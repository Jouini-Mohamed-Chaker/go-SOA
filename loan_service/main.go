package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	_ "github.com/lib/pq"
)

// Loan model matching the documentation
type Loan struct {
	ID         int       `json:"id"`
	UserID     int       `json:"userId"`
	BookID     int       `json:"bookId"`
	LoanDate   time.Time `json:"loanDate"`
	DueDate    time.Time `json:"dueDate"`
	ReturnDate *time.Time `json:"returnDate"`
	Status     string    `json:"status"` // ACTIVE or RETURNED
}

// Book model matching the documentation
type Book struct {
	ID                int    `json:"id"`
	ISBN              string `json:"isbn"`
	Title             string `json:"title"`
	Author            string `json:"author"`
	PublishYear       int    `json:"publishYear"`
	Category          string `json:"category"`
	AvailableQuantity int    `json:"availableQuantity"`
}

type LoanResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Loan    *Loan  `json:"loan,omitempty"`
}

type LoansResult struct {
	Success bool   `json:"success"`
	Loans   []Loan `json:"loans"`
}

var db *sql.DB

func main() {
	var err error
	
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "library")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection with retries
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Database not ready, retrying in 2 seconds... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		log.Fatal("Failed to ping database after retries:", err)
	}

	log.Println("Connected to database successfully")

	// Setup routes
	http.HandleFunc("/loan", handleLoan)

	port := "8083"
	log.Printf("Loan Service (Manual SOAP) listening on port %s\n", port)
	log.Printf("SOAP endpoint available at http://localhost:%s/loan\n", port)
	
	if err := http.ListenAndServe(":"+port, corsMiddleware(http.DefaultServeMux)); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, SOAPAction, soapaction")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		serveWSDL(w, r)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(buildErrorResponse("Failed to read request body")))
		return
	}

	soapBody := string(body)
	var responseXML string

	if contains(soapBody, "createLoan") {
		userID := extractValue(soapBody, "userId")
		bookID := extractValue(soapBody, "bookId")
		result := createLoan(userID, bookID)
		responseXML = buildCreateLoanResponse(result)
	} else if contains(soapBody, "returnLoan") {
		loanID := extractValue(soapBody, "loanId")
		result := returnLoan(loanID)
		responseXML = buildReturnLoanResponse(result)
	} else if contains(soapBody, "getLoansByUser") {
		userID := extractValue(soapBody, "userId")
		result := getLoansByUser(userID)
		responseXML = buildGetLoansByUserResponse(result)
	} else if contains(soapBody, "getLoanById") {
		loanID := extractValue(soapBody, "loanId")
		result := getLoanById(loanID)
		responseXML = buildGetLoanByIdResponse(result)
	} else if contains(soapBody, "getAllLoans") {
		result := getAllLoans()
		responseXML = buildGetAllLoansResponse(result)
	} else {
		responseXML = buildErrorResponse("Unknown operation")
	}

	w.Write([]byte(responseXML))
}

func serveWSDL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	wsdl, err := os.ReadFile("./loan.wsdl")
	if err != nil {
		http.Error(w, "WSDL not found", http.StatusNotFound)
		return
	}
	w.Write(wsdl)
}

func extractValue(xml, tagName string) string {
	pattern := fmt.Sprintf(`<%s>(.*?)</%s>`, tagName, tagName)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(xml)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func contains(s, substr string) bool {
	return regexp.MustCompile(substr).MatchString(s)
}

// createLoan implements the SOAP operation as per documentation:
// 1. Check if book exists (call Book Service)
// 2. Check if book is available (availableQuantity > 0)
// 3. Create loan with status ACTIVE
// 4. Decrease book's availableQuantity by 1
func createLoan(userID, bookID string) LoanResult {
	// Step 1: Check if book exists
	book, err := fetchBook(bookID)
	if err != nil {
		return LoanResult{Success: false, Message: err.Error(), Loan: nil}
	}

	// Step 2: Check if book is available
	if book.AvailableQuantity <= 0 {
		return LoanResult{Success: false, Message: "Book is not available", Loan: nil}
	}

	// Step 3: Create loan with status ACTIVE, dueDate = loanDate + 14 days
	loanDate := time.Now()
	dueDate := loanDate.AddDate(0, 0, 14) // Add 14 days as per documentation

	var loan Loan
	err = db.QueryRow(
		`INSERT INTO loans (user_id, book_id, loan_date, due_date, status) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, book_id, loan_date, due_date, return_date, status`,
		userID, bookID, loanDate, dueDate, "ACTIVE",
	).Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &loan.ReturnDate, &loan.Status)

	if err != nil {
		return LoanResult{Success: false, Message: err.Error(), Loan: nil}
	}

	// Step 4: Decrease book's availableQuantity by 1
	book.AvailableQuantity--
	if err := updateBook(bookID, book); err != nil {
		return LoanResult{Success: false, Message: err.Error(), Loan: nil}
	}

	return LoanResult{Success: true, Message: "Loan created successfully", Loan: &loan}
}

// returnLoan implements the SOAP operation as per documentation:
// 1. Find loan by ID
// 2. Set returnDate to current date
// 3. Set status to RETURNED
// 4. Increase book's availableQuantity by 1
func returnLoan(loanID string) LoanResult {
	// Step 1: Find loan by ID
	var loan Loan
	var returnDate sql.NullTime
	
	err := db.QueryRow("SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans WHERE id = $1", loanID).
		Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status)

	if err == sql.ErrNoRows {
		return LoanResult{Success: false, Message: "Loan not found", Loan: nil}
	}
	if err != nil {
		return LoanResult{Success: false, Message: err.Error(), Loan: nil}
	}

	if loan.Status == "RETURNED" {
		return LoanResult{Success: false, Message: "Loan already returned", Loan: nil}
	}

	// Step 2: Set returnDate to current date
	// Step 3: Set status to RETURNED
	returnTime := time.Now()
	_, err = db.Exec(
		"UPDATE loans SET return_date = $1, status = $2 WHERE id = $3",
		returnTime, "RETURNED", loanID,
	)
	if err != nil {
		return LoanResult{Success: false, Message: err.Error(), Loan: nil}
	}

	// Step 4: Increase book's availableQuantity by 1
	book, err := fetchBook(fmt.Sprintf("%d", loan.BookID))
	if err != nil {
		return LoanResult{Success: false, Message: err.Error(), Loan: nil}
	}

	book.AvailableQuantity++
	if err := updateBook(fmt.Sprintf("%d", loan.BookID), book); err != nil {
		return LoanResult{Success: false, Message: err.Error(), Loan: nil}
	}

	loan.Status = "RETURNED"
	loan.ReturnDate = &returnTime

	return LoanResult{Success: true, Message: "Loan returned successfully", Loan: &loan}
}

func getLoansByUser(userID string) LoansResult {
	rows, err := db.Query("SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans WHERE user_id = $1", userID)
	if err != nil {
		return LoansResult{Success: false, Loans: []Loan{}}
	}
	defer rows.Close()

	var loans []Loan
	for rows.Next() {
		var loan Loan
		var returnDate sql.NullTime
		if err := rows.Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status); err != nil {
			continue
		}
		if returnDate.Valid {
			loan.ReturnDate = &returnDate.Time
		}
		loans = append(loans, loan)
	}

	return LoansResult{Success: true, Loans: loans}
}

func getLoanById(loanID string) LoanResult {
	var loan Loan
	var returnDate sql.NullTime
	
	err := db.QueryRow("SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans WHERE id = $1", loanID).
		Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status)

	if err == sql.ErrNoRows {
		return LoanResult{Success: false, Message: "Loan not found", Loan: nil}
	}
	if err != nil {
		return LoanResult{Success: false, Message: err.Error(), Loan: nil}
	}

	if returnDate.Valid {
		loan.ReturnDate = &returnDate.Time
	}

	return LoanResult{Success: true, Message: "Loan found", Loan: &loan}
}

func getAllLoans() LoansResult {
	rows, err := db.Query("SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans ORDER BY loan_date DESC")
	if err != nil {
		return LoansResult{Success: false, Loans: []Loan{}}
	}
	defer rows.Close()

	var loans []Loan
	for rows.Next() {
		var loan Loan
		var returnDate sql.NullTime
		if err := rows.Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status); err != nil {
			continue
		}
		if returnDate.Valid {
			loan.ReturnDate = &returnDate.Time
		}
		loans = append(loans, loan)
	}

	return LoansResult{Success: true, Loans: loans}
}

func fetchBook(bookID string) (*Book, error) {
	resp, err := http.Get(fmt.Sprintf("http://book_service:8081/api/books/%s", bookID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("book not found")
	}

	var book Book
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return nil, err
	}

	return &book, nil
}

func updateBook(bookID string, book *Book) error {
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://book_service:8081/api/books/%s", bookID), bytes.NewBuffer(bookJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update book")
	}

	return nil
}

func buildCreateLoanResponse(result LoanResult) string {
	loanXML := ""
	if result.Loan != nil {
		returnDate := ""
		if result.Loan.ReturnDate != nil {
			returnDate = result.Loan.ReturnDate.Format(time.RFC3339)
		}
		loanXML = fmt.Sprintf(`
    <loan>
      <id>%d</id>
      <userId>%d</userId>
      <bookId>%d</bookId>
      <loanDate>%s</loanDate>
      <dueDate>%s</dueDate>
      <returnDate>%s</returnDate>
      <status>%s</status>
    </loan>`, result.Loan.ID, result.Loan.UserID, result.Loan.BookID,
			result.Loan.LoanDate.Format(time.RFC3339),
			result.Loan.DueDate.Format(time.RFC3339),
			returnDate, result.Loan.Status)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <createLoanResponse xmlns="http://loan.service">
      <success>%t</success>
      <message>%s</message>
      %s
    </createLoanResponse>
  </soap:Body>
</soap:Envelope>`, result.Success, xmlEscape(result.Message), loanXML)
}

func buildReturnLoanResponse(result LoanResult) string {
	loanXML := ""
	if result.Loan != nil {
		returnDate := ""
		if result.Loan.ReturnDate != nil {
			returnDate = result.Loan.ReturnDate.Format(time.RFC3339)
		}
		loanXML = fmt.Sprintf(`
    <loan>
      <id>%d</id>
      <userId>%d</userId>
      <bookId>%d</bookId>
      <loanDate>%s</loanDate>
      <dueDate>%s</dueDate>
      <returnDate>%s</returnDate>
      <status>%s</status>
    </loan>`, result.Loan.ID, result.Loan.UserID, result.Loan.BookID,
			result.Loan.LoanDate.Format(time.RFC3339),
			result.Loan.DueDate.Format(time.RFC3339),
			returnDate, result.Loan.Status)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <returnLoanResponse xmlns="http://loan.service">
      <success>%t</success>
      <message>%s</message>
      %s
    </returnLoanResponse>
  </soap:Body>
</soap:Envelope>`, result.Success, xmlEscape(result.Message), loanXML)
}

func buildGetAllLoansResponse(result LoansResult) string {
	loansXML := ""
	for _, loan := range result.Loans {
		returnDate := ""
		if loan.ReturnDate != nil {
			returnDate = loan.ReturnDate.Format(time.RFC3339)
		}
		loansXML += fmt.Sprintf(`
    <loan>
      <id>%d</id>
      <userId>%d</userId>
      <bookId>%d</bookId>
      <loanDate>%s</loanDate>
      <dueDate>%s</dueDate>
      <returnDate>%s</returnDate>
      <status>%s</status>
    </loan>`, loan.ID, loan.UserID, loan.BookID,
			loan.LoanDate.Format(time.RFC3339),
			loan.DueDate.Format(time.RFC3339),
			returnDate, loan.Status)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getAllLoansResponse xmlns="http://loan.service">
      <success>%t</success>
      %s
    </getAllLoansResponse>
  </soap:Body>
</soap:Envelope>`, result.Success, loansXML)
}

func buildGetLoansByUserResponse(result LoansResult) string {
	loansXML := ""
	for _, loan := range result.Loans {
		returnDate := ""
		if loan.ReturnDate != nil {
			returnDate = loan.ReturnDate.Format(time.RFC3339)
		}
		loansXML += fmt.Sprintf(`
    <loan>
      <id>%d</id>
      <userId>%d</userId>
      <bookId>%d</bookId>
      <loanDate>%s</loanDate>
      <dueDate>%s</dueDate>
      <returnDate>%s</returnDate>
      <status>%s</status>
    </loan>`, loan.ID, loan.UserID, loan.BookID,
			loan.LoanDate.Format(time.RFC3339),
			loan.DueDate.Format(time.RFC3339),
			returnDate, loan.Status)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getLoansByUserResponse xmlns="http://loan.service">
      <success>%t</success>
      %s
    </getLoansByUserResponse>
  </soap:Body>
</soap:Envelope>`, result.Success, loansXML)
}

func buildGetLoanByIdResponse(result LoanResult) string {
	loanXML := ""
	if result.Loan != nil {
		returnDate := ""
		if result.Loan.ReturnDate != nil {
			returnDate = result.Loan.ReturnDate.Format(time.RFC3339)
		}
		loanXML = fmt.Sprintf(`
    <loan>
      <id>%d</id>
      <userId>%d</userId>
      <bookId>%d</bookId>
      <loanDate>%s</loanDate>
      <dueDate>%s</dueDate>
      <returnDate>%s</returnDate>
      <status>%s</status>
    </loan>`, result.Loan.ID, result.Loan.UserID, result.Loan.BookID,
			result.Loan.LoanDate.Format(time.RFC3339),
			result.Loan.DueDate.Format(time.RFC3339),
			returnDate, result.Loan.Status)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getLoanByIdResponse xmlns="http://loan.service">
      <success>%t</success>
      <message>%s</message>
      %s
    </getLoanByIdResponse>
  </soap:Body>
</soap:Envelope>`, result.Success, xmlEscape(result.Message), loanXML)
}

func buildErrorResponse(message string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <soap:Fault>
      <faultcode>soap:Server</faultcode>
      <faultstring>%s</faultstring>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`, xmlEscape(message))
}

func xmlEscape(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
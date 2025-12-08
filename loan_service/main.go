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
	ID         int        `json:"id"`
	UserID     int        `json:"userId"`
	BookID     int        `json:"bookId"`
	LoanDate   time.Time  `json:"loanDate"`
	DueDate    time.Time  `json:"dueDate"`
	ReturnDate *time.Time `json:"returnDate"`
	Status     string     `json:"status"` // ACTIVE or RETURNED
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
	Loan  *Loan  `json:"loan,omitempty"`
	Error string `json:"error,omitempty"`
}

type LoansResult struct {
	Loans []Loan `json:"loans"`
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

	// Setup routes - handle both /ws and /loan for compatibility
	http.HandleFunc("/ws", handleLoan)
	http.HandleFunc("/loan", handleLoan)

	port := "8083"
	log.Printf("Loan Service listening on port %s\n", port)
	log.Printf("SOAP endpoints available at:")
	log.Printf("  http://localhost:%s/ws", port)
	log.Printf("  http://localhost:%s/loan", port)
	
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
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered: %v", r)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(buildErrorResponse("Internal server error")))
		}
	}()

	if r.Method == "GET" {
		// For GET requests, return a simple message or WSDL
		w.Header().Set("Content-Type", "text/xml")
		wsdl := `<?xml version="1.0"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/loan"
             targetNamespace="http://example.com/loan">
  <types>
    <xsd:schema xmlns:xsd="http://www.w3.org/2001/XMLSchema"
                targetNamespace="http://example.com/loan">
      <xsd:element name="createLoanRequest">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="userId" type="xsd:integer"/>
            <xsd:element name="bookId" type="xsd:integer"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
      <xsd:element name="createLoanResponse">
        <xsd:complexType>
          <xsd:sequence>
            <xsd:element name="loan" type="tns:loanType" minOccurs="0"/>
            <xsd:element name="error" type="xsd:string" minOccurs="0"/>
          </xsd:sequence>
        </xsd:complexType>
      </xsd:element>
      <xsd:complexType name="loanType">
        <xsd:sequence>
          <xsd:element name="id" type="xsd:integer"/>
          <xsd:element name="userId" type="xsd:integer"/>
          <xsd:element name="bookId" type="xsd:integer"/>
          <xsd:element name="loanDate" type="xsd:dateTime"/>
          <xsd:element name="dueDate" type="xsd:dateTime"/>
          <xsd:element name="returnDate" type="xsd:dateTime" minOccurs="0"/>
          <xsd:element name="status" type="xsd:string"/>
        </xsd:sequence>
      </xsd:complexType>
    </xsd:schema>
  </types>
  <message name="createLoanRequest">
    <part name="parameters" element="tns:createLoanRequest"/>
  </message>
  <message name="createLoanResponse">
    <part name="parameters" element="tns:createLoanResponse"/>
  </message>
  <portType name="LoanServicePortType">
    <operation name="createLoan">
      <input message="tns:createLoanRequest"/>
      <output message="tns:createLoanResponse"/>
    </operation>
  </portType>
  <binding name="LoanServiceBinding" type="tns:LoanServicePortType">
    <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
    <operation name="createLoan">
      <soap:operation soapAction="createLoan"/>
      <input>
        <soap:body use="literal"/>
      </input>
      <output>
        <soap:body use="literal"/>
      </output>
    </operation>
  </binding>
  <service name="LoanService">
    <port name="LoanServicePort" binding="tns:LoanServiceBinding">
      <soap:address location="http://localhost:8083/ws"/>
    </port>
  </service>
</definitions>`
		w.Write([]byte(wsdl))
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
	log.Printf("Received SOAP request: %s", soapBody[:min(200, len(soapBody))])

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

	log.Printf("Sending response: %s", responseXML[:min(200, len(responseXML))])
	w.Write([]byte(responseXML))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// createLoan implements the SOAP operation as per documentation
func createLoan(userID, bookID string) LoanResult {
	// Validate inputs
	if userID == "" || bookID == "" {
		return LoanResult{Error: "User ID and Book ID are required"}
	}

	// Step 1: Check if book exists
	book, err := fetchBook(bookID)
	if err != nil {
		log.Printf("Error fetching book %s: %v", bookID, err)
		return LoanResult{Error: "Book not found or book service unavailable"}
	}

	// Step 2: Check if book is available
	if book.AvailableQuantity <= 0 {
		return LoanResult{Error: "Book is not available"}
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
		log.Printf("Error creating loan: %v", err)
		return LoanResult{Error: "Failed to create loan: " + err.Error()}
	}

	// Step 4: Decrease book's availableQuantity by 1
	book.AvailableQuantity--
	if err := updateBook(bookID, book); err != nil {
		// If we fail to update the book, rollback the loan
		db.Exec("DELETE FROM loans WHERE id = $1", loan.ID)
		log.Printf("Error updating book quantity: %v", err)
		return LoanResult{Error: "Failed to update book quantity"}
	}

	log.Printf("Created loan ID %d for user %s, book %s", loan.ID, userID, bookID)
	return LoanResult{Loan: &loan}
}

// returnLoan implements the SOAP operation as per documentation
func returnLoan(loanID string) LoanResult {
	if loanID == "" {
		return LoanResult{Error: "Loan ID is required"}
	}

	// Step 1: Find loan by ID
	var loan Loan
	var returnDate sql.NullTime
	
	err := db.QueryRow("SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans WHERE id = $1", loanID).
		Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status)

	if err == sql.ErrNoRows {
		return LoanResult{Error: "Loan not found"}
	}
	if err != nil {
		log.Printf("Error finding loan %s: %v", loanID, err)
		return LoanResult{Error: "Database error: " + err.Error()}
	}

	if loan.Status == "RETURNED" {
		return LoanResult{Error: "Loan already returned"}
	}

	// Step 2: Set returnDate to current date
	// Step 3: Set status to RETURNED
	returnTime := time.Now()
	_, err = db.Exec(
		"UPDATE loans SET return_date = $1, status = $2 WHERE id = $3",
		returnTime, "RETURNED", loanID,
	)
	if err != nil {
		log.Printf("Error updating loan %s: %v", loanID, err)
		return LoanResult{Error: "Failed to update loan: " + err.Error()}
	}

	// Step 4: Increase book's availableQuantity by 1
	book, err := fetchBook(fmt.Sprintf("%d", loan.BookID))
	if err != nil {
		log.Printf("Error fetching book for return: %v", err)
		return LoanResult{Error: "Book service error during return"}
	}

	book.AvailableQuantity++
	if err := updateBook(fmt.Sprintf("%d", loan.BookID), book); err != nil {
		log.Printf("Error updating book quantity on return: %v", err)
		return LoanResult{Error: "Failed to update book quantity on return"}
	}

	loan.Status = "RETURNED"
	loan.ReturnDate = &returnTime

	log.Printf("Returned loan ID %s", loanID)
	return LoanResult{Loan: &loan}
}

func getLoansByUser(userID string) LoansResult {
	if userID == "" {
		return LoansResult{Loans: []Loan{}}
	}

	rows, err := db.Query("SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans WHERE user_id = $1 ORDER BY loan_date DESC", userID)
	if err != nil {
		log.Printf("Error querying loans for user %s: %v", userID, err)
		return LoansResult{Loans: []Loan{}}
	}
	defer rows.Close()

	var loans []Loan
	for rows.Next() {
		var loan Loan
		var returnDate sql.NullTime
		if err := rows.Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status); err != nil {
			log.Printf("Error scanning loan row: %v", err)
			continue
		}
		if returnDate.Valid {
			loan.ReturnDate = &returnDate.Time
		}
		loans = append(loans, loan)
	}

	return LoansResult{Loans: loans}
}

func getLoanById(loanID string) LoanResult {
	if loanID == "" {
		return LoanResult{Error: "Loan ID is required"}
	}

	var loan Loan
	var returnDate sql.NullTime
	
	err := db.QueryRow("SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans WHERE id = $1", loanID).
		Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status)

	if err == sql.ErrNoRows {
		return LoanResult{Error: "Loan not found"}
	}
	if err != nil {
		log.Printf("Error getting loan %s: %v", loanID, err)
		return LoanResult{Error: "Database error: " + err.Error()}
	}

	if returnDate.Valid {
		loan.ReturnDate = &returnDate.Time
	}

	return LoanResult{Loan: &loan}
}

func getAllLoans() LoansResult {
	rows, err := db.Query("SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans ORDER BY loan_date DESC")
	if err != nil {
		log.Printf("Error querying all loans: %v", err)
		return LoansResult{Loans: []Loan{}}
	}
	defer rows.Close()

	var loans []Loan
	for rows.Next() {
		var loan Loan
		var returnDate sql.NullTime
		if err := rows.Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status); err != nil {
			log.Printf("Error scanning loan row: %v", err)
			continue
		}
		if returnDate.Valid {
			loan.ReturnDate = &returnDate.Time
		}
		loans = append(loans, loan)
	}

	return LoansResult{Loans: loans}
}

func fetchBook(bookID string) (*Book, error) {
	// Try localhost first for testing, then the service name
	urls := []string{
		fmt.Sprintf("http://localhost:8081/api/books/%s", bookID),
		fmt.Sprintf("http://book_service:8081/api/books/%s", bookID),
	}

	var lastErr error
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var book Book
			if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
				lastErr = err
				continue
			}
			return &book, nil
		}
		lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return nil, fmt.Errorf("failed to fetch book: %v", lastErr)
}

func updateBook(bookID string, book *Book) error {
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return err
	}

	// Try localhost first for testing, then the service name
	urls := []string{
		fmt.Sprintf("http://localhost:8081/api/books/%s", bookID),
		fmt.Sprintf("http://book_service:8081/api/books/%s", bookID),
	}

	var lastErr error
	for _, url := range urls {
		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(bookJSON))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}
		lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return fmt.Errorf("failed to update book: %v", lastErr)
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

	errorXML := xmlEscape(result.Error)

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <createLoanResponse xmlns="http://example.com/loan">
      %s
      <error>%s</error>
    </createLoanResponse>
  </soap:Body>
</soap:Envelope>`, loanXML, errorXML)
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

	errorXML := xmlEscape(result.Error)

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <returnLoanResponse xmlns="http://example.com/loan">
      %s
      <error>%s</error>
    </returnLoanResponse>
  </soap:Body>
</soap:Envelope>`, loanXML, errorXML)
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
    <getAllLoansResponse xmlns="http://example.com/loan">
      %s
    </getAllLoansResponse>
  </soap:Body>
</soap:Envelope>`, loansXML)
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
    <getLoansByUserResponse xmlns="http://example.com/loan">
      %s
    </getLoansByUserResponse>
  </soap:Body>
</soap:Envelope>`, loansXML)
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

	errorXML := xmlEscape(result.Error)

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getLoanByIdResponse xmlns="http://example.com/loan">
      %s
      <error>%s</error>
    </getLoanByIdResponse>
  </soap:Body>
</soap:Envelope>`, loanXML, errorXML)
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
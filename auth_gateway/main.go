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
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var (
	db        *sql.DB
	jwtSecret = []byte("24abd7d0df965baabb514fc50c30f30a04e82ac50260700c35089ab593479015")
)

func main() {
	var err error
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "library")
	jwtSecretEnv := getEnv("JWT_SECRET", "")
	if jwtSecretEnv != "" {
		jwtSecret = []byte(jwtSecretEnv)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Auth Gateway connected to database")

	router := mux.NewRouter()
	router.HandleFunc("/auth/login", handleLogin)
	router.HandleFunc("/auth/register", handleRegister)
	router.HandleFunc("/auth/validate", handleValidate)
	router.HandleFunc("/api/books", jwtMiddleware(proxyBooks))
	router.PathPrefix("/api/books/").HandlerFunc(jwtMiddleware(proxyBooks))
	router.HandleFunc("/api/users", jwtMiddleware(proxyUsers))
	router.PathPrefix("/api/users/").HandlerFunc(jwtMiddleware(proxyUsers))
	router.HandleFunc("/api/loans", jwtMiddleware(proxyLoans))
	router.PathPrefix("/api/loans/").HandlerFunc(jwtMiddleware(proxyLoans))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Println("Auth Gateway starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), "Invalid request body", http.StatusBadRequest)
		return
	}

	var userID int64
	var passwordHash string
	err := db.QueryRow("SELECT user_id, password_hash FROM user_credentials WHERE username = $1", req.Username).Scan(&userID, &passwordHash)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid username or password"})
		return
	}
	if err != nil {
		sendError(w, err.Error(), "Database error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid username or password"})
		return
	}

	token, err := generateJWT(req.Username)
	if err != nil {
		sendError(w, err.Error(), "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:     token,
		Username:  req.Username,
		ExpiresIn: 36000,
	})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), "Invalid request body", http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", req.Username).Scan(&exists)
	if err != nil {
		sendError(w, err.Error(), "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Username already exists"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		sendError(w, err.Error(), "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var userID int64
	err = tx.QueryRow("INSERT INTO users (username, email, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Username, req.Email, req.FirstName, req.LastName).Scan(&userID)
	if err != nil {
		sendError(w, err.Error(), "Failed to create user", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		sendError(w, err.Error(), "Failed to hash password", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO user_credentials (user_id, username, password_hash) VALUES ($1, $2, $3)",
		userID, req.Username, string(hashedPassword))
	if err != nil {
		sendError(w, err.Error(), "Failed to create credentials", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		sendError(w, err.Error(), "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserResponse{
		ID:        userID,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), "Invalid request body", http.StatusBadRequest)
		return
	}

	username, err := validateJWT(req.Token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ValidateResponse{
			Valid: false,
			Error: "Invalid or expired token",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ValidateResponse{
		Valid:    true,
		Username: username,
	})
}

func generateJWT(username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": username,
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["sub"].(string)
		if !ok {
			return "", fmt.Errorf("invalid token claims")
		}
		return username, nil
	}
	return "", fmt.Errorf("invalid token")
}

func jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Unauthorized",
				Message: "Valid JWT token required",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Unauthorized",
				Message: "Valid JWT token required",
			})
			return
		}

		_, err := validateJWT(parts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Unauthorized",
				Message: "Valid JWT token required",
			})
			return
		}

		next(w, r)
	}
}

func proxyBooks(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/books")
	targetURL := "http://book_service:8081/api/books" + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	proxyRequest(w, r, targetURL)
}

func proxyUsers(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/users")
	targetURL := "http://user_service:8082/api/users" + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	proxyRequest(w, r, targetURL)
}

func proxyLoans(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/loans")

	if r.Method == http.MethodPost && path == "" {
		handleCreateLoan(w, r)
		return
	}

	if r.Method == http.MethodPut && strings.HasSuffix(path, "/return") {
		handleReturnLoan(w, r, path)
		return
	}

	if r.Method == http.MethodGet && strings.HasPrefix(path, "/user/") {
		handleGetLoansByUser(w, r, path)
		return
	}

	if r.Method == http.MethodGet && path != "" && path != "/" {
		handleGetLoanById(w, r, path)
		return
	}

	if r.Method == http.MethodGet && (path == "" || path == "/") {
		handleGetAllLoans(w, r)
		return
	}

	sendError(w, "", "Not found", http.StatusNotFound)
}

func handleCreateLoan(w http.ResponseWriter, r *http.Request) {
	var req CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), "Invalid request body", http.StatusBadRequest)
		return
	}

	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:loan="http://example.com/loan">
   <soapenv:Header/>
   <soapenv:Body>
      <loan:createLoan>
         <userId>%d</userId>
         <bookId>%d</bookId>
      </loan:createLoan>
   </soapenv:Body>
</soapenv:Envelope>`, req.UserID, req.BookID)

	resp, err := http.Post("http://loan_service:8083/ws", "text/xml", bytes.NewBufferString(soapBody))
	if err != nil {
		sendError(w, err.Error(), "Failed to contact loan service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	type CreateLoanResponse struct {
		XMLName struct{} `xml:"Envelope"`
		Body    struct {
			CreateLoanResponse struct {
				Loan struct {
					ID         int64  `xml:"id"`
					UserID     int64  `xml:"userId"`
					BookID     int64  `xml:"bookId"`
					LoanDate   string `xml:"loanDate"`
					DueDate    string `xml:"dueDate"`
					ReturnDate string `xml:"returnDate"`
					Status     string `xml:"status"`
				} `xml:"loan"`
				Error string `xml:"error"`
			} `xml:"createLoanResponse"`
		} `xml:"Body"`
	}

	var soapResp CreateLoanResponse
	if err := xml.Unmarshal(body, &soapResp); err != nil {
		sendError(w, err.Error(), "Failed to parse response", http.StatusInternalServerError)
		return
	}

	if soapResp.Body.CreateLoanResponse.Error != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: soapResp.Body.CreateLoanResponse.Error})
		return
	}

	loan := soapResp.Body.CreateLoanResponse.Loan
	var returnDate *string
	if loan.ReturnDate != "" {
		returnDate = &loan.ReturnDate
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(LoanResponse{
		ID:         loan.ID,
		UserID:     loan.UserID,
		BookID:     loan.BookID,
		LoanDate:   loan.LoanDate,
		DueDate:    loan.DueDate,
		ReturnDate: returnDate,
		Status:     loan.Status,
	})
}

func handleReturnLoan(w http.ResponseWriter, r *http.Request, path string) {
	loanID := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/return")

	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:loan="http://example.com/loan">
   <soapenv:Header/>
   <soapenv:Body>
      <loan:returnLoan>
         <loanId>%s</loanId>
      </loan:returnLoan>
   </soapenv:Body>
</soapenv:Envelope>`, loanID)

	resp, err := http.Post("http://loan_service:8083/ws", "text/xml", bytes.NewBufferString(soapBody))
	if err != nil {
		sendError(w, err.Error(), "Failed to contact loan service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	type ReturnLoanResponse struct {
		XMLName struct{} `xml:"Envelope"`
		Body    struct {
			ReturnLoanResponse struct {
				Loan struct {
					ID         int64  `xml:"id"`
					UserID     int64  `xml:"userId"`
					BookID     int64  `xml:"bookId"`
					LoanDate   string `xml:"loanDate"`
					DueDate    string `xml:"dueDate"`
					ReturnDate string `xml:"returnDate"`
					Status     string `xml:"status"`
				} `xml:"loan"`
				Error string `xml:"error"`
			} `xml:"returnLoanResponse"`
		} `xml:"Body"`
	}

	var soapResp ReturnLoanResponse
	if err := xml.Unmarshal(body, &soapResp); err != nil {
		sendError(w, err.Error(), "Failed to parse response", http.StatusInternalServerError)
		return
	}

	if soapResp.Body.ReturnLoanResponse.Error != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: soapResp.Body.ReturnLoanResponse.Error})
		return
	}

	loan := soapResp.Body.ReturnLoanResponse.Loan
	var returnDate *string
	if loan.ReturnDate != "" {
		returnDate = &loan.ReturnDate
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoanResponse{
		ID:         loan.ID,
		UserID:     loan.UserID,
		BookID:     loan.BookID,
		LoanDate:   loan.LoanDate,
		DueDate:    loan.DueDate,
		ReturnDate: returnDate,
		Status:     loan.Status,
	})
}

func handleGetLoansByUser(w http.ResponseWriter, r *http.Request, path string) {
	userID := strings.TrimPrefix(path, "/user/")

	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:loan="http://example.com/loan">
   <soapenv:Header/>
   <soapenv:Body>
      <loan:getLoansByUser>
         <userId>%s</userId>
      </loan:getLoansByUser>
   </soapenv:Body>
</soapenv:Envelope>`, userID)

	resp, err := http.Post("http://loan_service:8083/ws", "text/xml", bytes.NewBufferString(soapBody))
	if err != nil {
		sendError(w, err.Error(), "Failed to contact loan service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	type GetLoansByUserResponse struct {
		XMLName struct{} `xml:"Envelope"`
		Body    struct {
			GetLoansByUserResponse struct {
				Loans []struct {
					ID         int64  `xml:"id"`
					UserID     int64  `xml:"userId"`
					BookID     int64  `xml:"bookId"`
					LoanDate   string `xml:"loanDate"`
					DueDate    string `xml:"dueDate"`
					ReturnDate string `xml:"returnDate"`
					Status     string `xml:"status"`
				} `xml:"loan"`
			} `xml:"getLoansByUserResponse"`
		} `xml:"Body"`
	}

	var soapResp GetLoansByUserResponse
	if err := xml.Unmarshal(body, &soapResp); err != nil {
		sendError(w, err.Error(), "Failed to parse response", http.StatusInternalServerError)
		return
	}

	loans := make([]LoanResponse, 0)
	for _, loan := range soapResp.Body.GetLoansByUserResponse.Loans {
		var returnDate *string
		if loan.ReturnDate != "" {
			returnDate = &loan.ReturnDate
		}
		loans = append(loans, LoanResponse{
			ID:         loan.ID,
			UserID:     loan.UserID,
			BookID:     loan.BookID,
			LoanDate:   loan.LoanDate,
			DueDate:    loan.DueDate,
			ReturnDate: returnDate,
			Status:     loan.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loans)
}

func handleGetLoanById(w http.ResponseWriter, r *http.Request, path string) {
	loanID := strings.TrimPrefix(path, "/")

	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:loan="http://example.com/loan">
   <soapenv:Header/>
   <soapenv:Body>
      <loan:getLoanById>
         <loanId>%s</loanId>
      </loan:getLoanById>
   </soapenv:Body>
</soapenv:Envelope>`, loanID)

	resp, err := http.Post("http://loan_service:8083/ws", "text/xml", bytes.NewBufferString(soapBody))
	if err != nil {
		sendError(w, err.Error(), "Failed to contact loan service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	type GetLoanByIdResponse struct {
		XMLName struct{} `xml:"Envelope"`
		Body    struct {
			GetLoanByIdResponse struct {
				Loan struct {
					ID         int64  `xml:"id"`
					UserID     int64  `xml:"userId"`
					BookID     int64  `xml:"bookId"`
					LoanDate   string `xml:"loanDate"`
					DueDate    string `xml:"dueDate"`
					ReturnDate string `xml:"returnDate"`
					Status     string `xml:"status"`
				} `xml:"loan"`
				Error string `xml:"error"`
			} `xml:"getLoanByIdResponse"`
		} `xml:"Body"`
	}

	var soapResp GetLoanByIdResponse
	if err := xml.Unmarshal(body, &soapResp); err != nil {
		sendError(w, err.Error(), "Failed to parse response", http.StatusInternalServerError)
		return
	}

	if soapResp.Body.GetLoanByIdResponse.Error != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: soapResp.Body.GetLoanByIdResponse.Error})
		return
	}

	loan := soapResp.Body.GetLoanByIdResponse.Loan
	var returnDate *string
	if loan.ReturnDate != "" {
		returnDate = &loan.ReturnDate
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoanResponse{
		ID:         loan.ID,
		UserID:     loan.UserID,
		BookID:     loan.BookID,
		LoanDate:   loan.LoanDate,
		DueDate:    loan.DueDate,
		ReturnDate: returnDate,
		Status:     loan.Status,
	})
}

func handleGetAllLoans(w http.ResponseWriter, r *http.Request) {
	soapBody := `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:loan="http://example.com/loan">
   <soapenv:Header/>
   <soapenv:Body>
      <loan:getAllLoans/>
   </soapenv:Body>
</soapenv:Envelope>`

	resp, err := http.Post("http://loan_service:8083/ws", "text/xml", bytes.NewBufferString(soapBody))
	if err != nil {
		sendError(w, err.Error(), "Failed to contact loan service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	type GetAllLoansResponse struct {
		XMLName struct{} `xml:"Envelope"`
		Body    struct {
			GetAllLoansResponse struct {
				Loans []struct {
					ID         int64  `xml:"id"`
					UserID     int64  `xml:"userId"`
					BookID     int64  `xml:"bookId"`
					LoanDate   string `xml:"loanDate"`
					DueDate    string `xml:"dueDate"`
					ReturnDate string `xml:"returnDate"`
					Status     string `xml:"status"`
				} `xml:"loan"`
			} `xml:"getAllLoansResponse"`
		} `xml:"Body"`
	}

	var soapResp GetAllLoansResponse
	if err := xml.Unmarshal(body, &soapResp); err != nil {
		sendError(w, err.Error(), "Failed to parse response", http.StatusInternalServerError)
		return
	}

	loans := make([]LoanResponse, 0)
	for _, loan := range soapResp.Body.GetAllLoansResponse.Loans {
		var returnDate *string
		if loan.ReturnDate != "" {
			returnDate = &loan.ReturnDate
		}
		loans = append(loans, LoanResponse{
			ID:         loan.ID,
			UserID:     loan.UserID,
			BookID:     loan.BookID,
			LoanDate:   loan.LoanDate,
			DueDate:    loan.DueDate,
			ReturnDate: returnDate,
			Status:     loan.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loans)
}

func proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	client := &http.Client{Timeout: 10 * time.Second}

	var body io.Reader
	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(r.Method, targetURL, body)
	if err != nil {
		sendError(w, err.Error(), "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		sendError(w, err.Error(), "Failed to contact service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func sendError(w http.ResponseWriter, err, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	customError := fmt.Sprintf("%s: %v", message, err)
	json.NewEncoder(w).Encode(ErrorResponse{Error: customError})
}

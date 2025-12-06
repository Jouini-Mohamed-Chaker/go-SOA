package main

type UserCredentials struct {
	UserID int64 `json:"userId"`
	Username string `json:"username"`
	PasswordHash string `json:"passwordHash"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Username string `json:"username"`
	ExpiresIn int `json:"expiresIn"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email string `json:"email"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID int64 `json:"id"`
	Username string `json:"username"`
	Email string `json:"email"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
}

type ValidateRequest struct {
	Token string `json:"token"`
}

type ValidateResponse struct {
	Valid bool `json:"valid"`
	Username string `json:"username,omitempty"`
	Error string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Message string `json:"message,omitempty"`
}

type CreateLoanRequest struct {
	UserID int64 `json:"userId"`
	BookID int64 `json:"bookId"`
}

type LoanResponse struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"userId"`
	BookID     int64   `json:"bookId"`
	LoanDate   string  `json:"loanDate"`
	DueDate    string  `json:"dueDate"`
	ReturnDate *string `json:"returnDate"`
	Status     string  `json:"status"`
}
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	_ "github.com/lib/pq"
)

type Book struct {
	ID                int64  `json:"id"`
	ISBN              string `json:"isbn"`
	Title             string `json:"title"`
	Author            string `json:"author"`
	PublishYear       uint   `json:"publishYear"`
	Category          string `json:"category"`
	AvailableQuantity uint   `json:"availableQuantity"`
}

type PaginatedResponse struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Total int    `json:"total,omitempty"`
	Data  []Book `json:"data"`
}

var db *sql.DB

func main() {
	var err error

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "library")

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("sql.Open error: %v", err)
	}

	const maxAttempts = 30
	for i := 1; i <= maxAttempts; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for Postgres (%d/%d): %v", i, maxAttempts, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("could not connect to db after %d attempts: %v", maxAttempts, err)
	}

	log.Println("Book Service connected to database")

	router := mux.NewRouter()
	router.HandleFunc("/api/books/search", searchBookByTitle).Methods("GET")
	router.HandleFunc("/api/books", getAllBooks).Methods("GET")
	router.HandleFunc("/api/books/{id}", getBookByID).Methods("GET")
	router.HandleFunc("/api/books", createBook).Methods("POST")
	router.HandleFunc("/api/books/{id}", updateBook).Methods("PUT")
	router.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Println("Book Service running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", handler))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getAllBooks(w http.ResponseWriter, r *http.Request) {
	page, limit := getPaginationParams(r)
	offset := (page - 1) * limit

	rows, err := db.Query(`
	SELECT id, isbn, title, author, publish_year, category, available_quantity 
	FROM books
	ORDER BY id
	LIMIT $1 OFFSET $2`,
		limit, offset)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.ISBN, &b.Title, &b.Author, &b.PublishYear, &b.Category, &b.AvailableQuantity); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	// Initialize empty slice if nil
	if books == nil {
		books = []Book{}
	}

	response := PaginatedResponse{
		Page:  page,
		Limit: limit,
		Data:  books,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return // FIX: Added missing return
	}

	var b Book
	err = db.QueryRow("SELECT id, isbn, title, author, publish_year, category, available_quantity FROM books WHERE id = $1", id).Scan(&b.ID, &b.ISBN, &b.Title, &b.Author, &b.PublishYear, &b.Category, &b.AvailableQuantity)

	if err == sql.ErrNoRows {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func searchBookByTitle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "Missing title query parameter", http.StatusBadRequest)
		return
	}

	page, limit := getPaginationParams(r)
	offset := (page - 1) * limit

	// Get total count
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM books WHERE title ILIKE $1", "%"+title+"%").Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`
	SELECT id, isbn, title, author, publish_year, category, available_quantity
	FROM books
	WHERE title ILIKE $1
	ORDER BY id
	LIMIT $2 OFFSET $3`,
		"%"+title+"%", limit, offset,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		err := rows.Scan(&b.ID, &b.ISBN, &b.Title, &b.Author, &b.PublishYear, &b.Category, &b.AvailableQuantity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	// Initialize empty slice if nil
	if books == nil {
		books = []Book{}
	}

	response := PaginatedResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  books,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if b.ISBN == "" || b.Title == "" || b.Author == "" {
		http.Error(w, "ISBN, title, and author are required fields", http.StatusBadRequest)
		return
	}

	err := db.QueryRow("INSERT INTO books (isbn, title, author, publish_year, category, available_quantity) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", b.ISBN, b.Title, b.Author, b.PublishYear, b.Category, b.AvailableQuantity).Scan(&b.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// FIX: Use id from URL path, not b.ID from request body
	_, err = db.Exec("UPDATE books SET isbn = $1, title = $2, author = $3, publish_year = $4, category = $5, available_quantity = $6 WHERE id = $7", b.ISBN, b.Title, b.Author, b.PublishYear, b.Category, b.AvailableQuantity, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getPaginationParams(r *http.Request) (page, limit int) {
	query := r.URL.Query()

	page, _ = strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ = strconv.Atoi(query.Get("limit"))
	if limit < 1 {
		limit = 10
	}

	return
}
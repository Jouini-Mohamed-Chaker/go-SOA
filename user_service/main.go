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
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type PaginatedResponse struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Total int    `json:"total"`
	Data  []User `json:"data"`
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

	log.Println("User Service connected to database")

	router := mux.NewRouter()
	router.HandleFunc("/api/users", getAllUsers).Methods("GET")
	router.HandleFunc("/api/users/{id}", getUserByID).Methods("GET")
	router.HandleFunc("/api/users", createUser).Methods("POST")
	router.HandleFunc("/api/users/{id}", updateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id}", deleteUser).Methods("DELETE")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Println("User Service running on port 8082")
	log.Fatal(http.ListenAndServe(":8082", handler))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	page, limit := getPaginationParams(r)
	offset := (page - 1) * limit

	// Get total count
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`
	SELECT id, username, email, first_name, last_name 
	FROM users
	ORDER BY id
	LIMIT $1 OFFSET $2`,
		limit, offset)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// Initialize empty slice if nil
	if users == nil {
		users = []User{}
	}

	response := PaginatedResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return // FIX: Added missing return
	}

	var u User
	err = db.QueryRow("SELECT id, username, email, first_name, last_name FROM users WHERE id = $1", id).Scan(&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName)

	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if u.Username == "" || u.Email == "" {
		http.Error(w, "Username and email are required fields", http.StatusBadRequest)
		return
	}

	err := db.QueryRow("INSERT INTO users (username, email, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id", u.Username, u.Email, u.FirstName, u.LastName).Scan(&u.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return // FIX: Added missing return
	}

	_, err = db.Exec(
		"UPDATE users SET username = $1, email = $2, first_name = $3, last_name = $4 WHERE id = $5",
		u.Username, u.Email, u.FirstName, u.LastName, id,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM users WHERE id = $1", id)
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
Here's the updated documentation with the Authentication Gateway Service added:

# SIMPLIFIED LIBRARY MANAGEMENT SYSTEM

## 1. Architecture Overview

```
┌─────────────────┐
│   Frontend      │
│   (Optional)    │
└────────┬────────┘
         │
    ┌────▼────────────┐
    │ Auth Gateway    │
    │   (Port 8080)   │
    └────────┬────────┘
             │
        ┌────┴────┐
        │         │
   ┌────▼────┐  ┌▼──────┐
   │  REST   │  │ SOAP  │
   │Services │  │Service│
   └────┬────┘  └┬──────┘
        │        │
        └────┬───┘
             │
        ┌────▼────┐
        │Database │
        └─────────┘
```

### Services and Ports
- **Auth Gateway Service (REST)** : Port 8080
- **Book Service (REST)** : Port 8081
- **User Service (REST)** : Port 8082
- **Loan Service (SOAP)** : Port 8083
- **PostgreSQL Database** : Port 5432

---

## 2. Auth Gateway Service (REST)

### 2.1 Overview
The Auth Gateway Service provides JWT-based authentication and acts as a proxy/gateway to all other services. All client requests should go through this gateway. It handles:
- User authentication and JWT token generation
- Token validation for protected endpoints
- Request forwarding to backend services (Book, User, Loan)

### 2.2 Authentication Flow
1. Client sends login request with credentials
2. Gateway validates credentials against User Service
3. Gateway generates JWT token (valid for 10 hours)
4. Client includes token in Authorization header for subsequent requests
5. Gateway validates token and forwards requests to backend services

### 2.3 User Credentials Model
```java
UserCredentials {
  userId: Long (references users.id)
  username: String (unique)
  passwordHash: String (BCrypt hashed)
}
```

### 2.4 REST Endpoints

#### Authentication Endpoints (Public - No JWT Required)

**POST /auth/login**
- Authenticate user and generate JWT token
- Request body:
```json
{
  "username": "alice",
  "password": "password123"
}
```
- Response 200 OK:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbGljZSIsImlhdCI6MTYzMjc1MDAwMCwiZXhwIjoxNjMyNzg2MDAwfQ.abc123def456",
  "username": "alice",
  "expiresIn": 36000
}
```
- Response 401 Unauthorized (invalid credentials):
```json
{
  "error": "Invalid username or password"
}
```

**POST /auth/register**
- Register new user with credentials
- Request body:
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "firstName": "New",
  "lastName": "User",
  "password": "securepassword"
}
```
- Response 201 Created:
```json
{
  "userId": 3,
  "username": "newuser",
  "email": "newuser@example.com",
  "firstName": "New",
  "lastName": "User"
}
```
- Response 400 Bad Request (username exists):
```json
{
  "error": "Username already exists"
}
```

**POST /auth/validate**
- Validate JWT token
- Request body:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
- Response 200 OK:
```json
{
  "valid": true,
  "username": "alice"
}
```
- Response 401 Unauthorized:
```json
{
  "valid": false,
  "error": "Invalid or expired token"
}
```

#### Protected Endpoints (JWT Required)

All requests to protected endpoints must include:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Missing or invalid token returns:
```json
{
  "error": "Unauthorized",
  "message": "Valid JWT token required"
}
```
Status: 401 Unauthorized

---

**Book Service Proxied Endpoints**

All Book Service endpoints are accessible through the gateway with `/api/books` prefix:

- GET /api/books?page={page}&limit={limit}
- GET /api/books/{id}
- POST /api/books
- PUT /api/books/{id}
- DELETE /api/books/{id}
- GET /api/books/search?title={title}&page={page}&limit={limit}

Request/Response formats remain identical to Book Service specification (see section 3.2).

---

**User Service Proxied Endpoints**

All User Service endpoints are accessible through the gateway with `/api/users` prefix:

- GET /api/users?page={page}&limit={limit}
- GET /api/users/{id}
- POST /api/users
- PUT /api/users/{id}
- DELETE /api/users/{id}

Request/Response formats remain identical to User Service specification (see section 4.2).

---

**Loan Service Proxied Endpoints**

SOAP operations converted to REST endpoints:

**POST /api/loans**
- Create new loan
- Request body:
```json
{
  "userId": 1,
  "bookId": 2
}
```
- Response 201 Created:
```json
{
  "id": 1,
  "userId": 1,
  "bookId": 2,
  "loanDate": "2024-01-15",
  "dueDate": "2024-01-29",
  "returnDate": null,
  "status": "ACTIVE"
}
```
- Response 400 Bad Request (book unavailable):
```json
{
  "error": "Book not available for loan"
}
```

**PUT /api/loans/{loanId}/return**
- Return a loaned book
- Request body: (empty or optional)
```json
{}
```
- Response 200 OK:
```json
{
  "id": 1,
  "userId": 1,
  "bookId": 2,
  "loanDate": "2024-01-15",
  "dueDate": "2024-01-29",
  "returnDate": "2024-01-20",
  "status": "RETURNED"
}
```
- Response 404 Not Found:
```json
{
  "error": "Loan not found"
}
```

**GET /api/loans/user/{userId}**
- Get all loans for a specific user
- Response 200 OK:
```json
[
  {
    "id": 1,
    "userId": 1,
    "bookId": 2,
    "loanDate": "2024-01-15",
    "dueDate": "2024-01-29",
    "returnDate": "2024-01-20",
    "status": "RETURNED"
  },
  {
    "id": 3,
    "userId": 1,
    "bookId": 5,
    "loanDate": "2024-01-22",
    "dueDate": "2024-02-05",
    "returnDate": null,
    "status": "ACTIVE"
  }
]
```

**GET /api/loans/{loanId}**
- Get specific loan by ID
- Response 200 OK: Loan object
- Response 404 Not Found

**GET /api/loans**
- Get all loans (admin function)
- Response 200 OK: Array of loan objects

### 2.5 JWT Token Specification

**Token Structure:**
- Algorithm: HS256 (HMAC with SHA-256)
- Expiration: 10 hours from issuance
- Payload claims:
  - `sub`: username
  - `iat`: issued at timestamp
  - `exp`: expiration timestamp

**Token Example (decoded):**
```json
{
  "sub": "alice",
  "iat": 1632750000,
  "exp": 1632786000
}
```

### 2.6 Security Configuration

- **Public endpoints:** `/auth/login`, `/auth/register`, `/auth/validate`
- **Protected endpoints:** All other endpoints require valid JWT
- **Session management:** Stateless (no server-side sessions)
- **Password storage:** BCrypt hashed with salt
- **Token secret:** 256-bit secret key (configurable via environment variable)

---

## 3. Book Service (REST)

### 3.1 Book Model
```java
Book {
  id: Long (auto-generated)
  isbn: String (unique)
  title: String
  author: String
  publishYear: Integer
  category: String
  availableQuantity: Integer
}
```

### 3.2 REST Endpoints

**Note:** In production, all requests should go through Auth Gateway (Port 8080). Direct access to Port 8081 should be blocked by firewall.

**GET /api/books?page={page}&limit={limit}**
- Returns a paginated list of books
- Query Params:
  - page (optional, default = 1)
  - limit (optional, default = 10)
- Response: List of books
```json
{
  "page": 1,
  "limit": 10,
  "data": [
    {
      "id": 1,
      "isbn": "9781234567890",
      "title": "Sample Book",
      "author": "John Doe",
      "publishYear": 2020,
      "category": "Fiction",
      "availableQuantity": 5
    }, 
    {...}, 
    ...
  ]
}
```

**GET /api/books/{id}**
- Get book by ID
- Response: Book object or 404

**POST /api/books**
- Create new book
- Request body:
```json
{
  "isbn": "9781234567890",
  "title": "New Book",
  "author": "Jane Doe",
  "publishYear": 2023,
  "category": "Science",
  "availableQuantity": 3
}
```
- Response: 201 Created with book object

**PUT /api/books/{id}**
- Update book
- Request body: Book object with changes
- Response: 200 OK

**DELETE /api/books/{id}**
- Delete book
- Response: 204 No Content

**GET /api/books/search?title={title}**
- Search books by title
- Response: List of matching books

**GET /api/books/search?title={title}&page={page}&limit={limit}**
- Search books by title substring (case-insensitive).
- Supports pagination.
- Query params:
  - title (required)
  - page (optional, default = 1)
  - limit (optional, default = 10)
- Response:
```json
{
  "page": 1,
  "limit": 10,
  "total": 3,
  "data": [
    {
      "id": 1,
      "isbn": "9781234567890",
      "title": "Sample Book",
      "author": "John Doe",
      "publishYear": 2020,
      "category": "Fiction",
      "availableQuantity": 5
    }, 
    {...},
    ...
  ]
}
```

---

## 4. User Service (REST)

### 4.1 User Model
```java
User {
  id: Long (auto-generated)
  username: String (unique)
  email: String
  firstName: String
  lastName: String
}
```

### 4.2 REST Endpoints

**Note:** In production, all requests should go through Auth Gateway (Port 8080). Direct access to Port 8082 should be blocked by firewall.

**GET /api/users?page={page}&limit={limit}**
- Returns paginated list of users.
- Query params:
  - page (optional, default = 1)
  - limit (optional, default = 10)
- Response:
```json
{
  "page": 1,
  "limit": 10,
  "total": 12,
  "data": [
    {
      "id": 1,
      "username": "johndoe",
      "email": "john@example.com",
      "firstName": "John",
      "lastName": "Doe"
    }, 
    {...}
    ...
  ]
}
```

**GET /api/users/{id}**
- Get user by ID
- Response: User object or 404

**POST /api/users**
- Create new user
- Request body:
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "firstName": "John",
  "lastName": "Doe"
}
```
- Response: 201 Created

**PUT /api/users/{id}**
- Update user
- Response: 200 OK

**DELETE /api/users/{id}**
- Delete user
- Response: 204 No Content

---

## 5. Loan Service (SOAP)

### 5.1 Loan Model
```java
Loan {
  id: Long (auto-generated)
  userId: Long
  bookId: Long
  loanDate: Date
  dueDate: Date (loanDate + 14 days)
  returnDate: Date (nullable)
  status: String (ACTIVE, RETURNED)
}
```

### 5.2 SOAP Operations

**Note:** In production, SOAP service should only be accessible by Auth Gateway. Direct access to Port 8083 should be blocked by firewall.

**createLoan**
- Input: userId, bookId
- Process:
  1. Check if book exists (call Book Service)
  2. Check if book is available (availableQuantity > 0)
  3. Create loan with status ACTIVE
  4. Decrease book's availableQuantity by 1
- Output: Loan object

**returnLoan**
- Input: loanId
- Process:
  1. Find loan by ID
  2. Set returnDate to current date
  3. Set status to RETURNED
  4. Increase book's availableQuantity by 1
- Output: Updated loan object

**getLoansByUser**
- Input: userId
- Output: List of loans for that user

**getLoanById**
- Input: loanId
- Output: Loan object

**getAllLoans**
- Output: List of all loans

---

## 6. Database Schema

### 6.1 Tables

**books**
```sql
CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    isbn VARCHAR(20) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    author VARCHAR(100) NOT NULL,
    publish_year INTEGER,
    category VARCHAR(50),
    available_quantity INTEGER DEFAULT 0
);
```

**users**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50)
);
```

**user_credentials**
```sql
CREATE TABLE user_credentials (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**loans**
```sql
CREATE TABLE loans (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    book_id INTEGER REFERENCES books(id),
    loan_date DATE NOT NULL,
    due_date DATE NOT NULL,
    return_date DATE,
    status VARCHAR(20) NOT NULL
);
```

### 6.2 Sample Data

**Books:**
```sql
INSERT INTO books (isbn, title, author, publish_year, category, available_quantity) VALUES
('9781234567890', 'Introduction to Java', 'James Gosling', 2020, 'Programming', 5),
('9780987654321', 'Web Services Guide', 'Mary Smith', 2021, 'Technology', 3),
('9781111111111', 'Database Design', 'Peter Chen', 2019, 'Database', 4);
```

**Users:**
```sql
INSERT INTO users (username, email, first_name, last_name) VALUES
('alice', 'alice@example.com', 'Alice', 'Johnson'),
('bob', 'bob@example.com', 'Bob', 'Smith');
```

**User Credentials:**
```sql
-- Password for both users is 'password123' (BCrypt hashed)
INSERT INTO user_credentials (user_id, username, password_hash) VALUES
(1, 'alice', '$2a$10$N9qo8uLOickgx2ZMRZoMye8vKOhJTOFT.FV8VhN8ihJJ6FqPJq3uC'),
(2, 'bob', '$2a$10$N9qo8uLOickgx2ZMRZoMye8vKOhJTOFT.FV8VhN8ihJJ6FqPJq3uC');
```

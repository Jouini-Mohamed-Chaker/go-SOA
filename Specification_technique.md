# SIMPLIFIED LIBRARY MANAGEMENT SYSTEM

## 1. Architecture Overview

```
┌─────────────────┐
│   Frontend      │
│   (Optional)    │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼────┐  ┌▼──────┐
│  REST  │  │ SOAP  │
│Services│  │Service│
└───┬────┘  └┬──────┘
    │        │
    └────┬───┘
         │
    ┌────▼────┐
    │Database │
    └─────────┘
```

### Services and Ports
- **Book Service (REST)** : Port 8081
- **User Service (REST)** : Port 8082
- **Loan Service (SOAP)** : Port 8083
- **PostgreSQL Database** : Port 5432

---

## 2. Book Service (REST)

### 2.1 Book Model
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

### 2.2 REST Endpoints

**GET /api/books**
- Get all books
- Response: List of books
```json
[
  {
    "id": 1,
    "isbn": "9781234567890",
    "title": "Sample Book",
    "author": "John Doe",
    "publishYear": 2020,
    "category": "Fiction",
    "availableQuantity": 5
  }
]
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

---

## 3. User Service (REST)

### 3.1 User Model
```java
User {
  id: Long (auto-generated)
  username: String (unique)
  email: String
  firstName: String
  lastName: String
}
```

### 3.2 REST Endpoints

**GET /api/users**
- Get all users
- Response: List of users

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

## 4. Loan Service (SOAP)

### 4.1 Loan Model
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

### 4.2 SOAP Operations

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

## 5. Database Schema

### 5.1 Tables

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

### 5.2 Sample Data

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


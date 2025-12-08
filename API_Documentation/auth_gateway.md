# Auth Gateway API - Essential Docs

## Base URL
`http://localhost:8080`

## Authentication
Protected endpoints require:
```
Authorization: Bearer <jwt_token>
```

## Public Endpoints (No Auth)

### 1. POST `/auth/login` - Get JWT token
**Request:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:** `200 OK`
```json
{
  "token": "jwt_token_string",
  "username": "string",
  "expiresIn": 36000
}
```

### 2. POST `/auth/register` - Create new user
**Request:**
```json
{
  "username": "string",
  "email": "string",
  "firstName": "string",
  "lastName": "string",
  "password": "string"
}
```

**Response:** `201 Created`
```json
{
  "id": 0,
  "username": "string",
  "email": "string",
  "firstName": "string",
  "lastName": "string"
}
```

### 3. POST `/auth/validate` - Check token validity
**Request:**
```json
{
  "token": "string"
}
```

**Response:** `200 OK`
```json
{
  "valid": true,
  "username": "string"
}
```

---

## Protected Endpoints (Require `Authorization: Bearer <token>`)

### Books Proxy
All endpoints from Book Service available at `/api/books/*`
- `GET /api/books` - Get all books
- `GET /api/books/search?title={title}` - Search books
- `GET /api/books/{id}` - Get book by ID
- `POST /api/books` - Create book
- `PUT /api/books/{id}` - Update book
- `DELETE /api/books/{id}` - Delete book

### Users Proxy
All endpoints from User Service available at `/api/users/*`
- `GET /api/users` - Get all users
- `GET /api/users/{id}` - Get user by ID
- `POST /api/users` - Create user
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user

### Loans (REST Wrapper)

#### 1. POST `/api/loans` - Create loan
**Request:**
```json
{
  "userId": 0,
  "bookId": 0
}
```

**Response:** `201 Created`
```json
{
  "id": 0,
  "userId": 0,
  "bookId": 0,
  "loanDate": "2024-01-15T10:30:00Z",
  "dueDate": "2024-01-29T10:30:00Z",
  "returnDate": null,
  "status": "ACTIVE"
}
```

#### 2. PUT `/api/loans/{id}/return` - Return loan
**Response:** `200 OK` with updated loan (status: RETURNED)

#### 3. GET `/api/loans/user/{userId}` - Get user's loans
**Response:** `200 OK` with array of loans

#### 4. GET `/api/loans/{id}` - Get loan by ID
**Response:** `200 OK` with loan object

#### 5. GET `/api/loans` - Get all loans
**Response:** `200 OK` with array of all loans

---

## Quick Examples

### Register & Login
```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "email": "user@test.com", "password": "pass123"}'

# Login (get token)
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "password": "pass123"}'
# Save the token from response
```

### Use Protected Endpoints
```bash
# Create loan (with token from login)
curl -X POST http://localhost:8080/api/loans \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"userId": 1, "bookId": 5}'

# Search books
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  "http://localhost:8080/api/books/search?title=harry"

# Get user loans
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  "http://localhost:8080/api/loans/user/1"
```

### Return Loan
```bash
curl -X PUT http://localhost:8080/api/loans/15/return \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## Notes
- Token expires in 10 hours
- Use the same token for all protected endpoints
- Books/Users endpoints mirror the original services exactly
- Loans endpoints provide REST interface to SOAP service
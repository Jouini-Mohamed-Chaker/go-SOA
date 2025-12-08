# Book Service API - Essential Docs

## Base URL
`http://localhost:8081/api/books`

## Data Models

### Book Object
```json
{
  "id": 0,
  "isbn": "string",
  "title": "string",
  "author": "string",
  "publishYear": 0,
  "category": "string",
  "availableQuantity": 0
}
```

### Paginated Response
```json
{
  "page": 0,
  "limit": 0,
  "total": 0,      // sometimes included, sometimes not
  "data": [Book, Book, ...]
}
```

---

## Endpoints

### 1. GET `/api/books` - Get all books
**Query Params:**
- `page` (optional, default: 1)
- `limit` (optional, default: 10)

**Response:** `200 OK` with `PaginatedResponse` (no `total` field)

---

### 2. GET `/api/books/search` - Search by title
**Query Params:**
- `title` (required) - partial match, case-insensitive
- `page` (optional, default: 1)
- `limit` (optional, default: 10)

**Response:** `200 OK` with `PaginatedResponse` (includes `total`)

---

### 3. GET `/api/books/{id}` - Get book by ID
**Path Param:** `id` (integer)

**Response:** `200 OK` with `Book` object  
**Errors:** `400` (bad ID), `404` (not found)

---

### 4. POST `/api/books` - Create book
**Request Body:**
```json
{
  "isbn": "string",        // required
  "title": "string",       // required
  "author": "string",      // required
  "publishYear": 0,        // optional
  "category": "string",    // optional
  "availableQuantity": 0   // optional
}
```

**Response:** `201 Created` with created `Book` object (includes generated ID)

---

### 5. PUT `/api/books/{id}` - Update book
**Path Param:** `id` (integer)

**Request Body:** (same structure as POST, ID in body is ignored)

**Response:** `200 OK` with updated `Book` object

---

### 6. DELETE `/api/books/{id}` - Delete book
**Path Param:** `id` (integer)

**Response:** `204 No Content`

---

## Quick Examples

### Create Book
```bash
curl -X POST http://localhost:8081/api/books \
  -H "Content-Type: application/json" \
  -d '{
    "isbn": "978-1234567890",
    "title": "Book Title",
    "author": "Author Name"
  }'
```

### Search Books
```bash
curl "http://localhost:8081/api/books/search?title=harry&page=1&limit=10"
```

### Get All Books
```bash
curl "http://localhost:8081/api/books?page=1&limit=20"
```

### Get Book
```bash
curl "http://localhost:8081/api/books/123"
```

### Update Book (e.g., after loan)
```bash
curl -X PUT http://localhost:8081/api/books/123 \
  -H "Content-Type: application/json" \
  -d '{
    "isbn": "978-1234567890",
    "title": "Book Title",
    "author": "Author Name",
    "availableQuantity": 5  // Updated quantity
  }'
```

### Delete Book
```bash
curl -X DELETE "http://localhost:8081/api/books/123"
```

---

## Important Notes
- Required fields: `isbn`, `title`, `author`
- `availableQuantity` is used by Loan Service
- Search is partial match (LIKE '%search%')
- All endpoints return JSON
- No auth required
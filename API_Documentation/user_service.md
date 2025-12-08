# User Service API - Essential Docs

## Base URL
`http://localhost:8082/api/users`

## Data Models

### User Object
```json
{
  "id": 0,
  "username": "string",
  "email": "string",
  "firstName": "string",
  "lastName": "string"
}
```

### Paginated Response
```json
{
  "page": 0,
  "limit": 0,
  "total": 0,
  "data": [User, User, ...]
}
```

---

## Endpoints

### 1. GET `/api/users` - Get all users (paginated)
**Query Params:**
- `page` (optional, default: 1)
- `limit` (optional, default: 10)

**Response:** `200 OK` with `PaginatedResponse`

---

### 2. GET `/api/users/{id}` - Get user by ID
**Path Param:** `id` (integer)

**Response:** `200 OK` with `User` object  
**Errors:** `400` (bad ID), `404` (not found)

---

### 3. POST `/api/users` - Create user
**Request Body:**
```json
{
  "username": "string",  // required
  "email": "string",     // required
  "firstName": "string", // optional
  "lastName": "string"   // optional
}
```

**Response:** `201 Created` with created `User` object (includes generated ID)

---

### 4. PUT `/api/users/{id}` - Update user
**Path Param:** `id` (integer)

**Request Body:** (same structure as POST)

**Response:** `200 OK` with updated `User` object

---

### 5. DELETE `/api/users/{id}` - Delete user
**Path Param:** `id` (integer)

**Response:** `204 No Content`

---

## Quick Examples

### Create User
```bash
curl -X POST http://localhost:8082/api/users \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "email": "john@example.com"}'
```

### Get All Users
```bash
curl "http://localhost:8082/api/users?page=1&limit=20"
```

### Get User
```bash
curl "http://localhost:8082/api/users/123"
```

### Update User
```bash
curl -X PUT http://localhost:8082/api/users/123 \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "email": "new@example.com", "firstName": "John"}'
```

### Delete User
```bash
curl -X DELETE "http://localhost:8082/api/users/123"
```

---

## Notes
- All requests/responses use `application/json`
- `username` and `email` must be unique
- No auth required (open API)
- Errors return plain text messages
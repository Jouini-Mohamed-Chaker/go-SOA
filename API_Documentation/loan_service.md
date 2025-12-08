# Loan Service API - Essential Docs

## Base URL
`http://localhost:8083/ws` (or `/loan`)

## Protocol
SOAP 1.1 - Use `Content-Type: text/xml; charset=utf-8`

## Data Models

### Loan Object
```xml
<loanType>
  <id>integer</id>
  <userId>integer</userId>
  <bookId>integer</bookId>
  <loanDate>dateTime</loanDate>
  <dueDate>dateTime</dueDate>
  <returnDate>dateTime</returnDate>  <!-- optional -->
  <status>ACTIVE or RETURNED</status>
</loanType>
```

## SOAP Operations

### 1. createLoan - Create new loan
**SOAP Action:** `createLoan`

**Request:**
```xml
<soap:Envelope>
  <soap:Body>
    <createLoanRequest>
      <userId>integer</userId>
      <bookId>integer</bookId>
    </createLoanRequest>
  </soap:Body>
</soap:Envelope>
```

**Response:**
```xml
<soap:Envelope>
  <soap:Body>
    <createLoanResponse>
      <loan>...</loan>
      <error>string</error>  <!-- if error occurs -->
    </createLoanResponse>
  </soap:Body>
</soap:Envelope>
```

**Notes:**
- Book must be available (quantity > 0)
- Auto-sets due date = loan date + 14 days
- Decreases book quantity by 1

---

### 2. returnLoan - Return a loan
**SOAP Action:** `returnLoan`

**Request:**
```xml
<soap:Envelope>
  <soap:Body>
    <returnLoanRequest>
      <loanId>integer</loanId>
    </returnLoanRequest>
  </soap:Body>
</soap:Envelope>
```

**Response:**
```xml
<soap:Envelope>
  <soap:Body>
    <returnLoanResponse>
      <loan>...</loan>
      <error>string</error>
    </returnLoanResponse>
  </soap:Body>
</soap:Envelope>
```

---

### 3. getLoansByUser - Get user's loans
**SOAP Action:** `getLoansByUser`

**Request:**
```xml
<soap:Envelope>
  <soap:Body>
    <getLoansByUserRequest>
      <userId>integer</userId>
    </getLoansByUserRequest>
  </soap:Body>
</soap:Envelope>
```

**Response:** Returns all loans for user

---

### 4. getLoanById - Get specific loan
**SOAP Action:** `getLoanById`

**Request:**
```xml
<soap:Envelope>
  <soap:Body>
    <getLoanByIdRequest>
      <loanId>integer</loanId>
    </getLoanByIdRequest>
  </soap:Body>
</soap:Envelope>
```

---

### 5. getAllLoans - Get all loans
**SOAP Action:** `getAllLoans`

**Request:**
```xml
<soap:Envelope>
  <soap:Body>
    <getAllLoansRequest/>
  </soap:Body>
</soap:Envelope>
```

## Quick Examples

### Create Loan
```xml
<soap:Envelope>
  <soap:Body>
    <createLoanRequest>
      <userId>123</userId>
      <bookId>456</bookId>
    </createLoanRequest>
  </soap:Body>
</soap:Envelope>
```

### Return Loan
```xml
<soap:Envelope>
  <soap:Body>
    <returnLoanRequest>
      <loanId>789</loanId>
    </returnLoanRequest>
  </soap:Body>
</soap:Envelope>
```

### Get User Loans
```xml
<soap:Envelope>
  <soap:Body>
    <getLoansByUserRequest>
      <userId>123</userId>
    </getLoansByUserRequest>
  </soap:Body>
</soap:Envelope>
```

## Important Notes
- Loans are for 14 days (auto-calculated)
- Book availability is checked/updated automatically
- Status: `ACTIVE` or `RETURNED`
- Uses SOAP, not REST - send XML requests
- WSDL available at `/ws` or `/loan`
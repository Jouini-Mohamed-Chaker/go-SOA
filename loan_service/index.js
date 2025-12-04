const express = require('express');
const { Pool } = require('pg');
const axios = require('axios');
const cors = require('cors');
const bodyParser = require('body-parser');

const app = express();
const port = 8083;

// CRITICAL: Apply CORS before any other middleware
app.use(cors({
  origin: '*',
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'SOAPAction', 'soapaction'],
  credentials: true
}));

// Handle preflight
app.options('*', cors());

// Parse text/xml as text
app.use(bodyParser.text({ type: 'text/xml' }));

// Database connection
const pool = new Pool({
  host: process.env.DB_HOST || 'localhost',
  port: process.env.DB_PORT || 5432,
  user: process.env.DB_USER || 'postgres',
  password: process.env.DB_PASSWORD || 'postgres',
  database: process.env.DB_NAME || 'library',
});

// SOAP request handler
app.post('/loan', async (req, res) => {
  // Set CORS headers explicitly
  res.set({
    'Access-Control-Allow-Origin': '*',
    'Content-Type': 'text/xml; charset=utf-8'
  });

  const soapBody = req.body;
  
  try {
    let responseXml = '';

    // Parse which operation is being called
    if (soapBody.includes('createLoan')) {
      const userId = extractValue(soapBody, 'userId');
      const bookId = extractValue(soapBody, 'bookId');
      const result = await createLoan(userId, bookId);
      responseXml = buildCreateLoanResponse(result);
    } 
    else if (soapBody.includes('returnLoan')) {
      const loanId = extractValue(soapBody, 'loanId');
      const result = await returnLoan(loanId);
      responseXml = buildReturnLoanResponse(result);
    }
    else if (soapBody.includes('getLoansByUser')) {
      const userId = extractValue(soapBody, 'userId');
      const result = await getLoansByUser(userId);
      responseXml = buildGetLoansByUserResponse(result);
    }
    else if (soapBody.includes('getLoanById')) {
      const loanId = extractValue(soapBody, 'loanId');
      const result = await getLoanById(loanId);
      responseXml = buildGetLoanByIdResponse(result);
    }
    else if (soapBody.includes('getAllLoans')) {
      const result = await getAllLoans();
      responseXml = buildGetAllLoansResponse(result);
    }
    
    res.send(responseXml);
  } catch (error) {
    console.error('SOAP Error:', error);
    res.status(500).send(buildErrorResponse(error.message));
  }
});

// Serve WSDL
app.get('/loan', (req, res) => {
  res.set('Content-Type', 'text/xml');
  const fs = require('fs');
  const wsdl = fs.readFileSync('./loan.wsdl', 'utf8');
  res.send(wsdl);
});

// Helper function to extract values from SOAP XML
function extractValue(xml, tagName) {
  const regex = new RegExp(`<${tagName}>(.*?)</${tagName}>`, 's');
  const match = xml.match(regex);
  return match ? match[1] : null;
}

// Business Logic Functions
async function createLoan(userId, bookId) {
  try {
    const bookResponse = await axios.get(`http://book_service:8081/api/books/${bookId}`);
    const book = bookResponse.data;

    if (book.availableQuantity <= 0) {
      return {
        success: false,
        message: 'Book is not available',
        loan: null
      };
    }

    const loanDate = new Date();
    const dueDate = new Date(loanDate);
    dueDate.setDate(dueDate.getDate() + 14);

    const result = await pool.query(
      `INSERT INTO loans (user_id, book_id, loan_date, due_date, status) 
       VALUES ($1, $2, $3, $4, $5) RETURNING *`,
      [userId, bookId, loanDate, dueDate, 'ACTIVE']
    );

    await axios.put(`http://book_service:8081/api/books/${bookId}`, {
      ...book,
      availableQuantity: book.availableQuantity - 1
    });

    const loan = result.rows[0];
    return {
      success: true,
      message: 'Loan created successfully',
      loan: {
        id: loan.id,
        userId: loan.user_id,
        bookId: loan.book_id,
        loanDate: loan.loan_date.toISOString(),
        dueDate: loan.due_date.toISOString(),
        returnDate: loan.return_date ? loan.return_date.toISOString() : null,
        status: loan.status
      }
    };
  } catch (error) {
    return {
      success: false,
      message: error.message,
      loan: null
    };
  }
}

async function returnLoan(loanId) {
  try {
    const loanResult = await pool.query('SELECT * FROM loans WHERE id = $1', [loanId]);

    if (loanResult.rows.length === 0) {
      return { success: false, message: 'Loan not found', loan: null };
    }

    const loan = loanResult.rows[0];

    if (loan.status === 'RETURNED') {
      return { success: false, message: 'Loan already returned', loan: null };
    }

    const returnDate = new Date();
    const result = await pool.query(
      `UPDATE loans SET return_date = $1, status = $2 WHERE id = $3 RETURNING *`,
      [returnDate, 'RETURNED', loanId]
    );

    const bookResponse = await axios.get(`http://book_service:8081/api/books/${loan.book_id}`);
    const book = bookResponse.data;

    await axios.put(`http://book_service:8081/api/books/${loan.book_id}`, {
      ...book,
      availableQuantity: book.availableQuantity + 1
    });

    const updatedLoan = result.rows[0];
    return {
      success: true,
      message: 'Loan returned successfully',
      loan: {
        id: updatedLoan.id,
        userId: updatedLoan.user_id,
        bookId: updatedLoan.book_id,
        loanDate: updatedLoan.loan_date.toISOString(),
        dueDate: updatedLoan.due_date.toISOString(),
        returnDate: updatedLoan.return_date.toISOString(),
        status: updatedLoan.status
      }
    };
  } catch (error) {
    return { success: false, message: error.message, loan: null };
  }
}

async function getLoansByUser(userId) {
  try {
    const result = await pool.query('SELECT * FROM loans WHERE user_id = $1', [userId]);
    const loans = result.rows.map(loan => ({
      id: loan.id,
      userId: loan.user_id,
      bookId: loan.book_id,
      loanDate: loan.loan_date.toISOString(),
      dueDate: loan.due_date.toISOString(),
      returnDate: loan.return_date ? loan.return_date.toISOString() : null,
      status: loan.status
    }));
    return { success: true, loans };
  } catch (error) {
    return { success: false, loans: [] };
  }
}

async function getLoanById(loanId) {
  try {
    const result = await pool.query('SELECT * FROM loans WHERE id = $1', [loanId]);
    if (result.rows.length === 0) {
      return { success: false, message: 'Loan not found', loan: null };
    }
    const loan = result.rows[0];
    return {
      success: true,
      message: 'Loan found',
      loan: {
        id: loan.id,
        userId: loan.user_id,
        bookId: loan.book_id,
        loanDate: loan.loan_date.toISOString(),
        dueDate: loan.due_date.toISOString(),
        returnDate: loan.return_date ? loan.return_date.toISOString() : null,
        status: loan.status
      }
    };
  } catch (error) {
    return { success: false, message: error.message, loan: null };
  }
}

async function getAllLoans() {
  try {
    const result = await pool.query('SELECT * FROM loans ORDER BY loan_date DESC');
    const loans = result.rows.map(loan => ({
      id: loan.id,
      userId: loan.user_id,
      bookId: loan.book_id,
      loanDate: loan.loan_date.toISOString(),
      dueDate: loan.due_date.toISOString(),
      returnDate: loan.return_date ? loan.return_date.toISOString() : null,
      status: loan.status
    }));
    return { success: true, loans };
  } catch (error) {
    return { success: false, loans: [] };
  }
}

// Response Builders
function buildCreateLoanResponse(result) {
  const loanXml = result.loan ? `
    <loan>
      <id>${result.loan.id}</id>
      <userId>${result.loan.userId}</userId>
      <bookId>${result.loan.bookId}</bookId>
      <loanDate>${result.loan.loanDate}</loanDate>
      <dueDate>${result.loan.dueDate}</dueDate>
      <returnDate>${result.loan.returnDate || ''}</returnDate>
      <status>${result.loan.status}</status>
    </loan>` : '';

  return `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <createLoanResponse xmlns="http://loan.service">
      <success>${result.success}</success>
      <message>${result.message}</message>
      ${loanXml}
    </createLoanResponse>
  </soap:Body>
</soap:Envelope>`;
}

function buildReturnLoanResponse(result) {
  const loanXml = result.loan ? `
    <loan>
      <id>${result.loan.id}</id>
      <userId>${result.loan.userId}</userId>
      <bookId>${result.loan.bookId}</bookId>
      <loanDate>${result.loan.loanDate}</loanDate>
      <dueDate>${result.loan.dueDate}</dueDate>
      <returnDate>${result.loan.returnDate}</returnDate>
      <status>${result.loan.status}</status>
    </loan>` : '';

  return `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <returnLoanResponse xmlns="http://loan.service">
      <success>${result.success}</success>
      <message>${result.message}</message>
      ${loanXml}
    </returnLoanResponse>
  </soap:Body>
</soap:Envelope>`;
}

function buildGetAllLoansResponse(result) {
  const loansXml = result.loans.map(loan => `
    <loan>
      <id>${loan.id}</id>
      <userId>${loan.userId}</userId>
      <bookId>${loan.bookId}</bookId>
      <loanDate>${loan.loanDate}</loanDate>
      <dueDate>${loan.dueDate}</dueDate>
      <returnDate>${loan.returnDate || ''}</returnDate>
      <status>${loan.status}</status>
    </loan>`).join('');

  return `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getAllLoansResponse xmlns="http://loan.service">
      <success>${result.success}</success>
      ${loansXml}
    </getAllLoansResponse>
  </soap:Body>
</soap:Envelope>`;
}

function buildGetLoansByUserResponse(result) {
  const loansXml = result.loans.map(loan => `
    <loan>
      <id>${loan.id}</id>
      <userId>${loan.userId}</userId>
      <bookId>${loan.bookId}</bookId>
      <loanDate>${loan.loanDate}</loanDate>
      <dueDate>${loan.dueDate}</dueDate>
      <returnDate>${loan.returnDate || ''}</returnDate>
      <status>${loan.status}</status>
    </loan>`).join('');

  return `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getLoansByUserResponse xmlns="http://loan.service">
      <success>${result.success}</success>
      ${loansXml}
    </getLoansByUserResponse>
  </soap:Body>
</soap:Envelope>`;
}

function buildGetLoanByIdResponse(result) {
  const loanXml = result.loan ? `
    <loan>
      <id>${result.loan.id}</id>
      <userId>${result.loan.userId}</userId>
      <bookId>${result.loan.bookId}</bookId>
      <loanDate>${result.loan.loanDate}</loanDate>
      <dueDate>${result.loan.dueDate}</dueDate>
      <returnDate>${result.loan.returnDate || ''}</returnDate>
      <status>${result.loan.status}</status>
    </loan>` : '';

  return `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getLoanByIdResponse xmlns="http://loan.service">
      <success>${result.success}</success>
      <message>${result.message}</message>
      ${loanXml}
    </getLoanByIdResponse>
  </soap:Body>
</soap:Envelope>`;
}

function buildErrorResponse(message) {
  return `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <soap:Fault>
      <faultcode>soap:Server</faultcode>
      <faultstring>${message}</faultstring>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`;
}

app.listen(port, () => {
  console.log(`Loan Service (Manual SOAP) listening on port ${port}`);
  console.log(`SOAP endpoint available at http://localhost:${port}/loan`);
});
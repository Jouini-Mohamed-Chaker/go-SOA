const soap = require('soap');
const express = require('express');
const { Pool } = require('pg');
const axios = require('axios');

const app = express();
const port = 8083;

// Database connection
const pool = new Pool({
  host: process.env.DB_HOST || 'localhost',
  port: process.env.DB_PORT || 5432,
  user: process.env.DB_USER || 'postgres',
  password: process.env.DB_PASSWORD || 'postgres',
  database: process.env.DB_NAME || 'library',
});

// SOAP Service Implementation
const loanService = {
  LoanService: {
    LoanServicePort: {
      // Create a new loan
      createLoan: async function(args) {
        const { userId, bookId } = args;
        
        try {
          // Check if book exists and is available
          const bookResponse = await axios.get(`http://book_service:8081/api/books/${bookId}`);
          const book = bookResponse.data;
          
          if (book.availableQuantity <= 0) {
            return {
              success: false,
              message: 'Book is not available',
              loan: null
            };
          }
          
          // Create loan
          const loanDate = new Date();
          const dueDate = new Date(loanDate);
          dueDate.setDate(dueDate.getDate() + 14); // 14 days loan period
          
          const result = await pool.query(
            `INSERT INTO loans (user_id, book_id, loan_date, due_date, status) 
             VALUES ($1, $2, $3, $4, $5) RETURNING *`,
            [userId, bookId, loanDate, dueDate, 'ACTIVE']
          );
          
          // Decrease available quantity
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
          console.error('Error creating loan:', error);
          return {
            success: false,
            message: error.message,
            loan: null
          };
        }
      },
      
      // Return a loan
      returnLoan: async function(args) {
        const { loanId } = args;
        
        try {
          // Get loan details
          const loanResult = await pool.query('SELECT * FROM loans WHERE id = $1', [loanId]);
          
          if (loanResult.rows.length === 0) {
            return {
              success: false,
              message: 'Loan not found',
              loan: null
            };
          }
          
          const loan = loanResult.rows[0];
          
          if (loan.status === 'RETURNED') {
            return {
              success: false,
              message: 'Loan already returned',
              loan: null
            };
          }
          
          // Update loan
          const returnDate = new Date();
          const result = await pool.query(
            `UPDATE loans SET return_date = $1, status = $2 WHERE id = $3 RETURNING *`,
            [returnDate, 'RETURNED', loanId]
          );
          
          // Increase available quantity
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
          console.error('Error returning loan:', error);
          return {
            success: false,
            message: error.message,
            loan: null
          };
        }
      },
      
      // Get loans by user
      getLoansByUser: async function(args) {
        const { userId } = args;
        
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
          
          return {
            success: true,
            loans: loans
          };
        } catch (error) {
          console.error('Error getting loans:', error);
          return {
            success: false,
            loans: []
          };
        }
      },
      
      // Get loan by ID
      getLoanById: async function(args) {
        const { loanId } = args;
        
        try {
          const result = await pool.query('SELECT * FROM loans WHERE id = $1', [loanId]);
          
          if (result.rows.length === 0) {
            return {
              success: false,
              message: 'Loan not found',
              loan: null
            };
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
          console.error('Error getting loan:', error);
          return {
            success: false,
            message: error.message,
            loan: null
          };
        }
      },
      
      // Get all loans
      getAllLoans: async function(args) {
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
          
          return {
            success: true,
            loans: loans
          };
        } catch (error) {
          console.error('Error getting all loans:', error);
          return {
            success: false,
            loans: []
          };
        }
      }
    }
  }
};

// Start server
app.listen(port, async () => {
  console.log(`Loan Service listening on port ${port}`);
  
  // Load WSDL and create SOAP service
  const xml = require('fs').readFileSync('./loan.wsdl', 'utf8');
  soap.listen(app, '/loan', loanService, xml);
  
  console.log(`SOAP service available at http://localhost:${port}/loan?wsdl`);
});
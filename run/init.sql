-- Database Initialization Script for Library Management System
-- This script creates all necessary tables and inserts sample data

-- Drop tables if they exist (for clean re-initialization)
DROP TABLE IF EXISTS loans CASCADE;
DROP TABLE IF EXISTS user_credentials CASCADE;
DROP TABLE IF EXISTS books CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Create Users Table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50)
);

-- Create Books Table
CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    isbn VARCHAR(20) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    author VARCHAR(100) NOT NULL,
    publish_year INTEGER,
    category VARCHAR(50),
    available_quantity INTEGER DEFAULT 0
);

-- Create User Credentials Table
CREATE TABLE user_credentials (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Loans Table
CREATE TABLE loans (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    loan_date DATE NOT NULL,
    due_date DATE NOT NULL,
    return_date DATE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'RETURNED'))
);

-- Create Indexes for Better Performance
CREATE INDEX idx_books_title ON books(title);
CREATE INDEX idx_books_isbn ON books(isbn);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_loans_user_id ON loans(user_id);
CREATE INDEX idx_loans_book_id ON loans(book_id);
CREATE INDEX idx_loans_status ON loans(status);

-- Insert Sample Users
INSERT INTO users (username, email, first_name, last_name) VALUES
('alice', 'alice@example.com', 'Alice', 'Johnson'),
('bob', 'bob@example.com', 'Bob', 'Smith'),
('charlie', 'charlie@example.com', 'Charlie', 'Brown'),
('david', 'david@example.com', 'David', 'Wilson'),
('emma', 'emma@example.com', 'Emma', 'Davis');

-- Insert Sample Books
INSERT INTO books (isbn, title, author, publish_year, category, available_quantity) VALUES
('9781234567890', 'Introduction to Java', 'James Gosling', 2020, 'Programming', 5),
('9780987654321', 'Web Services Guide', 'Mary Smith', 2021, 'Technology', 3),
('9781111111111', 'Database Design', 'Peter Chen', 2019, 'Database', 4),
('9782222222222', 'Advanced Go Programming', 'Rob Pike', 2023, 'Programming', 2),
('9783333333333', 'RESTful API Design', 'Leonard Richardson', 2022, 'Technology', 6),
('9784444444444', 'Microservices Architecture', 'Sam Newman', 2021, 'Architecture', 3),
('9785555555555', 'Clean Code', 'Robert Martin', 2020, 'Programming', 8),
('9786666666666', 'Design Patterns', 'Gang of Four', 2019, 'Programming', 4);

-- Insert Sample Loans (some active, some returned)
INSERT INTO loans (user_id, book_id, loan_date, due_date, return_date, status) VALUES
(1, 1, '2024-11-01', '2024-11-15', '2024-11-14', 'RETURNED'),
(2, 2, '2024-11-10', '2024-11-24', NULL, 'ACTIVE'),
(1, 3, '2024-11-15', '2024-11-29', NULL, 'ACTIVE'),
(3, 4, '2024-10-20', '2024-11-03', '2024-11-02', 'RETURNED'),
(4, 5, '2024-11-20', '2024-12-04', NULL, 'ACTIVE');

-- Display summary
SELECT 'Database initialized successfully!' AS message;
SELECT COUNT(*) AS total_users FROM users;
SELECT COUNT(*) AS total_books FROM books;
SELECT COUNT(*) AS total_loans FROM loans;
SELECT COUNT(*) AS active_loans FROM loans WHERE status = 'ACTIVE';
SELECT COUNT(*) AS users_with_credentials FROM user_credentials;
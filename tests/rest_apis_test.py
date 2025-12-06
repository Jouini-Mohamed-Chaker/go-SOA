#!/usr/bin/env python3
"""
REST Services Test Script
Tests Book Service and User Service endpoints
"""

import requests
import json
from colorama import Fore, Back, Style, init
from datetime import datetime
import sys

# Initialize colorama
init(autoreset=True)

# Configuration
BOOK_SERVICE = "http://localhost:8081"
USER_SERVICE = "http://localhost:8082"

# Test counters
total_tests = 0
passed_tests = 0
failed_tests = 0

# Storage for dynamic IDs
created_book_ids = []
created_user_ids = []

# Print functions
def print_header(text):
    print(f"\n{Fore.CYAN}╔════════════════════════════════════════════════════════════╗{Style.RESET_ALL}")
    print(f"{Fore.CYAN}║{Fore.WHITE}{Style.BRIGHT}  {text:<56}{Fore.CYAN}║{Style.RESET_ALL}")
    print(f"{Fore.CYAN}╚════════════════════════════════════════════════════════════╝{Style.RESET_ALL}\n")

def print_section(text):
    print(f"\n{Fore.MAGENTA}▶ {text}{Style.RESET_ALL}")
    print(f"{Fore.MAGENTA}{'━' * 60}{Style.RESET_ALL}")

def print_test(text):
    print(f"{Fore.BLUE}Testing:{Style.RESET_ALL} {text}")

def print_success(text):
    global passed_tests, total_tests
    passed_tests += 1
    total_tests += 1
    print(f"{Fore.GREEN}✓ PASS{Style.RESET_ALL} - {text}")

def print_failure(text):
    global failed_tests, total_tests
    failed_tests += 1
    total_tests += 1
    print(f"{Fore.RED}✗ FAIL{Style.RESET_ALL} - {text}")

def print_info(text):
    print(f"{Fore.YELLOW}ℹ{Style.RESET_ALL} {text}")

def print_response(data):
    print(f"{Fore.WHITE}Response:{Style.RESET_ALL}")
    if isinstance(data, (dict, list)):
        print(json.dumps(data, indent=2))
    else:
        print(data)

def test_endpoint(method, url, data=None, expected_status=None, description=""):
    """Test an endpoint and return response"""
    print_test(description)
    
    try:
        if method == "GET":
            response = requests.get(url, timeout=5)
        elif method == "POST":
            response = requests.post(url, json=data, timeout=5)
        elif method == "PUT":
            response = requests.put(url, json=data, timeout=5)
        elif method == "DELETE":
            response = requests.delete(url, timeout=5)
        
        if response.status_code == expected_status:
            print_success(f"Status {response.status_code} (Expected {expected_status})")
            if response.text and response.text != "null":
                try:
                    print_response(response.json())
                except:
                    print_response(response.text)
            print()
            return response
        else:
            print_failure(f"Status {response.status_code} (Expected {expected_status})")
            if response.text:
                try:
                    print_response(response.json())
                except:
                    print_response(response.text)
            print()
            return None
            
    except requests.exceptions.ConnectionError:
        print_failure(f"Connection error - Service not reachable")
        print()
        return None
    except Exception as e:
        print_failure(f"Error: {str(e)}")
        print()
        return None

def check_services():
    """Check if services are running"""
    print_section("Checking Services")
    
    try:
        response = requests.get(f"{BOOK_SERVICE}/api/books", timeout=2)
        print_success("Book Service is running on port 8081")
    except:
        print_failure("Book Service is not reachable on port 8081")
        print(f"{Fore.RED}Please start the Book Service first!{Style.RESET_ALL}")
        sys.exit(1)
    
    try:
        response = requests.get(f"{USER_SERVICE}/api/users", timeout=2)
        print_success("User Service is running on port 8082")
    except:
        print_failure("User Service is not reachable on port 8082")
        print(f"{Fore.RED}Please start the User Service first!{Style.RESET_ALL}")
        sys.exit(1)

def test_book_create():
    """Test book creation"""
    print_section("CREATE Operations")
    
    books = [
        {
            "isbn": "9789999999991",
            "title": "Introduction to Go Programming",
            "author": "John Doe",
            "publishYear": 2023,
            "category": "Programming",
            "availableQuantity": 5
        },
        {
            "isbn": "9789999999992",
            "title": "Advanced Go Patterns",
            "author": "Jane Smith",
            "publishYear": 2024,
            "category": "Programming",
            "availableQuantity": 3
        },
        {
            "isbn": "9789999999993",
            "title": "Database Design Fundamentals",
            "author": "Peter Chen",
            "publishYear": 2022,
            "category": "Database",
            "availableQuantity": 4
        }
    ]
    
    for i, book in enumerate(books, 1):
        response = test_endpoint("POST", f"{BOOK_SERVICE}/api/books", book, 201, 
                                f"Create book {i}")
        if response and response.status_code == 201:
            book_id = response.json().get('id')
            created_book_ids.append(book_id)

def test_book_read():
    """Test book reading operations"""
    print_section("READ Operations")
    
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books", None, 200, 
                 "Get all books (default pagination)")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books?page=1&limit=2", None, 200,
                 "Get books with pagination (page=1, limit=2)")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books?page=2&limit=2", None, 200,
                 "Get books with pagination (page=2, limit=2)")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/1", None, 200,
                 "Get book by ID (ID=1)")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/2", None, 200,
                 "Get book by ID (ID=2)")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/999", None, 404,
                 "Get non-existent book (ID=999)")

def test_book_search():
    """Test book search operations"""
    print_section("SEARCH Operations")
    
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/search?title=Go", None, 200,
                 "Search books by title (Go)")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/search?title=Database", None, 200,
                 "Search books by title (Database)")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/search?title=NonExistent", None, 200,
                 "Search books with no results")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/search?title=Go&page=1&limit=1", None, 200,
                 "Search with pagination")

def test_book_update():
    """Test book update operations"""
    print_section("UPDATE Operations")
    
    # Get the first created book ID
    if created_book_ids:
        book_id = created_book_ids[0]
    else:
        # Fallback: search for the book
        response = requests.get(f"{BOOK_SERVICE}/api/books/search?title=Introduction%20to%20Go%20Programming")
        if response.status_code == 200:
            data = response.json()
            if data.get('data') and len(data['data']) > 0:
                book_id = data['data'][0]['id']
            else:
                book_id = 1
        else:
            book_id = 1
    
    update_data = {
        "isbn": "9789999999991",
        "title": "Introduction to Go Programming (2nd Edition)",
        "author": "John Doe",
        "publishYear": 2024,
        "category": "Programming",
        "availableQuantity": 10
    }
    
    test_endpoint("PUT", f"{BOOK_SERVICE}/api/books/{book_id}", update_data, 200,
                 f"Update book (ID={book_id})")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/{book_id}", None, 200,
                 "Verify book update")

def test_book_delete():
    """Test book delete operations"""
    print_section("DELETE Operations")
    
    # Get the third created book ID
    if len(created_book_ids) >= 3:
        book_id = created_book_ids[2]
    else:
        # Fallback: search for the book
        response = requests.get(f"{BOOK_SERVICE}/api/books/search?title=Database%20Design%20Fundamentals")
        if response.status_code == 200:
            data = response.json()
            if data.get('data') and len(data['data']) > 0:
                book_id = data['data'][0]['id']
            else:
                book_id = 3
        else:
            book_id = 3
    
    test_endpoint("DELETE", f"{BOOK_SERVICE}/api/books/{book_id}", None, 204,
                 f"Delete book (ID={book_id})")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/{book_id}", None, 404,
                 "Verify book deletion")

def test_user_create():
    """Test user creation"""
    print_section("CREATE Operations")
    
    users = [
        {
            "username": "alice123",
            "email": "alice@example.com",
            "firstName": "Alice",
            "lastName": "Johnson"
        },
        {
            "username": "bob456",
            "email": "bob@example.com",
            "firstName": "Bob",
            "lastName": "Smith"
        },
        {
            "username": "charlie789",
            "email": "charlie@example.com",
            "firstName": "Charlie",
            "lastName": "Brown"
        }
    ]
    
    for i, user in enumerate(users, 1):
        response = test_endpoint("POST", f"{USER_SERVICE}/api/users", user, 201,
                                f"Create user {i}")
        if response and response.status_code == 201:
            user_id = response.json().get('id')
            created_user_ids.append(user_id)

def test_user_read():
    """Test user reading operations"""
    print_section("READ Operations")
    
    test_endpoint("GET", f"{USER_SERVICE}/api/users", None, 200,
                 "Get all users (default pagination)")
    test_endpoint("GET", f"{USER_SERVICE}/api/users?page=1&limit=2", None, 200,
                 "Get users with pagination (page=1, limit=2)")
    test_endpoint("GET", f"{USER_SERVICE}/api/users?page=2&limit=2", None, 200,
                 "Get users with pagination (page=2, limit=2)")
    test_endpoint("GET", f"{USER_SERVICE}/api/users/1", None, 200,
                 "Get user by ID (ID=1)")
    test_endpoint("GET", f"{USER_SERVICE}/api/users/2", None, 200,
                 "Get user by ID (ID=2)")
    test_endpoint("GET", f"{USER_SERVICE}/api/users/999", None, 404,
                 "Get non-existent user (ID=999)")

def test_user_update():
    """Test user update operations"""
    print_section("UPDATE Operations")
    
    update_data = {
        "username": "alice_updated",
        "email": "alice.new@example.com",
        "firstName": "Alice",
        "lastName": "Johnson-Smith"
    }
    
    test_endpoint("PUT", f"{USER_SERVICE}/api/users/1", update_data, 200,
                 "Update user (ID=1)")
    test_endpoint("GET", f"{USER_SERVICE}/api/users/1", None, 200,
                 "Verify user update")

def test_user_delete():
    """Test user delete operations"""
    print_section("DELETE Operations")
    
    test_endpoint("DELETE", f"{USER_SERVICE}/api/users/3", None, 204,
                 "Delete user (ID=3)")
    test_endpoint("GET", f"{USER_SERVICE}/api/users/3", None, 404,
                 "Verify user deletion")

def test_edge_cases():
    """Test edge cases"""
    print_section("Invalid Requests")
    
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/abc", None, 400,
                 "Get book with invalid ID (non-numeric)")
    test_endpoint("GET", f"{USER_SERVICE}/api/users/xyz", None, 400,
                 "Get user with invalid ID (non-numeric)")
    test_endpoint("GET", f"{BOOK_SERVICE}/api/books/search", None, 400,
                 "Search without title parameter")
    
    invalid_book = {"title": "Missing Required Fields"}
    test_endpoint("POST", f"{BOOK_SERVICE}/api/books", invalid_book, 400,
                 "Create book with missing fields")

def print_summary():
    """Print test summary"""
    print_header("TEST SUMMARY")
    
    print(f"{Style.BRIGHT}Total Tests:{Style.RESET_ALL}   {total_tests}")
    print(f"{Fore.GREEN}{Style.BRIGHT}Passed:{Style.RESET_ALL}        {passed_tests}")
    print(f"{Fore.RED}{Style.BRIGHT}Failed:{Style.RESET_ALL}        {failed_tests}")
    
    if failed_tests == 0:
        print(f"\n{Fore.GREEN}{Style.BRIGHT}🎉 ALL TESTS PASSED! 🎉{Style.RESET_ALL}\n")
        return 0
    else:
        print(f"\n{Fore.RED}{Style.BRIGHT}⚠️  SOME TESTS FAILED ⚠️{Style.RESET_ALL}\n")
        return 1

def main():
    """Main test runner"""
    print_header("LIBRARY MANAGEMENT SYSTEM - API TESTS")
    
    check_services()
    
    # Book Service Tests
    print_header("BOOK SERVICE TESTS")
    test_book_create()
    test_book_read()
    test_book_search()
    test_book_update()
    test_book_delete()
    
    # User Service Tests
    print_header("USER SERVICE TESTS")
    test_user_create()
    test_user_read()
    test_user_update()
    test_user_delete()
    
    # Edge Cases
    print_header("EDGE CASE TESTS")
    test_edge_cases()
    
    # Summary
    exit_code = print_summary()
    sys.exit(exit_code)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print(f"\n\n{Fore.YELLOW}Tests interrupted by user{Style.RESET_ALL}\n")
        sys.exit(1)
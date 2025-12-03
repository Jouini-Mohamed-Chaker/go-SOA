#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Service URLs
BOOK_SERVICE="http://localhost:8081"
USER_SERVICE="http://localhost:8082"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Print functions
print_header() {
    echo -e "\n${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${WHITE}${BOLD}  $1${NC}${CYAN}║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}\n"
}

print_section() {
    echo -e "\n${MAGENTA}▶ $1${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_test() {
    echo -e "${BLUE}Testing:${NC} $1"
}

print_success() {
    ((PASSED_TESTS++))
    ((TOTAL_TESTS++))
    echo -e "${GREEN}✓ PASS${NC} - $1"
}

print_failure() {
    ((FAILED_TESTS++))
    ((TOTAL_TESTS++))
    echo -e "${RED}✗ FAIL${NC} - $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

print_response() {
    echo -e "${WHITE}Response:${NC}"
    echo "$1" | jq '.' 2>/dev/null || echo "$1"
}

# Test function
test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    print_test "$description"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" -H "Content-Type: application/json" -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" == "$expected_status" ]; then
        print_success "Status $http_code (Expected $expected_status)"
        if [ ! -z "$body" ] && [ "$body" != "null" ]; then
            print_response "$body"
        fi
        echo ""
        return 0
    else
        print_failure "Status $http_code (Expected $expected_status)"
        if [ ! -z "$body" ]; then
            print_response "$body"
        fi
        echo ""
        return 1
    fi
}

# Check if services are running
check_services() {
    print_section "Checking Services"
    
    if curl -s "$BOOK_SERVICE/api/books" > /dev/null 2>&1; then
        print_success "Book Service is running on port 8081"
    else
        print_failure "Book Service is not reachable on port 8081"
        echo -e "${RED}Please start the Book Service first!${NC}"
        exit 1
    fi
    
    if curl -s "$USER_SERVICE/api/users" > /dev/null 2>&1; then
        print_success "User Service is running on port 8082"
    else
        print_failure "User Service is not reachable on port 8082"
        echo -e "${RED}Please start the User Service first!${NC}"
        exit 1
    fi
}

# Main test execution
main() {
    clear
    print_header "LIBRARY MANAGEMENT SYSTEM - API TESTS"
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_info "jq is not installed. JSON formatting will be disabled."
        print_info "Install with: sudo apt-get install jq (Ubuntu) or brew install jq (macOS)"
        echo ""
    fi
    
    check_services
    
    # ==================== BOOK SERVICE TESTS ====================
    print_header "BOOK SERVICE TESTS"
    
    # Test 1: Create Book (using unique ISBNs not in init.sql)
    print_section "CREATE Operations"
    book1_data='{
        "isbn": "9789999999991",
        "title": "Introduction to Go Programming",
        "author": "John Doe",
        "publishYear": 2023,
        "category": "Programming",
        "availableQuantity": 5
    }'
    test_endpoint "POST" "$BOOK_SERVICE/api/books" "$book1_data" "201" "Create first book"
    
    book2_data='{
        "isbn": "9789999999992",
        "title": "Advanced Go Patterns",
        "author": "Jane Smith",
        "publishYear": 2024,
        "category": "Programming",
        "availableQuantity": 3
    }'
    test_endpoint "POST" "$BOOK_SERVICE/api/books" "$book2_data" "201" "Create second book"
    
    book3_data='{
        "isbn": "9789999999993",
        "title": "Database Design Fundamentals",
        "author": "Peter Chen",
        "publishYear": 2022,
        "category": "Database",
        "availableQuantity": 4
    }'
    test_endpoint "POST" "$BOOK_SERVICE/api/books" "$book3_data" "201" "Create third book"
    
    # Test 2: Get All Books
    print_section "READ Operations"
    test_endpoint "GET" "$BOOK_SERVICE/api/books" "" "200" "Get all books (default pagination)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books?page=1&limit=2" "" "200" "Get books with pagination (page=1, limit=2)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books?page=2&limit=2" "" "200" "Get books with pagination (page=2, limit=2)"
    
    # Test 3: Get Book by ID
    test_endpoint "GET" "$BOOK_SERVICE/api/books/1" "" "200" "Get book by ID (ID=1)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/2" "" "200" "Get book by ID (ID=2)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/999" "" "404" "Get non-existent book (ID=999)"
    
    # Test 4: Search Books
    print_section "SEARCH Operations"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/search?title=Go" "" "200" "Search books by title (Go)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/search?title=Database" "" "200" "Search books by title (Database)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/search?title=NonExistent" "" "200" "Search books with no results"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/search?title=Go&page=1&limit=1" "" "200" "Search with pagination"
    
    # Test 5: Update Book (use an ID from newly created books)
    print_section "UPDATE Operations"
    update_data='{
        "isbn": "9789999999991",
        "title": "Introduction to Go Programming (2nd Edition)",
        "author": "John Doe",
        "publishYear": 2024,
        "category": "Programming",
        "availableQuantity": 10
    }'
    # Get the ID of the first created book dynamically
    first_book_id=$(curl -s "$BOOK_SERVICE/api/books/search?title=Introduction%20to%20Go%20Programming" | jq -r '.data[0].id')
    test_endpoint "PUT" "$BOOK_SERVICE/api/books/$first_book_id" "$update_data" "200" "Update book (ID=$first_book_id)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/$first_book_id" "" "200" "Verify book update"
    
    # Test 6: Delete Book (use third created book)
    print_section "DELETE Operations"
    third_book_id=$(curl -s "$BOOK_SERVICE/api/books/search?title=Database%20Design%20Fundamentals" | jq -r '.data[0].id')
    test_endpoint "DELETE" "$BOOK_SERVICE/api/books/$third_book_id" "" "204" "Delete book (ID=$third_book_id)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/$third_book_id" "" "404" "Verify book deletion"
    
    # ==================== USER SERVICE TESTS ====================
    print_header "USER SERVICE TESTS"
    
    # Test 7: Create Users
    print_section "CREATE Operations"
    user1_data='{
        "username": "alice123",
        "email": "alice@example.com",
        "firstName": "Alice",
        "lastName": "Johnson"
    }'
    test_endpoint "POST" "$USER_SERVICE/api/users" "$user1_data" "201" "Create first user"
    
    user2_data='{
        "username": "bob456",
        "email": "bob@example.com",
        "firstName": "Bob",
        "lastName": "Smith"
    }'
    test_endpoint "POST" "$USER_SERVICE/api/users" "$user2_data" "201" "Create second user"
    
    user3_data='{
        "username": "charlie789",
        "email": "charlie@example.com",
        "firstName": "Charlie",
        "lastName": "Brown"
    }'
    test_endpoint "POST" "$USER_SERVICE/api/users" "$user3_data" "201" "Create third user"
    
    # Test 8: Get All Users
    print_section "READ Operations"
    test_endpoint "GET" "$USER_SERVICE/api/users" "" "200" "Get all users (default pagination)"
    test_endpoint "GET" "$USER_SERVICE/api/users?page=1&limit=2" "" "200" "Get users with pagination (page=1, limit=2)"
    test_endpoint "GET" "$USER_SERVICE/api/users?page=2&limit=2" "" "200" "Get users with pagination (page=2, limit=2)"
    
    # Test 9: Get User by ID
    test_endpoint "GET" "$USER_SERVICE/api/users/1" "" "200" "Get user by ID (ID=1)"
    test_endpoint "GET" "$USER_SERVICE/api/users/2" "" "200" "Get user by ID (ID=2)"
    test_endpoint "GET" "$USER_SERVICE/api/users/999" "" "404" "Get non-existent user (ID=999)"
    
    # Test 10: Update User
    print_section "UPDATE Operations"
    update_user_data='{
        "username": "alice_updated",
        "email": "alice.new@example.com",
        "firstName": "Alice",
        "lastName": "Johnson-Smith"
    }'
    test_endpoint "PUT" "$USER_SERVICE/api/users/1" "$update_user_data" "200" "Update user (ID=1)"
    test_endpoint "GET" "$USER_SERVICE/api/users/1" "" "200" "Verify user update"
    
    # Test 11: Delete User
    print_section "DELETE Operations"
    test_endpoint "DELETE" "$USER_SERVICE/api/users/3" "" "204" "Delete user (ID=3)"
    test_endpoint "GET" "$USER_SERVICE/api/users/3" "" "404" "Verify user deletion"
    
    # ==================== EDGE CASES ====================
    print_header "EDGE CASE TESTS"
    
    print_section "Invalid Requests"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/abc" "" "400" "Get book with invalid ID (non-numeric)"
    test_endpoint "GET" "$USER_SERVICE/api/users/xyz" "" "400" "Get user with invalid ID (non-numeric)"
    test_endpoint "GET" "$BOOK_SERVICE/api/books/search" "" "400" "Search without title parameter"
    
    invalid_book='{
        "title": "Missing Required Fields"
    }'
    # Now expects 400 with proper validation
    test_endpoint "POST" "$BOOK_SERVICE/api/books" "$invalid_book" "400" "Create book with missing fields"
    
    # ==================== FINAL SUMMARY ====================
    print_header "TEST SUMMARY"
    
    echo -e "${BOLD}Total Tests:${NC}   $TOTAL_TESTS"
    echo -e "${GREEN}${BOLD}Passed:${NC}        $PASSED_TESTS"
    echo -e "${RED}${BOLD}Failed:${NC}        $FAILED_TESTS"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}${BOLD}🎉 ALL TESTS PASSED! 🎉${NC}\n"
        exit 0
    else
        echo -e "\n${RED}${BOLD}⚠️  SOME TESTS FAILED ⚠️${NC}\n"
        exit 1
    fi
}

# Run tests
main
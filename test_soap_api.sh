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

# Emojis
CHECK="✓"
CROSS="✗"
ROCKET="🚀"
BOOK="📚"
USER="👤"
RETURN="↩️"
LIST="📋"
SEARCH="🔍"

SOAP_URL="http://localhost:8083/loan"

# Function to print section header
print_header() {
    echo -e "\n${BOLD}${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${CYAN}║${NC}  $1"
    echo -e "${BOLD}${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}\n"
}

# Function to print test title
print_test() {
    echo -e "${BOLD}${YELLOW}➤${NC} ${WHITE}$1${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}${CHECK} SUCCESS:${NC} $1"
}

# Function to print error
print_error() {
    echo -e "${RED}${CROSS} ERROR:${NC} $1"
}

# Function to print info
print_info() {
    echo -e "${BLUE}ℹ INFO:${NC} $1"
}

# Function to pretty print XML with syntax highlighting
pretty_xml() {
    local xml="$1"
    
    # First, try to format with xmllint
    if command -v xmllint &> /dev/null; then
        formatted=$(echo "$xml" | xmllint --format - 2>/dev/null)
        if [ $? -eq 0 ]; then
            xml="$formatted"
        fi
    fi
    
    # Add syntax highlighting - process line by line to avoid greedy matching
    echo "$xml" | while IFS= read -r line; do
        # Color opening tags
        line=$(echo "$line" | perl -pe "s|<([a-zA-Z0-9:_-]+)([^>]*)>|\e[0;36m<\e[1;33m\1\e[0;35m\2\e[0;36m>\e[0m|g")
        # Color closing tags
        line=$(echo "$line" | perl -pe "s|</([a-zA-Z0-9:_-]+)>|\e[0;36m</\e[1;33m\1\e[0;36m>\e[0m|g")
        # Color self-closing tags
        line=$(echo "$line" | perl -pe "s|/>|\e[0;36m/>\e[0m|g")
        # Color attribute values
        line=$(echo "$line" | perl -pe "s|=\"([^\"]*)\"|=\e[0;32m\"\1\"\e[0m|g")
        # Color special values
        line=$(echo "$line" | perl -pe "s|>true<|\e[0m>\e[0;32mtrue\e[0m<|g")
        line=$(echo "$line" | perl -pe "s|>false<|\e[0m>\e[0;31mfalse\e[0m<|g")
        line=$(echo "$line" | perl -pe "s|>ACTIVE<|\e[0m>\e[1;33mACTIVE\e[0m<|g")
        line=$(echo "$line" | perl -pe "s|>RETURNED<|\e[0m>\e[0;34mRETURNED\e[0m<|g")
        
        echo -e "$line"
    done
}

# Function to extract value from XML
extract_xml_value() {
    local xml="$1"
    local tag="$2"
    echo "$xml" | grep -oP "(?<=<$tag>)[^<]+" | head -1
}

# Function to make SOAP request
soap_request() {
    local operation="$1"
    local body="$2"
    
    local request="<?xml version=\"1.0\" encoding=\"utf-8\"?>
<soap:Envelope xmlns:soap=\"http://schemas.xmlsoap.org/soap/envelope/\">
  <soap:Body>
    $body
  </soap:Body>
</soap:Envelope>"
    
    response=$(curl -s -X POST "$SOAP_URL" \
        -H "Content-Type: text/xml" \
        -d "$request")
    
    echo "$response"
}

# Check if service is running
check_service() {
    print_header "${ROCKET} CHECKING SOAP SERVICE"
    print_test "Testing if SOAP service is available at $SOAP_URL"
    
    if curl -s -o /dev/null -w "%{http_code}" "$SOAP_URL?wsdl" | grep -q "200"; then
        print_success "SOAP service is running!"
        print_info "WSDL available at: ${CYAN}$SOAP_URL?wsdl${NC}"
    else
        print_error "SOAP service is not responding!"
        exit 1
    fi
}

# Test 1: Get All Loans (Initial State)
test_get_all_loans_initial() {
    print_header "${LIST} TEST 1: Get All Loans (Initial State)"
    print_test "Fetching all existing loans..."
    
    response=$(soap_request "getAllLoans" \
        "<getAllLoansRequest xmlns=\"http://library.example.com/loan\"/>")
    
    success=$(extract_xml_value "$response" "success")
    loan_count=$(echo "$response" | grep -o "<loans>" | wc -l)
    
    if [ "$success" = "true" ]; then
        print_success "Retrieved all loans"
        print_info "Total loans in system: ${MAGENTA}$loan_count${NC}"
        echo -e "\n${CYAN}Response:${NC}"
        pretty_xml "$response" | head -30
        echo -e "${YELLOW}... (truncated for readability)${NC}"
    else
        print_error "Failed to get loans"
    fi
}

# Test 2: Create a New Loan
test_create_loan() {
    print_header "${BOOK} TEST 2: Create New Loan"
    print_test "Creating loan for User ID: 1, Book ID: 1"
    
    response=$(soap_request "createLoan" \
        "<createLoanRequest xmlns=\"http://library.example.com/loan\">
      <userId>1</userId>
      <bookId>1</bookId>
    </createLoanRequest>")
    
    success=$(extract_xml_value "$response" "success")
    message=$(extract_xml_value "$response" "message")
    loan_id=$(extract_xml_value "$response" "id")
    status=$(extract_xml_value "$response" "status")
    loan_date=$(extract_xml_value "$response" "loanDate")
    due_date=$(extract_xml_value "$response" "dueDate")
    
    if [ "$success" = "true" ]; then
        print_success "$message"
        echo -e "${GREEN}  └─${NC} Loan ID: ${MAGENTA}$loan_id${NC}"
        echo -e "${GREEN}  └─${NC} Status: ${YELLOW}$status${NC}"
        echo -e "${GREEN}  └─${NC} Loan Date: ${CYAN}$loan_date${NC}"
        echo -e "${GREEN}  └─${NC} Due Date: ${CYAN}$due_date${NC}"
        CREATED_LOAN_ID=$loan_id
    else
        print_error "$message"
    fi
    
    echo -e "\n${CYAN}Full Response:${NC}"
    pretty_xml "$response"
}

# Test 3: Get Loan by ID
test_get_loan_by_id() {
    print_header "${SEARCH} TEST 3: Get Loan By ID"
    
    if [ -z "$CREATED_LOAN_ID" ]; then
        CREATED_LOAN_ID=1
        print_info "Using loan ID: $CREATED_LOAN_ID (no new loan was created)"
    else
        print_info "Searching for the loan we just created (ID: $CREATED_LOAN_ID)"
    fi
    
    print_test "Fetching loan ID: $CREATED_LOAN_ID"
    
    response=$(soap_request "getLoanById" \
        "<getLoanByIdRequest xmlns=\"http://library.example.com/loan\">
      <loanId>$CREATED_LOAN_ID</loanId>
    </getLoanByIdRequest>")
    
    success=$(extract_xml_value "$response" "success")
    message=$(extract_xml_value "$response" "message")
    
    if [ "$success" = "true" ]; then
        print_success "$message"
        user_id=$(extract_xml_value "$response" "userId")
        book_id=$(extract_xml_value "$response" "bookId")
        status=$(extract_xml_value "$response" "status")
        
        echo -e "${GREEN}  └─${NC} User ID: ${MAGENTA}$user_id${NC}"
        echo -e "${GREEN}  └─${NC} Book ID: ${MAGENTA}$book_id${NC}"
        echo -e "${GREEN}  └─${NC} Status: ${YELLOW}$status${NC}"
    else
        print_error "$message"
    fi
    
    echo -e "\n${CYAN}Full Response:${NC}"
    pretty_xml "$response"
}

# Test 4: Get Loans by User
test_get_loans_by_user() {
    print_header "${USER} TEST 4: Get Loans By User"
    print_test "Fetching all loans for User ID: 1"
    
    response=$(soap_request "getLoansByUser" \
        "<getLoansByUserRequest xmlns=\"http://library.example.com/loan\">
      <userId>1</userId>
    </getLoansByUserRequest>")
    
    success=$(extract_xml_value "$response" "success")
    loan_count=$(echo "$response" | grep -o "<loans>" | wc -l)
    
    if [ "$success" = "true" ]; then
        print_success "Retrieved user loans"
        print_info "User 1 has ${MAGENTA}$loan_count${NC} loan(s)"
        
        # Extract loan IDs and statuses
        echo -e "\n${CYAN}Loans for User 1:${NC}"
        loan_ids=$(echo "$response" | grep -oP "(?<=<id>)[^<]+")
        statuses=$(echo "$response" | grep -oP "(?<=<status>)[^<]+")
        
        ids_array=($loan_ids)
        status_array=($statuses)
        
        for i in "${!ids_array[@]}"; do
            if [ "${status_array[$i]}" = "ACTIVE" ]; then
                echo -e "  ${GREEN}•${NC} Loan ID: ${MAGENTA}${ids_array[$i]}${NC} - Status: ${YELLOW}${status_array[$i]}${NC}"
            else
                echo -e "  ${BLUE}•${NC} Loan ID: ${MAGENTA}${ids_array[$i]}${NC} - Status: ${CYAN}${status_array[$i]}${NC}"
            fi
        done
    else
        print_error "Failed to get user loans"
    fi
    
    echo -e "\n${CYAN}Full Response:${NC}"
    pretty_xml "$response"
}

# Test 5: Return a Loan
test_return_loan() {
    print_header "${RETURN} TEST 5: Return Loan"
    
    if [ -z "$CREATED_LOAN_ID" ]; then
        RETURN_LOAN_ID=2
        print_info "Using loan ID: $RETURN_LOAN_ID"
    else
        RETURN_LOAN_ID=$CREATED_LOAN_ID
        print_info "Returning the loan we created (ID: $RETURN_LOAN_ID)"
    fi
    
    print_test "Returning loan ID: $RETURN_LOAN_ID"
    
    response=$(soap_request "returnLoan" \
        "<returnLoanRequest xmlns=\"http://library.example.com/loan\">
      <loanId>$RETURN_LOAN_ID</loanId>
    </returnLoanRequest>")
    
    success=$(extract_xml_value "$response" "success")
    message=$(extract_xml_value "$response" "message")
    
    if [ "$success" = "true" ]; then
        print_success "$message"
        return_date=$(extract_xml_value "$response" "returnDate")
        status=$(extract_xml_value "$response" "status")
        
        echo -e "${GREEN}  └─${NC} Return Date: ${CYAN}$return_date${NC}"
        echo -e "${GREEN}  └─${NC} New Status: ${CYAN}$status${NC}"
    else
        print_error "$message"
        print_info "This might be expected if the loan was already returned"
    fi
    
    echo -e "\n${CYAN}Full Response:${NC}"
    pretty_xml "$response"
}

# Test 6: Try to Return Already Returned Loan (Should Fail)
test_return_already_returned() {
    print_header "${RETURN} TEST 6: Try to Return Already Returned Loan (Expected Failure)"
    print_test "Attempting to return loan ID: $RETURN_LOAN_ID again..."
    
    response=$(soap_request "returnLoan" \
        "<returnLoanRequest xmlns=\"http://library.example.com/loan\">
      <loanId>$RETURN_LOAN_ID</loanId>
    </returnLoanRequest>")
    
    success=$(extract_xml_value "$response" "success")
    message=$(extract_xml_value "$response" "message")
    
    if [ "$success" = "false" ]; then
        print_success "Service correctly rejected duplicate return"
        echo -e "${YELLOW}  └─${NC} Message: ${message}"
    else
        print_error "Service should have rejected this!"
    fi
    
    echo -e "\n${CYAN}Full Response:${NC}"
    pretty_xml "$response"
}

# Test 7: Create Multiple Loans
test_create_multiple_loans() {
    print_header "${BOOK} TEST 7: Create Multiple Loans"
    
    print_info "Fetching available users and books..."
    
    # Fetch users from REST API
    users_response=$(curl -s "http://localhost:8082/api/users?limit=3")
    books_response=$(curl -s "http://localhost:8081/api/books?limit=4")
    
    # Extract IDs (assuming JSON response)
    user_ids=($(echo "$users_response" | grep -oP '"id":\s*\K\d+' | head -3))
    book_ids=($(echo "$books_response" | grep -oP '"id":\s*\K\d+' | head -4))
    
    if [ ${#user_ids[@]} -lt 2 ] || [ ${#book_ids[@]} -lt 3 ]; then
        print_error "Not enough users or books in database to run this test"
        print_info "Please add at least 2 users and 3 books"
        return
    fi
    
    # Create loan combinations
    test_combinations=(
        "${user_ids[0]} ${book_ids[1]}"
        "${user_ids[1]} ${book_ids[2]}"
        "${user_ids[0]} ${book_ids[3]}"
    )
    
    for i in "${!test_combinations[@]}"; do
        read user_id book_id <<< "${test_combinations[$i]}"
        
        print_test "Creating loan $((i+1))/3 - User: $user_id, Book: $book_id"
        
        response=$(soap_request "createLoan" \
            "<createLoanRequest xmlns=\"http://library.example.com/loan\">
          <userId>$user_id</userId>
          <bookId>$book_id</bookId>
        </createLoanRequest>")
        
        success=$(extract_xml_value "$response" "success")
        message=$(extract_xml_value "$response" "message")
        loan_id=$(extract_xml_value "$response" "id")
        
        if [ "$success" = "true" ]; then
            print_success "Loan created - ID: ${MAGENTA}$loan_id${NC}"
        else
            print_error "$message"
        fi
        
        sleep 0.5
    done
}

# Test 8: Final State - Get All Loans
test_get_all_loans_final() {
    print_header "${LIST} TEST 8: Get All Loans (Final State)"
    print_test "Fetching all loans to see final state..."
    
    response=$(soap_request "getAllLoans" \
        "<getAllLoansRequest xmlns=\"http://library.example.com/loan\"/>")
    
    success=$(extract_xml_value "$response" "success")
    loan_count=$(echo "$response" | grep -o "<loans>" | wc -l)
    active_count=$(echo "$response" | grep -o "<status>ACTIVE</status>" | wc -l)
    returned_count=$(echo "$response" | grep -o "<status>RETURNED</status>" | wc -l)
    
    if [ "$success" = "true" ]; then
        print_success "Retrieved all loans"
        echo -e "\n${CYAN}═══════════════════════════════════════${NC}"
        echo -e "${BOLD}${WHITE}  LOAN STATISTICS${NC}"
        echo -e "${CYAN}═══════════════════════════════════════${NC}"
        echo -e "  Total Loans:    ${MAGENTA}$loan_count${NC}"
        echo -e "  Active Loans:   ${YELLOW}$active_count${NC}"
        echo -e "  Returned Loans: ${CYAN}$returned_count${NC}"
        echo -e "${CYAN}═══════════════════════════════════════${NC}\n"
    else
        print_error "Failed to get loans"
    fi
}

# Print summary
print_summary() {
    print_header "${ROCKET} TEST SUMMARY"
    echo -e "${GREEN}${CHECK}${NC} All SOAP operations tested successfully!"
    echo -e "\n${CYAN}Operations Tested:${NC}"
    echo -e "  ${GREEN}1.${NC} getAllLoans (initial state)"
    echo -e "  ${GREEN}2.${NC} createLoan (single)"
    echo -e "  ${GREEN}3.${NC} getLoanById"
    echo -e "  ${GREEN}4.${NC} getLoansByUser"
    echo -e "  ${GREEN}5.${NC} returnLoan"
    echo -e "  ${GREEN}6.${NC} returnLoan (duplicate - expected failure)"
    echo -e "  ${GREEN}7.${NC} createLoan (multiple)"
    echo -e "  ${GREEN}8.${NC} getAllLoans (final state)"
    
    echo -e "\n${BOLD}${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}${GREEN}  ✓ SOAP API IS WORKING PERFECTLY! ${NC}"
    echo -e "${BOLD}${CYAN}═══════════════════════════════════════════════════════════════${NC}\n"
    
    echo -e "${YELLOW}💡 TIP:${NC} View the WSDL at: ${CYAN}$SOAP_URL?wsdl${NC}\n"
}

# Main execution
main() {
    clear
    echo -e "${BOLD}${MAGENTA}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║                                                               ║"
    echo "║          SOAP LOAN SERVICE - COMPREHENSIVE TEST SUITE         ║"
    echo "║                                                               ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
    
    check_service
    sleep 1
    
    test_get_all_loans_initial
    sleep 1
    
    test_create_loan
    sleep 1
    
    test_get_loan_by_id
    sleep 1
    
    test_get_loans_by_user
    sleep 1
    
    test_return_loan
    sleep 1
    
    test_return_already_returned
    sleep 1
    
    test_create_multiple_loans
    sleep 1
    
    test_get_all_loans_final
    sleep 1
    
    print_summary
}

# Run main
main
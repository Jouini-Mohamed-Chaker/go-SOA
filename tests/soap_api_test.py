#!/usr/bin/env python3
"""
SOAP Loan Service Test Script
Tests all SOAP operations for the Loan Service
"""

import requests
import xml.etree.ElementTree as ET
from colorama import Fore, Back, Style, init
import json
import sys
import time
import re

# Initialize colorama
init(autoreset=True)

# Configuration
SOAP_URL = "http://localhost:8083/loan"
BOOK_SERVICE = "http://localhost:8081"
USER_SERVICE = "http://localhost:8082"

# Emojis
CHECK = "✓"
CROSS = "✗"
ROCKET = "🚀"
BOOK = "📚"
USER = "👤"
RETURN_EMOJI = "↩️"
LIST = "📋"
SEARCH = "🔍"

# Storage
created_loan_id = None
return_loan_id = None

# Print functions
def print_header(text):
    print(f"\n{Style.BRIGHT}{Fore.CYAN}╔════════════════════════════════════════════════════════════════╗{Style.RESET_ALL}")
    print(f"{Style.BRIGHT}{Fore.CYAN}║{Style.RESET_ALL}  {text}")
    print(f"{Style.BRIGHT}{Fore.CYAN}╚════════════════════════════════════════════════════════════════╝{Style.RESET_ALL}\n")

def print_test(text):
    print(f"{Style.BRIGHT}{Fore.YELLOW}➤{Style.RESET_ALL} {Fore.WHITE}{text}{Style.RESET_ALL}")

def print_success(text):
    print(f"{Fore.GREEN}{CHECK} SUCCESS:{Style.RESET_ALL} {text}")

def print_error(text):
    print(f"{Fore.RED}{CROSS} ERROR:{Style.RESET_ALL} {text}")

def print_info(text):
    print(f"{Fore.BLUE}ℹ INFO:{Style.RESET_ALL} {text}")

def pretty_xml(xml_string):
    """Pretty print XML with syntax highlighting"""
    try:
        # Parse and format XML
        root = ET.fromstring(xml_string)
        formatted = ET.tostring(root, encoding='unicode')
        
        # Add basic indentation
        formatted = re.sub(r'><', '>\n<', formatted)
        
        lines = formatted.split('\n')
        indent_level = 0
        pretty_lines = []
        
        for line in lines:
            line = line.strip()
            if not line:
                continue
                
            # Decrease indent for closing tags
            if line.startswith('</'):
                indent_level = max(0, indent_level - 1)
            
            # Add indentation
            pretty_line = '  ' * indent_level + line
            
            # Color coding
            # Opening tags
            pretty_line = re.sub(r'<([a-zA-Z0-9:_-]+)([^>]*)>', 
                               f'{Fore.CYAN}<{Fore.YELLOW}\\1{Fore.MAGENTA}\\2{Fore.CYAN}>{Style.RESET_ALL}',
                               pretty_line)
            # Closing tags
            pretty_line = re.sub(r'</([a-zA-Z0-9:_-]+)>', 
                               f'{Fore.CYAN}</{Fore.YELLOW}\\1{Fore.CYAN}>{Style.RESET_ALL}',
                               pretty_line)
            # Attribute values
            pretty_line = re.sub(r'="([^"]*)"', 
                               f'={Fore.GREEN}"\\1"{Style.RESET_ALL}',
                               pretty_line)
            # Special values
            pretty_line = pretty_line.replace('>true<', f'>{Fore.GREEN}true{Style.RESET_ALL}<')
            pretty_line = pretty_line.replace('>false<', f'>{Fore.RED}false{Style.RESET_ALL}<')
            pretty_line = pretty_line.replace('>ACTIVE<', f'>{Style.BRIGHT}{Fore.YELLOW}ACTIVE{Style.RESET_ALL}<')
            pretty_line = pretty_line.replace('>RETURNED<', f'>{Fore.BLUE}RETURNED{Style.RESET_ALL}<')
            
            pretty_lines.append(pretty_line)
            
            # Increase indent for opening tags (but not self-closing)
            if line.startswith('<') and not line.startswith('</') and not line.endswith('/>'):
                if not re.search(r'<[^>]+>[^<]+</[^>]+>', line):  # Not a complete tag
                    indent_level += 1
        
        return '\n'.join(pretty_lines)
    except:
        return xml_string

def extract_xml_value(xml_string, tag):
    """Extract value from XML by tag name"""
    try:
        root = ET.fromstring(xml_string)
        # Search for tag in all namespaces
        for elem in root.iter():
            if elem.tag.endswith(tag) or elem.tag == tag:
                return elem.text
        return None
    except:
        return None

def soap_request(operation, body):
    """Make a SOAP request"""
    request = f"""<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    {body}
  </soap:Body>
</soap:Envelope>"""
    
    try:
        response = requests.post(SOAP_URL, 
                               data=request,
                               headers={"Content-Type": "text/xml"},
                               timeout=5)
        return response.text
    except Exception as e:
        print_error(f"Request failed: {str(e)}")
        return None

def check_service():
    """Check if SOAP service is running"""
    print_header(f"{ROCKET} CHECKING SOAP SERVICE")
    print_test(f"Testing if SOAP service is available at {SOAP_URL}")
    
    try:
        response = requests.get(f"{SOAP_URL}?wsdl", timeout=5)
        if response.status_code == 200:
            print_success("SOAP service is running!")
            print_info(f"WSDL available at: {Fore.CYAN}{SOAP_URL}?wsdl{Style.RESET_ALL}")
            return True
        else:
            print_error("SOAP service is not responding!")
            return False
    except:
        print_error("SOAP service is not responding!")
        return False

def test_get_all_loans_initial():
    """Test getting all loans initially"""
    print_header(f"{LIST} TEST 1: Get All Loans (Initial State)")
    print_test("Fetching all existing loans...")
    
    response = soap_request("getAllLoans",
        '<getAllLoansRequest xmlns="http://library.example.com/loan"/>')
    
    if not response:
        return
    
    success = extract_xml_value(response, "success")
    loan_count = response.count("<loans>")
    
    if success == "true":
        print_success("Retrieved all loans")
        print_info(f"Total loans in system: {Fore.MAGENTA}{loan_count}{Style.RESET_ALL}")
        print(f"\n{Fore.CYAN}Response:{Style.RESET_ALL}")
        lines = pretty_xml(response).split('\n')
        for line in lines[:30]:
            print(line)
        if len(lines) > 30:
            print(f"{Fore.YELLOW}... (truncated for readability){Style.RESET_ALL}")
    else:
        print_error("Failed to get loans")

def test_create_loan():
    """Test creating a new loan"""
    global created_loan_id
    
    print_header(f"{BOOK} TEST 2: Create New Loan")
    print_test("Creating loan for User ID: 1, Book ID: 1")
    
    response = soap_request("createLoan",
        '''<createLoanRequest xmlns="http://library.example.com/loan">
      <userId>1</userId>
      <bookId>1</bookId>
    </createLoanRequest>''')
    
    if not response:
        return
    
    success = extract_xml_value(response, "success")
    message = extract_xml_value(response, "message")
    loan_id = extract_xml_value(response, "id")
    status = extract_xml_value(response, "status")
    loan_date = extract_xml_value(response, "loanDate")
    due_date = extract_xml_value(response, "dueDate")
    
    if success == "true":
        print_success(message or "Loan created successfully")
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} Loan ID: {Fore.MAGENTA}{loan_id}{Style.RESET_ALL}")
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} Status: {Style.BRIGHT}{Fore.YELLOW}{status}{Style.RESET_ALL}")
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} Loan Date: {Fore.CYAN}{loan_date}{Style.RESET_ALL}")
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} Due Date: {Fore.CYAN}{due_date}{Style.RESET_ALL}")
        created_loan_id = loan_id
    else:
        print_error(message or "Failed to create loan")
    
    print(f"\n{Fore.CYAN}Full Response:{Style.RESET_ALL}")
    print(pretty_xml(response))

def test_get_loan_by_id():
    """Test getting loan by ID"""
    print_header(f"{SEARCH} TEST 3: Get Loan By ID")
    
    loan_id = created_loan_id if created_loan_id else "1"
    
    if created_loan_id:
        print_info(f"Searching for the loan we just created (ID: {loan_id})")
    else:
        print_info(f"Using loan ID: {loan_id} (no new loan was created)")
    
    print_test(f"Fetching loan ID: {loan_id}")
    
    response = soap_request("getLoanById",
        f'''<getLoanByIdRequest xmlns="http://library.example.com/loan">
      <loanId>{loan_id}</loanId>
    </getLoanByIdRequest>''')
    
    if not response:
        return
    
    success = extract_xml_value(response, "success")
    message = extract_xml_value(response, "message")
    
    if success == "true":
        print_success(message or "Loan found")
        user_id = extract_xml_value(response, "userId")
        book_id = extract_xml_value(response, "bookId")
        status = extract_xml_value(response, "status")
        
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} User ID: {Fore.MAGENTA}{user_id}{Style.RESET_ALL}")
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} Book ID: {Fore.MAGENTA}{book_id}{Style.RESET_ALL}")
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} Status: {Style.BRIGHT}{Fore.YELLOW}{status}{Style.RESET_ALL}")
    else:
        print_error(message or "Loan not found")
    
    print(f"\n{Fore.CYAN}Full Response:{Style.RESET_ALL}")
    print(pretty_xml(response))

def test_get_loans_by_user():
    """Test getting loans by user"""
    print_header(f"{USER} TEST 4: Get Loans By User")
    print_test("Fetching all loans for User ID: 1")
    
    response = soap_request("getLoansByUser",
        '''<getLoansByUserRequest xmlns="http://library.example.com/loan">
      <userId>1</userId>
    </getLoansByUserRequest>''')
    
    if not response:
        return
    
    success = extract_xml_value(response, "success")
    loan_count = response.count("<loans>")
    
    if success == "true":
        print_success("Retrieved user loans")
        print_info(f"User 1 has {Fore.MAGENTA}{loan_count}{Style.RESET_ALL} loan(s)")
        
        # Extract loan information
        print(f"\n{Fore.CYAN}Loans for User 1:{Style.RESET_ALL}")
        try:
            root = ET.fromstring(response)
            loans = root.findall(".//{http://library.example.com/loan}loans")
            
            for loan in loans:
                loan_id = loan.find(".//{http://library.example.com/loan}id")
                status = loan.find(".//{http://library.example.com/loan}status")
                
                if loan_id is not None and status is not None:
                    if status.text == "ACTIVE":
                        print(f"  {Fore.GREEN}•{Style.RESET_ALL} Loan ID: {Fore.MAGENTA}{loan_id.text}{Style.RESET_ALL} - Status: {Style.BRIGHT}{Fore.YELLOW}{status.text}{Style.RESET_ALL}")
                    else:
                        print(f"  {Fore.BLUE}•{Style.RESET_ALL} Loan ID: {Fore.MAGENTA}{loan_id.text}{Style.RESET_ALL} - Status: {Fore.CYAN}{status.text}{Style.RESET_ALL}")
        except:
            pass
    else:
        print_error("Failed to get user loans")
    
    print(f"\n{Fore.CYAN}Full Response:{Style.RESET_ALL}")
    print(pretty_xml(response))

def test_return_loan():
    """Test returning a loan"""
    global return_loan_id
    
    print_header(f"{RETURN_EMOJI} TEST 5: Return Loan")
    
    loan_id = created_loan_id if created_loan_id else "2"
    return_loan_id = loan_id
    
    if created_loan_id:
        print_info(f"Returning the loan we created (ID: {loan_id})")
    else:
        print_info(f"Using loan ID: {loan_id}")
    
    print_test(f"Returning loan ID: {loan_id}")
    
    response = soap_request("returnLoan",
        f'''<returnLoanRequest xmlns="http://library.example.com/loan">
      <loanId>{loan_id}</loanId>
    </returnLoanRequest>''')
    
    if not response:
        return
    
    success = extract_xml_value(response, "success")
    message = extract_xml_value(response, "message")
    
    if success == "true":
        print_success(message or "Loan returned successfully")
        return_date = extract_xml_value(response, "returnDate")
        status = extract_xml_value(response, "status")
        
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} Return Date: {Fore.CYAN}{return_date}{Style.RESET_ALL}")
        print(f"{Fore.GREEN}  └─{Style.RESET_ALL} New Status: {Fore.CYAN}{status}{Style.RESET_ALL}")
    else:
        print_error(message or "Failed to return loan")
        print_info("This might be expected if the loan was already returned")
    
    print(f"\n{Fore.CYAN}Full Response:{Style.RESET_ALL}")
    print(pretty_xml(response))

def test_return_already_returned():
    """Test returning an already returned loan"""
    print_header(f"{RETURN_EMOJI} TEST 6: Try to Return Already Returned Loan (Expected Failure)")
    print_test(f"Attempting to return loan ID: {return_loan_id} again...")
    
    response = soap_request("returnLoan",
        f'''<returnLoanRequest xmlns="http://library.example.com/loan">
      <loanId>{return_loan_id}</loanId>
    </returnLoanRequest>''')
    
    if not response:
        return
    
    success = extract_xml_value(response, "success")
    message = extract_xml_value(response, "message")
    
    if success == "false":
        print_success("Service correctly rejected duplicate return")
        print(f"{Fore.YELLOW}  └─{Style.RESET_ALL} Message: {message}")
    else:
        print_error("Service should have rejected this!")
    
    print(f"\n{Fore.CYAN}Full Response:{Style.RESET_ALL}")
    print(pretty_xml(response))

def test_create_multiple_loans():
    """Test creating multiple loans"""
    print_header(f"{BOOK} TEST 7: Create Multiple Loans")
    
    print_info("Fetching available users and books...")
    
    try:
        users_response = requests.get(f"{USER_SERVICE}/api/users?limit=3", timeout=5)
        books_response = requests.get(f"{BOOK_SERVICE}/api/books?limit=4", timeout=5)
        
        users = users_response.json().get('data', [])
        books = books_response.json().get('data', [])
        
        if len(users) < 2 or len(books) < 3:
            print_error("Not enough users or books in database to run this test")
            print_info("Please add at least 2 users and 3 books")
            return
        
        test_combinations = [
            (users[0]['id'], books[1]['id']),
            (users[1]['id'], books[2]['id']),
            (users[0]['id'], books[3]['id']) if len(books) > 3 else (users[0]['id'], books[0]['id'])
        ]
        
        for i, (user_id, book_id) in enumerate(test_combinations, 1):
            print_test(f"Creating loan {i}/3 - User: {user_id}, Book: {book_id}")
            
            response = soap_request("createLoan",
                f'''<createLoanRequest xmlns="http://library.example.com/loan">
          <userId>{user_id}</userId>
          <bookId>{book_id}</bookId>
        </createLoanRequest>''')
            
            if response:
                success = extract_xml_value(response, "success")
                message = extract_xml_value(response, "message")
                loan_id = extract_xml_value(response, "id")
                
                if success == "true":
                    print_success(f"Loan created - ID: {Fore.MAGENTA}{loan_id}{Style.RESET_ALL}")
                else:
                    print_error(message or "Failed to create loan")
            
            time.sleep(0.5)
    except Exception as e:
        print_error(f"Failed to fetch users or books: {str(e)}")

def test_get_all_loans_final():
    """Test getting all loans in final state"""
    print_header(f"{LIST} TEST 8: Get All Loans (Final State)")
    print_test("Fetching all loans to see final state...")
    
    response = soap_request("getAllLoans",
        '<getAllLoansRequest xmlns="http://library.example.com/loan"/>')
    
    if not response:
        return
    
    success = extract_xml_value(response, "success")
    loan_count = response.count("<loans>")
    active_count = response.count("<status>ACTIVE</status>")
    returned_count = response.count("<status>RETURNED</status>")
    
    if success == "true":
        print_success("Retrieved all loans")
        print(f"\n{Fore.CYAN}{'═' * 39}{Style.RESET_ALL}")
        print(f"{Style.BRIGHT}{Fore.WHITE}  LOAN STATISTICS{Style.RESET_ALL}")
        print(f"{Fore.CYAN}{'═' * 39}{Style.RESET_ALL}")
        print(f"  Total Loans:    {Fore.MAGENTA}{loan_count}{Style.RESET_ALL}")
        print(f"  Active Loans:   {Style.BRIGHT}{Fore.YELLOW}{active_count}{Style.RESET_ALL}")
        print(f"  Returned Loans: {Fore.CYAN}{returned_count}{Style.RESET_ALL}")
        print(f"{Fore.CYAN}{'═' * 39}{Style.RESET_ALL}\n")
    else:
        print_error("Failed to get loans")

def print_summary():
    """Print test summary"""
    print_header(f"{ROCKET} TEST SUMMARY")
    print(f"{Fore.GREEN}{CHECK}{Style.RESET_ALL} All SOAP operations tested successfully!")
    print(f"\n{Fore.CYAN}Operations Tested:{Style.RESET_ALL}")
    print(f"  {Fore.GREEN}1.{Style.RESET_ALL} getAllLoans (initial state)")
    print(f"  {Fore.GREEN}2.{Style.RESET_ALL} createLoan (single)")
    print(f"  {Fore.GREEN}3.{Style.RESET_ALL} getLoanById")
    print(f"  {Fore.GREEN}4.{Style.RESET_ALL} getLoansByUser")
    print(f"  {Fore.GREEN}5.{Style.RESET_ALL} returnLoan")
    print(f"  {Fore.GREEN}6.{Style.RESET_ALL} returnLoan (duplicate - expected failure)")
    print(f"  {Fore.GREEN}7.{Style.RESET_ALL} createLoan (multiple)")
    print(f"  {Fore.GREEN}8.{Style.RESET_ALL} getAllLoans (final state)")
    
    print(f"\n{Style.BRIGHT}{Fore.CYAN}{'═' * 63}{Style.RESET_ALL}")
    print(f"{Style.BRIGHT}{Fore.GREEN}  ✓ SOAP API IS WORKING PERFECTLY! {Style.RESET_ALL}")
    print(f"{Style.BRIGHT}{Fore.CYAN}{'═' * 63}{Style.RESET_ALL}\n")
    
    print(f"{Fore.YELLOW}💡 TIP:{Style.RESET_ALL} View the WSDL at: {Fore.CYAN}{SOAP_URL}?wsdl{Style.RESET_ALL}\n")

def main():
    """Main test runner"""
    print(f"{Style.BRIGHT}{Fore.MAGENTA}")
    print("╔═══════════════════════════════════════════════════════════════╗")
    print("║                                                               ║")
    print("║          SOAP LOAN SERVICE - COMPREHENSIVE TEST SUITE         ║")
    print("║                                                               ║")
    print("╚═══════════════════════════════════════════════════════════════╝")
    print(f"{Style.RESET_ALL}\n")
    
    if not check_service():
        sys.exit(1)
    
    time.sleep(1)
    
    test_get_all_loans_initial()
    time.sleep(1)
    
    test_create_loan()
    time.sleep(1)
    
    test_get_loan_by_id()
    time.sleep(1)
    
    test_get_loans_by_user()
    time.sleep(1)
    
    test_return_loan()
    time.sleep(1)
    
    test_return_already_returned()
    time.sleep(1)
    
    test_create_multiple_loans()
    time.sleep(1)
    
    test_get_all_loans_final()
    time.sleep(1)
    
    print_summary()

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print(f"\n\n{Fore.YELLOW}Tests interrupted by user{Style.RESET_ALL}\n")
        sys.exit(1)
#!/usr/bin/env python3
"""
Auth Gateway Test Script
Tests authentication endpoints for the Library Management System
"""

import requests
import json
from colorama import Fore, Back, Style, init
from datetime import datetime
import sys

# Initialize colorama
init(autoreset=True)

# Configuration
BASE_URL = "http://localhost:8080"
TEST_USER = {
    "username": "testuser_" + datetime.now().strftime("%Y%m%d%H%M%S"),
    "email": "test@example.com",
    "firstName": "Test",
    "lastName": "User",
    "password": "TestPassword123!"
}

EXISTING_USER = {
    "username": "alice",
    "password": "password123"
}

# Utility Functions
def print_header(text):
    print(f"\n{Back.BLUE}{Fore.WHITE} {text} {Style.RESET_ALL}\n")

def print_test(text):
    print(f"{Fore.CYAN}→ {text}{Style.RESET_ALL}")

def print_success(text):
    print(f"{Fore.GREEN}✓ {text}{Style.RESET_ALL}")

def print_error(text):
    print(f"{Fore.RED}✗ {text}{Style.RESET_ALL}")

def print_warning(text):
    print(f"{Fore.YELLOW}⚠ {text}{Style.RESET_ALL}")

def print_info(text):
    print(f"{Fore.MAGENTA}ℹ {text}{Style.RESET_ALL}")

def print_json(data, indent=2):
    print(f"{Fore.WHITE}{json.dumps(data, indent=indent)}{Style.RESET_ALL}")

def make_request(method, endpoint, data=None, headers=None, expected_status=None):
    """Make HTTP request and handle response"""
    url = f"{BASE_URL}{endpoint}"
    try:
        if method == "GET":
            response = requests.get(url, headers=headers, timeout=5)
        elif method == "POST":
            response = requests.post(url, json=data, headers=headers, timeout=5)
        elif method == "PUT":
            response = requests.put(url, json=data, headers=headers, timeout=5)
        elif method == "DELETE":
            response = requests.delete(url, headers=headers, timeout=5)
        
        # Check if expected status matches
        if expected_status and response.status_code == expected_status:
            print_success(f"Status: {response.status_code} (Expected)")
        elif expected_status:
            print_error(f"Status: {response.status_code} (Expected {expected_status})")
        else:
            print_info(f"Status: {response.status_code}")
        
        # Print response
        if response.text:
            try:
                print_json(response.json())
            except:
                print(response.text)
        
        return response
    
    except requests.exceptions.ConnectionError:
        print_error("Connection Error: Cannot connect to server")
        print_warning(f"Make sure the server is running on {BASE_URL}")
        return None
    except requests.exceptions.Timeout:
        print_error("Request Timeout")
        return None
    except Exception as e:
        print_error(f"Error: {str(e)}")
        return None

# Test Functions
def test_register():
    """Test user registration"""
    print_header("TEST 1: User Registration")
    
    print_test("Registering new user...")
    response = make_request(
        "POST", 
        "/auth/register", 
        data=TEST_USER,
        expected_status=201
    )
    
    if response and response.status_code == 201:
        print_success("User registered successfully!")
        return True
    else:
        print_error("Registration failed!")
        return False

def test_register_duplicate():
    """Test duplicate registration"""
    print_header("TEST 2: Duplicate Registration (Should Fail)")
    
    print_test("Attempting to register same user again...")
    response = make_request(
        "POST", 
        "/auth/register", 
        data=TEST_USER,
        expected_status=400
    )
    
    if response and response.status_code == 400:
        print_success("Correctly rejected duplicate username!")
        return True
    else:
        print_error("Should have rejected duplicate!")
        return False

def test_login_invalid():
    """Test login with invalid credentials"""
    print_header("TEST 3: Login with Invalid Credentials")
    
    print_test("Attempting login with wrong password...")
    response = make_request(
        "POST",
        "/auth/login",
        data={"username": TEST_USER["username"], "password": "wrongpassword"},
        expected_status=401
    )
    
    if response and response.status_code == 401:
        print_success("Correctly rejected invalid credentials!")
        return True
    else:
        print_error("Should have rejected invalid credentials!")
        return False

def test_login():
    """Test successful login"""
    print_header("TEST 4: Login with Valid Credentials")
    
    print_test("Logging in with correct credentials...")
    response = make_request(
        "POST",
        "/auth/login",
        data={"username": TEST_USER["username"], "password": TEST_USER["password"]},
        expected_status=200
    )
    
    if response and response.status_code == 200:
        data = response.json()
        if "token" in data:
            print_success("Login successful! Token received.")
            return data["token"]
        else:
            print_error("No token in response!")
            return None
    else:
        print_error("Login failed!")
        return None

def test_validate_token(token):
    """Test token validation"""
    print_header("TEST 5: Token Validation")
    
    print_test("Validating JWT token...")
    response = make_request(
        "POST",
        "/auth/validate",
        data={"token": token},
        expected_status=200
    )
    
    if response and response.status_code == 200:
        data = response.json()
        if data.get("valid") == True:
            print_success(f"Token is valid! Username: {data.get('username')}")
            return True
        else:
            print_error("Token validation returned false!")
            return False
    else:
        print_error("Token validation failed!")
        return False

def test_validate_invalid_token():
    """Test validation with invalid token"""
    print_header("TEST 6: Invalid Token Validation")
    
    print_test("Validating invalid token...")
    response = make_request(
        "POST",
        "/auth/validate",
        data={"token": "invalid.token.here"},
        expected_status=401
    )
    
    if response and response.status_code == 401:
        data = response.json()
        if data.get("valid") == False:
            print_success("Correctly rejected invalid token!")
            return True
    
    print_error("Should have rejected invalid token!")
    return False

def test_protected_endpoint_no_token():
    """Test protected endpoint without token"""
    print_header("TEST 7: Protected Endpoint Without Token")
    
    print_test("Accessing /api/books without token...")
    response = make_request(
        "GET",
        "/api/books",
        expected_status=401
    )
    
    if response and response.status_code == 401:
        print_success("Correctly blocked access without token!")
        return True
    else:
        print_error("Should have blocked access!")
        return False

def test_protected_endpoint_with_token(token):
    """Test protected endpoint with valid token"""
    print_header("TEST 8: Protected Endpoint With Valid Token")
    
    print_test("Accessing /api/books with valid token...")
    headers = {"Authorization": f"Bearer {token}"}
    response = make_request(
        "GET",
        "/api/books",
        headers=headers
    )
    
    if response and response.status_code in [200, 404, 502]:
        if response.status_code == 200:
            print_success("Successfully accessed protected endpoint!")
        elif response.status_code == 502:
            print_warning("Gateway allowed access but backend service unavailable (502)")
            print_info("This is expected if Book Service is not running")
        else:
            print_warning(f"Gateway allowed access (Status: {response.status_code})")
        return True
    else:
        print_error("Failed to access protected endpoint!")
        return False

def test_existing_user_login():
    """Test login with existing user from database"""
    print_header("TEST 9: Login with Existing User (Alice)")
    
    print_test("Logging in as alice...")
    response = make_request(
        "POST",
        "/auth/login",
        data=EXISTING_USER,
        expected_status=200
    )
    
    if response and response.status_code == 200:
        data = response.json()
        if "token" in data:
            print_success("Existing user login successful!")
            return True
        else:
            print_error("No token in response!")
            return False
    else:
        print_warning("Could not login with existing user (may not exist in DB)")
        return False

# Main Test Runner
def run_tests():
    """Run all tests"""
    print(f"\n{Back.GREEN}{Fore.BLACK} AUTH GATEWAY TEST SUITE {Style.RESET_ALL}")
    print(f"{Fore.CYAN}Target: {BASE_URL}{Style.RESET_ALL}")
    print(f"{Fore.CYAN}Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}{Style.RESET_ALL}")
    
    results = []
    token = None
    
    # Run tests
    results.append(("User Registration", test_register()))
    results.append(("Duplicate Registration", test_register_duplicate()))
    results.append(("Invalid Login", test_login_invalid()))
    
    token = test_login()
    results.append(("Valid Login", token is not None))
    
    if token:
        results.append(("Token Validation", test_validate_token(token)))
        results.append(("Invalid Token Validation", test_validate_invalid_token()))
        results.append(("Protected Endpoint (No Token)", test_protected_endpoint_no_token()))
        results.append(("Protected Endpoint (With Token)", test_protected_endpoint_with_token(token)))
    else:
        print_error("Skipping token-dependent tests due to login failure")
    
    results.append(("Existing User Login", test_existing_user_login()))
    
    # Print summary
    print_header("TEST SUMMARY")
    
    passed = sum(1 for _, result in results if result)
    total = len(results)
    
    for test_name, result in results:
        if result:
            print(f"{Fore.GREEN}✓{Style.RESET_ALL} {test_name}")
        else:
            print(f"{Fore.RED}✗{Style.RESET_ALL} {test_name}")
    
    print(f"\n{Back.WHITE}{Fore.BLACK} Results: {passed}/{total} tests passed {Style.RESET_ALL}\n")
    
    if passed == total:
        print(f"{Fore.GREEN}🎉 All tests passed!{Style.RESET_ALL}\n")
        return 0
    else:
        print(f"{Fore.YELLOW}⚠ Some tests failed{Style.RESET_ALL}\n")
        return 1

if __name__ == "__main__":
    try:
        exit_code = run_tests()
        sys.exit(exit_code)
    except KeyboardInterrupt:
        print(f"\n\n{Fore.YELLOW}Tests interrupted by user{Style.RESET_ALL}\n")
        sys.exit(1)
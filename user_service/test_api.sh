#!/bin/bash

### ------------------------------------
### Pretty colors
### ------------------------------------
GREEN="\e[32m"
RED="\e[31m"
BLUE="\e[34m"
YELLOW="\e[33m"
NC="\e[0m"

BASE_URL="http://localhost:8082/api/users"

print_title() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

pass() {
    echo -e "  ${GREEN}✔ PASS${NC} $1"
}

fail() {
    echo -e "  ${RED}✘ FAIL${NC} $1"
    exit 1
}

### ------------------------------------
### Test Helpers
### ------------------------------------
request() {
    METHOD=$1
    URL=$2
    DATA=$3

    if [ -n "$DATA" ]; then
        RESPONSE=$(curl -s -w "\n%{http_code}" -X $METHOD -H "Content-Type: application/json" -d "$DATA" "$URL")
    else
        RESPONSE=$(curl -s -w "\n%{http_code}" -X $METHOD "$URL")
    fi

    BODY=$(echo "$RESPONSE" | head -n1)
    STATUS=$(echo "$RESPONSE" | tail -n1)

    echo "$BODY"
    echo "$STATUS"
}

assert_status() {
    [ "$1" -eq "$2" ] && pass "HTTP $1 == $2" || fail "expected $2, got $1"
}

### ------------------------------------
### TESTS START
### ------------------------------------

print_title "1. CREATE USER"
CREATE_DATA='{
  "username": "sam",
  "email": "sam@test.com",
  "firstName": "Sam",
  "lastName": "Smith"
}'

OUT=$(request POST "$BASE_URL" "$CREATE_DATA")
BODY=$(echo "$OUT" | head -n1)
STATUS=$(echo "$OUT" | tail -n1)

assert_status "$STATUS" 201

USER_ID=$(echo "$BODY" | jq -r '.id' 2>/dev/null)
if [ "$USER_ID" == "null" ] || [ -z "$USER_ID" ]; then
    fail "User ID not returned in CREATE"
else
    pass "User created with ID: $USER_ID"
fi


print_title "2. GET USER BY ID"

OUT=$(request GET "$BASE_URL/$USER_ID")
BODY=$(echo "$OUT" | head -n1)
STATUS=$(echo "$OUT" | tail -n1)

assert_status "$STATUS" 200
pass "Fetched user: $(echo "$BODY")"


print_title "3. UPDATE USER"

UPDATE_DATA='{
  "username": "neo",
  "email": "neo@test.com",
  "firstName": "Neo",
  "lastName": "Matrix"
}'

OUT=$(request PUT "$BASE_URL/$USER_ID" "$UPDATE_DATA")
STATUS=$(echo "$OUT" | tail -n1)

assert_status "$STATUS" 200
pass "User updated"


print_title "4. GET ALL USERS"

OUT=$(request GET "$BASE_URL")
STATUS=$(echo "$OUT" | tail -n1)

assert_status "$STATUS" 200
pass "Users list retrieved"


print_title "5. DELETE USER"

OUT=$(request DELETE "$BASE_URL/$USER_ID")
STATUS=$(echo "$OUT" | tail -n1)

assert_status "$STATUS" 204
pass "User deleted"


print_title "6. VERIFY USER IS GONE"

OUT=$(request GET "$BASE_URL/$USER_ID")
STATUS=$(echo "$OUT" | tail -n1)

assert_status "$STATUS" 404
pass "User not found (as expected)"

echo -e "\n${GREEN}✨ ALL TESTS PASSED SUCCESSFULLY ✨${NC}\n"

#!/bin/bash

# Contract Testing Script for User Service
# Tests all endpoints according to the contract specification

set -e

BASE_URL="http://localhost:8082/api/v1"
PASSED=0
FAILED=0

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=================================="
echo "User Service Contract Testing"
echo "=================================="
echo ""

# Helper function to test endpoint
test_endpoint() {
    local test_name="$1"
    local expected_status="$2"
    local response="$3"
    local actual_status=$(echo "$response" | tail -n1)

    if [ "$actual_status" = "$expected_status" ]; then
        echo -e "${GREEN}✓${NC} $test_name (HTTP $actual_status)"
        ((PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name (Expected $expected_status, got $actual_status)"
        ((FAILED++))
    fi
}

# Test 1: Health Check
echo -e "${YELLOW}[1] Health Check${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" http://localhost:8082/health)
test_endpoint "Health endpoint" "200" "$RESPONSE"
echo ""

# Test 2: Create User (Valid)
echo -e "${YELLOW}[2] Create User - Valid Data${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@test.com"}')
test_endpoint "Create user with valid data" "201" "$RESPONSE"

# Extract user ID for later tests
USER_ID=$(echo "$RESPONSE" | head -n1 | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "  Created user ID: $USER_ID"
echo ""

# Test 3: Email Normalization
echo -e "${YELLOW}[3] Email Normalization${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe", "email": "JANE@TEST.COM"}')
test_endpoint "Create user with uppercase email" "201" "$RESPONSE"

EMAIL=$(echo "$RESPONSE" | head -n1 | grep -o '"email":"[^"]*"' | cut -d'"' -f4)
if [ "$EMAIL" = "jane@test.com" ]; then
    echo -e "  ${GREEN}✓${NC} Email correctly normalized to lowercase: $EMAIL"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} Email not normalized (got: $EMAIL)"
    ((FAILED++))
fi
echo ""

# Test 4: Duplicate Email
echo -e "${YELLOW}[4] Duplicate Email (Edge Case)${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Another John", "email": "john@test.com"}')
test_endpoint "Reject duplicate email" "409" "$RESPONSE"

ERROR_MSG=$(echo "$RESPONSE" | head -n1 | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
if [ "$ERROR_MSG" = "email already exists" ]; then
    echo -e "  ${GREEN}✓${NC} Correct error message: $ERROR_MSG"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} Incorrect error message: $ERROR_MSG"
    ((FAILED++))
fi
echo ""

# Test 5: Invalid Email Format
echo -e "${YELLOW}[5] Invalid Email Format (Edge Case)${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "not-an-email"}')
test_endpoint "Reject invalid email format" "400" "$RESPONSE"

ERROR_MSG=$(echo "$RESPONSE" | head -n1 | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
if [ "$ERROR_MSG" = "invalid email format" ]; then
    echo -e "  ${GREEN}✓${NC} Correct error message: $ERROR_MSG"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} Incorrect error message: $ERROR_MSG"
    ((FAILED++))
fi
echo ""

# Test 6: Empty Name
echo -e "${YELLOW}[6] Empty Name (Edge Case)${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name": "", "email": "empty@test.com"}')
test_endpoint "Reject empty name" "400" "$RESPONSE"

ERROR_MSG=$(echo "$RESPONSE" | head -n1 | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
if [ "$ERROR_MSG" = "name is required" ]; then
    echo -e "  ${GREEN}✓${NC} Correct error message: $ERROR_MSG"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} Incorrect error message: $ERROR_MSG"
    ((FAILED++))
fi
echo ""

# Test 7: Get User by ID
echo -e "${YELLOW}[7] Get User by ID${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET $BASE_URL/users/$USER_ID)
test_endpoint "Get existing user" "200" "$RESPONSE"

RETRIEVED_ID=$(echo "$RESPONSE" | head -n1 | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ "$RETRIEVED_ID" = "$USER_ID" ]; then
    echo -e "  ${GREEN}✓${NC} Retrieved correct user ID"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} Retrieved wrong user ID"
    ((FAILED++))
fi
echo ""

# Test 8: Get Non-Existent User
echo -e "${YELLOW}[8] Get Non-Existent User (Edge Case)${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET $BASE_URL/users/00000000-0000-0000-0000-000000000000)
test_endpoint "Get non-existent user" "404" "$RESPONSE"

ERROR_MSG=$(echo "$RESPONSE" | head -n1 | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
if [ "$ERROR_MSG" = "user not found" ]; then
    echo -e "  ${GREEN}✓${NC} Correct error message: $ERROR_MSG"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} Incorrect error message: $ERROR_MSG"
    ((FAILED++))
fi
echo ""

# Test 9: List Users
echo -e "${YELLOW}[9] List All Users${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET $BASE_URL/users)
test_endpoint "List all users" "200" "$RESPONSE"

USER_COUNT=$(echo "$RESPONSE" | head -n1 | grep -o '"id":"[^"]*"' | wc -l | tr -d ' ')
echo -e "  Found $USER_COUNT users in the system"
echo ""

# Test 10: Validate User (Valid)
echo -e "${YELLOW}[10] Validate User - Valid ID${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users/validate \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": \"$USER_ID\"}")
test_endpoint "Validate existing user" "200" "$RESPONSE"

VALID=$(echo "$RESPONSE" | head -n1 | grep -o '"valid":[^,}]*' | cut -d':' -f2)
if [ "$VALID" = "true" ]; then
    echo -e "  ${GREEN}✓${NC} User validation returned valid=true"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} User validation returned valid=$VALID"
    ((FAILED++))
fi
echo ""

# Test 11: Validate User (Invalid)
echo -e "${YELLOW}[11] Validate User - Invalid ID${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users/validate \
  -H "Content-Type: application/json" \
  -d '{"user_id": "00000000-0000-0000-0000-000000000000"}')
test_endpoint "Validate non-existent user" "404" "$RESPONSE"

VALID=$(echo "$RESPONSE" | head -n1 | grep -o '"valid":[^,}]*' | cut -d':' -f2)
if [ "$VALID" = "false" ]; then
    echo -e "  ${GREEN}✓${NC} User validation returned valid=false"
    ((PASSED++))
else
    echo -e "  ${RED}✗${NC} User validation returned valid=$VALID"
    ((FAILED++))
fi
echo ""

# Summary
echo "=================================="
echo "Test Summary"
echo "=================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All contract tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some contract tests failed.${NC}"
    exit 1
fi

#!/bin/bash

# Test script for todo-service
# This script assumes both user-service and todo-service are running

set -e

BASE_URL="http://localhost:8081/api/v1"
USER_SERVICE_URL="http://localhost:8080/api/v1"

echo "========================================="
echo "Todo Service Contract Test"
echo "========================================="
echo ""

# Step 1: Create a user first
echo "[1] Creating a test user..."
USER_RESPONSE=$(curl -s -X POST "$USER_SERVICE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com"
  }')

USER_ID=$(echo "$USER_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Created user with ID: $USER_ID"
echo ""

# Step 2: Create a TODO with description
echo "[2] Creating TODO with description..."
TODO1=$(curl -s -X POST "$BASE_URL/todos" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, bread, eggs"
  }')

TODO1_ID=$(echo "$TODO1" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Response: $TODO1"
echo ""

# Step 3: Create a TODO without description
echo "[3] Creating TODO without description..."
TODO2=$(curl -s -X POST "$BASE_URL/todos" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d '{
    "title": "Write report"
  }')

TODO2_ID=$(echo "$TODO2" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Response: $TODO2"
echo ""

# Step 4: List all TODOs
echo "[4] Listing all TODOs..."
curl -s -X GET "$BASE_URL/todos" \
  -H "X-User-ID: $USER_ID" | jq .
echo ""

# Step 5: Get a specific TODO
echo "[5] Getting TODO by ID..."
curl -s -X GET "$BASE_URL/todos/$TODO1_ID" \
  -H "X-User-ID: $USER_ID" | jq .
echo ""

# Step 6: Update TODO status to completed
echo "[6] Updating TODO status to completed..."
curl -s -X PATCH "$BASE_URL/todos/$TODO1_ID" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d '{
    "status": "completed"
  }' | jq .
echo ""

# Step 7: List only pending TODOs
echo "[7] Listing only pending TODOs..."
curl -s -X GET "$BASE_URL/todos?status=pending" \
  -H "X-User-ID: $USER_ID" | jq .
echo ""

# Step 8: List only completed TODOs
echo "[8] Listing only completed TODOs..."
curl -s -X GET "$BASE_URL/todos?status=completed" \
  -H "X-User-ID: $USER_ID" | jq .
echo ""

# Step 9: Test error case - empty title
echo "[9] Testing error case - empty title..."
curl -s -X POST "$BASE_URL/todos" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d '{
    "title": "   "
  }' | jq .
echo ""

# Step 10: Test error case - invalid status
echo "[10] Testing error case - invalid status..."
curl -s -X GET "$BASE_URL/todos?status=invalid" \
  -H "X-User-ID: $USER_ID" | jq .
echo ""

# Step 11: Test error case - TODO not found
echo "[11] Testing error case - TODO not found..."
curl -s -X GET "$BASE_URL/todos/00000000-0000-0000-0000-000000000000" \
  -H "X-User-ID: $USER_ID" | jq .
echo ""

# Step 12: Delete a TODO
echo "[12] Deleting a TODO..."
curl -s -X DELETE "$BASE_URL/todos/$TODO2_ID" \
  -H "X-User-ID: $USER_ID" \
  -w "\nHTTP Status: %{http_code}\n"
echo ""

# Step 13: Verify deletion
echo "[13] Verifying TODO was deleted..."
curl -s -X GET "$BASE_URL/todos/$TODO2_ID" \
  -H "X-User-ID: $USER_ID" | jq .
echo ""

# Step 14: Test user isolation - create another user
echo "[14] Testing user isolation - creating another user..."
USER2_RESPONSE=$(curl -s -X POST "$USER_SERVICE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Another User",
    "email": "another@example.com"
  }')

USER2_ID=$(echo "$USER2_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Created user with ID: $USER2_ID"
echo ""

# Step 15: Verify user2 cannot see user1's TODOs
echo "[15] Verifying user2 cannot see user1's TODOs..."
curl -s -X GET "$BASE_URL/todos" \
  -H "X-User-ID: $USER2_ID" | jq .
echo ""

# Step 16: Verify user2 cannot access user1's TODO by ID
echo "[16] Verifying user2 cannot access user1's TODO by ID..."
curl -s -X GET "$BASE_URL/todos/$TODO1_ID" \
  -H "X-User-ID: $USER2_ID" | jq .
echo ""

echo "========================================="
echo "All tests completed!"
echo "========================================="

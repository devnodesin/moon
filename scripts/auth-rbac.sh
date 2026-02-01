#!/bin/bash
# RBAC (Role-Based Access Control) test script for Moon
# Tests: admin vs user roles, can_write permission
# Supports PREFIX environment variable for custom URL prefixes
# Usage: PREFIX=/api/v1 ./scripts/auth-rbac.sh
# Requires: jq, curl

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Default to empty prefix if not set
PREFIX=${PREFIX:-}

# Base URL
BASE_URL="http://localhost:6006${PREFIX}"

# Test counters
PASSED=0
FAILED=0

# Admin credentials
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-change-me-on-first-login}"

# Variables to store created resources for cleanup
ADMIN_TOKEN=""
USER_TOKEN=""
WRITER_TOKEN=""
TEST_USER_ID=""
TEST_WRITER_ID=""
TEST_COLLECTION_CREATED=false

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Moon RBAC (Role-Based Access Control) Tests${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo "Base URL: ${BASE_URL}"
    echo ""
}

pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    ((PASSED++))
}

fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    echo "   Response: $2"
    ((FAILED++))
}

cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up test data...${NC}"
    
    # Delete test users if they exist
    if [ -n "$TEST_USER_ID" ] && [ -n "$ADMIN_TOKEN" ]; then
        curl -s -X POST "${BASE_URL}/users:destroy?id=${TEST_USER_ID}" \
            -H "Authorization: Bearer ${ADMIN_TOKEN}" > /dev/null 2>&1 || true
        echo "Cleaned up test user: ${TEST_USER_ID}"
    fi
    
    if [ -n "$TEST_WRITER_ID" ] && [ -n "$ADMIN_TOKEN" ]; then
        curl -s -X POST "${BASE_URL}/users:destroy?id=${TEST_WRITER_ID}" \
            -H "Authorization: Bearer ${ADMIN_TOKEN}" > /dev/null 2>&1 || true
        echo "Cleaned up test writer: ${TEST_WRITER_ID}"
    fi
    
    # Delete test collection if created
    if [ "$TEST_COLLECTION_CREATED" = true ] && [ -n "$ADMIN_TOKEN" ]; then
        curl -s -X POST "${BASE_URL}/collections:destroy" \
            -H "Authorization: Bearer ${ADMIN_TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"name": "rbac_test_items"}' > /dev/null 2>&1 || true
        echo "Cleaned up test collection: rbac_test_items"
    fi
}

# Set up cleanup trap
trap cleanup EXIT

# Check for jq
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed${NC}"
    exit 1
fi

print_header

# Authenticate as admin
echo "[0] Authenticating as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth:login" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${ADMIN_USERNAME}\", \"password\": \"${ADMIN_PASSWORD}\"}")

ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo -e "${RED}Failed to authenticate as admin${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo "Admin authenticated successfully"

# Create test users
echo ""
echo "[1] Creating test user (role: user, can_write: false)..."
CREATE_USER_RESPONSE=$(curl -s -X POST "${BASE_URL}/users:create" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "rbac_test_user",
        "email": "rbac_test@example.com",
        "password": "TestPass123",
        "role": "user",
        "can_write": false
    }')

TEST_USER_ID=$(echo "$CREATE_USER_RESPONSE" | jq -r '.id // empty')

if [ -n "$TEST_USER_ID" ] && [ "$TEST_USER_ID" != "null" ]; then
    pass "Created read-only user (ID: ${TEST_USER_ID})"
else
    fail "Failed to create test user" "$CREATE_USER_RESPONSE"
fi

echo ""
echo "[2] Creating test writer (role: user, can_write: true)..."
CREATE_WRITER_RESPONSE=$(curl -s -X POST "${BASE_URL}/users:create" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "rbac_test_writer",
        "email": "rbac_writer@example.com",
        "password": "TestPass123",
        "role": "user",
        "can_write": true
    }')

TEST_WRITER_ID=$(echo "$CREATE_WRITER_RESPONSE" | jq -r '.id // empty')

if [ -n "$TEST_WRITER_ID" ] && [ "$TEST_WRITER_ID" != "null" ]; then
    pass "Created writer user (ID: ${TEST_WRITER_ID})"
else
    fail "Failed to create writer user" "$CREATE_WRITER_RESPONSE"
fi

# Login as test users
echo ""
echo "[3] Logging in as read-only user..."
USER_LOGIN=$(curl -s -X POST "${BASE_URL}/auth:login" \
    -H "Content-Type: application/json" \
    -d '{"username": "rbac_test_user", "password": "TestPass123"}')

USER_TOKEN=$(echo "$USER_LOGIN" | jq -r '.access_token // empty')

if [ -n "$USER_TOKEN" ] && [ "$USER_TOKEN" != "null" ]; then
    pass "Read-only user logged in"
else
    fail "Failed to login as read-only user" "$USER_LOGIN"
fi

echo ""
echo "[4] Logging in as writer user..."
WRITER_LOGIN=$(curl -s -X POST "${BASE_URL}/auth:login" \
    -H "Content-Type: application/json" \
    -d '{"username": "rbac_test_writer", "password": "TestPass123"}')

WRITER_TOKEN=$(echo "$WRITER_LOGIN" | jq -r '.access_token // empty')

if [ -n "$WRITER_TOKEN" ] && [ "$WRITER_TOKEN" != "null" ]; then
    pass "Writer user logged in"
else
    fail "Failed to login as writer user" "$WRITER_LOGIN"
fi

# Create a test collection for data tests
echo ""
echo "[5] Creating test collection (as admin)..."
CREATE_COLL_RESPONSE=$(curl -s -X POST "${BASE_URL}/collections:create" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "rbac_test_items",
        "columns": [
            {"name": "title", "type": "string", "required": true},
            {"name": "value", "type": "integer", "required": false}
        ]
    }')

if echo "$CREATE_COLL_RESPONSE" | jq -e '.name == "rbac_test_items"' > /dev/null 2>&1; then
    pass "Test collection created"
    TEST_COLLECTION_CREATED=true
else
    fail "Failed to create test collection" "$CREATE_COLL_RESPONSE"
fi

# Test: Admin can manage users
echo ""
echo "[6] Testing: Admin can list users..."
ADMIN_LIST_USERS=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/users:list" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}")
ADMIN_LIST_STATUS=$(echo "$ADMIN_LIST_USERS" | tail -n1)

if [ "$ADMIN_LIST_STATUS" = "200" ]; then
    pass "Admin can list users"
else
    fail "Admin should be able to list users, got ${ADMIN_LIST_STATUS}" "$(echo "$ADMIN_LIST_USERS" | sed '$d')"
fi

# Test: User cannot manage users
echo ""
echo "[7] Testing: User cannot list users (403 Forbidden)..."
USER_LIST_USERS=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/users:list" \
    -H "Authorization: Bearer ${USER_TOKEN}")
USER_LIST_STATUS=$(echo "$USER_LIST_USERS" | tail -n1)

if [ "$USER_LIST_STATUS" = "403" ]; then
    pass "User correctly denied access to user management"
else
    fail "User should get 403 for user management, got ${USER_LIST_STATUS}" "$(echo "$USER_LIST_USERS" | sed '$d')"
fi

# Test: Admin can manage collections
echo ""
echo "[8] Testing: Admin can update collections..."
ADMIN_UPDATE_COLL=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/collections:update" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "rbac_test_items",
        "add_columns": [{"name": "extra", "type": "string", "required": false}]
    }')
ADMIN_UPDATE_STATUS=$(echo "$ADMIN_UPDATE_COLL" | tail -n1)

if [ "$ADMIN_UPDATE_STATUS" = "200" ]; then
    pass "Admin can update collection schema"
else
    fail "Admin should be able to update collections, got ${ADMIN_UPDATE_STATUS}" "$(echo "$ADMIN_UPDATE_COLL" | sed '$d')"
fi

# Test: User cannot manage collections
echo ""
echo "[9] Testing: User cannot create collections (403 Forbidden)..."
USER_CREATE_COLL=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/collections:create" \
    -H "Authorization: Bearer ${USER_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"name": "user_attempt", "columns": [{"name": "test", "type": "string"}]}')
USER_CREATE_STATUS=$(echo "$USER_CREATE_COLL" | tail -n1)

if [ "$USER_CREATE_STATUS" = "403" ]; then
    pass "User correctly denied collection creation"
else
    fail "User should get 403 for collection creation, got ${USER_CREATE_STATUS}" "$(echo "$USER_CREATE_COLL" | sed '$d')"
fi

# Test: Both users can read collections
echo ""
echo "[10] Testing: Read-only user can read collections..."
USER_READ_COLL=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/collections:list" \
    -H "Authorization: Bearer ${USER_TOKEN}")
USER_READ_STATUS=$(echo "$USER_READ_COLL" | tail -n1)

if [ "$USER_READ_STATUS" = "200" ]; then
    pass "Read-only user can read collection metadata"
else
    fail "User should be able to read collections, got ${USER_READ_STATUS}" "$(echo "$USER_READ_COLL" | sed '$d')"
fi

# Test: Read-only user cannot write data
echo ""
echo "[11] Testing: Read-only user cannot create data (403 Forbidden)..."
USER_CREATE_DATA=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/rbac_test_items:create" \
    -H "Authorization: Bearer ${USER_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"data": {"title": "User Item", "value": 100}}')
USER_CREATE_DATA_STATUS=$(echo "$USER_CREATE_DATA" | tail -n1)

if [ "$USER_CREATE_DATA_STATUS" = "403" ]; then
    pass "Read-only user correctly denied data creation"
else
    fail "Read-only user should get 403 for data creation, got ${USER_CREATE_DATA_STATUS}" "$(echo "$USER_CREATE_DATA" | sed '$d')"
fi

# Test: Writer user can write data
echo ""
echo "[12] Testing: Writer user can create data..."
WRITER_CREATE_DATA=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/rbac_test_items:create" \
    -H "Authorization: Bearer ${WRITER_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"data": {"title": "Writer Item", "value": 200}}')
WRITER_CREATE_STATUS=$(echo "$WRITER_CREATE_DATA" | tail -n1)
WRITER_ITEM_ID=$(echo "$WRITER_CREATE_DATA" | sed '$d' | jq -r '.id // empty')

if [ "$WRITER_CREATE_STATUS" = "201" ] || [ "$WRITER_CREATE_STATUS" = "200" ]; then
    pass "Writer user can create data"
else
    fail "Writer should be able to create data, got ${WRITER_CREATE_STATUS}" "$(echo "$WRITER_CREATE_DATA" | sed '$d')"
fi

# Test: Read-only user can read data
echo ""
echo "[13] Testing: Read-only user can read data..."
USER_READ_DATA=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/rbac_test_items:list" \
    -H "Authorization: Bearer ${USER_TOKEN}")
USER_READ_DATA_STATUS=$(echo "$USER_READ_DATA" | tail -n1)

if [ "$USER_READ_DATA_STATUS" = "200" ]; then
    pass "Read-only user can read data"
else
    fail "User should be able to read data, got ${USER_READ_DATA_STATUS}" "$(echo "$USER_READ_DATA" | sed '$d')"
fi

# Test: Writer user can update data
echo ""
echo "[14] Testing: Writer user can update data..."
if [ -n "$WRITER_ITEM_ID" ] && [ "$WRITER_ITEM_ID" != "null" ]; then
    WRITER_UPDATE_DATA=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/rbac_test_items:update" \
        -H "Authorization: Bearer ${WRITER_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"id\": \"${WRITER_ITEM_ID}\", \"data\": {\"value\": 300}}")
    WRITER_UPDATE_STATUS=$(echo "$WRITER_UPDATE_DATA" | tail -n1)

    if [ "$WRITER_UPDATE_STATUS" = "200" ]; then
        pass "Writer user can update data"
    else
        fail "Writer should be able to update data, got ${WRITER_UPDATE_STATUS}" "$(echo "$WRITER_UPDATE_DATA" | sed '$d')"
    fi
else
    fail "No item ID to test update" ""
fi

# Test: Read-only user cannot update data
echo ""
echo "[15] Testing: Read-only user cannot update data (403 Forbidden)..."
if [ -n "$WRITER_ITEM_ID" ] && [ "$WRITER_ITEM_ID" != "null" ]; then
    USER_UPDATE_DATA=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/rbac_test_items:update" \
        -H "Authorization: Bearer ${USER_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"id\": \"${WRITER_ITEM_ID}\", \"data\": {\"value\": 400}}")
    USER_UPDATE_STATUS=$(echo "$USER_UPDATE_DATA" | tail -n1)

    if [ "$USER_UPDATE_STATUS" = "403" ]; then
        pass "Read-only user correctly denied data update"
    else
        fail "Read-only user should get 403 for data update, got ${USER_UPDATE_STATUS}" "$(echo "$USER_UPDATE_DATA" | sed '$d')"
    fi
fi

# Test: Writer user can delete data
echo ""
echo "[16] Testing: Writer user can delete data..."
if [ -n "$WRITER_ITEM_ID" ] && [ "$WRITER_ITEM_ID" != "null" ]; then
    WRITER_DELETE_DATA=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/rbac_test_items:destroy" \
        -H "Authorization: Bearer ${WRITER_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"id\": \"${WRITER_ITEM_ID}\"}")
    WRITER_DELETE_STATUS=$(echo "$WRITER_DELETE_DATA" | tail -n1)

    if [ "$WRITER_DELETE_STATUS" = "200" ]; then
        pass "Writer user can delete data"
    else
        fail "Writer should be able to delete data, got ${WRITER_DELETE_STATUS}" "$(echo "$WRITER_DELETE_DATA" | sed '$d')"
    fi
fi

# Test: User cannot manage API keys
echo ""
echo "[17] Testing: User cannot create API keys (403 Forbidden)..."
USER_CREATE_KEY=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/apikeys:create" \
    -H "Authorization: Bearer ${USER_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"name": "user_attempt_key", "role": "user"}')
USER_CREATE_KEY_STATUS=$(echo "$USER_CREATE_KEY" | tail -n1)

if [ "$USER_CREATE_KEY_STATUS" = "403" ]; then
    pass "User correctly denied API key creation"
else
    fail "User should get 403 for API key creation, got ${USER_CREATE_KEY_STATUS}" "$(echo "$USER_CREATE_KEY" | sed '$d')"
fi

# Test: Admin can update user permissions
echo ""
echo "[18] Testing: Admin can update user permissions..."
ADMIN_UPDATE_USER=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/users:update?id=${TEST_USER_ID}" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"can_write": true}')
ADMIN_UPDATE_USER_STATUS=$(echo "$ADMIN_UPDATE_USER" | tail -n1)

if [ "$ADMIN_UPDATE_USER_STATUS" = "200" ]; then
    pass "Admin can update user permissions"
else
    fail "Admin should be able to update user, got ${ADMIN_UPDATE_USER_STATUS}" "$(echo "$ADMIN_UPDATE_USER" | sed '$d')"
fi

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}RBAC Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All RBAC tests passed! ✅${NC}"
    exit 0
else
    echo -e "${RED}Some RBAC tests failed! ❌${NC}"
    exit 1
fi

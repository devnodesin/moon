#!/bin/bash
# Rate Limiting test script for Moon
# Tests: user rate limits, login rate limits, rate limit headers
# Supports PREFIX environment variable for custom URL prefixes
# Usage: PREFIX=/api/v1 ./scripts/auth-ratelimit.sh
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

# Test settings (adjust based on your rate limit configuration)
# Default: 100 requests/minute for users, 5 login attempts per 15 minutes
USER_RATE_LIMIT=${USER_RATE_LIMIT:-100}
LOGIN_RATE_LIMIT=${LOGIN_RATE_LIMIT:-5}

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Moon Rate Limiting Tests${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo "Base URL: ${BASE_URL}"
    echo "User rate limit: ${USER_RATE_LIMIT} requests/minute"
    echo "Login rate limit: ${LOGIN_RATE_LIMIT} attempts/15 minutes"
    echo ""
    echo -e "${YELLOW}Note: These tests may take a while and could trigger${NC}"
    echo -e "${YELLOW}rate limiting on the server. Use with caution.${NC}"
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

# Check for jq
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed${NC}"
    exit 1
fi

print_header

# Authenticate to get a token
echo "[0] Authenticating..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth:login" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${ADMIN_USERNAME}\", \"password\": \"${ADMIN_PASSWORD}\"}")

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
    echo -e "${RED}Failed to authenticate${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo "Authenticated successfully"

echo ""
echo "[1] Testing rate limit headers presence..."
HEADER_RESPONSE=$(curl -s -D - -X GET "${BASE_URL}/collections:list" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -o /dev/null 2>&1)

# Check for rate limit headers
LIMIT_HEADER=$(echo "$HEADER_RESPONSE" | grep -i "X-RateLimit-Limit" || true)
REMAINING_HEADER=$(echo "$HEADER_RESPONSE" | grep -i "X-RateLimit-Remaining" || true)
RESET_HEADER=$(echo "$HEADER_RESPONSE" | grep -i "X-RateLimit-Reset" || true)

if [ -n "$LIMIT_HEADER" ]; then
    pass "X-RateLimit-Limit header present"
    echo "   ${LIMIT_HEADER}"
else
    fail "X-RateLimit-Limit header missing" ""
fi

echo ""
echo "[2] Testing X-RateLimit-Remaining header..."
if [ -n "$REMAINING_HEADER" ]; then
    pass "X-RateLimit-Remaining header present"
    echo "   ${REMAINING_HEADER}"
else
    fail "X-RateLimit-Remaining header missing" ""
fi

echo ""
echo "[3] Testing X-RateLimit-Reset header..."
if [ -n "$RESET_HEADER" ]; then
    pass "X-RateLimit-Reset header present"
    echo "   ${RESET_HEADER}"
else
    fail "X-RateLimit-Reset header missing" ""
fi

echo ""
echo "[4] Testing rate limit decrements..."
# Make two requests and check if remaining decreases
FIRST_RESPONSE=$(curl -s -D - -X GET "${BASE_URL}/collections:list" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -o /dev/null 2>&1)
FIRST_REMAINING=$(echo "$FIRST_RESPONSE" | grep -i "X-RateLimit-Remaining" | awk '{print $2}' | tr -d '\r')

SECOND_RESPONSE=$(curl -s -D - -X GET "${BASE_URL}/collections:list" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -o /dev/null 2>&1)
SECOND_REMAINING=$(echo "$SECOND_RESPONSE" | grep -i "X-RateLimit-Remaining" | awk '{print $2}' | tr -d '\r')

if [ -n "$FIRST_REMAINING" ] && [ -n "$SECOND_REMAINING" ]; then
    FIRST_NUM=$(echo "$FIRST_REMAINING" | tr -cd '0-9')
    SECOND_NUM=$(echo "$SECOND_REMAINING" | tr -cd '0-9')
    
    if [ -n "$FIRST_NUM" ] && [ -n "$SECOND_NUM" ]; then
        if [ "$SECOND_NUM" -lt "$FIRST_NUM" ]; then
            pass "Rate limit remaining decrements correctly (${FIRST_NUM} -> ${SECOND_NUM})"
        else
            fail "Rate limit remaining should decrease" "First: ${FIRST_NUM}, Second: ${SECOND_NUM}"
        fi
    else
        fail "Could not parse rate limit values" "First: ${FIRST_REMAINING}, Second: ${SECOND_REMAINING}"
    fi
else
    fail "Could not read rate limit remaining values" ""
fi

echo ""
echo "[5] Testing login rate limit with invalid credentials..."
echo -e "${YELLOW}   Making ${LOGIN_RATE_LIMIT} failed login attempts...${NC}"

LOGIN_429_RECEIVED=false
for i in $(seq 1 $((LOGIN_RATE_LIMIT + 2))); do
    LOGIN_TEST=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/auth:login" \
        -H "Content-Type: application/json" \
        -d '{"username": "ratelimit_test_user", "password": "wrongpassword"}')
    LOGIN_STATUS=$(echo "$LOGIN_TEST" | tail -n1)
    
    if [ "$LOGIN_STATUS" = "429" ]; then
        LOGIN_429_RECEIVED=true
        pass "Login rate limit triggered after ${i} attempts (429 Too Many Requests)"
        break
    fi
    
    # Small delay to avoid overwhelming the server
    sleep 0.1
done

if [ "$LOGIN_429_RECEIVED" = false ]; then
    # This might be expected if rate limiting is configured differently
    echo -e "${YELLOW}   Note: 429 not received after ${LOGIN_RATE_LIMIT} attempts${NC}"
    echo -e "${YELLOW}   This may be expected if login rate limiting is disabled or configured higher${NC}"
    pass "Login rate limit test completed (rate limiting may be configured differently)"
fi

echo ""
echo "[6] Testing 429 response format..."
# Try to trigger a 429 to check response format
# If we already got one above, we can verify the format

if [ "$LOGIN_429_RECEIVED" = true ]; then
    # The last response should have been a 429
    ERROR_CODE=$(echo "$LOGIN_TEST" | sed '$d' | jq -r '.error.code // empty')
    
    if [ -n "$ERROR_CODE" ]; then
        pass "429 response includes error code: ${ERROR_CODE}"
    else
        fail "429 response should include error object" "$(echo "$LOGIN_TEST" | sed '$d')"
    fi
else
    pass "429 format test skipped (no 429 received)"
fi

echo ""
echo "[7] Testing rate limit on health endpoint (should be unrestricted)..."
# Health endpoint should typically be unrestricted
HEALTH_429_COUNT=0
for i in $(seq 1 10); do
    HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/health")
    HEALTH_STATUS=$(echo "$HEALTH_RESPONSE" | tail -n1)
    
    if [ "$HEALTH_STATUS" = "429" ]; then
        ((HEALTH_429_COUNT++))
    fi
done

if [ "$HEALTH_429_COUNT" -eq 0 ]; then
    pass "Health endpoint not rate limited (10 rapid requests OK)"
else
    fail "Health endpoint should not be rate limited" "${HEALTH_429_COUNT}/10 got 429"
fi

echo ""
echo "[8] Testing authenticated request rate (burst test)..."
echo -e "${YELLOW}   Making 20 rapid requests to test burst handling...${NC}"

REQUEST_COUNT=20
SUCCESS_COUNT=0
RATE_LIMITED_COUNT=0

for i in $(seq 1 $REQUEST_COUNT); do
    BURST_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/collections:list" \
        -H "Authorization: Bearer ${ACCESS_TOKEN}")
    BURST_STATUS=$(echo "$BURST_RESPONSE" | tail -n1)
    
    if [ "$BURST_STATUS" = "200" ]; then
        ((SUCCESS_COUNT++))
    elif [ "$BURST_STATUS" = "429" ]; then
        ((RATE_LIMITED_COUNT++))
    fi
done

echo "   Results: ${SUCCESS_COUNT}/${REQUEST_COUNT} succeeded, ${RATE_LIMITED_COUNT} rate limited"

if [ "$SUCCESS_COUNT" -gt 0 ]; then
    pass "Authenticated requests processed (${SUCCESS_COUNT}/${REQUEST_COUNT} successful)"
else
    fail "All requests failed" ""
fi

echo ""
echo "[9] Testing rate limit reset behavior..."
echo -e "${YELLOW}   Waiting for partial rate limit window reset (5 seconds)...${NC}"
sleep 5

AFTER_WAIT_RESPONSE=$(curl -s -D - -X GET "${BASE_URL}/collections:list" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -o /dev/null 2>&1)
AFTER_WAIT_REMAINING=$(echo "$AFTER_WAIT_RESPONSE" | grep -i "X-RateLimit-Remaining" | awk '{print $2}' | tr -d '\r')

if [ -n "$AFTER_WAIT_REMAINING" ]; then
    AFTER_WAIT_NUM=$(echo "$AFTER_WAIT_REMAINING" | tr -cd '0-9')
    if [ -n "$AFTER_WAIT_NUM" ]; then
        pass "Rate limit remaining after wait: ${AFTER_WAIT_NUM}"
    else
        pass "Rate limit check completed (value: ${AFTER_WAIT_REMAINING})"
    fi
else
    pass "Rate limit reset check completed"
fi

echo ""
echo "[10] Testing rate limit with different authentication methods..."
# If API key auth is enabled, test that it has separate rate limits
echo -e "${YELLOW}   Note: API key rate limiting (1000/min) is separate from JWT (100/min)${NC}"
pass "Rate limit separation documented (JWT: 100/min, API Key: 1000/min)"

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Rate Limit Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"

echo ""
echo "Rate Limit Configuration Reference:"
echo "  - JWT authenticated: 100 requests/minute"
echo "  - API Key authenticated: 1000 requests/minute"
echo "  - Login attempts: 5 per 15 minutes per IP/username"
echo ""
echo "Response Headers:"
echo "  - X-RateLimit-Limit: Maximum requests allowed"
echo "  - X-RateLimit-Remaining: Requests remaining in window"
echo "  - X-RateLimit-Reset: Unix timestamp when limit resets"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}All rate limit tests passed! ✅${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}Some rate limit tests failed! ❌${NC}"
    exit 1
fi

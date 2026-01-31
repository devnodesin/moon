#!/bin/bash
# Data pagination test script for Moon
# Supports PREFIX environment variable for custom URL prefixes
# Usage: PREFIX=/api/v1 ./scripts/data-paginate.sh
# Usage: PREFIX="" ./scripts/data-paginate.sh (for no prefix)

# Default to empty prefix if not set
PREFIX=${PREFIX:-}

# Base URL
BASE_URL="http://localhost:6006${PREFIX}"

echo "Testing Moon Data Pagination API"
echo "Using prefix: ${PREFIX:-<empty>}"
echo "Base URL: ${BASE_URL}"
echo ""

echo "[1] List products with limit=1 (first page):"
RESPONSE=$(curl -s -X GET "${BASE_URL}/products:list?limit=1")
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
echo ""

# Extract cursor from response
CURSOR=$(echo "$RESPONSE" | jq -r '.next_cursor // empty' 2>/dev/null)

if [ -n "$CURSOR" ] && [ "$CURSOR" != "null" ]; then
    echo "[2] List products with limit=1 and after=$CURSOR (second page):"
    RESPONSE2=$(curl -s -X GET "${BASE_URL}/products:list?limit=1&after=${CURSOR}")
    echo "$RESPONSE2" | jq . 2>/dev/null || echo "$RESPONSE2"
    echo ""
    
    # Extract second cursor
    CURSOR2=$(echo "$RESPONSE2" | jq -r '.next_cursor // empty' 2>/dev/null)
    
    if [ -n "$CURSOR2" ] && [ "$CURSOR2" != "null" ]; then
        echo "[3] List products with limit=1 and after=$CURSOR2 (third page):"
        curl -s -X GET "${BASE_URL}/products:list?limit=1&after=${CURSOR2}" | jq . 2>/dev/null || curl -s -X GET "${BASE_URL}/products:list?limit=1&after=${CURSOR2}"
        echo ""
    fi
else
    echo "No more pages available (next_cursor is null)"
fi

echo "Pagination test complete!"


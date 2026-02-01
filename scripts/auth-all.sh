#!/bin/bash
# Master authentication test runner for Moon
# Runs all auth tests: JWT, API Key, RBAC, Rate Limiting
# Supports PREFIX environment variable for custom URL prefixes
# Usage: PREFIX=/api/v1 ./scripts/auth-all.sh
# Usage: ./scripts/auth-all.sh [jwt|apikey|rbac|ratelimit]
# Requires: jq, curl

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Track results
TOTAL_PASSED=0
TOTAL_FAILED=0
declare -A TEST_RESULTS

print_banner() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                                                              ║${NC}"
    echo -e "${BLUE}║          Moon Authentication Test Suite                      ║${NC}"
    echo -e "${BLUE}║                                                              ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Base URL: http://localhost:6006${PREFIX:-}"
    echo "Scripts directory: ${SCRIPT_DIR}"
    echo ""
}

run_test() {
    local test_name=$1
    local script_name=$2
    local description=$3
    
    echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"
    echo -e "${BLUE}Running: ${description}${NC}"
    echo -e "${BLUE}Script: ${script_name}${NC}"
    echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"
    echo ""
    
    if [ -f "${SCRIPT_DIR}/${script_name}" ]; then
        if bash "${SCRIPT_DIR}/${script_name}"; then
            TEST_RESULTS[$test_name]="PASS"
            echo ""
            echo -e "${GREEN}✅ ${description} completed successfully${NC}"
        else
            TEST_RESULTS[$test_name]="FAIL"
            echo ""
            echo -e "${RED}❌ ${description} failed${NC}"
        fi
    else
        echo -e "${RED}Script not found: ${SCRIPT_DIR}/${script_name}${NC}"
        TEST_RESULTS[$test_name]="MISSING"
    fi
    
    echo ""
}

print_summary() {
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                    Test Suite Summary                        ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    local pass_count=0
    local fail_count=0
    
    for test_name in "${!TEST_RESULTS[@]}"; do
        local result=${TEST_RESULTS[$test_name]}
        if [ "$result" = "PASS" ]; then
            echo -e "  ${GREEN}✅ ${test_name}: PASSED${NC}"
            ((pass_count++))
        elif [ "$result" = "FAIL" ]; then
            echo -e "  ${RED}❌ ${test_name}: FAILED${NC}"
            ((fail_count++))
        else
            echo -e "  ${YELLOW}⚠️  ${test_name}: ${result}${NC}"
            ((fail_count++))
        fi
    done
    
    echo ""
    echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"
    echo -e "  Total: ${#TEST_RESULTS[@]} test suites"
    echo -e "  ${GREEN}Passed: ${pass_count}${NC}"
    echo -e "  ${RED}Failed: ${fail_count}${NC}"
    echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"
    echo ""
    
    if [ $fail_count -eq 0 ]; then
        echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║           All authentication tests passed! ✅               ║${NC}"
        echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
        return 0
    else
        echo -e "${RED}╔══════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${RED}║           Some authentication tests failed! ❌               ║${NC}"
        echo -e "${RED}╚══════════════════════════════════════════════════════════════╝${NC}"
        return 1
    fi
}

print_usage() {
    echo "Usage: $0 [OPTIONS] [TESTS...]"
    echo ""
    echo "Run Moon authentication tests."
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo ""
    echo "Tests (run all if none specified):"
    echo "  jwt            JWT authentication tests"
    echo "  apikey         API key authentication tests"
    echo "  rbac           Role-based access control tests"
    echo "  ratelimit      Rate limiting tests"
    echo ""
    echo "Environment Variables:"
    echo "  PREFIX         URL prefix (e.g., /api/v1)"
    echo "  ADMIN_USERNAME Admin username (default: admin)"
    echo "  ADMIN_PASSWORD Admin password (default: change-me-on-first-login)"
    echo ""
    echo "Examples:"
    echo "  $0                          # Run all tests"
    echo "  $0 jwt                      # Run only JWT tests"
    echo "  $0 jwt rbac                 # Run JWT and RBAC tests"
    echo "  PREFIX=/api/v1 $0           # Run all tests with prefix"
    echo ""
}

# Check for help flag
if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
    print_usage
    exit 0
fi

# Check for jq
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed${NC}"
    echo "Install with: apt-get install jq (Debian/Ubuntu) or brew install jq (macOS)"
    exit 1
fi

# Check for curl
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required but not installed${NC}"
    exit 1
fi

print_banner

# Check if Moon is running
echo "Checking if Moon is running..."
if curl -s --connect-timeout 5 "http://localhost:6006${PREFIX:-}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}Moon is running${NC}"
else
    echo -e "${RED}Error: Cannot connect to Moon at http://localhost:6006${PREFIX:-}/health${NC}"
    echo ""
    echo "Please ensure Moon is running with authentication enabled."
    echo "See INSTALL.md for setup instructions."
    exit 1
fi
echo ""

# Determine which tests to run
TESTS_TO_RUN=()

if [ $# -eq 0 ]; then
    # Run all tests
    TESTS_TO_RUN=("jwt" "apikey" "rbac" "ratelimit")
else
    # Run specified tests
    for arg in "$@"; do
        case "$arg" in
            jwt|apikey|rbac|ratelimit)
                TESTS_TO_RUN+=("$arg")
                ;;
            *)
                echo -e "${YELLOW}Warning: Unknown test '$arg' - skipping${NC}"
                ;;
        esac
    done
fi

if [ ${#TESTS_TO_RUN[@]} -eq 0 ]; then
    echo -e "${RED}No valid tests specified${NC}"
    print_usage
    exit 1
fi

echo "Tests to run: ${TESTS_TO_RUN[*]}"
echo ""

# Run selected tests
for test in "${TESTS_TO_RUN[@]}"; do
    case "$test" in
        jwt)
            run_test "JWT Authentication" "auth-jwt.sh" "JWT Authentication Tests"
            ;;
        apikey)
            run_test "API Key Authentication" "auth-apikey.sh" "API Key Authentication Tests"
            ;;
        rbac)
            run_test "RBAC" "auth-rbac.sh" "Role-Based Access Control Tests"
            ;;
        ratelimit)
            run_test "Rate Limiting" "auth-ratelimit.sh" "Rate Limiting Tests"
            ;;
    esac
done

# Print summary and exit with appropriate code
echo ""
if print_summary; then
    exit 0
else
    exit 1
fi

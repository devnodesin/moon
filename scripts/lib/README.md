# API Testing Framework Library

This library provides a modular, type-safe framework for automated API testing with intelligent authentication and token management.

## Modules

### `types.py`
Type definitions and data classes:
- `TestDefinition` - Single API test case
- `TestSuite` - Complete test suite from JSON
- `TestResponse` - Response from test execution
- `AuthState` - Authentication state tracking
- `PlaceholderContext` - Context for placeholder replacements
- `TestResult` - Aggregated test results

### `http_client.py`
HTTP request execution:
- `execute_request()` - Execute single HTTP request
- `check_health()` - Check API server health
- `_build_curl_command()` - Build formatted curl commands

### `auth.py`
Authentication and token management:
- `perform_login()` - Login and get tokens
- `detect_password_change()` - Detect password change operations
- `detect_token_refresh()` - Detect token refresh operations
- `detect_login()` - Detect login operations
- `extract_tokens_from_response()` - Extract tokens from response
- `should_relogin_after_test()` - Determine if re-login needed
- `relogin_with_new_password()` - Re-login after password change

### `placeholders.py`
Placeholder replacement utilities:
- `replace_auth_placeholders()` - Replace $ACCESS_TOKEN and $REFRESH_TOKEN
- `replace_record_placeholders()` - Replace $ULID and $NEXT_CURSOR
- `extract_record_id_from_response()` - Extract record IDs from responses
- `fetch_record_id_from_collection()` - Fetch record ID from collection
- `extract_collection_name()` - Extract collection name from endpoint

### `formatters.py`
Output formatting:
- `format_markdown_result()` - Format test results as Markdown
- `sanitize_curl_for_documentation()` - Sanitize curl commands for docs
- `write_markdown_output()` - Write markdown to file

### `test_runner.py`
Test execution orchestration:
- `run_test_suite()` - Execute complete test suite with smart token management

## Key Features

### Smart Token Management
The framework automatically:
- Detects login operations and captures tokens
- Detects token refresh operations and updates tokens
- Replaces $ACCESS_TOKEN and $REFRESH_TOKEN placeholders
- Tracks all tokens for documentation sanitization

### Password Change Handling
When a test changes a password:
```json
{
  "endpoint": "/auth:me",
  "data": {
    "old_password": "OldPass123#",
    "password": "NewPass456#"
  }
}
```
The framework automatically:
1. Detects the password change operation
2. Captures the new password
3. Re-logins with the new password after the test completes
4. Updates all subsequent tests with the new token

### Token Refresh Handling
When a test refreshes tokens:
```json
{
  "endpoint": "/auth:refresh",
  "data": {
    "refresh_token": "$REFRESH_TOKEN"
  }
}
```
The framework automatically:
1. Detects the refresh operation
2. Extracts new tokens from the response
3. Updates tokens for subsequent tests

### Record ID Placeholders
The framework supports dynamic record ID placeholders:
- `$ULID` - Generic record ID placeholder
- `$NEXT_CURSOR` - Pagination cursor placeholder

Record IDs are automatically captured from `:create` and `:list` endpoints.

## Type Safety

All functions include comprehensive type hints for:
- Better IDE support and autocomplete
- Static type checking with mypy
- Clearer function signatures
- Reduced runtime errors

## Usage Example

```python
from lib import TestSuite, run_test_suite, perform_login

# Load test suite
test_suite = load_test_suite("tests/auth.json")

# Perform initial login if needed
auth_state = perform_login(
    test_suite.serverURL,
    test_suite.username,
    test_suite.password
)

# Run all tests
result = run_test_suite(test_suite, auth_state, "out/auth.md")

print(f"Status: {result.status}")
print(f"All tests passed: {result.all_tests_passed}")
```

## Error Handling

The framework includes comprehensive error handling:
- HTTP request errors are caught and reported
- JSON parsing errors are handled gracefully
- Server health checks prevent running tests against unhealthy servers
- Login failures are reported clearly
- All exceptions are caught and logged

## Documentation Generation

The framework automatically sanitizes output for documentation:
- Replaces actual server URLs with documentation URLs
- Replaces all captured tokens with placeholders
- Replaces record IDs with their placeholder types
- Formats curl commands with proper indentation
- Generates clean Markdown output

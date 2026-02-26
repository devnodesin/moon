# api-check.py

A modular, type-safe automated API test runner for Moon API endpoints.

> **Note**: Test files follow the API patterns documented in [SPEC_API.md](../SPEC_API.md). For endpoint details, request/response formats, and error codes, refer to that specification.

## Features

- **Modular Architecture**: Split into focused modules under `lib/` for easy maintenance
- **Type Safety**: Full type hints throughout for better IDE support and fewer errors
- **Smart Authentication**: Automatic token management and renewal
- **Password Change Detection**: Automatically re-logins after password changes
- **Token Refresh Handling**: Updates tokens immediately after refresh operations
- **Placeholder Replacement**: Dynamic $ACCESS_TOKEN, $REFRESH_TOKEN, $ULID, $NEXT_CURSOR
- **Health Checks**: Validates server health before running tests
- **Documentation Generation**: Clean Markdown output with sanitized commands

## Usage

```sh
python api-check.py [-i TESTFILE.json] [-o OUTDIR] [-t TESTDIR]
```

- `-i`: Input test JSON file (default: all in `tests/`)
- `-o`: Output directory for Markdown (default: `./out`)
- `-t`: Directory with test JSONs (default: `./tests`)

PREREQUEST: make sure you run this tool against a fresh installation of Moon server.

```sh
cd 
python api-check.py -i ./tests/020-auth.json

# or run all 
python api-check.py

# or against a different server than the one specified in the test files
python api-check.py --server=http://localhost:6000
```

Results are saved as Markdown in the output directory. this file can be copied to `cmd\moon\internal\handlers\templates\md\` to update the documentations

```sh
cp .\out\*.md ..\cmd\moon\internal\handlers\templates\md\
```

## Requirements

- Python 3.x
- `requests` library

## Test Files

Test files are JSON files describing a sequence of API requests and expected behaviors.

### Structure

- `serverURL`: Base URL of the API server (e.g., <http://localhost:8080>)
- `docURL`: URL used for documentation output (optional)
- `prefix`: (optional) API prefix (e.g., /api)
- `username`/`password`: (optional) for login/auth tests
- `tests`: Array of test objects, each with:
  - `name`: Short description
  - `cmd`: HTTP method (GET, POST, etc.)
  - `endpoint`: API endpoint (e.g., /users:list)
  - `headers`: (optional) Dict of headers
  - `data`: (optional) Request body (dict or string)
  - `details`/`notes`: (optional) Markdown for docs

### Example

```json
{
  "serverURL": "http://localhost:8080",
  "docURL": "https://api.example.com",
  "prefix": "/api",
  "username": "admin",
  "password": "moonadmin12#",
  "tests": [
    {
      "name": "List users",
      "cmd": "GET",
      "endpoint": "/users:list"
    },
    {
      "name": "Create user",
      "cmd": "POST",
      "endpoint": "/users:create",
      "headers": {"Content-Type": "application/json"},
      "data": {"username": "bob", "password": "secret"}
    }
  ]
}
```

You can add placeholders like `$ACCESS_TOKEN`, `$ULID`, and `$NEXT_CURSOR` in endpoints, headers, or data. These will be replaced automatically during test execution.

## Advanced Features

### Smart Token Management

The framework automatically handles authentication complexities:

#### Password Change Detection
When a test changes a password:
```json
{
  "name": "Change Password",
  "endpoint": "/auth:me",
  "data": {
    "old_password": "OldPass123#",
    "password": "NewPass456#"
  }
}
```
The framework:
1. Detects the password change (by `old_password` + `password` fields)
2. Executes the test
3. Automatically re-logins with the new password
4. Updates tokens for all subsequent tests

#### Token Refresh Handling
When a test refreshes tokens:
```json
{
  "name": "Refresh Token",
  "endpoint": "/auth:refresh",
  "data": {
    "refresh_token": "$REFRESH_TOKEN"
  }
}
```
The framework:
1. Detects the token refresh operation
2. Extracts new tokens from the response
3. Immediately updates tokens for subsequent tests

This ensures tests never fail due to stale tokens after password changes or token refreshes.

## Architecture

The codebase is organized into focused modules:

- **`lib/types.py`** - Type definitions and data classes
- **`lib/http_client.py`** - HTTP request execution
- **`lib/auth.py`** - Authentication and token management
- **`lib/placeholders.py`** - Placeholder replacement logic
- **`lib/formatters.py`** - Output formatting utilities
- **`lib/test_runner.py`** - Test execution orchestration

See [`lib/README.md`](lib/README.md) for detailed module documentation.

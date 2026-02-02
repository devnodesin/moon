## Overview

- **Problem**: The system lacks standardized implementation for cross-cutting API concerns including rate limit headers, CORS configuration, and sensitive data logging/redaction, leading to inconsistent client experience, security vulnerabilities, and potential exposure of sensitive information in logs.
- **Context**: Moon currently implements basic error responses and HTTP status codes across various endpoints (documented in PRD-046, PRD-047, PRD-048), but lacks: (1) rate limit header exposure for client-side throttling, (2) CORS configuration for cross-origin browser access, (3) comprehensive sensitive data redaction in logs and error responses, and (4) standardized error code catalog for programmatic error handling.
- **Solution**: Implement comprehensive API standards covering rate limit headers (X-RateLimit-*), CORS configuration via YAML, sensitive data redaction (password, token, secret, api_key fields), standardized error codes, and consistent HTTP status code usage across all endpoints.

## Requirements

### Functional Requirements

#### FR-1: Rate Limit Headers

**Requirement**: All API responses MUST include rate limit headers when rate limiting is enabled.

**Required Headers**:
- `X-RateLimit-Limit`: Maximum requests allowed per window (e.g., `100`)
- `X-RateLimit-Remaining`: Remaining requests in current window (e.g., `87`)
- `X-RateLimit-Reset`: Unix timestamp when the rate limit window resets (e.g., `1704067200`)
- `Retry-After`: Seconds until retry allowed (only on HTTP 429 responses, e.g., `60`)

**Header Format**:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1704067200
```

**HTTP 429 Response** (when rate limit exceeded):
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1704067200
Retry-After: 60
Content-Type: application/json

{
  "error": "rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED"
}
```

**Rate Limit Window**:
- Default window: 60 seconds (sliding window)
- Configurable via `limits.rate_limit_window` in YAML
- Per-IP address for anonymous requests
- Per-user for authenticated requests (JWT or API key)

**Enforcement**:
- Rate limit headers MUST be present on ALL responses (200, 400, 429, etc.) when rate limiting is enabled
- If rate limiting is disabled (`limits.rate_limit_enabled: false`), headers MUST NOT be included
- `X-RateLimit-Remaining` MUST accurately reflect remaining quota AFTER current request is processed

**Rationale**:
- Allows clients to implement intelligent backoff strategies
- Prevents surprise 429 errors by exposing remaining quota
- Industry-standard headers (RFC 6585, draft-ietf-httpapi-ratelimit-headers)

#### FR-2: CORS Configuration

**Requirement**: Support Cross-Origin Resource Sharing (CORS) for browser-based API clients with configurable policies.

**Configuration Structure**:
```yaml
cors:
  enabled: true                    # Default: false
  allowed_origins:                 # List of allowed origins
    - "https://app.example.com"
    - "https://dashboard.example.com"
  allowed_methods:                 # Default: GET, POST, PUT, DELETE, OPTIONS
    - GET
    - POST
    - PUT
    - DELETE
    - PATCH
    - OPTIONS
  allowed_headers:                 # Default: Content-Type, Authorization, X-API-KEY
    - Content-Type
    - Authorization
    - X-API-KEY
    - X-Request-ID
  exposed_headers:                 # Headers exposed to browser
    - X-RateLimit-Limit
    - X-RateLimit-Remaining
    - X-RateLimit-Reset
  allow_credentials: true          # Default: false (cookies, auth headers)
  max_age: 3600                    # Preflight cache duration (seconds)
```

**CORS Headers (on actual requests)**:
```
Access-Control-Allow-Origin: https://app.example.com
Access-Control-Allow-Credentials: true
Access-Control-Expose-Headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
```

**CORS Preflight (OPTIONS request)**:
```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: https://app.example.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-API-KEY, X-Request-ID
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 3600
```

**Wildcard Origins**:
- Support `*` in `allowed_origins` for development
- **WARNING**: Using `*` with `allow_credentials: true` is NOT allowed by CORS spec (must specify exact origins)
- If `allowed_origins: ["*"]` and `allow_credentials: true`, server MUST return error on startup

**Origin Validation**:
- Request `Origin` header MUST match one of `allowed_origins` (exact match)
- If no match, CORS headers MUST NOT be included (request fails in browser)
- Empty `allowed_origins` list means CORS is effectively disabled

**Preflight Request Handling**:
- All `OPTIONS` requests MUST return HTTP 204 (No Content)
- No authentication required for `OPTIONS` requests
- Preflight cache duration controlled by `max_age` (default: 3600 seconds / 1 hour)

**Rationale**:
- Enables browser-based Single Page Applications (SPAs) to access Moon API
- Configurable policies prevent unauthorized cross-origin access
- Preflight caching reduces overhead for complex requests

#### FR-3: Sensitive Data Redaction

**Requirement**: Automatically redact sensitive fields in logs, error responses, and debug output to prevent credential leakage.

**Sensitive Field Detection**:
Detect fields by name (case-insensitive):
- `password`
- `token`
- `secret`
- `api_key`
- `apikey`
- `authorization`
- `jwt`
- `refresh_token`
- `access_token`
- `client_secret`
- `private_key`
- `credential`
- `auth`

**Redaction Placeholder**: `***REDACTED***`

**Redaction Scope**:

**1. Log Output**:
```go
// Before redaction (UNSAFE):
log.Info().Interface("user", user).Msg("User created")
// Output: {"user": {"name": "Alice", "password": "secret123"}}

// After redaction (SAFE):
log.Info().Interface("user", RedactSensitive(user)).Msg("User created")
// Output: {"user": {"name": "Alice", "password": "***REDACTED***"}}
```

**2. Error Responses**:
```json
// Before redaction (UNSAFE):
{
  "error": "failed to create user",
  "details": {
    "user_input": {"name": "Alice", "password": "secret123"}
  }
}

// After redaction (SAFE):
{
  "error": "failed to create user",
  "details": {
    "user_input": {"name": "Alice", "password": "***REDACTED***"}
  }
}
```

**3. Debug Output**:
- Stack traces MUST NOT include request bodies with sensitive fields
- Database query logs MUST redact parameter values for sensitive columns

**Redaction Algorithm**:
1. Recursively traverse map/struct
2. For each key, check if name matches sensitive field list (case-insensitive)
3. If match, replace value with `***REDACTED***`
4. Continue traversing nested objects/arrays

**Exceptions**:
- Hashed passwords (e.g., bcrypt hashes) MAY be logged (they are already one-way hashes)
- ULIDs, UUIDs, and public IDs are NOT considered sensitive (they are public identifiers)

**Configuration**:
```yaml
logging:
  path: "/var/log/moon"
  redact_sensitive: true           # Default: true
  additional_sensitive_fields:     # User-defined sensitive fields
    - "ssn"
    - "credit_card"
    - "phone_number"
```

**Rationale**:
- Prevents credential leakage in logs (a common security vulnerability)
- Complies with security best practices (OWASP, CWE-532)
- Reduces risk of insider threats and log aggregation exposure

#### FR-4: Standardized Error Response Format

**Requirement**: All error responses MUST follow a consistent JSON structure with error codes for programmatic handling.

**Error Response Structure**:
```json
{
  "error": "human-readable error message",
  "code": "ERROR_CODE"
}
```

**With Optional Details**:
```json
{
  "error": "validation failed",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "email",
    "expected": "valid email format",
    "received": "not-an-email"
  }
}
```

**Validation Errors (Multiple Fields)**:
```json
{
  "error": "validation failed",
  "code": "VALIDATION_ERROR",
  "errors": [
    {
      "field": "name",
      "message": "field 'name' is required",
      "code": "REQUIRED_FIELD"
    },
    {
      "field": "age",
      "message": "field 'age' must be an integer",
      "code": "INVALID_TYPE",
      "expected_type": "integer",
      "actual_value": "***REDACTED***"
    }
  ]
}
```

**Error Code Catalog**:

| Error Code | HTTP Status | Description | Example |
|------------|-------------|-------------|---------|
| `VALIDATION_ERROR` | 400 | Input validation failed | Invalid field type, missing required field |
| `INVALID_JSON` | 400 | Malformed JSON in request body | Syntax error, unexpected token |
| `INVALID_ULID` | 400 | Invalid ULID format | Wrong length, invalid characters |
| `INVALID_CURSOR` | 400 | Invalid pagination cursor | Malformed ULID in `?after` param |
| `PAGE_SIZE_EXCEEDED` | 400 | Page size exceeds maximum | `?limit=2000` when max is 1000 |
| `COLLECTION_NOT_FOUND` | 404 | Collection does not exist | GET `/nonexistent:list` |
| `RECORD_NOT_FOUND` | 404 | Record with given ID not found | GET `/users/01HQXYZ...` |
| `DUPLICATE_COLLECTION` | 409 | Collection name already exists | POST `/collections:create` with existing name |
| `UNIQUE_CONSTRAINT_VIOLATION` | 409 | Unique column constraint violated | Insert duplicate value in unique column |
| `MAX_COLLECTIONS_REACHED` | 409 | Maximum collections limit reached | 1000+ collections already exist |
| `MAX_COLUMNS_REACHED` | 409 | Maximum columns limit reached | 100+ columns in collection |
| `UNAUTHORIZED` | 401 | Authentication required | Missing or invalid JWT/API key |
| `FORBIDDEN` | 403 | Insufficient permissions | Non-admin user accessing admin endpoint |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests | Exceeded rate limit quota |
| `QUERY_TIMEOUT` | 504 | Query execution timeout | Query took longer than 30 seconds |
| `INTERNAL_ERROR` | 500 | Unexpected server error | Database connection failed, panic |

**Error Code Consistency**:
- Same error condition MUST always return the same `code`
- Error messages MAY vary (localization, context), but `code` MUST NOT
- Clients should handle errors by `code`, not by parsing `error` message

**Details Field** (optional):
- Include additional context when helpful for debugging
- MUST NOT include sensitive data (apply redaction)
- Examples: field name, expected value range, constraint violated

**Rationale**:
- Enables programmatic error handling in client applications
- Consistent structure simplifies API client development
- Error codes allow for client-side error message customization (i18n)

#### FR-5: HTTP Status Code Standards

**Requirement**: Enforce consistent HTTP status code usage across all endpoints.

**Status Code Rules**:

**2xx Success**:
- `200 OK`: Successful GET, PUT, DELETE (returns data)
- `201 Created`: Successful POST that creates a resource (collection, record, user)
- `204 No Content`: Successful DELETE (no response body), OPTIONS preflight

**4xx Client Errors**:
- `400 Bad Request`: Validation errors, malformed JSON, invalid query parameters
- `401 Unauthorized`: Missing authentication credentials (no JWT, no API key)
- `403 Forbidden`: Valid credentials but insufficient permissions (non-admin accessing admin endpoint)
- `404 Not Found`: Resource (collection, record, endpoint) not found
- `409 Conflict`: Resource conflict (duplicate collection, unique constraint violation, limit exceeded)
- `429 Too Many Requests`: Rate limit exceeded

**5xx Server Errors**:
- `500 Internal Server Error`: Database errors, unexpected panics, unhandled errors
- `504 Gateway Timeout`: Query timeout, slow query exceeds limit

**Status Code Selection Logic**:

```
Request received
├─ Parsing failed? → 400 Bad Request (malformed JSON, invalid syntax)
├─ Authentication missing? → 401 Unauthorized
├─ Authentication invalid? → 401 Unauthorized
├─ Permission denied? → 403 Forbidden
├─ Resource not found? → 404 Not Found
├─ Validation failed? → 400 Bad Request
├─ Resource conflict? → 409 Conflict
├─ Rate limit exceeded? → 429 Too Many Requests
├─ Query timeout? → 504 Gateway Timeout
├─ Unexpected error? → 500 Internal Server Error
└─ Success → 200 OK / 201 Created / 204 No Content
```

**Authentication vs. Authorization**:
- `401 Unauthorized`: "Who are you?" - No credentials or invalid credentials
- `403 Forbidden`: "I know who you are, but you can't do that" - Insufficient role/permissions

**Examples**:

```
# Missing JWT
GET /users:list
→ 401 Unauthorized {"error": "authentication required", "code": "UNAUTHORIZED"}

# Valid JWT but non-admin user
GET /users:list (requires admin role)
→ 403 Forbidden {"error": "admin role required", "code": "FORBIDDEN"}

# Collection not found
GET /nonexistent:list
→ 404 Not Found {"error": "collection 'nonexistent' not found", "code": "COLLECTION_NOT_FOUND"}

# Duplicate collection
POST /collections:create {"name": "users"}
→ 409 Conflict {"error": "collection 'users' already exists", "code": "DUPLICATE_COLLECTION"}
```

**Rationale**:
- Consistent status codes improve API predictability
- Clients can handle errors generically by status code family (4xx, 5xx)
- Proper authentication vs. authorization distinction prevents information leakage

### Technical Requirements

#### TR-1: Rate Limit Header Middleware

File: `cmd/moon/internal/middleware/ratelimit.go`

```go
type RateLimitMiddleware struct {
    enabled   bool
    limiter   *RateLimiter
    limit     int
    window    time.Duration
}

func (m *RateLimitMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !m.enabled {
            next.ServeHTTP(w, r)
            return
        }
        
        // Identify client (IP or user ID)
        clientID := getClientIdentifier(r)
        
        // Check rate limit
        allowed, remaining, resetTime := m.limiter.Allow(clientID)
        
        // Set rate limit headers
        w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.limit))
        w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
        w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
        
        if !allowed {
            retryAfter := resetTime - time.Now().Unix()
            w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
            writeError(w, http.StatusTooManyRequests, "rate limit exceeded", "RATE_LIMIT_EXCEEDED")
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

#### TR-2: CORS Middleware

File: `cmd/moon/internal/middleware/cors.go`

```go
type CORSConfig struct {
    Enabled          bool
    AllowedOrigins   []string
    AllowedMethods   []string
    AllowedHeaders   []string
    ExposedHeaders   []string
    AllowCredentials bool
    MaxAge           int
}

func CORSMiddleware(cfg *CORSConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !cfg.Enabled {
                next.ServeHTTP(w, r)
                return
            }
            
            origin := r.Header.Get("Origin")
            if origin == "" {
                next.ServeHTTP(w, r)
                return
            }
            
            // Validate origin
            if !isOriginAllowed(origin, cfg.AllowedOrigins) {
                next.ServeHTTP(w, r) // No CORS headers
                return
            }
            
            // Set CORS headers
            w.Header().Set("Access-Control-Allow-Origin", origin)
            if cfg.AllowCredentials {
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }
            if len(cfg.ExposedHeaders) > 0 {
                w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
            }
            
            // Handle preflight
            if r.Method == http.MethodOptions {
                w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
                w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
                w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
                w.WriteHeader(http.StatusNoContent)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

func isOriginAllowed(origin string, allowed []string) bool {
    for _, a := range allowed {
        if a == "*" || a == origin {
            return true
        }
    }
    return false
}
```

#### TR-3: Sensitive Data Redaction

File: `cmd/moon/internal/logging/redact.go`

```go
var SensitiveFields = []string{
    "password", "token", "secret", "api_key", "apikey",
    "authorization", "jwt", "refresh_token", "access_token",
    "client_secret", "private_key", "credential", "auth",
}

const RedactedPlaceholder = "***REDACTED***"

func RedactSensitive(data interface{}) interface{} {
    return redactValue(reflect.ValueOf(data))
}

func redactValue(v reflect.Value) interface{} {
    switch v.Kind() {
    case reflect.Map:
        result := make(map[string]interface{})
        for _, key := range v.MapKeys() {
            keyStr := key.String()
            value := v.MapIndex(key)
            
            if isSensitiveField(keyStr) {
                result[keyStr] = RedactedPlaceholder
            } else {
                result[keyStr] = redactValue(value)
            }
        }
        return result
        
    case reflect.Struct:
        // Similar logic for structs
        
    case reflect.Slice, reflect.Array:
        // Recursively redact array elements
        
    default:
        return v.Interface()
    }
}

func isSensitiveField(field string) bool {
    lower := strings.ToLower(field)
    for _, sensitive := range SensitiveFields {
        if strings.Contains(lower, sensitive) {
            return true
        }
    }
    return false
}
```

#### TR-4: Error Response Helper

File: `cmd/moon/internal/handlers/error.go`

```go
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Code    string                 `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}

type ValidationErrorResponse struct {
    Error  string            `json:"error"`
    Code   string            `json:"code"`
    Errors []ValidationError `json:"errors"`
}

type ValidationError struct {
    Field        string      `json:"field"`
    Message      string      `json:"message"`
    Code         string      `json:"code"`
    ExpectedType string      `json:"expected_type,omitempty"`
    ActualValue  interface{} `json:"actual_value,omitempty"`
}

func writeError(w http.ResponseWriter, status int, message string, code string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    resp := ErrorResponse{
        Error: message,
        Code:  code,
    }
    
    json.NewEncoder(w).Encode(RedactSensitive(resp))
}

func writeValidationError(w http.ResponseWriter, errors []ValidationError) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    
    resp := ValidationErrorResponse{
        Error:  "validation failed",
        Code:   "VALIDATION_ERROR",
        Errors: errors,
    }
    
    json.NewEncoder(w).Encode(RedactSensitive(resp))
}
```

#### TR-5: Configuration Structure

File: `cmd/moon/internal/config/config.go`

```yaml
# Rate Limiting
limits:
  rate_limit_enabled: true         # Default: false
  rate_limit_window: 60            # Seconds, default: 60
  rate_limit_max_requests: 100     # Per window, default: 100

# CORS
cors:
  enabled: false                   # Default: false
  allowed_origins:
    - "https://app.example.com"
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - PATCH
    - OPTIONS
  allowed_headers:
    - Content-Type
    - Authorization
    - X-API-KEY
  exposed_headers:
    - X-RateLimit-Limit
    - X-RateLimit-Remaining
    - X-RateLimit-Reset
  allow_credentials: true
  max_age: 3600

# Logging
logging:
  path: "/var/log/moon"
  redact_sensitive: true           # Default: true
  additional_sensitive_fields:
    - "ssn"
    - "credit_card"
```

### Error Handling and Failure Modes

**EH-1: Rate Limit Headers on Error Responses**
- Rate limit headers MUST be present even on 4xx/5xx responses
- If rate limiter itself fails, proceed without rate limiting (fail open)
- Log rate limiter failures at ERROR level

**EH-2: CORS Misconfiguration**
- If `allowed_origins: ["*"]` and `allow_credentials: true`, server MUST fail to start with error
- If `allowed_origins` is empty and `enabled: true`, server MUST fail to start
- Invalid origin patterns (e.g., malformed URLs) MUST be rejected at startup

**EH-3: Redaction Failure**
- If redaction function panics, log original error and return generic error to client
- Redaction MUST NOT break JSON serialization
- If field value is complex object, redact entire object (not just string fields)

**EH-4: Missing Error Codes**
- If error code is not provided, use `"INTERNAL_ERROR"` as fallback
- All validation errors MUST include field-level error codes
- Database errors MUST NOT expose raw SQL in `error` field (log only)

### Validation Rules and Constraints

**Rate Limit Configuration**:
- `rate_limit_window` MUST be > 0 (minimum: 1 second)
- `rate_limit_max_requests` MUST be > 0
- If `rate_limit_enabled: false`, rate limit headers MUST NOT be included

**CORS Configuration**:
- `allowed_origins` MUST NOT be empty if `enabled: true`
- `allowed_origins: ["*"]` with `allow_credentials: true` MUST be rejected
- `max_age` MUST be >= 0 (0 disables caching)

**Sensitive Fields**:
- Additional sensitive fields MUST be lowercase
- Field matching is case-insensitive and substring-based (e.g., "user_password" matches "password")

## Acceptance

### AC-1: Rate Limit Headers - Configuration

- [ ] Add `limits.rate_limit_enabled`, `limits.rate_limit_window`, `limits.rate_limit_max_requests` to config
- [ ] Validate `rate_limit_window > 0` and `rate_limit_max_requests > 0` at startup
- [ ] Default values: `enabled: false`, `window: 60`, `max_requests: 100`
- [ ] If `enabled: false`, rate limit headers MUST NOT be present

### AC-2: Rate Limit Headers - Middleware Implementation

- [ ] Implement `RateLimitMiddleware` in `cmd/moon/internal/middleware/ratelimit.go`
- [ ] Track requests per client (IP address for anonymous, user ID for authenticated)
- [ ] Use sliding window algorithm for rate limiting
- [ ] Set headers on ALL responses (200, 400, 429, 500, etc.)
- [ ] `X-RateLimit-Remaining` reflects quota AFTER current request

### AC-3: Rate Limit Headers - Response Verification

- [ ] Test: Make request, verify `X-RateLimit-Limit: 100` header present
- [ ] Test: Make 87 requests, verify `X-RateLimit-Remaining: 13`
- [ ] Test: Exceed limit, verify HTTP 429 with `Retry-After` header
- [ ] Test: After reset time, verify rate limit resets
- [ ] Test: Authenticated vs. anonymous users have separate quotas

### AC-4: Rate Limit Headers - 429 Response

- [ ] Test: Exceed rate limit, verify HTTP 429 response
- [ ] Response includes: `X-RateLimit-Limit`, `X-RateLimit-Remaining: 0`, `X-RateLimit-Reset`, `Retry-After`
- [ ] Response body: `{"error": "rate limit exceeded", "code": "RATE_LIMIT_EXCEEDED"}`
- [ ] Test: Wait for `Retry-After` seconds, verify request succeeds

### AC-5: CORS Configuration - Validation

- [ ] Add `cors` section to YAML config
- [ ] Validate `allowed_origins` not empty if `enabled: true`
- [ ] Validate `allowed_origins: ["*"]` with `allow_credentials: true` is rejected at startup
- [ ] Default: `enabled: false`, `allow_credentials: false`, `max_age: 3600`

### AC-6: CORS Middleware - Implementation

- [ ] Implement `CORSMiddleware` in `cmd/moon/internal/middleware/cors.go`
- [ ] Validate request `Origin` header against `allowed_origins`
- [ ] Set `Access-Control-Allow-Origin` to matching origin (not `*` if credentials enabled)
- [ ] Set `Access-Control-Allow-Credentials: true` if configured

### AC-7: CORS Middleware - Preflight Requests

- [ ] All `OPTIONS` requests return HTTP 204 (No Content)
- [ ] No authentication required for `OPTIONS` requests
- [ ] Response includes: `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, `Access-Control-Max-Age`
- [ ] Test: Send `OPTIONS` request with `Access-Control-Request-Method: POST`
- [ ] Verify response includes `Access-Control-Allow-Methods: GET, POST, ...`

### AC-8: CORS Middleware - Actual Requests

- [ ] Test: Send `GET` request with `Origin: https://app.example.com`
- [ ] Verify response includes `Access-Control-Allow-Origin: https://app.example.com`
- [ ] Verify response includes `Access-Control-Expose-Headers` with rate limit headers
- [ ] Test: Send request with disallowed origin, verify NO CORS headers

### AC-9: CORS Middleware - Credentials

- [ ] Test: `allow_credentials: true`, verify `Access-Control-Allow-Credentials: true` present
- [ ] Test: `allow_credentials: false`, verify header NOT present
- [ ] Test: Wildcard origin with credentials, verify server startup fails

### AC-10: Sensitive Data Redaction - Implementation

- [ ] Implement `RedactSensitive()` in `cmd/moon/internal/logging/redact.go`
- [ ] Support map, struct, slice, and nested object redaction
- [ ] Field matching is case-insensitive and substring-based
- [ ] Redaction placeholder: `***REDACTED***`

### AC-11: Sensitive Data Redaction - Field List

- [ ] Default sensitive fields: `password`, `token`, `secret`, `api_key`, `apikey`, `authorization`, `jwt`, `refresh_token`, `access_token`, `client_secret`, `private_key`, `credential`, `auth`
- [ ] Support `logging.additional_sensitive_fields` in config
- [ ] Test: Field name `"user_password"` is redacted (substring match)
- [ ] Test: Field name `"username"` is NOT redacted (no match)

### AC-12: Sensitive Data Redaction - Log Output

- [ ] Apply `RedactSensitive()` to all logged objects
- [ ] Test: Log object with `{"name": "Alice", "password": "secret"}` → `{"name": "Alice", "password": "***REDACTED***"}`
- [ ] Test: Log nested object with sensitive field
- [ ] Test: Log array of objects with sensitive fields

### AC-13: Sensitive Data Redaction - Error Responses

- [ ] Apply `RedactSensitive()` to all error response bodies before JSON encoding
- [ ] Test: Error includes user input with password field, verify redacted
- [ ] Test: Validation error with sensitive field in `actual_value`, verify redacted

### AC-14: Standardized Error Response - Format

- [ ] All error responses MUST include `error` and `code` fields
- [ ] Implement `writeError()` helper in `cmd/moon/internal/handlers/error.go`
- [ ] Optional `details` field for additional context
- [ ] Test: All endpoints return consistent error format

### AC-15: Standardized Error Response - Error Codes

- [ ] Define error code constants in `cmd/moon/internal/constants/errors.go`
- [ ] Implement all error codes from catalog (VALIDATION_ERROR, INVALID_JSON, etc.)
- [ ] Test: Each error condition returns correct error code
- [ ] Document error codes in `SPEC.md`

### AC-16: Standardized Error Response - Validation Errors

- [ ] Implement `ValidationErrorResponse` with `errors` array
- [ ] Each validation error includes: `field`, `message`, `code`
- [ ] Optional fields: `expected_type`, `actual_value` (redacted if sensitive)
- [ ] Test: Multi-field validation failure returns all errors in array

### AC-17: HTTP Status Code Standards - Enforcement

- [ ] Review all endpoints for correct status code usage
- [ ] Test: Missing authentication returns 401 (not 403)
- [ ] Test: Insufficient permissions returns 403 (not 401)
- [ ] Test: Resource not found returns 404
- [ ] Test: Duplicate resource returns 409 (not 400)
- [ ] Test: Query timeout returns 504 (not 500)

### AC-18: HTTP Status Code Standards - Documentation

- [ ] Update `SPEC.md` with "Error Handling" section
- [ ] Document all HTTP status codes with examples
- [ ] Document status code selection logic (flowchart)
- [ ] Document 401 vs. 403 distinction

### AC-19: Integration Testing - Rate Limiting

- [ ] Test: Enable rate limiting, make 100 requests, verify headers
- [ ] Test: Exceed limit, verify 429 response with headers
- [ ] Test: Wait for reset, verify limit resets
- [ ] Test: Disable rate limiting, verify no headers

### AC-20: Integration Testing - CORS

- [ ] Test: Enable CORS, send preflight request, verify 204 response
- [ ] Test: Send actual request with allowed origin, verify CORS headers
- [ ] Test: Send request with disallowed origin, verify no CORS headers
- [ ] Test: Test with wildcard origin (`*`)

### AC-21: Integration Testing - Sensitive Data

- [ ] Test: Log object with sensitive fields, verify redacted in log file
- [ ] Test: Error response with sensitive data, verify redacted in JSON
- [ ] Test: Add custom sensitive field to config, verify redacted

### AC-22: Integration Testing - Error Responses

- [ ] Test: Every error scenario returns correct format and code
- [ ] Test: Validation error with multiple fields
- [ ] Test: Database error returns 500 with generic message (not raw SQL)
- [ ] Test: Rate limit exceeded returns correct error code

### AC-23: Performance Testing

- [ ] Measure rate limit middleware overhead: < 1ms per request
- [ ] Measure CORS middleware overhead: < 1ms per request
- [ ] Measure redaction overhead: < 5ms for typical objects
- [ ] Verify rate limiter can handle 1000+ req/s

### AC-24: Configuration Validation

- [ ] Test: Invalid rate limit config (window: 0) fails startup
- [ ] Test: Invalid CORS config (wildcard + credentials) fails startup
- [ ] Test: Empty `allowed_origins` with `enabled: true` fails startup

### AC-25: Documentation Updates

- [ ] Update `SPEC.md` with "Rate Limiting" section
- [ ] Update `SPEC.md` with "CORS Configuration" section
- [ ] Update `SPEC.md` with "Error Handling" section
- [ ] Update `SPEC.md` with "Sensitive Data Logging" section
- [ ] Update API doc template with error response examples
- [ ] Update `INSTALL.md` with CORS configuration examples

### AC-26: Testing Checklist

- [ ] All unit tests pass for redaction, CORS, rate limiting
- [ ] All integration tests pass for error responses
- [ ] All error scenarios produce correct format and code
- [ ] Rate limit headers verified on all response types
- [ ] CORS preflight and actual requests tested
- [ ] Test coverage >= 90% for new code

---

### Implementation Checklist

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `USAGE.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

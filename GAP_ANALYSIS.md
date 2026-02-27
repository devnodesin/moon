# Gap Analysis: Specification vs. Source Code

**Repository:** `/home/runner/work/moon/moon`  
**Date:** 2025-07-27  
**Scope:** All spec files (`SPEC.md`, `SPEC_API.md`, `SPEC_AUTH.md`, `SPEC_API/*.md`) vs. source code in `cmd/`

---

## Summary

The codebase is substantially complete and implements the majority of features described in the specification. Most API endpoints, authentication flows, and data handling work as specified. The gaps identified below range from spec inconsistencies (where the spec contradicts itself) to features that are partially or fully unimplemented.

---

## 1. Specified but Not Implemented (or Partially Implemented)

### 1.1 Health Endpoint Missing `status` Field

**Spec (SPEC.md):**
> "Returns a JSON response with four fields: `moon` (Service version string), `status` (Service health status `ok` or `down`), `timestamp` (RFC3339 timestamp of the health check)"

**Code (`cmd/moon/internal/server/server.go`):**
```go
response := map[string]any{
    "moon":      s.version,
    "timestamp": time.Now().UTC().Format(time.RFC3339),
}
```
The `status` field is not returned. The health response contains only two fields: `moon` and `timestamp`.

**Note:** `SPEC_API/010-health.md` also shows only `moon` and `timestamp` in the response example, which creates an internal spec inconsistency (see Section 3).

---

### 1.2 Pagination `prev` Cursor Always Null

**Spec (SPEC_API.md):**
> "`prev` (string | null): Cursor pointing to the record before the current page. Pass to `?after` to return to the previous page. `null` on the first page."

**Code (all list handlers):**
All `:list` endpoints — data records, users, API keys, and collections — always return `"prev": null` regardless of the current page. Backward pagination is not implemented. The spec's description of using `?after={meta.prev}` to navigate backwards is non-functional.

---

### 1.3 Rate Limit Response Code (Ambiguous/Inconsistent in Spec — Code Follows SPEC_API.md)

**SPEC_AUTH.md:**
> "Response: `429 Too Many Requests` when exceeded"

**SPEC_API.md:**
> `HTTP/1.1 400 Too Many Requests`

**Code:** Returns `http.StatusBadRequest` (400).

The code follows `SPEC_API.md` by returning HTTP 400 for rate limit violations. However, `SPEC_AUTH.md` specifies HTTP 429 which is the correct HTTP standard for rate limiting. The two spec files contradict each other. The code's use of 400 for rate limiting is semantically incorrect per HTTP standards.

---

### 1.4 `readonly` Role Not Specified

**Spec:** Only two roles are defined — `admin` and `user`.

**Code (`cmd/moon/internal/auth/models.go`):**
```go
RoleReadOnly UserRole = "readonly"
```
A `readonly` role is defined in the code and included in `ValidRoles()`. This role is not specified in `SPEC_AUTH.md`, `SPEC_API.md`, or `SPEC.md`. If accepted during user creation, it could create an undocumented and untested permission state.

---

### 1.5 `errors` Package Not Used by Handlers

**Spec (SPEC_API.md):** Defines a simple error format `{"message": "..."}` with specific HTTP status codes.

**Code:** There is a sophisticated `cmd/moon/internal/errors/errors.go` package defining `ErrorCode`, `APIError`, `ErrorHandler`, `normalizeStatusCode`, etc. However, the actual HTTP handlers do not use this package — they use a local `writeError(w, code, msg)` helper. The `errors` package appears to be unused infrastructure, creating dead code.

---

## 2. Implemented but Not Specified (Code Exceeds Spec)

### 2.1 Third-Party Dependencies Beyond Standard Library

**Spec (AGENTS.md):**
> "Use only the Go standard library unless a third-party dependency is absolutely essential."

**Code (`go.mod`)** includes non-essential dependencies:
- `github.com/rs/zerolog` — structured logging; standard `log` package would suffice
- `github.com/spf13/viper` — configuration; standard library YAML parsing could be used
- `github.com/google/uuid` — used only for request ID generation; a simpler approach using `crypto/rand` is possible
- `github.com/yuin/goldmark` — Markdown-to-HTML conversion for `/doc/` endpoint; arguable whether this is "absolutely essential"

Clearly essential (justified) dependencies: `golang-jwt/jwt`, `lib/pq`, `go-sql-driver/mysql`, `modernc.org/sqlite`, `oklog/ulid`, `golang.org/x/crypto`.

---

### 2.2 `X-Request-ID` Header Not Documented in API Spec

**Spec:** The exposed CORS headers listed are `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `X-Request-ID`.

**Code:** The `X-Request-ID` header is exposed and referenced in the `errors` package and CORS middleware. This header is not described in `SPEC_API.md` or `SPEC_AUTH.md` beyond being listed as an exposed CORS header. Its generation and lifecycle are not specified.

---

### 2.3 Root Message Endpoint (`GET /`)

**Spec:** No root endpoint is defined.

**Code:** When no URL prefix is configured, a `GET /{$}` route returns a plain-text "root message" (`config.RootMessage`). This is an undocumented endpoint.

---

### 2.4 `admin` Cannot Modify Own Account via User Management Endpoints

**Spec:** Not explicitly stated.

**Code:** The `UsersHandler.Update` and `UsersHandler.Destroy` methods explicitly reject requests where the admin's own ID matches the target user ID:
```go
if claims.UserID == userID {
    writeError(w, http.StatusBadRequest, "cannot modify own account via user management endpoints")
    return
}
```
This is a reasonable security constraint but is not specified in `SPEC_AUTH.md` which only states "Admin cannot modify own role."

---

### 2.5 Login Rate Limiter Per IP and Username

**Spec (SPEC_AUTH.md):** "Login attempts: 5 attempts per 15 minutes per IP/username"

**Code:** Implemented with `LoginRateLimiter` using composite key of `clientIP + username`. This is consistent with the spec but the implementation detail (composite key) is not specified.

---

### 2.6 Token Blacklist for Immediate Logout Revocation

**Spec (SPEC_AUTH.md):** "Token Blacklist: In-database blacklist for revoked tokens (logout, password changes)"

**Code:** Fully implemented via `auth.TokenBlacklist`. The logout handler also blacklists the access token — the spec mentions the blacklist exists but does not detail that logout should immediately revoke the access token (only the refresh token). The code goes beyond the spec.

---

### 2.7 Description Length Limit on API Keys

**Spec:** API key name must be "3-100 chars". No description length limit is specified.

**Code:** Enforces `MaxDescriptionLength = 500` characters on the description field. This limit is not in the spec.

---

### 2.8 Middleware Order Differs from Spec

**Spec (SPEC_AUTH.md):**
> "1. CORS → 2. Rate Limiting → 3. Authentication → 4. Authorization → 5. Validation → 6. Logging → 7. Handler → 8. Error Handling"

**Code (`server.go`):**
The actual order is: CORS → Logging → Authentication → Rate Limiting → Authorization → Handler

Rate limiting and logging are in different positions than specified. Logging occurs before authentication (spec says after authorization), and rate limiting occurs after authentication (spec says before authentication). This difference is functionally important: the spec's order would rate-limit unauthenticated requests, while the code only rate-limits authenticated requests.

---

## 3. Inconsistencies Between Spec Files

### 3.1 Health Response Field Count

**SPEC.md:** States the health response has "four fields" but lists only three: `moon`, `status`, `timestamp`.

**SPEC_API/010-health.md:** Shows only two fields: `moon` and `timestamp` (no `status` field, different from SPEC.md).

These two spec files directly contradict each other on the health response structure.

---

### 3.2 `users:create` Request Body Format

**SPEC_AUTH.md:** Shows a flat request body:
```json
{ "username": "newuser", "email": "...", "password": "...", "role": "user", "can_write": false }
```

**SPEC_API/030-users.md:** Shows a wrapped request body:
```json
{ "data": { "username": "moonuser", "email": "...", "password": "...", "role": "user" } }
```

The code implements the wrapped format (`{ "data": { ... } }`), consistent with `SPEC_API/030-users.md`. The `SPEC_AUTH.md` format is incorrect/stale.

---

### 3.3 `apikeys:update` Rotate Action Format

**SPEC_AUTH.md:** Shows rotate with flat body:
```json
{ "action": "rotate" }
```

**SPEC_API/040-apikeys.md:** Shows rotate with wrapped body:
```json
{ "data": { "action": "rotate" } }
```

The code uses the wrapped format, consistent with `SPEC_API/040-apikeys.md`. The `SPEC_AUTH.md` format is incorrect/stale.

---

### 3.4 Pagination Maximum Page Size

**SPEC_API.md:** States "maximum allowed is 100" in two places: the `:list` endpoint documentation and the query options table.

**SPEC.md:** States "Max page size: 200 (configurable via `pagination.max_page_size`)"

**Code:** `MaxPaginationLimit = 200`, consistent with `SPEC.md`.

The two spec files disagree. The code follows `SPEC.md`. Any client following `SPEC_API.md` would not expect page sizes above 100 to be accepted.

---

### 3.5 Forbidden (403) vs. Unauthorized (401) for Authorization Failures

**SPEC_API.md:** "403 (Forbidden) is intentionally omitted. Authorization or permission failures should be handled via 401."

**SPEC_AUTH.md:** "If valid credentials but insufficient permissions, return `403 Forbidden`"

**Code:** The authorization middleware internally uses `http.StatusForbidden` (403) but converts it to `http.StatusUnauthorized` (401) before writing the response. This matches `SPEC_API.md` at the HTTP level but the internal code structure follows `SPEC_AUTH.md`.

---

### 3.6 API Keys `updated_at` Field in Rotation Response

**SPEC_AUTH.md:** The key rotation response example includes `"created_at"` in the response.

**SPEC_API/040-apikeys.md:** The rotation response includes only `id`, `name`, and `key` fields (no `created_at`).

**Code:** Returns only `id`, `name`, `key` — consistent with `SPEC_API/040-apikeys.md`.

---

### 3.7 `users:list` Role Filter Not Documented in SPEC_API

**SPEC_AUTH.md:** Documents `?role` as an optional query parameter for `GET /users:list`.

**SPEC_API/030-users.md:** Does not mention the `?role` filter parameter.

**Code:** Implements `?role` filtering. The feature exists in the code and is partially specified, but the two spec files are not aligned.

---

### 3.8 `updated_at` in API Key Metadata Response

**SPEC_AUTH.md:** The `apikeys:update` (metadata update) response shows an `"updated_at"` field.

**SPEC_API/040-apikeys.md:** The update response does not include `"updated_at"`.

**Code:** Does not return `updated_at` in the update response. Consistent with `SPEC_API/040-apikeys.md`.

---

## 4. Other Observations

### 4.1 Logging Uses Mixed Approaches

The code mixes `log.Printf` (standard library) and `github.com/rs/zerolog` for logging. Some packages (handlers, auth) use `log.Printf`, while the `logging` package wraps zerolog. This creates inconsistent log output formats across the application and contradicts the goal of a unified logging system.

---

### 4.2 `errors` Package Defines HTTP Error Codes Not in Spec

The `errors` package defines internal error codes (`CodeValidationFailed`, `CodeNotFound`, `CodeRateLimitExceeded`, etc.) that go well beyond what the spec permits. The spec states: "No internal error codes or additional error metadata are used." These codes are not exposed in API responses, but the package infrastructure is confusing and could lead to accidental exposure.

---

### 4.3 `collections:list` Does Not Support Filtering or Sorting

The `SPEC_API.md` defines query options (`filter`, `sort`, `q`, `fields`) for `:list` endpoints. The `/collections:list` endpoint only supports pagination (`limit`, `after`) and does not support any of these query parameters. This is a reasonable product decision (collections are schema metadata), but it is not explicitly excluded in the spec.

---

### 4.4 Schema Response Uses `fields` Key (Not `columns`)

**SPEC_API/060-data.md (`:schema` endpoint):** Response uses `"fields"` key.

**SPEC_API/050-collection.md (`collections:get` and `collections:create`):** Response uses `"columns"` key.

These two different terms (`fields` vs. `columns`) refer to the same concept in the spec itself. The code correctly uses `fields` for `:schema` and `columns` for collection management endpoints, following the spec — but the dual terminology in the spec is a potential source of confusion.

---

### 4.5 `can_write` Default Differs Between Users and API Keys

**SPEC_AUTH.md:** "Write-enabled by default (`can_write: true`)" for the `user` role.

**Code (users):** `canWrite = true` by default when creating a user without `can_write` specified.

**Code (API keys):** `canWrite = false` by default when creating an API key without `can_write` specified.

**SPEC_AUTH.md:** The API key section does not explicitly state a default for `can_write`. The asymmetry between users (default `true`) and API keys (default `false`) in the code is not clearly documented in the spec.

---

### 4.6 Aggregation Returns Float for Integer Fields

**Spec:** For aggregation results on `integer` fields, the spec examples show integer values (e.g., `"value": 85`, `"value": 55`).

**Code:** Aggregation results are always scanned into `sql.NullFloat64` and returned as `float64`. This means integer field aggregations return floating-point numbers (e.g., `85.0` instead of `85`). For `decimal` fields this is appropriate; for `integer` fields the response type may differ from spec examples.

---

### 4.7 Collections `:list` Records Count May Be Slow at Scale

**Spec (SPEC_API/050-collection.md):** The `collections:list` response includes a `"records"` field showing the row count per collection.

**Code:** The `CollectionInfo` struct includes `Records int` and the handler counts records per collection using individual `COUNT(*)` queries for each collection. At scale (e.g., 1000 collections), this could be expensive. The spec does not mention any performance considerations for this field.

---

### 4.8 `doc` Endpoint Serves System Information

The `/doc/` endpoint serves auto-generated API documentation including the list of available collections and API endpoint details. The spec defines this endpoint, but access control is not specified — the endpoint is public (no authentication required) and returns collection names, which could be considered metadata leakage in sensitive deployments.

---

### 4.9 No Token Cleanup / Expiry Purge

**Spec (SPEC_AUTH.md):** "Cleanup: Expired tokens should be purged periodically."

**Code:** There is no background goroutine or scheduled task to purge expired refresh tokens or blacklisted access tokens from the database. Expired tokens are detected on access but never proactively deleted, leading to unbounded table growth over time.

---

## 5. Summary Table

| # | Category | Item | Severity |
|---|----------|------|----------|
| 1.1 | Missing | Health endpoint `status` field | Medium |
| 1.2 | Missing | Pagination `prev` cursor (always null) | Medium |
| 1.3 | Inconsistent | Rate limit uses 400 (spec says 429) | Low |
| 1.4 | Extra | `readonly` role not in spec | Medium |
| 1.5 | Extra | `errors` package unused (dead code) | Low |
| 2.1 | Extra | Third-party deps beyond std lib | Low |
| 2.7 | Extra | API key description length limit (undocumented) | Low |
| 2.8 | Inconsistent | Middleware order differs from spec | Low |
| 3.1 | Spec Bug | Health: "four fields" listed but only two/three | High |
| 3.2 | Spec Bug | `users:create` request format differs across specs | High |
| 3.3 | Spec Bug | `apikeys:update` rotate format differs across specs | High |
| 3.4 | Spec Bug | Max page size: SPEC.md (200) vs SPEC_API.md (100) | High |
| 3.5 | Spec Bug | 403 vs 401 for authorization failures | Medium |
| 3.6 | Spec Bug | API key rotation response fields differ across specs | Low |
| 3.7 | Spec Bug | `users:list` role filter not in SPEC_API/030 | Low |
| 4.1 | Quality | Mixed logging approaches (zerolog + log.Printf) | Low |
| 4.6 | Quality | `can_write` default inconsistent (users vs API keys) | Low |
| 4.6 | Quality | Aggregation returns float64 for integer fields | Low |
| 4.9 | Missing | No expired token/blacklist cleanup | Low |

---

*This analysis was generated by comparing source code at commit time against the specification files listed above. Severity ratings: High = breaks API contract or causes client confusion; Medium = functional gap or undocumented behaviour; Low = code quality or minor inconsistency.*

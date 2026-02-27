# Moon API — Gap Analysis

This document compares the source code implementation against the specification files:

- `SPEC.md` — Architecture, schema management, data types, and system behavior
- `SPEC_API.md` — Standard response patterns, query options, and API conventions
- `SPEC_AUTH.md` — Authentication flows, JWT tokens, API keys, and permissions
- `SPEC_API/010-health.md` through `SPEC_API/090-errors.md` — Individual endpoint specifications

---

## Summary

| Area | Status | Severity |
|------|--------|----------|
| API Key via `Authorization: Bearer` header | ❌ Not implemented | High |
| Rate limit HTTP status code (429 vs 400) | ❌ Mismatch | Medium |
| API key `readonly` role support | ❌ Missing | Medium |
| `readonly` role write enforcement | ⚠️ Partial | Medium |
| JWT access token default expiry | ⚠️ Inconsistency | Low |
| Error format: SPEC.md vs SPEC_API.md | ⚠️ Spec inconsistency | Low |
| `health.Service` not wired to routes | ⚠️ Dead code | Low |
| `UnifiedAuthMiddleware` not used | ⚠️ Dead code | Low |
| Pre-existing doc handler test failures | ❌ Failing tests | Low |
| All SPEC_API/010-090 endpoints | ✅ Implemented | — |

---

## 1. API Key Authentication via `Authorization: Bearer` Header

### Specification (SPEC_AUTH.md)

> Both JWT tokens and API keys use the same `Authorization: Bearer` header format.
>
> **Token Type Detection:** The server automatically detects the token type:
> - **JWT tokens:** Three base64-encoded segments separated by dots
> - **API keys:** Start with `moon_live_` prefix

### Implementation (server.go — `authMiddleware`)

The server's `authMiddleware` method processes the `Authorization: Bearer` header as follows:

1. Extracts the Bearer token.
2. Checks token blacklist.
3. **Calls JWT validation unconditionally.**
4. If JWT validation succeeds → authenticated as user.
5. If JWT validation **fails → immediately returns `"invalid or expired token"` (HTTP 401).**
6. Falls through to check `X-API-Key` header for API key auth — only reached if **no** `Authorization: Bearer` header was present.

```go
// server.go
claims, err := s.tokenService.ValidateAccessToken(token)
if err == nil {
    // Valid JWT
    return
}
// Invalid JWT token
s.writeAuthError(w, http.StatusUnauthorized, "invalid or expired token")
return  // <-- API key check never reached when Bearer is present
```

### Impact

- Sending `Authorization: Bearer moon_live_abc123...` returns **HTTP 401 "invalid or expired token"** instead of authenticating the API key.
- API keys currently only work via the `X-API-Key` header, which is **deprecated** (marked "DEPRECATED: X-API-Key header is no longer supported (removed in PRD-059)").

### Note on `UnifiedAuthMiddleware`

A `UnifiedAuthMiddleware` exists in `middleware/unified_auth.go` that correctly implements Bearer token type detection via `isAPIKey()` (checks `moon_live_` prefix). However, this middleware is **never instantiated or used** by the server. The server uses its own custom `authMiddleware` instead.

---

## 2. Rate Limiting HTTP Status Code

### Specification (SPEC_API.md)

```
If You Exceed the Limit (429 Too Many Requests):
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 60
```

### Implementation (middleware/ratelimit.go)

```go
// writeRateLimitError in ratelimit.go
w.WriteHeader(http.StatusBadRequest)  // Returns 400, not 429
```

The `normalizeErrorStatus` function in `handlers/collections.go` deliberately maps HTTP 429 → 400:

```go
// normalizeErrorStatus: 405, 409, 413, 429 → 400; 503 → 500
```

### Conflict Within Specification

`SPEC_API.md` contains an internal inconsistency:
- The "Standard Error Response" section states: _"Only the codes listed above are permitted"_ (200, 201, 400, 401, 404, 500).
- The "Rate Limiting" section shows an example returning `429 Too Many Requests`.

**Current behavior:** HTTP 400 is returned with `{"message": "rate limit exceeded"}` and a `Retry-After` header. This aligns with the "only these codes" rule but conflicts with the rate limiting example.

---

## 3. API Key `readonly` Role Not Supported

### Specification (SPEC_AUTH.md)

> Both methods support role-based access control (RBAC) with three roles: `admin`, `user`, and **`readonly`**.
>
> **API Key Access:** Each key assigned a role (`admin`, `user`, or `readonly`).

### Implementation (handlers/apikeys.go)

```go
func ValidAPIKeyRoles() []string {
    return []string{"admin", "user"}  // readonly is missing
}
```

Creating an API key with `"role": "readonly"` returns:

```json
{"message": "role must be 'admin' or 'user'"}
```

### Impact

- Machine-to-machine integrations requiring read-only access cannot use API keys with the `readonly` role.
- Users with `role: "readonly"` exist in the system, but API keys cannot be assigned this role.

---

## 4. `readonly` Role Write Enforcement

### Specification (SPEC_AUTH.md)

> **readonly Role:** Read-only access (**enforced regardless of `can_write` flag**)

### Implementation (middleware/authorization.go — `RequireWrite`)

```go
func (m *AuthorizationMiddleware) RequireWrite(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Admin always has write access
        if entity.Role == string(auth.RoleAdmin) {
            next(w, r)
            return
        }
        // Check if user has write permission (flag only, no role check)
        if !entity.CanWrite {
            // deny
        }
        next(w, r)  // allowed if CanWrite is true, regardless of role
    }
}
```

The middleware does NOT explicitly check for `role == "readonly"`. If a `readonly` user is created with `can_write: true` explicitly set in the request body, they would pass write checks.

### Mitigating Factor

User creation defaults `can_write` to `false` for `readonly` role users (when `can_write` is not explicitly provided). However, a caller can still explicitly set `can_write: true` when creating a `readonly` user, producing an inconsistent state where `RequireWrite` does not enforce the spec's "regardless of flag" guarantee.

---

## 5. JWT Access Token Default Expiry Discrepancy

### Specification (SPEC_AUTH.md)

> **Access Token:** Short-lived (configurable, **default 1 hour**)

### Implementation

**Config defaults** (`config.go`):
```go
AccessExpiry: 3600,   // 1 hour ✓
```

**Server fallback** (`server.go`):
```go
const defaultAccessExpirySeconds = 900  // 15 minutes ✗
```

The server uses this constant as a fallback when `cfg.JWT.AccessExpiry` is 0. Under normal operation with the config file, the 3600-second default is used correctly. But if config is loaded without `jwt.access_expiry` being set explicitly and the Viper default is somehow bypassed, the token would expire in 15 minutes instead of 1 hour.

---

## 6. Error Format Inconsistency in SPEC.md

### SPEC.md Example

In the "Default Values" section, SPEC.md shows this error format:

```json
{
  "code": 400,
  "error": "unknown field 'default' in column 0"
}
```

### SPEC_API.md Specification

```json
{
  "message": "A human-readable description of the error"
}
```

### Implementation

The implementation correctly uses `{"message": "..."}` format per SPEC_API.md. The example in SPEC.md is inconsistent with the actual API contract defined in SPEC_API.md.

**Recommendation:** Update the SPEC.md example to match the `{"message": "..."}` format.

---

## 7. `health.Service` Not Wired to Routes

### Implementation

The `health` package (`internal/health/health.go`) provides a full `health.Service` with:
- `LivenessHandler` — responds with a complex `HealthResponse` struct including `status: "live"`, database dialect, and collection count.
- `ReadinessHandler` — performs comprehensive checks (database ping, registry, custom checkers).

However, **the server does not use `health.Service` at all**. Instead, it has a simple inline `healthHandler` method:

```go
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    response := map[string]any{
        "moon":      s.version,
        "status":    "ok",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    }
    s.writeJSON(w, http.StatusOK, map[string]any{"data": response})
}
```

The inline handler **does match the SPEC_API/010-health.md format**. The `health.Service` is dead code.

**Recommendation:** Either wire `health.Service.LivenessHandler` to the route or remove the unused package to avoid confusion.

---

## 8. `UnifiedAuthMiddleware` Not Used

The `middleware/unified_auth.go` implements a sophisticated `UnifiedAuthMiddleware` that:
- Correctly detects API keys vs JWT tokens via Bearer header.
- Supports configurable protected/unprotected paths.
- Integrates with CORS middleware for auth bypass.

This middleware is only instantiated in tests (`unified_auth_test.go`) and is **never used by the server**. The server's `authMiddleware` in `server.go` duplicates some of the same logic but does not handle the API key via Bearer case.

**Recommendation:** Either replace the server's `authMiddleware` with `UnifiedAuthMiddleware` (fixing the Bearer API key issue), or remove the unused middleware.

---

## 9. Pre-existing Test Failures

Three tests in `internal/handlers/doc_test.go` fail on the current codebase:

| Test | Failure |
|------|---------|
| `TestDocHandler_WithPrefix` | `expected prefix in curl examples` |
| `TestDocHandler_QuickstartSection` | `expected {collection} placeholder` |
| `TestDocHandler_ErrorSection` | `expected Error Responses section` |

These failures indicate the `DocHandler` HTML/Markdown output does not include:
1. Prefix-aware curl examples when a URL prefix is configured.
2. `{collection}` placeholder in data endpoint examples.
3. An "Error Responses" section in the HTML documentation.

These are pre-existing failures separate from the spec compliance issues documented above, but they indicate the documentation endpoint is out of sync with test expectations.

---

## 10. Verified Implementations (Compliant)

The following areas are **fully compliant** with their respective specifications:

### SPEC_API/010-health.md
- `GET /health` returns `{"data": {"moon": "<version>", "status": "ok", "timestamp": "<RFC3339>"}}`
- No authentication required ✓

### SPEC_API/020-auth.md
- `POST /auth:login` ✓
- `GET /auth:me` ✓
- `POST /auth:me` (email + password update) ✓
- `POST /auth:refresh` ✓
- `POST /auth:logout` ✓
- Response formats match spec ✓
- Login rate limiting (5 attempts/15 min) ✓

### SPEC_API/030-users.md
- `POST /users:create` (admin only, wrapped in `{"data": {...}}`) ✓
- `GET /users:list` (paginated with `meta`) ✓
- `GET /users:get?id=` ✓
- `POST /users:update?id=` (email/role/can_write, reset_password, revoke_sessions) ✓
- `POST /users:destroy?id=` ✓
- Last-admin protection ✓
- Cannot modify own account via user management ✓

### SPEC_API/040-apikeys.md
- `POST /apikeys:create` ✓
- `GET /apikeys:list` ✓
- `GET /apikeys:get?id=` ✓
- `POST /apikeys:update?id=` (metadata update and rotate) ✓
- `POST /apikeys:destroy?id=` ✓
- Key only returned on creation/rotation with warning ✓

### SPEC_API/050-collection.md
- `POST /collections:create` ✓
- `GET /collections:list` ✓
- `GET /collections:get?name=` ✓
- `POST /collections:update` (add/rename/modify/remove columns, combined) ✓
- `POST /collections:destroy?name=` ✓
- Default values rejected on create/update ✓

### SPEC_API/060-data.md
- `GET /{collection}:schema` ✓
- `POST /{collection}:create` (single and batch) ✓
- `GET /{collection}:list` ✓
- `GET /{collection}:get?id=` ✓
- `POST /{collection}:update` (single and batch) ✓
- `POST /{collection}:destroy` (single and batch) ✓

### SPEC_API/070-query.md
- Filter operators: `eq`, `ne`, `gt`, `lt`, `gte`, `lte`, `like`, `in` ✓
- Sorting: `?sort=-field1,field2` ✓
- Full-text search: `?q=term` ✓
- Field selection: `?fields=field1,field2` ✓
- Limit: `?limit=N` (default 15, max 100) ✓
- Cursor pagination: `?after=<cursor>` ✓

### SPEC_API/080-aggregation.md
- `GET /{collection}:count` ✓
- `GET /{collection}:sum?field=` ✓
- `GET /{collection}:avg?field=` ✓
- `GET /{collection}:min?field=` ✓
- `GET /{collection}:max?field=` ✓

### SPEC_API/090-errors.md
- Error format `{"message": "..."}` ✓
- HTTP 400 for validation errors ✓
- HTTP 401 for authentication failures ✓
- HTTP 404 for not-found resources ✓
- HTTP 500 for server errors ✓

---

## Recommended Actions (Priority Order)

1. **[High]** Fix `authMiddleware` in `server.go` to detect `moon_live_` prefix in Bearer tokens and route to API key authentication instead of returning "invalid or expired token".

2. **[Medium]** Add `"readonly"` to `ValidAPIKeyRoles()` in `handlers/apikeys.go` to allow API key creation with the readonly role.

3. **[Medium]** Update `RequireWrite` middleware to explicitly deny write access for `role == "readonly"` regardless of the `CanWrite` flag, matching the spec guarantee.

4. **[Low]** Resolve the 429 vs 400 rate limit status code: either update the spec to explicitly permit 429 and change `writeRateLimitError` to return `http.StatusTooManyRequests`, or update the SPEC_API.md rate limiting example to show 400.

5. **[Low]** Fix the SPEC.md error format example (`{"code": 400, "error": "..."}`) to match the `{"message": "..."}` format used by all implementations and SPEC_API.md.

6. **[Low]** Fix the three failing doc handler tests or update the `DocHandler` to generate the expected content.

7. **[Low]** Remove or wire `health.Service` — if the full readiness endpoint behavior is desired, wire `health.Service.LivenessHandler` and register a `/health/ready` route; otherwise remove the dead code.

8. **[Low]** Remove or integrate `UnifiedAuthMiddleware` — either use it as the server's auth middleware (fixing item #1) or remove it to reduce dead code.

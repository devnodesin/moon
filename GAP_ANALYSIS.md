# Moon API — Gap Analysis

This document compares the source code implementation against the specification files:

- `SPEC.md` — Architecture, schema management, data types, and system behavior
- `SPEC_API.md` — Standard response patterns, query options, and API conventions
- `SPEC_AUTH.md` — Authentication flows, JWT tokens, API keys, and permissions
- `SPEC_API/010-health.md` through `SPEC_API/090-errors.md` — Individual endpoint specifications

---

## Resolved Gaps (Fixed by PRD-001 through PRD-004)

| # | Issue | Resolution |
|---|---|---|
| 1 | X-API-Key fallback in server authMiddleware | Bearer-only auth with token type detection (PRD-001) |
| 2 | X-API-Key in CORS default allowed headers | Removed from config.go and doc.go JSON appendix (PRD-001) |
| 3 | JWT access expiry default mismatch (3600 vs 900) | Centralized to 900s in config defaults (PRD-002) |
| 4 | Dead health.Service package not wired to server | Removed unused package (PRD-003) |
| 5 | Health response included `status` field not in spec | Aligned with SPEC_API/010-health.md (PRD-003) |
| 6 | Doc handler missing prefix in curl examples | Include function replaces base URL with prefix-aware URL (PRD-004) |
| 7 | Doc handler missing `{collection}` placeholder | Updated template to use `{collection}` (PRD-004) |
| 8 | Doc handler missing Error Responses heading | Added `### Error Responses` subsection (PRD-004) |
| 9 | server.go authMiddleware did not detect API keys in Bearer tokens | Token type detection routes API key prefix to API key validation (PRD-001) |

---

## Remaining Gaps

### 1. API Key `readonly` Role Not Supported

**Specification (SPEC_AUTH.md):**

> Both methods support role-based access control (RBAC) with three roles: `admin`, `user`, and **`readonly`**.
>
> **API Key Access:** Each key assigned a role (`admin`, `user`, or `readonly`).

**Implementation (handlers/apikeys.go):**

```go
func ValidAPIKeyRoles() []string {
    return []string{"admin", "user"}  // readonly is missing
}
```

Creating an API key with `"role": "readonly"` returns:

```json
{ "message": "role must be 'admin' or 'user'" }
```

**Impact:** Machine-to-machine integrations requiring read-only access cannot use API keys with the `readonly` role.

---

### 2. `readonly` Role Write Enforcement

**Specification (SPEC_AUTH.md):**

> **readonly Role:** Read-only access (**enforced regardless of `can_write` flag**)

**Implementation (middleware/authorization.go — `RequireWrite`):**

The middleware does NOT explicitly check for `role == "readonly"`. If a `readonly` user is created with `can_write: true` explicitly set, they would pass write checks. The spec requires that `readonly` role enforces read-only regardless of the flag.

**Mitigating Factor:** User creation defaults `can_write` to `false` for `readonly` role users when not explicitly provided.

---

### 3. `UnifiedAuthMiddleware` Not Used

The `middleware/unified_auth.go` implements a sophisticated `UnifiedAuthMiddleware` with:
- Token type detection for API keys vs JWT via Bearer header
- Configurable protected/unprotected paths
- CORS middleware integration for auth bypass

This middleware is only instantiated in tests (`unified_auth_test.go`) and is **never used by the server**. The server's `authMiddleware` in `server.go` now implements equivalent Bearer-only token type detection logic (fixed by PRD-001), making `UnifiedAuthMiddleware` redundant dead code.

**Recommendation:** Remove `UnifiedAuthMiddleware` or consolidate the server's authMiddleware to use it.

---

## Compliance Summary by Spec File

### SPEC_API/010-health.md ✅
Health endpoint returns exactly `{"data": {"moon": version, "timestamp": RFC3339}}` with no extra fields.

### SPEC_API/020-auth.md ✅
Login, refresh, logout, and me endpoints all aligned. Login rate limiting enforced (5/15min).

### SPEC_API/030-users.md ✅
Admin-only CRUD with password reset, session revocation, and last-admin protection.

### SPEC_API/040-apikeys.md ⚠️
Aligned except `readonly` role not accepted for API key creation (Gap #1 above).

### SPEC_API/050-collection.md ✅
Schema modification operations (add/rename/modify/remove columns) all implemented with validation.

### SPEC_API/060-data.md ✅
Batch operations with `meta.succeeded`/`meta.failed` tracking. Array format enforced.

### SPEC_API/070-query.md ✅
All filter operators, sorting, full-text search, field selection, and cursor pagination implemented.

### SPEC_API/080-aggregation.md ✅
Count, sum, avg, min, max endpoints with filter support. Integer/decimal type restriction enforced.

### SPEC_API/090-errors.md ✅
Single `message` field in error responses. HTTP status codes per spec. No 403 (mapped to 401).

### SPEC_AUTH.md ⚠️
Bearer-only auth aligned. JWT/refresh token lifecycle aligned. Rate limits aligned. `readonly` role enforcement gap (Gap #2 above).

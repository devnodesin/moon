# Moon API — Gap Analysis

This document compares the source code implementation against the specification files:

- `SPEC.md` — Architecture, schema management, data types, and system behavior
- `SPEC_API.md` — Standard response patterns, query options, and API conventions
- `SPEC_AUTH.md` — Authentication flows, JWT tokens, API keys, and permissions
- `SPEC_API/010-health.md` through `SPEC_API/090-errors.md` — Individual endpoint specifications


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
{ "message": "role must be 'admin' or 'user'" }
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

## 8. `UnifiedAuthMiddleware` Not Used

The `middleware/unified_auth.go` implements a sophisticated `UnifiedAuthMiddleware` that:

- Correctly detects API keys vs JWT tokens via Bearer header.
- Supports configurable protected/unprotected paths.
- Integrates with CORS middleware for auth bypass.

This middleware is only instantiated in tests (`unified_auth_test.go`) and is **never used by the server**. The server's `authMiddleware` in `server.go` duplicates some of the same logic but does not handle the API key via Bearer case.

**Recommendation:** Either replace the server's `authMiddleware` with `UnifiedAuthMiddleware` (fixing the Bearer API key issue), or remove the unused middleware.

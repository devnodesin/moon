## Overview

- Every protected route requires an authenticated caller. Authentication identifies the caller; authorization decides whether the caller may perform the requested operation.
- Moon supports two bearer credential types (JWT and API key) and two roles (`admin`, `user`). Write capability for the `user` role is controlled by the `can_write` flag.
- This PRD defines the authentication middleware (credential parsing and validation) and the authorization middleware (role and capability enforcement).
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Authentication Middleware

Authentication applies to all routes except the public routes (`/`, `/health`, and `POST /auth:session`).

#### Credential Extraction

- The bearer credential must be extracted from the `Authorization: Bearer <token>` header.
- If the header is absent or malformed (not `Bearer <value>`), the request must be rejected with `401 Unauthorized`.

#### Credential Type Detection

| Pattern | Type |
|---------|------|
| Three dot-separated segments | JWT access token |
| Value starting with `moon_live_` | API key |
| Anything else | rejected as `401` |

Mixed or unrecognized bearer values must be rejected immediately.

#### JWT Validation

1. Parse and verify the JWT signature using `jwt_secret`.
2. Verify the token is not expired (`exp` claim).
3. Verify the token includes a `jti` claim.
4. Check the `jti` against implementation-private revocation storage; if revoked, return `401`.
5. Extract claims: `sub` (user id), `role`, `can_write`, `jti`.
6. Look up the user by `sub` in `users`; if not found, return `401`.
7. Attach the resolved identity to the request context.

#### API Key Validation

1. Verify the key starts with `moon_live_` and is 74 characters total (10-character prefix + 64-character base62 suffix).
2. Hash the submitted key with SHA-256.
3. Look up the hash in `apikeys.key_hash`.
4. If not found, return `401`.
5. Extract `role` and `can_write` from the matching `apikeys` row.
6. Update `apikeys.last_used_at` to the current time (non-blocking; best-effort is acceptable).
7. Attach the resolved identity to the request context.

#### Endpoints That Reject API Keys

- `GET /auth:me` must reject API key credentials with `401`.
- `POST /auth:me` must reject API key credentials with `401`.
- These endpoints are JWT-only.

#### Identity Context

Authentication middleware must attach to the request context:

- Credential type (`jwt` or `apikey`)
- Caller id (user id or api key id)
- Role (`admin` or `user`)
- `can_write` (boolean)

All subsequent middleware and handlers must read this context rather than re-validating credentials.

---

### Authorization Middleware

Authorization runs after authentication and before handlers. It enforces route-level access rules based on the resolved identity.

#### Capability Matrix

| Operation | `admin` | `user` `can_write=false` | `user` `can_write=true` |
|-----------|---------|--------------------------|------------------------|
| Read public endpoints | ✅ | ✅ | ✅ |
| Query collection metadata | ✅ | ✅ | ✅ |
| Query record data | ✅ | ✅ | ✅ |
| Create/update/destroy records | ✅ | ❌ `403` | ✅ |
| Mutate collection schema | ✅ | ❌ `403` | ❌ `403` |
| Manage users (create/update/destroy) | ✅ | ❌ `403` | ❌ `403` |
| Manage API keys (create/update/destroy) | ✅ | ❌ `403` | ❌ `403` |
| Privileged resource actions | ✅ | ❌ `403` | ❌ `403` |

Privileged resource actions include:

- `action=reset_password` on `/data/users:mutate`
- `action=revoke_sessions` on `/data/users:mutate`
- `action=rotate` on `/data/apikeys:mutate`

#### Admin Safety Rules

These rules must be enforced in the relevant handlers but are part of the authorization model:

- The last remaining admin user must not be deleted.
- The last remaining admin user must not be demoted to `user` role.
- An admin must not change their own role.

#### System Collection Protection

- `users` and `apikeys` must not be creatable, renameable, modifiable, or destroyable through `/collections:mutate`. Requests attempting these operations must return `403 Forbidden`.
- `moon_*` tables must never be accessible through any API surface; requests naming them must return `400 Bad Request`.

#### Authorization Failure

All authorization failures must return `403 Forbidden` with the standard error body `{ "message": "Forbidden" }`.

---

### JWT Revocation (Implementation Note)

- JWT revocation state is implementation-private and must not be exposed through public APIs.
- The revocation mechanism must use the JWT `jti` claim as the lookup key.
- The revocation store may be in-memory, database-backed, or another mechanism as long as revocation is effective immediately and correctness does not depend on background workers.

---

### Rate Limiting (Summary; See PRD 015)

Rate limits enforced by the rate-limiting middleware:

| Traffic type | Limit |
|-------------|-------|
| Login failures | 5 per 15 minutes per IP + username |
| Authenticated JWT traffic | 100 requests per minute per user |
| Authenticated API key traffic | 1000 requests per minute per key |

Rate-limit failures return `429` with the standard error body.

## Acceptance

- `GET /auth:me` with no `Authorization` header returns `401`.
- `GET /auth:me` with an expired JWT returns `401`.
- `GET /auth:me` with a revoked JWT `jti` returns `401`.
- `GET /auth:me` with an API key returns `401`.
- `GET /data/users:query` with a valid API key returns `200`.
- `POST /data/products:mutate` (`op=create`) with a `user` whose `can_write=false` returns `403`.
- `POST /data/products:mutate` (`op=create`) with a `user` whose `can_write=true` returns `201`.
- `POST /collections:mutate` with a `user` role (any `can_write`) returns `403`.
- `POST /data/users:mutate` with `op=destroy` from a non-admin returns `403`.
- `POST /data/users:mutate` with `op=destroy` targeting the last admin returns `403`.
- `POST /data/users:mutate` with `action=reset_password` from a non-admin returns `403`.
- `POST /collections:mutate` with `op=update` targeting `users` returns `403`.
- `GET /data/moon_secret:query` returns `400`.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

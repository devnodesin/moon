## Overview

- `POST /auth:session` is the unified credential-exchange endpoint for interactive user sessions. It handles login (credential verification and token issuance), refresh (token rotation), and logout (session invalidation).
- This endpoint does not require a bearer token. Credentials are always provided in the request body.
- Refresh-session state is stored in `moon_auth_refresh_tokens`.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Endpoint

| Method | Path | Auth Required | Accepted Bearer |
|--------|------|--------------|----------------|
| `POST` | `/auth:session` | No | None |

### Request Shape

```json
{
  "op": "login | refresh | logout",
  "data": {}
}
```

- `op` is required; must be exactly one of `login`, `refresh`, or `logout`.
- `data` is required and must be a JSON object.
- Unknown fields in `data` must be ignored or rejected consistently; they must not cause a server error.

### `op=login`

#### Required `data` Fields

| Field | Type | Constraint |
|-------|------|------------|
| `username` | string | required |
| `password` | string | required |

#### Behavior

1. Look up the user by `username` (case-insensitive, after normalization to lowercase).
2. If the user does not exist, return `401 Unauthorized` with `{ "message": "Invalid credentials" }`.
3. Verify the submitted `password` against the stored `password_hash` using bcrypt.
4. If verification fails, return `401 Unauthorized` with `{ "message": "Invalid credentials" }`. Do not distinguish between unknown username and wrong password.
5. If verification succeeds:
   a. Generate a ULID `jti` and issue a signed JWT access token with claims: `sub` (user id), `jti`, `role`, `can_write`, `exp` (now + `jwt_access_expiry`).
   b. Generate a cryptographically random refresh token (high entropy, e.g. 32 bytes → base64url).
   c. Hash the refresh token with SHA-256 and insert a row into `moon_auth_refresh_tokens`.
   d. Update `users.last_login_at` to the current time.
   e. Return `200 OK` with the session payload.

#### Rate Limiting

- Failed login attempts must be rate-limited: 5 failures per 15 minutes per IP address and username combination.
- On violation, return `429 Too Many Requests` with the standard error body.

#### Success Response (`200 OK`)

```json
{
  "message": "Login successful",
  "data": [
    {
      "access_token": "eyJhbGciOi...",
      "refresh_token": "SEb54NKd...",
      "expires_at": "2026-02-28T06:52:38Z",
      "token_type": "Bearer",
      "user": {
        "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
        "username": "newuser",
        "email": "newuser@example.com",
        "role": "user",
        "can_write": true,
        "created_at": "2026-02-01T10:00:00Z",
        "updated_at": "2026-02-28T06:52:38Z",
        "last_login_at": "2026-02-28T06:52:38Z"
      }
    }
  ]
}
```

- `expires_at` is the access token expiry in RFC3339.
- `token_type` is always `"Bearer"`.
- `user` contains only API-visible user fields; never `password_hash`.

### `op=refresh`

#### Required `data` Fields

| Field | Type | Constraint |
|-------|------|------------|
| `refresh_token` | string | required |

#### Behavior

1. Hash the submitted refresh token with SHA-256.
2. Look up the hash in `moon_auth_refresh_tokens`.
3. If not found, expired (`expires_at` < now), or revoked (`revoked_at` is not null), return `401 Unauthorized`.
4. Refresh tokens are single-use: mark the found row as revoked (`revoked_at = now`, `revocation_reason = "rotated"`) and `last_used_at = now` in a single atomic operation.
5. Issue a new JWT access token and a new refresh token (same logic as login step 5a–5c).
6. Return `200 OK` with the session payload (same shape as login success).

```json
{
  "message": "Token refreshed successfully",
  "data": [{ ... }]
}
```

### `op=logout`

#### Required `data` Fields

| Field | Type | Constraint |
|-------|------|------------|
| `refresh_token` | string | required |

#### Behavior

1. Hash the submitted refresh token with SHA-256.
2. Look up the hash in `moon_auth_refresh_tokens`.
3. If found and not already revoked, set `revoked_at = now`, `revocation_reason = "logout"`.
4. If not found or already revoked, still return `200 OK` (idempotent logout).
5. Return `200 OK` with message-only response.

#### Success Response (`200 OK`)

```json
{
  "message": "Logged out successfully"
}
```

This is the only documented endpoint that returns a message-only response with no `data` array.

### JWT Claims

All issued JWTs must include:

| Claim | Value |
|-------|-------|
| `sub` | user id (ULID string) |
| `jti` | unique ULID |
| `role` | `admin` or `user` |
| `can_write` | boolean |
| `exp` | Unix epoch expiry |
| `iat` | Unix epoch issued-at |

JWT signing must use HS256 with `jwt_secret`.

### Refresh Token Storage

- Raw refresh tokens must never be stored.
- Only the SHA-256 hash of the raw token is persisted.
- Each `op=refresh` or `op=login` inserts a new row in `moon_auth_refresh_tokens`.
- Each `op=refresh` or `op=logout` updates the matching row.

### Validation Failures

| Condition | Status | Message |
|-----------|--------|---------|
| `op` missing | `400` | descriptive |
| `op` unknown value | `400` | descriptive |
| `data` missing or not an object | `400` | descriptive |
| required `data` field missing | `400` | descriptive |
| invalid credentials (login) | `401` | `"Invalid credentials"` |
| expired or revoked refresh token | `401` | descriptive |
| login rate limit exceeded | `429` | standard error body |

## Acceptance

- `POST /auth:session` with `op=login` and valid credentials returns `200` with `access_token`, `refresh_token`, `user` (no `password_hash`), and `expires_at`.
- `POST /auth:session` with `op=login` and wrong password returns `401` with `{ "message": "Invalid credentials" }`.
- `POST /auth:session` with `op=login` and unknown username returns `401` with the same message (no distinction).
- `POST /auth:session` with `op=refresh` and a valid refresh token returns `200` with new tokens.
- Using the same refresh token a second time returns `401` (single-use enforcement).
- `POST /auth:session` with `op=logout` and a valid refresh token returns `200` with message only and no `data` field.
- Calling logout with an already-revoked token returns `200` (idempotent).
- `POST /auth:session` with missing `op` returns `400`.
- `POST /auth:session` with `op=login` missing `username` or `password` returns `400`.
- 6 consecutive failed logins from the same IP + username within 15 minutes: the 6th returns `429`.
- JWT access token contains `jti`, `sub`, `role`, `can_write`, and `exp` claims.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

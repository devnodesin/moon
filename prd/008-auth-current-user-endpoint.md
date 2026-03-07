## Overview

- `GET /auth:me` returns the current authenticated user's profile. `POST /auth:me` allows the user to update their own email or password.
- Both endpoints require a JWT bearer token. API keys must be rejected.
- Password changes must immediately invalidate all active sessions for the user.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Endpoints

| Method | Path | Auth Required | Accepted Bearer |
|--------|------|--------------|----------------|
| `GET` | `/auth:me` | Yes | JWT only |
| `POST` | `/auth:me` | Yes | JWT only |

- API key bearer credentials must be rejected on both endpoints with `401 Unauthorized`.
- The authenticated user is identified from the validated JWT `sub` claim.

---

## `GET /auth:me`

### Behavior

1. Validate the JWT bearer token (signature, expiry, revocation via `jti`).
2. Look up the user by the `sub` claim (user id) in `users`.
3. If the user does not exist, return `401 Unauthorized`.
4. Return the API-visible user fields.

### Success Response (`200 OK`)

```json
{
  "message": "Current user retrieved successfully",
  "data": [
    {
      "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
      "username": "newuser",
      "email": "newuser@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-01T10:00:00Z",
      "updated_at": "2026-02-28T07:52:40Z",
      "last_login_at": "2026-02-28T06:52:38Z"
    }
  ]
}
```

- `data` is always an array with exactly one user object.
- `password_hash` must never appear in the response.
- All other implementation-private fields must not appear in the response.

---

## `POST /auth:me`

### Request Shape

```json
{
  "data": {}
}
```

- No `op` field is used on this endpoint.
- `data` is required and must be a JSON object.
- At least one of `email` or `password` must be provided.

### Updatable Fields

| Field | Type | Constraint |
|-------|------|------------|
| `email` | string | optional; must be a valid, unique email |
| `password` | string | optional; must satisfy password policy |
| `old_password` | string | required when `password` is provided |

### Non-Writable Fields

The following fields must be rejected if submitted; writing them must return `400 Bad Request`:

- `id`, `username`, `role`, `can_write`, `created_at`, `updated_at`, `last_login_at`, `password_hash`

### Validation Rules

- `data` must not be empty; at least `email` or `password` must be present.
- If `email` is provided:
  - Must be a syntactically valid email address.
  - Must be unique across `users` (case-insensitive).
  - Must be normalized to lowercase before storage and uniqueness check.
- If `password` is provided:
  - `old_password` is required.
  - `old_password` must match the user's current `password_hash` (bcrypt verification).
  - New `password` must satisfy the password policy:
    - Minimum 8 characters.
    - At least one lowercase letter.
    - At least one uppercase letter.
    - At least one digit.
  - New password must be hashed with bcrypt at cost 12.

### Behavior for Email Update

1. Validate the JWT.
2. Look up the user.
3. Validate the new email (syntactically valid, unique).
4. Update `users.email` and `users.updated_at`.
5. Return `200 OK` with the updated user profile.

### Behavior for Password Update

1. Validate the JWT.
2. Look up the user.
3. Verify `old_password` against `password_hash`.
4. Validate new `password` against password policy.
5. Hash new password with bcrypt cost 12.
6. Update `users.password_hash` and `users.updated_at`.
7. Revoke all active refresh tokens for the user in `moon_auth_refresh_tokens` (set `revoked_at = now`, `revocation_reason = "password_changed"` for all non-revoked rows with matching `user_id`).
8. Return `200 OK` with the updated user profile and the message `"Password updated successfully. Sign in again."`.

### Success Response — Email Update (`200 OK`)

```json
{
  "message": "Current user updated successfully",
  "data": [
    {
      "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-01T10:00:00Z",
      "updated_at": "2026-02-28T08:10:00Z",
      "last_login_at": "2026-02-28T06:52:38Z"
    }
  ]
}
```

### Success Response — Password Update (`200 OK`)

```json
{
  "message": "Password updated successfully. Sign in again.",
  "data": [
    {
      "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-01T10:00:00Z",
      "updated_at": "2026-02-28T08:20:00Z",
      "last_login_at": "2026-02-28T06:52:38Z"
    }
  ]
}
```

### Validation Failures

| Condition | Status | Note |
|-----------|--------|------|
| Bearer token is an API key | `401` | API keys not accepted |
| JWT invalid, expired, or revoked | `401` | standard auth failure |
| `data` missing or not an object | `400` | |
| `data` is empty (no updatable field) | `400` | |
| Non-writable field submitted | `400` | name the field in message |
| `email` not a valid address | `400` | |
| `email` already in use | `409` | conflict |
| `password` provided without `old_password` | `400` | |
| `old_password` incorrect | `401` | |
| `password` fails policy | `400` | |

## Acceptance

- `GET /auth:me` with a valid JWT returns `200` with the user profile; `password_hash` is absent.
- `GET /auth:me` with an API key returns `401`.
- `GET /auth:me` with an expired JWT returns `401`.
- `POST /auth:me` with `{ "data": { "email": "new@example.com" } }` updates the email and returns the updated profile.
- `POST /auth:me` with an already-used email returns `409`.
- `POST /auth:me` with `{ "data": { "password": "NewPass123", "old_password": "OldPass123" } }` updates the password and returns the updated profile with message `"Password updated successfully. Sign in again."`.
- After a password change, the previous refresh token cannot be used (`401` on refresh).
- `POST /auth:me` with a wrong `old_password` returns `401`.
- `POST /auth:me` with `{ "data": { "role": "admin" } }` returns `400`.
- `POST /auth:me` with an empty `data` object returns `400`.
- `POST /auth:me` with an API key returns `401`.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

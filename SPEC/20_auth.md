# Authentication API

Moon exposes two authentication surfaces:

- `/auth:session` for login, refresh, and logout
- `/auth:me` for the current authenticated user

## Authentication Rules by Endpoint

| Endpoint | Method | Bearer Token Required | Accepted Bearer Type |
| -------- | ------ | --------------------- | -------------------- |
| `/auth:session` | `POST` | No | None |
| `/auth:me` | `GET` | Yes | JWT only |
| `/auth:me` | `POST` | Yes | JWT only |

Additional rules:

- `/auth:session` uses credentials in the request body, not bearer authentication.
- API keys must not be accepted on `/auth:me`.
- Access-token revocation is checked using JWT `jti`.
- Refresh-session state lives in `moon_auth_refresh_tokens` and must never be exposed through public APIs.
- JWT revocation state is implementation-private and must never be exposed through public APIs.

## `POST /auth:session`

`POST /auth:session` is the credential-exchange endpoint.

### Request Shape

```json
{
  "op": "login | refresh | logout",
  "data": {}
}
```

Rules:

- `op` is required.
- `op` must be exactly one of `login`, `refresh`, or `logout`.
- `data` is required and must be an object.

### Validation by Operation

#### `op=login`

Required fields in `data`:

- `username`
- `password`

#### `op=refresh`

Required fields in `data`:

- `refresh_token`

#### `op=logout`

Required fields in `data`:

- `refresh_token`

### Session Response Payload

Successful `login` and `refresh` responses return one session payload inside `data`:

```json
{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "base64-or-similar-token",
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
```

Rules:

- `access_token` is a JWT access token.
- The JWT must include a unique `jti` claim.
- `refresh_token` is a stateful refresh credential.
- `user` contains the API-visible user fields only.

### Login Example

Request:

```json
{
  "op": "login",
  "data": {
    "username": "newuser",
    "password": "UserPass123"
  }
}
```

Response `200 OK`:

```json
{
  "message": "Login successful",
  "data": [
    {
      "access_token": "eyJhbGciOi...",
      "refresh_token": "SEb54NKdpecktQN0s2qjSziWlhdWM8r-Ts6TzQ-jOT4=",
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

### Refresh Example

Request:

```json
{
  "op": "refresh",
  "data": {
    "refresh_token": "SEb54NKdpecktQN0s2qjSziWlhdWM8r-Ts6TzQ-jOT4="
  }
}
```

Response `200 OK`:

```json
{
  "message": "Token refreshed successfully",
  "data": [
    {
      "access_token": "eyJhbGciOi...",
      "refresh_token": "aDSM1M5z61WgwHfEHcgTZxqhMgjC0PbrCtg1iaKU7bw=",
      "expires_at": "2026-02-28T07:52:40Z",
      "token_type": "Bearer",
      "user": {
        "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
        "username": "newuser",
        "email": "newuser@example.com",
        "role": "user",
        "can_write": true,
        "created_at": "2026-02-01T10:00:00Z",
        "updated_at": "2026-02-28T07:52:40Z",
        "last_login_at": "2026-02-28T06:52:38Z"
      }
    }
  ]
}
```

### Logout Example

Request:

```json
{
  "op": "logout",
  "data": {
    "refresh_token": "aDSM1M5z61WgwHfEHcgTZxqhMgjC0PbrCtg1iaKU7bw="
  }
}
```

Response `200 OK`:

```json
{
  "message": "Logged out successfully"
}
```

## `GET /auth:me`

Returns the current authenticated user.

Rules:

- Requires `Authorization: Bearer <jwt>`.
- API keys must be rejected.
- The response returns one user object inside `data`.

Response `200 OK`:

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

## `POST /auth:me`

Updates the current authenticated user.

### Request Shape

```json
{
  "data": {}
}
```

Rules:

- Requires `Authorization: Bearer <jwt>`.
- API keys must be rejected.
- No `op` field is used on this endpoint.
- `data` is required and must be an object.
- At least one supported updatable field must be present.

### Supported Updatable Fields

- `email`
- `password`
- `old_password`

Validation rules:

- At least one of `email` or `password` must be provided.
- If `email` is provided, it must be a valid and unique email address.
- If `password` is provided, `old_password` is required and must match the current password.
- Password changes must satisfy the password policy defined in `SPEC.md`.
- Fields such as `id`, `username`, `role`, `can_write`, `created_at`, `updated_at`, and `last_login_at` are not writable through `/auth:me`.

### Change Email Example

Request:

```json
{
  "data": {
    "email": "newemail@example.com"
  }
}
```

Response `200 OK`:

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

### Change Password Example

Request:

```json
{
  "data": {
    "old_password": "UserPass123",
    "password": "NewSecurePass456"
  }
}
```

Response `200 OK`:

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

Additional rule:

- Successful password changes must invalidate affected sessions immediately.

See `SPEC/10_error.md` for error handling.

---

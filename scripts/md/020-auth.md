### Login

Authenticate user and retrieve access and refresh tokens.

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "smart_login",
        "data": {
          "email": "moonuser@example.com",
          "password": "UserPass123#"
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Login successful",
  "data": [
    {
      "access_token": "$ACCESS_TOKEN",
      "refresh_token": "$REFRESH_TOKEN",
      "expires_at": "2026-11-09T05:23:30Z",
      "token_type": "Bearer",
      "user": {
        "id": "01M6Z8R1890XKF2BPRWPBB5SWZ",
        "username": "moonuser",
        "email": "moonuser@example.com",
        "role": "user",
        "can_write": false,
        "created_at": "2026-11-09T04:23:29Z",
        "updated_at": "2026-11-09T04:23:29Z",
        "last_login_at": "2026-11-09T04:23:30Z"
      }
    }
  ]
}
```

### Get Current User

Fetch details of the currently authenticated user.

```bash
curl -s -X GET "http://localhost:6000/auth:me" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Current user retrieved successfully",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-11-09T04:23:29Z",
      "email": "moonuser@example.com",
      "id": "01M6Z8R1890XKF2BPRWPBB5SWZ",
      "last_login_at": "2026-11-09T04:23:30Z",
      "role": "user",
      "updated_at": "2026-11-09T04:23:30Z",
      "username": "moonuser"
    }
  ]
}
```

### Update Current User (Change Email)

Update email address for the current user.

```bash
curl -s -X POST "http://localhost:6000/auth:me" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": {
          "email": "admin_updated@example.com"
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Current user updated successfully",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-11-09T04:23:29Z",
      "email": "admin_updated@example.com",
      "id": "01M6Z8R1890XKF2BPRWPBB5SWZ",
      "last_login_at": "2026-11-09T04:23:30Z",
      "role": "user",
      "updated_at": "2026-11-09T04:23:30Z",
      "username": "moonuser"
    }
  ]
}
```

### Update Current User (Change Password)

Change password for the current user. Should require old_password and new password, and invalidate session on success.

```bash
curl -s -X POST "http://localhost:6000/auth:me" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": {
          "old_password": "UserPass123#",
          "password": "UserPass456#"
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Password updated successfully. Sign in again.",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-11-09T04:23:29Z",
      "email": "admin_updated@example.com",
      "id": "01M6Z8R1890XKF2BPRWPBB5SWZ",
      "last_login_at": "2026-11-09T04:23:30Z",
      "role": "user",
      "updated_at": "2026-11-09T04:23:30Z",
      "username": "moonuser"
    }
  ]
}
```

### Refresh Token

Generate a new access token using the refresh token.

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "refresh",
        "data": {
          "refresh_token": "$REFRESH_TOKEN"
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Token refreshed successfully",
  "data": [
    {
      "access_token": "$ACCESS_TOKEN",
      "refresh_token": "$REFRESH_TOKEN",
      "expires_at": "2026-11-09T05:23:30Z",
      "token_type": "Bearer",
      "user": {
        "id": "01M6Z8R1890XKF2BPRWPBB5SWZ",
        "username": "moonuser",
        "email": "moonuser@example.com",
        "role": "user",
        "can_write": false,
        "created_at": "2026-11-09T04:23:29Z",
        "updated_at": "2026-11-09T04:23:29Z",
        "last_login_at": "2026-11-09T04:23:30Z"
      }
    }
  ]
}
```

**Response (401 Unauthorized):**

```json
{
  "message": "Invalid or expired refresh token"
}
```

### Logout

Invalidate the current session and revoke the refresh token.

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "logout",
        "data": {
          "refresh_token": "$REFRESH_TOKEN"
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Logged out successfully"
}
```

### Integration Clarifications — Verify Against Live Server

The following items are not confirmed against your local server at `http://localhost:6000`. Treat them as open questions, not documented behavior.

- **Smart login (password) success shape — documented.** `POST /auth:session` with `op=smart_login` and `email`+`password` returns `200 OK` with `data[0]` containing `access_token`, `refresh_token`, `expires_at`, `token_type`, and `user`. See the "Login" section above.
- **Smart login (OTP) steps 3 & 4 response shape — NOT documented.** It is unconfirmed whether steps 3 and 4 return the same token envelope (`data[0]` with `access_token`/`refresh_token`/`user`) as the password flow. Confirm against the live server before coding the client.
- **`challenge` field semantics — NOT documented.** It is unconfirmed whether the client should branch on `password` vs `code` presence in the request rather than parsing `challenge` strings. Confirm against live behavior.
- **Remember-me / session persistence — conditional.** If smart login returns the same tokens as login, the existing `authStore.setSession(tokens, rememberMe)` storage logic can be reused unchanged. This is conditional on the token envelope matching the password flow (see above).
- **Rate-limit UX (429) — documented.** `auth:session` returns `429` when rate limited (5 requests / 15 min per email). A `retry-after` header is not documented; confirm against the live server whether one is sent.

### Smart Login Client Guide

Authenticate via `POST /auth:session` with `op=smart_login`.

#### Step 1 — Initiate

Request a smart login challenge for an email address:

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "smart_login",
        "data": {
          "email": "user@example.com"
        }
      }
    ' | jq .
```

Response:

```json
{
  "challenge": "credentials",
  "message": "Enter your password",
  "login_method": "password"
}
```

`login_method` is either `password` (message: "Enter your password") or `otp`
(message: "Enter your OTP"). The client uses `login_method` to decide which
credential field to display. An unknown email defaults to `password`.

Accepted exception (documented in SPEC.md "Email / OTP Login"): the initiate
step deliberately reveals `login_method` to keep the UI fully backend-driven.
This allows probing whether an email is registered as an OTP account.

#### Step 2 — Request an OTP (optional)

Trigger an OTP email for accounts that use OTP login:

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "smart_login",
        "data": {
          "email": "user@example.com",
          "request_otp": true
        }
      }
    ' | jq .
```

Response:

```json
{
  "challenge": "credentials",
  "message": "Enter your password",
  "login_method": "password"
}
```

OTP email is sent if the account uses OTP login. The response `message` /
`login_method` reflect the account's configured login method (see Step 1).

#### Step 3 — Verify password

Complete login with the user's password:

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "smart_login",
        "data": {
          "email": "user@example.com",
          "password": "..."
        }
      }
    ' | jq .
```

#### Step 4 — Verify OTP

Complete login with the OTP code:

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "smart_login",
        "data": {
          "email": "user@example.com",
          "code": "123456"
        }
      }
    ' | jq .
```

#### Errors

| Status | Meaning |
|--------|---------|
| 400 | Missing or invalid fields |
| 401 | Invalid credentials or OTP |
| 429 | Rate limited (5 requests / 15 min per email) |
| 503 | OTP not configured |

#### Migration

Replace `login`, `otp_request`, and `otp_verify` ops with `smart_login`.

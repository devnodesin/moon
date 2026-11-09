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

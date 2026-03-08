### Login

Authenticate user and retrieve access and refresh tokens.

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "login",
        "data": {
          "username": "moonuser",
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
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjYW5fd3JpdGUiOmZhbHNlLCJleHAiOjE3NzI5OTE1MTQsImlhdCI6MTc3Mjk4NzkxNCwianRpIjoiMDFLSzc1MTZKWlZDUlpaSlIzU0ZWR0QwMzUiLCJyb2xlIjoidXNlciIsInN1YiI6IjAxS0s3NTE1RllFWE40U0I5RFZXRkVDNEJCIn0.F3Q0FypBDT6xcwlHPWwBdiC7CgSUNLCw2AJdOJNL0Co",
      "refresh_token": "-8ktGghg2DnS2xc0sHchDqvlyMYZ-LFYx3JmcwpiIts",
      "expires_at": "2026-03-08T17:38:34Z",
      "token_type": "Bearer",
      "user": {
        "id": "01KK7515FYEXN4SB9DVWFEC4BB",
        "username": "moonuser",
        "email": "moonuser@example.com",
        "role": "user",
        "can_write": false,
        "created_at": "2026-03-08T16:38:33Z",
        "updated_at": "2026-03-08T16:38:33Z",
        "last_login_at": "2026-03-08T16:38:34Z"
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
      "created_at": "2026-03-08T16:38:33Z",
      "email": "moonuser@example.com",
      "id": "01KK7515FYEXN4SB9DVWFEC4BB",
      "last_login_at": "2026-03-08T16:38:34Z",
      "role": "user",
      "updated_at": "2026-03-08T16:38:34Z",
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
      "created_at": "2026-03-08T16:38:33Z",
      "email": "admin_updated@example.com",
      "id": "01KK7515FYEXN4SB9DVWFEC4BB",
      "last_login_at": "2026-03-08T16:38:34Z",
      "role": "user",
      "updated_at": "2026-03-08T16:38:35Z",
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
      "created_at": "2026-03-08T16:38:33Z",
      "email": "admin_updated@example.com",
      "id": "01KK7515FYEXN4SB9DVWFEC4BB",
      "last_login_at": "2026-03-08T16:38:34Z",
      "role": "user",
      "updated_at": "2026-03-08T16:38:36Z",
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
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjYW5fd3JpdGUiOmZhbHNlLCJleHAiOjE3NzI5OTE1MTcsImlhdCI6MTc3Mjk4NzkxNywianRpIjoiMDFLSzc1MTlLN05DVFk1N1lXUkEwWjBKOVEiLCJyb2xlIjoidXNlciIsInN1YiI6IjAxS0s3NTE1RllFWE40U0I5RFZXRkVDNEJCIn0.ru2B7SMAvLWw-hQsf9NWz1mg80-avAJxzzyOvtPg3X8",
      "refresh_token": "HbgoTmbahI5NX9oRrX9V5sUtZGYdJcmVPpR8SRtKWNA",
      "expires_at": "2026-03-08T17:38:37Z",
      "token_type": "Bearer",
      "user": {
        "id": "01KK7515FYEXN4SB9DVWFEC4BB",
        "username": "moonuser",
        "email": "admin_updated@example.com",
        "role": "user",
        "can_write": false,
        "created_at": "2026-03-08T16:38:33Z",
        "updated_at": "2026-03-08T16:38:37Z",
        "last_login_at": "2026-03-08T16:38:37Z"
      }
    }
  ]
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

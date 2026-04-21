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
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjYW5fd3JpdGUiOmZhbHNlLCJleHAiOjE3NzY3ODcxMTEsImlhdCI6MTc3Njc4MzUxMSwianRpIjoiMDFLUFI4U0tOQlZYV0ZOQTIxRkRENENHRFMiLCJyb2xlIjoidXNlciIsInN1YiI6IjAxS1BSOFNKVkNYWkE1MTg1SzFNVzMxOE40In0.7IJGtc5_R3I1xfh7vcDC4df1cDiNB4t68h8Z-Ei_YC4",
      "refresh_token": "OIDUSNLYbbkVF50h9yUIFHpeeme7LJznuAGVzWozing",
      "expires_at": "2026-04-21T15:58:31Z",
      "token_type": "Bearer",
      "user": {
        "id": "01KPR8SJVCXZA5185K1MW318N4",
        "username": "moonuser",
        "email": "moonuser@example.com",
        "role": "user",
        "can_write": false,
        "created_at": "2026-04-21T14:58:30Z",
        "updated_at": "2026-04-21T14:58:30Z",
        "last_login_at": "2026-04-21T14:58:31Z"
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
      "created_at": "2026-04-21T14:58:30Z",
      "email": "moonuser@example.com",
      "id": "01KPR8SJVCXZA5185K1MW318N4",
      "last_login_at": "2026-04-21T14:58:31Z",
      "role": "user",
      "updated_at": "2026-04-21T14:58:31Z",
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
      "created_at": "2026-04-21T14:58:30Z",
      "email": "admin_updated@example.com",
      "id": "01KPR8SJVCXZA5185K1MW318N4",
      "last_login_at": "2026-04-21T14:58:31Z",
      "role": "user",
      "updated_at": "2026-04-21T14:58:31Z",
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
      "created_at": "2026-04-21T14:58:30Z",
      "email": "admin_updated@example.com",
      "id": "01KPR8SJVCXZA5185K1MW318N4",
      "last_login_at": "2026-04-21T14:58:31Z",
      "role": "user",
      "updated_at": "2026-04-21T14:58:32Z",
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
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjYW5fd3JpdGUiOmZhbHNlLCJleHAiOjE3NzY3ODcxMTMsImlhdCI6MTc3Njc4MzUxMywianRpIjoiMDFLUFI4U05UV1lQNUNSWFdQVzBGWVc1WkUiLCJyb2xlIjoidXNlciIsInN1YiI6IjAxS1BSOFNKVkNYWkE1MTg1SzFNVzMxOE40In0.GOqe3k4-CwuOX3q2Ek5kX978p_V3Qhs_LdL7-x32dCU",
      "refresh_token": "krd0cdFFMhhHbmJ5OOVDAn7kexI_VN-8jhfNaH6Bvkc",
      "expires_at": "2026-04-21T15:58:33Z",
      "token_type": "Bearer",
      "user": {
        "id": "01KPR8SJVCXZA5185K1MW318N4",
        "username": "moonuser",
        "email": "admin_updated@example.com",
        "role": "user",
        "can_write": false,
        "created_at": "2026-04-21T14:58:30Z",
        "updated_at": "2026-04-21T14:58:33Z",
        "last_login_at": "2026-04-21T14:58:33Z"
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

### Login

Authenticate user and retrieve access and refresh tokens.

```bash
curl -s -X POST "http://localhost:6006/auth:login" \
    -H "Content-Type: application/json" \
    -d '
      {
        "username": "newuser",
        "password": "UserPass123#"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSkNESko1Q0dOWUI0WDZQMlhHN0pHVlQiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSkNESko1Q0dOWUI0WDZQMlhHN0pHVlQiLCJleHAiOjE3NzIwOTQ1MDMsIm5iZiI6MTc3MjA5MDg3MywiaWF0IjoxNzcyMDkwOTAzfQ.aAu_3Ax4A0PBIXtA91AlQHPMl7a6bh6CzBiPKI4_Pjw",
    "refresh_token": "HvYdfIspl7S_D8MrM0rHDbHMikSloPfCZJfOxxS-kME=",
    "expires_at": "2026-02-26T08:28:23.831470496Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KJCDJJ5CGNYB4X6P2XG7JGVT",
      "username": "newuser",
      "email": "newuser@example.com",
      "role": "user",
      "can_write": true
    }
  },
  "message": "Login successful"
}
```

### Get Current User

Fetch details of the currently authenticated user.

```bash
curl -s -X GET "http://localhost:6006/auth:me" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJCDJJ5CGNYB4X6P2XG7JGVT",
    "username": "newuser",
    "email": "newuser@example.com",
    "role": "user",
    "can_write": true
  }
}
```

### Update Current User (Change email)

Change email for current user.

```bash
curl -s -X POST "http://localhost:6006/auth:me" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "email": "newemail@example.com"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJCDJJ5CGNYB4X6P2XG7JGVT",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  },
  "message": "User updated successfully"
}
```

### Update Current User (Change Password)

Change password for current user.

```bash
curl -s -X POST "http://localhost:6006/auth:me" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "old_password": "UserPass123#",
        "password": "NewSecurePass456"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJCDJJ5CGNYB4X6P2XG7JGVT",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  },
  "message": "Password updated successfully. Please login again."
}
```

### Refresh Token

Generate new access token using refresh token.

```bash
curl -s -X POST "http://localhost:6006/auth:refresh" \
    -H "Content-Type: application/json" \
    -d '
      {
        "refresh_token": "$REFRESH_TOKEN"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSkNESko1Q0dOWUI0WDZQMlhHN0pHVlQiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSkNESko1Q0dOWUI0WDZQMlhHN0pHVlQiLCJleHAiOjE3NzIwOTQ1MDUsIm5iZiI6MTc3MjA5MDg3NSwiaWF0IjoxNzcyMDkwOTA1fQ.3hayP_WUaGEUlcnSWotMzjt42KHY5G7Oblt-4EziYDs",
    "refresh_token": "gD_4bWQeYhu--AoSpGBB3h2N9By0vkmQy3WFNDv2pwI=",
    "expires_at": "2026-02-26T08:28:25.573036556Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KJCDJJ5CGNYB4X6P2XG7JGVT",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true
    }
  },
  "message": "Token refreshed successfully"
}
```

### Logout

Invalidate current session and refresh token.

```bash
curl -s -X POST "http://localhost:6006/auth:logout" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "refresh_token": "$REFRESH_TOKEN"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Logged out successfully"
}
```

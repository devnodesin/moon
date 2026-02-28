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
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSkhDV05ESjNRTjJaM0NSM1k5SDM2QTYiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSkhDV05ESjNRTjJaM0NSM1k5SDM2QTYiLCJleHAiOjE3NzIyNjE1NTgsIm5iZiI6MTc3MjI1NzkyOCwiaWF0IjoxNzcyMjU3OTU4fQ.lZ8oFckKcKAKLkWAAQ-CibKrNCKN55cUrDr1zbxadAI",
    "refresh_token": "SEb54NKdpecktQN0s2qjSziWlhdWM8r-Ts6TzQ-jOT4=",
    "expires_at": "2026-02-28T06:52:38.69599201Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
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
    "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
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
    "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
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
    "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
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
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSkhDV05ESjNRTjJaM0NSM1k5SDM2QTYiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSkhDV05ESjNRTjJaM0NSM1k5SDM2QTYiLCJleHAiOjE3NzIyNjE1NjAsIm5iZiI6MTc3MjI1NzkzMCwiaWF0IjoxNzcyMjU3OTYwfQ.b3miIPvXZGt-7-58mayTA3Zy79q53S1MOnx0beT59mg",
    "refresh_token": "aDSM1M5z61WgwHfEHcgTZxqhMgjC0PbrCtg1iaKU7bw=",
    "expires_at": "2026-02-28T06:52:40.914567576Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
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

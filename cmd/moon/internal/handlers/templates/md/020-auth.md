### Login

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
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSFNUOEE3V1I0Q1pNRU03SzQ4S0ZFN1EiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSFNUOEE3V1I0Q1pNRU03SzQ4S0ZFN1EiLCJleHAiOjE3NzE0NzAyNjUsIm5iZiI6MTc3MTQ2NjYzNSwiaWF0IjoxNzcxNDY2NjY1fQ.Osb7IIMBx2iCMhvhpKTN0lz-_dum0JfYGUMfEC3SWhs",
    "refresh_token": "xWqY732xkDmydNXtuyDbNagw3kBb8ZnHXC_439mCVS4=",
    "expires_at": "2026-02-19T03:04:25.550813514Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KHST8A7WR4CZMEM7K48KFE7Q",
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

```bash
curl -s -X GET "http://localhost:6006/auth:me" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHST8A7WR4CZMEM7K48KFE7Q",
    "username": "newuser",
    "email": "newuser@example.com",
    "role": "user",
    "can_write": true
  }
}
```

### Update Current User (Change email)

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
    "id": "01KHST8A7WR4CZMEM7K48KFE7Q",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  },
  "message": "User updated successfully"
}
```

### Update Current User (Change Password)

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
    "id": "01KHST8A7WR4CZMEM7K48KFE7Q",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  },
  "message": "Password updated successfully. Please login again."
}
```

### Refresh Token

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
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSFNUOEE3V1I0Q1pNRU03SzQ4S0ZFN1EiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSFNUOEE3V1I0Q1pNRU03SzQ4S0ZFN1EiLCJleHAiOjE3NzE0NzAyNjcsIm5iZiI6MTc3MTQ2NjYzNywiaWF0IjoxNzcxNDY2NjY3fQ.zZ2DDNVADdtP6kz2xe6jqAiSJwJWZWiPHT7KR-V5ems",
    "refresh_token": "vSyq-5vOAtnPCVO7W6zZslmzb53Ca9Py235G2gjg91g=",
    "expires_at": "2026-02-19T03:04:27.449751842Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KHST8A7WR4CZMEM7K48KFE7Q",
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

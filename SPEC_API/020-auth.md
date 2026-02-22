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
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSjI3VlpSUDBLUTFETUJGM1I3RjNNRjEiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSjI3VlpSUDBLUTFETUJGM1I3RjNNRjEiLCJleHAiOjE3NzE3NTI5NzYsIm5iZiI6MTc3MTc0OTM0NiwiaWF0IjoxNzcxNzQ5Mzc2fQ.s7kgoJP_zZdVk9oZrevBeQaaHwJCdEFep_g0tqzgzkc",
    "refresh_token": "C8DvtpvU1n0jP0xmvSUAcZfl2Qs1pZoqb00JBO6QWHU=",
    "expires_at": "2026-02-22T09:36:16.955786225Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KJ27VZRP0KQ1DMBF3R7F3MF1",
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
    "id": "01KJ27VZRP0KQ1DMBF3R7F3MF1",
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
    "id": "01KJ27VZRP0KQ1DMBF3R7F3MF1",
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
    "id": "01KJ27VZRP0KQ1DMBF3R7F3MF1",
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
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSjI3VlpSUDBLUTFETUJGM1I3RjNNRjEiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSjI3VlpSUDBLUTFETUJGM1I3RjNNRjEiLCJleHAiOjE3NzE3NTI5NzgsIm5iZiI6MTc3MTc0OTM0OCwiaWF0IjoxNzcxNzQ5Mzc4fQ.1F3VTk1HWlRD4BY19C2pTpJUYsAMYwpqkyILpoMEP8c",
    "refresh_token": "9PQ_AubqD_OHgWyyGo0_vu7WqdCp9VtkpQz4cKf8mQg=",
    "expires_at": "2026-02-22T09:36:18.744549895Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KJ27VZRP0KQ1DMBF3R7F3MF1",
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

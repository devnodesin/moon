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
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSENaR1dXUkJRQlJFTUcwSzIzQzZDNUgiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSENaR1dXUkJRQlJFTUcwSzIzQzZDNUgiLCJleHAiOjE3NzEwMzk2NTMsIm5iZiI6MTc3MTAzNjAyMywiaWF0IjoxNzcxMDM2MDUzfQ.EeUuX_36FPb4oh-G9YNICgHm08Tq7Cp30GgJqGezgBU",
  "refresh_token": "hyTTpweINXOKltH6r5Cl7--_8VKl58Z6fE7W0fjlHls=",
  "expires_at": "2026-02-14T03:27:33.935149435Z",
  "token_type": "Bearer",
  "user": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  }
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
  "user": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
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
  "message": "user updated successfully",
  "user": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  }
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
  "message": "password updated successfully, please login again",
  "user": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  }
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
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSENaR1dXUkJRQlJFTUcwSzIzQzZDNUgiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSENaR1dXUkJRQlJFTUcwSzIzQzZDNUgiLCJleHAiOjE3NzEwMzk2NTYsIm5iZiI6MTc3MTAzNjAyNiwiaWF0IjoxNzcxMDM2MDU2fQ.PBeaXDTl-Bk46sR-7875N4D-Bdledwx_QPCHlqo3dwk",
  "refresh_token": "Yke6FxWxoqPfagJCfD13Rbb8SZz_4SMG9TuI_a61YEE=",
  "expires_at": "2026-02-14T03:27:36.386965511Z",
  "token_type": "Bearer",
  "user": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  }
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
  "message": "logged out successfully"
}
```

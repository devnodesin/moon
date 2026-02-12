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
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSDk0Q0RUU0tKNUUzRUMwRURQNktTOFIiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSDk0Q0RUU0tKNUUzRUMwRURQNktTOFIiLCJleHAiOjE3NzA5MTA0NjEsIm5iZiI6MTc3MDkwNjgzMSwiaWF0IjoxNzcwOTA2ODYxfQ.aCcrwZQhxQw9Z4LLCKl-hrRGGewUb4XCDVEJC8aHFF0",
  "refresh_token": "tUQ-qGOWpBSfIEznoedcUtqirEHBzEamDQmbleFRHMI=",
  "expires_at": "2026-02-12T15:34:21.361302468Z",
  "token_type": "Bearer",
  "user": {
    "id": "01KH94CDTSKJ5E3EC0EDP6KS8R",
    "username": "newuser",
    "email": "newuser@example.com",
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
    "id": "01KH94CDTSKJ5E3EC0EDP6KS8R",
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
  "message": "user updated successfully",
  "user": {
    "id": "01KH94CDTSKJ5E3EC0EDP6KS8R",
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
    "id": "01KH94CDTSKJ5E3EC0EDP6KS8R",
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
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSDk0Q0RUU0tKNUUzRUMwRURQNktTOFIiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSDk0Q0RUU0tKNUUzRUMwRURQNktTOFIiLCJleHAiOjE3NzA5MTA0NzAsIm5iZiI6MTc3MDkwNjg0MCwiaWF0IjoxNzcwOTA2ODcwfQ.22FFClsMEFKewEsc4EJgyHVP122mNmg8tolrKJJL8o8",
  "refresh_token": "hRLoFVwYQrwLN-o5h4Xj5ws8IItytYKSq9WVB_TQNs4=",
  "expires_at": "2026-02-12T15:34:30.113574954Z",
  "token_type": "Bearer",
  "user": {
    "id": "01KH94CDTSKJ5E3EC0EDP6KS8R",
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

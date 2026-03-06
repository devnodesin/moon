`POST /auth:session` request shape:

```json
{
  "op": "login | refresh | logout",
  "data": {
    "username": "newuser",
    "password": "secret",
    "refresh_token": "..."
  }
}
```

The required fields in the `data` object depend on the value of `op`:

- For `login`: `username`, `password`
- For `refresh`: `refresh_token`
- For `logout`: `refresh_token`

### Login `POST /auth:session`

Request

```json
{
  "op": "login",
  "data": {
    "username": "newuser",
    "password": "UserPass123#"
  }
}
```

Response (200 OK):

```json
{
  "data": [
    {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSkhDV05ESjNRTjJaM0NSM1k5SDM2QTYiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSkhDV05ESjNRTjJaM0NSM1k5SDM2QTYiLCJleHAiOjE3NzIyNjE1NTgsIm5iZiI6MTc3MjI1NzkyOCwiaWF0IjoxNzcyMjU3OTU4fQ.lZ8oFckKcKAKLkWAAQ-CibKrNCKN55cUrDr1zbxadAI",
      "refresh_token": "SEb54NKdpecktQN0s2qjSziWlhdWM8r-Ts6TzQ-jOT4=",
      "expires_at": "2026-02-28T06:52:38.69599201Z",
      "token_type": "Bearer",
      "user": {
        "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
        "username": "newuser",
        "email": "newuser@example.com", // email is for contact only
        "role": "user",
        "can_write": true
      }
    }
  ],
  "message": "Login successful"
}
```

### Refresh `POST /auth:session`

Request

```json
{
  "op": "refresh",
  "data": {
    "refresh_token": "$REFRESH_TOKEN"
  }
}
```

Response (200 OK):

```json
{
  "data": [
    {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFLSkhDV05ESjNRTjJaM0NSM1k5SDM2QTYiLCJ1c2VybmFtZSI6Im5ld3VzZXIiLCJyb2xlIjoidXNlciIsImNhbl93cml0ZSI6dHJ1ZSwic3ViIjoiMDFLSkhDV05ESjNRTjJaM0NSM1k5SDM2QTYiLCJleHAiOjE3NzIyNjE1NjAsIm5iZiI6MTc3MjI1NzkzMCwiaWF0IjoxNzcyMjU3OTYwfQ.b3miIPvXZGt-7-58mayTA3Zy79q53S1MOnx0beT59mg",
      "refresh_token": "aDSM1M5z61WgwHfEHcgTZxqhMgjC0PbrCtg1iaKU7bw=",
      "expires_at": "2026-02-28T06:52:40.914567576Z",
      "token_type": "Bearer",
      "user": {
        "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
        "username": "newuser",
        "email": "newemail@example.com", // email is for contact only
        "role": "user",
        "can_write": true
      }
    }
  ],
  "message": "Token refreshed successfully"
}
```

### Logout `POST /auth:session`

Request

```json
{
  "op": "logout",
  "data": {
    "refresh_token": "$REFRESH_TOKEN"
  }
}
```

Response (200 OK):

```json
{
  "message": "Logged out successfully"
}
```

### Get Current User `GET  /auth:me`

Request `GET  /auth:me`

Response (200 OK):

```json
{
  "data": {
    "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
    "username": "newuser",
    "email": "newuser@example.com", // email is for contact only
    "role": "user",
    "can_write": true
  }
}
```

### Update Current User `POST /auth:me`

**Change email for current user**

Request

```json
{
  "op": "logout",
  "data": {
    "email": "newemail@example.com"
  }
}
```

Response (200 OK):

```json
{
  "data": {
    "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
    "username": "newuser",
    "email": "newemail@example.com", // email is for contact only
    "role": "user",
    "can_write": true
  },
  "message": "User updated successfully"
}
```

**Update Current User (Change Password)**

Request

```json
{
  "op": "logout",
  "data": {
    "old_password": "UserPass123#",
    "password": "NewSecurePass456"
  }
}
```

Response (200 OK):

```json
{
  "data": {
    "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
    "username": "newuser",
    "email": "newemail@example.com", // email is for contact only
    "role": "user",
    "can_write": true
  },
  "message": "Password updated successfully. Please login again."
}
```

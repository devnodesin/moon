### Create New User

```bash
curl -s -X POST "http://localhost:6006/users:create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": {
          "username": "moonuser",
          "email": "moonuser@example.com",
          "password": "UserPass123#",
          "role": "user"
        }
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "data": {
    "id": "01KHST8F8JRVTHJ6ZHBYWFGB22",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-19T02:04:29Z",
    "updated_at": "2026-02-19T02:04:29Z"
  },
  "message": "User created successfully"
}
```

### List All Users

```bash
curl -s -X GET "http://localhost:6006/users:list" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KHST646GJ82W429E5VN0WC4S",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-19T02:03:12Z",
      "updated_at": "2026-02-19T02:04:28Z",
      "last_login_at": "2026-02-19T02:04:28Z"
    },
    {
      "id": "01KHST8A7WR4CZMEM7K48KFE7Q",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-19T02:04:24Z",
      "updated_at": "2026-02-19T02:04:27Z",
      "last_login_at": "2026-02-19T02:04:27Z"
    },
    {
      "id": "01KHST8F8JRVTHJ6ZHBYWFGB22",
      "username": "moonuser",
      "email": "moonuser@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-19T02:04:29Z",
      "updated_at": "2026-02-19T02:04:29Z"
    }
  ],
  "meta": {
    "count": 3,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Get Specific User by ID

```bash
curl -s -X GET "http://localhost:6006/users:get?id=01KHST8F8JRVTHJ6ZHBYWFGB22" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHST8F8JRVTHJ6ZHBYWFGB22",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-19T02:04:29Z",
    "updated_at": "2026-02-19T02:04:29Z"
  }
}
```

### Update User

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHST8F8JRVTHJ6ZHBYWFGB22" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "email": "updateduser@example.com",
        "role": "admin"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHST8F8JRVTHJ6ZHBYWFGB22",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-19T02:04:29Z",
    "updated_at": "2026-02-19T02:04:29Z"
  },
  "message": "User updated successfully"
}
```

### Reset User Password

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHST8F8JRVTHJ6ZHBYWFGB22" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "action": "reset_password",
        "new_password": "NewSecurePassword123#"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHST8F8JRVTHJ6ZHBYWFGB22",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-19T02:04:29Z",
    "updated_at": "2026-02-19T02:04:30Z"
  },
  "message": "Password reset successfully"
}
```

### Revoke All User Sessions

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHST8F8JRVTHJ6ZHBYWFGB22" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "action": "revoke_sessions"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHST8F8JRVTHJ6ZHBYWFGB22",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-19T02:04:29Z",
    "updated_at": "2026-02-19T02:04:30Z"
  },
  "message": "All sessions revoked successfully"
}
```

### Delete User Account

```bash
curl -s -X POST "http://localhost:6006/users:destroy?id=01KHST8F8JRVTHJ6ZHBYWFGB22" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "User deleted successfully"
}
```

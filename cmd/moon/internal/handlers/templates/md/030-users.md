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
    "id": "01KJHCWTYTG4ZFFFHAWGJ26XG2",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-28T05:52:42Z",
    "updated_at": "2026-02-28T05:52:42Z"
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
      "id": "01KJHCW4P59AQVEPY55668SDBY",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-28T05:52:20Z",
      "updated_at": "2026-02-28T05:52:42Z",
      "last_login_at": "2026-02-28T05:52:42Z"
    },
    {
      "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-28T05:52:37Z",
      "updated_at": "2026-02-28T05:52:40Z",
      "last_login_at": "2026-02-28T05:52:40Z"
    },
    {
      "id": "01KJHCWTYTG4ZFFFHAWGJ26XG2",
      "username": "moonuser",
      "email": "moonuser@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-28T05:52:42Z",
      "updated_at": "2026-02-28T05:52:42Z"
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
curl -s -X GET "http://localhost:6006/users:get?id=01KJHCWTYTG4ZFFFHAWGJ26XG2" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJHCWTYTG4ZFFFHAWGJ26XG2",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-28T05:52:42Z",
    "updated_at": "2026-02-28T05:52:42Z"
  }
}
```

### Update User

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJHCWTYTG4ZFFFHAWGJ26XG2" \
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
    "id": "01KJHCWTYTG4ZFFFHAWGJ26XG2",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-28T05:52:42Z",
    "updated_at": "2026-02-28T05:52:43Z"
  },
  "message": "User updated successfully"
}
```

### Reset User Password

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJHCWTYTG4ZFFFHAWGJ26XG2" \
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
    "id": "01KJHCWTYTG4ZFFFHAWGJ26XG2",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-28T05:52:42Z",
    "updated_at": "2026-02-28T05:52:44Z"
  },
  "message": "Password reset successfully"
}
```

### Revoke All User Sessions

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJHCWTYTG4ZFFFHAWGJ26XG2" \
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
    "id": "01KJHCWTYTG4ZFFFHAWGJ26XG2",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-28T05:52:42Z",
    "updated_at": "2026-02-28T05:52:44Z"
  },
  "message": "All sessions revoked successfully"
}
```

### Delete User Account

```bash
curl -s -X POST "http://localhost:6006/users:destroy?id=01KJHCWTYTG4ZFFFHAWGJ26XG2" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "User deleted successfully"
}
```

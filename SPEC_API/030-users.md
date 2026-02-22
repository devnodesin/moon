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
    "id": "01KJ27W44GKNM4D3TR8C29KVSY",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-22T08:36:20Z",
    "updated_at": "2026-02-22T08:36:20Z"
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
      "id": "01KJ27VD9QPE6PVH7696HTWC82",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-22T08:35:56Z",
      "updated_at": "2026-02-22T08:36:19Z",
      "last_login_at": "2026-02-22T08:36:19Z"
    },
    {
      "id": "01KJ27VZRP0KQ1DMBF3R7F3MF1",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-22T08:36:15Z",
      "updated_at": "2026-02-22T08:36:18Z",
      "last_login_at": "2026-02-22T08:36:18Z"
    },
    {
      "id": "01KJ27W44GKNM4D3TR8C29KVSY",
      "username": "moonuser",
      "email": "moonuser@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-22T08:36:20Z",
      "updated_at": "2026-02-22T08:36:20Z"
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
curl -s -X GET "http://localhost:6006/users:get?id=01KJ27W44GKNM4D3TR8C29KVSY" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJ27W44GKNM4D3TR8C29KVSY",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-22T08:36:20Z",
    "updated_at": "2026-02-22T08:36:20Z"
  }
}
```

### Update User

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJ27W44GKNM4D3TR8C29KVSY" \
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
    "id": "01KJ27W44GKNM4D3TR8C29KVSY",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-22T08:36:20Z",
    "updated_at": "2026-02-22T08:36:20Z"
  },
  "message": "User updated successfully"
}
```

### Reset User Password

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJ27W44GKNM4D3TR8C29KVSY" \
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
    "id": "01KJ27W44GKNM4D3TR8C29KVSY",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-22T08:36:20Z",
    "updated_at": "2026-02-22T08:36:21Z"
  },
  "message": "Password reset successfully"
}
```

### Revoke All User Sessions

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJ27W44GKNM4D3TR8C29KVSY" \
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
    "id": "01KJ27W44GKNM4D3TR8C29KVSY",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-22T08:36:20Z",
    "updated_at": "2026-02-22T08:36:21Z"
  },
  "message": "All sessions revoked successfully"
}
```

### Delete User Account

```bash
curl -s -X POST "http://localhost:6006/users:destroy?id=01KJ27W44GKNM4D3TR8C29KVSY" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "User deleted successfully"
}
```

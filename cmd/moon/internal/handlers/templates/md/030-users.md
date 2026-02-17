### Create New User

```bash
curl -s -X POST "http://localhost:6006/users:create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "username": "moonuser",
        "email": "moonuser@example.com",
        "password": "UserPass123#",
        "role": "user"
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "data": {
    "id": "01KHNM1FNQ4ERBPNZNNFCXVTTC",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-17T10:58:51Z",
    "updated_at": "2026-02-17T10:58:51Z"
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
      "id": "01KHFW9ME0PV6P17AYG2BGVNS2",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-15T05:27:40Z",
      "updated_at": "2026-02-17T10:58:50Z",
      "last_login_at": "2026-02-17T10:58:50Z"
    },
    {
      "id": "01KHNGH2MTS2B6BRF6VEWD30ZF",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-17T09:57:27Z",
      "updated_at": "2026-02-17T09:57:30Z",
      "last_login_at": "2026-02-17T09:57:30Z"
    },
    {
      "id": "01KHNM1FNQ4ERBPNZNNFCXVTTC",
      "username": "moonuser",
      "email": "moonuser@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-17T10:58:51Z",
      "updated_at": "2026-02-17T10:58:51Z"
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
curl -s -X GET "http://localhost:6006/users:get?id=01KHNM1FNQ4ERBPNZNNFCXVTTC" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHNM1FNQ4ERBPNZNNFCXVTTC",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-17T10:58:51Z",
    "updated_at": "2026-02-17T10:58:51Z"
  }
}
```

### Update User

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHNM1FNQ4ERBPNZNNFCXVTTC" \
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
    "id": "01KHNM1FNQ4ERBPNZNNFCXVTTC",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-17T10:58:51Z",
    "updated_at": "2026-02-17T10:58:51Z"
  },
  "message": "User updated successfully"
}
```

### Reset User Password

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHNM1FNQ4ERBPNZNNFCXVTTC" \
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
    "id": "01KHNM1FNQ4ERBPNZNNFCXVTTC",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-17T10:58:51Z",
    "updated_at": "2026-02-17T10:58:52Z"
  },
  "message": "Password reset successfully"
}
```

### Revoke All User Sessions

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHNM1FNQ4ERBPNZNNFCXVTTC" \
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
    "id": "01KHNM1FNQ4ERBPNZNNFCXVTTC",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-17T10:58:51Z",
    "updated_at": "2026-02-17T10:58:52Z"
  },
  "message": "All sessions revoked successfully"
}
```

### Delete User Account

```bash
curl -s -X POST "http://localhost:6006/users:destroy?id=01KHNM1FNQ4ERBPNZNNFCXVTTC" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "User deleted successfully"
}
```

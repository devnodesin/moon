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
    "id": "01KJMQ3JK54PVC7S41QGNYPNKP",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-03-01T12:48:52Z",
    "updated_at": "2026-03-01T12:48:52Z"
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
      "id": "01KJHQ9H3T8V9D7ZT8M0EXYNDS",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-28T08:54:24Z",
      "updated_at": "2026-03-01T12:48:51Z",
      "last_login_at": "2026-03-01T12:48:51Z"
    },
    {
      "id": "01KJJKBK89SAKJ4NV49NJV769K",
      "username": "Wow",
      "email": "W@wow.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-28T17:04:52Z",
      "updated_at": "2026-02-28T17:04:52Z"
    },
    {
      "id": "01KJMG4GYD01KNTBYA7ANQ4DK2",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-03-01T10:47:03Z",
      "updated_at": "2026-03-01T10:47:10Z",
      "last_login_at": "2026-03-01T10:47:10Z"
    },
    {
      "id": "01KJMQ3JK54PVC7S41QGNYPNKP",
      "username": "moonuser",
      "email": "moonuser@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-03-01T12:48:52Z",
      "updated_at": "2026-03-01T12:48:52Z"
    }
  ],
  "meta": {
    "count": 4,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Get Specific User by ID

```bash
curl -s -X GET "http://localhost:6006/users:get?id=01KJMQ3JK54PVC7S41QGNYPNKP" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJMQ3JK54PVC7S41QGNYPNKP",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-03-01T12:48:52Z",
    "updated_at": "2026-03-01T12:48:52Z"
  }
}
```

### Update User

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJMQ3JK54PVC7S41QGNYPNKP" \
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
    "id": "01KJMQ3JK54PVC7S41QGNYPNKP",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-03-01T12:48:52Z",
    "updated_at": "2026-03-01T12:48:53Z"
  },
  "message": "User updated successfully"
}
```

### Reset User Password

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJMQ3JK54PVC7S41QGNYPNKP" \
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
    "id": "01KJMQ3JK54PVC7S41QGNYPNKP",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-03-01T12:48:52Z",
    "updated_at": "2026-03-01T12:48:54Z"
  },
  "message": "Password reset successfully"
}
```

### Revoke User Sessions

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KJMQ3JK54PVC7S41QGNYPNKP" \
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
    "id": "01KJMQ3JK54PVC7S41QGNYPNKP",
    "username": "moonuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-03-01T12:48:52Z",
    "updated_at": "2026-03-01T12:48:54Z"
  },
  "message": "All sessions revoked successfully"
}
```

### List Users with Role Filter (admin)

```bash
curl -s -X GET "http://localhost:6006/users:list?role=admin" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KJHQ9H3T8V9D7ZT8M0EXYNDS",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-28T08:54:24Z",
      "updated_at": "2026-03-01T12:48:51Z",
      "last_login_at": "2026-03-01T12:48:51Z"
    },
    {
      "id": "01KJMQ3JK54PVC7S41QGNYPNKP",
      "username": "moonuser",
      "email": "updateduser@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-03-01T12:48:52Z",
      "updated_at": "2026-03-01T12:48:54Z"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 15,
    "next": null,
    "prev": null
  }
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
      "id": "01KJHQ9H3T8V9D7ZT8M0EXYNDS",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-28T08:54:24Z",
      "updated_at": "2026-03-01T12:48:51Z",
      "last_login_at": "2026-03-01T12:48:51Z"
    },
    {
      "id": "01KJJKBK89SAKJ4NV49NJV769K",
      "username": "Wow",
      "email": "W@wow.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-28T17:04:52Z",
      "updated_at": "2026-02-28T17:04:52Z"
    },
    {
      "id": "01KJMG4GYD01KNTBYA7ANQ4DK2",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-03-01T10:47:03Z",
      "updated_at": "2026-03-01T10:47:10Z",
      "last_login_at": "2026-03-01T10:47:10Z"
    },
    {
      "id": "01KJMQ3JK54PVC7S41QGNYPNKP",
      "username": "moonuser",
      "email": "updateduser@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-03-01T12:48:52Z",
      "updated_at": "2026-03-01T12:48:54Z"
    }
  ],
  "meta": {
    "count": 4,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### List Users with After ID Pagination

```bash
curl -s -X GET "http://localhost:6006/users:list?limit=1" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KJHQ9H3T8V9D7ZT8M0EXYNDS",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-28T08:54:24Z",
      "updated_at": "2026-03-01T12:48:51Z",
      "last_login_at": "2026-03-01T12:48:51Z"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 1,
    "next": "01KJHQ9H3T8V9D7ZT8M0EXYNDS",
    "prev": null
  }
}
```

### List Users with After ID Pagination

```bash
curl -s -X GET "http://localhost:6006/users:list?after=01KJHQ9H3T8V9D7ZT8M0EXYNDS&limit=1" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KJJKBK89SAKJ4NV49NJV769K",
      "username": "Wow",
      "email": "W@wow.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-28T17:04:52Z",
      "updated_at": "2026-02-28T17:04:52Z"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 1,
    "next": "01KJJKBK89SAKJ4NV49NJV769K",
    "prev": null
  }
}
```

### Delete User Account

```bash
curl -s -X POST "http://localhost:6006/users:destroy?id=01KJMQ3JK54PVC7S41QGNYPNKP" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "User deleted successfully"
}
```

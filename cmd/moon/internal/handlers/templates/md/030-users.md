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
  "message": "user created successfully",
  "user": {
    "id": "01KHCZK95DPBAT04EH0WWDZYR7",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-14T02:27:38Z",
    "updated_at": "2026-02-14T02:27:38Z"
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
  "users": [
    {
      "id": "01KHCZFXAFJPS9SKSFKNBMHTP5",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2026-02-14T02:25:48Z",
      "updated_at": "2026-02-14T02:27:37Z",
      "last_login_at": "2026-02-14T02:27:37Z"
    },
    {
      "id": "01KHCZGWWRBQBREMG0K23C6C5H",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-14T02:26:20Z",
      "updated_at": "2026-02-14T02:27:36Z",
      "last_login_at": "2026-02-14T02:27:36Z"
    },
    {
      "id": "01KHCZK95DPBAT04EH0WWDZYR7",
      "username": "moonuser",
      "email": "moonuser@example.com",
      "role": "user",
      "can_write": true,
      "created_at": "2026-02-14T02:27:38Z",
      "updated_at": "2026-02-14T02:27:38Z"
    }
  ],
  "next_cursor": null,
  "limit": 15
}
```

### Get Specific User by ID

```bash
curl -s -X GET "http://localhost:6006/users:get?id=01KHCZGWWRBQBREMG0K23C6C5H" \
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
    "can_write": true,
    "created_at": "2026-02-14T02:26:20Z",
    "updated_at": "2026-02-14T02:27:36Z",
    "last_login_at": "2026-02-14T02:27:36Z"
  }
}
```

### Update User

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHCZGWWRBQBREMG0K23C6C5H" \
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
  "message": "user updated successfully",
  "user": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-14T02:26:20Z",
    "updated_at": "2026-02-14T02:27:39Z",
    "last_login_at": "2026-02-14T02:27:36Z"
  }
}
```

### Reset User Password

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHCZGWWRBQBREMG0K23C6C5H" \
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
  "message": "password reset successfully",
  "user": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-14T02:26:20Z",
    "updated_at": "2026-02-14T02:27:40Z",
    "last_login_at": "2026-02-14T02:27:36Z"
  }
}
```

### Revoke All User Sessions

```bash
curl -s -X POST "http://localhost:6006/users:update?id=01KHCZGWWRBQBREMG0K23C6C5H" \
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
  "message": "all sessions revoked successfully",
  "user": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "can_write": true,
    "created_at": "2026-02-14T02:26:20Z",
    "updated_at": "2026-02-14T02:27:40Z",
    "last_login_at": "2026-02-14T02:27:36Z"
  }
}
```

### Delete User Account

```bash
curl -s -X POST "http://localhost:6006/users:destroy?id=01KHCZGWWRBQBREMG0K23C6C5H" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "user deleted successfully"
}
```

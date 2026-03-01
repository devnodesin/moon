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

**Query Parameters:**

| Parameter | Type   | Description                                         |
|-----------|--------|-----------------------------------------------------|
| `limit`   | int    | Maximum users per page (default: 15, max: 100)      |
| `after`   | string | Cursor (ULID) for forward pagination                |
| `role`    | string | Filter by role: `admin` or `user`                   |

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
    "prev": null,
    "total": 3
  }
}
```

**Meta fields:**

| Field   | Type        | Description                                                        |
|---------|-------------|--------------------------------------------------------------------|
| `count` | int         | Number of users returned in this page                              |
| `limit` | int         | Page size limit used for this request                              |
| `next`  | string/null | Cursor of the last user on this page; use as `?after=<next>` to fetch the next page. `null` if there are no more pages. |
| `prev`  | string/null | Cursor to fetch the previous page using `?after=<prev>`. `null` on page 1 (no previous page) and page 2 (page 1 requires no cursor; navigate there by removing the `?after` parameter entirely). |
| `total` | int         | Total number of users matching the filter (ignores pagination cursor) |

### List Users with Pagination

Fetch the first page (limit=2):

```bash
curl -s -X GET "http://localhost:6006/users:list?limit=2" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    { "id": "01KJHCW4P59AQVEPY55668SDBY", "username": "admin", ... },
    { "id": "01KJHCWNDJ3QN2Z3CR3Y9H36A6", "username": "newuser", ... }
  ],
  "meta": {
    "count": 2,
    "limit": 2,
    "next": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
    "prev": null,
    "total": 4
  }
}
```

Fetch the next page using the `next` cursor:

```bash
curl -s -X GET "http://localhost:6006/users:list?limit=2&after=01KJHCWNDJ3QN2Z3CR3Y9H36A6" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK) — page 2:**

```json
{
  "data": [
    { "id": "01KJHCWTYTG4ZFFFHAWGJ26XG2", "username": "moonuser", ... },
    { "id": "01KJHCWZ12345ABCDEFGHJKLMN", "username": "anotheruser", ... }
  ],
  "meta": {
    "count": 2,
    "limit": 2,
    "next": "01KJHCWZ12345ABCDEFGHJKLMN",
    "prev": null,
    "total": 4
  }
}
```

Fetch the next page (page 3), which also provides a `prev` cursor for backward navigation:

```bash
curl -s -X GET "http://localhost:6006/users:list?limit=2&after=01KJHCWZ12345ABCDEFGHJKLMN" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK) — page 3 with backward cursor:**

```json
{
  "data": [
    { "id": "01KJHCXABC123ABCDEFGHJKLMN", "username": "user5", ... }
  ],
  "meta": {
    "count": 1,
    "limit": 2,
    "next": null,
    "prev": "01KJHCWNDJ3QN2Z3CR3Y9H36A6",
    "total": 4
  }
}
```

Navigate backward by passing the `prev` cursor as `?after`:

```bash
curl -s -X GET "http://localhost:6006/users:list?limit=2&after=01KJHCWNDJ3QN2Z3CR3Y9H36A6" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

This returns page 2 (the page just before page 3).

**Pagination notes:**
- `next` is the cursor to advance forward; use `?after=<next>` to fetch the next page.
- `prev` is the cursor to navigate backward; use `?after=<prev>` to fetch the previous page. It is `null` on page 1 (no previous page) and page 2 (page 1 has no cursor representation; navigate there by removing the `?after` parameter).
- `total` reflects the total count of users matching the role filter, independent of the cursor position.


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

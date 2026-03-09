### Create User

Create a new user account.

```bash
curl -s -X POST "http://localhost:6000/data/users:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "username": "moonuser",
            "email": "moonuser@example.com",
            "password": "UserPass123#",
            "role": "user",
            "can_write": false
          }
        ]
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "message": "Resource created successfully",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-03-09T17:13:00Z",
      "email": "moonuser@example.com",
      "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J",
      "role": "user",
      "updated_at": "2026-03-09T17:13:00Z",
      "username": "moonuser"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### List All Users

Retrieve all users.

```bash
curl -s -X GET "http://localhost:6000/data/users:query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "can_write": true,
      "created_at": "2026-03-09T10:15:50Z",
      "email": "admin@example.com",
      "id": "01KK91H3WK7KH2H1H7AT4NYMA6",
      "last_login_at": "2026-03-09T17:13:00Z",
      "role": "admin",
      "updated_at": "2026-03-09T17:13:00Z",
      "username": "admin"
    },
    {
      "can_write": true,
      "created_at": "2026-03-09T15:07:48Z",
      "email": "mohamed@asensar.com",
      "id": "01KK9J7Q51N1SW86VBWHB1NKH0",
      "last_login_at": null,
      "role": "user",
      "updated_at": "2026-03-09T15:07:48Z",
      "username": "mohamed"
    },
    {
      "can_write": false,
      "created_at": "2026-03-09T17:13:00Z",
      "email": "moonuser@example.com",
      "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J",
      "last_login_at": null,
      "role": "user",
      "updated_at": "2026-03-09T17:13:00Z",
      "username": "moonuser"
    }
  ],
  "meta": {
    "count": 3,
    "current_page": 1,
    "per_page": 15,
    "total": 3,
    "total_pages": 1
  },
  "links": {
    "first": "/data/users:query?page=1&per_page=15",
    "last": "/data/users:query?page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```

### Get User by ID

Retrieve a specific user by their ULID.

```bash
curl -s -X GET "http://localhost:6000/data/users:query?id=01KK9SCZ3TMBYP9E8KV34GEQ9J" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource retrieved successfully",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-03-09T17:13:00Z",
      "email": "moonuser@example.com",
      "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J",
      "last_login_at": null,
      "role": "user",
      "updated_at": "2026-03-09T17:13:00Z",
      "username": "moonuser"
    }
  ]
}
```

### Update User

Update an existing user's fields: email, role and can_write permissions.

```bash
curl -s -X POST "http://localhost:6000/data/users:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "update",
        "data": [
          {
            "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J",
            "email": "moonuser_updated@example.com"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource updated successfully",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-03-09T17:13:00Z",
      "email": "moonuser_updated@example.com",
      "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J",
      "last_login_at": null,
      "role": "user",
      "updated_at": "2026-03-09T17:13:01Z",
      "username": "moonuser"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Reset User Password

Reset a user's password via action.

```bash
curl -s -X POST "http://localhost:6000/data/users:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "action",
        "action": "reset_password",
        "data": [
          {
            "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J",
            "password": "NewSecurePassword123"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Action completed successfully",
  "data": [
    {
      "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Revoke User Sessions

Revoke all sessions for a user via action.

```bash
curl -s -X POST "http://localhost:6000/data/users:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "action",
        "action": "revoke_sessions",
        "data": [
          {
            "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Action completed successfully",
  "data": [
    {
      "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Delete User

Delete a user account.

```bash
curl -s -X POST "http://localhost:6000/data/users:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "destroy",
        "data": [
          {
            "id": "01KK9SCZ3TMBYP9E8KV34GEQ9J"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource destroyed successfully",
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

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
      "created_at": "2026-03-08T16:38:42Z",
      "email": "moonuser@example.com",
      "id": "01KK751DR3JH9VH80B17H114FA",
      "role": "user",
      "updated_at": "2026-03-08T16:38:42Z",
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
      "created_at": "2026-03-08T12:33:26Z",
      "email": "admin@example.com",
      "id": "01KK6Q0AKZ52X0V10S7QG9MEVV",
      "last_login_at": "2026-03-08T16:38:41Z",
      "role": "admin",
      "updated_at": "2026-03-08T16:38:41Z",
      "username": "admin"
    },
    {
      "can_write": false,
      "created_at": "2026-03-08T16:38:42Z",
      "email": "moonuser@example.com",
      "id": "01KK751DR3JH9VH80B17H114FA",
      "last_login_at": null,
      "role": "user",
      "updated_at": "2026-03-08T16:38:42Z",
      "username": "moonuser"
    }
  ],
  "meta": {
    "count": 2,
    "current_page": 1,
    "per_page": 15,
    "total": 2,
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
curl -s -X GET "http://localhost:6000/data/users:query?id=01KK751DR3JH9VH80B17H114FA" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource retrieved successfully",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-03-08T16:38:42Z",
      "email": "moonuser@example.com",
      "id": "01KK751DR3JH9VH80B17H114FA",
      "last_login_at": null,
      "role": "user",
      "updated_at": "2026-03-08T16:38:42Z",
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
            "id": "01KK751DR3JH9VH80B17H114FA",
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
      "created_at": "2026-03-08T16:38:42Z",
      "email": "moonuser_updated@example.com",
      "id": "01KK751DR3JH9VH80B17H114FA",
      "last_login_at": null,
      "role": "user",
      "updated_at": "2026-03-08T16:38:44Z",
      "username": "moonuser"
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
            "id": "01KK751DR3JH9VH80B17H114FA"
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

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
      "created_at": "2026-11-09T04:23:32Z",
      "email": "moonuser@example.com",
      "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6",
      "role": "user",
      "updated_at": "2026-11-09T04:23:32Z",
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
      "created_at": "2026-11-09T04:23:20Z",
      "email": "admin@example.com",
      "id": "01M6Z8QQYXRFP3TA7BXRXAGQA8",
      "last_login_at": "2026-11-09T04:23:31Z",
      "login_method": "password",
      "role": "admin",
      "updated_at": "2026-11-09T04:23:31Z",
      "username": "admin"
    },
    {
      "can_write": false,
      "created_at": "2026-11-09T04:23:32Z",
      "email": "moonuser@example.com",
      "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6",
      "last_login_at": null,
      "login_method": "password",
      "role": "user",
      "updated_at": "2026-11-09T04:23:32Z",
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
curl -s -X GET "http://localhost:6000/data/users:query?id=01M6Z8R3GFH5ZC0J4SK6TRBQH6" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource retrieved successfully",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-11-09T04:23:32Z",
      "email": "moonuser@example.com",
      "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6",
      "last_login_at": null,
      "login_method": "password",
      "role": "user",
      "updated_at": "2026-11-09T04:23:32Z",
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
            "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6",
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
      "created_at": "2026-11-09T04:23:32Z",
      "email": "moonuser_updated@example.com",
      "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6",
      "last_login_at": null,
      "login_method": "password",
      "role": "user",
      "updated_at": "2026-11-09T04:23:32Z",
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
            "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6",
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
      "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6"
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
            "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6"
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
      "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6"
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
            "id": "01M6Z8R3GFH5ZC0J4SK6TRBQH6"
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

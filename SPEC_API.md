# Standard API Response Patterns

This document describes the standard response patterns, query options, and aggregation operations for the Moon API. All endpoints follow consistent conventions for success and error responses.


## Public Endpoints

Health and documentation endpoints are public; no authentication is required. All other endpoints require authentication.

### API Documentation

API documentation is available in multiple formats:

- HTML: `GET /doc/`
- Markdown: `GET /doc/llms.md`
- Plain Text: `GET /doc/llms.txt` (alias for `/doc/llms.md`)
- JSON: `GET /doc/llms.json`

### Health Endpoint

- `GET /health`: Returns API service health and version information.

**Response (200 OK):**

```json
{
  "data": {
    "moon": "1.0.0",
    "status": "ok",    
    "timestamp": "2026-02-03T13:58:53Z"
  }
}
```

### Important Notes

**Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Authentication Endpoints

- `POST /auth:login`: Login
- `POST /auth:logout`: Logout
- `POST /auth:refresh`: Refresh access token
- `GET /auth:me`: Get current user
- `POST /auth:me`: Update current user

### Login

`POST /auth:login`

Authenticate user and receive access token.

**Request body:**

```json
{
  "username": "newuser",
  "password": "UserPass123#"
}
```

**Response (200 OK):**

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "hyTTpweINXOKltH6r5Cl7--_8VKl58Z6fE7W0fjlHls=",
    "expires_at": "2026-02-14T03:27:33.935149435Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KHCZGWWRBQBREMG0K23C6C5H",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true
    }
  },
  "message": "Login successful"
}
```

### Get Current User

`GET /auth:me`

Retrieve authenticated user information.

**Headers:**

- `Authorization: Bearer {access_token}` (required)

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  }
}
```

### Update Current User

`POST /auth:me`

Update authenticated user's email or password.

**Headers:**

- `Authorization: Bearer {access_token}` (required)

**Update email:**

```json
{
  "email": "newemail@example.com"
}
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  },
  "message": "User updated successfully"
}
```

**Change password:**

```json
{
  "old_password": "UserPass123#",
  "password": "NewSecurePass456"
}
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  },
  "message": "Password updated successfully. Please login again."
}
```

### Refresh Token

`POST /auth:refresh`

Generate new access token using refresh token.

**Request body:**

```json
{
  "refresh_token": "hyTTpweINXOKltH6r5Cl7--_8VKl58Z6fE7W0fjlHls="
}
```

**Response (200 OK):**

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "Yke6FxWxoqPfagJCfD13Rbb8SZz_4SMG9TuI_a61YEE=",
    "expires_at": "2026-02-14T03:27:36.386965511Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KHCZGWWRBQBREMG0K23C6C5H",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true
    }
  },
  "message": "Token refreshed successfully"
}
```

### Logout

`POST /auth:logout`

Invalidate current session and refresh token.

**Headers:**

- `Authorization: Bearer {access_token}` (required)

**Request body:**

```json
{
  "refresh_token": "hyTTpweINXOKltH6r5Cl7--_8VKl58Z6fE7W0fjlHls="
}
```

**Response (200 OK):**

```json
{
  "message": "Logged out successfully"
}
```

### Important Notes

- **Token expiration**: Access tokens expire in 1 hour (configurable). Use refresh token to obtain new access token without re-authentication.
- **Refresh token**: Single-use tokens. Each refresh returns a new access token AND a new refresh token. Store the new refresh token for subsequent refreshes.
- **Password change**: Changing password invalidates all existing sessions. User must login again with new credentials.
- **Authorization header**: Format is `Authorization: Bearer {access_token}`. Include this header in all authenticated requests.
- **Token storage**: Store tokens securely. Never expose tokens in URLs or logs.
- **Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Standard Response Pattern for `:list` Endpoints

List endpoints return paginated collections of resources. All list endpoints share a consistent request/response pattern described in this section.

**Applicable Endpoints:**

- List Users: `GET /users:list`
- List API Keys: `GET /apikeys:list`
- List Collections: `GET /collections:list`
- List Collection Records: `GET /{collection_name}:list`

**Response Structure:**
Every list endpoint returns a JSON object with two top-level keys: `data` and `meta`.

```json
{
  "data": [
    {
      "id": "01KHCZKMM0N808MKSHBNWF464F",
      "title": "Wireless Mouse",
      "price": "29.99"
    }
  ],
  "meta": {
    "count": 15,
    "limit": 15,
    "next": "01KHCZKMM0N808MKSHBNWF464F",
    "prev": "01KHCZFXAFJPS9SKSFKNBMHTP5"
  }
}
```

- `data` - An array of resource objects. Each record always includes an `id` field (ULID), except collections which use `name` as the identifier.
- `meta` - Pagination metadata for the current page.
  - `count` (integer): Number of records returned in this response
  - `limit` (integer): The page size limit that was applied. Default is 15; maximum allowed is 100.
  - `next` (string | null): Cursor pointing to the last record on the current page. Pass to ?after to get the next page. null on the last page.
  - `prev` (string | null): Cursor pointing to the record before the current page. Pass to ?after to return to the previous page. null on the first page.

---

The `:list` endpoint supports the following query parameters: `limit`, `after`, `sort`, `filter`, `q` (full-text search), and `fields` (field selection).

### Pagination

For pagination use parameter `?after={cursor}` to return records after the specified ULID cursor. Omit this parameter to start from the first page.

This API uses cursor-based pagination. Each response includes `meta.next` and `meta.prev` cursors, both of which are used with the `?after` parameter.

```sh
# First page (no cursor needed)
GET /products:list

# Next page — use meta.next, meta.prev, or any valid record id from the previous response
GET /products:list?after=01KHCZKMM0N808MKSHBNWF464F
```

**Notes:**

- `meta.prev` is `null` on the first page and `meta.next` is `null` on the last page.
- Records are always returned in chronological order (by ULID/creation time).
 - To page backwards: pass `?after={meta.prev}` from the current response. This returns the previous page of records (the record matching the cursor is excluded). Example: `GET /products:list?after=01KHCZFXAFJPS9SKSFKNBMHTP5`.
- For `?after={cursor}`, the cursor must always be a record's id (ULID). It can be:
  - A valid id of an existing record,
  - The value of `meta.prev` from the current response,
  - The value of `meta.next` from the current response.
- When `?after={cursor}` is used, only records that follow the specified id (ULID) are returned; the record matching the cursor is excluded from the results.
- If an invalid or non-existent cursor is provided, return an error response as specified in the [Standard Error Response](#standard-error-response) section.

### Limit

Use the query option `?limit={number}` to set the number of records returned per page. The default is 15; the maximum is 100.

```sh
GET /products:list?limit=2
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KHCZKSPHB01TBEWKYQDKG5KS",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 2,
    "next": "01KHCZKSPHB01TBEWKYQDKG5KS",
    "prev": null
  }
}
```

### Filtering

Filter results by column value using the syntax `?{column_name}[operator]=value`. You can combine multiple filters in a single request.

**Supported operators:**

- `eq`: Equal to
- `ne`: Not equal to
- `gt`: Greater than
- `lt`: Less than
- `gte`: Greater than or equal to
- `lte`: Less than or equal to
- `like`: Pattern match. Use `%` as a wildcard, e.g. `brand[like]=Wo%`
- `in`: Matches any value in a comma-separated list, e.g. `brand[in]=Wow,Orange`

**Example:**

```sh
GET /products:list?quantity[gt]=5&brand[eq]=Wow
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KHCZKT086EEB3EKM3PZ3N2Q0",
      "price": "199.99",
      "quantity": 20,
      "title": "Monitor 21 inch"
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

### Sorting

Use `?sort={field1,-field2,...}` to sort by one or more fields. Prefix a field name with `-` for descending order. Separate multiple fields with commas.

```sh
GET /products:list?sort=-quantity,title
```

Above sorts by `quantity` descending, then by `title` ascending.

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Orange",
      "details": "Gaming keyboard",
      "id": "01KHCZKSPHB01TBEWKYQDKG5KS",
      "price": "19.99",
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "brand": "Wow",
      "details": "Full HD monitor",
      "id": "01KHCZKT086EEB3EKM3PZ3N2Q0",
      "price": "199.99",
      "quantity": 20,
      "title": "Monitor 21 inch"
    },
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
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

### Full-Text Search

Use `?q` to search across all string and text fields in the collection.

```sh
GET /products:list?q=mouse
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "brand": "Wow",
      "details": "Ergonomic wireless mouse",
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
      "price": "29.99",
      "quantity": 10,
      "title": "Wireless Mouse"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Field Selection

Return only the fields you need. `id` is always included.

```sh
GET /products:list?fields=quantity,title
```

**Response (200 OK):**

```json
{
  "data": [
    { "id": "01KHCZKSBQV1KH69AA6PVS12MM", "quantity": 10, "title": "Wireless Mouse" },
    { "id": "01KHCZKSPHB01TBEWKYQDKG5KS", "quantity": 55, "title": "USB Keyboard" },
    { "id": "01KHCZKT086EEB3EKM3PZ3N2Q0", "quantity": 20, "title": "Monitor 21 inch" }
  ],
  "meta": {
    "count": 3,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Combined Examples

All query parameters can be combined in a single request.

```sh
# Filter by price range, sort descending, limit results
GET /products:list?quantity[gte]=10&price[lt]=100&sort=-price&limit=5

# Full-text search with a brand filter, returning only select fields
GET /products:list?q=laptop&brand[eq]=Wow&fields=title,price,quantity

# Multi-filter with pagination
GET /products:list?price[gte]=100&quantity[gt]=0&sort=-price&limit=10&after=01KHCZKMM0N808MKSHBNWF464F
```

### Important Notes

- **Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Standard Response Pattern for `:get` Endpoints

Get endpoints retrieve a single resource by its identifier.

**Applicable Endpoints:**

- Get User: `GET /users:get?id={id}`
- Get API Key: `GET /apikeys:get?id={id}`
- Get Collection: `GET /collections:get?name={collection_name}`
- Get Collection Record: `GET /{collection_name}:get?id={id}`

### Response Structure

**For Users, API Keys, and Collection Records:**

```sh
GET /users:get?id=01KHCZGWWRBQBREMG0K23C6C5H
GET /apikeys:get?id=01KHCZGWWRBQBREMG0K23C6C5H
GET /products:get?id=01KHCZGWWRBQBREMG0K23C6C5H
```

**Example Response:**

```json
{
  "data": {
    "id": "01KHCZKMM0N808MKSHBNWF464F",
    "title": "Wireless Mouse",
    "price": "29.99"
  }
}
```

**For Collections:**

```sh
GET /collections:get?name=products
```

**Example Response:**

```json
{
  "data": {
    "name": "products",
    "columns": [
      {
        "name": "title",
        "type": "string",
        "nullable": false,
        "unique": true
      }
    ]
  }
}
```

### Parameters

- `id` (string): ULID of the resource (required for users, API keys, and records)
- `name` (string): Name of the collection (required for collections)

### Important Notes

- **Single object**: The `data` field contains a single object (not an array).
- **No meta field**: Get endpoints don't need pagination metadata.
- **Consistent wrapper**: All `:get` endpoints use the `data` wrapper, matching `:list` endpoints.
- **Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Standard Response Pattern for `:create` Endpoints

Create endpoints add new resources to the system.

**Applicable Endpoints:**

- Create User: `POST /users:create`
- Create API Key: `POST /apikeys:create`
- Create Collection: `POST /collections:create`
- Create Collection Record(s): `POST /{collection_name}:create`

### Response Structure

**For Users, API Keys, and Collections:**

```sh
POST /users:create
POST /apikeys:create
POST /collections:create
```

**Request Body:**

```json
{
  "data": {
    "username": "moonuser",
    "email": "moonuser@example.com",
    "password": "UserPass123#"
  }
}
```

**Response (201 Created):**

```json
{
  "data": {
    "id": "01KHCZK95DPBAT04EH0WWDZYR7",
    "username": "moonuser",
    "email": "moonuser@example.com",
    "created_at": "2026-02-14T02:27:38Z"
  },
  "message": "User created successfully"
}
```

**For Collection Records:**

Single record:

```json
{
  "data": [{ "title": "Wireless Mouse", "price": "29.99" }]
}
```

Multiple records:

```json
{
  "data": [
    { "title": "Keyboard", "price": "49.99" },
    { "title": "Monitor", "price": "199.99" },
    { "title": "Keyboard", "price": "39.99" }
  ]
}
```

**Response (201 Created) - All Succeeded:**

```json
{
  "data": [
    {
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "title": "Keyboard",
      "price": "49.99"
    },
    {
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SY",
      "title": "Monitor",
      "price": "199.99"
    },
    {
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SZ",
      "title": "Keyboard",
      "price": "39.99"
    }
  ],
  "meta": {
    "total": 3,
    "succeeded": 3,
    "failed": 0
  },
  "message": "3 record(s) created successfully"
}
```

**Response (201 Created) - Partial Success:**

```json
{
  "data": [
    {
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "title": "Keyboard",
      "price": "49.99"
    },
    {
      "id": "01KHCZKMY28ERJFPCVBQEKQ4SY",
      "title": "Monitor",
      "price": "199.99"
    }
  ],
  "meta": {
    "total": 3,
    "succeeded": 2,
    "failed": 1
  },
  "message": "2 of 3 record(s) created successfully"
}
```

**For Collection Creation:**

**Request Body:**

```json
{
  "data": {
    "name": "products",
    "columns": [
      {
        "name": "title",
        "type": "string",
        "nullable": false,
        "unique": true
      },
      {
        "name": "price",
        "type": "integer",
        "nullable": false
      }
    ]
  }
}
```

**Response (201 Created):**

```json
{
  "data": {
    "name": "products",
    "columns": [
      {
        "name": "title",
        "type": "string",
        "nullable": false,
        "unique": true
      },
      {
        "name": "price",
        "type": "integer",
        "nullable": false,
        "unique": false
      }
    ]
  },
  "message": "Collection 'products' created successfully"
}
```

### Important Notes

- **ID field**: The `id` field is system-generated and read-only. Do not include it in create requests.
- **Array format**: Collection records must always be sent as an array in `data`, even for single records.
- **Partial success**: If some records fail validation, successfully created records are returned in `data`.
- **Failed records**: Failed records are excluded from the `data` array. Check `meta.failed` count to detect partial failures.
- **Status code**: Always returns `201 Created` if at least one record was created successfully.
- **Consistent wrapper**: All `:create` endpoints use the `data` field for created resource(s).
- **Message field**: Always includes a human-readable success message.
- **API Key security**: The `key` field appears in `data` only once during creation.
- **Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Standard Response Pattern for `:destroy` Endpoints

Destroy endpoints permanently delete resources from the system.

**Applicable Endpoints:**

- Delete User: `POST /users:destroy?id={id}`
- Delete API Key: `POST /apikeys:destroy?id={id}`
- Delete Collection: `POST /collections:destroy?name={collection_name}`
- Delete Collection Record(s): `POST /{collection_name}:destroy`

### Response Structure

**For Users, API Keys, and Collections:**

```sh
POST /users:destroy?id=01KHCZGWWRBQBREMG0K23C6C5H
POST /apikeys:destroy?id=01KHCZKCR7MHB0Q69KM63D6AXF
POST /collections:destroy?name=products
```

**Response (200 OK):**

```json
{
  "message": "User deleted successfully"
}
```

**For Collection Records:**

Single record:

```json
{
  "data": ["01KHCZKMM0N808MKSHBNWF464F"]
}
```

Multiple records:

```json
{
  "data": [
    "01KHCZKMXYVC1NRHDZ83XMHY4N",
    "01KHCZKMY28ERJFPCVBQEKQ4SY",
    "01KHCZKMY28ERJFPCVBQEKQ4SZ"
  ]
}
```

**Response (200 OK) - All Succeeded:**

```json
{
  "data": [
    "01KHCZKMXYVC1NRHDZ83XMHY4N",
    "01KHCZKMY28ERJFPCVBQEKQ4SY",
    "01KHCZKMY28ERJFPCVBQEKQ4SZ"
  ],
  "meta": {
    "total": 3,
    "succeeded": 3,
    "failed": 0
  },
  "message": "3 record(s) deleted successfully"
}
```

**Response (200 OK) - Partial Success:**

```json
{
  "data": ["01KHCZKMXYVC1NRHDZ83XMHY4N", "01KHCZKMY28ERJFPCVBQEKQ4SY"],
  "meta": {
    "total": 3,
    "succeeded": 2,
    "failed": 1
  },
  "message": "2 of 3 record(s) deleted successfully"
}
```

### Parameters

- `id` (string): ULID of the resource (required for users, apikeys)
- `name` (string): Name of the collection (required for collections)
- `data` (array): Array of record IDs to delete (required for records)

### Important Notes

- **Array format**: Collection records must be sent as an array in `data`, even for single deletions.
- **Deleted IDs returned**: Response includes `data` array with IDs of successfully deleted records.
- **Partial success**: If some records fail to delete, the successfully deleted count is shown in `meta`.
- **Failed records**: Check `meta.failed` count to detect partial failures. Failed record IDs are excluded from the `data` array.
- **Status code**: Returns `200 OK` if at least one record was deleted successfully.
- **Message field**: Always includes a human-readable success message.
- **Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Standard Response Pattern for `:update` Endpoints

Update endpoints modify existing resources in the system.

**Applicable Endpoints:**

- Update User: `POST /users:update?id={id}`
- Update API Key: `POST /apikeys:update?id={id}`
- Update Collection: `POST /collections:update`
- Update Collection Record(s): `POST /{collection_name}:update`

### Response Structure

**For Users and API Keys:**

```sh
POST /users:update?id=01KHCZGWWRBQBREMG0K23C6C5H
POST /apikeys:update?id=01KHCZKCR7MHB0Q69KM63D6AXF
```

Standard update:

```json
{
  "email": "updateduser@example.com",
  "role": "admin"
}
```

Special actions:

```json
{
  "action": "reset_password",
  "new_password": "NewSecurePassword123#"
}
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "updateduser@example.com",
    "role": "admin",
    "updated_at": "2026-02-14T02:27:39Z"
  },
  "message": "User updated successfully"
}
```

**For Collections:**

```sh
POST /collections:update
```

**Request Body:**

```json
{
  "data": {
    "name": "products",
    "add_columns": [
      {
        "name": "stock",
        "type": "integer",
        "nullable": false
      }
    ]
  }
}
```

**Response (200 OK):**

```json
{
  "data": {
    "name": "products",
    "columns": [...]
  },
  "message": "Collection 'products' updated successfully"
}
```

**For Collection Records:**

Single record:

```json
{
  "data": [
    {
      "id": "01KHCZKMM0N808MKSHBNWF464F",
      "price": "100.00",
      "title": "Updated Product"
    }
  ]
}
```

Multiple records:

```json
{
  "data": [
    {
      "id": "01KHCZKMM0N808MKSHBNWF464F",
      "price": "100.00"
    },
    {
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "price": "200.00"
    }
  ]
}
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KHCZKMM0N808MKSHBNWF464F",
      "price": "100.00",
      "title": "Updated Product"
    }
  ],
  "meta": {
    "total": 2,
    "succeeded": 2,
    "failed": 0
  },
  "message": "2 record(s) updated successfully"
}
```

### Special Actions

**User Actions:**

Reset user password:

```json
{
  "action": "reset_password",
  "new_password": "NewSecurePassword123#"
}
```

Revoke all active sessions:

```json
{
  "action": "revoke_sessions"
}
```

**API Key Actions:**

Rotate API key (generate new key and invalidate old one):

```json
{
  "action": "rotate"
}
```

Response includes new `key` field:

```json
{
  "data": {
    "id": "01KHCZKCR7MHB0Q69KM63D6AXF",
    "name": "Updated Service Name",
    "key": "moon_live_9wAtNeHqBdYmQaf3Dvm8YhVM4FK880X4dj5dqFSqLJ6qJmApxHuYe6gqkI3ipgKG"
  },
  "message": "API key rotated successfully",
  "warning": "Store this key securely. It will not be shown again."
}
```

### Parameters

- `id` (string): ULID of the resource (required for users, apikeys)
- Request Body (object): Fields to update OR `action` parameter for special operations
- `action` (string): Special operation to perform (`reset_password`, `revoke_sessions`, `rotate`)
- `name` (string): Collection name (required for collection operations)
- `data` (array): Array with objects containing `id` plus fields to update (for records)

### Important Notes

- **Array format**: Collection records must be sent as an array in `data`, even for single updates.
- **Partial updates**: Only fields provided are updated; other fields remain unchanged.
- **Actions vs updates**: When `action` is specified, it takes precedence over field updates.
- **Action-specific fields**: Some actions require additional fields (e.g., `new_password` for `reset_password`).
- **Updated data returned**: Response includes the full updated resource(s) in `data`.
- **Partial success**: For batch updates, successfully updated records are returned in `data`.
- **Status code**: Returns `200 OK` if at least one record was updated successfully.
- **Key rotation**: `rotate` action returns the new key in `data.key` field (shown only once).
- **Warning field**: Optional field for security warnings (e.g., key rotation, password reset).
- **Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Standard Response Pattern for `:schema` Endpoints

Retrieve the schema definition for a collection, including all fields, their types, constraints, and defaults.

`GET /{collection_name}:schema`

**Example Request:**

```sh
GET /products:schema
```

**Response (200 OK):**

```json
{
  "data": {
    "collection": "products",
    "fields": [
      {
        "name": "id",
        "type": "string",
        "nullable": false,
        "readonly": true
      },
      {
        "name": "title",
        "type": "string",
        "nullable": false
      },
      {
        "name": "price",
        "type": "decimal",
        "nullable": false
      },
      {
        "name": "details",
        "type": "string",
        "nullable": true,
        "default": "''"
      },
      {
        "name": "quantity",
        "type": "integer",
        "nullable": true,
        "default": "0"
      },
      {
        "name": "brand",
        "type": "string",
        "nullable": true,
        "default": "''"
      }
    ],
    "total": 6
  }
}
```

**Field Properties**

- `name`: The field's name.
- `type`: The data type (`string`, `integer`, `decimal`, `boolean`, `timestamp`, etc., as defined in the specification).
- `nullable`: Indicates if the field can be omitted or set to `null` in API requests.
- `readonly`: Indicates if the field is system-generated and cannot be modified (e.g., `id`).
- `default`: The default value assigned when the field is not provided.
- `unique`: Specifies whether the field must have unique values (optional).

### Important Notes

- **System fields**: The `id` and `created_at` fields are automatically included in every collection and are readonly.
- **Total count**: Represents the total number of fields in the collection schema
- **Schema introspection**: Use this endpoint to dynamically discover collection structure
- **Validation**: Schema information helps clients validate data before submission
- **Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Aggregation Operations

Moon provides dedicated aggregation endpoints that perform calculations directly on the server. This enables fast, efficient analytics—such as counting records, summing numeric fields, computing averages, and finding minimum or maximum values—without transferring unnecessary data.

**Aggregation Endpoints:**

- Count Records: `GET /{collection_name}:count`
- Sum Numeric Field: `GET /{collection_name}:sum` (requires `?field=...`)
- Average Numeric Field: `GET /{collection_name}:avg` (requires `?field=...`)
- Minimum Value: `GET /{collection_name}:min` (requires `?field=...`)
- Maximum Value: `GET /{collection_name}:max` (requires `?field=...`)

**Note:**

- Replace `{collection_name}` with your collection name.
- Aggregation can be combined with filters (e.g., `?quantity[gt]=10`) to perform calculations on specific subsets of data.
- Aggregation functions (`sum`, `avg`, `min`, `max`) are supported only on `integer` and `decimal` field types.

**Example Request:**

```sh
GET /products:sum?field=quantity
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 55
  }
}
```

**Aggregation with Filters:** Combine aggregation with query filters for calculations on specific subsets:

- `/products:count?quantity[gt]=10`
- `/products:sum?field=quantity&brand[eq]=Wow`
- `/products:max?field=quantity`

**Validation note:** Returns `400` (Standard Error Response) if `field` is missing or the named field is not a numeric type (`integer` or `decimal`) in the collection schema.

**Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling


## Standard Error Response

The API uses a simple, consistent error handling approach and strictly follows standard HTTP semantics.

- `400`: Invalid request (validation error, invalid parameter, malformed request)
- `401`: Authentication required
- `404`: Resource not found
- `500`: Server error
 - Only the codes listed above are permitted; do not use any others.

Note: `403` (Forbidden) is intentionally omitted in this specification to keep the error surface small. Authorization or permission failures should be handled via `401` per this document. If an implementation needs to distinguish "authenticated but not allowed" cases, add `403` and follow the same single-`message` JSON body pattern.
- Errors are indicated by standard HTTP status codes (for machines).
- Each error response includes only a single `message` field (for humans), intended for direct display to users.
- No internal error codes or additional error metadata are used.
- The HTTP status code is the only machine-readable error signal.
- Clients are not expected to parse or branch on error types.

When an error occurs, the API responds with the appropriate HTTP status code and a JSON body:

```json
{
  "message": "A human-readable description of the error"
}
```

### Examples

**HTTP 400: Validation Error**

```json
{
  "message": "Email format is invalid"
}
```

**HTTP 401: Unauthorized**

```json
{
  "message": "Authentication required"
}
```

**HTTP 404: Not Found**

```json
{
  "message": "User with id '01KHCZGWWRBQBREMG0K23C6C5H' not found"
}
```

**HTTP 500: Server Error**

```json
{
  "message": "An unexpected error occurred"
}
```



# Standard API Response Patterns

This document describes the standard response patterns, query options, and aggregation operations for the Moon API. All endpoints follow consistent conventions for success and error responses.

## Documentation and Health Endpoints

### Documentation Endpoints

Access API documentation in multiple formats.

#### View HTML Documentation

`GET /doc/`

View interactive HTML documentation in browser.

**URL:** [http://localhost:6006/doc/](http://localhost:6006/doc/)

---

#### Get Markdown Documentation

`GET /doc/llms.md`

Retrieve documentation in Markdown format (for humans and AI coding agents).

**Response (200 OK):**

Returns Markdown-formatted documentation.

**URL:** [http://localhost:6006/doc/llms.md](http://localhost:6006/doc/llms.md)

---

#### Get Text Documentation

`GET /doc/llms.txt`

Retrieve documentation in plain text format.

**Response (200 OK):**

Returns plain text documentation.

---

#### Get JSON Schema

`GET /doc/llms.json`

Retrieve machine-readable API schema in JSON format.

**Response (200 OK):**

```json
{
  "data": {
    "version": "1.0",
    "endpoints": [...],
    "schemas": {...}
  }
}
```

---

#### Refresh Documentation Cache

`POST /doc:refresh`

Force refresh of the documentation cache.

**Headers:**

- `Authorization: Bearer {access_token}` (required)

**Response (200 OK):**

```json
{
  "message": "Documentation cache refreshed successfully"
}
```

---

### Health Check Endpoint

`GET /health`

Check API service health and version information.

**Response (200 OK):**

```json
{
  "data": {
    "name": "moon",
    "status": "live",
    "version": "1.0"
  }
}
```

### Important Notes

- **Documentation formats**: Available in HTML (interactive), Markdown (human/AI readable), plain text, and JSON (machine-readable)
- **Cache refresh**: Documentation is cached for performance. Use `/doc:refresh` after configuration changes or schema updates
- **Health check**: No authentication required. Use for monitoring and uptime checks
- **Version tracking**: The version field in health response indicates the current API version

### Error Response

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

**Common error codes:**

- `UNAUTHORIZED`: Missing or invalid authentication (for `/doc:refresh`)
- `NOT_FOUND`: Documentation format not available

## Authentication Endpoints

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

### Error Response

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid credentials"
  }
}
```

**Common error codes:**

- `UNAUTHORIZED`: Invalid credentials or expired token
- `INVALID_PARAMETER`: Missing required fields
- `VALIDATION_ERROR`: Invalid email format or weak password
- `FORBIDDEN`: Insufficient permissions

## Standard Response Pattern for `:list` Endpoints

List endpoints return paginated collections of resources.

**Applicable Endpoints:**

- List Users: `GET /users:list`
- List API Keys: `GET /apikeys:list`
- List Collections: `GET /collections:list`
- List Collection Records: `GET /{collection_name}:list`

### Response Structure

All list endpoints return a consistent JSON structure with two main sections: `data` (the array of records) and `meta` (pagination information).

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

### Pagination

The API uses **unidirectional cursor-based pagination**. The `after` parameter always returns records that come after the specified cursor.

**Example Usage:**

```sh
# First page
GET /products:list?limit=15

# Next page - use meta.next cursor
GET /products:list?after=01KHCZKMM0N808MKSHBNWF464F&limit=15

# Previous page - use meta.prev cursor
GET /products:list?after=01KHCZFXAFJPS9SKSFKNBMHTP5&limit=15
```

**How It Works:**

- `meta.next`: Cursor pointing to the last record in the current page. Using this with `after` returns the next page.
- `meta.prev`: Cursor pointing to a record before the current page. Using this with `after` returns the previous page.
- Both cursors use the same `after` parameter for simplicity and consistency.

### Parameters

| Parameter | Type    | Description                                      |
| --------- | ------- | ------------------------------------------------ |
| `limit`   | integer | Number of items per page (default: 15, max: 100) |
| `after`   | string  | ULID cursor - returns records after this cursor  |

### Important Notes

- **Null cursors**: `prev` is `null` on the first page; `next` is `null` on the last page.
- **Sort order**: Records are always returned in consistent chronological order (by ULID/creation time).
- **No cursor**: Omitting `after` returns the first page.
- **ID requirement**: Each record in `data` must include an `id` field (ULID), except for collections which use `name` as the identifier.
- **Cursor format**: Cursors are ULIDs pointing to specific records.
- **Invalid cursor**: Returns error `RECORD_NOT_FOUND` if the cursor doesn't exist.

### Error Response

When an error occurs, the API returns a structured error response:

```json
{
  "error": {
    "code": "RECORD_NOT_FOUND",
    "message": "The requested cursor does not exist"
  }
}
```

**Common Error Codes:**

- `RECORD_NOT_FOUND`: Invalid cursor or resource doesn't exist.
- `INVALID_PARAMETER`: Invalid limit value (e.g., exceeds max of 100).
- `UNAUTHORIZED`: Missing or invalid authentication.

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

| Parameter | Type   | Description                                                 |
| --------- | ------ | ----------------------------------------------------------- |
| `id`      | string | ULID of the resource (required for users, apikeys, records) |
| `name`    | string | Name of the collection (required for collections)           |

### Important Notes

- **Single object**: The `data` field contains a single object (not an array).
- **No meta field**: Get endpoints don't need pagination metadata.
- **Consistent wrapper**: All `:get` endpoints use the `data` wrapper, matching `:list` endpoints.
- **404 error**: Returns `RECORD_NOT_FOUND` if the resource doesn't exist.

### Error Response

```json
{
  "error": {
    "code": "RECORD_NOT_FOUND",
    "message": "User with id '01KHCZGWWRBQBREMG0K23C6C5H' not found"
  }
}
```

**Common Error Codes:**

- `RECORD_NOT_FOUND`: Resource with specified ID/name doesn't exist.
- `INVALID_PARAMETER`: Missing or invalid `id` or `name` parameter.
- `UNAUTHORIZED`: Missing or invalid authentication.

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

### Error Response

**Full Failure (No Records Created):**

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "All records failed validation"
  }
}
```

**Common Error Codes:**

- `VALIDATION_ERROR`: Invalid input or constraint violation.
- `DUPLICATE_RECORD`: Resource with unique field already exists.
- `INVALID_PARAMETER`: Missing required fields.
- `UNAUTHORIZED`: Missing or invalid authentication.

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

| Parameter | Type   | Description                                          |
| --------- | ------ | ---------------------------------------------------- |
| `id`      | string | ULID of the resource (required for users, apikeys)   |
| `name`    | string | Name of the collection (required for collections)    |
| `data`    | array  | Array of record IDs to delete (required for records) |

### Important Notes

- **Array format**: Collection records must be sent as an array in `data`, even for single deletions.
- **Deleted IDs returned**: Response includes `data` array with IDs of successfully deleted records.
- **Partial success**: If some records fail to delete, the successfully deleted count is shown in `meta`.
- **Failed records**: Check `meta.failed` count to detect partial failures. Failed record IDs are excluded from the `data` array.
- **Status code**: Returns `200 OK` if at least one record was deleted successfully.
- **Message field**: Always includes a human-readable success message.

### Error Response

**Full Failure (No Records Deleted):**

```json
{
  "error": {
    "code": "RECORD_NOT_FOUND",
    "message": "No records found with provided IDs"
  }
}
```

**Common Error Codes:**

- `RECORD_NOT_FOUND`: Resource with specified ID/name doesn't exist.
- `INVALID_PARAMETER`: Missing required `id`, `name`, or `data` parameter.
- `UNAUTHORIZED`: Missing or invalid authentication.
- `FORBIDDEN`: Cannot delete protected resource.

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
  "name": "products",
  "add_columns": [
    {
      "name": "stock",
      "type": "integer",
      "nullable": false
    }
  ]
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

| Parameter    | Type   | Description                                                                  |
| ------------ | ------ | ---------------------------------------------------------------------------- |
| `id`         | string | ULID of the resource (required for users, apikeys)                           |
| Request Body | object | Fields to update OR `action` parameter for special operations                |
| `action`     | string | Special operation to perform (`reset_password`, `revoke_sessions`, `rotate`) |
| `name`       | string | Collection name (required for collection operations)                         |
| `data`       | array  | Array with objects containing `id` + fields to update (for records)          |

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

### Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid field values provided"
  }
}
```

**Common Error Codes:**

- `RECORD_NOT_FOUND`: Resource with specified ID doesn't exist.
- `VALIDATION_ERROR`: Invalid input or constraint violation.
- `INVALID_ACTION`: Unsupported action specified.
- `INVALID_PARAMETER`: Missing required fields for action.
- `UNAUTHORIZED`: Missing or invalid authentication.
- `FORBIDDEN`: Cannot update protected resource.

## Query Options

Query parameters for filtering, sorting, searching, field selection, and pagination when listing records. These options allow you to retrieve specific subsets of data based on your criteria.

| Query Option              | Description                                                  |
| ------------------------- | ------------------------------------------------------------ |
| `?column[operator]=value` | Filter records by column values using comparison operators   |
| `?sort={fields}`          | Sort by one or more fields (prefix `-` for descending order) |
| `?q={term}`               | Full-text search across all text columns                     |
| `?fields={field1,field2}` | Select specific fields to return (`id` is always included)   |
| `?limit={number}`         | Limit the number of records returned (default: 15, max: 100) |
| `?after={cursor}`         | Get records after the specified cursor                       |

### Filtering

**Query Option:** `?column[operator]=value`

**Supported Operators:** `eq`, `ne`, `gt`, `lt`, `gte`, `lte`, `like`, `in`

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:list?quantity[gt]=5&brand[eq]=Wow" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
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

**Query Option:** `?sort={-field1,field2}`

Sort by `field` (ascending) or `-field` (descending). Multiple fields can be specified.

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:list?sort=-quantity,title" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

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

**Query Option:** `?q={search_term}`

Searches across all string/text fields in the collection.

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:list?q=mouse" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
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

**Query Option:** `?fields={field1,field2}`

Returns only the specified fields (plus `id`, which is always included).

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:list?fields=quantity,title" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KHCZKSBQV1KH69AA6PVS12MM",
      "quantity": 10,
      "title": "Wireless Mouse"
    },
    {
      "id": "01KHCZKSPHB01TBEWKYQDKG5KS",
      "quantity": 55,
      "title": "USB Keyboard"
    },
    {
      "id": "01KHCZKT086EEB3EKM3PZ3N2Q0",
      "quantity": 20,
      "title": "Monitor 21 inch"
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

### Limit

**Query Option:** `?limit={number}`

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:list?limit=2" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
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

### Pagination

**Query Option:** `?after={cursor}`

Refer to the `:list` endpoint documentation for detailed pagination behavior.

### Combined Query Examples

Query parameters can be combined to perform complex queries. Here are some examples:

**Filter, sort, and limit:**

```bash
curl -g "http://localhost:6006/products:list?quantity[gte]=10&price[lt]=100&sort=-price&limit=5" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Search with category filter and field selection:**

```bash
curl -g "http://localhost:6006/products:list?q=laptop&brand[eq]=Wow&fields=title,price,quantity" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Multiple filters with pagination:**

```bash
curl -g "http://localhost:6006/products:list?price[gte]=100&quantity[gt]=0&sort=-price&limit=10&after=01ARZ3NDEK" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

## Aggregation Operations

Traditional analytics and reporting often require downloading large datasets to the client for processing, which is inefficient and slow—especially when counting, summing, or calculating averages across millions of records.

Moon provides dedicated aggregation endpoints that perform calculations directly on the server. This enables fast, efficient analytics—such as counting records, summing numeric fields, computing averages, and finding minimum or maximum values—without transferring unnecessary data.

**Server-Side Aggregation Endpoints:**

Replace `{collection}` with your collection name.

| Endpoint              | Method | Description                                   |
| --------------------- | ------ | --------------------------------------------- |
| `/{collection}:count` | GET    | Count records                                 |
| `/{collection}:sum`   | GET    | Sum numeric field (requires `?field=...`)     |
| `/{collection}:avg`   | GET    | Average numeric field (requires `?field=...`) |
| `/{collection}:min`   | GET    | Minimum value (requires `?field=...`)         |
| `/{collection}:max`   | GET    | Maximum value (requires `?field=...`)         |

**Note:**

- Aggregation can be combined with filters (e.g., `?quantity[gt]=10`) to perform calculations on specific subsets of data.
- Aggregation functions (`sum`, `avg`, `min`, `max`) are supported only on `integer` and `decimal` field types.

### Count Records

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:count" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 3
  }
}
```

### Sum Numeric Field

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:sum?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 85
  }
}
```

### Average Numeric Field

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:avg?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 28.333333333333332
  }
}
```

### Minimum Value

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:min?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 10
  }
}
```

### Maximum Value

**Example:**

```bash
curl -s -X GET "http://localhost:6006/products:max?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "value": 55
  }
}
```

### Aggregation with Filters

Combine aggregation with query filters for calculations on specific subsets:

**Count products with quantity greater than 10:**

```bash
curl -s -X GET "http://localhost:6006/products:count?quantity[gt]=10" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Sum quantity for specific brand:**

```bash
curl -s -X GET "http://localhost:6006/products:sum?field=quantity&brand[eq]=Wow" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Average price for products in stock:**

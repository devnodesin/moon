# Standard API Response Patterns

This document describes the standard response patterns, query options, and aggregation operations for the Moon API. All endpoints follow consistent conventions for success and error responses.


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

**For Users:**

```sh
POST /users:update?id=01KHCZGWWRBQBREMG0K23C6C5H
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

**For API Keys:**

```sh
POST /apikeys:update?id=01KHCZKCR7MHB0Q69KM63D6AXF
```

Standard update (request body wrapped in `data`):

```json
{
  "data": {
    "name": "Updated Service Name",
    "description": "Updated description",
    "can_write": true
  }
}
```

Special actions:

```json
{
  "data": {
    "action": "rotate"
  }
}
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZKCR7MHB0Q69KM63D6AXF",
    "name": "Updated Service Name",
    "description": "Updated description",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-14T02:27:38Z"
  },
  "message": "API key updated successfully"
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
- **Total count**: Represents the total number of fields in the collection schema.
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



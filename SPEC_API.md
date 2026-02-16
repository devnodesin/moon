## Standard API Response

### Standard Response Pattern for `:list` Endpoints

- List Users `GET /users:list`
- List API Keys `GET /apikeys:list`
- List Collections `GET /collections:list`
- List Collection Records `GET /{collection_name}:list`

**Response Structure:**

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

**Pagination:**

The API uses **unidirectional cursor pagination**. The `after` parameter always returns records that come after the specified cursor.

```sh
# First page
GET /products:list?limit=15

# Next page - use meta.next cursor
GET /products:list?after=01KHCZKMM0N808MKSHBNWF464F&limit=15

# Previous page - use meta.prev cursor
GET /products:list?after=01KHCZFXAFJPS9SKSFKNBMHTP5&limit=15
```

**How it works:**

- `meta.next`: Cursor pointing to the last record in the current page
  - Using this with `after` returns the next page
- `meta.prev`: Cursor pointing to a record **before** the current page
  - Using this with `after` returns the previous page
- Both cursors use the same `after` parameter for simplicity

**Parameters:**

- `limit` (optional): Number of items per page (default: 15, max: 100)
- `after` (optional): ULID cursor - returns records after this cursor

**Important Notes:**

- **Null cursors**: `prev` is `null` on the first page; `next` is `null` on the last page
- **Sort order**: Records are always returned in consistent chronological order (by ULID/creation time)
- **No cursor**: Omitting `after` returns the first page
- **ID requirement**: Each record in `data` must include an `id` field (ULID), except for collections which use `name` as identifier
- **Cursor format**: Cursors are ULIDs pointing to specific records
- **Invalid cursor**: Returns error `RECORD_NOT_FOUND` if cursor doesn't exist

**Error Response:**

```json
{
  "error": {
    "code": "RECORD_NOT_FOUND",
    "message": "The requested cursor does not exist"
  }
}
```

**Common error codes:**

- `RECORD_NOT_FOUND`: Invalid cursor or resource doesn't exist
- `INVALID_PARAMETER`: Invalid limit value (e.g., exceeds max of 100)
- `UNAUTHORIZED`: Missing or invalid authentication

### Standard Response Pattern for `:get` Endpoints

- Get User `GET /users:get?id={id}`
- Get API Key `GET /apikeys:get?id={id}`
- Get Collection `GET /collections:get?name={collection_name}`
- Get Collection Record `GET /{collection_name}:get?id={id}`

**Response Structure:**

**Get User, API Key and Collection Record:**

```sh
GET /users:get?id=01KHCZGWWRBQBREMG0K23C6C5H
GET /apikeys:get?id=01KHCZGWWRBQBREMG0K23C6C5H
GET /products:get?id=01KHCZGWWRBQBREMG0K23C6C5H
```

See the example responses below.

```json
{
  "data": {
    "id": "01KHCZKMM0N808MKSHBNWF464F",
    "title": "Wireless Mouse",
    "price": "29.99"
  }
}
```

**For Collection:**

```sh
GET /collections:get?name=products
```

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

**Parameters:**

- `id` (required for users, apikeys, records): ULID of the resource
- `name` (required for collections): Name of the collection

**Important Notes:**

- **Single object**: The `data` field contains a single object (not an array)
- **No meta field**: Get endpoints don't need pagination metadata
- **Consistent wrapper**: All `:get` endpoints use `data` wrapper, matching `:list` endpoints
- **404 error**: Returns `RECORD_NOT_FOUND` if the resource doesn't exist

**Error Response:**

```json
{
  "error": {
    "code": "RECORD_NOT_FOUND",
    "message": "User with id '01KHCZGWWRBQBREMG0K23C6C5H' not found"
  }
}
```

**Common error codes:**

- `RECORD_NOT_FOUND`: Resource with specified ID/name doesn't exist
- `INVALID_PARAMETER`: Missing or invalid `id` or `name` parameter
- `UNAUTHORIZED`: Missing or invalid authentication

### Standard Response Pattern for `:create` Endpoints

- Create User `POST /users:create`
- Create API Key `POST /apikeys:create`
- Create Collection `POST /collections:create`
- Create Collection Record(s) `POST /{collection_name}:create`

**Response Structure:**

**Create User, API Key, Collection:**

```sh
POST /users:create
POST /apikeys:create
POST /collections:create
```

Request body:

```json
{
  "data": {
    "username": "moonuser",
    "email": "moonuser@example.com",
    "password": "UserPass123#"
  }
}
```

Response (201 Created):

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

**Create Collection Record(s):**

Single record:

```json
{
  "data": [
    { "title": "Wireless Mouse", "price": "29.99" }
  ]
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

Response (201 Created) - All succeeded:

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

Response (201 Created) - Partial success:

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

**Collections Create:**

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

**Important Notes:**

- **id field**: The `id` field is system-generated and read-only. Do not include it in create requests.
- **Array format**: Collection records must always be sent as an array in `data`, even for single records
- **Partial success**: If some records fail validation, successfully created records are returned in `data`
- **Failed records**: Failed records are excluded from `data` array. Check `meta.failed` count to detect partial failures.
- **Status code**: Always returns `201 Created` if at least one record was created successfully
- **Consistent wrapper**: All `:create` endpoints use `data` field for created resource(s)
- **Message field**: Always includes a human-readable success message
- **API Key security**: The `key` field appears in `data` only once during creation

**Error Response:**

Full failure (no records created):

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "All records failed validation"
  }
}
```

**Common error codes:**

- `VALIDATION_ERROR`: Invalid input or constraint violation
- `DUPLICATE_RECORD`: Resource with unique field already exists
- `INVALID_PARAMETER`: Missing required fields
- `UNAUTHORIZED`: Missing or invalid authentication

### Standard Response Pattern for `:destroy` Endpoints

- Delete User `POST /users:destroy?id={id}`
- Delete API Key `POST /apikeys:destroy?id={id}`
- Delete Collection `POST /collections:destroy?name={collection_name}`
- Delete Collection Record(s) `POST /{collection_name}:destroy`

**Response Structure:**

**Delete User, API Key, Collection:**

```sh
POST /users:destroy?id=01KHCZGWWRBQBREMG0K23C6C5H
POST /apikeys:destroy?id=01KHCZKCR7MHB0Q69KM63D6AXF
POST /collections:destroy?name=products
```

Response (200 OK):

```json
{
  "message": "User deleted successfully"
}
```

**Delete Collection Record(s):**

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

Response (200 OK) - All succeeded:

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

Response (200 OK) - Partial success:

```json
{
  "data": [
    "01KHCZKMXYVC1NRHDZ83XMHY4N",
    "01KHCZKMY28ERJFPCVBQEKQ4SY"
  ],
  "meta": {
    "total": 3,
    "succeeded": 2,
    "failed": 1
  },
  "message": "2 of 3 record(s) deleted successfully"
}
```

**Parameters:**

- `id` (required for users, apikeys): ULID of the resource
- `name` (required for collections): Name of the collection
- Request body `data` (required for records): Array of record IDs to delete

**Important Notes:**

- **Array format**: Collection records must be sent as an array in `data`, even for single deletions
- **Deleted IDs returned**: Response includes `data` array with IDs of successfully deleted records
- **Partial success**: If some records fail to delete, successfully deleted count is shown in `meta`
- **Failed records**: Check `meta.failed` count to detect partial failures. Failed record IDs are excluded from `data` array.
- **Status code**: Returns `200 OK` if at least one record was deleted successfully
- **Message field**: Always includes a human-readable success message

**Error Response:**

Full failure (no records deleted):

```json
{
  "error": {
    "code": "RECORD_NOT_FOUND",
    "message": "No records found with provided IDs"
  }
}
```

**Common error codes:**

- `RECORD_NOT_FOUND`: Resource with specified ID/name doesn't exist
- `INVALID_PARAMETER`: Missing required `id`, `name`, or `data` parameter
- `UNAUTHORIZED`: Missing or invalid authentication
- `FORBIDDEN`: Cannot delete protected resource

### Standard Response Pattern for `:update` Endpoints

- Update User `POST /users:update?id={id}`
- Update API Key `POST /apikeys:update?id={id}`
- Update Collection `POST /collections:update`
- Update Collection Record(s) `POST /{collection_name}:update`

#### Response Structure

**Update User, API Key:**

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

Response (200 OK):

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

**Update Collection:**

```sh
POST /collections:update
```

Request body:

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

Response (200 OK):

```json
{
  "data": {
    "name": "products",
    "columns": [...]
  },
  "message": "Collection 'products' updated successfully"
}
```

**Update Collection Record(s):**

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

Response (200 OK):

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

#### Special Actions

**User Actions:**

`reset_password` - Reset user password:

```json
{
  "action": "reset_password",
  "new_password": "NewSecurePassword123#"
}
```

`revoke_sessions` - Revoke all active sessions:

```json
{
  "action": "revoke_sessions"
}
```

**API Key Actions:**

`rotate` - Generate new key and invalidate old one:

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

#### Parameters

- `id` (required for users, apikeys): ULID of the resource
- Request body: Fields to update OR `action` parameter for special operations
- `action` (optional): Special operation to perform (`reset_password`, `revoke_sessions`, `rotate`)
- Collection operations: `name` (required) + operation fields
- Record updates: `data` array with objects containing `id` + fields to update

#### Important Notes

- **Array format**: Collection records must be sent as an array in `data`, even for single updates
- **Partial updates**: Only fields provided are updated; other fields remain unchanged
- **Actions vs updates**: When `action` is specified, it takes precedence over field updates
- **Action-specific fields**: Some actions require additional fields (e.g., `new_password` for `reset_password`)
- **Updated data returned**: Response includes full updated resource(s) in `data`
- **Partial success**: For batch updates, successfully updated records are returned in `data`
- **Status code**: Returns `200 OK` if at least one record was updated successfully
- **Key rotation**: `rotate` action returns new key in `data.key` field (shown only once)
- **Warning field**: Optional field for security warnings (e.g., key rotation, password reset)

#### Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid field values provided"
  }
}
```

**Common error codes:**

- `RECORD_NOT_FOUND`: Resource with specified ID doesn't exist
- `VALIDATION_ERROR`: Invalid input or constraint violation
- `INVALID_ACTION`: Unsupported action specified
- `INVALID_PARAMETER`: Missing required fields for action
- `UNAUTHORIZED`: Missing or invalid authentication
- `FORBIDDEN`: Cannot update protected resource

### Query Options

Query parameters for filtering, sorting, searching, field selection, and pagination when listing records. Using these options allows you to retrieve specific subsets of data based on your criteria.

| Query Options | Description |
|---------------|-------------|
| `?column[operator]=value` | Filter records by column values using comparison operators |
| `?sort={fields}` | Sort by one or more fields (prefix `-` for descending) |
| `?q={term}` | Full-text search across all text columns |
| `?fields={field1,field2}` | Select specific fields to return (id always included) |
| `?limit={number}` | Limit number of records returned (default: 15, max: 100) |
| `?after={cursor}` | Get records after the specified cursor |

#### Filtering

**Query Option:** `?column[operator]=value`

**Operators:** eq, ne, gt, lt, gte, lte, like, in

```bash
curl -s -X GET "http://localhost:6006/products:list?quantity[gt]=5&brand[eq]=Wow" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

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

#### Sorting

**Query Option:** `?sort={-field1,field2}`

Sort by `field` (ascending) or `-field` (descending).

```bash
curl -s -X GET "http://localhost:6006/products:list?sort=-quantity,title" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

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

#### Full-Text Search

**Query Option:** `?q={search_term}` (across all text columns)

Searches across all string/text fields in the collection.

```bash
curl -s -X GET "http://localhost:6006/products:list?q=mouse" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

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

#### Field Selection

**Query Option:** `?fields={field1,field2}`

Returns only the specified fields (plus `id` which is always included).

```bash
curl -s -X GET "http://localhost:6006/products:list?fields=quantity,title" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

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

#### Limit

**Query Option:** `?limit={limit}`

```bash
curl -s -X GET "http://localhost:6006/products:list?limit=2" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

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

#### Pagination

**Query Option:** `?after={cursor}`

Refer to `:list` endpoint documentation for detailed pagination behavior.

#### Combined Query Examples

Query parameters can be combined to perform complex queries. Here are some examples:

Filter, sort, and limit:

```bash
curl -g "http://localhost:6006/products:list?quantity[gte]=10&price[lt]=100&sort=-price&limit=5" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Search with category filter and field selection:

```bash
curl -g "http://localhost:6006/products:list?q=laptop&brand[eq]=Wow&fields=title,price,quantity" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Multiple filters with pagination:

```bash
curl -g "http://localhost:6006/products:list?price[gte]=100&quantity[gt]=0&sort=-price&limit=10&after=01ARZ3NDEK" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

### Aggregation Operations

Traditional analytics and reporting often require downloading large datasets to the client for processing, which is inefficient and slow—especially for counting, summing, or calculating averages across millions of records.

Moon provides dedicated aggregation endpoints that perform calculations directly on the server. This enables fast, efficient analytics—such as counting records, summing numeric fields, computing averages, and finding minimum or maximum values—without transferring unnecessary data.

Server-side aggregation endpoints for analytics. Replace `{collection}` with your collection name.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/{collection}:count` | GET | Count records |
| `/{collection}:sum` | GET | Sum numeric field (requires `?field=...`) |
| `/{collection}:avg` | GET | Average numeric field (requires `?field=...`) |
| `/{collection}:min` | GET | Minimum value (requires `?field=...`) |
| `/{collection}:max` | GET | Maximum value (requires `?field=...`) |

**Note:**

- Aggregation can be combined with filters (e.g., `?quantity[gt]=10`) to perform calculations on specific subsets of data.
- Aggregation functions (`sum`, `avg`, `min`, `max`) are supported only on `integer` and `decimal` field types.

#### Count Records

```bash
curl -s -X GET "http://localhost:6006/products:count" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

```json
{
  "data": {
    "value": 3
  }
}
```

#### Sum Numeric Field

```bash
curl -s -X GET "http://localhost:6006/products:sum?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

```json
{
  "data": {
    "value": 85
  }
}
```

#### Average Numeric Field

```bash
curl -s -X GET "http://localhost:6006/products:avg?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

```json
{
  "data": {
    "value": 28.333333333333332
  }
}
```

#### Minimum Value

```bash
curl -s -X GET "http://localhost:6006/products:min?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

```json
{
  "data": {
    "value": 10
  }
}
```

#### Maximum Value

```bash
curl -s -X GET "http://localhost:6006/products:max?field=quantity" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

Response (200 OK):

```json
{
  "data": {
    "value": 55
  }
}
```

#### Aggregation with Filters

Combine aggregation with query filters for calculations on specific subsets:

```bash
# Count products with quantity greater than 10
curl -s -X GET "http://localhost:6006/products:count?quantity[gt]=10" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

```bash
# Sum quantity for specific brand
curl -s -X GET "http://localhost:6006/products:sum?field=quantity&brand[eq]=Wow" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

```bash
# Average price for products in stock
curl -s -X GET "http://localhost:6006/products:avg?field=price&quantity[gt]=0" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

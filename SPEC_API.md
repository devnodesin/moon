## Standard Response Pattern for `:list` Endpoints

- List Users `GET /users:list`
- List API Keys `GET /apikeys:list`
- List Collections `GET /collections:list`
- List Collection Records `GET /{collection_name}:list`

**Response Structure**

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

**Important Notes**

- **Null cursors**: `prev` is `null` on the first page; `next` is `null` on the last page
- **Sort order**: Records are always returned in consistent chronological order (by ULID/creation time)
- **No cursor**: Omitting `after` returns the first page
- **ID requirement**: Each record in `data` must include an `id` field (ULID), except for collections which use `name` as identifier
- **Cursor format**: Cursors are ULIDs pointing to specific records
- **Invalid cursor**: Returns error `RECORD_NOT_FOUND` if cursor doesn't exist

**Error Response**

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

## Standard Response Pattern for `:get` Endpoints

- Get User `GET /users:get?id={id}`
- Get API Key `GET /apikeys:get?id={id}`
- Get Collection `GET /collections:get?name={collection_name}`
- Get Collection Record `GET /{collection_name}:get?id={id}`

**Response Structure**

```json
{
  "data": {
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

### Examples

**Get User:**

```sh
GET /users:get?id=01KHCZGWWRBQBREMG0K23C6C5H
```

```json
{
  "data": {
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

**Get API Key:**

```sh
GET /apikeys:get?id=01KHCZKCR7MHB0Q69KM63D6AXF
```

```json
{
  "data": {
    "id": "01KHCZKCR7MHB0Q69KM63D6AXF",
    "name": "Integration Service",
    "description": "Key for integration",
    "role": "user",
    "can_write": false,
    "created_at": "2026-02-14T02:27:42Z"
  }
}
```

**Get Collection:**

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
      },
      {
        "name": "price",
        "type": "integer",
        "nullable": false,
        "unique": false
      }
    ]
  }
}
```

**Get Collection Record:**

```sh
GET /products:get?id=01KHCZKMM0N808MKSHBNWF464F
```

```json
{
  "data": {
    "id": "01KHCZKMM0N808MKSHBNWF464F",
    "title": "Wireless Mouse",
    "brand": "Wow",
    "details": "Ergonomic wireless mouse",
    "price": "29.99",
    "quantity": 10
  }
}
```

**Parameters**

- `id` (required for users, apikeys, records): ULID of the resource
- `name` (required for collections): Name of the collection

**Important Notes**

- **Single object**: The `data` field contains a single object (not an array)
- **No meta field**: Get endpoints don't need pagination metadata
- **Consistent wrapper**: All `:get` endpoints use `data` wrapper, matching `:list` endpoints
- **404 error**: Returns `RECORD_NOT_FOUND` if the resource doesn't exist

**Error Response**

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

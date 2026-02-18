## Standard Response Pattern for `:update` Endpoints

Update endpoints modify existing resources in the system.

### Applicable Endpoints

- Update User: `POST /users:update?id={id}`
- Update API Key: `POST /apikeys:update?id={id}`
- Update Collection: `POST /collections:update`
- Update Collection Record(s): `POST /{collection_name}:update`

### Response Structure

**Update User or API Key:**

Request parameters:

- Query: `?id={user_id}` or `?id={apikey_id}`
- Body: Fields to update

Request body:

```json
{
  "email": "updateduser@example.com",
  "role": "admin"
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
    "can_write": true,
    "updated_at": "2026-02-14T02:27:39Z"
  },
  "message": "User updated successfully"
}
```

**Update Collection:**

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

Other collection operations:

- `rename_columns`: Rename existing columns
- `modify_columns`: Change column properties
- `remove_columns`: Delete columns

Response (200 OK):

```json
{
  "data": {
    "name": "products",
    "columns": [
      {
        "name": "stock",
        "type": "integer",
        "nullable": false,
        "unique": false
      }
    ]
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

Response (200 OK) - All succeeded:

```json
{
  "data": [
    {
      "id": "01KHCZKMM0N808MKSHBNWF464F",
      "price": "100.00",
      "title": "Updated Product"
    },
    {
      "id": "01KHCZKMXYVC1NRHDZ83XMHY4N",
      "price": "200.00",
      "title": "Updated Monitor"
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

Response (200 OK) - Partial success:

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
    "succeeded": 1,
    "failed": 1
  },
  "message": "1 of 2 record(s) updated successfully"
}
```

### Special Actions

Special actions use the `action` parameter to perform operations beyond standard field updates.

**User Actions:**

Reset password:

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
    "can_write": true,
    "updated_at": "2026-02-14T02:27:40Z"
  },
  "message": "Password reset successfully"
}
```

Revoke all sessions:

```json
{
  "action": "revoke_sessions"
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
    "can_write": true,
    "updated_at": "2026-02-14T02:27:40Z"
  },
  "message": "All sessions revoked successfully"
}
```

**API Key Actions:**

Rotate API key (generate new key and invalidate old one):

```json
{
  "action": "rotate"
}
```

Response (200 OK):

```json
{
  "data": {
    "id": "01KHCZKCR7MHB0Q69KM63D6AXF",
    "name": "Updated Service Name",
    "description": "Updated description",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-14T02:27:42Z",
    "key": "moon_live_9wAtNeHqBdYmQaf3Dvm8YhVM4FK880X4dj5dqFSqLJ6qJmApxHuYe6gqkI3ipgKG"
  },
  "message": "API key regenerated successfully. Store this key securely. It will not be shown again."
}
```

### Parameters

- `id` (required for users, apikeys): ULID of the resource (in query string)
- Request body: Fields to update OR `action` parameter for special operations
- `action` (optional): Special operation to perform (`reset_password`, `revoke_sessions`, `rotate`)
- Collection operations: `name` (required) + operation fields (`add_columns`, `rename_columns`, `modify_columns`, `remove_columns`)
- Record updates: `data` array with objects containing `id` + fields to update

### Important Notes

- **Array format**: Collection records must be sent as an array in `data`, even for single updates
- **Query parameters**: User and API Key updates require `?id={id}` in the URL
- **Partial updates**: Only fields provided are updated; other fields remain unchanged
- **Actions vs updates**: When `action` is specified, it takes precedence over field updates
- **Action-specific fields**: Some actions require additional fields (e.g., `new_password` for `reset_password`)
- **Updated data returned**: Response includes full updated resource(s) in `data`
- **Partial success**: For batch updates, successfully updated records are returned in `data`
- **Status code**: Returns `200 OK` if at least one record was updated successfully
- **Key rotation**: `rotate` action returns new key in `data.key` field (shown only once)
- **Warning field**: Optional field for security warnings (e.g., key rotation, password reset)

### Error Handling

**Error Response:** Follow [Standard Error Response](#standard-error-response) for any error handling

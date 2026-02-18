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

**Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling

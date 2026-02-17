### Create API Key

```bash
curl -s -X POST "http://localhost:6006/apikeys:create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "name": "Integration Service",
        "description": "Key for integration",
        "role": "user",
        "can_write": false
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "data": {
    "can_write": false,
    "created_at": "2026-02-17T10:58:53Z",
    "description": "Key for integration",
    "id": "01KHNM1HRQBFQECV2RX8EEMVWH",
    "key": "moon_live_bJoL6mveHqWBssN5jXWpqchOK2z2HI3QnwqgA5Xaxnyxw6Vn5VLXjRXj0ui0DLxH",
    "name": "Integration Service",
    "role": "user"
  },
  "message": "API key created successfully",
  "warning": "Store this key securely. It will not be shown again."
}
```

### List API Keys

```bash
curl -s -X GET "http://localhost:6006/apikeys:list" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01KHNGH8D21N949H4H7ZSC4JMH",
      "name": "Another Service",
      "description": "Another key for testing",
      "role": "user",
      "can_write": false,
      "created_at": "2026-02-17T09:57:33Z"
    },
    {
      "id": "01KHNM1HRQBFQECV2RX8EEMVWH",
      "name": "Integration Service",
      "description": "Key for integration",
      "role": "user",
      "can_write": false,
      "created_at": "2026-02-17T10:58:53Z"
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

### Get API Key

```bash
curl -s -X GET "http://localhost:6006/apikeys:get?id=01KHNM1HRQBFQECV2RX8EEMVWH" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHNM1HRQBFQECV2RX8EEMVWH",
    "name": "Integration Service",
    "description": "Key for integration",
    "role": "user",
    "can_write": false,
    "created_at": "2026-02-17T10:58:53Z"
  }
}
```

### Update API Key Metadata

***Note:*** Update metadata fields (name, description, can_write) without changing the API key itself. The key remains valid. To generate a new key, use the rotation action.

```bash
curl -s -X POST "http://localhost:6006/apikeys:update?id=01KHNM1HRQBFQECV2RX8EEMVWH" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "name": "Updated Service Name",
        "description": "Updated description",
        "can_write": true
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHNM1HRQBFQECV2RX8EEMVWH",
    "name": "Updated Service Name",
    "description": "Updated description",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-17T10:58:53Z"
  },
  "message": "API key updated successfully"
}
```

### Rotate API Key

Use `rotate` to securely generate a new API key and invalidate the old one in a single step, minimizing overlap between valid keys.

```bash
curl -s -X POST "http://localhost:6006/apikeys:update?id=01KHNM1HRQBFQECV2RX8EEMVWH" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "action": "rotate"
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHNM1HRQBFQECV2RX8EEMVWH",
    "key": "moon_live_a5184k1gXaqGztjUJ5lsLSi7zsY3fZCuVJxpq7t6arsjb84hT5WnmSwZWUwPoCzk",
    "name": "Updated Service Name"
  },
  "message": "API key rotated successfully",
  "warning": "Store this key securely. The old key is now invalid."
}
```

### Delete API Key

```bash
curl -s -X POST "http://localhost:6006/apikeys:destroy?id=01KHNM1HRQBFQECV2RX8EEMVWH" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "API key deleted successfully"
}
```

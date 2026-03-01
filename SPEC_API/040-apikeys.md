### Create API Key

```bash
curl -s -X POST "http://localhost:6006/apikeys:create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": {
          "name": "Integration Service",
          "description": "Key for integration",
          "role": "user",
          "can_write": false
        }
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "data": {
    "can_write": false,
    "created_at": "2026-03-01T12:48:57Z",
    "description": "Key for integration",
    "id": "01KJMQ3PZX9543A1YX340S108D",
    "key": "moon_live_irpswNXjQMNDRsoJYW1eKVG0szqKPhPksCV4XZ1o9UBDIDGUI1sGBxVRUkh5Ec40",
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
      "id": "01KJHQA9D1EWJMGX785STJWSYH",
      "name": "Another Service",
      "description": "Another key for testing",
      "role": "user",
      "can_write": false,
      "created_at": "2026-02-28T08:54:49Z"
    },
    {
      "id": "01KJJK5862YTHS837B69ZWZV2G",
      "name": "Testing",
      "description": "Testing",
      "role": "user",
      "can_write": false,
      "created_at": "2026-02-28T17:01:24Z"
    },
    {
      "id": "01KJKZXAZCAQATV75PH96AGDF9",
      "name": "Wow Key",
      "description": "Wwwok",
      "role": "admin",
      "can_write": false,
      "created_at": "2026-03-01T06:03:31Z"
    },
    {
      "id": "01KJMQ3PZX9543A1YX340S108D",
      "name": "Integration Service",
      "description": "Key for integration",
      "role": "user",
      "can_write": false,
      "created_at": "2026-03-01T12:48:57Z"
    }
  ],
  "meta": {
    "count": 4,
    "limit": 15,
    "next": null,
    "prev": null
  }
}
```

### Get API Key

```bash
curl -s -X GET "http://localhost:6006/apikeys:get?id=01KJMQ3PZX9543A1YX340S108D" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJMQ3PZX9543A1YX340S108D",
    "name": "Integration Service",
    "description": "Key for integration",
    "role": "user",
    "can_write": false,
    "created_at": "2026-03-01T12:48:57Z"
  }
}
```

### Update API Key Metadata

***Note:*** Update metadata fields (name, description, can_write) without changing the API key itself. The key remains valid. To generate a new key, use the rotation action.

```bash
curl -s -X POST "http://localhost:6006/apikeys:update?id=01KJMQ3PZX9543A1YX340S108D" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": {
          "name": "Updated Service Name",
          "description": "Updated description",
          "can_write": true
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJMQ3PZX9543A1YX340S108D",
    "name": "Updated Service Name",
    "description": "Updated description",
    "role": "user",
    "can_write": true,
    "created_at": "2026-03-01T12:48:57Z"
  },
  "message": "API key updated successfully"
}
```

### Rotate API Key

Use `rotate` to securely generate a new API key and invalidate the old one in a single step, minimizing overlap between valid keys.

```bash
curl -s -X POST "http://localhost:6006/apikeys:update?id=01KJMQ3PZX9543A1YX340S108D" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": {
          "action": "rotate"
        }
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJMQ3PZX9543A1YX340S108D",
    "key": "moon_live_I7T1uNRduazIASRIIucsgctuktM2Rk1J9O0E3ezfAaxREEgMaQBoxqJzoAY1A6Gk",
    "name": "Updated Service Name"
  },
  "message": "API key rotated successfully",
  "warning": "Store this key securely. The old key is now invalid."
}
```

### Delete API Key

```bash
curl -s -X POST "http://localhost:6006/apikeys:destroy?id=01KJMQ3PZX9543A1YX340S108D" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "API key deleted successfully"
}
```

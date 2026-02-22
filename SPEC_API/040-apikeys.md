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
    "created_at": "2026-02-22T08:36:22Z",
    "description": "Key for integration",
    "id": "01KJ27W66RTNN823BHM1YSD90Z",
    "key": "moon_live_GqDqrGEIEfXhJMcENCykMkHH7A3I0wuEPYTZNvT6qGTnGvUZgiYf6bViNRo0C23w",
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
      "id": "01KJ27W66RTNN823BHM1YSD90Z",
      "name": "Integration Service",
      "description": "Key for integration",
      "role": "user",
      "can_write": false,
      "created_at": "2026-02-22T08:36:22Z"
    },
    {
      "id": "01KJ27W6D91D148FAWX8MF22NR",
      "name": "Another Service",
      "description": "Another key for testing",
      "role": "user",
      "can_write": false,
      "created_at": "2026-02-22T08:36:22Z"
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
curl -s -X GET "http://localhost:6006/apikeys:get?id=01KJ27W66RTNN823BHM1YSD90Z" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KJ27W66RTNN823BHM1YSD90Z",
    "name": "Integration Service",
    "description": "Key for integration",
    "role": "user",
    "can_write": false,
    "created_at": "2026-02-22T08:36:22Z"
  }
}
```

### Update API Key Metadata

***Note:*** Update metadata fields (name, description, can_write) without changing the API key itself. The key remains valid. To generate a new key, use the rotation action.

```bash
curl -s -X POST "http://localhost:6006/apikeys:update?id=01KJ27W66RTNN823BHM1YSD90Z" \
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
    "id": "01KJ27W66RTNN823BHM1YSD90Z",
    "name": "Updated Service Name",
    "description": "Updated description",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-22T08:36:22Z"
  },
  "message": "API key updated successfully"
}
```

### Rotate API Key

Use `rotate` to securely generate a new API key and invalidate the old one in a single step, minimizing overlap between valid keys.

```bash
curl -s -X POST "http://localhost:6006/apikeys:update?id=01KJ27W66RTNN823BHM1YSD90Z" \
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
    "id": "01KJ27W66RTNN823BHM1YSD90Z",
    "key": "moon_live_YfzXnixP5NJT334R9vy5s9uEhPnpinvNvZ6Jy2jATEo6GmOG1CDkrvSmEyKzUR0k",
    "name": "Updated Service Name"
  },
  "message": "API key rotated successfully",
  "warning": "Store this key securely. The old key is now invalid."
}
```

### Delete API Key

```bash
curl -s -X POST "http://localhost:6006/apikeys:destroy?id=01KJ27W66RTNN823BHM1YSD90Z" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "API key deleted successfully"
}
```

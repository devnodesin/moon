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
  "message": "API key created successfully",
  "warning": "Store this key securely. It will not be shown again.",
  "apikey": {
    "id": "01KH94D5BDE52VNTYK36268V9E",
    "name": "Integration Service",
    "description": "Key for integration",
    "role": "user",
    "can_write": false,
    "created_at": "2026-02-12T14:34:43Z"
  },
  "key": "moon_live_XcHL33LuDzu0uJ4vykHQg3MZgUWPXxq1AfO5Bd56LPWEoKj6Lu1Adu2Uc2tgSU9u"
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
  "apikeys": [
    {
      "id": "01KH92WRNXMGXY6KBRJAJ6TZ98",
      "name": "Another Service",
      "description": "Another key for testing",
      "role": "user",
      "can_write": false,
      "created_at": "2026-02-12T14:08:17Z"
    },
    {
      "id": "01KH94D5BDE52VNTYK36268V9E",
      "name": "Integration Service",
      "description": "Key for integration",
      "role": "user",
      "can_write": false,
      "created_at": "2026-02-12T14:34:43Z"
    }
  ],
  "next_cursor": null,
  "limit": 15
}
```

### Get API Key

```bash
curl -s -X GET "http://localhost:6006/apikeys:get?id=01KH92WRNXMGXY6KBRJAJ6TZ98" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "apikey": {
    "id": "01KH92WRNXMGXY6KBRJAJ6TZ98",
    "name": "Another Service",
    "description": "Another key for testing",
    "role": "user",
    "can_write": false,
    "created_at": "2026-02-12T14:08:17Z"
  }
}
```

### Update API Key Metadata

***Note:*** Update metadata fields (name, description, can_write) without changing the API key itself. The key remains valid. To generate a new key, use the rotation action.

```bash
curl -s -X POST "http://localhost:6006/apikeys:update?id=01KH92WRNXMGXY6KBRJAJ6TZ98" \
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
  "message": "API key updated successfully",
  "apikey": {
    "id": "01KH92WRNXMGXY6KBRJAJ6TZ98",
    "name": "Updated Service Name",
    "description": "Updated description",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-12T14:08:17Z"
  }
}
```

### Rotate API Key

Use `rotate` to securely generate a new API key and invalidate the old one in a single step, minimizing overlap between valid keys.

```bash
curl -s -X POST "http://localhost:6006/apikeys:update?id=01KH92WRNXMGXY6KBRJAJ6TZ98" \
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
  "message": "API key rotated successfully",
  "warning": "Store this key securely. It will not be shown again.",
  "apikey": {
    "id": "01KH92WRNXMGXY6KBRJAJ6TZ98",
    "name": "Updated Service Name",
    "description": "Updated description",
    "role": "user",
    "can_write": true,
    "created_at": "2026-02-12T14:08:17Z"
  },
  "key": "moon_live_q8wYDAKcA5IeRrdooLLNGbrx012ME0urYMWR3p8oHc6M4qzm20kwwT9hOPbfo1vH"
}
```

### Delete API Key

```bash
curl -s -X POST "http://localhost:6006/apikeys:destroy?id=01KH92WRNXMGXY6KBRJAJ6TZ98" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "API key deleted successfully"
}
```

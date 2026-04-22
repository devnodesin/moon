### Create Device API Key

Create a device API key after logging in with username and password. The returned key is reused by later requests through `$API_KEY`.

```bash
curl -s -X POST "http://localhost:6000/data/apikeys:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "name": "Moon Usage Key",
            "role": "user",
            "can_write": false,
            "collections": [
              "apikey_usage_devices"
            ],
            "is_website": false,
            "rate_limit": 15,
            "captcha_required": false,
            "enabled": true
          }
        ]
      }
    ' | jq .
```

**Response (201 Created):**

```json
{
  "message": "Resource created successfully",
  "data": [
    {
      "allowed_origins": null,
      "can_write": false,
      "captcha_required": false,
      "collections": [
        "apikey_usage_devices"
      ],
      "created_at": "2026-04-22T00:55:37Z",
      "enabled": true,
      "id": "01KPSAYYC4QHKMC6B47313B9W1",
      "is_website": false,
      "key": "$API_KEY",
      "name": "Moon Usage Key",
      "rate_limit": 15,
      "role": "user",
      "updated_at": "2026-04-22T00:55:37Z"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Device Key Lists Granted Collections

List collections with the device API key. The response should be limited to granted collections.

```bash
curl -s -X GET "http://localhost:6000/collections:query" \
    -H "Authorization: Bearer $API_KEY" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Collections retrieved successfully",
  "data": [
    {
      "count": 0,
      "name": "apikey_usage_devices",
      "system": false
    }
  ],
  "meta": {
    "count": 1,
    "current_page": 1,
    "per_page": 15,
    "total": 1,
    "total_pages": 1
  },
  "links": {
    "first": "/collections:query?page=1&per_page=15",
    "last": "/collections:query?page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```

### Device Key Reads Granted Data

Read data from the collection granted to the device API key.

```bash
curl -s -X GET "http://localhost:6000/data/apikey_usage_devices:query" \
    -H "Authorization: Bearer $API_KEY" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "meta": {
    "count": 0,
    "current_page": 1,
    "per_page": 15,
    "total": 0,
    "total_pages": 1
  },
  "links": {
    "first": "/data/apikey_usage_devices:query?page=1&per_page=15",
    "last": "/data/apikey_usage_devices:query?page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```

### Device Key Rejects Unlisted Collection

Attempt to access a collection that is not in the device key allowlist.

```bash
curl -s -X GET "http://localhost:6000/collections:query?name=apikey_usage_pages" \
    -H "Authorization: Bearer $API_KEY" | jq .
```

**Response (403 Forbidden):**

```json
{
  "message": "Forbidden"
}
```

### Device Key Rejects Unlisted Data

Attempt to read data from a collection that is not in the device key allowlist.

```bash
curl -s -X GET "http://localhost:6000/data/apikey_usage_pages:query" \
    -H "Authorization: Bearer $API_KEY" | jq .
```

**Response (403 Forbidden):**

```json
{
  "message": "Forbidden"
}
```

### Convert to Website API Key

Update the same API key to website mode and move access to the website collection.

```bash
curl -s -X POST "http://localhost:6000/data/apikeys:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "update",
        "data": [
          {
            "id": "01KPSAYYC4QHKMC6B47313B9W1",
            "name": "Moon Usage Key",
            "role": "user",
            "can_write": false,
            "collections": [
              "apikey_usage_pages"
            ],
            "is_website": true,
            "allowed_origins": [
              "https://moon.devnodes.in"
            ],
            "rate_limit": 10,
            "captcha_required": false,
            "enabled": true
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource updated successfully",
  "data": [
    {
      "allowed_origins": [
        "https://moon.devnodes.in"
      ],
      "can_write": false,
      "captcha_required": false,
      "collections": [
        "apikey_usage_pages"
      ],
      "created_at": "2026-04-22T00:55:37Z",
      "enabled": true,
      "id": "01KPSAYYC4QHKMC6B47313B9W1",
      "is_website": true,
      "last_used_at": "2026-04-22T00:55:38Z",
      "name": "Moon Usage Key",
      "rate_limit": 10,
      "role": "user",
      "updated_at": "2026-04-22T00:55:38Z"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Rotate Website API Key

Rotate the website API key to capture the fresh website credential for subsequent requests.

```bash
curl -s -X POST "http://localhost:6000/data/apikeys:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "action",
        "action": "rotate",
        "data": [
          {
            "id": "01KPSAYYC4QHKMC6B47313B9W1"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Action completed successfully",
  "data": [
    {
      "allowed_origins": [
        "https://moon.devnodes.in"
      ],
      "can_write": false,
      "captcha_required": false,
      "collections": [
        "apikey_usage_pages"
      ],
      "enabled": true,
      "id": "01KPSAYYC4QHKMC6B47313B9W1",
      "is_website": true,
      "key": "$API_KEY",
      "name": "Moon Usage Key",
      "rate_limit": 10,
      "role": "user"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Website Key Reads Granted Collection

Read an allowed collection with the website API key and a matching `Origin` header.

```bash
curl -s -X GET "http://localhost:6000/collections:query?name=apikey_usage_pages" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Origin: https://moon.devnodes.in" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Collection retrieved successfully",
  "data": [
    {
      "count": 0,
      "name": "apikey_usage_pages",
      "system": false
    }
  ]
}
```

### Website Key Reads Granted Data

Read data with the website API key and a matching `Origin` header.

```bash
curl -s -X GET "http://localhost:6000/data/apikey_usage_pages:query" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Origin: https://moon.devnodes.in" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "meta": {
    "count": 0,
    "current_page": 1,
    "per_page": 15,
    "total": 0,
    "total_pages": 1
  },
  "links": {
    "first": "/data/apikey_usage_pages:query?page=1&per_page=15",
    "last": "/data/apikey_usage_pages:query?page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```

### Website Key Rejects Missing Origin

Website API keys must reject requests without a matching `Origin` header.

```bash
curl -s -X GET "http://localhost:6000/data/apikey_usage_pages:query" \
    -H "Authorization: Bearer $API_KEY" | jq .
```

**Response (403 Forbidden):**

```json
{
  "message": "Forbidden"
}
```

### Website Key Rejects Unlisted Collection

Attempt to access a collection that is no longer granted to the website API key.

```bash
curl -s -X GET "http://localhost:6000/collections:query?name=apikey_usage_devices" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Origin: https://moon.devnodes.in" | jq .
```

**Response (403 Forbidden):**

```json
{
  "message": "Forbidden"
}
```

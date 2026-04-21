### Create API Key

Create a new API key.

```bash
curl -s -X POST "http://localhost:6000/data/apikeys:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "create",
        "data": [
          {
            "name": "Integration Service",
            "role": "user",
            "can_write": false,
            "collections": [
              "products",
              "orders"
            ],
            "is_website": true,
            "allowed_origins": [
              "https://moon.devnodes.in"
            ],
            "rate_limit": 5,
            "captcha_required": true,
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
      "allowed_origins": [
        "https://moon.devnodes.in"
      ],
      "can_write": false,
      "captcha_required": true,
      "collections": [
        "products",
        "orders"
      ],
      "created_at": "2026-04-21T14:58:39Z",
      "enabled": true,
      "id": "01KPR8SVKNGXSCC3EY1DB6RSRZ",
      "is_website": true,
      "key": "moon_live_YQh3eBJNm0AWUr8L9LXDbTX7Y3wP3qGZJ4f5h3MG2mfICMBVpHgQwbC4QwYH0RL0",
      "name": "Integration Service",
      "rate_limit": 5,
      "role": "user",
      "updated_at": "2026-04-21T14:58:39Z"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### List API Keys

Retrieve all API keys.

```bash
curl -s -X GET "http://localhost:6000/data/apikeys:query" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "allowed_origins": [
        "https://moon.devnodes.in"
      ],
      "can_write": false,
      "captcha_required": true,
      "collections": [
        "products",
        "orders"
      ],
      "created_at": "2026-04-21T14:58:39Z",
      "enabled": true,
      "id": "01KPR8SVKNGXSCC3EY1DB6RSRZ",
      "is_website": true,
      "last_used_at": null,
      "name": "Integration Service",
      "rate_limit": 5,
      "role": "user",
      "updated_at": "2026-04-21T14:58:39Z"
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
    "first": "/data/apikeys:query?page=1&per_page=15",
    "last": "/data/apikeys:query?page=1&per_page=15",
    "next": null,
    "prev": null
  }
}
```

### Get API Key by ID

Retrieve a specific API key by its ULID.

```bash
curl -s -X GET "http://localhost:6000/data/apikeys:query?id=01KPR8SVKNGXSCC3EY1DB6RSRZ" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource retrieved successfully",
  "data": [
    {
      "allowed_origins": [
        "https://moon.devnodes.in"
      ],
      "can_write": false,
      "captcha_required": true,
      "collections": [
        "products",
        "orders"
      ],
      "created_at": "2026-04-21T14:58:39Z",
      "enabled": true,
      "id": "01KPR8SVKNGXSCC3EY1DB6RSRZ",
      "is_website": true,
      "last_used_at": null,
      "name": "Integration Service",
      "rate_limit": 5,
      "role": "user",
      "updated_at": "2026-04-21T14:58:39Z"
    }
  ]
}
```

### Update API Key Metadata

Update API key metadata (name, description) without changing the key itself.

```bash
curl -s -X POST "http://localhost:6000/data/apikeys:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "update",
        "data": [
          {
            "id": "01KPR8SVKNGXSCC3EY1DB6RSRZ",
            "name": "Updated Integration Service",
            "collections": [
              "products"
            ],
            "allowed_origins": [
              "https://moon.devnodes.in",
              "https://www.moon.devnodes.in"
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
        "https://moon.devnodes.in",
        "https://www.moon.devnodes.in"
      ],
      "can_write": false,
      "captcha_required": false,
      "collections": [
        "products"
      ],
      "created_at": "2026-04-21T14:58:39Z",
      "enabled": true,
      "id": "01KPR8SVKNGXSCC3EY1DB6RSRZ",
      "is_website": true,
      "last_used_at": null,
      "name": "Updated Integration Service",
      "rate_limit": 10,
      "role": "user",
      "updated_at": "2026-04-21T14:58:40Z"
    }
  ],
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

### Rotate API Key

Generate a new key value and invalidate the old one in a single step.

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
            "id": "01KPR8SVKNGXSCC3EY1DB6RSRZ"
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
        "https://moon.devnodes.in",
        "https://www.moon.devnodes.in"
      ],
      "can_write": false,
      "captcha_required": false,
      "collections": [
        "products"
      ],
      "enabled": true,
      "id": "01KPR8SVKNGXSCC3EY1DB6RSRZ",
      "is_website": true,
      "key": "moon_live_FcxzPdRuR3l31Fd6Kh9OUtaJ9NAZ2kYl9XdwGk2AOvEDLKl8NnXI8QmDLlvoNzTD",
      "name": "Updated Integration Service",
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

### Delete API Key

Permanently delete an API key.

```bash
curl -s -X POST "http://localhost:6000/data/apikeys:mutate" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "destroy",
        "data": [
          {
            "id": "01KPR8SVKNGXSCC3EY1DB6RSRZ"
          }
        ]
      }
    ' | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource destroyed successfully",
  "meta": {
    "failed": 0,
    "success": 1
  }
}
```

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
            "can_write": false
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
      "can_write": false,
      "created_at": "2026-03-08T16:38:47Z",
      "id": "01KK751JJD1GR629JX7RF2Q1N2",
      "key": "moon_live_6PL5p63vTbGTF9fss4DUj10vusx0oS0R4a01l4P8qeLhmEa9WF20OhbtKhaBatV9",
      "name": "Integration Service",
      "role": "user",
      "updated_at": "2026-03-08T16:38:47Z"
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
      "can_write": false,
      "created_at": "2026-03-08T16:38:47Z",
      "id": "01KK751JJD1GR629JX7RF2Q1N2",
      "last_used_at": null,
      "name": "Integration Service",
      "role": "user",
      "updated_at": "2026-03-08T16:38:47Z"
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
curl -s -X GET "http://localhost:6000/data/apikeys:query?id=01KK751JJD1GR629JX7RF2Q1N2" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (200 OK):**

```json
{
  "message": "Resource retrieved successfully",
  "data": [
    {
      "can_write": false,
      "created_at": "2026-03-08T16:38:47Z",
      "id": "01KK751JJD1GR629JX7RF2Q1N2",
      "last_used_at": null,
      "name": "Integration Service",
      "role": "user",
      "updated_at": "2026-03-08T16:38:47Z"
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
            "id": "01KK751JJD1GR629JX7RF2Q1N2",
            "name": "Updated Integration Service"
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
      "can_write": false,
      "created_at": "2026-03-08T16:38:47Z",
      "id": "01KK751JJD1GR629JX7RF2Q1N2",
      "last_used_at": null,
      "name": "Updated Integration Service",
      "role": "user",
      "updated_at": "2026-03-08T16:38:48Z"
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
            "id": "01KK751JJD1GR629JX7RF2Q1N2"
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
      "can_write": false,
      "id": "01KK751JJD1GR629JX7RF2Q1N2",
      "key": "moon_live_55XwTPU7On334v2DAXhlnlSrjKq9BPUhTASgomEB79ULKqFPxG1ZX2ZsBGAJXt23",
      "name": "Updated Integration Service",
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
            "id": "01KK751JJD1GR629JX7RF2Q1N2"
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

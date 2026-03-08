### Health Check

Returns server health status. No authentication required.

```bash
curl -s -X GET "http://localhost:6000/health" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "moon": "1.00",
    "timestamp": "2026-03-08T16:38:31Z"
  }
}
```

### Root Endpoint

`/` is an alias for `/health`.

```bash
curl -s -X GET "http://localhost:6000/" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "moon": "1.00",
    "timestamp": "2026-03-08T16:38:31Z"
  }
}
```

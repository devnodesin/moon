### Check Health

```bash
curl -s -X GET "http://localhost:6006/health" | jq .
```

The root path `GET /` is an alias for `GET /health` when no URL prefix is configured.

**Response (200 OK):**

```json
{
  "data": {
    "moon": "1.99",
    "timestamp": "2026-02-27T18:21:43Z"
  }
}
```

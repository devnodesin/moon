### Check Health

```bash
curl -s -X GET "http://localhost:6006/health" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "moon": "1.99",
    "timestamp": "2026-03-01T12:48:49Z"
  }
}
```

### Get Root

`/` is alias for `/health`

```bash
curl -s -X GET "http://localhost:6006/" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "moon": "1.99",
    "timestamp": "2026-03-01T12:48:49Z"
  }
}
```

### Check Health

```bash
curl -s -X GET "http://localhost:6006/health" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "moon": "1.0.0",
    "timestamp": "2026-02-03T13:58:53Z"
  }
}
```

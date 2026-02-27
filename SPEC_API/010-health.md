### Check Health

```bash
curl -s -X GET "http://localhost:6006/health" | jq .
```

**Response (200 OK):**

```json
{
  "data": {
    "moon": "1.99",
    "timestamp": "2026-02-27T18:21:43Z"
  }
}
```

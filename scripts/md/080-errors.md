### HTTP 400: Missing Required Fields

**Error:** `400 Bad Request` — Returned when required fields are missing or the request is malformed.

```bash
curl -s -X POST "http://localhost:6000/auth:session" \
    -H "Content-Type: application/json" \
    -d '
      {
        "op": "login",
        "data": {
          "username": "",
          "password": ""
        }
      }
    ' | jq .
```

**Response (400 Bad Request):**

```json
{
  "message": "Missing required field: data.username"
}
```

### HTTP 401: Authentication Required

**Error:** `401 Unauthorized` — Returned when no Authorization header is provided on a protected endpoint.

```bash
curl -s -X GET "http://localhost:6000/auth:me" | jq .
```

**Response (401 Unauthorized):**

```json
{
  "message": "Unauthorized"
}
```

### HTTP 401: Invalid Token

**Error:** `401 Unauthorized` — Returned when the provided token is invalid or expired.

```bash
curl -s -X GET "http://localhost:6000/data/users:query" \
    -H "Authorization: Bearer invalid_token_value" | jq .
```

**Response (401 Unauthorized):**

```json
{
  "message": "Unauthorized"
}
```

### HTTP 404: Resource Not Found

**Error:** `404 Not Found` — Returned when the requested resource does not exist.

```bash
curl -s -X GET "http://localhost:6000/data/users:query?id=01ZZZZZZZZZZZZZZZZZZZZZZZ0" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (404 Not Found):**

```json
{
  "message": "Resource not found"
}
```

### HTTP 404: Collection Not Found

**Error:** `404 Not Found` — Returned when the requested collection does not exist.

```bash
curl -s -X GET "http://localhost:6000/collections:query?name=nonexistentcollection" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (404 Not Found):**

```json
{
  "message": "Collection 'nonexistentcollection' not found"
}
```

### HTTP 405: Method Not Allowed

**Error:** `405 Method Not Allowed` — Returned when an unsupported HTTP method is used.

```bash
curl -s -X DELETE "http://localhost:6000/health" | jq .
```

**Response (405 Method Not Allowed):**

```json
{
  "message": "Method not allowed"
}
```

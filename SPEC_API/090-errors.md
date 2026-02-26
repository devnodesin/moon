### HTTP 400: Missing Required Fields

**Error:** `400 Bad Request` — Returned when required fields are missing or the request is malformed.

```bash
curl -s -X POST "http://localhost:6006/auth:login" \
    -H "Content-Type: application/json" \
    -d '
      {
        "username": "",
        "password": ""
      }
    ' | jq .
```

**Response (400 Bad Request):**

```json
{
  "message": "username and password are required"
}
```

### HTTP 400: Validation Error

**Error:** `400 Bad Request` — Returned when field values fail validation (e.g., invalid email format).

```bash
curl -s -X POST "http://localhost:6006/users:create" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '
      {
        "data": {
          "username": "baduser",
          "email": "not-a-valid-email",
          "password": "Pass123#",
          "role": "user"
        }
      }
    ' | jq .
```

**Response (400 Bad Request):**

```json
{
  "message": "invalid email format"
}
```

### HTTP 401: Authentication Required

**Error:** `401 Unauthorized` — Returned when no Authorization header is provided on a protected endpoint.

```bash
curl -s -X GET "http://localhost:6006/auth:me" | jq .
```

**Response (401 Unauthorized):**

```json
{
  "message": "authentication required"
}
```

### HTTP 401: Invalid Token

**Error:** `401 Unauthorized` — Returned when the provided token is invalid or expired.

```bash
curl -s -X GET "http://localhost:6006/users:list" \
    -H "Authorization: Bearer invalid_token_value" | jq .
```

**Response (401 Unauthorized):**

```json
{
  "message": "invalid or expired token"
}
```

### HTTP 404: User Not Found

**Error:** `404 Not Found` — Returned when the requested resource does not exist.

```bash
curl -s -X GET "http://localhost:6006/users:get?id=01ZZZZZZZZZZZZZZZZZZZZZZZ0" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (404 Not Found):**

```json
{
  "message": "user with id '01ZZZZZZZZZZZZZZZZZZZZZZZ0' not found"
}
```

### HTTP 404: Collection Not Found

**Error:** `404 Not Found` — Returned when the requested collection does not exist.

```bash
curl -s -X GET "http://localhost:6006/collections:get?name=nonexistentcollection" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Response (404 Not Found):**

```json
{
  "message": "collection 'nonexistentcollection' not found"
}
```

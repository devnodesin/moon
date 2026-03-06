## Standard Error Response

The API uses a simple, consistent error handling approach and strictly follows standard HTTP semantics.

- `200`: OK – Successful GET request |
- `201`: Created – Successful POST request creating resource |
- `400`: Invalid request (validation error, invalid parameter, malformed request)
- `401`: Authentication required
- `403`: Forbidden
- `404`: Resource not found
- `429`: Too Many Requests
- `500`: Server error
- Only the codes listed above are permitted; do not use any others.

- Errors are indicated by standard HTTP status codes (for machines).
- Each error response includes only a single `message` field (for humans), intended for direct display to users.
- No internal error codes or additional error metadata are used.
- The HTTP status code is the only machine-readable error signal.
- Clients are not expected to parse or branch on error types.

When an error occurs, the API responds with the appropriate HTTP status code and a JSON body:

```json
{
  "message": "A human-readable description of the error"
}
```


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
  "message": "invalid token format"
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
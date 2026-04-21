### Standard Error Response

Moon uses a single error body shape across all endpoints:

```json
{
  "message": "A human-readable description of the error"
}
```

Rules:

- Error bodies contain only `message`.
- No error codes, validation maps, or extra metadata are allowed.
- The HTTP status code is the only machine-readable error signal.
- Clients must not expect structured error codes or error metadata.
- Documented exception: CAPTCHA challenges use `message` plus a `captcha` object.

Rate-limit rule:

- `429` guarantees only the standard error body.
- No rate-limit response headers are guaranteed unless a future specification adds them.

CAPTCHA challenge rule:

- `403` may return a CAPTCHA challenge body when an authenticated API key requires CAPTCHA validation on `POST` requests.

### Error Status Codes

Moon uses only these error statuses for API errors:

| Status | Meaning |
| ------ | ------- |
| `400 Bad Request` | The request is malformed or fails validation |
| `401 Unauthorized` | Authentication is missing, invalid, expired, or revoked |
| `403 Forbidden` | Authentication succeeded but the caller is not allowed to perform the operation |
| `404 Not Found` | The requested endpoint target, collection, or record does not exist |
| `405 Method Not Allowed` | The HTTP method is not supported for the route |
| `429 Too Many Requests` | The caller exceeded a rate limit |
| `500 Internal Server Error` | The server failed to complete a valid request |

### Error Examples

#### 400 Bad Request

Example request:

`POST /auth:session`

```json
{
  "op": "signin",
  "data": {
    "username": "newuser",
    "password": "UserPass123"
  }
}
```

Example response:

```json
{
  "message": "invalid session operation"
}
```

#### 401 Unauthorized

Example request:

`GET /auth:me`

No `Authorization` header.

Example response:

```json
{
  "message": "authentication required"
}
```

#### 403 Forbidden

Example request:

`POST /collections:mutate`

```json
{
  "op": "create",
  "data": [
    {
      "name": "products",
      "columns": [
        { "name": "title", "type": "string" }
      ]
    }
  ]
}
```

Authenticated as a non-admin user.

Example response:

```json
{
  "message": "forbidden"
}
```

CAPTCHA challenge response:

```json
{
  "message": "Captcha required",
  "captcha": {
    "id": "01KTESTCAPTCHA1234567890AB",
    "image_base64": "PHN2ZyB4bWxucz0iLi4uIj48L3N2Zz4=",
    "expires_in": 300
  }
}
```

#### 404 Not Found

Example request:

`GET /collections:query?name=missing_collection`

Example response:

```json
{
  "message": "collection 'missing_collection' not found"
}
```

Example request:

`GET /data/products:query?id=01ZZZZZZZZZZZZZZZZZZZZZZZ0`

Example response:

```json
{
  "message": "record with id '01ZZZZZZZZZZZZZZZZZZZZZZZ0' not found"
}
```

#### 405 Method Not Allowed

Example request:

`DELETE /data/products:query`

Example response:

```json
{
  "message": "method not allowed"
}
```

#### 429 Too Many Requests

Example response:

```json
{
  "message": "too many requests"
}
```

Rate-limit note:

- `429` guarantees only the standard error body.
- No rate-limit response headers are guaranteed unless they are defined in a future specification.

#### 500 Internal Server Error

Example response:

```json
{
  "message": "internal server error"
}
```

### Message Rules

- Messages must be concise and human-readable.
- Messages must not expose secrets or internal implementation details.
- Authentication errors must not reveal sensitive credential state beyond what is necessary for client handling.

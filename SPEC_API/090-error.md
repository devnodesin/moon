## Standard Error Response

The API uses a simple, consistent error handling approach and strictly follows standard HTTP semantics.

- `200`: OK – Successful GET request |
- `201`: Created – Successful POST request creating resource |
- `400`: Invalid request (validation error, invalid parameter, malformed request)
- `401`: Authentication required
- `404`: Resource not found
- `500`: Server error
 - Only the codes listed above are permitted; do not use any others.

Note: `403` (Forbidden) is intentionally omitted in this specification to keep the error surface small. Authorization or permission failures should be handled via `401` per this document. If an implementation needs to distinguish "authenticated but not allowed" cases, add `403` and follow the same single-`message` JSON body pattern.
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

### Examples

**HTTP 400: Validation Error**

```json
{
  "message": "Email format is invalid"
}
```

**HTTP 401: Unauthorized**

```json
{
  "message": "Authentication required"
}
```

**HTTP 404: Not Found**

```json
{
  "message": "User with id '01KHCZGWWRBQBREMG0K23C6C5H' not found"
}
```

**HTTP 500: Server Error**

```json
{
  "message": "An unexpected error occurred"
}
```

## Public Endpoints

Health and documentation endpoints are public; no authentication is required. All other endpoints require authentication.

### API Documentation

API documentation is available in multiple formats:

- HTML: `GET /doc/`
- Markdown: `GET /doc/llms.md`
- Plain Text: `GET /doc/llms.txt` (alias for `/doc/llms.md`)
- JSON: `GET /doc/llms.json`

### Health Endpoint

- `GET /health`: Returns API service health and version information.

**Response (200 OK):**

```json
{
  "data": {
    "name": "moon",
    "status": "live",
    "version": "1.0"
  }
}
```

**Error Response:** Follow [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling

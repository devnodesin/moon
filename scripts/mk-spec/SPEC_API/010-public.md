
## Documentation and Health Endpoints

### Documentation Endpoints

Access API documentation in multiple formats.

#### View HTML Documentation

`GET /doc/`

View interactive HTML documentation in browser.

**URL:** [http://localhost:6006/doc/](http://localhost:6006/doc/)

---

#### Get Markdown Documentation

`GET /doc/llms.md`

Retrieve documentation in Markdown format (for humans and AI coding agents).

**Response (200 OK):**

Returns Markdown-formatted documentation.

**URL:** [http://localhost:6006/doc/llms.md](http://localhost:6006/doc/llms.md)

---

#### Get Text Documentation

`GET /doc/llms.txt`

Retrieve documentation in plain text format.

**Response (200 OK):**

Returns plain text documentation.

---

#### Get JSON Schema

`GET /doc/llms.json`

Retrieve machine-readable API schema in JSON format.

**Response (200 OK):**

```json
{
  "data": {
    "version": "1.0",
    "endpoints": [...],
    "schemas": {...}
  }
}
```

---

#### Refresh Documentation Cache

`POST /doc:refresh`

Force refresh of the documentation cache.

**Headers:**

- `Authorization: Bearer {access_token}` (required)

**Response (200 OK):**

```json
{
  "message": "Documentation cache refreshed successfully"
}
```

---

### Health Check Endpoint

`GET /health`

Check API service health and version information.

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

### Important Notes

- **Documentation formats**: Available in HTML (interactive), Markdown (human/AI readable), plain text, and JSON (machine-readable)
- **Cache refresh**: Documentation is cached for performance. Use `/doc:refresh` after configuration changes or schema updates
- **Health check**: No authentication required. Use for monitoring and uptime checks
- **Version tracking**: The version field in health response indicates the current API version

**Error Response:** For details on error handling, see [Error Response](#error-response).

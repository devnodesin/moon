# PRD-054-schema-endpoint-refactor.md

## 1. Overview

### Problem Statement
Schema retrieval is currently supported via a `?schema` query parameter on collection endpoints. This approach is non-standard, complicates routing, and introduces legacy code paths. There is a need to migrate schema retrieval to a dedicated, RESTful endpoint: `GET /{collection}:schema`, and to remove all legacy support for the old query parameter.

### Context and Background
The current implementation allows clients to retrieve a collection's schema by appending `?schema` to the collection endpoint. This pattern is inconsistent with modern API design and complicates handler logic, documentation, and testing. A dedicated endpoint will simplify the codebase, improve clarity, and align with best practices.

### High-Level Solution Summary
- Remove all code, tests, and documentation related to the `?schema` query parameter.
- Implement schema retrieval exclusively at `GET /{collection}:schema`.
- Update all relevant documentation and templates.
- No backward compatibility is required.

## 2. Requirements

### Functional Requirements
- The API must expose a dedicated endpoint: `GET /{collection}:schema` for schema retrieval.
- The legacy `?schema` query parameter must be fully removed from code, tests, and documentation.
- The new endpoint must return the schema in the following format:

```json
{
  "collection": "<collection_name>",
  "fields": [
    {
      "name": "<field_name>",
      "type": "<type>",
      "nullable": <bool>,
      // optional: "system": <bool>, "unique": <bool>
    },
    ...
  ]
}
```

- The endpoint must require authentication (e.g., Bearer token).
- The endpoint must return appropriate error responses for:
  - Unauthorized access (401)
  - Collection not found (404)
  - Internal server errors (500)

### Technical Requirements
- All routing and handler logic must be refactored to remove support for the `?schema` query parameter.
- Tests must be updated or added to cover only the new endpoint and its output.
- Documentation must be updated:
  - `SPEC.md` must describe the new endpoint and remove all references to the old pattern.
  - `cmd/moon/internal/handlers/templates/doc.md.tmpl` must reflect the new endpoint only.
- No code or documentation may reference the `?schema` query parameter after this change.
- No backward compatibility or migration logic is required.

### API Specifications
- **Endpoint:** `GET /{collection}:schema`
- **Auth:** Required (Bearer token)
- **Request Parameters:**
  - `collection` (path): Name of the collection
- **Response:**
  - 200 OK: Schema object as above
  - 401 Unauthorized: If auth is missing/invalid
  - 404 Not Found: If collection does not exist
  - 500 Internal Server Error: On unexpected errors

### Validation Rules and Constraints
- Only valid, existing collections may be queried
- The endpoint must not accept or process the `?schema` query parameter

### Error Handling and Failure Modes
- 401 if no or invalid auth
- 404 if collection does not exist
- 500 for all other errors
- Error responses must be JSON and include a clear `error` message

### Filtering, Sorting, Permissions, Limits
- No filtering or sorting is required for this endpoint
- Permissions are enforced via authentication
- No pagination or limits apply

## 3. Acceptance Criteria

### Verification Steps
- [ ] All code and tests referencing the `?schema` query parameter are removed
- [ ] `GET /{collection}:schema` returns the correct schema for valid collections
- [ ] The endpoint requires valid authentication
- [ ] The endpoint returns 404 for non-existent collections
- [ ] The endpoint returns 401 for missing/invalid auth
- [ ] The endpoint returns 500 for internal errors
- [ ] All documentation and templates reference only the new endpoint

### Test Scenarios
- Request schema for an existing collection with valid auth → 200 + schema JSON
- Request schema for a non-existent collection → 404 + error JSON
- Request schema with missing/invalid auth → 401 + error JSON
- Attempt to use `?schema` query parameter → 404 or no effect (no legacy support)

### Expected API Responses
- **Success:**
```json
{
  "collection": "todos",
  "fields": [
    { "name": "id", "type": "string", "nullable": false, "system": true },
    { "name": "title", "type": "string", "nullable": false, "unique": true },
    { "name": "due_date", "type": "datetime", "nullable": true },
    { "name": "priority", "type": "integer", "nullable": false }
  ]
}
```
- **Error:**
```json
{
  "error": "Collection not found"
}
```

### Edge Cases and Negative Paths
- Requesting schema for a collection that does not exist
- Requesting schema with invalid or missing authentication
- Attempting to use the `?schema` query parameter (should not be supported)

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

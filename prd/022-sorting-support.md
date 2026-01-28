## Overview

- Add dynamic sorting to the data list endpoint
- Currently, sorting is hardcoded to `ORDER BY ulid ASC` with no client control
- Enable clients to specify sort order via query parameters
- Support syntax like `?sort=price` (ascending) or `?sort=-price` (descending)
- Support multiple sort fields (e.g., `?sort=-created_at,name`)

## Requirements

- Parse `sort` query parameter to determine sort fields and directions
- Support ascending sort: `?sort=column_name` or `?sort=+column_name`
- Support descending sort: `?sort=-column_name`
- Support multiple sort fields: `?sort=field1,-field2` (comma-separated)
- Validate that sort columns exist in the collection schema
- Default to `ORDER BY ulid ASC` when no sort parameter is provided
- Return 400 Bad Request for invalid column names in sort parameter
- Update `query.Builder` to accept order by clauses from the handler
- Add tests for single field sort, multiple field sort, and mixed directions

## Acceptance

- Clients can sort collections using the `sort` query parameter
- Ascending and descending directions work correctly
- Multiple sort fields are applied in the correct order
- Invalid column names in sort parameter return 400 Bad Request
- Default sorting behavior (ulid ASC) is preserved when no sort parameter given
- Tests cover all sort variations and error cases
- API documentation includes sort syntax examples

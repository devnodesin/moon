## Overview

- Add field selection capability to reduce payload size and improve performance
- Currently, all endpoints use `SELECT *` which returns all columns
- Enable clients to request only specific fields via query parameter
- Support syntax like `?fields=id,name,price` to select subset of columns
- Particularly valuable for mobile clients and large tables with many columns

## Requirements

- Parse `fields` query parameter to determine requested columns
- Support comma-separated field list: `?fields=id,name,price`
- Validate that requested fields exist in the collection schema
- Always include `ulid` field even if not explicitly requested (for cursor pagination)
- Update `query.Builder.Select` to accept a field list instead of defaulting to `*`
- Return 400 Bad Request when invalid field names are provided
- Default to all fields (`SELECT *`) when no `fields` parameter is given
- Work correctly with filters, sorting, and search parameters
- Add tests for field selection with various combinations

## Acceptance

- Clients can request specific fields using the `?fields=` parameter
- Only requested fields are returned in the response
- The `ulid` field is always included for pagination consistency
- Invalid field names return 400 Bad Request with clear error messages
- Field selection works correctly with filters, sorting, and search
- Default behavior (all fields) is preserved when no fields parameter given
- Tests cover field selection alone and combined with other features
- API documentation includes field selection examples

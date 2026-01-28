## Overview

- Add comprehensive filtering capabilities to the data list endpoint
- Currently, the `DataListRequest` has a `Filter` field that is completely ignored by the handler
- Enable clients to filter data using query parameters with standard operators
- Support syntax like `?price[gt]=100`, `?name[like]=moon`, `?status[eq]=active`
- Replace manual SQL construction in handlers with the query builder

## Requirements

- Parse filter query parameters from the URL (e.g., `?column[operator]=value`)
- Support multiple filters combined with AND logic (e.g., `?price[gt]=100&category[eq]=electronics`)
- Validate that filter columns exist in the collection schema before applying
- Convert query parameter filters into `query.Condition` structs
- Update `DataHandler.List` to use `query.Builder` instead of manual `fmt.Sprintf` SQL
- Return meaningful error messages when invalid columns or operators are used
- Support filtering on all data types (strings, numbers, booleans, dates)
- Document filter syntax in API documentation
- Add tests for various filter combinations

## Acceptance

- Clients can filter collections using query parameters with operators
- Invalid column names return 400 Bad Request with clear error messages
- Invalid operators return 400 Bad Request with clear error messages
- Multiple filters work correctly with AND logic
- All manual SQL construction in `DataHandler.List` is replaced with query builder
- Tests cover single filters, multiple filters, and error cases
- API documentation includes filter syntax examples

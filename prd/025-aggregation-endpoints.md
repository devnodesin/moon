## Overview

- Add server-side aggregation endpoints for analytics and dashboards
- Provide custom-action endpoints per collection for `count`, `sum`, `avg`, `min`, and `max`
- Avoid client-side aggregation that requires fetching full datasets
- Return only the computed aggregate value to keep responses fast and bandwidth-efficient

## Requirements

- Add the following aggregation endpoints for any existing collection:
  - `GET /api/v1/{collection}:count`
  - `GET /api/v1/{collection}:sum?field={field}`
  - `GET /api/v1/{collection}:avg?field={field}`
  - `GET /api/v1/{collection}:min?field={field}`
  - `GET /api/v1/{collection}:max?field={field}`

- Request parameters
  - `collection` (path): required, the target collection name
  - `field` (query):
    - Required for `:sum`, `:avg`, `:min`, `:max`
    - Not used for `:count` (if provided, it is ignored)

- Filtering support
  - All aggregation endpoints must support filtering via query parameters so aggregates can be computed over a subset of records
  - Filtering semantics and operators must match the existing list filtering conventions:
    - Syntax: `?column[operator]=value`
    - Operators: `eq`, `ne`, `gt`, `lt`, `gte`, `lte`, `like`, `in`
    - Multiple filters are combined with AND semantics

- Server-side execution and performance
  - Aggregation must be computed in the database (e.g., SQL aggregate functions) and must not fetch and iterate over the full dataset in application memory
  - Aggregation queries must apply filters at the database level (via `WHERE` clauses) before aggregation

- Validation
  - Validate that the collection exists (using the schema registry)
  - Validate that `field` exists in the collection schema for endpoints that require it
  - Validate that `field` is a numeric type for `:sum` and `:avg`
  - Validate that `field` is a numeric type for `:min` and `:max`
  - Needs Clarification: whether `:min`/`:max` should support non-numeric fields (e.g., strings, timestamps). This PRD currently requires numeric-only to keep behavior unambiguous.

- Response shape
  - Each endpoint must return only the computed value in a minimal JSON payload:
    - `200 OK` with body: `{ "value": <number> }`
  - `:count` returns an integer value
  - `:sum`, `:avg`, `:min`, `:max` return a JSON number (implementation may use floating-point internally)
  - Needs Clarification: numeric precision/rounding requirements for `:avg` and `:sum` (e.g., decimals for currency). Default behavior is to return the database-computed numeric result without additional rounding.

- Error handling
  - Errors must use the standard error response format (code, message, details, request_id)
  - Missing `field` for `:sum`/`:avg`/`:min`/`:max` returns `400 Bad Request`
  - Unknown collection returns `404 Not Found`
  - Unknown field returns `400 Bad Request`
  - Non-numeric field used for numeric-only aggregations returns `400 Bad Request`
  - Unsupported aggregation action (anything other than `count|sum|avg|min|max`) returns `404 Not Found`

- OpenAPI
  - Dynamic OpenAPI generation must include the aggregation endpoints for each collection
  - OpenAPI must document:
    - Required/optional query parameters (`field`, filtering params)
    - Response schema `{ value: number }`
    - Error response schema
    - Example requests for each aggregation type

- Tests
  - Add automated tests covering:
    - Successful aggregation for each endpoint
    - Filtering behavior (e.g., `total[gt]=...` affects aggregate)
    - Field validation errors
    - Collection-not-found errors
    - Numeric-only validation errors

- Sample test script
  - Add `samples/test_scripts/aggregation.sh` that:
    - Creates a collection named `orders` with fields:
      - `order_id` (string)
      - `total` (float)
      - `products` (json)
      - `subtotal` (float)
      - `tax` (float)
      - `customer_name` (string)
      - Any additional fields needed for realism
    - Inserts 10 dummy `orders` records with varied values for `total`, `subtotal`, `tax`, and `products`
    - Retrieves and prints all orders to verify insertion
    - Demonstrates all aggregation endpoints on `orders`:
      - `GET /api/v1/orders:count`
      - `GET /api/v1/orders:sum?field=total`
      - `GET /api/v1/orders:avg?field=total`
      - `GET /api/v1/orders:min?field=total`
      - `GET /api/v1/orders:max?field=total`
    - Optionally demonstrates aggregations on other numeric fields like `tax` or `subtotal`
    - Prints each aggregation result clearly for review

## Acceptance

- `GET /api/v1/{collection}:count` returns `{ "value": <integer> }` and does not return dataset rows
- `GET /api/v1/{collection}:sum|avg|min|max?field=...` returns `{ "value": <number> }` and does not return dataset rows
- Aggregation endpoints support filtering via query parameters and apply filters before computing aggregates
- Aggregation is computed by the database (no full-table scans into application memory)
- Invalid requests return correct status codes and standard error format:
  - Missing field -> 400
  - Unknown field -> 400
  - Non-numeric field -> 400
  - Unknown collection -> 404
  - Unsupported action -> 404
- OpenAPI documentation includes aggregation endpoints with parameters, response schema, errors, and examples
- Automated tests cover success paths and negative paths for all aggregation types
- `samples/test_scripts/aggregation.sh` creates `orders`, inserts 10 rows, prints them, and demonstrates all aggregation endpoints with clearly printed results
- Updated the scripts, documentations and specifications.
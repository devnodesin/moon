# Standard API Response Patterns

This document describes the standard response patterns, query options, and aggregation operations for the Moon API. All endpoints follow consistent conventions for success and error responses.

## Public Endpoints

Health and documentation endpoints are accessible without authentication. All other endpoints require authentication.

- **Health Endpoint:** See [Health Endpoint](./010-health.md).
- **API Documentation:**
  - HTML: `GET /doc/`
  - Markdown: `GET /doc/llms.md`
  - Plain Text: `GET /doc/llms.txt` (alias for `/doc/llms.md`)
  - JSON: `GET /doc/llms.json`

## Authentication

| Endpoint       | Method | Description                                 |
|----------------|--------|---------------------------------------------|
| `/auth:login`    | POST   | Authenticate user, receive tokens           |
| `/auth:logout`   | POST   | Invalidate current session's refresh token  |
| `/auth:refresh`  | POST   | Exchange refresh token for new tokens       |
| `/auth:me`       | GET    | Get current authenticated user info         |
| `/auth:me`       | POST   | Update current user's profile/password      |


See [Authentication API](./020-auth.md).

## Manage User (Admin Only)

| Endpoint        | Method | Description                              |
|-----------------|--------|------------------------------------------|
| `/users:list`     | GET    | List all users                           |
| `/users:get`      | GET    | Get specific user by ID                  |
| `/users:create`   | POST   | Create new user                          |
| `/users:update`   | POST   | Update user properties or admin actions  |
| `/users:destroy`  | POST   | Delete user account                      |

## Manage API Keys (Admin Only)

| Endpoint         | Method | Description                                 |
|------------------|--------|---------------------------------------------|
| `/apikeys:list`    | GET    | List all API keys                           |
| `/apikeys:get`     | GET    | Get specific API key                        |
| `/apikeys:create`  | POST   | Create new API key                          |
| `/apikeys:update`  | POST   | Update API key metadata or rotate key       |
| `/apikeys:destroy` | POST   | Delete API key                              |

## Manage Collections

These endpoints manage database tables (collections) and their schemas.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/collections:list` | GET | List all collections |
| `/collections:get` | GET | Get collection schema (requires `?name=...`) |
| `/collections:create` | POST | Create a new collection |
| `/collections:update` | POST | Update collection schema |
| `/collections:destroy` | POST | Delete a collection |

Update collection support following schema modification operations:

- `add_columns` - Add new columns
- `rename_columns` - Rename existing columns
- `modify_columns` - Change column types or attributes
- `remove_columns` - Remove existing columns

### Aggregation Operations

Moon provides dedicated aggregation endpoints that perform calculations directly on the server. This enables fast, efficient analytics—such as counting records, summing numeric fields, computing averages, and finding minimum or maximum values—without transferring unnecessary data.

**Aggregation Endpoints:**

- Count Records: `GET /{collection_name}:count`
- Sum Numeric Field: `GET /{collection_name}:sum` (requires `?field=...`)
- Average Numeric Field: `GET /{collection_name}:avg` (requires `?field=...`)
- Minimum Value: `GET /{collection_name}:min` (requires `?field=...`)
- Maximum Value: `GET /{collection_name}:max` (requires `?field=...`)

**Note:**

- Replace `{collection_name}` with your collection name.
- Aggregation can be combined with filters (e.g., `?quantity[gt]=10`) to perform calculations on specific subsets of data.
- Aggregation functions (`sum`, `avg`, `min`, `max`) are supported only on `integer` and `decimal` field types.
- Combine aggregation with query filters for calculations on specific subsets:
  - `/products:count?quantity[gt]=10`
  - `/products:sum?field=quantity&brand[eq]=Wow`
  - `/products:max?field=quantity`

Refer Detailed API [080-aggregation.md](./080-aggregation.md)

**Validation note:** Returns `400` (Standard Error Response) if `field` is missing or the named field is not a numeric type (`integer` or `decimal`) in the collection schema.

## Data Access

### Query Options

Query parameters for filtering, sorting, searching, field selection, and pagination when listing records. Using these options allows you to retrieve specific subsets of data based on your criteria.

| Query Options | Description |
|---------------|-------------|
| `?column[operator]=value` | Filter records by column values using comparison operators |
| `?sort={fields}` | Sort by one or more fields (prefix `-` for descending) |
| `?q={term}` | Full-text search across all text columns |
| `?fields={field1,field2}` | Select specific fields to return (id always included) |
| `?limit={number}` | Limit number of records returned (default: 15, max: 100) |
| `?after={cursor}` | Get records after the specified cursor |

## Security

## Standard Error Response

**Error Response:** Follow [090-error.md](./090-error.md) for any error handling

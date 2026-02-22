# Standard API Response Patterns

This document describes the standard response patterns, query options, and aggregation operations for the Moon API. All endpoints follow consistent conventions for success and error responses.

## Public Endpoints

Health and documentation endpoints are accessible without authentication. All other endpoints require authentication.

| Endpoint            | Method | Description                                 |
|---------------------|--------|---------------------------------------------|
| `/health`           | GET    | Health Endpoint (see [010-health.md](./010-health.md)) |
| `/doc/`             | GET    | API Documentation (HTML)                    |
| `/doc/llms.md`      | GET    | API Documentation (Markdown)                |
| `/doc/llms.txt`     | GET    | API Documentation (Plain Text, alias for `/doc/llms.md`) |
| `/doc/llms.json`    | GET    | API Documentation (JSON)                    |

See [Documentation Endpoints](./002-doc.md).


## Authentication

| Endpoint        | Method | Description                                |
| --------------- | ------ | ------------------------------------------ |
| `/auth:login`   | POST   | Authenticate user, receive tokens          |
| `/auth:logout`  | POST   | Invalidate current session's refresh token |
| `/auth:refresh` | POST   | Exchange refresh token for new tokens      |
| `/auth:me`      | GET    | Get current authenticated user info        |
| `/auth:me`      | POST   | Update current user's profile/password     |

See [Authentication API](./020-auth.md).

## Manage User (Admin Only)

| Endpoint         | Method | Description                             |
| ---------------- | ------ | --------------------------------------- |
| `/users:list`    | GET    | List all users                          |
| `/users:get`     | GET    | Get specific user by ID                 |
| `/users:create`  | POST   | Create new user                         |
| `/users:update`  | POST   | Update user properties or admin actions |
| `/users:destroy` | POST   | Delete user account                     |

See [Users API](./030-users.md).

## Manage API Keys (Admin Only)

| Endpoint           | Method | Description                           |
| ------------------ | ------ | ------------------------------------- |
| `/apikeys:list`    | GET    | List all API keys                     |
| `/apikeys:get`     | GET    | Get specific API key                  |
| `/apikeys:create`  | POST   | Create new API key                    |
| `/apikeys:update`  | POST   | Update API key metadata or rotate key |
| `/apikeys:destroy` | POST   | Delete API key                        |

See [APIKeys API](./040-apikeys.md).

## Manage Collections

These endpoints manage database tables (collections) and their schemas.

| Endpoint               | Method | Description                                  |
| ---------------------- | ------ | -------------------------------------------- |
| `/collections:list`    | GET    | List all collections                         |
| `/collections:get`     | GET    | Get collection schema (requires `?name=...`) |
| `/collections:create`  | POST   | Create a new collection                      |
| `/collections:update`  | POST   | Update collection schema                     |
| `/collections:destroy` | POST   | Delete a collection                          |

Update collection support following schema modification operations:

- `add_columns` - Add new columns
- `rename_columns` - Rename existing columns
- `modify_columns` - Change column types or attributes
- `remove_columns` - Remove existing columns

See [Collection Managment API](./050-collection.md).

## Data Access

These endpoints manage records within a specific collection. Replace `{collection_name}` with your collection name.

| Endpoint                     | Method | Description                              |
| ---------------------------- | ------ | ---------------------------------------- |
| `/{collection_name}:list`    | GET    | List all records                         |
| `/{collection_name}:schema`  | GET    | Get collection schema (read-only)        |
| `/{collection_name}:get`     | GET    | Get a single record (requires `?id=...`) |
| `/{collection_name}:create`  | POST   | Create a new record                      |
| `/{collection_name}:update`  | POST   | Update an existing record                |
| `/{collection_name}:destroy` | POST   | Delete a record                          |

For complete details on API request and response formats, supported endpoints, and data examples, see [060-data.md](./060-data.md). All error handling must follow [Standard Error Response](./090-error.md).

### Query Options

Query parameters for filtering, sorting, searching, field selection, and pagination when listing records. Using these options allows you to retrieve specific subsets of data based on your criteria.

| Query Options             | Description                                                |
| ------------------------- | ---------------------------------------------------------- |
| `?column[operator]=value` | Filter records by column values using comparison operators |
| `?sort={fields}`          | Sort by one or more fields (prefix `-` for descending)     |
| `?q={term}`               | Full-text search across all text columns                   |
| `?fields={field1,field2}` | Select specific fields to return (id always included)      |
| `?limit={number}`         | Limit number of records returned (default: 15, max: 100)   |
| `?after={cursor}`         | Get records after the specified cursor                     |


### Aggregation Operations

Moon provides dedicated aggregation endpoints that perform calculations directly on the server. This enables fast, efficient analytics—such as counting records, summing numeric fields, computing averages, and finding minimum or maximum values—without transferring unnecessary data.

Server-side aggregation endpoints for analytics. Replace `{collection_name}` with your collection name.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/{collection_name}:count` | GET | Count records |
| `/{collection_name}:sum` | GET | Sum numeric field (requires `?field=...`) |
| `/{collection_name}:avg` | GET | Average numeric field (requires `?field=...`) |
| `/{collection_name}:min` | GET | Minimum value (requires `?field=...`) |
| `/{collection_name}:max` | GET | Maximum value (requires `?field=...`) |

**Note:**

- Replace `{collection_name}` with your collection name.
- Aggregation can be combined with filters (e.g., `?quantity[gt]=10`) to perform calculations on specific subsets of data.
- Aggregation functions (`sum`, `avg`, `min`, `max`) are supported only on `integer` and `decimal` field types.
- Combine aggregation with query filters for calculations on specific subsets:
  - `/products:count?quantity[gt]=10`
  - `/products:sum?field=quantity&brand[eq]=Wow`
  - `/products:max?field=quantity`

Refer Detailed API [080-aggregation.md](./080-aggregation.md)

## Security

## Standard Error Response

**Error Response:** Follow [090-error.md](./090-error.md) for any error handling

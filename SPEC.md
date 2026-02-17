# Moon - Dynamic Headless Engine

This document outlines the architecture and design for a high-performance, API-first backend built in **Go**. The system allows for real-time, migration-less database management via REST APIs using a **custom-action pattern** and **in-memory schema caching**.

> **📖 API Reference**: For complete API documentation including all endpoints, request/response formats, query options, and error codes, see **[SPEC_API.md](SPEC_API.md)**.
>
> **🔐 Authentication**: For authentication flows, JWT tokens, API keys, roles, and permissions, see **[SPEC_AUTH.md](SPEC_AUTH.md)**.

## Table of Contents

1. [System Philosophy](#1-system-philosophy)
2. [Data Types](#data-types)
3. [Default Values](#default-values)
4. [Validation Constraints](#validation-constraints)
5. [API Standards](#api-standards)
6. [Configuration Architecture](#configuration-architecture)
7. [API Endpoint Specification](#2-api-endpoint-specification)
8. [Database Schema Design](#database-schema-design)
9. [Interface & Integration](#interface--integration)
10. [Authentication & Authorization](#authentication--authorization)
11. [Design for AI Maintainability](#4-design-for-ai-maintainability)
12. [Persistence Layer & Agnosticism](#5-persistence-layer--agnosticism)
13. [End-User Testing](#6-end-user-testing)

## 1. System Philosophy

- **Migration-Less Data Modeling:** Database tables and columns are created, modified, and deleted via API calls rather than manual migration files.
- **AIP-136 Custom Actions:** APIs use a colon separator (`:`) to distinguish between the resource and the action, providing a predictable and AI-friendly interface.
- **Zero-Latency Validation:** An **In-Memory Schema Registry** (using `sync.Map`) stores the current database structure, allowing the server to validate requests in nanoseconds before hitting the disk.
- **Resource Efficiency:** Targeted to run with a memory footprint under **50MB**, optimized for cloud-native and edge deployments.
- **Database Default:** SQLite is used as the default database if no other is specified. For most development and testing scenarios, you do not need to configure a database connection string unless you want to use Postgres or MySQL.

## Data Types

Moon supports a simplified, portable type system that maps consistently across all supported databases (SQLite, PostgreSQL, MySQL). This design prioritizes simplicity and predictability over fine-grained type control.

### Supported Data Types

| API Type   | Description                             | SQLite   | PostgreSQL   | MySQL        |
| ---------- | --------------------------------------- | -------- | ------------ | ------------ |
| `string`   | Text values of any length               | TEXT     | TEXT         | TEXT         |
| `integer`  | 64-bit integer values                   | INTEGER  | BIGINT       | BIGINT       |
| `decimal`  | Exact numeric values (e.g., price)      | NUMERIC  | NUMERIC(19,2)| DECIMAL(19,2)|
| `boolean`  | True/false values                       | INTEGER  | BOOLEAN      | BOOLEAN      |
| `datetime` | Date and time (RFC3339/ISO 8601 format) | TEXT     | TIMESTAMP    | TIMESTAMP    |
| `json`     | Arbitrary JSON objects or arrays        | TEXT     | JSON         | JSON         |

### Decimal Type

The `decimal` type provides **exact, deterministic numeric handling** for precision-critical values such as price, amount, weight, tax, and quantity. This addresses the inherent precision errors in floating-point arithmetic.

**API Representation:**
- Input and output are **strings** (e.g., `"199.99"`, `"-42.75"`, `"0.01"`)
- Preserves precision across serialization and deserialization
- Supports SQL aggregation functions (`SUM`, `AVG`, `MIN`, `MAX`)

**Validation:**
- Default scale: 2 decimal places
- Maximum scale: 10 decimal places
- No scientific notation allowed
- No locale-specific separators (e.g., no comma thousands separator)

**Valid formats:**
- `"10"`, `"10.50"`, `"1299.99"`, `"-42.75"`, `"0.01"`

**Invalid formats:**
- `"abc"` (non-numeric)
- `"1e10"` (scientific notation)
- `"10.999"` (exceeds default scale of 2)
- `"10."` (trailing decimal point)
- `".50"` (leading decimal point)

**Example usage:**
```json
{
  "name": "products",
  "columns": [
    {"name": "price", "type": "decimal", "nullable": false},
    {"name": "tax", "type": "decimal", "nullable": true}
  ]
}
```

### Design Rationale

- **No `float` type:** Floating-point numbers are discouraged due to precision issues. Use `integer` for whole numbers or `decimal` for exact precision values like currency and measurements.
- **No `text` vs `string` distinction:** All string data maps to `TEXT` for simplicity. There is no VARCHAR length limit enforcement at the database level.
- **JSON storage:** JSON data is stored as TEXT in SQLite and native JSON in PostgreSQL/MySQL.
- **Boolean storage:** SQLite uses INTEGER (0/1) for boolean values; PostgreSQL and MySQL use native BOOLEAN.
- **Decimal storage:** Uses native NUMERIC/DECIMAL types for exact arithmetic. API exposes values as strings to preserve precision in JSON serialization.

## Default Values

Moon handles default values strictly at the database column level during collection creation. Default values are NOT applied record-by-record during insert operations.

### Nullable Field Behavior

The `nullable` property controls API request validation and default value application:

**`nullable: false` (Required Fields):**
- Field **MUST** be present in every API request with a valid value
- Omitting the field or setting it to `null` results in a validation error
- **No automatic default values** are applied at the application level
- Non-nullable fields cannot have default values

**`nullable: true` (Optional Fields):**
- Field **MAY** be omitted from API requests
- When omitted, the database column default is used (automatically set by Moon backend during table creation)
- Can explicitly be set to `null` in requests (stored as NULL in database)

### Collection Creation Defaults

**Important:** Default values for columns are managed internally by the Moon backend and **cannot be set or modified via API requests** to `/collections:create` or `/collections:update`. Any request containing `default` fields will be rejected with a 400 Bad Request error.

When collections are created, Moon automatically applies type-based default values for nullable fields at the database level.

**Default Values by Type** (automatically applied during collection creation for nullable fields):

| Type | Default Value | Notes |
|------|--------------|-------|
| `string` | `""` (empty string) | Applied only if field is nullable |
| `integer` | `0` | Applied only if field is nullable |
| `decimal` | `"0.00"` | Applied only if field is nullable |
| `boolean` | `false` | Applied only if field is nullable |
| `datetime` | `null` | Applied for nullable fields |
| `json` | `"{}"` (empty object) | Applied only if field is nullable |

**Important:** Defaults are automatically set for nullable fields during collection creation by the Moon backend. Non-nullable fields have NO default and must always be provided in API requests.

### API Restrictions on Default Values

**Default values cannot be set or modified via API endpoints:**

- The `/collections:create` endpoint does NOT accept `default` fields in column definitions
- The `/collections:update` endpoint does NOT accept `default` fields in `add_columns` or `modify_columns` operations
- Any request containing these fields will be rejected with a 400 Bad Request error

**Example of rejected request:**

```json
// ❌ REJECTED - This request will fail with 400 Bad Request
POST /collections:create
{
  "name": "tasks",
  "columns": [
    {
      "name": "priority",
      "type": "integer",
      "nullable": true,
      "default": "5"  // ❌ Not allowed - will trigger validation error
    }
  ]
}
```

**Error Response:**

```json
{
  "code": 400,
  "error": "unknown field 'default' in column 0"
}
```

**Correct usage:**

```json
// ✓ ACCEPTED - Default values are applied automatically by the backend
POST /collections:create
{
  "name": "tasks",
  "columns": [
    {
      "name": "priority",
      "type": "integer",
      "nullable": true
      // Type default (0) will be applied automatically
    }
  ]
}
```

### Reading Default Values from Schema

While default values cannot be set via API, they **are visible in schema responses**. Use `/collections:get` to inspect the default values that the backend has configured for a collection:

```json
GET /collections:get?name=tasks

Response:
{
  "collection": {
    "name": "tasks",
    "columns": [
      {
        "name": "priority",
        "type": "integer",
        "nullable": true,
        "default": "0"  // ✓ Visible in schema responses
      }
    ]
  }
}
```

### nullable vs. default

- **`nullable` (API):** Controls whether a field must be provided in API requests
  - `nullable: false` → field **MUST** be provided in every API request (validation error if omitted)
  - `nullable: true` → field **MAY** be omitted from API requests (uses database default if omitted)

- **`default` (Schema/Database - Read-Only):** The database column default value for nullable fields
  - Automatically set by the backend for nullable fields
  - Enforced by the database layer (not application logic)
  - Only applies to nullable fields
  - **Cannot be set or modified via API** - managed internally by Moon
  - Visible in schema responses from `/collections:get`

**Behavior Matrix:**

| nullable | default | field omitted | field = null | Result |
|----------|---------------|---------------|--------------|---------|
| `false` | N/A* | ✗ | ✗ | **Validation error** - field is required |
| `true` | (type default) | ✓ | ✓ | Type default applied (if omitted), NULL (if explicit) |

*Note: Default values are automatically applied by the backend for nullable fields. Non-nullable fields must always be provided in API requests.

**Examples:**

```json
// Example 1: Required field (no default)
{
  "name": "title",
  "type": "string",
  "nullable": false
  // Must be provided in every API request - no default
}

// Example 2: Optional field with automatic type default
{
  "name": "description",
  "type": "string",
  "nullable": true
  // Type default ("") automatically applied by backend when omitted
}

// Example 3: Optional integer field with automatic type default
{
  "name": "priority",
  "type": "integer",
  "nullable": true
  // Type default (0) automatically applied by backend when omitted
}
```

## Validation Constraints

Moon enforces strict validation rules to ensure data integrity and prevent naming conflicts.

### Collection Name Constraints

| Constraint | Value | Notes |
|------------|-------|-------|
| Minimum length | 2 characters | Single-character names are not allowed |
| Maximum length | 63 characters | Matches PostgreSQL identifier limit |
| Pattern | `^[a-zA-Z][a-zA-Z0-9_]*$` | Must start with letter, alphanumeric + underscores |
| Case normalization | Lowercase | Names are automatically converted to lowercase |
| Reserved endpoints | `collections`, `auth`, `users`, `apikeys`, `doc`, `health` | Case-insensitive |
| System prefix | `moon_*`, `moon` | Reserved for internal system tables |
| SQL keywords | 100+ keywords | `select`, `insert`, `update`, `delete`, `table`, etc. |

### Column Name Constraints

| Constraint | Value | Notes |
|------------|-------|-------|
| Minimum length | 3 characters | Short names like `id`, `at` are not allowed |
| Maximum length | 63 characters | Matches PostgreSQL identifier limit |
| Pattern | `^[a-z][a-z0-9_]*$` | Lowercase only, must start with letter |
| Reserved names | `pkid`, `id` | System columns, automatically created |
| SQL keywords | 100+ keywords | Same list as collection names |

**Important:** Unlike collection names, column names are NOT auto-normalized to lowercase. Uppercase characters will be rejected with an error.

### System Limits

| Limit | Default | Configurable | Notes |
|-------|---------|--------------|-------|
| Max collections | 1000 | Yes (`limits.max_collections`) | Per server |
| Max columns | 100 | Yes (`limits.max_columns_per_collection`) | Per collection (includes system columns) |
| Max filters | 20 | Yes (`limits.max_filters_per_request`) | Per request |
| Max sort fields | 5 | Yes (`limits.max_sort_fields_per_request`) | Per request |

### Pagination Limits

| Limit | Default | Configurable | Notes |
|-------|---------|--------------|-------|
| Min page size | 1 | No | Hardcoded minimum |
| Default page size | 15 | Yes (`pagination.default_page_size`) | When no limit specified |
| Max page size | 200 | Yes (`pagination.max_page_size`) | Maximum allowed |

## API Standards

> **📖 Complete API Reference**: For detailed API documentation including all endpoints, request/response formats, query options, aggregation operations, and error codes, see **[SPEC_API.md](SPEC_API.md)**.
>
> This section covers API design principles and patterns. Implementation details are in SPEC_API.md.

Moon implements industry-standard API patterns for consistent client experience.

### HTTP Methods

**Moon only supports GET and POST HTTP methods.** All other HTTP methods (PUT, DELETE, PATCH, etc.) are not supported and will return a `405 Method Not Allowed` error.

- **GET** is used for read operations (list, get, count, aggregation queries)
- **POST** is used for write operations (create, update, destroy) and authentication operations
- **OPTIONS** is supported for CORS preflight requests only

This design choice:
- Simplifies routing and middleware logic
- Works universally with all HTTP clients and proxies
- Follows the AIP-136 custom actions pattern where the action is in the URL (`:create`, `:update`, `:destroy`)
- Ensures compatibility with restrictive network environments that may filter uncommon HTTP methods

### Rate Limiting Headers

When rate limiting is enabled, all responses include rate limit headers:

| Header | Description | Example |
|--------|-------------|---------|
| `X-RateLimit-Limit` | Maximum requests per window | `100` |
| `X-RateLimit-Remaining` | Remaining requests in window | `87` |
| `X-RateLimit-Reset` | Unix timestamp when window resets | `1704067200` |
| `Retry-After` | Seconds until retry (429 responses only) | `60` |

### Error Response Format

All error responses follow a consistent JSON structure.

> **See SPEC_API.md** for complete error response documentation and examples across all endpoint types.

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "human-readable error message"
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Input validation failed |
| `INVALID_JSON` | 400 | Malformed JSON |
| `INVALID_ULID` | 400 | Invalid ULID format |
| `INVALID_PARAMETER` | 400 | Invalid or unsupported query or body parameter |
| `PAGE_SIZE_EXCEEDED` | 400 | Page size exceeds maximum |
| `COLLECTION_NOT_FOUND` | 404 | Collection does not exist |
| `DUPLICATE_RECORD` | 404 | Resource with unique field already exists |
| `RECORD_NOT_FOUND` | 404 | Record not found |
| `DUPLICATE_COLLECTION` | 409 | Collection name already exists |
| `MAX_COLLECTIONS_REACHED` | 409 | Maximum collections limit reached |
| `MAX_COLUMNS_REACHED` | 409 | Maximum columns limit reached |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

### CORS Support

Cross-Origin Resource Sharing (CORS) can be enabled via configuration:

```yaml
cors:
  enabled: true
  allowed_origins:
    - "https://app.example.com"
  allowed_methods:
    - GET
    - POST
    - OPTIONS
  allowed_headers:
    - Content-Type
    - Authorization
  allow_credentials: true
  max_age: 3600
  
  # Endpoint-specific CORS registration
  endpoints:
    - path: "/health"
      pattern_type: "exact"
      allowed_origins: ["*"]
      allowed_methods: ["GET", "OPTIONS"]
      allowed_headers: ["Content-Type"]
      allow_credentials: false
      bypass_auth: true
    
    - path: "/doc/"
      pattern_type: "prefix"
      allowed_origins: ["*"]
      allowed_methods: ["GET", "OPTIONS"]
      allowed_headers: ["Content-Type"]
      allow_credentials: false
      bypass_auth: true
```

**Note:** Only GET, POST, and OPTIONS methods are supported by the Moon server. Including other methods (PUT, DELETE, PATCH) in the `allowed_methods` configuration will not enable them on the server.

**CORS Endpoint Registration:**

Moon supports dynamic CORS endpoint registration with pattern matching:

- **Pattern Types:**
  - `exact`: Matches exact path only (e.g., `/health` matches `/health` but not `/health/status`)
  - `prefix`: Matches path prefix (e.g., `/doc/` matches `/doc/api`, `/doc/llms.md`, `/doc/llms.txt`, `/doc/llms.json`. Note: `/doc/` does not match `/doc` without trailing slash)
  - `suffix`: Matches path suffix (e.g., `*.json` matches `/data/users.json`)
  - `contains`: Matches if path contains substring (e.g., `/public/` matches any path with `/public/`)

- **Priority:** When multiple patterns match, the most specific match is used:
  1. Exact matches (highest priority)
  2. Longest prefix matches
  3. Longest suffix matches
  4. Longest contains matches
  5. Global CORS configuration (fallback)

- **Authentication Bypass:** Set `bypass_auth: true` to skip authentication for public endpoints (health, docs, status).

- **Default Endpoints:** If `cors.endpoints` is not specified, these defaults are applied:
  - `/health` (exact, `*`, no auth)
  - `/doc/` (prefix, `*`, no auth - matches all paths starting with `/doc/` including `/doc/`, `/doc/llms.md`, `/doc/llms.txt`, and `/doc/llms.json`)

CORS headers exposed to browsers:
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`
- `X-Request-ID`

### Sensitive Data Redaction

Moon automatically redacts sensitive fields in logs to prevent credential leakage:

**Default Sensitive Fields:**
- `password`, `token`, `secret`, `api_key`, `apikey`
- `authorization`, `jwt`, `refresh_token`, `access_token`
- `client_secret`, `private_key`, `credential`, `auth`

**Configuration:**
```yaml
logging:
  redact_sensitive: true  # Default: true
  additional_sensitive_fields:
    - "ssn"
    - "credit_card"
```

## Configuration Architecture

The system uses YAML-only configuration with centralized defaults:

- **YAML Configuration Only:** All configuration is stored in YAML format at `/etc/moon.conf` (default) or custom path via `--config` flag
- **No Environment Variables:** Configuration values must be set in the YAML file - no environment variable overrides
- **Centralized Defaults:** All default values are defined in the `config.Defaults` struct to eliminate hardcoded literals
- **Immutable State:** On startup, the configuration is parsed into a global, read-only `AppConfig` struct to prevent accidental runtime mutations and ensure thread safety

> **📖 Configuration Reference**: See **[moon.conf](moon.conf)** in the project root for the complete, self-documented configuration file with all available options, defaults, and inline documentation.

### Configuration Principles

- **Single Source:** `moon.conf` is the only configuration file - fully documented with inline comments
- **Secure Defaults:** Most options have sensible defaults; only `jwt.secret` and database settings require customization
- **Location:** Default `/etc/moon.conf` or custom path via `--config` flag
- **Format:** YAML with inline documentation for all options

### Quick Start

1. Copy `moon.conf` to `/etc/moon.conf`
2. Set `jwt.secret` to a secure random value (use `openssl rand -base64 32`)
3. Configure database connection (SQLite is default)
4. Start Moon: `moon --config /etc/moon.conf`

### Recovery and Consistency Checking

Moon includes robust consistency checking and recovery logic that ensures the in-memory schema registry remains synchronized with the physical database tables across restarts and failures.

**On Startup:**

- Moon performs an automatic consistency check comparing the registry with physical database tables
- If inconsistencies are detected, they are logged with detailed information
- With `auto_repair: true` (default), Moon automatically repairs inconsistencies:
  - **Orphaned registry entries** (registered but table doesn't exist): Removed from registry
  - **Orphaned tables** (table exists but not registered):
    - If `drop_orphans: false` (default): Table schema is inferred and registered
    - If `drop_orphans: true`: Table is dropped from database

**Consistency Check:**

- Runs within the configured timeout (default 5 seconds)
- Non-blocking with configurable timeout to prevent indefinite startup delays
- Results are logged and displayed during startup
- Startup fails if critical issues cannot be repaired

**Health Endpoint:**

- The `/health` endpoint provides health check information for liveness and readiness checks
- Returns a JSON response with four fields:
  - `status`: Service health status (`ok` or `down`)
  - `database`: Database connectivity status (`ok` or `error`)
  - `version`: Service version string (e.g., `1.0.0`)
  - `timestamp`: ISO 8601 timestamp of the health check
- Always returns HTTP 200, even when the service is down
- Clients must check the `status` field to determine service health
- Should not expose internal details like database type, collection count, or consistency status

**Example health response:**

```json
{
  "status": "ok",
  "database": "ok",
  "moon": "1.0.0",
  "timestamp": "2026-02-03T13:58:53Z"
}
```


### Running Modes

#### Preflight Checks

Before the server starts, Moon performs filesystem preflight checks:

- Ensures the logging directory exists (and creates it if missing)
- For SQLite, ensures the database parent directory exists (and creates it if missing)
- In daemon mode, truncates the log file to start fresh

#### Console Mode (Default)

```bash
moon 
```

- Runs in foreground attached to terminal
- The `--config /{path}/moon.conf` flag is optional. If not provided, Moon will attempt to load the configuration from the default location `/etc/moon.conf`.
- Logs output to both stdout/stderr AND log file (`/var/log/moon/main.log` or path specified in config)
- Stdout logs use console format (colorized, human-readable)
- File logs use simple format (`[LEVEL](TIMESTAMP): {MESSAGE}`)
- Process terminates when terminal closes or Ctrl+C is pressed

#### Daemon Mode

```bash
moon --daemon
# or shorthand
moon -d
```

- Runs as background daemon process
- Logs written only to `/var/log/moon/main.log` (or path specified in config)
- PID file written to `/var/run/moon.pid`
- Process continues after terminal closes
- Supports graceful shutdown via SIGTERM/SIGINT

## 2. API Endpoint Specification

The system uses a strict pattern to ensure that AI agents and developers can interact with any collection without new code deployment.

- **RESTful API:** A standardized API following strict predictable patterns, making it easy for AI to generate documentation.
- **Configurable Prefix:** All API endpoints are mounted under a configurable URL prefix (default: empty string).
  - Default (no prefix): `/health`, `/collections:list`, `/{collection}:list`
  - With custom prefix: `/{prefix}/health`, `/{prefix}/collections:list`, `/{prefix}/{collection}:list`
  - Example With `/api/v1` prefix: `/api/v1/health`, `/api/v1/collections:list`, `/api/v1/{collection}:list`

### A. Schema Management (`/collections`)

These endpoints manage the database tables and metadata.

> **See SPEC_API.md** for complete endpoint documentation including request/response formats, parameters, and examples.

**Note:** All endpoints below are shown without a prefix. If a prefix is configured (e.g., `/api/v1`), prepend it to all paths.

| Endpoint                    | Method | Purpose                                                |
| --------------------------- | ------ | ------------------------------------------------------ |
| `GET /collections:list`     | `GET`  | List all managed collections from the cache.           |
| `GET /collections:get`      | `GET`  | Retrieve the schema (fields/types) for one collection. |
| `POST /collections:create`  | `POST` | Create a new table in the database.                    |
| `POST /collections:update`  | `POST` | Modify table columns (add/remove/rename).              |
| `POST /collections:destroy` | `POST` | Drop the table and purge it from the cache.            |

#### Collections List Response Format

The `GET /collections:list` endpoint returns detailed information about each collection, including record counts.

> **📖 Complete Format**: See [SPEC_API.md § Collections Endpoints](SPEC_API.md) for full response schema and examples.

**Response includes:**
- `collections` (array): Collection objects with `name` and `records` count
- `count` (integer): Total number of collections
- `records` value: 0 for empty collections, -1 if count cannot be retrieved (database error)

### B. Data Access (`/{collectionName}`)

These endpoints manage the records within a specific collection.

> **See SPEC_API.md** for complete endpoint documentation including request/response formats, query options, pagination, filtering, and aggregation operations.

**Note:** All endpoints below are shown without a prefix. If a prefix is configured, prepend it to all paths.

| Endpoint               | Method | Purpose                                            |
| ---------------------- | ------ | -------------------------------------------------- |
| `GET /{name}:list`     | `GET`  | Fetch all records from the specified table.        |
| `GET /{name}:get`      | `GET`  | Fetch a single record by its unique ID.            |
| `GET /{name}:schema`   | `GET`  | Retrieve the schema for a specific collection.     |
| `POST /{name}:create`  | `POST` | Insert a new record (validated against the cache). |
| `POST /{name}:update`  | `POST` | Update an existing record.                         |
| `POST /{name}:destroy` | `POST` | Delete a record from the table.                    |

#### Batch Operations

The `:create`, `:update`, and `:destroy` endpoints support both **single-object** and **batch** modes.

> **📖 Complete Documentation**: See [SPEC_API.md § Batch Operations](SPEC_API.md) for detailed request/response formats, atomic mode, error handling, examples, and configuration.

**Key Features:**
- **Automatic Detection:** Array = batch mode, object = single mode
- **Best-Effort (Default):** Process each record independently, return HTTP 207 with per-record results
- **Atomic Mode:** `?atomic=true` - all succeed or all fail (transaction)
- **Size Limits:** Default 50 records, 2MB payload (configurable)

**Configuration:**
```yaml
batch:
  max_size: 50               # Maximum records per batch
  max_payload_bytes: 2097152 # Maximum payload (2MB)
```

#### Identifiers

- Records use a ULID as the external identifier.
- The database stores a `pkid` column (auto-increment integer, internal use only) and an `id` column (ULID string).
- API responses expose the `id` column directly (which contains the ULID value).
- The internal `pkid` column is never exposed via the API.
- System columns (`pkid`, `id`) are automatically created and protected from modification, deletion, or renaming.

#### Query Parameters

The list endpoint supports powerful query parameters for filtering, sorting, searching, and field selection.

> **📖 Complete Documentation**: See [SPEC_API.md § Query Options](SPEC_API.md) for all operators, pagination, field selection, full-text search, and examples.

**Filtering:** `?column[operator]=value` (operators: `eq`, `ne`, `gt`, `lt`, `gte`, `lte`, `like`, `contains`, `in`, `null`)  
**Sorting:** `?sort=field` or `?sort=-field` (descending), multiple: `?sort=-created_at,name`  
**Search:** `?q=searchterm` (searches all text columns)  
**Fields:** `?fields=field1,field2` (select specific fields)  
**Pagination:** `?after=<ulid>&limit=N` (cursor-based)

**Example:** `GET /products:list?q=laptop&price[gt]=500&sort=-price&fields=name,price&limit=10`

#### Schema Retrieval

Retrieve collection schema using `GET /{collection}:schema`.

> **📖 Complete Documentation**: See [SPEC_API.md § Schema Endpoint](SPEC_API.md) for full response format and field properties.

**Response includes:** Collection name, field definitions (name, type, nullable, readonly), and total record count.

**Authentication:** Required (Bearer token or API key)

**Error Responses:**
- `401 Unauthorized`: Missing or invalid authentication
- `404 Not Found`: Collection does not exist
- `500 Internal Server Error`: Unexpected errors

**Example:**
```bash
curl -H "Authorization: Bearer $ACCESS_TOKEN" https://api.example.com/products:schema
```

### C. Aggregation Operations (`/{collectionName}`)

Server-side aggregation for analytics without fetching full datasets.

> **📖 Complete Documentation**: See [SPEC_API.md § Aggregation Operations](SPEC_API.md) for all operations, filtering, response formats, and examples.

| Endpoint                    | Purpose                                |
| --------------------------- | -------------------------------------- |
| `GET /{name}:count`         | Count records                          |
| `GET /{name}:sum?field=...` | Sum numeric field values               |
| `GET /{name}:avg?field=...` | Calculate average of numeric field     |
| `GET /{name}:min?field=...` | Find minimum value of numeric field    |
| `GET /{name}:max?field=...` | Find maximum value of numeric field    |

**Key Features:**
- Supports filtering (same syntax as `:list`)
- Works with `integer` and `decimal` types
- Returns `{"value": <number>}` format
- Database-level processing for performance

### D. Documentation Endpoints

Moon provides auto-generated documentation endpoints.

> **📖 Complete Documentation**: See [SPEC_API.md § Documentation Endpoints](SPEC_API.md) for all formats and caching details.

| Endpoint                     | Purpose                                    |
| ---------------------------- | ------------------------------------------ |
| `GET /doc/`                  | HTML documentation                         |
| `GET /doc/llms.md`           | Markdown for AI agents                     |
| `GET /doc/llms.txt`          | Text format (alias)                        |
| `GET /doc/llms.json`         | JSON schema for machine consumption        |
| `POST /doc:refresh`          | Clear cache and regenerate                 |

**Features:** Automatic collection discovery, curl examples, query documentation, caching with ETag/Last-Modified headers.

### E. Collection Column Operations

The `POST /collections:update` endpoint manages column lifecycle (add, remove, rename, modify).

> **📖 Complete Documentation**: See [SPEC_API.md § Collection Update Endpoint](SPEC_API.md) for request formats, operation order, and examples.

**Key Rules:**
- System columns (`pkid`, `id`) are protected
- Operations execute in order: rename → modify → add → remove
- All operations can be combined in a single request
- Registry atomically updated after successful DDL
- Validation before execution; rollback on failure

## Database Schema Design

### In-Memory Schema Registry

The server maintains a **sync.Map** cache of collection schemas for zero-latency validation.

**Data Flow:**
1. **Ingress:** Router parses `/:name:action`
2. **Validation:** Check in-memory registry; validate JSON against cached schema
3. **SQL Generation:** Build parameterized SQL statement
4. **Persistence:** Execute against database (PostgreSQL/MySQL/SQLite)
5. **Reactive Cache:** Refresh registry entry on schema changes

**Schema Cache Structure:**
- Collection name → field definitions (name, type, nullable, unique)
- System columns (`pkid`, `id`) automatically included
- Cache refreshed on collection create/update/destroy

### Identifiers

- **External ID:** ULID (exposed via API as `id` field)
- **Internal ID:** Auto-increment integer (`pkid`, never exposed)
- System columns automatically created and protected

## Interface & Integration

**Documentation:** Auto-generated docs via `/doc/` (HTML), `/doc/llms.md` (Markdown), `/doc/llms.json` (JSON)  
**Security Middleware:** JWT and API Key validation with role-based authorization  
**Advanced Controls:** Per-endpoint permissions, protected/unprotected path lists

> **📖 Details**: See [SPEC_AUTH.md](SPEC_AUTH.md) for authentication flows and [SPEC_API.md](SPEC_API.md) for all endpoints.

## Authentication & Authorization

> **📖 Complete Specification**: See [SPEC_AUTH.md](SPEC_AUTH.md) for detailed authentication flows, JWT tokens, API keys, roles, permissions, and security configuration.

**Authentication Methods:**

| Method | Header Format | Use Case | Rate Limit |
|--------|--------------|----------|------------|
| JWT | `Authorization: Bearer <token>` | Interactive users | 100 req/min |
| API Key | `Authorization: Bearer moon_live_*` | Machine-to-machine | 1000 req/min |

**Roles:** `admin` (full access), `user` (read + optional write), `readonly` (read-only)  
**Permissions:** Role-based with `can_write` flag for user role  
**Rate Limits:** Per-user/key with headers (`X-RateLimit-*`)

## 4. Design for AI Maintainability

- **Predictable Interface:** By standardizing the `:action` suffix, AI agents can guess the correct endpoint for any new collection with 100% accuracy.
- **Statically Typed Logic:** Although data is dynamic (`map[string]any`), the internal validation logic is strictly typed, preventing AI-generated bugs from corrupting the database.
- **Test-Driven Development:** Every module and feature is covered by automated tests (unit, integration, and API tests). Integration tests mock the database to ensure safe refactoring (e.g., of the SQL builder) and to guarantee the API contract is never broken. Tests are the foundation for all new code and refactoring. The project aims for maximum possible test coverage to ensure reliability and maintainability.
- **Strict Conventions:** By adhering to standard Go patterns, the codebase remains "recognizably structured." Both AI agents and human developers can navigate the project with 99% accuracy because files are exactly where they are expected to be.

---

## 5. Persistence Layer & Agnosticism

- **Dialect-Agnostic:** The server uses a driver-based approach. The user provides a connection string, and Moon-Go detects if it needs to use `Postgres`, `MySQL`, or `SQLite` syntax.
- **Database Type Fixed at Startup:** The database type is selected at server startup and cannot be changed at runtime.
- **Single-Tenant Focus:** Optimized as a high-speed core for a single application, ensuring maximum simplicity and maintainability.

## 6. End-User Testing

**Recommended:** Use curl for endpoint testing with JWT/API key authentication  
**Payloads:** JSON via `-d` flag with `Content-Type: application/json`  
**Prefix Support:** All endpoints respect configured URL prefix (default: none)

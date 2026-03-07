## Overview

- Moon supports three database backends: SQLite (default), PostgreSQL, and MySQL. All external API behavior must remain consistent regardless of which backend is selected.
- This PRD defines the persistence adapter interface, the backend implementations, query execution, timeout enforcement, slow-query logging, and the field-type mapping between Moon's API types and backend column types.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Adapter Interface

- A single internal interface must define all database operations used by Moon's services. Upper-layer code must depend only on this interface, never on backend-specific types.
- The interface must support at minimum:
  - executing raw DDL statements (for schema management)
  - querying rows with filter, sort, pagination, and field projection
  - inserting one or more rows
  - updating one or more rows by id
  - deleting one or more rows by id
  - introspecting physical table names in the connected database or schema
  - introspecting physical column definitions for a given table
  - pinging the database to verify connectivity
  - closing the connection cleanly

### Backend Selection

- The backend is selected from `database.connection` in the configuration.
- Valid values are `sqlite`, `postgres`, and `mysql` (defined as constants in `Config.go`).
- Any other value must fail startup.

### SQLite Backend

- Uses the file path from `database.database` (default: `/opt/moon/sqlite.db`).
- Must use WAL mode for concurrent read access.
- Column type mapping:

| Moon type | SQLite column type |
|-----------|--------------------|
| `id` | `TEXT` |
| `string` | `TEXT` |
| `integer` | `INTEGER` |
| `decimal` | `NUMERIC` |
| `boolean` | `INTEGER` (0/1) |
| `datetime` | `TEXT` |
| `json` | `TEXT` |

- Schema introspection must use `PRAGMA table_list` and `PRAGMA table_info({table})`.

### PostgreSQL Backend

- Requires `database.user`, `database.password`, `database.host`, and `database.database`.
- Column type mapping:

| Moon type | PostgreSQL column type |
|-----------|----------------------|
| `id` | `TEXT` |
| `string` | `TEXT` |
| `integer` | `BIGINT` |
| `decimal` | `NUMERIC(19,2)` |
| `boolean` | `BOOLEAN` |
| `datetime` | `TIMESTAMP` |
| `json` | `JSON` |

- Schema introspection must use `information_schema.tables` and `information_schema.columns`.

### MySQL Backend

- Requires `database.user`, `database.password`, `database.host`, and `database.database`.
- Column type mapping:

| Moon type | MySQL column type |
|-----------|-----------------|
| `id` | `TEXT` |
| `string` | `TEXT` |
| `integer` | `BIGINT` |
| `decimal` | `DECIMAL(19,2)` |
| `boolean` | `BOOLEAN` |
| `datetime` | `TIMESTAMP` |
| `json` | `JSON` |

- Schema introspection must use `SHOW TABLES` and `information_schema.columns`.

### External Type Invariants

Regardless of backend, these invariants must hold at the API layer:

| Moon type | API representation | Invariant |
|-----------|--------------------|-----------|
| `string` | JSON string | returned as-is |
| `integer` | JSON number | returned as integer number |
| `decimal` | JSON string | returned as string without scientific notation; max 10 fractional digits |
| `boolean` | JSON boolean | returned as `true` or `false` |
| `datetime` | JSON string | returned as RFC3339 |
| `json` | JSON object or array | returned as parsed JSON |

### Query Execution

- All queries must be parameterized; raw string interpolation into SQL is forbidden.
- The adapter must apply `database.query_timeout` as a `context.Context` deadline on every query.
- The adapter must measure query duration and emit a slow-query log line when duration exceeds `database.slow_query_threshold` milliseconds.
- Slow-query log lines must include: table name, operation type, duration in milliseconds.

### Connectivity Verification

- During startup the adapter must verify connectivity by executing a lightweight ping or equivalent health check.
- If the connection cannot be established, startup must fail.

### Error Propagation

- Backend-specific error types must be wrapped and translated into adapter-level errors before being returned to upper layers.
- SQL error details must not leak backend-specific syntax or internal query text into API error responses.

### No Direct SQL in Upper Layers

- Auth service, collection service, and resource service must not contain SQL strings.
- All SQL must reside in the adapter layer.

## Acceptance

- Configuring `database.connection = sqlite` starts Moon using the SQLite file at the configured path.
- Configuring `database.connection = postgres` with valid credentials connects to PostgreSQL.
- Configuring `database.connection = mysql` with valid credentials connects to MySQL.
- Configuring an unknown `database.connection` value causes startup failure.
- Querying a `decimal` field always returns a string in the API response with no scientific notation.
- Querying a `boolean` field always returns `true` or `false` in the API response (never `0` or `1`).
- Querying a `datetime` field always returns an RFC3339 string.
- A query that exceeds `database.query_timeout` seconds returns a timeout error; the API returns an appropriate error response.
- A query whose duration exceeds `database.slow_query_threshold` milliseconds produces a slow-query log line containing the table and duration.
- If the database cannot be reached during startup, the service exits with a non-zero code and a descriptive error.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

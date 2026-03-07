# Moon Software Design Specification

## 1. Document Authority and Scope

Moon is a single-process, API-first backend written in Go. This document is the authoritative specification for Moon's software design, including architecture, runtime behavior, data and schema management, configuration, validation, security, audit logging, and operational expectations.

This document is implementation-neutral. It defines required behavior, invariants, and boundaries. It does not prescribe package names, directory layout, or internal helper structure.

| Document      | Authority                                                                                                                                             |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `SPEC.md`     | Architecture, runtime behavior, database and schema design, configuration, security requirements, validation rules, and operational expectations.     |
| `SPEC_API.md` | Endpoint paths, request and response shapes, query parameter encodings, and HTTP status code contract.                                                |
| `SPEC/*.md`   | Supporting examples and endpoint notes. These files are subordinate to `SPEC.md` and `SPEC_API.md` and must not introduce contradictory design rules. |

Conflict resolution rules:

- If documents disagree about architecture, runtime behavior, configuration, validation, security, or operations, `SPEC.md` controls.
- If documents disagree about endpoint paths, request or response bodies, query parameter formats, or HTTP status codes, `SPEC_API.md` controls.
- Supporting files under `SPEC/` must be updated to match these two source-of-truth documents.

## 2. Product Model and Terminology

Moon manages data through runtime-defined collections rather than migration files.

| Moon term  | Database term |
| ---------- | ------------- |
| Collection | Table         |
| Field      | Column        |
| Record     | Row           |

Moon has two collection classes:

- system collections: `users`, `apikeys`
- dynamic collections: every other collection created through the collection APIs

System collections and dynamic collections are both accessible through the canonical data endpoints defined in `SPEC_API.md`. System collections have additional protection and lifecycle rules defined by this document.

## 3. Design Goals

Moon must prioritize:

- deterministic behavior over implicit behavior
- small, explicit internal boundaries
- portability across supported database backends
- runtime schema management without migration files
- predictable authentication and authorization behavior
- simple, testable rules that humans and AI agents can implement consistently

## 4. Non-Goals and Explicit Exclusions

Moon does not provide any of the following features:

- SQL joins
- database triggers or hooks
- foreign keys
- migration files or schema versioning
- built-in backup or restore
- background job processing or scheduled task runners
- realtime or WebSocket transport
- built-in file or binary storage
- fine-grained ACL beyond the authorization model defined here
- built-in admin UI or dashboard
- built-in encryption at rest
- API versioning

These exclusions are deliberate. Implementations must not introduce them implicitly or depend on them for correctness.

## 5. Global Invariants

The following rules are mandatory across the entire system:

| Area                      | Requirement                                                                                                                                                                       |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HTTP methods              | Only `GET`, `POST`, and `OPTIONS` are supported. All other methods must return `405 Method Not Allowed`.                                                                          |
| Public routes             | Only `/` and `/health` are public. All other routes require authentication. If `server.prefix` is set, these routes are prefixed like every other route.                          |
| Endpoint style            | Endpoints must follow the AIP-136 custom action pattern and use `:` to separate the resource from the action.                                                                     |
| Error body                | All error responses must use `{ "message": "..." }` only.                                                                                                                         |
| Identifiers               | Records, users, and API keys use server-generated ULID `id` values. Collections use `name`.                                                                                       |
| Schema authority          | The in-memory schema registry is the runtime source of truth for schema validation and request planning.                                                                          |
| Schema changes            | Schema changes must occur through the API. Migration files and out-of-band schema changes are not part of the design.                                                             |
| System collections        | `users` and `apikeys` must always exist and must not be created, renamed, modified, or destroyed through collection schema mutation APIs.                                         |
| Internal system tables    | Internal system tables must use the reserved prefix `moon_`, must never be returned by collection or data APIs, and must not be user-mutable.                                     |
| Database support          | SQLite is the default backend. PostgreSQL and MySQL are optional backends and must preserve the same external behavior within the limits of this specification.                   |
| Canonical resource routes | Canonical resource routes are `/data/{resource}:query`, `/data/{resource}:mutate`, and `/data/{resource}:schema`. Additional alias routes are not required by this specification. |

## 6. Architecture Overview

Moon is a single HTTP service with clearly separated runtime layers.

```text
Client
  |
  v
HTTP Transport
  |
  +--> Route and prefix resolution
  +--> CORS
  +--> Audit logging context
  +--> Authentication
  +--> Rate limiting
  +--> Authorization
  |
  v
Application Services
  |
  +--> Auth service
  +--> Collection service
  +--> Resource service
  |
  v
Schema Registry
  |
  v
Persistence Adapter
  |
  v
SQLite | PostgreSQL | MySQL
```

### 6.1 Layer Responsibilities

| Layer               | Responsibilities                                                                                         | Must Not Do                                             |
| ------------------- | -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| HTTP transport      | route requests, decode input, apply middleware, shape responses                                          | contain domain rules beyond request-level validation    |
| Authentication      | identify the caller from a bearer credential, distinguish JWT from API key, reject malformed credentials | grant resource permissions by itself                    |
| Authorization       | enforce public-route rules, role checks, and write capability                                            | bypass schema validation or resource rules              |
| Auth service        | login, refresh, logout, current-user behavior, API key lifecycle rules                                   | own HTTP parsing or response serialization              |
| Collection service  | schema validation, collection mutation, system collection protections                                    | execute persistence operations without prior validation |
| Resource service    | record query and record mutation rules                                                                   | bypass the schema registry                              |
| Schema registry     | store normalized schema metadata, support runtime lookup, refresh atomically after schema mutation       | serve stale schema after a successful schema change     |
| Persistence adapter | execute validated plans against the selected database backend                                            | leak backend-specific behavior into upper layers        |

### 6.2 Middleware Order

Requests must pass through middleware in this order:

1. route and prefix resolution
2. CORS handling
3. audit logging context creation
4. authentication for protected routes
5. rate limiting
6. authorization
7. handler and service execution
8. response shaping

Rationale:

- CORS must run early so browser preflight behavior is deterministic.
- Audit context must exist before authentication so rejected requests are still traceable.
- Authorization must occur before handlers perform domain work.
- Response shaping must be centralized so all errors and success envelopes remain consistent.

## 7. Runtime Lifecycle

### 7.1 Startup Sequence

The service must complete the following sequence before it accepts traffic:

1. load built-in configuration defaults from `Config.go`
2. resolve the configuration file path from `-c <path>` or the default `/etc/moon.conf`
3. load and apply configuration-file overrides from the resolved path
4. validate the resulting configuration
5. initialize logging to both the console and the configured log file
6. initialize the selected database adapter and verify connectivity
7. ensure required API-visible system collections and `moon_auth_refresh_tokens` exist
8. inspect the physical database schema and build the in-memory schema registry
9. start the HTTP server

Startup must fail if the configuration file cannot be read, if any required configuration is missing or invalid, if the configured log file cannot be opened, if the selected backend cannot be reached, or if the required system state cannot be reconciled safely.

### 7.2 Request Lifecycles

Read request flow:

```text
request
  -> middleware
  -> authentication and authorization
  -> schema lookup
  -> query validation
  -> adapter execution
  -> response shaping
```

Record mutation flow:

```text
request
  -> middleware
  -> authentication and authorization
  -> schema lookup
  -> payload validation
  -> adapter execution
  -> response shaping
```

Schema mutation flow:

```text
request
  -> middleware
  -> authentication and authorization
  -> schema mutation validation
  -> adapter schema change
  -> atomic schema registry refresh
  -> response shaping
```

### 7.3 State Refresh and Cleanup

The schema registry must remain on the previous committed state unless a schema mutation completes successfully.

Expired rows in `moon_auth_refresh_tokens` and other implementation-private authentication state may be cleaned up during startup or normal request handling, but correctness must not depend on background schedulers or out-of-band workers.

## 8. Configuration Model

### 8.1 Service Command and Configuration Sources

Moon standardizes the following service command parameter:

| Parameter    | Default          | Requirement                    |
| ------------ | ---------------- | ------------------------------ |
| `-c <path>`  | `/etc/moon.conf` | configuration file path to use |

Configuration sources and precedence:

1. built-in defaults defined as named constants in `Config.go`
2. overrides from the YAML configuration file resolved from `-c <path>` or `/etc/moon.conf`

Configuration-source rules:

- `Config.go` must be the single source for configuration keys, built-in defaults, and path constants used by the service.
- Magic values for configuration keys, default paths, limits, or timeouts must not be duplicated outside `Config.go`.
- If `-c` is omitted, the service must attempt to load `/etc/moon.conf`.
- If `-c` is provided, the service must load the specified file instead of the default path.
- No environment-variable override mechanism or additional runtime flag override mechanism is part of this specification unless it is added here in a future revision.

### 8.2 Configuration Validation and Failure Behavior

Configuration validation rules are mandatory:

- unknown keys must fail startup
- the resolved configuration file must exist and be readable
- malformed values must fail startup
- conditional keys must be present when the selected backend requires them
- insecure placeholder secrets must be rejected in production
- optional keys must use documented defaults when omitted
- backend-specific keys that do not apply to the selected backend must not alter behavior
- if any bootstrap admin field is provided, all bootstrap admin fields must be provided
- `server.logpath` must resolve to a writable file location

### 8.3 Configuration Reference

| Key                             | Required                                        | Default                                                 | Requirements                                                  |
| ------------------------------- | ----------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------------- |
| `server.host`                   | no                                              | `0.0.0.0`                                               | listen address                                                |
| `server.port`                   | no                                              | `6006`                                                  | integer in the valid TCP port range                           |
| `server.prefix`                 | no                                              | `""`                                                    | empty or a single leading-slash path prefix                   |
| `server.logpath`                | no                                              | `/var/log/moon.log`                                     | writable file path used in addition to console logging        |
| `database.connection`           | no                                              | `sqlite`                                                | `sqlite`, `postgres`, or `mysql`                              |
| `database.database`             | no for `sqlite`, yes for `postgres` and `mysql` | `/opt/moon/sqlite.db` when `database.connection=sqlite` | SQLite file path or database name                             |
| `database.user`                 | conditional                                     | none                                                    | required for backends that require a username                 |
| `database.password`             | conditional                                     | none                                                    | required for backends that require a password                 |
| `database.host`                 | conditional                                     | none                                                    | required for networked backends                               |
| `database.query_timeout`        | no                                              | `30`                                                    | positive integer seconds                                      |
| `database.slow_query_threshold` | no                                              | `500`                                                   | positive integer milliseconds                                 |
| `jwt_secret`                    | yes                                             | none                                                    | minimum 32 characters                                         |
| `jwt_access_expiry`             | no                                              | `3600`                                                  | positive integer seconds                                      |
| `jwt_refresh_expiry`            | no                                              | `604800`                                                | positive integer seconds and greater than `jwt_access_expiry` |
| `bootstrap_admin_username`      | conditional                                     | none                                                    | first-run only                                                |
| `bootstrap_admin_email`         | conditional                                     | none                                                    | first-run only, valid email                                   |
| `bootstrap_admin_password`      | conditional                                     | none                                                    | first-run only, must satisfy the password policy              |
| `cors.enabled`                  | no                                              | `true`                                                  | boolean                                                       |
| `cors.allowed_origins`          | no                                              | `["*"]`                                                 | list of allowed origins                                       |

### 8.4 Configuration Behavior

#### Server

- The service must bind to `server.host:server.port`.
- `server.prefix` must prepend every route exactly once, including public routes.
- Route prefixing must not change route semantics.

Example:

```text
server.prefix = "/api"
"/"       -> "/api"
"/health" -> "/api/health"
```

#### Logging

- All service logs, including audit logs, startup logs, and shutdown logs, must be written to both the console and the file at `server.logpath`.
- The default log file path is `/var/log/moon.log`.
- The service must open or create the configured log file during startup. If that fails, startup must fail.
- This specification does not standardize log rotation or retention behavior.

#### Database

- SQLite is the default backend.
- Adapter behavior must remain externally consistent across SQLite, PostgreSQL, and MySQL.
- Query timeout enforcement must be applied through the persistence layer.
- Slow query logging must use `database.slow_query_threshold` when configured.

#### JWT

- JWT signing and verification must use `jwt_secret`.
- Access and refresh token lifetimes must use the configured expiry values.
- Expired credentials must be rejected.

#### Bootstrap admin

- Bootstrap admin creation is first-run only.
- The bootstrap admin must be created only when no admin user exists.
- Bootstrap credentials should be removed from configuration immediately after successful initialization.

#### CORS

- If `cors.enabled` is `false`, the service must not add CORS headers.
- If `cors.enabled` is `true`, `cors.allowed_origins` controls the browser origin allowlist.
- Production deployments should use explicit origins and should not rely on wildcard origins.

## 9. Data Model and Persistence

### 9.1 Identifier and Ownership Rules

- The server must generate all record `id` values.
- Clients must not provide `id` on create operations.
- Update and destroy operations must identify the target record by `id`.
- Server-owned fields must be read-only from the client perspective.

### 9.2 Supported Field Types

| Type       | API representation | Rules                                   |
| ---------- | ------------------ | --------------------------------------- |
| `id`       | string             | read-only ULID generated by the server  |
| `string`   | string             | arbitrary text                          |
| `integer`  | number             | signed integer                          |
| `decimal`  | string             | fixed-point decimal encoded as a string |
| `boolean`  | boolean            | true or false                           |
| `datetime` | string             | RFC3339 timestamp                       |
| `json`     | object or array    | valid JSON document                     |

### 9.3 Adapter Mapping and External Invariants

| API type   | SQLite    | PostgreSQL      | MySQL           | External invariant                                  |
| ---------- | --------- | --------------- | --------------- | --------------------------------------------------- |
| `string`   | `TEXT`    | `TEXT`          | `TEXT`          | returned as JSON string                             |
| `integer`  | `INTEGER` | `BIGINT`        | `BIGINT`        | returned as JSON number                             |
| `decimal`  | `NUMERIC` | `NUMERIC(19,2)` | `DECIMAL(19,2)` | returned as JSON string without scientific notation |
| `boolean`  | `INTEGER` | `BOOLEAN`       | `BOOLEAN`       | returned as JSON boolean                            |
| `datetime` | `TEXT`    | `TIMESTAMP`     | `TIMESTAMP`     | returned as RFC3339 string                          |
| `json`     | `TEXT`    | `JSON`          | `JSON`          | returned as JSON object or array                    |

Adapter-specific storage may vary, but external API behavior must remain consistent.

### 9.4 Value Constraints

- `decimal` values must be accepted and returned as strings.
- `decimal` values must not use scientific notation or locale-specific separators.
- `decimal` scale must not exceed 10 fractional digits.
- `datetime` values must be valid RFC3339 timestamps.
- `json` values must be valid JSON objects or arrays.
- `boolean` values must be real booleans.
- `integer` values must fit the supported integer range.

### 9.5 Collection and Field Naming Rules

| Item                     | Requirement                                       |
| ------------------------ | ------------------------------------------------- |
| Collection name length   | 2 to 63 characters                                |
| Field name length        | 3 to 63 characters                                |
| Allowed pattern          | lowercase snake_case starting with a letter       |
| Uniqueness               | field names must be unique within a collection    |
| Reserved names and words | must be rejected consistently across all backends |

Collection and field naming rules must be enforced centrally so every backend behaves the same way.

In addition to reserved words, the exact collection names `users` and `apikeys` and the prefix `moon_` are reserved. Dynamic collections must not use them.

### 9.6 System Persistence Topology

Moon standardizes the system collections required for core functionality plus one internal refresh-token table. The collection list and field definitions for API-visible collections must be derived from the physical database schema at runtime rather than stored in Moon-managed metadata tables.

| Object                     | Kind                  | API-visible | Purpose                                                |
| -------------------------- | --------------------- | ----------- | ------------------------------------------------------ |
| `users`                    | system collection     | yes         | interactive identity, role, and write-capability state |
| `apikeys`                  | system collection     | yes         | machine credential metadata and authorization context  |
| `moon_auth_refresh_tokens` | internal system table | no          | refresh-session storage and rotation state             |

System-persistence rules:

- `users` and `apikeys` are the only always-present standardized system collections.
- `moon_auth_refresh_tokens` is a required internal system table for refresh-session correctness and must never surface through public APIs.
- `moon_schema_collections`, `moon_schema_fields`, and `moon_auth_jwt_revocations` are not part of Moon's canonical persistence model and must not be required for correctness.
- The reserved prefix `moon_` remains unavailable to dynamic collections and may be used only for implementation-private tables that never surface through public APIs.

### 9.7 API-Visible System Collections

Moon must maintain the `users` and `apikeys` collections at all times.

Both system collections:

- are queryable through the canonical data endpoints
- are discovered and described through runtime schema introspection
- are included in the schema registry
- must not be created, renamed, modified, or destroyed through collection schema mutation APIs

### 9.8 `users` Collection Schema

The `users` collection is the authoritative store for interactive user identity and authorization data.

```sql
-- users table for Moon system collection
CREATE TABLE users (
    id TEXT PRIMARY KEY, -- ULID, server-generated, immutable
    username TEXT NOT NULL, -- unique, 3-63 chars, lowercase snake_case
    email TEXT NOT NULL, -- unique, normalized lowercase email
    password_hash TEXT NOT NULL, -- bcrypt hash, never returned by APIs
    role TEXT NOT NULL, -- 'admin' or 'user'
    can_write BOOLEAN NOT NULL DEFAULT 0, -- default false; ignored when role=admin
    created_at TEXT NOT NULL, -- RFC3339 timestamp, immutable
    updated_at TEXT NOT NULL, -- RFC3339 timestamp, system-managed
    last_login_at TEXT, -- RFC3339 timestamp, nullable
    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE INDEX idx_users_role ON users(role);
```

Required constraints and indexes:

- primary key on `id`
- unique index on `username`
- unique index on `email`
- index on `role`

Additional rules:

- `username` comparison and uniqueness must be case-insensitive after normalization to lowercase.
- `email` comparison and uniqueness must be case-insensitive after normalization to lowercase.
- The physical row may contain internal implementation fields only if they do not change API behavior and are never exposed through public APIs.

### 9.9 `apikeys` Collection Schema

The `apikeys` collection stores machine credentials and their authorization context.

```sql
-- apikeys table for Moon system collection
CREATE TABLE apikeys (
    id TEXT PRIMARY KEY, -- ULID, server-generated, immutable
    name TEXT NOT NULL, -- unique administrative label, 3-100 chars
    role TEXT NOT NULL, -- 'admin' or 'user'
    can_write BOOLEAN NOT NULL DEFAULT 0, -- default false; ignored when role=admin
    key_hash TEXT NOT NULL, -- SHA-256 or stronger one-way hash of the raw API key; never returned by APIs
    created_at TEXT NOT NULL, -- RFC3339 timestamp, immutable
    updated_at TEXT NOT NULL, -- RFC3339 timestamp, system-managed
    last_used_at TEXT, -- RFC3339 timestamp, nullable
    CONSTRAINT apikeys_name_unique UNIQUE (name),
    CONSTRAINT apikeys_key_hash_unique UNIQUE (key_hash)
);

CREATE INDEX idx_apikeys_last_used_at ON apikeys(last_used_at);
```

Required constraints and indexes:

- primary key on `id`
- unique index on `name`
- unique index on `key_hash`
- index on `last_used_at`

Additional rules:

- Raw API key material must never be stored after creation or rotation.
- The key format must remain `moon_live_` plus a 64-character base62 suffix.
- Rotation replaces the stored credential immediately while preserving the logical API key record identified by `id`.

### 9.10 `moon_auth_refresh_tokens` Internal Table

`moon_auth_refresh_tokens` stores stateful refresh-session data for interactive user sessions.

```sql
-- moon_auth_refresh_tokens table for refresh-session state
CREATE TABLE moon_auth_refresh_tokens (
    id TEXT PRIMARY KEY, -- ULID, server-generated, immutable
    user_id TEXT NOT NULL, -- logical reference to users.id
    refresh_token_hash TEXT NOT NULL, -- SHA-256 or stronger one-way hash of the raw refresh token
    expires_at TEXT NOT NULL, -- RFC3339 timestamp, hard expiry
    created_at TEXT NOT NULL, -- RFC3339 timestamp, issue timestamp
    last_used_at TEXT, -- RFC3339 timestamp, nullable, set when the token is successfully exchanged
    revoked_at TEXT, -- RFC3339 timestamp, nullable, set when the token is invalidated
    revocation_reason TEXT -- nullable, implementation-controlled audit reason
);

CREATE UNIQUE INDEX idx_refresh_tokens_hash ON moon_auth_refresh_tokens(refresh_token_hash);
CREATE INDEX idx_refresh_tokens_user_revoked ON moon_auth_refresh_tokens(user_id, revoked_at);
CREATE INDEX idx_refresh_tokens_expires_at ON moon_auth_refresh_tokens(expires_at);
```

Required constraints and indexes:

- primary key on `id`
- unique index on `refresh_token_hash`
- index on (`user_id`, `revoked_at`)
- index on `expires_at`

Additional rules:

- Raw refresh tokens must never be stored after issuance.
- Refresh tokens are single-use credentials.
- A successful refresh must revoke the presented token and persist a replacement row.
- The table is implementation-owned and must never be exposed through collection or resource APIs.
- Deleting a user must delete or invalidate all rows with the matching `user_id`.

### 9.11 Dynamic Schema Discovery

Moon must discover API-visible collections and field definitions from the physical database schema instead of storing a Moon-managed catalog in the database.

Discovery requirements:

- collection discovery must use adapter-native SQL catalog inspection such as `SHOW TABLES`, `information_schema` queries, `PRAGMA table_list`, or equivalent backend mechanisms
- field discovery must use adapter-native column inspection queries rather than stored metadata rows
- `users` and `apikeys` must always be present in the discovered schema
- the configured database or schema namespace is owned by Moon; non-Moon application tables must not share that namespace
- tables whose names start with `moon_` are reserved and must never be exposed through public APIs
- a discovered table is API-visible only if its name satisfies section 9.5 and each exposed column can be mapped to a supported Moon field type
- if a candidate API-visible table cannot be mapped to a valid Moon schema, startup must fail
- schema discovery results must be normalized into the in-memory schema registry before the service accepts traffic

### 9.12 Dynamic Collection Physical Table Template

Every API-visible collection table, including dynamic collections, must follow this physical shape:

- `id`: required primary key, server-generated ULID, not client-writable
- user-defined fields: derived from physical columns and validated against the declared Moon field types
- unique fields: must create database-level unique constraints or unique indexes
- timestamps: system tables require explicit timestamps as defined above; dynamic collections do not receive implicit timestamps unless the API later standardizes them

Additional rules:

- The service, not the client, creates the physical `id` column for every API-visible collection.
- Physical column order or equivalent adapter metadata order must determine field order for schema responses.
- Dynamic collection tables must not use foreign keys, triggers, or hidden generated columns that change API semantics.
- Implementation-private columns may exist only if they do not change documented API behavior and are never exposed through public APIs.

### 9.13 Record Semantics and Defaults

- Field values must be validated against the active schema before persistence.
- Nullable and unique flags default to `false` when omitted in collection schema operations.
- User-defined schema default values are not supported.
- Relations between records must be managed at the application layer because Moon does not provide joins or foreign keys.
- System-managed fields such as `id`, `created_at`, `updated_at`, `password_hash`, `key_hash`, and equivalent implementation-private auth or session fields must not be client-writable.

## 10. Schema Management

### 10.1 Schema Registry Requirements

Moon must maintain an in-memory schema registry that:

- is loaded before the server accepts requests
- includes both system and dynamic collections
- is reconstructed from physical schema inspection
- stores normalized schema metadata
- is concurrency-safe
- is the runtime source of truth for validation and planning
- is refreshed atomically after successful schema mutations
- does not rely on a dedicated Moon metadata table as its source of truth

### 10.2 Schema Mutation Rules

Schema mutation behavior must follow these rules:

- each request must contain exactly one top-level mutation intent
- `op=update` must express exactly one schema sub-operation set for a collection mutation request
- all requested schema changes must be validated before persistence begins
- system collections must be protected from collection schema mutation
- implementation-private tables, including reserved `moon_*` tables, must never be addressable through schema mutation APIs
- unsupported schema features must be rejected explicitly
- cache publication must occur only after persistence succeeds

If a schema mutation fails, the service must return an error and keep the previously committed schema registry state.

### 10.3 Startup Reconciliation

On startup, Moon must reconcile the runtime schema registry with the physical database schema.

If required API-visible system collections or `moon_auth_refresh_tokens` are missing, the service must create them. If a discovered API-visible table cannot be mapped to a valid Moon schema or the physical schema cannot be reconciled safely, startup must fail rather than serve inconsistent behavior.

## 11. Query and Mutation Semantics

### 11.1 Canonical Resource Surface

Canonical resource endpoints are:

- `/data/{resource}:query`
- `/data/{resource}:mutate`
- `/data/{resource}:schema`

System collections and dynamic collections must both use this surface. Implementation-private tables, including reserved `moon_*` tables, must never use it. Additional top-level resource aliases are not required by this specification.

### 11.2 Query Rules

For collection and record queries, the service must:

- resolve the target collection from the schema registry
- reject requests for collections that do not exist
- validate query parameters before adapter execution
- reject query fields that are not present in the schema
- preserve the response contract defined by `SPEC_API.md`

Query constraints:

| Area       | Requirement                                                                 |
| ---------- | --------------------------------------------------------------------------- |
| `page`     | default `1`; must be at least `1`                                           |
| `per_page` | default `15`; maximum `200`                                                 |
| `sort`     | every sort field must exist in the target schema; `-field` means descending |
| `q`        | applies only to text-searchable fields                                      |
| `fields`   | every projected field must exist; `id` is always included                   |
| `filter`   | only operators valid for the field type are allowed                         |

Supported filter operators are `eq`, `ne`, `gt`, `lt`, `gte`, `lte`, `like`, and `in`, subject to field-type compatibility.

### 11.3 Record Mutation Rules

For record mutations, the service must enforce all of the following:

- `op` must be one of `create`, `update`, `destroy`, or `action`
- `data` must always be an array
- create items must not include `id`
- update items must include `id`
- destroy items must include `id`
- action requests must include an `action` value and an action-specific payload as defined by the API contract
- every submitted field must be validated against the active schema
- writes to server-owned or read-only fields must be rejected
- authorization must be enforced before persistence

### 11.4 Batch Semantics

Batch create, update, and destroy operations are supported for records.

Partial success is allowed for record mutations. The service must report outcomes using the response contract defined in `SPEC_API.md`. Moon must not claim multi-item transactional atomicity.

### 11.5 Collection Mutation Semantics

Collection mutations are schema operations, not record operations. They must be handled separately from record mutations and must use stricter validation because they change the rules for future requests.

## 12. Authentication and Authorization

### 12.1 Authentication Modes

Moon supports exactly two bearer credential types:

- JWT access token
- API key

Both must use:

`Authorization: Bearer <token>`

The authentication layer must distinguish credential types deterministically:

- JWT: three dot-separated segments
- API key: `moon_live_` prefix

JWT access tokens must include a unique `jti` claim so revocation checks can be performed without storing raw token material.

Malformed, mixed, or unrecognized bearer values must be rejected with the standard error response.

### 12.2 Authorization Model

Moon authorization is based on:

- endpoint protection rules
- role
- write capability

At minimum, the system must support:

- `admin`
- `user`

`admin` always implies effective write access. `can_write` controls record mutation capability for the `user` role only.

| Capability                             | `admin` | `user` with `can_write=false` | `user` with `can_write=true` |
| -------------------------------------- | ------- | ----------------------------- | ---------------------------- |
| Read public endpoints                  | yes     | yes                           | yes                          |
| Query collections metadata             | yes     | yes                           | yes                          |
| Query record data                      | yes     | yes                           | yes                          |
| Create, update, or destroy record data | yes     | no                            | yes                          |
| Mutate collection schema               | yes     | no                            | no                           |
| Manage users                           | yes     | no                            | no                           |
| Manage API keys                        | yes     | no                            | no                           |
| Perform privileged resource actions    | yes     | no                            | no                           |

Current-user endpoints apply to authenticated user sessions backed by the `users` collection, not API keys.

### 12.3 Session Rules

The system must support the session flows defined by the API contract:

- login
- refresh
- logout
- current-user read
- current-user update

Session requirements:

- multiple concurrent sessions per user are allowed
- refresh tokens are stateful credentials
- refresh tokens must be single-use
- refresh token rotation must occur on refresh
- logout must invalidate the targeted session immediately
- password change and administrative session revocation must invalidate affected sessions immediately
- access token revocation must be effective immediately, even when JWTs are otherwise stateless

Refresh-session state must be persisted in `moon_auth_refresh_tokens`. Immediate JWT revocation may use implementation-private storage, but that state must remain outside the public API surface and correctness must not depend on background workers.

### 12.4 Password Policy and User Safety

Password rules are mandatory:

- minimum length: 8 characters
- must contain at least one lowercase letter
- must contain at least one uppercase letter
- must contain at least one digit

Special characters are allowed. Passwords must be hashed with bcrypt using cost `12`.

The password policy must be enforced on:

- user creation
- current-user password change
- administrative password reset

Administrative safety rules:

- the last remaining admin must not be deleted
- the last remaining admin must not be demoted
- an admin must not change their own role
- deleting a user must also remove or invalidate that user's rows in `moon_auth_refresh_tokens` and any other implementation-private session state

### 12.5 API Key Rules

API key requirements are mandatory:

- every API key must start with `moon_live_`
- the generated secret portion must be fixed length and high entropy
- the current canonical format is a 64-character base62 suffix after the prefix, for 74 total characters
- the raw secret must be returned only at create or rotate time
- stored key material must be non-reversible and hashed at rest
- subsequent reads must return metadata only, not the raw secret
- key rotation must invalidate the previous key immediately
- API key usage should update `last_used_at` or equivalent usage metadata

API keys must carry enough authorization context to enforce role and write capability consistently with the authorization model.

## 13. Validation and Error Handling

### 13.1 Validation Requirements

The service must reject requests that violate any of the following:

- configuration: unknown, malformed, or incomplete required values
- authentication: missing, malformed, expired, revoked, or unsupported credentials
- collection selection: collection does not exist or is protected from the requested operation
- naming: invalid collection or field names
- schema: unsupported field types or unsupported schema features
- query: unsupported query parameter values or field references
- record mutation: missing required identifiers, forbidden identifiers on create, or writes to read-only fields
- authorization: insufficient role or write permission

Validation behavior must be adapter-independent.

### 13.2 Error Status Codes

Moon must use only these error status codes for error responses:

See [Standard Error Response](./SPEC/10_error.md) for any error handling

The service must not return internal error codes or additional machine-readable error metadata.

### 13.4 Error Content Rules

- Error messages must be concise and human-readable.
- Error messages must not leak secrets or internal implementation details.
- Authentication failures must be clear without revealing sensitive state.
- Error shaping must be centralized so all handlers use the same output rules.

## 14. Security, Audit Logging, and Operations

### 14.1 Transport and Secret Management

- Production deployments must use HTTPS.
- Secrets must not be committed to source control.
- Secret material must not appear in logs.
- Placeholder or sample secrets must not be used in production.
- Secret files should use restrictive file permissions.

### 14.2 Sensitive Data Redaction

The service must redact or omit sensitive values from logs, including:

- passwords
- authorization headers
- access tokens
- refresh tokens
- API keys
- JWT signing secrets
- equivalent credential or secret material

### 14.3 Rate Limiting

Moon must enforce the following rate limits:

| Traffic type                  | Limit                                         |
| ----------------------------- | --------------------------------------------- |
| login failures                | 5 attempts per 15 minutes per IP and username |
| authenticated JWT traffic     | 100 requests per minute per user              |
| authenticated API key traffic | 1000 requests per minute per key              |

Rate-limit failures must use the standard error format. Any rate-limit headers or retry metadata must be documented in `SPEC_API.md` before clients can rely on them.

### 14.4 Audit Logging

Moon must produce audit logs for security-relevant and operationally significant events.

Required audit events:

- startup success and startup failure
- configuration validation failure
- authentication success and authentication failure
- logout and token refresh
- rate-limit violations
- schema mutation attempts and outcomes
- privileged record mutations
- API key creation and API key rotation
- administrative user-management actions

Audit logs should include, when available:

- timestamp
- request method
- request path
- actor identity or anonymous state
- target collection or resource
- operation name
- outcome
- HTTP status
- request duration

Audit logs must exclude sensitive values.

Audit logging is required for correctness and observability, but a database-backed audit table is optional. The service must always emit the required audit events through the shared logger to both the console and `server.logpath`. Structured logs or an additional external log sink are acceptable as long as the required events and redaction rules are preserved.

### 14.5 Operational Best Practices

- Health endpoints must remain lightweight and public.
- Slow queries should be logged using `database.slow_query_threshold`.
- The service should fail fast on invalid startup state.
- Backups are an external operational responsibility.
- Relations between records must be enforced by the application layer, not the database.
- Production deployments should use explicit CORS origins and strong JWT secrets.
- Bootstrap admin credentials should be removed immediately after initialization.

## 15. Implementation Boundaries and Verification

The implementation must preserve these structural boundaries:

- HTTP handlers should remain thin and delegate domain logic to services.
- Adapter-specific behavior must stay inside the persistence layer.
- The schema registry must be the single runtime read path for schema validation and request planning.
- Configuration keys, default values, and path constants must be centralized in `Config.go` and referenced through named constants.
- Modules should remain single-purpose and independently testable.
- Changes to behavior governed by this specification must be accompanied by automated tests.
- No fallback behavior may silently bypass explicit validation or security rules.

This specification generally does not require a specific package tree or file layout. The exception is configuration centralization: runtime configuration keys and defaults must live in `Config.go`.

## 16. Rationale for Key Design Decisions

- Single-process service: reduces operational complexity and keeps the runtime easy to reason about
- API-managed schema: makes schema changes explicit and keeps design authority inside the public control surface
- In-memory schema registry: enables fast, centralized validation and adapter-independent planning
- Physical schema introspection instead of metadata tables: removes duplicate state and avoids drift between stored metadata and real tables
- Limited HTTP method surface: reduces protocol complexity and keeps the API predictable
- Simple error body: keeps client behavior stable and avoids fragile internal error taxonomies
- Centralized `Config.go`: eliminates magic values and keeps configuration behavior consistent across services
- No joins, migrations, or background jobs: preserves portability, keeps scope small, and reduces hidden state

## 17. Document Maintenance

`SPEC.md` must be reviewed and updated whenever a change affects:

- architecture
- runtime behavior
- schema rules
- configuration behavior
- validation rules
- authentication or authorization behavior
- security requirements
- operational guarantees

`SPEC_API.md` must be updated whenever wire-level API behavior changes. Files under `SPEC/` should be kept as aligned examples and must not redefine architecture or contradict this document.

This document should also be reviewed after production incidents, major implementation lessons, or repeated ambiguity discovered during development. Cross-document conflicts must be resolved in the same change that introduces them.

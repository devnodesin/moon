# Moon Architecture Specification

## 1. Purpose and Scope

This document defines Moon's top-level architecture and implementation model for a Go-based, API-first backend.

Moon's HTTP contract is frozen and defined by:

- `SPEC_API.md`
- `SPEC/10_error.md`
- `SPEC/20_auth.md`
- `SPEC/30_collection.md`
- `SPEC/40_resource.md`

This file defines how the system is structured to implement that contract cleanly and maintainably.

## 2. Design Principles

- Keep the runtime small, explicit, and easy to reason about.
- Favor a single service process with clear module boundaries.
- Treat API contract consistency as a first-class requirement.
- Keep features minimal and aligned with documented non-goals.
- Prefer deterministic behavior over implicit or magical behavior.

## 3. Product Model

Moon manages data as:

- **Collection** (database table)
- **Field** (table column)
- **Record** (table row)

There are two system collections:

- `users`
- `apikeys`

All other collections are dynamic and managed via collection APIs.

## 4. High-Level Architecture

Moon is a single HTTP service with modular internal components.

### 4.1 API Transport Layer

Responsibilities:

- Route requests for `GET`, `POST`, and `OPTIONS`
- Enforce route naming and action style (`resource:action`)
- Apply shared middleware (auth, rate limiting, CORS, request normalization)
- Return responses in the standardized envelope defined in API specs

### 4.2 Authentication and Authorization Layer

Responsibilities:

- Validate `Authorization: Bearer <token>`
- Support JWT and API key bearer formats
- Resolve caller identity and permissions (`user` / `admin`, write capability)
- Enforce endpoint access rules before hitting domain handlers

### 4.3 Collection Schema Layer

Responsibilities:

- Implement `/collections:query` and `/collections:mutate`
- Validate collection and field naming/type constraints
- Protect system collections from unsupported schema mutation
- Apply schema changes and publish updated schema state to cache

### 4.4 Resource Layer

Responsibilities:

- Implement `/data/{resource}:query`, `:mutate`, and `:schema`
- Validate operations (`create`, `update`, `destroy`, `action`)
- Enforce system field protections (`id` behavior, readonly fields)
- Support single and batch mutation behavior

### 4.5 Persistence Layer

Responsibilities:

- Provide database adapter abstraction for:
  - SQLite (default)
  - Postgres
  - MySQL
- Execute validated query/mutation plans
- Keep behavior uniform across adapters within API constraints

### 4.6 In-Memory Schema Cache

Responsibilities:

- Hold normalized schema metadata for system and dynamic collections
- Provide fast schema lookup for request validation and query planning
- Refresh atomically after successful schema changes
- Warm at startup before serving traffic

### 4.7 Cross-Cutting Services

- **Rate limiting:** enforce limits documented by API specs
- **CORS:** apply configured origin policy
- **Health:** expose `/health` and `/`
- **Error shaping:** ensure `{ "message": "..." }` for failures

## 5. Data and Persistence Model

### 5.1 Identifiers

- Records, users, and API keys use server-generated ULID `id`.
- Collections use `name`.

### 5.2 Supported Types

Moon supports `id`, `string`, `integer`, `decimal`, `boolean`, `datetime`, and `json` as documented in `SPEC_API.md`.

### 5.3 Dynamic Schema

- Collection definitions are mutable at runtime through API calls.
- Schema mutations are constrained to one operation kind per request.
- Collection metadata and physical schema must remain synchronized.

### 5.4 Non-Goals (Enforced)

Moon does not implement joins, triggers, background jobs, foreign keys, migrations/versioning, built-in backup/restore, WebSockets, file storage, or API versioning.

### 5.5 System Collection Baseline

The implementation must keep stable system collections for identity and service authentication.

- `users` (minimum externally visible fields): `id`, `username`, `email`, `role`, `can_write`
- `apikeys` (minimum externally visible fields): `id`, `name`

Internal security and lifecycle fields are implementation-owned but must support the frozen API flows (login/refresh/logout, password reset, session revocation, API key rotation, and one-time key disclosure on create/rotate).

## 6. Schema Lifecycle

### 6.1 Startup

1. Load configuration (`moon.conf`).
2. Initialize database adapter.
3. Ensure system collections exist.
4. Load schema metadata into in-memory cache.
5. Start HTTP server only after cache is ready.

### 6.2 Runtime Changes

For schema mutation requests:

1. Validate operation shape and naming constraints.
2. Validate compatibility with existing schema and protected system rules.
3. Apply schema change through the persistence adapter.
4. Rebuild or update cache atomically.
5. Serve subsequent requests from refreshed cache state.

### 6.3 Failure Handling

- If schema mutation fails, return an error response and keep prior cache state.
- Do not partially apply schema definitions in cache.

## 7. Request Processing Model

### 7.1 Read Path

1. Authenticate and authorize caller.
2. Resolve target collection/resource via cache.
3. Validate query options (`page`, `per_page`, `sort`, `q`, `fields`, `filter`).
4. Execute adapter query.
5. Return standardized list/get response.

### 7.2 Write Path

1. Authenticate and authorize caller.
2. Validate operation contract (`op`, `action`, `data`).
3. Validate payload fields against schema cache.
4. Execute mutation via adapter.
5. Return standardized mutation response, including success/failure counts where required.

## 8. Authentication and Session Model

- `POST /auth:session` handles `login`, `refresh`, and `logout`.
- `GET /auth:me` and `POST /auth:me` operate on the current authenticated user.
- User-specific security actions are implemented via resource actions on `users`.
- API key rotation is implemented via resource actions on `apikeys`.
- Access and refresh token expiries are configuration-driven.
- JWT claims must include identity and authorization context required by API responses and authorization checks.
- Refresh tokens are stateful credentials and must support rotation and revocation.

## 9. Configuration Model (`moon.conf`)

Moon is configured by `moon.conf` with these groups:

- **server:** `host`, `port`, optional `prefix`
- **database:** connection type and backend-specific settings
- **jwt:** secret and TTL configuration
- **bootstrap admin:** first-run credentials
- **cors:** enabled flag and allowed origins

Configuration rules:

- SQLite is the default backend when no external database is configured.
- Invalid or incomplete required config must fail startup.
- Bootstrap admin credentials are first-run only and should be removed after initial setup.

## 10. Security and Operational Rules

- Require HTTPS in production deployments.
- Never expose sensitive secret material in logs.
- Enforce role/write checks before mutations.
- Enforce documented rate limits consistently.
- Keep CORS policy explicit in production.

## 11. Implementation Boundaries

To keep implementation AI-friendly and maintainable:

- Keep modules single-purpose and independently testable.
- Keep request handlers thin; place domain logic in dedicated services.
- Keep adapter-specific behavior inside persistence layer only.
- Use schema cache as the single read path for field/collection validation.
- Preserve API compatibility with frozen API specs; track ambiguities in `REVIEW.md`.

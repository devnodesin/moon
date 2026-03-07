## Overview

- Moon's schema registry is the in-memory, concurrency-safe store of all API-visible collection schemas. Every validation and query-planning operation reads from it; it is rebuilt from physical database introspection at startup and refreshed atomically after successful schema mutations.
- No Moon-managed metadata table stores schema information; the physical database catalog is the authoritative source of truth.
- This PRD defines the schema registry data structures, population sequence, introspection queries, field-type mapping, visibility filtering, and atomic refresh semantics.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Registry Data Model

The registry must store, for each API-visible collection:

- `name`: collection name (string)
- `fields`: ordered list of field descriptors

Each field descriptor must contain:

| Property | Type | Description |
|----------|------|-------------|
| `name` | string | field name |
| `type` | string | Moon field type (`id`, `string`, `integer`, `decimal`, `boolean`, `datetime`, `json`) |
| `nullable` | bool | whether the field accepts null |
| `unique` | bool | whether the field has a unique constraint |
| `readonly` | bool | `true` for server-owned fields (`id` always; `password_hash`, `key_hash`, and equivalent system fields) |

### Startup Population

After system table initialization and before the HTTP server starts:

1. Discover all physical tables using the backend-native catalog query (SQLite: `PRAGMA table_list`; PostgreSQL: `information_schema.tables`; MySQL: `SHOW TABLES`).
2. Filter the table list to include only API-visible candidates:
   - names that satisfy collection naming rules (SPEC.md §9.5)
   - names that do not start with `moon_`
3. For each candidate table, discover its columns using the backend-native column query (SQLite: `PRAGMA table_info({table})`; PostgreSQL/MySQL: `information_schema.columns`).
4. Map each physical column to a Moon field type using the adapter mapping table (SPEC.md §9.3).
5. If any column in an otherwise-valid candidate table cannot be mapped to a supported Moon type, startup must fail with a descriptive error naming the table and column.
6. Populate the in-memory registry with the resulting collection list.

### Field Ordering

- Field order in the registry must match the physical column order as returned by the catalog query.
- The `id` field must always be first.

### System Collection Visibility Filtering

- `users` and `apikeys` must always be present in the registry.
- System-only fields must be marked `readonly = true` and must not appear in API schema responses:
  - `users`: `password_hash` is implementation-private and must not appear in schema or data responses.
  - `apikeys`: `key_hash` is implementation-private and must not appear in schema or data responses.
- `moon_auth_refresh_tokens` and any other `moon_*` table must never appear in the registry.

### `moon_` Prefix Rule

- Any table whose name begins with `moon_` must be excluded from the registry unconditionally.
- Any API request naming a `moon_*` resource must be rejected before reaching the registry.

### Concurrency Safety

- The registry must be safe for concurrent reads from multiple goroutines.
- Writes (atomic refresh) must not block readers for longer than necessary.
- A read-write mutex (`sync.RWMutex`) or equivalent mechanism must protect the registry.

### Atomic Refresh After Schema Mutation

- When a schema mutation completes successfully, the registry must be rebuilt from the physical schema immediately.
- If the rebuild fails, the previous registry state must be restored and the mutation must be rolled back or rejected.
- The refresh must be atomic: readers must see either the old complete state or the new complete state, never a partial intermediate state.
- The registry must never serve a stale schema after a successful committed mutation.

### Lookup Interface

The registry must expose at minimum:

- `Get(name string) (*Collection, bool)` — return a single collection schema by name, or false if not found.
- `List() []*Collection` — return all API-visible collections in insertion order.
- `Refresh()` error — rebuild the registry from the physical database schema.

### Validation Support

Upper layers must use the registry to:

- confirm a collection exists before executing any query or mutation.
- confirm every field name in a query or mutation payload exists in the target collection.
- determine whether a field is read-only before accepting a client write.
- determine field type before applying type-specific filter operators.

## Acceptance

- On startup against a database with `users`, `apikeys`, and a `products` table, the registry contains exactly those three collections.
- `moon_auth_refresh_tokens` is never present in the registry.
- Any table whose name starts with `moon_` is never present in the registry.
- `password_hash` does not appear in `GET /data/users:schema` response.
- `key_hash` does not appear in `GET /data/apikeys:schema` response.
- A table with an unmappable column type causes startup failure with the table and column named.
- After a successful `POST /collections:mutate` that adds a new field, the registry immediately reflects the new field for subsequent requests.
- If a schema refresh fails after a mutation, the mutation is rejected and the old registry state remains.
- Concurrent read requests during a registry refresh do not panic or return partial data.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

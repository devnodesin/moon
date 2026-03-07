## Overview

- The collection API surfaces the schema registry to clients. `GET /collections:query` lists or retrieves API-visible collections. `POST /collections:mutate` creates, updates (schema changes), or destroys dynamic collections.
- System collections (`users`, `apikeys`) are visible through queries but protected from schema mutation.
- Internal `moon_*` tables are never visible through this API.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Endpoints

| Method | Path | Required Role |
|--------|------|--------------|
| `GET` | `/collections:query` | Any authenticated |
| `POST` | `/collections:mutate` | `admin` only |

---

## `GET /collections:query`

### Query Modes

| Mode | Trigger | Behavior |
|------|---------|---------|
| List | No `name` parameter | Returns all API-visible collections with record counts |
| Get-one | `?name=<value>` | Returns the named collection with record count |

### List Response (`200 OK`)

```json
{
  "message": "Collections retrieved successfully",
  "data": [
    { "name": "users", "count": 5 },
    { "name": "apikeys", "count": 2 },
    { "name": "products", "count": 55 }
  ],
  "meta": {
    "total": 3,
    "count": 3,
    "per_page": 15,
    "current_page": 1,
    "total_pages": 1
  },
  "links": {
    "first": "/collections:query?page=1&per_page=15",
    "last": "/collections:query?page=1&per_page=15",
    "prev": null,
    "next": null
  }
}
```

- `count` is the number of records in the collection at query time.
- List results must include standard pagination `meta` and `links`.
- Internal `moon_*` tables must never appear in the list.
- Pagination applies using the standard `page` and `per_page` query parameters.

### Get-One Response (`200 OK`)

```json
{
  "message": "Collection retrieved successfully",
  "data": [
    { "name": "products", "count": 55 }
  ]
}
```

- If the named collection does not exist, return `404 Not Found`.
- Requests for `moon_*` names must return `400 Bad Request`.

### Validation Rules

- `name` query parameter selects get-one mode.
- Unknown query parameters may be ignored or rejected.
- The endpoint must only return collections visible in the schema registry.

---

## `POST /collections:mutate`

### Request Shape

```json
{
  "op": "create | update | destroy",
  "data": []
}
```

- `op` is required and must be exactly one of `create`, `update`, or `destroy`.
- `data` is required and must be a JSON array.
- Each request must contain exactly one top-level mutation intent (one `op` value).

### Protected Names

The following names must be rejected on `create`, `update`, and `destroy`:

- `users`
- `apikeys`
- Any name starting with `moon_`

Rejection response: `403 Forbidden` for system collections; `400 Bad Request` for `moon_*` names.

### Single-Intent Rule

Each request must contain exactly one mutation operation. For `op=update`, each collection item must define exactly one schema sub-operation set:

- `add_columns`
- `rename_columns`
- `modify_columns`
- `remove_columns`

Mixing sub-operation sets in the same collection item is invalid and must return `400 Bad Request`.

---

### `op=create`

#### Request

```json
{
  "op": "create",
  "data": [
    {
      "name": "products",
      "columns": [
        { "name": "title", "type": "string", "unique": true },
        { "name": "price", "type": "decimal", "nullable": true }
      ]
    }
  ]
}
```

#### Validation

- `name` must satisfy collection naming rules: 2–63 characters, lowercase snake_case starting with a letter.
- `name` must not already exist.
- `columns` is required and must be a non-empty array.
- Each column `name` must satisfy field naming rules: 3–63 characters, lowercase snake_case starting with a letter.
- Each column `name` must be unique within the collection.
- Each column `type` must be a supported Moon field type (`string`, `integer`, `decimal`, `boolean`, `datetime`, `json`).
- `nullable` defaults to `false` when omitted.
- `unique` defaults to `false` when omitted.
- The `id` column must not be declared; the server manages it implicitly.

#### Behavior

1. Validate all fields (naming rules, types, uniqueness).
2. Execute the DDL to create the physical table with the `id` primary key column plus the declared columns.
3. Refresh the schema registry atomically.
4. Return `201 Created`.

#### Response (`201 Created`)

```json
{
  "message": "Collection created successfully",
  "data": [
    {
      "name": "products",
      "columns": [
        { "name": "title", "type": "string", "nullable": false, "unique": true },
        { "name": "price", "type": "decimal", "nullable": true, "unique": false }
      ]
    }
  ],
  "meta": { "success": 1, "failed": 0 }
}
```

---

### `op=update`

#### Supported Sub-Operations

**`add_columns`**

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "add_columns": [
        { "name": "description", "type": "string", "nullable": true }
      ]
    }
  ]
}
```

**`rename_columns`**

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "rename_columns": [{ "old_name": "title", "new_name": "name" }]
    }
  ]
}
```

**`modify_columns`**

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "modify_columns": [{ "name": "price", "type": "decimal", "nullable": false }]
    }
  ]
}
```

**`remove_columns`**

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "remove_columns": ["description"]
    }
  ]
}
```

#### Validation

- The target collection must exist and must not be a system collection.
- Renamed columns must exist; new names must satisfy naming rules.
- Added columns must satisfy naming and type rules.
- Removed columns must exist and must not be `id`.
- Exactly one sub-operation set per collection item.

#### Response (`200 OK`)

```json
{
  "message": "Collection updated successfully",
  "data": [
    {
      "name": "products",
      "columns": [...]
    }
  ],
  "meta": { "success": 1, "failed": 0 }
}
```

Schema registry must be refreshed atomically after success.

---

### `op=destroy`

#### Request

```json
{
  "op": "destroy",
  "data": [{ "name": "products" }]
}
```

#### Validation

- The target collection must exist.
- Must not be `users` or `apikeys`.
- Must not start with `moon_`.

#### Behavior

1. Validate.
2. Drop the physical table.
3. Remove the collection from the schema registry.
4. Return `200 OK`.

#### Response (`200 OK`)

```json
{
  "message": "Collection destroyed successfully",
  "data": [{ "name": "products" }],
  "meta": { "success": 1, "failed": 0 }
}
```

---

### Error Conditions

| Condition | Status |
|-----------|--------|
| Non-admin caller | `403` |
| `op` missing or invalid | `400` |
| `data` missing or not array | `400` |
| Collection name violates naming rules | `400` |
| Collection already exists (create) | `409` |
| Collection not found (update/destroy) | `404` |
| System collection targeted (mutation) | `403` |
| `moon_*` name targeted | `400` |
| Column type unsupported | `400` |
| Multiple sub-operation sets in one item | `400` |

## Acceptance

- `GET /collections:query` returns `users`, `apikeys`, and any dynamic collections with their record counts.
- `moon_auth_refresh_tokens` is never in the list response.
- `GET /collections:query?name=products` returns just `products`.
- `GET /collections:query?name=moon_auth_refresh_tokens` returns `400`.
- `GET /collections:query?name=missing` returns `404`.
- `POST /collections:mutate` with `op=create` and a valid payload creates the collection and it appears in subsequent list queries.
- `POST /collections:mutate` with `op=create` targeting `users` returns `403`.
- `POST /collections:mutate` with `op=update` adding a column to a dynamic collection makes the new field queryable immediately.
- `POST /collections:mutate` with `op=destroy` removes the collection from subsequent queries.
- Mixing `add_columns` and `rename_columns` in one item returns `400`.
- A non-admin caller attempting any `POST /collections:mutate` returns `403`.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

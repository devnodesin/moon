## Overview

- `GET /data/{resource}:schema` returns the field definitions for any API-visible collection as known to the schema registry.
- System collection schemas (`users`, `apikeys`) must expose only API-visible fields; implementation-private fields must be excluded.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Endpoint

| Method | Path | Auth Required |
|--------|------|--------------|
| `GET` | `/data/{resource}:schema` | Yes (any authenticated) |

- `{resource}` must be an API-visible collection name.
- Requests where `{resource}` starts with `moon_` must return `400 Bad Request`.
- Requests for a non-existent collection must return `404 Not Found`.

### Behavior

1. Validate that `{resource}` exists in the schema registry and is not a `moon_*` name.
2. Retrieve the field list from the registry.
3. For system collections, exclude implementation-private fields before returning.
4. Return `200 OK` with the schema payload.

### Response Shape (`200 OK`)

```json
{
  "message": "Schema retrieved successfully",
  "data": [
    {
      "name": "products",
      "fields": [
        { "name": "id",       "type": "id",      "nullable": false, "unique": false, "readonly": true  },
        { "name": "title",    "type": "string",  "nullable": false, "unique": true,  "readonly": false },
        { "name": "price",    "type": "decimal", "nullable": false, "unique": false, "readonly": false },
        { "name": "details",  "type": "string",  "nullable": true,  "unique": false, "readonly": false },
        { "name": "quantity", "type": "integer", "nullable": true,  "unique": false, "readonly": false },
        { "name": "brand",    "type": "string",  "nullable": true,  "unique": false, "readonly": false }
      ]
    }
  ]
}
```

- `data` is always an array containing exactly one schema object.
- `fields` is ordered by physical column order (matching the order returned by `PRAGMA table_info`, `information_schema.columns`, or equivalent).
- `id` is always first and always has `type=id`, `readonly=true`.

### Field Descriptor Properties

| Property | Type | Description |
|----------|------|-------------|
| `name` | string | field name |
| `type` | string | Moon field type: `id`, `string`, `integer`, `decimal`, `boolean`, `datetime`, `json` |
| `nullable` | bool | whether the field accepts null values |
| `unique` | bool | whether the field has a unique constraint |
| `readonly` | bool | `true` if the field is server-managed and not client-writable |

### System Collection Field Visibility

| Collection | Excluded fields |
|------------|----------------|
| `users` | `password_hash` |
| `apikeys` | `key_hash` |

- Any field marked as implementation-private in the schema registry must not appear in the schema response for system collections.
- These rules apply even if the physical column exists in the database.

### API-Visible Fields for `users`

Included fields: `id`, `username`, `email`, `role`, `can_write`, `created_at`, `updated_at`, `last_login_at`

### API-Visible Fields for `apikeys`

Included fields: `id`, `name`, `role`, `can_write`, `created_at`, `updated_at`, `last_used_at`

### Validation Rules

- No query parameters are required or defined for this endpoint.
- Unknown query parameters may be ignored.

### Error Conditions

| Condition | Status |
|-----------|--------|
| `{resource}` starts with `moon_` | `400` |
| Collection does not exist in registry | `404` |

## Acceptance

- `GET /data/products:schema` returns `200` with a schema object containing all fields for `products` in the correct order.
- `GET /data/users:schema` returns `200` with fields for `users` but never includes `password_hash`.
- `GET /data/apikeys:schema` returns `200` with fields for `apikeys` but never includes `key_hash`.
- The `id` field is always first in the `fields` array and has `type=id` and `readonly=true`.
- `GET /data/moon_auth_refresh_tokens:schema` returns `400`.
- `GET /data/nonexistent:schema` returns `404`.
- After adding a column to a collection via `/collections:mutate`, the new field appears in the schema response for that collection.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

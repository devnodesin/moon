# Collection Management API

The collection endpoints manage API-visible collection schemas.

Rules:

- `/collections:query` lists only API-visible collections.
- Internal `moon_*` tables must never be returned.
- `users` and `apikeys` are API-visible system collections.
- `users` and `apikeys` must not be created, renamed, modified, or destroyed through `/collections:mutate`.
- Dynamic collections must not use the reserved `moon_` prefix.
- Collection schema changes must follow single-intent rules.

## `GET /collections:query`

### Modes

`GET /collections:query` supports:

1. **List mode**: no `name`
2. **Get-one mode**: `?name=...`

In list mode, the endpoint returns only API-visible collections. Internal `moon_*` tables are excluded.

When the caller uses an API key, the response must be filtered to collections listed in that key's `collections` allowlist.

### List Response

Response `200 OK`:

```json
{
  "message": "Collections retrieved successfully",
  "data": [
    { "name": "users", "count": 5, "system": true },
    { "name": "apikeys", "count": 2, "system": true },
    { "name": "products", "count": 55, "system": false }
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

### Get-One Response

Request:

`GET /collections:query?name=products`

Response `200 OK`:

```json
{
  "message": "Collection retrieved successfully",
  "data": [
    { "name": "products", "count": 55, "system": false }
  ]
}
```

If the caller uses an API key and `name` is not present in that key's `collections` allowlist, the request must be rejected.

Validation rules:

- `name` selects get-one mode.
- If the named collection does not exist, return `404 Not Found`.
- Requests for `moon_*` tables must not succeed through this endpoint.

## `POST /collections:mutate`

### Request Shape

```json
{
  "op": "create | update | destroy",
  "data": []
}
```

Rules:

- `op` is required.
- `data` is required and must be an array.
- `users` and `apikeys` must be rejected on `create`, `update`, and `destroy`.
- Internal `moon_*` names must be rejected.
- Nullable and unique default to `false` when omitted.
- The server manages the implicit `id` field for every collection. Clients must not declare, rename, modify, or remove it through this API.

### Single-Intent Rules

Each request must contain exactly one top-level mutation intent.

For `op=update`, each collection item must define exactly one schema sub-operation set:

- `add_columns`
- `rename_columns`
- `modify_columns`
- `remove_columns`

Mixing these sub-operation sets in the same collection item is invalid.

## Create Collection

### Request

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

### Response

Response `201 Created`:

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
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

## Update Collection

### Supported Update Payloads

#### Add Columns

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "add_columns": [
        { "name": "description", "type": "string", "nullable": true },
        { "name": "sku", "type": "string", "unique": true }
      ]
    }
  ]
}
```

#### Rename Columns

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "rename_columns": [
        { "old_name": "title", "new_name": "name" }
      ]
    }
  ]
}
```

#### Modify Columns

```json
{
  "op": "update",
  "data": [
    {
      "name": "products",
      "modify_columns": [
        { "name": "price", "type": "decimal", "nullable": false }
      ]
    }
  ]
}
```

#### Remove Columns

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

### Response

Response `200 OK`:

```json
{
  "message": "Collection updated successfully",
  "data": [
    {
      "name": "products",
      "columns": [
        { "name": "title", "type": "string", "nullable": false, "unique": true },
        { "name": "price", "type": "decimal", "nullable": false, "unique": false },
        { "name": "sku", "type": "string", "nullable": false, "unique": true }
      ]
    }
  ],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

## Destroy Collection

### Request

```json
{
  "op": "destroy",
  "data": [
    {
      "name": "products"
    }
  ]
}
```

### Response

Response `200 OK`:

```json
{
  "message": "Collection destroyed successfully",
  "data": [
    {
      "name": "products"
    }
  ],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

See `SPEC/10_error.md` for error handling.

---

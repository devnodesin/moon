## Overview

- `GET /data/{resource}:query` retrieves records from any API-visible collection. It supports two modes: listing records (with filtering, sorting, and pagination) and fetching a single record by `id`.
- The endpoint applies to both system collections (`users`, `apikeys`) and dynamic collections.
- Internal `moon_*` resources are never queryable.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Endpoint

| Method | Path | Auth Required |
|--------|------|--------------|
| `GET` | `/data/{resource}:query` | Yes (any authenticated) |

- `{resource}` must be an API-visible collection name.
- Requests where `{resource}` starts with `moon_` must return `400 Bad Request`.
- Requests for a non-existent collection must return `404 Not Found`.

### Query Modes

| Mode | Trigger | Behavior |
|------|---------|---------|
| List | No `id` parameter | Returns paginated records with `meta` and `links` |
| Get-one | `?id=<value>` | Returns a single record; `404` if not found |

---

## Query Parameters

All query parameters are validated before execution. Unknown parameters must be rejected.

| Parameter | Default | Rules |
|-----------|---------|-------|
| `page` | `1` | Must be ≥ 1 |
| `per_page` | `15` | Must be between 1 and 200 (inclusive) |
| `sort` | none | Comma-separated field names; prefix `-` means descending; every field must exist in the schema |
| `q` | none | Full-text search applied to text-searchable fields (`string` type only) |
| `fields` | all | Comma-separated field projection; every field must exist; `id` is always included |
| `filter` | none | Field filters; see filter rules below |

### Filter Rules

Filters apply per-field constraints on query results.

Supported operators:

| Operator | Applicable types | Description |
|----------|-----------------|-------------|
| `eq` | all | equal |
| `ne` | all | not equal |
| `gt` | `integer`, `decimal`, `datetime` | greater than |
| `lt` | `integer`, `decimal`, `datetime` | less than |
| `gte` | `integer`, `decimal`, `datetime` | greater than or equal |
| `lte` | `integer`, `decimal`, `datetime` | less than or equal |
| `like` | `string` | pattern match (case-insensitive substring) |
| `in` | all except `boolean`, `json` | value is in list |

- An operator applied to an incompatible field type must return `400 Bad Request`.
- An unknown field in a filter must return `400 Bad Request`.
- Filter values must be type-checked against the field's Moon type before execution.

### Validation Rules

- Every field name in `sort`, `fields`, or `filter` must exist in the target collection schema.
- Unknown fields must return `400 Bad Request`.
- Invalid query parameter values (e.g. `page=abc`) must return `400 Bad Request`.
- `per_page` > 200 must return `400 Bad Request`.

---

## System Collection Visibility

- `users` query results must never include `password_hash` or any implementation-private field.
- `apikeys` query results must never include `key_hash` or raw key material.
- Only API-visible fields as defined in the schema registry appear in results.

---

## List Response (`200 OK`)

```json
{
  "message": "Resources retrieved successfully",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "title": "Wireless Mouse",
      "price": "29.99",
      "quantity": 10,
      "brand": "Wow",
      "details": "Ergonomic wireless mouse"
    },
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5U",
      "title": "Mechanical Keyboard",
      "price": "89.99",
      "quantity": 5,
      "brand": "KeyPro",
      "details": "RGB backlit, blue switches"
    }
  ],
  "meta": {
    "total": 42,
    "count": 2,
    "per_page": 15,
    "current_page": 1,
    "total_pages": 3
  },
  "links": {
    "first": "/data/products:query?page=1&per_page=15",
    "last": "/data/products:query?page=3&per_page=15",
    "prev": null,
    "next": "/data/products:query?page=2&per_page=15"
  }
}
```

- `meta.total`: total matching records (before pagination).
- `meta.count`: number of records in this page.
- `meta.per_page`: the effective `per_page` value.
- `meta.current_page`: the effective `page` value.
- `meta.total_pages`: `ceil(total / per_page)`.
- `links.prev` is `null` on the first page; `links.next` is `null` on the last page.
- Links must include all active query parameters (e.g. `sort`, `filter`) in addition to `page` and `per_page`.

---

## Get-One Response (`200 OK`)

```json
{
  "message": "Resource retrieved successfully",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "title": "Wireless Mouse",
      "price": "29.99",
      "quantity": 10,
      "brand": "Wow",
      "details": "Ergonomic wireless mouse"
    }
  ]
}
```

- `data` is always an array; in get-one mode it contains exactly one item.
- If the record does not exist, return `404 Not Found`.

---

## Field Type Representation

Returned field values must conform to Moon's external type invariants regardless of backend:

| Moon type | JSON representation |
|-----------|-------------------|
| `string` | string |
| `integer` | number |
| `decimal` | string (no scientific notation) |
| `boolean` | `true` or `false` |
| `datetime` | RFC3339 string |
| `json` | object or array |

---

## Error Conditions

| Condition | Status |
|-----------|--------|
| `{resource}` starts with `moon_` | `400` |
| Collection does not exist | `404` |
| Unknown query field in `sort`, `fields`, or `filter` | `400` |
| Invalid operator for field type | `400` |
| `per_page` > 200 | `400` |
| `page` < 1 | `400` |
| `?id=<value>` record not found | `404` |

## Acceptance

- `GET /data/products:query` returns a paginated list with `meta` and `links`.
- `GET /data/products:query?id=01KJMQ3XZF5H1P2DDNGWGVXB5T` returns a single record in `data` array.
- `GET /data/products:query?id=missing` returns `404`.
- `GET /data/moon_auth_refresh_tokens:query` returns `400`.
- `GET /data/nonexistent:query` returns `404`.
- `GET /data/users:query` never includes `password_hash` in any record.
- `GET /data/apikeys:query` never includes `key_hash` or raw key material.
- `GET /data/products:query?sort=price` returns records sorted ascending by price.
- `GET /data/products:query?sort=-price` returns records sorted descending by price.
- `GET /data/products:query?sort=nonexistent` returns `400`.
- `GET /data/products:query?fields=title,price` returns only `id`, `title`, and `price`.
- `GET /data/products:query?fields=nonexistent` returns `400`.
- `GET /data/products:query?per_page=201` returns `400`.
- `decimal` fields are returned as strings with no scientific notation.
- `boolean` fields are returned as `true` or `false`, never `0` or `1`.
- `datetime` fields are returned as RFC3339 strings.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

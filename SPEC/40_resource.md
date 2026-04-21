# Resource API

The resource API manages records for one API-visible collection at a time.

Canonical routes:

- `/data/{resource}:query`
- `/data/{resource}:mutate`
- `/data/{resource}:schema`

Rules:

- These are the only canonical resource routes defined by this specification.
- `users` and `apikeys` are valid system resources on this surface.
- Internal `moon_*` tables must never be exposed on this surface.
- Additional top-level `/{resource}:...` aliases are not part of this specification.

## Resource Visibility

- Dynamic collections and API-visible system collections are addressable through `/data/{resource}:...`.
- `users` and `apikeys` must return only API-visible fields.
- System-resource schemas must not expose implementation-only fields such as `password_hash` or `key_hash`.
- Query responses must never expose raw API key material.
- Internal `moon_*` tables are never queryable, mutable, or schema-visible.

## `GET /data/{resource}:query`

### Modes

`GET /data/{resource}:query` supports:

1. **List mode**: no `id`
2. **Get-one mode**: `?id=...`

### List Response

Request:

`GET /data/products:query?page=1&per_page=15`

Response `200 OK`:

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

### Get-One Response

Request:

`GET /data/products:query?id=01KJMQ3XZF5H1P2DDNGWGVXB5T`

Response `200 OK`:

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

Validation rules:

- The target resource must exist and be API-visible.
- Unknown query fields or invalid query values must be rejected.
- Unknown records must return `404 Not Found`.

## `GET /data/{resource}:schema`

Returns the schema for one API-visible resource.

Response `200 OK`:

```json
{
  "message": "Schema retrieved successfully",
  "data": [
    {
      "name": "products",
      "fields": [
        { "name": "id", "type": "id", "nullable": false, "unique": false, "readonly": true },
        { "name": "title", "type": "string", "nullable": false, "unique": true, "readonly": false },
        { "name": "price", "type": "decimal", "nullable": false, "unique": false, "readonly": false },
        { "name": "details", "type": "string", "nullable": true, "unique": false, "readonly": false },
        { "name": "quantity", "type": "integer", "nullable": true, "unique": false, "readonly": false },
        { "name": "brand", "type": "string", "nullable": true, "unique": false, "readonly": false }
      ]
    }
  ]
}
```

System-resource rule:

- `/data/users:schema` and `/data/apikeys:schema` must include only API-visible fields.
- Fields such as `password_hash` and `key_hash` must not appear.

## `POST /data/{resource}:mutate`

### Request Shape

```json
{
  "op": "create | update | destroy | action",
  "data": [],
  "action": "required only when op=action"
}
```

Rules:

- `op` is required.
- `data` is required and must always be an array.
- `action` is required only when `op=action`.
- The target resource must exist and be API-visible.
- `moon_*` resources must be rejected.

### Operation Rules

#### `op=create`

- Each item in `data` must omit `id`.
- Client writes to read-only or server-owned fields must be rejected.
- Successful responses use `201 Created` when at least one record is created.

#### `op=update`

- Each item in `data` must include `id`.
- Client writes to read-only or server-owned fields must be rejected.

#### `op=destroy`

- Each item in `data` must include `id`.

#### `op=action`

- `action` is required.
- Each item in `data` must satisfy the documented payload requirements for that action.
- Action responses use the same mutation envelope and must include `meta.success` and `meta.failed`.

## Batch Semantics

Batch create, update, destroy, and action operations are supported.

Successful mutation responses follow these rules:

- `data` contains only successful items.
- Successful items remain in request order relative to other successful items.
- Failed items are omitted from `data`.
- `meta.success` and `meta.failed` summarize the result.

Example successful mutation envelope:

```json
{
  "message": "Mutation completed successfully",
  "data": [],
  "meta": {
    "success": 2,
    "failed": 1
  }
}
```

This contract does not add per-item error payloads to successful mutation responses.

## Create Example

Request:

```json
{
  "op": "create",
  "data": [
    {
      "title": "Wireless Mouse",
      "price": "29.99",
      "quantity": 10,
      "brand": "Wow",
      "details": "Ergonomic wireless mouse"
    }
  ]
}
```

Response `201 Created`:

```json
{
  "message": "Resource created successfully",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "title": "Wireless Mouse",
      "price": "29.99",
      "quantity": 10,
      "brand": "Wow",
      "details": "Ergonomic wireless mouse"
    }
  ],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

## Update Example

Request:

```json
{
  "op": "update",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "price": "24.99",
      "quantity": 12
    }
  ]
}
```

Response `200 OK`:

```json
{
  "message": "Resource updated successfully",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "title": "Wireless Mouse",
      "price": "24.99",
      "quantity": 12,
      "brand": "Wow",
      "details": "Ergonomic wireless mouse"
    }
  ],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

## Destroy Example

Request:

```json
{
  "op": "destroy",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T"
    }
  ]
}
```

Response `200 OK`:

```json
{
  "message": "Resource destroyed successfully",
  "data": [],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

## Action Examples

### Reset User Password

Request:

```json
{
  "op": "action",
  "action": "reset_password",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "password": "NewSecurePassword123"
    }
  ]
}
```

Response `200 OK`:

```json
{
  "message": "Action completed successfully",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T"
    }
  ],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

### Revoke User Sessions

Request:

```json
{
  "op": "action",
  "action": "revoke_sessions",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T"
    }
  ]
}
```

Response `200 OK`:

```json
{
  "message": "Action completed successfully",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T"
    }
  ],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

### Rotate API Key

Request:

```json
{
  "op": "action",
  "action": "rotate",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGW12542T"
    }
  ]
}
```

Response `200 OK`:

```json
{
  "message": "Action completed successfully",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGW12542T",
      "name": "Updated Service Name",
      "role": "user",
      "can_write": true,
      "is_website": true,
      "allowed_origins": ["https://moon.devnodes.in"],
      "rate_limit": 5,
      "captcha_required": true,
      "enabled": true,
      "key": "moon_live_I7T1uNRduazIASRIIucsgctuktM2Rk1J9O0E3ezfAaxREEgMaQBoxqJzoAY1A6Gk"
    }
  ],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

API key rule:

- Raw `key` material is returned only when an API key is created or rotated.
- Raw `key` material is never returned by query or schema endpoints.
- API key query and schema responses include `is_website`, `allowed_origins`, `rate_limit`, `captcha_required`, and `enabled`.
- When `captcha_required=true`, authenticated `POST` requests may include `captcha_id` and `captcha_value` at the top level of the JSON body.

See `SPEC/10_error.md` for error handling.

---

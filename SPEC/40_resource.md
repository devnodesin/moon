These endpoints manage records within a specific collection. Replace `{resource}` with your collection name (e.g., `products`).

The `/data/{resource}:{query, schema, mutate}` endpoints:

- System resources include `users` and `apikeys`, accessible via `/data/users` and `/data/apikeys`.
- New dynamic collections are accessible via `/data/product:{schema,query,mutate}`.
- All resources, both system and dynamic, are available under `/data/{resource}`.
- Each collection also provides its own top-level endpoints for `:query`, `:mutate`, and `:schema` (outside of `/data/`).

### `GET /{resource}:query`

**List Response (200 OK)::**

To request a list query (multiple resources), use the endpoint with optional pagination or filter parameters, for example:

`GET /data/products:query`
`GET /data/products:query?page=1&per_page=15`

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

**Single Resource Response (200 OK)::**

`GET /data/products:query?id=01KJMQ3XZF5H1P2DDNGWGVXB5T`

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

### `GET /{resource}:schema`

Response (200 OK):

```json
{
  "message": "Schema retrieved successfully",
  "data": [
    {
      "collection": "products",
      "total": 6,
      "fields": [
        {
          "name": "id",
          "type": "id",
          "nullable": false,
          "unique": false,          
          "readonly": true
        },
        {
          "name": "title",
          "type": "string",
          "nullable": false,
          "unique": true
        },
        {
          "name": "price",
          "type": "decimal",
          "nullable": false,
          "unique": false
        },
        {
          "name": "details",
          "type": "string",
          "nullable": true,
          "unique": false
        },
        {
          "name": "quantity",
          "type": "integer",
          "nullable": true,
          "unique": false
        },
        {
          "name": "brand",
          "type": "string",
          "nullable": true,
          "unique": false
        }
      ]
    }
  ]
}
```

### `POST /data/{resource}:mutate`

`POST /data/{resource}:mutate` request shape:

```json
{
  "op": "create | update | destroy | action",
  "data": [],
  "action": "optional-custom-action"
}
```

Rules:

- `data` must always be an array (single or batch).
- `op`:
  - `create`: Each object in `data` must not include the system `id`.
  - `update`: Each object in `data` must include the identifier `id`.
  - `destroy`: Each object in `data` must include the `id`; one or more objects may be provided.
- `action`: Use the `action` field to specify a required custom operation (e.g., password reset, API key rotation).

- Create, Update, and Destroy operations support both single-object and batch (multiple objects) in the `data` array.

**Create `POST /data/{resource}:mutate`**

Request: `POST /data/product:mutate`

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

Response (200 OK)::

```json
{
  "message": "Product created successfully",
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

**Update `POST /data/{resource}:mutate`**

Request: `POST /data/product:mutate`

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

Response (200 OK)::

```json
{
  "message": "Product updated successfully",
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

**Destroy `POST /data/{resource}:mutate`**

Request: `POST /data/product:mutate`

```json
{
  "op": "destroy",
  "data": [{ "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T" }]
}
```

Response (200 OK)::

```json
{
  "message": "Product deleted successfully",
  "data": [],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

- `201 Created` when `op=create` and at least one record is created.
- `200 OK` for `update`, `destroy`, `action` with at least one successful operation.
- Partial success is allowed for batch writes; report counts via `meta.success` and `meta.failed`.

See [Standard Error Response (200 OK):](10_error.md) for any error handling

### Action `POST /data/{resource}:mutate`

**Reset User Password: `POST /data/users:mutate`**

Request:

```json
{
  "op": "action",
  "action": "reset_password",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T",
      "password": "NewSecurePassword123#"
    }
  ]
}
```

Response (200 OK)::

```json
{
  "message": "Reset password successful",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T"
    }
  ]
}
```

**Revoke User Sessions: `POST /data/users:mutate`**

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

Response (200 OK)::

```json
{
  "message": "Revoke session successful",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T"
    }
  ]
}
```

**Rotate API Key: `POST /data/apikeys:mutate`**

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

Response (200 OK)::

```json
{
  "message": "API key rotated successfully",
  "warning": "Store this key securely. The old key is now invalid.",
  "data": [
    {
      "id": "01KJMQ3XZF5H1P2DDNGW12542T",
      "name": "Updated Service Name",
      "key": "moon_live_I7T1uNRduazIASRIIucsgctuktM2Rk1J9O0E3ezfAaxREEgMaQBoxqJzoAY1A6Gk"      
    }
  ]
}
```

- `key` values are returned only on creation/rotation Response (200 OK): and must not be retrievable later.

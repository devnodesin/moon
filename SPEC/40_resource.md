
These endpoints manage records within a specific collection. Replace `{resource}` with your collection name (e.g., `products`).

The `/data/{resource}:{query, schema, mutate}` endpoints:

- System resources include `users` and `apikeys`, accessible via `/data/users` and `/data/apikeys`.
- New dynamic collections are accessible via `/data/product:{schema,query,mutate}`.
- All resources, both system and dynamic, are available under `/data/{resource}`.
- Each collection also provides its own top-level endpoints for `:query`, `:mutate`, and `:schema` (outside of `/data/`).

### `GET /collections:query`

### `GET /collections:schema`

### `GET /collections:mutate`

`POST /data/{resource}:mutate` request shape:

```json
{
  "op": "create | update | destroy | action",
  "data": [],
  "action": "optional-custom-action"
}
```

Rules:

- `data` is always an array (single or batch).
- `create`: objects in `data` must not include system `id`.
- `update`: each object in `data` must include identifier (`id` or `name`).
- `destroy`: each object in `data` must include identifier only.
- `action`: use `action` (special action to perform via `action` field which mandatory).

Status behavior:

- `201 Created` when `op=create` and at least one record is created.
- `200 OK` for `update`, `destroy`, `action` with at least one successful operation.
- Partial success is allowed for batch writes; report counts via `meta.success` and `meta.failed`.
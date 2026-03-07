## Overview

- `POST /data/{resource}:mutate` creates, updates, destroys, or executes an action on records in any API-visible collection.
- Mutations support batch operations. Partial success is allowed; the response reports success and failure counts.
- System collections (`users`, `apikeys`) have additional protection and support privileged actions (`reset_password`, `revoke_sessions`, `rotate`).
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Endpoint

| Method | Path | Required Role / Capability |
|--------|------|--------------------------|
| `POST` | `/data/{resource}:mutate` | `admin` for system resource mutations and privileged actions; `can_write=true` for dynamic collection mutations |

- `{resource}` must be an API-visible collection name.
- Requests where `{resource}` starts with `moon_` must return `400 Bad Request`.
- Requests for a non-existent collection must return `404 Not Found`.

### Request Shape

```json
{
  "op": "create | update | destroy | action",
  "data": [],
  "action": "required only when op=action"
}
```

- `op` is required and must be exactly one of `create`, `update`, `destroy`, or `action`.
- `data` is required and must always be a JSON array.
- `action` is required only when `op=action`; must be omitted otherwise.

---

## `op=create`

### Rules

- Each item in `data` must not include `id`; if `id` is present, return `400 Bad Request`.
- Client writes to read-only or server-owned fields (`id`, `created_at`, `updated_at`, `password_hash`, `key_hash`, etc.) must be rejected with `400`.
- Every submitted field must exist in the collection schema; unknown fields must be rejected.
- Every field value must be type-valid for the declared Moon type.
- Unique constraints must be enforced; violations return `409 Conflict`.

### Behavior

1. Validate each item in `data`.
2. For each valid item: generate a ULID `id`, set system timestamps, insert the row.
3. Collect results; partial success is allowed.
4. Return `201 Created`.

### Response (`201 Created`)

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
  "meta": { "success": 1, "failed": 0 }
}
```

---

## `op=update`

### Rules

- Each item in `data` must include `id`; if `id` is missing, return `400 Bad Request`.
- Client writes to read-only or server-owned fields must be rejected.
- Unknown fields must be rejected.
- Field values must be type-valid.
- The target record must exist; if not, that item is counted as failed.
- Unique constraint violations are counted as failed.

### Behavior

1. Validate each item.
2. For each valid item: look up the record by `id`, apply the updates, set `updated_at`.
3. Collect results; partial success is allowed.
4. Return `200 OK`.

### Response (`200 OK`)

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
  "meta": { "success": 1, "failed": 0 }
}
```

---

## `op=destroy`

### Rules

- Each item in `data` must include `id`; if missing, return `400 Bad Request`.
- If the record does not exist, that item is counted as failed.
- When destroying a `users` record, all rows with matching `user_id` in `moon_auth_refresh_tokens` must also be deleted or invalidated.
- The last admin user must not be destroyed; the attempt must be counted as failed (return `403`-equivalent error in the response item).

### Response (`200 OK`)

```json
{
  "message": "Resource destroyed successfully",
  "data": [],
  "meta": { "success": 1, "failed": 0 }
}
```

- `data` is always an empty array for destroy.

---

## `op=action`

### Rules

- `action` field is required.
- Each item in `data` must satisfy the documented payload requirements for the named action.
- Unsupported `action` values must return `400 Bad Request`.
- Action responses use the same mutation envelope and must include `meta.success` and `meta.failed`.
- Only `admin` may perform privileged actions.

### Documented Actions

#### `action=reset_password` (resource: `users`)

Resets a user's password. Requires `admin` role.

Request item:

```json
{ "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T", "password": "NewSecurePassword123" }
```

- `password` must satisfy the password policy (min 8 chars, ≥1 uppercase, ≥1 lowercase, ≥1 digit).
- Hash new password with bcrypt cost 12.
- Invalidate all active sessions for the target user in `moon_auth_refresh_tokens`.
- Return the user id in `data`.

Response item in `data`:

```json
{ "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T" }
```

#### `action=revoke_sessions` (resource: `users`)

Revokes all active sessions for a user. Requires `admin` role.

Request item:

```json
{ "id": "01KJMQ3XZF5H1P2DDNGWGVXB5T" }
```

- Set `revoked_at = now`, `revocation_reason = "admin_revoked"` on all non-revoked rows in `moon_auth_refresh_tokens` with matching `user_id`.
- Revoke corresponding JWT `jti` values through the implementation-private revocation store.
- Return the user id in `data`.

#### `action=rotate` (resource: `apikeys`)

Rotates an API key. Requires `admin` role.

Request item:

```json
{ "id": "01KJMQ3XZF5H1P2DDNGW12542T" }
```

- Generate a new cryptographically random API key in the format `moon_live_` + 64-character base62 suffix.
- Hash the new key with SHA-256 and update `apikeys.key_hash`.
- The previous raw key is immediately invalidated.
- Return the raw key **only at rotation time**; it must never be returned again.

Response item:

```json
{
  "id": "01KJMQ3XZF5H1P2DDNGW12542T",
  "name": "Service Name",
  "role": "user",
  "can_write": true,
  "key": "moon_live_I7T1uNRduazIASRIIucsgctuktM2Rk1J9O0E3ezfAaxREEgMaQBoxqJzoAY1A6Gk"
}
```

---

## Batch Semantics

- Batch operations process all items in the submitted `data` array.
- Moon does not guarantee multi-item transactional atomicity.
- `data` in the response contains only successfully processed items, in request order.
- Failed items are omitted from `data`.
- `meta.success` counts successful items; `meta.failed` counts failures.
- Partial success (some items succeed, some fail) is allowed and returns `200 OK` (or `201 Created` if any item was created).

Example partial success response:

```json
{
  "message": "Mutation completed successfully",
  "data": [],
  "meta": { "success": 2, "failed": 1 }
}
```

---

## System Resource Mutation Rules

### `users` mutations

- Only `admin` may create, update, destroy, or perform actions on `users`.
- `password_hash` must never appear in responses.
- On destroy: cascade-invalidate `moon_auth_refresh_tokens` rows for the user.
- Last admin cannot be destroyed or demoted.

### `apikeys` mutations

- Only `admin` may create, update, destroy, or rotate `apikeys`.
- Raw `key` material is returned only at create or rotate time.
- `key_hash` must never appear in responses.
- On create: return the raw key in the response item.
- On subsequent reads: return metadata only, never raw key material.

---

## Error Conditions

| Condition | Status |
|-----------|--------|
| `{resource}` starts with `moon_` | `400` |
| Collection does not exist | `404` |
| `op` missing or invalid | `400` |
| `data` missing or not array | `400` |
| `op=action` missing `action` | `400` |
| `id` provided on create | `400` |
| `id` missing on update/destroy | `400` |
| Unknown field in payload | `400` |
| Type-invalid field value | `400` |
| Write to read-only field | `400` |
| Unique constraint violation | `409` |
| Insufficient role or `can_write` | `403` |
| Attempting to destroy last admin | `403` |

## Acceptance

- `POST /data/products:mutate` with `op=create` and valid items returns `201` with `data` containing each created record.
- `POST /data/products:mutate` with `op=create` and `id` in the item returns `400`.
- `POST /data/products:mutate` with `op=update` and a missing `id` returns `400`.
- `POST /data/products:mutate` with `op=destroy` returns `200` with empty `data` and correct counts.
- Batch create with 3 items where 2 are valid: returns `201` with 2 items in `data`, `meta.success=2`, `meta.failed=1`.
- `POST /data/users:mutate` with `op=action`, `action=reset_password` by admin: password changes and active sessions are revoked.
- `POST /data/users:mutate` with `action=revoke_sessions` by admin: all refresh tokens for the user are invalidated.
- `POST /data/apikeys:mutate` with `action=rotate` by admin: returns new raw key in response; subsequent queries do not return raw key.
- `POST /data/users:mutate` with `op=destroy` on the last admin returns an error.
- A non-admin calling `POST /data/users:mutate` returns `403`.
- A `user` with `can_write=false` calling `POST /data/products:mutate` returns `403`.
- A `user` with `can_write=true` calling `POST /data/products:mutate` with `op=create` returns `201`.
- `password_hash` never appears in any `users` mutation response.
- `key_hash` never appears in any `apikeys` mutation response.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

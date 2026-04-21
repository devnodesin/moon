# Moon API Specification

## Scope

This document defines the wire-level contract for the Moon HTTP API.

`SPEC.md` is the architectural source of truth. This document defines the externally visible HTTP behavior and must remain consistent with it.

## Core Rules

- Only `GET`, `POST`, and `OPTIONS` are supported.
- Any other HTTP method must return `405 Method Not Allowed`.
- Only `/` and `/health` are public.
- All other routes require authentication unless this document explicitly states otherwise.
- Canonical resource routes are:
  - `/data/{resource}:query`
  - `/data/{resource}:mutate`
  - `/data/{resource}:schema`
- Internal system tables use the `moon_` prefix and must never be exposed through collection or resource APIs.
- API-visible system collections are `users` and `apikeys`.
- Collection schema mutation APIs must not create, rename, modify, or destroy `users` or `apikeys`.
- Error responses always use `{ "message": "..." }` only.

## Terminology

Moon uses the following terms:

- **Collection**: a table
- **Field**: a column
- **Record**: a row

Identifiers:

- Records, users, and API keys use server-generated ULID `id`
- Collections use `name`

## Authentication Model

Moon supports exactly two bearer credential types:

- JWT access token
- API key

Both use:

`Authorization: Bearer <token>`

Credential rules:

- JWT access tokens are used for interactive user sessions.
- API keys are used for service access.
- Website API keys are browser-facing API keys and must enforce a matching `Origin` header from their `allowed_origins` list.
- Disabled API keys must be rejected.
- JWT access tokens must include a unique `jti` claim.
- Malformed, expired, revoked, or unsupported bearer credentials must be rejected with the standard error body.
- `/auth:session` is the credential-exchange endpoint. It does not require a bearer token.
- `GET /auth:me` and `POST /auth:me` require a JWT bearer token.
- API keys must not be accepted on `/auth:me`.

## Standard Success Responses

Success responses use one of the following documented shapes.

### Standard Success Envelope

Use this envelope when the endpoint returns a resource payload:

```json
{
  "message": "Request completed successfully",
  "data": []
}
```

Rules:

- `message` is always present.
- When present, `data` is always an array.
- `meta` is present only when defined by the endpoint.
- `links` is present only when defined by the endpoint.

### List Query Success

List query endpoints return pagination metadata:

```json
{
  "message": "Resources retrieved successfully",
  "data": [],
  "meta": {
    "total": 42,
    "count": 15,
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

### Mutation Success

Mutation endpoints return mutation counts:

```json
{
  "message": "Mutation completed successfully",
  "data": [],
  "meta": {
    "success": 1,
    "failed": 0
  }
}
```

Rules:

- Mutation responses always include `meta.success` and `meta.failed`.
- This applies to create, update, destroy, and action responses.
- `201 Created` is used when at least one resource or collection is created.
- `200 OK` is used for successful get, list, update, destroy, action, and schema responses.

### Message-Only Success

Message-only success responses are allowed only when an endpoint intentionally returns no resource payload.

Documented exception:

- `POST /auth:session` with `op=logout`

```json
{
  "message": "Logged out successfully"
}
```

## Standard Error Response

All errors use this shape:

```json
{
  "message": "A human-readable description of the error"
}
```

Documented error statuses: See [Standard Error Response](./SPEC/10_error.md)

## CAPTCHA Challenge Response

When an authenticated API key has `captcha_required=true` and a `POST` request is missing or fails CAPTCHA validation, Moon returns `403 Forbidden` with this shape:

```json
{
  "message": "Captcha required",
  "captcha": {
    "id": "01KTESTCAPTCHA1234567890AB",
    "image_base64": "PHN2ZyB4bWxucz0iLi4uIj48L3N2Zz4=",
    "expires_in": 300
  }
}
```

Request rules:

- Clients retry the original `POST` request and include `captcha_id` and `captcha_value` in the top-level JSON body.
- CAPTCHA challenges are single-use and expire after the documented lifetime.

## Endpoint Surface

### Public Endpoints

| Endpoint  | Method | Description        |
| --------- | ------ | ------------------ |
| `/`       | GET    | Alias of `/health` |
| `/health` | GET    | Service health     |

This document does not standardize public health response bodies beyond normal HTTP success semantics.

**Check Health: `GET /auth:me`**

Response `200 OK`:

```json
{
  "data": {
    "moon": "1.99",
    "timestamp": "2026-03-01T12:48:49Z"
  }
}
```

- The response will include `moon: {version}`, where the version is loaded from Config.go.
- The response will also include `timestamp: {current UTC time in RFC3339 timestamp}`.

### Authentication Endpoints

| Endpoint        | Method | Description                                           |
| --------------- | ------ | ----------------------------------------------------- |
| `/auth:session` | POST   | Unified session actions: `login`, `refresh`, `logout` |
| `/auth:me`      | GET    | Get the current authenticated user                    |
| `/auth:me`      | POST   | Update the current authenticated user                 |

See [Authentication API](./SPEC/20_auth.md)

### Collection Managment Endpoints

| Endpoint              | Method | Description                            |
| --------------------- | ------ | -------------------------------------- |
| `/collections:query`  | GET    | List collections or get one by `name`  |
| `/collections:mutate` | POST   | Create, update, or destroy collections |

See [Collection Managment API](./SPEC/30_collection.md)

### Resource Endpoints

| Endpoint                  | Method | Description                               |
| ------------------------- | ------ | ----------------------------------------- |
| `/data/{resource}:query`  | GET    | List records or get one by `id`           |
| `/data/{resource}:mutate` | POST   | Create, update, destroy, or run an action |
| `/data/{resource}:schema` | GET    | Read the resource schema                  |

See `SPEC/40_resource.md`.

See [Resource API](./SPEC/40_resource.md)

## Query Modes

### Collection Query Modes

`GET /collections:query` supports:

1. **List mode**: no `name`
2. **Get-one mode**: `?name=...`

### Resource Query Modes

`GET /data/{resource}:query` supports:

1. **List mode**: no `id`
2. **Get-one mode**: `?id=...`

## Query Options

Unless a more specific endpoint contract says otherwise, query endpoints use these options:

| Parameter  | Rules                                                                                                       |
| ---------- | ----------------------------------------------------------------------------------------------------------- |
| `page`     | Default `1`; must be at least `1`                                                                           |
| `per_page` | Default `15`; maximum `200`                                                                                 |
| `sort`     | Comma-separated fields; `-field` means descending                                                           |
| `q`        | Full-text search across text-searchable fields only                                                         |
| `fields`   | Comma-separated field projection; every field must exist; `id` is always included for record queries        |
| `filter`   | Field filters using `eq`, `ne`, `gt`, `lt`, `gte`, `lte`, `like`, `in`, subject to field-type compatibility |

Validation rules:

- Unknown fields in `sort`, `fields`, or `filter` must be rejected.
- Invalid query values must be rejected.
- Query parameters are validated before execution.
- Collection and resource names that start with `moon_` are invalid on public APIs.

## Visibility Rules

- `users` and `apikeys` are API-visible system collections.
- Internal `moon_*` tables are never API-visible.
- System-resource schemas must not expose implementation-only fields such as `password_hash`, `key_hash`, or any internal `moon_*` structures.
- Resource queries for system collections must return only API-visible fields.

---

## Overview

- This PRD defines the HTTP server, route table, middleware chain, response-shaping rules, CORS handling, and the public health endpoints. It establishes the structural skeleton that every other feature depends on.
- The service binds to `server.host:server.port`, applies `server.prefix` to all routes, and processes each request through a fixed middleware pipeline before reaching a handler.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### HTTP Server

- The service must use Go's standard `net/http` package.
- It must bind to the address `server.host:server.port` as configured.
- Graceful shutdown must be supported: on SIGTERM or SIGINT the server must stop accepting new connections and wait for in-flight requests to complete before exiting.

### Route Table

All routes below are relative to `server.prefix` (which may be empty). If `server.prefix` is set to `/api`, then `/health` becomes `/api/health`, and so on.

| Method | Path | Visibility | Handler |
|--------|------|-----------|---------|
| `GET` | `/` | public | alias of `/health` |
| `GET` | `/health` | public | health check |
| `POST` | `/auth:session` | public | session operations |
| `GET` | `/auth:me` | protected | current user |
| `POST` | `/auth:me` | protected | update current user |
| `GET` | `/collections:query` | protected | list/get collections |
| `POST` | `/collections:mutate` | protected | schema mutation |
| `GET` | `/data/{resource}:query` | protected | list/get records |
| `POST` | `/data/{resource}:mutate` | protected | record mutation |
| `GET` | `/data/{resource}:schema` | protected | resource schema |

- Only `GET`, `POST`, and `OPTIONS` are accepted. Any other method must return `405 Method Not Allowed` with the standard error body `{ "message": "Method not allowed" }`.
- Unknown routes must return `404 Not Found` with the standard error body.

### Route Prefix

- `server.prefix` must prepend every route exactly once, including public routes.
- Route prefixing must not change route semantics.
- Example: `server.prefix = "/api"` → `/health` becomes `/api/health`, `/auth:session` becomes `/api/auth:session`.

### Middleware Chain (in order)

Requests must pass through middleware in this exact sequence:

1. **Route and prefix resolution** — determine the matched route; reject unknown routes (404) and unsupported methods (405) before any further processing.
2. **CORS handling** — apply CORS headers based on `cors.enabled` and `cors.allowed_origins`; respond to `OPTIONS` preflight immediately.
3. **Audit logging context creation** — attach a request-scoped context for correlation (request ID, start time); this context must be available to all subsequent middleware and handlers.
4. **Authentication** — for protected routes only: extract and validate the bearer credential; reject missing or invalid credentials with `401`.
5. **Rate limiting** — apply per-IP, per-user, or per-key rate limits; reject excess requests with `429`.
6. **Authorization** — enforce role and write-capability rules; reject unauthorized requests with `403`.
7. **Handler and service execution** — delegate to the appropriate handler.
8. **Response shaping** — emit the final response with correct status code and body.

### CORS Handling

- If `cors.enabled = false`, the service must not add any CORS headers.
- If `cors.enabled = true`:
  - `Access-Control-Allow-Origin` must be set to the matched allowed origin or the wildcard when `allowed_origins = ["*"]`.
  - Preflight `OPTIONS` requests must receive a `200 OK` response with CORS headers and an empty body immediately, before reaching authentication or subsequent middleware.

### Response Shaping Rules

All responses must conform to the envelopes defined in `SPEC_API.md`. Response shaping is the final middleware step and is applied to every response including errors.

#### Success Responses

- `message` is always present.
- When present, `data` is always an array.
- `meta` and `links` are present only when defined by the endpoint contract.
- `201 Created` is used when at least one resource or collection is created.
- `200 OK` is used for all other success responses.

#### Error Responses

All errors use this shape exactly:

```json
{ "message": "A human-readable description of the error" }
```

No additional fields are allowed in error responses.

#### Standard HTTP Status Codes

| Status | Condition |
|--------|-----------|
| `200` | success (non-create) |
| `201` | at least one resource created |
| `400` | bad request / validation failure |
| `401` | missing or invalid authentication |
| `403` | insufficient authorization |
| `404` | resource or collection not found |
| `405` | unsupported HTTP method |
| `409` | conflict (duplicate name, unique violation) |
| `422` | unprocessable entity (schema constraint violation) |
| `429` | rate limit exceeded |
| `500` | internal server error |

### Public Health Endpoints

- `GET /` must behave identically to `GET /health`.
- `GET /health` must return `200 OK` with a minimal success body.
- Health endpoints must not require authentication.
- Health endpoints must remain lightweight; they must not execute database queries.

### Error Centralization

- Error serialization must be centralized in a single helper; handlers must not construct raw JSON error strings.
- Panic recovery middleware must catch unhandled panics, log them (without sensitive context), and return `500`.

### `{resource}` Path Variable

- `{resource}` in `/data/{resource}:...` routes is a URL path segment.
- The value must be extracted and passed to the handler before middleware completes.
- Requests where `{resource}` starts with `moon_` must be rejected with `400 Bad Request`.

## Acceptance

- `GET /health` returns `200 OK` with no authentication required.
- `GET /` returns the same response as `GET /health`.
- `DELETE /health` returns `405 Method Not Allowed` with `{ "message": "Method not allowed" }`.
- `PUT /auth:session` returns `405`.
- `GET /unknown-route` returns `404` with `{ "message": "..." }`.
- `GET /data/moon_secret:query` returns `400`.
- With `server.prefix = "/api"`, `GET /api/health` returns `200`; `GET /health` returns `404`.
- With `cors.enabled = true` and `OPTIONS /auth:session`, the response is `200` with CORS headers and an empty body.
- With `cors.enabled = false`, no CORS headers appear in any response.
- Unhandled panics in handlers return `500` and produce a log line; the process continues.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

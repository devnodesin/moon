## Overview
- What is this and why
- Overview
- Overview

This PRD addresses a production bug where CORS preflight (OPTIONS) requests to dynamic endpoints (for example `/thalib:schema`) return HTTP 405 Method Not Allowed because those routes were registered without the CORS middleware. The root cause is routing registration that bypasses the CORS middleware for dynamic/catch-all handlers. The goal is to ensure all endpoints that require browser access and CORS support are registered with the correct CORS wrapper, and that OPTIONS preflight requests are handled consistently and predictably.

Context:
- Moon supports only `GET`, `POST`, and `OPTIONS` (preflight) for browser clients. Preflight `OPTIONS` must return `204 No Content` with the appropriate `Access-Control-*` headers when the origin and headers/methods are allowed.
- Public endpoints (`/health`, `/doc/*`) must always respond with `Access-Control-Allow-Origin: *`.

High-level solution summary:
- Audit routing registration to ensure every API endpoint that can be called from a browser is wrapped with the CORS middleware (`CORSMiddleware.Handle` for normal endpoints or `CORSMiddleware.HandlePublic` for public endpoints).
- Ensure the CORS middleware returns `204 No Content` for preflight `OPTIONS` requests and sets `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials` (where applicable) and `Access-Control-Max-Age`.
- Add automated tests and a short runtime check to prevent regressions.

## Requirements
- What must it do?

1) Functional requirements
- FR-1: OPTIONS preflight handling: All endpoints that accept cross-origin requests must respond to `OPTIONS` preflight requests with `204 No Content` and appropriate `Access-Control-*` headers when the `Origin` header is allowed.
- FR-2: Route registration: Dynamic/catch-all routes (the AIP-136 `{collection}:{action}` pattern) MUST be registered with the standard CORS middleware wrapper so preflight requests never reach the dynamic handler logic which returns `405`.
- FR-3: Public endpoints: The `/health` and `/doc/*` endpoints MUST use public CORS behavior and return `Access-Control-Allow-Origin: *` for both normal and preflight requests.
- FR-4: Preflight unauthenticated: Preflight `OPTIONS` requests MUST not require authentication or rate-limiting and must be short-circuited by CORS middleware.
- FR-5: Config-driven policy: CORS behavior must be driven by `CORSConfig` (enabled, allowed_origins, allowed_methods, allowed_headers, allow_credentials, max_age). If CORS is disabled globally, preflight handling for public endpoints still applies per PRD rules.

2) Technical requirements
- TR-1: Middleware contract: The CORS middleware must provide two entry-points: `Handle` (standard CORS for protected endpoints) and `HandlePublic` (public endpoints with `Access-Control-Allow-Origin: *`). Both must short-circuit `OPTIONS` with `204` when appropriate.
- TR-2: Route-level enforcement: All calls to `mux.HandleFunc` that register API or dynamic handlers must be wrapped with the appropriate CORS middleware wrapper before logging, auth, or other middleware where preflight should not be blocked.
- TR-3: No-auth for OPTIONS: Authorization and rate-limit middlewares MUST allow `OPTIONS` and should not block preflight requests (i.e., they must check `r.Method == http.MethodOptions` and bypass auth/rate-limit checks or be applied after CORS middleware).
- TR-4: Tests: Unit tests must cover:
  - OPTIONS on a dynamic endpoint returns `204` with the expected headers.
  - GET/POST on the same endpoint return their normal status and include `Access-Control-Allow-Origin` when origin is allowed.
  - Public endpoints return `Access-Control-Allow-Origin: *` for GET and OPTIONS.
- TR-5: Runtime safety: On startup, warn or fail if `CORSConfig.AllowedOrigins` contains `*` while `AllowCredentials` is true (CORS spec incompatibility).

3) API specifications (endpoints impacted)
- Dynamic endpoints (prefix + `/{collection}:{action}`)
  - Example: `GET /{prefix}/thalib:schema` — requires `Access-Control-Allow-Origin` on actual GETs; preflight: `OPTIONS /{prefix}/thalib:schema` must return `204` with `Access-Control-Allow-Methods: GET, OPTIONS` (or configured methods) and `Access-Control-Allow-Headers` matching allowed headers.
- Auth endpoints
  - `POST {prefix}/auth:login` and `POST {prefix}/auth:refresh` — `OPTIONS` must return `204` when preflighted.
- Admin and management endpoints
  - `GET/POST {prefix}/users:*`, `{prefix}/apikeys:*`, `{prefix}/collections:*` — all must support `OPTIONS` via CORS middleware.
- Public endpoints
  - `GET {prefix}/health`, `GET {prefix}/doc/*` and their `OPTIONS` — must use `HandlePublic` behavior and return `Access-Control-Allow-Origin: *`.

Validation rules and constraints
- VR-1: If `CORSConfig.Enabled == false`, public endpoints still expose `Access-Control-Allow-Origin: *` per product rules. All other endpoints will have no CORS headers (browsers will block cross-origin calls).
- VR-2: `AllowedMethods` may contain only `GET`, `POST`, and `OPTIONS` (server does not implement other methods). Including other methods has no effect.
- VR-3: `Access-Control-Allow-Origin` must echo the request `Origin` when `AllowedOrigins` contains the exact origin; `*` must be used only for `HandlePublic` responses. If `AllowCredentials` is true, do NOT return `*`; return the explicit origin.

Error handling and failure modes
- EM-1: If a preflight `OPTIONS` reaches a dynamic handler (i.e., not short-circuited by CORS middleware), the handler may return `405`. This must be prevented by route registration changes and tests.
- EM-2: If `AllowedOrigins` is empty, the server must not set any `Access-Control-*` response header and browsers will block cross-origin requests.
- EM-3: If configuration is invalid (e.g., `AllowedOrigins == ["*"]` and `AllowCredentials == true`), the server should log an error and either fail fast on startup or normalize the configuration and emit a loud warning; define behavior in implementation notes.

Filtering, sorting, permissions, and limits
- Not applicable to CORS. Permissions: CORS does not grant authentication — server authorization remains in place for actual requests.

## Acceptance
- How do we know it’s done?

1) Verification steps for each major requirement
- AC-1: Preflight for dynamic endpoint
  - Run:
    ```bash
    curl -i -X OPTIONS 'http://localhost:8080/thalib:schema' \
      -H 'Origin: https://example.com' \
      -H 'Access-Control-Request-Method: GET'
    ```
  - Expected: `HTTP/1.1 204 No Content` and headers include `Access-Control-Allow-Origin: https://example.com` (if configured), `Access-Control-Allow-Methods: GET, OPTIONS`, and `Access-Control-Allow-Headers` per config.

- AC-2: Actual GET includes CORS header
  - Run:
    ```bash
    curl -i -X GET 'http://localhost:8080/thalib:schema' -H 'Origin: https://example.com'
    ```
  - Expected: 200 with `Access-Control-Allow-Origin: https://example.com` when origin allowed.

- AC-3: Public endpoints use wildcard origin
  - Run `curl -I -X OPTIONS http://localhost:8080/health -H 'Origin: https://other.com'` and expect `Access-Control-Allow-Origin: *` and `204`.

- AC-4: No auth for OPTIONS
  - Ensure that sending OPTIONS without auth (no Authorization header) returns `204` and not `401` or `429` for endpoints that are normally protected.

2) Test scenarios or scripts
- Unit tests (Go `*_test.go`):
  - Test that `CORSMiddleware.Handle` returns `204` for `OPTIONS` when `Origin` is allowed.
  - Test that `CORSMiddleware.HandlePublic` returns `204` and `Access-Control-Allow-Origin: *` for public paths.
  - Test that dynamic route registrations are wrapped with `CORSMiddleware` by inspecting the server setup via a small integration test that registers a fake handler and asserts `OPTIONS` behavior.

3) Expected API responses (examples)
- Preflight success (allowed origin):

  Status: 204 No Content
  Headers:
  - Access-Control-Allow-Origin: https://example.com
  - Access-Control-Allow-Methods: GET, OPTIONS
  - Access-Control-Allow-Headers: Content-Type, Authorization, X-API-Key
  - Access-Control-Max-Age: 3600

- Preflight disallowed origin:

  Status: 204 No Content (or 204 with no CORS headers) — but browsers will block the request because no `Access-Control-Allow-Origin` header present. Implementation choice: middleware should return 204 with no CORS headers when origin not allowed, mirroring PRD-049 guidance.

4) Edge cases and negative paths
- EC-1: Missing `Origin` header: CORS middleware should not treat the request as cross-origin; normal flow should continue — do not add CORS headers.
- EC-2: `AllowedOrigins` empty: Preflight returns 204 with no CORS headers; browsers will block.
- EC-3: Non-preflight `OPTIONS` usage by other clients: If an API genuinely needs `OPTIONS` for some custom functionality, ensure that route-specific handlers override middleware behavior or that CORS middleware is configurable per-route.

## Implementation Notes (developer guidance)
- Audit routing registration at `cmd/moon/internal/server/server.go`. Ensure `mux.HandleFunc` registrations use either `s.corsMiddle.Handle(...)` or `s.corsMiddle.HandlePublic(...)` wrappers as appropriate. The dynamic catch-all route must be wrapped.
- Ensure `corsPreflightHandler` exists as a no-op handler for explicit `OPTIONS` routes, and that it is wrapped with the CORS middleware so headers are set and `204` returned.
- Ensure auth and rate-limit middlewares skip `OPTIONS` early or are applied after CORS middleware so they do not block preflight.
- Add unit tests in `cmd/moon/internal/server/server_test.go` and middleware tests in `cmd/moon/internal/middleware/cors_test.go` covering preflight behavior.
- Add a CI gating test that runs the new tests to prevent regressions.

Assumptions / Needs Clarification:
- NC-1: Exact list of endpoints that should be considered "browser-accessible" (I assumed all documented API endpoints and dynamic collection endpoints). If some internal endpoints should remain CORS-disabled, list them explicitly.
- NC-2: Startup policy on invalid CORS config (`*` with credentials). I recommend failing startup to avoid unsafe configuration, but the team may prefer a warning and normalization.

--

Checklist (required):
- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

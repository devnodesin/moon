## Overview

- Moon must enforce rate limits to protect against brute-force attacks and API abuse. Three distinct rate limits apply: login failure limiting per IP+username, and per-caller request limits for JWT and API key traffic.
- Rate limiting is enforced in the middleware chain, after authentication and before authorization.
- All rate-limit violations must return `429 Too Many Requests` with the standard error body.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Rate Limit Table

| Traffic type | Limit | Window | Key |
|-------------|-------|--------|-----|
| Login failures | 5 attempts | 15 minutes | IP address + username (case-insensitive) |
| Authenticated JWT requests | 100 requests | 1 minute | user id (from JWT `sub` claim) |
| Authenticated API key requests | 1000 requests | 1 minute | api key id (from `apikeys.id`) |

### Login Failure Rate Limiting

- Applies to `POST /auth:session` with `op=login`.
- Counted only on authentication failure (wrong password or unknown username).
- Successful logins do not increment the failure counter.
- The rate-limit key is the combination of the client IP address and the submitted `username` (normalized to lowercase).
- After 5 failures within a 15-minute sliding window, return `429 Too Many Requests`.
- The counter must reset automatically after the window expires.
- On `429`, the response body must be the standard error body: `{ "message": "Too many requests" }`.

### JWT Request Rate Limiting

- Applies to all authenticated requests where the bearer credential is a JWT.
- The rate-limit key is the user id extracted from the JWT `sub` claim.
- Limit: 100 requests per 60-second sliding window.
- On limit exceeded, return `429 Too Many Requests` with the standard error body.

### API Key Request Rate Limiting

- Applies to all authenticated requests where the bearer credential is an API key.
- The rate-limit key is the api key id from the resolved `apikeys` row.
- Limit: 1000 requests per 60-second sliding window.
- On limit exceeded, return `429 Too Many Requests` with the standard error body.

### Implementation Requirements

- Rate-limit state may be stored in-memory. Correctness must not depend on database storage.
- The implementation must be concurrency-safe.
- A sliding window or token-bucket algorithm is acceptable.
- Rate-limit state does not need to persist across service restarts.
- All limit values (counts, window durations) must be defined as named constants in `Config.go`; no magic values.

### Headers

- No rate-limit response headers (e.g. `Retry-After`, `X-RateLimit-*`) are guaranteed by this specification.
- If headers are added in a future revision, they must be documented in `SPEC_API.md` before clients can rely on them.

### Audit Logging

- Every rate-limit violation must produce an audit log entry with:
  - `event`: `rate_limit_violation`
  - `limit_type`: one of `login_failure`, `jwt_traffic`, `apikey_traffic`
  - `actor`: the rate-limit key (IP+username for login; user id for JWT; api key id for API key)
  - `timestamp`: RFC3339

### Error Response (`429 Too Many Requests`)

```json
{
  "message": "Too many requests"
}
```

## Acceptance

- 5 consecutive failed login attempts from the same IP + username within 15 minutes: the 5th fails with `401`, the 6th attempt (same window) returns `429`.
- A successful login resets the failure counter for that IP + username combination.
- After 15 minutes have elapsed from the first failure, the counter resets and login attempts are accepted again.
- 101 authenticated JWT requests from the same user within 60 seconds: the 101st returns `429`.
- 1001 authenticated API key requests from the same key within 60 seconds: the 1001st returns `429`.
- Each `429` response body is `{ "message": "Too many requests" }`.
- Each rate-limit violation produces an audit log entry with the correct `limit_type`.
- Rate-limit counters are concurrency-safe: simultaneous requests from multiple goroutines do not cause data races.
- All limit values are defined in `Config.go`; no magic numbers appear in rate-limiting code.
- `go vet ./cmd/...` reports zero issues.
- `go test -race ./cmd/...` passes without race conditions.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

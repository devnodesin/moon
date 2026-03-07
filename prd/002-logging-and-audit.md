## Overview

- Moon must write all logs — startup, shutdown, audit, request, and error — to both the console (stdout/stderr) and a file simultaneously. If the log file cannot be opened, the service must refuse to start.
- Sensitive values such as passwords, tokens, API keys, and secrets must never appear in any log line.
- This PRD covers the logging infrastructure: the dual writer, log levels, sensitive-value redaction, and the required set of audit events that must be emitted.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Dual Log Writer

- All log output must be written to both the console (stdout) and the file at `server.logpath` simultaneously.
- The default log file path is `/var/log/moon.log` (defined as a constant in `Config.go`).
- During startup the service must open or create the configured log file. If that operation fails, startup must fail with a descriptive error.
- The log writer must be initialized before any other component emits log output.
- Log rotation and retention are not part of this specification; the log file is append-only.

### Log Format

- Each log line must include at minimum: timestamp (RFC3339), log level, and message.
- Structured key-value pairs are acceptable as long as they are consistently formatted.
- The format must be consistent across all components.

### Sensitive-Value Redaction

The following values must never appear in any log line, in any form:

- passwords (plaintext or hashed)
- authorization header values
- JWT access tokens
- refresh tokens
- raw API keys
- JWT signing secrets (`jwt_secret`)
- any equivalent credential or secret material

Redaction rules:

- Values must be omitted or replaced with a fixed placeholder (e.g. `[REDACTED]`).
- The presence of these fields in a request may be logged (e.g. `authorization: present`) but not their values.
- Error messages that would reveal secret material must be sanitized before logging.

### Required Audit Events

The service must emit a structured log entry for each of the following events:

| Event | Fields to include |
|-------|-------------------|
| Startup success | timestamp, service version if available |
| Startup failure | timestamp, reason (no secrets) |
| Configuration validation failure | timestamp, failing key name |
| Authentication success | timestamp, actor identity, credential type (jwt/apikey) |
| Authentication failure | timestamp, credential type attempted, reason |
| Logout | timestamp, actor identity |
| Token refresh | timestamp, actor identity |
| Rate-limit violation | timestamp, actor identity or IP, limit type |
| Schema mutation attempt | timestamp, actor identity, target collection, op |
| Schema mutation outcome | timestamp, actor identity, target collection, op, outcome |
| Privileged record mutation | timestamp, actor identity, target collection, op |
| API key creation | timestamp, actor identity, new key id |
| API key rotation | timestamp, actor identity, key id |
| Administrative user-management action | timestamp, actor identity, target user id, action name |
| Shutdown | timestamp |

Audit event fields when available:

- `timestamp` (RFC3339)
- `method` (HTTP method)
- `path` (request path, without query string)
- `actor` (user id, api key id, or `anonymous`)
- `target` (collection or resource name)
- `op` (operation name)
- `outcome` (`success` or `failure`)
- `status` (HTTP status code)
- `duration_ms` (request duration in milliseconds)

### Logging in Request Handlers

- Every inbound request must produce at minimum one log line: method, path, status, duration.
- The authentication middleware must log auth success and failure events using the audit event format.
- No handler may log raw credential values.

### Log File Failure During Runtime

- If the log file becomes unwritable during runtime, the service must log the error to console and continue serving. Dropping the file writer at runtime is acceptable; failing the process is not required.

## Acceptance

- On startup, the service creates or opens the file at `server.logpath` and writes an initial startup message to both console and file.
- If `server.logpath` points to a non-writable location, startup fails with a descriptive error.
- Sending a valid login request produces a log line on both console and file containing the auth success event fields.
- Sending a request with an invalid bearer token produces a log line containing the auth failure event — no token value appears.
- Performing a schema mutation produces two log lines: one for attempt, one for outcome.
- No log line in any test scenario contains a raw password, token, API key, or JWT secret.
- All required audit event types can be triggered by exercising the corresponding API operations and verified in the log output.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

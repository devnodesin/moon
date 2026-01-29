## SPEC Compliance Audit (2026-01-29)

### Missing from Implementation (SPEC-required)
- **Auth middleware enforcement**: JWT and API key checks exist but are not applied before handlers. The SPEC requires a security layer that enforces allow/deny on every request.
- **Dynamic OpenAPI endpoint**: OpenAPI generation exists, but there is no HTTP endpoint serving the live spec. The SPEC requires dynamic OpenAPI that reflects the in-memory registry and includes auth requirements and example payloads.
- **Collection schema updates**: `/collections:update` only supports adding columns. The SPEC requires add/remove/rename support.

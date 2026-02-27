## Overview

### Problem Statement
Moon currently fails to authenticate API keys when they are sent in the required `Authorization: Bearer <token>` format if the token does not parse as JWT. The current flow attempts JWT validation first and returns `401` immediately on JWT failure, which prevents API key authentication via Bearer token.

### Context and Background
- `SPEC_AUTH.md` requires both JWT and API keys to use `Authorization: Bearer`.
- Token type detection must distinguish JWT format (3 dot-separated segments) from API key format (`moon_live_` prefix).
- Current behavior still relies on legacy `X-API-Key` fallback in parts of the flow.
- This PRD explicitly requires **no backward compatibility** for legacy `X-API-Key` header.

### High-Level Solution Summary
Implement a single Bearer-token authentication path for protected endpoints:
1. Parse `Authorization: Bearer`.
2. Detect token type:
   - JWT: validate via JWT service and continue as user principal.
   - API key (`moon_live_`): validate via API key repository and continue as API key principal.
3. Remove `X-API-Key` header authentication logic completly.
4. Preserve existing RBAC and write-permission enforcement after successful principal resolution.

## Requirements

### Functional Requirements
1. Protected endpoints MUST authenticate using only `Authorization: Bearer <token>`.
2. Server MUST detect Bearer token type before validation:
   - JWT candidate: exactly 3 non-empty dot-separated segments.
   - API key candidate: begins with `moon_live_`.
3. For JWT candidates, server MUST run JWT validation and load user principal context.
4. For API key candidates, server MUST run API key validation and load API key principal context.
5. If Bearer token does not match JWT or API key format, server MUST return `401` with standard error format.
6. If JWT validation fails for a JWT candidate, server MUST return `401` with standard error format.
7. If API key validation fails for an API key candidate, server MUST return `401` with standard error format.
8. `X-API-Key` header authentication MUST be removed (hard fail/ignored for auth); no fallback behavior is allowed.
9. Authorization layer MUST continue to apply role-based checks exactly as currently defined for authenticated principal type.

### Technical Requirements
1. Consolidate auth decision in one middleware path to avoid duplicate auth logic.
2. Ensure context values (principal ID, role, can_write, auth type) are populated consistently for JWT and API key principals.
3. Keep HTTP method support unchanged (`GET`, `POST`, `OPTIONS`) and preserve existing routing behavior.
4. Ensure all auth error responses use the standard API error shape from `SPEC_API.md` (`{"message": "..."}`).
5. Add/adjust unit tests first (TDD) covering token type detection and Bearer-only behavior.
6. No warnings introduced in build/test pipeline.

### API Specifications
- **Affected inputs**
  - Header: `Authorization: Bearer <token>` (required for protected routes).
- **Removed/unsupported input**
  - Header: `X-API-Key` (unsupported for authentication).
- **Behavior matrix**
  - Bearer JWT valid -> authenticated.
  - Bearer JWT invalid -> `401`.
  - Bearer API key valid -> authenticated.
  - Bearer API key invalid -> `401`.
  - Bearer malformed/unknown token type -> `401`.
  - Missing Bearer on protected route -> `401`.

### Validation Rules and Constraints
1. Authorization header MUST start with `Bearer ` (case-insensitive scheme handling allowed if already standard in project).
2. Empty token after `Bearer` MUST be rejected as `401`.
3. API key detection MUST be prefix-based on `moon_live_` and MUST NOT attempt JWT validation first for such tokens.
4. JWT detection MUST require three token segments separated by `.`.

### Error Handling and Failure Modes
1. Authentication failures return `401` with `{"message": "..."}`.
2. Unknown token formats return `401` with a deterministic message (for testability).
3. Internal validation/repository failures return `500` with standard error envelope only when failure is server-side, not credential-invalid.
4. Token blacklist behavior for JWT remains enforced before granting access.

### Permissions and Limits
1. Existing roles (`admin`, `user`, `readonly`) remain in effect for authorization.
2. No change to endpoint-level permission model in this PRD.
3. No change to rate-limiting behavior in this PRD.

### Use Cases
1. **JWT user access**: Client sends valid JWT via Bearer and accesses protected endpoint.
2. **API key service access**: Integration sends valid `moon_live_...` via Bearer and accesses protected endpoint.
3. **Invalid JWT**: Client sends expired/invalid JWT via Bearer and receives `401`.
4. **Invalid API key**: Integration sends unknown/revoked `moon_live_...` via Bearer and receives `401`.
5. **Legacy header removed**: Client sends only `X-API-Key`; request is unauthenticated and receives `401`.

### Non-Goals
1. No changes to JWT issuance/refresh semantics.
2. No changes to API key creation/rotation endpoints.
3. No migration layer for legacy `X-API-Key` clients.
4. No auth protocol additions beyond Bearer token support.

### Needs Clarification
1. Exact `401` message text per failure category (invalid JWT vs invalid API key vs malformed token) should be finalized to avoid brittle tests if message wording is expected externally.

## Acceptance Criteria

1. **Bearer JWT success path**
   - Given a valid JWT in `Authorization: Bearer`, protected endpoint returns success response per endpoint contract.
2. **Bearer API key success path**
   - Given a valid `moon_live_...` token in `Authorization: Bearer`, protected endpoint returns success response per endpoint contract.
3. **JWT failure path**
   - Given an invalid JWT-form token in Bearer, endpoint returns `401` with `{"message": "..."}`.
4. **API key failure path**
   - Given an invalid `moon_live_...` token in Bearer, endpoint returns `401` with `{"message": "..."}`.
5. **Malformed Bearer token path**
   - Given non-empty Bearer token that is neither JWT-form nor `moon_live_...`, endpoint returns `401` with `{"message": "..."}`.
6. **Missing Authorization path**
   - Given no Bearer header on protected endpoint, endpoint returns `401` with `{"message": "..."}`.
7. **Legacy header removal**
   - Given only `X-API-Key` header, endpoint returns `401`; no fallback authentication occurs.
8. **Regression coverage**
   - Existing auth tests pass after update.
   - New tests cover token type detection branching and legacy-header rejection.
9. **Documentation alignment**
   - `SPEC_AUTH.md` and `SPEC_API.md` remain aligned with Bearer-only auth behavior.

### Test Scenarios (Minimum)
1. `Authorization: Bearer <valid-jwt>` -> authenticated principal type `jwt`.
2. `Authorization: Bearer moon_live_valid...` -> authenticated principal type `api_key`.
3. `Authorization: Bearer <invalid-jwt-format-3-segments>` -> `401`.
4. `Authorization: Bearer moon_live_invalid...` -> `401`.
5. `Authorization: Bearer abc123` -> `401`.
6. `X-API-Key: moon_live_valid...` without Bearer -> `401`.
7. Ensure role middleware receives equivalent context fields for both principal types.

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

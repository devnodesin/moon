# PRD: Remove Legacy X-API-Key Header Support

## 1. Overview

### Problem Statement
The codebase and documentation currently support the legacy `X-API-Key` header for API key authentication. This transitional support is now obsolete and must be fully removed in favor of the unified `Authorization: Bearer` header, as specified in `prd/057-unified-authentication-header.md`. Retaining legacy header support increases maintenance burden, creates ambiguity, and risks security inconsistencies.

### Context and Background
Moon's authentication system has transitioned to a unified header format for both JWT and API key authentication. The legacy `X-API-Key` header was retained for backward compatibility but is now deprecated. All references, code paths, and documentation supporting `X-API-Key` must be eliminated to enforce a single, predictable authentication mechanism.

### High-Level Solution Summary
Remove all code, configuration, and documentation related to the legacy `X-API-Key` header. Ensure only the `Authorization: Bearer` header is accepted for authentication. Update all relevant files, tests, and samples to reflect this change. Validate via grep search that no legacy header references remain.

## 2. Requirements

### Functional Requirements
- Remove all support for the `X-API-Key` header in source code, configuration, and documentation
- Accept only `Authorization: Bearer` for API key and JWT authentication
- Update all endpoint examples, samples, and tests to use the unified header
- Remove transitional/deprecation notices and sunset configuration for legacy header
- Ensure no references to `X-API-Key` remain in:
  - Source code (`cmd/`)
  - Documentation (`SPEC.md`, `SPEC_AUTH.md`, `AGENTS.md`)
  - Configuration samples (`moon.conf`)
  - Test files and samples

### Technical Requirements
- Remove all code paths, middleware, and logic handling `X-API-Key`
- Remove configuration options related to legacy header support (e.g., `legacy_header_support`, `legacy_header_sunset`)
- Update authentication error handling to reference only the unified header
- Update all documentation and samples to use `Authorization: Bearer` exclusively
- Run grep search for `X-API-Key` and related terms to confirm removal

### API Specifications
- No changes to endpoint URLs or request/response formats except header usage
- All authentication must use `Authorization: Bearer <token>`

### Validation Rules and Constraints
- No code, documentation, or configuration may reference or accept `X-API-Key`
- All authentication examples must use the unified header
- No transitional or deprecated header support remains

### Error Handling and Failure Modes
- If `X-API-Key` is provided, return `401 Unauthorized` with error message referencing only `Authorization: Bearer`
- No deprecation headers or migration links returned

### Filtering, Sorting, Permissions, and Limits
- No changes to filtering, sorting, or permissions logic; only header handling is affected

## 3. Acceptance Criteria

- All code, documentation, and configuration supporting `X-API-Key` are removed
- Only `Authorization: Bearer` is accepted and documented for authentication
- All endpoint examples, samples, and tests use the unified header
- No references to `X-API-Key` remain in the codebase (validated via grep search)
- No transitional or deprecation notices remain
- Authentication error messages reference only the unified header

### Test Scenarios
- Attempt authentication with `X-API-Key` header: expect `401 Unauthorized` referencing `Authorization: Bearer`
- Attempt authentication with `Authorization: Bearer` header: expect normal behavior
- Run grep search for `X-API-Key` and related terms: expect zero matches
- Review all documentation and samples for header usage

### Edge Cases and Negative Paths
- Requests with both headers: only `Authorization: Bearer` is processed; `X-API-Key` ignored and triggers error
- Configuration files with legacy header options: expect error or ignore
- Tests referencing legacy header: update or remove

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

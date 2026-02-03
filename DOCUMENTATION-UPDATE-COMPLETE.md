# API Documentation Update - Version 1.2.0

## Overview

Successfully updated the API documentation template (`cmd/moon/internal/handlers/templates/doc.md.tmpl`) from version 1.1.0 to 1.2.0 with comprehensive improvements following the requirements in `.github/prompts/UpdateApiDocumentation.prompt.md`.

## Changes Summary

### 1. Version Update
- **Documentation Version:** `1.1.0` → `1.2.0`
- Location: Properties table in the header section

### 2. New Sections Added

#### Rate Limiting (Line 862)
- Documents rate limits per authentication method:
  - JWT Authentication: 100 requests/minute
  - API Key Authentication: 1,000 requests/minute
- Rate limit response headers:
  - `X-RateLimit-Limit`
  - `X-RateLimit-Remaining`
  - `X-RateLimit-Reset`
- 429 error response example

#### CORS Configuration (Line 900)
- Supported HTTP methods: GET, POST, OPTIONS
- OPTIONS request documentation
- CORS headers and configuration
- Reference to `cors.allowed_origins` in `moon.conf`

#### Security Best Practices (Line 937)
- Authentication best practices (JWT vs API Key)
- Authorization principles (least privilege, admin endpoints)
- Input validation guidelines
- Network security recommendations (HTTPS, rate limiting)

#### JSON Appendix (Line 1013)
- Machine-readable schema for AI coding agents
- Complete endpoint catalog with authentication requirements
- Data types with SQL mappings
- Query operators and syntax
- HTTP status codes
- CORS and rate limiting configuration
- System guarantees and limitations
- AIP-136 standard reference

### 3. Enhanced Existing Sections

#### Response Format (Line 187)
- Added 429 (Too Many Requests) error response example

#### User Management (Lines 352-372)
- **reset_password action:** Admin can reset user passwords
  ```bash
  curl -X POST "{{$ApiURL}}/users:update?id=<USER_ID>" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"action": "reset_password", "password": "NewSecurePassword123#"}' | jq .
  ```
- **revoke_sessions action:** Admin can invalidate all user refresh tokens
  ```bash
  curl -X POST "{{$ApiURL}}/users:update?id=<USER_ID>" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"action": "revoke_sessions"}' | jq .
  ```

#### Intro Section (Lines 72-80)
- Added "What Moon Does NOT Do" subsection clarifying limitations:
  - No transactions
  - No joins
  - No triggers/hooks
  - No background jobs

#### Table of Contents (Lines 26-44)
- Added links to new sections:
  - Rate Limiting
  - CORS Configuration
  - Security Best Practices
  - JSON Appendix

## File Statistics

- **Total lines added:** 585
- **Total lines removed:** 1
- **Net change:** +584 lines
- **Sections added:** 4 major sections
- **Subsections enhanced:** 3

## Template Variables Preserved

All Go template syntax maintained:
- `{{.Version}}`
- `{{.BaseURL}}`
- `{{.Prefix}}`
- `{{$ApiURL}}`
- `{{.JWTEnabled}}`
- `{{.APIKeyEnabled}}`
- `{{.APIKeyHeader}}`

## Document Structure

Final structure follows the required format:

```
1. Header and Metadata
2. Table of Contents
3. Intro
   - Terminology
   - Design Constraints
   - What Moon Does NOT Do
4. Data Types
5. Health Check
6. Authentication Endpoints
7. User Management (Admin Only)
8. API Key Management (Admin Only)
9. Collection Management (Admin Only)
10. Data Access (Dynamic Collections)
11. Query Options
12. Aggregation Operations
13. Rate Limiting
14. CORS Configuration
15. Security Best Practices
16. Documentation
17. JSON Appendix
```

## Validation

- ✅ Build successful: `go build -o moon ./cmd/moon`
- ✅ Template syntax preserved
- ✅ All existing examples maintained
- ✅ All curl examples use `{{$ApiURL}}` variable
- ✅ All examples use `$ACCESS_TOKEN` environment variable
- ✅ Single-page format maintained
- ✅ Go template rendering compatible

## Deployment Notes

1. The template is embedded at compile time using `go:embed`
2. Server must be rebuilt for changes to take effect: `go build -o moon ./cmd/moon`
3. Documentation cache can be refreshed via: `POST /doc:refresh` (requires admin auth)
4. HTML docs available at: `/doc/`
5. Markdown docs available at: `/doc/md`

## JSON Appendix Highlights

The new JSON appendix provides machine-readable documentation including:
- Complete endpoint catalog with auth requirements
- Data type definitions with SQL mappings
- Query syntax and operators
- HTTP status code reference
- Rate limiting configuration
- CORS configuration
- System guarantees and limitations

This enables:
- Automated client generation
- AI agent integration
- API schema validation
- Developer tooling

## Future Considerations

1. Monitor user feedback on new sections
2. Consider adding more complex workflow examples
3. May need to update JSON appendix as API evolves
4. Security best practices should be reviewed periodically

## References

- Source: `cmd/moon/internal/handlers/templates/doc.md.tmpl`
- Spec: `SPEC.md`, `SPEC_AUTH.md`
- Prompt: `.github/prompts/UpdateApiDocumentation.prompt.md`
- Commit: 5014e53

---

**Status:** ✅ Complete and Deployed
**Date:** 2025
**Version:** 1.2.0

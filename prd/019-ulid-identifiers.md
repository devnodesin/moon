ulid  CHAR(26) NOT NULL UNIQUE,
## Overview
- Refactor backend identifier architecture to use ULID as the sole public identifier for all API resources.
- Internal SQL auto-increment IDs are strictly for internal logic and never exposed externally.
- This is a greenfield implementation—no migration or backward compatibility required.

## Requirements
- Every table uses an auto-increment SQL `id` (`BIGINT` or `INT`) as the internal primary key, used only for joins, foreign keys, and internal lookups.
- SQL `id` is never exposed in API responses, requests, logs, or error messages.
- Every table/collection has a `ulid` column:
  - Type: `CHAR(26)`
  - Generated at record creation
  - Globally unique
  - Indexed
  - `NOT NULL`, unique constraint
- ULID is generated using a Go library (e.g. `oklog/ulid`).
- All API identifiers are ULIDs. API field name is always `id`, but the value is a ULID. SQL `id` is never accepted or returned by the API.
- Records are ordered by `ulid ASC`. Pagination uses ULID as cursor:
  - `WHERE ulid > :after ORDER BY ulid ASC LIMIT :limit`
- API response format:
  - `{ "data": [...], "next_cursor": "<ulid or null>", "limit": 20 }`
- Reject any request containing numeric IDs or SQL `id` values. All record access is via ULID only. Prevent ID enumeration and scraping.
- Error handling:
  - Invalid ULID → `400 Bad Request`
  - ULID not found → `404 Not Found`
  - Never leak SQL errors or internal IDs

## Acceptance
- All API endpoints use ULID as the only identifier.
- SQL IDs are never exposed externally.
- Security and error handling requirements are met.
- All new tables/collections follow the schema and contract above.

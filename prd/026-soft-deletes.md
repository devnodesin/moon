## Overview

- Add soft delete support so records are not permanently removed immediately
- Provide an explicit, on-demand purge mechanism to permanently remove soft-deleted data when the user chooses


## Requirements

- Soft delete data model
  - A soft delete is represented by a nullable timestamp column named `deleted_at`
  - `deleted_at = NULL` means the record is active
  - `deleted_at != NULL` means the record is soft-deleted

- Collection schema support
  - Collections must support adding a `deleted_at` column via normal schema management (`/collections:create` and `/collections:update`)
  - The schema registry must cache and validate the `deleted_at` field when present
  - Soft delete behavior is enabled per-collection when (and only when) the collection schema contains a `deleted_at` column

- Default query semantics (read paths)
  - For collections that include `deleted_at`:
    - `GET /api/v1/{collection}:list` must exclude soft-deleted rows by default by applying `deleted_at IS NULL`
    - `GET /api/v1/{collection}:get` must treat soft-deleted rows as not found by default
  - For collections that do not include `deleted_at`, behavior remains unchanged

- Opt-in access to soft-deleted data
  - For collections that include `deleted_at`, support the following query parameters on `:list` and `:get`:
    - `include_deleted` (boolean, default: false)
      - When true, do not apply the default `deleted_at IS NULL` constraint
      - Allows clients/admins to read both active and soft-deleted rows
    - `only_deleted` (boolean, default: false; valid only when `include_deleted=true`)
      - When true, return only soft-deleted rows by applying `deleted_at IS NOT NULL`
  - Validation
    - If `only_deleted=true` and `include_deleted=false`, return `400 Bad Request`

- Delete semantics (write paths)
  - For collections that include `deleted_at`:
    - `POST /api/v1/{collection}:destroy` must perform a soft delete by setting `deleted_at` to the current server time for the targeted record
    - If the record is already soft-deleted, `:destroy` must be idempotent (no-op) and return success
  - For collections that do not include `deleted_at`:
    - `POST /api/v1/{collection}:destroy` continues to hard delete as it does today

- Restore endpoint
  - For collections that include `deleted_at`, add:
    - `POST /api/v1/{collection}:restore`
  - Behavior
    - Restores a soft-deleted record by setting `deleted_at` back to `NULL`
    - If the record is not soft-deleted, `:restore` is idempotent (no-op) and returns success
  - Validation
    - If the collection does not have `deleted_at`, return `400 Bad Request`

- On-demand purge (hard delete of soft-deleted rows)
  - Add an explicit purge endpoint for collections that include `deleted_at`:
    - `POST /api/v1/{collection}:purge_deleted`
  - Purpose
    - Permanently removes soft-deleted data when the user explicitly requests cleanup
  - Request parameters
    - `before` (optional, RFC3339 timestamp)
      - When provided, only purge rows where `deleted_at <= before`
      - When omitted, purge all rows where `deleted_at IS NOT NULL`
    - Filtering (optional)
      - Support existing filtering semantics (`?column[operator]=value`) to allow narrowing which deleted rows are purged
      - Purge must always include the base constraint `deleted_at IS NOT NULL` regardless of filters
  - Response shape
    - `200 OK` with body: `{ "purged": <integer> }` indicating number of rows permanently deleted
  - Safety requirements
    - Purge is a deliberate hard-delete operation and must only be triggered by this explicit endpoint
    - Purge must never run automatically in the background

- Error handling
  - All errors must use the standard error response format (code, message, details, request_id)
  - Soft delete related errors
    - Invalid boolean query values for `include_deleted` / `only_deleted` -> `400 Bad Request`
    - `only_deleted=true` with `include_deleted=false` -> `400 Bad Request`
    - `:restore` or `:purge_deleted` on a collection without `deleted_at` -> `400 Bad Request`
    - Invalid `before` timestamp format -> `400 Bad Request`

- OpenAPI
  - Dynamic OpenAPI generation must reflect soft delete behavior and the additional endpoints for collections that contain `deleted_at`
  - OpenAPI must document:
    - `include_deleted` / `only_deleted` query parameters for `:list` and `:get`
    - `POST :restore` and `POST :purge_deleted` endpoints and their request/response bodies

- Tests
  - Add automated tests covering (at minimum):
    - Default exclusion of deleted rows on `:list` and `:get`
    - `include_deleted=true` returns deleted rows
    - `only_deleted=true` returns only deleted rows and rejects invalid combinations
    - `:destroy` performs soft delete when `deleted_at` exists, hard delete when it does not
    - `:restore` restores records and is idempotent
    - `:purge_deleted` deletes only soft-deleted rows and respects optional `before` and filters

## Acceptance

- Soft delete behavior
  - When a collection includes a `deleted_at` column:
    - `:list` excludes soft-deleted rows by default
    - `:get` treats soft-deleted rows as not found by default
    - `include_deleted=true` returns both active and deleted rows
    - `only_deleted=true` returns only deleted rows and returns `400` if `include_deleted=false`
    - `:destroy` sets `deleted_at` (soft delete) and is idempotent
    - `:restore` sets `deleted_at` back to `NULL` and is idempotent
    - `:purge_deleted` permanently deletes soft-deleted rows only, returns `{ "purged": N }`, and never runs automatically
  - When a collection does not include `deleted_at`, existing delete/list/get behavior is unchanged

- Error handling
  - Invalid soft delete query parameters or timestamps return `400 Bad Request` with standard error format
  - `:restore` and `:purge_deleted` on non-soft-delete collections return `400 Bad Request` with standard error format

- OpenAPI
  - OpenAPI docs include soft delete query parameters and the additional endpoints (`:restore`, `:purge_deleted`) when `deleted_at` is present

- Tests
  - Automated tests exist for all major soft delete flows and purge edge cases (including idempotency and invalid parameter 
  combinations)
- Updated the scripts, documentations and specifications.

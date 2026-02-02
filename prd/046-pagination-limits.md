## Overview

- **Problem**: The current pagination implementation lacks proper enforcement of maximum page sizes and configurable limits, exposing the system to potential memory exhaustion and DoS attacks through unbounded data retrieval.
- **Context**: Moon currently implements cursor-based pagination for `/{collection}:list` endpoints and aggregation endpoints (`/{collection}:count`, `/:sum`, `/:avg`, `/:min`, `/:max`) with hardcoded limits (`DefaultPaginationLimit = 15`, `MaxPaginationLimit = 100`). There is no enforcement mechanism to prevent clients from bypassing these limits or requesting excessively large page sizes.
- **Solution**: Implement robust pagination limit enforcement across all list and aggregation endpoints with configurable limits, strict validation, proper cursor handling, and comprehensive error responses.

## Requirements

### Functional Requirements

**FR-1: Maximum Page Size Enforcement**
- All `/{collection}:list` endpoints MUST enforce a maximum page size of `MaxPageSize` (default: 200)
- All aggregation endpoints MUST enforce the same `MaxPageSize` limit when returning result sets
- If `?limit` query parameter exceeds `MaxPageSize`, return HTTP 400 with error message: `"page size exceeds maximum allowed: {MaxPageSize}"`
- If `?limit` is not provided, use `DefaultPageSize` (default: 15)
- If `?limit` is less than `MinPageSize` (hardcoded: 1), return HTTP 400 with error message: `"page size must be at least 1"`

**FR-2: Minimum Page Size Validation**
- `MinPageSize` MUST be hardcoded to `1`
- Any request with `limit < 1` MUST be rejected with HTTP 400
- Zero or negative limit values MUST return error: `"page size must be at least 1"`

**FR-3: Cursor Pagination Validation**
- All `?after` cursor values MUST be valid ULIDs (26 characters, Crockford Base32 encoding)
- Invalid ULID format MUST return HTTP 400 with error: `"invalid cursor: {validation_error}"`
- Empty cursor (`after=""`) MUST be treated as "start from beginning"
- Non-existent ULID cursors MUST NOT cause errors (return empty result set)

**FR-4: Configuration Integration**
- Add `pagination.default_page_size` to YAML configuration (default: 15)
- Add `pagination.max_page_size` to YAML configuration (default: 200)
- `MinPageSize` remains hardcoded at `1` (not configurable)
- Configuration MUST be loaded at startup and stored in `config.AppConfig`
- Invalid configuration values (negative, zero, or default > max) MUST prevent server startup with clear error message

**FR-5: Affected Endpoints**
All endpoints MUST enforce pagination limits:

**Data Endpoints:**
- `GET /{collection}:list`

**Collection Management:**
- `GET /collections:list`

**User Management (Auth):**
- `GET /users:list` (admin only)

**API Key Management (Auth):**
- `GET /apikeys:list` (admin only)

**Aggregation Endpoints (when returning result sets):**
- `GET /{collection}:count`
- `GET /{collection}:sum?field={field}`
- `GET /{collection}:avg?field={field}`
- `GET /{collection}:min?field={field}`
- `GET /{collection}:max?field={field}`

**FR-6: Response Format**
All paginated responses MUST include:
```json
{
  "data": [...],
  "next_cursor": "01HQXYZ...",  // null if no more data
  "limit": 15
}
```

### Technical Requirements

**TR-1: Constants Definition**
File: `cmd/moon/internal/constants/pagination.go`
```go
const (
    MinPageSize     = 1     // Hardcoded, not configurable
    DefaultPageSize = 15    // Default if not configured
    MaxPageSize     = 200   // Default if not configured
)
```

**TR-2: Configuration Structure**
File: `cmd/moon/internal/config/config.go`
```yaml
pagination:
  default_page_size: 15  # Default: 15 (range: 1-200)
  max_page_size: 200     # Default: 200 (must be >= default_page_size)
```

**TR-3: Validation Logic**
Create reusable pagination validation function:
```go
// File: cmd/moon/internal/pagination/validator.go
func ValidatePageSize(limit int, cfg *config.AppConfig) error {
    if limit < MinPageSize {
        return fmt.Errorf("page size must be at least %d", MinPageSize)
    }
    maxLimit := cfg.Pagination.MaxPageSize
    if limit > maxLimit {
        return fmt.Errorf("page size exceeds maximum allowed: %d", maxLimit)
    }
    return nil
}

func ValidateCursor(cursor string) error {
    if cursor == "" {
        return nil // Empty cursor is valid
    }
    return ulid.Validate(cursor)
}
```

**TR-4: Query Parameter Parsing**
All handlers MUST:
1. Parse `?limit` as integer (default to `cfg.Pagination.DefaultPageSize`)
2. Parse `?after` as string (default to empty)
3. Validate limit using `ValidatePageSize()`
4. Validate cursor using `ValidateCursor()`
5. Return HTTP 400 on validation failure

**TR-5: Database Query Enforcement**
- Query builder MUST NOT allow `limit` parameter to exceed configured `MaxPageSize`
- Always fetch `limit + 1` records to determine if more pages exist
- Trim result set to `limit` before returning response
- Set `next_cursor` to ULID of last record if more data exists

**TR-6: Error Response Format**
```json
{
  "error": "page size exceeds maximum allowed: 200"
}
```

**TR-7: Cursor Implementation**
- Use existing ULID validation: `cmd/moon/internal/ulid/ulid.go:Validate()`
- Cursor condition: `WHERE ulid > ?` (exclusive, not inclusive)
- Ordering MUST be `ORDER BY ulid ASC` for consistent pagination
- `next_cursor` is ULID of last record in current page

### Edge Cases and Error Handling

**EC-1: Invalid Query Parameters**
- `?limit=abc` → HTTP 400: `"invalid limit: strconv.ParseInt: parsing \"abc\": invalid syntax"`
- `?limit=0` → HTTP 400: `"page size must be at least 1"`
- `?limit=-10` → HTTP 400: `"page size must be at least 1"`
- `?limit=999` → HTTP 400: `"page size exceeds maximum allowed: 200"`

**EC-2: Invalid Cursor**
- `?after=invalid` → HTTP 400: `"invalid cursor: expected 26 characters, got 7"`
- `?after=ZZZZZZZZZZZZZZZZZZZZZZZZZZ` → HTTP 400: `"invalid cursor: ulid: bad data size when unmarshaling"`

**EC-3: Boundary Conditions**
- `?limit=1` → Valid, return 1 record
- `?limit=200` (if MaxPageSize=200) → Valid, return up to 200 records
- `?limit=201` (if MaxPageSize=200) → HTTP 400
- Empty result set with cursor → Return `{"data": [], "next_cursor": null, "limit": 15}`
- Last page with exactly `limit` records → Fetch `limit+1`, determine `next_cursor` is null

**EC-4: Configuration Edge Cases**
- `default_page_size` > `max_page_size` → Server startup failure
- `default_page_size` < 1 → Server startup failure
- `max_page_size` < 1 → Server startup failure
- Missing `pagination` config block → Use hardcoded defaults (15, 200)

**EC-5: Concurrent Modification**
- If new records are inserted with ULID < current cursor → Not visible in current pagination session (expected behavior for cursor pagination)
- If records are deleted → Cursor may skip records (acceptable trade-off for cursor pagination)

### Constraints and Limits

- **MinPageSize**: 1 (hardcoded, cannot be changed)
- **DefaultPageSize**: 15 (configurable, recommended range: 10-50)
- **MaxPageSize**: 200 (configurable, recommended range: 100-1000)
- **ULID Format**: Exactly 26 characters, Crockford Base32 encoding
- **Cursor Type**: String (ULID)
- **Response Payload Limit**: No explicit limit, but bounded by `MaxPageSize * record_size`

### Non-Functional Requirements

**NFR-1: Performance**
- Validation overhead MUST be < 1ms per request
- No additional database queries for validation
- Use indexed ULID column for cursor queries

**NFR-2: Security**
- Prevent DoS via unbounded pagination
- Validate all user input before database queries
- Log attempts to exceed `MaxPageSize` for monitoring

**NFR-3: Backward Compatibility**
- Existing clients using `?limit` within bounds continue to work
- Default behavior remains unchanged for clients not specifying `?limit`
- Error messages are clear and actionable

**NFR-4: Observability**
- Log validation errors at `WARN` level
- Include attempted `limit` value in error logs
- No PII in error messages or logs

## Acceptance

**AC-1: Configuration Loading**
- [ ] Create `cmd/moon/internal/config/config.go` with `PaginationConfig` struct containing `DefaultPageSize` and `MaxPageSize` fields
- [ ] Load `pagination.default_page_size` and `pagination.max_page_size` from YAML configuration
- [ ] Validate that `1 <= default_page_size <= max_page_size`
- [ ] Validate that `max_page_size >= 1`
- [ ] Server startup fails with error if validation fails
- [ ] Missing pagination config uses hardcoded defaults (15, 200)

**AC-2: Constants and Validator**
- [ ] Update `cmd/moon/internal/constants/pagination.go` with `MinPageSize = 1`, `DefaultPageSize = 15`, `MaxPageSize = 200`
- [ ] Create `cmd/moon/internal/pagination/validator.go` with `ValidatePageSize()` and `ValidateCursor()` functions
- [ ] Write unit tests for `ValidatePageSize()` covering: valid limits, below min, above max, edge cases
- [ ] Write unit tests for `ValidateCursor()` covering: valid ULID, invalid format, empty cursor, wrong length

**AC-3: Data Endpoint Enforcement**
- [ ] Update `cmd/moon/internal/handlers/data.go:HandleList()` to:
  - Parse `?limit` query parameter (default to config default)
  - Validate limit using `ValidatePageSize()`
  - Return HTTP 400 with appropriate error if validation fails
  - Pass validated limit to query builder
- [ ] Test: `GET /products:list?limit=1` returns 1 record
- [ ] Test: `GET /products:list?limit=200` returns up to 200 records (if MaxPageSize=200)
- [ ] Test: `GET /products:list?limit=201` returns HTTP 400 (if MaxPageSize=200)
- [ ] Test: `GET /products:list?limit=0` returns HTTP 400
- [ ] Test: `GET /products:list?limit=-5` returns HTTP 400
- [ ] Test: `GET /products:list` (no limit) returns default page size (15)

**AC-4: Cursor Validation**
- [ ] Update all `:list` handlers to validate `?after` cursor using `ValidateCursor()`
- [ ] Test: Valid ULID cursor `?after=01HQXYZ1234567890ABCDEFGH` returns records after cursor
- [ ] Test: Invalid cursor `?after=invalid` returns HTTP 400
- [ ] Test: Empty cursor `?after=` returns records from start
- [ ] Test: Non-existent ULID returns empty result set (not error)

**AC-5: Collection List Endpoint**
- [ ] Update `cmd/moon/internal/handlers/collection.go:HandleListCollections()` to enforce pagination limits
- [ ] Test: `GET /collections:list?limit=300` returns HTTP 400 (if MaxPageSize=200)
- [ ] Test: `GET /collections:list` returns default page size with `next_cursor`

**AC-6: User List Endpoint (Auth)**
- [ ] Update `cmd/moon/internal/auth/user_repository.go:ListPaginated()` to enforce pagination limits
- [ ] Update handler to validate limit and cursor
- [ ] Test: `GET /users:list?limit=500` returns HTTP 400 (admin auth required)
- [ ] Test: `GET /users:list?after=invalidulid` returns HTTP 400

**AC-7: API Key List Endpoint (Auth)**
- [ ] Update `cmd/moon/internal/auth/apikey_repository.go:ListPaginated()` to enforce pagination limits
- [ ] Update handler to validate limit and cursor
- [ ] Test: `GET /apikeys:list?limit=999` returns HTTP 400 (admin auth required)

**AC-8: Aggregation Endpoints**
- [ ] Update `cmd/moon/internal/handlers/aggregation.go` to validate `?limit` for endpoints returning result sets
- [ ] Test: Each aggregation endpoint with `?limit=201` returns HTTP 400 (if applicable)
- [ ] Verify that count/sum/avg/min/max respect pagination limits when returning multiple records

**AC-9: Error Response Format**
- [ ] All validation errors return JSON: `{"error": "descriptive message"}`
- [ ] HTTP status code is 400 for all validation errors
- [ ] Error messages are clear and include the failed constraint
- [ ] Test each error scenario and verify response format

**AC-10: Next Cursor Logic**
- [ ] Query always fetches `limit + 1` records
- [ ] If result count > limit, set `next_cursor` to ULID of last record (before trim)
- [ ] If result count <= limit, set `next_cursor` to null
- [ ] Trim result set to `limit` before returning
- [ ] Test: Page with exactly `limit` records returns `next_cursor: null`
- [ ] Test: Page with `limit + 1` records returns valid `next_cursor`

**AC-11: Configuration Edge Cases**
- [ ] Test: `default_page_size: 0` causes server startup failure
- [ ] Test: `max_page_size: 0` causes server startup failure
- [ ] Test: `default_page_size: 300` and `max_page_size: 200` causes startup failure
- [ ] Test: Missing `pagination` config uses defaults (15, 200)

**AC-12: Integration Tests**
- [ ] Test: Paginate through 500 records using cursor pagination with `limit=20`
- [ ] Test: Request `limit=MaxPageSize` on collection with 1000+ records
- [ ] Test: Concurrent requests with different cursors return consistent results
- [ ] Test: Pagination works correctly with filters, sorting, and field selection

**AC-13: Performance Validation**
- [ ] Measure validation overhead: < 1ms per request
- [ ] Verify no N+1 queries introduced
- [ ] Test with MaxPageSize records to ensure memory usage is bounded

**AC-14: Documentation Updates**
- [ ] Update `SPEC.md` with pagination limits section
- [ ] Update API documentation template in `cmd/moon/internal/handlers/templates/doc.md.tmpl`
- [ ] Document configuration options in `INSTALL.md`
- [ ] Add pagination examples to API documentation

**AC-15: Testing Checklist**
- [ ] All unit tests pass for validation functions
- [ ] All integration tests pass for each endpoint
- [ ] All error scenarios produce correct HTTP status and error message
- [ ] Configuration loading tested with valid/invalid values
- [ ] Cursor pagination tested end-to-end with real data
- [ ] Edge cases (min, max, boundary values) tested
- [ ] Test coverage >= 90% for new code

---

### Implementation Checklist

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `USAGE.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

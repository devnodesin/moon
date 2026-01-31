## Overview

- Cursor-based pagination in the `/{collection}:list` endpoint is skipping records when using the `after` query parameter
- The root cause is that `next_cursor` is incorrectly set to the ID of the extra fetched record (at index `limit`) instead of the last returned record (at index `limit-1`)
- When clients use this cursor in subsequent requests, they skip one record because the cursor points to a record that was never returned in the previous page
- This prevents users from navigating through all records in a collection when using pagination
- Fix requires changing the cursor assignment logic to use the last returned record's ID instead of the extra fetched record's ID

## Requirements

### Functional Requirements

- **Cursor Assignment Logic:**
  - When `len(data) > limit` (indicating more records exist), set `next_cursor` to the ULID of the record at index `limit-1` (the last record being returned in the current response)
  - Do NOT set `next_cursor` to the ULID of the record at index `limit` (the extra record used only to determine if more data exists)
  - The extra record at index `limit` should be discarded and never exposed to clients

- **Pagination Behavior:**
  - First request: `?limit=N` returns records 0 to N-1, with `next_cursor` set to the ID of record N-1
  - Second request: `?limit=N&after=<cursor>` returns records N to 2N-1, with `next_cursor` set to the ID of record 2N-1
  - Continue until all records are returned
  - Last page: `next_cursor` is `null` when no more records exist

- **Existing Cursor Filter Logic:**
  - Keep the existing `ulid > cursor_value` logic (line 128-132 in data.go)
  - This correctly excludes the cursor record itself from subsequent queries

### Technical Requirements

- **File to Modify:** `cmd/moon/internal/handlers/data.go`
- **Function:** `List` method in `DataHandler`
- **Lines to Change:** Approximately lines 210-217

- **Current Incorrect Logic:**
  ```go
  if len(data) > limit {
      // More data available, use the ULID of the last item as cursor
      lastItem := data[len(data)-1]  // ← This is the extra record at index `limit`
      if ulidVal, ok := lastItem["id"].(string); ok {
          nextCursor = &ulidVal
      }
      // Remove the extra item we fetched
      data = data[:limit]
  }
  ```

- **Required Fixed Logic:**
  ```go
  if len(data) > limit {
      // More data available, use the ULID of the last returned record as cursor
      // Truncate to limit first
      data = data[:limit]
      // Now get the last item from the returned data
      lastItem := data[len(data)-1]  // ← This is the last returned record at index `limit-1`
      if ulidVal, ok := lastItem["id"].(string); ok {
          nextCursor = &ulidVal
      }
  }
  ```

### Validation Requirements

- **Unit Tests:**
  - Create test case in `cmd/moon/internal/handlers/data_test.go` (or create file if missing)
  - Test scenario: Collection with 5 records, paginate with limit=1
  - Verify all 5 records are accessible through sequential pagination
  - Verify no records are skipped
  - Verify cursor values point to last returned record on each page

- **Integration Tests:**
  - Test with SQLite, Postgres, and MySQL dialects
  - Create collection with 10+ records
  - Paginate with limit=1, limit=2, limit=5
  - Verify all records are returned exactly once
  - Verify `next_cursor` is correct on each page
  - Verify `next_cursor` is null on the last page

- **Edge Cases:**
  - Single record in collection (limit=1 should return it with null cursor)
  - Exactly `limit` records in collection (should return all with null cursor)
  - Empty collection (should return empty array with null cursor)
  - Pagination with filters and sorting applied

### Error Handling

- No new error conditions introduced
- Existing error handling for invalid cursors remains unchanged
- If `lastItem["id"]` is missing or not a string, cursor should remain nil (existing behavior)

## Acceptance

- [x] Code changes applied to `cmd/moon/internal/handlers/data.go`
- [x] Unit tests added to verify cursor logic with multiple pagination scenarios
- [x] Integration tests verify no records are skipped across SQLite, Postgres, MySQL
- [x] Manual testing confirms:
  - Collection with 3 records, limit=1: Three requests return record 1, record 2, record 3 sequentially
  - No records are skipped when using `next_cursor` from each response
  - Last page correctly returns `next_cursor: null`
- [x] All existing tests pass without modification
- [x] No new compilation warnings introduced
- [x] Code formatted with `gofmt`
- [x] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `USAGE.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [x] Ensure all test scripts in `scripts/*.sh` are working properly and up to date with the latest code and API changes.

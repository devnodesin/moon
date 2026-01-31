# PRD 041: Data Type Refactor - SQL Native Mapping

## Overview

This PRD defines a comprehensive refactor of Moon's data type system to use SQL-native type mappings instead of the current generic type abstraction. The current system uses API-level types (`string`, `integer`, `float`, `boolean`, `datetime`, `json`, `text`) that are mapped to database-specific SQL types. The new system will eliminate the intermediate abstraction layer and expose SQL-native types directly in the API, while maintaining database-agnostic compatibility through a clear mapping table.

**Status:** Breaking change - No backward compatibility maintained  
**Impact:** Complete refactor of type system across API, registry, validation, query builder, handlers, documentation, and tests

## Current State

### Current API Data Types
- `string` - Short, length-limited text (e.g., names, titles, emails)
- `text` - Long, unbounded text (e.g., descriptions, comments)
- `integer` - Whole numbers
- `float` - Decimal numbers
- `boolean` - true/false values
- `datetime` - Date/time in RFC3339 or ISO 8601 format (e.g., 2023-01-31T13:45:00Z)
- `json` - Arbitrary JSON object or array

### Current SQL Mappings

**PostgreSQL:**
- `string` → `VARCHAR(255)`
- `integer` → `INTEGER`
- `float` → `DOUBLE PRECISION`
- `boolean` → `BOOLEAN`
- `datetime` → `TIMESTAMP`
- `text` → `TEXT`
- `json` → `JSONB`

**MySQL:**
- `string` → `VARCHAR(255)`
- `integer` → `INT`
- `float` → `DOUBLE`
- `boolean` → `BOOLEAN`
- `datetime` → `DATETIME`
- `text` → `TEXT`
- `json` → `JSON`

**SQLite:**
- `string` → `TEXT`
- `integer` → `INTEGER`
- `float` → `REAL`
- `boolean` → `INTEGER` (0/1)
- `datetime` → `TEXT` (ISO-8601)
- `text` → `TEXT`
- `json` → `TEXT`

## Requirements

### 1. New Data Type System

**API Data Types** (used in collection creation/modification APIs):
- `string` - Maps to TEXT in all databases
- `integer` - Maps to INTEGER (SQLite) or BIGINT (PostgreSQL/MySQL)
- `boolean` - Maps to INTEGER 0/1 (SQLite) or BOOLEAN (PostgreSQL/MySQL)
- `datetime` - Maps to TEXT ISO-8601 UTC (SQLite) or TIMESTAMP (PostgreSQL/MySQL)
- `json` - Maps to TEXT (SQLite) or JSON (PostgreSQL/MySQL)

**Remove the following types:**
- `text` - No longer used (use `string` instead)
- `float` - No longer used (use `integer` instead, or document workaround)

### 2. SQL Type Mapping Table

| API Data Type | SQLite                 | MySQL / PostgreSQL / MariaDB |
| ------------- | ---------------------- | ---------------------------- |
| **string**    | `TEXT`                 | `TEXT`                       |
| **integer**   | `INTEGER`              | `BIGINT`                     |
| **boolean**   | `INTEGER` (0/1)        | `BOOLEAN`                    |
| **datetime**  | `TEXT` (ISO-8601, UTC) | `TIMESTAMP`                  |
| **json**      | `TEXT`                 | `JSON`                       |

**Rationale:**
- **Simplification:** Eliminates the distinction between `string` and `text` by mapping both to TEXT
- **Precision:** Uses BIGINT for integers to support larger numbers and avoid overflow issues
- **Compatibility:** Maintains SQLite's affinity system while providing proper types for other databases
- **Storage:** All string data now stored as TEXT, removing VARCHAR length constraints
- **Consistency:** Clear, predictable mapping across all databases

### 3. Code Changes Required

#### 3.1 Registry Package (`cmd/moon/internal/registry/registry.go`)
**Current:**
```go
const (
	TypeString   ColumnType = "string"
	TypeInteger  ColumnType = "integer"
	TypeFloat    ColumnType = "float"
	TypeBoolean  ColumnType = "boolean"
	TypeDatetime ColumnType = "datetime"
	TypeText     ColumnType = "text"
	TypeJSON     ColumnType = "json"
)
```

**New:**
```go
const (
	TypeString   ColumnType = "string"
	TypeInteger  ColumnType = "integer"
	TypeBoolean  ColumnType = "boolean"
	TypeDatetime ColumnType = "datetime"
	TypeJSON     ColumnType = "json"
)
```

**Changes:**
- Remove `TypeText` and `TypeFloat` constants
- Update `MapGoTypeToColumnType` function to remove float mappings
- Update validation logic to reject float and text types

#### 3.2 Validation Package (`cmd/moon/internal/validation/validator.go`)
**Changes:**
- Remove all `TypeText` references in validation logic
- Remove all `TypeFloat` validation logic
- Update `validateType` function to only support new type set
- Remove float parsing and validation rules
- Update error messages to reflect new type system
- Update `validateStringConstraints` to remove VARCHAR length restrictions (since all strings now TEXT)

#### 3.3 Query Builder Package (`cmd/moon/internal/query/builder.go`)
**Update SQL Type Mappings:**

**PostgreSQL:**
```go
func mapColumnTypeToPostgres(colType registry.ColumnType) string {
	switch colType {
	case registry.TypeString:
		return "TEXT"
	case registry.TypeInteger:
		return "BIGINT"
	case registry.TypeBoolean:
		return "BOOLEAN"
	case registry.TypeDatetime:
		return "TIMESTAMP"
	case registry.TypeJSON:
		return "JSON"
	default:
		return "TEXT"
	}
}
```

**MySQL:**
```go
func mapColumnTypeToMySQL(colType registry.ColumnType) string {
	switch colType {
	case registry.TypeString:
		return "TEXT"
	case registry.TypeInteger:
		return "BIGINT"
	case registry.TypeBoolean:
		return "BOOLEAN"
	case registry.TypeDatetime:
		return "TIMESTAMP"
	case registry.TypeJSON:
		return "JSON"
	default:
		return "TEXT"
	}
}
```

**SQLite:**
```go
func mapColumnTypeToSQLite(colType registry.ColumnType) string {
	switch colType {
	case registry.TypeString:
		return "TEXT"
	case registry.TypeInteger:
		return "INTEGER"
	case registry.TypeBoolean:
		return "INTEGER"
	case registry.TypeDatetime:
		return "TEXT"
	case registry.TypeJSON:
		return "TEXT"
	default:
		return "TEXT"
	}
}
```

**Changes:**
- Remove `TypeText` and `TypeFloat` from all mapping functions
- Update `TypeString` mapping from `VARCHAR(255)` to `TEXT` for PostgreSQL/MySQL
- Update `TypeInteger` mapping from `INTEGER`/`INT` to `BIGINT` for PostgreSQL/MySQL
- Ensure SQLite continues to use `INTEGER` for integers (affinity-based)
- Remove DOUBLE PRECISION, REAL, DOUBLE mappings

#### 3.4 Collections Handler (`cmd/moon/internal/handlers/collections.go`)
**Changes:**
- Update `mapColumnTypeToSQL` and dialect-specific functions
- Update `generateCreateTableDDL` to use new mappings
- Update `generateAddColumnDDL` to use new mappings
- Update `generateModifyColumnDDL` to use new mappings
- Remove text and float type handling
- Update validation error messages

#### 3.5 Data Handler (`cmd/moon/internal/handlers/data.go`)
**Changes:**
- Remove float conversion logic in `convertValue` function
- Update `validateFieldType` to remove float validation
- Remove text type handling (treat same as string)
- Update error messages

#### 3.6 Aggregation Handler (`cmd/moon/internal/handlers/aggregation.go`)
**Changes:**
- Update `validateNumericField` to only accept `registry.TypeInteger`
- Remove float field validation
- Update error messages for numeric-only fields

#### 3.7 Database Inspector (`cmd/moon/internal/database/inspector.go`)
**Changes:**
- Update `InferColumnType` function to map database types back to API types:
  - FLOAT, DOUBLE, REAL, DECIMAL, NUMERIC → Reject or map to `TypeInteger` with warning
  - TEXT, CLOB → `TypeString`
  - VARCHAR, CHAR → `TypeString`
  - Remove float inference logic

#### 3.8 Documentation Template (`cmd/moon/internal/handlers/doc.go`)
**Changes:**
- Update embedded documentation template with new type table
- Add migration guide for users upgrading from old system
- Update all examples to use new types
- Remove references to `text` and `float` types

### 4. Test Updates

**Files to Update:**
- `cmd/moon/internal/registry/registry_test.go`
- `cmd/moon/internal/validation/validator_test.go`
- `cmd/moon/internal/query/builder_test.go`
- `cmd/moon/internal/handlers/collections_test.go`
- `cmd/moon/internal/handlers/data_test.go`
- `cmd/moon/internal/handlers/aggregation_test.go`
- `cmd/moon/internal/database/inspector_test.go`

**Changes Required:**
- Replace all `TypeText` with `TypeString`
- Replace all `TypeFloat` with `TypeInteger`
- Update expected SQL output in query builder tests
- Update validation test cases
- Remove float validation tests
- Update aggregation tests to only accept integers
- Update type inference tests

### 5. Documentation Updates

#### 5.1 SPEC.md
**Sections to Update:**
- Section 1.0: System Philosophy - Update data type references
- Section 2.0: API Endpoint Specification - Update field type examples
- Section 3.0: Architecture - Update type mapping tables
- Add new section on SQL Native Type Mapping with complete table

**Add New Section:**
```markdown
## SQL Native Type Mapping

Moon uses SQL-native type mappings for maximum database compatibility and performance:

| API Data Type | SQLite                 | MySQL / PostgreSQL / MariaDB |
| ------------- | ---------------------- | ---------------------------- |
| **string**    | `TEXT`                 | `TEXT`                       |
| **integer**   | `INTEGER`              | `BIGINT`                     |
| **boolean**   | `INTEGER` (0/1)        | `BOOLEAN`                    |
| **datetime**  | `TEXT` (ISO-8601, UTC) | `TIMESTAMP`                  |
| **json**      | `TEXT`                 | `JSON`                       |

### Type Details

**string:** 
- All databases: TEXT (unlimited length)
- No VARCHAR length constraints
- Suitable for all text data

**integer:**
- SQLite: INTEGER (affinity-based, can store up to 64-bit)
- PostgreSQL/MySQL: BIGINT (64-bit signed integer)
- Range: -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807

**boolean:**
- SQLite: INTEGER storing 0 (false) or 1 (true)
- PostgreSQL/MySQL: Native BOOLEAN type
- API accepts: true/false JSON boolean values

**datetime:**
- SQLite: TEXT storing ISO-8601 format in UTC (e.g., "2023-01-31T13:45:00Z")
- PostgreSQL/MySQL: TIMESTAMP type
- API accepts: RFC3339 or ISO 8601 string format
- Always stored in UTC

**json:**
- SQLite: TEXT storing JSON string
- PostgreSQL: JSON type
- MySQL: JSON type
- API accepts: Any valid JSON (object, array, primitive)

### Removed Types

The following types from previous versions are no longer supported:
- **text:** Use `string` instead (all strings now TEXT)
- **float:** Use `integer` for whole numbers; for decimal precision, store as string or use external computation
```

#### 5.2 USAGE.md
**Changes:**
- Update all collection creation examples to use new types
- Replace `text` with `string` in all examples
- Replace `float` with `integer` in all examples
- Add note about removed types
- Update type validation examples

#### 5.3 README.md
**Changes:**
- Update feature list if it mentions data types
- Ensure type references are accurate

#### 5.4 INSTALL.md
**Changes:**
- Add migration notes for users upgrading from previous versions
- Document type mapping changes

### 6. Scripts and Samples Updates

#### 6.1 Test Scripts (`scripts/`)
**Files to Update:**
- `scripts/collection.sh`
- Any other scripts creating collections with type definitions

**Changes:**
- Replace `"type": "text"` with `"type": "string"`
- Replace `"type": "float"` with `"type": "integer"`
- Update test data to use integers instead of floats

#### 6.2 Sample Configuration (`samples/`)
**Changes:**
- Review sample configs for any type references
- Update documentation comments if they mention types

### 7. Build and Install Scripts

#### 7.1 build.sh
**Changes:**
- Add version bump or changelog note about breaking change
- No functional changes required

#### 7.2 install.sh
**Changes:**
- No changes required (installation process unchanged)

## Acceptance Criteria

### A1. Core Type System
- [ ] `TypeText` constant removed from `registry.ColumnType`
- [ ] `TypeFloat` constant removed from `registry.ColumnType`
- [ ] Only 5 types remain: `string`, `integer`, `boolean`, `datetime`, `json`
- [ ] All type validation updated to reject text and float types
- [ ] `MapGoTypeToColumnType` function updated to remove float mappings

### A2. SQL Mapping Updates
- [ ] PostgreSQL: string → TEXT, integer → BIGINT
- [ ] MySQL: string → TEXT, integer → BIGINT
- [ ] SQLite: string → TEXT, integer → INTEGER
- [ ] All datetime mappings use TIMESTAMP (PostgreSQL/MySQL) or TEXT (SQLite)
- [ ] All boolean mappings use BOOLEAN (PostgreSQL/MySQL) or INTEGER (SQLite)
- [ ] All JSON mappings use JSON (PostgreSQL/MySQL) or TEXT (SQLite)

### A3. Code Compilation
- [ ] All Go files compile without errors
- [ ] All Go files compile without warnings
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` shows no issues
- [ ] `gofmt -w .` produces no changes

### A4. Test Coverage
- [ ] All existing tests updated to use new type system
- [ ] All tests pass with new type mappings
- [ ] Registry tests updated (no text/float types)
- [ ] Validation tests updated (no text/float validation)
- [ ] Query builder tests updated (new SQL output)
- [ ] Handler tests updated (collections, data, aggregation)
- [ ] Database inspector tests updated (type inference)
- [ ] Integration tests pass with real databases (SQLite, PostgreSQL, MySQL if available)

### A5. Documentation Updates
- [ ] SPEC.md updated with new type mapping table
- [ ] SPEC.md includes removed types section
- [ ] USAGE.md updated with new type examples
- [ ] README.md reviewed and updated if needed
- [ ] INSTALL.md includes migration notes
- [ ] Inline code documentation updated (godoc comments)
- [ ] API documentation template updated (`doc.go`)

### A6. Scripts and Samples
- [ ] `scripts/collection.sh` updated to use new types
- [ ] All test scripts use new type system
- [ ] Sample configs reviewed and updated if needed
- [ ] All sample requests use new types

### A7. Validation and Error Handling
- [ ] API rejects collection creation with `text` type (clear error message)
- [ ] API rejects collection creation with `float` type (clear error message)
- [ ] API accepts all 5 valid types: string, integer, boolean, datetime, json
- [ ] Error messages reference new type system
- [ ] Validation error codes updated

### A8. Aggregation Operations
- [ ] Sum, Avg, Min, Max operations only accept `integer` fields
- [ ] Operations reject `float` fields (no longer exists)
- [ ] Error messages for non-numeric fields updated

### A9. Database Inspector
- [ ] `InferColumnType` no longer returns `TypeFloat` or `TypeText`
- [ ] FLOAT/DOUBLE/REAL database columns map to `TypeInteger` or are handled gracefully
- [ ] TEXT/VARCHAR database columns map to `TypeString`
- [ ] Type inference tests updated

### A10. Backward Compatibility
- [ ] **No backward compatibility maintained** (breaking change)
- [ ] Existing databases with `text` or `float` types handled gracefully on startup:
  - Option 1: Orphaned tables dropped (if `drop_orphans: true`)
  - Option 2: Orphaned tables registered with inferred types (if `drop_orphans: false`)
  - Type inference maps old types to new types
- [ ] Clear error messages when old types are used

### A11. Performance and Behavior
- [ ] Query performance unchanged or improved (TEXT vs VARCHAR)
- [ ] No memory leaks or performance regressions
- [ ] String storage works correctly without length limits
- [ ] Integer storage uses BIGINT for large numbers
- [ ] Boolean storage and retrieval works correctly (0/1 in SQLite, native in others)
- [ ] Datetime storage and parsing works correctly
- [ ] JSON storage and parsing works correctly

### A12. Integration Testing
- [ ] Create collection with all 5 types succeeds
- [ ] Insert data with all 5 types succeeds
- [ ] Query data with all 5 types succeeds
- [ ] Update data with all 5 types succeeds
- [ ] Delete data succeeds
- [ ] Aggregation operations on integer fields succeed
- [ ] Full-text search on string fields succeeds
- [ ] Filtering/sorting on all types works correctly
- [ ] Cross-database compatibility verified (SQLite, PostgreSQL, MySQL)

## Migration Notes for Users

### Breaking Changes
1. **Type Removal:** `text` and `float` types no longer supported
2. **Type Mappings Changed:** `string` now maps to TEXT instead of VARCHAR(255)
3. **Integer Size:** Integers now use BIGINT (64-bit) in PostgreSQL/MySQL

### Migration Path
1. **Backup Data:** Export all data before upgrading
2. **Update Schemas:** Replace `text` with `string`, `float` with `integer`
3. **Update Application Code:** Ensure clients send integers instead of floats
4. **Test:** Verify all APIs work with new types
5. **Deploy:** Upgrade Moon server

### Workarounds for Removed Types
- **For `text` type:** Use `string` type (now TEXT, unlimited length)
- **For `float` type:**
  - Option 1: Store as `integer` (multiply by 10^n, divide on client)
  - Option 2: Store as `string` (format as needed)
  - Option 3: Use external computation service

## Implementation Plan

### Phase 1: Core Type System (2-3 hours)
1. Update `registry/registry.go` - Remove TypeText and TypeFloat
2. Update `validation/validator.go` - Remove validation logic
3. Update registry tests
4. Update validation tests
5. Verify compilation

### Phase 2: SQL Mapping (2-3 hours)
1. Update `query/builder.go` - Update all mapping functions
2. Update `handlers/collections.go` - Update DDL generation
3. Update `handlers/data.go` - Remove float handling
4. Update `handlers/aggregation.go` - Integer-only validation
5. Update `database/inspector.go` - Type inference
6. Update all handler tests
7. Update query builder tests

### Phase 3: Documentation (1-2 hours)
1. Update SPEC.md with new type table and removed types section
2. Update USAGE.md with new examples
3. Update README.md if needed
4. Update INSTALL.md with migration notes
5. Update inline godoc comments
6. Update `handlers/doc.go` template

### Phase 4: Scripts and Integration (1 hour)
1. Update `scripts/collection.sh`
2. Update any other test scripts
3. Review and update sample configs

### Phase 5: Testing and Validation (2-3 hours)
1. Run full test suite (`go test ./...`)
2. Fix any failing tests
3. Run integration tests with real databases
4. Test collection creation with all types
5. Test data insertion/retrieval
6. Test aggregation operations
7. Test error handling for invalid types
8. Verify documentation accuracy

### Phase 6: Final Review (1 hour)
1. Code review and cleanup
2. Format all files with `gofmt`
3. Run `go vet` and address any issues
4. Update AGENTS.md if needed
5. Commit changes with clear message

**Total Estimated Time:** 9-13 hours

## Risks and Mitigation

### Risk 1: Breaking Existing Deployments
**Impact:** High  
**Mitigation:** 
- Clear migration documentation
- Version bump with breaking change notice
- Changelog entry highlighting breaking changes
- Consider providing migration script or tool

### Risk 2: Data Loss During Migration
**Impact:** High  
**Mitigation:**
- Document backup procedures
- Test migration path thoroughly
- Provide rollback instructions

### Risk 3: Test Failures
**Impact:** Medium  
**Mitigation:**
- Systematic test updates
- Run tests frequently during development
- Address failures immediately

### Risk 4: Performance Regression
**Impact:** Low  
**Mitigation:**
- TEXT vs VARCHAR performance should be similar or better for most use cases
- BIGINT uses more storage but provides better range
- Monitor performance during testing

### Risk 5: Documentation Inconsistencies
**Impact:** Medium  
**Mitigation:**
- Comprehensive documentation review
- Update all references to old types
- Provide clear examples with new types

## Success Metrics

1. **Code Quality:**
   - Zero compilation errors
   - Zero compilation warnings
   - All tests passing
   - 90%+ test coverage maintained

2. **Documentation Quality:**
   - All type references updated
   - Clear migration guide provided
   - Examples use new types consistently

3. **Functionality:**
   - All API operations work with new types
   - Cross-database compatibility maintained
   - No performance degradation

4. **User Experience:**
   - Clear error messages for invalid types
   - Smooth upgrade path documented
   - No data loss during migration

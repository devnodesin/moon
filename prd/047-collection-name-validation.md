## Overview

- **Problem**: Current collection name validation does not enforce system prefix protection, length limits, or consistent case-sensitivity rules, leading to potential conflicts with system tables and cross-database portability issues.
- **Context**: Moon currently validates collection names using regex pattern `^[a-zA-Z][a-zA-Z0-9_]*$` and blocks SQL reserved keywords. However, it lacks enforcement for: (1) `moon_` and `moon` prefix protection to prevent conflicts with system tables (`moon_users`, `moon_apikeys`, etc.), (2) minimum/maximum length constraints to ensure database portability (PostgreSQL 63-char limit, MySQL 64-char limit), and (3) case-sensitivity normalization to avoid platform-dependent behavior across SQLite, PostgreSQL, and MySQL.
- **Solution**: Implement comprehensive collection name validation with system prefix protection, length constraints (2-63 characters), lowercase normalization, and enhanced character restrictions to ensure consistency, security, and cross-database compatibility.

## Requirements

### Functional Requirements

**FR-1: System Prefix Protection**
- Collection names MUST NOT start with `moon_` (case-insensitive)
- Collection names MUST NOT be exactly `moon` (case-insensitive)
- Attempting to create a collection with these prefixes MUST return HTTP 400 with error: `"collection name cannot start with 'moon_' or be 'moon' (reserved for system tables)"`
- This rule applies to all collection operations: `collections:create`, `collections:rename`

**Rationale:**
- System tables use `moon_` prefix: `moon_users`, `moon_refresh_tokens`, `moon_apikeys`, `moon_blacklisted_tokens`
- Prevents user collections from conflicting with current or future system tables
- Clear namespace separation between user-managed and system-managed tables

**FR-2: Length Constraints**
- Collection names MUST have minimum length of **2 characters** (after validation and normalization)
- Collection names MUST have maximum length of **63 characters**
- Names shorter than 2 characters MUST return HTTP 400: `"collection name must be at least 2 characters"`
- Names longer than 63 characters MUST return HTTP 400: `"collection name must not exceed 63 characters"`

**Rationale:**
- **PostgreSQL**: Max identifier length is 63 characters (NAMEDATALEN - 1)
- **MySQL**: Max table name length is 64 characters
- **SQLite**: No hard limit, but 63 ensures portability
- Minimum of 2 prevents single-letter collection names that are unclear
- Ensures collection names fit in URLs without encoding issues

**FR-3: Case-Sensitivity Normalization**
- Collection names MUST be automatically converted to **lowercase** before storage
- User input `"Products"`, `"PRODUCTS"`, or `"products"` all normalize to `"products"`
- Case-insensitive duplicate detection: Creating `"Users"` when `"users"` exists MUST return HTTP 409: `"collection 'users' already exists"`
- Return normalized name in response: `{"name": "products", ...}`

**Rationale:**
- **PostgreSQL**: Folds identifiers to lowercase unless quoted
- **MySQL**: Case-sensitivity depends on OS (Linux: case-sensitive, Windows: case-insensitive)
- **SQLite**: Case-sensitive but lacks normalization
- Enforcing lowercase ensures consistent behavior across all databases and prevents duplicate collections with different casing

**FR-4: Character Restrictions**
- Collection names MUST start with an **alphabetic character** (a-z after normalization)
- Collection names MUST contain only:
  - Lowercase letters: `a-z`
  - Numbers: `0-9`
  - Underscores: `_`
- No special characters allowed except underscore
- No spaces, hyphens, or other punctuation allowed
- Regex pattern (after lowercase conversion): `^[a-z][a-z0-9_]*$`
- Invalid characters MUST return HTTP 400: `"collection name must start with a letter and contain only lowercase letters, numbers, and underscores"`

**FR-5: Reserved Keyword Protection (Existing)**
- Continue blocking SQL reserved keywords (SELECT, INSERT, UPDATE, DELETE, FROM, WHERE, JOIN, TABLE, etc.)
- Error message: `"'{name}' is a reserved keyword and cannot be used as a collection name"`

**FR-6: Empty Name Validation (Existing)**
- Empty or whitespace-only names MUST return HTTP 400: `"collection name cannot be empty"`
- Trim whitespace before validation

**FR-7: Validation Order**
All validation steps MUST execute in this order:
1. Empty/whitespace check
2. Trim and convert to lowercase
3. Length validation (min 2, max 63)
4. System prefix check (`moon_`, `moon`)
5. Character pattern validation (`^[a-z][a-z0-9_]*$`)
6. Reserved keyword check
7. Duplicate collection check

**FR-8: Affected Endpoints**
Apply validation to:
- `POST /collections:create` (existing)
- `POST /collections/{name}:rename` (if implemented)
- Any future endpoint that creates or renames collections

### Technical Requirements

**TR-1: Validation Function Signature**
File: `cmd/moon/internal/handlers/collections.go`

```go
// validateCollectionName validates and normalizes collection names.
// Returns the normalized (lowercase) name and validation error if any.
func validateCollectionName(name string) (normalizedName string, err error) {
    // Implementation details in acceptance criteria
}
```

**TR-2: Validation Constants**
File: `cmd/moon/internal/constants/validation.go`

```go
const (
    MinCollectionNameLength = 2
    MaxCollectionNameLength = 63
    
    // Pattern for normalized names (lowercase only)
    CollectionNamePattern = `^[a-z][a-z0-9_]*$`
    
    // Reserved prefixes
    SystemPrefix = "moon_"
    SystemNamespace = "moon"
)
```

**TR-3: Case-Insensitive Duplicate Check**
File: `cmd/moon/internal/handlers/collections.go`

```go
// Before creating collection, check if normalized name exists
normalizedName, err := validateCollectionName(req.Name)
if err != nil {
    return err
}

if h.registry.Exists(normalizedName) {
    return fmt.Errorf("collection '%s' already exists", normalizedName)
}
```

**TR-4: System Table Detection Enhancement**
File: `cmd/moon/internal/constants/tables.go`

```go
// IsSystemTableOrPrefix checks if name is system table or uses reserved prefix.
func IsSystemTableOrPrefix(name string) bool {
    lower := strings.ToLower(name)
    return strings.HasPrefix(lower, SystemPrefix) || lower == SystemNamespace
}
```

**TR-5: Error Response Format**
All validation errors MUST return:
```json
{
  "error": "descriptive validation error message"
}
```

HTTP status codes:
- `400 Bad Request`: Validation failure (length, format, prefix, reserved keyword)
- `409 Conflict`: Duplicate collection name

**TR-6: Schema Registry Integration**
- Store collections in schema registry using **normalized lowercase names**
- All lookups MUST use normalized names
- `registry.Exists(name)` MUST use case-insensitive comparison (normalize input before lookup)

**TR-7: Response Format**
Collection creation response MUST include normalized name:
```json
{
  "name": "products",
  "columns": [...]
}
```

**TR-8: Database Table Creation**
- Execute `CREATE TABLE` with normalized lowercase name
- No quoting of identifiers (let database handle folding)
- Example: Input `"Products"` → SQL `CREATE TABLE products (...)`

### Validation Rules and Constraints

**Rule 1: Empty Name**
- Input: `""`, `"   "`, `null`
- Error: `"collection name cannot be empty"`

**Rule 2: Length Constraints**
- Min: 2 characters (after normalization)
- Max: 63 characters
- Input: `"a"` → Error: `"collection name must be at least 2 characters"`
- Input: `"a_very_long_collection_name_that_exceeds_the_maximum_allowed_limit"` (64 chars) → Error: `"collection name must not exceed 63 characters"`

**Rule 3: System Prefix Protection**
- Input: `"moon_"` → Error: `"collection name cannot start with 'moon_' or be 'moon' (reserved for system tables)"`
- Input: `"Moon_Users"` → Error (case-insensitive check)
- Input: `"moon"` → Error
- Input: `"MOON"` → Error
- Input: `"moonbase"` → Valid (not exact match)

**Rule 4: Character Pattern**
- Must start with letter (a-z after normalization)
- Must contain only: a-z, 0-9, _
- Input: `"123products"` → Error: `"collection name must start with a letter and contain only lowercase letters, numbers, and underscores"`
- Input: `"products-v2"` → Error (hyphen not allowed)
- Input: `"products v2"` → Error (space not allowed)
- Input: `"products@home"` → Error (special character not allowed)

**Rule 5: Reserved Keywords**
- Input: `"select"`, `"TABLE"`, `"Join"` → Error: `"'{name}' is a reserved keyword and cannot be used as a collection name"`

**Rule 6: Case Normalization**
- Input: `"Products"` → Normalized: `"products"`
- Input: `"USER_ACCOUNTS"` → Normalized: `"user_accounts"`
- Input: `"MixedCase123"` → Normalized: `"mixedcase123"`

### Error Handling and Failure Modes

**EH-1: Validation Failure**
- All validation errors return HTTP 400 with descriptive error message
- Do not execute database queries if validation fails
- Log validation failures at `WARN` level with attempted name

**EH-2: Duplicate Collection**
- If normalized name exists in schema registry, return HTTP 409
- Error message includes normalized name: `"collection 'products' already exists"`
- Prevent case-insensitive duplicates (e.g., `"Products"` and `"products"`)

**EH-3: Database Constraint Violation**
- If database rejects table creation (e.g., duplicate table from manual SQL), return HTTP 500
- Log database error details
- Clean up schema registry entry if database creation fails

**EH-4: Invalid Rename**
- If renaming to existing collection name, return HTTP 409
- Apply same validation rules as collection creation
- Cannot rename to system table name

### Edge Cases and Constraints

**EC-1: Boundary Conditions**
- Input: `"ab"` → Valid (exactly 2 characters)
- Input: `"a"` → Invalid (below minimum)
- Input: 63-character name → Valid (at maximum)
- Input: 64-character name → Invalid (exceeds maximum)

**EC-2: System Prefix Edge Cases**
- Input: `"moon_"` → Invalid (exact prefix)
- Input: `"moon_custom"` → Invalid (starts with prefix)
- Input: `"moon"` → Invalid (exact match)
- Input: `"moonbase"` → Valid (prefix match but not exact)
- Input: `"my_moon_table"` → Valid (moon_ not at start)

**EC-3: Case Variations**
- Input: `"users"` then `"Users"` → Second request returns HTTP 409
- Input: `"PRODUCTS"` then `"products"` → Second request returns HTTP 409
- Input: `"MixedCase"` then `"mixedcase"` → Second request returns HTTP 409

**EC-4: Special Characters**
- Input: `"products_v2"` → Valid (underscore allowed)
- Input: `"products__v2"` → Valid (multiple underscores allowed)
- Input: `"_products"` → Invalid (must start with letter)
- Input: `"products_"` → Valid (underscore at end)

**EC-5: Unicode and Non-ASCII**
- Input: `"prödücts"` → Invalid (non-ASCII characters)
- Input: `"用户"` → Invalid (non-ASCII characters)
- Error message: `"collection name must start with a letter and contain only lowercase letters, numbers, and underscores"`

**EC-6: Reserved Keyword Cases**
- Input: `"SELECT"`, `"select"`, `"Select"` → All invalid (case-insensitive check)

### Non-Functional Requirements

**NFR-1: Performance**
- Validation overhead MUST be < 1ms per request
- No database queries for validation (use in-memory schema registry)
- Case-insensitive duplicate check via normalized name lookup

**NFR-2: Security**
- Prevent SQL injection via strict character validation
- Block system table namespace to prevent privilege escalation attempts
- Log suspicious attempts (e.g., SQL keywords, system prefixes)

**NFR-3: Backward Compatibility**
- Existing collections with uppercase or mixed-case names remain accessible
- New validation applies only to new collection creation
- No automatic migration of existing collection names
- **Migration Path**: Document manual steps to rename existing collections if needed

**NFR-4: Database Portability**
- 63-character limit ensures compatibility with PostgreSQL (strictest limit)
- Lowercase normalization ensures consistent behavior across SQLite, PostgreSQL, MySQL
- Character restrictions prevent database-specific identifier issues

**NFR-5: User Experience**
- Clear, actionable error messages
- Include attempted name in error message for clarity
- Return normalized name in success response

**NFR-6: Observability**
- Log validation errors at `WARN` level
- Include attempted name and error reason in logs
- No PII in logs (collection names are not PII)

## Acceptance

**AC-1: Update Validation Constants**
- [ ] Update `cmd/moon/internal/constants/validation.go` with:
  - `MinCollectionNameLength = 2`
  - `MaxCollectionNameLength = 63`
  - `CollectionNamePattern = ^[a-z][a-z0-9_]*$` (lowercase only)
  - `SystemPrefix = "moon_"`
  - `SystemNamespace = "moon"`
- [ ] Write unit tests for constants

**AC-2: Implement Enhanced Validation Function**
- [ ] Update `validateCollectionName()` in `cmd/moon/internal/handlers/collections.go` to:
  - Accept input name and return `(normalizedName string, err error)`
  - Trim whitespace and check for empty string
  - Convert to lowercase using `strings.ToLower()`
  - Validate length (2-63 characters)
  - Check for system prefix/namespace using `IsSystemTableOrPrefix()`
  - Validate character pattern using updated regex
  - Check reserved keywords (case-insensitive)
- [ ] Write comprehensive unit tests covering all validation rules

**AC-3: System Prefix Protection Tests**
- [ ] Test: `"moon_"` returns error
- [ ] Test: `"moon_users"` returns error
- [ ] Test: `"Moon_Custom"` returns error (case-insensitive)
- [ ] Test: `"MOON"` returns error
- [ ] Test: `"moon"` returns error
- [ ] Test: `"moonbase"` is valid (not exact match)
- [ ] Test: `"my_moon_table"` is valid (moon_ not at start)

**AC-4: Length Validation Tests**
- [ ] Test: `"a"` returns error (below minimum)
- [ ] Test: `"ab"` is valid (exactly 2 characters)
- [ ] Test: `"products"` is valid
- [ ] Test: 63-character name is valid
- [ ] Test: 64-character name returns error (exceeds maximum)
- [ ] Test: Empty string returns error
- [ ] Test: Whitespace-only string returns error

**AC-5: Case Normalization Tests**
- [ ] Test: `"Products"` normalizes to `"products"`
- [ ] Test: `"USER_ACCOUNTS"` normalizes to `"user_accounts"`
- [ ] Test: `"MixedCase123"` normalizes to `"mixedcase123"`
- [ ] Test: Response includes normalized name: `{"name": "products"}`
- [ ] Test: Database table created with lowercase name

**AC-6: Character Pattern Tests**
- [ ] Test: `"products_v2"` is valid
- [ ] Test: `"_products"` returns error (must start with letter)
- [ ] Test: `"123products"` returns error (must start with letter)
- [ ] Test: `"products-v2"` returns error (hyphen not allowed)
- [ ] Test: `"products v2"` returns error (space not allowed)
- [ ] Test: `"products@home"` returns error (special character not allowed)
- [ ] Test: `"prödücts"` returns error (non-ASCII)

**AC-7: Duplicate Detection (Case-Insensitive)**
- [ ] Create collection `"products"` successfully
- [ ] Attempt to create `"Products"` returns HTTP 409: `"collection 'products' already exists"`
- [ ] Attempt to create `"PRODUCTS"` returns HTTP 409
- [ ] Attempt to create `"PrOdUcTs"` returns HTTP 409
- [ ] Schema registry lookup uses normalized name

**AC-8: Reserved Keyword Tests**
- [ ] Test: `"select"` returns error
- [ ] Test: `"SELECT"` returns error (case-insensitive)
- [ ] Test: `"table"` returns error
- [ ] Test: `"users"` is valid (not reserved by default)

**AC-9: Update Collection Creation Endpoint**
- [ ] Update `POST /collections:create` handler to:
  - Call `validateCollectionName()` and capture normalized name
  - Return HTTP 400 on validation failure with error message
  - Use normalized name for schema registry and database operations
  - Return normalized name in response
- [ ] Test: `POST /collections:create` with `"Products"` creates `"products"`
- [ ] Test: Response includes `{"name": "products"}`

**AC-10: Error Response Format**
- [ ] All validation errors return JSON: `{"error": "message"}`
- [ ] HTTP 400 for validation failures
- [ ] HTTP 409 for duplicate collections
- [ ] Error messages are clear and include attempted name where appropriate

**AC-11: Integration with Schema Registry**
- [ ] Schema registry stores collections using normalized names
- [ ] `registry.Exists(name)` normalizes input before lookup
- [ ] `registry.Get(name)` normalizes input before lookup
- [ ] `registry.List()` returns collections with normalized names

**AC-12: Database Table Creation**
- [ ] `CREATE TABLE` statement uses normalized lowercase name
- [ ] No identifier quoting (let database handle folding)
- [ ] Verify table created successfully in SQLite, PostgreSQL, MySQL
- [ ] Query `SELECT * FROM products` works (case-insensitive)

**AC-13: Edge Case Coverage**
- [ ] Test all boundary conditions (2 chars, 63 chars, 1 char, 64 chars)
- [ ] Test all system prefix variations (moon_, Moon_, MOON_, moon, MOON)
- [ ] Test all character pattern edge cases (_, --, spaces, special chars)
- [ ] Test Unicode and non-ASCII characters
- [ ] Test reserved keywords (case variations)

**AC-14: Logging and Observability**
- [ ] Validation errors logged at `WARN` level
- [ ] Log includes attempted name and error reason
- [ ] No sensitive data in logs
- [ ] Duplicate collection attempts logged

**AC-15: Documentation Updates**
- [ ] Update `SPEC.md` with collection name validation rules
- [ ] Update API documentation template in `cmd/moon/internal/handlers/templates/doc.md.tmpl`
- [ ] Add validation rules to `INSTALL.md` or `README.md`
- [ ] Document migration path for existing collections (if needed)

**AC-16: Backward Compatibility**
- [ ] Existing collections remain accessible (no breaking changes)
- [ ] New validation applies only to new collections
- [ ] Document migration steps for renaming existing collections (if applicable)
- [ ] No automatic migration of existing collection names

**AC-17: Testing Checklist**
- [ ] All unit tests pass for validation functions
- [ ] All integration tests pass for collection creation
- [ ] All error scenarios produce correct HTTP status and message
- [ ] Case-insensitive duplicate detection tested end-to-end
- [ ] Cross-database compatibility tested (SQLite, PostgreSQL, MySQL)
- [ ] Test coverage >= 90% for new validation code

**AC-18: Rename Endpoint (Future)**
- [ ] If `POST /collections/{name}:rename` exists, apply same validation
- [ ] Validate new name using `validateCollectionName()`
- [ ] Prevent renaming to existing collection (case-insensitive)
- [ ] Cannot rename to system table name

---

### Implementation Checklist

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `USAGE.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

## Overview

- **Problem**: The system lacks comprehensive validation and constraint enforcement across collection names, column names, data types, schema limits, query parameters, and transaction handling, leading to potential security vulnerabilities, database inconsistencies, and unbounded resource usage.
- **Context**: Moon currently implements basic validation for collection and column names using regex patterns, but lacks enforcement for: (1) reserved endpoint names (collections, auth, users, apikeys, doc), (2) deprecated type removal (text, float), (3) schema size limits (max collections: 1000, max columns per collection: 100), (4) query parameter limits (filters, sort fields), (5) default value validation, (6) uniqueness constraint documentation, (7) query timeout enforcement, and (8) slow query logging.
- **Solution**: Implement comprehensive validation and constraint enforcement system covering naming rules, data type validation, schema limits, query constraints, transaction atomicity, and performance monitoring to ensure security, consistency, and resource efficiency.

## Requirements

### Functional Requirements

#### FR-1: Collection Name Validation - Reserved Endpoints

**Rule**: Collection names MUST NOT conflict with reserved endpoint names.

**Reserved Names** (case-insensitive):
- `collections`
- `auth`
- `users`
- `apikeys`
- `doc`
- `health` (system endpoint)

**Enforcement**:
- Validation MUST occur at collection creation time (`POST /collections:create`)
- Validation MUST occur at collection rename time (if `POST /collections/{name}:rename` exists)
- Attempting to create a collection with a reserved name MUST return HTTP 400
- Error message: `"collection name '{name}' is reserved for system endpoints"`
- Check is case-insensitive: `"Collections"`, `"USERS"`, `"Auth"` all rejected

**Rationale**:
- Prevents URL routing conflicts: `/{collectionName}:list` vs `/collections:list`
- Protects system endpoints from being shadowed by user collections
- Ensures consistent API behavior across all deployments

#### FR-2: Collection Name Validation - Pattern and Length

**Pattern**: `^[a-zA-Z][a-zA-Z0-9_]*$`
- MUST start with a letter (a-z, A-Z)
- MUST contain only letters, numbers, and underscores
- Underscore is allowed (e.g., `user_accounts`, `order_items`)
- No spaces, hyphens, or special characters allowed

**Length Constraints**:
- Minimum: 2 characters
- Maximum: 63 characters

**URL Safety**:
- Pattern already prevents problematic characters
- No additional URL encoding validation required
- Collection names appear in URLs as `/{collectionName}:list`

**Error Messages**:
- Empty/whitespace: `"collection name cannot be empty"`
- Invalid pattern: `"collection name must start with a letter and contain only letters, numbers, and underscores"`
- Too short: `"collection name must be at least 2 characters"`
- Too long: `"collection name must not exceed 63 characters"`

#### FR-3: Column Name Validation - Pattern and Length

**Pattern**: `^[a-z][a-z0-9_]*$` (lowercase only)
- MUST start with a lowercase letter (a-z)
- MUST contain only lowercase letters, numbers, and underscores
- Case enforcement: Input `"UserName"` MUST be rejected (not auto-converted)

**Length Constraints**:
- Minimum: 3 characters
- Maximum: 63 characters

**System Column Protection**:
- Cannot add columns named `id` or `ulid` (reserved, auto-created)
- Cannot remove system columns `id` or `ulid`
- Cannot rename system columns `id` or `ulid`
- Error message: `"cannot add/remove/rename system column '{name}'"`

**Error Messages**:
- Empty: `"column name cannot be empty"`
- Invalid pattern: `"column name must start with a lowercase letter and contain only lowercase letters, numbers, and underscores"`
- Too short: `"column name must be at least 3 characters"`
- Too long: `"column name must not exceed 63 characters"`
- System column conflict: `"cannot add system column '{name}'"`

#### FR-4: Data Type Validation

**Supported Types** (exactly 6):
1. `string` - Text values (maps to TEXT)
2. `integer` - 64-bit integer (maps to BIGINT)
3. `decimal` - Exact numeric precision (maps to NUMERIC/DECIMAL)
4. `boolean` - True/false (maps to BOOLEAN/INTEGER)
5. `datetime` - RFC3339 timestamp (maps to TIMESTAMP/TEXT)
6. `json` - JSON objects/arrays (maps to JSON/TEXT)

**Deprecated Types** (REMOVED - no backward compatibility):
- `text` - REMOVED (use `string` instead)
- `float` - REMOVED (use `decimal` or `integer` instead)

**Validation**:
- All column types MUST be validated using `registry.ValidateColumnType()`
- Invalid types MUST return HTTP 400: `"invalid column type '{type}'. Supported types: string, integer, decimal, boolean, datetime, json"`
- Deprecated types MUST return HTTP 400: `"type '{type}' is deprecated and no longer supported. Use 'string' instead of 'text', or 'decimal'/'integer' instead of 'float'"`

**Decimal Type Constraints**:
- Default scale: 2 decimal places (e.g., `19.99`)
- Maximum scale: 10 decimal places
- Scale limit MUST be enforced via named constant: `constants.DecimalMaxScale = 10`
- Invalid decimal values MUST be rejected before insert/update

#### FR-5: Schema Size Limits

**Collection Limit**:
- Maximum collections per server: **1000**
- Count MUST be checked before creating a new collection
- Error message: `"maximum number of collections (1000) reached"`
- HTTP status: 409 Conflict

**Column Limit**:
- Maximum columns per collection: **100**
- Count MUST be checked before adding a new column
- System columns (`id`, `ulid`) count toward limit
- Error message: `"maximum number of columns (100) reached for collection '{name}'"`
- HTTP status: 409 Conflict

**Rationale**:
- Prevents unbounded schema growth and excessive memory usage
- Schema registry uses `sync.Map` with all collections in memory
- Limits chosen for broad database compatibility and API usability
- PostgreSQL default max columns: 1600 (100 is conservative)
- Memory estimate: 1000 collections × 100 columns × ~100 bytes = ~10MB

#### FR-6: Uniqueness Constraints

**Single Column Unique Constraints**:
- Support `unique: true` on individual columns
- Unique constraint enforced at database level (CREATE UNIQUE INDEX)
- Constraint naming convention: `idx_{collection}_{column}_unique`

**NULL Value Behavior**:
- SQL standard: Multiple NULL values are allowed in unique columns (NULLs are distinct)
- PostgreSQL, MySQL, SQLite: All follow SQL standard (multiple NULLs allowed)
- Document explicitly in SPEC.md

**Error Handling**:
- Duplicate value insert/update MUST return HTTP 409 Conflict
- Error message: `"duplicate value for unique column '{column}'"`
- Include original database error for debugging (in logs, not response)

**Compound Unique Constraints**:
- NOT currently supported
- If requested, return HTTP 400: `"compound unique constraints are not supported"`
- Future enhancement: `unique_together: [["field1", "field2"]]`

#### FR-7: Default Value Validation

**Type Matching**:
- Default value MUST match column type
- Type conversion/coercion is NOT allowed
- Validation MUST occur at column creation/modification time

**Format Requirements by Type**:
- `string`: Any string value (e.g., `"default text"`)
- `integer`: Numeric string parseable as int64 (e.g., `"42"`, `"-100"`)
- `decimal`: Valid decimal string (e.g., `"19.99"`, `"0.00"`)
- `boolean`: `"true"` or `"false"` (case-insensitive)
- `datetime`: RFC3339 format (e.g., `"2024-01-01T00:00:00Z"`)
- `json`: Valid JSON string (e.g., `"{}"`, `"[]"`, `"null"`)

**NULL as Default**:
- If column is `nullable: true`, default can be `null` (omit `default_value` field)
- If column is `nullable: false`, default MUST be provided or insert will require value
- Explicit `default_value: null` MUST be rejected for non-nullable columns

**Validation Errors**:
- Type mismatch: `"default value '{value}' is invalid for type '{type}'"`
- Invalid format: `"default value '{value}' does not match required format for type '{type}'"`
- NULL for non-nullable: `"default value cannot be null for non-nullable column"`

**Examples**:
```json
{
  "name": "status",
  "type": "string",
  "nullable": false,
  "default_value": "active"
}

{
  "name": "price",
  "type": "decimal",
  "nullable": false,
  "default_value": "0.00"
}

{
  "name": "created_at",
  "type": "datetime",
  "nullable": false,
  "default_value": "2024-01-01T00:00:00Z"
}
```

#### FR-8: Transaction Support and Atomicity

**Schema Operations**:
- All schema modifications MUST be atomic
- Use database transactions for all CREATE/ALTER/DROP operations
- On transaction failure, rollback database AND registry changes

**Transaction Workflow**:
1. Begin database transaction
2. Execute DDL statement (CREATE TABLE, ALTER TABLE, DROP TABLE)
3. On success: Commit transaction → Update schema registry
4. On failure: Rollback transaction → Do NOT update registry

**Registry Rollback**:
- If database operation succeeds but registry update fails, rollback database transaction
- Registry uses `sync.Map` for concurrent reads (no locks on read)
- Schema modifications are serialized at database level (table locks)

**Error Recovery**:
- Database error: Return HTTP 500, log full error, registry unchanged
- Registry error: Rollback database, return HTTP 500, log error
- Partial failure: Always rollback to maintain consistency

**Concurrency**:
- Multiple read operations can run concurrently (no locks)
- Schema modifications acquire database-level table locks
- No distributed locking (single-instance deployment only)
- Consistency check on startup detects and repairs registry/database mismatches

#### FR-9: Query String Validation

**Filter Limits**:
- Maximum filters per request: **20**
- Each query parameter `field[operator]=value` counts as one filter
- Exceeding limit MUST return HTTP 400: `"maximum number of filters (20) exceeded"`

**Sort Field Limits**:
- Maximum sort fields per request: **5**
- `?sort=field1,-field2,+field3` counts as 3 sort fields
- Exceeding limit MUST return HTTP 400: `"maximum number of sort fields (5) exceeded"`

**Valid Filter Operators**:
- `eq` - equals
- `ne` - not equals
- `gt` - greater than
- `gte` - greater than or equal
- `lt` - less than
- `lte` - less than or equal
- `contains` - string contains (case-sensitive)
- `icontains` - string contains (case-insensitive)
- `startswith` - string starts with
- `endswith` - string ends with
- `in` - value in list (comma-separated)
- `null` - is null (value ignored)
- `notnull` - is not null (value ignored)

**Invalid operators** MUST return HTTP 400: `"invalid filter operator '{operator}'. Valid operators: eq, ne, gt, gte, lt, lte, contains, icontains, startswith, endswith, in, null, notnull"`

**Filter Value Type Checking**:
- Value type MUST be validated against column type BEFORE query execution
- Integer column with non-numeric value: HTTP 400: `"invalid value '{value}' for integer column '{column}'"`
- Boolean column with non-boolean value: HTTP 400: `"invalid value '{value}' for boolean column '{column}'"`
- Datetime column with invalid RFC3339: HTTP 400: `"invalid datetime value '{value}' for column '{column}'"`

**SQL Injection Prevention**:
- ALL queries MUST use parameterized statements (prepared statements)
- NO string concatenation of user input into SQL
- Filter values MUST be passed as parameters, not interpolated
- Column names MUST be validated against schema registry before query construction

#### FR-10: Query Timeout and Performance Limits

**Query Timeout**:
- Default timeout: **30 seconds**
- Configurable via `database.query_timeout` in YAML config
- Timeout MUST be enforced using `context.WithTimeout()`
- Timeout error MUST return HTTP 504 Gateway Timeout: `"query execution timeout (30s)"`

**Slow Query Threshold**:
- Default threshold: **500ms**
- Configurable via `database.slow_query_threshold` in YAML config
- Queries exceeding threshold MUST be logged at `WARN` level
- Log format: `"Slow query detected: {duration}ms - {sql} - params: {params}"`

**Records Per Collection**:
- No hard limit enforced (recommended: unlimited)
- If limit required in future, enforce via configuration
- Storage limits should be managed at database level, not application level

**Configuration Structure**:
```yaml
database:
  connection: "sqlite"
  database: "/opt/moon/sqlite.db"
  query_timeout: 30          # Default: 30 seconds
  slow_query_threshold: 500  # Default: 500 milliseconds
```

**Performance Monitoring**:
- Log all query execution times
- Aggregate slow query statistics (future enhancement)
- Expose query metrics via health endpoint (future enhancement)

### System Limits

The following limits are enforced to ensure system stability:

| Resource                       | Limit         | Rationale                                        |
| ------------------------------ | ------------- | ------------------------------------------------ |
| Maximum Collections            | 1000          | Prevents unbounded schema growth                 |
| Maximum Columns per Collection | 100           | Ensures API usability and DB compatibility       |
| Default Page Size              | 50 records    | Balances response size and performance           |
| Maximum Page Size              | 1000 records  | Prevents memory exhaustion                       |
| Maximum Collection Name Length | 64 characters | PostgreSQL identifier limit (63) + safety margin |
| Maximum Column Name Length     | 64 characters | Consistency with collection names                |
| Query Timeout                  | 30 seconds    | Prevents long-running queries                    |
| Slow Query Threshold           | 500ms         | Logged for performance monitoring                |

**Note:** These limits may be configurable in future versions.

### Naming Conventions

#### Collection Names

Collection names MUST follow these rules:

1. **Character Set:** Start with a letter (a-z, A-Z), followed by letters, numbers, or underscores
2. **Pattern:** `^[a-zA-Z][a-zA-Z0-9_]*$`
3. **Length:** 2-63 characters (64 in system limits table for safety margin)
4. **Case:** Lowercase enforced for portability (auto-converted to lowercase, see PRD-047)
5. **Reserved Prefixes:** Cannot start with `moon_` or be `moon` (reserved for system tables)
6. **Reserved Names:** Cannot be: `collections`, `auth`, `users`, `apikeys`, `doc`, `health` (reserved endpoints)
7. **SQL Keywords:** Cannot use SQL reserved words (select, insert, update, delete, etc. - see full list below)

**Valid Examples:**

- `users`
- `blog_posts`
- `product_inventory`
- `order_items_2024`

**Invalid Examples:**

- `123_products` (starts with number)
- `moon_orders` (reserved prefix)
- `user-profiles` (contains hyphen)
- `select` (SQL reserved word)
- `auth` (reserved endpoint)
- `a` (too short, minimum 2 characters)

#### Column Names

Column names MUST follow these rules:

1. **Character Set:** Start with a lowercase letter (a-z), followed by lowercase letters, numbers, or underscores
2. **Pattern:** `^[a-z][a-z0-9_]*$`
3. **Length:** 3-63 characters
4. **System Columns:** Cannot use `id` or `ulid` (automatically created)
5. **SQL Keywords:** Cannot use SQL reserved words
6. **Case:** Lowercase only (uppercase rejected, not auto-converted)

**Valid Examples:**

- `username`
- `email_address`
- `created_at`
- `product_id`

**Invalid Examples:**

- `UserName` (uppercase not allowed)
- `id` (system column)
- `ulid` (system column)
- `ab` (too short, minimum 3 characters)
- `select` (SQL reserved word)

**Protected System Columns:**

- `id`: Auto-increment primary key (internal use only)
- `ulid`: ULID identifier (exposed as `id` in API responses)

### Technical Requirements

#### TR-1: Validation Constants

File: `cmd/moon/internal/constants/validation.go`

```go
const (
    // Collection constraints
    MinCollectionNameLength = 2
    MaxCollectionNameLength = 63
    MaxCollectionsPerServer = 1000
    
    // Column constraints
    MinColumnNameLength = 3
    MaxColumnNameLength = 63
    MaxColumnsPerCollection = 100
    
    // Data type constraints
    DecimalDefaultScale = 2
    DecimalMaxScale = 10
    
    // Query constraints
    MaxFiltersPerRequest = 20
    MaxSortFieldsPerRequest = 5
    
    // Performance constraints
    DefaultQueryTimeout = 30          // seconds
    DefaultSlowQueryThreshold = 500   // milliseconds
    
    // Regex patterns
    CollectionNamePattern = `^[a-zA-Z][a-zA-Z0-9_]*$`
    ColumnNamePattern = `^[a-z][a-z0-9_]*$`
)

// ReservedEndpointNames are collection names that conflict with system endpoints
var ReservedEndpointNames = []string{
    "collections",
    "auth",
    "users",
    "apikeys",
    "doc",
    "health",
}

// IsReservedEndpointName checks if a name conflicts with system endpoints (case-insensitive)
func IsReservedEndpointName(name string) bool {
    lower := strings.ToLower(name)
    for _, reserved := range ReservedEndpointNames {
        if lower == reserved {
            return true
        }
    }
    return false
}

// SQLReservedKeywords are SQL keywords that cannot be used as collection or column names
var SQLReservedKeywords = map[string]bool{
    // DDL Keywords
    "alter": true, "create": true, "drop": true, "truncate": true,
    "rename": true, "comment": true,
    
    // DML Keywords
    "select": true, "insert": true, "update": true, "delete": true,
    "merge": true, "replace": true,
    
    // Transaction Keywords
    "commit": true, "rollback": true, "savepoint": true, "transaction": true,
    "begin": true, "end": true, "start": true,
    
    // Query Keywords
    "from": true, "where": true, "join": true, "inner": true, "outer": true,
    "left": true, "right": true, "full": true, "cross": true,
    "on": true, "using": true, "natural": true,
    
    // Filtering/Grouping Keywords
    "order": true, "group": true, "having": true, "limit": true,
    "offset": true, "fetch": true, "distinct": true, "all": true,
    
    // Set Operations
    "union": true, "intersect": true, "except": true, "minus": true,
    
    // Logical Operators
    "and": true, "or": true, "not": true, "xor": true,
    
    // Comparison Operators
    "in": true, "exists": true, "between": true, "like": true,
    "is": true, "null": true, "isnull": true, "notnull": true,
    
    // Database Objects
    "table": true, "view": true, "index": true, "trigger": true,
    "function": true, "procedure": true, "database": true, "schema": true,
    "sequence": true, "constraint": true,
    
    // Constraint Keywords
    "primary": true, "foreign": true, "key": true, "unique": true,
    "check": true, "references": true, "cascade": true, "restrict": true,
    "default": true, "auto_increment": true, "serial": true,
    
    // User/Permission Keywords
    "user": true, "grant": true, "revoke": true, "role": true,
    "privilege": true, "with": true,
    
    // Data Types (commonly reserved)
    "int": true, "integer": true, "bigint": true, "smallint": true,
    "varchar": true, "char": true, "text": true, "blob": true,
    "decimal": true, "numeric": true, "float": true, "double": true,
    "real": true, "boolean": true, "bool": true, "date": true,
    "time": true, "timestamp": true, "datetime": true, "interval": true,
    "json": true, "jsonb": true, "array": true,
    
    // Other Common Keywords
    "as": true, "case": true, "when": true, "then": true, "else": true,
    "if": true, "elseif": true, "while": true, "loop": true, "repeat": true,
    "for": true, "do": true, "return": true, "declare": true,
    "set": true, "values": true, "into": true, "by": true,
    "asc": true, "desc": true, "nulls": true, "first": true, "last": true,
}

// IsReservedKeyword checks if a name is a SQL reserved keyword (case-insensitive)
func IsReservedKeyword(name string) bool {
    return SQLReservedKeywords[strings.ToLower(name)]
}
```

#### TR-2: Collection Validation Function

File: `cmd/moon/internal/handlers/collections.go`

```go
func validateCollectionName(name string, registry *registry.SchemaRegistry) error {
    // 1. Empty check
    if strings.TrimSpace(name) == "" {
        return fmt.Errorf("collection name cannot be empty")
    }
    
    // 2. Length validation
    if len(name) < constants.MinCollectionNameLength {
        return fmt.Errorf("collection name must be at least %d characters", constants.MinCollectionNameLength)
    }
    if len(name) > constants.MaxCollectionNameLength {
        return fmt.Errorf("collection name must not exceed %d characters", constants.MaxCollectionNameLength)
    }
    
    // 3. Reserved endpoint check (case-insensitive)
    if constants.IsReservedEndpointName(name) {
        return fmt.Errorf("collection name '%s' is reserved for system endpoints", name)
    }
    
    // 4. Pattern validation
    if !collectionNameRegex.MatchString(name) {
        return fmt.Errorf("collection name must start with a letter and contain only letters, numbers, and underscores")
    }
    
    // 5. Reserved keyword check
    if constants.IsReservedKeyword(name) {
        return fmt.Errorf("'%s' is a reserved keyword and cannot be used as a collection name", name)
    }
    
    // 6. System table/prefix check (from PRD-047)
    if constants.IsSystemTableOrPrefix(name) {
        return fmt.Errorf("collection name cannot start with 'moon_' or be 'moon' (reserved for system tables)")
    }
    
    // 7. Collection count limit
    collections := registry.List()
    if len(collections) >= constants.MaxCollectionsPerServer {
        return fmt.Errorf("maximum number of collections (%d) reached", constants.MaxCollectionsPerServer)
    }
    
    return nil
}
```

#### TR-3: Column Validation Function

File: `cmd/moon/internal/handlers/collections.go`

```go
func validateColumnName(name string) error {
    // 1. Empty check
    if strings.TrimSpace(name) == "" {
        return fmt.Errorf("column name cannot be empty")
    }
    
    // 2. Length validation
    if len(name) < constants.MinColumnNameLength {
        return fmt.Errorf("column name must be at least %d characters", constants.MinColumnNameLength)
    }
    if len(name) > constants.MaxColumnNameLength {
        return fmt.Errorf("column name must not exceed %d characters", constants.MaxColumnNameLength)
    }
    
    // 3. System column check
    if name == "id" || name == "ulid" {
        return fmt.Errorf("cannot add system column '%s'", name)
    }
    
    // 4. Pattern validation (lowercase only)
    columnNameRegex := regexp.MustCompile(constants.ColumnNamePattern)
    if !columnNameRegex.MatchString(name) {
        return fmt.Errorf("column name must start with a lowercase letter and contain only lowercase letters, numbers, and underscores")
    }
    
    // 5. Reserved keyword check
    if constants.IsReservedKeyword(name) {
        return fmt.Errorf("'%s' is a reserved keyword and cannot be used as a column name", name)
    }
    
    return nil
}

func validateColumnCount(collection *registry.Collection) error {
    // Count includes system columns (id, ulid)
    if len(collection.Columns) >= constants.MaxColumnsPerCollection {
        return fmt.Errorf("maximum number of columns (%d) reached for collection '%s'", constants.MaxColumnsPerCollection, collection.Name)
    }
    return nil
}
```

#### TR-4: Data Type Validation

File: `cmd/moon/internal/registry/registry.go`

```go
var validTypes = map[ColumnType]bool{
    TypeString:   true,
    TypeInteger:  true,
    TypeDecimal:  true,
    TypeBoolean:  true,
    TypeDatetime: true,
    TypeJSON:     true,
}

func ValidateColumnType(t ColumnType) bool {
    return validTypes[t]
}

func ValidateColumnTypeString(typeStr string) error {
    t := ColumnType(typeStr)
    if !ValidateColumnType(t) {
        // Check for deprecated types
        if typeStr == "text" {
            return fmt.Errorf("type 'text' is deprecated and no longer supported. Use 'string' instead")
        }
        if typeStr == "float" {
            return fmt.Errorf("type 'float' is deprecated and no longer supported. Use 'decimal' or 'integer' instead")
        }
        return fmt.Errorf("invalid column type '%s'. Supported types: string, integer, decimal, boolean, datetime, json", typeStr)
    }
    return nil
}
```

#### TR-5: Default Value Validation

File: `cmd/moon/internal/handlers/collections.go`

```go
func validateDefaultValue(column *registry.Column) error {
    if column.DefaultValue == nil {
        return nil // No default value specified
    }
    
    value := *column.DefaultValue
    
    // Check nullable constraint
    if value == "null" && !column.Nullable {
        return fmt.Errorf("default value cannot be null for non-nullable column '%s'", column.Name)
    }
    
    // Validate format based on type
    switch column.Type {
    case registry.TypeString:
        // Any string is valid
        return nil
        
    case registry.TypeInteger:
        if _, err := strconv.ParseInt(value, 10, 64); err != nil {
            return fmt.Errorf("default value '%s' is invalid for type 'integer'", value)
        }
        return nil
        
    case registry.TypeDecimal:
        if err := validateDecimalFormat(value); err != nil {
            return fmt.Errorf("default value '%s' is invalid for type 'decimal': %v", value, err)
        }
        return nil
        
    case registry.TypeBoolean:
        lower := strings.ToLower(value)
        if lower != "true" && lower != "false" {
            return fmt.Errorf("default value '%s' is invalid for type 'boolean'. Use 'true' or 'false'", value)
        }
        return nil
        
    case registry.TypeDatetime:
        if _, err := time.Parse(time.RFC3339, value); err != nil {
            return fmt.Errorf("default value '%s' is invalid for type 'datetime'. Use RFC3339 format (e.g., '2024-01-01T00:00:00Z')", value)
        }
        return nil
        
    case registry.TypeJSON:
        if !json.Valid([]byte(value)) {
            return fmt.Errorf("default value '%s' is invalid JSON", value)
        }
        return nil
        
    default:
        return fmt.Errorf("unknown column type '%s'", column.Type)
    }
}
```

#### TR-6: Query Validation

File: `cmd/moon/internal/handlers/data.go`

```go
func validateQueryParams(r *http.Request) error {
    query := r.URL.Query()
    
    // Count filters
    filterCount := 0
    for key := range query {
        if strings.Contains(key, "[") && strings.Contains(key, "]") {
            filterCount++
        }
    }
    if filterCount > constants.MaxFiltersPerRequest {
        return fmt.Errorf("maximum number of filters (%d) exceeded", constants.MaxFiltersPerRequest)
    }
    
    // Count sort fields
    sortParam := query.Get("sort")
    if sortParam != "" {
        sortFields := strings.Split(sortParam, ",")
        if len(sortFields) > constants.MaxSortFieldsPerRequest {
            return fmt.Errorf("maximum number of sort fields (%d) exceeded", constants.MaxSortFieldsPerRequest)
        }
    }
    
    return nil
}
```

#### TR-7: Transaction Wrapper

File: `cmd/moon/internal/db/transaction.go`

```go
func ExecuteSchemaChange(db *sql.DB, registry *registry.SchemaRegistry, fn func(tx *sql.Tx) (*registry.Collection, error)) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    collection, err := fn(tx)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    // Update registry only after successful commit
    if collection != nil {
        if err := registry.Set(collection); err != nil {
            // Critical error: database committed but registry update failed
            // Log error and trigger consistency check
            log.Error().Err(err).Str("collection", collection.Name).Msg("Failed to update registry after successful database commit")
            return fmt.Errorf("registry update failed: %w", err)
        }
    }
    
    return nil
}
```

#### TR-8: Query Timeout Enforcement

File: `cmd/moon/internal/db/query.go`

```go
func ExecuteQueryWithTimeout(ctx context.Context, db *sql.DB, query string, args []interface{}, timeout time.Duration) (*sql.Rows, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    start := time.Now()
    rows, err := db.QueryContext(ctx, query, args...)
    duration := time.Since(start)
    
    // Log slow queries
    if duration > time.Duration(constants.DefaultSlowQueryThreshold)*time.Millisecond {
        log.Warn().
            Dur("duration", duration).
            Str("query", query).
            Interface("params", args).
            Msg("Slow query detected")
    }
    
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("query execution timeout (%v)", timeout)
        }
        return nil, err
    }
    
    return rows, nil
}
```

#### TR-9: Configuration Structure

File: `cmd/moon/internal/config/config.go`

```yaml
database:
  connection: "sqlite"
  database: "/opt/moon/sqlite.db"
  query_timeout: 30          # seconds
  slow_query_threshold: 500  # milliseconds

limits:
  max_collections: 1000
  max_columns_per_collection: 100
  max_filters_per_request: 20
  max_sort_fields_per_request: 5
```

### Error Handling and Failure Modes

**EH-1: Validation Failure Responses**
- HTTP 400 Bad Request: Invalid input (name, type, format, pattern)
- HTTP 409 Conflict: Limit exceeded (collections, columns), duplicate constraint
- HTTP 504 Gateway Timeout: Query timeout exceeded
- HTTP 500 Internal Server Error: Database or registry error

**EH-2: Transaction Rollback Scenarios**
- Database DDL failure → Rollback transaction, registry unchanged
- Registry update failure after commit → Log critical error, trigger consistency check on next startup
- Partial failure → Always rollback to maintain consistency

**EH-3: Slow Query Handling**
- Log slow queries at WARN level with duration, SQL, and parameters
- Do NOT return error to client (query succeeded, just slow)
- Aggregate slow query metrics for performance monitoring

**EH-4: Reserved Name Conflicts**
- Check at collection creation time, not just routing level
- Case-insensitive comparison prevents bypass attempts
- Clear error message indicating which reserved name conflicted

## Acceptance

### AC-1: Reserved Endpoint Name Validation

- [ ] Update `validateCollectionName()` to check `IsReservedEndpointName()`
- [ ] Add reserved names list to `constants/validation.go`: `collections`, `auth`, `users`, `apikeys`, `doc`, `health`
- [ ] Test: `POST /collections:create` with `"collections"` returns HTTP 400
- [ ] Test: `"Collections"` (case-insensitive) returns HTTP 400
- [ ] Test: `"auth"`, `"users"`, `"apikeys"`, `"doc"` all return HTTP 400
- [ ] Test: `"products"` (non-reserved) is accepted

### AC-2: Collection Name Pattern and Length

- [ ] Enforce pattern: `^[a-zA-Z][a-zA-Z0-9_]*$`
- [ ] Enforce min length: 2 characters
- [ ] Enforce max length: 63 characters
- [ ] Test: `"ab"` is valid (min length)
- [ ] Test: `"a"` returns error (too short)
- [ ] Test: 63-char name is valid
- [ ] Test: 64-char name returns error
- [ ] Test: `"_products"` returns error (must start with letter)
- [ ] Test: `"123products"` returns error

### AC-3: Collection Count Limit

- [ ] Add `MaxCollectionsPerServer = 1000` to constants
- [ ] Check collection count before creating new collection
- [ ] Test: Create 1000 collections successfully
- [ ] Test: Attempt to create 1001st collection returns HTTP 409: `"maximum number of collections (1000) reached"`
- [ ] Test: Delete collection and create new one succeeds

### AC-4: Column Name Pattern and Length

- [ ] Enforce pattern: `^[a-z][a-z0-9_]*$` (lowercase only)
- [ ] Enforce min length: 3 characters
- [ ] Enforce max length: 63 characters
- [ ] Test: `"abc"` is valid (min length)
- [ ] Test: `"ab"` returns error (too short)
- [ ] Test: `"UserName"` returns error (uppercase not allowed)
- [ ] Test: `"user_name"` is valid
- [ ] Test: `"_name"` returns error (must start with letter)

### AC-5: System Column Protection

- [ ] Prevent adding columns named `id` or `ulid`
- [ ] Prevent removing columns `id` or `ulid` (if remove operation exists)
- [ ] Prevent renaming columns `id` or `ulid` (if rename operation exists)
- [ ] Test: `POST /collections/{name}:add_column` with `"id"` returns HTTP 400
- [ ] Test: Adding `"ulid"` returns HTTP 400
- [ ] Error message: `"cannot add system column 'id'"`

### AC-6: Column Count Limit

- [ ] Add `MaxColumnsPerCollection = 100` to constants
- [ ] Check column count (including system columns) before adding new column
- [ ] Test: Add 98 user columns (98 + 2 system = 100 total) succeeds
- [ ] Test: Attempt to add 99th user column returns HTTP 409: `"maximum number of columns (100) reached for collection 'products'"`

### AC-7: Data Type Validation - Supported Types

- [ ] Validate all column types using `registry.ValidateColumnType()`
- [ ] Accept only: `string`, `integer`, `decimal`, `boolean`, `datetime`, `json`
- [ ] Test: Create column with each supported type succeeds
- [ ] Test: Create column with `"varchar"` returns HTTP 400
- [ ] Error message: `"invalid column type 'varchar'. Supported types: string, integer, decimal, boolean, datetime, json"`

### AC-8: Data Type Validation - Deprecated Types

- [ ] Remove all references to `text` and `float` types from codebase
- [ ] Reject `"text"` with error: `"type 'text' is deprecated and no longer supported. Use 'string' instead"`
- [ ] Reject `"float"` with error: `"type 'float' is deprecated and no longer supported. Use 'decimal' or 'integer' instead"`
- [ ] Test: Attempt to create column with `"text"` returns HTTP 400
- [ ] Test: Attempt to create column with `"float"` returns HTTP 400
- [ ] Search codebase for `"text"` and `"float"` type references and remove all

### AC-9: Decimal Type Constraints

- [ ] Add `DecimalDefaultScale = 2` to constants
- [ ] Add `DecimalMaxScale = 10` to constants
- [ ] Validate decimal values before insert/update (if validation exists)
- [ ] Test: Insert `"19.99"` into decimal column succeeds
- [ ] Test: Insert `"19.999"` into decimal(2) column returns error (if scale enforcement exists)

### AC-10: Default Value Validation - Type Matching

- [ ] Implement `validateDefaultValue()` function
- [ ] Validate default value matches column type at creation time
- [ ] Test: String column with default `"active"` succeeds
- [ ] Test: Integer column with default `"42"` succeeds
- [ ] Test: Integer column with default `"abc"` returns HTTP 400
- [ ] Test: Decimal column with default `"19.99"` succeeds
- [ ] Test: Boolean column with default `"true"` succeeds
- [ ] Test: Boolean column with default `"yes"` returns HTTP 400

### AC-11: Default Value Validation - Format Requirements

- [ ] Test: Datetime column with default `"2024-01-01T00:00:00Z"` succeeds
- [ ] Test: Datetime column with default `"2024-01-01"` returns HTTP 400 (invalid RFC3339)
- [ ] Test: JSON column with default `"{}"` succeeds
- [ ] Test: JSON column with default `"invalid json"` returns HTTP 400

### AC-12: Default Value Validation - NULL Handling

- [ ] Test: Nullable column with `default_value: null` succeeds
- [ ] Test: Non-nullable column with `default_value: null` returns HTTP 400
- [ ] Error message: `"default value cannot be null for non-nullable column"`

### AC-13: Uniqueness Constraints

- [ ] Document unique constraint behavior in SPEC.md
- [ ] Document NULL value behavior: "Multiple NULL values are allowed in unique columns (SQL standard)"
- [ ] Test: Insert two NULL values into unique column succeeds
- [ ] Test: Insert duplicate non-NULL value returns HTTP 409: `"duplicate value for unique column '{column}'"`
- [ ] Verify constraint naming: `idx_{collection}_{column}_unique`

### AC-14: Transaction Atomicity

- [ ] Wrap all schema operations in database transactions
- [ ] Implement `ExecuteSchemaChange()` transaction wrapper
- [ ] Test: Simulate database failure during CREATE TABLE → Registry unchanged
- [ ] Test: Simulate registry failure after successful commit → Log critical error
- [ ] Test: Rollback restores database state (if possible)

### AC-15: Query Parameter Validation - Filter Limits

- [ ] Add `MaxFiltersPerRequest = 20` to constants
- [ ] Count filter parameters: `field[operator]=value`
- [ ] Test: Request with 20 filters succeeds
- [ ] Test: Request with 21 filters returns HTTP 400: `"maximum number of filters (20) exceeded"`

### AC-16: Query Parameter Validation - Sort Limits

- [ ] Add `MaxSortFieldsPerRequest = 5` to constants
- [ ] Count sort fields in `?sort=field1,-field2,field3`
- [ ] Test: Request with 5 sort fields succeeds
- [ ] Test: Request with 6 sort fields returns HTTP 400: `"maximum number of sort fields (5) exceeded"`

### AC-17: Query Parameter Validation - Valid Operators

- [ ] Define valid operators list in constants
- [ ] Validate operator before query execution
- [ ] Test: Each valid operator (`eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `contains`, `icontains`, `startswith`, `endswith`, `in`, `null`, `notnull`) works
- [ ] Test: Invalid operator `"equals"` returns HTTP 400: `"invalid filter operator 'equals'. Valid operators: ..."`

### AC-18: Query Parameter Validation - Type Checking

- [ ] Validate filter value types against column types before query execution
- [ ] Test: Integer column filter with `"abc"` returns HTTP 400
- [ ] Test: Boolean column filter with `"yes"` returns HTTP 400
- [ ] Test: Datetime column filter with `"2024-01-01"` (invalid RFC3339) returns HTTP 400

### AC-19: SQL Injection Prevention

- [ ] Verify ALL queries use parameterized statements
- [ ] Verify NO string concatenation of user input into SQL
- [ ] Validate column names against schema registry before query construction
- [ ] Test: Inject SQL keywords in filter values (e.g., `field[eq]='; DROP TABLE--`) → Query executes safely as parameter
- [ ] Review all query construction code for SQL injection vulnerabilities

### AC-20: Query Timeout Enforcement

- [ ] Add `query_timeout: 30` to database config (default: 30 seconds)
- [ ] Implement `ExecuteQueryWithTimeout()` using `context.WithTimeout()`
- [ ] Test: Long-running query (e.g., large table scan) times out after 30 seconds
- [ ] Test: Timeout returns HTTP 504: `"query execution timeout (30s)"`
- [ ] Verify timeout value is configurable via YAML

### AC-21: Slow Query Logging

- [ ] Add `slow_query_threshold: 500` to database config (default: 500ms)
- [ ] Log queries exceeding threshold at WARN level
- [ ] Log format includes: duration, SQL, parameters
- [ ] Test: Query taking 600ms is logged as slow query
- [ ] Test: Query taking 400ms is NOT logged as slow
- [ ] Verify threshold is configurable via YAML

### AC-22: Configuration Updates

- [ ] Add `limits` section to YAML config with:
  - `max_collections: 1000`
  - `max_columns_per_collection: 100`
  - `max_filters_per_request: 20`
  - `max_sort_fields_per_request: 5`
- [ ] Add to `database` section:
  - `query_timeout: 30`
  - `slow_query_threshold: 500`
- [ ] Load configuration at startup
- [ ] Validate configuration values (positive integers)

### AC-23: Documentation Updates - SPEC.md

- [ ] Document reserved endpoint names
- [ ] Document collection and column naming rules
- [ ] Document schema size limits (1000 collections, 100 columns)
- [ ] Document uniqueness constraint behavior (NULL handling)
- [ ] Document default value format requirements
- [ ] Document query parameter limits
- [ ] Document transaction atomicity guarantees
- [ ] Document query timeout and slow query threshold

### AC-24: Documentation Updates - API Template

- [ ] Update `cmd/moon/internal/handlers/templates/doc.md.tmpl` with validation rules
- [ ] Document reserved names
- [ ] Document naming patterns and length limits
- [ ] Document supported data types (remove text/float)
- [ ] Document query parameter limits
- [ ] Add examples for default values

### AC-25: Error Message Consistency

- [ ] Ensure all validation errors return JSON: `{"error": "message"}`
- [ ] Ensure HTTP status codes are correct (400, 409, 504, 500)
- [ ] Test each error scenario for consistent format
- [ ] Verify error messages are clear and actionable

### AC-26: Integration Testing

- [ ] Test end-to-end collection creation with all validations
- [ ] Test end-to-end column addition with all validations
- [ ] Test query with maximum filters and sort fields
- [ ] Test query timeout with long-running query
- [ ] Test transaction rollback on database failure
- [ ] Test schema limit enforcement (1000 collections, 100 columns)

### AC-27: Performance Testing

- [ ] Measure validation overhead: < 1ms per request
- [ ] Test with 1000 collections in registry (memory usage)
- [ ] Test with collection having 100 columns (query performance)
- [ ] Verify query timeout works with slow queries
- [ ] Verify slow query logging does not impact performance

### AC-28: Testing Checklist

- [ ] All unit tests pass for validation functions
- [ ] All integration tests pass for endpoints
- [ ] All error scenarios produce correct HTTP status and message
- [ ] Transaction rollback tested with simulated failures
- [ ] SQL injection prevention verified with malicious inputs
- [ ] Query timeout tested with long-running queries
- [ ] Test coverage >= 90% for new validation code

---

### Implementation Checklist

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `USAGE.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

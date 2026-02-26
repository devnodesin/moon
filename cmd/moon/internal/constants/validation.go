package constants

import "strings"

// Validation constants for input validation and data constraints.
const (
	// DefaultVarcharMaxLength is the default maximum length for VARCHAR fields
	// when no specific length is specified in the schema.
	// Used in: validation/validator.go
	// Purpose: Standard SQL VARCHAR constraint for string fields
	// Default: 255 characters
	DefaultVarcharMaxLength = 255

	// MinAPIKeyLength is the minimum required length for API keys.
	// Used in: middleware/apikey.go
	// Purpose: Security requirement to ensure API keys have sufficient entropy
	// Default: 40 characters
	MinAPIKeyLength = 40

	// Collection name constraints (PRD-047, PRD-048)
	// MinCollectionNameLength is the minimum length for collection names.
	MinCollectionNameLength = 2
	// MaxCollectionNameLength is the maximum length for collection names.
	// PostgreSQL has 63-char limit, MySQL has 64-char limit.
	MaxCollectionNameLength = 63
	// MaxCollections is the maximum number of collections allowed per server.
	MaxCollections = 1000
	// MaxCollectionsPerServer is an alias for MaxCollections (deprecated).
	MaxCollectionsPerServer = MaxCollections

	// Column name constraints (PRD-048)
	// MinColumnNameLength is the minimum length for column names.
	MinColumnNameLength = 3
	// MaxColumnNameLength is the maximum length for column names.
	MaxColumnNameLength = 63
	// MaxColumnsPerCollection is the maximum number of columns per collection.
	// This includes system columns (id, ulid).
	MaxColumnsPerCollection = 100
	// SystemColumnsCount is the number of automatically added system columns.
	// System columns are: id (auto-increment primary key), ulid (external ID).
	SystemColumnsCount = 2

	// Data type constraints (PRD-048)
	// DecimalDefaultScale is the default number of decimal places.
	DecimalDefaultScale = 2
	// DecimalMaxScale is the maximum number of decimal places.
	DecimalMaxScale = 10

	// Query constraints (PRD-048)
	// MaxFiltersPerRequest is the maximum number of filter parameters per request.
	MaxFiltersPerRequest = 20
	// MaxSortFieldsPerRequest is the maximum number of sort fields per request.
	MaxSortFieldsPerRequest = 5

	// Performance constraints (PRD-048)
	// DefaultQueryTimeout is the default query timeout in seconds.
	DefaultQueryTimeout = 30
	// DefaultSlowQueryThreshold is the default threshold for slow query logging in milliseconds.
	DefaultSlowQueryThreshold = 500

	// System prefix protection (PRD-047)
	// SystemPrefix is the reserved prefix for system tables.
	SystemPrefix = "moon_"
	// SystemNamespace is the reserved namespace name.
	SystemNamespace = "moon"
)

// DatetimeFormat is the ISO 8601 datetime format used throughout the application.
// This is equivalent to time.RFC3339 but named to reflect the ISO 8601 standard.
const DatetimeFormat = "2006-01-02T15:04:05Z07:00"

// Regular expression patterns for validation.
const (
	// CollectionNamePattern is the regex pattern for valid collection names.
	// Pattern: Must start with a letter, followed by letters, numbers, or underscores.
	// Used in: handlers/collections.go
	// Purpose: Ensures collection names are valid SQL identifiers
	// Note: Collection names are normalized to lowercase before storage (PRD-047)
	CollectionNamePattern = `^[a-zA-Z][a-zA-Z0-9_]*$`

	// CollectionNamePatternLowercase is the regex for normalized (lowercase) collection names.
	CollectionNamePatternLowercase = `^[a-z][a-z0-9_]*$`

	// ColumnNamePattern is the regex pattern for valid column names.
	// Pattern: Must start with a lowercase letter, followed by lowercase letters, numbers, or underscores.
	// Note: Uppercase is rejected, not auto-converted (PRD-048)
	ColumnNamePattern = `^[a-z][a-z0-9_]*$`
)

// ReservedEndpointNames are collection names that conflict with system endpoints.
// These names cannot be used as collection names (case-insensitive).
var ReservedEndpointNames = []string{
	"collections",
	"auth",
	"users",
	"apikeys",
	"doc",
	"health",
}

// IsReservedEndpointName checks if a name conflicts with system endpoints (case-insensitive).
func IsReservedEndpointName(name string) bool {
	lower := strings.ToLower(name)
	for _, reserved := range ReservedEndpointNames {
		if lower == reserved {
			return true
		}
	}
	return false
}

// IsSystemTableOrPrefix checks if a name is a system table or uses reserved prefix.
// Returns true if the name starts with "moon_" or is exactly "moon" (case-insensitive).
func IsSystemTableOrPrefix(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasPrefix(lower, SystemPrefix) || lower == SystemNamespace
}

// SQLReservedKeywords are SQL keywords that cannot be used as collection or column names.
// This map provides O(1) lookup for reserved keyword checking.
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

// IsReservedKeyword checks if a name is a SQL reserved keyword (case-insensitive).
func IsReservedKeyword(name string) bool {
	return SQLReservedKeywords[strings.ToLower(name)]
}

// ValidFilterOperators are the supported filter operators for query parameters.
var ValidFilterOperators = map[string]bool{
	"eq":         true, // equals
	"ne":         true, // not equals
	"gt":         true, // greater than
	"gte":        true, // greater than or equal
	"lt":         true, // less than
	"lte":        true, // less than or equal
	"like":       true, // pattern match
	"in":         true, // value in list
	"contains":   true, // string contains (case-sensitive)
	"icontains":  true, // string contains (case-insensitive)
	"startswith": true, // string starts with
	"endswith":   true, // string ends with
	"null":       true, // is null
	"notnull":    true, // is not null
}

// IsValidFilterOperator checks if an operator is valid.
func IsValidFilterOperator(op string) bool {
	return ValidFilterOperators[strings.ToLower(op)]
}

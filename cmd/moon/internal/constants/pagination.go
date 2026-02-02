package constants

// Pagination constants for list operations and data retrieval.
// These define default limits and offsets for paginated API responses.
const (
	// MinPageSize is the minimum allowed page size for paginated responses.
	// This is hardcoded and not configurable.
	// Used in: handlers/data.go, pagination/validator.go
	// Purpose: Ensures at least 1 record is returned
	// Default: 1 record
	MinPageSize = 1

	// DefaultPaginationLimit is the default number of records to return
	// in a single paginated response when no limit is specified.
	// This is the fallback if not configured in the config file.
	// Used in: handlers/data.go
	// Purpose: Prevents excessive memory usage and improves response times
	// Default: 15 records
	DefaultPaginationLimit = 15

	// MaxPaginationLimit is the maximum allowed limit for paginated responses.
	// This is the fallback if not configured in the config file.
	// Used in: handlers/data.go, handlers/users.go
	// Purpose: Prevents clients from requesting too many records at once
	// Default: 200 records (configurable via pagination.max_page_size)
	MaxPaginationLimit = 200

	// DefaultPaginationOffset is the default starting position for pagination
	// when no offset is specified.
	// Used in: handlers/data.go
	// Purpose: Standard starting point for paginated queries
	// Default: 0 (start from beginning)
	DefaultPaginationOffset = 0
)

// Query parameter names for pagination.
const (
	// QueryParamLimit is the URL query parameter name for specifying page size.
	// Used in: handlers/data.go
	QueryParamLimit = "limit"

	// QueryParamOffset is the URL query parameter name for specifying page offset.
	// Used in: handlers/data.go
	QueryParamOffset = "offset"

	// QueryParamID is the URL query parameter name for resource ID.
	// Used in: handlers/data.go
	QueryParamID = "id"
)

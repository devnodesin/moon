package main

import (
	"context"
	"fmt"
	"time"
)

// ---------------------------------------------------------------------------
// DatabaseAdapter defines the persistence interface used by all upper layers.
// Every backend (SQLite, PostgreSQL, MySQL) must implement this contract.
// ---------------------------------------------------------------------------

// DatabaseAdapter is the single internal interface for all database
// operations. Implementations must be safe for concurrent use.
type DatabaseAdapter interface {
	// Ping verifies that the database is reachable.
	Ping(ctx context.Context) error

	// Close releases the underlying connection.
	Close() error

	// ExecDDL executes a raw DDL statement (CREATE TABLE, ALTER, etc.).
	ExecDDL(ctx context.Context, ddl string) error

	// QueryRows returns rows matching the given options. It returns the
	// result rows, the total count of matching rows (before pagination),
	// and any error.
	QueryRows(ctx context.Context, table string, opts QueryOptions) ([]map[string]any, int, error)

	// InsertRow inserts a single row into the given table.
	InsertRow(ctx context.Context, table string, data map[string]any) error

	// UpdateRow updates the row identified by id in the given table.
	UpdateRow(ctx context.Context, table string, id string, data map[string]any) error

	// DeleteRow deletes the row identified by id from the given table.
	DeleteRow(ctx context.Context, table string, id string) error

	// ListTables returns the names of all physical user tables.
	ListTables(ctx context.Context) ([]string, error)

	// DescribeTable returns column definitions for the given table.
	DescribeTable(ctx context.Context, table string) ([]ColumnInfo, error)

	// CountRows returns the number of rows in the given table.
	CountRows(ctx context.Context, table string) (int, error)
}

// ---------------------------------------------------------------------------
// Query option types
// ---------------------------------------------------------------------------

// Filter represents a single column filter.
type Filter struct {
	Field string
	Op    string // "eq", "neq", "gt", "gte", "lt", "lte", "like"
	Value any
}

// SortField represents a single sort directive.
type SortField struct {
	Field string
	Desc  bool
}

// QueryOptions carries filtering, sorting, pagination, and projection
// parameters for QueryRows.
type QueryOptions struct {
	Filters      []Filter
	Sort         []SortField
	Page         int
	PerPage      int
	Fields       []string
	Search       string
	SearchFields []string
}

// ---------------------------------------------------------------------------
// Column introspection
// ---------------------------------------------------------------------------

// ColumnInfo describes a single column in a physical table.
type ColumnInfo struct {
	Name     string
	Type     string
	Nullable bool
	PK       bool
	Unique   bool // single-column UNIQUE constraint (excludes PK)
}

// ---------------------------------------------------------------------------
// Adapter errors
// ---------------------------------------------------------------------------

// AdapterError wraps backend-specific errors so SQL details never leak
// into API responses.
type AdapterError struct {
	Op      string // operation name (e.g. "InsertRow", "QueryRows")
	Table   string
	Message string
	Err     error // underlying backend error
}

func (e *AdapterError) Error() string {
	if e.Table != "" {
		return fmt.Sprintf("adapter: %s on %q: %s", e.Op, e.Table, e.Message)
	}
	return fmt.Sprintf("adapter: %s: %s", e.Op, e.Message)
}

func (e *AdapterError) Unwrap() error {
	return e.Err
}

// newAdapterError creates an AdapterError wrapping the underlying error.
func newAdapterError(op, table, message string, err error) *AdapterError {
	return &AdapterError{
		Op:      op,
		Table:   table,
		Message: message,
		Err:     err,
	}
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

// NewDatabaseAdapter creates the appropriate adapter based on the database
// configuration. The adapter is ready to use after construction; callers
// should call Ping to verify connectivity.
func NewDatabaseAdapter(cfg DatabaseConfig, logger *Logger) (DatabaseAdapter, error) {
	switch cfg.Connection {
	case DBConnectionSQLite:
		return NewSQLiteAdapter(cfg, logger)
	case DBConnectionPostgres:
		return NewPostgresAdapter(cfg, logger)
	case DBConnectionMySQL:
		return NewMySQLAdapter(cfg, logger)
	default:
		return nil, fmt.Errorf("unsupported database connection type: %q", cfg.Connection)
	}
}

// ---------------------------------------------------------------------------
// Slow-query logging helper
// ---------------------------------------------------------------------------

// logSlowQuery emits a warning if duration exceeds the configured threshold.
func logSlowQuery(logger *Logger, table, op string, start time.Time, thresholdMs int) {
	elapsed := time.Since(start)
	ms := elapsed.Milliseconds()
	if ms > int64(thresholdMs) {
		logger.Warn("slow query",
			"table", table,
			"op", op,
			"duration_ms", ms,
			"threshold_ms", thresholdMs,
		)
	}
}

package main

import (
	"context"
	"fmt"
)

// ---------------------------------------------------------------------------
// PostgreSQL type mapping constants
// ---------------------------------------------------------------------------

const (
	PGTypeID       = "TEXT"
	PGTypeString   = "TEXT"
	PGTypeInteger  = "BIGINT"
	PGTypeDecimal  = "NUMERIC(19,2)"
	PGTypeBoolean  = "BOOLEAN"
	PGTypeDatetime = "TIMESTAMP"
	PGTypeJSON     = "JSON"
)

// ---------------------------------------------------------------------------
// PostgresAdapter is a stub implementation of DatabaseAdapter for PostgreSQL.
// ---------------------------------------------------------------------------

// PostgresAdapter implements DatabaseAdapter for PostgreSQL. Currently a stub
// that returns "not implemented" errors for all operations.
type PostgresAdapter struct {
	cfg    DatabaseConfig
	logger *Logger
}

// NewPostgresAdapter creates a stub PostgreSQL adapter.
func NewPostgresAdapter(cfg DatabaseConfig, logger *Logger) (*PostgresAdapter, error) {
	return &PostgresAdapter{cfg: cfg, logger: logger}, nil
}

func (a *PostgresAdapter) Ping(ctx context.Context) error {
	return fmt.Errorf("postgres adapter not implemented")
}

func (a *PostgresAdapter) Close() error {
	return nil
}

func (a *PostgresAdapter) ExecDDL(ctx context.Context, ddl string) error {
	return fmt.Errorf("postgres adapter not implemented")
}

func (a *PostgresAdapter) QueryRows(ctx context.Context, table string, opts QueryOptions) ([]map[string]any, int, error) {
	return nil, 0, fmt.Errorf("postgres adapter not implemented")
}

func (a *PostgresAdapter) InsertRow(ctx context.Context, table string, data map[string]any) error {
	return fmt.Errorf("postgres adapter not implemented")
}

func (a *PostgresAdapter) UpdateRow(ctx context.Context, table string, id string, data map[string]any) error {
	return fmt.Errorf("postgres adapter not implemented")
}

func (a *PostgresAdapter) DeleteRow(ctx context.Context, table string, id string) error {
	return fmt.Errorf("postgres adapter not implemented")
}

func (a *PostgresAdapter) ListTables(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("postgres adapter not implemented")
}

func (a *PostgresAdapter) DescribeTable(ctx context.Context, table string) ([]ColumnInfo, error) {
	return nil, fmt.Errorf("postgres adapter not implemented")
}

func (a *PostgresAdapter) CountRows(ctx context.Context, table string) (int, error) {
	return 0, fmt.Errorf("postgres adapter not implemented")
}

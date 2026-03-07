package main

import (
	"context"
	"fmt"
)

// ---------------------------------------------------------------------------
// MySQL type mapping constants
// ---------------------------------------------------------------------------

const (
	MySQLTypeID       = "TEXT"
	MySQLTypeString   = "TEXT"
	MySQLTypeInteger  = "BIGINT"
	MySQLTypeDecimal  = "DECIMAL(19,2)"
	MySQLTypeBoolean  = "BOOLEAN"
	MySQLTypeDatetime = "TIMESTAMP"
	MySQLTypeJSON     = "JSON"
)

// ---------------------------------------------------------------------------
// MySQLAdapter is a stub implementation of DatabaseAdapter for MySQL.
// ---------------------------------------------------------------------------

// MySQLAdapter implements DatabaseAdapter for MySQL. Currently a stub
// that returns "not implemented" errors for all operations.
type MySQLAdapter struct {
	cfg    DatabaseConfig
	logger *Logger
}

// NewMySQLAdapter creates a stub MySQL adapter.
func NewMySQLAdapter(cfg DatabaseConfig, logger *Logger) (*MySQLAdapter, error) {
	return &MySQLAdapter{cfg: cfg, logger: logger}, nil
}

func (a *MySQLAdapter) Ping(ctx context.Context) error {
	return fmt.Errorf("mysql adapter not implemented")
}

func (a *MySQLAdapter) Close() error {
	return nil
}

func (a *MySQLAdapter) ExecDDL(ctx context.Context, ddl string) error {
	return fmt.Errorf("mysql adapter not implemented")
}

func (a *MySQLAdapter) QueryRows(ctx context.Context, table string, opts QueryOptions) ([]map[string]any, int, error) {
	return nil, 0, fmt.Errorf("mysql adapter not implemented")
}

func (a *MySQLAdapter) InsertRow(ctx context.Context, table string, data map[string]any) error {
	return fmt.Errorf("mysql adapter not implemented")
}

func (a *MySQLAdapter) UpdateRow(ctx context.Context, table string, id string, data map[string]any) error {
	return fmt.Errorf("mysql adapter not implemented")
}

func (a *MySQLAdapter) DeleteRow(ctx context.Context, table string, id string) error {
	return fmt.Errorf("mysql adapter not implemented")
}

func (a *MySQLAdapter) ListTables(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("mysql adapter not implemented")
}

func (a *MySQLAdapter) DescribeTable(ctx context.Context, table string) ([]ColumnInfo, error) {
	return nil, fmt.Errorf("mysql adapter not implemented")
}

func (a *MySQLAdapter) CountRows(ctx context.Context, table string) (int, error) {
	return 0, fmt.Errorf("mysql adapter not implemented")
}

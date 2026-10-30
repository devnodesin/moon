package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ---------------------------------------------------------------------------
// PostgreSQL type mapping constants (SPEC.md §9.3)
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
// PostgresAdapter implements DatabaseAdapter for PostgreSQL.
//
// It uses the pgx stdlib driver (registered as "pgx"), which exposes the
// standard library database/sql interface. The shared query helpers in
// adapter_shared.go emit "?" placeholders, which this adapter rewrites to
// PostgreSQL-native "$n" form via pgQuestionToDollar before execution because
// the pgx stdlib driver does not perform that conversion.
// ---------------------------------------------------------------------------

// PostgresAdapter implements DatabaseAdapter for PostgreSQL.
type PostgresAdapter struct {
	db                 *sql.DB
	cfg                DatabaseConfig
	logger             *Logger
	slowQueryThreshold int
	queryTimeout       int
}

// NewPostgresAdapter opens a PostgreSQL connection using the pgx stdlib driver.
// Connection pooling is configured conservatively for a single-process service.
func NewPostgresAdapter(cfg DatabaseConfig, logger *Logger) (*PostgresAdapter, error) {
	dsn := buildPGDSN(cfg)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, newAdapterError("NewPostgresAdapter", "", "failed to open database", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresAdapter{
		db:                 db,
		cfg:                cfg,
		logger:             logger,
		slowQueryThreshold: cfg.SlowQueryThreshold,
		queryTimeout:       cfg.QueryTimeout,
	}, nil
}

// buildPGDSN constructs a PostgreSQL connection string from the resolved
// configuration. sslmode defaults to disable for local/dev deployments; set
// the host connection to require TLS in production via the DSN override if
// needed.
func buildPGDSN(cfg DatabaseConfig) string {
	// Allow an explicit DSN override through the Database field when it looks
	// like a full connection string (contains "://"). Otherwise build one.
	if strings.Contains(cfg.Database, "://") {
		return cfg.Database
	}
	sslmode := "disable"
	host := cfg.Host
	port := "5432"
	if h, p, ok := strings.Cut(cfg.Host, ":"); ok {
		host = h
		port = p
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, host, port, cfg.Database, sslmode)
}

// withTimeout derives a context with the configured query timeout.
func (a *PostgresAdapter) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(a.queryTimeout)*time.Second)
}

// Ping verifies that the database is reachable.
func (a *PostgresAdapter) Ping(ctx context.Context) error {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	if err := a.db.PingContext(ctx2); err != nil {
		return newAdapterError("Ping", "", "database unreachable", err)
	}
	return nil
}

// Close releases the underlying database connection.
func (a *PostgresAdapter) Close() error {
	return a.db.Close()
}

// ExecDDL executes a raw DDL statement.
func (a *PostgresAdapter) ExecDDL(ctx context.Context, ddl string) error {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()
	_, err := a.db.ExecContext(ctx2, ddl)
	logSlowQuery(a.logger, "", "ExecDDL", start, a.slowQueryThreshold)
	if err != nil {
		return newAdapterError("ExecDDL", "", "DDL execution failed", err)
	}
	return nil
}

// QueryRows returns rows matching the given options.
func (a *PostgresAdapter) QueryRows(ctx context.Context, table string, opts QueryOptions) ([]map[string]any, int, error) {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	where, args := buildWhereClause(opts)

	// Total count query. Rewrite "?" placeholders to PostgreSQL "$n" form.
	countSQL := pgQuestionToDollar(fmt.Sprintf("SELECT COUNT(*) FROM %s%s", quoteIdent(table), where))
	var total int
	if err := a.db.QueryRowContext(ctx2, countSQL, args...).Scan(&total); err != nil {
		logSlowQuery(a.logger, table, "QueryRows/count", start, a.slowQueryThreshold)
		return nil, 0, newAdapterError("QueryRows", table, "count query failed", err)
	}

	fields := "*"
	if len(opts.Fields) > 0 {
		quoted := make([]string, len(opts.Fields))
		for i, f := range opts.Fields {
			quoted[i] = quoteIdent(f)
		}
		fields = strings.Join(quoted, ", ")
	}

	orderClause := ""
	if len(opts.Sort) > 0 {
		parts := make([]string, len(opts.Sort))
		for i, s := range opts.Sort {
			dir := "ASC"
			if s.Desc {
				dir = "DESC"
			}
			parts[i] = fmt.Sprintf("%s %s", quoteIdent(s.Field), dir)
		}
		orderClause = " ORDER BY " + strings.Join(parts, ", ")
	}

	page := opts.Page
	if page < 1 {
		page = 1
	}
	perPage := opts.PerPage
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	offset := (page - 1) * perPage

	selectSQL := pgQuestionToDollar(fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT ? OFFSET ?",
		fields, quoteIdent(table), where, orderClause))
	selectArgs := append(args, perPage, offset)

	rows, err := a.db.QueryContext(ctx2, selectSQL, selectArgs...)
	logSlowQuery(a.logger, table, "QueryRows", start, a.slowQueryThreshold)
	if err != nil {
		return nil, 0, newAdapterError("QueryRows", table, "select query failed", err)
	}
	defer rows.Close()

	results, err := scanRows(rows)
	if err != nil {
		return nil, 0, newAdapterError("QueryRows", table, "row scan failed", err)
	}

	return results, total, nil
}

// InsertRow inserts a single row into the given table.
func (a *PostgresAdapter) InsertRow(ctx context.Context, table string, data map[string]any) error {
	if len(data) == 0 {
		return newAdapterError("InsertRow", table, "no data provided", nil)
	}

	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]any, 0, len(data))
	idx := 1
	for col, val := range data {
		columns = append(columns, quoteIdent(col))
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		values = append(values, val)
		idx++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quoteIdent(table),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	_, err := a.db.ExecContext(ctx2, query, values...)
	logSlowQuery(a.logger, table, "InsertRow", start, a.slowQueryThreshold)
	if err != nil {
		return newAdapterError("InsertRow", table, "insert failed", err)
	}
	return nil
}

// UpdateRow updates the row identified by id in the given table.
func (a *PostgresAdapter) UpdateRow(ctx context.Context, table string, id string, data map[string]any) error {
	if len(data) == 0 {
		return newAdapterError("UpdateRow", table, "no data provided", nil)
	}

	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	setClauses := make([]string, 0, len(data))
	values := make([]any, 0, len(data)+1)
	idx := 1
	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdent(col), idx))
		values = append(values, val)
		idx++
	}
	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d",
		quoteIdent(table),
		strings.Join(setClauses, ", "),
		quoteIdent("id"), idx)

	_, err := a.db.ExecContext(ctx2, query, values...)
	logSlowQuery(a.logger, table, "UpdateRow", start, a.slowQueryThreshold)
	if err != nil {
		return newAdapterError("UpdateRow", table, "update failed", err)
	}
	return nil
}

// DeleteRow deletes the row identified by id from the given table.
func (a *PostgresAdapter) DeleteRow(ctx context.Context, table string, id string) error {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1",
		quoteIdent(table), quoteIdent("id"))

	_, err := a.db.ExecContext(ctx2, query, id)
	logSlowQuery(a.logger, table, "DeleteRow", start, a.slowQueryThreshold)
	if err != nil {
		return newAdapterError("DeleteRow", table, "delete failed", err)
	}
	return nil
}

// ListTables returns the names of all physical base tables in the public
// schema. Internal PostgreSQL tables and those prefixed with "pg_" are excluded
// by the information_schema query itself; the reserved "moon_" prefix is a
// physical table that remains discoverable (the upper layer applies API
// visibility rules per SPEC.md §9.11).
func (a *PostgresAdapter) ListTables(ctx context.Context) ([]string, error) {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	const q = `SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`

	rows, err := a.db.QueryContext(ctx2, q)
	logSlowQuery(a.logger, "", "ListTables", start, a.slowQueryThreshold)
	if err != nil {
		return nil, newAdapterError("ListTables", "", "table list failed", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, newAdapterError("ListTables", "", "scan failed", err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return nil, newAdapterError("ListTables", "", "iteration failed", err)
	}
	return tables, nil
}

// DescribeTable returns column definitions for the given table using
// information_schema.
func (a *PostgresAdapter) DescribeTable(ctx context.Context, table string) ([]ColumnInfo, error) {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	const q = `SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position`

	rows, err := a.db.QueryContext(ctx2, q, table)
	logSlowQuery(a.logger, table, "DescribeTable", start, a.slowQueryThreshold)
	if err != nil {
		return nil, newAdapterError("DescribeTable", table, "columns query failed", err)
	}

	var columns []ColumnInfo
	for rows.Next() {
		var name, dataType, nullable string
		if err := rows.Scan(&name, &dataType, &nullable); err != nil {
			rows.Close()
			return nil, newAdapterError("DescribeTable", table, "scan failed", err)
		}
		columns = append(columns, ColumnInfo{
			Name:     name,
			Type:     pgDataTypeToMoon(dataType),
			Nullable: nullable == "YES",
		})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, newAdapterError("DescribeTable", table, "iteration failed", err)
	}

	pkCols, err := a.primaryKeyColumns(ctx2, table)
	if err != nil {
		return nil, err
	}
	for i := range columns {
		if pkCols[columns[i].Name] {
			columns[i].PK = true
		}
	}

	uniqueCols, err := a.detectUniqueColumns(ctx2, table)
	if err != nil {
		return nil, err
	}
	for i := range columns {
		if uniqueCols[columns[i].Name] {
			columns[i].Unique = true
		}
	}

	return columns, nil
}

// primaryKeyColumns returns the set of column names that are part of the
// primary key for the given table.
func (a *PostgresAdapter) primaryKeyColumns(ctx context.Context, table string) (map[string]bool, error) {
	result := make(map[string]bool)
	const q = `SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.table_schema = 'public'
			AND tc.table_name = $1
			AND tc.constraint_type = 'PRIMARY KEY'`
	rows, err := a.db.QueryContext(ctx, q, table)
	if err != nil {
		return nil, newAdapterError("DescribeTable", table, "primary key query failed", err)
	}
	defer rows.Close()
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, newAdapterError("DescribeTable", table, "scan failed", err)
		}
		result[col] = true
	}
	return result, rows.Err()
}

// detectUniqueColumns returns the set of column names that have a single-column
// UNIQUE constraint, excluding primary keys.
func (a *PostgresAdapter) detectUniqueColumns(ctx context.Context, table string) (map[string]bool, error) {
	result := make(map[string]bool)

	const q = `SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.table_schema = 'public'
			AND tc.table_name = $1
			AND tc.constraint_type = 'UNIQUE'`
	rows, err := a.db.QueryContext(ctx, q, table)
	if err != nil {
		return nil, newAdapterError("DescribeTable", table, "unique query failed", err)
	}
	defer rows.Close()

	var candidates []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, newAdapterError("DescribeTable", table, "scan failed", err)
		}
		candidates = append(candidates, col)
	}
	if err := rows.Err(); err != nil {
		return nil, newAdapterError("DescribeTable", table, "iteration failed", err)
	}

	// Only single-column unique constraints are reported (multi-column
	// uniqueness is not exposed as a per-column unique flag).
	for _, col := range candidates {
		n, err := a.columnUniqueConstraintCount(ctx, table, col)
		if err != nil {
			return nil, err
		}
		if n == 1 {
			result[col] = true
		}
	}
	return result, nil
}

// columnUniqueConstraintCount returns how many UNIQUE constraints a single
// column participates in (used to exclude columns that are only part of
// multi-column unique constraints).
func (a *PostgresAdapter) columnUniqueConstraintCount(ctx context.Context, table, column string) (int, error) {
	const q = `SELECT COUNT(*)
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.table_schema = 'public'
			AND tc.table_name = $1
			AND tc.constraint_type = 'UNIQUE'
			AND kcu.column_name = $2`
	var count int
	if err := a.db.QueryRowContext(ctx, q, table, column).Scan(&count); err != nil {
		return 0, newAdapterError("DescribeTable", table, "unique count query failed", err)
	}
	return count, nil
}

// CountRows returns the number of rows in the given table.
func (a *PostgresAdapter) CountRows(ctx context.Context, table string) (int, error) {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(table))
	var count int
	err := a.db.QueryRowContext(ctx2, query).Scan(&count)
	logSlowQuery(a.logger, table, "CountRows", start, a.slowQueryThreshold)
	if err != nil {
		return 0, newAdapterError("CountRows", table, "count failed", err)
	}
	return count, nil
}

// ---------------------------------------------------------------------------
// Placeholder conversion
//
// The shared query builder (buildWhereClause) emits "?" placeholders so the
// SQLite adapter can use them directly. The pgx stdlib driver only accepts
// PostgreSQL-native "$n" positional placeholders, so the PostgreSQL adapter
// rewrites "?" to "$1, $2, ..." immediately before execution.
// ---------------------------------------------------------------------------

// pgQuestionToDollar rewrites "?" placeholders in a SQL string to PostgreSQL
// positional "$n" placeholders, preserving order.
func pgQuestionToDollar(sql string) string {
	var b strings.Builder
	n := 1
	for i := 0; i < len(sql); i++ {
		if sql[i] == '?' {
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
			n++
			continue
		}
		b.WriteByte(sql[i])
	}
	return b.String()
}
func pgDataTypeToMoon(dataType string) string {
	switch {
	case dataType == "text":
		return "TEXT"
	case dataType == "bigint", dataType == "integer", dataType == "smallint":
		return "BIGINT"
	case strings.HasPrefix(dataType, "numeric"), strings.HasPrefix(dataType, "decimal"):
		return "NUMERIC"
	case dataType == "boolean":
		return "BOOLEAN"
	case strings.HasPrefix(dataType, "timestamp"), strings.HasPrefix(dataType, "date"):
		return "TIMESTAMP"
	case dataType == "json", dataType == "jsonb":
		return "JSON"
	default:
		return "TEXT"
	}
}

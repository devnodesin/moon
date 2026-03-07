package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------
// SQLite type mapping constants
// ---------------------------------------------------------------------------

const (
	SQLiteTypeID       = "TEXT"
	SQLiteTypeString   = "TEXT"
	SQLiteTypeInteger  = "INTEGER"
	SQLiteTypeDecimal  = "NUMERIC"
	SQLiteTypeBoolean  = "INTEGER"
	SQLiteTypeDatetime = "TEXT"
	SQLiteTypeJSON     = "TEXT"
)

// ---------------------------------------------------------------------------
// SQLiteAdapter implements DatabaseAdapter for SQLite.
// ---------------------------------------------------------------------------

// SQLiteAdapter provides a SQLite-backed implementation of the
// DatabaseAdapter interface. It opens the database in WAL mode for
// concurrent read access.
type SQLiteAdapter struct {
	db                 *sql.DB
	cfg                DatabaseConfig
	logger             *Logger
	slowQueryThreshold int
	queryTimeout       int
}

// NewSQLiteAdapter opens a SQLite database at the path specified in
// cfg.Database. WAL journal mode is enabled immediately.
func NewSQLiteAdapter(cfg DatabaseConfig, logger *Logger) (*SQLiteAdapter, error) {
	dsn := cfg.Database + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, newAdapterError("NewSQLiteAdapter", "", "failed to open database", err)
	}

	// Enable WAL mode explicitly.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, newAdapterError("NewSQLiteAdapter", "", "failed to set WAL mode", err)
	}

	return &SQLiteAdapter{
		db:                 db,
		cfg:                cfg,
		logger:             logger,
		slowQueryThreshold: cfg.SlowQueryThreshold,
		queryTimeout:       cfg.QueryTimeout,
	}, nil
}

// withTimeout derives a context with the configured query timeout.
func (a *SQLiteAdapter) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(a.queryTimeout)*time.Second)
}

// Ping verifies that the database is reachable.
func (a *SQLiteAdapter) Ping(ctx context.Context) error {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	if err := a.db.PingContext(ctx2); err != nil {
		return newAdapterError("Ping", "", "database unreachable", err)
	}
	return nil
}

// Close releases the underlying database connection.
func (a *SQLiteAdapter) Close() error {
	return a.db.Close()
}

// ExecDDL executes a raw DDL statement.
func (a *SQLiteAdapter) ExecDDL(ctx context.Context, ddl string) error {
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
func (a *SQLiteAdapter) QueryRows(ctx context.Context, table string, opts QueryOptions) ([]map[string]any, int, error) {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	where, args := buildWhereClause(opts)

	// Total count query.
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", quoteIdent(table), where)
	var total int
	if err := a.db.QueryRowContext(ctx2, countSQL, args...).Scan(&total); err != nil {
		logSlowQuery(a.logger, table, "QueryRows/count", start, a.slowQueryThreshold)
		return nil, 0, newAdapterError("QueryRows", table, "count query failed", err)
	}

	// Build SELECT.
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

	selectSQL := fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT ? OFFSET ?",
		fields, quoteIdent(table), where, orderClause)
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
func (a *SQLiteAdapter) InsertRow(ctx context.Context, table string, data map[string]any) error {
	if len(data) == 0 {
		return newAdapterError("InsertRow", table, "no data provided", nil)
	}

	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]any, 0, len(data))
	for col, val := range data {
		columns = append(columns, quoteIdent(col))
		placeholders = append(placeholders, "?")
		values = append(values, val)
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
func (a *SQLiteAdapter) UpdateRow(ctx context.Context, table string, id string, data map[string]any) error {
	if len(data) == 0 {
		return newAdapterError("UpdateRow", table, "no data provided", nil)
	}

	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	setClauses := make([]string, 0, len(data))
	values := make([]any, 0, len(data)+1)
	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", quoteIdent(col)))
		values = append(values, val)
	}
	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",
		quoteIdent(table),
		strings.Join(setClauses, ", "),
		quoteIdent("id"))

	_, err := a.db.ExecContext(ctx2, query, values...)
	logSlowQuery(a.logger, table, "UpdateRow", start, a.slowQueryThreshold)
	if err != nil {
		return newAdapterError("UpdateRow", table, "update failed", err)
	}
	return nil
}

// DeleteRow deletes the row identified by id from the given table.
func (a *SQLiteAdapter) DeleteRow(ctx context.Context, table string, id string) error {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?",
		quoteIdent(table), quoteIdent("id"))

	_, err := a.db.ExecContext(ctx2, query, id)
	logSlowQuery(a.logger, table, "DeleteRow", start, a.slowQueryThreshold)
	if err != nil {
		return newAdapterError("DeleteRow", table, "delete failed", err)
	}
	return nil
}

// ListTables returns the names of all physical user tables. Internal SQLite
// tables and those prefixed with "sqlite_" are excluded.
func (a *SQLiteAdapter) ListTables(ctx context.Context) ([]string, error) {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	rows, err := a.db.QueryContext(ctx2, "PRAGMA table_list")
	logSlowQuery(a.logger, "", "ListTables", start, a.slowQueryThreshold)
	if err != nil {
		return nil, newAdapterError("ListTables", "", "table list failed", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, newAdapterError("ListTables", "", "columns read failed", err)
	}

	// PRAGMA table_list returns: schema, name, type, ncol, wr, strict
	nameIdx := -1
	typeIdx := -1
	for i, c := range cols {
		switch c {
		case "name":
			nameIdx = i
		case "type":
			typeIdx = i
		}
	}
	if nameIdx < 0 || typeIdx < 0 {
		return nil, newAdapterError("ListTables", "", "unexpected PRAGMA table_list schema", nil)
	}

	var tables []string
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, newAdapterError("ListTables", "", "scan failed", err)
		}

		name := fmt.Sprintf("%v", vals[nameIdx])
		ttype := fmt.Sprintf("%v", vals[typeIdx])

		// Include only regular tables, skip views, shadow tables, etc.
		if ttype != "table" {
			continue
		}
		// Skip internal SQLite tables.
		if strings.HasPrefix(name, "sqlite_") {
			continue
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return nil, newAdapterError("ListTables", "", "iteration failed", err)
	}
	return tables, nil
}

// DescribeTable returns column definitions for the given table using
// PRAGMA table_info.
func (a *SQLiteAdapter) DescribeTable(ctx context.Context, table string) ([]ColumnInfo, error) {
	ctx2, cancel := a.withTimeout(ctx)
	defer cancel()
	start := time.Now()

	query := fmt.Sprintf("PRAGMA table_info(%s)", quoteIdent(table))
	rows, err := a.db.QueryContext(ctx2, query)
	logSlowQuery(a.logger, table, "DescribeTable", start, a.slowQueryThreshold)
	if err != nil {
		return nil, newAdapterError("DescribeTable", table, "table_info failed", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return nil, newAdapterError("DescribeTable", table, "scan failed", err)
		}
		columns = append(columns, ColumnInfo{
			Name:     name,
			Type:     colType,
			Nullable: notNull == 0,
			PK:       pk == 1,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, newAdapterError("DescribeTable", table, "iteration failed", err)
	}

	uniqueCols := a.detectUniqueColumns(ctx2, table)
	for i := range columns {
		if uniqueCols[columns[i].Name] {
			columns[i].Unique = true
		}
	}

	return columns, nil
}

// detectUniqueColumns returns the set of column names that have a
// single-column UNIQUE constraint or index, excluding primary keys.
func (a *SQLiteAdapter) detectUniqueColumns(ctx context.Context, table string) map[string]bool {
	result := make(map[string]bool)

	query := fmt.Sprintf("PRAGMA index_list(%s)", quoteIdent(table))
	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return result
	}
	defer rows.Close()

	var uniqueIndexNames []string
	for rows.Next() {
		var seq int
		var name string
		var isUnique int
		var origin string
		var partial int
		if err := rows.Scan(&seq, &name, &isUnique, &origin, &partial); err != nil {
			return result
		}
		if isUnique == 1 && origin != "pk" {
			uniqueIndexNames = append(uniqueIndexNames, name)
		}
	}

	for _, idxName := range uniqueIndexNames {
		infoQuery := fmt.Sprintf("PRAGMA index_info(%s)", quoteIdent(idxName))
		infoRows, err := a.db.QueryContext(ctx, infoQuery)
		if err != nil {
			continue
		}
		var colNames []string
		for infoRows.Next() {
			var seqno, cid int
			var colName string
			if err := infoRows.Scan(&seqno, &cid, &colName); err != nil {
				break
			}
			colNames = append(colNames, colName)
		}
		infoRows.Close()

		if len(colNames) == 1 {
			result[colNames[0]] = true
		}
	}

	return result
}

// CountRows returns the number of rows in the given table.
func (a *SQLiteAdapter) CountRows(ctx context.Context, table string) (int, error) {
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
// SQL helpers
// ---------------------------------------------------------------------------

// quoteIdent wraps a SQL identifier in double quotes to prevent injection.
// Any embedded double quotes are escaped by doubling.
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// filterOpSQL maps filter operator names to SQL operators.
var filterOpSQL = map[string]string{
	"eq":   "=",
	"ne":   "!=",
	"neq":  "!=",
	"gt":   ">",
	"gte":  ">=",
	"lt":   "<",
	"lte":  "<=",
	"like": "LIKE",
}

// buildWhereClause builds a WHERE clause from QueryOptions filters and
// search parameters. Returns the clause string (including " WHERE " prefix
// if non-empty) and the corresponding parameter values.
func buildWhereClause(opts QueryOptions) (string, []any) {
	var conditions []string
	var args []any

	for _, f := range opts.Filters {
		if f.Op == "in" {
			values, ok := f.Value.([]string)
			if !ok || len(values) == 0 {
				continue
			}
			placeholders := make([]string, len(values))
			for i, v := range values {
				placeholders[i] = "?"
				args = append(args, v)
			}
			conditions = append(conditions,
				fmt.Sprintf("%s IN (%s)", quoteIdent(f.Field), strings.Join(placeholders, ", ")))
			continue
		}
		sqlOp, ok := filterOpSQL[f.Op]
		if !ok {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s %s ?", quoteIdent(f.Field), sqlOp))
		args = append(args, f.Value)
	}

	if opts.Search != "" && len(opts.SearchFields) > 0 {
		var searchConds []string
		for _, sf := range opts.SearchFields {
			searchConds = append(searchConds, fmt.Sprintf("%s LIKE ?", quoteIdent(sf)))
			args = append(args, "%"+opts.Search+"%")
		}
		conditions = append(conditions, "("+strings.Join(searchConds, " OR ")+")")
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

// scanRows reads all rows from a *sql.Rows into a slice of maps.
func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}
	return results, nil
}

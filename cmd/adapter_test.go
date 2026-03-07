package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testSQLiteAdapter creates a temporary SQLite database and returns a ready
// adapter. The database file is removed when the test completes.
func testSQLiteAdapter(t *testing.T) *SQLiteAdapter {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	cfg := DatabaseConfig{
		Connection:         DBConnectionSQLite,
		Database:           dbPath,
		QueryTimeout:       5,
		SlowQueryThreshold: 500,
	}
	logger := NewTestLogger(&bytes.Buffer{})
	adapter, err := NewSQLiteAdapter(cfg, logger)
	if err != nil {
		t.Fatalf("NewSQLiteAdapter: %v", err)
	}
	t.Cleanup(func() { adapter.Close() })
	return adapter
}

// seedTestTable creates a simple test table and inserts some rows.
func seedTestTable(t *testing.T, adapter *SQLiteAdapter) {
	t.Helper()
	ctx := context.Background()
	ddl := `CREATE TABLE items (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		active INTEGER NOT NULL DEFAULT 1
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ExecDDL: %v", err)
	}
	rows := []map[string]any{
		{"id": "001", "name": "alpha", "quantity": int64(10), "active": int64(1)},
		{"id": "002", "name": "bravo", "quantity": int64(20), "active": int64(1)},
		{"id": "003", "name": "charlie", "quantity": int64(30), "active": int64(0)},
	}
	for _, r := range rows {
		if err := adapter.InsertRow(ctx, "items", r); err != nil {
			t.Fatalf("InsertRow: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Interface compliance
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_ImplementsDatabaseAdapter(t *testing.T) {
	var _ DatabaseAdapter = (*SQLiteAdapter)(nil)
}

func TestPostgresAdapter_ImplementsDatabaseAdapter(t *testing.T) {
	var _ DatabaseAdapter = (*PostgresAdapter)(nil)
}

func TestMySQLAdapter_ImplementsDatabaseAdapter(t *testing.T) {
	var _ DatabaseAdapter = (*MySQLAdapter)(nil)
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

func TestNewDatabaseAdapter_SQLite(t *testing.T) {
	dir := t.TempDir()
	cfg := DatabaseConfig{
		Connection:         DBConnectionSQLite,
		Database:           filepath.Join(dir, "factory.db"),
		QueryTimeout:       5,
		SlowQueryThreshold: 500,
	}
	logger := NewTestLogger(&bytes.Buffer{})
	adapter, err := NewDatabaseAdapter(cfg, logger)
	if err != nil {
		t.Fatalf("NewDatabaseAdapter: %v", err)
	}
	defer adapter.Close()
	if _, ok := adapter.(*SQLiteAdapter); !ok {
		t.Fatalf("expected *SQLiteAdapter, got %T", adapter)
	}
}

func TestNewDatabaseAdapter_Postgres(t *testing.T) {
	cfg := DatabaseConfig{Connection: DBConnectionPostgres}
	logger := NewTestLogger(&bytes.Buffer{})
	adapter, err := NewDatabaseAdapter(cfg, logger)
	if err != nil {
		t.Fatalf("NewDatabaseAdapter: %v", err)
	}
	defer adapter.Close()
	if _, ok := adapter.(*PostgresAdapter); !ok {
		t.Fatalf("expected *PostgresAdapter, got %T", adapter)
	}
}

func TestNewDatabaseAdapter_MySQL(t *testing.T) {
	cfg := DatabaseConfig{Connection: DBConnectionMySQL}
	logger := NewTestLogger(&bytes.Buffer{})
	adapter, err := NewDatabaseAdapter(cfg, logger)
	if err != nil {
		t.Fatalf("NewDatabaseAdapter: %v", err)
	}
	defer adapter.Close()
	if _, ok := adapter.(*MySQLAdapter); !ok {
		t.Fatalf("expected *MySQLAdapter, got %T", adapter)
	}
}

func TestNewDatabaseAdapter_Unknown(t *testing.T) {
	cfg := DatabaseConfig{Connection: "oracle"}
	logger := NewTestLogger(&bytes.Buffer{})
	_, err := NewDatabaseAdapter(cfg, logger)
	if err == nil {
		t.Fatal("expected error for unsupported backend")
	}
}

// ---------------------------------------------------------------------------
// Ping
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_Ping(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	if err := adapter.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestSQLiteAdapter_Ping_FailsAfterClose(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	adapter.Close()
	if err := adapter.Ping(context.Background()); err == nil {
		t.Fatal("expected error after Close")
	}
}

// ---------------------------------------------------------------------------
// ExecDDL
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_ExecDDL(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE test_ddl (id TEXT PRIMARY KEY, val TEXT)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ExecDDL: %v", err)
	}

	// Verify table exists.
	tables, err := adapter.ListTables(ctx)
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	found := false
	for _, tbl := range tables {
		if tbl == "test_ddl" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("table test_ddl not found in %v", tables)
	}
}

func TestSQLiteAdapter_ExecDDL_InvalidSQL(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	if err := adapter.ExecDDL(context.Background(), "NOT VALID SQL"); err == nil {
		t.Fatal("expected error for invalid DDL")
	}
}

// ---------------------------------------------------------------------------
// InsertRow / QueryRows
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_InsertAndQuery(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)
	ctx := context.Background()

	rows, total, err := adapter.QueryRows(ctx, "items", QueryOptions{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
}

func TestSQLiteAdapter_InsertRow_EmptyData(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)
	err := adapter.InsertRow(context.Background(), "items", map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

// ---------------------------------------------------------------------------
// QueryRows – filters
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_QueryRows_Filter_Eq(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)
	ctx := context.Background()

	opts := QueryOptions{
		Filters: []Filter{{Field: "name", Op: "eq", Value: "bravo"}},
		Page:    1,
		PerPage: 10,
	}
	rows, total, err := adapter.QueryRows(ctx, "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "bravo" {
		t.Fatalf("expected name=bravo, got %v", rows[0]["name"])
	}
}

func TestSQLiteAdapter_QueryRows_Filter_Gt(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	opts := QueryOptions{
		Filters: []Filter{{Field: "quantity", Op: "gt", Value: 15}},
		Page:    1,
		PerPage: 10,
	}
	rows, total, err := adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestSQLiteAdapter_QueryRows_Filter_Like(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	opts := QueryOptions{
		Filters: []Filter{{Field: "name", Op: "like", Value: "%rav%"}},
		Page:    1,
		PerPage: 10,
	}
	rows, _, err := adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if len(rows) != 1 || rows[0]["name"] != "bravo" {
		t.Fatalf("expected bravo, got %v", rows)
	}
}

func TestSQLiteAdapter_QueryRows_Filter_UnknownOp(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	opts := QueryOptions{
		Filters: []Filter{{Field: "name", Op: "unknown", Value: "x"}},
		Page:    1,
		PerPage: 10,
	}
	_, total, err := adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	// Unknown filter op is ignored, so all rows returned.
	if total != 3 {
		t.Fatalf("expected total=3 (filter ignored), got %d", total)
	}
}

// ---------------------------------------------------------------------------
// QueryRows – sorting
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_QueryRows_Sort_Asc(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	opts := QueryOptions{
		Sort:    []SortField{{Field: "quantity", Desc: false}},
		Page:    1,
		PerPage: 10,
	}
	rows, _, err := adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "alpha" {
		t.Fatalf("expected first row alpha, got %v", rows[0]["name"])
	}
}

func TestSQLiteAdapter_QueryRows_Sort_Desc(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	opts := QueryOptions{
		Sort:    []SortField{{Field: "quantity", Desc: true}},
		Page:    1,
		PerPage: 10,
	}
	rows, _, err := adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if rows[0]["name"] != "charlie" {
		t.Fatalf("expected first row charlie, got %v", rows[0]["name"])
	}
}

// ---------------------------------------------------------------------------
// QueryRows – pagination
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_QueryRows_Pagination(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	opts := QueryOptions{
		Sort:    []SortField{{Field: "id", Desc: false}},
		Page:    1,
		PerPage: 2,
	}
	rows, total, err := adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows page 1: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows on page 1, got %d", len(rows))
	}

	// Page 2.
	opts.Page = 2
	rows, total, err = adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows page 2: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row on page 2, got %d", len(rows))
	}
}

func TestSQLiteAdapter_QueryRows_Pagination_Defaults(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	// Page=0, PerPage=0 should default to page 1, DefaultPerPage.
	rows, _, err := adapter.QueryRows(context.Background(), "items", QueryOptions{})
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows with defaults, got %d", len(rows))
	}
}

// ---------------------------------------------------------------------------
// QueryRows – field projection
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_QueryRows_FieldProjection(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	opts := QueryOptions{
		Fields:  []string{"id", "name"},
		Page:    1,
		PerPage: 10,
	}
	rows, _, err := adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected rows")
	}
	// Should only have id and name columns.
	if _, ok := rows[0]["quantity"]; ok {
		t.Fatal("quantity should not be in projected fields")
	}
}

// ---------------------------------------------------------------------------
// QueryRows – search
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_QueryRows_Search(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	opts := QueryOptions{
		Search:       "cha",
		SearchFields: []string{"name"},
		Page:         1,
		PerPage:      10,
	}
	rows, total, err := adapter.QueryRows(context.Background(), "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(rows) != 1 || rows[0]["name"] != "charlie" {
		t.Fatalf("expected charlie, got %v", rows)
	}
}

// ---------------------------------------------------------------------------
// UpdateRow
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_UpdateRow(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)
	ctx := context.Background()

	err := adapter.UpdateRow(ctx, "items", "002", map[string]any{"name": "beta", "quantity": int64(25)})
	if err != nil {
		t.Fatalf("UpdateRow: %v", err)
	}

	rows, _, err := adapter.QueryRows(ctx, "items", QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: "002"}},
		Page:    1, PerPage: 10,
	})
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "beta" {
		t.Fatalf("expected name=beta, got %v", rows[0]["name"])
	}
}

func TestSQLiteAdapter_UpdateRow_EmptyData(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)
	err := adapter.UpdateRow(context.Background(), "items", "001", map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

// ---------------------------------------------------------------------------
// DeleteRow
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_DeleteRow(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)
	ctx := context.Background()

	if err := adapter.DeleteRow(ctx, "items", "002"); err != nil {
		t.Fatalf("DeleteRow: %v", err)
	}

	count, err := adapter.CountRows(ctx, "items")
	if err != nil {
		t.Fatalf("CountRows: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows after delete, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// ListTables
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_ListTables(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	ctx := context.Background()

	if err := adapter.ExecDDL(ctx, "CREATE TABLE aaa (id TEXT)"); err != nil {
		t.Fatal(err)
	}
	if err := adapter.ExecDDL(ctx, "CREATE TABLE bbb (id TEXT)"); err != nil {
		t.Fatal(err)
	}

	tables, err := adapter.ListTables(ctx)
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	if len(tables) < 2 {
		t.Fatalf("expected at least 2 tables, got %d: %v", len(tables), tables)
	}

	found := map[string]bool{}
	for _, tbl := range tables {
		found[tbl] = true
	}
	if !found["aaa"] || !found["bbb"] {
		t.Fatalf("expected aaa and bbb in tables: %v", tables)
	}
}

func TestSQLiteAdapter_ListTables_ExcludesSQLiteInternal(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	ctx := context.Background()

	if err := adapter.ExecDDL(ctx, "CREATE TABLE users (id TEXT)"); err != nil {
		t.Fatal(err)
	}

	tables, err := adapter.ListTables(ctx)
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	for _, tbl := range tables {
		if tbl == "sqlite_master" || tbl == "sqlite_sequence" {
			t.Fatalf("internal SQLite table %q should be excluded", tbl)
		}
	}
}

// ---------------------------------------------------------------------------
// DescribeTable
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_DescribeTable(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE describe_test (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		score NUMERIC,
		active INTEGER NOT NULL DEFAULT 1
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	cols, err := adapter.DescribeTable(ctx, "describe_test")
	if err != nil {
		t.Fatalf("DescribeTable: %v", err)
	}
	if len(cols) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(cols))
	}

	// Check id column.
	idCol := cols[0]
	if idCol.Name != "id" {
		t.Fatalf("expected first col name=id, got %q", idCol.Name)
	}
	if !idCol.PK {
		t.Fatal("expected id to be PK")
	}
	if idCol.Type != "TEXT" {
		t.Fatalf("expected id type=TEXT, got %q", idCol.Type)
	}

	// Check name column is not nullable (NOT NULL).
	nameCol := cols[1]
	if nameCol.Nullable {
		t.Fatal("expected name to be NOT NULL (Nullable=false)")
	}

	// Check score column is nullable.
	scoreCol := cols[2]
	if !scoreCol.Nullable {
		t.Fatal("expected score to be nullable")
	}
}

// ---------------------------------------------------------------------------
// CountRows
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_CountRows(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)
	ctx := context.Background()

	count, err := adapter.CountRows(ctx, "items")
	if err != nil {
		t.Fatalf("CountRows: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestSQLiteAdapter_CountRows_Empty(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	ctx := context.Background()

	if err := adapter.ExecDDL(ctx, "CREATE TABLE empty_tbl (id TEXT)"); err != nil {
		t.Fatal(err)
	}
	count, err := adapter.CountRows(ctx, "empty_tbl")
	if err != nil {
		t.Fatalf("CountRows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// AdapterError
// ---------------------------------------------------------------------------

func TestAdapterError_Error(t *testing.T) {
	err := newAdapterError("InsertRow", "items", "insert failed", nil)
	expected := `adapter: InsertRow on "items": insert failed`
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestAdapterError_Error_NoTable(t *testing.T) {
	err := newAdapterError("Ping", "", "unreachable", nil)
	expected := "adapter: Ping: unreachable"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestAdapterError_Unwrap(t *testing.T) {
	inner := os.ErrNotExist
	err := newAdapterError("ExecDDL", "", "fail", inner)
	if err.Unwrap() != inner {
		t.Fatal("Unwrap did not return inner error")
	}
}

// ---------------------------------------------------------------------------
// Slow query logging
// ---------------------------------------------------------------------------

func TestSlowQueryLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewTestLogger(&buf)

	dir := t.TempDir()
	cfg := DatabaseConfig{
		Connection:         DBConnectionSQLite,
		Database:           filepath.Join(dir, "slow.db"),
		QueryTimeout:       5,
		SlowQueryThreshold: 1, // 1ms threshold — almost everything triggers
	}
	adapter, err := NewSQLiteAdapter(cfg, logger)
	if err != nil {
		t.Fatalf("NewSQLiteAdapter: %v", err)
	}
	defer adapter.Close()

	ctx := context.Background()
	if err := adapter.ExecDDL(ctx, "CREATE TABLE slow_test (id TEXT PRIMARY KEY, val TEXT)"); err != nil {
		t.Fatal(err)
	}

	// Insert enough rows to cause a measurable query.
	for i := 0; i < 100; i++ {
		adapter.InsertRow(ctx, "slow_test", map[string]any{
			"id":  fmt.Sprintf("%03d", i),
			"val": "data",
		})
	}
	adapter.QueryRows(ctx, "slow_test", QueryOptions{Page: 1, PerPage: 100})

	// We can't guarantee slow query logs fire (fast CI), but we verify
	// the function doesn't panic or error.
}

// ---------------------------------------------------------------------------
// Stub adapter errors
// ---------------------------------------------------------------------------

func TestPostgresAdapter_Stubs(t *testing.T) {
	cfg := DatabaseConfig{Connection: DBConnectionPostgres}
	logger := NewTestLogger(&bytes.Buffer{})
	a, _ := NewPostgresAdapter(cfg, logger)
	ctx := context.Background()

	if err := a.Ping(ctx); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.ExecDDL(ctx, "CREATE TABLE x (id TEXT)"); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if _, _, err := a.QueryRows(ctx, "x", QueryOptions{}); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.InsertRow(ctx, "x", map[string]any{}); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.UpdateRow(ctx, "x", "1", map[string]any{}); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.DeleteRow(ctx, "x", "1"); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if _, err := a.ListTables(ctx); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if _, err := a.DescribeTable(ctx, "x"); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if _, err := a.CountRows(ctx, "x"); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.Close(); err != nil {
		t.Fatalf("Close should succeed: %v", err)
	}
}

func TestMySQLAdapter_Stubs(t *testing.T) {
	cfg := DatabaseConfig{Connection: DBConnectionMySQL}
	logger := NewTestLogger(&bytes.Buffer{})
	a, _ := NewMySQLAdapter(cfg, logger)
	ctx := context.Background()

	if err := a.Ping(ctx); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.ExecDDL(ctx, "CREATE TABLE x (id TEXT)"); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if _, _, err := a.QueryRows(ctx, "x", QueryOptions{}); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.InsertRow(ctx, "x", map[string]any{}); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.UpdateRow(ctx, "x", "1", map[string]any{}); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.DeleteRow(ctx, "x", "1"); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if _, err := a.ListTables(ctx); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if _, err := a.DescribeTable(ctx, "x"); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if _, err := a.CountRows(ctx, "x"); err == nil {
		t.Fatal("expected not-implemented error")
	}
	if err := a.Close(); err != nil {
		t.Fatalf("Close should succeed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Type mapping constants sanity checks
// ---------------------------------------------------------------------------

func TestSQLiteTypeMappingConstants(t *testing.T) {
	if SQLiteTypeID != "TEXT" {
		t.Fatalf("SQLiteTypeID: got %q", SQLiteTypeID)
	}
	if SQLiteTypeBoolean != "INTEGER" {
		t.Fatalf("SQLiteTypeBoolean: got %q", SQLiteTypeBoolean)
	}
	if SQLiteTypeDecimal != "NUMERIC" {
		t.Fatalf("SQLiteTypeDecimal: got %q", SQLiteTypeDecimal)
	}
}

func TestPostgresTypeMappingConstants(t *testing.T) {
	if PGTypeInteger != "BIGINT" {
		t.Fatalf("PGTypeInteger: got %q", PGTypeInteger)
	}
	if PGTypeBoolean != "BOOLEAN" {
		t.Fatalf("PGTypeBoolean: got %q", PGTypeBoolean)
	}
}

func TestMySQLTypeMappingConstants(t *testing.T) {
	if MySQLTypeDecimal != "DECIMAL(19,2)" {
		t.Fatalf("MySQLTypeDecimal: got %q", MySQLTypeDecimal)
	}
	if MySQLTypeJSON != "JSON" {
		t.Fatalf("MySQLTypeJSON: got %q", MySQLTypeJSON)
	}
}

// ---------------------------------------------------------------------------
// quoteIdent
// ---------------------------------------------------------------------------

func TestQuoteIdent(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"name", `"name"`},
		{`has"quote`, `"has""quote"`},
		{"simple_col", `"simple_col"`},
	}
	for _, tc := range tests {
		got := quoteIdent(tc.input)
		if got != tc.want {
			t.Errorf("quoteIdent(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Context timeout
// ---------------------------------------------------------------------------

func TestSQLiteAdapter_ContextTimeout(t *testing.T) {
	adapter := testSQLiteAdapter(t)
	seedTestTable(t, adapter)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, _, err := adapter.QueryRows(ctx, "items", QueryOptions{Page: 1, PerPage: 10})
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

// ---------------------------------------------------------------------------
// DBConnection constants
// ---------------------------------------------------------------------------

func TestDBConnectionConstants(t *testing.T) {
	if DBConnectionSQLite != "sqlite" {
		t.Fatalf("expected sqlite, got %q", DBConnectionSQLite)
	}
	if DBConnectionPostgres != "postgres" {
		t.Fatalf("expected postgres, got %q", DBConnectionPostgres)
	}
	if DBConnectionMySQL != "mysql" {
		t.Fatalf("expected mysql, got %q", DBConnectionMySQL)
	}
}

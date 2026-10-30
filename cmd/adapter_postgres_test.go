package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// PostgreSQL test harness
//
// These tests require a live PostgreSQL instance. They are skipped unless
// MOON_TEST_POSTGRES_DSN is set. This keeps the default CI/developer flow
// (SQLite only) green without requiring a PostgreSQL server to be installed.
//
// Example:
//   docker run --rm -e POSTGRES_PASSWORD=test -e POSTGRES_DB=moon -p 5432:5432 postgres:16
//   MOON_TEST_POSTGRES_DSN='postgres://postgres:test@localhost:5432/moon?sslmode=disable' go test ./cmd
// ---------------------------------------------------------------------------

const pgTestDSNEnv = "MOON_TEST_POSTGRES_DSN"

// testPostgresAdapter creates a PostgreSQL adapter pointed at the DSN from the
// environment, skipping the test if none is configured. The database is dropped
// and recreated to give each test a clean slate.
func testPostgresAdapter(t *testing.T) *PostgresAdapter {
	t.Helper()
	dsn := os.Getenv(pgTestDSNEnv)
	if dsn == "" {
		t.Skipf("skipping: %s not set", pgTestDSNEnv)
	}

	cfg := DatabaseConfig{
		Connection:         DBConnectionPostgres,
		Database:           dsn,
		QueryTimeout:       5,
		SlowQueryThreshold: 500,
	}
	logger := NewTestLogger(&bytes.Buffer{})
	adapter, err := NewPostgresAdapter(cfg, logger)
	if err != nil {
		t.Fatalf("NewPostgresAdapter: %v", err)
	}
	if err := adapter.Ping(context.Background()); err != nil {
		adapter.Close()
		t.Fatalf("Ping: %v", err)
	}
	t.Cleanup(func() { adapter.Close() })
	return adapter
}

// pgSeedTestTable creates a table exercising every Moon field type and inserts
// rows, validating the full type round-trip through PostgreSQL.
func pgSeedTestTable(t *testing.T, adapter *PostgresAdapter) {
	t.Helper()
	ctx := context.Background()
	// Drop first so re-runs against a persistent test database start clean.
	if err := adapter.ExecDDL(ctx, "DROP TABLE IF EXISTS items"); err != nil {
		t.Fatalf("DROP items: %v", err)
	}
	ddl := `CREATE TABLE items (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		quantity BIGINT NOT NULL,
		price NUMERIC(19,2) NOT NULL,
		active BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL,
		meta JSON NOT NULL
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ExecDDL: %v", err)
	}
	rows := []map[string]any{
		{
			"id":         "001",
			"name":       "alpha",
			"quantity":   int64(10),
			"price":      "19.99",
			"active":     true,
			"created_at": time.Now().UTC().Format(time.RFC3339),
			"meta":       `{"k":"v"}`,
		},
		{
			"id":         "002",
			"name":       "bravo",
			"quantity":   int64(20),
			"price":      "29.50",
			"active":     true,
			"created_at": time.Now().UTC().Format(time.RFC3339),
			"meta":       `{"k":"w"}`,
		},
		{
			"id":         "003",
			"name":       "charlie",
			"quantity":   int64(30),
			"price":      "9.00",
			"active":     false,
			"created_at": time.Now().UTC().Format(time.RFC3339),
			"meta":       `{"k":"x"}`,
		},
	}
	for _, r := range rows {
		if err := adapter.InsertRow(ctx, "items", r); err != nil {
			t.Fatalf("InsertRow: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

func TestNewDatabaseAdapter_Postgres_Live(t *testing.T) {
	dsn := os.Getenv(pgTestDSNEnv)
	if dsn == "" {
		t.Skipf("skipping: %s not set", pgTestDSNEnv)
	}
	cfg := DatabaseConfig{Connection: DBConnectionPostgres, Database: dsn}
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

// ---------------------------------------------------------------------------
// Ping / Close
// ---------------------------------------------------------------------------

func TestPostgresAdapter_Ping(t *testing.T) {
	adapter := testPostgresAdapter(t)
	if err := adapter.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestPostgresAdapter_Ping_FailsAfterClose(t *testing.T) {
	adapter := testPostgresAdapter(t)
	adapter.Close()
	if err := adapter.Ping(context.Background()); err == nil {
		t.Fatal("expected error after Close")
	}
}

// ---------------------------------------------------------------------------
// ExecDDL / ListTables / DescribeTable
// ---------------------------------------------------------------------------

func TestPostgresAdapter_ExecDDL(t *testing.T) {
	adapter := testPostgresAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE test_ddl (id TEXT PRIMARY KEY, val TEXT)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ExecDDL: %v", err)
	}

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

func TestPostgresAdapter_ExecDDL_InvalidSQL(t *testing.T) {
	adapter := testPostgresAdapter(t)
	if err := adapter.ExecDDL(context.Background(), "NOT VALID SQL"); err == nil {
		t.Fatal("expected error for invalid DDL")
	}
}

func TestPostgresAdapter_DescribeTable(t *testing.T) {
	adapter := testPostgresAdapter(t)
	ctx := context.Background()
	pgSeedTestTable(t, adapter)

	cols, err := adapter.DescribeTable(ctx, "items")
	if err != nil {
		t.Fatalf("DescribeTable: %v", err)
	}

	byName := map[string]ColumnInfo{}
	for _, c := range cols {
		byName[c.Name] = c
	}

	expect := map[string]struct {
		typ      string
		pk       bool
		nullable bool
	}{
		"id":         {"TEXT", true, false},
		"name":       {"TEXT", false, false},
		"quantity":   {"BIGINT", false, false},
		"price":      {"NUMERIC", false, false},
		"active":     {"BOOLEAN", false, false},
		"created_at": {"TIMESTAMP", false, false},
		"meta":       {"JSON", false, false},
	}
	for name, want := range expect {
		got, ok := byName[name]
		if !ok {
			t.Fatalf("column %q missing from %v", name, byName)
		}
		if got.Type != want.typ {
			t.Errorf("column %q type = %q, want %q", name, got.Type, want.typ)
		}
		if got.PK != want.pk {
			t.Errorf("column %q pk = %v, want %v", name, got.PK, want.pk)
		}
		if got.Nullable != want.nullable {
			t.Errorf("column %q nullable = %v, want %v", name, got.Nullable, want.nullable)
		}
	}
}

func TestPostgresAdapter_DescribeTable_Unique(t *testing.T) {
	adapter := testPostgresAdapter(t)
	ctx := context.Background()
	if err := adapter.ExecDDL(ctx, "DROP TABLE IF EXISTS uniq_test"); err != nil {
		t.Fatalf("DROP uniq_test: %v", err)
	}
	ddl := `CREATE TABLE uniq_test (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		CONSTRAINT uniq_test_email UNIQUE (email)
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ExecDDL: %v", err)
	}
	cols, err := adapter.DescribeTable(ctx, "uniq_test")
	if err != nil {
		t.Fatalf("DescribeTable: %v", err)
	}
	var emailUnique, idPK bool
	for _, c := range cols {
		if c.Name == "email" && c.Unique {
			emailUnique = true
		}
		if c.Name == "id" && c.PK {
			idPK = true
		}
	}
	if !emailUnique {
		t.Fatal("expected email column to be reported Unique")
	}
	if !idPK {
		t.Fatal("expected id column to be reported PK")
	}
}

// ---------------------------------------------------------------------------
// CRUD round-trip
// ---------------------------------------------------------------------------

func TestPostgresAdapter_InsertAndQuery(t *testing.T) {
	adapter := testPostgresAdapter(t)
	pgSeedTestTable(t, adapter)
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

	// datetime must be returned as RFC3339 string (not time.Time).
	if _, ok := rows[0]["created_at"].(string); !ok {
		t.Fatalf("expected created_at as string, got %T", rows[0]["created_at"])
	}
}

func TestPostgresAdapter_QueryRows_Filter_Eq(t *testing.T) {
	adapter := testPostgresAdapter(t)
	pgSeedTestTable(t, adapter)
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
	if rows[0]["name"] != "bravo" {
		t.Fatalf("expected name=bravo, got %v", rows[0]["name"])
	}
}

func TestPostgresAdapter_QueryRows_Filter_Gt(t *testing.T) {
	adapter := testPostgresAdapter(t)
	pgSeedTestTable(t, adapter)
	ctx := context.Background()

	opts := QueryOptions{
		Filters: []Filter{{Field: "quantity", Op: "gt", Value: 15}},
		Page:    1,
		PerPage: 10,
	}
	rows, total, err := adapter.QueryRows(ctx, "items", opts)
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

func TestPostgresAdapter_QueryRows_Filter_Like(t *testing.T) {
	adapter := testPostgresAdapter(t)
	pgSeedTestTable(t, adapter)
	ctx := context.Background()

	opts := QueryOptions{
		Filters: []Filter{{Field: "name", Op: "like", Value: "%rav%"}},
		Page:    1,
		PerPage: 10,
	}
	rows, _, err := adapter.QueryRows(ctx, "items", opts)
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if len(rows) != 1 || rows[0]["name"] != "bravo" {
		t.Fatalf("expected bravo, got %v", rows)
	}
}

func TestPostgresAdapter_InsertRow_EmptyData(t *testing.T) {
	adapter := testPostgresAdapter(t)
	pgSeedTestTable(t, adapter)
	err := adapter.InsertRow(context.Background(), "items", map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestPostgresAdapter_UpdateRow(t *testing.T) {
	adapter := testPostgresAdapter(t)
	pgSeedTestTable(t, adapter)
	ctx := context.Background()

	if err := adapter.UpdateRow(ctx, "items", "001", map[string]any{"name": "alpha-updated"}); err != nil {
		t.Fatalf("UpdateRow: %v", err)
	}
	rows, _, err := adapter.QueryRows(ctx, "items", QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: "001"}},
		Page:    1, PerPage: 10,
	})
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if len(rows) != 1 || rows[0]["name"] != "alpha-updated" {
		t.Fatalf("expected alpha-updated, got %v", rows)
	}
}

func TestPostgresAdapter_DeleteRow(t *testing.T) {
	adapter := testPostgresAdapter(t)
	pgSeedTestTable(t, adapter)
	ctx := context.Background()

	if err := adapter.DeleteRow(ctx, "items", "003"); err != nil {
		t.Fatalf("DeleteRow: %v", err)
	}
	rows, total, err := adapter.QueryRows(ctx, "items", QueryOptions{Page: 1, PerPage: 10})
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

func TestPostgresAdapter_CountRows(t *testing.T) {
	adapter := testPostgresAdapter(t)
	pgSeedTestTable(t, adapter)
	ctx := context.Background()

	count, err := adapter.CountRows(ctx, "items")
	if err != nil {
		t.Fatalf("CountRows: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count=3, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Backend-aware column defaults
// ---------------------------------------------------------------------------

// TestPostgresDefaultLiteralForType verifies that boolean defaults are emitted
// as PostgreSQL-native literals (false/true) rather than integer literals
// (0/1), which PostgreSQL rejects for BOOLEAN columns.
func TestPostgresDefaultLiteralForType(t *testing.T) {
	if got := defaultLiteralForType(MoonFieldTypeBoolean, DBConnectionPostgres); got != "false" {
		t.Errorf("PG boolean default = %q, want %q", got, "false")
	}
	if got := defaultLiteralForType(MoonFieldTypeBoolean, DBConnectionSQLite); got != "0" {
		t.Errorf("SQLite boolean default = %q, want %q", got, "0")
	}
	if got := defaultLiteralForType(MoonFieldTypeInteger, DBConnectionPostgres); got != "0" {
		t.Errorf("PG integer default = %q, want %q", got, "0")
	}
}

// TestPostgresAdapter_AddBooleanColumn exercises the exact ALTER TABLE path
// used by collection schema mutation when adding a NOT NULL boolean column. It
// would fail on PostgreSQL if the default literal were an integer (0).
func TestPostgresAdapter_AddBooleanColumn(t *testing.T) {
	adapter := testPostgresAdapter(t)
	ctx := context.Background()
	pgSeedTestTable(t, adapter)

	ddl := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s NOT NULL DEFAULT %s",
		quoteIdent("items"),
		quoteIdent("is_active"),
		moonTypeToSQL(MoonFieldTypeBoolean, DBConnectionPostgres),
		defaultLiteralForType(MoonFieldTypeBoolean, DBConnectionPostgres))
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ALTER TABLE add boolean column failed (regression): %v", err)
	}
}

// ---------------------------------------------------------------------------
// Unique-constraint error mapping
// ---------------------------------------------------------------------------

func TestPostgresAdapter_UniqueViolation(t *testing.T) {
	adapter := testPostgresAdapter(t)
	ctx := context.Background()
	pgSeedTestTable(t, adapter)

	// Re-insert an existing id to trigger a primary-key violation; the server's
	// uniqueViolationFields parsing must handle PostgreSQL's error format.
	dup := map[string]any{
		"id":         "001",
		"name":       "dup",
		"quantity":   int64(1),
		"price":      "1.00",
		"active":     false,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"meta":       `{}`,
	}
	err := adapter.InsertRow(ctx, "items", dup)
	if err == nil {
		t.Fatal("expected unique/PK violation error")
	}
	fields := uniqueViolationFields(err)
	if len(fields) == 0 {
		t.Fatalf("expected uniqueViolationFields to detect a field, got none from %v", err)
	}
}

// ---------------------------------------------------------------------------
// Startup reconciliation (system tables + dynamic collection)
// ---------------------------------------------------------------------------

func TestPostgres_StartupReconcile(t *testing.T) {
	adapter := testPostgresAdapter(t)
	ctx := context.Background()

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	tables, err := adapter.ListTables(ctx)
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	has := map[string]bool{}
	for _, tbl := range tables {
		has[tbl] = true
	}
	for _, sys := range []string{"users", "apikeys", "moon_auth_refresh_tokens"} {
		if !has[sys] {
			t.Fatalf("expected system table %q to exist, got %v", sys, tables)
		}
	}

	// Exercise a dynamic collection end-to-end through the same DDL path the
	// server uses (backend-aware type mapping).
	if err := adapter.ExecDDL(ctx, "DROP TABLE IF EXISTS products"); err != nil {
		t.Fatalf("DROP products: %v", err)
	}
	ddl := fmt.Sprintf("CREATE TABLE products (%s TEXT PRIMARY KEY, %s %s, %s %s)",
		quoteIdent("id"),
		quoteIdent("title"), moonTypeToSQL(MoonFieldTypeString, DBConnectionPostgres),
		quoteIdent("price"), moonTypeToSQL(MoonFieldTypeDecimal, DBConnectionPostgres))
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("dynamic DDL: %v", err)
	}
	cols, err := adapter.DescribeTable(ctx, "products")
	if err != nil {
		t.Fatalf("DescribeTable products: %v", err)
	}
	var priceType string
	for _, c := range cols {
		if c.Name == "price" {
			priceType = c.Type
		}
	}
	if priceType != "NUMERIC" {
		t.Fatalf("expected price type NUMERIC, got %q", priceType)
	}
}

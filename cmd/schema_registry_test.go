package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testRegistryAdapter creates a temporary SQLite adapter with system tables.
func testRegistryAdapter(t *testing.T) *SQLiteAdapter {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "registry_test.db")
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

	ctx := context.Background()
	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}
	return adapter
}

// ---------------------------------------------------------------------------
// Population from system tables
// ---------------------------------------------------------------------------

func TestSchemaRegistry_PopulatesSystemTables(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	// users and apikeys must be present.
	users, ok := reg.Get("users")
	if !ok {
		t.Fatal("expected users collection in registry")
	}
	if users.Name != "users" {
		t.Fatalf("expected name=users, got %q", users.Name)
	}

	apikeys, ok := reg.Get("apikeys")
	if !ok {
		t.Fatal("expected apikeys collection in registry")
	}
	if apikeys.Name != "apikeys" {
		t.Fatalf("expected name=apikeys, got %q", apikeys.Name)
	}
}

func TestSchemaRegistry_MoonPrefixExcluded(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	// moon_auth_refresh_tokens must NOT be in the registry.
	if _, ok := reg.Get("moon_auth_refresh_tokens"); ok {
		t.Fatal("moon_auth_refresh_tokens must not be in registry")
	}
}

// ---------------------------------------------------------------------------
// Users collection field verification
// ---------------------------------------------------------------------------

func TestSchemaRegistry_UsersFields(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	users, ok := reg.Get("users")
	if !ok {
		t.Fatal("users not found")
	}

	// id must be first field.
	if len(users.Fields) == 0 {
		t.Fatal("users has no fields")
	}
	if users.Fields[0].Name != "id" {
		t.Fatalf("expected first field=id, got %q", users.Fields[0].Name)
	}
	if users.Fields[0].Type != MoonFieldTypeID {
		t.Fatalf("expected id type=%q, got %q", MoonFieldTypeID, users.Fields[0].Type)
	}
	if !users.Fields[0].ReadOnly {
		t.Fatal("id must be readonly")
	}

	// Check field types.
	fieldMap := make(map[string]Field)
	for _, f := range users.Fields {
		fieldMap[f.Name] = f
	}

	// password_hash must be present and readonly.
	ph, ok := fieldMap["password_hash"]
	if !ok {
		t.Fatal("password_hash not in users fields")
	}
	if !ph.ReadOnly {
		t.Fatal("password_hash must be readonly")
	}
	if ph.Type != MoonFieldTypeString {
		t.Fatalf("password_hash type=%q, want %q", ph.Type, MoonFieldTypeString)
	}

	// can_write must be boolean.
	cw, ok := fieldMap["can_write"]
	if !ok {
		t.Fatal("can_write not in users fields")
	}
	if cw.Type != MoonFieldTypeBoolean {
		t.Fatalf("can_write type=%q, want %q", cw.Type, MoonFieldTypeBoolean)
	}

	// username must be unique.
	un, ok := fieldMap["username"]
	if !ok {
		t.Fatal("username not in users fields")
	}
	if !un.Unique {
		t.Fatal("username must be unique")
	}

	// email must be unique.
	em, ok := fieldMap["email"]
	if !ok {
		t.Fatal("email not in users fields")
	}
	if !em.Unique {
		t.Fatal("email must be unique")
	}

	// created_at must be readonly for system collection.
	ca, ok := fieldMap["created_at"]
	if !ok {
		t.Fatal("created_at not in users fields")
	}
	if !ca.ReadOnly {
		t.Fatal("created_at must be readonly for system collection")
	}
}

// ---------------------------------------------------------------------------
// APIFields hides system fields
// ---------------------------------------------------------------------------

func TestSchemaRegistry_APIFieldsHidesPasswordHash(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	users, _ := reg.Get("users")
	apiFields := users.APIFields()

	for _, f := range apiFields {
		if f.Name == "password_hash" {
			t.Fatal("password_hash must not appear in API fields")
		}
	}

	// Ensure other fields are still present.
	found := false
	for _, f := range apiFields {
		if f.Name == "username" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("username must appear in API fields")
	}
}

func TestSchemaRegistry_APIFieldsHidesKeyHash(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	apikeys, _ := reg.Get("apikeys")
	apiFields := apikeys.APIFields()

	for _, f := range apiFields {
		if f.Name == "key_hash" {
			t.Fatal("key_hash must not appear in API fields")
		}
	}
}

func TestSchemaRegistry_APIFieldsNoHiddenForDynamic(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE products (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		price NUMERIC NOT NULL
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	p, ok := reg.Get("products")
	if !ok {
		t.Fatal("products not found")
	}
	apiFields := p.APIFields()
	if len(apiFields) != len(p.Fields) {
		t.Fatalf("dynamic collection should not hide any fields: api=%d, all=%d",
			len(apiFields), len(p.Fields))
	}
}

// ---------------------------------------------------------------------------
// Dynamic collection discovery
// ---------------------------------------------------------------------------

func TestSchemaRegistry_DiscoversDynamicTable(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE orders (
		id TEXT PRIMARY KEY,
		customer TEXT NOT NULL,
		total NUMERIC NOT NULL,
		shipped BOOLEAN NOT NULL DEFAULT 0
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	orders, ok := reg.Get("orders")
	if !ok {
		t.Fatal("orders not found in registry")
	}

	fieldMap := make(map[string]Field)
	for _, f := range orders.Fields {
		fieldMap[f.Name] = f
	}

	if fieldMap["id"].Type != MoonFieldTypeID {
		t.Fatalf("id type=%q, want %q", fieldMap["id"].Type, MoonFieldTypeID)
	}
	if fieldMap["customer"].Type != MoonFieldTypeString {
		t.Fatalf("customer type=%q, want %q", fieldMap["customer"].Type, MoonFieldTypeString)
	}
	if fieldMap["total"].Type != MoonFieldTypeDecimal {
		t.Fatalf("total type=%q, want %q", fieldMap["total"].Type, MoonFieldTypeDecimal)
	}
	if fieldMap["shipped"].Type != MoonFieldTypeBoolean {
		t.Fatalf("shipped type=%q, want %q", fieldMap["shipped"].Type, MoonFieldTypeBoolean)
	}

	// id is readonly for dynamic collections too.
	if !fieldMap["id"].ReadOnly {
		t.Fatal("id must be readonly in dynamic collections")
	}
	// Other fields are NOT readonly in dynamic collections.
	if fieldMap["customer"].ReadOnly {
		t.Fatal("customer should not be readonly in dynamic collection")
	}
}

// ---------------------------------------------------------------------------
// Type mapping
// ---------------------------------------------------------------------------

func TestPhysicalToMoonType(t *testing.T) {
	tests := []struct {
		name   string
		col    ColumnInfo
		want   string
		errMsg string
	}{
		{
			name: "TEXT PK id → id",
			col:  ColumnInfo{Name: "id", Type: "TEXT", PK: true},
			want: MoonFieldTypeID,
		},
		{
			name: "TEXT non-PK → string",
			col:  ColumnInfo{Name: "title", Type: "TEXT"},
			want: MoonFieldTypeString,
		},
		{
			name: "INTEGER → integer",
			col:  ColumnInfo{Name: "quantity", Type: "INTEGER"},
			want: MoonFieldTypeInteger,
		},
		{
			name: "BIGINT → integer",
			col:  ColumnInfo{Name: "count", Type: "BIGINT"},
			want: MoonFieldTypeInteger,
		},
		{
			name: "NUMERIC → decimal",
			col:  ColumnInfo{Name: "price", Type: "NUMERIC"},
			want: MoonFieldTypeDecimal,
		},
		{
			name: "NUMERIC(19,2) → decimal",
			col:  ColumnInfo{Name: "price", Type: "NUMERIC(19,2)"},
			want: MoonFieldTypeDecimal,
		},
		{
			name: "DECIMAL(19,2) → decimal",
			col:  ColumnInfo{Name: "price", Type: "DECIMAL(19,2)"},
			want: MoonFieldTypeDecimal,
		},
		{
			name: "BOOLEAN → boolean",
			col:  ColumnInfo{Name: "active", Type: "BOOLEAN"},
			want: MoonFieldTypeBoolean,
		},
		{
			name: "JSON → json",
			col:  ColumnInfo{Name: "metadata", Type: "JSON"},
			want: MoonFieldTypeJSON,
		},
		{
			name: "JSONB → json",
			col:  ColumnInfo{Name: "metadata", Type: "JSONB"},
			want: MoonFieldTypeJSON,
		},
		{
			name: "TIMESTAMP → datetime",
			col:  ColumnInfo{Name: "created_at", Type: "TIMESTAMP"},
			want: MoonFieldTypeDatetime,
		},
		{
			name: "TIMESTAMP WITH TIME ZONE → datetime",
			col:  ColumnInfo{Name: "created_at", Type: "TIMESTAMP WITH TIME ZONE"},
			want: MoonFieldTypeDatetime,
		},
		{
			name:   "REAL → unmappable",
			col:    ColumnInfo{Name: "weight", Type: "REAL"},
			errMsg: "unmappable SQL type",
		},
		{
			name:   "VARCHAR → unmappable",
			col:    ColumnInfo{Name: "name", Type: "VARCHAR"},
			errMsg: "unmappable SQL type",
		},
		{
			name: "non-id PK remains its type",
			col:  ColumnInfo{Name: "code", Type: "INTEGER", PK: true},
			want: MoonFieldTypeInteger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := physicalToMoonType(tt.col)
			if tt.errMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Fatalf("error %q does not contain %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unmappable column causes startup failure
// ---------------------------------------------------------------------------

func TestSchemaRegistry_UnmappableColumnFails(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE bad_table (
		id TEXT PRIMARY KEY,
		weight REAL NOT NULL
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	_, err := NewSchemaRegistry(adapter)
	if err == nil {
		t.Fatal("expected error for unmappable column, got nil")
	}
	if !strings.Contains(err.Error(), "bad_table") {
		t.Fatalf("error should mention table name: %v", err)
	}
	if !strings.Contains(err.Error(), "weight") {
		t.Fatalf("error should mention column name: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ID field always first
// ---------------------------------------------------------------------------

func TestEnsureIDFirst(t *testing.T) {
	fields := []Field{
		{Name: "name", Type: MoonFieldTypeString},
		{Name: "id", Type: MoonFieldTypeID},
		{Name: "price", Type: MoonFieldTypeDecimal},
	}
	result := ensureIDFirst(fields)
	if result[0].Name != "id" {
		t.Fatalf("expected first field=id, got %q", result[0].Name)
	}
	if result[1].Name != "name" {
		t.Fatalf("expected second field=name, got %q", result[1].Name)
	}
	if result[2].Name != "price" {
		t.Fatalf("expected third field=price, got %q", result[2].Name)
	}
}

func TestEnsureIDFirst_AlreadyFirst(t *testing.T) {
	fields := []Field{
		{Name: "id", Type: MoonFieldTypeID},
		{Name: "name", Type: MoonFieldTypeString},
	}
	result := ensureIDFirst(fields)
	if result[0].Name != "id" {
		t.Fatalf("expected first field=id, got %q", result[0].Name)
	}
}

func TestEnsureIDFirst_NoID(t *testing.T) {
	fields := []Field{
		{Name: "name", Type: MoonFieldTypeString},
		{Name: "price", Type: MoonFieldTypeDecimal},
	}
	result := ensureIDFirst(fields)
	if len(result) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// List returns sorted collections
// ---------------------------------------------------------------------------

func TestSchemaRegistry_ListAlphabetical(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	for _, name := range []string{"zeta_items", "alpha_items"} {
		ddl := "CREATE TABLE " + name + " (id TEXT PRIMARY KEY, val TEXT NOT NULL)"
		if err := adapter.ExecDDL(ctx, ddl); err != nil {
			t.Fatal(err)
		}
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	list := reg.List()
	if len(list) < 4 {
		t.Fatalf("expected at least 4 collections (users, apikeys, alpha, zeta), got %d", len(list))
	}

	// Verify sorted order.
	for i := 1; i < len(list); i++ {
		if list[i].Name < list[i-1].Name {
			t.Fatalf("list not sorted: %q comes after %q", list[i].Name, list[i-1].Name)
		}
	}
}

// ---------------------------------------------------------------------------
// Get returns nil for unknown collection
// ---------------------------------------------------------------------------

func TestSchemaRegistry_GetUnknown(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := reg.Get("nonexistent")
	if ok {
		t.Fatal("expected ok=false for nonexistent collection")
	}
}

// ---------------------------------------------------------------------------
// Refresh rebuilds registry
// ---------------------------------------------------------------------------

func TestSchemaRegistry_Refresh(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}

	// Initially no "tasks" collection.
	if _, ok := reg.Get("tasks"); ok {
		t.Fatal("tasks should not exist initially")
	}

	// Create a new table.
	ctx := context.Background()
	ddl := `CREATE TABLE tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT 0
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	// Refresh and verify.
	if err := reg.Refresh(); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	tasks, ok := reg.Get("tasks")
	if !ok {
		t.Fatal("tasks should exist after refresh")
	}
	if len(tasks.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(tasks.Fields))
	}
}

func TestSchemaRegistry_RefreshPreservesOnError(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE good_table (id TEXT PRIMARY KEY, name TEXT NOT NULL)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}

	// Verify good_table exists.
	if _, ok := reg.Get("good_table"); !ok {
		t.Fatal("good_table should exist")
	}

	// Create a table with unmappable column.
	ddl2 := `CREATE TABLE bad_refresh (id TEXT PRIMARY KEY, weight REAL NOT NULL)`
	if err := adapter.ExecDDL(ctx, ddl2); err != nil {
		t.Fatal(err)
	}

	// Refresh should fail.
	if err := reg.Refresh(); err == nil {
		t.Fatal("expected refresh to fail with unmappable column")
	}

	// Previous state should be preserved.
	if _, ok := reg.Get("good_table"); !ok {
		t.Fatal("good_table should still exist after failed refresh")
	}
	if _, ok := reg.Get("users"); !ok {
		t.Fatal("users should still exist after failed refresh")
	}
}

// ---------------------------------------------------------------------------
// Concurrency safety
// ---------------------------------------------------------------------------

func TestSchemaRegistry_ConcurrentReads(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reg.Get("users")
			reg.List()
		}()
	}
	wg.Wait()
}

func TestSchemaRegistry_ConcurrentReadAndRefresh(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	// Concurrent readers.
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				reg.Get("users")
				reg.List()
			}
		}()
	}
	// Concurrent refreshes.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = reg.Refresh()
		}()
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// Collection naming validation
// ---------------------------------------------------------------------------

func TestIsValidCollectionName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// Valid.
		{"ab", true},
		{"my_table", true},
		{"a1", true},
		{"products", true},
		{"order_items", true},
		{"a_very_long_name_that_is_still_within_limit_01234567890123456789", false}, // 64 chars, exceeds max 63

		// Too short.
		{"a", false},
		{"", false},

		// Invalid pattern.
		{"1abc", false},
		{"MyTable", false},
		{"my-table", false},
		{"my table", false},
		{"_hidden", false},

		// Reserved names.
		{"users", false},
		{"apikeys", false},
		{"collections", false},
		{"auth", false},
		{"doc", false},
		{"health", false},

		// moon_ prefix.
		{"moon_custom", false},
		{"moon_anything", false},

		// SQL keywords.
		{"select", false},
		{"table", false},
		{"insert", false},
		{"delete", false},
		{"update", false},
		{"create", false},
		{"drop", false},
		{"index", false},
		{"where", false},
		{"from", false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+boolStr(tt.want), func(t *testing.T) {
			got := IsValidCollectionName(tt.name)
			if got != tt.want {
				t.Fatalf("IsValidCollectionName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsValidCollectionName_MaxLen(t *testing.T) {
	// Exactly 63 characters.
	name := "a" + repeatStr("b", 62)
	if len(name) != 63 {
		t.Fatalf("test setup: expected 63 chars, got %d", len(name))
	}
	if !IsValidCollectionName(name) {
		t.Fatal("63-char name should be valid")
	}

	// 64 characters.
	name64 := name + "c"
	if IsValidCollectionName(name64) {
		t.Fatal("64-char name should be invalid")
	}
}

// ---------------------------------------------------------------------------
// Field naming validation
// ---------------------------------------------------------------------------

func TestIsValidFieldName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// Valid.
		{"abc", true},
		{"field_name", true},
		{"a12", true},
		{"title", true},
		{"created_at", true},

		// Too short.
		{"ab", false},
		{"a", false},
		{"", false},

		// Invalid pattern.
		{"1abc", false},
		{"MyField", false},
		{"my-field", false},

		// SQL keywords.
		{"select", false},
		{"table", false},
		{"where", false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+boolStr(tt.want), func(t *testing.T) {
			got := IsValidFieldName(tt.name)
			if got != tt.want {
				t.Fatalf("IsValidFieldName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsValidFieldName_MaxLen(t *testing.T) {
	name := "abc" + repeatStr("d", 60)
	if len(name) != 63 {
		t.Fatalf("test setup: expected 63 chars, got %d", len(name))
	}
	if !IsValidFieldName(name) {
		t.Fatal("63-char field name should be valid")
	}

	name64 := name + "e"
	if IsValidFieldName(name64) {
		t.Fatal("64-char field name should be invalid")
	}
}

// ---------------------------------------------------------------------------
// IsSQLKeyword
// ---------------------------------------------------------------------------

func TestIsSQLKeyword(t *testing.T) {
	if !IsSQLKeyword("SELECT") {
		t.Fatal("SELECT should be a keyword")
	}
	if !IsSQLKeyword("select") {
		t.Fatal("select should be a keyword")
	}
	if !IsSQLKeyword("Select") {
		t.Fatal("Select should be a keyword")
	}
	if IsSQLKeyword("products") {
		t.Fatal("products should not be a keyword")
	}
}

// ---------------------------------------------------------------------------
// matchesCollectionPattern
// ---------------------------------------------------------------------------

func TestMatchesCollectionPattern(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"ab", true},
		{"users", true},
		{"moon_custom", true}, // Pattern match only, no reserved check.
		{"a", false},          // Too short.
		{"1abc", false},       // Starts with digit.
		{"ABC", false},        // Uppercase.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesCollectionPattern(tt.name); got != tt.want {
				t.Fatalf("matchesCollectionPattern(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Invalid name tables are skipped during discovery
// ---------------------------------------------------------------------------

func TestSchemaRegistry_SkipsInvalidNameTables(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	// Create tables with names that don't match collection pattern.
	ddls := []string{
		`CREATE TABLE "X" (id TEXT PRIMARY KEY)`,                 // single char
		`CREATE TABLE "123abc" (id TEXT PRIMARY KEY)`,            // starts with digit
		`CREATE TABLE "CamelCase" (id TEXT PRIMARY KEY, a TEXT)`, // uppercase
	}
	for _, ddl := range ddls {
		if err := adapter.ExecDDL(ctx, ddl); err != nil {
			t.Fatal(err)
		}
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	for _, name := range []string{"X", "123abc", "CamelCase"} {
		if _, ok := reg.Get(name); ok {
			t.Fatalf("table %q should not be in registry", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Unique constraint detection
// ---------------------------------------------------------------------------

func TestSchemaRegistry_UniqueConstraintDetection(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE items (
		id TEXT PRIMARY KEY,
		sku TEXT NOT NULL,
		name TEXT NOT NULL,
		CONSTRAINT items_sku_unique UNIQUE (sku)
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}

	items, ok := reg.Get("items")
	if !ok {
		t.Fatal("items not found")
	}

	fieldMap := make(map[string]Field)
	for _, f := range items.Fields {
		fieldMap[f.Name] = f
	}

	// sku has a unique constraint.
	if !fieldMap["sku"].Unique {
		t.Fatal("sku should be marked unique")
	}
	// name does NOT have a unique constraint.
	if fieldMap["name"].Unique {
		t.Fatal("name should not be marked unique")
	}
	// id (PK) should NOT be marked unique (PK uniqueness is implicit).
	if fieldMap["id"].Unique {
		t.Fatal("id PK should not be marked unique in Unique field")
	}
}

func TestSchemaRegistry_UniqueIndexDetection(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE tags (
		id TEXT PRIMARY KEY,
		tag_name TEXT NOT NULL,
		category TEXT NOT NULL
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}
	idx := `CREATE UNIQUE INDEX idx_tags_name ON tags(tag_name)`
	if err := adapter.ExecDDL(ctx, idx); err != nil {
		t.Fatal(err)
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}

	tags, _ := reg.Get("tags")
	fieldMap := make(map[string]Field)
	for _, f := range tags.Fields {
		fieldMap[f.Name] = f
	}

	if !fieldMap["tag_name"].Unique {
		t.Fatal("tag_name should be unique via CREATE UNIQUE INDEX")
	}
	if fieldMap["category"].Unique {
		t.Fatal("category should not be unique")
	}
}

// ---------------------------------------------------------------------------
// Readonly field detection
// ---------------------------------------------------------------------------

func TestIsReadOnlyField(t *testing.T) {
	tests := []struct {
		table string
		field string
		isPK  bool
		want  bool
	}{
		// id is always readonly.
		{"users", "id", true, true},
		{"apikeys", "id", true, true},
		{"products", "id", true, true},

		// System readonly fields.
		{"users", "password_hash", false, true},
		{"users", "created_at", false, true},
		{"users", "updated_at", false, true},
		{"users", "last_login_at", false, true},
		{"apikeys", "key_hash", false, true},
		{"apikeys", "created_at", false, true},
		{"apikeys", "last_used_at", false, true},

		// Non-readonly system fields.
		{"users", "username", false, false},
		{"users", "email", false, false},
		{"users", "role", false, false},
		{"apikeys", "name", false, false},
		{"apikeys", "role", false, false},

		// Dynamic collection fields (not readonly).
		{"products", "title", false, false},
		{"products", "price", false, false},
	}

	for _, tt := range tests {
		name := tt.table + "." + tt.field
		t.Run(name, func(t *testing.T) {
			got := isReadOnlyField(tt.table, tt.field, tt.isPK)
			if got != tt.want {
				t.Fatalf("isReadOnlyField(%q, %q, %v) = %v, want %v",
					tt.table, tt.field, tt.isPK, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Apikeys collection field verification
// ---------------------------------------------------------------------------

func TestSchemaRegistry_ApikeysFields(t *testing.T) {
	adapter := testRegistryAdapter(t)
	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	apikeys, ok := reg.Get("apikeys")
	if !ok {
		t.Fatal("apikeys not found")
	}

	fieldMap := make(map[string]Field)
	for _, f := range apikeys.Fields {
		fieldMap[f.Name] = f
	}

	// key_hash must be readonly.
	kh, ok := fieldMap["key_hash"]
	if !ok {
		t.Fatal("key_hash not in apikeys fields")
	}
	if !kh.ReadOnly {
		t.Fatal("key_hash must be readonly")
	}

	// name must be unique.
	n, ok := fieldMap["name"]
	if !ok {
		t.Fatal("name not in apikeys fields")
	}
	if !n.Unique {
		t.Fatal("apikeys.name must be unique")
	}

	// id first.
	if apikeys.Fields[0].Name != "id" {
		t.Fatalf("expected first field=id, got %q", apikeys.Fields[0].Name)
	}
}

// ---------------------------------------------------------------------------
// Nullable detection
// ---------------------------------------------------------------------------

func TestSchemaRegistry_NullableDetection(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	ddl := `CREATE TABLE notes (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		body TEXT,
		rating INTEGER
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}

	notes, _ := reg.Get("notes")
	fieldMap := make(map[string]Field)
	for _, f := range notes.Fields {
		fieldMap[f.Name] = f
	}

	if fieldMap["title"].Nullable {
		t.Fatal("title should not be nullable")
	}
	if !fieldMap["body"].Nullable {
		t.Fatal("body should be nullable")
	}
	if !fieldMap["rating"].Nullable {
		t.Fatal("rating should be nullable")
	}
}

// ---------------------------------------------------------------------------
// All SQLite type constant types map correctly
// ---------------------------------------------------------------------------

func TestSchemaRegistry_SQLiteTypeConstants(t *testing.T) {
	adapter := testRegistryAdapter(t)
	ctx := context.Background()

	// Create a table using all SQLite type constants.
	ddl := `CREATE TABLE all_types (
		id TEXT PRIMARY KEY,
		str_field TEXT NOT NULL,
		int_field INTEGER NOT NULL,
		dec_field NUMERIC NOT NULL,
		bool_field BOOLEAN NOT NULL DEFAULT 0,
		json_field JSON,
		dt_field TIMESTAMP
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	reg, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	coll, ok := reg.Get("all_types")
	if !ok {
		t.Fatal("all_types not found")
	}

	fieldMap := make(map[string]Field)
	for _, f := range coll.Fields {
		fieldMap[f.Name] = f
	}

	expectations := map[string]string{
		"id":         MoonFieldTypeID,
		"str_field":  MoonFieldTypeString,
		"int_field":  MoonFieldTypeInteger,
		"dec_field":  MoonFieldTypeDecimal,
		"bool_field": MoonFieldTypeBoolean,
		"json_field": MoonFieldTypeJSON,
		"dt_field":   MoonFieldTypeDatetime,
	}
	for name, wantType := range expectations {
		f, ok := fieldMap[name]
		if !ok {
			t.Fatalf("field %q not found", name)
		}
		if f.Type != wantType {
			t.Fatalf("field %q: type=%q, want %q", name, f.Type, wantType)
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func boolStr(b bool) string {
	if b {
		return "valid"
	}
	return "invalid"
}

func repeatStr(s string, n int) string {
	return strings.Repeat(s, n)
}

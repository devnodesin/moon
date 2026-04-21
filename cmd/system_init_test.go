package main

import (
	"bytes"
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testAdapter creates a temporary SQLite adapter for system init tests.
func testAdapter(t *testing.T) DatabaseAdapter {
	t.Helper()
	return testSQLiteAdapter(t)
}

// testConfig returns an AppConfig with bootstrap admin fields populated.
func testConfig(t *testing.T) *AppConfig {
	t.Helper()
	return &AppConfig{
		BootstrapAdminUsername: "admin",
		BootstrapAdminEmail:    "admin@example.com",
		BootstrapAdminPassword: "SecurePass1",
	}
}

// testLogger returns a Logger backed by a bytes.Buffer for test inspection.
func testLogger(t *testing.T) (*Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	return NewTestLogger(&buf), &buf
}

// ---------------------------------------------------------------------------
// GenerateULID
// ---------------------------------------------------------------------------

func TestGenerateULID(t *testing.T) {
	id := GenerateULID()
	if len(id) != 26 {
		t.Fatalf("ULID length = %d; want 26", len(id))
	}
}

func TestGenerateULID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateULID()
		if seen[id] {
			t.Fatalf("duplicate ULID: %s", id)
		}
		seen[id] = true
	}
}

// ---------------------------------------------------------------------------
// HashPassword
// ---------------------------------------------------------------------------

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("SecurePass1")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("SecurePass1")); err != nil {
		t.Fatal("hash does not match original password")
	}
}

func TestHashPassword_Cost(t *testing.T) {
	hash, err := HashPassword("SecurePass1")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt.Cost: %v", err)
	}
	if cost != BcryptCost {
		t.Fatalf("cost = %d; want %d", cost, BcryptCost)
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	h1, _ := HashPassword("SecurePass1")
	h2, _ := HashPassword("SecurePass1")
	if h1 == h2 {
		t.Fatal("identical hashes for same password; expected unique salts")
	}
}

// ---------------------------------------------------------------------------
// EnsureSystemTables
// ---------------------------------------------------------------------------

func TestEnsureSystemTables_CreatesAllTables(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	tables, err := adapter.ListTables(ctx)
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}

	want := map[string]bool{
		"users":                    false,
		"apikeys":                  false,
		"moon_auth_refresh_tokens": false,
	}
	for _, tbl := range tables {
		if _, ok := want[tbl]; ok {
			want[tbl] = true
		}
	}
	for tbl, found := range want {
		if !found {
			t.Errorf("table %q not created", tbl)
		}
	}
}

func TestEnsureSystemTables_Idempotent(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("second call: %v", err)
	}
}

func TestEnsureSystemTables_UsersColumns(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	cols, err := adapter.DescribeTable(ctx, "users")
	if err != nil {
		t.Fatalf("DescribeTable: %v", err)
	}

	wantCols := []string{"id", "username", "email", "password_hash", "role",
		"can_write", "created_at", "updated_at", "last_login_at"}
	got := make(map[string]bool)
	for _, c := range cols {
		got[c.Name] = true
	}
	for _, name := range wantCols {
		if !got[name] {
			t.Errorf("users: missing column %q", name)
		}
	}
}

func TestEnsureSystemTables_ApikeysColumns(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	cols, err := adapter.DescribeTable(ctx, "apikeys")
	if err != nil {
		t.Fatalf("DescribeTable: %v", err)
	}

	wantCols := []string{"id", "name", "role", "can_write", "collections", "is_website",
		"allowed_origins", "rate_limit", "captcha_required", "enabled",
		"key_hash", "created_at", "updated_at", "last_used_at"}
	got := make(map[string]bool)
	for _, c := range cols {
		got[c.Name] = true
	}
	for _, name := range wantCols {
		if !got[name] {
			t.Errorf("apikeys: missing column %q", name)
		}
	}
}

func TestEnsureSystemTables_RefreshTokensColumns(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	cols, err := adapter.DescribeTable(ctx, "moon_auth_refresh_tokens")
	if err != nil {
		t.Fatalf("DescribeTable: %v", err)
	}

	wantCols := []string{"id", "user_id", "refresh_token_hash", "expires_at",
		"created_at", "last_used_at", "revoked_at", "revocation_reason"}
	got := make(map[string]bool)
	for _, c := range cols {
		got[c.Name] = true
	}
	for _, name := range wantCols {
		if !got[name] {
			t.Errorf("moon_auth_refresh_tokens: missing column %q", name)
		}
	}
}

// ---------------------------------------------------------------------------
// CreateBootstrapAdmin
// ---------------------------------------------------------------------------

func TestCreateBootstrapAdmin_CreatesAdmin(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	cfg := testConfig(t)
	logger, _ := testLogger(t)

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	if err := CreateBootstrapAdmin(ctx, adapter, cfg, logger); err != nil {
		t.Fatalf("CreateBootstrapAdmin: %v", err)
	}

	rows, total, err := adapter.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "role", Op: "eq", Value: "admin"}},
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if total != 1 {
		t.Fatalf("admin count = %d; want 1", total)
	}

	admin := rows[0]
	if admin["username"] != "admin" {
		t.Errorf("username = %v; want %q", admin["username"], "admin")
	}
	if admin["email"] != "admin@example.com" {
		t.Errorf("email = %v; want %q", admin["email"], "admin@example.com")
	}
	if admin["role"] != "admin" {
		t.Errorf("role = %v; want %q", admin["role"], "admin")
	}

	// can_write should be true (stored as 1, returned as bool or int64)
	switch v := admin["can_write"].(type) {
	case bool:
		if !v {
			t.Error("can_write = false; want true")
		}
	case int64:
		if v != 1 {
			t.Errorf("can_write = %d; want 1", v)
		}
	default:
		t.Fatalf("can_write type = %T; want bool or int64", admin["can_write"])
	}

	// password_hash must be a valid bcrypt hash
	hash, ok := admin["password_hash"].(string)
	if !ok || hash == "" {
		t.Fatal("password_hash is empty or not a string")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("SecurePass1")); err != nil {
		t.Error("password_hash does not match configured password")
	}

	// id must be a ULID (26 chars)
	id, ok := admin["id"].(string)
	if !ok || len(id) != 26 {
		t.Errorf("id = %q; want 26-character ULID", id)
	}
}

func TestCreateBootstrapAdmin_SkipsWhenNoBootstrapFields(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	cfg := &AppConfig{}
	logger, _ := testLogger(t)

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	if err := CreateBootstrapAdmin(ctx, adapter, cfg, logger); err != nil {
		t.Fatalf("CreateBootstrapAdmin: %v", err)
	}

	count, err := adapter.CountRows(ctx, "users")
	if err != nil {
		t.Fatalf("CountRows: %v", err)
	}
	if count != 0 {
		t.Fatalf("users count = %d; want 0", count)
	}
}

func TestCreateBootstrapAdmin_SkipsWhenAdminExists(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	cfg := testConfig(t)
	logger, _ := testLogger(t)

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	// Create first admin
	if err := CreateBootstrapAdmin(ctx, adapter, cfg, logger); err != nil {
		t.Fatalf("first CreateBootstrapAdmin: %v", err)
	}

	// Second call with different config should be a no-op
	cfg2 := &AppConfig{
		BootstrapAdminUsername: "admin2",
		BootstrapAdminEmail:    "admin2@example.com",
		BootstrapAdminPassword: "SecurePass2",
	}
	if err := CreateBootstrapAdmin(ctx, adapter, cfg2, logger); err != nil {
		t.Fatalf("second CreateBootstrapAdmin: %v", err)
	}

	count, err := adapter.CountRows(ctx, "users")
	if err != nil {
		t.Fatalf("CountRows: %v", err)
	}
	if count != 1 {
		t.Fatalf("users count = %d; want 1 (second call should skip)", count)
	}
}

func TestCreateBootstrapAdmin_LogsWarningWhenFieldsConfigured(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	cfg := testConfig(t)
	logger, buf := testLogger(t)

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	if err := CreateBootstrapAdmin(ctx, adapter, cfg, logger); err != nil {
		t.Fatalf("CreateBootstrapAdmin: %v", err)
	}

	logOutput := buf.String()
	if !bytes.Contains([]byte(logOutput), []byte("bootstrap admin fields")) {
		t.Error("expected log warning about bootstrap admin fields remaining in config")
	}
}

func TestCreateBootstrapAdmin_Idempotent(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	cfg := testConfig(t)
	logger, _ := testLogger(t)

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	// Run twice with same config
	for i := 0; i < 2; i++ {
		if err := CreateBootstrapAdmin(ctx, adapter, cfg, logger); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
	}

	count, err := adapter.CountRows(ctx, "users")
	if err != nil {
		t.Fatalf("CountRows: %v", err)
	}
	if count != 1 {
		t.Fatalf("users count = %d; want 1", count)
	}
}

func TestCreateBootstrapAdmin_TimestampsSet(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	cfg := testConfig(t)
	logger, _ := testLogger(t)

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}
	if err := CreateBootstrapAdmin(ctx, adapter, cfg, logger); err != nil {
		t.Fatalf("CreateBootstrapAdmin: %v", err)
	}

	rows, _, err := adapter.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "role", Op: "eq", Value: "admin"}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		t.Fatalf("QueryRows: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("no admin user found")
	}

	admin := rows[0]
	for _, field := range []string{"created_at", "updated_at"} {
		v, ok := admin[field].(string)
		if !ok || v == "" {
			t.Errorf("%s is empty or not a string", field)
		}
	}
}

func TestCreateBootstrapAdmin_SkipsWhenNonAdminUserExists(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	cfg := testConfig(t)
	logger, _ := testLogger(t)

	if err := EnsureSystemTables(ctx, adapter); err != nil {
		t.Fatalf("EnsureSystemTables: %v", err)
	}

	// Insert a non-admin user first
	hash, _ := HashPassword("UserPass1")
	err := adapter.InsertRow(ctx, "users", map[string]any{
		"id":            GenerateULID(),
		"username":      "regularuser",
		"email":         "user@example.com",
		"password_hash": hash,
		"role":          "user",
		"can_write":     int64(0),
		"created_at":    "2024-01-01T00:00:00Z",
		"updated_at":    "2024-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("InsertRow: %v", err)
	}

	// Bootstrap admin should still be created since no admin role exists
	if err := CreateBootstrapAdmin(ctx, adapter, cfg, logger); err != nil {
		t.Fatalf("CreateBootstrapAdmin: %v", err)
	}

	count, err := adapter.CountRows(ctx, "users")
	if err != nil {
		t.Fatalf("CountRows: %v", err)
	}
	if count != 2 {
		t.Fatalf("users count = %d; want 2 (regular + admin)", count)
	}
}

// ---------------------------------------------------------------------------
// Additional HashPassword and EnsureSystemTables tests
// ---------------------------------------------------------------------------

func TestHashPassword_Success(t *testing.T) {
	hash, err := HashPassword("ValidPass123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "ValidPass123" {
		t.Fatal("expected hash to differ from input")
	}
}

func TestHashPassword_Empty(t *testing.T) {
	// bcrypt allows empty passwords; this verifies the function works for them
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword empty: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash for empty password")
	}
}

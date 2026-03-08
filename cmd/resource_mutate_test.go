package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func setupMutateTest(t *testing.T) (*ResourceMutateHandler, *SQLiteAdapter, *SchemaRegistry) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mutate_test.db")
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

	// Create products table for dynamic collection tests
	productsDDL := `CREATE TABLE products (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		price NUMERIC NOT NULL DEFAULT 0,
		quantity INTEGER NOT NULL DEFAULT 0,
		active BOOLEAN NOT NULL DEFAULT 1,
		description TEXT,
		created_at TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT ''
	)`
	if err := adapter.ExecDDL(ctx, productsDDL); err != nil {
		t.Fatalf("ExecDDL products: %v", err)
	}

	registry, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	appCfg := &AppConfig{
		Server:    ServerConfig{Prefix: ""},
		JWTSecret: "test-secret-key-that-is-long-enough-for-jwt",
	}
	jtiStore := NewJTIRevocationStore()
	handler := NewResourceMutateHandler(adapter, registry, appCfg, jtiStore)
	return handler, adapter, registry
}

func doMutateRequest(t *testing.T, handler *ResourceMutateHandler, resource string, body any, identity *AuthIdentity) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/data/%s:mutate", resource), bytes.NewReader(b))
	if identity != nil {
		ctx := SetAuthIdentity(req.Context(), identity)
		req = req.WithContext(ctx)
	}
	w := httptest.NewRecorder()
	handler.HandleMutate(w, req)
	return w
}

func adminIdentity() *AuthIdentity {
	return &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "admin-id",
		Role:           "admin",
		CanWrite:       true,
		JTI:            "test-jti",
	}
}

func userWriteIdentity() *AuthIdentity {
	return &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user-id",
		Role:           "user",
		CanWrite:       true,
		JTI:            "test-jti",
	}
}

func userReadOnlyIdentity() *AuthIdentity {
	return &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user-id",
		Role:           "user",
		CanWrite:       false,
		JTI:            "test-jti",
	}
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v (body=%s)", err, w.Body.String())
	}
	return resp
}

func seedAdminUser(t *testing.T, adapter *SQLiteAdapter) string {
	t.Helper()
	hash, err := HashPassword("AdminPass123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	id := GenerateULID()
	err = adapter.InsertRow(context.Background(), "users", map[string]any{
		"id":            id,
		"username":      "admin",
		"email":         "admin@test.com",
		"password_hash": hash,
		"role":          "admin",
		"can_write":     int64(1),
		"created_at":    "2025-01-01T00:00:00Z",
		"updated_at":    "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	return id
}

// ---------------------------------------------------------------------------
// Tests: Authorization
// ---------------------------------------------------------------------------

func TestMutate_SystemResource_RequiresAdmin(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"username": "test", "email": "a@b.com", "password": "Pass1234", "role": "user"}},
	}

	w := doMutateRequest(t, handler, "users", body, userWriteIdentity())
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_DynamicResource_RequiresCanWrite(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"title": "Mouse", "price": "9.99", "quantity": 1}},
	}

	w := doMutateRequest(t, handler, "products", body, userReadOnlyIdentity())
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_DynamicResource_CanWriteAllowed(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"title": "Mouse", "price": "9.99", "quantity": 1, "active": true}},
	}

	w := doMutateRequest(t, handler, "products", body, userWriteIdentity())
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: op=create validation
// ---------------------------------------------------------------------------

func TestMutate_Create_RejectsIDInData(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"id": "abc", "title": "Mouse", "price": "9.99", "quantity": 1, "active": true}},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(t, w)
	if !strings.Contains(resp["message"].(string), "id") {
		t.Fatalf("expected error about id, got: %s", resp["message"])
	}
}

func TestMutate_Create_RejectsReadonlyFields(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"username": "u", "email": "a@b.com", "password": "Pass1234", "role": "user", "created_at": "2025-01-01T00:00:00Z"}},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_RejectsUnknownFields(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"title": "Mouse", "price": "9.99", "quantity": 1, "active": true, "nonexistent": "val"}},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_RejectsInvalidType(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"title": 123, "price": "9.99", "quantity": 1, "active": true}},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_MissingOp(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"data": []any{map[string]any{"title": "Mouse"}},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_MissingData(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "create",
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_EmptyData(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_UnknownOp(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "invalid",
		"data": []any{map[string]any{"title": "Mouse"}},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_NonexistentResource(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"title": "Mouse"}},
	}

	w := doMutateRequest(t, handler, "nonexistent", body, adminIdentity())
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: op=create dynamic resource
// ---------------------------------------------------------------------------

func TestMutate_Create_DynamicResource_Success(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{"title": "Mouse", "price": "29.99", "quantity": 10, "active": true},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}

	record := data[0].(map[string]any)
	if record["id"] == nil || record["id"] == "" {
		t.Fatal("expected id to be set")
	}
	if record["title"] != "Mouse" {
		t.Fatalf("expected title=Mouse, got %v", record["title"])
	}

	meta := resp["meta"].(map[string]any)
	if int(meta["success"].(float64)) != 1 {
		t.Fatalf("expected success=1, got %v", meta["success"])
	}
	if int(meta["failed"].(float64)) != 0 {
		t.Fatalf("expected failed=0, got %v", meta["failed"])
	}
}

func TestMutate_Create_DynamicResource_Batch(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{"title": "Item1", "price": "10.00", "quantity": 1, "active": true},
			map[string]any{"title": "Item2", "price": "20.00", "quantity": 2, "active": false},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 2 {
		t.Fatalf("expected 2 results, got %d", len(data))
	}

	meta := resp["meta"].(map[string]any)
	if int(meta["success"].(float64)) != 2 {
		t.Fatalf("expected success=2, got %v", meta["success"])
	}
}

// ---------------------------------------------------------------------------
// Tests: op=create users
// ---------------------------------------------------------------------------

func TestMutate_Create_User_Success(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{
				"username":  "newuser",
				"email":     "newuser@test.com",
				"password":  "SecurePass123",
				"role":      "user",
				"can_write": true,
			},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}

	record := data[0].(map[string]any)
	if record["id"] == nil || record["id"] == "" {
		t.Fatal("expected id to be set")
	}
	if record["username"] != "newuser" {
		t.Fatalf("expected username=newuser, got %v", record["username"])
	}
	// password_hash must not appear
	if _, ok := record["password_hash"]; ok {
		t.Fatal("password_hash must not appear in response")
	}
}

func TestMutate_Create_User_MissingFields(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"username": "u"}},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_User_WeakPassword(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{
				"username": "weakuser",
				"email":    "weak@test.com",
				"password": "abc",
				"role":     "user",
			},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_User_DuplicateUsername(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)
	seedAdminUser(t, adapter)

	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{
				"username": "admin",
				"email":    "other@test.com",
				"password": "SecurePass123",
				"role":     "user",
			},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: op=create apikeys
// ---------------------------------------------------------------------------

func TestMutate_Create_APIKey_Success(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{
				"name":      "my-service",
				"role":      "user",
				"can_write": true,
			},
		},
	}

	w := doMutateRequest(t, handler, "apikeys", body, adminIdentity())
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)

	key, ok := record["key"].(string)
	if !ok || !strings.HasPrefix(key, APIKeyPrefix) {
		t.Fatalf("expected key with prefix %s, got %v", APIKeyPrefix, record["key"])
	}
	if len(key) != APIKeyTotalLen {
		t.Fatalf("expected key length %d, got %d", APIKeyTotalLen, len(key))
	}

	// key_hash must not appear
	if _, ok := record["key_hash"]; ok {
		t.Fatal("key_hash must not appear in response")
	}
}

func TestMutate_Create_APIKey_MissingName(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"role": "user"}},
	}

	w := doMutateRequest(t, handler, "apikeys", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: op=update
// ---------------------------------------------------------------------------

func TestMutate_Update_Success(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)

	// Seed a product
	id := GenerateULID()
	err := adapter.InsertRow(context.Background(), "products", map[string]any{
		"id":         id,
		"title":      "Old Title",
		"price":      "10.00",
		"quantity":   int64(5),
		"active":     int64(1),
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}

	body := map[string]any{
		"op": "update",
		"data": []any{
			map[string]any{"id": id, "title": "New Title", "quantity": 10},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}

	record := data[0].(map[string]any)
	if record["title"] != "New Title" {
		t.Fatalf("expected title=New Title, got %v", record["title"])
	}

	meta := resp["meta"].(map[string]any)
	if int(meta["success"].(float64)) != 1 {
		t.Fatalf("expected success=1, got %v", meta["success"])
	}
}

func TestMutate_Update_MissingID(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "update",
		"data": []any{
			map[string]any{"title": "New Title"},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Update_NonexistentRecord(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "update",
		"data": []any{
			map[string]any{"id": "nonexistent", "title": "Updated"},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	meta := resp["meta"].(map[string]any)
	if int(meta["failed"].(float64)) != 1 {
		t.Fatalf("expected failed=1, got %v", meta["failed"])
	}
	if int(meta["success"].(float64)) != 0 {
		t.Fatalf("expected success=0, got %v", meta["success"])
	}
}

func TestMutate_Update_RejectsReadonly(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "update",
		"data": []any{
			map[string]any{"id": "some-id", "created_at": "2025-01-01T00:00:00Z"},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: op=destroy
// ---------------------------------------------------------------------------

func TestMutate_Destroy_Success(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)

	id := GenerateULID()
	err := adapter.InsertRow(context.Background(), "products", map[string]any{
		"id":         id,
		"title":      "ToDelete",
		"price":      "10.00",
		"quantity":   int64(1),
		"active":     int64(1),
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}

	body := map[string]any{
		"op": "destroy",
		"data": []any{
			map[string]any{"id": id},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	if dataRaw, ok := resp["data"]; ok && dataRaw != nil {
		data := dataRaw.([]any)
		if len(data) != 0 {
			t.Fatalf("expected empty data, got %d", len(data))
		}
	}

	meta := resp["meta"].(map[string]any)
	if int(meta["success"].(float64)) != 1 {
		t.Fatalf("expected success=1, got %v", meta["success"])
	}
}

func TestMutate_Destroy_MissingID(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "destroy",
		"data": []any{map[string]any{}},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Destroy_NonexistentRecord(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "destroy",
		"data": []any{
			map[string]any{"id": "nonexistent"},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	meta := resp["meta"].(map[string]any)
	if int(meta["failed"].(float64)) != 1 {
		t.Fatalf("expected failed=1, got %v", meta["failed"])
	}
}

func TestMutate_Destroy_LastAdmin_Protected(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)
	adminID := seedAdminUser(t, adapter)

	body := map[string]any{
		"op": "destroy",
		"data": []any{
			map[string]any{"id": adminID},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	meta := resp["meta"].(map[string]any)
	if int(meta["failed"].(float64)) != 1 {
		t.Fatalf("expected failed=1 (last admin protection), got %v", meta["failed"])
	}
	if int(meta["success"].(float64)) != 0 {
		t.Fatalf("expected success=0, got %v", meta["success"])
	}
}

func TestMutate_Destroy_User_CascadeTokens(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)

	// Create two admins so we can delete one
	admin1 := seedAdminUser(t, adapter)
	hash2, _ := HashPassword("Admin2Pass123")
	admin2 := GenerateULID()
	err := adapter.InsertRow(context.Background(), "users", map[string]any{
		"id":            admin2,
		"username":      "admin2",
		"email":         "admin2@test.com",
		"password_hash": hash2,
		"role":          "admin",
		"can_write":     int64(1),
		"created_at":    "2025-01-01T00:00:00Z",
		"updated_at":    "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed admin2: %v", err)
	}

	// Add a refresh token for admin1
	err = adapter.InsertRow(context.Background(), "moon_auth_refresh_tokens", map[string]any{
		"id":                 GenerateULID(),
		"user_id":            admin1,
		"refresh_token_hash": "somehash",
		"expires_at":         "2099-01-01T00:00:00Z",
		"created_at":         "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed refresh token: %v", err)
	}

	body := map[string]any{
		"op":   "destroy",
		"data": []any{map[string]any{"id": admin1}},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify refresh tokens were deleted
	rows, _, err := adapter.QueryRows(context.Background(), "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "user_id", Op: "eq", Value: admin1}},
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("query tokens: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 tokens after cascade delete, got %d", len(rows))
	}
}

// ---------------------------------------------------------------------------
// Tests: op=action reset_password
// ---------------------------------------------------------------------------

func TestMutate_Action_ResetPassword_Success(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)
	userID := seedAdminUser(t, adapter)

	body := map[string]any{
		"op":     "action",
		"action": "reset_password",
		"data": []any{
			map[string]any{"id": userID, "password": "NewSecure123"},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}
	record := data[0].(map[string]any)
	if record["id"] != userID {
		t.Fatalf("expected id=%s, got %v", userID, record["id"])
	}
}

func TestMutate_Action_ResetPassword_MissingPassword(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)
	userID := seedAdminUser(t, adapter)

	body := map[string]any{
		"op":     "action",
		"action": "reset_password",
		"data": []any{
			map[string]any{"id": userID},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Action_ResetPassword_WeakPassword(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)
	userID := seedAdminUser(t, adapter)

	body := map[string]any{
		"op":     "action",
		"action": "reset_password",
		"data": []any{
			map[string]any{"id": userID, "password": "weak"},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Action_MissingAction(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "action",
		"data": []any{map[string]any{"id": "abc"}},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Action_UnsupportedAction(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":     "action",
		"action": "fly_to_moon",
		"data":   []any{map[string]any{"id": "abc"}},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: op=action revoke_sessions
// ---------------------------------------------------------------------------

func TestMutate_Action_RevokeSessions_Success(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)
	userID := seedAdminUser(t, adapter)

	// Add refresh token
	err := adapter.InsertRow(context.Background(), "moon_auth_refresh_tokens", map[string]any{
		"id":                 GenerateULID(),
		"user_id":            userID,
		"refresh_token_hash": "hash1",
		"expires_at":         "2099-01-01T00:00:00Z",
		"created_at":         "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed token: %v", err)
	}

	body := map[string]any{
		"op":     "action",
		"action": "revoke_sessions",
		"data":   []any{map[string]any{"id": userID}},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify token was revoked
	rows, _, err := adapter.QueryRows(context.Background(), "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "user_id", Op: "eq", Value: userID}},
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("query tokens: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 token, got %d", len(rows))
	}
	if rows[0]["revoked_at"] == nil {
		t.Fatal("expected token to be revoked")
	}
	reason, _ := rows[0]["revocation_reason"].(string)
	if reason != "admin_revoked" {
		t.Fatalf("expected reason=admin_revoked, got %s", reason)
	}
}

// ---------------------------------------------------------------------------
// Tests: op=action rotate (apikeys)
// ---------------------------------------------------------------------------

func TestMutate_Action_RotateAPIKey_Success(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)

	// Create an API key
	origRaw, origHash := GenerateAPIKey()
	_ = origRaw
	id := GenerateULID()
	err := adapter.InsertRow(context.Background(), "apikeys", map[string]any{
		"id":         id,
		"name":       "service-key",
		"role":       "user",
		"can_write":  int64(1),
		"key_hash":   origHash,
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed apikey: %v", err)
	}

	body := map[string]any{
		"op":     "action",
		"action": "rotate",
		"data":   []any{map[string]any{"id": id}},
	}

	w := doMutateRequest(t, handler, "apikeys", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}

	record := data[0].(map[string]any)
	key, ok := record["key"].(string)
	if !ok || !strings.HasPrefix(key, APIKeyPrefix) {
		t.Fatalf("expected key with prefix, got %v", record["key"])
	}
	if record["name"] != "service-key" {
		t.Fatalf("expected name=service-key, got %v", record["name"])
	}
	if record["role"] != "user" {
		t.Fatalf("expected role=user, got %v", record["role"])
	}
	if record["can_write"] != true {
		t.Fatalf("expected can_write=true, got %v", record["can_write"])
	}
}

func TestMutate_Action_RotateAPIKey_NotFound(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":     "action",
		"action": "rotate",
		"data":   []any{map[string]any{"id": "nonexistent"}},
	}

	w := doMutateRequest(t, handler, "apikeys", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	meta := resp["meta"].(map[string]any)
	if int(meta["failed"].(float64)) != 1 {
		t.Fatalf("expected failed=1, got %v", meta["failed"])
	}
}

// ---------------------------------------------------------------------------
// Tests: GenerateAPIKey
// ---------------------------------------------------------------------------

func TestGenerateAPIKey_Format(t *testing.T) {
	raw, hash := GenerateAPIKey()

	if !strings.HasPrefix(raw, APIKeyPrefix) {
		t.Fatalf("expected prefix %s, got %s", APIKeyPrefix, raw[:10])
	}
	if len(raw) != APIKeyTotalLen {
		t.Fatalf("expected length %d, got %d", APIKeyTotalLen, len(raw))
	}
	if len(hash) != 64 {
		t.Fatalf("expected 64 char hash, got %d", len(hash))
	}
}

func TestGenerateAPIKey_Unique(t *testing.T) {
	raw1, _ := GenerateAPIKey()
	raw2, _ := GenerateAPIKey()
	if raw1 == raw2 {
		t.Fatal("expected unique keys")
	}
}

// ---------------------------------------------------------------------------
// Tests: isTypeValid
// ---------------------------------------------------------------------------

func TestIsTypeValid(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		fieldType string
		want      bool
	}{
		{"string ok", "hello", MoonFieldTypeString, true},
		{"string bad", 123, MoonFieldTypeString, false},
		{"integer ok float", float64(42), MoonFieldTypeInteger, true},
		{"integer ok float negative", float64(-5), MoonFieldTypeInteger, true},
		{"integer bad float", float64(42.5), MoonFieldTypeInteger, false},
		{"integer bad string", "42", MoonFieldTypeInteger, false},
		{"decimal ok string", "29.99", MoonFieldTypeDecimal, true},
		{"decimal ok float", float64(29.99), MoonFieldTypeDecimal, true},
		{"decimal bad int", 42, MoonFieldTypeDecimal, false},
		{"boolean ok", true, MoonFieldTypeBoolean, true},
		{"boolean bad", "true", MoonFieldTypeBoolean, false},
		{"datetime ok", "2025-01-01T00:00:00Z", MoonFieldTypeDatetime, true},
		{"datetime bad", "2025-01-01", MoonFieldTypeDatetime, false},
		{"json ok map", map[string]any{"a": 1}, MoonFieldTypeJSON, true},
		{"json ok slice", []any{1, 2}, MoonFieldTypeJSON, true},
		{"json bad", "string", MoonFieldTypeJSON, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTypeValid(tt.value, tt.fieldType)
			if got != tt.want {
				t.Fatalf("isTypeValid(%v, %s) = %v, want %v", tt.value, tt.fieldType, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: No auth identity
// ---------------------------------------------------------------------------

func TestMutate_NoAuthIdentity(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op":   "create",
		"data": []any{map[string]any{"title": "Mouse"}},
	}

	w := doMutateRequest(t, handler, "products", body, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: UniqueViolation detection
// ---------------------------------------------------------------------------

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"UNIQUE constraint failed: users.username", true},
		{"unique constraint violation", true},
		{"duplicate key value violates unique constraint", true},
		{"some other error", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.msg)
			got := isUniqueViolation(err)
			if got != tt.want {
				t.Fatalf("isUniqueViolation(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestIsUniqueViolation_Nil(t *testing.T) {
	if isUniqueViolation(nil) {
		t.Fatal("expected false for nil error")
	}
}

// ---------------------------------------------------------------------------
// Tests: Update with unique violation counted as failed
// ---------------------------------------------------------------------------

func TestMutate_Update_UniqueViolation_CountedAsFailed(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)

	// Seed two users
	hash, _ := HashPassword("Pass1234aa")
	id1 := GenerateULID()
	id2 := GenerateULID()
	ctx := context.Background()
	adapter.InsertRow(ctx, "users", map[string]any{
		"id": id1, "username": "user1", "email": "user1@test.com",
		"password_hash": hash, "role": "user", "can_write": int64(0),
		"created_at": "2025-01-01T00:00:00Z", "updated_at": "2025-01-01T00:00:00Z",
	})
	adapter.InsertRow(ctx, "users", map[string]any{
		"id": id2, "username": "user2", "email": "user2@test.com",
		"password_hash": hash, "role": "user", "can_write": int64(0),
		"created_at": "2025-01-01T00:00:00Z", "updated_at": "2025-01-01T00:00:00Z",
	})

	// Try updating user2's username to user1's
	body := map[string]any{
		"op": "update",
		"data": []any{
			map[string]any{"id": id2, "username": "user1"},
		},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	meta := resp["meta"].(map[string]any)
	if int(meta["failed"].(float64)) != 1 {
		t.Fatalf("expected failed=1, got %v", meta["failed"])
	}
}

// ---------------------------------------------------------------------------
// Tests: Action reset_password revokes refresh tokens
// ---------------------------------------------------------------------------

func TestMutate_Action_ResetPassword_RevokesTokens(t *testing.T) {
	handler, adapter, _ := setupMutateTest(t)
	userID := seedAdminUser(t, adapter)

	// Add a non-revoked refresh token
	err := adapter.InsertRow(context.Background(), "moon_auth_refresh_tokens", map[string]any{
		"id":                 GenerateULID(),
		"user_id":            userID,
		"refresh_token_hash": "testhash",
		"expires_at":         "2099-01-01T00:00:00Z",
		"created_at":         "2025-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("seed token: %v", err)
	}

	body := map[string]any{
		"op":     "action",
		"action": "reset_password",
		"data":   []any{map[string]any{"id": userID, "password": "NewPass123xyz"}},
	}

	w := doMutateRequest(t, handler, "users", body, adminIdentity())
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Check token was revoked
	rows, _, err := adapter.QueryRows(context.Background(), "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "user_id", Op: "eq", Value: userID}},
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("query tokens: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 token, got %d", len(rows))
	}
	if rows[0]["revoked_at"] == nil {
		t.Fatal("expected token to be revoked")
	}
	reason, _ := rows[0]["revocation_reason"].(string)
	if reason != "password_reset" {
		t.Fatalf("expected reason=password_reset, got %s", reason)
	}
}

// ---------------------------------------------------------------------------
// Tests: Boolean type validation
// ---------------------------------------------------------------------------

func TestMutate_Create_BooleanField(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{"title": "Item", "price": "5.00", "quantity": 1, "active": true},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Create_BooleanField_Invalid(t *testing.T) {
	handler, _, _ := setupMutateTest(t)

	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{"title": "Item", "price": "5.00", "quantity": 1, "active": "yes"},
		},
	}

	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// validationError type
// ---------------------------------------------------------------------------

func TestValidationError_Error(t *testing.T) {
err := &validationError{msg: "test validation error"}
if got := err.Error(); got != "test validation error" {
t.Errorf("validationError.Error() = %q, want %q", got, "test validation error")
}
}

// ---------------------------------------------------------------------------
// prepareValueForDB
// ---------------------------------------------------------------------------

func TestPrepareValueForDB(t *testing.T) {
tests := []struct {
name      string
value     any
fieldType string
want      any
}{
{"nil value", nil, MoonFieldTypeString, nil},
{"bool true", true, MoonFieldTypeBoolean, int64(1)},
{"bool false", false, MoonFieldTypeBoolean, int64(0)},
{"json map", map[string]any{"a": 1}, MoonFieldTypeJSON, `{"a":1}`},
{"json array", []any{1, 2}, MoonFieldTypeJSON, `[1,2]`},
{"string passthrough", "hello", MoonFieldTypeString, "hello"},
{"integer passthrough", int64(42), MoonFieldTypeInteger, int64(42)},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := prepareValueForDB(tt.value, tt.fieldType)
// For JSON, check the string representation
if tt.fieldType == MoonFieldTypeJSON && tt.value != nil {
gotStr, ok := got.(string)
if !ok {
t.Fatalf("expected string for JSON, got %T", got)
}
wantStr, _ := tt.want.(string)
if gotStr != wantStr {
t.Errorf("prepareValueForDB JSON = %q, want %q", gotStr, wantStr)
}
return
}
if got != tt.want {
t.Errorf("prepareValueForDB(%v, %q) = %v (%T), want %v (%T)",
tt.value, tt.fieldType, got, got, tt.want, tt.want)
}
})
}
}

// ---------------------------------------------------------------------------
// isTypeValid - covered by existing TestIsTypeValid above
// ---------------------------------------------------------------------------


// ---------------------------------------------------------------------------
// Additional action tests for uncovered paths
// ---------------------------------------------------------------------------

func TestMutate_Action_ResetPassword_MissingID(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":     "action",
"action": "reset_password",
"data":   []any{map[string]any{"password": "ValidPass1"}},
}
w := doMutateRequest(t, handler, "users", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

func TestMutate_Action_ResetPassword_InvalidIDType(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":     "action",
"action": "reset_password",
"data":   []any{map[string]any{"id": 123, "password": "ValidPass1"}},
}
w := doMutateRequest(t, handler, "users", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

func TestMutate_Action_ResetPassword_InvalidPasswordType(t *testing.T) {
handler, adapter, _ := setupMutateTest(t)
userID := seedAdminUser(t, adapter)
body := map[string]any{
"op":     "action",
"action": "reset_password",
"data":   []any{map[string]any{"id": userID, "password": 12345}},
}
w := doMutateRequest(t, handler, "users", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

func TestMutate_Action_ResetPassword_UserNotFound(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":     "action",
"action": "reset_password",
"data":   []any{map[string]any{"id": "nonexistent-id", "password": "ValidPass123"}},
}
w := doMutateRequest(t, handler, "users", body, adminIdentity())
if w.Code != http.StatusOK {
t.Fatalf("expected 200 (with failed count), got %d: %s", w.Code, w.Body.String())
}
resp := parseResponse(t, w)
meta := resp["meta"].(map[string]any)
if meta["failed"].(float64) != 1 {
t.Errorf("expected failed=1, got %v", meta["failed"])
}
}

func TestMutate_Action_RevokeSessions_MissingID(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":     "action",
"action": "revoke_sessions",
"data":   []any{map[string]any{}},
}
w := doMutateRequest(t, handler, "users", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

func TestMutate_Action_RevokeSessions_InvalidIDType(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":     "action",
"action": "revoke_sessions",
"data":   []any{map[string]any{"id": 456}},
}
w := doMutateRequest(t, handler, "users", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

func TestMutate_Action_RevokeSessions_UserNotFound(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":     "action",
"action": "revoke_sessions",
"data":   []any{map[string]any{"id": "nonexistent-user"}},
}
w := doMutateRequest(t, handler, "users", body, adminIdentity())
if w.Code != http.StatusOK {
t.Fatalf("expected 200 (with failed), got %d: %s", w.Code, w.Body.String())
}
resp := parseResponse(t, w)
meta := resp["meta"].(map[string]any)
if meta["failed"].(float64) != 1 {
t.Errorf("expected failed=1, got %v", meta["failed"])
}
}

func TestMutate_Action_RotateAPIKey_MissingID(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":     "action",
"action": "rotate",
"data":   []any{map[string]any{}},
}
w := doMutateRequest(t, handler, "apikeys", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

func TestMutate_Action_RotateAPIKey_InvalidIDType(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":     "action",
"action": "rotate",
"data":   []any{map[string]any{"id": true}},
}
w := doMutateRequest(t, handler, "apikeys", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

// ---------------------------------------------------------------------------
// validateFieldTypes edge cases
// ---------------------------------------------------------------------------

func TestValidateFieldTypes_NullableField(t *testing.T) {
fieldMap := map[string]Field{
"title": {Name: "title", Type: MoonFieldTypeString, Nullable: true},
}
item := map[string]any{"title": nil}
if err := validateFieldTypes(item, fieldMap); err != nil {
t.Errorf("unexpected error for nullable nil: %v", err)
}
}

func TestValidateFieldTypes_NonNullableNilField(t *testing.T) {
fieldMap := map[string]Field{
"title": {Name: "title", Type: MoonFieldTypeString, Nullable: false},
}
item := map[string]any{"title": nil}
if err := validateFieldTypes(item, fieldMap); err == nil {
t.Error("expected error for non-nullable nil field")
}
}

func TestValidateFieldTypes_InvalidType(t *testing.T) {
fieldMap := map[string]Field{
"count": {Name: "count", Type: MoonFieldTypeInteger, Nullable: false},
}
item := map[string]any{"count": "not-an-int"}
if err := validateFieldTypes(item, fieldMap); err == nil {
t.Error("expected type error for string in integer field")
}
}

func TestValidateFieldTypes_UnknownField(t *testing.T) {
fieldMap := map[string]Field{
"title": {Name: "title", Type: MoonFieldTypeString},
}
// Unknown fields are ignored (no error)
item := map[string]any{"unknown_field": "value"}
if err := validateFieldTypes(item, fieldMap); err != nil {
t.Errorf("unexpected error for unknown field: %v", err)
}
}

// ---------------------------------------------------------------------------
// Additional destroy/update/handleCreate coverage tests
// ---------------------------------------------------------------------------

func TestMutate_Destroy_InvalidIDType(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":   "destroy",
"data": []any{map[string]any{"id": 12345}},
}
w := doMutateRequest(t, handler, "products", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

func TestMutate_Update_InvalidIDType(t *testing.T) {
handler, _, _ := setupMutateTest(t)
body := map[string]any{
"op":   "update",
"data": []any{map[string]any{"id": true, "title": "X"}},
}
w := doMutateRequest(t, handler, "products", body, adminIdentity())
if w.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
}
}

func TestMutate_Update_RecordNotFound(t *testing.T) {
	handler, _, _ := setupMutateTest(t)
	body := map[string]any{
		"op":   "update",
		"data": []any{map[string]any{"id": "nonexistent-id", "title": "Updated"}},
	}
	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	// handleUpdate returns 200 with failed=1 when record not found
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(t, w)
	meta := resp["meta"].(map[string]any)
	if meta["failed"].(float64) != 1 {
		t.Errorf("expected failed=1, got %v", meta["failed"])
	}
}

func TestMutate_Create_NullableField(t *testing.T) {
	handler, _, _ := setupMutateTest(t)
	body := map[string]any{
		"op": "create",
		"data": []any{
			map[string]any{"title": "Item", "price": "5.00", "quantity": 1, "description": nil},
		},
	}
	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMutate_Destroy_EmptyData(t *testing.T) {
	handler, _, _ := setupMutateTest(t)
	body := map[string]any{
		"op":   "destroy",
		"data": []any{},
	}
	w := doMutateRequest(t, handler, "products", body, adminIdentity())
	// Empty data is rejected at HandleMutate level
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

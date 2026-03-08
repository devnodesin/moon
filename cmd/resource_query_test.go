package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func setupResourceQueryTest(t *testing.T) (*ResourceQueryHandler, *SQLiteAdapter, *SchemaRegistry) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "rq_test.db")
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

	// Create products table
	ddl := `CREATE TABLE products (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		price NUMERIC NOT NULL DEFAULT 0,
		quantity INTEGER NOT NULL DEFAULT 0,
		active BOOLEAN NOT NULL DEFAULT 1,
		description TEXT,
		metadata JSON,
		created_at TIMESTAMP NOT NULL DEFAULT ''
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ExecDDL products: %v", err)
	}

	// Create users table (system collection)
	usersDDL := `CREATE TABLE users (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT '',
		last_login_at TEXT
	)`
	if err := adapter.ExecDDL(ctx, usersDDL); err != nil {
		t.Fatalf("ExecDDL users: %v", err)
	}

	// Create apikeys table (system collection)
	apikeysDDL := `CREATE TABLE apikeys (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		can_write INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT '',
		last_used_at TEXT
	)`
	if err := adapter.ExecDDL(ctx, apikeysDDL); err != nil {
		t.Fatalf("ExecDDL apikeys: %v", err)
	}

	registry, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	appCfg := &AppConfig{
		Server: ServerConfig{Prefix: ""},
	}
	handler := NewResourceQueryHandler(adapter, registry, appCfg)

	return handler, adapter, registry
}

func seedProducts(t *testing.T, adapter *SQLiteAdapter) {
	t.Helper()
	ctx := context.Background()
	products := []map[string]any{
		{"id": "01J0001", "title": "Widget", "price": 9.99, "quantity": int64(100), "active": int64(1), "description": "A nice widget", "metadata": `{"color":"red"}`, "created_at": "2024-01-01T00:00:00Z"},
		{"id": "01J0002", "title": "Gadget", "price": 19.99, "quantity": int64(50), "active": int64(1), "description": "A cool gadget", "metadata": `{"color":"blue"}`, "created_at": "2024-01-02T00:00:00Z"},
		{"id": "01J0003", "title": "Doohickey", "price": 5.50, "quantity": int64(200), "active": int64(0), "description": "A doohickey", "metadata": `{"color":"green"}`, "created_at": "2024-01-03T00:00:00Z"},
		{"id": "01J0004", "title": "Thingamajig", "price": 29.99, "quantity": int64(10), "active": int64(1), "description": nil, "metadata": nil, "created_at": "2024-01-04T00:00:00Z"},
		{"id": "01J0005", "title": "Whatchamacallit", "price": 15.00, "quantity": int64(75), "active": int64(1), "description": "Quite useful", "metadata": `{"size":"large"}`, "created_at": "2024-01-05T00:00:00Z"},
	}
	for _, p := range products {
		if err := adapter.InsertRow(ctx, "products", p); err != nil {
			t.Fatalf("InsertRow: %v", err)
		}
	}
}

func seedUsers(t *testing.T, adapter *SQLiteAdapter) {
	t.Helper()
	ctx := context.Background()
	if err := adapter.InsertRow(ctx, "users", map[string]any{
		"id":            "U001",
		"username":      "admin",
		"email":         "admin@test.com",
		"password_hash": "$2a$12$fakehash",
		"role":          "admin",
		"created_at":    "2024-01-01T00:00:00Z",
		"updated_at":    "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("InsertRow users: %v", err)
	}
}

func seedAPIKeys(t *testing.T, adapter *SQLiteAdapter) {
	t.Helper()
	ctx := context.Background()
	if err := adapter.InsertRow(ctx, "apikeys", map[string]any{
		"id":         "K001",
		"name":       "test-key",
		"key_hash":   "abc123hash",
		"role":       "user",
		"can_write":  int64(0),
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("InsertRow apikeys: %v", err)
	}
}

func makeQueryRequest(path string) *http.Request {
	return httptest.NewRequest(http.MethodGet, path, nil)
}

func decodeRQResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

// ---------------------------------------------------------------------------
// Tests: List mode
// ---------------------------------------------------------------------------

func TestResourceQuery_ListMode_BasicPagination(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	if resp["message"] != "Resources retrieved successfully" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}

	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("data is not an array: %T", resp["data"])
	}
	if len(data) != 5 {
		t.Fatalf("expected 5 records, got %d", len(data))
	}

	meta := resp["meta"].(map[string]any)
	if meta["total"].(float64) != 5 {
		t.Fatalf("expected total=5, got %v", meta["total"])
	}
	if meta["per_page"].(float64) != float64(DefaultPerPage) {
		t.Fatalf("expected per_page=%d, got %v", DefaultPerPage, meta["per_page"])
	}
	if meta["current_page"].(float64) != 1 {
		t.Fatalf("expected current_page=1, got %v", meta["current_page"])
	}

	if resp["links"] == nil {
		t.Fatal("expected links to be present")
	}
}

func TestResourceQuery_ListMode_CustomPagination(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?page=2&per_page=2")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 2 {
		t.Fatalf("expected 2 records, got %d", len(data))
	}

	meta := resp["meta"].(map[string]any)
	if meta["total"].(float64) != 5 {
		t.Fatalf("expected total=5, got %v", meta["total"])
	}
	if meta["current_page"].(float64) != 2 {
		t.Fatalf("expected current_page=2, got %v", meta["current_page"])
	}
	if meta["total_pages"].(float64) != 3 {
		t.Fatalf("expected total_pages=3, got %v", meta["total_pages"])
	}
}

func TestResourceQuery_ListMode_EmptyCollection(t *testing.T) {
	h, _, _ := setupResourceQueryTest(t)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	// data may be nil or empty array when no records exist
	if resp["data"] != nil {
		data := resp["data"].([]any)
		if len(data) != 0 {
			t.Fatalf("expected 0 records, got %d", len(data))
		}
	}

	meta := resp["meta"].(map[string]any)
	if meta["total"].(float64) != 0 {
		t.Fatalf("expected total=0, got %v", meta["total"])
	}
}

// ---------------------------------------------------------------------------
// Tests: Get-one mode
// ---------------------------------------------------------------------------

func TestResourceQuery_GetOne_Found(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id=01J0001")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	if resp["message"] != "Resource retrieved successfully" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}

	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 record, got %d", len(data))
	}

	record := data[0].(map[string]any)
	if record["id"] != "01J0001" {
		t.Fatalf("expected id=01J0001, got %v", record["id"])
	}
	if record["title"] != "Widget" {
		t.Fatalf("expected title=Widget, got %v", record["title"])
	}

	// meta and links should NOT be present in get-one
	if resp["meta"] != nil {
		t.Fatal("expected no meta in get-one response")
	}
	if resp["links"] != nil {
		t.Fatal("expected no links in get-one response")
	}
}

func TestResourceQuery_GetOne_NotFound(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id=NONEXISTENT")
	h.HandleQuery(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Resource not found
// ---------------------------------------------------------------------------

func TestResourceQuery_ResourceNotFound(t *testing.T) {
	h, _, _ := setupResourceQueryTest(t)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/nonexistent:query")
	h.HandleQuery(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Sort
// ---------------------------------------------------------------------------

func TestResourceQuery_Sort_Ascending(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?sort=title")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	first := data[0].(map[string]any)
	if first["title"] != "Doohickey" {
		t.Fatalf("expected first item to be Doohickey, got %v", first["title"])
	}
}

func TestResourceQuery_Sort_Descending(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?sort=-quantity")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	first := data[0].(map[string]any)
	// Highest quantity is 200 (Doohickey)
	if first["title"] != "Doohickey" {
		t.Fatalf("expected first item to be Doohickey (qty 200), got %v", first["title"])
	}
}

func TestResourceQuery_Sort_UnknownField(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?sort=nonexistent")
	h.HandleQuery(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Fields projection
// ---------------------------------------------------------------------------

func TestResourceQuery_Fields_Projection(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?fields=title,price")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)

	// id should always be included
	if _, ok := record["id"]; !ok {
		t.Fatal("expected id to always be included")
	}
	if _, ok := record["title"]; !ok {
		t.Fatal("expected title in projection")
	}
	if _, ok := record["price"]; !ok {
		t.Fatal("expected price in projection")
	}
	// quantity should NOT be present
	if _, ok := record["quantity"]; ok {
		t.Fatal("expected quantity to be excluded from projection")
	}
}

func TestResourceQuery_Fields_UnknownField(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?fields=nonexistent")
	h.HandleQuery(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Full-text search (q)
// ---------------------------------------------------------------------------

func TestResourceQuery_Search(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?q=widget")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result for 'widget' search, got %d", len(data))
	}
	record := data[0].(map[string]any)
	if record["title"] != "Widget" {
		t.Fatalf("expected Widget, got %v", record["title"])
	}
}

// ---------------------------------------------------------------------------
// Tests: Filters
// ---------------------------------------------------------------------------

func TestResourceQuery_Filter_Eq(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?title[eq]=Gadget")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}
	if data[0].(map[string]any)["title"] != "Gadget" {
		t.Fatalf("expected Gadget")
	}
}

func TestResourceQuery_Filter_Ne(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?title[ne]=Gadget")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 4 {
		t.Fatalf("expected 4 results (all except Gadget), got %d", len(data))
	}
}

func TestResourceQuery_Filter_Gt(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?quantity[gt]=50")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	// qty > 50: Widget(100), Doohickey(200), Whatchamacallit(75)
	if len(data) != 3 {
		t.Fatalf("expected 3 results (qty > 50: 100, 200, 75), got %d", len(data))
	}
}

func TestResourceQuery_Filter_Gte(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?quantity[gte]=50")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	// qty >= 50: Widget(100), Gadget(50), Doohickey(200), Whatchamacallit(75)
	if len(data) != 4 {
		t.Fatalf("expected 4 results (qty >= 50), got %d", len(data))
	}
}

func TestResourceQuery_Filter_Lt(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?quantity[lt]=50")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result (qty < 50: 10), got %d", len(data))
	}
}

func TestResourceQuery_Filter_Lte(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?quantity[lte]=50")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 2 {
		t.Fatalf("expected 2 results (qty <= 50: 50, 10), got %d", len(data))
	}
}

func TestResourceQuery_Filter_Like(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?title[like]=dget")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	// "Widget" and "Gadget" both contain "dget"
	if len(data) != 2 {
		t.Fatalf("expected 2 results (Widget, Gadget contain 'dget'), got %d", len(data))
	}
}

func TestResourceQuery_Filter_In(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id[in]=01J0001,01J0003")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 2 {
		t.Fatalf("expected 2 results, got %d", len(data))
	}
}

func TestResourceQuery_Filter_InvalidOperatorForType(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	// "like" is not valid for integer fields
	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?quantity[like]=10")
	h.HandleQuery(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestResourceQuery_Filter_UnknownField(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?nonexistent[eq]=value")
	h.HandleQuery(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestResourceQuery_Filter_UnknownOperator(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?title[contains]=test")
	h.HandleQuery(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Unknown query parameters
// ---------------------------------------------------------------------------

func TestResourceQuery_UnknownQueryParam(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?unknown=value")
	h.HandleQuery(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: System collection visibility
// ---------------------------------------------------------------------------

func TestResourceQuery_Users_HidesPasswordHash(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedUsers(t, adapter)

	// Override the path to point to users
	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/users:query")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(data))
	}

	record := data[0].(map[string]any)
	if _, ok := record["password_hash"]; ok {
		t.Fatal("password_hash should not be visible in API response")
	}
	if _, ok := record["username"]; !ok {
		t.Fatal("username should be visible")
	}
}

func TestResourceQuery_Users_GetOne_HidesPasswordHash(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedUsers(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/users:query?id=U001")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)
	if _, ok := record["password_hash"]; ok {
		t.Fatal("password_hash should not be visible")
	}
}

func TestResourceQuery_APIKeys_HidesKeyHash(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedAPIKeys(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/apikeys:query")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 apikey, got %d", len(data))
	}

	record := data[0].(map[string]any)
	if _, ok := record["key_hash"]; ok {
		t.Fatal("key_hash should not be visible in API response")
	}
	if _, ok := record["name"]; !ok {
		t.Fatal("name should be visible")
	}
}

// ---------------------------------------------------------------------------
// Tests: Type conversion
// ---------------------------------------------------------------------------

func TestResourceQuery_TypeConversion_Boolean(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id=01J0001")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)

	active, ok := record["active"]
	if !ok {
		t.Fatal("expected active field")
	}
	if active != true {
		t.Fatalf("expected active=true (boolean), got %v (%T)", active, active)
	}
}

func TestResourceQuery_TypeConversion_BooleanFalse(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id=01J0003")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)

	active := record["active"]
	if active != false {
		t.Fatalf("expected active=false (boolean), got %v (%T)", active, active)
	}
}

func TestResourceQuery_TypeConversion_Decimal(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id=01J0001")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)

	price, ok := record["price"]
	if !ok {
		t.Fatal("expected price field")
	}
	priceStr, ok := price.(string)
	if !ok {
		t.Fatalf("expected price as string, got %T: %v", price, price)
	}
	if priceStr != "9.99" {
		t.Fatalf("expected price=9.99, got %s", priceStr)
	}
}

func TestResourceQuery_TypeConversion_Integer(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id=01J0001")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)

	qty := record["quantity"]
	// JSON numbers are float64 in Go
	qtyFloat, ok := qty.(float64)
	if !ok {
		t.Fatalf("expected quantity as number, got %T: %v", qty, qty)
	}
	if qtyFloat != 100 {
		t.Fatalf("expected quantity=100, got %v", qtyFloat)
	}
}

func TestResourceQuery_TypeConversion_JSON(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id=01J0001")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)

	metadata := record["metadata"]
	metaMap, ok := metadata.(map[string]any)
	if !ok {
		t.Fatalf("expected metadata as object, got %T: %v", metadata, metadata)
	}
	if metaMap["color"] != "red" {
		t.Fatalf("expected metadata.color=red, got %v", metaMap["color"])
	}
}

func TestResourceQuery_TypeConversion_NullJSON(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?id=01J0004")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	record := data[0].(map[string]any)

	if record["metadata"] != nil {
		t.Fatalf("expected metadata=null for Thingamajig, got %v", record["metadata"])
	}
}

// ---------------------------------------------------------------------------
// Tests: Pagination links include query params
// ---------------------------------------------------------------------------

func TestResourceQuery_PaginationLinks_IncludeQueryParams(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?per_page=2&sort=-title")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := decodeRQResponse(t, w)
	links := resp["links"].(map[string]any)

	nextLink, ok := links["next"].(string)
	if !ok {
		t.Fatal("expected next link as string")
	}

	if !strings.Contains(nextLink, "sort=-title") {
		t.Fatalf("expected next link to include sort param, got %s", nextLink)
	}
}

// ---------------------------------------------------------------------------
// Tests: Multiple filters combined
// ---------------------------------------------------------------------------

func TestResourceQuery_MultipleFilters(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	// active=1 AND quantity >= 50
	r := makeQueryRequest("/data/products:query?active[eq]=1&quantity[gte]=50")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	// active=1: Widget(100), Gadget(50), Whatchamacallit(75)
	// qty >= 50: Widget(100), Gadget(50), Doohickey(200), Whatchamacallit(75)
	// Intersection: Widget(100), Gadget(50), Whatchamacallit(75)
	if len(data) != 3 {
		t.Fatalf("expected 3 results, got %d", len(data))
	}
}

// ---------------------------------------------------------------------------
// Tests: Router integration
// ---------------------------------------------------------------------------

func TestResourceQuery_RouterIntegration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "router_test.db")
	dbCfg := DatabaseConfig{
		Connection:         DBConnectionSQLite,
		Database:           dbPath,
		QueryTimeout:       5,
		SlowQueryThreshold: 500,
	}
	logBuf := &bytes.Buffer{}
	logger := NewTestLogger(logBuf)
	adapter, err := NewSQLiteAdapter(dbCfg, logger)
	if err != nil {
		t.Fatalf("NewSQLiteAdapter: %v", err)
	}
	t.Cleanup(func() { adapter.Close() })

	ctx := context.Background()
	ddl := `CREATE TABLE tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT 0
	)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ExecDDL: %v", err)
	}
	if err := adapter.InsertRow(ctx, "tasks", map[string]any{
		"id": "T001", "title": "Do stuff", "done": int64(0),
	}); err != nil {
		t.Fatalf("InsertRow: %v", err)
	}

	registry, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	appCfg := &AppConfig{
		Server: ServerConfig{Prefix: ""},
		CORS: CORSConfig{
			Enabled:        true,
			AllowedOrigins: []string{"*"},
		},
	}

	mux := NewRouter(appCfg.Server.Prefix, logger, adapter, appCfg, registry)
	handler := BuildHandler(mux, appCfg, logger)

	// Without auth middleware (no JWT secret), the request should succeed
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/data/tasks:query", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 without auth middleware, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 task, got %d", len(data))
	}

	record := data[0].(map[string]any)
	if record["title"] != "Do stuff" {
		t.Fatalf("expected title=Do stuff, got %v", record["title"])
	}
	// Verify boolean conversion through the full stack
	if record["done"] != false {
		t.Fatalf("expected done=false, got %v (%T)", record["done"], record["done"])
	}
}

// ---------------------------------------------------------------------------
// Tests: buildResourcePaginationLinks
// ---------------------------------------------------------------------------

func TestBuildResourcePaginationLinks(t *testing.T) {
	q := url.Values{}
	q.Set("sort", "-title")
	q.Set("per_page", "2")
	q.Set("page", "1")

	links := buildResourcePaginationLinks("/data/products:query", 1, 2, 3, q)

	first := links["first"].(string)
	if !strings.Contains(first, "page=1") {
		t.Fatalf("first link should have page=1: %s", first)
	}
	if !strings.Contains(first, "sort=-title") {
		t.Fatalf("first link should include sort param: %s", first)
	}

	last := links["last"].(string)
	if !strings.Contains(last, "page=3") {
		t.Fatalf("last link should have page=3: %s", last)
	}

	if links["prev"] != nil {
		t.Fatal("prev should be nil on page 1")
	}

	next := links["next"].(string)
	if !strings.Contains(next, "page=2") {
		t.Fatalf("next link should have page=2: %s", next)
	}
}

// ---------------------------------------------------------------------------
// Tests: convertToMoonType
// ---------------------------------------------------------------------------

func TestConvertToMoonType(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		fieldType string
		check     func(any) bool
	}{
		{"nil value", nil, MoonFieldTypeString, func(v any) bool { return v == nil }},
		{"bool true from int64", int64(1), MoonFieldTypeBoolean, func(v any) bool { return v == true }},
		{"bool false from int64", int64(0), MoonFieldTypeBoolean, func(v any) bool { return v == false }},
		{"integer from int64", int64(42), MoonFieldTypeInteger, func(v any) bool { return v == int64(42) }},
		{"decimal from float64", float64(3.14), MoonFieldTypeDecimal, func(v any) bool { return v == "3.14" }},
		{"decimal from int64", int64(10), MoonFieldTypeDecimal, func(v any) bool { return v == "10" }},
		{"json from string", `{"a":1}`, MoonFieldTypeJSON, func(v any) bool {
			m, ok := v.(map[string]any)
			return ok && m["a"] == float64(1)
		}},
		{"string", "hello", MoonFieldTypeString, func(v any) bool { return v == "hello" }},
		{"datetime", "2024-01-01T00:00:00Z", MoonFieldTypeDatetime, func(v any) bool { return v == "2024-01-01T00:00:00Z" }},
		{"id", "ABC123", MoonFieldTypeID, func(v any) bool { return v == "ABC123" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToMoonType(tt.value, tt.fieldType)
			if !tt.check(result) {
				t.Fatalf("convertToMoonType(%v, %s) = %v (%T)", tt.value, tt.fieldType, result, result)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: filterHiddenFields
// ---------------------------------------------------------------------------

func TestFilterHiddenFields(t *testing.T) {
	record := map[string]any{
		"id": "U001", "username": "admin", "password_hash": "secret",
	}
	result := filterHiddenFields("users", record)
	if _, ok := result["password_hash"]; ok {
		t.Fatal("password_hash should be filtered")
	}
	if result["username"] != "admin" {
		t.Fatal("username should remain")
	}
}

func TestFilterHiddenFields_NonSystem(t *testing.T) {
	record := map[string]any{"id": "P001", "title": "Test"}
	result := filterHiddenFields("products", record)
	if len(result) != 2 {
		t.Fatal("non-system collection should not filter fields")
	}
}

// ---------------------------------------------------------------------------
// Tests: parseFilterParams edge cases
// ---------------------------------------------------------------------------

func TestParseFilterParams_BooleanOnlyEqNe(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	// "gt" is not valid for boolean fields
	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?active[gt]=0")
	h.HandleQuery(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for gt on boolean field, got %d", w.Code)
	}
}

func TestParseFilterParams_InOperator_Boolean(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	// "in" is not valid for boolean
	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?active[in]=0,1")
	h.HandleQuery(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for in on boolean field, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: validateQueryParams
// ---------------------------------------------------------------------------

func TestValidateQueryParams_AllKnown(t *testing.T) {
	h, _, _ := setupResourceQueryTest(t)
	col := &Collection{
		Name: "test",
		Fields: []Field{
			{Name: "id", Type: MoonFieldTypeID},
			{Name: "title", Type: MoonFieldTypeString},
		},
	}

	q := url.Values{}
	q.Set("page", "1")
	q.Set("per_page", "10")
	q.Set("sort", "title")
	q.Set("q", "test")
	q.Set("fields", "title")
	q.Set("id", "123")
	q.Set("title[eq]", "hello")

	if err := h.validateQueryParams(q, col); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateQueryParams_Unknown(t *testing.T) {
	h, _, _ := setupResourceQueryTest(t)
	col := &Collection{
		Name:   "test",
		Fields: []Field{{Name: "id", Type: MoonFieldTypeID}},
	}

	q := url.Values{}
	q.Set("bogus", "value")

	err := h.validateQueryParams(q, col)
	if err == nil {
		t.Fatal("expected error for unknown param")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("error should mention 'bogus': %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: sort with multiple fields
// ---------------------------------------------------------------------------

func TestResourceQuery_Sort_MultipleFields(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?sort=active,-quantity")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	// active ASC: Doohickey(0) first, then the rest (1)
	// Within active=1, quantity DESC: Widget(100), Whatchamacallit(75), Gadget(50), Thingamajig(10)
	first := data[0].(map[string]any)
	if first["title"] != "Doohickey" {
		t.Fatalf("expected Doohickey first, got %v", first["title"])
	}
}

// ---------------------------------------------------------------------------
// Tests: Prefix integration
// ---------------------------------------------------------------------------

func TestResourceQuery_WithPrefix(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "prefix_test.db")
	dbCfg := DatabaseConfig{
		Connection:         DBConnectionSQLite,
		Database:           dbPath,
		QueryTimeout:       5,
		SlowQueryThreshold: 500,
	}
	logger := NewTestLogger(&bytes.Buffer{})
	adapter, err := NewSQLiteAdapter(dbCfg, logger)
	if err != nil {
		t.Fatalf("NewSQLiteAdapter: %v", err)
	}
	t.Cleanup(func() { adapter.Close() })

	ctx := context.Background()
	ddl := `CREATE TABLE notes (id TEXT PRIMARY KEY, body TEXT NOT NULL)`
	if err := adapter.ExecDDL(ctx, ddl); err != nil {
		t.Fatalf("ExecDDL: %v", err)
	}
	if err := adapter.InsertRow(ctx, "notes", map[string]any{
		"id": "N001", "body": "Hello world",
	}); err != nil {
		t.Fatalf("InsertRow: %v", err)
	}

	registry, err := NewSchemaRegistry(adapter)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	appCfg := &AppConfig{
		Server: ServerConfig{Prefix: "/api/v1"},
	}

	rqh := NewResourceQueryHandler(adapter, registry, appCfg)

	w := httptest.NewRecorder()
	r := makeQueryRequest("/api/v1/data/notes:query")
	rqh.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 note, got %d", len(data))
	}

	// Verify links include prefix
	links := resp["links"].(map[string]any)
	first := links["first"].(string)
	if !strings.Contains(first, "/api/v1/data/notes:query") {
		t.Fatalf("first link should include prefix path: %s", first)
	}
}

// ---------------------------------------------------------------------------
// Tests: Combined sort, filter, fields, search
// ---------------------------------------------------------------------------

func TestResourceQuery_CombinedQueryParams(t *testing.T) {
	h, adapter, _ := setupResourceQueryTest(t)
	seedProducts(t, adapter)

	// Search for products with "widget" in text fields, sort by price desc, project fields
	w := httptest.NewRecorder()
	r := makeQueryRequest("/data/products:query?q=widget&sort=-price&fields=title,price")
	h.HandleQuery(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeRQResponse(t, w)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}

	record := data[0].(map[string]any)
	if _, ok := record["id"]; !ok {
		t.Fatal("id should always be present")
	}
	if _, ok := record["title"]; !ok {
		t.Fatal("title should be present in projection")
	}
	if _, ok := record["quantity"]; ok {
		t.Fatal("quantity should NOT be present in projection")
	}
}

// ---------------------------------------------------------------------------
// Tests: Implicit behavior
// ---------------------------------------------------------------------------

// Suppress unused import warnings
var _ = fmt.Sprintf

// ---------------------------------------------------------------------------
// Type conversion functions coverage
// ---------------------------------------------------------------------------

func TestToInteger(t *testing.T) {
tests := []struct {
name  string
input any
want  any
}{
{"int64", int64(42), int64(42)},
{"float64", float64(10), int64(10)},
{"int", int(7), int64(7)},
{"string valid", "99", int64(99)},
{"string invalid", "abc", "abc"},
{"bytes valid", []byte("55"), int64(55)},
{"bytes invalid", []byte("xyz"), "xyz"},
{"nil", nil, nil},
{"bool", true, true},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := toInteger(tt.input)
if got != tt.want {
t.Errorf("toInteger(%v) = %v (%T), want %v (%T)", tt.input, got, got, tt.want, tt.want)
}
})
}
}

func TestToDecimalString(t *testing.T) {
tests := []struct {
name  string
input any
want  string
}{
{"float64", float64(3.14), "3.14"},
{"int64", int64(100), "100"},
{"int", int(5), "5"},
{"string passthrough", "2.718", "2.718"},
{"bytes", []byte("1.23"), "1.23"},
{"bool default fmt", true, "true"},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := toDecimalString(tt.input)
if got != tt.want {
t.Errorf("toDecimalString(%v) = %q, want %q", tt.input, got, tt.want)
}
})
}
}

func TestToJSONValue(t *testing.T) {
tests := []struct {
name  string
input any
check func(t *testing.T, got any)
}{
{
"string valid JSON object",
`{"key":"val"}`,
func(t *testing.T, got any) {
m, ok := got.(map[string]any)
if !ok {
t.Fatalf("expected map, got %T", got)
}
if m["key"] != "val" {
t.Errorf("expected val, got %v", m["key"])
}
},
},
{
"string invalid JSON passthrough",
"not json",
func(t *testing.T, got any) {
if got != "not json" {
t.Errorf("expected passthrough, got %v", got)
}
},
},
{
"bytes valid JSON array",
[]byte(`[1,2,3]`),
func(t *testing.T, got any) {
arr, ok := got.([]any)
if !ok {
t.Fatalf("expected array, got %T", got)
}
if len(arr) != 3 {
t.Errorf("expected 3 elements, got %d", len(arr))
}
},
},
{
"bytes invalid JSON passthrough",
[]byte("not json"),
func(t *testing.T, got any) {
if got != "not json" {
t.Errorf("expected passthrough string, got %v", got)
}
},
},
{
"int64 passthrough",
int64(42),
func(t *testing.T, got any) {
if got != int64(42) {
t.Errorf("expected int64(42), got %v", got)
}
},
},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := toJSONValue(tt.input)
tt.check(t, got)
})
}
}

func TestToString(t *testing.T) {
tests := []struct {
name  string
input any
want  string
}{
{"string passthrough", "hello", "hello"},
{"bytes", []byte("world"), "world"},
{"int default fmt", 42, "42"},
{"bool default fmt", true, "true"},
{"nil", nil, "<nil>"},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := toString(tt.input)
if got != tt.want {
t.Errorf("toString(%v) = %q, want %q", tt.input, got, tt.want)
}
})
}
}

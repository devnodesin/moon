package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thalib/moon/cmd/moon/internal/config"
	"github.com/thalib/moon/cmd/moon/internal/database"
	"github.com/thalib/moon/cmd/moon/internal/registry"
)

// TestDefaultValues_Integration tests the complete default value workflow
func TestDefaultValues_Integration(t *testing.T) {
	// Setup database
	dbConfig := database.Config{
		ConnectionString: "sqlite://:memory:",
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Minute * 5,
	}

	driver, err := database.NewDriver(dbConfig)
	if err != nil {
		t.Fatalf("Failed to create database driver: %v", err)
	}

	ctx := context.Background()
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer driver.Close()

	// Setup registry and handlers
	reg := registry.NewSchemaRegistry()
	collectionsHandler := NewCollectionsHandler(driver, reg)
	dataHandler := NewDataHandler(driver, reg, &config.AppConfig{
		Batch: config.BatchConfig{
			MaxSize:         100,
			MaxPayloadBytes: 2097152,
		},
	})

	// 1. Create collection with type default values for nullable fields (no custom defaults)
	createReq := CreateRequest{
		Name: "test_products",
		Columns: []registry.Column{
			{Name: "name", Type: registry.TypeString, Nullable: false},     // required, no default
			{Name: "status", Type: registry.TypeString, Nullable: true},    // nullable with type default ""
			{Name: "price", Type: registry.TypeInteger, Nullable: false},   // required, no default
			{Name: "stock", Type: registry.TypeInteger, Nullable: true},    // nullable with type default 0
			{Name: "discount", Type: registry.TypeDecimal, Nullable: true}, // nullable with type default "0.00"
			{Name: "featured", Type: registry.TypeBoolean, Nullable: true}, // nullable with type default false
			{Name: "verified", Type: registry.TypeBoolean, Nullable: true}, // nullable with type default false
			{Name: "metadata", Type: registry.TypeJSON, Nullable: true},    // nullable with type default "{}"
			{Name: "notes", Type: registry.TypeString, Nullable: true},     // nullable with type default ""
		},
	}
	createBody, _ := json.Marshal(map[string]any{"data": createReq})
	createHTTPReq := httptest.NewRequest(http.MethodPost, "/collections:create", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	collectionsHandler.Create(createW, createHTTPReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("Failed to create collection: %s", createW.Body.String())
	}

	// 2. Insert record with only required fields (nullable fields should use database defaults)
	insertReq := BatchCreateDataRequest{
		Data: json.RawMessage(`[{"name": "Test Product", "price": 99}]`),
	}
	insertBody, _ := json.Marshal(insertReq)
	insertHTTPReq := httptest.NewRequest(http.MethodPost, "/test_products:create", bytes.NewReader(insertBody))
	insertW := httptest.NewRecorder()
	dataHandler.Create(insertW, insertHTTPReq, "test_products")

	if insertW.Code != http.StatusCreated {
		t.Fatalf("Failed to insert record: %s", insertW.Body.String())
	}

	// 3. Verify the response includes only provided fields (defaults are not in response)
	var insertRespRaw map[string]any
	if err := json.NewDecoder(insertW.Body).Decode(&insertRespRaw); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	insertData := insertRespRaw["data"].([]any)
	if len(insertData) == 0 {
		t.Fatal("expected at least one record in insert response")
	}
	insertRespItem := insertData[0].(map[string]any)

	// Check provided fields are in response
	if name, exists := insertRespItem["name"]; !exists || name != "Test Product" {
		t.Errorf("name: expected 'Test Product', got %v", name)
	}
	if price, exists := insertRespItem["price"]; !exists || price != float64(99) {
		t.Errorf("price: expected 99, got %v", price)
	}

	// Omitted fields should not be in response (database defaults were applied)
	// To verify defaults, we need to query the record
	listReq := httptest.NewRequest(http.MethodGet, "/test_products:list", nil)
	listW := httptest.NewRecorder()
	dataHandler.List(listW, listReq, "test_products")

	if listW.Code != http.StatusOK {
		t.Fatalf("Failed to list records: %s", listW.Body.String())
	}

	var listResp DataListResponse
	if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	if len(listResp.Data) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(listResp.Data))
	}

	record := listResp.Data[0]

	// Verify database defaults were applied (all type defaults now)
	tests := []struct {
		field    string
		expected any
	}{
		{"name", "Test Product"}, // provided
		{"status", ""},           // type default (empty string)
		{"price", float64(99)},   // provided
		{"stock", float64(0)},    // type default
		{"discount", "0.00"},     // type default
		{"featured", false},      // type default (stored as 0 in SQLite)
		{"verified", false},      // type default
		{"metadata", "{}"},       // type default
		{"notes", ""},            // type default (empty string)
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := record[tt.field]
			if got != tt.expected {
				t.Errorf("%s: expected %v (type %T), got %v (type %T)",
					tt.field, tt.expected, tt.expected, got, got)
			}
		})
	}

	// 4. Verify we can override defaults by providing values
	insertReq2 := BatchCreateDataRequest{
		Data: json.RawMessage(`[{"name": "Product 2", "status": "active", "price": 150, "stock": 5, "discount": "5.50", "featured": true, "verified": false, "metadata": "{\"key\":\"value\"}", "notes": "Some notes"}]`),
	}
	insertBody2, _ := json.Marshal(insertReq2)
	insertHTTPReq2 := httptest.NewRequest(http.MethodPost, "/test_products:create", bytes.NewReader(insertBody2))
	insertW2 := httptest.NewRecorder()
	dataHandler.Create(insertW2, insertHTTPReq2, "test_products")

	if insertW2.Code != http.StatusCreated {
		t.Fatalf("Failed to insert second record: %s", insertW2.Body.String())
	}

	var insertResp2Raw map[string]any
	if err := json.NewDecoder(insertW2.Body).Decode(&insertResp2Raw); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify overrides worked
	insertResp2Data := insertResp2Raw["data"].([]any)
	if len(insertResp2Data) == 0 {
		t.Fatal("expected at least one record in second insert response")
	}
	insertResp2Item := insertResp2Data[0].(map[string]any)
	if insertResp2Item["status"] != "active" {
		t.Errorf("status should be overridden to 'active', got %v", insertResp2Item["status"])
	}
	if insertResp2Item["verified"] != false {
		t.Errorf("verified should be overridden to false, got %v", insertResp2Item["verified"])
	}
}

// TestDefaultValues_BatchCreate tests batch create with defaults
func TestDefaultValues_BatchCreate(t *testing.T) {
	// Setup database
	dbConfig := database.Config{
		ConnectionString: "sqlite://:memory:",
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Minute * 5,
	}

	driver, err := database.NewDriver(dbConfig)
	if err != nil {
		t.Fatalf("Failed to create database driver: %v", err)
	}

	ctx := context.Background()
	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer driver.Close()

	// Setup registry and handlers
	reg := registry.NewSchemaRegistry()
	collectionsHandler := NewCollectionsHandler(driver, reg)
	dataHandler := NewDataHandler(driver, reg, &config.AppConfig{
		Batch: config.BatchConfig{
			MaxSize:         100,
			MaxPayloadBytes: 2097152,
		},
	})

	// Create collection with nullable field that has type default (0)
	createReq := CreateRequest{
		Name: "batch_test",
		Columns: []registry.Column{
			{Name: "title", Type: registry.TypeString, Nullable: false},
			{Name: "count", Type: registry.TypeInteger, Nullable: true}, // type default is 0
		},
	}
	createBody, _ := json.Marshal(map[string]any{"data": createReq})
	createHTTPReq := httptest.NewRequest(http.MethodPost, "/collections:create", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	collectionsHandler.Create(createW, createHTTPReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("Failed to create collection: %s", createW.Body.String())
	}

	// Batch insert with some records missing "count" field (should use database default)
	batchReq := BatchCreateDataRequest{
		Data: json.RawMessage(`[
			{"title": "Item 1"},
			{"title": "Item 2", "count": 5},
			{"title": "Item 3"}
		]`),
	}
	batchBody, _ := json.Marshal(batchReq)
	batchHTTPReq := httptest.NewRequest(http.MethodPost, "/batch_test:create?atomic=true", bytes.NewReader(batchBody))
	batchW := httptest.NewRecorder()
	dataHandler.Create(batchW, batchHTTPReq, "batch_test")

	if batchW.Code != http.StatusCreated {
		t.Fatalf("Failed to batch insert: %s", batchW.Body.String())
	}

	var batchRespRaw map[string]any
	if err := json.NewDecoder(batchW.Body).Decode(&batchRespRaw); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	batchData := batchRespRaw["data"].([]any)
	if len(batchData) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(batchData))
	}

	// Verify only provided fields are in response
	// Item 1 and 3 should NOT have count in response (it was omitted, database default applied)
	item0 := batchData[0].(map[string]any)
	item1 := batchData[1].(map[string]any)
	item2 := batchData[2].(map[string]any)

	if _, exists := item0["count"]; exists {
		t.Errorf("Item 1: count should not be in response (was omitted)")
	}
	if count := item1["count"]; count != float64(5) {
		t.Errorf("Item 2: expected count 5, got %v", count)
	}
	if _, exists := item2["count"]; exists {
		t.Errorf("Item 3: count should not be in response (was omitted)")
	}

	// To verify defaults were actually applied, query the records
	listReq := httptest.NewRequest(http.MethodGet, "/batch_test:list", nil)
	listW := httptest.NewRecorder()
	dataHandler.List(listW, listReq, "batch_test")

	if listW.Code != http.StatusOK {
		t.Fatalf("Failed to list records: %s", listW.Body.String())
	}

	var listResp DataListResponse
	if err := json.NewDecoder(listW.Body).Decode(&listResp); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}

	if len(listResp.Data) != 3 {
		t.Fatalf("Expected 3 records in list, got %d", len(listResp.Data))
	}

	// Verify database defaults were applied by checking each record by title
	for _, record := range listResp.Data {
		title := record["title"].(string)
		count := record["count"].(float64)

		switch title {
		case "Item 1":
			if count != float64(0) {
				t.Errorf("Item 1: expected count 0 from database, got %v", count)
			}
		case "Item 2":
			if count != float64(5) {
				t.Errorf("Item 2: expected count 5, got %v", count)
			}
		case "Item 3":
			if count != float64(0) {
				t.Errorf("Item 3: expected count 0 from database, got %v", count)
			}
		default:
			t.Errorf("Unexpected title: %s", title)
		}
	}
}

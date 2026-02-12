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

	// 1. Create collection with various default values
	createReq := CreateRequest{
		Name: "test_products",
		Columns: []registry.Column{
			{Name: "name", Type: registry.TypeString, Nullable: false},                                       // global default: ""
			{Name: "status", Type: registry.TypeString, Nullable: false, DefaultValue: stringPtr("pending")}, // custom default
			{Name: "price", Type: registry.TypeInteger, Nullable: false},                                     // global default: 0
			{Name: "stock", Type: registry.TypeInteger, Nullable: false, DefaultValue: stringPtr("10")},      // custom default
			{Name: "discount", Type: registry.TypeDecimal, Nullable: false},                                  // global default: "0.00"
			{Name: "featured", Type: registry.TypeBoolean, Nullable: false},                                  // global default: false
			{Name: "verified", Type: registry.TypeBoolean, Nullable: false, DefaultValue: stringPtr("true")}, // custom default
			{Name: "metadata", Type: registry.TypeJSON, Nullable: false},                                     // global default: "{}"
			{Name: "notes", Type: registry.TypeString, Nullable: true},                                       // nullable: NULL
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHTTPReq := httptest.NewRequest(http.MethodPost, "/collections:create", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	collectionsHandler.Create(createW, createHTTPReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("Failed to create collection: %s", createW.Body.String())
	}

	// 2. Insert record with only some fields (others should get defaults)
	insertReq := CreateDataRequest{
		Data: map[string]any{
			"name":  "Test Product",
			"price": 99,
			// All other fields omitted - should get defaults
		},
	}
	insertBody, _ := json.Marshal(insertReq)
	insertHTTPReq := httptest.NewRequest(http.MethodPost, "/test_products:create", bytes.NewReader(insertBody))
	insertW := httptest.NewRecorder()
	dataHandler.Create(insertW, insertHTTPReq, "test_products")

	if insertW.Code != http.StatusCreated {
		t.Fatalf("Failed to insert record: %s", insertW.Body.String())
	}

	// 3. Verify the response includes all defaults
	var insertResp CreateDataResponse
	if err := json.NewDecoder(insertW.Body).Decode(&insertResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check all expected defaults
	tests := []struct {
		field    string
		expected any
	}{
		{"name", "Test Product"}, // provided
		{"status", "pending"},    // custom default
		{"price", float64(99)},   // provided (JSON decodes to float64)
		{"stock", float64(10)},   // custom default (JSON decodes to float64)
		{"discount", "0.00"},     // global default
		{"featured", false},      // global default
		{"verified", true},       // custom default
		{"metadata", "{}"},       // global default
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := insertResp.Data[tt.field]
			if got != tt.expected {
				t.Errorf("%s: expected %v (type %T), got %v (type %T)",
					tt.field, tt.expected, tt.expected, got, got)
			}
		})
	}

	// Verify nullable field is nil
	if notes, exists := insertResp.Data["notes"]; exists && notes != nil {
		t.Errorf("notes: expected nil for nullable field, got %v", notes)
	}

	// 4. Verify we can override defaults by providing values
	insertReq2 := CreateDataRequest{
		Data: map[string]any{
			"name":     "Product 2",
			"status":   "active", // override custom default
			"price":    150,
			"stock":    5, // override custom default
			"discount": "5.50",
			"featured": true,  // override global default
			"verified": false, // override custom default
			"metadata": `{"key":"value"}`,
			"notes":    "Some notes",
		},
	}
	insertBody2, _ := json.Marshal(insertReq2)
	insertHTTPReq2 := httptest.NewRequest(http.MethodPost, "/test_products:create", bytes.NewReader(insertBody2))
	insertW2 := httptest.NewRecorder()
	dataHandler.Create(insertW2, insertHTTPReq2, "test_products")

	if insertW2.Code != http.StatusCreated {
		t.Fatalf("Failed to insert second record: %s", insertW2.Body.String())
	}

	var insertResp2 CreateDataResponse
	if err := json.NewDecoder(insertW2.Body).Decode(&insertResp2); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify overrides worked
	if insertResp2.Data["status"] != "active" {
		t.Errorf("status should be overridden to 'active', got %v", insertResp2.Data["status"])
	}
	if insertResp2.Data["verified"] != false {
		t.Errorf("verified should be overridden to false, got %v", insertResp2.Data["verified"])
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

	// Create collection
	createReq := CreateRequest{
		Name: "batch_test",
		Columns: []registry.Column{
			{Name: "title", Type: registry.TypeString, Nullable: false},
			{Name: "count", Type: registry.TypeInteger, Nullable: false, DefaultValue: stringPtr("0")},
		},
	}
	createBody, _ := json.Marshal(createReq)
	createHTTPReq := httptest.NewRequest(http.MethodPost, "/collections:create", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	collectionsHandler.Create(createW, createHTTPReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("Failed to create collection: %s", createW.Body.String())
	}

	// Batch insert with some records missing "count" field
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

	var batchResp BatchCreateResponse
	if err := json.NewDecoder(batchW.Body).Decode(&batchResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(batchResp.Data) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(batchResp.Data))
	}

	// Verify defaults were applied
	if count := batchResp.Data[0]["count"]; count != float64(0) {
		t.Errorf("Item 1: expected count 0, got %v", count)
	}
	if count := batchResp.Data[1]["count"]; count != float64(5) {
		t.Errorf("Item 2: expected count 5, got %v", count)
	}
	if count := batchResp.Data[2]["count"]; count != float64(0) {
		t.Errorf("Item 3: expected count 0, got %v", count)
	}
}

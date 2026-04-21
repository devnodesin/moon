package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResourceSchemaHandler_HandleSchema(t *testing.T) {
	registry := &SchemaRegistry{
		collections: map[string]*Collection{
			"products": {
				Name: "products",
				Fields: []Field{
					{Name: "id", Type: MoonFieldTypeID, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "title", Type: MoonFieldTypeString, Nullable: false, Unique: true, ReadOnly: false},
					{Name: "price", Type: MoonFieldTypeDecimal, Nullable: false, Unique: false, ReadOnly: false},
				},
			},
			"users": {
				Name: "users",
				Fields: []Field{
					{Name: "id", Type: MoonFieldTypeID, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "username", Type: MoonFieldTypeString, Nullable: false, Unique: true, ReadOnly: false},
					{Name: "email", Type: MoonFieldTypeString, Nullable: false, Unique: true, ReadOnly: false},
					{Name: "password_hash", Type: MoonFieldTypeString, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "role", Type: MoonFieldTypeString, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "can_write", Type: MoonFieldTypeBoolean, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "created_at", Type: MoonFieldTypeDatetime, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "updated_at", Type: MoonFieldTypeDatetime, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "last_login_at", Type: MoonFieldTypeDatetime, Nullable: true, Unique: false, ReadOnly: true},
				},
			},
			"apikeys": {
				Name: "apikeys",
				Fields: []Field{
					{Name: "id", Type: MoonFieldTypeID, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "name", Type: MoonFieldTypeString, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "role", Type: MoonFieldTypeString, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "can_write", Type: MoonFieldTypeBoolean, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "is_website", Type: MoonFieldTypeBoolean, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "allowed_origins", Type: MoonFieldTypeJSON, Nullable: true, Unique: false, ReadOnly: false},
					{Name: "rate_limit", Type: MoonFieldTypeInteger, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "captcha_required", Type: MoonFieldTypeBoolean, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "enabled", Type: MoonFieldTypeBoolean, Nullable: false, Unique: false, ReadOnly: false},
					{Name: "key_hash", Type: MoonFieldTypeString, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "created_at", Type: MoonFieldTypeDatetime, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "updated_at", Type: MoonFieldTypeDatetime, Nullable: false, Unique: false, ReadOnly: true},
					{Name: "last_used_at", Type: MoonFieldTypeDatetime, Nullable: true, Unique: false, ReadOnly: true},
				},
			},
		},
	}

	h := NewResourceSchemaHandler(registry, "/api")

	t.Run("success_dynamic_collection", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data/products:schema", nil)
		w := httptest.NewRecorder()
		h.HandleSchema(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
		}

		var resp SuccessResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp.Message != "Schema retrieved successfully" {
			t.Fatalf("got message %q, want %q", resp.Message, "Schema retrieved successfully")
		}
		if len(resp.Data) != 1 {
			t.Fatalf("got %d data items, want 1", len(resp.Data))
		}

		raw, _ := json.Marshal(resp.Data[0])
		var schema schemaObject
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("unmarshal schema: %v", err)
		}
		if schema.Name != "products" {
			t.Fatalf("got name %q, want %q", schema.Name, "products")
		}
		if len(schema.Fields) != 3 {
			t.Fatalf("got %d fields, want 3", len(schema.Fields))
		}
		if schema.Fields[0].Name != "id" || schema.Fields[0].Type != "id" || !schema.Fields[0].ReadOnly {
			t.Fatalf("id field mismatch: %+v", schema.Fields[0])
		}
	})

	t.Run("not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data/nonexistent:schema", nil)
		w := httptest.NewRecorder()
		h.HandleSchema(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("users_excludes_password_hash", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data/users:schema", nil)
		w := httptest.NewRecorder()
		h.HandleSchema(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
		}

		var resp SuccessResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		raw, _ := json.Marshal(resp.Data[0])
		var schema schemaObject
		json.Unmarshal(raw, &schema)

		for _, f := range schema.Fields {
			if f.Name == "password_hash" {
				t.Fatal("password_hash should be hidden from users schema")
			}
		}
		if len(schema.Fields) != 8 {
			t.Fatalf("got %d fields for users, want 8", len(schema.Fields))
		}
	})

	t.Run("apikeys_excludes_key_hash", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data/apikeys:schema", nil)
		w := httptest.NewRecorder()
		h.HandleSchema(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
		}

		var resp SuccessResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		raw, _ := json.Marshal(resp.Data[0])
		var schema schemaObject
		json.Unmarshal(raw, &schema)

		for _, f := range schema.Fields {
			if f.Name == "key_hash" {
				t.Fatal("key_hash should be hidden from apikeys schema")
			}
		}
		if len(schema.Fields) != 12 {
			t.Fatalf("got %d fields for apikeys, want 12", len(schema.Fields))
		}
	})

	t.Run("missing_resource", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/data/:schema", nil)
		w := httptest.NewRecorder()
		h.HandleSchema(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

package schema

import (
	"testing"

	"github.com/thalib/moon/cmd/moon/internal/registry"
)

func TestFromCollection_FiltersInternalColumns(t *testing.T) {
	builder := NewBuilder()

	// Test case 1: Collection with id and ulid columns should filter them out
	t.Run("filters_id_and_ulid_columns", func(t *testing.T) {
		collection := &registry.Collection{
			Name: "products",
			Columns: []registry.Column{
				{Name: "id", Type: registry.TypeInteger, Nullable: true},
				{Name: "ulid", Type: registry.TypeString, Nullable: false},
				{Name: "name", Type: registry.TypeString, Nullable: false},
				{Name: "price", Type: registry.TypeInteger, Nullable: false},
			},
		}

		schema := builder.FromCollection(collection)

		// Verify that schema has exactly 3 fields (id from builder + name + price)
		if len(schema.Fields) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(schema.Fields))
		}

		// Verify the first field is 'id' (string, non-nullable) - the external identifier
		if schema.Fields[0].Name != "id" {
			t.Errorf("Expected first field to be 'id', got '%s'", schema.Fields[0].Name)
		}
		if schema.Fields[0].Type != "string" {
			t.Errorf("Expected id type to be 'string', got '%s'", schema.Fields[0].Type)
		}
		if schema.Fields[0].Nullable {
			t.Error("Expected id to be non-nullable")
		}

		// Verify there's no duplicate 'id' field
		idCount := 0
		for _, field := range schema.Fields {
			if field.Name == "id" {
				idCount++
			}
		}
		if idCount != 1 {
			t.Errorf("Expected exactly 1 'id' field, got %d", idCount)
		}

		// Verify there's no 'ulid' field in the output
		for _, field := range schema.Fields {
			if field.Name == "ulid" {
				t.Error("Schema should not expose internal 'ulid' column")
			}
		}

		// Verify user-defined fields are present
		fieldNames := make(map[string]bool)
		for _, field := range schema.Fields {
			fieldNames[field.Name] = true
		}

		if !fieldNames["name"] {
			t.Error("Expected 'name' field in schema")
		}
		if !fieldNames["price"] {
			t.Error("Expected 'price' field in schema")
		}
	})

	// Test case 2: Collection without id/ulid columns (normal case)
	t.Run("normal_collection_without_system_columns", func(t *testing.T) {
		collection := &registry.Collection{
			Name: "users",
			Columns: []registry.Column{
				{Name: "username", Type: registry.TypeString, Nullable: false},
				{Name: "email", Type: registry.TypeString, Nullable: false},
			},
		}

		schema := builder.FromCollection(collection)

		// Should have 3 fields: id + username + email
		if len(schema.Fields) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(schema.Fields))
		}

		// First field should be 'id'
		if schema.Fields[0].Name != "id" {
			t.Errorf("Expected first field to be 'id', got '%s'", schema.Fields[0].Name)
		}

		// Verify other fields
		fieldNames := make(map[string]bool)
		for _, field := range schema.Fields {
			fieldNames[field.Name] = true
		}

		expectedFields := []string{"id", "username", "email"}
		for _, expected := range expectedFields {
			if !fieldNames[expected] {
				t.Errorf("Expected field '%s' in schema", expected)
			}
		}
	})

	// Test case 3: Collection with only id column (edge case)
	t.Run("collection_with_only_id_column", func(t *testing.T) {
		collection := &registry.Collection{
			Name: "test",
			Columns: []registry.Column{
				{Name: "id", Type: registry.TypeInteger, Nullable: true},
			},
		}

		schema := builder.FromCollection(collection)

		// Should have only 1 field: the external 'id' from the builder
		if len(schema.Fields) != 1 {
			t.Errorf("Expected 1 field, got %d", len(schema.Fields))
		}

		if schema.Fields[0].Name != "id" || schema.Fields[0].Type != "string" {
			t.Errorf("Expected single 'id' field of type string")
		}
	})

	// Test case 4: Verify primary key is always 'id'
	t.Run("primary_key_is_always_id", func(t *testing.T) {
		collection := &registry.Collection{
			Name: "orders",
			Columns: []registry.Column{
				{Name: "total", Type: registry.TypeInteger, Nullable: false},
			},
		}

		schema := builder.FromCollection(collection)

		if schema.PrimaryKey != "id" {
			t.Errorf("Expected primary_key to be 'id', got '%s'", schema.PrimaryKey)
		}
	})
}

func TestFromCollection_PreservesFieldProperties(t *testing.T) {
	builder := NewBuilder()

	defaultValue := "default_value"
	collection := &registry.Collection{
		Name: "test",
		Columns: []registry.Column{
			{Name: "field1", Type: registry.TypeString, Nullable: true},
			{Name: "field2", Type: registry.TypeInteger, Nullable: false, DefaultValue: &defaultValue},
		},
	}

	schema := builder.FromCollection(collection)

	// Find field1 and verify properties
	var field1 *FieldSchema
	for i := range schema.Fields {
		if schema.Fields[i].Name == "field1" {
			field1 = &schema.Fields[i]
			break
		}
	}

	if field1 == nil {
		t.Fatal("field1 not found in schema")
	}

	if field1.Type != "string" {
		t.Errorf("Expected field1 type 'string', got '%s'", field1.Type)
	}
	if !field1.Nullable {
		t.Error("Expected field1 to be nullable")
	}

	// Find field2 and verify properties
	var field2 *FieldSchema
	for i := range schema.Fields {
		if schema.Fields[i].Name == "field2" {
			field2 = &schema.Fields[i]
			break
		}
	}

	if field2 == nil {
		t.Fatal("field2 not found in schema")
	}

	if field2.Type != "integer" {
		t.Errorf("Expected field2 type 'integer', got '%s'", field2.Type)
	}
	if field2.Nullable {
		t.Error("Expected field2 to be non-nullable")
	}
	if field2.Default == nil {
		t.Error("Expected field2 to have a default value")
	} else {
		if *field2.Default != defaultValue {
			t.Errorf("Expected field2 default value '%s', got '%v'", defaultValue, *field2.Default)
		}
	}
}

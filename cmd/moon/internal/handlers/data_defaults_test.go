package handlers

import (
	"testing"

	"github.com/thalib/moon/cmd/moon/internal/registry"
)

// TestGetDefaultValue tests the getDefaultValue function for all data types
func TestGetDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		column   registry.Column
		expected any
	}{
		// Test global defaults for required (non-nullable) fields
		{
			name: "string: global default",
			column: registry.Column{
				Name:     "name",
				Type:     registry.TypeString,
				Nullable: false,
			},
			expected: "",
		},
		{
			name: "integer: global default",
			column: registry.Column{
				Name:     "count",
				Type:     registry.TypeInteger,
				Nullable: false,
			},
			expected: int64(0),
		},
		{
			name: "decimal: global default",
			column: registry.Column{
				Name:     "price",
				Type:     registry.TypeDecimal,
				Nullable: false,
			},
			expected: "0.00",
		},
		{
			name: "boolean: global default",
			column: registry.Column{
				Name:     "active",
				Type:     registry.TypeBoolean,
				Nullable: false,
			},
			expected: false,
		},
		{
			name: "datetime: global default (null)",
			column: registry.Column{
				Name:     "created",
				Type:     registry.TypeDatetime,
				Nullable: false,
			},
			expected: nil,
		},
		{
			name: "json: global default",
			column: registry.Column{
				Name:     "metadata",
				Type:     registry.TypeJSON,
				Nullable: false,
			},
			expected: "{}",
		},

		// Test nullable fields without explicit default
		{
			name: "nullable string without default",
			column: registry.Column{
				Name:     "description",
				Type:     registry.TypeString,
				Nullable: true,
			},
			expected: nil,
		},
		{
			name: "nullable integer without default",
			column: registry.Column{
				Name:     "score",
				Type:     registry.TypeInteger,
				Nullable: true,
			},
			expected: nil,
		},

		// Test custom defaults
		{
			name: "string: custom default",
			column: registry.Column{
				Name:         "status",
				Type:         registry.TypeString,
				Nullable:     false,
				DefaultValue: stringPtr("pending"),
			},
			expected: "pending",
		},
		{
			name: "integer: custom default",
			column: registry.Column{
				Name:         "priority",
				Type:         registry.TypeInteger,
				Nullable:     false,
				DefaultValue: stringPtr("5"),
			},
			expected: int64(5),
		},
		{
			name: "decimal: custom default",
			column: registry.Column{
				Name:         "discount",
				Type:         registry.TypeDecimal,
				Nullable:     false,
				DefaultValue: stringPtr("10.50"),
			},
			expected: "10.50",
		},
		{
			name: "boolean: custom default true",
			column: registry.Column{
				Name:         "verified",
				Type:         registry.TypeBoolean,
				Nullable:     false,
				DefaultValue: stringPtr("true"),
			},
			expected: true,
		},
		{
			name: "boolean: custom default false",
			column: registry.Column{
				Name:         "deleted",
				Type:         registry.TypeBoolean,
				Nullable:     false,
				DefaultValue: stringPtr("false"),
			},
			expected: false,
		},
		{
			name: "datetime: custom default",
			column: registry.Column{
				Name:         "scheduled",
				Type:         registry.TypeDatetime,
				Nullable:     false,
				DefaultValue: stringPtr("2026-01-01T00:00:00Z"),
			},
			expected: "2026-01-01T00:00:00Z",
		},
		{
			name: "json: custom default array",
			column: registry.Column{
				Name:         "tags",
				Type:         registry.TypeJSON,
				Nullable:     false,
				DefaultValue: stringPtr("[]"),
			},
			expected: "[]",
		},

		// Test null keyword in default
		{
			name: "nullable with explicit null default",
			column: registry.Column{
				Name:         "notes",
				Type:         registry.TypeString,
				Nullable:     true,
				DefaultValue: stringPtr("null"),
			},
			expected: nil,
		},
		{
			name: "nullable with explicit NULL default",
			column: registry.Column{
				Name:         "comments",
				Type:         registry.TypeString,
				Nullable:     true,
				DefaultValue: stringPtr("NULL"),
			},
			expected: nil,
		},

		// Edge cases
		{
			name: "integer: zero default",
			column: registry.Column{
				Name:         "retries",
				Type:         registry.TypeInteger,
				Nullable:     false,
				DefaultValue: stringPtr("0"),
			},
			expected: int64(0),
		},
		{
			name: "integer: negative default",
			column: registry.Column{
				Name:         "offset",
				Type:         registry.TypeInteger,
				Nullable:     false,
				DefaultValue: stringPtr("-1"),
			},
			expected: int64(-1),
		},
		{
			name: "boolean: 1 as true",
			column: registry.Column{
				Name:         "enabled",
				Type:         registry.TypeBoolean,
				Nullable:     false,
				DefaultValue: stringPtr("1"),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultValue(tt.column)

			// Compare results
			if got != tt.expected {
				t.Errorf("getDefaultValue() = %v (type: %T), want %v (type: %T)",
					got, got, tt.expected, tt.expected)
			}
		})
	}
}

// TestGetDefaultValueInvalidParsing tests fallback behavior for invalid default values
func TestGetDefaultValueInvalidParsing(t *testing.T) {
	// Test integer with invalid default - should fallback to 0
	col := registry.Column{
		Name:         "count",
		Type:         registry.TypeInteger,
		Nullable:     false,
		DefaultValue: stringPtr("not-a-number"),
	}
	got := getDefaultValue(col)
	if got != int64(0) {
		t.Errorf("getDefaultValue() with invalid integer = %v, want 0", got)
	}
}

// stringPtr is a helper to create string pointers
func stringPtr(s string) *string {
	return &s
}

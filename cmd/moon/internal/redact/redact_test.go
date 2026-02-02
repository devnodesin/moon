package redact

import (
	"testing"
)

func TestIsSensitiveField(t *testing.T) {
	r := New()

	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"password lowercase", "password", true},
		{"password uppercase", "PASSWORD", true},
		{"password mixed case", "Password", true},
		{"token", "token", true},
		{"secret", "secret", true},
		{"api_key", "api_key", true},
		{"apikey", "apikey", true},
		{"authorization", "authorization", true},
		{"jwt", "jwt", true},
		{"refresh_token", "refresh_token", true},
		{"access_token", "access_token", true},
		{"non-sensitive name", "name", false},
		{"non-sensitive email", "email", false},
		{"non-sensitive id", "id", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.IsSensitiveField(tt.field)
			if result != tt.expected {
				t.Errorf("IsSensitiveField(%q) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestRedactMap(t *testing.T) {
	r := New()

	t.Run("simple map with sensitive field", func(t *testing.T) {
		input := map[string]any{
			"name":     "Alice",
			"password": "secret123",
		}
		result := r.RedactMap(input)

		if result["name"] != "Alice" {
			t.Errorf("name = %v, want Alice", result["name"])
		}
		if result["password"] != RedactedPlaceholder {
			t.Errorf("password = %v, want %s", result["password"], RedactedPlaceholder)
		}
	})

	t.Run("nested map with sensitive field", func(t *testing.T) {
		input := map[string]any{
			"user": map[string]any{
				"name":     "Alice",
				"password": "secret123",
			},
		}
		result := r.RedactMap(input)

		userMap := result["user"].(map[string]any)
		if userMap["name"] != "Alice" {
			t.Errorf("user.name = %v, want Alice", userMap["name"])
		}
		if userMap["password"] != RedactedPlaceholder {
			t.Errorf("user.password = %v, want %s", userMap["password"], RedactedPlaceholder)
		}
	})

	t.Run("array with sensitive fields", func(t *testing.T) {
		input := map[string]any{
			"users": []any{
				map[string]any{
					"name":  "Alice",
					"token": "abc123",
				},
				map[string]any{
					"name":  "Bob",
					"token": "xyz789",
				},
			},
		}
		result := r.RedactMap(input)

		users := result["users"].([]any)
		for _, u := range users {
			user := u.(map[string]any)
			if user["token"] != RedactedPlaceholder {
				t.Errorf("user.token = %v, want %s", user["token"], RedactedPlaceholder)
			}
		}
	})

	t.Run("nil map", func(t *testing.T) {
		result := r.RedactMap(nil)
		if result != nil {
			t.Errorf("RedactMap(nil) = %v, want nil", result)
		}
	})
}

func TestNewWithFields(t *testing.T) {
	r := NewWithFields([]string{"ssn", "credit_card"})

	tests := []struct {
		field    string
		expected bool
	}{
		{"ssn", true},
		{"credit_card", true},
		{"password", true}, // default field
		{"name", false},
	}

	for _, tt := range tests {
		result := r.IsSensitiveField(tt.field)
		if result != tt.expected {
			t.Errorf("IsSensitiveField(%q) = %v, want %v", tt.field, result, tt.expected)
		}
	}
}

func TestGlobalFunctions(t *testing.T) {
	t.Run("IsSensitive", func(t *testing.T) {
		if !IsSensitive("password") {
			t.Error("IsSensitive(password) = false, want true")
		}
		if IsSensitive("name") {
			t.Error("IsSensitive(name) = true, want false")
		}
	})

	t.Run("Map", func(t *testing.T) {
		input := map[string]any{"password": "secret"}
		result := Map(input)
		if result["password"] != RedactedPlaceholder {
			t.Errorf("Map() password = %v, want %s", result["password"], RedactedPlaceholder)
		}
	})

	t.Run("String", func(t *testing.T) {
		result := String("password", "secret")
		if result != RedactedPlaceholder {
			t.Errorf("String(password, secret) = %v, want %s", result, RedactedPlaceholder)
		}

		result = String("name", "Alice")
		if result != "Alice" {
			t.Errorf("String(name, Alice) = %v, want Alice", result)
		}
	})
}

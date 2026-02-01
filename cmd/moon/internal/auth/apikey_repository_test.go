package auth

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	rawKey, keyHash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}

	// Check prefix
	if !strings.HasPrefix(rawKey, APIKeyPrefix) {
		t.Errorf("GenerateAPIKey() key doesn't have prefix %q, got %q", APIKeyPrefix, rawKey[:len(APIKeyPrefix)])
	}

	// Check length
	expectedLen := len(APIKeyPrefix) + APIKeyLength
	if len(rawKey) != expectedLen {
		t.Errorf("GenerateAPIKey() key length = %d, want %d", len(rawKey), expectedLen)
	}

	// Check hash is not empty
	if keyHash == "" {
		t.Error("GenerateAPIKey() keyHash is empty")
	}

	// Check hash is different from raw key
	if keyHash == rawKey {
		t.Error("GenerateAPIKey() keyHash equals rawKey")
	}

	// Check that hashing the raw key produces the same hash
	if HashAPIKey(rawKey) != keyHash {
		t.Error("GenerateAPIKey() HashAPIKey(rawKey) != keyHash")
	}
}

func TestGenerateAPIKey_Uniqueness(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		rawKey, _, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("GenerateAPIKey() error = %v", err)
		}
		if keys[rawKey] {
			t.Errorf("GenerateAPIKey() generated duplicate key")
		}
		keys[rawKey] = true
	}
}

func TestHashAPIKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"simple key", "moon_live_abc123"},
		{"empty key", ""},
		{"long key", strings.Repeat("a", 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashAPIKey(tt.key)

			// Hash should be 64 characters (SHA-256 in hex)
			if len(hash) != 64 {
				t.Errorf("HashAPIKey() length = %d, want 64", len(hash))
			}

			// Same input should produce same output
			if HashAPIKey(tt.key) != hash {
				t.Error("HashAPIKey() not deterministic")
			}
		})
	}
}

func TestHashAPIKey_DifferentInputs(t *testing.T) {
	hash1 := HashAPIKey("key1")
	hash2 := HashAPIKey("key2")

	if hash1 == hash2 {
		t.Error("HashAPIKey() produced same hash for different inputs")
	}
}

func TestBase62Charset(t *testing.T) {
	if len(base62Charset) != 62 {
		t.Errorf("base62Charset length = %d, want 62", len(base62Charset))
	}

	// Check it contains expected characters
	for _, c := range "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" {
		if !strings.ContainsRune(base62Charset, c) {
			t.Errorf("base62Charset missing character %c", c)
		}
	}
}

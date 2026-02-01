package auth

import (
	"strings"
	"testing"
)

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"simple token", "abc123"},
		{"empty token", ""},
		{"long token", strings.Repeat("x", 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashToken(tt.token)

			// Hash should be 64 characters (SHA-256 in hex)
			if len(hash) != 64 {
				t.Errorf("HashToken() length = %d, want 64", len(hash))
			}

			// Same input should produce same output
			if HashToken(tt.token) != hash {
				t.Error("HashToken() not deterministic")
			}
		})
	}
}

func TestHashToken_DifferentInputs(t *testing.T) {
	hash1 := HashToken("token1")
	hash2 := HashToken("token2")

	if hash1 == hash2 {
		t.Error("HashToken() produced same hash for different inputs")
	}
}

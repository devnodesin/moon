package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid password",
			password: "securePassword123!",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "a",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: strings.Repeat("a", 72), // bcrypt max is 72 bytes
			wantErr:  false,
		},
		{
			name:        "empty password",
			password:    "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("HashPassword() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("HashPassword() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("HashPassword() unexpected error = %v", err)
				return
			}

			if hash == "" {
				t.Error("HashPassword() returned empty hash")
				return
			}

			// Hash should be different from password
			if hash == tt.password {
				t.Error("HashPassword() hash equals password")
			}

			// Hash should start with bcrypt identifier
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
				t.Errorf("HashPassword() hash doesn't look like bcrypt: %s", hash)
			}
		})
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "testPassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Same password should produce different hashes (due to salt)
	if hash1 == hash2 {
		t.Error("HashPassword() produced same hash for same password (should use random salt)")
	}
}

func TestComparePassword(t *testing.T) {
	password := "securePassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  bool
	}{
		{
			name:     "correct password",
			hash:     hash,
			password: password,
			wantErr:  false,
		},
		{
			name:     "wrong password",
			hash:     hash,
			password: "wrongPassword",
			wantErr:  true,
		},
		{
			name:     "empty password",
			hash:     hash,
			password: "",
			wantErr:  true,
		},
		{
			name:     "empty hash",
			hash:     "",
			password: password,
			wantErr:  true,
		},
		{
			name:     "invalid hash",
			hash:     "not-a-valid-hash",
			password: password,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePassword(tt.hash, tt.password)

			if tt.wantErr && err == nil {
				t.Errorf("ComparePassword() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ComparePassword() unexpected error = %v", err)
			}
		})
	}
}

func TestBcryptCost(t *testing.T) {
	if BcryptCost != 12 {
		t.Errorf("BcryptCost = %d, want 12", BcryptCost)
	}
}

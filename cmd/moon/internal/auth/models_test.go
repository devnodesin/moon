package auth

import (
	"testing"
	"time"
)

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{"admin", true},
		{"user", true},
		{"readonly", true},
		{"", false},
		{"superadmin", false},
		{"ADMIN", false},
		{"Admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			if got := IsValidRole(tt.role); got != tt.valid {
				t.Errorf("IsValidRole(%q) = %v, want %v", tt.role, got, tt.valid)
			}
		})
	}
}

func TestValidRoles(t *testing.T) {
	roles := ValidRoles()
	if len(roles) != 3 {
		t.Errorf("ValidRoles() returned %d roles, want 3", len(roles))
	}

	expected := map[UserRole]bool{
		RoleAdmin:    true,
		RoleUser:     true,
		RoleReadOnly: true,
	}

	for _, role := range roles {
		if !expected[role] {
			t.Errorf("unexpected role %q in ValidRoles()", role)
		}
	}
}

func TestRefreshToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired",
			expiresAt: time.Now().Add(time.Hour),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().Add(-time.Hour),
			want:      true,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RefreshToken{ExpiresAt: tt.expiresAt}
			if got := r.IsExpired(); got != tt.want {
				t.Errorf("RefreshToken.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIKeyPrefix(t *testing.T) {
	if APIKeyPrefix != "moon_live_" {
		t.Errorf("APIKeyPrefix = %q, want %q", APIKeyPrefix, "moon_live_")
	}
}

func TestAPIKeyLength(t *testing.T) {
	if APIKeyLength != 64 {
		t.Errorf("APIKeyLength = %d, want 64", APIKeyLength)
	}
}

func TestUserRoleConstants(t *testing.T) {
	if RoleAdmin != "admin" {
		t.Errorf("RoleAdmin = %q, want %q", RoleAdmin, "admin")
	}
	if RoleUser != "user" {
		t.Errorf("RoleUser = %q, want %q", RoleUser, "user")
	}
	if RoleReadOnly != "readonly" {
		t.Errorf("RoleReadOnly = %q, want %q", RoleReadOnly, "readonly")
	}
}

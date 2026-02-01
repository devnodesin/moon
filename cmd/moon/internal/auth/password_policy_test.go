package auth

import (
	"testing"
)

func TestDefaultPasswordPolicy(t *testing.T) {
	policy := DefaultPasswordPolicy()

	if policy.MinLength != 8 {
		t.Errorf("MinLength = %d, want 8", policy.MinLength)
	}
	if !policy.RequireUppercase {
		t.Error("RequireUppercase should be true")
	}
	if !policy.RequireLowercase {
		t.Error("RequireLowercase should be true")
	}
	if !policy.RequireDigit {
		t.Error("RequireDigit should be true")
	}
	if policy.RequireSpecialChar {
		t.Error("RequireSpecialChar should be false by default")
	}
}

func TestPasswordPolicy_Validate(t *testing.T) {
	policy := DefaultPasswordPolicy()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid password",
			password: "Password123",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1",
			wantErr:  true,
			errMsg:   "password must be at least 8 characters",
		},
		{
			name:     "no uppercase",
			password: "password123",
			wantErr:  true,
			errMsg:   "password must include at least one uppercase letter",
		},
		{
			name:     "no lowercase",
			password: "PASSWORD123",
			wantErr:  true,
			errMsg:   "password must include at least one lowercase letter",
		},
		{
			name:     "no digit",
			password: "Passwordabc",
			wantErr:  true,
			errMsg:   "password must include at least one number",
		},
		{
			name:     "exactly 8 chars valid",
			password: "Passwo1d",
			wantErr:  false,
		},
		{
			name:     "unicode uppercase",
			password: "passWord123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestPasswordPolicy_ValidateWithSpecialChar(t *testing.T) {
	policy := &PasswordPolicy{
		MinLength:          8,
		RequireUppercase:   true,
		RequireLowercase:   true,
		RequireDigit:       true,
		RequireSpecialChar: true,
		SpecialChars:       "!@#$%^&*",
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid with special char",
			password: "Password1!",
			wantErr:  false,
		},
		{
			name:     "missing special char",
			password: "Password123",
			wantErr:  true,
		},
		{
			name:     "different special char",
			password: "Password1@",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPasswordPolicy_ValidationErrors(t *testing.T) {
	policy := DefaultPasswordPolicy()

	tests := []struct {
		name          string
		password      string
		expectedCount int
	}{
		{
			name:          "valid password",
			password:      "Password123",
			expectedCount: 0,
		},
		{
			name:          "all failures except length",
			password:      "12345678",
			expectedCount: 2, // no uppercase, no lowercase
		},
		{
			name:          "too short only",
			password:      "Pa1",
			expectedCount: 1,
		},
		{
			name:          "multiple failures",
			password:      "abc",
			expectedCount: 3, // too short, no uppercase, no digit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := policy.ValidationErrors(tt.password)
			if len(errors) != tt.expectedCount {
				t.Errorf("ValidationErrors() count = %d, want %d, errors: %v", len(errors), tt.expectedCount, errors)
			}
		})
	}
}

func TestPasswordPolicy_EmptyPassword(t *testing.T) {
	policy := DefaultPasswordPolicy()

	err := policy.Validate("")
	if err == nil {
		t.Error("Validate() should return error for empty password")
	}
}

func TestPasswordPolicy_CustomMinLength(t *testing.T) {
	policy := &PasswordPolicy{
		MinLength:        12,
		RequireUppercase: false,
		RequireLowercase: false,
		RequireDigit:     false,
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "exactly 12 chars",
			password: "abcdefghijkl",
			wantErr:  false,
		},
		{
			name:     "11 chars",
			password: "abcdefghijk",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

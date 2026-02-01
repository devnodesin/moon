package auth

import (
	"strings"
	"testing"
)

func TestValidateBootstrapConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *BootstrapConfig
		wantErr     bool
		errContains string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: false,
		},
		{
			name:    "empty config",
			cfg:     &BootstrapConfig{},
			wantErr: false,
		},
		{
			name: "valid config",
			cfg: &BootstrapConfig{
				Username: "admin",
				Email:    "admin@example.com",
				Password: "securepassword123",
			},
			wantErr: false,
		},
		{
			name: "missing username",
			cfg: &BootstrapConfig{
				Username: "",
				Email:    "admin@example.com",
				Password: "securepassword123",
			},
			wantErr:     true,
			errContains: "username",
		},
		{
			name: "missing email",
			cfg: &BootstrapConfig{
				Username: "admin",
				Email:    "",
				Password: "securepassword123",
			},
			wantErr:     true,
			errContains: "email",
		},
		{
			name: "missing password",
			cfg: &BootstrapConfig{
				Username: "admin",
				Email:    "admin@example.com",
				Password: "",
			},
			wantErr:     true,
			errContains: "password is required",
		},
		{
			name: "password too short",
			cfg: &BootstrapConfig{
				Username: "admin",
				Email:    "admin@example.com",
				Password: "short",
			},
			wantErr:     true,
			errContains: "at least 8 characters",
		},
		{
			name: "password exactly 8 chars",
			cfg: &BootstrapConfig{
				Username: "admin",
				Email:    "admin@example.com",
				Password: "12345678",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBootstrapConfig(tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateBootstrapConfig() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateBootstrapConfig() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateBootstrapConfig() unexpected error = %v", err)
			}
		})
	}
}

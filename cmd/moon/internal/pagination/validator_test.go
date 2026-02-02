package pagination

import (
	"testing"

	"github.com/thalib/moon/cmd/moon/internal/config"
	"github.com/thalib/moon/cmd/moon/internal/constants"
)

func TestValidatePageSize(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		cfg     *config.AppConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid limit",
			limit:   50,
			cfg:     nil,
			wantErr: false,
		},
		{
			name:    "minimum valid limit",
			limit:   1,
			cfg:     nil,
			wantErr: false,
		},
		{
			name:    "maximum valid limit (default max)",
			limit:   200,
			cfg:     nil,
			wantErr: false,
		},
		{
			name:    "zero limit",
			limit:   0,
			cfg:     nil,
			wantErr: true,
			errMsg:  "page size must be at least 1",
		},
		{
			name:    "negative limit",
			limit:   -5,
			cfg:     nil,
			wantErr: true,
			errMsg:  "page size must be at least 1",
		},
		{
			name:    "exceeds default max limit",
			limit:   201,
			cfg:     nil,
			wantErr: true,
			errMsg:  "page size exceeds maximum allowed: 200",
		},
		{
			name:  "within configured max limit",
			limit: 150,
			cfg: &config.AppConfig{
				Pagination: config.PaginationConfig{
					MaxPageSize: 500,
				},
			},
			wantErr: false,
		},
		{
			name:  "exceeds configured max limit",
			limit: 600,
			cfg: &config.AppConfig{
				Pagination: config.PaginationConfig{
					MaxPageSize: 500,
				},
			},
			wantErr: true,
			errMsg:  "page size exceeds maximum allowed: 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePageSize(tt.limit, tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePageSize() expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("ValidatePageSize() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("ValidatePageSize() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCursor(t *testing.T) {
	tests := []struct {
		name    string
		cursor  string
		wantErr bool
	}{
		{
			name:    "empty cursor",
			cursor:  "",
			wantErr: false,
		},
		{
			name:    "valid ULID",
			cursor:  "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantErr: false,
		},
		{
			name:    "invalid cursor too short",
			cursor:  "invalid",
			wantErr: true,
		},
		{
			name:    "invalid cursor wrong length",
			cursor:  "01ARZ3NDEKTSV4RRFFQ69G5FAVX",
			wantErr: true,
		},
		{
			name:    "invalid cursor bad characters",
			cursor:  "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCursor(tt.cursor)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateCursor() expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("ValidateCursor() unexpected error: %v", err)
			}
		})
	}
}

func TestGetDefaultPageSize(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.AppConfig
		want int
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: constants.DefaultPaginationLimit,
		},
		{
			name: "zero config value",
			cfg: &config.AppConfig{
				Pagination: config.PaginationConfig{
					DefaultPageSize: 0,
				},
			},
			want: constants.DefaultPaginationLimit,
		},
		{
			name: "configured value",
			cfg: &config.AppConfig{
				Pagination: config.PaginationConfig{
					DefaultPageSize: 25,
				},
			},
			want: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDefaultPageSize(tt.cfg)
			if got != tt.want {
				t.Errorf("GetDefaultPageSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetMaxPageSize(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.AppConfig
		want int
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: constants.MaxPaginationLimit,
		},
		{
			name: "zero config value",
			cfg: &config.AppConfig{
				Pagination: config.PaginationConfig{
					MaxPageSize: 0,
				},
			},
			want: constants.MaxPaginationLimit,
		},
		{
			name: "configured value",
			cfg: &config.AppConfig{
				Pagination: config.PaginationConfig{
					MaxPageSize: 500,
				},
			},
			want: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMaxPageSize(tt.cfg)
			if got != tt.want {
				t.Errorf("GetMaxPageSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNormalizePageSize(t *testing.T) {
	cfg := &config.AppConfig{
		Pagination: config.PaginationConfig{
			DefaultPageSize: 20,
			MaxPageSize:     100,
		},
	}

	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{
			name:  "zero returns default",
			limit: 0,
			want:  20,
		},
		{
			name:  "negative returns default",
			limit: -5,
			want:  20,
		},
		{
			name:  "valid limit unchanged",
			limit: 50,
			want:  50,
		},
		{
			name:  "exceeds max returns max",
			limit: 150,
			want:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePageSize(tt.limit, cfg)
			if got != tt.want {
				t.Errorf("NormalizePageSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

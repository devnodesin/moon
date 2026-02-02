// Package pagination provides validation and utilities for cursor-based pagination.
// It enforces page size limits and validates ULID cursors for consistent API responses.
package pagination

import (
	"fmt"

	"github.com/thalib/moon/cmd/moon/internal/config"
	"github.com/thalib/moon/cmd/moon/internal/constants"
	moonulid "github.com/thalib/moon/cmd/moon/internal/ulid"
)

// ValidatePageSize validates the requested page size against configured limits.
// Returns an error if the limit is below MinPageSize or above MaxPageSize.
func ValidatePageSize(limit int, cfg *config.AppConfig) error {
	if limit < constants.MinPageSize {
		return fmt.Errorf("page size must be at least %d", constants.MinPageSize)
	}

	maxLimit := constants.MaxPaginationLimit
	if cfg != nil && cfg.Pagination.MaxPageSize > 0 {
		maxLimit = cfg.Pagination.MaxPageSize
	}

	if limit > maxLimit {
		return fmt.Errorf("page size exceeds maximum allowed: %d", maxLimit)
	}

	return nil
}

// ValidateCursor validates a ULID cursor for pagination.
// Empty cursor is valid (means start from beginning).
// Returns an error if the cursor is not a valid ULID format.
func ValidateCursor(cursor string) error {
	if cursor == "" {
		return nil // Empty cursor is valid
	}
	return moonulid.Validate(cursor)
}

// GetDefaultPageSize returns the configured default page size or the constant default.
func GetDefaultPageSize(cfg *config.AppConfig) int {
	if cfg != nil && cfg.Pagination.DefaultPageSize > 0 {
		return cfg.Pagination.DefaultPageSize
	}
	return constants.DefaultPaginationLimit
}

// GetMaxPageSize returns the configured maximum page size or the constant default.
func GetMaxPageSize(cfg *config.AppConfig) int {
	if cfg != nil && cfg.Pagination.MaxPageSize > 0 {
		return cfg.Pagination.MaxPageSize
	}
	return constants.MaxPaginationLimit
}

// NormalizePageSize adjusts the requested page size to be within valid bounds.
// If limit is 0 or negative, returns the default page size.
// If limit exceeds maximum, returns the maximum.
func NormalizePageSize(limit int, cfg *config.AppConfig) int {
	defaultSize := GetDefaultPageSize(cfg)
	maxSize := GetMaxPageSize(cfg)

	if limit <= 0 {
		return defaultSize
	}
	if limit > maxSize {
		return maxSize
	}
	return limit
}

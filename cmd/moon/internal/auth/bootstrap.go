package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/thalib/moon/cmd/moon/internal/database"
)

// BootstrapConfig holds the configuration for bootstrapping the admin user.
type BootstrapConfig struct {
	Username string
	Email    string
	Password string
}

// Bootstrap initializes the auth tables and creates the bootstrap admin user if needed.
func Bootstrap(ctx context.Context, db database.Driver, cfg *BootstrapConfig) error {
	// Initialize auth tables
	if err := initializeSchema(ctx, db); err != nil {
		return fmt.Errorf("failed to initialize auth schema: %w", err)
	}

	// Create bootstrap admin if configured and no users exist
	if cfg != nil && cfg.Username != "" && cfg.Password != "" {
		if err := createBootstrapAdmin(ctx, db, cfg); err != nil {
			return fmt.Errorf("failed to create bootstrap admin: %w", err)
		}
	}

	return nil
}

// initializeSchema creates the auth tables if they don't exist.
func initializeSchema(ctx context.Context, db database.Driver) error {
	statements := GetSchemaSQL(db.Dialect())

	for _, stmt := range statements {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute schema statement: %w", err)
		}
	}

	return nil
}

// createBootstrapAdmin creates the initial admin user if no users exist.
func createBootstrapAdmin(ctx context.Context, db database.Driver, cfg *BootstrapConfig) error {
	repo := NewUserRepository(db)

	// Check if any users exist
	count, err := repo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count > 0 {
		// Users already exist, skip bootstrap
		log.Println("Users already exist, skipping bootstrap admin creation")
		return nil
	}

	// Hash the password
	passwordHash, err := HashPassword(cfg.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the admin user
	admin := &User{
		Username:     cfg.Username,
		Email:        cfg.Email,
		PasswordHash: passwordHash,
		Role:         string(RoleAdmin),
		CanWrite:     true,
	}

	if err := repo.Create(ctx, admin); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Printf("Bootstrap admin user '%s' created successfully", cfg.Username)
	return nil
}

// ValidateBootstrapConfig validates the bootstrap configuration.
func ValidateBootstrapConfig(cfg *BootstrapConfig) error {
	if cfg == nil {
		return nil // Empty config is valid (no bootstrap)
	}

	if cfg.Username == "" && cfg.Email == "" && cfg.Password == "" {
		return nil // All empty is valid (no bootstrap)
	}

	if cfg.Username == "" {
		return fmt.Errorf("bootstrap admin username is required")
	}

	if cfg.Email == "" {
		return fmt.Errorf("bootstrap admin email is required")
	}

	if cfg.Password == "" {
		return fmt.Errorf("bootstrap admin password is required")
	}

	if len(cfg.Password) < 8 {
		return fmt.Errorf("bootstrap admin password must be at least 8 characters")
	}

	return nil
}

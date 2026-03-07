package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// System table DDL
// ---------------------------------------------------------------------------

const ddlUsersTable = `CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    can_write BOOLEAN NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    last_login_at TEXT,
    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_unique UNIQUE (email)
)`

const ddlUsersRoleIndex = `CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)`

const ddlApikeysTable = `CREATE TABLE IF NOT EXISTS apikeys (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    can_write BOOLEAN NOT NULL DEFAULT 0,
    key_hash TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    last_used_at TEXT,
    CONSTRAINT apikeys_name_unique UNIQUE (name),
    CONSTRAINT apikeys_key_hash_unique UNIQUE (key_hash)
)`

const ddlApikeysLastUsedIndex = `CREATE INDEX IF NOT EXISTS idx_apikeys_last_used_at ON apikeys(last_used_at)`

const ddlRefreshTokensTable = `CREATE TABLE IF NOT EXISTS moon_auth_refresh_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    refresh_token_hash TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    last_used_at TEXT,
    revoked_at TEXT,
    revocation_reason TEXT
)`

const ddlRefreshTokensHashIndex = `CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON moon_auth_refresh_tokens(refresh_token_hash)`

const ddlRefreshTokensUserRevokedIndex = `CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_revoked ON moon_auth_refresh_tokens(user_id, revoked_at)`

const ddlRefreshTokensExpiresIndex = `CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON moon_auth_refresh_tokens(expires_at)`

// systemDDL lists every DDL statement executed during startup reconciliation,
// in the order they must run.
var systemDDL = []string{
	ddlUsersTable,
	ddlUsersRoleIndex,
	ddlApikeysTable,
	ddlApikeysLastUsedIndex,
	ddlRefreshTokensTable,
	ddlRefreshTokensHashIndex,
	ddlRefreshTokensUserRevokedIndex,
	ddlRefreshTokensExpiresIndex,
}

// ---------------------------------------------------------------------------
// EnsureSystemTables creates the required system tables if they do not exist.
// All DDL uses IF NOT EXISTS so calls are idempotent.
// ---------------------------------------------------------------------------

func EnsureSystemTables(ctx context.Context, db DatabaseAdapter) error {
	for _, ddl := range systemDDL {
		if err := db.ExecDDL(ctx, ddl); err != nil {
			return fmt.Errorf("ensure system tables: %w", err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// CreateBootstrapAdmin creates the initial admin user when all bootstrap
// fields are configured and no admin user exists yet.
// ---------------------------------------------------------------------------

func CreateBootstrapAdmin(ctx context.Context, db DatabaseAdapter, cfg *AppConfig, logger *Logger) error {
	if cfg.BootstrapAdminUsername == "" || cfg.BootstrapAdminEmail == "" || cfg.BootstrapAdminPassword == "" {
		return nil
	}

	logger.Warn("bootstrap admin fields are present in config; remove them after initial setup",
		"event", "bootstrap admin fields")

	// Check whether an admin already exists.
	rows, _, err := db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "role", Op: "eq", Value: "admin"}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		return fmt.Errorf("bootstrap admin: check existing: %w", err)
	}
	if len(rows) > 0 {
		logger.Info("admin user already exists; skipping bootstrap")
		return nil
	}

	hash, err := HashPassword(cfg.BootstrapAdminPassword)
	if err != nil {
		return fmt.Errorf("bootstrap admin: hash password: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	admin := map[string]any{
		"id":            GenerateULID(),
		"username":      cfg.BootstrapAdminUsername,
		"email":         cfg.BootstrapAdminEmail,
		"password_hash": hash,
		"role":          "admin",
		"can_write":     int64(1),
		"created_at":    now,
		"updated_at":    now,
	}

	if err := db.InsertRow(ctx, "users", admin); err != nil {
		return fmt.Errorf("bootstrap admin: insert: %w", err)
	}

	logger.Info("bootstrap admin user created",
		"username", cfg.BootstrapAdminUsername)
	return nil
}

// ---------------------------------------------------------------------------
// GenerateULID returns a new ULID string using crypto/rand.
// ---------------------------------------------------------------------------

func GenerateULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

// ---------------------------------------------------------------------------
// HashPassword returns a bcrypt hash of the given password at BcryptCost.
// ---------------------------------------------------------------------------

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

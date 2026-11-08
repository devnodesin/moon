package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
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
    password_hash TEXT,
    role TEXT NOT NULL,
    can_write BOOLEAN NOT NULL DEFAULT 0,
    login_method TEXT NOT NULL DEFAULT 'password',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    last_login_at TEXT,
    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_login_method_check CHECK (login_method IN ('password', 'otp'))
)`

const ddlUsersRoleIndex = `CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)`

const ddlApikeysTable = `CREATE TABLE IF NOT EXISTS apikeys (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    can_write BOOLEAN NOT NULL DEFAULT 0,
    collections JSON NOT NULL DEFAULT '[]',
    is_website BOOLEAN NOT NULL DEFAULT 0,
    allowed_origins JSON,
    rate_limit INTEGER NOT NULL DEFAULT 15,
    captcha_required BOOLEAN NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT 1,
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
const ddlOTPCodesTable = `CREATE TABLE IF NOT EXISTS moon_auth_otp_codes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    consumed_at TEXT
)`

const ddlOTPCodesUserPurposeIndex = `CREATE INDEX IF NOT EXISTS idx_otp_codes_user_purpose ON moon_auth_otp_codes(user_id, purpose)`

const ddlOTPCodesExpiresIndex = `CREATE INDEX IF NOT EXISTS idx_otp_codes_expires_at ON moon_auth_otp_codes(expires_at)`

const ddlSchemaMigrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL
)`

var systemDDL = []string{
	ddlUsersTable,
	ddlUsersRoleIndex,
	ddlApikeysTable,
	ddlApikeysLastUsedIndex,
	ddlRefreshTokensTable,
	ddlRefreshTokensHashIndex,
	ddlRefreshTokensUserRevokedIndex,
	ddlRefreshTokensExpiresIndex,
	ddlOTPCodesTable,
	ddlOTPCodesUserPurposeIndex,
	ddlOTPCodesExpiresIndex,
	ddlSchemaMigrationsTable,
}

// ---------------------------------------------------------------------------
// EnsureSystemTables creates the required system tables if they do not exist.
// All DDL uses IF NOT EXISTS so calls are idempotent. The DDL is written in
// SQLite-compatible form; for the PostgreSQL backend a few defaults are
// normalized (e.g. BOOLEAN DEFAULT 0 is invalid in PostgreSQL, which requires
// DEFAULT false).
// ---------------------------------------------------------------------------

func EnsureSystemTables(ctx context.Context, db DatabaseAdapter) error {
	_, isPostgres := db.(*PostgresAdapter)
	for _, ddl := range systemDDL {
		normalized := ddl
		if isPostgres {
			normalized = normalizeDDLForPostgres(ddl)
		}
		if err := db.ExecDDL(ctx, normalized); err != nil {
			return fmt.Errorf("ensure system tables: %w", err)
		}
	}

	// Migrate existing system tables that predate schema changes: add any
	// missing columns so deployments don't require manual migration.
	if err := ensureSystemTableColumns(ctx, db, isPostgres); err != nil {
		return fmt.Errorf("ensure system tables: migrate columns: %w", err)
	}
	return nil
}

// systemTableColumns declares, for each system table, the columns that must exist
// and the ALTER TABLE ... ADD COLUMN statement used to create them when missing.
// Only columns added after the initial schema need entries here; CREATE TABLE
// IF NOT EXISTS covers brand-new databases.
var systemTableColumns = map[string]map[string]string{
	"users": {
		"login_method": "ALTER TABLE users ADD COLUMN login_method TEXT NOT NULL DEFAULT 'password'",
	},
	"apikeys":                  {},
	"moon_auth_refresh_tokens": {},
	"moon_auth_otp_codes":      {},
}

// ensureSystemTableColumns adds any missing columns to existing system tables.
func ensureSystemTableColumns(ctx context.Context, db DatabaseAdapter, isPostgres bool) error {
	for table, cols := range systemTableColumns {
		existing, err := db.DescribeTable(ctx, table)
		if err != nil {
			// Table may not exist yet (handled by DDL); skip.
			continue
		}
		have := make(map[string]bool)
		for _, c := range existing {
			have[c.Name] = true
		}
		for name, ddl := range cols {
			if have[name] {
				continue
			}
			if err := db.ExecDDL(ctx, ddl); err != nil {
				return fmt.Errorf("table %s column %s: %w", table, name, err)
			}
		}
		// Make password_hash nullable on users table for OTP-only accounts.
		if table == "users" {
			for _, c := range existing {
				if c.Name == "password_hash" && !c.Nullable {
					_, isPostgres := db.(*PostgresAdapter)
					_, isMySQL := db.(*MySQLAdapter)
					if isPostgres {
						if err := db.ExecDDL(ctx, "ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL"); err != nil {
							return fmt.Errorf("table %s column password_hash: %w", table, err)
						}
					} else if isMySQL {
						if err := db.ExecDDL(ctx, "ALTER TABLE users MODIFY password_hash TEXT"); err != nil {
							return fmt.Errorf("table %s column password_hash: %w", table, err)
						}
					} else {
						// SQLite: rebuild the table to drop the NOT NULL constraint.
						// Verify row count before and after to guard against data loss.
						beforeCount, _ := db.CountRows(ctx, "users")
						newDDL := strings.Replace(ddlUsersTable, "CREATE TABLE IF NOT EXISTS users (", "CREATE TABLE users_new (", 1)
						if err := db.ExecDDL(ctx, newDDL); err != nil {
							return fmt.Errorf("table %s rebuild create: %w", table, err)
						}
						copyCols := "id, username, email, password_hash, role, can_write, login_method, created_at, updated_at, last_login_at"
						if err := db.ExecDDL(ctx, "INSERT INTO users_new ("+copyCols+") SELECT "+copyCols+" FROM users"); err != nil {
							_ = db.ExecDDL(ctx, "DROP TABLE IF EXISTS users_new")
							return fmt.Errorf("table %s rebuild copy: %w", table, err)
						}
						afterCount, _ := db.CountRows(ctx, "users_new")
						if beforeCount > 0 && afterCount != beforeCount {
							_ = db.ExecDDL(ctx, "DROP TABLE IF EXISTS users_new")
							return fmt.Errorf("table %s rebuild: row count mismatch (%d vs %d)", table, beforeCount, afterCount)
						}
						if err := db.ExecDDL(ctx, "DROP TABLE users"); err != nil {
							return fmt.Errorf("table %s rebuild drop: %w", table, err)
						}
						if err := db.ExecDDL(ctx, "ALTER TABLE users_new RENAME TO users"); err != nil {
							return fmt.Errorf("table %s rebuild rename: %w", table, err)
						}
						if err := db.ExecDDL(ctx, ddlUsersRoleIndex); err != nil {
							return fmt.Errorf("table %s rebuild index: %w", table, err)
						}
					}
				}
			}
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

// ---------------------------------------------------------------------------
// Backend DDL normalization
// ---------------------------------------------------------------------------

// normalizeDDLForPostgres rewrites SQLite-oriented DDL fragments that are
// invalid in PostgreSQL. PostgreSQL requires boolean literals (true/false)
// rather than integer literals (0/1) for BOOLEAN columns.
func normalizeDDLForPostgres(ddl string) string {
	replacer := strings.NewReplacer(
		"BOOLEAN NOT NULL DEFAULT 0", "BOOLEAN NOT NULL DEFAULT false",
		"BOOLEAN NOT NULL DEFAULT 1", "BOOLEAN NOT NULL DEFAULT true",
	)
	return replacer.Replace(ddl)
}

package auth

import (
	"github.com/thalib/moon/cmd/moon/internal/database"
)

// Schema SQL statements for authentication tables.
// These are used during database initialization.

// GetSchemaSQL returns the SQL statements to create auth tables for the given dialect.
func GetSchemaSQL(dialect database.DialectType) []string {
	switch dialect {
	case database.DialectPostgres:
		return getPostgresSchema()
	case database.DialectMySQL:
		return getMySQLSchema()
	default:
		return getSQLiteSchema()
	}
}

func getSQLiteSchema() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ulid TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			can_write INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			last_login_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_ulid ON users(ulid)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			last_used_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)`,

		`CREATE TABLE IF NOT EXISTS apikeys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ulid TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			key_hash TEXT NOT NULL UNIQUE,
			role TEXT NOT NULL DEFAULT 'user',
			can_write INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			last_used_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_apikeys_ulid ON apikeys(ulid)`,
		`CREATE INDEX IF NOT EXISTS idx_apikeys_key_hash ON apikeys(key_hash)`,
	}
}

func getPostgresSchema() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			ulid VARCHAR(26) NOT NULL UNIQUE,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			can_write BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			last_login_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_ulid ON users(ulid)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(64) NOT NULL UNIQUE,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			last_used_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)`,

		`CREATE TABLE IF NOT EXISTS apikeys (
			id BIGSERIAL PRIMARY KEY,
			ulid VARCHAR(26) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			key_hash VARCHAR(64) NOT NULL UNIQUE,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			can_write BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL,
			last_used_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_apikeys_ulid ON apikeys(ulid)`,
		`CREATE INDEX IF NOT EXISTS idx_apikeys_key_hash ON apikeys(key_hash)`,
	}
}

func getMySQLSchema() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			ulid VARCHAR(26) NOT NULL UNIQUE,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			can_write BOOLEAN NOT NULL DEFAULT true,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			last_login_at DATETIME,
			INDEX idx_users_ulid (ulid),
			INDEX idx_users_username (username),
			INDEX idx_users_email (email)
		)`,

		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			token_hash VARCHAR(64) NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			last_used_at DATETIME NOT NULL,
			INDEX idx_refresh_tokens_token_hash (token_hash),
			INDEX idx_refresh_tokens_user_id (user_id),
			INDEX idx_refresh_tokens_expires_at (expires_at),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS apikeys (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			ulid VARCHAR(26) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			key_hash VARCHAR(64) NOT NULL UNIQUE,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			can_write BOOLEAN NOT NULL DEFAULT true,
			created_at DATETIME NOT NULL,
			last_used_at DATETIME,
			INDEX idx_apikeys_ulid (ulid),
			INDEX idx_apikeys_key_hash (key_hash)
		)`,
	}
}

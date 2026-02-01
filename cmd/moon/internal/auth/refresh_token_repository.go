package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/thalib/moon/cmd/moon/internal/database"
)

// RefreshTokenRepository provides database operations for refresh tokens.
type RefreshTokenRepository struct {
	db database.Driver
}

// NewRefreshTokenRepository creates a new refresh token repository.
func NewRefreshTokenRepository(db database.Driver) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// HashToken hashes a refresh token using SHA-256.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Create creates a new refresh token in the database.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *RefreshToken) error {
	token.CreatedAt = time.Now()
	token.LastUsedAt = time.Now()

	var query string
	switch r.db.Dialect() {
	case database.DialectPostgres:
		query = `INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at, last_used_at)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`
		err := r.db.QueryRow(ctx, query,
			token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt, token.LastUsedAt,
		).Scan(&token.ID)
		return err
	default:
		query = `INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at, last_used_at)
			VALUES (?, ?, ?, ?, ?)`
		result, err := r.db.Exec(ctx, query,
			token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt, token.LastUsedAt,
		)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		token.ID = id
		return nil
	}
}

// GetByHash retrieves a refresh token by its hash.
func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := "SELECT id, user_id, token_hash, expires_at, created_at, last_used_at FROM refresh_tokens WHERE token_hash = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "SELECT id, user_id, token_hash, expires_at, created_at, last_used_at FROM refresh_tokens WHERE token_hash = $1"
	}

	token := &RefreshToken{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt, &token.LastUsedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return token, nil
}

// UpdateLastUsed updates the last used time for a refresh token.
func (r *RefreshTokenRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	var query string
	switch r.db.Dialect() {
	case database.DialectPostgres:
		query = "UPDATE refresh_tokens SET last_used_at = $1 WHERE id = $2"
	default:
		query = "UPDATE refresh_tokens SET last_used_at = ? WHERE id = ?"
	}

	_, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}

// Delete deletes a refresh token from the database.
func (r *RefreshTokenRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM refresh_tokens WHERE id = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "DELETE FROM refresh_tokens WHERE id = $1"
	}

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// DeleteByHash deletes a refresh token by its hash.
func (r *RefreshTokenRepository) DeleteByHash(ctx context.Context, tokenHash string) error {
	query := "DELETE FROM refresh_tokens WHERE token_hash = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "DELETE FROM refresh_tokens WHERE token_hash = $1"
	}

	_, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all refresh tokens for a user.
func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := "DELETE FROM refresh_tokens WHERE user_id = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "DELETE FROM refresh_tokens WHERE user_id = $1"
	}

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user tokens: %w", err)
	}
	return nil
}

// DeleteExpired deletes all expired refresh tokens.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := "DELETE FROM refresh_tokens WHERE expires_at < ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "DELETE FROM refresh_tokens WHERE expires_at < $1"
	}

	result, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return count, nil
}

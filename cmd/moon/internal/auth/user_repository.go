package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thalib/moon/cmd/moon/internal/database"
	moonulid "github.com/thalib/moon/cmd/moon/internal/ulid"
)

// UserRepository provides database operations for users.
type UserRepository struct {
	db database.Driver
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db database.Driver) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database.
func (r *UserRepository) Create(ctx context.Context, user *User) error {
	user.ULID = moonulid.Generate()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	var query string
	switch r.db.Dialect() {
	case database.DialectPostgres:
		query = `INSERT INTO users (ulid, username, email, password_hash, role, can_write, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
		err := r.db.QueryRow(ctx, query,
			user.ULID, user.Username, user.Email, user.PasswordHash,
			user.Role, user.CanWrite, user.CreatedAt, user.UpdatedAt,
		).Scan(&user.ID)
		return err
	default:
		query = `INSERT INTO users (ulid, username, email, password_hash, role, can_write, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		result, err := r.db.Exec(ctx, query,
			user.ULID, user.Username, user.Email, user.PasswordHash,
			user.Role, user.CanWrite, user.CreatedAt, user.UpdatedAt,
		)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		user.ID = id
		return nil
	}
}

// GetByID retrieves a user by internal ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	query := "SELECT id, ulid, username, email, password_hash, role, can_write, created_at, updated_at, last_login_at FROM users WHERE id = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "SELECT id, ulid, username, email, password_hash, role, can_write, created_at, updated_at, last_login_at FROM users WHERE id = $1"
	}

	user := &User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.ULID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CanWrite, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetByULID retrieves a user by ULID.
func (r *UserRepository) GetByULID(ctx context.Context, ulid string) (*User, error) {
	query := "SELECT id, ulid, username, email, password_hash, role, can_write, created_at, updated_at, last_login_at FROM users WHERE ulid = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "SELECT id, ulid, username, email, password_hash, role, can_write, created_at, updated_at, last_login_at FROM users WHERE ulid = $1"
	}

	user := &User{}
	err := r.db.QueryRow(ctx, query, ulid).Scan(
		&user.ID, &user.ULID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CanWrite, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := "SELECT id, ulid, username, email, password_hash, role, can_write, created_at, updated_at, last_login_at FROM users WHERE username = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "SELECT id, ulid, username, email, password_hash, role, can_write, created_at, updated_at, last_login_at FROM users WHERE username = $1"
	}

	user := &User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.ULID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CanWrite, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := "SELECT id, ulid, username, email, password_hash, role, can_write, created_at, updated_at, last_login_at FROM users WHERE email = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "SELECT id, ulid, username, email, password_hash, role, can_write, created_at, updated_at, last_login_at FROM users WHERE email = $1"
	}

	user := &User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.ULID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CanWrite, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// Update updates a user in the database.
func (r *UserRepository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	var query string
	switch r.db.Dialect() {
	case database.DialectPostgres:
		query = `UPDATE users SET username = $1, email = $2, password_hash = $3, role = $4, 
			can_write = $5, updated_at = $6, last_login_at = $7 WHERE id = $8`
	default:
		query = `UPDATE users SET username = ?, email = ?, password_hash = ?, role = ?, 
			can_write = ?, updated_at = ?, last_login_at = ? WHERE id = ?`
	}

	_, err := r.db.Exec(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.Role,
		user.CanWrite, user.UpdatedAt, user.LastLoginAt, user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// UpdateLastLogin updates the last login time for a user.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	now := time.Now()
	var query string
	switch r.db.Dialect() {
	case database.DialectPostgres:
		query = "UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3"
	default:
		query = "UPDATE users SET last_login_at = ?, updated_at = ? WHERE id = ?"
	}

	_, err := r.db.Exec(ctx, query, now, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// Delete deletes a user from the database.
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "DELETE FROM users WHERE id = $1"
	}

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Count returns the total number of users.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// Exists checks if a user exists by username or email.
func (r *UserRepository) Exists(ctx context.Context, username, email string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE username = ? OR email = ?"
	if r.db.Dialect() == database.DialectPostgres {
		query = "SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2"
	}

	var count int64
	err := r.db.QueryRow(ctx, query, username, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}

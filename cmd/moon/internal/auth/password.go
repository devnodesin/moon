// Package auth provides authentication services including password hashing,
// JWT token management, API key generation, and user/session management.
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is the cost factor for bcrypt hashing (2^12 iterations).
const BcryptCost = 12

// HashPassword hashes a password using bcrypt with the configured cost factor.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// ComparePassword compares a plaintext password with a bcrypt hash.
// Returns nil if they match, or an error if they don't.
func ComparePassword(hashedPassword, password string) error {
	if hashedPassword == "" || password == "" {
		return fmt.Errorf("password or hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}

	return nil
}

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CreateAccessToken signs a JWT with the standard Moon claims.
func CreateAccessToken(userID, jti, role string, canWrite bool, secret string, expirySeconds int) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(time.Duration(expirySeconds) * time.Second)

	claims := jwt.MapClaims{
		"sub":       userID,
		"jti":       jti,
		"role":      role,
		"can_write": canWrite,
		"exp":       exp.Unix(),
		"iat":       now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}
	return signed, exp, nil
}

// GenerateRefreshToken creates a cryptographically random refresh token
// and returns both the raw base64url-encoded value and its SHA-256 hash
// (hex-encoded).
func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	raw = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	hash = HashRefreshToken(raw)
	return raw, hash, nil
}

// HashRefreshToken returns the hex-encoded SHA-256 hash of a raw refresh token.
func HashRefreshToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

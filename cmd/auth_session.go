package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthSessionHandler implements POST /auth:session with login, refresh, and logout operations.
type AuthSessionHandler struct {
	db  DatabaseAdapter
	cfg *AppConfig
}

type authSessionRequest struct {
	Op   string         `json:"op"`
	Data map[string]any `json:"data"`
}

type sessionUser struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	CanWrite    bool    `json:"can_write"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	LastLoginAt *string `json:"last_login_at"`
}

type sessionPayload struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    string      `json:"expires_at"`
	TokenType    string      `json:"token_type"`
	User         sessionUser `json:"user"`
}

// HandleSession dispatches to the appropriate operation based on the "op" field.
func (h *AuthSessionHandler) HandleSession(w http.ResponseWriter, r *http.Request) {
	var req authSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Op == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: op")
		return
	}

	if req.Data == nil {
		WriteError(w, http.StatusBadRequest, "Missing required field: data")
		return
	}

	switch req.Op {
	case "login":
		h.handleLogin(w, r, req.Data)
	case "refresh":
		h.handleRefresh(w, r, req.Data)
	case "logout":
		h.handleLogout(w, r, req.Data)
	default:
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Unknown op: %s", req.Op))
	}
}

func (h *AuthSessionHandler) handleLogin(w http.ResponseWriter, r *http.Request, data map[string]any) {
	usernameRaw, ok := data["username"]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.username")
		return
	}
	username, ok := usernameRaw.(string)
	if !ok || username == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.username")
		return
	}

	passwordRaw, ok := data["password"]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.password")
		return
	}
	password, ok := passwordRaw.(string)
	if !ok || password == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.password")
		return
	}

	ctx := r.Context()
	username = strings.ToLower(username)

	rows, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "username", Op: "eq", Value: username}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(rows) == 0 {
		WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	user := rows[0]
	storedHash, _ := user["password_hash"].(string)
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	userID, _ := user["id"].(string)
	role, _ := user["role"].(string)
	canWrite := toBool(user["can_write"])

	payload, err := h.issueSession(ctx, userID, role, canWrite, user)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_ = h.db.UpdateRow(ctx, "users", userID, map[string]any{
		"last_login_at": now,
		"updated_at":    now,
	})

	payload.User.LastLoginAt = &now

	WriteSuccess(w, http.StatusOK, "Login successful", []any{payload})
}

func (h *AuthSessionHandler) handleRefresh(w http.ResponseWriter, r *http.Request, data map[string]any) {
	tokenRaw, ok := data["refresh_token"]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.refresh_token")
		return
	}
	tokenStr, ok := tokenRaw.(string)
	if !ok || tokenStr == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.refresh_token")
		return
	}

	ctx := r.Context()
	tokenHash := HashRefreshToken(tokenStr)

	tokenRows, _, err := h.db.QueryRows(ctx, "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "refresh_token_hash", Op: "eq", Value: tokenHash}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(tokenRows) == 0 {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	tokenRow := tokenRows[0]

	if tokenRow["revoked_at"] != nil {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	expiresAtStr, _ := tokenRow["expires_at"].(string)
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil || time.Now().UTC().After(expiresAt) {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	tokenID, _ := tokenRow["id"].(string)
	now := time.Now().UTC().Format(time.RFC3339)
	_ = h.db.UpdateRow(ctx, "moon_auth_refresh_tokens", tokenID, map[string]any{
		"revoked_at":        now,
		"revocation_reason": "rotated",
		"last_used_at":      now,
	})

	userID, _ := tokenRow["user_id"].(string)
	userRows, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: userID}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil || len(userRows) == 0 {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	user := userRows[0]
	role, _ := user["role"].(string)
	canWrite := toBool(user["can_write"])

	payload, err := h.issueSession(ctx, userID, role, canWrite, user)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	WriteSuccess(w, http.StatusOK, "Token refreshed successfully", []any{payload})
}

func (h *AuthSessionHandler) handleLogout(w http.ResponseWriter, r *http.Request, data map[string]any) {
	tokenRaw, ok := data["refresh_token"]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.refresh_token")
		return
	}
	tokenStr, ok := tokenRaw.(string)
	if !ok || tokenStr == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.refresh_token")
		return
	}

	ctx := r.Context()
	tokenHash := HashRefreshToken(tokenStr)

	tokenRows, _, err := h.db.QueryRows(ctx, "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "refresh_token_hash", Op: "eq", Value: tokenHash}},
		Page:    1,
		PerPage: 1,
	})
	if err == nil && len(tokenRows) > 0 {
		tokenRow := tokenRows[0]
		if tokenRow["revoked_at"] == nil {
			tokenID, _ := tokenRow["id"].(string)
			now := time.Now().UTC().Format(time.RFC3339)
			_ = h.db.UpdateRow(ctx, "moon_auth_refresh_tokens", tokenID, map[string]any{
				"revoked_at":        now,
				"revocation_reason": "logout",
			})
		}
	}

	WriteMessage(w, http.StatusOK, "Logged out successfully")
}

// issueSession creates a new JWT + refresh token pair and stores the refresh token.
func (h *AuthSessionHandler) issueSession(ctx context.Context, userID, role string, canWrite bool, user map[string]any) (*sessionPayload, error) {
	jti := GenerateULID()

	accessToken, expiresAt, err := CreateAccessToken(userID, jti, role, canWrite, h.cfg.JWTSecret, h.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("issue session: %w", err)
	}

	rawRefresh, refreshHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("issue session: %w", err)
	}

	now := time.Now().UTC()
	refreshExpiry := now.Add(time.Duration(h.cfg.JWTRefreshExpiry) * time.Second)

	err = h.db.InsertRow(ctx, "moon_auth_refresh_tokens", map[string]any{
		"id":                 GenerateULID(),
		"user_id":            userID,
		"refresh_token_hash": refreshHash,
		"expires_at":         refreshExpiry.Format(time.RFC3339),
		"created_at":         now.Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("issue session: store refresh token: %w", err)
	}

	var lastLogin *string
	if v, ok := user["last_login_at"].(string); ok && v != "" {
		lastLogin = &v
	}

	return &sessionPayload{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
		TokenType:    "Bearer",
		User: sessionUser{
			ID:          userID,
			Username:    stringVal(user, "username"),
			Email:       stringVal(user, "email"),
			Role:        role,
			CanWrite:    canWrite,
			CreatedAt:   stringVal(user, "created_at"),
			UpdatedAt:   stringVal(user, "updated_at"),
			LastLoginAt: lastLogin,
		},
	}, nil
}

func stringVal(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func toBool(v any) bool {
	switch b := v.(type) {
	case bool:
		return b
	case int64:
		return b != 0
	case float64:
		return b != 0
	case int:
		return b != 0
	case string:
		return b == "1" || b == "true"
	default:
		return false
	}
}

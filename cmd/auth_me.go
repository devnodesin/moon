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

// AuthMeHandler implements GET /auth:me and POST /auth:me.
type AuthMeHandler struct {
	db  DatabaseAdapter
	cfg *AppConfig
}

// NewAuthMeHandler creates a new AuthMeHandler with its dependencies.
func NewAuthMeHandler(db DatabaseAdapter, cfg *AppConfig) *AuthMeHandler {
	return &AuthMeHandler{db: db, cfg: cfg}
}

// nonWritableFields lists fields that cannot be set via POST /auth:me.
var nonWritableFields = map[string]bool{
	"id":            true,
	"username":      true,
	"role":          true,
	"can_write":     true,
	"created_at":    true,
	"updated_at":    true,
	"last_login_at": true,
	"password_hash": true,
}

// apiVisibleUserFields are the fields returned in auth:me responses.
var apiVisibleUserFields = []string{
	"id", "username", "email", "role", "can_write",
	"created_at", "updated_at", "last_login_at",
}

// GetMe handles GET /auth:me — returns the current authenticated user.
func (h *AuthMeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	identity, ok := GetAuthIdentity(r.Context())
	if !ok || identity.CredentialType != CredentialTypeJWT {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.lookupUser(r.Context(), identity.CallerID)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	WriteSuccess(w, http.StatusOK, "Current user retrieved successfully", []any{buildUserResponse(user)})
}

// UpdateMe handles POST /auth:me — updates email and/or password.
func (h *AuthMeHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	identity, ok := GetAuthIdentity(r.Context())
	if !ok || identity.CredentialType != CredentialTypeJWT {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if body.Data == nil {
		WriteError(w, http.StatusBadRequest, "Missing required field: data")
		return
	}

	if len(body.Data) == 0 {
		WriteError(w, http.StatusBadRequest, "No updatable fields provided")
		return
	}

	// Reject non-writable fields.
	for key := range body.Data {
		if nonWritableFields[key] {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("Field %q is not writable", key))
			return
		}
	}

	emailRaw, hasEmail := body.Data["email"]
	passwordRaw, hasPassword := body.Data["password"]
	_, hasOldPassword := body.Data["old_password"]

	// At least one of email or password must be present.
	if !hasEmail && !hasPassword {
		WriteError(w, http.StatusBadRequest, "At least one of email or password must be provided")
		return
	}

	user, err := h.lookupUser(r.Context(), identity.CallerID)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx := r.Context()

	if hasPassword {
		newPassword, ok := passwordRaw.(string)
		if !ok || newPassword == "" {
			WriteError(w, http.StatusBadRequest, "Field \"password\" must be a non-empty string")
			return
		}

		if !hasOldPassword {
			WriteError(w, http.StatusBadRequest, "Field \"old_password\" is required when changing password")
			return
		}
		oldPassword, ok := body.Data["old_password"].(string)
		if !ok || oldPassword == "" {
			WriteError(w, http.StatusBadRequest, "Field \"old_password\" must be a non-empty string")
			return
		}

		storedHash, _ := user["password_hash"].(string)
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(oldPassword)); err != nil {
			WriteError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		if err := validatePasswordPolicy(newPassword); err != nil {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("Password policy violation: %s", err.Error()))
			return
		}

		hash, err := HashPassword(newPassword)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		userID, _ := user["id"].(string)
		if err := h.db.UpdateRow(ctx, "users", userID, map[string]any{
			"password_hash": hash,
			"updated_at":    now,
		}); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Revoke all active refresh tokens for this user.
		if err := h.revokeAllRefreshTokens(ctx, userID, "password_changed"); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Re-fetch the user to return updated state.
		user, err = h.lookupUser(ctx, userID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		WriteSuccess(w, http.StatusOK, "Password updated successfully. Sign in again.", []any{buildUserResponse(user)})
		return
	}

	// Email-only update.
	newEmail, ok := emailRaw.(string)
	if !ok || newEmail == "" {
		WriteError(w, http.StatusBadRequest, "Field \"email\" must be a non-empty string")
		return
	}

	newEmail = strings.ToLower(newEmail)
	if !isValidEmail(newEmail) {
		WriteError(w, http.StatusBadRequest, "Invalid email address")
		return
	}

	// Check uniqueness.
	existing, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "email", Op: "eq", Value: newEmail}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	userID, _ := user["id"].(string)
	if len(existing) > 0 {
		existingID, _ := existing[0]["id"].(string)
		if existingID != userID {
			WriteError(w, http.StatusConflict, "Email already in use")
			return
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if err := h.db.UpdateRow(ctx, "users", userID, map[string]any{
		"email":      newEmail,
		"updated_at": now,
	}); err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	user, err = h.lookupUser(ctx, userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	WriteSuccess(w, http.StatusOK, "Current user updated successfully", []any{buildUserResponse(user)})
}

// lookupUser fetches a user by ID, returning the full row or an error.
func (h *AuthMeHandler) lookupUser(ctx context.Context, userID string) (map[string]any, error) {
	rows, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: userID}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return rows[0], nil
}

// buildUserResponse converts a raw user row to an API-visible map.
func buildUserResponse(user map[string]any) map[string]any {
	out := make(map[string]any, len(apiVisibleUserFields))
	for _, f := range apiVisibleUserFields {
		v, ok := user[f]
		if !ok {
			continue
		}
		if f == "can_write" {
			out[f] = toBool(v)
		} else {
			out[f] = v
		}
	}
	return out
}

// revokeAllRefreshTokens revokes all active (non-revoked) refresh tokens
// for the given user by setting revoked_at and revocation_reason.
func (h *AuthMeHandler) revokeAllRefreshTokens(ctx context.Context, userID, reason string) error {
	rows, _, err := h.db.QueryRows(ctx, "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{
			{Field: "user_id", Op: "eq", Value: userID},
		},
		Page:    1,
		PerPage: MaxPerPage,
	})
	if err != nil {
		return fmt.Errorf("revoke tokens: query: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	for _, row := range rows {
		if row["revoked_at"] != nil {
			continue
		}
		tokenID, _ := row["id"].(string)
		if tokenID == "" {
			continue
		}
		if err := h.db.UpdateRow(ctx, "moon_auth_refresh_tokens", tokenID, map[string]any{
			"revoked_at":        now,
			"revocation_reason": reason,
		}); err != nil {
			return fmt.Errorf("revoke tokens: update %s: %w", tokenID, err)
		}
	}
	return nil
}

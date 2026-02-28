package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/thalib/moon/cmd/moon/internal/auth"
	"github.com/thalib/moon/cmd/moon/internal/constants"
	"github.com/thalib/moon/cmd/moon/internal/database"
)

// APIKeysHandler handles API key management endpoints (admin only).
type APIKeysHandler struct {
	db           database.Driver
	apiKeyRepo   *auth.APIKeyRepository
	tokenService *auth.TokenService
}

// NewAPIKeysHandler creates a new API keys handler.
func NewAPIKeysHandler(db database.Driver, jwtSecret string, accessExpiry, refreshExpiry int) *APIKeysHandler {
	return &APIKeysHandler{
		db:           db,
		apiKeyRepo:   auth.NewAPIKeyRepository(db),
		tokenService: auth.NewTokenService(jwtSecret, accessExpiry, refreshExpiry),
	}
}

// API key name validation constants.
const (
	MinKeyNameLength     = 3
	MaxKeyNameLength     = 100
	MaxDescriptionLength = 500
)

// APIKeyPublicInfo represents public API key information (no actual key).
type APIKeyPublicInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Role        string  `json:"role"`
	CanWrite    bool    `json:"can_write"`
	CreatedAt   string  `json:"created_at"`
	LastUsedAt  *string `json:"last_used_at,omitempty"`
}

// CreateAPIKeyRequest represents a request to create an API key.
type CreateAPIKeyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Role        string `json:"role"`
	CanWrite    *bool  `json:"can_write,omitempty"`
}

// UpdateAPIKeyRequest represents a request to update an API key.
type UpdateAPIKeyRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	CanWrite    *bool   `json:"can_write,omitempty"`
	Action      string  `json:"action,omitempty"`
}

// ValidAPIKeyRoles returns the valid roles for API keys.
func ValidAPIKeyRoles() []string {
	return []string{"admin", "user"}
}

// IsValidAPIKeyRole checks if a role is valid for API keys.
func IsValidAPIKeyRole(role string) bool {
	for _, r := range ValidAPIKeyRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// List handles GET /apikeys:list
func (h *APIKeysHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, err := h.validateAdminAccess(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	ctx := r.Context()

	limitStr := r.URL.Query().Get(constants.QueryParamLimit)
	after := r.URL.Query().Get("after")

	limit := constants.DefaultPaginationLimit
	if limitStr != "" {
		if l := parseIntWithDefault(limitStr, constants.DefaultPaginationLimit); l > 0 && l <= constants.MaxPaginationLimit {
			limit = l
		}
	}

	keys, err := h.apiKeyRepo.ListPaginated(ctx, auth.APIKeyListOptions{
		Limit:   limit + 1,
		AfterID: after,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}

	var nextCursor *string
	if len(keys) > limit {
		keys = keys[:limit]
		cursor := keys[len(keys)-1].ID
		nextCursor = &cursor
	}

	// Determine prev cursor for backward pagination
	var prevCursor *string
	if after != "" && len(keys) > 0 {
		prevID := h.apiKeyRepo.FindPrevCursorID(ctx, keys[0].ID, limit)
		if prevID != "" {
			prevCursor = &prevID
		}
	}

	publicKeys := make([]APIKeyPublicInfo, len(keys))
	for i, key := range keys {
		publicKeys[i] = apiKeyToPublicInfo(key)
	}

	h.logAdminAction("apikey_list", claims.UserID, "")

	// Build meta with prev/next cursors per SPEC_API.md
	meta := map[string]any{
		"count": len(publicKeys),
		"limit": limit,
		"next":  nextCursor,
		"prev":  prevCursor,
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": publicKeys,
		"meta": meta,
	})
}

// Get handles GET /apikeys:get?id={ulid}
func (h *APIKeysHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	_, err := h.validateAdminAccess(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	ctx := r.Context()

	keyID := r.URL.Query().Get("id")
	if keyID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	apiKey, err := h.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get API key")
		return
	}

	if apiKey == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("API key with id '%s' not found", keyID))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": apiKeyToPublicInfo(apiKey),
	})
}

// Create handles POST /apikeys:create
func (h *APIKeysHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, err := h.validateAdminAccess(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	var wrapper struct {
		Data CreateAPIKeyRequest `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&wrapper); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req := wrapper.Data

	ctx := r.Context()

	// Validate required fields
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Role == "" {
		writeError(w, http.StatusBadRequest, "role is required")
		return
	}

	// Validate name length
	if len(req.Name) < MinKeyNameLength || len(req.Name) > MaxKeyNameLength {
		writeError(w, http.StatusBadRequest, "name must be between 3 and 100 characters")
		return
	}

	// Validate description length
	if len(req.Description) > MaxDescriptionLength {
		writeError(w, http.StatusBadRequest, "description must not exceed 500 characters")
		return
	}

	// Validate role
	if !IsValidAPIKeyRole(req.Role) {
		writeError(w, http.StatusBadRequest, "role must be 'admin' or 'user'")
		return
	}

	// Check if name exists
	exists, err := h.apiKeyRepo.NameExists(ctx, req.Name, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check name")
		return
	}
	if exists {
		writeError(w, http.StatusBadRequest, "API key name already exists")
		return
	}

	// Generate API key
	rawKey, keyHash, err := auth.GenerateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate API key")
		return
	}

	// Determine can_write (default false)
	canWrite := false
	if req.CanWrite != nil {
		canWrite = *req.CanWrite
	}

	apiKey := &auth.APIKey{
		Name:        req.Name,
		Description: req.Description,
		KeyHash:     keyHash,
		Role:        req.Role,
		CanWrite:    canWrite,
	}

	if err := h.apiKeyRepo.Create(ctx, apiKey); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create API key")
		return
	}

	h.logAdminAction("apikey_created", claims.UserID, apiKey.ID)

	dataResp := map[string]any{
		"id":          apiKey.ID,
		"name":        apiKey.Name,
		"description": apiKey.Description,
		"role":        apiKey.Role,
		"can_write":   apiKey.CanWrite,
		"key":         rawKey,
		"created_at":  apiKey.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data":    dataResp,
		"message": "API key created successfully",
		"warning": "Store this key securely. It will not be shown again.",
	})
}

// Update handles POST /apikeys:update?id={ulid}
func (h *APIKeysHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, err := h.validateAdminAccess(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	ctx := r.Context()

	keyID := r.URL.Query().Get("id")
	if keyID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	var wrapper struct {
		Data UpdateAPIKeyRequest `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&wrapper); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req := wrapper.Data

	apiKey, err := h.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get API key")
		return
	}

	if apiKey == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("API key with id '%s' not found", keyID))
		return
	}

	// Handle rotate action
	if req.Action == "rotate" {
		rawKey, keyHash, err := auth.GenerateAPIKey()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate new API key")
			return
		}

		if err := h.apiKeyRepo.UpdateKeyHash(ctx, apiKey.PKID, keyHash); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to rotate API key")
			return
		}

		h.logAdminAction("apikey_rotated", claims.UserID, apiKey.ID)

		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"id":   apiKey.ID,
				"name": apiKey.Name,
				"key":  rawKey,
			},
			"message": "API key rotated successfully",
			"warning": "Store this key securely. The old key is now invalid.",
		})
		return
	}

	// Handle invalid action
	if req.Action != "" {
		writeError(w, http.StatusBadRequest, "invalid action")
		return
	}

	// Normal update: name, description, can_write
	updated := false

	if req.Name != nil {
		// Validate name length
		if len(*req.Name) < MinKeyNameLength || len(*req.Name) > MaxKeyNameLength {
			writeError(w, http.StatusBadRequest, "name must be between 3 and 100 characters")
			return
		}

		// Check if name exists for another key
		exists, err := h.apiKeyRepo.NameExists(ctx, *req.Name, apiKey.PKID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check name")
			return
		}
		if exists {
			writeError(w, http.StatusBadRequest, "API key name already exists")
			return
		}

		apiKey.Name = *req.Name
		updated = true
	}

	if req.Description != nil {
		// Validate description length
		if len(*req.Description) > MaxDescriptionLength {
			writeError(w, http.StatusBadRequest, "description must not exceed 500 characters")
			return
		}
		apiKey.Description = *req.Description
		updated = true
	}

	if req.CanWrite != nil {
		apiKey.CanWrite = *req.CanWrite
		updated = true
	}

	if !updated {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	if err := h.apiKeyRepo.UpdateMetadata(ctx, apiKey); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update API key")
		return
	}

	h.logAdminAction("apikey_updated", claims.UserID, apiKey.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"data":    apiKeyToPublicInfo(apiKey),
		"message": "API key updated successfully",
	})
}

// Destroy handles POST /apikeys:destroy?id={ulid}
func (h *APIKeysHandler) Destroy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, err := h.validateAdminAccess(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	ctx := r.Context()

	keyID := r.URL.Query().Get("id")
	if keyID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	apiKey, err := h.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get API key")
		return
	}

	if apiKey == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("API key with id '%s' not found", keyID))
		return
	}

	if err := h.apiKeyRepo.Delete(ctx, apiKey.PKID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete API key")
		return
	}

	h.logAdminAction("apikey_deleted", claims.UserID, keyID)

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "API key deleted successfully",
	})
}

// validateAdminAccess validates that the request is from an admin user.
func (h *APIKeysHandler) validateAdminAccess(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get(constants.HeaderAuthorization)
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != strings.ToLower(constants.AuthSchemeBearer) {
		return nil, http.ErrNoCookie
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return nil, http.ErrNoCookie
	}

	claims, err := h.tokenService.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	if claims.Role != string(auth.RoleAdmin) {
		return nil, http.ErrNoCookie
	}

	return claims, nil
}

// logAdminAction logs an admin action for audit purposes.
func (h *APIKeysHandler) logAdminAction(action, adminULID, targetULID string) {
	if targetULID != "" {
		log.Printf("INFO: ADMIN_ACTION %s by=%s key_id=%s", action, adminULID, targetULID)
	} else {
		log.Printf("INFO: ADMIN_ACTION %s by=%s", action, adminULID)
	}
}

// apiKeyToPublicInfo converts an APIKey to public info.
func apiKeyToPublicInfo(apiKey *auth.APIKey) APIKeyPublicInfo {
	info := APIKeyPublicInfo{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Description: apiKey.Description,
		Role:        apiKey.Role,
		CanWrite:    apiKey.CanWrite,
		CreatedAt:   apiKey.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if apiKey.LastUsedAt != nil {
		lastUsed := apiKey.LastUsedAt.Format("2006-01-02T15:04:05Z")
		info.LastUsedAt = &lastUsed
	}

	return info
}

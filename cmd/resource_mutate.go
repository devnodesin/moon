package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ResourceMutateHandler implements POST /data/{resource}:mutate.
type ResourceMutateHandler struct {
	db       DatabaseAdapter
	registry *SchemaRegistry
	cfg      *AppConfig
	jtiStore *JTIRevocationStore
	prefix   string
}

// NewResourceMutateHandler creates a ResourceMutateHandler with the given dependencies.
func NewResourceMutateHandler(db DatabaseAdapter, registry *SchemaRegistry, cfg *AppConfig, jtiStore *JTIRevocationStore) *ResourceMutateHandler {
	return &ResourceMutateHandler{
		db:       db,
		registry: registry,
		cfg:      cfg,
		jtiStore: jtiStore,
		prefix:   strings.TrimRight(cfg.Server.Prefix, "/"),
	}
}

// resourceMutateRequest is the JSON body for POST /data/{resource}:mutate.
type resourceMutateRequest struct {
	Op     string            `json:"op"`
	Data   []json.RawMessage `json:"data"`
	Action string            `json:"action,omitempty"`
}

// HandleMutate handles POST /data/{resource}:mutate requests.
func (h *ResourceMutateHandler) HandleMutate(w http.ResponseWriter, r *http.Request) {
	resource := extractResource(r.URL.Path)
	if resource == "" {
		WriteError(w, http.StatusBadRequest, "Missing resource name")
		return
	}

	col, ok := h.registry.Get(resource)
	if !ok {
		WriteError(w, http.StatusNotFound, fmt.Sprintf("Resource '%s' not found", resource))
		return
	}

	identity, ok := GetAuthIdentity(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.authorize(resource, identity); err != nil {
		WriteError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var req resourceMutateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
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

	if len(req.Data) == 0 {
		WriteError(w, http.StatusBadRequest, "Data must not be empty")
		return
	}

	switch req.Op {
	case "create":
		h.handleCreate(w, r, resource, col, req.Data)
	case "update":
		h.handleUpdate(w, r, resource, col, req.Data)
	case "destroy":
		h.handleDestroy(w, r, resource, col, req.Data)
	case "action":
		h.handleAction(w, r, resource, col, req)
	default:
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Unknown op: %s", req.Op))
	}
}

// authorize checks authorization for mutate operations.
func (h *ResourceMutateHandler) authorize(resource string, identity *AuthIdentity) error {
	if resource == "users" || resource == "apikeys" {
		if identity.Role != "admin" {
			return fmt.Errorf("forbidden")
		}
		return nil
	}
	if identity.Role != "admin" && !identity.CanWrite {
		return fmt.Errorf("forbidden")
	}
	return nil
}

// ---------------------------------------------------------------------------
// op=create
// ---------------------------------------------------------------------------

func (h *ResourceMutateHandler) handleCreate(w http.ResponseWriter, _ *http.Request, resource string, col *Collection, rawItems []json.RawMessage) {
	ctx := context.Background()
	fieldMap := buildFieldMap(col)

	var results []any
	failed := 0

	for _, raw := range rawItems {
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid create item")
			return
		}

		if _, hasID := item["id"]; hasID {
			WriteError(w, http.StatusBadRequest, "Field 'id' must not be provided for create")
			return
		}

		if err := validateWritableFields(item, col, resource); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := validateFieldsExist(item, fieldMap, resource); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := validateFieldTypes(item, fieldMap); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		var record map[string]any
		var insertErr error

		switch resource {
		case "users":
			record, insertErr = h.createUser(ctx, item)
		case "apikeys":
			record, insertErr = h.createAPIKey(ctx, item)
		default:
			record, insertErr = h.createDynamic(ctx, resource, item, col)
		}

		if insertErr != nil {
			if ve, ok := insertErr.(*validationError); ok {
				WriteError(w, http.StatusBadRequest, ve.msg)
				return
			}
			if isUniqueViolation(insertErr) {
				WriteError(w, http.StatusConflict, uniqueViolationMessage(insertErr))
				return
			}
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		results = append(results, record)
	}

	status := http.StatusCreated
	if len(results) == 0 {
		status = http.StatusOK
	}
	meta := map[string]any{"success": len(results), "failed": failed}
	WriteSuccessFull(w, status, "Resource created successfully", results, meta, nil)
}

func (h *ResourceMutateHandler) createUser(ctx context.Context, item map[string]any) (map[string]any, error) {
	username, _ := item["username"].(string)
	email, _ := item["email"].(string)
	password, _ := item["password"].(string)
	role, _ := item["role"].(string)

	if username == "" {
		return nil, &validationError{msg: "Field 'username' is required"}
	}
	if email == "" {
		return nil, &validationError{msg: "Field 'email' is required"}
	}
	if password == "" {
		return nil, &validationError{msg: "Field 'password' is required"}
	}
	if role == "" {
		return nil, &validationError{msg: "Field 'role' is required"}
	}
	if role != "admin" && role != "user" {
		return nil, &validationError{msg: "Field 'role' must be 'admin' or 'user'"}
	}

	if err := validatePasswordPolicy(password); err != nil {
		return nil, &validationError{msg: fmt.Sprintf("Password policy violation: %s", err.Error())}
	}

	if !isValidEmail(email) {
		return nil, &validationError{msg: "Invalid email address"}
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	canWrite := false
	if v, ok := item["can_write"]; ok {
		canWrite = toBool(v)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := GenerateULID()
	row := map[string]any{
		"id":            id,
		"username":      strings.ToLower(username),
		"email":         strings.ToLower(email),
		"password_hash": hash,
		"role":          role,
		"can_write":     boolToInt(canWrite),
		"created_at":    now,
		"updated_at":    now,
	}

	if err := h.db.InsertRow(ctx, "users", row); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":         id,
		"username":   row["username"],
		"email":      row["email"],
		"role":       role,
		"can_write":  canWrite,
		"created_at": now,
		"updated_at": now,
	}, nil
}

func (h *ResourceMutateHandler) createAPIKey(ctx context.Context, item map[string]any) (map[string]any, error) {
	name, _ := item["name"].(string)
	role, _ := item["role"].(string)

	if name == "" {
		return nil, &validationError{msg: "Field 'name' is required"}
	}
	if role == "" {
		return nil, &validationError{msg: "Field 'role' is required"}
	}
	if role != "admin" && role != "user" {
		return nil, &validationError{msg: "Field 'role' must be 'admin' or 'user'"}
	}

	isWebsiteRaw, ok := item["is_website"]
	if !ok {
		return nil, &validationError{msg: "Field 'is_website' is required"}
	}
	isWebsite, ok := isWebsiteRaw.(bool)
	if !ok {
		return nil, &validationError{msg: "Field 'is_website' must be a boolean"}
	}

	canWrite := false
	if v, ok := item["can_write"]; ok {
		canWrite = toBool(v)
	}

	collections, err := validateCollections(item["collections"], true)
	if err != nil {
		return nil, &validationError{msg: err.Error()}
	}

	allowedOrigins, err := validateAllowedOrigins(item["allowed_origins"])
	if err != nil {
		return nil, &validationError{msg: err.Error()}
	}

	rateLimit := DefaultAPIKeyRateLimit
	if value, ok := item["rate_limit"]; ok {
		rateLimit, err = validatePositiveInteger("rate_limit", value)
		if err != nil {
			return nil, &validationError{msg: err.Error()}
		}
	}

	captchaRequired := false
	if value, ok := item["captcha_required"]; ok {
		captchaRequired = toBool(value)
	}

	enabled := true
	if value, ok := item["enabled"]; ok {
		enabled = toBool(value)
	}

	rawKey, keyHash := GenerateAPIKey()
	now := time.Now().UTC().Format(time.RFC3339)
	id := GenerateULID()
	row := map[string]any{
		"id":               id,
		"name":             name,
		"role":             role,
		"can_write":        boolToInt(canWrite),
		"collections":      prepareValueForDB(collections, MoonFieldTypeJSON),
		"is_website":       boolToInt(isWebsite),
		"allowed_origins":  prepareValueForDB(allowedOrigins, MoonFieldTypeJSON),
		"rate_limit":       int64(rateLimit),
		"captcha_required": boolToInt(captchaRequired),
		"enabled":          boolToInt(enabled),
		"key_hash":         keyHash,
		"created_at":       now,
		"updated_at":       now,
	}

	if err := h.db.InsertRow(ctx, "apikeys", row); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":               id,
		"name":             name,
		"role":             role,
		"can_write":        canWrite,
		"collections":      collections,
		"is_website":       isWebsite,
		"allowed_origins":  allowedOrigins,
		"rate_limit":       int64(rateLimit),
		"captcha_required": captchaRequired,
		"enabled":          enabled,
		"key":              rawKey,
		"created_at":       now,
		"updated_at":       now,
	}, nil
}

func (h *ResourceMutateHandler) createDynamic(ctx context.Context, resource string, item map[string]any, col *Collection) (map[string]any, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	id := GenerateULID()
	row := map[string]any{"id": id}
	for k, v := range item {
		row[k] = prepareValueForDB(v, buildFieldMap(col)[k].Type)
	}

	fieldMap := buildFieldMap(col)
	if _, hasCreated := fieldMap["created_at"]; hasCreated {
		row["created_at"] = now
	}
	if _, hasUpdated := fieldMap["updated_at"]; hasUpdated {
		row["updated_at"] = now
	}

	if err := h.db.InsertRow(ctx, resource, row); err != nil {
		return nil, err
	}

	rows, _, err := h.db.QueryRows(ctx, resource, QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: id}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil || len(rows) == 0 {
		return row, nil
	}

	record := formatRecord(rows[0], col)
	record = filterHiddenFields(resource, record)
	return record, nil
}

// ---------------------------------------------------------------------------
// op=update
// ---------------------------------------------------------------------------

func (h *ResourceMutateHandler) handleUpdate(w http.ResponseWriter, _ *http.Request, resource string, col *Collection, rawItems []json.RawMessage) {
	ctx := context.Background()
	fieldMap := buildFieldMap(col)

	var results []any
	failed := 0

	for _, raw := range rawItems {
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid update item")
			return
		}

		idRaw, hasID := item["id"]
		if !hasID {
			WriteError(w, http.StatusBadRequest, "Each update item must include 'id'")
			return
		}
		id, ok := idRaw.(string)
		if !ok || id == "" {
			WriteError(w, http.StatusBadRequest, "Field 'id' must be a non-empty string")
			return
		}

		updateData := make(map[string]any)
		for k, v := range item {
			if k == "id" {
				continue
			}
			updateData[k] = v
		}

		if err := validateWritableFields(updateData, col, resource); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := validateFieldsExist(updateData, fieldMap, resource); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := validateFieldTypes(updateData, fieldMap); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if resource == "apikeys" {
			if err := validateAPIKeyMutationFields(updateData); err != nil {
				WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		// Check record exists
		existing, _, err := h.db.QueryRows(ctx, resource, QueryOptions{
			Filters: []Filter{{Field: "id", Op: "eq", Value: id}},
			Page:    1,
			PerPage: 1,
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if len(existing) == 0 {
			failed++
			continue
		}

		dbData := make(map[string]any)
		for k, v := range updateData {
			f, fOK := fieldMap[k]
			if fOK {
				dbData[k] = prepareValueForDB(v, f.Type)
			} else {
				dbData[k] = v
			}
		}

		if _, hasUpdated := fieldMap["updated_at"]; hasUpdated {
			dbData["updated_at"] = time.Now().UTC().Format(time.RFC3339)
		}

		if err := h.db.UpdateRow(ctx, resource, id, dbData); err != nil {
			if isUniqueViolation(err) {
				failed++
				continue
			}
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		rows, _, err := h.db.QueryRows(ctx, resource, QueryOptions{
			Filters: []Filter{{Field: "id", Op: "eq", Value: id}},
			Page:    1,
			PerPage: 1,
		})
		if err != nil || len(rows) == 0 {
			failed++
			continue
		}

		record := formatRecord(rows[0], col)
		record = filterHiddenFields(resource, record)
		results = append(results, record)
	}

	meta := map[string]any{"success": len(results), "failed": failed}
	WriteSuccessFull(w, http.StatusOK, "Resource updated successfully", results, meta, nil)
}

// ---------------------------------------------------------------------------
// op=destroy
// ---------------------------------------------------------------------------

func (h *ResourceMutateHandler) handleDestroy(w http.ResponseWriter, _ *http.Request, resource string, col *Collection, rawItems []json.RawMessage) {
	ctx := context.Background()

	failed := 0
	success := 0

	for _, raw := range rawItems {
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid destroy item")
			return
		}

		idRaw, hasID := item["id"]
		if !hasID {
			WriteError(w, http.StatusBadRequest, "Each destroy item must include 'id'")
			return
		}
		id, ok := idRaw.(string)
		if !ok || id == "" {
			WriteError(w, http.StatusBadRequest, "Field 'id' must be a non-empty string")
			return
		}

		// Check record exists
		existing, _, err := h.db.QueryRows(ctx, resource, QueryOptions{
			Filters: []Filter{{Field: "id", Op: "eq", Value: id}},
			Page:    1,
			PerPage: 1,
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if len(existing) == 0 {
			failed++
			continue
		}

		// Last admin protection
		if resource == "users" {
			userRole, _ := existing[0]["role"].(string)
			if userRole == "admin" {
				adminCount, err := h.countAdmins(ctx)
				if err != nil {
					WriteError(w, http.StatusInternalServerError, "Internal server error")
					return
				}
				if adminCount <= 1 {
					failed++
					continue
				}
			}
		}

		// For users, cascade-delete refresh tokens
		if resource == "users" {
			if err := h.cascadeDeleteRefreshTokens(ctx, id); err != nil {
				WriteError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
		}

		if err := h.db.DeleteRow(ctx, resource, id); err != nil {
			failed++
			continue
		}

		success++
	}

	data := make([]any, 0)
	meta := map[string]any{"success": success, "failed": failed}
	WriteSuccessFull(w, http.StatusOK, "Resource destroyed successfully", data, meta, nil)
}

func (h *ResourceMutateHandler) countAdmins(ctx context.Context) (int, error) {
	rows, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "role", Op: "eq", Value: "admin"}},
		Page:    1,
		PerPage: MaxPerPage,
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (h *ResourceMutateHandler) cascadeDeleteRefreshTokens(ctx context.Context, userID string) error {
	rows, _, err := h.db.QueryRows(ctx, "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "user_id", Op: "eq", Value: userID}},
		Page:    1,
		PerPage: MaxPerPage,
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		tokenID, _ := row["id"].(string)
		if tokenID == "" {
			continue
		}
		if err := h.db.DeleteRow(ctx, "moon_auth_refresh_tokens", tokenID); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// op=action
// ---------------------------------------------------------------------------

func (h *ResourceMutateHandler) handleAction(w http.ResponseWriter, _ *http.Request, resource string, col *Collection, req resourceMutateRequest) {
	if req.Action == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: action")
		return
	}

	switch {
	case resource == "users" && req.Action == "reset_password":
		h.actionResetPassword(w, req.Data)
	case resource == "users" && req.Action == "revoke_sessions":
		h.actionRevokeSessions(w, req.Data)
	case resource == "apikeys" && req.Action == "rotate":
		h.actionRotateAPIKey(w, req.Data)
	default:
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported action '%s' for resource '%s'", req.Action, resource))
	}
}

func (h *ResourceMutateHandler) actionResetPassword(w http.ResponseWriter, rawItems []json.RawMessage) {
	ctx := context.Background()
	var results []any
	failed := 0

	for _, raw := range rawItems {
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid action item")
			return
		}

		idRaw, hasID := item["id"]
		if !hasID {
			WriteError(w, http.StatusBadRequest, "Each item must include 'id'")
			return
		}
		id, ok := idRaw.(string)
		if !ok || id == "" {
			WriteError(w, http.StatusBadRequest, "Field 'id' must be a non-empty string")
			return
		}

		passwordRaw, hasPwd := item["password"]
		if !hasPwd {
			WriteError(w, http.StatusBadRequest, "Field 'password' is required for reset_password")
			return
		}
		password, ok := passwordRaw.(string)
		if !ok || password == "" {
			WriteError(w, http.StatusBadRequest, "Field 'password' must be a non-empty string")
			return
		}

		if err := validatePasswordPolicy(password); err != nil {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("Password policy violation: %s", err.Error()))
			return
		}

		// Check user exists
		existing, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
			Filters: []Filter{{Field: "id", Op: "eq", Value: id}},
			Page:    1,
			PerPage: 1,
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if len(existing) == 0 {
			failed++
			continue
		}

		hash, err := HashPassword(password)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		if err := h.db.UpdateRow(ctx, "users", id, map[string]any{
			"password_hash": hash,
			"updated_at":    now,
		}); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Invalidate all refresh tokens
		if err := h.revokeAllRefreshTokens(ctx, id, "password_reset"); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		results = append(results, map[string]any{"id": id})
	}

	meta := map[string]any{"success": len(results), "failed": failed}
	WriteSuccessFull(w, http.StatusOK, "Action completed successfully", results, meta, nil)
}

func (h *ResourceMutateHandler) actionRevokeSessions(w http.ResponseWriter, rawItems []json.RawMessage) {
	ctx := context.Background()
	var results []any
	failed := 0

	for _, raw := range rawItems {
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid action item")
			return
		}

		idRaw, hasID := item["id"]
		if !hasID {
			WriteError(w, http.StatusBadRequest, "Each item must include 'id'")
			return
		}
		id, ok := idRaw.(string)
		if !ok || id == "" {
			WriteError(w, http.StatusBadRequest, "Field 'id' must be a non-empty string")
			return
		}

		// Check user exists
		existing, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
			Filters: []Filter{{Field: "id", Op: "eq", Value: id}},
			Page:    1,
			PerPage: 1,
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if len(existing) == 0 {
			failed++
			continue
		}

		if err := h.revokeAllRefreshTokens(ctx, id, "admin_revoked"); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		results = append(results, map[string]any{"id": id})
	}

	meta := map[string]any{"success": len(results), "failed": failed}
	WriteSuccessFull(w, http.StatusOK, "Action completed successfully", results, meta, nil)
}

func (h *ResourceMutateHandler) actionRotateAPIKey(w http.ResponseWriter, rawItems []json.RawMessage) {
	ctx := context.Background()
	var results []any
	failed := 0

	for _, raw := range rawItems {
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid action item")
			return
		}

		idRaw, hasID := item["id"]
		if !hasID {
			WriteError(w, http.StatusBadRequest, "Each item must include 'id'")
			return
		}
		id, ok := idRaw.(string)
		if !ok || id == "" {
			WriteError(w, http.StatusBadRequest, "Field 'id' must be a non-empty string")
			return
		}

		existing, _, err := h.db.QueryRows(ctx, "apikeys", QueryOptions{
			Filters: []Filter{{Field: "id", Op: "eq", Value: id}},
			Page:    1,
			PerPage: 1,
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if len(existing) == 0 {
			failed++
			continue
		}

		rawKey, keyHash := GenerateAPIKey()
		now := time.Now().UTC().Format(time.RFC3339)

		if err := h.db.UpdateRow(ctx, "apikeys", id, map[string]any{
			"key_hash":   keyHash,
			"updated_at": now,
		}); err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		row := existing[0]
		results = append(results, map[string]any{
			"id":               id,
			"name":             stringVal(row, "name"),
			"role":             stringVal(row, "role"),
			"can_write":        toBool(row["can_write"]),
			"collections":      apiKeyCollectionsValue(row["collections"]),
			"is_website":       toBool(row["is_website"]),
			"allowed_origins":  apiKeyAllowedOriginsValue(row["allowed_origins"]),
			"rate_limit":       int64(apiKeyRateLimitValue(row["rate_limit"])),
			"captcha_required": toBool(row["captcha_required"]),
			"enabled":          apiKeyEnabledValue(row),
			"key":              rawKey,
		})
	}

	meta := map[string]any{"success": len(results), "failed": failed}
	WriteSuccessFull(w, http.StatusOK, "Action completed successfully", results, meta, nil)
}

// revokeAllRefreshTokens revokes all non-revoked refresh tokens for a user.
func (h *ResourceMutateHandler) revokeAllRefreshTokens(ctx context.Context, userID, reason string) error {
	rows, _, err := h.db.QueryRows(ctx, "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "user_id", Op: "eq", Value: userID}},
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

// ---------------------------------------------------------------------------
// API Key generation
// ---------------------------------------------------------------------------

// GenerateAPIKey generates a new API key with the moon_live_ prefix and a
// 64-character base62 suffix. Returns the raw key and its SHA-256 hex hash.
func GenerateAPIKey() (raw string, hash string) {
	const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	suffix := make([]byte, 64)
	for i := range suffix {
		b := make([]byte, 1)
		rand.Read(b)
		suffix[i] = base62Chars[int(b[0])%len(base62Chars)]
	}
	raw = APIKeyPrefix + string(suffix)
	h := sha256.Sum256([]byte(raw))
	hash = fmt.Sprintf("%x", h)
	return raw, hash
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

type validationError struct {
	msg string
}

func (e *validationError) Error() string { return e.msg }

// readonlyFieldsForResource returns the set of fields that are read-only for
// the given resource and must not be set by the client on create or update.
func readonlyFieldsForResource(resource string) map[string]bool {
	if sysFields, ok := systemReadOnlyFields[resource]; ok {
		return sysFields
	}
	return map[string]bool{"id": true}
}

// validateWritableFields rejects writes to read-only or server-owned fields.
func validateWritableFields(item map[string]any, col *Collection, resource string) error {
	readonly := readonlyFieldsForResource(resource)

	// For system resources, also block password/password_hash and key_hash writes
	// through create/update (password is handled as a special input field for users)
	for key := range item {
		if readonly[key] {
			return fmt.Errorf("Field '%s' is read-only", key)
		}
	}
	return nil
}

// validateFieldsExist ensures every field in item exists in the schema.
// For system resources, we also allow special input fields like 'password'.
func validateFieldsExist(item map[string]any, fieldMap map[string]Field, resource string) error {
	for key := range item {
		if _, ok := fieldMap[key]; ok {
			continue
		}
		// Allow "password" as a special input field for users create
		if resource == "users" && key == "password" {
			continue
		}
		return fmt.Errorf("Unknown field '%s'", key)
	}
	return nil
}

// validateFieldTypes checks that each value is type-valid for its field.
func validateFieldTypes(item map[string]any, fieldMap map[string]Field) error {
	for key, value := range item {
		f, ok := fieldMap[key]
		if !ok {
			continue
		}
		if value == nil {
			if !f.Nullable {
				return fmt.Errorf("Field '%s' cannot be null", key)
			}
			continue
		}
		if !isTypeValid(value, f.Type) {
			return fmt.Errorf("Invalid value for field '%s' of type '%s'", key, f.Type)
		}
	}
	return nil
}

func validateAPIKeyMutationFields(item map[string]any) error {
	if _, ok := item["allowed_origins"]; ok {
		if _, err := validateAllowedOrigins(item["allowed_origins"]); err != nil {
			return err
		}
	}
	if _, ok := item["collections"]; ok {
		if _, err := validateCollections(item["collections"], false); err != nil {
			return err
		}
	}

	if value, ok := item["rate_limit"]; ok {
		if _, err := validatePositiveInteger("rate_limit", value); err != nil {
			return err
		}
	}

	if value, ok := item["is_website"]; ok {
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("Field 'is_website' must be a boolean")
		}
	}

	if value, ok := item["captcha_required"]; ok {
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("Field 'captcha_required' must be a boolean")
		}
	}

	if value, ok := item["enabled"]; ok {
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("Field 'enabled' must be a boolean")
		}
	}

	return nil
}

func validateAllowedOrigins(value any) ([]string, error) {
	return validateStringArrayField("allowed_origins", value, false)
}

func validateCollections(value any, required bool) ([]string, error) {
	return validateStringArrayField("collections", value, required)
}

func validateStringArrayField(field string, value any, required bool) ([]string, error) {
	if value == nil {
		if required {
			return nil, fmt.Errorf("Field '%s' is required", field)
		}
		return nil, nil
	}

	switch v := value.(type) {
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("Field '%s' must contain only strings", field)
			}
			result = append(result, s)
		}
		return result, nil
	case []string:
		result := make([]string, len(v))
		copy(result, v)
		return result, nil
	default:
		return nil, fmt.Errorf("Field '%s' must be an array of strings", field)
	}
}

func validatePositiveInteger(field string, value any) (int, error) {
	switch v := value.(type) {
	case float64:
		if v != math.Trunc(v) || v < 1 {
			return 0, fmt.Errorf("Field '%s' must be a positive integer", field)
		}
		return int(v), nil
	case int:
		if v < 1 {
			return 0, fmt.Errorf("Field '%s' must be a positive integer", field)
		}
		return v, nil
	case int64:
		if v < 1 {
			return 0, fmt.Errorf("Field '%s' must be a positive integer", field)
		}
		return int(v), nil
	default:
		return 0, fmt.Errorf("Field '%s' must be a positive integer", field)
	}
}

func apiKeyAllowedOriginsValue(value any) []string {
	allowedOrigins, err := parseAllowedOrigins(value)
	if err != nil {
		return nil
	}
	return allowedOrigins
}

func apiKeyCollectionsValue(value any) []string {
	collections, err := parseCollections(value)
	if err != nil {
		return nil
	}
	return collections
}

func apiKeyRateLimitValue(value any) int {
	rateLimit, err := parseAPIKeyRateLimit(value)
	if err != nil {
		return DefaultAPIKeyRateLimit
	}
	return rateLimit
}

func apiKeyEnabledValue(row map[string]any) bool {
	value, ok := row["enabled"]
	if !ok {
		return true
	}
	return toBool(value)
}

// isTypeValid checks if a JSON value is compatible with the given Moon field type.
func isTypeValid(value any, fieldType string) bool {
	switch fieldType {
	case MoonFieldTypeString:
		_, ok := value.(string)
		return ok
	case MoonFieldTypeInteger:
		switch v := value.(type) {
		case float64:
			return v == math.Trunc(v)
		case int:
			return true
		case int64:
			return true
		default:
			return false
		}
	case MoonFieldTypeDecimal:
		switch value.(type) {
		case string:
			return true
		case float64:
			return true
		default:
			return false
		}
	case MoonFieldTypeBoolean:
		_, ok := value.(bool)
		return ok
	case MoonFieldTypeDatetime:
		s, ok := value.(string)
		if !ok {
			return false
		}
		_, err := time.Parse(time.RFC3339, s)
		return err == nil
	case MoonFieldTypeJSON:
		switch value.(type) {
		case map[string]any:
			return true
		case []any:
			return true
		default:
			return false
		}
	default:
		return true
	}
}

// prepareValueForDB converts a JSON value to the appropriate database storage format.
func prepareValueForDB(value any, fieldType string) any {
	if value == nil {
		return nil
	}
	switch fieldType {
	case MoonFieldTypeBoolean:
		if toBool(value) {
			return int64(1)
		}
		return int64(0)
	case MoonFieldTypeJSON:
		b, err := json.Marshal(value)
		if err != nil {
			return value
		}
		return string(b)
	default:
		return value
	}
}

// boolToInt converts a bool to 0 or 1 for SQLite storage.
func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// isUniqueViolation checks if an error indicates a unique constraint violation.
func isUniqueViolation(err error) bool {
	for _, msg := range errorMessages(err) {
		if strings.Contains(msg, "UNIQUE constraint failed") ||
			strings.Contains(msg, "unique constraint") ||
			strings.Contains(msg, "duplicate key") {
			return true
		}
	}
	return false
}

// postgresUniqueFieldsRe extracts field names from PostgreSQL duplicate key errors.
var postgresUniqueFieldsRe = regexp.MustCompile(`Key \(([^)]+)\)=`)

const uniqueFieldNameTrimCutset = "\"'`"

func uniqueViolationMessage(err error) string {
	fields := uniqueViolationFields(err)
	switch len(fields) {
	case 0:
		return "Unique constraint violation"
	case 1:
		return fmt.Sprintf("Unique constraint violation for field: %s", fields[0])
	default:
		return fmt.Sprintf("Unique constraint violation for fields: %s", strings.Join(fields, ", "))
	}
}

func uniqueViolationFields(err error) []string {
	const sqlitePrefix = "UNIQUE constraint failed:"

	for _, msg := range errorMessages(err) {
		if idx := strings.Index(msg, sqlitePrefix); idx >= 0 {
			fields := parseUniqueFieldList(msg[idx+len(sqlitePrefix):])
			if len(fields) > 0 {
				return fields
			}
		}

		matches := postgresUniqueFieldsRe.FindStringSubmatch(msg)
		if len(matches) == 2 {
			fields := parseUniqueFieldList(matches[1])
			if len(fields) > 0 {
				return fields
			}
		}
	}

	return nil
}

func parseUniqueFieldList(raw string) []string {
	parts := strings.Split(raw, ",")
	fields := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		field := strings.TrimSpace(part)
		if dot := strings.LastIndex(field, "."); dot >= 0 {
			field = field[dot+1:]
		}
		field = strings.Trim(field, uniqueFieldNameTrimCutset)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	return fields
}

func errorMessages(err error) []string {
	if err == nil {
		return nil
	}

	var messages []string
	for current := err; current != nil; current = errors.Unwrap(current) {
		messages = append(messages, current.Error())
	}
	return messages
}

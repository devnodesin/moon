package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// setupAuthMeTest creates a fresh in-memory DB, system tables, a test user,
// and returns the handler, a valid JWT, and the DB adapter.
func setupAuthMeTest(t *testing.T) (*AuthMeHandler, string, DatabaseAdapter) {
	t.Helper()

	db, err := NewSQLiteAdapter(DatabaseConfig{
		Connection:         "sqlite",
		Database:           ":memory:",
		QueryTimeout:       30,
		SlowQueryThreshold: 500,
	}, NewTestLogger(&bytes.Buffer{}))
	if err != nil {
		t.Fatalf("create db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	if err := EnsureSystemTables(ctx, db); err != nil {
		t.Fatalf("ensure system tables: %v", err)
	}

	cfg := &AppConfig{
		JWTSecret:        "this-is-a-test-secret-that-is-long-enough",
		JWTAccessExpiry:  3600,
		JWTRefreshExpiry: 604800,
	}

	hash, err := HashPassword("TestPass1")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	userID := "01TESTUSER000000000000001"
	if err := db.InsertRow(ctx, "users", map[string]any{
		"id":            userID,
		"username":      "testuser",
		"email":         "test@example.com",
		"password_hash": hash,
		"role":          "admin",
		"can_write":     int64(1),
		"created_at":    now,
		"updated_at":    now,
	}); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	jti := GenerateULID()
	token, _, err := CreateAccessToken(userID, jti, "admin", true, cfg.JWTSecret, cfg.JWTAccessExpiry)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	handler := NewAuthMeHandler(db, cfg)
	return handler, token, db
}

// reqWithJWT creates a request with the auth identity set in context.
func reqWithJWT(method, path string, body []byte, userID, role string, canWrite bool) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       userID,
		Role:           role,
		CanWrite:       canWrite,
		JTI:            GenerateULID(),
	}
	ctx := SetAuthIdentity(req.Context(), identity)
	return req.WithContext(ctx)
}

func reqWithAPIKey(method, path string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	identity := &AuthIdentity{
		CredentialType: CredentialTypeAPIKey,
		CallerID:       "01TESTAPIKEY0000000000001",
		Role:           "admin",
		CanWrite:       true,
	}
	ctx := SetAuthIdentity(req.Context(), identity)
	return req.WithContext(ctx)
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

// ---------------------------------------------------------------------------
// GET /auth:me tests
// ---------------------------------------------------------------------------

func TestGetMe_Success(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	req := reqWithJWT("GET", "/auth:me", nil, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.GetMe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeResponse(t, w)
	if resp["message"] != "Current user retrieved successfully" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}

	data, ok := resp["data"].([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("expected 1 data item, got %v", resp["data"])
	}

	user, ok := data[0].(map[string]any)
	if !ok {
		t.Fatal("data[0] is not a map")
	}

	if user["id"] != "01TESTUSER000000000000001" {
		t.Fatalf("unexpected id: %v", user["id"])
	}
	if user["username"] != "testuser" {
		t.Fatalf("unexpected username: %v", user["username"])
	}
	if user["email"] != "test@example.com" {
		t.Fatalf("unexpected email: %v", user["email"])
	}
	if user["role"] != "admin" {
		t.Fatalf("unexpected role: %v", user["role"])
	}
	if user["can_write"] != true {
		t.Fatalf("unexpected can_write: %v", user["can_write"])
	}
	// password_hash must not be present
	if _, exists := user["password_hash"]; exists {
		t.Fatal("password_hash should not be in response")
	}
}

func TestGetMe_NoIdentity(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	req := httptest.NewRequest("GET", "/auth:me", nil)
	w := httptest.NewRecorder()
	handler.GetMe(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetMe_APIKeyRejected(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	req := reqWithAPIKey("GET", "/auth:me", nil)
	w := httptest.NewRecorder()
	handler.GetMe(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetMe_UserNotFound(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	req := reqWithJWT("GET", "/auth:me", nil, "01NONEXISTENT00000000000", "admin", true)
	w := httptest.NewRecorder()
	handler.GetMe(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// POST /auth:me — email update tests
// ---------------------------------------------------------------------------

func TestUpdateMe_EmailSuccess(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"email": "new@example.com"},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeResponse(t, w)
	if resp["message"] != "Current user updated successfully" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}

	data := resp["data"].([]any)
	user := data[0].(map[string]any)
	if user["email"] != "new@example.com" {
		t.Fatalf("expected email new@example.com, got %v", user["email"])
	}
}

func TestUpdateMe_EmailNormalizesLowercase(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"email": "NEW@Example.COM"},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	data := decodeResponse(t, w)["data"].([]any)
	user := data[0].(map[string]any)
	if user["email"] != "new@example.com" {
		t.Fatalf("expected normalized email, got %v", user["email"])
	}
}

func TestUpdateMe_EmailInvalid(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"email": "not-an-email"},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_EmailAlreadyInUse(t *testing.T) {
	handler, _, db := setupAuthMeTest(t)

	// Insert a second user with a different email.
	hash, _ := HashPassword("TestPass1")
	now := time.Now().UTC().Format(time.RFC3339)
	_ = db.InsertRow(context.Background(), "users", map[string]any{
		"id":            "01TESTUSER000000000000002",
		"username":      "otheruser",
		"email":         "other@example.com",
		"password_hash": hash,
		"role":          "user",
		"can_write":     int64(0),
		"created_at":    now,
		"updated_at":    now,
	})

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"email": "other@example.com"},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateMe_EmailSameAsCurrent(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"email": "test@example.com"},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	// Setting email to same value should succeed (no uniqueness conflict with self).
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// POST /auth:me — password update tests
// ---------------------------------------------------------------------------

func TestUpdateMe_PasswordSuccess(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"old_password": "TestPass1",
			"password":     "NewPass123",
		},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeResponse(t, w)
	if resp["message"] != "Password updated successfully. Sign in again." {
		t.Fatalf("unexpected message: %v", resp["message"])
	}
}

func TestUpdateMe_PasswordMissingOld(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"password": "NewPass123",
		},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_PasswordOldIncorrect(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"old_password": "WrongPass1",
			"password":     "NewPass123",
		},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUpdateMe_PasswordPolicyViolation(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"old_password": "TestPass1",
			"password":     "short",
		},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_PasswordRevokesRefreshTokens(t *testing.T) {
	handler, _, db := setupAuthMeTest(t)

	ctx := context.Background()
	// Insert a refresh token for the user.
	now := time.Now().UTC().Format(time.RFC3339)
	tokenID := GenerateULID()
	_ = db.InsertRow(ctx, "moon_auth_refresh_tokens", map[string]any{
		"id":                 tokenID,
		"user_id":            "01TESTUSER000000000000001",
		"refresh_token_hash": "somehash123",
		"expires_at":         time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		"created_at":         now,
	})

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"old_password": "TestPass1",
			"password":     "NewPass123",
		},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the refresh token was revoked.
	rows, _, err := db.QueryRows(ctx, "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: tokenID}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil || len(rows) == 0 {
		t.Fatal("failed to find refresh token")
	}
	if rows[0]["revoked_at"] == nil {
		t.Fatal("expected refresh token to be revoked")
	}
	if rows[0]["revocation_reason"] != "password_changed" {
		t.Fatalf("expected revocation_reason password_changed, got %v", rows[0]["revocation_reason"])
	}
}

// ---------------------------------------------------------------------------
// POST /auth:me — validation tests
// ---------------------------------------------------------------------------

func TestUpdateMe_APIKeyRejected(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"email": "new@example.com"},
	})
	req := reqWithAPIKey("POST", "/auth:me", body)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUpdateMe_NoIdentity(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"email": "new@example.com"},
	})
	req := httptest.NewRequest("POST", "/auth:me", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUpdateMe_InvalidJSON(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	req := reqWithJWT("POST", "/auth:me", []byte("not json"), "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_MissingData(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_EmptyData(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_NonWritableField(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	fields := []string{"id", "username", "role", "can_write", "created_at", "updated_at", "last_login_at", "password_hash"}
	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			body, _ := json.Marshal(map[string]any{
				"data": map[string]any{field: "some_value"},
			})
			req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
			w := httptest.NewRecorder()
			handler.UpdateMe(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("field %q: expected 400, got %d: %s", field, w.Code, w.Body.String())
			}

			resp := decodeResponse(t, w)
			msg, _ := resp["message"].(string)
			if msg == "" || !strings.Contains(msg, field) {
				t.Fatalf("field %q: expected error message to contain field name, got %q", field, msg)
			}
		})
	}
}

func TestUpdateMe_NoEmailOrPassword(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"old_password": "TestPass1"},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Additional edge case tests for UpdateMe coverage
// ---------------------------------------------------------------------------

func TestUpdateMe_PasswordInvalidType(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"password":     12345,
			"old_password": "TestPass1",
		},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_OldPasswordInvalidType(t *testing.T) {
	handler, _, _ := setupAuthMeTest(t)

	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"password":     "NewValidPass1",
			"old_password": 12345,
		},
	})
	req := reqWithJWT("POST", "/auth:me", body, "01TESTUSER000000000000001", "admin", true)
	w := httptest.NewRecorder()
	handler.UpdateMe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRefresh_InvalidTokenType(t *testing.T) {
	handler, _ := setupAuthTest(t)
	// Send refresh_token as a non-string
	body := map[string]any{
		"op":   "refresh",
		"data": map[string]any{"refresh_token": 12345},
	}
	w := doAuthRequest(t, handler, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupAuthTest(t *testing.T) (*AuthSessionHandler, DatabaseAdapter) {
	t.Helper()
	db, err := NewSQLiteAdapter(DatabaseConfig{
		Connection:         "sqlite",
		Database:           ":memory:",
		QueryTimeout:       30,
		SlowQueryThreshold: 500,
	}, NewTestLogger(&bytes.Buffer{}))
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}

	ctx := context.Background()
	if err := EnsureSystemTables(ctx, db); err != nil {
		t.Fatalf("failed to create tables: %v", err)
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
	if err := db.InsertRow(ctx, "users", map[string]any{
		"id":            "01TESTUSER000000000000001",
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

	handler := &AuthSessionHandler{db: db, cfg: cfg}
	t.Cleanup(func() { db.Close() })
	return handler, db
}

func doAuthRequest(t *testing.T, handler *AuthSessionHandler, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth:session", bytes.NewReader(b))
	w := httptest.NewRecorder()
	handler.HandleSession(w, req)
	return w
}

// --- Login tests ---

func TestLogin_Success(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": "TestPass1"},
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp SuccessResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Message != "Login successful" {
		t.Fatalf("expected 'Login successful', got %q", resp.Message)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data item, got %d", len(resp.Data))
	}

	payload, ok := resp.Data[0].(map[string]any)
	if !ok {
		t.Fatal("data[0] is not a map")
	}
	if payload["access_token"] == nil || payload["access_token"] == "" {
		t.Fatal("missing access_token")
	}
	if payload["refresh_token"] == nil || payload["refresh_token"] == "" {
		t.Fatal("missing refresh_token")
	}
	if payload["token_type"] != "Bearer" {
		t.Fatalf("expected Bearer, got %v", payload["token_type"])
	}
	if payload["expires_at"] == nil {
		t.Fatal("missing expires_at")
	}

	user, ok := payload["user"].(map[string]any)
	if !ok {
		t.Fatal("missing user object")
	}
	if user["username"] != "testuser" {
		t.Fatalf("expected testuser, got %v", user["username"])
	}
	if user["role"] != "admin" {
		t.Fatalf("expected admin, got %v", user["role"])
	}
}

func TestLogin_CaseInsensitiveUsername(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "TestUser", "password": "TestPass1"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": "WrongPass1"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "nobody", "password": "TestPass1"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_MissingUsername(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"password": "TestPass1"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Validation tests ---

func TestMissingOp(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"data": map[string]any{"username": "testuser"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUnknownOp(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "destroy",
		"data": map[string]any{},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMissingData(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op": "login",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestInvalidJSON(t *testing.T) {
	handler, _ := setupAuthTest(t)
	req := httptest.NewRequest(http.MethodPost, "/auth:session", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	handler.HandleSession(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Refresh tests ---

func TestRefresh_Success(t *testing.T) {
	handler, _ := setupAuthTest(t)

	// Login first
	loginW := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": "TestPass1"},
	})
	if loginW.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", loginW.Code)
	}

	var loginResp SuccessResponse
	json.NewDecoder(loginW.Body).Decode(&loginResp)
	payload := loginResp.Data[0].(map[string]any)
	refreshToken := payload["refresh_token"].(string)

	// Refresh
	refreshW := doAuthRequest(t, handler, map[string]any{
		"op":   "refresh",
		"data": map[string]any{"refresh_token": refreshToken},
	})
	if refreshW.Code != http.StatusOK {
		t.Fatalf("refresh: expected 200, got %d: %s", refreshW.Code, refreshW.Body.String())
	}

	var refreshResp SuccessResponse
	json.NewDecoder(refreshW.Body).Decode(&refreshResp)
	if refreshResp.Message != "Token refreshed successfully" {
		t.Fatalf("expected 'Token refreshed successfully', got %q", refreshResp.Message)
	}

	newPayload := refreshResp.Data[0].(map[string]any)
	if newPayload["access_token"] == payload["access_token"] {
		t.Fatal("expected new access token")
	}
	if newPayload["refresh_token"] == refreshToken {
		t.Fatal("expected new refresh token (rotation)")
	}
}

func TestRefresh_ReuseRevoked(t *testing.T) {
	handler, _ := setupAuthTest(t)

	// Login
	loginW := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": "TestPass1"},
	})
	var loginResp SuccessResponse
	json.NewDecoder(loginW.Body).Decode(&loginResp)
	refreshToken := loginResp.Data[0].(map[string]any)["refresh_token"].(string)

	// First refresh (succeeds, revokes old token)
	doAuthRequest(t, handler, map[string]any{
		"op":   "refresh",
		"data": map[string]any{"refresh_token": refreshToken},
	})

	// Second refresh with same token (should fail - already revoked)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "refresh",
		"data": map[string]any{"refresh_token": refreshToken},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "refresh",
		"data": map[string]any{"refresh_token": "invalid-token-value"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRefresh_MissingToken(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "refresh",
		"data": map[string]any{},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Logout tests ---

func TestLogout_Success(t *testing.T) {
	handler, _ := setupAuthTest(t)

	// Login
	loginW := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": "TestPass1"},
	})
	var loginResp SuccessResponse
	json.NewDecoder(loginW.Body).Decode(&loginResp)
	refreshToken := loginResp.Data[0].(map[string]any)["refresh_token"].(string)

	// Logout
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "logout",
		"data": map[string]any{"refresh_token": refreshToken},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] != "Logged out successfully" {
		t.Fatalf("expected 'Logged out successfully', got %v", resp["message"])
	}

	// Refresh with revoked token should fail
	refreshW := doAuthRequest(t, handler, map[string]any{
		"op":   "refresh",
		"data": map[string]any{"refresh_token": refreshToken},
	})
	if refreshW.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", refreshW.Code)
	}
}

func TestLogout_Idempotent(t *testing.T) {
	handler, _ := setupAuthTest(t)

	// Logout with unknown token — should still return 200
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "logout",
		"data": map[string]any{"refresh_token": "nonexistent-token"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogout_MissingToken(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "logout",
		"data": map[string]any{},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Rate limiting tests ---

// setupAuthTestWithRateLimiter creates a handler with a real RateLimiter attached.
func setupAuthTestWithRateLimiter(t *testing.T) *AuthSessionHandler {
	t.Helper()
	handler, _ := setupAuthTest(t)
	handler.rateLimiter = NewRateLimiter()
	handler.logger = middlewareTestLogger()
	return handler
}

func TestLogin_RateLimit_BlocksAfterLimit(t *testing.T) {
	handler := setupAuthTestWithRateLimiter(t)

	// Make RateLoginFailureLimit failed attempts (each returns 401).
	for i := range RateLoginFailureLimit {
		w := doAuthRequest(t, handler, map[string]any{
			"op":   "login",
			"data": map[string]any{"username": "testuser", "password": "WrongPass"},
		})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, w.Code)
		}
	}

	// The next attempt must be rate-limited (429).
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": "WrongPass"},
	})
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after %d failures, got %d", RateLoginFailureLimit, w.Code)
	}
}

func TestLogin_RateLimit_ResetOnSuccess(t *testing.T) {
	handler := setupAuthTestWithRateLimiter(t)

	// Exhaust failures (one less than the limit so the correct password still works).
	for range RateLoginFailureLimit - 1 {
		doAuthRequest(t, handler, map[string]any{
			"op":   "login",
			"data": map[string]any{"username": "testuser", "password": "WrongPass"},
		})
	}

	// Successful login must reset the counter.
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": "TestPass1"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("successful login failed: %d %s", w.Code, w.Body.String())
	}

	// Another failed attempt should be allowed (counter was reset).
	w = doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": "WrongPass"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after reset, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// toBool helper tests
// ---------------------------------------------------------------------------

func TestToBool(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"int64 nonzero", int64(1), true},
		{"int64 zero", int64(0), false},
		{"float64 nonzero", float64(1.5), true},
		{"float64 zero", float64(0), false},
		{"int nonzero", int(3), true},
		{"int zero", int(0), false},
		{"string 1", "1", true},
		{"string true", "true", true},
		{"string 0", "0", false},
		{"string false", "false", false},
		{"string other", "yes", false},
		{"nil", nil, false},
		{"struct", struct{}{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toBool(tt.input)
			if got != tt.want {
				t.Errorf("toBool(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Additional auth_session tests for uncovered paths
// ---------------------------------------------------------------------------

func TestLogin_InvalidUsernameType(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": 12345, "password": "TestPass1"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_InvalidPasswordType(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "login",
		"data": map[string]any{"username": "testuser", "password": true},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_Expired_Refresh(t *testing.T) {
	handler, adapter := setupAuthTest(t)

	// Insert an expired refresh token
	expiredAt := "2020-01-01T00:00:00Z"
	hash := HashRefreshToken("test-expired-token")
	_ = adapter.InsertRow(context.Background(), "moon_auth_refresh_tokens", map[string]any{
		"id":                 GenerateULID(),
		"user_id":            "01TESTUSER000000000000001",
		"refresh_token_hash": hash,
		"expires_at":         expiredAt,
		"created_at":         expiredAt,
	})

	w := doAuthRequest(t, handler, map[string]any{
		"op":   "refresh",
		"data": map[string]any{"refresh_token": "test-expired-token"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", w.Code)
	}
}

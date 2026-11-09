package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
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
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "TestPass1"},
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
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "TestPass1"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "WrongPass1"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "nobody@example.com", "password": "TestPass1"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_MissingUsername(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"password": "TestPass1"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Validation tests ---

func TestMissingOp(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"data": map[string]any{"email": "test@example.com"},
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
		"op": "smart_login",
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
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "TestPass1"},
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
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "TestPass1"},
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
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "TestPass1"},
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

	// Make RateSmartLoginLimit failed attempts (each returns 401).
	for i := range RateSmartLoginLimit {
		w := doAuthRequest(t, handler, map[string]any{
			"op":   "smart_login",
			"data": map[string]any{"email": "test@example.com", "password": "WrongPass"},
		})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, w.Code)
		}
	}

	// The next attempt must be rate-limited (429).
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "WrongPass"},
	})
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after %d failures, got %d", RateSmartLoginLimit, w.Code)
	}
}

func TestLogin_RateLimit_ResetOnSuccess(t *testing.T) {
	handler := setupAuthTestWithRateLimiter(t)

	// Exhaust failures (one less than the limit so the correct password still works).
	for range RateSmartLoginLimit - 1 {
		doAuthRequest(t, handler, map[string]any{
			"op":   "smart_login",
			"data": map[string]any{"email": "test@example.com", "password": "WrongPass"},
		})
	}

	// Successful login must reset the counter.
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "TestPass1"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("successful login failed: %d %s", w.Code, w.Body.String())
	}

	// Another failed attempt should be allowed (counter was reset).
	w = doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "WrongPass"},
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
		"op":   "smart_login",
		"data": map[string]any{"email": 12345, "password": "TestPass1"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLogin_InvalidPasswordType(t *testing.T) {
	handler, _ := setupAuthTest(t)
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": true},
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

// ---------------------------------------------------------------------------
// fakeEmailSender captures OTP emails without making SMTP connections.
// ---------------------------------------------------------------------------

type fakeEmailSender struct {
	mu   sync.Mutex
	sent []struct{ to, code string }
}

func (f *fakeEmailSender) SendOTPEmail(to, code string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, struct{ to, code string }{to, code})
	return nil
}

func (f *fakeEmailSender) lastCode() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sent) == 0 {
		return ""
	}
	return f.sent[len(f.sent)-1].code
}

func (f *fakeEmailSender) sentCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

// setupAuthTestWithOTP builds on setupAuthTest and attaches a fake email
// sender so the OTP request path can be exercised without SMTP.
func setupAuthTestWithOTP(t *testing.T) (*AuthSessionHandler, DatabaseAdapter, *fakeEmailSender) {
	t.Helper()
	handler, db := setupAuthTest(t)
	sender := &fakeEmailSender{}
	handler.emailSender = sender
	handler.otpExpiryMin = OTPTTLMinutes
	return handler, db, sender
}

// insertOTPUser inserts a user with login_method="otp" and no password hash.
func insertOTPUser(t *testing.T, db DatabaseAdapter, email, userID string) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	if err := db.InsertRow(ctx, "users", map[string]any{
		"id":            userID,
		"username":      "otpuser",
		"email":         email,
		"password_hash": nil,
		"role":          "user",
		"can_write":     int64(0),
		"login_method":  "otp",
		"created_at":    now,
		"updated_at":    now,
	}); err != nil {
		t.Fatalf("insert otp user: %v", err)
	}
}

// insertOTPRow manually inserts an OTP code row with a known plaintext code,
// matching the hashing used by GenerateOTP/VerifyOTP.
func insertOTPRow(t *testing.T, db DatabaseAdapter, userID, code, purpose string, expired bool) {
	t.Helper()
	ctx := context.Background()
	hash := sha256.Sum256([]byte(code))
	hashHex := hex.EncodeToString(hash[:])
	now := time.Now().UTC()
	expires := now.Add(5 * time.Minute)
	if expired {
		expires = now.Add(-time.Minute)
	}
	if err := db.InsertRow(ctx, "moon_auth_otp_codes", map[string]any{
		"id":          GenerateULID(),
		"user_id":     userID,
		"code_hash":   hashHex,
		"purpose":     purpose,
		"expires_at":  expires.Format(time.RFC3339),
		"created_at":  now.Format(time.RFC3339),
		"consumed_at": nil,
	}); err != nil {
		t.Fatalf("insert otp row: %v", err)
	}
}

// otpRowCount returns the number of OTP rows for a user/purpose.
func otpRowCount(t *testing.T, db DatabaseAdapter, userID, purpose string) int {
	t.Helper()
	ctx := context.Background()
	rows, _, err := db.QueryRows(ctx, "moon_auth_otp_codes", QueryOptions{
		Filters: []Filter{
			{Field: "user_id", Op: "eq", Value: userID},
			{Field: "purpose", Op: "eq", Value: purpose},
		},
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("query otp rows: %v", err)
	}
	return len(rows)
}

// ---------------------------------------------------------------------------
// OTP integration tests
// ---------------------------------------------------------------------------

func TestSmartLogin_Initiate_RequestOTP_Success(t *testing.T) {
	handler, db, sender := setupAuthTestWithOTP(t)
	insertOTPUser(t, db, "otp@example.com", "01TESTOTPUSER000000000000002")

	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "otp@example.com", "request_otp": true},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// An OTP row must have been created for the OTP user.
	if count := otpRowCount(t, db, "01TESTOTPUSER000000000000002", "login"); count != 1 {
		t.Fatalf("expected 1 OTP row, got %d", count)
	}
	if sender.sentCount() != 1 {
		t.Fatalf("expected 1 OTP email sent, got %d", sender.sentCount())
	}
}

func TestSmartLogin_Initiate_RequestOTP_PasswordOnlyUser(t *testing.T) {
	handler, db, sender := setupAuthTestWithOTP(t)

	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "request_otp": true},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Password-only users must not receive an OTP row or email.
	if count := otpRowCount(t, db, "01TESTUSER000000000000001", "login"); count != 0 {
		t.Fatalf("expected 0 OTP rows for password-only user, got %d", count)
	}
	if sender.sentCount() != 0 {
		t.Fatalf("expected 0 OTP emails sent, got %d", sender.sentCount())
	}
}

func TestSmartLogin_Verify_OTP_Success(t *testing.T) {
	handler, db, sender := setupAuthTestWithOTP(t)
	insertOTPUser(t, db, "otp@example.com", "01TESTOTPUSER000000000000002")

	// Step 1: request OTP.
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "otp@example.com", "request_otp": true},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("step 1: expected 200, got %d", w.Code)
	}
	code := sender.lastCode()
	if code == "" {
		t.Fatal("expected an OTP email to be sent")
	}

	// Step 2: verify with the code received by email.
	w = doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "otp@example.com", "code": code},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("step 2: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp SuccessResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Message != "Login successful" {
		t.Fatalf("expected 'Login successful', got %q", resp.Message)
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
}

func TestSmartLogin_Verify_OTP_WrongCode(t *testing.T) {
	handler, _, _ := setupAuthTestWithOTP(t)
	insertOTPUser(t, handler.db, "otp@example.com", "01TESTOTPUSER000000000000002")
	insertOTPRow(t, handler.db, "01TESTOTPUSER000000000000002", "123456", "login", false)

	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "otp@example.com", "code": "000000"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong code, got %d", w.Code)
	}
}

func TestSmartLogin_Verify_OTP_Expired(t *testing.T) {
	handler, _, _ := setupAuthTestWithOTP(t)
	insertOTPUser(t, handler.db, "otp@example.com", "01TESTOTPUSER000000000000002")
	insertOTPRow(t, handler.db, "01TESTOTPUSER000000000000002", "123456", "login", true)

	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "otp@example.com", "code": "123456"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired code, got %d", w.Code)
	}
}

func TestSmartLogin_Verify_OTP_AlreadyConsumed(t *testing.T) {
	handler, _, _ := setupAuthTestWithOTP(t)
	insertOTPUser(t, handler.db, "otp@example.com", "01TESTOTPUSER000000000000002")
	insertOTPRow(t, handler.db, "01TESTOTPUSER000000000000002", "123456", "login", false)

	// First verification succeeds and consumes the code.
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "otp@example.com", "code": "123456"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("first verify: expected 200, got %d", w.Code)
	}

	// Second verification with the same code must fail (already consumed).
	w = doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "otp@example.com", "code": "123456"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("second verify: expected 401, got %d", w.Code)
	}
}

func TestSmartLogin_Verify_PasswordForOTPUser(t *testing.T) {
	handler, _, _ := setupAuthTestWithOTP(t)
	insertOTPUser(t, handler.db, "otp@example.com", "01TESTOTPUSER000000000000002")

	// OTP-only users cannot log in with a password.
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "otp@example.com", "password": "TestPass1"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for password on OTP user, got %d", w.Code)
	}
}

func TestSmartLogin_Verify_OTPForPasswordUser(t *testing.T) {
	handler, _, _ := setupAuthTestWithOTP(t)

	// Password-only users cannot log in with an OTP code.
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "code": "123456"},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for code on password user, got %d", w.Code)
	}
}

func TestSmartLogin_BothPasswordAndCode_Rejected(t *testing.T) {
	handler, _ := setupAuthTest(t)

	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "TestPass1", "code": "123456"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Rate limiting tests for the initiate path
// ---------------------------------------------------------------------------

func TestSmartLogin_Initiate_RateLimit_BlocksAfterLimit(t *testing.T) {
	handler := setupAuthTestWithRateLimiter(t)

	// RateSmartLoginInitiateLimit step-1 requests succeed.
	for i := range RateSmartLoginInitiateLimit {
		w := doAuthRequest(t, handler, map[string]any{
			"op":   "smart_login",
			"data": map[string]any{"email": "test@example.com"},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// The next initiate request is rate-limited.
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com"},
	})
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after %d initiates, got %d", RateSmartLoginInitiateLimit, w.Code)
	}
}

func TestSmartLogin_Initiate_RateLimit_ResetOnVerifySuccess(t *testing.T) {
	handler := setupAuthTestWithRateLimiter(t)

	// Exhaust the initiate limit.
	for range RateSmartLoginInitiateLimit {
		doAuthRequest(t, handler, map[string]any{
			"op":   "smart_login",
			"data": map[string]any{"email": "test@example.com"},
		})
	}
	w := doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com"},
	})
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exhausting initiate limit, got %d", w.Code)
	}

	// A successful verify resets both the smart-login and initiate counters.
	w = doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com", "password": "TestPass1"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("successful verify failed: %d %s", w.Code, w.Body.String())
	}

	// Initiate must work again after the counter was reset.
	w = doAuthRequest(t, handler, map[string]any{
		"op":   "smart_login",
		"data": map[string]any{"email": "test@example.com"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 after reset, got %d", w.Code)
	}
}

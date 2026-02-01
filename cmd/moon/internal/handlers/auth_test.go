package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thalib/moon/cmd/moon/internal/auth"
	"github.com/thalib/moon/cmd/moon/internal/database"
)

func setupTestAuthHandler(t *testing.T) (*AuthHandler, database.Driver) {
	t.Helper()

	// Create in-memory SQLite database
	cfg := database.Config{
		ConnectionString: "sqlite://:memory:",
		MaxOpenConns:     1,
		MaxIdleConns:     1,
	}

	db, err := database.NewDriver(cfg)
	if err != nil {
		t.Fatalf("failed to create database driver: %v", err)
	}

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Initialize auth schema
	if err := auth.Bootstrap(ctx, db, nil); err != nil {
		t.Fatalf("failed to bootstrap auth: %v", err)
	}

	handler := NewAuthHandler(db, "test-secret-key", 3600, 604800)
	return handler, db
}

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a test user
	passwordHash, _ := auth.HashPassword("testpassword123")
	userRepo := auth.NewUserRepository(db)
	user := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "user",
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Test login
	body := LoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth:login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Login() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Login() access_token is empty")
	}
	if resp.RefreshToken == "" {
		t.Error("Login() refresh_token is empty")
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("Login() token_type = %q, want %q", resp.TokenType, "Bearer")
	}
	if resp.User.Username != "testuser" {
		t.Errorf("Login() user.username = %q, want %q", resp.User.Username, "testuser")
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a test user
	passwordHash, _ := auth.HashPassword("testpassword123")
	userRepo := auth.NewUserRepository(db)
	user := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "user",
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"wrong password", "testuser", "wrongpassword"},
		{"wrong username", "wronguser", "testpassword123"},
		{"empty username", "", "testpassword123"},
		{"empty password", "testuser", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := LoginRequest{
				Username: tt.username,
				Password: tt.password,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/auth:login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code == http.StatusOK {
				t.Errorf("Login() with %s should fail", tt.name)
			}
		})
	}
}

func TestAuthHandler_Login_WrongMethod(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/auth:login", nil)
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Login() with GET status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	body := RefreshRequest{
		RefreshToken: "some-refresh-token",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth:logout", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Logout() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthHandler_Logout_MissingToken(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	body := RefreshRequest{}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth:logout", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Logout() without token status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a test user and login to get tokens
	passwordHash, _ := auth.HashPassword("testpassword123")
	userRepo := auth.NewUserRepository(db)
	user := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "user",
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Login first
	loginBody := LoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}
	loginBytes, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest(http.MethodPost, "/auth:login", bytes.NewReader(loginBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	handler.Login(loginW, loginReq)

	var loginResp LoginResponse
	if err := json.NewDecoder(loginW.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	// Now test refresh
	body := RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth:refresh", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Refresh() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Refresh() access_token is empty")
	}
	if resp.RefreshToken == "" {
		t.Error("Refresh() refresh_token is empty")
	}
	// New refresh token should be different
	if resp.RefreshToken == loginResp.RefreshToken {
		t.Error("Refresh() should return new refresh token")
	}
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	body := RefreshRequest{
		RefreshToken: "invalid-refresh-token",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth:refresh", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Refresh() with invalid token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_GetMe_Unauthorized(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/auth:me", nil)
	w := httptest.NewRecorder()

	handler.GetMe(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetMe() without auth status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_GetMe_Success(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a test user
	passwordHash, _ := auth.HashPassword("testpassword123")
	userRepo := auth.NewUserRepository(db)
	user := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "user",
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Login to get access token
	loginBody := LoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}
	loginBytes, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest(http.MethodPost, "/auth:login", bytes.NewReader(loginBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	handler.Login(loginW, loginReq)

	var loginResp LoginResponse
	if err := json.NewDecoder(loginW.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	// Test GetMe with valid token
	req := httptest.NewRequest(http.MethodGet, "/auth:me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	w := httptest.NewRecorder()

	handler.GetMe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetMe() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	userInfo := resp["user"].(map[string]any)
	if userInfo["username"] != "testuser" {
		t.Errorf("GetMe() username = %v, want %q", userInfo["username"], "testuser")
	}
}

func TestAuthHandler_UpdateMe_Success(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a test user
	passwordHash, _ := auth.HashPassword("testpassword123")
	userRepo := auth.NewUserRepository(db)
	user := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "user",
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Login to get access token
	loginBody := LoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}
	loginBytes, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest(http.MethodPost, "/auth:login", bytes.NewReader(loginBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	handler.Login(loginW, loginReq)

	var loginResp LoginResponse
	if err := json.NewDecoder(loginW.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	// Test UpdateMe - update email
	updateBody := UpdateMeRequest{
		Email: "newemail@example.com",
	}
	updateBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPost, "/auth:me", bytes.NewReader(updateBytes))
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateMe() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Verify email was updated
	updatedUser, _ := userRepo.GetByUsername(ctx, "testuser")
	if updatedUser.Email != "newemail@example.com" {
		t.Errorf("UpdateMe() email = %q, want %q", updatedUser.Email, "newemail@example.com")
	}
}

func TestAuthHandler_UpdateMe_ChangePassword(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a test user
	passwordHash, _ := auth.HashPassword("testpassword123")
	userRepo := auth.NewUserRepository(db)
	user := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "user",
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Login to get access token
	loginBody := LoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}
	loginBytes, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest(http.MethodPost, "/auth:login", bytes.NewReader(loginBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	handler.Login(loginW, loginReq)

	var loginResp LoginResponse
	if err := json.NewDecoder(loginW.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	// Test UpdateMe - change password
	updateBody := UpdateMeRequest{
		Password:    "newpassword456",
		OldPassword: "testpassword123",
	}
	updateBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPost, "/auth:me", bytes.NewReader(updateBytes))
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateMe() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Verify new password works
	updatedUser, _ := userRepo.GetByUsername(ctx, "testuser")
	if err := auth.ComparePassword(updatedUser.PasswordHash, "newpassword456"); err != nil {
		t.Error("UpdateMe() new password doesn't work")
	}
}

func TestAuthHandler_UpdateMe_WrongOldPassword(t *testing.T) {
	handler, db := setupTestAuthHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a test user
	passwordHash, _ := auth.HashPassword("testpassword123")
	userRepo := auth.NewUserRepository(db)
	user := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "user",
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Login to get access token
	loginBody := LoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}
	loginBytes, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest(http.MethodPost, "/auth:login", bytes.NewReader(loginBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	handler.Login(loginW, loginReq)

	var loginResp LoginResponse
	if err := json.NewDecoder(loginW.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	// Test UpdateMe - change password with wrong old password
	updateBody := UpdateMeRequest{
		Password:    "newpassword456",
		OldPassword: "wrongoldpassword",
	}
	updateBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPost, "/auth:me", bytes.NewReader(updateBytes))
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateMe(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UpdateMe() with wrong old password status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestNewAuthHandler(t *testing.T) {
	cfg := database.Config{
		ConnectionString: "sqlite://:memory:",
		MaxOpenConns:     1,
		MaxIdleConns:     1,
	}

	db, err := database.NewDriver(cfg)
	if err != nil {
		t.Fatalf("failed to create database driver: %v", err)
	}
	defer db.Close()

	handler := NewAuthHandler(db, "secret", 3600, 604800)
	if handler == nil {
		t.Error("NewAuthHandler() returned nil")
	}
	if handler.userRepo == nil {
		t.Error("NewAuthHandler() userRepo is nil")
	}
	if handler.tokenRepo == nil {
		t.Error("NewAuthHandler() tokenRepo is nil")
	}
	if handler.tokenService == nil {
		t.Error("NewAuthHandler() tokenService is nil")
	}
}

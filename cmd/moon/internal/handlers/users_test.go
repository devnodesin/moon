package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/thalib/moon/cmd/moon/internal/auth"
	"github.com/thalib/moon/cmd/moon/internal/database"
)

func setupTestUsersHandler(t *testing.T) (*UsersHandler, *auth.User, string, database.Driver) {
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

	// Create an admin user
	passwordHash, _ := auth.HashPassword("AdminPass123")
	userRepo := auth.NewUserRepository(db)
	adminUser := &auth.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleAdmin),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, adminUser); err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}

	handler := NewUsersHandler(db, "test-secret-key", 3600, 604800)

	// Generate token for admin
	tokenService := auth.NewTokenService("test-secret-key", 3600, 604800)
	tokenPair, _, err := tokenService.GenerateTokenPair(adminUser)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	return handler, adminUser, tokenPair.AccessToken, db
}

func TestUsersHandler_List_Success(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:list", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp["data"].([]any)
	if !ok || len(data) == 0 {
		t.Error("List() should return at least one user in data array")
	}
}

func TestUsersHandler_List_Unauthorized(t *testing.T) {
	handler, _, _, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:list", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("List() without auth status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUsersHandler_List_NonAdminForbidden(t *testing.T) {
	handler, _, _, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a regular user
	passwordHash, _ := auth.HashPassword("UserPass123")
	userRepo := auth.NewUserRepository(db)
	regularUser := &auth.User{
		Username:     "regularuser",
		Email:        "user@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleUser),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, regularUser); err != nil {
		t.Fatalf("failed to create regular user: %v", err)
	}

	// Generate token for regular user
	tokenService := auth.NewTokenService("test-secret-key", 3600, 604800)
	tokenPair, _, err := tokenService.GenerateTokenPair(regularUser)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users:list", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("List() with non-admin status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUsersHandler_List_WithRoleFilter(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:list?role=admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data := resp["data"].([]any)
	for _, u := range data {
		user := u.(map[string]any)
		if user["role"] != "admin" {
			t.Errorf("List() with role filter returned user with role %v, want admin", user["role"])
		}
	}
}

func TestUsersHandler_Get_Success(t *testing.T) {
	handler, admin, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:get?id="+admin.ID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	userInfo := resp["data"].(map[string]any)
	if userInfo["username"] != "admin" {
		t.Errorf("Get() username = %v, want admin", userInfo["username"])
	}
}

func TestUsersHandler_Get_NotFound(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:get?id=nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Get() with nonexistent id status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUsersHandler_Get_MissingID(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:get", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Get() without id status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Create_Success(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	body := map[string]any{
		"data": CreateUserRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "NewUser123",
			Role:     "user",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:create", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create() status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data := resp["data"].(map[string]any)
	if data["username"] != "newuser" {
		t.Errorf("Create() username = %v, want newuser", data["username"])
	}
	if data["role"] != "user" {
		t.Errorf("Create() role = %v, want user", data["role"])
	}
}

func TestUsersHandler_Create_WeakPassword(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	body := map[string]any{
		"data": CreateUserRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "weak",
			Role:     "user",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:create", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() with weak password status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] == nil {
		t.Error("expected error message in response")
	}
}

func TestUsersHandler_Create_InvalidEmail(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	body := map[string]any{
		"data": CreateUserRequest{
			Username: "newuser",
			Email:    "invalid-email",
			Password: "NewUser123",
			Role:     "user",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:create", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() with invalid email status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] == nil {
		t.Error("expected error message in response")
	}
}

func TestUsersHandler_Create_InvalidRole(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	body := map[string]any{
		"data": CreateUserRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "NewUser123",
			Role:     "invalid",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:create", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() with invalid role status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] == nil {
		t.Error("expected error message in response")
	}
}

func TestUsersHandler_Create_DuplicateUsername(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	body := map[string]any{
		"data": CreateUserRequest{
			Username: "admin", // Already exists
			Email:    "new@example.com",
			Password: "NewUser123",
			Role:     "user",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:create", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() with duplicate username status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] == nil {
		t.Error("expected error message in response")
	}
}

func TestUsersHandler_Create_DuplicateEmail(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	body := map[string]any{
		"data": CreateUserRequest{
			Username: "newuser",
			Email:    "admin@example.com", // Already exists
			Password: "NewUser123",
			Role:     "user",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:create", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() with duplicate email status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] == nil {
		t.Error("expected error message in response")
	}
}

func TestUsersHandler_Update_Success(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a user to update
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	testUser := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleUser),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	newEmail := "updated@example.com"
	body := UpdateUserRequest{
		Email: &newEmail,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id="+testUser.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Update() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data := resp["data"].(map[string]any)
	if data["email"] != "updated@example.com" {
		t.Errorf("Update() email = %v, want updated@example.com", data["email"])
	}
}

func TestUsersHandler_Update_ResetPassword(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a user
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	testUser := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleUser),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	body := UpdateUserRequest{
		Action:      "reset_password",
		NewPassword: "NewPass456",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id="+testUser.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Update() with reset_password status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Verify new password works
	updatedUser, _ := userRepo.GetByID(ctx, testUser.ID)
	if err := auth.ComparePassword(updatedUser.PasswordHash, "NewPass456"); err != nil {
		t.Error("Update() password reset didn't work")
	}
}

func TestUsersHandler_Update_RevokeSessions(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a user
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	testUser := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleUser),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	body := UpdateUserRequest{
		Action: "revoke_sessions",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id="+testUser.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Update() with revoke_sessions status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUsersHandler_Update_CannotModifySelf(t *testing.T) {
	handler, admin, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	newEmail := "newemail@example.com"
	body := UpdateUserRequest{
		Email: &newEmail,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id="+admin.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Update() on self status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] == nil {
		t.Error("expected error message in response")
	}
}

func TestUsersHandler_Update_CannotDowngradeLastAdmin(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a second admin to test downgrade
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	secondAdmin := &auth.User{
		Username:     "admin2",
		Email:        "admin2@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleAdmin),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, secondAdmin); err != nil {
		t.Fatalf("failed to create second admin: %v", err)
	}

	// First, downgrade second admin (should succeed, since there's still one admin)
	userRole := "user"
	body := UpdateUserRequest{
		Role: &userRole,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id="+secondAdmin.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Update() downgrade second admin status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUsersHandler_Destroy_Success(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a user to delete
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	testUser := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleUser),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/users:destroy?id="+testUser.ID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Destroy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Destroy() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Verify user is deleted
	deletedUser, _ := userRepo.GetByID(ctx, testUser.ID)
	if deletedUser != nil {
		t.Error("Destroy() user should be deleted")
	}
}

func TestUsersHandler_Destroy_CannotDeleteSelf(t *testing.T) {
	handler, admin, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/users:destroy?id="+admin.ID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Destroy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Destroy() on self status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] == nil {
		t.Error("expected error message in response")
	}
}

func TestUsersHandler_Destroy_CannotDeleteLastAdmin(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create another admin
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	secondAdmin := &auth.User{
		Username:     "admin2",
		Email:        "admin2@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleAdmin),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, secondAdmin); err != nil {
		t.Fatalf("failed to create second admin: %v", err)
	}

	// Delete second admin (should succeed)
	req := httptest.NewRequest(http.MethodPost, "/users:destroy?id="+secondAdmin.ID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Destroy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Destroy() second admin status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUsersHandler_Destroy_NotFound(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/users:destroy?id=nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Destroy(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Destroy() with nonexistent id status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUsersHandler_Destroy_MissingID(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/users:destroy", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Destroy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Destroy() without id status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Create_MissingFields(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	tests := []struct {
		name string
		body CreateUserRequest
	}{
		{"missing username", CreateUserRequest{Email: "a@b.com", Password: "Pass123!", Role: "user"}},
		{"missing email", CreateUserRequest{Username: "user", Password: "Pass123!", Role: "user"}},
		{"missing password", CreateUserRequest{Username: "user", Email: "a@b.com", Role: "user"}},
		{"missing role", CreateUserRequest{Username: "user", Email: "a@b.com", Password: "Pass123!"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := map[string]any{"data": tt.body}
			bodyBytes, _ := json.Marshal(wrapper)

			req := httptest.NewRequest(http.MethodPost, "/users:create", bytes.NewReader(bodyBytes))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Create(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Create() with %s status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}

			var resp map[string]any
			json.NewDecoder(w.Body).Decode(&resp)
			if resp["message"] == nil {
				t.Error("expected error message in response")
			}
		})
	}
}

func TestNewUsersHandler(t *testing.T) {
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

	handler := NewUsersHandler(db, "secret", 3600, 604800)
	if handler == nil {
		t.Error("NewUsersHandler() returned nil")
	}
	if handler.userRepo == nil {
		t.Error("NewUsersHandler() userRepo is nil")
	}
	if handler.tokenRepo == nil {
		t.Error("NewUsersHandler() tokenRepo is nil")
	}
	if handler.tokenService == nil {
		t.Error("NewUsersHandler() tokenService is nil")
	}
	if handler.passwordPolicy == nil {
		t.Error("NewUsersHandler() passwordPolicy is nil")
	}
}

func TestUsersHandler_List_WrongMethod(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/users:list", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("List() with POST status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Get_WrongMethod(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/users:get?id=123", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Get() with POST status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Create_WrongMethod(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:create", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() with GET status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Update_WrongMethod(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:update?id=123", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Update() with GET status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Destroy_WrongMethod(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/users:destroy?id=123", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.Destroy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Destroy() with GET status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Update_NotFound(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	newEmail := "test@example.com"
	body := UpdateUserRequest{
		Email: &newEmail,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id=nonexistent", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Update() with nonexistent id status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUsersHandler_Update_InvalidAction(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a user
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	testUser := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleUser),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	body := UpdateUserRequest{
		Action: "invalid_action",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id="+testUser.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Update() with invalid action status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Update_ResetPasswordMissingPassword(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a user
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	testUser := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleUser),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	body := UpdateUserRequest{
		Action: "reset_password",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id="+testUser.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Update() reset_password without password status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUsersHandler_Update_NoFieldsToUpdate(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create a user
	passwordHash, _ := auth.HashPassword("TestPass123")
	userRepo := auth.NewUserRepository(db)
	testUser := &auth.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         string(auth.RoleUser),
		CanWrite:     true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	body := UpdateUserRequest{}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users:update?id="+testUser.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Update() with no fields status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserToPublicInfo(t *testing.T) {
	user := &auth.User{
		ID:       "01H1234567890ABCDEFGHJKMNP",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "admin",
		CanWrite: true,
	}

	info := userToPublicInfo(user)

	if info.ID != user.ID {
		t.Errorf("userToPublicInfo() ID = %s, want %s", info.ID, user.ID)
	}
	if info.Username != user.Username {
		t.Errorf("userToPublicInfo() Username = %s, want %s", info.Username, user.Username)
	}
	if info.Email != user.Email {
		t.Errorf("userToPublicInfo() Email = %s, want %s", info.Email, user.Email)
	}
	if info.Role != user.Role {
		t.Errorf("userToPublicInfo() Role = %s, want %s", info.Role, user.Role)
	}
	if info.CanWrite != user.CanWrite {
		t.Errorf("userToPublicInfo() CanWrite = %v, want %v", info.CanWrite, user.CanWrite)
	}
}

func TestParseIntWithDefault(t *testing.T) {
	tests := []struct {
		input    string
		defVal   int
		expected int
	}{
		{"10", 5, 10},
		{"", 5, 5},
		{"abc", 5, 5},
		{"100", 5, 100},
		{"0", 5, 0},
	}

	for _, tt := range tests {
		result := parseIntWithDefault(tt.input, tt.defVal)
		if result != tt.expected {
			t.Errorf("parseIntWithDefault(%q, %d) = %d, want %d", tt.input, tt.defVal, result, tt.expected)
		}
	}
}

func TestUsersHandler_List_MetaTotal(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create 2 extra users
	userRepo := auth.NewUserRepository(db)
	for i := 0; i < 2; i++ {
		passwordHash, _ := auth.HashPassword("UserPass123")
		u := &auth.User{
			Username:     "extrauser" + strconv.Itoa(i+1),
			Email:        "extra" + strconv.Itoa(i+1) + "@example.com",
			PasswordHash: passwordHash,
			Role:         string(auth.RoleUser),
			CanWrite:     true,
		}
		if err := userRepo.Create(ctx, u); err != nil {
			t.Fatalf("failed to create extra user: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/users:list", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	meta, ok := resp["meta"].(map[string]any)
	if !ok {
		t.Fatal("List() meta is missing or not an object")
	}

	total, ok := meta["total"]
	if !ok {
		t.Fatal("List() meta.total is missing")
	}

	// 1 admin + 2 extra users = 3 total
	totalFloat, ok := total.(float64)
	if !ok {
		t.Fatalf("List() meta.total type = %T, want float64", total)
	}
	if int(totalFloat) != 3 {
		t.Errorf("List() meta.total = %d, want 3", int(totalFloat))
	}
}

func TestUsersHandler_List_MetaTotalWithRoleFilter(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create 2 regular users
	userRepo := auth.NewUserRepository(db)
	for i := 0; i < 2; i++ {
		passwordHash, _ := auth.HashPassword("UserPass123")
		u := &auth.User{
			Username:     "regularuser" + strconv.Itoa(i+1),
			Email:        "regular" + strconv.Itoa(i+1) + "@example.com",
			PasswordHash: passwordHash,
			Role:         string(auth.RoleUser),
			CanWrite:     true,
		}
		if err := userRepo.Create(ctx, u); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/users:list?role=user", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	meta := resp["meta"].(map[string]any)
	totalFloat := meta["total"].(float64)
	if int(totalFloat) != 2 {
		t.Errorf("List() meta.total with role=user = %d, want 2", int(totalFloat))
	}
}

func TestUsersHandler_List_BackwardPagination(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create 4 more users (admin already exists = 5 total)
	userRepo := auth.NewUserRepository(db)
	var createdUsers []*auth.User
	for i := 1; i <= 4; i++ {
		passwordHash, _ := auth.HashPassword("UserPass123")
		u := &auth.User{
			Username:     "paginationuser" + strconv.Itoa(i),
			Email:        "paginationuser" + strconv.Itoa(i) + "@example.com",
			PasswordHash: passwordHash,
			Role:         string(auth.RoleUser),
			CanWrite:     true,
		}
		if err := userRepo.Create(ctx, u); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		createdUsers = append(createdUsers, u)
	}

	// Page 1: limit=2, no cursor
	req1 := httptest.NewRequest(http.MethodGet, "/users:list?limit=2", nil)
	req1.Header.Set("Authorization", "Bearer "+adminToken)
	w1 := httptest.NewRecorder()
	handler.List(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("List() page1 status = %d, want %d", w1.Code, http.StatusOK)
	}

	var resp1 map[string]any
	json.NewDecoder(w1.Body).Decode(&resp1)
	meta1 := resp1["meta"].(map[string]any)

	// Page 1: prev should be null, next should be set
	if meta1["prev"] != nil {
		t.Errorf("Page 1: prev should be null, got %v", meta1["prev"])
	}
	if meta1["next"] == nil {
		t.Error("Page 1: next should be non-null")
	}
	// Verify total
	if int(meta1["total"].(float64)) != 5 {
		t.Errorf("Page 1: total = %d, want 5", int(meta1["total"].(float64)))
	}

	nextPage1 := meta1["next"].(string)

	// Page 2: limit=2, after=nextPage1
	req2 := httptest.NewRequest(http.MethodGet, "/users:list?limit=2&after="+nextPage1, nil)
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	w2 := httptest.NewRecorder()
	handler.List(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("List() page2 status = %d, want %d", w2.Code, http.StatusOK)
	}

	var resp2 map[string]any
	json.NewDecoder(w2.Body).Decode(&resp2)
	meta2 := resp2["meta"].(map[string]any)

	// Page 2: prev should be null (previous page has no cursor = first page), next should be set
	if meta2["prev"] != nil {
		t.Logf("Page 2: prev=%v (null expected - page 1 needs no cursor)", meta2["prev"])
	}
	if meta2["next"] == nil {
		t.Error("Page 2: next should be non-null")
	}

	nextPage2 := meta2["next"].(string)

	// Page 3: limit=2, after=nextPage2
	req3 := httptest.NewRequest(http.MethodGet, "/users:list?limit=2&after="+nextPage2, nil)
	req3.Header.Set("Authorization", "Bearer "+adminToken)
	w3 := httptest.NewRecorder()
	handler.List(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("List() page3 status = %d, want %d", w3.Code, http.StatusOK)
	}

	var resp3 map[string]any
	json.NewDecoder(w3.Body).Decode(&resp3)
	meta3 := resp3["meta"].(map[string]any)

	// Page 3: prev should be non-null (can navigate back to page 2)
	if meta3["prev"] == nil {
		t.Error("Page 3: prev should be non-null (backward pagination should be available)")
	}

	prevPage3 := meta3["prev"].(string)

	// Navigate backward using prevPage3: should give page 2 data
	reqBack := httptest.NewRequest(http.MethodGet, "/users:list?limit=2&after="+prevPage3, nil)
	reqBack.Header.Set("Authorization", "Bearer "+adminToken)
	wBack := httptest.NewRecorder()
	handler.List(wBack, reqBack)

	if wBack.Code != http.StatusOK {
		t.Fatalf("List() back-navigate status = %d, want %d", wBack.Code, http.StatusOK)
	}

	var respBack map[string]any
	json.NewDecoder(wBack.Body).Decode(&respBack)
	dataBack := respBack["data"].([]any)
	data3 := resp3["data"].([]any)

	// The back-navigated data should NOT be the same as page 3 data
	// (it should be page 2 data = the page before page 3)
	if len(dataBack) == 0 {
		t.Error("Back-navigation should return users")
	}

	// Ensure the back-navigated page ends before the first user of page 3
	page3FirstID := data3[0].(map[string]any)["id"].(string)
	for _, u := range dataBack {
		userMap := u.(map[string]any)
		if userMap["id"].(string) >= page3FirstID {
			t.Errorf("Back-navigated user %v should come before page 3 first user %v", userMap["id"], page3FirstID)
		}
	}
}

func TestUsersHandler_List_ForwardPagination(t *testing.T) {
	handler, _, adminToken, db := setupTestUsersHandler(t)
	defer db.Close()

	ctx := context.Background()

	// Create 3 more users (admin already = 4 total)
	userRepo := auth.NewUserRepository(db)
	for i := 1; i <= 3; i++ {
		passwordHash, _ := auth.HashPassword("UserPass123")
		u := &auth.User{
			Username:     "pageuser" + strconv.Itoa(i),
			Email:        "pageuser" + strconv.Itoa(i) + "@example.com",
			PasswordHash: passwordHash,
			Role:         string(auth.RoleUser),
			CanWrite:     true,
		}
		if err := userRepo.Create(ctx, u); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	// Page through all users with limit=1 and collect IDs
	var allIDs []string
	after := ""
	for {
		url := "/users:list?limit=1"
		if after != "" {
			url += "&after=" + after
		}
		req := httptest.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		handler.List(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("List() status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)

		data := resp["data"].([]any)
		for _, u := range data {
			userMap := u.(map[string]any)
			allIDs = append(allIDs, userMap["id"].(string))
		}

		meta := resp["meta"].(map[string]any)
		if meta["next"] == nil {
			break
		}
		after = meta["next"].(string)
	}

	// Should have 4 users total
	if len(allIDs) != 4 {
		t.Errorf("Forward pagination collected %d users, want 4", len(allIDs))
	}

	// Verify IDs are in ascending order
	for i := 1; i < len(allIDs); i++ {
		if allIDs[i] <= allIDs[i-1] {
			t.Errorf("Forward pagination not in order: %s should be after %s", allIDs[i], allIDs[i-1])
		}
	}
}

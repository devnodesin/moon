package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thalib/moon/cmd/moon/internal/auth"
)

func TestNewAuthorizationMiddleware(t *testing.T) {
	m := NewAuthorizationMiddleware()
	if m == nil {
		t.Fatal("NewAuthorizationMiddleware returned nil")
	}
}

func TestGetAuthEntity(t *testing.T) {
	t.Run("entity present in context", func(t *testing.T) {
		entity := &AuthEntity{
			ID:       "test-ulid",
			Type:     EntityTypeUser,
			Role:     string(auth.RoleAdmin),
			CanWrite: true,
			Username: "testuser",
		}

		ctx := context.WithValue(context.Background(), AuthEntityContextKey, entity)
		retrieved, ok := GetAuthEntity(ctx)

		if !ok {
			t.Fatal("GetAuthEntity returned false for context with entity")
		}
		if retrieved.ID != "test-ulid" {
			t.Errorf("Expected ID 'test-ulid', got '%s'", retrieved.ID)
		}
		if retrieved.Type != EntityTypeUser {
			t.Errorf("Expected type '%s', got '%s'", EntityTypeUser, retrieved.Type)
		}
	})

	t.Run("entity not present in context", func(t *testing.T) {
		ctx := context.Background()
		_, ok := GetAuthEntity(ctx)
		if ok {
			t.Error("GetAuthEntity returned true for context without entity")
		}
	})
}

func TestSetAuthEntity(t *testing.T) {
	entity := &AuthEntity{
		ID:       "test-ulid",
		Type:     EntityTypeAPIKey,
		Role:     string(auth.RoleUser),
		CanWrite: false,
	}

	ctx := SetAuthEntity(context.Background(), entity)
	retrieved, ok := GetAuthEntity(ctx)

	if !ok {
		t.Fatal("Entity not found after SetAuthEntity")
	}
	if retrieved.ID != entity.ID {
		t.Errorf("Expected ID '%s', got '%s'", entity.ID, retrieved.ID)
	}
}

func TestRequireRole_AdminAccess(t *testing.T) {
	m := NewAuthorizationMiddleware()

	t.Run("admin role has access to admin endpoint", func(t *testing.T) {
		entity := &AuthEntity{
			ID:   "admin-ulid",
			Type: EntityTypeUser,
			Role: string(auth.RoleAdmin),
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireRole(string(auth.RoleAdmin))(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for admin user")
		}
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("user role denied access to admin endpoint", func(t *testing.T) {
		entity := &AuthEntity{
			ID:   "user-ulid",
			Type: EntityTypeUser,
			Role: string(auth.RoleUser),
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}

		req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireRole(string(auth.RoleAdmin))(handler)(w, req)

		if handlerCalled {
			t.Error("Handler should not be called for user role on admin endpoint")
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("no auth entity returns forbidden", func(t *testing.T) {
		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}

		req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
		w := httptest.NewRecorder()

		m.RequireRole(string(auth.RoleAdmin))(handler)(w, req)

		if handlerCalled {
			t.Error("Handler should not be called without auth entity")
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestRequireAdmin(t *testing.T) {
	m := NewAuthorizationMiddleware()

	t.Run("admin role allowed", func(t *testing.T) {
		entity := &AuthEntity{
			ID:   "admin-ulid",
			Type: EntityTypeUser,
			Role: string(auth.RoleAdmin),
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireAdmin(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for admin user")
		}
	})

	t.Run("user role denied", func(t *testing.T) {
		entity := &AuthEntity{
			ID:   "user-ulid",
			Type: EntityTypeUser,
			Role: string(auth.RoleUser),
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireAdmin(handler)(w, req)

		if handlerCalled {
			t.Error("Handler should not be called for user role")
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestRequireWrite(t *testing.T) {
	m := NewAuthorizationMiddleware()

	t.Run("admin always has write access", func(t *testing.T) {
		entity := &AuthEntity{
			ID:       "admin-ulid",
			Type:     EntityTypeUser,
			Role:     string(auth.RoleAdmin),
			CanWrite: false, // Even with CanWrite false, admin should have write access
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodPost, "/products:create", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireWrite(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for admin user")
		}
	})

	t.Run("user with can_write has write access", func(t *testing.T) {
		entity := &AuthEntity{
			ID:       "user-ulid",
			Type:     EntityTypeUser,
			Role:     string(auth.RoleUser),
			CanWrite: true,
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodPost, "/products:create", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireWrite(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for user with can_write")
		}
	})

	t.Run("user without can_write denied write access", func(t *testing.T) {
		entity := &AuthEntity{
			ID:       "user-ulid",
			Type:     EntityTypeUser,
			Role:     string(auth.RoleUser),
			CanWrite: false,
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}

		req := httptest.NewRequest(http.MethodPost, "/products:create", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireWrite(handler)(w, req)

		if handlerCalled {
			t.Error("Handler should not be called for user without can_write")
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("API key with can_write has write access", func(t *testing.T) {
		entity := &AuthEntity{
			ID:       "apikey-ulid",
			Type:     EntityTypeAPIKey,
			Role:     string(auth.RoleUser),
			CanWrite: true,
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodPost, "/products:create", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireWrite(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for API key with can_write")
		}
	})

	t.Run("no auth entity returns forbidden", func(t *testing.T) {
		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}

		req := httptest.NewRequest(http.MethodPost, "/products:create", nil)
		w := httptest.NewRecorder()

		m.RequireWrite(handler)(w, req)

		if handlerCalled {
			t.Error("Handler should not be called without auth entity")
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestRequireAuthenticated(t *testing.T) {
	m := NewAuthorizationMiddleware()

	t.Run("any authenticated user allowed", func(t *testing.T) {
		entity := &AuthEntity{
			ID:       "user-ulid",
			Type:     EntityTypeUser,
			Role:     string(auth.RoleUser),
			CanWrite: false,
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodGet, "/auth:me", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireAuthenticated(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for authenticated user")
		}
	})

	t.Run("API key authenticated allowed", func(t *testing.T) {
		entity := &AuthEntity{
			ID:   "apikey-ulid",
			Type: EntityTypeAPIKey,
			Role: string(auth.RoleUser),
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodGet, "/collections:list", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireAuthenticated(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for authenticated API key")
		}
	})

	t.Run("no auth entity returns forbidden", func(t *testing.T) {
		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}

		req := httptest.NewRequest(http.MethodGet, "/collections:list", nil)
		w := httptest.NewRecorder()

		m.RequireAuthenticated(handler)(w, req)

		if handlerCalled {
			t.Error("Handler should not be called without auth entity")
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestRequireRole_UserRoleAccess(t *testing.T) {
	m := NewAuthorizationMiddleware()

	t.Run("user role has access to user endpoint", func(t *testing.T) {
		entity := &AuthEntity{
			ID:   "user-ulid",
			Type: EntityTypeUser,
			Role: string(auth.RoleUser),
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodGet, "/products:list", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireRole(string(auth.RoleUser))(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for user role")
		}
	})

	t.Run("admin role has access to user endpoint", func(t *testing.T) {
		entity := &AuthEntity{
			ID:   "admin-ulid",
			Type: EntityTypeUser,
			Role: string(auth.RoleAdmin),
		}

		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}

		req := httptest.NewRequest(http.MethodGet, "/products:list", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RequireRole(string(auth.RoleUser))(handler)(w, req)

		if !handlerCalled {
			t.Error("Handler was not called for admin role on user endpoint")
		}
	})
}

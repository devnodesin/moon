package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Mock database adapter for auth tests
// ---------------------------------------------------------------------------

type mockAuthDB struct {
	users   []map[string]any
	apikeys []map[string]any
	updates []mockUpdate
}

type mockUpdate struct {
	table string
	id    string
	data  map[string]any
}

func (m *mockAuthDB) Ping(_ context.Context) error              { return nil }
func (m *mockAuthDB) Close() error                              { return nil }
func (m *mockAuthDB) ExecDDL(_ context.Context, _ string) error { return nil }
func (m *mockAuthDB) InsertRow(_ context.Context, _ string, _ map[string]any) error {
	return nil
}
func (m *mockAuthDB) UpdateRow(_ context.Context, table, id string, data map[string]any) error {
	m.updates = append(m.updates, mockUpdate{table: table, id: id, data: data})
	return nil
}
func (m *mockAuthDB) DeleteRow(_ context.Context, _ string, _ string) error { return nil }
func (m *mockAuthDB) ListTables(_ context.Context) ([]string, error)        { return nil, nil }
func (m *mockAuthDB) DescribeTable(_ context.Context, _ string) ([]ColumnInfo, error) {
	return nil, nil
}
func (m *mockAuthDB) CountRows(_ context.Context, _ string) (int, error) { return 0, nil }

func (m *mockAuthDB) QueryRows(_ context.Context, table string, opts QueryOptions) ([]map[string]any, int, error) {
	switch table {
	case "users":
		for _, u := range m.users {
			for _, f := range opts.Filters {
				if f.Field == "id" && u["id"] == f.Value {
					return []map[string]any{u}, 1, nil
				}
			}
		}
		return nil, 0, nil
	case "apikeys":
		for _, k := range m.apikeys {
			for _, f := range opts.Filters {
				if f.Field == "key_hash" && k["key_hash"] == f.Value {
					return []map[string]any{k}, 1, nil
				}
			}
		}
		return nil, 0, nil
	}
	return nil, 0, nil
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func testJWTSecret() string {
	return "test-secret-that-is-at-least-32-chars-long"
}

func createTestJWT(t *testing.T, userID, jti, role string, canWrite bool, expirySeconds int) string {
	t.Helper()
	token, _, err := CreateAccessToken(userID, jti, role, canWrite, testJWTSecret(), expirySeconds)
	if err != nil {
		t.Fatalf("failed to create test JWT: %v", err)
	}
	return token
}

func createTestAPIKey() (raw string, hash string) {
	// moon_live_ is 10 chars, need 74 total, so 64 more chars
	raw = APIKeyPrefix + strings.Repeat("a", APIKeyTotalLen-len(APIKeyPrefix))
	h := sha256.Sum256([]byte(raw))
	hash = fmt.Sprintf("%x", h)
	return raw, hash
}

func testAuthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := GetAuthIdentity(r.Context())
		if !ok {
			WriteJSON(w, http.StatusOK, map[string]string{"auth": "none"})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{
			"credential_type": identity.CredentialType,
			"caller_id":       identity.CallerID,
			"role":            identity.Role,
			"can_write":       identity.CanWrite,
			"jti":             identity.JTI,
		})
	})
}

func buildAuthTestServer(t *testing.T, db DatabaseAdapter, prefix string) http.Handler {
	t.Helper()
	cfg := &AppConfig{
		Server: ServerConfig{Prefix: prefix},
		CORS:   CORSConfig{Enabled: false},
	}
	logger := NewTestLogger(&bytes.Buffer{})
	mux := NewRouter(prefix, logger, db, cfg)
	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), prefix, jtiStore)
	return BuildHandler(mux, cfg, logger, WithAuthMiddleware(am))
}

// ---------------------------------------------------------------------------
// JTI Revocation Store tests
// ---------------------------------------------------------------------------

func TestJTIRevocationStore_RevokeAndCheck(t *testing.T) {
	store := NewJTIRevocationStore()

	if store.IsRevoked("jti1") {
		t.Fatal("expected jti1 to not be revoked")
	}

	store.Revoke("jti1")

	if !store.IsRevoked("jti1") {
		t.Fatal("expected jti1 to be revoked")
	}

	if store.IsRevoked("jti2") {
		t.Fatal("expected jti2 to not be revoked")
	}
}

func TestJTIRevocationStore_ConcurrentAccess(t *testing.T) {
	store := NewJTIRevocationStore()
	done := make(chan struct{})

	go func() {
		for i := 0; i < 100; i++ {
			store.Revoke(fmt.Sprintf("jti-%d", i))
		}
		close(done)
	}()

	for i := 0; i < 100; i++ {
		store.IsRevoked(fmt.Sprintf("jti-%d", i))
	}
	<-done
}

// ---------------------------------------------------------------------------
// Auth Identity context tests
// ---------------------------------------------------------------------------

func TestAuthIdentityContext(t *testing.T) {
	ctx := context.Background()

	_, ok := GetAuthIdentity(ctx)
	if ok {
		t.Fatal("expected no identity in empty context")
	}

	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user1",
		Role:           "admin",
		CanWrite:       true,
		JTI:            "jti1",
	}
	ctx = SetAuthIdentity(ctx, identity)

	got, ok := GetAuthIdentity(ctx)
	if !ok {
		t.Fatal("expected identity in context")
	}
	if got.CallerID != "user1" {
		t.Fatalf("expected user1, got %s", got.CallerID)
	}
}

// ---------------------------------------------------------------------------
// Credential type detection tests
// ---------------------------------------------------------------------------

func TestDetectCredentialType(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"jwt", "header.payload.signature", CredentialTypeJWT},
		{"api key", APIKeyPrefix + strings.Repeat("x", APIKeyTotalLen-len(APIKeyPrefix)), CredentialTypeAPIKey},
		{"empty", "", ""},
		{"random string", "not-a-valid-token", ""},
		{"two dots", "a.b", ""},
		{"four dots", "a.b.c.d", ""},
		{"empty segment", "a..c", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectCredentialType(tt.token)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Bearer token extraction tests
// ---------------------------------------------------------------------------

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
		ok       bool
	}{
		{"valid", "Bearer mytoken", "mytoken", true},
		{"no prefix", "mytoken", "", false},
		{"empty bearer", "Bearer ", "", false},
		{"no header", "", "", false},
		{"basic auth", "Basic abc", "", false},
		{"bearer lowercase", "bearer mytoken", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("Authorization", tt.header)
			}
			got, ok := extractBearerToken(r)
			if ok != tt.ok {
				t.Fatalf("ok: expected %v, got %v", tt.ok, ok)
			}
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Public route bypass tests
// ---------------------------------------------------------------------------

func TestAuthenticate_PublicRoutes(t *testing.T) {
	db := &mockAuthDB{}
	handler := buildAuthTestServer(t, db, "")

	publicRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/"},
		{http.MethodGet, "/health"},
		{http.MethodPost, "/auth:session"},
	}

	for _, route := range publicRoutes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Public routes should not return 401
			if w.Code == http.StatusUnauthorized {
				t.Fatalf("public route %s %s returned 401", route.method, route.path)
			}
		})
	}
}

func TestAuthenticate_PublicRoutesWithPrefix(t *testing.T) {
	db := &mockAuthDB{}
	handler := buildAuthTestServer(t, db, "/api")

	publicRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api"},
		{http.MethodGet, "/api/"},
		{http.MethodGet, "/api/health"},
		{http.MethodPost, "/api/auth:session"},
	}

	for _, route := range publicRoutes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code == http.StatusUnauthorized {
				t.Fatalf("public route %s %s returned 401", route.method, route.path)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Protected route requires auth tests
// ---------------------------------------------------------------------------

func TestAuthenticate_ProtectedRouteWithoutToken(t *testing.T) {
	db := &mockAuthDB{}
	handler := buildAuthTestServer(t, db, "")

	protectedRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/auth:me"},
		{http.MethodPost, "/auth:me"},
		{http.MethodGet, "/collections:query"},
		{http.MethodPost, "/collections:mutate"},
		{http.MethodGet, "/data/products:query"},
		{http.MethodPost, "/data/products:mutate"},
	}

	for _, route := range protectedRoutes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
		})
	}
}

func TestAuthenticate_MalformedBearer(t *testing.T) {
	db := &mockAuthDB{}
	handler := buildAuthTestServer(t, db, "")

	req := httptest.NewRequest(http.MethodGet, "/auth:me", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_UnknownCredentialType(t *testing.T) {
	db := &mockAuthDB{}
	handler := buildAuthTestServer(t, db, "")

	req := httptest.NewRequest(http.MethodGet, "/auth:me", nil)
	req.Header.Set("Authorization", "Bearer some-random-string")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// JWT authentication tests
// ---------------------------------------------------------------------------

func TestAuthenticate_ValidJWT(t *testing.T) {
	userID := GenerateULID()
	db := &mockAuthDB{
		users: []map[string]any{
			{"id": userID, "role": "admin", "can_write": true},
		},
	}

	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	token := createTestJWT(t, userID, "test-jti", "admin", true, 3600)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["credential_type"] != CredentialTypeJWT {
		t.Fatalf("expected jwt, got %v", body["credential_type"])
	}
	if body["caller_id"] != userID {
		t.Fatalf("expected %s, got %v", userID, body["caller_id"])
	}
	if body["role"] != "admin" {
		t.Fatalf("expected admin, got %v", body["role"])
	}
}

func TestAuthenticate_ExpiredJWT(t *testing.T) {
	userID := GenerateULID()
	db := &mockAuthDB{
		users: []map[string]any{
			{"id": userID, "role": "user", "can_write": false},
		},
	}

	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	// Create token that expired 1 second ago
	token := createTestJWT(t, userID, "test-jti", "user", false, -1)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_RevokedJTI(t *testing.T) {
	userID := GenerateULID()
	db := &mockAuthDB{
		users: []map[string]any{
			{"id": userID, "role": "admin", "can_write": true},
		},
	}

	jtiStore := NewJTIRevocationStore()
	jtiStore.Revoke("revoked-jti")
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	token := createTestJWT(t, userID, "revoked-jti", "admin", true, 3600)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_JWT_UserNotFound(t *testing.T) {
	db := &mockAuthDB{users: nil}

	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	token := createTestJWT(t, "nonexistent-user", "test-jti", "admin", true, 3600)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_JWT_WrongSecret(t *testing.T) {
	userID := GenerateULID()
	db := &mockAuthDB{
		users: []map[string]any{
			{"id": userID, "role": "admin", "can_write": true},
		},
	}

	// Sign with different secret
	token, _, err := CreateAccessToken(userID, "test-jti", "admin", true, "wrong-secret-that-is-at-least-32-chars", 3600)
	if err != nil {
		t.Fatal(err)
	}

	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_JWT_MissingJTI(t *testing.T) {
	userID := GenerateULID()
	db := &mockAuthDB{
		users: []map[string]any{
			{"id": userID, "role": "admin", "can_write": true},
		},
	}

	// Create JWT without jti
	token := createTestJWT(t, userID, "", "admin", true, 3600)

	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// API key authentication tests
// ---------------------------------------------------------------------------

func TestAuthenticate_ValidAPIKey(t *testing.T) {
	raw, hash := createTestAPIKey()
	keyID := GenerateULID()
	db := &mockAuthDB{
		apikeys: []map[string]any{
			{"id": keyID, "key_hash": hash, "role": "user", "can_write": true},
		},
	}

	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["credential_type"] != CredentialTypeAPIKey {
		t.Fatalf("expected apikey, got %v", body["credential_type"])
	}
	if body["caller_id"] != keyID {
		t.Fatalf("expected %s, got %v", keyID, body["caller_id"])
	}
}

func TestAuthenticate_APIKey_WrongLength(t *testing.T) {
	db := &mockAuthDB{}
	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	shortKey := APIKeyPrefix + "tooshort"
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+shortKey)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_APIKey_NotFound(t *testing.T) {
	raw, _ := createTestAPIKey()
	db := &mockAuthDB{apikeys: nil}

	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_APIKey_UpdatesLastUsedAt(t *testing.T) {
	raw, hash := createTestAPIKey()
	keyID := GenerateULID()
	db := &mockAuthDB{
		apikeys: []map[string]any{
			{"id": keyID, "key_hash": hash, "role": "user", "can_write": false},
		},
	}

	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)

	inner := testAuthHandler()
	handler := am.Authenticate(inner)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if len(db.updates) == 0 {
		t.Fatal("expected last_used_at update")
	}
	upd := db.updates[0]
	if upd.table != "apikeys" || upd.id != keyID {
		t.Fatalf("unexpected update: %+v", upd)
	}
	if _, ok := upd.data["last_used_at"]; !ok {
		t.Fatal("expected last_used_at in update data")
	}
}

// ---------------------------------------------------------------------------
// API key rejected on auth:me
// ---------------------------------------------------------------------------

func TestAuthenticate_APIKeyRejectedOnAuthMe(t *testing.T) {
	raw, hash := createTestAPIKey()
	keyID := GenerateULID()
	db := &mockAuthDB{
		apikeys: []map[string]any{
			{"id": keyID, "key_hash": hash, "role": "admin", "can_write": true},
		},
	}

	handler := buildAuthTestServer(t, db, "")

	paths := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/auth:me"},
		{http.MethodPost, "/auth:me"},
	}

	for _, tt := range paths {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+raw)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Authorization tests
// ---------------------------------------------------------------------------

func TestAuthorize_AdminCanDoEverything(t *testing.T) {
	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "admin1",
		Role:           "admin",
		CanWrite:       true,
	}

	paths := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/collections:query"},
		{http.MethodPost, "/collections:mutate"},
		{http.MethodGet, "/data/products:query"},
		{http.MethodPost, "/data/products:mutate"},
		{http.MethodPost, "/data/users:mutate"},
		{http.MethodPost, "/data/apikeys:mutate"},
	}

	for _, tt := range paths {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := Authorize("", inner)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			ctx := SetAuthIdentity(req.Context(), identity)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
		})
	}
}

func TestAuthorize_UserCanWriteFalse_ReadAllowed(t *testing.T) {
	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user1",
		Role:           "user",
		CanWrite:       false,
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Authorize("", inner)

	req := httptest.NewRequest(http.MethodGet, "/data/products:query", nil)
	ctx := SetAuthIdentity(req.Context(), identity)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthorize_UserCanWriteFalse_MutateForbidden(t *testing.T) {
	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user1",
		Role:           "user",
		CanWrite:       false,
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Authorize("", inner)

	req := httptest.NewRequest(http.MethodPost, "/data/products:mutate", nil)
	ctx := SetAuthIdentity(req.Context(), identity)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	var body ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Message != "Forbidden" {
		t.Fatalf("expected 'Forbidden', got %q", body.Message)
	}
}

func TestAuthorize_UserCanWriteTrue_MutateAllowed(t *testing.T) {
	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user1",
		Role:           "user",
		CanWrite:       true,
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Authorize("", inner)

	req := httptest.NewRequest(http.MethodPost, "/data/products:mutate", nil)
	ctx := SetAuthIdentity(req.Context(), identity)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthorize_UserCanWriteTrue_CollectionMutateForbidden(t *testing.T) {
	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user1",
		Role:           "user",
		CanWrite:       true,
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Authorize("", inner)

	req := httptest.NewRequest(http.MethodPost, "/collections:mutate", nil)
	ctx := SetAuthIdentity(req.Context(), identity)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestAuthorize_UserCannotMutateSystemCollections(t *testing.T) {
	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user1",
		Role:           "user",
		CanWrite:       true,
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Authorize("", inner)

	systemRoutes := []string{"/data/users:mutate", "/data/apikeys:mutate"}
	for _, path := range systemRoutes {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			ctx := SetAuthIdentity(req.Context(), identity)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d for %s", w.Code, path)
			}
		})
	}
}

func TestAuthorize_WithPrefix(t *testing.T) {
	identity := &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       "user1",
		Role:           "user",
		CanWrite:       false,
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Authorize("/api", inner)

	// Read should be OK
	req := httptest.NewRequest(http.MethodGet, "/api/data/products:query", nil)
	ctx := SetAuthIdentity(req.Context(), identity)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Write should be forbidden
	req = httptest.NewRequest(http.MethodPost, "/api/data/products:mutate", nil)
	ctx = SetAuthIdentity(req.Context(), identity)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Integration tests - full middleware chain
// ---------------------------------------------------------------------------

func TestIntegration_ProtectedRoute_JWT_AdminAccess(t *testing.T) {
	userID := GenerateULID()
	db := &mockAuthDB{
		users: []map[string]any{
			{"id": userID, "role": "admin", "can_write": true},
		},
	}

	cfg := &AppConfig{
		Server:    ServerConfig{Prefix: ""},
		CORS:      CORSConfig{Enabled: false},
		JWTSecret: testJWTSecret(),
	}
	logger := NewTestLogger(&bytes.Buffer{})
	mux := NewRouter("", logger, db, cfg)
	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)
	handler := BuildHandler(mux, cfg, logger, WithAuthMiddleware(am))

	token := createTestJWT(t, userID, GenerateULID(), "admin", true, 3600)

	req := httptest.NewRequest(http.MethodGet, "/collections:query", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should get past auth (501 = handler stub)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestIntegration_ProtectedRoute_JWT_UserReadOnly(t *testing.T) {
	userID := GenerateULID()
	db := &mockAuthDB{
		users: []map[string]any{
			{"id": userID, "role": "user", "can_write": false},
		},
	}

	cfg := &AppConfig{
		Server:    ServerConfig{Prefix: ""},
		CORS:      CORSConfig{Enabled: false},
		JWTSecret: testJWTSecret(),
	}
	logger := NewTestLogger(&bytes.Buffer{})
	mux := NewRouter("", logger, db, cfg)
	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)
	handler := BuildHandler(mux, cfg, logger, WithAuthMiddleware(am))

	token := createTestJWT(t, userID, GenerateULID(), "user", false, 3600)

	// Read should work
	req := httptest.NewRequest(http.MethodGet, "/data/products:query", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}

	// Write should be forbidden
	req = httptest.NewRequest(http.MethodPost, "/data/products:mutate", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestIntegration_APIKey_AuthMe_Rejected(t *testing.T) {
	raw, hash := createTestAPIKey()
	keyID := GenerateULID()
	db := &mockAuthDB{
		apikeys: []map[string]any{
			{"id": keyID, "key_hash": hash, "role": "admin", "can_write": true},
		},
	}

	cfg := &AppConfig{
		Server:    ServerConfig{Prefix: ""},
		CORS:      CORSConfig{Enabled: false},
		JWTSecret: testJWTSecret(),
	}
	logger := NewTestLogger(&bytes.Buffer{})
	mux := NewRouter("", logger, db, cfg)
	jtiStore := NewJTIRevocationStore()
	am := NewAuthMiddleware(db, testJWTSecret(), "", jtiStore)
	handler := BuildHandler(mux, cfg, logger, WithAuthMiddleware(am))

	req := httptest.NewRequest(http.MethodGet, "/auth:me", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestIntegration_NoAuthMiddleware_SkipsAuth(t *testing.T) {
	cfg := defaultTestConfig()
	logger := NewTestLogger(&bytes.Buffer{})
	mux := NewRouter(cfg.Server.Prefix, logger, nil, cfg)
	handler := BuildHandler(mux, cfg, logger)

	// Without auth middleware, routes should work without token
	req := httptest.NewRequest(http.MethodGet, "/auth:me", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Without auth middleware, the handler still requires JWT identity → 401
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Constants tests
// ---------------------------------------------------------------------------

func TestConstants(t *testing.T) {
	if APIKeyPrefix != "moon_live_" {
		t.Fatalf("APIKeyPrefix = %q, want %q", APIKeyPrefix, "moon_live_")
	}
	if APIKeyTotalLen != 74 {
		t.Fatalf("APIKeyTotalLen = %d, want 74", APIKeyTotalLen)
	}
	if CredentialTypeJWT != "jwt" {
		t.Fatalf("CredentialTypeJWT = %q, want %q", CredentialTypeJWT, "jwt")
	}
	if CredentialTypeAPIKey != "apikey" {
		t.Fatalf("CredentialTypeAPIKey = %q, want %q", CredentialTypeAPIKey, "apikey")
	}
}

// ---------------------------------------------------------------------------
// isPublicRoute edge cases
// ---------------------------------------------------------------------------

func TestIsPublicRoute_GETAuthSessionIsNotPublic(t *testing.T) {
	am := NewAuthMiddleware(nil, testJWTSecret(), "", NewJTIRevocationStore())

	req := httptest.NewRequest(http.MethodGet, "/auth:session", nil)
	if am.isPublicRoute(req) {
		t.Fatal("GET /auth:session should not be public")
	}
}

func TestIsPublicRoute_PostHealthIsNotPublic(t *testing.T) {
	am := NewAuthMiddleware(nil, testJWTSecret(), "", NewJTIRevocationStore())

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	if am.isPublicRoute(req) {
		t.Fatal("POST /health should not be public")
	}
}

// ---------------------------------------------------------------------------
// Validate that Revoke is immediate
// ---------------------------------------------------------------------------

func TestJTIRevocation_Immediate(t *testing.T) {
	store := NewJTIRevocationStore()

	jti := "immediate-test"
	if store.IsRevoked(jti) {
		t.Fatal("should not be revoked before Revoke call")
	}

	store.Revoke(jti)

	// Must be revoked immediately, no delay
	if !store.IsRevoked(jti) {
		t.Fatal("should be revoked immediately after Revoke call")
	}

	// Verify timestamp is recent
	store.mu.RLock()
	ts := store.store[jti]
	store.mu.RUnlock()
	if time.Since(ts) > time.Second {
		t.Fatal("revocation timestamp should be very recent")
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func defaultTestConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Host:   DefaultServerHost,
			Port:   DefaultServerPort,
			Prefix: DefaultServerPrefix,
		},
		CORS: CORSConfig{
			Enabled:        true,
			AllowedOrigins: []string{"*"},
		},
	}
}

func buildTestServer(t *testing.T, cfg *AppConfig) http.Handler {
	t.Helper()
	logger := NewTestLogger(&bytes.Buffer{})
	mux := NewRouter(cfg.Server.Prefix, logger, nil, cfg)
	return BuildHandler(mux, cfg, logger)
}

// --- Health endpoint tests ---

func TestHealthEndpoint(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"root", "/", http.StatusOK},
		{"health", "/health", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestHealthEndpoint_WithPrefix(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Server.Prefix = "/api"
	handler := buildTestServer(t, cfg)

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"prefixed root", "/api", http.StatusOK},
		{"prefixed root trailing", "/api/", http.StatusOK},
		{"prefixed health", "/api/health", http.StatusOK},
		{"unprefixed root", "/", http.StatusNotFound},
		{"unprefixed health", "/health", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Fatalf("%s: expected %d, got %d", tt.name, tt.status, w.Code)
			}
		})
	}
}

// --- Method validation tests ---

func TestMethodNotAllowed(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	methods := []string{http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Fatalf("expected 405, got %d", w.Code)
			}

			var got ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}
			if got.Message != "Method not allowed" {
				t.Fatalf("expected 'Method not allowed', got %q", got.Message)
			}
		})
	}
}

// --- 404 tests ---

func TestUnknownRoute_404(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Auth route tests ---

func TestAuthSessionRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodPost, "/auth:session", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAuthMeGetRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/auth:me", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

func TestAuthMePostRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodPost, "/auth:me", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

// --- Collection route tests ---

func TestCollectionsQueryRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/collections:query", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

func TestCollectionsMutateRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodPost, "/collections:mutate", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

// --- Resource route tests ---

func TestResourceQueryRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/data/products:query", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

func TestResourceMutateRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodPost, "/data/products:mutate", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

func TestResourceSchemaRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/data/products:schema", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

// --- Resource validation tests ---

func TestResourceMoonPrefixRejected(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/data/moon_internal:query", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var got ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got.Message == "" {
		t.Fatal("expected error message")
	}
}

func TestResourceMoonPrefixRejected_POST(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodPost, "/data/moon_tokens:mutate", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Data route without action ---

func TestDataRouteWithoutAction_404(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/data/products", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Data route with unknown action ---

func TestDataRouteWithUnknownAction_404(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/data/products:unknown", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Prefix tests for all routes ---

func TestPrefixedRoutes(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Server.Prefix = "/api/v1"
	handler := buildTestServer(t, cfg)

	tests := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"health", http.MethodGet, "/api/v1/health", http.StatusOK},
		{"root", http.MethodGet, "/api/v1", http.StatusOK},
		{"auth session", http.MethodPost, "/api/v1/auth:session", http.StatusBadRequest},
		{"auth me get", http.MethodGet, "/api/v1/auth:me", http.StatusNotImplemented},
		{"auth me post", http.MethodPost, "/api/v1/auth:me", http.StatusNotImplemented},
		{"collections query", http.MethodGet, "/api/v1/collections:query", http.StatusNotImplemented},
		{"collections mutate", http.MethodPost, "/api/v1/collections:mutate", http.StatusNotImplemented},
		{"resource query", http.MethodGet, "/api/v1/data/products:query", http.StatusNotImplemented},
		{"resource mutate", http.MethodPost, "/api/v1/data/products:mutate", http.StatusNotImplemented},
		{"resource schema", http.MethodGet, "/api/v1/data/products:schema", http.StatusNotImplemented},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Fatalf("%s %s: expected %d, got %d", tt.method, tt.path, tt.status, w.Code)
			}
		})
	}
}

// --- CORS integration with routes ---

func TestCORS_OptionsPreflightOnRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected *, got %q", got)
	}
}

// --- Panic recovery integration ---

func TestPanicRecoveryIntegration(t *testing.T) {
	cfg := defaultTestConfig()
	logger := NewTestLogger(&bytes.Buffer{})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /panic-test", func(w http.ResponseWriter, r *http.Request) {
		panic("integration test panic")
	})
	handler := BuildHandler(mux, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/panic-test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// --- Wrong method on specific route ---

func TestWrongMethodOnRoute(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	// GET on auth:session — only POST is registered, so this is 404 or 405
	req := httptest.NewRequest(http.MethodGet, "/auth:session", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// ServeMux may return 404 when pattern doesn't match method
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 404 or 405, got %d", w.Code)
	}
}

// --- Resource missing name ---

func TestResourceMissingName(t *testing.T) {
	handler := buildTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/data/:query", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

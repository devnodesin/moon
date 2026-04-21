package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func middlewareTestLogger() *Logger {
	return NewTestLogger(&bytes.Buffer{})
}

func TestMethodValidationMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := methodValidationMiddleware(inner)

	tests := []struct {
		method     string
		wantStatus int
	}{
		{http.MethodGet, http.StatusOK},
		{http.MethodPost, http.StatusOK},
		{http.MethodOptions, http.StatusOK},
		{http.MethodPut, http.StatusMethodNotAllowed},
		{http.MethodDelete, http.StatusMethodNotAllowed},
		{http.MethodPatch, http.StatusMethodNotAllowed},
		{http.MethodHead, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("method %s: expected %d, got %d", tt.method, tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusMethodNotAllowed {
				var got ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode: %v", err)
				}
				if got.Message != "Method not allowed" {
					t.Fatalf("expected 'Method not allowed', got %q", got.Message)
				}
				if allow := w.Header().Get("Allow"); allow != "GET, POST, OPTIONS" {
					t.Fatalf("expected Allow header, got %q", allow)
				}
			}
		})
	}
}

func TestCORSMiddleware_Enabled(t *testing.T) {
	cfg := CORSConfig{Enabled: true, AllowedOrigins: []string{"*"}}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(cfg, inner)

	t.Run("adds CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("expected *, got %q", got)
		}
	})

	t.Run("OPTIONS preflight returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("expected *, got %q", got)
		}
	})

	t.Run("no Origin header means no CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("expected no CORS header, got %q", got)
		}
	})
}

func TestCORSMiddleware_Disabled(t *testing.T) {
	cfg := CORSConfig{Enabled: false}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(cfg, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS header when disabled, got %q", got)
	}
}

func TestCORSMiddleware_SpecificOrigins(t *testing.T) {
	cfg := CORSConfig{Enabled: true, AllowedOrigins: []string{"http://allowed.com"}}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(cfg, inner)

	t.Run("matching origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://allowed.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://allowed.com" {
			t.Fatalf("expected http://allowed.com, got %q", got)
		}
	})

	t.Run("non-matching origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://notallowed.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("expected no CORS header, got %q", got)
		}
	})
}

func TestWebsiteAPIKeyMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := websiteAPIKeyMiddleware(inner)

	t.Run("allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/data/contact:query", nil)
		req.Header.Set("Origin", "https://example.com")
		req = req.WithContext(SetAuthIdentity(req.Context(), &AuthIdentity{
			CredentialType: CredentialTypeAPIKey,
			CallerID:       "key-1",
			IsWebsite:      true,
			AllowedOrigins: []string{"https://example.com"},
		}))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("missing origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/data/contact:query", nil)
		req = req.WithContext(SetAuthIdentity(req.Context(), &AuthIdentity{
			CredentialType: CredentialTypeAPIKey,
			CallerID:       "key-1",
			IsWebsite:      true,
			AllowedOrigins: []string{"https://example.com"},
		}))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}

func TestPanicRecoveryMiddleware(t *testing.T) {
	logger := middlewareTestLogger()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	handler := panicRecoveryMiddleware(logger, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var got ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got.Message != "Internal server error" {
		t.Fatalf("expected 'Internal server error', got %q", got.Message)
	}
}

func TestAuditContextMiddleware(t *testing.T) {
	logger := middlewareTestLogger()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := auditContextMiddleware(logger, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestExtractResource(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/data/products:query", "products"},
		{"/data/orders:mutate", "orders"},
		{"/data/items:schema", "items"},
		{"/api/data/products:query", "products"},
		{"/data/moon_internal:query", "moon_internal"},
		{"/health", ""},
		{"/data/", ""},
		{"/data/products", "products"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractResource(tt.path)
			if got != tt.want {
				t.Fatalf("extractResource(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestMatchOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed []string
		want    string
	}{
		{"wildcard", "http://example.com", []string{"*"}, "*"},
		{"exact match", "http://example.com", []string{"http://example.com"}, "http://example.com"},
		{"no match", "http://bad.com", []string{"http://good.com"}, ""},
		{"empty origin", "", []string{"*"}, ""},
		{"case insensitive", "HTTP://EXAMPLE.COM", []string{"http://example.com"}, "HTTP://EXAMPLE.COM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchOrigin(tt.origin, tt.allowed)
			if got != tt.want {
				t.Fatalf("matchOrigin(%q, %v) = %q, want %q", tt.origin, tt.allowed, got, tt.want)
			}
		})
	}
}

func TestExtractCaptchaFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/data/contact:mutate", bytes.NewBufferString(`{"captcha_id":"abc","captcha_value":"123456","op":"create"}`))
	id, value, ok := extractCaptchaFields(req)
	if !ok {
		t.Fatal("expected captcha fields to parse")
	}
	if id != "abc" || value != "123456" {
		t.Fatalf("unexpected captcha fields: %q %q", id, value)
	}
}

func TestResourceValidationMiddleware_RejectsReservedPrefix(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := resourceValidationMiddleware(inner)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"reserved moon_ prefix blocked", "/data/moon_internal:query", http.StatusBadRequest},
		{"reserved moon_ prefix mutate blocked", "/data/moon_system:mutate", http.StatusBadRequest},
		{"normal resource allowed", "/data/products:query", http.StatusOK},
		{"health route allowed", "/health", http.StatusOK},
		{"no data segment allowed", "/api/v1/status", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("path %s: expected status %d, got %d", tt.path, tt.wantStatus, w.Code)
			}
		})
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	bucket := NewTokenBucket(100, 10)

	if bucket == nil {
		t.Fatal("NewTokenBucket returned nil")
	}
	if bucket.maxTokens != 100 {
		t.Errorf("Expected maxTokens 100, got %d", bucket.maxTokens)
	}
	if bucket.refillRate != 10 {
		t.Errorf("Expected refillRate 10, got %d", bucket.refillRate)
	}
	if bucket.tokens != 100 {
		t.Errorf("Expected tokens 100, got %d", bucket.tokens)
	}
}

func TestTokenBucket_Allow(t *testing.T) {
	bucket := NewTokenBucket(3, 1)

	// Should allow first 3 requests
	for i := 0; i < 3; i++ {
		if !bucket.Allow() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if bucket.Allow() {
		t.Error("4th request should be denied")
	}
}

func TestTokenBucket_Remaining(t *testing.T) {
	bucket := NewTokenBucket(10, 1)

	if bucket.Remaining() != 10 {
		t.Errorf("Expected 10 remaining, got %d", bucket.Remaining())
	}

	bucket.Allow()
	bucket.Allow()

	if bucket.Remaining() != 8 {
		t.Errorf("Expected 8 remaining, got %d", bucket.Remaining())
	}
}

func TestTokenBucket_Limit(t *testing.T) {
	bucket := NewTokenBucket(100, 10)

	if bucket.Limit() != 100 {
		t.Errorf("Expected limit 100, got %d", bucket.Limit())
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	// Create bucket with 1 token per second refill
	bucket := NewTokenBucket(5, 5)

	// Use all tokens
	for i := 0; i < 5; i++ {
		bucket.Allow()
	}

	if bucket.Remaining() != 0 {
		t.Errorf("Expected 0 remaining, got %d", bucket.Remaining())
	}

	// Wait for refill (at least 1 second for 5 tokens at 5/sec)
	time.Sleep(1100 * time.Millisecond)

	remaining := bucket.Remaining()
	if remaining < 4 {
		t.Errorf("Expected at least 4 tokens after refill, got %d", remaining)
	}
}

func TestRateLimiter_NewRateLimiter(t *testing.T) {
	config := RateLimiterConfig{
		UserRPM:   100,
		APIKeyRPM: 1000,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	if rl.userRPM != 100 {
		t.Errorf("Expected userRPM 100, got %d", rl.userRPM)
	}
	if rl.apiKeyRPM != 1000 {
		t.Errorf("Expected apiKeyRPM 1000, got %d", rl.apiKeyRPM)
	}
}

func TestRateLimiter_DefaultValues(t *testing.T) {
	config := RateLimiterConfig{} // zero values

	rl := NewRateLimiter(config)
	defer rl.Stop()

	if rl.userRPM != 100 {
		t.Errorf("Expected default userRPM 100, got %d", rl.userRPM)
	}
	if rl.apiKeyRPM != 1000 {
		t.Errorf("Expected default apiKeyRPM 1000, got %d", rl.apiKeyRPM)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	config := RateLimiterConfig{
		UserRPM:   10, // small for testing
		APIKeyRPM: 20,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	t.Run("user rate limit", func(t *testing.T) {
		entityID := "user-test-1"

		// First requests should be allowed
		for i := 0; i < 5; i++ {
			allowed, remaining, _, limit := rl.Allow(entityID, EntityTypeUser)
			if !allowed {
				t.Errorf("Request %d should be allowed", i+1)
			}
			if limit != 10 {
				t.Errorf("Expected limit 10, got %d", limit)
			}
			expectedRemaining := 10 - (i + 1)
			if remaining != expectedRemaining {
				t.Errorf("Expected remaining %d, got %d", expectedRemaining, remaining)
			}
		}
	})

	t.Run("API key rate limit", func(t *testing.T) {
		entityID := "apikey-test-1"

		allowed, _, _, limit := rl.Allow(entityID, EntityTypeAPIKey)
		if !allowed {
			t.Error("First request should be allowed")
		}
		if limit != 20 {
			t.Errorf("Expected limit 20, got %d", limit)
		}
	})
}

func TestRateLimitMiddleware_Headers(t *testing.T) {
	config := RateLimiterConfig{
		UserRPM:   100,
		APIKeyRPM: 1000,
	}

	m := NewRateLimitMiddleware(config)
	defer m.Stop()

	entity := &AuthEntity{
		ID:   "test-user-headers",
		Type: EntityTypeUser,
		Role: "user",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(SetAuthEntity(req.Context(), entity))
	w := httptest.NewRecorder()

	m.RateLimit(handler)(w, req)

	// Check headers
	limitHeader := w.Header().Get(HeaderRateLimitLimit)
	if limitHeader == "" {
		t.Error("X-RateLimit-Limit header not set")
	}
	limit, _ := strconv.Atoi(limitHeader)
	if limit != 100 {
		t.Errorf("Expected limit 100, got %d", limit)
	}

	remainingHeader := w.Header().Get(HeaderRateLimitRemaining)
	if remainingHeader == "" {
		t.Error("X-RateLimit-Remaining header not set")
	}

	resetHeader := w.Header().Get(HeaderRateLimitReset)
	if resetHeader == "" {
		t.Error("X-RateLimit-Reset header not set")
	}
}

func TestRateLimitMiddleware_NoAuthEntity(t *testing.T) {
	config := RateLimiterConfig{
		UserRPM:   100,
		APIKeyRPM: 1000,
	}

	m := NewRateLimitMiddleware(config)
	defer m.Stop()

	handlerCalled := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	m.RateLimit(handler)(w, req)

	if !handlerCalled {
		t.Error("Handler should be called when no auth entity")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRateLimitMiddleware_ExceedLimit(t *testing.T) {
	config := RateLimiterConfig{
		UserRPM:   3, // very small for testing
		APIKeyRPM: 3,
	}

	m := NewRateLimitMiddleware(config)
	defer m.Stop()

	entity := &AuthEntity{
		ID:   "test-exceed-limit",
		Type: EntityTypeUser,
		Role: "user",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// Make requests until limit is exceeded
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(SetAuthEntity(req.Context(), entity))
		w := httptest.NewRecorder()

		m.RateLimit(handler)(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed, got status %d", i+1, w.Code)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(SetAuthEntity(req.Context(), entity))
	w := httptest.NewRecorder()

	m.RateLimit(handler)(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestLoginRateLimiter_NewLoginRateLimiter(t *testing.T) {
	config := LoginRateLimiterConfig{
		MaxAttempts:   5,
		WindowSeconds: 900,
	}

	lrl := NewLoginRateLimiter(config)
	defer lrl.Stop()

	if lrl == nil {
		t.Fatal("NewLoginRateLimiter returned nil")
	}
	if lrl.maxAttempts != 5 {
		t.Errorf("Expected maxAttempts 5, got %d", lrl.maxAttempts)
	}
	if lrl.windowSeconds != 900 {
		t.Errorf("Expected windowSeconds 900, got %d", lrl.windowSeconds)
	}
}

func TestLoginRateLimiter_DefaultValues(t *testing.T) {
	config := LoginRateLimiterConfig{}

	lrl := NewLoginRateLimiter(config)
	defer lrl.Stop()

	if lrl.maxAttempts != 5 {
		t.Errorf("Expected default maxAttempts 5, got %d", lrl.maxAttempts)
	}
	if lrl.windowSeconds != 900 {
		t.Errorf("Expected default windowSeconds 900, got %d", lrl.windowSeconds)
	}
}

func TestLoginRateLimiter_CheckAndRecord(t *testing.T) {
	config := LoginRateLimiterConfig{
		MaxAttempts:   3,
		WindowSeconds: 60,
	}

	lrl := NewLoginRateLimiter(config)
	defer lrl.Stop()

	ip := "192.168.1.1"
	username := "testuser"

	// First 3 attempts should be allowed
	for i := 0; i < 3; i++ {
		allowed, remaining, _ := lrl.CheckAndRecord(ip, username)
		if !allowed {
			t.Errorf("Attempt %d should be allowed", i+1)
		}
		expectedRemaining := 3 - (i + 1)
		if remaining != expectedRemaining {
			t.Errorf("Expected remaining %d, got %d", expectedRemaining, remaining)
		}
	}

	// 4th attempt should be denied
	allowed, remaining, _ := lrl.CheckAndRecord(ip, username)
	if allowed {
		t.Error("4th attempt should be denied")
	}
	if remaining != 0 {
		t.Errorf("Expected 0 remaining, got %d", remaining)
	}
}

func TestLoginRateLimiter_IsBlocked(t *testing.T) {
	config := LoginRateLimiterConfig{
		MaxAttempts:   2,
		WindowSeconds: 60,
	}

	lrl := NewLoginRateLimiter(config)
	defer lrl.Stop()

	ip := "192.168.1.2"
	username := "blockeduser"

	// Initially not blocked
	blocked, _ := lrl.IsBlocked(ip, username)
	if blocked {
		t.Error("User should not be blocked initially")
	}

	// Use all attempts
	lrl.CheckAndRecord(ip, username)
	lrl.CheckAndRecord(ip, username)

	// Should be blocked now
	blocked, resetAt := lrl.IsBlocked(ip, username)
	if !blocked {
		t.Error("User should be blocked after max attempts")
	}
	if resetAt.IsZero() {
		t.Error("Reset time should not be zero")
	}
}

func TestLoginRateLimiter_ResetForUser(t *testing.T) {
	config := LoginRateLimiterConfig{
		MaxAttempts:   2,
		WindowSeconds: 60,
	}

	lrl := NewLoginRateLimiter(config)
	defer lrl.Stop()

	ip := "192.168.1.3"
	username := "resetuser"

	// Use attempts
	lrl.CheckAndRecord(ip, username)
	lrl.CheckAndRecord(ip, username)

	// Should be blocked
	blocked, _ := lrl.IsBlocked(ip, username)
	if !blocked {
		t.Error("User should be blocked")
	}

	// Reset
	lrl.ResetForUser(ip, username)

	// Should not be blocked after reset
	blocked, _ = lrl.IsBlocked(ip, username)
	if blocked {
		t.Error("User should not be blocked after reset")
	}

	// Should be able to make attempts again
	allowed, _, _ := lrl.CheckAndRecord(ip, username)
	if !allowed {
		t.Error("Attempt should be allowed after reset")
	}
}

func TestLoginRateLimiter_DifferentUsers(t *testing.T) {
	config := LoginRateLimiterConfig{
		MaxAttempts:   2,
		WindowSeconds: 60,
	}

	lrl := NewLoginRateLimiter(config)
	defer lrl.Stop()

	ip := "192.168.1.4"
	user1 := "user1"
	user2 := "user2"

	// Block user1
	lrl.CheckAndRecord(ip, user1)
	lrl.CheckAndRecord(ip, user1)

	blocked, _ := lrl.IsBlocked(ip, user1)
	if !blocked {
		t.Error("User1 should be blocked")
	}

	// User2 should not be blocked
	blocked, _ = lrl.IsBlocked(ip, user2)
	if blocked {
		t.Error("User2 should not be blocked")
	}

	allowed, _, _ := lrl.CheckAndRecord(ip, user2)
	if !allowed {
		t.Error("User2 attempt should be allowed")
	}
}

func TestLoginRateLimiter_DifferentIPs(t *testing.T) {
	config := LoginRateLimiterConfig{
		MaxAttempts:   2,
		WindowSeconds: 60,
	}

	lrl := NewLoginRateLimiter(config)
	defer lrl.Stop()

	ip1 := "192.168.1.5"
	ip2 := "192.168.1.6"
	username := "sameuser"

	// Block from ip1
	lrl.CheckAndRecord(ip1, username)
	lrl.CheckAndRecord(ip1, username)

	blocked, _ := lrl.IsBlocked(ip1, username)
	if !blocked {
		t.Error("Should be blocked from ip1")
	}

	// Same user from different IP should not be blocked
	blocked, _ = lrl.IsBlocked(ip2, username)
	if blocked {
		t.Error("Should not be blocked from ip2")
	}

	allowed, _, _ := lrl.CheckAndRecord(ip2, username)
	if !allowed {
		t.Error("Attempt from ip2 should be allowed")
	}
}

func TestWriteLoginRateLimitError(t *testing.T) {
	w := httptest.NewRecorder()
	resetAt := time.Now().Add(15 * time.Minute)

	WriteLoginRateLimitError(w, resetAt)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// slidingWindowLimiter tests
// ---------------------------------------------------------------------------

func TestSlidingWindowLimiter_Allow(t *testing.T) {
	l := newSlidingWindowLimiter(3, time.Minute)

	// First three requests must be allowed.
	for i := range 3 {
		if !l.Allow("key") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	// Fourth request must be denied.
	if l.Allow("key") {
		t.Fatal("fourth request should be denied")
	}
}

func TestSlidingWindowLimiter_Allow_DifferentKeys(t *testing.T) {
	l := newSlidingWindowLimiter(1, time.Minute)

	if !l.Allow("a") {
		t.Fatal("first request for key 'a' should be allowed")
	}
	if !l.Allow("b") {
		t.Fatal("first request for key 'b' should be allowed")
	}
	if l.Allow("a") {
		t.Fatal("second request for key 'a' should be denied")
	}
}

func TestSlidingWindowLimiter_IsExceeded(t *testing.T) {
	l := newSlidingWindowLimiter(2, time.Minute)

	if l.IsExceeded("key") {
		t.Fatal("should not be exceeded before any hits")
	}

	l.RecordHit("key")
	l.RecordHit("key")

	if !l.IsExceeded("key") {
		t.Fatal("should be exceeded after limit hits")
	}
}

func TestSlidingWindowLimiter_Reset(t *testing.T) {
	l := newSlidingWindowLimiter(1, time.Minute)

	l.Allow("key") // consume the slot

	if l.Allow("key") {
		t.Fatal("second request should be denied before reset")
	}

	l.Reset("key")

	if !l.Allow("key") {
		t.Fatal("first request after reset should be allowed")
	}
}

func TestSlidingWindowLimiter_WindowExpiry(t *testing.T) {
	l := newSlidingWindowLimiter(1, 50*time.Millisecond)

	if !l.Allow("key") {
		t.Fatal("first request should be allowed")
	}
	if l.Allow("key") {
		t.Fatal("second request should be denied within window")
	}

	time.Sleep(60 * time.Millisecond) // wait for the window to expire

	if !l.Allow("key") {
		t.Fatal("request should be allowed after window expiry")
	}
}

func TestSlidingWindowLimiter_ConcurrencySafe(t *testing.T) {
	l := newSlidingWindowLimiter(50, time.Minute)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Allow("key")
		}()
	}
	wg.Wait()

	// Exactly 50 requests should have been allowed, so the next must be denied.
	if l.Allow("key") {
		t.Fatal("request beyond limit should be denied")
	}
}

// ---------------------------------------------------------------------------
// RateLimiter tests
// ---------------------------------------------------------------------------

func TestRateLimiter_LoginFailure(t *testing.T) {
	rl := NewRateLimiter()

	const ip = "127.0.0.1"
	const user = "testuser"

	// Should not be exceeded initially.
	if rl.LoginFailureExceeded(ip, user) {
		t.Fatal("should not be exceeded with zero failures")
	}

	// Record RateLoginFailureLimit failures.
	for range RateLoginFailureLimit {
		rl.RecordLoginFailure(ip, user)
	}

	// Now must be exceeded.
	if !rl.LoginFailureExceeded(ip, user) {
		t.Fatal("should be exceeded after limit failures")
	}
}

func TestRateLimiter_LoginFailureReset(t *testing.T) {
	rl := NewRateLimiter()

	const ip = "10.0.0.1"
	const user = "alice"

	for range RateLoginFailureLimit {
		rl.RecordLoginFailure(ip, user)
	}

	rl.ResetLoginFailures(ip, user)

	if rl.LoginFailureExceeded(ip, user) {
		t.Fatal("should not be exceeded after reset")
	}
}

func TestRateLimiter_LoginFailure_CaseInsensitive(t *testing.T) {
	rl := NewRateLimiter()

	const ip = "192.168.1.1"

	for range RateLoginFailureLimit {
		rl.RecordLoginFailure(ip, "Admin")
	}

	// "admin" (lowercase) must share the same counter.
	if !rl.LoginFailureExceeded(ip, "admin") {
		t.Fatal("login failure key must be case-insensitive")
	}
}

func TestRateLimiter_JWT(t *testing.T) {
	rl := NewRateLimiter()

	const userID = "01TESTUSER000000000000001"

	for range RateJWTRequestLimit {
		if !rl.AllowJWT(userID) {
			t.Fatalf("request within limit should be allowed")
		}
	}

	if rl.AllowJWT(userID) {
		t.Fatal("request beyond limit should be denied")
	}
}

func TestRateLimiter_APIKey(t *testing.T) {
	rl := NewRateLimiter()

	const keyID = "01TESTAPIKEY0000000000001"

	for range RateAPIKeyRequestLimit {
		if !rl.AllowAPIKey(keyID) {
			t.Fatalf("request within limit should be allowed")
		}
	}

	if rl.AllowAPIKey(keyID) {
		t.Fatal("request beyond limit should be denied")
	}
}

// ---------------------------------------------------------------------------
// clientIP tests
// ---------------------------------------------------------------------------

func TestClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	got := clientIP(req)
	if got != "192.168.1.50" {
		t.Fatalf("expected 192.168.1.50, got %q", got)
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.1")

	got := clientIP(req)
	if got != "203.0.113.10" {
		t.Fatalf("expected 203.0.113.10, got %q", got)
	}
}

func TestClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	req.Header.Set("X-Real-IP", "203.0.113.20")

	got := clientIP(req)
	if got != "203.0.113.20" {
		t.Fatalf("expected 203.0.113.20, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// rateLimitMiddleware integration tests
// ---------------------------------------------------------------------------

func TestRateLimitMiddleware_JWT(t *testing.T) {
	rl := NewRateLimiter()
	logger := middlewareTestLogger()

	// Exhaust the JWT limit for userID directly.
	userID := "01JWTUSER000000000000001"
	for range RateJWTRequestLimit {
		rl.AllowJWT(userID)
	}

	// The next request through the middleware should get 429.
	req := httptest.NewRequest("GET", "/data/test:query", nil)
	identity := &AuthIdentity{CredentialType: CredentialTypeJWT, CallerID: userID}
	req = req.WithContext(SetAuthIdentity(req.Context(), identity))

	calls := 0
	w := httptest.NewRecorder()
	handler := rateLimitMiddleware(rl, logger, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		calls++
		rw.WriteHeader(200)
	}))

	handler.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	if calls != 0 {
		t.Fatal("inner handler should not have been called")
	}
}

func TestRateLimitMiddleware_NoIdentity(t *testing.T) {
	rl := NewRateLimiter()
	logger := middlewareTestLogger()

	called := false
	handler := rateLimitMiddleware(rl, logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("inner handler should have been called for unauthenticated request")
	}
}

package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Sliding window rate limiter
// ---------------------------------------------------------------------------

// slidingWindowLimiter is a concurrency-safe in-memory sliding window rate limiter.
type slidingWindowLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

// newSlidingWindowLimiter creates a sliding window limiter with the given limit and window.
func newSlidingWindowLimiter(limit int, window time.Duration) *slidingWindowLimiter {
	return &slidingWindowLimiter{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
}

// Allow returns true if the key is below the limit and records the hit.
// Returns false without recording a hit if the limit is already reached.
func (l *slidingWindowLimiter) Allow(key string) bool {
	return l.AllowWithLimit(key, l.limit)
}

// AllowWithLimit returns true if the key is below the provided limit and
// records the hit. Returns false without recording a hit if the limit is
// already reached.
func (l *slidingWindowLimiter) AllowWithLimit(key string, limit int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)
	l.hits[key] = keepAfter(l.hits[key], cutoff)

	if len(l.hits[key]) >= limit {
		return false
	}
	l.hits[key] = append(l.hits[key], now)
	return true
}

// IsExceeded returns true if the key is at or over the limit without recording a hit.
func (l *slidingWindowLimiter) IsExceeded(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)
	l.hits[key] = keepAfter(l.hits[key], cutoff)
	return len(l.hits[key]) >= l.limit
}

// RecordHit records a hit for key without checking the limit.
func (l *slidingWindowLimiter) RecordHit(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hits[key] = append(l.hits[key], time.Now())
}

// Reset removes all recorded hits for key.
func (l *slidingWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.hits, key)
}

// keepAfter returns the subset of ts that is strictly after cutoff.
// It reuses the underlying array to avoid allocation.
func keepAfter(ts []time.Time, cutoff time.Time) []time.Time {
	out := ts[:0]
	for _, t := range ts {
		if t.After(cutoff) {
			out = append(out, t)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Aggregate rate limiter
// ---------------------------------------------------------------------------

// RateLimiter aggregates rate limiters for the traffic types defined in SPEC.md.
type RateLimiter struct {
	smartLogin         *slidingWindowLimiter
	smartLoginInitiate *slidingWindowLimiter
	jwtRequest         *slidingWindowLimiter
	apikeyRequest      *slidingWindowLimiter
}

// RateLimiter uses in-memory sliding windows.
// NOT safe for multi-instance deployments — use a Redis-backed limiter for production.
//
// NewRateLimiter creates a RateLimiter with limits taken from the constants in
// Config.go.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		smartLogin:         newSlidingWindowLimiter(RateSmartLoginLimit, time.Duration(RateSmartLoginWindow)*time.Second),
		smartLoginInitiate: newSlidingWindowLimiter(RateSmartLoginInitiateLimit, time.Duration(RateSmartLoginInitiateWindow)*time.Second),
		jwtRequest:         newSlidingWindowLimiter(RateJWTRequestLimit, time.Duration(RateJWTRequestWindow)*time.Second),
		apikeyRequest:      newSlidingWindowLimiter(RateAPIKeyRequestLimit, time.Duration(RateAPIKeyRequestWindow)*time.Second),
	}
}

// SmartLoginInitiateExceeded returns true if the initiate rate limit for the given key has been reached.
func (r *RateLimiter) SmartLoginInitiateExceeded(key string) bool {
	if r == nil || r.smartLoginInitiate == nil {
		return false
	}
	return r.smartLoginInitiate.IsExceeded(strings.ToLower(key))
}

// RecordSmartLoginInitiate records one initiate attempt for the given key.
func (r *RateLimiter) RecordSmartLoginInitiate(key string) {
	if r == nil || r.smartLoginInitiate == nil {
		return
	}
	r.smartLoginInitiate.RecordHit(strings.ToLower(key))
}

// ResetSmartLoginInitiate clears recorded initiate attempts for the given key.
func (r *RateLimiter) ResetSmartLoginInitiate(key string) {
	if r == nil || r.smartLoginInitiate == nil {
		return
	}
	r.smartLoginInitiate.Reset(strings.ToLower(key))
}

// SmartLoginExceeded returns true if the smart login limit for the given key has been reached.
func (r *RateLimiter) SmartLoginExceeded(key string) bool {
	if r == nil || r.smartLogin == nil {
		return false
	}
	return r.smartLogin.IsExceeded(strings.ToLower(key))
}

// RecordSmartLogin records one smart login attempt for the given key.
func (r *RateLimiter) RecordSmartLogin(key string) {
	if r == nil || r.smartLogin == nil {
		return
	}
	r.smartLogin.RecordHit(strings.ToLower(key))
}

// ResetSmartLogin clears recorded smart login attempts for the given email.
func (r *RateLimiter) ResetSmartLogin(email string) {
	if r == nil || r.smartLogin == nil {
		return
	}
	r.smartLogin.Reset(strings.ToLower(email))
}

// AllowJWT returns true if the JWT request is within the per-user limit.
func (r *RateLimiter) AllowJWT(userID string) bool {
	return r.jwtRequest.Allow(userID)
}

// AllowAPIKey returns true if the API key request is within the per-key limit.
func (r *RateLimiter) AllowAPIKey(keyID string) bool {
	return r.apikeyRequest.Allow(keyID)
}

// AllowAPIKeyWithLimit returns true if the API key request is within the
// provided per-minute limit.
func (r *RateLimiter) AllowAPIKeyWithLimit(keyID string, limit int) bool {
	return r.apikeyRequest.AllowWithLimit(keyID, limit)
}

// Client IP extraction
// ---------------------------------------------------------------------------

// clientIP extracts the real client IP address from the request, respecting
// X-Forwarded-For and X-Real-IP headers set by reverse proxies.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first (leftmost) IP which is the original client.
		parts := strings.SplitN(xff, ",", 2)
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

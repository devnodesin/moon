// Package middleware provides HTTP middleware for authentication, authorization,
// request logging, and error handling.
package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/thalib/moon/cmd/moon/internal/constants"
)

const (
	// HeaderRateLimitLimit is the header for maximum requests allowed.
	HeaderRateLimitLimit = "X-RateLimit-Limit"
	// HeaderRateLimitRemaining is the header for remaining requests.
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	// HeaderRateLimitReset is the header for when the limit resets.
	HeaderRateLimitReset = "X-RateLimit-Reset"
)

// TokenBucket implements the token bucket algorithm for rate limiting.
type TokenBucket struct {
	tokens     int
	maxTokens  int
	refillRate int // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket.
func NewTokenBucket(maxTokens, refillRate int) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow attempts to consume a token. Returns true if allowed.
func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// Remaining returns the number of remaining tokens.
func (b *TokenBucket) Remaining() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()
	return b.tokens
}

// Reset returns the time when the bucket will be fully refilled.
func (b *TokenBucket) Reset() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.tokens >= b.maxTokens {
		return time.Now()
	}

	tokensNeeded := b.maxTokens - b.tokens
	secondsToFull := float64(tokensNeeded) / float64(b.refillRate)
	return time.Now().Add(time.Duration(secondsToFull * float64(time.Second)))
}

// Limit returns the maximum number of tokens.
func (b *TokenBucket) Limit() int {
	return b.maxTokens
}

// refill adds tokens based on time elapsed (must hold lock).
func (b *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)

	tokensToAdd := int(elapsed.Seconds() * float64(b.refillRate))
	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > b.maxTokens {
			b.tokens = b.maxTokens
		}
		b.lastRefill = now
	}
}

// RateLimiter manages rate limits for multiple entities.
type RateLimiter struct {
	buckets       sync.Map // entity ID -> *TokenBucket
	userRPM       int      // requests per minute for JWT users
	apiKeyRPM     int      // requests per minute for API keys
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// RateLimiterConfig holds rate limiter configuration.
type RateLimiterConfig struct {
	UserRPM   int // requests per minute for JWT users
	APIKeyRPM int // requests per minute for API keys
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.UserRPM <= 0 {
		config.UserRPM = 100 // default
	}
	if config.APIKeyRPM <= 0 {
		config.APIKeyRPM = 1000 // default
	}

	rl := &RateLimiter{
		userRPM:       config.UserRPM,
		apiKeyRPM:     config.APIKeyRPM,
		cleanupTicker: time.NewTicker(5 * time.Minute),
		stopCleanup:   make(chan struct{}),
	}

	// Start background cleanup
	go rl.cleanupLoop()

	return rl
}

// Stop stops the rate limiter's background cleanup.
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
	rl.cleanupTicker.Stop()
}

// cleanupLoop periodically cleans up stale buckets.
func (rl *RateLimiter) cleanupLoop() {
	for {
		select {
		case <-rl.stopCleanup:
			return
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		}
	}
}

// cleanup removes stale buckets that are fully refilled.
func (rl *RateLimiter) cleanup() {
	rl.buckets.Range(func(key, value any) bool {
		bucket := value.(*TokenBucket)
		if bucket.Remaining() >= bucket.Limit() {
			rl.buckets.Delete(key)
		}
		return true
	})
}

// getOrCreateBucket gets or creates a bucket for an entity.
func (rl *RateLimiter) getOrCreateBucket(entityID, entityType string) *TokenBucket {
	key := entityType + ":" + entityID

	// Calculate RPM based on entity type
	var rpm int
	if entityType == EntityTypeAPIKey {
		rpm = rl.apiKeyRPM
	} else {
		rpm = rl.userRPM
	}

	// Convert RPM to tokens per second
	refillRate := rpm / 60
	if refillRate < 1 {
		refillRate = 1 // minimum 1 token per second
	}

	// Try to load existing bucket
	if existing, ok := rl.buckets.Load(key); ok {
		return existing.(*TokenBucket)
	}

	// Create new bucket
	bucket := NewTokenBucket(rpm, refillRate)
	actual, _ := rl.buckets.LoadOrStore(key, bucket)
	return actual.(*TokenBucket)
}

// Allow checks if a request should be allowed for an entity.
func (rl *RateLimiter) Allow(entityID, entityType string) (allowed bool, remaining int, reset time.Time, limit int) {
	bucket := rl.getOrCreateBucket(entityID, entityType)
	allowed = bucket.Allow()
	remaining = bucket.Remaining()
	reset = bucket.Reset()
	limit = bucket.Limit()
	return
}

// RateLimitMiddleware provides rate limiting middleware.
type RateLimitMiddleware struct {
	limiter *RateLimiter
}

// NewRateLimitMiddleware creates a new rate limit middleware.
func NewRateLimitMiddleware(config RateLimiterConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: NewRateLimiter(config),
	}
}

// Stop stops the rate limiter.
func (m *RateLimitMiddleware) Stop() {
	m.limiter.Stop()
}

// RateLimit returns middleware that enforces rate limits.
func (m *RateLimitMiddleware) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entity, ok := GetAuthEntity(r.Context())
		if !ok {
			// No auth entity, skip rate limiting
			next(w, r)
			return
		}

		allowed, remaining, reset, limit := m.limiter.Allow(entity.ID, entity.Type)

		// Set rate limit headers
		w.Header().Set(HeaderRateLimitLimit, strconv.Itoa(limit))
		w.Header().Set(HeaderRateLimitRemaining, strconv.Itoa(remaining))
		w.Header().Set(HeaderRateLimitReset, strconv.FormatInt(reset.Unix(), 10))

		if !allowed {
			m.logRateLimitExceeded(r, entity.ID, entity.Type)
			m.writeRateLimitError(w, limit, reset)
			return
		}

		next(w, r)
	}
}

// logRateLimitExceeded logs rate limit violations.
func (m *RateLimitMiddleware) logRateLimitExceeded(r *http.Request, entityID, entityType string) {
	log.Printf("WARN: RATE_LIMIT_EXCEEDED entity_id=%s entity_type=%s endpoint=%s",
		entityID, entityType, r.URL.Path)
}

// writeRateLimitError writes a rate limit error response.
func (m *RateLimitMiddleware) writeRateLimitError(w http.ResponseWriter, limit int, reset time.Time) {
	// Add Retry-After header (PRD-049)
	retryAfter := int(time.Until(reset).Seconds())
	if retryAfter < 0 {
		retryAfter = 0
	}
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.Header().Set(constants.HeaderContentType, constants.MIMEApplicationJSON)
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "rate limit exceeded",
	})
}

// LoginRateLimiter manages rate limits for login attempts.
type LoginRateLimiter struct {
	attempts      sync.Map // key -> *loginAttempt
	maxAttempts   int      // max login attempts
	windowSeconds int      // window in seconds
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

type loginAttempt struct {
	count        int
	firstAttempt time.Time
	mu           sync.Mutex
}

// LoginRateLimiterConfig holds login rate limiter configuration.
type LoginRateLimiterConfig struct {
	MaxAttempts   int // max login attempts before lockout
	WindowSeconds int // lockout window in seconds
}

// NewLoginRateLimiter creates a new login rate limiter.
func NewLoginRateLimiter(config LoginRateLimiterConfig) *LoginRateLimiter {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 5 // default
	}
	if config.WindowSeconds <= 0 {
		config.WindowSeconds = 900 // 15 minutes default
	}

	lrl := &LoginRateLimiter{
		maxAttempts:   config.MaxAttempts,
		windowSeconds: config.WindowSeconds,
		cleanupTicker: time.NewTicker(5 * time.Minute),
		stopCleanup:   make(chan struct{}),
	}

	go lrl.cleanupLoop()

	return lrl
}

// Stop stops the login rate limiter's background cleanup.
func (lrl *LoginRateLimiter) Stop() {
	close(lrl.stopCleanup)
	lrl.cleanupTicker.Stop()
}

// cleanupLoop periodically cleans up expired entries.
func (lrl *LoginRateLimiter) cleanupLoop() {
	for {
		select {
		case <-lrl.stopCleanup:
			return
		case <-lrl.cleanupTicker.C:
			lrl.cleanup()
		}
	}
}

// cleanup removes expired entries.
func (lrl *LoginRateLimiter) cleanup() {
	now := time.Now()
	window := time.Duration(lrl.windowSeconds) * time.Second

	lrl.attempts.Range(func(key, value any) bool {
		attempt := value.(*loginAttempt)
		attempt.mu.Lock()
		if now.Sub(attempt.firstAttempt) > window {
			lrl.attempts.Delete(key)
		}
		attempt.mu.Unlock()
		return true
	})
}

// makeKey creates a key for IP:username combination.
func (lrl *LoginRateLimiter) makeKey(ip, username string) string {
	return ip + ":" + username
}

// CheckAndRecord checks if login attempt is allowed and records it.
// Returns true if allowed, false if rate limited.
func (lrl *LoginRateLimiter) CheckAndRecord(ip, username string) (allowed bool, remaining int, resetAt time.Time) {
	key := lrl.makeKey(ip, username)
	window := time.Duration(lrl.windowSeconds) * time.Second
	now := time.Now()

	// Get or create attempt record
	val, _ := lrl.attempts.LoadOrStore(key, &loginAttempt{
		count:        0,
		firstAttempt: now,
	})
	attempt := val.(*loginAttempt)

	attempt.mu.Lock()
	defer attempt.mu.Unlock()

	// Check if window has expired, reset if so
	if now.Sub(attempt.firstAttempt) > window {
		attempt.count = 0
		attempt.firstAttempt = now
	}

	// Calculate reset time
	resetAt = attempt.firstAttempt.Add(window)

	// Check if over limit
	if attempt.count >= lrl.maxAttempts {
		return false, 0, resetAt
	}

	// Record attempt
	attempt.count++
	remaining = lrl.maxAttempts - attempt.count

	return true, remaining, resetAt
}

// RecordFailure records a failed login attempt.
func (lrl *LoginRateLimiter) RecordFailure(ip, username string) {
	// CheckAndRecord already increments, so this is for explicit failure recording
	// when we want to record without checking first
}

// ResetForUser resets the login attempts for a user (called on successful login).
func (lrl *LoginRateLimiter) ResetForUser(ip, username string) {
	key := lrl.makeKey(ip, username)
	lrl.attempts.Delete(key)
}

// IsBlocked checks if a user is blocked from logging in.
func (lrl *LoginRateLimiter) IsBlocked(ip, username string) (blocked bool, resetAt time.Time) {
	key := lrl.makeKey(ip, username)
	window := time.Duration(lrl.windowSeconds) * time.Second
	now := time.Now()

	val, ok := lrl.attempts.Load(key)
	if !ok {
		return false, time.Time{}
	}

	attempt := val.(*loginAttempt)
	attempt.mu.Lock()
	defer attempt.mu.Unlock()

	// Check if window has expired
	if now.Sub(attempt.firstAttempt) > window {
		return false, time.Time{}
	}

	resetAt = attempt.firstAttempt.Add(window)
	blocked = attempt.count >= lrl.maxAttempts
	return
}

// WriteLoginRateLimitError writes a login rate limit error response.
func WriteLoginRateLimitError(w http.ResponseWriter, resetAt time.Time) {
	w.Header().Set(constants.HeaderContentType, constants.MIMEApplicationJSON)
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "too many login attempts",
	})
}

// LogLoginRateLimitExceeded logs login rate limit violations.
func LogLoginRateLimitExceeded(ip, username, endpoint string) {
	log.Printf("WARN: LOGIN_RATE_LIMIT ip=%s username=%s endpoint=%s",
		ip, username, endpoint)
}

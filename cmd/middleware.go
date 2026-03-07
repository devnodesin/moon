package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// corsMiddleware adds CORS headers when cors.enabled is true and handles
// OPTIONS preflight requests by returning 200 immediately.
func corsMiddleware(cfg CORSConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cfg.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		allowed := matchOrigin(origin, cfg.AllowedOrigins)
		if allowed != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowed)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// matchOrigin returns the origin if it matches the allowed list, or empty string.
func matchOrigin(origin string, allowed []string) string {
	if origin == "" {
		return ""
	}
	for _, a := range allowed {
		if a == "*" {
			return "*"
		}
		if strings.EqualFold(a, origin) {
			return origin
		}
	}
	return ""
}

// auditContextMiddleware injects a request ID and start time into audit logs.
func auditContextMiddleware(logger *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := ulid.Make().String()
		start := time.Now()

		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		logger.AuditEvent("http.request",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", w.Header().Get("X-Status-Code"),
			"duration_ms", duration.Milliseconds(),
		)
	})
}

// methodValidationMiddleware rejects methods other than GET, POST, OPTIONS with 405.
func methodValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPost, http.MethodOptions:
			next.ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "GET, POST, OPTIONS")
			WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}

// panicRecoveryMiddleware catches panics from downstream handlers, logs them,
// and returns a 500 error response.
func panicRecoveryMiddleware(logger *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered",
					"error", fmt.Sprintf("%v", rec),
					"method", r.Method,
					"path", r.URL.Path,
				)
				WriteError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware enforces per-caller rate limits for authenticated JWT and
// API key requests. It must run after the authentication middleware so that the
// caller identity is available in the request context.
func rateLimitMiddleware(rl *RateLimiter, logger *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := GetAuthIdentity(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		switch identity.CredentialType {
		case CredentialTypeJWT:
			if !rl.AllowJWT(identity.CallerID) {
				logger.AuditEvent(AuditRateLimitViolation,
					"limit_type", "jwt_traffic",
					"actor", identity.CallerID,
					"timestamp", time.Now().UTC().Format(time.RFC3339),
				)
				WriteError(w, http.StatusTooManyRequests, "Too many requests")
				return
			}
		case CredentialTypeAPIKey:
			if !rl.AllowAPIKey(identity.CallerID) {
				logger.AuditEvent(AuditRateLimitViolation,
					"limit_type", "apikey_traffic",
					"actor", identity.CallerID,
					"timestamp", time.Now().UTC().Format(time.RFC3339),
				)
				WriteError(w, http.StatusTooManyRequests, "Too many requests")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// routes and rejects names starting with "moon_".
func resourceValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resource := extractResource(r.URL.Path)
		if resource != "" && strings.HasPrefix(resource, "moon_") {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("Resource name %q is reserved", resource))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// extractResource extracts the resource name from paths like /data/{resource}:action
// or /prefix/data/{resource}:action. Returns empty string if not a data path.
func extractResource(path string) string {
	// Find "/data/" in the path
	idx := strings.Index(path, "/data/")
	if idx < 0 {
		return ""
	}
	rest := path[idx+len("/data/"):]
	// The resource name is everything before the colon
	colonIdx := strings.Index(rest, ":")
	if colonIdx < 0 {
		return rest
	}
	return rest[:colonIdx]
}

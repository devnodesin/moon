// Package middleware provides HTTP middleware for authentication, authorization,
// request logging, and error handling.
package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/thalib/moon/cmd/moon/internal/auth"
	"github.com/thalib/moon/cmd/moon/internal/constants"
)

// AuthEntity represents an authenticated entity (user or API key).
type AuthEntity struct {
	ID       string // ULID of the entity
	Type     string // "user" or "apikey"
	Role     string // "admin" or "user"
	CanWrite bool   // Write permission flag
	Username string // Username (only for users)
}

const (
	// AuthEntityContextKey is the context key for storing auth entity info.
	AuthEntityContextKey ContextKey = "auth_entity"

	// EntityTypeUser represents a user entity type.
	EntityTypeUser = "user"
	// EntityTypeAPIKey represents an API key entity type.
	EntityTypeAPIKey = "apikey"
)

// GetAuthEntity extracts auth entity from request context.
func GetAuthEntity(ctx context.Context) (*AuthEntity, bool) {
	entity, ok := ctx.Value(AuthEntityContextKey).(*AuthEntity)
	return entity, ok
}

// SetAuthEntity adds auth entity to request context.
func SetAuthEntity(ctx context.Context, entity *AuthEntity) context.Context {
	return context.WithValue(ctx, AuthEntityContextKey, entity)
}

// AuthorizationMiddleware provides authorization middleware.
type AuthorizationMiddleware struct{}

// NewAuthorizationMiddleware creates a new authorization middleware instance.
func NewAuthorizationMiddleware() *AuthorizationMiddleware {
	return &AuthorizationMiddleware{}
}

// RequireRole returns middleware that checks if the user has the required role.
// Admin role has access to all admin endpoints.
// User role has access to user-level endpoints.
func (m *AuthorizationMiddleware) RequireRole(role string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			entity, ok := GetAuthEntity(r.Context())
			if !ok {
				m.logAuthzFailure(r, "", "", "no auth entity in context")
				m.writeAuthzError(w, http.StatusForbidden, "access denied", "AUTHZ_NO_ENTITY")
				return
			}

			// Admin always has access
			if entity.Role == string(auth.RoleAdmin) {
				next(w, r)
				return
			}

			// Check if user has the required role
			if entity.Role != role && role != string(auth.RoleUser) {
				m.logAuthzFailure(r, entity.ID, entity.Type, "insufficient role")
				m.writeAuthzError(w, http.StatusForbidden, "admin access required", "AUTHZ_ROLE_REQUIRED")
				return
			}

			next(w, r)
		}
	}
}

// RequireAdmin returns middleware that requires admin role.
func (m *AuthorizationMiddleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole(string(auth.RoleAdmin))(next)
}

// RequireWrite returns middleware that checks if the user has write permission.
// Admin role always has write permission.
// User role needs can_write: true to have write permission.
func (m *AuthorizationMiddleware) RequireWrite(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entity, ok := GetAuthEntity(r.Context())
		if !ok {
			m.logAuthzFailure(r, "", "", "no auth entity in context")
			m.writeAuthzError(w, http.StatusForbidden, "access denied", "AUTHZ_NO_ENTITY")
			return
		}

		// Admin always has write access
		if entity.Role == string(auth.RoleAdmin) {
			next(w, r)
			return
		}

		// Check if user has write permission
		if !entity.CanWrite {
			m.logAuthzFailure(r, entity.ID, entity.Type, "write permission required")
			m.writeAuthzError(w, http.StatusForbidden, "write permission required", "AUTHZ_WRITE_REQUIRED")
			return
		}

		next(w, r)
	}
}

// RequireAuthenticated returns middleware that only checks if the user is authenticated.
// Any role (admin or user) is allowed.
func (m *AuthorizationMiddleware) RequireAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetAuthEntity(r.Context())
		if !ok {
			m.logAuthzFailure(r, "", "", "authentication required")
			m.writeAuthzError(w, http.StatusForbidden, "authentication required", "AUTHZ_NOT_AUTHENTICATED")
			return
		}

		next(w, r)
	}
}

// logAuthzFailure logs authorization failures.
func (m *AuthorizationMiddleware) logAuthzFailure(r *http.Request, entityID, entityType, reason string) {
	if entityID == "" {
		entityID = "unknown"
	}
	if entityType == "" {
		entityType = "unknown"
	}
	log.Printf("WARN: AUTHZ_FAILURE entity_id=%s entity_type=%s endpoint=%s reason=%s",
		entityID, entityType, r.URL.Path, reason)
}

// writeAuthzError writes an authorization error response per SPEC_API.md.
func (m *AuthorizationMiddleware) writeAuthzError(w http.ResponseWriter, statusCode int, message, _ string) {
	if statusCode == http.StatusForbidden {
		statusCode = http.StatusUnauthorized
	}
	w.Header().Set(constants.HeaderContentType, constants.MIMEApplicationJSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"message": message,
	})
}

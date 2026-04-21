package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ---------------------------------------------------------------------------
// Identity context
// ---------------------------------------------------------------------------

// AuthIdentity represents the authenticated caller attached to the request.
type AuthIdentity struct {
	CredentialType  string // "jwt" or "apikey"
	CallerID        string // user id or api key id
	Role            string // "admin" or "user"
	CanWrite        bool
	JTI             string // only for JWT credentials
	IsWebsite       bool
	AllowedOrigins  []string
	RateLimit       int
	CaptchaRequired bool
	Enabled         bool
}

type contextKey string

const authIdentityKey contextKey = "auth_identity"

// errBadRequest signals a 400 Bad Request condition in authorization checks.
var errBadRequest = errors.New("bad_request")

// SetAuthIdentity stores the identity in the request context.
func SetAuthIdentity(ctx context.Context, id *AuthIdentity) context.Context {
	return context.WithValue(ctx, authIdentityKey, id)
}

// GetAuthIdentity retrieves the identity from the request context.
func GetAuthIdentity(ctx context.Context) (*AuthIdentity, bool) {
	id, ok := ctx.Value(authIdentityKey).(*AuthIdentity)
	return id, ok
}

// ---------------------------------------------------------------------------
// JTI Revocation Store
// ---------------------------------------------------------------------------

// JTIRevocationStore is an in-memory, concurrency-safe store for revoked JTIs.
type JTIRevocationStore struct {
	mu    sync.RWMutex
	store map[string]time.Time
}

// NewJTIRevocationStore creates a new empty revocation store.
func NewJTIRevocationStore() *JTIRevocationStore {
	return &JTIRevocationStore{
		store: make(map[string]time.Time),
	}
}

// Revoke marks a JTI as revoked immediately.
func (s *JTIRevocationStore) Revoke(jti string) {
	s.mu.Lock()
	s.store[jti] = time.Now().UTC()
	s.mu.Unlock()
}

// IsRevoked returns true if the JTI has been revoked.
func (s *JTIRevocationStore) IsRevoked(jti string) bool {
	s.mu.RLock()
	_, ok := s.store[jti]
	s.mu.RUnlock()
	return ok
}

// ---------------------------------------------------------------------------
// Authentication Middleware
// ---------------------------------------------------------------------------

// AuthMiddleware extracts and validates bearer credentials.
type AuthMiddleware struct {
	db        DatabaseAdapter
	jwtSecret string
	jtiStore  *JTIRevocationStore
	prefix    string
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(db DatabaseAdapter, jwtSecret, prefix string, jtiStore *JTIRevocationStore) *AuthMiddleware {
	return &AuthMiddleware{
		db:        db,
		jwtSecret: jwtSecret,
		jtiStore:  jtiStore,
		prefix:    strings.TrimRight(prefix, "/"),
	}
}

// Authenticate wraps the next handler with bearer credential validation.
// Public routes (/, /health, POST /auth:session) bypass authentication.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.isPublicRoute(r) {
			next.ServeHTTP(w, r)
			return
		}

		token, ok := extractBearerToken(r)
		if !ok {
			WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		identity, err := m.validateCredential(r.Context(), token)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Reject API keys on auth:me endpoints
		if identity.CredentialType == CredentialTypeAPIKey {
			path := r.URL.Path
			authMePath := m.prefix + "/auth:me"
			if path == authMePath {
				WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
		}

		ctx := SetAuthIdentity(r.Context(), identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// isPublicRoute returns true for routes that don't require authentication.
func (m *AuthMiddleware) isPublicRoute(r *http.Request) bool {
	path := r.URL.Path
	method := r.Method

	// OPTIONS requests are handled by CORS middleware before auth
	if method == http.MethodOptions {
		return true
	}

	if m.prefix == "" {
		if method == http.MethodGet && (path == "/" || path == "/health") {
			return true
		}
		if method == http.MethodPost && path == "/auth:session" {
			return true
		}
		return false
	}

	if method == http.MethodGet && (path == m.prefix || path == m.prefix+"/" || path == m.prefix+"/health") {
		return true
	}
	if method == http.MethodPost && path == m.prefix+"/auth:session" {
		return true
	}
	return false
}

// extractBearerToken extracts the token from the Authorization header.
func extractBearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}

// detectCredentialType determines if a token is a JWT or API key.
func detectCredentialType(token string) string {
	if strings.HasPrefix(token, APIKeyPrefix) {
		return CredentialTypeAPIKey
	}
	parts := strings.Split(token, ".")
	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return CredentialTypeJWT
	}
	return ""
}

// validateCredential dispatches to JWT or API key validation.
func (m *AuthMiddleware) validateCredential(ctx context.Context, token string) (*AuthIdentity, error) {
	credType := detectCredentialType(token)
	switch credType {
	case CredentialTypeJWT:
		return m.validateJWT(ctx, token)
	case CredentialTypeAPIKey:
		return m.validateAPIKey(ctx, token)
	default:
		return nil, fmt.Errorf("unknown credential type")
	}
}

// validateJWT parses and verifies a JWT token.
func (m *AuthMiddleware) validateJWT(ctx context.Context, tokenStr string) (*AuthIdentity, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(m.jwtSecret), nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("jwt parse: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid jwt claims")
	}

	sub, _ := claims["sub"].(string)
	jti, _ := claims["jti"].(string)
	role, _ := claims["role"].(string)
	canWrite := toBool(claims["can_write"])

	if sub == "" || jti == "" {
		return nil, fmt.Errorf("missing required jwt claims")
	}

	if m.jtiStore.IsRevoked(jti) {
		return nil, fmt.Errorf("token revoked")
	}

	// Look up user by sub
	rows, _, err := m.db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: sub}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &AuthIdentity{
		CredentialType: CredentialTypeJWT,
		CallerID:       sub,
		Role:           role,
		CanWrite:       canWrite,
		JTI:            jti,
	}, nil
}

// validateAPIKey verifies an API key credential.
func (m *AuthMiddleware) validateAPIKey(ctx context.Context, key string) (*AuthIdentity, error) {
	if !strings.HasPrefix(key, APIKeyPrefix) || len(key) != APIKeyTotalLen {
		return nil, fmt.Errorf("invalid api key format")
	}

	hash := sha256.Sum256([]byte(key))
	keyHash := fmt.Sprintf("%x", hash)

	rows, _, err := m.db.QueryRows(ctx, "apikeys", QueryOptions{
		Filters: []Filter{{Field: "key_hash", Op: "eq", Value: keyHash}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("api key not found")
	}

	row := rows[0]
	id, _ := row["id"].(string)
	role, _ := row["role"].(string)
	canWrite := toBool(row["can_write"])
	isWebsite := toBool(row["is_website"])
	allowedOrigins, err := parseAllowedOrigins(row["allowed_origins"])
	if err != nil {
		return nil, fmt.Errorf("parse allowed origins: %w", err)
	}
	rateLimit, err := parseAPIKeyRateLimit(row["rate_limit"])
	if err != nil {
		return nil, fmt.Errorf("parse rate limit: %w", err)
	}
	enabled := true
	if rawEnabled, ok := row["enabled"]; ok {
		enabled = toBool(rawEnabled)
	}
	if !enabled {
		return nil, fmt.Errorf("api key disabled")
	}
	captchaRequired := false
	if rawCaptcha, ok := row["captcha_required"]; ok {
		captchaRequired = toBool(rawCaptcha)
	}

	// Best-effort update of last_used_at
	now := time.Now().UTC().Format(time.RFC3339)
	_ = m.db.UpdateRow(ctx, "apikeys", id, map[string]any{
		"last_used_at": now,
	})

	return &AuthIdentity{
		CredentialType:  CredentialTypeAPIKey,
		CallerID:        id,
		Role:            role,
		CanWrite:        canWrite,
		IsWebsite:       isWebsite,
		AllowedOrigins:  allowedOrigins,
		RateLimit:       rateLimit,
		CaptchaRequired: captchaRequired,
		Enabled:         enabled,
	}, nil
}

func parseAllowedOrigins(value any) ([]string, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, nil
		}
		var parsed []string
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil, err
		}
		return parsed, nil
	case []byte:
		if len(v) == 0 {
			return nil, nil
		}
		var parsed []string
		if err := json.Unmarshal(v, &parsed); err != nil {
			return nil, err
		}
		return parsed, nil
	case []string:
		result := make([]string, len(v))
		copy(result, v)
		return result, nil
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("allowed origins must contain only strings")
			}
			result = append(result, s)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported allowed origins type %T", value)
	}
}

func parseAPIKeyRateLimit(value any) (int, error) {
	if value == nil {
		return DefaultAPIKeyRateLimit, nil
	}
	var limit int
	switch v := value.(type) {
	case int:
		limit = v
	case int64:
		limit = int(v)
	case float64:
		limit = int(v)
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		limit = parsed
	case []byte:
		parsed, err := strconv.Atoi(string(v))
		if err != nil {
			return 0, err
		}
		limit = parsed
	default:
		return 0, fmt.Errorf("unsupported rate limit type %T", value)
	}
	if limit < 1 {
		return 0, fmt.Errorf("rate limit must be positive")
	}
	return limit, nil
}

// ---------------------------------------------------------------------------
// Authorization Middleware
// ---------------------------------------------------------------------------

// Authorize enforces role-based access control after authentication.
func Authorize(prefix string, next http.Handler) http.Handler {
	p := strings.TrimRight(prefix, "/")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := GetAuthIdentity(r.Context())
		if !ok {
			// No identity = public route or unauthenticated; let upstream handle it.
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path

		// Determine what kind of operation this is
		if isAdminOnlyRoute(path, r.Method, p) {
			if identity.Role != "admin" {
				WriteError(w, http.StatusForbidden, "Forbidden")
				return
			}
		}

		if isCollectionMutateRoute(path, r.Method, p) {
			if err := authorizeCollectionMutate(identity); err != nil {
				if errors.Is(err, errBadRequest) {
					WriteError(w, http.StatusBadRequest, "Resource name is reserved")
					return
				}
				WriteError(w, http.StatusForbidden, "Forbidden")
				return
			}
		}

		if isWriteRoute(path, r.Method, p) {
			if identity.Role != "admin" && !identity.CanWrite {
				WriteError(w, http.StatusForbidden, "Forbidden")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// isAdminOnlyRoute returns true for routes that require admin role.
func isAdminOnlyRoute(path, method, prefix string) bool {
	// Manage users/apikeys via data endpoints
	dataPrefix := prefix + "/data/"
	if strings.HasPrefix(path, dataPrefix) {
		rest := path[len(dataPrefix):]
		colonIdx := strings.LastIndex(rest, ":")
		if colonIdx > 0 {
			resource := rest[:colonIdx]
			action := rest[colonIdx+1:]
			if (resource == "users" || resource == "apikeys") && action == "mutate" && method == http.MethodPost {
				return true
			}
		}
	}

	return false
}

// isCollectionMutateRoute returns true for POST /collections:mutate.
func isCollectionMutateRoute(path, method, prefix string) bool {
	return method == http.MethodPost && path == prefix+"/collections:mutate"
}

// authorizeCollectionMutate checks collection mutation authorization.
func authorizeCollectionMutate(identity *AuthIdentity) error {
	if identity.Role != "admin" {
		return fmt.Errorf("forbidden")
	}
	return nil
}

// isWriteRoute returns true for routes that perform create/update/destroy
// on records (POST /data/{resource}:mutate).
func isWriteRoute(path, method, prefix string) bool {
	if method != http.MethodPost {
		return false
	}
	dataPrefix := prefix + "/data/"
	if strings.HasPrefix(path, dataPrefix) {
		rest := path[len(dataPrefix):]
		colonIdx := strings.LastIndex(rest, ":")
		if colonIdx > 0 {
			action := rest[colonIdx+1:]
			return action == "mutate"
		}
	}
	return false
}

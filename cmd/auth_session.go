package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthSessionHandler implements POST /auth:session with login, refresh, and logout operations.
type AuthSessionHandler struct {
	db           DatabaseAdapter
	cfg          *AppConfig
	logger       *Logger
	rateLimiter  *RateLimiter
	emailSender  otpSender
	otpExpiryMin int
}

// otpSender is the interface used to deliver OTP codes. It is an interface
// (rather than the concrete *EmailSender) so tests can inject a fake sender
// without making real SMTP connections.
type otpSender interface {
	SendOTPEmail(to, code string) error
}

type authSessionRequest struct {
	Op   string         `json:"op"`
	Data map[string]any `json:"data"`
}

type sessionUser struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	CanWrite    bool    `json:"can_write"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	LastLoginAt *string `json:"last_login_at"`
}

type sessionPayload struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    string      `json:"expires_at"`
	TokenType    string      `json:"token_type"`
	User         sessionUser `json:"user"`
}

// HandleSession dispatches to the appropriate operation based on the "op" field.
func (h *AuthSessionHandler) HandleSession(w http.ResponseWriter, r *http.Request) {
	var req authSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Op == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: op")
		return
	}

	if req.Data == nil {
		WriteError(w, http.StatusBadRequest, "Missing required field: data")
		return
	}

	switch req.Op {
	case "smart_login":
		h.handleSmartLogin(w, r, req.Data)
	case "refresh":
		h.handleRefresh(w, r, req.Data)
	case "logout":
		h.handleLogout(w, r, req.Data)
	default:
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Unknown op: %s", req.Op))
	}
}

func (h *AuthSessionHandler) handleSmartLogin(w http.ResponseWriter, r *http.Request, data map[string]any) {
	emailRaw, ok := data["email"]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.email")
		return
	}
	email, ok := emailRaw.(string)
	if !ok || email == "" || !isValidEmail(email) {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.email")
		return
	}
	email = strings.ToLower(email)

	passwordRaw, hasPassword := data["password"]
	codeRaw, hasCode := data["code"]
	password, _ := passwordRaw.(string)
	code, _ := codeRaw.(string)

	authMethod := "password"
	if hasCode {
		authMethod = "otp"
	}

	if hasPassword && password == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.password")
		return
	}
	if hasCode && code == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.code")
		return
	}
	if hasPassword && hasCode {
		WriteError(w, http.StatusBadRequest, "Provide either password or code, not both")
		return
	}

	ctx := r.Context()
	rateLimitKey := email + ":" + clientIP(r)

	// Step 2: verify credentials.
	if hasPassword || hasCode {
		if h.rateLimiter != nil && h.rateLimiter.SmartLoginExceeded(rateLimitKey) {
			if h.logger != nil {
				h.logger.AuditEvent(AuditRateLimitViolation,
					"limit_type", "smart_login",
					"actor", RedactEmail(email),
					"timestamp", time.Now().UTC().Format(time.RFC3339),
				)
			}
			WriteError(w, http.StatusTooManyRequests, "Too many requests")
			return
		}
		// Record the attempt now so the counter reflects this request.
		if h.rateLimiter != nil {
			h.rateLimiter.RecordSmartLogin(rateLimitKey)
		}

		rows, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
			Filters: []Filter{{Field: "email", Op: "eq", Value: email}},
			Page:    1,
			PerPage: 1,
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if len(rows) == 0 {
			// Constant-time dummy comparison to prevent timing-based enumeration.
			dummyHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
			_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
			if h.rateLimiter != nil {
				// Already recorded above; no additional action needed.
			}
			if h.logger != nil {
				h.logger.AuditEvent(AuditAuthFailure,
					"actor", RedactEmail(email),
					"method", authMethod,
					"outcome", "failure",
					"reason", "user_not_found",
				)
			}
			WriteError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		user := rows[0]
		loginMethod, _ := user["login_method"].(string)
		if loginMethod == "" {
			loginMethod = "password"
		}

		userID, _ := user["id"].(string)

		if hasPassword {
			if loginMethod == "otp" {
				if h.logger != nil {
					h.logger.AuditEvent(AuditAuthFailure,
						"actor", RedactEmail(email),
						"method", authMethod,
						"outcome", "failure",
						"reason", "method_mismatch",
					)
				}
				WriteError(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}
			storedHash, _ := user["password_hash"].(string)
			if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
				if h.logger != nil {
					h.logger.AuditEvent(AuditAuthFailure,
						"actor", RedactEmail(email),
						"method", authMethod,
						"outcome", "failure",
						"reason", "invalid_password",
					)
				}
				WriteError(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}
		} else {
			if loginMethod != "otp" {
				if h.logger != nil {
					h.logger.AuditEvent(AuditAuthFailure,
						"actor", RedactEmail(email),
						"method", authMethod,
						"outcome", "failure",
						"reason", "method_mismatch",
					)
				}
				WriteError(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}
			ok2, err := VerifyOTP(ctx, h.db, userID, code, "login")
			if err != nil {
				if h.logger != nil {
					h.logger.AuditEvent(AuditAuthFailure,
						"actor", RedactEmail(email),
						"method", authMethod,
						"outcome", "error",
						"reason", "verify_error",
					)
				}
				WriteError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
			if !ok2 {
				if h.logger != nil {
					h.logger.AuditEvent(AuditAuthFailure,
						"actor", RedactEmail(email),
						"method", authMethod,
						"outcome", "failure",
						"reason", "invalid_otp",
					)
				}
				WriteError(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}
		}

		if h.rateLimiter != nil {
			h.rateLimiter.ResetSmartLogin(rateLimitKey)
			h.rateLimiter.ResetSmartLoginInitiate(rateLimitKey)
		}

		role, _ := user["role"].(string)
		canWrite := toBool(user["can_write"])

		payload, err := h.issueSession(ctx, userID, role, canWrite, user)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		_ = h.db.UpdateRow(ctx, "users", userID, map[string]any{
			"last_login_at": now,
			"updated_at":    now,
		})

		payload.User.LastLoginAt = &now

		if h.logger != nil {
			h.logger.AuditEvent(AuditAuthSuccess,
				"actor", RedactEmail(email),
				"login_method", authMethod,
				"outcome", "success",
			)
		}

		WriteSuccess(w, http.StatusOK, "Login successful", []any{payload})
		return
	}

	// Step 1: initiate. Record rate limit hit for ALL emails to prevent
	// asymmetry between existing and non-existing users.
	if h.rateLimiter != nil && h.rateLimiter.SmartLoginInitiateExceeded(rateLimitKey) {
		if h.logger != nil {
			h.logger.AuditEvent(AuditRateLimitViolation,
				"limit_type", "smart_login_initiate",
				"actor", email,
				"timestamp", time.Now().UTC().Format(time.RFC3339),
			)
		}
		WriteError(w, http.StatusTooManyRequests, "Too many requests")
		return
	}

	// Optionally send OTP if explicitly requested.
	requestOTP, _ := data["request_otp"].(bool)
	if requestOTP && h.emailSender == nil {
		WriteError(w, http.StatusServiceUnavailable, "OTP login not configured")
		return
	}
	// Record the rate limit hit only after passing the service availability check.
	if h.rateLimiter != nil {
		h.rateLimiter.RecordSmartLoginInitiate(rateLimitKey)
	}

	// Look up the user silently to determine the login method and, when
	// requested, to send an OTP email for OTP accounts.
	loginMethod := "password"
	userID := ""
	rows, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "email", Op: "eq", Value: email}},
		Page:    1,
		PerPage: 1,
	})
	if err == nil && len(rows) > 0 {
		user := rows[0]
		lm, _ := user["login_method"].(string)
		if lm == "" {
			lm = "password"
		}
		loginMethod = lm
		userID, _ = user["id"].(string)
	}

	if requestOTP && h.emailSender != nil && loginMethod == "otp" {
		code2, hash, err := GenerateOTP(OTPCodeLength)
		if err != nil {
			if h.logger != nil {
				h.logger.Error("generate OTP failed", "error", err)
			}
		} else {
			expiry := h.otpExpiryMin
			if expiry <= 0 {
				expiry = OTPTTLMinutes
			}
			if err := StoreOTP(ctx, h.db, userID, hash, "login", expiry); err != nil {
				if h.logger != nil {
					h.logger.Error("store OTP failed", "error", err)
				}
			} else {
				if h.logger != nil {
					h.logger.Info("otp email delivery attempt", "recipient", RedactEmail(email))
				}
				if err := h.emailSender.SendOTPEmail(email, code2); err != nil {
					if h.logger != nil {
						h.logger.Error("send OTP email failed", "recipient", RedactEmail(email), "error", err)
					}
				} else {
					if h.logger != nil {
						h.logger.Info("otp email delivered", "recipient", RedactEmail(email), "outcome", "success")
					}
				}
			}
		}
	}

	// Accepted exception (documented in SPEC.md "Email / OTP Login"): this
	// response deliberately reveals the account's login_method and therefore
	// whether a queried email is registered as an OTP account. This is an
	// accepted trade-off so the UI can be fully backend-driven (the client
	// renders only the credential kind the server dictates) and does not
	// implement any manual "password vs OTP" selector.
	msg := "Enter your password"
	if loginMethod == "otp" {
		msg = "Enter your OTP"
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"challenge":    "credentials",
		"message":      msg,
		"login_method": loginMethod,
	})
}

func (h *AuthSessionHandler) handleRefresh(w http.ResponseWriter, r *http.Request, data map[string]any) {
	tokenRaw, ok := data["refresh_token"]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.refresh_token")
		return
	}
	tokenStr, ok := tokenRaw.(string)
	if !ok || tokenStr == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.refresh_token")
		return
	}

	ctx := r.Context()
	tokenHash := HashRefreshToken(tokenStr)

	tokenRows, _, err := h.db.QueryRows(ctx, "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "refresh_token_hash", Op: "eq", Value: tokenHash}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(tokenRows) == 0 {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	tokenRow := tokenRows[0]

	if tokenRow["revoked_at"] != nil {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	expiresAtStr, _ := tokenRow["expires_at"].(string)
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil || time.Now().UTC().After(expiresAt) {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	tokenID, _ := tokenRow["id"].(string)
	now := time.Now().UTC().Format(time.RFC3339)
	_ = h.db.UpdateRow(ctx, "moon_auth_refresh_tokens", tokenID, map[string]any{
		"revoked_at":        now,
		"revocation_reason": "rotated",
		"last_used_at":      now,
	})

	userID, _ := tokenRow["user_id"].(string)
	userRows, _, err := h.db.QueryRows(ctx, "users", QueryOptions{
		Filters: []Filter{{Field: "id", Op: "eq", Value: userID}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil || len(userRows) == 0 {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	user := userRows[0]
	role, _ := user["role"].(string)
	canWrite := toBool(user["can_write"])

	payload, err := h.issueSession(ctx, userID, role, canWrite, user)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if h.logger != nil {
		actor := "anonymous"
		if e, ok := user["email"].(string); ok && e != "" {
			actor = RedactEmail(e)
		}
		h.logger.AuditEvent(AuditTokenRefresh,
			"actor", actor,
			"outcome", "success",
		)
	}

	WriteSuccess(w, http.StatusOK, "Token refreshed successfully", []any{payload})
}

func (h *AuthSessionHandler) handleLogout(w http.ResponseWriter, r *http.Request, data map[string]any) {
	tokenRaw, ok := data["refresh_token"]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.refresh_token")
		return
	}
	tokenStr, ok := tokenRaw.(string)
	if !ok || tokenStr == "" {
		WriteError(w, http.StatusBadRequest, "Missing required field: data.refresh_token")
		return
	}

	ctx := r.Context()
	tokenHash := HashRefreshToken(tokenStr)

	tokenRows, _, err := h.db.QueryRows(ctx, "moon_auth_refresh_tokens", QueryOptions{
		Filters: []Filter{{Field: "refresh_token_hash", Op: "eq", Value: tokenHash}},
		Page:    1,
		PerPage: 1,
	})
	actor := "anonymous"
	if err == nil && len(tokenRows) > 0 {
		tokenRow := tokenRows[0]
		userID, _ := tokenRow["user_id"].(string)
		if userID != "" {
			userRows, _, _ := h.db.QueryRows(ctx, "users", QueryOptions{
				Filters: []Filter{{Field: "id", Op: "eq", Value: userID}},
				Page:    1,
				PerPage: 1,
			})
			if len(userRows) > 0 {
				if e, ok := userRows[0]["email"].(string); ok && e != "" {
					actor = RedactEmail(e)
				}
			}
		}
		if tokenRow["revoked_at"] == nil {
			tokenID, _ := tokenRow["id"].(string)
			now := time.Now().UTC().Format(time.RFC3339)
			_ = h.db.UpdateRow(ctx, "moon_auth_refresh_tokens", tokenID, map[string]any{
				"revoked_at":        now,
				"revocation_reason": "logout",
			})
		}
	}

	if h.logger != nil {
		h.logger.AuditEvent(AuditLogout,
			"actor", actor,
			"outcome", "success",
		)
	}

	WriteMessage(w, http.StatusOK, "Logged out successfully")
}

// issueSession creates a new JWT + refresh token pair and stores the refresh token.
func (h *AuthSessionHandler) issueSession(ctx context.Context, userID, role string, canWrite bool, user map[string]any) (*sessionPayload, error) {
	jti := GenerateULID()

	accessToken, expiresAt, err := CreateAccessToken(userID, jti, role, canWrite, h.cfg.JWTSecret, h.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("issue session: %w", err)
	}

	rawRefresh, refreshHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("issue session: %w", err)
	}

	now := time.Now().UTC()
	refreshExpiry := now.Add(time.Duration(h.cfg.JWTRefreshExpiry) * time.Second)

	err = h.db.InsertRow(ctx, "moon_auth_refresh_tokens", map[string]any{
		"id":                 GenerateULID(),
		"user_id":            userID,
		"refresh_token_hash": refreshHash,
		"expires_at":         refreshExpiry.Format(time.RFC3339),
		"created_at":         now.Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("issue session: store refresh token: %w", err)
	}

	var lastLogin *string
	if v, ok := user["last_login_at"].(string); ok && v != "" {
		lastLogin = &v
	}

	return &sessionPayload{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
		TokenType:    "Bearer",
		User: sessionUser{
			ID:          userID,
			Username:    stringVal(user, "username"),
			Email:       stringVal(user, "email"),
			Role:        role,
			CanWrite:    canWrite,
			CreatedAt:   stringVal(user, "created_at"),
			UpdatedAt:   stringVal(user, "updated_at"),
			LastLoginAt: lastLogin,
		},
	}, nil
}

func stringVal(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func toBool(v any) bool {
	switch b := v.(type) {
	case bool:
		return b
	case int64:
		return b != 0
	case float64:
		return b != 0
	case int:
		return b != 0
	case string:
		return b == "1" || b == "true"
	default:
		return false
	}
}

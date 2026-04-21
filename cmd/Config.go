package main

// ---------------------------------------------------------------------------
// Configuration key names
// ---------------------------------------------------------------------------

const (
	KeyServerHost    = "server.host"
	KeyServerPort    = "server.port"
	KeyServerPrefix  = "server.prefix"
	KeyServerLogpath = "server.logpath"

	KeyDatabaseConnection         = "database.connection"
	KeyDatabaseDatabase           = "database.database"
	KeyDatabaseUser               = "database.user"
	KeyDatabasePassword           = "database.password"
	KeyDatabaseHost               = "database.host"
	KeyDatabaseQueryTimeout       = "database.query_timeout"
	KeyDatabaseSlowQueryThreshold = "database.slow_query_threshold"

	KeyJWTSecret        = "jwt_secret"
	KeyJWTAccessExpiry  = "jwt_access_expiry"
	KeyJWTRefreshExpiry = "jwt_refresh_expiry"

	KeyBootstrapAdminUsername = "bootstrap_admin_username"
	KeyBootstrapAdminEmail    = "bootstrap_admin_email"
	KeyBootstrapAdminPassword = "bootstrap_admin_password"

	KeyCORSEnabled        = "cors.enabled"
	KeyCORSAllowedOrigins = "cors.allowed_origins"
)

// ---------------------------------------------------------------------------
// Database backend identifiers
// ---------------------------------------------------------------------------

const (
	DBConnectionSQLite   = "sqlite"
	DBConnectionPostgres = "postgres"
	DBConnectionMySQL    = "mysql"
)

// ---------------------------------------------------------------------------
// Built-in default values
// ---------------------------------------------------------------------------

const (
	DefaultServerHost    = "0.0.0.0"
	DefaultServerPort    = 6006
	DefaultServerPrefix  = ""
	DefaultServerLogpath = "/var/log/moon.log"

	DefaultDatabaseConnection         = "sqlite"
	DefaultDatabaseDatabase           = "/opt/moon/sqlite.db"
	DefaultDatabaseQueryTimeout       = 30
	DefaultDatabaseSlowQueryThreshold = 500

	DefaultJWTAccessExpiry  = 3600
	DefaultJWTRefreshExpiry = 604800

	DefaultCORSEnabled = true
)

// DefaultCORSAllowedOrigins is the default list of allowed CORS origins.
var DefaultCORSAllowedOrigins = []string{"*"}

// ---------------------------------------------------------------------------
// Default file paths
// ---------------------------------------------------------------------------

const (
	DefaultConfigPath = "/etc/moon.conf"
)

// ---------------------------------------------------------------------------
// Logging and redaction
// ---------------------------------------------------------------------------

const (
	RedactedPlaceholder = "[REDACTED]"
)

// SensitiveKeys lists configuration and header keys whose values must never
// appear in log output. All comparisons are case-insensitive.
var SensitiveKeys = []string{
	"password",
	"authorization",
	"jwt_secret",
	"refresh_token",
	"api_key",
	"token",
}

// ---------------------------------------------------------------------------
// Audit event names
// ---------------------------------------------------------------------------

const (
	AuditStartupSuccess      = "startup.success"
	AuditStartupFailure      = "startup.failure"
	AuditConfigValidation    = "config.validation_failure"
	AuditAuthSuccess         = "auth.success"
	AuditAuthFailure         = "auth.failure"
	AuditLogout              = "auth.logout"
	AuditTokenRefresh        = "auth.token_refresh"
	AuditRateLimitViolation  = "rate_limit.violation"
	AuditSchemaMutation      = "schema.mutation"
	AuditPrivilegedMutation  = "privileged.mutation"
	AuditAPIKeyCreate        = "api_key.create"
	AuditAPIKeyRotation      = "api_key.rotation"
	AuditAdminUserManagement = "admin.user_management"
	AuditShutdown            = "shutdown"
)

// ---------------------------------------------------------------------------
// Version
// ---------------------------------------------------------------------------

const (
	MoonVersion = "1.00"
)

// ---------------------------------------------------------------------------
// Fixed limits
// ---------------------------------------------------------------------------

const (
	MaxPerPage             = 200
	DefaultPerPage         = 15
	BcryptCost             = 12
	MinJWTSecretLength     = 32
	MinPasswordLength      = 8
	DefaultAPIKeyRateLimit = 15
)

// ---------------------------------------------------------------------------
// API key constants
// ---------------------------------------------------------------------------

const (
	APIKeyPrefix   = "moon_live_"
	APIKeyTotalLen = 74
)

// ---------------------------------------------------------------------------
// Credential type identifiers
// ---------------------------------------------------------------------------

const (
	CredentialTypeJWT    = "jwt"
	CredentialTypeAPIKey = "apikey"
)

// ---------------------------------------------------------------------------
// Rate limiting constants
// ---------------------------------------------------------------------------

const (
	RateLoginFailureLimit   = 5
	RateLoginFailureWindow  = 900 // 15 minutes in seconds
	RateJWTRequestLimit     = 100
	RateJWTRequestWindow    = 60 // 1 minute
	RateAPIKeyRequestLimit  = DefaultAPIKeyRateLimit
	RateAPIKeyRequestWindow = 60 // 1 minute
)

// ---------------------------------------------------------------------------
// CAPTCHA constants
// ---------------------------------------------------------------------------

const (
	CaptchaCodeLength   = 6
	CaptchaTTLSeconds   = 300
	CaptchaImageWidth   = 240
	CaptchaImageHeight  = 80
	MaxCaptchaBodyBytes = 1 << 20
)

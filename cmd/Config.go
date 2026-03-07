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
// Fixed limits
// ---------------------------------------------------------------------------

const (
	MaxPerPage         = 200
	DefaultPerPage     = 15
	BcryptCost         = 12
	MinJWTSecretLength = 32
	MinPasswordLength  = 8
)

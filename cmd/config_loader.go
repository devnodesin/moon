package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// YAML-mapped structures (unexported — used only during parsing)
// ---------------------------------------------------------------------------

type rawServerConfig struct {
	Host    *string `yaml:"host"`
	Port    *int    `yaml:"port"`
	Prefix  *string `yaml:"prefix"`
	Logpath *string `yaml:"logpath"`
}

type rawDatabaseConfig struct {
	Connection         *string `yaml:"connection"`
	Database           *string `yaml:"database"`
	User               *string `yaml:"user"`
	Password           *string `yaml:"password"`
	Host               *string `yaml:"host"`
	QueryTimeout       *int    `yaml:"query_timeout"`
	SlowQueryThreshold *int    `yaml:"slow_query_threshold"`
}

type rawCORSConfig struct {
	Enabled        *bool    `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type rawConfig struct {
	Server   *rawServerConfig   `yaml:"server"`
	Database *rawDatabaseConfig `yaml:"database"`

	JWTSecret        *string `yaml:"jwt_secret"`
	JWTAccessExpiry  *int    `yaml:"jwt_access_expiry"`
	JWTRefreshExpiry *int    `yaml:"jwt_refresh_expiry"`

	BootstrapAdminUsername *string `yaml:"bootstrap_admin_username"`
	BootstrapAdminEmail    *string `yaml:"bootstrap_admin_email"`
	BootstrapAdminPassword *string `yaml:"bootstrap_admin_password"`

	CORS *rawCORSConfig `yaml:"cors"`
}

// ---------------------------------------------------------------------------
// Resolved configuration (exported for use by the service)
// ---------------------------------------------------------------------------

// ServerConfig holds resolved server settings.
type ServerConfig struct {
	Host    string
	Port    int
	Prefix  string
	Logpath string
}

// DatabaseConfig holds resolved database settings.
type DatabaseConfig struct {
	Connection         string
	Database           string
	User               string
	Password           string
	Host               string
	QueryTimeout       int
	SlowQueryThreshold int
}

// CORSConfig holds resolved CORS settings.
type CORSConfig struct {
	Enabled        bool
	AllowedOrigins []string
}

// AppConfig is the fully validated application configuration.
type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig

	JWTSecret        string
	JWTAccessExpiry  int
	JWTRefreshExpiry int

	BootstrapAdminUsername string
	BootstrapAdminEmail    string
	BootstrapAdminPassword string

	CORS CORSConfig
}

// ---------------------------------------------------------------------------
// Loading & validation
// ---------------------------------------------------------------------------

// LoadConfig reads and validates the YAML configuration file at path.
func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read configuration file %q: %w", path, err)
	}

	if err := rejectUnknownKeys(data); err != nil {
		return nil, err
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("malformed YAML: %w", err)
	}

	cfg := applyDefaults(&raw)

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ---------------------------------------------------------------------------
// Unknown-key detection
// ---------------------------------------------------------------------------

// knownTopLevel lists every permitted top-level YAML key.
var knownTopLevel = map[string]bool{
	"server":                   true,
	"database":                 true,
	"jwt_secret":               true,
	"jwt_access_expiry":        true,
	"jwt_refresh_expiry":       true,
	"bootstrap_admin_username": true,
	"bootstrap_admin_email":    true,
	"bootstrap_admin_password": true,
	"cors":                     true,
}

var knownServerKeys = map[string]bool{
	"host": true, "port": true, "prefix": true, "logpath": true,
}

var knownDatabaseKeys = map[string]bool{
	"connection": true, "database": true, "user": true,
	"password": true, "host": true, "query_timeout": true,
	"slow_query_threshold": true,
}

var knownCORSKeys = map[string]bool{
	"enabled": true, "allowed_origins": true,
}

func rejectUnknownKeys(data []byte) error {
	var generic map[string]interface{}
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return fmt.Errorf("malformed YAML: %w", err)
	}

	for key, val := range generic {
		if !knownTopLevel[key] {
			return fmt.Errorf("unknown configuration key %q", key)
		}
		switch key {
		case "server":
			if err := checkSubKeys(val, knownServerKeys, "server"); err != nil {
				return err
			}
		case "database":
			if err := checkSubKeys(val, knownDatabaseKeys, "database"); err != nil {
				return err
			}
		case "cors":
			if err := checkSubKeys(val, knownCORSKeys, "cors"); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkSubKeys(val interface{}, allowed map[string]bool, parent string) error {
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	for k := range m {
		if !allowed[k] {
			return fmt.Errorf("unknown configuration key %q", parent+"."+k)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Defaults
// ---------------------------------------------------------------------------

func applyDefaults(raw *rawConfig) *AppConfig {
	cfg := &AppConfig{
		Server: ServerConfig{
			Host:    DefaultServerHost,
			Port:    DefaultServerPort,
			Prefix:  DefaultServerPrefix,
			Logpath: DefaultServerLogpath,
		},
		Database: DatabaseConfig{
			Connection:         DefaultDatabaseConnection,
			Database:           DefaultDatabaseDatabase,
			QueryTimeout:       DefaultDatabaseQueryTimeout,
			SlowQueryThreshold: DefaultDatabaseSlowQueryThreshold,
		},
		JWTAccessExpiry:  DefaultJWTAccessExpiry,
		JWTRefreshExpiry: DefaultJWTRefreshExpiry,
		CORS: CORSConfig{
			Enabled:        DefaultCORSEnabled,
			AllowedOrigins: DefaultCORSAllowedOrigins,
		},
	}

	if raw.Server != nil {
		s := raw.Server
		if s.Host != nil {
			cfg.Server.Host = *s.Host
		}
		if s.Port != nil {
			cfg.Server.Port = *s.Port
		}
		if s.Prefix != nil {
			cfg.Server.Prefix = *s.Prefix
		}
		if s.Logpath != nil {
			cfg.Server.Logpath = *s.Logpath
		}
	}

	if raw.Database != nil {
		d := raw.Database
		if d.Connection != nil {
			cfg.Database.Connection = *d.Connection
		}
		if d.Database != nil {
			cfg.Database.Database = *d.Database
		}
		if d.User != nil {
			cfg.Database.User = *d.User
		}
		if d.Password != nil {
			cfg.Database.Password = *d.Password
		}
		if d.Host != nil {
			cfg.Database.Host = *d.Host
		}
		if d.QueryTimeout != nil {
			cfg.Database.QueryTimeout = *d.QueryTimeout
		}
		if d.SlowQueryThreshold != nil {
			cfg.Database.SlowQueryThreshold = *d.SlowQueryThreshold
		}
	}

	// Clear sqlite default database when using non-sqlite backend without
	// an explicit database value, so validation can detect the missing field.
	if cfg.Database.Connection != DefaultDatabaseConnection {
		if raw.Database == nil || raw.Database.Database == nil {
			cfg.Database.Database = ""
		}
	}

	if raw.JWTSecret != nil {
		cfg.JWTSecret = *raw.JWTSecret
	}
	if raw.JWTAccessExpiry != nil {
		cfg.JWTAccessExpiry = *raw.JWTAccessExpiry
	}
	if raw.JWTRefreshExpiry != nil {
		cfg.JWTRefreshExpiry = *raw.JWTRefreshExpiry
	}

	if raw.BootstrapAdminUsername != nil {
		cfg.BootstrapAdminUsername = *raw.BootstrapAdminUsername
	}
	if raw.BootstrapAdminEmail != nil {
		cfg.BootstrapAdminEmail = *raw.BootstrapAdminEmail
	}
	if raw.BootstrapAdminPassword != nil {
		cfg.BootstrapAdminPassword = *raw.BootstrapAdminPassword
	}

	if raw.CORS != nil {
		c := raw.CORS
		if c.Enabled != nil {
			cfg.CORS.Enabled = *c.Enabled
		}
		if c.AllowedOrigins != nil {
			cfg.CORS.AllowedOrigins = c.AllowedOrigins
		}
	}

	return cfg
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

func validate(cfg *AppConfig) error {
	if err := validateServer(cfg); err != nil {
		return err
	}
	if err := validateDatabase(cfg); err != nil {
		return err
	}
	if err := validateJWT(cfg); err != nil {
		return err
	}
	if err := validateBootstrapAdmin(cfg); err != nil {
		return err
	}
	return nil
}

func validateServer(cfg *AppConfig) error {
	if cfg.Server.Host == "" {
		return fmt.Errorf("server.host must not be empty")
	}
	if net.ParseIP(cfg.Server.Host) == nil {
		if _, err := net.LookupHost(cfg.Server.Host); err != nil {
			return fmt.Errorf("server.host %q is not a valid host", cfg.Server.Host)
		}
	}

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", cfg.Server.Port)
	}

	if cfg.Server.Prefix != "" && !strings.HasPrefix(cfg.Server.Prefix, "/") {
		return fmt.Errorf("server.prefix must be empty or start with '/', got %q", cfg.Server.Prefix)
	}

	if err := validateLogpath(cfg.Server.Logpath); err != nil {
		return err
	}

	return nil
}

func validateLogpath(path string) error {
	if path == "" {
		return fmt.Errorf("server.logpath must not be empty")
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("server.logpath %q is not writable: %w", path, err)
	}
	f.Close()
	return nil
}

func validateDatabase(cfg *AppConfig) error {
	switch cfg.Database.Connection {
	case "sqlite":
		// sqlite needs no extra fields
	case "postgres", "mysql":
		if cfg.Database.Database == "" {
			return fmt.Errorf("database.database is required for %s", cfg.Database.Connection)
		}
		if cfg.Database.User == "" {
			return fmt.Errorf("database.user is required for %s", cfg.Database.Connection)
		}
		if cfg.Database.Password == "" {
			return fmt.Errorf("database.password is required for %s", cfg.Database.Connection)
		}
		if cfg.Database.Host == "" {
			return fmt.Errorf("database.host is required for %s", cfg.Database.Connection)
		}
	default:
		return fmt.Errorf("database.connection must be sqlite, postgres, or mysql, got %q", cfg.Database.Connection)
	}

	if cfg.Database.QueryTimeout <= 0 {
		return fmt.Errorf("database.query_timeout must be a positive integer, got %d", cfg.Database.QueryTimeout)
	}
	if cfg.Database.SlowQueryThreshold <= 0 {
		return fmt.Errorf("database.slow_query_threshold must be a positive integer, got %d", cfg.Database.SlowQueryThreshold)
	}

	return nil
}

func validateJWT(cfg *AppConfig) error {
	if cfg.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}
	if len(cfg.JWTSecret) < MinJWTSecretLength {
		return fmt.Errorf("jwt_secret must be at least %d characters", MinJWTSecretLength)
	}
	if cfg.JWTAccessExpiry <= 0 {
		return fmt.Errorf("jwt_access_expiry must be a positive integer")
	}
	if cfg.JWTRefreshExpiry <= 0 {
		return fmt.Errorf("jwt_refresh_expiry must be a positive integer")
	}
	if cfg.JWTRefreshExpiry <= cfg.JWTAccessExpiry {
		return fmt.Errorf("jwt_refresh_expiry (%d) must be greater than jwt_access_expiry (%d)",
			cfg.JWTRefreshExpiry, cfg.JWTAccessExpiry)
	}
	return nil
}

func validateBootstrapAdmin(cfg *AppConfig) error {
	hasUsername := cfg.BootstrapAdminUsername != ""
	hasEmail := cfg.BootstrapAdminEmail != ""
	hasPassword := cfg.BootstrapAdminPassword != ""

	count := 0
	if hasUsername {
		count++
	}
	if hasEmail {
		count++
	}
	if hasPassword {
		count++
	}

	if count == 0 {
		return nil
	}
	if count != 3 {
		return fmt.Errorf("all bootstrap admin fields (username, email, password) must be provided together")
	}

	if !isValidEmail(cfg.BootstrapAdminEmail) {
		return fmt.Errorf("bootstrap_admin_email %q is not a valid email address", cfg.BootstrapAdminEmail)
	}

	if err := validatePasswordPolicy(cfg.BootstrapAdminPassword); err != nil {
		return fmt.Errorf("bootstrap_admin_password: %w", err)
	}

	return nil
}

var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegexp.MatchString(email)
}

func validatePasswordPolicy(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("must be at least %d characters", MinPasswordLength)
	}

	var hasLower, hasUpper, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasLower {
		return fmt.Errorf("must contain at least one lowercase letter")
	}
	if !hasUpper {
		return fmt.Errorf("must contain at least one uppercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("must contain at least one digit")
	}

	return nil
}

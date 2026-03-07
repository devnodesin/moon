package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper: write a temp YAML config and return its path.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "moon.conf")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// helper: return a minimal valid config YAML snippet.
func minimalValidYAML(t *testing.T) string {
	t.Helper()
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	return `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
`
}

// ---------------------------------------------------------------------------
// Happy path
// ---------------------------------------------------------------------------

func TestLoadConfig_ValidMinimal(t *testing.T) {
	path := writeTempConfig(t, minimalValidYAML(t))
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify defaults applied
	assertEqual(t, cfg.Server.Host, DefaultServerHost)
	assertEqual(t, cfg.Server.Port, DefaultServerPort)
	assertEqual(t, cfg.Server.Prefix, DefaultServerPrefix)
	assertEqual(t, cfg.Database.Connection, DefaultDatabaseConnection)
	assertEqual(t, cfg.Database.Database, DefaultDatabaseDatabase)
	assertEqual(t, cfg.Database.QueryTimeout, DefaultDatabaseQueryTimeout)
	assertEqual(t, cfg.Database.SlowQueryThreshold, DefaultDatabaseSlowQueryThreshold)
	assertEqual(t, cfg.JWTAccessExpiry, DefaultJWTAccessExpiry)
	assertEqual(t, cfg.JWTRefreshExpiry, DefaultJWTRefreshExpiry)
	assertEqual(t, cfg.CORS.Enabled, DefaultCORSEnabled)
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "*" {
		t.Errorf("expected AllowedOrigins=[*], got %v", cfg.CORS.AllowedOrigins)
	}
}

func TestLoadConfig_ValidFull(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `
server:
  host: "127.0.0.1"
  port: 8080
  prefix: "/api"
  logpath: "` + logPath + `"
database:
  connection: sqlite
  database: "/tmp/test.db"
  query_timeout: 60
  slow_query_threshold: 1000
jwt_secret: "supersecretkeythatisatleast32characters!!"
jwt_access_expiry: 1800
jwt_refresh_expiry: 86400
bootstrap_admin_username: admin
bootstrap_admin_email: admin@example.com
bootstrap_admin_password: "Admin123"
cors:
  enabled: false
  allowed_origins:
    - "https://example.com"
`
	path := writeTempConfig(t, yaml)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, cfg.Server.Host, "127.0.0.1")
	assertEqual(t, cfg.Server.Port, 8080)
	assertEqual(t, cfg.Server.Prefix, "/api")
	assertEqual(t, cfg.Database.Connection, "sqlite")
	assertEqual(t, cfg.Database.Database, "/tmp/test.db")
	assertEqual(t, cfg.Database.QueryTimeout, 60)
	assertEqual(t, cfg.Database.SlowQueryThreshold, 1000)
	assertEqual(t, cfg.JWTAccessExpiry, 1800)
	assertEqual(t, cfg.JWTRefreshExpiry, 86400)
	assertEqual(t, cfg.BootstrapAdminUsername, "admin")
	assertEqual(t, cfg.BootstrapAdminEmail, "admin@example.com")
	assertEqual(t, cfg.BootstrapAdminPassword, "Admin123")
	assertEqual(t, cfg.CORS.Enabled, false)
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("expected AllowedOrigins=[https://example.com], got %v", cfg.CORS.AllowedOrigins)
	}
}

// ---------------------------------------------------------------------------
// Missing config file
// ---------------------------------------------------------------------------

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/moon.conf")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "cannot read configuration file") {
		t.Errorf("error should mention cannot read: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Unknown keys
// ---------------------------------------------------------------------------

func TestLoadConfig_UnknownTopLevelKey(t *testing.T) {
	yaml := minimalValidYAML(t) + "bogus_key: 42\n"
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	if !strings.Contains(err.Error(), "bogus_key") {
		t.Errorf("error should name the unknown key: %v", err)
	}
}

func TestLoadConfig_UnknownNestedKey(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
  unknown_server_key: true
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for unknown nested key")
	}
	if !strings.Contains(err.Error(), "server.unknown_server_key") {
		t.Errorf("error should name the unknown key: %v", err)
	}
}

func TestLoadConfig_UnknownDatabaseKey(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
database:
  connection: sqlite
  extra: "nope"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for unknown database key")
	}
	if !strings.Contains(err.Error(), "database.extra") {
		t.Errorf("error should name the unknown key: %v", err)
	}
}

func TestLoadConfig_UnknownCORSKey(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
cors:
  badkey: true
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for unknown CORS key")
	}
	if !strings.Contains(err.Error(), "cors.badkey") {
		t.Errorf("error should name the unknown key: %v", err)
	}
}

// ---------------------------------------------------------------------------
// JWT validation
// ---------------------------------------------------------------------------

func TestLoadConfig_MissingJWTSecret(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for missing jwt_secret")
	}
	if !strings.Contains(err.Error(), "jwt_secret is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_JWTSecretTooShort(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "short"
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for short jwt_secret")
	}
	if !strings.Contains(err.Error(), "at least 32 characters") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_JWTRefreshNotGreaterThanAccess(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
jwt_access_expiry: 3600
jwt_refresh_expiry: 3600
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error when jwt_refresh_expiry == jwt_access_expiry")
	}
	if !strings.Contains(err.Error(), "greater than") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_JWTRefreshLessThanAccess(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
jwt_access_expiry: 7200
jwt_refresh_expiry: 3600
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error when jwt_refresh_expiry < jwt_access_expiry")
	}
	if !strings.Contains(err.Error(), "greater than") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Bootstrap admin validation
// ---------------------------------------------------------------------------

func TestLoadConfig_PartialBootstrapAdmin_UsernameOnly(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
bootstrap_admin_username: admin
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for partial bootstrap admin")
	}
	if !strings.Contains(err.Error(), "all bootstrap admin fields") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_PartialBootstrapAdmin_TwoFields(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
bootstrap_admin_username: admin
bootstrap_admin_email: admin@example.com
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for partial bootstrap admin")
	}
	if !strings.Contains(err.Error(), "all bootstrap admin fields") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_BootstrapAdminInvalidEmail(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
bootstrap_admin_username: admin
bootstrap_admin_email: not-an-email
bootstrap_admin_password: "Admin123"
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
	if !strings.Contains(err.Error(), "not a valid email") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_BootstrapAdminWeakPassword(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
bootstrap_admin_username: admin
bootstrap_admin_email: admin@example.com
bootstrap_admin_password: "short"
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for weak password")
	}
	if !strings.Contains(err.Error(), "bootstrap_admin_password") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_BootstrapAdminNoUppercase(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
bootstrap_admin_username: admin
bootstrap_admin_email: admin@example.com
bootstrap_admin_password: "alllower1"
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for password without uppercase")
	}
	if !strings.Contains(err.Error(), "uppercase") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_BootstrapAdminNoDigit(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
bootstrap_admin_username: admin
bootstrap_admin_email: admin@example.com
bootstrap_admin_password: "Abcdefgh"
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for password without digit")
	}
	if !strings.Contains(err.Error(), "digit") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Server validation
// ---------------------------------------------------------------------------

func TestLoadConfig_InvalidPort(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  port: 99999
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
	if !strings.Contains(err.Error(), "server.port") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_InvalidPrefix(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  prefix: "no-leading-slash"
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid prefix")
	}
	if !strings.Contains(err.Error(), "server.prefix") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_UnwritableLogpath(t *testing.T) {
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "/nonexistent/dir/test.log"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for unwritable logpath")
	}
	if !strings.Contains(err.Error(), "server.logpath") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Database validation
// ---------------------------------------------------------------------------

func TestLoadConfig_InvalidDatabaseConnection(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
database:
  connection: mongodb
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid database connection")
	}
	if !strings.Contains(err.Error(), "database.connection") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_PostgresMissingFields(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
database:
  connection: postgres
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for postgres without required fields")
	}
	if !strings.Contains(err.Error(), "required for postgres") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_PostgresComplete(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
database:
  connection: postgres
  database: moondb
  user: moonuser
  password: moonpass
  host: localhost
`
	path := writeTempConfig(t, yaml)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, cfg.Database.Connection, "postgres")
	assertEqual(t, cfg.Database.Database, "moondb")
	assertEqual(t, cfg.Database.User, "moonuser")
	assertEqual(t, cfg.Database.Host, "localhost")
}

func TestLoadConfig_MySQLMissingFields(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
database:
  connection: mysql
  database: moondb
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for mysql without required fields")
	}
	if !strings.Contains(err.Error(), "required for mysql") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Backend-specific keys silently ignored
// ---------------------------------------------------------------------------

func TestLoadConfig_SqliteIgnoresBackendKeys(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
database:
  connection: sqlite
  user: ignored
  password: ignored
  host: ignored
`
	path := writeTempConfig(t, yaml)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, cfg.Database.Connection, "sqlite")
	assertEqual(t, cfg.Database.User, "ignored")
}

// ---------------------------------------------------------------------------
// Malformed YAML
// ---------------------------------------------------------------------------

func TestLoadConfig_MalformedYAML(t *testing.T) {
	path := writeTempConfig(t, "{{{{not yaml at all")
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "malformed YAML") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Default values applied when keys omitted
// ---------------------------------------------------------------------------

func TestLoadConfig_DefaultsApplied(t *testing.T) {
	path := writeTempConfig(t, minimalValidYAML(t))
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, cfg.Server.Host, DefaultServerHost)
	assertEqual(t, cfg.Server.Port, DefaultServerPort)
	assertEqual(t, cfg.Server.Prefix, DefaultServerPrefix)
	assertEqual(t, cfg.Database.Connection, DefaultDatabaseConnection)
	assertEqual(t, cfg.Database.Database, DefaultDatabaseDatabase)
	assertEqual(t, cfg.Database.QueryTimeout, DefaultDatabaseQueryTimeout)
	assertEqual(t, cfg.Database.SlowQueryThreshold, DefaultDatabaseSlowQueryThreshold)
	assertEqual(t, cfg.JWTAccessExpiry, DefaultJWTAccessExpiry)
	assertEqual(t, cfg.JWTRefreshExpiry, DefaultJWTRefreshExpiry)
	assertEqual(t, cfg.CORS.Enabled, DefaultCORSEnabled)
}

// ---------------------------------------------------------------------------
// Negative database values
// ---------------------------------------------------------------------------

func TestLoadConfig_NegativeQueryTimeout(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
database:
  query_timeout: -1
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for negative query_timeout")
	}
	if !strings.Contains(err.Error(), "query_timeout") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadConfig_ZeroSlowQueryThreshold(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  logpath: "` + logPath + `"
database:
  slow_query_threshold: 0
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for zero slow_query_threshold")
	}
	if !strings.Contains(err.Error(), "slow_query_threshold") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Email validation
// ---------------------------------------------------------------------------

func TestIsValidEmail(t *testing.T) {
	valid := []string{
		"user@example.com",
		"admin@sub.domain.org",
		"test+tag@gmail.com",
	}
	invalid := []string{
		"not-an-email",
		"@missing-local.com",
		"missing-at.com",
		"",
	}

	for _, e := range valid {
		if !isValidEmail(e) {
			t.Errorf("expected %q to be valid", e)
		}
	}
	for _, e := range invalid {
		if isValidEmail(e) {
			t.Errorf("expected %q to be invalid", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Password policy
// ---------------------------------------------------------------------------

func TestValidatePasswordPolicy(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
		errMsg   string
	}{
		{"Admin123", false, ""},
		{"short", true, "at least 8"},
		{"alllower1", true, "uppercase"},
		{"ALLUPPER1", true, "lowercase"},
		{"Abcdefgh", true, "digit"},
		{"Aa1!@#$%", false, ""},
	}

	for _, tt := range tests {
		err := validatePasswordPolicy(tt.password)
		if tt.wantErr && err == nil {
			t.Errorf("password %q: expected error containing %q", tt.password, tt.errMsg)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("password %q: unexpected error: %v", tt.password, err)
		}
		if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
			t.Errorf("password %q: expected error containing %q, got %v", tt.password, tt.errMsg, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Port zero
// ---------------------------------------------------------------------------

func TestLoadConfig_PortZero(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
server:
  port: 0
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

// ---------------------------------------------------------------------------
// JWT access expiry non-positive
// ---------------------------------------------------------------------------

func TestLoadConfig_JWTAccessExpiryZero(t *testing.T) {
	logDir := t.TempDir()
	logPath := filepath.Join(logDir, "test.log")
	yaml := `jwt_secret: "this-is-a-very-long-secret-that-is-at-least-32-chars!"
jwt_access_expiry: 0
server:
  logpath: "` + logPath + `"
`
	path := writeTempConfig(t, yaml)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for zero jwt_access_expiry")
	}
}

// ---------------------------------------------------------------------------
// No bootstrap admin fields — should succeed
// ---------------------------------------------------------------------------

func TestLoadConfig_NoBootstrapAdmin(t *testing.T) {
	path := writeTempConfig(t, minimalValidYAML(t))
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, cfg.BootstrapAdminUsername, "")
	assertEqual(t, cfg.BootstrapAdminEmail, "")
	assertEqual(t, cfg.BootstrapAdminPassword, "")
}

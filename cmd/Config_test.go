package main

import (
	"testing"
)

// TestConfigConstants verifies that all named constants in Config.go are set to
// the values specified in the SPEC and PRD.
func TestConfigConstants(t *testing.T) {
	t.Run("default server values", func(t *testing.T) {
		assertEqual(t, DefaultServerHost, "0.0.0.0")
		assertEqual(t, DefaultServerPort, 6006)
		assertEqual(t, DefaultServerPrefix, "")
		assertEqual(t, DefaultServerLogpath, "/var/log/moon.log")
	})

	t.Run("default database values", func(t *testing.T) {
		assertEqual(t, DefaultDatabaseConnection, "sqlite")
		assertEqual(t, DefaultDatabaseDatabase, "/opt/moon/sqlite.db")
		assertEqual(t, DefaultDatabaseQueryTimeout, 30)
		assertEqual(t, DefaultDatabaseSlowQueryThreshold, 500)
	})

	t.Run("default JWT values", func(t *testing.T) {
		assertEqual(t, DefaultJWTAccessExpiry, 3600)
		assertEqual(t, DefaultJWTRefreshExpiry, 604800)
	})

	t.Run("default CORS values", func(t *testing.T) {
		assertEqual(t, DefaultCORSEnabled, true)
		if len(DefaultCORSAllowedOrigins) != 1 || DefaultCORSAllowedOrigins[0] != "*" {
			t.Fatalf("expected DefaultCORSAllowedOrigins=[\"*\"], got %v", DefaultCORSAllowedOrigins)
		}
	})

	t.Run("default config path", func(t *testing.T) {
		assertEqual(t, DefaultConfigPath, "/etc/moon.conf")
	})

	t.Run("fixed limits", func(t *testing.T) {
		assertEqual(t, MaxPerPage, 200)
		assertEqual(t, DefaultPerPage, 15)
		assertEqual(t, BcryptCost, 12)
		assertEqual(t, MinJWTSecretLength, 32)
		assertEqual(t, MinPasswordLength, 8)
	})
}

// TestConfigKeyConstants ensures key name constants are the expected YAML paths.
func TestConfigKeyConstants(t *testing.T) {
	keys := map[string]string{
		"KeyServerHost":                 KeyServerHost,
		"KeyServerPort":                 KeyServerPort,
		"KeyServerPrefix":               KeyServerPrefix,
		"KeyServerLogpath":              KeyServerLogpath,
		"KeyDatabaseConnection":         KeyDatabaseConnection,
		"KeyDatabaseDatabase":           KeyDatabaseDatabase,
		"KeyDatabaseUser":               KeyDatabaseUser,
		"KeyDatabasePassword":           KeyDatabasePassword,
		"KeyDatabaseHost":               KeyDatabaseHost,
		"KeyDatabaseQueryTimeout":       KeyDatabaseQueryTimeout,
		"KeyDatabaseSlowQueryThreshold": KeyDatabaseSlowQueryThreshold,
		"KeyJWTSecret":                  KeyJWTSecret,
		"KeyJWTAccessExpiry":            KeyJWTAccessExpiry,
		"KeyJWTRefreshExpiry":           KeyJWTRefreshExpiry,
		"KeyBootstrapAdminUsername":     KeyBootstrapAdminUsername,
		"KeyBootstrapAdminEmail":        KeyBootstrapAdminEmail,
		"KeyBootstrapAdminPassword":     KeyBootstrapAdminPassword,
		"KeyCORSEnabled":                KeyCORSEnabled,
		"KeyCORSAllowedOrigins":         KeyCORSAllowedOrigins,
	}

	expected := map[string]string{
		"KeyServerHost":                 "server.host",
		"KeyServerPort":                 "server.port",
		"KeyServerPrefix":               "server.prefix",
		"KeyServerLogpath":              "server.logpath",
		"KeyDatabaseConnection":         "database.connection",
		"KeyDatabaseDatabase":           "database.database",
		"KeyDatabaseUser":               "database.user",
		"KeyDatabasePassword":           "database.password",
		"KeyDatabaseHost":               "database.host",
		"KeyDatabaseQueryTimeout":       "database.query_timeout",
		"KeyDatabaseSlowQueryThreshold": "database.slow_query_threshold",
		"KeyJWTSecret":                  "jwt_secret",
		"KeyJWTAccessExpiry":            "jwt_access_expiry",
		"KeyJWTRefreshExpiry":           "jwt_refresh_expiry",
		"KeyBootstrapAdminUsername":     "bootstrap_admin_username",
		"KeyBootstrapAdminEmail":        "bootstrap_admin_email",
		"KeyBootstrapAdminPassword":     "bootstrap_admin_password",
		"KeyCORSEnabled":                "cors.enabled",
		"KeyCORSAllowedOrigins":         "cors.allowed_origins",
	}

	for name, got := range keys {
		want := expected[name]
		if got != want {
			t.Errorf("%s = %q; want %q", name, got, want)
		}
	}
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

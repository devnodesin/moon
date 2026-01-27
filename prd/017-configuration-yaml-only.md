## Overview

- Simplify and standardize configuration management to use YAML-only format
- Remove legacy .env and TOML configuration support for cleaner, maintainable codebase
- Centralize all default values in a Defaults struct to eliminate hardcoded literals
- Implement --config command-line flag for custom configuration file paths
- Align defaults with production requirements (port 6006, proper paths for logs and database)
- Updated the `SPEC.md` to match this changes
- Updated the DOCUMENTATIONS: `README.md`, `INSTALL.md` `USAGE.md` to match this changes

## Requirements

- Remove all `.env` file loading logic and godotenv dependency from `internal/config/config.go`
- Remove TOML configuration support (Viper still allows it but should be explicitly disabled)
- Support only YAML configuration format
- Default configuration file location: `/etc/moon.conf`
- Add `--config {filename}` command-line flag to override default config location in `cmd/moon/main.go`
- Create centralized `Defaults` struct in `internal/config/config.go` containing all default values
- Never use hardcoded string literals for configuration defaults in business logic
- Update `AppConfig` struct to include:
  - `Server.Port` (default: 6006)
  - `Server.Host` (default: "0.0.0.0")
  - `Logging.Path` (default: "/var/log/moon") - new field
  - `Database.Connection` (default: "sqlite") - database type
  - `Database.Database` (default: "/opt/moon/sqlite.db") - database file/name
  - `Database.User` (default: "")
  - `Database.Password` (default: "")
  - `Database.Host` (default: "0.0.0.0")
  - `JWT.Secret` (required, no default)
  - `JWT.Expiry` (default: 3600)
  - `APIKey.Enabled` (default: false)
  - `APIKey.Header` (default: "X-API-KEY")
- Create `samples/moon.conf` with minimal quick-start configuration
- Create `samples/moon-full.conf` with comprehensive documentation for all fields
- Remove `samples/.env.example` file
- Update `samples/config.example.yaml` or remove if redundant
- Default log file location: `{LOG_PATH}/main.log` = `/var/log/moon/main.log`
- Default SQLite database location: `/opt/moon/sqlite.db`
- If no database configuration provided, default to SQLite at `/opt/moon/sqlite.db`
- Remove environment variable loading (MOON_* env vars) - YAML-only approach
- Configuration validation must fail fast with clear error messages
- Update all config-related unit tests in `internal/config/config_test.go`

## Acceptance

- Application successfully loads configuration from YAML file only
- Default configuration path is `/etc/moon.conf`
- Custom config path works via `--config /path/to/config.yaml` flag
- All godotenv imports and `.env` loading code removed
- TOML support explicitly disabled in Viper configuration
- Defaults struct defined with all default values (port 6006, LOG_PATH, DB paths, etc.)
- No hardcoded default values scattered in business logic
- `samples/moon.conf` provides minimal working example for quick start
- `samples/moon-full.conf` documents all available configuration options with comments
- `samples/.env.example` file deleted
- SQLite defaults to `/opt/moon/sqlite.db` when no database configured
- Default log file is `/var/log/moon/main.log`
- Configuration validation fails with descriptive error for missing required fields (JWT secret)
- Unit tests cover YAML loading, defaults application, validation, and --config flag
- Documentation updated to reflect YAML-only configuration approach
- No references to .env or TOML in code or documentation
- Updated the `SPEC.md` to match this changes
- Updated the DOCUMENTATIONS: `README.md`, `INSTALL.md` `USAGE.md` to match this changes

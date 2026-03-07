## Overview

- Moon is a single-process, API-first Go backend. Before it can serve any traffic it must initialize itself from a known, validated configuration state. This PRD defines the Go module setup, the central `Config.go` constant source, the `-c <path>` CLI flag, configuration file loading, and full startup validation.
- All Go source files live in `cmd/`. The compiled output binary is named `moon`. There are no sub-packages.
- The goal is a deterministic, fail-fast bootstrap that either starts with a fully valid configuration or exits with a clear error message before opening any network socket.

## Requirements

### Module and Entry Point

- The Go module must be initialized with an appropriate module path.
- `cmd/main.go` is the entry point. It must parse flags, load configuration, and start the service.
- The output binary must be named `moon`.
- All Go source files reside in `cmd/`.

### `Config.go` — Central Constant Source

- A file named `cmd/Config.go` must exist and be the single source for:
  - all configuration key names (as named string constants)
  - all built-in default values (as named typed constants)
  - all default file paths (e.g. default config path, default log path, default DB path)
  - all fixed limits (e.g. max `per_page`, bcrypt cost, minimum JWT secret length)
- No magic literals for configuration keys, defaults, paths, or limits may appear outside `Config.go`.
- Named constants must use `UPPER_SNAKE_CASE` or idiomatic Go `CamelCase`; the style must be consistent throughout the file.

### CLI Flag

| Flag | Default | Description |
|------|---------|-------------|
| `-c <path>` | `/etc/moon.conf` | Path to the YAML configuration file |

- If `-c` is omitted the service must attempt to load `/etc/moon.conf`.
- If `-c` is provided the service must load the specified file instead of the default path.
- No other flag override mechanism is part of this specification.

### Configuration File Format

- The configuration file is YAML.
- The service must parse only the keys defined in the configuration reference table (SPEC.md §8.3).
- Unknown keys must cause startup failure with a descriptive error message.

### Configuration Reference

| Key | Required | Default | Constraint |
|-----|----------|---------|------------|
| `server.host` | no | `0.0.0.0` | valid host string |
| `server.port` | no | `6006` | integer 1–65535 |
| `server.prefix` | no | `""` | empty or single leading-slash path segment |
| `server.logpath` | no | `/var/log/moon.log` | writable file path |
| `database.connection` | no | `sqlite` | one of `sqlite`, `postgres`, `mysql` |
| `database.database` | no for sqlite; yes for postgres/mysql | `/opt/moon/sqlite.db` | file path or DB name |
| `database.user` | conditional | none | required when backend needs username |
| `database.password` | conditional | none | required when backend needs password |
| `database.host` | conditional | none | required for networked backends |
| `database.query_timeout` | no | `30` | positive integer (seconds) |
| `database.slow_query_threshold` | no | `500` | positive integer (milliseconds) |
| `jwt_secret` | yes | none | minimum 32 characters |
| `jwt_access_expiry` | no | `3600` | positive integer (seconds) |
| `jwt_refresh_expiry` | no | `604800` | positive integer seconds and greater than `jwt_access_expiry` |
| `bootstrap_admin_username` | conditional | none | first-run only; required if any bootstrap field is set |
| `bootstrap_admin_email` | conditional | none | first-run only; valid email; required if any bootstrap field is set |
| `bootstrap_admin_password` | conditional | none | first-run only; must satisfy password policy; required if any bootstrap field is set |
| `cors.enabled` | no | `true` | boolean |
| `cors.allowed_origins` | no | `["*"]` | list of origin strings |

### Validation Rules

- Unknown keys must fail startup.
- The resolved configuration file must exist and be readable; if not, startup fails.
- Malformed YAML or malformed values must fail startup.
- `jwt_secret` is required; absence or a value shorter than 32 characters must fail startup.
- `jwt_refresh_expiry` must be strictly greater than `jwt_access_expiry`.
- `server.logpath` must resolve to a writable file location; if the file cannot be opened or created, startup fails.
- If any bootstrap admin field is provided, all three bootstrap admin fields (`username`, `email`, `password`) must be provided; otherwise startup fails.
- Backend-specific keys that do not apply to the selected backend must be silently ignored (no cross-backend validation error).
- Optional keys must use documented defaults when omitted; defaults must be read from `Config.go` constants.

### Startup Failure Behavior

- If any validation step fails the service must log a descriptive error message and exit with a non-zero exit code before any network socket is opened.
- Error messages must not expose secret values.

### Configuration Precedence

1. Built-in defaults from `Config.go`
2. YAML file overrides from the resolved config path

No environment variable or additional runtime override mechanism is defined.

## Acceptance

- Running `moon -c /path/to/valid.conf` starts successfully when the file is valid and complete.
- Running `moon` (no `-c`) attempts to load `/etc/moon.conf`; if missing, startup fails with a clear error.
- Running `moon -c /missing.conf` fails with a descriptive error and non-zero exit code.
- A config file with an unknown key causes startup failure with the unknown key named in the error.
- A config file missing `jwt_secret` causes startup failure.
- A config file with `jwt_secret` shorter than 32 characters causes startup failure.
- A config file with `jwt_refresh_expiry` ≤ `jwt_access_expiry` causes startup failure.
- Providing only one or two bootstrap admin fields causes startup failure.
- All default values originate from named constants in `Config.go`; no magic literals exist elsewhere.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

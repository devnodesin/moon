## Overview

- On every startup, Moon must verify that its required system tables exist and, if they are missing, create them before accepting any traffic.
- The required system objects are the `users` collection, the `apikeys` collection, and the `moon_auth_refresh_tokens` internal table.
- If a bootstrap admin is configured and no admin user exists yet, the bootstrap admin must be created during this phase.
- This PRD covers the reconciliation logic, the exact DDL for each system object, and the bootstrap admin creation flow.
- All Go source files reside in `cmd/`. The compiled output binary is named `moon`.

## Requirements

### Startup Reconciliation Order

During startup, after the database adapter is initialized and connectivity is verified, the service must execute the following steps in order:

1. Check if `users` table exists; if not, create it using the DDL below.
2. Check if `apikeys` table exists; if not, create it using the DDL below.
3. Check if `moon_auth_refresh_tokens` table exists; if not, create it using the DDL below.
4. If all bootstrap admin fields are configured and no admin user exists, create the bootstrap admin user.
5. Proceed to schema registry population.

If any step fails, startup must fail with a descriptive error.

### `users` Table DDL

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    can_write BOOLEAN NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    last_login_at TEXT,
    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE INDEX idx_users_role ON users(role);
```

Field rules:

- `id`: server-generated ULID, immutable, primary key.
- `username`: 3–63 characters, lowercase snake_case, case-insensitive uniqueness enforced after normalization.
- `email`: valid email, case-insensitive uniqueness enforced after normalization to lowercase.
- `password_hash`: bcrypt hash at cost 12; never returned by any API.
- `role`: one of `admin` or `user`.
- `can_write`: boolean; default false; ignored for `admin` role.
- `created_at`, `updated_at`: RFC3339 timestamps, server-managed.
- `last_login_at`: nullable RFC3339 timestamp.

### `apikeys` Table DDL

```sql
CREATE TABLE apikeys (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    can_write BOOLEAN NOT NULL DEFAULT 0,
    key_hash TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    last_used_at TEXT,
    CONSTRAINT apikeys_name_unique UNIQUE (name),
    CONSTRAINT apikeys_key_hash_unique UNIQUE (key_hash)
);

CREATE INDEX idx_apikeys_last_used_at ON apikeys(last_used_at);
```

Field rules:

- `id`: server-generated ULID, immutable, primary key.
- `name`: 3–100 characters, unique administrative label.
- `role`: one of `admin` or `user`.
- `can_write`: boolean; default false; ignored for `admin` role.
- `key_hash`: SHA-256 or stronger one-way hash of the raw API key; never returned by any API.
- `created_at`, `updated_at`: RFC3339 timestamps, server-managed.
- `last_used_at`: nullable RFC3339 timestamp.

### `moon_auth_refresh_tokens` Table DDL

```sql
CREATE TABLE moon_auth_refresh_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    refresh_token_hash TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    last_used_at TEXT,
    revoked_at TEXT,
    revocation_reason TEXT
);

CREATE UNIQUE INDEX idx_refresh_tokens_hash ON moon_auth_refresh_tokens(refresh_token_hash);
CREATE INDEX idx_refresh_tokens_user_revoked ON moon_auth_refresh_tokens(user_id, revoked_at);
CREATE INDEX idx_refresh_tokens_expires_at ON moon_auth_refresh_tokens(expires_at);
```

Field rules:

- `id`: server-generated ULID, immutable, primary key.
- `user_id`: logical reference to `users.id`; not a foreign key constraint.
- `refresh_token_hash`: SHA-256 or stronger hash of the raw token; unique.
- `expires_at`: RFC3339 hard expiry timestamp.
- `created_at`: RFC3339 issue timestamp.
- `last_used_at`: nullable RFC3339; set on successful exchange.
- `revoked_at`: nullable RFC3339; set on invalidation.
- `revocation_reason`: nullable string; implementation-controlled audit reason.
- This table must never be exposed through collection or resource APIs.

### Idempotency

- All existence checks and table-creation steps must be idempotent (e.g. `CREATE TABLE IF NOT EXISTS`).
- Running reconciliation on an already-initialized database must succeed without modifying existing data.

### Bootstrap Admin Creation

- Bootstrap admin creation occurs only when `bootstrap_admin_username`, `bootstrap_admin_email`, and `bootstrap_admin_password` are all present in configuration AND no admin user exists in the `users` table at startup.
- The bootstrap admin must be created with `role = admin` and `can_write = true`.
- The password must be hashed with bcrypt at cost 12 before storage.
- If bootstrap admin creation fails (e.g. username conflict), startup must fail with a descriptive error.
- Bootstrap admin credentials should be removed from configuration after successful initialization; the spec does not enforce removal but the implementation should log a warning if they remain.

### System Table Visibility Protection

- `users` and `apikeys` are API-visible system collections and must appear in schema discovery results.
- `moon_auth_refresh_tokens` is an internal system table and must never appear in schema discovery or any public API response.

## Acceptance

- Starting against an empty database creates `users`, `apikeys`, and `moon_auth_refresh_tokens` tables with the correct schemas and indexes.
- Starting against a database that already has these tables does not alter them and does not fail.
- With bootstrap admin fields configured and an empty `users` table, the bootstrap admin user is created.
- With bootstrap admin fields configured but an admin already in `users`, no new user is created.
- `moon_auth_refresh_tokens` is never returned by `/collections:query` or any data endpoint.
- `users` and `apikeys` are returned by `/collections:query`.
- If the `users` table cannot be created (e.g. permission error), startup fails with a descriptive error.
- `go vet ./cmd/...` reports zero issues.

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

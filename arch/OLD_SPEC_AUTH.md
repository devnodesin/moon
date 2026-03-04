# Authentication Design

**Rules** to Follow for this document

- Be written from the user's perspective (outside-in: user → server)
- Focus on simplicity and clarity
- Define only what is needed for real use, not hypothetical features
- Use sections and bullet points unless told to be used
- Keep roles and permissions minimal and easy to manage
- **AIP-136 Custom Actions:** APIs use a colon separator (`:`) to distinguish between the resource and the action, providing a predictable and AI-friendly interface.

## Overview

Moon's authentication system provides two authentication methods:

1. **JWT-based authentication** for interactive users (web/mobile applications)
2. **API Key authentication** for machine-to-machine integrations

Both methods support role-based access control (RBAC) with two roles: `admin` and `user`.

## Access Types

### User Login (JWT-based)

**Purpose:** For interactive users accessing the system via web or mobile clients.

**Token Types:**

- **Access Token:** Short-lived (configurable, default 15 minutes)
- **Refresh Token:** Longer-lived (7 days), single-use

**Authentication Header:**

```
Authorization: Bearer <access_token>
```

**Flow:**

1. User sends credentials to `POST /auth:login`
2. Server validates credentials and returns both access and refresh tokens
3. Client stores tokens securely (httpOnly cookies or secure storage)
4. Client includes access token in `Authorization: Bearer <token>` header for all requests
5. Before access token expires, client calls `POST /auth:refresh` with refresh token
6. Server validates refresh token, invalidates it, and issues new token pair
7. If refresh token expires or is invalid, user must re-authenticate via `/auth:login`

**Token Properties:**

- **JWT Algorithm:** HS256 (HMAC with SHA-256)
- Access tokens are stateless (JWT claims validated cryptographically)
- Refresh tokens are single-use and invalidated after use
- Multiple concurrent sessions supported (each gets separate refresh token)
- Logout invalidates both the current session's refresh token and access token
- **Token Blacklist:** In-database blacklist for immediately revoking tokens (logout, password changes). On logout, both the refresh token is deleted from the database and the access token is added to the blacklist so it cannot be reused before expiry.

**JWT Claims Structure:**

Access tokens contain the following claims:

- `user_id`: User's ULID identifier (string, from `id` column)
- `username`: User's username (string)
- `email`: User's email address (string)
- `role`: User's role (`admin`, `user`)
- `can_write`: Write permission flag (boolean)
- Standard JWT claims: `iss`, `exp`, `iat`, `sub`

**Rate Limits:**

- Standard requests: 100 requests/minute per user
- Login attempts: 5 attempts per 15 minutes per IP/username

### API Key Access

**Purpose:** For machine-to-machine integrations, automation, and service accounts.

**Key Properties:**

- **Prefix:** All keys start with `moon_live_` for easy identification
- Long-lived credentials with no expiration
- Must be manually rotated or revoked
- Minimum 64 characters after prefix (base62: alphanumeric + `-` + `_`)
- Total length: ~74 characters (`moon_live_` + 64 chars)
- Stored as SHA-256 hashes in database
- Each key assigned a role (`admin`, `user`)
- **Usage Tracking:** `last_used_at` timestamp updated on each request

**Authentication Header:**

```
Authorization: Bearer <api_key>
```

**Flow:**

1. Admin creates API key via `POST /apikeys:create` (specifying name, role, and optional description)
2. Server generates cryptographically secure key with `moon_live_` prefix, returns it once
3. Service stores key securely and includes it in `Authorization: Bearer` header for all requests
4. Keys can be rotated via `POST /apikeys:update` or destroyed via `POST /apikeys:destroy`

**Key Management:**

- API key value returned only once during creation - must be stored securely
- Subsequent API calls only return key metadata (id, name, role, created_at)
- Keys stored as SHA-256 hashes; original value never retrievable
- Admin can list all keys and their metadata via `/apikeys:list`

**Rate Limits:**

- 1000 requests/minute per API key

### Unified Authentication Header

**Standard:**

Both JWT tokens and API keys use the same `Authorization: Bearer` header format:

```
Authorization: Bearer <TOKEN>
```

**Token Type Detection:**

The server automatically detects the token type:

- **JWT tokens:** Three base64-encoded segments separated by dots (e.g., `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`)
- **API keys:** Start with `moon_live_` prefix (e.g., `moon_live_abc123...`)

**Examples:**

```bash
# JWT authentication
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  https://api.moon.example.com/data/users:list

# API key authentication (same header format)
curl -H "Authorization: Bearer moon_live_abc123..." \
  https://api.moon.example.com/data/users:list
```

### Authentication Priority

**Precedence Rules:**

- If no `Authorization: Bearer` header is present on protected endpoint, return `401 Unauthorized`
- If invalid/expired credentials provided, return `401 Unauthorized`
- If valid credentials but insufficient permissions, return `403 Forbidden`

**Configuration Control:**

- JWT authentication: Always enabled (controlled via `jwt.secret` in config)
- API Key authentication: Opt-in (controlled via `apikey.enabled` in config)

## Roles and Permissions

### Role Definitions

**admin Role:**

- Full system access
- Can manage users and API keys
- Can create, read, update, and delete all collections
- Can create, read, update, and delete all data in any collection
- Can access all aggregation and query endpoints
- **Admin Override:** The `can_write` flag is ignored for admin role (always has write access)

**user Role:**

- **Write-enabled by default** (`can_write: true`)
- Can read collections metadata
- Can read data from all collections
- Can use query, filter, and aggregation endpoints
- **Write access controlled per-user via `can_write` flag:**
  - When `can_write: true` (default), user can create, update, and delete data
  - When `can_write: false`, user can only read data
- **Cannot** manage collections schema (create/update/destroy collections)
- **Cannot** manage users or API keys

### Permission Matrix

| Action | Admin | User (can_write: false) | User (can_write: true) |
|--------|-------|-------------------------|------------------------|
| Manage users/apikeys | ✓ | ✗ | ✗ |
| Create/update/delete collections | ✓ | ✗ | ✗ |
| Read collections metadata | ✓ | ✓ | ✓ |
| Read data from collections | ✓ | ✓ | ✓ |
| Create/update/delete data | ✓ | ✗ | ✓ |
| Query/filter/aggregate data | ✓ | ✓ | ✓ |

## Security Configuration

### Password Policy

**Requirements:**

- Minimum 8 characters (configurable)
- Must include: uppercase, lowercase, number
- Optional special characters (configurable)
- Passwords hashed with bcrypt (cost factor: 12)

**Enforcement:** User creation, password change, and admin password reset.

**Password Reset:** Admin-only via `POST /users:update` with `action: reset_password`. Invalidates all user sessions.

### Validation Constraints

**Users:** Email (RFC-compliant), unique username, role (`admin`, `user`)  
**API Keys:** Name (3-100 chars, unique), role, key format (`moon_live_` + 64 base62 chars)  
**Protection:** Cannot delete/demote last admin. Admin cannot modify own role.  
**Cascade:** Deleting user removes all refresh tokens.

### Rate Limiting

- **Login:** 5 failed attempts per 15 minutes per IP/username (auth-specific)
- **Standard limits, response headers, and rate-limit error format:** See [SPEC_API.md § Security > Rate Limiting](SPEC_API.md#rate-limiting)

### Session Management

**Sessions:** Multiple concurrent sessions per user. Each login creates unique refresh token.  
**Refresh Tokens:** Single-use, stored in database, invalidated after use. New token issued on refresh.  
**Invalidation:** Logout (deletes refresh token + blacklists access token for immediate revocation), password change (all sessions revoked + access token blacklisted), admin revoke (`revoke_sessions` action).  
**Cleanup:** Expired tokens should be purged periodically.

### First Admin Bootstrap

On first start, if no admin exists, create from `auth.bootstrap_admin` config section. Remove config section after first login and change password immediately. Never commit credentials to version control.

### Security Best Practices

**Transport:** HTTPS required in production (TLS 1.2+)  
**Token Storage:** httpOnly cookies (web), secure storage (mobile). Never use localStorage/sessionStorage.  
**Secrets:** Never log API keys/passwords. Store JWT secret with restricted permissions (chmod 600).  
**CORS:** Configure via `security.cors.allowed_origins`. No wildcards in production.  
**Audit:** Log all auth attempts, admin actions, rate limit violations. Never log sensitive data.

## Server Implementation

### Middleware Order

1. CORS → 2. Logging → 3. Authentication → 4. Rate Limiting → 5. Authorization → 6. Handler

**Rationale:**
- **CORS** is always first so that browser preflight requests are handled before any other processing.
- **Logging** comes before authentication to ensure all requests (including unauthorized ones) are captured for security auditing and debugging.
- **Authentication** validates the identity of the requester (JWT token or API key).
- **Rate Limiting** is applied after authentication to enable per-user/per-key rate limiting. Authenticated identity is required for accurate rate limit tracking. Login endpoints use a separate IP-based rate limiter.
- **Authorization** verifies that the authenticated entity has the required permissions for the requested action.

JWT takes precedence over API key. Rate limiting uses user/key ID after authentication.

### Database Schema

**users Table:**

```sql
CREATE TABLE users (
  pkid INTEGER PRIMARY KEY AUTOINCREMENT,  -- Internal ID (not exposed)
  id TEXT UNIQUE NOT NULL,                 -- ULID, exposed as "id" in API
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,             -- bcrypt hash
  role TEXT NOT NULL,                      -- "admin" or "user"
  can_write BOOLEAN DEFAULT FALSE,         -- Write permission for user role
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_login_at TIMESTAMP
);

CREATE INDEX idx_users_id ON users(id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
```

**refresh_tokens Table:**

```sql
CREATE TABLE refresh_tokens (
  pkid INTEGER PRIMARY KEY AUTOINCREMENT,
  user_pkid INTEGER NOT NULL,             -- Foreign key to users.pkid
  token_hash TEXT UNIQUE NOT NULL,        -- SHA-256 hash of refresh token
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP,
  FOREIGN KEY (user_pkid) REFERENCES users(pkid) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user_pkid ON refresh_tokens(user_pkid);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);
```

**apikeys Table:**

```sql
CREATE TABLE apikeys (
  pkid INTEGER PRIMARY KEY AUTOINCREMENT,
  id TEXT UNIQUE NOT NULL,                -- ULID, exposed as "id" in API
  name TEXT NOT NULL,
  description TEXT,
  key_hash TEXT UNIQUE NOT NULL,          -- SHA-256 hash of API key
  role TEXT NOT NULL,                     -- "admin" or "user"
  can_write BOOLEAN DEFAULT FALSE,        -- Write permission for user role
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP
);

CREATE INDEX idx_apikeys_id ON apikeys(id);
CREATE INDEX idx_apikeys_hash ON apikeys(key_hash);
```

**Note:** Rate limiting uses in-memory tracking (token bucket or sliding window). Database storage optional.

## API Endpoints

All auth endpoints follow the AIP-136 custom actions pattern (`resource:action`). For all request/response formats, error codes, and pagination patterns, see the references below.

### Authentication Endpoints (Public)

> **📖 Full Reference**: [SPEC_API/020-auth.md](SPEC_API/020-auth.md)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/auth:login` | POST | Authenticate user and receive access + refresh tokens |
| `/auth:logout` | POST | Invalidate current session's refresh token |
| `/auth:refresh` | POST | Exchange refresh token for new access + refresh token pair |
| `/auth:me` | GET | Get current authenticated user information |
| `/auth:me` | POST | Update current user's profile (email, password) |

**Auth-specific notes:**

- `POST /auth:me` password change invalidates all refresh tokens (forces re-login)

---

### User Management Endpoints (Admin Only)

> **📖 Full Reference**: [SPEC_API/030-users.md](SPEC_API/030-users.md)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/users:list` | GET | List all users with pagination |
| `/users:get` | GET | Get specific user by ID |
| `/users:create` | POST | Create new user (admin only — no self-registration) |
| `/users:update` | POST | Update user properties or perform admin actions (`reset_password`, `revoke_sessions`) |
| `/users:destroy` | POST | Delete user account |

**Auth-specific notes:**

- `reset_password` action invalidates all user's refresh tokens
- `revoke_sessions` action invalidates all user's refresh tokens
- Cannot downgrade the last admin user (must always have at least one admin)
- Cannot delete the last admin user
- Deleting a user cascades to all their refresh tokens (via foreign key)

---

### API Key Management Endpoints (Admin Only)

> **📖 Full Reference**: [SPEC_API/040-apikeys.md](SPEC_API/040-apikeys.md)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/apikeys:list` | GET | List all API keys (metadata only — key value never returned) |
| `/apikeys:get` | GET | Get specific API key metadata by ID |
| `/apikeys:create` | POST | Create new API key |
| `/apikeys:update` | POST | Update metadata or rotate key (`rotate` action) |
| `/apikeys:destroy` | POST | Delete API key |

**Auth-specific notes:**

- API key value returned only once during creation (`key` field in response) — store it securely
- `rotate` action invalidates the old key immediately; new key returned once only
- Key value is never retrievable after creation (stored as SHA-256 hash)

---

## Authentication-Specific Error Codes

The following error codes are unique to authentication flows and not covered in SPEC_API.md. For standard error codes and HTTP status codes, see [SPEC_API.md](SPEC_API.md).

> **Error Response:** All errors follow the [Standard Error Response](SPEC_API.md#standard-error-response) for any error handling

---

## Configuration Reference

> **📖 Complete Configuration**: See **[moon.conf](moon.conf)** in the project root for the complete, self-documented configuration file with all authentication options, defaults, security recommendations, and inline documentation.

### Configuration Principles

Authentication configuration is managed via the `moon.conf` YAML file:

- **Location:** Default `/etc/moon.conf` or custom path via `--config` flag
- **Format:** YAML with inline documentation
- **Required:** `jwt.secret` must be set to a secure random string (min 32 chars)
- **Security:** Never commit secrets to version control

### Quick Configuration

**Minimum Required:**

```yaml
jwt:
  secret: "your-secret-key-min-32-chars"  # Generate with: openssl rand -base64 32
```

**Common Options:**

- `jwt.access_expiry`: Access token lifetime (default: 900s / 15 minutes)
- `jwt.refresh_expiry`: Refresh token lifetime (default: 604800s / 7 days)
- `apikey.enabled`: Enable API key authentication (default: false)
- `rate_limit.user_rpm`: JWT user request limit (default: 100/min)
- `rate_limit.apikey_rpm`: API key request limit (default: 1000/min)
- `auth.bootstrap_admin`: First-time admin account setup (remove after creation)

See `moon.conf` for all available options with detailed inline documentation.

**Note:** Configuration loaded from YAML file only (no environment variables). See SPEC.md for general configuration patterns.

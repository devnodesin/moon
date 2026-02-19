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

Both methods support role-based access control (RBAC) with three roles: `admin`, `user`, and `readonly`.

## Access Types

### User Login (JWT-based)

**Purpose:** For interactive users accessing the system via web or mobile clients.

**Token Types:**

- **Access Token:** Short-lived (configurable, default 1 hour)
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
- Logout only invalidates current session's refresh token
- **Token Blacklist:** In-database blacklist for revoked tokens (logout, password changes)

**JWT Claims Structure:**

Access tokens contain the following claims:
- `user_id`: User's ULID identifier (string, from `id` column)
- `username`: User's username (string)
- `email`: User's email address (string)
- `role`: User's role (`admin`, `user`, or `readonly`)
- `can_write`: Write permission flag (boolean)
- `active`: User Active (boolean)
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
- Each key assigned a role (`admin`, `user`, or `readonly`)
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

**readonly Role:**

- **Read-only access** (enforced regardless of `can_write` flag)
- Can read collections metadata
- Can read data from all collections
- Can use query, filter, and aggregation endpoints
- **Cannot** write data even if `can_write` flag is set to true
- **Cannot** manage collections schema
- **Cannot** manage users or API keys

### Permission Matrix

| Action | Admin | User (can_write: false) | User (can_write: true) | Readonly |
|--------|-------|-------------------------|------------------------|----------|
| Manage users/apikeys | ✓ | ✗ | ✗ | ✗ |
| Create/update/delete collections | ✓ | ✗ | ✗ | ✗ |
| Read collections metadata | ✓ | ✓ | ✓ | ✓ |
| Read data from collections | ✓ | ✓ | ✓ | ✓ |
| Create/update/delete data | ✓ | ✗ | ✓ | ✗ |
| Query/filter/aggregate data | ✓ | ✓ | ✓ | ✓ |

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

**Users:** Email (RFC-compliant), unique username, role (`admin`, `user`, `readonly`)  
**API Keys:** Name (3-100 chars, unique), role, key format (`moon_live_` + 64 base62 chars)  
**Protection:** Cannot delete/demote last admin. Admin cannot modify own role.  
**Cascade:** Deleting user removes all refresh tokens.

### Rate Limiting

**Limits:** 100 req/min per JWT user, 1000 req/min per API key, 5 failed login attempts per 15 min  
**Headers:** `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`  
**Response:** `429 Too Many Requests` when exceeded

### Session Management

**Sessions:** Multiple concurrent sessions per user. Each login creates unique refresh token.  
**Refresh Tokens:** Single-use, stored in database, invalidated after use. New token issued on refresh.  
**Invalidation:** Logout (current session only), password change (all sessions), admin revoke (`revoke_sessions` action).  
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

1. CORS → 2. Rate Limiting → 3. Authentication → 4. Authorization → 5. Validation → 6. Logging → 7. Handler → 8. Error Handling

JWT takes precedence over API key. Rate limiting uses user/key ID.

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

> **📖 Complete API Reference**: For standardized request/response formats, error codes, and pagination patterns used by authentication endpoints, see **[SPEC_API.md](SPEC_API.md)**.
>
> This section documents authentication-specific endpoints. These endpoints follow the same response patterns as other Moon APIs.

All auth endpoints follow the AIP-136 custom actions pattern (resource:action).

### Authentication Endpoints (Public)

#### POST /auth:login

**Purpose:** Authenticate user and receive access + refresh tokens

> **📖 Request/Response**: See [SPEC_API.md § Login](SPEC_API.md#login) for full request/response format.
>
> **📖 Error Responses**: See [SPEC_API.md § Standard Error Response](SPEC_API.md#standard-error-response). Common codes: `INVALID_CREDENTIALS` (401), `MISSING_REQUIRED_FIELD` (400), `RATE_LIMIT_EXCEEDED` (429).

---

#### POST /auth:logout

**Purpose:** Invalidate current session's refresh token

> **📖 Request/Response**: See [SPEC_API.md § Logout](SPEC_API.md#logout) for full request/response format.
>
> **📖 Error Responses**: See [SPEC_API.md § Standard Error Response](SPEC_API.md#standard-error-response). Common codes: `UNAUTHORIZED` (401), `MISSING_REQUIRED_FIELD` (400).

---

#### POST /auth:refresh

**Purpose:** Exchange refresh token for new access + refresh token pair

> **📖 Request/Response**: See [SPEC_API.md § Refresh Token](SPEC_API.md#refresh-token) for full request/response format.
>
> **📖 Error Responses**: See [SPEC_API.md § Standard Error Response](SPEC_API.md#standard-error-response). Common codes: `EXPIRED_TOKEN` (401), `REVOKED_TOKEN` (401), `MISSING_REQUIRED_FIELD` (400).

---

#### GET /auth:me

**Purpose:** Get current authenticated user information

> **📖 Request/Response**: See [SPEC_API.md § Get Current User](SPEC_API.md#get-current-user) for full response format.
>
> **📖 Error Responses**: See [SPEC_API.md § Standard Error Response](SPEC_API.md#standard-error-response). Common codes: `UNAUTHORIZED` (401).

---

#### POST /auth:me

**Purpose:** Update current user's profile (email, password)

> **📖 Request/Response**: See [SPEC_API.md § Update Current User](SPEC_API.md#update-current-user) for full request/response format.
>
> **📖 Error Responses**: See [SPEC_API.md § Standard Error Response](SPEC_API.md#standard-error-response). Common codes: `UNAUTHORIZED` (401), `VALIDATION_ERROR` (400), `EMAIL_EXISTS` (409).

**Notes:**

- Password change invalidates all refresh tokens (forces re-login)
- Email change requires re-verification (if email verification enabled)

---

### User Management Endpoints (Admin Only)

#### GET /users:list

**Purpose:** List all users with pagination

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**

- `limit` (optional, default: 50, max: 100)
- `after` (optional): Cursor for pagination (user ULID)
- `role` (optional): Filter by role ("admin" or "user")

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "can_write": true,
      "created_at": "2024-01-10T08:00:00Z",
      "last_login_at": "2024-01-16T09:00:00Z"
    },
    {
      "id": "01ARZ3NDEKTSV4RRFFQ69G5FBX",
      "username": "user1",
      "email": "user1@example.com",
      "role": "user",
      "can_write": false,
      "created_at": "2024-01-12T10:30:00Z",
      "last_login_at": "2024-01-15T14:20:00Z"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 50,
    "next": "01ARZ3NDEKTSV4RRFFQ69G5FBX",
    "prev": null
  }
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:list` Endpoints](SPEC_API.md#standard-response-pattern-for-list-endpoints) for pagination details.
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format. Common codes: `UNAUTHORIZED` (401), `FORBIDDEN` (403).

---

#### GET /users:get

**Purpose:** Get specific user by ID

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**

- `id` (required): User ULID

**Response (200 OK):**

```json
{
  "data": {
    "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "username": "user1",
    "email": "user1@example.com",
    "role": "user",
    "can_write": false,
    "created_at": "2024-01-12T10:30:00Z",
    "updated_at": "2024-01-15T11:00:00Z",
    "last_login_at": "2024-01-15T14:20:00Z"
  }
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:get` Endpoints](SPEC_API.md#standard-response-pattern-for-get-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format. Common codes: `UNAUTHORIZED` (401), `FORBIDDEN` (403), `RECORD_NOT_FOUND` (404).

---

#### POST /users:create

**Purpose:** Create new user (admin only - no self-registration)

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Request:**

```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "SecurePass123",
  "role": "user",
  "can_write": false
}
```

**Response (201 Created):**

```json
{
  "data": {
    "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "username": "newuser",
    "email": "newuser@example.com",
    "role": "user",
    "can_write": false,
    "created_at": "2024-01-16T15:30:00Z"
  },
  "message": "User created successfully"
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:create` Endpoints](SPEC_API.md#standard-response-pattern-for-create-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format. Common codes: `UNAUTHORIZED` (401), `FORBIDDEN` (403), `VALIDATION_ERROR` (400), `USERNAME_EXISTS` (409), `EMAIL_EXISTS` (409).

---

#### POST /users:update

**Purpose:** Update user properties or perform admin actions

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**

- `id` (required): User ULID

**Request (update role and permissions):**

```json
{
  "role": "user",
  "can_write": true
}
```

**Request (reset password):**

```json
{
  "action": "reset_password",
  "new_password": "NewSecurePass456"
}
```

**Request (revoke all sessions):**

```json
{
  "action": "revoke_sessions"
}
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "username": "user1",
    "email": "user1@example.com",
    "role": "user",
    "can_write": true,
    "created_at": "2024-01-12T10:30:00Z",
    "updated_at": "2024-01-16T16:00:00Z"
  },
  "message": "User updated successfully"
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:update` Endpoints](SPEC_API.md#standard-response-pattern-for-update-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format. Common codes: `UNAUTHORIZED` (401), `FORBIDDEN` (403), `RECORD_NOT_FOUND` (404), `INVALID_ACTION` (400).

**Notes:**

- Password reset invalidates all user's refresh tokens
- Revoking sessions invalidates all user's refresh tokens
- Cannot downgrade the last admin user to regular user (must have at least one admin)

---

#### POST /users:destroy

**Purpose:** Delete user account

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**

- `id` (required): User ULID

**Response (200 OK):**

```json
{
  "message": "User deleted successfully"
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:destroy` Endpoints](SPEC_API.md#standard-response-pattern-for-destroy-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format. Common codes: `UNAUTHORIZED` (401), `FORBIDDEN` (403), `RECORD_NOT_FOUND` (404).

**Notes:**

- Cannot delete the last admin user (must have at least one admin)
- Deleting user cascades to refresh_tokens (via foreign key)

---

### API Key Management Endpoints (Admin Only)

#### GET /apikeys:list

**Purpose:** List all API keys (metadata only)

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**

- `limit` (optional, default: 50, max: 100)
- `after` (optional): Cursor for pagination (key ULID)

**Response (200 OK):**

```json
{
  "data": [
    {
      "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
      "name": "Production Service",
      "description": "Main API integration",
      "role": "user",
      "can_write": true,
      "created_at": "2024-01-10T10:00:00Z",
      "last_used_at": "2024-01-16T14:30:00Z"
    },
    {
      "id": "01ARZ3NDEKTSV4RRFFQ69G5FBX",
      "name": "Analytics Service",
      "description": "Read-only analytics integration",
      "role": "user",
      "can_write": false,
      "created_at": "2024-01-12T11:00:00Z",
      "last_used_at": "2024-01-16T15:00:00Z"
    }
  ],
  "meta": {
    "count": 2,
    "limit": 50,
    "next": "01ARZ3NDEKTSV4RRFFQ69G5FBX",
    "prev": null
  }
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:list` Endpoints](SPEC_API.md#standard-response-pattern-for-list-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format.

**Notes:**

- Actual API key value is never returned (only metadata)

---

#### GET /apikeys:get

**Purpose:** Get specific API key metadata by ID

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**

- `id` (required): API key ULID

**Response (200 OK):**

```json
{
  "data": {
    "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "name": "Production Service",
    "description": "Main API integration",
    "role": "user",
    "can_write": true,
    "created_at": "2024-01-10T10:00:00Z",
    "last_used_at": "2024-01-16T14:30:00Z"
  }
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:get` Endpoints](SPEC_API.md#standard-response-pattern-for-get-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format.

---

#### POST /apikeys:create

**Purpose:** Create new API key

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Request:**

```json
{
  "name": "New Service",
  "description": "Optional description",
  "role": "user",
  "can_write": false
}
```

**Response (201 Created):**

```json
{
  "data": {
    "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "name": "New Service",
    "description": "Optional description",
    "role": "user",
    "can_write": false,
    "key": "moon_live_abc123...xyz789",
    "created_at": "2024-01-16T16:30:00Z"
  },
  "message": "API key created successfully",
  "warning": "Store this key securely. It will not be shown again."
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:create` Endpoints](SPEC_API.md#standard-response-pattern-for-create-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format. Common codes: `VALIDATION_ERROR` (400), `APIKEY_NAME_EXISTS` (409).

**Notes:**

- API key value returned only once during creation
- Key format: `moon_live_` prefix + 64 characters (base62)
- Key stored as SHA-256 hash in database

---

#### POST /apikeys:update

**Purpose:** Update API key metadata or rotate key

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**

- `id` (required): API key ULID

**Request (update metadata):**

```json
{
  "name": "Updated Service Name",
  "description": "Updated description",
  "can_write": true
}
```

**Request (rotate key):**

```json
{
  "action": "rotate"
}
```

**Response (200 OK - metadata update):**

```json
{
  "data": {
    "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "name": "Updated Service Name",
    "description": "Updated description",
    "role": "user",
    "can_write": true,
    "updated_at": "2024-01-16T17:00:00Z"
  },
  "message": "API key updated successfully"
}
```

**Response (200 OK - key rotation):**

```json
{
  "data": {
    "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "name": "Production Service",
    "key": "moon_live_def456...uvw012",
    "created_at": "2024-01-16T17:00:00Z"
  },
  "message": "API key rotated successfully",
  "warning": "Store this key securely. The old key is now invalid."
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:update` Endpoints](SPEC_API.md#standard-response-pattern-for-update-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format. Common codes: `RECORD_NOT_FOUND` (404), `INVALID_ACTION` (400).

**Notes:**

- Key rotation invalidates old key immediately
- New key returned only once after rotation

---

#### POST /apikeys:destroy

**Purpose:** Delete API key

**Headers:**

```
Authorization: Bearer <admin_access_token>
```

**Query Parameters:**

- `id` (required): API key ULID

**Response (200 OK):**

```json
{
  "message": "API key deleted successfully"
}
```

> **📖 Response Pattern**: See [SPEC_API.md § Standard Response Pattern for `:destroy` Endpoints](SPEC_API.md#standard-response-pattern-for-destroy-endpoints).
>
> **📖 Error Responses**: See [SPEC_API.md](SPEC_API.md) for standard error format.

---

## Authentication-Specific Error Codes

The following error codes are unique to authentication flows and not covered in SPEC_API.md. For standard error codes and HTTP status codes, see [SPEC_API.md](SPEC_API.md).

**Authentication Errors (401):**

- `MISSING_AUTH_HEADER`: No Authorization header provided
- `INVALID_TOKEN_FORMAT`: Authorization header not in "Bearer <token>" format
- `INVALID_TOKEN`: Token signature invalid or token malformed
- `EXPIRED_TOKEN`: Access token or refresh token has expired
- `REVOKED_TOKEN`: Refresh token has been revoked or already used
- `INVALID_CREDENTIALS`: Username/password combination incorrect
- `INVALID_API_KEY`: API key does not exist or is invalid

**Authorization Errors (403):**

- `INSUFFICIENT_PERMISSIONS`: User/API key lacks required role or permission
- `ADMIN_REQUIRED`: Endpoint requires admin role
- `WRITE_PERMISSION_REQUIRED`: User role requires can_write flag for this action
- `CANNOT_DELETE_LAST_ADMIN`: Cannot delete the last admin user
- `CANNOT_MODIFY_SELF_ROLE`: Admin cannot change their own role

**Validation Errors (400):**

- `WEAK_PASSWORD`: Password does not meet security policy
- `INVALID_ROLE`: Role must be "admin", "user", or "readonly"

**Conflict Errors (409):**

- `USERNAME_EXISTS`: Username already taken
- `EMAIL_EXISTS`: Email already registered
- `APIKEY_NAME_EXISTS`: API key name already in use

**Rate Limit Errors (429):**

- `LOGIN_ATTEMPTS_EXCEEDED`: Too many failed login attempts

> **📖 Standard Error Format**: All errors follow the `{error: {code, message}}` format defined in [SPEC_API.md](SPEC_API.md).

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
- `jwt.access_expiry`: Access token lifetime (default: 3600s / 1 hour)
- `jwt.refresh_expiry`: Refresh token lifetime (default: 604800s / 7 days)
- `apikey.enabled`: Enable API key authentication (default: false)
- `rate_limit.user_rpm`: JWT user request limit (default: 100/min)
- `rate_limit.apikey_rpm`: API key request limit (default: 1000/min)
- `auth.bootstrap_admin`: First-time admin account setup (remove after creation)

See `moon.conf` for all available options with detailed inline documentation.

**Note:** Configuration loaded from YAML file only (no environment variables). See SPEC.md for general configuration patterns.

# Authentication Middleware Design Document

**Project:** Moon - Dynamic Headless Engine  
**Document Type:** Product Requirements Document (PRD)  
**Version:** 1.0  
**Date:** 2026-01-31  
**Status:** Design Phase

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Goals and Objectives](#goals-and-objectives)
3. [Critical Design Questions](#critical-design-questions)
4. [System Architecture](#system-architecture)
5. [Authentication Mechanisms](#authentication-mechanisms)
6. [User Management](#user-management)
7. [API Key Management](#api-key-management)
8. [Authorization and Access Control](#authorization-and-access-control)
9. [Security Best Practices](#security-best-practices)
10. [API Endpoints Specification](#api-endpoints-specification)
11. [Database Schema](#database-schema)
12. [Configuration](#configuration)
13. [Implementation Phases](#implementation-phases)
14. [Testing Strategy](#testing-strategy)
15. [Monitoring and Auditing](#monitoring-and-auditing)
16. [Migration and Backwards Compatibility](#migration-and-backwards-compatibility)

---

## 1. Executive Summary

This document outlines the design for a comprehensive authentication and authorization middleware system for the Moon API server. The system will support both **JWT-based authentication** for users and **API key authentication** for machine-to-machine access, with robust role-based access control (RBAC) for fine-grained permissions.

### Key Features

- Dual authentication: JWT tokens and API keys
- Role-based access control (RBAC) with granular permissions
- Secure root user bootstrapping
- Token refresh and rotation mechanisms
- API key lifecycle management (create, rotate, revoke)
- Comprehensive audit logging
- Zero-downtime migration from unauthenticated to authenticated mode

### Design Principles

1. **Security by Default:** All endpoints protected unless explicitly marked public
2. **Least Privilege:** Users and API keys granted minimal necessary permissions
3. **Zero Trust:** Every request validated and authorized
4. **Auditability:** All authentication and authorization events logged
5. **Simplicity:** Clear, predictable API patterns following existing AIP-136 conventions
6. **Performance:** Minimal latency impact (<1ms for token validation)

---

## 2. Goals and Objectives

### Primary Goals

1. **Secure the API:** Protect all endpoints with authentication and authorization
2. **Enable Multi-tenancy:** Support multiple users with different permission levels
3. **Support Automation:** Provide API keys for service-to-service communication
4. **Maintain Compatibility:** Preserve existing API patterns and conventions
5. **Operational Excellence:** Enable secure bootstrapping and management

### Non-Goals

- OAuth 2.0 / OpenID Connect (future consideration)
- Multi-factor authentication (MFA) (Phase 2)
- Social login providers (out of scope)
- Session-based authentication (stateless JWT only)

---

## 3. Critical Design Questions

### Q1: How will users authenticate?

**Answer:**
- **JWT Tokens:** Users authenticate via username/password and receive:
  - **Access Token:** Short-lived (15 minutes), used for API requests
  - **Refresh Token:** Longer-lived (7 days), used to obtain new access tokens
- **API Keys:** Long-lived static keys for machine-to-machine access

**Rationale:**
- JWT provides stateless, scalable authentication suitable for distributed systems
- Refresh tokens balance security (short-lived access) with UX (fewer re-authentications)
- API keys support automation and service accounts without user credentials

---

### Q2: How will the root user be created?

**Answer:**
- **Bootstrap Mode:** On first startup when no users exist:
  - Server checks for `MOON_ROOT_PASSWORD` environment variable
  - If set, creates root user with predefined username `root`
  - If not set, generates secure random password and logs to console/file
  - Root user creation is logged and can only happen once
- **Post-Bootstrap:** Root user can create additional admin users
- **Security:** Root password must be changed on first login (force password reset)

**Rationale:**
- Environment variable allows automated deployment while keeping secrets out of code
- Console output enables manual setup for development
- One-time bootstrap prevents accidental overwrites
- Forced password change ensures default/generated passwords aren't kept

---

### Q3: Where will users and API keys be stored?

**Answer:**
- **Users Table:** Stores user credentials and metadata
  - `id` (ULID): Unique identifier
  - `username`: Unique, indexed
  - `email`: Unique, indexed, nullable
  - `password_hash`: bcrypt hash (cost factor 12)
  - `role`: Enum (root, admin, editor, viewer, api)
  - `must_change_password`: Boolean flag
  - `created_at`, `updated_at`: Timestamps
  - `last_login_at`: Timestamp, nullable

- **API Keys Table:** Stores API key credentials and permissions
  - `id` (ULID): Unique identifier
  - `name`: Human-readable name/description
  - `key_hash`: SHA-256 hash of the API key
  - `key_prefix`: First 8 chars (e.g., `moon_abc1`) for identification
  - `permissions`: JSON array of permission strings
  - `expires_at`: Timestamp, nullable
  - `last_used_at`: Timestamp, nullable
  - `created_by`: User ID who created the key
  - `created_at`, `revoked_at`: Timestamps

- **Refresh Tokens Table:** Tracks issued refresh tokens for rotation/revocation
  - `id` (ULID): Unique identifier
  - `token_hash`: SHA-256 hash of the refresh token
  - `user_id`: Foreign key to users table
  - `expires_at`: Timestamp
  - `revoked_at`: Timestamp, nullable
  - `replaced_by`: ULID of replacement token (for rotation tracking)
  - `created_at`: Timestamp

**Storage Location:** Same database as application data (SQLite/PostgreSQL/MySQL)

**Rationale:**
- Centralized storage simplifies deployment and backup
- Hashed credentials never store plaintext secrets
- Key prefixes enable safe logging and debugging
- Separate tables allow independent lifecycle management
- Refresh token tracking enables secure rotation and revocation

---

### Q4: What are the best ways to expose authentication endpoints?

**Answer:**
- **Follow AIP-136 Pattern:** Use custom actions with colon separator
- **Prefix Consistency:** Respect configured URL prefix (e.g., `/api/v1`)
- **Public Endpoints:** Authentication endpoints are public (no auth required)
- **Grouping:** All auth endpoints under `/auth` resource

**Endpoint Structure:**
```
POST /auth:login           # User login (returns access + refresh tokens)
POST /auth:refresh         # Refresh access token
POST /auth:logout          # Revoke refresh token
POST /auth:change-password # Change own password
GET  /auth:whoami          # Get current user info

POST /users:create         # Create new user (admin only)
POST /users:update         # Update user (admin or self)
POST /users:delete         # Delete user (admin only)
GET  /users:list           # List users (admin only)
GET  /users:get            # Get user details (admin or self)

POST /apikeys:create       # Create API key (admin only)
POST /apikeys:revoke       # Revoke API key (admin or owner)
GET  /apikeys:list         # List API keys (admin or own)
GET  /apikeys:get          # Get API key details (admin or owner)
```

**Rationale:**
- Consistent with existing collection management patterns
- Clear separation between authentication and user management
- Custom actions clearly indicate state-changing operations
- RESTful conventions for resource management

---

### Q5: What endpoints are needed?

**Answer:**

**Authentication Endpoints (Public):**
1. `POST /auth:login` - User login
2. `POST /auth:refresh` - Refresh access token
3. `POST /auth:logout` - Logout (revoke refresh token)
4. `GET /auth:whoami` - Get current user info (authenticated)
5. `POST /auth:change-password` - Change own password (authenticated)

**User Management Endpoints (Authenticated):**
6. `POST /users:create` - Create new user
7. `GET /users:list` - List all users
8. `GET /users:get` - Get user details
9. `POST /users:update` - Update user
10. `POST /users:delete` - Delete user

**API Key Management Endpoints (Authenticated):**
11. `POST /apikeys:create` - Create API key
12. `GET /apikeys:list` - List API keys
13. `GET /apikeys:get` - Get API key details
14. `POST /apikeys:revoke` - Revoke API key
15. `POST /apikeys:rotate` - Rotate API key

**Permission Management Endpoints (Authenticated, Admin Only):**
16. `GET /permissions:list` - List available permissions
17. `GET /roles:list` - List available roles and their permissions

**See Section 10 for detailed specifications.**

---

### Q6: How to add, remove, and manage users and API keys?

**Answer:**

**User Lifecycle:**
1. **Create:** Admin calls `POST /users:create` with username, email, initial password, and role
2. **Update:** Admin or user calls `POST /users:update` to change email, role (admin only), or reset password
3. **Delete:** Admin calls `POST /users:delete` to deactivate user (soft delete recommended)
4. **List/Search:** Admin calls `GET /users:list` with optional filters

**API Key Lifecycle:**
1. **Create:** 
   - Admin calls `POST /apikeys:create` with name, permissions, and optional expiry
   - Server generates cryptographically secure random key (e.g., 32 bytes, base64-encoded)
   - Server returns full key **once** (client must save it)
   - Server stores only hashed key and prefix
2. **Rotate:**
   - User calls `POST /apikeys:rotate` with key ID
   - Server generates new key, revokes old key atomically
   - Returns new key once
3. **Revoke:**
   - Admin or owner calls `POST /apikeys:revoke` with key ID
   - Server marks key as revoked (soft delete with timestamp)
4. **List:**
   - Admin sees all keys; regular users see only their own
   - Response shows prefix, name, permissions, creation date, last used date
   - **Never** returns full key or hash

**Audit Trail:**
- All create/update/delete operations logged with actor, timestamp, and details
- Last login time tracked for users
- Last used time tracked for API keys

---

### Q7: How to check and refresh authentication?

**Answer:**

**Token Validation (Every Request):**
1. **Extract Token:** Check `Authorization: Bearer <token>` header or `X-API-Key` header
2. **Validate Signature:** Verify JWT signature or hash API key
3. **Check Expiration:** Ensure access token or API key hasn't expired
4. **Check Revocation:** For refresh tokens and API keys, check database for revocation
5. **Extract Claims:** Get user ID, role, and permissions from token/key
6. **Cache Results:** Cache successful validations for 1 minute to reduce database load

**Refresh Flow:**
1. Client receives 401 Unauthorized when access token expires
2. Client calls `POST /auth:refresh` with refresh token (from cookie or storage)
3. Server validates refresh token:
   - Check signature and expiration
   - Verify token exists in database and isn't revoked
4. Server generates new access token and new refresh token
5. Server marks old refresh token as replaced (for audit trail)
6. Server returns both tokens
7. Client uses new access token for subsequent requests

**Rotation Strategy:**
- **Access Tokens:** Never refreshed, always short-lived (15 min)
- **Refresh Tokens:** Rotated on every use (single-use tokens)
- **API Keys:** Manually rotated by user (recommended every 90 days)

**Performance:**
- Access token validation: <1ms (signature check only, no database)
- API key validation: <5ms (database lookup with caching)
- Refresh token validation: <10ms (database lookup and write)

---

### Q8: What authentication methods should be supported?

**Answer:**

**Supported Methods:**
1. **JWT Bearer Tokens:** Primary method for user authentication
2. **API Keys:** Secondary method for service accounts and automation

**Not Supported (Initial Release):**
- Basic Auth (username/password on every request)
- OAuth 2.0 / OpenID Connect
- SAML
- Multi-factor authentication (MFA)

**Rationale:**
- JWT provides stateless, scalable authentication
- API keys support automation without complex token management
- Basic auth sends credentials on every request (security risk)
- OAuth/MFA can be added in future phases

---

### Q9: How to handle password security?

**Answer:**

**Password Requirements:**
- Minimum 8 characters
- Must contain: uppercase, lowercase, number, special character
- Maximum 72 characters (bcrypt limit)
- Common password dictionary check (e.g., "password123" rejected)

**Password Storage:**
- Use bcrypt with cost factor 12 (recommended for 2024)
- Never store plaintext passwords
- Never log or transmit passwords (except during initial login)

**Password Reset:**
- Admin can reset user password (generates new temporary password)
- User forced to change password on next login (`must_change_password` flag)
- Old password required to change password (except on forced reset)

**Password Change Policy:**
- Root user must change password on first login
- Recommend password rotation every 90 days (logged warning, not enforced)
- Cannot reuse last 3 passwords (stored as hashes in password_history table)

**Account Lockout:**
- Lock account after 5 failed login attempts in 15 minutes
- Unlock automatically after 30 minutes or manual admin unlock
- Lockout events logged for security monitoring

---

### Q10: What security headers and middleware should be applied?

**Answer:**

**Required Security Headers:**
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Referrer-Policy: no-referrer
```

**Middleware Stack (Order Matters):**
1. **CORS Middleware:** Handle cross-origin requests
2. **Rate Limiting:** Prevent brute force (10 req/min for auth endpoints)
3. **Request ID:** Generate unique ID for tracing
4. **Logging:** Log all requests with sanitized headers
5. **Authentication:** Validate JWT/API key and extract claims
6. **Authorization:** Check permissions against required role/permissions
7. **Business Logic:** Execute actual request handler

**Token Transmission:**
- **JWT Tokens:** `Authorization: Bearer <token>` header only
- **Refresh Tokens:** HTTP-only, Secure, SameSite=Strict cookies preferred
- **API Keys:** `X-API-Key: <key>` header only
- **Never** send tokens in URL query parameters (logged, cached)

---

## 4. System Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Application                      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       API Gateway                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Rate Limiter → CORS → Request ID → Logger          │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Authentication Middleware                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Token Validator → User Context → Permission Check  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Route Handler (Business Logic)             │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                         Database                             │
│  Users │ API Keys │ Refresh Tokens │ Collections │ Data    │
└─────────────────────────────────────────────────────────────┘
```

### Authentication Flow

```
Client                    Server                   Database
  │                         │                         │
  ├─ POST /auth:login ────▶│                         │
  │  (username, password)   │                         │
  │                         ├─ Verify credentials ──▶│
  │                         │◀─ User record ──────────┤
  │                         │                         │
  │                         ├─ Generate tokens        │
  │                         │  - Access (15m)         │
  │                         │  - Refresh (7d)         │
  │                         │                         │
  │                         ├─ Store refresh hash ──▶│
  │◀─ 200 OK ───────────────┤                         │
  │  {access, refresh}      │                         │
  │                         │                         │
  ├─ GET /collections:list ▶│                         │
  │  Authorization: Bearer  │                         │
  │                         │                         │
  │                         ├─ Validate token         │
  │                         │  - Check signature      │
  │                         │  - Check expiration     │
  │                         │  - Extract user_id      │
  │                         │                         │
  │                         ├─ Check permissions     │
  │◀─ 200 OK ───────────────┤                         │
  │                         │                         │
  ├─ POST /auth:refresh ───▶│                         │
  │  (refresh_token)        │                         │
  │                         ├─ Validate refresh ────▶│
  │                         │◀─ Token record ─────────┤
  │                         │                         │
  │                         ├─ Generate new tokens    │
  │                         ├─ Revoke old token ────▶│
  │◀─ 200 OK ───────────────┤                         │
  │  {access, refresh}      │                         │
```

---

## 5. Authentication Mechanisms

### JWT Token Structure

**Access Token (15-minute expiry):**
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "01HJ8X1234567890ABCDEF",     // User ID (ULID)
    "username": "admin",
    "role": "admin",
    "permissions": ["*"],
    "iat": 1706716800,                   // Issued at
    "exp": 1706717700,                   // Expires at (15 min)
    "iss": "moon-api",                   // Issuer
    "aud": "moon-api",                   // Audience
    "jti": "01HJ8Y9876543210ZYXWVU"      // JWT ID (ULID, for revocation)
  },
  "signature": "..."
}
```

**Refresh Token (7-day expiry):**
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "01HJ8X1234567890ABCDEF",     // User ID (ULID)
    "type": "refresh",
    "iat": 1706716800,
    "exp": 1707321600,                   // 7 days
    "iss": "moon-api",
    "aud": "moon-api",
    "jti": "01HJ8Z1111111111REFRESH"     // Unique ID for rotation tracking
  },
  "signature": "..."
}
```

**Key Management:**
- JWT secret stored in `jwt.secret` config (must be set, no default)
- Minimum 256-bit (32 bytes) random key
- Rotate secret with graceful transition:
  - Support two secrets temporarily (old and new)
  - Issue tokens with new secret
  - Accept tokens signed with either secret during transition (24 hours)
  - After transition, reject old secret

---

## 6. User Management

### User Roles

| Role     | Description                                      | Permissions                                    |
|----------|--------------------------------------------------|------------------------------------------------|
| `root`   | Superuser with full system access               | All permissions (implicitly granted)           |
| `admin`  | Administrator with user and key management       | Manage users, API keys, collections, and data  |
| `editor` | Can create and modify data                       | Create, read, update collections and data      |
| `viewer` | Read-only access                                 | Read collections and data                      |
| `api`    | Service account for API key usage (no login)     | Custom permissions per key                     |

### Permission Model

**Permissions are hierarchical:**
- `collections:*` - All collection operations
  - `collections:create` - Create collections
  - `collections:read` - Read collection schemas
  - `collections:update` - Update collection schemas
  - `collections:delete` - Delete collections
- `data:*` - All data operations
  - `data:create` - Create records
  - `data:read` - Read records
  - `data:update` - Update records
  - `data:delete` - Delete records
- `users:*` - All user management operations
  - `users:create` - Create users
  - `users:read` - Read user info
  - `users:update` - Update users
  - `users:delete` - Delete users
- `apikeys:*` - All API key operations
  - `apikeys:create` - Create API keys
  - `apikeys:read` - Read API key info
  - `apikeys:revoke` - Revoke API keys

**Role → Permission Mapping:**
- `root`: `*` (wildcard, all permissions)
- `admin`: `collections:*`, `data:*`, `users:*`, `apikeys:*`
- `editor`: `collections:*`, `data:*`
- `viewer`: `collections:read`, `data:read`
- `api`: Custom per key (typically `data:*` for read/write services)

**Special Rules:**
- Users can always read and update their own profile
- Users can change their own password
- Root user cannot be deleted
- At least one admin user must exist

---

## 7. API Key Management

### API Key Format

```
moon_<environment>_<random>_<checksum>

Example: moon_prod_a7b9c3d2e1f4g5h6_k8j9
```

**Components:**
- `moon_` - Fixed prefix for identification
- `<environment>` - `dev`, `test`, `prod` (from config)
- `<random>` - 16 bytes of cryptographically secure random data (base32-encoded)
- `<checksum>` - 4-byte CRC32 checksum for error detection

**Storage:**
- Full key displayed to user **once** on creation
- Only SHA-256 hash stored in database
- Key prefix (first 12 chars) stored for identification in logs

### Key Permissions

API keys have explicit permissions (no roles):

```json
{
  "id": "01HJ8X1234567890ABCDEF",
  "name": "Production Data Sync Service",
  "key_prefix": "moon_prod_a7",
  "permissions": [
    "data:create",
    "data:read",
    "data:update"
  ],
  "expires_at": "2027-01-31T00:00:00Z",
  "created_by": "01HJ8X9876543210ADMIN",
  "created_at": "2026-01-31T12:00:00Z",
  "last_used_at": "2026-01-31T18:45:12Z"
}
```

### Key Rotation Strategy

**Recommended Rotation Frequency:**
- Production keys: Every 90 days
- Development keys: Every 180 days
- Keys with elevated permissions: Every 30 days

**Rotation Process:**
1. User calls `POST /apikeys:rotate` with key ID
2. Server generates new key with same permissions
3. Server marks old key as "pending revocation" (grace period: 24 hours)
4. Server returns new key to user
5. After grace period, old key fully revoked

**Grace Period Benefits:**
- Allows time to update services with new key
- Prevents immediate downtime
- Old key still visible in logs/audit for correlation

---

## 8. Authorization and Access Control

### Middleware Authorization Check

For each protected endpoint:

1. **Extract Authentication:**
   - Check `Authorization: Bearer <token>` header for JWT
   - Check `X-API-Key: <key>` header for API key
   - If neither present, return 401 Unauthorized

2. **Validate Authentication:**
   - Verify token signature and expiration
   - Look up API key hash in database
   - If invalid, return 401 Unauthorized

3. **Load User Context:**
   - Extract user ID from token or API key
   - Load user role and permissions
   - Attach to request context

4. **Check Endpoint Permission:**
   - Determine required permission for endpoint (e.g., `collections:create`)
   - Check if user has required permission (exact match or wildcard)
   - If unauthorized, return 403 Forbidden

5. **Additional Checks:**
   - Resource-level permissions (e.g., can user update *this* collection?)
   - Rate limiting per user/key
   - Account lockout status

### Public Endpoints

Endpoints that bypass authentication:
- `GET /health` - Health check
- `GET /doc/*` - Documentation
- `POST /auth:login` - Login
- `POST /auth:refresh` - Refresh token

### Protected Endpoint Examples

```go
// Pseudo-code for endpoint declaration
endpoints := []Endpoint{
  {
    Path: "/collections:create",
    Method: "POST",
    Handler: CreateCollectionHandler,
    RequiredPermission: "collections:create",
  },
  {
    Path: "/users:list",
    Method: "GET",
    Handler: ListUsersHandler,
    RequiredPermission: "users:read",
  },
  {
    Path: "/auth:whoami",
    Method: "GET",
    Handler: WhoAmIHandler,
    RequiredPermission: "", // Any authenticated user
  },
}
```

---

## 9. Security Best Practices

### Password Security

- ✅ Bcrypt with cost factor 12
- ✅ Enforce strong password requirements
- ✅ Rate limit login attempts (10/min per IP)
- ✅ Account lockout after 5 failed attempts
- ✅ Never log or transmit passwords (except initial login)
- ✅ Forced password change for default/temporary passwords

### Token Security

- ✅ Short-lived access tokens (15 minutes)
- ✅ Refresh token rotation (single-use)
- ✅ Strong JWT secret (256-bit minimum)
- ✅ Signature verification on every request
- ✅ Refresh token revocation support
- ✅ HTTP-only, Secure cookies for refresh tokens (web)

### API Key Security

- ✅ Cryptographically secure random generation
- ✅ Hash keys with SHA-256 before storage
- ✅ Display full key only once on creation
- ✅ Regular rotation (90-day recommended)
- ✅ Per-key permissions (least privilege)
- ✅ Expiration dates on keys

### Transport Security

- ✅ HTTPS/TLS 1.2+ required (enforced via HSTS)
- ✅ Reject HTTP connections in production
- ✅ Secure cookie flags (HttpOnly, Secure, SameSite)
- ✅ Never send tokens in URL query parameters

### Database Security

- ✅ Parameterized queries (prevent SQL injection)
- ✅ Encrypted credentials at rest
- ✅ Limited database user permissions
- ✅ Regular backups of user and key tables

### Audit and Monitoring

- ✅ Log all authentication attempts (success and failure)
- ✅ Log all authorization failures
- ✅ Track API key and token usage
- ✅ Alert on anomalous patterns (impossible travel, mass failures)
- ✅ Retain logs for 90 days minimum

---

## 10. API Endpoints Specification

### Authentication Endpoints

#### POST /auth:login

**Description:** Authenticate user and receive access and refresh tokens.

**Request:**
```json
{
  "username": "admin",
  "password": "securePassword123!"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "01HJ8X1234567890ABCDEF",
    "username": "admin",
    "email": "admin@example.com",
    "role": "admin",
    "must_change_password": false
  }
}
```

**Errors:**
- `401 Unauthorized`: Invalid credentials
- `403 Forbidden`: Account locked
- `429 Too Many Requests`: Rate limit exceeded

---

#### POST /auth:refresh

**Description:** Obtain new access token using refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Errors:**
- `401 Unauthorized`: Invalid or expired refresh token
- `403 Forbidden`: Token revoked

---

#### POST /auth:logout

**Description:** Revoke refresh token and end session.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "message": "Successfully logged out"
}
```

---

#### GET /auth:whoami

**Description:** Get current authenticated user information.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "id": "01HJ8X1234567890ABCDEF",
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "permissions": ["*"],
  "last_login_at": "2026-01-31T12:00:00Z"
}
```

---

#### POST /auth:change-password

**Description:** Change own password.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "old_password": "currentPassword123!",
  "new_password": "newSecurePassword456!"
}
```

**Response (200 OK):**
```json
{
  "message": "Password changed successfully"
}
```

**Errors:**
- `400 Bad Request`: Password doesn't meet requirements
- `401 Unauthorized`: Old password incorrect

---

### User Management Endpoints

#### POST /users:create

**Description:** Create a new user (admin only).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Required Permission:** `users:create`

**Request:**
```json
{
  "username": "editor1",
  "email": "editor@example.com",
  "password": "temporaryPassword123!",
  "role": "editor",
  "must_change_password": true
}
```

**Response (201 Created):**
```json
{
  "message": "User created successfully",
  "user": {
    "id": "01HJ8Y9876543210NEWUSER",
    "username": "editor1",
    "email": "editor@example.com",
    "role": "editor",
    "must_change_password": true,
    "created_at": "2026-01-31T12:00:00Z"
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid input (username taken, weak password)
- `403 Forbidden`: Insufficient permissions

---

#### GET /users:list

**Description:** List all users (admin only).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Required Permission:** `users:read`

**Query Parameters:**
- `role` (optional): Filter by role
- `limit` (optional): Results per page (default: 50)
- `after` (optional): Cursor for pagination

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": "01HJ8X1234567890ABCDEF",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "created_at": "2026-01-01T00:00:00Z",
      "last_login_at": "2026-01-31T12:00:00Z"
    }
  ],
  "next_cursor": "01HJ8Y9876543210CURSOR"
}
```

---

#### POST /users:update

**Description:** Update user (admin or self).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Required Permission:** `users:update` (or updating self)

**Request:**
```json
{
  "id": "01HJ8Y9876543210NEWUSER",
  "email": "newemail@example.com",
  "role": "admin"
}
```

**Response (200 OK):**
```json
{
  "message": "User updated successfully",
  "user": {
    "id": "01HJ8Y9876543210NEWUSER",
    "username": "editor1",
    "email": "newemail@example.com",
    "role": "admin",
    "updated_at": "2026-01-31T12:00:00Z"
  }
}
```

---

#### POST /users:delete

**Description:** Delete user (admin only, soft delete).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Required Permission:** `users:delete`

**Request:**
```json
{
  "id": "01HJ8Y9876543210NEWUSER"
}
```

**Response (200 OK):**
```json
{
  "message": "User deleted successfully"
}
```

**Errors:**
- `400 Bad Request`: Cannot delete root user or last admin
- `403 Forbidden`: Insufficient permissions

---

### API Key Management Endpoints

#### POST /apikeys:create

**Description:** Create a new API key (admin only).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Required Permission:** `apikeys:create`

**Request:**
```json
{
  "name": "Production Sync Service",
  "permissions": ["data:create", "data:read", "data:update"],
  "expires_at": "2027-01-31T00:00:00Z"
}
```

**Response (201 Created):**
```json
{
  "message": "API key created successfully",
  "api_key": {
    "id": "01HJ8Z1111111111APIKEY",
    "name": "Production Sync Service",
    "key": "moon_prod_a7b9c3d2e1f4g5h6_k8j9",
    "key_prefix": "moon_prod_a7",
    "permissions": ["data:create", "data:read", "data:update"],
    "expires_at": "2027-01-31T00:00:00Z",
    "created_at": "2026-01-31T12:00:00Z"
  },
  "warning": "Save this key now. You won't be able to see it again."
}
```

**Important:** Full key is only returned once. Client must save it securely.

---

#### GET /apikeys:list

**Description:** List API keys (admin sees all, users see own).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Required Permission:** `apikeys:read`

**Query Parameters:**
- `user_id` (optional, admin only): Filter by creator
- `limit` (optional): Results per page (default: 50)

**Response (200 OK):**
```json
{
  "api_keys": [
    {
      "id": "01HJ8Z1111111111APIKEY",
      "name": "Production Sync Service",
      "key_prefix": "moon_prod_a7",
      "permissions": ["data:create", "data:read", "data:update"],
      "expires_at": "2027-01-31T00:00:00Z",
      "created_at": "2026-01-31T12:00:00Z",
      "last_used_at": "2026-01-31T18:45:12Z",
      "revoked_at": null
    }
  ]
}
```

**Note:** Full key is never returned in list responses.

---

#### POST /apikeys:rotate

**Description:** Rotate API key (generate new, revoke old).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Required Permission:** `apikeys:create` (or key owner)

**Request:**
```json
{
  "id": "01HJ8Z1111111111APIKEY"
}
```

**Response (200 OK):**
```json
{
  "message": "API key rotated successfully",
  "api_key": {
    "id": "01HJ8Z2222222222NEWKEY",
    "name": "Production Sync Service",
    "key": "moon_prod_x9y8z7w6v5u4t3s2_m1n2",
    "key_prefix": "moon_prod_x9",
    "permissions": ["data:create", "data:read", "data:update"],
    "expires_at": "2027-01-31T00:00:00Z",
    "created_at": "2026-01-31T18:00:00Z"
  },
  "old_key_grace_period": "24 hours",
  "warning": "Save this key now. Old key will be revoked after grace period."
}
```

---

#### POST /apikeys:revoke

**Description:** Revoke API key immediately.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Required Permission:** `apikeys:revoke` (or key owner)

**Request:**
```json
{
  "id": "01HJ8Z1111111111APIKEY"
}
```

**Response (200 OK):**
```json
{
  "message": "API key revoked successfully"
}
```

---

### Permission and Role Endpoints

#### GET /permissions:list

**Description:** List all available permissions.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "permissions": [
    {
      "name": "collections:*",
      "description": "All collection operations"
    },
    {
      "name": "collections:create",
      "description": "Create collections"
    },
    {
      "name": "data:read",
      "description": "Read data records"
    }
  ]
}
```

---

#### GET /roles:list

**Description:** List all roles and their permissions.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "roles": [
    {
      "name": "admin",
      "description": "Administrator with full access",
      "permissions": ["collections:*", "data:*", "users:*", "apikeys:*"]
    },
    {
      "name": "editor",
      "description": "Can create and modify data",
      "permissions": ["collections:*", "data:*"]
    },
    {
      "name": "viewer",
      "description": "Read-only access",
      "permissions": ["collections:read", "data:read"]
    }
  ]
}
```

---

## 11. Database Schema

### Users Table

```sql
CREATE TABLE users (
  id TEXT PRIMARY KEY,                  -- ULID
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE,
  password_hash TEXT NOT NULL,          -- bcrypt hash
  role TEXT NOT NULL DEFAULT 'viewer',  -- root, admin, editor, viewer, api
  must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
  locked_until TIMESTAMP,               -- Account lockout
  failed_login_attempts INTEGER NOT NULL DEFAULT 0,
  last_failed_login_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_login_at TIMESTAMP,
  deleted_at TIMESTAMP                  -- Soft delete
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
```

### API Keys Table

```sql
CREATE TABLE api_keys (
  id TEXT PRIMARY KEY,                  -- ULID
  name TEXT NOT NULL,
  key_hash TEXT UNIQUE NOT NULL,        -- SHA-256 hash
  key_prefix TEXT NOT NULL,             -- First 12 chars (for logs)
  permissions TEXT NOT NULL,            -- JSON array of permission strings
  expires_at TIMESTAMP,
  created_by TEXT NOT NULL,             -- User ID (foreign key)
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP,
  revoked_at TIMESTAMP,                 -- Soft delete
  replaced_by TEXT,                     -- ID of replacement key (rotation)
  
  FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_created_by ON api_keys(created_by);
CREATE INDEX idx_api_keys_revoked_at ON api_keys(revoked_at);
```

### Refresh Tokens Table

```sql
CREATE TABLE refresh_tokens (
  id TEXT PRIMARY KEY,                  -- ULID (jti claim)
  token_hash TEXT UNIQUE NOT NULL,      -- SHA-256 hash
  user_id TEXT NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_at TIMESTAMP,
  replaced_by TEXT,                     -- ID of replacement token (rotation)
  
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

### Auth Audit Log Table

```sql
CREATE TABLE auth_audit_log (
  id TEXT PRIMARY KEY,                  -- ULID
  event_type TEXT NOT NULL,             -- login, logout, refresh, create_user, etc.
  user_id TEXT,
  username TEXT,
  api_key_id TEXT,
  ip_address TEXT,
  user_agent TEXT,
  success BOOLEAN NOT NULL,
  error_message TEXT,
  metadata TEXT,                        -- JSON additional data
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auth_audit_log_user_id ON auth_audit_log(user_id);
CREATE INDEX idx_auth_audit_log_event_type ON auth_audit_log(event_type);
CREATE INDEX idx_auth_audit_log_created_at ON auth_audit_log(created_at);
CREATE INDEX idx_auth_audit_log_success ON auth_audit_log(success);
```

---

## 12. Configuration

### Authentication Configuration Block

Add to existing `moon.conf` YAML:

```yaml
auth:
  enabled: true                          # Default: false (phase in carefully)
  jwt:
    secret: ""                           # REQUIRED if enabled - must be 256+ bits
    access_token_expiry: 900             # Default: 900 seconds (15 minutes)
    refresh_token_expiry: 604800         # Default: 604800 seconds (7 days)
    issuer: "moon-api"                   # Default: moon-api
    audience: "moon-api"                 # Default: moon-api
  
  apikey:
    enabled: true                        # Default: true
    header: "X-API-Key"                  # Default: X-API-Key
    prefix: "moon"                       # Default: moon
  
  password:
    min_length: 8                        # Default: 8
    require_uppercase: true              # Default: true
    require_lowercase: true              # Default: true
    require_number: true                 # Default: true
    require_special: true                # Default: true
    bcrypt_cost: 12                      # Default: 12 (2024 recommended)
  
  account_lockout:
    enabled: true                        # Default: true
    max_attempts: 5                      # Default: 5
    window_seconds: 900                  # Default: 900 (15 minutes)
    lockout_duration: 1800               # Default: 1800 (30 minutes)
  
  rate_limiting:
    enabled: true                        # Default: true
    login_requests_per_minute: 10        # Default: 10 per IP
    api_requests_per_minute: 60          # Default: 60 per user/key
  
  bootstrap:
    root_password_env: "MOON_ROOT_PASSWORD"  # Default: MOON_ROOT_PASSWORD
    force_password_change: true          # Default: true

  public_endpoints:                      # Endpoints that skip auth
    - "GET /health"
    - "GET /doc/*"
    - "POST /auth:login"
    - "POST /auth:refresh"
```

### Environment Variables

```bash
# Required for production
MOON_ROOT_PASSWORD=superSecurePassword123!

# Optional JWT secret (if not in config file)
MOON_JWT_SECRET=32-byte-base64-encoded-secret-here
```

### Configuration Validation

On startup, server validates:
- If `auth.enabled = true`, `jwt.secret` must be set and >= 32 bytes
- If `auth.enabled = false`, warn user that API is unprotected
- If root user doesn't exist and `MOON_ROOT_PASSWORD` not set, generate and log password
- Validate bcrypt cost is between 10-14 (performance vs security)
- Ensure at least one public endpoint exists (`/health`)

---

## 13. Implementation Phases

### Phase 1: Foundation (Week 1-2)

**Goals:**
- Database schema and migrations
- JWT token generation and validation
- Basic authentication middleware

**Deliverables:**
- [ ] Create users, api_keys, refresh_tokens tables
- [ ] Implement bcrypt password hashing
- [ ] Implement JWT signing and verification
- [ ] Create authentication middleware (extract token, validate, load user)
- [ ] Bootstrap root user on startup
- [ ] Unit tests for core crypto functions

**Risks:**
- JWT secret management complexity
- Migration path for existing deployments

---

### Phase 2: Authentication Endpoints (Week 3)

**Goals:**
- Login, refresh, logout endpoints
- User management endpoints
- API key creation

**Deliverables:**
- [ ] `POST /auth:login` with password validation and token generation
- [ ] `POST /auth:refresh` with token rotation
- [ ] `POST /auth:logout` with refresh token revocation
- [ ] `GET /auth:whoami`
- [ ] `POST /auth:change-password`
- [ ] `POST /users:create`, `GET /users:list`, `POST /users:update`, `POST /users:delete`
- [ ] `POST /apikeys:create`, `GET /apikeys:list`
- [ ] Integration tests for auth flows

**Risks:**
- Token rotation complexity
- User enumeration attacks

---

### Phase 3: Authorization (Week 4)

**Goals:**
- Role-based access control
- Permission checking middleware
- Protect existing endpoints

**Deliverables:**
- [ ] Permission model implementation
- [ ] Authorization middleware (check permissions)
- [ ] Protect all existing endpoints with required permissions
- [ ] API key permission validation
- [ ] `GET /permissions:list`, `GET /roles:list`
- [ ] E2E tests for authorization scenarios

**Risks:**
- Breaking existing API clients
- Performance impact of permission checks

---

### Phase 4: Security Hardening (Week 5)

**Goals:**
- Rate limiting
- Account lockout
- Audit logging
- API key rotation

**Deliverables:**
- [ ] Rate limiting middleware (per-IP for auth, per-user for API)
- [ ] Account lockout logic
- [ ] Auth audit log table and logging
- [ ] `POST /apikeys:rotate` with grace period
- [ ] `POST /apikeys:revoke`
- [ ] Security headers middleware
- [ ] Penetration testing

**Risks:**
- False positive lockouts
- Performance impact of audit logging

---

### Phase 5: Production Readiness (Week 6)

**Goals:**
- Documentation
- Monitoring and alerting
- Migration tools
- Performance testing

**Deliverables:**
- [ ] Update API documentation (`/doc/md` and `/doc/`)
- [ ] Add authentication examples to docs
- [ ] Migration guide for existing deployments
- [ ] Monitoring dashboard for auth events
- [ ] Alerts for security events (mass failures, lockouts)
- [ ] Load testing (target: <1ms auth overhead)
- [ ] Go-live checklist

**Risks:**
- Documentation gaps
- Production issues not caught in testing

---

## 14. Testing Strategy

### Unit Tests

**Coverage Target:** 90%+

**Key Test Areas:**
- Password hashing and verification (bcrypt)
- JWT token generation and validation
- API key generation and validation
- Permission checking logic
- Token rotation logic
- Account lockout logic

**Example Test Cases:**
```go
TestPasswordHashing()
TestJWTSigningAndVerification()
TestJWTExpiration()
TestJWTInvalidSignature()
TestAPIKeyGeneration()
TestAPIKeyValidation()
TestPermissionCheck()
TestWildcardPermission()
TestAccountLockout()
TestAccountLockoutReset()
```

---

### Integration Tests

**Key Scenarios:**
1. **Login Flow:**
   - Valid credentials → access and refresh tokens
   - Invalid credentials → 401 Unauthorized
   - Locked account → 403 Forbidden
   - Must change password → special response

2. **Refresh Flow:**
   - Valid refresh token → new tokens, old token revoked
   - Invalid refresh token → 401 Unauthorized
   - Already-used token (replay) → 401 Unauthorized

3. **Authorization:**
   - Admin creates collection → 201 Created
   - Viewer creates collection → 403 Forbidden
   - Editor reads data → 200 OK
   - Unauthenticated request to protected endpoint → 401 Unauthorized

4. **API Key:**
   - Valid key → access granted
   - Revoked key → 401 Unauthorized
   - Expired key → 401 Unauthorized
   - Key with insufficient permissions → 403 Forbidden

5. **User Management:**
   - Admin creates user → 201 Created
   - Non-admin creates user → 403 Forbidden
   - User updates own profile → 200 OK
   - User changes own password → 200 OK

---

### Security Tests

**Penetration Testing:**
- SQL injection in auth endpoints
- JWT algorithm confusion attack
- Token replay attacks
- Brute force password guessing
- Account enumeration via timing
- CORS misconfiguration
- XSS in error messages

**Automated Security Scanning:**
- OWASP ZAP or Burp Suite
- Static analysis (gosec for Go)
- Dependency vulnerability scanning (Dependabot)

---

### Performance Tests

**Load Testing:**
- Baseline: Measure request latency without auth
- With Auth: Measure latency with JWT validation
- Target: <1ms overhead for token validation
- Target: <10ms for API key validation (with caching)
- Stress: 1000 concurrent login attempts
- Stress: 10,000 concurrent authenticated requests

**Tools:**
- `wrk` or `hey` for HTTP load testing
- Custom Go benchmarks for crypto operations

---

## 15. Monitoring and Auditing

### Metrics to Track

**Authentication Metrics:**
- Login attempts (success/failure) per minute
- Refresh token usage per minute
- API key usage per minute
- Account lockouts per hour
- Password changes per day

**Performance Metrics:**
- Token validation latency (p50, p95, p99)
- API key validation latency (p50, p95, p99)
- Auth endpoint response times

**Security Metrics:**
- Failed login attempts by IP
- Failed login attempts by username
- Revoked tokens/keys per day
- Admin actions per day

### Audit Log Events

**Critical Events (Always Log):**
- User login (success/failure, IP, user agent)
- User logout
- Password change
- User created/updated/deleted
- API key created/rotated/revoked
- Permission changes
- Account lockout
- Token refresh
- Failed authorization (permission denied)

**Audit Log Format:**
```json
{
  "id": "01HJ8Z3333333333AUDIT",
  "timestamp": "2026-01-31T12:00:00Z",
  "event_type": "login",
  "user_id": "01HJ8X1234567890USER",
  "username": "admin",
  "ip_address": "203.0.113.42",
  "user_agent": "Mozilla/5.0...",
  "success": true,
  "metadata": {
    "login_method": "password"
  }
}
```

### Alerts

**High-Priority Alerts:**
- 🚨 10+ failed logins from single IP in 1 minute
- 🚨 Root user login from new location
- 🚨 Mass account lockouts (5+ in 5 minutes)
- 🚨 API key with elevated permissions created
- 🚨 User role changed to admin

**Medium-Priority Alerts:**
- ⚠️ API key not rotated in 90+ days
- ⚠️ User password not changed in 180+ days
- ⚠️ Failed authorization spike (100+ in 1 hour)

### Dashboards

**Real-time Dashboard:**
- Active users (logged in last 5 minutes)
- Login success rate (last hour)
- Top failed login IPs
- Recent audit events (last 100)
- API key usage by prefix

**Security Dashboard:**
- Locked accounts (active)
- Password reset requests (last 24 hours)
- Admin actions (last 7 days)
- API keys by permission level
- Failed authorization attempts by endpoint

---

## 16. Migration and Backwards Compatibility

### Enabling Authentication on Existing Deployment

**Challenge:**
- Existing Moon deployments have no authentication
- Cannot break existing API clients immediately
- Need graceful migration path

**Solution: Phased Rollout**

#### Phase A: Authentication Optional (Deploy)

1. Deploy auth system with `auth.enabled = false`
2. Endpoints work without authentication (existing behavior)
3. Auth endpoints available but optional
4. Clients can start using authentication if desired

**Configuration:**
```yaml
auth:
  enabled: false  # Auth available but not enforced
  jwt:
    secret: "your-secret-here"
```

#### Phase B: Warning Period (Week 1-2)

1. Set `auth.enabled = true` but add grace period
2. Unauthenticated requests succeed with warning header
3. Log all unauthenticated requests
4. Send deprecation warnings to clients

**Response Headers:**
```
X-Auth-Warning: Authentication will be required after 2026-02-15
X-Auth-Docs: https://moon.asensar.in/doc/md#authentication
```

**Configuration:**
```yaml
auth:
  enabled: true
  grace_period_until: "2026-02-15T00:00:00Z"
```

#### Phase C: Enforcement (After Grace Period)

1. Remove grace period
2. All protected endpoints require authentication
3. Unauthenticated requests return 401 Unauthorized

**Configuration:**
```yaml
auth:
  enabled: true
  # grace_period_until removed
```

### Migration Tools

**Admin Bootstrap Script:**
```bash
#!/bin/bash
# bootstrap-auth.sh - Helper script for initial auth setup

# 1. Generate secure root password
ROOT_PASSWORD=$(openssl rand -base64 32)

# 2. Set environment variable
export MOON_ROOT_PASSWORD="$ROOT_PASSWORD"

# 3. Start Moon with auth enabled
./moon --config /etc/moon.conf

# 4. Save root password securely
echo "Root password: $ROOT_PASSWORD" | gpg --encrypt --recipient admin@example.com > root-password.gpg

# 5. Create admin API key for migration
curl -X POST http://localhost:6006/auth:login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"root\",\"password\":\"$ROOT_PASSWORD\"}" \
  | jq -r '.access_token' > /tmp/access_token

curl -X POST http://localhost:6006/apikeys:create \
  -H "Authorization: Bearer $(cat /tmp/access_token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Migration Key",
    "permissions": ["*"],
    "expires_at": "2026-03-01T00:00:00Z"
  }' | jq -r '.api_key.key' > migration-key.txt

echo "Migration API key saved to migration-key.txt"
echo "Update your clients with this key before grace period ends."
```

### Client Migration Guide

**For Existing Clients:**

1. **Obtain API Key:**
   - Contact admin to create API key with required permissions
   - Or create user account and use JWT tokens

2. **Update API Calls:**
   ```bash
   # Before (no auth)
   curl http://localhost:6006/collections:list
   
   # After (API key)
   curl http://localhost:6006/collections:list \
     -H "X-API-Key: moon_prod_a7b9c3d2e1f4g5h6_k8j9"
   
   # After (JWT token)
   curl http://localhost:6006/collections:list \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
   ```

3. **Test Before Grace Period Ends:**
   - Verify all API calls work with authentication
   - Check for 401/403 errors
   - Update error handling for auth failures

### Rollback Plan

If critical issues arise:

1. **Immediate:**
   - Set `auth.enabled = false` in config
   - Restart server
   - All endpoints become public again

2. **Investigate:**
   - Review audit logs for issues
   - Check error rates and latency
   - Gather client feedback

3. **Fix and Re-enable:**
   - Address issues
   - Deploy fix
   - Re-enable with extended grace period

---

## Appendix A: Security Checklist

### Pre-Deployment

- [ ] JWT secret is 256+ bits and cryptographically random
- [ ] JWT secret stored in environment variable or secret manager (not config file)
- [ ] bcrypt cost factor set to 12 (or higher for sensitive applications)
- [ ] HTTPS enforced in production (HSTS header)
- [ ] Rate limiting enabled for auth endpoints
- [ ] Account lockout enabled and tested
- [ ] Password requirements enforced and tested
- [ ] Root password set via environment variable (not logged)
- [ ] Public endpoints whitelist reviewed and minimal
- [ ] All sensitive endpoints protected with permissions
- [ ] Audit logging enabled and tested
- [ ] Security headers configured (CSP, X-Frame-Options, etc.)
- [ ] CORS policy configured and restrictive
- [ ] Error messages don't leak sensitive information
- [ ] SQL injection tests passed
- [ ] XSS tests passed
- [ ] JWT algorithm confusion tests passed

### Post-Deployment

- [ ] Monitor failed login attempts
- [ ] Monitor account lockouts
- [ ] Monitor API key usage
- [ ] Review audit logs daily (first week)
- [ ] Set up alerts for security events
- [ ] Test disaster recovery (revoke all keys, force password resets)
- [ ] Rotate JWT secret after 90 days
- [ ] Review and prune unused API keys monthly
- [ ] Review and deactivate inactive users quarterly

---

## Appendix B: Common Attack Vectors and Mitigations

| Attack Vector                  | Mitigation                                                     |
|--------------------------------|----------------------------------------------------------------|
| Brute force password guessing  | Rate limiting (10 req/min), account lockout (5 attempts)       |
| Credential stuffing            | Rate limiting, CAPTCHA (future), anomaly detection             |
| SQL injection                  | Parameterized queries, ORM usage                               |
| XSS in error messages          | Sanitize all output, CSP headers                               |
| JWT algorithm confusion        | Validate `alg` header, reject `none`                           |
| JWT secret brute force         | 256-bit secret, rotate regularly                               |
| Token replay                   | Short expiration, refresh token rotation                       |
| Session fixation               | Generate new token on login, invalidate old sessions           |
| Account enumeration            | Same response time for valid/invalid users, generic errors     |
| Timing attacks                 | Constant-time password comparison (bcrypt handles this)        |
| CSRF                           | SameSite cookies, CSRF tokens (if using cookies)               |
| Man-in-the-middle              | HTTPS only, HSTS, secure cookies                               |
| Clickjacking                   | X-Frame-Options: DENY                                          |
| API key leakage in logs        | Only log key prefix, sanitize logs                             |
| Privilege escalation           | Strict permission checks, audit role changes                   |
| Mass assignment                | Explicit field whitelisting in update endpoints                |

---

## Appendix C: References and Further Reading

### Industry Standards

- **OWASP Top 10:** https://owasp.org/www-project-top-ten/
- **OWASP Authentication Cheat Sheet:** https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- **OWASP JWT Cheat Sheet:** https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html
- **NIST Password Guidelines:** https://pages.nist.gov/800-63-3/sp800-63b.html

### Best Practices Articles

- JWT Authentication Security Guide: Refresh Token Rotation and Replay Protection
- Building Secure REST APIs with JWT Authentication: Complete 2025 Guide
- API Key Management Best Practices: Keeping Your APIs Secure and Efficient
- Role-Based Access Control for APIs: Implementation Guide

### Tools and Libraries

- **Go JWT Library:** github.com/golang-jwt/jwt/v5
- **Go bcrypt:** golang.org/x/crypto/bcrypt
- **Security Headers Testing:** https://securityheaders.com
- **JWT Debugger:** https://jwt.io

---

**End of Document**

---

**Document Metadata:**
- **Version:** 1.0
- **Date:** 2026-01-31
- **Author:** Moon Development Team
- **Status:** Ready for Review
- **Next Steps:** Review by team, approval, begin Phase 1 implementation

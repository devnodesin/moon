## SYSTEM DB TABLES

- Tables for `users`, `refresh_token`, and `apikeys` should be prefixed with `moon_`.
- These tables are internal only and should never be accessible/visible on collections endpoints.
- They should only be accessible with special endpoints: `Auth Endpoints`, `User Management Endpoints`, and `API Key Management Endpoints`.
- Refer to the doc template `cmd\moon\internal\handlers\templates\doc.md.tmpl`.
- They should not be included in the count.
- If this is not properly defined in ``SPEC.md``, `SPEC_AUTH.md`, then update it properly in the SPEC files.

```bash
$ curl -s http://localhost:6006/collections:list \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
{
  "collections": [
    "refresh_tokens",
    "users",
    "apikeys",
    "products"
  ],
  "count": 4
}
```

## After logout I am able to fetch the protected endpoints

```bash
$ curl -X POST "http://localhost:6006/auth:logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"$REFRESH_TOKEN"}' | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    72  100    38  100    34  17257  15440 --:--:-- --:--:-- --:--:-- 36000
{
  "message": "logged out successfully"
}
mohamed@asensar-ubuntu-s-2vcpu-4gb-amd-blr1-01:~/docker/apps/moon$ curl -s "http://localhost:6006/collections:list" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
{
  "collections": [
    "users",
    "refresh_tokens",
    "apikeys"
  ],
  "count": 3
}

```

## After changing the password the user should be logged out, but I am still able to access the API

```bash
$ curl -X POST "http://localhost:6006/auth:me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"old_password":"moonadmin12#","password":"NewSecurePass456"}' | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   226  100   165  100    61    315    116 --:--:-- --:--:-- --:--:--   432
{
  "message": "user updated successfully",
  "user": {
    "id": "01KGD682TEA2MRAF0NXA07BV41",
    "username": "admin",
    "email": "newemail@example.com",
    "role": "admin",
    "can_write": true
  }
}
mohamed@asensar-ubuntu-s-2vcpu-4gb-amd-blr1-01:~/docker/apps/moon$ curl -X GET "http://localhost:6006/auth:me"   -H "Authorization: Bearer $ACCESS_TOKEN" | jq .       % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   127  100   127    0     0  57104      0 --:--:-- --:--:-- --:--:-- 63500
{
  "user": {
    "id": "01KGD682TEA2MRAF0NXA07BV41",
    "username": "admin",
    "email": "newemail@example.com",
    "role": "admin",
    "can_write": true
  }
}
mohamed@asensar-ubuntu-s-2vcpu-4gb-amd-blr1-01:~/docker/apps/moon$ curl -X GET "http://localhost:6006/auth:me"   -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   127  100   127    0     0  64895      0 --:--:-- --:--:-- --:--:--  124k
{
  "user": {
    "id": "01KGD682TEA2MRAF0NXA07BV41",
    "username": "admin",
    "email": "newemail@example.com",
    "role": "admin",
    "can_write": true
  }
}
```

## CORS is not configured but I am able to access it from outside

```bash
$ curl  http://localhost:6006/health
{"name":"moon","status":"live","version":"1.99"}

$ curl   https://moon.asensar.in/health
{"name":"moon","status":"live","version":"1.99"}
mohamed@asensar-ubuntu-s-2vcpu-4gb-amd-blr1-01:~/docker/apps/moon$ cat ./moon.conf
# Moon - Dynamic Headless Engine Configuration
# Minimal quick-start configuration file
# Copy this file to /etc/moon.conf and adjust values as needed

server:
  host: "0.0.0.0"
  port: 6006
  prefix: ""  # Optional URL prefix (e.g., "/api/v1", "/moon/api")

database:
  connection: "sqlite"
  database: "/opt/moon/sqlite.db"

logging:
  path: "/var/log/moon"

# ============================================================================
# JWT Authentication Configuration (REQUIRED)
# ============================================================================
jwt:
  # JWT signing secret, Use a unique, random string of at least 32 characters
  # Generate a secure secret with: openssl rand -base64 32
  secret: "your-secret-key-change-in-production"
  access_expiry: 3600 # Default: 3600 (1 hour)
  refresh_expiry: 604800 # Default: 604800 (7 days)

# ============================================================================
# API Key Authentication Configuration (Optional)
# ============================================================================
apikey:
  enabled: true # Enable API key authentication
  header: "X-API-Key" # Clients must send their API key in this header

# On first startup, if no admin users exist, Moon will create an admin
# user with these credentials. You MUST change the password immediately
# after first login.
auth:
  bootstrap_admin:
    username: "admin"
    email: "admin@example.com"
    password: "moonadmin12#" # CHANGE THIS PASSWORD IMMEDIATELY AFTER FIRST LOGIN!

# ============================================================================
# Rate Limiting Configuration (Optional)
# ============================================================================
# Rate limits protect against abuse and ensure fair resource usage
# ratelimit:
#   user_rpm: 100  # Requests per minute for JWT-authenticated users (Default: 100)
#   apikey_rpm: 1000 # Requests per minute for API key-authenticated requests,  Default: 1000 (higher for machine-to-machine)
#   login_attempts: 5 # Maximum failed login attempts per 15 minutes per IP/username (Default: 5)
```

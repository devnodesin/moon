## Authentication Endpoints

- `POST /auth:login`: Login
- `POST /auth:logout`: Logout
- `POST /auth:refresh`: Refresh access token
- `GET /auth:me`: Get current user
- `POST /auth:me`: Update current user

### Login

`POST /auth:login`

Authenticate user and receive access token.

**Request body:**

```json
{
  "username": "newuser",
  "password": "UserPass123#"
}
```

**Response (200 OK):**

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "hyTTpweINXOKltH6r5Cl7--_8VKl58Z6fE7W0fjlHls=",
    "expires_at": "2026-02-14T03:27:33.935149435Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KHCZGWWRBQBREMG0K23C6C5H",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true
    }
  },
  "message": "Login successful"
}
```

### Get Current User

`GET /auth:me`

Retrieve authenticated user information.

**Headers:**

- `Authorization: Bearer {access_token}` (required)

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  }
}
```

### Update Current User

`POST /auth:me`

Update authenticated user's email or password.

**Headers:**

- `Authorization: Bearer {access_token}` (required)

**Update email:**

```json
{
  "email": "newemail@example.com"
}
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  },
  "message": "User updated successfully"
}
```

**Change password:**

```json
{
  "old_password": "UserPass123#",
  "password": "NewSecurePass456"
}
```

**Response (200 OK):**

```json
{
  "data": {
    "id": "01KHCZGWWRBQBREMG0K23C6C5H",
    "username": "newuser",
    "email": "newemail@example.com",
    "role": "user",
    "can_write": true
  },
  "message": "Password updated successfully. Please login again."
}
```

### Refresh Token

`POST /auth:refresh`

Generate new access token using refresh token.

**Request body:**

```json
{
  "refresh_token": "hyTTpweINXOKltH6r5Cl7--_8VKl58Z6fE7W0fjlHls="
}
```

**Response (200 OK):**

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "Yke6FxWxoqPfagJCfD13Rbb8SZz_4SMG9TuI_a61YEE=",
    "expires_at": "2026-02-14T03:27:36.386965511Z",
    "token_type": "Bearer",
    "user": {
      "id": "01KHCZGWWRBQBREMG0K23C6C5H",
      "username": "newuser",
      "email": "newemail@example.com",
      "role": "user",
      "can_write": true
    }
  },
  "message": "Token refreshed successfully"
}
```

### Logout

`POST /auth:logout`

Invalidate current session and refresh token.

**Headers:**

- `Authorization: Bearer {access_token}` (required)

**Request body:**

```json
{
  "refresh_token": "hyTTpweINXOKltH6r5Cl7--_8VKl58Z6fE7W0fjlHls="
}
```

**Response (200 OK):**

```json
{
  "message": "Logged out successfully"
}
```

### Important Notes

- **Token expiration**: Access tokens expire in 1 hour (configurable). Use refresh token to obtain new access token without re-authentication.
- **Refresh token**: Single-use tokens. Each refresh returns a new access token AND a new refresh token. Store the new refresh token for subsequent refreshes.
- **Password change**: Changing password invalidates all existing sessions. User must login again with new credentials.
- **Authorization header**: Format is `Authorization: Bearer {access_token}`. Include this header in all authenticated requests.
- **Token storage**: Store tokens securely. Never expose tokens in URLs or logs.

### Error Handling

**Error Response:** Follow [Standard Error Response](#standard-error-response) for any error handling
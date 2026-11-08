# Smart Login Client Guide

## Flow

1. Initiate: `POST /auth:session` with `{ "op": "smart_login", "data": { "email": "user@example.com" } }`
   - Response: `{ "challenge": "credentials", "message": "Enter your password or request an OTP" }`
2. To request an OTP: `POST /auth:session` with `{ "op": "smart_login", "data": { "email": "user@example.com", "request_otp": true } }`
   - Response: `{ "challenge": "credentials", "message": "Enter your password or request an OTP" }`
   - OTP email is sent if the account uses OTP login
3. Verify password: `POST /auth:session` with `{ "op": "smart_login", "data": { "email": "...", "password": "..." } }`
4. Verify OTP: `POST /auth:session` with `{ "op": "smart_login", "data": { "email": "...", "code": "123456" } }`

## Errors

- 400 missing/invalid fields
- 401 invalid credentials / OTP
- 429 rate limited (5 requests / 15 min per email)
- 503 OTP not configured

## Migration

Replace `login`, `otp_request`, `otp_verify` ops with `smart_login`.

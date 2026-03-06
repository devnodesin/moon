# Moon Specification Review

## Summary

`SPEC.md` is now substantially complete and no longer empty. It now documents architecture, schema lifecycle, configuration, security, and operational boundaries aligned to the frozen API documents.  
This review now keeps only issues that remain outstanding across `SPEC_API.md` and `SPEC/*.md`.

---

## Critical Contract Inconsistencies

### 4. `405` Missing from Allowed Error Codes

`SPEC_API.md` says unsupported methods return `405`, but `SPEC/10_error.md` says only listed codes are allowed and does not include `405`.

### 5. `/collections:query` Example Missing `message`

`SPEC_API.md` says `message` is always present, but `SPEC/30_collection.md` list response example omits it.

### 6. Action Mutation Responses Missing `meta`

Mutation response shape includes `meta.success` / `meta.failed`, but `reset_password` and `revoke_sessions` responses in `SPEC/40_resource.md` omit `meta`.

---

## Missing Documentation

### 7. `GET /auth:me` Has No Request/Response Contract

Endpoint is listed in `SPEC_API.md` but lacks full request/response details in specs.

### 8. `POST /auth:me` Has No Request/Response Contract

Endpoint is listed in `SPEC_API.md` but lacks payload rules, validation behavior, and response examples.

### 9. API Key Lifecycle Is Incomplete

`SPEC/40_resource.md` documents key rotation only. Missing docs for:

- create API key (`op=create`)
- list/query API keys
- destroy API keys (`op=destroy`)
- canonical API key object fields
- explicit statement that raw key is only returned on create/rotate

### 10. User Management Lifecycle Is Incomplete

Missing docs for user create/list/get-one patterns and canonical writable/system-managed fields.

### 11. Partial Batch Failure Shape Is Undefined

The spec allows partial success (`meta.success`, `meta.failed`) but does not define how failed items are represented in response `data`.

### 12. Rate Limit Response Semantics Are Undefined

`429` is listed but behavior is unspecified (`Retry-After`, reset window semantics, client backoff guidance).

### 13. `can_write` Semantics Are Not Defined

`can_write` appears in auth responses/JWT examples but no spec defines exact meaning, source of truth, or interaction with role-based authorization.

---

## Additional Findings and Suggestions

- **`revoke_sessions` message wording:** current `"Revoke session successful"` is inconsistent with surrounding response language; prefer `"Sessions revoked successfully"`.
- **Filter syntax ambiguity:** `?filter=quantity[gt]=5&brand[eq]=Wow` mixes patterns; define one canonical filter encoding.
- **Top-level resource route scope:** `SPEC/40_resource.md` allows `/products:query` in addition to `/data/products:query`; clarify in `SPEC_API.md` whether this applies to all resources including system collections.
- **`rotate` response `name` field origin:** rotation request only sends `id`, but response includes `name`; document whether this is always the persisted existing name.
- **`/auth:me` token type support:** clarify whether endpoint accepts both JWT and API keys, or JWT only.


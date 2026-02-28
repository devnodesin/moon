# Gap Analysis: Code vs Specification

Analysis of gaps between the source code and `SPEC.md`, `SPEC_API.md`, `SPEC_AUTH.md`, and `SPEC_API/*.md`.

---

## Issues Fixed in This PR

| Item | Status |
|------|--------|
| Dead `errors` package removed (was unused by handlers) | ✅ Fixed |
| `readonly` role removed from code and tests | ✅ Fixed |
| `prev` cursor always null — backward pagination implemented | ✅ Fixed |
| Root `GET /` endpoint now aliases `/health` instead of plain text | ✅ Fixed |
| `google/uuid` third-party dep replaced with `crypto/rand` | ✅ Fixed |
| Middleware order documented with rationale in `SPEC_AUTH.md` | ✅ Fixed |
| Logout blacklisting both tokens documented in `SPEC_AUTH.md` | ✅ Fixed |
| Root endpoint documented in `SPEC_API.md` and `SPEC_API/010-health.md` | ✅ Fixed |
| Rate limit HTTP status code fixed from `400` to `429 Too Many Requests` | ✅ Fixed |

---

## Remaining Gaps

### Code features not in spec

#### 1. `X-Request-ID` Header

- **Code:** The `logging/logger.go` middleware generates and sets an `X-Request-ID` header on every response. If the client provides `X-Request-ID` in the request, it is echoed back.
- **Spec:** Listed as an exposed CORS header but not described in detail anywhere.
- **Assessment:** Useful for distributed tracing; low risk. The CORS spec lists it as exposed. No action needed beyond the CORS header declaration.

#### 2. `total` Field in List Response (Spec only mentions `count`)

- **Code:** Some list endpoints (e.g., data `:create`, `:destroy`) return `meta.total`. The `:list` response does not include `total`.
- **Spec (`SPEC_API.md`):** `count`, `limit`, `next`, `prev` are the four meta fields for `:list`. `total` is mentioned in `:create` and `:destroy` responses as the number of items attempted.
- **Assessment:** Compliant. The `total` in create/destroy is separate from list pagination. No gap.

#### 3. In-Memory Login Rate Limiter

- **Code:** `middleware/loginratelimit.go` implements an in-memory IP+username rate limiter for login attempts.
- **Spec (`SPEC_AUTH.md`):** Specifies "5 attempts per 15 minutes per IP/username".
- **Assessment:** Compliant. Implemented and matching spec.

#### 4. `last_login_at` Not Returned by Default in `/auth:me`

- **Code:** The `/auth:me` response returns `UserInfo{id, username, email, role, can_write}` — no `last_login_at`.
- **Spec (`SPEC_API/030-users.md`):** The admin user list includes `last_login_at` but the `/auth:me` response in `SPEC_API/020-auth.md` does not include it.
- **Assessment:** Compliant. No gap.

#### 5. Collections List Does Not Use Cursor Pagination

- **Code:** `/collections:list` returns all collections with `next: null, prev: null`.
- **Spec:** No explicit pagination is defined for collections list — it's not a large dataset.
- **Assessment:** Acceptable. Collections are typically small (bounded by schema design). No gap.

#### 6. `doc` Endpoints (HTML, Markdown, JSON, Text)

- **Code:** Documentation endpoints at `/doc`, `/doc/llms.md`, `/doc/llms.txt`, `/doc/llms.json` are implemented.
- **Spec (`SPEC_API.md`):** Listed as public endpoints.
- **Assessment:** Compliant. No gap.

#### 7. `readonly` Field on Schema Columns

- **Code:** The `:schema` endpoint returns a `readonly` boolean flag for `id` and `pkid` columns.
- **Spec (`SPEC_API/060-data.md`):** Shows `"readonly": true` for the `id` field in schema response.
- **Assessment:** Compliant. No gap.

#### 8. `DuplicateKeyPatterns` and `ConnectionErrorPatterns` Constants

- **Code:** These constants in `constants/errors.go` are defined but only used in tests.
- **Spec:** Not mentioned.
- **Assessment:** Low-value dead code but harmless. No functional gap.

---

### Spec features not fully implemented

#### 1. `meta.count` in List Response vs Actual Count

- **Spec:** `count` should be the "number of records returned in this response."
- **Code:** `count` = `len(data)` after truncation to `limit`. This is correct.
- **Assessment:** Compliant.

#### 2. No `total` Count Across All Pages in `:list`

- **Code:** The data list computes a `total` (all records matching filters, regardless of cursor), but this value is not returned in the response meta.
- **Spec:** No `total` field is specified in `:list` meta — only `count`, `limit`, `next`, `prev`.
- **Assessment:** Compliant (total is computed internally but not exposed, matching spec).

#### 3. Collections `:update` — `RenameColumns` Not Verified at Schema Level

- **Code:** Rename operations are implemented but may not fully validate that old column names exist before renaming.
- **Spec:** Rename should fail gracefully with an error if the old column doesn't exist.
- **Assessment:** Minor gap; worth validating in integration tests.

---

## Summary

| Category | Count |
|----------|-------|
| Gaps fixed in this PR | 9 |
| Code features beyond spec (acceptable) | 8 |
| Spec features not implemented / mismatched | 1 |

### Action Items (Priority)

1. **LOW:** Document `X-Request-ID` header lifecycle in `SPEC_API.md` (currently only listed in CORS exposed headers).

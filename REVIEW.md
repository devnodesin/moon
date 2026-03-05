# Moon Spec Review

**Score: 6.5 / 10**

The spec has a solid conceptual foundation — intentional minimalism, AIP-136 action pattern, clear data model — but contains enough internal contradictions and gaps to cause real implementation confusion. Issues are grouped below by severity.

---

## Critical Inconsistencies

### 1. Login field: `email` vs `username`
- `SPEC_API.md` (20_auth top schema) states login requires `email` and `password`.
- `SPEC/20_auth.md` login request example sends `username` instead of `email`.
- **Pick one and use it consistently.**

.

### 4. `data` shape violation — schema endpoint
- `SPEC_API.md` core rule: "`data` is always an array of objects."
- `SPEC/40_resource.md` schema response returns `data` as a plain object `{}`, not an array.
- Either the rule or the example must change.

### 5. Pagination meta field names disagree
- `SPEC_API.md` list response meta: `per_page`, `current_page`
- `SPEC/30_collection.md` list response meta: `limit`, `page`
- Fields must match across all list endpoints.

---

## Minor Issues


### 10. Surprising default for `unique`
- `SPEC/30_collection.md`: "defaults to `nullable: false` and `unique: true`"
- Having `unique: true` as the default is an unusual and breaking default for most fields. Likely should be `unique: false`. Confirm intent.

### 11. Mutation response shape inconsistency
- `SPEC_API.md` mutation response shows `data: []` (empty array).
- `SPEC/40_resource.md` create/update response returns the full record(s) inside `data`.
- These are contradictory. Full record in response is the better UX — update the spec shape.

### 12. SPEC.md is almost empty
- `SPEC.md` is only 6 lines: a title, a one-liner, and a pointer to `SPEC_API.md`.
- Per `AGENTS.md`, SPEC.md should cover architecture, database design, schema management, and system behavior. It currently has none of that.

---

## Strengths

- **Clear intentional scope.** The "What Moon Does NOT Do" list is excellent — sets honest expectations.
- **AIP-136 pattern.** Colon-separated actions are consistent and AI-friendly.
- **Error model.** Simple, strict, human-readable — exactly right.
- **ULID identifiers.** Good choice; lexicographically sortable, collision-safe.
- **Unified auth header.** JWT and API keys both use `Authorization: Bearer` — clean.
- **Atomic schema operations.** One op per request avoids complex rollback logic.
- **Partial batch success.** `meta.success` / `meta.failed` is a pragmatic pattern.
- **Rate limit model.** Separate limits for JWT vs API key users is well-considered.

---

## Recommended Actions (Priority Order)

1. Fix login field: standardize to `email` or `username` across all files.
2. Fix `/collection` vs `/collections` — pick one, update all examples.
3. Fix schema response `data` to be an array.
4. Align pagination meta field names (`per_page`/`current_page`) everywhere.
5. Fix links in 40_resource.md to include `:query` suffix.
6. Correct `GET` → `POST` in the mutate section heading.
7. Clarify `text` vs `string` type.
8. Clarify `unique: true` default intent.
9. Align mutation response — empty array vs full record — and document the winner.
10. Flesh out `SPEC.md` with architecture and system behavior content.

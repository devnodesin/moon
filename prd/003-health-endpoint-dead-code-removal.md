## Overview

### Problem Statement
The codebase contains an unused `health.Service` implementation while the server uses a separate inline health handler. This creates dead code, duplicated health logic, and maintenance ambiguity.

### Context and Background
- The authoritative response contract for health is defined in [SPEC_API/010-health.md](../SPEC_API/010-health.md).
- This work must keep the health endpoint implementation lean and minimal.
- Backward compatibility with old/unused health paths is not required.

### High-Level Solution Summary
Standardize health behavior on a single minimal implementation for `GET /health` that strictly follows [SPEC_API/010-health.md](../SPEC_API/010-health.md), and remove unused/dead health code paths (including non-routed legacy service components) to eliminate duplication.

## Requirements

### Functional Requirements
1. `GET /health` MUST return the exact contract defined by [SPEC_API/010-health.md](../SPEC_API/010-health.md).
2. Health response contract MUST be treated as authoritative from [SPEC_API/010-health.md](../SPEC_API/010-health.md), not inferred from legacy code.
3. Health endpoint MUST remain publicly accessible (no auth), consistent with API specification.
4. Health endpoint behavior MUST be minimal and deterministic:
   - include only fields required by [SPEC_API/010-health.md](../SPEC_API/010-health.md),
   - avoid extra readiness/dependency payloads.

### Technical Requirements
1. Keep one active health implementation path only.
2. Remove dead/unwired health code that is not part of the selected minimal `GET /health` flow.
3. Remove unused types, helpers, and tests tied only to removed dead health path.
4. Do not introduce compatibility shims for removed health implementations.
5. Preserve existing server routing pattern while ensuring only the minimal health handler is registered.

### API Specifications
- Endpoint: `GET /health`
- Source of truth: [SPEC_API/010-health.md](../SPEC_API/010-health.md)
- Required response status: `200 OK`
- Required response shape:
  ```json
  {
    "data": {
      "moon": "<version>",
      "timestamp": "<RFC3339>"
    }
  }
  ```
- `status` or other non-specified keys MUST NOT be returned unless added to [SPEC_API/010-health.md](../SPEC_API/010-health.md).

### Validation Rules and Constraints
1. `timestamp` MUST be RFC3339 UTC format.
2. `moon` MUST contain server version string. version should be get from config 
3. Response wrapper key MUST be `data`.
4. No readiness diagnostics, DB details, registry counters, or internal health metadata in `/health` response.

### Error Handling and Failure Modes
1. `GET /health` SHOULD remain resilient and return `200` for liveness unless server cannot process request at HTTP layer.
2. Internal subsystem state MUST NOT change `/health` response schema.
3. Any removed endpoint/path formerly tied to dead health code is out of scope for compatibility preservation.

### Permissions, Filtering, Sorting, and Limits
1. Authentication/authorization: not required for `/health`.
2. No query parameters, filtering, sorting, or pagination apply.

### Use Cases
1. **Liveness check by orchestrator:** Poll `/health` and parse minimal `data` payload.
2. **Client compatibility with spec:** API consumers rely only on fields defined in [SPEC_API/010-health.md](../SPEC_API/010-health.md).
3. **Code maintainability:** Developers maintain one health implementation without dead alternatives.

### Non-Goals
1. No `/health/ready` readiness endpoint in this change.
2. No deep dependency diagnostics in health payload.
3. No backward compatibility for legacy/unused health service behavior.

### Needs Clarification
1. None.

## Acceptance Criteria

1. **Spec-Conformant Output**
   - `GET /health` response matches [SPEC_API/010-health.md](../SPEC_API/010-health.md) exactly.
   - Response does not include removed/legacy fields not defined by the spec.

2. **Dead Code Removed**
   - Unused `health.Service` implementation path and related non-routed legacy code are removed.
   - No references remain to removed health path from active server route registration.

3. **Lean Implementation**
   - Health handler remains minimal and single-purpose.
   - No dependency-heavy checks are executed for `/health`.

4. **Verification Scenarios**
   - Scenario A: `GET /health` returns `200` with required `data.moon` and `data.timestamp`.
   - Scenario B: Response schema remains stable and aligned with [SPEC_API/010-health.md](../SPEC_API/010-health.md).
   - Scenario C: Build/tests pass without references to removed dead health code.

5. **Regression Safety**
   - Public endpoint routing for `/health` remains available.
   - No unintended changes to unrelated endpoints.

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

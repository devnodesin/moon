## Overview

### Problem Statement
Three pre-existing tests in `internal/handlers/doc_test.go` are failing because generated documentation output is missing required content:
1. Prefix-aware curl examples when URL prefix is configured.
2. `{collection}` placeholder in quickstart/data endpoint examples.
3. An `Error Responses` section in HTML documentation.

### Context and Background
- The failures are documentation-generation mismatches, not API runtime behavior mismatches.
- The documentation endpoint must remain synchronized with test expectations and API specs.

### High-Level Solution Summary
Update the `DocHandler` generation logic/templates to include all required output elements expected by tests, ensure prefix-aware example rendering, keep placeholders consistent (`{collection}`), and add the missing `Error Responses` section. Add/adjust tests only as needed to codify the intended output contract and prevent regressions.

## Requirements

### Functional Requirements
1. `DocHandler` MUST generate curl examples that include configured URL prefix when prefix is set.
2. `DocHandler` MUST preserve and render `{collection}` placeholder in quickstart/data endpoint examples.
3. HTML documentation output MUST include an `Error Responses` section.
4. Generated docs MUST remain deterministic for test validation (stable headings/content blocks).
5. Existing documentation endpoints (`/doc`, `/doc/`, `/doc/llms.md`, `/doc/llms.txt`, `/doc/llms.json`) MUST remain available.

### Technical Requirements
1. Update the documentation template and/or handler composition logic used by `DocHandler`.
2. Ensure prefix injection is centralized in one formatter/helper path to avoid duplicated string logic.
3. Ensure placeholder rendering does not accidentally resolve `{collection}` to a concrete name in generic examples.
4. Keep changes minimal and scoped to documentation generation only.
5. Maintain existing response wrapper/route behavior for doc endpoints.

### Content/Output Requirements
1. **Prefix-aware examples**
   - With prefix unset: examples use base paths like `/health`, `/collections:list`, `/{collection}:list`.
   - With prefix set (e.g., `/api/v1`): examples use `/api/v1/health`, `/api/v1/collections:list`, `/api/v1/{collection}:list`.
2. **Quickstart placeholder**
   - Quickstart and data examples MUST use literal `{collection}` placeholder where generic collection examples are intended.
3. **Error Responses section**
   - HTML docs MUST contain a clearly labeled `Error Responses` section.
   - Section MUST summarize standard error shape and common HTTP codes consistent with current API docs.

### Validation Rules and Constraints
1. Do not change API behavior; this PRD only addresses generated documentation output.
2. Do not introduce backward-compatibility formatting shims for incorrect legacy output.
3. Keep markdown and HTML outputs logically consistent for equivalent sections.

### Error Handling and Failure Modes
1. If prefix configuration is empty or invalid, fallback to unprefixed examples without crashing doc generation.
2. Doc generation failures must return standard internal server error response handling already used by docs handler path.

### Use Cases
1. **Operator with API prefix:** Sees correctly prefixed curl examples in docs.
2. **Developer reading quickstart:** Sees generic `{collection}` placeholder and understands dynamic endpoints.
3. **Client developer handling errors:** Finds `Error Responses` section in HTML docs quickly.

### Non-Goals
1. No API endpoint contract changes.
2. No redesign of documentation site/theme.
3. No new documentation endpoints.

### Needs Clarification
1. None.

## Acceptance Criteria

1. **Failing Tests Resolved**
   - `TestDocHandler_WithPrefix` passes.
   - `TestDocHandler_QuickstartSection` passes.
   - `TestDocHandler_ErrorSection` passes.

2. **Prefix-Aware Output**
   - With configured prefix, rendered curl examples include the prefix consistently across major examples.

3. **Placeholder Correctness**
   - Quickstart/data endpoint examples include literal `{collection}` placeholder where expected.

4. **Error Section Presence**
   - HTML output contains `Error Responses` heading/section with expected baseline content.

5. **Regression Safety**
   - No unrelated handler/test regressions introduced.

- [x] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [x] Ensure all unit tests and integration tests in `tests/*` are passing successfully.

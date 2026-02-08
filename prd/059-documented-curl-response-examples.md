# PRD: Documented curl Response Examples in API Docs

## 1. Overview

### Problem Statement
Current API documentation in `cmd/moon/internal/handlers/templates/doc.md.tmpl` lacks consistent, explicit response examples for each `curl` request. This prevents automated tools and AI-based code generators from reliably parsing request/response pairs, leading to incomplete or incorrect code generation.

### Context and Background
The project aims to provide high-quality, machine-readable API documentation. Existing endpoint examples often show only the request, omitting the response or failing to document HTTP status codes. This gap hinders both human and automated consumption of the docs.

### High-Level Solution Summary
Update the template so every `curl` example includes a short title, a one-line description, a bash-formatted request, and a JSON-formatted response with explicit HTTP status. Where endpoints have multiple significant responses (success, validation error, auth error), include each with clear labeling.

## 2. Requirements

### Functional Requirements
- For every endpoint example in `doc.md.tmpl`, provide:
  - Short title and one-line description
  - Request block (bash fenced)
  - Response block (JSON fenced) with HTTP status
- Use `bash` for request blocks and `json` for response blocks
- Indicate expected HTTP status code next to Response heading (e.g., `Response (200):`)
- For endpoints with multiple significant responses, include examples for each (success, validation error, auth error)
- Example responses must be minimal but representative, including all important fields and typical values
- Requests requiring headers (Authorization, Content-Type, etc.) must include them inline in the `curl` example

### Technical Requirements
- Target file: `cmd/moon/internal/handlers/templates/doc.md.tmpl`
- All examples must follow the required format:
  1. Short title and one-line description
  2. Request (bash fenced block)
  3. Response (JSON fenced block) and HTTP status
- Automated parser must be able to extract request and response blocks unambiguously
- Examples must cover success and at least one common failure mode where applicable

### API Specifications
- No changes to API endpoints themselves; only documentation template is updated
- No new endpoints or parameters introduced

### Validation Rules and Constraints
- All example responses must be syntactically valid JSON
- HTTP status codes must be accurate and representative
- No extraneous fields or values in example responses

### Error Handling and Failure Modes
- For endpoints with common failure modes (validation error, auth error), include at least one example for each
- Clearly label each response with its HTTP status code

### Filtering, Sorting, Permissions, and Limits
- If an endpoint requires headers (e.g., Authorization), include them in the request example
- No changes to filtering, sorting, or permissions logic; documentation only

## 3. Acceptance Criteria

- Every `curl` example in `doc.md.tmpl` follows the Required Format:
  - Short title and one-line description
  - Request (bash fenced block)
  - Response (JSON fenced block) and HTTP status
- Automated parser can extract request and response blocks unambiguously
- Examples cover success and at least one common failure mode where applicable
- Example responses are minimal, representative, and syntactically valid JSON
- Requests requiring headers include them inline
- No undocumented or implied behavior

### Test Scenarios
- Review each endpoint example for compliance with required format
- Run automated parser to verify extraction of request and response blocks
- Validate example responses for JSON syntax and accuracy of HTTP status codes
- Confirm inclusion of at least one failure mode example per applicable endpoint

### Edge Cases and Negative Paths
- Endpoints with multiple failure modes: ensure each is documented
- Endpoints with non-JSON responses: document response in appropriate format and label status
- Endpoints requiring headers: verify headers are included in request example

---

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

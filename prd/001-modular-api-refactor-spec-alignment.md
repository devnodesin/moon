## Overview

This PRD defines the comprehensive refactor of the Moon API to achieve full alignment with the updated specifications (SPEC.md, SPEC_API.md, and SPEC_AUTH.md). The refactor removes all legacy and obsolete code, implements strict test-driven and specification-driven development, and ensures all documentation and scripts remain synchronized with code changes.

The refactor is executed as a series of discrete, self-contained modules—each with its own PRD, unit tests, integration tests, and verification steps. No backward compatibility or legacy support is required. Each module is independently implemented, tested, and merged before proceeding to the next.

This approach ensures surgical precision, minimal risk of cross-module interference, and complete specification compliance for each component before integration.

## Requirements

### 1. Preparation & Discovery

**1.1 Specification Analysis**
- Review SPEC.md for architectural requirements, data types, default values, validation constraints, and database schema design
- Review SPEC_API.md for all API endpoints, request/response formats, pagination patterns, query options, filtering, sorting, aggregation, and error codes
- Review SPEC_AUTH.md for authentication flows, JWT-based authentication, API key authentication, role-based access control, session management, security policies, and authentication-specific error codes

**1.2 Codebase Mapping**
- Identify all current API endpoints in `cmd/moon/internal/handlers/`
- Map each handler to its corresponding specification in SPEC_API.md and SPEC_AUTH.md
- Identify obsolete endpoints, features, or code not present in updated specifications
- Document all modules requiring refactor:
  - Authentication (`/auth:login`, `/auth:logout`, `/auth:refresh`, `/auth:me`)
  - User Management (`/users:list`, `/users:get`, `/users:create`, `/users:update`, `/users:destroy`)
  - API Key Management (`/apikeys:list`, `/apikeys:get`, `/apikeys:create`, `/apikeys:update`, `/apikeys:destroy`)
  - Collection Management (`/collections:list`, `/collections:get`, `/collections:create`, `/collections:update`, `/collections:destroy`)
  - Data Operations (`/data/{collection}:list`, `/data/{collection}:get`, `/data/{collection}:create`, `/data/{collection}:update`, `/data/{collection}:destroy`, `/data/{collection}:batch`)
  - Aggregation and Querying (`/data/{collection}:aggregate`, `/data/{collection}:count`)
  - Documentation (`/doc/`, `/doc/llms.md`, `/doc/llms.txt`, `/doc/llms.json`, `/doc:refresh`)
  - Health Check (`/health`)

**1.3 Dependency Identification**
- Identify shared modules: middleware, authentication, authorization, validation, pagination, query building, schema registry, error handling, logging
- Map dependencies between modules to determine refactor sequence
- Ensure shared modules are refactored before dependent modules

### 2. Task Decomposition

**2.1 Module Breakdown**
For each identified module:
- Create a dedicated PRD outlining scope, requirements, and acceptance criteria per SPEC.md, SPEC_API.md, and SPEC_AUTH.md
- Define test coverage requirements (unit tests, integration tests, edge cases, negative paths)
- Specify documentation updates required (README.md, INSTALL.md, moon.conf, scripts, samples)

**2.2 Prioritization**
Refactor modules in dependency order:
1. **Core Infrastructure** (constants, errors, logging, decimal handling, ULID generation, validation, pagination)
2. **Database Layer** (driver, inspector, schema builder, registry, query builder)
3. **Authentication & Authorization** (JWT tokens, API keys, middleware, password policy, token blacklist, session management)
4. **User Management** (user repository, user handlers)
5. **API Key Management** (API key repository, API key handlers)
6. **Collection Management** (collection schema, collection handlers)
7. **Data Operations** (data handlers, batch operations, filtering, sorting)
8. **Aggregation & Querying** (aggregation handlers, count operations)
9. **Documentation & Health** (documentation handlers, health check)
10. **Server & Daemon** (server initialization, graceful shutdown, preflight checks, bootstrap admin)

### 3. Test-Driven & Spec-Driven Development (Per Module)

**3.1 Remove Obsolete Code**
- Delete all tests for obsolete features or endpoints not present in updated specifications
- Remove all legacy code, deprecated endpoints, and obsolete handlers
- Remove obsolete middleware, helpers, or utilities no longer required

**3.2 Write New Tests**
- Write unit tests for all new or updated functions, methods, and logic per SPEC.md and SPEC_API.md
- Write integration tests for all API endpoints per SPEC_API.md and SPEC_AUTH.md
- Test all success paths, error paths, edge cases, validation rules, and security constraints
- Test all filtering, sorting, pagination, aggregation, and query options per SPEC_API.md
- Ensure all tests follow Go testing conventions and use table-driven patterns where applicable
- Target 90% test coverage for all modules

**3.3 Refactor or Rewrite Implementation**
- Refactor or rewrite module implementation to match updated specifications exactly
- Ensure all request/response formats, error codes, and status codes match SPEC_API.md
- Ensure all authentication flows, token handling, and permission checks match SPEC_AUTH.md
- Ensure all data types, validation constraints, and default values match SPEC.md
- Run tests continuously during implementation to ensure correctness
- Fix all test failures before proceeding to next module

**3.4 Verify Specification Compliance**
- Cross-reference implementation against SPEC.md, SPEC_API.md, and SPEC_AUTH.md
- Verify all endpoints, request/response patterns, query options, error codes, and authentication flows
- Ensure no undocumented behavior or assumptions
- Ensure strict adherence to AIP-136 custom actions pattern (resource:action)

### 4. Documentation & Script Updates

**4.1 Update Documentation**
- Update README.md to reflect new API structure, endpoints, and features
- Update INSTALL.md with correct installation, setup, and configuration instructions
- Ensure moon.conf includes all required configuration options with inline documentation
- Update API documentation template at `cmd/moon/internal/handlers/templates/doc.md.tmpl`

**4.2 Update Scripts**
- Update `scripts/api-check.py` to test all new API endpoints, request/response formats, and error codes
- Update test scripts in `scripts/tests/*.json` to match new API behaviors
- Ensure `/doc/llms.json` endpoint reflects latest API structure and capabilities

**4.3 Update Samples**
- Update all sample files in `samples/` to match new API request/response formats
- Ensure all samples are executable and valid against new API

### 5. Verification & Quality Assurance

**5.1 Test Execution**
- Run all unit tests for the module: `go test -v ./cmd/moon/internal/{module}/...`
- Run all integration tests for the module
- Ensure 100% pass rate before proceeding
- If any test failure is unrelated to the module, investigate and fix before marking module complete

**5.2 Code Review**
- Require code review for each module
- Verify strict adherence to SPEC.md, SPEC_API.md, and SPEC_AUTH.md
- Verify test coverage meets 90% target
- Verify all documentation and scripts are updated

**5.3 Manual Verification**
- Manually test all API endpoints for the module using `curl` or API client
- Verify request/response formats, error codes, and status codes match specifications
- Verify filtering, sorting, pagination, and aggregation behavior
- Verify authentication, authorization, and permission enforcement

### 6. Iteration & Completion

**6.1 Repeat Process**
- Complete steps 3-5 for each module in dependency order
- Do not proceed to next module until current module passes all tests, code review, and verification

**6.2 Final Integration**
- After all modules are refactored, run full test suite: `go test -v ./...`
- Ensure 100% pass rate across all modules
- Run `scripts/api-check.py` to verify all endpoints
- Verify `/doc/llms.json` is current and accurate

**6.3 Cleanup**
- Remove all obsolete test files, code, and documentation
- Remove all TODOs, dead code, and partial fixes
- Format all Go files: `gofmt -w .`
- Run `go mod tidy` to clean up dependencies

## Acceptance

### Acceptance Criteria

**AC1: Specification Compliance**
- All API endpoints, request/response formats, error codes, and status codes match SPEC_API.md exactly
- All authentication flows, token handling, and permission checks match SPEC_AUTH.md exactly
- All data types, validation constraints, default values, and database schema match SPEC.md exactly
- No undocumented or legacy behavior remains

**AC2: Test Coverage**
- All modules have unit tests and integration tests
- Test coverage is at least 90% across all modules
- All tests pass with 100% success rate
- All edge cases, error paths, and negative paths are tested

**AC3: Code Quality**
- No obsolete code, legacy endpoints, or deprecated features remain
- Code follows Go best practices and idiomatic patterns
- All code is formatted with `gofmt`
- No compilation warnings or errors
- All dependencies are current and minimal

**AC4: Documentation Accuracy**
- README.md, INSTALL.md, moon.conf, and all documentation are current and accurate
- API documentation template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` reflects all changes
- `scripts/api-check.py` and `scripts/tests/*.json` are updated and pass
- `/doc/llms.json` endpoint is current and accurate
- All samples in `samples/` are valid and executable

**AC5: Module Independence**
- Each module is self-contained with its own PRD and test suite
- No cross-module dependencies except through well-defined interfaces
- Each module can be tested and verified independently

**AC6: Security & Performance**
- All authentication and authorization checks are enforced per SPEC_AUTH.md
- Password policies, rate limiting, and session management are implemented correctly
- Memory footprint remains under 50MB as per SPEC.md
- No security vulnerabilities or credential leaks

### Verification Steps

**Step 1: Run All Tests**
```bash
go test -v ./...
```
Expected: 100% pass rate, no failures, coverage ≥ 90%

**Step 2: Verify API Endpoints**
```bash
python scripts/api-check.py
```
Expected: All endpoints respond correctly per SPEC_API.md

**Step 3: Verify Documentation Endpoint**
```bash
curl http://localhost:6006/doc/llms.json
```
Expected: Current API schema reflecting all refactored endpoints

**Step 4: Manual Endpoint Verification**
Test each module's endpoints manually:
- Authentication: `/auth:login`, `/auth:logout`, `/auth:refresh`, `/auth:me`
- Users: `/users:list`, `/users:get`, `/users:create`, `/users:update`, `/users:destroy`
- API Keys: `/apikeys:list`, `/apikeys:get`, `/apikeys:create`, `/apikeys:update`, `/apikeys:destroy`
- Collections: `/collections:list`, `/collections:get`, `/collections:create`, `/collections:update`, `/collections:destroy`
- Data: `/data/{collection}:list`, `/data/{collection}:get`, `/data/{collection}:create`, `/data/{collection}:update`, `/data/{collection}:destroy`, `/data/{collection}:batch`
- Aggregation: `/data/{collection}:aggregate`, `/data/{collection}:count`
- Documentation: `/doc/`, `/doc/llms.md`, `/doc/llms.txt`, `/doc/llms.json`, `/doc:refresh`
- Health: `/health`

Expected: All endpoints return correct responses per specifications

**Step 5: Code Review Checklist**
- [ ] All obsolete and legacy code removed
- [ ] All endpoints follow AIP-136 custom actions pattern
- [ ] All request/response formats match SPEC_API.md
- [ ] All authentication flows match SPEC_AUTH.md
- [ ] All data types and validation match SPEC.md
- [ ] All tests pass with ≥ 90% coverage
- [ ] All documentation and scripts updated
- [ ] No compilation warnings or errors
- [ ] Code formatted with `gofmt`

### Post-Implementation Checklist

- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [ ] Update API doc template at `cmd/moon/internal/handlers/templates/doc.md.tmpl` to reflect these changes.
- [ ] Run all tests and ensure 100% pass rate.
- [ ] If any test failure is unrelated to your feature, investigate and fix it before marking the task as complete.

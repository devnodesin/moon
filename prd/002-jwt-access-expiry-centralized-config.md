## Overview

### Problem Statement
JWT access token expiry behavior is inconsistent because defaults are defined in more than one place. Configuration-level default indicates 1 hour while server fallback logic uses 15 minutes, creating ambiguity and potential drift.

### Context and Background
- Current implementation can apply different defaults depending on where fallback is evaluated.
- This increases maintenance risk and can produce unexpected token lifetimes.

### High-Level Solution Summary
Centralize JWT access token expiry resolution in exactly one configuration source-of-truth. The effective value must be loaded from `moon.conf` when provided. If not explicitly specified by user config, use config defaults from the centralized config module. Runtime/server code must not define a separate fallback constant.

## Requirements

### Functional Requirements
1. JWT access token expiry MUST be configurable through `moon.conf` using `jwt.access_expiry`.
2. System MUST resolve effective access expiry from centralized configuration only.
3. If `jwt.access_expiry` is explicitly set in `moon.conf`, that value MUST be used.
4. If `jwt.access_expiry` is not specified in `moon.conf`, system MUST fall back to centralized config defaults.
5. Fallback default expiry for access tokens MUST be **15 minutes** (`900` seconds).
6. No secondary default/fallback for access expiry may exist in server/runtime layers.
7. If `moon.conf` cannot be loaded or `jwt.access_expiry` cannot be parsed/validated, service MUST continue startup using the centralized default (`900`).
8. On fallback due to config error, system MUST emit clear logs describing the issue, expected value format, and applied default.

### Technical Requirements
1. Define the default for `jwt.access_expiry` in one place only (configuration default registration path).
2. Remove or replace any hardcoded server-level fallback constant for access expiry.
3. Token service initialization MUST consume already-resolved config values rather than applying additional defaults.
4. Existing config parsing/validation flow MUST remain compatible with current config loading architecture.
5. Add tests to verify precedence and defaulting:
   - explicit value in `moon.conf` overrides defaults,
   - missing key uses centralized default,
   - no duplicate fallback path in server runtime.

### Configuration Specification
- Config file: `moon.conf`
- Key: `jwt.access_expiry`
- Unit: seconds
- Expected behavior:
  - `jwt.access_expiry: <positive integer>` -> use specified value
  - key omitted -> use centralized default (`900`)
   - file missing/unreadable/invalid format -> continue startup with centralized default (`900`) and warning logs
   - key present but invalid (`<=0` or non-integer) -> continue startup with centralized default (`900`) and warning logs

### Validation Rules and Constraints
1. `jwt.access_expiry` MUST be validated as a positive integer.
2. Zero or negative values MUST be treated as invalid and trigger fallback to centralized default (`900`).
3. Invalid types (string/float/object) MUST trigger fallback to centralized default (`900`).
4. Default assignment MUST occur during config load/default registration, not during auth request handling.
5. Invalid config input MUST NOT terminate service startup for this setting.

### Logging Requirements
1. When fallback is used, log entry MUST include:
   - config source (`moon.conf`),
   - key name (`jwt.access_expiry`),
   - invalid/failed input value when safely available,
   - reason (missing key, parse error, invalid type, invalid range, file read/decode error),
   - expected format (`positive integer in seconds`),
   - applied default value (`900`).
2. Log severity SHOULD be `WARN` for recoverable config issues.
3. Log message MUST be actionable, including what operator should change in `moon.conf`.
4. Startup success log SHOULD include final effective access expiry value.

### Error Handling and Failure Modes
1. Invalid `jwt.access_expiry` in `moon.conf` MUST NOT fail startup; system logs warning and applies default (`900`).
2. Missing `jwt` section or missing `access_expiry` key MUST not fail startup; centralized defaults apply.
3. Config file read/decode errors for `moon.conf` MUST NOT fail startup for this setting; system logs warning and applies default (`900`).
4. Runtime auth flow MUST not alter expiry value after config load.

### Use Cases
1. **Configured expiry:** Operator sets `jwt.access_expiry: 1800` in `moon.conf`; tokens use 30 minutes.
2. **Defaulted expiry:** Operator omits `jwt.access_expiry`; tokens use centralized default 900 seconds.
3. **Invalid config value:** Operator sets `jwt.access_expiry: 0`; startup continues, warning is logged, and 900 is used.
4. **Invalid config file:** `moon.conf` has parse error; startup continues, warning is logged, and 900 is used.

### Non-Goals
1. No changes to refresh token expiry behavior unless already tied to the same centralized config path.
2. No changes to JWT claims structure beyond `exp` calculation input.
3. No API contract changes for auth endpoints.

### Needs Clarification
1. None.

## Acceptance Criteria

1. **Single Source of Truth**
   - Codebase contains one authoritative default definition for `jwt.access_expiry`.
   - No server/runtime fallback constant remains for access expiry.

2. **Configurability via `moon.conf`**
   - Given `jwt.access_expiry` is set in `moon.conf`, issued access tokens use that exact TTL.

3. **Centralized Default Behavior**
   - Given `jwt.access_expiry` is omitted in `moon.conf`, issued access tokens use 900-second default from centralized config defaults.

4. **Validation Enforcement**
   - Invalid values (`<=0`, non-integer) do not stop startup; warning log is emitted and default 900 is applied.

5. **Resilient Startup on Config Errors**
   - If `moon.conf` is unreadable or malformed, service still starts and uses default 900 for access expiry.
   - Warning log clearly states the config issue, expected format, and applied default.

6. **Test Coverage**
   - Unit tests cover explicit value precedence, omitted-key fallback to centralized defaults, and rejection of invalid values.
   - Unit tests verify invalid values/file parse failures produce warning logs and use default 900 without startup failure.
   - Tests verify no duplicate fallback logic path in server initialization.

7. **Regression Safety**
   - Existing authentication and token issuance tests continue to pass.

- [x] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.
- [x] Ensure all unit tests and integration tests in `tests/*` are passing successfully.

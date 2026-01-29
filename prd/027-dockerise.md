## Overview

- Add containerized deployment so Moon can run in a consistent, portable runtime
- Provide a small production image built via a multi-stage Docker build
- Support the project’s default config path and SQLite defaults via volume mounts

## Requirements

- Repository artifacts
  - Add a `Dockerfile` at the repository root

- Build requirements
  - The Docker build must compile the `moon` binary from `cmd/moon`
  - Use a multi-stage build:
    - Builder stage uses the Go toolchain to compile
    - Final stage contains only what is required to run the binary
  - The final image must be as small as practical while remaining compatible with the project requirements

- Runtime requirements
  - The container must run a single foreground `moon` process (console mode)
  - Needs Clarification: whether daemon mode should be supported inside the container. Default is console mode only.
  - The container must expose the configured HTTP port (default: 6006)

- Configuration and persistence
  - The container must support mounting a YAML config at `/etc/moon.conf` (the default config path)
  - The container must support persisting SQLite data by mounting the database path directory (default: `/opt/moon/`)

- Operational constraints
  - The container must not require environment variables for configuration (YAML-only configuration)
  - The container must start successfully when provided a valid YAML config file mounted to `/etc/moon.conf`

- Tests / validation
  - Add an automated smoke test that validates the Docker build succeeds
  - Needs Clarification: whether CI exists and where to place Docker build validation. If no CI pipeline exists, document validation steps in acceptance and rely on local verification.

## Acceptance

- A root-level `Dockerfile` exists and builds successfully
- The built image runs `moon` in the foreground using `/etc/moon.conf`
- The container exposes port 6006 by default
- Moon can be run with:
  - A mounted config at `/etc/moon.conf`
  - A mounted data directory for `/opt/moon/` so SQLite persists across container restarts
- No environment variables are required for configuration
- Updated the scripts, documentations and specifications.

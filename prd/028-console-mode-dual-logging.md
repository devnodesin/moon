## Overview

- In console (foreground) mode, Moon currently logs only to stdout
- For local development and container usage, it should also write the same logs to the configured log file
- Add “tee” logging in console mode: send logs to both terminal and `logging.path/main.log`

## Requirements

- Current behavior
  - Daemon mode writes logs to a file: `filepath.Join(cfg.Logging.Path, "main.log")`
  - Console mode currently logs to stdout only

- Desired behavior (console mode)
  - When running in console mode (no `--daemon` / `-d`), Moon must write logs to:
    - stdout (terminal)
    - the log file at `filepath.Join(cfg.Logging.Path, "main.log")`
  - Log messages must be identical in content (same fields, level, message) across both outputs

- Logging format
  - Console mode must continue to use the existing console-friendly formatting for stdout
  - The file output should use the existing file format used elsewhere (currently “simple”), unless changed globally
  - Needs Clarification: whether console mode should write “console” format to the file as well. Default: keep file in “simple” format.

- File handling
  - The file writer must open the log file in append mode
  - The log directory must be created if missing (same behavior as current logger file path handling)
  - Fail-safe behavior:
    - If the log file cannot be opened/created, Moon must still log to stdout
    - The failure to open the file must be reported to stderr (consistent with current logger behavior)

- Configuration
  - No new YAML settings are required for the initial implementation
  - Console-mode file logging must respect the existing `logging.path` configuration
  - Needs Clarification: whether to add an explicit opt-out flag (e.g., `logging.console_file=false`). Default: no opt-out.

- Docker/containers
  - This feature must work in containers where stdout is the primary log sink
  - File logging must still work when `/var/log/moon` is bind-mounted to the host

- Tests
  - Add unit tests for the logging package to verify “tee” behavior:
    - When both outputs are available, a log call writes to both
    - When the file output fails, logging continues to stdout
  - Add an integration-style test (where practical in the current test suite) validating:
    - In console mode, logs are produced and written to `logging.path/main.log`

## Acceptance

- Running Moon in console mode produces logs in the terminal and appends logs to `logging.path/main.log`
- If the log file cannot be opened, Moon continues logging to stdout and reports the file-open failure to stderr
- Existing daemon mode logging behavior is unchanged
- Automated tests cover:
  - dual-write (stdout + file)
  - file-open failure fallback

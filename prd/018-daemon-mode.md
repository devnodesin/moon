## Overview

- Implement daemon mode to run Moon as a background service
- By default, Moon runs in foreground (console mode) with logs to stdout/stderr
- Add `-d` or `--daemon` flag to enable background daemon mode with file-based logging
- Support proper Unix daemon behavior with process detachment and PID management
- Updated the `SPEC.md` to match this changes
- Updated the DOCUMENTATIONS: `README.md`, `INSTALL.md` `USAGE.md` to match this changes

## Requirements

- Add command-line flag `-d` or `--daemon` in `cmd/moon/main.go` to enable daemon mode
- Console mode (default behavior):
  - Run in foreground attached to terminal
  - Display logs directly to stdout/stderr
  - Process terminates when terminal closes or Ctrl+C pressed
- Daemon mode (when `-d` flag provided):
  - Detach process from terminal (fork/setsid on Unix, service mode on Windows)
  - Redirect logs to file defined by `LOG_PATH` configuration
  - Write PID file to `/var/run/moon.pid` or configurable location
  - Process continues running after terminal closes
  - Handle signals properly (SIGTERM, SIGINT for graceful shutdown)
- Use `LOG_PATH` configuration value for log file location (default: `/var/log/moon`)
- Default log file in daemon mode: `{LOG_PATH}/main.log` = `/var/log/moon/main.log`
- Integrate with existing graceful shutdown mechanism in `internal/shutdown/shutdown.go`
- Daemon process must properly clean up PID file on exit
- Support both console and daemon modes with `--config` flag for configuration
- On Unix/Linux: use standard daemonization (double fork, setsid, chdir to /)
- On Windows: run as foreground service (daemonization handled externally by service managers)
- Update `internal/logging/logger.go` to support file output configuration
- Add `LoggingConfig` to `AppConfig` with `Path` and `File` fields
- Ensure log directory exists or create it on startup in daemon mode
- Handle log file permissions and ownership appropriately
- Document daemon mode usage in README and INSTALL documentation

## Acceptance

- Running `moon` starts server in console mode with logs to stdout/stderr
- Running `moon -d` or `moon --daemon` starts server in background daemon mode
- Daemon process successfully detaches from terminal and continues running after terminal closes
- Logs written to configured `LOG_PATH` location in daemon mode (default: `/var/log/moon/main.log`)
- PID file created at `/var/run/moon.pid` when running in daemon mode
- PID file deleted on graceful shutdown
- Graceful shutdown works correctly in both console and daemon modes
- SIGTERM and SIGINT signals handled properly in daemon mode
- Configuration flag `--config` works with both console and daemon modes
- Daemon mode respects LOG_PATH configuration from YAML file
- Log directory created if it doesn't exist (with proper error handling if permissions denied)
- Unit tests verify daemon initialization, PID file creation, and cleanup
- Integration tests verify signal handling and graceful shutdown in daemon mode
- Documentation updated with daemon mode usage examples and systemd service file
- `samples/moon.service` systemd unit file provided for production deployment
- Updated the `SPEC.md` to match this changes
- Updated the DOCUMENTATIONS: `README.md`, `INSTALL.md` `USAGE.md` to match this changes

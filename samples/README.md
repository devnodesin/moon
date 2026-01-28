# Moon Samples

This directory contains example configuration files and scripts to help you get started with Moon.

## Configuration Files

### config.example.yaml

Example YAML configuration file showing all available options. Copy this to `config.yaml` in the project root and customize as needed:

```bash
cp samples/config.example.yaml config.yaml
```

### .env.example

Example environment variables file. Environment variables take precedence over config file settings. Copy this to `.env` in the project root:

```bash
cp samples/.env.example .env
```

**Important:** Make sure to change the `MOON_JWT_SECRET` to a secure random string in production!

## Scripts

### api-demo.sh

A comprehensive demonstration script that shows all major Moon API operations:
- Creating collections (database tables)
- Managing collection schemas
- CRUD operations on data
- Pagination and filtering

**Usage:**

```bash
# Standalone mode (auto-starts server)
# Script will create a temporary server and clean up automatically
./samples/api-demo.sh

# With existing server
# Start Moon server first, then run the demo
./moon --config config.yaml &
./samples/api-demo.sh
```

**Features:**
- **Auto-start mode**: If no server is running, the script automatically creates a temporary configuration and starts a Moon server for the demo
- **Temporary environment**: Uses `/tmp` for database and logs - no special permissions needed
- **Auto-cleanup**: Automatically stops the server and removes temporary files when complete
- **Existing server support**: Detects if a server is already running and uses it instead

The script will walk through:
1. Health check
2. Collection management (create, list, get, update, destroy)
3. Data operations (create, read, update, delete)
4. Schema modifications

### test-runner.sh

A convenient test runner with multiple modes:

```bash
# Run all tests
./samples/test-runner.sh

# Run only unit tests
./samples/test-runner.sh unit

# Run tests with coverage report
./samples/test-runner.sh coverage

# Run tests with race detector
./samples/test-runner.sh race

# Run benchmarks
./samples/test-runner.sh bench
```

## Quick Start

1. Build Moon:
   ```bash
   go build -o moon ./cmd/moon
   ```

2. Try the API demo (standalone mode - no config needed):
   ```bash
   ./samples/api-demo.sh
   ```

The demo script will automatically:
- Create a temporary configuration
- Start a Moon server
- Run through all API operations
- Clean up when complete

For production use, copy configuration files:
```bash
cp samples/moon.conf /etc/moon.conf
# Edit /etc/moon.conf and set jwt.secret
```

For detailed documentation, see:
- [Installation Guide](../docs/INSTALL.md)
- [Usage Guide](../docs/USAGE.md)
- [Project README](../README.md)

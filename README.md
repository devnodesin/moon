# Moon - Dynamic Headless Engine

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE) [![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)

Moon is an API-first, migration-less backend in Go. Manage database schemas and data via REST APIs—no migration files needed.

## Features

- Migration-less schema management (create/modify tables via API)
- In-memory schema registry for fast validation
- Multi-database: SQLite, PostgreSQL, MySQL
- Predictable API pattern (AIP-136 custom actions)
- Built-in HTML & Markdown documentation (`/doc/`, `/doc/llms.md`)
- Docker-ready, efficient (<50MB RAM)
- **JWT & API key authentication** (mandatory)
- **Role-based access control** (admin, user with can_write)
- **Rate limiting** (100 req/min JWT, 1000 req/min API key)
- ULID identifiers
- Headless Backend for API-first apps like CMS, E-Commerce, CRM, Blog, Datastores etc.

## Quick Start

```bash
git clone https://github.com/thalib/moon.git
cd moon
```

See [INSTALL.md](INSTALL.md) for complete setup including Docker deployment.

## Documentation

Moon provides comprehensive, auto-generated API documentation:

- **HTML Documentation**: Visit `http://localhost:6006/doc/` in your browser for a complete, interactive API reference
- **Markdown Documentation**: Access `http://localhost:6006/doc/llms-full.md` for terminal-friendly or AI-agent documentation
- Configuration: See `moon.conf` in the project root for comprehensive, self-documented configuration

### Additional Resources

- [INSTALL.md](INSTALL.md): Installation and deployment guide (includes authentication setup)
- [SPEC.md](SPEC.md): Architecture and technical specifications
- [SPEC_API.md](SPEC_API.md): Complete API reference (endpoints, request/response formats, query options)
- [samples/](samples/): Sample and Install scripts
- [scripts/](scripts/): Python API Test Suite
- [LICENSE](LICENSE): MIT License
- [GitHub Issues](https://github.com/thalib/moon/issues)
- [GitHub Discussions](https://github.com/thalib/moon/discussions)

## License & Credits

MIT License ([LICENSE](LICENSE))

---

Made by [Devnodes.in](https://github.com/devnodesin)

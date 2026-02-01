# Moon - Dynamic Headless Engine

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE) [![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)

Moon is an API-first, migration-less backend in Go. Manage database schemas and data via REST APIs—no migration files needed.

## Features

- Migration-less schema management (create/modify tables via API)
- In-memory schema registry for fast validation
- Multi-database: SQLite, PostgreSQL, MySQL
- Predictable API pattern (AIP-136 custom actions)
- Built-in HTML & Markdown documentation (`/doc/`, `/doc/md`)
- Server-side aggregations (`:count`, `:sum`, etc.)
- Docker-ready, efficient (<50MB RAM)
- JWT & API key auth
- ULID identifiers
- Headless Backend for API-first apps like CMS, E-Commerce, CRM, Blog, Datastores etc.

## Quick Start

```bash
git clone https://github.com/thalib/moon.git
```

See [INSTALL.md](INSTALL.md) for setup.

## Documentation

Moon provides comprehensive, auto-generated API documentation:

- **HTML Documentation**: Visit `http://localhost:6006/doc/` in your browser for a complete, interactive API reference
- **Markdown Documentation**: Access `http://localhost:6006/doc/md` for terminal-friendly or AI-agent documentation
- Configuration: See `samples/moon.conf` and `samples/moon-full.conf`.
- Testing: See `scripts/test-runner.sh`

### Additional Resources

- [INSTALL.md](INSTALL.md): Installation and deployment guide
- [SPEC.md](SPEC.md): Architecture and technical specifications
- [samples/](samples/): Sample configuration files
- [scripts/](scripts/): Test and demo scripts
- [LICENSE](LICENSE): MIT License
- [GitHub Issues](https://github.com/thalib/moon/issues)
- [GitHub Discussions](https://github.com/thalib/moon/discussions)

## License & Credits

MIT License ([LICENSE](LICENSE))
Built with [Go](https://go.dev/), [Viper](https://github.com/spf13/viper), [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql), [lib/pq](https://github.com/lib/pq), [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)

---

Made by [Devnodes.in](https://github.com/devnodesin)

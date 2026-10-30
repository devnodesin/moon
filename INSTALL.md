# Moon Installation & Deployment

Moon is a single-process, API-first backend. It supports three database backends:
**SQLite** (default), **PostgreSQL**, and **MySQL**. Only SQLite and PostgreSQL
are fully implemented; MySQL is a declared but unimplemented backend.

## 1. Prerequisites

- Go 1.25+ (see `go.mod` for the exact toolchain version)
- A database:
  - SQLite: none (a file path is used)
  - PostgreSQL: a reachable PostgreSQL 12+ server
  - MySQL: not yet supported

## 2. Build

```sh
go build -o moon ./cmd
```

## 3. Configuration

Copy `moon.conf` and edit it. The relevant database keys:

| Key                | SQLite default      | PostgreSQL / MySQL requirement        |
| ------------------ | ------------------- | ------------------------------------- |
| `database.connection` | `sqlite`          | `postgres` or `mysql`                 |
| `database.database`   | `/opt/moon/sqlite.db` | database name (PG/MySQL)             |
| `database.user`       | —                 | required for PG/MySQL                 |
| `database.password`   | —                 | required for PG/MySQL                 |
| `database.host`       | —                 | required for PG/MySQL (`host:port`)   |

### SQLite example

```yaml
database:
  connection: "sqlite"
  database: "/opt/moon/sqlite.db"
```

### PostgreSQL example

```yaml
database:
  connection: "postgres"
  database: "moon"
  user: "moon_user"
  password: "strong-password"
  host: "127.0.0.1:5432"
```

Moon validates that `user`, `password`, and `host` are present when the backend
is `postgres` or `mysql`. Startup fails if they are missing.

## 4. PostgreSQL driver & behavior notes

- The PostgreSQL adapter uses the [`pgx/v5`](https://github.com/jackc/pgx)
  driver via its `stdlib` layer (driver name `pgx`). `lib/pq` is intentionally
  avoided (maintenance mode).
- External API behavior is identical across backends (SPEC.md §9.3). Type
  mapping for PostgreSQL:

  | Moon type | PostgreSQL column type |
  | --------- | --------------------- |
  | string    | `TEXT`                |
  | integer   | `BIGINT`              |
  | decimal   | `NUMERIC(19,2)`       |
  | boolean   | `BOOLEAN`             |
  | datetime  | `TIMESTAMP`           |
  | json      | `JSON`                |

- Datetime values are stored as `TIMESTAMP` in PostgreSQL but returned as RFC3339
  strings through the API, exactly like SQLite's `TEXT` storage.
- System tables (`users`, `apikeys`, `moon_auth_refresh_tokens`) are created
  automatically on startup via idempotent DDL. Minor backend-specific defaults
  (e.g. `BOOLEAN DEFAULT false` instead of `DEFAULT 0`) are normalized
  internally.

## 5. Running

```sh
./moon -c /etc/moon.conf
```

The service binds to `server.host:server.port` and logs to both the console and
`server.logpath`. On first start, if bootstrap admin fields are present in the
config, an admin user is created; remove those fields afterward.

## 6. Data migration: SQLite → PostgreSQL

Moon has **no built-in backup/restore** (SPEC.md §4). Migration is an external
operational task. Two approaches:

### Option A — API-based (recommended, safe, validates every record)

1. **Stop the SQLite instance.**
2. **Export** every collection (system + dynamic) from the SQLite instance:
   ```sh
   curl -H "Authorization: Bearer $ADMIN_JWT" http://localhost:6006/collections:query
   # for each collection:
   curl -H "Authorization: Bearer $ADMIN_JWT" "http://localhost:6006/data/<collection>:query?per_page=200"
   ```
3. **Start a fresh Moon instance pointed at PostgreSQL** (empty database). It
   creates the system tables automatically.
4. **Re-create dynamic collections** via `POST /collections:mutate` using the
   schema obtained from the SQLite instance's `/data/<collection>:schema`.
5. **Import records** via `POST /data/<collection>:mutate` (`op=create`).

This preserves ULIDs, bcrypt password hashes, and SHA-256 API-key hashes, all of
which are backend-portable.

### Option B — SQL dump (faster for large datasets)

1. Dump SQLite: `sqlite3 /opt/moon/sqlite.db .dump > moon.sql`
2. Load into PostgreSQL with a conversion tool such as `pgloader`, or manually
   cast columns: `decimal` must remain a string in the API layer, and `json`
   columns must contain valid JSON. Be aware SQLite stores `boolean` as `0/1`
   integers and `datetime` as RFC3339 `TEXT`; map them to PG `BOOLEAN` /
   `TIMESTAMP` only if you also adjust the physical schema consistently.

> **Note:** After migration, verify with `GET /health` and a login
> (`POST /auth:session`) plus a collection listing before decommissioning the
> SQLite instance.

## 7. Backups

Because there is no built-in backup:

- **SQLite:** stop the service (briefly) and copy the `.db` file plus its
  `-wal`/`-shm` companions, or run `sqlite3 <file> .dump`.
- **PostgreSQL:** use `pg_dump` (e.g. `pg_dump -Fc moon > moon.dump`) on a
  schedule; restore with `pg_restore`.

## 8. Running the test suite

Tests default to SQLite and require no database server. PostgreSQL tests are
**opt-in** via an environment variable and are skipped when it is unset:

```sh
# SQLite only (default)
go test ./cmd

# With PostgreSQL (requires a running server)
docker run --rm -e POSTGRES_PASSWORD=test -e POSTGRES_DB=moon -p 5432:5432 postgres:16
export MOON_TEST_POSTGRES_DSN='postgres://postgres:test@localhost:5432/moon?sslmode=disable'
go test ./cmd
```

`go vet ./...` and `gofmt` must pass cleanly before changes are considered
complete.

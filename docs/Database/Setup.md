# Database setup

SSTorytime stores its semantic-spacetime cache in PostgreSQL (17+). This page
walks through the three supported ways to bring a database online — Docker,
manual/package install, and a RAM-disk variant for bulk-load benchmarking —
and then explains how the Go client picks up credentials at `Open()` time.

!!! tip "Which setup should I pick?"
    - **Docker** — easiest for development laptops, zero host pollution.
    - **Manual / package install** — right for servers you already run Postgres on.
    - **RAM disk** — only for throwaway bulk-load experiments; see the trade-off
      warnings below.

## Docker (recommended)

The `postgres-docker/` directory contains a complete single-node setup.

```yaml title="postgres-docker/docker-compose.yml"
services:
  postgres:
    image: postgres:17-alpine
    container_name: sstorytime-postgres
    environment:
      POSTGRES_USER: sstoryline
      POSTGRES_PASSWORD: sst_1234
      POSTGRES_DB: sstoryline
    ports:
      - "127.0.0.1:5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
```

See [`postgres-docker/docker-compose.yml`](https://github.com/markburgess/SSTorytime/blob/main/postgres-docker/docker-compose.yml).
Three things worth noting:

1. **Bind to loopback only** (`127.0.0.1:5432:5432`). The default password is
   trivial; do not expose this container to an untrusted network.
2. **Named volume `postgres_data`** holds the cluster on disk between restarts.
   Delete it to start clean (`docker compose down -v`).
3. **`init-db.sql`** is mounted into `/docker-entrypoint-initdb.d/`, which
   Postgres runs once on first startup. Its critical line is:

    ```sql title="postgres-docker/init-db.sql:5"
    CREATE EXTENSION IF NOT EXISTS unaccent;
    ```

    The `unaccent` extension is mandatory — every accent-insensitive search in
    SSTorytime goes through it (see the
    [`sst_unaccent` wrapper at postgres_types_functions.go:1747](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L1747)).

Bring the container up:

```sh
cd postgres-docker
docker compose up -d
docker compose logs -f postgres   # until you see "database system is ready"
```

Then proceed to a [connection check](#connection-config) below.

## Manual / package install

If Postgres is already installed on the host, the repository ships a setup
script at [`contrib/makedb.sh`](https://github.com/markburgess/SSTorytime/blob/main/contrib/makedb.sh):

```sh
sh contrib/makedb.sh
```

The script checks whether the `sstoryline` database already exists; if not it
creates the role, database, grants, and loads the `unaccent` extension:

```sh title="contrib/makedb.sh:14"
echo "CREATE USER sstoryline PASSWORD 'sst_1234' superuser;
      CREATE DATABASE sstoryline;
      GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline;
      CREATE EXTENSION UNACCENT;" | sudo su postgres -c "psql"
```

!!! warning "Change the password"
    The shipped default `sst_1234` is a placeholder. In any non-throwaway
    environment, edit the script (or create the role by hand) with a real
    password and drop it into `~/.SSTorytime` — see
    [Connection config](#connection-config).

The equivalent by hand, if you prefer:

```sh
sudo -u postgres psql <<'SQL'
CREATE ROLE sstoryline LOGIN PASSWORD 'choose-a-real-one' SUPERUSER;
CREATE DATABASE sstoryline OWNER sstoryline;
\c sstoryline
CREATE EXTENSION IF NOT EXISTS unaccent;
SQL
```

Installing the `unaccent` extension itself is a packaging step some distros
gate behind `postgresql-contrib` (Debian/Ubuntu) or `postgresqlN-contrib`
(RHEL/Fedora). If the `CREATE EXTENSION` above fails with "could not open
extension control file", install the contrib package first and retry.

## RAM disk (bulk-load experiments)

For large N4L imports where you want to measure the upper bound of ingest
speed, the repo provides a tmpfs-backed Postgres:

```sh
sudo sh contrib/ramify.sh      # mounts 800 MB tmpfs at /mnt/pg_ram
sudo sh contrib/makeramdb.sh   # initdb + pg_ctl start on the tmpfs
```

See [`contrib/ramify.sh`](https://github.com/markburgess/SSTorytime/blob/main/contrib/ramify.sh)
and [`contrib/makeramdb.sh`](https://github.com/markburgess/SSTorytime/blob/main/contrib/makeramdb.sh).

!!! danger "Data is lost on reboot"
    A tmpfs Postgres cluster vanishes when the machine reboots or when the
    filesystem is unmounted. Treat RAM mode as a disposable cache: keep your
    canonical source of truth as version-controlled `.n4l` files and rebuild
    with `N4L -wipe -u *.n4l` after each reboot. Never use RAM mode for the
    only copy of a corpus.

!!! warning "`tmpfs` is not a security boundary — pages can swap"
    "RAM-only" here is a *persistence* claim, not a secrecy one. Linux
    tmpfs pages are eligible for swap-out to the configured swap device
    under memory pressure, so cluster contents (including data in
    memory-mapped buffers) can land on the block device anyway. If your
    threat model requires data never to touch disk, either disable swap
    on the host or use a distinct encrypted swap. Otherwise treat a
    tmpfs cluster's confidentiality the same as any disk-backed cluster.

**When to use it**: comparing ingest throughput, profiling queries without
disk I/O noise, one-shot ad-hoc experiments. **When not to use it**: any
workflow where losing the DB without warning would cost more than five
minutes of re-upload.

## Connection config

The Go library's `Open()` function picks credentials in this priority order —
see [`pkg/SSTorytime/session.go:20-69`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L20-L69):

1. **`POSTGRESQL_URI` environment variable** (highest priority). If set, it is
   passed verbatim to `sql.Open()`:

    ```sh
    export POSTGRESQL_URI='postgresql://sstoryline:sst_1234@localhost:5432/sstoryline?sslmode=disable'
    ```

    Checked at [`session.go:41`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L41).

2. **`~/.SSTorytime` credentials file**. Parsed by `OverrideCredentials()` at
   [`session.go:73-133`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L73-L133).
   Simple `key: value` format, one per line:

    ```text title="~/.SSTorytime"
    dbname:    sstoryline
    user:      sstoryline
    passwd:    my-real-password
    ```

    Recognised keys are `user:`, `passwd:` (alias `password:`), and `dbname:`
    (alias `db:`). Whitespace between key and value is ignored; anything
    unrecognised is skipped.

    !!! warning "File permissions — set `chmod 600 ~/.SSTorytime`"
        The credentials file stores the database password in plaintext.
        Immediately after creating it, tighten permissions:

        ```bash
        chmod 600 ~/.SSTorytime
        ```

        The loader at
        [`session.go:85`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L85)
        does **not** check file mode — a world-readable `~/.SSTorytime`
        silently leaks credentials to every user on the host. This matters
        on shared dev boxes, CI runners, and containers with
        multi-tenant `$HOME` mounts. Enforce the mode yourself.

3. **Hardcoded defaults** (lowest priority) — fallback only, for a
   zero-config local dev box:

    ```go title="pkg/SSTorytime/session.go:27-30"
    user     = "sstoryline"
    password = "sst_1234"
    dbname   = "sstoryline"
    ```

!!! info "Connection is always over `sslmode=disable`"
    The fallback connection string is built without TLS
    ([`session.go:35`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L35)).
    If you need TLS to the database, use `POSTGRESQL_URI` and set
    `sslmode=require` (or stronger). The HTTPS server in
    `src/server/http_server.go` is a separate layer and does not affect the
    DB connection.

!!! warning "Fallback-path `sql.Open` error is swallowed"
    When `POSTGRESQL_URI` is **not** set and the hardcoded / `~/.SSTorytime`
    credentials are used, the `err` returned by
    [`sql.Open` at `session.go:44`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L44)
    is ignored. The `POSTGRESQL_URI` branch at
    [`:46-50`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L46-L50)
    checks `err` correctly; the non-env branch does not. A malformed
    driver string or an unreachable socket on the fallback path therefore
    does not fail at `Open()` — it surfaces later as a confusing error from
    the subsequent `Ping()` or from `Configure()`. If your troubleshooting
    trail leads to a "connection" error that points at a query instead of
    `Open()`, suspect this. Set `POSTGRESQL_URI` explicitly to get loud
    failures.

### Verify

A minimal smoke test from a Go program:

```go
sst := SSTorytime.Open(true)  // true = also load arrow/context directories
defer SSTorytime.Close(sst)
// If this returns without exit(-1), the connection works
// and Configure() has installed all 6 tables + ~34 stored functions.
```

`Open()` calls `Configure()`, which creates the 3 custom types, 6 tables, and
installs the ~34 PL/pgSQL functions (see
[`session.go:185-302`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L185-L302)
and the [Stored functions](Functions.md) reference). It also calls
`CREATE EXTENSION unaccent` on every open —
[`session.go:248`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L248) —
so if the extension is missing the very first query will fail loudly.

## Next steps

- [Schema](Schema.md) — the six tables and three custom types.
- [Stored functions](Functions.md) — everything `Configure()` installs.
- [Performance](Performance.md) — UNLOGGED→LOGGED lifecycle, cone limits,
  the `unaccent` dependency, RAM-disk trade-offs.

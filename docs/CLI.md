# SSTorytime CLI

The project ships **one binary**: `sstorytime`. Subcommands replace the old
multi-binary layout under `src/`.

```text
go build -o bin/sstorytime ./cmd/sstorytime
```

## Commands

| Command | Former binary | Purpose |
|---------|---------------|---------|
| `sstorytime n4l` | `N4L` | Compile / upload N4L |
| `sstorytime search` | `searchN4L` | CLI graph search |
| `sstorytime pathsolve` | `pathsolve` | Path finding + centrality |
| `sstorytime notes` | `notes` | Page-map note browser |
| `sstorytime remove` | `removeN4L` | Delete a chapter (`--force`) |
| `sstorytime graph-report` | `graph_report` | Graph analytics |
| `sstorytime text2n4l` | `text2N4L` | Intentional sentence skim |
| `sstorytime serve` | `http_server` | Web UI + JSON API |
| `sstorytime migrate` | _(new)_ | Schema migrations |
| `sstorytime examples` | API demos / pocs | list, load, run demos |

Global flags: `--database-url`, `-v` / `--verbose`.

Also accepts (in order): `DATABASE_URL`, `POSTGRES_URL`, `POSTGRESQL_URI`,
`~/.SSTorytime` credentials file, then the historical local default DSN.

## Busybox-style names

Symlink or hardlink the same binary under the old names:

```text
ln -s sstorytime N4L
ln -s sstorytime searchN4L
ln -s sstorytime pathsolve
ln -s sstorytime notes
ln -s sstorytime removeN4L
ln -s sstorytime graph_report
ln -s sstorytime text2N4L
ln -s sstorytime http_server
```

Then `N4L -u doors.n4l` is equivalent to `sstorytime n4l -u doors.n4l`.

## N4L config (arrows / annotations)

Defaults are **embedded** in the binary (`internal/sstconfig`).

To use an on-disk tree (old `./SSTconfig` behaviour):

```text
sstorytime n4l --config ./SSTconfig -u notes.n4l
```

## Web server and TLS

**Default is plain HTTP** on `--addr` (`:8080`) — reverse-proxy friendly.
The app does **not** implement ACME; terminate TLS (and ACME) on the proxy.

```text
sstorytime serve
# open http://localhost:8080
```

**Optional local HTTPS** (self-signed, classic dual-port style):

```text
sstorytime serve --tls
# https://localhost:8443  (browser will warn)
# http://localhost:8080 → redirect
```

With `--tls`, missing `cert.pem` / `key.pem` are generated via stdlib `crypto/x509`.

| Flag | Default | Meaning |
|------|---------|---------|
| `--addr` | `:8080` | HTTP app listen (or redirect listen with `--tls`) |
| `--tls` | false | Self-signed HTTPS + HTTP→HTTPS redirect |
| `--https-addr` | `:8443` | HTTPS listen (`--tls` only) |
| `--http-addr` | (equals `--addr`) | HTTP redirect listen (`--tls` only) |
| `--cert` / `--key` | `cert.pem` / `key.pem` | PEM paths (`--tls` only) |
| `--resources` | `/mnt` | Root for `/Resources/` file URLs |

## Database schema and upgrades

Schema and PL/pgSQL live in versioned migrations under
`internal/db/migrations/`. Opening a session (or `sstorytime migrate up`)
applies them via [golang-migrate](https://github.com/golang-migrate/migrate).

There is **no binary dump/export tool**. The durable source of truth is your
**N4L sources**. To rebuild a database after a breaking schema change:

```text
# optional: sstorytime remove --force "chapter" for selective delete
sstorytime n4l -u --wipe path/to/*.n4l
# or load packaged examples:
sstorytime examples load all
```

If the project follows semantic versioning for releases, treat major versions
as allowed to break on-disk layout; re-upload from N4L after upgrading.

## text2n4l percent flag

```text
sstorytime text2n4l --percent 30 file.txt
sstorytime text2n4l -% 30 file.txt          # upstream alias
```

## pathsolve Dirac notation

```text
sstorytime pathsolve --begin A1 --end B6
sstorytime pathsolve '<B6|A1>'
sstorytime pathsolve '<B6|context tags|A1>'
```

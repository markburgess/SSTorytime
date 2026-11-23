# Persistent Docker environment (CLI + HTTPS web)

A local Linux environment for hands-on SSTorytime work. It runs PostgreSQL 17, a
workspace container with all SSTorytime tools built from this checkout, and the
web UI at **https://localhost:8443**. It is **persistent**:
your data survives stop/start, and nothing is ever deleted unless you ask for
it explicitly. It is a development environment only, not a deployment.

## Requirements

- **A copy of this repository:** `git clone https://github.com/markburgess/SSTorytime`,
  or download the ZIP from GitHub and unpack it.
- **Docker with Compose v2** (the `docker compose` command), e.g. Docker Desktop.

That's all: no Go, PostgreSQL or Make on your computer. More in
[Platform details](#platform-details).

## First steps

From the repository root:

```sh
docker compose -f docker/compose.yaml run --rm start
```

The first run takes a minute or two: it downloads images, builds the tools,
sets up the database and loads one small example (`examples/doors.n4l`).

1. Open **https://localhost:8443**. Type `https://`: there is no plain `http://`
   address.
2. By default the certificate is self-signed, so the browser warns about it.
   This is expected for this local address: choose **Advanced**, then
   **Proceed to localhost (unsafe)** in Chrome, **Continue to localhost
   (unsafe)** in Edge, or **Accept the Risk and Continue** in Firefox (wording
   may vary by version). Accept it only for `localhost`. **To avoid the
   warning entirely**, put an mkcert certificate into `docker/tls/` (see
   [docker/tls/README.md](tls/README.md)); start picks it up automatically.
3. Search for **`door`**. Right after setup only this small example is loaded,
   so other searches find nothing yet. To follow the [Tutorial](../docs/Tutorial.md),
   load its examples (see [Loading the tutorial examples](#loading-the-tutorial-examples)).

The command line works too:

```sh
docker compose -f docker/compose.yaml exec workspace bash
searchN4L door                       # inside the workspace shell
```

**After restarting your computer or Docker,** the environment does not start by
itself: run the `start` command above again. Your data is kept.

**Before shutting down your computer,** stop the environment:
`docker compose -f docker/compose.yaml stop`. SSTorytime's graph tables are
emptied if the database is not shut down cleanly (see
[Stop, resume, clean up](#stop-resume-clean-up)). Should that happen, reload
your `.n4l` files.

## Commands

Run them from the repository root.

| What | Command |
|---|---|
| provision / resume | `docker compose -f docker/compose.yaml run --rm start` |
| shell in the workspace | `docker compose -f docker/compose.yaml exec workspace bash` |
| original baseline check | `docker compose -f docker/compose.yaml run --rm verify` |
| rebuild (all, or only some tools) | `docker compose -f docker/compose.yaml run --rm rebuild [tool ...]` |
| rebuild on save | `docker compose -f docker/compose.yaml run --rm watch` |
| state of everything, read-only | `docker compose -f docker/compose.yaml run --rm inspect` |
| web server logs | `docker compose -f docker/compose.yaml logs -f web` |
| stop (keeps everything) | `docker compose -f docker/compose.yaml stop` |
| quick resume without checks | `docker compose -f docker/compose.yaml up -d --wait` |

`start` brings up `db`, `workspace` and `web`, then runs every step in the
foreground. Each stage prints `[stage]`, and a failure names the stage, exits
nonzero and leaves all containers and volumes in place for diagnosis.

In the workspace shell, all tools are on the search path, and the checkout is
mounted **read-only** at `/src`. Use `/work/run` for your own scratch files.

| Tool | Purpose (see [docs/README.md](../docs/README.md#the-tools)) |
|---|---|
| `N4L` | compile and upload `.n4l` notes |
| `searchN4L` | search the graph from the command line |
| `http_server` | the web UI (runs in the `web` service) |
| `text2N4L` | turn a text file into N4L notes for further editing |
| `removeN4L` | remove an uploaded chapter |
| `notes` | browse notes in page layout |
| `pathsolve`, `graph_report` | path solving; graph analysis (loops, sources, sinks, centrality) |
| `API_EXAMPLE_1` … `API_EXAMPLE_4` | programming examples for the Go API |

`cmd/demo_pocs` (internal test programs) is not built.

## Using a different host port

The web UI uses host port 8443 unless you set `SST_WEB_PORT`, e.g. when 8443 is
already taken:

```sh
SST_WEB_PORT=9443 docker compose -f docker/compose.yaml run --rm start       # bash, zsh
```

```powershell
$env:SST_WEB_PORT = "9443"; docker compose -f docker/compose.yaml run --rm start   # PowerShell
```

The address is then `https://localhost:9443`. Use the same value for every
command against that environment. Without it, Compose falls back to 8443 and
recreates `web` on that port; the data is not affected.

## What start does

| Step | First start | Later starts |
|---|---|---|
| start `db`, `workspace`, `web` | yes | if stopped |
| build binaries into `/work/bin` | yes | only if missing (otherwise use `rebuild`) |
| import `examples/doors.n4l` (`N4L -u -wipe`) | only into a confirmed-fresh database | **never** |
| TLS certificate | mkcert files from `docker/tls/` if present (validated), else self-signed, generated once | the same rule; self-signed is reused, never overwritten |
| web server | starts once binary, certificate and initialized database exist | reloads if the binary or certificate changed |
| readiness (web answers, its database-backed search works) | yes | yes |
| baseline check | yes | automatically until it has passed once; afterwards via `verify` |
| delete data | never | never |

The web service waits until the database carries this environment's
initialization marker. `http_server` alters the schema of whatever database it
opens, so it never starts against an unrecognized one.

## verify: the original baseline

This checks that the original `doors.n4l` example still answers in two ways:

- `searchN4L door` lists `(leads to) - passage`.
- `POST https://localhost:8443/searchN4L name=door` returns HTTP 200 with
  `Response == "Orbits"` and `passage` in the content. The request trusts
  exactly the active certificate authority, never "any certificate".

It runs no `-wipe`, no import and no upload, stops no containers and changes no
volumes. The application itself runs idempotent schema setup (such as
`CREATE OR REPLACE FUNCTION`) whenever it opens the database. If you rebuilt
the graph without `doors.n4l` (see below), this check fails and says so. The
environment can still be perfectly healthy.

## Loading data

N4L builds the graph from `.n4l` files. **The normal way to change data is to
rebuild the whole graph: all files in one `N4L -u -wipe` call.** The compiler
resolves references between files only when they are loaded together; rebuilding
is fast. `-wipe` replaces everything in the graph, so always list **all** files
you want to keep.

### Loading the tutorial examples

This is the Docker equivalent of the Tutorial's `cd examples; make` plus
`LoopyLoo.n4l`. First open the workspace shell (from the repository root):

```sh
docker compose -f docker/compose.yaml exec workspace bash
```

Then paste this **into the workspace shell**, not into PowerShell or cmd:

```sh
cd /src/examples
N4L -u -wipe doors.n4l Mary.n4l chinese*n4l branches.n4l doubleslit.n4l \
  ConstructionProcesses.n4l wardleymap.n4l brains.n4l kubernetes.n4l SSTorytime.n4l \
  SSTsearchexamples.n4l integral.n4l reasoning.n4l moon.n4l astronomy.n4l promises.n4l \
  json_example_1.n4l json_example_2.n4l NDA.n4l reminders.n4l LoopyLoo.n4l
```

Then the Tutorial's searches work, e.g. `searchN4L please in chinese`. Don't run
`make` in `examples/`: it tries to build into the read-only checkout.

Like every `-wipe` call, this **replaces the whole graph**, including files of
your own that you loaded earlier. To keep them, add them to the same call (see
the next section).

### Your own files

Keep your `.n4l` files in the checkout (they are visible at `/src`) and rebuild
with all of them in one call. Including `doors.n4l` keeps `verify` passing:

```sh
cd /work/run
N4L -u -wipe /src/examples/doors.n4l /src/path/to/notes.n4l /src/path/to/more.n4l
```

This keeps **only** the files listed: tutorial examples loaded earlier are gone
afterwards. To keep them too, use the command from
[Loading the tutorial examples](#loading-the-tutorial-examples) (run in
`/src/examples`) and append your files to the end of it, e.g.
`... reminders.n4l LoopyLoo.n4l /src/path/to/notes.n4l`.

To check a file without uploading it, leave out `-u`: `N4L /src/path/to/notes.n4l`.

### Appending without -wipe: only for new, independent chapters

`N4L -u` without `-wipe` adds to the existing graph, but only safely in one case:

| Situation | What N4L does |
|---|---|
| new chapter, no node names in common with the database | imported correctly |
| a chapter that already exists (e.g. you edited the file and load it again) | **refuses the whole upload** ("Database already contains a chapter … might corrupt the data") and exits **without an error code**: your edit is silently not applied |
| new chapter that reuses existing node names | imported, but those nodes now exist **twice** ("Matches for … already exist … reload all overlapping segments together") |

So after editing a file, rebuild with `-wipe` as shown above. `-force` exists,
but N4L itself warns that it might corrupt the data.

## Inspecting the environment

Three read-only views answer the questions that otherwise take several
commands. None of them changes data.

```sh
docker compose -f docker/compose.yaml run --rm inspect                       # everything on one screen
docker compose -f docker/compose.yaml run --rm inspect duplicates            # nodes registered twice
docker compose -f docker/compose.yaml run --rm inspect search "(zai chengshi)"   # where a search gets lost
```

**`inspect`** shows the initialization state, the chapters with node counts
(marking the one `verify` needs), duplicates, whether the tools are built and
up to date with the source (judged by timestamps), whether the web server
answers and runs the current binary, and the active certificate:

```
Database
  initialized   seeded 2026-11-23 10:07 (doors.n4l) · baseline verified 2026-11-23 10:07
  graph         1655 nodes in 3 chapters  (+4 empty placeholder nodes)
                   1640  AIP Research Landscape
                     13  multi slit interference   <- needed by verify
  duplicates    none
Tools
  12 of 12 built · source unchanged since the last build   -> up to date (by timestamps)
Web
  https://localhost:8443 · answering · running the current http_server and certificate
```

The empty placeholder nodes at `(0,0)` are internal to SSTorytime (one more
appeared with each upload in tests); they are not counted as data or duplicates.

**`inspect duplicates`** lists node texts that exist more than once, with their
chapters: the typical result of appending a chapter without `-wipe` (see
[Loading data](#loading-data)). The same text in different chapters can also be
intended, so check the list.

**`inspect search "<term>"`** compares, for one search term, what the database
matches with what `searchN4L` and the web search return, so you can tell
missing data from a search that loses the term:

```
Term "(zai chengshi)"  -> parentheses = accent-insensitive search
  database, accent-insensitive           2   zài chéngshì lǐ · zài chéngshì de xīběi lǐmiàn
  database, full-text unaccented         2
  searchN4L (command line)               0   <- term arrives empty ("Intentional:" is blank)
  web search (/searchN4L)                0
Diagnosis: the data is there and the database finds it, but both the command-line and the
           web search return nothing: the term gets lost in the search -> application, not your data.
```

The search view runs `searchN4L` and the web search like a normal search; like
`verify`, that only triggers the application's usual schema setup.

## Development workflow

Edit and use git on the host; `/src` is read-only in the containers.

- **`watch`** rebuilds on save; stop it with Ctrl-C. It polls file timestamps
  every second, because Docker Desktop and WSL2 don't reliably pass
  file-change notifications into containers. That also works for checkouts on
  a Windows drive (`C:`). It notices edited, new, deleted and renamed files
  (shown as `~`, `+` and `-`) and rebuilds only what's affected:

  | Change in | Rebuilds |
  |---|---|
  | `pkg/`, `go.mod`, `go.sum` | all tools |
  | `cmd/<tool>`, e.g. `cmd/N4L` | that tool only |
  | `cmd/server`, including `public/` HTML/CSS/JS (compiled into the binary) | `http_server` |

- **`rebuild`** does the same by hand, e.g. `run --rm rebuild` for all tools
  (about 3 s) or `run --rm rebuild searchN4L notes` for some (about 2 s).
- The web service runs `http_server` under `docker/web-supervisor.sh`, which
  restarts it when a new binary is published. The changed page is served about
  2–3 s after a save. Reload the browser afterwards.
- A compile error prints the Go error, publishes nothing and leaves the running
  binaries and server untouched. Fix the error and save again.
- If the terminal running `watch` is closed instead of stopped with Ctrl-C, the
  `watch` container may keep running. Remove it with
  `docker compose -f docker/compose.yaml --profile tools down --remove-orphans`;
  without `-v` this keeps the data. Don't run two watchers at once.

## Stop, resume, clean up

```sh
docker compose -f docker/compose.yaml stop     # stop; keeps everything
docker compose -f docker/compose.yaml run --rm start   # resume
docker compose -f docker/compose.yaml down     # remove containers; keeps volumes
docker compose -f docker/compose.yaml down -v  # FULL RESET: deletes all data + certificates
```

Nothing is ever cleaned up automatically. Only `down -v` removes the volumes
`db_data` (the database) and `work` (binaries, certificates, caches). Your
files in `docker/tls/` and the checkout are never touched.

**Important:** SSTorytime keeps its graph in PostgreSQL `UNLOGGED` tables, and
PostgreSQL **empties them after an unclean shutdown**. This includes a killed
container, Docker Desktop crashing, or a hard `wsl --shutdown`. Prefer
`docker compose ... stop`; the `db` service gets 120 s to shut down cleanly.
Since your `.n4l` files are the source, you can always rebuild the graph.

## Platform details

- **Docker version:** Compose v2 (`docker compose`) is required; the old
  `docker-compose` (v1) is not supported. Docker Engine 25 or newer is needed
  for a healthcheck setting; older versions reject the Compose file.
- **Nothing else on the host:** no Go, PostgreSQL, Make, OpenSSL or POSIX
  shell. The commands are the same on every platform.
- **Tested on:** Windows with Docker Desktop, from PowerShell and from WSL2
  (Ubuntu). Native Linux and macOS are expected to work but have not been
  tested yet.
- **Network:** the first start downloads images and Go modules.
- **Windows checkouts:** `.gitattributes` keeps `docker/*.sh` at LF line
  endings, since CRLF would break them inside the Linux containers. If you
  copied the files some other way and see `set: Illegal option -`, convert them
  to LF.

## Isolation

The Compose project is `sstorytime-install-test`, a name kept from the first
version so that existing environments keep their volumes. It has its own
network and volumes. PostgreSQL is not published to the host; only the web UI
is, on `127.0.0.1` (port 8443 or `SST_WEB_PORT`). This environment does not
touch `postgres-docker/` or a host PostgreSQL. The containers get their DB
connection from `POSTGRESQL_URI` and their config from
`SST_CONFIG_PATH=/src/SSTconfig`.

## Testing a fresh installation (maintainers / CI)

```sh
./docker/test-fresh-install.sh          # --keep leaves the test project for inspection
```

This provisions a brand-new, isolated Compose project with its own volumes,
network and host port (`SST_TEST_PORT`, default 18443). It runs alongside an
existing environment without touching it, and checks three things:

1. **First start:** build, one-time fixture import, markers, certificate, web
   published on loopback only, and a host HTTPS request with a verified
   certificate.
2. **Verification on demand:** CLI and HTTPS baseline, with the node count
   unchanged.
3. **Own data, stop, resume:** no re-import or rebuild, and the data, marker
   and certificate are retained. The data is searchable via CLI and HTTPS
   afterwards.

It removes only its own project at the end, also after a failure, and exits
nonzero if any check fails. Unlike normal use it needs a POSIX shell and the
Docker CLI on the host, because it drives Compose from the outside.

## Troubleshooting

| Message | Meaning / fix |
|---|---|
| `service "workspace" is not running` (or the browser can't connect) | The environment is stopped, e.g. after restarting your computer or Docker. Run `docker compose -f docker/compose.yaml run --rm start`; your data is kept. |
| browser shows an HTTP 400 error, e.g. "Client sent an HTTP request to an HTTPS server" | You opened `http://`. Use `https://localhost:8443` (or your `SST_WEB_PORT`). |
| `Bind for 127.0.0.1:8443 failed: port is already allocated` | Another process holds the port. Free it, or use another one (see [Using a different host port](#using-a-different-host-port)), then follow the next row. |
| `the web container is not on the Compose network` | Docker left web without a network because it was first created while the port was busy. Run `docker compose -f docker/compose.yaml up -d --force-recreate web` (no data lost), then `start` again. |
| `previous initialization did not complete (fixture=seeding)` | The first import was interrupted or could not be confirmed. Nothing was wiped. Inspect with `exec db psql -U sstoryline`, or start over with `down -v`. |
| `unrecognized database` | The database has tables but no install marker. start refuses to seed it, and the web server won't open it. |
| `invalid marker` | The `sst_install_test.state` table is missing or holds an unknown state. Nothing was changed. |
| `go build ... failed` | Fix the source and run `rebuild`. The previous binaries are still in use. |
| `searchN4L: command not found` in the workspace | You opened a login shell (`bash -l`), which resets the search path. Use `exec workspace bash`, or call `/work/bin/searchN4L`. |
| N4L: `Database already contains a chapter` | Nothing was uploaded. Rebuild with all files and `-wipe` (see [Loading data](#loading-data)). |
| `docker/tls/...` (mkcert: expired, not for localhost, not signed by rootCA.pem, incomplete) | Regenerate as described in `docker/tls/README.md`. The server keeps the last good certificate meanwhile. |
| `rootCA-key.pem found` | mkcert's CA private key is in the checkout. Delete it there; it belongs only in `mkcert -CAROOT`. |
| self-signed `expired` / `do not match` / only one file | In the workspace, `rm -r /work/tls/self`, then run start again to generate a new one. |
| `search endpoint did not answer` | Web is up but cannot use the DB. See `logs web`. |
| web log: `waiting for setup` | Normal before the first `start`. Run `start`. |

Initialization state lives in the database itself, in table
`sst_install_test.state` (`fixture=seeding|seeded`, `baseline=verified`), so
`down -v` resets it together with the data.

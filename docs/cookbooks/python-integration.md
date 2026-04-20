# SSTorytime from Python

SSTorytime ships a Python module at [`src/SSTorytime.py`](https://github.com/markburgess/SSTorytime/blob/main/src/SSTorytime.py) that mirrors a subset of the Go library. This cookbook walks through using it to insert nodes, connect them, and run a cone search — end-to-end, under 50 lines of Python.

!!! info "Prerequisites"
    - Python 3.10+ (the module uses `match`/`case`).
    - `pip install psycopg2-binary` (or system `psycopg2`).
    - PostgreSQL with the SSTorytime schema and at least the default arrow directory loaded. Run `make db` at the repo root once, then `../src/bin/N4L -u examples/doors.n4l` (or any N4L file) to populate arrows.
    - Either run your script from `src/` (the module sits there) or add `src/` to `PYTHONPATH`.

## 1. Connect

Credentials and DSN defaults match the hardcoded values in Go — user `sstoryline`, password `sst_1234`, database `sstoryline`, host `localhost`, port `5432`.

```python
import SSTorytime as SST

ok, sst = SST.Open("sstoryline", "sst_1234", "sstoryline", "localhost")
if not ok:
    raise SystemExit("could not open SSTorytime DB")
```

`Open` downloads the arrow directory and the inverse-arrow table into module-level globals (`ARROW_DIRECTORY`, `INVERSE_ARROWS`). Subsequent calls use those to translate arrow short names into STtype channels, so your first call is a little slow and later calls are fast.

!!! warning "Override the DSN for production"
    The hardcoded credentials are fine for local development. In production, read them from
    environment variables or a config file and pass them to `Open`. The `psycopg2.connect`
    call inside `Open` takes the usual libpq keywords.

## 2. Create and link nodes

```python
chapter = "python cookbook"
context = ["demo", "python"]

v1 = SST.Vertex(sst, "First Python node", chapter)
v2 = SST.Vertex(sst, "Second Python node", chapter)
v3 = SST.Vertex(sst, "Third Python node", chapter)

SST.Edge(sst, v1, "then", v2, context, 1.0)
SST.Edge(sst, v2, "then", v3, context, 1.0)

print("v1 NPtr:", v1.NPtr if hasattr(v1, "NPtr") else v1)
print("v2 NPtr:", v2.NPtr if hasattr(v2, "NPtr") else v2)
```

Like the Go API, `Vertex` and `Edge` are idempotent. Running this block twice does not duplicate anything.

The arrow short name passed to `Edge` (`"then"`) must already exist in the arrow directory. To define new arrows, use N4L — the Python module does not (yet) expose arrow creation.

## 3. Read nodes back

```python
fetch1 = SST.GetDBNodeByNodePtr(sst, v1)
print("Node v1 (text, chapter, class, cptr):", fetch1)

fetch2 = SST.GetDBNodeByNodePtr(sst, v2)
print("Node v2:", fetch2)
```

The returned tuple matches the layout in [`SSTorytime.py`](https://github.com/markburgess/SSTorytime/blob/main/src/SSTorytime.py); expect `(S, Chap, Class, CPtr)` or similar depending on the version.

## 4. Cone search

The simplest path query — forward [cone](../concepts/glossary.md#cone--cone-search) starting at a [NodePtr](../concepts/glossary.md#nodeptr), following a single STtype.

```python
leadsto   = 1       # +leads-to; see STtype table in docs/graph_report.md
depth     = 5       # max hops
maxresult = 100     # cap on paths returned

link_paths, _ = SST.GetFwdPathsAsLinks(sst, "(1,0)", leadsto, depth, maxresult)

for path in link_paths:
    print("Path:", end=" ")
    for link in path:
        node = SST.GetDBNodeByNodePtr(sst, link[3])
        print(f"{link[3]}={node[0]}", end=" -> ")
    print()
```

The NodePtr `"(1,0)"` is a string literal; change it to one of the NPtr values printed in step 3 to anchor the cone on your own data.

## 5. Constrained multi-context cone search

The more capable search function takes a context filter, a chapter filter, and a direction:

```python
context   = ["path"]
startset  = [str(v1.NPtr)] if hasattr(v1, "NPtr") else ["(1,0)"]

paths, _ = SST.GetEntireNCConePathsAsLinks(
    sst,
    "fwd",          # direction: fwd or bwd
    startset,       # list of NPtr strings
    10,             # max depth
    "python",       # chapter substring
    context,
    30,             # result limit
)

for path in paths:
    print("Path:", end=" ")
    for link in path:
        node = SST.GetDBNodeByNodePtr(sst, link[3])
        print(f"{link[3]}={node[0]}", end=" -> ")
    print()
```

Compare the output to running the equivalent shell:

```bash
./src/bin/searchN4L "\\from \"First Python\" \\depth 10 \\context path"
```

## 6. Close the connection

```python
SST.Close(sst)
```

Not strictly required — `psycopg2` closes on program exit — but good practice inside longer-running scripts.

## Full runnable script

The repo ships a working reference implementation at [`src/python_integration_example.py`](https://github.com/markburgess/SSTorytime/blob/main/src/python_integration_example.py). Run it directly:

```bash
cd src
python3 python_integration_example.py
```

It covers: connect, insert with context, fetch node, cone search, multi-context search, close — the same flow as this cookbook, against the default sample data.

## Differences from the Go library

The Python module (`src/SSTorytime.py`) is a deliberate subset of the Go
library. Each bullet below names a specific gap and tells you what to
reach for instead. "Drop to psycopg2" means call the PL/pgSQL stored
functions directly from a raw cursor; "use the Go HTTP API" means POST to
`/searchN4L` on `:8443` via the `requests` library (see
[WebAPI.md](../WebAPI.md) for the wire shape).

- **No `HubJoin`.** Hub-pattern inserts are not exposed. Do them through
  N4L (recommended) or a Go helper. Do not try to fake it from Python by
  calling `IdempDBAddNode` — see the next bullet.
- **No `GetPathsAndSymmetries`.** The high-level path solver is Go-only.
  For simple cone walks use `GetFwdPathsAsLinks`/`GetEntireNCConePathsAsLinks`
  (covered above); for full path-solve go via the Go HTTP API with a
  `\from … \to …` query.
- **No `WaveFrontsOverlap`.** The path-integral-style wave intersection
  is Go-only. Same workaround: use `/searchN4L` with `\from … \to …`.
- **No `GetDBArrowByPtr`.** Reverse lookup from an `ArrowPtr` to its
  short/long name and STtype is not wrapped. Drop to psycopg2 and
  `SELECT * FROM arrowdirectory WHERE arrptr = $1`.
- **No `/Upload` or `/SearchAssets` wrappers.** Asset attachment is not
  exposed in the Python module. Use the Go HTTP API via `requests`
  (POST `/Upload` or `/SearchAssets`); the wire protocol is documented
  in [WebAPI.md](../WebAPI.md).
- **No idempotent mock layer.** `Node.IdempDBAddNode` in the Python
  source returns a dummy `(1,2)` tuple — it is a stub, not the real
  insertion path. Real inserts go through `Vertex()`, which runs direct
  SQL. Do not build hub-style logic on top of the stub.
- **String NPtrs.** NodePtrs round-trip as strings (`"(1,0)"`), not
  structs. The Go code uses the `NodePtr` type. Parse before comparing.
- **Subset coverage more broadly.** When a function you need is not in
  the Python module, the two escape hatches are (a) drop to psycopg2
  and call the PL/pgSQL stored functions directly, or (b) use the Go
  HTTP API via `requests` and parse the JSON envelope.

## Troubleshooting

- **`ImportError: No module named SSTorytime`** — run the script from `src/`, or add it to `PYTHONPATH`: `PYTHONPATH=src python3 my_script.py`.
- **`psycopg2.OperationalError: connection refused`** — PostgreSQL is not listening on `localhost:5432`. Check with `pg_isready`.
- **`FATAL: role "sstoryline" does not exist`** — run `make db` from the repo root (or `sh contrib/makedb.sh`) to create the SSTorytime role and database.

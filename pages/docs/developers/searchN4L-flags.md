# searchN4L — query DSL reference

Full flag and query-DSL reference for `searchN4L`. This page is kept in
the repo for contributors. The user-facing page — "the shape of a
question" — lives at `../searchN4L.md`.

## Command line

```
searchN4L [-v] <query>
```

- `-v` — verbose mode; prints the parsed `SearchParameters` struct and diagnostic detail. See [`src/searchN4L/searchN4L.go:67-97`](https://github.com/markburgess/SSTorytime/blob/main/src/searchN4L/searchN4L.go#L67-L97).

Everything else is expressed through the query DSL below. Multiple arguments on the command line are joined into a single query string before parsing.

## The query DSL

Queries are parsed by `DecodeSearchField` in [`pkg/SSTorytime/service_search_cmd.go:113-192`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go#L113-L192). Commands begin with a backslash; on the shell command line escape it as `\\` or quote the token.

The tool recognises a number of words. This is a mixed blessing — the parser may misread a search word as a command. Quote search terms (`"..."`) to protect them.

### Command reference

All commands are defined as constants in [`pkg/SSTorytime/service_search_cmd.go:50-107`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go#L50-L107). Short aliases (without backslash) are accepted where unambiguous.

| Command | Short aliases | Meaning | Example |
|---|---|---|---|
| `\on` | `on` | Subject of the search (synonym of plain search term) | `\on brain` |
| `\for` | `for` | Same as `\on` | `\for lamb` |
| `\about` | | Same as `\on` | `\about fox` |
| `\notes` / `\browse` | | Show notes in original input order for a chapter | `\notes brain` |
| `\page <N>` | | Pagination for `\notes` | `\notes brain \page 3` |
| `\path` / `\paths` | | Search for paths between nodes | `\path \from a1 \to b6` |
| `\from` | | Start node(s) for a path | `\from start` |
| `\to` | `to` | End node(s) for a path | `\to "target 1"` |
| `\seq` / `\sequence` | | Sequence mode (stories linked by `then`) | `\seq "Mary had"` |
| `\story` / `\stories` | | Same as `\seq` | `\stories fox` |
| `\context` / `\ctx` | | Context filter (scopes which contexts to match) | `\context restaurant` |
| `\as` | `as` | Same as `\context` (role/classifier) | `\as brain` |
| `\chapter` / `\section` / `\in` | `in` | Scope to a chapter | `\chapter "chinese"` |
| `\contents` / `\toc` / `\map` | `toc` | Table of contents | `\toc any` |
| `\arrow` / `\arrows` | | Arrow introspection (by name, number, or STtype) | `\arrow ph,pe` |
| `\limit` / `\range` / `\depth` / `\distance` `<N>` | | All four aliases set `param.Range` — cap on results / path length / search-cone radius. See [`service_search_cmd.go:270`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go#L270). | `\depth 16` |
| `\min <N>` / `\atleast <N>` / `\gt <N>` | | Minimum path length | `\min 2` |
| `\max <N>` / `\atmost <N>` / `\lt <N>` | | Maximum arrow or path length | `\max 8` |
| `\stats` | `stats` | Print graph statistics | `\stats \in brain` |
| `\finds` / `\finding` | | What to find on an orbit | `\finds fleece` |
| `\remind` | | Use recent-activity filtering | `\remind` |
| `\new` | | Restrict to items last-seen within 4 hours (`RECENT = 4`, see [`service_search_cmd.go:411-413`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go#L411-L413)) | `\new` |
| `\never` | | Disable the recent-activity horizon (`Horizon = NEVER = -1` — unbounded, not "items never seen"; see [`service_search_cmd.go:414-416`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go#L414-L416)) | `\never` |
| `\help` | `help` | Show help chapter | `\help` |

### Literal and precise matches

- Bang/pling quotes — `!term!` or `|term|` — force an **exact** string match. Useful for tokens that are common substrings (e.g. `!A!` to find just `A` rather than every string containing `A`).
- Parenthesised, unaccented terms — `"(fangzi)"` — search an **unaccented** tsvector column. This lets you find `fángzǐ` without typing the accents. See [`pkg/SSTorytime/postgres_types_functions.go`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go) for the `UnSearch` column.
- `NodePtr` references — `"(1,1)"` — match a node by its `(Class, CPtr)` address directly. See `IsLiteralNptr` in [`service_search_cmd.go`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go).
- Dirac notation — `"<end|start>"` — path from `start` to `end`. Dispatched at [`service_search_cmd.go:178-189`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go#L178-L189) inside `DecodeSearchField`; the `DiracNotation` function itself lives in [`pkg/SSTorytime/tools.go:523`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/tools.go#L523). Order is **end first, start second**.

### Word-level and ts_vector search

Plain substrings are the default. For exact-word, multi-word, logical, and
prefix queries the underlying Postgres `ts_vector` operators are exposed:

```
!a1!                         exact whole-node match
|a1|                         equivalent
"|deep purple|"              exact multi-word match (quote the whole thing)
strange<->kind<->of<->woman  neighbouring lexemes (ts_vector)
strange<2>woman              skip two lexemes
"a1 & !b6"                   AND + NOT (quoted so ! is not delimiter)
pink<->flo:*                 prefix completion
```

`ts_vector` ignores stopwords (`a`, `in`, `of`, …) and works only with the
English dictionary for now.

### Why `--` before `\arrow -2`?

Go's `flag` package treats any token starting with `-` as a potential
flag. Without `--`, `searchN4L \arrow -2` is rejected because `-2` is
read as an undefined flag. The bare `--` sentinel tells the flag parser
"everything after this is a positional argument," so `-2` reaches the
query DSL intact.

## Exit codes

- **`0`** — success (may print zero results if nothing matches).
- **`2`** — usage error (e.g. `-h` / invalid flag). `Usage()` calls `os.Exit(2)` at [`src/searchN4L/searchN4L.go:137`](https://github.com/markburgess/SSTorytime/blob/main/src/searchN4L/searchN4L.go#L137).
- **`-1`** — any other error (e.g. DB unreachable, malformed Dirac notation).

## Environment variables

- `POSTGRESQL_URI` — overrides the hardcoded DSN in [`pkg/SSTorytime/session.go:41`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L41). Point `searchN4L` at a remote or alternate instance.
- `SST_CONFIG_PATH` — location of `SSTconfig/` arrow definitions. Not required at query time (arrows are loaded from the DB), but used if the CLI needs to fall back to file-based lookup.

!!! warning "Database must be reachable"
    `searchN4L` connects on every invocation. If PostgreSQL is down, the tool prints a
    connection error and exits with `-1`. Use `pg_isready` or `psql -c 'select 1'` to verify
    the database is up before troubleshooting query-DSL issues.

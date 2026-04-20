# Performance and operational notes

This page catalogues the hardcoded thresholds, lifecycle phases, and
extension dependencies that shape SSTorytime's database performance. Most
of the numbers here are **not tuneable at runtime** — they are source-level
constants. If you are profiling a workload and one of them is a bottleneck,
you will be recompiling.

## UNLOGGED → LOGGED lifecycle

See the [state diagram in Schema.md](Schema.md#unlogged-logged-bulk-load-lifecycle)
for the full picture. The short version:

- **During bulk upload**: `Node`, `PageMap`, `ArrowDirectory`, and
  `ArrowInverses` are declared `UNLOGGED`
  ([`postgres_types_functions.go:32`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L32),
  [`:50`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L50),
  [`:59`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L59),
  [`:67`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L67)).
  Skipping WAL makes ingest 2-3× faster on typical hardware but means an
  unclean Postgres shutdown **truncates the tables to zero**.
- **After upload**: `GraphToDB()` calls
  `ALTER TABLE Node SET LOGGED` and `ALTER TABLE PageMap SET LOGGED`
  ([`db_upload.go:119-120`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go#L119-L120)).
  `ArrowDirectory` and `ArrowInverses` are **not** flipped — they are
  rebuilt from the in-memory cache on the next `Configure()`, so their
  durability does not matter.

!!! warning "Crash during upload → re-run the uploader"
    If `N4L -u` crashes (SIGKILL, OOM, Postgres crash) before
    `db_upload.go:119` runs, `Node` and `PageMap` stay `UNLOGGED`. Any
    subsequent Postgres restart that is not strictly clean (including
    `docker compose down` without `-f`) will truncate them. Recovery is
    always `N4L -wipe -u *.n4l`. This is fine provided your `.n4l` sources
    are in version control — **treat the database as a cache, never as
    source of truth**.

## `CAUSAL_CONE_MAXLIMIT` — cone cardinality cap

```go title="globals.go:29"
CAUSAL_CONE_MAXLIMIT = 100
```

Every cone and path search in the library threads a `maxlimit` parameter
through — fan-out per hop is capped at this value. In practice the HTTP
server and CLIs pass `CAUSAL_CONE_MAXLIMIT` unless the caller overrides it.

**Why a cap at all?** The PL/pgSQL path walkers recurse; without a
cardinality bound a depth-8 walk in a densely connected graph explodes
into megarows of intermediate text. `SumConstraintPaths` and its siblings
implement a dynamic-budget refinement on top:

```text
horizon := maxlimit - array_length(fwdlinks, 1);
IF horizon < 0 THEN
  horizon = 0;
  maxdepth = depth + 1;   -- force termination next level
END IF;
```

(See
[`postgres_types_functions.go:1275-1280`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L1275-L1280).)
So the effective budget shrinks with each hop, and once it goes negative
the walker cuts the search at the next level rather than risking runaway.

**Tuning**: bump the constant in `globals.go` and rebuild. Realistic ceiling
before paths get unreadable is around 500-1000; at that point the returned
text blob itself becomes a bottleneck (one big text string per call,
parsed client-side).

## Dual tsvector strategy

The `Node` table carries two `tsvector` columns, both
`GENERATED ALWAYS AS STORED`:

```sql title="postgres_types_functions.go:37-38"
Search    TSVECTOR GENERATED ALWAYS AS
            (to_tsvector('english', S)) STORED,
UnSearch  TSVECTOR GENERATED ALWAYS AS
            (to_tsvector('english', sst_unaccent(S))) STORED,
```

- `Search` preserves accents — a query for "café" matches only text with
  the accented form.
- `UnSearch` strips accents before tokenising — a query for "cafe" matches
  "café", "Café", "CAFE", and vice versa.

Each has its own [GIN index](Indexes.md): `sst_gin` over `Search`, `sst_ungin`
over `UnSearch`. Storage cost is roughly 2× the textual index footprint —
acceptable for the query power gained.

**Why not pick one?** Accent-preserving is correct for proper nouns and
disambiguation ("Peña" vs "Pena"). Accent-stripping is correct for casual
English-language search where typing accents is an accessibility burden.
Shipping both lets every caller pick without schema migration.

## The `unaccent` dependency

`sst_unaccent` is a thin immutable wrapper around PostgreSQL's
`unaccent()` extension function:

```sql title="postgres_types_functions.go:1747-1755"
CREATE OR REPLACE FUNCTION sst_unaccent(this text) RETURNS text AS $fn$
DECLARE
  s text;
BEGIN
  s = unaccent(this);
  RETURN s;
END;
$fn$ LANGUAGE plpgsql IMMUTABLE;
```

The extension is required — there is no fallback. It is loaded from three
different places so every supported deployment path has it by the time
`Configure()` needs it:

1. **Docker**: [`postgres-docker/init-db.sql:5`](https://github.com/markburgess/SSTorytime/blob/main/postgres-docker/init-db.sql#L5)
   runs on first container start.
2. **Manual install**: [`contrib/makedb.sh:14`](https://github.com/markburgess/SSTorytime/blob/main/contrib/makedb.sh#L14)
   includes `CREATE EXTENSION UNACCENT`.
3. **Every `Open()` call**: as a belt-and-braces measure,
   `Configure()` issues `CREATE EXTENSION unaccent` on every connection
   ([`session.go:248`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L248)).
   Postgres treats this as a no-op if the extension is already present.

### Failure modes if `unaccent` is missing

If the `unaccent` extension is not installed at the **cluster** level
(i.e., the files in `$PGSHARE/extension/` are absent — typically because
the `-contrib` package wasn't installed):

- **Symptom 1**: `Open()` → `Configure()` issues
  `CREATE EXTENSION unaccent` and it fails silently in the current code
  (errors from `QueryRow` are not checked at
  [`session.go:248`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L248)).
- **Symptom 2**: The first time `Node` is created, PostgreSQL tries to
  evaluate the `GENERATED ALWAYS AS (... sst_unaccent(S) ...)` clause;
  because `sst_unaccent` wraps the missing `unaccent`, the whole
  `CreateTable` call errors out with "function unaccent(text) does not
  exist".
- **Recovery**: install `postgresql-contrib` (or your distro's equivalent),
  run `CREATE EXTENSION unaccent` as a superuser in the `sstoryline`
  database, retry `Open()`.

See also the Setup page's
[manual-install section](Setup.md#manual-package-install) for distro notes.

## RAM-disk trade-offs

`contrib/ramify.sh` + `contrib/makeramdb.sh` set up a tmpfs-backed
PostgreSQL cluster at `/mnt/pg_ram`. The gain is real — bulk ingest avoids
disk I/O entirely — but comes with bright warnings:

- **Data is lost on reboot**. The tmpfs unmounts cleanly, taking the
  cluster with it.
- **Data is lost on tmpfs unmount**. Any admin who `umount /mnt/pg_ram`
  without coordinating destroys the DB.
- **Limited to 800 MB** by default (`contrib/ramify.sh:13`: `size=800M`).
  Adjust the `size=` in the `mount -t tmpfs` command for larger corpora.
  Make sure the host has the RAM to back it.
- **No backups**. `pg_dump` works, but anyone who reaches for RAM mode
  should be treating the `.n4l` source files as canonical and re-uploading
  after each reboot.

Use RAM mode for: ingest-speed benchmarks, query profiling without disk
noise, ad-hoc single-session experiments. Never use it for long-lived
knowledge bases.

## `LastSeen` 60-second sampling threshold

The two activity-logging functions apply a hardcoded dead zone before
recording a new observation:

```sql title="postgres_types_functions.go:1727"
-- 1 minute dead time
IF deltat > 60 THEN
   UPDATE LastSeen SET last=NOW(), delta=avdeltat, freq=f WHERE nptr = this;
ELSE
   return false;
END IF;
```

The same guard appears in `LastSawSection` at
[`:1691`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L1691).

**Why 60 seconds?** Rapid repeat observations (e.g. a user double-clicking
a search result, an MCP proxy issuing the same query twice in 300 ms)
would otherwise flood `LastSeen` with near-zero deltas and corrupt the
exponentially-smoothed `Delta` column. The 60-second gate turns that into
one counted sighting and lets the smoothed average represent genuine
user-visible sessions.

**Tuning**: the value is hardcoded in both functions. If you change it,
change both copies and re-run `Configure()` (which re-issues
`CREATE OR REPLACE FUNCTION`). No ALTER needed.

!!! info "Honesty about limitations"
    SSTorytime's observability surface is deliberately minimal — one
    activity table, one hardcoded threshold, no dynamic session/visit
    model. If your application needs click-level analytics, layer your
    own analytics on top of the underlying queries rather than trying to
    repurpose `LastSeen`.

## Other constants worth knowing

| Name | Value | Defined at | Role |
|---|---|---|---|
| `FORGOTTEN` | 10800 | [`globals.go:83`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L83) | 3-hour cutoff used by session-state heuristics. |
| `TEXT_SIZE_LIMIT` | 30 | [`globals.go:84`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L84) | Threshold used by text fractionation / short-form classification. |
| `LT128`, `LT1024` | 128, 1024 (bytes) | [`globals.go:51-52`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L51-L52) | Size-class breakpoints for the 6-bucket `NodePtr` addressing. |

None of these are runtime-tuneable. All bulk-loader tuning happens at the
upload end — e.g. running under `-wipe -u` to get a clean rebuild rather
than incremental reconciliation.

## See also

- [Setup](Setup.md) — picking Docker, manual, or RAM-disk.
- [Schema](Schema.md#unlogged-logged-bulk-load-lifecycle) — the state
  diagram behind the UNLOGGED→LOGGED flow.
- [Indexes](Indexes.md) — why all 5 GIN indexes are built post-upload.
- [Stored functions](Functions.md) — the PL/pgSQL code that honours every
  threshold on this page.

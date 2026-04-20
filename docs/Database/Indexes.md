# Indexes

SSTorytime creates **5 GIN indexes** on the `Node` and `ContextDirectory`
tables. They are all declared at the end of the bulk-upload path in
`GraphToDB()` at
[`db_upload.go:113-118`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go#L113-L118),
**after** every row has been inserted. Timing is not an accident — see the
[callout below](#why-indexes-are-built-last).

## The 5 indexes

```sql title="db_upload.go:113-118"
-- CREATE INDEX IF NOT EXISTS sst_type on Node (((NPtr).Chan),L,S)
CREATE INDEX IF NOT EXISTS sst_gin   on Node USING GIN (to_tsvector('english', Search));
CREATE INDEX IF NOT EXISTS sst_ungin on Node USING GIN (to_tsvector('english', UnSearch));
CREATE INDEX IF NOT EXISTS sst_s     on Node USING GIN (S);
CREATE INDEX IF NOT EXISTS sst_n     on Node USING GIN (NPtr);
CREATE INDEX IF NOT EXISTS sst_cnt   on ContextDirectory USING GIN (Context);
```

### `sst_gin` — accent-aware full-text search

- **Target:** `to_tsvector('english', Search)` on `Node`.
- **Why GIN:** tsvectors are sparse token-sets; GIN is the canonical
  PostgreSQL index for them and supports the `@@ to_tsquery(...)` operator.
- **Queries it serves:** every accent-*preserving* full-text search.
  The `Node.Search` column is itself a `tsvector GENERATED ALWAYS AS
  (to_tsvector('english', S))` ([Schema → Node](Schema.md#node)), so the
  index closes the loop between user query and stored representation.

### `sst_ungin` — accent-insensitive full-text search

- **Target:** `to_tsvector('english', UnSearch)` on `Node`.
- **Why GIN:** same reasoning — tsvector indexing.
- **Queries it serves:** searches that should succeed regardless of accents
  (so "cafe" matches "café"). `Node.UnSearch` is generated from
  `to_tsvector('english', sst_unaccent(S))`, and the immutable wrapper is
  what makes the generated column legal — see
  [`postgres_types_functions.go:1747`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L1747)
  and [Performance → the `unaccent` dependency](Performance.md#the-unaccent-dependency).

The two-index pair is the database side of SSTorytime's [dual tsvector
strategy](Performance.md#dual-tsvector-strategy): the caller picks whether
accent fidelity matters per-query.

### `sst_s` — raw text substring matching

- **Target:** the raw `S` text column on `Node`.
- **Why GIN:** GIN over `text` (with the right opclass) supports fast
  containment queries (`LIKE '%foo%'`, trigram lookups when `pg_trgm` is
  loaded). The default `text_ops` opclass also speeds up `ANY(ARRAY[...])`
  scans.
- **Queries it serves:** direct string lookups in `postgres_retrieval.go`
  that don't go through tsvector — e.g. chapter-filtered `WHERE S LIKE
  '%needle%'` fallbacks when full-text tokenisation would normalize away
  punctuation the user actually wants.

### `sst_n` — NodePtr direct lookup

- **Target:** composite `NPtr` on `Node`.
- **Why GIN:** `NPtr` is a `(Chan, CPtr)` composite. GIN with the right
  opclass lets the planner find all rows matching a `NPtr` value without
  scanning the whole table.
- **Queries it serves:** every `WHERE Nptr=start` clause inside
  `GetNeighboursByType`, `GetNCNeighboursByType`, `GetAppointments`, etc.
  Without this index, every single-hop fetch would degrade to a sequential
  scan — catastrophic for cone search.

### `sst_cnt` — context string membership

- **Target:** `Context` column on `ContextDirectory`.
- **Why GIN:** contexts are short strings that appear in `match_context` as
  arguments to case-insensitive overlap checks. GIN supports the prefix and
  substring patterns the matching code uses.
- **Queries it serves:** `SELECT Context INTO ctxstr FROM ContextDirectory
  WHERE ctxPtr=thisctxptr` inside `match_context`
  ([`postgres_types_functions.go:856`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L856)),
  plus the `sstorytime.py` and Go helpers that look up contexts by text.

## The commented-out `sst_type` btree

One line above the five GIN statements sits a disabled btree:

```sql title="db_upload.go:113"
-- sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_type on Node (((NPtr).Chan),L,S)")
```

It would have indexed `(NPtr.Chan, L, S)` as a btree — useful for
ordered scans by size-class then length. It is commented out because the
dominant query pattern is not ordered scans but random-access lookup by
`NPtr` or full-text match; maintaining another index on every insert cost
more than it saved. Left in source as a hint for anyone benchmarking a
workload that actually wants it.

## Why indexes are built last

!!! warning "Indexes are created after the bulk insert, not before"
    Building GIN indexes **incrementally** during a high-volume insert is
    pathologically slow — every row insert triggers GIN's internal pending
    list flush machinery, and on large imports this alone can dominate the
    upload time. SSTorytime sidesteps this by:

    1. Creating the tables `UNLOGGED` with **no indexes**
       ([`postgres_types_functions.go:32`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L32)).
    2. Bulk-inserting all nodes, arrows, contexts, pagemap events.
    3. **Then** creating the 5 GIN indexes in one pass
       ([`db_upload.go:114-118`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go#L114-L118)).
    4. Flipping the tables to `LOGGED` for durability
       ([`db_upload.go:119-120`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go#L119-L120)).

    Step 3's GIN builds scan the whole column once and batch-build the
    index — dramatically faster than incrementally maintaining five indexes
    during the inserts. The trade-off is that queries issued mid-upload
    would miss the indexes; SSTorytime's upload path is effectively
    offline, so this is fine.

## Dropped on wipe

The indexes (along with the btree placeholder) are explicitly dropped at
the top of `Configure()` when `WIPE_DB` is set:

```go title="session.go:195-201"
sst.DB.QueryRow("DROP INDEX sst_nan")
sst.DB.QueryRow("DROP INDEX sst_type")
sst.DB.QueryRow("DROP INDEX sst_gin")
sst.DB.QueryRow("DROP INDEX sst_ungin")
sst.DB.QueryRow("DROP INDEX sst_s")
sst.DB.QueryRow("DROP INDEX sst_n")
sst.DB.QueryRow("DROP INDEX sst_cnt")
```

(Note `sst_nan` is a historical name — the current `GraphToDB` does not
create it, but the `DROP` remains so an older database can be wiped without
error.) See
[`session.go:189-244`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L189-L244).

## See also

- [Schema → Node](Schema.md#node) — the columns the indexes cover.
- [Performance](Performance.md) — the dual tsvector strategy, accent
  handling, and bulk-load lifecycle that justify this index set.

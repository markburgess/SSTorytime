# Backup, restore, chapter deletion

The SSTorytime graph is a projection of your N4L source files. That framing — **N4L is the source of truth, PostgreSQL is a fast cache** — makes backup and disaster recovery almost trivial, and shapes every tradeoff discussed here.

!!! info "Prerequisites"
    - `src/bin/N4L`, `src/bin/removeN4L` compiled (`make` from repo root).
    - A PostgreSQL client (`psql`, `pg_dump`) in `$PATH`.
    - A location to store backups (filesystem directory or version-control remote).

## Backup strategy: version-control your N4L

The recommended backup strategy is:

```
your-notes/
├── .git/
├── brain.n4l
├── chinese.n4l
├── reminders.n4l
└── README.md
```

Keep your `.n4l` files in a git repository. Every edit is a commit. Every commit is a backup. Every diff is a meaningful history of how your thinking evolved.

```bash
cd your-notes
git init
git add *.n4l README.md
git commit -m "initial knowledge snapshot"
git remote add origin git@github.com:you/your-notes.git
git push -u origin main
```

This handles the **90% case**: if your laptop dies, you clone the repo and re-upload. No database snapshot required.

## Restore from N4L sources

Re-upload everything atomically:

```bash
cd your-notes
../src/bin/N4L -wipe -u *.n4l
```

- `-wipe` drops and recreates all SSTorytime tables before loading.
- `-u` triggers the upload.
- Passing every `.n4l` file at once means cross-file references resolve consistently in a single pass.

On a laptop with a few thousand nodes this takes seconds. On a large corpus (book-sized) it can take minutes to an hour; the upload is dominated by PL/pgSQL stored-function calls per link.

!!! tip "RAM-disk option for repeated wipes"
    If you iterate frequently, set PostgreSQL to use a tmpfs for its data directory
    (`contrib/ramify.sh`). The wipe-and-reload cycle becomes dramatically faster at the
    cost of losing everything on reboot — which is fine, because your source of truth
    is the git repo.

## Chapter-level removal

To delete a single [chapter](../concepts/glossary.md#chapter) without nuking the whole graph:

```bash
../src/bin/removeN4L -force "chapter name"
```

!!! warning "`-force` is required"
    Without `-force`, `removeN4L` exits immediately without doing anything. This is
    intentional — deletion is irreversible.

Behaviour:

- Calls the `DeleteChapter` PL/pgSQL stored procedure.
- Removes all nodes, links, and `PageMap` rows whose `Chap` matches the string.
- Does **not** touch the arrow directory or context directory, so arrows you defined inline in that chapter linger in `ArrowDirectory`. This is usually what you want (other chapters may reuse them), but it means "clean re-upload" is better achieved with `N4L -wipe -u`.

To list chapters before you delete:

```bash
../src/bin/searchN4L "\\chapter any"
```

or in psql:

```sql
SELECT DISTINCT Chap FROM Node ORDER BY Chap;
```

## Full-database snapshot with `pg_dump`

!!! danger "Never run `pg_dump` during `N4L -u`"
    The bulk-load tables (`Node`, `PageMap`, `ArrowDirectory`,
    `ArrowInverses`) are `UNLOGGED` for the duration of an ingest.
    A `pg_dump` taken mid-ingest captures a partial, inconsistent graph
    — and because UNLOGGED tables do not produce WAL, there is no
    transaction log you can recover the missing rows from. Dump only
    when the database is in its `LOGGED` steady state: after `N4L -u`
    has exited cleanly, or when you are certain no upload is in flight.
    If you automate snapshots, gate them on an "ingest lock" your
    operator owns, or on a timer that cannot overlap with your upload
    window.

When you want a point-in-time image of the whole graph — not just the sources — use PostgreSQL's native tooling:

```bash
# Compressed custom format, restorable with pg_restore
pg_dump \
  --host=localhost \
  --username=sstoryline \
  --format=custom \
  --file=sstorytime-$(date +%Y%m%d).dump \
  sstoryline

# Plain SQL, restorable with psql
pg_dump \
  --host=localhost \
  --username=sstoryline \
  --format=plain \
  --file=sstorytime-$(date +%Y%m%d).sql \
  sstoryline
```

Restore:

```bash
# Custom format
pg_restore --clean --if-exists --dbname=sstoryline sstorytime-20260420.dump

# Plain SQL
psql --username=sstoryline --dbname=sstoryline --file=sstorytime-20260420.sql
```

**Pros:**

- Captures schema, data, indexes, and stored functions in one file.
- Restore does not require running `N4L`, which is useful on machines that don't have the Go toolchain.
- Binary-exact restore: node `NPtr`s are preserved.

**Cons:**

- Snapshot-size grows with graph size; can be many multiples of the N4L source.
- Restores the exact state including any drift from your sources; if you have un-committed experiments in the DB you restore those too.
- PostgreSQL version coupling — a dump from 17 may not load cleanly into 15.

## Trade-offs: N4L sources vs DB snapshot

| Question | N4L sources in git | `pg_dump` snapshot |
|---|---|---|
| Human-readable? | Yes (plain text) | No (or verbose SQL) |
| Diff-friendly? | Yes, line-level | No |
| Captures manual DB edits? | No | Yes |
| Captures index statistics? | No (rebuilt on upload) | Yes |
| Fast to restore on a fresh host? | Seconds to minutes | Seconds |
| Cross-version safe? | Yes (text is text) | PostgreSQL-version-dependent |
| Fits on a flash drive? | Yes (KB-MB) | Often MB-GB |
| Preserves node `NPtr` addresses? | No (reassigned on upload) | Yes |

**Recommendation:** keep the N4L sources in version control as the primary backup. Take a `pg_dump` before large risky experiments (as a throwaway fallback), and delete those dumps when the experiment succeeds. Resist the urge to treat the dump as authoritative — as soon as you do, the sources drift and the abstraction leaks.

!!! tip "`pg_basebackup` as a physical-backup alternative"
    Operators running Postgres as a shared cluster (several application
    databases on the same server) may prefer a physical base backup over
    a per-database `pg_dump`. One `pg_basebackup -D /path/to/backup -X
    stream -P` captures the entire cluster — including `sstoryline`,
    other databases, roles, and WAL — in one image. This is the right
    tool for site-wide disaster recovery; it is overkill for a laptop
    that only hosts SSTorytime.

### `pg_restore -t` filters by **table**, not chapter

A common misconception: `pg_restore --table=Node` or similar flags do
**not** let you restore a single chapter. `-t/--table` selects a
PostgreSQL table object (e.g. `Node`, `PageMap`) from the dump; there is
no `pg_restore` equivalent to `removeN4L`'s per-chapter scoping. If you
need a subset restore, the pattern is:

1. Restore the full dump into a scratch database
   (`pg_restore --clean --if-exists -d sstoryline_scratch
    sstorytime-YYYYMMDD.dump`).
2. Use `removeN4L -force` or SQL (`DELETE FROM Node WHERE Chap = 'X'`)
   in the scratch database to keep only what you want.
3. `pg_dump` the scratch, then restore into production — or use `COPY`
   to shuttle the desired rows.

There is no "restore chapter X as-of last Tuesday" workflow built in.
That property, if you need it, lives in your N4L git history.

### Corruption vs data loss — the distinction that matters

The N4L-as-source-of-truth pattern protects against **data loss**
(host fails, disk dies, you ran `N4L -wipe -u` unintentionally): you
clone the repo, re-upload, done. It does **not** protect against
**corruption at the source** — if one of your `.n4l` files got
poisoned with wrong facts and committed six weeks ago, re-uploading
re-introduces the poison. The only defence is the history in the
`.n4l` repository itself (`git log -p`, `git bisect`, peer review of
edits). Treat N4L sources the way you treat application code: review
changes, keep history, know how to blame.

Relatedly: **SSTorytime does not support point-in-time recovery (PITR)**
in the sense of continuous WAL-archiving-plus-restore-to-a-moment.
The database is a derived artefact. The "recovery point" concept
belongs in your N4L source repo's commit history, not in Postgres WAL.
Do not set up `archive_command` expecting cross-`wipe -u` replay to
work — `N4L -wipe` drops and recreates the tables, which is not a
traditional transactional rollback.

## Restore checklist after a disaster

1. Ensure PostgreSQL 17 is installed and running.
2. Run `make db` from the repo root to create the role and database.
3. `git clone` your N4L source repo.
4. From the source directory: `../src/bin/N4L -wipe -u *.n4l`.
5. Smoke-test: `../src/bin/searchN4L "%%" \\limit 5` — expect five random hits.
6. If you kept a `pg_dump` for pre-experiment state, restore it into a
   **separate** scratch database (`pg_restore -d sstoryline_scratch
   …`), not on top of the freshly-loaded production DB. From there,
   `COPY` or `SELECT INTO` the rows you need back into production —
   remembering that `NPtr` addresses in the dump are scoped to the old
   graph and may collide with the new one.

That's it. A complete knowledge graph from source, end to end, in under five minutes for modest corpora.

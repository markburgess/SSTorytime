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

To delete a single chapter without nuking the whole graph:

```bash
../src/bin/removeN4L -force "chapter name"
```

!!! warning "`-force` is required"
    Without `-force`, `removeN4L` exits immediately without doing anything. This is
    intentional — deletion is irreversible. See [removeN4L.md](../removeN4L.md) for the
    guard details.

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

## Restore checklist after a disaster

1. Ensure PostgreSQL 17 is installed and running.
2. Run `make db` from the repo root to create the role and database.
3. `git clone` your N4L source repo.
4. From the source directory: `../src/bin/N4L -wipe -u *.n4l`.
5. Smoke-test: `../src/bin/searchN4L "%%" \\limit 5` — expect five random hits.
6. If you also rely on a dump, use `pg_restore` to the chapter or subset you care about, and re-run step 5 to verify.

That's it. A complete knowledge graph from source, end to end, in under five minutes for modest corpora.

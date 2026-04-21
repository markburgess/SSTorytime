# When things go wrong

> **Most of what breaks breaks in one of a handful of ways. Here are the ones
> that show up after the quick checks in [Install in 5 minutes](../GettingStarted.md) have already passed.**

Symptom, one-line cause, one-line fix. If none of these match, the essays
and the glossary are there; the repo's `developers/` folder has the deeper
reference.

---

## `psql` can't connect after a reboot

**Cause.** PostgreSQL isn't running, or is running on a port the tools
aren't pointing at.

**Fix.** Start it (`systemctl start postgresql`, or `pg_ctl -D … start`
if you are using a RAM-disk data directory). If the daemon is up but
`psql` still refuses, check for a stale socket file in `/var/run/postgresql/`
or `/tmp/` — a hard shutdown can leave one behind, and the client will
fail to connect until it is removed.

---

## `N4L` complains that an arrow has not been declared

**Cause.** You used an arrow in a `(…)` slot that isn't in `SSTconfig/`. N4L
does not invent arrows; every shorthand has to be declared somewhere.

**Fix.** Pick one of the existing arrows ([Thinking in arrows](../arrows.md)
has the list), or add the new declaration to a file in `SSTconfig/` and
re-ingest. Typos count — `(citess)` and `(cites)` are different arrows as
far as N4L is concerned.

---

## `searchN4L` returns nothing

**Cause.** Either the ingest didn't land, or the string you searched for
doesn't exactly match a node.

**Fix.** Re-run `N4L -u examples/reading-list.n4l` (or whatever your file
is) and watch for errors. If ingest looks clean, try a substring or a
less specific phrase — the match is on node text, not on keywords inside
it.

---

## A chapter keeps growing duplicates every time you ingest

**Cause.** You re-ran `N4L -u` without wiping the old version first.
Re-ingesting does not replace; it adds. The old nodes are still there,
and now you have two of everything.

**Fix.** Either wipe the whole graph with `N4L -wipe -u` before ingesting,
or drop just the one chapter with `removeN4L -force "chapter name"` and
then ingest. Remember `-force` is required — without it, `removeN4L`
exits silently.

---

## `pg_dump` gave you a partial graph

**Cause.** You took the dump while `N4L -u` was still running. The
bulk-load tables are held in an unlogged state during ingest, and a
snapshot taken then captures an inconsistent subset.

**Fix.** Wait for ingest to exit cleanly before dumping. See
[Backing up and restoring](../cookbooks/backup-restore-delete.md) for
the timing rule.

---

## Tools refuse to start, complaining about `~/.SSTorytime`

**Cause.** The credentials file is not mode 600. It holds a plaintext
password and the tools will not read a world-readable copy.

**Fix.**

```
chmod 600 ~/.SSTorytime
```

---

## Still stuck

The [glossary](../concepts/glossary.md) explains the terms that show up
in error messages. The essays under [Concepts](../concepts/why-semantic-spacetime.md)
cover the mental model, which is often the real fix. Deeper reference —
the Go API, the HTTP API, the PostgreSQL schema — lives under the
`developers/` folder in the repo.

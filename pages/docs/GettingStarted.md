# Install in 5 minutes

![A retro CRT terminal on a wooden desk flanked by a PostgreSQL elephant and a Go gopher, connected by ochre flow arrows.](figs/getting_started_hero.jpg){ align=center }

> **By the end of this page you'll have a working graph on your machine, a small
> reading list loaded into it, and one answered question to prove it.**

One happy path, straight through. About five minutes on Linux with PostgreSQL
already installed. If you are on macOS or would rather run the database in
Docker, those notes live in the repo's `developers/` folder — they were cut
from this page so the main path stays short.

---

## Before you start

You need four things on your machine. One line each to check.

| You need | Check it with | What you should see |
|---|---|---|
| `git` | `git --version` | `git version 2.x` or newer |
| Go 1.21+ | `go version` | `go version go1.21` or newer |
| PostgreSQL 14+ | `pg_config --version` | `PostgreSQL 14.x` or newer |
| `make` | `make --version` | `GNU Make 4.x` |

If any command is missing, install it with your distro's package manager and
come back.

---

## 1. Get the code

```
git clone https://github.com/markburgess/SSTorytime.git
cd SSTorytime
```

---

## 2. Set up the database

The database is a cache — your N4L files on disk are the source of truth — so
running it in RAM is fine, fast, and easier on your SSD. The snippet below
finds your Postgres binary directory via `pg_config`, which works across
distros and major versions.

!!! tip "Don't want a RAM-disk?"
    Skip the `mkdir` / `mount` / `chown` lines and the `initdb` into
    `/mnt/pg_ram/pgdata`. Use your distro's default data directory instead
    (`systemctl start postgresql` and continue from the `psql` step). You'll
    lose nothing except the speed and the SSD-wear saving.

As root:

```
sudo su -

mkdir -p /mnt/pg_ram
mount -t tmpfs -o size=800M tmpfs /mnt/pg_ram
chown postgres:postgres /mnt/pg_ram
systemctl stop postgresql

su - postgres

PG_BINDIR=$(pg_config --bindir 2>/dev/null || dirname $(which initdb))
$PG_BINDIR/initdb -D /mnt/pg_ram/pgdata
$PG_BINDIR/pg_ctl -D /mnt/pg_ram/pgdata -l /mnt/pg_ram/logfile start

psql <<'SQL'
CREATE USER sstoryline PASSWORD 'sst_1234' SUPERUSER;
CREATE DATABASE sstoryline;
GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline;
CREATE EXTENSION UNACCENT;
SQL
```

!!! warning "Credentials file must be chmod 600"
    If you change the default database name, user, or password, write the new
    values to `~/.SSTorytime` and **`chmod 600` that file**. It holds a
    plaintext password; the tools will refuse to read it if it's
    world-readable. This requirement is audit-flagged and is not optional,
    even on a single-user machine.

    ```
    chmod 600 ~/.SSTorytime
    ```

Back in your own shell, `psql sstoryline` should drop you into a shell without
prompting for a password. Ctrl-D to leave it.

---

## 3. Build

From the repo root:

```
make -C src
```

Binaries land in `src/bin/`. Put that on your `$PATH` so `N4L`, `searchN4L`,
and the rest work as bare commands:

```
export PATH=$PWD/src/bin:$PATH
```

---

## 4. Load your first story

The repo ships with a small reading list — seven books, what each is about,
which cite each other, when they were read. Load it and ask one question:

```
N4L -u examples/reading-list.n4l
searchN4L "decision making"
```

Expected output shape — three books, each annotated with the topics and
takeaway you'd expect from the corpus:

```
"Thinking Fast and Slow"   (is about) decision making
                           (is about) dual-process cognition
                           (one-line takeaway) two systems, one of them lazy, both of them you

"Superforecasting"         (is about) decision making
                           (cites)    "Thinking Fast and Slow"
                           (one-line takeaway) calibration beats cleverness over long horizons

"Thinking in Systems"      (is about) decision making
                           (is about) feedback loops
                           (one-line takeaway) most problems are the shape of a feedback loop
```

Three books you wrote about a shared topic, pulled back together by asking
the graph. That's the whole loop.

---

## When things don't work

- **`psql` can't connect.** PostgreSQL isn't running where your tools look for
  it. Check `pg_ctl -D /mnt/pg_ram/pgdata status` (or `systemctl status
  postgresql` for the disk path) and start it if it's down.
- **`N4L` says "chapter empty" or refuses to ingest.** Run the command from
  the repo root so `examples/reading-list.n4l` resolves; N4L doesn't search
  upward for files.
- **`searchN4L "decision making"` returns nothing.** The ingest didn't land.
  Re-run `N4L -u examples/reading-list.n4l` and watch for errors; a common
  cause is stale credentials in `~/.SSTorytime` from a previous install.

---

## Next

One reading list loaded, one question answered. Now write your own:
[Your first story](Tutorial.md).

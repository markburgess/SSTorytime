# How you can contribute to SSTorytime

This project is, in the end, a teaching project — an argument that knowledge
is what happens in the head of someone who has sat with material for long
enough to know it like a friend. Teachers and students are both welcome here,
and the line between them is meant to be porous: the best teachers are still
learning, and the best students are already teaching the person sitting next
to them. If you found the tool useful, you are in a position to make it
useful for the next person. This page walks you through the practical side —
setting up a development environment, the repository layout, how code and
documentation changes flow through the project, and where to ask when you
get stuck.

!!! tip "The single biggest thing you can do"
    Use SSTorytime on something you actually want to know. A book you're
    reading. A problem you're stuck on. A subject you'd like to teach, or a
    family history you'd like to hold on to. The tool improves when real
    humans push on it with real questions — and, not incidentally, so do
    the humans.

## Development setup

### Prerequisites

| Tool | Minimum version | Why |
|---|---|---|
| Go | 1.24.2 | [`go.mod`](https://github.com/markburgess/SSTorytime/blob/main/go.mod) pins `go 1.24.2`. |
| PostgreSQL | 17 | Runtime dependency for the library and tools. |
| `make` | any | Orchestrates the two-tier build. |
| `git` | any | Source control + fork workflow. |
| `openssl` | any | Generates the self-signed TLS cert for the web server. |

Optional but recommended: Docker (if you'd rather run PostgreSQL in a
container — see [`postgres-docker/`](https://github.com/markburgess/SSTorytime/blob/main/postgres-docker/README.md)),
and Python 3.11+ if you want to preview the docs site locally.

### Clone and build

```bash
git clone https://github.com/markburgess/SSTorytime.git
cd SSTorytime
make all
```

After `make all`, the compiled binaries land in `src/bin/` (the library CLIs
plus `http_server`) and `src/demo_pocs/bin/` (integration-test programs). See
the [Build system](build-system.md) page for the full pipeline.

### Initialise the database

Two options, depending on where you want Postgres to keep its data files:

```bash
make ramdb    # RAM-backed, recommended for development (fast, ephemeral)
make db       # Disk-backed (persistent across reboots)
```

!!! info "`make db` fix landed in Phase 0"
    The root `Makefile`'s `make db` target historically had a broken
    invocation (`(cd contrib/makedb.sh)` — `cd` into a file fails). The fix
    is part of the documentation upleveling; see
    [`contrib/makedb.sh`](https://github.com/markburgess/SSTorytime/blob/main/contrib/makedb.sh)
    for what the script does.

Then load the examples so you have something to query:

```bash
cd examples/
make
```

## Repository layout

| Path | What lives here |
|---|---|
| [`pkg/SSTorytime/`](https://github.com/markburgess/SSTorytime/tree/main/pkg/SSTorytime) | The Go library — types, graph operations, PostgreSQL wiring, N4L parser. This is the public API. |
| [`src/`](https://github.com/markburgess/SSTorytime/tree/main/src) | CLI tools (`N4L`, `searchN4L`, `pathsolve`, `notes`, `graph_report`, `text2N4L`, `removeN4L`), the HTTPS web server, and the `API_EXAMPLE_*` reference programs. |
| [`src/demo_pocs/`](https://github.com/markburgess/SSTorytime/tree/main/src/demo_pocs) | Proof-of-concept integration tests (`dotest_getnodes`, `dotest_entirecone`, `definecontext`, `postgres_testdb`). |
| [`examples/`](https://github.com/markburgess/SSTorytime/tree/main/examples) | N4L source files showing how to model knowledge — history, recipes, physics, language, etc. |
| [`docs/`](https://github.com/markburgess/SSTorytime/tree/main/docs) | The source for this MkDocs Material site. |
| [`tests/`](https://github.com/markburgess/SSTorytime/tree/main/tests) | Three-tier test harness (parser / library / integration). See the [Testing](testing.md) page. |
| [`contrib/`](https://github.com/markburgess/SSTorytime/tree/main/contrib) | Helper scripts — `makedb.sh`, `ramify.sh`, `makeramdb.sh`. |
| [`postgres-docker/`](https://github.com/markburgess/SSTorytime/tree/main/postgres-docker) | Docker Compose setup for running PostgreSQL in a container. |
| [`SSTconfig`](https://github.com/markburgess/SSTorytime/blob/main/SSTconfig) | Default Postgres connection parameters (host, port, user, password, database). |

## Development workflow

SSTorytime uses the standard GitHub fork-and-PR flow:

1. **Fork** [`markburgess/SSTorytime`](https://github.com/markburgess/SSTorytime)
   on GitHub into your own account.
2. **Clone** your fork locally and add the upstream as a remote:
   ```bash
   git clone https://github.com/<you>/SSTorytime.git
   cd SSTorytime
   git remote add upstream https://github.com/markburgess/SSTorytime.git
   ```
3. **Branch** from `main` using a descriptive name:
   ```bash
   git checkout -b feature/my-feature
   ```
   Conventional prefixes: `feature/…`, `fix/…`, `docs/…`, `test/…`, `chore/…`.
4. **Commit** in small, logical chunks (see message conventions below).
5. **Push** your branch to your fork:
   ```bash
   git push -u origin feature/my-feature
   ```
6. **Open a pull request** against `markburgess/SSTorytime:main`. Describe
   *why* the change is needed, not just what it does. Link to any related
   issue.

Keep your branch current by rebasing on upstream `main` before requesting
review:

```bash
git fetch upstream
git rebase upstream/main
```

## Code style

- Follow [Effective Go](https://go.dev/doc/effective_go). The Go standard
  layout is our baseline — packages in `pkg/`, executables in `src/<tool>/`.
- Every exported identifier (function, type, constant) should have a GoDoc
  comment that begins with the identifier's name: `// MyFunction does X…`.
  This is how `go doc` and IDE tooltips discover intent.
- Prefer descriptive names. `NextArrowIndex` beats `nai`; `adjacency` beats
  `m`. Names are read far more often than they are typed.
- Keep external dependencies minimal. The project deliberately uses only
  [`lib/pq`](https://github.com/lib/pq) for PostgreSQL —
  [`go.mod`](https://github.com/markburgess/SSTorytime/blob/main/go.mod) is
  two lines for a reason. Before adding a dependency, ask if the stdlib can
  do it.
- `gofmt` (or your editor's Go-on-save hook) before committing.

## Commit messages

Use the conventional-commits-ish style the repo already follows:

```
<type>(<scope>): short imperative subject line

Optional longer body explaining the why. Wrap at ~72 columns.
```

Types you'll see in history:

| Type | When to use |
|---|---|
| `feat` | A new user-visible feature. |
| `fix` | A bug fix that changes observable behaviour. |
| `docs` | Documentation-only changes. |
| `test` | Adding or adjusting tests. |
| `chore` | Build-system, dependency, or housekeeping work. |
| `refactor` | Internal rework that doesn't change behaviour. |

Examples:

```
docs(tutorial): fix binary path after make reorganized src/bin output
feat(searchN4L): add -context flag to scope queries to a chapter
fix(Makefile): repair make db target (cd into file → sh invocation)
```

Keep commits scoped — one logical change per commit. A twenty-file
"WIP: stuff" commit is hard to review and impossible to revert cleanly.

## Running tests

The test harness is a shell script with three tiers (parser / library /
integration). From the repo root:

```bash
cd tests
./run_tests
```

You need a running PostgreSQL (on `localhost:5432` with the default
credentials from [`SSTconfig`](https://github.com/markburgess/SSTorytime/blob/main/SSTconfig))
and the example data loaded first (`cd examples && make`). See the
dedicated [Testing](testing.md) page for how to add a new test case and
what the output means.

## Documentation style

This doc site is built with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/).
A few conventions to keep it coherent:

- **Image paths are relative.** Write `![alt](figs/foo.png)`, never a
  GitHub blob URL. That way the site works offline and on GitHub's file
  renderer both.
- **Cite code with `[file:line](github-blob-url)`.** Every claim about
  behaviour should point to the line in the code that produces it.
  Example:
  [`src/server/http_server.go:132`](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L132).
- **Use the `code-ref` admonition** for "see in code" callouts:

  ```markdown
  !!! code-ref "See in code"
      [`pkg/SSTorytime/postgres_types_functions.go:1203`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L1203)
      defines `GetNodesStartingStoryMatchingName`.
  ```

- **Preview locally** before submitting:

  ```bash
  pip install -r docs/requirements.txt
  mkdocs serve
  ```

  Open <http://localhost:8000/SSTorytime/>. If `mkdocs build --strict`
  emits warnings, CI will fail on the PR — fix them before pushing.

## Where to ask

- **GitHub Issues** — bug reports and feature requests:
  <https://github.com/markburgess/SSTorytime/issues>
- **GitHub Discussions** (if enabled on the repo) — general design
  discussion and Q&A.
- **LinkedIn group** — longer-form conversations with Mark Burgess and
  other users: <https://www.linkedin.com/groups/15875004/>

When reporting a bug, include: Go version (`go version`), PostgreSQL
version (`psql --version`), the exact command you ran, the full error
output, and a minimal N4L input file that reproduces the issue if
relevant.

## What's a good first contribution?

Small, self-contained, visible. Suggestions:

- **Fix a cookbook typo.** The [Search Examples](search_examples.md) and
  [Example Walkthrough](example.md) pages are battle-tested but still
  collect the occasional typo.
- **Add an N4L example file.** If you've used N4L to model something real
  (a recipe, a reading list, a project plan), contribute it to the
  [`examples/`](https://github.com/markburgess/SSTorytime/tree/main/examples)
  directory with a short README.
- **Clarify a confusing doc section.** If you stumbled over an
  explanation, fix it while the confusion is still fresh — that's a
  high-value PR. (It's also how this file came to exist.)
- **Add a unit test.** Pick an exported function in
  [`pkg/SSTorytime/`](https://github.com/markburgess/SSTorytime/tree/main/pkg/SSTorytime)
  that lacks one. Small tests that document intent are worth a lot.
- **Document an undocumented function.** Add a GoDoc comment to an
  exported identifier that's missing one. The `code-index.md` page
  tracks what's still bare.

!!! tip "Don't try to do too much at once"
    A 30-line PR that lands in a day is worth ten 2000-line PRs that
    never merge. Start tiny, get the workflow right, then grow the scope.

## Beyond the code

- Are you a teacher? Get your students to try it — SSTorytime is ideal
  for class projects and can be used in any language.
- Anyone can contribute N4L notes without writing a line of Go. If you
  take notes, you can use SSTorytime.
- If you work in web UI / visualization / information design and have
  ideas for how to render a knowledge graph in new ways, please get in
  touch. The [`src/server/`](https://github.com/markburgess/SSTorytime/tree/main/src/server)
  frontend is deliberately minimalist — there's a lot of room.
- A Python interface is on the horizon (see
  [`src/python_integration_example.py`](https://github.com/markburgess/SSTorytime/blob/main/src/python_integration_example.py)
  and [`src/SSTorytime.py`](https://github.com/markburgess/SSTorytime/blob/main/src/SSTorytime.py)).
  Extending it is a great way in for Python folk.

Thank you for being here. The project is stronger for every person who
takes the time to read the code, try it on their own data, and write
down what they learn.

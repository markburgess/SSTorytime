# User-Facing Documentation Restructure — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task.

**Goal:** Execute the four-phase restructure specified in `2026-04-21-user-facing-restructure-design.md` — nav rewrite, file moves, content rewrites in user register — landing a published `/pages/docs/` site that reads for writers and queryers, not developers.

**Architecture:** Four sequential phases (A → B → C → D). Phase A is mechanical (nav + file moves + deletes + link audit, ~1h). Phase B is the anchor-three voice-setter (sequential single-author, ~4–8h, **hard review checkpoint for Mark before Phase C**). Phase C is parallel two-agent track rewrites with strict file ownership (~4h). Phase D is concepts + polish (~2–3h). `mkdocs build --strict` must pass at the exit of every phase.

**Tech stack:** MkDocs 1.6.1 + `mkdocs-material` 9.5.44 (already installed via `pages/requirements.txt`). Verification command throughout: `mkdocs build --strict -f pages/mkdocs.yml`. Git commits are always `Co-Authored-By: Claude` via HEREDOC.

**Branch posture:** Land Phase A as a small standalone commit on `docs/phase3-8-content` (or a stacked branch from it). Phases B/C/D on a fresh branch `docs/phase9-user-facing` stacked on top of the post-Phase-A tip, so upstream rebases affect at most one phase at a time.

**Reference docs:**
- Design doc: `pages/docs/plans/2026-04-21-user-facing-restructure-design.md` (226 lines; read this first if you didn't write it)
- Prior plan (developer-facing baseline): `pages/docs/plans/2026-04-20-documentation-upleveling.md`
- Voice exemplars (already Burgess-register post-audit): `pages/docs/concepts/why-semantic-spacetime.md`, `pages/docs/concepts/glossary.md` (Wave-front / Orbit / Story entries), `pages/docs/cookbooks/bring-your-own-corpus.md` (intro only)

---

## Phase A — Structure (mechanical, ~1h, 1 agent)

**Exit criterion:** `mkdocs build --strict -f pages/mkdocs.yml` is clean. Site renders the 7-tab nav. Developer files exist in repo at `pages/docs/developers/` but not on the site. Three stale files are gone. All internal cross-links resolve.

### Task A1: Pre-flight — verify baseline builds clean

**Files:** None (read-only check)

**Step 1:** Run the verification command before changing anything, so any later break is clearly caused by this phase, not pre-existing state.

```bash
cd /home/msvoboda/github/SSTorytime
mkdocs build --strict -f pages/mkdocs.yml 2>&1 | tail -20
```

**Expected:** Build completes, `INFO - Documentation built in Ns` line. Zero warnings. If you see warnings, **stop and triage before proceeding** — they will compound with this phase's changes.

### Task A2: Create the `developers/` scaffold

**Files:**
- Create: `pages/docs/developers/README.md`

**Step 1:** Create the directory with a one-page README that explains what lives here and why it's excluded from the site.

```markdown
# Developer Documentation

These pages are kept in the repo for contributors but are **excluded from the published site**
(see `exclude_docs: developers/*` in `pages/mkdocs.yml`). The published site at
https://markburgess.github.io/SSTorytime/ is user-facing — this folder holds the deeper
technical reference: Go API, HTTP API, PostgreSQL schema, internals, and contributor guides.

If you arrived here from a user-facing page looking for detail: you're in the right place.
If you want the user-facing site, start at [Home](../index.md).
```

**Step 2:** Commit the scaffold alone before moving files into it, so the `git mv` operations are a clean second commit.

```bash
git add pages/docs/developers/README.md
git commit -m "docs(structure): add developers/ scaffold (excluded from site)"
```

### Task A3: Move the 18 developer-facing files

**Files:** `git mv` only. No content changes.

**Step 1:** Execute moves in one batch so the resulting commit is atomic.

```bash
cd /home/msvoboda/github/SSTorytime
git mv pages/docs/API.md pages/docs/developers/API.md
git mv pages/docs/API_WALKTHROUGH.md pages/docs/developers/API_WALKTHROUGH.md
git mv pages/docs/WebAPI.md pages/docs/developers/WebAPI.md
git mv pages/docs/http-api pages/docs/developers/http-api
git mv pages/docs/http_server.md pages/docs/developers/http_server.md
git mv pages/docs/Database pages/docs/developers/Database
git mv pages/docs/dynamic_functions.md pages/docs/developers/dynamic_functions.md
git mv pages/docs/graph_report.md pages/docs/developers/graph_report.md
git mv pages/docs/notes.md pages/docs/developers/notes.md
git mv pages/docs/removeN4L.md pages/docs/developers/removeN4L.md
git mv pages/docs/code-index.md pages/docs/developers/code-index.md
git mv pages/docs/debug.md pages/docs/developers/debug.md
git mv pages/docs/namespaces.md pages/docs/developers/namespaces.md
git mv pages/docs/build-system.md pages/docs/developers/build-system.md
git mv pages/docs/testing.md pages/docs/developers/testing.md
git mv pages/docs/versioning.md pages/docs/developers/versioning.md
git mv pages/docs/tls-certificates.md pages/docs/developers/tls-certificates.md
git mv pages/docs/operations pages/docs/developers/operations
git mv pages/docs/howtocontribute.md pages/docs/developers/howtocontribute.md
git mv pages/docs/cookbooks/first-go-program.md pages/docs/developers/first-go-program.md
git mv pages/docs/cookbooks/python-integration.md pages/docs/developers/python-integration.md
git mv pages/docs/cookbooks/llm-via-mcp-sst.md pages/docs/developers/llm-via-mcp-sst.md
```

**Step 2:** Verify the moves.

```bash
ls pages/docs/developers/ | wc -l
# Expected: 18+ entries (files + the 3 subdirs http-api/, Database/, operations/, + README.md)
ls pages/docs/
# Expected: only user-facing files remain at the top level
```

**Step 3:** Do not commit yet — Task A4 deletes the three stale files in the same commit.

### Task A4: Delete the three stale files

**Files:**
- Delete: `pages/docs/ToDo.md`
- Delete: `pages/docs/concepts/architecture.md`
- Delete: `pages/docs/README.md`

**Step 1:** Remove and verify.

```bash
git rm pages/docs/ToDo.md pages/docs/concepts/architecture.md pages/docs/README.md
```

**Step 2:** Commit the moves + deletes together.

```bash
git commit -m "$(cat <<'EOF'
docs(structure): move developer pages to developers/; delete stale files

Per 2026-04-21 restructure design, 18 developer-facing files move to
pages/docs/developers/ (to be excluded from the published site in a
follow-up commit) while remaining in the repo for contributors.

Deleted:
- ToDo.md (project hygiene, not documentation)
- concepts/architecture.md (dev-framed stack diagram; user-facing
  Concepts overview to be authored in Phase D)
- README.md (stub)

Next: mkdocs.yml nav rewrite + exclude_docs + link audit.
EOF
)"
```

### Task A5: Rewrite `mkdocs.yml` nav and excludes

**Files:**
- Modify: `pages/mkdocs.yml` (nav section at line 103–167; `exclude_docs` at line 13)

**Step 1:** Replace the `exclude_docs` block to add `developers/*`.

Current (line 13–19):

```yaml
exclude_docs: |
  /README.md
  requirements.txt
  plans/*
  # NOTE (deferred): kebab-case file renames ...
```

Replace with:

```yaml
exclude_docs: |
  /README.md
  requirements.txt
  plans/*
  developers/*
  # NOTE: developers/* is kept in the repo for contributors but
  # excluded from the published site per the 2026-04-21 restructure.
```

**Step 2:** Replace the `nav:` block (line 103 onwards) with the 7-tab structure. Note: page file paths reference *user-facing* names. Several pages (like `N4L.md` "user half") will be split in Phase C; for now point the nav at the original paths and let Phase C supersede.

```yaml
nav:
  - Home: index.md

  - Start Here:
      - Install in 5 minutes: GettingStarted.md
      - Your first story: Tutorial.md
      - Try the examples: example.md
      - FAQ: FAQ.md

  - Writing Stories:
      - Thinking in arrows: arrows.md
      - Writing N4L by hand: N4L.md
      - Turning documents into stories: text2N4L.md
      - Patterns — research notes: cookbooks/bring-your-own-corpus.md

  - Asking Questions:
      - Finding things: searchN4L.md
      - Finding paths: pathsolve.md
      - Context — asking the same question different ways: howdoescontextwork.md
      - Patterns — search recipes: cookbooks/search-recipes.md

  - Living With Your Graph:
      - Backing up, restoring: cookbooks/backup-restore-delete.md

  - Concepts:
      - Semantic spacetime in plain English: concepts/why-semantic-spacetime.md
      - Glossary: concepts/glossary.md
      - Storytelling: Storytelling.md
      - Knowledge and learning: KnowledgeAndLearning.md
      - Outreach: outreach.md
```

**Step 3:** Also update `search_examples.md` — it either folds into `Finding things` (preferred) or moves to developers/. Simplest Phase A choice: leave it in place but omit from nav. Phase C's Queryer agent decides whether to consolidate or delete.

**Step 4:** Do **not** commit yet — Task A6 runs the link audit so link fixes land in the same commit.

### Task A6: Link audit

**Files:** All remaining files at `pages/docs/*.md` (not under `developers/`)

**Step 1:** Build with strict. Strict mode fails on any broken internal link, giving us a free exhaustive audit.

```bash
cd /home/msvoboda/github/SSTorytime
mkdocs build --strict -f pages/mkdocs.yml 2>&1 | tee /tmp/mkdocs-phaseA-audit.log | tail -60
```

**Expected initial state:** Build fails with one or more `WARNING -  Doc file 'foo.md' contains a link 'bar.md' (or 'directory/foo.md'), but the target is not found among documentation files.` Each such warning points at a cross-link into a moved or deleted file.

**Step 2:** For each warning, the fix is one of:
- **Link to a moved file from a surviving user-facing page** → rewrite the link to `developers/<file>.md` **AND** reconsider whether to keep the link at all (users shouldn't be pushed at dev pages from the user-facing site). When in doubt, remove the link and let the footer "For developers, see the repo" line do the work.
- **Link from a moved file (now under `developers/`) to another moved file** → rewrite to relative path within `developers/`. Example: `[Schema](Database/Schema.md)` → `[Schema](Database/Schema.md)` if both are now under `developers/` (relative path unchanged); `[API](API.md)` from a `developers/cookbooks/*` file → `[API](../API.md)`. `mkdocs build` complains about these only if a `developers/` file references a *user-facing* file, because `developers/*` is excluded. Cross-references between `developers/*` pages are fine — they render in the repo rendering (GitHub markdown view) even though not on the built site.
- **Link to a deleted file** → remove the link; if the linked content was load-bearing, leave a `<!-- TODO Phase D: replace link to deleted architecture.md with new Concepts overview -->` marker.

**Step 3:** Re-run `mkdocs build --strict` until zero warnings.

**Step 4:** Do a human sanity pass — open 3–5 of the reframe-candidate pages (`index.md`, `Tutorial.md`, `GettingStarted.md`, `arrows.md`, `searchN4L.md`) in the built site (`mkdocs serve -f pages/mkdocs.yml`) to confirm the nav is navigable and no nav entries 404.

### Task A7: Commit Phase A

**Step 1:** Single squash commit for the nav + excludes + link-audit fixes.

```bash
git add pages/mkdocs.yml pages/docs/
git commit -m "$(cat <<'EOF'
docs(nav): user-facing 7-tab structure; exclude developers/* from site

Per 2026-04-21 restructure design, the nav pivots from developer/
operator-facing (12 tabs: APIs, Operations, Reference, Tools,
Contributing, Project, ...) to user-facing (7 tabs: Start Here,
Writing Stories, Asking Questions, Living With Your Graph,
Concepts, plus Home).

- pages/mkdocs.yml: new 7-tab nav; developers/* added to exclude_docs
- internal cross-links updated throughout pages/docs/*.md to reflect
  the move of 18 files to developers/ (see prior commit)
- mkdocs build --strict passes clean

This commit carries *structure only*. Content rewrites (Phase B
anchor-three, Phase C tracks, Phase D concepts+polish) land in
subsequent commits on docs/phase9-user-facing.
EOF
)"
```

**Step 2:** Verify the branch tip.

```bash
git log --oneline -5
mkdocs build --strict -f pages/mkdocs.yml 2>&1 | tail -3
```

**Exit Phase A.** Decide with Mike whether to (a) push Phase A to `fork/docs/phase3-8-content` as a PR #57 follow-up, or (b) hold and stack Phase B on top locally first.

---

## Phase B — Anchor three (sequential, 1 agent, ~4–8h) — **REVIEW CHECKPOINT**

**Exit criterion:** `examples/reading-list.n4l` authored and tested; `index.md`, `GettingStarted.md` (→ "Install in 5 minutes"), `Tutorial.md` (→ "Your first story") rewritten in user register; `mkdocs build --strict` clean. **Present to Mark before Phase C dispatches.**

### Task B1: Author `examples/reading-list.n4l`

**Files:**
- Create: `examples/reading-list.n4l`
- Reference: `examples/Mary.n4l`, `examples/chinese.n4l` for N4L syntax patterns

**Target content:** ~20 lines describing 5–8 books/papers, what they're about, what they cite, and (for 2–3 of them) who read them and when. The corpus must support three target queries:

1. "What connects these two papers?" — needs at least one chain of 2+ arrows linking two items via a shared topic.
2. "What have I read that's about X?" — needs a content/topic property on each item.
3. "What came before Y in my reading?" — needs temporal ordering (read-on-date).

**Step 1:** Draft the corpus. Example scaffold (fill in with real-feeling titles and topics):

```n4l
- Reading list context

 "Thinking Fast and Slow" (is about) dual-process cognition
 "Thinking Fast and Slow" (cites) "Judgment under Uncertainty"
 "Thinking Fast and Slow" (read on) 2024-03-15

 "Judgment under Uncertainty" (is about) heuristics
 "Judgment under Uncertainty" (read on) 2024-04-02

 "The Dictator's Handbook" (is about) political coalitions
 ...
```

**Step 2:** Ingest and run the three target queries to verify the corpus supports the value prop.

```bash
cd /home/msvoboda/github/SSTorytime/src
# adjust paths per GettingStarted.md's install output
./bin/N4L -u ../examples/reading-list.n4l
./bin/searchN4L "what connects 'Thinking Fast and Slow' and 'The Dictator's Handbook'"
./bin/searchN4L "what is about heuristics"
./bin/pathsolve "Judgment under Uncertainty" -fwd -depth 3
```

**Expected:** Each query returns a non-trivial, readable result. If query #1 returns nothing, add a linking concept; if query #3 returns nothing, add more `read on` entries.

**Step 3:** Once the corpus is validated, commit.

```bash
git add examples/reading-list.n4l
git commit -m "feat(examples): add reading-list.n4l — running example for user-facing docs"
```

### Task B2: Rewrite `index.md` (Home / hero)

**Files:**
- Modify: `pages/docs/index.md` (full rewrite)
- Reference voice: `pages/docs/concepts/why-semantic-spacetime.md` (already Burgess-register post-audit)

**Goal:** 30-second triage. Reader arrives, and within 30 seconds knows:
1. **What this is** (in one sentence, no jargon)
2. **Why they might care** (value-add, not feature list)
3. **Whether it's for them** (writer? queryer? evaluator? — pointer into each track)
4. **One piece of proof** (one query screenshot or code block using the reading-list corpus)

**Shape suggestion:**

```
# SSTorytime

A one-paragraph honest answer to "what is this?" Mark's voice. Not:
"knowledge graph database" — that's the market-survey framing we're
killing. More like: "A way to keep notes that remember how they're
connected, so you can ask questions of them later."

## What you can do with it

Three bullets, each a job-to-be-done:
- Capture what you know, with the connections
- Ask questions that follow those connections
- Keep the shape of your thinking

## Is this for you?

- If you want to capture → [Writing Stories]
- If you have a corpus already → [Asking Questions]
- If you want to see why → [Concepts]

## One concrete example

Show the reading-list corpus being queried. Three lines of N4L, one
query, one answer. No install required to read the page.

## Install

Single link to Start Here → Install in 5 minutes.
```

**Voice rules** (from the design doc — tape to the monitor):
1. Outcome before mechanism.
2. No `file:line` citations.
3. No flag tables.
4. Thread the reading-list corpus.
5. 30-second bail-out test.
6. No internals named.
7. Show, don't define.

**Step 1:** Draft the page.

**Step 2:** Run `mkdocs serve -f pages/mkdocs.yml` and view at http://localhost:8000. Read it as a stranger. Does the page answer the three triage questions in 30 seconds? Iterate until yes.

**Step 3:** `mkdocs build --strict` must still pass.

**Step 4:** Commit.

```bash
git add pages/docs/index.md
git commit -m "docs(home): rewrite index.md in user register — 30-sec triage + value-add"
```

### Task B3: Rewrite `GettingStarted.md` → "Install in 5 minutes"

**Files:**
- Modify: `pages/docs/GettingStarted.md` (full rewrite)
- Also update: `pages/mkdocs.yml` nav title if desired — the nav already shows "Install in 5 minutes"; the H1 in the file should match

**Goal:** One happy path to a working install. No branching. No "if you're on macOS, if you're on Linux, if you prefer Docker ..." — pick one (Linux + source) and say clearly at the top "this page assumes Linux; Mac and Docker notes are in the developers folder".

**Shape:**

```
# Install in 5 minutes

One-sentence opener: what you'll have at the end.

Check you have: go 1.21+, postgres 15+, git. (One `command; # expected output` per row.)

Step 1: Get the code. (Two commands.)
Step 2: Set up the database. (Three commands, using `pg_config --bindir` detection, mode-600 creds.)
Step 3: Build. (One command.)
Step 4: Verify — ingest and query the reading-list corpus. Link to the actual commands using `examples/reading-list.n4l` from Task B1.

> If you hit a problem, see [When things go wrong](...). If you want the
> full build matrix (Mac, Docker, cross-compile, ...), see [developers/build-system.md](...).

Next: [Your first story →](Tutorial.md)
```

**Voice rules apply** — same seven. Don't over-explain what each command does; pick a working path and trust the reader.

**Step 1:** Draft.

**Step 2:** **Actually run the steps on a clean machine if possible** (or a clean worktree — `git clean -fdx` a new clone). If a command fails or produces unexpected output, fix the page, don't paper over. The user's first impression is load-bearing.

**Step 3:** `mkdocs build --strict`.

**Step 4:** Commit.

```bash
git add pages/docs/GettingStarted.md pages/mkdocs.yml
git commit -m "docs(start): rewrite GettingStarted as 'Install in 5 minutes' — one happy path"
```

### Task B4: Rewrite `Tutorial.md` → "Your first story"

**Files:**
- Modify: `pages/docs/Tutorial.md` (full rewrite)
- Reference: `examples/reading-list.n4l` (from Task B1)

**Goal:** 15-minute narrative. The user types (or copies) the reading-list corpus, ingests it, runs one query, sees the aha. Ends with "you just built a knowledge graph; here's what to do next."

**Shape:**

```
# Your first story

One-sentence opener: what you'll have at the end (a queryable graph of
your own reading list, plus a feel for how N4L thinks).

## The story we'll tell (2 min)

Introduce the reading-list corpus. What it is. Why it's a good first
example. One sentence each for 2–3 of the entries.

## Writing it down (5 min)

Walk through the N4L file section by section. Don't dump the whole file
at the top. Build it up:
  - First: one book and what it's about.
  - Then: a second book and its topic.
  - Then: a link between them via a shared concept.
  - Then: temporal context — who read what when.

Each section ends with "here's what your graph now contains". One
sentence of payoff.

## Loading it (1 min)

One `N4L -u examples/reading-list.n4l` command.
Expected output — show the actual output.

## Asking a question (5 min)

Three small queries against the corpus. For each:
  - The question in plain English.
  - The query command.
  - The answer, with a line or two reading the output as a sentence.

At least one query is "what connects X and Y" — the aha.

## What just happened

Three-bullet recap: you wrote it, you loaded it, you asked it. No code,
no API, no database.

## Where to go next

- Add more to your story → Writing Stories
- Ask bigger questions → Asking Questions
- Understand why it works → Concepts
```

**Voice rules apply.** Especially rule 1: outcome before mechanism, on every sub-section.

**Step 1:** Draft.

**Step 2:** Run the tutorial yourself end-to-end. If the query output doesn't read as an aha, go back to the corpus and fix it.

**Step 3:** `mkdocs build --strict`.

**Step 4:** Commit.

```bash
git add pages/docs/Tutorial.md
git commit -m "docs(start): rewrite Tutorial as 'Your first story' — 15-min narrative w/ reading-list corpus"
```

### Task B5: Phase B exit — push and request review

**Step 1:** Build-strict one more time; open `mkdocs serve` and click through Home → Install → Your first story as a stranger.

**Step 2:** Push to `fork/docs/phase9-user-facing` (or whichever branch name was chosen at Phase A exit).

**Step 3:** Present to Mark (and Mike) for voice review. **Do not dispatch Phase C until voice is approved.** This is the non-negotiable checkpoint from the design doc.

Suggested review ask:

> Phase B is complete — index.md, Install in 5 minutes, Your first story. These three pages set the voice for the whole user-facing site. Please read as a stranger and tell me:
> 1. Does the Home page answer "is this for me?" in 30 seconds?
> 2. Does "Install in 5 minutes" actually take 5 minutes on your machine?
> 3. Does "Your first story" land the aha on the last query?
> 4. Any single sentence that jars against the voice you want?
>
> If yes/yes/yes/none, I dispatch the parallel Phase C tracks.

---

## Phase C — Tracks (parallel, 2 agents, ~4h)

**Exit criterion:** Writing Stories and Asking Questions tabs fully populated in user register; `mkdocs build --strict` clean; the reading-list corpus threads through every page.

**Dispatch pattern:** Two agents in isolated git worktrees, strict file ownership, same pattern as the 8-agent audit (see `2026-04-20-documentation-upleveling.md` for precedent). Use `superpowers:using-git-worktrees` to create them and `superpowers:dispatching-parallel-agents` to run them.

### Task C1: Dispatch Writer agent

**Files owned:** `pages/docs/arrows.md`, `pages/docs/N4L.md` (user half; developer half — grammar, parser — moves to `pages/docs/developers/N4L-grammar.md` as part of this task), `pages/docs/text2N4L.md`, `pages/docs/cookbooks/bring-your-own-corpus.md` (retitled "Patterns — research notes" inline), one new file `pages/docs/writing-stories/chapters-and-context.md`.

**Task brief (pass to agent verbatim):**

> You are the Writer agent for the SSTorytime user-facing restructure.
>
> Read these first, in order:
> 1. `pages/docs/plans/2026-04-21-user-facing-restructure-design.md` — your charter
> 2. `pages/docs/plans/2026-04-21-user-facing-restructure-plan.md` — Task C1 (your task) is in this file
> 3. `pages/docs/index.md`, `pages/docs/GettingStarted.md`, `pages/docs/Tutorial.md` — the three voice anchors. Your pages must read in the same register.
>
> Your owned files are listed above. Do not touch any other file under `pages/docs/`.
>
> For every page, apply the seven voice rules from the design doc. Thread `examples/reading-list.n4l` through your examples — same corpus the anchor three use.
>
> `arrows.md` specifically: strip the 7-channel encoding callout (move it to `developers/arrows-grammar.md` if it needs a home). The user page teaches four arrow types: contains, leadsto, expresses, near. Nothing more.
>
> `N4L.md` specifically: split. User half stays as `N4L.md` under Writing Stories and teaches "here's how you write down a story so SSTorytime can load it". Developer half (grammar rules, parser flags, edge cases) moves to `pages/docs/developers/N4L-grammar.md`.
>
> Exit: `mkdocs build --strict -f pages/mkdocs.yml` clean. One commit per file (per the TDD-for-docs pattern: draft → build-strict → render-check → commit). Final commit message on branch: `docs(writing): Phase C writer-track complete`.

### Task C2: Dispatch Queryer agent

**Files owned:** `pages/docs/searchN4L.md`, `pages/docs/pathsolve.md`, `pages/docs/search_examples.md` (consolidate into `searchN4L.md` or delete), `pages/docs/howdoescontextwork.md`, `pages/docs/cookbooks/search-recipes.md` (retitled "Patterns — search recipes" inline).

**Task brief (pass to agent verbatim):**

> You are the Queryer agent for the SSTorytime user-facing restructure.
>
> Read these first, in order:
> 1. `pages/docs/plans/2026-04-21-user-facing-restructure-design.md` — your charter
> 2. `pages/docs/plans/2026-04-21-user-facing-restructure-plan.md` — Task C2 (your task) is in this file
> 3. `pages/docs/index.md`, `pages/docs/GettingStarted.md`, `pages/docs/Tutorial.md` — the three voice anchors. Your pages must read in the same register.
>
> Your owned files are listed above. Do not touch any other file under `pages/docs/`.
>
> `searchN4L.md` specifically: the current page is a CLI reference with flag tables. Rewrite as "Finding things" — teach the shape of a question (what/who/how/when/where), show a few queries against `examples/reading-list.n4l`, point at `--help` for the flag reference. Consolidate `search_examples.md` into this page.
>
> `pathsolve.md` specifically: rewrite as "Finding paths between things". Lead with the value ("what connects X and Y") and the canonical reading-list query ("what connects these two books?"). Flag reference goes to `developers/pathsolve-flags.md` if it needs a home.
>
> `howdoescontextwork.md` specifically: rewrite as "Context — asking the same question different ways". The current page has the "Precedence" admonition from the audit — preserve its *content* but reframe its *voice*.
>
> Exit: `mkdocs build --strict -f pages/mkdocs.yml` clean. One commit per file. Final commit message on branch: `docs(asking): Phase C queryer-track complete`.

### Task C3: Merge the two worktrees

**Step 1:** Merge Writer and Queryer branches back to `docs/phase9-user-facing`. Non-overlapping file ownership means conflicts should be zero; if any, resolve per file-owner intent.

**Step 2:** `mkdocs build --strict`.

**Step 3:** Manual click-through via `mkdocs serve`. Click every nav link in Writing Stories and Asking Questions. Any 404 or wrong-register page is a blocker.

---

## Phase D — Concepts + polish (1 agent, ~2–3h)

**Exit criterion:** Concepts tab + Living With Your Graph tab populated; FAQ in user register; final link audit clean; Phase D commit pushed.

### Task D1: Concepts overview (replacing deleted `architecture.md`)

**Files:**
- Create: `pages/docs/concepts/index.md` OR extend `pages/docs/concepts/why-semantic-spacetime.md` with a front-matter overview section

**Goal:** One-page orientation for the Concepts tab. Two paragraphs + three bullets pointing into the four sub-pages (why / glossary / arrows / context).

### Task D2: Essay voice pass

**Files:**
- Modify: `pages/docs/Storytelling.md`, `pages/docs/KnowledgeAndLearning.md`, `pages/docs/outreach.md`

**Goal:** Light touch only. These are Mark's essays; don't restructure. Look for (and fix) places where a sentence assumes the reader has been through the dev-facing nav (stale cross-references, "as we saw in the API reference", etc.).

### Task D3: Living With Your Graph

**Files:**
- Modify: `pages/docs/cookbooks/backup-restore-delete.md` → retitle "Backing up, restoring"
- Create: `pages/docs/living/sharing-a-graph.md` (can be short — 20 lines is fine; what files to hand someone else)
- Create: `pages/docs/living/when-things-go-wrong.md` (short — common errors + where to look)

**Goal:** Tiny but present. These pages exist so the tab isn't a one-link tab. They can grow later; starting short is fine.

### Task D4: FAQ rewrite

**Files:**
- Modify: `pages/docs/FAQ.md`

**Goal:** User-register voice pass. If any question is about internals (e.g., "how does the cache work"), move it to `developers/FAQ-internals.md`.

### Task D5: Final link audit and push

**Step 1:** Full `mkdocs build --strict`. Zero warnings is the target.

**Step 2:** `mkdocs serve` and click every nav link. Every link in every page body. This is tedious; do it.

**Step 3:** Push `docs/phase9-user-facing` to fork; either open a fresh PR or append to PR #57.

**Step 4:** Present to Mark.

---

## Execution notes

**Worktrees:** Use `superpowers:using-git-worktrees` for Phase C. Phase A is mechanical enough to run in the main workspace. Phase B should run in a worktree if you want to keep main clean while the voice gets iterated.

**Rebases:** If upstream `markburgess/main` moves during B/C/D, rebase at phase boundaries only, not mid-phase. An in-flight Phase C with partial agent work is painful to rebase.

**Commit cadence:** One commit per file for content pages (Phase B/C/D). One commit per logical structural change for Phase A (scaffold, moves+deletes, nav). Matches the design doc's "frequent commits" principle.

**Build verification:** `mkdocs build --strict -f pages/mkdocs.yml` is the equivalent of "tests pass". Run it after every change. Do not batch a day's changes and then build — you'll spend the evening bisecting.

**Voice review:** The design doc's seven rules aren't decorative; they're the pass/fail bar. If a page violates rule 2 (file:line cites), the correct fix is to delete the cite, not to "justify" it. Same for the others.

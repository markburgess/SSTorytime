# SSTorytime User-Facing Documentation Restructure — Design

> **For Claude:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement the eventual plan derived from this design. This is the *design* doc; the implementation plan will live alongside as a sibling `-plan.md` once authored.

**Goal:** Re-shape the published `/pages/docs/` site from its current developer/operator-facing posture into a user-facing guide. Simplicity and value-add over code internals. Driven by direct feedback from the project maintainer (Mark Burgess).

**Constraint:** The pristine upstream `/docs/` tree (identical to `markburgess/main`) stays untouched. All changes happen under `/pages/`. Developer-facing content is preserved in the repo but excluded from the built site via `exclude_docs`.

**Baseline:** The documentation landed in PR #57 (`docs/phase3-8-content` on `mikesvoboda/SSTorytime`, rebased onto upstream `main` at commit `953a732` on 2026-04-21). That PR is the **developer-facing** version. This design is the **user-facing** successor, intended to ship as either a follow-up PR on the same branch or a fresh branch stacked on top.

---

## Design decisions locked by Mike (2026-04-21)

1. **Primary user** = evaluator front door + two parallel tracks (writer / queryer). Hobbyists and researchers fall out as derivatives of those two modes.
2. **First user win** = "Your first story" — user writes ~15 lines of N4L about something they know, ingests, asks one question, sees a meaningful answer. Supplemented by a 2-minute "try the examples" warmup on the landing page.
3. **Running example** — one small reading-list / research-notes corpus (~20 lines of N4L) threaded through `index → Install → First story → Writing patterns → Asking patterns`. Authored fresh for this restructure.
4. **Developer content disposition** — excluded from the site via `exclude_docs`, kept in the repo under `/pages/docs/developers/` so contributors can still find it. A single footer note links out.
5. **Voice baseline** — Burgess register (established in `why-semantic-spacetime`, glossary Wave-front/Orbit/Story during the 8-agent audit) plus seven user-facing rules below.
6. **Phase B (anchor three) is a review checkpoint** for Mark before Phase C dispatches. Voice has to cohere across the whole site; getting the three anchor pages right de-risks everything after.

---

## Problem statement

The existing `/pages/docs/` tree is structured for developers, operators, and contributors:

- Twelve top-level nav tabs, of which six (`APIs`, `Operations`, `Reference`, `Tools`, `Contributing`, `Project`) are explicitly developer-coded.
- Load-bearing front-door pages (`index.md`, `Tutorial.md`, `GettingStarted.md`) lead with stack diagrams, `go install`, and `psql` invocations — not with "what is this and what will it do for me?"
- Reference pages (CLI flag tables, stored-functions catalog, OpenAPI, schema ER diagram) dominate the sidebar and push user-relevant content below the fold.
- The running examples are fragmented across `example.md`, `API_WALKTHROUGH.md`, and six cookbooks, none of which thread continuously.

Result: an evaluator hits the landing page and can't answer "is this for me?" in under 30 seconds. A writer can't find a path from "I have an idea to capture" to "I ran one query against it." A queryer can't find a path from "I have a corpus" to "I got an answer."

---

## Target users and first win

**Front door (evaluator).** Someone who has heard of semantic spacetime, knowledge graphs, or this project specifically, and is asking "what is this, why would I use it, is it for me?" They may never install. Job-to-be-done of the landing page: *answer that triage question in 30 seconds, then convert into one of the two tracks.*

**Track A — writer / knowledge-keeper.** Wants to capture notes, research, a personal knowledge base, or a story as a semantic graph. Writes N4L files by hand or via `text2N4L`. Doesn't code. Job-to-be-done: *my idea becomes a graph I can ask questions of.*

**Track B — queryer / researcher.** Has a corpus (ingested already, or via an example file) and wants to explore relationships, context, and paths. Cares about insights, not about how the graph was built. Job-to-be-done: *the graph answers a question I couldn't easily answer any other way.*

**First win = "Your first story."** ~15 minutes, one N4L file (the reading-list corpus), one ingest, one query, one aha. Demonstrates capture + query end-to-end. The user feels ownership because they typed their own data (even if it's the canonical reading list we provide as a template).

---

## Proposed nav

```
Home                       ← hero: what is this, why, "is it for me?"

Start Here
    Install in 5 minutes   ← one-page, happy-path only, no branching
    Your first story       ← the 15-min narrative: one N4L file, one query
    Try the examples       ← 2-min warmup: ingest + query included data

Writing Stories            ← Track A: for people capturing knowledge
    Thinking in arrows
    Writing N4L by hand
    Turning documents into stories      (text2N4L, reframed)
    Organising with chapters & context
    Patterns: family trees, research notes, meeting logs, trips

Asking Questions           ← Track B: for people querying a graph
    The shape of a question
    Finding things (search)
    Finding paths between things
    Context: asking the same question different ways
    Patterns: "what led to X", "who connects these two", "what's near this"

Living With Your Graph     ← light-touch ops
    Backing up, restoring
    Sharing a graph
    When things go wrong

Concepts                   ← the "why" — optional deep reads
    Semantic spacetime in plain English
    The four arrows
    Context as a first-class citizen
    Storytelling & knowledge (Mark's essays)

For Developers             ← single link out + footer note
    (points to /pages/docs/developers/ in the repo)
```

Seven tabs, down from twelve. The persona split (Writing / Asking) is the spine. "Cookbooks", "Operations", "Reference", "Tools", "Contributing", "Project" all dissolve into the two user tracks or move to the developers folder.

---

## Disposition of the existing ~40 pages

Three buckets:

### Keep & reframe (~22 pages — stay on the site)

| Current path | Becomes |
|---|---|
| `index.md` | Home (hero) — full rewrite |
| `GettingStarted.md` | Start Here → Install in 5 minutes — full rewrite |
| `Tutorial.md` | Start Here → Your first story — full rewrite |
| `FAQ.md` | Home → FAQ — user-register pass |
| `arrows.md` | Writing Stories → Thinking in arrows — strip 7-channel encoding internals |
| `N4L.md` | Writing Stories → Writing N4L by hand — user half only; ref half moves to developers/ |
| `example.md` | Writing Stories → Patterns: (one of) — fold in |
| `text2N4L.md` | Writing Stories → Turning documents into stories — reframe from CLI reference to "when and why" |
| `searchN4L.md`, `search_examples.md`, `pathsolve.md` | Collapse into Asking Questions → Finding things / Finding paths between things |
| `howdoescontextwork.md` | Asking Questions → Context: asking the same question different ways |
| `concepts/why-semantic-spacetime.md` | Concepts → Semantic spacetime in plain English — already Burgess-voiced post-audit |
| `concepts/glossary.md` | Concepts → Glossary (inline admonition links from other pages) |
| `Storytelling.md`, `KnowledgeAndLearning.md`, `outreach.md` | Concepts → Mark's essays — minor voice edits, no restructure |
| `cookbooks/bring-your-own-corpus.md` | Writing Stories → Patterns: research notes — absorb |
| `cookbooks/search-recipes.md` | Asking Questions → Patterns — absorb |
| `cookbooks/backup-restore-delete.md` | Living With Your Graph → Backing up, restoring |

### Move to `/pages/docs/developers/` + `exclude_docs` (~18 pages — kept in repo, off the site)

- **Go / HTTP API:** `API.md`, `API_WALKTHROUGH.md`, `WebAPI.md`, `http-api/mcp-sst.md`, `http_server.md`
- **Database internals:** `Database/Schema.md`, `Database/Functions.md`, `Database/Indexes.md`, `Database/Performance.md`, `Database/Setup.md`
- **Internals & reference:** `dynamic_functions.md`, `graph_report.md`, `notes.md`, `removeN4L.md`, `code-index.md`, `debug.md`, `namespaces.md`
- **Build / ops / contributor:** `build-system.md`, `testing.md`, `versioning.md`, `tls-certificates.md`, `operations/observability.md`, `howtocontribute.md`
- **Dev cookbooks:** `cookbooks/first-go-program.md`, `cookbooks/python-integration.md`, `cookbooks/llm-via-mcp-sst.md`

### Delete (~3 pages)

- `ToDo.md` — project hygiene, not documentation
- `concepts/architecture.md` — dev-framed stack diagram; replaced by a short user-facing "Concepts" overview
- `README.md` — stub banner, no content

---

## Voice principles (seven rules)

1. **Outcome before mechanism.** "You'll be able to ask *what led to X*" — not "N4L uses channel-encoded arrows".
2. **No `file:line` citations in body.** Those anchored the 8-agent audit work; user pages don't need them.
3. **No flag tables in body.** Point readers at `--help` or the developers folder.
4. **One concrete running example threaded through the whole site** (the reading-list corpus).
5. **Every page opens with "why does this matter to me?"** — 30-second bail-out test. If the reader can't tell within 30 seconds whether this page is for them, the opening is wrong.
6. **Don't name internals.** No `ScoreContext`, `NODE_CACHE`, `GraphToDB`, `HubJoin`, etc. If a concept only makes sense by naming one, it's a developers-folder page.
7. **Show, don't define.** Introduce concepts via what the reader does with them, not glossary-style.

---

## The running example — reading-list corpus

One small `.n4l` file (~20 lines) authored fresh for this restructure. Concept: a personal reading list — a handful of books/papers, what each one is about, what it cites, who read it when.

Why this corpus:
- Universally relatable across writer / queryer / hobbyist / researcher personas.
- Demonstrates the project's value in one query ("what connects these two papers?") — something harder to do in SQL or a flat note-taking app.
- Avoids the domain-specificity of `chinese.n4l` (language-learning flavor limits audience) and the contrived feel of `Mary.n4l` (family-tree toy).
- Small enough to display in full on the "Your first story" page, big enough for the Asking Questions track to demonstrate multiple query shapes.

The same corpus appears on:
- Home (screenshot of one query result in the hero)
- Install in 5 minutes (final verification step queries it)
- Your first story (built up line by line)
- Writing Stories → Patterns: research notes (variant extensions)
- Asking Questions → every page (different query angles)

A new file `examples/reading-list.n4l` is authored as part of Phase B.

---

## Sequencing — four phases

### Phase A — Structure (mechanical, ~1h, 1 agent)

- Rewrite `mkdocs.yml` nav to the 7-tab shape above.
- `git mv` ~18 files into `/pages/docs/developers/`.
- Update `exclude_docs` to add `developers/*`.
- Delete the three stale files.
- Add a footer note in `mkdocs.yml` pointing at `developers/` for contributors.
- Update any cross-links broken by the moves (mechanical; grep-able).
- `mkdocs build --strict` passes.

**Output:** site *shape* is user-facing; content still reads dev-flavored. Safe to ship alone as an interim commit.

### Phase B — Anchor three (sequential, 1 agent, ~4–8h) — **REVIEW CHECKPOINT**

Rewrite in order:
1. `index.md` — the hero. What is this, why, is it for you, the 30-second triage.
2. `GettingStarted.md` → Install in 5 minutes. Happy path only. No branching. Ends with "now go to Your first story".
3. `Tutorial.md` → Your first story. The 15-minute narrative. Introduces the reading-list corpus. One N4L file, one ingest, one query, one aha.

Also authors `examples/reading-list.n4l`.

**Non-negotiable checkpoint:** Mark reviews the three pages before Phase C dispatches. Voice has to cohere across the whole site; if these three are wrong, everything built on top inherits the error.

### Phase C — Tracks (parallel, 2 agents, ~4h)

Strict file-ownership split (same pattern as the 8-agent audit):

- **Writer agent** — rewrites Writing Stories tab (`arrows.md`, user half of `N4L.md`, `text2N4L.md`, context/chapters page, patterns).
- **Queryer agent** — rewrites Asking Questions tab (`searchN4L.md`, `pathsolve.md`, `search_examples.md` consolidation, context page, patterns).

Both agents thread the reading-list corpus through their examples.

### Phase D — Concepts + polish (1 agent, ~2–3h)

- Concepts pages: light rewrite of `why-semantic-spacetime` (already Burgess-voiced), glossary audit, new short "Concepts" overview replacing `architecture.md`.
- Essay pages (`Storytelling`, `KnowledgeAndLearning`, `outreach`): voice pass only, no restructure.
- Living With Your Graph: `backup-restore-delete.md` reframe, new "Sharing a graph" and "When things go wrong" stubs (may be short).
- FAQ rewrite.
- Final link audit. `mkdocs build --strict` clean.

---

## Risk flags

1. **Phase B is load-bearing.** If the anchor three set the wrong voice, Phase C agents inherit the error and we rewrite twice. Mitigation: hard checkpoint with Mark before Phase C dispatches.
2. **Developer content drift.** Moving the developer pages to `developers/` without changes means cross-links from the old nav will 404 on the deployed site until Phase A's link audit is complete. Mitigation: link audit is a Phase A exit criterion; `mkdocs build --strict` enforces it (strict fails on broken internal links).
3. **Upstream rebase churn.** Upstream `main` is active (7 commits in the last day). A Phase B/C rewrite that takes a week could eat multiple rebases. Mitigation: land Phase A as a small standalone commit/PR immediately; do B/C/D on a fresh branch stacked on top so rebases affect at most one phase at a time.
4. **Reading-list corpus is unvalidated.** We haven't actually authored it yet or tested that a 20-line file demonstrates the value proposition. Mitigation: Phase B starts by authoring `examples/reading-list.n4l` and running the three target queries against it — if the corpus doesn't yield a good aha query, pivot before writing the "Your first story" page.
5. **`N4L.md` split.** Splitting N4L reference into "user half" (writing by hand) and "dev half" (grammar / parser internals) may leave the dev half orphaned in `developers/`. Acceptable — it's for contributors, not readers — but worth flagging in the commit message.

---

## What this design does **not** cover

- The implementation plan itself. That's a sibling doc (`*-plan.md`) once this design is validated, produced via `superpowers:writing-plans`.
- The copy of the reading-list corpus. Authored at the start of Phase B against the target queries.
- Visual changes (new mermaid diagrams, new AI images). Phase B may commission one new hero image for the Home page; otherwise reuse the existing 18 AI-generated assets.
- Changes to the pristine upstream `/docs/` tree. Out of scope — that tree remains byte-for-byte identical to upstream `main`.
- Developer-site production. The `/pages/docs/developers/` folder is *repo-only*, not a separate published MkDocs site. If a separate dev site is wanted later, it's a follow-up design.

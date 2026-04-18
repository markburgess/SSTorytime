---
title: "feat: n4l:md-import — structure-first / arrow-second markdown→.n4l skill"
type: feat
status: active
date: 2026-04-18
origin: docs/brainstorms/2026-04-18-n4l-md-import-requirements.md
---

# feat: n4l:md-import — structure-first / arrow-second markdown→.n4l skill

## Overview

Add a new Claude skill `n4l:md-import` to the SSTorytime n4l-plugin that converts a single markdown file into a valid `.n4l` knowledge-graph file. The skill uses a **two-pass prompt-based architecture**: Pass 1 deterministically walks markdown structure (headings, lists, tables, links, blockquotes, code fences, frontmatter) and emits a `.n4l` skeleton with generic placeholder arrows plus provenance comments; Pass 2 uses Claude judgment to specialize placeholders into specific `SSTconfig/` arrow codes. Re-runs detect hand-edits via content hashing and preserve user changes. The skill composes with the existing `n4l:upload` validate-fix-upload loop (see origin: `docs/brainstorms/2026-04-18-n4l-md-import-requirements.md`).

## Problem Frame

SSTorytime users with existing markdown notes currently have no structural path into the graph — the existing `n4l:import` handles CSV only, and `text2N4L` handles prose via sentence fractionation. Structured markdown (headings, lists, tables) is where most user knowledge already lives. The skill closes this gap while respecting the seed constraint "never block on questions": ambiguity surfaces as inline `# REVIEW:` comments rather than prompts.

## Requirements Trace

**Invocation and I/O (origin R1–R4)**
- R1. Accept single markdown file path: `/n4l:md-import notes.md` → Unit 1 (scaffold), Unit 2 (Pass 1)
- R2. Default output `<source-name>.n4l`; `--output` override → Unit 1
- R3. Marker-gated merge-or-abort on existing output; `--force` replaces → Unit 1, Unit 4 (merge)
- R4. Non-interactive; ambiguity only surfaces as `# REVIEW:` / `# SUGGEST:` comments → Units 2, 3

**Pass 1 — structural parse (origin R5–R9)**
- R5. Parse frontmatter, H1–H6, ul/ol lists, tables, links, blockquotes, code fences → Unit 2
- R6. Skip paragraph prose; emit hint when prose-bytes ratio crosses threshold → Unit 2, Unit 5
- R7. Node naming = stripped text, with context-stack-scoped identity for collision avoidance → Unit 2, Unit 5
- R8. Placeholder vocabulary = `contain`, `fwd`, `see`, `note` (all verified against SSTconfig; see Key Decisions) → Unit 2
- R9. Provenance comment `# src:<path>:L<line> h:<hash>` on every edge → Unit 2

**Pass 2 — Claude arrow specialization (origin R10–R13)**
- R10. Claude specializes each placeholder using skeleton + heading breadcrumb + SSTconfig candidates → Unit 3
- R11. `# REVIEW:` when multiple arrows plausible; keep placeholder + `# SUGGEST:` when none better → Unit 3
- R12. SSTconfig lookup: CWD → `$SSTORYTIME_HOME` → none (no prompting) → Unit 1, Unit 3
- R13. Hash computed over Pass-1 output (placeholder arrow), never Pass-2 specialized form → Unit 2, Unit 3

**Output structure (origin R14–R15)**
- R14. Trimmed top-of-file marker block: source path + conditional R6/R12 warnings (versioning fields cut per review) → Unit 2
- R15. Chapter + context blocks, ditto-aligned, blank-line-separated → Unit 2

**Merge on re-run (origin R16–R18)**
- R16. Reframed as **two-way hash-based merge** (old-.n4l vs regenerated skeleton) per review P0 → Unit 4
- R17. **Tuple-based edit/removal detection** (not line-number-based) per review P0 → Unit 4. *Tuple = (normalized-node-text, arrow-meta-type, normalized-target-text)* — used to identify whether a source still exists in the markdown; see Key Technical Decisions for the full specification.
- R18. **REMOVED from v1** per review P0. The origin's three-way merge with `.n4l.backup` fallback is not implemented: line-number-only provenance cannot reconstruct old markdown, so the three-way framing is infeasible without persisting old md as a sidecar (deferred, see Scope Boundaries). R17's four-state detection covers all cases via tuple matching.

**Integration (origin R19–R20)**
- R19. Output valid for `n4l:upload`; skill does NOT invoke upload in v1 → Unit 2, Unit 6
- R20. Per-skill `./references/` copies of `n4l-syntax.md` and `arrow-types.md` (verified byte-identical across the 5 skills that carry them: import, interpret, learn, scaffold, upload. The plugin has 11 skills total; the other 6 don't carry per-skill references. `md-import` follows the 5-skill convention.) → Unit 1

## Scope Boundaries

Preserved from origin plus two review-driven additions:

- **Single file only.** Vault bulk import / wikilink resolution is ideation 6a.3, deferred.
- **No learning loop.** Pass 2 is stateless; author profile harvesting is ideation 6a.5.
- **No pipeline composition.** `--pipe` to upload/interpret is ideation 6a.6.
- **No signature recipes.** `glossary-list`, `academic-paper` presets are ideation 6a.7.
- **No prose fractionation.** `text2N4L` remains the tool for prose content.
- **No SSTconfig modification.** Read-only.
- **No database interaction.**
- **No wikilinks or Obsidian `#tags`.** Literal text in v1.
- **No interactive prompts.** Per seed constraint.
- **Two-way merge, not three-way** (*new, from review*). Full three-way would require persisting old md as a sidecar; deferred until a concrete failure mode demands it.
- **Versioning fields in top-of-file marker omitted** (*new, from review*). Added only when a second output format version exists and a migration is needed.

## Context & Research

### Relevant Code and Patterns

- `n4l-plugin/plugins/n4l/skills/import/SKILL.md` — canonical template. Mirror: frontmatter (`name`, `description`, `argument-hint`), `## Reference Material` with `@./references/...` imports, numbered `## Workflow` steps, Step 1 = argument parsing, final step = Write tool + summary report.
- `n4l-plugin/plugins/n4l/AGENTS.md` — plugin-wide conventions including SSTconfig lookup order.
- `n4l-plugin/plugins/n4l/skills/{import,scaffold,upload,interpret,learn}/references/{n4l-syntax.md,arrow-types.md}` — verified byte-identical across all 5 skills (md5 checksums match). Pattern: per-skill copies, not shared path.
- `SSTconfig/arrows-{NR-0,LT-1,CN-2,EP-3}.sst` — arrow definition files. Placeholder verification hits:
  - `SSTconfig/arrows-CN-2.sst:9`: `+ contains (contain) - belongs to (belong)` ✓
  - `SSTconfig/arrows-LT-1.sst:17`: `+ leads to (fwd) - comes from (bwd)` ✓
  - `SSTconfig/arrows-NR-0.sst:48`: `see also (see)` ✓
  - `SSTconfig/arrows-EP-3.sst:10`: `+ note/remark (note) - is a note or remark about (isnotefor)` ✓
- `examples/knowledge.n4l` lines 37–49 — canonical example of `+:: _sequence_ ::` nested under `:: getting started ::`. Confirms R5 sequence-nesting assumption.
- `examples/openshift.n4l:40` — additional sequence-mode example.
- `src/N4L/N4L.go:396, 1327, 1479` — parser's `#` comment handling. Both full-line and trailing-comment positions accepted.
- `docs/text2N4L.md` — confirmed non-overlapping scope (sentence fractionation binary, not structural parsing).

### Institutional Learnings

- `docs/ideation/2026-04-17-n4l-claude-skills-ideation.md` 6a.1 — original ideation for this two-pass architecture. 6a.4 (parser-error back-propagator, future skill) will consume the same provenance comment format — **lock the format now as a cross-skill interface.**
- `docs/brainstorms/2026-04-17-n4l-skills-plugin-requirements.md` — earlier plugin brainstorm convention `SSTconfig/ required with prompt fallback` is superseded; this skill follows shipped `n4l:import` convention (silent generic fallback with warning comment).
- `docs/plans/2026-04-17-001-feat-n4l-skills-plugin-plan.md` — existing plan for the base plugin; md-import adds a sibling skill without touching shared plumbing.
- `docs/solutions/` does not exist — no prior documented learnings database.

### Gaps (no prior art in repo)

- **Content hashing for edit detection** — novel mechanism; no existing skill does this.
- **Two-way (formerly "three-way") merge with provenance** — novel; future-skill cross-interface.
- **Markdown → .n4l structural parsing** — novel entirely.

## Key Technical Decisions

- **Placeholder vocabulary is REAL SSTconfig codes.** Verified during planning: all four (`contain`, `fwd`, `see`, `note`) appear in `SSTconfig/arrows-{CN-2,LT-1,NR-0,EP-3}.sst`. The brainstorm P0 finding is resolved. **Decision:** no runtime validation needed in v1 beyond the existing R12 lookup-failure warning. Planning-time verification is documented here; if a fork of SSTconfig removes any, the user gets the existing no-SSTconfig warning and the output still parses because N4L's parser accepts any arrow token.
- **Merge is two-way, not three-way.** Review P0 established that line-number-only provenance cannot reconstruct old markdown. Rather than persist old md as a sidecar (new state management surface), drop the three-way framing entirely. The four-state detection in R17 (hash-match / hash-mismatch / no-provenance / source-removed) is sufficient when source-removal is detected by **tuple matching**, not line numbers.
- **Edge identity tuple = (normalized-node-text, arrow-meta-type, normalized-target-text).** Used for: (a) Pass-1 hash computation (R9); (b) merge detection of "source line no longer in markdown" (R17). Line numbers in provenance become *informational only* (useful for parser-error back-propagation in future skill 6a.4), not load-bearing for merge correctness.
- **Node identity is scoped by context stack.** R7 collision policy: a node's full identity is `(context_stack, stripped_text)`. Two `## Usage` headings under different H1s produce distinct identities because their context stacks differ. Pass 1 emits nodes under their full `:: ctx1, ctx2 ::` block, so N4L's own context-scoped resolution does the work. For bare same-context collisions (two bullets with identical stripped text under the same heading), suffix with `(2)`, `(3)`. For link collisions `[Database](postgres.md)` vs `[Database](mysql.md)` in the same context, qualify as `Database [postgres.md]` vs `Database [mysql.md]` to preserve disambiguation.
- **Parser is prompt-based.** Pass 1 is Claude following SKILL.md instructions to walk the markdown (via `Read`) and emit the skeleton (via `Write`). No external binary, no goldmark, no python/node shell-out. This resolves the P1 parser-choice deferred question. Pass-1 determinism means "given the same markdown, the emitted skeleton is structurally identical run-to-run"; arrow specialization (Pass 2) has normal LLM non-determinism mitigated by hashing Pass-1 output (R13) not Pass-2 output.
- **Pass-2 chunking policy.** For skeletons under ~40k tokens estimated, single prompt over the whole skeleton. For larger files, chunk by top-level heading with one-heading overlap so context stacks align across chunks. Candidate arrow catalogs are always included per-chunk (four `SSTconfig/arrows-*.sst` files total ~24KB ≈ 6K tokens at ~280 entries — cheap to duplicate for modest chunk counts; at >~5 chunks consider restricting per chunk to the meta-types actually used).
- **Prose-bytes definition for R6 threshold.** Prose bytes = bytes in paragraphs NOT inside `{code fence, table, list item, blockquote, heading line, frontmatter block}`, excluding whitespace-only lines. Ratio = `prose_bytes / (prose_bytes + structural_bytes)`. Threshold retained at 50% but treated as *provisional*; tune against the Reference Corpus in Unit 6. This replaces the review-flagged "false precision" by making the metric testable and tunable.
- **Top-of-file marker = one line + conditional warnings.** Stripped per review: `# n4l:md-import :: src=<relative-md-path>` as the single marker line (used by R3 marker-check to identify skill-owned output). R6 and R12 warnings appear as separate `# NOTE:` lines below it only when triggered. Timestamp and version fields dropped.
- **Skill name committed to `n4l:md-import`.** Matches `n4l:import`, `n4l:upload` explicit-verb convention.
- **Per-skill references copies.** Verified convention; new skill ships its own `references/n4l-syntax.md` and `references/arrow-types.md` copies.

## Open Questions

### Resolved During Planning

- **Placeholder vocabulary real?** Resolved by grep against `SSTconfig/arrows-*.sst` — all four codes present.
- **Is `+:: _sequence_ ::` nestable under heading-derived `:: ::` contexts?** Resolved — `examples/knowledge.n4l:37-49` shows the exact pattern uploaded to production.
- **Parser execution model?** Resolved — prompt-based, Claude reads markdown directly.
- **Three-way vs two-way merge?** Resolved — two-way, with tuple-based source-removal detection.
- **Node-naming collision policy?** Resolved — context-stack scoping + `(N)` suffix for same-context duplicates + URL-qualifier for same-context link duplicates.
- **R14 top-of-file block scope?** Resolved — single marker line + conditional NOTE lines.
- **Skill name?** Resolved — `n4l:md-import`.

### Deferred to Implementation

- Exact prose-bytes threshold empirical tuning (50% is provisional; Unit 6 sets final value based on Reference Corpus behavior).
- Exact hash algorithm (6 hex chars of SHA-256 is the starting recommendation; Unit 2 confirms collision rate on realistic-size corpus).
- Pass-2 chunking threshold (40k tokens is the starting recommendation; Unit 3 confirms against representative skeletons).
- Fixture set for Reference Corpus (Unit 6 pins the exact file list drawn from `docs/brainstorms/`, `docs/ideation/`, `n4l_patterns/`).
- Exact form of normalized node/target text for tuple matching (case, whitespace, trailing punctuation) — specified at Unit 2 time against fixture behavior.

## High-Level Technical Design

> *This illustrates the intended approach and is directional guidance for review, not implementation specification. The implementing agent should treat it as context, not code to reproduce.*

### Pass-1 skeleton output shape (per feature)

```
# n4l:md-import :: src=notes.md
# NOTE: this file contains substantial prose; consider running text2N4L for sentence-level extraction.  # R6 conditional
# NOTE: SSTconfig/ not found. Placeholders kept. Set $SSTORYTIME_HOME or run from your project root.  # R12 conditional

- <frontmatter.title or filename-sans-ext>
:: <frontmatter.tags joined ::            # R5 frontmatter, bare ::

+:: <heading1 text> ::                    # R5 H1 opens
+:: <heading2 text> ::                    # R5 H2 opens (stacked)

<node1> (contain) <bullet-text>           # src:notes.md:L12 h:a3f4c2
"       (contain) <bullet-text-2>         # src:notes.md:L13 h:b8e501
"       (see)    <link-text>              # src:notes.md:L14 h:c19a42 url=https://example.com

+:: _sequence_ ::                         # R5 ordered-list wrapper
<step1>                                   # src:notes.md:L20 h:d4... (then chains implicit)
<step2>                                   # src:notes.md:L21 h:e9...
-:: _sequence_ ::

-:: <heading2 text> ::                    # close when next same-or-shallower heading appears
```

### Merge state machine (two-way, per edge)

| Condition on re-run | Action |
|---|---|
| Old-n4l edge's `h:` hash matches fresh Pass-1 recomputation of same tuple | Replace line with fresh Pass-1+Pass-2 output |
| Hash mismatch (user hand-edited arrow or text) | Preserve existing line verbatim, update provenance comment to append ` edited`, skip Pass-2 specialization |
| Edge has no `# src:` provenance | User-added edge; preserve verbatim, tag with ` user-added` on first merge |
| Fresh Pass-1 over current md produces no edge whose (normalized-node, arrow-meta-type, normalized-target) tuple matches the old edge's tuple | Emit old edge commented-out above a `# REMOVED: source no longer in markdown` comment |

### Pass-2 specialization prompt shape

Input: skeleton chapter-by-chapter, candidate arrows from the four `arrows-*.sst` files grouped by meta-type, heading breadcrumb carried in context stack. Output: same skeleton with each `(placeholder)` replaced by a specific `(code)` and optional trailing `# REVIEW:` or `# SUGGEST:` comment. Hashes unchanged (R13).

## Implementation Units

- [x] **Unit 1: Scaffold skill directory and SKILL.md**

**Goal:** Create the skill directory, SKILL.md skeleton, per-skill references, and argument-parsing logic. Establish the R3 marker-check and `--force` / `--output` flag handling.

**Requirements:** R1, R2, R3, R12, R20

**Dependencies:** None.

**Files:**
- Create: `n4l-plugin/plugins/n4l/skills/md-import/SKILL.md`
- Create: `n4l-plugin/plugins/n4l/skills/md-import/references/n4l-syntax.md` (copy of existing)
- Create: `n4l-plugin/plugins/n4l/skills/md-import/references/arrow-types.md` (copy of existing)
- Test: `n4l-plugin/plugins/n4l/skills/md-import/tests/fixtures/` (directory for acceptance fixtures)

**Approach:**
- Mirror the `n4l:import/SKILL.md` structure: frontmatter (`name: n4l:md-import`, `description:` with trigger phrases "import markdown", "md to n4l", "markdown to knowledge graph", `argument-hint: "<markdown-file> [--output <path>] [--force]"`), `## Reference Material` imports, numbered `## Workflow` sections.
- Step 1 = argument parsing. Output path defaults to swapping `.md` → `.n4l`.
- Step 2 = existing-file / marker check. If output exists and starts with `# n4l:md-import :: src=`, proceed to merge (Unit 4). If exists without marker, abort with "foreign file" error. `--force` skips both checks.
- Step 3 = SSTconfig locator (CWD → `$SSTORYTIME_HOME` → none; silent fallback with `# NOTE:` warning in output).
- Remaining steps delegate to Pass 1 (Unit 2) / Pass 2 (Unit 3) / merge (Unit 4).
- Copy references files byte-for-byte from any sibling skill (e.g., `skills/import/references/`).

**Patterns to follow:**
- `n4l-plugin/plugins/n4l/skills/import/SKILL.md` overall shape and step granularity.
- `--force` is a new flag with no existing precedent in the plugin. Specify its exact semantics in SKILL.md Step 2: `--force` = bypass marker-check and bypass merge; always overwrite. Idempotent when output does not exist.

**Test scenarios:**
- Happy path: user invokes `/n4l:md-import notes.md` → skill proceeds to Pass 1 with `notes.n4l` as output path.
- Happy path: user invokes `/n4l:md-import notes.md --output graph/notes.n4l` → skill uses override path.
- Edge case: `notes.md` missing → skill reports error, exits.
- Edge case: output `notes.n4l` exists with marker line → skill enters merge flow (Unit 4).
- Error path: output `notes.n4l` exists WITHOUT marker → skill aborts with error text naming the `--force` and `--output` escape hatches.
- Error path: `--force` + existing file → skill replaces without merge.
- Integration: SSTconfig/ absent from CWD and `$SSTORYTIME_HOME` → skill proceeds with placeholders + adds top-of-file NOTE (verified once Units 2+3 produce real output).

**Verification:**
- SKILL.md loads in Claude Code without frontmatter errors.
- References imports resolve (no `@./references/...` path errors).
- Marker-check branches correctly on fixture files with and without the marker.

- [x] **Unit 2: Pass 1 — structural parse + skeleton emission**

**Goal:** Implement the prompt steps that walk markdown structure and emit a valid `.n4l` skeleton with placeholder arrows and provenance comments on every edge.

**Requirements:** R5, R6, R7, R8, R9, R13, R14, R15

**Dependencies:** Unit 1.

**Files:**
- Modify: `n4l-plugin/plugins/n4l/skills/md-import/SKILL.md` (add `## Workflow - Pass 1` steps)
- Modify: `n4l-plugin/plugins/n4l/skills/md-import/references/` (optionally add `markdown-mapping.md` if feature-to-placeholder table gets too long for SKILL.md inline — evaluate during authoring)
- Test: `n4l-plugin/plugins/n4l/skills/md-import/tests/fixtures/simple-notes.md` + expected `simple-notes.n4l`
- Test: `n4l-plugin/plugins/n4l/skills/md-import/tests/fixtures/table-heavy.md` + expected output
- Test: `n4l-plugin/plugins/n4l/skills/md-import/tests/fixtures/prose-heavy.md` (for R6 trigger)
- Test: `n4l-plugin/plugins/n4l/skills/md-import/tests/fixtures/collision.md` (for R7 context-scoped identity)

**Approach:**
- One step per markdown feature in R5. For each, specify the placeholder arrow, how node naming strips inline markdown, how provenance is emitted, and examples.
- Frontmatter step: bare `:: tags ::` line at top; `title:` overrides chapter name.
- Heading step: emit `+:: <heading-text> ::` on open, `-:: <heading-text> ::` when next same-or-shallower heading appears or at EOF.
- List step: bullets → `(contain)` placeholder with ditto alignment; ordered lists → wrap in `+:: _sequence_ ::`/`-:: _sequence_ ::`.
- Table step: row-per-node, column-per-edge with `(note)` placeholder and `# SUGGEST: column <name>` comment.
- Link step: `[text](url)` → `text` as node, `(see)` placeholder, `url=<href>` trailing provenance annotation.
- Blockquote step: content-per-node, `(note)` placeholder (note: `quote` is not a standard SSTconfig code — use `note` as placeholder, suggest `(quote)` in comment; Pass 2 may specialize).
- Code fence step: fence content as quoted node, `(note)` placeholder with `# SUGGEST: language=<lang>` if fenced lang present.
- Node-name normalization step: strip `**bold**`, `_italic_`, `` `code` ``, `[text](url)` → text; trim whitespace; quote if value contains parens/commas.
- Context-scoped node identity: when emitting a bullet, Pass 1 tracks an in-flight collision map `(context_stack → {stripped_text: count})`. Second occurrence under same stack gets `(2)` suffix, third gets `(3)`, etc. Link-collision case: if same-context produces two bullets with identical stripped text from different link URLs, qualify as `Database [postgres.md]` form.
- Provenance step: every emitted edge line ends with ` # src:<rel-path>:L<line> h:<hash>`. Hash is `sha256(normalized-line)[:6]` where normalized-line is `"<node> (<placeholder>) <target>"` post-quoting, post-normalization.
- Prose-bytes counter: walk file once before structural parsing, count bytes per line type (prose / structural / frontmatter), compute ratio; if ≥ threshold (default 50%, provisional), emit `# NOTE:` hint line at top.
- Top-of-file marker: always emit `# n4l:md-import :: src=<rel-path>` as first line.

**Execution note:** Draft against `simple-notes.md` fixture first; use output diff to validate behavior before iterating.

**Technical design:** *(directional — not implementation specification)*
```
pass1(md_path) -> n4l_string:
  src_lines = read md_path
  prose_ratio = compute_prose_ratio(src_lines)   # R6 trigger

  emit marker line
  emit R6 NOTE if prose_ratio >= 0.50
  emit R12 NOTE if SSTconfig not located

  emit chapter line from frontmatter.title or filename-sans-ext
  emit bare :: frontmatter.tags ::

  for each structural block in src_lines:
    open/close heading contexts via +::/-::
    for each edge derivable from block:
      node = normalize(stripped_text_with_context_scope)
      ph   = placeholder_for_feature(block_type)   # contain/fwd/see/note
      tgt  = normalize(target_text)
      h    = sha256(f"{node} ({ph}) {tgt}")[:6]
      emit f"{node} ({ph}) {tgt}   # src:{rel}:L{lineno} h:{h}"
```

**Patterns to follow:**
- `n4l-plugin/plugins/n4l/skills/import/SKILL.md` Step 8 "Generate N4L Output" rules 1–9 (ditto alignment, quoting, blank-line-between-blocks).
- `examples/knowledge.n4l:37-49` for exact `+:: _sequence_ ::` nesting form.
- `examples/italian_cooking.n4l` (via `references/n4l-syntax.md:107-134`) for chapter+context+sequence complete shape.

**Test scenarios:**
- **Happy path — simple notes:** `simple-notes.md` with 1 H1, 1 H2, 3 bullets, 1 ordered list of 2 items → produces marker + chapter + one `:: ::` + 3 `(contain)` edges + `+:: _sequence_ ::` block with 2 items + closing `-:: ::`. Every edge has `# src:` provenance.
- **Happy path — frontmatter tags:** fixture with `title: Cookbook` and `tags: [food, italian]` → chapter line is `- Cookbook`, first non-comment line is `:: food, italian ::`.
- **Happy path — table:** `table-heavy.md` with 3-column, 4-row table under H2 "Recipes" → 4 nodes keyed by first column, 2 edges per node with `(note)` placeholder and `# SUGGEST:` comments naming columns.
- **Edge case — empty file:** 0-byte input → skill emits marker + empty chapter + exits cleanly (no edges).
- **Edge case — frontmatter-only:** input with only `---\ntitle: X\n---` and nothing else → marker + `- X` + nothing.
- **Edge case — nested bullets:** `- parent\n  - child` → `parent (contain) child` emitted as hierarchical CONTAINS.
- **Edge case — code fence without language:** ` ``` ` / body / ` ``` ` → node with quoted body and `(note)` placeholder, no language SUGGEST.
- **Edge case — unclosed code fence:** enters fence but EOF before close → skill emits a `# REVIEW: unclosed code fence starting at L<n>` and treats remaining lines as fence content.
- **Edge case — link with same text, different URLs, same context:** `[Database](postgres.md)` and `[Database](mysql.md)` under same heading → nodes named `Database [postgres.md]` and `Database [mysql.md]` (no collision).
- **Edge case — two `## Usage` headings under different H1s:** context stack differs; no identity collision.
- **Edge case — same-context bullet text collision:** two bullets with identical stripped text under the same heading → second becomes `text (2)`.
- **R6 trigger:** `prose-heavy.md` where paragraph bytes dominate → `# NOTE: this file contains substantial prose...` emitted.
- **R6 non-trigger (false-positive guard):** `table-heavy.md` where most bytes are table cells → no NOTE emitted (tables are structural, not prose).
- **R6 non-trigger (code-fence guard):** `code-only.md` with one huge fenced block → no NOTE emitted.
- **Integration — hash stability:** running Pass 1 twice against the same md produces byte-identical skeleton (modulo Pass-2 arrow specialization), i.e., hashes are deterministic.
- **Integration — N4L -v acceptance:** skeleton for every fixture passes `N4L -v` (requires Unit 6 infra to automate).

**Verification:**
- For each fixture, diffing actual vs expected `.n4l` yields zero structural differences.
- Every edge line has a `# src:` and `h:` stanza.
- Hash collisions on the fixture set are zero (verifies 6-hex-char truncation is safe for realistic file sizes).

- [x] **Unit 3: Pass 2 — Claude arrow specialization**

**Goal:** Given a Pass-1 skeleton and the active SSTconfig/ arrow files, walk the skeleton and replace placeholder arrows with specific codes where a confident pick exists. Leave placeholder + `# SUGGEST:` when no candidate is better; emit `# REVIEW:` when multiple candidates are plausible.

**Requirements:** R10, R11, R12, R13

**Dependencies:** Unit 2.

**Files:**
- Modify: `n4l-plugin/plugins/n4l/skills/md-import/SKILL.md` (add `## Workflow - Pass 2` steps)

**Approach:**
- Pass-2 step reads the skeleton emitted by Pass 1 and the four `SSTconfig/arrows-*.sst` files (CN-2, LT-1, NR-0, EP-3).
- Build a per-meta-type candidate list (both directional `+ name (code)` and symmetric `name (code)` forms — NEAR is symmetric-only; others are directional pairs).
- For each edge with a placeholder arrow, Claude evaluates: (a) heading breadcrumb (walk context stack), (b) source node text, (c) target node text, (d) candidate arrow labels for the matching meta-type. Pick one.
- If pick is confident AND a second candidate is plausibly equally good → emit `# REVIEW: or (<alt>)?`. If pick is confident and distinctly better → emit specialized arrow, no comment. If no candidate beats the placeholder → keep placeholder, emit `# SUGGEST: review against arrows-<code>.sst`.
- **Chunking:** compute skeleton token estimate; if > 40k tokens, chunk by top-level heading (each chunk = one chapter's contexts). Include heading breadcrumb prefix for chunks that aren't at file start. Each chunk independently receives the candidate catalog.
- Hash is NOT recomputed. Preserve the Pass-1 `h:<hash>` comment unchanged.

**Patterns to follow:**
- `n4l-plugin/plugins/n4l/skills/interpret/SKILL.md` for the pattern of reading SSTconfig/ to produce human-readable annotations. The arrow-file parsing logic is reusable.
- `n4l-plugin/plugins/n4l/skills/scaffold/SKILL.md` for context-aware arrow selection from SSTconfig/.

**Test scenarios:**
- **Happy path — clear pick:** skeleton edge `"sauté" (contain) "onions"` under `:: recipe ::` breadcrumb → Pass 2 emits `(ingred)`. Hash comment preserved.
- **Happy path — no better pick:** generic relationship under a vague heading → placeholder retained, `# SUGGEST:` comment added.
- **Happy path — REVIEW with alternative:** two candidates equally plausible → arrow picked, `# REVIEW: or (<alt>)?` trailing comment.
- **Edge case — empty skeleton:** zero edges → Pass 2 no-ops, skeleton passes through.
- **Edge case — no SSTconfig available:** Pass 2 detects R12 warning line at top and short-circuits, leaving placeholders intact.
- **Edge case — all arrows of a meta-type are symmetric (NEAR):** no forward/reverse ambiguity; picks work on label match alone.
- **Integration — determinism check:** re-running Pass 2 with temperature-0-equivalent settings on same skeleton yields high-agreement output (not byte-identical — LLM — but the same specialized code for most edges). Non-determinism is bounded by R13's hash-over-Pass-1 rule: spurious `edited` tags don't appear on re-runs even when Pass 2 flips a choice, because merge reads Pass-1 hashes.
- **Integration — chunking boundary:** a >40k-token skeleton spanning 5 top-level headings chunks into 5 prompts, each carrying the full arrow catalog; specializations at chunk boundaries stay consistent.

**Verification:**
- Every placeholder arrow in the skeleton is either: (a) specialized into a non-placeholder code, (b) kept with a `# SUGGEST:` comment, or (c) specialized with `# REVIEW:` flagging an alternative. No placeholders without one of these markers.
- Provenance `h:<hash>` unchanged on every line.
- On the Reference Corpus (Unit 6), REVIEW+SUGGEST density ≤ 1 per 10 edges (origin success criterion #2).

- [ ] **Unit 4: Two-way merge on re-run**

**Goal:** When the output `.n4l` already exists with the skill's marker, merge fresh Pass-1+Pass-2 output against the existing file, preserving hand-edits and flagging removed sources. Drop the old three-way framing and `.n4l.backup` fallback.

**Requirements:** R3, R16, R17 (revised to tuple-based), R18 (removed from v1 — three-way merge with `.n4l.backup` deferred to a follow-up plan once a concrete failure mode emerges)

**Dependencies:** Unit 2 (needs stable Pass-1 hash), Unit 3 (needs Pass-2 output for the "fresh" side).

**Files:**
- Modify: `n4l-plugin/plugins/n4l/skills/md-import/SKILL.md` (add `## Workflow - Merge` step after Unit 1's marker-check, before Pass 1 final-write)
- Test: `n4l-plugin/plugins/n4l/skills/md-import/tests/fixtures/merge/` (directory of md+n4l pairs)

**Approach:**
- Load the existing `.n4l` file. Parse each edge line: extract node, arrow, target, and provenance (`# src:... h:...`). Record the line's full text for hash-mismatch preservation.
- Run Pass 1 fresh on the current md to get the candidate skeleton, indexed by tuple (normalized-node, meta-type-of-placeholder, normalized-target). Compute fresh hashes.
- For each edge in the existing `.n4l`:
  - Compute tuple T = (normalized-node, meta-type, normalized-target) from the *stored Pass-1 form* — the meta-type is determined by the arrow's meta-type category in SSTconfig (CONTAINS/LEADSTO/NEAR/EXPRESS). For an `edited` edge whose arrow was user-modified, the meta-type is derived from whichever arrow-file the current arrow code belongs to. If the arrow code is absent from all four files (malformed edit), preserve the edge verbatim and tag `# user-orphan`.
  - Lookup tuple T in the fresh skeleton index.
  - If present AND `h:` matches fresh hash of same tuple → Pass-1 regen is equivalent; use fresh Pass-1+Pass-2 output (user has not edited).
  - If present AND `h:` mismatches → user edited the line; preserve verbatim, append ` edited` to provenance (idempotent if already present), skip Pass-2 for this edge.
  - If NOT present in fresh skeleton AND provenance exists → source no longer in markdown; emit old edge commented out with a leading `# REMOVED: source no longer in markdown` line.
  - If provenance is absent entirely → user-added edge; preserve verbatim, tag `# user-added` on first merge only.
- After processing existing edges, walk fresh Pass-1 skeleton for any tuple not seen in existing → new edge from new md content, emit normally via Pass-1+Pass-2.
- Preserve or update the top-of-file marker line; re-evaluate R6/R12 warnings against current file.

**Patterns to follow:**
- No existing skill does merge; this is a novel mechanism. Document the four-state logic inline in SKILL.md as a table (mirror the design doc table above).

**Test scenarios:**
- **Happy path — unchanged md, unchanged n4l:** re-run on a file where nothing changed → output byte-identical (or Pass-2-specialization-identical) to previous run. No `edited` or `REMOVED` tags.
- **Happy path — user edits one arrow:** existing `.n4l` has `Alice (contain) Acme` user-edited to `Alice (works-at) Acme` → re-run preserves the `(works-at)` line verbatim and adds ` edited` to provenance.
- **Happy path — md adds a new bullet:** fresh skeleton has one tuple not in existing → new edge inserted with fresh provenance.
- **Edge case — md renames heading:** `## Usage` → `## Usage Guide` under same H1 → tuple of child bullets changes (context stack changed), old tuples not found in fresh → commented out as REMOVED; new tuples emitted with fresh provenance.
- **Edge case — md reorders sections:** identical tuples in different positions → tuple match succeeds regardless of line number; no spurious REMOVED/added.
- **Edge case — user adds a non-md-derived edge:** bullet in existing `.n4l` with no `# src:` provenance → preserved verbatim with `# user-added` tag.
- **Edge case — user deletes an edge from existing n4l:** on re-run, fresh skeleton contains the tuple → edge re-emitted. Document this behavior (merge reintroduces deletions). User must re-delete or use `--force`. Add a success-criterion note.
- **Error path — existing n4l is malformed (missing marker):** Unit 1's marker-check catches this before merge runs.
- **Error path — hash-mismatch on a line with modified TARGET text (not arrow):** tuple has changed → tuple-lookup fails → treated as "source removed" rather than "edited". Document this as a known limitation — prose edits to the target change the tuple and look like deletion.
- **Integration — three-run sequence:** (1) first run, (2) hand-edit one arrow, (3) edit md to add a bullet, re-run → second run has `edited` tag on edit, third run has both `edited` tag preserved + new bullet inserted.

**Verification:**
- All four R17 states (match, mismatch, no-provenance, source-removed) are exercised by fixtures.
- Tuple-based detection correctly handles section reordering (no false REMOVED).
- Edited lines keep their original user text verbatim across re-runs (idempotent `edited` tagging).

- [ ] **Unit 5: Edge-case documentation and test fixtures**

**Goal:** Codify R5 edge-case behavior (empty files, unclosed fences, frontmatter errors, setext headings, etc.) in SKILL.md with exact fixtures. Close the review P1 finding on R5 edge cases.

**Requirements:** R5 (edge cases), R6 (threshold tuning), R7 (collision policy)

**Dependencies:** Unit 2 (fixtures align with Pass 1 logic).

**Files:**
- Modify: `n4l-plugin/plugins/n4l/skills/md-import/SKILL.md` (add `## Special Handling Notes` section matching `n4l:import/SKILL.md:209+`)
- Create: `n4l-plugin/plugins/n4l/skills/md-import/tests/fixtures/edge-cases/` (empty, frontmatter-only, unclosed-fence, setext-heading, multi-h1, mixed-language, nested-lists-deep, html-embedded)
- Modify: tests/fixtures/collision.md with three collision types documented

**Approach:**
- For each edge case, write the fixture + expected output + a line in SKILL.md's special-handling section explaining the rule.
- Multi-H1 decision: each H1 opens its own context (no chapter split); chapter remains filename/frontmatter.title. Document this explicitly.
- Setext headings (`===` / `---` underline syntax): treat identically to their ATX equivalents (H1/H2).
- HTML embedded: strip tags, treat content as prose (contributes to R6 ratio).
- Mixed-language node names: preserved as-is (UTF-8); aligns with existing `c.s.peirce_semiotics.n4l` pattern.
- Nested lists deeper than 4 levels: chain CONTAINS without ditto-alignment artifacts; document that deep nesting is valid but `(contain)` is the only arrow used.

**Patterns to follow:**
- `n4l-plugin/plugins/n4l/skills/import/SKILL.md` lines 209–217 "Special Handling Notes" format.

**Test scenarios:**
- Happy path — setext H1/H2: `Title\n=====\n\nSection\n-------\n- bullet` → same output as `# Title\n## Section\n- bullet`.
- Happy path — multi-H1: file with two H1s → both open contexts under the single chapter, stack closes one H1's scope when the next opens.
- Edge case — HTML tags in bullet: `- <b>Alan Turing</b> invented the machine` → stripped node text `Alan Turing invented the machine`.
- Edge case — mixed-language: `- 知识 (knowledge)` → single node `知识 (knowledge)` (quote-wrapped because parens).
- Edge case — YAML frontmatter parse error: skill falls back to filename as chapter and emits `# NOTE: frontmatter parse failed at L<n>` hint.
- Edge case — deeply nested ul (5 levels): each level chains CONTAINS to its parent bullet.
- R6 threshold tuning: run against `prose-heavy.md`, `mixed.md`, `table-heavy.md`, `code-only.md` fixtures; adjust threshold (default 50%) if miscategorization occurs. Document final value in Key Decisions table.

**Verification:**
- Every fixture in `tests/fixtures/edge-cases/` has an expected output; diff is zero.
- SKILL.md's Special Handling Notes names every edge case covered by fixtures.
- R6 threshold is empirically validated: no false positive on table-heavy/code-heavy, no false negative on prose-heavy.

- [ ] **Unit 6: Reference Corpus + end-to-end acceptance**

**Goal:** Pin the Reference Corpus files, run end-to-end acceptance (Pass 1 + Pass 2 + `N4L -v`) against them, and confirm origin success criteria. Document the final REVIEW+SUGGEST density achieved.

**Requirements:** All origin success criteria

**Dependencies:** Units 1–5.

**Files:**
- Create: `n4l-plugin/plugins/n4l/skills/md-import/tests/corpus/README.md` (names the corpus files + rationale for each)
- Create: symlinks or copies of 5–8 corpus files from `docs/brainstorms/`, `docs/ideation/`, `n4l_patterns/` at `tests/corpus/`
- Test: `n4l-plugin/plugins/n4l/skills/md-import/tests/corpus-acceptance.md` (documenting expected REVIEW+SUGGEST density per file)

**Approach:**
- Pick files with variety: one heading-dense (ideation doc), one table-heavy, one mixed prose+structure, one frontmatter-rich, one code-fence-heavy.
- For each, run the skill, pipe output to `N4L -v`, record density metrics.
- Confirm origin success criteria:
  1. Skeleton + specialization passes `N4L -v` without syntax errors.
  2. REVIEW+SUGGEST density ≤ 1 per 10 edges on corpus.
  3. Re-run after hand-editing the .n4l preserves the edit (marked `edited`).
  4. Pointing at a prose-heavy file emits the R6 hint.

**Test scenarios:**
- Happy path — each corpus file: skill produces output, `N4L -v` exits 0.
- Happy path — density metric: aggregate REVIEW+SUGGEST tags per emitted edge across corpus ≤ 0.10.
- Happy path — re-run preservation: for each corpus file, first run + hand-edit one arrow + second run → edit preserved with `edited` tag.
- Happy path — R6 trigger: include at least one prose-heavy file in corpus to exercise the hint path.
- **Test expectation: none -- acceptance-level unit; covered by scenarios above rather than standalone test entries.**

**Verification:**
- Corpus README names every file and why it was chosen.
- Aggregated density metrics documented in `corpus-acceptance.md` and meet origin success criterion #2.
- All four origin success criteria verifiably met.

## System-Wide Impact

- **Interaction graph:** The skill is additive — no existing skill changes. Downstream consumer: `n4l:upload` validates the output. Future consumer (ideation 6a.4): parser-error back-propagator will read the `# src:...` provenance comments. The provenance format is therefore a **cross-skill interface** and must not change without a migration plan.
- **Error propagation:** The skill's failure modes are local (bad md file, missing SSTconfig, merge-state confusion). No upstream errors to propagate. Downstream `N4L -v` errors appear only when the user pipes the output further.
- **State lifecycle risks:** The only persistent state is the `.n4l` file itself. The marker line (R14) prevents confusion with hand-written or CSV-derived `.n4l` files. No temp files, no cache, no hidden state.
- **API surface parity:** No other skills need parity changes. The origin brainstorm explicitly defers pipe mode (6a.6) and vault import (6a.3).
- **Integration coverage:** Unit-level fixtures cover single-file behaviors. Corpus acceptance in Unit 6 covers `N4L -v` integration. Merge integration tested in Unit 4. Real `n4l:upload` pipe is deferred but the contract (valid `.n4l`) is tested.
- **Unchanged invariants:**
  - Existing skills (`n4l:import`, `n4l:scaffold`, `n4l:upload`, `n4l:interpret`, etc.) are not modified.
  - Shared `references/n4l-syntax.md` and `arrow-types.md` are copied, not modified.
  - `SSTconfig/arrows-*.sst` is read-only.
  - The `N4L` binary, `N4L.go` parser, and N4L file format are unchanged.
  - The `+:: _sequence_ ::` nesting pattern under `:: ::` contexts is an existing capability (evidenced by `examples/knowledge.n4l`) — the new skill exercises it but does not alter it.

## Risks & Dependencies

| Risk | Mitigation |
|---|---|
| Pass-2 non-determinism causes spurious `edited` tags on re-run | R13 pins hash to Pass-1 output only (pre-specialization); Unit 4 tests confirm re-runs are stable. |
| Hash collision at 6 hex chars (~16M space) | Unit 2 fixture validates zero collisions at realistic file sizes. If collision observed, widen to 8 chars (still fits in ~12-char provenance suffix). |
| User edits target text, not arrow — merge sees as deletion, not edit | Documented as known limitation (Unit 4 test). Target-text edits are rare in practice; if common, future iteration can hash target separately. |
| User deletes an edge but re-run reintroduces it | Documented behavior — user must re-delete or hand-edit markdown to eliminate the source. Aligns with "markdown is canonical" framing. |
| Pass-2 context window exceeded on very large files | Unit 3 chunking threshold (~40k tokens) with per-chapter fallback. No plan-time evidence of files this large in the target corpus; revisit if Unit 6 surfaces a counter-example. |
| `SSTconfig/` forked with one of the four placeholder codes removed | Unlikely (all four are canonical SST arrows, verified in shipped SSTconfig). If it happens, R12 fallback already covers the missing-arrow case with a NOTE; output still parses because N4L accepts any arrow token. |
| Provenance comment format becomes load-bearing for 6a.4 and changes later break that skill | Lock the format in Unit 2 with an explicit comment in SKILL.md; treat as stable interface. Any change requires coordinated update. |
| Content-hash provenance overlaps with git's change tracking (review P2 finding) | Accepted tradeoff: git tracks file-level history; our hashes track per-edge skill-derived-vs-hand-edited state. Both coexist; no conflict. |

## Documentation / Operational Notes

- The SKILL.md itself serves as user-facing documentation. Update `n4l-plugin/plugins/n4l/AGENTS.md` to list the new skill.
- Worked example in SKILL.md should use a small sample markdown file to demonstrate the structure-first / arrow-second output shape.
- No deployment, no rollback, no monitoring — the skill runs in the Claude Code CLI and produces local files only.
- No migration — additive new skill.

## Sources & References

- **Origin document:** [docs/brainstorms/2026-04-18-n4l-md-import-requirements.md](../brainstorms/2026-04-18-n4l-md-import-requirements.md)
- **Ideation lineage:** [docs/ideation/2026-04-17-n4l-claude-skills-ideation.md](../ideation/2026-04-17-n4l-claude-skills-ideation.md) (idea 6a.1)
- **Prior plan for sibling skills:** [docs/plans/2026-04-17-001-feat-n4l-skills-plugin-plan.md](./2026-04-17-001-feat-n4l-skills-plugin-plan.md)
- **SKILL.md template:** `n4l-plugin/plugins/n4l/skills/import/SKILL.md`
- **Plugin conventions:** `n4l-plugin/plugins/n4l/AGENTS.md`
- **Shared references (to be copied):** `n4l-plugin/plugins/n4l/skills/import/references/n4l-syntax.md`, `arrow-types.md`
- **N4L syntax canonical example:** `n4l-plugin/plugins/n4l/skills/*/references/n4l-syntax.md:107-134`
- **Nested sequence exemplar:** `examples/knowledge.n4l:37-49`
- **Parser comment handling:** `src/N4L/N4L.go:396, 1327, 1479`
- **SSTconfig arrow files:** `SSTconfig/arrows-{NR-0,LT-1,CN-2,EP-3}.sst`
- **Related future skill:** ideation 6a.4 (parser-error back-propagator) will consume the provenance format specified in R9

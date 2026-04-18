---
date: 2026-04-17
topic: n4l-claude-skills
focus: Claude skills to assist in using SSTorytime
---

# Ideation: Claude Skills for SSTorytime

## Codebase Context

SSTorytime is a Go-based knowledge graph database on PostgreSQL using Semantic Spacetime (SST) principles. Users write knowledge in N4L (Notes for Learning) plain text format, upload to PostgreSQL via the N4L tool, and query via searchN4L, pathsolve, notes, and an http_server web UI.

Key pain points:
- N4L syntax has many features (chapters, contexts, arrows, sequences, references) -- easy to forget or misuse
- Arrow selection requires understanding 4 meta-types (NEAR, LEADSTO, CONTAINS, EXPRESS) and dozens of specific arrows
- Context management is called "the hard problem" -- designing good context schemas is difficult
- searchN4L query syntax uses backslash commands that must be composed correctly
- No CLAUDE.md or AI assistant guidance exists
- No validation of N4L files before upload

## Ranked Ideas

### 1. Arrow Advisor (with Socratic teaching)
**Description:** User describes a relationship in natural language and the skill recommends the correct SST arrow type and name. Includes interactive teaching through concrete examples.
**Rationale:** Arrow selection is the steepest learning curve and most consequential design decision. Compounds with every authoring session.
**Downsides:** Must keep in sync with SSTconfig/ files. May oversimplify nuanced cases.
**Confidence:** 90%
**Complexity:** Low
**Status:** Unexplored

### 2. N4L Scaffolder / Domain-to-N4L
**Description:** User describes a knowledge domain and the skill generates a complete starter N4L file with chapters, contexts, appropriate arrows, and example entries.
**Rationale:** Eliminates the blank-page problem. 40+ example .n4l files provide rich patterns. Context design gets addressed as part of scaffolding.
**Downsides:** Generated scaffolds may not match user's mental model.
**Confidence:** 85%
**Complexity:** Medium
**Status:** Explored (brainstormed 2026-04-17 as n4l-scaffold)

### 3. Search Query Builder
**Description:** Translates natural language search intent into correct searchN4L/pathsolve command syntax with proper escaping and execution.
**Rationale:** searchN4L has 15+ backslash commands with subtle distinctions and shell escaping gotchas.
**Downsides:** Cannot verify queries work without DB access.
**Confidence:** 85%
**Complexity:** Low
**Status:** Explored (brainstormed 2026-04-17 as n4l-search)

### 4. N4L Refactoring / Tidy Daily (with lint mode)
**Description:** Analyzes existing N4L files for structural improvements: placeholder arrows, capitalization collisions, orphan nodes, oversized chapters. Includes lint mode for pre-upload validation.
**Rationale:** The docs encourage writing (tbd) first and refining later, and the five-step program calls for daily tidying -- but no tool supports this.
**Downsides:** Re-parsing N4L risks diverging from Go parser behavior.
**Confidence:** 80%
**Complexity:** Medium
**Status:** Unexplored

### 5. Arrow Extension Wizard
**Description:** Helps users design new arrows for SSTconfig/: determines STtype category, checks short-code collisions across all 4 config files, generates correct forward/inverse pair syntax.
**Rationale:** Arrow definitions are the vocabulary of SSTorytime. Getting the format wrong silently breaks the compiler.
**Downsides:** Niche use case -- only needed when extending arrow vocabulary.
**Confidence:** 75%
**Complexity:** Low
**Status:** Unexplored

### 6. Import Pipeline (CSV/Markdown/Obsidian -> N4L)
**Description:** Converts structured knowledge from other formats into well-formed N4L files with appropriate arrow types.
**Rationale:** Cold-start bottleneck -- people have existing knowledge in other tools.
**Downsides:** Each source format is its own converter. Scope creep risk.
**Confidence:** 70%
**Complexity:** Medium-High
**Status:** Explored (brainstormed 2026-04-17 as n4l-import, CSV shipped v1; markdown/Obsidian ideation expanded 2026-04-18 below)

#### 6a. Markdown -> .n4l Conversion Skill — Expanded Ideation (2026-04-18)

Sub-ideation for the markdown branch of idea #6. Ranked survivors from 32 candidates across 4 frames (user pain, inversion/automation, leverage/compounding, edge cases) -> 25 deduped + 4 cross-cutting syntheses -> 7 survivors.

**Grounding hooks:** No md->n4l skill exists. `n4l:import` CSV pattern is the template (shared `@./references/arrow-types.md`, `@./references/n4l-syntax.md`, `SSTconfig/` lookup order CWD -> `$SSTORYTIME_HOME` -> fallback, "safe default + suggest in comments"). `n4l_patterns/` directory already exists (untracked) — natural home for harvested rules. `text2N4L` is explicitly out of scope (statistical sentence fractionation, not structural parsing).

**Natural mappings anchor:** H1 -> `- chapter`; H2/H3 -> sub-chapter or `:: context ::`; bulleted list -> CONTAINS `(contain)` with ditto alignment; numbered list -> LEADSTO `+:: _sequence_ ::` block; `[text](url)` -> `(see)` NEAR default; emphasis/callouts -> EXPRESS `(note)`, `(e.g.)`; frontmatter `tags:` -> context tags.

##### 6a.1 Structure-first / arrow-second converter core *(synthesized: zero-question + deferred-arrow + provenance)*
**Description:** Two-pass architecture. Pass 1 parses md structure (headings, lists, tables, links, emphasis, frontmatter) and emits valid n4l with *generic* arrow placeholders (`fwd`, `contain`, `note`, `see`) plus stable `# src:file.md:L42` provenance comments on every edge. Pass 2 walks the skeleton and specializes each placeholder by consulting `SSTconfig/arrows-*.sst` with node-text context. Never blocks on questions; ambiguous calls become `# REVIEW:` comments rather than prompts.
**Rationale:** Decouples the boring 90% (structure) from the interesting 10% (arrow precision). First pass is deterministic and always emits valid n4l. Second pass is where SSTconfig expertise lives and can be re-run independently as the `SSTconfig/` vocabulary evolves. Provenance comments underpin error back-propagation (6a.4) and round-trip edits.
**Downsides:** Two parser runs per file. Placeholders need a review pass before upload; tempting to ship un-specialized n4l.
**Confidence:** 90% **Complexity:** Medium **Status:** Explored (brainstorming 2026-04-18)

##### 6a.2 Dry-run diff-first review mode
**Description:** Default first run is non-destructive: emits a side-by-side "left = md line, right = proposed n4l line with arrow-code annotated" review. Re-runs against edited markdown produce a three-way diff (old md / new md / current .n4l) so drift is visible before commit. A `--yes` flag skips review for batch/CI use.
**Rationale:** Builds trust before committing an author's whole notes archive. Makes repeated runs safe — silent drift is the top risk of naive converters. Borrows the review-before-commit pattern from `n4l:upload`'s validate-fix-upload loop and applies it one step upstream.
**Downsides:** Adds a confirmation step that batch users will want to skip.
**Confidence:** 85% **Complexity:** Low **Status:** Unexplored

##### 6a.3 Obsidian / vault bulk importer
**Description:** `--vault <path>` mode walks a whole Obsidian/Logseq vault. Folder hierarchy -> nested contexts. Frontmatter `tags:` -> `:: ctx ::`. `[[wikilinks]]` -> cross-file node references with canonical-name resolution. `aliases:` -> NEAR `(same)` arrows. Dangling-ref report emitted as a sidecar. One multi-chapter `.n4l` output (or chapter-per-file set) with shared context stack.
**Rationale:** A single file is a toy; a vault is a graph. This is the killer use case — a 5-year Obsidian archive becomes a live SSTorytime graph in one command. Cross-file node identity is what makes the graph worth having. Real SSTorytime corpora like `c.s.peirce_semiotics.n4l` and `jungian_psychology.n4l` are exactly the kind of knowledge that starts life in markdown notes.
**Downsides:** Canonical-name resolution is genuinely hard (alias collisions, case variants). Very large vaults may need streaming.
**Confidence:** 85% **Complexity:** Medium-High **Status:** Unexplored

##### 6a.4 Parser-error back-propagator
**Description:** After emitting n4l, runs `N4L -v` and when a cryptic error appears ("arrow redefinition at line 412"), maps the n4l line back to the originating markdown line via the provenance comments from 6a.1 and rewrites the error as: *"In notes.md line 87, the bullet under 'Jung' conflicts with the arrow used earlier for 'Freud' on line 34. Choose: [A] rename, [B] use (also), [C] open a new context scope."*
**Rationale:** Cryptic parser errors after a long import are the single most discouraging moment in SSTorytime authoring (surfaced in the n4l:upload FAQ). md-source provenance + repair menu turns dead-ends into recoverable edits. Extends `n4l:upload`'s validate-fix loop upstream into the source format.
**Downsides:** Requires N4L binary available and strict line-mapping discipline. Some errors (semantic, not syntactic) resist auto-repair.
**Confidence:** 75% **Complexity:** Medium **Status:** Unexplored

##### 6a.5 Learning loop: author profile + correction harvester
**Description:** When the user hand-edits generated n4l, the skill diffs pre/post and extracts rules like *"H3 under H1='Concept' became `(example_of)`"* into `n4l_patterns/learned-rules.yaml`. Subsequent runs consult harvested rules before falling back to defaults. Profile is per-author and shareable across a team to converge on consistent arrow vocabulary.
**Rationale:** Compounds across sessions — file 100 converts almost identically to file 1 without re-prompting. Turns every correction into durable capital rather than one-off effort. The `n4l_patterns/` directory already exists in the repo (untracked per git status) — natural home.
**Downsides:** Rule extraction heuristics can misgeneralize; needs a "forget this rule" escape hatch. Profile format needs to stay stable across skill versions.
**Confidence:** 70% **Complexity:** Medium **Status:** Unexplored

##### 6a.6 Pipeline-native composition (`--pipe`)
**Description:** A flag streams the converted `.n4l` directly to `n4l:upload` and optionally `n4l:interpret`, collapsing a 4-tool workflow (author -> convert -> validate -> upload -> interpret) into one command. Honors the plugin's chaining convention (skills compose via Claude, not via direct calls).
**Rationale:** This skill becomes infrastructure that multiplies the utility of every other n4l skill. Low-cost to ship and exercises the full pipeline, which is how users discover the downstream skills in the first place.
**Downsides:** Failure-mode complexity (which step failed?); needs careful error surfacing so debugging doesn't force manual decomposition.
**Confidence:** 85% **Complexity:** Low **Status:** Unexplored

##### 6a.7 Structural-pattern signature recipes
**Description:** Ships with named recipes — `glossary-list`, `how-to-steps`, `meeting-notes`, `academic-paper`, `zettel-atomic` — each a pre-tuned arrow palette + context strategy + heading mapping. Author picks per file (CLI flag) or per subtree (`<!-- n4l:signature how-to-steps -->`). Absorbs table-row-explosion and citation-reifier patterns as first-class recipes. New signatures learnable from a hand-authored example pair.
**Rationale:** Eliminates the cold-start choice for typical note styles. A 100-file Zettelkasten converts consistently without per-file tuning. Provides the onramp that a raw converter lacks. Parallels `n4l:scaffold`'s domain-description input at the conversion-time end.
**Downsides:** Curating recipes is ongoing work; wrong recipe on wrong file produces subtle mismatches. Signature discovery UX (how does the author know which to pick?) needs thought.
**Confidence:** 80% **Complexity:** Low-Medium **Status:** Unexplored

### 7. N4L Merge & Diff
**Description:** Semantic-level diffing and merging of N4L files, understanding that different line orderings may express the same graph.
**Rationale:** Text-level diff tools mangle N4L because of ditto marks and context inheritance. Enables collaborative knowledge graph building.
**Downsides:** Building a semantic diff engine for N4L is genuinely complex.
**Confidence:** 65%
**Complexity:** High
**Status:** Unexplored

### 8. n4l:interpret — Search Result Interpreter
**Description:** Takes searchN4L or pathsolve output and annotates it by looking up arrow short codes in SSTconfig/. When output shows `(ph)`, the skill reads the arrow definition files and annotates: `(ph) = "pinyin has hanzi" [EXPRESS type]`. Also explains orbit structure, centrality scores, and context tags in plain English.
**Rationale:** searchN4L output is dense with short codes that are meaningless without memorizing arrow definitions. Natural complement to n4l:search: one translates the query, the other translates the response. Reads the same SSTconfig/ files already used by scaffold and import.
**Downsides:** Requires parsing searchN4L's specific output format (indented orbits, arrow annotations). Output format may vary.
**Confidence:** 85%
**Complexity:** Low-Medium
**Status:** Explored (brainstormed and built 2026-04-17 as n4l-interpret)

### 9. n4l:upload — Upload with Error Auto-Fix
**Description:** Runs `N4L -v <file>` first to validate, then `N4L -u <file>` to upload. If errors occur, captures error output, reads the source N4L file, identifies and fixes problems (syntax errors, undefined arrows, capitalization collisions, arrow redefinition warnings), then retries up to 3 times.
**Rationale:** Currently errors are discovered only after running N4L -u and reading cryptic output. The FAQ shows common errors like "Warning: Redefinition of arrow" that require understanding N4L internals. Closes the validate-fix-upload loop into one command. Absorbs the "lint mode" from idea #4 (Tidy Daily).
**Downsides:** Requires N4L binary on PATH and running PostgreSQL. Some errors (arrow redefinition) are semantic, not syntactic — harder to auto-fix. Fixing may change user intent.
**Confidence:** 80%
**Complexity:** Medium
**Status:** Explored (brainstormed and built 2026-04-17 as n4l-upload)

## Rejection Summary

### Top-level skills ideation (2026-04-17)

| # | Idea | Reason Rejected |
|---|------|-----------------|
| 1 | Database Setup Automation | Wrong tool -- Makefile/Docker problem, not a Claude skill |
| 2 | N4L Explain & Annotate | Too obvious -- any LLM explains text in context |
| 3 | Go API Scaffolder | Too generic -- Claude generates Go code with docs in context |
| 4 | Five-Step Learning Coach | 2 of 5 steps need DB access the skill doesn't have |
| 5 | MCP-SST Bridge | Infrastructure project, not a prompt-based skill |
| 6 | CLAUDE.md | One-time documentation task, not a reusable skill |
| 7 | Talk-to-Graph | Same as N4L Scaffolder with different branding |
| 8 | Living Graph / Emergent Schema | Requires persistent state + DB writes |
| 9 | Knowledge Testing | Needs DB access to verify assertions |
| 10 | Socratic SST Tutor | Absorbed into Arrow Advisor |
| 11 | Context Schema Designer | Absorbed into N4L Scaffolder |
| 12 | Ghost Nodes / Gap Detection | Interesting analysis requires DB |
| 13 | Graph Report Interpreter | Paste-output doesn't need a dedicated skill |
| 14 | Pathsolve Optimizer | Absorbed into Search Query Builder |
| 15 | Polyglot Bridge | N4L already handles UTF-8 |
| 16 | Graph Health Dashboard | Needs DB for structural monitoring |
| 17 | Cross-Graph Inference | Would require reimplementing graph engine |
| 18 | Closure Composer | Audience of approximately one person |
| 19 | MCP Session Coach | Documentation, not a skill |

### Markdown -> .n4l sub-ideation (2026-04-18)

| # | Idea | Reason Rejected |
|---|------|-----------------|
| 1 | Zero-question silent conversion + structure-only inference | Design principle of 6a.1 core, not a distinct idea — absorbed |
| 2 | Paste-a-paragraph arrow coach / live-preview authoring template | Belongs as a capability of the main converter; overlaps with `n4l:import` "safe default + suggest" pattern |
| 3 | Capitalization-collision linter / auto-normalizer | Feature of main converter's node-canonicalization, not a distinct product move |
| 4 | Ambient-context sidecar (.n4l.context.md) | Introduces a 3rd file format; HTML comments in the md achieve the same goal more simply |
| 5 | Link-resolution strategy picker (5-option interactive) | Default `(see)` with url in comment is sufficient; power users use palette overrides |
| 6 | Round-trippable md<->n4l (full bijection) | Full reverse is genuinely lossy; forward-with-provenance (6a.1) captures ~90% of the value |
| 7 | Watch-mode daemon | Service, not a prompt-based skill; doesn't fit the skill contract |
| 8 | In-place markdown enrichment (edit author's source) | Mutating the author's source is a trust hazard |
| 9 | Conversion report as separate artifact | Absorbed as top-of-file comment block in generated n4l; separate file rots out of sync |
| 10 | Seed-graph foundry for public corpora | Program-level initiative, not skill-level; absorbed into 6a.7 signature recipes |
| 11 | Arrow vocabulary coverage telemetry | Requires persistent collection infrastructure; not a skill deliverable |
| 12 | Book-length streaming chunker with symbol table | Niche; H1-splitting handles most cases; cross-shard reconciliation is future work |
| 13 | Structured-table row explosion (standalone) | Absorbed as a pattern inside 6a.7 signature recipes |
| 14 | Mermaid diagram direct translator | Valuable but narrow; merits a dedicated mermaid->n4l skill later, not v1 of md->n4l |
| 15 | Citation / footnote reifier (standalone) | Absorbed as the `academic-paper` signature in 6a.7 |
| 16 | Per-section HTML-comment arrow palette overrides | Power-user escape hatch; defer to v2 once real user friction emerges |
| 17 | Multilingual node alias binder | Specific case of entity aliasing; handle inside main converter's canonicalization logic |
| 18 | Dirty-markdown lint-and-heal pass | Generic md-tooling concern; rely on external linters |

## Session Log

- 2026-04-17: Initial ideation -- ~48 candidates generated across 6 frames, deduped to 27, 7 survived adversarial filtering
- 2026-04-17: Brainstormed ideas #2, #3, #6 as n4l-scaffold, n4l-search, n4l-import (requirements doc created)
- 2026-04-17: Added ideas #8 (n4l:interpret) and #9 (n4l:upload) from user request — both survived assessment
- 2026-04-17: Brainstormed and built #8 and #9 as n4l-interpret and n4l-upload skills
- 2026-04-18: Expanded idea #6 with markdown->.n4l sub-ideation (6a) — 32 candidates across 4 frames (pain/friction, inversion/automation, leverage/compounding, edge cases) -> 25 deduped + 4 syntheses -> 7 survivors (6a.1-6a.7). Core move: two-pass structure-first / arrow-second architecture with md-source provenance comments
- 2026-04-18: Handed 6a.1 (structure-first/arrow-second core) to /compound-engineering:ce-brainstorm

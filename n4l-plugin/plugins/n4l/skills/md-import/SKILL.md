---
name: n4l:md-import
description: "Convert a Markdown file into N4L knowledge graph format using two-pass parsing. Pass 1 walks markdown structure (headings → contexts, lists → CONTAINS/LEADSTO edges, tables → row-per-node, links/blockquotes/code-fences → EXPRESS edges) and emits a skeleton with placeholder arrows and provenance comments. Pass 2 reads the active SSTconfig/arrows-*.sst and specializes each placeholder into a specific arrow code, flagging ambiguous picks with # REVIEW. Use when the user says 'import markdown', 'md to n4l', 'convert markdown', or 'markdown to knowledge graph'."
argument-hint: "<markdown-file-path> [--output filename.n4l] [--force]"
---

# N4L Markdown Import Skill

Convert a single Markdown file into a valid N4L file by walking the document's structure and specializing arrow choices against the active `SSTconfig/`. Two-pass architecture:

- **Pass 1 (structural):** walk markdown, emit skeleton with generic placeholder arrows (`contain`, `fwd`, `see`, `note`) and `# src:<path>:L<n>` provenance comments on every edge.
- **Pass 2 (semantic):** read the skeleton + `SSTconfig/arrows-*.sst` + heading breadcrumb per edge, specialize each placeholder into a specific SSTconfig arrow code. Ambiguous calls surface as `# REVIEW:` or `# SUGGEST:` trailing comments — never as prompts.

If `SSTconfig/` cannot be located, Pass 2 is skipped and placeholders are kept (the four placeholders are real SSTconfig codes, so output still parses). Merge-on-rerun is deferred to a future iteration; v1 aborts on existing output unless `--force` is used.

## Reference Material

@./references/n4l-syntax.md
@./references/arrow-types.md

## What this skill does (and doesn't)

**Does:**
- Parses markdown structure: YAML frontmatter, H1–H6 headings, unordered/ordered lists, tables, links, blockquotes, code fences.
- Emits a valid `.n4l` file with arrow codes drawn from the active `SSTconfig/` (Pass 2 specialization) or generic placeholders when `SSTconfig/` is unavailable.
- Tags every edge with `# src:<path>:L<line>` provenance comments.
- Flags ambiguous arrow picks with `# REVIEW: or (<alt>)?` trailing comments for human review.
- Warns when a file is mostly prose (suggest running `text2N4L` instead).
- Aborts on existing output unless `--force` is used (prevents clobbering hand-edited files).

**Does not (deferred to future iterations):**
- Detect hand-edits in existing `.n4l` and preserve them on re-run (merge).
- Import from Obsidian vaults with wikilink resolution (multi-file).
- Fractionate prose into sentences (use `text2N4L` for that).
- Interpret wikilinks `[[foo]]` or Obsidian tags `#foo` as structure (treated as literal text).

**Known Pass 2 limitation:** Claude's arrow specialization is non-deterministic. Re-running the skill on the same markdown may produce slightly different specialized arrows for the same edge. Provenance comments and hashes are computed from Pass-1 (placeholder) form, so re-runs do not falsely flag edges as "edited" when merge ships. But the diff of two `.n4l` outputs from the same markdown will show arrow variance. Review `# REVIEW:` tags before upload.

## Workflow

### Step 1: Parse Arguments

Extract:

- **markdown-file-path** (required): Path to the `.md` file. Single file only.
- **--output filename.n4l** (optional): Override output path. Default: swap `.md` → `.n4l` in the same directory.
- **--force** (optional): Overwrite any existing output file without checking for marker or merge. Default: refuse to overwrite unless the file is this skill's own previous output (which requires the merge feature — deferred — so v1 always aborts on non-`--force` existing files).

If the markdown file path is missing, ask: "Which markdown file would you like to import?"

### Step 2: Existing-File Check

Check whether the output path already exists:

1. Use the Read tool on the output path. If it doesn't exist, proceed to Step 3.
2. If it exists, read its first non-blank line.
3. If `--force` was provided, proceed to Step 3 (overwrite).
4. If the first non-blank line is `# n4l:md-import :: src=<anything>`, the file is a previous output of this skill. Merge-on-rerun is deferred to v2 — for v1, abort with:
   > `Existing output notes.n4l was produced by this skill. Merge-on-rerun is deferred to v2. Pass --force to overwrite, or use --output to pick a different path.`
5. If the existing file does NOT start with the marker, it's a foreign file (hand-written N4L, or output from `/n4l:import` on a sibling CSV). Abort with:
   > `Existing file notes.n4l was not produced by this skill. Pass --force to overwrite, or use --output <different-path>.`

### Step 3: Read the Markdown File

Use the Read tool on the markdown file path. If missing or unreadable, report the error and stop.

Record the absolute path and a relative-to-CWD path — the relative path is used in provenance comments.

### Step 4: Classify Line Types (for prose-ratio check)

Walk the file once line-by-line and classify each non-blank line into one of these buckets:

| Bucket | What counts |
|---|---|
| Frontmatter | Lines inside the leading `---` … `---` YAML block (including the fences). |
| Heading | Lines matching `#{1,6} ` (ATX) or underline form `====` / `----` on the line after a text line (setext). |
| List item | Lines starting with (optional whitespace) `-`, `*`, `+`, or `N.` / `N)`. |
| Table | Lines containing `|` with at least 2 `|` separators. |
| Blockquote | Lines starting with (optional whitespace) `>`. |
| Code fence | Between matching ```` ``` ```` or `~~~` markers (including the fences). |
| Prose | Any other non-blank line. |

Compute `prose_ratio = prose_bytes / (prose_bytes + structural_bytes)` where structural_bytes = sum of bytes on lines in any non-prose bucket. Whitespace-only lines are excluded from both sides.

If `prose_ratio >= 0.50`, remember to emit a `# NOTE:` at the top of the output (Step 7).

### Step 5: Walk the Structure and Emit Edges

Produce an in-memory list of `(context_stack, source_node, placeholder_arrow, target_node, md_line_number, extra_annotation)` tuples. Rules per feature:

**YAML frontmatter**
- `title: X` → use `X` as the chapter name (overrides filename default).
- `tags: [a, b]` → produce a bare `:: a, b ::` context line at the top of the output (sets the base context). Heading contexts add onto this with `+::` / `-::`.
- Ignore other frontmatter keys in v1.

**Headings (H1–H6, ATX and setext)**
- On opening a heading, push its stripped text onto the context stack and emit `+:: <stripped_text> ::`.
- On encountering the next heading at the same or shallower depth, pop all deeper contexts with `-:: <stripped_text> ::` in LIFO order, then open the new one.
- At EOF, pop all remaining open contexts.
- A heading itself does not create a node — it only scopes subsequent edges.

**Unordered lists (bulleted)**
- Each list item becomes a node named by the bullet's stripped text (see "Node naming" below).
- Source node = the most recent preceding heading text (last element of context stack). If no heading has opened yet, source = chapter name.
- Placeholder arrow = `contain` (CONTAINS placeholder).
- Nested bullets: source = the parent bullet's node (not the heading).

**Ordered lists (numbered)**
- Wrap the list in `+:: _sequence_ ::` / `-:: _sequence_ ::` to activate auto-chaining with `(then)`.
- First item of the list is linked from the nearest heading via `contain`.
- N4L handles the chain between items automatically in sequence mode.

**Tables**
- Row 1 is the header row. Column 1 is the key; remaining columns produce edges from the key to the cell value.
- Each data row: first-column value becomes a node; remaining columns → `(note)` edges to cell values.
- Include a `# SUGGEST: column '<header>'` trailing comment on each edge so the user can refine arrow choice per column.

**Links `[text](url)`**
- `text` becomes the target node (stripped).
- Source = current nearest heading / bullet. Placeholder arrow = `see` (NEAR placeholder).
- Append `# url:<href>` to the provenance comment (informational; preserves the URL).

**Blockquotes**
- The blockquote content (joined across continuation lines) becomes a node. Wrap in quotes.
- Edge: nearest heading `(note)` blockquote-node with `# SUGGEST: consider (quote) from EXPRESS`.

**Fenced code blocks**
- The fence content becomes a single quoted node (preserve newlines as `\n` within the quoted string or emit as a multi-line quoted block — pick whichever your fixture set validates).
- Edge: nearest heading `(note)` code-node with `# SUGGEST: consider (example) from EXPRESS, language=<lang>` if a language tag is present.

**Paragraph prose**
- Skipped entirely in v1. This skill does not fractionate prose. If the file is prose-heavy, the top-of-file NOTE from Step 4 tells the user to run `text2N4L` instead.

### Step 6: Node Naming and Collision Handling

**Strip inline markdown** from every node's source text before use:
- `**bold**` → `bold`
- `*italic*` → `italic`
- `_italic_` → `italic`
- `` `code` `` → `code`
- `[text](url)` → `text` (when the link appears inside a bullet/table cell; the link itself separately creates a node via Step 5)
- HTML tags (e.g., `<b>`, `<i>`) → strip tags, keep inner text.

Trim leading and trailing whitespace.

**Quote the result** (wrap in `"..."`) if it contains any of: `(`, `)`, `,`, a leading/trailing space after trim, or begins with `"`, `@`, `$`, `+`, `-`, `:`, `#`, `/`.

**Context-scoped node identity.** Node identity in the graph is the pair `(current context stack, stripped text)`. N4L already scopes nodes by context — two nodes with identical stripped text under different `+:: ::` contexts are distinct in the graph. The skill does not need to do anything special for cross-context identity.

**Within-context collisions.** If two structural elements under the same context produce the same stripped text:
- Track an in-flight map `{stripped_text: count}` per context block.
- First occurrence is emitted as-is; second appends ` (2)`, third ` (3)`, and so on.
- Apply this rule uniformly regardless of whether the source is a bullet, a link, or a table row. Do not use URL-qualifier forms like `Database [postgres.md]` in v1 — the numeric-suffix rule is the single collision policy.

**Pre-heading content.** If a bullet or list appears before the first heading (e.g., directly after frontmatter), its context stack is just whatever frontmatter tags set. The source node for these top-level bullets is the chapter name.

**After-last-heading content.** After the final heading's scope closes at EOF, any remaining bullets are rare in practice — treat them the same way as pre-heading content (source = chapter; context = frontmatter tags, if any).

### Step 7: Emit Pass-1 Skeleton (in memory)

Build the Pass-1 skeleton string in this order. This is still an in-memory skeleton — not yet written to disk. Pass 2 (Steps 8–10) may modify arrow codes before the final write in Step 11.

1. **Marker line (always first):**
   ```
   # n4l:md-import :: src=<relative-md-path>
   ```

2. **Conditional NOTE lines** (each on its own line):
   - If `prose_ratio >= 0.50`: `# NOTE: this file contains substantial prose; consider running text2N4L for sentence-level extraction.`

3. **Chapter header:**
   ```
   - <frontmatter.title or filename-without-extension>
   ```

4. **Frontmatter tags as base context** (if `tags:` was provided):
   ```
   :: tag1, tag2, tag3 ::
   ```

5. **Walk the edge list from Step 5.** Open and close `+:: ::` / `-:: ::` blocks around edges as the context stack enters/leaves each heading. Within a block, group edges by source node and use ditto `"` for second and subsequent edges from the same source. Align arrow codes vertically within each group for readability.

6. **Provenance comment on every edge line:**
   ```
   <source> (<placeholder>) <target>   # src:<rel-path>:L<line>
   ```
   - For link edges: append ` url:<href>` to the comment.
   - For table edges: append ` SUGGEST: column '<header>'` to the comment.
   - For blockquote edges: append ` SUGGEST: consider (quote)` to the comment.
   - For code-fence edges: append ` SUGGEST: consider (example)` plus ` language=<lang>` if present.

7. **Blank line between edge-groups** (each group is all edges sharing a source node).

### Step 8: Pass 2 — Locate SSTconfig/ and Load Arrow Catalogs

Find the arrow definition files in this order:

1. Use the Glob tool to check for `SSTconfig/` in the current working directory.
2. If not found, check `$SSTORYTIME_HOME` using Bash: `echo $SSTORYTIME_HOME`; if set, look for `SSTconfig/` there.
3. If still not found, **skip Pass 2 entirely**. Insert a second conditional NOTE line near the top of the skeleton:
   ```
   # NOTE: SSTconfig/ not found. Placeholders kept. Set $SSTORYTIME_HOME or run from your project root.
   ```
   Then jump to Step 11.

Do NOT prompt the user — matches the `n4l:import` silent-fallback convention.

If `SSTconfig/` is located, read all four arrow definition files:

- `arrows-NR-0.sst` — NEAR (symmetric)
- `arrows-LT-1.sst` — LEADSTO (directional)
- `arrows-CN-2.sst` — CONTAINS (directional)
- `arrows-EP-3.sst` — EXPRESS (directional)

Parse per the format documented in `@./references/arrow-types.md` ("Arrow Definition File Format" section). Build four per-meta-type candidate lists:

- NEAR candidates: `[(short_code, label), ...]` from symmetric `name (code)` lines.
- LEADSTO / CONTAINS / EXPRESS candidates: for each `+ forward (code) - reverse (code)` line, emit **both** directions as candidates so Pass 2 can pick either orientation. Keep them tagged: `{"code": "cause", "label": "causes", "direction": "forward", "pair_code": "cause-by"}`.

Skip lines matching `:: tag1, tag2 ::` (these are organizational groupings within arrow files, not arrow definitions).

### Step 9: Pass 2 — Specialize Each Placeholder Arrow

For each edge line in the Pass-1 skeleton whose arrow is exactly one of the four placeholders (`contain`, `fwd`, `see`, `note`), decide whether to specialize.

**Per-edge decision procedure:**

1. **Identify the meta-type** from the placeholder:
   - `contain` → CONTAINS candidates
   - `fwd` → LEADSTO candidates
   - `see` → NEAR candidates
   - `note` → EXPRESS candidates

2. **Gather decision context:**
   - Source node text
   - Target node text
   - Heading breadcrumb (full chain of open `+:: ::` contexts at this edge's position)
   - Any existing `# SUGGEST:` hint already on the edge line (e.g., `SUGGEST: column 'Director'`, `SUGGEST: consider (quote)`, `SUGGEST: consider (example) language=python`, `url:<href>`)
   - The candidate list for the matching meta-type

3. **Pick the best code** using these heuristics (in order):

   a. **Honor existing hints.** If the edge already carries a `SUGGEST: consider (code)` hint (blockquote → `quote`, code fence → `example`), and that code exists in the candidate list, pick it.

   b. **Keyword match the node text.** Look for explicit relationship words in the source or target: "causes"/"leads to" → `(cause)`/`(fwd)`; "means"/"definition" → `(means)`; "example"/"e.g." → `(e.g.)`; "contains"/"has"/"ingredient" → `(contain)`/`(has-pt)`/`(ingred)`; "consists of" → `(consists)`; "note"/"remark" → `(note)`; "describes" → `(describe)`; "synonym"/"same as" → `(syn)`/`(=)`; "see also"/"related" → `(see)`/`(related)`.

   c. **Match the heading breadcrumb.** Recipes/cooking contexts favor `(ingred)` over `(contain)`. Processes/steps/workflow contexts favor `(then)`/`(next)` over `(fwd)`. Glossary/definitions favor `(means)`/`(describe)` over `(note)`. Taxonomy/categories favor `(is-a)` (if present) or `(contain)`.

   d. **Link edges specifically.** For a `see` placeholder that came from a markdown link, prefer `(ref)` / `(cite)` / `(source)` if present, else keep `(see)`.

   e. **Table column edges.** The `SUGGEST: column '<header>'` hint carries a column name; treat it as a potential arrow-label match (e.g., column `role` → `(role)`, column `description` → `(describe)`, column `category` → `(is-a)` or `(contain)`).

4. **Confidence gating:**

   - If a single candidate is distinctly better than the placeholder AND clearly better than all alternatives → emit the specialized code with **no extra comment**.
   - If two (or more) candidates are equally plausible → pick the first one and append ` # REVIEW: or (<alt>)?` listing **exactly one** alternative. Never list more than one REVIEW alternative per edge.
   - If no candidate is confidently better than the placeholder → keep the placeholder and append ` # SUGGEST: review against <arrow-file>.sst` (where `<arrow-file>` is the matching meta-type file name).

5. **Preserve the provenance comment unchanged.** The `# src:<path>:L<n>` and any `url:...` / `SUGGEST: column '<x>'` annotations remain on the line. REVIEW/SUGGEST arrow-choice comments are appended after the provenance comment with a space separator:
   ```
   <source> (ingred) <target>   # src:notes.md:L12 # REVIEW: or (contain)?
   ```

6. **Do not change node text, ditto alignment, or context blocks.** Pass 2 only rewrites the arrow code in parentheses; everything else is byte-preserved from the Pass-1 skeleton.

### Step 10: Pass 2 — Chunking for Large Skeletons

If the in-memory skeleton exceeds approximately 40k tokens (rough guide: >30,000 bytes of skeleton text), process it in chunks rather than one pass:

1. Split the skeleton at top-level heading boundaries (where a `+:: <top-heading-text> ::` opens at indent level 0 of the context stack). Each chunk is one top-level heading's scope.
2. Include the full chapter header + frontmatter tag context at the top of every chunk so Claude has consistent base context.
3. Include the full candidate list (all four meta-types) in every chunk prompt — it's ~6K tokens, cheap to duplicate for modest chunk counts.
4. Process each chunk independently. Each chunk self-contains its `+::` / `-::` blocks.
5. If the heuristic would split inside a `+:: _sequence_ ::` / `-:: _sequence_ ::` block, bump the chunk boundary to include the whole sequence.
6. Reassemble chunks in order; the output is a single file.

For skeletons under the threshold, process the whole thing in one pass.

### Step 11: Write the Output and Report

Use the Write tool to save the final (post-Pass-2) output. Report:

- Output file path.
- Edge count.
- Chapter name.
- Distinct context count (number of headings encountered).
- Prose-ratio (so the user knows whether the NOTE was emitted).
- Pass 2 stats: count of specialized edges, count of `# REVIEW:` tags, count of `# SUGGEST:` tags, and whether Pass 2 ran at all (skipped if no SSTconfig).
- Review hint: *"Review `# REVIEW:` tags before upload. Run `/n4l:upload <output-path>` to validate and upload."*

## Worked Example

### Input (`cooking.md`)

```markdown
---
title: Italian Cooking Notes
tags: [food, cooking]
---

# Italian Cooking

## Ingredients
- pasta
- tomato sauce
- [parmesan cheese](https://en.wikipedia.org/wiki/Parmigiano-Reggiano)

## Recipe

1. Boil water
2. Add salt
3. Add pasta
4. Drain and serve
```

### Command

```
/n4l:md-import cooking.md
```

### Pass-1 skeleton (intermediate, not written to disk)

```
# n4l:md-import :: src=cooking.md

- Italian Cooking Notes
:: food, cooking ::

+:: Italian Cooking ::
+:: Ingredients ::

Ingredients (contain) pasta                # src:cooking.md:L9
"          (contain) tomato sauce          # src:cooking.md:L10
"          (see)     parmesan cheese       # src:cooking.md:L11 url:https://en.wikipedia.org/wiki/Parmigiano-Reggiano

-:: Ingredients ::
+:: Recipe ::
+:: _sequence_ ::

Recipe (contain) Boil water                # src:cooking.md:L15
Boil water
Add salt
Add pasta
Drain and serve

-:: _sequence_ ::
-:: Recipe ::
-:: Italian Cooking ::
```

### Output (`cooking.n4l`, after Pass 2 specialization)

Pass 2 reads the skeleton + `SSTconfig/arrows-*.sst` + heading breadcrumb. Under `Ingredients`, the placeholder `(contain)` for food items specializes to `(ingred)` (food domain). The link to a Wikipedia article stays `(see)` — it's a generic reference, not a domain-specific relationship. Pass 2 does not touch the sequence-mode items (they're implicitly `(then)`-chained by `+:: _sequence_ ::`, no placeholder present).

```
# n4l:md-import :: src=cooking.md

- Italian Cooking Notes
:: food, cooking ::

+:: Italian Cooking ::
+:: Ingredients ::

Ingredients (ingred) pasta                 # src:cooking.md:L9
"          (ingred) tomato sauce           # src:cooking.md:L10
"          (see)    parmesan cheese        # src:cooking.md:L11 url:https://en.wikipedia.org/wiki/Parmigiano-Reggiano

-:: Ingredients ::
+:: Recipe ::
+:: _sequence_ ::

Recipe (contain) Boil water                # src:cooking.md:L15 # REVIEW: or (has-pt)?
Boil water
Add salt
Add pasta
Drain and serve

-:: _sequence_ ::
-:: Recipe ::
-:: Italian Cooking ::
```

The `Recipe (contain) Boil water` edge was flagged because `(contain)` and `(has-pt)` are both reasonable for "a recipe is made up of these steps" — the user can pick during review. If `SSTconfig/` was not found at Pass 2 time, the skeleton above is written as-is with a `# NOTE: SSTconfig/ not found.` line near the top and no REVIEW tags.

## Special Handling Notes

- **Empty file** → emit marker + empty chapter + exit cleanly. No edges, no NOTE.
- **Frontmatter-only file** → emit marker + chapter from `title:` + optional base context. No edges.
- **Unclosed code fence** → treat the unclosed region as a single code-fence node; emit a trailing `# REVIEW: unclosed code fence starting at L<n>` comment.
- **Setext headings** (`Title\n=====`) → treat identically to ATX equivalents (`====` = H1, `----` = H2).
- **Multi-H1** → each H1 opens its own `+:: ::` context; the previous H1's scope closes implicitly when the next H1 opens. Chapter remains the filename or `frontmatter.title`.
- **HTML embedded in bullets** → strip tags; keep inner text.
- **Mixed-language content** → preserved as-is (UTF-8); aligns with existing `c.s.peirce_semiotics.n4l` pattern.
- **Deeply nested lists** → chain CONTAINS from each parent bullet; no special handling beyond the nesting structure.
- **Wikilinks `[[foo]]`** → treated as literal text in v1 (included in the stripped node text unchanged). Multi-file wikilink resolution is future work.
- **Obsidian tags `#foo` inline** → treated as literal text; NOT promoted to context tags. (Only YAML frontmatter `tags:` becomes context.)
- **YAML frontmatter parse failure** → fall back to filename as chapter; emit `# NOTE: frontmatter parse failed at L<n>` as the first NOTE line.
- **Duplicate stripped text within a context** → second occurrence gets `(2)`, third `(3)`, etc.
- **Large files (>500 lines)** → no special handling in v1; the prompt processes the whole file in one pass. If this becomes a performance concern, report it and the skill will be updated to chunk by top-level heading.

## Relationship to Other Skills

- **`/n4l:upload`** — run this after `md-import` to validate (`N4L -v`) and upload the generated `.n4l`.
- **`/n4l:interpret`** — use on search results if you want short-code annotations explained.
- **`text2N4L`** (binary, not a skill) — use for prose-heavy files that `md-import` skips.
- **`/n4l:import`** — for CSV files. `md-import` is the markdown equivalent; the two do not share code.

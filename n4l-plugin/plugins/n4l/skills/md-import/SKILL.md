---
name: n4l:md-import
description: "Convert a Markdown file into N4L knowledge graph format using structure-first parsing. Maps headings to contexts, lists to CONTAINS/LEADSTO edges, tables to row-per-node, and links/blockquotes/code fences to EXPRESS annotations. Emits generic placeholder arrows with SUGGEST comments for later specialization. Use when the user says 'import markdown', 'md to n4l', 'convert markdown', or 'markdown to knowledge graph'."
argument-hint: "<markdown-file-path> [--output filename.n4l] [--force]"
---

# N4L Markdown Import Skill

Convert a single Markdown file into a valid N4L file by walking the document's structure and emitting generic placeholder arrows with provenance comments. This is **Pass 1 only** (structural parse) — arrow specialization against SSTconfig/ (Pass 2) and merge-on-rerun are deferred to future iterations. The output is valid N4L that uploads cleanly via `/n4l:upload`; review the `# SUGGEST:` comments to refine arrow choices by hand.

## Reference Material

@./references/n4l-syntax.md
@./references/arrow-types.md

## What this skill does (and doesn't)

**Does:**
- Parses markdown structure: YAML frontmatter, H1–H6 headings, unordered/ordered lists, tables, links, blockquotes, code fences.
- Emits a valid `.n4l` file with generic placeholder arrows (`contain`, `fwd`, `see`, `note`) that are real SSTconfig codes — unspecialized output uploads cleanly.
- Tags every edge with `# src:<path>:L<line>` provenance comments.
- Warns when a file is mostly prose (suggest running `text2N4L` instead).
- Aborts on existing output unless `--force` is used (prevents clobbering hand-edited files).

**Does not (deferred to future iterations):**
- Specialize placeholder arrows using SSTconfig/ and heading context (Pass 2).
- Detect hand-edits in existing `.n4l` and preserve them on re-run (merge).
- Import from Obsidian vaults with wikilink resolution (multi-file).
- Fractionate prose into sentences (use `text2N4L` for that).
- Interpret wikilinks `[[foo]]` or Obsidian tags `#foo` as structure (treated as literal text).

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

### Step 7: Emit the Output File

Build the output in this order:

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

### Step 8: Write the Output and Report

Use the Write tool to save the output. Report:

- Output file path.
- Edge count.
- Chapter name.
- Distinct context count (number of headings encountered).
- Prose-ratio (so the user knows whether the NOTE was emitted).
- Review hint: *"Output uses generic placeholder arrows (`contain`, `fwd`, `see`, `note`). Specialize by hand against SSTconfig/ or run `/n4l:upload` to validate and upload."*

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

### Output (`cooking.n4l`)

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

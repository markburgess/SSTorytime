# md-import fixtures

Acceptance inputs for `/n4l:md-import`. Each `.md` file has a sibling `.expected.n4l` showing the intended shape of the skill's output.

## Why structural-property assertions, not byte-diff

The skill is prompt-based — Pass 1 is Claude following `../SKILL.md` instructions. LLM output has minor run-to-run variation in whitespace, ditto alignment, and quoting that would break strict byte-diff tests. Expected outputs are **representative targets**, not contract. Validate by checking:

1. **Marker line present:** first non-blank line matches `^# n4l:md-import :: src=`.
2. **Every edge has `# src:` provenance** with a line number.
3. **Edge tuple set** (source, arrow, target) matches the expected output's tuple set exactly.
4. **Context stack balance:** every `+:: X ::` has a matching `-:: X ::` in LIFO order.
5. **Placeholder arrows only:** arrows in emitted edges are drawn from `{contain, fwd, see, note}` (Pass 2 specialization is deferred to a future iteration).
6. **Chapter line present:** matches `^- <title-or-filename>`.
7. **`N4L -v` accepts the output** without syntax errors (integration check; requires the `N4L` binary on PATH).

## Fixtures

| File | Validates |
|---|---|
| `simple-notes.md` | Frontmatter (title, tags), heading hierarchy, unordered list, ordered list → `_sequence_`, link-with-url. |
| `table-heavy.md` | Markdown table → row-per-node with column-header SUGGEST comments. First-column quoting when value contains `:`. |
| `collision.md` | Numeric-suffix collision rule for duplicate stripped text within a context. Cross-context duplicates (two `## Usage` under different H1s) are NOT collisions because contexts scope identity. |

## Not covered yet

The following fixtures would round out coverage and can be added when the skill moves past v1 scaffold:

- Empty file / frontmatter-only / YAML parse failure.
- Unclosed code fence.
- Setext headings (`Title\n=====`).
- Multi-H1 files.
- HTML tags embedded in bullets.
- Mixed-language (zh/en) node names.
- Deeply nested lists (5+ levels).
- Prose-heavy file (R6 NOTE trigger).
- Code-fence-heavy file (R6 NOTE non-trigger — false-positive guard).

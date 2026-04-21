# Bringing your own corpus — from prose to graph

The helper tool `text2N4L` will take a plaintext document and propose an N4L
file from it. The operative word is *propose*. The draft it produces is
deliberately imperfect: it picks out sentences it thinks are high-signal,
sketches containment links, and decorates everything with n-gram context
tags, and then it hands the result to you to argue with. That argument is
the point. A corpus you have not argued with is not a corpus you know.

Expect to spend real time on this — certainly longer than it takes to run
the command. The first pass gives you a rough shape; the second pass gives
you chapters you believe in; later passes turn the chapters into something
you can tell a story from. Everything in SSTorytime is designed to make
that iteration cheap: the `.n4l` file is plain text, the upload is
idempotent, `-wipe` lets you start over without regret. Treat the tool
chain as a loom rather than a printer.

This cookbook walks the mechanics: generate a draft, refine it, upload it,
search it. The judgement happens between the steps, not inside them.

!!! info "Prerequisites"
    - `src/bin/N4L`, `src/bin/text2N4L`, `src/bin/searchN4L` compiled (run `make` from repo root).
    - A running PostgreSQL instance with SSTorytime schema loaded (`make db`).
    - A plaintext `.txt` document you want to explore. For this walkthrough we'll pretend it's a collection of project retrospectives about a home-security platform, but any prose works.

## 1. Drop the source document in place

We follow the convention of keeping raw corpora under `examples/` alongside the existing sample data.

```bash
cd examples
cat > mycorpus.txt <<'EOF'
The system was originally designed as a single Raspberry Pi
running in the garage. Early tests showed that the camera
latency was too high for real-time motion alerts. After a
redesign we moved the inference pipeline onto a small x86 box
and kept the Pi as a network relay. The first production cut
went live in September and caught a delivery driver who tried
to move a package from the porch.
EOF
```

For a real run, substitute any plain UTF-8 text — meeting notes, essay drafts, an interview transcript, a book chapter.

## 2. Fractionate the text

`text2N4L` reads the file and picks out the highest-signal sentences, writing a proposed N4L skeleton alongside the source:

```bash
../src/bin/text2N4L -% 30 mycorpus.txt
```

What this produces:

- `mycorpus.txt_edit_me.n4l` — a new N4L file ready for manual refinement.
- The `-% 30` asks for approximately 30% of the sentences; in practice you will get slightly more (see [text2N4L](../text2N4L.md#percentage-semantics)).
- Each selected sentence becomes a `@senN` item with `(is in) partN of mycorpus.txt` links.
- A `_sequence_` context pulls the sentences into a running narrative.
- N-gram phrases from the source are added as **context tags** (this is what makes the output searchable by topic rather than just by sentence).

Sample output header:

```
 - Samples from mycorpus.txt

 # TABLE OF CONTENTS ...
 # themes and topics
 # selected samples
 # final fraction 40.00 of requested 30.00
```

## 3. Open and refine

This is where the human judgement goes. Open the generated file in your editor of choice:

```bash
$EDITOR mycorpus.txt_edit_me.n4l
```

Things to do while reading:

- **Split into chapters.** Replace the single `- Samples from mycorpus.txt` header with one `- <chapter>` line per conceptual section. Chapters are how you scope later searches.
- **Add arrows.** The generator only emits `(is in)` containment links. Wherever two sentences share a concept (a cause, a quote, a continuation), add an explicit arrow. Use the four STtype families described in [arrows.md](../arrows.md):
    - `(leads to)` / `(then)` — causal or temporal
    - `(contains)` / `(is part of)` — composition
    - `(expresses)` / `(is described by)` — property
    - `(is similar to)` / `(=)` — proximity
- **Fix ambiguity.** If the text talks about "the system" in several places, rename each `@senN` anchor (e.g. `@garage_system`, `@production_system`) so that later searches distinguish them.
- **Delete noise.** Boilerplate sentences that `text2N4L` picked up but that add no signal — kill them.

!!! tip "Iterate small"
    Don't try to perfect every chapter on the first pass. Get a rough structure, upload it,
    search it, see what's missing, then go back and refine. The tool chain is cheap to re-run.

## 4. Upload to the graph

Once the file parses cleanly (`N4L mycorpus.txt_edit_me.n4l` with no errors), upload:

```bash
../src/bin/N4L -u mycorpus.txt_edit_me.n4l
```

For an atomic re-upload that clears previous state (useful during iteration), use the wipe pattern:

```bash
../src/bin/N4L -wipe -u mycorpus.txt_edit_me.n4l
```

Expect a few seconds of progress output followed by `Upload complete.` (or an error pointing to the offending line).

!!! warning "`-wipe` is destructive"
    `-wipe` drops **all** SSTorytime tables and recreates them. If you have other corpora
    loaded alongside, re-upload them in the same invocation: `N4L -wipe -u *.n4l`.

## 5. Search and browse

The graph is now queryable. Three useful first queries:

```bash
# Substring search across all chapters
../src/bin/searchN4L "delivery driver"

# Browse notes in original order
../src/bin/searchN4L "\\notes mycorpus"

# Find paths from one idea to another
../src/bin/searchN4L "\\from Raspberry \\to production"
```

Full query grammar is in [searchN4L.md](../searchN4L.md#the-query-dsl). For more recipes see [10 search recipes](search-recipes.md).

## Next steps

- When the structure feels right, commit the refined `.n4l` file to version control. That file — not the database — is your source of truth.
- To delete a draft chapter cleanly: `../src/bin/removeN4L -force "chapter name"`.
- To see graph-level statistics: `../src/bin/graph_report -chapter mycorpus -sttype L,C`.

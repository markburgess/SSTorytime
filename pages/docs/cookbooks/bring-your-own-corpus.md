# Patterns — research notes

> **You have a document. You want a graph of it. This is the walk from
> plaintext to queryable notes, with the honest bits in plain sight.**

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

!!! info "Before you start"
    - `N4L`, `text2N4L`, and `searchN4L` on your `$PATH` (from
      [Install in 5 minutes](../GettingStarted.md)).
    - A running PostgreSQL instance loaded with the SSTorytime schema.
    - A plaintext `.txt` document you want to explore. This
      walkthrough uses an imagined set of project retrospectives
      from a home-security platform — any prose works equally well:
      meeting notes, interview transcripts, book chapters, essay
      drafts. If you'd rather see the pattern on something smaller
      and already in the repo, the reading list at
      `examples/reading-list.n4l` was written by hand and makes a
      good comparison when you want to see what a polished,
      fully-edited N4L file looks like.

---

## 1. Drop the source document in place

A habit worth picking up: keep raw corpora under `examples/`
alongside the sample data that ships with the project. The files
are text, git handles them happily, and future-you will thank
past-you for keeping the source next to the N4L that came out of
it.

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

For a real run, substitute any plain UTF-8 text.

---

## 2. Fractionate the text

`text2N4L` reads the file and picks out the highest-signal
sentences, writing a proposed N4L skeleton alongside the source:

```bash
text2N4L -% 30 mycorpus.txt
```

What you'll find:

- A new file `mycorpus.txt_edit_me.n4l` sitting next to the source.
- `-% 30` asked for about 30% of the sentences; the output will
  usually be a little more than that, because two internal
  heuristics vote and their union gets kept.
- Each selected sentence becomes an aliased item (`@senN`) with
  a containment link back to the source document.
- A `_sequence_` context pulls the sentences into a running
  narrative so you can still read the document in order.
- N-gram phrases from the source are added as context tags —
  this is what makes the output searchable by topic rather than
  only by sentence.

A typical file header looks like:

```
 - Samples from mycorpus.txt

 # TABLE OF CONTENTS ...
 # themes and topics
 # selected samples
 # final fraction 40.00 of requested 30.00
```

See [Turning documents into stories](../text2N4L.md) for the
bigger picture on when to reach for `text2N4L` and when to write
N4L by hand instead.

---

## 3. Open and refine

This is where the human judgement goes. Open the generated file in
your editor of choice:

```bash
$EDITOR mycorpus.txt_edit_me.n4l
```

Things to do while reading:

- **Split into chapters.** Replace the single `- Samples from
  mycorpus.txt` header with one `- <chapter>` line per conceptual
  section. Chapters are how you scope later searches, and a file
  with one huge chapter is painful to query in the same way a
  book with no table of contents is painful to re-read.
- **Add arrows.** The generator only emits containment links — "this
  sentence is part of the document." Everywhere two sentences
  share a concept, add an explicit arrow. Pick from the four
  shapes: `(then)` for sequence, `(contains)` for membership,
  `(about)` or `(by)` for properties, `(see also)` for adjacency.
  [Thinking in arrows](../arrows.md) has the full catalogue.
- **Fix ambiguity.** If "the system" appears in several places and
  means different things, rename the anchors (`@garage_system`,
  `@production_system`) so later searches keep them apart.
- **Delete noise.** Sentences the tool picked up that add no
  signal: kill them.

!!! tip "Iterate small"
    Don't try to perfect every chapter on the first pass. Get a
    rough structure, upload it, search it, see what's missing, then
    go back and refine. The tool chain is cheap to re-run.

---

## 4. Upload to the graph

Once the file parses cleanly (`N4L mycorpus.txt_edit_me.n4l` with
no errors), upload:

```bash
N4L -u mycorpus.txt_edit_me.n4l
```

For an atomic re-upload that clears previous state — useful
during iteration — run with the wipe flag:

```bash
N4L -wipe -u mycorpus.txt_edit_me.n4l
```

Expect a few seconds of progress output followed by a `Finally
done!` line, or an error pointing at the offending input line.

!!! warning "`-wipe` is destructive"
    `-wipe` drops all SSTorytime state and rebuilds it from the
    files you give it. If you have other corpora already loaded,
    re-upload them in the same command: `N4L -wipe -u *.n4l`.

---

## 5. Search and browse

The graph is now queryable. Three useful first queries:

```bash
# Substring search across all chapters
searchN4L "delivery driver"

# Browse notes in original input order for one chapter
searchN4L "\\notes mycorpus"

# Find paths from one idea to another
searchN4L "\\from Raspberry \\to production"
```

[Finding things](../searchN4L.md) has the shape of a question and
the commands you'll type. For patterns on larger corpora, see
[Patterns — search recipes](search-recipes.md).

---

## Next steps

- When the structure feels right, commit the refined `.n4l` file
  to version control. That file — not the database — is your
  source of truth.
- Re-run `N4L -wipe -u *.n4l` any time you want a clean slate.
  The upload is idempotent and fast.
- If you need to remove a chapter without wiping everything, the
  `removeN4L` tool in the repo's `developers/` folder handles
  targeted deletion; most users never reach for it.

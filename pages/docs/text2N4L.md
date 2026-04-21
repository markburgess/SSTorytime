# Turning documents into stories

![Six tools arranged on a worktable — sieve, builder's square, magnifying glass, open book, twin flashlights, microscope — representing the family of CLI utilities in the SSTorytime toolkit.](figs/tools_hero.jpg){ align=center }

> **`text2N4L` takes a plain-text document and writes a first-draft N4L file
> from it — so you can start from something instead of from nothing.**

Sometimes the thing you want to put in the graph is already written:
a book, a transcript, a set of meeting minutes, a long essay. Turning
that by hand into N4L — splitting it into items, picking out the
relationships, choosing arrows — would take hours you don't have.

`text2N4L` gives you a head start. It reads the source file, picks
out the sentences it thinks carry the most signal, and writes out an
`.n4l` file with the skeleton of a graph inside it — one item per
retained sentence, containment links back to the source, topic tags
derived from the prose. Nothing is uploaded to the database; the
output is a text file you open and argue with.

That argument is the point. The draft is deliberately imperfect.
A corpus you haven't edited is a corpus you don't know.

---

## When to reach for it

**Reach for it when:**

- The source is already written down and you want to work *with* it,
  not *rewrite* it.
- You're exploring an unfamiliar text and want a scaffold to think
  against — which sentences matter, which topics recur.
- You're going to spend time editing regardless; starting from a
  draft is faster than starting from blank.

**Skip it when:**

- You're writing something new. Just write N4L directly — the
  tutorial walks you through it in fifteen minutes.
- The source is short (a paragraph, a page). Hand-writing the N4L
  will be faster than running the tool and cleaning up its guesses.
- The source is a table or a list with structure `text2N4L`
  can't see. A spreadsheet export or a JSON file is a better
  starting point via other tooling.

---

## The shape of the workflow

Four steps, not all at the keyboard:

1. **Run the tool** against your `.txt` file. Get back an
   `.n4l` file to edit.
2. **Open the result in your editor.** Read what it picked up.
   Delete the noise, rename the chapters, group things that belong
   together.
3. **Add arrows.** The generator only emits containment links —
   "this sentence is part of the document." You add the
   interesting ones: "this statement leads to that", "this thing is
   a property of that".
4. **Ingest the edited file** with `N4L -u`, and start asking
   questions.

Budget real time for step 2. More than for step 1. Possibly more
than steps 1, 3, and 4 put together. That's the design.

---

## One run, end to end

A full walkthrough of this — with a concrete example corpus, what
the draft looks like, what to do with it — lives in the
[Patterns — research notes](cookbooks/bring-your-own-corpus.md)
cookbook. Read that if you have a document in mind; it's the same
tool shown in action.

If you just want to see it work, the project ships sample texts
under `examples/example_data/`. Try one of the shorter pieces:

```
text2N4L examples/example_data/promisetheory1.dat
```

You'll get back a sibling `.n4l` file (`..._edit_me.n4l`) with
roughly half the source's sentences retained, each one an item in
a chapter, with n-gram topic tags derived from the prose. Open
that file in an editor and you're at step 2.

---

## What comes out

The generated file is a normal N4L file — the same format you'd
write by hand. A typical entry looks like this (from a Moby-Dick
sample):

```n4l
@sen9471   Towards thee I roll, thou all-destroying but unconquering whale,
              " (is in) part 210 of ../examples/example_data/MobyDick.dat

@sen9473   and since neither can be mine, let me then tow to pieces,
              " (is in) part 210 of ../examples/example_data/MobyDick.dat
```

Each retained sentence becomes an aliased item (`@senN`) with a
containment link back to the source. Topic tags from n-gram
analysis appear as contexts. Nothing else is linked yet — that's
your job.

You can reference a generated item from your own additions using
the alias:

```n4l
 $sen9471.1 (note) This line was made famous by Khan in Wrath of Khan.
```

---

## Honest caveats

The tool picks sentences using a heuristic for "intent" — how
load-bearing the sentence looks. It oversamples on purpose: if two
different heuristics both flag a sentence, it's kept. That means
asking for 30% usually gets you more than 30%, which matters when
your source is long.

Large sources are expensive. A book-sized text — *Moby Dick*,
*Origin of Species* — will produce enough cross-referenced output
that the downstream `N4L -u` ingest takes hours, not minutes. If
that's your use case, split the source into chapters before running
`text2N4L`, and ingest chapter-by-chapter.

The tool is not a replacement for reading the text. It's a
scaffold. The reading is yours.

---

## Adjusting the sample fraction

One flag you'll reach for: the target fraction of sentences to
keep.

```
text2N4L -% 30 mycorpus.txt
```

30% asks for fewer sentences, 77% asks for more, 100% is "every
sentence, no selection". Run `text2N4L --help` for the full flag
list; it's a short list.

---

## Where to go next

<div class="grid cards" markdown>

-   :material-notebook-edit:{ .lg .middle } **Walk through a real corpus**

    ---

    Full cookbook: drop in a `.txt`, fractionate, refine, upload,
    search.

    [:octicons-arrow-right-24: Patterns — research notes](cookbooks/bring-your-own-corpus.md)

-   :material-pencil:{ .lg .middle } **Write the edited version**

    ---

    Chapters, contexts, arrows — the notation the tool's output is
    in, so you can read and refine it.

    [:octicons-arrow-right-24: Writing N4L by hand](N4L.md)

-   :material-magnify:{ .lg .middle } **Query what you ingested**

    ---

    Once the corpus is in, ask it questions.

    [:octicons-arrow-right-24: Finding things](searchN4L.md)

</div>

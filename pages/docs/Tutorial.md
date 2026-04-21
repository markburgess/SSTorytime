# Your first story

> **In fifteen minutes you will have a queryable graph of your own reading list — and a feel for how N4L thinks.**

You will write a small reading list by hand, load it, and ask it three
questions. The last question's answer is the point of the whole page:
connections you never explicitly asked for surface on their own.

This page assumes you have built the binaries and loaded the database per
[Install in 5 minutes](GettingStarted.md). If `./src/bin/N4L` exists and
`./src/bin/searchN4L` runs without a connection error, you are ready.

---

## What we're going to build

A short file in N4L — roughly seven books, a handful of topics they are
about, who wrote them, which ones cite each other, and a one-line
takeaway for each. Small enough to keep in your head, rich enough to ask
real questions of:

- *What have I read about decision making?*
- *What are my notes on Thinking Fast and Slow?*
- *What else does Daniel Kahneman connect to in my reading?*

You will type ~20 lines, run one ingest command, and run three queries.
Total time: fifteen minutes, give or take however long you spend squinting
at your first query result.

---

## Writing it down

We are going to build the file up in pieces, not dump it all at once.
Each piece adds one more thing the graph can answer. Open an empty file
called `my-reading.n4l` in your editor.

### Your first book

Start with a title line and one book:

```n4l
- my reading list

 :: books, topics, authors ::

 Thinking Fast and Slow   (about)    decision making
        "                 (by)       Daniel Kahneman
```

Three things to notice.

The line starting with `-` names the chapter. Everything else is inside it.
The line starting with `::` is a context — a set of tags that frame the
section. You can ignore both of these for now; they matter when your graph
gets big enough that you want to scope searches.

Bare phrases like `Thinking Fast and Slow` and `Daniel Kahneman` are
nodes. The things in parentheses — `(about)`, `(by)` — are arrows. An
arrow goes from the thing on its left to the thing on its right. So the
first line says: *this book is about decision making.* The second line
says: *the book is by Daniel Kahneman.*

The big quote on the second line (`"`) is a ditto mark. It means "the
node from the line above." You do not have to retype the book's title.

Your graph now contains one book, one topic, one author, and two arrows
connecting them.

### A second book, and a shared topic

Add a second book that happens to be about the same thing:

```n4l
 Superforecasting         (about)    decision making
        "                 (by)       Philip Tetlock
```

You did not tell the graph that *Thinking Fast and Slow* and
*Superforecasting* have anything to do with each other. You just said each
of them points at `decision making`. But they do now: two arrows,
same target. Ask about the topic and both books come back.

This is the core of it. You do not model relationships directly. You write
down what each thing points at, and the relationships fall out.

### A citation

Add one more arrow:

```n4l
 Superforecasting         (bib-cite) Thinking Fast and Slow
```

`(bib-cite)` is the *has bibtex citation* arrow — N4L's shorthand for
"one work cites another." Your graph now has *two* paths from one book
to the other: the shared-topic path we already had, and the direct
citation you just added. Both surface in the query outputs below.

### A few more books

Add three more so we have something to query:

```n4l
 Thinking Fast and Slow     (bib-cite) Judgment under Uncertainty
        "                   (note)     two systems, one of them lazy, both of them you

 Judgment under Uncertainty (about)    heuristics
        "                   (by)       Daniel Kahneman
        "                   (by)       Amos Tversky

 Thinking in Systems        (about)    decision making
        "                   (by)       Donella Meadows
```

`(note)` attaches a short remark to a node; you can hang as many notes
off a book as you like. `(by)` on *Judgment under Uncertainty* appears
twice — a book with two authors is just a node with two `(by)` arrows.

Arrow types in N4L are pre-declared: `about`, `by`, `bib-cite`, `note`,
and dozens more ship in the project's config. [Thinking in arrows](arrows.md)
has the full list; for this page the ones above are all you need.

### The full corpus

What you have is a decent start. The complete version of this reading list
is already in the repo at `examples/reading-list.n4l` — seven books, more
topics, more citations, one-line takeaways. Same shape as what you just
wrote; more of it.

For the rest of this page we will use the full file. Either keep editing
your own, or switch to the canonical one.

---

## Loading it

One command from the repo root:

```
./src/bin/N4L -u examples/reading-list.n4l
```

`-u` means *upload to the database*. You will see a short progress block:

```
Uploading nodes..

Storing primary nodes ...

.

(0.0%) uploading . . .

 1 )   J O T   I T   D O W N
Storing Arrows...
Storing inverse Arrows...
Storing contexts...
Storing page map...
 W H E N   Y O U   T H I N K   O F Indexing ....
Finally done!
```

If you see `Finally done!` and no red error lines, the corpus is in. If
you see a connection error instead, your database is not running — back
to [Install in 5 minutes](GettingStarted.md).

---

## Asking questions

Three queries. Each is a sentence in English, a command, and a reading of
the answer.

### What have I read about decision making?

```
./src/bin/searchN4L "decision making"
```

You search for a string; the graph finds it as a node and shows what
points at it. Something like:

```
0: "decision making"	in chapter: reading list

      -    (is the topic/theme of) - Thinking Fast and Slow
           -    (is a bibtex citation label for) - Superforecasting
      -    (is the topic/theme of) - Superforecasting
      -    (is the topic/theme of) - Thinking in Systems
```

Three books, reachable because each wrote an `(about)` arrow at the same
target. The arrow names render in their inverse form — you wrote
`(about)`, the graph reads it back as `(is the topic/theme of)` from
the topic's side. Same fact, other direction.

Look at the indented line under *Thinking Fast and Slow*: the graph also
noticed that it is the citation target of *Superforecasting*, and tucked
that in. You asked about a topic; you got back the books AND a citation
relationship between two of them, for free.

### What are my notes on Thinking Fast and Slow?

```
./src/bin/searchN4L notes about "Thinking Fast and Slow"
```

This pulls the full orbit of one book:

```
0: "Thinking Fast and Slow"	in chapter: reading list

      -    (is about topic/them) - dual-process cognition
      -    (note/remark) - two systems, one of them lazy, both of them you
      -    (is about topic/them) - decision making
      -    (is about topic/them) - heuristics
      -    (has author) - Daniel Kahneman
      -    (has bibtex citation) - Judgment under Uncertainty
           -    (is about topic/them) - cognitive bias
           -    (has author) - Amos Tversky
           -    (note/remark) - the anchor is almost always the wrong number
      -    (is a bibtex citation label for) - Superforecasting
```

Everything you ever said about this book, plus one step further: the book
it cites, *Judgment under Uncertainty*, shows its own topics, author, and
takeaway indented underneath. The graph walked one hop out without being
asked — that is the orbit.

### What does Daniel Kahneman connect to in my reading?

This is the one.

```
./src/bin/searchN4L "Daniel Kahneman"
```

You ask about a person. You get back:

```
0: "Daniel Kahneman"	in chapter: reading list

      -    (is the author of) - Thinking Fast and Slow
           -    (is a bibtex citation label for) - Superforecasting
      -    (is the author of) - Judgment under Uncertainty
```

Kahneman wrote two books — you told the graph that directly. But the
indented line says *Thinking Fast and Slow* is cited by *Superforecasting*.
You did not ask about Superforecasting. You did not write anything
connecting it to Kahneman. It surfaced because it is one hop away from
something Kahneman authored, and the graph walks one hop out of every
neighbourhood.

That is the whole idea.

---

## What just happened

- **You wrote it.** Twenty lines of plain text. No schema designed in
  advance. No code. No SQL.
- **You loaded it.** One command. The database is a cache; your N4L file
  is still the source of truth and lives in your editor and your version
  control.
- **You asked it.** Three questions in English-shaped commands. The hard
  one — "what does Kahneman connect to?" — surfaced a citation
  relationship you never mentioned in the question, because the graph
  walks one hop out from whatever you ask about.

You just built a knowledge graph. If you extend the file and run the
upload again, it grows. If you throw the database away, `N4L -u` rebuilds
it from your file in seconds.

---

## Where to go next

<div class="grid cards" markdown>

-   :material-pencil:{ .lg .middle } **Keep writing**

    ---

    The arrow types, the chapters, the contexts — more on what they are
    for and when to reach for each.

    [:octicons-arrow-right-24: Thinking in arrows](arrows.md)

-   :material-magnify:{ .lg .middle } **Ask bigger questions**

    ---

    Orbits, paths, context-scoped searches, sequences. The shape of a
    question, not the flags.

    [:octicons-arrow-right-24: Finding things](searchN4L.md)

-   :material-lightbulb-on:{ .lg .middle } **Understand why it works**

    ---

    Semantic spacetime in plain English. Why connections are written the
    way they are, and what that buys you.

    [:octicons-arrow-right-24: Concepts](concepts/why-semantic-spacetime.md)

</div>

# Your first story

> **In fifteen minutes you will have a queryable graph of your own reading list — and a feel for how N4L thinks.**

You will write a small reading list by hand, load it, and ask it three
questions. The last question's answer is the point of the whole page:
connections you never explicitly wrote become paths the tool walks for you.

This page assumes you have built the binaries and loaded the database per
[Install in 5 minutes](GettingStarted.md). If `./src/bin/N4L` exists and
`./src/bin/searchN4L` runs without a connection error, you are ready.

---

## What we're going to build

A short file in N4L — roughly seven books, a handful of topics they are
about, a few citations between them, and the date I read each one. Small
enough to keep in your head, rich enough to ask real questions of:

- *What have I read about decision making?*
- *What was I reading in spring of 2024?*
- *What connects these two books?*

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

 :: books, topics, reading history ::

 "Thinking Fast and Slow"   (is about) decision making
       "                    (read on)  2024-03-15
```

Three things to notice.

The line starting with `-` names the chapter. Everything else is inside it.
The line starting with `::` is a context — a set of tags that frame the
section. You can ignore both of these for now; they matter when your graph
gets big enough that you want to scope searches.

The things in double quotes — `"Thinking Fast and Slow"` — are nodes.
The things in parentheses — `(is about)`, `(read on)` — are arrows. An
arrow goes from the thing on its left to the thing on its right. So that
second line says: *this book is about decision making.* Third line: *I read
it on 2024-03-15.*

The big quote on the second and third lines (`"`) is a ditto mark. It means
"the node from the line above." You do not have to retype the book's
title.

Your graph now contains one book, one topic, one date, and two arrows
connecting them.

### A second book, and a shared topic

Add a second book that happens to be about the same thing:

```n4l
 "Superforecasting"         (is about) decision making
       "                    (read on)  2024-05-20
```

You did not tell the graph that *Thinking Fast and Slow* and
*Superforecasting* have anything to do with each other. You just said each
of them points at `decision making`. But they do now: two arrows,
same target. A query that starts at one and follows "is about" then
follows "is about" backwards lands at the other.

This is the core of it. You do not model relationships directly. You write
down what each thing points at, and the relationships fall out.

### A citation

Add one more arrow:

```n4l
 "Superforecasting"         (cites)    "Thinking Fast and Slow"
```

`(cites)` is just another arrow. The tool does not know what "cites" means
— you defined it by using it — but it will treat it as a first-class
connection. Your graph now has *two* paths from one book to the other: the
shared-topic path we already had, and the direct citation you just added.

### Dates, and a couple more books

Add three more books so we have something to query:

```n4l
 "Thinking Fast and Slow"   (cites)    "Judgment under Uncertainty"

 "Judgment under Uncertainty" (is about) heuristics
       "                     (read on)  2024-04-02

 "Thinking in Systems"      (is about) decision making
       "                    (read on)  2024-09-14
       "                    (author)   Donella Meadows
```

Dates are just strings — nothing parses them as dates — but you can still
search for them, and you can still read them off an orbit. The
`(author)` arrow is new; you made it up by writing it. N4L accepts new
arrow types as you need them.

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

`-u` means *upload to the database*. You will see a brief summary —
something in the shape of:

```
Uploading nodes..

-------------------------------------
Incidence summary of raw declarations
-------------------------------------
Total nodes ...
Total directed links of type LeadsTo ...
Total directed links of type Contains ...
Total directed links of type Express ...
Total links ... sparseness (fraction of completeness) ...
```

If you see that summary and no red error lines, the corpus is in. If you
see a connection error instead, your database is not running — back to
[Install in 5 minutes](GettingStarted.md).

---

## Asking questions

Three queries. Each one is a sentence in English, a command, and a reading
of the answer.

### What is about decision making?

```
./src/bin/searchN4L "decision making"
```

You will see matches whose text contains `decision making`, each with the
nodes that point at it. Something like:

```
0: decision making
      -    (is about) - Thinking Fast and Slow
      -    (is about) - Superforecasting
      -    (is about) - Thinking in Systems
```

Three books, all reachable because each wrote an `(is about)` arrow at the
same target. You did not build an index. You did not write a join. The
graph answered it because that is how you wrote it down.

### What was I reading around spring of 2024?

```
./src/bin/searchN4L "2024-03"
```

Date strings are just text, so a substring search finds them. The orbit
around a match pulls in whatever points at it:

```
0: 2024-03-15
      -    (read on) - Thinking Fast and Slow
```

Swap `2024-03` for `2024-04` or any other month and you will see what
landed in that window. This is the temporal dimension you get for free:
you wrote the dates in the same notation as everything else, so they are
queryable in the same way.

### What connects Thinking Fast and Slow and Superforecasting?

This is the one.

```
./src/bin/searchN4L "\\from \"Thinking Fast and Slow\" \\to Superforecasting"
```

You are asking for paths from the first book to the second. You get
something in the shape of:

```
     - story path:  Thinking Fast and Slow  -(is cited by)->  Superforecasting

     - story path:  Thinking Fast and Slow  -(is about)-> decision making  -(is about ←)-> Superforecasting
```

Two answers. One direct — *Superforecasting* cites the earlier book, so
there is a one-hop path from one to the other through `cites`. One
through a shared topic — both books point at `decision making`, so
there is a two-hop path through the topic.

You wrote neither of these paths. You wrote that *Superforecasting* cites
*Thinking Fast and Slow*. You wrote that each of them is about decision
making. The tool composed those facts into paths when you asked.

That is the whole idea.

---

## What just happened

- **You wrote it.** Twenty lines of plain text. No schema designed in
  advance. No code. New arrow types as you needed them.
- **You loaded it.** One command. The database is a cache; your N4L file
  is still the source of truth and lives in your editor and your version
  control.
- **You asked it.** Three questions in English-shaped commands. The hard
  one — "what connects these two?" — was answered with connections you
  did not explicitly write.

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

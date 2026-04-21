# Glossary

A pocket dictionary of the words SSTorytime throws at you. Each entry leads
with what you *do* with the concept, then defines it. The aim is narrow:
read the rest of the docs without keeping a second tab open.

Terms are alphabetical.

---

## Ambient context

The scene a note was written inside — chapter, place, mood, time-of-day,
whatever tags stood over the block in your N4L file. You do not usually
write ambient context directly; it comes from the `::` headers and the
chapter line, and gets attached to every link declared underneath. At
query time it is compared against what you are *looking for* — see
[Intentional context](#intentional-context) — and the overlap scores
relevance.

## Arrow

The named relation between two nodes. You write arrows in parentheses
between the left-hand and right-hand node: `Thinking Fast and Slow (about)
decision making`. Arrow *names* (`about`, `by`, `bib-cite`, `note`,
`leads-to` …) are yours to compose, but every arrow has to be
**pre-declared** in the project's `SSTconfig/` files as one of the four
spacetime meta-types — NEAR, LEADSTO, CONTAINS, EXPRESSES — before N4L
will accept it. You cannot invent an arrow mid-file. See
[Thinking in arrows](../arrows.md) for the declared vocabulary and how
to add to it.

## Chapter

A scoping label you put on notes so queries can be filtered to the
section you mean. Declared with a leading `-` at the top of a block:
`- reading list`. Everything below sits inside that chapter until
another `-` line opens a new one. Chapter is not a namespace — the
same node can appear in several chapters — it is a *filter*. Searches
can be restricted with `in chapter: reading list` when you want a
crowded graph to answer for only one corner of itself.

## Cone / cone search

The breadth-first neighbourhood of a node, expanding outward to a
depth you set. The word comes from the spacetime picture: the set of
things reachable from an event, outward in every direction, forms a
cone. In SSTorytime you use a cone search when you want *everything
near* a starting point — all topics, authors, citations, notes
attached to a book and one or two hops beyond — as opposed to a
single directed walk.

## Context

The metadata that answers *under what conditions does this relationship
hold?* Alice trusts Bob on a rope pitch; Alice trusts Bob with the
company credit card; same arrow, different contexts. Every link in
SSTorytime carries a context. Context splits into
[ambient](#ambient-context) (the scene the note was written in) and
[intentional](#intentional-context) (what the reader is looking for);
a query's relevance is the overlap between the two. See
[How does context work?](../howdoescontextwork.md) for worked examples.

## Intentional context

The context a *query* brings — what you are looking for right now,
regardless of what you once wrote. You supply it at search time
(`searchN4L "trust" context rope-pitch`) and it is matched against the
[ambient context](#ambient-context) stored alongside each link. Intent
is the half only you can provide: the text can tell you what was
written down, it cannot tell you why you are reading it today.

## N4L

*Notes For Learning* — the plain text notation you write stories in.
A line says: left-hand node, arrow in parentheses, right-hand node.
Ditto marks (`"`) re-use the previous node. Contexts sit between `::`
markers. Chapters open with `-`. That is most of the language; the
rest is comfort features. See [Writing N4L by hand](../N4L.md) for the
full shape.

## Node

Any text-addressed point in the graph: a single word, a phrase, a
paragraph, a whole document. You do not declare nodes separately —
they come into being when you use them on either side of an arrow.
If you write `Thinking Fast and Slow (about) decision making`, two
nodes now exist. Reuse the same text later and it is the same node.

## Orbit

The immediate neighbourhood of a focal node — the things it touches
closely enough that knowing the node means roughly knowing them too.
When you ask `searchN4L "Thinking Fast and Slow"`, what comes back
is the orbit: topics it is about, who wrote it, what it cites, what
it is cited by, any notes hanging off it, plus one hop further out.
The orbit is how the graph *answers* a query — not with a single
fact but with the neighbourhood of significance around it.

## PageMap

A parallel record of the order you originally wrote things in — the
line sequence of your N4L file. Most queries return the graph; a few
(like "render this chapter as it was written") use the PageMap to
walk the source in order. The graph tells you how things connect;
the PageMap tells you what came *before* what on the page.

## Path

A chain of arrows leading from one node to another. `pathsolve` hands
you paths: the sequence of steps the graph found between start and
end. Paths in `pathsolve` trace **LEADSTO**-type arrows only — the
causal/sequential family — because that is what "came before / came
after" means in practice. If you want to follow NEAR, CONTAINS, or
EXPRESSES chains, reach for a cone search instead. See
[Finding paths](../pathsolve.md).

## Re-uploading (idempotence)

What you do when you edit your N4L and want the graph updated: run
`N4L -u` again. Re-uploading the same file does not duplicate your
graph — nodes are addressed by their text, links by their endpoints
and arrow, so a second ingest sees "already there" and moves on.
The consequence is practical: your N4L file stays the source of
truth, the database is a cache, and rebuilding from the file is
cheap.

## Sequence mode

A shorthand for writing an ordered list without repeating the arrow
on every line. Open a block with `+:: _sequence_ ::` and successive
items get auto-linked with `(then)` — useful for flow charts, recipe
steps, timelines, anywhere the point is *this, then this, then this*.
See [Writing N4L by hand](../N4L.md).

## Story

The walk a reader takes through a corner of the graph — a chapter
read in order, following the thread. SSTorytime treats stories as
the point rather than a presentation wrapper: knowledge that cannot
be told as a story is incomplete, and the structure of the story is
what you remember after individual facts have faded. See the longer
argument in [Storytelling](../Storytelling.md).

## Supernode

A cluster of nodes that turn up together in path results — the graph
noticing "these always appear in the same region, treat them as a
group". You do not write supernodes; they emerge from running path
and cone searches over a graph large enough to have regions. Useful
for spotting the hubs in a corpus you did not design.

## Wave-front (bidirectional search)

![Pen-and-ink physics-style diagram: two points labelled START and END on a horizontal line, concentric ripple-arcs spreading outward from each, meeting in the middle; at the overlap a path of labelled nodes A through E emerges, drawn darker than the ripples.](../figs/wavefront_bidirectional.jpg){ align=center }

The picture Mark uses for path-finding in the Smart Spacetime work:
an influence spreading outward from a point the way a wave spreads
across water, crossing one neighbourhood of significance after
another. To solve a path from A to B, SSTorytime starts a wave at A
and a second, inverse wave at B, and lets them travel until they
meet. Where the fronts overlap is the path. This is cheaper than a
single-direction sweep because each wave only has to reach the
meeting place, not the far shore.

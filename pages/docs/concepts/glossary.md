# Glossary

A pocket dictionary of the words SSTorytime throws at you. The goal of this page
is narrow: define each term just well enough to read the rest of the docs
without needing a second tab open. Every entry points at the line of code where
the concept actually lives, because *that* is the source of truth — prose drifts,
code does not.

Terms are ordered alphabetically.

---

## Ambient context

The background metadata that colours every item written under a `::` block in
N4L — chapter, place, mood, time-of-day, anything the notes were *written
inside*. Ambient context is compared against **intentional context** at query
time to score relevance. Implemented as a flat string merged through
[`NormalizeContextString`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L88).

See also: [Context](#context), [Intentional context](#intentional-context).

## Appointment

A first-class structure that records "this node points at these other nodes by
the same arrow" — useful for render-time grouping, not a separate DB table.
Defined at
[`types_structures.go:108-119`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L108-L119)
and populated by
[`GetAppointedNodesByArrow`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L772).

## Arrow (STtype)

The *relation* between two nodes. SST uses **4 conceptual types** —
`NEAR`, `LEADSTO`, `CONTAINS`, `EXPRESS` — which become **7 signed channels**
(±3 directional plus the symmetric `NEAR`) in storage. Constants at
[`globals.go:23-26`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L23-L26);
channel mapping at
[`STtype.go:82-109`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/STtype.go#L82-L109).
See the full breakdown in [Arrows & Relationships](../arrows.md).

## ArrowDirectory

The in-memory (and on-disk) table that numbers every arrow. Every arrow has a
long form ("leads to"), a short form (`lt`), an `STAindex` (its signed STtype),
and an `ArrPtr`. Defined at
[`types_structures.go:71-77`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L71-L77)
and kept in the global `ARROW_DIRECTORY`
([`globals.go:96`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L96)).

## Chapter

A string label on a node that groups it with other nodes written together —
usually a file, a topic, an essay. Chapter is a scoping filter on every search;
it is *not* a namespace. Field at
[`types_structures.go:39`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L39).

## Cone / cone search

A breadth-first horizon of links expanding outward from a root node to a given
depth. "Cone" because in spacetime terms the graph neighbourhood is a causal
cone around an event. Bounded by `CAUSAL_CONE_MAXLIMIT = 100`
([`globals.go:29`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L29)).
Implemented by
[`GetEntireConePathsAsLinks`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L1089)
and
[`GetConstraintConePathsAsLinks`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L1164).

## Context

The metadata that scopes a link: *under what conditions does this relationship
hold?* Stored deduplicated in `CONTEXT_DIRECTORY` with an integer `ContextPtr`
into it, so every link carries a cheap pointer rather than a string. Context is
split into **ambient** (scene-level) and **intentional** (query-level), each
evaluated through
[`RegisterContext`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L33-L54)
and
[`TryContext`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L58-L69).

## Idempotence (insertion)

The guarantee that uploading the same N4L twice produces the same graph, not a
duplicated graph. Nodes are hashed by text+chapter; links are deduped per
channel. Implemented at
[`IdempDBAddNode`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_insertion.go#L47)
and
[`IdempDBAddLink`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_insertion.go#L97).

## Intentional context

The metadata the *query* brings — what the user is looking for, regardless of
what was written. Compared against [ambient context](#ambient-context) at
retrieval time. Treated symmetrically to ambient by
[`CompileContextString`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L73).

## Link

The concrete row stored inside a `Node`'s link-array column, combining an
`ArrowPtr`, a weight, a `ContextPtr`, and the destination `NodePtr`. Defined at
[`types_structures.go:49-55`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L49-L55).

## N4L

*Notes For Learning* — SSTorytime's plaintext DSL for authoring graphs. Humans
(and LLMs) write N4L; the parser at
[`src/N4L/N4L.go`](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go)
turns it into nodes, arrows, and context declarations and feeds them to
PostgreSQL. See [N4L Reference](../N4L.md).

## Node

A text-addressed point in the graph. Nodes can be tokens, phrases, paragraphs,
or whole documents — length is not fixed. Each carries its own 7-channel link
array. Struct at
[`types_structures.go:33-45`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L33-L45).

## NodePtr

A compact `(Class, CPtr)` address for a node. `Class` is a **size bucket** (see
[`globals.go:48-53`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L48-L53)):

| Class | Range |
|---|---|
| `N1GRAM` = 1 | single-word tokens |
| `N2GRAM` = 2 | two-word phrases |
| `N3GRAM` = 3 | three-word phrases |
| `LT128` = 4 | strings under 128 B |
| `LT1024` = 5 | strings under 1 KB |
| `GT1024` = 6 | strings over 1 KB |

`CPtr` is the index inside the class's lane. Defined at
[`types_structures.go:59-67`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L59-L67).

## Orbit

The immediate neighbourhood of a focal node — the set of nodes within a small
radius, organised by STtype. In Semantic Spacetime terms this is a node's
*neighbourhood of significance*: the handful of things it touches closely
enough that knowing the node means roughly knowing the orbit too. Default
`probe_radius = 3`. Implemented at
[`GetNodeOrbit`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go#L273-L306)
which fans out one goroutine per STtype channel; the result is shaped for
JSON and web rendering (struct at
[`types_structures.go:220-231`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L220-L231)).

## PageMap

A parallel overlay of the graph that preserves **narrative line order** — "this
note came before that one in the original file." Used by the web server to
render the source as written while also offering graph navigation. Struct at
[`types_structures.go:97-104`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L97-L104);
table DDL at
[`postgres_types_functions.go:50-57`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L50-L57).

## Path

The result of a [cone](#cone-cone-search) or [wave-front](#wave-front-bidirectional-search)
search — a sequence of `Link`s from one node to another, optionally
constrained by arrow type, chapter, and context. Paths are what `pathsolve`
returns; supernodes and betweenness fall out of them.

## PoSST

The connection handle: `type PoSST struct { DB *sql.DB }`
([`types_structures.go:17-20`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L17-L20)).
Every API call takes a `PoSST`; acquire one with `Open(load_arrows bool)` and
release with `Close(sst)`.

## Sequence mode

A special N4L context marker — `+:: _sequence_ ::` — that tells the parser to
auto-link successive list items with the `(then)` arrow, saving the author from
typing the relation on every line. Matched at
[`CheckSequenceMode`](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go#L2154-L2172)
and applied by
[`LinkUpStorySequence`](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go#L2176-L2206).

## Story

The path a reader walks through a corner of the graph — a chapter plus an
ordered axis of `NodeEvent`s. Stories are the *point*, not a presentation
wrapper: knowledge that can't be told as a story is incomplete, and the
structure of the story is what you remember after the individual facts have
faded (see [`Storytelling.md`](../Storytelling.md)). SSTorytime prefers to
return a traversal that reads like prose rather than a flat query result; the
struct lives at
[`types_structures.go:185-205`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L185-L205).

## Supernode

A cluster of nodes that emerge as co-located in path results — betweenness
centrality applied to cone search output. Computed by
[`SuperNodes`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/centrality_clustering.go#L113)
and
[`BetweenNessCentrality`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/centrality_clustering.go#L33).

## Wave-front (bidirectional search)

A *propagation* through the graph. The image is the one Mark uses in the
Smart Spacetime work: an influence spreads outward from a point the way a
wave spreads across water, crossing one neighbourhood of significance after
another. To solve a path from A to B, SSTorytime starts a wave at A and a
second, inverse wave at B, and lets them travel until they meet. Where the
fronts overlap is the path. This is cheaper than a single-direction sweep
because each wave only has to reach the meeting place, not the far shore.
Core at
[`GetPathsAndSymmetries`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go#L18-L75)
and
[`WaveFrontsOverlap`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go#L309).

# pathsolve — flag reference

Full flag and internals reference for `pathsolve`. This page is kept in
the repo for contributors. The user-facing page — "Finding paths" —
lives at `../pathsolve.md`.

## Command line

```
pathsolve [-v] -begin <string> -end <string> [-chapter <string>] [-bwd] [subject] [context]
```

Flag declarations live at [`src/pathsolve/pathsolve.go:59-63`](https://github.com/markburgess/SSTorytime/blob/main/src/pathsolve/pathsolve.go#L59-L63).

- `-v` — verbose mode. Prints the boundary condition match sets and internal wave-front state.
- `-begin <string>` — text for the **start** set. Matches by substring via `GetDBNodePtrMatchingName`.
- `-end <string>` — text for the **end** set.
- `-chapter <string>` — optional substring filter. Restricts the start/end node lookups and the path search to nodes tagged with this chapter. Default: empty (search the whole graph).
- `-bwd` — reverse the search direction. Internally swaps the forward/backward wave-front labels (`FWD` and `BWD` in the source). Use this when you want paths from `end` to `begin` along reverse-arrow semantics.

The single positional argument, if present, is checked against `DiracNotation`. If it is a Dirac-form string, the boundary conditions parsed from it override `-begin` / `-end`.

## Arrow-type constraint

`pathsolve` walks only `+leads to` (LT-family) arrows. These are the
forward arrows whose STtype classification puts them in the sequence /
causation family — `(then)`, `(leads to)`, `(next)`, `(=>)`,
`(causes)`, `(prec)`, and their inverses. Arrows of other families
(`+expresses`, `+contains`, `+close to`) are ignored by the path
expansion.

This is enforced by the underlying path-search machinery
(`GetConstraintConePathsAsLinks`), which `pathsolve` invokes with a
fixed LT-family constraint. See
[`src/pathsolve/pathsolve.go`](https://github.com/markburgess/SSTorytime/blob/main/src/pathsolve/pathsolve.go)
for the wiring.

## Hardcoded search depth

`pathsolve` searches paths of length **2 to 20** hops. These are `const` values in the source:

```go
const mindepth = 2
const maxdepth = 20
```

See [`src/pathsolve/pathsolve.go:121-122`](https://github.com/markburgess/SSTorytime/blob/main/src/pathsolve/pathsolve.go#L121-L122).

- `mindepth = 2` — skips the trivial "start and end are the same node" result. If your search is entirely scoped to one node (common when the start/end strings overlap), you need at least 2 hops for the result to be a genuine path.
- `maxdepth = 20` — caps the wave-front expansion. Graphs with very long paths may exceed this; paths longer than 20 hops simply will not be found. Edit the source and rebuild if your use case needs a larger horizon.

## Dirac `<end|start>` notation

```
pathsolve "<end|start>"
pathsolve "<B6|A1>"
pathsolve "<target|start>"
```

The **end set comes first** (the bra `<end|`) and the **start set comes second** (the ket `|start>`). This mirrors quantum-mechanical transition-matrix notation: read `<end|start>` as "amplitude for the system to evolve into `end`, given it starts in `start`."

Parsing is dispatched from `DecodeSearchField` at
[`pkg/SSTorytime/service_search_cmd.go:180`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go#L180)
(the call site of `DiracNotation`); the `DiracNotation` function itself
is defined in
[`pkg/SSTorytime/tools.go:523`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/tools.go#L523).
`pathsolve` wires it into its own argument loop at
[`src/pathsolve/pathsolve.go:102-110`](https://github.com/markburgess/SSTorytime/blob/main/src/pathsolve/pathsolve.go#L102-L110).
The parser also extracts an optional trailing context string.

When Dirac notation is used, the `-begin`/`-end` flags are overridden.

## The `examples/door.n4l` / maze walkthrough

A classic demonstration corpus in the repo uses a maze of named nodes:

```
$ cd examples
$ make
$ pathsolve -begin A1 -end B6

 Paths < end_set= {B6, b6, } | {A1, } = start set>

     - story path: 1 * A1  -(forwards)->  A3  -(forwards)->  A5  -(forwards)->  S1
      -(forwards)->  B1  -(forwards)->  B4  -(forwards)->  B6

    Linkage process: -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)-> .

     - story path: 2 * A1  -(forwards)->  A3  -(forwards)->  A5  -(forwards)->  S2
      -(forwards)->  B2  -(forwards)->  B4  -(forwards)->  B6
     ...
```

and a betweenness / supernode analysis underneath:

```
 * PATH ANALYSIS: into node flow equivalence groups

    - Super node 0 = {A1,}
    - Super node 1 = {A3,A2,}
    - Super node 2 = {A5,A6,}
    - Super node 3 = {S1,}
    ...

 * FLOW IMPORTANCE:

    -Rank (betweenness centrality): 1.00 - B4,A1,B6,
    -Rank (betweenness centrality): 0.80 - A5,
    ...
```

Adjoint path search:

```
$ pathsolve -begin B6 -end A1 -bwd
```

Dirac notation:

```
$ pathsolve "<B6|A1>"
$ pathsolve "<end|start>"
$ pathsolve "<target|start>"
```

## Using in the web browser

In the search field, enter the Dirac notation, e.g. `<target|start>` and relevant chapter `interference`, then click on `geometry`.

![Alpha interface](../figs/pathsolve1.png 'pathsolving in a web interface')
![Alpha interface](../figs/pathsolve2.png 'pathsolving in a web interface')

## Speeding up path searches with arrow constraints

When searching for paths, searches with few constraints are expensive
because the graph branches at every step — the number of possible paths
grows exponentially. One way to reduce this complexity is to specify
the kinds of arrows allowed. If your graph has consistent link types,
this greatly reduces the search space.

The directed nature of arrows makes this complicated: when specifying
arrows, you need to give both the forward and backwards arrows, because
the search is made from start and end simultaneously. The start sees
outgoing forward links and the end sees outgoing backwards links. The
general tool for path searching is the `GetConstraintConePathsAsLinks()`
function, with or without arrows.

```
searchN4L -v \\from \!a1\! \\to b6 \\arrow 20,21
searchN4L -v \\from \!gun\! \\to scarlet \\arrow +3,-3,0
```

Always give pairs of arrow,inverse since the FROM and the TO match
opposite arrow directions. The power of SST becomes more apparent when
using the STtypes `0,1,2,3` for matching arrows by general family rather
than specific name.

## Notes about path searching

When we search for a path, we supply boundary conditions for the start
and end. If we reverse start and end (the adjoint path), the direction
of arrows along the path reverses, but we should find a meaningful
solution in both directions.

Paths need to be bounded by minimum and maximum lengths:

- **Maximum** — we don't know whether there is actually a meaningful
  solution linking start and end, so we give up searching at some point
  (assuming the search doesn't end because we've reached the edge of
  the graph).
- **Minimum** — start and end criteria may contain the same nodes, so a
  single node already satisfies the path criteria. In the default data
  there is a `door.n4l` example of paths from `start` to `target 1`,
  `target 2`, `target 3`, but other chapters contain nodes whose
  strings partially match "start" and "target". A minimum length of 2
  skips the premature match.

Longer (non-trivial) paths can contain arrows of mixed causality — nodes
that go forwards and backwards along arrows. These correspond to
"higher perturbations" in the quantum loop expansion (see
[Searching in Graphs, Artificial Reasoning, and Quantum Loop Corrections with Semantic Spacetime](https://medium.com/@mark-burgess-oslo-mb/searching-in-graphs-artificial-reasoning-and-quantum-loop-corrections-with-semantics-spacetime-ea8df54ba1c5)).

## Exit codes

- **`0`** — at least one path found and printed.
- **`-1`** — no paths satisfy the constraints, or any library/DB error (see [`src/pathsolve/pathsolve.go:164-167`](https://github.com/markburgess/SSTorytime/blob/main/src/pathsolve/pathsolve.go#L164-L167)).
- **`2`** — invalid flag.

## Environment variables

- `POSTGRESQL_URI` — overrides the hardcoded DSN in [`pkg/SSTorytime/session.go:41`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L41).
- `SST_CONFIG_PATH` — location of `SSTconfig/`. Arrows are loaded from the DB via `Open(true)`, so this is usually not needed.

If the database is unreachable, `pathsolve` prints a connection error and exits `-1` before any wave-front work runs.

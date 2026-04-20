
# API walkthroughs

This page works through the four `API_EXAMPLE_*` programs in
[`src/`](https://github.com/markburgess/SSTorytime/tree/main/src). Each
example is self-contained, compiles independently, and exercises a
slightly different corner of the Go library. Read them in order — each one
builds on the primitives introduced in the previous.

If you already have SSTorytime built, you can run any example with

```
cd src/API_EXAMPLE_N
make
./API_EXAMPLE_N
```

The `make` target in each example's directory wraps `go build` and drops
the binary alongside the source.

## API_EXAMPLE_1 — node-by-node insertion and path lookup

**Goal:** demonstrate the minimum viable Go program that (a) writes a
small story into the graph using `Vertex` / `Edge`, and (b) reads the
story back as a forward [cone](concepts/glossary.md#cone-cone-search) of linked nodes.

This is the shape of every custom SSTorytime client. `Open` establishes
the database connection; `Vertex` creates idempotent nodes; `Edge` wires
them together with a named arrow; `GetDBNodePtrMatchingName` and
`GetFwdPathsAsLinks` walk the graph. `Close` releases the connection.

Source:
[`src/API_EXAMPLE_1/API_EXAMPLE_1.go`](https://github.com/markburgess/SSTorytime/blob/main/src/API_EXAMPLE_1/API_EXAMPLE_1.go).

```go
package main

import (
    "fmt"
    SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

func main() {
    load_arrows := false          // arrows are already registered in the DB
    sst := SST.Open(load_arrows)  // opens a persistent PoSST connection

    AddStory(sst)
    LookupStory(sst)

    SST.Close(sst)                // drains the lib/pq pool
}

func AddStory(sst SST.PoSST) {
    chap := "home and away"
    context := []string{""}
    var w float32 = 1.0

    // Vertex() is the high-level "create-or-reuse" primitive.
    // It hashes the text, checks NODE_CACHE, and otherwise
    // calls IdempDBAddNode in pkg/SSTorytime/API.go.
    n1 := SST.Vertex(sst, "Mary had a little lamb", chap)
    n2 := SST.Vertex(sst, "Whose fleece was dull and grey", chap)
    n3 := SST.Vertex(sst, "And every time she washed it clean", chap)
    n4 := SST.Vertex(sst, "It just went to roll in the hay", chap)
    n5 := SST.Vertex(sst, "And when it reached a certain age ", chap)
    n6 := SST.Vertex(sst, "She'd serve it on a tray", chap)

    // Edge() writes an idempotent directed link with an arrow
    // name ("then"), an optional context tag set, and a weight.
    // The arrow must already exist in the ArrowDirectory —
    // user code cannot register new arrows via Edge().
    SST.Edge(sst, n1, "then", n2, context, w)

    // Bifurcation: two possible continuations from n2.
    SST.Edge(sst, n2, "then", n3, context, w/2)
    SST.Edge(sst, n2, "then", n5, context, w/2)

    // Endings.
    SST.Edge(sst, n3, "then", n4, context, w)
    SST.Edge(sst, n5, "then", n6, context, w)
}

func LookupStory(sst SST.PoSST) {
    // Random-access lookup: find all nodes whose text matches "Mary had a".
    start_set := SST.GetDBNodePtrMatchingName(sst, "Mary had a", "")

    // Discover the STtype associated with the "then" arrow.
    _, sttype := SST.GetDBArrowsWithArrowName(sst, "then")

    path_length := 4
    const maxlimit = SST.CAUSAL_CONE_MAXLIMIT // 100, from globals.go

    for n := range start_set {
        // Walk the forward cone of depth path_length from each start node.
        paths, _ := SST.GetFwdPathsAsLinks(sst, start_set[n], sttype, path_length, maxlimit)

        for p := range paths {
            if len(paths[p]) > 1 {
                fmt.Println("    Path", p, " len", len(paths[p]))
                for l := 0; l < len(paths[p]); l++ {
                    name := SST.GetDBNodeByNodePtr(sst, paths[p][l].Dst).S
                    fmt.Println("    ", l, "xx  --> ",
                        paths[p][l].Dst, "=", name, "  , weight",
                        paths[p][l].Wgt, "context", paths[p][l].Ctx)
                }
            }
        }
    }
}
```

**Expected output** (truncated):

```
    Path 0  len 4
     0 xx  -->  {4 0} = Mary had a little lamb   , weight 1 context []
     1 xx  -->  {4 987} = Whose fleece was dull and grey   , weight 1 context []
     2 xx  -->  {4 988} = And every time she washed it clean   , weight 0.5 context []
     3 xx  -->  {4 989} = It just went to roll in the hay   , weight 1 context []
    Path 1  len 4
     0 xx  -->  {4 0} = Mary had a little lamb   , weight 1 context []
     1 xx  -->  {4 987} = Whose fleece was dull and grey   , weight 1 context []
     2 xx  -->  {4 990} = And when it reached a certain age    , weight 0.5 context []
     3 xx  -->  {4 991} = She'd serve it on a tray   , weight 1 context []
```

**Exercises:**

1. Add a fifth continuation from `n2` to a new node `n7`; confirm
   `GetFwdPathsAsLinks` now returns three paths instead of two.
2. Replace `"then"` with an arrow that does not exist in the directory
   and observe what `Edge` does — you will need `load_arrows := true` and
   an `InsertArrowDirectory` call to register a new arrow name before
   `Edge` can use it.
3. Query by context rather than name: pass a non-empty context slice to
   `Edge`, then modify `LookupStory` to filter `paths[p][l].Ctx` against
   the same tag set.

## API_EXAMPLE_2 — hub joins

**Goal:** show how `HubJoin` creates hyperlink-style many-to-one
relationships — a single hub node that a set of child nodes point to
through a shared arrow.

A hub join is the right primitive when a set of items all share a
relationship to a common anchor (a bibliography, a category, a topic).
It respects the semantic-spacetime rules: the hub becomes a first-class
node, and the links are normal `Link` records, not a separate
hyperedge table.

Source:
[`src/API_EXAMPLE_2/API_EXAMPLE_2.go`](https://github.com/markburgess/SSTorytime/blob/main/src/API_EXAMPLE_2/API_EXAMPLE_2.go).

```go
package main

import (
    "fmt"
    SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

func main() {
    load_arrows := false
    sst := SST.Open(load_arrows)

    names := []string{"test_node1", "test_node2", "test_node3"}
    weights := []float32{0.2, 0.4, 1.0}
    context := []string{"some", "context", "tags"}

    var nodes []SST.Node
    var nptrs []SST.NodePtr

    // Create a set of leaf nodes, then harvest their NodePtrs.
    for n := range names {
        nodes = append(nodes, SST.Vertex(sst, names[n], "my chapter"))
        nptrs = append(nptrs, nodes[n].NPtr)
    }

    // HubJoin with empty hub name — the library synthesises
    // "hub_<arrow>_<nodelist>" as the hub's text.
    // Each leaf gets a "then" link to the new hub.
    created1 := SST.HubJoin(sst, "", "", nptrs, "then", context, weights)
    fmt.Println("Creates hub node", created1)

    // HubJoin with an explicit hub name and a different arrow.
    // Passing nil for context and weights uses per-call defaults.
    created2 := SST.HubJoin(sst, "mummy_node", "", nptrs, "belongs to", nil, nil)
    fmt.Println("Creates hub node", created2)

    SST.Close(sst)
}
```

**Expected output:**

```
Creates hub node {4 ...}   // synthesised hub NPtr
Creates hub node {4 ...}   // mummy_node NPtr
```

The specific `Class`/`CPtr` values depend on how many nodes already live
in your database.

**Exercises:**

1. `HubJoin` creates links **leaves → hub** (see
   [`pkg/SSTorytime/API.go:97-106`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go#L97-L106)
   — each leaf becomes the `from` and the hub is `link.Dst`). So a
   forward query from the hub reaches nothing. To confirm the wiring,
   either (a) call `GetFwdPathsAsLinks` from each leaf and verify the
   path lands on `mummy_node`, or (b) use the inverse-arrow orientation
   (`"bwd"` in the search APIs) from `mummy_node` to reach every leaf.
2. Change the third `HubJoin` arrow to one of the 4 canonical types
   (`contains`, `leads to`, `expresses`, `near`) and observe how the
   storage channel (`Im3`..`Ie3`) the links land in changes.
3. Use `HubJoin` to build a small bibliography — one paper per leaf node,
   one citation hub per topic — and retrieve all papers on a topic with
   a single orbit lookup.

## API_EXAMPLE_3 — maze path solve

**Goal:** demonstrate the high-level `GetPathsAndSymmetries` path solver
over a realistic branching graph (a pen-and-paper maze). Shows how to go
from raw insertions to a completed walk between named endpoints.

The maze is encoded as a set of straight-line fragments (`path[0]`..`path[8]`);
each adjacent pair of cells becomes an `Edge`. The solver is then asked
to find `maze_a7` → `maze_i6`.

Source:
[`src/API_EXAMPLE_3/API_EXAMPLE_3.go`](https://github.com/markburgess/SSTorytime/blob/main/src/API_EXAMPLE_3/API_EXAMPLE_3.go).

```go
package main

import (
    "fmt"
    SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

var path [9][]string

func main() {
    // Nine fragments of a maze; path[0] is the main trail from a7 to i6,
    // the others are side passages and dead-ends that the solver must
    // navigate through or around.
    path[0] = []string{"maze_a7", "maze_b7", "maze_b6", "maze_c6", "maze_c5",
        "maze_b5", "maze_b4", "maze_a4", "maze_a3", "maze_b3", "maze_c3",
        "maze_d3", "maze_d2", "maze_e2", "maze_e3", "maze_f3", "maze_f4",
        "maze_e4", "maze_e5", "maze_f5", "maze_f6", "maze_g6", "maze_g5",
        "maze_g4", "maze_h4", "maze_h5", "maze_h6", "maze_i6"}
    path[1] = []string{"maze_d1", "maze_d2"}
    path[2] = []string{"maze_f1", "maze_f2", "maze_e2"}
    path[3] = []string{"maze_f2", "maze_g2", "maze_h2", "maze_h3", "maze_g3", "maze_g2"}
    path[4] = []string{"maze_b1", "maze_c1", "maze_c2", "maze_b2", "maze_b1"}
    path[5] = []string{"maze_b7", "maze_b8", "maze_c8", "maze_c7", "maze_d7",
        "maze_d6", "maze_e6", "maze_e7", "maze_f7", "maze_f8"}
    path[6] = []string{"maze_d7", "maze_d8", "maze_e8", "maze_e7"}
    path[7] = []string{"maze_f7", "maze_g7", "maze_g8", "maze_h8", "maze_h7"}
    path[8] = []string{"maze_a2", "maze_a1"}

    load_arrows := true          // register the "fwd" arrow on startup
    sst := SST.Open(load_arrows)

    // Build the graph: every adjacent pair of cells in every fragment
    // becomes a fwd-linked edge.
    for p := range path {
        for leg := 1; leg < len(path[p]); leg++ {
            chap := "solve maze"
            context := []string{""}
            var w float32 = 1.0

            nfrom := SST.Vertex(sst, path[p][leg-1], chap)
            nto := SST.Vertex(sst, path[p][leg], chap)

            SST.Edge(sst, nfrom, "fwd", nto, context, w)
        }
    }

    Solve(sst)
    SST.Close(sst)
}

func Solve(sst SST.PoSST) {
    const mindepth = 1
    const maxdepth = 16
    var count int
    var arrowptrs []SST.ArrowPtr
    var sttype []int
    var context []string

    start_bc := "maze_a7"
    end_bc := "maze_i6"
    chapter := ""

    // Resolve the endpoint names to NodePtrs.
    leftptrs := SST.GetDBNodePtrMatchingName(sst, start_bc, "")
    rightptrs := SST.GetDBNodePtrMatchingName(sst, end_bc, "")

    if leftptrs == nil || rightptrs == nil {
        fmt.Println("No paths available from end points")
        return
    }

    // High-level path solver — runs bidirectional wave-front search and
    // returns one []Link per solution, already collapsed against symmetries.
    solutions := SST.GetPathsAndSymmetries(sst,
        leftptrs, rightptrs,
        chapter, context, arrowptrs, sttype,
        mindepth, maxdepth)

    if len(solutions) > 0 {
        for s := 0; s < len(solutions); s++ {
            prefix := fmt.Sprintf(" - story %d: ", s)
            // PrintLinkPath is a convenience renderer in the lib.
            SST.PrintLinkPath(sst, solutions, s, prefix, "", nil)
        }
        count++
    }
}
```

**Expected output** (shape):

```
 - story 0:  (fwd) -> maze_a7 (fwd) -> maze_b7 (fwd) -> maze_b6 ... (fwd) -> maze_i6
```

One line per solution; the exact number of solutions depends on how the
solver resolves symmetric sub-paths through the side passages.

**Exercises:**

1. Lower `maxdepth` to 8. Does the solver still find a solution, or is
   the shortest route longer than 8 hops?
2. Add a weighted dead-end (e.g. `maze_x1, maze_x2` with weight `0.01`)
   from `maze_d3`. Does it appear in the solution set? Why/why not?
3. Replace the shared `"fwd"` arrow with two different arrows on
   alternate legs and pass an explicit `sttype` slice to restrict the
   solver to only one of them.

## API_EXAMPLE_4 — wave-front path solver, layered view

**Goal:** open up the `GetPathsAndSymmetries` primitive used in
EXAMPLE_3 and show what happens underneath. This example calls
`GetEntireConePathsAsLinks` and `WaveFrontsOverlap` directly, one
depth-tick at a time, surfacing both DAG solutions and loop-corrected
ones.

!!! note "Prerequisite data"
    This example expects the graph to already contain the `double.n4l`
    notes, which define nodes `A1` through `B6` with cycles. Load those
    first (`N4L -u double.n4l`) or the solver will report no endpoints.

Source:
[`src/API_EXAMPLE_4/API_EXAMPLE_4.go`](https://github.com/markburgess/SSTorytime/blob/main/src/API_EXAMPLE_4/API_EXAMPLE_4.go).

```go
package main

import (
    "fmt"
    SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

func main() {
    load_arrows := true
    sst := SST.Open(load_arrows)

    // Wave-front solver parameters.
    const branching_limit = 2   // max successors explored per node
    const maxdepth = 7          // global depth cap
    var ldepth, rdepth int = 2, 2   // current left/right frontier depths
    var Lnum, Rnum int
    var count int
    var left_paths, right_paths [][]SST.Link

    start_bc := "A1"
    end_bc := "B6"

    leftptrs := SST.GetDBNodePtrMatchingName(sst, start_bc, "")
    rightptrs := SST.GetDBNodePtrMatchingName(sst, end_bc, "")

    if leftptrs == nil || rightptrs == nil {
        fmt.Println("No paths available from end points")
        return
    }

    // Alternating-depth expansion: expand left, then right, then left...
    // WaveFrontsOverlap is the library's path-integral-style collision test.
    for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

        // Forward cone from the start, current depth.
        left_paths, Lnum = SST.GetEntireConePathsAsLinks(sst, "any", leftptrs[0], ldepth, branching_limit)

        // Forward cone from the end node — the wave-front collision
        // logic (WaveFrontsOverlap, below) treats this as the reverse
        // frontier when intersecting the two cones. Note both calls use
        // "any" orientation; no sign flip happens at the call site.
        right_paths, Rnum = SST.GetEntireConePathsAsLinks(sst, "any", rightptrs[0], rdepth, branching_limit)

        // Intersect the two wave fronts. Returns DAG solutions and
        // separately "loop_corrections" — paths that contain a cycle.
        solutions, loop_corrections := SST.WaveFrontsOverlap(sst,
            left_paths, right_paths, Lnum, Rnum, ldepth, rdepth)

        if len(solutions) > 0 {
            fmt.Println("-- T R E E ----------------------------------")
            fmt.Println("Path solution", count, "from", start_bc, "to", end_bc,
                "with lengths", ldepth, -rdepth)
            for s := 0; s < len(solutions); s++ {
                prefix := fmt.Sprintf(" - story %d: ", s)
                SST.PrintLinkPath(sst, solutions, s, prefix, "", nil)
            }
            count++
            fmt.Println("-------------------------------------------")
        }

        if len(loop_corrections) > 0 {
            fmt.Println("++ L O O P S +++++++++++++++++++++++++++++++")
            for s := 0; s < len(loop_corrections); s++ {
                prefix := fmt.Sprintf(" - story %d: ", s)
                SST.PrintLinkPath(sst, loop_corrections, s, prefix, "", nil)
            }
            count++
            fmt.Println("+++++++++++++++++++++++++++++++++++++++++++")
        }

        // Alternate which frontier expands next.
        if turn%2 == 0 {
            ldepth++
        } else {
            rdepth++
        }
    }
}
```

**Expected output** (shape):

```
-- T R E E ----------------------------------
Path solution 0 from A1 to B6 with lengths 3 -2
 - story 0:  (fwd) -> A1 (fwd) -> ... -> B6
-------------------------------------------
++ L O O P S +++++++++++++++++++++++++++++++
 - story 1:  (fwd) -> A1 ... (back edge) ... -> B6
+++++++++++++++++++++++++++++++++++++++++++
```

You will see several TREE blocks as the solver widens the depth, and
intermittent LOOPS blocks whenever cycles appear.

**Exercises:**

1. Raise `branching_limit` to 4 and watch the solution count change.
   Where is the sweet spot between recall and runtime on your graph?
2. Remove the alternating-depth trick (always expand left) and compare
   how many iterations it takes to find the first TREE solution.
3. Instrument `WaveFrontsOverlap` with print statements to visualise
   exactly when the two frontiers collide.

## See also

- [Go API reference](API.md) — catalogue of library functions used above.
- [Web API reference](WebAPI.md) — the HTTPS equivalent of these calls.
- [`src/demo_pocs/`](https://github.com/markburgess/SSTorytime/tree/main/src/demo_pocs) —
  many more Go programs exercising the library.

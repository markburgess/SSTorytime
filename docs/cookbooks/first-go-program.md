# Build your first SSTorytime program

This cookbook walks through writing a small Go program that imports the SSTorytime library, creates a handful of nodes, links them, and reads a path back out. Model: `src/API_EXAMPLE_1/API_EXAMPLE_1.go`. At the end you will have a complete, compilable program.

!!! tip "Keep the glossary open"
    We'll use *arrow*, *orbit*, *chapter*, *context*, *NodePtr*, *PoSST* without re-defining them here.  
    See the [Glossary](../concepts/glossary.md) for quick reference.

!!! info "Prerequisites"
    - Go 1.24 or newer in `$PATH`.
    - PostgreSQL running with the SSTorytime schema loaded (`make db` from the repo root).
    - The SSTorytime [arrow](../concepts/glossary.md#arrow-sttype) directory populated — run `make` at the repo root once to ensure the default arrow set is uploaded.
    - The SSTorytime repository cloned at a known path. We'll use `/home/you/SSTorytime` as a placeholder.
    - This cookbook does **not** require the GOPATH symlink mentioned in older docs — just the `replace` directive below. A modern Go-modules checkout is all you need.

## 1. Set up a module

Create a new directory alongside the repo (not inside it):

```bash
mkdir -p ~/sst-hello && cd ~/sst-hello
go mod init example.com/sst-hello
```

Then tell Go where to find the SSTorytime package by adding a `replace` directive to `go.mod`:

```
module example.com/sst-hello

go 1.24

require github.com/markburgess/SSTorytime v0.0.0

replace github.com/markburgess/SSTorytime => /home/you/SSTorytime
```

The replace directive points at your local checkout. In a production app you would publish a tagged release and drop the replace, but the library is not yet on the public Go proxy.

Pull in the transitive dependency (`lib/pq`):

```bash
go get github.com/lib/pq
```

## 2. Write the program

Save the following as `main.go`:

```go
package main

import (
    "fmt"

    SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

func main() {
    // Open the database. load_arrows=false is fine if the default
    // arrow directory was already loaded by N4L.
    load_arrows := false
    sst := SST.Open(load_arrows)
    defer SST.Close(sst)

    chapter := "first program"
    context := []string{"demo", "cookbook"}
    var weight float32 = 1.0

    // 1. Create five nodes.
    n1 := SST.Vertex(sst, "The ship leaves port", chapter)
    n2 := SST.Vertex(sst, "Storms are forecast", chapter)
    n3 := SST.Vertex(sst, "The crew secures the deck", chapter)
    n4 := SST.Vertex(sst, "A squall hits at midnight", chapter)
    n5 := SST.Vertex(sst, "The captain orders shelter", chapter)

    // 2. Link them with causal arrows. "then" is one of the reserved
    //    leads-to arrows the library ships with.
    SST.Edge(sst, n1, "then", n2, context, weight)
    SST.Edge(sst, n2, "then", n3, context, weight)
    SST.Edge(sst, n3, "then", n4, context, weight)
    SST.Edge(sst, n4, "then", n5, context, weight)

    // Add a "note" property arrow so we can see a non-causal edge
    // in the resulting orbit.
    SST.Edge(sst, n4, "note", n2, context, weight)

    // 3. Query: find forward paths starting from n1, along "then"
    //    arrows, up to 5 hops deep.
    _, sttype := SST.GetDBArrowsWithArrowName(sst, "then")
    pathLen := 5
    const maxResults = SST.CAUSAL_CONE_MAXLIMIT

    startSet := SST.GetDBNodePtrMatchingName(sst, "The ship leaves", chapter)
    if len(startSet) == 0 {
        fmt.Println("no starting node matched; did the inserts succeed?")
        return
    }

    for i := range startSet {
        paths, _ := SST.GetFwdPathsAsLinks(sst, startSet[i], sttype, pathLen, maxResults)
        for p := range paths {
            if len(paths[p]) <= 1 {
                continue
            }
            fmt.Printf("Path %d (len=%d):\n", p, len(paths[p]))
            for hop, link := range paths[p] {
                node := SST.GetDBNodeByNodePtr(sst, link.Dst)
                fmt.Printf("   %d. %s  (weight=%.2f)\n", hop, node.S, link.Wgt)
            }
        }
    }
}
```

## 3. Run it

```bash
go run ./main.go
```

Expected output (tabs may differ):

```
Path 0 (len=5):
   0. The ship leaves port  (weight=1.00)
   1. Storms are forecast  (weight=1.00)
   2. The crew secures the deck  (weight=1.00)
   3. A squall hits at midnight  (weight=1.00)
   4. The captain orders shelter  (weight=1.00)
```

## 4. Understanding what happened

Each call hit a specific layer of the library:

| Call | What it does | Code reference |
|---|---|---|
| `SST.Open(false)` | Connects to PostgreSQL via `lib/pq`. Returns a [`PoSST`](../concepts/glossary.md#posst) session handle. | [`session.go`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go) |
| `SST.Vertex(sst, name, chapter)` | Idempotent node insert. Creates the node if missing; reuses the existing [`NodePtr`](../concepts/glossary.md#nodeptr) if the `(name, [chapter](../concepts/glossary.md#chapter))` pair is already in the graph. | [`API.go:18-28`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go#L18-L28) |
| `SST.Edge(sst, from, arrow, to, context, weight)` | Idempotent link insert; looks up the arrow by its short name, translates to a channel, and writes both the forward and inverse links. | [`API.go:32-46`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go#L32-L46) |
| `SST.GetDBArrowsWithArrowName(sst, "then")` | Resolves the arrow's pointer and STtype. | [`postgres_retrieval.go`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) |
| `SST.GetDBNodePtrMatchingName(sst, prefix, chapter)` | Substring match on node text within a chapter. | [`postgres_retrieval.go`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) |
| `SST.GetFwdPathsAsLinks(sst, start, sttype, depth, max)` | Forward wave-front cone search. Returns every path of up to `depth` hops that follows arrows of the given STtype. | [`path_wave_search.go`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go) |
| `SST.Close(sst)` | Tears down the DB connection. | [`session.go`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go) |

## 5. Things worth knowing

- **Idempotency.** Running the program twice does not double-insert. The second invocation's `Vertex` calls return the same `NodePtr`s and the `Edge` calls no-op.
- **Arrow vocabulary.** `Vertex/Edge` callers cannot invent new arrows on the fly. Use one of the short names from the loaded arrow directory (list them with the command below), or define new arrows in an N4L file uploaded with `N4L -u`.

    ```bash
    searchN4L '\arrow' any
    ```

    The single leading backslash is a `searchN4L` query sigil — quote the argument so bash does not try to interpret it or expand an alias.
- **[Context](../concepts/glossary.md#context) is cheap.** Pass any `[]string` of tags as context; SSTorytime interns the set in the `ContextDirectory` table and keys links by the interned pointer.
- **Cleanup.** To wipe the chapter between runs, call `removeN4L -force "first program"` (assuming `src/bin/` is on your `$PATH` — see [Getting Started](../GettingStarted.md)).

## Next steps

- For Python, see [Python integration](python-integration.md).
- The four bundled examples (`src/API_EXAMPLE_1` through `src/API_EXAMPLE_4`) demonstrate hub joins, multi-context queries, and path constraints — each is under 100 lines.
- To expose your program's data to an LLM, continue with [Connecting an LLM via MCP-SST](llm-via-mcp-sst.md).

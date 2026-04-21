# Concept → Code index

This index is a reverse lookup. Use it when a concept in the docs does not
make sense and you want to read the code.

Every row points at the most load-bearing definition or entry point for a
concept. Functions are listed in the order a reader typically encounters them
— public surface first, helpers second.

!!! warning "Link drift"
    All links target `main`. The code moves; line numbers occasionally drift
    between a release and this page's next refresh. If a link opens in the
    wrong place, use the *file* link and re-find the function by name —
    function names are stable, line numbers are not.

## Library core — `pkg/SSTorytime/`

### Public API surface

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Graph construction (vertex) | [API.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go) | [18-28](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go#L18-L28) | `Vertex` | Auto-numbered `NodePtr`. |
| Graph construction (edge) | [API.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go) | [32-46](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go#L32-L46) | `Edge` | Takes named arrow + context. |
| Hub join (multi-node) | [API.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go) | [50-109](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/API.go#L50-L109) | `HubJoin` | Hyperedge-style container. |
| Session open | [session.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go) | [20-69](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L20-L69) | `Open` | Reads `POSTGRESQL_URI` env var. |
| Credentials file | [session.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go) | [73-136](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L73-L136) | `OverrideCredentials` | `~/.SSTorytime`. |
| Memory init | [session.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go) | [159-183](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L159-L183) | `MemoryInit` | Allocates ARROW/CONTEXT/NODE dirs. |
| Schema bootstrap | [session.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go) | [185-305](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L185-L305) | `Configure` | Creates types, tables, functions. |
| Session close | [session.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go) | [307-318](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L307-L318) | `Close` | Final DB flush + disconnect. |

### Insertion & idempotence

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Idempotent node add | [db_insertion.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_insertion.go) | [47-95](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_insertion.go#L47-L95) | `IdempDBAddNode` | Text+chap dedupe. |
| Idempotent link add | [db_insertion.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_insertion.go) | [97-135](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_insertion.go#L97-L135) | `IdempDBAddLink` | Stores in correct channel col. |
| Link append (in-memory) | [N4L_parsing.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/N4L_parsing.go) | [275-304](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/N4L_parsing.go#L275-L304) | `AppendLinkToNode` | Parser-side accumulator. |
| Bulk upload | [db_upload.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go) | [19-128](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go#L19-L128) | `GraphToDB` | Walks memory dirs → DB. |
| Node upload | [db_upload.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go) | [129-159](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go#L129-L159) | `UploadNodeToDB` | One SQL per node. |
| Context upload | [db_upload.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go) | [218-245](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go#L218-L245) | `UploadContextToDB` | Dedup via CONTEXT_DIR. |
| PageMap upload | [db_upload.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go) | [247-280](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/db_upload.go#L247-L280) | `UploadPageMapEvent` | One row per source line. |

### Retrieval & search

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Name → NodePtr resolution | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [21-63](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L21-L63) | `SolveNodePtrs` | Entry point for search. |
| Multi-criteria match | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [74-107](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L74-L107) | `GetDBNodePtrMatchingNCCS` | Name + chap + ctx + arrow + seq. |
| Node fetch by ptr | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [327-386](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L327-L386) | `GetDBNodeByNodePtr` | Single-row → `Node`. |
| Forward cone (nodes) | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [1003-1031](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L1003-L1031) | `GetFwdConeAsNodes` | BFS to depth `d`. |
| Forward cone (links) | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [1032-1061](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L1032-L1061) | `GetFwdConeAsLinks` | Same, keeping edges. |
| Forward path expansion | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [1062-1088](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L1062-L1088) | `GetFwdPathsAsLinks` | All simple paths. |
| Entire-cone paths | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [1089-1124](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L1089-L1124) | `GetEntireConePathsAsLinks` | Multi-root variant. |
| Constrained cone paths | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [1164+](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L1164) | `GetConstraintConePathsAsLinks` | Arrow + STtype filters. |
| Bidirectional path solve | [path_wave_search.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go) | [18-75](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go#L18-L75) | `GetPathsAndSymmetries` | Wave-front core. |
| Wave-front overlap test | [path_wave_search.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go) | [309-347](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go#L309-L347) | `WaveFrontsOverlap` | L ∩ R each turn. |
| Adjoint arrows/sttypes | [path_wave_search.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go) | [241-271](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/path_wave_search.go#L241-L271) | `AdjointArrows`, `AdjointSTtype` | For inverse frontier. |
| Story selection | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [502-595](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L502-L595) | `SelectStoriesByArrow`, `GetSequenceContainers` | Narrative axes. |
| PageMap fetch | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [949-1002](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L949-L1002) | `GetDBPageMap` | Paginated narrative. |
| Appointed nodes | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [772-872](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L772-L872) | `GetAppointedNodesByArrow`, `GetAppointedNodesBySTType` | Cluster-by-relation. |

### Context layer

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Context registration | [eval_context.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go) | [33-54](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L33-L54) | `RegisterContext` | Populates `CONTEXT_DIRECTORY`. |
| Context DB upload | [eval_context.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go) | [58-69](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L58-L69) | `TryContext` | Idempotent. |
| Context normalize | [eval_context.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go) | [88-120](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L88-L120) | `NormalizeContextString` | Ambient + intentional merge. |
| Node context lookup | [eval_context.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go) | [122-160](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L122-L160) | `GetNodeContext`, `GetNodeContextString` | For rendering. |
| Time-of-day context | [eval_context.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go) | [254-310](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/eval_context.go#L254-L310) | `GetTimeContext`, `Season`, `GetTimeFromSemantics` | Automatic ambient. |

### Arrows & STtype encoding

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| STtype constants | [globals.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go) | [23-34](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L23-L34) | `NEAR`, `LEADSTO`, `CONTAINS`, `EXPRESS`, `ST_ZERO` | 4 types + signed. |
| Channel column names | [globals.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go) | [38-44](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L38-L44) | `I_MEXPR` .. `I_PEXPR` | 7 DB channels. |
| STtype → channel | [STtype.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/STtype.go) | [82-109](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/STtype.go#L82-L109) | `STTypeDBChannel` | Maps signed int → col name. |
| Index ↔ type | [STtype.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/STtype.go) | [113-127](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/STtype.go#L113-L127) | `STIndexToSTType`, `STTypeToSTIndex` | `ST_ZERO` shift. |
| STtype name → index | [STtype.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/STtype.go) | [20-46](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/STtype.go#L20-L46) | `GetSTIndexByName` | Parser-side. |
| Arrow directory download | [cache.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/cache.go) | [92-164](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/cache.go#L92-L164) | `DownloadArrowsFromDB` | Loads on `Open(true)`. |
| Arrow by name | [postgres_retrieval.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go) | [596-690](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_retrieval.go#L596-L690) | `GetDBArrowsWithArrowName`, `GetDBArrowByName` | Long + short resolution. |
| Arrow directory insert | [N4L_parsing.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/N4L_parsing.go) | [223-273](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/N4L_parsing.go#L223-L273) | `InsertArrowDirectory`, `InsertInverseArrowDirectory` | Parser-side. |

### Cache & directories

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| NodePtr cache | [cache.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/cache.go) | [25-90](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/cache.go#L25-L90) | `GetNodeTxtFromPtr`, `GetMemoryNodeFromPtr`, `CacheNode` | `NODE_CACHE` global. |
| Context download | [cache.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/cache.go) | [165-205](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/cache.go#L165-L205) | `DownloadContextsFromDB` | `Open(true)` fills `CONTEXT_DIRECTORY`. |
| NPtr synchronisation | [cache.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/cache.go) | [206-260](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/cache.go#L206-L260) | `SynchronizeNPtrs` | Post-upload reconcile. |
| Size-class bucketing | [globals.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go) | [48-53](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/globals.go#L48-L53) | `N1GRAM`..`GT1024` | Six lanes. |

### JSON & web-render

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Orbit retrieval (concurrent) | [json_marshalling.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go) | [273-306](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go#L273-L306) | `GetNodeOrbit` | One goroutine per STtype. |
| Satellite assembly | [json_marshalling.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go) | [306-377](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go#L306-L377) | `AssembleSatellitesBySTtype` | Per-channel radial sweep. |
| NodeEvent JSON | [json_marshalling.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go) | [21-37](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go#L21-L37) | `JSONNodeEvent` | For story axes. |
| Web-cone paths | [json_marshalling.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go) | [38-107](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go#L38-L107) | `LinkWebPaths` | Paints paths for UI. |
| Page JSON | [json_marshalling.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go) | [188-268](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go#L188-L268) | `JSONPage` | PageMap → web JSON. |
| Axial path | [json_marshalling.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go) | [394-422](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/json_marshalling.go#L394-L422) | `GetLongestAxialPath` | Story spine. |

### Analytics

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Betweenness centrality | [centrality_clustering.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/centrality_clustering.go) | [33-75](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/centrality_clustering.go#L33-L75) | `BetweenNessCentrality` | Over path solutions. |
| Supernode clustering | [centrality_clustering.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/centrality_clustering.go) | [76-160](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/centrality_clustering.go#L76-L160) | `SuperNodes`, `SuperNodesByConicPath` | Clusters on paths. |
| Adjacency matrices | [matrices.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/matrices.go) | [18-156](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/matrices.go#L18-L156) | `GetDBAdjacentNodePtrBySTType` | Dense adj by STtype. |
| Eigenvector centrality | [matrices.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/matrices.go) | [319-395](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/matrices.go#L319-L395) | `ComputeEVC`, `NormalizeVec`, `CompareVec` | Power iteration. |

### Access tracking

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Section-level access | [lastseen.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/lastseen.go) | [18-32](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/lastseen.go#L18-L32) | `UpdateLastSawSection` | 60s sampling threshold. |
| Node-level access | [lastseen.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/lastseen.go) | [26-32](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/lastseen.go#L26-L32) | `UpdateLastSawNPtr` | Stamps `LastSeen` row. |
| Newly-seen filter | [lastseen.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/lastseen.go) | [102+](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/lastseen.go#L102) | `GetNewlySeenNPtrs` | Forgetting curve. |

### Types & structures

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| `Node` struct | [types_structures.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go) | [33-45](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L33-L45) | — | 7-channel `I[ST_TOP][]Link`. |
| `Link` struct | [types_structures.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go) | [49-55](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L49-L55) | — | Arr · Wgt · Ctx · Dst. |
| `NodePtr` struct | [types_structures.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go) | [59-67](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L59-L67) | — | `(Class, CPtr)`. |
| `Story` struct | [types_structures.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go) | [185-205](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L185-L205) | — | Chapter + `[]NodeEvent`. |
| `PageMap` struct | [types_structures.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go) | [97-104](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L97-L104) | — | Narrative overlay. |
| `Orbit` struct | [types_structures.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go) | [220-231](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L220-L231) | — | Web render. |
| `Appointment` struct | [types_structures.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go) | [108-119](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L108-L119) | — | Node-cluster-by-arrow. |
| `LastSeen` struct | [types_structures.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go) | [254-263](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/types_structures.go#L254-L263) | — | Section/node access. |

### Database schema

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| `NodePtr` composite type | [postgres_types_functions.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go) | [18-22](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L18-L22) | `NODEPTR_TYPE` const | Chan + CPtr. |
| `Link` composite type | [postgres_types_functions.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go) | [24-30](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L24-L30) | `LINK_TYPE` const | Arr · Wgt · Ctx · Dst. |
| `Node` table DDL | [postgres_types_functions.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go) | [32-48](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L32-L48) | `NODE_TABLE` const | 7 channel cols + tsvectors. |
| `PageMap` table DDL | [postgres_types_functions.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go) | [50-57](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L50-L57) | `PAGEMAP_TABLE` const | Narrative overlay. |
| `ArrowDirectory` DDL | [postgres_types_functions.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go) | [59-65](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L59-L65) | `ARROW_DIRECTORY_TABLE` | Arrow registry. |
| Stored function install | [postgres_types_functions.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go) | [143+](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/postgres_types_functions.go#L143) | `DefineStoredFunctions` | 35 PL/pgSQL funcs. |

## Parser — `src/N4L/`

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Main entry | [src/N4L/N4L.go](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go) | [190-224](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go#L190-L224) | `main` | Iterates args, calls `Upload`. |
| Flag parse | [src/N4L/N4L.go](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go) | [228-280](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go#L228-L280) | `Init` | `-v`, `-d`, `-u`, `-s`, `-adj`, `-force`, `-wipe`. |
| Sequence-mode switch | [src/N4L/N4L.go](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go) | [2154-2172](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go#L2154-L2172) | `CheckSequenceMode` | `_sequence_` marker. |
| Story auto-link | [src/N4L/N4L.go](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go) | [2176-2206](https://github.com/markburgess/SSTorytime/blob/main/src/N4L/N4L.go#L2176-L2206) | `LinkUpStorySequence` | Applies `(then)`. |

## HTTPS server — `src/server/`

| Concept | File | Lines | Primary functions | Notes |
|---|---|---|---|---|
| Server init | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [57-86](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L57-L86) | `Init` | Flag parse + resources path. |
| Server start | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [87-178](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L87-L178) | `Start` | HTTPS :8443, :8080 redirect. |
| CORS middleware | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [179-205](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L179-L205) | `EnableCORS` | Origin reflection. |
| Search endpoint | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [206-241](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L206-L241) | `SearchN4LHandler` | `/searchN4L`. |
| Upload endpoint | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [242-413](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L242-L413) | `UploadHandler`, `UploadURI`, `UploadInline` | `/Upload` (multipart + URI). |
| Asset lookup | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [414-431](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L414-L431) | `AssetsHandler` | `/Assets/*`. |
| Orbit response | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [622-661](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L622-L661) | `HandleOrbit` | Wraps `GetNodeOrbit`. |
| Causal-cone response | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [662-729](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L662-L729) | `HandleCausalCones`, `PackageConeFromOrigin` | Wraps cone search. |
| Path-solve response | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [730-777](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L730-L777) | `HandlePathSolve` | Wraps `GetPathsAndSymmetries`. |
| PageMap response | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [778-835](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L778-L835) | `HandlePageMap`, `FilterSeen` | Wraps `GetDBPageMap`. |
| Stories response | [http_server.go](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go) | [836-868](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L836-L868) | `HandleStories` | Narrative axes JSON. |

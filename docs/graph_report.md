
# graph_report tool

There's a few things one would like to know about pretty much every graph.
the `graph_report` tool offers a simple analysis for determining these basic
features, for each link meta-type, chapter by chapter etc. These are particularly
useful when we don't know much about the graph to begin with, because it was
collected from a data set rather than as a set of personal notes. The `graph_report`
tool helps us to get a technical overview of the graph.

* *Loops*: graphs that contain loops (cyclic graphs)

* *Sources* and *Sinks*: these are nodes that start and end a path through the graph.
They exchange places if one changes the sign of the link type.

* *Appointed nodes*: when several nodes point to a single hub that appointee is called an appointed
agent in Promise Theory. The cluster of nodes all pointing / electing a single individual are
thus correlated by the appointee (they have it in common). Such structures help us to see
processes and process histories.

* * For "leads to" arrows, these structures are confluences of arrows or explosions from a point.
* * For "contains" arrows, these structures are the containers or shared members
* * For "property expression" arrows, these structures are compositions of attributes or shared attributes common to several compositions
* * For "near" arrows, these structures are synonym / alias / or density clusters

* *Eigenvector centrality*: undirected graphs have a property by virtue of the 
Frobenious-Perron theorem that every undirected graph has a non-negative principal
eigenvector. It ranks the 'connectedness' of nodes, or their importance, by measuring
the amount of 'weight' propagated to each node. We can even calculate this for a directed
graph by symmetrizing all the links, and this will then tell us something about relative
utilization of the nodes for transport and connectivity. It can be compared to the
betweenness centrality scores for directed paths, which are reported by `pathsolve`.
If we think of each node as being a reservoir of 'weight' and each directed arrow as being a gradient, then
all the weight in a directed graph flows to the sinks immediately, leaving all others empty.
In a symmetrized (undirected graph), the flows reach equilibrium and the highest levels settle
in the nodes that are best connected. The capacitance of these nodes tells us something about
their topological connectivity, setting aside the directedness of the arrows.  The graph_report tool
reports the unbiased vector normalization of the principal eigenvector when removing all arrow
directions but preserving their weights.

For example, the report on the "demo_pocs/search_maze" example graph, for leadsto links:
<pre>
go run graph_report.go  -chapter multi|more
----------------------------------------------------------------
Analysing chapter "multi slit interference", context [] to path length 6
----------------------------------------------------------------

* TOTAL NODES IN THE SEARCH REGION 13

* TOTAL DIRECTED LINKS = 17 of possible 156 = 0.11 %

* DISTRIBUTION OF NAME TYPE/LENGTHS:
  - single word ngram : 10 / 13
  - two word ngram : 3 / 13

* PROCESS ORIGINS / ROOT DEPENDENCIES / PATH SOURCES for ("+leads to") in multi slit interference

   - NPtr(1,0) -> start

* FINAL END-STATES / PATH SINK NODES for ("+leads to") in multi slit interference

   - NPtr(2,2) -> target 3
   - NPtr(2,0) -> target 1
   - NPtr(2,1) -> target 2

* DIRECTED LOOPS AND CYCLES:

   - Acyclic

* APPOINTED NODES (nodes pointed to by at least 2 others thus correlating them) 

   Appointer correlates -> 2 appointed nodes (gate ...) in chapter "multi slit interference"

     tram --(comes from / arriving from : -comes from)--> gate...   - in context []
     bike --(comes from / arriving from : -comes from)--> gate...   - in context []

   Appointer correlates -> 4 appointed nodes (start ...) in chapter "multi slit interference"

     door --(comes from / arriving from : -comes from)--> start...   - in context []
     port --(comes from / arriving from : -comes from)--> start...   - in context []
     hole --(comes from / arriving from : -comes from)--> start...   - in context []
     gate --(comes from / arriving from : -comes from)--> start...   - in context []

   Appointer correlates -> 3 appointed nodes (door ...) in chapter "multi slit interference"

     passage --(comes from / arriving from : -comes from)--> door...   - in context []
     road --(comes from / arriving from : -comes from)--> door...   - in context []
     river --(comes from / arriving from : -comes from)--> door...   - in context []

   Appointer correlates -> 2 appointed nodes (port ...) in chapter "multi slit interference"

     river --(comes from / arriving from : -comes from)--> port...   - in context []
     tram --(comes from / arriving from : -comes from)--> port...   - in context []

* SYMMETRIZED EIGENVECTOR CENTRALITY = FLOW RESERVOIR CAPACITANCE AT EQUILIBRIUM = 

   ( 0.993 ) <- 0 = tram
   ( 0.768 ) <- 1 = target 3
   ( 0.847 ) <- 2 = gate
   ( 0.496 ) <- 3 = bike
   ( 1.000 ) <- 4 = start
   ( 0.787 ) <- 5 = door
   ( 0.940 ) <- 6 = port
   ( 0.678 ) <- 7 = hole
   ( 0.767 ) <- 8 = river
   ( 0.272 ) <- 9 = passage
   ( 0.093 ) <- 10 = target 1
   ( 0.272 ) <- 11 = road
   ( 0.093 ) <- 12 = target 2

* THERE ARE 2 LOCAL MAXIMA IN THE EQUILIBRIUM EVC LANDSCAPE:

  - subregion of maximum 0 consisting of nodes [0 1]
     - where 0 -> tram
     - where 1 -> target 3
  - subregion of maximum 4 consisting of nodes [2 3 4 5 6 7 8 9 10 11 12]
     - where 2 -> gate
     - where 3 -> bike
     - where 4 -> start
     - where 5 -> door
     - where 6 -> port
     - where 7 -> hole
     - where 8 -> river
     - where 9 -> passage
     - where 10 -> target 1
     - where 11 -> road
     - where 12 -> target 2

* HILL-CLIMBING EVC-LAMDSCAPE GRADIENT PATHS:

     - Path node 0 has local maximum at node * 0 *, hop distance 0 along [0]
     - Path node 1 has local maximum at node * 0 *, hop distance 1 along [1 0]
     - Path node 2 has local maximum at node * 4 *, hop distance 1 along [2 4]
     - Path node 3 has local maximum at node * 4 *, hop distance 2 along [3 2 4]
     - Path node 4 has local maximum at node * 4 *, hop distance 0 along [4]
     - Path node 5 has local maximum at node * 4 *, hop distance 1 along [5 4]
     - Path node 6 has local maximum at node * 4 *, hop distance 1 along [6 4]
     - Path node 7 has local maximum at node * 4 *, hop distance 1 along [7 4]
     - Path node 8 has local maximum at node * 4 *, hop distance 2 along [8 6 4]
     - Path node 9 has local maximum at node * 4 *, hop distance 2 along [9 5 4]
     - Path node 10 has local maximum at node * 4 *, hop distance 3 along [10 9 5 4]
     - Path node 11 has local maximum at node * 4 *, hop distance 2 along [11 5 4]
     - Path node 12 has local maximum at node * 4 *, hop distance 3 along [12 11 5 4]
</pre>

Another example:
<pre>
$ go run graph_report.go -chapter maze -sttype L
----------------------------------------------------------------
Analysing chapter "maze", context [] to path length 6
----------------------------------------------------------------

* TOTAL NODES IN THE SEARCH REGION 54

* TOTAL DIRECTED LINKS = 55 of possible 2862 = 0.02 %

* DISTRIBUTION OF NAME TYPE/LENGTHS:
  - single word ngram : 54 / 54


* PROCESS ORIGINS / ROOT DEPENDENCIES / PATH SOURCES for ("+leads to") in maze

   - NPtr(1,3107) -> a7
   - NPtr(1,3135) -> d1
   - NPtr(1,3136) -> f1

* FINAL END-STATES / PATH SINK NODES for ("+leads to") in maze

   - NPtr(1,3154) -> f8
   - NPtr(1,3134) -> i6
   - NPtr(1,3160) -> h7

* DIRECTED LOOPS AND CYCLES:

  - Cycle of length 4 with members (33)(34)(35)(36)
  - Cycle of length 4 with members (37)(38)(39)(40)

* SYMMETRIZED EIGENVECTOR CENTRALITY = FLOW RESERVOIR CAPACITANCE AT EQUILIBRIUM = 

   ( 0.151 ) <- 0 = a7
   ( 0.284 ) <- 1 = b7
   ( 0.359 ) <- 2 = g7
   ( 0.205 ) <- 3 = g8
   ( 0.178 ) <- 4 = e4
   ( 0.149 ) <- 5 = e5
   ( 0.246 ) <- 6 = b6
   ( 0.177 ) <- 7 = c6
   ( 0.167 ) <- 8 = c5
   ( 0.143 ) <- 9 = b5
   ( 0.144 ) <- 10 = b4
   ( 0.147 ) <- 11 = a4
   ( 0.161 ) <- 12 = a3
   ( 0.209 ) <- 13 = b3
   ( 0.278 ) <- 14 = c3
   ( 0.459 ) <- 15 = d3
   ( 0.730 ) <- 16 = d2
   ( 1.000 ) <- 17 = e2
   ( 0.558 ) <- 18 = e3
   ( 0.355 ) <- 19 = f3
   ( 0.231 ) <- 20 = f4
   ( 0.138 ) <- 21 = f5
   ( 0.132 ) <- 22 = f6
   ( 0.129 ) <- 23 = g6
   ( 0.122 ) <- 24 = g5
   ( 0.117 ) <- 25 = g4
   ( 0.101 ) <- 26 = h4
   ( 0.086 ) <- 27 = h5
   ( 0.059 ) <- 28 = h6
   ( 0.032 ) <- 29 = i6
   ( 0.329 ) <- 30 = d1
   ( 0.422 ) <- 31 = f1
   ( 0.971 ) <- 32 = f2
   ( 0.604 ) <- 33 = h2
   ( 0.510 ) <- 34 = h3
   ( 0.933 ) <- 35 = g2
   ( 0.604 ) <- 36 = g3
   ( 0.131 ) <- 37 = c1
   ( 0.131 ) <- 38 = c2
   ( 0.131 ) <- 39 = b2
   ( 0.131 ) <- 40 = b1
   ( 0.298 ) <- 41 = b8
   ( 0.293 ) <- 42 = c8
   ( 0.426 ) <- 43 = c7
   ( 0.597 ) <- 44 = d7
   ( 0.520 ) <- 45 = d8
   ( 0.519 ) <- 46 = e8
   ( 0.520 ) <- 47 = d6
   ( 0.519 ) <- 48 = e6
   ( 0.742 ) <- 49 = e7
   ( 0.568 ) <- 50 = f7
   ( 0.263 ) <- 51 = f8
   ( 0.122 ) <- 52 = h8
   ( 0.054 ) <- 53 = h7


* THERE ARE 8 LOCAL MAXIMA IN THE EQUILIBRIUM EVC LANDSCAPE:

  - subregion of maximum 37 consisting of nodes [37]
     - where 37 -> c1
  - subregion of maximum 39 consisting of nodes [39]
     - where 39 -> b2
  - subregion of maximum 44 consisting of nodes [42 43 44 45 47]
     - where 42 -> c8
     - where 43 -> c7
     - where 44 -> d7
     - where 45 -> d8
     - where 47 -> d6
  - subregion of maximum 41 consisting of nodes [0 1 6 7 8 41]
     - where 0 -> a7
     - where 1 -> b7
     - where 6 -> b6
     - where 7 -> c6
     - where 8 -> c5
     - where 41 -> b8
  - subregion of maximum 38 consisting of nodes [38]
     - where 38 -> c2
  - subregion of maximum 40 consisting of nodes [40]
     - where 40 -> b1
  - subregion of maximum 49 consisting of nodes [2 3 46 48 49 50 51 52 53]
     - where 2 -> g7
     - where 3 -> g8
     - where 46 -> e8
     - where 48 -> e6
     - where 49 -> e7
     - where 50 -> f7
     - where 51 -> f8
     - where 52 -> h8
     - where 53 -> h7
  - subregion of maximum 17 consisting of nodes [4 5 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36]
     - where 4 -> e4
     - where 5 -> e5
     - where 9 -> b5
     - where 10 -> b4
     - where 11 -> a4
     - where 12 -> a3
     - where 13 -> b3
     - where 14 -> c3
     - where 15 -> d3
     - where 16 -> d2
     - where 17 -> e2
     - where 18 -> e3
     - where 19 -> f3
     - where 20 -> f4
     - where 21 -> f5
     - where 22 -> f6
     - where 23 -> g6
     - where 24 -> g5
     - where 25 -> g4
     - where 26 -> h4
     - where 27 -> h5
     - where 28 -> h6
     - where 29 -> i6
     - where 30 -> d1
     - where 31 -> f1
     - where 32 -> f2
     - where 33 -> h2
     - where 34 -> h3
     - where 35 -> g2
     - where 36 -> g3

* HILL-CLIMBING EVC-LAMDSCAPE GRADIENT PATHS:

     - Path node 0 has local maximum at node * 41 *, hop distance 2 along [0 1 41]
     - Path node 1 has local maximum at node * 41 *, hop distance 1 along [1 41]
     - Path node 2 has local maximum at node * 49 *, hop distance 2 along [2 50 49]
     - Path node 3 has local maximum at node * 49 *, hop distance 3 along [3 2 50 49]
     - Path node 4 has local maximum at node * 17 *, hop distance 4 along [4 20 19 18 17]
     - Path node 5 has local maximum at node * 17 *, hop distance 5 along [5 4 20 19 18 17]
     - Path node 6 has local maximum at node * 41 *, hop distance 2 along [6 1 41]
     - Path node 7 has local maximum at node * 41 *, hop distance 3 along [7 6 1 41]
     - Path node 8 has local maximum at node * 41 *, hop distance 4 along [8 7 6 1 41]
     - Path node 9 has local maximum at node * 17 *, hop distance 8 along [9 10 11 12 13 14 15 16 17]
     - Path node 10 has local maximum at node * 17 *, hop distance 7 along [10 11 12 13 14 15 16 17]
     - Path node 11 has local maximum at node * 17 *, hop distance 6 along [11 12 13 14 15 16 17]
     - Path node 12 has local maximum at node * 17 *, hop distance 5 along [12 13 14 15 16 17]
     - Path node 13 has local maximum at node * 17 *, hop distance 4 along [13 14 15 16 17]
     - Path node 14 has local maximum at node * 17 *, hop distance 3 along [14 15 16 17]
     - Path node 15 has local maximum at node * 17 *, hop distance 2 along [15 16 17]
     - Path node 16 has local maximum at node * 17 *, hop distance 1 along [16 17]
     - Path node 17 has local maximum at node * 17 *, hop distance 0 along [17]
     - Path node 18 has local maximum at node * 17 *, hop distance 1 along [18 17]
     - Path node 19 has local maximum at node * 17 *, hop distance 2 along [19 18 17]
     - Path node 20 has local maximum at node * 17 *, hop distance 3 along [20 19 18 17]
     - Path node 21 has local maximum at node * 17 *, hop distance 6 along [21 5 4 20 19 18 17]
     - Path node 22 has local maximum at node * 17 *, hop distance 7 along [22 21 5 4 20 19 18 17]
     - Path node 23 has local maximum at node * 17 *, hop distance 8 along [23 22 21 5 4 20 19 18 17]
     - Path node 24 has local maximum at node * 17 *, hop distance 9 along [24 23 22 21 5 4 20 19 18 17]
     - Path node 25 has local maximum at node * 17 *, hop distance 10 along [25 24 23 22 21 5 4 20 19 18 17]
     - Path node 26 has local maximum at node * 17 *, hop distance 11 along [26 25 24 23 22 21 5 4 20 19 18 17]
     - Path node 27 has local maximum at node * 17 *, hop distance 12 along [27 26 25 24 23 22 21 5 4 20 19 18 17]
     - Path node 28 has local maximum at node * 17 *, hop distance 13 along [28 27 26 25 24 23 22 21 5 4 20 19 18 17]
     - Path node 29 has local maximum at node * 17 *, hop distance 14 along [29 28 27 26 25 24 23 22 21 5 4 20 19 18 17]
     - Path node 30 has local maximum at node * 17 *, hop distance 2 along [30 16 17]
     - Path node 31 has local maximum at node * 17 *, hop distance 2 along [31 32 17]
     - Path node 32 has local maximum at node * 17 *, hop distance 1 along [32 17]
     - Path node 33 has local maximum at node * 17 *, hop distance 3 along [33 35 32 17]
     - Path node 34 has local maximum at node * 17 *, hop distance 4 along [34 36 35 32 17]
     - Path node 35 has local maximum at node * 17 *, hop distance 2 along [35 32 17]
     - Path node 36 has local maximum at node * 17 *, hop distance 3 along [36 35 32 17]
     - Path node 37 has local maximum at node * 37 *, hop distance 0 along [37]
     - Path node 38 has local maximum at node * 38 *, hop distance 0 along [38]
     - Path node 39 has local maximum at node * 39 *, hop distance 0 along [39]
     - Path node 40 has local maximum at node * 40 *, hop distance 0 along [40]
     - Path node 41 has local maximum at node * 41 *, hop distance 0 along [41]
     - Path node 42 has local maximum at node * 44 *, hop distance 2 along [42 43 44]
     - Path node 43 has local maximum at node * 44 *, hop distance 1 along [43 44]
     - Path node 44 has local maximum at node * 44 *, hop distance 0 along [44]
     - Path node 45 has local maximum at node * 44 *, hop distance 1 along [45 44]
     - Path node 46 has local maximum at node * 49 *, hop distance 1 along [46 49]
     - Path node 47 has local maximum at node * 44 *, hop distance 1 along [47 44]
     - Path node 48 has local maximum at node * 49 *, hop distance 1 along [48 49]
     - Path node 49 has local maximum at node * 49 *, hop distance 0 along [49]
     - Path node 50 has local maximum at node * 49 *, hop distance 1 along [50 49]
     - Path node 51 has local maximum at node * 49 *, hop distance 2 along [51 50 49]
     - Path node 52 has local maximum at node * 49 *, hop distance 4 along [52 3 2 50 49]
     - Path node 53 has local maximum at node * 49 *, hop distance 5 along [53 52 3 2 50 49]

</pre>
and the report on the "doors.n4l" graph, for leadsto links:
<pre>

$ go run graph_report.go -chapter multi
----------------------------------------------------------------
Analysing chapter "multi slit interference", context [] to path length 6
----------------------------------------------------------------

* TOTAL NODES IN THE SEARCH REGION 13

* TOTAL DIRECTED LINKS = 17 of possible 156 = 0.11 %

* DISTRIBUTION OF NAME TYPE/LENGTHS:
  - single word ngram : 10 / 13
  - two word ngram : 3 / 13


* PROCESS ORIGINS / ROOT DEPENDENCIES / PATH SOURCES for ("+leads to") in multi slit interference

   - NPtr(1,0) -> start



* FINAL END-STATES / PATH SINK NODES for ("+leads to") in multi slit interference

   - NPtr(2,0) -> target 1
   - NPtr(2,1) -> target 2
   - NPtr(2,2) -> target 3

* DIRECTED LOOPS AND CYCLES:

   - Acyclic

* SYMMETRIZED EIGENVECTOR CENTRALITY = FLOW RESERVOIR CAPACITANCE AT EQUILIBRIUM = 

   ( 0.993 ) <- 0 = tram
   ( 0.768 ) <- 1 = target 3
   ( 0.847 ) <- 2 = gate
   ( 0.496 ) <- 3 = bike
   ( 1.000 ) <- 4 = start
   ( 0.787 ) <- 5 = door
   ( 0.940 ) <- 6 = port
   ( 0.678 ) <- 7 = hole
   ( 0.767 ) <- 8 = river
   ( 0.272 ) <- 9 = passage
   ( 0.093 ) <- 10 = target 1
   ( 0.272 ) <- 11 = road
   ( 0.093 ) <- 12 = target 2


* THERE ARE 2 LOCAL MAXIMA IN THE EQUILIBRIUM EVC LANDSCAPE:

  - subregion of maximum 0 consisting of nodes [0 1]
     - where 0 -> tram
     - where 1 -> target 3
  - subregion of maximum 4 consisting of nodes [2 3 4 5 6 7 8 9 10 11 12]
     - where 2 -> gate
     - where 3 -> bike
     - where 4 -> start
     - where 5 -> door
     - where 6 -> port
     - where 7 -> hole
     - where 8 -> river
     - where 9 -> passage
     - where 10 -> target 1
     - where 11 -> road
     - where 12 -> target 2

* HILL-CLIMBING EVC-LAMDSCAPE GRADIENT PATHS:

     - Path node 0 has local maximum at node * 0 *, hop distance 0 along [0]
     - Path node 1 has local maximum at node * 0 *, hop distance 1 along [1 0]
     - Path node 2 has local maximum at node * 4 *, hop distance 1 along [2 4]
     - Path node 3 has local maximum at node * 4 *, hop distance 2 along [3 2 4]
     - Path node 4 has local maximum at node * 4 *, hop distance 0 along [4]
     - Path node 5 has local maximum at node * 4 *, hop distance 1 along [5 4]
     - Path node 6 has local maximum at node * 4 *, hop distance 1 along [6 4]
     - Path node 7 has local maximum at node * 4 *, hop distance 1 along [7 4]
     - Path node 8 has local maximum at node * 4 *, hop distance 2 along [8 6 4]
     - Path node 9 has local maximum at node * 4 *, hop distance 2 along [9 5 4]
     - Path node 10 has local maximum at node * 4 *, hop distance 3 along [10 9 5 4]
     - Path node 11 has local maximum at node * 4 *, hop distance 2 along [11 5 4]
     - Path node 12 has local maximum at node * 4 *, hop distance 3 along [12 11 5 4]

</pre>
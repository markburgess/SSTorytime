
# pathsolve

`pathsolve` is an experimental tool for finding contiguous paths between node sets.
It can also be accessed through the web browser.

`pathsolve` also reports about two deeper analyses of the paths:

* *Betweenness centrality*:  a score for how many times each path passes through each node in the path sets.
The hiighest scoring nodes are 'most central' in the sense of flow throughput.

* *Supernodes*: these are nodes that form equivalence sets. The members of a supernode are interchangeable as far
as the path process is concerned. The map to and from the same locations, so they are symmetrical.

## Command line

For now, you can get started by trying the examples, e.g.
<pre>
$ cd examples
$ make
$ ../src/pathsolve -begin A1 -end B6 

mark% go run pathsolve.go -begin a1 -end b6 

 Paths < end_set= {B6, b6, } | {A1, } = start set>

     - story path: 1 * A1  -(forwards)->  A3  -(forwards)->  A5  -(forwards)->  S1
      -(forwards)->  B1  -(forwards)->  B4  -(forwards)->  B6

    Linkage process: -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)-> . 


     - story path: 2 * A1  -(forwards)->  A3  -(forwards)->  A5  -(forwards)->  S2
      -(forwards)->  B2  -(forwards)->  B4  -(forwards)->  B6

    Linkage process: -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)-> . 


     - story path: 3 * A1  -(forwards)->  A3  -(forwards)->  A6  -(forwards)->  S2
      -(forwards)->  B2  -(forwards)->  B4  -(forwards)->  B6

    Linkage process: -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)-> . 


     - story path: 4 * A1  -(forwards)->  A2  -(forwards)->  A5  -(forwards)->  S1
      -(forwards)->  B1  -(forwards)->  B4  -(forwards)->  B6

    Linkage process: -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)-> . 


     - story path: 5 * A1  -(forwards)->  A2  -(forwards)->  A5  -(forwards)->  S2
      -(forwards)->  B2  -(forwards)->  B4  -(forwards)->  B6

    Linkage process: -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)->  -(+leads to)-> . 

 *
 *
 * PATH ANALYSIS: into node flow equivalence groups
 *
 *

    - Super node 0 = {A1,}

    - Super node 1 = {A3,A2,}

    - Super node 2 = {A5,A6,}

    - Super node 3 = {S1,}

    - Super node 4 = {S2,}

    - Super node 5 = {B1,}

    - Super node 6 = {B2,}

    - Super node 7 = {B4,}

    - Super node 8 = {B6,}
 *
 *
 * FLOW IMPORTANCE:
 *
 *

    -Rank (betweenness centrality): 1.00 - B4,A1,B6,

    -Rank (betweenness centrality): 0.80 - A5,

    -Rank (betweenness centrality): 0.60 - S2,B2,A3,

    -Rank (betweenness centrality): 0.40 - A2,B1,S1,

    -Rank (betweenness centrality): 0.20 - A6,

</pre>

Or the adjoint path search:

<pre>

$ go run pathsolve.go -begin B6 -end A1 -bwd

</pre>
You can also use Dirac transition matrix notation like this:
<pre>

$ go run pathsolve.go "<B6|A1>"
$ go run pathsolve.go "<end|start>"
$ go run pathsolve.go "<target|start>"

</pre>
Notice the order of the start and end sets.

## Using in the web browser

In the search field, enter the Dirac notation, e.g. `<target|start>` and relevant chapter `interference`, then click on `geometry`.

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/pathsolve1.png 'pathsolving in a web interface')
![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/pathsolve2.png 'pathsolving in a web interface')


Notice the reporting about supernodes and betweenness centrality scores. 

## Speeding up path searches with restricted arrows

When searching for paths, the most powerful searches involve free association. However, searching with few constraints
is expensive, because the graph branches at every step, and therefore the number of possible paths grows exponentially.
One way to reduce this complexity is to specify the kinds of arrrows that are allowed. 
Arrows complicate searches, without necessarily offering much value, but--if your graph has consistent and simple link
types--this can greatly reduce the complexity of searches.

The directed nature of arrows makes this complicated too. When specifying arrows, you need to give both the forward and backwards arrows, because the search is made from start and end. The start sees outgoing forward links and the end sees outgoing backwards links. The general tool for path searching is the `GetConstraintConePathsAsLinks()` function, with or without arrows. This will, no doubt. improve with future versions, as there is still a lot to do to make graph searches smarter, but for now this is the most powerful approach.

Remember: the power of SST becomes more apparent when using the STTypes 0,1,2,3 for matching arrows by general type rather than by specific name.

### How?

Path searches grow exponentially with the length of the path, so they get slower and slower as the distance between nodes
increases. If you know the type of arrow along the whole path, you can speed up the search by specifying the arrow types, or the sttypes, e.g. using the STtypes:
<pre>
./searchN4L -v \\from \!gun\! \\to scarlet \\arrow +3,-3,0
</pre>
And using the arrows:
<pre>
./searchN4L -v \\from \!a1\! \\to b6 \\arrow 20,21
</pre>
Remember to always give pairs of arrow,inverse since the FROM and the TO match opposite arrow directions.
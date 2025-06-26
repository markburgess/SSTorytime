

# An API for interacting with the SST graph

You will find many examples of using Go(lang) code to write custom scripts
that interact with the database through the Go API
[here](https://github.com/markburgess/SSTorytime/tree/main/src/demo_pocs).

Notable examples:
* [Maze solver 1](../src/demo_pocs/search_maze.go) and * [Maze solver 2](../src/demo_pocs/search_maze2.go)
* [Path solver](../src/pathsolve.go)
* [Graph reporter](../src/graph_report.go)

## Creating an SST graph from data

See the [example](../src/API_EXAMPLE_1.go). To make node registration as easy as possible, you can use two functions
`Vertex()` and `Edge()` to create nodes and links respectively. These names are chosen to distance themselves
from the underlying `Node` and `Link`naming, by using the more mathematical names for these objects.

### Structure

Assuming the arrow names have been defined (e.g. by uploading them using N4L),
then to open the context channel for the database, we bracket the meat of a program with
Open and Close functions:

<pre>
func main() {

	load_arrows := false
	ctx := SST.Open(load_arrows)

	AddStory(ctx)
	LookupStory(ctx)

	SST.Close(ctx)
}

</pre>

### Add nodes and simple links from data

For the meat of an AddStory function, we can use the Vertex and Edge functions to avoid low level details.
Adding nodes to a database, without using the N4L language is straightforward:
<pre>
	chap := "home and away"
	context := []string{""}
	var w float32 = 1.0

	n1 := SST.Vertex(ctx,"Mary had a little lamb",chap)
	n2 := SST.Vertex(ctx,"Whose fleece was dull and grey",chap)

	n3 := SST.Vertex(ctx,"And every time she washed it clean",chap)
	n4 := SST.Vertex(ctx,"It just went to roll in the hay",chap)

	n5 := SST.Vertex(ctx,"And when it reached a certain age ",chap)
	n6 := SST.Vertex(ctx,"She'd serve it on a tray",chap)

	SST.Edge(ctx,n1,"then",n2,context,w)

	// bifurcation!

	SST.Edge(ctx,n2,"then",n3,context,w/2)
	SST.Edge(ctx,n2,"then",n5,context,w/2)

	// endings

	SST.Edge(ctx,n3,"then",n4,context,w)
	SST.Edge(ctx,n5,"then",n6,context,w)

</pre>

### Adding hub-joins or hyperlinks from data

See `API_EXAMPLE_2.go`. In a `HubJoin()` we provide a list of node pointers
to be linked together via  a common hub. This respects the Semantic Spacetime
rules and simulates a hyperlink.
<pre>
	// Then create a container mummy_node for all and arrow type is "is contained by"

	created := SST.HubJoin(ctx,"mummy_node","my chapter",nptrs,"is contained by",nil,nil)
	fmt.Println("Creates hub node",created2)

        // Create 
</pre>
If you don't set the chapter, and all the nodes belong to the same chapter,
then the method will adopt the same chapter as the children  belong to.
Similarly, if you don't want to give the hub node a name, then leave it empty:
<pre>
	created := SST.HubJoin(ctx,"","",nptrs,"then",context,weights)
	fmt.Println("Creates hub node",created1)

</pre>
the name of the node will be uniquely formed from a list of the node pointers,
starting "hub_<arrow>_<nodelist>".

### Reading the graph back

Looking up up the data is more complicated because there are many options.
This example looks for story paths starting from a node that we search for by name.
<ol>
<li>First we get a pointer to the starting node by random access lookup:
<pre>
	start_set := SST.GetDBNodePtrMatchingName(ctx,"Mary had a","")
</pre>
Because there might be several nodes that match your name description, this returns
an array of pointers.

<li>Next we want to know the Semantic Spacetime type of link to follow.
If you remember the numbers -3,-2,-1,0,1,2,3 of the link type (leadsto,contains,property,near)
you can select `sttype` directly. If you only remember the name of the relation, you can search
for it:
<pre>
	_,sttype := SST.GetDBArrowsWithArrowName(ctx,"then")
</pre>
<li>Setting a limit on the path length to explore, you search for the forward cone
of type `sttype` from the starting set of node pointers.
<pre>
	path_length := 4

	for n := range start_set {

		paths,_ := SST.GetFwdPathsAsLinks(ctx,start_set[n],sttype,path_length)

		for p := range paths {

			if len(paths[p]) > 1 {
			
				fmt.Println("    Path",p," len",len(paths[p]))

				for l := 0; l < len(paths[p]); l++ {

					// Find the long node name details from the pointer

					name := SST.GetDBNodeByNodePtr(ctx,paths[p][l].Dst).S

					fmt.Println("    ",l,"xx  --> ",
						paths[p][l].Dst,"=",name,"  , weight",
						paths[p][l].Wgt,"context",paths[p][l].Ctx)
				}
			}
		}
	}

</pre>
</ol>

### Checking the result

Running the `API_EXAMPLE.go` program:
<pre>
$ cd src
$ make
go build -o API_EXAMPLE API_EXAMPLE.go
$ ./API_EXAMPLE 
    Path 0  len 4
     0 xx  -->  {4 0} = Mary had a little lamb   , weight 1 context []
     1 xx  -->  {4 2} = Whose fleece was white as snow   , weight 1 context [cutting edge high brow poem]
     2 xx  -->  {4 3} = And everywhere that Mary went   , weight 1 context [cutting edge high brow poem]
     3 xx  -->  {4 4} = The lamb was sure to go   , weight 1 context [cutting edge high brow poem]
    Path 1  len 4
     0 xx  -->  {4 0} = Mary had a little lamb   , weight 1 context []
     1 xx  -->  {4 987} = Whose fleece was dull and grey   , weight 1 context []
     2 xx  -->  {4 988} = And every time she washed it clean   , weight 0.5 context []
     3 xx  -->  {4 989} = It just went to roll in the hay   , weight 1 context []
    Path 2  len 4
     0 xx  -->  {4 0} = Mary had a little lamb   , weight 1 context []
     1 xx  -->  {4 987} = Whose fleece was dull and grey   , weight 1 context []
     2 xx  -->  {4 990} = And when it reached a certain age    , weight 0.5 context []
     3 xx  -->  {4 991} = She'd serve it on a tray   , weight 1 context []

</pre>

## Searching a graph

We need to respect the geometry of the semantic spacetime when tracing and presenting paths.
Out major focus will tend to be by STtype.

* Find a starting place (random lookup).
* Decide on the capture region and criterion for search.
* * Select the cone of a particular STtype, for a picture of its relationships.
* * Find all possible paths, without restriction on semantics.

The search patterns can be:
<pre>
            by Name                                    GetNodeNotes/Orbits
START match by Chapter     ---> (set of NodePtr)  -->  GetFwdPaths (by STtype)
            by first Arrow                             GetFwdBwdPaths (by signless STtype)
            by Context                                 GetEntireCone (for all types)
</pre>

## Low level wrapper functions 

In general, you will want to use the special functions written for
querying the data.  These return data into Go structures directly,
performing all the marshalling and de-marshalling. The following are
basic workhorses. You will not normally use these.
For example, [see demo](https://github.com/markburgess/SSTorytime/blob/main/src/demo_pocs/postgres_stories.go).

<pre>
  :: low level API, golang, go programming ::

 +::data types::

 PoSST     (for) establishing a connection to the SST library service
 Node      (for) representing core aspects of a single graph node
 NodePtr   (for) unique key referring to a node and pointing to its data
 Link      (for) representing a graph link, with arrow and destination node and weight
 ArrowPtr  (for) A unique key for a type of link arrow and its properties
 PageMap   (for) representing the original N4L intended layout of notes

 -::data types::
 +::database upload functions::

"CreateDBNode(ctx PoSST, n Node) Node" (for) establishing a node structure in postgres
"UploadNodeToDB(ctx PoSST, org Node)"  (for) uploading an existing Node in memory to postgres
"UploadArrowToDB(ctx PoSST,arrow ArrowPtr)" (for) uploading an arrow definition from memory to postgres
"UploadInverseArrowToDB(ctx PoSST,arrow ArrowPtr)" (for) uploading an inverse arrow definition
"UploadPageMapEvent(ctx PoSST, line PageMap)" (for) uploading a PageMap structure from memory to postgres

"IdempDBAddLink(ctx PoSST,from Node,link Link,to Node)" (for) entry point for adding a link to a node in postgres
"CreateDBNodeArrowNode(ctx PoSST, org NodePtr, dst Link, sttype int) bool" (for) adding a NodeArrowNode secondary/derived structure to postgres

 -::database upload functions::
 +::database retrieve structural parts, retrieving::


"GetDBChaptersMatchingName(ctx PoSST,src string) []string" (for) retrieving chapter names
"GetDBContextsMatchingName(ctx PoSST,src string) []string" (for) retrieving context elements/dictionary with Node.S matching src string
"GetDBNodePtrMatchingName(ctx PoSST,src,chap string) []NodePtr" (for) retrieving a NodePtr to nodes with Node.S matching src string, node.Chap matching chap
"GetDBNodeByNodePtr(ctx PoSST,db_nptr NodePtr) Node" (for) retrieving a full Node structure from a NodePtr reference
"GetDBNodeArrowNodeMatchingArrowPtrs(ctx PoSST,chap string,cn []string,arrows []ArrowPtr) []NodeArrowNode" (for) retrieving a NodeArrowNode record in a given chapter and context by arrow type
"GetDBNodeContextsMatchingArrow(ctx PoSST,searchtext string,chap string,cn []string,arrow []ArrowPtr,page int) []QNodePtr" (for) retrieving contextualized node pointers involved in arrow criteria
"GetNodesStartingStoriesForArrow(ctx PoSST,arrow string) []NodePtr" (for) retrieving singleton nodes starting paths with a particular arrow
    " (see) "GetDBSingletonBySTType(ctx PoSST,sttypes []int,chap string,cn []string) ([]NodePtr,[]NodePtr)"
    " (see) "GetNCCNodesStartingStoriesForArrow(ctx PoSST,arrow string,chapter string,context []string) []NodePtr"
"GetNCCNodesStartingStoriesForArrow(ctx PoSST,arrow string,chapter string,context []string) []NodePtr" (for) retrieving singleton nodes starting paths with a particular arrow and matching context and chapter 
    " (see) "GetDBSingletonBySTType(ctx PoSST,sttypes []int,chap string,cn []string) ([]NodePtr,[]NodePtr)"
    " (see) "GetNodesStartingStoriesForArrow(ctx PoSST,arrow string) []NodePtr"
"GetDBSingletonBySTType(ctx PoSST,sttypes []int,chap string,cn []string) ([]NodePtr,[]NodePtr)" (for) retrieving a list of nodes that are sources or sinks in chapters and contexts of the graph with respect to a given link meta SSType
    "  (see) "GetNCCNodesStartingStoriesForArrow(ctx PoSST,arrow string,chapter string,context []string) []NodePtr"
    "  (see) "GetNodesStartingStoriesForArrow(ctx PoSST,arrow string) []NodePtr"

"GetDBArrowsWithArrowName(ctx PoSST,s string) ArrowPtr"       (for) retrieving all arrow details matching exact name
    " (see) "GetDBArrowByName(ctx PoSST,name string) ArrowPtr" 
"GetDBArrowsMatchingArrowName(ctx PoSST,s string) []ArrowPtr" (for) retrieving list of all arrow details matching name
"GetDBArrowByName(ctx PoSST,name string) ArrowPtr"   (for) retrieving all arrow details matching name from arrow directory 
     " (see) "GetDBArrowsWithArrowName(ctx PoSST,s string) ArrowPtr"
"GetDBArrowByPtr(ctx PoSST,arrowptr ArrowPtr) ArrowDirectory"
"GetDBPageMap(ctx PoSST,chap string,cn []string,page int) []PageMap" (for) retrieving a PageMap matching chapter, context and logical page number (note) pages are currently 60 items long
"GetFwdConeAsNodes(ctx PoSST, start NodePtr, sttype,depth int) []NodePtr" (for) retrieving the future cone set of Nodes from a given NodePtr, returned as NodePtr for orbit description
"GetFwdConeAsLinks(ctx PoSST, start NodePtr, sttype,depth int) []Link" (for) retrieving the future cone set of Nodes from a given NodePtr, returned as Link structures for path description
"GetFwdPathsAsLinks(ctx PoSST, start NodePtr, sttype,depth int) ([][]Link,int)" (for) retrieving the future cone set of Links from a given NodePtr as an array of paths, i.e. a double array of Link
"GetEntireConePathsAsLinks(ctx PoSST,orientation string,start NodePtr,depth int) ([][]Link,int)" (for) retrieving the cone set of Nodes from a given NodePtr in all directions, returned as Link structures for path description
"GetEntireNCConePathsAsLinks(ctx PoSST,orientation string,start NodePtr,depth int,chapter string,context []string) ([][]Link,int)" (for) retrieving the cone set of Nodes from a given NodePtr in all directions, returned as Link structures for path description and filtered by chapter and context, specifying direction fwd/bwd/any
"GetEntireNCSuperConePathsAsLinks(ctx PoSST,orientation string,start []NodePtr,depth int,chapter string,context []string) ([][]Link,int)" (for) retrieving the cone set of Nodes from a given multinode start set of NodePtr in all directions, returned as Link structures for path description, filtered by chapter and context, specifying direction fwd/bwd/any

 -::database retrieve structural parts::
 +::path integral:::

"GetPathsAndSymmetries(ctx PoSST,start_set,end_set []NodePtr,chapter string,context []string,maxdepth int) [][]Link" (for) retrieve solution paths between a starting set and and final set like +'<final|start>' in generalized way
"GetPathTransverseSuperNodes(ctx PoSST,solutions [][]Link,maxdepth int) [][]NodePtr" (for) establish the nodes that play idential roles in a set of paths from +'<final|start>' to see which nodes are redundant

  -::path integral:::
  +::adjacency matrix representation, graph vector support::

"GetDBAdjacentNodePtrBySTType(ctx PoSST,sttypes []int,chap string,cn []string) ([][]float32,[]NodePtr)" (for) retrieving the graph adjacenccy matrix as a square matrix of float32 link weights and an index to NodePointer lookup directory

 -::path integral:::
 +::orbits::

"GetNodeOrbit(ctx PoSST,nptr NodePtr,exclude_vector string) [ST_TOP][]Orbit" (for) retrieving the nearest neighbours of a NodePtr to maximum radius of three layers
"PrintNodeOrbit(ctx PoSST, nptr NodePtr,width int)" (for) printing a Node orbit in human readable form on the console, calling GetNodeOrbit
"PrintLinkOrbit(notes [ST_TOP][]Orbit,sttype int)" (for) printing an orbit in human readable form
"PrintLinkPath(ctx PoSST, cone [][]Link, p int, prefix string,chapter string,context []string)" (for) printing a Link array of paths in human readable form

</pre>


## Matroid Analysis Functions (nodes by appointed roles)

For examples, [see demo](https://github.com/markburgess/SSTorytime/blob/main/src/demo_pocs/search_clusters_functions.go) and [example](https://github.com/markburgess/SSTorytime/blob/main/src/demo_pocs/search_clusters.go).

These functions will most likely be used during browsing of data, when getting a feel for the size and shape of the data.

* `GetAppointmentArrayByArrow(ctx PoSST, context []string,chapter string) map[ArrowPtr][]NodePtr` - return a map of groups of nodes formed as matroids to arrows of all types, classified by arrow type.

* `GetAppointmentArrayBySSType(ctx PoSST) map[int][]NodePtr` - return a map of groups of nodes formed as matroids to arrows classified by STType.

* `GetAppointmentHistogramByArrow(ctx PoSST) map[ArrowPtr]int` - Find the group (member frequency) sizes of the above groups by arrow.

* `GetAppointmentHistogramBySSType(ctx PoSST) map[int]int` - Find the group (member frequency) sizes of the above groups by STType.

* `GetAppointmentNodesByArrow(ctx PoSST) []ArrowAppointment` - Return the group members of a matroid by arrow type.

* `GetAppointmentNodesBySTType(ctx PoSST) []STTypeAppointment` - Return the group members of a matroid by STType.



## Basic queries from SQL

Using perfectly standard SQL, you can interrogate the database established by N4L or the low level API
functions.

### Tables

* To show the different tables:
<pre>
$ psql storyline

storyline=# \dt
              List of relations
 Schema |      Name      | Type  |   Owner    
--------+----------------+-------+------------
 public | arrowdirectory | table | sstoryline
 public | arrowinverses  | table | sstoryline
 public | node           | table | sstoryline
 public | nodearrownode  | table | sstoryline
(4 rows)

</pre>
* To query these, we look at the members:
<pre>
storyline=# \d node
                Table "public.node"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 nptr   | nodeptr |           |          | 
 l      | integer |           |          | 
 s      | text    |           |          | 
 chap   | text    |           |          | 
 im3    | link[]  |           |          | 
 im2    | link[]  |           |          | 
 im1    | link[]  |           |          | 
 in0    | link[]  |           |          | 
 il1    | link[]  |           |          | 
 ic2    | link[]  |           |          | 
 ie3    | link[]  |           |          | 
Indexes:
    "node_chan_l_s_idx" btree (((nptr).chan), l, s)

</pre>

### Nodes

Now try:
<pre>
storyline=# select S,chap from Node limit 10;
     s      |       chap       
------------+------------------
 please     | notes on chinese
 yes        | notes on chinese
 请          | notes on chinese
 qǐng       | notes on chinese
 thankyou   | notes on chinese
 Méiyǒu     | notes on chinese
 谢谢        | notes on chinese
 xièxiè     | notes on chinese
 是的        | notes on chinese
 请在这里等    | notes on chinese
(10 rows)

</pre>

* An alternative view of relations is provided by NodeArrowNode:
<pre>
storyline=# select *  from NodeArrowNode LIMIT 10;
 nfrom | sttype | arr | wgt |              ctx              |   nto   
-------+--------+-----+-----+-------------------------------+---------
 (1,0) |     -1 |  69 |   1 | {please,"thank you",thankyou} | (1,1)
 (1,1) |     -1 |  67 |   1 | {thankyou,please,"thank you"} | (1,2)
 (1,1) |      1 |  68 |   1 | {thankyou,please,"thank you"} | (1,0)
 (1,1) |      1 |  68 |   1 | {news,online}                 | (2,291)
 (1,2) |      1 |  66 |   1 | {thankyou,please,"thank you"} | (1,1)
 (1,3) |     -1 |  69 |   1 | {please,"thank you",thankyou} | (1,4)
 (1,4) |     -1 |  67 |   1 | {please,"thank you",thankyou} | (1,5)
 (1,4) |      1 |  68 |   1 | {please,"thank you",thankyou} | (1,3)
 (1,5) |      1 |  66 |   1 | {please,"thank you",thankyou} | (1,4)
 (1,6) |     -1 |  67 |   1 | {please,"thank you",thankyou} | (4,0)
(10 rows)

</pre>

Notice how nodes (`nfrom`,`nto`,`nptr? ) and arrows (`arr`) are represented by pointer references
that are integers. When working with the graph, we often don't need to know the names
of things, we can get away with deferring the lookup of the actual data until we find what we're
looking for. That information can be cached so as to minimize the data transferred over the wire.

<pre>
storyline=# select S from Node where NPtr=(1,5);
   s    
--------
 xièxiè
(1 row)

</pre>

### Links and Arrows

A link is a composite relation that involves an arrow (pointer), a context,
and a destination node. Links are anchored to their origin nodes in the `Node` table
in the six columns `im3`, `im2`, `im1`, `in0`, `il1`, `ic2`, `ie3`.  
To find the links of type `Leads to':
<pre>
storyline=# select Il1 from Node where NPtr=(1,5);
                                       il1                                        
----------------------------------------------------------------------------------
 {"(66,1,\"{ \"\"please\"\", \"\"thank you\"\", \"\"thankyou\"\" }\",\"(1,4)\")"}
(1 row)

</pre>

Arrows are defined for each arrow pointer in the arrow directory:

<pre>
storyline=# select * from arrowdirectory limit 10;
 staindex |         long         | short | arrptr 
----------+----------------------+-------+--------
        4 | leads to             | lt    |      0
        2 | arriving from        | af    |      1
        4 | forwards             | fwd   |      2
        2 | backwards            | bwd   |      3
        4 | affects              | aff   |      4
        2 | affected by          | baff  |      5
        4 | causes               | cf    |      6
        2 | is caused by         | cb    |      7
        4 | used for             | for   |      8
        2 | is a possible use of | use   |      9
(10 rows)

</pre>

## The Go(lang) interfaces

The SSToryline package tries to make querying the data structures easy, by providing
generic scriptable functions that can be used easily in Go.

The open a database connection, to make any kind of query, with the help of the SSToryline package:
<pre>

package main

import (
	"fmt"
        SST "SSTorytime"
)

//******************************************************************

const (
	host     = "localhost"
	port     = 5432
	user     = "sstoryline"
	password = "sst_1234"
	dbname   = "storyline"
)

//******************************************************************

func main() {

	load_arrows := false

	ctx := SST.Open(load_arrows)

	row,err := ctx.DB.Query("SELECT NFrom,Arr,NTo FROM NodeArrowNode LIMIT 10")

	var a,c string	
	var b int

	for row.Next() {		
		err = row.Scan(&a,&b,&c)
		fmt.Println(a,b,c)
	}
	
	row.Close()

	SST.Close(ctx)
}

</pre>












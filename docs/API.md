

# An API for interacting with the SST graph

*(These preliminary functions are unlikely to be final public functions in the final API, they
are still in the development and research phase of testing )*

Once data have been entered into a SSToryline database, we want to be able to extract it again.
It's possible to create tools for this, but ultimately any set of tools will tend to limit the user.
A user's imagination should be the only limit. 

Many specialized graph databases offer graph languages, but they
expose an important problem with Domain Specific Languages, which is
that by trying to make simple things easy, they make less-simple
things hard. The most well known standard for data (Structured Query
Language, or SQL) is itself a Domain Specific Language with exactly
these problems. However, in Open Source Postgres there are plenty of
extensions that make it possible to overcome the limitations of SQL.

*This project uses Postgres because of that compromise between a well known
standard, and a battle-tested and extensible data platform.*

You will find examples of using Go(lang) code to write custom scripts
that interact with the database through the Go API [here](https://github.com/markburgess/SSTorytime/tree/main/src/demo_pocs).

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
$ psql newdb

newdb=# \dt
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
newdb=# \d node
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
newdb=# select S,chap from Node limit 10;
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
newdb=# select *  from NodeArrowNode LIMIT 10;
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
newdb=# select S from Node where NPtr=(1,5);
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
newdb=# select Il1 from Node where NPtr=(1,5);
                                       il1                                        
----------------------------------------------------------------------------------
 {"(66,1,\"{ \"\"please\"\", \"\"thank you\"\", \"\"thankyou\"\" }\",\"(1,4)\")"}
(1 row)

</pre>

Arrows are defined for each arrow pointer in the arrow directory:

<pre>
newdb=# select * from arrowdirectory limit 10;
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
	dbname   = "newdb"
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












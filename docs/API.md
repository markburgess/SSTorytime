

# An API for interacting with the SST graph

The simplest way to manage graphs is to use the N4L language to create
and manage them, then rebuild the whole database consistently as a
cache of the information. This makes editing very intuitive and simple.
However, sometimes you want to work directly with a programming interface.

The most flexible way to deal with graphs is to program using the
API. There are two API approaches you can use (in principle):

- Adding nodes and links without caring about assigning node pointers. This uses the `SST.Vertex(ctx,str,chap)` and `SST.Edge(ctx,nfrom,"fwd",nto,context,w)` functions.
- Adding nodes and links in bulk and managing node pointers yourself. This uses the `SST.IdempDBAddNode(ctx, np)`, `SST.IdempDBAddLink(ctx,np,lnk,nt)`, and `SST.CreateDBNodeArrowNode(ctx,np,lnk,sttype)`.

Most will use only the first of these.  Both interfaces behave
idempotently and work in the same way. The only difference is whether
you allocate `NodePtr` values yourself (which is efficient for large
numbers of entrites).

Because the database behaves like a cache, optimized for quite different use-cases,
we have to maintain several tables.

<pre>
sstoryline=# \dt
              List of relations
 Schema |      Name      | Type  |   Owner    
--------+----------------+-------+------------
 public | arrowdirectory | table | sstoryline
 public | arrowinverses  | table | sstoryline
 public | node           | table | sstoryline
 public | nodearrownode  | table | sstoryline
 public | pagemap        | table | sstoryline

</pre>

Another thing we need to do is register arrow definitions used in links/edges.
For this we use two functions: `SST.InsertArrowDirectory(stname,alias,name,pm string)` and
`SST.InsertInverseArrowDirectory(fwd,bwd ArrowPtr)`. We need to register arrows before using them
in links.

## Examples

The API examples is the `src` directory are stripped down to the minimum, and the special
tools llike the pathsolver and notes tool show how to make simple wrappers for the API functions too.
You will also find many examples of using Go(lang) code to write custom scripts
that interact with the database through the Go API
[here](https://github.com/markburgess/SSTorytime/tree/main/src/demo_pocs).


## Creating an SST graph from data

See the [example](../src/API_EXAMPLE_1.go). To make node registration as easy as possible, you can use two functions
`Vertex()` and `Edge()` to create nodes and links respectively. These names are chosen to distance themselves
from the underlying `Node` and `Link`naming, by using the more mathematical names for these objects.

### Open/Close the connection to SST

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

### Add nodes and links from data

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

### Adding hub-joins (hyperlinks) from data

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

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/geometry.png 'Testing a web interface')

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
 :: golang API, database upload functions, self-managed NPtr  ::

"CreateDBNode(ctx PoSST, n Node) Node" (use-for) establishing a node in postgres without auto %NPtr assignment
"UploadNodeToDB(ctx PoSST, org Node)"  (use-for) uploading an existing Node in memory to postgres
"UploadArrowToDB(ctx PoSST,arrow ArrowPtr)" (use-for) uploading an arrow definition from memory to postgres
"UploadInverseArrowToDB(ctx PoSST,arrow ArrowPtr)" (use-for) uploading an inverse arrow definition
"UploadPageMapEvent(ctx PoSST, line PageMap)" (use-for) uploading a PageMap structure from memory to postgres

 :: database upload functions, DB-managed NPtr  ::

"IdempDBAddNode(ctx PoSST,n Node) Node" (use-for) appending a node when you don't want to manage the NPtr values.
"IdempDBAddLink(ctx PoSST,from Node,link Link,to Node)" (use-for) entry point for adding a link to a node in postgres

"CreateDBNodeArrowNode(ctx PoSST, org NodePtr, dst Link, sttype int) bool" (use-for) adding a NodeArrowNode secondary/derived structure to postgres

 :: golang API, searching functions ::

"GetDBNodePtrMatchingName(ctx PoSST,name,chap string) []NodePtr"  (use-for) finding a list of %NPtr matching a substring by name
"GetDBNodePtrMatchingNCCS(ctx PoSST,nm,chap string,cn []string,arrow []ArrowPtr,seq bool,limit int) []NodePtr"  (use-for) Comprehensive search by %"NCCS criteria"
"GetDBChaptersMatchingName(ctx PoSST,src string) []string"  (use-for) obtaining a list of chapters matching by name
"GetDBContextByName(ctx PoSST,src string) (string,ContextPtr)"  (use-for) obtaining context sets that match by string name
"GetDBContextByPtr(ctx PoSST,ptr ContextPtr) (string,ContextPtr)"  (use-for) obtaining the context set with given index pointer
"GetSTtypesFromArrows(arrows []ArrowPtr) []int"  (use-for) obtaining the generic semantic spacetime type for a given a list of arrow pointers
"GetDBSingletonBySTType(ctx PoSST,sttypes []int,chap string,cn []string) ([]NodePtr,[]NodePtr)"  (use-for) find nodes that are sources or sinks for a specific STType
"SelectStoriesByArrow(ctx PoSST,nodeptrs []NodePtr, arrowptrs []ArrowPtr, sttypes []int, limit int) []NodePtr"  (use-for) finding nodes that are sources for story sequences matching the arrow types

"GetDBArrowsWithArrowName(ctx PoSST,s string) (ArrowPtr,int)"  (use-for) obtaining an arrowpointer and STType matching a precise name
"GetDBArrowsMatchingArrowName(ctx PoSST,s string) []ArrowPtr" (use-for) obtaining a list of arrowpointers matching the approximate name
"GetDBArrowByName(ctx PoSST,name string) ArrowPtr"  (use-for) obtaining an arrowpointer for precise arrowname - redundant
"GetDBArrowByPtr(ctx PoSST,arrowptr ArrowPtr) ArrowDirectory" (use-for) obtains an arrow directory entry for a given arrow pointer
"GetDBArrowBySTType(ctx PoSST,sttype int) []ArrowDirectory" (use-for) obtains the arrow directory for a given STtype
"GetDBPageMap(ctx PoSST,chap string,cn []string,page int) []PageMap"  (use-for) obtains a page from the named chapter as a page map

 :: golang API, causal cones ::

"GetFwdConeAsNodes(ctx PoSST, start NodePtr, sttype,depth int,limit int) []NodePtr"  (use-for) obtains the orbit around a starting node as a set of nodes to given depth
"GetFwdPathsAsLinks(ctx PoSST, start NodePtr, sttype,depth int, maxlimit int) ([][]Link,int)"  (use-for) obtains all possible paths from a node along STtype links
"GetEntireNCSuperConePathsAsLinks(ctx PoSST,orientation string,start []NodePtr,depth int,chapter string,context []string,limit int) ([][]Link,int)" (use-for) obtains all possible paths from a starting node, with orientation "fwd,bwd,any" as link arrays matching the chapter and context criteria

 :: search language ::

"SolveNodePtrs(ctx PoSST,nodenames []string,search SearchParameters,arr []ArrowPtr,limit int) []NodePtr"  (use-for) finding a set of matching NPtrs satisfying the search parameters compiled by a search command


</pre>


## Basic queries from SQL

Using perfectly standard SQL, you can interrogate the database established by N4L or the low level API
functions.

### Tables

* To show the different tables:
<pre>
$ psql storyline

storyline=# \dt
               List of relations
 Schema |       Name       | Type  |   Owner    
--------+------------------+-------+------------
 public | arrowdirectory   | table | sstoryline
 public | arrowinverses    | table | sstoryline
 public | contextdirectory | table | sstoryline
 public | lastseen         | table | sstoryline
 public | node             | table | sstoryline
 public | pagemap          | table | sstoryline
(6 rows)

</pre>
* To query these, we look at the members:
<pre>
storyline=# \d node
sstoryline-# \d Node
                                                     Table "public.node"
  Column  |   Type   | Collation | Nullable |                                     Default                                     
----------+----------+-----------+----------+---------------------------------------------------------------------------------
 nptr     | nodeptr  |           |          | 
 l        | integer  |           |          | 
 s        | text     |           |          | 
 search   | tsvector |           |          | generated always as (to_tsvector('english'::regconfig, s)) stored
 unsearch | tsvector |           |          | generated always as (to_tsvector('english'::regconfig, sst_unaccent(s))) stored
 chap     | text     |           |          | 
 seq      | boolean  |           |          | 
 im3      | link[]   |           |          | 
 im2      | link[]   |           |          | 
 im1      | link[]   |           |          | 
 in0      | link[]   |           |          | 
 il1      | link[]   |           |          | 
 ic2      | link[]   |           |          | 
 ie3      | link[]   |           |          | 
Indexes:
    "sst_gin" gin (search)
    "sst_type" btree (((nptr).chan), l, s)
    "sst_ungin" gin (unsearch)


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













# Web JSON queries

Query data returned by searches are addressed to the `http_server.go`
program handler as POST requests, either with a command string into a
`name` field, or as an NPtr reference sent as two separate form
variables `nclass` and `ncptr` for the two components of and `NPtr`.

The returned data are best understood as one of two paradigms: either
i) story paths, which are arrays of `Link` structures (the first of
which is a singleton (a destination Node without an arrow), or ii)
orbitals, which model a single node at a time with a cloud of arrow
paths up to two arrows in length, and corresponding to satellites of
the original node (see the figure below to see the geometry).

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/geometry.png 'Testing a web interface')

Data are returned in JSON packages by the `PackageResponse()` function, which
wraps data in a structure:
<pre>
{
 "Response" : SWITCHCASE,
 "Content"  : data,
 "Time"     : <internal>,
 "Intent"   : <internal>, 
 "Ambient"  : <internal> 
}
</pre>
where the `data` are returned in one of a number of formats (below).
The `SWITCHCASE` is handled in the web javascript parser as follows:
<pre>
  switch (SWITCHCASE) 
     {
     case "Orbits":
          DoOrbitPanel(resp);
          break;
     case "ConePaths":
          DoEntireConePanel(resp);
          break;
     case "PathSolve":
          DoEntireConePanel(resp);
          break;
     case "Sequence":
          DoSeqPanel(resp);
          break;
     case "PageMap":
          DoPageMapPanel(resp);
          break;
     case "TOC":
          DoTOCPanel(resp);
          break;
     case "Arrows":
          DoArrowsPanel(resp);
          break;
     case "STAT":
          DoStatsPanel(resp);
          break;
     }
</pre>
So each of these functions basically renders a fixed type JSON structure, in a manner appropriate to its purpose.


### NodeEvents and their Orbits

The simplest kind of lookup is a single topic match, returning a
Node. When we get a Node, we return the `S` string, the `NPtr`
reference and the `Chapter` and `Context` sets associated with each
match. Each orbiting is a collection of `NodeEvent`s in JSON:

### NodeEvent

A `NodeEvent` is a structur containing the data for a single Node information.
It refers to `Orbits` which are precompiled paths reaching out from the Node
to a depth of two hops. Coordinates `XYZ` are pre-calculated by the server based on the
semantics of the search.

<pre>
type NodeEvent struct {

	Text    string
	L       int
	Chap    string
	Context string
        NPtr    NodePtr
	XYZ     Coords
	Orbits  [ST_TOP][]Orbit
}
</pre>

The orbital references are listed as a set of arrays much like the
STtype collection of `Link` arrays in the server and database's
internal Node representation. There are 7 lists indicating the arrow
name, it's spacetime time and the destination Node Ptr and its text
along with relative coordinates. `OOO` is the origin coordinate of the
body a satellite orbits, (i.e. radius one orbits the original search
node, radius 2 orbits one of the radius one satellites).  There is
sufficient information in this to collate all the satellites of a
particular STtype into a single list.  <pre> type Orbit struct {

	Radius  int
	Arrow   string
	STindex int
	Dst     NodePtr
	Ctx     string
	Text    string
	XYZ     Coords  // coords
	OOO     Coords  // origin
}
</pre>

### WebPath

Go internal `Link` arrays are transformed into `WebPath` objects that have all pointers expanded.
<pre>
type WebPath struct {
	NPtr    NodePtr
	Arr     ArrowPtr
	STindex int
	Line    int
	Name    string
	Chp     string
	Ctx     string
	XYZ     Coords
}
</pre>

### Story

Story searches are a simple wrapping of a `NodeEvent` array, with encapsulating Chapter.
<pre>
type Story struct {

	Chapter   string
	Axis      []NodeEvent
}
</pre>


### WebConePaths

`WebConePaths` are wrappers around arrays of `WebPath` arrays
(analogous to the `[][]Link` arrays in the figure above), which are
held in the `Paths` field. Additional information is provided by the
server: The title of the cone, as some explanation of its semantics,
an array of 'betweenness centrality' scores for the paths, and an
array of supernode sets that are calculated by the server based on
path
[symmetries](https://mark-burgess-oslo-mb.medium.com/semantic-spacetime-1-the-shape-of-knowledge-86daced424a5).

<pre>
type WebConePaths struct {

	RootNode   NodePtr
	Title      string
	BTWC       []string
	Paths      [][]WebPath
	SuperNodes []string
}
</pre>

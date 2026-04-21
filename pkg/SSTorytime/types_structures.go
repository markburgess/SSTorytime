//**************************************************************
//
// types_structures.go
//
//**************************************************************

package SSTorytime

import (
	"database/sql"
	_ "github.com/lib/pq"

)

//**************************************************************

type PoSST struct {

	DB *sql.DB
	// Session globals
	
	NODE_DIRECTORY NodeDirectory  // Internal histo-representations
	NODE_CACHE map[NodePtr]NodePtr
        BASE_DB_CHANNEL_STATE[7] ClassedNodePtr

	ARROW_DIRECTORY []ArrowDirectory
	ARROW_SHORT_DIR map[string]ArrowPtr
	ARROW_LONG_DIR map[string]ArrowPtr
	ARROW_DIRECTORY_TOP ArrowPtr
	INVERSE_ARROWS map[ArrowPtr]ArrowPtr

	// Context array factorization

	CONTEXT_DIRECTORY []ContextDirectory
	CONTEXT_DIR map[string]ContextPtr
	CONTEXT_TOP ContextPtr

	// Page layout
	
	PAGE_MAP []PageMap

}

//******************************************************************

type Etc struct {

	E bool  // event
	T bool  // thing
	C bool  // concept
}

//******************************************************************

type Node struct {

	L         int     // length of text string
	S         string  // text string itself

	Seq       bool    // true if this node begins an intended sequence, otherwise ambiguous
	Chap      string  // section/chapter name in which this was added
	NPtr      NodePtr // Pointer to self index
	Psi       Etc     // induced node type (experimental)

	I [ST_TOP][]Link  // link incidence list, by STindex - these are the "vectors" +/-
  	                  // NOTE: carefully how STindex offsets represent negative SSTtypes
}

//**************************************************************

type Link struct {  // A link is a type of arrow, with context
                    // and maybe with a weight for package math
	Arr ArrowPtr         // type of arrow, presorted
	Wgt float32          // numerical weight of this link
	Ctx ContextPtr       // context for this pathway
	Dst NodePtr          // adjacent event/item/node
}

//**************************************************************

type NodePtr struct {

	Class int            // Text size-class, used mainly in memory
	CPtr  ClassedNodePtr // index of within name class lane
}

//**************************************************************

type ClassedNodePtr int  // Internal pointer type of size-classified text

//**************************************************************

type ArrowDirectory struct {

	STAindex  int
	Long    string
	Short   string
	Ptr     ArrowPtr
}

//**************************************************************

type ArrowPtr int // ArrowDirectory index

//**************************************************************

type ContextDirectory struct {

	Context string
	Ptr     ContextPtr
}

//**************************************************************

type ContextPtr int // ContextDirectory index

//**************************************************************

type PageMap struct {  // Thereis additional intent in the layout

	Chapter string
	Alias   string
	Context ContextPtr
	Line    int
	Path    []Link
}

//**************************************************************

type Appointment struct {

        // An appointed from node points to a collection of to nodes 
        // by the same arrow

	Arr ArrowPtr
	STType int
	Chap string
	Ctx []string
	NTo NodePtr
	NFrom []NodePtr
}

//**************************************************************

type NodeDirectory struct {

	// Power law n-gram frequencies

	N1grams map[string]ClassedNodePtr
	N1directory []Node
	N1_top ClassedNodePtr

	N2grams map[string]ClassedNodePtr
	N2directory []Node
	N2_top ClassedNodePtr

	N3grams map[string]ClassedNodePtr
	N3directory []Node
	N3_top ClassedNodePtr

	// Use linear search on these exp fewer long strings

	LT128 []Node
	LT128_top ClassedNodePtr
	LT1024 []Node
	LT1024_top ClassedNodePtr
	GT1024 []Node
	GT1024_top ClassedNodePtr
}


//**************************************************************

type PageView struct {
	Title   string
	Context string
	Notes   [][]WebPath
}

//**************************************************************

type Coords struct {
	X   float64
	Y   float64
	Z   float64
	R   float64
	Lat float64
	Lon float64
}

//**************************************************************

type WebPath struct {
	NPtr    NodePtr
	Arr     ArrowPtr
	STindex int
	Line    int     // used for pagemap
	Name    string
	Chp     string
	Ctx     string
	XYZ     Coords
	Wgt     float32
}

//******************************************************************

type Story struct {

	// The title of a story is a property of the sequence
        // not a container for it. It belongs to the sequence context.

	Chapter   string  // chapter it belongs to
	Axis      []NodeEvent
}

//******************************************************************

type NodeEvent struct {

	Text    string
	L       int
	Chap    string
	Context string
        NPtr    NodePtr
	XYZ     Coords
	Orbits  [ST_TOP][]Orbit
}

//******************************************************************

type WebConePaths struct {

	RootNode   NodePtr
	Title      string
	BTWC       []string
	Paths      [][]WebPath
	SuperNodes []string
}

//******************************************************************

type Orbit struct {  // union, JSON transformer

	Radius  int
	Arrow   string
	STindex int
	Dst     NodePtr
	Ctx     string
	Wgt     float32
	Text    string
	XYZ     Coords  // coords
	OOO     Coords  // origin
}

//******************************************************************

type Loc struct {

	Text string
	Reln []int
	XYZ  Coords
}

//******************************************************************

type ChCtx struct {
	Chapter  string
	XYZ      Coords
	Context  []Loc
	Single   []Loc
	Common   []Loc
}

//******************************************************************

type LastSeen struct {
	Section string
	First   int64    // timestamp of first access
	Last    int64    // timestamp of last access
        Pdelta  float64  // previous average of access intervals
	Ndelta  float64  // current last access interval
	Freq    int      // count of total accesses
	NPtr    NodePtr
	XYZ     Coords
}

//**************************************************************

type History struct {

	Freq  float64  // just use float becasue we'll want to calculate
	Last  int64    // calc gradient and purge
	Delta int64
	Time  string
}



//
// types_structures.go
//



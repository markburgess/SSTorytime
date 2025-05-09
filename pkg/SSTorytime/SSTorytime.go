
//**************************************************************
//
// An interface for postgres for graph analytics and semantics
//
//**************************************************************

package SSTorytime

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"unicode"
	"sort"
	"encoding/json"

	_ "github.com/lib/pq"

)

//**************************************************************
// Errors
//**************************************************************

const (
	ERR_ST_OUT_OF_BOUNDS="Link STtype is out of bounds (must be -3 to +3)"
	ERR_ILLEGAL_LINK_CLASS="ILLEGAL LINK CLASS"
	ERR_NO_SUCH_ARROW = "No such arrow has been declared in the configuration: "
	ERR_MEMORY_DB_ARROW_MISMATCH = "Arrows in database are not in synch (shouldn't happen)"
	WARN_DIFFERENT_CAPITALS = "WARNING: Another capitalization exists"

	SCREENWIDTH = 100
	RIGHTMARGIN = 5
	LEFTMARGIN = 5

	NEAR = 0
	LEADSTO = 1   // +/-
	CONTAINS = 2  // +/-
	EXPRESS = 3   // +/-

	// And shifted indices for array indicesin Go

	ST_ZERO = EXPRESS
	ST_TOP = ST_ZERO + EXPRESS + 1

	// For the SQL table, as 2d arrays not good

	I_MEXPR = "Im3"
	I_MCONT = "Im2"
	I_MLEAD = "Im1"
	I_NEAR  = "In0"
	I_PLEAD = "Il1"
	I_PCONT = "Ic2"
	I_PEXPR = "Ie3"

	// For separating text types

	N1GRAM = 1
	N2GRAM = 2
	N3GRAM = 3
	LT128 = 4
	LT1024 = 5
	GT1024 = 6
)

//**************************************************************

type Node struct {

	L         int     // length of text string
	S         string  // text string itself

	Chap      string  // section/chapter name in which this was added
	NPtr      NodePtr // Pointer to self index

	I [ST_TOP][]Link  // link incidence list, by STindex
  	                  // NOTE: carefully how STindex offsets represent negative SSTtypes
}

//**************************************************************

type NodeArrowNode struct {

	NFrom NodePtr
	STType int
	Arr ArrowPtr
	Wgt float64
	Ctx []string
	NTo NodePtr
}

//**************************************************************

type QNodePtr struct {

	// A Qualified NodePtr 

	NPtr    NodePtr
	Context string  // array in string form
	Chapter string
}

//**************************************************************

type PageMap struct {  // Thereis additional intent in the layout

	Chapter string
	Alias   string
	Context []string
	Line    int
	Path    []Link
}

type PageView struct {
	Title   string
	Context string
	Notes   [][]WebPath
}

type WebPath struct {
	NPtr    NodePtr
	Arr     ArrowPtr
	STindex int
	Name    string
}

//**************************************************************

type STTypeAppointment struct {

	NFrom NodePtr
	STType int
	NTo []NodePtr
}

//**************************************************************

type ArrowAppointment struct {

	NFrom NodePtr
	Arr ArrowPtr
	NTo []NodePtr
}

//**************************************************************

type Link struct {  // A link is a type of arrow, with context
                    // and maybe with a weightfor package math
	Arr ArrowPtr         // type of arrow, presorted
	Wgt float64          // numerical weight of this link
	Ctx []string         // context for this pathway
	Dst NodePtr          // adjacent event/item/node
}

const NODEPTR_TYPE = "CREATE TYPE NodePtr AS  " +
	"(                    " +
	"Chan     int,        " +
	"CPtr     int         " +
	")"

const LINK_TYPE = "CREATE TYPE Link AS  " +
	"(                    " +
	"Arr      int,        " +
	"Wgt      real,       " +
	"Ctx      text,       " +
	"Dst      NodePtr     " +
	")"

const NODE_TABLE = "CREATE TABLE IF NOT EXISTS Node " +
	"( " +
	"NPtr      NodePtr,        " +
	"L         int,            " +
	"S         text,           " +
	"Chap      text,           " +
	I_MEXPR+"  Link[],         " + // Im3
	I_MCONT+"  Link[],         " + // Im2
	I_MLEAD+"  Link[],         " + // Im1
	I_NEAR +"  Link[],         " + // In0
	I_PLEAD+"  Link[],         " + // Il1
	I_PCONT+"  Link[],         " + // Ic2
	I_PEXPR+"  Link[]          " + // Ie3
	")"

const LINK_TABLE = "CREATE TABLE IF NOT EXISTS NodeArrowNode " +
	"( " +
	"NFrom    NodePtr, " +
	"STtype   int,     " +
	"Arr      int,     " +
	"Wgt      int,     " +
	"Ctx      text[],  " +
	"NTo      NodePtr  " +
	")"

const PAGEMAP_TABLE = "CREATE TABLE IF NOT EXISTS PageMap " +
	"( " +
	"Chap     Text,  " +
	"Alias    Text,  " +
	"Ctx      Text[]," +
	"Line     Int,   " +
	"Path     Link[] " +
	")"

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

var NODE_CACHE = make(map[NodePtr]NodePtr)

//**************************************************************

type NodePtr struct {

	Class int            // Text size-class
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

const ARROW_DIRECTORY_TABLE = "CREATE TABLE IF NOT EXISTS ArrowDirectory " +
	"(    " +
	"STAindex int,             " +
	"Long text,              " +
	"Short text,             " +
	"ArrPtr int primary key  " +
	")"

const ARROW_INVERSES_TABLE = "CREATE TABLE IF NOT EXISTS ArrowInverses " +
	"(    " +
	"Plus int,  " +
	"Minus int,  " +
	"Primary Key(Plus,Minus)" +
	")"

//**************************************************************
// Lookup tables
//**************************************************************

var ( 
	ARROW_DIRECTORY []ArrowDirectory
	ARROW_SHORT_DIR = make(map[string]ArrowPtr) // Look up short name int referene
	ARROW_LONG_DIR = make(map[string]ArrowPtr)  // Look up long name int referene
	ARROW_DIRECTORY_TOP ArrowPtr = 0
	INVERSE_ARROWS = make(map[ArrowPtr]ArrowPtr)

	PAGE_MAP []PageMap

	NODE_DIRECTORY NodeDirectory  // Internal histo-representations

	NO_NODE_PTR NodePtr // see Init()

	WIPE_DB bool = false
        SILLINESS_COUNTER int
        SILLINESS_POS int
	SILLINESS bool
)

//******************************************************************
// LIBRARY
//******************************************************************

type PoSST struct {

   DB *sql.DB
}

//******************************************************************

type Story struct {

	ContainNPtr NodePtr
	Text     string
        Arrow     string  // arrow connecting the story to its container
	Axis      []NodeEvent
}

//******************************************************************

type NodeEvent struct {

	Text    string
	L       int
	Chap    string
        NPtr    NodePtr
	Orbits  [ST_TOP][]Orbit
}

//******************************************************************

type Orbit struct {  // union, JSON transformer

	Radius  int
	Arrow   string
	STindex int
	Dst     NodePtr
	Ctx     string
	Text    string
}

//******************************************************************

func Open(load_arrows bool) PoSST {

	var ctx PoSST
	var err error

	// Replace this with a private file

	const (
		host     = "localhost"
		port     = 5432
		user     = "sstoryline"
		password = "sst_1234"
		dbname   = "sstoryline"
	)

        connStr := "user="+user+" dbname="+dbname+" password="+password+" sslmode=disable"

        ctx.DB, err = sql.Open("postgres", connStr)

	if err != nil {
	   	fmt.Println("Error connecting to the database: ", err)
		os.Exit(-1)
	}
	
	err = ctx.DB.Ping()
	
	if err != nil {
		fmt.Println("Error pinging the database: ", err)
		os.Exit(-1)
	}

	MemoryInit()
	Configure(ctx,load_arrows)

	NO_NODE_PTR.Class = 0
	NO_NODE_PTR.CPtr =  -1

	return ctx
}

// **************************************************************************

func MemoryInit() {

	if NODE_DIRECTORY.N1grams == nil {
		NODE_DIRECTORY.N1grams = make(map[string]ClassedNodePtr)
	}

	if NODE_DIRECTORY.N2grams == nil {
		NODE_DIRECTORY.N2grams = make(map[string]ClassedNodePtr)
	}

	if NODE_DIRECTORY.N3grams == nil {
		NODE_DIRECTORY.N3grams = make(map[string]ClassedNodePtr)
	}
}

// **************************************************************************

func Configure(ctx PoSST,load_arrows bool) {

	// Tmp reset

	if WIPE_DB {

		fmt.Println("***********************")
		fmt.Println("* WIPING DB")
		fmt.Println("***********************")
		
		ctx.DB.QueryRow("drop function fwdconeaslinks")
		ctx.DB.QueryRow("drop function fwdconeasnodes")
		ctx.DB.QueryRow("drop function fwdpathsaslinks")
		ctx.DB.QueryRow("drop function getfwdlinks")
		ctx.DB.QueryRow("drop function getfwdnodes")
		ctx.DB.QueryRow("drop function getneighboursbytype")
		ctx.DB.QueryRow("drop function getsingletonaslink")
		ctx.DB.QueryRow("drop function AllNCPathsAsLinks")
		ctx.DB.QueryRow("drop function AllSuperNCPathsAsLinks")
		ctx.DB.QueryRow("drop function SumAllNCPaths")
		ctx.DB.QueryRow("drop function GetNCFwdLinks")
		ctx.DB.QueryRow("drop function GetNCCLinks")

		ctx.DB.QueryRow("drop function getsingletonaslinkarray")
		ctx.DB.QueryRow("drop function idempinsertnode")
		ctx.DB.QueryRow("drop function sumfwdpaths")
		ctx.DB.QueryRow("drop function match_context")
		ctx.DB.QueryRow("drop function empty_path")
		ctx.DB.QueryRow("drop function match_arrows")
		ctx.DB.QueryRow("drop function ArrowInList")
		ctx.DB.QueryRow("drop function GetStoryStartNodes")
		ctx.DB.QueryRow("drop function GetNCCStoryStartNodes")

		ctx.DB.QueryRow("drop table Node")
		ctx.DB.QueryRow("drop table PageMap")
		ctx.DB.QueryRow("drop table NodeArrowNode")
		ctx.DB.QueryRow("drop type NodePtr")
		ctx.DB.QueryRow("drop type Link")

		ctx.DB.QueryRow("drop table ArrowDirectory")
		ctx.DB.QueryRow("drop table ArrowInverses")
	}

	// Ignore error
	ctx.DB.QueryRow("CREATE EXTENSION unaccent")

	if !CreateType(ctx,NODEPTR_TYPE) {
		fmt.Println("Unable to create type as, ",NODEPTR_TYPE)
		os.Exit(-1)
	}

	if !CreateType(ctx,LINK_TYPE) {
		fmt.Println("Unable to create type as, ",LINK_TYPE)
		os.Exit(-1)
	}

	if !CreateTable(ctx,PAGEMAP_TABLE) {
		fmt.Println("Unable to create table as, ",PAGEMAP_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(ctx,NODE_TABLE) {
		fmt.Println("Unable to create table as, ",NODE_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(ctx,LINK_TABLE) {
		fmt.Println("Unable to create table as, ",LINK_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(ctx,ARROW_INVERSES_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_INVERSES_TABLE)
		os.Exit(-1)
	}
	if !CreateTable(ctx,ARROW_DIRECTORY_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_DIRECTORY_TABLE)
		os.Exit(-1)
	}

	DefineStoredFunctions(ctx)

	if load_arrows {
		DownloadArrowsFromDB(ctx)
	}
}

// **************************************************************************

func Close(ctx PoSST) {
	ctx.DB.Close()
}

// **************************************************************************
// In memory representation structures
// **************************************************************************

func GetNodeTxtFromPtr(frptr NodePtr) string {

	class := frptr.Class
	index := frptr.CPtr

	var node Node

	switch class {
	case N1GRAM:
		node = NODE_DIRECTORY.N1directory[index]
	case N2GRAM:
		node = NODE_DIRECTORY.N2directory[index]
	case N3GRAM:
		node = NODE_DIRECTORY.N3directory[index]
	case LT128:
		node = NODE_DIRECTORY.LT128[index]
	case LT1024:
		node = NODE_DIRECTORY.LT1024[index]
	case GT1024:
		node = NODE_DIRECTORY.GT1024[index]
	}

	return node.S
}

// **************************************************************************

func GetNodeFromPtr(frptr NodePtr) Node {

	class := frptr.Class
	index := frptr.CPtr

	var node Node

	switch class {
	case N1GRAM:
		node = NODE_DIRECTORY.N1directory[index]
	case N2GRAM:
		node = NODE_DIRECTORY.N2directory[index]
	case N3GRAM:
		node = NODE_DIRECTORY.N3directory[index]
	case LT128:
		node = NODE_DIRECTORY.LT128[index]
	case LT1024:
		node = NODE_DIRECTORY.LT1024[index]
	case GT1024:
		node = NODE_DIRECTORY.GT1024[index]
	}

	return node
}

//**************************************************************

func AppendTextToDirectory(event Node,ErrFunc func(string)) NodePtr {

	var cnode_slot ClassedNodePtr = -1
	var ok bool = false
	var node_alloc_ptr NodePtr

	cnode_slot,ok = CheckExistingOrAltCaps(event,ErrFunc)

	node_alloc_ptr.Class = event.NPtr.Class

	if ok {
		node_alloc_ptr.CPtr = cnode_slot
		IdempAddChapterToNode(node_alloc_ptr.Class,node_alloc_ptr.CPtr,event.Chap)
		return node_alloc_ptr
	}

	switch event.NPtr.Class {
	case N1GRAM:
		cnode_slot = NODE_DIRECTORY.N1_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.N1directory = append(NODE_DIRECTORY.N1directory,event)
		NODE_DIRECTORY.N1grams[event.S] = cnode_slot
		NODE_DIRECTORY.N1_top++ 
		return node_alloc_ptr
	case N2GRAM:
		cnode_slot = NODE_DIRECTORY.N2_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.N2directory = append(NODE_DIRECTORY.N2directory,event)
		NODE_DIRECTORY.N2grams[event.S] = cnode_slot
		NODE_DIRECTORY.N2_top++
		return node_alloc_ptr
	case N3GRAM:
		cnode_slot = NODE_DIRECTORY.N3_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.N3directory = append(NODE_DIRECTORY.N3directory,event)
		NODE_DIRECTORY.N3grams[event.S] = cnode_slot
		NODE_DIRECTORY.N3_top++
		return node_alloc_ptr
	case LT128:
		cnode_slot = NODE_DIRECTORY.LT128_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.LT128 = append(NODE_DIRECTORY.LT128,event)
		NODE_DIRECTORY.LT128_top++
		return node_alloc_ptr
	case LT1024:
		cnode_slot = NODE_DIRECTORY.LT1024_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.LT1024 = append(NODE_DIRECTORY.LT1024,event)
		NODE_DIRECTORY.LT1024_top++
		return node_alloc_ptr
	case GT1024:
		cnode_slot = NODE_DIRECTORY.GT1024_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.GT1024 = append(NODE_DIRECTORY.GT1024,event)
		NODE_DIRECTORY.GT1024_top++
		return node_alloc_ptr
	}

	return NO_NODE_PTR
}

//**************************************************************

func CheckExistingOrAltCaps(event Node,ErrFunc func(string)) (ClassedNodePtr,bool) {

	var cnode_slot ClassedNodePtr = -1
	var ok bool = false
	ignore_caps := false

	switch event.NPtr.Class {
	case N1GRAM:
		cnode_slot,ok = NODE_DIRECTORY.N1grams[event.S]
	case N2GRAM:
		cnode_slot,ok = NODE_DIRECTORY.N2grams[event.S]
	case N3GRAM:
		cnode_slot,ok = NODE_DIRECTORY.N3grams[event.S]
	case LT128:
		cnode_slot,ok = LinearFindText(NODE_DIRECTORY.LT128,event,ignore_caps)
	case LT1024:
		cnode_slot,ok = LinearFindText(NODE_DIRECTORY.LT1024,event,ignore_caps)
	case GT1024:
		cnode_slot,ok = LinearFindText(NODE_DIRECTORY.GT1024,event,ignore_caps)
	}

	if ok {
		return cnode_slot,ok
	} else {
		// Check for alternative caps

		ignore_caps = true
		alternative_caps := false
		
		switch event.NPtr.Class {
		case N1GRAM:
			for key := range NODE_DIRECTORY.N1grams {
				if strings.ToLower(key) == strings.ToLower(event.S) {
					alternative_caps = true
				}
			}
		case N2GRAM:
			for key := range NODE_DIRECTORY.N2grams {
				if strings.ToLower(key) == strings.ToLower(event.S) {
					alternative_caps = true
				}
			}
		case N3GRAM:
			for key := range NODE_DIRECTORY.N3grams {
				if strings.ToLower(key) == strings.ToLower(event.S) {
					alternative_caps = true
				}
			}

		case LT128:
			_,alternative_caps = LinearFindText(NODE_DIRECTORY.LT128,event,ignore_caps)
		case LT1024:
			_,alternative_caps = LinearFindText(NODE_DIRECTORY.LT1024,event,ignore_caps)
		case GT1024:
			_,alternative_caps = LinearFindText(NODE_DIRECTORY.GT1024,event,ignore_caps)
		}

		if alternative_caps {
			ErrFunc(WARN_DIFFERENT_CAPITALS+" ("+event.S+")")
		}

	}
	return cnode_slot,ok
}

//**************************************************************

func IdempDBAddNode(ctx PoSST,n Node) Node {

	// alternative for np = SST.CreateDBNode(ctx, np)
	// without assuming management/control of the Nptr increments

	var qstr string

	// No need to trust the values, ignore/overwrite CPtr

        n.L,n.NPtr.Class = StorageClass(n.S)

	es := SQLEscape(n.S)
	ec := SQLEscape(n.Chap)

	// Wrap BEGIN/END a single transaction

	qstr = fmt.Sprintf("SELECT IdempAppendNode(%d,%d,'%s','%s')",n.L,n.NPtr.Class,es,ec)

	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to add node",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		return n
	}

	var whole string
	var cl,ch int

	for row.Next() {		
		err = row.Scan(&whole)
		fmt.Sscanf(whole,"(%d,%d)",&cl,&ch)
	}

	n.NPtr.Class = cl
	n.NPtr.CPtr = ClassedNodePtr(ch)

	row.Close()

	return n

}

//**************************************************************

func IdempAddChapterToNode(class int,cptr ClassedNodePtr,chap string) {

	/* In the DB version, we have handle chapter collisions
           we want all similar names to have a single node for lateral
           association, but we need to be able to search by chapter too,
           so merge the chapters as an attribute list */

	var node Node

	switch class {
	case N1GRAM:
		node = NODE_DIRECTORY.N1directory[cptr]
	case N2GRAM:
		node = NODE_DIRECTORY.N2directory[cptr]
	case N3GRAM:
		node = NODE_DIRECTORY.N3directory[cptr]
	case LT128:
		node = NODE_DIRECTORY.LT128[cptr]
	case LT1024:
		node = NODE_DIRECTORY.LT1024[cptr]
	case GT1024:
		node = NODE_DIRECTORY.GT1024[cptr]
	}

	if strings.Contains(node.Chap,chap) {
		return
	}
	
	newchap := node.Chap + "," + chap

	switch class {
	case N1GRAM:
		NODE_DIRECTORY.N1directory[cptr].Chap = newchap
	case N2GRAM:
		NODE_DIRECTORY.N2directory[cptr].Chap = newchap
	case N3GRAM:
		NODE_DIRECTORY.N3directory[cptr].Chap = newchap
	case LT128:
		NODE_DIRECTORY.LT128[cptr].Chap = newchap
	case LT1024:
		NODE_DIRECTORY.LT1024[cptr].Chap = newchap
	case GT1024:
		NODE_DIRECTORY.GT1024[cptr].Chap = newchap
	}
}

//**************************************************************

func AppendLinkToNode(frptr NodePtr,link Link,toptr NodePtr) {

	frclass := frptr.Class
	frm := frptr.CPtr
	stindex := ARROW_DIRECTORY[link.Arr].STAindex

	link.Dst = toptr // fill in the last part of the reference

	// Add idempotently ...

	switch frclass {

	case N1GRAM:
		NODE_DIRECTORY.N1directory[frm].I[stindex] = MergeLinks(NODE_DIRECTORY.N1directory[frm].I[stindex],link)
	case N2GRAM:
		NODE_DIRECTORY.N2directory[frm].I[stindex] = MergeLinks(NODE_DIRECTORY.N2directory[frm].I[stindex],link)
	case N3GRAM:
		NODE_DIRECTORY.N3directory[frm].I[stindex] = MergeLinks(NODE_DIRECTORY.N3directory[frm].I[stindex],link)
	case LT128:
		NODE_DIRECTORY.LT128[frm].I[stindex] = MergeLinks(NODE_DIRECTORY.LT128[frm].I[stindex],link)
	case LT1024:
		NODE_DIRECTORY.LT1024[frm].I[stindex] = MergeLinks(NODE_DIRECTORY.LT1024[frm].I[stindex],link)
	case GT1024:
		NODE_DIRECTORY.GT1024[frm].I[stindex] = MergeLinks(NODE_DIRECTORY.GT1024[frm].I[stindex],link)
	}
}

//**************************************************************

func MergeLinks(list []Link,lnk Link) []Link {

	var ctx []string

	for c := range lnk.Ctx { // strip redundant signal
		if lnk.Ctx[c] != "_sequence_" {
			ctx = append(ctx,lnk.Ctx[c])
		}
	}

	lnk.Ctx = ctx

	for l := range list {
		if list[l].Arr == lnk.Arr && list[l].Dst == lnk.Dst {
			list[l].Ctx = MergeContexts(list[l].Ctx,ctx)
			return list
		}
	}

	list = append(list,lnk)
	return list
}

//**************************************************************

func MergeContexts(one,two []string) []string {

	var merging = make(map[string]bool)
	var merged []string

	for s := range one {
		merging[one[s]] = true
	}

	for s := range two {
		merging[two[s]] = true
	}

	for s := range merging {
		if s != "_sequence_" {
			merged = append(merged,s)
		}
	}

	return merged
}

//**************************************************************

func LinearFindText(in []Node,event Node,ignore_caps bool) (ClassedNodePtr,bool) {

	for i := 0; i < len(in); i++ {

		if event.L != in[i].L {
			continue
		}

		if ignore_caps {
			if strings.ToLower(in[i].S) == strings.ToLower(event.S) {
				return ClassedNodePtr(i),true
			}
		} else {
			if in[i].S == event.S {
				return ClassedNodePtr(i),true
			}
		}
	}

	return -1,false
}

//**************************************************************

func GetSTIndexByName(stname,pm string) int {

	var encoding  int
	var sign int

	switch pm {
	case "+":
		sign = 1
	case "-":
		sign = -1
	}

	switch stname {

	case "leadsto":
		encoding = ST_ZERO + LEADSTO * sign
	case "contains":
		encoding = ST_ZERO + CONTAINS * sign
	case "properties":
		encoding = ST_ZERO + EXPRESS * sign
	case "similarity":
		encoding = ST_ZERO + NEAR
	}

	return encoding

}

//**************************************************************

func PrintSTAIndex(stindex int) string {

	sttype := stindex - ST_ZERO
	var ty string

	switch sttype {
	case -EXPRESS:
		ty = "-(expressed by)"
	case -CONTAINS:
		ty = "-(part of)"
	case -LEADSTO:
		ty = "-(arriving from)"
	case NEAR:
		ty = "(close to)"
	case LEADSTO:
		ty = "+(leading to)"
	case CONTAINS:
		ty = "+(containing)"
	case EXPRESS:
		ty = "+(expressing)"
	default:
		ty = "unknown relation!"
	}

	const green = "\x1b[36m"
	const endgreen = "\x1b[0m"

	return green + ty + endgreen
}

//**************************************************************

func InsertArrowDirectory(stname,alias,name,pm string) ArrowPtr {

	// Insert an arrow into the forward/backward indices

	var newarrow ArrowDirectory

	newarrow.STAindex = GetSTIndexByName(stname,pm)
	newarrow.Long = name
	newarrow.Short = alias
	newarrow.Ptr = ARROW_DIRECTORY_TOP

	ARROW_DIRECTORY = append(ARROW_DIRECTORY,newarrow)
	ARROW_SHORT_DIR[alias] = ARROW_DIRECTORY_TOP
	ARROW_LONG_DIR[name] = ARROW_DIRECTORY_TOP
	ARROW_DIRECTORY_TOP++

	return ARROW_DIRECTORY_TOP-1
}

//**************************************************************

func InsertInverseArrowDirectory(fwd,bwd ArrowPtr) {

	// Lookup inverse by long name, only need this in search presentation

	INVERSE_ARROWS[fwd] = bwd
	INVERSE_ARROWS[bwd] = fwd
}

//**************************************************************
// Write to database
//**************************************************************

func GraphToDB(ctx PoSST) {

	fmt.Println("Storing nodes...")

	for class := N1GRAM; class <= GT1024; class++ {
		switch class {
		case N1GRAM:
			for n := range NODE_DIRECTORY.N1directory {
				org := NODE_DIRECTORY.N1directory[n]
				UploadNodeToDB(ctx,org)
			}
		case N2GRAM:
			for n := range NODE_DIRECTORY.N2directory {
				org := NODE_DIRECTORY.N2directory[n]
				UploadNodeToDB(ctx,org)
			}
		case N3GRAM:
			for n := range NODE_DIRECTORY.N3directory {
				org := NODE_DIRECTORY.N3directory[n]
				UploadNodeToDB(ctx,org)
			}
		case LT128:
			for n := range NODE_DIRECTORY.LT128 {
				org := NODE_DIRECTORY.LT128[n]
				UploadNodeToDB(ctx,org)
			}
		case LT1024:
			for n := range NODE_DIRECTORY.LT1024 {
				org := NODE_DIRECTORY.LT1024[n]
				UploadNodeToDB(ctx,org)
			}

		case GT1024:
			for n := range NODE_DIRECTORY.GT1024 {
				org := NODE_DIRECTORY.GT1024[n]
				UploadNodeToDB(ctx,org)
			}
		}
	}


	fmt.Println("\nStoring Arrows...")

	for arrow := range ARROW_DIRECTORY {

		UploadArrowToDB(ctx,ArrowPtr(arrow))
	}

	fmt.Println("Storing inverse Arrows...")

	for arrow := range INVERSE_ARROWS {

		UploadInverseArrowToDB(ctx,ArrowPtr(arrow))
	}

	fmt.Println("Storing page map...")

	for line := 0; line < len(PAGE_MAP); line ++ {
		UploadPageMapEvent(ctx,PAGE_MAP[line])
	}

	// CREATE INDICES

	fmt.Println("Indexing ....")

	ctx.DB.QueryRow("CREATE INDEX on NodeArrowNode (Arr,STType)")
	ctx.DB.QueryRow("CREATE INDEX on Node (((NPtr).Chan),L,S)")
}

// **************************************************************************
// Postgres
// **************************************************************************

func CreateType(ctx PoSST, defn string) bool {

	row,err := ctx.DB.Query(defn)

	if err != nil {
		s := fmt.Sprintln("Failed to create datatype PGLink ",err)
		
		if strings.Contains(s,"already exists") {
			return true
		} else {
			return false
		}
	}

	row.Close();
	return true
}

// **************************************************************************

func CreateTable(ctx PoSST,defn string) bool {

	row,err := ctx.DB.Query(defn)
	
	if err != nil {
		s := fmt.Sprintln("Failed to create a table %.10 ...",defn,err)
		
		if strings.Contains(s,"already exists") {
			return true
		} else {
			return false
		}
	}

	row.Close()
	return true
}

// **************************************************************************
// Store
// **************************************************************************

func CreateDBNode(ctx PoSST, n Node) Node {

	var qstr string

	// No need to trust the values

        n.L,n.NPtr.Class = StorageClass(n.S)
	
	cptr := n.NPtr.CPtr
	es := SQLEscape(n.S)
	ec := SQLEscape(n.Chap)

	qstr = fmt.Sprintf("SELECT IdempInsertNode(%d,%d,%d,'%s','%s')",n.L,n.NPtr.Class,cptr,es,ec)

	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to insert",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		return n
	}

	var whole string
	var cl,ch int

	for row.Next() {		
		err = row.Scan(&whole)
		fmt.Sscanf(whole,"(%d,%d)",&cl,&ch)
	}

	n.NPtr.Class = cl
	n.NPtr.CPtr = ClassedNodePtr(ch)

	row.Close()

	return n
}

// **************************************************************************

func UploadNodeToDB(ctx PoSST, org Node) {

	CreateDBNode(ctx, org)

	const nolink = 999
	var empty Link

	for stindex := range org.I {

		for lnk := range org.I[stindex] {

			dstlnk := org.I[stindex][lnk]
			sttype := STIndexToSTType(stindex)

			AppendDBLinkToNode(ctx,org.NPtr,dstlnk,sttype)
			CreateDBNodeArrowNode(ctx,org.NPtr,dstlnk,sttype)
			Waiting()
		}

		CreateDBNodeArrowNode(ctx,org.NPtr,empty,nolink)
		Waiting()
	}
}

// **************************************************************************

func UploadArrowToDB(ctx PoSST,arrow ArrowPtr) {

	staidx := ARROW_DIRECTORY[arrow].STAindex
	long := SQLEscape(ARROW_DIRECTORY[arrow].Long)
	short := SQLEscape(ARROW_DIRECTORY[arrow].Short)

	qstr := fmt.Sprintf("INSERT INTO ArrowDirectory (STAindex,Long,Short,ArrPtr) VALUES (%d,'%s','%s',%d)",staidx,long,short,arrow)

	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to insert",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		return
	}

	row.Close()
}

// **************************************************************************

func UploadInverseArrowToDB(ctx PoSST,arrow ArrowPtr) {

	plus := arrow
	minus := INVERSE_ARROWS[arrow]

	qstr := fmt.Sprintf("INSERT INTO ArrowInverses (Plus,Minus) VALUES (%d,%d)",plus,minus)

	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to insert",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		return
	}

	row.Close()
}

//**************************************************************

func UploadPageMapEvent(ctx PoSST, line PageMap) {

	qstr := fmt.Sprintf("INSERT INTO PageMap (Chap,Alias,Ctx,Line) VALUES ('%s','%s',%s,%d)",line.Chapter,line.Alias,FormatSQLStringArray(line.Context),line.Line)

	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to insert pagemap event",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		row.Close()
		return
	}

	row.Close()

	for lnk := 0; lnk < len(line.Path); lnk++ {

		linkval := fmt.Sprintf("(%d, %f, %s, (%d,%d)::NodePtr)",line.Path[lnk].Arr,line.Path[lnk].Wgt,FormatSQLStringArray(line.Path[lnk].Ctx),line.Path[lnk].Dst.Class,line.Path[lnk].Dst.CPtr)

		literal := fmt.Sprintf("%s::Link",linkval)
		
		qstr := fmt.Sprintf("UPDATE PageMap SET Path=array_append(Path,%s) WHERE Chap = '%s' AND Line = '%d'",literal,line.Chapter,line.Line)
		
		row,err := ctx.DB.Query(qstr)
		
		if err != nil {
			fmt.Println("Failed to append",err,qstr)
		}
		
		row.Close()
	}
}

//**************************************************************

func IdempDBAddLink(ctx PoSST,from Node,link Link,to Node) {

	frptr := from.NPtr
	toptr := to.NPtr

	link.Dst = toptr // it might have changed, so override

	if frptr == toptr {
		fmt.Println("Self-loops are not allowed",from.S)
		os.Exit(-1)
	}

	if link.Arr < 0 {
		fmt.Println("No arrows have yet been defined, so you can't rely on the arrow names")
		os.Exit(-1)
	}

	sttype := STIndexToSTType(ARROW_DIRECTORY[link.Arr].STAindex)

	AppendDBLinkToNode(ctx,frptr,link,sttype)

	// Double up the reverse definition for easy indexing of both in/out arrows
	// But be careful not the make the graph undirected by mistake

	var invlink Link
	invlink.Arr = INVERSE_ARROWS[link.Arr]
	invlink.Wgt = link.Wgt
	invlink.Dst = frptr
	AppendDBLinkToNode(ctx,toptr,invlink,-sttype)
}

// **************************************************************************

func AppendDBLinkToNode(ctx PoSST, n1ptr NodePtr, lnk Link, sttype int) bool {

	// Want to make this idempotent, because SQL is not (and not clause)

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS,sttype)
		os.Exit(-1)
	}

	if n1ptr == lnk.Dst {
		return false
	}

	//                       Arr,Wgt,Ctx,  Dst
	linkval := fmt.Sprintf("(%d, %f, %s, (%d,%d)::NodePtr)",lnk.Arr,lnk.Wgt,FormatSQLStringArray(lnk.Ctx),lnk.Dst.Class,lnk.Dst.CPtr)

	literal := fmt.Sprintf("%s::Link",linkval)

	link_table := STTypeDBChannel(sttype)

	qstr := fmt.Sprintf("UPDATE NODE set %s=array_append(%s,%s) where (NPtr).CPtr = '%d' and (NPtr).Chan = '%d' and (%s is null or not %s = ANY(%s))",
		link_table,
		link_table,
		literal,
		n1ptr.CPtr,
		n1ptr.Class,
		link_table,
		literal,
		link_table)

	row,err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("Failed to append",err,qstr)
	       return false
	}

	row.Close()
	return true
}

// **************************************************************************

func CreateDBNodeArrowNode(ctx PoSST, org NodePtr, dst Link, sttype int) bool {

	qstr := fmt.Sprintf("SELECT IdempInsertNodeArrowNode(" +
		"%d," + //infromptr
		"%d," + //infromchan
		"%d," + //isttype
		"%d," + //iarr
		"%.2f," + //iwgt
		"%s," + //ictx
		"%d," + //intoptr
		"%d " + //intochan,
		")",
		org.CPtr,
		org.Class,
		sttype,
		dst.Arr,
		dst.Wgt,
		FormatSQLStringArray(dst.Ctx),
		dst.Dst.CPtr,
		dst.Dst.Class)

	row,err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("Failed to make node-arrow-node",err,qstr)
	       return false
	}

	row.Close()
	return true
}

// **************************************************************************

func DefineStoredFunctions(ctx PoSST) {

	// NB! these functions are in "plpgsql" language, NOT SQL. They look similar but they are DIFFERENT!
	
	// Insert a node structure, also an anchor for and containing link arrays
	
	cols := I_MEXPR+","+I_MCONT+","+I_MLEAD+","+I_NEAR +","+I_PLEAD+","+I_PCONT+","+I_PEXPR

	qstr := fmt.Sprintf("CREATE OR REPLACE FUNCTION IdempInsertNode(iLi INT, iszchani INT, icptri INT, iSi TEXT, ichapi TEXT)\n" +
		"RETURNS TABLE (    \n" +
		"    ret_cptr INTEGER," +
		"    ret_channel INTEGER" +
		") AS $fn$ " +
		"DECLARE \n" +
		"BEGIN\n" +
		"  IF NOT EXISTS (SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE lower(s) = lower(iSi)) THEN\n" +
		"     INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap,%s) VALUES (iszchani,icptri,iLi,iSi,ichapi,'{}','{}','{}','{}','{}','{}','{}');" +
		"  END IF;\n" +
		"  RETURN QUERY SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;",cols);

	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	qstr = "CREATE OR REPLACE FUNCTION IdempAppendNode(iLi INT, iszchani INT, iSi TEXT, ichapi TEXT)\n" +
		"RETURNS TABLE (    \n" +
		"    ret_cptr INTEGER," +
		"    ret_channel INTEGER" +
		") AS $fn$ " +
		"DECLARE \n" +
		"    icptri INT = 0;" +
		"BEGIN\n" +
		"  IF NOT EXISTS (SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi) THEN\n" +
		"     SELECT max((Nptr).CPtr) INTO icptri FROM Node WHERE (Nptr).Chan=iszchani;\n"+
		"     INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap) VALUES (iszchani,icptri+1,iLi,iSi,ichapi);" +
		"  END IF;\n" +
		"  RETURN QUERY SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;";

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// For lookup by arrow

	qstr = "CREATE OR REPLACE FUNCTION IdempInsertNodeArrowNode\n" +
		"(\n" +
		"infromptr  int,   \n" +
		"infromchan int,   \n" +
		"isttype    int,   \n" +
		"iarr       int,   \n" +
		"iwgt       real,  \n" +
		"ictx       text[],\n" +
		"intoptr    int,   \n" +
		"intochan   int    \n" +
		")\n" +

		"RETURNS real AS $fn$ " +

		"DECLARE \n" +
		"  ret_wgt real;\n" +
		"BEGIN\n" +

		"  IF NOT EXISTS (SELECT Wgt FROM NodeArrowNode WHERE (NFrom).Cptr=infromptr AND Arr=iarr AND (NTo).Cptr=intoptr) THEN\n" +

		"     INSERT INTO NodeArrowNode (nfrom.Cptr,nfrom.Chan,sttype,arr,wgt,ctx,nto.Cptr,nto.Chan) \n" +
		"       VALUES (infromptr,infromchan,isttype,iarr,iwgt,ictx,intoptr,intochan);" +

		"  END IF;\n" +
		"  SELECT Wgt into ret_wgt FROM NodeArrowNode WHERE (NFrom).Cptr=infromptr AND Arr=iarr AND (NTo).Cptr=intoptr;\n" +
		"  RETURN ret_wgt;" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;";

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct an empty link pointing nowhere as a starting node

	qstr = "CREATE OR REPLACE FUNCTION GetSingletonAsLinkArray(start NodePtr)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    level Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,Array[]::text[],(0,0));\n"+
		"BEGIN\n"+
		" lnk.Dst = start;\n"+
		" level = array_append(level,lnk);\n"+
		"RETURN level; \n"+
		"END ;\n"+
		"$fn$ LANGUAGE plpgsql;"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct an empty link pointing nowhere as a starting node

	qstr = "CREATE OR REPLACE FUNCTION GetSingletonAsLink(start NodePtr)\n"+
		"RETURNS Link AS $fn$\n"+
		"DECLARE \n"+
		"    lnk Link := (0,1.0,Array[]::text[],(0,0));\n"+
		"BEGIN\n"+
		" lnk.Dst = start;\n"+
		"RETURN lnk; \n"+
		"END ;\n"+
		"$fn$ LANGUAGE plpgsql;"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct search by sttype. since table names are static we need a case statement

	qstr = "CREATE OR REPLACE FUNCTION GetNeighboursByType(start NodePtr, sttype int)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    fwdlinks Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,Array[]::text[],(0,0));\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"     SELECT %s INTO fwdlinks FROM Node WHERE Nptr=start;\n",st,STTypeDBChannel(st));
	}
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN fwdlinks; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Get the nearest neighbours as NPtr, with respect to each of the four STtype

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetFwdNodes(start NodePtr,exclude NodePtr[],sttype int)\n"+
		"RETURNS NodePtr[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours NodePtr[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks =GetNeighboursByType(start,sttype);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +

		"    neighbours := ARRAY[]::NodePtr[];\n" +

		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+
		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk.dst);\n" +
		"      END IF; \n" +
		"    END LOOP;\n" +

		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Basic quick neighbour probe

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetFwdLinks(start NodePtr,exclude NodePtr[],sttype int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNeighboursByType(start,sttype);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +
		"    neighbours := ARRAY[]::Link[];\n" +
		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+
		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 
		"    END LOOP;\n" +
		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")
	
	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()
	
	// Get the forward cone / half-ball as NPtr

	qstr = "CREATE OR REPLACE FUNCTION FwdConeAsNodes(start NodePtr,sttype INT, maxdepth INT)\n"+
		"RETURNS NodePtr[] AS $fn$\n" +
		"DECLARE \n" +
		"    nextlevel NodePtr[];\n" +
		"    partlevel NodePtr[];\n" +
		"    level NodePtr[] = ARRAY[start]::NodePtr[];\n" +
		"    exclude NodePtr[] = ARRAY['(0,0)']::NodePtr[];\n" +
		"    cone NodePtr[];\n" +
		"    neigh NodePtr;\n" +
		"    frn NodePtr;\n" +
		"    counter int := 0;\n" +

		"BEGIN\n" +

		"LOOP\n" +
		"  EXIT WHEN counter = maxdepth+1;\n" +

		"  IF level IS NULL THEN\n" +
		"     RETURN cone;\n" +
		"  END IF;\n" +

		"  nextlevel := ARRAY[]::NodePtr[];\n" +

		"  FOREACH frn IN ARRAY level "+
		"  LOOP \n"+
		"     nextlevel = array_append(nextlevel,frn);\n" +
		"  END LOOP;\n" +

		"  IF nextlevel IS NULL THEN\n" +
		"     RETURN cone;\n" +
		"  END IF;\n" +

		"  FOREACH neigh IN ARRAY nextlevel LOOP \n"+
		"    IF NOT neigh = ANY(exclude) THEN\n" +
		"      cone = array_append(cone,neigh);\n" +
		"      exclude := array_append(exclude,neigh);\n" +
		"      partlevel := GetFwdNodes(neigh,exclude,sttype);\n" +
		"    END IF;" +
		"    IF partlevel IS NOT NULL THEN\n" +
		"         level = array_cat(level,partlevel);\n"+
		"    END IF;\n" +
		"  END LOOP;\n" +

		// Next, continue, foreach
		"  counter = counter + 1;\n" +
		"END LOOP;\n" +
		
		"RETURN cone; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
          /* e.g. select unnest(fwdconeaslinks) from FwdConeAsLinks('(4,1)',1,4);
                           unnest                           
             ------------------------------------------------------------
              (0,0,{},"(4,1)")
              (77,0.34,"{ ""fairy castles"", ""angel air"" }","(4,2)")
              (77,0.34,"{ ""fairy castles"", ""angel air"" }","(4,3)")
              (77,0.34,"{ ""steamy hot tubs"" }","(4,5)")
              (77,0.34,"{ ""fairy castles"", ""angel air"" }","(4,4)")
              (77,0.34,"{ ""steamy hot tubs"", ""lady gaga"" }","(4,6)")
             (6 rows)

          */

	qstr = "CREATE OR REPLACE FUNCTION FwdConeAsLinks(start NodePtr,sttype INT,maxdepth INT)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    nextlevel Link[];\n" +
		"    partlevel Link[];\n" +
		"    level Link[] = ARRAY[]::Link[];\n" +
		"    exclude NodePtr[] = ARRAY['(0,0)']::NodePtr[];\n" +
		"    cone Link[];\n" +
		"    neigh Link;\n" +
		"    frn Link;\n" +
		"    counter int := 0;\n" +

		"BEGIN\n" +

		"level := GetSingletonAsLinkArray(start);\n"+

		"LOOP\n" +
		"  EXIT WHEN counter = maxdepth+1;\n" +

		"  IF level IS NULL THEN\n" +
		"     RETURN cone;\n" +
		"  END IF;\n" +

		"  nextlevel := ARRAY[]::Link[];\n" +

		"  FOREACH frn IN ARRAY level "+
		"  LOOP \n"+
		"     nextlevel = array_append(nextlevel,frn);\n" +
		"  END LOOP;\n" +

		"  IF nextlevel IS NULL THEN\n" +
		"     RETURN cone;\n" +
		"  END IF;\n" +

		"  FOREACH neigh IN ARRAY nextlevel LOOP \n"+
		"    IF NOT neigh.Dst = ANY(exclude) THEN\n" +
		"      cone = array_append(cone,neigh);\n" +
		"      exclude := array_append(exclude,neigh.Dst);\n" +
		"      partlevel := GetFwdLinks(neigh.Dst,exclude,sttype);\n" +
		"    END IF;" +
		"    IF partlevel IS NOT NULL THEN\n" +
		"         level = array_cat(level,partlevel);\n"+
		"    END IF;\n" +
		"  END LOOP;\n" +

		"  counter = counter + 1;\n" +
		"END LOOP;\n" +

		"RETURN cone; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Orthogonal (depth first) paths from origin spreading out

	qstr = "CREATE OR REPLACE FUNCTION FwdPathsAsLinks(start NodePtr,sttype INT,maxdepth INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   hop Text;\n" +
		"   path Text;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = ARRAY[start]::NodePtr[];\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;"+

		"BEGIN\n" +

		"startlnk := GetSingletonAsLink(start);\n"+
		"path := Format('%s',startlnk::Text);\n"+
		"ret_paths := SumFwdPaths(startlnk,path,sttype,1,maxdepth,exclude);" +

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

        // select FwdPathsAsLinks('(4,1)',1,3)

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Return end of path branches as aggregated text summaries

	qstr = "CREATE OR REPLACE FUNCTION SumFwdPaths(start Link,path TEXT, sttype INT,depth int, maxdepth INT,exclude NodePtr[])\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE \n" + 
		"    fwdlinks Link[];\n" +
		"    empty Link[] = ARRAY[]::Link[];\n" +
		"    lnk Link;\n" +
		"    fwd Link;\n" +
		"    ret_paths Text;\n" +
		"    appendix Text;\n" +
		"    tot_path Text;\n"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		"fwdlinks := GetFwdLinks(start.Dst,exclude,sttype);\n" +

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL THEN" +
		          // set end of path as return val
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"         RETURN ret_paths;"+
		"      ELSE\n"+
		          // Add to the path and descend into new link
		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumFwdPaths(lnk,tot_path,sttype,depth+1,maxdepth,exclude);\n" +
		          // when we return, we reached the end of one path
		"         IF appendix IS NOT NULL THEN\n"+
	                     // append full path to list of all paths, separated by newlines
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"         ELSE"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;"+
		"      END IF;"+
		"   END IF;"+
		"END LOOP;"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Typeless cone searches

	qstr = "CREATE OR REPLACE FUNCTION AllPathsAsLinks(start NodePtr,orientation text,maxdepth INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   hop Text;\n" +
		"   path Text;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = ARRAY[start]::NodePtr[];\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;"+

		"BEGIN\n" +

		"startlnk := GetSingletonAsLink(start);\n"+
		"path := Format('%s',startlnk::Text);\n"+
		"ret_paths := SumAllPaths(startlnk,path,orientation,1,maxdepth,exclude);" +
		
		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
        // select AllPathsAsLinks('(4,1)',3)

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// SumAllPaths

	qstr = "CREATE OR REPLACE FUNCTION SumAllPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,exclude NodePtr[])\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE \n" + 
		"    fwdlinks Link[];\n" +
		"    stlinks  Link[];\n" +
		"    empty Link[] = ARRAY[]::Link[];\n" +
		"    lnk Link;\n" +
		"    fwd Link;\n" +
		"    ret_paths Text;\n" +
		"    appendix Text;\n" +
		"    tot_path Text;\n"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		// Get *All* in/out Links
		"CASE \n" +
		"   WHEN orientation = 'bwd' THEN\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   WHEN orientation = 'fwd' THEN\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   ELSE\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"END CASE;\n" +

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL THEN\n" +
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"         RETURN ret_paths;"+
		"      ELSE\n"+
		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumAllPaths(lnk,tot_path,orientation,depth+1,maxdepth,exclude);\n" +
		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"         ELSE\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;\n"+
		"      END IF;\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Check if linkpath representation is just one item

	qstr = "CREATE OR REPLACE FUNCTION empty_path(path text)\n"+
		"RETURNS boolean AS $fn$\n" +
		"BEGIN \n" +
		"   IF strpos(path,';') THEN \n" + // exact match
		"      RETURN true;\n" +
		"   END IF;\n" +
		"RETURN false;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Matching context strings with fuzzy criteria

	qstr = "CREATE OR REPLACE FUNCTION match_context(db_set text[],user_set text[])\n"+
		"RETURNS boolean AS $fn$\n" +
		"DECLARE\n" +
		"   db_ref text[];\n" +
		"   unicode text;\n" +
		"   item text;\n" +
		"   try text;\n"+
		"BEGIN \n" +
		"IF array_length(user_set,1) IS NULL THEN\n"+
		"   RETURN true;\n"+
		"END IF;\n"+

		"IF array_length(db_set,1) IS NULL THEN\n"+
		"   RETURN true;\n"+
		"END IF;\n"+

		"FOREACH item IN ARRAY db_set LOOP\n" +
		"   db_ref := array_append(db_ref,lower(unaccent(item)));\n" +
		"END LOOP;\n" +

		"FOREACH item IN ARRAY user_set LOOP\n" +
		"   IF item = 'any' OR item = '' THEN\n"+
		"     RETURN true;"+
		"   END IF;"+
		"  unicode := replace(item,'|','');\n" +
		"  FOREACH try IN ARRAY db_ref LOOP\n"+
		"     IF length(substring(try from lower(unicode))) > 3 THEN \n" + // unaccented unicode match
	        "        RETURN true;\n" +
		"     END IF;\n" +
		"  END LOOP;"+
		"END LOOP;\n" +
		"RETURN false;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Matching integer ranges

	qstr = "CREATE OR REPLACE FUNCTION match_arrows(arr int,user_set int[])\n"+
		"RETURNS boolean AS $fn$\n" +
		"BEGIN \n" +
		"   IF array_length(user_set,1) IS NULL THEN \n" + // empty arrows
                "      RETURN true;"+
		"   END IF;"+
		"   IF arr = ANY(user_set) THEN \n" + // exact match
		"      RETURN true;\n" +
		"   END IF;\n" +
		"RETURN false;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Helper to find arrows by type

	qstr = "CREATE OR REPLACE FUNCTION ArrowInList(arrow int,links Link[])\n"+
		"RETURNS boolean AS $fn$\n"+
		"DECLARE \n"+
		"   lnk Link;\n"+
		"BEGIN\n"+
		"IF links IS NULL THEN\n"+
		"   RETURN false;"+
		"END IF;"+
		"FOREACH lnk IN ARRAY links LOOP\n"+
		"  IF lnk.Arr = arrow THEN\n"+
		"     RETURN true;\n"+
		"  END IF;\n"+
		"END LOOP;"+
		"RETURN false;"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// NC version

	qstr = "CREATE OR REPLACE FUNCTION ArrowInContextList(arrow int,links Link[],context text[])\n"+
		"RETURNS boolean AS $fn$\n"+
		"DECLARE \n"+
		"   lnk Link;\n"+
		"BEGIN\n"+
		"IF links IS NULL THEN\n"+
		"   RETURN false;"+
		"END IF;"+
		"FOREACH lnk IN ARRAY links LOOP\n"+
		"  IF lnk.Arr = arrow AND match_context(lnk.Ctx::text[],context) THEN\n"+
		"     RETURN true;\n"+
		"  END IF;\n"+
		"END LOOP;"+
		"RETURN false;"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ***********************************
	// Find the start of story paths, where outgoing nodes match but no incoming
	// This means we've reached the top of a hierarchy
	// ***********************************

	// Find the node that sit's at the start/top of a causal chain

	qstr =  "CREATE OR REPLACE FUNCTION GetStoryStartNodes(arrow int,inverse int,sttype int)\n"+
		"RETURNS NodePtr[] AS $fn$\n"+
		"DECLARE \n"+
		"   retval nodeptr[] = ARRAY[]::nodeptr[];\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"   SELECT array_agg(Nptr) into retval FROM Node WHERE ArrowInList(arrow,%s) AND NOT ArrowInList(inverse,%s);\n",st,STTypeDBChannel(st),STTypeDBChannel(-st));
	}
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN retval; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	row.Close()


	// Find the node that sit's at the start/top of a causal chain

	qstr =  "CREATE OR REPLACE FUNCTION GetNCCStoryStartNodes(arrow int,inverse int,sttype int,chapter text,context text[])\n"+
		"RETURNS NodePtr[] AS $fn$\n"+
		"DECLARE \n"+
		"   retval nodeptr[] = ARRAY[]::nodeptr[];\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"   SELECT array_agg(Nptr) into retval FROM Node WHERE lower(Chap) LIKE lower(chapter) AND ArrowInContextList(arrow,%s,context) AND NOT ArrowInContextList(inverse,%s,context);\n",st,STTypeDBChannel(st),STTypeDBChannel(-st));
	}
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN retval; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)

	qstr = "CREATE OR REPLACE FUNCTION AllNCPathsAsLinks(start NodePtr,chapter text,context text[],orientation text,maxdepth INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   hop Text;\n" +
		"   path Text;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = ARRAY[start]::NodePtr[];\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;"+
		"   chp text = Format('%s%s%s','%',chapter,'%');"+
		"BEGIN\n" +
		"startlnk := GetSingletonAsLink(start);\n"+
		"path := Format('%s',startlnk::Text);"+
		"ret_paths := SumAllNCPaths(startlnk,path,orientation,1,maxdepth,chapter,context,exclude);" +

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
        // select AllNCPathsAsLinks('(1,46)','chinese','{"food","example"}','fwd',4);

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// SumAllNCPaths - a filtering version of the SumAllPaths recursive helper function, slower but more powerful

	qstr = "CREATE OR REPLACE FUNCTION SumAllNCPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,chapter text,context text[],exclude NodePtr[])\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE \n" + 
		"    fwdlinks Link[];\n" +
		"    stlinks  Link[];\n" +
		"    empty Link[] = ARRAY[]::Link[];\n" +
		"    lnk Link;\n" +
		"    fwd Link;\n" +
		"    ret_paths Text;\n" +
		"    appendix Text;\n" +
		"    tot_path Text;\n"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		// We order the link types to respect the geometry of the temporal links
		// so that (then) will always come last for visual sensemaking

		// Get *All* in/out Links
		"CASE \n" +
		"   WHEN orientation = 'bwd' THEN\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,-1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,-2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,-3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   WHEN orientation = 'fwd' THEN\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   ELSE\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,-1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,-2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,-3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,context,exclude,1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +

		"END CASE;\n" +

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL THEN\n" +
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"      ELSE\n"+
		"         IF context is not NULL AND NOT match_context(lnk.Ctx::text[],context::text[]) THEN\n"+
                "            CONTINUE;\n"+
                "         END IF;\n"+

		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumAllNCPaths(lnk,tot_path,orientation,depth+1,maxdepth,chapter,context,exclude);\n" +

		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"         ELSE\n"+
		"            ret_paths := tot_path;\n"+
		"         END IF;\n"+
		"      END IF;\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)
        // with a start set of more than one node

	qstr = "CREATE OR REPLACE FUNCTION AllSuperNCPathsAsLinks(start NodePtr[],chapter text,context text[],orientation text,maxdepth INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   root Text;\n" +
		"   path Text;\n"+
		"   node NodePtr;"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = start;\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;"+
		"   chp text = Format('%s%s%s','%',chapter,'%');"+
		"BEGIN\n" +

		// Aggregate array of starting set
		"FOREACH node IN ARRAY start LOOP\n"+
		"   startlnk := GetSingletonAsLink(node);\n"+
		"   path := Format('%s',startlnk::Text);"+
		"   root := SumAllNCPaths(startlnk,path,orientation,1,maxdepth,chapter,context,exclude);" +
		"ret_paths := Format('%s\n%s',ret_paths,root);\n"+
		"END LOOP;"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
        // select AllNCPathsAsLinks('(1,46)','chinese','{"food","example"}','fwd',4);

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

        // An NC/C filtering version of the neighbour scan

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetNCFwdLinks(start NodePtr,chapter text,context text[],exclude NodePtr[],sttype int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNCNeighboursByType(start,chapter,sttype);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +
		"    neighbours := ARRAY[]::Link[];\n" +
		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+

                "      IF context is not NULL AND NOT match_context(lnk.Ctx::text[],context::text[]) THEN\n"+
                "         CONTINUE;\n"+
                "      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 
		"    END LOOP;\n" +
		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")
	
	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

        // This one includes an NCC chapter and context filter so slower! 

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetNCCLinks(start NodePtr,exclude NodePtr[],sttype int,chapter text,context text[])\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks =GetNCNeighboursByType(start,chapter,sttype);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +
		"    neighbours := ARRAY[]::Link[];\n" +
		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+
                "      IF context is not NULL AND NOT match_context(lnk.Ctx,context) THEN\n"+
                "        CONTINUE;\n"+
                "      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 
		"    END LOOP;\n" +
		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")


        // This one includes an NC chapter filter

	qstr = "CREATE OR REPLACE FUNCTION GetNCNeighboursByType(start NodePtr, chapter text,sttype int)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    fwdlinks Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,Array[]::text[],(0,0));\n"+
                "    chp text = Format('%s%s%s','%',chapter,'%');"+
		"BEGIN\n"+

		"   CASE sttype \n"
	
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"     SELECT %s INTO fwdlinks FROM Node WHERE Nptr=start AND lower(Chap) LIKE lower(chp);\n",st,STTypeDBChannel(st));
	}

	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +

		"    RETURN fwdlinks; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

        // elect GetNCNeighboursByType('(1,116)','chinese',-1);


	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

}

// **************************************************************************
// Retrieve
// **************************************************************************

func GetDBChaptersMatchingName(ctx PoSST,src string) []string {

	var qstr string

	remove_accents,stripped := IsBracketedSearchTerm(src)

	if remove_accents {
		search := "%"+stripped+"%"
		qstr = fmt.Sprintf("SELECT DISTINCT Chap FROM Node WHERE lower(unaccent(Chap)) LIKE lower('%s')",search)
	} else {
		search := "%"+src+"%"
		qstr = fmt.Sprintf("SELECT DISTINCT Chap FROM Node WHERE lower(Chap) LIKE lower('%s')",search)
	}

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBChaptersMatchingName",err)
	}

	var whole string
	var retval []string

	for row.Next() {		
		err = row.Scan(&whole)
		retval = append(retval,whole)
	}

	sort.Strings(retval)
	row.Close()
	return retval
}

// **************************************************************************

func GetDBContextsMatchingName(ctx PoSST,src string) []string {

	var qstr string

	remove_accents,stripped := IsBracketedSearchTerm(src)

	if remove_accents {
		search := stripped
		qstr = fmt.Sprintf("SELECT DISTINCT Ctx FROM NodeArrowNode WHERE match_context(Ctx,'{%s}')",search)
	} else {
		search := src
		qstr = fmt.Sprintf("SELECT DISTINCT Ctx FROM NodeArrowNode WHERE match_context(Ctx,'{%s}')",search)
	}

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBChaptersMatchingName",err)
	}

	var whole string
	var retval []string
	var idemp = make(map[string]int)

	for row.Next() {		
		err = row.Scan(&whole)
		a := ParseSQLArrayString(whole)
		for i := range a {
			idemp[a[i]]++
		}
	}

	for s := range idemp {
		retval = append(retval,s)
	}

	row.Close()

	sort.Strings(retval)
	return retval

}

// **************************************************************************

func GetDBNodePtrMatchingName(ctx PoSST,src,chap string) []NodePtr {

	var qstr string

	if src == "" || src == "empty" {
		return nil
	}
 
	remove_accents,stripped := IsBracketedSearchTerm(src)

	if remove_accents {
		search := "%"+stripped+"%"
		qstr = fmt.Sprintf("select NPtr from Node where lower(unaccent(S)) LIKE lower('%s')",search)
	} else {
		search := "%"+src+"%"
		qstr = fmt.Sprintf("select NPtr from Node where lower(S) LIKE lower('%s')",search)
	}

	if chap != "any" && chap != "" {

		remove_accents,stripped := IsBracketedSearchTerm(chap)
		if remove_accents {
			chapter := "%"+stripped+"%"
			qstr += fmt.Sprintf(" AND lower(unaccent(chap)) LIKE '%s'",chapter)
		} else {
			chapter := "%"+chap+"%"
			qstr += fmt.Sprintf(" AND lower(chap) LIKE '%s'",chapter)
		}
	}

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetNodePtrMatchingName Failed",err)
	}

	var whole string
	var n NodePtr
	var retval []NodePtr

	for row.Next() {		
		err = row.Scan(&whole)
		fmt.Sscanf(whole,"(%d,%d)",&n.Class,&n.CPtr)
		retval = append(retval,n)
	}

	row.Close()
	return retval

}

// **************************************************************************

func GetDBNodePtrMatching(ctx PoSST,nm,chap string,cn []string,arrow []ArrowPtr) []NodePtr {

	// Match name, context, chapter

	var chap_col, nm_col string
	var context string

	remove_name_accents,nm_stripped := IsBracketedSearchTerm(nm)

	if remove_name_accents {
		nm_search := "%"+nm_stripped+"%"
		nm_col = fmt.Sprintf("AND lower(unaccent(S)) LIKE lower('%s')",nm_search)
	} else {
		nm_search := "%"+nm+"%"
		nm_col = fmt.Sprintf("AND lower(S) LIKE lower('%s')",nm_search)
	}

	if chap != "any" && chap != "" {

		remove_chap_accents,chap_stripped := IsBracketedSearchTerm(chap)

		if remove_chap_accents {
			chap_search := "%"+chap_stripped+"%"
			chap_col = fmt.Sprintf("AND lower(unaccent(chap)) LIKE lower('%s')",chap_search)
		} else {
			chap_search := "%"+chap+"%"
			chap_col = fmt.Sprintf("AND lower(chap) LIKE lower('%s')",chap_search)
		}
	}

	_,cn_stripped := IsBracketedSearchList(cn)
	context = FormatSQLStringArray(cn_stripped)

	arrows := FormatSQLIntArray(Arrow2Int(arrow))

	qstr := fmt.Sprintf("WITH matching_nodes AS "+
		"  (SELECT NFrom,ctx,match_context(ctx,%s) AS match,match_arrows(Arr,%s) AS matcha FROM NodeArrowNode)"+
		"     SELECT DISTINCT nfrom FROM matching_nodes "+
		"      JOIN Node ON nptr=nfrom WHERE match=true AND matcha=true %s %s",
		context,arrows,nm_col,chap_col)

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetNodePtrMatchingNCC Failed",err,qstr)
	}

	var whole string
	var n NodePtr
	var retval []NodePtr

	for row.Next() {		
		err = row.Scan(&whole)
		fmt.Sscanf(whole,"(%d,%d)",&n.Class,&n.CPtr)
		retval = append(retval,n)
	}

	row.Close()
	return retval

}

// **************************************************************************

func GetDBNodeByNodePtr(ctx PoSST,db_nptr NodePtr) Node {

	im_nptr,cached := NODE_CACHE[db_nptr]

	if cached {
		return GetNodeFromPtr(im_nptr)
	}

	// This ony works if we insert non-null arrays in initialization
	cols := I_MEXPR+","+I_MCONT+","+I_MLEAD+","+I_NEAR +","+I_PLEAD+","+I_PCONT+","+I_PEXPR
	qstr := fmt.Sprintf("select L,S,Chap,%s from Node where NPtr='(%d,%d)'::NodePtr",cols,db_nptr.Class,db_nptr.CPtr)

	row, err := ctx.DB.Query(qstr)

	var n Node
	var count int = 0

	if err != nil {
		fmt.Println("GetDBNodeByNodePointer Failed:",err)
		return n
	}

	var whole [ST_TOP]string

	// NB, there seems to be a bug in the SQL package, which cannot always populate the links, so try not to
	//     rely on this and work around when needed using GetEntireCone(any,2..) separately

	for row.Next() {
		err = row.Scan(&n.L,&n.S,&n.Chap,&whole[0],&whole[1],&whole[2],&whole[3],&whole[4],&whole[5],&whole[6])

		for i := 0; i < ST_TOP; i++ {
			n.I[i] = ParseLinkArray(whole[i])
		}
		count++
	}

	if count > 1 {
		fmt.Println("GetDBNodeByNodePtr returned too many matches (multi-model conflict?):",count,"for ptr",db_nptr)
		os.Exit(-1)
	}

	row.Close()

	if !cached {
		CacheNode(n)
	}

	n.NPtr = db_nptr
	return n
}

// **************************************************************************

func GetDBArrowsWithArrowName(ctx PoSST,s string) ArrowPtr {

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(ctx)
	}

	for a := range ARROW_DIRECTORY {
		if s == ARROW_DIRECTORY[a].Long || s == ARROW_DIRECTORY[a].Short {
			return ARROW_DIRECTORY[a].Ptr
		}
	}

	fmt.Println("No such arrow found in database:",s)
	os.Exit(-1)
	return -1
}

// **************************************************************************

func GetDBArrowsMatchingArrowName(ctx PoSST,s string) []ArrowPtr {

	var list []ArrowPtr

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(ctx)
	}

	for a := range ARROW_DIRECTORY {
		if SimilarString(s,ARROW_DIRECTORY[a].Long) || SimilarString(s,ARROW_DIRECTORY[a].Short) {
			list = append(list,ARROW_DIRECTORY[a].Ptr)
		}
	}

	return list
}

// **************************************************************************

func GetDBNodeArrowNodeMatchingArrowPtrs(ctx PoSST,chap string,cn []string,arrows []ArrowPtr) []NodeArrowNode {

	var intarrows []int

	for i := range arrows {
		intarrows = append(intarrows,int(arrows[i]))
	}

	qstr := fmt.Sprintf("SELECT NFrom,STType,Arr,Wgt,Ctx,NTo FROM NodeArrowNode where Arr=ANY(%s::int[])",FormatSQLIntArray(intarrows))

	if cn != nil {
		context := FormatSQLStringArray(cn)
		chapter := "%"+chap+"%"
		
		qstr = fmt.Sprintf("WITH matching_rel AS "+
			" (SELECT NFrom,STType,Arr,Wgt,Ctx,NTo,match_context(ctx,%s) AS match FROM NodeArrowNode)"+
			"   SELECT DISTINCT NFrom,STType,Arr,Wgt,Ctx,NTo FROM matching_rel "+
			"    JOIN Node ON nptr=nfrom WHERE match=true AND lower(chap) LIKE lower('%s')",context,chapter)	
	}

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("GetDBNodeArrowNodeMatchingArrowPtrs Failed:",err,qstr)
	}

	var from_node string
	var to_node string
	var actx string
	var st,arr int
	var wgt float64

	var nfr,nto NodePtr
	var nan NodeArrowNode
	var nanlist []NodeArrowNode

	for row.Next() {		
		err = row.Scan(&from_node,&st,&arr,&wgt,&actx,&to_node)

		fmt.Sscanf(from_node,"(%d,%d)",&nfr.Class,&nfr.CPtr)
		fmt.Sscanf(to_node,"(%d,%d)",&nto.Class,&nto.CPtr)

		nan.NFrom = nfr
		nan.STType = st
		nan.Arr = ArrowPtr(arr)
		nan.Wgt = wgt
		nan.Ctx = ParseSQLArrayString(actx)
		nan.NTo = nto

		nanlist = append(nanlist,nan)

	}

	row.Close()

	return nanlist
}

// **************************************************************************

func GetDBNodeContextsMatchingArrow(ctx PoSST,searchtext string,chap string,cn []string,arrow []ArrowPtr,page int) []QNodePtr {
	var qstr string

	context := FormatSQLStringArray(cn)
	chapter := "%"+chap+"%"
	arrows := FormatSQLIntArray(Arrow2Int(arrow))

	const hits_per_page = 30
	offset := (page-1) * hits_per_page;

	// sufficient to search NFrom to get all nodes in context, as +/- relations complete
	
	qstr = fmt.Sprintf("WITH matching_nodes AS \n"+
		" (SELECT DISTINCT NFrom,Arr,Ctx,match_context(Ctx,%s) AS matchc,match_arrows(Arr,%s) AS matcha FROM NodeArrowNode)\n"+
		"   SELECT DISTINCT NFrom,Ctx,Chap FROM matching_nodes \n"+
		"    JOIN Node ON nptr=nfrom WHERE matchc=true AND matcha=true AND lower(Chap) LIKE lower('%s') ORDER BY Ctx,NFrom DESC OFFSET %d LIMIT %d",context,arrows,chapter,offset,hits_per_page)

	row, err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetDBNodeArrowNodeByContext Failed:",err,qstr)
	}

	var return_value []QNodePtr

	var qptr QNodePtr
	var nptr NodePtr
	var nctx string
	var nchap string
	var nptrs string

	for row.Next() {		

		nctx = ""
		nchap = ""
		err = row.Scan(&nptrs,&nctx,&nchap)
		fmt.Sscanf(nptrs,"(%d,%d)",&nptr.Class,&nptr.CPtr)
		qptr.NPtr = nptr
		qptr.Chapter = nchap

		if nctx == "" {
			nctx = "(no context)"
		}

		qptr.Context = nctx

		return_value = append(return_value,qptr)
	}

	row.Close()
	return return_value
}

// **************************************************************************

func GetNodesStartingStoriesForArrow(ctx PoSST,arrow string) []NodePtr {

	// Find the head / starting node matching an arrow sequence.
	// It has outgoing (+sttype) but not incoming (-sttype) arrow

	var matches []NodePtr

	arrowptr := GetDBArrowsWithArrowName(ctx,arrow)

	sttype := STIndexToSTType(ARROW_DIRECTORY[arrowptr].STAindex)

	qstr := fmt.Sprintf("select GetStoryStartNodes(%d,%d,%d)",arrowptr,INVERSE_ARROWS[arrowptr],sttype)
		
	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("GetNodesStartingStoriesForArrow failed\n",qstr,err)
		return nil
	}
	
	var nptrstring string
	
	for row.Next() {		
		err = row.Scan(&nptrstring)
		matches = ParseSQLNPtrArray(nptrstring)
	}
	
	row.Close()

	return matches
}

// **************************************************************************

func GetNCCNodesStartingStoriesForArrow(ctx PoSST,arrow string,chapter string,context []string) []NodePtr {

	// Filtered version of function
	// Find the head / starting node matching an arrow sequence.
	// It has outgoing (+sttype) but not incoming (-sttype) arrow

	var matches []NodePtr

	arrowptr := GetDBArrowsWithArrowName(ctx,arrow)

	sttype := STIndexToSTType(ARROW_DIRECTORY[arrowptr].STAindex)

	chp := "%"+chapter+"%"
	cntx := FormatSQLStringArray(context)
	
	qstr := fmt.Sprintf("select GetNCCStoryStartNodes(%d,%d,%d,'%s',%s)",arrowptr,INVERSE_ARROWS[arrowptr],sttype,chp,cntx)
	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("GetNodesNCCStartingStoriesForArrow failed\n",qstr,err)
		return nil
	}
	
	var nptrstring string
	
	for row.Next() {		
		err = row.Scan(&nptrstring)
		
		matches = ParseSQLNPtrArray(nptrstring)
	}
	
	row.Close()
	return matches
}

// **************************************************************************

func GetDBArrowByName(ctx PoSST,name string) ArrowPtr {

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(ctx)
	}

	ptr, ok := ARROW_SHORT_DIR[name]
	
	// If not, then check longname
	
	if !ok {
		ptr, ok = ARROW_LONG_DIR[name]
		
		if !ok {
			ptr, ok = ARROW_SHORT_DIR[name]
			
			// If not, then check longname
			
			if !ok {
				ptr, ok = ARROW_LONG_DIR[name]
				fmt.Println(ERR_NO_SUCH_ARROW,"("+name+") - no arrows defined in database yet?")
			}
		}
	}

	return ptr
}

// **************************************************************************

func GetDBArrowByPtr(ctx PoSST,arrowptr ArrowPtr) ArrowDirectory {

	if ARROW_DIRECTORY_TOP > 0 {
		a := ARROW_DIRECTORY[arrowptr]
		return a
	}

	DownloadArrowsFromDB(ctx)

	if len(ARROW_DIRECTORY) < int(arrowptr) {
		fmt.Println(ERR_NO_SUCH_ARROW,"(",arrowptr,")")
		os.Exit(-1)
	}

	return ARROW_DIRECTORY[arrowptr]

}

// **************************************************************************

func GetDBPageMap(ctx PoSST,chap string,cn []string,page int) []PageMap {

	var qstr string

	context := FormatSQLStringArray(cn)
	chapter := "%"+chap+"%"

	const hits_per_page = 30
	offset := (page-1) * hits_per_page;

	qstr = fmt.Sprintf("SELECT DISTINCT Chap,Ctx,Line,Path FROM PageMap\n"+
		"WHERE match_context(Ctx,%s)=true AND lower(Chap) LIKE lower('%s') ORDER BY Line OFFSET %d LIMIT %d",
		context,chapter,offset,hits_per_page)

	row, err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetDBPageMap Failed:",err,qstr)
	}

	var path string
	var pagemap []PageMap
	var line int
	for row.Next() {		

		var event PageMap
		err = row.Scan(&chap,&context,&line,&path)

		if err != nil {
			fmt.Println("Error reading GetDBPageMap",err)
		}

		event.Path = ParseMapLinkArray(path)

		event.Chapter = chap
		event.Context = ParseSQLArrayString(context)
		pagemap = append(pagemap,event)
	}

	row.Close()
	return pagemap
}

// **************************************************************************
// Retrieval
// **************************************************************************

func CacheNode(n Node) {

	NODE_CACHE[n.NPtr] = AppendTextToDirectory(n,RunErr)
}

// **************************************************************************

func DownloadArrowsFromDB(ctx PoSST) {

	// These must be ordered to match in-memory array

	qstr := fmt.Sprintf("SELECT STAindex,Long,Short,ArrPtr FROM ArrowDirectory ORDER BY ArrPtr")

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY Download Arrows Failed",err)
	}

	var staidx int
	var long string
	var short string
	var ptr ArrowPtr
	var ad ArrowDirectory

	for row.Next() {		
		err = row.Scan(&staidx,&long,&short,&ptr)
		ad.STAindex = staidx
		ad.Long = long
		ad.Short = short
		ad.Ptr = ptr

		ARROW_DIRECTORY = append(ARROW_DIRECTORY,ad)
		ARROW_SHORT_DIR[short] = ARROW_DIRECTORY_TOP
		ARROW_LONG_DIR[long] = ARROW_DIRECTORY_TOP

		if ad.Ptr != ARROW_DIRECTORY_TOP {
			fmt.Println(ERR_MEMORY_DB_ARROW_MISMATCH,ad,ad.Ptr,ARROW_DIRECTORY_TOP)
			os.Exit(-1)
		}

		ARROW_DIRECTORY_TOP++
	}

	row.Close()

	// Get Inverses

	qstr = fmt.Sprintf("SELECT Plus,Minus FROM ArrowInverses ORDER BY Plus")

	row, err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY Download Inverses Failed",err)
	}

	var plus,minus ArrowPtr

	for row.Next() {		

		err = row.Scan(&plus,&minus)

		if err != nil {
			fmt.Println("QUERY Download Arrows Failed",err)
		}

		INVERSE_ARROWS[plus] = minus
	}
}

// **************************************************************************

func GetFwdConeAsNodes(ctx PoSST, start NodePtr, sttype,depth int) []NodePtr {

	qstr := fmt.Sprintf("select unnest(fwdconeasnodes) from FwdConeAsNodes('(%d,%d)',%d,%d);",start.Class,start.CPtr,sttype,depth)

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdConeAsNodes Failed",err)
	}

	var whole string
	var n NodePtr
	var retval []NodePtr

	for row.Next() {		
		err = row.Scan(&whole)
		fmt.Sscanf(whole,"(%d,%d)",&n.Class,&n.CPtr)
		retval = append(retval,n)
	}

	row.Close()
	return retval
}

// **************************************************************************

func GetFwdConeAsLinks(ctx PoSST, start NodePtr, sttype,depth int) []Link {

	// This function may be misleading as it doesn't respect paths

	qstr := fmt.Sprintf("select unnest(fwdconeaslinks) from FwdConeAsLinks('(%d,%d)',%d,%d);",start.Class,start.CPtr,sttype,depth)

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdConeAsLinks Failed",err)
	}

	var whole string
	var retval []Link

	for row.Next() {		
		err = row.Scan(&whole)
		l := ParseSQLLinkString(whole)
		retval = append(retval,l)
	}

	row.Close()

	return retval
}

// **************************************************************************

func GetFwdPathsAsLinks(ctx PoSST, start NodePtr, sttype,depth int) ([][]Link,int) {

	qstr := fmt.Sprintf("select FwdPathsAsLinks from FwdPathsAsLinks('(%d,%d)',%d,%d);",start.Class,start.CPtr,sttype,depth)

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdPathsAsLinks Failed",err)
	}

	var whole string
	var retval [][]Link

	for row.Next() {		
		err = row.Scan(&whole)
		retval = ParseLinkPath(whole)
	}

	row.Close()
	return retval,len(retval)
}

// **************************************************************************

func GetEntireConePathsAsLinks(ctx PoSST,orientation string,start NodePtr,depth int) ([][]Link,int) {

	// orientation should be "fwd" or "bwd" else "both"

	qstr := fmt.Sprintf("select AllPathsAsLinks from AllPathsAsLinks('(%d,%d)','%s',%d);",
		start.Class,start.CPtr,orientation,depth)

	row, err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to AllPathsAsLinks Failed",err,qstr)
	}

	var whole string
	var retval [][]Link

	for row.Next() {		
		err = row.Scan(&whole)
		retval = ParseLinkPath(whole)
	}

	row.Close()

	sort.Slice(retval, func(i,j int) bool {
		return len(retval[i]) < len(retval[j])
	})

	return retval,len(retval)
}

// **************************************************************************

func GetEntireNCConePathsAsLinks(ctx PoSST,orientation string,start NodePtr,depth int,chapter string,context []string) ([][]Link,int) {

	// orientation should be "fwd" or "bwd" else "both"

	qstr := fmt.Sprintf("select AllNCPathsAsLinks from AllNCPathsAsLinks('(%d,%d)','%s',%s,'%s',%d);",
		start.Class,start.CPtr,chapter,FormatSQLStringArray(context),orientation,depth)

	row, err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to AllNCPathsAsLinks Failed",err,qstr)
	}

	var whole string
	var retval [][]Link

	for row.Next() {		
		err = row.Scan(&whole)
		retval = ParseLinkPath(whole)
	}

	row.Close()
	return retval,len(retval)
}

// **************************************************************************

func GetEntireNCSuperConePathsAsLinks(ctx PoSST,orientation string,start []NodePtr,depth int,chapter string,context []string) ([][]Link,int) {
	// orientation should be "fwd" or "bwd" else "both"

	qstr := fmt.Sprintf("select AllSuperNCPathsAsLinks(%s,'%s',%s,'%s',%d);",FormatSQLNodePtrArray(start),
		chapter,FormatSQLStringArray(context),orientation,depth)

	row, err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to AllSuperNCPathsAsLinks Failed",err,qstr)
		os.Exit(-1)
	}

	var whole string
	var retval [][]Link

	for row.Next() {		
		err = row.Scan(&whole)
		retval = ParseLinkPath(whole)
	}

	row.Close()
	return retval,len(retval)
}

// **************************************************************************
// Transition/path integral matrix
// **************************************************************************

func GetPathsAndSymmetries(ctx PoSST,start_set,end_set []NodePtr,chapter string,context []string,maxdepth int) [][]Link {

	var left_paths, right_paths [][]Link
	var ldepth,rdepth int = 1,1
	var Lnum,Rnum int
	var solutions [][]Link

	if start_set == nil || end_set == nil {
		return nil
	}

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		left_paths,Lnum = GetEntireNCSuperConePathsAsLinks(ctx,"fwd",start_set,ldepth,chapter,context)
		right_paths,Rnum = GetEntireNCSuperConePathsAsLinks(ctx,"bwd",end_set,rdepth,chapter,context)		
		solutions,_ = WaveFrontsOverlap(ctx,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		if len(solutions) > 0 {
			break
		}

		if turn % 2 == 0 {
			ldepth++
		} else {
			rdepth++
		}
	}

	// Calculate the supernode layer sets S[path][depth], factoring process symmetries

	return solutions
}

// **************************************************************************

func GetPathTransverseSuperNodes(ctx PoSST,solutions [][]Link,maxdepth int) [][]NodePtr {

	var supernodes [][]NodePtr

	for depth := 0; depth < maxdepth; depth++ {

		for p_i := 0; p_i < len(solutions); p_i++ {

			if depth == len(solutions[p_i])-1 {
				supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_i][depth].Dst)
			}

			if depth > len(solutions[p_i])-1 {
				continue
			}

			supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_i][depth].Dst)

			for p_j := p_i+1; p_j < len(solutions); p_j++ {

				if depth < 1 || depth > len(solutions[p_j])-2 {
					break
				}

				if solutions[p_i][depth-1].Dst == solutions[p_j][depth-1].Dst && 
				   solutions[p_i][depth+1].Dst == solutions[p_j][depth+1].Dst {
					   supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_j][depth].Dst)
				}
			}
		}		
	}

	return supernodes	
}

// **********************************************************

func WaveFrontsOverlap(ctx PoSST,left_paths,right_paths [][]Link,Lnum,Rnum,ldepth,rdepth int) ([][]Link,[][]Link) {

	// The wave front consists of Lnum and Rnum points left_paths[len()-1].
	// Any of the

	var solutions [][]Link
	var loops [][]Link

	// Start expanding the waves from left and right, one step at a time, alternately

	leftfront := WaveFront(left_paths,Lnum)
	rightfront := WaveFront(right_paths,Rnum)

	incidence := NodesOverlap(ctx,leftfront,rightfront)

	for lp := range incidence {

		for alternative := range incidence[lp] {

			rp := incidence[lp][alternative]

			var LRsplice []Link		
			
			LRsplice = LeftJoin(LRsplice,left_paths[lp])
			adjoint := AdjointLinkPath(right_paths[rp])
			LRsplice = RightComplementJoin(LRsplice,adjoint)

			if IsDAG(LRsplice) {
				solutions = append(solutions,LRsplice)
			} else {
				loops = append(loops,LRsplice)
			}
		}
	}

	return solutions,loops
}

// **********************************************************

func WaveFront(path [][]Link,num int) []NodePtr {

	// assemble the cross cutting nodeptrs of the wavefronts

	var front []NodePtr

	for l := 0; l < num; l++ {
		front = append(front,path[l][len(path[l])-1].Dst)
	}

	return front
}

// **********************************************************

func NodesOverlap(ctx PoSST,left,right []NodePtr) map[int][]int {

	var LRsplice = make(map[int][]int)

	// Return coordinate pairs of partial paths to splice

	for l := 0; l < len(left); l++ {
		for r := 0; r < len(right); r++ {
			if left[l] == right[r] {
				LRsplice[l] = append(LRsplice[l],r)
			}
		}
	}

	return LRsplice
}

// **********************************************************

func LeftJoin(LRsplice,seq []Link) []Link {

	for i := 0; i < len(seq); i++ {

		LRsplice = append(LRsplice,seq[i])
	}

	return LRsplice
}

// **********************************************************

func RightComplementJoin(LRsplice,adjoint []Link) []Link {

	// len(seq)-1 matches the last node of right join
	// when we invert, links and destinations are shifted

	for j := 1; j < len(adjoint); j++ {
		LRsplice = append(LRsplice,adjoint[j])
	}

	return LRsplice
}

// **********************************************************

func IsDAG(seq []Link) bool {

	var freq = make(map[NodePtr]int)

	for i := range seq {
		freq[seq[i].Dst]++
	}

	for n := range freq {
		if freq[n] > 1 {
			return false
		}
	}

	return true
}

// **********************************************************

func Together(matroid [][]NodePtr,n1 NodePtr,n2 NodePtr) [][]NodePtr {

        // matroid [snode][member]

	if len(matroid) == 0 {
		var newsuper []NodePtr
		newsuper = append(newsuper,n1)
		if n1 != n2 {
			newsuper = append(newsuper,n2)
		}
		matroid = append(matroid,newsuper)
		return matroid
	}

	for i := range matroid {
		if InNodeSet(matroid[i],n1) || InNodeSet(matroid[i],n2) {
			matroid[i] = IdempAddNodePtr(matroid[i],n1)
			matroid[i] = IdempAddNodePtr(matroid[i],n2)
			return matroid
		}
	}

	var newsuper []NodePtr

	newsuper = IdempAddNodePtr(newsuper,n1)
	newsuper = IdempAddNodePtr(newsuper,n2)
	matroid = append(matroid,newsuper)

	return matroid
}

// **********************************************************

func IdempAddNodePtr(set []NodePtr, n NodePtr) []NodePtr {

	if !InNodeSet(set,n) {
		set = append(set,n)
	}
	return set
}

// **********************************************************

func InNodeSet(list []NodePtr,node NodePtr) bool {

	for n := range list {
		if list[n] == node {
			return true
		}
	}
	return false
}

// **************************************************************************
// Path tools
// **************************************************************************

func AdjointLinkPath(LL []Link) []Link {

	var adjoint []Link

	// len(seq)-1 matches the last node of right join
	// when we invert, links and destinations are shifted

	var prevarrow ArrowPtr = INVERSE_ARROWS[0]

	for j := len(LL)-1; j >= 0; j-- {

		var lnk Link = LL[j]
		lnk.Arr = INVERSE_ARROWS[prevarrow]
		adjoint = append(adjoint,lnk)
		prevarrow = LL[j].Arr
	}

	return adjoint
}

// **************************************************************************

func NextLinkArrow(ctx PoSST,path []Link,arrows []ArrowPtr) string {

	var rstring string

	if len(path) > 1 {

		for l := 1; l < len(path); l++ {

			if !MatchArrows(arrows,path[l].Arr) {
				break
			}

			nextnode := GetDBNodeByNodePtr(ctx,path[l].Dst)
			
			arr := GetDBArrowByPtr(ctx,path[l].Arr)
			
			if l < len(path) {
				rstring += fmt.Sprint("  -(",arr.Long,")->  ")
			}
			
			rstring += fmt.Sprint(nextnode.S)
		}
	}

	return rstring
}

// **************************************************************************

func GetNodeOrbit(ctx PoSST,nptr NodePtr,exclude_vector string) [ST_TOP][]Orbit {

	// Start with properties of node, within orbit

	const probe_radius = 3

	// radius = 0 is the starting node

	sweep,_ := GetEntireConePathsAsLinks(ctx,"any",nptr,probe_radius)

	var notes [ST_TOP][]Orbit

	// Organize by the leading nearest-neighbour by vector/link type

	for stindex := 0; stindex < ST_TOP; stindex++ {

		// Sweep different radial paths

		for angle := range sweep {

			// len(sweep[angle]) is the length of the probe path at angle

			if sweep[angle] != nil && len(sweep[angle]) > 1 {

				const nearest_satellite = 1
				start := sweep[angle][nearest_satellite]

				arrow := GetDBArrowByPtr(ctx,start.Arr)

				if arrow.STAindex == stindex {
					txt := GetDBNodeByNodePtr(ctx,start.Dst)
					var nt Orbit
					nt.Arrow = arrow.Long
                                        nt.STindex = arrow.STAindex
					nt.Dst = start.Dst
					nt.Text = txt.S
					nt.Radius = 1
					if arrow.Long == exclude_vector || arrow.Short == exclude_vector {
						continue
					}

					notes[stindex] = IdempAddNote(notes[stindex],nt)

					// are there more satellites at this angle?

					for depth := 2; depth < probe_radius && depth < len(sweep[angle]); depth++ {

						arprev := STIndexToSTType(arrow.STAindex)
						next := sweep[angle][depth]
						arrow = GetDBArrowByPtr(ctx,next.Arr)
						subtxt := GetDBNodeByNodePtr(ctx,next.Dst)

						if arrow.Long == exclude_vector || arrow.Short == exclude_vector {
							break
						}

						nt.Arrow = arrow.Long
						nt.STindex = arrow.STAindex
						nt.Dst = next.Dst
						nt.Ctx = Array2Str(next.Ctx)
						nt.Text = subtxt.S
						nt.Radius = depth

						arthis := STIndexToSTType(arrow.STAindex)
						// No backtracking
						if arthis != -arprev {	
							notes[stindex] = IdempAddNote(notes[stindex],nt)
							arprev = arthis
						}
					}
				}
			}
		}
	}
	return notes
}

// **************************************************************************

func IdempAddNote(list []Orbit, item Orbit) []Orbit {

	for o := range list {
		if list[o].Dst == item.Dst && list[o].Arrow == item.Arrow &&
			list[o].Text == item.Text {
			return list
		}
	}

	return append(list,item)
}

// **************************************************************************
// Axial paths
// **************************************************************************

func GetSequenceContainers(ctx PoSST,arrname string,search,chapter string,context []string) []Story {

	var stories []Story

	if arrname == "" {
		arrname = "then"
	}

	arrowptr := GetDBArrowsWithArrowName(ctx,arrname)

	openings := GetNCCNodesStartingStoriesForArrow(ctx,arrname,chapter,context)
	
	if len(openings) > 1 {

		for nptr := range openings {
			var story Story
			node := GetDBNodeByNodePtr(ctx,openings[nptr])
			story.Arrow = node.Chap
			story.Text = node.S
			stories = append(stories,story)
		}
		// If Axis is null, then this is just the toc
		return stories
	}

	// return one story
	
	for nptr := range openings {

		var story Story
		
		node := GetDBNodeByNodePtr(ctx,openings[nptr])
		orbit := GetNodeOrbit(ctx,openings[nptr],arrname)

		container := orbit[ST_ZERO-CONTAINS] // Does the sequence have a container?
		
		if container != nil {
			story.ContainNPtr = container[0].Dst // generalize..tbd
			story.Text = container[0].Text
			story.Arrow = container[0].Arrow
		} else {
			var none NodePtr
			story.ContainNPtr = none // generalize..tbd
			story.Text = "(Story without an external title container)"
			story.Arrow = "  -- title may be included in the _sequence_, consider moving title (contains) ::_sequence_::"
		}

		if OrbitMatching(ctx,node,orbit,search) {

			axis := GetLongestAxialPath(ctx,openings[nptr],arrowptr)

			for lnk := 0; lnk < len(axis); lnk++ {

				// Now add the orbit at this node, not including the axis
				var ne NodeEvent
				nd := GetDBNodeByNodePtr(ctx,axis[lnk].Dst)
				ne.Text = nd.S
				ne.L = nd.L
				ne.NPtr = axis[lnk].Dst
				ne.Orbits = GetNodeOrbit(ctx,axis[lnk].Dst,arrname)

				story.Axis = append(story.Axis,ne)
			}
		}

		if story.Axis != nil {
			stories = append(stories,story)
		}
		
	}

	return stories
}

// **************************************************************************

func OrbitMatching(ctx PoSST,node Node,orbit [ST_TOP][]Orbit,search string) bool {

	// Check whether the search string occurs within the near orbit of the node

	if strings.Contains(node.S,search) || strings.Contains(search,node.S) {
		return true
	}

	for st := 0; st < ST_TOP; st++ {
		for r := range orbit[st] {
			if strings.Contains(orbit[st][r].Text,search) || 
			strings.Contains(search,orbit[st][r].Text) {
				return true
			}
		}
	}

	return false
}

// **************************************************************************

func GetLongestAxialPath(ctx PoSST,nptr NodePtr,arrowptr ArrowPtr) []Link {

	var max int = 1
	const maxdepth = 100 // Hard limit on story length, what?

	sttype := STIndexToSTType(ARROW_DIRECTORY[arrowptr].STAindex)
	paths,dim := GetFwdPathsAsLinks(ctx,nptr,sttype,maxdepth)

	for pth := 0; pth < dim; pth++ {

		var depth int
		paths[pth],depth = TruncatePathsByArrow(paths[pth],arrowptr)

		if len(paths[pth]) == 1 {
			paths[pth] = nil
		}

		if depth > max {
			max = pth
		}
	}

	return paths[max]
}

// **************************************************************************

func TruncatePathsByArrow(path []Link,arrow ArrowPtr) ([]Link,int) {

	for hop := 1; hop < len(path); hop++ {

		if path[hop].Arr != arrow {
			return path[:hop],hop
		}
	}

	return path,len(path)
}

// **************************************************************************
// Presentation on command line
// **************************************************************************

func PrintNodeOrbit(ctx PoSST, nptr NodePtr,width int) {

	node := GetDBNodeByNodePtr(ctx,nptr)		

	ShowText(node.S,width)
	fmt.Println()

	notes := GetNodeOrbit(ctx,nptr,"")

	PrintLinkOrbit(notes,EXPRESS)
	PrintLinkOrbit(notes,-EXPRESS)
	PrintLinkOrbit(notes,-CONTAINS)
	PrintLinkOrbit(notes,LEADSTO)
	PrintLinkOrbit(notes,-LEADSTO)
	PrintLinkOrbit(notes,NEAR)

	fmt.Println()
}

// **************************************************************************

func PrintLinkOrbit(notes [ST_TOP][]Orbit,sttype int) {

	t := STTypeToSTIndex(sttype)

	for n := range notes[t] {		

		if notes[t][n].Ctx != "" {
			txt := fmt.Sprintf(" -    (%s) - %s  .. %s\n",notes[t][n].Arrow,notes[t][n].Text,notes[t][n].Ctx)
			text := Indent(LEFTMARGIN) + txt
			ShowText(text,SCREENWIDTH)
		} else {
			txt := fmt.Sprintf(" -    (%s) - %s\n",notes[t][n].Arrow,notes[t][n].Text)
			text := Indent(LEFTMARGIN) + txt
			ShowText(text,SCREENWIDTH)
		}

	}

}

// **************************************************************************

func PrintLinkPath(ctx PoSST, cone [][]Link, p int, prefix string,chapter string,context []string) {

	if len(cone[p]) > 1 {

		path_start := GetDBNodeByNodePtr(ctx,cone[p][0].Dst)		
		
		start_shown := false

		var format int
		var stpath []string
		
		for l := 1; l < len(cone[p]); l++ {

			if !MatchContexts(context,cone[p][l].Ctx) {
				return
			}

			NewLine(format)

			if !start_shown {
				if len(cone) > 1 {
					fmt.Print(prefix,p+1," * ",path_start.S)
				} else {
					fmt.Print(prefix," * ",path_start.S)
				}
				start_shown = true
			}

			nextnode := GetDBNodeByNodePtr(ctx,cone[p][l].Dst)

			if !SimilarString(nextnode.Chap,chapter) {
				break
			}
			
			arr := GetDBArrowByPtr(ctx,cone[p][l].Arr)

			if arr.Short == "then" {
				fmt.Print("\n   >>> ")
				format = 0
			}

			if arr.Short == "prior" {
				fmt.Print("\n   <<< ")
			}

			stpath = append(stpath,STTypeName(STIndexToSTType(arr.STAindex)))
	
			if l < len(cone[p]) {
				fmt.Print("  -(",arr.Long,")->  ")
			}
			
			fmt.Print(nextnode.S)
			format += 2
		}

		fmt.Print("\n\n    Linkage process:")

		for s := range stpath {
			fmt.Print(" -(",stpath[s],")-> ")
		}
		fmt.Println(". \n")
	}
}

// **************************************************************************
// Presentation in JSON
// **************************************************************************

func JSONNodeEvent(ctx PoSST, nptr NodePtr) string {

	node := GetDBNodeByNodePtr(ctx,nptr)

	var event NodeEvent
	event.Text = node.S
	event.L = node.L
	event.Chap = node.Chap
	event.NPtr = nptr
	event.Orbits = GetNodeOrbit(ctx,nptr,"")

	jstr,_ := json.Marshal(event)

	return string(jstr)
}

// **************************************************************************

func JSONCone(ctx PoSST, cone [][]Link,chapter string,context []string) string {

	var jstr string = "["

	for p := 0; p < len(cone); p++ {

		path_start := GetDBNodeByNodePtr(ctx,cone[p][0].Dst)		
		
		start_shown := false

		var path []WebPath
		
		for l := 1; l < len(cone[p]); l++ {

			if !MatchContexts(context,cone[p][l].Ctx) {
				return "[]"
			}

			nextnode := GetDBNodeByNodePtr(ctx,cone[p][l].Dst)

			if !SimilarString(nextnode.Chap,chapter) {
				break
			}
			
			if !start_shown {
				var ws WebPath
				ws.Name = path_start.S
				ws.NPtr = cone[p][0].Dst
				path = append(path,ws)
				start_shown = true
			}

			arr := GetDBArrowByPtr(ctx,cone[p][l].Arr)
	
			if l < len(cone[p]) {
				var wl WebPath
				wl.Name = arr.Long
				wl.Arr = cone[p][l].Arr
				wl.STindex = arr.STAindex
				path = append(path,wl)
			}

			var wn WebPath
			wn.Name = nextnode.S
			wn.NPtr = cone[p][l].Dst
			path = append(path,wn)

		}

		encoded, _ := json.Marshal(path)
		jstr += fmt.Sprintf("%s",string(encoded))

		if p < len(cone)-1 {
			jstr += ",\n"
		}
	}

	jstr += "]"

	return jstr
}

// **************************************************************************

func JSON_TableOfContents(ctx PoSST,chap string,cn []string) string {

	chap_col := ""

	if chap != "any" && chap != "" {

		remove_chap_accents,chap_stripped := IsBracketedSearchTerm(chap)

		if remove_chap_accents {
			chap_search := "%"+chap_stripped+"%"
			chap_col = fmt.Sprintf("AND lower(unaccent(chap)) LIKE lower('%s')",chap_search)
		} else {
			chap_search := "%"+chap+"%"
			chap_col = fmt.Sprintf("AND lower(chap) LIKE lower('%s')",chap_search)
		}
	}

	_,cn_stripped := IsBracketedSearchList(cn)
	context := FormatSQLStringArray(cn_stripped)

	qstr := fmt.Sprintf("WITH matching_nodes AS "+
		"  (SELECT NFrom,ctx,match_context(ctx,%s) AS match FROM NodeArrowNode)"+
		"     SELECT DISTINCT chap,ctx FROM matching_nodes "+
		"      JOIN Node ON nptr=nfrom WHERE match=true %s",
		context,chap_col)
fmt.Println(qstr)
	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY TableOfContents Failed",err,qstr)
	}

	var rchap,rcontext string
	var toc = make(map[string][]string)

	for row.Next() {		
		err = row.Scan(&rchap,&rcontext)

		chps := strings.Split(rchap,",")
		for c := range chps {
			rc := chps[c]
			cn := ParseSQLArrayString(rcontext)
			for s := 0; s < len(cn); s++ {
				//cn[s] = strings.Replace(cn[s]," ","-",-1)
				toc[rc] = append(toc[rc],cn[s])
			}
		}
	}

	// JSON

	json_toc := "{ \"TOC\": ["
	var order []string

	for keys := range toc {
		order = append(order,keys)
	}

	sort.Strings(order)

	for key := 0; key < len(order); key++ {
		json_toc += fmt.Sprintf("{\n\"Chapter\": \"%s\",\n",order[key])

		var idemp = make(map[string]bool)
		var list []string

		for s := range toc[order[key]] {
			idemp[toc[order[key]][s]] = true
		}
		
		for vals := range idemp {
			list = append(list,vals)
		}

		sort.Strings(list)
		arr,_ := json.Marshal(list)

		json_toc += fmt.Sprintf("\"Contexts\": %s\n",string(arr))
		json_toc += "}"
		if key != len(order)-1 {
			json_toc += ",\n"
		}
	}

	json_toc += " ]}"
	
	row.Close()
	return json_toc
}

// **************************************************************************

func JSONPage(ctx PoSST, maplines []PageMap) string {

	var webnotes PageView
	var last,lastc string

	for n := 0; n < len(maplines); n++ {

		var path []WebPath

		txtctx := ContextString(maplines[n].Context)

		if last != maplines[n].Chapter || lastc != txtctx {
			webnotes.Title = maplines[n].Chapter
			webnotes.Context = txtctx
			last = maplines[n].Chapter
			lastc = txtctx
		}
		
		for lnk := 0; lnk < len(maplines[n].Path); lnk++ {
			
			text := GetDBNodeByNodePtr(ctx,maplines[n].Path[lnk].Dst)
			
			if lnk == 0 {
				var ws WebPath
				ws.Name = text.S
				ws.NPtr = maplines[n].Path[lnk].Dst
				path = append(path,ws)
				
			} else {
				arr := GetDBArrowByPtr(ctx,maplines[n].Path[lnk].Arr)
				var wl WebPath
				wl.Name = arr.Long
				wl.Arr = maplines[n].Path[lnk].Arr
				wl.STindex = arr.STAindex
				path = append(path,wl)
				
				var ws WebPath
				ws.Name = text.S
				ws.NPtr = maplines[n].Path[lnk].Dst
				path = append(path,ws)
				
			}
		}
		webnotes.Notes = append(webnotes.Notes,path)
	}
	
	encoded, _ := json.Marshal(webnotes)
	jstr := fmt.Sprintf("%s",string(encoded))

	return jstr
}

// **************************************************************************
// Retrieve Analysis
// **************************************************************************

func GetAppointmentArrayByArrow(ctx PoSST, context []string,chapter string) map[ArrowPtr][]NodePtr {

          /* arr |             x             
            -----+---------------------------
              18 | {"(2,4)","(3,4)","(4,4)"}
             138 | {"(4,4)","(0,4)"}
              97 | {"(1,2)"}
              96 | {"(0,4)"}
             137 | {"(1,4)","(0,3)"}
              52 | {"(0,4)"}
              53 | {"(0,2)"} */

	// Postgres && operator on arrays is SET OVERLAP .. how to solve this?

	qstr := "SELECT arr,array_agg(DISTINCT NTo) FROM NodeArrowNode"

	if context != nil {
		qstr += fmt.Sprintf(" WHERE match_context(ctx,%s)",FormatSQLStringArray(context))
	}

	qstr += " GROUP BY arr"

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointmentArrayByArrow Failed",err,qstr)
	}

	var arry string
	var arr ArrowPtr
	var retval = make(map[ArrowPtr][]NodePtr)

	for row.Next() {		
		err = row.Scan(&arr,&arry)
		retval[arr] = ParseSQLNPtrArray(arry)
	}

	row.Close()
	
	return retval
}

// **************************************************************************

func GetAppointmentArrayBySSType(ctx PoSST) map[int][]NodePtr {


          /* sttype |             array_agg             
             --------+-----------------------------------
                 -3 | {"(0,4)","(4,4)"}
                 -2 | {"(1,2)"}
                 -1 | {"(0,2)"}
                  1 | {"(0,4)","(2,4)","(3,4)","(4,4)"}
                  2 | {"(0,4)"}
                  3 | {"(0,3)","(1,4)"} */


	qstr := "SELECT sttype, array_agg(DISTINCT NTo) FROM NodeArrowNode GROUP BY Sttype order by sttype"

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointmentArrayByArrow Failed",err)
	}

	var arry string
	var sttype int
	var retval = make(map[int][]NodePtr)

	for row.Next() {		
		err = row.Scan(&sttype,&arry)
		retval[sttype] = ParseSQLNPtrArray(arry)
	}

	row.Close()
	
	return retval
}

// **************************************************************************

func GetAppointmentHistogramByArrow(ctx PoSST) map[ArrowPtr]int {

/* arr | count 
-----+-------
  18 |     3
  52 |     1
  53 |     1
  96 |     1
  97 |     1
 137 |     2
 138 |     2 */

	qstr := "SELECT arr,count(NTo) FROM NodeArrowNode GROUP BY arr order by Arr"

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointmentArrayByArrow Failed",err)
	}

	var freq int
	var arr ArrowPtr
	var retval = make(map[ArrowPtr]int)

	for row.Next() {		
		err = row.Scan(&arr,&freq)
		retval[arr] = freq
	}

	row.Close()
	
	return retval
}

// **************************************************************************

func GetAppointmentHistogramBySSType(ctx PoSST) map[int]int {

/* sttype | x 
--------+---
      1 | 4
     -1 | 1
      2 | 1
     -2 | 1
      3 | 2
     -3 | 2 */

	qstr := "SELECT sttype,count(NTo) FROM NodeArrowNode GROUP BY Sttype"

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointmentArrayByArrow Failed",err)
	}

	var freq int
	var sttype int
	var retval = make(map[int]int)

	for row.Next() {		
		err = row.Scan(&sttype,&freq)
		retval[sttype] = freq
	}
	return retval
}

// **************************************************************************

func GetAppointmentNodesByArrow(ctx PoSST) []ArrowAppointment {

/*  nfrom | arr | array_agg 
-------+-----+-----------
 (4,0) |  18 | {"(2,4)"}
 (4,2) |  18 | {"(3,4)"}
 (4,3) |  18 | {"(4,4)"}
 (2,0) |  52 | {"(0,4)"}
 (4,0) |  53 | {"(0,2)"}
 (2,1) |  96 | {"(0,4)"}
 (4,0) |  97 | {"(1,2)"}
 (4,0) | 137 | {"(1,4)"}
 (4,4) | 137 | {"(0,3)"}
 (3,0) | 138 | {"(4,4)"}
 (4,1) | 138 | {"(0,4)"} */

	qstr := "SELECT NFrom,Arr,array_agg(NTo) FROM NodeArrowNode GROUP BY Arr,Nfrom HAVING count(NTo) > 1 ORDER BY Arr "

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointmentArrayByArrow Failed",err)
	}

	var nptr,arry string
	var retval []ArrowAppointment
	var this ArrowAppointment

	for row.Next() {		
		err = row.Scan(&nptr,&this.Arr,&arry)
		fmt.Sscanf(nptr,"(%d,%d)",&this.NFrom.Class,&this.NFrom.CPtr)
		this.NTo = ParseSQLNPtrArray(arry)
		retval = append(retval,this)
	}

	row.Close()

	return retval
}

// **************************************************************************

func GetAppointmentNodesBySTType(ctx PoSST) []STTypeAppointment {

/*  nfrom | sttype | array_agg 
-------+--------+-----------
 (3,0) |     -3 | {"(4,4)"}
 (4,1) |     -3 | {"(0,4)"}
 (4,0) |     -2 | {"(1,2)"}
 (4,0) |     -1 | {"(0,2)"}
 (2,0) |      1 | {"(0,4)"}
 (4,0) |      1 | {"(2,4)"}
 (4,2) |      1 | {"(3,4)"}
 (4,3) |      1 | {"(4,4)"}
 (2,1) |      2 | {"(0,4)"}
 (4,0) |      3 | {"(1,4)"}
 (4,4) |      3 | {"(0,3)"}*/

	qstr := "SELECT NFrom,sttype,array_agg(NTo) FROM NodeArrowNode GROUP BY sttype,Nfrom HAVING count(NTo) > 1 ORDER BY sttype"

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointmentArrayBySTType Failed",err)
	}

	var retval []STTypeAppointment
	var this STTypeAppointment
	var nptr,arr,arry string

	for row.Next() {		
		err = row.Scan(&nptr,&arr,&arry)
		err = row.Scan(&nptr,&this.STType,&arry)
		fmt.Sscanf(nptr,"(%d,%d)",&this.NFrom.Class,&this.NFrom.CPtr)
		this.NTo = ParseSQLNPtrArray(arry)
		retval = append(retval,this)
	}

	row.Close()
	return retval
}

// **************************************************************************
// CENTRALITY
// **************************************************************************

func TallyPath(ctx PoSST,path []Link,between map[string]int) map[string]int {

	// count how often each node appears in the different path solutions

	for leg := range path {
		n := GetDBNodeByNodePtr(ctx,path[leg].Dst)
		between[n.S]++
	}

	return between
}

// **************************************************************************

func BetweenNessCentrality(ctx PoSST,solutions [][]Link) string {

	var betweenness = make(map[string]int)

	for s := 0; s < len(solutions); s++ {
		betweenness = TallyPath(ctx,solutions[s],betweenness)
	}

	var inv = make(map[int][]string)
 	var order []int

	for key := range betweenness {
		inv[betweenness[key]] = append(inv[betweenness[key]],key)
	}

	for key := range inv {
		order = append(order,key)
	}

	sort.Ints(order)

	var betw,retval string

	for key := len(order)-1; key >= 0; key-- {
		betw = fmt.Sprintf("%.2f : ",float64(order[key])/float64(len(solutions)))
		for el := 0; el < len(inv[order[key]]); el++ {
			betw += fmt.Sprintf("%s",inv[order[key]][el])
			if el < len(inv[order[key]])-1 {
				betw += ", "
			}
		}
		retval += fmt.Sprintf("\"%s\"",betw)
		if key > 0 {
			retval += ","
		}
	}
	return retval
}

// **************************************************************************

func SuperNodesByConicPath(solutions [][]Link, maxdepth int) [][]NodePtr {

	var supernodes [][]NodePtr
	
	for depth := 0; depth < maxdepth*2; depth++ {
		
		for p_i := 0; p_i < len(solutions); p_i++ {

			if depth == len(solutions[p_i])-1 {
				supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_i][depth].Dst)
			}

			if depth > len(solutions[p_i])-1 {
				continue
			}

			supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_i][depth].Dst)

			for p_j := p_i+1; p_j < len(solutions); p_j++ {

				if depth < 1 || depth > len(solutions[p_j])-2 {
					break
				}

				if solutions[p_i][depth-1].Dst == solutions[p_j][depth-1].Dst && 
				   solutions[p_i][depth+1].Dst == solutions[p_j][depth+1].Dst {
					   supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_j][depth].Dst)
				}
			}
		}		
	}

	return supernodes
}

// **************************************************************************

func SuperNodes(ctx PoSST,solutions [][]Link, maxdepth int) string {

	supernodes := SuperNodesByConicPath(solutions,maxdepth)

	var retval string

	for g := range supernodes {

		super := ""

		for n := range supernodes[g] {
			node := GetDBNodeByNodePtr(ctx,supernodes[g][n])
			super += fmt.Sprintf("%s",node.S)
			if n < len(supernodes[g])-1 {
				super += ", "
			}
		}
		retval += fmt.Sprintf("\"%s\"",super)
		if g < len(supernodes)-1 {
			retval += ", "
		}
	}

	return retval
}

// **************************************************************************
// SQL marshalling Tools
// **************************************************************************

func SQLEscape(s string) string {

	return strings.Replace(s, `'`, `''`, -1)
}

// **************************************************************************

func Array2Str(arr []string) string {

	var s string

	for a := 0; a < len(arr); a++ {
		s += arr[a]
		if a < len(arr)-1 {
			s += ", "
		}
	}

	return s
}

// **************************************************************************

func Str2Array(s string) ([]string,int) {

	var non_zero int
	s = strings.Replace(s,"{","",-1)
	s = strings.Replace(s,"}","",-1)
	s = strings.Replace(s,"\"","",-1)

	arr := strings.Split(s,",")

	for a := 0; a < len(arr); a++ {
		arr[a] = strings.TrimSpace(arr[a])
		if len(arr[a]) > 0 {
			non_zero++
		}
	}

	return arr,non_zero
}

// **************************************************************************

func ParseSQLNPtrArray(s string) []NodePtr {

	stringify := ParseSQLArrayString(s)

	var retval []NodePtr
	var nptr NodePtr

	for n := 0; n < len(stringify); n++ {
		fmt.Sscanf(stringify[n],"(%d,%d)",&nptr.Class,&nptr.CPtr)
		retval = append(retval,nptr)
	}

	return retval
}

// **************************************************************************

func ParseSQLArrayString(whole_array string) []string {

	// array as {"(1,2,3)","(4,5,6)",spacelessstring}

      	var l []string

    	whole_array = strings.Replace(whole_array,"{","",-1)
    	whole_array = strings.Replace(whole_array,"}","",-1)

	uni_array := []rune(whole_array)

	var items []string
	var item []rune
	var protected = false

	for u := range uni_array {

		if uni_array[u] == '"' {
			protected = !protected
			continue
		}

		if !protected && uni_array[u] == ',' {
			items = append(items,string(item))
			item = nil
			continue
		}

		item = append(item,uni_array[u])
	}

	if item != nil {
		items = append(items,string(item))
	}

	for i := range items {

	    s := strings.TrimSpace(items[i])

	    l = append(l,s)
	    }

	return l
}

// **************************************************************************

func FormatSQLIntArray(array []int) string {

        if len(array) == 0 {
		return "'{ }'"
        }

	sort.Slice(array, func(i, j int) bool {
		return array[i] < array[j]
	})

	var ret string = "'{ "
	
	for i := 0; i < len(array); i++ {
		ret += fmt.Sprintf("%d",array[i])
	    if i < len(array)-1 {
	    ret += ", "
	    }
        }

	ret += " }' "

	return ret
}

// **************************************************************************

func FormatSQLStringArray(array []string) string {

        if len(array) == 0 {
		return "'{ }'"
        }

	sort.Strings(array) // Avoids ambiguities in db comparisons

	var ret string = "'{ "
	
	for i := 0; i < len(array); i++ {
		ret += fmt.Sprintf("\"%s\"",SQLEscape(array[i]))
	    if i < len(array)-1 {
	    ret += ", "
	    }
        }

	ret += " }' "

	return ret
}

// **************************************************************************

func FormatSQLNodePtrArray(array []NodePtr) string {

        if len(array) == 0 {
		return "'{ }'"
        }

	var ret string = "'{ "
	
	for i := 0; i < len(array); i++ {
		ret += fmt.Sprintf("\"(%d,%d)\"",array[i].Class,array[i].CPtr)
	    if i < len(array)-1 {
	    ret += ", "
	    }
        }

	ret += " }' "

	return ret
}

// **************************************************************************

func ParseSQLLinkString(s string) Link {

        // e.g. (77,0.34,"{ ""fairy castles"", ""angel air"" }","(4,2)")
	// This feels dangerous. Is postgres consistent here?

      	var l Link

    	s = strings.Replace(s,"(","",-1)
    	s = strings.Replace(s,")","",-1)
	s = strings.Replace(s,"\"\"",";",-1)
	s = strings.Replace(s,"\"","",-1)
	s = strings.Replace(s,"\\","",-1)
	
        items := strings.Split(s,",")

	for i := 0; i < len(items); i++ {
		items[i] = strings.Replace(items[i],"{","",-1)
		items[i] = strings.Replace(items[i],"}","",-1)
		items[i] = strings.Replace(items[i],";","",-1)
		items[i] = strings.TrimSpace(items[i])
	}

	// Arrow type
	fmt.Sscanf(items[0],"%d",&l.Arr)

	// Link weight
	fmt.Sscanf(items[1],"%f",&l.Wgt)

	// These are the context array

	var array []string

	for i := 2; i <= len(items)-3; i++ {
		array = append(array,items[i])
	}

	l.Ctx = array

	// the last two are the NPtr

	fmt.Sscanf(items[len(items)-2],"%d",&l.Dst.Class)
	fmt.Sscanf(items[len(items)-1],"%d",&l.Dst.CPtr)

	return l
}

//**************************************************************

func ParseLinkArray(s string) []Link {

	var array []Link

	s = strings.TrimSpace(s)

	if len(s) <= 2 {
		return array
	}

	strarray := strings.Split(s,"\n")

	for i := 0; i < len(strarray); i++ {
		link := ParseSQLLinkString(strarray[i])
		array = append(array,link)
	}
	
	return array
}

//**************************************************************

func ParseMapLinkArray(s string) []Link {

	var array []Link

	s = strings.TrimSpace(s)

	if len(s) <= 2 {
		return array
	}

	strarray := strings.Split(s,"\",\"")

	for i := 0; i < len(strarray); i++ {
		link := ParseSQLLinkString(strarray[i])
		array = append(array,link)
	}
	
	return array
}

//**************************************************************

func ParseLinkPath(s string) [][]Link {

	// Each path will start on a new line, with comma sep Link encodings

	var array [][]Link
	var index int = 0
	s = strings.TrimSpace(s)

	lines := strings.Split(s,"\n")

	for line := range lines {

		if len(lines[line]) > 0 {

			links := strings.Split(lines[line],";")

			if len(links) < 2 {
				continue
			}

			array = append(array,make([]Link,0))

			for l := 0; l < len(links); l++ {

				lnk := ParseSQLLinkString(links[l])
				array[index] = append(array[index],lnk)
			}
			index++
		}
	}

	if index < 1 {
		return nil
	}
	return array
}

//**************************************************************

func StorageClass(s string) (int,int) {
	
	var spaces int = 0

	var l = len(s)
	
	for i := 0; i < l; i++ {
		
		if s[i] == ' ' {
			spaces++
		}
		
		if spaces > 2 {
			break
		}
	}
	
	// Text usage tends to fall into a number of different roles, with a power law
	// frequency of occurrence in a text, so let's classify in order of likely usage
	// for small and many, we use a hashmap/btree
	
	switch spaces {
	case 0:
		return l,N1GRAM
	case 1:
		return l,N2GRAM
	case 2:
		return l,N3GRAM
	}
	
	// For longer strings, a linear search is probably fine here
        // (once it gets into a database, it's someone else's problem)
	
	if l < 128 {
		return l,LT128
	}
	
	if l < 1024 {
		return l,LT1024
	}
	
	return l,GT1024
}

// **************************************************************************

func DiracNotation(s string) (bool,string,string,string) {

	var begin,end,context string

	if s[0] == '<' && s[len(s)-1] == '>' {
		matrix := s[1:len(s)-1]
		params := strings.Split(matrix,"|")
		
		switch len(params) {
			
		case 2: 
			begin = params[0]
			end = params[1]
		case 3:
			begin = params[0]
			context = params[1]
			end = params[2]			
		default:
			fmt.Println("Bad Dirac notation, should be <a|b> or <a|context|b>")
			os.Exit(-1)
		}
	} else {
		return false,"","",""
	}

	return true,begin,end,context
}

// **************************************************************************
// Semantic Spacetime names and channels
// **************************************************************************

func STTypeDBChannel(sttype int) string {

	// This expects the range for sttype to be unshifted 0,+/-

	var link_channel string
	switch sttype {

	case NEAR:
		link_channel = I_NEAR
	case LEADSTO:
		link_channel = I_PLEAD
	case CONTAINS:
		link_channel = I_PCONT
	case EXPRESS:
		link_channel = I_PEXPR
	case -LEADSTO:
		link_channel = I_MLEAD
	case -CONTAINS:
		link_channel = I_MCONT
	case -EXPRESS:
		link_channel = I_MEXPR
	default:
		fmt.Println(ERR_ILLEGAL_LINK_CLASS,sttype)
		os.Exit(-1)
	}

	return link_channel
}

// **************************************************************************

func STIndexToSTType(stindex int) int {

	// Convert shifted array index to symmetrical type

	return stindex - ST_ZERO
}

// **************************************************************************

func STTypeToSTIndex(stindex int) int {

	// Convert shifted array index to symmetrical type

	return stindex + ST_ZERO
}

// **************************************************************************

func STTypeName(sttype int) string {

	switch sttype {
	case -EXPRESS:
		return "-is property of"
	case -CONTAINS:
		return "-contained by"
	case -LEADSTO:
		return "-comes from"
	case NEAR:
		return "=Similarity"
	case LEADSTO:
		return "+leads to"
	case CONTAINS:
		return "+contains"
	case EXPRESS:
		return "+property"
	}

	return "Unknown ST type"
}

// **************************************************************************
// String matching - keep this simple for now
// **************************************************************************

func SimilarString(s1,s2 string) bool {

	// Placeholder
	// Need to handle pluralisation patterns etc... multi-language

	if s1 == s2 {
		return true
	}

	if s1 == "" || s2 == "" || s1 == "any" || s2 == "any" {  // same as any
		return true
	}

	if (s1[0] == '!' && s2[0] != '!') || (s1[0] != '!' && s2[0] == '!') {
		if !strings.Contains(s2,s1) || !strings.Contains(s1,s2) {
			return true
		}
	}

	if strings.Contains(s2,s1) || strings.Contains(s1,s2) {
		return true
	}

	return false
}

//****************************************************************************

func MatchArrows(arrows []ArrowPtr,arr ArrowPtr) bool {

	for a := range arrows {
		if arrows[a] == arr {
			return true
		}
	}

	return false
}

//****************************************************************************

func MatchContexts(context1 []string,context2 []string) bool {

	if context1 == nil || context2 == nil {
		return true
	}

	for c := range context1 {

		if MatchesInContext(context1[c],context2) {
			return true
		}
	}
	return false 
}

//****************************************************************************

func MatchesInContext(s string,context []string) bool {
	
	for c := range context {
		if SimilarString(s,context[c]) {
			return true
		}
	}
	return false 
}

// **************************************************************************
// Misc tools
// **************************************************************************

func RunErr(message string) {

	const red = "\033[31;1;1m"
	const endred = "\033[0m"

	fmt.Println("SSTorytime",message,endred)

}

// **************************************************************************

func EscapeString(s string) string {

	// Don't do this here, move to SQLEscape()
	return s
}







//******************************************************************

func ContextString(context []string) string {

	var s string

	for c := 0; c < len(context); c++ {

		s += context[c] + " "
	}

	return s
}

//****************************************************************************

func ShowText(s string, width int) {

	var spacecounter int
	var linecounter int
	var indent string = Indent(LEFTMARGIN)

	if width < 40 {
		width = SCREENWIDTH
	}

	// Check is the string has a large number of spaces, in which case it's
	// probably preformatted,

	runes := []rune(s)

	for r := 0; r < len(runes); r++ {
		if unicode.IsSpace(runes[r]) {
			spacecounter++
		}
	} 

	if len(runes) > SCREENWIDTH - LEFTMARGIN - RIGHTMARGIN {
		if spacecounter > len(runes) / 3 {
			fmt.Println()
			fmt.Println(s)
			return
		}
	}

	// Format

	linecounter = 0

	for r := 0; r < len(runes); r++ {
		
		if unicode.IsSpace(runes[r]) && linecounter > width-RIGHTMARGIN {
			if runes[r] != '\n' {
				fmt.Print("\n",indent)
				linecounter = 0
				continue
			} else {
				linecounter = 0
			}
		}
		if unicode.IsPunct(runes[r]) && linecounter > width-RIGHTMARGIN {
			if runes[r] != '\n' {
				fmt.Print("\n",indent)
				linecounter = 0
				continue
			} else {
				linecounter = 0
			}
		}
		fmt.Print(string(runes[r]))
		linecounter++
		
	}
}

//****************************************************************************

func Indent(indent int) string {

	spc := ""

	for i := 0; i < indent; i++ {
		spc += " "
	}

	return spc
}

//****************************************************************************

func NewLine(n int) {

	if n % 6 == 0 {
		fmt.Print("\n    ",)
	}
}

// **************************************************************************

func Waiting() {

	const propaganda = "IT.ISN'T.KNOWLEDGE.UNLESS.YOU.KNOW.IT.!!"
	const interval = 3

	if SILLINESS {
		if SILLINESS_COUNTER % interval != 0 {
			fmt.Print(".")
		} else {
			fmt.Print(string(propaganda[SILLINESS_POS]))
			SILLINESS_POS++
			if SILLINESS_POS > len(propaganda)-1 {
				SILLINESS_POS = 0
				SILLINESS = false
			}
		}
	} else {
		fmt.Print(".")
	}

	if SILLINESS_COUNTER % (len(propaganda)*interval*interval) == 0 {
		SILLINESS = !SILLINESS
	}

	SILLINESS_COUNTER++
}

// **************************************************************************

func Already (s string, cone map[int][]string) bool {

	for l := range cone {
		for n := 0; n < len(cone[l]); n++ {
			if s == cone[l][n] {
				return true
			}
		}
	}

	return false
}

//****************************************************************************

func Arrow2Int(arr []ArrowPtr) []int {

	var ret []int

	for a := range arr {
		ret = append(ret,int(arr[a]))
	}

	return ret
}

//****************************************************************************
// Unicode
//****************************************************************************

func IsBracketedSearchList(list []string) (bool,[]string) {

	var stripped_list []string
	retval := false

	for i := range list {

		isbrack,stripped := IsBracketedSearchTerm(list[i])

		if isbrack {
			retval = true
			stripped_list = append(stripped_list,"|"+stripped+"|")
		} else {
			stripped_list = append(stripped_list,list[i])
		}

	}

	return retval,stripped_list
}

//****************************************************************************

func IsBracketedSearchTerm(src string) (bool,string) {

	retval := false
	stripped := src

	decomp := strings.TrimSpace(src)

	if len(decomp) == 0 {
		return false, ""
	}

	if decomp[0] == '(' && decomp[len(decomp)-1] == ')' {
		retval = true
		stripped = decomp[1:len(decomp)-1]
		stripped = strings.TrimSpace(stripped)
	}

	return retval,stripped
}



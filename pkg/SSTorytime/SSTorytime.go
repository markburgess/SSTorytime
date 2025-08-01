//**************************************************************
//
// Library of methods and tools for Semantic Spacetime Graph Processes
// (All in one file for easy searching)
//
//**************************************************************

package SSTorytime

import (
	"database/sql"
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"strconv"
	"unicode"
	"sort"
	"encoding/json"
	"regexp"
	"math"
	"time"
	_ "github.com/lib/pq"

)

//**************************************************************
//
// Part 1: N4L language and Graph Representation
//
//**************************************************************

//**************************************************************
// Errors
//**************************************************************

const (
	CREDENTIALS_FILE = ".SSTorytime" // user's home directory

	ERR_ST_OUT_OF_BOUNDS = "Link STtype is out of bounds (must be -3 to +3)"
	ERR_ILLEGAL_LINK_CLASS = "ILLEGAL LINK CLASS"
	ERR_NO_SUCH_ARROW = "No such arrow has been declared in the configuration: "
	ERR_MEMORY_DB_ARROW_MISMATCH = "Arrows in database are not in synch (shouldn't happen)"
	WARN_DIFFERENT_CAPITALS = "WARNING: Another capitalization exists"

	SCREENWIDTH = 120
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

	// semantics, resverved names
)

var BASE_DB_CHANNEL_STATE[7] ClassedNodePtr

var CLASS_CHANNEL_DESCRIPTION = []string{"","single word ngram","two word ngram","three word ngram",
	"string less than 128 chars","string less than 1024 chars","string greater than 1024 chars"}

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
	Wgt float32
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

//**************************************************************

type PageView struct {
	Title   string
	Context string
	Notes   [][]WebPath
}

//**************************************************************

type Coords struct {
	X float32
	Y float32
	Z float32
}

//**************************************************************

type WebPath struct {
	NPtr    NodePtr
	Arr     ArrowPtr
	STindex int
	Name    string
	XYZ     Coords
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

type Link struct {  // A link is a type of arrow, with context
                    // and maybe with a weightfor package math
	Arr ArrowPtr         // type of arrow, presorted
	Wgt float32          // numerical weight of this link
	Ctx []string         // context for this pathway
	Dst NodePtr          // adjacent event/item/node
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

const APPOINTMENT_TYPE = "CREATE TYPE Appointment AS  " +
	"(                    " +
	"Arr    int," +
	"STType int," +
	"Chap   text," +
	"Ctx    text[]," +
	"NTo    NodePtr," +
	"NFrom  NodePtr[]" +
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

const ARROW_DIRECTORY_TABLE = "CREATE TABLE IF NOT EXISTS ArrowDirectory " +
	"(    " +
	"STAindex int,           " +
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
        NPtr    NodePtr
	XYZ     Coords
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
	XYZ     Coords  // coords
	OOO     Coords  // origin
}

//******************************************************************

func Open(load_arrows bool) PoSST {

	var ctx PoSST
	var err error

	// Replace this with a private file

	var (
		user     = "sstoryline"
		password = "sst_1234"
		dbname   = "sstoryline"
	)

	user,password,dbname = OverrideCredentials(user,password,dbname)

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

func OverrideCredentials(u,p,d string) (string,string,string) {

	dirname, err := os.UserHomeDir()

	if err != nil && len(dirname) > 1 {
		fmt.Println("Unable to determine user's home directory")
		os.Exit(-1)
	}

	filename := dirname+"/"+CREDENTIALS_FILE
	content,err := ioutil.ReadFile(filename)

	if err != nil {
		return u,p,d
	}

	/* format
          dbname: sstoryline 
          user:sstoryline 
          passwd: sst_1234
        */

	var (
		offset,delta int
		user=u
		password=p
		dbname=d
	)

	for offset = 0; offset < len(content); offset = offset {

		var conf string
		fmt.Sscanf(string(content[offset:]),"%s",&conf)

		if len(conf) > 0 && conf[len(conf)-1] != ':' { // missing space

			for delta = 0; delta < len(conf); delta++ {
				if conf[delta] == ':' {
					conf = conf[:delta+1]
				}
			}
		}

		switch(conf) {
		case "user:":
			delta = len(conf)
			user,offset = GetLine(content,offset+delta)
		case "passwd:","password:":
			delta = len(conf)
			password,offset = GetLine(content,offset+delta)
		case "db:","dbname:":
			delta = len(conf)
			dbname,offset = GetLine(content,offset+delta)
		default:
			offset++
		}
	}

	return user,password,dbname
}

// **************************************************************************

func GetLine(s []byte,i int) (string,int) {

	var result []byte

	for o := i; o < len(s); o++ {

		if s[o] == '\n' {
			i = o
			break
		}

		result = append(result,s[o])
	}

	return string(result),i
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

	for i := N_GRAM_MIN; i < N_GRAM_MAX; i++ {

		STM_NGRAM_FREQ[i] = make(map[string]float64)
		STM_NGRAM_LOCA[i] = make(map[string][]int)
		STM_NGRAM_LAST[i] = make(map[string]int)
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
		ctx.DB.QueryRow("drop function GetAppointments")
		ctx.DB.QueryRow("drop function UnCmp")

		ctx.DB.QueryRow("drop table Node")
		ctx.DB.QueryRow("drop table PageMap")
		ctx.DB.QueryRow("drop table NodeArrowNode")
		ctx.DB.QueryRow("drop type NodePtr")
		ctx.DB.QueryRow("drop type Link")
		ctx.DB.QueryRow("drop type Appointment")

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

	if !CreateType(ctx,APPOINTMENT_TYPE) {
		fmt.Println("Unable to create type as, ",APPOINTMENT_TYPE)
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
	DownloadArrowsFromDB(ctx)
	SynchronizeNPtrs(ctx)

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

func GetMemoryNodeFromPtr(frptr NodePtr) Node {

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

	for a := range ARROW_DIRECTORY {
		if ARROW_DIRECTORY[a].Long == name || ARROW_DIRECTORY[a].Short == alias {
			return ArrowPtr(-1)
		}
	}

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

	if fwd == ArrowPtr(-1) || bwd == ArrowPtr(-1) {
		return
	}

	// Lookup inverse by long name, only need this in search presentation

	INVERSE_ARROWS[fwd] = bwd
	INVERSE_ARROWS[bwd] = fwd
}

//**************************************************************
// Write to database
//**************************************************************

func GraphToDB(ctx PoSST,wait_counter bool) {

	for class := N1GRAM; class <= GT1024; class++ {

		offset := int(BASE_DB_CHANNEL_STATE[class])

		switch class {
		case N1GRAM:
			for n := offset; n < len(NODE_DIRECTORY.N1directory); n++ {
				org := NODE_DIRECTORY.N1directory[n]
				UploadNodeToDB(ctx,org,class)
				Waiting(wait_counter)
			}
		case N2GRAM:
			for n := offset; n < len(NODE_DIRECTORY.N2directory); n++ {
				org := NODE_DIRECTORY.N2directory[n]
				UploadNodeToDB(ctx,org,class)
				Waiting(wait_counter)
			}
		case N3GRAM:
			for n := offset; n < len(NODE_DIRECTORY.N3directory); n++ {
				org := NODE_DIRECTORY.N3directory[n]
				UploadNodeToDB(ctx,org,class)
				Waiting(wait_counter)
			}
		case LT128:
			for n := offset; n < len(NODE_DIRECTORY.LT128); n++ {
				org := NODE_DIRECTORY.LT128[n]
				UploadNodeToDB(ctx,org,class)
				Waiting(wait_counter)
			}
		case LT1024:
			for n := offset; n < len(NODE_DIRECTORY.LT1024); n++ {
				org := NODE_DIRECTORY.LT1024[n]
				UploadNodeToDB(ctx,org,class)
				Waiting(wait_counter)
			}

		case GT1024:
			for n := offset; n < len(NODE_DIRECTORY.GT1024); n++ {
				org := NODE_DIRECTORY.GT1024[n]
				UploadNodeToDB(ctx,org,class)
				Waiting(wait_counter)
			}
		}
	}

	fmt.Println("\nStoring Arrows...")

	// Avoid duplicates, even if we're not wiping. since we loaded everything

	ctx.DB.QueryRow("drop table ArrowDirectory")
	ctx.DB.QueryRow("drop table ArrowInverses")

	if !CreateTable(ctx,ARROW_INVERSES_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_INVERSES_TABLE)
		os.Exit(-1)
	}
	if !CreateTable(ctx,ARROW_DIRECTORY_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_DIRECTORY_TABLE)
		os.Exit(-1)
	}

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
// Store - High level API
// **************************************************************************

func Vertex(ctx PoSST, name,chap string) Node {

	var n Node

	n.S = name
	n.Chap = chap

	return IdempDBAddNode(ctx,n)
}

// **************************************************************************

func Edge(ctx PoSST,from Node,arrow string,to Node,context []string,weight float32) (ArrowPtr,int) {

	arrowptr,sttype := GetDBArrowsWithArrowName(ctx,arrow)

	var link Link

	link.Arr = arrowptr
	link.Dst = to.NPtr
	link.Wgt = weight
	link.Ctx = context

	IdempDBAddLink(ctx,from,link,to)
	CreateDBNodeArrowNode(ctx,from.NPtr,link,sttype)

	return arrowptr,sttype
}

// **************************************************************************

func HubJoin(ctx PoSST,name,chap string,nptrs []NodePtr,arrow string,context []string,weight []float32) Node {

	// Create a container node joining several other nodes in a list, like a hyperlink

	if nptrs == nil {
		fmt.Println("Call to HubJoin with a null list of pointers")
		os.Exit(-1)
	}

	if weight == nil {
		for n := 0; n < len(nptrs); n++ {
			weight = append(weight,1.0)
		}
	}

	if len(nptrs) != len(weight) {
		fmt.Println("Call to HubJoin with inconsistent node/weight pointer arrays: dimensions ",len(nptrs),"vs",len(weight))
		os.Exit(-1)
	}

	var chaps = make(map[string]int)

	if name == "" {
		name = "hub_"+arrow+"_"
		for n := range nptrs {
			name += fmt.Sprintf("(%d,%d)",nptrs[n].Class,nptrs[n].CPtr)
			node := GetDBNodeByNodePtr(ctx,nptrs[n])
			chaps[node.Chap]++
		}
	}

	var to Node

	to.S = name

	if chap != "" {
		to.Chap = chap
	} else 	if chap == "" && len(chaps) == 1 {
		for ch := range chaps {
			to.Chap = ch
		}
	}

	container := IdempDBAddNode(ctx,to)

	arrowptr,sttype := GetDBArrowsWithArrowName(ctx,arrow)

	for nptr := range nptrs {

		var link Link
		link.Arr = arrowptr
		link.Dst = container.NPtr
		link.Wgt = weight[nptr]
		link.Ctx = context
		from := GetDBNodeByNodePtr(ctx,nptrs[nptr])
		IdempDBAddLink(ctx,from,link,container)
		CreateDBNodeArrowNode(ctx,nptrs[nptr],link,sttype)
	}

	return GetDBNodeByNodePtr(ctx,container.NPtr)
}

// **************************************************************************
// Lower level functions
// **************************************************************************

func CreateDBNode(ctx PoSST, n Node) Node {

	// Add node version setting explicit CPtr value, note different function call
	// We use this function when we ARE managing/counting CPtr values ourselves

	var qstr string

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

func IdempDBAddNode(ctx PoSST,n Node) Node {

	// We use this function when we aren't counting CPtr values
	// This functon may be deprecated in future, replaced by Node()

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

// **************************************************************************

func UploadNodeToDB(ctx PoSST, org Node,channel int) {

	CreateDBNode(ctx,org)

	const nolink = 999
	var empty Link

	for stindex := 0; stindex < len(org.I); stindex++ {

		for lnk := range org.I[stindex] {

			dstlnk := org.I[stindex][lnk]
			sttype := STIndexToSTType(stindex)

			AppendDBLinkToNode(ctx,org.NPtr,dstlnk,sttype)
			CreateDBNodeArrowNode(ctx,org.NPtr,dstlnk,sttype)
		}

		CreateDBNodeArrowNode(ctx,org.NPtr,empty,nolink)
	}
}

// **************************************************************************

func UploadArrowToDB(ctx PoSST,arrow ArrowPtr) {

	staidx := ARROW_DIRECTORY[arrow].STAindex
	long := SQLEscape(ARROW_DIRECTORY[arrow].Long)
	short := SQLEscape(ARROW_DIRECTORY[arrow].Short)

	qstr := fmt.Sprintf("INSERT INTO ArrowDirectory (STAindex,Long,Short,ArrPtr) SELECT %d,'%s','%s',%d WHERE NOT EXISTS (SELECT Long,Short,ArrPtr FROM ArrowDirectory WHERE lower(Long) = lower('%s') OR lower(Short) = lower('%s') OR ArrPtr = %d)",staidx,long,short,arrow,long,short,arrow)

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

	qstr := fmt.Sprintf("INSERT INTO ArrowInverses (Plus,Minus) SELECT %d,%d WHERE NOT EXISTS (SELECT Plus,Minus FROM ArrowInverses WHERE Plus = %d OR minus = %d)",plus,minus,plus,minus)

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

	if link.Wgt == 0 {
		fmt.Println("Attempt to register a link with zero weight is pointless")
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

	qstr := fmt.Sprintf("UPDATE NODE SET %s=array_append(%s,%s) WHERE (NPtr).CPtr = '%d' AND (NPtr).Chan = '%d' AND (%s IS NULL OR NOT %s = ANY(%s))",
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

	// And the reverse arrow

	qstr = fmt.Sprintf("SELECT IdempInsertNodeArrowNode(" +
		"%d," + //infromptr
		"%d," + //infromchan
		"%d," + //isttype
		"%d," + //iarr
		"%.2f," + //iwgt
		"%s," + //ictx
		"%d," + //intoptr
		"%d " + //intochan,
		")",
		dst.Dst.CPtr,
		dst.Dst.Class,
		-sttype,
		INVERSE_ARROWS[dst.Arr],
		dst.Wgt,
		FormatSQLStringArray(dst.Ctx),
		org.CPtr,
		org.Class,)

	row,err = ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("Failed to make inverse node-arrow-node",err,qstr)
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

		"IF array_length(db_set,1) IS NOT NULL THEN\n"+
		"   FOREACH item IN ARRAY db_set LOOP\n" +
		"      db_ref := array_append(db_ref,lower(unaccent(item)));\n" +
		"   END LOOP;\n" +

		"   FOREACH item IN ARRAY user_set LOOP\n" +
		"      IF item = 'any' OR item = '' THEN\n"+
		"        RETURN true;"+
		"      END IF;"+
		"      unicode := replace(item,'|','');\n" +
		"     FOREACH try IN ARRAY db_ref LOOP\n"+
		"        IF length(substring(try from lower(unicode))) > 3 THEN \n" + // unaccented unicode match
	        "           RETURN true;\n" +
		"        END IF;\n" +
		"     END LOOP;"+
		"   END LOOP;\n" +
		"END IF;\n"+
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

	// HELPER

	qstr =  "CREATE OR REPLACE FUNCTION UnCmp(value text,unacc boolean)\n"+
		"RETURNS text AS $fn$\n"+
		"DECLARE \n"+
		"   retval nodeptr[] = ARRAY[]::nodeptr[];\n"+
		"BEGIN\n"+
		//"  RAISE NOTICE 'VALUE= %',value;\n"+
		"  IF unacc THEN\n"+
		"    RETURN lower(unaccent(value)); \n" +
		"  ELSE\n"+
		"    RETURN lower(value); \n" +
		"  END IF;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED UnCmp definition\n",qstr,err)
	}

	row.Close()


	// Find the node that sits at the start/top of a causal chain

	qstr =  "CREATE OR REPLACE FUNCTION GetNCCStoryStartNodes(arrow int,inverse int,sttype int,name text,chapter text,context text[],rm_nm boolean, rm_ch boolean)\n"+
		"RETURNS NodePtr[] AS $fn$\n"+
		"DECLARE \n"+
		"   retval nodeptr[] = ARRAY[]::nodeptr[];\n"+
		"   lowname text = lower(name);"+
		"   lowchap text = lower(chapter);"+
		"BEGIN\n"+
		"     CASE sttype \n"
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"     SELECT array_agg(Nptr) into retval FROM Node WHERE (UnCmp(S,rm_nm) LIKE lower(name)) AND (UnCmp(Chap,rm_ch) LIKE lower(chapter)) AND ArrowInContextList(arrow,%s,context) AND NOT ArrowInContextList(inverse,%s,context);\n",st,STTypeDBChannel(st),STTypeDBChannel(-st));
	}
	qstr += "        ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"     END CASE;\n" +
		"  RETURN retval; \n" +
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

	qstr = "CREATE OR REPLACE FUNCTION AllNCPathsAsLinks(start NodePtr,chapter text,rm_acc boolean,context text[],orientation text,maxdepth INT)\n"+
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
		"path := Format('%s',startlnk::Text);"+
		"ret_paths := SumAllNCPaths(startlnk,path,orientation,1,maxdepth,chapter,rm_acc,context,exclude);" +

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
	qstr = "CREATE OR REPLACE FUNCTION SumAllNCPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,chapter text,rm_acc boolean,context text[],exclude NodePtr[])\n"+
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
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   WHEN orientation = 'fwd' THEN\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   ELSE\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-3);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,1);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,2);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,3);\n" +
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
		"         appendix := SumAllNCPaths(lnk,tot_path,orientation,depth+1,maxdepth,chapter,rm_acc,context,exclude);\n" +

		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"         ELSE\n"+
//		"            ret_paths := tot_path;\n"+
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

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)
        // with a start set of more than one node

	qstr = "CREATE OR REPLACE FUNCTION AllSuperNCPathsAsLinks(start NodePtr[],chapter text,rm_acc boolean,context text[],orientation text,maxdepth INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   root Text;\n" +
		"   path Text;\n"+
		"   node NodePtr;"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = start;\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;"+
		"BEGIN\n" +

		// Aggregate array of starting set
		"FOREACH node IN ARRAY start LOOP\n"+
		"   startlnk := GetSingletonAsLink(node);\n"+
		"   path := Format('%s',startlnk::Text);"+
		"   root := SumAllNCPaths(startlnk,path,orientation,1,maxdepth,chapter,rm_acc,context,exclude);" +
		"   ret_paths := Format('%s\n%s',ret_paths,root);\n"+
		"END LOOP;"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

        // An NC/C filtering version of the neighbour scan

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetNCFwdLinks(start NodePtr,chapter text,rm_acc boolean,context text[],exclude NodePtr[],sttype int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNCNeighboursByType(start,chapter,rm_acc,sttype);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +
		"    neighbours := ARRAY[]::Link[];\n" +
		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+

		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+

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

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetNCCLinks(start NodePtr,exclude NodePtr[],sttype int,chapter text,rm_acc boolean,context text[])\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks =GetNCNeighboursByType(start,chapter,rm_acc,sttype);\n"+

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
	
	qstr = "CREATE OR REPLACE FUNCTION GetNCNeighboursByType(start NodePtr, chapter text,rm_acc boolean,sttype int)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    fwdlinks Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,Array[]::text[],(0,0));\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"     SELECT %s INTO fwdlinks FROM Node WHERE Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter);\n",st,STTypeDBChannel(st));
	}
	
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN fwdlinks; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
        // select GetNCNeighboursByType('(1,116)','chinese',-1);
	
	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
	// **************************************
	// Looking for hub / authority search
	// **************************************
	
	qstr = "CREATE OR REPLACE FUNCTION GetAppointments(arrow int,sttype int,min int,chaptxt text,context text[],with_accents bool)\n"+

		"RETURNS Appointment[] AS $fn$\n" +
		"DECLARE \n" +
		"    app       Appointment;\n" +
		"    appointed Appointment[];\n" +
		"    this      RECORD;" +
		"    thischap  text;" +
		"    arrscalar text;"+
		"    thisarray Link[];" +
		"    count     int;"+
		"    lnk       Link;"+


		"BEGIN\n" +		
		"   CASE sttype \n"
	
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n",st);

		qstr += "   IF with_accents THEN\n"
		// -------------------------------------------------
		qstr += fmt.Sprintf("      FOR this IN SELECT NPtr as thptr,Chap as thchap,%s as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)\n",STTypeDBChannel(st));
		qstr += "      LOOP\n" +
		"         count := 0;\n" +
		"         app.NFrom = null;"+
		"         app.NTo = this.thptr::NodePtr;\n" +
		"         app.Chap = this.thchap;\n" +
		"         app.Arr = arrow;"+
		"         app.STType = sttype;"+
		"         app.Ctx = lnk.Ctx;\n\n" +

		"         IF this.chn::Link[] IS NOT NULL THEN\n"+
		"           FOREACH lnk IN ARRAY this.chn::Link[]\n" +
		"           LOOP\n" +
		"	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx::text[],context) THEN\n" +
		"  	          count = count + 1;\n" +
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              ELSIF arrow < 0 AND match_context(lnk.Ctx::text[],context) THEN\n"+
		"  	          count = count + 1;\n" +
		"                 app.Arr = lnk.Arr;"+
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              END IF;\n" +
		"           END LOOP;\n" +

		"         END IF;\n" +
		
		"         IF count >= min THEN\n" +
		"	    appointed = array_append(appointed,app);\n" +
		"         END IF;\n" +
		"      END LOOP;\n"
		// -------------------------------------------------
		qstr += "   ELSE\n"
		// -------------------------------------------------
		qstr += fmt.Sprintf("      FOR this IN SELECT NPtr as thptr,Chap as thchap,%s as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)\n",STTypeDBChannel(st));
		qstr += "      LOOP\n" +
		"         count := 0;\n" +
		"         app.NFrom = null;"+
		"         app.NTo = this.thptr::NodePtr;\n" +
		"         app.Chap = this.thchap;\n" +
		"         app.Arr = arrow;"+
		"         app.STType = sttype;"+
		"         app.Ctx = lnk.Ctx;\n\n" +

		"         IF this.chn::Link[] IS NOT NULL THEN\n"+
		"           FOREACH lnk IN ARRAY this.chn::Link[]\n" +
		"           LOOP\n" +
		"	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx::text[],context) THEN\n" +
		"  	          count = count + 1;\n" +
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              ELSIF arrow < 0 AND match_context(lnk.Ctx::text[],context) THEN\n"+
		"  	          count = count + 1;\n" +
		"                 app.Arr = lnk.Arr;"+
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              END IF;\n" +
		"           END LOOP;\n" +
		"         END IF;\n" +
		
		"         IF count >= min THEN\n" +
		"	    appointed = array_append(appointed,app);\n" +
		"         END IF;\n" +
		"      END LOOP;\n" +
		// -------------------------------------------------
		"   END IF;\n"
	}
	
	qstr += "END CASE;\n"
	qstr += "    RETURN appointed;\n"
	qstr += "END ;\n"
	qstr += "$fn$ LANGUAGE plpgsql;\n"
	
	row,err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()

}

// **************************************************************************
// Retrieve structural parts
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
	var chapters = make(map[string]int)
	var retval []string

	for row.Next() {		
		err = row.Scan(&whole)
		several := strings.Split(whole,",")

		for s := range several {
			chapters[several[s]]++
		}
	}

	for c := range chapters {
		if strings.Contains(c,src) {
			retval = append(retval,c)
		}
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
		fmt.Println("QUERY GetDBContextssMatchingName",err)
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

func GetDBNodePtrMatchingNCC(ctx PoSST,nm,chap string,cn []string,arrow []ArrowPtr) []NodePtr {

	// Match name, context, chapter, with arrows

/*	if cn == nil && arrow == nil {

		// in case a lazy user hasn't filled out NodeArrowNode
		return GetDBNodePtrMatchingName(ctx,nm,chap)
	}
*/
	var chap_col, nm_col string
	var context string
	var qstr string

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

	qstr = fmt.Sprintf("WITH matching_nodes AS "+
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
		return GetMemoryNodeFromPtr(im_nptr)
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

func GetDBSingletonBySTType(ctx PoSST,sttypes []int,chap string,cn []string) ([]NodePtr,[]NodePtr) {

	var qstr,qwhere string
	var dim = len(sttypes)

	context := FormatSQLStringArray(cn)
	chapter := "%"+chap+"%"

	if dim == 0 || dim > 4 {
		fmt.Println("Maximum 4 sttypes in GetDBSingletonBySTType")
		return nil,nil
	}

	for st := 0; st < len(sttypes); st++ {

		if sttypes[st] < 0 {
			fmt.Println("WARNING! Only give positive STType arguments to GetDBSingletonBySTType as both signs are returned as sources (+) and sinks (-)")
			return nil,nil
		}

		stname := STTypeDBChannel(sttypes[st])
		stinv := STTypeDBChannel(-sttypes[st])
		qwhere += fmt.Sprintf("(array_length(%s::text[],1) IS NOT NULL AND array_length(%s::text[],1) IS NULL AND match_context((%s)[0].Ctx::text[],%s))",stname,stinv,stname,context)
		
		if st != dim-1 {
			qwhere += " OR "
		}
	}

	qstr = fmt.Sprintf("SELECT NPtr FROM Node WHERE lower(Chap) LIKE lower('%s') AND (%s)",chapter,qwhere)

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBSingletonBySTType Failed",err,"IN",qstr)
		return nil,nil
	}

	var src_nptrs,snk_nptrs []NodePtr

	for row.Next() {		
		
		var n NodePtr
		var nstr string
		
		err = row.Scan(&nstr)
		
		if err != nil {
			fmt.Println("Error scanning sql data case",dim,"gave error",err,qstr)
			row.Close()
			return nil,nil
		}
		
		fmt.Sscanf(nstr,"(%d,%d)",&n.Class,&n.CPtr)
		
		src_nptrs = append(src_nptrs,n)
	}
	row.Close()

	// and sinks  -> -

	qwhere = ""

	for st := 0; st < len(sttypes); st++ {

		stname := STTypeDBChannel(-sttypes[st])
		stinv := STTypeDBChannel(sttypes[st])
		qwhere += fmt.Sprintf("(array_length(%s::text[],1) IS NOT NULL AND array_length(%s::text[],1) IS NULL AND match_context((%s)[0].Ctx::text[],%s))",stname,stinv,stname,context)
		
		if st != dim-1 {
			qwhere += " OR "
		}
	}

	qstr = fmt.Sprintf("SELECT NPtr FROM Node WHERE lower(Chap) LIKE lower('%s') AND (%s)",chapter,qwhere)

	row, err = ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBSingletonBySTType 2 Failed",err,"IN",qstr)
		return nil,nil
	}

	for row.Next() {		
		
		var n NodePtr
		var nstr string
		
		err = row.Scan(&nstr)
		
		if err != nil {
			fmt.Println("Error scanning sql data case",dim,"gave error",err,qstr)
			row.Close()
			return nil,nil
		}
		
		fmt.Sscanf(nstr,"(%d,%d)",&n.Class,&n.CPtr)
		
		snk_nptrs = append(snk_nptrs,n)
	}
	row.Close()
	
	return src_nptrs,snk_nptrs
	
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
	var wgt float32

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

	const hits_per_page = 60
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

func GetDBNodeArrowNodeByContexts(ctx PoSST,chap string,cn []string) []NodeArrowNode {

	var qstr string

	_,cn_stripped := IsBracketedSearchList(cn)
	context := FormatSQLStringArray(cn_stripped)

	qstr = fmt.Sprintf("SELECT DISTINCT NFrom,Arr,STType,NTo,Ctx FROM NodeArrowNode WHERE match_context(Ctx,%s)",context)
	row, err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetDBNodesInContexts Failed:",err,qstr)
	}

	var return_value []NodeArrowNode

	var nan NodeArrowNode
	var fromptr NodePtr
	var toptr NodePtr
	var from string
	var to string
	var nctx string
	var arr ArrowPtr
	var sttype int

	for row.Next() {
		nctx = ""
		err = row.Scan(&from,&arr,&sttype,&to,&nctx)

		fmt.Sscanf(from,"(%d,%d)",&fromptr.Class,&fromptr.CPtr)
		nan.NFrom = fromptr
		fmt.Sscanf(to,"(%d,%d)",&toptr.Class,&toptr.CPtr)
		nan.NTo = toptr
		nan.Ctx = ParseSQLArrayString(nctx)
		nan.Arr = arr
		nan.STType = sttype
		return_value = append(return_value,nan)
	}

	row.Close()
	return return_value
}

// **************************************************************************

func GetNodesStartingStoriesForArrow(ctx PoSST,arrow string) ([]NodePtr,int) {

	// Find the head / starting node matching an arrow sequence.
	// It has outgoing (+sttype) but not incoming (-sttype) arrow

	var matches []NodePtr

	arrowptr,sttype := GetDBArrowsWithArrowName(ctx,arrow)

	qstr := fmt.Sprintf("select GetStoryStartNodes(%d,%d,%d)",arrowptr,INVERSE_ARROWS[arrowptr],sttype)
		
	row,err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("GetNodesStartingStoriesForArrow failed\n",qstr,err)
		return nil,0
	}
	
	var nptrstring string
	
	for row.Next() {		
		err = row.Scan(&nptrstring)
		matches = ParseSQLNPtrArray(nptrstring)
	}
	
	row.Close()

	return matches,sttype
}

// **************************************************************************

func GetNCCNodesStartingStoriesForArrow(ctx PoSST,arrow string,name,chapter string,context []string) []NodePtr {

	// Filtered version of function
	// Find the head / starting node matching an arrow sequence.
	// It has outgoing (+sttype) but not incoming (-sttype) arrow

	var matches []NodePtr
	var qstr string

	arrowptr,sttype := GetDBArrowsWithArrowName(ctx,arrow)

	remove_name_accents,nm_stripped := IsBracketedSearchTerm(name)
	remove_chap_accents,chap_stripped := IsBracketedSearchTerm(chapter)

	chp := "%"+chap_stripped+"%"
	nm := "%"+nm_stripped+"%"
	cntx := FormatSQLStringArray(context)

	rm_nm := "false"
	rm_ch := "false"

	if remove_name_accents {
		rm_nm = "true"
	}

	if remove_chap_accents {
		rm_ch = "true"
	}

	// look for _title_ in context

	qstr = fmt.Sprintf("select GetNCCStoryStartNodes(%d,%d,%d,'%s','%s',%s,%s,%s)",arrowptr,INVERSE_ARROWS[arrowptr],sttype,nm,chp,cntx,rm_nm,rm_ch)

	row,err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetNodesNCCStartingStoriesForArrow failed\n",qstr,err)
		return nil
	}
	
	var nptrstring string

	for row.Next() {		
		err = row.Scan(&nptrstring)
		match := ParseSQLNPtrArray(nptrstring)
		matches = append(matches,match...)
	}
	
	row.Close()

	return matches
}

// **************************************************************************
// Retrieve Arrow type data
// **************************************************************************

func GetDBArrowsWithArrowName(ctx PoSST,s string) (ArrowPtr,int) {

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(ctx)
	}

	for a := range ARROW_DIRECTORY {
		if s == ARROW_DIRECTORY[a].Long || s == ARROW_DIRECTORY[a].Short {
			sttype := STIndexToSTType(ARROW_DIRECTORY[a].STAindex)
			return ARROW_DIRECTORY[a].Ptr,sttype
		}
	}

	fmt.Println("No such arrow found in database:",s)
	os.Exit(-1)
	return -1,0
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

	if int(arrowptr) > len(ARROW_DIRECTORY) {
		DownloadArrowsFromDB(ctx)
	}
	
	if int(arrowptr) < len(ARROW_DIRECTORY) {
		a := ARROW_DIRECTORY[arrowptr]
		return a
	} else {
		fmt.Println(ERR_NO_SUCH_ARROW,"(",arrowptr,")")
		os.Exit(-1)
	}
		
	return ARROW_DIRECTORY[arrowptr]

}

// **************************************************************************

func GetDBArrowBySTType(ctx PoSST,sttype int) []ArrowDirectory {

	var retval []ArrowDirectory

	DownloadArrowsFromDB(ctx)

	for a := range ARROW_DIRECTORY {
		sta := ARROW_DIRECTORY[a].STAindex
		if STIndexToSTType(sta) == sttype {
			retval = append(retval,ARROW_DIRECTORY[a])
		}
	}

	return retval
}

//******************************************************************
// Parsing and handling of search strings
//******************************************************************

func ArrowPtrFromArrowsNames(ctx PoSST,arrows []string) ([]ArrowPtr,[]int) {

	// Parse input and discern arrow types, best guess

	var arr []ArrowPtr
	var stt []int

	for a := range arrows {

		// is the entry a number? sttype?

		number, err := strconv.Atoi(arrows[a])
		notnumber := err != nil

		if notnumber {
			arrowptr,_ := GetDBArrowsWithArrowName(ctx,arrows[a])
			if arrowptr != -1 {
				arrdir := GetDBArrowByPtr(ctx,arrowptr)
				arr = append(arr,arrdir.Ptr)
			}
		} else {
			if number < -EXPRESS {
				fmt.Println("Negative arrow value doesn't make sense",number)
			} else if number >= -EXPRESS && number <= EXPRESS {
				stt = append(stt,number)
			} else {
				// whatever remains can only be an arrowpointer
				arrdir := GetDBArrowByPtr(ctx,ArrowPtr(number))
				arr = append(arr,arrdir.Ptr)
			}
		}
	}

	return arr,stt
}

//******************************************************************

func SolveNodePtrs(ctx PoSST,nodenames []string,chap string,cntx []string, arr []ArrowPtr) []NodePtr {

	nodeptrs,rest := ParseLiteralNodePtrs(nodenames)

	var idempotence = make(map[NodePtr]bool)
	var result []NodePtr

	for n := range nodeptrs {
		idempotence[nodeptrs[n]] = true
	}

	for r := range rest {

		nptrs := GetDBNodePtrMatchingNCC(ctx,rest[r],chap,cntx,arr)

		for n := range nptrs {
			idempotence[nptrs[n]] = true
		}
	}

	for uniqnptr := range idempotence {
		result = append(result,uniqnptr)
	}

	return result
}

//******************************************************************

func ParseLiteralNodePtrs(names []string) ([]NodePtr,[]string) {

	var current []rune
	var rest []string
	var nodeptrs []NodePtr

	for n := range names {

		line := []rune(names[n])
		
		for i := 0; i < len(line); i++ {
			
			if line[i] == '(' {
				rs := strings.TrimSpace(string(current))
				if len(rs) > 0 {
					rest = append(rest,string(current))
					current = nil
				}
				continue
			}
			
			if line[i] == ')' {
				np := string(current)
				var nptr NodePtr
				var a,b int = -1,-1
				fmt.Sscanf(np,"%d,%d",&a,&b)
				if a >= 0 && b >= 0 {
					nptr.Class = a
					nptr.CPtr = ClassedNodePtr(b)
					nodeptrs = append(nodeptrs,nptr)
					current = nil
				} else {
					rest = append(rest,"("+np+")")
					current = nil
				}
				continue
			}

			current = append(current,line[i])
		}
		rs := strings.TrimSpace(string(current))

		if len(rs) > 0 {
			rest = append(rest,rs)
		}
		current = nil
	}

	return nodeptrs,rest
}

// **************************************************************************
// Page format, preserving N4L intent
// **************************************************************************

func GetDBPageMap(ctx PoSST,chap string,cn []string,page int) []PageMap {

	var qstr string

	context := FormatSQLStringArray(cn)
	chapter := "%"+chap+"%"

	const hits_per_page = 60
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
// Bulk DB Retrieval
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

	remove_accents,stripped := IsBracketedSearchTerm(chapter)
	chapter = "%"+stripped+"%"
	rm_acc := "false"

	if remove_accents {
		rm_acc = "true"
	}

	qstr := fmt.Sprintf("select AllNCPathsAsLinks from AllNCPathsAsLinks('(%d,%d)','%s',%s,%s,'%s',%d);",
		start.Class,start.CPtr,chapter,rm_acc,FormatSQLStringArray(context),orientation,depth)

	row, err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to AllNCPathsAsLinks Failed",err,qstr)
	}

	var whole string
	var retval [][]Link

	for row.Next() {		
		err = row.Scan(&whole)
		if err != nil {
			fmt.Println("reading AllNCPathsAsLinks",err)
		}

		retval = ParseLinkPath(whole)
	}

	sort.Slice(retval, func(i,j int) bool {
		return len(retval[i]) < len(retval[j])
	})

	row.Close()
	return retval,len(retval)
}

// **************************************************************************

func GetEntireNCSuperConePathsAsLinks(ctx PoSST,orientation string,start []NodePtr,depth int,chapter string,context []string) ([][]Link,int) {
	// orientation should be "fwd" or "bwd" else "both"

	remove_accents,stripped := IsBracketedSearchTerm(chapter)
	chapter = "%"+stripped+"%"
	rm_acc := "false"

	if remove_accents {
		rm_acc = "true"
	}

	qstr := fmt.Sprintf("select AllSuperNCPathsAsLinks(%s,'%s',%s,%s,'%s',%d);",FormatSQLNodePtrArray(start),
		chapter,rm_acc,FormatSQLStringArray(context),orientation,depth)

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
// Bulk retrieval helper functions
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

	ARROW_DIRECTORY = nil
	ARROW_DIRECTORY_TOP = 0

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

func SynchronizeNPtrs(ctx PoSST) {

	// If we're merging (not recommended) N4L into an existing db, we need to synch

	for channel := N1GRAM; channel <= GT1024; channel++ {
		
		qstr := fmt.Sprintf("SELECT max((Nptr).CPtr) FROM Node WHERE (Nptr).Chan=%d",channel)

		row, err := ctx.DB.Query(qstr)
		
		if err != nil {
			fmt.Println("QUERY Synchronizing nptrs",err)
		}

		var cptr int

		for row.Next() {			
			err = row.Scan(&cptr)
			
			if err != nil {
				continue // maybe not defined yet
			}

			if cptr > 0 {

				var empty Node

				// Remember this for uploading later ..
				BASE_DB_CHANNEL_STATE[channel] = ClassedNodePtr(cptr)

				for n := 0; n <= cptr; n++ {

					switch channel {
					case N1GRAM:
						NODE_DIRECTORY.N1_top++
						NODE_DIRECTORY.N1directory = append(NODE_DIRECTORY.N1directory,empty)
					case N2GRAM:
						NODE_DIRECTORY.N2directory = append(NODE_DIRECTORY.N2directory,empty)
						NODE_DIRECTORY.N2_top++
					case N3GRAM:
						NODE_DIRECTORY.N3directory = append(NODE_DIRECTORY.N3directory,empty)
						NODE_DIRECTORY.N3_top++
					case LT128:
						NODE_DIRECTORY.LT128 = append(NODE_DIRECTORY.LT128,empty)
						NODE_DIRECTORY.LT128_top++
					case LT1024:
						NODE_DIRECTORY.LT1024 = append(NODE_DIRECTORY.LT1024,empty)
						NODE_DIRECTORY.LT1024_top++
					case GT1024:
						NODE_DIRECTORY.GT1024 = append(NODE_DIRECTORY.GT1024,empty)
						NODE_DIRECTORY.GT1024_top++
					}
				}
			}
		}
	}

}

// **************************************************************************
// Graphical view, axial geometry
// **************************************************************************

const R0 = 0.4    // radii should not overlap
const R1 = 0.3
const R2 = 0.1

// **************************************************************************

func RelativeOrbit(origin Coords,radius float64,n int,max int) Coords {

	var xyz Coords
	var offset float64

	// splay the vector positions so links not collinear
	switch radius {
	case R1:
		offset = -math.Pi/6.0
	case R2:
		offset = +math.Pi/6.0
	}

	angle := offset + 2 * math.Pi * float64(n)/float64(max)

	xyz.X = origin.X + float32(radius * math.Cos(angle))
	xyz.Y = origin.Y + float32(radius * math.Sin(angle))
	xyz.Z = origin.Z

	return xyz
}

// **************************************************************************

func SetOrbitCoords(xyz Coords,orb [ST_TOP][]Orbit) [ST_TOP][]Orbit {
	
	var r1max,r2max int
	
	// Count all the orbital nodes at this location to calc space
	
	for sti := 0; sti < ST_TOP; sti++ {
		
		for o := range orb[sti] {
			switch orb[sti][o].Radius {
			case 1:
				r1max++
			case 2:
				r2max++
			}
		}
	}
	
	// Place + and - cones on opposite sides, by ordering of sti
	
	var r1,r2 int
	
	for sti := 0; sti < ST_TOP; sti++ {
		
		for o := 0; o < len(orb[sti]); o++ {
			if orb[sti][o].Radius == 1 {
				anchor := RelativeOrbit(xyz,R1,r1,r1max)
				orb[sti][o].OOO = xyz
				orb[sti][o].XYZ = anchor
				r1++
				for op := o+1; op < len(orb[sti]) && orb[sti][op].Radius == 2; op++ {
					orb[sti][op].OOO = anchor
					orb[sti][op].XYZ = RelativeOrbit(anchor,R2,r2,r2max)
					r2++
					o = op-1
				}
			}
		}
	}

	return orb
}

// **************************************************************************

func AssignConeCoordinates(cone [][]Link,nth,swimlanes int) map[NodePtr]Coords {

	var unique = make([][]NodePtr,0)
	var already = make(map[NodePtr]bool)
	var maxlen_tz int

	// If we have multiple cones, each needs a separate name/graph space in X

	if swimlanes == 0 {
		swimlanes = 1
	}

	// Find the longest path length

	for x := 0; x < len(cone); x++ {
		if len(cone[x]) > maxlen_tz {
			maxlen_tz = len(cone[x])
		}
	}

	// Count the expanding wavefront sections for unique node entries

	XChannels := make([]float32,maxlen_tz) // node widths along each path step

	// Find the total number of parallel swimlanes

	for tz := 0; tz < maxlen_tz; tz++ {
		var unique_section = make([]NodePtr,0)
		for x := 0; x < len(cone); x++ {
			if tz < len(cone[x]) {
				if !already[cone[x][tz].Dst] {
					unique_section = append(unique_section,cone[x][tz].Dst)
					already[cone[x][tz].Dst] = true
					XChannels[tz]++
				}
			}
		}
		unique = append(unique,unique_section)
	}

	return MakeCoordinateDirectory(XChannels,unique,maxlen_tz,nth,swimlanes)
}

// **************************************************************************

func AssignStoryCoordinates(axis []Link,nth,swimlanes int,limit int) map[NodePtr]Coords {

	var unique = make([][]NodePtr,0)

	// Nth is segment nth of swimlanes, which has range (width=1.0)/swimlanes * [nth-nth+1]

	if swimlanes == 0 {
		swimlanes = 1
	}

	maxlen_tz := len(axis)

	if limit < maxlen_tz {
		maxlen_tz = limit
	}

	XChannels := make([]float32,maxlen_tz)        // node widths along the path
	already := make(map[NodePtr]bool)

	for tz := 0; tz < maxlen_tz; tz++ {

		var unique_section = make([]NodePtr,0)
		
		if !already[axis[tz].Dst] {
			unique_section = append(unique_section,axis[tz].Dst)
			already[axis[tz].Dst] = true
			XChannels[tz]++
		}

		unique = append(unique,unique_section)
	}

	return MakeCoordinateDirectory(XChannels,unique,maxlen_tz,nth,swimlanes)
}

// **************************************************************************

func AssignPageCoordinates(mapline []Link,nth,swimlanes int) map[NodePtr]Coords {

	XChannels := make([]float32,len(mapline))        // node widths along the path

	var unique = make([][]NodePtr,len(mapline))
	var unique_section = make([]NodePtr,1)

	for tz := 0; tz < len(mapline); tz++ {
		nptr := mapline[tz].Dst
		XChannels[tz] = 1
		unique_section[0] = nptr
		unique[tz] = unique_section
	}

	return MakeCoordinateDirectory(XChannels,unique,len(mapline),nth,swimlanes)
}

// **************************************************************************

func MakeCoordinateDirectory(XChannels []float32, unique [][]NodePtr,maxzlen,nth,swimlanes int) map[NodePtr]Coords {

	var directory = make(map[NodePtr]Coords)

	const totwidth = 2.0 // This is the depth dimenion of the paths -1 to +1
	const arbitrary_elevation = 0.0

	x_lanewidth := totwidth / (float32(swimlanes))
	tz_steplength := totwidth / float32(maxzlen) 

	x_lane_start := float32(nth) * x_lanewidth - totwidth/2.0

	// Start allocating swimlane into XChannels parallel spaces
	// x now runs from (x_lane_start to += x_lanewidth)

	for tz := 0; tz < maxzlen && tz < len(unique); tz++ {

		x_increment := x_lanewidth / (XChannels[tz]+1)

		z_left := -float32(totwidth/2)
		x_left := float32(x_lane_start) + x_increment 

		var xyz Coords

		xyz.X = x_left
		xyz.Y = arbitrary_elevation
		xyz.Z = z_left + tz_steplength * float32(tz)

		for uniqptr := 0; uniqptr < len(unique[tz]); uniqptr++ {
			directory[unique[tz][uniqptr]] = xyz
			xyz.X += x_increment
		}
	}

	return directory
}

// **************************************************************************
// Path integral matrix and coarse graining
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
//
// Part 2: Adjacency matrix representation and graph vector support
//
// **************************************************************************

func GetDBAdjacentNodePtrBySTType(ctx PoSST,sttypes []int,chap string,cn []string,transpose bool) ([][]float32,[]NodePtr) {

	// Return a weighted adjacency matrix by nptr, and an index:nptr lookup table
	// Returns a connected adjacency matrix for the subgraph and a lookup table
	// A bit memory intensive, but possibly unavoidable
	
	var qstr,qwhere,qsearch string
	var dim = len(sttypes)

	context := FormatSQLStringArray(cn)
	chapter := "%"+chap+"%"

	if dim > 4 {
		fmt.Println("Maximum 4 sttypes in GetDBAdjacentNodePtrBySTType")
		return nil,nil
	}

	for st := 0; st < len(sttypes); st++ {

		stname := STTypeDBChannel(sttypes[st])
		qwhere += fmt.Sprintf("array_length(%s::text[],1) IS NOT NULL AND match_context((%s)[0].Ctx::text[],%s)",stname,stname,context)

		if st != dim-1 {
			qwhere += " OR "
		}

		qsearch += "," + stname

	}

	qstr = fmt.Sprintf("SELECT NPtr%s FROM Node WHERE lower(Chap) LIKE lower('%s') AND (%s)",qsearch,chapter,qwhere)

	row, err := ctx.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY GetDBAdjacentNodePtrBySTType Failed",err)
		return nil,nil
	}

	var linkstr = make([]string,dim+1)
	var protoadj = make(map[int][]Link)
	var lookup = make(map[NodePtr]int)
	var rowindex int
	var nodekey []NodePtr
	var counter int

	for row.Next() {		

		var n NodePtr
		var nstr string

		switch dim {

		case 1: err = row.Scan(&nstr,&linkstr[0])
		case 2: err = row.Scan(&nstr,&linkstr[0],&linkstr[1])
		case 3: err = row.Scan(&nstr,&linkstr[0],&linkstr[1],&linkstr[2])
		case 4: err = row.Scan(&nstr,&linkstr[0],&linkstr[1],&linkstr[2],&linkstr[3])

		default:
			fmt.Println("Maximum 4 sttypes in GetDBAdjacentNodePtrBySTType - shouldn't happen")
			row.Close()
			return nil,nil
		}

		if err != nil {
			fmt.Println("Error scanning sql data case",dim,"gave error",err,qstr)
			row.Close()
			return nil,nil
		}

		fmt.Sscanf(nstr,"(%d,%d)",&n.Class,&n.CPtr)

		// idempotently gather nptrs into a map, keeping linked nodes close in order

		index,already := lookup[n]

		if already {
			rowindex = index
		} else {
			rowindex = counter
			lookup[n] = counter
			counter++
			nodekey = append(nodekey,n)
		}

		// Run through the nodes linked and add them now

		for lnks := range linkstr {

			links := ParseMapLinkArray(linkstr[lnks])

			// we have to go through one by one to avoid duplicates
			// and keep adjacent nodes closer in order
			
			for l := range links {	
				_,already := lookup[links[l].Dst]
				
				if !already {
					lookup[links[l].Dst] = counter
					counter++
					nodekey = append(nodekey,links[l].Dst)
				}
			}
			// Now we have a vector row for each NPtr, with a list of links
			protoadj[rowindex] = append(protoadj[rowindex],links...)
		}
	}

	// Now we know the dimension of the square matrix = counter
        // and an ordered directory vector[index] ->  NPtr, as well as lookup table
	// So we assemble the adjacency matrix (or its transpose on request)

	adj := make([][]float32,counter)

	for r := 0; r < counter; r++ {

		adj[r] = make([]float32,counter)

		row := protoadj[r]

		for l := 0; l < len(row); l++ {

			lnk := row[l]
			c := lookup[lnk.Dst]

			if transpose {
				adj[c][r] = lnk.Wgt
			} else {
				adj[r][c] = lnk.Wgt
			}
		}
	}
	
	row.Close()
	return adj,nodekey
}

// **************************************************************************

func SymbolMatrix(m [][]float32) [][]string {
	
	var symbol [][]string
	dim := len(m)

	for r := 0; r < dim; r++ {

		var srow []string
		
		for c := 0; c < dim; c++ {

			var sym string = ""

			if m[r][c] != 0 {
				sym = fmt.Sprintf("%d*%d",r,c)
			}
			srow = append(srow,sym)
		}
		symbol = append(symbol,srow)
	}
	return symbol
}

//**************************************************************

func SymbolicMultiply(m1,m2 [][]float32,s1,s2 [][]string) ([][]float32,[][]string) {

	// trace the elements in a multiplication for path mapping

	var m [][]float32
	var sym [][]string

	dim := len(m1)

	for r := 0; r < dim; r++ {

		var newrow []float32
		var symrow []string

		for c := 0; c < dim; c++ {

			var value float32
			var symbols string

			for j := 0; j < dim; j++ {

				if  m1[r][j] != 0 && m2[j][c] != 0 {
					value += m1[r][j] * m2[j][c]
					symbols += fmt.Sprintf("%s*%s",s1[r][j],s2[j][c])
				}
			}
			newrow = append(newrow,value)
			symrow = append(symrow,symbols)

		}
		m  = append(m,newrow)
		sym  = append(sym,symrow)
	}

	return m,sym
}

//**************************************************************

func GetSparseOccupancy(m [][]float32,dim int) []int {

	var sparse_count = make([]int,dim)

	for r := 0; r < dim; r++ {
		for c := 0; c < dim; c++ {
			sparse_count[r]+= int(m[r][c])
		}
	}

	return sparse_count
}

//**************************************************************

func SymmetrizeMatrix(m [][]float32) [][]float32 {

	// CAUTION! unless we make a copy, go actually changes the original m!!! :o
	// There is some very weird pathological memory behaviour here .. but this
	// workaround seems to be stable

	var dim int = len(m)
	var symm [][]float32 = make([][]float32,dim)

	for r := 0; r < dim; r++ {
		var row []float32 = make([]float32,dim)
		symm[r] = row
	}
	
	for r := 0; r < dim; r++ {
		for c := r; c < dim; c++ {
			v := m[r][c]+m[c][r]
			symm[r][c] = v
			symm[c][r] = v
		}
	}

	return symm
}

//**************************************************************

func TransposeMatrix(m [][]float32) [][]float32 {

	var dim int = len(m)
	var mt [][]float32 = make([][]float32,dim)

	for r := 0; r < dim; r++ {
		var row []float32 = make([]float32,dim)
		mt[r] = row
	}

	for r := 0; r < len(m); r++ {
		for c := r; c < len(m); c++ {

			v := m[r][c]
			vt := m[c][r]
			mt[r][c] = vt
			mt[c][r] = v
		}
	}

	return mt
}

//**************************************************************

func MakeInitVector(dim int,init_value float32) []float32 {

	var v = make([]float32,dim)

	for r := 0; r < dim; r++ {
		v[r] = init_value
	}

	return v
}

//**************************************************************

func MatrixOpVector(m [][]float32, v []float32) []float32 {

	var vp = make([]float32,len(m))

	for r := 0; r < len(m); r++ {
		for c := 0; c < len(m); c++ {

			if m[r][c] != 0 {
				vp[r] += m[r][c] * v[c]
			}
		}
	}
	return vp
}

//**************************************************************

func ComputeEVC(adj [][]float32) []float32 {

	v := MakeInitVector(len(adj),1.0)
	vlast := v

	const several = 10

	for i := 0; i < several; i++ {

		v = MatrixOpVector(adj,vlast)

		if CompareVec(v,vlast) < 0.01 {
			break
		}
		vlast = v
	}

	maxval,_ := GetVecMax(v)
	v = NormalizeVec(v,maxval)
	return v
}

//**************************************************************

func GetVecMax(v []float32) (float32,int) {

	var max float32 = -1
	var index int

	for r := range v {
		if v[r] > max {
			max = v[r]
			index = r
		}
	}

	return max,index
}

//**************************************************************

func NormalizeVec(v []float32, div float32) []float32 {

	if div == 0 {
		div = 1
	}

	for r := range v {
		v[r] = v[r] / div
	}

	return v
}

//**************************************************************

func CompareVec(v1,v2 []float32) float32 {

	var max float32 = -1

	for r := range v1 {
		diff := v1[r]-v2[r]

		if diff < 0 {
			diff = -diff
		}

		if diff > max {
			max = diff
		}
	}

	return max
}

//**************************************************************

func FindGradientFieldTop(sadj [][]float32,evc []float32) (map[int][]int,[]int,[][]int) {

	// Hill climbing gradient search

	dim := len(evc)

	var localtop []int
	var paths [][]int
	var regions = make(map[int][]int)

	for index := 0; index < dim; index++ {

		// foreach neighbour

		ltop,path := GetHillTop(index,sadj,evc)

		regions[ltop] = append(regions[ltop],index)
		localtop = append(localtop,ltop)
		paths = append(paths,path)
	}

	return regions,localtop,paths
}

//**************************************************************

func GetHillTop(index int,sadj [][]float32,evc []float32) (int,[]int) {

	topnode := index
	visited := make(map[int]bool)
	visited[index] = true

	var path []int

	dim := len(evc)
	finished := false
	path = append(path,index)

	for {
		finished = true
		winner := topnode
		
		for ngh := 0; ngh < dim; ngh++ {
			
			if (sadj[topnode][ngh] > 0) && !visited[ngh] {
				visited[ngh] = true
				
				if evc[ngh] > evc[topnode] {
					winner = ngh
					finished = false
				}
			}
		}
		if finished {
			break
		}

		topnode = winner
		path = append(path,topnode)
	}

	return topnode,path
}

// **************************************************************************
// Matrix/Path tools
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
//
// Part 3: Model data retrieval, and data marshalling, with JSON etc
//
// **************************************************************************

func GetSequenceContainers(ctx PoSST,arrname string,search,chapter string,context []string, limit int) []Story {

	var stories []Story

	if arrname == "" {
		arrname = "then"
	}

	var count int

	arrowptr,_ := GetDBArrowsWithArrowName(ctx,arrname)
	openings := GetNCCNodesStartingStoriesForArrow(ctx,arrname,search,chapter,context)

	for nth := range openings {

		var story Story

		node := GetDBNodeByNodePtr(ctx,openings[nth])

		story.Chapter = node.Chap

		axis := GetLongestAxialPath(ctx,openings[nth],arrowptr)

		directory := AssignStoryCoordinates(axis,nth,len(openings),limit)

		for lnk := 0; lnk < len(axis); lnk++ {
			
			// Now add the orbit at this node, not including the axis
			var ne NodeEvent
			nd := GetDBNodeByNodePtr(ctx,axis[lnk].Dst)
			ne.Text = nd.S
			ne.L = nd.L
			ne.Chap = nd.Chap
			ne.NPtr = axis[lnk].Dst
			ne.XYZ = directory[ne.NPtr]
			ne.Orbits = GetNodeOrbit(ctx,axis[lnk].Dst,arrname)
			ne.Orbits = SetOrbitCoords(ne.XYZ,ne.Orbits)

			if lnk > limit {
				break
			}

			story.Axis = append(story.Axis,ne)
		}

		if story.Axis != nil {
			stories = append(stories,story)
			count ++
		}

		count++

		if count > limit {
			return stories
		}
		
	}

	return stories
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

		for angle := 0; angle < len(sweep); angle++ {

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

//******************************************************************

func ContextIntentAnalysis(spectrum map[string]int,clusters []string) ([]string,[]string) {

	var intentional []string
	const intent_limit = 3  // policy from research

	for f := range spectrum {
		if spectrum[f] < intent_limit {
			intentional = append(intentional,f)
			delete(spectrum,f)
		}
	}

	for cl := range clusters {
		for deletions := range intentional {
			clusters[cl] = strings.Replace(clusters[cl],intentional[deletions]+", ","",-1)
			clusters[cl] = strings.Replace(clusters[cl],intentional[deletions],"",-1)
		}
	}

	spectrum = make(map[string]int)

	for cl := range clusters {
		if strings.TrimSpace(clusters[cl]) != "" {
			pruned := strings.Trim(clusters[cl],", ")
			spectrum[pruned]++
		}
	}

	// Now we have a small set of largely separated major strings.
	// One more round of diffs for a final separation

	var ambient = make(map[string]int)

	context := Map2List(spectrum)

	for ci := 0; ci < len(context); ci++ {
		for cj := ci+1; cj < len(context); cj++ {

			s,i := DiffClusters(context[ci],context[cj])

			if len(s) > 0 && len(i) > 0 && len(i) <= len(context[ci])+len(context[ci]) {
				ambient[strings.TrimSpace(s)]++
				ambient[strings.TrimSpace(i)]++
			}
		}
	}
	
	return intentional,Map2List(ambient)
}

// **************************************************************************

func IntersectContextParts(context_clusters []string) (int,[]string,[][]int)  {

	var idemp = make(map[string]int)
	var cluster_list []string

	for s := range context_clusters {
		idemp[context_clusters[s]]++
	}

	for each_unique_cluster := range idemp {
		cluster_list = append(cluster_list,each_unique_cluster)
	}

	sort.Strings(cluster_list)

	var adj [][]int

	for ci := 0; ci < len(cluster_list); ci++ {

		var row []int

		for cj := ci+1; cj < len(cluster_list); cj++ {			
			s,_ := DiffClusters(cluster_list[ci],cluster_list[cj])
			row = append(row,len(s))
		}

		adj = append(adj,row)
	}

	return len(cluster_list),cluster_list,adj
}

// **************************************************************************
// These functions are about text fractionation of the context strings
// which is similar to text2N4L scanning but applied to lists of phrases
// on a much smaller scale. Still looking for "mass spectrum" of fragments ..
// **************************************************************************

func DiffClusters(l1,l2 string) (string,string) {

	// The fragments arrive as comma separated strings that are
        // already composed or ordered n-grams

	spectrum1 := strings.Split(l1,", ")
	spectrum2 := strings.Split(l2,", ")

	// Get orderless idempotent directory of all 1-grams

	m1 := List2Map(spectrum1)
	m2 := List2Map(spectrum2)

	// split the lists into words into directories for common and individual ngrams

	return OverlapMatrix(m1,m2)
}

// **************************************************************************

func OverlapMatrix(m1,m2 map[string]int) (string,string) {

	var common = make(map[string]int)
	var separate = make(map[string]int)

	// sieve shared / individual parts

	for ng := range m1 {
		if m2[ng] > 0 {
			common[ng]++
		} else {
			separate[ng]++
		}
	}

	for ng := range m2 {
		if m1[ng] > 0 {
			delete(separate,ng)
			common[ng]++
		} else {
			_,exists := common[ng]
			if  !exists {
				separate[ng]++
			}
		}
	}

	return List2String(Map2List(common)),List2String(Map2List(separate))
}

// **************************************************************************

func GetContextTokenFrequencies(fraglist []string) map[string]int {

	var spectrum = make(map[string]int)

	for l := range fraglist {
		fragments := strings.Split(fraglist[l],", ")
		partial := List2Map(fragments)

		// Merge all strands

		for f := range partial {
			spectrum[f] += partial[f]
		}
	}

	return spectrum
}

// **************************************************************************
// Presentation on command line
// **************************************************************************

func PrintNodeOrbit(ctx PoSST, nptr NodePtr,width int) {

	node := GetDBNodeByNodePtr(ctx,nptr)		

	ShowText(node.S,width)
	fmt.Println("\tin chapter:",node.Chap)
	fmt.Println()

	notes := GetNodeOrbit(ctx,nptr,"")

	PrintLinkOrbit(notes,EXPRESS,0)
	PrintLinkOrbit(notes,-EXPRESS,0)
	PrintLinkOrbit(notes,-CONTAINS,0)
	PrintLinkOrbit(notes,LEADSTO,0)
	PrintLinkOrbit(notes,-LEADSTO,0)
	PrintLinkOrbit(notes,NEAR,0)

	fmt.Println()
}

// **************************************************************************

func PrintLinkOrbit(notes [ST_TOP][]Orbit,sttype int,indent_level int) {

	t := STTypeToSTIndex(sttype)

	for n := range notes[t] {		

		r := notes[t][n].Radius + indent_level

		if notes[t][n].Ctx != "" {
			txt := fmt.Sprintf(" -    (%s) - %s  \t.. in the context of %s\n",notes[t][n].Arrow,notes[t][n].Text,notes[t][n].Ctx)
			text := Indent(LEFTMARGIN * r) + txt
			ShowText(text,SCREENWIDTH)
		} else {
			txt := fmt.Sprintf(" -    (%s) - %s\n",notes[t][n].Arrow,notes[t][n].Text)
			text := Indent(LEFTMARGIN * r) + txt
			ShowText(text,SCREENWIDTH)
		}

	}

}

// **************************************************************************

func PrintLinkPath(ctx PoSST, cone [][]Link, p int, prefix string,chapter string,context []string) {

	PrintSomeLinkPath(ctx,cone, p,prefix,chapter,context,10000)
}

// **************************************************************************

func PrintSomeLinkPath(ctx PoSST, cone [][]Link, p int, prefix string,chapter string,context []string,limit int) {

	count := 0

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

			count++

			if count > limit {
				return
			}

			if !start_shown {
				if len(cone) > 1 {
					fmt.Printf("%s (%d) %s",prefix,p+1,path_start.S)
				} else {
					fmt.Printf("%s %s",prefix,path_start.S)
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

		fmt.Print("\n     -  [ Link STTypes:")

		for s := range stpath {
			fmt.Print(" -(",stpath[s],")-> ")
		}
		fmt.Println(". ]\n")
	}
}

// **************************************************************************
// Presentation in JSON
// **************************************************************************

func JSONNodeEvent(ctx PoSST, nptr NodePtr,xyz Coords,orbits [ST_TOP][]Orbit) string {

	node := GetDBNodeByNodePtr(ctx,nptr)

	var event NodeEvent
	event.Text = node.S
	event.L = node.L
	event.Chap = node.Chap
	event.NPtr = nptr
	event.XYZ = xyz
	event.Orbits = orbits

	jstr,_ := json.Marshal(event)

	return strings.ReplaceAll(string(jstr),"null","[]")
}

// **************************************************************************

func LinkWebPaths(ctx PoSST,cone [][]Link,nth int,chapter string,context []string,swimlanes,limit int) [][]WebPath {

	// This is dealing in good faith with one of swimlanes cones, assigning equal width to all
	// The cone is a flattened array, we can assign spatial coordinates for visualization

	var conepaths [][]WebPath

	directory := AssignConeCoordinates(cone,nth,swimlanes)

	// JSONify the cone structure, converting []Link into []WebPath

	for p := 0; p < len(cone); p++ {

		path_start := GetDBNodeByNodePtr(ctx,cone[p][0].Dst)		
		
		start_shown := false

		var path []WebPath
		
		for l := 1; l < len(cone[p]); l++ {

			if !MatchContexts(context,cone[p][l].Ctx) {
				break
			}

			nextnode := GetDBNodeByNodePtr(ctx,cone[p][l].Dst)

			if !SimilarString(nextnode.Chap,chapter) {
				break
			}
			
			if !start_shown {
				var ws WebPath
				ws.Name = path_start.S
				ws.NPtr = cone[p][0].Dst
				ws.XYZ = directory[cone[p][0].Dst]
				path = append(path,ws)
				start_shown = true
			}

			arr := GetDBArrowByPtr(ctx,cone[p][l].Arr)
	
			if l < len(cone[p]) {
				var wl WebPath
				wl.Name = arr.Long
				wl.Arr = cone[p][l].Arr
				wl.STindex = arr.STAindex
				wl.XYZ = directory[cone[p][l].Dst]
				path = append(path,wl)
			}

			var wn WebPath
			wn.Name = nextnode.S
			wn.NPtr = cone[p][l].Dst
			wn.XYZ = directory[cone[p][l].Dst]
			path = append(path,wn)

		}
		conepaths = append(conepaths,path)
	}

	return conepaths
}

// **************************************************************************

func GetChaptersByChapContext(ctx PoSST,chap string,cn []string,limit int) map[string][]string {

	qstr := ""
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

	if chap == "TableOfContents" {
		chap_col = ""
	}

	_,cn_stripped := IsBracketedSearchList(cn)
	context := FormatSQLStringArray(cn_stripped)

	qstr = fmt.Sprintf("WITH matching_nodes AS "+
		"  (SELECT NFrom,ctx,match_context(ctx,%s) AS match FROM NodeArrowNode)"+
		"     SELECT DISTINCT chap,ctx FROM matching_nodes "+
		"      JOIN Node ON nptr=nfrom WHERE match=true %s ORDER BY Chap",
		context,chap_col)

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetChaptersByChapContext Failed",err,qstr)
	}

	var rchap,rcontext string
	var toc = make(map[string][]string)

	for row.Next() {		
		err = row.Scan(&rchap,&rcontext)

		// Each chapter can be a comma separated list

		chps := SplitChapters(rchap)

		for c := 0; c < len(chps); c++ {

			if len(toc) == limit {
				row.Close()
				return toc
			}

			rc := chps[c]

			cn := ParseSQLArrayString(rcontext)
			ctx_grp := ""

			for s := 0; s < len(cn); s++ {
				ctx_grp += cn[s]
				if s < len(cn)-1 {
					ctx_grp += ", "
				}
			}

			if len(ctx_grp) > 0 {
				toc[rc] = append(toc[rc],ctx_grp)
			}
		}
	}

	row.Close()
	return toc
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

		directory := AssignPageCoordinates(maplines[n].Path,n,len(maplines))

		// Next line item

		for lnk := 0; lnk < len(maplines[n].Path); lnk++ {
			
			text := GetDBNodeByNodePtr(ctx,maplines[n].Path[lnk].Dst)
			
			if lnk == 0 {
				var ws WebPath
				ws.Name = text.S
				ws.NPtr = maplines[n].Path[lnk].Dst
				ws.XYZ = directory[ws.NPtr]
				path = append(path,ws)
				
			} else {// ARROW
				arr := GetDBArrowByPtr(ctx,maplines[n].Path[lnk].Arr)
				var wl WebPath
				wl.Name = arr.Long
				wl.Arr = maplines[n].Path[lnk].Arr
				wl.STindex = arr.STAindex
				path = append(path,wl)
				// NODE
				var ws WebPath
				ws.Name = text.S
				ws.NPtr = maplines[n].Path[lnk].Dst
				ws.XYZ = directory[ws.NPtr]
				path = append(path,ws)
				
			}
		}
		// Next line
		webnotes.Notes = append(webnotes.Notes,path)
	}
	
	encoded, _ := json.Marshal(webnotes)
	jstr := fmt.Sprintf("%s",string(encoded))

	return jstr
}

// **************************************************************************
// Retrieve cluster Analysis
// **************************************************************************

func GetAppointedNodesByArrow(ctx PoSST,arrow ArrowPtr,cn []string,chap string,size int) map[ArrowPtr][]Appointment {

	// return a map of all the nodes in chap,context that are pointed to by the same type of arrow
        // grouped by arrow

	reverse_arrow := INVERSE_ARROWS[arrow]
	arr := GetDBArrowByPtr(ctx,reverse_arrow)
	sttype := STIndexToSTType(arr.STAindex)

	_,cn_stripped := IsBracketedSearchList(cn)
	context := FormatSQLStringArray(cn_stripped)

	var chap_col,chap_stripped string
	var remove_chap_accents bool

	if chap != "any" && chap != "" {	
		remove_chap_accents,chap_stripped = IsBracketedSearchTerm(chap)
		
		if remove_chap_accents {
			chap_col = "%"+chap_stripped+"%"
		} else {
			chap_col = "%"+chap+"%"
		}
	}

	qstr := fmt.Sprintf("SELECT unnest(GetAppointments(%d,%d,%d,'%s',%s,%v))",int(reverse_arrow),sttype,size,chap_col,context,remove_chap_accents)

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointedNodesByArrow Failed",err,qstr)
	}

	var whole string

	var retval = make(map[ArrowPtr][]Appointment)
	
	for row.Next() {
		err = row.Scan(&whole) //arrint,&sttype,&rchap,&rctx,&apex,&arry)

		next := ParseAppointedNodeCluster(whole)
		retval[next.Arr] = append(retval[next.Arr],next)
	}
	
	row.Close()
	
	return retval
}

// **************************************************************************

func GetAppointedNodesBySTType(ctx PoSST,sttype int,cn []string,chap string,size int) map[ArrowPtr][]Appointment {

	// return a map of all the nodes in chap,context that are pointed to by the same type of arrow
        // grouped by arrow

	_,cn_stripped := IsBracketedSearchList(cn)
	context := FormatSQLStringArray(cn_stripped)

	var chap_col,chap_stripped string
	var remove_chap_accents bool

	if chap != "any" && chap != "" {	
		remove_chap_accents,chap_stripped = IsBracketedSearchTerm(chap)
		
		if remove_chap_accents {
			chap_col = "%"+chap_stripped+"%"
		} else {
			chap_col = "%"+chap+"%"
		}
	}

	qstr := fmt.Sprintf("SELECT unnest(GetAppointments(%d,%d,%d,'%s',%s,%v))",-1,sttype,size,chap_col,context,remove_chap_accents)

	row, err := ctx.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointedNodesByArrow Failed",err,qstr)
	}

	var whole string

	var retval = make(map[ArrowPtr][]Appointment)
	
	for row.Next() {
		err = row.Scan(&whole) //arrint,&sttype,&rchap,&rctx,&apex,&arry)

		next := ParseAppointedNodeCluster(whole)
		retval[next.Arr] = append(retval[next.Arr],next)
	}
	
	row.Close()
	
	return retval
}

// **************************************************************************

func ParseAppointedNodeCluster(whole string) Appointment {

    //  (13,-1,maze,{},"(1,3122)","{""(1,3121)"",""(1,3138)""}")

	var next Appointment
      	var l []string

    	whole = strings.Trim(whole,"(")
    	whole = strings.Trim(whole,")")

	uni_array := []rune(whole)

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

	var arrp ArrowPtr
	fmt.Sscanf(l[0],"%d",&arrp)
	fmt.Sscanf(l[1],"%d",&next.STType)

	// invert arrow
	next.Arr = INVERSE_ARROWS[ArrowPtr(arrp)]
	next.STType = -next.STType

	next.Chap = l[2]
	next.Ctx = ParseSQLArrayString(l[3])

	fmt.Sscanf(l[4],"(%d,%d)",&next.NTo.Class,&next.NTo.CPtr)

	// Postgres is inconsistent in adding \" to arrays (hack)

	l[5] = strings.Replace(l[5],"(","\"(",-1)
	l[5] = strings.Replace(l[5],")",")\"",-1)
	next.NFrom = ParseSQLNPtrArray(l[5])

	return next
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
		betw = fmt.Sprintf("%.2f : ",float32(order[key])/float32(len(solutions)))
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

// ******************************************************************
//
// Part 4 : SEARCH LANGUAGE
//
// ******************************************************************

type SearchParameters struct {

	Name     []string
	From     []string
	To       []string
	Chapter  string
	Context  []string
	Arrows   []string
	PageNr   int
	Range    int
	Sequence bool
}

const (

	CMD_ON = "on"
	CMD_FOR = "for"
	CMD_ABOUT = "about"
	CMD_NOTES = "note"
	CMD_PAGE = "page"
	CMD_PATH = "path"
	CMD_SEQ = "seq"
	CMD_FROM = "from"
	CMD_TO = "to"
	CMD_CTX = "ctx"
	CMD_CONTEXT = "context"
	CMD_AS = "as"
	CMD_CHAPTER = "chapter"
	CMD_SECTION = "section"
	CMD_IN = "in"
	CMD_ARROW = "arrow"
	CMD_LIMIT = "limit"
	CMD_DEPTH = "depth"
	CMD_RANGE = "range"
	CMD_DISTANCE = "distance"
)

//******************************************************************

func DecodeSearchField(cmd string) SearchParameters {

	var keywords = []string{ 
		CMD_NOTES, CMD_PATH,
		CMD_PATH,CMD_FROM,CMD_TO,
		CMD_SEQ,
		CMD_CONTEXT,CMD_CTX,CMD_AS,
		CMD_CHAPTER,CMD_IN,CMD_SECTION,
		CMD_ARROW,
		CMD_ON,CMD_ABOUT,CMD_FOR,
		CMD_PAGE,
		CMD_LIMIT,CMD_RANGE,CMD_DISTANCE,CMD_DEPTH,
        }
	
	// parentheses are reserved for unaccenting

	m := regexp.MustCompile("[ \t]+") 
	cmd = m.ReplaceAllString(cmd," ") 

	cmd = strings.TrimSpace(cmd)
	pts := SplitQuotes(cmd)

	var parts [][]string
	var part []string

	for p := 0; p < len(pts); p++ {

		subparts := SplitQuotes(pts[p])

		for w := 0; w < len(subparts); w++ {

			if InList(subparts[w],keywords) {
				// special case for TO with implicit FROM, and USED AS
				if p > 0 && subparts[w] == "to" {
					part = append(part,subparts[w])
					continue
				}
				if w > 0 && strings.HasPrefix(subparts[w],"to") {
					part = append(part,subparts[w])
				} else {
					parts = append(parts,part)
					part = nil
					part = append(part,subparts[w])
				}
			} else {
				// Try to override command line splitting behaviour
				part = append(part,subparts[w])
			}
		}
	}

	parts = append(parts,part) // add straggler to complete

	// command is now segmented

	param := FillInParameters(parts,keywords)

	return param
}

//******************************************************************

func FillInParameters(cmd_parts [][]string,keywords []string) SearchParameters {

	var param SearchParameters 

	for c := 0; c < len(cmd_parts); c++ {

		lenp := len(cmd_parts[c])

		for p := 0; p < lenp; p++ {

			switch SomethingLike(cmd_parts[c][p],keywords) {

			case CMD_CHAPTER, CMD_IN:
				if lenp > p+1 {
					param.Chapter = cmd_parts[c][p+1]
					break
				} else {
					param.Chapter = "TableOfContents"
					break					
				}
				continue

			case CMD_NOTES:
				if param.PageNr < 1 {
					param.PageNr = 1
				}
			
				if lenp > p+1 && lenp == 2 {
					param.Chapter = cmd_parts[c][p+1]
				} else {
					if lenp > 1 {
						param = AddOrphan(param,cmd_parts[c][p])
					}
				}
				continue

			case CMD_PAGE:
				// if followed by a number, else could be search term
				if lenp > p+1 {
					p++
					var no int = -1
					fmt.Sscanf(cmd_parts[c][p],"%d",&no)
					if no > 0 {
						param.PageNr = no
					} else {
						param = AddOrphan(param,cmd_parts[c][p-1])
						param = AddOrphan(param,cmd_parts[c][p])
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_RANGE,CMD_DEPTH,CMD_LIMIT,CMD_DISTANCE:
				// if followed by a number, else could be search term
				if lenp > p+1 {
					p++
					var no int = -1
					fmt.Sscanf(cmd_parts[c][p],"%d",&no)
					if no > 0 {
						param.Range = no
					} else {
						param = AddOrphan(param,cmd_parts[c][p-1])
						param = AddOrphan(param,cmd_parts[c][p])
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_ARROW:
				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							param.Arrows = append(param.Arrows,DeQ(ult[u]))
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue
				
			case CMD_CONTEXT,CMD_CTX,CMD_AS:

				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							param.Context = append(param.Context,DeQ(ult[u]))
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_FROM:
				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							param.From = append(param.From,DeQ(ult[u]))
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_TO:
				if p > 0 && lenp > p+1 {
					if param.From == nil {
						param.From = append(param.From,cmd_parts[c][p-1])
					}

					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							param.To = append(param.To,DeQ(ult[u]))
						}
					}
					continue
				}
				// TO is too short to be an independent search term

				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							param.To = append(param.To,DeQ(ult[u]))
						}
					}
					continue
				}

			case CMD_PATH,CMD_SEQ:
				param.Sequence = true
				continue

			case CMD_ON,CMD_ABOUT,CMD_FOR:
				if lenp > p+1 {

					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp]," ")
						for u := range ult {
							param.Name = append(param.Name,DeQ(ult[u]))
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			default:

				if lenp > p+1 && cmd_parts[c][p+1] == CMD_TO {
					continue
				}

				for pp := p; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
					p++
					ult := SplitQuotes(cmd_parts[c][pp])
					for u := range ult {
						param.Name = append(param.Name,DeQ(ult[u]))
					}
				}
				continue
			}
			break
		}
	}

	return param
}

//******************************************************************

func IsParam(i,lenp int,keys []string,keywords []string) bool {

	// Make sure the next item is not the start of a new token

	if i >= lenp {
		return false
	}

	key := keys[i]

	for k := 0; k < len(keywords); k++ {
		if strings.HasPrefix(key,keywords[k]) {
			return false
		}
	}

	return true
}

//******************************************************************

func SomethingLike(s string,keywords []string) string {

	for k := 0; k < len(keywords); k++ {
		if strings.HasPrefix(s,keywords[k]) {
			return keywords[k]
		}
	}

	return s
}

//******************************************************************

func AddOrphan(param SearchParameters,orphan string) SearchParameters {

	// if a keyword isn't followed by the right param it was possibly
	// intended as a search term not a command, so add back

	if param.To != nil {
		param.To = append(param.To,orphan)
		return param
	}

	if param.From != nil {
		param.From = append(param.From,orphan)
		return param
	}

	param.Name = append(param.Name,orphan)

	return param
}

//******************************************************************

func InList(s string,list []string) bool {

	for w := range list {
		if strings.HasPrefix(s,list[w]) {
			return true
		}
	}
	return false
}

//******************************************************************

func SplitQuotes(s string) []string {

	var items []string
	var upto []rune
	cmd := []rune(s)

	for r := 0; r < len(cmd); r++ {

		if IsQuote(cmd[r]) {
			if len(upto) > 0 {
				items = append(items,string(upto))
			}

			qstr,offset := ReadToNext(cmd,r,cmd[r])

			if len(qstr) > 0 {
				items = append(items,qstr)
				r += offset
			}
			continue
		}

		switch cmd[r] {
		case ' ':
			if len(upto) > 0 {
				items = append(items,string(upto))
			}
			upto = nil
			continue

		case '(':
			if len(upto) > 0 {
				items = append(items,string(upto))
			}

			qstr,offset := ReadToNext(cmd,r,')')

			if len(qstr) > 0 {
				items = append(items,qstr)
				r += offset
			}
			continue

		}

		upto = append(upto,cmd[r])
	}

	items = append(items,string(upto))

	return items
}

// **************************************************************************

func DeQ(s string) string {

	return strings.Trim(s,"\"")
}

// **************************************************************************
//
// Part 5: Context processing
//
// **************************************************************************

// ****************************************************************************
// Semantic 2D time
// ****************************************************************************

var GR_DAY_TEXT = []string{
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday",
        "Sunday",
    }
        
var GR_MONTH_TEXT = []string{
        "January",
        "February",
        "March",
        "April",
        "May",
        "June",
        "July",
        "August",
        "September",
        "October",
        "November",
        "December",
}
        
var GR_SHIFT_TEXT = []string{
        "Night",
        "Morning",
        "Afternoon",
        "Evening",
    }

// For second resolution Unix time

const CF_MONDAY_MORNING = 345200
const CF_MEASURE_INTERVAL = 5*60
const CF_SHIFT_INTERVAL = 6*3600

const MINUTES_PER_HOUR = 60
const SECONDS_PER_MINUTE = 60
const SECONDS_PER_HOUR = (60 * SECONDS_PER_MINUTE)
const SECONDS_PER_DAY = (24 * SECONDS_PER_HOUR)
const SECONDS_PER_WEEK = (7 * SECONDS_PER_DAY)
const SECONDS_PER_YEAR = (365 * SECONDS_PER_DAY)
const HOURS_PER_SHIFT = 6
const SECONDS_PER_SHIFT = (HOURS_PER_SHIFT * SECONDS_PER_HOUR)
const SHIFTS_PER_DAY = 4
const SHIFTS_PER_WEEK = (4*7)

// ****************************************************************************
// Semantic spacetime timeslots
// ****************************************************************************

func DoNowt(then time.Time) (string,string) {

	//then := given.UnixNano()

	// Time on the torus (donut/doughnut) (CFEngine style)
	// The argument is a Golang time unit e.g. then := time.Now()
	// Return a db-suitable keyname reflecting the coarse-grained SST time
	// The function also returns a printable summary of the time

	year := fmt.Sprintf("Yr%d",then.Year())
	month := GR_MONTH_TEXT[int(then.Month())-1]
	day := then.Day()
	hour := fmt.Sprintf("Hr%02d",then.Hour())
	mins := fmt.Sprintf("Min%02d",then.Minute())
	quarter := fmt.Sprintf("Q%d",then.Minute()/15 + 1)
	shift :=  fmt.Sprintf("%s",GR_SHIFT_TEXT[then.Hour()/6])

	//secs := then.Second()
	//nano := then.Nanosecond()

	dayname := then.Weekday()
	dow := fmt.Sprintf("%.3s",dayname)
	daynum := fmt.Sprintf("Day%d",day)

	// 5 minute resolution capture
        interval_start := (then.Minute() / 5) * 5
        interval_end := (interval_start + 5) % 60
        minD := fmt.Sprintf("Min%02d_%02d",interval_start,interval_end)

	var when string = fmt.Sprintf("%s,%s,%s,%s,%s at %s %s %s %s",shift,dayname,daynum,month,year,hour,mins,quarter,minD)
	var key string = fmt.Sprintf("%s:%s:%s",dow,hour,minD)

	return when, key
}

// ****************************************************************************

func GetUnixTimeKey(now int64) string {

	// Time on the torus (donut/doughnut) (CFEngine style)
	// The argument is in traditional UNIX "time_t" unit e.g. then := time.Unix()
	// This is a simple wrapper to DoNowt() returning only a db-suitable keyname

	t := time.Unix(now, 0)
	_,slot := DoNowt(t)

	return slot
}

//*****************************************************************
// Read text file
//*****************************************************************

func ReadFile(filename string) string {

	// Read a string and strip out characters that can't be used in kenames
	// to yield a "pure" text for n-gram classification, with fewer special chars
	// The text marks end of sentence with a # for later splitting

	content,err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Couldn't find or open",filename)
		os.Exit(-1)
	}

	// Start by stripping HTML / XML tags before para-split
	// if they haven't been removed already

	m1 := regexp.MustCompile("<[^>]*>") 
	cleaned := m1.ReplaceAllString(string(content),";") 
	return cleaned
}

//**************************************************************
// Text Fractionation (alphabetic language)
//**************************************************************

const N_GRAM_MAX = 6
const N_GRAM_MIN = 1

const DUNBAR_5 =5
const DUNBAR_15 = 15
const DUNBAR_30 = 45
const DUNBAR_150 = 150

// **************************************************************

var EXCLUSIONS []string

var STM_NGRAM_FREQ [N_GRAM_MAX]map[string]float64
var STM_NGRAM_LOCA [N_GRAM_MAX]map[string][]int
var STM_NGRAM_LAST [N_GRAM_MAX]map[string]int

type TextRank struct {
	Significance float64
	Fragment     string
	Order        int
	Partition    int
}

//**************************************************************

func NewNgramMap() [N_GRAM_MAX]map[string]float64 {

	var thismap [N_GRAM_MAX]map[string]float64

	for i := 1; i < N_GRAM_MAX; i++ {
		thismap[i] = make(map[string]float64)
	}

	return thismap
}

//**************************************************************

func CleanText(s string) string {

	// Start by stripping HTML / XML tags before para-split
	// if they haven't been removed already

	m := regexp.MustCompile("<[^>]*>") 
	s = m.ReplaceAllString(s,":\n") 

	// Weird English abbrev
	s = strings.Replace(s,"Mr.","Mr",-1) 
	s = strings.Replace(s,"Ms.","Ms",-1) 
	s = strings.Replace(s,"Mrs.","Mrs",-1) 
	s = strings.Replace(s,"Dr.","Dr",-1)
	s = strings.Replace(s,"St.","St",-1) 
	s = strings.Replace(s,"[","",-1) 
	s = strings.Replace(s,"]","",-1) 

	// Encode sentence space boudnaries and end of sentence markers with a # for later splitting

	m = regexp.MustCompile("([?!.。]+[ \n])")  // end of sentence punctuation
	s = m.ReplaceAllString(s,"$0#")

	// ellipsis
	m = regexp.MustCompile("([.][.][.])+")  // end of sentence punctuation
	s = m.ReplaceAllString(s,"---")

	m = regexp.MustCompile("[—]+")  // endash
	s = m.ReplaceAllString(s,", ")

	m = regexp.MustCompile("[\n][\n]")     // paragraph or highlighted sentence
	s = m.ReplaceAllString(s,">>\n")

	m = regexp.MustCompile("[\n]+")        // spurious spaces
	s = m.ReplaceAllString(s," ")

	return s
}

//******************************************************************

func FractionateTextFile(name string) ([][][]string,int) {

	file := ReadFile(name)
	proto_text := CleanText(file)
	pbsf := SplitIntoParaSentences(proto_text)

	count := 0

	for p := range pbsf {
		for s := range pbsf[p] {

			count++

			for f := range pbsf[p][s] {

				change_set := Fractionate(pbsf[p][s][f],count,STM_NGRAM_FREQ,N_GRAM_MIN)

				// Update global n-gram frequencies for fragment, and location histories

				for n := N_GRAM_MIN; n < N_GRAM_MAX; n++ {
					for ng := range change_set[n] {
						ngram := change_set[n][ng]
						STM_NGRAM_FREQ[n][ngram]++
						STM_NGRAM_LOCA[n][ngram] = append(STM_NGRAM_LOCA[n][ngram],count)
					}
				}
			}
		}
	}
	return pbsf,count
}

//**************************************************************

func SplitIntoParaSentences(text string) [][][]string {

	// Take arbitrary text (preprocessed by clean text) and return coherent
	// semantic fragments as separate elements

	// If the text contains parenthetic remarks, these appear unexpanded and expanded

	var pbsf [][][]string

	paras := strings.Split(text,">>")

	for p := 0; p < len(paras); p++ {

		re := regexp.MustCompile("[^#]#")

		sentences := re.Split(paras[p], -1)
		
		var cleaned [][]string
		
		for s := range sentences{

			// NB, if parentheses contain multiple sentences, this complains, TBD

			frags := SplitPunctuationText(sentences[s])

			var codons []string

			for f := range frags {
				content := strings.TrimSpace(frags[f])
				if len(content) > 2 {			
					codons = append(codons,content)
				}
			}
			cleaned = append(cleaned,codons)
		}
		pbsf = append(pbsf,cleaned)
	}
	return pbsf
}

//**************************************************************

func SplitCommandText(s string) []string {

	return SplitPunctuationTextWork(s,true)
}

//**************************************************************

func SplitPunctuationText(s string) []string {

	return SplitPunctuationTextWork(s,false)
}

//**************************************************************

func SplitPunctuationTextWork(s string,allow_small bool) []string {

	// first split sentence on intentional separators

	var subfrags []string

	frags := CountParens(s)

	for f := 0; f < len(frags); f++ {

		contents,hasparen := UnParen(frags[f])

		var sfrags []string

		if hasparen {
			// contiguous parenthesis
			subfrags = append(subfrags,frags[f])
			// and fractionated contents (recurse)
			sfrags = SplitPunctuationTextWork(contents,allow_small)
			sfrags = nil // count but don't repeat
		} else {
			re := regexp.MustCompile("([\"—“”!?,:;—]+[ \n])")
			sfrags = re.Split(contents, -1)
		}

		for sf := range sfrags {
			sfrags[sf] = strings.TrimSpace(sfrags[sf])
			
			if allow_small || len(sfrags[sf]) > 1 {
				subfrags = append(subfrags,sfrags[sf])
			}
		}
	}

	// handle parentheses first as a single fragment because this could mean un-accenting

	// now split on any punctuation that's not a hyphen
	
	return subfrags
}

//**************************************************************

func UnParen(s string) (string,bool) {

	var counter byte = ' '

	switch s[0] {
	case '(':
		counter = ')'
	case '[':
		counter = ']'
	case '{':
		counter = '}'
	}

	if counter != ' ' {
		if s[len(s)-1] == counter {
			trimmed := strings.TrimSpace(s[1:len(s)-1])
			return trimmed,true
		}
	}
	return strings.TrimSpace(s),false
}

//**************************************************************

func CountParens(s string) []string {

	var text = []rune(strings.TrimSpace(s))

	var match rune = ' '
	var count = make(map[rune]int)

	var subfrags []string
	var fragstart int = 0

	for i := 0; i < len(text); i++ {

		switch text[i] {
		case '(':
			count[')']++
			if match == ' ' {
				match = ')'
				frag := strings.TrimSpace(string(text[fragstart:i]))
				fragstart = i
				if len(frag) > 0 {
					subfrags = append(subfrags,frag)
				}
			}
		case '[':
			count[']']++
			if match == ' ' {
				match = ']'
				frag := strings.TrimSpace(string(text[fragstart:i]))
				fragstart = i
				if len(frag) > 0 {
					subfrags = append(subfrags,frag)
				}
			}
		case '{':
			count['}']++
			if match == ' ' {
				match = '}'
				frag := strings.TrimSpace(string(text[fragstart:i]))
				fragstart = i
				if len(frag) > 0 {
					subfrags = append(subfrags,frag)
				}
			}

			// end

		case ')',']','}':
			count[text[i]]--
			if count[match] == 0 {
				frag := text[fragstart:i+1]
				fragstart = i+1
				subfrags = append(subfrags,string(frag))
			}
		}

	}

	lastfrag := strings.TrimSpace(string(text[fragstart:len(text)]))

	if len(lastfrag) > 0 {
		subfrags = append(subfrags,string(lastfrag))
	}

	// Ignore unbalanced parentheses, because it's unclear why in natural language

	return subfrags
}

//**************************************************************

func Fractionate(frag string,L int,frequency [N_GRAM_MAX]map[string]float64,min int) [N_GRAM_MAX][]string {

	// A round robin cyclic buffer for taking fragments and extracting
	// n-ngrams of 1,2,3,4,5,6 words separateed by whitespace, passing

	var rrbuffer [N_GRAM_MAX][]string
	var change_set [N_GRAM_MAX][]string

	words := strings.Split(frag," ")

	for w := range words {
		rrbuffer,change_set = NextWord(words[w],rrbuffer)
	}

	return change_set
}

//**************************************************************

func AssessStaticIntent(frag string,L int,frequency [N_GRAM_MAX]map[string]float64,min int) float64 {

	// A round robin cyclic buffer for taking fragments and extracting
	// n-ngrams of 1,2,3,4,5,6 words separateed by whitespace, passing

	var change_set [N_GRAM_MAX][]string
	var rrbuffer [N_GRAM_MAX][]string
	var score float64

	words := strings.Split(frag," ")

	for w := range words {

		rrbuffer,change_set = NextWord(words[w],rrbuffer)

		for n := min; n < N_GRAM_MAX; n++ {
			for ng := range change_set[n] {
				ngram := change_set[n][ng]
				score += StaticIntentionality(L,ngram,STM_NGRAM_FREQ[n][ngram])
			}
		}
	}

	return score
}

//**************************************************************

func AssessStaticTextAnomalies(L int,frequencies [N_GRAM_MAX]map[string]float64,locations [N_GRAM_MAX]map[string][]int) ([N_GRAM_MAX][]TextRank,[N_GRAM_MAX][]TextRank) {

	// Try to split a text into anomalous/ambient i.e. intentional + contextual  parts

	const coherence_length = DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var anomalous [N_GRAM_MAX][]TextRank
	var ambient [N_GRAM_MAX][]TextRank

	for n := N_GRAM_MIN; n < N_GRAM_MAX; n++ {

		for ngram := range STM_NGRAM_LOCA[n] {

			var ns TextRank
			ns.Significance = AssessStaticIntent(ngram,L,STM_NGRAM_FREQ,1)
			ns.Fragment = ngram

			if IntentionalNgram(n,ngram,L,coherence_length) {
				anomalous[n] = append(anomalous[n],ns)
			} else {
				ambient[n] = append(ambient[n],ns)
			}
		}
		
		sort.Slice(anomalous[n], func(i, j int) bool {
			return anomalous[n][i].Significance > anomalous[n][j].Significance
		})

		sort.Slice(ambient[n], func(i, j int) bool {
			return ambient[n][i].Significance > ambient[n][j].Significance
		})
	}

	var intent [N_GRAM_MAX][]TextRank
	var context [N_GRAM_MAX][]TextRank
	var max_intentional = [N_GRAM_MAX]int{0,0,DUNBAR_150,DUNBAR_150,DUNBAR_30,DUNBAR_15}

	for n := N_GRAM_MIN; n < N_GRAM_MAX; n++ {

		for i := 0; i < max_intentional[n] && i < len(anomalous[n]); i++ {
			intent[n] = append(intent[n],anomalous[n][i])
		}

		for i := 0; i < max_intentional[n] && i < len(ambient[n]); i++ {
			context[n] = append(context[n],ambient[n][i])
		}
	}

	return intent,context
}

//**************************************************************

func IntentionalNgram(n int,ngram string,L int,coherence_length int) bool {

	// If short file, everything is probably significant

	if n == 1 {
		return false 
	}

	if L < coherence_length {
		return true
	}

	occurrences,minr,maxr := IntervalRadius(n,ngram)

	// if too few occurrences, no difference between max and min delta

	if occurrences < 2 {
		return true
	}

	// the distribution of intraspacings is broad, so not just a regular pattern

	return maxr > minr + coherence_length
}

//**************************************************************

func IntervalRadius(n int, ngram string) (int,int,int) {

	// find minimax distances between n-grams (in sentences)

	occurrences := len(STM_NGRAM_LOCA[n][ngram])
	var dl int = 0
	var dlmin int = 99
	var dlmax int = 0

	// Find the width of the intraspacing distribution

	for occ := 0; occ < occurrences; occ++ {

		d := STM_NGRAM_LOCA[n][ngram][occ]
		delta := d - dl
		dl = d
		
		if dl == 0 {
			continue
		}
		
		if dl > dlmax {
			dlmax = delta
		}
		
		if dl < dlmin {
			dlmin = delta
		}
	}

	return occurrences,dlmin,dlmax
}

//**************************************************************

func AssessTextCoherentCoactivation(L int,ngram_loc [N_GRAM_MAX]map[string][]int) ([N_GRAM_MAX]map[string]int,[N_GRAM_MAX]map[string]int,int) {

	// In this global assessment of coherence intervals, we separate each into text that is unique (intentional)
	// and fragments that are repeated in any other interval, so this is an extreme view. Compare to fast/slow method
	// below

	const coherence_length = DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var overlap [N_GRAM_MAX]map[string]int
	var condensate [N_GRAM_MAX]map[string]int

	C,partitions := CoherenceSet(ngram_loc,L,coherence_length)

	for n := 1; n < N_GRAM_MAX; n++ {

		overlap[n] = make(map[string]int)
		condensate[n] = make(map[string]int)

		// now run through linearly and split nearest neighbours

		// very short excerpts,there is nothing we can do in a single coherence set
		if partitions < 2 {
			for ngram := range C[n][0] {
				overlap[n][ngram]++
			}
		// multiple coherence zones
		} else {
			for pi := 0; pi < len(C[n]); pi++ {
				for pj := pi+1; pj < len(C[n]); pj++ {
					for ngram := range C[n][pi] {
						if C[n][pi][ngram] > 0 && C[n][pj][ngram] > 0 {
							// ambients
							delete(condensate[n],ngram)
							overlap[n][ngram]++
						} else {
							// unique things here
							_,ambient := overlap[n][ngram]
							if !ambient {
								condensate[n][ngram]++
							}
						}
					}
				}
			}
		}
	}
	return overlap,condensate,partitions
}

//**************************************************************

func AssessTextFastSlow(L int,ngram_loc [N_GRAM_MAX]map[string][]int) ([N_GRAM_MAX][]map[string]int,[N_GRAM_MAX][]map[string]int,int) {

	// Use a running evaluation of context intervals to separate ngrams that are varying quickly (intentional)
	// from those changing slowly (context). For each region, what if different from the last in fast and what
	// remains the same as last is slow. This is remarkably effective and quick to calculate.

	const coherence_length = DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var slow [N_GRAM_MAX][]map[string]int
	var fast [N_GRAM_MAX][]map[string]int

	C,partitions := CoherenceSet(ngram_loc,L,coherence_length)

	for n := 1; n < N_GRAM_MAX; n++ {

		slow[n] = make([]map[string]int,partitions)
		fast[n] = make([]map[string]int,partitions)

		// now run through linearly and split nearest neighbours

		// very short excerpts,there is nothing we can do in a single coherence set
		if partitions < 2 {
			slow[n][0] = make(map[string]int)
			fast[n][0] = make(map[string]int)

			for ngram := range C[n][0] {
				fast[n][0][ngram]++
			}
		// multiple coherence zones
		} else {
			for p := 1; p < partitions; p++ {

				slow[n][p-1] = make(map[string]int)
				fast[n][p-1] = make(map[string]int)

				for ngram := range C[n][p-1] {
					if C[n][p][ngram] > 0 && C[n][p-1][ngram] > 0 {
						// ambients
						slow[n][p-1][ngram]++
					} else {
						// unique things here
						fast[n][p-1][ngram]++
					}
				}
			}
		}
	}

	return slow,fast,partitions
}

//**************************************************************

func CoherenceSet(ngram_loc [N_GRAM_MAX]map[string][]int, L,coherence_length int) ([N_GRAM_MAX][]map[string]int,int) {

	var C [N_GRAM_MAX][]map[string]int

	partitions := L/coherence_length + 1

	for n := 1; n < N_GRAM_MAX; n++ {
		
		C[n] = make([]map[string]int,partitions)
		for p := 0; p < partitions; p++ {
			C[n][p] = make(map[string]int)
		}

		for ngram := range ngram_loc[n] {
			
			// commute indices and expand to a sparse representation for simplicity

			for s := range ngram_loc[n][ngram] {
				p := ngram_loc[n][ngram][s] / coherence_length
				C[n][p][ngram]++
			}
		}
	}

	return C,partitions
}

//**************************************************************

func NextWord(frag string,rrbuffer [N_GRAM_MAX][]string) ([N_GRAM_MAX][]string,[N_GRAM_MAX][]string) {

	// Word by word, we form a superposition of scores from n-grams of different lengths
	// as a simple sum. This means lower lengths will dominate as there are more of them
	// so we define intentionality proportional to the length also as compensation

	var change_set [N_GRAM_MAX][]string

	for n := 1; n < N_GRAM_MAX; n++ {
		
		// Pop from round-robin

		if (len(rrbuffer[n]) > n-1) {
			rrbuffer[n] = rrbuffer[n][1:n]
		}
		
		// Push new to maintain length

		rrbuffer[n] = append(rrbuffer[n],frag)

		// Assemble the key, only if complete cluster
		
		if (len(rrbuffer[n]) > n-1) {
			
			var key string
			
			for j := 0; j < n; j++ {
				key = key + rrbuffer[n][j]
				if j < n-1 {
					key = key + " "
				}
			}

			key = CleanNgram(key)

			if ExcludedByBindings(CleanNgram(rrbuffer[n][0]),CleanNgram(rrbuffer[n][n-1])) {
				continue
			}

			change_set[n] = append(change_set[n],key)
		}
	}

	frag = CleanNgram(frag)
	
	if N_GRAM_MIN <= 1 && !ExcludedByBindings(frag,frag) {
		change_set[1] = append(change_set[1],frag)
	}

	return rrbuffer,change_set
}

//**************************************************************

func CleanNgram(s string) string {

	re := regexp.MustCompile("[-][-][-].*")
	s = re.ReplaceAllString(s,"")
	re = regexp.MustCompile("[\"—“”!?`,.:;—()_]+")
	s = re.ReplaceAllString(s,"")
	s = strings.Replace(s,"  "," ",-1)
	s = strings.Trim(s,"-")
	s = strings.Trim(s,"'")

	return strings.ToLower(s)
}

//**************************************************************

func ExtractIntentionalTokens(L int) ([][]string,[][]string,[]string,[]string) {

	slow,fast,doc_parts := AssessTextFastSlow(L,STM_NGRAM_LOCA)

	var grad_amb [N_GRAM_MAX]map[string]float64
	var grad_oth [N_GRAM_MAX]map[string]float64

	// returns

	var fastparts = make([][]string,doc_parts)
	var slowparts = make([][]string,doc_parts)
	var fastwhole []string
	var slowwhole []string

	for n := 1; n < N_GRAM_MAX; n++ {
		grad_amb[n] = make(map[string]float64)
		grad_oth[n] = make(map[string]float64)
	}

	for p := 0; p < doc_parts; p++ {

		for n := 1; n < N_GRAM_MAX; n++ {

			var amb []string
			var other []string

			for ngram := range fast[n][p] {
				other = append(other,ngram)
			}

			for ngram := range slow[n][p] {
				amb = append(amb,ngram)
			}
			
			// Sort by intentionality

			sort.Slice(amb, func(i, j int) bool {
				ambi :=	StaticIntentionality(L,amb[i],STM_NGRAM_FREQ[n][amb[i]])
				ambj := StaticIntentionality(L,amb[j],STM_NGRAM_FREQ[n][amb[j]])
				return ambi > ambj
			})

			sort.Slice(other, func(i, j int) bool {
				inti := StaticIntentionality(L,other[i],STM_NGRAM_FREQ[n][other[i]])
				intj := StaticIntentionality(L,other[j],STM_NGRAM_FREQ[n][other[j]])
				return inti > intj
			})
			
			for i := 0 ; i < 150 && i < len(amb); i++ {
				v := StaticIntentionality(L,amb[i],STM_NGRAM_FREQ[n][amb[i]])
				slowparts[p] = append(slowparts[p],amb[i])
				grad_amb[n][amb[i]] += v
			}
			
			for i := 0 ; i < 150 && i < len(other); i++ {
				v := StaticIntentionality(L,other[i],STM_NGRAM_FREQ[n][other[i]])
				fastparts[p] = append(fastparts[p],other[i])
				grad_oth[n][other[i]] += v
			}
		}
	}
	
	// Summary ranking of whole doc
	
	for n := 1; n < N_GRAM_MAX; n++ {
		
		var amb []string
		var other []string
				
		// there is possible overlap

		for ngram := range grad_oth[n] {
			_,dup := grad_amb[n][ngram]
			if dup {
				continue
			}
			other = append(other,ngram)
		}

		for ngram := range grad_amb[n] {
			amb = append(amb,ngram)
		}

		// Sort by intentionality
		
		sort.Slice(amb, func(i, j int) bool {
			ambi := StaticIntentionality(L,amb[i],STM_NGRAM_FREQ[n][amb[i]])
			ambj := StaticIntentionality(L,amb[j],STM_NGRAM_FREQ[n][amb[j]])
			return ambi > ambj
		})
		sort.Slice(other, func(i, j int) bool {
			inti := StaticIntentionality(L,other[i],STM_NGRAM_FREQ[n][other[i]])
			intj := StaticIntentionality(L,other[j],STM_NGRAM_FREQ[n][other[j]])
			return inti > intj
		})
		
		for i := 0 ; i < 150 && i < len(amb); i++ {
			slowwhole = append(slowwhole,amb[i])
		}

		for i := 0 ; i < 150 && i < len(other); i++ {
			fastwhole = append(fastwhole,other[i])
		}
		fmt.Println()
	}	

	return fastparts,slowparts,fastwhole,slowwhole
}

//**************************************************************
// Heuristics for Text Processing
//**************************************************************

func ExcludedByBindings(firstword,lastword string) bool {

	// A standalone fragment can't start/end with these words, because they
	// Promise to bind to something else...
	// Rather than looking for semantics, look at spacetime promises only - words that bind strongly
	// to a prior or posterior word.

	// Promise bindings in English. This domain knowledge saves us a lot of training analysis
	
	var forbidden_ending = []string{"but", "and", "the", "or", "a", "an", "its", "it's", "their", "your", "my", "of", "as", "are", "is", "was", "has", "be", "with", "using", "that", "who", "to" ,"no", "because","at","but","yes","no","yeah","yay", "in", "which", "what","as","he","she","they","all","I","they","from","for","then"}
	
	var forbidden_starter = []string{"and","or","of","the","it","because","in","that","these","those","is","are","was","were","but","yes","no","yeah","yay","also","me","them","him","but"}

	if (len(firstword) <= 2) || len(lastword) <= 2 {
		return true
	}

	for s := range forbidden_ending {
		if strings.ToLower(lastword) == forbidden_ending[s] {
			return true
		}
	}
	
	for s := range forbidden_starter {
		if strings.ToLower(firstword) == forbidden_starter[s] {
			return true
		}
	}

	return false 
}

//**************************************************************

func RunningIntentionality(t int, frag string) float64 {

	// A round robin cyclic buffer for taking fragments and extracting
	// n-ngrams of 1,2,3,4,5,6 words separateed by whitespace, passing

	var change_set [N_GRAM_MAX][]string
	var rrbuffer [N_GRAM_MAX][]string
	var score float64

	words := strings.Split(frag," ")
	decayrate := float64(DUNBAR_30)

	for w := range words {

		rrbuffer,change_set = NextWord(words[w],rrbuffer)

		for n := 1; n < N_GRAM_MAX; n++ {
			for ng := range change_set[n] {
				ngram := change_set[n][ng]
				work := float64(len(ngram))
				lastseen := STM_NGRAM_LAST[n][ngram]

				if lastseen == 0 {
					score = work
				} else {
					score += work * (1 - math.Exp(-float64(t-lastseen)/decayrate))
				}

				STM_NGRAM_LAST[n][ngram] = t
			}
		}
	}

	return score

}

//**************************************************************

func StaticIntentionality(L int, s string, freq float64) float64 {

	// Compute the effective significance of a string s
	// within a document of many sentences. The weighting due to
	// inband learning uses an exponential deprecation based on
	// SST scales (see "leg" meaning).

	work := float64(len(s)) 

	// tempting to measure occurrences relative to total length L in sentences
	// but this is not the relevant scale. Coherence is on a shorter scale
	// set by cognitive limits, not author expansiveness / article scope ...

	phi := freq
	phi_0 := float64(DUNBAR_30) // not float64(L)

	// How often is too often for a concept?
	const rho = 1/30.0 

	crit := phi/phi_0 - rho

	meaning := phi * work / (1.0 + math.Exp(crit))

	return meaning
}

// **************************************************************************
//
// Toolkits: generic helper functions
//
// **************************************************************************

func SplitChapters(str string) []string {

	run := []rune(str)

	var part []rune
	var retval []string

	for r := 0; r < len(run); r++ {
		if run[r] == ',' && (r+1 < len(run) && run[r+1] != ' ') {
			retval = append(retval,string(part))
			part = nil
		} else {
			part = append(part,run[r])
		}
	}

	retval = append(retval,string(part))
	return retval
}

// **************************************************************************

func List2Map(l []string) map[string]int {

	var retvar = make(map[string]int)

	for s := range l {
		retvar[strings.TrimSpace(l[s])]++
	}

	return retvar
}

// **************************************************************************

func Map2List(m map[string]int) []string {

	var retvar []string

	for s := range m {
		retvar = append(retvar,strings.TrimSpace(s))
	}

	sort.Strings(retvar)
	return retvar
}

// **************************************************************************

func List2String(list []string) string {

	var s string

	sort.Strings(list)

	for i := 0; i < len(list); i++ {
		s += list[i]
		if i < len(list)-1 {
			s+= ", "
		}
	}

	return s
}

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

		if len(array[i]) == 0 {
			continue
		}

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

	if s == "" {
		return false,"","",""
	}

	if s[0] == '<' && s[len(s)-1] == '>' {
		matrix := s[1:len(s)-1]
		params := strings.Split(matrix,"|")
		
		switch len(params) {
			
		case 2: 
			end = params[0]
			begin = params[1]
		case 3:
			end = params[0]
			context = params[1]
			begin = params[2]			
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

	run := []rune(s)
	var res []rune

	for r := range run {
		if run[r] == '"' {
			res = append(res,'\\')
			res = append(res,'"')
		} else {
			res = append(res,run[r])
		}
	}

	s = string(res)
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
			fmt.Print(string(runes[r]))
			r++
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

func Waiting(output bool) {

	if !output {
		return
	}

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
// Unicod
//****************************************************************************

const (
	NON_ASCII_LQUOTE = '“'
	NON_ASCII_RQUOTE = '”'
)

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

//****************************************************************************

func IsQuote(r rune) bool {

	switch r {
	case '"','\'',NON_ASCII_LQUOTE,NON_ASCII_RQUOTE:
		return true
	}

	return false
}

//****************************************************************************

func ReadToNext(array []rune,pos int,r rune) (string,int) {

	var buff []rune

	for i := pos; i < len(array); i++ {

		buff = append(buff,array[i])

		if i > pos && array[i] == r {
			ret := string(buff)
			return ret,len(ret)
		}
	}

	ret := string(buff)
	return ret,len(ret)
}



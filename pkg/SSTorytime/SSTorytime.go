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
	"sync"
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
	ERR_MEMORY_DB_CONTEXT_MISMATCH = "Contexts in database are not in synch (shouldn't happen)"
	WARN_DIFFERENT_CAPITALS = "WARNING: Another capitalization exists"

	SCREENWIDTH = 120
	RIGHTMARGIN = 5
	LEFTMARGIN = 5

	NEAR = 0
	LEADSTO = 1   // +/-
	CONTAINS = 2  // +/-
	EXPRESS = 3   // +/-

	// Letting a cone search get too large is unresponsive
	CAUSAL_CONE_MAXLIMIT = 100

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

	// mandatory relations used in text processing, but we should never follow these links
        // when presenting results

	EXPR_INTENT_L = "has contextual theme"
	EXPR_INTENT_S = "has_theme"
        INV_EXPR_INTENT_L = "is a context theme in"
	INV_EXPR_INTENT_S ="theme_of"

	EXPR_AMBIENT_L = "has contextual highlight"
	EXPR_AMBIENT_S = "has_highlight"
        INV_EXPR_AMBIENT_L = "is contextual highlight of"
	INV_EXPR_AMBIENT_S = "highlight_of"

	CONT_FINDS_L = "contains extract/quote"
	CONT_FINDS_S = "has-extract"
	INV_CONT_FOUND_IN_L = "extract/quote from"
	INV_CONT_FOUND_IN_S = "extract-fr"

	// This is a "contained-by something that expresses" shortcut => NEAR

	CONT_FRAG_L = "contains intented characteristic"      // intentional characteristic
	CONT_FRAG_S = "has-frag"
	INV_CONT_FRAG_IN_L = "characteristic of"   // explains intentional context
	INV_CONT_FRAG_IN_S = "charct-in"

)

var BASE_DB_CHANNEL_STATE[7] ClassedNodePtr

var CLASS_CHANNEL_DESCRIPTION = []string{"","single word ngram","two word ngram","three word ngram",
	"string less than 128 chars","string less than 1024 chars","string greater than 1024 chars"}

//**************************************************************
// Data structures
//**************************************************************

type PoSST struct {

   DB *sql.DB
}

//******************************************************************

type Node struct {

	L         int     // length of text string
	S         string  // text string itself

	Seq       bool    // true if this node begins an intended sequence, otherwise ambiguous
	Chap      string  // section/chapter name in which this was added
	NPtr      NodePtr // Pointer to self index

	I [ST_TOP][]Link  // link incidence list, by STindex - these are the "vectors" +/-
  	                  // NOTE: carefully how STindex offsets represent negative SSTtypes
}

//**************************************************************

type Link struct {  // A link is a type of arrow, with context
                    // and maybe with a weightfor package math
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

var NODE_CACHE = make(map[NodePtr]NodePtr)

//**************************************************************
// POSTGRES DATA TYPES
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
	"Ctx      int,        " +
	"Dst      NodePtr     " +
	")"

const NODE_TABLE = "CREATE UNLOGGED TABLE IF NOT EXISTS Node " +
	"( " +
	"NPtr      NodePtr,        \n" +
	"L         int,            \n" +
	"S         text,           \n" +
	"Search    TSVECTOR GENERATED ALWAYS AS (to_tsvector('english',S)) STORED,\n" +
	"UnSearch  TSVECTOR GENERATED ALWAYS AS (to_tsvector('english',sst_unaccent(S))) STORED,\n" +
	"Chap      text,           \n" +
	"Seq       boolean,        \n" +
	I_MEXPR+"  Link[],         \n" + // Im3
	I_MCONT+"  Link[],         \n" + // Im2
	I_MLEAD+"  Link[],         \n" + // Im1
	I_NEAR +"  Link[],         \n" + // In0
	I_PLEAD+"  Link[],         \n" + // Il1
	I_PCONT+"  Link[],         \n" + // Ic2
	I_PEXPR+"  Link[]          \n" + // Ie3
	")"

const PAGEMAP_TABLE = "CREATE UNLOGGED TABLE IF NOT EXISTS PageMap " +
	"( " +
	"Chap     Text,  " +
	"Alias    Text,  " +
	"Ctx      int,   " +
	"Line     Int,   " +
	"Path     Link[] " +
	")"

const ARROW_DIRECTORY_TABLE = "CREATE UNLOGGED TABLE IF NOT EXISTS ArrowDirectory " +
	"(    " +
	"STAindex int,           " +
	"Long text,              " +
	"Short text,             " +
	"ArrPtr int primary key  " +
	")"

const ARROW_INVERSES_TABLE = "CREATE UNLOGGED TABLE IF NOT EXISTS ArrowInverses " +
	"(    " +
	"Plus int,  " +
	"Minus int,  " +
	"Primary Key(Plus,Minus)" +
	")"

const LASTSEEN_TABLE = "CREATE TABLE IF NOT EXISTS LastSeen " +
	"(    " +
	"Section text," +
	"NPtr    NodePtr," +
	"Last    timestamp," +
	"Delta   real," +
	"Freq    int" +
	")"

const CONTEXT_DIRECTORY_TABLE = "CREATE TABLE IF NOT EXISTS ContextDirectory " +
	"(    " +
	"Context text,            " +
	"CtxPtr  int primary key  " +
	")"

const APPOINTMENT_TYPE = "CREATE TYPE Appointment AS  " +
	"(                    " +
	"Arr    int," +
	"STType int," +
	"Chap   text," +
	"Ctx    int," +
	"NTo    NodePtr," +
	"NFrom  NodePtr[]" +
	")"

//**************************************************************
// Lookup tables
//**************************************************************

var ( 
	// Arrow multi-name factorization

	ARROW_DIRECTORY []ArrowDirectory
	ARROW_SHORT_DIR = make(map[string]ArrowPtr) // Look up short name int referene
	ARROW_LONG_DIR = make(map[string]ArrowPtr)  // Look up long name int referene
	ARROW_DIRECTORY_TOP ArrowPtr = 0
	INVERSE_ARROWS = make(map[ArrowPtr]ArrowPtr)

	IGNORE_ARROWS []ArrowPtr

	// Context array factorization

	CONTEXT_DIRECTORY []ContextDirectory
	CONTEXT_DIR = make(map[string]ContextPtr)   // Look up long name int referene
	CONTEXT_TOP ContextPtr

	PAGE_MAP []PageMap

	NODE_DIRECTORY NodeDirectory  // Internal histo-representations
	NO_NODE_PTR NodePtr // see Init()
	NONODE NodePtr

	// Uploading

	WIPE_DB bool = false
        SILLINESS_COUNTER int
        SILLINESS_POS int
        SILLINESS_SLOGAN int
	SILLINESS bool
)

//**************************************************************
// WEB INTEFACE REPRESENTATIONS FOR JSON RENDERING
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
	Last    string   // timestamp of last access
        Pdelta  float64  // previous average of access intervals
	Ndelta  float64  // current last access interval
	Freq    int      // count of total accesses
	NPtr    NodePtr
	XYZ     Coords
}

//******************************************************************
// Open Library Context
//******************************************************************

func Open(load_arrows bool) PoSST {

	var sst PoSST
	var err error

	// Replace this with a private file

	var (
		user     = "sstoryline"
		password = "sst_1234"
		dbname   = "sstoryline"
	)

	user,password,dbname = OverrideCredentials(user,password,dbname)

        connStr := "user="+user+" dbname="+dbname+" password="+password+" sslmode=disable"

        sst.DB, err = sql.Open("postgres", connStr)

	if err != nil {
	   	fmt.Println("Error connecting to the database: ", err)
		os.Exit(-1)
	}
	
	err = sst.DB.Ping()
	
	if err != nil {
		fmt.Println("Error pinging the database: ", err)
		os.Exit(-1)
	}

	MemoryInit()
	Configure(sst,load_arrows)

	NO_NODE_PTR.Class = 0
	NO_NODE_PTR.CPtr =  -1
	NONODE.Class = 0
	NONODE.CPtr = 0

	return sst
}

// **************************************************************************

func OverrideCredentials(u,p,d string) (string,string,string) {

	// Store database/postgres credentials in a system file instead of hardcoding

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

	// For parsing the password credential file

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
//  When opening a connection, restore config
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

	for i := 1; i < N_GRAM_MAX; i++ {

		STM_NGRAM_FREQ[i] = make(map[string]float64)
		STM_NGRAM_LOCA[i] = make(map[string][]int)
		STM_NGRAM_LAST[i] = make(map[string]int)
	}
}

// **************************************************************************

func Configure(sst PoSST,load_arrows bool) {

	// Tmp reset

	if WIPE_DB {

		fmt.Println("***********************")
		fmt.Println("* WIPING DB")
		fmt.Println("***********************")
		
		sst.DB.QueryRow("DROP INDEX sst_nan")
		sst.DB.QueryRow("DROP INDEX sst_type")
		sst.DB.QueryRow("DROP INDEX sst_gin")
		sst.DB.QueryRow("DROP INDEX sst_ungin")
		sst.DB.QueryRow("DROP INDEX sst_s")
		sst.DB.QueryRow("DROP INDEX sst_n")
		sst.DB.QueryRow("DROP INDEX sst_cnt")

		sst.DB.QueryRow("drop function fwdconeaslinks")
		sst.DB.QueryRow("drop function fwdconeasnodes")
		sst.DB.QueryRow("drop function fwdpathsaslinks")
		sst.DB.QueryRow("drop function getfwdlinks")
		sst.DB.QueryRow("drop function getfwdnodes")
		sst.DB.QueryRow("drop function getneighboursbytype")
		sst.DB.QueryRow("drop function getsingletonaslink")
		sst.DB.QueryRow("drop function AllNCPathsAsLinks")
		sst.DB.QueryRow("drop function AllSuperNCPathsAsLinks")
		sst.DB.QueryRow("drop function SumAllNCPaths")
		sst.DB.QueryRow("drop function GetNCFwdLinks")
		sst.DB.QueryRow("drop function GetNCCLinks")

		sst.DB.QueryRow("drop function getsingletonaslinkarray")
		sst.DB.QueryRow("drop function idempinsertnode")
		sst.DB.QueryRow("drop function sumfwdpaths")
		sst.DB.QueryRow("drop function match_context")
		sst.DB.QueryRow("drop function empty_path")
		sst.DB.QueryRow("drop function match_arrows")
		sst.DB.QueryRow("drop function ArrowInList")
		sst.DB.QueryRow("drop function GetNCCStoryStartNodes")
		sst.DB.QueryRow("drop function GetStoryStartNodes")
		sst.DB.QueryRow("drop function GetAppointments")
		sst.DB.QueryRow("drop function UnCmp")
		sst.DB.QueryRow("drop function DeleteChapter")

		sst.DB.QueryRow("drop function lastsawsection(text)")
		sst.DB.QueryRow("drop function lastsawnptr(nodeptr)")

		sst.DB.QueryRow("drop type NodePtr")
		sst.DB.QueryRow("drop type Link")
		sst.DB.QueryRow("drop type Appointment")

		sst.DB.QueryRow("drop table Node")
		sst.DB.QueryRow("drop table PageMap")
		sst.DB.QueryRow("drop table NodeArrowNode")
		sst.DB.QueryRow("drop table ArrowDirectory")
		sst.DB.QueryRow("drop table ArrowInverses")
		sst.DB.QueryRow("drop table ContextDirectory")
		sst.DB.QueryRow("drop table LastSeen")

	}

	// Create functions, some we use in autocreating index columns

	sst.DB.QueryRow("CREATE EXTENSION unaccent")

	if !CreateType(sst,NODEPTR_TYPE) {
		fmt.Println("Unable to create type as, ",NODEPTR_TYPE)
		os.Exit(-1)
	}

	if !CreateType(sst,LINK_TYPE) {
		fmt.Println("Unable to create type as, ",LINK_TYPE)
		os.Exit(-1)
	}

	if !CreateType(sst,APPOINTMENT_TYPE) {
		fmt.Println("Unable to create type as, ",APPOINTMENT_TYPE)
		os.Exit(-1)
	}

	if !CreateTable(sst,CONTEXT_DIRECTORY_TABLE) {
		fmt.Println("Unable to create table as, ",CONTEXT_DIRECTORY_TABLE)
		os.Exit(-1)
	}

	DefineStoredFunctions(sst)

	if !CreateTable(sst,PAGEMAP_TABLE) {
		fmt.Println("Unable to create table as, ",PAGEMAP_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(sst,NODE_TABLE) {
		fmt.Println("Unable to create table as, ",NODE_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(sst,ARROW_INVERSES_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_INVERSES_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(sst,ARROW_DIRECTORY_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_DIRECTORY_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(sst,LASTSEEN_TABLE) {
		fmt.Println("Unable to create table as, ",LASTSEEN_TABLE)
		os.Exit(-1)
	}

	DownloadArrowsFromDB(sst)
	DownloadContextsFromDB(sst)
	SynchronizeNPtrs(sst)

	// Find ignorable arrows
}

// **************************************************************************

func Close(sst PoSST) {
	sst.DB.Close()
}

// **************************************************************************
//  Context registratation and directory management
// **************************************************************************

func GetContext(contextptr ContextPtr) string {

	exists := int(contextptr) < len(CONTEXT_DIRECTORY)

	if exists {
		return CONTEXT_DIRECTORY[contextptr].Context
	}

	return "unknown context"
}

// ****************************************************************************

func RegisterContext(parse_state map[string]bool,context []string) ContextPtr {

	ctxstr := NormalizeContextString(parse_state,context)

	if len(ctxstr) == 0 {
		return 0
	}

	ctxptr,exists := CONTEXT_DIR[ctxstr] 

	if !exists {
		var cd ContextDirectory
		cd.Context = ctxstr
		cd.Ptr = CONTEXT_TOP
		CONTEXT_DIRECTORY = append(CONTEXT_DIRECTORY,cd)
		CONTEXT_DIR[ctxstr] = CONTEXT_TOP
		ctxptr = CONTEXT_TOP
		CONTEXT_TOP++
	}

	return ctxptr
}

// **************************************************************************

func TryContext(sst PoSST,context []string) ContextPtr {

	ctxstr := CompileContextString(context)
	str,ctxptr := GetDBContextByName(sst,ctxstr)

	if ctxptr == -1 || str != ctxstr {
		ctxptr = UploadContextToDB(sst,ctxstr,-1)
		RegisterContext(nil,context)
	}

	return ctxptr
}

// **************************************************************************

func CompileContextString(context []string) string {

	// Ensure idempotence

	var merge = make(map[string]int)

	for c := range context {
		merge[context[c]]++
	}

	return List2String(Map2List(merge))	
}

// **************************************************************************

func NormalizeContextString(contextmap map[string]bool,ctx []string) string {

	// Mitigate combinatoric explosion

	var merge = make(map[string]bool)
	var clist []string

	// Merge sources into single map
	
	if contextmap != nil {
		for c := range contextmap {
			merge[c] = true
		}
	}

	for c := range ctx {
		merge[ctx[c]] = true
	}

	for c := range merge {
		s := strings.Split(c,",")
		for i := range s {
			s[i] = strings.TrimSpace(s[i])
			if s[i] != "_sequence_" {
				clist = append(clist,s[i])
			}
		}
	}

	return List2String(clist)
}

// **************************************************************************

func GetNodeContext(sst PoSST,node Node) []string {

	str := GetNodeContextString(sst,node)

	if str != "" {
		return strings.Split(str,",")
	}

	return nil
}

// **************************************************************************

func GetNodeContextString(sst PoSST,node Node) string {

	// This reads the ghost link planted for the purpose of attaching
	// a context to floating nodes

	empty := GetDBArrowByName(sst,"empty")

	for _,lnk := range node.I[ST_ZERO+LEADSTO] {

		if lnk.Arr == empty {
			return GetContext(lnk.Ctx)
		}
	}

	return ""
}

// **************************************************************************
//  Node registration and memory management
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
		IdempAddChapterSeqToNode(node_alloc_ptr.Class,node_alloc_ptr.CPtr,event.Chap,event.Seq)
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

func IdempAddChapterSeqToNode(class int,cptr ClassedNodePtr,chap string,seq bool) {

	/* In the DB version, we have handle chapter collisions
           we want all similar names to have a single node for lateral
           association, but we need to be able to search by chapter too,
           so merge the chapters as an attribute list */

	var node Node

	node = UpdateSeqStatus(class,cptr,seq)

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

func UpdateSeqStatus(class int,cptr ClassedNodePtr,seq bool) Node {

	switch class {
	case N1GRAM:
		NODE_DIRECTORY.N1directory[cptr].Seq = NODE_DIRECTORY.N1directory[cptr].Seq || seq
		return NODE_DIRECTORY.N1directory[cptr]
	case N2GRAM:
		NODE_DIRECTORY.N2directory[cptr].Seq = NODE_DIRECTORY.N2directory[cptr].Seq || seq
		return NODE_DIRECTORY.N2directory[cptr]
	case N3GRAM:
		NODE_DIRECTORY.N3directory[cptr].Seq = NODE_DIRECTORY.N3directory[cptr].Seq || seq
		return NODE_DIRECTORY.N3directory[cptr]
	case LT128:
		NODE_DIRECTORY.LT128[cptr].Seq = NODE_DIRECTORY.LT128[cptr].Seq || seq
		return NODE_DIRECTORY.LT128[cptr]
	case LT1024:
		NODE_DIRECTORY.LT1024[cptr].Seq = NODE_DIRECTORY.LT1024[cptr].Seq || seq
		return NODE_DIRECTORY.LT1024[cptr]
	case GT1024:
		NODE_DIRECTORY.GT1024[cptr].Seq = NODE_DIRECTORY.GT1024[cptr].Seq || seq
		return NODE_DIRECTORY.GT1024[cptr]
	}

	fmt.Println("Non existent node class (shouldn't happen)")
	os.Exit(-1)
	var dummy Node
	return dummy
}

//**************************************************************
// Link registration and management
//**************************************************************

func AppendLinkToNode(frptr NodePtr,link Link,toptr NodePtr) {

	frclass := frptr.Class
	frm := frptr.CPtr
	stindex := ARROW_DIRECTORY[link.Arr].STAindex

	link.Dst = toptr // fill in the last part of the reference

	// Idempotently add any new context strings to the current list
	// between from and to nodes -- stindex tells us which link type, so implicit in the arrow type
	// the empty arrow is used to record node context, which is type LEADSTO

	switch frclass {

	case N1GRAM:
		NODE_DIRECTORY.N1directory[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.N1directory[frm].I[stindex],link)
	case N2GRAM:
		NODE_DIRECTORY.N2directory[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.N2directory[frm].I[stindex],link)
	case N3GRAM:
		NODE_DIRECTORY.N3directory[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.N3directory[frm].I[stindex],link)
	case LT128:
		NODE_DIRECTORY.LT128[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.LT128[frm].I[stindex],link)
	case LT1024:
		NODE_DIRECTORY.LT1024[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.LT1024[frm].I[stindex],link)
	case GT1024:
		NODE_DIRECTORY.GT1024[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.GT1024[frm].I[stindex],link)
	}
}

//**************************************************************

func MergeLinkLists(linklist []Link,lnk Link) []Link {

	// Ensure all arrows and contexts in lnk are in list for the appropriate arrows

	new_ctxstr := GetContext(lnk.Ctx)
	new_ctxlist := strings.Split(new_ctxstr,",")

	// Check if the arrow is already there to add to its context

	for l := range linklist {
		if linklist[l].Arr == lnk.Arr && linklist[l].Dst == lnk.Dst {

			already_ctxstr := GetContext(linklist[l].Ctx)
			already_ctxlist := strings.Split(already_ctxstr,",")

			linklist[l].Ctx = MergeContextLists(already_ctxlist,new_ctxlist)

			return linklist
		}
	}

	// if not already there, add this arrow

	linklist = append(linklist,lnk)
	return linklist
}

//**************************************************************

func MergeContextLists(one,two []string) ContextPtr {

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

	ctxstr := List2String(merged)

	// Register the merger of contexts

	ctxptr,ok := CONTEXT_DIR[ctxstr]

	if ok {
		return ctxptr
	} else {
		var cd ContextDirectory
		cd.Context = ctxstr
		cd.Ptr = CONTEXT_TOP
		CONTEXT_DIRECTORY = append(CONTEXT_DIRECTORY,cd)
		CONTEXT_DIR[ctxstr] = CONTEXT_TOP
		ctxptr = CONTEXT_TOP
		CONTEXT_TOP++
	}

	return ctxptr
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
// STtype naming and conversion
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
// Arrow registration and management
//**************************************************************

func InsertArrowDirectory(stname,alias,name,pm string) ArrowPtr {

	// Insert an arrow into the forward/backward indices

	var newarrow ArrowDirectory

	// Check is already exists - harmless

	prev_alias,a_exists := ARROW_SHORT_DIR[alias]
	prev_name,n_exists := ARROW_LONG_DIR[name]

	if a_exists && n_exists {
		if prev_alias == prev_name {
			return prev_alias
		}
	}

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
// Upload managed N4L graph / Write to database
//**************************************************************

func GraphToDB(sst PoSST,wait_counter bool) {

	total := len(NODE_DIRECTORY.N1directory) + len(NODE_DIRECTORY.N2directory) + len(NODE_DIRECTORY.N3directory) + len(NODE_DIRECTORY.LT128) + len(NODE_DIRECTORY.LT1024) + len(NODE_DIRECTORY.GT1024) + len(PAGE_MAP)

	fmt.Println("\nStoring primary nodes ...\n")

	for class := N1GRAM; class <= GT1024; class++ {

		offset := int(BASE_DB_CHANNEL_STATE[class])

		switch class {
		case N1GRAM:
			for n := offset; n < len(NODE_DIRECTORY.N1directory); n++ {
				org := NODE_DIRECTORY.N1directory[n]
				UploadNodeToDB(sst,org)
				Waiting(wait_counter,total)
			}
		case N2GRAM:
			for n := offset; n < len(NODE_DIRECTORY.N2directory); n++ {
				org := NODE_DIRECTORY.N2directory[n]
				UploadNodeToDB(sst,org)
				Waiting(wait_counter,total)
			}
		case N3GRAM:
			for n := offset; n < len(NODE_DIRECTORY.N3directory); n++ {
				org := NODE_DIRECTORY.N3directory[n]
				UploadNodeToDB(sst,org)
				Waiting(wait_counter,total)
			}
		case LT128:
			for n := offset; n < len(NODE_DIRECTORY.LT128); n++ {
				org := NODE_DIRECTORY.LT128[n]
				UploadNodeToDB(sst,org)
				Waiting(wait_counter,total)
			}
		case LT1024:
			for n := offset; n < len(NODE_DIRECTORY.LT1024); n++ {
				org := NODE_DIRECTORY.LT1024[n]
				UploadNodeToDB(sst,org)
				Waiting(wait_counter,total)
			}

		case GT1024:
			for n := offset; n < len(NODE_DIRECTORY.GT1024); n++ {
				org := NODE_DIRECTORY.GT1024[n]
				UploadNodeToDB(sst,org)
				Waiting(wait_counter,total)
			}
		}
	}

	// Arrows etc

	fmt.Println("\nStoring Arrows...")

	sst.DB.QueryRow("drop table ArrowDirectory")
	sst.DB.QueryRow("drop table ArrowInverses")

	if !CreateTable(sst,ARROW_INVERSES_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_INVERSES_TABLE)
		os.Exit(-1)
	}
	if !CreateTable(sst,ARROW_DIRECTORY_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_DIRECTORY_TABLE)
		os.Exit(-1)
	}

	for arrow := range ARROW_DIRECTORY {

		UploadArrowToDB(sst,ArrowPtr(arrow))
	}

	fmt.Println("Storing inverse Arrows...")

	for arrow := range INVERSE_ARROWS {

		UploadInverseArrowToDB(sst,ArrowPtr(arrow))
	}

	fmt.Println("Storing contexts...")

	UploadContextsToDB(sst)

	fmt.Println("Storing page map...")

	for line := 0; line < len(PAGE_MAP); line ++ {
		UploadPageMapEvent(sst,PAGE_MAP[line])
		Waiting(wait_counter,total)
	}

	// CREATE INDICES

	fmt.Println("Indexing ....")

//	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_type on Node (((NPtr).Chan),L,S)")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_gin on Node USING GIN (to_tsvector('english',Search))")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_ungin on Node USING GIN (to_tsvector('english',UnSearch))")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_s on Node USING GIN (S)")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_n on Node USING GIN (NPtr)")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_cnt on ContextDirectory USING GIN (Context)")
	sst.DB.QueryRow("ALTER TABLE Node SET LOGGED")
	sst.DB.QueryRow("ALTER TABLE PageMap SET LOGGED")

	fmt.Println("Finally done!")
}

// **************************************************************************
// Store - High level API, for automatic NPtr numbering
// **************************************************************************

func Vertex(sst PoSST,name,chap string) Node {

	var n Node

	n.S = name
	n.Chap = chap

	return IdempDBAddNode(sst,n)
}

// **************************************************************************

func Edge(sst PoSST,from Node,arrow string,to Node,context []string,weight float32) (ArrowPtr,int) {

	arrowptr,sttype := GetDBArrowsWithArrowName(sst,arrow)

	var link Link

	link.Arr = arrowptr
	link.Dst = to.NPtr
	link.Wgt = weight
	link.Ctx = TryContext(sst,context)

	IdempDBAddLink(sst,from,link,to)

	return arrowptr,sttype
}

// **************************************************************************

func HubJoin(sst PoSST,name,chap string,nptrs []NodePtr,arrow string,context []string,weight []float32) Node {

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
			node := GetDBNodeByNodePtr(sst,nptrs[n])
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

	container := IdempDBAddNode(sst,to)

	arrowptr,_ := GetDBArrowsWithArrowName(sst,arrow)

	for nptr := range nptrs {

		var link Link
		link.Arr = arrowptr
		link.Dst = container.NPtr
		link.Wgt = weight[nptr]
		link.Ctx = TryContext(sst,context)
		from := GetDBNodeByNodePtr(sst,nptrs[nptr])
		IdempDBAddLink(sst,from,link,container)
	}

	return GetDBNodeByNodePtr(sst,container.NPtr)
}

// **************************************************************************
// Lower level functions, for self-managed NPtr values
// **************************************************************************

func FormDBNode(sst PoSST, n Node) string {

	// Add node version setting explicit CPtr value, note different function call
	// We use this function when we ARE managing/counting CPtr values ourselves

	var qstr,seqstr string

        n.L,n.NPtr.Class = StorageClass(n.S)
	
	cptr := n.NPtr.CPtr

	es := SQLEscape(n.S)
	ec := SQLEscape(n.Chap)

	if n.Seq {
		seqstr = "true"
	} else {
		seqstr = "false"
	}

	qstr = fmt.Sprintf("SELECT InsertNode(%d,%d,%d,'%s','%s',%s);\n",n.L,n.NPtr.Class,cptr,es,ec,seqstr)
	return qstr
}


// **************************************************************************

func CreateDBNode(sst PoSST, n Node) Node {

	// Add node version setting explicit CPtr value, note different function call
	// We use this function when we ARE managing/counting CPtr values ourselves

	var qstr string

        n.L,n.NPtr.Class = StorageClass(n.S)
	
	cptr := n.NPtr.CPtr

	es := SQLEscape(n.S)
	ec := SQLEscape(n.Chap)

	qstr = fmt.Sprintf("SELECT IdempInsertNode(%d,%d,%d,'%s','%s')",n.L,n.NPtr.Class,cptr,es,ec)

	row,err := sst.DB.Query(qstr)
	
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

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			fmt.Sscanf(whole,"(%d,%d)",&cl,&ch)
		}
		
		n.NPtr.Class = cl
		n.NPtr.CPtr = ClassedNodePtr(ch)
		
		row.Close()
	}

	return n
}

// **************************************************************************
//  Low level append/insert for auto-managed NPtr values
// **************************************************************************

func IdempDBAddNode(sst PoSST,n Node) Node {

	// We use this function when we aren't counting CPtr values
	// This functon may be deprecated in future

	var qstr string

	// No need to trust the values, ignore/overwrite CPtr

        n.L,n.NPtr.Class = StorageClass(n.S)

	es := SQLEscape(n.S)
	ec := SQLEscape(n.Chap)

	// Wrap BEGIN/END a single transaction

	qstr = fmt.Sprintf("SELECT IdempAppendNode(%d,%d,'%s','%s')",n.L,n.NPtr.Class,es,ec)

	row,err := sst.DB.Query(qstr)
	
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

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			fmt.Sscanf(whole,"(%d,%d)",&cl,&ch)
		}
		
		n.NPtr.Class = cl
		n.NPtr.CPtr = ClassedNodePtr(ch)
		
		row.Close()
	}

	return n
}

// **************************************************************************
//  Uploading memory cache to database
// **************************************************************************

func UploadNodeToDB(sst PoSST, org Node) {

	const nolink = 999

	qstr := "BEGIN;\n" + FormDBNode(sst,org)

	for stindex := 0; stindex < len(org.I); stindex++ {

		lnkarray := FormatSQLLinkArray(org.I[stindex])
		sttype := STIndexToSTType(stindex)
		qstr += AppendDBLinkArrayToNode(sst,org.NPtr,lnkarray,sttype)
	}

	qstr += "\nCOMMIT;"

	row,err := sst.DB.Query(qstr)

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

func UploadArrowToDB(sst PoSST,arrow ArrowPtr) {

	staidx := ARROW_DIRECTORY[arrow].STAindex
	long := SQLEscape(ARROW_DIRECTORY[arrow].Long)
	short := SQLEscape(ARROW_DIRECTORY[arrow].Short)

	qstr := fmt.Sprintf("INSERT INTO ArrowDirectory (STAindex,Long,Short,ArrPtr) SELECT %d,'%s','%s',%d WHERE NOT EXISTS (SELECT Long,Short,ArrPtr FROM ArrowDirectory WHERE lower(Long) = lower('%s') OR lower(Short) = lower('%s') OR ArrPtr = %d)",staidx,long,short,arrow,long,short,arrow)

	row,err := sst.DB.Query(qstr)
	
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

func UploadInverseArrowToDB(sst PoSST,arrow ArrowPtr) {

	plus := arrow
	minus := INVERSE_ARROWS[arrow]

	qstr := fmt.Sprintf("INSERT INTO ArrowInverses (Plus,Minus) SELECT %d,%d WHERE NOT EXISTS (SELECT Plus,Minus FROM ArrowInverses WHERE Plus = %d OR minus = %d)",plus,minus,plus,minus)

	row,err := sst.DB.Query(qstr)
	
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

func UploadContextsToDB(sst PoSST) {

	for ctxdir := range CONTEXT_DIRECTORY {
		UploadContextToDB(sst,CONTEXT_DIRECTORY[ctxdir].Context,CONTEXT_DIRECTORY[ctxdir].Ptr)
	}
}

// **************************************************************************

func UploadContextToDB(sst PoSST,contextstring string,ptr ContextPtr) ContextPtr {

	a := SQLEscape(contextstring)
	b := ptr

	// Make sure neither a nor b are previously defined

	qstr := fmt.Sprintf("SELECT IdempInsertContext('%s',%d)",a,b)

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	var cptr ContextPtr

	if row != nil {
		for row.Next() {
			err = row.Scan(&cptr)
		}
		row.Close()
	}

	return cptr
}

//**************************************************************

func UploadPageMapEvent(sst PoSST, line PageMap) {

	chap := SQLEscape(line.Chapter)

	qstr := "BEGIN;\n"

	qstr += fmt.Sprintf("INSERT INTO PageMap (Chap,Alias,Ctx,Line) VALUES ('%s','%s',%d,%d);\n",chap,line.Alias,line.Context,line.Line)

	lnkarray := FormatSQLLinkArray(line.Path)

	qstr += fmt.Sprintf("\nUPDATE PageMap SET Path='%s' WHERE Chap = '%s' AND Line = '%d';",lnkarray,chap,line.Line)

	qstr += "COMMIT;"

	row,err := sst.DB.Query(qstr)
	
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
}

//**************************************************************

func IdempDBAddLink(sst PoSST,from Node,link Link,to Node) {

	// API Entry point for registering links

	frptr := from.NPtr
	toptr := to.NPtr

	link.Dst = toptr // it might have changed, so override

	if frptr == toptr {
		fmt.Println("Self-loops are not allowed",from.S,from,link,to)
		os.Exit(-1)
	}

	if link.Arr < 0 || len(ARROW_DIRECTORY) == 0 {
		fmt.Println("No arrows have yet been defined, so you can't rely on the arrow names")
		os.Exit(-1)
	}

	if link.Wgt == 0 {
		fmt.Println("Attempt to register a link with zero weight is pointless")
		os.Exit(-1)
	}

	sttype := STIndexToSTType(ARROW_DIRECTORY[link.Arr].STAindex)

	AppendDBLinkToNode(sst,frptr,link,sttype)

	// Double up the reverse definition for easy indexing of both in/out arrows
	// But be careful not the make the graph undirected by mistake

	var invlink Link
	invlink.Arr = INVERSE_ARROWS[link.Arr]
	invlink.Wgt = link.Wgt
	invlink.Dst = frptr
	AppendDBLinkToNode(sst,toptr,invlink,-sttype)
}

// **************************************************************************

func AppendDBLinkToNode(sst PoSST, n1ptr NodePtr, lnk Link, sttype int) bool {

	qstr := AppendDBLinkToNodeCommand(sst,n1ptr,lnk,sttype)

	row,err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("Failed to append",err,qstr)
	       return false
	}

	row.Close()
	return true
}

// **************************************************************************

func AppendDBLinkToNodeCommand(sst PoSST, n1ptr NodePtr, lnk Link, sttype int) string {

	// Want to make this idempotent, because SQL is not (and not clause)

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS,sttype)
		os.Exit(-1)
	}

	if n1ptr == lnk.Dst {
		return ""
	}

	//                       Arr,Wgt,Ctx,  Dst
	linkval := fmt.Sprintf("(%d, %f, %d, (%d,%d)::NodePtr)",lnk.Arr,lnk.Wgt,lnk.Ctx,lnk.Dst.Class,lnk.Dst.CPtr)

	literal := fmt.Sprintf("%s::Link",linkval)

	link_table := STTypeDBChannel(sttype)

	qstr := fmt.Sprintf("UPDATE NODE SET %s=array_append(%s,%s) WHERE (NPtr).CPtr = '%d' AND (NPtr).Chan = '%d' AND (%s IS NULL OR NOT %s = ANY(%s));\n",
		link_table,
		link_table,
		literal,
		n1ptr.CPtr,
		n1ptr.Class,
		link_table,
		literal,
		link_table)

	return qstr
}

// **************************************************************************

func AppendDBLinkArrayToNode(sst PoSST, nptr NodePtr, array string, sttype int) string {

	// Want to make this idempotent, because SQL is not (and not clause)

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS,sttype)
		os.Exit(-1)
	}

	link_table := STTypeDBChannel(sttype)

	qstr := fmt.Sprintf("UPDATE NODE SET %s='%s' WHERE (NPtr).CPtr = '%d' AND (NPtr).Chan = '%d';\n",
		link_table,
		array,
		nptr.CPtr,
		nptr.Class)

	return qstr
}

// **************************************************************************
// Postgres interface
// **************************************************************************

func CreateType(sst PoSST, defn string) bool {

	row,err := sst.DB.Query(defn)

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

func CreateTable(sst PoSST,defn string) bool {

	row,err := sst.DB.Query(defn)
	
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

func DefineStoredFunctions(sst PoSST) {

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

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Force for managed input

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION InsertNode(iLi INT, iszchani INT, icptri INT, iSi TEXT, ichapi TEXT,sequence boolean)\n" +
		"RETURNS bool AS $fn$ " +
		"DECLARE \n" +
		"BEGIN\n" +
		"   INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap,Seq,%s) VALUES (iszchani,icptri,iLi,iSi,ichapi,sequence,'{}','{}','{}','{}','{}','{}','{}');" +
		"   RETURN true;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;",cols);

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Without controlling nptr

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
		"     IF icptri IS NULL THEN"+
		"         icptri = 0;"+
		"     END IF;"+
		"     INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap) VALUES (iszchani,icptri+1,iLi,iSi,ichapi);" +
		"  END IF;\n" +
		"  RETURN QUERY SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;";

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Insert Context from API

	qstr = "CREATE OR REPLACE FUNCTION IdempInsertContext(constr text,conptr int)\n" +
		"RETURNS int AS $fn$ " +
		"DECLARE \n" +
		"    cptr INT = 0;\n" +
		"    found int=-99;\n" +
		"BEGIN\n" +
		"IF conptr=-1 THEN\n"+
		"   SELECT Context,CtxPtr INTO found FROM ContextDirectory WHERE Context=constr AND CtxPtr=conptr;\n"+
		"   SELECT max(CtxPtr) INTO cptr FROM ContextDirectory;\n"+
		"   INSERT INTO ContextDirectory (Context,CtxPtr) VALUES (constr,cptr+1);\n"+
		"   RETURN cptr+1;\n" +
		"END IF;\n" +
		"IF NOT EXISTS (SELECT CtxPtr FROM ContextDirectory WHERE CtxPtr=conptr OR Context=constr) THEN\n" +
		"   INSERT INTO ContextDirectory (Context,CtxPtr) VALUES (constr,conptr);\n"+
		"   RETURN conptr;\n" +
		"END IF;"+
		"SELECT CtxPtr INTO cptr FROM ContextDirectory WHERE CtxPtr=conptr OR Context=constr;\n"+
		"RETURN cptr;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;";

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// For lookup include name,chapter,context,arrow
	
	qstr = "CREATE OR REPLACE FUNCTION NCC_match(thisnptr NodePtr,context text[],arrows int[],sttypes int[],lm3 Link[],lm2 Link[],lm1 Link[],ln0 Link[],lp1 Link[],lp2 Link[],lp3 Link[])\n"+
		"RETURNS boolean AS $fn$\n"+
		"DECLARE \n"+
		"    emptyarray Link[] := Array[] :: Link[];\n"+
		"    lnkarray Link[] := Array[] :: Link[];\n"+
		"    lnk Link;\n"+
		"    st int;\n"+
		"BEGIN\n"+
		
		// If there are no arrows, we only need to look for the Node context in lp1 for "empty" == 0

		"IF array_length(arrows,1) IS NULL THEN\n"+

		"   IF lp1 IS NOT NULL THEN"+		
		"      FOREACH lnk IN ARRAY lp1 LOOP\n"+
                "         IF lnk.Arr = 0 AND match_context(lnk.Ctx,context) THEN\n"+
		"            RETURN true;"+
		"         END IF;"+
		"      END LOOP;\n"+
		"   END IF;\n"+

		"ELSE\n"+

		// If there are arrows

		"   FOREACH st IN ARRAY sttypes LOOP\n"+
		"      CASE st \n"		
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("   WHEN %d THEN\n"+
			"         SELECT %s INTO lnkarray FROM Node WHERE Nptr=thisnptr;\n",st,STTypeDBChannel(st));
	}
	qstr +=	"      ELSE RAISE EXCEPTION 'No such sttype in NCC_match %', sttype;\n" +
		"      END CASE;\n" +
		
		"      FOREACH lnk IN ARRAY lnkarray LOOP\n"+
		"         IF match_arrows(lnk.arr,arrows) AND match_context(lnk.ctx,context) THEN\n"+
		"            RETURN true;\n"+

		"         END IF;\n"+
		"      END LOOP;\n"+
		"   END LOOP;\n"+
		"END IF;\n"+

		"RETURN false; \n"+
		"END ;\n"+
		"$fn$ LANGUAGE plpgsql;"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct an empty link pointing nowhere as a starting node

	qstr = "CREATE OR REPLACE FUNCTION GetSingletonAsLinkArray(start NodePtr)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    level Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,0,(0,0));\n"+
		"BEGIN\n"+
		" lnk.Dst = start;\n"+
		" level = array_append(level,lnk);\n"+
		"RETURN level; \n"+
		"END ;\n"+
		"$fn$ LANGUAGE plpgsql;"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct an empty link pointing nowhere as a starting node

	qstr = "CREATE OR REPLACE FUNCTION GetSingletonAsLink(start NodePtr)\n"+
		"RETURNS Link AS $fn$\n"+
		"DECLARE \n"+
		"    lnk Link := (0,1.0,0,(0,0));\n"+
		"BEGIN\n"+
		" lnk.Dst = start;\n"+
		"RETURN lnk; \n"+
		"END ;\n"+
		"$fn$ LANGUAGE plpgsql;"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct search by sttype. since table names are static we need a case statement

	qstr = "CREATE OR REPLACE FUNCTION GetNeighboursByType(start NodePtr, sttype int, maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    fwdlinks Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,0,(0,0));\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"     SELECT %s INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;\n",st,STTypeDBChannel(st));
	}
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN fwdlinks; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Get the nearest neighbours as NPtr, with respect to each of the four STtype

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetFwdNodes(start NodePtr,exclude NodePtr[],sttype int,maxlimit int)\n"+
		"RETURNS NodePtr[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours NodePtr[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNeighboursByType(start,sttype,maxlimit);\n"+

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

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Basic quick neighbour probe

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetFwdLinks(start NodePtr,exclude NodePtr[],sttype int,maxlimit int)\n")
	qstr +=	"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNeighboursByType(start,sttype,maxlimit);\n"+

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
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()
	
	// Get the forward cone / half-ball as NPtr

	qstr = "CREATE OR REPLACE FUNCTION FwdConeAsNodes(start NodePtr,sttype INT, maxdepth INT,maxlimit int)\n"+
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
		"      partlevel := GetFwdNodes(neigh,exclude,sttype,maxlimit);\n" +
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
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
	qstr = "CREATE OR REPLACE FUNCTION FwdConeAsLinks(start NodePtr,sttype INT,maxdepth INT,maxlimit int)\n"+
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
		"      partlevel := GetFwdLinks(neigh.Dst,exclude,sttype,maxlimit);\n" +
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

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Orthogonal (depth first) paths from origin spreading out

	qstr = "CREATE OR REPLACE FUNCTION FwdPathsAsLinks(start NodePtr,sttype INT,maxdepth INT, maxlimit INT)\n"+
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
		"ret_paths := SumFwdPaths(startlnk,path,sttype,1,maxdepth,exclude, maxlimit);" +

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Return end of path branches as aggregated text summaries

	qstr = "CREATE OR REPLACE FUNCTION SumFwdPaths(start Link,path TEXT, sttype INT,depth int, maxdepth INT,exclude NodePtr[], maxlimit INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE \n" + 
		"    fwdlinks Link[];\n" +
		"    empty Link[] = ARRAY[]::Link[];\n" +
		"    lnk Link;\n" +
		"    fwd Link;\n" +
		"    ret_paths Text;\n" +
		"    appendix Text;\n" +
		"    tot_path Text;\n"+
		"    count    int = 0;\n"+
		"    horizon  int = 0;\n"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		"fwdlinks := GetFwdLinks(start.Dst,exclude,sttype, maxlimit);\n" +
		
		// limit recursion explosions
		"horizon := maxlimit - array_length(fwdlinks,1);"+

		"IF horizon < 0 THEN\n"+
		"  horizon = 0;\n"+
		"  maxdepth = depth + 1;"+
		"END IF;\n"+

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL OR count >= maxlimit THEN" +
		          // set end of path as return val
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"         RETURN ret_paths;"+
		"      ELSE\n"+
		"         count = count + 1;"+
		          // Add to the path and descend into new link
		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumFwdPaths(lnk,tot_path,sttype,depth+1,maxdepth,exclude,horizon);\n" +
		          // when we return, we reached the end of one path
		"         IF appendix IS NOT NULL THEN\n"+
	                     // append full path to list of all paths, separated by newlines
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"            count = count + regexp_count(appendix,';');"+
		"         ELSE"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;"+
		"      END IF;"+
		"   END IF;"+
		"END LOOP;"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Typeless cone searches

	qstr = "CREATE OR REPLACE FUNCTION AllPathsAsLinks(start NodePtr,orientation text,maxdepth INT, maxlimit INT)\n"+
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
		"ret_paths := SumAllPaths(startlnk,path,orientation,1,maxdepth,exclude, maxlimit);" +
		
		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
        // select AllPathsAsLinks('(4,1)',3)

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// SumAllPaths

	qstr = "CREATE OR REPLACE FUNCTION SumAllPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,exclude NodePtr[],maxlimit int)\n"+
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
		"    counter int = 0;"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		// Get *All* in/out Links
		"CASE \n" +
		"   WHEN orientation = 'bwd' THEN\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   WHEN orientation = 'fwd' THEN\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   ELSE\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"END CASE;\n" +

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF counter > maxlimit THEN\n"+
		"      RETURN ret_paths;"+
		"   END IF;"+
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL THEN\n" +
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"         RETURN ret_paths;"+
		"      ELSE\n"+
		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumAllPaths(lnk,tot_path,orientation,depth+1,maxdepth,exclude,maxlimit);\n" +
		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"         ELSE\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;\n"+
		"         counter = counter + 1;\n"+
		"      END IF;\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
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

	row,err = sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Matching context strings with fuzzy criteria. The policy/notes expression is db_set
	// the client/lookup set is user_set - both COULD use AND expressions.
	// We are looking for sets that overlap for a true result

	qstr = "CREATE OR REPLACE FUNCTION match_context(thisctxptr int,user_set text[])\n"+
		"RETURNS boolean AS $fn$\n" +
		"DECLARE\n" +
		"   ctxstr text;\n" +
		"   db_set text[];\n" +
		"   notes text[];\n" +
		"   client text[];\n" +
		"   pattern text;\n" +
		"   and_list text[];\n"+
		"   and_count int;\n"+
		"   and_result int = 0;\n"+
		"   end_result int = 0;\n"+
		"   or_list text[] = ARRAY[]::text[];\n"+
		"   item text;\n" +
		"   item_db text;\n" +
		"   item_us text;\n" +
		"   ref text;\n" +
		"   c text;\n"+
		"BEGIN \n" +

		// If no constraints at all, then match

		"IF array_length(user_set,1) IS NULL THEN\n" +  // Shouldn't happen anymore
		"   RETURN true;\n"+
		"END IF;\n"+

		"IF user_set[0] = '' THEN\n"+
		"   RETURN true;\n"+
		"END IF;\n"+

		// Convert context ptr into a list from the new factored cache
		"SELECT Context INTO ctxstr FROM ContextDirectory WHERE ctxPtr=thisctxptr;" +
		"db_set = regexp_split_to_array(ctxstr,',');\n" +

		// If there is a constraint, but no db membership, then no match
		"IF array_length(db_set,1) IS NULL AND array_length(user_set,1) IS NOT NULL THEN\n"+
		"   RETURN false;\n"+
		"END IF;\n"+

		// if both are empty, then match
		"IF array_length(db_set,1) IS NULL AND array_length(user_set,1) IS NULL THEN\n"+
		"   RETURN true;\n"+
		"END IF;\n"+

		// clean and unaccent sets

		"FOREACH item_db IN ARRAY db_set LOOP\n" +
		"   FOREACH item_us IN ARRAY user_set LOOP\n" +
		"      IF item_db = item_us THEN\n" +
		"         RETURN true;\n"+
		"      END IF;\n"+
		"   END LOOP;\n" +
		"END LOOP;\n" +

		"FOREACH item IN ARRAY db_set LOOP\n" +
		"   notes = array_append(notes,lower(unaccent(item)));\n" +
		"END LOOP;\n" +

		"FOREACH item IN ARRAY user_set LOOP\n" +
		"   client = array_append(client,lower(unaccent(item)));\n" +
		"END LOOP;\n" +

	       // First split check AND strings in the notes

		"FOREACH item IN ARRAY notes LOOP\n" +

		"   and_list = regexp_split_to_array(item, '\\.');\n" +
		"   and_count = array_length(and_list,1);\n"+

		"   IF and_count > 1 THEN\n"+

		// end_result = MatchANDExpression(and_list,client)

	        "      and_result = 0;\n"+

                       // check each and expression first

		"      FOREACH ref IN ARRAY and_list LOOP\n"+
		"         FOREACH c IN ARRAY client LOOP\n"+
		             // AND need an exact match
		"            IF ref = c THEN \n" +
	        "               and_result = and_result + 1;\n" +
		"            END IF;\n" +
		"         END LOOP;\n"+
		"      END LOOP;\n"+

		"      IF and_result = and_count THEN\n"+
		"         end_result = end_result + 1;\n"+
	        "      END IF;\n" +
		"   ELSE\n"+
		"      or_list = array_append(or_list,item);\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"IF end_result > 0 THEN\n"+
		"   RETURN true;\n" +
		"END IF;\n"+

		// if still not match, check any left overs, client AND matches are still unresolved
		"FOREACH ref IN ARRAY or_list LOOP\n" +
		    // now we can look at substring partial matches
		"   FOREACH c IN ARRAY client LOOP\n"+
		"      pattern := Format('[^.]*%s[^.]*',c);\n" +
		       // substring too greedy if there is a .
		"      IF substring(ref,pattern) IS NOT NULL THEN \n" +
	        "         return true;\n" +
		"      END IF;\n" +
		"   END LOOP;\n"+
		"END LOOP;\n" +

		"RETURN false;\n" +

		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
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

	row,err = sst.DB.Query(qstr)

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

	row,err = sst.DB.Query(qstr)

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
		"  IF lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN\n"+
		"     RETURN true;\n"+
		"  END IF;\n"+
		"END LOOP;"+
		"RETURN false;"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	//

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

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED UnCmp definition\n",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)
	// SumAllNCPaths - a filtering version of the SumAllPaths recursive helper function, slower but more powerful

	qstr = "CREATE OR REPLACE FUNCTION SumAllNCPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,chapter text,rm_acc boolean,context text[],exclude NodePtr[],maxlimit int)\n"+
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
		"    count    int = 0;\n"+
		"    horizon  int = 0;\n"+
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
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   WHEN orientation = 'fwd' THEN\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   ELSE\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"END CASE;\n" +

		"horizon := maxlimit - array_length(fwdlinks,1);"+

		"IF horizon < 0 THEN\n"+
		"  horizon = 0;\n"+
		"  maxdepth = depth + 1;"+
		"END IF;\n"+

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL OR count > maxlimit THEN\n" +
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"      ELSE\n"+
		"         count = count + 1;"+
		"         IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN\n"+
                "            CONTINUE;\n"+
                "         END IF;\n"+

		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumAllNCPaths(lnk,tot_path,orientation,depth+1,maxdepth,chapter,rm_acc,context,exclude,horizon);\n" +

		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"            count = count + regexp_count(appendix,';');"+
		"         ELSE\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;\n"+
		"      END IF;\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)
        // with a start set of more than one node

	qstr = "CREATE OR REPLACE FUNCTION AllNCPathsAsLinks(start NodePtr[],chapter text,rm_acc boolean,context text[],orientation text,maxdepth INT,maxlimit int)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   root Text;\n" +
		"   path Text;\n"+
		"   node NodePtr;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = start;\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;\n"+
		"BEGIN\n" +

		// Aggregate array of starting set
		"FOREACH node IN ARRAY start LOOP\n"+
		"   startlnk := GetSingletonAsLink(node);\n"+
		"   path := Format('%s',startlnk::Text);\n"+
		"   root := SumAllNCPaths(startlnk,path,orientation,1,maxdepth,chapter,rm_acc,context,exclude,maxlimit);\n" +
		"   ret_paths := Format('%s\n%s',ret_paths,root);\n"+
		"END LOOP;"+

		"RETURN ret_paths;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)
        // with a start set of more than one node

	qstr = "CREATE OR REPLACE FUNCTION ConstraintPathsAsLinks(start NodePtr[],chapter text,rm_acc boolean,context text[],arrows int[],sttypes int[],maxdepth INT,maxlimit int)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   root Text;\n" +
		"   path Text;\n"+
		"   node NodePtr;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = start;\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;\n"+
		"BEGIN\n" +

		"IF sttypes IS NULL OR array_length(sttypes,1) IS NULL THEN\n" +
		"   sttypes = ARRAY[-3,-2,-1,0,1,2,3];\n" +
		"END IF;\n"+

		// Aggregate array of starting set

		"FOREACH node IN ARRAY start LOOP\n"+
		"   startlnk := GetSingletonAsLink(node);\n"+
		"   path := Format('%s',startlnk::Text);\n"+
		"   root := SumConstraintPaths(startlnk,path,1,maxdepth,chapter,rm_acc,context,arrows,sttypes,exclude,maxlimit);\n" +
		"   ret_paths := Format('%s\n%s',ret_paths,root);\n"+
		"END LOOP;"+

		"RETURN ret_paths;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // Generalized path search
	// SumConstraintPaths - a filtering version of the SumAllPaths recursive helper function, slower but more powerful

	qstr = "CREATE OR REPLACE FUNCTION SumConstraintPaths(start Link,path TEXT,depth int,maxdepth INT,chapter text,rm_acc boolean,context text[],arrows int[],sttypes int[],exclude NodePtr[],maxlimit int)\n"+
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
		"    count    int = 0;\n"+
		"    horizon  int = 0;\n"+
		"    sttype   int;\n"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		"FOREACH sttype IN ARRAY sttypes LOOP\n"+
		"   stlinks := GetConstrainedFwdLinks(start.Dst,chapter,rm_acc,context,exclude,sttype,arrows,maxlimit);\n" +
		"   fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"END LOOP;\n"+

		"horizon := maxlimit - array_length(fwdlinks,1);"+

		"IF horizon < 0 THEN\n"+
		"  horizon = 0;\n"+
		"  maxdepth = depth + 1;"+
		"END IF;\n"+

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL OR count > maxlimit THEN\n" +
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"      ELSE\n"+
		"         count = count + 1;"+
		"         IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN\n"+
                "            CONTINUE;\n"+
                "         END IF;\n"+

		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumConstraintPaths(lnk,tot_path,depth+1,maxdepth,chapter,rm_acc,context,arrows,sttypes,exclude,horizon);\n" +

		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"            count = count + regexp_count(appendix,';');"+
		"         ELSE\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;\n"+
		"      END IF;\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ****
        // Fully filtering version of the neighbour scan
	// ****

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetConstrainedFwdLinks(start NodePtr,chapter text,rm_acc boolean,context text[],exclude NodePtr[],sttype int,arrows int[],maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"   fwdlinks = GetNCNeighboursByType(start,chapter,rm_acc,sttype,maxlimit);\n"+

		"   IF fwdlinks IS NULL THEN\n" +
		"       RETURN '{}';\n" +
		"   END IF;\n" +

		"    neighbours := ARRAY[]::Link[];\n" +

		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+

		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+

		"      IF arrows IS NOT NULL AND array_length(arrows,1) > 0 AND NOT lnk.Arr=ANY(arrows) THEN\n"+
		"         CONTINUE;\n"+
		"      END IF;\n"+

                "      IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN\n"+
                "         CONTINUE;\n"+
                "      END IF;\n"+

		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 

		"    END LOOP;\n" +

		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ****
        // An NC/C filtering version of the neighbour scan
	// ****

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetNCFwdLinks(start NodePtr,chapter text,rm_acc boolean,context text[],exclude NodePtr[],sttype int,maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNCNeighboursByType(start,chapter,rm_acc,sttype,maxlimit);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +
		"    neighbours := ARRAY[]::Link[];\n" +
		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+

		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+

                "      IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN\n"+
                "         CONTINUE;\n"+
                "      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 
		"    END LOOP;\n" +
		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

        // This one includes an NCC chapter and context filter so slower! 

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetNCCLinks(start NodePtr,exclude NodePtr[],sttype int,chapter text,rm_acc boolean,context text[],maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks =GetNCNeighboursByType(start,chapter,rm_acc,sttype,maxlimit);\n"+

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
	
	qstr = "CREATE OR REPLACE FUNCTION GetNCNeighboursByType(start NodePtr, chapter text,rm_acc boolean,sttype int,maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    fwdlinks Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,0,(0,0));\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"     SELECT %s INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;\n",st,STTypeDBChannel(st));
	}
	
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN fwdlinks; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
	// **************************************
	// Looking for hub / appointed node matroid search
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
		"	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN\n" +
		"  	          count = count + 1;\n" +
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN\n"+
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

		qstr += "   ELSE\n"

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
		"	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN\n" +
		"  	          count = count + 1;\n" +
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN\n"+
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

		"   END IF;\n"
	}
	
	qstr += "END CASE;\n"
	qstr += "    RETURN appointed;\n"
	qstr += "END ;\n"
	qstr += "$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()

	// **************************************
	// Maintenance/deletion transactions
	// **************************************

	qstr = "CREATE OR REPLACE FUNCTION DeleteChapter(chapter text)\n"+
		"RETURNS boolean AS $fn$\n" +
		"DECLARE\n" +
		"   marked    NodePtr[];\n"+
		"   autoset   NodePtr[];\n"+
		"   nnptr     NodePtr;\n"+
		"   lnk       Link;\n"+
		"   links     Link[];\n"+
		"   ed_list   Link[];\n"+
		"   oleft     text;\n"+
		"   oright    text;\n"+
		"   chaparray text[];\n"+
		"   chaplist  text;\n"+
	        "   ed_chap   text;\n"+
		"   chp       text;\n"+
	
		"BEGIN \n"+

		// First get all NPtrs contained in the chapter for deletion
		// To avoid deleting overlaps, select only the automorphic links

		"chp := Format('%%%s%%',chapter);\n"+
		"SELECT array_agg(NPtr) into autoset FROM Node WHERE Chap LIKE chp;\n"+

		"IF autoset IS NULL THEN\n"+
		"   RETURN false;\n"+
		"END IF;\n"+

		// Look for overlapping chapters

		"oleft := Format('%%%s,%%',chapter);\n"+
		"oright := Format('%%,%s%%',chapter);\n"+

		"SELECT array_agg(NPtr) into marked FROM Node WHERE Chap LIKE oleft OR Chap LIKE oright;\n"+

		"IF marked IS NULL THEN\n"+
		"   DELETE FROM Node WHERE Chap = chapter;\n"+
		"   RETURN true;\n"+
		"END IF;\n"+

		"FOREACH nnptr IN ARRAY marked LOOP\n"+
		"   SELECT Chap into chaplist FROM Node WHERE NPtr = nnptr;\n"+
		"   chaparray = string_to_array(chaplist,',');\n"+

		// Remove the chapter reference
		"IF chaparray IS NOT NULL AND array_length(chaparray,1) > 1 THEN"+
		"   FOREACH chp IN ARRAY chaparray LOOP\n"+
		"      IF NOT chp = chapter THEN"+
		"         IF length(ed_chap) > 0 THEN\n"+
		"            ed_chap = Format('%s,%s',ed_chap,chp);\n"+
		"         ELSE"+
		"            ed_chap = chp;"+
		"         END IF;"+
		"      END IF;"+
		"   END LOOP;"+
		"   UPDATE Node SET Chap = ed_chap WHERE NPtr = nnptr;\n"+
		"   marked = array_remove(marked,nnptr);"+
		"END IF;\n"

	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf(
			
			"SELECT %s into links FROM Node WHERE NPtr = nnptr;\n"+

			"   IF links IS NOT NULL THEN\n"+
			"      ed_list = ARRAY[]::Link[];\n"+         // delete reference links
			"      FOREACH lnk in ARRAY links LOOP\n"+
			"         IF NOT lnk.Dst = ANY(marked) THEN\n"+
			"            ed_list = array_append(ed_list,lnk);\n"+
			"         END IF;\n"+
			"      END LOOP;\n"+
			"      UPDATE Node SET %s = ed_list WHERE NPtr = nnptr;\n"+
			"   END IF;\n",
			STTypeDBChannel(st),STTypeDBChannel(st))
	}
	
	qstr += "END LOOP;\n"+

		"DELETE FROM Node WHERE Nptr = ANY(marked);\n"+
		"DELETE FROM Node WHERE Chap = chapter;\n"+

		"RETURN true;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ************ LAST SEEN **************'

	qstr = "CREATE OR REPLACE FUNCTION LastSawSection(this text)\n"+
		"RETURNS bool AS $fn$\n"+
		"DECLARE \n"+
		"  prev      timestamp = NOW();\n"+
		"  prevdelta int;\n"+
		"  deltat    int;\n"+
		"  avdeltat  real;\n"+
		"  nowt      int;\n"+
		"  f         int = 0;"+
		"BEGIN\n"+
		"  SELECT last,EXTRACT(EPOCH FROM NOW()-last),delta,freq INTO prev,deltat,prevdelta,f FROM LastSeen WHERE section=this;\n"+
		"  IF NOT FOUND THEN\n"+
		"     INSERT INTO LastSeen (section,last,delta,freq,nptr) VALUES (this, NOW(),0,1,'(-1,-1)');\n"+
		"  ELSE\n"+
		"     avdeltat = 0.5 * deltat::real + 0.5 * prevdelta::real;\n"+
		"     f = f + 1;\n"+
		      // 1 minute dead time
		"     IF deltat > 60 THEN\n"+
		"       UPDATE LastSeen SET last=NOW(),delta=avdeltat,freq=f WHERE section = this;\n"+
		"     ELSE\n"+
		"        return false;\n"+
		"     END IF;\n"+
		"  END IF;\n"+
		"  RETURN true;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
	qstr = "CREATE OR REPLACE FUNCTION LastSawNPtr(this NodePtr,name text)\n"+
		"RETURNS bool AS $fn$\n"+
		"DECLARE \n"+
		"  prev      timestamp = NOW();\n"+
		"  prevdelta int;\n"+
		"  avdeltat  real;\n"+
		"  deltat    int;\n"+
		"  nowt      int;\n"+
		"  ep        int = 0;"+
		"  f         int = 0;"+
		"BEGIN\n"+
		"  SELECT last,EXTRACT(EPOCH FROM NOW()-last),delta,freq INTO prev,deltat,prevdelta,f FROM LastSeen WHERE nptr=this;\n"+
		"  IF NOT FOUND THEN\n"+
		"     INSERT INTO LastSeen (section,nptr,last,freq,delta) VALUES (name,this,NOW(),1,0);\n"+
		"  ELSE\n"+
		"     avdeltat = 0.5 * deltat::real + 0.5 * prevdelta::real;\n"+
		"     f = f + 1;\n"+
		      // 1 minute dead time
		"     IF deltat > 60 THEN\n"+
		"        UPDATE LastSeen SET last=NOW(),delta=avdeltat,freq=f WHERE nptr = this;\n"+
		"     ELSE\n"+
		"        return false;\n"+
		"     END IF;\n"+
		"  END IF;\n"+
		"  RETURN true;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()

	// Finally an immutable wrapper

	qstr = "CREATE OR REPLACE FUNCTION sst_unaccent(this text)\n"+
		"RETURNS text AS $fn$\n"+
		"DECLARE \n"+
		"  s text;\n"+
		"BEGIN\n"+
		"  s = unaccent(this);\n"+
		"  RETURN s;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql IMMUTABLE;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()

}

// **************************************************************************
// Query / Retrieve structural parts
// **************************************************************************

func GetDBNodePtrMatchingName(sst PoSST,name,chap string) []NodePtr {

	// simplified, retain for compatibility

	return GetDBNodePtrMatchingNCCS(sst,name,chap,nil,nil,false,CAUSAL_CONE_MAXLIMIT)
}

// **************************************************************************

func GetDBNodePtrMatchingNCCS(sst PoSST,nm,chap string,cn []string,arrow []ArrowPtr,seq bool,limit int) []NodePtr {

	// Order by L to favour exact matches

	nm = SQLEscape(nm)
	chap = SQLEscape(chap)

	qstr := fmt.Sprintf("SELECT NPtr FROM Node WHERE %s ORDER BY L ASC,(CARDINALITY(Ie3)+CARDINALITY(Im3)+CARDINALITY(Il1)) DESC LIMIT %d",NodeWhereString(nm,chap,cn,arrow,seq),limit)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY GetNodePtrMatchingNCC Failed",err,qstr)
	}

	var whole string
	var n NodePtr
	var retval []NodePtr

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			fmt.Sscanf(whole,"(%d,%d)",&n.Class,&n.CPtr)
			retval = append(retval,n)
		}

		row.Close()
	}

	return retval
}

// **************************************************************************

func NodeWhereString(name,chap string,context []string,arrow []ArrowPtr,seq bool) string {

	var chap_col, nm_col string
	var ctx_col string
	var qstr string

	// Format a WHERE clause for a Node search satisfying constraints

	// Chapter first to limit search by block

	if chap != "any" && chap != "" {

		remove_chap_accents,chap_stripped := IsBracketedSearchTerm(chap)

		if remove_chap_accents {
			chap_search := "%"+chap_stripped+"%"
			chap_col = fmt.Sprintf("lower(unaccent(Chap)) LIKE lower('%s')",chap_search)
		} else {
			chap_search := "%"+chap+"%"
			chap_col = fmt.Sprintf("lower(Chap) LIKE lower('%s')",chap_search)
		}
	} else {
		chap_col = "true"
	}

	// Name search using tsquery for wildcards and additional S = exact_constraint for !exact!

	outer_exact_match,nopling := IsExactMatch(name)
	remove_name_accents,nobrack := IsBracketedSearchTerm(nopling)
	inner_exact_match,bare_name := IsExactMatch(nobrack)

	is_exact_match := outer_exact_match || inner_exact_match

	// First ignore technical references from ad hoc search results, like img paths

	if !strings.HasPrefix(bare_name,"/") {
		nm_col = "AND S NOT LIKE '/%'"
	}

	if is_exact_match {

		nm_col += fmt.Sprintf(" AND lower(S) = '%s'",bare_name)

	} else if IsStringFragment(bare_name) {

		if name == "any" || name == "%%" {
			nm_col = fmt.Sprintf(" AND lower(S) LIKE '%%%%'")
		} else {
			nm_col = fmt.Sprintf(" AND lower(S) LIKE '%%%s%%'",bare_name)
		}
	} else {

		if name == "any" || name == "%%" {
			nm_col = ""
		} else {
			if remove_name_accents {
				nm_col = fmt.Sprintf(" AND Unsearch @@ to_tsquery('english', '%s')",bare_name)
			} else {
				nm_col = fmt.Sprintf(" AND Search @@ to_tsquery('english', '%s')",bare_name)
			}
		}
	}

        var seq_col string
        
        if seq {
                seq_col = "AND Seq=true"
        }

	// context and arrows

	_,cn_stripped := IsBracketedSearchList(context)
	ctx_col = FormatSQLStringArray(cn_stripped)

	arrows := FormatSQLIntArray(Arrow2Int(arrow))
	sttypes := FormatSQLIntArray(GetSTtypesFromArrows(arrow))

	dbcols := I_MEXPR+","+I_MCONT+","+I_MLEAD+","+I_NEAR +","+I_PLEAD+","+I_PCONT+","+I_PEXPR

	qstr = fmt.Sprintf("%s %s %s AND NCC_match(NPtr,%s,%s,%s,%s)",
		chap_col,nm_col,seq_col, ctx_col,arrows,sttypes,dbcols)

	return qstr
}

// **************************************************************************

func GetDBChaptersMatchingName(sst PoSST,src string) []string {

	var qstr string

	remove_accents,stripped := IsBracketedSearchTerm(SQLEscape(src))

	if remove_accents {
		search := "%"+stripped+"%"
		qstr = fmt.Sprintf("SELECT DISTINCT Chap FROM Node WHERE lower(unaccent(Chap)) LIKE lower('%s')",search)
	} else {
		search := "%"+src+"%"
		qstr = fmt.Sprintf("SELECT DISTINCT Chap FROM Node WHERE lower(Chap) LIKE lower('%s')",search)
	}

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBChaptersMatchingName",err)
	}

	var whole string
	var chapters = make(map[string]int)
	var retval []string

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			several := strings.Split(whole,",")
			
			for s := range several {
				chapters[several[s]]++
			}
		}

		for c := range chapters {
			if strings.Contains(c,src) {
				if len(c) > 0 {
					retval = append(retval,c)
				}
			}
		}

		sort.Strings(retval)
		row.Close()
	}

	return retval
}

// **************************************************************************

func GetDBContextByName(sst PoSST,src string) (string,ContextPtr) {

	var qstr string

	remove_accents,stripped := IsBracketedSearchTerm(src)

	if remove_accents {
		search := stripped
		qstr = fmt.Sprintf("SELECT DISTINCT Context,CtxPtr FROM ContextDirectory WHERE unaccent(Context)='%s'",search)
	} else {
		search := src
		qstr = fmt.Sprintf("SELECT DISTINCT Context,CtxPtr FROM ContextDirectory WHERE Context='%s'",search)
	}

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY GetDBContextByName",err)
	}

	var whole string
	var ptr int

	// Assume unique match for this, to be fixed elsewhere

	if row != nil {
		for row.Next() {
			err = row.Scan(&whole,&ptr)
		}
		row.Close()
	}

	return whole,ContextPtr(ptr)

}

// **************************************************************************

func GetDBContextByPtr(sst PoSST,ptr ContextPtr) (string,ContextPtr) {

	qstr := fmt.Sprintf("SELECT DISTINCT Context,CtxPtr FROM ContextDirectory WHERE CtxPtr=%d",ptr)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBContextssByPtr",err)
	}

	var retctx string
	var retptr int

	// Assume unique match for this, to be fixed elsewhere

	if row != nil {
		for row.Next() {
			err = row.Scan(&retctx,&retptr)
		}

		row.Close()
	}

	return retctx,ContextPtr(retptr)
}

// **************************************************************************

func GetSTtypesFromArrows(arrows []ArrowPtr) []int {

	var sttypes []int

	for a := range arrows {
		sta := ARROW_DIRECTORY[arrows[a]].STAindex
		st := STIndexToSTType(sta)
		sttypes = append(sttypes,st)
	}

	return sttypes
}

// **************************************************************************

func GetDBNodeByNodePtr(sst PoSST,db_nptr NodePtr) Node {

	im_nptr,cached := NODE_CACHE[db_nptr]

	if cached {
		return GetMemoryNodeFromPtr(im_nptr)
	}

	// This ony works if we insert non-null arrays like '[]' during initialization
	cols := I_MEXPR+","+I_MCONT+","+I_MLEAD+","+I_NEAR +","+I_PLEAD+","+I_PCONT+","+I_PEXPR
	qstr := fmt.Sprintf("select L,S,Chap,%s from Node where NPtr='(%d,%d)'::NodePtr AND NOT L=0",cols,db_nptr.Class,db_nptr.CPtr)

	row, err := sst.DB.Query(qstr)

	var n Node
	var count int = 0

	if err != nil {
		fmt.Println("GetDBNodeByNodePointer Failed:",err)
		return n
	}

	var whole [ST_TOP]string

	// NB, there seems to be a "bug" in the SQL package, which cannot always populate the links, so try not to
	//     rely on this and work around when needed using GetEntireCone(any,2..) separately

	if row != nil {
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

		// Expand any dynamic inbuilt functions

		if strings.HasPrefix(n.S,"Dynamic: ") {
			n.S = ExpandDynamicFunctions(n.S)
		}

		row.Close()

		if !cached {
			CacheNode(n)
		}
	}

	n.NPtr = db_nptr
	return n
}

// **************************************************************************

func GetDBSingletonBySTType(sst PoSST,sttypes []int,chap string,cn []string) ([]NodePtr,[]NodePtr) {

	// Used in graph report, analysis

	var qstr,qwhere string
	var dim = len(sttypes)

	context := FormatSQLStringArray(cn)
	chapter := "%"+SQLEscape(chap)+"%"

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
		qwhere += fmt.Sprintf("(array_length(%s::text[],1) IS NOT NULL AND array_length(%s::text[],1) IS NULL AND match_context((%s)[0].Ctx,%s))",stname,stinv,stname,context)
		
		if st != dim-1 {
			qwhere += " OR "
		}
	}

	qstr = fmt.Sprintf("SELECT NPtr FROM Node WHERE lower(Chap) LIKE lower('%s') AND (%s)",chapter,qwhere)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBSingletonBySTType Failed",err,"IN",qstr)
		return nil,nil
	}

	var src_nptrs,snk_nptrs []NodePtr

	if row != nil {
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
	}

	// and sinks  -> -

	qwhere = ""

	for st := 0; st < len(sttypes); st++ {

		stname := STTypeDBChannel(-sttypes[st])
		stinv := STTypeDBChannel(sttypes[st])
		qwhere += fmt.Sprintf("(array_length(%s::text[],1) IS NOT NULL AND array_length(%s::text[],1) IS NULL AND match_context((%s)[0].Ctx,%s))",stname,stinv,stname,context)
		
		if st != dim-1 {
			qwhere += " OR "
		}
	}

	qstr = fmt.Sprintf("SELECT NPtr FROM Node WHERE lower(Chap) LIKE lower('%s') AND (%s)",chapter,qwhere)

	row, err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBSingletonBySTType 2 Failed",err,"IN",qstr)
		return nil,nil
	}

	if row != nil {
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
	}

	return src_nptrs,snk_nptrs
	
}

// **************************************************************************

func SelectStoriesByArrow(sst PoSST,nodeptrs []NodePtr, arrowptrs []ArrowPtr, sttypes []int, limit int) []NodePtr {

	var matches []NodePtr

	// Need to take each arrow type at a time. We can't possibly know if an
	// intentionally promised sequence start (in Node) refers to one arrow or another,
	// but, the chance of being a start for several different independent stories is unlikely.

	// We can always search for ad hoc cases with dream/post-processing if not from N4L
	// Thus a valid story is defined from a start node. It is normally a node with an out-arrow
	// |- NODE --ARROW-->, i.e. no in-arrow entering, but this may be false if the story has
	// loops, like a repeated line in a song chorus.

	for _,n := range nodeptrs {

		// After changes, all these nodes should have Seq = true already from "SolveNodePtrs()"
		// So all the searching is finished, we just need to match the requested arrow

		node := GetDBNodeByNodePtr(sst,n)  // we are now caching this for later
		matches = append(matches,node.NPtr)
	}

	return matches
}

// **************************************************************************
// Retrieve Arrow type data
// **************************************************************************

func GetDBArrowsWithArrowName(sst PoSST,s string) (ArrowPtr,int) {

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	s = strings.Trim(s,"!")

	if s == "" {
		fmt.Println("No such arrow found in database:",s)
		return 0,0
	}

	for a := range ARROW_DIRECTORY {
		if s == ARROW_DIRECTORY[a].Long || s == ARROW_DIRECTORY[a].Short {
			sttype := STIndexToSTType(ARROW_DIRECTORY[a].STAindex)
			return ARROW_DIRECTORY[a].Ptr,sttype
		}
	}

	fmt.Println("No such arrow found in database:",s)
	return 0,0
}

// **************************************************************************

func GetDBArrowsMatchingArrowName(sst PoSST,s string) []ArrowPtr {

	var list []ArrowPtr

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	trimmed := strings.Trim(s,"!")

	if trimmed == "" {
		return list
	}

	if trimmed != s {
		for a := range ARROW_DIRECTORY {
			if ARROW_DIRECTORY[a].Long==trimmed || ARROW_DIRECTORY[a].Short==trimmed {
				list = append(list,ARROW_DIRECTORY[a].Ptr)
			}
		}
	} else {
		for a := range ARROW_DIRECTORY {
			if SimilarString(ARROW_DIRECTORY[a].Long,s) || SimilarString(ARROW_DIRECTORY[a].Short,s) {
				list = append(list,ARROW_DIRECTORY[a].Ptr)
			}
		}
	}

	return list
}

// **************************************************************************

func GetDBArrowByName(sst PoSST,name string) ArrowPtr {

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	name = strings.Trim(name,"!")

	if name == "" {
		return 0
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
				return 0
			}
		}
	}

	return ptr
}

// **************************************************************************

func GetDBArrowByPtr(sst PoSST,arrowptr ArrowPtr) ArrowDirectory {

	if int(arrowptr) > len(ARROW_DIRECTORY) {
		DownloadArrowsFromDB(sst)
	}

	if int(arrowptr) < len(ARROW_DIRECTORY) {
		a := ARROW_DIRECTORY[arrowptr]
		return a
	} else {
		return ARROW_DIRECTORY[0]
	}
		
	return ARROW_DIRECTORY[arrowptr]

}

// **************************************************************************

func GetDBArrowBySTType(sst PoSST,sttype int) []ArrowDirectory {

	var retval []ArrowDirectory

	DownloadArrowsFromDB(sst)

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

func ArrowPtrFromArrowsNames(sst PoSST,arrows []string) ([]ArrowPtr,[]int) {

	// Parse input and discern arrow types, best guess

	var arr []ArrowPtr
	var stt []int

	for a := range arrows {

		// is the entry a number? sttype?

		number, err := strconv.Atoi(arrows[a])
		notnumber := err != nil

		if notnumber {
			arrs := GetDBArrowsMatchingArrowName(sst,arrows[a])
			for  ar := range arrs {
				arrowptr := arrs[ar]
				if arrowptr > 0 {
					arrdir := GetDBArrowByPtr(sst,arrowptr)
					arr = append(arr,arrdir.Ptr)
					stt = append(stt,STIndexToSTType(arrdir.STAindex))
				}
			}
		} else {
			if number < -EXPRESS {
				fmt.Println("Negative arrow value doesn't make sense",number)
			} else if number >= -EXPRESS && number <= EXPRESS {
				stt = append(stt,number)
			} else {
				// whatever remains can only be an arrowpointer
				arrdir := GetDBArrowByPtr(sst,ArrowPtr(number))
				arr = append(arr,arrdir.Ptr)
				stt = append(stt,STIndexToSTType(arrdir.STAindex))
			}
		}
	}

	return arr,stt
}

//******************************************************************

func SolveNodePtrs(sst PoSST,nodenames []string,search SearchParameters,arr []ArrowPtr,limit int) []NodePtr {

	chap := search.Chapter
	cntx := search.Context
	seq := search.Sequence

	// This is a UI/UX wrapper for the underlying lookup, avoiding
	// duplicate results and ordering according to interest

	nodeptrs,rest := ParseLiteralNodePtrs(nodenames)

	var idempotence = make(map[NodePtr]bool)
	var result []NodePtr

	// If we give a precise reference, then that was obviously intended

	for n := range nodeptrs {
		idempotence[nodeptrs[n]] = true
	}

	for r := 0; r < len(rest); r++ {

		// Takes care of general context matching

		nptrs := GetDBNodePtrMatchingNCCS(sst,rest[r],chap,cntx,arr,seq,limit)

		for n := 0; n < len(nptrs); n++ {
			idempotence[nptrs[n]] = true
		}
	}

	// Currently disordered, sort by additional scoring by running context ..

	for uniqnptr := range idempotence {
		result = append(result,uniqnptr)
	}

	sort.Slice(result, ScoreContext)

	return result
}

//******************************************************************

func ScoreContext(i,j int) bool {

	// the more matching items the more relevant

	return true
}

//******************************************************************

func ParseLiteralNodePtrs(names []string) ([]NodePtr,[]string) {

	var current []rune
	var rest []string
	var nodeptrs []NodePtr

	// Note that, when we get here (a,b) is already splut into (a and b)

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

func GetDBPageMap(sst PoSST,chap string,cn []string,page int) []PageMap {

	var qstr string

	chap = strings.Trim(chap,"\"")

	context := FormatSQLStringArray(cn)
	chapter := "%"+chap+"%"

	const hits_per_page = 60
	offset := (page-1) * hits_per_page;

	qstr = fmt.Sprintf("SELECT DISTINCT Chap,Ctx,Line,Path FROM PageMap\n"+
		"WHERE match_context(Ctx,%s)=true AND lower(Chap) LIKE lower('%s') ORDER BY Chap,Line OFFSET %d LIMIT %d",
		context,chapter,offset,hits_per_page)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetDBPageMap Failed:",err,qstr)
	}

	var path string
	var pagemap []PageMap
	var line int
	var ctxptr ContextPtr

	if row != nil {
		for row.Next() {		

			var event PageMap

			err = row.Scan(&chap,&ctxptr,&line,&path)

			if err != nil {
				fmt.Println("Error reading GetDBPageMap",err)
			}

			event.Path = ParseMapLinkArray(path)

			event.Chapter = chap
			event.Context = ctxptr
			event.Line = line;

			pagemap = append(pagemap,event)
		}

		row.Close()
	}

	return pagemap
}

// **************************************************************************
// Bulk DB Retrieval
// **************************************************************************

func GetFwdConeAsNodes(sst PoSST, start NodePtr, sttype,depth int,limit int) []NodePtr {

	qstr := fmt.Sprintf("select unnest(fwdconeasnodes) from FwdConeAsNodes('(%d,%d)',%d,%d,%d);",start.Class,start.CPtr,sttype,depth,limit)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdConeAsNodes Failed",err)
	}

	var whole string
	var n NodePtr
	var retval []NodePtr

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			fmt.Sscanf(whole,"(%d,%d)",&n.Class,&n.CPtr)
			retval = append(retval,n)
		}

		row.Close()
	}

	return retval
}

// **************************************************************************

func GetFwdConeAsLinks(sst PoSST, start NodePtr, sttype,depth int) []Link {

	// This function may be misleading as it doesn't respect paths, may be deprecated in future

	qstr := fmt.Sprintf("select unnest(fwdconeaslinks) from FwdConeAsLinks('(%d,%d)',%d,%d);",start.Class,start.CPtr,sttype,depth)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdConeAsLinks Failed",err)
	}

	var whole string
	var retval []Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			l := ParseSQLLinkString(whole)
			retval = append(retval,l)
		}

		row.Close()
	}

	return retval
}

// **************************************************************************

func GetFwdPathsAsLinks(sst PoSST, start NodePtr, sttype,depth int, maxlimit int) ([][]Link,int) {

	qstr := fmt.Sprintf("SELECT FwdPathsAsLinks from FwdPathsAsLinks('(%d,%d)',%d,%d,%d);",start.Class,start.CPtr,sttype,depth,maxlimit)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdPathsAsLinks Failed",err)
	}

	var whole string
	var retval [][]Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			retval = ParseLinkPath(whole)
		}

		row.Close()
	}

	return retval,len(retval)
}

// **************************************************************************

func GetEntireConePathsAsLinks(sst PoSST,orientation string,start NodePtr,depth int,limit int) ([][]Link,int) {

	// orientation should be "fwd" or "bwd" else "both"

	// Todo: how to limit path search? Usually solutions are small..?

	qstr := fmt.Sprintf("select AllPathsAsLinks from AllPathsAsLinks('(%d,%d)','%s',%d, %d);",
		start.Class,start.CPtr,orientation,depth,limit)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to AllPathsAsLinks Failed",err,qstr)
	}

	var whole string
	var retval [][]Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			retval = ParseLinkPath(whole)
		}

		row.Close()
	}

	sort.Slice(retval, func(i,j int) bool {
		return len(retval[i]) < len(retval[j])
	})

	return retval,len(retval)
}

// **************************************************************************

func GetEntireNCConePathsAsLinks(sst PoSST,orientation string,start []NodePtr,depth int,chapter string,context []string,limit int) ([][]Link,int) {

	// See also GetConstraintConePathsAsLinks for an interface with arrow matching
	// orientation should be "fwd" or "bwd" else "both"

	remove_accents,stripped := IsBracketedSearchTerm(chapter)
	chapter = "%"+stripped+"%"
	rm_acc := "false"

	if remove_accents {
		rm_acc = "true"
	}

	qstr := fmt.Sprintf("select AllNCPathsAsLinks(%s,'%s',%s,%s,'%s',%d,%d);",FormatSQLNodePtrArray(start),chapter,rm_acc,FormatSQLStringArray(context),orientation,depth,limit)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to AllNCPathsAsLinks Failed",err,qstr)
		os.Exit(-1)
	}

	var whole string
	var retval [][]Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			retval = ParseLinkPath(whole)
		}

		row.Close()
	}

	return retval,len(retval)
}

// **************************************************************************

func GetConstraintConePathsAsLinks(sst PoSST,start []NodePtr,depth int,chapter string,context []string,arrowptrs []ArrowPtr,sttypes []int,limit int) ([][]Link,int) {

	// See also GetEntireNCConePathsAsLinks() for a differently optimized interface
	// orientation should be "fwd" or "bwd" else "both"

	remove_accents,stripped := IsBracketedSearchTerm(chapter)
	chapter = "%"+stripped+"%"
	rm_acc := "false"

	if remove_accents {
		rm_acc = "true"
	}

	nod := FormatSQLNodePtrArray(start)
	arr := FormatSQLIntArray(Arrow2Int(arrowptrs))
	stt := FormatSQLIntArray(sttypes)
	cnt := FormatSQLStringArray(context)

	qstr := fmt.Sprintf("select ConstraintPathsAsLinks(%s,'%s',%s,%s,%s,%s,%d,%d);",nod,chapter,rm_acc,cnt,arr,stt,depth,limit)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to ConstraintPathsAsLinks Failed",err,qstr)
		os.Exit(-1)
	}

	var whole string
	var retval [][]Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			retval = ParseLinkPath(whole)
			break
		}

		row.Close()
	}

	return retval,len(retval)
}

// **************************************************************************
// Bulk retrieval helper functions
// **************************************************************************

var MUTEX sync.Mutex

// **************************************************************************

func CacheNode(n Node) {

	_,already := NODE_CACHE[n.NPtr]

	if !already {
		MUTEX.Lock()
		defer MUTEX.Unlock()
		NODE_CACHE[n.NPtr] = AppendTextToDirectory(n,RunErr)
	}
}

// **************************************************************************

func DownloadArrowsFromDB(sst PoSST) {

	// These must be ordered to match in-memory array

	qstr := fmt.Sprintf("SELECT STAindex,Long,Short,ArrPtr FROM ArrowDirectory ORDER BY ArrPtr")

	row, err := sst.DB.Query(qstr)
	
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

	if row != nil {
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
	}

	// Get Inverses

	qstr = fmt.Sprintf("SELECT Plus,Minus FROM ArrowInverses ORDER BY Plus")

	row, err = sst.DB.Query(qstr)
	
	if err != nil {    
		fmt.Println("QUERY Download Inverses Failed",err)
	}

	var plus,minus ArrowPtr

	if row != nil {
		for row.Next() {		

			err = row.Scan(&plus,&minus)

			if err != nil {
				fmt.Println("QUERY Download Arrows Failed",err)
			}

			INVERSE_ARROWS[plus] = minus
		}
		row.Close()
	}
}

// **************************************************************************

func DownloadContextsFromDB(sst PoSST) {

	qstr := fmt.Sprintf("SELECT Context,CtxPtr FROM ContextDirectory ORDER BY CtxPtr")

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY Download Arrows Failed",err)
	}

	CONTEXT_DIRECTORY = nil
	CONTEXT_TOP = 0

	var context string
	var ptr ContextPtr

	if row != nil {
		for row.Next() {		
			err = row.Scan(&context,&ptr)

			var c ContextDirectory

			c.Context = context
			c.Ptr = ptr

			if c.Ptr != CONTEXT_TOP {
				fmt.Println(ERR_MEMORY_DB_CONTEXT_MISMATCH,c,CONTEXT_TOP)
				os.Exit(-1)
			}

			CONTEXT_DIRECTORY = append(CONTEXT_DIRECTORY,c)
			CONTEXT_DIR[context] = CONTEXT_TOP
			CONTEXT_TOP++
		}

		row.Close()
	}
}

// **************************************************************************

func SynchronizeNPtrs(sst PoSST) {

	// If we're merging (not recommended) N4L into an existing db, we need to synch

	for channel := N1GRAM; channel <= GT1024; channel++ {
		
		qstr := fmt.Sprintf("SELECT max((Nptr).CPtr) FROM Node WHERE (Nptr).Chan=%d",channel)

		row, err := sst.DB.Query(qstr)
		
		if err != nil {
			fmt.Println("QUERY Synchronizing nptrs",err)
		}

		var cptr int

		if row != nil {
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
			row.Close()
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

	xyz.X = origin.X + float64(radius * math.Cos(angle))
	xyz.Y = origin.Y + float64(radius * math.Sin(angle))
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

	XChannels := make([]float64,maxlen_tz) // node widths along each path step

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

	XChannels := make([]float64,maxlen_tz)        // node widths along the path
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

func AssignPageCoordinates(maplines []PageMap) map[NodePtr]Coords {

	// Make a quasi causal cone [width][depth] to span the geometry

	var directory = make(map[NodePtr]Coords)
	var already = make(map[NodePtr]bool)
	var axis []NodePtr
	var satellites = make(map[NodePtr][]NodePtr)
	var allnotes int

	// Order unique axial leads and satellite notes

	for depth := 0; depth < len(maplines); depth++ {

		axial_nptr := maplines[depth].Path[0].Dst

		if !already[axial_nptr] {
			allnotes++
			already[axial_nptr] = true
			axis = append(axis,axial_nptr)
		}

		axis := maplines[depth].Path[0].Dst

		for sat := 1; sat < len(maplines[depth].Path); sat++ {
			orbit := maplines[depth].Path[sat].Dst
			if !already[orbit] {
				satellites[axis] = append(satellites[axis],orbit)
				already[orbit] = true
			}
		}
	}

	const screen = 2.0
	const z_start = -1.0
	var zinc = screen / float64(allnotes)

	for tz := 0; tz < len(axis); tz++ {

		var leader Coords

		leader.X = 0
		leader.Y = 0
		leader.Z = z_start + float64(tz) * zinc

		directory[axis[tz]] = leader

		// Arrange the notes orbitally around the leader

		satrange := float64(len(satellites[axis[tz]]))

		for i,sat := range(satellites[axis[tz]]) {

			pos := float64(i)
			radius := 0.3
			var satc Coords
			nptr := sat
			satc.X = pos * radius * math.Cos(2.0 * pos * math.Pi/satrange)
			satc.Y = pos * radius * math.Sin(2.0 * pos * math.Pi/satrange)
			satc.Z = leader.Z

			directory[nptr] = satc
		}

	}

	return directory
}

// **************************************************************************

func AssignChapterCoordinates(nth,swimlanes int) Coords {

	// Place chapters uniformly over the surface of a sphere, using
	// the Fibonacci lattice

	N := float64(swimlanes)
	n := float64(nth)
	const fibratio = 1.618
	const rho = 0.75

	latitude := math.Asin(2 * n / (2 * N + 1))
	longitude := 2 * math.Pi * n/fibratio

	if longitude < -math.Pi {
		longitude += 2 * math.Pi
	}

	if longitude > math.Pi {
		longitude -= 2 * math.Pi
	}

	var fxyz Coords

	fxyz.X = float64(-rho * math.Sin(longitude))
	fxyz.Y = float64(rho * math.Sin(latitude))
	fxyz.Z = float64(rho * math.Cos(longitude) * math.Cos(latitude))

	fxyz.R = rho
	fxyz.Lat = latitude
	fxyz.Lon = longitude

	return fxyz
}

// **************************************************************************

func AssignContextSetCoordinates(origin Coords,nth,swimlanes int) Coords {

	N := float64(swimlanes)
	n := float64(nth)
	latitude := float64(origin.Lat)
	longitude := float64(origin.Lon)
	rho := 0.85

	orbital_angle := math.Pi / 8

	var fxyz Coords

	if N == 1 {
		fxyz.X = -rho * math.Sin(longitude)
		fxyz.Y = rho * math.Sin(latitude)
		fxyz.Z = rho * math.Cos(longitude) * math.Cos(latitude)
		return fxyz
	}

	delta_lon := orbital_angle * math.Sin(2 * math.Pi * n / N)
	delta_lat := orbital_angle * math.Cos(2 * math.Pi * n / N)

	fxyz.X = -rho * math.Sin(longitude+delta_lon)
	fxyz.Y = rho * math.Sin(latitude+delta_lat)
	fxyz.Z = rho * math.Cos(longitude+delta_lon) * math.Cos(latitude+delta_lat)

	return fxyz
}

// **************************************************************************

func AssignFragmentCoordinates(origin Coords,nth,swimlanes int) Coords {

	// These are much more crowded, so stagger radius

	N := float64(swimlanes)
	n := float64(nth)
	latitude := float64(origin.Lat)
	longitude := float64(origin.Lon)

	rho := 0.3 + float64(nth % 2) * 0.1

	orbital_angle := math.Pi / 12

	var fxyz Coords

	if N == 1 {
		fxyz.X = -rho * math.Sin(longitude)
		fxyz.Y = rho * math.Sin(latitude)
		fxyz.Z = rho * math.Cos(longitude) * math.Cos(latitude)
		return fxyz
	}

	delta_lon := orbital_angle * math.Sin(2 * math.Pi * n / N)
	delta_lat := orbital_angle * math.Cos(2 * math.Pi * n / N)

	fxyz.X = -rho * math.Sin(longitude+delta_lon)
	fxyz.Y = rho * math.Sin(latitude+delta_lat)
	fxyz.Z = rho * math.Cos(longitude+delta_lon) * math.Cos(latitude+delta_lat)

	return fxyz
}

// **************************************************************************

func MakeCoordinateDirectory(XChannels []float64, unique [][]NodePtr,maxzlen,nth,swimlanes int) map[NodePtr]Coords {

	var directory = make(map[NodePtr]Coords)

	const totwidth = 2.0 // This is the width dimenion of the paths -1 to +1
	const totdepth = 2.0 // This is the depth dimenion of the paths -1 to +1
	const arbitrary_elevation = 0.0

	x_lanewidth := totwidth / (float64(swimlanes))
	tz_steplength := totdepth / float64(maxzlen) 

	x_lane_start := float64(nth) * x_lanewidth - totwidth/2.0

	// Start allocating swimlane into XChannels parallel spaces
	// x now runs from (x_lane_start to += x_lanewidth)

	for tz := 0; tz < maxzlen && tz < len(unique); tz++ {

		x_increment := x_lanewidth / (XChannels[tz]+1)

		z_left := -float64(totwidth/2)
		x_left := float64(x_lane_start) + x_increment 

		var xyz Coords

		xyz.X = x_left
		xyz.Y = arbitrary_elevation
		xyz.Z = z_left + tz_steplength * float64(tz)

		// Each cross section, at depth tz

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

func GetPathsAndSymmetries(sst PoSST,start_set,end_set []NodePtr,chapter string,context []string,arrowptrs []ArrowPtr,sttypes []int,mindepth,maxdepth int) [][]Link {

	var left_paths, right_paths [][]Link
	var ldepth,rdepth int = 1,1
	var Lnum,Rnum int
	var solutions [][]Link
	var loop_corrections [][]Link

	if start_set == nil || end_set == nil {
		return nil
	}

	if sttypes == nil || len(sttypes)== 0 {
		sttypes = []int{1,2,3,0,-1,-2,-3}
	}

	// Complete Adjoint types for inverse/acceptor wave

	adj_arrowptrs := AdjointArrows(arrowptrs)
	adj_sttypes := AdjointSTtype(sttypes)

	// Prime paths - the different starting points could be parallelized in principle, but we might not win much

	left_paths,Lnum = GetConstraintConePathsAsLinks(sst,start_set,ldepth,chapter,context,arrowptrs,sttypes,maxdepth)
	right_paths,Rnum = GetConstraintConePathsAsLinks(sst,end_set,rdepth,chapter,context,adj_arrowptrs,adj_sttypes,maxdepth)

	fmt.Println("Constraint primer: \nleft",left_paths,"\n\nright",right_paths)

	// There is a minimum depth to avoid finding oneself in the mirror

	fmt.Println("Min-path:",mindepth)
	fmt.Println("Max-path:",maxdepth)
	fmt.Println("STtypes:",sttypes)

	// Expand waves

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		fmt.Print("\r   ..Waves searching: ",ldepth,rdepth)

		solutions,loop_corrections = WaveFrontsOverlap(sst,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		if len(solutions) > mindepth {
			fmt.Println("   ..DAG solutions:",ldepth,rdepth)
			return solutions
		}

		if len(loop_corrections) > mindepth {
			fmt.Println("   ..Only non-DAG solutions:",ldepth,rdepth)
			return loop_corrections
		}

		if turn % 2 == 0 {
			left_paths = IncConstraintConeLinks(sst,left_paths,chapter,context,arrowptrs,sttypes,maxdepth)
			ldepth++
		} else {
			right_paths = IncConstraintConeLinks(sst,right_paths,chapter,context,adj_arrowptrs,adj_sttypes,maxdepth)
			rdepth++
		}
	}

	// Calculate the supernode layer sets S[path][depth], factoring process symmetries

	fmt.Println("HINT: specify \\arrow fwd,bwd inverse-pairs to speed restrict search and speed up search")
	return solutions
}

// **************************************************************************

func IncConstraintConeLinks(sst PoSST,cone [][]Link,chapter string ,context []string,arrowptrs []ArrowPtr,sttypes []int,maxdepth int) [][]Link {

	// Provide an incremental cone expander, so we can preserve state to avoid recomputation
	// This will be increasingly effective as path length increases

	var expanded_cone [][]Link

	for p := 0; p < len(cone); p++ {

		branch := cone[p]
		var exclude = make(map[NodePtr]bool)

		for _,prev := range branch {
			exclude[prev.Dst] = true
		}

		tip := []NodePtr{branch[len(branch)-1].Dst}

		shoots := GetConstrainedFwdLinks(sst,tip,chapter,context,sttypes,arrowptrs,maxdepth)

		// unfurl branches, checking for retracing

		for _,satellite := range shoots {

			if !exclude[satellite.Dst] {
				exclude[satellite.Dst] = true
				var delta []Link
				for _,prev := range branch {
					delta = append(delta,prev)
				}

				delta = append(delta,satellite)
				expanded_cone = append(expanded_cone,delta)
			}
		}
	}

	return expanded_cone
}

// **************************************************************************

func GetConstrainedFwdLinks(sst PoSST,start []NodePtr,chapter string,context []string,sttypes []int,arrows []ArrowPtr,maxlimit int) []Link {

	var ret []Link

	remove_accents,stripped := IsBracketedSearchTerm(chapter)
	chapter = "%"+stripped+"%"
	rm_acc := "false"

	if remove_accents {
		rm_acc = "true"
	}

	start = append(start,NONODE)
	excl := FormatSQLNodePtrArray(start)
	arr := FormatSQLIntArray(Arrow2Int(arrows))
	cnt := FormatSQLStringArray(context)

	startnode := fmt.Sprintf("(%d,%d)",start[0].Class,start[0].CPtr)

	for _,st := range sttypes { 

		qstr := fmt.Sprintf("select GetConstrainedFwdLinks('%s','%s',%s,%s,%s,%d,%s,%d);",startnode,chapter,rm_acc,cnt,excl,st,arr,maxlimit)

		row, err := sst.DB.Query(qstr)
		
		if err != nil {
			fmt.Println("QUERY to ConstraintPathsAsLinks Failed",err,qstr)
			return ret
		}
		
		var whole string

		if row != nil {		
			for row.Next() {		
				err = row.Scan(&whole)
				orbit := ParseLinkArray(whole)
				for _,lnk := range orbit {
					ret = append(ret,lnk)
				}
			}
			row.Close()
		}
	}

	return ret
}

// **************************************************************************

func GetPathsAndSymmetries_legacy(sst PoSST,start_set,end_set []NodePtr,chapter string,context []string,arrowptrs []ArrowPtr,sttypes []int,mindepth,maxdepth int) [][]Link {

	var left_paths, right_paths [][]Link
	var ldepth,rdepth int = 1,1
	var Lnum,Rnum int
	var solutions [][]Link
	var loop_corrections [][]Link

	if start_set == nil || end_set == nil {
		return nil
	}

	// Complete Adjoint types for inverse/acceptor wave

	adj_arrowptrs := AdjointArrows(arrowptrs)
	adj_sttypes := AdjointSTtype(sttypes)

	// Expand waves

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		fmt.Println("   ..Waves searching",ldepth,rdepth)

		// Keep these inside the loop, because there helps curtail exponential growth, despite repetition
		// The interaction of limits can lead to obvious paths being dropped in favour of weird ones if we try 
		// to actor out the search from the start. Compromise by parallelizing the waves.

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done() 
			left_paths,Lnum = GetConstraintConePathsAsLinks(sst,start_set,ldepth,chapter,context,arrowptrs,sttypes,maxdepth)
		}()

		go func() {
			defer wg.Done() 
			right_paths,Rnum = GetConstraintConePathsAsLinks(sst,end_set,rdepth,chapter,context,adj_arrowptrs,adj_sttypes,maxdepth)
		}()

		wg.Wait()

		// end threads

		solutions,loop_corrections = WaveFrontsOverlap(sst,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		if len(solutions) > 0 {
			fmt.Println("   ..DAG solutions:")
			return solutions
		}

		if len(loop_corrections) > 0 {
			fmt.Println("   ..Only non-DAG solutions:")
			return loop_corrections
		}

		if turn % 2 == 0 {
			ldepth++
		} else {
			rdepth++
		}
	}

	// Calculate the supernode layer sets S[path][depth], factoring process symmetries

	fmt.Println("HINT: specify \\arrow fwd,bwd inverse-pairs to speed restrict search and speed up search")
	return solutions
}

// **************************************************************************

func AdjointArrows(arrowptrs []ArrowPtr) []ArrowPtr {

	var idemp = make(map[ArrowPtr]bool)
	var result []ArrowPtr

	for _,a := range arrowptrs {
		idemp[INVERSE_ARROWS[a]] = true
	}

	for a := range idemp {
		result = append(result,a)
	}

	return result
}

// **************************************************************************

func AdjointSTtype(sttypes []int) []int {

	var result []int

	for i := len(sttypes)-1; i >= 0; i-- {
		result = append(result,-sttypes[i])
	}

	return result
}

// **************************************************************************

func GetPathTransverseSuperNodes(sst PoSST,solutions [][]Link,maxdepth int) [][]NodePtr {

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

func WaveFrontsOverlap(sst PoSST,left_paths,right_paths [][]Link,Lnum,Rnum,ldepth,rdepth int) ([][]Link,[][]Link) {

	// The wave front consists of Lnum and Rnum points left_paths[len()-1].
	// Any of the

	var solutions [][]Link
	var loops [][]Link

	// Start expanding the waves from left and right, one step at a time, alternately

	leftfront := WaveFront(left_paths,Lnum)
	rightfront := WaveFront(right_paths,Rnum)

	incidence := NodesOverlap(sst,leftfront,rightfront)
	
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

	for l := 0; l < len(path); l++ {
		front = append(front,path[l][len(path[l])-1].Dst)
	}

	return front
}

// **********************************************************

func NodesOverlap(sst PoSST,left,right []NodePtr) map[int][]int {

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

func GetDBAdjacentNodePtrBySTType(sst PoSST,sttypes []int,chap string,cn []string,transpose bool) ([][]float32,[]NodePtr) {

	// Return a weighted adjacency matrix by nptr, and an index:nptr lookup table
	// Returns a connected adjacency matrix for the subgraph and a lookup table
	// A bit memory intensive, but possibly unavoidable
	
	var qstr,qwhere,qsearch string
	var dim = len(sttypes)

	context := FormatSQLStringArray(cn)
	chapter := "%"+SQLEscape(chap)+"%"

	if dim > 4 {
		fmt.Println("Maximum 4 sttypes in GetDBAdjacentNodePtrBySTType")
		return nil,nil
	}

	for st := 0; st < len(sttypes); st++ {

		stname := STTypeDBChannel(sttypes[st])
		qwhere += fmt.Sprintf("array_length(%s::text[],1) IS NOT NULL AND match_context((%s)[0].Ctx,%s)",stname,stname,context)

		if st != dim-1 {
			qwhere += " OR "
		}

		qsearch += "," + stname

	}

	qstr = fmt.Sprintf("SELECT NPtr%s FROM Node WHERE lower(Chap) LIKE lower('%s') AND (%s)",qsearch,chapter,qwhere)

	row, err := sst.DB.Query(qstr)

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

	if row != nil {
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
		row.Close()
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

func NextLinkArrow(sst PoSST,path []Link,arrows []ArrowPtr) string {

	var rstring string

	if len(path) > 1 {

		for l := 1; l < len(path); l++ {

			if !MatchArrows(arrows,path[l].Arr) {
				break
			}

			nextnode := GetDBNodeByNodePtr(sst,path[l].Dst)
			
			arr := GetDBArrowByPtr(sst,path[l].Arr)
			
			if l < len(path) {
				rstring += fmt.Sprint("  -(",arr.Long,")->  ")
			}
			
			rstring += fmt.Sprint(nextnode.S)
		}
	}

	return rstring
}

// **************************************************************************
//
// Part 3: Model data retrieval, and data marshalling, with JSON etc
//
// **************************************************************************

func GetSequenceContainers(sst PoSST,nodeptrs []NodePtr, arrowptrs []ArrowPtr, sttypes []int, limit int) []Story {

	// Story search

	var stories []Story

	openings := SelectStoriesByArrow(sst,nodeptrs,arrowptrs,sttypes,limit)

	arrname := ""
	count := 0

	for nth := range openings {

		var story Story

		node := GetDBNodeByNodePtr(sst,openings[nth])

		story.Chapter = node.Chap

		axis := GetLongestAxialPath(sst,openings[nth],arrowptrs[0],limit)

		directory := AssignStoryCoordinates(axis,nth,len(openings),limit)

		for lnk := 0; lnk < len(axis); lnk++ {
			
			// Now add the orbit at this node, not including the axis

			var ne NodeEvent

			nd := GetDBNodeByNodePtr(sst,axis[lnk].Dst)

			ne.Text = nd.S
			ne.L = nd.L
			ne.Chap = nd.Chap
			ne.Context = GetContext(axis[lnk].Ctx)
			ne.NPtr = axis[lnk].Dst
			ne.XYZ = directory[ne.NPtr]
			ne.Orbits = GetNodeOrbit(sst,axis[lnk].Dst,arrname,limit)
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

func GetNodeOrbit(sst PoSST,nptr NodePtr,exclude_vector string,limit int) [ST_TOP][]Orbit {

	// radius = 0 is the starting node

	const probe_radius = 3

	// Find the orbiting linked nodes of NPtr, start with properties of node

	sweep,_ := GetEntireConePathsAsLinks(sst,"any",nptr,probe_radius,limit)

	var satellites [ST_TOP][]Orbit
	var thread_wg sync.WaitGroup

	for stindex := 0; stindex < ST_TOP; stindex++ {

		// Go routines remain a mystery
		thread_wg.Add(1)
		
		go func(idx int) {
			defer thread_wg.Done()  // threading
			
			satellites[idx] = AssembleSatellitesBySTtype(sst,idx,satellites[idx],sweep,exclude_vector,probe_radius,limit)
			
		} (stindex)
	}
	
	thread_wg.Wait()

	return satellites
}

// **************************************************************************

func AssembleSatellitesBySTtype(sst PoSST,stindex int,satellite []Orbit,sweep [][]Link,exclude_vector string,probe_radius int,limit int) []Orbit {

	var already = make(map[string]bool)

	// Sweep different radial paths [angle][depth]
	
	for angle := 0; angle < len(sweep); angle++ {
		
		// len(sweep[angle]) is the length of the probe path at angle
		
		if sweep[angle] != nil && len(sweep[angle]) > 1 {
			
			const nearest_satellite = 1
			start := sweep[angle][nearest_satellite]
			
			arrow := GetDBArrowByPtr(sst,start.Arr)
			
			if arrow.STAindex == stindex {

				txt := GetDBNodeByNodePtr(sst,start.Dst)

				var nt Orbit				
				nt.Arrow = arrow.Long
				nt.STindex = arrow.STAindex
				nt.Dst = start.Dst
				nt.Text = txt.S
				if txt.I[LEADSTO] != nil {
					nt.Ctx = GetContext(txt.I[LEADSTO][0].Ctx)  // node context
				} else {
					nt.Ctx = "any"
				}
				nt.Radius = 1
				if arrow.Long == exclude_vector || arrow.Short == exclude_vector {
					continue
				}
				
				satellite = IdempAddSatellite(satellite,nt,already)
				
				// are there more satellites at this angle?
				
				for depth := 2; depth < probe_radius && depth < len(sweep[angle]); depth++ {
					
					arprev := STIndexToSTType(arrow.STAindex)
					next := sweep[angle][depth]
					arrow = GetDBArrowByPtr(sst,next.Arr)
					subtxt := GetDBNodeByNodePtr(sst,next.Dst)
					
					if arrow.Long == exclude_vector || arrow.Short == exclude_vector {
						break
					}
					
					nt.Arrow = arrow.Long
					nt.STindex = arrow.STAindex
					nt.Dst = next.Dst
					nt.Ctx = GetContext(next.Ctx)
					nt.Text = subtxt.S
					nt.Radius = depth
					
					arthis := STIndexToSTType(arrow.STAindex)
					// No backtracking
					if arthis != -arprev {	
						satellite = IdempAddSatellite(satellite,nt,already)
						arprev = arthis
					}
				}
			}
		}
	}

	return satellite
}

// **************************************************************************

func IdempAddSatellite(list []Orbit, item Orbit,already map[string]bool) []Orbit {

	// crude check but effective, since the list is fairly short unless the graph is sick

	key := fmt.Sprintf("%v,%s",item.Dst,item.Arrow)

	if already[key] {
		return list
	} else {
		already[key] = true
		return append(list,item)
	}
}

// **************************************************************************

func GetLongestAxialPath(sst PoSST,nptr NodePtr,arrowptr ArrowPtr,limit int) []Link {

	// Used in story search along extended STtype paths

	var max int = 1

	sttype := STIndexToSTType(ARROW_DIRECTORY[arrowptr].STAindex)

	paths,dim := GetFwdPathsAsLinks(sst,nptr,sttype,limit,limit)

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

// *********************************************************************
// Access stats, LTM progress tracking
// *********************************************************************

func UpdateLastSawSection(sst PoSST,name string) {

	s := fmt.Sprintf("select LastSawSection('%s')",name)
	sst.DB.QueryRow(s)
}

// *********************************************************************

func UpdateLastSawNPtr(sst PoSST,class,cptr int,name string) {

	s := fmt.Sprintf("select LastSawNPtr('(%d,%d)','%s')",class,cptr,name)
	sst.DB.QueryRow(s)
}

//******************************************************************

func GetLastSawSection(sst PoSST) []LastSeen {

	qstr := fmt.Sprintf("SELECT section,nptr,last,freq,delta as pdelta,EXTRACT(EPOCH FROM NOW()-last) as ndelta from Lastseen ORDER BY section")

	row,err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetLastSawSection failed\n",qstr,err)
		return nil
	}

	var ret []LastSeen

	if row != nil {
		for row.Next() {		
			var ls LastSeen
			var nptrstr string // last because if empty fails
			
			err = row.Scan(&ls.Section,&nptrstr,&ls.Last,&ls.Freq,&ls.Pdelta,&ls.Ndelta)
			fmt.Sscanf(nptrstr,"(%d,%d)",&ls.NPtr.Class,&ls.NPtr.CPtr)
			
			ret = append(ret,ls)
		}

		for c := 0; c < len(ret); c++ {
			ret[c].XYZ = AssignChapterCoordinates(c,len(ret))
		}

		row.Close()
	}

	return ret
}

//******************************************************************

func GetLastSawNPtr(sst PoSST, nptr NodePtr) LastSeen {

	var ls LastSeen

	qstr := fmt.Sprintf("SELECT section,last,freq,delta as pdelta,EXTRACT(EPOCH FROM NOW()-last) as ndelta from Lastseen WHERE NPTR='(%d,%d)'::NodePtr",nptr.Class,nptr.CPtr)

	row,err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetLastSawNPtr failed\n",qstr,err)
		return ls
	}

	if row != nil {
		for row.Next() {		
			err = row.Scan(&ls.Section,&ls.Last,&ls.Freq,&ls.Pdelta,&ls.Ndelta)
		}

		ls.NPtr = nptr

		row.Close()
	}

	return ls
}

// *********************************************************************
// Dynamic context / sensory input evaluation STM tracking
// *********************************************************************

type History struct {

	Freq  float64  // just use float becasue we'll want to calculate
	Last  int64    // calc gradient and purge
	Delta int64
	Time  string
}

// *********************************************************************

var STM_INT_FRAG = make(map[string]History) // for intentional (exceptional) fragments
var STM_AMB_FRAG = make(map[string]History) // for ambient (repeated) fragments
var STM_INV_GROUP = make(map[string]History) // look for invariants

const FORGOTTEN = 10800
const TEXT_SIZE_LIMIT = 30

// *********************************************************************

func UpdateSTMContext(sst PoSST,ambient,key string,now int64,params SearchParameters) string {

	var context []string

	if params.Sequence || params.From != nil || params.To != nil {
		// path / cone are intended
		context = append(context,params.Name...)
		context = append(context,params.From...)
		context = append(context,params.To...)
		return AddContext(sst,ambient,key,now,context)
	} else {
		// ongoing / adhoc are ambient
		context = append(context,params.Name...)

		for _,ct := range params.Context {
			if ct != "" {
				context = append(context,ct)
			}
		}

		if params.Chapter != "" {
			context = append(context,"Chapter:"+params.Chapter)
		}

		return AddContext(sst,ambient,key,now,context)
	}

	return ""
}

// *********************************************************************

func AddContext(sst PoSST,ambient,key string,now int64,tokens []string) string {

	for t := range tokens {

		token := tokens[t]

		if len(token) == 0 || token == "%%" {
			continue
		}

		// Check for direct NPtr click, watch out for long text

		if token[0] == '(' {
			var nptr NodePtr
			fmt.Sscanf(token,"(%d,%d)",&nptr.Class,&nptr.CPtr)

			if nptr.Class > 0 {
				node := GetDBNodeByNodePtr(sst,nptr)
				if node.L < TEXT_SIZE_LIMIT {
					token = node.S
				} else {
					token = node.S[0:TEXT_SIZE_LIMIT] + "..."
				}
			} else {
				continue
			}
		}
		CommitContextToken(token,now,ambient)
	}

	var format = make(map[string]int)

	for fr := range STM_AMB_FRAG {

		if STM_AMB_FRAG[fr].Delta > FORGOTTEN {
			delete(STM_AMB_FRAG,fr)
			continue
		} 

		format[fr]++
	}

	for fr := range STM_INT_FRAG {

		if STM_INT_FRAG[fr].Delta > FORGOTTEN {
			delete(STM_INT_FRAG,fr)
			continue
		} 

		format[fr]++
	}

	full_context := List2String(Map2List(format))

	return full_context
}

// *********************************************************************

func CommitContextToken(token string,now int64,key string) {
	
	var last,obs History
	
	// Check if already known ambient
	last,already := STM_AMB_FRAG[token]
	
	// if not, then check if already seen
	if !already {
		last,already = STM_INT_FRAG[token]
	}
	
	if !already {
		last.Last = now
	}
	
	obs.Freq = last.Freq + 1
	obs.Last = now
	obs.Time = key
	obs.Delta = now - last.Last
	
	if obs.Freq > 1 {
		pr,okey := DoNowt(time.Unix(last.Last,0))
		fmt.Printf("    - last saw \"%s\" at %s (%s)\n",token,pr,okey)
	}
	
	if already {
		delete(STM_INT_FRAG,token)
		STM_AMB_FRAG[token] = obs
	} else {
		STM_INT_FRAG[token] = obs
	}
}

// *********************************************************************

func ContextInterferometry(now_sst string) {

 // deleted

}

// *********************************************************************

func ShowContext(amb,intent,key string) {

	fmt.Println()
	fmt.Println("  .......................................................")
	fmt.Printf("    Recurrent now: %s\n",key)
	fmt.Printf("    Intentional  : %s\n",intent)
	fmt.Printf("    Ambient      : %s\n",amb)
	fmt.Println("  .......................................................")

}

// **************************************************************************

func IntersectContextParts(context_clusters []string) (int,[]string,[][]int)  {

	// return a weighted upper triangular matrix of overlaps between frags,
	// and an idempotent list of fragments

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

func PrintNodeOrbit(sst PoSST, nptr NodePtr,limit int) {

	node := GetDBNodeByNodePtr(sst,nptr)		
	fmt.Print("\"")
	ShowText(node.S,SCREENWIDTH)
	fmt.Print("\"")
	fmt.Println("\tin chapter:",node.Chap)
	fmt.Println()

	satellites := GetNodeOrbit(sst,nptr,"",limit)

	PrintLinkOrbit(satellites,EXPRESS,0)
	PrintLinkOrbit(satellites,-EXPRESS,0)
	PrintLinkOrbit(satellites,-CONTAINS,0)
	PrintLinkOrbit(satellites,LEADSTO,0)
	PrintLinkOrbit(satellites,-LEADSTO,0)
	PrintLinkOrbit(satellites,NEAR,0)

	fmt.Println()
}

// **************************************************************************

func PrintLinkOrbit(satellites [ST_TOP][]Orbit,sttype int,indent_level int) {

	t := STTypeToSTIndex(sttype)

	for n := range satellites[t] {		

		r := satellites[t][n].Radius + indent_level

		if satellites[t][n].Ctx != "" {
			txt := fmt.Sprintf(" -    (%s) - %s  \t.. in the context of %s\n",satellites[t][n].Arrow,satellites[t][n].Text,satellites[t][n].Ctx)
			text := Indent(LEFTMARGIN * r) + txt
			ShowText(text,SCREENWIDTH)
		} else {
			txt := fmt.Sprintf(" -    (%s) - %s\n",satellites[t][n].Arrow,satellites[t][n].Text)
			text := Indent(LEFTMARGIN * r) + txt
			ShowText(text,SCREENWIDTH)
		}

	}

}

// **************************************************************************

func PrintLinkPath(sst PoSST, cone [][]Link, p int, prefix string,chapter string,context []string) {

	PrintSomeLinkPath(sst,cone, p,prefix,chapter,context,10000)
}

// **************************************************************************

func PrintSomeLinkPath(sst PoSST, cone [][]Link, p int, prefix string,chapter string,context []string,limit int) {

	count := 0

	if len(cone[p]) > 1 {

		path_start := GetDBNodeByNodePtr(sst,cone[p][0].Dst)		
		
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

			nextnode := GetDBNodeByNodePtr(sst,cone[p][l].Dst)

			if !SimilarString(nextnode.Chap,chapter) {
				break
			}

			arr := GetDBArrowByPtr(sst,cone[p][l].Arr)

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
// Rendering, Marshalling, Presentation in JSON
// **************************************************************************

func JSONNodeEvent(sst PoSST, nptr NodePtr,xyz Coords,orbits [ST_TOP][]Orbit) NodeEvent {

	node := GetDBNodeByNodePtr(sst,nptr)

	var event NodeEvent
	event.Text = node.S
	event.L = node.L
	event.Chap = node.Chap
	event.Context = GetNodeContextString(sst,node)
	event.NPtr = nptr
	event.XYZ = xyz
	event.Orbits = orbits
	return event
}

// **************************************************************************

func LinkWebPaths(sst PoSST,cone [][]Link,nth int,chapter string,context []string,swimlanes,limit int) [][]WebPath {

	// This is dealing in good faith with one of swimlanes cones, assigning equal width to all
	// The cone is a flattened array, we can assign spatial coordinates for visualization

	var conepaths [][]WebPath

	directory := AssignConeCoordinates(cone,nth,swimlanes)

	// JSONify the cone structure, converting []Link into []WebPath

	for p := 0; p < len(cone); p++ {

		path_start := GetDBNodeByNodePtr(sst,cone[p][0].Dst)		
		
		start_shown := false

		var path []WebPath
		
		for l := 1; l < len(cone[p]); l++ {

			if !MatchContexts(context,cone[p][l].Ctx) {
				break
			}

			nextnode := GetDBNodeByNodePtr(sst,cone[p][l].Dst)

			if !SimilarString(nextnode.Chap,chapter) {
				break
			}
			
			if !start_shown {
				var ws WebPath
				ws.Name = path_start.S
				ws.NPtr = cone[p][0].Dst
				ws.Chp = nextnode.Chap
				ws.XYZ = directory[cone[p][0].Dst]
				path = append(path,ws)
				start_shown = true
			}

			arr := GetDBArrowByPtr(sst,cone[p][l].Arr)
	
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
			wn.Chp = nextnode.Chap
			wn.NPtr = cone[p][l].Dst
			wn.XYZ = directory[cone[p][l].Dst]
			path = append(path,wn)

		}
		conepaths = append(conepaths,path)
	}

	return conepaths
}

// **************************************************************************

func GetChaptersByChapContext(sst PoSST,chap string,cn []string,limit int) map[string][]string {

	qstr := ""
	chap_col := ""

	chap = strings.Trim(chap,"\"")

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

	qstr = fmt.Sprintf("SELECT DISTINCT chap,ctx FROM PageMap WHERE match_context(ctx,%s) %s ORDER BY Chap",context,chap_col)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetChaptersByChapContext Failed",err,qstr)
	}

	var rchap string
	var rcontext ContextPtr
	var toc = make(map[string][]string)

	if row != nil {
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

				cn := strings.Split(GetContext(rcontext),",")
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
	}

	return toc
}

// **************************************************************************

func JSONPage(sst PoSST, maplines []PageMap) string {

	var webnotes PageView
	var lastchap,lastctx string
	var signalchap, signalctx, signalchange string
	var warned bool = false

	directory := AssignPageCoordinates(maplines)

	for n := 0; n < len(maplines); n++ {

		var path []WebPath

		txtctx := GetContext(maplines[n].Context)

		// Format superheader aggregate summary

		if lastchap != maplines[n].Chapter {
			if !warned {
				webnotes.Title = webnotes.Title
				warned = true
			}
			webnotes.Title += maplines[n].Chapter + ", "
			lastchap = maplines[n].Chapter
			signalchap = maplines[n].Chapter
		} else {
			signalchap = ""
		}

		if lastctx != txtctx {
			webnotes.Context += txtctx + ", " 
			lastctx = txtctx
			signalctx = txtctx
		} else {
			signalctx = txtctx
		}

		signalchange = signalchap + " :: " + signalctx

		// Next line item

		for lnk := 0; lnk < len(maplines[n].Path); lnk++ {
			
			text := GetDBNodeByNodePtr(sst,maplines[n].Path[lnk].Dst)
			
			if lnk == 0 {
				var ws WebPath
				ws.Name = text.S
				ws.NPtr = maplines[n].Path[lnk].Dst
				ws.XYZ = directory[ws.NPtr]
				ws.Chp = maplines[n].Chapter
				ws.Line = maplines[n].Line
				ws.Ctx = GetContext(maplines[n].Context)
				path = append(path,ws)
				
			} else {// ARROW
				arr := GetDBArrowByPtr(sst,maplines[n].Path[lnk].Arr)
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
				ws.Chp = maplines[n].Chapter
				ws.Ctx = signalchange
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

func GetAppointedNodesByArrow(sst PoSST,arrow ArrowPtr,cn []string,chap string,size int) map[ArrowPtr][]Appointment {

	// return a map of all the nodes in chap,context that are pointed to by the same type of arrow
        // grouped by arrow

	reverse_arrow := INVERSE_ARROWS[arrow]
	arr := GetDBArrowByPtr(sst,reverse_arrow)
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

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointedNodesByArrow Failed",err,qstr)
	}

	var whole string

	var retval = make(map[ArrowPtr][]Appointment)
	
	if row != nil {
		for row.Next() {
			err = row.Scan(&whole) //arrint,&sttype,&rchap,&rctx,&apex,&arry)

			next := ParseAppointedNodeCluster(whole)
			retval[next.Arr] = append(retval[next.Arr],next)
		}
	
		row.Close()
	}

	return retval
}

// **************************************************************************

func GetAppointedNodesBySTType(sst PoSST,sttype int,cn []string,chap string,size int) map[ArrowPtr][]Appointment {

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

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointedNodesByArrow Failed",err,qstr)
	}

	var whole string

	var retval = make(map[ArrowPtr][]Appointment)
	
	if row != nil {
		for row.Next() {
			err = row.Scan(&whole) //arrint,&sttype,&rchap,&rctx,&apex,&arry)

			next := ParseAppointedNodeCluster(whole)
			retval[next.Arr] = append(retval[next.Arr],next)
		}	
		row.Close()
	}

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

func TallyPath(sst PoSST,path []Link,between map[string]int) map[string]int {

	// count how often each node appears in the different path solutions

	for leg := range path {
		n := GetDBNodeByNodePtr(sst,path[leg].Dst)
		between[n.S]++
	}

	return between
}

// **************************************************************************

func BetweenNessCentrality(sst PoSST,solutions [][]Link) []string {

	var betweenness = make(map[string]int)

	for s := 0; s < len(solutions); s++ {
		betweenness = TallyPath(sst,solutions[s],betweenness)
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

	var retval []string
	var betw string

	for key := len(order)-1; key >= 0; key-- {

		betw = fmt.Sprintf("%.2f : ",float32(order[key])/float32(len(solutions)))

		for el := 0; el < len(inv[order[key]]); el++ {

			betw += fmt.Sprintf("%s",inv[order[key]][el])
			if el < len(inv[order[key]])-1 {
				betw += ", "
			}
		}

		retval =  append(retval,betw)
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

func SuperNodes(sst PoSST,solutions [][]Link, maxdepth int) []string {

	supernodes := SuperNodesByConicPath(solutions,maxdepth)

	var retval []string

	for g := range supernodes {

		super := ""

		for n := range supernodes[g] {
			node := GetDBNodeByNodePtr(sst,supernodes[g][n])
			super += fmt.Sprintf("%s",node.S)
			if n < len(supernodes[g])-1 {
				super += ", "
			}
		}

		if g < len(supernodes)-1 {
			retval = append(retval,super)
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
	Min      []int
	Max      []int
	Finds    []string
	Sequence bool
	Stats    bool
}

// ******************************************************************
// Short Term Memory (STM) intent event capture
// ******************************************************************

type STM struct {

	Query SearchParameters

}

// ******************************************************************

const (

	CMD_ON = "\\on"
	CMD_ON_2 = "on"    // _2 are too short to be intentional
	CMD_FOR = "\\for"  // so double these for "smarter" accident avoidance
	CMD_FOR_2 = "for"
	CMD_ABOUT = "\\about"
	CMD_NOTES = "\\notes"
	CMD_BROWSE = "\\browse"
	CMD_PAGE = "\\page"
	CMD_PATH = "\\path"
	CMD_SEQ1 = "\\sequence"
	CMD_SEQ2 = "\\seq"
	CMD_STORY = "\\story"
	CMD_STORIES = "\\stories"
	CMD_FROM = "\\from"
	CMD_TO = "\\to"
	CMD_TO_2 = "to"
	CMD_CTX = "\\ctx"
	CMD_CONTEXT = "\\context"
	CMD_AS = "\\as"
	CMD_AS_2 = "as"
	CMD_CHAPTER = "\\chapter"
	CMD_CONTENTS = "\\contents"
	CMD_TOC = "\\toc"
	CMD_TOC_2 = "toc"
	CMD_MAP = "\\map"
	CMD_SECTION = "\\section"
	CMD_IN = "\\in"
	CMD_IN_2 = "in"
	CMD_ARROW = "\\arrow"
	CMD_ARROWS = "\\arrows"
	CMD_LIMIT = "\\limit"
	CMD_DEPTH = "\\depth"
	CMD_RANGE = "\\range"
	CMD_DISTANCE = "\\distance"
	CMD_STATS = "\\stats"
	CMD_STATS_2 = "stats"
	CMD_REMIND = "\\remind"
	CMD_HELP = "\\help"
	CMD_HELP_2 = "help"
	// What to find in orbit
	CMD_FINDS = "\\finds"
	CMD_FINDING = "\\finding"
	// bounding linear path and parallel arrows
	CMD_GT = "\\gt"
	CMD_LT = "\\lt"
	CMD_MIN = "\\min"	
	CMD_MAX = "\\max"
	CMD_ATLEAST = "\\atleast"
	CMD_ATMOST = "\\atmost"

)

//******************************************************************
// Decoding local receiver (-) intent
//******************************************************************

func DecodeSearchField(cmd string) SearchParameters {

	var keywords = []string{ 
		CMD_NOTES, CMD_BROWSE, CMD_PATH,
		CMD_PATH,CMD_FROM,CMD_TO,CMD_TO_2,
		CMD_SEQ1,CMD_SEQ2,CMD_STORY,CMD_STORIES,
		CMD_CONTEXT,CMD_CTX,CMD_AS,CMD_AS_2,
		CMD_CHAPTER,CMD_IN,CMD_IN_2,CMD_SECTION,CMD_CONTENTS,CMD_TOC,CMD_TOC_2,CMD_MAP,
		CMD_ARROW,CMD_ARROWS,
		CMD_GT,CMD_MIN,CMD_ATLEAST,
		CMD_LT,CMD_MAX,CMD_ATMOST,
		CMD_ON,CMD_ON_2,CMD_ABOUT,CMD_FOR,CMD_FOR_2,
		CMD_PAGE,
		CMD_LIMIT,CMD_RANGE,CMD_DISTANCE,CMD_DEPTH,
		CMD_STATS,CMD_STATS_2,
		CMD_REMIND,
		CMD_HELP,CMD_HELP_2,
		CMD_FINDS,CMD_FINDING,
        }
	
	// parentheses are reserved for unaccenting

	cmd = strings.ToLower(cmd)

	m := regexp.MustCompile("[ \t]+") 
	cmd = m.ReplaceAllString(cmd," ") 

	cmd = strings.TrimSpace(cmd)
	pts := SplitQuotes(cmd)

	var parts [][]string
	var part []string

	for p := 0; p < len(pts); p++ {

		subparts := SplitQuotes(pts[p])

		for w := 0; w < len(subparts); w++ {

			if IsCommand(subparts[w],keywords) {
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

	for arg := range param.Name {

		isdirac,beg,end,cnt := DiracNotation(param.Name[arg])
		
		if isdirac {
			param.Name = nil
			param.From = []string{beg}
			param.To = []string{end}
			param.Context = []string{cnt}
			break
		}
	}

	return param
}

//******************************************************************

func FillInParameters(cmd_parts [][]string,keywords []string) SearchParameters {

	var param SearchParameters 

	for c := 0; c < len(cmd_parts); c++ {

		lenp := len(cmd_parts[c])

		for p := 0; p < lenp; p++ {

			switch SomethingLike(cmd_parts[c][p],keywords) {

			case CMD_STATS, CMD_STATS_2:
				param.Stats = true
				continue

			case CMD_HELP, CMD_HELP_2:
				param.Chapter = "SSTorytime help"
				param.Name = []string{"any"}
				continue

			case CMD_CHAPTER,CMD_SECTION,CMD_IN,CMD_IN_2,CMD_CONTENTS,CMD_TOC,CMD_TOC_2,CMD_MAP:
				if lenp > p+1 {
					str := cmd_parts[c][p+1]
					str = strings.TrimSpace(str)
					str = strings.Trim(str,"'")
					str = strings.Trim(str,"\"")
					param.Chapter = str
					break
				} else {
					param.Chapter = "TableOfContents"
					break					
				}
				continue

			case CMD_NOTES, CMD_BROWSE:
				if param.PageNr < 1 {
					param.PageNr = 1
				}
			
				if lenp > p+1 {
					param.Chapter = cmd_parts[c][p+1]
				} else {
					if lenp > 1 {
						param = AddOrphan(param,cmd_parts[c][p+1])
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

			case CMD_GT,CMD_MIN,CMD_ATLEAST:
				// if followed by a number, else could be search term
				if lenp > p+1 {
					p++
					var no int = -1
					fmt.Sscanf(cmd_parts[c][p],"%d",&no)
					if no > 0 {
						param.Min = append(param.Min,no)
					} else {
						param = AddOrphan(param,cmd_parts[c][p-1])
						param = AddOrphan(param,cmd_parts[c][p])
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_LT,CMD_MAX,CMD_ATMOST:
				// if followed by a number, else could be search term
				if lenp > p+1 {
					p++
					var no int = -1
					fmt.Sscanf(cmd_parts[c][p],"%d",&no)
					if no > 0 {
						param.Max = append(param.Max,no)
					} else {
						param = AddOrphan(param,cmd_parts[c][p-1])
						param = AddOrphan(param,cmd_parts[c][p])
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_ARROW,CMD_ARROWS:
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
				
				case CMD_CONTEXT,CMD_CTX,CMD_AS,CMD_AS_2:
				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							class := strings.TrimSpace(DeQ(ult[u]))
							if len(class) > 0 {
								param.Context = append(param.Context,class)
							}
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
						if !IsLiteralNptr(cmd_parts[c][pp]) {
							ult := strings.Split(cmd_parts[c][pp],",")
							for u := range ult {
								param.From = append(param.From,DeQ(ult[u]))
							}
						} else {
							param.From = append(param.From,cmd_parts[c][pp])
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_TO,CMD_TO_2:
				if p > 0 && lenp > p+1 {
					if param.From == nil {
						param.From = append(param.From,cmd_parts[c][p-1])
					}

					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						if !IsLiteralNptr(cmd_parts[c][pp]) {
							ult := strings.Split(cmd_parts[c][pp],",")
							for u := range ult {
								param.To = append(param.To,DeQ(ult[u]))
							}
						} else {
							param.To = append(param.From,cmd_parts[c][pp])
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

			case CMD_PATH,CMD_SEQ1,CMD_SEQ2,CMD_STORY,CMD_STORIES:
				param.Sequence = true
				continue

			case CMD_ON,CMD_ON_2,CMD_ABOUT,CMD_FOR,CMD_FOR_2:
				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						if param.PageNr > 0 {
							param.Chapter = cmd_parts[c][pp]
						} else {
							ult := strings.Split(cmd_parts[c][pp]," ")
							for u := range ult {
								param.Name = append(param.Name,DeQ(ult[u]))
							} 
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_FINDS,CMD_FINDING:

				for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
					p++
					ult := SplitQuotes(cmd_parts[c][pp])
					for u := range ult {
						if ult[u] == "any" {
							ult[u] = "%%"
						}
						param.Finds = append(param.Finds,DeQ(ult[u]))
					}
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
						if ult[u] == "any" {
							ult[u] = "%%"
						}
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

	const min_sense = 4

	if i >= lenp {
		return false
	}

	key := keys[i]

	if IsCommand(key,keywords) {
		return false
	}

	return true
}

//******************************************************************

func MinMaxPolicy(search SearchParameters) (int,int) {

	// The min max doubles as context dependent role as
	// i) limits on path length and ii) limits on matches arrow matches

	minlimit := 1
	maxlimit := 0
	from := search.From != nil
	to := search.To != nil

	// Validate

	if len(search.Min) > 4 {
		fmt.Println("\nWARNING: minimum arrow matches exceeds the number of ST-types")
	}

	if len(search.Max) > 4 {
		fmt.Println("\nWARNING: maximum arrow matches exceeds the number of ST-types")
	}

	if len(search.Min) != 4 && len(search.Max) != 4 {

		// Only "abusing" the min max for linear path length or search depth

		if len(search.Min) > 0 && len(search.Max) > 0 {
			if search.Min[0] > search.Max[0] {
				fmt.Println("\nWARNING: minimum arrow limit greater than maximum limit!")
				fmt.Println("Depth/range:","min =",search.Min[0],", max =",search.Max[0])
			}
			
			if len(search.Min) == 1 {
				minlimit = search.Min[0]
			}
			
		} else if len(search.Max) == 1 && search.Range > 0 {
			fmt.Println("\nWARNING: conflict between \\depth,\\range and \\max,\\lt,\\atmost ")
		}
	} else {
		// Full ST-type arrow match limits

		for i := 0; i < 4; i++ {
			if search.Min[i] > search.Max[i] {
				fmt.Println("\nWARNING: minimum arrow limit greater than maximum limit!")
				fmt.Println("ST-type:",i,"min =",search.Min[i],", max =",search.Max[i])
			}
		}
	}

	// Defaults

	if search.Chapter == "TableOfContents" {

		// We want to see all contents
		maxlimit = 50

	} else if search.Range > 0 {

		maxlimit = search.Range

	} else if len(search.Max) == 1 {  // if only one, we probably meant Range

		maxlimit = search.Max[0]

	} else {

		if from || to || search.Sequence {
			maxlimit = 30 // many paths make hard work
		} else {
			const common_word = 5

			if SearchTermLen(search.Name) < common_word {
				maxlimit = 5
			} else {
				maxlimit = 10
			}

			if len(search.Name) < 3 && AllExact(search.Name) {
				maxlimit = 30
			}
		}
	}

	return minlimit,maxlimit
}

//******************************************************************

func AllExact(list []string) bool {

	is_exact := false

	for _,s := range list {
		is,_ := IsExactMatch(s)
		is_exact = is_exact || is
	}

	return is_exact
}

//******************************************************************

func IsLiteralNptr(s string) bool {
	
	var a,b int = -1,-1

	s = strings.TrimSpace(s)
	
	fmt.Sscanf(s,"(%d,%d)",&a,&b)
	
	if a >= 0 && b >= 0 {
		return true
	}
	
	return false
}

//******************************************************************

func SomethingLike(s string,keywords []string) string {

	const min_sense = 4

	for k := 0; k < len(keywords); k++ {

		if s == keywords[k] {
			return keywords[k]
		}

		if len(s) > min_sense && len(keywords[k]) > min_sense {
			if strings.HasPrefix(s,keywords[k]) {
				return keywords[k]
			}
		}
	}
	return s
}

//******************************************************************

func IsCommand(s string,list []string) bool {

	const min_sense = 5

	for w := range list {
		if list[w] == s {
			return true
		}

		// Allow likely abbreviations ?

		if len(list[w]) > min_sense && strings.HasPrefix(s,list[w]) {
			return true
		}
	}
	return false
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

	if len(upto) > 0 {
		items = append(items,string(upto))
	}

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
	"NONE",
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

	// In this version, we need less accuracy and greater semantic distinction
	// so prefix temporal classes with a :

	year := fmt.Sprintf("Yr%d",then.Year())
	month := GR_MONTH_TEXT[int(then.Month())]
	day := then.Day()
	hour := fmt.Sprintf("Hr%02d",then.Hour())
	quarter := fmt.Sprintf("Qu%d",then.Minute()/15 + 1)
	shift :=  fmt.Sprintf("%s",GR_SHIFT_TEXT[then.Hour()/6])

	//secs := then.Second()
	//nano := then.Nanosecond()
	//mins := fmt.Sprintf("Min%02d",then.Minute())

	n_season,s_season := Season(month)

	dayname := then.Weekday()
	dow := fmt.Sprintf("%.3s",dayname)
	daynum := fmt.Sprintf("Day%d",day)

	// 5 minute resolution capture is too fine grained for most human interest
        interval_start := (then.Minute() / 5) * 5
        interval_end := (interval_start + 5) % 60
        minD := fmt.Sprintf("Min%02d_%02d",interval_start,interval_end)

	// Don't include the time key in general context, as it varies too fast to be meaningful

	var when string = fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s, %s",n_season,s_season,shift,dayname,daynum,month,year,hour,quarter)
	var key string = fmt.Sprintf("%s:%s:%s-%s",dow,hour,quarter,minD)

	return when, key
}

// ****************************************************************************

func GetTimeContext() (string,string,int64) {

	now := time.Now()
	context,keyslot := DoNowt(now)

	return context,keyslot,now.Unix()
}

// ****************************************************************************

func Season (month string) (string,string) {

	switch month {

	case "December","January","February":
		return "N_Winter","S_Summer"
	case "March","April","May":
		return "N_Spring","S_Autumn"
	case "June","July","August":
		return "N_Summer","S_Winter"
	case "September","October","November":
		return "N_Autumn","S_Spring"
	}

	return "hurricane","typhoon"
}

// ****************************************************************************
 
func GetTimeFromSemantics(speclist []string,now time.Time) time.Time {

	day := 0
	hour := 0
	mins := 0
	weekday := 0
	month := time.Month(0)
	year := 0
	days_to_next := 0

	hasweekday := false
	hasmonth := false

	// Parse semantic time array

	for i,v := range speclist {

		if i == 0 {
			continue
		}

		if strings.HasPrefix(v,"Day") {
			fmt.Sscanf(v[3:],"%d",&day)
			continue
		}

		if strings.HasPrefix(v,"Yr") {
			fmt.Sscanf(v[2:],"%d",&year)
			continue
		}

		if strings.HasPrefix(v,"Min") {
			fmt.Sscanf(v[3:],"%d",&mins)
			continue
		}

		if strings.HasPrefix(v,"Hr") {
			fmt.Sscanf(v[2:],"%d",&hour)
			continue
		}

		if !hasweekday {
			weekday,hasweekday = InList(v,GR_DAY_TEXT)
			if hasweekday {
				intended := weekday
				todayis := fmt.Sprintf("%s",now.Weekday())
				actual,_ := InList(todayis,GR_DAY_TEXT)
				days_to_next = (intended - actual + 7) % 7
				continue
			}
		}

		if !hasmonth {
			var index int
			index,hasmonth = InList(v,GR_MONTH_TEXT)
			if hasmonth {
				month = time.Month(index)
				continue
			}
		}

		fmt.Println("Semantic time parameter without semantic prefix (Day,Hr,Min, etc)",v)
	}

	if hasweekday && (day > 0 || hasmonth || year > 0) {
		fmt.Println("Weekday only makes sense as the next applicable occurrence, without a date")
	} else if hasweekday {

		// We're looking for the next upcoming day
		day = now.Day()
		month = now.Month()
		year = now.Year()
		newnow := time.Date(year,month,day,0,0,0,0,time.UTC)
		newnow = newnow.AddDate(0,0,days_to_next)
		return newnow
	}

	if year == 0 {
		year = now.Year()
	}

	if day == 0 {
		day = now.Day()
	}

	if month == 0 {
		month = now.Month()
	}

	if hour == 0 {
		hour = now.Hour()
	}

	// Note the local timezone is very problematic in Go

	_,offset_secs := now.Zone()

	offset := offset_secs / 3600

	newnow := time.Date(year,month,day,hour-offset,mins,0,0,time.UTC)
	return newnow
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
const N_GRAM_MIN = 2  // fragments that are too small are exponentially large in number and meaningless

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
			ns.Significance = AssessStaticIntent(ngram,L,STM_NGRAM_FREQ,N_GRAM_MIN)
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

func ExtractIntentionalTokens(L int,selected []TextRank,Nmin,Nmax int) ([][]string,[][]string,[]string,[]string) {

	// This function examines a fractionation of text for fractions, only for
	// sentences that are selected, and extracts some shared context

	const policy_skim = 15
	const reuse_threshold = 0
	const intent_threshold = 1

	slow,fast,doc_parts := AssessTextFastSlow(L,STM_NGRAM_LOCA)

	var grad_amb [N_GRAM_MAX]map[string]float64
	var grad_oth [N_GRAM_MAX]map[string]float64

	// returns

	var fastparts = make([][]string,doc_parts)
	var slowparts = make([][]string,doc_parts)
	var fastwhole []string
	var slowwhole []string

	for n := 1; n < Nmax; n++ {
		grad_amb[n] = make(map[string]float64)
		grad_oth[n] = make(map[string]float64)
	}

	for p := 0; p < doc_parts; p++ {

		for n := Nmin; n < Nmax; n++ {

			var amb []string
			var other []string

			for ngram := range fast[n][p] {
				if fast[n][p][ngram] > reuse_threshold {
					other = append(other,ngram)
				}
			}

			for ngram := range slow[n][p] {
				if slow[n][p][ngram] > reuse_threshold {
					amb = append(amb,ngram)
				}
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
			
			for i := 0 ; i < policy_skim && i < len(amb); i++ {
				v := StaticIntentionality(L,amb[i],STM_NGRAM_FREQ[n][amb[i]])
				slowparts[p] = append(slowparts[p],amb[i])
				if v > intent_threshold {
					grad_amb[n][amb[i]] += v
				}
			}
			
			for i := 0 ; i < policy_skim && i < len(other); i++ {
				v := StaticIntentionality(L,other[i],STM_NGRAM_FREQ[n][other[i]])
				fastparts[p] = append(fastparts[p],other[i])
				if v > intent_threshold {
					grad_oth[n][other[i]] += v
				}
			}
		}
	}
	
	// Summary ranking of whole doc, but pick only if selected
	
	for n := Nmin; n < Nmax; n++ {
		
		var amb []string
		var other []string

		for s := range selected {
			for ngram := range grad_amb[n] {
				if !strings.Contains(selected[s].Fragment,ngram) {
					delete(grad_amb[n],ngram)
				}
			}

			for ngram := range grad_oth[n] {
				if !strings.Contains(selected[s].Fragment,ngram) {
					delete(grad_oth[n],ngram)
				}
			}
		}
				
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
		
		for i := 0 ; i < policy_skim && i < len(amb); i++ {
			slowwhole = append(slowwhole,amb[i])
		}

		for i := 0 ; i < policy_skim && i < len(other); i++ {
			fastwhole = append(fastwhole,other[i])
		}
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

		for n := N_GRAM_MIN; n < N_GRAM_MAX; n++ {

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

	// if this doesn't occur at least 3 times, then why do we care?

	const ignore = 2

	if freq < ignore {
		return 0
	}

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
			s+= ","
		}
	}

	return s
}

// **************************************************************************

func SQLEscape(s string) string {

	undo := strings.ReplaceAll(s,"''","'")
	escaped := strings.ReplaceAll(undo,"'","''")

	return string(escaped)
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

func FormatSQLLinkArray(array []Link) string {

	// {"(81,1,2,\"(1,0)\")","(108,1,2,\"(3,11)\")","(118,1,2,\"(2,1348)\")"}

	var s string

	for _,lnk := range array {

		l := fmt.Sprintf("(%d, %f, %d, \\\"(%d,%d)\\\")",lnk.Arr,lnk.Wgt,lnk.Ctx,lnk.Dst.Class,lnk.Dst.CPtr)
		s += fmt.Sprintf("\"%s\",",l)
	}

	s = "{" + strings.Trim(s,",") + "}"

	return s
}

// **************************************************************************

func ParseSQLLinkString(s string) Link {

        // e.g. (77,0.34,334,"(4,2)")

      	var l Link

	s = strings.Replace(s,"\"","",-1)
	s = strings.Replace(s,"\\","",-1)
	s = strings.Replace(s,"(","",-1)
	s = strings.Replace(s,")","",-1)
	
        items := strings.Split(s,",")

	for i := 0; i < len(items); i++ {
		items[i] = strings.Replace(items[i],";","",-1)
		items[i] = strings.TrimSpace(items[i])
	}

	// Arrow type
	fmt.Sscanf(items[0],"%d",&l.Arr)

	// Link weight
	fmt.Sscanf(items[1],"%f",&l.Wgt)

	// Context pointer
	fmt.Sscanf(items[2],"%d",&l.Ctx)

	// DstNPtr
	fmt.Sscanf(items[3],"%d",&l.Dst.Class)
	fmt.Sscanf(items[4],"%d",&l.Dst.CPtr)

	return l
}

//**************************************************************

func ParseLinkArray(s string) []Link {

	var array []Link

	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s,"{","")
	s = strings.ReplaceAll(s,"}","")

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

			// Actual paths need len > 1, but this is also used to seed longer paths

			if len(links) < 1 {
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

func SimilarString(full,like string) bool {

	// Placeholder
	// Need to handle pluralisation patterns etc... multi-language

	if full == like {
		return true
	}

	if full == "" || like == "" || full == "any" || like == "any" {  // same as any
		return true
	}

	if strings.Contains(full,like) {
		return true
	}

	return false
}

//****************************************************************************

func InList(s string, list []string) (int,bool) {

	for i,v := range list {
		if s == v {
			return i,true
		}
	}

	return -1,false
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

func MatchContexts(context1 []string,context2ptr ContextPtr) bool {

	if context1 == nil || context2ptr == 0 {
		return true
	}

	context2 := strings.Split(GetContext(context2ptr),",")

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

func SearchTermLen(names []string) int {

	var maxlen int

	for _,s := range names {
		if !IsNPtrStr(s) && len(s) > maxlen {
			maxlen = len(s)
		}
	}

	return maxlen
}

// **************************************************************************

func IsNPtrStr(s string) bool {

	s = strings.TrimSpace(s)

	if s[0] == '(' && s[len(s)-1] == ')' {
		var a,b int = -1,-1
		fmt.Sscanf(s,"(%d,%d)",&a,&b)
		if a >= 0 && b >= 0 {
			return true
		}
	}
	return false
}

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
		if run[r] == '\n' {
		} else if run[r] == '"' {
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
			if r < len(runes) && runes[r] != '\n' {
				fmt.Print("\n",indent)
				linecounter = 0
				continue
			} else {
				linecounter = 0
			}
		}

		if r < len(runes) {
			fmt.Print(string(runes[r]))
		}
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

func Waiting(output bool,total int) {

	if !output {
		return
	}

	percent := float64(SILLINESS_COUNTER) / float64(total) * 100

	var propaganda = []string{"\n1) JOT IT DOWN WHEN YOU THINK OF IT. . .\n","\n2) TYPE IT INTO N4L AS SOON AS YOU CAN. . .\n","\n3) ORGANIZE AND TIDY YOUR NOTES EVERY DAY. . .\n","\n4) UPLOAD AND BROWSE THEM ONLINE. . .\n","\n5) AND REMEMBER, IT ISN'T KNOWLEDGE IF YOU DON'T ACTUALLY KNOW IT !!\n"}

	const interval = 2

	if SILLINESS {
		if SILLINESS_COUNTER % interval != 0 {
			fmt.Print(" ")
		} else {
			fmt.Print(string(propaganda[SILLINESS_SLOGAN][SILLINESS_POS]))
			SILLINESS_POS++

			if SILLINESS_POS > len(propaganda[SILLINESS_SLOGAN])-1 {
				SILLINESS_POS = 0
				SILLINESS = false
				SILLINESS_SLOGAN++
				if SILLINESS_SLOGAN >= len(propaganda) {
					SILLINESS_SLOGAN = 0
				}
			}
		}
	} else {
		fmt.Print(".")
	}

	if SILLINESS_COUNTER % (2000) == 0 {
		SILLINESS = !SILLINESS
		if percent > 100 {
			fmt.Printf("\n(%.1f%% - oops, have to work overtime!)\n",percent)
		} else {
			fmt.Printf("\n\n(%.1f%%) uploading . . .\n",percent)
		}
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

	return retval,SQLEscape(stripped)
}

//****************************************************************************

func IsExactMatch(org string) (bool,string) {

	org = strings.TrimSpace(org)

	if len(org) == 0 {
		return false,org
	}

	if org[0] == '!' && org[len(org)-1] == '!' {
		tr := strings.Trim(org,"!")
		return true,strings.ToLower(tr)
	}

	if org[0] == '|' && org[len(org)-1] == '|' {
		tr := strings.Trim(org,"|")
		return true,strings.ToLower(tr)
	}

	return false,org
}

//****************************************************************************

func IsStringFragment(s string) bool {

	tsvec_patterns := []string{"|","&","!","<->","<1>","<2>","<3>","<4>"}

	// if this is a ts_vec pattern, it's not for us

	for _,p := range tsvec_patterns {
		if strings.Contains(s,p) {
			return false
		}
	}

	// The tsvector cannot handle spaces or apostrophes(!), so fall back on LIKE %%

	str_patterns := []string{" ","-","_","'","\""}

	for _,p := range str_patterns {
		if strings.Contains(s,p) {
			return true
		}
	}

	const theshold_for_uniqueness = 12 // skjønn

	if len(s) > theshold_for_uniqueness {
		return true
	}

	return false
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

//****************************************************************************
// Allow for simple dynamic content for some approved realtime functions
//****************************************************************************

func ExpandDynamicFunctions(s string) string {

	if !strings.Contains(s,"{") {
		return s
	}

	if !strings.Contains(s,"}") {
		return s
	}

	chars := []rune(s[len("Dynamic:"):])

	var news string

	for pos := 0; pos < len(chars); pos++ {

		if chars[pos] != '{' {
			news += string(chars[pos])
		} else {
			newpos,result := EvaluateInBuilt(chars,pos)
			news += result
			pos = newpos
		}
	}

	return news
}

//****************************************************************************

func EvaluateInBuilt(chars []rune,pos int) (int,string) {

	var fntext string
	var endpos int

	for r := pos; chars[r] != '}' && r < len(chars); r++ {
		fntext += string(chars[r])
		endpos = r+1
	}

	fntext = fntext[1:len(fntext)]

	delim := func(c rune) bool {
		return c == ' ' || c == ',' || c == ';'
	}

	fn := strings.FieldsFunc(fntext,delim)
	result := DoInBuiltFunction(fn)
	return endpos,result
}

//****************************************************************************

func DoInBuiltFunction(fn []string) string {

	// Placeholder - this needs to support sandboxed read only user functions

	var result string

	switch fn[0] {
	case "TimeUntil":
		result = InBuiltTimeUntil(fn)
	case "TimeSince":
		result = InBuiltTimeSince(fn)
	}

	return result
}

//****************************************************************************

func InBuiltTimeUntil(fn []string) string {

	now := time.Now().Local()
	intended_time := GetTimeFromSemantics(fn,now)
	duration := intended_time.Sub(now)

	interval := int(duration / 1000000000)  // nanoseconds -> seconds

	years := interval / (365 * 24 * 3600)
	r1 := interval % (365 * 24 * 3600)

	days := r1 / (24 * 3600)
	r2 := r1 % (24 * 3600)

	hours := r2 / 3600
	r3 :=  r2 % 3600

	mins := r3 / 60

	return ShowTime(years,days,hours,mins)
}

//****************************************************************************

func InBuiltTimeSince(fn []string) string {

	now := time.Now().Local()
	intended_time := GetTimeFromSemantics(fn,now)

	duration := now.Sub(intended_time)

	interval := int(duration / 1000000000)  // nanoseconds -> seconds

	years := interval / (365 * 24 * 3600)
	r1 := interval % (365 * 24 * 3600)

	days := r1 / (24 * 3600)
	r2 := r1 % (24 * 3600)

	hours := r2 / 3600
	r3 :=  r2 % 3600

	mins := r3 / 60

	return ShowTime(years,days,hours,mins)
}

//****************************************************************************

func ShowTime(years,days,hours,mins int) string {

	var s string

	if years > 0 {
		s += fmt.Sprintf("%d Years, ",years)
	}

	if days > 0 {
		s += fmt.Sprintf("%d Days, ",days)
	}

	if hours > 0 {
		s += fmt.Sprintf("%d Hours, ",hours)
	}

	s += fmt.Sprintf("%d Mins ",mins)

	if mins < 0 {
		s += " [already passed or waiting for next occurrence]"
	}

	return s
}


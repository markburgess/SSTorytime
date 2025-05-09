
//
// N4LParser
//

package main

import (
	"strings"
	"os"
	"io/ioutil"
	"flag"
	"fmt"
	"unicode/utf8"
	"unicode"
	"regexp"
	"sort"
	"strconv"


        SST "SSTorytime"
)

//**************************************************************
// Parsing state variables
//**************************************************************

const (
	ALPHATEXT = 'x'
	NON_ASCII_LQUOTE = '“'
	NON_ASCII_RQUOTE = '”'
        HAVE_PLUS = 11
        HAVE_MINUS = 22
	ROLE_ABBR = 33

	ROLE_EVENT = 1
	ROLE_RELATION = 2
	ROLE_SECTION = 3
	ROLE_CONTEXT = 4
	ROLE_CONTEXT_ADD = 5
	ROLE_CONTEXT_SUBTRACT = 6
	ROLE_BLANK_LINE = 7
	ROLE_LINE_ALIAS = 8
	ROLE_LOOKUP = 9

	WORD_MISTAKE_LEN = 3 // a string shorter than this is probably a mistake

	WARN_NOTE_TO_SELF = "WARNING: Found a note to self in the text"
	WARN_INADVISABLE_CONTEXT_EXPRESSION = "WARNING: Inadvisably complex/parenthetic context expression - simplify?"
	WARN_CHAPTER_CLASS_MIXUP="WARNING: possible space between class cancellation -:: <class> :: ambiguous chapter name, in: "

	ERR_NO_SUCH_FILE_FOUND = "No file found in the name "
	ERR_MISSING_EVENT = "Missing item? Dangling section, relation, or context"
	ERR_MISSING_SECTION = "Declarations outside a section or chapter"
	ERR_NO_SUCH_ALIAS = "No such alias or \" reference exists to fill in - aborting"
	ERR_MISSING_ITEM_SOMEWHERE = "Missing item somewhere"
	ERR_MISSING_ITEM_RELN = "Missing item or double relation"
	ERR_MISMATCH_QUOTE = "Apparent missing or mismatch in ', \" or ( )"
	ERR_ILLEGAL_CONFIGURATION = "Error in configuration, no such section"
	ERR_BAD_LABEL_OR_REF = "Badly formed label or reference (@label becomes $label.n) in "
	ERR_ILLEGAL_QUOTED_STRING_OR_REF = "WARNING: Something wrong, bad quoted string or mistaken back reference. Close any space after a quote..."
	ERR_ANNOTATION_BAD = "Annotation marker should be short mark of non-space, non-alphanumeric character "
	ERR_BAD_ABBRV = "abbreviation out of place"
	ERR_BAD_ALIAS_REFERENCE = "Alias references start from $name.1"
	ERR_ANNOTATION_MISSING = "Missing non-alphnumeric annotation marker or stray relation"
	ERR_ANNOTATION_REDEFINE = "Redefinition of annotation character"
	ERR_SIMILAR_NO_SIGN = "Arrows for similarity do not have signs, they are directionless"
	ERR_ARROW_SELFLOOP = "Arrow's origin points to itself"
	ERR_ARR_REDEFINITION="Redefinition of arrow "
	ERR_NEGATIVE_WEIGHT = "Arrow relation has a negative weight, which is disallowed. Use a NOT relation if you want to signify inhibition: "
	ERR_TOO_MANY_WEIGHTS = "More than one weight value in the arrow relation "
        ERR_STRAY_PAREN="Stray ) in an event/item - illegal character"
	ERR_MISSING_LINE_LABEL_IN_REFERENCE="Missing a line label in reference, should be in the form $label.n"
	ERR_NON_WORD_WHITE="Non word (whitespace) character after an annotation: "
	ERR_SHORT_WORD="Short word, probably a mistake: "
	ERR_ILLEGAL_ANNOT_CHAR="Cannot use +/- reserved tokens for annotation"
)

//**************************************************************

var ( 
	LINE_NUM int = 1
	LINE_ITEM_CACHE = make(map[string][]string)  // contains current and labelled line elements
	LINE_ITEM_REFS []SST.NodePtr                     // contains current line integer references
	LINE_RELN_CACHE = make(map[string][]SST.Link)
	LINE_ITEM_STATE int = ROLE_BLANK_LINE
	LINE_ALIAS string = ""
	LINE_ITEM_COUNTER int = 1
	LINE_RELN_COUNTER int = 0
	LINE_PATH []SST.Link

	FWD_ARROW string
	BWD_ARROW string
	FWD_INDEX SST.ArrowPtr
	BWD_INDEX SST.ArrowPtr
	ANNOTATION = make(map[string]string)

	CONTEXT_STATE = make(map[string]bool)
	SECTION_STATE string

	SEQUENCE_MODE bool = false
	SEQUENCE_RELN string = "then" 
	LAST_IN_SEQUENCE string = ""

	// Flags

	VERBOSE bool = false
	DIAGNOSTIC bool = false
	UPLOAD bool = false
	SUMMARIZE bool = false
	CREATE_ADJACENCY bool = false
	ADJ_LIST string

	CURRENT_FILE string
	TEST_DIAG_FILE string

	RELN_BY_SST [4][]SST.ArrowPtr // From an EventItemNode
)

//**************************************************************
// DATA structures for input
//**************************************************************

type RCtype struct {
	Row SST.NodePtr
	Col SST.NodePtr
}

//**************************************************************
// BEGIN
//**************************************************************

func main() {

	args := Init()

	NewFile("N4Lconfig.in")
	config := ReadFile(CURRENT_FILE)

	AddMandatory()
	ParseConfig(config)

	for input := 0; input < len(args); input++ {
		NewFile(args[input])
		input := ReadFile(CURRENT_FILE)
		ParseN4L(input)
	}

	if SUMMARIZE {
		SummarizeGraph()
	}

	if CREATE_ADJACENCY {
		dim, key, d_adj, u_adj := CreateAdjacencyMatrix(ADJ_LIST)
		PrintMatrix("directed adjacency sub-matrix",dim,key,d_adj)
		PrintMatrix("undirected adjacency sub-matrix",dim,key,u_adj)
		evc := ComputeEVC(dim,u_adj)
		PrintNZVector("Eigenvector centrality (EVC) score for symmetrized graph",dim,key,evc)
	}

	if UPLOAD {
		load_arrows := false
		ctx := SST.Open(load_arrows)
		fmt.Println("Uploading nodes..")
		SST.GraphToDB(ctx)
		SST.Close(ctx)
	}
}

//**************************************************************

func Init() []string {

	flag.Usage = Usage
	verbosePtr := flag.Bool("v", false,"verbose")
	diagPtr := flag.Bool("d", false,"diagnostic mode")
	uploadPtr := flag.Bool("u", false,"upload")
	wipePtr := flag.Bool("wipe", false,"wipe and reset")
	incidencePtr := flag.Bool("s", false,"summary (node,links...)")
	adjacencyPtr := flag.String("adj", "none", "a quoted, comma-separated list of short link names")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		Usage()
		os.Exit(1);
	}

	if *verbosePtr {
		VERBOSE = true
	}

	if *wipePtr {
		SST.WIPE_DB = true
	}

	if *diagPtr {
		VERBOSE = true
		DIAGNOSTIC = true
	}

	if *uploadPtr {
		UPLOAD = true
	}
	if *incidencePtr {
		SUMMARIZE = true
	}

	if *adjacencyPtr != "none" {
		CREATE_ADJACENCY = true
		ADJ_LIST = *adjacencyPtr
	}

	SST.MemoryInit()

	return args
}

//**************************************************************

func NewFile(filename string) {

	CURRENT_FILE = filename
	TEST_DIAG_FILE = DiagnosticName(filename)

	Box("Parsing new file",filename)

	LINE_ITEM_STATE = ROLE_BLANK_LINE
	LINE_NUM = 1
	LINE_ITEM_CACHE = make(map[string][]string)
	LINE_RELN_CACHE = make(map[string][]SST.Link)
	LINE_ITEM_REFS = nil
	LINE_ITEM_COUNTER = 1
	LINE_RELN_COUNTER = 0
	LINE_ALIAS = ""
	LAST_IN_SEQUENCE = ""
	FWD_ARROW = ""
	BWD_ARROW = ""
	SECTION_STATE = ""
	CONTEXT_STATE = make(map[string]bool)
}

//**************************************************************
// N4L configuration
//**************************************************************

func ParseConfig(src []rune) {

	var token string

	for pos := 0; pos < len(src); {

		pos = SkipWhiteSpace(src,pos)
		token,pos = GetConfigToken(src,pos)

		ClassifyConfigRole(token)
	}
}

//**************************************************************

func GetConfigToken(src []rune, pos int) (string,int) {

	// Handle concatenation of words/lines and separation of types

	var token string

	if pos >= len(src) {
		return "", pos
	}

	switch (src[pos]) {

	case '+':
		token,pos = ReadToLast(src,pos,ALPHATEXT)

	case '-':
		token,pos = ReadToLast(src,pos,ALPHATEXT)

	case '(':
		token,pos = ReadToLast(src,pos,')')  // alias

	case '#':
		return "",pos

	case '/':
		if src[pos+1] == '/' {
			return "",pos
		}

	default: // similarity
		token,pos = ReadToLast(src,pos,ALPHATEXT)

	}

	return token, pos
}

//**************************************************************

func ClassifyConfigRole(token string) {

	if len(token) == 0 {
		return
	}

	if token[0] == '-' && LINE_ITEM_STATE == ROLE_BLANK_LINE {
		SECTION_STATE = strings.TrimSpace(token[1:])
		Box("Configuration of",SECTION_STATE)
		LINE_ITEM_STATE = ROLE_SECTION
		return
	}

	switch SECTION_STATE {

	case "leadsto","contains","properties":

		switch token[0] {

		case '+':
			FWD_ARROW = strings.TrimSpace(token[1:])
			LINE_ITEM_STATE = HAVE_PLUS
			Diag("fwd arrow in",SECTION_STATE, token)
			
		case '-':
			BWD_ARROW = strings.TrimSpace(token[1:])
			LINE_ITEM_STATE = HAVE_MINUS
			Diag("bwd arrow in",SECTION_STATE, token)

		case '(':
			reln := token[1:len(token)-1]
			reln = strings.TrimSpace(reln)

			if LINE_ITEM_STATE == HAVE_MINUS {
				CheckArrow(reln,BWD_ARROW)
				BWD_INDEX = SST.InsertArrowDirectory(SECTION_STATE,reln,BWD_ARROW,"-")
				SST.InsertInverseArrowDirectory(FWD_INDEX,BWD_INDEX)
				PVerbose("In",SECTION_STATE,"short name",reln,"for",BWD_ARROW,", direction","-")
			} else if LINE_ITEM_STATE == HAVE_PLUS {
				CheckArrow(reln,FWD_ARROW)
				FWD_INDEX = SST.InsertArrowDirectory(SECTION_STATE,reln,FWD_ARROW,"+")
				PVerbose("In",SECTION_STATE,"short name",reln,"for",FWD_ARROW,", direction","+")
			} else {
				ParseError(ERR_BAD_ABBRV)
				os.Exit(-1)
			}
		}

	case "similarity":

		switch token[0] {

		case '(':
			reln := token[1:len(token)-1]
			reln = strings.TrimSpace(reln)

			if LINE_ITEM_STATE == HAVE_MINUS {
				index := SST.InsertArrowDirectory(SECTION_STATE,reln,BWD_ARROW,"both")
				SST.InsertInverseArrowDirectory(index,index)
				PVerbose("In",SECTION_STATE,reln,"for",BWD_ARROW,", direction","both")
			} else {
				PVerbose(SECTION_STATE,"abbreviation out of place")
			}

		case '+','-':
			ParseError(ERR_SIMILAR_NO_SIGN)
			os.Exit(-1)

		default:
			similarity := strings.TrimSpace(token)
			FWD_ARROW = similarity
			BWD_ARROW = similarity
			LINE_ITEM_STATE = HAVE_MINUS
		}

	case "annotations":

		switch token[0] {

		case '(':
			if LINE_ITEM_STATE != HAVE_PLUS {
				ParseError(ERR_ANNOTATION_MISSING)
			}

			FWD_ARROW = StripParen(token)
			PVerbose("Annotation marker",LAST_IN_SEQUENCE,"defined as arrow:",FWD_ARROW)

			value,defined := ANNOTATION[LAST_IN_SEQUENCE]

			if defined && value != FWD_ARROW {
				ParseError(ERR_ANNOTATION_REDEFINE)
				os.Exit(-1)
			}

			ANNOTATION[LAST_IN_SEQUENCE] = FWD_ARROW
			LINE_ITEM_STATE = ROLE_BLANK_LINE
			
		default:

			for r := range token {
				if unicode.IsLetter(rune(token[r])) {
					ParseError(ERR_ANNOTATION_BAD)
				}
			}

			if token[0] == '+' || token[0] == '-' {
				ParseError(ERR_ILLEGAL_ANNOT_CHAR)
				os.Exit(-1)
			}

			Diag("Markup character defined in",SECTION_STATE, token)
			LINE_ITEM_STATE = HAVE_PLUS
			LAST_IN_SEQUENCE = token

		}

	default:
		ParseError(ERR_ILLEGAL_CONFIGURATION+" "+SECTION_STATE)
		os.Exit(-1)
	}
}

//**************************************************************

func CheckArrow(alias,name string) {

	prev,ok := SST.ARROW_SHORT_DIR[alias]
	if ok {
		ParseError(ERR_ARR_REDEFINITION+"\""+alias+"\" previous short name: "+SST.ARROW_DIRECTORY[prev].Short)
		os.Exit(-1)
	}
	
	prev,ok = SST.ARROW_LONG_DIR[name]
	if ok {
		ParseError(ERR_ARR_REDEFINITION+"\""+name+"\" previous long name: "+SST.ARROW_DIRECTORY[prev].Long)
		os.Exit(-1)
	}
}

//**************************************************************

func GetLinkArrowByName(token string) SST.Link {

	// Return a preregistered link/arrow ptr bythe name of a link

	var reln []string
	var weight float64 = 1
	var weightcount int
	var ctx []string
	var name string

	if token[0] == '(' {
		name = token[1:len(token)-1]
	} else {
		name = token
	}

	name = strings.TrimSpace(name)

	if strings.Contains(name,",") {
		reln = strings.Split(name,",")
		name = reln[0]

		// look at any comma separated notes after the arrow name
		for i := 1; i < len(reln); i++ {

			v, err := strconv.ParseFloat(reln[i], 64)

			if err == nil {
				if weight < 0 {
					ParseError(ERR_NEGATIVE_WEIGHT+token)
					os.Exit(-1)
				}
				if weightcount > 1 {
					ParseError(ERR_TOO_MANY_WEIGHTS+token)
					os.Exit(-1)
				}
				weight = v
				weightcount++
			} else {
				ctx = append(ctx,reln[i])
			}
		}
	}

	// First check if this is an alias/short name

	ptr, ok := SST.ARROW_SHORT_DIR[name]
	
	// If not, then check longname
	
	if !ok {
		ptr, ok = SST.ARROW_LONG_DIR[name]
		
		if !ok {
			ParseError(SST.ERR_NO_SUCH_ARROW+"("+name+")")
			os.Exit(-1)
		}
	}

	var link SST.Link

	link.Arr = ptr
	link.Wgt = weight
	link.Ctx = GetContext(ctx)
	return link
}

//**************************************************************

func LookupAlias(alias string, counter int) string {
	
	value,ok := LINE_ITEM_CACHE[alias]

	if !ok || counter > len(value) {
		ParseError(ERR_NO_SUCH_ALIAS)
		os.Exit(1)
	}

	return LINE_ITEM_CACHE[alias][counter-1]

}

//**************************************************************

func ResolveAliasedItem(token string) string {

	// split $alias.n into (alias string,n int)

	var contig string
	fmt.Sscanf(token,"%s",&contig)
	
	if len(contig) == 1 {
		return token
	}

	if contig == "$$" {
		return token
	}

	split := strings.Split(token[1:],".")

	if len(split) < 2 {
		ParseError(ERR_MISSING_LINE_LABEL_IN_REFERENCE)
		os.Exit(-1)
	}

	name := strings.TrimSpace(split[0])

	var number int = 0
	fmt.Sscanf(split[1],"%d",&number)

	if number < 1 {
		ParseError(ERR_BAD_ALIAS_REFERENCE)
		os.Exit(-1)
	}

	return LookupAlias(name,number)
}

//**************************************************************

func SummarizeAndTestConfig() {

	Box("Raw Summary")
	fmt.Println("..\n")
	fmt.Println("ANNOTATION MARKS", ANNOTATION)
	fmt.Println("..\n")
	fmt.Println("DIRECTORY", SST.ARROW_DIRECTORY)
	fmt.Println("..\n")
	fmt.Println("SHORT",SST.ARROW_SHORT_DIR)
	fmt.Println("..\n")
	fmt.Println("LONG",SST.ARROW_LONG_DIR)
	fmt.Println("\nTEXT\n\n",SST.NODE_DIRECTORY)
}

//**************************************************************

func SummarizeGraph() {

	Box("SUMMARIZE GRAPH.....\n")

	var count_nodes int = 0
	var count_links [4]int
	var total int

	for class := SST.N1GRAM; class <= SST.GT1024; class++ {
		switch class {
		case SST.N1GRAM:
			for n := range SST.NODE_DIRECTORY.N1directory {
				org := SST.NODE_DIRECTORY.N1directory[n]
				count_nodes++
				PrintNodeSystem(n,org,&count_links)
			}
		case SST.N2GRAM:
			for n := range SST.NODE_DIRECTORY.N2directory {
				org := SST.NODE_DIRECTORY.N2directory[n]
				count_nodes++
				PrintNodeSystem(n,org,&count_links)
			}
		case SST.N3GRAM:
			for n := range SST.NODE_DIRECTORY.N3directory {
				org := SST.NODE_DIRECTORY.N3directory[n]
				count_nodes++
				PrintNodeSystem(n,org,&count_links)
			}
		case SST.LT128:
			for n := range SST.NODE_DIRECTORY.LT128 {
				org := SST.NODE_DIRECTORY.LT128[n]
				count_nodes++
				PrintNodeSystem(n,org,&count_links)
			}
		case SST.LT1024:
			for n := range SST.NODE_DIRECTORY.LT1024 {
				org := SST.NODE_DIRECTORY.LT1024[n]
				count_nodes++
				PrintNodeSystem(n,org,&count_links)
			}
		case SST.GT1024:
			for n := range SST.NODE_DIRECTORY.GT1024 {
				org := SST.NODE_DIRECTORY.GT1024[n]
				count_nodes++
				PrintNodeSystem(n,org,&count_links)
			}
		}
	}
		
	fmt.Println("-------------------------------------")
	fmt.Println("Incidence summary of raw declarations")
	fmt.Println("-------------------------------------")

	fmt.Println("Total nodes",count_nodes)

	for st := 0; st < 4; st++ {
		total += count_links[st]
		fmt.Println("Total directed links of type",SST.STTypeName(st),count_links[st])
	}

	complete := count_nodes * (count_nodes-1)
	fmt.Println("Total links",total,"sparseness (fraction of completeness)",float64(total)/float64(complete))
}

//**************************************************************

func CreateAdjacencyMatrix(searchlist string) (int,[]SST.NodePtr,[][]float64,[][]float64) {

	search_list := ValidateLinkArgs(searchlist)

	// the matrix is dim x dim

	filtered_node_list,path_weights := AssembleInvolvedNodes(search_list)

	dim := len(filtered_node_list)

	for f := 0; f < len(filtered_node_list); f++ {
		Verbose("    - row/col key [",f,"/",dim,"]",SST.GetNodeTxtFromPtr(filtered_node_list[f]))
	}

	// Debugging mainly
	//for f := range path_weights {
	//	Verbose("    - path weight",path_weights[f],"from",GetNodeTxtFromPtr(f.Row),"to",GetNodeTxtFromPtr(f.Col))
	//}

	var subadj_matrix [][]float64 = make([][]float64,dim)
	var symadj_matrix [][]float64 = make([][]float64,dim)

	for row := 0; row < dim; row++ {
		subadj_matrix [row] = make([]float64,dim)
		symadj_matrix [row] = make([]float64,dim)
	}

	for row := 0; row < dim; row++ {
		for col := 0; col < dim; col++ {

			var rc, rcT RCtype
			rc.Row = filtered_node_list[row]
			rc.Col = filtered_node_list[col]

			rcT.Row = filtered_node_list[col]
			rcT.Col = filtered_node_list[row]

			subadj_matrix[row][col] = path_weights[rc]

			symadj_matrix[row][col] = path_weights[rc] + path_weights[rcT]
			symadj_matrix[col][row] = path_weights[rc] + path_weights[rcT]
		}
	}

	return dim, filtered_node_list, subadj_matrix, symadj_matrix
}

//**************************************************************

func PrintMatrix(name string, dim int, key []SST.NodePtr, matrix [][]float64) {


	s := fmt.Sprintln("\n",name,"...\n")
	Verbose(s)

	for row := 0; row < dim; row++ {
		
		s = fmt.Sprintf("%20.15s ..\r\t\t\t(",SST.GetNodeTxtFromPtr(key[row]))
		
		for col := 0; col < dim; col++ {
			
			const screenwidth = 12
			
			if col > screenwidth {
				s += fmt.Sprint("\t...")
				break
			} else {
				s += fmt.Sprintf("  %4.1f",matrix[row][col])
			}
			
		}
		s += fmt.Sprint(")")
		Verbose(s)
	}
}

//**************************************************************

func PrintNZVector(name string, dim int, key []SST.NodePtr, vector[]float64) {

	s := fmt.Sprintln("\n",name,"...\n")

	Verbose(s)

	type KV struct {
		Key string
		Value float64
	}

	var vec []KV = make([]KV,dim)

	for row := 0; row < dim; row++ {
		vec[row].Key = SST.GetNodeTxtFromPtr(key[row])
		vec[row].Value = vector[row]
	}

	sort.SliceStable(vec, func(i, j int) bool {
		return vec[i].Value > vec[j].Value
	})

	for row := 0; row < dim; row++ {
		if vec[row].Value > 0.1 {
			s = fmt.Sprintf("ordered by EVC:  (%4.1f)  ",vec[row].Value)
			s += fmt.Sprintf("%-80.79s",vec[row].Key)
			Verbose(s)
		}
	}
}

//**************************************************************

func ComputeEVC(dim int,adj [][]float64) []float64 {

	v := MakeInitVector(dim,1.0)
	vlast := v

	const several = 6

	for i := 0; i < several; i++ {

		v = MatrixOpVector(dim,adj,vlast)

		if CompareVec(v,vlast) < 0.1 {
			break
		}
		vlast = v
	}

	maxval := GetVecMax(v)
	v = NormalizeVec(v,maxval)

	return v
}

//**************************************************************

func MakeInitVector(dim int, init_value float64) []float64 {

	var v = make([]float64,dim)

	for r := 0; r < dim; r++ {
		v[r] = init_value
	}

	return v
}

//**************************************************************

func MatrixOpVector(dim int,m [][]float64, v []float64) []float64 {

	var vp = make([]float64,dim)

	for r := 0; r < dim; r++ {
		for c := 0; c < dim; c++ {
			if m[r][c] != 0 {
				vp[r] += m[r][c] * v[c]
			}
		}
	}
	return vp
}

//**************************************************************

func GetVecMax(v []float64) float64 {

	var max float64 = -1

	for r := range v {
		if v[r] > max {
			max = v[r]
		}
	}

	return max
}

//**************************************************************

func NormalizeVec(v []float64, div float64) []float64 {

	for r := range v {
		v[r] = v[r] / div
	}

	return v
}

//**************************************************************

func CompareVec(v1,v2 []float64) float64 {

	var max float64 = -1

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

func FlatSTType(i int) int {

	n := i - SST.ST_ZERO
	if n < 0 {
		n = -n
	}

	return n
}

//**************************************************************

func ValidateLinkArgs(s string) []SST.ArrowPtr {

	list := strings.Split(s,",")
	var search_list []SST.ArrowPtr

	if s == "" || s == "all" {
		return nil
	}

	for i := range list {
		v,ok := SST.ARROW_SHORT_DIR[list[i]]

		if ok {
			typ := SST.ARROW_DIRECTORY[v].STAindex - SST.ST_ZERO
			if typ < 0 {
				typ = -typ
			}

			name := SST.ARROW_DIRECTORY[v].Long
			ptr := SST.ARROW_DIRECTORY[v].Ptr

			fmt.Println(" - including search pathway STtype",SST.STTypeName(typ),"->",name)
			search_list = append(search_list,ptr)

			if typ != SST.NEAR {
				inverse := SST.INVERSE_ARROWS[ptr]
				fmt.Println("   including inverse meaning",SST.ARROW_DIRECTORY[inverse].Long)
				search_list = append(search_list,inverse)
			}
		} else {
			fmt.Println("\nThere is no link abbreviation called ",list[i])
			os.Exit(-1)
		}
	}

	return search_list
}

//**************************************************************

func AssembleInvolvedNodes(search_list []SST.ArrowPtr) ([]SST.NodePtr,map[RCtype]float64) {

	var node_list []SST.NodePtr
	var weights = make(map[RCtype]float64)

	for class := SST.N1GRAM; class <= SST.GT1024; class++ {

		switch class {
		case SST.N1GRAM:
			for n := range SST.NODE_DIRECTORY.N1directory {
				node_list = SearchIncidentRowClass(SST.NODE_DIRECTORY.N1directory[n],search_list,node_list,weights)
			}
		case SST.N2GRAM:
			for n := range SST.NODE_DIRECTORY.N2directory {
				node_list = SearchIncidentRowClass(SST.NODE_DIRECTORY.N2directory[n],search_list,node_list,weights)
			}
		case SST.N3GRAM:
			for n := range SST.NODE_DIRECTORY.N3directory {
				node_list = SearchIncidentRowClass(SST.NODE_DIRECTORY.N3directory[n],search_list,node_list,weights)
			}
		case SST.LT128:
			for n := range SST.NODE_DIRECTORY.LT128 {
				node_list = SearchIncidentRowClass(SST.NODE_DIRECTORY.LT128[n],search_list,node_list,weights)
			}
		case SST.LT1024:
			for n := range SST.NODE_DIRECTORY.LT1024 {
				node_list = SearchIncidentRowClass(SST.NODE_DIRECTORY.LT1024[n],search_list,node_list,weights)
			}
		case SST.GT1024:
			for n := range SST.NODE_DIRECTORY.GT1024 {
				node_list = SearchIncidentRowClass(SST.NODE_DIRECTORY.GT1024[n],search_list,node_list,weights)
			}
		}
	}

	return node_list,weights
}

//**************************************************************

func SearchIncidentRowClass(node SST.Node, searcharrows []SST.ArrowPtr,node_list []SST.NodePtr,ret_weights map[RCtype]float64) []SST.NodePtr {

	var row_nodes = make(map[SST.NodePtr]bool)
	var ret_nodes []SST.NodePtr

        var rc,cr RCtype

	rc.Row = node.NPtr // transposes
        cr.Col = node.NPtr

	// flip backward facing arrows
	const inverse_flip_arrow = SST.ST_ZERO

        // Only sum over outgoing (+) links
	
	for sttype := SST.ST_ZERO; sttype < len(node.I); sttype++ {
		
		for lnk := range node.I[sttype] {
			arrowptr := node.I[sttype][lnk].Arr
			
			if len(searcharrows) == 0 {
				match := node.I[sttype][lnk]
				row_nodes[match.Dst] = true
				rc.Col = match.Dst
				cr.Row = match.Dst

				if sttype < inverse_flip_arrow {
					ret_weights[cr] += match.Wgt  // flip arrow
				} else {
					ret_weights[rc] += match.Wgt
				}
			} else {
				for l := range searcharrows {
					if arrowptr == searcharrows[l] {
						match := node.I[sttype][lnk]
						row_nodes[match.Dst] = true
						rc.Col = match.Dst
						cr.Row = match.Dst
						if sttype < inverse_flip_arrow {
							ret_weights[cr] += match.Wgt  // flip arrow
						} else {
							ret_weights[rc] += match.Wgt
						}
						
					}
				}
			}
		}
	}
	
	if len(row_nodes) > 0 {
		row_nodes[node.NPtr] = true // Add the parent if it has children
	}

	for nptr := range node_list {
		row_nodes[node_list[nptr]] = true
	}
	
	// Merge idempotently
	
	for nptr := range row_nodes {
		ret_nodes = append(ret_nodes,nptr)
	}

	return ret_nodes
}

//**************************************************************
// N4L language
//**************************************************************

func ParseN4L(src []rune) {

	var token string

	for pos := 0; pos < len(src); {

		pos = SkipWhiteSpace(src,pos)
		token,pos = GetToken(src,pos)

		ClassifyTokenRole(token)

	}

	if Dangler() {
		ParseError(ERR_MISSING_EVENT)
	}
}

//**************************************************************

func SkipWhiteSpace(src []rune, pos int) int {

	for ; pos < len(src) && IsWhiteSpace(src[pos],src[pos]); pos++ {

		if src[pos] == '\n' {
			UpdateLastLineCache() 
		} else {
		
			if src[pos] == '#' || (src[pos] == '/' && src[pos+1] == '/') {
				
				for ; pos < len(src) && src[pos] != '\n'; pos++ {
				}

				UpdateLastLineCache() 
			}
		}
	}

	return pos
}

//**************************************************************

func AddMandatory() {

	arr := SST.InsertArrowDirectory("leadsto","empty","debug","+")
	inv := SST.InsertArrowDirectory("leadsto","void","unbug","-")
	SST.InsertInverseArrowDirectory(arr,inv)

	arr = SST.InsertArrowDirectory("leadsto",SEQUENCE_RELN,SEQUENCE_RELN,"+")
	inv = SST.InsertArrowDirectory("leadsto","prev","follows on from","-")
	SST.InsertInverseArrowDirectory(arr,inv)

	arr = SST.InsertArrowDirectory("properties","url","has URL","+")
        inv = SST.InsertArrowDirectory("properties","isurl","is a URL for","-")
	SST.InsertInverseArrowDirectory(arr,inv)

	arr = SST.InsertArrowDirectory("properties","img","has image","+")
        inv = SST.InsertArrowDirectory("properties","isimg","is an image for","-")
	SST.InsertInverseArrowDirectory(arr,inv)
}

//**************************************************************

func GetToken(src []rune, pos int) (string,int) {

	// Handle concatenation of words/lines and separation of types

	var token string

	if pos >= len(src) {
		return "", pos
	}

	switch (src[pos]) {

	case '+':  // could be +:: 

		switch (src[pos+1]) {

		case ':':
			token,pos = ReadToLast(src,pos,':')
		default:
			token,pos = ReadToLast(src,pos,ALPHATEXT)
		}

	case '-':  // could -:: or -section

		switch (src[pos+1]) {

		case ':':
			token,pos = ReadToLast(src,pos,':')
		default:
			token,pos = ReadToLast(src,pos,ALPHATEXT)
		}

	case ':':
		token,pos = ReadToLast(src,pos,':')

	case '(':
		token,pos = ReadToLast(src,pos,')')

        case '"','\'':
		quote := src[pos]

		if IsQuote(quote) && IsBackReference(src,pos) {
			token = "\""
			pos++
		} else {
			if pos+2 < len(src) && IsWhiteSpace(src[pos+1],src[pos+2]) {
				ParseError(ERR_ILLEGAL_QUOTED_STRING_OR_REF)
				os.Exit(-1)
			}
			token,pos = ReadToLast(src,pos,quote)
			strip := strings.Split(token,string(quote))
			token = strip[1]
		}

	case '#':
		return "",pos

	case '/':
		if src[pos+1] == '/' {
			return "",pos
		}

	case '@':
		token,pos = ReadToLast(src,pos,' ')

	default: // a text item that could end with any of the above
		token,pos = ReadToLast(src,pos,ALPHATEXT)

	}

	return token, pos
}

//**************************************************************

func ClassifyTokenRole(token string) {

	if len(token) == 0 {
		return
	}

	switch token[0] {

	case ':':
		expression := ExtractContextExpression(token)
		CheckSequenceMode(expression,'+')
		LINE_ITEM_STATE = ROLE_CONTEXT
		AssessGrammarCompletions(expression,LINE_ITEM_STATE)

	case '+':
		expression := ExtractContextExpression(token)
		CheckSequenceMode(expression,'+')
		LINE_ITEM_STATE = ROLE_CONTEXT_ADD
		AssessGrammarCompletions(expression,LINE_ITEM_STATE)

	case '-':
		if token[1:2] == string(':') {
			expression := ExtractContextExpression(token)
			CheckSequenceMode(expression,'-')
			LINE_ITEM_STATE = ROLE_CONTEXT_SUBTRACT
			AssessGrammarCompletions(expression,LINE_ITEM_STATE)
		} else {
			section := strings.TrimSpace(token[1:])
			LINE_ITEM_STATE = ROLE_SECTION
			AssessGrammarCompletions(section,LINE_ITEM_STATE)
		}

		// No quotes here in a string, we need to allow quoting in excerpts.

	case '(':
		if LINE_ITEM_STATE == ROLE_RELATION {
			ParseError(ERR_MISSING_ITEM_RELN)
			os.Exit(-1)
		}
		link := GetLinkArrowByName(token)
		LINE_ITEM_STATE = ROLE_RELATION
		LINE_RELN_CACHE["THIS"] = append(LINE_RELN_CACHE["THIS"],link)
		LINE_RELN_COUNTER++

	case '"': // prior reference
		result := LookupAlias("PREV",LINE_ITEM_COUNTER)
		LINE_ITEM_CACHE["THIS"] = append(LINE_ITEM_CACHE["THIS"],result)
		StoreAlias(result)
		AssessGrammarCompletions(result,LINE_ITEM_STATE)
		LINE_ITEM_STATE = ROLE_EVENT
		LINE_ITEM_COUNTER++

	case '@':
		LINE_ITEM_STATE = ROLE_LINE_ALIAS
		token  = strings.TrimSpace(token)
		LINE_ALIAS = token[1:]
		CheckLineAlias(token)

	case '$':
		CheckLineAlias(token)
		actual := ResolveAliasedItem(token)
		LINE_ITEM_CACHE["THIS"] = append(LINE_ITEM_CACHE["THIS"],actual)
		PVerbose("fyi, line reference",token,"resolved to",actual)
		AssessGrammarCompletions(actual,LINE_ITEM_STATE)
		LINE_ITEM_STATE = ROLE_LOOKUP
		LINE_ITEM_COUNTER++

	default:
		LINE_ITEM_CACHE["THIS"] = append(LINE_ITEM_CACHE["THIS"],token)
		StoreAlias(token)
		AssessGrammarCompletions(token,LINE_ITEM_STATE)

		LINE_ITEM_STATE = ROLE_EVENT
		LINE_ITEM_COUNTER++
	}
}

//**************************************************************

func AssessGrammarCompletions(token string, prior_state int) {

	if len(token) == 0 {
		return
	}

	this_item := token

	switch prior_state {

	case ROLE_RELATION:

		CheckNonNegative(LINE_ITEM_COUNTER-2)
		last_item := LINE_ITEM_CACHE["THIS"][LINE_ITEM_COUNTER-2]
		last_reln := LINE_RELN_CACHE["THIS"][LINE_RELN_COUNTER-1]
		last_iptr := LINE_ITEM_REFS[LINE_ITEM_COUNTER-2]
		this_iptr := HandleNode(this_item)
		IdempAddLink(last_item,last_iptr,last_reln,this_item,this_iptr)
		CheckSection()

	case ROLE_CONTEXT:
		Box("Reset context: ->",this_item)
		ContextEval(this_item,"=")
		CheckSection()

	case ROLE_CONTEXT_ADD:
		PVerbose("Add to context:",this_item)
		ContextEval(this_item,"+")
		CheckSection()

	case ROLE_CONTEXT_SUBTRACT:
		PVerbose("Remove from context:",this_item)
		ContextEval(this_item,"-")
		CheckSection()

	case ROLE_SECTION:
		Box("Set chapter/section: ->",this_item)
		CheckChapter(this_item)
		SECTION_STATE = this_item

	default:
		CheckSection()

		if AllCaps(token) {
			ParseError(WARN_NOTE_TO_SELF+" ("+token+")")
		}

		HandleNode(this_item)
		LinkUpStorySequence(this_item)
	}
}

//**************************************************************

func CheckLineAlias(token string) {

	var contig string
	fmt.Sscanf(token,"%s",&contig)
	
	if token[0] == '@' && len(contig) == 1 {
		ParseError(ERR_BAD_LABEL_OR_REF+token)
		os.Exit(-1)
	}
}

//**************************************************************

func CheckChapter(name string) {

	if name[0] == ':' {
		ParseError(WARN_CHAPTER_CLASS_MIXUP+name)
		os.Exit(-1)
	}

	SEQUENCE_MODE = false
}

//**************************************************************

func StoreAlias(name string) {

	if LINE_ALIAS != "" {
		PVerbose("-- Storing alias",LINE_ITEM_CACHE[LINE_ALIAS],name,"as",LINE_ALIAS)
		LINE_ITEM_CACHE[LINE_ALIAS] = append(LINE_ITEM_CACHE[LINE_ALIAS],name)
	}
}


//**************************************************************
// Memory representation
//**************************************************************

func IdempAddLink(from string, frptr SST.NodePtr, link SST.Link,to string, toptr SST.NodePtr) {

	// Add a link index cache pointer directly to a from node

	if from == to {
		ParseError(ERR_ARROW_SELFLOOP)
		os.Exit(-1)
	}

	if link.Wgt != 1 {
		PVerbose("... Relation:",from,"--(",SST.ARROW_DIRECTORY[link.Arr].Long,",",link.Wgt,")->",to,link.Ctx)
	} else {
		PVerbose("... Relation:",from,"--",SST.ARROW_DIRECTORY[link.Arr].Long,"->",to,link.Ctx)
	}

        // Build PageMap

	link.Dst = toptr
	LINE_PATH = append(LINE_PATH,link)

        // Add to graph

	SST.AppendLinkToNode(frptr,link,toptr)

	// Double up the reverse definition for easy indexing of both in/out arrows
	// But be careful not the make the graph undirected by mistake

	invlink := GetLinkArrowByName(SST.ARROW_DIRECTORY[SST.INVERSE_ARROWS[link.Arr]].Short)

	SST.AppendLinkToNode(toptr,invlink,frptr)

}

//**************************************************************

func HandleNode(annotated string) SST.NodePtr {

	clean_ptr,clean_version := IdempAddNode(annotated)

	PVerbose("Event/item/node:",clean_version,"in chapter",SECTION_STATE)

	LINE_ITEM_REFS = append(LINE_ITEM_REFS,clean_ptr)
	
	if len(clean_version) != len(annotated) {
		AddBackAnnotations(clean_version,clean_ptr,annotated)
	}

	return clean_ptr
}

//**************************************************************

func IdempAddNode(s string) (SST.NodePtr,string) {

	clean_version := StripAnnotations(s)

	l,c := SST.StorageClass(s)

	var new_nodetext SST.Node
	new_nodetext.S = clean_version
	new_nodetext.L = l
	new_nodetext.Chap = SECTION_STATE
	new_nodetext.NPtr.Class = c

	iptr := SST.AppendTextToDirectory(new_nodetext,ParseError)

	// Build page map
	if LINE_PATH == nil {
		var leg SST.Link
		leg.Dst = iptr
		LINE_PATH = append(LINE_PATH,leg)
	}

	return iptr,clean_version
}

//**************************************************************
// Scan text input
//**************************************************************

func ReadFile(filename string) []rune {

	text := ReadTUF8File(filename)

	// clean unicode nonsense

	for r := range text {
		switch text[r] {
		case NON_ASCII_LQUOTE,NON_ASCII_RQUOTE:
			text[r] = '"'
		}
	}

	return text
}


//**************************************************************

func ReadToLast(src []rune,pos int, stop rune) (string,int) {

	var cpy []rune

	var starting_at = LINE_NUM

	for ; Collect(src,pos,stop,cpy) && pos < len(src); pos++ {
		cpy = append(cpy,src[pos])
	}

	if IsQuote(stop) && src[pos-1] != stop {
		e := fmt.Sprintf("%s starting at line %d (found token %s)",ERR_MISMATCH_QUOTE,starting_at,string(cpy))
		ParseError(e)
		os.Exit(-1)
	}

	token := string(cpy)

	token = strings.TrimSpace(token)

	return token,pos
}

//**************************************************************

func Collect(src []rune,pos int, stop rune,cpy []rune) bool {

	var collect bool = true

	// Quoted strings are tricky

	if IsQuote(stop) {
		var is_end bool

		if pos+1 >= len(src) {
			is_end= true
		} else {
			is_end = IsWhiteSpace(src[pos],src[pos+1])
		}

		if src[pos-1] == stop && is_end {
			return false
		} else {
			return true
		}

	}

	if pos >= len(src) || src[pos] == '\n' {
		return false
	}

	if stop == ALPHATEXT {
		collect = IsGeneralString(src,pos)
	} else {
		// a ::: cluster is special, we don't care how many

		if stop != ':' && !IsQuote(stop) { 
			return !LastSpecialChar(src,pos,stop)
		} else {
			var groups int = 0

			for r := 1; r < len(cpy)-1; r++ {

				if cpy[r] != ':' && cpy[r-1] == ':' {
					groups++
				}

				if cpy[r] != '"' && cpy[r-1] == '"' {
					groups++
				}
			}

			if groups > 1 {
				collect = !LastSpecialChar(src,pos,stop)
			}
		} 
	}

	return collect
}

//**************************************************************

func IsGeneralString(src []rune,pos int) bool {

	switch src[pos] {

        case ')':
	        ParseError(ERR_STRAY_PAREN)
		os.Exit(-1)
	case '(':
		return false
	case '#':
		return false
	case '\n':
		return false

	case '/':
		if src[pos+1] == '/' {
			return false
		}
	}

	return true
}

//**************************************************************

func IsQuote(r rune) bool {

	switch r {
	case '"','\'',NON_ASCII_LQUOTE,NON_ASCII_RQUOTE:
		return true
	}

	return false
}

//**************************************************************

func LastSpecialChar(src []rune,pos int, stop rune) bool {

	if src[pos] == '\n' {
		if stop != '"' {
			return true
		}
	}

	// Special case, but still don't understand why?!

	if src[pos] == '@' {
		return false
	}

	if pos > 0 && src[pos-1] == stop && src[pos] != stop {
		return true
	}

	return false
}

//**************************************************************

func UpdateLastLineCache() {

	if Dangler() {
		ParseError(ERR_MISSING_EVENT)
	}

	if CURRENT_FILE != "N4Lconfig.in" {
		PageMap(SECTION_STATE,CONTEXT_STATE,LINE_PATH,LINE_NUM,LINE_ALIAS)
	}

	LINE_NUM++

	// If this line was not blank, overwrite previous settings and reset

	if LINE_ITEM_STATE != ROLE_BLANK_LINE {
		
		if LINE_ITEM_CACHE["THIS"] != nil {
			LINE_ITEM_CACHE["PREV"] = LINE_ITEM_CACHE["THIS"]
		}
		if LINE_RELN_CACHE["THIS"] != nil {
			LINE_RELN_CACHE["PREV"] = LINE_RELN_CACHE["THIS"]
		}
	} 

	LINE_ITEM_CACHE["THIS"] = nil
	LINE_RELN_CACHE["THIS"] = nil
	LINE_ITEM_REFS = nil
	LINE_ITEM_COUNTER = 1
	LINE_RELN_COUNTER = 0
	LINE_ALIAS = ""
	LINE_PATH = nil

	LINE_ITEM_STATE = ROLE_BLANK_LINE
}

//**************************************************************

func PageMap(chapter string,ctxmap map[string]bool,path []SST.Link,line int,alias string) {

	if len(path) == 0 {
		return
	}

	var page_event SST.PageMap;
	var context []string
	var contextstr string

	for c := range ctxmap {
		context = append(context,c)
	}

	sort.Strings(context)

	for c := 0; c < len(context); c++ {
		contextstr += context[c]
		if c < len(context)-1 {
			contextstr += ", "
		}
	}

	page_event.Chapter = chapter
	page_event.Alias = alias
	page_event.Context = GetContext(nil)
	page_event.Line = line
	page_event.Path = path

	SST.PAGE_MAP = append(SST.PAGE_MAP,page_event)
}

//**************************************************************

func IsWhiteSpace(r,rn rune) bool {

	return (unicode.IsSpace(r) || r == '#' || r == '/' && rn == '/')
}

//**************************************************************

func IsBackReference(src []rune,pos int) bool {

	// Any non-whitespace before \n or ( means it's not a back reference

	for pos++; pos < len(src); pos++ {

		if src[pos] == '(' || src[pos] == '\n' || src[pos] == '#' {
			return true
		} else {
			if !unicode.IsSpace(src[pos]) {
				return false
			}
		}
	}

	return false
}

//**************************************************************

func Dangler() bool {

	switch LINE_ITEM_STATE {

	case ROLE_EVENT:
		return false
	case ROLE_LOOKUP:
		return false
	case ROLE_BLANK_LINE:
		return false
	case ROLE_SECTION:
		return false
	case ROLE_CONTEXT:
		return false
	case ROLE_CONTEXT_ADD:
		return false
	case ROLE_CONTEXT_SUBTRACT:
		return false
	case HAVE_MINUS:
		return false
	}

	return true
}

//**************************************************************

func ExtractContextExpression(token string) string {

	var expression string

	s := strings.Split(token, ":")

	for i := 1; i < len(s); i++ {
		if len(s[i]) > 1 {
			expression = strings.TrimSpace(s[i])
			break
		}
	}
	
	return expression
}

//**************************************************************

func CheckSequenceMode(context string, mode rune) {

	if (strings.Contains(context,"_sequence_")) {

		switch mode {
		case '+':
			PVerbose("\nStart sequence mode for items")
			SEQUENCE_MODE = true
			LAST_IN_SEQUENCE = ""

		case '-':
			PVerbose("End sequence mode for items\n")
			SEQUENCE_MODE = false
		}
	}

}

//**************************************************************

func LinkUpStorySequence(this string) {

	// Join together a sequence of nodes using default "(then)"

	if SEQUENCE_MODE && this != LAST_IN_SEQUENCE {

		if LINE_ITEM_COUNTER == 1 && LAST_IN_SEQUENCE != "" {
			
			PVerbose("* ... Sequence addition: ",LAST_IN_SEQUENCE,"-(",SEQUENCE_RELN,")->",this,"\n")
			
			last_iptr,_ := IdempAddNode(LAST_IN_SEQUENCE)
			this_iptr,_ := IdempAddNode(this)
			link := GetLinkArrowByName("(then)")
			SST.AppendLinkToNode(last_iptr,link,this_iptr)

			invlink := GetLinkArrowByName(SST.ARROW_DIRECTORY[SST.INVERSE_ARROWS[link.Arr]].Short)
			SST.AppendLinkToNode(this_iptr,invlink,last_iptr)

		}
		
		LAST_IN_SEQUENCE = this
	}
}

//**************************************************************

func StripAnnotations(fulltext string) string {

	var protected bool = false
	var deloused []rune
	var preserve_unicode = []rune(fulltext)

	for r := 0; r < len(preserve_unicode); r++ {

		if preserve_unicode[r] == '"' {
			protected = !protected
		}

		if !protected {
			skip,symb := EmbeddedSymbol(preserve_unicode,r)
			if skip > 0 {
				r += skip-1
				if unicode.IsSpace(preserve_unicode[r]) {
					ParseError(ERR_NON_WORD_WHITE+symb)
				}
				continue
			}
		}

		deloused = append(deloused,preserve_unicode[r])
	}

	return string(deloused)
}

//**************************************************************

func AddBackAnnotations(cleantext string,cleanptr SST.NodePtr,annotated string) {

	var protected bool = false

	reminder := fmt.Sprintf("%.30s...",cleantext)
	PVerbose("\n        Checking annotations from \""+reminder+"\"")

	for r := 0; r < len(annotated); r++ {

		if annotated[r] == '"' {
			protected = !protected
		} else {
			if !protected {
				skip,symb := EmbeddedSymbol([]rune(annotated),r)
				if skip > 0 {
					link := GetLinkArrowByName(ANNOTATION[symb])
					this_item := ExtractWord(annotated,r)
					this_iptr,_ := IdempAddNode(this_item)
					IdempAddLink(reminder,cleanptr,link,this_item,this_iptr)
					r += skip-1
					continue
				}
			}
		}
	}
}

//**************************************************************

func EmbeddedSymbol(fulltext []rune,offset int) (int,string) {

	for an := range ANNOTATION {

		// Careful of unicode, convert to runes
		uni := []rune(an)
		match := true

		for r := 0; r < len(uni) && r+offset < len(fulltext); r++ {
			if uni[r] != fulltext[offset+r] {
				match = false
				continue
			}

			// No space between marker and text
			if unicode.IsSpace(fulltext[offset+r+1]) {
				match = false
				continue
			}
		}

		if match {
			return len(an),an
		}
	}

	return 0,"UNKNOWN SYMBOL"
}

//**************************************************************

func ExtractWord(fulltext string,offset int) string {

	var protected bool = false
	var word string

	for r := offset+1; r < len(fulltext); r++ {

		if fulltext[r] == '"' {
			protected = !protected
		}

		if !protected && !unicode.IsLetter(rune(fulltext[r])) {
			word = strings.Trim(strings.TrimSpace(word),"\" ")
			return word
		}

		word += string(fulltext[r])
	}
	
	word = strings.Trim(strings.TrimSpace(word),"\" ")
	
	if len(word) <= WORD_MISTAKE_LEN {
		ParseError(ERR_SHORT_WORD+word)
	}
	
	return word
}

//**************************************************************
// Context logic
//**************************************************************

func ContextEval(s,op string) {

	expr := CleanExpression(s)

	or_parts := SplitWithParensIntact(expr,'|')

	if strings.Contains(s,"(") {
		ParseError(WARN_INADVISABLE_CONTEXT_EXPRESSION)
	}

	// +,-,= on CONTEXT_STATE

	switch op {
		
	case "=": 
		CONTEXT_STATE = make(map[string]bool)
		ModContext(or_parts,"+")
	default:
		ModContext(or_parts,op)
	}
}

//**************************************************************

func GetContext(ctx []string) []string {

	var merge = make(map[string]bool)
	var clist []string

	for c := range CONTEXT_STATE {
		merge[c] = true
	}


	for c := range ctx {
		merge[ctx[c]] = true
	}

	for c := range merge {
		clist = append(clist,c)
	}

	sort.Strings(clist)

	return clist
}

//**************************************************************

func CleanExpression(s string) string {

	s = TrimParen(s)
	r1 := regexp.MustCompile("[|,]+") 
	s = r1.ReplaceAllString(s,"|") 
	r2 := regexp.MustCompile("[&]+") 
	s = r2.ReplaceAllString(s,".") 
	r3 := regexp.MustCompile("[.]+") 
	s = r3.ReplaceAllString(s,".") 

	return s
}

// ***********************************************************************

func SplitWithParensIntact(expr string,split_ch byte) []string {

	var token string = ""
	var set []string

	for c := 0; c < len(expr); c++ {

		switch expr[c] {

		case split_ch:
			set = append(set,token)
			token = ""

		case '(':
			subtoken,offset := Paren(expr,c)
			token += subtoken
			c = offset-1

		default:
			token += string(expr[c])
		}
	}

	if len(token) > 0 {
		set = append(set,token)
	}

	return set
} 

// ***********************************************************************

func Paren(s string, offset int) (string,int) {

	var level int = 0

	for c := offset; c < len(s); c++ {

		if s[c] == '(' {
			level++
			continue
		}

		if s[c] == ')' {
			level--
			if level == 0 {
				token := s[offset:c+1]
				return token, c+1
			}
		}
	}

	return "bad expression", -1
}


// ***********************************************************************

func TrimParen(s string) string {

	var level int = 0
	var trim = true

	if len(s) == 0 {
		return s
	}

	s = strings.TrimSpace(s)

	if s[0] != '(' {
		return s
	}

	for c := 0; c < len(s); c++ {

		if s[c] == '(' {
			level++
			continue
		}

		if level == 0 && c < len(s)-1 {
			trim = false
		}
		
		if s[c] == ')' {
			level--

			if level == 0 && c == len(s)-1 {
				
				var token string
				
				if trim {
					token = s[1:len(s)-1]
				} else {
					token = s
				}
				return token
			}
		}
	}
	
	return s
}

//**************************************************************

func ModContext(list []string,op string) {

	for or_frag := range list {

		frag := strings.TrimSpace(list[or_frag])

		if len(frag) == 0 {
			continue
		}

		switch op {
		case "+":
			CONTEXT_STATE[frag] = true

		case "-": // to remove, we also need to look at children
			for cand := range CONTEXT_STATE {
				and_parts := SplitWithParensIntact(cand,'.') 
		
				for part := range and_parts {

					if strings.Contains(and_parts[part],frag) {
						delete(CONTEXT_STATE,cand)
					}
				}
			}
		}

	}
}

//**************************************************************

func CheckNonNegative(i int) {

	if i < 0 {
		ParseError(ERR_MISSING_ITEM_SOMEWHERE)
		os.Exit(-1)
	}
}

//**************************************************************

func CheckSection() {

	if len(SECTION_STATE) == 0 {
		ParseError(ERR_MISSING_SECTION)
		os.Exit(-1)
	}
}

//**************************************************************

func AllCaps(s string) bool {

	if len(s) <= WORD_MISTAKE_LEN {
		return false
	}

	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) || unicode.IsNumber(r) {
			return false
		}
	}

	return true
}

//**************************************************************

func StripParen(token string) string {

	token =	strings.TrimSpace(token[1:])

	if token[0] == '(' {
		token =	strings.TrimSpace(token[1:])
	}

	if token[len(token)-1] == ')' {
		token =	token[:len(token)-1]
	}

	return token
}

//**************************************************************
// Tools
//**************************************************************

func PrintNodeSystem(n int,org SST.Node, count_links *[4]int) {

	fmt.Println(n,"\t",org.S)

	for sttype := range org.I {
		for lnk := range org.I[sttype] {
			count_links[FlatSTType(sttype)]++
			PrintLink(org.I[sttype][lnk])
		}
	}
	fmt.Println()
}

//**************************************************************

func PrintLink(l SST.Link) {

	to := SST.GetNodeTxtFromPtr(l.Dst)
	arrow := SST.ARROW_DIRECTORY[l.Arr]
	Verbose("\t ... --(",arrow.Long,",",l.Wgt,")->",to,l.Ctx," \t . . .",SST.PrintSTAIndex(arrow.STAindex))
}

// **************************************************************************

func ParseError(message string) {

	const red = "\033[31;1;1m"
	const endred = "\033[0m"

	fmt.Print("\n",LINE_NUM,":",red)
	fmt.Println("N4L",CURRENT_FILE,message,"at line", LINE_NUM,endred)
	Diag("N4L",CURRENT_FILE,message,"at line", LINE_NUM)

}

//**************************************************************

func ReadTUF8File(filename string) []rune {
	
	content,err := ioutil.ReadFile(filename)
	
	if err != nil {
		ParseError(ERR_NO_SUCH_FILE_FOUND+filename)
		os.Exit(-1)
	}

	var unicode []rune
	
	for i, w := 0, 0; i < len(content); i += w {
                runeValue, width := utf8.DecodeRuneInString(string(content)[i:])
                w = width

		unicode = append(unicode,runeValue)
	}

	return unicode
}

//**************************************************************

func Usage() {
	
	fmt.Printf("usage: N4L [-v] [-u] [-s] [file].dat\n")
	flag.PrintDefaults()
	os.Exit(2)
}

//**************************************************************

func Verbose(a ...interface{}) {

	line := fmt.Sprintln(a...)
	
	if DIAGNOSTIC {
		AppendStringToFile(TEST_DIAG_FILE,line)
	}


	if VERBOSE {
		fmt.Print(line)
	}
}

//**************************************************************

func PVerbose(a ...interface{}) {

	const green = "\x1b[36m"
	const endgreen = "\x1b[0m"

	if VERBOSE {
		fmt.Print(LINE_NUM,":\t",green)
		fmt.Println(a...)
		fmt.Print(endgreen)
	}
}

//**************************************************************

func Box(a ...interface{}) {

	if VERBOSE {

		fmt.Println("\n------------------------------------")
		fmt.Println(a...)
		fmt.Println("------------------------------------\n")
	}
}

//**************************************************************

func DiagnosticName(filename string) string {

	return "test_output/"+filename+"_test_log"

}

//**************************************************************

func Diag(a ...interface{}) {

	// Log diagnostic output for self-diagnostic tests

	if DIAGNOSTIC {
		s := fmt.Sprintln(a...)
		prefix := fmt.Sprint(LINE_NUM,":")
		AppendStringToFile(TEST_DIAG_FILE,prefix+s)
	}
}

//**************************************************************

func AppendStringToFile(name string, s string) {

	// strip out \r that mess up the file format but are useful for term

	san := strings.Replace(s,"\r","",-1)

	f, err := os.OpenFile(name,os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Couldn't open for write/append to",name,err)
		f.Close()
		return
	}

	_, err = f.WriteString(san)

	if err != nil {
		fmt.Println("Couldn't write/append to",name,err)
	}

	f.Close()
}


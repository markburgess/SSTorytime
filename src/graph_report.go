//******************************************************************
//
// Study graph properties - print a basic report of key map features
//
//******************************************************************

package main

import (
	"fmt"
	"strings"
	"sort"
	"flag"
	"os"
        SST "SSTorytime"
)

var CHAPTER string
var CONTEXT []string
var STTYPES []int
var DEPTH int

//******************************************************************

func main() {

	Init()

	load_arrows := true
	sst := SST.Open(load_arrows)

	chaps := SST.GetDBChaptersMatchingName(sst,CHAPTER)

	for chap := range chaps {
		AnalyzeGraph(sst,chaps[chap],CONTEXT,STTYPES,DEPTH) 
	}

	SST.Close(sst)
}

//**************************************************************

func Usage() {
	
	fmt.Printf("usage: graph_report [-sttype comma separated L,C,P,N] [-depth integer] [-chapter comma separated string] [context]\n")
	flag.PrintDefaults()

	os.Exit(2)
}

//**************************************************************

func Init() []string {

	flag.Usage = Usage

	chapterPtr := flag.String("chapter", "", "a optional substring to match specific chapters")
	sttypePtr := flag.String("sttype", "+L", "link st-types e.g. L,C,P,N")
	depthPtr := flag.Int("depth", 3, "maximum probe depth for loop detection")

	flag.Parse()
	args := flag.Args()

	CHAPTER = "none"

	if *chapterPtr != "" {
		CHAPTER = *chapterPtr
	}

	if *sttypePtr != "" {

		var sttypes = make(map[int]bool)
		array := strings.Split(*sttypePtr,",")
		for t := range array {
			switch array[t] {
			case "L","+L": 
				sttypes[1] = true
			case "C","+C": 
				sttypes[2] = true
			case "E","+E": 
				sttypes[3] = true
			case "P","+P": 
				sttypes[3] = true
			case "N","+N","-N": 
				sttypes[4] = true
			case "-L": 
				sttypes[-1] = true
			case "-C": 
				sttypes[-2] = true
			case "-E": 
				sttypes[-3] = true
			case "-P": 
				sttypes[-3] = true
			default:
				fmt.Println("Unknown sttype",array[t],"(should be in { L,C,E,N } +/-)")
				os.Exit(-1)
			}
		}

		for t := range sttypes {
			STTYPES = append(STTYPES,t)
		}
	}

	DEPTH = *depthPtr

	SST.MemoryInit()

	return args
}

//******************************************************************

func AnalyzeGraph(sst SST.PoSST,chapter string,context []string,sttypes []int,depth int) {


	adj,nodekey := SST.GetDBAdjacentNodePtrBySTType(sst,sttypes,chapter,context,false)
	symb := SST.SymbolMatrix(adj)
	sadj := SST.SymmetrizeMatrix(adj)
	num := GetNumberOfLinks(adj)
	distribution := GetNameDistribution(nodekey)
	total := len(nodekey)
	max := total*(total-1)

	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("Analysing chapter \"%s\", context %v to path length %d\n",chapter,context,depth)
	fmt.Println("----------------------------------------------------------------\n")

	fmt.Println("\n* TOTAL NODES IN THE SEARCH REGION",total)
	fmt.Printf("\n* TOTAL DIRECTED LINKS = %d of possible %d = %2.2f %%\n",num,max,float64(num)/float64(max))
	fmt.Printf("\n* DISTRIBUTION OF NAME TYPE/LENGTHS:\n")
		for class := 1; class < 7; class++ {
			if distribution[class] > 0 {
				fmt.Printf("  - %s : %d / %d\n",SST.CLASS_CHANNEL_DESCRIPTION[class],distribution[class],total)
			}
		}
	
	sources,sinks := SST.GetDBSingletonBySTType(sst,sttypes,chapter,context)

	fmt.Print("\n\n* PROCESS ORIGINS / ROOT DEPENDENCIES / PATH SOURCES for (")
	for st := range sttypes {
		fmt.Print("\"",SST.STTypeName(sttypes[st]),"\"")
	}
	fmt.Println(") in",chapter)
	fmt.Println("")

	PrintNodes(sst,sources)

	fmt.Println("")
	fmt.Print("\n\n* FINAL END-STATES / PATH SINK NODES for (")
	for st := range sttypes {
		fmt.Print("\"",SST.STTypeName(sttypes[st]),"\"")
	}
	fmt.Println(") in",chapter)
	fmt.Println("")

	PrintNodes(sst,sinks)

	fmt.Println("")
	fmt.Println("* DIRECTED LOOPS AND CYCLES (max depth < ",depth,"):\n")
	fmt.Println("\n")

	// Find power matrices

	an := make([][][]float32,depth+1)
	sn := make([][][]string,depth+1)

	an[1] = adj
	sn[1] = symb
	acyclic := true

	for power := 2; power <= depth; power++ {

		if power % 2 == 0 {
			an[power],sn[power] = SST.SymbolicMultiply(an[power-1],adj,sn[power-1],symb)
		} else {
			an[power],sn[power] = SST.SymbolicMultiply(an[power-1],adj,sn[power-1],symb)
		}

		loop,_ := AnalyzePowerMatrix(sst,sn[power])

		for m := range loop {
			acyclic = false
			length := len(strings.Split(m,")("))
			fmt.Println("  - Cycle of length",length,"with members",m)
		}
	}

	if acyclic {
		fmt.Println("   - Acyclic")
	}

	// Look for appointed nodes

	fmt.Println("\n* APPOINTED NODES (nodes pointed to by at least 2 others thus correlating them) ")

	for st := range sttypes {
		var ama map[SST.ArrowPtr][]SST.Appointment
		
		ama = SST.GetAppointedNodesBySTType(sst,sttypes[st],context,chapter,2)
		
		for arrowptr := range ama {
			
			arr_dir := SST.GetDBArrowByPtr(sst,arrowptr)
			
			// Appointment list
			for n := 0; n < len(ama[arrowptr]); n++ {
				
				appointed_nptr := ama[arrowptr][n].NTo
				appointed := SST.GetDBNodeByNodePtr(sst,appointed_nptr)
				dim := len(ama[arrowptr][n].NFrom)
				
				fmt.Printf("\n   Appointer correlates -> %d appointed nodes (%s ...) in chapter \"%s\"\n\n",dim,appointed.S,chapter)
				
				// Appointers list
				for m := range ama[arrowptr][n].NFrom {
					node := SST.GetDBNodeByNodePtr(sst,ama[arrowptr][n].NFrom[m])
					stname := SST.STTypeName(SST.STIndexToSTType(arr_dir.STAindex))
					fmt.Printf("     %.40s --(%s : %s)--> %.40s...   - in context %v\n",node.S,arr_dir.Long,stname,appointed.S,context)
				}
			}
			
			fmt.Println()
		}
	}

	// Now find the undirected graph properties 

	fmt.Println("")
	evc := SST.ComputeEVC(sadj)

	fmt.Println("* SYMMETRIZED EIGENVECTOR CENTRALITY = FLOW RESERVOIR CAPACITANCE AT EQUILIBRIUM = \n")

	PrintVector(sst,evc,nodekey)

	regions,evctop,path := SST.FindGradientFieldTop(sadj,evc)

	fmt.Println("")
	if len(regions) == 1 {
		fmt.Println("* THERE IS",len(regions),"LOCAL MAXIMA IN THE EQUILIBRIUM EVC LANDSCAPE:\n")
	} else {
		fmt.Println("* THERE ARE",len(regions),"LOCAL MAXIMA IN THE EQUILIBRIUM EVC LANDSCAPE:\n")
	}

	for reg := range regions {
		fmt.Println("  - subregion of maximum",reg,"consisting of nodes",regions[reg])
		PrintKeyNodes(sst,regions[reg],nodekey)
	}

	fmt.Println("\n* HILL-CLIMBING EVC-LAMDSCAPE GRADIENT PATHS:\n")

	for index := 0; index < len(evc); index++ {
		fmt.Println("     - Path node",index,"has local maximum at node *",evctop[index],"*, hop distance",len(path[index])-1,"along",path[index])		
	}

}

//**************************************************************

func GetNumberOfLinks(a [][]float32) int {

	count := 0
	for i := range a {
		for j := range a[i] {
			if a[i][j] > 0 {
				count++
			}
		}
	}
	return count
}

//**************************************************************

func GetNameDistribution(nodeptr []SST.NodePtr) [7]int {

	var dist [7]int

	for n := range nodeptr {
		dist[nodeptr[n].Class]++
	}

	return dist
}

//**************************************************************

func AnalyzePowerMatrix(sst SST.PoSST,symbolic [][]string) (map[string]int,map[string][]int) {

	var loop = make(map[string]int)
	var memberlist = make(map[string][]int)

	for r := 0; r < len(symbolic); r++ {

		// check the diagonal

		if len(symbolic[r][r]) == 0 {
			continue
		}

		var distrib = make(map[string]int)
		var nodes []string

		vec := strings.Split(symbolic[r][r],"*")
		
		for i := 0; i < len(vec); i++ {
			distrib[vec[i]]++
		}

		var degeneracy int

		for d := range distrib {
			degeneracy = distrib[d] / 2
			break
		}

		for r := range distrib {
			nodes = append(nodes,r)
		}

		sort.Strings(nodes)
		var members string
		var membindex []int
		var v int

		for n := 0; n < len(nodes); n++ {
			members += fmt.Sprintf("(%s)",nodes[n])
			fmt.Sscanf(nodes[n],"%d",&v)
			membindex = append(membindex,v)
		}

		loop[members] = degeneracy
		memberlist[members] = membindex
	}

	return loop,memberlist
}

//**************************************************************

func PrintNodes(sst SST.PoSST,nptrs []SST.NodePtr) {

	for n := range nptrs {
		node := SST.GetDBNodeByNodePtr(sst,nptrs[n])
		fmt.Printf("   - NPtr(%d,%d) -> %s\n",nptrs[n].Class,nptrs[n].CPtr,node.S)
	}
}

//**************************************************************

func PrintKeyNodes(sst SST.PoSST,m []int,nodekey []SST.NodePtr) {

	for member := range m {
		nptr := nodekey[m[member]]
		node := SST.GetDBNodeByNodePtr(sst,nptr)
		fmt.Printf("     - where %d -> %s\n",m[member],node.S)
	}
}

//**************************************************************

func PrintVector(sst SST.PoSST,vector []float32,nodekey []SST.NodePtr) {

	for row := 0; row < len(vector); row++ {
		nptr := nodekey[row]
		node := SST.GetDBNodeByNodePtr(sst,nptr)
		fmt.Printf("   ( %3.3f ) <- %d = %s\n",vector[row],row,node.S)
	}
	fmt.Println()
}

//**************************************************************

func PrintMatrix(matrix [][]float32,symbolic [][]string,str string) {

	fmt.Printf("                 DIAG %s \n",str)

	for row := 0; row < len(matrix); row++ {
		for col := 0; col < len(matrix[row]); col++ {
			fmt.Printf("%2.0f ",matrix[row][col])
		}

		fmt.Printf(" %1.1f   ...",matrix[row][row])
		if matrix[row][row] > 0 {
			fmt.Printf("      %s    (loop)\n",symbolic[row][row])
		} else {
			fmt.Println()
		}
	}
	fmt.Println()
}


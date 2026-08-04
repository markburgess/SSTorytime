// Graph-report CLI logic (ported from upstream graph_report).
package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

var CLASS_CHANNEL_DESCRIPTION = []string{"", "single word ngram", "two word ngram", "three word ngram",
	"string less than 128 chars", "string less than 1024 chars", "string greater than 1024 chars"}

func AnalyzeGraphCLI(ctx context.Context, sst SST.PoSST, chapter string, context []string, sttypes []int, depth int) {

	adj, nodekey := SST.GetDBAdjacentNodePtrBySTType(ctx, sst, sttypes, chapter, context, false)
	symb := SST.SymbolMatrix(adj)
	sadj := SST.SymmetrizeMatrix(adj)
	num := GraphReportGetNumberOfLinks(adj)
	distribution := GraphReportGetNameDistribution(nodekey)
	total := len(nodekey)
	max := total * (total - 1)

	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("Analysing chapter \"%s\", context %v to path length %d\n", chapter, context, depth)
	fmt.Println("----------------------------------------------------------------\n")

	fmt.Println("\n* TOTAL NODES IN THE SEARCH REGION", total)
	fmt.Printf("\n* TOTAL DIRECTED LINKS = %d of possible %d = %2.2f %%\n", num, max, float64(num)/float64(max))
	fmt.Printf("\n* DISTRIBUTION OF NAME TYPE/LENGTHS:\n")
	for class := 1; class < 7; class++ {
		if distribution[class] > 0 {
			fmt.Printf("  - %s : %d / %d\n", CLASS_CHANNEL_DESCRIPTION[class], distribution[class], total)
		}
	}

	sources, sinks := SST.GetDBSingletonBySTType(ctx, sst, sttypes, chapter, context)

	fmt.Print("\n\n* PROCESS ORIGINS / ROOT DEPENDENCIES / PATH SOURCES for (")
	for st := range sttypes {
		fmt.Print("\"", SST.STTypeName(sttypes[st]), "\"")
	}
	fmt.Println(") in", chapter)
	fmt.Println("")

	GraphReportPrintNodes(ctx, &sst, sources)

	fmt.Println("")
	fmt.Print("\n\n* FINAL END-STATES / PATH SINK NODES for (")
	for st := range sttypes {
		fmt.Print("\"", SST.STTypeName(sttypes[st]), "\"")
	}
	fmt.Println(") in", chapter)
	fmt.Println("")

	GraphReportPrintNodes(ctx, &sst, sinks)

	fmt.Println("")
	fmt.Println("* DIRECTED LOOPS AND CYCLES (max depth < ", depth, "):\n")
	fmt.Println("\n")

	// Find power matrices

	an := make([][][]float32, depth+1)
	sn := make([][][]string, depth+1)

	an[1] = adj
	sn[1] = symb
	acyclic := true

	for power := 2; power <= depth; power++ {

		if power%2 == 0 {
			an[power], sn[power] = SST.SymbolicMultiply(an[power-1], adj, sn[power-1], symb)
		} else {
			an[power], sn[power] = SST.SymbolicMultiply(an[power-1], adj, sn[power-1], symb)
		}

		loop, layers := GraphReportAnalyzePowerMatrix(ctx, sst, sn[power])
		if len(layers) < 0 {
			return
		}

		for m := range loop {
			acyclic = false
			length := len(strings.Split(m, ")("))
			fmt.Println("  - Cycle of length", length, "with members", m)
		}
	}

	if acyclic {
		fmt.Println("   - Acyclic")
	}

	// Look for appointed nodes

	fmt.Println("\n* APPOINTED NODES (nodes pointed to by at least 2 others thus correlating them) ")

	for st := range sttypes {
		var ama map[SST.ArrowPtr][]SST.Appointment

		ama = SST.GetAppointedNodesBySTType(ctx, &sst, sttypes[st], context, chapter, 2)

		for arrowptr := range ama {

			arr_dir := SST.GetDBArrowByPtr(ctx, &sst, arrowptr)

			// Appointment list
			for n := 0; n < len(ama[arrowptr]); n++ {

				appointed_nptr := ama[arrowptr][n].NTo
				appointed := SST.GetDBNodeByNodePtr(ctx, &sst, appointed_nptr)
				dim := len(ama[arrowptr][n].NFrom)

				fmt.Printf("\n   Appointer correlates -> %d appointed nodes (%s ...) in chapter \"%s\"\n\n", dim, appointed.S, chapter)

				// Appointers list
				for m := range ama[arrowptr][n].NFrom {
					node := SST.GetDBNodeByNodePtr(ctx, &sst, ama[arrowptr][n].NFrom[m])
					stname := SST.STTypeName(SST.STIndexToSTType(arr_dir.STAindex))
					fmt.Printf("     %.40s --(%s : %s)--> %.40s...   - in context %v\n", node.S, arr_dir.Long, stname, appointed.S, context)
				}
			}

			fmt.Println()
		}
	}

	// Now find the undirected graph properties

	fmt.Println("")
	evc := SST.ComputeEVC(sadj)

	fmt.Println("* SYMMETRIZED EIGENVECTOR CENTRALITY = FLOW RESERVOIR CAPACITANCE AT EQUILIBRIUM = \n")

	GraphReportPrintVector(ctx, &sst, evc, nodekey)

	regions, evctop, path := SST.FindGradientFieldTop(sadj, evc)

	fmt.Println("")
	if len(regions) == 1 {
		fmt.Println("* THERE IS", len(regions), "LOCAL MAXIMA IN THE EQUILIBRIUM EVC LANDSCAPE:\n")
	} else {
		fmt.Println("* THERE ARE", len(regions), "LOCAL MAXIMA IN THE EQUILIBRIUM EVC LANDSCAPE:\n")
	}

	for reg := range regions {
		fmt.Println("  - subregion of maximum", reg, "consisting of nodes", regions[reg])
		GraphReportPrintKeyNodes(ctx, &sst, regions[reg], nodekey)
	}

	fmt.Println("\n* HILL-CLIMBING EVC-LAMDSCAPE GRADIENT PATHS:\n")

	for index := 0; index < len(evc); index++ {
		fmt.Println("     - Path node", index, "has local maximum at node *", evctop[index], "*, hop distance", len(path[index])-1, "along", path[index])
	}

}

//**************************************************************

func GraphReportGetNumberOfLinks(a [][]float32) int {

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

func GraphReportGetNameDistribution(nodeptr []SST.NodePtr) [7]int {

	var dist [7]int

	for n := range nodeptr {
		dist[nodeptr[n].Class]++
	}

	return dist
}

//**************************************************************

func GraphReportAnalyzePowerMatrix(ctx context.Context, sst SST.PoSST, symbolic [][]string) (map[string]int, map[string][]int) {

	var loop = make(map[string]int)
	var memberlist = make(map[string][]int)

	for r := 0; r < len(symbolic); r++ {

		// check the diagonal

		if len(symbolic[r][r]) == 0 {
			continue
		}

		var distrib = make(map[string]int)
		var nodes []string

		vec := strings.Split(symbolic[r][r], "*")

		for i := 0; i < len(vec); i++ {
			distrib[vec[i]]++
		}

		var degeneracy int

		for d := range distrib {
			degeneracy = distrib[d] / 2
			break
		}

		for r := range distrib {
			nodes = append(nodes, r)
		}

		sort.Strings(nodes)
		var members string
		var membindex []int
		var v int

		for n := 0; n < len(nodes); n++ {
			members += fmt.Sprintf("(%s)", nodes[n])
			fmt.Sscanf(nodes[n], "%d", &v)
			membindex = append(membindex, v)
		}

		loop[members] = degeneracy
		memberlist[members] = membindex
	}

	return loop, memberlist
}

//**************************************************************

func GraphReportPrintNodes(ctx context.Context, sst *SST.PoSST, nptrs []SST.NodePtr) {

	for n := range nptrs {
		node := SST.GetDBNodeByNodePtr(ctx, sst, nptrs[n])
		fmt.Printf("   - NPtr(%d,%d) -> %s\n", nptrs[n].Class, nptrs[n].CPtr, node.S)
	}
}

//**************************************************************

func GraphReportPrintKeyNodes(ctx context.Context, sst *SST.PoSST, m []int, nodekey []SST.NodePtr) {

	for member := range m {
		nptr := nodekey[m[member]]
		node := SST.GetDBNodeByNodePtr(ctx, sst, nptr)
		fmt.Printf("     - where %d -> %s\n", m[member], node.S)
	}
}

//**************************************************************

func GraphReportPrintVector(ctx context.Context, sst *SST.PoSST, vector []float32, nodekey []SST.NodePtr) {

	for row := 0; row < len(vector); row++ {
		nptr := nodekey[row]
		node := SST.GetDBNodeByNodePtr(ctx, sst, nptr)
		fmt.Printf("   ( %3.3f ) <- %d = %s\n", vector[row], row, node.S)
	}
	fmt.Println()
}

//**************************************************************

func GraphReportPrintMatrix(matrix [][]float32, symbolic [][]string, str string) {

	fmt.Printf("                 DIAG %s \n", str)

	for row := 0; row < len(matrix); row++ {
		for col := 0; col < len(matrix[row]); col++ {
			fmt.Printf("%2.0f ", matrix[row][col])
		}

		fmt.Printf(" %1.1f   ...", matrix[row][row])
		if matrix[row][row] > 0 {
			fmt.Printf("      %s    (loop)\n", symbolic[row][row])
		} else {
			fmt.Println()
		}
	}
	fmt.Println()
}

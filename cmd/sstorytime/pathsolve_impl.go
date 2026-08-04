// Pathsolve CLI logic (ported from upstream pathsolve).
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

func PathSolveCLI(ctx context.Context, sst SST.PoSST, chapter, cntext, begin, end string) {

	const mindepth = 2
	const maxdepth = 20
	var count int
	var arrowptrs []SST.ArrowPtr
	var sttype []int

	start_bc := []string{begin}
	end_bc := []string{end}
	context := strings.Split(cntext, ",")

	var leftptrs, rightptrs []SST.NodePtr

	for n := range start_bc {
		leftptrs = append(leftptrs, SST.GetDBNodePtrMatchingName(ctx, sst, start_bc[n], chapter)...)
	}

	for n := range end_bc {
		rightptrs = append(rightptrs, SST.GetDBNodePtrMatchingName(ctx, sst, end_bc[n], chapter)...)
	}

	if leftptrs == nil || rightptrs == nil {
		fmt.Println("No paths available from end points", begin, "TO", end, "in chapter", chapter)
		return
	}

	fmt.Printf("\n\n Paths < end_set= {%s} | {%s} = start set>\n\n", PathSolveShowNode(ctx, &sst, rightptrs), PathSolveShowNode(ctx, &sst, leftptrs))

	solutions := SST.GetPathsAndSymmetries(ctx, &sst, leftptrs, rightptrs, chapter, context, arrowptrs, sttype, mindepth, maxdepth)

	// Find the path matrix

	var betweenness = make(map[string]int)

	if len(solutions) > 0 {

		for s := 0; s < len(solutions); s++ {
			prefix := fmt.Sprintf(" - story path: ")
			SST.PrintLinkPath(ctx, &sst, solutions, s, prefix, "", nil)
			betweenness = PathSolveTallyPath(ctx, &sst, solutions[s], betweenness)
		}
		count++
	}

	if len(solutions) == 0 {
		fmt.Println("No paths satisfy constraints", context, " between end points", begin, "TO", end, "in chapter", chapter)
		os.Exit(-1)
	}

	// Calculate the node layer sets S[path][depth]

	fmt.Println(" *\n *\n * PATH ANALYSIS: into node flow equivalence groups\n *\n *\n\n")

	//supernodes := SST.SuperNodesByConicPath(solutions,maxdepth)

	// *** Summarize paths

	supers := SST.SuperNodes(ctx, sst, solutions, maxdepth)

	for s := range supers {
		fmt.Println("   - Supernode:", supers[s])
	}

	fmt.Println("\n *\n *\n * FLOW IMPORTANCE:\n *\n *\n")

	betw := SST.BetweenNessCentrality(ctx, sst, solutions)

	for b := range betw {
		fmt.Println("   - Betweenness centrality:", betw[b])
	}

}

// **********************************************************

func PathSolveTallyPath(ctx context.Context, sst *SST.PoSST, path []SST.Link, between map[string]int) map[string]int {

	// count how often each node appears in the different path solutions

	for leg := range path {
		n := SST.GetDBNodeByNodePtr(ctx, sst, path[leg].Dst)
		between[n.S]++
	}

	return between
}

// **********************************************************

func PathSolveShowNode(ctx context.Context, sst *SST.PoSST, nptr []SST.NodePtr) string {

	var ret string

	for n := range nptr {
		node := SST.GetDBNodeByNodePtr(ctx, sst, nptr[n])
		ret += fmt.Sprintf("%.30s, ", node.S)
	}

	return ret
}

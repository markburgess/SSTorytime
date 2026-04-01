// **************************************************************************
//
// path_wave_search.go 
//
// **************************************************************************

package SSTorytime

import (
	"fmt"
	"sync"
	_ "github.com/lib/pq"

)


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



//
// path_wave_search.go 
//



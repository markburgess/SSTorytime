
package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := true
	sst := SST.Open(load_arrows)

	start_bc := "a1"
	end_bc := "b6"
	chapter := "double slit example"
	context := []string{""}
	arrowptrs := []SST.ArrowPtr{20,21}
	sttype := []int{1}
	maxdepth := 10

	leftptrs := SST.GetDBNodePtrMatchingName(sst,start_bc,"")
	rightptrs := SST.GetDBNodePtrMatchingName(sst,end_bc,"")

	// Contra colliding wavefronts as path integral solver

	solutions := GetPathsAndSymmetries2(sst,leftptrs,rightptrs,chapter,context,arrowptrs,sttype,maxdepth)

	if len(solutions) > 0 {		
		for s := 0; s < len(solutions); s++ {
			PrintConstrainedLinkPath(sst,solutions,s,"-",chapter,context,arrowptrs,sttype)
		}
	}
}

//******************************************************************

func GetPathsAndSymmetries2(sst SST.PoSST,start_set,end_set []SST.NodePtr,chapter string,context []string,arrowptrs []SST.ArrowPtr,sttypes []int,maxdepth int) [][]SST.Link {

	var left_paths, right_paths [][]SST.Link
	var ldepth,rdepth int = 2,2
	var Lnum,Rnum int
	var solutions [][]SST.Link
	var loop_corrections [][]SST.Link

	if start_set == nil || end_set == nil {
		return nil
	}

	// Complete Adjoint types for inverse/acceptor wave

	adj_arrowptrs := SST.AdjointArrows(arrowptrs)
	adj_sttypes := SST.AdjointSTtype(sttypes)

	// Prime paths

//	AUTOMATICALLY ALIGN + and - , prioritize intended direction

//ADJOINT SHOULD REVERSE THE ORDER OF SSTYPE and ARROW

//This is why we chose bwd if no fwd found.... because there might not be solutions +1 only -1 from a node
//	If we choose "any", we need to relax the exclusion rules

	left_paths,Lnum = SST.GetConstraintConePathsAsLinks(sst,start_set,ldepth,chapter,context,arrowptrs,sttypes,maxdepth)
	right_paths,Rnum = SST.GetConstraintConePathsAsLinks(sst,end_set,rdepth,chapter,context,adj_arrowptrs,adj_sttypes,maxdepth)

	// Expand waves

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		fmt.Println("   ..Waves searching",ldepth,rdepth)

		solutions,loop_corrections = SST.WaveFrontsOverlap(sst,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		if len(solutions) > 0 {
			fmt.Println("   ..DAG solutions:")
			return solutions
		}

		if len(loop_corrections) > 0 {
			fmt.Println("   ..Only non-DAG solutions:")
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

func IncConstraintConeLinks(sst SST.PoSST,cone [][]SST.Link,chapter string ,context []string,arrowptrs []SST.ArrowPtr,sttypes []int,maxdepth int) [][]SST.Link {

	var expanded_cone [][]SST.Link

	for p := 0; p < len(cone); p++ {
		
		branch := cone[p]
		var exclude = make(map[SST.NodePtr]bool)

		for _,prev := range branch {
			exclude[prev.Dst] = true
		}

		tip := []SST.NodePtr{branch[len(branch)-1].Dst}

		shoots := GetConstrainedFwdLinks(sst,tip,chapter,context,sttypes,arrowptrs,maxdepth)

		// unfurl branches

		for _,satellite := range shoots {
			if !exclude[satellite.Dst] {
				exclude[satellite.Dst] = true
				var delta []SST.Link
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

func GetConstrainedFwdLinks(sst SST.PoSST,start []SST.NodePtr,chapter string,context []string,sttypes []int,arrows []SST.ArrowPtr,maxlimit int) []SST.Link {

	var ret []SST.Link

	remove_accents,stripped := SST.IsBracketedSearchTerm(chapter)
	chapter = "%"+stripped+"%"
	rm_acc := "false"

	if remove_accents {
		rm_acc = "true"
	}

	start = append(start,SST.NONODE)
	excl := SST.FormatSQLNodePtrArray(start)
	arr := SST.FormatSQLIntArray(SST.Arrow2Int(arrows))
	cnt := SST.FormatSQLStringArray(context)

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
				orbit := SST.ParseLinkArray(whole)

				for _,lnk := range orbit {
					ret = append(ret,lnk)
				}
			}
			row.Close()
		}
	}

	return ret
}


// **********************************************************

func PrintConstrainedLinkPath(sst SST.PoSST, cone [][]SST.Link, p int, prefix string,chapter string,context []string,arrows []SST.ArrowPtr,sttype []int) {

	for l := 1; l < len(cone[p]); l++ {
		link := cone[p][l]

		if !ArrowAllowed(sst,link.Arr,arrows,sttype) {
			return
		}
	}

	SST.PrintLinkPath(sst,cone,p,prefix,chapter,context)
}

// **********************************************************

func ArrowAllowed(sst SST.PoSST,arr SST.ArrowPtr, arrlist []SST.ArrowPtr, stlist []int) bool {

	st_ok := false
	arr_ok := false

	staidx := SST.GetDBArrowByPtr(sst,arr).STAindex
	st := SST.STIndexToSTType(staidx)

	if arrlist != nil {
		for a := range arrlist {
			if arr == arrlist[a] {
				arr_ok = true
				break
			}
		}
	} else {
		arr_ok = true
	}

	if stlist != nil {
		for i := range stlist {
			if stlist[i] == st {
				st_ok = true
				break
			}
		}
	} else {
		st_ok = true
	}

	if st_ok || arr_ok {
		return true
	}

	return false
}


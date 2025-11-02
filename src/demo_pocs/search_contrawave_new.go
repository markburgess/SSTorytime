
package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := true
	sst := SST.Open(load_arrows)

	start_bc := "!gun!"
	end_bc := "scarlet"
//	start_bc := "!A1!"
//	end_bc := "B6"

//	start_bc := "start"
//	end_bc := "target"

//	start_bc := "maze_a7"
//	end_bc := "maze_i6"

	chapter := ""
	context := []string{""}
	arrowptrs := []SST.ArrowPtr{}
	sttype := []int{1,2,3,0,-1,-2,-3}
	maxdepth := 50
	mindepth := 2

	leftptrs := SST.GetDBNodePtrMatchingName(sst,start_bc,"")
	rightptrs := SST.GetDBNodePtrMatchingName(sst,end_bc,"")

	fmt.Println("Boundary conditions: \nLEFT:",start_bc,leftptrs,"\n\nRIGHT:",end_bc,rightptrs)
	// Contra colliding wavefronts as path integral solver

	solutions := SST.GetPathsAndSymmetries(sst,leftptrs,rightptrs,chapter,context,arrowptrs,sttype,mindepth,maxdepth)

	if len(solutions) > 0 {		
		for s := 0; s < len(solutions); s++ {
			PrintConstrainedLinkPath(sst,solutions,s,"-",chapter,context,arrowptrs,sttype)
		}
	}
}

//******************************************************************

func GetPathsAndSymmetries2(sst SST.PoSST,start_set,end_set []SST.NodePtr,chapter string,context []string,arrowptrs []SST.ArrowPtr,sttypes []int,mindepth,maxdepth int) [][]SST.Link {

	var left_paths, right_paths [][]SST.Link
	var ldepth,rdepth int = 1,1
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

	left_paths,Lnum = SST.GetConstraintConePathsAsLinks(sst,start_set,ldepth,chapter,context,arrowptrs,sttypes,maxdepth)
	right_paths,Rnum = SST.GetConstraintConePathsAsLinks(sst,end_set,rdepth,chapter,context,adj_arrowptrs,adj_sttypes,maxdepth)

	fmt.Println("Constraint primer: \nleft",left_paths,"\n\nright",right_paths)

	// Expand waves

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		fmt.Println("   ..Waves searching",ldepth,rdepth)

		solutions,loop_corrections = SST.WaveFrontsOverlap(sst,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		if len(solutions) > mindepth {
			fmt.Println("   ..DAG solutions:",solutions)
			return solutions
		}

		if len(loop_corrections) > mindepth {
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


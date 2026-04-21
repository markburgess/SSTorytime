// **************************************************************************
//
// matrices.go
//
// **************************************************************************

package SSTorytime

import (
	"fmt"
	_ "github.com/lib/pq"

)


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

func AdjointLinkPath(sst *PoSST,LL []Link) []Link {

	var adjoint []Link

	// len(seq)-1 matches the last node of right join
	// when we invert, links and destinations are shifted

	var prevarrow ArrowPtr = sst.INVERSE_ARROWS[0]

	for j := len(LL)-1; j >= 0; j-- {

		var lnk Link = LL[j]
		lnk.Arr = sst.INVERSE_ARROWS[prevarrow]
		adjoint = append(adjoint,lnk)
		prevarrow = LL[j].Arr
	}

	return adjoint
}

// **************************************************************************

func NextLinkArrow(sst *PoSST,path []Link,arrows []ArrowPtr) string {

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


//
// matrices.go
//



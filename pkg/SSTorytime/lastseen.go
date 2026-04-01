// *********************************************************************
//
// lastseen.go
//
// *********************************************************************

package SSTorytime

import (
	"fmt"
	_ "github.com/lib/pq"

)


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

	qstr := fmt.Sprintf("SELECT section,nptr,EXTRACT(EPOCH FROM first),EXTRACT(EPOCH FROM last),freq,delta as pdelta,EXTRACT(EPOCH FROM NOW()-last) as ndelta from Lastseen ORDER BY section")

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
			var first,last float64
			err = row.Scan(&ls.Section,&nptrstr,&first,&last,&ls.Freq,&ls.Pdelta,&ls.Ndelta)
			ls.Last = int64(last)
			ls.First = int64(first)
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

	qstr := fmt.Sprintf("SELECT section,EXTRACT(EPOCH FROM first),EXTRACT(EPOCH FROM last),freq,delta as pdelta,EXTRACT(EPOCH FROM NOW()-last) as ndelta from Lastseen WHERE NPTR='(%d,%d)'::NodePtr",nptr.Class,nptr.CPtr)

	row,err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetLastSawNPtr failed\n",qstr,err)
		return ls
	}

	if row != nil {
		for row.Next() {
			var first,last float64
			err = row.Scan(&ls.Section,&first,&last,&ls.Freq,&ls.Pdelta,&ls.Ndelta)
			ls.Last = int64(last)
			ls.First = int64(first)
		}

		ls.NPtr = nptr

		row.Close()
	}

	return ls
}

// *********************************************************************

func GetNewlySeenNPtrs(sst PoSST,search SearchParameters) map[NodePtr]bool {

	var qstr string
	var nptrs = make (map[NodePtr]bool)

	switch search.Horizon {

	case RECENT:
		qstr = fmt.Sprintf("SELECT NPtr FROM LastSeen WHERE last > NOW() - INTERVAL '%d hour'",search.Horizon)
	case NEVER:
		qstr = "SELECT NPtr FROM LastSeen"
	default:
		return nptrs
	}

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Failed to get LastSeen",err)
	}

	var whole string
	var nptr NodePtr

	if row != nil {
		for row.Next() {
			err = row.Scan(&whole)
			fmt.Sscanf(whole,"(%d,%d)",&nptr.Class,&nptr.CPtr)
			nptrs[nptr] = true
		}
		
		row.Close()
	}

	return nptrs
}



//
// lastseen.go
//


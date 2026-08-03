// **************************************************************************
//
// cache.go
//
// **************************************************************************

package sst

import (
	"fmt"
	"os"
	"sync"

)

// **************************************************************************

var MUTEX sync.Mutex

// **************************************************************************
//  Node registration and memory management
// **************************************************************************

func GetNodeTxtFromPtr(sst *PoSST,frptr NodePtr) string {

	class := frptr.Class
	index := frptr.CPtr

	var node Node

	switch class {
	case N1GRAM:
		node = sst.NODE_DIRECTORY.N1directory[index]
	case N2GRAM:
		node = sst.NODE_DIRECTORY.N2directory[index]
	case N3GRAM:
		node = sst.NODE_DIRECTORY.N3directory[index]
	case LT128:
		node = sst.NODE_DIRECTORY.LT128directory[index]
	case LT1024:
		node = sst.NODE_DIRECTORY.LT1024[index]
	case GT1024:
		node = sst.NODE_DIRECTORY.GT1024[index]
	}

	return node.S
}

// **************************************************************************

func GetMemoryNodeFromPtr(sst *PoSST,frptr NodePtr) Node {

	class := frptr.Class
	index := frptr.CPtr

	var node Node

	switch class {
	case N1GRAM:
		node = sst.NODE_DIRECTORY.N1directory[index]
	case N2GRAM:
		node = sst.NODE_DIRECTORY.N2directory[index]
	case N3GRAM:
		node = sst.NODE_DIRECTORY.N3directory[index]
	case LT128:
		node = sst.NODE_DIRECTORY.LT128directory[index]
	case LT1024:
		node = sst.NODE_DIRECTORY.LT1024[index]
	case GT1024:
		node = sst.NODE_DIRECTORY.GT1024[index]
	}

	return node
}

// **************************************************************************

func CacheNode(sst *PoSST,n Node) {

	_,already := sst.NODE_CACHE[n.NPtr]

	if !already {
		MUTEX.Lock()
		defer MUTEX.Unlock()
		sst.NODE_CACHE[n.NPtr] = AppendTextToDirectory(sst,n,RunErr)
	}
}


// **************************************************************************

func DownloadArrowsFromDB(sst *PoSST) {

	rows, err := sst.Q.ListArrows(sst.Ctx)
	if err != nil {
		fmt.Println("QUERY Download Arrows Failed", err)
		return
	}

	sst.ARROW_DIRECTORY = nil
	sst.ARROW_DIRECTORY_TOP = 0

	for _, row := range rows {
		ad := ArrowDirectory{
			STAindex: int(row.StaIndex),
			Long:     row.LongName,
			Short:    row.ShortName,
			Ptr:      ArrowPtr(row.ArrPtr),
		}
		sst.ARROW_DIRECTORY = append(sst.ARROW_DIRECTORY, ad)
		sst.ARROW_SHORT_DIR[ad.Short] = sst.ARROW_DIRECTORY_TOP
		sst.ARROW_LONG_DIR[ad.Long] = sst.ARROW_DIRECTORY_TOP
		if ad.Ptr != sst.ARROW_DIRECTORY_TOP {
			fmt.Println(ERR_MEMORY_DB_ARROW_MISMATCH, ad, ad.Ptr, sst.ARROW_DIRECTORY_TOP)
			os.Exit(-1)
		}
		sst.ARROW_DIRECTORY_TOP++
	}

	inv, err := sst.Q.ListArrowInverses(sst.Ctx)
	if err != nil {
		fmt.Println("QUERY Download Inverses Failed", err)
		return
	}
	for _, row := range inv {
		sst.INVERSE_ARROWS[ArrowPtr(row.Plus)] = ArrowPtr(row.Minus)
	}
}

// **************************************************************************

func DownloadContextsFromDB(sst *PoSST) {

	rows, err := sst.Q.ListContexts(sst.Ctx)
	if err != nil {
		fmt.Println("QUERY Download Contexts Failed", err)
		return
	}

	sst.CONTEXT_DIRECTORY = nil
	sst.CONTEXT_TOP = 0

	for _, row := range rows {
		c := ContextDirectory{Context: row.Context, Ptr: ContextPtr(row.CtxPtr)}
		if c.Ptr != sst.CONTEXT_TOP {
			fmt.Println(ERR_MEMORY_DB_CONTEXT_MISMATCH, c, sst.CONTEXT_TOP)
			os.Exit(-1)
		}
		sst.CONTEXT_DIRECTORY = append(sst.CONTEXT_DIRECTORY, c)
		sst.CONTEXT_DIR[c.Context] = sst.CONTEXT_TOP
		sst.CONTEXT_TOP++
	}
}

// **************************************************************************

func SynchronizeNPtrs(sst *PoSST) {
	// Pad in-memory directories so new N4L nodes append after existing DB cptrs.
	for channel := N1GRAM; channel <= GT1024; channel++ {
		maxC, err := sst.Q.MaxCPtrForClass(sst.Ctx, int32(channel))
		if err != nil {
			fmt.Println("QUERY Synchronizing nptrs", err)
			continue
		}
		if maxC < 0 {
			continue
		}
		cptr := int(maxC)
		sst.BASE_DB_CHANNEL_STATE[channel] = ClassedNodePtr(cptr)
		var empty Node
		for n := 0; n <= cptr; n++ {
			switch channel {
			case N1GRAM:
				sst.NODE_DIRECTORY.N1_top++
				sst.NODE_DIRECTORY.N1directory = append(sst.NODE_DIRECTORY.N1directory, empty)
			case N2GRAM:
				sst.NODE_DIRECTORY.N2directory = append(sst.NODE_DIRECTORY.N2directory, empty)
				sst.NODE_DIRECTORY.N2_top++
			case N3GRAM:
				sst.NODE_DIRECTORY.N3directory = append(sst.NODE_DIRECTORY.N3directory, empty)
				sst.NODE_DIRECTORY.N3_top++
			case LT128:
				sst.NODE_DIRECTORY.LT128directory = append(sst.NODE_DIRECTORY.LT128directory, empty)
				sst.NODE_DIRECTORY.LT128_top++
			case LT1024:
				sst.NODE_DIRECTORY.LT1024 = append(sst.NODE_DIRECTORY.LT1024, empty)
				sst.NODE_DIRECTORY.LT1024_top++
			case GT1024:
				sst.NODE_DIRECTORY.GT1024 = append(sst.NODE_DIRECTORY.GT1024, empty)
				sst.NODE_DIRECTORY.GT1024_top++
			}
		}
	}
}

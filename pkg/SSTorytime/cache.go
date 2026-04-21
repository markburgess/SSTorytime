// **************************************************************************
//
// cache.go
//
// **************************************************************************

package SSTorytime

import (
	"fmt"
	"os"
	"sync"
	_ "github.com/lib/pq"

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
		node = sst.NODE_DIRECTORY.LT128[index]
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
		node = sst.NODE_DIRECTORY.LT128[index]
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

	// These must be ordered to match in-memory array

	qstr := fmt.Sprintf("SELECT STAindex,Long,Short,ArrPtr FROM ArrowDirectory ORDER BY ArrPtr")

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY Download Arrows Failed",err)
	}

	sst.ARROW_DIRECTORY = nil
	sst.ARROW_DIRECTORY_TOP = 0

	var staidx int
	var long string
	var short string
	var ptr ArrowPtr
	var ad ArrowDirectory

	if row != nil {
		for row.Next() {		
			err = row.Scan(&staidx,&long,&short,&ptr)
			ad.STAindex = staidx
			ad.Long = long
			ad.Short = short
			ad.Ptr = ptr

			sst.ARROW_DIRECTORY = append(sst.ARROW_DIRECTORY,ad)
			sst.ARROW_SHORT_DIR[short] = sst.ARROW_DIRECTORY_TOP
			sst.ARROW_LONG_DIR[long] = sst.ARROW_DIRECTORY_TOP

			if ad.Ptr != sst.ARROW_DIRECTORY_TOP {
				fmt.Println(ERR_MEMORY_DB_ARROW_MISMATCH,ad,ad.Ptr,sst.ARROW_DIRECTORY_TOP)
				os.Exit(-1)
			}

			sst.ARROW_DIRECTORY_TOP++
		}

		row.Close()
	}

	// Get Inverses

	qstr = fmt.Sprintf("SELECT Plus,Minus FROM ArrowInverses ORDER BY Plus")

	row, err = sst.DB.Query(qstr)
	
	if err != nil {    
		fmt.Println("QUERY Download Inverses Failed",err)
	}

	var plus,minus ArrowPtr

	if row != nil {
		for row.Next() {		

			err = row.Scan(&plus,&minus)

			if err != nil {
				fmt.Println("QUERY Download Arrows Failed",err)
			}

			sst.INVERSE_ARROWS[plus] = minus
		}
		row.Close()
	}
}

// **************************************************************************

func DownloadContextsFromDB(sst *PoSST) {

	qstr := fmt.Sprintf("SELECT Context,CtxPtr FROM ContextDirectory ORDER BY CtxPtr")

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY Download Arrows Failed",err)
	}

	sst.CONTEXT_DIRECTORY = nil
	sst.CONTEXT_TOP = 0

	var context string
	var ptr ContextPtr

	if row != nil {
		for row.Next() {		
			err = row.Scan(&context,&ptr)

			var c ContextDirectory

			c.Context = context
			c.Ptr = ptr

			if c.Ptr != sst.CONTEXT_TOP {
				fmt.Println(ERR_MEMORY_DB_CONTEXT_MISMATCH,c,sst.CONTEXT_TOP)
				os.Exit(-1)
			}

			sst.CONTEXT_DIRECTORY = append(sst.CONTEXT_DIRECTORY,c)
			sst.CONTEXT_DIR[context] = sst.CONTEXT_TOP
			sst.CONTEXT_TOP++
		}

		row.Close()
	}
}

// **************************************************************************

func SynchronizeNPtrs(sst *PoSST) {

	// If we're merging (not recommended) N4L into an existing db, we need to synch

	for channel := N1GRAM; channel <= GT1024; channel++ {
		
		qstr := fmt.Sprintf("SELECT max((Nptr).CPtr) FROM Node WHERE (Nptr).Chan=%d",channel)

		row, err := sst.DB.Query(qstr)
		
		if err != nil {
			fmt.Println("QUERY Synchronizing nptrs",err)
		}

		var cptr int

		if row != nil {
			for row.Next() {			
				err = row.Scan(&cptr)
			
				if err != nil {
					continue // maybe not defined yet
				}

				if cptr > 0 {

					var empty Node

					// Remember this for uploading later ..
					sst.BASE_DB_CHANNEL_STATE[channel] = ClassedNodePtr(cptr)

					for n := 0; n <= cptr; n++ {

						switch channel {
						case N1GRAM:
							sst.NODE_DIRECTORY.N1_top++
							sst.NODE_DIRECTORY.N1directory = append(sst.NODE_DIRECTORY.N1directory,empty)
						case N2GRAM:
							sst.NODE_DIRECTORY.N2directory = append(sst.NODE_DIRECTORY.N2directory,empty)
							sst.NODE_DIRECTORY.N2_top++
						case N3GRAM:
							sst.NODE_DIRECTORY.N3directory = append(sst.NODE_DIRECTORY.N3directory,empty)
							sst.NODE_DIRECTORY.N3_top++
						case LT128:
							sst.NODE_DIRECTORY.LT128 = append(sst.NODE_DIRECTORY.LT128,empty)
							sst.NODE_DIRECTORY.LT128_top++
						case LT1024:
							sst.NODE_DIRECTORY.LT1024 = append(sst.NODE_DIRECTORY.LT1024,empty)
							sst.NODE_DIRECTORY.LT1024_top++
						case GT1024:
							sst.NODE_DIRECTORY.GT1024 = append(sst.NODE_DIRECTORY.GT1024,empty)
							sst.NODE_DIRECTORY.GT1024_top++
						}
					}
				}
			}
			row.Close()
		}
	}

}



//
// cache.go
//


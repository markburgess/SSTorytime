//**************************************************************
//
// db_upload.go
//
//**************************************************************

package SSTorytime

import (
        "fmt"
        "os"
	"strings"
	_ "github.com/lib/pq"

)

//**************************************************************

func GraphToDB(sst PoSST,wait_counter bool) {

	total := len(sst.NODE_DIRECTORY.N1directory) + len(sst.NODE_DIRECTORY.N2directory) + len(sst.NODE_DIRECTORY.N3directory) + len(sst.NODE_DIRECTORY.LT128) + len(sst.NODE_DIRECTORY.LT1024) + len(sst.NODE_DIRECTORY.GT1024) + len(sst.PAGE_MAP)

	fmt.Println("\nStoring primary nodes ...\n")

	for class := N1GRAM; class <= GT1024; class++ {

		offset := int(sst.BASE_DB_CHANNEL_STATE[class])

		switch class {
		case N1GRAM:
			for n := offset; n < len(sst.NODE_DIRECTORY.N1directory); n++ {
				org := sst.NODE_DIRECTORY.N1directory[n]
				UploadNodeToDB(&sst,org)
				Waiting(wait_counter,total)
			}
		case N2GRAM:
			for n := offset; n < len(sst.NODE_DIRECTORY.N2directory); n++ {
				org := sst.NODE_DIRECTORY.N2directory[n]
				UploadNodeToDB(&sst,org)
				Waiting(wait_counter,total)
			}
		case N3GRAM:
			for n := offset; n < len(sst.NODE_DIRECTORY.N3directory); n++ {
				org := sst.NODE_DIRECTORY.N3directory[n]
				UploadNodeToDB(&sst,org)
				Waiting(wait_counter,total)
			}
		case LT128:
			for n := offset; n < len(sst.NODE_DIRECTORY.LT128); n++ {
				org := sst.NODE_DIRECTORY.LT128[n]
				UploadNodeToDB(&sst,org)
				Waiting(wait_counter,total)
			}
		case LT1024:
			for n := offset; n < len(sst.NODE_DIRECTORY.LT1024); n++ {
				org := sst.NODE_DIRECTORY.LT1024[n]
				UploadNodeToDB(&sst,org)
				Waiting(wait_counter,total)
			}

		case GT1024:
			for n := offset; n < len(sst.NODE_DIRECTORY.GT1024); n++ {
				org := sst.NODE_DIRECTORY.GT1024[n]
				UploadNodeToDB(&sst,org)
				Waiting(wait_counter,total)
			}
		}
	}

	// Arrows etc

	fmt.Println("\nStoring Arrows...")

	sst.DB.QueryRow("drop table ArrowDirectory")
	sst.DB.QueryRow("drop table ArrowInverses")

	if !CreateTable(sst,ARROW_INVERSES_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_INVERSES_TABLE)
		os.Exit(-1)
	}
	if !CreateTable(sst,ARROW_DIRECTORY_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_DIRECTORY_TABLE)
		os.Exit(-1)
	}

	for arrow := range sst.ARROW_DIRECTORY {

		UploadArrowToDB(sst,ArrowPtr(arrow))
	}

	fmt.Println("Storing inverse Arrows...")

	for arrow := range sst.INVERSE_ARROWS {

		UploadInverseArrowToDB(sst,ArrowPtr(arrow))
	}

	fmt.Println("Storing contexts...")

	UploadContextsToDB(&sst)

	fmt.Println("Storing page map...")

	for line := 0; line < len(sst.PAGE_MAP); line ++ {
		UploadPageMapEvent(sst,sst.PAGE_MAP[line])
		Waiting(wait_counter,total)
	}

	// CREATE INDICES

	fmt.Println("Indexing ....")

//	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_type on Node (((NPtr).Chan),L,S)")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_gin on Node USING GIN (to_tsvector('english',Search))")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_ungin on Node USING GIN (to_tsvector('english',UnSearch))")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_s on Node USING GIN (S)")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_n on Node USING GIN (NPtr)")
	sst.DB.QueryRow("CREATE INDEX IF NOT EXISTS sst_cnt on ContextDirectory USING GIN (Context)")
	sst.DB.QueryRow("ALTER TABLE Node SET LOGGED")
	sst.DB.QueryRow("ALTER TABLE PageMap SET LOGGED")

	fmt.Println("Finally done!")
}

// **************************************************************************
//  Uploading memory cache to database
// **************************************************************************

func UploadNodeToDB(sst *PoSST, org Node) {

	const nolink = 999

	qstr := "BEGIN;\n" + FormDBNode(sst,org)

	for stindex := 0; stindex < len(org.I); stindex++ {

		lnkarray := FormatSQLLinkArray(org.I[stindex])
		sttype := STIndexToSTType(stindex)
		qstr += AppendDBLinkArrayToNode(sst,org.NPtr,lnkarray,sttype)
	}

	qstr += "\nCOMMIT;"

	row,err := sst.DB.Query(qstr)

	if err != nil {
		s := fmt.Sprint("Failed to insert",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		return
	}

	row.Close()
}

// **************************************************************************

func UploadArrowToDB(sst PoSST,arrow ArrowPtr) {

	staidx := sst.ARROW_DIRECTORY[arrow].STAindex
	long := SQLEscape(sst.ARROW_DIRECTORY[arrow].Long)
	short := SQLEscape(sst.ARROW_DIRECTORY[arrow].Short)

	qstr := fmt.Sprintf("INSERT INTO ArrowDirectory (STAindex,Long,Short,ArrPtr) SELECT %d,'%s','%s',%d WHERE NOT EXISTS (SELECT Long,Short,ArrPtr FROM ArrowDirectory WHERE lower(Long) = lower('%s') OR lower(Short) = lower('%s') OR ArrPtr = %d)",staidx,long,short,arrow,long,short,arrow)

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to insert",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		return
	}

	row.Close()
}

// **************************************************************************

func UploadInverseArrowToDB(sst PoSST,arrow ArrowPtr) {

	plus := arrow
	minus := sst.INVERSE_ARROWS[arrow]

	qstr := fmt.Sprintf("INSERT INTO ArrowInverses (Plus,Minus) SELECT %d,%d WHERE NOT EXISTS (SELECT Plus,Minus FROM ArrowInverses WHERE Plus = %d OR minus = %d)",plus,minus,plus,minus)

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to insert",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		return
	}
	row.Close()
}

// **************************************************************************

func UploadContextsToDB(sst *PoSST) {

	for ctxdir := range sst.CONTEXT_DIRECTORY {
		UploadContextToDB(sst,sst.CONTEXT_DIRECTORY[ctxdir].Context,sst.CONTEXT_DIRECTORY[ctxdir].Ptr)
	}
}

// **************************************************************************

func UploadContextToDB(sst *PoSST,contextstring string,ptr ContextPtr) ContextPtr {

	a := SQLEscape(contextstring)
	b := ptr

	// Make sure neither a nor b are previously defined

	qstr := fmt.Sprintf("SELECT IdempInsertContext('%s',%d)",a,b)

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	var cptr ContextPtr

	if row != nil {
		for row.Next() {
			err = row.Scan(&cptr)
		}
		row.Close()
	}

	return cptr
}

//**************************************************************

func UploadPageMapEvent(sst PoSST, line PageMap) {

	chap := SQLEscape(line.Chapter)

	qstr := "BEGIN;\n"

	qstr += fmt.Sprintf("INSERT INTO PageMap (Chap,Alias,Ctx,Line) VALUES ('%s','%s',%d,%d);\n",chap,line.Alias,line.Context,line.Line)

	lnkarray := FormatSQLLinkArray(line.Path)

	qstr += fmt.Sprintf("\nUPDATE PageMap SET Path='%s' WHERE Chap = '%s' AND Line = '%d';",lnkarray,chap,line.Line)

	qstr += "COMMIT;"

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to insert pagemap event",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		row.Close()
		return
	}

	row.Close()
}


//
// db_upload.go
//



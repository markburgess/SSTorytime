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
	
	fmt.Println("\nStoring primary nodes ...")
	
	for class := N1GRAM; class <= GT1024; class++ {
		
		offset := int(sst.BASE_DB_CHANNEL_STATE[class])
		
		switch class {
		case N1GRAM:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.N1directory[offset:])
		case N2GRAM:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.N2directory[offset:])
		case N3GRAM:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.N3directory[offset:])
		case LT128:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.LT128directory[offset:])
		case LT1024:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.LT1024[offset:])
		case GT1024:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.GT1024[offset:])
		}

	}
	
	// Arrows etc

	fmt.Println("Storing Arrows...")

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

	UploadArrowsToDB(sst)

	fmt.Println("Storing inverse Arrows...")

	UploadInverseArrowsToDB(sst)

	fmt.Println("Storing contexts...")

	UploadContextsToDB(&sst)

	fmt.Println("Storing page map...")

	UploadPageMapBatch(&sst, sst.PAGE_MAP)
	
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
		
	fmt.Println("Done!",total)
}

// **************************************************************************
//  Uploading memory cache to database
// **************************************************************************

func UploadNodesBatch(sst *PoSST, nodes []Node) {

	const chunk = 200

	var qstr string
	
	for i := 0; i < len(nodes); i++ {
	
		if (i % chunk == 0) {
			DBCommit(sst,qstr)
			qstr = ""
		}

		qstr += UploadNodeToDB(sst,nodes[i])
	}

	DBCommit(sst,qstr)
}

// **************************************************************************

func DBCommit(sst *PoSST, qstr string) {

	cstr := "BEGIN;\n"
	cstr += qstr
	cstr += "\nCOMMIT;"
	
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
	return

}

// **************************************************************************

func UploadNodeToDB(sst *PoSST, n Node) string {

	const nodecols  = "(NPtr.Chan,NPtr.Cptr,L,S,Chap,Seq," +
		I_MEXPR + "," + I_MCONT + "," + I_MLEAD + "," + I_NEAR + "," + I_PLEAD + "," + I_PCONT + "," + I_PEXPR + ")"

	n.L, n.NPtr.Class = StorageClass(n.S)

	seq := "false"

	if n.Seq {
		seq = "true"
	}

	// Im3, Im2, Im1, In0, Il1, Ic2, Ie3.

	cols := [7]string{"{}", "{}", "{}", "{}", "{}", "{}", "{}"}

	for stindex := 0; stindex < len(n.I) && stindex < ST_TOP; stindex++ {
		cols[stindex] = FormatSQLLinkArray(n.I[stindex])
	}

	vals := fmt.Sprintf("(%d,%d,%d,'%s','%s',%s,'%s','%s','%s','%s','%s','%s','%s')",
		n.NPtr.Class, n.NPtr.CPtr, n.L,
		SQLEscape(n.S), SQLEscape(n.Chap), seq,
		cols[0], cols[1], cols[2], cols[3], cols[4], cols[5], cols[6])

	
	qstr := "INSERT INTO Node " + nodecols + " VALUES " + vals + ";\n"

	return qstr
}

// **************************************************************************

func UploadArrowsToDB(sst PoSST) {
	
	qstr := "BEGIN;\n"
	
	for arrow := range sst.ARROW_DIRECTORY {
		
		staidx := sst.ARROW_DIRECTORY[arrow].STAindex
		long := SQLEscape(sst.ARROW_DIRECTORY[arrow].Long)
		short := SQLEscape(sst.ARROW_DIRECTORY[arrow].Short)
		
		qstr += fmt.Sprintf("INSERT INTO ArrowDirectory (STAindex,Long,Short,ArrPtr) SELECT %d,'%s','%s',%d WHERE NOT EXISTS (SELECT Long,Short,ArrPtr FROM ArrowDirectory WHERE lower(Long) = lower('%s') OR lower(Short) = lower('%s') OR ArrPtr = %d);\n",staidx,long,short,arrow,long,short,arrow)
		
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

func UploadInverseArrowsToDB(sst PoSST) {

	qstr := "BEGIN;\n"
	
	for arrow := range sst.INVERSE_ARROWS {
		plus := arrow
		minus := sst.INVERSE_ARROWS[arrow]
		
		qstr += fmt.Sprintf("INSERT INTO ArrowInverses (Plus,Minus) SELECT %d,%d WHERE NOT EXISTS (SELECT Plus,Minus FROM ArrowInverses WHERE Plus = %d OR minus = %d);\n",plus,minus,plus,minus)
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

func UploadPageMapBatch(sst *PoSST, lines []PageMap) {

	const chunk = 200

	var qstr string
	
	for i := 0; i < len(lines); i++ {
	
		if (i % chunk == 0) {
			DBCommit(sst,qstr)
			qstr = ""
		}

		line := lines[i]
		qstr += fmt.Sprintf("INSERT INTO PageMap (Chap,Alias,Ctx,Line,Path) VALUES ('%s','%s',%d,%d,'%s');\n",
		SQLEscape(line.Chapter), line.Alias, line.Context, line.Line,
		FormatSQLLinkArray(line.Path))
	}

	DBCommit(sst,qstr)

}


//
// db_upload.go
//



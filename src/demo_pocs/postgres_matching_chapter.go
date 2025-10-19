
//
// Simplest text based set-overlap match test
//

package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

const (
	host     = "localhost"
	port     = 5432
	user     = "sstoryline"
	password = "sst_1234"
	dbname   = "newdb"
)

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)

	qstr := "drop function match_context"

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	} else {
		row.Close()
	}

	qstr = "CREATE OR REPLACE FUNCTION match_context(set1 text[],set2 text[]) RETURNS boolean AS $fn$" +
		"DECLARE "+
		"BEGIN "+
		"  IF set1 && set2 THEN " +
		"     RETURN true;" +
		"  END IF;" +
		"  RETURN false;" +
		"END ;" +
		"$fn$ LANGUAGE plpgsql;"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	row.Close()

	// Show me the nodes in this context

	arr1 := []string{ "yes", "thankyou", "rhyme"}
	set1 := SST.FormatSQLStringArray(arr1)
	chapter := "chinese"
	chapmatch := "%"+chapter+"%"

	// Try matching to nodes in the db
	// qstr = fmt.Sprintf("SELECT match_context(%s,%s)",set1,set2)

	qstr = fmt.Sprintf("WITH matching_nodes AS "+
		"  (SELECT NFrom,ctx,match_context(ctx,%s) AS match FROM NodeArrowNode)"+
		"     SELECT DISTINCT ctx,chap,nfrom,S FROM matching_nodes JOIN Node ON nptr=nfrom  WHERE match=true and chap LIKE '%s'",set1,chapmatch)

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	var a,b,c,d string

	for row.Next() {		
		err = row.Scan(&a,&b,&c,&d)
		fmt.Println("GOT",a,b,c,d)
	}

	row.Close()

	SST.Close(sst)
}

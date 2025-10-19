
//
// Simplest text based set-overlap match test
//

package main

import (
	"fmt"

        SST "SSTorytime"
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

	// Return true/false if the sets overlap

	arr1 := []string{ "one", "two", "three"}
	arr2 := []string{ "four", "five", "x_one"}

	set1 := SST.FormatSQLStringArray(arr1)
	set2 := SST.FormatSQLStringArray(arr2)

	qstr = fmt.Sprintf("SELECT match_context(%s,%s)",set1,set2)

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	var whole string

	for row.Next() {		
		err = row.Scan(&whole)
		fmt.Println("GOT",whole)
	}

	row.Close()



//WITH matching_nodes AS (SELECT match_ctx(ctx,'param') AS match FROM Node)
//   SELECT * FROM matching_nodes WHERE match = true


	SST.Close(sst)
}

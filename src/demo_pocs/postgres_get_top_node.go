
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

	qstr := "drop function ArrowInList"
	row,err := sst.DB.Query(qstr)
	qstr = "drop function GetStoryStartNodes"
	row,err = sst.DB.Query(qstr)

	qstr = "CREATE OR REPLACE FUNCTION ArrowInList(arrow int,links Link[])\n"+
		"RETURNS boolean AS $fn$\n"+
		"DECLARE \n"+
		"   lnk Link;\n"+
		"BEGIN\n"+
		"IF links IS NULL THEN\n"+
		"   RETURN false;"+
		"END IF;"+
		"FOREACH lnk IN ARRAY links LOOP\n"+
		"  IF lnk.Arr = arrow THEN\n"+
		"     RETURN true;\n"+
		"  END IF;\n"+
		"END LOOP;"+
		"RETURN false;"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	row.Close()

	// Find the node that sit's at the start/top of a causal chain

	qstr =  "CREATE OR REPLACE FUNCTION GetStoryStartNodes(arrow int,inverse int,sttype int)\n"+
		"RETURNS NodePtr[] AS $fn$\n"+
		"DECLARE \n"+
		"   retval nodeptr[] = ARRAY[]::nodeptr[];\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	
	for st := -SST.EXPRESS; st <= SST.EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"   SELECT array_agg(Nptr) into retval FROM Node WHERE ArrowInList(arrow,%s) AND NOT ArrowInList(inverse,%s);\n",st,SST.STTypeDBChannel(st),SST.STTypeDBChannel(-st));
	}
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN retval; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	row.Close()

	arrow := "then"

	//matches := SST.GetNodesStartingStoriesForArrow(sst,arrow)

	chapter := ""
	context := []string{"poem"}

	matches := SST.GetNCCNodesStartingStoriesForArrow(sst,arrow,"",chapter,context)

	for p := range matches {

		n := SST.GetDBNodeByNodePtr(sst,matches[p])

		fmt.Println("Story start with",n.S)
	}
	
	SST.Close(sst)
}

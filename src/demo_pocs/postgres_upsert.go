package main

import (
        SST "SSTorytime"
	"fmt"
)


func main() {

	sst := SST.Open(false)

	var qstr string

	sst.DB.QueryRow("drop function lastsawsection(text)")
	sst.DB.QueryRow("drop function lastsawnptr(nodeptr)")

	// ************ LAST SEEN **************'

	qstr = "CREATE OR REPLACE FUNCTION LastSawSection(this text)\n"+
		"RETURNS bool AS $fn$\n"+
		"DECLARE \n"+
		"  prev      timestamp = NOW();\n"+
		"  prevdelta interval;\n"+
		"  deltat    interval;\n"+
		"  nowt      timestamp;\n"+
		"  f    int = 0;"+
		"BEGIN\n"+
		"  nowt = NOW();\n"+
		"  SELECT last,delta,freq INTO prev,prevdelta,f FROM LastSeen WHERE section=this;\n"+
		"  IF NOT FOUND THEN\n"+
		"     INSERT INTO LastSeen (section,last,freq) VALUES (this, NOW(),1);\n"+
		"  ELSE\n"+
		"     deltat = nowt - prev;\n"+
		"     f = f + 1;\n"+
		"     IF deltat > interval '2 minutes' THEN\n"+
		"       UPDATE LastSeen SET last=nowt,delta=deltat,freq=f WHERE section = this;\n"+
		"     END IF;\n"+
		"  END IF;\n"+
		"  RETURN true;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
        // select GetNCNeighboursByType('(1,116)','chinese',-1);
	
	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
	qstr = "CREATE OR REPLACE FUNCTION LastSawNPtr(this NodePtr)\n"+
		"RETURNS bool AS $fn$\n"+
		"DECLARE \n"+
		"  prev      timestamp = NOW();\n"+
		"  prevdelta interval;\n"+
		"  deltat    interval;\n"+
		"  nowt      timestamp;\n"+
		"  f    int = 0;"+
		"BEGIN\n"+
		"  nowt = NOW();\n"+
		"  SELECT last,delta,freq INTO prev,prevdelta,f FROM LastSeen WHERE nptr=this;\n"+
		"  IF NOT FOUND THEN\n"+
		"     INSERT INTO LastSeen (nptr,last,freq) VALUES (this,nowt,1);\n"+
		"  ELSE\n"+
		"     deltat = nowt - prev;\n"+
		"     f = f + 1;\n"+
		"     IF deltat > interval '2 minutes' THEN\n"+
		"        UPDATE LastSeen SET last=nowt,delta=deltat,freq=f WHERE nptr = this;\n"+
		"     END IF;\n"+
		"  END IF;\n"+
		"  RETURN true;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()

	SST.Close(sst)
}


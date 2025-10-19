
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

	var whole1,whole2,whole3,whole4 string
	var retval1,retval2,retval3,retval4 [][]SST.Link

	// Show me the nodes in this context

	qstr := "select AllNCPathsAsLinks('(1,116)','chinese','{}','any',-1)"

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED \n",qstr,err)
	}

	for row.Next() {		
		err = row.Scan(&whole1)
		retval1 = SST.ParseLinkPath(whole1)
	}

	row.Close()

	fmt.Println("GOT EMPTY MATCH for",qstr,retval1)

	// Show me the nodes in this context

	qstr2 := "select AllNCPathsAsLinks('(1,116)','chinese','{trivia}','any',8)"

	row2,err2 := sst.DB.Query(qstr2)
	
	if err2 != nil {
		fmt.Println("FAILED \n",qstr2,err2)
	}

	for row2.Next() {		
		err = row2.Scan(&whole2)
		retval2 = SST.ParseLinkPath(whole2)
	}

	row2.Close()

	fmt.Println("GOT CONTEXT MATCH",qstr2,retval2)

	qstr3 := "select AllNCPathsAsLinks('(1,116)','wrong section','{trivia}','any',8)"

	row3,err3 := sst.DB.Query(qstr3)
	
	if err3 != nil {
		fmt.Println("FAILED \n",qstr3,err3)
	}

	for row3.Next() {		
		err3 = row3.Scan(&whole3)
		retval3 = SST.ParseLinkPath(whole3)
	}

	row3.Close()

	fmt.Println("GOT WRONG CHAPTER MATCH",qstr3,retval3)

	qstr4 := "select AllNCPathsAsLinks('(1,116)','chinese','{wrong,context}','any',8)"

	row4,err4 := sst.DB.Query(qstr4)
	
	if err4 != nil {
		fmt.Println("FAILED \n",qstr4,err4)
	}

	for row4.Next() {		
		err4 = row4.Scan(&whole4)
		retval4 = SST.ParseLinkPath(whole4)
	}

	row4.Close()

	fmt.Println("GOT WRONG CONTEXT MATCH",qstr4,retval4)

	start := SST.GetDBNodePtrMatchingName(sst,"important","chinese")
	a,_ := SST.GetEntireConePathsAsLinks(sst,"any",start[0],4)

	fmt.Println("wrapper call should work (reference)",a)

	chap := "chinese"
	context := []string{"trivia"}

	b,_ := SST.GetEntireNCConePathsAsLinks(sst,"any",start[0],4,chap,context)

	fmt.Println("NC wrapper call should work",b)

	chap = "chinese"
	context = []string{"not work"}

	c,_ := SST.GetEntireNCConePathsAsLinks(sst,"any",start[0],4,chap,context)

	fmt.Println("NC wrapper call should BE EMPTY",c)

	chap = "NOTchinese"
	context = []string{"trivia"}

	d,_ := SST.GetEntireNCConePathsAsLinks(sst,"any",start[0],4,chap,context)

	fmt.Println("NC wrapper call should BE EMPTY",d)

	SST.Close(sst)
}


//
// SQL method
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


// Doesn't compile--remember method


	if remove_name_accents {
		nm_search := "%"+nm_stripped+"%"
		nm_col = fmt.Sprintf("AND lower(unaccent(S)) LIKE lower('%s')",nm_search)
	} else {
		nm_search := "%"+nm+"%"
		nm_col = fmt.Sprintf("AND lower(S) LIKE lower('%s')",nm_search)
	}

	if chap != "any" && chap != "" {

		remove_chap_accents,chap_stripped := IsBracketedSearchTerm(chap)

		if remove_chap_accents {
			chap_search := "%"+chap_stripped+"%"
			chap_col = fmt.Sprintf("AND lower(unaccent(chap)) LIKE lower('%s')",chap_search)
		} else {
			chap_search := "%"+chap+"%"
			chap_col = fmt.Sprintf("AND lower(chap) LIKE lower('%s')",chap_search)
		}
	}

	_,cn_stripped := IsBracketedSearchList(cn)
	context = FormatSQLStringArray(cn_stripped)

	arrows := FormatSQLIntArray(Arrow2Int(arrow))

	qstr = fmt.Sprintf("WITH matching_nodes AS "+
		"  (SELECT NFrom,ctx,match_context(ctx,%s) AS match,match_arrows(Arr,%s) AS matcha FROM NodeArrowNode)"+
		"     SELECT DISTINCT nfrom FROM matching_nodes "+
		"      JOIN Node ON nptr=nfrom WHERE match=true AND matcha=true %s %s",
		context,arrows,nm_col,chap_col)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY GetNodePtrMatchingNCC Failed",err,qstr)
	}

	var whole string
	var n NodePtr
	var retval []NodePtr

	for row.Next() {		
		err = row.Scan(&whole)
		fmt.Sscanf(whole,"(%d,%d)",&n.Class,&n.CPtr)
		retval = append(retval,n)
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

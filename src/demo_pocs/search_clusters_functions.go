//******************************************************************
//
// Exploring how to present a search text, with API
//
// Prepare:
// cd examples
// ../src/N4L-db -u Mary.n4l, e.g. try type Mary example, type 1
//
//******************************************************************

package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)

	context := []string{""}
	//chapter := "double slit"
	//arrow := SST.GetDBArrowByName(sst,"backwards")

	chapter := "maze"
	arrow := SST.GetDBArrowByName(sst,"fwd")
	UseGetAppointmentArrayByArrow(sst,arrow,chapter,context,2)

	chapter = "double slit"
	UseGetAppointmentArrayBySTType(sst,1,chapter,context,2)

	SST.Close(sst)
}

//******************************************************************

func UseGetAppointmentArrayByArrow(sst SST.PoSST,arrow SST.ArrowPtr,chapter string,context []string,min int) {

	var ama map[SST.ArrowPtr][]SST.Appointment

	arr_search := SST.GetDBArrowByPtr(sst,arrow)

	ama = SST.GetAppointedNodesByArrow(sst,arrow,context,chapter,min)

	fmt.Println("--------------------------------------------------")
	fmt.Println("FEATURE: GetAppointmentArrayByArrow:")
	fmt.Println(" return a map of all the nodes in chap,context that are ")
	fmt.Println(" pointed to by at least",min,"arrows of type:",arr_search.Long)
	fmt.Println("--------------------------------------------------")

	for arrowptr := range ama {
		
		arr_dir := SST.GetDBArrowByPtr(sst,arrowptr)
		
		// Appointment list
		for n := 0; n < len(ama[arrowptr]); n++ {

			appointed_nptr := ama[arrowptr][n].NTo
			appointed := SST.GetDBNodeByNodePtr(sst,appointed_nptr)
			
			fmt.Printf("\nAppointed node (%s ...) in chapter \"%s\" correlates/is selected by:\n",appointed.S,chapter)

			// Appointers list
			for m := range ama[arrowptr][n].NFrom {
				node := SST.GetDBNodeByNodePtr(sst,ama[arrowptr][n].NFrom[m])
				stname := SST.STTypeName(SST.STIndexToSTType(arr_dir.STAindex))
				fmt.Printf("     %.40s --(%s : %s)--> %.40s...   - in context %v\n",node.S,arr_dir.Long,stname,appointed.S,context)
			}
		}

		fmt.Println()
		fmt.Println("............................................")
	}
}
//******************************************************************

func UseGetAppointmentArrayBySTType(sst SST.PoSST,sttype int,chapter string,context []string,min int) {

	var ama map[SST.ArrowPtr][]SST.Appointment

	ama = SST.GetAppointedNodesBySTType(sst,sttype,context,chapter,min)

	fmt.Println("--------------------------------------------------")
	fmt.Println("FEATURE: GetAppointmentArrayBySTType:")
	fmt.Println(" return a map of all the nodes in chap,context that are ")
	fmt.Println(" pointed to by at least",min,"arrows of STtype:",sttype)
	fmt.Println("--------------------------------------------------")

	for arrowptr := range ama {
		
		arr_dir := SST.GetDBArrowByPtr(sst,arrowptr)
		
		// Appointment list
		for n := 0; n < len(ama[arrowptr]); n++ {

			appointed_nptr := ama[arrowptr][n].NTo
			appointed := SST.GetDBNodeByNodePtr(sst,appointed_nptr)
			
			fmt.Printf("\nAppointed node (%s ...) in chapter \"%s\" correlates/is selected by:\n",appointed.S,chapter)

			// Appointers list
			for m := range ama[arrowptr][n].NFrom {
				node := SST.GetDBNodeByNodePtr(sst,ama[arrowptr][n].NFrom[m])
				stname := SST.STTypeName(SST.STIndexToSTType(arr_dir.STAindex))
				fmt.Printf("     %.40s --(%s : %s)--> %.40s...   - in context %v\n",node.S,arr_dir.Long,stname,appointed.S,context)
			}
		}

		fmt.Println()
		fmt.Println("............................................")
	}
}


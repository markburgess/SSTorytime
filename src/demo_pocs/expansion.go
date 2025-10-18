
//
// transform random ngrams to graph
//

package main

import (
	"fmt"
        SST "SSTorytime"
)

//**************************************************************

func main() {

	var str = []string{"Dynamic: Time remaining until Christmas ... {TimeUntil Day25 December}",
		"Dynamic: {TimeSince Day25 May Yr2018 Hr18} have elapsed since ChiTek-i was started",
		"Dynamic: {TimeSince Day10 October Yr2016 Hr18} Ldays",
		"Dynamic: Time to regular coordination meeting {TimeUntil Hr11} at 11:00",
		"Dynamic: Time to Monday week start {TimeUntil Monday} Monday Morning"}

	for i,s := range str {
		fmt.Println("\n",i,SST.ExpandDynamicFunctions(s))
	}

	fmt.Println("\n\n")

}


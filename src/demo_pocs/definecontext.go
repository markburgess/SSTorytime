//******************************************************************
//
// Test context registry
//
//******************************************************************

package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := true
	ctx := SST.Open(load_arrows)

	context1 := []string{"giddy","up","horsey"}
	context2 := []string{"get","on","down","pony"}

	newptr1 := SST.TryContext(ctx,context1)
	fmt.Println("defined/found",newptr1)
	newptr2 := SST.TryContext(ctx,context2)
	fmt.Println("defined/found",newptr2)

	str,ptr := SST.GetDBContextByPtr(ctx,newptr1)
	fmt.Println("confirming",ptr,"=",str)

	str,ptr = SST.GetDBContextByPtr(ctx,newptr2)
	fmt.Println("confirming",ptr,"=",str)

	fmt.Println("DIRECTORY CACHE",SST.CONTEXT_DIRECTORY[newptr1])
	fmt.Println("DIRECTORY CACHE",SST.CONTEXT_DIRECTORY[newptr2])

	SST.Close(ctx)	
}






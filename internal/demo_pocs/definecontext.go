package demopocs

import (
	"context"
	"fmt"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

// DefineContext exercises context registry insert/lookup (upstream definecontext).
func DefineContext(ctx context.Context) error {
	sst := SST.Open(ctx, true)
	defer SST.Close(sst)

	context1 := []string{"giddy", "up", "horsey"}
	context2 := []string{"get", "on", "down", "pony"}

	newptr1 := SST.TryContext(ctx, &sst, context1)
	fmt.Println("defined/found", newptr1)
	newptr2 := SST.TryContext(ctx, &sst, context2)
	fmt.Println("defined/found", newptr2)

	str, ptr := SST.GetDBContextByPtr(ctx, &sst, newptr1)
	fmt.Println("confirming", ptr, "=", str)

	str, ptr = SST.GetDBContextByPtr(ctx, &sst, newptr2)
	fmt.Println("confirming", ptr, "=", str)

	fmt.Println("DIRECTORY CACHE", sst.CONTEXT_DIRECTORY[newptr1])
	fmt.Println("DIRECTORY CACHE", sst.CONTEXT_DIRECTORY[newptr2])
	return nil
}

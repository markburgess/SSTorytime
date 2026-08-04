package demopocs

import (
	"context"
	"fmt"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

// DotestGetNodes compares name lookup helpers (upstream dotest_getnodes).
// Prepare: load chinese.n4l (or similar) so node "i6" exists.
func DotestGetNodes(ctx context.Context) error {
	sst := SST.Open(ctx, true)
	defer SST.Close(sst)

	startBC := "i6"
	p1 := SST.GetDBNodePtrMatchingName(ctx, sst, startBC, "")
	p2 := SST.GetDBNodePtrMatchingNCCS(ctx, sst, startBC, "", nil, nil, false, 10)

	if nodePtrSetsDiffer(p1, p2) {
		return fmt.Errorf("%w: %v vs %v", ErrGetNodesMismatch, p1, p2)
	}
	fmt.Println("dotest_getnodes: match ok", p1)
	return nil
}

func nodePtrSetsDiffer(left, right []SST.NodePtr) bool {
	if len(left) != len(right) {
		return true
	}
	for l := 0; l < len(left); l++ {
		if left[l] != right[l] {
			fmt.Println("Mismatch:", left[l], right[l])
			return true
		}
	}
	return false
}

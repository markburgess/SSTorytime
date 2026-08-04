package demopocs

import (
	"context"
	"fmt"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

// DotestEntireCone compares NC cone vs plain cone retrieval (upstream dotest_entirecone).
// Prepare: maze-style data (e.g. examples run api-3) so a7 / i6 exist.
func DotestEntireCone(ctx context.Context) error {
	sst := SST.Open(ctx, true)
	defer SST.Close(sst)

	const maxdepth = 16
	ldepth, rdepth := 1, 1
	startBC := "a7"
	endBC := "i6"

	leftptrs := SST.GetDBNodePtrMatchingName(ctx, sst, startBC, "")
	rightptrs := SST.GetDBNodePtrMatchingName(ctx, sst, endBC, "")
	if leftptrs == nil || rightptrs == nil {
		return fmt.Errorf("%w: %q / %q", ErrEntireConeNoEnds, startBC, endBC)
	}

	cntx := []string{""}
	const limit = 30

	for ldepth < maxdepth && rdepth < maxdepth {
		left_paths, Lnum := SST.GetEntireNCConePathsAsLinks(ctx, &sst, "fwd", leftptrs, ldepth, "", cntx, limit)
		xleft_paths, Lnumx := SST.GetEntireConePathsAsLinks(ctx, &sst, "fwd", leftptrs[0], ldepth, limit)

		right_paths, Rnum := SST.GetEntireNCConePathsAsLinks(ctx, &sst, "bwd", rightptrs, rdepth, "", cntx, limit)
		xright_paths, Rnumx := SST.GetEntireConePathsAsLinks(ctx, &sst, "bwd", rightptrs[0], rdepth, limit)

		if Lnum != Lnumx {
			return fmt.Errorf("%w: depth %d sizes %d vs %d\nEntireNCcone %v\nEntirecone %v",
				ErrEntireConeLeftSz, ldepth, Lnum, Lnumx, left_paths, xleft_paths)
		}
		if linkPathSetsDiffer(left_paths, xleft_paths) {
			return fmt.Errorf("%w: depth %d\nEntireNCcone %v\nEntirecone %v",
				ErrEntireConeLeftSet, ldepth, left_paths, xleft_paths)
		}
		if Rnum != Rnumx {
			return fmt.Errorf("%w: depth %d sizes %d vs %d", ErrEntireConeRightSz, rdepth, Rnum, Rnumx)
		}
		if linkPathSetsDiffer(right_paths, xright_paths) {
			return fmt.Errorf("%w: depth %d", ErrEntireConeRightSet, rdepth)
		}
		ldepth++
		rdepth++
	}
	fmt.Println("dotest_entirecone: NC vs plain cone match through depth", maxdepth-1)
	return nil
}

func linkPathSetsDiffer(left, right [][]SST.Link) bool {
	L := make(map[SST.Link]bool)
	R := make(map[SST.Link]bool)
	retval := false

	for path := 0; path < len(left); path++ {
		for l := 0; l < len(left[path]); l++ {
			L[left[path][l]] = true
		}
	}
	for path := 0; path < len(right); path++ {
		for l := 0; l < len(right[path]); l++ {
			R[right[path][l]] = true
		}
	}
	for r := range L {
		if !R[r] {
			fmt.Println("L not in R", r)
			retval = true
		}
	}
	for l := range R {
		if !L[l] {
			fmt.Println("R not in L", l)
			retval = true
		}
	}
	return retval
}

package sst

import (
	"fmt"

	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

func UpdateLastSawSection(sst PoSST, name string) {
	if sst.Q == nil {
		return
	}
	if err := sst.Q.CallLastSawSection(sst.ctx(), name); err != nil {
		fmt.Println("LastSawSection failed", err)
	}
}

func UpdateLastSawNPtr(sst PoSST, class, cptr int, name string) {
	if sst.Q == nil {
		return
	}
	if err := sst.Q.CallLastSawNPtr(sst.ctx(), sqlc.CallLastSawNPtrParams{
		Column1: int32(class),
		Column2: int32(cptr),
		Name:    name,
	}); err != nil {
		fmt.Println("LastSawNPtr failed", err)
	}
}

func GetLastSawSection(sst PoSST) []LastSeen {
	if sst.Q == nil {
		return nil
	}
	rows, err := sst.Q.ListLastSeen(sst.ctx())
	if err != nil {
		fmt.Println("GetLastSawSection failed", err)
		return nil
	}
	ret := make([]LastSeen, 0, len(rows))
	for _, row := range rows {
		var ls LastSeen
		ls.Section = derefStr(row.Section)
		ls.NPtr.Class = int(row.Chan)
		ls.NPtr.CPtr = ClassedNodePtr(row.Cptr)
		ls.First = int64(row.FirstEpoch)
		ls.Last = int64(row.LastEpoch)
		if row.Delta != nil {
			ls.Pdelta = float64(*row.Delta)
		}
		if row.Freq != nil {
			ls.Freq = int(*row.Freq)
		}
		ls.Ndelta = row.Ndelta
		ret = append(ret, ls)
	}
	for c := 0; c < len(ret); c++ {
		ret[c].XYZ = AssignChapterCoordinates(c, len(ret))
	}
	return ret
}

func GetLastSawNPtr(sst PoSST, nptr NodePtr) LastSeen {
	var ls LastSeen
	ls.NPtr = nptr
	if sst.Q == nil {
		return ls
	}
	row, err := sst.Q.GetLastSeenByNPtr(sst.ctx(), sqlc.GetLastSeenByNPtrParams{
		Column1: int32(nptr.Class),
		Column2: int32(nptr.CPtr),
	})
	if err != nil {
		return ls
	}
	ls.Section = derefStr(row.Section)
	ls.First = int64(row.FirstEpoch)
	ls.Last = int64(row.LastEpoch)
	if row.Delta != nil {
		ls.Pdelta = float64(*row.Delta)
	}
	if row.Freq != nil {
		ls.Freq = int(*row.Freq)
	}
	ls.Ndelta = row.Ndelta
	return ls
}

func GetNewlySeenNPtrs(sst PoSST, search SearchParameters) map[NodePtr]bool {
	nptrs := make(map[NodePtr]bool)
	if sst.Q == nil {
		return nptrs
	}
	var (
		rows []struct {
			Chan int32
			Cptr int32
		}
		err error
	)
	switch search.Horizon {
	case RECENT:
		r, e := sst.Q.ListRecentLastSeenNPtrs(sst.ctx(), int32(search.Horizon))
		err = e
		for _, x := range r {
			rows = append(rows, struct {
				Chan int32
				Cptr int32
			}{x.Chan, x.Cptr})
		}
	case NEVER:
		r, e := sst.Q.ListAllLastSeenNPtrs(sst.ctx())
		err = e
		for _, x := range r {
			rows = append(rows, struct {
				Chan int32
				Cptr int32
			}{x.Chan, x.Cptr})
		}
	default:
		return nptrs
	}
	if err != nil {
		fmt.Println("Failed to get LastSeen", err)
		return nptrs
	}
	for _, r := range rows {
		nptrs[NodePtr{Class: int(r.Chan), CPtr: ClassedNodePtr(r.Cptr)}] = true
	}
	return nptrs
}

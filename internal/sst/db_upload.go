package sst

import (
	"fmt"

	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

func GraphToDB(sst PoSST, wait_counter bool) {
	fmt.Println(".  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ")
	fmt.Println("\nStoring primary nodes ...")

	for class := N1GRAM; class <= GT1024; class++ {
		offset := int(sst.BASE_DB_CHANNEL_STATE[class])
		switch class {
		case N1GRAM:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.N1directory[offset:])
		case N2GRAM:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.N2directory[offset:])
		case N3GRAM:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.N3directory[offset:])
		case LT128:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.LT128directory[offset:])
		case LT1024:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.LT1024[offset:])
		case GT1024:
			UploadNodesBatch(&sst, sst.NODE_DIRECTORY.GT1024[offset:])
		}
	}

	fmt.Println(".  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ")
	fmt.Println("Storing Arrows...")
	UploadArrowsToDB(sst)
	fmt.Println("Storing inverse Arrows...")
	UploadInverseArrowsToDB(sst)
	fmt.Println("Storing contexts...")
	UploadContextsToDB(&sst)
	fmt.Println("Storing page map...")
	UploadPageMapBatch(&sst, sst.PAGE_MAP)
	fmt.Println(".  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ")
	fmt.Println("Done storing graph.")
}

func BookmarksToDB(sst PoSST, marks map[string]string) {
	ctx := sst.Ctx
	for b, q := range marks {
		_ = sst.Q.InsertBookmark(ctx, sqlc.InsertBookmarkParams{
			Bookmark: b,
			Query:    q,
		})
	}
}

func UploadNodesBatch(sst *PoSST, nodes []Node) {
	for i := range nodes {
		UploadNodeToDB(sst, nodes[i])
	}
}

func DBCommit(sst *PoSST, qstr string) {}

func UploadNodeToDB(sst *PoSST, n Node) string {
	ctx := sst.Ctx
	_ = sst.Q.InsertNode(ctx, sqlc.InsertNodeParams{
		Class: int32(n.NPtr.Class),
		Cptr:  int32(n.NPtr.CPtr),
		L:     int32(n.L),
		S:     n.S,
		Chap:  n.Chap,
		Seq:   n.Seq,
	})

	for stindex := 0; stindex < len(n.I) && stindex < ST_TOP; stindex++ {
		sttype := STIndexToSTType(stindex)
		for _, lnk := range n.I[stindex] {
			if lnk.Arr == 0 && lnk.Dst.CPtr == 0 && lnk.Dst.Class == 0 {
				continue
			}
			_ = sst.Q.InsertEdge(ctx, sqlc.InsertEdgeParams{
				SrcClass: int32(n.NPtr.Class),
				SrcCptr:  int32(n.NPtr.CPtr),
				DstClass: int32(lnk.Dst.Class),
				DstCptr:  int32(lnk.Dst.CPtr),
				Arr:      int32(lnk.Arr),
				Wgt:      lnk.Wgt,
				Ctx:      int32(lnk.Ctx),
				St:       StTypeFromInt(sttype),
			})
		}
	}
	return ""
}

func UploadArrowsToDB(sst PoSST) {
	ctx := sst.Ctx
	for _, ad := range sst.ARROW_DIRECTORY {
		_ = sst.Q.InsertArrow(ctx, sqlc.InsertArrowParams{
			ArrPtr:    int32(ad.Ptr),
			StaIndex:  int32(ad.STAindex),
			LongName:  ad.Long,
			ShortName: ad.Short,
		})
	}
}

func UploadInverseArrowsToDB(sst PoSST) {
	ctx := sst.Ctx
	for plus, minus := range sst.INVERSE_ARROWS {
		_ = sst.Q.InsertArrowInverse(ctx, sqlc.InsertArrowInverseParams{
			Plus:  int32(plus),
			Minus: int32(minus),
		})
	}
}

func UploadContextsToDB(sst *PoSST) {
	for _, c := range sst.CONTEXT_DIRECTORY {
		UploadContextToDB(sst, c.Context, c.Ptr)
	}
}

func UploadContextToDB(sst *PoSST, contextstring string, ptr ContextPtr) ContextPtr {
	ctx := sst.Ctx
	_ = sst.Q.InsertContext(ctx, sqlc.InsertContextParams{
		CtxPtr:  int32(ptr),
		Context: contextstring,
	})
	return ptr
}

func UploadPageMapBatch(sst *PoSST, lines []PageMap) {
	ctx := sst.Ctx
	for _, line := range lines {
		// path as empty json for now; full link path serialization later
		_ = sst.Q.InsertPageMap(ctx, sqlc.InsertPageMapParams{
			Chap:  line.Chapter,
			Alias: line.Alias,
			Ctx:   int32(line.Context),
			Line:  int32(line.Line),
			Path:  []byte("[]"),
		})
	}
}

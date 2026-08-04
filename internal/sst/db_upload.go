//**************************************************************
//
// db_upload.go
//
//**************************************************************

package sst

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

//**************************************************************

func GraphToDB(ctx context.Context, sst PoSST, wait_counter bool) {

	fmt.Println(".  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ")
	fmt.Println("\nStoring primary nodes ...")

	for class := N1GRAM; class <= GT1024; class++ {

		offset := int(sst.BASE_DB_CHANNEL_STATE[class])

		switch class {
		case N1GRAM:
			UploadNodesBatch(ctx, &sst, sst.NODE_DIRECTORY.N1directory[offset:])
		case N2GRAM:
			UploadNodesBatch(ctx, &sst, sst.NODE_DIRECTORY.N2directory[offset:])
		case N3GRAM:
			UploadNodesBatch(ctx, &sst, sst.NODE_DIRECTORY.N3directory[offset:])
		case LT128:
			UploadNodesBatch(ctx, &sst, sst.NODE_DIRECTORY.LT128directory[offset:])
		case LT1024:
			UploadNodesBatch(ctx, &sst, sst.NODE_DIRECTORY.LT1024[offset:])
		case GT1024:
			UploadNodesBatch(ctx, &sst, sst.NODE_DIRECTORY.GT1024[offset:])
		}

	}

	// Arrows etc

	fmt.Println(".  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ")
	fmt.Println("Storing Arrows...")

	if sst.Q != nil {
		if err := sst.Q.TruncateArrowDirectory(ctx); err != nil {
			fmt.Println("truncate arrowdirectory:", err)
		}
		if err := sst.Q.TruncateArrowInverses(ctx); err != nil {
			fmt.Println("truncate arrowinverses:", err)
		}
	}

	UploadArrowsToDB(ctx, sst)

	fmt.Println("Storing inverse Arrows...")

	UploadInverseArrowsToDB(ctx, sst)

	fmt.Println("Storing contexts...")

	UploadContextsToDB(ctx, &sst)

	fmt.Println("Storing page map...")

	UploadPageMapBatch(ctx, &sst, sst.PAGE_MAP)

	// CREATE INDICES

	fmt.Println(".  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  ")
	fmt.Println("Indexing ....")

	Waiting()

	if sst.Q != nil {
		for name, fn := range map[string]func() error{
			"sst_gin":        func() error { return sst.Q.CreateNodeIndexes(ctx) },
			"sst_ungin":      func() error { return sst.Q.CreateNodeIndexes2(ctx) },
			"sst_s":          func() error { return sst.Q.CreateNodeIndexes3(ctx) },
			"sst_n":          func() error { return sst.Q.CreateNodeIndexes4(ctx) },
			"sst_cnt":        func() error { return sst.Q.CreateContextIndex(ctx) },
			"node logged":    func() error { return sst.Q.AlterNodeLogged(ctx) },
			"pagemap logged": func() error { return sst.Q.AlterPageMapLogged(ctx) },
		} {
			if err := fn(); err != nil {
				fmt.Println("index/alter", name, ":", err)
			}
		}
	}

}

// **************************************************************************

func BookmarksToDB(ctx context.Context, sst PoSST, marks map[string]string) {

	if sst.Q == nil {
		return
	}
	for b, q := range marks {
		if err := sst.Q.InsertBookmark(ctx, sqlc.InsertBookmarkParams{
			Bookmark: strPtr(b),
			Query:    strPtr(q),
		}); err != nil && !isUniqueViolation(err) {
			fmt.Println("Failed to insert bookmark", err)
		}
	}
}

// **************************************************************************
//  Uploading memory cache to database
// **************************************************************************

func UploadNodesBatch(ctx context.Context, sst *PoSST, nodes []Node) {

	for i := 0; i < len(nodes); i++ {
		if err := UploadNodeToDB(ctx, sst, nodes[i]); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			fmt.Println("Failed to insert", err, "FAILED")
		}
	}
}

// **************************************************************************

// UploadNodeToDB inserts one in-memory node (with link arrays) via sqlc.
func UploadNodeToDB(ctx context.Context, sst *PoSST, n Node) error {
	if sst.Q == nil {
		return ErrNoQuerier
	}
	cols := [7]string{"{}", "{}", "{}", "{}", "{}", "{}", "{}"}
	for stindex := 0; stindex < len(n.I) && stindex < ST_TOP; stindex++ {
		cols[stindex] = FormatSQLLinkArray(n.I[stindex])
	}
	s := n.S
	chap := n.Chap
	return sst.Q.InsertNodeRow(ctx, sqlc.InsertNodeRowParams{
		Column1:  int32(n.NPtr.Class),
		Column2:  int32(n.NPtr.CPtr),
		Column3:  int32(n.L),
		S:        &s,
		Chap:     &chap,
		Column6:  n.Seq,
		Column7:  cols[0],
		Column8:  cols[1],
		Column9:  cols[2],
		Column10: cols[3],
		Column11: cols[4],
		Column12: cols[5],
		Column13: cols[6],
	})
}

// **************************************************************************

func UploadArrowsToDB(ctx context.Context, sst PoSST) {

	if sst.Q == nil {
		return
	}
	for arrow := range sst.ARROW_DIRECTORY {
		staidx := int32(sst.ARROW_DIRECTORY[arrow].STAindex)
		long := sst.ARROW_DIRECTORY[arrow].Long
		short := sst.ARROW_DIRECTORY[arrow].Short
		if err := sst.Q.InsertArrowDirectory(ctx, sqlc.InsertArrowDirectoryParams{
			Staindex: &staidx,
			Long:     &long,
			Short:    &short,
			Arrptr:   int32(arrow),
		}); err != nil && !isUniqueViolation(err) {
			fmt.Println("Failed to insert arrow", err)
		}
	}
}

// **************************************************************************

func UploadInverseArrowsToDB(ctx context.Context, sst PoSST) {

	if sst.Q == nil {
		return
	}
	for arrow := range sst.INVERSE_ARROWS {
		plus := int32(arrow)
		minus := int32(sst.INVERSE_ARROWS[arrow])
		if err := sst.Q.InsertArrowInverse(ctx, sqlc.InsertArrowInverseParams{
			Plus:  plus,
			Minus: minus,
		}); err != nil && !isUniqueViolation(err) {
			fmt.Println("Failed to insert inverse", err)
		}
	}
}

// **************************************************************************

func UploadContextsToDB(ctx context.Context, sst *PoSST) {

	for ctxdir := range sst.CONTEXT_DIRECTORY {
		UploadContextToDB(ctx, sst, sst.CONTEXT_DIRECTORY[ctxdir].Context, sst.CONTEXT_DIRECTORY[ctxdir].Ptr)
	}
}

// **************************************************************************

func UploadContextToDB(ctx context.Context, sst *PoSST, contextstring string, ptr ContextPtr) ContextPtr {

	if sst.Q == nil {
		return ptr
	}

	cptr, err := sst.Q.IdempInsertContext(ctx, sqlc.IdempInsertContextParams{
		Constr: contextstring,
		Conptr: int32(ptr),
	})
	if err != nil {
		fmt.Println("FAILED IdempInsertContext", err)
		return ptr
	}
	return ContextPtr(cptr)
}

//**************************************************************

func UploadPageMapBatch(ctx context.Context, sst *PoSST, lines []PageMap) {

	if sst.Q == nil {
		return
	}
	for _, line := range lines {
		path := FormatSQLLinkArray(line.Path)
		if err := sst.Q.InsertPageMap(ctx, sqlc.InsertPageMapParams{
			Chap:    strPtr(line.Chapter),
			Alias:   strPtr(line.Alias),
			Column3: int32(line.Context),
			Column4: int32(line.Line),
			Column5: path,
		}); err != nil && !isUniqueViolation(err) {
			fmt.Println("Failed to insert pagemap", err)
		}
	}
}

// isUniqueViolation reports Postgres unique_violation (23505) or common duplicate wording.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	// Fallback when drivers wrap without PgError
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}

//
// db_upload.go
//

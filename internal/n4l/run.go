package n4l

import (
	"context"
	"fmt"
	"os"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

// Options for the N4L compiler.
type Options struct {
	Files     []string
	Upload    bool
	Force     bool
	Wipe      bool
	Verbose   bool
	Diag      bool
	Summary   bool
	AdjList   string // empty = no adjacency matrix
	DatabaseURL string
}

// Run compiles N4L files and optionally uploads to the database.
func Run(ctx context.Context, opt Options) error {
	if len(opt.Files) < 1 {
		return fmt.Errorf("n4l: at least one input file required")
	}

	VERBOSE = opt.Verbose
	DIAGNOSTIC = opt.Diag
	UPLOAD = opt.Upload
	FORCE_UPLOAD = opt.Force
	SUMMARIZE = opt.Summary
	if opt.Diag {
		VERBOSE = true
	}
	if opt.AdjList != "" && opt.AdjList != "none" {
		CREATE_ADJACENCY = true
		ADJ_LIST = opt.AdjList
	}
	if opt.Wipe {
		SST.WIPE_DB = true
	}

	var sst SST.PoSST
	if UPLOAD || opt.Wipe {
		// DB required for upload / wipe (migrations + sqlc).
		sst = SST.OpenWithDSN(ctx, opt.DatabaseURL)
	} else {
		// Parse-only: in-memory graph, no Postgres.
		sst.Ctx = ctx
		SST.MemoryInit(&sst)
	}
	AddMandatory(&sst)

	config := ReadConfig()
	CONFIGURING = true
	for input := 0; input < len(config); input++ {
		NewFile(config[input])
		con := ReadFile(CURRENT_FILE)
		ParseConfig(&sst, con)
	}
	CONFIGURING = false

	for input := 0; input < len(opt.Files); input++ {
		NewFile(opt.Files[input])
		src := ReadFile(CURRENT_FILE)
		ParseN4L(&sst, src)
	}

	CompleteInferences(&sst)

	if SUMMARIZE {
		SummarizeGraph(sst)
	}
	if CREATE_ADJACENCY {
		dim, key, d_adj, u_adj := CreateAdjacencyMatrix(sst, ADJ_LIST)
		PrintMatrix(sst, "directed adjacency sub-matrix", dim, key, d_adj)
		PrintMatrix(sst, "undirected adjacency sub-matrix", dim, key, u_adj)
		evc := ComputeEVC(dim, u_adj)
		PrintNZVector(sst, "Eigenvector centrality (EVC) score for symmetrized graph", dim, key, evc)
	}
	if UPLOAD {
		if err := upload(sst); err != nil {
			SST.Close(sst)
			return err
		}
	}
	SST.Close(sst)
	return nil
}

func upload(sst SST.PoSST) error {
	dbchapters := SST.GetDBChaptersMatchingName(sst, "")
	memchapters := GetMemChapters(sst)
	conflict := false
	for m := range memchapters {
		for d := range dbchapters {
			if memchapters[m] == dbchapters[d] {
				fmt.Println(" Database already contains a chapter: ", dbchapters[d])
				conflict = true
			}
		}
	}
	if conflict && !FORCE_UPLOAD {
		fmt.Println("\nUploading to a pre-existing chapter might corrupt the data. You can remove it first with remove or force using --force.")
		return fmt.Errorf("chapter conflict; use --force to override")
	}
	fmt.Println("\n\nUploading nodes..")
	SST.GraphToDB(sst, true)
	marks := GetBookMarks()
	fmt.Println("\n\nUploading bookmarks..")
	SST.BookmarksToDB(sst, marks)
	return nil
}

// ensure unused imports stay quiet in compiler.go after main rename
var _ = os.Args

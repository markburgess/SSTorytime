package sst

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/markburgess/SSTorytime/internal/db"
)

// Open connects via internal/db (sync.Once + migrate on first load).
// load_arrows is kept for API compatibility; arrows are always loaded from DB when present.
func Open(load_arrows bool) PoSST {
	return OpenWithDSN(context.Background(), "")
}

// OpenWithDSN is like Open but accepts an explicit DSN (CLI --database-url).
func OpenWithDSN(ctx context.Context, dsn string) PoSST {
	var sst PoSST
	sst.Ctx = ctx

	if WIPE_DB {
		if err := wipeDatabase(ctx, dsn); err != nil {
			fmt.Println("wipe database:", err)
			os.Exit(-1)
		}
	}

	pool, q, err := db.OpenDSN(ctx, dsn)
	if err != nil {
		fmt.Println("Error connecting to the database: ", err)
		os.Exit(-1)
	}

	sst.Pool = pool
	sst.Q = q
	sst.DB = stdlib.OpenDBFromPool(pool)

	MemoryInit(&sst)

	DownloadArrowsFromDB(&sst)
	DownloadContextsFromDB(&sst)
	SynchronizeNPtrs(&sst)

	NO_NODE_PTR.Class = 0
	NO_NODE_PTR.CPtr = -1
	NONODE.Class = 0
	NONODE.CPtr = 0

	return sst
}

func wipeDatabase(ctx context.Context, dsn string) error {
	// Truncate after migrate so schema exists.
	pool, q, err := db.OpenDSN(ctx, dsn)
	if err != nil {
		return err
	}
	_ = q
	tables := []string{
		"edge", "page_map", "bookmarks", "last_seen",
		"node", "arrow_inverses", "arrow_directory", "context_directory",
	}
	for _, t := range tables {
		if _, err := pool.Exec(ctx, "TRUNCATE TABLE "+t+" CASCADE"); err != nil {
			// ignore missing on first boot
			_ = err
		}
	}
	fmt.Println("***********************")
	fmt.Println("* WIPED DB TABLES")
	fmt.Println("***********************")
	return nil
}

// OverrideCredentials reads ~/.SSTorytime (legacy). Prefer DATABASE_URL / --database-url.
func OverrideCredentials(u, p, d string) (string, string, string) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return u, p, d
	}
	filename := filepath.Join(dirname, CREDENTIALS_FILE)
	content, err := os.ReadFile(filename)
	if err != nil {
		return u, p, d
	}
	user, password, dbname := u, p, d
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)
		switch key {
		case "user":
			user = val
		case "passwd", "password":
			password = val
		case "db", "dbname":
			dbname = val
		}
	}
	return user, password, dbname
}

func MemoryInit(sst *PoSST) {
	if sst.NODE_DIRECTORY.N1grams == nil {
		sst.NODE_DIRECTORY.N1grams = make(map[string]ClassedNodePtr)
	}
	if sst.NODE_DIRECTORY.N2grams == nil {
		sst.NODE_DIRECTORY.N2grams = make(map[string]ClassedNodePtr)
	}
	if sst.NODE_DIRECTORY.N3grams == nil {
		sst.NODE_DIRECTORY.N3grams = make(map[string]ClassedNodePtr)
	}
	if sst.NODE_DIRECTORY.LT128 == nil {
		sst.NODE_DIRECTORY.LT128 = make(map[string]ClassedNodePtr)
	}
	sst.NODE_CACHE = make(map[NodePtr]NodePtr)
	sst.INVERSE_ARROWS = make(map[ArrowPtr]ArrowPtr)
	sst.ARROW_SHORT_DIR = make(map[string]ArrowPtr)
	sst.ARROW_LONG_DIR = make(map[string]ArrowPtr)
	sst.ARROW_DIRECTORY_TOP = 0
	sst.CONTEXT_DIR = make(map[string]ContextPtr)
}

func Close(sst PoSST) {
	if sst.DB != nil {
		_ = sst.DB.Close()
	}
	// pool is shared (sync.Once); do not close process-wide pool here
}

package sst

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/markburgess/SSTorytime/internal/db"
)

//**************************************************************
//
// session.go
//
//**************************************************************

// Open migrates and opens a session using the default DSN resolution.
// ctx must be non-nil (callers pass cmd.Context() or r.Context()).
func Open(ctx context.Context, load_arrows bool) PoSST {
	return OpenWithDSN(ctx, "", load_arrows)
}

// OpenWithDSN migrates schema/functions then loads arrow/context caches.
// load_arrows is kept for API compatibility (arrows always loaded from DB when present).
// ctx must be non-nil.
func OpenWithDSN(ctx context.Context, dsn string, load_arrows bool) PoSST {
	if ctx == nil {
		panic("sst: OpenWithDSN called with nil context")
	}
	var sst PoSST

	if WIPE_DB {
		// Full reset: drop dirty migrate state by truncating app data after migrate.
		// Schema comes only from migrations; PL/pgSQL is in 000002.
	}

	pool, q, err := db.OpenDSN(ctx, dsn)
	if err != nil {
		fmt.Println("Error connecting/migrating the database: ", err)
		os.Exit(-1)
	}
	sst.Pool = pool
	sst.Q = q

	if WIPE_DB {
		if err := q.TruncateAllData(ctx); err != nil {
			fmt.Println("wipe truncate:", err)
			os.Exit(-1)
		}
		fmt.Println("***********************")
		fmt.Println("* WIPED DB DATA")
		fmt.Println("***********************")
		WIPE_DB = false
	}

	MemoryInit(&sst)
	// Schema+functions: migrations only. No runtime DefineStoredFunctions.
	_ = load_arrows

	DownloadArrowsFromDB(ctx, &sst)
	DownloadContextsFromDB(ctx, &sst)
	SynchronizeNPtrs(ctx, &sst)

	NO_NODE_PTR.Class = 0
	NO_NODE_PTR.CPtr = -1
	NONODE.Class = 0
	NONODE.CPtr = 0

	return sst
}

// **************************************************************************

func OverrideCredentials(u, p, d string) (string, string, string) {

	// Store database/postgres credentials in a system file instead of hardcoding

	dirname, err := os.UserHomeDir()

	if err != nil && len(dirname) > 1 {
		fmt.Println("Unable to determine user's home directory")
		os.Exit(-1)
	}

	filename := dirname + "/" + CREDENTIALS_FILE
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		return u, p, d
	}

	/* format
	   dbname: sstoryline
	   user:sstoryline
	   passwd: sst_1234
	*/

	var (
		offset, delta int
		user          = u
		password      = p
		dbname        = d
	)

	for offset = 0; offset < len(content); offset = offset {

		var conf string
		fmt.Sscanf(string(content[offset:]), "%s", &conf)

		if len(conf) > 0 && conf[len(conf)-1] != ':' { // missing space

			for delta = 0; delta < len(conf); delta++ {
				if conf[delta] == ':' {
					conf = conf[:delta+1]
				}
			}
		}

		switch conf {
		case "user:":
			delta = len(conf)
			user, offset = GetLine(content, offset+delta)
		case "passwd:", "password:":
			delta = len(conf)
			password, offset = GetLine(content, offset+delta)
		case "db:", "dbname:":
			delta = len(conf)
			dbname, offset = GetLine(content, offset+delta)
		default:
			offset++
		}
	}

	return user, password, dbname
}

// **************************************************************************

func GetLine(s []byte, i int) (string, int) {

	// For parsing the password credential file

	var result []byte

	for o := i; o < len(s); o++ {

		if s[o] == '\n' {
			i = o
			break
		}

		result = append(result, s[o])
	}

	return string(result), i
}

// **************************************************************************

func MemoryInit(sst *PoSST) {

	//  When opening a connection, restore config and allocate maps

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

// **************************************************************************

func Configure(sst PoSST, load_arrows bool) {
	// No-op: schema and PL/pgSQL come from golang-migrate (000001 + 000002).
	_ = sst
	_ = load_arrows
}

func Close(sst PoSST) {
	// Shared process pool; sessions do not close it.
}

//
// session.go
//

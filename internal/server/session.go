package server

import (
	"context"
	"sync"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

// Shared session for HTTP handlers. Opening a full session per request reloads
// arrow/context caches from Postgres and races package globals; under load or a
// slow DB that looks like a hang in the browser.
var (
	serveSSTOnce sync.Once
	serveSST     SST.PoSST
)

// WarmSession opens the process-wide graph session (migrate + caches).
// Call once at serve startup after the DB is reachable.
func WarmSession(ctx context.Context) {
	serveSSTOnce.Do(func() {
		// Independent of request cancel; pool is process-lifetime.
		serveSST = SST.Open(context.WithoutCancel(ctx), true)
	})
}

// Session returns the warmed graph session. Safe for concurrent read-ish use
// (same model as upstream open-per-process tools).
func Session() SST.PoSST {
	return serveSST
}

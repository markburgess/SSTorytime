package demopocs

import (
	"context"
	"fmt"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

// PostgresTestDB opens and closes a session (upstream postgres_testdb).
func PostgresTestDB(ctx context.Context) error {
	sst := SST.Open(ctx, false)
	defer SST.Close(sst)
	fmt.Println("postgres_testdb: open/close ok")
	return nil
}

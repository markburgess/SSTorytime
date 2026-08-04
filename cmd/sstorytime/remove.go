package main

import (
	"fmt"
	"strings"

	SST "github.com/markburgess/SSTorytime/internal/sst"
	"github.com/spf13/cobra"
)

var removeForce bool

var removeCmd = &cobra.Command{
	Use:   "remove [chapter...]",
	Short: "Remove a chapter from the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		chap := strings.Join(args, " ")
		if chap == "" {
			return ErrChapterRequired
		}
		if !removeForce {
			return fmt.Errorf("%w: pass --force to confirm removing %q", ErrRemove, chap)
		}
		sst := SST.Open(ctx, false)
		defer SST.Close(sst)
		if err := SST.DeleteChapter(ctx, sst, chap); err != nil {
			return fmt.Errorf("%w: %w", ErrRemove, err)
		}
		fmt.Println("Deleted", chap)
		return nil
	},
}

func init() {
	registerMultiCall("remove", "removeN4L")
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "force remove")
}

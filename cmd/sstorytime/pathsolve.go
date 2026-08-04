package main

import (
	SST "github.com/markburgess/SSTorytime/internal/sst"
	"github.com/spf13/cobra"
)

var (
	psBegin   string
	psEnd     string
	psChapter string
	psBwd     bool
)

var pathsolveCmd = &cobra.Command{
	Use:   "pathsolve",
	Short: "Solve paths between begin and end nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if psBegin == "" || psEnd == "" {
			return ErrPathsolveArgs
		}
		// Upstream stores FWD/BWD labels from -bwd; PathSolve itself uses begin/end sets.
		_ = psBwd
		sst := SST.Open(ctx, true)
		defer SST.Close(sst)
		PathSolveCLI(ctx, sst, psChapter, "", psBegin, psEnd)
		return nil
	},
}

func init() {
	registerMultiCall("pathsolve", "pathsolve", "PathSolve")
	pathsolveCmd.Flags().StringVar(&psBegin, "begin", "", "start match")
	pathsolveCmd.Flags().StringVar(&psEnd, "end", "", "end match")
	pathsolveCmd.Flags().StringVar(&psChapter, "chapter", "", "optional chapter filter")
	pathsolveCmd.Flags().BoolVar(&psBwd, "bwd", false, "reverse search direction")
}

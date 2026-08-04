package main

import (
	"fmt"

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
	Use:   "pathsolve [dirac]",
	Short: "Solve paths between begin and end nodes",
	Long: `Solve paths between begin and end node sets.

Flags --begin/--end set the endpoints. Alternatively pass Dirac notation as
the first positional argument (same as upstream pathsolve):

  <end|begin>
  <end|context|begin>

Example: pathsolve '<B6|A1>'  or  pathsolve --begin A1 --end B6`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		begin, end, cntext := psBegin, psEnd, ""

		// Upstream: optional free arg is Dirac <end|begin> or <end|context|begin>.
		if len(args) == 1 {
			ok, beg, en, cnt := SST.DiracNotation(args[0])
			if !ok {
				return fmt.Errorf("%w: positional arg must be Dirac <end|begin> or <end|context|begin>, got %q",
					ErrPathsolveArgs, args[0])
			}
			begin, end, cntext = beg, en, cnt
		}

		if begin == "" || end == "" {
			return ErrPathsolveArgs
		}
		// Upstream sets FWD/BWD labels from -bwd; PathSolve body never used them.
		_ = psBwd

		sst := SST.Open(ctx, true)
		defer SST.Close(sst)
		PathSolveCLI(ctx, sst, psChapter, cntext, begin, end)
		return nil
	},
}

func init() {
	registerMultiCall("pathsolve", "pathsolve", "PathSolve")
	pathsolveCmd.Flags().StringVar(&psBegin, "begin", "", "start match")
	pathsolveCmd.Flags().StringVar(&psEnd, "end", "", "end match")
	pathsolveCmd.Flags().StringVar(&psChapter, "chapter", "", "optional chapter filter")
	pathsolveCmd.Flags().BoolVar(&psBwd, "bwd", false, "reverse search direction (accepted; unused, same as upstream)")
}

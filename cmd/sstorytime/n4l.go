package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// n4lCmd is the N4L compiler/uploader (logic ported from former src/N4L).
var n4lCmd = &cobra.Command{
	Use:   "n4l [files...]",
	Short: "Compile and optionally upload N4L notes into the graph database",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Full N4L pipeline lives in internal/n4l; wire-up continues as domain is ported.
		return fmt.Errorf("n4l: not fully wired yet — domain port from internal/sst in progress (flags: upload/wipe/adj ready for next commit)")
	},
}

var (
	n4lUpload bool
	n4lWipe   bool
	n4lForce  bool
	n4lDiag   bool
	n4lSummary bool
	n4lAdj    string
)

func init() {
	n4lCmd.Flags().BoolVarP(&n4lUpload, "upload", "u", false, "upload into database")
	n4lCmd.Flags().BoolVar(&n4lWipe, "wipe", false, "wipe and reset before upload")
	n4lCmd.Flags().BoolVar(&n4lForce, "force", false, "force upload")
	n4lCmd.Flags().BoolVarP(&n4lDiag, "diagnostic", "d", false, "diagnostic mode")
	n4lCmd.Flags().BoolVarP(&n4lSummary, "summary", "s", false, "print summary")
	n4lCmd.Flags().StringVar(&n4lAdj, "adj", "none", "comma-separated short link names for adjacency")
}

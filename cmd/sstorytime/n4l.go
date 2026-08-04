package main

import (
	"github.com/markburgess/SSTorytime/internal/n4l"
	"github.com/markburgess/SSTorytime/internal/sstconfig"
	"github.com/spf13/cobra"
)

var (
	n4lUpload    bool
	n4lWipe      bool
	n4lForce     bool
	n4lDiag      bool
	n4lSummary   bool
	n4lAdj       string
	n4lConfigDir string
)

var n4lCmd = &cobra.Command{
	Use:   "n4l [files...]",
	Short: "Compile and optionally upload N4L notes into the graph database",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := n4l.Options{
			Files:       args,
			Upload:      n4lUpload,
			Force:       n4lForce,
			Wipe:        n4lWipe,
			Verbose:     verbose,
			Diag:        n4lDiag,
			Summary:     n4lSummary,
			AdjList:     n4lAdj,
			DatabaseURL: databaseURL,
		}
		if n4lConfigDir != "" {
			fsys, err := sstconfig.Dir(n4lConfigDir)
			if err != nil {
				return err
			}
			opt.ConfigFS = fsys
		}
		return n4l.Run(cmd.Context(), opt)
	},
}

func init() {
	registerMultiCall("n4l", "N4L", "N4L-db")
	n4lCmd.Flags().BoolVarP(&n4lUpload, "upload", "u", false, "upload into database")
	n4lCmd.Flags().BoolVar(&n4lWipe, "wipe", false, "wipe and reset before upload")
	n4lCmd.Flags().BoolVar(&n4lForce, "force", false, "force upload")
	n4lCmd.Flags().BoolVarP(&n4lDiag, "diagnostic", "d", false, "diagnostic mode")
	n4lCmd.Flags().BoolVarP(&n4lSummary, "summary", "s", false, "print summary")
	n4lCmd.Flags().StringVar(&n4lAdj, "adj", "none", "comma-separated short link names for adjacency")
	n4lCmd.Flags().StringVar(&n4lConfigDir, "config", "", "arrow config directory (default: embedded internal/sstconfig); explicit opt-in only")
}

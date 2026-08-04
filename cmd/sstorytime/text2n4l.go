package main

import (
	"github.com/spf13/cobra"
)

var text2Pct float64

var text2n4lCmd = &cobra.Command{
	Use:   "text2n4l [file]",
	Short: "Extract high-intentionality sentences into N4L-ish output",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		RipFile2File(args[0], text2Pct)
		return nil
	},
}

func init() {
	text2n4lCmd.Flags().Float64Var(&text2Pct, "percent", 50, "approximate percentage of file to skim")
}

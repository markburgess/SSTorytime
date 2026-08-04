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
	registerMultiCall("text2n4l", "text2N4L", "Text2N4L")
	text2n4lCmd.Flags().Float64Var(&text2Pct, "percent", 50, "approximate percentage of file to skim")
	// Upstream text2N4L used flag name "%" (e.g. -% 30). Same variable.
	text2n4lCmd.Flags().Float64Var(&text2Pct, "%", 50, "alias for --percent (upstream text2N4L)")
	if err := text2n4lCmd.Flags().MarkHidden("%"); err != nil {
		panic(err) // flag must exist; programming error
	}
}

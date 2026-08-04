package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	grChapter string
	grSTType  string
	grDepth   int
)

var graphReportCmd = &cobra.Command{
	Use:   "graph-report",
	Short: "Print a graph analytics report",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("%w: chapter=%q sttype=%q depth=%d", ErrGraphReport, grChapter, grSTType, grDepth)
	},
}

func init() {
	graphReportCmd.Flags().StringVar(&grChapter, "chapter", "", "chapter substring")
	graphReportCmd.Flags().StringVar(&grSTType, "sttype", "+L", "link ST types e.g. L,C,P,N")
	graphReportCmd.Flags().IntVar(&grDepth, "depth", 3, "max probe depth")
}

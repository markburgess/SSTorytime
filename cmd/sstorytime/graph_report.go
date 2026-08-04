package main

import (
	"strings"

	SST "github.com/markburgess/SSTorytime/internal/sst"
	"github.com/spf13/cobra"
)

var (
	grChapter string
	grSTType  string
	grDepth   int
)

var graphReportCmd = &cobra.Command{
	Use:   "graph-report [context...]",
	Short: "Print a graph analytics report",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		chapter := grChapter
		if chapter == "" {
			chapter = "none"
		}
		var sttypes []int
		if grSTType != "" {
			sttypes = make([]int, 0)
			seen := make(map[int]bool)
			for _, t := range strings.Split(grSTType, ",") {
				switch t {
				case "L", "+L":
					seen[SST.LEADSTO] = true
				case "-L":
					seen[-SST.LEADSTO] = true
				case "C", "+C":
					seen[SST.CONTAINS] = true
				case "-C":
					seen[-SST.CONTAINS] = true
				case "P", "E", "+P", "+E":
					seen[SST.EXPRESS] = true
				case "-P", "-E":
					seen[-SST.EXPRESS] = true
				case "N", "+N", "-N":
					seen[SST.NEAR] = true
					seen[-SST.NEAR] = true
				}
			}
			for k := range seen {
				sttypes = append(sttypes, k)
			}
		}
		if len(sttypes) == 0 {
			sttypes = []int{SST.LEADSTO}
		}
		context := []string{""}
		if len(args) > 0 {
			context = args
		}
		sst := SST.Open(ctx, true)
		defer SST.Close(sst)
		chaps := SST.GetDBChaptersMatchingName(ctx, sst, chapter)
		for _, chap := range chaps {
			AnalyzeGraphCLI(ctx, sst, chap, context, sttypes, grDepth)
		}
		return nil
	},
}

func init() {
	registerMultiCall("graph-report", "graph_report")
	graphReportCmd.Flags().StringVar(&grChapter, "chapter", "", "chapter substring")
	graphReportCmd.Flags().StringVar(&grSTType, "sttype", "+L", "link ST types e.g. L,C,P,N")
	graphReportCmd.Flags().IntVar(&grDepth, "depth", 3, "max probe depth")
}

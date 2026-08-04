package main

import (
	"fmt"
	"strings"

	SST "github.com/markburgess/SSTorytime/internal/sst"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query...]",
	Short: "Search the graph (former searchN4L)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if len(args) < 1 {
			return ErrQueryRequired
		}
		var search_string string
		for a := 0; a < len(args); a++ {
			if strings.Contains(args[a], " ") {
				search_string += fmt.Sprintf("\"%s\"", args[a]) + " "
			} else {
				search_string += args[a] + " "
			}
		}
		search_string = SST.CheckRemindQuery(search_string)
		search_string = SST.CheckHelpQuery(search_string)
		search_string = SST.CheckConceptQuery(search_string)
		search := SST.DecodeSearchField(search_string)

		sst := SST.Open(ctx, false)
		defer SST.Close(sst)
		RunSearchCLI(ctx, sst, search, search_string)
		return nil
	},
}

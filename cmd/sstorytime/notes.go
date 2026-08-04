package main

import (
	"strings"

	SST "github.com/markburgess/SSTorytime/internal/sst"
	"github.com/spf13/cobra"
)

var notesPage int

var notesCmd = &cobra.Command{
	Use:   "notes [chapter words...]",
	Short: "Browse notes by chapter (former notes tool)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if len(args) == 0 {
			return ErrNotes
		}
		chapter := strings.Join(args, " ")
		sst := SST.Open(ctx, true)
		defer SST.Close(sst)
		Page(ctx, sst, chapter, []string{""}, notesPage)
		return nil
	},
}

func init() {
	registerMultiCall("notes", "notes")
	notesCmd.Flags().IntVar(&notesPage, "page", 1, "page number")
}

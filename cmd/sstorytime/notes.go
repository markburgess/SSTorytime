package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var notesPage int

var notesCmd = &cobra.Command{
	Use:   "notes [chapter words...]",
	Short: "Browse notes by chapter (former notes tool)",
	RunE: func(cmd *cobra.Command, args []string) error {
		chap := strings.Join(args, " ")
		return fmt.Errorf("%w: chapter=%q page=%d", ErrNotes, chap, notesPage)
	},
}

func init() {
	notesCmd.Flags().IntVar(&notesPage, "page", 1, "page number")
}

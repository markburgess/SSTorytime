package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query...]",
	Short: "Search the graph (former searchN4L)",
	RunE: func(cmd *cobra.Command, args []string) error {
		q := strings.Join(args, " ")
		if q == "" {
			return ErrQueryRequired
		}
		return fmt.Errorf("%w: query=%q", ErrSearch, q)
	},
}

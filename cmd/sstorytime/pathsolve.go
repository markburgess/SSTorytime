package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	psBegin   string
	psEnd     string
	psChapter string
	psBwd     bool
)

var pathsolveCmd = &cobra.Command{
	Use:   "pathsolve",
	Short: "Solve paths between begin and end nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if psBegin == "" || psEnd == "" {
			return ErrPathsolveArgs
		}
		return fmt.Errorf("%w: begin=%q end=%q chapter=%q bwd=%v", ErrPathsolve, psBegin, psEnd, psChapter, psBwd)
	},
}

func init() {
	pathsolveCmd.Flags().StringVar(&psBegin, "begin", "", "start match")
	pathsolveCmd.Flags().StringVar(&psEnd, "end", "", "end match")
	pathsolveCmd.Flags().StringVar(&psChapter, "chapter", "", "optional chapter filter")
	pathsolveCmd.Flags().BoolVar(&psBwd, "bwd", false, "reverse search direction")
}

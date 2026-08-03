package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var removeForce bool

var removeCmd = &cobra.Command{
	Use:   "remove [chapter...]",
	Short: "Remove a chapter from the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		chap := strings.Join(args, " ")
		if chap == "" {
			return fmt.Errorf("remove: chapter name required")
		}
		return fmt.Errorf("remove: not fully wired yet — chapter=%q force=%v", chap, removeForce)
	},
}

func init() {
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "force remove")
}

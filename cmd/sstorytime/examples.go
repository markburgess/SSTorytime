package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "List/load N4L corpora and run API demos",
}

var examplesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List example N4L files under examples/",
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "examples"
		return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".n4l") {
				fmt.Println(path)
			}
			return nil
		})
	},
}

var examplesLoadCmd = &cobra.Command{
	Use:   "load [name|all]",
	Short: "Load example N4L into the database (requires n4l upload)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "all"
		if len(args) == 1 {
			target = args[0]
		}
		return fmt.Errorf("%w: target=%q (will call n4l -u)", ErrExamplesLoad, target)
	},
}

var examplesRunCmd = &cobra.Command{
	Use:   "run [demo]",
	Short: "Run a Go API demo (api-1, api-2, api-3, api-4, …)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("%w: demo=%q", ErrExamplesRun, args[0])
	},
}

func init() {
	examplesCmd.AddCommand(examplesListCmd, examplesLoadCmd, examplesRunCmd)
}

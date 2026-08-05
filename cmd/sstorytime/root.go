package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "sstorytime",
	Short: "Semantic Spacetime story graph tools",
	Long: `sstorytime is the SSTorytime CLI: N4L compile/upload, search, notes,
path solve, HTTP UI, and related tools.

Busybox-style installs: symlink or hardlink this binary as N4L, searchN4L,
http_server, etc. Those names dispatch to the matching subcommand.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output (where supported by subcommands)")
}

// Execute runs the root command under ctx (from main's signal.NotifyContext).
func Execute(ctx context.Context) error {
	os.Args = applyMultiCall(os.Args)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

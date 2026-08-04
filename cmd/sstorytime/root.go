package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	databaseURL string
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "sstorytime",
	Short: "Semantic Spacetime story graph tools",
	Long: `sstorytime is the SSTorytime CLI: N4L compile/upload, search, notes,
path solve, HTTP UI, examples, and database migrations.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&databaseURL, "database-url", "", "Postgres DSN (overrides DATABASE_URL / POSTGRESQL_URI / ~/.SSTorytime)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(n4lCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(notesCmd)
	rootCmd.AddCommand(pathsolveCmd)
	rootCmd.AddCommand(graphReportCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(text2n4lCmd)
	rootCmd.AddCommand(examplesCmd)
}

// Execute runs the root command under ctx (from main's signal.NotifyContext).
// Busybox-style names (N4L, searchN4L, …) are rewritten via multiCall before Cobra runs.
func Execute(ctx context.Context) error {
	os.Args = applyMultiCall(os.Args)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

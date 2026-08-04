package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	demopocs "github.com/markburgess/SSTorytime/internal/demo_pocs"
	"github.com/spf13/cobra"
)

// demoRunner is the shared shape for examples run <name> subcommands.
type demoRunner func(context.Context) error

type demoEntry struct {
	Run  demoRunner
	Desc string
}

// demoRunners is filled from init() via registerDemo (api_* and demopocs).
var demoRunners = map[string]demoEntry{}

// registerDemo binds a name under "examples run" to a func(ctx) error.
func registerDemo(name, desc string, run demoRunner) {
	if name == "" || run == nil {
		return
	}
	demoRunners[name] = demoEntry{Run: run, Desc: desc}
}

func demoNames() []string {
	names := make([]string, 0, len(demoRunners))
	for n := range demoRunners {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// wireDemoSubcommands attaches one Cobra child per registered demo under parent
// (typically "examples run").
func wireDemoSubcommands(parent *cobra.Command) {
	for _, name := range demoNames() {
		entry := demoRunners[name]
		n, e := name, entry
		parent.AddCommand(&cobra.Command{
			Use:   n,
			Short: e.Desc,
			RunE: func(cmd *cobra.Command, args []string) error {
				return e.Run(cmd.Context())
			},
		})
	}
	var b strings.Builder
	b.WriteString("Run a named demo. Available:\n")
	for _, name := range demoNames() {
		fmt.Fprintf(&b, "  %-20s %s\n", name, demoRunners[name].Desc)
	}
	parent.Long = strings.TrimSpace(b.String())
}

func init() {
	// Development pocs (internal/demo_pocs).
	registerDemo("definecontext", "context registry smoke test", demopocs.DefineContext)
	registerDemo("postgres_testdb", "open/close DB", demopocs.PostgresTestDB)
	registerDemo("dotest_getnodes", "name lookup parity (needs data)", demopocs.DotestGetNodes)
	registerDemo("dotest_entirecone", "NC vs plain cone (needs a7/i6)", demopocs.DotestEntireCone)
	// API examples registered in examples.go next to their Run bodies.
}

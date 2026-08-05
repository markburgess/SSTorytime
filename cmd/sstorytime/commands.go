package main

import (
	"os"

	"github.com/markburgess/SSTorytime/internal/app/apiexample1"
	"github.com/markburgess/SSTorytime/internal/app/apiexample2"
	"github.com/markburgess/SSTorytime/internal/app/apiexample3"
	"github.com/markburgess/SSTorytime/internal/app/apiexample4"
	"github.com/markburgess/SSTorytime/internal/app/definecontext"
	"github.com/markburgess/SSTorytime/internal/app/dotestentirecone"
	"github.com/markburgess/SSTorytime/internal/app/dotestgetnodes"
	"github.com/markburgess/SSTorytime/internal/app/graphreport"
	"github.com/markburgess/SSTorytime/internal/app/httpserver"
	"github.com/markburgess/SSTorytime/internal/app/n4l"
	"github.com/markburgess/SSTorytime/internal/app/notes"
	"github.com/markburgess/SSTorytime/internal/app/pathsolve"
	"github.com/markburgess/SSTorytime/internal/app/postgrestestdb"
	"github.com/markburgess/SSTorytime/internal/app/removen4l"
	"github.com/markburgess/SSTorytime/internal/app/searchn4l"
	"github.com/markburgess/SSTorytime/internal/app/text2n4l"
	"github.com/spf13/cobra"
)

// runLegacy sets os.Args for a tool that still uses the stdlib flag package
// on the process argv, then invokes Main. Subcommands use DisableFlagParsing
// so cobra does not strip flags before we pass them through.
func runLegacy(argv0 string, args []string, main func()) {
	old := os.Args
	os.Args = append([]string{argv0}, args...)
	defer func() { os.Args = old }()
	main()
}

func legacyCmd(use, short, argv0 string, multiNames []string, main func()) *cobra.Command {
	c := &cobra.Command{
		Use:                use,
		Short:              short,
		DisableFlagParsing: true,
		Run: func(_ *cobra.Command, args []string) {
			runLegacy(argv0, args, main)
		},
	}
	registerMultiCall(use, multiNames...)
	// also register primary Use as multicall target name for the map value
	// multiNames map to `use` (first segment)
	return c
}

func init() {
	// use is the cobra subcommand name (multiCall value)
	cmds := []*cobra.Command{
		legacyCmd("n4l", "Compile and optionally upload N4L notes (former N4L)", "N4L",
			[]string{"N4L", "N4L-db"}, n4l.Main),
		legacyCmd("search", "Search the graph (former searchN4L)", "searchN4L",
			[]string{"searchN4L"}, searchn4l.Main),
		legacyCmd("remove", "Remove an uploaded chapter (former removeN4L)", "removeN4L",
			[]string{"removeN4L"}, removen4l.Main),
		legacyCmd("text2n4l", "Turn text into N4L notes (former text2N4L)", "text2N4L",
			[]string{"text2N4L"}, text2n4l.Main),
		legacyCmd("notes", "Browse notes in page layout", "notes",
			[]string{"notes"}, notes.Main),
		legacyCmd("pathsolve", "Path solving experiments", "pathsolve",
			[]string{"pathsolve"}, pathsolve.Main),
		legacyCmd("graph-report", "Graph reports, loops, centrality (former graph_report)", "graph_report",
			[]string{"graph_report"}, graphreport.Main),
		legacyCmd("serve", "HTTP UI and JSON API (former http_server)", "http_server",
			[]string{"http_server"}, httpserver.Main),
		legacyCmd("api-example-1", "API example 1", "API_EXAMPLE_1",
			[]string{"API_EXAMPLE_1"}, apiexample1.Main),
		legacyCmd("api-example-2", "API example 2", "API_EXAMPLE_2",
			[]string{"API_EXAMPLE_2"}, apiexample2.Main),
		legacyCmd("api-example-3", "API example 3", "API_EXAMPLE_3",
			[]string{"API_EXAMPLE_3"}, apiexample3.Main),
		legacyCmd("api-example-4", "API example 4", "API_EXAMPLE_4",
			[]string{"API_EXAMPLE_4"}, apiexample4.Main),
		legacyCmd("definecontext", "Demo: define context", "definecontext",
			[]string{"definecontext"}, definecontext.Main),
		legacyCmd("dotest-entirecone", "Demo: entire cone", "dotest_entirecone",
			[]string{"dotest_entirecone"}, dotestentirecone.Main),
		legacyCmd("dotest-getnodes", "Demo: get nodes", "dotest_getnodes",
			[]string{"dotest_getnodes"}, dotestgetnodes.Main),
		legacyCmd("postgres-testdb", "Demo: postgres test db", "postgres_testdb",
			[]string{"postgres_testdb"}, postgrestestdb.Main),
	}
	for _, c := range cmds {
		rootCmd.AddCommand(c)
	}
}

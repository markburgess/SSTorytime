package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/markburgess/SSTorytime/internal/n4l"
	SST "github.com/markburgess/SSTorytime/internal/sst"
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
		return filepath.WalkDir("examples", func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(path, ".n4l") {
				fmt.Println(path)
			}
			return nil
		})
	},
}

var examplesLoadCmd = &cobra.Command{
	Use:   "load [name|all]",
	Short: "Load example N4L into the database (n4l -u --force)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		target := "all"
		if len(args) == 1 {
			target = args[0]
		}
		var files []string
		err := filepath.WalkDir("examples", func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".n4l") {
				return nil
			}
			if target == "all" || strings.Contains(path, target) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("%w: %w", ErrExamplesLoad, err)
		}
		if len(files) == 0 {
			return fmt.Errorf("%w: no .n4l matching %q under examples/", ErrExamplesLoad, target)
		}
		for _, f := range files {
			fmt.Println("loading", f)
			if err := n4l.Run(ctx, n4l.Options{
				Files:       []string{f},
				Upload:      true,
				Force:       true,
				DatabaseURL: databaseURL,
				Verbose:     verbose,
			}); err != nil {
				return fmt.Errorf("%w: %s: %w", ErrExamplesLoad, f, err)
			}
		}
		return nil
	},
}

var examplesRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Go API demo or development poc",
	// Long and children filled by wireDemoSubcommands from demoRunners.
}

func init() {
	registerMultiCall("examples", "sstorytime-examples")
	registerDemo("api-1", "upstream API_EXAMPLE_1 (lamb story)", runAPI1)
	registerDemo("api-2", "upstream API_EXAMPLE_2 (HubJoin)", runAPI2)
	registerDemo("api-3", "upstream API_EXAMPLE_3 (maze paths)", runAPI3)
	registerDemo("api-4", "upstream API_EXAMPLE_4 (needs double.n4l)", runAPI4)
	wireDemoSubcommands(examplesRunCmd)
	examplesCmd.AddCommand(examplesListCmd, examplesLoadCmd, examplesRunCmd)
}

// Ported from upstream API_EXAMPLE_1
func runAPI1(ctx context.Context) error {
	sst := SST.Open(ctx, false)
	defer SST.Close(sst)
	chap := "home and away"
	context := []string{""}
	var w float32 = 1.0
	n1 := SST.Vertex(ctx, &sst, "Mary had a little lamb", chap)
	n2 := SST.Vertex(ctx, &sst, "Whose fleece was dull and grey", chap)
	n3 := SST.Vertex(ctx, &sst, "And every time she washed it clean", chap)
	n4 := SST.Vertex(ctx, &sst, "It just went to roll in the hay", chap)
	n5 := SST.Vertex(ctx, &sst, "And when it reached a certain age ", chap)
	n6 := SST.Vertex(ctx, &sst, "She'd serve it on a tray", chap)
	SST.Edge(ctx, &sst, n1, "then", n2, context, w)
	SST.Edge(ctx, &sst, n2, "then", n3, context, w/2)
	SST.Edge(ctx, &sst, n2, "then", n5, context, w/2)
	SST.Edge(ctx, &sst, n3, "then", n4, context, w)
	SST.Edge(ctx, &sst, n5, "then", n6, context, w)
	start := SST.GetDBNodePtrMatchingName(ctx, sst, "Mary had a little lamb", chap)
	if len(start) == 0 {
		return fmt.Errorf("%w: api-1 lookup empty", ErrExamplesRun)
	}
	paths, npaths := SST.GetFwdPathsAsLinks(ctx, &sst, start[0], SST.LEADSTO, 6, 20)
	if npaths == 0 && len(paths) == 0 {
		fmt.Println("api-1: no forward paths")
	}
	for s := range paths {
		SST.PrintLinkPath(ctx, &sst, paths, s, " story: ", "", nil)
	}
	return nil
}

// Ported from upstream API_EXAMPLE_2
func runAPI2(ctx context.Context) error {
	sst := SST.Open(ctx, false)
	defer SST.Close(sst)
	names := []string{"test_node1", "test_node2", "test_node3"}
	weights := []float32{0.2, 0.4, 1.0}
	context := []string{"some", "context", "tags"}
	var nptrs []SST.NodePtr
	for _, name := range names {
		n := SST.Vertex(ctx, &sst, name, "my chapter")
		nptrs = append(nptrs, n.NPtr)
	}
	created1 := SST.HubJoin(ctx, &sst, "", "", nptrs, "then", context, weights)
	fmt.Println("Creates hub node", created1)
	created2 := SST.HubJoin(ctx, &sst, "mummy_node", "", nptrs, "belongs to", nil, nil)
	fmt.Println("Creates hub node", created2)
	return nil
}

// Ported from upstream API_EXAMPLE_3 (maze)
func runAPI3(ctx context.Context) error {
	var path [9][]string
	path[0] = []string{"maze_a7", "maze_b7", "maze_b6", "maze_c6", "maze_c5", "maze_b5", "maze_b4", "maze_a4", "maze_a3", "maze_b3", "maze_c3", "maze_d3", "maze_d2", "maze_e2", "maze_e3", "maze_f3", "maze_f4", "maze_e4", "maze_e5", "maze_f5", "maze_f6", "maze_g6", "maze_g5", "maze_g4", "maze_h4", "maze_h5", "maze_h6", "maze_i6"}
	path[1] = []string{"maze_d1", "maze_d2"}
	path[2] = []string{"maze_f1", "maze_f2", "maze_e2"}
	path[3] = []string{"maze_f2", "maze_g2", "maze_h2", "maze_h3", "maze_g3", "maze_g2"}
	path[4] = []string{"maze_b1", "maze_c1", "maze_c2", "maze_b2", "maze_b1"}
	path[5] = []string{"maze_b7", "maze_b8", "maze_c8", "maze_c7", "maze_d7", "maze_d6", "maze_e6", "maze_e7", "maze_f7", "maze_f8"}
	path[6] = []string{"maze_d7", "maze_d8", "maze_e8", "maze_e7"}
	path[7] = []string{"maze_f7", "maze_g7", "maze_g8", "maze_h8", "maze_h7"}
	path[8] = []string{"maze_a2", "maze_a1"}

	sst := SST.Open(ctx, true)
	defer SST.Close(sst)
	for p := range path {
		for leg := 1; leg < len(path[p]); leg++ {
			chap := "solve maze"
			context := []string{""}
			var w float32 = 1.0
			nfrom := SST.Vertex(ctx, &sst, path[p][leg-1], chap)
			nto := SST.Vertex(ctx, &sst, path[p][leg], chap)
			SST.Edge(ctx, &sst, nfrom, "fwd", nto, context, w)
		}
	}
	start := SST.GetDBNodePtrMatchingName(ctx, sst, "maze_a7", "solve maze")
	end := SST.GetDBNodePtrMatchingName(ctx, sst, "maze_i6", "solve maze")
	if len(start) == 0 || len(end) == 0 {
		return fmt.Errorf("%w: api-3 maze ends missing", ErrExamplesRun)
	}
	sols := SST.GetPathsAndSymmetries(ctx, &sst, start, end, "solve maze", nil, nil, nil, 2, 40)
	for s := range sols {
		SST.PrintLinkPath(ctx, &sst, sols, s, " maze path: ", "", nil)
	}
	return nil
}

// Ported from upstream API_EXAMPLE_4 (needs double.n4l loaded)
func runAPI4(ctx context.Context) error {
	sst := SST.Open(ctx, true)
	defer SST.Close(sst)
	const branchingLimit = 2
	const maxdepth = 7
	var ldepth, rdepth = 2, 2
	startBC, endBC := "A1", "B6"
	leftptrs := SST.GetDBNodePtrMatchingName(ctx, sst, startBC, "")
	rightptrs := SST.GetDBNodePtrMatchingName(ctx, sst, endBC, "")
	if leftptrs == nil || rightptrs == nil {
		return fmt.Errorf("%w: api-4 needs double.n4l data (A1..B6)", ErrExamplesRun)
	}
	count := 0
	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {
		left_paths, Lnum := SST.GetEntireConePathsAsLinks(ctx, &sst, "any", leftptrs[0], ldepth, branchingLimit)
		right_paths, Rnum := SST.GetEntireConePathsAsLinks(ctx, &sst, "any", rightptrs[0], rdepth, branchingLimit)
		solutions, loop_corrections := SST.WaveFrontsOverlap(&sst, left_paths, right_paths, Lnum, Rnum, ldepth, rdepth)
		if len(solutions) > 0 {
			fmt.Println("-- T R E E ----------------------------------")
			fmt.Println("Path solution", count, "from", startBC, "to", endBC, "with lengths", ldepth, -rdepth)
			for s := 0; s < len(solutions); s++ {
				SST.PrintLinkPath(ctx, &sst, solutions, s, fmt.Sprintf(" - story %d: ", s), "", nil)
			}
			count++
		}
		if len(loop_corrections) > 0 {
			fmt.Println("-- L O O P S ----------------------------------")
			for s := 0; s < len(loop_corrections); s++ {
				SST.PrintLinkPath(ctx, &sst, loop_corrections, s, fmt.Sprintf(" - loop %d: ", s), "", nil)
			}
		}
		if turn%2 == 0 {
			ldepth++
		} else {
			rdepth++
		}
	}
	return nil
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/francescobianco/define/internal/dsl"
	"github.com/francescobianco/define/internal/graph"
	"github.com/francescobianco/define/internal/report"
)

var checkCmd = &cobra.Command{
	Use:   "check <file.def> [symbol]",
	Short: "Verify a .def model file",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runVerify,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runVerify(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return rootCmd.Help()
	}

	filePath := args[0]
	var rootSymbol string
	if len(args) > 1 {
		rootSymbol = args[1]
	}

	src, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", filePath, err)
	}

	file, err := dsl.Parse(string(src))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	if len(file.Definitions) == 0 {
		return fmt.Errorf("no definitions found in %s", filePath)
	}

	g := graph.Build(file)

	if rootSymbol == "" {
		rootSymbol = g.Order[0]
	}

	closure := graph.CheckClosure(g, rootSymbol)
	reachable := graph.Reachable(g, rootSymbol)
	dead := graph.FindDead(g, rootSymbol)

	report.Print(os.Stdout, rootSymbol, g, closure, dead, reachable)

	if !closure.Closed {
		os.Exit(1)
	}
	return nil
}
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/francescobianco/define/internal/dsl"
	"github.com/francescobianco/define/internal/extractor"
	"github.com/francescobianco/define/internal/graph"
	"github.com/francescobianco/define/internal/report"
)

var extractCmd = &cobra.Command{
	Use:   "extract <dir>",
	Short: "Extract a .def model from Go source code",
	Args:  cobra.ExactArgs(1),
	RunE:  runExtract,
}

func init() {
	extractCmd.Flags().StringP("out", "o", "", "write .def to file instead of stdout")
	extractCmd.Flags().Bool("no-tests", true, "ignore *_test.go files")
	extractCmd.Flags().Bool("check", false, "run verification after extraction")
	rootCmd.AddCommand(extractCmd)
}

func runExtract(cmd *cobra.Command, args []string) error {
	dir := args[0]
	outFile, _ := cmd.Flags().GetString("out")
	noTests, _ := cmd.Flags().GetBool("no-tests")
	doCheck, _ := cmd.Flags().GetBool("check")

	result, err := extractor.ExtractGoPackages(dir, noTests)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	if len(result.Packages) == 0 {
		return fmt.Errorf("no Go packages found in %s", dir)
	}

	var sb strings.Builder
	sb.WriteString("# extracted by: define extract " + dir + "\n\n")
	for _, pkg := range result.Packages {
		sb.WriteString("define " + pkg.ConceptName)
		if len(pkg.Deps) > 0 {
			sb.WriteString(" with " + strings.Join(pkg.Deps, ", "))
		}
		sb.WriteString("\n")
	}
	defSrc := sb.String()

	if outFile != "" {
		if err := os.WriteFile(outFile, []byte(defSrc), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outFile, err)
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", outFile)
	} else {
		fmt.Print(defSrc)
	}

	if doCheck {
		fmt.Fprintln(os.Stderr, "\n--- verification ---")
		file, err := dsl.Parse(defSrc)
		if err != nil {
			return fmt.Errorf("parse error in extracted model: %w", err)
		}
		g := graph.Build(file)
		rootSymbol := g.Order[0]
		closure := graph.CheckClosure(g, rootSymbol)
		reachable := graph.Reachable(g, rootSymbol)
		dead := graph.FindDead(g, rootSymbol)
		report.Print(os.Stderr, rootSymbol, g, closure, dead, reachable)
		if !closure.Closed {
			os.Exit(1)
		}
	}

	return nil
}

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
	Short: "Extract a .def model from source code",
	Long: `Extract a structural model from a codebase directory.

The --config flag points to a define.yml profile that lives outside the
codebase, so experiments never touch the upstream source tree.

Examples:
  define extract ./src
  define extract labs/repos/gin --config labs/profiles/go-generic.yml -o labs/reports/gin.def
  define extract labs/repos/laravel --config labs/profiles/php-laravel.yml`,
	Args: cobra.ExactArgs(1),
	RunE: runExtract,
}

func init() {
	extractCmd.Flags().StringP("config", "c", "", "profile file (define.yml) — overrides auto-detection")
	extractCmd.Flags().StringP("out", "o", "", "write .def to file instead of stdout")
	extractCmd.Flags().Bool("no-tests", true, "ignore test files (*_test.go, *.spec.ts, *Test.php, …)")
	extractCmd.Flags().Bool("check", false, "run verification after extraction")
	rootCmd.AddCommand(extractCmd)
}

func runExtract(cmd *cobra.Command, args []string) error {
	dir := args[0]
	outFile, _ := cmd.Flags().GetString("out")
	cfgFile, _ := cmd.Flags().GetString("config")
	noTests, _ := cmd.Flags().GetBool("no-tests")
	doCheck, _ := cmd.Flags().GetBool("check")

	cfg, err := extractor.LoadConfig(dir, cfgFile)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	result, err := extractor.Extract(dir, cfg, noTests)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	if len(result.Packages) == 0 {
		return fmt.Errorf("no concepts found in %s (language: %s)", dir, cfg.Language)
	}

	header := "# extracted by: define extract " + dir
	if cfgFile != "" {
		header += " --config " + cfgFile
	}
	var sb strings.Builder
	sb.WriteString(header + "\n\n")
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
		rootSymbol := cfg.Root
		if rootSymbol == "" {
			rootSymbol = g.Order[0]
		}
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

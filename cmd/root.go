package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "define [file.def] [symbol]",
	Short: "Structural CI gate for design closure and dead concept detection",
	Long: `define verifies closure, reachability and dead concepts in a .def model.

Examples:
  define model.def                   verify from first declared symbol
  define model.def Example1          verify from a specific root symbol
  define check model.def             explicit check subcommand
  define extract ./src               extract model from source code (v0.2)`,
	Args: cobra.RangeArgs(0, 2),
}

func init() {
	rootCmd.RunE = runVerify
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
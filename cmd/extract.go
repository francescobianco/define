package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract <dir>",
	Short: "Extract a .def model from source code (coming in v0.2)",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Println("extract is not yet implemented (planned for v0.2)")
		fmt.Printf("Target directory: %s\n", args[0])
		return nil
	},
}

func init() {
	extractCmd.Flags().StringP("config", "c", "define.yml", "projection config file")
	extractCmd.Flags().StringP("out", "o", "model.def", "output .def file")
	rootCmd.AddCommand(extractCmd)
}
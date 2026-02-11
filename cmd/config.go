package cmd

import (
	"fmt"
	"os"

	"github.com/anthropics/opencc/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configurations (TUI)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}
		return nil
	},
}

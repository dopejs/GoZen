package cmd

import (
	"fmt"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/tui"
	"github.com/spf13/cobra"
)

var pickCmd = &cobra.Command{
	Use:           "pick [cli args...]",
	Short:         "Select providers interactively and start proxy",
	Long:          "Launch a checkbox picker to select providers for this session, then start the proxy.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runPick,
}

var pickCLIFlag string

func init() {
	pickCmd.Flags().StringVar(&pickCLIFlag, "cli", "", "CLI to use (claude, codex, opencode)")
}

func runPick(cmd *cobra.Command, args []string) error {
	available := config.ProviderNames()
	if len(available) == 0 {
		fmt.Println("No providers configured. Run 'zen config' to set up providers.")
		return nil
	}

	selected, err := tui.RunPick()
	if err != nil {
		// User cancelled
		if err.Error() == "cancelled" {
			return nil
		}
		return err
	}
	if len(selected) == 0 {
		// User cancelled without selecting
		return nil
	}

	cli := pickCLIFlag
	if cli == "" {
		cli = config.GetDefaultCLI()
	}
	return startProxy(selected, nil, cli, args)
}

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

var pickClientFlag string

func init() {
	pickCmd.Flags().StringVarP(&pickClientFlag, "client", "c", "", "client to use (claude, codex, opencode)")
	pickCmd.Flags().String("cli", "", "alias for --client (deprecated)")
	pickCmd.Flags().Lookup("cli").Hidden = true
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

	// Support --cli as alias for --client
	client := pickClientFlag
	if client == "" {
		client, _ = cmd.Flags().GetString("cli")
	}
	if client == "" {
		client = config.GetDefaultClient()
	}
	return startProxy(selected, nil, client, args)
}

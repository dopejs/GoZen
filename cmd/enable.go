package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dopejs/gozen/internal/config"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:               "enable <provider>",
	Short:             "Clear unavailability marking for a provider",
	Long:              "Clear the unavailability marking for a provider, allowing the proxy to route requests to it again.",
	ValidArgsFunction: completeConfigNames,
	SilenceUsage:      true,
	SilenceErrors:     true,
	RunE:              runEnable,
}

func runEnable(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: zen enable <provider>")
		available := config.ProviderNames()
		if len(available) > 0 {
			fmt.Printf("\nAvailable providers: %s\n", strings.Join(available, ", "))
		}
		return nil
	}

	name := args[0]

	// Validate provider exists
	if config.GetProvider(name) == nil {
		available := config.ProviderNames()
		fmt.Fprintf(os.Stderr, "Error: provider %q not found\n", name)
		if len(available) > 0 {
			fmt.Printf("Available providers: %s\n", strings.Join(available, ", "))
		}
		return fmt.Errorf("provider %q not found", name)
	}

	// Check if actually disabled
	if !config.IsProviderDisabled(name) {
		fmt.Printf("Provider %q is not currently disabled.\n", name)
		return nil
	}

	if err := config.EnableProvider(name); err != nil {
		return err
	}

	fmt.Printf("Provider %q is now enabled.\n", name)
	return nil
}

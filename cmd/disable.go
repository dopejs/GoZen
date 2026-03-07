package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dopejs/gozen/internal/config"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:               "disable <provider>",
	Short:             "Mark a provider as unavailable",
	Long:              "Mark a provider as unavailable, preventing the proxy from routing requests to it.",
	ValidArgsFunction: completeConfigNames,
	SilenceUsage:      true,
	SilenceErrors:     true,
	RunE:              runDisable,
}

var (
	disableTodayFlag     bool
	disableMonthFlag     bool
	disablePermanentFlag bool
	disableListFlag      bool
)

func init() {
	disableCmd.Flags().BoolVar(&disableTodayFlag, "today", false, "mark unavailable for today only (default)")
	disableCmd.Flags().BoolVar(&disableMonthFlag, "month", false, "mark unavailable for this calendar month")
	disableCmd.Flags().BoolVar(&disablePermanentFlag, "permanent", false, "mark unavailable permanently until manually cleared")
	disableCmd.Flags().BoolVar(&disableListFlag, "list", false, "list all disabled providers")
}

func runDisable(cmd *cobra.Command, args []string) error {
	// Handle --list flag
	if disableListFlag {
		return runDisableList()
	}

	if len(args) == 0 {
		fmt.Println("Usage: zen disable <provider> [--today|--month|--permanent]")
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

	// Determine marking type
	markingType := config.MarkingTypeToday // default
	if disableMonthFlag {
		markingType = config.MarkingTypeMonth
	} else if disablePermanentFlag {
		markingType = config.MarkingTypePermanent
	}

	if err := config.DisableProvider(name, markingType); err != nil {
		return err
	}

	// Get the marking for display
	disabled := config.GetDisabledProviders()
	if m, ok := disabled[name]; ok {
		if m.ExpiresAt.IsZero() {
			fmt.Printf("Provider %q marked as unavailable (%s, no expiration)\n", name, m.Type)
		} else {
			fmt.Printf("Provider %q marked as unavailable (%s, expires %s)\n", name, m.Type, m.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}

func runDisableList() error {
	disabled := config.GetDisabledProviders()
	if len(disabled) == 0 {
		fmt.Println("No providers are currently disabled.")
		return nil
	}

	fmt.Println("Disabled providers:")
	for name, m := range disabled {
		if m.ExpiresAt.IsZero() {
			fmt.Printf("  %-20s %-12s no expiration\n", name, m.Type)
		} else {
			fmt.Printf("  %-20s %-12s expires %s\n", name, m.Type, m.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
	}
	return nil
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/dopejs/gozen/internal/config"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:               "use <config> [cli args...]",
	Short:             "Load config and exec CLI directly",
	ValidArgsFunction: completeConfigNames,
	SilenceUsage:      true,
	SilenceErrors:     true,
	RunE:              runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	available := config.ProviderNames()

	if len(args) == 0 {
		fmt.Println("Usage: zen use <provider> [cli args...]")
		if len(available) > 0 {
			fmt.Printf("\nAvailable providers: %s\n", strings.Join(available, ", "))
		} else {
			fmt.Println("\nNo providers configured. Run 'zen config' to set up providers.")
		}
		return nil
	}

	configName := args[0]
	cliArgs := args[1:]

	if err := config.ExportProviderToEnv(configName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if len(available) > 0 {
			fmt.Printf("Available providers: %s\n", strings.Join(available, ", "))
		} else {
			fmt.Println("No providers configured. Run 'zen config' to set up providers.")
		}
		return nil
	}

	// Get CLI binary name from config
	cliBin := config.GetDefaultCLI()
	if cliBin == "" {
		cliBin = "claude"
	}

	// Find CLI binary
	cliPath, err := exec.LookPath(cliBin)
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", cliBin, err)
	}

	// Replace process with CLI (like shell exec)
	argv := append([]string{cliBin}, cliArgs...)
	return syscall.Exec(cliPath, argv, os.Environ())
}

func completeConfigNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names := config.ProviderNames()
	return names, cobra.ShellCompDirectiveNoFileComp
}

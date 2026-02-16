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

var useClientFlag string

func init() {
	useCmd.Flags().StringVarP(&useClientFlag, "client", "c", "", "client to use (claude, codex, opencode)")
	useCmd.Flags().String("cli", "", "alias for --client (deprecated)")
	useCmd.Flags().Lookup("cli").Hidden = true
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

	// Get client binary name from flag or config
	// Support --cli as alias for --client
	clientBin := useClientFlag
	if clientBin == "" {
		clientBin, _ = cmd.Flags().GetString("cli")
	}
	if clientBin == "" {
		clientBin = config.GetDefaultClient()
	}
	if clientBin == "" {
		clientBin = "claude"
	}

	// Find client binary
	clientPath, err := exec.LookPath(clientBin)
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", clientBin, err)
	}

	// Replace process with client (like shell exec)
	argv := append([]string{clientBin}, cliArgs...)
	return syscall.Exec(clientPath, argv, os.Environ())
}

func completeConfigNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names := config.ProviderNames()
	return names, cobra.ShellCompDirectiveNoFileComp
}

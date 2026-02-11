package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/anthropics/opencc/internal/envfile"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:               "use <config> [claude args...]",
	Short:             "Load config and exec claude directly",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completeConfigNames,
	RunE:              runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	configName := args[0]
	claudeArgs := args[1:]

	cfg, err := envfile.LoadByName(configName)
	if err != nil {
		return fmt.Errorf("configuration '%s' not found: %w", configName, err)
	}

	// Export all env vars
	cfg.ExportToEnv()

	// Find claude binary
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}

	// Replace process with claude (like shell exec)
	argv := append([]string{"claude"}, claudeArgs...)
	return syscall.Exec(claudeBin, argv, os.Environ())
}

func completeConfigNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names := envfile.ConfigNames()
	return names, cobra.ShellCompDirectiveNoFileComp
}

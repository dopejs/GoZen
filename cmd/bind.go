package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dopejs/opencc/internal/config"
	"github.com/spf13/cobra"
)

var bindCmd = &cobra.Command{
	Use:   "bind <profile>",
	Short: "Bind current directory to a profile",
	Long: `Bind the current directory to a profile.
After binding, running 'opencc' in this directory will automatically use the bound profile.`,
	Args: cobra.ExactArgs(1),
	RunE: runBind,
}

var unbindCmd = &cobra.Command{
	Use:   "unbind",
	Short: "Remove profile binding for current directory",
	Long:  `Remove the profile binding for the current directory.`,
	Args:  cobra.NoArgs,
	RunE:  runUnbind,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show profile binding status for current directory",
	Long:  `Show the profile binding status for the current directory.`,
	Args:  cobra.NoArgs,
	RunE:  runStatus,
}

func runBind(cmd *cobra.Command, args []string) error {
	profile := args[0]

	// Get current directory (absolute path)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Clean the path
	cwd = filepath.Clean(cwd)

	// Bind the project
	if err := config.BindProject(cwd, profile); err != nil {
		return err
	}

	fmt.Printf("✓ Bound %s to profile '%s'\n", cwd, profile)
	fmt.Printf("\nNow running 'opencc' in this directory will use profile '%s'\n", profile)
	return nil
}

func runUnbind(cmd *cobra.Command, args []string) error {
	// Get current directory (absolute path)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Clean the path
	cwd = filepath.Clean(cwd)

	// Check if there's a binding
	profile := config.GetProjectBinding(cwd)
	if profile == "" {
		fmt.Printf("No binding found for %s\n", cwd)
		return nil
	}

	// Unbind the project
	if err := config.UnbindProject(cwd); err != nil {
		return err
	}

	fmt.Printf("✓ Removed binding for %s (was: '%s')\n", cwd, profile)
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Get current directory (absolute path)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Clean the path
	cwd = filepath.Clean(cwd)

	// Check for binding
	profile := config.GetProjectBinding(cwd)

	fmt.Printf("Directory: %s\n", cwd)
	if profile != "" {
		fmt.Printf("Bound to:  %s\n", profile)

		// Check if profile still exists
		if config.GetProfileConfig(profile) == nil {
			fmt.Printf("\n⚠️  Warning: Profile '%s' no longer exists (will use 'default')\n", profile)
		}
	} else {
		fmt.Printf("Bound to:  (none - will use 'default' or -f flag)\n")
	}

	// Show all bindings
	bindings := config.GetAllProjectBindings()
	if len(bindings) > 0 {
		fmt.Printf("\nAll project bindings:\n")
		for path, prof := range bindings {
			if path == cwd {
				fmt.Printf("  → %s → %s\n", path, prof)
			} else {
				fmt.Printf("    %s → %s\n", path, prof)
			}
		}
	}

	return nil
}

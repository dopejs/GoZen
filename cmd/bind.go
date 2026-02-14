package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dopejs/opencc/internal/config"
	"github.com/spf13/cobra"
)

var bindCmd = &cobra.Command{
	Use:   "bind [profile]",
	Short: "Bind current directory to a profile and/or CLI",
	Long: `Bind the current directory to a profile and/or CLI.
After binding, running 'opencc' in this directory will automatically use the bound settings.

Examples:
  opencc bind work              # Bind to profile 'work'
  opencc bind --cli codex       # Bind to use Codex CLI
  opencc bind work --cli codex  # Bind to profile 'work' with Codex CLI
  opencc bind --cli ""          # Clear CLI binding (use default)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBind,
}

var unbindCmd = &cobra.Command{
	Use:   "unbind",
	Short: "Remove binding for current directory",
	Long:  `Remove the profile and CLI binding for the current directory.`,
	Args:  cobra.NoArgs,
	RunE:  runUnbind,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show binding status for current directory",
	Long:  `Show the profile and CLI binding status for the current directory.`,
	Args:  cobra.NoArgs,
	RunE:  runStatus,
}

var bindCLI string

func init() {
	bindCmd.Flags().StringVar(&bindCLI, "cli", "", "CLI to use (claude, codex, opencode)")
}

func runBind(cmd *cobra.Command, args []string) error {
	var profile string
	if len(args) > 0 {
		profile = args[0]
	}

	// Check if --cli flag was explicitly set
	cliSet := cmd.Flags().Changed("cli")

	// If neither profile nor CLI specified, show usage
	if profile == "" && !cliSet {
		return fmt.Errorf("specify a profile name and/or --cli flag")
	}

	// Get current directory (absolute path)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Clean the path
	cwd = filepath.Clean(cwd)

	// Get existing binding to preserve values not being changed
	existing := config.GetProjectBinding(cwd)
	if existing != nil {
		if profile == "" {
			profile = existing.Profile
		}
		if !cliSet {
			bindCLI = existing.CLI
		}
	}

	// Bind the project
	if err := config.BindProject(cwd, profile, bindCLI); err != nil {
		return err
	}

	// Build status message
	var msg string
	if profile != "" && bindCLI != "" {
		msg = fmt.Sprintf("profile '%s' with CLI '%s'", profile, bindCLI)
	} else if profile != "" {
		msg = fmt.Sprintf("profile '%s'", profile)
	} else if bindCLI != "" {
		msg = fmt.Sprintf("CLI '%s'", bindCLI)
	}

	fmt.Printf("Bound %s to %s\n", cwd, msg)
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
	binding := config.GetProjectBinding(cwd)
	if binding == nil {
		fmt.Printf("No binding found for %s\n", cwd)
		return nil
	}

	// Unbind the project
	if err := config.UnbindProject(cwd); err != nil {
		return err
	}

	var was string
	if binding.Profile != "" && binding.CLI != "" {
		was = fmt.Sprintf("profile '%s', CLI '%s'", binding.Profile, binding.CLI)
	} else if binding.Profile != "" {
		was = fmt.Sprintf("profile '%s'", binding.Profile)
	} else if binding.CLI != "" {
		was = fmt.Sprintf("CLI '%s'", binding.CLI)
	}
	fmt.Printf("Removed binding for %s (was: %s)\n", cwd, was)
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
	binding := config.GetProjectBinding(cwd)

	fmt.Printf("Directory: %s\n", cwd)
	if binding != nil {
		if binding.Profile != "" {
			fmt.Printf("Profile:   %s\n", binding.Profile)
			// Check if profile still exists
			if config.GetProfileConfig(binding.Profile) == nil {
				fmt.Printf("           (profile no longer exists, will use default)\n")
			}
		} else {
			fmt.Printf("Profile:   (default)\n")
		}
		if binding.CLI != "" {
			fmt.Printf("CLI:       %s\n", binding.CLI)
		} else {
			fmt.Printf("CLI:       (default)\n")
		}
	} else {
		fmt.Printf("Profile:   (not bound, will use default)\n")
		fmt.Printf("CLI:       (not bound, will use default)\n")
	}

	// Show all bindings
	bindings := config.GetAllProjectBindings()
	if len(bindings) > 0 {
		fmt.Printf("\nAll project bindings:\n")
		for path, b := range bindings {
			marker := "  "
			if path == cwd {
				marker = "> "
			}
			var info string
			if b.Profile != "" && b.CLI != "" {
				info = fmt.Sprintf("%s (CLI: %s)", b.Profile, b.CLI)
			} else if b.Profile != "" {
				info = b.Profile
			} else if b.CLI != "" {
				info = fmt.Sprintf("(CLI: %s)", b.CLI)
			}
			fmt.Printf("%s%s -> %s\n", marker, path, info)
		}
	}

	return nil
}

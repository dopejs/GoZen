package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dopejs/gozen/internal/config"
	"github.com/spf13/cobra"
)

var bindCmd = &cobra.Command{
	Use:   "bind [profile]",
	Short: "Bind current directory to a profile and/or client",
	Long: `Bind the current directory to a profile and/or client.
After binding, running 'zen' in this directory will automatically use the bound settings.

Examples:
  zen bind work                 # Bind to profile 'work'
  zen bind --client codex       # Bind to use Codex
  zen bind work --client codex  # Bind to profile 'work' with Codex
  zen bind --client ""          # Clear client binding (use default)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBind,
}

var unbindCmd = &cobra.Command{
	Use:   "unbind",
	Short: "Remove binding for current directory",
	Long:  `Remove the profile and client binding for the current directory.`,
	Args:  cobra.NoArgs,
	RunE:  runUnbind,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show binding status for current directory",
	Long:  `Show the profile and client binding status for the current directory.`,
	Args:  cobra.NoArgs,
	RunE:  runStatus,
}

var bindClient string

func init() {
	bindCmd.Flags().StringVarP(&bindClient, "client", "c", "", "client to use (claude, codex, opencode)")
	bindCmd.Flags().String("cli", "", "alias for --client (deprecated)")
	bindCmd.Flags().Lookup("cli").Hidden = true
}
func runBind(cmd *cobra.Command, args []string) error {
	var profile string
	if len(args) > 0 {
		profile = args[0]
	}

	// Check if --client or --cli flag was explicitly set
	clientSet := cmd.Flags().Changed("client") || cmd.Flags().Changed("cli")
	if cmd.Flags().Changed("cli") && !cmd.Flags().Changed("client") {
		bindClient, _ = cmd.Flags().GetString("cli")
	}

	// If neither profile nor client specified, show usage
	if profile == "" && !clientSet {
		return fmt.Errorf("specify a profile name and/or --client flag")
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
		if !clientSet {
			bindClient = existing.Client
		}
	}

	// Bind the project
	if err := config.BindProject(cwd, profile, bindClient); err != nil {
		return err
	}

	// Build status message
	var msg string
	if profile != "" && bindClient != "" {
		msg = fmt.Sprintf("profile '%s' with client '%s'", profile, bindClient)
	} else if profile != "" {
		msg = fmt.Sprintf("profile '%s'", profile)
	} else if bindClient != "" {
		msg = fmt.Sprintf("client '%s'", bindClient)
	}

	fmt.Printf("Bound %s to %s\n", cwd, msg)
	return nil
}

func runUnbind(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	cwd = filepath.Clean(cwd)

	binding := config.GetProjectBinding(cwd)
	if binding == nil {
		fmt.Printf("No binding found for %s\n", cwd)
		return nil
	}

	if err := config.UnbindProject(cwd); err != nil {
		return err
	}

	var was string
	if binding.Profile != "" && binding.Client != "" {
		was = fmt.Sprintf("profile '%s', client '%s'", binding.Profile, binding.Client)
	} else if binding.Profile != "" {
		was = fmt.Sprintf("profile '%s'", binding.Profile)
	} else if binding.Client != "" {
		was = fmt.Sprintf("client '%s'", binding.Client)
	}
	fmt.Printf("Removed binding for %s (was: %s)\n", cwd, was)
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	cwd = filepath.Clean(cwd)

	binding := config.GetProjectBinding(cwd)

	fmt.Printf("Directory: %s\n", cwd)
	if binding != nil {
		if binding.Profile != "" {
			fmt.Printf("Profile:   %s\n", binding.Profile)
			if config.GetProfileConfig(binding.Profile) == nil {
				fmt.Printf("           (profile no longer exists, will use default)\n")
			}
		} else {
			fmt.Printf("Profile:   (default)\n")
		}
		if binding.Client != "" {
			fmt.Printf("Client:    %s\n", binding.Client)
		} else {
			fmt.Printf("Client:    (default)\n")
		}
	} else {
		fmt.Printf("Profile:   (not bound, will use default)\n")
		fmt.Printf("Client:    (not bound, will use default)\n")
	}

	bindings := config.GetAllProjectBindings()
	if len(bindings) > 0 {
		fmt.Printf("\nAll project bindings:\n")
		for path, b := range bindings {
			marker := "  "
			if path == cwd {
				marker = "> "
			}
			var info string
			if b.Profile != "" && b.Client != "" {
				info = fmt.Sprintf("%s (client: %s)", b.Profile, b.Client)
			} else if b.Profile != "" {
				info = b.Profile
			} else if b.Client != "" {
				info = fmt.Sprintf("(client: %s)", b.Client)
			}
			fmt.Printf("%s%s -> %s\n", marker, path, info)
		}
	}

	return nil
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/web"
	"github.com/dopejs/gozen/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage providers, profiles, and settings",
	Long: `Manage providers, profiles, and settings.

Usage:
  zen config [command]

Available Commands:
  add provider [name]    Add a new provider
  add profile [name]     Add a new profile
  edit provider <name>   Edit an existing provider
  edit profile <name>    Edit an existing profile
  delete provider <name> Delete a provider
  delete profile <name>  Delete a profile
  default-client         Set the default client
  default-profile        Set the default profile
  reset-password         Reset Web UI access password

Use "zen config [command] --help" for more information about a command.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Long)
	},
}

// --- add subcommands ---

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a provider or profile",
}

var configAddProviderCmd = &cobra.Command{
	Use:   "provider [name]",
	Short: "Add a new provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		_, err := tui.RunAddProvider(name)
		if err != nil && err.Error() == "cancelled" {
			return nil
		}
		return err
	},
}

var configAddGroupCmd = &cobra.Command{
	Use:   "profile [name]",
	Short: "Add a new profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		err := tui.RunAddGroup(name)
		if err != nil && err.Error() == "cancelled" {
			return nil
		}
		return err
	},
}

// --- delete subcommands ---

var configDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a provider or profile",
	Long:  "Delete a provider or profile.\n\nUsage:\n  zen config delete provider <name>\n  zen config delete profile <name>",
}

var configDeleteProviderCmd = &cobra.Command{
	Use:   "provider <name>",
	Short: "Delete a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}
		return deleteProvider(args[0])
	},
}

var configDeleteGroupCmd = &cobra.Command{
	Use:   "profile <name>",
	Short: "Delete a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}
		return deleteGroup(args[0])
	},
}

func deleteProvider(name string) error {
	names := config.ProviderNames()
	if len(names) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}
	if len(names) == 1 {
		fmt.Println("Cannot delete the last provider. At least one provider must remain.")
		return nil
	}

	if config.GetProvider(name) == nil {
		fmt.Printf("Provider %q not found.\n", name)
		return nil
	}

	if !confirmDelete(name) {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := config.DeleteProviderByName(name); err != nil {
		return err
	}
	fmt.Printf("Deleted provider %q.\n", name)
	return nil
}

func deleteGroup(name string) error {
	defaultProfile := config.GetDefaultProfile()
	if name == defaultProfile {
		fmt.Printf("Cannot delete the default profile '%s'.\n", defaultProfile)
		return nil
	}

	profiles := config.ListProfiles()
	found := false
	for _, p := range profiles {
		if p == name {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Profile %q not found.\n", name)
		return nil
	}

	if !confirmDelete(name) {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := config.DeleteProfile(name); err != nil {
		return err
	}
	fmt.Printf("Deleted profile %q.\n", name)
	return nil
}

func confirmDelete(name string) bool {
	fmt.Printf("Delete '%s'? This cannot be undone. (y/n): ", name)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// --- edit subcommands ---

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a provider or profile",
}

var configEditProviderCmd = &cobra.Command{
	Use:   "provider <name>",
	Short: "Edit a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}
		return editProvider(args[0])
	},
}

var configEditGroupCmd = &cobra.Command{
	Use:   "profile <name>",
	Short: "Edit a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}
		return editGroup(args[0])
	},
}

func editProvider(name string) error {
	if config.GetProvider(name) == nil {
		fmt.Printf("Provider %q not found.\n", name)
		return nil
	}

	err := tui.RunEditProvider(name)
	if err != nil && err.Error() == "cancelled" {
		return nil
	}
	return err
}

func editGroup(name string) error {
	profiles := config.ListProfiles()
	found := false
	for _, p := range profiles {
		if p == name {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Profile %q not found.\n", name)
		return nil
	}

	err := tui.RunEditGroup(name)
	if err != nil && err.Error() == "cancelled" {
		return nil
	}
	return err
}

// --- default-client / default-profile subcommands ---

var configDefaultClientCmd = &cobra.Command{
	Use:   "default-client",
	Short: "Set the default client",
	RunE: func(cmd *cobra.Command, args []string) error {
		current := config.GetDefaultClient()
		selected, err := tui.RunMinimalSelector(config.AvailableClients, current)
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
		if selected == current {
			return nil
		}
		if err := config.SetDefaultClient(selected); err != nil {
			return err
		}
		fmt.Printf("Default client set to %q.\n", selected)
		return nil
	},
}

var configDefaultProfileCmd = &cobra.Command{
	Use:   "default-profile",
	Short: "Set the default profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles := config.ListProfiles()
		if len(profiles) == 0 {
			fmt.Println("No profiles configured.")
			return nil
		}
		current := config.GetDefaultProfile()
		selected, err := tui.RunMinimalSelector(profiles, current)
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
		if selected == current {
			return nil
		}
		if err := config.SetDefaultProfile(selected); err != nil {
			return err
		}
		fmt.Printf("Default profile set to %q.\n", selected)
		return nil
	},
}

var configResetPasswordCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Reset Web UI access password",
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := web.GeneratePassword()
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
		fmt.Printf("New Web UI password: %s\n", password)
		return nil
	},
}

func init() {
	configAddCmd.AddCommand(configAddProviderCmd)
	configAddCmd.AddCommand(configAddGroupCmd)

	configDeleteCmd.AddCommand(configDeleteProviderCmd)
	configDeleteCmd.AddCommand(configDeleteGroupCmd)

	configEditCmd.AddCommand(configEditProviderCmd)
	configEditCmd.AddCommand(configEditGroupCmd)

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configDefaultClientCmd)
	configCmd.AddCommand(configDefaultProfileCmd)
	configCmd.AddCommand(configResetPasswordCmd)
}

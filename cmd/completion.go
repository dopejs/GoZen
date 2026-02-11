package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:       "completion <shell>",
	Short:     "Generate shell completion script",
	Long:      "Generate completion script for zsh, bash, fish, or powershell.",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"zsh", "bash", "fish", "powershell"},
	RunE:      runCompletion,
}

func runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "zsh":
		return rootCmd.GenZshCompletion(os.Stdout)
	case "bash":
		return rootCmd.GenBashCompletion(os.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		cmd.PrintErrf("Unsupported shell: %s\n", args[0])
		return nil
	}
}

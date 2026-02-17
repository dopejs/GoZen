package cmd

import (
	"fmt"
	"strings"

	"github.com/dopejs/gozen/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List providers and profiles",
	Run: func(cmd *cobra.Command, args []string) {
		printProviders()
		fmt.Println()
		printProfiles()
	},
}

var listProviderCmd = &cobra.Command{
	Use:   "provider",
	Short: "List providers",
	Run: func(cmd *cobra.Command, args []string) {
		printProviders()
	},
}

var listGroupCmd = &cobra.Command{
	Use:   "profile",
	Short: "List profiles",
	Run: func(cmd *cobra.Command, args []string) {
		printProfiles()
	},
}

func printProviders() {
	names := config.ProviderNames()
	if len(names) == 0 {
		fmt.Println("No providers configured.")
		return
	}
	fmt.Println("Providers:")
	for _, name := range names {
		p := config.GetProvider(name)
		if p == nil {
			continue
		}
		model := p.Model
		if model == "" {
			model = "-"
		}
		fmt.Printf("  %-14s %s  model=%s\n", name, p.BaseURL, model)
	}
}

func printProfiles() {
	profiles := config.ListProfiles()
	if len(profiles) == 0 {
		fmt.Println("No profiles configured.")
		return
	}
	defaultProfile := config.GetDefaultProfile()
	fmt.Println("Profiles:")
	for _, name := range profiles {
		order, _ := config.ReadProfileOrder(name)
		tag := ""
		if name == defaultProfile {
			tag = " (default)"
		}
		providers := "-"
		if len(order) > 0 {
			providers = strings.Join(order, " → ")
		}
		fmt.Printf("  %-14s %s%s\n", name, providers, tag)
	}
}

func init() {
	listCmd.AddCommand(listProviderCmd)
	listCmd.AddCommand(listGroupCmd)
}

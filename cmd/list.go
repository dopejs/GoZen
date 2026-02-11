package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anthropics/opencc/internal/config"
	"github.com/anthropics/opencc/internal/envfile"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [prefix]",
	Short: "List available configurations",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	prefix := ""
	if len(args) > 0 {
		prefix = args[0]
	}

	configs, err := envfile.ListConfigs()
	if err != nil || len(configs) == 0 {
		if prefix != "" {
			fmt.Printf("No configurations matching '%s*' in %s/\n", prefix, envfile.EnvsPath())
		} else {
			fmt.Printf("No configurations found in %s/\n", envfile.EnvsPath())
			fmt.Printf("Create a .env file, e.g.: %s/work.env\n", envfile.EnvsPath())
		}
		return nil
	}

	// Load fallback order for sorting
	fbOrder, _ := config.ReadFallbackOrder()
	fbMap := make(map[string]int)
	for i, name := range fbOrder {
		fbMap[name] = i + 1
	}

	// Sort: fallback configs first (by order), then the rest alphabetically
	sort.Slice(configs, func(i, j int) bool {
		fi, oki := fbMap[configs[i].Name]
		fj, okj := fbMap[configs[j].Name]
		if oki && okj {
			return fi < fj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return configs[i].Name < configs[j].Name
	})

	found := 0
	for _, cfg := range configs {
		if prefix != "" && !strings.HasPrefix(cfg.Name, prefix) {
			continue
		}

		baseURL := cfg.Get("ANTHROPIC_BASE_URL")
		model := cfg.Get("ANTHROPIC_MODEL")
		if model == "" {
			model = "-"
		}
		if baseURL == "" {
			baseURL = "-"
		}

		fbTag := "    "
		if idx, ok := fbMap[cfg.Name]; ok {
			fbTag = fmt.Sprintf("[%d] ", idx)
		}

		fmt.Printf("  %-12s %s model=%-20s  base_url=%s\n", cfg.Name, fbTag, model, baseURL)
		found++
	}

	if found == 0 {
		if prefix != "" {
			fmt.Printf("No configurations matching '%s*' in %s/\n", prefix, envfile.EnvsPath())
		} else {
			fmt.Printf("No configurations found in %s/\n", envfile.EnvsPath())
		}
	}
	return nil
}

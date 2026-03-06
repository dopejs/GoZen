package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/dopejs/gozen/internal/config"
	"github.com/spf13/cobra"
)

var experienceCmd = &cobra.Command{
	Use:    "experience [feature]",
	Short:  "Manage experimental feature gates",
	Hidden: true, // Not shown in zen --help
	RunE:   runExperience,
}

var closeFlag bool

func init() {
	experienceCmd.Flags().BoolVarP(&closeFlag, "close", "c", false, "disable the feature")
}

func runExperience(cmd *cobra.Command, args []string) error {
	// No arguments: list all features
	if len(args) == 0 {
		return runExperienceList(cmd, args)
	}

	// One argument: enable or disable feature
	feature := strings.ToLower(args[0])
	if closeFlag {
		return runExperienceDisable(cmd, []string{feature})
	}
	return runExperienceEnable(cmd, []string{feature})
}

func runExperienceList(cmd *cobra.Command, args []string) error {
	gates := config.GetFeatureGates()

	fmt.Fprintln(os.Stdout, "Experimental Features:")

	// Define feature metadata
	features := []struct {
		name        string
		enabled     bool
		description string
	}{
		{"bot", gates.Bot, "Bot gateway (BETA)"},
		{"compression", gates.Compression, "Context compression (BETA)"},
		{"middleware", gates.Middleware, "Middleware pipeline (BETA)"},
		{"agent", gates.Agent, "Agent infrastructure (BETA)"},
	}

	// Print each feature with status
	for _, f := range features {
		status := "[disabled]"
		if f.enabled {
			status = "[enabled] "
		}
		fmt.Fprintf(os.Stdout, "  %-12s %s  %s\n", f.name, status, f.description)
	}

	return nil
}

func runExperienceEnable(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("feature name required")
	}

	feature := strings.ToLower(args[0])

	// Validate feature name
	validFeatures := map[string]bool{
		"bot":         true,
		"compression": true,
		"middleware":  true,
		"agent":       true,
	}

	if !validFeatures[feature] {
		return fmt.Errorf("unknown feature: %s (valid: bot, compression, middleware, agent)", feature)
	}

	// Load current gates
	gates := config.GetFeatureGates()
	if gates == nil {
		gates = &config.FeatureGates{}
	}

	// Enable the feature
	switch feature {
	case "bot":
		gates.Bot = true
	case "compression":
		gates.Compression = true
	case "middleware":
		gates.Middleware = true
	case "agent":
		gates.Agent = true
	}

	// Save updated gates
	if err := config.SetFeatureGates(gates); err != nil {
		return fmt.Errorf("failed to enable feature: %w", err)
	}

	// Audit logging
	username := "unknown"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	log.Printf("[AUDIT] action=enable_feature_gate resource=%s user=%s", feature, username)

	fmt.Fprintf(os.Stdout, "Enabled feature: %s\n", feature)
	return nil
}

func runExperienceDisable(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("feature name required")
	}

	feature := strings.ToLower(args[0])

	// Validate feature name
	validFeatures := map[string]bool{
		"bot":         true,
		"compression": true,
		"middleware":  true,
		"agent":       true,
	}

	if !validFeatures[feature] {
		return fmt.Errorf("unknown feature: %s (valid: bot, compression, middleware, agent)", feature)
	}

	// Load current gates
	gates := config.GetFeatureGates()
	if gates == nil {
		gates = &config.FeatureGates{}
	}

	// Disable the feature
	switch feature {
	case "bot":
		gates.Bot = false
	case "compression":
		gates.Compression = false
	case "middleware":
		gates.Middleware = false
	case "agent":
		gates.Agent = false
	}

	// Save updated gates
	if err := config.SetFeatureGates(gates); err != nil {
		return fmt.Errorf("failed to disable feature: %w", err)
	}

	// Audit logging
	username := "unknown"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	log.Printf("[AUDIT] action=disable_feature_gate resource=%s user=%s", feature, username)

	fmt.Fprintf(os.Stdout, "Disabled feature: %s\n", feature)
	return nil
}

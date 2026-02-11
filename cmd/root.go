package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/anthropics/opencc/internal/config"
	"github.com/anthropics/opencc/internal/envfile"
	"github.com/anthropics/opencc/internal/proxy"
	"github.com/anthropics/opencc/tui"
	"github.com/spf13/cobra"
)

// stdinReader is the reader used for interactive prompts. Tests can replace it.
var stdinReader io.Reader = os.Stdin

var Version = "1.2.0"

var rootCmd = &cobra.Command{
	Use:   "opencc [claude args...]",
	Short: "Claude Code environment switcher with fallback proxy",
	Long:  "Load environment variables and start Claude Code, optionally with a fallback proxy.",
	// Allow unknown flags to pass through to claude
	DisableFlagParsing: false,
	SilenceUsage:       true,
	SilenceErrors:      true,
	RunE:               runProxy,
}

func init() {
	rootCmd.Flags().StringP("fallback", "f", "", "fallback profile name (use -f without value to pick interactively)")
	rootCmd.Flags().Lookup("fallback").NoOptDefVal = " "
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(pickCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func runProxy(cmd *cobra.Command, args []string) error {
	profileFlag, _ := cmd.Flags().GetString("fallback")

	providerNames, profile, err := resolveProviderNames(profileFlag)
	if err != nil {
		return err
	}

	providerNames, err = validateProviderNames(providerNames, profile)
	if err != nil {
		return err
	}

	return startProxy(providerNames, args)
}

func startProxy(names []string, args []string) error {
	providers, firstModel, err := buildProviders(names)
	if err != nil {
		return err
	}

	// Set up logger
	logDir := envfile.EnvsPath()
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logDir, "proxy.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logFile = nil
	}

	var logger *log.Logger
	if logFile != nil {
		logger = log.New(logFile, "", log.LstdFlags)
		defer logFile.Close()
	} else {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	logger.Printf("Starting proxy with %d providers:", len(providers))
	for i, p := range providers {
		logger.Printf("  [%d] %s → %s (model=%s)", i+1, p.Name, p.BaseURL.String(), p.Model)
	}

	// Start embedded proxy
	port, err := proxy.StartProxy(providers, "127.0.0.1:0", logger)
	if err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	logger.Printf("Proxy listening on 127.0.0.1:%d", port)

	// Set environment for claude
	os.Setenv("ANTHROPIC_BASE_URL", fmt.Sprintf("http://127.0.0.1:%d", port))
	os.Setenv("ANTHROPIC_AUTH_TOKEN", "opencc-proxy")
	if firstModel != "" {
		os.Setenv("ANTHROPIC_MODEL", firstModel)
	}

	// Find claude binary
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}

	// Start claude as subprocess (not exec, so proxy stays alive)
	claudeCmd := exec.Command(claudeBin, args...)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	// Forward signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			if claudeCmd.Process != nil {
				claudeCmd.Process.Signal(sig)
			}
		}
	}()

	if err := claudeCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func buildProviders(names []string) ([]*proxy.Provider, string, error) {
	var providers []*proxy.Provider
	var firstModel string

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		cfg, err := envfile.LoadByName(name)
		if err != nil {
			return nil, "", fmt.Errorf("configuration '%s' not found", name)
		}

		baseURL := cfg.Get("ANTHROPIC_BASE_URL")
		token := cfg.Get("ANTHROPIC_AUTH_TOKEN")
		model := cfg.Get("ANTHROPIC_MODEL")

		if baseURL == "" || token == "" {
			return nil, "", fmt.Errorf("%s.env missing ANTHROPIC_BASE_URL or ANTHROPIC_AUTH_TOKEN", name)
		}
		if model == "" {
			model = "claude-sonnet-4-20250514"
		}
		if firstModel == "" {
			firstModel = model
		}

		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, "", fmt.Errorf("invalid URL for provider %s: %w", name, err)
		}

		providers = append(providers, &proxy.Provider{
			Name:    name,
			BaseURL: u,
			Token:   token,
			Model:   model,
			Healthy: true,
		})
	}

	if len(providers) == 0 {
		return nil, "", fmt.Errorf("no valid providers")
	}
	return providers, firstModel, nil
}

// resolveProviderNames determines the provider list based on the -f flag value.
// Returns the provider names and the profile used.
func resolveProviderNames(profileFlag string) ([]string, string, error) {
	// -f (no value, NoOptDefVal=" ") → interactive profile picker
	if profileFlag == " " {
		profile, err := tui.RunProfilePicker()
		if err != nil {
			return nil, "", err
		}
		names, err := config.ReadProfileOrder(profile)
		if err != nil {
			return nil, "", fmt.Errorf("profile '%s' has no providers configured", profile)
		}
		if len(names) == 0 {
			return nil, "", fmt.Errorf("profile '%s' has no providers configured", profile)
		}
		return names, profile, nil
	}

	// -f <name> → use that specific profile
	if profileFlag != "" {
		names, err := config.ReadProfileOrder(profileFlag)
		if err != nil {
			return nil, "", fmt.Errorf("profile '%s' not found", profileFlag)
		}
		if len(names) == 0 {
			return nil, "", fmt.Errorf("profile '%s' has no providers configured", profileFlag)
		}
		return names, profileFlag, nil
	}

	// No flag → existing behavior (default profile, or interactive selection)
	fbNames, err := config.ReadFallbackOrder()
	if err == nil && len(fbNames) > 0 {
		return fbNames, "default", nil
	}

	// fallback.conf missing or empty — interactive selection
	names, err := interactiveSelectProviders()
	if err != nil {
		return nil, "", err
	}
	return names, "default", nil
}

// interactiveSelectProviders uses TUI to select providers.
// If no providers exist, launches the create-first editor.
// Otherwise launches the checkbox picker.
func interactiveSelectProviders() ([]string, error) {
	available := envfile.ConfigNames()
	if len(available) == 0 {
		// No providers at all — launch TUI editor to create one
		name, err := tui.RunCreateFirst()
		if err != nil {
			return nil, fmt.Errorf("no providers configured")
		}
		if name == "" {
			return nil, fmt.Errorf("no providers configured")
		}
		return []string{name}, nil
	}

	// Providers exist but no fallback.conf — launch picker
	selected, err := tui.RunPick()
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no providers selected")
	}

	// Write selection to fallback.conf
	if err := config.WriteFallbackOrder(selected); err != nil {
		return nil, fmt.Errorf("failed to save fallback order: %w", err)
	}
	fmt.Printf("Saved fallback order: %s\n", strings.Join(selected, ", "))

	return selected, nil
}

// validateProviderNames checks that each provider .env exists.
// Prompts user to confirm removal of missing providers from the profile's conf.
func validateProviderNames(names []string, profile string) ([]string, error) {
	var valid, missing []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		path := filepath.Join(envfile.EnvsPath(), name+".env")
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, name)
		} else {
			valid = append(valid, name)
		}
	}

	if len(missing) == 0 {
		return names, nil
	}

	fmt.Printf("%s provider(s) not found. Continue and remove from profile? (y/n): ", strings.Join(missing, ", "))
	reader := bufio.NewReader(stdinReader)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(line))
	if answer != "y" && answer != "yes" {
		return nil, fmt.Errorf("aborted")
	}

	// Remove missing from profile
	for _, name := range missing {
		config.RemoveFromProfileOrder(profile, name)
	}

	if len(valid) == 0 {
		return nil, fmt.Errorf("no valid providers remaining. Run 'opencc config' to set up providers")
	}

	return valid, nil
}

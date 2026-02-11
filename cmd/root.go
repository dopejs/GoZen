package cmd

import (
	"fmt"
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
	"github.com/spf13/cobra"
)

var Version = "1.1.0"

var fallbackFlag string

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
	rootCmd.Flags().StringVarP(&fallbackFlag, "fallback", "f", "", "provider list (comma-separated), or reads fallback.conf if flag present without value")
	// Make -f work with optional value
	rootCmd.Flags().Lookup("fallback").NoOptDefVal = " "

	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func runProxy(cmd *cobra.Command, args []string) error {
	// Determine provider list
	providerNames, err := resolveProviderNames(cmd.Flags().Changed("fallback"), fallbackFlag)
	if err != nil {
		return err
	}

	// Build provider list
	providers, firstModel, err := buildProviders(providerNames)
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

// resolveProviderNames determines the provider list from flags or fallback.conf.
func resolveProviderNames(fallbackChanged bool, fallbackValue string) ([]string, error) {
	var names []string

	if fallbackChanged {
		if v := strings.TrimSpace(fallbackValue); v != "" {
			names = strings.Split(v, ",")
		}
	}

	if len(names) == 0 {
		fbNames, err := config.ReadFallbackOrder()
		if err != nil {
			return nil, fmt.Errorf("no providers specified and fallback.conf not found: %w", err)
		}
		names = fbNames
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no providers configured. Run 'opencc config' to set up providers")
	}

	return names, nil
}

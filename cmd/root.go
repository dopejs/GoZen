package cmd

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/daemon"
	"github.com/dopejs/gozen/internal/proxy"
	"github.com/dopejs/gozen/internal/update"
	"github.com/dopejs/gozen/tui"
	"github.com/spf13/cobra"
)

// stdinReader is the reader used for interactive prompts. Tests can replace it.
var stdinReader io.Reader = os.Stdin

var Version = "2.1.1"

var updateChecker *update.Checker

var rootCmd = &cobra.Command{
	Use:   "zen [cli args...]",
	Short: "Multi-CLI environment switcher with proxy failover",
	Long:  "Load environment variables and start CLI (Claude Code, Codex, or OpenCode) with proxy failover.",
	// Allow unknown flags to pass through to claude
	DisableFlagParsing: false,
	SilenceUsage:       true,
	SilenceErrors:      true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip update check for commands where it's not useful
		name := cmd.Name()
		if name == "upgrade" || name == "version" || name == "completion" {
			return
		}
		updateChecker = update.NewChecker(Version)
		updateChecker.Start()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if updateChecker != nil {
			if msg := updateChecker.Notification(); msg != "" {
				fmt.Fprint(os.Stderr, msg)
			}
		}
	},
	RunE: runProxy,
}

var clientFlag string
var yesFlag bool

func init() {
	// -p/--profile is the new flag, -f/--fallback is kept for backward compatibility but hidden
	rootCmd.Flags().StringP("profile", "p", "", "profile name")
	rootCmd.Flags().StringP("fallback", "f", "", "alias for --profile (deprecated)")
	rootCmd.Flags().Lookup("fallback").Hidden = true
	rootCmd.Flags().StringVarP(&clientFlag, "client", "c", "", "client to use (claude, codex, opencode)")
	rootCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "auto-approve CLI permissions (claude --permission-mode acceptEdits, codex -a never)")
	rootCmd.Flags().String("cli", "", "alias for --client (deprecated)")
	rootCmd.Flags().Lookup("cli").Hidden = true
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(pickCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(bindCmd)
	rootCmd.AddCommand(unbindCmd)
	rootCmd.AddCommand(statusCmd)

	// Set custom help function only for root command
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd {
			cmd.Println(rootHelpText(cmd))
		} else {
			defaultHelp(cmd, args)
		}
	})
}

func rootHelpText(cmd *cobra.Command) string {
	return fmt.Sprintf(`%s

Usage:
  %s
  %s [command]

Quick Start:
  zen                       Start with default profile
  zen -p <profile>          Start with specific profile
  zen --cli codex           Start with specific CLI
  zen -y                    Start with auto-approve permissions

Configuration:
  config add provider [name]   Add a new provider
  config add profile [name]    Add a new profile
  config edit provider <name>  Edit an existing provider
  config edit profile <name>   Edit an existing profile
  config delete provider <name> Delete a provider
  config delete profile <name>  Delete a profile
  config default-client        Set the default client
  config default-profile       Set the default profile
  config reset-password        Reset Web UI access password

Project Binding:
  bind <profile>               Bind current directory to a profile
  bind --cli <cli>             Bind current directory to a CLI
  unbind                       Remove binding for current directory
  status                       Show binding status

Web Interface:
  web                          Open web UI in browser (starts daemon if needed)

Other Commands:
  list                         List all providers and profiles
  pick                         Interactively select providers
  use <provider>               Use a specific provider directly
  upgrade                      Upgrade to latest version
  version                      Show version
  completion                   Generate shell completion

Flags:
%s
Use "%s [command] --help" for more information about a command.`,
		cmd.Long,
		cmd.UseLine(),
		cmd.CommandPath(),
		cmd.LocalFlags().FlagUsages(),
		cmd.CommandPath())
}

func Execute() error {
	return rootCmd.Execute()
}

func runProxy(cmd *cobra.Command, args []string) error {
	// Support both -p/--profile (new) and -f/--fallback (deprecated)
	profileFlag, _ := cmd.Flags().GetString("profile")
	if profileFlag == "" {
		profileFlag, _ = cmd.Flags().GetString("fallback")
	}

	// Support --cli as alias for --client
	if clientFlag == "" {
		clientFlag, _ = cmd.Flags().GetString("cli")
	}

	providerNames, profile, client, err := resolveProviderNamesAndClient(profileFlag, clientFlag)
	if err != nil {
		return err
	}

	providerNames, err = validateProviderNames(providerNames, profile)
	if err != nil {
		return err
	}

	// Use zend daemon
	return startViaDaemon(profile, client, providerNames, args, yesFlag)
}

// startViaDaemon starts a client session through the zend daemon.
// 1. Ensure zend is running (auto-start if needed)
// 2. Generate session UUID
// 3. Set base URL to http://127.0.0.1:<proxy_port>/<profile>/<session>/v1
// 4. Merge provider env vars
// 5. Exec client binary
func startViaDaemon(profile, client string, providerNames []string, args []string, autoApprove bool) error {
	if err := ensureDaemonRunning(); err != nil {
		return fmt.Errorf("failed to start zend: %w", err)
	}

	clientBin := client
	if clientBin == "" {
		clientBin = "claude"
	}

	// Generate session UUID
	sessionID := generateSessionID()

	// Build proxy URL with profile and session in path
	proxyPort := config.GetProxyPort()
	baseURL := fmt.Sprintf("http://127.0.0.1:%d/%s/%s", proxyPort, profile, sessionID)

	// Merge env_vars from all providers for this client
	providers, err := buildProviders(providerNames)
	if err != nil {
		return err
	}
	mergedEnvVars := mergeProviderEnvVarsForCLI(providers, clientBin)
	for k, v := range mergedEnvVars {
		os.Setenv(k, v)
	}

	// Set environment variables based on client type
	logger := log.New(io.Discard, "", 0)
	setupClientEnvironment(clientBin, baseURL, logger)

	// Set X-Zen-Client header via env var (proxy strips it)
	os.Setenv("X_ZEN_CLIENT", clientBin)

	// Find client binary
	cliPath, err := exec.LookPath(clientBin)
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", clientBin, err)
	}

	// Inject auto-approve flags based on client type
	if autoApprove {
		args = prependAutoApproveArgs(clientBin, args)
	}

	// Start client as subprocess
	cliCmd := exec.Command(cliPath, args...)
	cliCmd.Stdin = os.Stdin
	cliCmd.Stdout = os.Stdout
	cliCmd.Stderr = os.Stderr

	// Forward signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			if cliCmd.Process != nil {
				cliCmd.Process.Signal(sig)
			}
		}
	}()

	if err := cliCmd.Run(); err != nil {
		signal.Stop(sigCh)
		close(sigCh)
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	signal.Stop(sigCh)
	close(sigCh)
	return nil
}

// ensureDaemonRunning checks if zend is running and starts it if not.
func ensureDaemonRunning() error {
	if _, running := daemon.IsDaemonRunning(); running {
		return nil
	}

	// Auto-start the daemon
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	logPath := daemon.DaemonLogPath()
	logDir := config.ConfigDirPath()
	os.MkdirAll(logDir, 0755)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("cannot open log file: %w", err)
	}
	defer logFile.Close()

	child := exec.Command(exe, "daemon", "start")
	child.Env = append(os.Environ(), "GOZEN_DAEMON=1")
	child.Stdout = logFile
	child.Stderr = logFile
	child.SysProcAttr = daemon.DaemonSysProcAttr()

	if err := child.Start(); err != nil {
		return fmt.Errorf("failed to start zend: %w", err)
	}

	daemon.WriteDaemonPid(child.Process.Pid)

	// Wait for daemon to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return waitForDaemonReady(ctx)
}

// generateSessionID generates a short random hex session ID.
func generateSessionID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use time-based
		t := time.Now().UnixNano()
		b[0] = byte(t >> 24)
		b[1] = byte(t >> 16)
		b[2] = byte(t >> 8)
		b[3] = byte(t)
	}
	return hex.EncodeToString(b)
}

func startLegacyProxy(names []string, pc *config.ProfileConfig, cli string, args []string) error {
	providers, err := buildProviders(names)
	if err != nil {
		return err
	}

	// Set up logger
	logDir := config.ConfigDirPath()
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

	// Initialize structured logger for web API access
	if err := proxy.InitGlobalLogger(logDir); err != nil {
		logger.Printf("Warning: failed to initialize structured logger: %v", err)
	}

	logger.Printf("Starting proxy with %d providers:", len(providers))
	for i, p := range providers {
		logger.Printf("  [%d] %s → %s (model=%s)", i+1, p.Name, p.BaseURL.String(), p.Model)
	}

	// Use CLI from parameter (already resolved from flag/binding/default)
	cliBin := cli
	if cliBin == "" {
		cliBin = "claude"
	}

	// Determine client format based on CLI type
	clientFormat := GetClientFormat(GetClientType(cliBin))
	logger.Printf("CLI: %s, Client format: %s", cliBin, clientFormat)

	// Start proxy — with routing if configured, otherwise plain
	var port int
	if pc != nil && len(pc.Routing) > 0 {
		routingCfg, err := buildRoutingConfig(pc, providers, logger)
		if err != nil {
			return fmt.Errorf("failed to build routing config: %w", err)
		}
		port, err = proxy.StartProxyWithRouting(routingCfg, clientFormat, "127.0.0.1:0", logger)
		if err != nil {
			return fmt.Errorf("failed to start proxy: %w", err)
		}
	} else {
		port, err = proxy.StartProxy(providers, clientFormat, "127.0.0.1:0", logger)
		if err != nil {
			return fmt.Errorf("failed to start proxy: %w", err)
		}
	}

	logger.Printf("Proxy listening on 127.0.0.1:%d", port)

	// Merge env_vars from all providers for this specific CLI
	// For numeric values like ANTHROPIC_MAX_CONTEXT_WINDOW, use the minimum value
	// This ensures the CLI respects the most restrictive provider's limit
	mergedEnvVars := mergeProviderEnvVarsForCLI(providers, cliBin)
	for k, v := range mergedEnvVars {
		os.Setenv(k, v)
		logger.Printf("Setting env: %s=%s", k, v)
	}

	// Set environment variables based on CLI type
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	setupClientEnvironment(cliBin, proxyURL, logger)

	// Find CLI binary
	cliPath, err := exec.LookPath(cliBin)
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", cliBin, err)
	}

	// Start CLI as subprocess (not exec, so proxy stays alive)
	cliCmd := exec.Command(cliPath, args...)
	cliCmd.Stdin = os.Stdin
	cliCmd.Stdout = os.Stdout
	cliCmd.Stderr = os.Stderr

	// Forward signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			if cliCmd.Process != nil {
				cliCmd.Process.Signal(sig)
			}
		}
	}()

	if err := cliCmd.Run(); err != nil {
		signal.Stop(sigCh)
		close(sigCh)
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	signal.Stop(sigCh)
	close(sigCh)
	return nil
}

func buildProviders(names []string) ([]*proxy.Provider, error) {
	var providers []*proxy.Provider

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		p := config.GetProvider(name)
		if p == nil {
			return nil, fmt.Errorf("configuration '%s' not found", name)
		}

		if p.BaseURL == "" || p.AuthToken == "" {
			return nil, fmt.Errorf("%s missing base_url or auth_token", name)
		}

		model := p.Model
		if model == "" {
			model = "claude-sonnet-4-5"
		}
		reasoningModel := p.ReasoningModel
		if reasoningModel == "" {
			reasoningModel = "claude-sonnet-4-5-thinking"
		}
		haikuModel := p.HaikuModel
		if haikuModel == "" {
			haikuModel = "claude-haiku-4-5"
		}
		opusModel := p.OpusModel
		if opusModel == "" {
			opusModel = "claude-opus-4-5"
		}
		sonnetModel := p.SonnetModel
		if sonnetModel == "" {
			sonnetModel = "claude-sonnet-4-5"
		}

		u, err := url.Parse(p.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL for provider %s: %w", name, err)
		}

		providers = append(providers, &proxy.Provider{
			Name:            name,
			Type:            p.GetType(),
			BaseURL:         u,
			Token:           p.AuthToken,
			Model:           model,
			ReasoningModel:  reasoningModel,
			HaikuModel:      haikuModel,
			OpusModel:       opusModel,
			SonnetModel:     sonnetModel,
			EnvVars:         p.EnvVars,
			ClaudeEnvVars:   p.ClaudeEnvVars,
			CodexEnvVars:    p.CodexEnvVars,
			OpenCodeEnvVars: p.OpenCodeEnvVars,
			Healthy:         true,
		})
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no valid providers")
	}
	return providers, nil
}

// mergeProviderEnvVarsForCLI merges env_vars from all providers for a specific CLI.
// For numeric values like ANTHROPIC_MAX_CONTEXT_WINDOW, uses the minimum value.
// For other values, first provider's value takes precedence.
func mergeProviderEnvVarsForCLI(providers []*proxy.Provider, cli string) map[string]string {
	result := make(map[string]string)

	// Env vars where we should take the minimum numeric value
	minValueKeys := map[string]bool{
		"ANTHROPIC_MAX_CONTEXT_WINDOW":          true,
		"OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": true,
	}

	for _, p := range providers {
		envVars := p.GetEnvVarsForClient(cli)
		if envVars == nil {
			continue
		}
		for k, v := range envVars {
			if k == "" || v == "" {
				continue
			}

			existing, exists := result[k]
			if !exists {
				result[k] = v
				continue
			}

			// For min-value keys, compare and keep the smaller value
			if minValueKeys[k] {
				existingVal, err1 := strconv.Atoi(existing)
				newVal, err2 := strconv.Atoi(v)
				if err1 == nil && err2 == nil && newVal < existingVal {
					result[k] = v
				}
			}
			// For other keys, first value wins (already set)
		}
	}

	return result
}

// buildRoutingConfig creates a RoutingConfig from a ProfileConfig.
// Provider instances are shared across scenarios: same name → same *Provider pointer.
func buildRoutingConfig(pc *config.ProfileConfig, defaultProviders []*proxy.Provider, logger *log.Logger) (*proxy.RoutingConfig, error) {
	// Build a map of all provider instances by name (from default providers)
	providerMap := make(map[string]*proxy.Provider)
	for _, p := range defaultProviders {
		providerMap[p.Name] = p
	}

	// Also build providers for any names that only appear in routing scenarios
	for _, route := range pc.Routing {
		for _, pr := range route.Providers {
			if _, ok := providerMap[pr.Name]; !ok {
				// Need to build this provider
				ps, err := buildProviders([]string{pr.Name})
				if err != nil {
					logger.Printf("[routing] skipping unknown provider %q in routing: %v", pr.Name, err)
					continue
				}
				providerMap[pr.Name] = ps[0]
			}
		}
	}

	// Build scenario routes
	scenarioRoutes := make(map[config.Scenario]*proxy.ScenarioProviders)
	for scenario, route := range pc.Routing {
		var chain []*proxy.Provider
		models := make(map[string]string)
		for _, pr := range route.Providers {
			if p, ok := providerMap[pr.Name]; ok {
				chain = append(chain, p)
				if pr.Model != "" {
					models[pr.Name] = pr.Model
				}
			}
		}
		if len(chain) > 0 {
			scenarioRoutes[scenario] = &proxy.ScenarioProviders{
				Providers: chain,
				Models:    models,
			}
			logger.Printf("[routing] scenario %s: %d providers, %d model overrides", scenario, len(chain), len(models))
		}
	}

	return &proxy.RoutingConfig{
		DefaultProviders:     defaultProviders,
		ScenarioRoutes:       scenarioRoutes,
		LongContextThreshold: pc.LongContextThreshold,
	}, nil
}

// resolveProviderNamesAndClient determines the provider list and client based on flags and bindings.
// Returns the provider names, the profile used, and the client to use.
func resolveProviderNamesAndClient(profileFlag string, clientFlag string) ([]string, string, string, error) {
	// Determine CLI: flag > binding > default
	cli := clientFlag

	// -p <name> → use that specific profile
	if profileFlag != "" {
		names, err := config.ReadProfileOrder(profileFlag)
		if err != nil {
			return nil, "", "", fmt.Errorf("profile '%s' not found", profileFlag)
		}
		if len(names) == 0 {
			return nil, "", "", fmt.Errorf("profile '%s' has no providers configured", profileFlag)
		}
		if cli == "" {
			cli = config.GetDefaultClient()
		}
		return names, profileFlag, cli, nil
	}

	// No profile flag → check for project binding first
	cwd, err := os.Getwd()
	if err == nil {
		cwd = filepath.Clean(cwd)
		if binding := config.GetProjectBinding(cwd); binding != nil {
			// Found project binding
			profile := binding.Profile
			if profile == "" {
				profile = config.GetDefaultProfile()
			}

			// Use binding CLI if not overridden by flag
			if cli == "" && binding.Client != "" {
				cli = binding.Client
			}

			names, err := config.ReadProfileOrder(profile)
			if err == nil && len(names) > 0 {
				if cli == "" {
					cli = config.GetDefaultClient()
				}
				return names, profile, cli, nil
			}
			// Profile was deleted, fall through to default
			if binding.Profile != "" {
				fmt.Fprintf(os.Stderr, "Warning: Bound profile '%s' not found, using default\n", binding.Profile)
			}
		}
	}

	// No binding → use default profile
	defaultProfile := config.GetDefaultProfile()
	fbNames, err := config.ReadFallbackOrder()
	if err == nil && len(fbNames) > 0 {
		if cli == "" {
			cli = config.GetDefaultClient()
		}
		return fbNames, defaultProfile, cli, nil
	}

	// default profile missing or empty — interactive selection
	names, err := interactiveSelectProviders()
	if err != nil {
		return nil, "", "", err
	}
	if names == nil {
		// User cancelled
		return nil, "", "", fmt.Errorf("cancelled")
	}
	if cli == "" {
		cli = config.GetDefaultClient()
	}
	return names, defaultProfile, cli, nil
}

// interactiveSelectProviders uses TUI to select providers.
// If no providers exist, launches the create-first editor.
// Otherwise launches the checkbox picker.
// Returns nil, nil if user cancels.
func interactiveSelectProviders() ([]string, error) {
	available := config.ProviderNames()
	if len(available) == 0 {
		// No providers at all — launch TUI editor to create one
		name, err := tui.RunCreateFirst()
		if err != nil {
			// User cancelled
			return nil, nil
		}
		if name == "" {
			return nil, nil
		}
		return []string{name}, nil
	}

	// Providers exist but no default profile — launch picker
	selected, err := tui.RunPick()
	if err != nil {
		// User cancelled
		return nil, nil
	}
	if len(selected) == 0 {
		return nil, nil
	}

	// Write selection to default profile
	if err := config.WriteFallbackOrder(selected); err != nil {
		return nil, fmt.Errorf("failed to save fallback order: %w", err)
	}
	fmt.Printf("Saved fallback order: %s\n", strings.Join(selected, ", "))

	return selected, nil
}

// validateProviderNames checks that each provider exists in the config.
// Prompts user to confirm removal of missing providers from the profile.
func validateProviderNames(names []string, profile string) ([]string, error) {
	var valid, missing []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if config.GetProvider(name) == nil {
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
		return nil, fmt.Errorf("no valid providers remaining. Run 'zen config' to set up providers")
	}

	return valid, nil
}

// ClientType represents the type of client being used.
type ClientType string

const (
	ClientClaude   ClientType = "claude"
	ClientCodex    ClientType = "codex"
	ClientOpenCode ClientType = "opencode"
)

// GetClientType returns the client type from the binary name.
func GetClientType(clientBin string) ClientType {
	switch clientBin {
	case "codex":
		return ClientCodex
	case "opencode":
		return ClientOpenCode
	default:
		return ClientClaude
	}
}

// GetClientFormat returns the API format used by the client.
func GetClientFormat(clientType ClientType) string {
	switch clientType {
	case ClientCodex:
		return config.ProviderTypeOpenAI
	default:
		// Claude Code and OpenCode use Anthropic format by default
		return config.ProviderTypeAnthropic
	}
}

// prependAutoApproveArgs prepends the appropriate auto-approve flags for each CLI.
// Claude Code: --permission-mode acceptEdits, Codex: -a never, OpenCode: auto-approves by default (no flag needed).
func prependAutoApproveArgs(clientBin string, args []string) []string {
	switch GetClientType(clientBin) {
	case ClientClaude:
		return append([]string{"--permission-mode", "acceptEdits"}, args...)
	case ClientCodex:
		return append([]string{"-a", "never"}, args...)
	default:
		// OpenCode auto-approves by default
		return args
	}
}

// setupClientEnvironment sets the appropriate environment variables for the client.
func setupClientEnvironment(clientBin string, proxyURL string, logger *log.Logger) {
	clientType := GetClientType(clientBin)

	switch clientType {
	case ClientCodex:
		// Codex uses OpenAI environment variables
		os.Setenv("OPENAI_BASE_URL", proxyURL)
		os.Setenv("OPENAI_API_KEY", "zen-proxy")
		logger.Printf("Setting Codex env: OPENAI_BASE_URL=%s", proxyURL)

	case ClientOpenCode:
		// OpenCode supports multiple providers, set both
		// It will use the appropriate one based on the model prefix
		os.Setenv("ANTHROPIC_BASE_URL", proxyURL)
		os.Setenv("ANTHROPIC_API_KEY", "zen-proxy")
		os.Setenv("OPENAI_BASE_URL", proxyURL)
		os.Setenv("OPENAI_API_KEY", "zen-proxy")
		logger.Printf("Setting OpenCode env: ANTHROPIC_BASE_URL=%s, OPENAI_BASE_URL=%s", proxyURL, proxyURL)

	default:
		// Claude Code uses Anthropic environment variables
		os.Setenv("ANTHROPIC_BASE_URL", proxyURL)
		os.Setenv("ANTHROPIC_AUTH_TOKEN", "zen-proxy")
		logger.Printf("Setting Claude env: ANTHROPIC_BASE_URL=%s", proxyURL)
	}
}


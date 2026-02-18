package config

import (
	"bytes"
	"encoding/json"
	"os"
)

const (
	ConfigDir  = ".zen"
	ConfigFile = "zen.json"
	LegacyDir  = ".cc_envs"

	// LegacyOpenCCDir is the old config directory name used before the GoZen rename.
	LegacyOpenCCDir  = ".opencc"
	LegacyOpenCCFile = "opencc.json"

	DefaultWebPort   = 19840
	DefaultProxyPort = 19841
	DaemonPidFile    = "zend.pid"
	DaemonLogFile    = "zend.log"

	DefaultProfileName  = "default"
	DefaultClientName   = "claude"

	// Supported client names
	ClientClaude   = "claude"
	ClientCodex    = "codex"
	ClientOpenCode = "opencode"

	// Provider API types
	ProviderTypeAnthropic = "anthropic"
	ProviderTypeOpenAI    = "openai"
)

// AvailableClients is the canonical list of supported client names.
var AvailableClients = []string{ClientClaude, ClientCodex, ClientOpenCode}

// IsValidClient reports whether name is a supported client name.
func IsValidClient(name string) bool {
	for _, c := range AvailableClients {
		if c == name {
			return true
		}
	}
	return false
}

// ProviderConfig holds connection and model settings for a single API provider.
type ProviderConfig struct {
	Type           string            `json:"type,omitempty"` // "anthropic" (default) or "openai"
	BaseURL        string            `json:"base_url"`
	AuthToken      string            `json:"auth_token"`
	Model          string            `json:"model,omitempty"`
	ReasoningModel string            `json:"reasoning_model,omitempty"`
	HaikuModel     string            `json:"haiku_model,omitempty"`
	OpusModel      string            `json:"opus_model,omitempty"`
	SonnetModel    string            `json:"sonnet_model,omitempty"`
	EnvVars        map[string]string `json:"env_vars,omitempty"`          // Claude Code env vars (legacy, for backward compat)
	ClaudeEnvVars  map[string]string `json:"claude_env_vars,omitempty"`   // Claude Code specific env vars
	CodexEnvVars   map[string]string `json:"codex_env_vars,omitempty"`    // Codex specific env vars
	OpenCodeEnvVars map[string]string `json:"opencode_env_vars,omitempty"` // OpenCode specific env vars
}

// GetType returns the provider type, defaulting to "anthropic".
func (p *ProviderConfig) GetType() string {
	if p.Type == "" {
		return ProviderTypeAnthropic
	}
	return p.Type
}

// GetEnvVarsForClient returns the environment variables for a specific client.
// Falls back to legacy EnvVars if client-specific vars are not set.
func (p *ProviderConfig) GetEnvVarsForClient(client string) map[string]string {
	switch client {
	case "codex":
		if len(p.CodexEnvVars) > 0 {
			return p.CodexEnvVars
		}
	case "opencode":
		if len(p.OpenCodeEnvVars) > 0 {
			return p.OpenCodeEnvVars
		}
	default: // claude
		if len(p.ClaudeEnvVars) > 0 {
			return p.ClaudeEnvVars
		}
	}
	// Fallback to legacy EnvVars
	return p.EnvVars
}

// ExportToEnv sets all ANTHROPIC_* environment variables from this provider config.
func (p *ProviderConfig) ExportToEnv() {
	os.Setenv("ANTHROPIC_BASE_URL", p.BaseURL)
	os.Setenv("ANTHROPIC_AUTH_TOKEN", p.AuthToken)

	// Clear optional model vars first to avoid stale values from previous provider
	os.Unsetenv("ANTHROPIC_MODEL")
	os.Unsetenv("ANTHROPIC_REASONING_MODEL")
	os.Unsetenv("ANTHROPIC_DEFAULT_HAIKU_MODEL")
	os.Unsetenv("ANTHROPIC_DEFAULT_OPUS_MODEL")
	os.Unsetenv("ANTHROPIC_DEFAULT_SONNET_MODEL")

	if p.Model != "" {
		os.Setenv("ANTHROPIC_MODEL", p.Model)
	}
	if p.ReasoningModel != "" {
		os.Setenv("ANTHROPIC_REASONING_MODEL", p.ReasoningModel)
	}
	if p.HaikuModel != "" {
		os.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", p.HaikuModel)
	}
	if p.OpusModel != "" {
		os.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", p.OpusModel)
	}
	if p.SonnetModel != "" {
		os.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", p.SonnetModel)
	}

	// Export custom environment variables
	for k, v := range p.EnvVars {
		if k != "" && v != "" {
			os.Setenv(k, v)
		}
	}
}

// Scenario represents a request scenario for routing decisions.
type Scenario string

const (
	ScenarioThink       Scenario = "think"
	ScenarioImage       Scenario = "image"
	ScenarioLongContext Scenario = "longContext"
	ScenarioWebSearch   Scenario = "webSearch"
	ScenarioBackground  Scenario = "background"
	ScenarioDefault     Scenario = "default"
)

// ProviderRoute defines a provider and its optional model override in a scenario.
type ProviderRoute struct {
	Name  string `json:"name"`
	Model string `json:"model,omitempty"`
}

// ScenarioRoute defines providers and their model overrides for a scenario.
type ScenarioRoute struct {
	Providers []*ProviderRoute `json:"providers"`
}

// UnmarshalJSON supports both old format (providers: ["p1"], model: "m") and new format (providers: [{name, model}]).
func (sr *ScenarioRoute) UnmarshalJSON(data []byte) error {
	// Try new format first
	type scenarioRouteAlias struct {
		Providers []*ProviderRoute `json:"providers"`
	}
	var alias scenarioRouteAlias
	if err := json.Unmarshal(data, &alias); err == nil && len(alias.Providers) > 0 {
		// Check if first provider is actually a ProviderRoute (has Name field)
		if alias.Providers[0].Name != "" {
			sr.Providers = alias.Providers
			return nil
		}
	}

	// Try old format: {providers: ["p1", "p2"], model: "m"}
	var oldFormat struct {
		Providers []string `json:"providers"`
		Model     string   `json:"model,omitempty"`
	}
	if err := json.Unmarshal(data, &oldFormat); err != nil {
		return err
	}

	// Convert old format to new
	sr.Providers = make([]*ProviderRoute, len(oldFormat.Providers))
	for i, name := range oldFormat.Providers {
		sr.Providers[i] = &ProviderRoute{
			Name:  name,
			Model: oldFormat.Model, // All providers share the same model in old format
		}
	}
	return nil
}

// ProviderNames returns the list of provider names in order.
func (sr *ScenarioRoute) ProviderNames() []string {
	names := make([]string, len(sr.Providers))
	for i, pr := range sr.Providers {
		names[i] = pr.Name
	}
	return names
}

// ModelForProvider returns the model override for a specific provider, or empty string.
func (sr *ScenarioRoute) ModelForProvider(name string) string {
	for _, pr := range sr.Providers {
		if pr.Name == name {
			return pr.Model
		}
	}
	return ""
}

// ProfileConfig holds a profile's provider list and optional scenario routing.
type ProfileConfig struct {
	Providers            []string                    `json:"providers"`
	Routing              map[Scenario]*ScenarioRoute `json:"routing,omitempty"`
	LongContextThreshold int                         `json:"long_context_threshold,omitempty"` // defaults to 32000 if not set
	Strategy             LoadBalanceStrategy         `json:"strategy,omitempty"`               // load balancing strategy
}

// UnmarshalJSON supports both old format (["p1","p2"]) and new format ({providers: [...], routing: {...}}).
func (pc *ProfileConfig) UnmarshalJSON(data []byte) error {
	// Trim whitespace to check first character
	for _, b := range data {
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			continue
		}
		if b == '[' {
			// Old format: plain string array
			var providers []string
			if err := json.Unmarshal(data, &providers); err != nil {
				return err
			}
			pc.Providers = providers
			pc.Routing = nil
			return nil
		}
		break
	}

	// New format: object with providers and optional routing
	type profileConfigAlias ProfileConfig
	var alias profileConfigAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*pc = ProfileConfig(alias)
	return nil
}

// Config version history:
// - Version 1 (implicit, no version field): profiles as string arrays
// - Version 2 (v1.3.2+): profiles as objects with routing support
// - Version 3 (v1.4.0+): project bindings support
// - Version 4 (v1.5.0+): default profile and web port settings
// - Version 5 (v1.5.0+): project bindings with CLI support
// - Version 6 (v2.0.0+): renamed config dir from .opencc to .zen
// - Version 7 (v2.1.0+): renamed default_cli→default_client, cli→client in JSON; added proxy_port, web_password_hash
// - Version 8 (v2.2.0+): added pricing, budgets, webhooks, health_check; profile strategy
const CurrentConfigVersion = 8

// --- Model Pricing ---

// ModelPricing defines the cost per million tokens for a model.
type ModelPricing struct {
	InputPerMillion  float64 `json:"input_per_million"`
	OutputPerMillion float64 `json:"output_per_million"`
}

// DefaultModelPricing provides built-in pricing for common Claude models.
var DefaultModelPricing = map[string]*ModelPricing{
	"claude-opus-4-20250514":     {InputPerMillion: 15.0, OutputPerMillion: 75.0},
	"claude-sonnet-4-20250514":   {InputPerMillion: 3.0, OutputPerMillion: 15.0},
	"claude-haiku-3-5-20241022":  {InputPerMillion: 0.80, OutputPerMillion: 4.0},
	"claude-3-5-sonnet-20241022": {InputPerMillion: 3.0, OutputPerMillion: 15.0},
	"claude-3-5-haiku-20241022":  {InputPerMillion: 0.80, OutputPerMillion: 4.0},
	"claude-3-opus-20240229":     {InputPerMillion: 15.0, OutputPerMillion: 75.0},
	"claude-3-sonnet-20240229":   {InputPerMillion: 3.0, OutputPerMillion: 15.0},
	"claude-3-haiku-20240307":    {InputPerMillion: 0.25, OutputPerMillion: 1.25},
}

// --- Budget Configuration ---

// BudgetAction defines what happens when a budget limit is reached.
type BudgetAction string

const (
	BudgetActionWarn      BudgetAction = "warn"
	BudgetActionDowngrade BudgetAction = "downgrade"
	BudgetActionBlock     BudgetAction = "block"
)

// BudgetLimit defines a spending limit and the action to take when exceeded.
type BudgetLimit struct {
	Amount float64      `json:"amount"`
	Action BudgetAction `json:"action,omitempty"`
}

// BudgetConfig holds budget limits for different time periods.
type BudgetConfig struct {
	Daily      *BudgetLimit `json:"daily,omitempty"`
	Weekly     *BudgetLimit `json:"weekly,omitempty"`
	Monthly    *BudgetLimit `json:"monthly,omitempty"`
	PerProject bool         `json:"per_project,omitempty"`
}

// --- Webhook Configuration ---

// WebhookEvent defines the types of events that can trigger webhooks.
type WebhookEvent string

const (
	WebhookEventBudgetWarning  WebhookEvent = "budget_warning"
	WebhookEventBudgetExceeded WebhookEvent = "budget_exceeded"
	WebhookEventProviderDown   WebhookEvent = "provider_down"
	WebhookEventProviderUp     WebhookEvent = "provider_up"
	WebhookEventFailover       WebhookEvent = "failover"
	WebhookEventDailySummary   WebhookEvent = "daily_summary"
)

// WebhookConfig defines a webhook endpoint configuration.
type WebhookConfig struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Events  []WebhookEvent    `json:"events"`
	Headers map[string]string `json:"headers,omitempty"`
	Enabled bool              `json:"enabled"`
}

// --- Health Check Configuration ---

// HealthCheckConfig defines settings for provider health monitoring.
type HealthCheckConfig struct {
	Enabled      bool `json:"enabled"`
	IntervalSecs int  `json:"interval_secs,omitempty"`
	TimeoutSecs  int  `json:"timeout_secs,omitempty"`
}

// --- Load Balance Strategy ---

// LoadBalanceStrategy defines how providers are selected for requests.
type LoadBalanceStrategy string

const (
	LoadBalanceFailover     LoadBalanceStrategy = "failover"
	LoadBalanceRoundRobin   LoadBalanceStrategy = "round-robin"
	LoadBalanceLeastLatency LoadBalanceStrategy = "least-latency"
	LoadBalanceLeastCost    LoadBalanceStrategy = "least-cost"
)

// ProjectBinding holds the configuration for a project directory.
type ProjectBinding struct {
	Profile string `json:"profile,omitempty"` // profile name (empty = use default)
	Client  string `json:"client,omitempty"`  // client name (empty = use default)
}

// SyncConfig holds configuration for remote config sync.
type SyncConfig struct {
	Backend      string `json:"backend"`                   // "webdav"|"s3"|"gist"|"repo"
	Endpoint     string `json:"endpoint,omitempty"`        // WebDAV URL or S3 endpoint
	Bucket       string `json:"bucket,omitempty"`          // S3
	Region       string `json:"region,omitempty"`          // S3
	AccessKey    string `json:"access_key,omitempty"`      // S3
	SecretKey    string `json:"secret_key,omitempty"`      // S3
	GistID       string `json:"gist_id,omitempty"`         // Gist
	RepoOwner    string `json:"repo_owner,omitempty"`      // Repo
	RepoName     string `json:"repo_name,omitempty"`       // Repo
	RepoPath     string `json:"repo_path,omitempty"`       // Repo (default: "zen-sync.json")
	RepoBranch   string `json:"repo_branch,omitempty"`     // Repo (default: "main")
	Token        string `json:"token,omitempty"`           // PAT or WebDAV password
	Username     string `json:"username,omitempty"`        // WebDAV
	Passphrase   string `json:"passphrase,omitempty"`      // encryption passphrase (local only)
	AutoPull     bool   `json:"auto_pull,omitempty"`       // enable periodic pull
	PullInterval int    `json:"pull_interval,omitempty"`   // seconds (default: 300)
}

// OpenCCConfig is the top-level configuration structure stored in opencc.json.
type OpenCCConfig struct {
	Version         int                         `json:"version,omitempty"`           // config file version
	DefaultProfile  string                      `json:"default_profile,omitempty"`   // default profile name (defaults to "default")
	DefaultClient   string                      `json:"default_client,omitempty"`    // default client (claude, codex, opencode)
	ProxyPort       int                         `json:"proxy_port,omitempty"`        // proxy port (defaults to 19841)
	WebPort         int                         `json:"web_port,omitempty"`          // web UI port (defaults to 19840)
	WebPasswordHash string                      `json:"web_password_hash,omitempty"` // bcrypt hash for Web UI access password
	Providers       map[string]*ProviderConfig  `json:"providers"`                   // provider configurations
	Profiles        map[string]*ProfileConfig   `json:"profiles"`                    // profile configurations
	ProjectBindings map[string]*ProjectBinding  `json:"project_bindings,omitempty"`  // directory path -> binding config
	Sync            *SyncConfig                 `json:"sync,omitempty"`              // remote sync configuration
	Pricing         map[string]*ModelPricing    `json:"pricing,omitempty"`           // custom model pricing overrides
	Budgets         *BudgetConfig               `json:"budgets,omitempty"`           // budget configuration
	Webhooks        []*WebhookConfig            `json:"webhooks,omitempty"`          // webhook configurations
	HealthCheck     *HealthCheckConfig          `json:"health_check,omitempty"`      // health check configuration
}

// UnmarshalJSON supports multiple config versions:
// - v7+: current format with "default_client" and "client" keys
// - v5-v6: "default_cli" and "cli" keys (auto-migrated)
// - v3: project_bindings as map[string]string (profile name only)
func (c *OpenCCConfig) UnmarshalJSON(data []byte) error {
	// Use a raw struct that reads both old and new field names.
	// This handles all versions in a single pass.
	var raw struct {
		Version         int                            `json:"version,omitempty"`
		DefaultProfile  string                         `json:"default_profile,omitempty"`
		DefaultClient   string                         `json:"default_client,omitempty"` // v7+
		DefaultCLI      string                         `json:"default_cli,omitempty"`    // v6 compat
		ProxyPort       int                            `json:"proxy_port,omitempty"`
		WebPort         int                            `json:"web_port,omitempty"`
		WebPasswordHash string                         `json:"web_password_hash,omitempty"` // v7+
		Providers       map[string]*ProviderConfig     `json:"providers"`
		Profiles        map[string]*ProfileConfig      `json:"profiles"`
		ProjectBindings map[string]json.RawMessage     `json:"project_bindings,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Version = raw.Version
	c.DefaultProfile = raw.DefaultProfile
	c.ProxyPort = raw.ProxyPort
	c.WebPort = raw.WebPort
	c.WebPasswordHash = raw.WebPasswordHash
	c.Providers = raw.Providers
	c.Profiles = raw.Profiles

	// Migrate default_cli → default_client
	c.DefaultClient = raw.DefaultClient
	if c.DefaultClient == "" && raw.DefaultCLI != "" {
		c.DefaultClient = raw.DefaultCLI
	}

	// Parse project bindings (handles v3 string format, v5-v6 "cli" key, v7+ "client" key)
	if len(raw.ProjectBindings) > 0 {
		c.ProjectBindings = make(map[string]*ProjectBinding, len(raw.ProjectBindings))
		for path, msg := range raw.ProjectBindings {
			// Try as object first (v5+ format with "cli"/"client" keys, or empty object)
			var pbRaw struct {
				Profile string `json:"profile,omitempty"`
				Client  string `json:"client,omitempty"` // v7+
				CLI     string `json:"cli,omitempty"`    // v5-v6 compat
			}
			// Check if it's a JSON object (starts with '{')
			trimmed := bytes.TrimSpace(msg)
			if len(trimmed) > 0 && trimmed[0] == '{' {
				if err := json.Unmarshal(msg, &pbRaw); err == nil {
					client := pbRaw.Client
					if client == "" {
						client = pbRaw.CLI
					}
					c.ProjectBindings[path] = &ProjectBinding{
						Profile: pbRaw.Profile,
						Client:  client,
					}
					continue
				}
			}
			// Try as plain string (v3 format: profile name only)
			var profileName string
			if err := json.Unmarshal(msg, &profileName); err == nil {
				c.ProjectBindings[path] = &ProjectBinding{Profile: profileName}
				continue
			}
			// Skip unrecognized entries
		}
	}

	return nil
}

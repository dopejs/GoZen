package config

import (
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

	DefaultWebPort = 19840
	WebPidFile     = "web.pid"
	WebLogFile     = "web.log"

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
const CurrentConfigVersion = 6

// ProjectBinding holds the configuration for a project directory.
type ProjectBinding struct {
	Profile string `json:"profile,omitempty"` // profile name (empty = use default)
	Client  string `json:"cli,omitempty"`     // client name (empty = use default); JSON key kept as "cli" for compat
}

// OpenCCConfig is the top-level configuration structure stored in opencc.json.
type OpenCCConfig struct {
	Version         int                         `json:"version,omitempty"`          // config file version
	DefaultProfile  string                      `json:"default_profile,omitempty"`  // default profile name (defaults to "default")
	DefaultClient   string                      `json:"default_cli,omitempty"`      // default client (claude, codex, opencode); JSON key kept as "default_cli" for compat
	WebPort         int                         `json:"web_port,omitempty"`         // web UI port (defaults to 19841)
	Providers       map[string]*ProviderConfig  `json:"providers"`                  // provider configurations
	Profiles        map[string]*ProfileConfig   `json:"profiles"`                   // profile configurations
	ProjectBindings map[string]*ProjectBinding  `json:"project_bindings,omitempty"` // directory path -> binding config
}

// UnmarshalJSON supports both current format (project_bindings as map[string]*ProjectBinding)
// and the v3 format (project_bindings as map[string]string where the value is just a profile name).
func (c *OpenCCConfig) UnmarshalJSON(data []byte) error {
	// Try standard unmarshal first (works for v5+ configs and configs without project_bindings)
	type openCCConfigAlias OpenCCConfig
	var alias openCCConfigAlias
	if err := json.Unmarshal(data, &alias); err == nil {
		*c = OpenCCConfig(alias)
		return nil
	}

	// Standard unmarshal failed — likely v3 project_bindings with string values.
	// Parse with raw messages for project_bindings.
	var raw struct {
		Version         int                            `json:"version,omitempty"`
		DefaultProfile  string                         `json:"default_profile,omitempty"`
		DefaultClient   string                         `json:"default_cli,omitempty"`
		WebPort         int                            `json:"web_port,omitempty"`
		Providers       map[string]*ProviderConfig     `json:"providers"`
		Profiles        map[string]*ProfileConfig      `json:"profiles"`
		ProjectBindings map[string]json.RawMessage     `json:"project_bindings,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Version = raw.Version
	c.DefaultProfile = raw.DefaultProfile
	c.DefaultClient = raw.DefaultClient
	c.WebPort = raw.WebPort
	c.Providers = raw.Providers
	c.Profiles = raw.Profiles

	if len(raw.ProjectBindings) > 0 {
		c.ProjectBindings = make(map[string]*ProjectBinding, len(raw.ProjectBindings))
		for path, msg := range raw.ProjectBindings {
			// Try as *ProjectBinding first (v5 format)
			var pb ProjectBinding
			if err := json.Unmarshal(msg, &pb); err == nil {
				c.ProjectBindings[path] = &pb
				continue
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

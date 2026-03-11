package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"
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
	Type            string            `json:"type,omitempty"` // "anthropic" (default) or "openai"
	BaseURL         string            `json:"base_url"`
	AuthToken       string            `json:"auth_token"`
	ProxyURL        string            `json:"proxy_url,omitempty"`
	Model           string            `json:"model,omitempty"`
	ReasoningModel  string            `json:"reasoning_model,omitempty"`
	HaikuModel      string            `json:"haiku_model,omitempty"`
	OpusModel       string            `json:"opus_model,omitempty"`
	SonnetModel     string            `json:"sonnet_model,omitempty"`
	Weight          int               `json:"weight,omitempty"`            // Weight for weighted load balancing (0 = equal weight)
	EnvVars         map[string]string `json:"env_vars,omitempty"`          // Claude Code env vars (legacy, for backward compat)
	ClaudeEnvVars   map[string]string `json:"claude_env_vars,omitempty"`   // Claude Code specific env vars
	CodexEnvVars    map[string]string `json:"codex_env_vars,omitempty"`    // Codex specific env vars
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

// Clone returns a deep copy of the ProviderConfig.
func (p *ProviderConfig) Clone() *ProviderConfig {
	if p == nil {
		return nil
	}
	clone := &ProviderConfig{
		Type:           p.Type,
		BaseURL:        p.BaseURL,
		AuthToken:      p.AuthToken,
		ProxyURL:       p.ProxyURL,
		Model:          p.Model,
		ReasoningModel: p.ReasoningModel,
		HaikuModel:     p.HaikuModel,
		OpusModel:      p.OpusModel,
		SonnetModel:    p.SonnetModel,
		Weight:         p.Weight,
	}
	if p.EnvVars != nil {
		clone.EnvVars = make(map[string]string, len(p.EnvVars))
		for k, v := range p.EnvVars {
			clone.EnvVars[k] = v
		}
	}
	if p.ClaudeEnvVars != nil {
		clone.ClaudeEnvVars = make(map[string]string, len(p.ClaudeEnvVars))
		for k, v := range p.ClaudeEnvVars {
			clone.ClaudeEnvVars[k] = v
		}
	}
	if p.CodexEnvVars != nil {
		clone.CodexEnvVars = make(map[string]string, len(p.CodexEnvVars))
		for k, v := range p.CodexEnvVars {
			clone.CodexEnvVars[k] = v
		}
	}
	if p.OpenCodeEnvVars != nil {
		clone.OpenCodeEnvVars = make(map[string]string, len(p.OpenCodeEnvVars))
		for k, v := range p.OpenCodeEnvVars {
			clone.OpenCodeEnvVars[k] = v
		}
	}
	return clone
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

	// Export proxy environment variables
	p.ExportProxyToEnv()
}

// ExportProxyToEnv sets proxy environment variables based on the ProxyURL scheme.
// HTTP/HTTPS proxies set HTTP_PROXY and HTTPS_PROXY. SOCKS5 proxies set ALL_PROXY.
// No-op when ProxyURL is empty.
func (p *ProviderConfig) ExportProxyToEnv() {
	if p.ProxyURL == "" {
		return
	}
	u, err := url.Parse(p.ProxyURL)
	if err != nil {
		return
	}
	switch u.Scheme {
	case "http", "https":
		os.Setenv("HTTP_PROXY", p.ProxyURL)
		os.Setenv("HTTPS_PROXY", p.ProxyURL)
	case "socks5":
		os.Setenv("ALL_PROXY", p.ProxyURL)
	}
}

// ValidateProxyURL validates a proxy URL string.
// Empty string is valid (means no proxy). Otherwise the URL must have
// a scheme of http, https, or socks5, be parseable, and have a non-empty host.
func ValidateProxyURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("proxy_url: invalid URL format")
	}
	switch u.Scheme {
	case "http", "https", "socks5":
		// valid
	default:
		return fmt.Errorf("proxy_url: unsupported scheme %q (must be http, https, or socks5)", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("proxy_url: missing host")
	}
	return nil
}

// MaskProxyURL returns the proxy URL with credentials masked for safe logging.
// Returns the empty string unchanged.
func MaskProxyURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Redacted()
}

// Scenario represents a request scenario for routing decisions.
type Scenario string

const (
	ScenarioThink       Scenario = "think"
	ScenarioImage       Scenario = "image"
	ScenarioLongContext Scenario = "longContext"
	ScenarioWebSearch   Scenario = "webSearch"
	ScenarioBackground  Scenario = "background"
	ScenarioCode        Scenario = "code"
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
		// Also check for nil to avoid panic
		if alias.Providers[0] != nil && alias.Providers[0].Name != "" {
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
	names := make([]string, 0, len(sr.Providers))
	for _, pr := range sr.Providers {
		if pr != nil {
			names = append(names, pr.Name)
		}
	}
	return names
}

// ModelForProvider returns the model override for a specific provider, or empty string.
func (sr *ScenarioRoute) ModelForProvider(name string) string {
	for _, pr := range sr.Providers {
		if pr != nil && pr.Name == name {
			return pr.Model
		}
	}
	return ""
}

// RoutePolicy defines routing configuration for a scenario (v15+).
// Replaces ScenarioRoute with per-scenario strategy, weights, and threshold support.
type RoutePolicy struct {
	Providers            []*ProviderRoute        `json:"providers"`
	Strategy             LoadBalanceStrategy     `json:"strategy,omitempty"`              // per-scenario strategy override
	ProviderWeights      map[string]int          `json:"provider_weights,omitempty"`      // per-scenario weights
	LongContextThreshold *int                    `json:"long_context_threshold,omitempty"` // per-scenario threshold (nil = use profile default)
	FallbackToDefault    *bool                   `json:"fallback_to_default,omitempty"`   // whether to fallback to default route on failure
}

// ProviderNames returns the list of provider names in order.
func (rp *RoutePolicy) ProviderNames() []string {
	names := make([]string, 0, len(rp.Providers))
	for _, pr := range rp.Providers {
		if pr != nil {
			names = append(names, pr.Name)
		}
	}
	return names
}

// ModelForProvider returns the model override for a specific provider, or empty string.
func (rp *RoutePolicy) ModelForProvider(name string) string {
	for _, pr := range rp.Providers {
		if pr != nil && pr.Name == name {
			return pr.Model
		}
	}
	return ""
}

// ProfileConfig holds a profile's provider list and optional scenario routing.
type ProfileConfig struct {
	Providers            []string                    `json:"providers"`
	Routing              map[string]*RoutePolicy     `json:"routing,omitempty"`                // v15: string keys, RoutePolicy values
	LongContextThreshold int                         `json:"long_context_threshold,omitempty"` // defaults to 32000 if not set
	Strategy             LoadBalanceStrategy         `json:"strategy,omitempty"`               // load balancing strategy
	ProviderWeights      map[string]int              `json:"provider_weights,omitempty"`       // weights for weighted strategy
	ScenarioPriority     []string                    `json:"scenario_priority,omitempty"`      // scenario priority order for builtin classifier
}

// Clone returns a deep copy of the ProfileConfig.
func (pc *ProfileConfig) Clone() *ProfileConfig {
	if pc == nil {
		return nil
	}
	clone := &ProfileConfig{
		LongContextThreshold: pc.LongContextThreshold,
		Strategy:             pc.Strategy,
	}
	if pc.Providers != nil {
		clone.Providers = make([]string, len(pc.Providers))
		copy(clone.Providers, pc.Providers)
	}
	if pc.ScenarioPriority != nil {
		clone.ScenarioPriority = make([]string, len(pc.ScenarioPriority))
		copy(clone.ScenarioPriority, pc.ScenarioPriority)
	}
	if pc.ProviderWeights != nil {
		clone.ProviderWeights = make(map[string]int, len(pc.ProviderWeights))
		for k, v := range pc.ProviderWeights {
			clone.ProviderWeights[k] = v
		}
	}
	if pc.Routing != nil {
		clone.Routing = make(map[string]*RoutePolicy, len(pc.Routing))
		for k, v := range pc.Routing {
			if v != nil {
				routeClone := &RoutePolicy{
					Strategy:             v.Strategy,
					FallbackToDefault:    v.FallbackToDefault,
				}
				if v.LongContextThreshold != nil {
					threshold := *v.LongContextThreshold
					routeClone.LongContextThreshold = &threshold
				}
				if v.Providers != nil {
					routeClone.Providers = make([]*ProviderRoute, len(v.Providers))
					for i, pr := range v.Providers {
						if pr != nil {
							routeClone.Providers[i] = &ProviderRoute{
								Name:  pr.Name,
								Model: pr.Model,
							}
						}
					}
				}
				if v.ProviderWeights != nil {
					routeClone.ProviderWeights = make(map[string]int, len(v.ProviderWeights))
					for pk, pv := range v.ProviderWeights {
						routeClone.ProviderWeights[pk] = pv
					}
				}
				clone.Routing[k] = routeClone
			}
		}
	}
	return clone
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
// - Version 8 (v2.2.0+): added pricing, budgets, webhooks, health_check; profile strategy; compression; middleware
// - Version 9 (v2.3.0+): added agent infrastructure (coordinator, observatory, guardrails, task queue, runtime)
// - Version 10 (v3.0.0+): added skills config to bot (skill-based intent recognition)
// - Version 11 (v3.0.0-alpha.14): added show_provider_tag (deprecated in v3.0.0-alpha.14)
// - Version 12 (v3.0.0+): added auto-permission configuration (claude_auto_permission, codex_auto_permission, opencode_auto_permission)
// - Version 13 (v3.0.0+): added feature_gates for experimental features (bot, compression, middleware, agent)
// - Version 14 (v3.0.0+): added disabled_providers map for manual provider unavailability marking
const CurrentConfigVersion = 15

// FeatureGates controls experimental features.
type FeatureGates struct {
	Bot         bool `json:"bot"`         // Bot gateway (BETA)
	Compression bool `json:"compression"` // Context compression (BETA)
	Middleware  bool `json:"middleware"`  // Middleware pipeline (BETA)
	Agent       bool `json:"agent"`       // Agent infrastructure (BETA)
}

// AutoPermissionConfig holds auto-permission settings for a specific client.
type AutoPermissionConfig struct {
	Enabled bool   `json:"enabled"`           // whether auto-permission is enabled
	Mode    string `json:"mode,omitempty"`    // permission mode (client-specific: bypassPermissions, acceptEdits, never, etc.)
}

// --- Model Pricing ---

// ModelPricing defines the cost per million tokens for a model.
type ModelPricing struct {
	InputPerMillion  float64 `json:"input_per_million"`
	OutputPerMillion float64 `json:"output_per_million"`
}

// DefaultModelPricing provides built-in pricing for common Claude models.
var DefaultModelPricing = map[string]*ModelPricing{
	// Anthropic Claude models
	"claude-opus-4-20250514":     {InputPerMillion: 15.0, OutputPerMillion: 75.0},
	"claude-sonnet-4-20250514":   {InputPerMillion: 3.0, OutputPerMillion: 15.0},
	"claude-haiku-3-5-20241022":  {InputPerMillion: 0.80, OutputPerMillion: 4.0},
	"claude-3-5-sonnet-20241022": {InputPerMillion: 3.0, OutputPerMillion: 15.0},
	"claude-3-5-haiku-20241022":  {InputPerMillion: 0.80, OutputPerMillion: 4.0},
	"claude-3-opus-20240229":     {InputPerMillion: 15.0, OutputPerMillion: 75.0},
	"claude-3-sonnet-20240229":   {InputPerMillion: 3.0, OutputPerMillion: 15.0},
	"claude-3-haiku-20240307":    {InputPerMillion: 0.25, OutputPerMillion: 1.25},

	// OpenAI models
	"gpt-4o":                {InputPerMillion: 2.5, OutputPerMillion: 10.0},
	"gpt-4o-2024-11-20":     {InputPerMillion: 2.5, OutputPerMillion: 10.0},
	"gpt-4o-2024-08-06":     {InputPerMillion: 2.5, OutputPerMillion: 10.0},
	"gpt-4o-mini":           {InputPerMillion: 0.15, OutputPerMillion: 0.6},
	"gpt-4o-mini-2024-07-18": {InputPerMillion: 0.15, OutputPerMillion: 0.6},
	"gpt-4-turbo":           {InputPerMillion: 10.0, OutputPerMillion: 30.0},
	"gpt-4-turbo-2024-04-09": {InputPerMillion: 10.0, OutputPerMillion: 30.0},
	"gpt-4":                 {InputPerMillion: 30.0, OutputPerMillion: 60.0},
	"gpt-4-32k":             {InputPerMillion: 60.0, OutputPerMillion: 120.0},
	"gpt-3.5-turbo":         {InputPerMillion: 0.5, OutputPerMillion: 1.5},
	"gpt-3.5-turbo-0125":    {InputPerMillion: 0.5, OutputPerMillion: 1.5},
	"o1":                    {InputPerMillion: 15.0, OutputPerMillion: 60.0},
	"o1-2024-12-17":         {InputPerMillion: 15.0, OutputPerMillion: 60.0},
	"o1-mini":               {InputPerMillion: 3.0, OutputPerMillion: 12.0},
	"o1-mini-2024-09-12":    {InputPerMillion: 3.0, OutputPerMillion: 12.0},
	"o3-mini":               {InputPerMillion: 1.1, OutputPerMillion: 4.4},
	"o3-mini-2025-01-31":    {InputPerMillion: 1.1, OutputPerMillion: 4.4},

	// DeepSeek models
	"deepseek-chat":      {InputPerMillion: 0.14, OutputPerMillion: 0.28},
	"deepseek-coder":     {InputPerMillion: 0.14, OutputPerMillion: 0.28},
	"deepseek-reasoner":  {InputPerMillion: 0.55, OutputPerMillion: 2.19},

	// MiniMax models
	"abab6.5s-chat":   {InputPerMillion: 1.0, OutputPerMillion: 1.0},
	"abab6.5-chat":    {InputPerMillion: 3.0, OutputPerMillion: 3.0},
	"abab6.5t-chat":   {InputPerMillion: 5.0, OutputPerMillion: 5.0},
	"abab5.5-chat":    {InputPerMillion: 1.0, OutputPerMillion: 1.0},

	// GLM (Zhipu AI) models
	"glm-4-plus":       {InputPerMillion: 7.14, OutputPerMillion: 7.14},
	"glm-4-0520":       {InputPerMillion: 14.29, OutputPerMillion: 14.29},
	"glm-4":            {InputPerMillion: 14.29, OutputPerMillion: 14.29},
	"glm-4-air":        {InputPerMillion: 0.14, OutputPerMillion: 0.14},
	"glm-4-airx":       {InputPerMillion: 1.43, OutputPerMillion: 1.43},
	"glm-4-long":       {InputPerMillion: 0.14, OutputPerMillion: 0.14},
	"glm-4-flash":      {InputPerMillion: 0.014, OutputPerMillion: 0.014},
	"glm-4-flashx":     {InputPerMillion: 0.07, OutputPerMillion: 0.07},
	"codegeex-4":       {InputPerMillion: 0.07, OutputPerMillion: 0.07},

	// Google Gemini models
	"gemini-2.0-flash":       {InputPerMillion: 0.1, OutputPerMillion: 0.4},
	"gemini-2.0-flash-lite":  {InputPerMillion: 0.075, OutputPerMillion: 0.3},
	"gemini-1.5-pro":         {InputPerMillion: 1.25, OutputPerMillion: 5.0},
	"gemini-1.5-flash":       {InputPerMillion: 0.075, OutputPerMillion: 0.3},
	"gemini-1.5-flash-8b":    {InputPerMillion: 0.0375, OutputPerMillion: 0.15},

	// Mistral models
	"mistral-large-latest":  {InputPerMillion: 2.0, OutputPerMillion: 6.0},
	"mistral-small-latest":  {InputPerMillion: 0.2, OutputPerMillion: 0.6},
	"codestral-latest":      {InputPerMillion: 0.3, OutputPerMillion: 0.9},
	"ministral-8b-latest":   {InputPerMillion: 0.1, OutputPerMillion: 0.1},
	"ministral-3b-latest":   {InputPerMillion: 0.04, OutputPerMillion: 0.04},
	"pixtral-large-latest":  {InputPerMillion: 2.0, OutputPerMillion: 6.0},

	// Qwen (Alibaba) models - prices in USD converted from CNY
	"qwen-max":         {InputPerMillion: 2.8, OutputPerMillion: 11.2},
	"qwen-plus":        {InputPerMillion: 0.56, OutputPerMillion: 2.24},
	"qwen-turbo":       {InputPerMillion: 0.042, OutputPerMillion: 0.168},
	"qwen-long":        {InputPerMillion: 0.07, OutputPerMillion: 0.28},
	"qwen-coder-plus":  {InputPerMillion: 0.49, OutputPerMillion: 1.96},
	"qwen-coder-turbo": {InputPerMillion: 0.28, OutputPerMillion: 1.12},
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

// --- Context Compression Configuration (BETA) ---

// CompressionConfig holds context compression settings.
// [BETA] This feature is experimental and disabled by default.
type CompressionConfig struct {
	Enabled         bool   `json:"enabled"`          // default: false (BETA)
	ThresholdTokens int    `json:"threshold_tokens"` // trigger compression above this (default: 50000)
	TargetTokens    int    `json:"target_tokens"`    // compress to this size (default: 20000)
	SummaryModel    string `json:"summary_model"`    // model for summarization (default: "claude-3-haiku-20240307")
	PreserveRecent  int    `json:"preserve_recent"`  // keep last N messages uncompressed (default: 4)
	SummaryProvider string `json:"summary_provider"` // provider to use for summarization (default: first healthy)
}

// --- Middleware Pipeline Configuration (BETA) ---

// MiddlewareConfig holds middleware pipeline settings.
// [BETA] This feature is experimental and disabled by default.
type MiddlewareConfig struct {
	Enabled     bool               `json:"enabled"`     // default: false (BETA)
	Middlewares []*MiddlewareEntry `json:"middlewares"` // ordered list of middleware
}

// MiddlewareEntry defines a single middleware in the pipeline.
type MiddlewareEntry struct {
	Name    string          `json:"name"`             // middleware identifier
	Enabled bool            `json:"enabled"`          // can disable individual middleware
	Source  string          `json:"source,omitempty"` // "builtin", "local", "remote"
	Path    string          `json:"path,omitempty"`   // path for local plugins
	URL     string          `json:"url,omitempty"`    // URL for remote plugins
	Config  json.RawMessage `json:"config,omitempty"` // middleware-specific config
}

// --- Agent Infrastructure Configuration (BETA) ---

// AgentConfig holds agent infrastructure settings.
// [BETA] This feature is experimental and disabled by default.
type AgentConfig struct {
	Enabled     bool               `json:"enabled"`               // default: false (BETA)
	Coordinator *CoordinatorConfig `json:"coordinator,omitempty"` // file coordination
	Observatory *ObservatoryConfig `json:"observatory,omitempty"` // session monitoring
	Guardrails  *GuardrailsConfig  `json:"guardrails,omitempty"`  // safety controls
	TaskQueue   *TaskQueueConfig   `json:"task_queue,omitempty"`  // task management
	Runtime     *RuntimeConfig     `json:"runtime,omitempty"`     // autonomous runtime
}

// CoordinatorConfig holds agent coordinator settings.
type CoordinatorConfig struct {
	Enabled        bool `json:"enabled"`
	LockTimeoutSec int  `json:"lock_timeout_sec,omitempty"` // default: 300
	InjectWarnings bool `json:"inject_warnings,omitempty"`  // inject lock info into context
}

// ObservatoryConfig holds agent observatory settings.
type ObservatoryConfig struct {
	Enabled        bool `json:"enabled"`
	StuckThreshold int  `json:"stuck_threshold,omitempty"` // consecutive errors before marking stuck (default: 5)
	IdleTimeoutMin int  `json:"idle_timeout_min,omitempty"` // minutes before marking idle (default: 30)
}

// GuardrailsConfig holds agent guardrails settings.
type GuardrailsConfig struct {
	Enabled            bool    `json:"enabled"`
	SessionSpendingCap float64 `json:"session_spending_cap,omitempty"` // max $ per session
	RequestRateLimit   int     `json:"request_rate_limit,omitempty"`   // max requests per minute
	SensitiveOpsDetect bool    `json:"sensitive_ops_detect,omitempty"` // detect dangerous operations
	AutoPauseOnCap     bool    `json:"auto_pause_on_cap,omitempty"`    // pause when cap hit
}

// TaskQueueConfig holds task queue settings.
type TaskQueueConfig struct {
	Enabled    bool `json:"enabled"`
	MaxRetries int  `json:"max_retries,omitempty"` // default: 3
	Workers    int  `json:"workers,omitempty"`     // concurrent workers (default: 1)
}

// RuntimeConfig holds autonomous agent runtime settings.
type RuntimeConfig struct {
	Enabled         bool   `json:"enabled"`
	PlanningModel   string `json:"planning_model,omitempty"`   // model for planning phase
	ExecutionModel  string `json:"execution_model,omitempty"`  // model for execution phase
	ValidationModel string `json:"validation_model,omitempty"` // model for validation phase
	MaxTurns        int    `json:"max_turns,omitempty"`        // max conversation turns (default: 50)
	MaxTokens       int    `json:"max_tokens,omitempty"`       // max total tokens (default: 500000)
}

// --- Skills Configuration ---

// SkillsConfig holds skill-based intent recognition settings.
type SkillsConfig struct {
	Enabled             bool              `json:"enabled"`                        // default: true
	ConfidenceThreshold float64           `json:"confidence_threshold,omitempty"` // default: 0.7
	LLMFallback         bool              `json:"llm_fallback"`                  // default: true
	LogBufferSize       int               `json:"log_buffer_size,omitempty"`     // default: 200
	Custom              []SkillDefinition `json:"custom,omitempty"`
}

// SkillDefinition defines a user-configured custom skill.
type SkillDefinition struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Intent      string              `json:"intent"`
	Priority    int                 `json:"priority"`
	Keywords    map[string][]string `json:"keywords"`            // lang code -> keyword list
	Synonyms    map[string]string   `json:"synonyms,omitempty"`  // variant -> canonical
	Examples    []string            `json:"examples,omitempty"`
}

// DefaultSkillsConfig returns the default SkillsConfig.
func DefaultSkillsConfig() *SkillsConfig {
	return &SkillsConfig{
		Enabled:             true,
		ConfidenceThreshold: 0.7,
		LLMFallback:         true,
		LogBufferSize:       200,
	}
}

// --- Bot Configuration (BETA) ---

// BotConfig holds bot gateway settings.
// [BETA] This feature is experimental and disabled by default.
type BotConfig struct {
	Enabled     bool                    `json:"enabled"`          // default: false (BETA)
	Profile     string                  `json:"profile,omitempty"`     // profile for NLU (recommend small model)
	Model       string                  `json:"model,omitempty"`       // LLM model ID (default: "claude-3-haiku-20240307")
	SocketPath  string                  `json:"socket_path,omitempty"` // IPC socket path
	Platforms   *BotPlatformsConfig     `json:"platforms,omitempty"`
	Interaction *BotInteractionConfig   `json:"interaction,omitempty"`
	Aliases     map[string]string       `json:"aliases,omitempty"` // alias -> project path
	Notify      *BotNotifyConfig        `json:"notify,omitempty"`
	HistorySize int                     `json:"history_size,omitempty"` // conversation history size, default 20
	Skills      *SkillsConfig           `json:"skills,omitempty"`      // skill-based intent recognition
}

// BotPlatformsConfig holds configuration for all chat platforms.
type BotPlatformsConfig struct {
	Telegram    *BotTelegramConfig    `json:"telegram,omitempty"`
	Discord     *BotDiscordConfig     `json:"discord,omitempty"`
	Slack       *BotSlackConfig       `json:"slack,omitempty"`
	Lark        *BotLarkConfig        `json:"lark,omitempty"`
	FBMessenger *BotFBMessengerConfig `json:"fbmessenger,omitempty"`
}

// BotTelegramConfig holds Telegram bot settings.
type BotTelegramConfig struct {
	Enabled      bool     `json:"enabled"`
	Token        string   `json:"token"`
	AllowedUsers []string `json:"allowed_users,omitempty"`
	AllowedChats []string `json:"allowed_chats,omitempty"`
}

// BotDiscordConfig holds Discord bot settings.
type BotDiscordConfig struct {
	Enabled         bool     `json:"enabled"`
	Token           string   `json:"token"`
	AllowedUsers    []string `json:"allowed_users,omitempty"`
	AllowedChannels []string `json:"allowed_channels,omitempty"`
	AllowedGuilds   []string `json:"allowed_guilds,omitempty"`
}

// BotSlackConfig holds Slack bot settings.
type BotSlackConfig struct {
	Enabled         bool     `json:"enabled"`
	BotToken        string   `json:"bot_token"`
	AppToken        string   `json:"app_token"`
	AllowedUsers    []string `json:"allowed_users,omitempty"`
	AllowedChannels []string `json:"allowed_channels,omitempty"`
}

// BotLarkConfig holds Lark/Feishu bot settings.
type BotLarkConfig struct {
	Enabled      bool     `json:"enabled"`
	AppID        string   `json:"app_id"`
	AppSecret    string   `json:"app_secret"`
	AllowedUsers []string `json:"allowed_users,omitempty"`
	AllowedChats []string `json:"allowed_chats,omitempty"`
}

// BotFBMessengerConfig holds Facebook Messenger bot settings.
type BotFBMessengerConfig struct {
	Enabled      bool     `json:"enabled"`
	PageToken    string   `json:"page_token"`
	VerifyToken  string   `json:"verify_token"`
	AppSecret    string   `json:"app_secret,omitempty"`
	AllowedUsers []string `json:"allowed_users,omitempty"`
}

// BotInteractionConfig controls how the bot responds to messages.
type BotInteractionConfig struct {
	RequireMention  bool     `json:"require_mention"`             // default: true
	MentionKeywords []string `json:"mention_keywords,omitempty"`  // default: ["@zen", "/zen"]
	DirectMsgMode   string   `json:"direct_message_mode,omitempty"` // "always" or "mention"
	ChannelMode     string   `json:"channel_mode,omitempty"`        // "always" or "mention"
}

// BotNotifyConfig controls notification behavior.
type BotNotifyConfig struct {
	DefaultPlatform string `json:"default_platform,omitempty"` // telegram, discord, etc.
	DefaultChatID   string `json:"default_chat_id,omitempty"`
	QuietHoursStart string `json:"quiet_hours_start,omitempty"` // "23:00"
	QuietHoursEnd   string `json:"quiet_hours_end,omitempty"`   // "08:00"
	QuietHoursZone  string `json:"quiet_hours_zone,omitempty"`  // "Asia/Shanghai"
}

// --- Load Balance Strategy ---

// LoadBalanceStrategy defines how providers are selected for requests.
type LoadBalanceStrategy string

const (
	LoadBalanceFailover     LoadBalanceStrategy = "failover"
	LoadBalanceRoundRobin   LoadBalanceStrategy = "round-robin"
	LoadBalanceLeastLatency LoadBalanceStrategy = "least-latency"
	LoadBalanceLeastCost    LoadBalanceStrategy = "least-cost"
	LoadBalanceWeighted     LoadBalanceStrategy = "weighted"
)

// --- Unavailability Marking ---

// Expiration type constants for UnavailableMarking.
const (
	MarkingTypeToday     = "today"
	MarkingTypeMonth     = "month"
	MarkingTypePermanent = "permanent"
)

// UnavailableMarking represents a user's manual decision to mark a provider as unavailable.
type UnavailableMarking struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"` // zero for permanent
}

// IsExpired returns true if the marking has expired.
// Permanent markings (zero ExpiresAt) never expire.
func (m *UnavailableMarking) IsExpired() bool {
	if m == nil {
		return true
	}
	if m.ExpiresAt.IsZero() {
		return false // permanent
	}
	return time.Now().After(m.ExpiresAt)
}

// IsActive returns true if the marking exists and is not expired.
func (m *UnavailableMarking) IsActive() bool {
	return m != nil && !m.IsExpired()
}

// NewUnavailableMarking creates a new marking with the given type.
// For "today", ExpiresAt is end of current day (23:59:59 local).
// For "month", ExpiresAt is end of current month (last day 23:59:59 local).
// For "permanent", ExpiresAt is zero (never expires).
func NewUnavailableMarking(markingType string) (*UnavailableMarking, error) {
	now := time.Now()
	m := &UnavailableMarking{
		Type:      markingType,
		CreatedAt: now,
	}
	switch markingType {
	case MarkingTypeToday:
		y, mo, d := now.Date()
		m.ExpiresAt = time.Date(y, mo, d, 23, 59, 59, 0, now.Location())
	case MarkingTypeMonth:
		y, mo, _ := now.Date()
		// First day of next month minus 1 second = last second of current month
		m.ExpiresAt = time.Date(y, mo+1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second)
	case MarkingTypePermanent:
		// ExpiresAt stays zero
	default:
		return nil, fmt.Errorf("invalid marking type: %q (must be %q, %q, or %q)", markingType, MarkingTypeToday, MarkingTypeMonth, MarkingTypePermanent)
	}
	return m, nil
}

// ProjectBinding holds the configuration for a project directory.
type ProjectBinding struct {
	Profile string `json:"profile,omitempty"` // profile name (empty = use default)
	Client  string `json:"client,omitempty"`  // client name (empty = use default)
}

// SyncConfig holds configuration for remote config sync.
type SyncConfig struct {
	Enabled      bool   `json:"enabled,omitempty"`         // enable/disable sync
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
	Version                int                         `json:"version,omitempty"`                  // config file version
	DefaultProfile         string                      `json:"default_profile,omitempty"`          // default profile name (defaults to "default")
	DefaultClient          string                      `json:"default_client,omitempty"`           // default client (claude, codex, opencode)
	ProxyPort              int                         `json:"proxy_port,omitempty"`               // proxy port (defaults to 19841)
	WebPort                int                         `json:"web_port,omitempty"`                 // web UI port (defaults to 19840)
	WebPasswordHash        string                      `json:"web_password_hash,omitempty"`        // bcrypt hash for Web UI access password
	ClaudeAutoPermission   *AutoPermissionConfig       `json:"claude_auto_permission,omitempty"`   // auto-permission config for Claude Code
	CodexAutoPermission    *AutoPermissionConfig       `json:"codex_auto_permission,omitempty"`    // auto-permission config for Codex
	OpenCodeAutoPermission *AutoPermissionConfig       `json:"opencode_auto_permission,omitempty"` // auto-permission config for OpenCode
	FeatureGates           *FeatureGates               `json:"feature_gates,omitempty"`            // experimental feature toggles
	Providers              map[string]*ProviderConfig  `json:"providers"`                          // provider configurations
	Profiles               map[string]*ProfileConfig   `json:"profiles"`                           // profile configurations
	ProjectBindings        map[string]*ProjectBinding  `json:"project_bindings,omitempty"`         // directory path -> binding config
	Sync                   *SyncConfig                 `json:"sync,omitempty"`                     // remote sync configuration
	Pricing                map[string]*ModelPricing    `json:"pricing,omitempty"`                  // custom model pricing overrides
	Budgets                *BudgetConfig               `json:"budgets,omitempty"`                  // budget configuration
	Webhooks               []*WebhookConfig            `json:"webhooks,omitempty"`                 // webhook configurations
	HealthCheck            *HealthCheckConfig          `json:"health_check,omitempty"`             // health check configuration
	Compression            *CompressionConfig          `json:"compression,omitempty"`              // [BETA] context compression
	Middleware             *MiddlewareConfig           `json:"middleware,omitempty"`               // [BETA] middleware pipeline
	Agent                  *AgentConfig                `json:"agent,omitempty"`                    // [BETA] agent infrastructure
	Bot                    *BotConfig                  `json:"bot,omitempty"`                      // [BETA] bot gateway configuration
	DisabledProviders      map[string]*UnavailableMarking `json:"disabled_providers,omitempty"`    // manually disabled providers
}

// UnmarshalJSON supports multiple config versions:
// - v12+: auto-permission fields (claude_auto_permission, codex_auto_permission, opencode_auto_permission)
// - v7+: current format with "default_client" and "client" keys
// - v5-v6: "default_cli" and "cli" keys (auto-migrated)
// - v3: project_bindings as map[string]string (profile name only)
func (c *OpenCCConfig) UnmarshalJSON(data []byte) error {
	// Use a raw struct that reads both old and new field names.
	// This handles all versions in a single pass.
	var raw struct {
		Version                int                            `json:"version,omitempty"`
		DefaultProfile         string                         `json:"default_profile,omitempty"`
		DefaultClient          string                         `json:"default_client,omitempty"`          // v7+
		DefaultCLI             string                         `json:"default_cli,omitempty"`             // v6 compat
		ProxyPort              int                            `json:"proxy_port,omitempty"`
		WebPort                int                            `json:"web_port,omitempty"`
		WebPasswordHash        string                         `json:"web_password_hash,omitempty"`        // v7+
		ShowProviderTag        bool                           `json:"show_provider_tag,omitempty"`        // v11+ (deprecated)
		ClaudeAutoPermission   *AutoPermissionConfig          `json:"claude_auto_permission,omitempty"`   // v12+
		CodexAutoPermission    *AutoPermissionConfig          `json:"codex_auto_permission,omitempty"`    // v12+
		OpenCodeAutoPermission *AutoPermissionConfig          `json:"opencode_auto_permission,omitempty"` // v12+
		FeatureGates           *FeatureGates                  `json:"feature_gates,omitempty"`            // v13+
		Providers              map[string]*ProviderConfig     `json:"providers"`
		Profiles               map[string]*ProfileConfig      `json:"profiles"`
		ProjectBindings        map[string]json.RawMessage     `json:"project_bindings,omitempty"`
		Sync                   *SyncConfig                    `json:"sync,omitempty"`
		Pricing                map[string]*ModelPricing       `json:"pricing,omitempty"`
		Budgets                *BudgetConfig                  `json:"budgets,omitempty"`
		Webhooks               []*WebhookConfig               `json:"webhooks,omitempty"`
		HealthCheck            *HealthCheckConfig             `json:"health_check,omitempty"`
		Compression            *CompressionConfig             `json:"compression,omitempty"`
		Middleware             *MiddlewareConfig              `json:"middleware,omitempty"`
		Agent                  *AgentConfig                   `json:"agent,omitempty"`
		Bot                    *BotConfig                     `json:"bot,omitempty"`
		DisabledProviders      map[string]*UnavailableMarking `json:"disabled_providers,omitempty"` // v14+
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Version = raw.Version
	c.DefaultProfile = raw.DefaultProfile
	c.ProxyPort = raw.ProxyPort
	c.WebPort = raw.WebPort
	c.WebPasswordHash = raw.WebPasswordHash
	// Note: ShowProviderTag is parsed but ignored (deprecated field)
	c.ClaudeAutoPermission = raw.ClaudeAutoPermission
	c.CodexAutoPermission = raw.CodexAutoPermission
	c.OpenCodeAutoPermission = raw.OpenCodeAutoPermission
	c.FeatureGates = raw.FeatureGates
	c.Providers = raw.Providers
	c.Profiles = raw.Profiles
	c.Sync = raw.Sync
	c.Pricing = raw.Pricing
	c.Budgets = raw.Budgets
	c.Webhooks = raw.Webhooks
	c.HealthCheck = raw.HealthCheck
	c.Compression = raw.Compression
	c.Middleware = raw.Middleware
	c.Agent = raw.Agent
	c.Bot = raw.Bot
	c.DisabledProviders = raw.DisabledProviders

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

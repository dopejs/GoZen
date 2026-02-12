package config

import (
	"encoding/json"
	"os"
)

const (
	ConfigDir  = ".opencc"
	ConfigFile = "opencc.json"
	LegacyDir  = ".cc_envs"

	WebPort    = 19840
	WebPidFile = "web.pid"
	WebLogFile = "web.log"
)

// ProviderConfig holds connection and model settings for a single API provider.
type ProviderConfig struct {
	BaseURL        string `json:"base_url"`
	AuthToken      string `json:"auth_token"`
	Model          string `json:"model,omitempty"`
	ReasoningModel string `json:"reasoning_model,omitempty"`
	HaikuModel     string `json:"haiku_model,omitempty"`
	OpusModel      string `json:"opus_model,omitempty"`
	SonnetModel    string `json:"sonnet_model,omitempty"`
}

// ExportToEnv sets all ANTHROPIC_* environment variables from this provider config.
func (p *ProviderConfig) ExportToEnv() {
	os.Setenv("ANTHROPIC_BASE_URL", p.BaseURL)
	os.Setenv("ANTHROPIC_AUTH_TOKEN", p.AuthToken)
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
}

// Scenario represents a request scenario for routing decisions.
type Scenario string

const (
	ScenarioThink      Scenario = "think"
	ScenarioImage      Scenario = "image"
	ScenarioLongContext Scenario = "longContext"
	ScenarioDefault    Scenario = "default"
)

// ScenarioRoute defines providers and optional model override for a scenario.
type ScenarioRoute struct {
	Providers []string `json:"providers"`
	Model     string   `json:"model,omitempty"`
}

// ProfileConfig holds a profile's provider list and optional scenario routing.
type ProfileConfig struct {
	Providers []string                    `json:"providers"`
	Routing   map[Scenario]*ScenarioRoute `json:"routing,omitempty"`
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

// OpenCCConfig is the top-level configuration structure stored in opencc.json.
type OpenCCConfig struct {
	Providers map[string]*ProviderConfig `json:"providers"`
	Profiles  map[string]*ProfileConfig  `json:"profiles"`
}

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

// OpenCCConfig is the top-level configuration structure stored in opencc.json.
type OpenCCConfig struct {
	Providers map[string]*ProviderConfig `json:"providers"`
	Profiles  map[string][]string        `json:"profiles"`
}

// UnmarshalJSON supports both profile formats:
//   - Old/simple: {"profiles": {"default": ["p1", "p2"]}}
//   - New/object: {"profiles": {"default": {"providers": ["p1", "p2"]}}}
func (c *OpenCCConfig) UnmarshalJSON(data []byte) error {
	// Decode providers normally, profiles as raw JSON
	var raw struct {
		Providers map[string]*ProviderConfig `json:"providers"`
		Profiles  map[string]json.RawMessage `json:"profiles"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Providers = raw.Providers
	c.Profiles = make(map[string][]string, len(raw.Profiles))

	for name, rawProfile := range raw.Profiles {
		// Try simple format: ["p1", "p2"]
		var simple []string
		if err := json.Unmarshal(rawProfile, &simple); err == nil {
			c.Profiles[name] = simple
			continue
		}

		// Try object format: {"providers": ["p1", "p2"], ...}
		var obj struct {
			Providers json.RawMessage `json:"providers"`
		}
		if err := json.Unmarshal(rawProfile, &obj); err == nil && obj.Providers != nil {
			// Providers could be ["p1"] or [{"name":"p1","model":"m"}]
			var names []string
			if err := json.Unmarshal(obj.Providers, &names); err == nil {
				c.Profiles[name] = names
				continue
			}
			// Object-style providers: [{"name":"p1"}]
			var providerRoutes []struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(obj.Providers, &providerRoutes); err == nil {
				names := make([]string, 0, len(providerRoutes))
				for _, pr := range providerRoutes {
					if pr.Name != "" {
						names = append(names, pr.Name)
					}
				}
				c.Profiles[name] = names
				continue
			}
		}

		// Fallback: empty profile
		c.Profiles[name] = []string{}
	}

	return nil
}

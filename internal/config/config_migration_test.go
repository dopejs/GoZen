package config

import (
	"encoding/json"
	"testing"
)

// T078: Test v14→v15 config migration (ScenarioRoute → RoutePolicy)
func TestConfigMigrationV14ToV15_ScenarioRouteToRoutePolicy(t *testing.T) {
	// v14 config with old ScenarioRoute format (no strategy/weights/threshold)
	v14Config := `{
		"version": 14,
		"providers": {
			"provider1": {
				"base_url": "https://api.provider1.com",
				"auth_token": "token1"
			},
			"provider2": {
				"base_url": "https://api.provider2.com",
				"auth_token": "token2"
			}
		},
		"profiles": {
			"default": {
				"providers": ["provider1", "provider2"],
				"routing": {
					"think": {
						"providers": [
							{"name": "provider1", "model": "claude-opus-4"},
							{"name": "provider2"}
						]
					},
					"code": {
						"providers": [{"name": "provider2"}]
					}
				}
			}
		}
	}`

	var cfg OpenCCConfig
	if err := json.Unmarshal([]byte(v14Config), &cfg); err != nil {
		t.Fatalf("failed to unmarshal v14 config: %v", err)
	}

	// Verify version
	if cfg.Version != 14 {
		t.Errorf("version = %d, want 14", cfg.Version)
	}

	// Verify providers
	if len(cfg.Providers) != 2 {
		t.Errorf("providers count = %d, want 2", len(cfg.Providers))
	}

	// Verify profile routing was parsed as RoutePolicy
	profile := cfg.Profiles["default"]
	if profile == nil {
		t.Fatal("default profile not found")
	}

	if len(profile.Routing) != 2 {
		t.Fatalf("routing count = %d, want 2", len(profile.Routing))
	}

	// Check think route
	thinkRoute := profile.Routing["think"]
	if thinkRoute == nil {
		t.Fatal("think route not found")
	}
	if len(thinkRoute.Providers) != 2 {
		t.Errorf("think providers count = %d, want 2", len(thinkRoute.Providers))
	}
	if thinkRoute.Providers[0].Name != "provider1" {
		t.Errorf("think provider[0] name = %q, want provider1", thinkRoute.Providers[0].Name)
	}
	if thinkRoute.Providers[0].Model != "claude-opus-4" {
		t.Errorf("think provider[0] model = %q, want claude-opus-4", thinkRoute.Providers[0].Model)
	}

	// Check code route
	codeRoute := profile.Routing["code"]
	if codeRoute == nil {
		t.Fatal("code route not found")
	}
	if len(codeRoute.Providers) != 1 {
		t.Errorf("code providers count = %d, want 1", len(codeRoute.Providers))
	}

	// Verify new fields have default values (nil/empty)
	if thinkRoute.Strategy != "" {
		t.Errorf("think strategy should be empty, got %q", thinkRoute.Strategy)
	}
	if thinkRoute.ProviderWeights != nil {
		t.Errorf("think provider_weights should be nil, got %v", thinkRoute.ProviderWeights)
	}
	if thinkRoute.LongContextThreshold != nil {
		t.Errorf("think long_context_threshold should be nil, got %v", *thinkRoute.LongContextThreshold)
	}
}

// T079: Test scenario key normalization (kebab-case → camelCase)
func TestConfigMigration_ScenarioKeyNormalization(t *testing.T) {
	// Config with kebab-case and snake_case scenario keys
	configJSON := `{
		"version": 15,
		"providers": {
			"provider1": {"base_url": "https://api.test.com", "auth_token": "token"}
		},
		"profiles": {
			"default": {
				"providers": ["provider1"],
				"routing": {
					"web-search": {
						"providers": [{"name": "provider1"}]
					},
					"long_context": {
						"providers": [{"name": "provider1"}]
					},
					"customPlan": {
						"providers": [{"name": "provider1"}]
					}
				}
			}
		}
	}`

	var cfg OpenCCConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	profile := cfg.Profiles["default"]
	if profile == nil {
		t.Fatal("default profile not found")
	}

	// Verify all scenario keys are preserved as-is (normalization happens at lookup time)
	if _, ok := profile.Routing["web-search"]; !ok {
		t.Error("web-search route not found (should be preserved)")
	}
	if _, ok := profile.Routing["long_context"]; !ok {
		t.Error("long_context route not found (should be preserved)")
	}
	if _, ok := profile.Routing["customPlan"]; !ok {
		t.Error("customPlan route not found (should be preserved)")
	}
}

// T080: Test builtin scenario preservation
func TestConfigMigration_BuiltinScenarioPreservation(t *testing.T) {
	// Config with builtin scenario keys
	configJSON := `{
		"version": 15,
		"providers": {
			"provider1": {"base_url": "https://api.test.com", "auth_token": "token"}
		},
		"profiles": {
			"default": {
				"providers": ["provider1"],
				"routing": {
					"think": {"providers": [{"name": "provider1"}]},
					"image": {"providers": [{"name": "provider1"}]},
					"code": {"providers": [{"name": "provider1"}]},
					"longContext": {"providers": [{"name": "provider1"}]},
					"webSearch": {"providers": [{"name": "provider1"}]},
					"background": {"providers": [{"name": "provider1"}]},
					"default": {"providers": [{"name": "provider1"}]}
				}
			}
		}
	}`

	var cfg OpenCCConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	profile := cfg.Profiles["default"]
	if profile == nil {
		t.Fatal("default profile not found")
	}

	// Verify all builtin scenarios are preserved
	builtinScenarios := []string{"think", "image", "code", "longContext", "webSearch", "background", "default"}
	for _, scenario := range builtinScenarios {
		if _, ok := profile.Routing[scenario]; !ok {
			t.Errorf("builtin scenario %q not found", scenario)
		}
	}
}

// T081: Test config round-trip (marshal/unmarshal)
func TestConfigMigration_RoundTrip(t *testing.T) {
	original := &OpenCCConfig{
		Version: 15,
		Providers: map[string]*ProviderConfig{
			"provider1": {
				BaseURL:   "https://api.test.com",
				AuthToken: "token1",
				Type:      ProviderTypeAnthropic,
			},
		},
		Profiles: map[string]*ProfileConfig{
			"default": {
				Providers: []string{"provider1"},
				Routing: map[string]*RoutePolicy{
					"think": {
						Providers: []*ProviderRoute{
							{Name: "provider1", Model: "claude-opus-4"},
						},
						Strategy:        LoadBalanceRoundRobin,
						ProviderWeights: map[string]int{"provider1": 10},
					},
					"customPlan": {
						Providers: []*ProviderRoute{
							{Name: "provider1"},
						},
					},
				},
				Strategy:             LoadBalanceFailover,
				LongContextThreshold: 50000,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Unmarshal back
	var restored OpenCCConfig
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	// Verify version
	if restored.Version != 15 {
		t.Errorf("version = %d, want 15", restored.Version)
	}

	// Verify providers
	if len(restored.Providers) != 1 {
		t.Errorf("providers count = %d, want 1", len(restored.Providers))
	}
	if restored.Providers["provider1"].BaseURL != "https://api.test.com" {
		t.Errorf("provider1 base_url = %q", restored.Providers["provider1"].BaseURL)
	}

	// Verify profile
	profile := restored.Profiles["default"]
	if profile == nil {
		t.Fatal("default profile not found")
	}
	if len(profile.Providers) != 1 {
		t.Errorf("profile providers count = %d, want 1", len(profile.Providers))
	}
	if profile.Strategy != LoadBalanceFailover {
		t.Errorf("profile strategy = %q, want %q", profile.Strategy, LoadBalanceFailover)
	}
	if profile.LongContextThreshold != 50000 {
		t.Errorf("profile threshold = %d, want 50000", profile.LongContextThreshold)
	}

	// Verify routing
	if len(profile.Routing) != 2 {
		t.Errorf("routing count = %d, want 2", len(profile.Routing))
	}

	thinkRoute := profile.Routing["think"]
	if thinkRoute == nil {
		t.Fatal("think route not found")
	}
	if len(thinkRoute.Providers) != 1 {
		t.Errorf("think providers count = %d, want 1", len(thinkRoute.Providers))
	}
	if thinkRoute.Providers[0].Model != "claude-opus-4" {
		t.Errorf("think model = %q, want claude-opus-4", thinkRoute.Providers[0].Model)
	}
	if thinkRoute.Strategy != LoadBalanceRoundRobin {
		t.Errorf("think strategy = %q, want %q", thinkRoute.Strategy, LoadBalanceRoundRobin)
	}
	if thinkRoute.ProviderWeights["provider1"] != 10 {
		t.Errorf("think weight = %d, want 10", thinkRoute.ProviderWeights["provider1"])
	}

	customRoute := profile.Routing["customPlan"]
	if customRoute == nil {
		t.Fatal("customPlan route not found")
	}
	if len(customRoute.Providers) != 1 {
		t.Errorf("customPlan providers count = %d, want 1", len(customRoute.Providers))
	}
}

// Test profile-level strategy/weights/threshold preservation during migration
func TestConfigMigration_ProfileLevelFieldsPreserved(t *testing.T) {
	v14Config := `{
		"version": 14,
		"providers": {
			"provider1": {"base_url": "https://api.test.com", "auth_token": "token"}
		},
		"profiles": {
			"default": {
				"providers": ["provider1"],
				"strategy": "round-robin",
				"provider_weights": {"provider1": 5},
				"long_context_threshold": 40000,
				"routing": {
					"think": {
						"providers": [{"name": "provider1"}]
					}
				}
			}
		}
	}`

	var cfg OpenCCConfig
	if err := json.Unmarshal([]byte(v14Config), &cfg); err != nil {
		t.Fatalf("failed to unmarshal v14 config: %v", err)
	}

	profile := cfg.Profiles["default"]
	if profile == nil {
		t.Fatal("default profile not found")
	}

	// Verify profile-level fields are preserved
	if profile.Strategy != LoadBalanceRoundRobin {
		t.Errorf("profile strategy = %q, want %q", profile.Strategy, LoadBalanceRoundRobin)
	}
	if profile.ProviderWeights["provider1"] != 5 {
		t.Errorf("profile weight = %d, want 5", profile.ProviderWeights["provider1"])
	}
	if profile.LongContextThreshold != 40000 {
		t.Errorf("profile threshold = %d, want 40000", profile.LongContextThreshold)
	}

	// Verify routing is also preserved
	if len(profile.Routing) != 1 {
		t.Errorf("routing count = %d, want 1", len(profile.Routing))
	}
}

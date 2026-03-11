package config

import (
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *OpenCCConfig
		wantErrorCount int
		wantWarnCount  int
		errorContains  string
		warnContains   string
	}{
		{
			name: "valid config",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{
					"provider1": {BaseURL: "https://api.example.com", AuthToken: "token1"},
				},
				Profiles: map[string]*ProfileConfig{
					"default": {Providers: []string{"provider1"}},
				},
			},
			wantErrorCount: 0,
			wantWarnCount:  0,
		},
		{
			name:           "nil config",
			cfg:            nil,
			wantErrorCount: 1,
			errorContains:  "config is nil",
		},
		{
			name: "no providers",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{},
				Profiles:  map[string]*ProfileConfig{},
			},
			wantErrorCount: 0,
			wantWarnCount:  2, // no providers + no profiles
			warnContains:   "no providers configured",
		},
		{
			name: "provider missing base_url",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{
					"provider1": {AuthToken: "token1"},
				},
				Profiles: map[string]*ProfileConfig{},
			},
			wantErrorCount: 1,
			wantWarnCount:  1, // no profiles configured
			errorContains:  "base_url is required",
		},
		{
			name: "provider missing auth_token (warning only)",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{
					"provider1": {BaseURL: "https://api.example.com"},
				},
				Profiles: map[string]*ProfileConfig{},
			},
			wantErrorCount: 0,
			wantWarnCount:  2, // auth_token empty + no profiles
			warnContains:   "auth_token is empty",
		},
		{
			name: "profile references non-existent provider",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{
					"provider1": {BaseURL: "https://api.example.com", AuthToken: "token1"},
				},
				Profiles: map[string]*ProfileConfig{
					"default": {Providers: []string{"provider1", "nonexistent"}},
				},
			},
			wantErrorCount: 0,
			wantWarnCount:  1,
			warnContains:   "references non-existent provider",
		},
		{
			name: "default profile does not exist",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{
					"provider1": {BaseURL: "https://api.example.com", AuthToken: "token1"},
				},
				Profiles: map[string]*ProfileConfig{
					"work": {Providers: []string{"provider1"}},
				},
				DefaultProfile: "nonexistent",
			},
			wantErrorCount: 0,
			wantWarnCount:  1,
			warnContains:   "default profile",
		},
		{
			name: "project binding references non-existent profile",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{
					"provider1": {BaseURL: "https://api.example.com", AuthToken: "token1"},
				},
				Profiles: map[string]*ProfileConfig{
					"default": {Providers: []string{"provider1"}},
				},
				ProjectBindings: map[string]*ProjectBinding{
					"/path/to/project": {Profile: "nonexistent"},
				},
			},
			wantErrorCount: 1,
			errorContains:  "references non-existent profile",
		},
		{
			name: "project binding has invalid client",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{
					"provider1": {BaseURL: "https://api.example.com", AuthToken: "token1"},
				},
				Profiles: map[string]*ProfileConfig{
					"default": {Providers: []string{"provider1"}},
				},
				ProjectBindings: map[string]*ProjectBinding{
					"/path/to/project": {Profile: "default", Client: "invalid"},
				},
			},
			wantErrorCount: 1,
			errorContains:  "invalid client",
		},
		{
			name: "routing config with invalid provider",
			cfg: &OpenCCConfig{
				Providers: map[string]*ProviderConfig{
					"provider1": {BaseURL: "https://api.example.com", AuthToken: "token1"},
				},
				Profiles: map[string]*ProfileConfig{
					"default": {
						Providers: []string{"provider1"},
						Routing: map[string]*RoutePolicy{
							"think": {
								Providers: []*ProviderRoute{
									{Name: "nonexistent"},
								},
							},
						},
					},
				},
			},
			wantErrorCount: 1,
			errorContains:  "references non-existent provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, warnings := ValidateConfig(tt.cfg)

			if len(errors) != tt.wantErrorCount {
				t.Errorf("got %d errors, want %d. Errors: %v", len(errors), tt.wantErrorCount, errors)
			}

			if len(warnings) != tt.wantWarnCount {
				t.Errorf("got %d warnings, want %d. Warnings: %v", len(warnings), tt.wantWarnCount, warnings)
			}

			if tt.errorContains != "" && len(errors) > 0 {
				found := false
				for _, err := range errors {
					if strings.Contains(err.Error(), tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got errors: %v", tt.errorContains, errors)
				}
			}

			if tt.warnContains != "" && len(warnings) > 0 {
				found := false
				for _, warn := range warnings {
					if strings.Contains(warn, tt.warnContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning containing %q, got warnings: %v", tt.warnContains, warnings)
				}
			}
		})
	}
}

// TestValidateOnSave verifies that validation is enforced when saving config
func TestValidateOnSave(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*Store) error
		wantErr       bool
		errorContains string
	}{
		{
			name: "valid config saves successfully",
			setup: func(s *Store) error {
				s.SetProvider("provider1", &ProviderConfig{
					BaseURL:   "https://api.example.com",
					AuthToken: "token1",
				})
				return s.SetProfileConfig("default", &ProfileConfig{
					Providers: []string{"provider1"},
				})
			},
			wantErr: false,
		},
		{
			name: "profile with non-existent provider allowed (warning only)",
			setup: func(s *Store) error {
				return s.SetProfileConfig("test", &ProfileConfig{
					Providers: []string{"nonexistent"},
				})
			},
			wantErr: false,
		},
		{
			name: "routing with non-existent provider rejected",
			setup: func(s *Store) error {
				s.SetProvider("provider1", &ProviderConfig{
					BaseURL:   "https://api.example.com",
					AuthToken: "token1",
				})
				return s.SetProfileConfig("test", &ProfileConfig{
					Providers: []string{"provider1"},
					Routing: map[string]*RoutePolicy{
						"think": {
							Providers: []*ProviderRoute{
								{Name: "nonexistent"},
							},
						},
					},
				})
			},
			wantErr:       true,
			errorContains: "references non-existent provider",
		},
		{
			name: "project binding with non-existent profile rejected",
			setup: func(s *Store) error {
				s.SetProvider("provider1", &ProviderConfig{
					BaseURL:   "https://api.example.com",
					AuthToken: "token1",
				})
				s.SetProfileConfig("default", &ProfileConfig{
					Providers: []string{"provider1"},
				})
				return s.BindProject("/path/to/project", "nonexistent", "")
			},
			wantErr:       true,
			errorContains: "does not exist",
		},
		{
			name: "project binding with invalid client rejected",
			setup: func(s *Store) error {
				s.SetProvider("provider1", &ProviderConfig{
					BaseURL:   "https://api.example.com",
					AuthToken: "token1",
				})
				s.SetProfileConfig("default", &ProfileConfig{
					Providers: []string{"provider1"},
				})
				return s.BindProject("/path/to/project", "default", "invalid-client")
			},
			wantErr:       true,
			errorContains: "invalid client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", dir)
			ResetDefaultStore()
			t.Cleanup(func() { ResetDefaultStore() })

			store := DefaultStore()
			err := tt.setup(store)

			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errorContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errorContains)
				}
			}
		})
	}
}

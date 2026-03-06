package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })
	return dir
}

func TestConfigVersion(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Test 1: New config should have current version
	store := DefaultStore()
	store.SetProvider("test", &ProviderConfig{
		BaseURL:   "https://api.test.com",
		AuthToken: "test-token",
	})

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Version != CurrentConfigVersion {
		t.Errorf("new config version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}
}

// T004: Test AutoPermissionConfig type validation
func TestAutoPermissionConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  AutoPermissionConfig
		wantErr bool
	}{
		{
			name: "valid config with all fields",
			config: AutoPermissionConfig{
				Enabled: true,
				Mode:    "bypassPermissions",
			},
			wantErr: false,
		},
		{
			name: "valid config disabled",
			config: AutoPermissionConfig{
				Enabled: false,
				Mode:    "",
			},
			wantErr: false,
		},
		{
			name: "valid config with empty mode when disabled",
			config: AutoPermissionConfig{
				Enabled: false,
				Mode:    "acceptEdits",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Test JSON unmarshaling
			var decoded AutoPermissionConfig
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if decoded.Enabled != tt.config.Enabled {
				t.Errorf("Enabled = %v, want %v", decoded.Enabled, tt.config.Enabled)
			}
			if decoded.Mode != tt.config.Mode {
				t.Errorf("Mode = %q, want %q", decoded.Mode, tt.config.Mode)
			}
		})
	}
}

// T005: Test config version migration (v11→v13)
func TestConfigMigrationV11ToV12(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write v11 config (without auto-permission fields)
	v11Config := `{
  "version": 11,
  "providers": {
    "test": {
      "base_url": "https://api.test.com",
      "auth_token": "test-token"
    }
  },
  "profiles": {
    "default": {
      "providers": ["test"]
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(v11Config), 0600); err != nil {
		t.Fatal(err)
	}

	// Load config (should trigger migration)
	store := DefaultStore()
	if err := store.Load(); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify version was upgraded
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Version != 13 {
		t.Errorf("version after migration = %d, want 13", cfg.Version)
	}

	// Verify new fields exist with default values (nil/empty)
	if cfg.ClaudeAutoPermission != nil {
		t.Errorf("claude_auto_permission should be nil after migration, got %+v", cfg.ClaudeAutoPermission)
	}
	if cfg.CodexAutoPermission != nil {
		t.Errorf("codex_auto_permission should be nil after migration, got %+v", cfg.CodexAutoPermission)
	}
	if cfg.OpenCodeAutoPermission != nil {
		t.Errorf("opencode_auto_permission should be nil after migration, got %+v", cfg.OpenCodeAutoPermission)
	}
}

// T006: Test backward compatibility (old configs without new fields)
func TestBackwardCompatibility(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write old config without auto-permission fields
	oldConfig := `{
  "version": 11,
  "providers": {
    "test": {
      "base_url": "https://api.test.com",
      "auth_token": "test-token"
    }
  },
  "profiles": {
    "default": {
      "providers": ["test"]
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(oldConfig), 0600); err != nil {
		t.Fatal(err)
	}

	// Load and verify it works
	store := DefaultStore()
	if err := store.Load(); err != nil {
		t.Fatalf("failed to load old config: %v", err)
	}

	// Verify provider still accessible
	provider := store.GetProvider("test")
	if provider == nil {
		t.Fatal("provider 'test' not found after loading old config")
	}
	if provider.BaseURL != "https://api.test.com" {
		t.Errorf("provider base_url = %q, want %q", provider.BaseURL, "https://api.test.com")
	}
}

// T007: Test forward compatibility (new configs read by old code)
func TestForwardCompatibility(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write v12 config with new fields
	v12Config := `{
  "version": 12,
  "claude_auto_permission": {
    "enabled": true,
    "mode": "bypassPermissions"
  },
  "codex_auto_permission": {
    "enabled": true,
    "mode": "never"
  },
  "providers": {
    "test": {
      "base_url": "https://api.test.com",
      "auth_token": "test-token"
    }
  },
  "profiles": {
    "default": {
      "providers": ["test"]
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(v12Config), 0600); err != nil {
		t.Fatal(err)
	}

	// Load config
	store := DefaultStore()
	if err := store.Load(); err != nil {
		t.Fatalf("failed to load v12 config: %v", err)
	}

	// Verify new fields are loaded
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.ClaudeAutoPermission == nil {
		t.Fatal("claude_auto_permission should not be nil")
	}
	if !cfg.ClaudeAutoPermission.Enabled {
		t.Error("claude_auto_permission.enabled should be true")
	}
	if cfg.ClaudeAutoPermission.Mode != "bypassPermissions" {
		t.Errorf("claude_auto_permission.mode = %q, want %q", cfg.ClaudeAutoPermission.Mode, "bypassPermissions")
	}

	if cfg.CodexAutoPermission == nil {
		t.Fatal("codex_auto_permission should not be nil")
	}
	if !cfg.CodexAutoPermission.Enabled {
		t.Error("codex_auto_permission.enabled should be true")
	}
	if cfg.CodexAutoPermission.Mode != "never" {
		t.Errorf("codex_auto_permission.mode = %q, want %q", cfg.CodexAutoPermission.Mode, "never")
	}
}

// T008: Test config round-trip marshaling
func TestConfigRoundTripMarshaling(t *testing.T) {
	original := OpenCCConfig{
		Version:        12,
		DefaultProfile: "default",
		DefaultClient:  "claude",
		ClaudeAutoPermission: &AutoPermissionConfig{
			Enabled: true,
			Mode:    "bypassPermissions",
		},
		CodexAutoPermission: &AutoPermissionConfig{
			Enabled: true,
			Mode:    "never",
		},
		OpenCodeAutoPermission: &AutoPermissionConfig{
			Enabled: false,
			Mode:    "",
		},
		Providers: map[string]*ProviderConfig{
			"test": {
				BaseURL:   "https://api.test.com",
				AuthToken: "test-token",
			},
		},
		Profiles: map[string]*ProfileConfig{
			"default": {
				Providers: []string{"test"},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded OpenCCConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify all fields match
	if decoded.Version != original.Version {
		t.Errorf("Version = %d, want %d", decoded.Version, original.Version)
	}
	if decoded.DefaultProfile != original.DefaultProfile {
		t.Errorf("DefaultProfile = %q, want %q", decoded.DefaultProfile, original.DefaultProfile)
	}
	if decoded.DefaultClient != original.DefaultClient {
		t.Errorf("DefaultClient = %q, want %q", decoded.DefaultClient, original.DefaultClient)
	}

	// Verify auto-permission configs
	if decoded.ClaudeAutoPermission == nil {
		t.Fatal("ClaudeAutoPermission is nil after round-trip")
	}
	if decoded.ClaudeAutoPermission.Enabled != original.ClaudeAutoPermission.Enabled {
		t.Errorf("ClaudeAutoPermission.Enabled = %v, want %v", decoded.ClaudeAutoPermission.Enabled, original.ClaudeAutoPermission.Enabled)
	}
	if decoded.ClaudeAutoPermission.Mode != original.ClaudeAutoPermission.Mode {
		t.Errorf("ClaudeAutoPermission.Mode = %q, want %q", decoded.ClaudeAutoPermission.Mode, original.ClaudeAutoPermission.Mode)
	}

	if decoded.CodexAutoPermission == nil {
		t.Fatal("CodexAutoPermission is nil after round-trip")
	}
	if decoded.CodexAutoPermission.Enabled != original.CodexAutoPermission.Enabled {
		t.Errorf("CodexAutoPermission.Enabled = %v, want %v", decoded.CodexAutoPermission.Enabled, original.CodexAutoPermission.Enabled)
	}
	if decoded.CodexAutoPermission.Mode != original.CodexAutoPermission.Mode {
		t.Errorf("CodexAutoPermission.Mode = %q, want %q", decoded.CodexAutoPermission.Mode, original.CodexAutoPermission.Mode)
	}

	if decoded.OpenCodeAutoPermission == nil {
		t.Fatal("OpenCodeAutoPermission is nil after round-trip")
	}
	if decoded.OpenCodeAutoPermission.Enabled != original.OpenCodeAutoPermission.Enabled {
		t.Errorf("OpenCodeAutoPermission.Enabled = %v, want %v", decoded.OpenCodeAutoPermission.Enabled, original.OpenCodeAutoPermission.Enabled)
	}
}

func TestConfigVersionOldFormat(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write old format config (no version field)
	oldConfig := `{
  "providers": {
    "test": {
      "base_url": "https://api.test.com",
      "auth_token": "test-token"
    }
  },
  "profiles": {
    "default": ["test"]
  }
}`
	if err := os.WriteFile(configPath, []byte(oldConfig), 0600); err != nil {
		t.Fatal(err)
	}

	// Load should succeed and auto-upgrade version
	ResetDefaultStore()
	store := DefaultStore()

	provider := store.GetProvider("test")
	if provider == nil {
		t.Fatal("expected provider to be loaded")
	}

	// Save should write version
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Version != CurrentConfigVersion {
		t.Errorf("upgraded config version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}
}

func TestConfigVersionTooNew(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write config with future version but compatible data structure
	futureConfig := `{
  "version": 999,
  "providers": {
    "test": {
      "base_url": "https://api.test.com",
      "auth_token": "test-token"
    }
  },
  "profiles": {
    "default": ["test"]
  }
}`
	if err := os.WriteFile(configPath, []byte(futureConfig), 0600); err != nil {
		t.Fatal(err)
	}

	// Load should succeed — data structure is compatible even though version is higher
	ResetDefaultStore()
	store := &Store{path: configPath}
	err := store.Load()

	if err != nil {
		t.Fatalf("compatible future config should load without error, got: %v", err)
	}

	// Version should be preserved (not downgraded)
	if store.config.Version != 999 {
		t.Errorf("version = %d, want 999 (should be preserved)", store.config.Version)
	}

	// Data should be loaded correctly
	p := store.GetProvider("test")
	if p == nil {
		t.Fatal("provider 'test' should be loaded")
	}
	if p.BaseURL != "https://api.test.com" {
		t.Errorf("base_url = %q, want %q", p.BaseURL, "https://api.test.com")
	}
}

func TestConfigVersionTooNewIncompatible(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write config with incompatible structure (providers as array instead of object)
	badConfig := `{
  "version": 999,
  "providers": ["not", "an", "object"]
}`
	if err := os.WriteFile(configPath, []byte(badConfig), 0600); err != nil {
		t.Fatal(err)
	}

	ResetDefaultStore()
	store := &Store{path: configPath}
	err := store.Load()

	if err == nil {
		t.Fatal("incompatible future config should fail to parse")
	}
}

func TestConfigVersionTooNewPreservedOnSave(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	futureConfig := `{
  "version": 42,
  "providers": {
    "test": {
      "base_url": "https://api.test.com",
      "auth_token": "test-token"
    }
  },
  "profiles": {
    "default": { "providers": ["test"] }
  }
}`
	if err := os.WriteFile(configPath, []byte(futureConfig), 0600); err != nil {
		t.Fatal(err)
	}

	ResetDefaultStore()
	store := &Store{path: configPath}
	if err := store.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Modify something and save
	if err := store.SetProvider("new-p", &ProviderConfig{BaseURL: "https://new.com", AuthToken: "tok"}); err != nil {
		t.Fatalf("SetProvider: %v", err)
	}

	// Re-read the file and verify version was not downgraded
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]interface{}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if v := int(saved["version"].(float64)); v != 42 {
		t.Errorf("saved version = %d, want 42 (should not be downgraded)", v)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}


func TestReadWriteFallbackOrder(t *testing.T) {
	setTestHome(t)

	names := []string{"yunyi", "cctq", "minimax"}
	if err := WriteFallbackOrder(names); err != nil {
		t.Fatalf("WriteFallbackOrder() error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}

	if len(got) != len(names) {
		t.Fatalf("got %d names, want %d", len(got), len(names))
	}
	for i, n := range names {
		if got[i] != n {
			t.Errorf("got[%d] = %q, want %q", i, got[i], n)
		}
	}
}

func TestReadFallbackOrderMissing(t *testing.T) {
	setTestHome(t)
	// Don't create default profile

	_, err := ReadFallbackOrder()
	if err == nil {
		t.Error("expected error for missing default profile")
	}
}

func TestWriteFallbackOrderEmpty(t *testing.T) {
	setTestHome(t)

	if err := WriteFallbackOrder(nil); err != nil {
		t.Fatalf("WriteFallbackOrder(nil) error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 names, got %d", len(got))
	}
}

func TestWriteFallbackOrderCreatesDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })

	if err := WriteFallbackOrder([]string{"a"}); err != nil {
		t.Fatalf("WriteFallbackOrder() error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("got %v, want [a]", got)
	}
}

func TestWriteFallbackOrderErrorBadDir(t *testing.T) {
	t.Setenv("HOME", "/dev/null/impossible")
	ResetDefaultStore()
	t.Cleanup(func() { ResetDefaultStore() })

	err := WriteFallbackOrder([]string{"a"})
	if err == nil {
		t.Error("expected error when dir can't be created")
	}
}

func TestRemoveFromFallbackOrder(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("b"); err != nil {
		t.Fatalf("RemoveFromFallbackOrder() error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("got %v, want [a c]", got)
	}
}

func TestRemoveFromFallbackOrderMissingProfile(t *testing.T) {
	setTestHome(t)
	// No default profile — should be a no-op
	if err := RemoveFromFallbackOrder("x"); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRemoveFromFallbackOrderNotPresent(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b"})

	if err := RemoveFromFallbackOrder("z"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 {
		t.Errorf("expected 2 names unchanged, got %v", got)
	}
}

func TestRemoveFromFallbackOrderFirst(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("a"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Errorf("got %v, want [b c]", got)
	}
}

func TestRemoveFromFallbackOrderLast(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("c"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

func TestRemoveFromFallbackOrderOnlyItem(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"solo"})

	if err := RemoveFromFallbackOrder("solo"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestRemoveFromFallbackOrderDuplicates(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "a", "c"})

	if err := RemoveFromFallbackOrder("a"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Errorf("got %v, want [b c]", got)
	}
}

// --- Profile tests ---

func TestReadWriteProfileOrder(t *testing.T) {
	setTestHome(t)

	names := []string{"p1", "p2"}
	if err := WriteProfileOrder("work", names); err != nil {
		t.Fatalf("WriteProfileOrder() error: %v", err)
	}

	got, err := ReadProfileOrder("work")
	if err != nil {
		t.Fatalf("ReadProfileOrder() error: %v", err)
	}
	if len(got) != 2 || got[0] != "p1" || got[1] != "p2" {
		t.Errorf("got %v, want [p1 p2]", got)
	}

	// Default profile should be unaffected
	_, err = ReadProfileOrder("default")
	if err == nil {
		t.Error("expected error for missing default profile")
	}
}

func TestListProfiles(t *testing.T) {
	setTestHome(t)

	WriteProfileOrder("default", []string{"a"})
	WriteProfileOrder("work", []string{"b"})
	WriteProfileOrder("staging", []string{"c"})

	profiles := ListProfiles()
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d: %v", len(profiles), profiles)
	}
	// Should be sorted
	if profiles[0] != "default" || profiles[1] != "staging" || profiles[2] != "work" {
		t.Errorf("got %v, want [default staging work]", profiles)
	}
}

func TestListProfilesEmpty(t *testing.T) {
	setTestHome(t)
	profiles := ListProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %v", profiles)
	}
}

func TestDeleteProfile(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("work", []string{"a"})

	if err := DeleteProfile("work"); err != nil {
		t.Fatalf("DeleteProfile() error: %v", err)
	}

	_, err := ReadProfileOrder("work")
	if err == nil {
		t.Error("expected error after deleting profile")
	}
}

func TestDeleteProfileDefault(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("default", []string{"a"})

	err := DeleteProfile("default")
	if err == nil {
		t.Error("expected error when deleting default profile")
	}

	// Default should still exist
	got, err := ReadProfileOrder("default")
	if err != nil {
		t.Fatalf("default profile should still exist: %v", err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("got %v, want [a]", got)
	}
}

func TestDeleteProfileEmpty(t *testing.T) {
	setTestHome(t)
	err := DeleteProfile("")
	if err == nil {
		t.Error("expected error when deleting empty profile name")
	}
}

func TestRemoveFromProfileOrder(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("work", []string{"a", "b", "c"})

	if err := RemoveFromProfileOrder("work", "b"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadProfileOrder("work")
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("got %v, want [a c]", got)
	}
}

func TestProviderConfigExportToEnvClearsStaleVars(t *testing.T) {
	// First export with all model fields set
	p1 := &ProviderConfig{
		BaseURL:        "https://first.com",
		AuthToken:      "tok-1",
		Model:          "m1",
		ReasoningModel: "m2",
		HaikuModel:     "m3",
		OpusModel:      "m4",
		SonnetModel:    "m5",
	}
	p1.ExportToEnv()

	// Second export with no optional model fields
	p2 := &ProviderConfig{
		BaseURL:   "https://second.com",
		AuthToken: "tok-2",
	}
	p2.ExportToEnv()

	// All optional model vars should be cleared
	staleVars := []string{
		"ANTHROPIC_MODEL",
		"ANTHROPIC_REASONING_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
	}
	for _, k := range staleVars {
		if got := os.Getenv(k); got != "" {
			t.Errorf("%s should be cleared after switching provider, got %q", k, got)
		}
	}

	// Base fields should be updated
	if got := os.Getenv("ANTHROPIC_BASE_URL"); got != "https://second.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", got, "https://second.com")
	}

	// Cleanup
	os.Unsetenv("ANTHROPIC_BASE_URL")
	os.Unsetenv("ANTHROPIC_AUTH_TOKEN")
}

func TestProviderConfigExportToEnv(t *testing.T) {
	p := &ProviderConfig{
		BaseURL:        "https://test.com",
		AuthToken:      "tok-test",
		Model:          "m1",
		ReasoningModel: "m2",
		HaikuModel:     "m3",
		OpusModel:      "m4",
		SonnetModel:    "m5",
	}

	p.ExportToEnv()

	tests := map[string]string{
		"ANTHROPIC_BASE_URL":              "https://test.com",
		"ANTHROPIC_AUTH_TOKEN":            "tok-test",
		"ANTHROPIC_MODEL":                 "m1",
		"ANTHROPIC_REASONING_MODEL":       "m2",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":   "m3",
		"ANTHROPIC_DEFAULT_OPUS_MODEL":    "m4",
		"ANTHROPIC_DEFAULT_SONNET_MODEL":  "m5",
	}

	for k, want := range tests {
		if got := os.Getenv(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	// Cleanup
	for k := range tests {
		os.Unsetenv(k)
	}
}

func TestConfigDirPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	got := ConfigDirPath()
	if got != dir+"/.zen" {
		t.Errorf("ConfigDirPath() = %q", got)
	}
}

func TestConfigFilePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	got := ConfigFilePath()
	if got != dir+"/.zen/zen.json" {
		t.Errorf("ConfigFilePath() = %q", got)
	}
}

func TestLogPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	got := LogPath()
	if got != dir+"/.zen/proxy.log" {
		t.Errorf("LogPath() = %q", got)
	}
}

// --- ProfileConfig JSON tests ---

func TestProfileConfigUnmarshalOldFormat(t *testing.T) {
	// Old format: ["p1", "p2"]
	data := []byte(`["p1", "p2"]`)
	var pc ProfileConfig
	if err := json.Unmarshal(data, &pc); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if len(pc.Providers) != 2 || pc.Providers[0] != "p1" || pc.Providers[1] != "p2" {
		t.Errorf("got providers %v, want [p1 p2]", pc.Providers)
	}
	if pc.Routing != nil {
		t.Errorf("routing should be nil for old format, got %v", pc.Routing)
	}
}

func TestProfileConfigUnmarshalNewFormat(t *testing.T) {
	data := []byte(`{
		"providers": ["a", "b"],
		"routing": {
			"think": {"providers": ["b", "a"], "model": "claude-opus-4-5"},
			"image": {"providers": ["a"]}
		}
	}`)
	var pc ProfileConfig
	if err := json.Unmarshal(data, &pc); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if len(pc.Providers) != 2 || pc.Providers[0] != "a" || pc.Providers[1] != "b" {
		t.Errorf("got providers %v, want [a b]", pc.Providers)
	}
	if pc.Routing == nil {
		t.Fatal("routing should not be nil")
	}
	if len(pc.Routing) != 2 {
		t.Fatalf("expected 2 routing entries, got %d", len(pc.Routing))
	}

	thinkRoute := pc.Routing[ScenarioThink]
	if thinkRoute == nil {
		t.Fatal("think route should exist")
	}
	if len(thinkRoute.Providers) != 2 || thinkRoute.Providers[0].Name != "b" {
		t.Errorf("think providers = %v", thinkRoute.Providers)
	}
	if thinkRoute.Providers[0].Model != "claude-opus-4-5" {
		t.Errorf("think model = %q", thinkRoute.Providers[0].Model)
	}

	imageRoute := pc.Routing[ScenarioImage]
	if imageRoute == nil {
		t.Fatal("image route should exist")
	}
	if len(imageRoute.Providers) != 1 || imageRoute.Providers[0].Name != "a" {
		t.Errorf("image providers = %v", imageRoute.Providers)
	}
	if imageRoute.Providers[0].Model != "" {
		t.Errorf("image model should be empty, got %q", imageRoute.Providers[0].Model)
	}
}

func TestProfileConfigUnmarshalNewFormatNoRouting(t *testing.T) {
	data := []byte(`{"providers": ["x", "y"]}`)
	var pc ProfileConfig
	if err := json.Unmarshal(data, &pc); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if len(pc.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(pc.Providers))
	}
	if pc.Routing != nil {
		t.Errorf("routing should be nil, got %v", pc.Routing)
	}
}

func TestProfileConfigRoundTrip(t *testing.T) {
	original := ProfileConfig{
		Providers: []string{"a", "b", "c"},
		Routing: map[Scenario]*ScenarioRoute{
			ScenarioThink: {
				Providers: []*ProviderRoute{
					{Name: "c", Model: "claude-opus-4-5"},
					{Name: "a"},
				},
			},
			ScenarioLongContext: {
				Providers: []*ProviderRoute{
					{Name: "b"},
				},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var restored ProfileConfig
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(restored.Providers) != 3 {
		t.Errorf("providers count: got %d, want 3", len(restored.Providers))
	}
	for i, want := range original.Providers {
		if restored.Providers[i] != want {
			t.Errorf("providers[%d] = %q, want %q", i, restored.Providers[i], want)
		}
	}

	if len(restored.Routing) != 2 {
		t.Fatalf("routing count: got %d, want 2", len(restored.Routing))
	}

	thinkRoute := restored.Routing[ScenarioThink]
	if thinkRoute == nil {
		t.Fatal("think route should exist")
	}
	if len(thinkRoute.Providers) != 2 || thinkRoute.Providers[0].Name != "c" {
		t.Errorf("think providers = %v", thinkRoute.Providers)
	}
	if thinkRoute.Providers[0].Model != "claude-opus-4-5" {
		t.Errorf("think model = %q", thinkRoute.Providers[0].Model)
	}

	lcRoute := restored.Routing[ScenarioLongContext]
	if lcRoute == nil || len(lcRoute.Providers) != 1 || lcRoute.Providers[0].Name != "b" {
		t.Errorf("longContext route not properly round-tripped")
	}
}

func TestProfileConfigRoundTripOldFormat(t *testing.T) {
	// Start with old format, marshal, unmarshal — should produce equivalent result
	oldData := []byte(`["x", "y"]`)
	var pc ProfileConfig
	if err := json.Unmarshal(oldData, &pc); err != nil {
		t.Fatalf("Unmarshal old format error: %v", err)
	}

	newData, err := json.Marshal(pc)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var restored ProfileConfig
	if err := json.Unmarshal(newData, &restored); err != nil {
		t.Fatalf("Unmarshal new format error: %v", err)
	}

	if len(restored.Providers) != 2 || restored.Providers[0] != "x" || restored.Providers[1] != "y" {
		t.Errorf("got providers %v, want [x y]", restored.Providers)
	}
	if restored.Routing != nil {
		t.Errorf("routing should be nil after round-trip from old format")
	}
}

func TestFullConfigRoundTrip(t *testing.T) {
	setTestHome(t)

	// Write config with routing
	pc := &ProfileConfig{
		Providers: []string{"p1", "p2"},
		Routing: map[Scenario]*ScenarioRoute{
			ScenarioThink: {Providers: []*ProviderRoute{{Name: "p2", Model: "model-x"}}},
		},
	}
	if err := SetProfileConfig("myprofile", pc); err != nil {
		t.Fatalf("SetProfileConfig error: %v", err)
	}

	// Read it back
	got := GetProfileConfig("myprofile")
	if got == nil {
		t.Fatal("GetProfileConfig returned nil")
	}
	if len(got.Providers) != 2 {
		t.Errorf("providers count = %d", len(got.Providers))
	}
	if got.Routing == nil || got.Routing[ScenarioThink] == nil {
		t.Fatal("routing not preserved")
	}
	if got.Routing[ScenarioThink].Providers[0].Model != "model-x" {
		t.Errorf("model = %q", got.Routing[ScenarioThink].Providers[0].Model)
	}
}

func TestDeleteProviderCascadeRouting(t *testing.T) {
	setTestHome(t)

	// Setup: provider "a" and "b", profile with routing referencing both
	store := DefaultStore()
	store.SetProvider("a", &ProviderConfig{BaseURL: "https://a.com", AuthToken: "t"})
	store.SetProvider("b", &ProviderConfig{BaseURL: "https://b.com", AuthToken: "t"})

	pc := &ProfileConfig{
		Providers: []string{"a", "b"},
		Routing: map[Scenario]*ScenarioRoute{
			ScenarioThink: {Providers: []*ProviderRoute{{Name: "a", Model: "m1"}, {Name: "b", Model: "m1"}}},
			ScenarioImage: {Providers: []*ProviderRoute{{Name: "a"}}},
		},
	}
	SetProfileConfig("default", pc)

	// Delete provider "a"
	DeleteProviderByName("a")

	// Check routing was updated
	got := GetProfileConfig("default")
	if got == nil {
		t.Fatal("profile should still exist")
	}

	// "a" should be removed from providers
	for _, p := range got.Providers {
		if p == "a" {
			t.Error("provider 'a' should have been removed from providers")
		}
	}

	// Check routing
	if got.Routing != nil {
		if think := got.Routing[ScenarioThink]; think != nil {
			for _, p := range think.Providers {
				if p.Name == "a" {
					t.Error("provider 'a' should have been removed from think route")
				}
			}
			if len(think.Providers) != 1 || think.Providers[0].Name != "b" {
				t.Errorf("think providers = %v, want [b]", think.Providers)
			}
		}
		// image route had only "a" — should be removed entirely
		if image := got.Routing[ScenarioImage]; image != nil {
			t.Error("image route should have been removed (no providers left)")
		}
	}
}

func TestProviderConfigWithEnvVarsExportToEnv(t *testing.T) {
	p := &ProviderConfig{
		BaseURL:   "https://test.com",
		AuthToken: "tok-test",
		Model:     "claude-sonnet-4-5",
		EnvVars: map[string]string{
			"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
			"CLAUDE_CODE_EFFORT_LEVEL":      "high",
			"MY_CUSTOM_VAR":                 "custom_value",
		},
	}

	p.ExportToEnv()

	// Check base fields
	if got := os.Getenv("ANTHROPIC_BASE_URL"); got != "https://test.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q", got)
	}
	if got := os.Getenv("ANTHROPIC_MODEL"); got != "claude-sonnet-4-5" {
		t.Errorf("ANTHROPIC_MODEL = %q", got)
	}

	// Check env vars
	if got := os.Getenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS"); got != "64000" {
		t.Errorf("CLAUDE_CODE_MAX_OUTPUT_TOKENS = %q", got)
	}
	if got := os.Getenv("CLAUDE_CODE_EFFORT_LEVEL"); got != "high" {
		t.Errorf("CLAUDE_CODE_EFFORT_LEVEL = %q", got)
	}
	if got := os.Getenv("MY_CUSTOM_VAR"); got != "custom_value" {
		t.Errorf("MY_CUSTOM_VAR = %q", got)
	}

	// Cleanup
	os.Unsetenv("ANTHROPIC_BASE_URL")
	os.Unsetenv("ANTHROPIC_MODEL")
	os.Unsetenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS")
	os.Unsetenv("CLAUDE_CODE_EFFORT_LEVEL")
	os.Unsetenv("MY_CUSTOM_VAR")
}

func TestEnvVarsJSONRoundTrip(t *testing.T) {
	original := &ProviderConfig{
		BaseURL:   "https://api.test.com",
		AuthToken: "token",
		EnvVars: map[string]string{
			"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "32000",
			"CLAUDE_CODE_EFFORT_LEVEL":      "low",
			"CUSTOM_VAR":                    "value",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var restored ProviderConfig
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if restored.EnvVars == nil {
		t.Fatal("EnvVars should not be nil")
	}
	if len(restored.EnvVars) != 3 {
		t.Errorf("expected 3 env vars, got %d", len(restored.EnvVars))
	}
	if restored.EnvVars["CLAUDE_CODE_MAX_OUTPUT_TOKENS"] != "32000" {
		t.Errorf("CLAUDE_CODE_MAX_OUTPUT_TOKENS not preserved")
	}
	if restored.EnvVars["CLAUDE_CODE_EFFORT_LEVEL"] != "low" {
		t.Errorf("CLAUDE_CODE_EFFORT_LEVEL not preserved")
	}
	if restored.EnvVars["CUSTOM_VAR"] != "value" {
		t.Errorf("CUSTOM_VAR not preserved")
	}
}

func TestEnvVarsEmptyMap(t *testing.T) {
	// Test that empty map doesn't export env vars
	p := &ProviderConfig{
		BaseURL:   "https://test.com",
		AuthToken: "token",
		EnvVars:   map[string]string{},
	}

	p.ExportToEnv()

	// These should not be set
	if got := os.Getenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS"); got != "" {
		t.Errorf("CLAUDE_CODE_MAX_OUTPUT_TOKENS should not be set, got %q", got)
	}
}

func TestEnvVarsNilMap(t *testing.T) {
	// Test that nil map doesn't cause panic
	p := &ProviderConfig{
		BaseURL:   "https://test.com",
		AuthToken: "token",
		EnvVars:   nil,
	}

	p.ExportToEnv() // Should not panic
}

func TestConfigVersionV3Bindings(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write a v3-style config where project_bindings values are plain strings
	v3Config := `{
  "version": 3,
  "providers": {
    "main": {
      "base_url": "https://api.example.com",
      "auth_token": "tok-123"
    }
  },
  "profiles": {
    "default": {
      "providers": ["main"]
    }
  },
  "project_bindings": {
    "/home/user/project-a": "default",
    "/home/user/project-b": "work"
  }
}`
	if err := os.WriteFile(configPath, []byte(v3Config), 0600); err != nil {
		t.Fatal(err)
	}

	ResetDefaultStore()
	store := DefaultStore()

	// Verify provider loaded
	if p := store.GetProvider("main"); p == nil {
		t.Fatal("expected provider 'main' to be loaded")
	}

	// Verify project bindings were migrated from string to *ProjectBinding
	bindings := store.GetAllProjectBindings()
	if len(bindings) != 2 {
		t.Fatalf("expected 2 project bindings, got %d", len(bindings))
	}

	bindingA := bindings["/home/user/project-a"]
	if bindingA == nil {
		t.Fatal("expected binding for /home/user/project-a")
	}
	if bindingA.Profile != "default" {
		t.Errorf("project-a profile = %q, want %q", bindingA.Profile, "default")
	}
	if bindingA.Client != "" {
		t.Errorf("project-a cli = %q, want empty", bindingA.Client)
	}

	bindingB := bindings["/home/user/project-b"]
	if bindingB == nil {
		t.Fatal("expected binding for /home/user/project-b")
	}
	if bindingB.Profile != "work" {
		t.Errorf("project-b profile = %q, want %q", bindingB.Profile, "work")
	}

	// Verify version was upgraded after save
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Version != CurrentConfigVersion {
		t.Errorf("saved version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}
}

func TestConfigVersionV3MixedBindings(t *testing.T) {
	// Test a config with mixed binding formats (some string, some object)
	v3MixedConfig := `{
  "version": 3,
  "providers": {},
  "profiles": {},
  "project_bindings": {
    "/path/old": "my-profile",
    "/path/new": {"profile": "other", "cli": "codex"}
  }
}`
	var cfg OpenCCConfig
	if err := json.Unmarshal([]byte(v3MixedConfig), &cfg); err != nil {
		t.Fatalf("failed to unmarshal mixed bindings: %v", err)
	}

	if len(cfg.ProjectBindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(cfg.ProjectBindings))
	}

	old := cfg.ProjectBindings["/path/old"]
	if old == nil || old.Profile != "my-profile" || old.Client != "" {
		t.Errorf("old binding = %+v, want {Profile:my-profile CLI:}", old)
	}

	newB := cfg.ProjectBindings["/path/new"]
	if newB == nil || newB.Profile != "other" || newB.Client != "codex" {
		t.Errorf("new binding = %+v, want {Profile:other CLI:codex}", newB)
	}
}

func TestOpenCCConfigUnmarshalEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		wantErr         bool
		wantBindingsLen int
		checkBinding    func(t *testing.T, bindings map[string]*ProjectBinding)
	}{
		{
			name: "no project_bindings field",
			json: `{"version":5,"providers":{},"profiles":{}}`,
			wantBindingsLen: 0,
		},
		{
			name: "empty project_bindings",
			json: `{"version":5,"providers":{},"profiles":{},"project_bindings":{}}`,
			wantBindingsLen: 0,
		},
		{
			name: "v5 object bindings (normal path)",
			json: `{"version":5,"providers":{},"profiles":{},"project_bindings":{"/a":{"profile":"p","cli":"claude"}}}`,
			wantBindingsLen: 1,
			checkBinding: func(t *testing.T, b map[string]*ProjectBinding) {
				if b["/a"].Profile != "p" || b["/a"].Client != "claude" {
					t.Errorf("/a = %+v", b["/a"])
				}
			},
		},
		{
			name: "v3 all string bindings (fallback path)",
			json: `{"version":3,"providers":{},"profiles":{},"project_bindings":{"/x":"prof1","/y":"prof2"}}`,
			wantBindingsLen: 2,
			checkBinding: func(t *testing.T, b map[string]*ProjectBinding) {
				if b["/x"].Profile != "prof1" || b["/x"].Client != "" {
					t.Errorf("/x = %+v", b["/x"])
				}
				if b["/y"].Profile != "prof2" {
					t.Errorf("/y = %+v", b["/y"])
				}
			},
		},
		{
			name: "v3 empty string binding",
			json: `{"version":3,"providers":{},"profiles":{},"project_bindings":{"/z":""}}`,
			wantBindingsLen: 1,
			checkBinding: func(t *testing.T, b map[string]*ProjectBinding) {
				if b["/z"] == nil || b["/z"].Profile != "" {
					t.Errorf("/z = %+v, want empty ProjectBinding", b["/z"])
				}
			},
		},
		{
			name: "v5 binding with empty object",
			json: `{"version":5,"providers":{},"profiles":{},"project_bindings":{"/e":{}}}`,
			wantBindingsLen: 1,
			checkBinding: func(t *testing.T, b map[string]*ProjectBinding) {
				if b["/e"] == nil || b["/e"].Profile != "" || b["/e"].Client != "" {
					t.Errorf("/e = %+v, want empty", b["/e"])
				}
			},
		},
		{
			name: "invalid json",
			json: `{not valid json`,
			wantErr: true,
		},
		{
			name:            "v4 config without bindings upgrades cleanly",
			json:            `{"version":4,"providers":{"p":{"base_url":"u","auth_token":"t"}},"profiles":{"default":{"providers":["p"]}}}`,
			wantBindingsLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg OpenCCConfig
			err := json.Unmarshal([]byte(tt.json), &cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(cfg.ProjectBindings) != tt.wantBindingsLen {
				t.Fatalf("bindings len = %d, want %d", len(cfg.ProjectBindings), tt.wantBindingsLen)
			}
			if tt.checkBinding != nil {
				tt.checkBinding(t, cfg.ProjectBindings)
			}
		})
	}
}

func TestOpenCCConfigUnmarshalPreservesAllFields(t *testing.T) {
	// Ensure the fallback path doesn't lose any top-level fields
	input := `{
  "version": 3,
  "default_profile": "work",
  "default_cli": "codex",
  "web_port": 9999,
  "providers": {"p1": {"base_url": "https://a.com", "auth_token": "tok"}},
  "profiles": {"work": {"providers": ["p1"]}},
  "project_bindings": {"/proj": "work"}
}`
	var cfg OpenCCConfig
	if err := json.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if cfg.Version != 3 {
		t.Errorf("Version = %d, want 3", cfg.Version)
	}
	if cfg.DefaultProfile != "work" {
		t.Errorf("DefaultProfile = %q, want %q", cfg.DefaultProfile, "work")
	}
	if cfg.DefaultClient != "codex" {
		t.Errorf("DefaultClient = %q, want %q", cfg.DefaultClient, "codex")
	}
	if cfg.WebPort != 9999 {
		t.Errorf("WebPort = %d, want 9999", cfg.WebPort)
	}
	if cfg.Providers["p1"] == nil || cfg.Providers["p1"].BaseURL != "https://a.com" {
		t.Errorf("Provider p1 not preserved: %+v", cfg.Providers["p1"])
	}
	if cfg.Profiles["work"] == nil || len(cfg.Profiles["work"].Providers) != 1 {
		t.Errorf("Profile work not preserved: %+v", cfg.Profiles["work"])
	}
	if cfg.ProjectBindings["/proj"] == nil || cfg.ProjectBindings["/proj"].Profile != "work" {
		t.Errorf("ProjectBinding /proj not migrated: %+v", cfg.ProjectBindings["/proj"])
	}
}

func TestOpenCCConfigMarshalRoundTrip(t *testing.T) {
	// After migrating a v3 config, marshal and re-unmarshal should produce v5 format
	v3Input := `{"version":3,"providers":{},"profiles":{},"project_bindings":{"/a":"prof1"}}`
	var cfg OpenCCConfig
	if err := json.Unmarshal([]byte(v3Input), &cfg); err != nil {
		t.Fatalf("unmarshal v3: %v", err)
	}

	// Marshal (produces v5 format with object bindings)
	data, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Re-unmarshal should take the fast path (no fallback needed)
	var cfg2 OpenCCConfig
	if err := json.Unmarshal(data, &cfg2); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}

	if cfg2.ProjectBindings["/a"] == nil || cfg2.ProjectBindings["/a"].Profile != "prof1" {
		t.Errorf("round-trip failed: /a = %+v", cfg2.ProjectBindings["/a"])
	}
}

func TestMigrateFromOpenCC(t *testing.T) {
	home := setTestHome(t)

	// Create legacy ~/.opencc/opencc.json with v5 config
	legacyDir := filepath.Join(home, ".opencc")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatal(err)
	}

	v5Config := `{
  "version": 5,
  "providers": {
    "main": {
      "base_url": "https://api.example.com",
      "auth_token": "tok-123"
    }
  },
  "profiles": {
    "default": {
      "providers": ["main"]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(legacyDir, "opencc.json"), []byte(v5Config), 0600); err != nil {
		t.Fatal(err)
	}

	// Create an auxiliary file to verify it gets copied
	if err := os.WriteFile(filepath.Join(legacyDir, "proxy.log"), []byte("log data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Load should migrate from ~/.opencc/ to ~/.zen/
	store := DefaultStore()

	// Verify provider was loaded
	if p := store.GetProvider("main"); p == nil {
		t.Fatal("expected provider 'main' to be loaded after migration")
	}

	// Verify new config file exists at ~/.zen/zen.json
	newConfigPath := filepath.Join(home, ".zen", "zen.json")
	data, err := os.ReadFile(newConfigPath)
	if err != nil {
		t.Fatalf("new config file should exist: %v", err)
	}

	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Version != CurrentConfigVersion {
		t.Errorf("migrated config version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}

	// Verify auxiliary file was copied
	logData, err := os.ReadFile(filepath.Join(home, ".zen", "proxy.log"))
	if err != nil {
		t.Fatalf("proxy.log should have been copied: %v", err)
	}
	if string(logData) != "log data" {
		t.Errorf("proxy.log content = %q, want %q", string(logData), "log data")
	}
}

// --- Compat convenience function tests ---

func TestCompatGetSetProvider(t *testing.T) {
	setTestHome(t)

	if err := SetProvider("test", &ProviderConfig{
		BaseURL:   "https://test.com",
		AuthToken: "tok",
	}); err != nil {
		t.Fatal(err)
	}

	p := GetProvider("test")
	if p == nil {
		t.Fatal("expected provider")
	}
	if p.BaseURL != "https://test.com" {
		t.Errorf("BaseURL = %q", p.BaseURL)
	}

	if GetProvider("nonexistent") != nil {
		t.Error("expected nil for nonexistent provider")
	}
}

func TestCompatProviderNames(t *testing.T) {
	setTestHome(t)

	SetProvider("b", &ProviderConfig{BaseURL: "u", AuthToken: "t"})
	SetProvider("a", &ProviderConfig{BaseURL: "u", AuthToken: "t"})

	names := ProviderNames()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("ProviderNames() = %v", names)
	}
}

func TestCompatExportProviderToEnv(t *testing.T) {
	setTestHome(t)

	SetProvider("x", &ProviderConfig{BaseURL: "https://x.com", AuthToken: "tok-x"})

	if err := ExportProviderToEnv("x"); err != nil {
		t.Fatal(err)
	}
	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "https://x.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q", v)
	}
	os.Unsetenv("ANTHROPIC_BASE_URL")
	os.Unsetenv("ANTHROPIC_AUTH_TOKEN")

	if err := ExportProviderToEnv("nonexistent"); err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestCompatDefaultProfileAndCLI(t *testing.T) {
	setTestHome(t)

	// Default profile should be "default"
	if p := GetDefaultProfile(); p != "default" {
		t.Errorf("GetDefaultProfile() = %q", p)
	}

	WriteProfileOrder("work", []string{"a"})
	if err := SetDefaultProfile("work"); err != nil {
		t.Fatal(err)
	}
	if p := GetDefaultProfile(); p != "work" {
		t.Errorf("GetDefaultProfile() = %q", p)
	}

	// CLI - set and get
	if err := SetDefaultClient("codex"); err != nil {
		t.Fatal(err)
	}
	if cli := GetDefaultClient(); cli != "codex" {
		t.Errorf("GetDefaultClient() = %q", cli)
	}
}

func TestCompatWebPort(t *testing.T) {
	setTestHome(t)

	if err := SetWebPort(9090); err != nil {
		t.Fatal(err)
	}
	if p := GetWebPort(); p != 9090 {
		t.Errorf("GetWebPort() = %d", p)
	}
}

func TestCompatProxyPort(t *testing.T) {
	setTestHome(t)

	// Default
	if p := GetProxyPort(); p != DefaultProxyPort {
		t.Errorf("GetProxyPort() default = %d, want %d", p, DefaultProxyPort)
	}

	if err := SetProxyPort(29841); err != nil {
		t.Fatal(err)
	}
	if p := GetProxyPort(); p != 29841 {
		t.Errorf("GetProxyPort() = %d, want 29841", p)
	}
}

func TestProviderConfigGetType(t *testing.T) {
	// GetType returns the Type field, defaulting to "anthropic" when empty
	p1 := &ProviderConfig{Type: ""}
	if got := p1.GetType(); got != ProviderTypeAnthropic {
		t.Errorf("GetType() for empty = %q, want %q", got, ProviderTypeAnthropic)
	}

	p2 := &ProviderConfig{Type: ProviderTypeOpenAI}
	if got := p2.GetType(); got != ProviderTypeOpenAI {
		t.Errorf("GetType() for openai = %q, want %q", got, ProviderTypeOpenAI)
	}
}

func TestProviderConfigGetEnvVarsForClient(t *testing.T) {
	p := &ProviderConfig{
		EnvVars:         map[string]string{"SHARED": "1"},
		ClaudeEnvVars:   map[string]string{"CLAUDE_VAR": "c"},
		CodexEnvVars:    map[string]string{"CODEX_VAR": "x"},
		OpenCodeEnvVars: map[string]string{"OC_VAR": "o"},
	}

	// When CLI-specific vars exist, they take precedence over shared
	vars := p.GetEnvVarsForClient("claude")
	if vars["CLAUDE_VAR"] != "c" {
		t.Errorf("claude vars = %v", vars)
	}

	vars = p.GetEnvVarsForClient("codex")
	if vars["CODEX_VAR"] != "x" {
		t.Errorf("codex vars = %v", vars)
	}

	vars = p.GetEnvVarsForClient("opencode")
	if vars["OC_VAR"] != "o" {
		t.Errorf("opencode vars = %v", vars)
	}

	// When no CLI-specific vars, falls back to shared
	p2 := &ProviderConfig{
		EnvVars: map[string]string{"SHARED": "1"},
	}
	vars = p2.GetEnvVarsForClient("claude")
	if vars["SHARED"] != "1" {
		t.Errorf("fallback vars = %v", vars)
	}
}

func TestExportProxyToEnv(t *testing.T) {
	tests := []struct {
		name     string
		proxyURL string
		wantVars map[string]string // expected env vars to be set
		noVars   []string          // env vars that should NOT be set
	}{
		{
			name:     "http sets HTTP_PROXY and HTTPS_PROXY",
			proxyURL: "http://proxy:8080",
			wantVars: map[string]string{
				"HTTP_PROXY":  "http://proxy:8080",
				"HTTPS_PROXY": "http://proxy:8080",
			},
			noVars: []string{"ALL_PROXY"},
		},
		{
			name:     "https sets HTTP_PROXY and HTTPS_PROXY",
			proxyURL: "https://proxy:8443",
			wantVars: map[string]string{
				"HTTP_PROXY":  "https://proxy:8443",
				"HTTPS_PROXY": "https://proxy:8443",
			},
			noVars: []string{"ALL_PROXY"},
		},
		{
			name:     "socks5 sets ALL_PROXY",
			proxyURL: "socks5://proxy:1080",
			wantVars: map[string]string{
				"ALL_PROXY": "socks5://proxy:1080",
			},
			noVars: []string{"HTTP_PROXY", "HTTPS_PROXY"},
		},
		{
			name:     "empty does not set any proxy vars",
			proxyURL: "",
			noVars:   []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all proxy env vars before each test
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("HTTPS_PROXY")
			os.Unsetenv("ALL_PROXY")

			p := &ProviderConfig{ProxyURL: tt.proxyURL}
			p.ExportProxyToEnv()

			for k, want := range tt.wantVars {
				got := os.Getenv(k)
				if got != want {
					t.Errorf("%s = %q, want %q", k, got, want)
				}
			}
			for _, k := range tt.noVars {
				if got := os.Getenv(k); got != "" {
					t.Errorf("%s should not be set, got %q", k, got)
				}
			}
		})
	}
}

func TestValidateProxyURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{"empty string", "", false, ""},
		{"valid http", "http://proxy:8080", false, ""},
		{"valid https", "https://proxy:8443", false, ""},
		{"valid socks5", "socks5://proxy:1080", false, ""},
		{"http with IP", "http://192.168.1.1:8080", false, ""},
		{"http with credentials", "http://user:pass@proxy:8080", false, ""},
		{"unsupported scheme ftp", "ftp://proxy:21", true, "unsupported scheme"},
		{"malformed URL", "://bad", true, "invalid URL"},
		{"missing host", "http://", true, "missing host"},
		{"no scheme", "proxy:8080", true, "unsupported scheme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProxyURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProxyURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateProxyURL(%q) error = %v, want error containing %q", tt.url, err, tt.errMsg)
				}
			}
		})
	}
}

func TestConfigMigrationProxyURL(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write v8 config without proxy_url
	v8Config := `{
		"version": 8,
		"providers": {
			"test": {
				"base_url": "https://api.test.com",
				"auth_token": "tok"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(v8Config), 0644); err != nil {
		t.Fatal(err)
	}

	store := DefaultStore()
	names := store.ProviderNames()
	if len(names) == 0 {
		t.Fatal("expected at least one provider")
	}

	pc := store.GetProvider("test")
	if pc == nil {
		t.Fatal("expected provider 'test'")
	}
	if pc.ProxyURL != "" {
		t.Errorf("expected empty ProxyURL for v8 config, got %q", pc.ProxyURL)
	}

	// Verify round-trip: set proxy_url and reload
	pc.ProxyURL = "socks5://proxy:1080"
	store.SetProvider("test", pc)

	ResetDefaultStore()
	store = DefaultStore()
	pc2 := store.GetProvider("test")
	if pc2 == nil {
		t.Fatal("expected provider 'test' after reload")
	}
	if pc2.ProxyURL != "socks5://proxy:1080" {
		t.Errorf("ProxyURL round-trip failed: got %q, want %q", pc2.ProxyURL, "socks5://proxy:1080")
	}
}

func TestMaskProxyURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"empty string", "", ""},
		{"no credentials", "http://proxy:8080", "http://proxy:8080"},
		{"with credentials", "http://user:pass@proxy:8080", "http://user:xxxxx@proxy:8080"},
		{"socks5 no credentials", "socks5://proxy:1080", "socks5://proxy:1080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskProxyURL(tt.url)
			if got != tt.want {
				t.Errorf("MaskProxyURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

// --- Skills Config Migration Tests (v9 → v10) ---

func TestSkillsConfigMigrationFromV9(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// v9 config with bot but no skills field
	v9Config := `{
  "version": 9,
  "providers": {
    "test": {
      "base_url": "https://api.test.com",
      "auth_token": "test-token"
    }
  },
  "profiles": {
    "default": { "providers": ["test"] }
  },
  "bot": {
    "enabled": true,
    "profile": "default",
    "model": "claude-3-haiku-20240307"
  }
}`
	if err := os.WriteFile(configPath, []byte(v9Config), 0600); err != nil {
		t.Fatal(err)
	}

	ResetDefaultStore()
	store := DefaultStore()

	// Should auto-upgrade to v10
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Version != CurrentConfigVersion {
		t.Errorf("version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}

	// Bot config should be preserved
	bot := store.GetBot()
	if bot == nil {
		t.Fatal("bot config should be preserved after migration")
	}
	if !bot.Enabled {
		t.Error("bot.enabled should be true")
	}
	if bot.Profile != "default" {
		t.Errorf("bot.profile = %q, want %q", bot.Profile, "default")
	}
	if bot.Model != "claude-3-haiku-20240307" {
		t.Errorf("bot.model = %q, want %q", bot.Model, "claude-3-haiku-20240307")
	}

	// Provider should be preserved
	p := store.GetProvider("test")
	if p == nil {
		t.Fatal("provider should be preserved after migration")
	}
	if p.BaseURL != "https://api.test.com" {
		t.Errorf("base_url = %q, want %q", p.BaseURL, "https://api.test.com")
	}
}

func TestSkillsConfigMigrationNoBotField(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// v9 config without bot field at all
	v9Config := `{
  "version": 9,
  "providers": {
    "test": {
      "base_url": "https://api.test.com",
      "auth_token": "test-token"
    }
  },
  "profiles": {
    "default": { "providers": ["test"] }
  }
}`
	if err := os.WriteFile(configPath, []byte(v9Config), 0600); err != nil {
		t.Fatal(err)
	}

	ResetDefaultStore()
	store := DefaultStore()

	// Should upgrade without error
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Version != CurrentConfigVersion {
		t.Errorf("version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}

	// Bot should be nil (not created from nothing)
	bot := store.GetBot()
	if bot != nil {
		t.Error("bot config should be nil when not present in v9 config")
	}
}

func TestSkillsConfigRoundTrip(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write config with skills
	store := DefaultStore()
	store.SetBot(&BotConfig{
		Enabled: true,
		Profile: "default",
		Skills: &SkillsConfig{
			Enabled:             true,
			ConfidenceThreshold: 0.8,
			LLMFallback:         true,
			LogBufferSize:       500,
			Custom: []SkillDefinition{
				{
					Name:        "code-review",
					Description: "请求代码审查",
					Intent:      "send_task",
					Priority:    50,
					Keywords: map[string][]string{
						"en": {"review", "code review"},
						"zh": {"审查", "代码审查"},
					},
					Synonyms: map[string]string{"审核": "审查"},
					Examples: []string{"帮我审查一下这段代码"},
				},
			},
		},
	})

	// Re-read and verify round-trip
	ResetDefaultStore()
	store2 := DefaultStore()
	bot := store2.GetBot()
	if bot == nil {
		t.Fatal("bot config should be loaded")
	}
	if bot.Skills == nil {
		t.Fatal("skills config should be loaded")
	}
	if bot.Skills.ConfidenceThreshold != 0.8 {
		t.Errorf("confidence_threshold = %f, want 0.8", bot.Skills.ConfidenceThreshold)
	}
	if bot.Skills.LogBufferSize != 500 {
		t.Errorf("log_buffer_size = %d, want 500", bot.Skills.LogBufferSize)
	}
	if len(bot.Skills.Custom) != 1 {
		t.Fatalf("custom skills count = %d, want 1", len(bot.Skills.Custom))
	}

	skill := bot.Skills.Custom[0]
	if skill.Name != "code-review" {
		t.Errorf("skill name = %q, want %q", skill.Name, "code-review")
	}
	if skill.Intent != "send_task" {
		t.Errorf("skill intent = %q, want %q", skill.Intent, "send_task")
	}
	if skill.Priority != 50 {
		t.Errorf("skill priority = %d, want 50", skill.Priority)
	}
	if len(skill.Keywords["en"]) != 2 {
		t.Errorf("en keywords count = %d, want 2", len(skill.Keywords["en"]))
	}
	if len(skill.Keywords["zh"]) != 2 {
		t.Errorf("zh keywords count = %d, want 2", len(skill.Keywords["zh"]))
	}
	if skill.Synonyms["审核"] != "审查" {
		t.Errorf("synonym 审核 = %q, want 审查", skill.Synonyms["审核"])
	}
}

func TestSkillsConfigFieldPreservation(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// v9 config with bot containing all existing fields
	v9Config := `{
  "version": 9,
  "providers": {
    "p1": {"base_url": "https://a.com", "auth_token": "t1"}
  },
  "profiles": {
    "default": {"providers": ["p1"]}
  },
  "bot": {
    "enabled": true,
    "profile": "default",
    "model": "claude-3-haiku-20240307",
    "history_size": 30,
    "interaction": {
      "require_mention": true,
      "mention_keywords": ["@zen", "/zen"]
    },
    "aliases": {"myproj": "/home/user/myproj"},
    "notify": {
      "default_platform": "telegram",
      "default_chat_id": "12345"
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(v9Config), 0600); err != nil {
		t.Fatal(err)
	}

	ResetDefaultStore()
	store := DefaultStore()
	bot := store.GetBot()
	if bot == nil {
		t.Fatal("bot config should be loaded")
	}

	// All existing fields should be preserved
	if bot.HistorySize != 30 {
		t.Errorf("history_size = %d, want 30", bot.HistorySize)
	}
	if bot.Interaction == nil {
		t.Fatal("interaction should be preserved")
	}
	if !bot.Interaction.RequireMention {
		t.Error("require_mention should be true")
	}
	if bot.Aliases["myproj"] != "/home/user/myproj" {
		t.Errorf("alias myproj = %q, want /home/user/myproj", bot.Aliases["myproj"])
	}
	if bot.Notify == nil {
		t.Fatal("notify should be preserved")
	}
	if bot.Notify.DefaultPlatform != "telegram" {
		t.Errorf("default_platform = %q, want telegram", bot.Notify.DefaultPlatform)
	}
	if bot.Notify.DefaultChatID != "12345" {
		t.Errorf("default_chat_id = %q, want 12345", bot.Notify.DefaultChatID)
	}
}

func TestConfig_DeprecatedFieldIgnored(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Write config with deprecated field
	deprecatedConfig := `{
		"version": 11,
		"show_provider_tag": true,
		"providers": {
			"test-provider": {
				"base_url": "https://api.test.com",
				"auth_token": "test-token"
			}
		},
		"profiles": {}
	}`

	if err := os.WriteFile(configPath, []byte(deprecatedConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Load config - should not error
	store := DefaultStore()
	providers := store.ProviderNames()

	// Verify provider was loaded correctly
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}

	if p := store.GetProvider("test-provider"); p == nil {
		t.Error("test-provider not found")
	} else if p.BaseURL != "https://api.test.com" {
		t.Errorf("provider BaseURL = %s, want https://api.test.com", p.BaseURL)
	}

	// Verify deprecated field is ignored (no error, field not accessible)
	// The field should be parsed but not stored or used
}

// TestFeatureGatesJSONSerialization tests FeatureGates struct JSON serialization (table-driven).
func TestFeatureGatesJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		gates    *FeatureGates
		wantJSON string
	}{
		{
			name:     "all disabled",
			gates:    &FeatureGates{},
			wantJSON: `{"bot":false,"compression":false,"middleware":false,"agent":false}`,
		},
		{
			name:     "bot enabled",
			gates:    &FeatureGates{Bot: true},
			wantJSON: `{"bot":true,"compression":false,"middleware":false,"agent":false}`,
		},
		{
			name:     "all enabled",
			gates:    &FeatureGates{Bot: true, Compression: true, Middleware: true, Agent: true},
			wantJSON: `{"bot":true,"compression":true,"middleware":true,"agent":true}`,
		},
		{
			name:     "mixed state",
			gates:    &FeatureGates{Bot: true, Agent: true},
			wantJSON: `{"bot":true,"compression":false,"middleware":false,"agent":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Marshal
			data, err := json.Marshal(tt.gates)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("Marshal = %s, want %s", string(data), tt.wantJSON)
			}

			// Test Unmarshal
			var gates FeatureGates
			if err := json.Unmarshal([]byte(tt.wantJSON), &gates); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if gates != *tt.gates {
				t.Errorf("Unmarshal = %+v, want %+v", gates, *tt.gates)
			}
		})
	}
}

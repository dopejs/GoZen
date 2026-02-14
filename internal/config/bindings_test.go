package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectBindings(t *testing.T) {
	home := setTestHome(t)

	// Create a test profile
	err := SetProfileConfig("test-profile", &ProfileConfig{
		Providers: []string{"test-provider"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test binding with profile only
	testPath := filepath.Join(home, "test-project")
	err = BindProject(testPath, "test-profile", "")
	if err != nil {
		t.Fatalf("BindProject() error: %v", err)
	}

	// Test getting binding
	binding := GetProjectBinding(testPath)
	if binding == nil {
		t.Fatal("GetProjectBinding() returned nil")
	}
	if binding.Profile != "test-profile" {
		t.Errorf("GetProjectBinding().Profile = %q, want %q", binding.Profile, "test-profile")
	}

	// Test getting all bindings
	bindings := GetAllProjectBindings()
	if len(bindings) != 1 {
		t.Errorf("GetAllProjectBindings() len = %d, want 1", len(bindings))
	}
	if bindings[testPath].Profile != "test-profile" {
		t.Errorf("bindings[%q].Profile = %q, want %q", testPath, bindings[testPath].Profile, "test-profile")
	}

	// Test unbinding
	err = UnbindProject(testPath)
	if err != nil {
		t.Fatalf("UnbindProject() error: %v", err)
	}

	// Verify unbinding
	binding = GetProjectBinding(testPath)
	if binding != nil {
		t.Errorf("GetProjectBinding() after unbind = %v, want nil", binding)
	}
}

func TestProjectBindingsWithCLI(t *testing.T) {
	home := setTestHome(t)

	// Create a test profile
	err := SetProfileConfig("cli-profile", &ProfileConfig{
		Providers: []string{"test-provider"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test binding with both profile and CLI
	testPath := filepath.Join(home, "cli-project")
	err = BindProject(testPath, "cli-profile", "codex")
	if err != nil {
		t.Fatalf("BindProject() error: %v", err)
	}

	binding := GetProjectBinding(testPath)
	if binding == nil {
		t.Fatal("GetProjectBinding() returned nil")
	}
	if binding.Profile != "cli-profile" {
		t.Errorf("binding.Profile = %q, want %q", binding.Profile, "cli-profile")
	}
	if binding.CLI != "codex" {
		t.Errorf("binding.CLI = %q, want %q", binding.CLI, "codex")
	}

	// Test binding with CLI only (no profile)
	testPath2 := filepath.Join(home, "cli-only-project")
	err = BindProject(testPath2, "", "opencode")
	if err != nil {
		t.Fatalf("BindProject() with CLI only error: %v", err)
	}

	binding2 := GetProjectBinding(testPath2)
	if binding2 == nil {
		t.Fatal("GetProjectBinding() returned nil for CLI-only binding")
	}
	if binding2.Profile != "" {
		t.Errorf("binding.Profile = %q, want empty", binding2.Profile)
	}
	if binding2.CLI != "opencode" {
		t.Errorf("binding.CLI = %q, want %q", binding2.CLI, "opencode")
	}
}

func TestBindNonexistentProfile(t *testing.T) {
	setTestHome(t)

	testPath := "/tmp/test-project"
	err := BindProject(testPath, "nonexistent", "")
	if err == nil {
		t.Error("BindProject() with nonexistent profile should error")
	}
}

func TestBindInvalidCLI(t *testing.T) {
	setTestHome(t)

	testPath := "/tmp/test-project"
	err := BindProject(testPath, "", "invalid-cli")
	if err == nil {
		t.Error("BindProject() with invalid CLI should error")
	}
}

func TestUnbindNonexistentPath(t *testing.T) {
	setTestHome(t)

	// Unbinding a path that was never bound should not error
	err := UnbindProject("/tmp/never-bound")
	if err != nil {
		t.Errorf("UnbindProject() error: %v", err)
	}
}

func TestProjectBindingPersistence(t *testing.T) {
	home := setTestHome(t)

	// Create a test profile
	err := SetProfileConfig("persist-profile", &ProfileConfig{
		Providers: []string{"test-provider"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Bind a project with both profile and CLI
	testPath := filepath.Join(home, "persist-project")
	err = BindProject(testPath, "persist-profile", "claude")
	if err != nil {
		t.Fatal(err)
	}

	// Reset store to force reload from disk
	ResetDefaultStore()

	// Verify binding persisted
	binding := GetProjectBinding(testPath)
	if binding == nil {
		t.Fatal("GetProjectBinding() after reload returned nil")
	}
	if binding.Profile != "persist-profile" {
		t.Errorf("binding.Profile after reload = %q, want %q", binding.Profile, "persist-profile")
	}
	if binding.CLI != "claude" {
		t.Errorf("binding.CLI after reload = %q, want %q", binding.CLI, "claude")
	}
}

func TestConfigVersionWithBindings(t *testing.T) {
	home := setTestHome(t)
	configPath := filepath.Join(home, ConfigDir, ConfigFile)

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a profile and binding
	SetProfileConfig("test", &ProfileConfig{Providers: []string{"p1"}})
	BindProject("/test/path", "test", "codex")

	// Read config and check version
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Version != CurrentConfigVersion {
		t.Errorf("config version = %d, want %d", cfg.Version, CurrentConfigVersion)
	}

	if len(cfg.ProjectBindings) != 1 {
		t.Errorf("project_bindings len = %d, want 1", len(cfg.ProjectBindings))
	}

	binding := cfg.ProjectBindings["/test/path"]
	if binding == nil {
		t.Fatal("binding not found in config")
	}
	if binding.Profile != "test" {
		t.Errorf("binding.Profile = %q, want %q", binding.Profile, "test")
	}
	if binding.CLI != "codex" {
		t.Errorf("binding.CLI = %q, want %q", binding.CLI, "codex")
	}
}

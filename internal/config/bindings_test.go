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

	// Test binding
	testPath := filepath.Join(home, "test-project")
	err = BindProject(testPath, "test-profile")
	if err != nil {
		t.Fatalf("BindProject() error: %v", err)
	}

	// Test getting binding
	profile := GetProjectBinding(testPath)
	if profile != "test-profile" {
		t.Errorf("GetProjectBinding() = %q, want %q", profile, "test-profile")
	}

	// Test getting all bindings
	bindings := GetAllProjectBindings()
	if len(bindings) != 1 {
		t.Errorf("GetAllProjectBindings() len = %d, want 1", len(bindings))
	}
	if bindings[testPath] != "test-profile" {
		t.Errorf("bindings[%q] = %q, want %q", testPath, bindings[testPath], "test-profile")
	}

	// Test unbinding
	err = UnbindProject(testPath)
	if err != nil {
		t.Fatalf("UnbindProject() error: %v", err)
	}

	// Verify unbinding
	profile = GetProjectBinding(testPath)
	if profile != "" {
		t.Errorf("GetProjectBinding() after unbind = %q, want empty", profile)
	}
}

func TestBindNonexistentProfile(t *testing.T) {
	setTestHome(t)

	testPath := "/tmp/test-project"
	err := BindProject(testPath, "nonexistent")
	if err == nil {
		t.Error("BindProject() with nonexistent profile should error")
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

	// Bind a project
	testPath := filepath.Join(home, "persist-project")
	err = BindProject(testPath, "persist-profile")
	if err != nil {
		t.Fatal(err)
	}

	// Reset store to force reload from disk
	ResetDefaultStore()

	// Verify binding persisted
	profile := GetProjectBinding(testPath)
	if profile != "persist-profile" {
		t.Errorf("GetProjectBinding() after reload = %q, want %q", profile, "persist-profile")
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
	BindProject("/test/path", "test")

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
}

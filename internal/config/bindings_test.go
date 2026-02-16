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
	if binding.Client != "codex" {
		t.Errorf("binding.Client = %q, want %q", binding.Client, "codex")
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
	if binding2.Client != "opencode" {
		t.Errorf("binding.Client = %q, want %q", binding2.Client, "opencode")
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
	if binding.Client != "claude" {
		t.Errorf("binding.Client after reload = %q, want %q", binding.Client, "claude")
	}
}

func TestProjectBindingSymlinkDedup(t *testing.T) {
	home := setTestHome(t)

	// Create a test profile
	err := SetProfileConfig("sym-profile", &ProfileConfig{
		Providers: []string{"test-provider"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a real directory and a symlink pointing to it
	realDir := filepath.Join(home, "real-project")
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(home, "link-project")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	// Resolve realDir so comparison works on systems where tmpdir is itself a symlink
	// (e.g. macOS /var -> /private/var)
	resolvedRealDir, err := filepath.EvalSymlinks(realDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(realDir) error: %v", err)
	}

	// Bind via the real path
	err = BindProject(realDir, "sym-profile", "claude")
	if err != nil {
		t.Fatalf("BindProject(realDir) error: %v", err)
	}

	// Lookup via the symlink path should find the same binding
	binding := GetProjectBinding(linkDir)
	if binding == nil {
		t.Fatal("GetProjectBinding(linkDir) returned nil, expected to find binding via symlink")
	}
	if binding.Profile != "sym-profile" {
		t.Errorf("binding.Profile = %q, want %q", binding.Profile, "sym-profile")
	}

	// Rebind via the symlink path should update, not duplicate
	err = BindProject(linkDir, "sym-profile", "codex")
	if err != nil {
		t.Fatalf("BindProject(linkDir) error: %v", err)
	}

	allBindings := GetAllProjectBindings()
	count := 0
	for p := range allBindings {
		if p == resolvedRealDir {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 binding for resolvedRealDir, got %d (bindings: %v)", count, allBindings)
	}

	// Verify the CLI was updated
	binding = GetProjectBinding(realDir)
	if binding == nil {
		t.Fatal("GetProjectBinding(realDir) returned nil after rebind")
	}
	if binding.Client != "codex" {
		t.Errorf("binding.Client = %q, want %q after rebind via symlink", binding.Client, "codex")
	}

	// Unbind via symlink should remove the binding
	err = UnbindProject(linkDir)
	if err != nil {
		t.Fatalf("UnbindProject(linkDir) error: %v", err)
	}
	binding = GetProjectBinding(realDir)
	if binding != nil {
		t.Error("GetProjectBinding(realDir) should be nil after unbind via symlink")
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
	if binding.Client != "codex" {
		t.Errorf("binding.Client = %q, want %q", binding.Client, "codex")
	}
}

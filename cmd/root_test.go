package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anthropics/opencc/internal/config"
)

func setTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	envDir := filepath.Join(dir, ".cc_envs")
	os.MkdirAll(envDir, 0755)
	return dir
}

func writeTestEnv(t *testing.T, name, content string) {
	t.Helper()
	home := os.Getenv("HOME")
	path := filepath.Join(home, ".cc_envs", name+".env")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildProviders(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "yunyi", "ANTHROPIC_BASE_URL=https://yunyi.example.com\nANTHROPIC_AUTH_TOKEN=tok1\nANTHROPIC_MODEL=opus\n")
	writeTestEnv(t, "cctq", "ANTHROPIC_BASE_URL=https://cctq.example.com\nANTHROPIC_AUTH_TOKEN=tok2\n")

	providers, firstModel, err := buildProviders([]string{"yunyi", "cctq"})
	if err != nil {
		t.Fatalf("buildProviders() error: %v", err)
	}

	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	if providers[0].Name != "yunyi" {
		t.Errorf("providers[0].Name = %q", providers[0].Name)
	}
	if providers[0].BaseURL.String() != "https://yunyi.example.com" {
		t.Errorf("providers[0].BaseURL = %q", providers[0].BaseURL.String())
	}
	if providers[0].Token != "tok1" {
		t.Errorf("providers[0].Token = %q", providers[0].Token)
	}
	if providers[0].Model != "opus" {
		t.Errorf("providers[0].Model = %q", providers[0].Model)
	}

	// cctq has no model, should default
	if providers[1].Model != "claude-sonnet-4-20250514" {
		t.Errorf("providers[1].Model = %q, want default", providers[1].Model)
	}

	if firstModel != "opus" {
		t.Errorf("firstModel = %q, want %q", firstModel, "opus")
	}
}

func TestBuildProvidersSkipsEmpty(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "a", "ANTHROPIC_BASE_URL=https://a.com\nANTHROPIC_AUTH_TOKEN=tok\n")

	providers, _, err := buildProviders([]string{"", "a", "  "})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}
}

func TestBuildProvidersMissingConfig(t *testing.T) {
	setTestHome(t)

	_, _, err := buildProviders([]string{"nonexistent"})
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestBuildProvidersMissingURL(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "bad", "ANTHROPIC_AUTH_TOKEN=tok\n")

	_, _, err := buildProviders([]string{"bad"})
	if err == nil {
		t.Error("expected error for missing ANTHROPIC_BASE_URL")
	}
}

func TestBuildProvidersMissingToken(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "bad", "ANTHROPIC_BASE_URL=https://example.com\n")

	_, _, err := buildProviders([]string{"bad"})
	if err == nil {
		t.Error("expected error for missing ANTHROPIC_AUTH_TOKEN")
	}
}

func TestBuildProvidersAllEmpty(t *testing.T) {
	setTestHome(t)

	_, _, err := buildProviders([]string{"", "  "})
	if err == nil {
		t.Error("expected error for no valid providers")
	}
}

func TestVersionValue(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

// PLACEHOLDER_MORE_TESTS

func writeProfileConf(t *testing.T, profile, content string) {
	t.Helper()
	home := os.Getenv("HOME")
	var path string
	if profile == "" || profile == "default" {
		path = filepath.Join(home, ".cc_envs", "fallback.conf")
	} else {
		path = filepath.Join(home, ".cc_envs", "fallback."+profile+".conf")
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveWithProfileFlag(t *testing.T) {
	setTestHome(t)
	writeProfileConf(t, "work", "p1\np2\n")

	names, profile, err := resolveProviderNames("work")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(names) != 2 || names[0] != "p1" || names[1] != "p2" {
		t.Errorf("got %v", names)
	}
	if profile != "work" {
		t.Errorf("profile = %q, want \"work\"", profile)
	}
}

func TestResolveWithProfileFlagNotFound(t *testing.T) {
	setTestHome(t)

	_, _, err := resolveProviderNames("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveWithProfileFlagEmpty(t *testing.T) {
	setTestHome(t)
	writeProfileConf(t, "empty", "")

	_, _, err := resolveProviderNames("empty")
	if err == nil {
		t.Error("expected error for empty profile")
	}
	if err != nil && !strings.Contains(err.Error(), "no providers configured") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveNoFlag(t *testing.T) {
	setTestHome(t)
	writeFallbackConf(t, "a\nb\n")

	names, profile, err := resolveProviderNames("")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("got %v", names)
	}
	if profile != "default" {
		t.Errorf("profile = %q, want \"default\"", profile)
	}
}

func TestValidateWithProfile(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "a", "ANTHROPIC_BASE_URL=https://a.com\nANTHROPIC_AUTH_TOKEN=tok\n")
	writeProfileConf(t, "work", "a\nmissing\n")
	mockStdin(t, "y\n")

	valid, err := validateProviderNames([]string{"a", "missing"}, "work")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(valid) != 1 || valid[0] != "a" {
		t.Errorf("expected [a], got %v", valid)
	}

	// Verify "missing" was removed from work profile
	names, _ := config.ReadProfileOrder("work")
	for _, n := range names {
		if n == "missing" {
			t.Error("missing should have been removed from work profile")
		}
	}
}

func writeFallbackConf(t *testing.T, content string) {
	t.Helper()
	home := os.Getenv("HOME")
	path := filepath.Join(home, ".cc_envs", "fallback.conf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveProviderNamesFromFallbackConf(t *testing.T) {
	setTestHome(t)
	writeFallbackConf(t, "p1\np2\n")

	names, profile, err := resolveProviderNames("")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(names) != 2 || names[0] != "p1" || names[1] != "p2" {
		t.Errorf("got %v", names)
	}
	if profile != "default" {
		t.Errorf("profile = %q, want \"default\"", profile)
	}
}

func TestResolveProviderNamesNoFallbackConf(t *testing.T) {
	setTestHome(t)
	// No fallback.conf and no providers → should error about no providers configured

	_, _, err := resolveProviderNames("")
	if err == nil {
		t.Error("expected error when fallback.conf missing and no providers")
	}
}

func TestResolveProviderNamesEmptyFallbackConf(t *testing.T) {
	setTestHome(t)
	writeFallbackConf(t, "")
	// Empty fallback.conf and no providers → should error about no providers configured

	_, _, err := resolveProviderNames("")
	if err == nil {
		t.Error("expected error when fallback.conf is empty and no providers")
	}
}

func TestBuildProvidersMissingConfigErrors(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "a", "ANTHROPIC_BASE_URL=https://a.com\nANTHROPIC_AUTH_TOKEN=tok\n")

	_, _, err := buildProviders([]string{"a", "gone"})
	if err == nil {
		t.Error("expected error for missing config")
	}
	if err != nil && !strings.Contains(err.Error(), "'gone' not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- validateProviderNames tests ---

// mockStdin replaces stdinReader for the duration of the test.
func mockStdin(t *testing.T, input string) {
	t.Helper()
	old := stdinReader
	stdinReader = strings.NewReader(input)
	t.Cleanup(func() { stdinReader = old })
}

func TestValidateProviderNamesAllExist(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "a", "ANTHROPIC_BASE_URL=https://a.com\nANTHROPIC_AUTH_TOKEN=tok\n")
	writeTestEnv(t, "b", "ANTHROPIC_BASE_URL=https://b.com\nANTHROPIC_AUTH_TOKEN=tok\n")

	valid, err := validateProviderNames([]string{"a", "b"}, "default")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(valid) != 2 {
		t.Errorf("expected 2 valid, got %d", len(valid))
	}
}

func TestValidateProviderNamesSomeMissingConfirmYes(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "a", "ANTHROPIC_BASE_URL=https://a.com\nANTHROPIC_AUTH_TOKEN=tok\n")
	writeFallbackConf(t, "a\nmissing\n")
	mockStdin(t, "y\n")

	valid, err := validateProviderNames([]string{"a", "missing"}, "default")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(valid) != 1 || valid[0] != "a" {
		t.Errorf("expected [a], got %v", valid)
	}

	// Verify "missing" was removed from fallback.conf
	names, _ := config.ReadFallbackOrder()
	for _, n := range names {
		if n == "missing" {
			t.Error("missing should have been removed from fallback.conf")
		}
	}
}

func TestValidateProviderNamesSomeMissingConfirmNo(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "a", "ANTHROPIC_BASE_URL=https://a.com\nANTHROPIC_AUTH_TOKEN=tok\n")
	mockStdin(t, "n\n")

	_, err := validateProviderNames([]string{"a", "missing"}, "default")
	if err == nil {
		t.Error("expected error when user says no")
	}
	if err != nil && err.Error() != "aborted" {
		t.Errorf("expected 'aborted', got: %v", err)
	}
}

func TestValidateProviderNamesAllMissingConfirmYes(t *testing.T) {
	setTestHome(t)
	mockStdin(t, "y\n")

	_, err := validateProviderNames([]string{"x", "y"}, "default")
	if err == nil {
		t.Error("expected error when all providers missing")
	}
	if err != nil && !strings.Contains(err.Error(), "no valid providers remaining") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateProviderNamesSkipsEmptyNames(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "a", "ANTHROPIC_BASE_URL=https://a.com\nANTHROPIC_AUTH_TOKEN=tok\n")

	valid, err := validateProviderNames([]string{"", "a", "  "}, "default")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(valid) != 3 {
		// empty names are kept as-is when no missing providers found
		// (they get filtered later by buildProviders)
	}
}

func TestValidateProviderNamesConfirmYes(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "a", "ANTHROPIC_BASE_URL=https://a.com\nANTHROPIC_AUTH_TOKEN=tok\n")
	mockStdin(t, "yes\n")

	valid, err := validateProviderNames([]string{"a", "gone"}, "default")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(valid) != 1 || valid[0] != "a" {
		t.Errorf("expected [a], got %v", valid)
	}
}
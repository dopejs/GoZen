package cmd

import (
	"os"
	"path/filepath"
	"testing"
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

func writeFallbackConf(t *testing.T, content string) {
	t.Helper()
	home := os.Getenv("HOME")
	path := filepath.Join(home, ".cc_envs", "fallback.conf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveProviderNamesFromFlag(t *testing.T) {
	setTestHome(t)

	names, err := resolveProviderNames(true, "a,b,c")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(names) != 3 || names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Errorf("got %v", names)
	}
}

func TestResolveProviderNamesFallbackFlagNoValue(t *testing.T) {
	setTestHome(t)
	writeFallbackConf(t, "x\ny\n")

	// -f with empty value → reads fallback.conf
	names, err := resolveProviderNames(true, " ")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(names) != 2 || names[0] != "x" || names[1] != "y" {
		t.Errorf("got %v", names)
	}
}

func TestResolveProviderNamesNoFlagReadsFallback(t *testing.T) {
	setTestHome(t)
	writeFallbackConf(t, "p1\np2\n")

	names, err := resolveProviderNames(false, "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(names) != 2 || names[0] != "p1" {
		t.Errorf("got %v", names)
	}
}

func TestResolveProviderNamesNoFallbackConf(t *testing.T) {
	setTestHome(t)
	// No fallback.conf

	_, err := resolveProviderNames(false, "")
	if err == nil {
		t.Error("expected error when fallback.conf missing")
	}
}

func TestResolveProviderNamesEmptyFallbackConf(t *testing.T) {
	setTestHome(t)
	writeFallbackConf(t, "")

	_, err := resolveProviderNames(false, "")
	if err == nil {
		t.Error("expected error when fallback.conf is empty")
	}
}

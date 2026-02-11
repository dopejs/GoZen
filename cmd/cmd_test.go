package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunListNoConfigs(t *testing.T) {
	setTestHome(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, nil)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runList() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No configurations found") {
		t.Errorf("expected 'No configurations found', got %q", output)
	}
}

func TestRunListWithConfigs(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "work", "ANTHROPIC_BASE_URL=https://work.example.com\nANTHROPIC_AUTH_TOKEN=tok\nANTHROPIC_MODEL=opus\n")
	writeTestEnv(t, "personal", "ANTHROPIC_BASE_URL=https://personal.example.com\nANTHROPIC_AUTH_TOKEN=tok2\nANTHROPIC_MODEL=sonnet\n")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, nil)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runList() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "work") {
		t.Errorf("expected 'work' in output, got %q", output)
	}
	if !strings.Contains(output, "personal") {
		t.Errorf("expected 'personal' in output, got %q", output)
	}
	if !strings.Contains(output, "opus") {
		t.Errorf("expected 'opus' in output, got %q", output)
	}
}

func TestRunListWithPrefix(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "work", "ANTHROPIC_BASE_URL=https://work.example.com\nANTHROPIC_AUTH_TOKEN=tok\nANTHROPIC_MODEL=opus\n")
	writeTestEnv(t, "personal", "ANTHROPIC_BASE_URL=https://personal.example.com\nANTHROPIC_AUTH_TOKEN=tok2\n")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, []string{"wo"})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runList() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "work") {
		t.Errorf("expected 'work' in output, got %q", output)
	}
	if strings.Contains(output, "personal") {
		t.Errorf("should not contain 'personal', got %q", output)
	}
}

func TestRunListPrefixNoMatch(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "work", "ANTHROPIC_BASE_URL=https://work.example.com\nANTHROPIC_AUTH_TOKEN=tok\n")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, []string{"zzz"})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runList() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No configurations matching") {
		t.Errorf("expected 'No configurations matching', got %q", output)
	}
}

func TestRunListWithFallbackAnnotation(t *testing.T) {
	home := setTestHome(t)
	writeTestEnv(t, "yunyi", "ANTHROPIC_BASE_URL=https://yunyi.example.com\nANTHROPIC_AUTH_TOKEN=tok\nANTHROPIC_MODEL=opus\n")
	writeTestEnv(t, "cctq", "ANTHROPIC_BASE_URL=https://cctq.example.com\nANTHROPIC_AUTH_TOKEN=tok2\n")

	// Write fallback.conf
	fbPath := filepath.Join(home, ".cc_envs", "fallback.conf")
	os.WriteFile(fbPath, []byte("yunyi\ncctq\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, nil)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runList() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "[1]") {
		t.Errorf("expected fallback annotation [1], got %q", output)
	}
	if !strings.Contains(output, "[2]") {
		t.Errorf("expected fallback annotation [2], got %q", output)
	}
}

func TestCompleteConfigNames(t *testing.T) {
	setTestHome(t)
	writeTestEnv(t, "alpha", "K=V\n")
	writeTestEnv(t, "beta", "K=V\n")

	names, directive := completeConfigNames(nil, nil, "")
	if directive != 4 { // cobra.ShellCompDirectiveNoFileComp = 4
		t.Errorf("directive = %d", directive)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d: %v", len(names), names)
	}
}

func TestRunCompletion(t *testing.T) {
	tests := []struct {
		shell   string
		wantErr bool
	}{
		{"zsh", false},
		{"bash", false},
		{"fish", false},
		{"powershell", false},
		{"invalid", false}, // prints error but doesn't return error
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			// Redirect stdout to avoid noise
			old := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			err := runCompletion(completionCmd, []string{tt.shell})

			w.Close()
			os.Stdout = old

			if (err != nil) != tt.wantErr {
				t.Errorf("runCompletion(%q) error = %v, wantErr %v", tt.shell, err, tt.wantErr)
			}
		})
	}
}

func TestRunListDefaultValues(t *testing.T) {
	setTestHome(t)
	// Config with no model and no base_url
	writeTestEnv(t, "minimal", "ANTHROPIC_AUTH_TOKEN=tok\n")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, nil)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runList() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should show "-" for missing model and base_url
	if !strings.Contains(output, "minimal") {
		t.Errorf("expected 'minimal' in output, got %q", output)
	}
}

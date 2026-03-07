package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func TestCompleteConfigNames(t *testing.T) {
	setTestHome(t)
	writeTestProvider(t, "alpha", &config.ProviderConfig{BaseURL: "https://a.com", AuthToken: "tok"})
	writeTestProvider(t, "beta", &config.ProviderConfig{BaseURL: "https://b.com", AuthToken: "tok"})

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

func TestPrintProviders(t *testing.T) {
	setTestHome(t)
	writeTestProvider(t, "work", &config.ProviderConfig{
		BaseURL:   "https://api.work.com",
		AuthToken: "tok",
		Model:     "claude-sonnet-4-5",
	})
	writeTestProvider(t, "personal", &config.ProviderConfig{
		BaseURL:   "https://api.personal.com",
		AuthToken: "tok",
	})

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printProviders()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Providers:") {
		t.Error("expected 'Providers:' header")
	}
	if !strings.Contains(output, "work") {
		t.Error("expected provider 'work' in output")
	}
	if !strings.Contains(output, "personal") {
		t.Error("expected provider 'personal' in output")
	}
	if !strings.Contains(output, "https://api.work.com") {
		t.Error("expected base URL in output")
	}
	if !strings.Contains(output, "claude-sonnet-4-5") {
		t.Error("expected model in output")
	}
}

func TestPrintProvidersWithDisabled(t *testing.T) {
	setTestHome(t)
	writeTestProvider(t, "work", &config.ProviderConfig{
		BaseURL:   "https://api.work.com",
		AuthToken: "tok",
		Model:     "claude-sonnet-4-5",
	})
	writeTestProvider(t, "personal", &config.ProviderConfig{
		BaseURL:   "https://api.personal.com",
		AuthToken: "tok",
	})
	config.DisableProvider("work", config.MarkingTypeToday)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printProviders()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "[disabled: today]") {
		t.Errorf("expected '[disabled: today]' tag, got: %s", output)
	}
	if strings.Contains(output, "personal") && strings.Contains(output, "[disabled") {
		// personal shouldn't have disabled tag — just check it appears without the tag
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "personal") && strings.Contains(line, "[disabled") {
				t.Errorf("personal should not have disabled tag, got: %s", line)
			}
		}
	}
}

func TestPrintProvidersEmpty(t *testing.T) {
	setTestHome(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printProviders()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No providers configured") {
		t.Error("expected 'No providers configured' message")
	}
}

func TestPrintProfiles(t *testing.T) {
	setTestHome(t)
	writeTestProvider(t, "a", &config.ProviderConfig{BaseURL: "https://a.com", AuthToken: "t"})
	writeTestProvider(t, "b", &config.ProviderConfig{BaseURL: "https://b.com", AuthToken: "t"})
	writeProfileConf(t, "default", []string{"a", "b"})
	writeProfileConf(t, "work", []string{"b"})

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printProfiles()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Profiles:") {
		t.Error("expected 'Profiles:' header")
	}
	if !strings.Contains(output, "default") {
		t.Error("expected 'default' profile")
	}
	if !strings.Contains(output, "work") {
		t.Error("expected 'work' profile")
	}
	if !strings.Contains(output, "(default)") {
		t.Error("expected '(default)' tag")
	}
	if !strings.Contains(output, "a → b") {
		t.Errorf("expected provider chain 'a → b', got: %s", output)
	}
}

// --- US4: zen use bypasses disabled check ---

func TestUseBypassesDisabledCheck(t *testing.T) {
	setTestHome(t)
	writeTestProvider(t, "work", &config.ProviderConfig{
		BaseURL:   "https://api.work.com",
		AuthToken: "tok",
		Model:     "claude-sonnet-4-5",
	})
	// Disable the provider
	config.DisableProvider("work", config.MarkingTypePermanent)
	if !config.IsProviderDisabled("work") {
		t.Fatal("provider should be disabled for test setup")
	}

	// Verify ExportProviderToEnv still works (used by zen use)
	if err := config.ExportProviderToEnv("work"); err != nil {
		t.Errorf("ExportProviderToEnv should succeed for disabled provider, got: %v", err)
	}
}

// --- disable/enable command tests ---

// resetDisableFlags resets the package-level flag variables that persist across subtests.
func resetDisableFlags(t *testing.T) {
	t.Helper()
	disableTodayFlag = false
	disableMonthFlag = false
	disablePermanentFlag = false
	disableListFlag = false
}

func TestRunDisable(t *testing.T) {
	t.Run("disable provider with default (today)", func(t *testing.T) {
		setTestHome(t)
		resetDisableFlags(t)
		writeTestProvider(t, "work", &config.ProviderConfig{
			BaseURL:   "https://api.work.com",
			AuthToken: "tok",
		})

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"disable", "work"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		buf.ReadFrom(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "unavailable") {
			t.Errorf("expected unavailable message, got: %s", output)
		}
		if !config.IsProviderDisabled("work") {
			t.Error("provider should be disabled")
		}
	})

	t.Run("disable with --month flag", func(t *testing.T) {
		setTestHome(t)
		resetDisableFlags(t)
		writeTestProvider(t, "work", &config.ProviderConfig{
			BaseURL:   "https://api.work.com",
			AuthToken: "tok",
		})

		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"disable", "--month", "work"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		disabled := config.GetDisabledProviders()
		if m, ok := disabled["work"]; !ok || m.Type != config.MarkingTypeMonth {
			t.Errorf("expected month marking, got %+v", disabled["work"])
		}
	})

	t.Run("disable with --permanent flag", func(t *testing.T) {
		setTestHome(t)
		resetDisableFlags(t)
		writeTestProvider(t, "work", &config.ProviderConfig{
			BaseURL:   "https://api.work.com",
			AuthToken: "tok",
		})

		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"disable", "--permanent", "work"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		disabled := config.GetDisabledProviders()
		if m, ok := disabled["work"]; !ok || m.Type != config.MarkingTypePermanent {
			t.Errorf("expected permanent marking, got %+v", disabled["work"])
		}
	})

	t.Run("disable nonexistent provider returns error", func(t *testing.T) {
		setTestHome(t)
		resetDisableFlags(t)

		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"disable", "nonexistent"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err == nil {
			t.Error("expected error for nonexistent provider")
		}
	})

	t.Run("disable --list shows disabled providers", func(t *testing.T) {
		setTestHome(t)
		resetDisableFlags(t)
		writeTestProvider(t, "work", &config.ProviderConfig{
			BaseURL:   "https://api.work.com",
			AuthToken: "tok",
		})
		config.DisableProvider("work", config.MarkingTypeToday)

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"disable", "--list"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		buf.ReadFrom(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "work") {
			t.Errorf("expected 'work' in disabled list, got: %s", output)
		}
	})

	t.Run("disable no args shows usage", func(t *testing.T) {
		setTestHome(t)
		resetDisableFlags(t)

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"disable"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		buf.ReadFrom(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "Usage:") {
			t.Errorf("expected usage message, got: %s", output)
		}
	})
}

func TestRunEnable(t *testing.T) {
	t.Run("enable disabled provider", func(t *testing.T) {
		setTestHome(t)
		writeTestProvider(t, "work", &config.ProviderConfig{
			BaseURL:   "https://api.work.com",
			AuthToken: "tok",
		})
		config.DisableProvider("work", config.MarkingTypeToday)

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"enable", "work"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		buf.ReadFrom(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "enabled") {
			t.Errorf("expected enabled message, got: %s", output)
		}
		if config.IsProviderDisabled("work") {
			t.Error("provider should no longer be disabled")
		}
	})

	t.Run("enable already enabled provider", func(t *testing.T) {
		setTestHome(t)
		writeTestProvider(t, "work", &config.ProviderConfig{
			BaseURL:   "https://api.work.com",
			AuthToken: "tok",
		})

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"enable", "work"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		buf.ReadFrom(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "not currently disabled") {
			t.Errorf("expected 'not currently disabled' message, got: %s", output)
		}
	})

	t.Run("enable nonexistent provider returns error", func(t *testing.T) {
		setTestHome(t)

		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"enable", "nonexistent"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err == nil {
			t.Error("expected error for nonexistent provider")
		}
	})

	t.Run("enable no args shows usage", func(t *testing.T) {
		setTestHome(t)

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"enable"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		buf.ReadFrom(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "Usage:") {
			t.Errorf("expected usage message, got: %s", output)
		}
	})
}

func TestConfigEditDeleteRequiresArgs(t *testing.T) {
	setTestHome(t)

	cmds := []struct {
		name string
		args []string
	}{
		{"edit provider", []string{"config", "edit", "provider"}},
		{"edit profile", []string{"config", "edit", "profile"}},
		{"delete provider", []string{"config", "delete", "provider"}},
		{"delete profile", []string{"config", "delete", "profile"}},
	}

	for _, tt := range cmds {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout to capture usage output
			old := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			w.Close()
			os.Stdout = old

			// Should not return an error (shows usage instead)
			if err != nil {
				t.Errorf("expected no error for missing args, got: %v", err)
			}
		})
	}
}

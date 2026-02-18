package proxy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func TestNewBudgetChecker(t *testing.T) {
	// Setup temp config
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	tracker := NewUsageTracker(nil)
	checker := NewBudgetChecker(tracker)

	if checker == nil {
		t.Fatal("Expected non-nil budget checker")
	}
	if checker.tracker != tracker {
		t.Error("Expected tracker to be set")
	}
}

func TestBudgetChecker_Check_NilTracker(t *testing.T) {
	checker := &BudgetChecker{tracker: nil}

	status, err := checker.Check("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("Expected non-nil status")
	}
}

func TestBudgetChecker_Check_NilConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	tracker := NewUsageTracker(nil)
	checker := NewBudgetChecker(tracker)

	status, err := checker.Check("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("Expected non-nil status")
	}
}

func TestBudgetChecker_ReloadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	checker := NewBudgetChecker(nil)

	// Set budget config
	config.SetBudgets(&config.BudgetConfig{
		Daily: &config.BudgetLimit{Amount: 10.0, Action: "warn"},
	})

	checker.ReloadConfig()

	if checker.config == nil {
		t.Error("Expected config to be reloaded")
	}
}

func TestBudgetChecker_ShouldBlock(t *testing.T) {
	checker := &BudgetChecker{tracker: nil}

	// With nil tracker, should not block
	if checker.ShouldBlock("") {
		t.Error("Expected ShouldBlock to return false with nil tracker")
	}
}

func TestBudgetChecker_ShouldDowngrade(t *testing.T) {
	checker := &BudgetChecker{tracker: nil}

	// With nil tracker, should not downgrade
	if checker.ShouldDowngrade("") {
		t.Error("Expected ShouldDowngrade to return false with nil tracker")
	}
}

func TestBudgetChecker_GetDowngradeModel(t *testing.T) {
	checker := &BudgetChecker{}

	model := checker.GetDowngradeModel("claude-3-opus-20240229")
	if model != "claude-3-5-haiku-20241022" {
		t.Errorf("Expected haiku model, got %s", model)
	}
}

func TestGlobalBudgetChecker(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Save old global
	oldGlobal := globalBudgetChecker
	defer func() { globalBudgetChecker = oldGlobal }()

	tracker := NewUsageTracker(nil)
	InitGlobalBudgetChecker(tracker)

	checker := GetGlobalBudgetChecker()
	if checker == nil {
		t.Error("Expected non-nil global budget checker")
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{50.0, "50%"},
		{100.0, "100%"},
		{150.0, "100%"},
		{0.0, "00%"},
		{99.9, "99%"},
	}

	for _, tt := range tests {
		result := formatPercent(tt.input)
		if result != tt.expected {
			t.Errorf("formatPercent(%v) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestBudgetChecker_Check_WithDailyLimit(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Set budget config with daily limit
	config.SetBudgets(&config.BudgetConfig{
		Daily: &config.BudgetLimit{Amount: 10.0, Action: "warn"},
	})

	tracker := NewUsageTracker(nil)
	checker := NewBudgetChecker(tracker)
	checker.ReloadConfig()

	status, err := checker.Check("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status.DailyLimit != 10.0 {
		t.Errorf("Expected daily limit 10.0, got %f", status.DailyLimit)
	}
}

func TestBudgetChecker_Check_WithWeeklyLimit(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Set budget config with weekly limit
	config.SetBudgets(&config.BudgetConfig{
		Weekly: &config.BudgetLimit{Amount: 50.0, Action: "downgrade"},
	})

	tracker := NewUsageTracker(nil)
	checker := NewBudgetChecker(tracker)
	checker.ReloadConfig()

	status, err := checker.Check("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status.WeeklyLimit != 50.0 {
		t.Errorf("Expected weekly limit 50.0, got %f", status.WeeklyLimit)
	}
}

func TestBudgetChecker_Check_WithMonthlyLimit(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Set budget config with monthly limit
	config.SetBudgets(&config.BudgetConfig{
		Monthly: &config.BudgetLimit{Amount: 200.0, Action: "block"},
	})

	tracker := NewUsageTracker(nil)
	checker := NewBudgetChecker(tracker)
	checker.ReloadConfig()

	status, err := checker.Check("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status.MonthlyLimit != 200.0 {
		t.Errorf("Expected monthly limit 200.0, got %f", status.MonthlyLimit)
	}
}

func TestBudgetChecker_Check_PerProject(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Set budget config with per-project enabled
	config.SetBudgets(&config.BudgetConfig{
		Daily:      &config.BudgetLimit{Amount: 10.0, Action: "warn"},
		PerProject: true,
	})

	tracker := NewUsageTracker(nil)
	checker := NewBudgetChecker(tracker)
	checker.ReloadConfig()

	status, err := checker.Check("/path/to/project")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("Expected non-nil status")
	}
}

func TestBudgetChecker_Check_AllLimits(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Set budget config with all limits
	config.SetBudgets(&config.BudgetConfig{
		Daily:   &config.BudgetLimit{Amount: 10.0, Action: "warn"},
		Weekly:  &config.BudgetLimit{Amount: 50.0, Action: "downgrade"},
		Monthly: &config.BudgetLimit{Amount: 200.0, Action: "block"},
	})

	tracker := NewUsageTracker(nil)
	checker := NewBudgetChecker(tracker)
	checker.ReloadConfig()

	status, err := checker.Check("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status.DailyLimit != 10.0 {
		t.Errorf("Expected daily limit 10.0, got %f", status.DailyLimit)
	}
	if status.WeeklyLimit != 50.0 {
		t.Errorf("Expected weekly limit 50.0, got %f", status.WeeklyLimit)
	}
	if status.MonthlyLimit != 200.0 {
		t.Errorf("Expected monthly limit 200.0, got %f", status.MonthlyLimit)
	}
}

package proxy

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

func TestNewUsageTracker(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	tracker := NewUsageTracker(nil)
	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}
	if tracker.pricing == nil {
		t.Error("Expected pricing to be loaded")
	}
}

func TestUsageTracker_ReloadPricing(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	tracker := NewUsageTracker(nil)
	tracker.ReloadPricing()

	if tracker.pricing == nil {
		t.Error("Expected pricing after reload")
	}
}

func TestUsageTracker_FindPricing(t *testing.T) {
	tracker := &UsageTracker{
		pricing: map[string]*config.ModelPricing{
			"claude-3-opus-20240229":     {InputPerMillion: 15.0, OutputPerMillion: 75.0},
			"claude-3-5-sonnet-20241022": {InputPerMillion: 3.0, OutputPerMillion: 15.0},
			"claude-3-5-haiku-20241022":  {InputPerMillion: 0.8, OutputPerMillion: 4.0},
		},
	}

	tests := []struct {
		model    string
		expected bool
	}{
		{"claude-3-opus-20240229", true},
		{"claude-3-5-sonnet-20241022", true},
		{"some-opus-model", true},    // partial match
		{"some-sonnet-model", true},  // partial match
		{"some-haiku-model", true},   // partial match
		{"unknown-model", false},
	}

	for _, tt := range tests {
		result := tracker.findPricing(tt.model)
		if (result != nil) != tt.expected {
			t.Errorf("findPricing(%s) = %v, want found=%v", tt.model, result, tt.expected)
		}
	}
}

func TestUsageTracker_Record_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	err := tracker.Record(UsageEntry{
		SessionID:    "test",
		Provider:     "test",
		Model:        "test",
		InputTokens:  100,
		OutputTokens: 50,
	})

	if err != nil {
		t.Errorf("Expected no error for nil db, got %v", err)
	}
}

func TestUsageTracker_GetSummary_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	summary, err := tracker.GetSummary("day", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("Expected non-nil summary")
	}
	if summary.ByProvider == nil {
		t.Error("Expected ByProvider map to be initialized")
	}
}

func TestUsageTracker_GetDailyCost_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	cost, err := tracker.GetDailyCost("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if cost != 0 {
		t.Errorf("Expected 0 cost for nil db, got %f", cost)
	}
}

func TestUsageTracker_GetWeeklyCost_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	cost, err := tracker.GetWeeklyCost("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if cost != 0 {
		t.Errorf("Expected 0 cost for nil db, got %f", cost)
	}
}

func TestUsageTracker_GetMonthlyCost_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	cost, err := tracker.GetMonthlyCost("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if cost != 0 {
		t.Errorf("Expected 0 cost for nil db, got %f", cost)
	}
}

func TestUsageTracker_AggregateHourly_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	err := tracker.AggregateHourly()
	if err != nil {
		t.Errorf("Expected no error for nil db, got %v", err)
	}
}

func TestUsageTracker_GetHourlySummary_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	result, err := tracker.GetHourlySummary(24)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for nil db")
	}
}

func TestUsageTracker_GetRecentUsage_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	result, err := tracker.GetRecentUsage(10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for nil db")
	}
}

func TestGlobalUsageTracker(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Save old global
	oldGlobal := globalUsageTracker
	defer func() { globalUsageTracker = oldGlobal }()

	InitGlobalUsageTracker(nil)

	tracker := GetGlobalUsageTracker()
	if tracker == nil {
		t.Error("Expected non-nil global usage tracker")
	}
}

func TestUsageEntry_Struct(t *testing.T) {
	entry := UsageEntry{
		SessionID:    "session-123",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
		LatencyMs:    1500,
		ProjectPath:  "/test/project",
		ClientType:   "claude",
	}

	if entry.SessionID != "session-123" {
		t.Errorf("Expected session-123, got %s", entry.SessionID)
	}
	if entry.InputTokens != 1000 {
		t.Errorf("Expected 1000 input tokens, got %d", entry.InputTokens)
	}
}

func TestUsageSummary_Struct(t *testing.T) {
	summary := &UsageSummary{
		TotalInputTokens:  10000,
		TotalOutputTokens: 5000,
		TotalCost:         1.50,
		RequestCount:      100,
		ByProvider:        make(map[string]*UsageStats),
		ByModel:           make(map[string]*UsageStats),
		ByProject:         make(map[string]*UsageStats),
	}

	if summary.TotalCost != 1.50 {
		t.Errorf("Expected 1.50 total cost, got %f", summary.TotalCost)
	}
	if summary.RequestCount != 100 {
		t.Errorf("Expected 100 requests, got %d", summary.RequestCount)
	}
}

func TestUsageStats_Struct(t *testing.T) {
	stats := &UsageStats{
		InputTokens:  5000,
		OutputTokens: 2500,
		Cost:         0.75,
		RequestCount: 50,
	}

	if stats.Cost != 0.75 {
		t.Errorf("Expected 0.75 cost, got %f", stats.Cost)
	}
}

func TestHourlyUsage_Struct(t *testing.T) {
	now := time.Now()
	hourly := &HourlyUsage{
		Hour:         now,
		InputTokens:  1000,
		OutputTokens: 500,
		Cost:         0.10,
		RequestCount: 10,
	}

	if hourly.Hour != now {
		t.Errorf("Expected hour time, got %v", hourly.Hour)
	}
	if hourly.RequestCount != 10 {
		t.Errorf("Expected 10 requests, got %d", hourly.RequestCount)
	}
}

func TestUsageTracker_Record_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	err = tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
		LatencyMs:    1500,
		ProjectPath:  "/test/project",
		ClientType:   "claude",
	})

	if err != nil {
		t.Errorf("Record() error: %v", err)
	}
}

func TestUsageTracker_GetSummary_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// Record some usage
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
	})

	// Test different periods
	for _, period := range []string{"day", "week", "month", "all"} {
		summary, err := tracker.GetSummary(period, "")
		if err != nil {
			t.Errorf("GetSummary(%s) error: %v", period, err)
		}
		if summary == nil {
			t.Errorf("GetSummary(%s) returned nil", period)
		}
	}
}

func TestUsageTracker_GetRecentUsage_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// Record some usage
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
	})

	entries, err := tracker.GetRecentUsage(10)
	if err != nil {
		t.Errorf("GetRecentUsage() error: %v", err)
	}
	if entries == nil {
		t.Error("GetRecentUsage() returned nil")
	}
}

func TestUsageTracker_GetHourlySummary_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// GetHourlySummary may return nil if no data, that's acceptable
	_, err = tracker.GetHourlySummary(24)
	if err != nil {
		t.Errorf("GetHourlySummary() error: %v", err)
	}
}

func TestUsageTracker_AggregateHourly_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// Record some usage
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
	})

	err = tracker.AggregateHourly()
	if err != nil {
		t.Errorf("AggregateHourly() error: %v", err)
	}
}

func TestUsageTracker_GetCosts_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// Record some usage
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
	})

	// Test GetDailyCost
	cost, err := tracker.GetDailyCost("")
	if err != nil {
		t.Errorf("GetDailyCost() error: %v", err)
	}
	if cost < 0 {
		t.Errorf("GetDailyCost() returned negative cost: %f", cost)
	}

	// Test GetWeeklyCost
	cost, err = tracker.GetWeeklyCost("")
	if err != nil {
		t.Errorf("GetWeeklyCost() error: %v", err)
	}
	if cost < 0 {
		t.Errorf("GetWeeklyCost() returned negative cost: %f", cost)
	}

	// Test GetMonthlyCost
	cost, err = tracker.GetMonthlyCost("")
	if err != nil {
		t.Errorf("GetMonthlyCost() error: %v", err)
	}
	if cost < 0 {
		t.Errorf("GetMonthlyCost() returned negative cost: %f", cost)
	}
}

func TestUsageTracker_GetHourlySummaryByProvider_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// Record some usage
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
	})

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now().Add(time.Hour)

	result, err := tracker.GetHourlySummaryByProvider(since, until)
	if err != nil {
		t.Errorf("GetHourlySummaryByProvider() error: %v", err)
	}
	// Result may be empty or have data, both are valid
	_ = result
}

func TestUsageTracker_GetHourlySummaryByModel_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// Record some usage
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
	})

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now().Add(time.Hour)

	result, err := tracker.GetHourlySummaryByModel(since, until)
	if err != nil {
		t.Errorf("GetHourlySummaryByModel() error: %v", err)
	}
	_ = result
}

func TestUsageTracker_GetSummaryByTimeRange_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// Record some usage with project path
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
		ProjectPath:  "/test/project",
	})

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now().Add(time.Hour)

	// Test without project filter
	summary, err := tracker.GetSummaryByTimeRange(since, until, "")
	if err != nil {
		t.Errorf("GetSummaryByTimeRange() error: %v", err)
	}
	if summary == nil {
		t.Error("GetSummaryByTimeRange() returned nil")
	}

	// Test with project filter
	summary, err = tracker.GetSummaryByTimeRange(since, until, "/test/project")
	if err != nil {
		t.Errorf("GetSummaryByTimeRange() with project error: %v", err)
	}
	if summary == nil {
		t.Error("GetSummaryByTimeRange() with project returned nil")
	}
}

func TestUsageTracker_GetSummaryByTimeRange_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	summary, err := tracker.GetSummaryByTimeRange(since, until, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("Expected non-nil summary")
	}
	if summary.ByProvider == nil {
		t.Error("Expected ByProvider map to be initialized")
	}
}

func TestUsageTracker_GetRecentPaths_WithDB(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	ldb, err := OpenLogDB(filepath.Join(configDir, "logs"))
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	tracker := &UsageTracker{db: ldb, pricing: make(map[string]*config.ModelPricing)}

	// Record usage with project paths
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session-1",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
		ProjectPath:  "/test/project1",
	})
	tracker.Record(UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    "test-session-2",
		Provider:     "anthropic",
		Model:        "claude-3-opus",
		InputTokens:  1000,
		OutputTokens: 500,
		CostUSD:      0.05,
		ProjectPath:  "/test/project2",
	})

	paths, err := tracker.GetRecentPaths(10)
	if err != nil {
		t.Errorf("GetRecentPaths() error: %v", err)
	}
	if len(paths) == 0 {
		t.Error("GetRecentPaths() returned empty")
	}

	// Test with default limit
	paths, err = tracker.GetRecentPaths(0)
	if err != nil {
		t.Errorf("GetRecentPaths(0) error: %v", err)
	}
	_ = paths
}

func TestUsageTracker_GetHourlySummaryByDimension_NilDB(t *testing.T) {
	tracker := &UsageTracker{db: nil}

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	result, err := tracker.GetHourlySummaryByProvider(since, until)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for nil db")
	}

	result, err = tracker.GetHourlySummaryByModel(since, until)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for nil db")
	}
}

func TestHourlyUsageByDimension_Struct(t *testing.T) {
	now := time.Now()
	h := HourlyUsageByDimension{
		Hour:         now,
		Dimension:    "anthropic",
		InputTokens:  1000,
		OutputTokens: 500,
		Cost:         0.10,
		RequestCount: 10,
	}

	if h.Dimension != "anthropic" {
		t.Errorf("Expected anthropic, got %s", h.Dimension)
	}
	if h.RequestCount != 10 {
		t.Errorf("Expected 10 requests, got %d", h.RequestCount)
	}
}

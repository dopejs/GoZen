package proxy

import (
	"testing"
	"time"
)

func TestRecordMetric(t *testing.T) {
	dir := t.TempDir()
	ldb, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	// Record a successful metric
	err = ldb.RecordMetric("provider1", 100, 200, false, false)
	if err != nil {
		t.Errorf("RecordMetric() error: %v", err)
	}

	// Record an error metric
	err = ldb.RecordMetric("provider1", 500, 500, true, false)
	if err != nil {
		t.Errorf("RecordMetric() error: %v", err)
	}

	// Record a rate limit metric
	err = ldb.RecordMetric("provider2", 50, 429, false, true)
	if err != nil {
		t.Errorf("RecordMetric() error: %v", err)
	}
}

func TestRecordMetric_NilDB(t *testing.T) {
	var ldb *LogDB
	err := ldb.RecordMetric("provider", 100, 200, false, false)
	if err != nil {
		t.Errorf("Expected nil error for nil db, got: %v", err)
	}
}

func TestGetProviderMetrics_NilDB(t *testing.T) {
	var ldb *LogDB
	metrics, err := ldb.GetProviderMetrics("provider", time.Now())
	if err != nil {
		t.Errorf("Expected nil error for nil db, got: %v", err)
	}
	if metrics == nil {
		t.Error("Expected non-nil metrics")
	}
}

func TestGetAllProviderMetrics_NilDB(t *testing.T) {
	var ldb *LogDB
	metrics, err := ldb.GetAllProviderMetrics(time.Now())
	if err != nil {
		t.Errorf("Expected nil error for nil db, got: %v", err)
	}
	if metrics == nil {
		t.Error("Expected non-nil map")
	}
}

func TestGetLatencyHistory_NilDB(t *testing.T) {
	var ldb *LogDB
	history, err := ldb.GetLatencyHistory("provider", time.Now(), 5)
	if err != nil {
		t.Errorf("Expected nil error for nil db, got: %v", err)
	}
	if history != nil {
		t.Error("Expected nil history for nil db")
	}
}

func TestCleanupOldMetrics_NilDB(t *testing.T) {
	var ldb *LogDB
	deleted, err := ldb.CleanupOldMetrics(time.Hour)
	if err != nil {
		t.Errorf("Expected nil error for nil db, got: %v", err)
	}
	if deleted != 0 {
		t.Errorf("Expected 0 deleted for nil db, got %d", deleted)
	}
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Error("Expected 1 for true")
	}
	if boolToInt(false) != 0 {
		t.Error("Expected 0 for false")
	}
}

func TestProviderMetrics_Struct(t *testing.T) {
	// Test ProviderMetrics struct initialization
	metrics := &ProviderMetrics{
		Provider:       "test",
		TotalRequests:  100,
		SuccessCount:   90,
		ErrorCount:     10,
		RateLimitCount: 5,
		AvgLatencyMs:   150.5,
		MinLatencyMs:   50,
		MaxLatencyMs:   500,
		SuccessRate:    90.0,
	}

	if metrics.Provider != "test" {
		t.Errorf("Expected provider 'test', got %s", metrics.Provider)
	}
	if metrics.TotalRequests != 100 {
		t.Errorf("Expected 100 total requests, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessRate != 90.0 {
		t.Errorf("Expected 90.0 success rate, got %f", metrics.SuccessRate)
	}
}

func TestProviderMetric_Struct(t *testing.T) {
	// Test ProviderMetric struct initialization
	now := time.Now()
	metric := ProviderMetric{
		Timestamp:   now,
		Provider:    "test",
		LatencyMs:   100,
		StatusCode:  200,
		IsError:     false,
		IsRateLimit: false,
	}

	if metric.Provider != "test" {
		t.Errorf("Expected provider 'test', got %s", metric.Provider)
	}
	if metric.LatencyMs != 100 {
		t.Errorf("Expected 100 latency, got %d", metric.LatencyMs)
	}
	if metric.StatusCode != 200 {
		t.Errorf("Expected 200 status code, got %d", metric.StatusCode)
	}
}

func TestLatencyPoint_Struct(t *testing.T) {
	// Test LatencyPoint struct initialization
	now := time.Now()
	point := LatencyPoint{
		Timestamp:    now,
		AvgLatency:   150.5,
		MinLatency:   50,
		MaxLatency:   500,
		Count:        10,
		ErrorCount:   2,
		TotalLatency: 1505,
	}

	if point.Count != 10 {
		t.Errorf("Expected 10 count, got %d", point.Count)
	}
	if point.ErrorCount != 2 {
		t.Errorf("Expected 2 error count, got %d", point.ErrorCount)
	}
	if point.AvgLatency != 150.5 {
		t.Errorf("Expected 150.5 avg latency, got %f", point.AvgLatency)
	}
}

func TestGetProviderMetrics_WithData(t *testing.T) {
	dir := t.TempDir()
	ldb, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	// Set since to before recording any metrics (use UTC to match RecordMetric)
	since := time.Now().UTC().Add(-time.Hour)

	// Record some metrics
	ldb.RecordMetric("test-provider", 100, 200, false, false)
	ldb.RecordMetric("test-provider", 150, 200, false, false)
	ldb.RecordMetric("test-provider", 200, 500, true, false)

	metrics, err := ldb.GetProviderMetrics("test-provider", since)
	if err != nil {
		t.Fatalf("GetProviderMetrics() error: %v", err)
	}

	if metrics.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessCount != 2 {
		t.Errorf("Expected 2 success count, got %d", metrics.SuccessCount)
	}
	if metrics.ErrorCount != 1 {
		t.Errorf("Expected 1 error count, got %d", metrics.ErrorCount)
	}
	if metrics.LastSuccess == nil {
		t.Error("Expected LastSuccess to be set")
	}
	if metrics.LastError == nil {
		t.Error("Expected LastError to be set")
	}
}

func TestGetAllProviderMetrics_WithData(t *testing.T) {
	dir := t.TempDir()
	ldb, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	// Set since to before recording any metrics (use UTC to match RecordMetric)
	since := time.Now().UTC().Add(-time.Hour)

	// Record metrics
	ldb.RecordMetric("provider1", 100, 200, false, false)
	ldb.RecordMetric("provider1", 150, 200, false, false)
	ldb.RecordMetric("provider2", 200, 200, false, false)
	ldb.RecordMetric("provider2", 300, 429, false, true)

	metrics, err := ldb.GetAllProviderMetrics(since)
	if err != nil {
		t.Fatalf("GetAllProviderMetrics() error: %v", err)
	}

	if len(metrics) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(metrics))
	}

	if m, ok := metrics["provider1"]; ok {
		if m.TotalRequests != 2 {
			t.Errorf("Expected 2 requests for provider1, got %d", m.TotalRequests)
		}
	} else {
		t.Error("Expected provider1 in metrics")
	}

	if m, ok := metrics["provider2"]; ok {
		if m.RateLimitCount != 1 {
			t.Errorf("Expected 1 rate limit for provider2, got %d", m.RateLimitCount)
		}
	} else {
		t.Error("Expected provider2 in metrics")
	}
}

func TestGetLatencyHistory_WithData(t *testing.T) {
	dir := t.TempDir()
	ldb, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	// Set since to before recording any metrics (use UTC to match RecordMetric)
	since := time.Now().UTC().Add(-time.Hour)

	// Record metrics
	ldb.RecordMetric("test-provider", 100, 200, false, false)
	ldb.RecordMetric("test-provider", 150, 200, false, false)
	ldb.RecordMetric("test-provider", 200, 200, false, false)

	history, err := ldb.GetLatencyHistory("test-provider", since, 5)
	if err != nil {
		t.Fatalf("GetLatencyHistory() error: %v", err)
	}

	if history == nil {
		t.Fatal("Expected non-nil history")
	}
	if len(history) == 0 {
		t.Error("Expected at least one latency point")
	}
}

func TestCleanupOldMetrics_WithData(t *testing.T) {
	dir := t.TempDir()
	ldb, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB() error: %v", err)
	}
	defer ldb.Close()

	// Record metrics
	ldb.RecordMetric("test-provider", 100, 200, false, false)
	ldb.RecordMetric("test-provider", 150, 200, false, false)

	// Cleanup with very short retention (should delete all)
	deleted, err := ldb.CleanupOldMetrics(0)
	if err != nil {
		t.Fatalf("CleanupOldMetrics() error: %v", err)
	}

	if deleted != 2 {
		t.Errorf("Expected 2 deleted, got %d", deleted)
	}
}

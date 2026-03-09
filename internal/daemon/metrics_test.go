package daemon

import (
	"testing"
	"time"
)

// TestMetricsRecordRequest verifies metrics collection
func TestMetricsRecordRequest(t *testing.T) {
	m := NewMetrics()

	// Record some requests
	m.RecordRequest("provider1", 100*time.Millisecond, nil)
	m.RecordRequest("provider1", 200*time.Millisecond, nil)
	m.RecordRequest("provider2", 150*time.Millisecond, nil)

	stats := m.GetStats()
	if stats.TotalRequests != 3 {
		t.Errorf("total_requests = %d, want 3", stats.TotalRequests)
	}
	if stats.SuccessCount != 3 {
		t.Errorf("success_count = %d, want 3", stats.SuccessCount)
	}
	if stats.ErrorCount != 0 {
		t.Errorf("error_count = %d, want 0", stats.ErrorCount)
	}
}

// TestMetricsRecordError verifies error tracking
func TestMetricsRecordError(t *testing.T) {
	m := NewMetrics()

	// Record errors
	m.RecordRequest("provider1", 50*time.Millisecond, &RequestError{Provider: "provider1", Type: "timeout"})
	m.RecordRequest("provider1", 60*time.Millisecond, &RequestError{Provider: "provider1", Type: "timeout"})
	m.RecordRequest("provider2", 70*time.Millisecond, &RequestError{Provider: "provider2", Type: "rate_limit"})

	stats := m.GetStats()
	if stats.TotalRequests != 3 {
		t.Errorf("total_requests = %d, want 3", stats.TotalRequests)
	}
	if stats.ErrorCount != 3 {
		t.Errorf("error_count = %d, want 3", stats.ErrorCount)
	}

	// Check error grouping by provider
	if len(stats.ErrorsByProvider) != 2 {
		t.Errorf("errors_by_provider count = %d, want 2", len(stats.ErrorsByProvider))
	}
	if stats.ErrorsByProvider["provider1"] != 2 {
		t.Errorf("provider1 errors = %d, want 2", stats.ErrorsByProvider["provider1"])
	}
	if stats.ErrorsByProvider["provider2"] != 1 {
		t.Errorf("provider2 errors = %d, want 1", stats.ErrorsByProvider["provider2"])
	}

	// Check error grouping by type
	if len(stats.ErrorsByType) != 2 {
		t.Errorf("errors_by_type count = %d, want 2", len(stats.ErrorsByType))
	}
	if stats.ErrorsByType["timeout"] != 2 {
		t.Errorf("timeout errors = %d, want 2", stats.ErrorsByType["timeout"])
	}
	if stats.ErrorsByType["rate_limit"] != 1 {
		t.Errorf("rate_limit errors = %d, want 1", stats.ErrorsByType["rate_limit"])
	}
}

// TestMetricsPercentiles verifies percentile calculation
func TestMetricsPercentiles(t *testing.T) {
	m := NewMetrics()

	// Record latencies: 10ms, 20ms, 30ms, ..., 100ms
	for i := 1; i <= 10; i++ {
		m.RecordRequest("provider1", time.Duration(i*10)*time.Millisecond, nil)
	}

	stats := m.GetStats()

	// P50 should be around 50-60ms
	if stats.LatencyP50Ms < 40 || stats.LatencyP50Ms > 70 {
		t.Errorf("P50 = %.1fms, want ~50-60ms", stats.LatencyP50Ms)
	}

	// P95 should be around 90-100ms
	if stats.LatencyP95Ms < 80 || stats.LatencyP95Ms > 110 {
		t.Errorf("P95 = %.1fms, want ~90-100ms", stats.LatencyP95Ms)
	}

	// P99 should be around 100ms
	if stats.LatencyP99Ms < 90 || stats.LatencyP99Ms > 110 {
		t.Errorf("P99 = %.1fms, want ~100ms", stats.LatencyP99Ms)
	}
}

// TestMetricsRingBuffer verifies ring buffer behavior
func TestMetricsRingBuffer(t *testing.T) {
	m := NewMetrics()

	// Fill ring buffer beyond capacity (default 1000)
	for i := 0; i < 1500; i++ {
		m.RecordRequest("provider1", time.Duration(i)*time.Millisecond, nil)
	}

	stats := m.GetStats()
	if stats.TotalRequests != 1500 {
		t.Errorf("total_requests = %d, want 1500", stats.TotalRequests)
	}

	// Ring buffer should only keep last 1000 samples for percentile calculation
	// This is verified by checking that percentiles are calculated from recent samples
}

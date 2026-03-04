package proxy

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// TestRequestMonitor_Add verifies buffer append and LRU eviction
func TestRequestMonitor_Add(t *testing.T) {
	t.Run("add to empty buffer", func(t *testing.T) {
		monitor := NewRequestMonitor(10)
		record := RequestRecord{
			ID:        "req1",
			Timestamp: time.Now(),
			Provider:  "test-provider",
			Model:     "claude-sonnet-4",
		}

		monitor.Add(record)

		records := monitor.GetRecent(10, RequestFilter{})
		if len(records) != 1 {
			t.Errorf("expected 1 record, got %d", len(records))
		}
		if records[0].ID != "req1" {
			t.Errorf("expected ID req1, got %s", records[0].ID)
		}
	})

	t.Run("LRU eviction when buffer full", func(t *testing.T) {
		monitor := NewRequestMonitor(3)

		// Add 5 records to a buffer of size 3
		for i := 1; i <= 5; i++ {
			record := RequestRecord{
				ID:        string(rune('0' + i)),
				Timestamp: time.Now().Add(time.Duration(i) * time.Second),
				Provider:  "test",
			}
			monitor.Add(record)
		}

		records := monitor.GetRecent(10, RequestFilter{})
		if len(records) != 3 {
			t.Errorf("expected 3 records after eviction, got %d", len(records))
		}

		// Should keep the 3 most recent (3, 4, 5)
		if records[0].ID != "5" {
			t.Errorf("expected newest record ID=5, got %s", records[0].ID)
		}
		if records[2].ID != "3" {
			t.Errorf("expected oldest kept record ID=3, got %s", records[2].ID)
		}
	})
}

// TestRequestMonitor_GetRecent verifies reverse chronological order
func TestRequestMonitor_GetRecent(t *testing.T) {
	monitor := NewRequestMonitor(100)

	// Add records with different timestamps
	now := time.Now()
	records := []RequestRecord{
		{ID: "req1", Timestamp: now.Add(-3 * time.Second), Provider: "p1"},
		{ID: "req2", Timestamp: now.Add(-2 * time.Second), Provider: "p2"},
		{ID: "req3", Timestamp: now.Add(-1 * time.Second), Provider: "p1"},
	}

	for _, r := range records {
		monitor.Add(r)
	}

	t.Run("reverse chronological order", func(t *testing.T) {
		result := monitor.GetRecent(10, RequestFilter{})
		if len(result) != 3 {
			t.Fatalf("expected 3 records, got %d", len(result))
		}

		// Newest first
		if result[0].ID != "req3" {
			t.Errorf("expected req3 first, got %s", result[0].ID)
		}
		if result[1].ID != "req2" {
			t.Errorf("expected req2 second, got %s", result[1].ID)
		}
		if result[2].ID != "req1" {
			t.Errorf("expected req1 third, got %s", result[2].ID)
		}
	})

	t.Run("limit results", func(t *testing.T) {
		result := monitor.GetRecent(2, RequestFilter{})
		if len(result) != 2 {
			t.Errorf("expected 2 records with limit=2, got %d", len(result))
		}
	})

	t.Run("filter by provider", func(t *testing.T) {
		result := monitor.GetRecent(10, RequestFilter{Provider: "p1"})
		if len(result) != 2 {
			t.Errorf("expected 2 records for provider p1, got %d", len(result))
		}
		for _, r := range result {
			if r.Provider != "p1" {
				t.Errorf("expected provider p1, got %s", r.Provider)
			}
		}
	})
}

// TestRequestMonitor_ThreadSafety verifies concurrent Add/GetRecent
func TestRequestMonitor_ThreadSafety(t *testing.T) {
	monitor := NewRequestMonitor(1000)
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				record := RequestRecord{
					ID:        string(rune('A' + id)),
					Timestamp: time.Now(),
					Provider:  "test",
				}
				monitor.Add(record)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				monitor.GetRecent(10, RequestFilter{})
			}
		}()
	}

	wg.Wait()

	// Verify no race conditions and buffer is within limits
	records := monitor.GetRecent(2000, RequestFilter{})
	if len(records) > 1000 {
		t.Errorf("buffer exceeded max size: got %d records", len(records))
	}
}

// TestRequestMonitor_FilterByStatus verifies status code filtering
func TestRequestMonitor_FilterByStatus(t *testing.T) {
	monitor := NewRequestMonitor(100)
	now := time.Now()

	// Add records with different status codes
	records := []RequestRecord{
		{ID: "req1", Timestamp: now, Provider: "p1", StatusCode: 200},
		{ID: "req2", Timestamp: now, Provider: "p1", StatusCode: 400},
		{ID: "req3", Timestamp: now, Provider: "p1", StatusCode: 500},
		{ID: "req4", Timestamp: now, Provider: "p1", StatusCode: 201},
		{ID: "req5", Timestamp: now, Provider: "p1", StatusCode: 429},
	}

	for _, r := range records {
		monitor.Add(r)
	}

	t.Run("filter by min status", func(t *testing.T) {
		result := monitor.GetRecent(10, RequestFilter{MinStatus: 400})
		if len(result) != 3 {
			t.Errorf("expected 3 records with status >= 400, got %d", len(result))
		}
		for _, r := range result {
			if r.StatusCode < 400 {
				t.Errorf("expected status >= 400, got %d", r.StatusCode)
			}
		}
	})

	t.Run("filter by max status", func(t *testing.T) {
		result := monitor.GetRecent(10, RequestFilter{MaxStatus: 299})
		if len(result) != 2 {
			t.Errorf("expected 2 records with status <= 299, got %d", len(result))
		}
		for _, r := range result {
			if r.StatusCode > 299 {
				t.Errorf("expected status <= 299, got %d", r.StatusCode)
			}
		}
	})

	t.Run("filter by status range", func(t *testing.T) {
		result := monitor.GetRecent(10, RequestFilter{MinStatus: 400, MaxStatus: 499})
		if len(result) != 2 {
			t.Errorf("expected 2 records with 400 <= status <= 499, got %d", len(result))
		}
		for _, r := range result {
			if r.StatusCode < 400 || r.StatusCode > 499 {
				t.Errorf("expected 400 <= status <= 499, got %d", r.StatusCode)
			}
		}
	})
}

// TestRequestMonitor_FilterByTimeRange verifies time range filtering
func TestRequestMonitor_FilterByTimeRange(t *testing.T) {
	monitor := NewRequestMonitor(100)
	now := time.Now()

	// Add records with different timestamps
	records := []RequestRecord{
		{ID: "req1", Timestamp: now.Add(-5 * time.Minute), Provider: "p1"},
		{ID: "req2", Timestamp: now.Add(-3 * time.Minute), Provider: "p1"},
		{ID: "req3", Timestamp: now.Add(-1 * time.Minute), Provider: "p1"},
		{ID: "req4", Timestamp: now, Provider: "p1"},
	}

	for _, r := range records {
		monitor.Add(r)
	}

	t.Run("filter by start time", func(t *testing.T) {
		startTime := now.Add(-2 * time.Minute)
		result := monitor.GetRecent(10, RequestFilter{StartTime: startTime})
		if len(result) != 2 {
			t.Errorf("expected 2 records after start time, got %d", len(result))
		}
		for _, r := range result {
			if r.Timestamp.Before(startTime) {
				t.Errorf("expected timestamp >= %v, got %v", startTime, r.Timestamp)
			}
		}
	})

	t.Run("filter by end time", func(t *testing.T) {
		endTime := now.Add(-2 * time.Minute)
		result := monitor.GetRecent(10, RequestFilter{EndTime: endTime})
		if len(result) != 2 {
			t.Errorf("expected 2 records before end time, got %d", len(result))
		}
		for _, r := range result {
			if r.Timestamp.After(endTime) {
				t.Errorf("expected timestamp <= %v, got %v", endTime, r.Timestamp)
			}
		}
	})

	t.Run("filter by time range", func(t *testing.T) {
		startTime := now.Add(-4 * time.Minute)
		endTime := now.Add(-2 * time.Minute)
		result := monitor.GetRecent(10, RequestFilter{StartTime: startTime, EndTime: endTime})
		if len(result) != 1 {
			t.Errorf("expected 1 record in time range, got %d", len(result))
		}
		if len(result) > 0 && result[0].ID != "req2" {
			t.Errorf("expected req2, got %s", result[0].ID)
		}
	})
}

// TestRequestMonitor_FilterByModel verifies model filtering
func TestRequestMonitor_FilterByModel(t *testing.T) {
	monitor := NewRequestMonitor(100)
	now := time.Now()

	// Add records with different models
	records := []RequestRecord{
		{ID: "req1", Timestamp: now, Provider: "p1", Model: "claude-sonnet-4"},
		{ID: "req2", Timestamp: now, Provider: "p1", Model: "claude-opus-4"},
		{ID: "req3", Timestamp: now, Provider: "p1", Model: "claude-sonnet-4"},
		{ID: "req4", Timestamp: now, Provider: "p1", Model: "claude-haiku-4"},
	}

	for _, r := range records {
		monitor.Add(r)
	}

	t.Run("filter by model", func(t *testing.T) {
		result := monitor.GetRecent(10, RequestFilter{Model: "claude-sonnet-4"})
		if len(result) != 2 {
			t.Errorf("expected 2 records for claude-sonnet-4, got %d", len(result))
		}
		for _, r := range result {
			if r.Model != "claude-sonnet-4" {
				t.Errorf("expected model claude-sonnet-4, got %s", r.Model)
			}
		}
	})

	t.Run("filter by non-existent model", func(t *testing.T) {
		result := monitor.GetRecent(10, RequestFilter{Model: "gpt-4"})
		if len(result) != 0 {
			t.Errorf("expected 0 records for gpt-4, got %d", len(result))
		}
	})
}

// TestRequestMonitor_FilterCombinations verifies multiple filters work together
func TestRequestMonitor_FilterCombinations(t *testing.T) {
	monitor := NewRequestMonitor(100)
	now := time.Now()

	// Add diverse records
	records := []RequestRecord{
		{ID: "req1", Timestamp: now.Add(-5 * time.Minute), Provider: "p1", Model: "claude-sonnet-4", StatusCode: 200},
		{ID: "req2", Timestamp: now.Add(-3 * time.Minute), Provider: "p2", Model: "claude-opus-4", StatusCode: 500},
		{ID: "req3", Timestamp: now.Add(-1 * time.Minute), Provider: "p1", Model: "claude-sonnet-4", StatusCode: 200},
		{ID: "req4", Timestamp: now, Provider: "p1", Model: "claude-haiku-4", StatusCode: 400},
	}

	for _, r := range records {
		monitor.Add(r)
	}

	t.Run("filter by provider and model", func(t *testing.T) {
		result := monitor.GetRecent(10, RequestFilter{Provider: "p1", Model: "claude-sonnet-4"})
		if len(result) != 2 {
			t.Errorf("expected 2 records, got %d", len(result))
		}
	})

	t.Run("filter by provider, status, and time", func(t *testing.T) {
		startTime := now.Add(-4 * time.Minute)
		result := monitor.GetRecent(10, RequestFilter{
			Provider:  "p1",
			MinStatus: 200,
			MaxStatus: 299,
			StartTime: startTime,
		})
		if len(result) != 1 {
			t.Errorf("expected 1 record, got %d", len(result))
		}
		if len(result) > 0 && result[0].ID != "req3" {
			t.Errorf("expected req3, got %s", result[0].ID)
		}
	})
}

// TestRequestRecord_DurationMs_JSON verifies that DurationMs serializes as
// actual milliseconds (int64), not nanoseconds (time.Duration default).
func TestRequestRecord_DurationMs_JSON(t *testing.T) {
	tests := []struct {
		name       string
		durationMs int64
	}{
		{"typical request", 2500},
		{"fast request", 150},
		{"slow request", 30000},
		{"zero duration", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := RequestRecord{
				ID:         "req_test",
				Timestamp:  time.Now(),
				Provider:   "test",
				Model:      "claude-sonnet-4",
				StatusCode: 200,
				DurationMs: tt.durationMs,
			}

			data, err := json.Marshal(record)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			got, ok := parsed["duration_ms"].(float64)
			if !ok {
				t.Fatalf("duration_ms not found or not a number in JSON output")
			}
			if int64(got) != tt.durationMs {
				t.Errorf("duration_ms = %v, want %d", got, tt.durationMs)
			}
		})
	}
}

// TestProviderAttempt_DurationMs_JSON verifies that ProviderAttempt.DurationMs
// serializes as actual milliseconds.
func TestProviderAttempt_DurationMs_JSON(t *testing.T) {
	attempt := ProviderAttempt{
		Provider:   "test-provider",
		StatusCode: 200,
		DurationMs: 1200,
	}

	data, err := json.Marshal(attempt)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	got, ok := parsed["duration_ms"].(float64)
	if !ok {
		t.Fatalf("duration_ms not found or not a number in JSON output")
	}
	if int64(got) != 1200 {
		t.Errorf("duration_ms = %v, want 1200", got)
	}
}

package proxy

import (
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

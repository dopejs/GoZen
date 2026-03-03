package web

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/proxy"
)

// TestGetRequests_Success verifies API returns JSON with correct structure
func TestGetRequests_Success(t *testing.T) {
	s := setupTestServer(t)

	// Add some test records to the monitor
	monitor := proxy.GetGlobalRequestMonitor()
	now := time.Now()

	records := []proxy.RequestRecord{
		{
			ID:           "req1",
			Timestamp:    now.Add(-2 * time.Second),
			Provider:     "test-provider",
			Model:        "claude-sonnet-4",
			StatusCode:   200,
			Duration:     time.Second,
			InputTokens:  100,
			OutputTokens: 50,
			Cost:         0.005,
		},
		{
			ID:           "req2",
			Timestamp:    now.Add(-1 * time.Second),
			Provider:     "another-provider",
			Model:        "claude-haiku-3-5",
			StatusCode:   200,
			Duration:     500 * time.Millisecond,
			InputTokens:  50,
			OutputTokens: 25,
			Cost:         0.001,
		},
	}

	for _, r := range records {
		monitor.Add(r)
	}

	// Make request
	w := doRequest(s, http.MethodGet, "/api/v1/monitoring/requests", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	// Parse response
	var resp struct {
		Requests []proxy.RequestRecord `json:"requests"`
		Total    int                   `json:"total"`
		Limit    int                   `json:"limit"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify structure
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}

	if len(resp.Requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(resp.Requests))
	}

	// Verify reverse chronological order (newest first)
	if resp.Requests[0].ID != "req2" {
		t.Errorf("first request ID = %s, want req2", resp.Requests[0].ID)
	}

	// Verify fields are populated
	req1 := resp.Requests[1]
	if req1.Provider != "test-provider" {
		t.Errorf("provider = %s, want test-provider", req1.Provider)
	}
	if req1.Model != "claude-sonnet-4" {
		t.Errorf("model = %s, want claude-sonnet-4", req1.Model)
	}
	if req1.InputTokens != 100 {
		t.Errorf("input_tokens = %d, want 100", req1.InputTokens)
	}
}

// TestGetRequests_WithFilters verifies provider/status/time filters work
func TestGetRequests_WithFilters(t *testing.T) {
	s := setupTestServer(t)
	monitor := proxy.GetGlobalRequestMonitor()
	now := time.Now()

	// Add test records with different providers and status codes
	records := []proxy.RequestRecord{
		{ID: "req1", Timestamp: now.Add(-3 * time.Second), Provider: "provider-a", StatusCode: 200},
		{ID: "req2", Timestamp: now.Add(-2 * time.Second), Provider: "provider-b", StatusCode: 500},
		{ID: "req3", Timestamp: now.Add(-1 * time.Second), Provider: "provider-a", StatusCode: 200},
	}

	for _, r := range records {
		monitor.Add(r)
	}

	t.Run("filter by provider", func(t *testing.T) {
		w := doRequest(s, http.MethodGet, "/api/v1/monitoring/requests?provider=provider-a", nil)

		var resp struct {
			Requests []proxy.RequestRecord `json:"requests"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if len(resp.Requests) != 2 {
			t.Errorf("expected 2 requests for provider-a, got %d", len(resp.Requests))
		}

		for _, r := range resp.Requests {
			if r.Provider != "provider-a" {
				t.Errorf("expected provider-a, got %s", r.Provider)
			}
		}
	})

	t.Run("filter by status code range", func(t *testing.T) {
		w := doRequest(s, http.MethodGet, "/api/v1/monitoring/requests?status_min=500&status_max=599", nil)

		var resp struct {
			Requests []proxy.RequestRecord `json:"requests"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if len(resp.Requests) != 1 {
			t.Errorf("expected 1 error request, got %d", len(resp.Requests))
		}

		if resp.Requests[0].StatusCode != 500 {
			t.Errorf("expected status 500, got %d", resp.Requests[0].StatusCode)
		}
	})

	t.Run("limit results", func(t *testing.T) {
		w := doRequest(s, http.MethodGet, "/api/v1/monitoring/requests?limit=1", nil)

		var resp struct {
			Requests []proxy.RequestRecord `json:"requests"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if len(resp.Requests) != 1 {
			t.Errorf("expected 1 request with limit=1, got %d", len(resp.Requests))
		}
	})
}

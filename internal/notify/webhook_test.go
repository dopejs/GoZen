package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

func TestNewWebhookDispatcher(t *testing.T) {
	d := NewWebhookDispatcher()
	if d == nil {
		t.Fatal("NewWebhookDispatcher returned nil")
	}
	if d.client == nil {
		t.Error("client is nil")
	}
	if d.client.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", d.client.Timeout)
	}
}

func TestMatchesEvent(t *testing.T) {
	d := NewWebhookDispatcher()

	tests := []struct {
		name     string
		wh       *config.WebhookConfig
		event    config.WebhookEvent
		expected bool
	}{
		{
			name: "matches single event",
			wh: &config.WebhookConfig{
				Events: []config.WebhookEvent{config.WebhookEventBudgetWarning},
			},
			event:    config.WebhookEventBudgetWarning,
			expected: true,
		},
		{
			name: "matches one of multiple events",
			wh: &config.WebhookConfig{
				Events: []config.WebhookEvent{
					config.WebhookEventBudgetWarning,
					config.WebhookEventProviderDown,
				},
			},
			event:    config.WebhookEventProviderDown,
			expected: true,
		},
		{
			name: "does not match",
			wh: &config.WebhookConfig{
				Events: []config.WebhookEvent{config.WebhookEventBudgetWarning},
			},
			event:    config.WebhookEventProviderDown,
			expected: false,
		},
		{
			name: "empty events list",
			wh: &config.WebhookConfig{
				Events: []config.WebhookEvent{},
			},
			event:    config.WebhookEventBudgetWarning,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.matchesEvent(tt.wh, tt.event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFormatSlack(t *testing.T) {
	d := NewWebhookDispatcher()
	payload := WebhookPayload{
		Event:     config.WebhookEventBudgetWarning,
		Timestamp: time.Now().UTC(),
		Data: &BudgetEventData{
			Period:  "daily",
			Spent:   8.5,
			Limit:   10.0,
			Percent: 85.0,
		},
	}

	body := d.formatSlack(payload)
	if len(body) == 0 {
		t.Fatal("formatSlack returned empty body")
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(body, &msg); err != nil {
		t.Fatalf("failed to unmarshal Slack message: %v", err)
	}

	if _, ok := msg["text"]; !ok {
		t.Error("Slack message missing 'text' field")
	}
	if _, ok := msg["blocks"]; !ok {
		t.Error("Slack message missing 'blocks' field")
	}
}

func TestFormatDiscord(t *testing.T) {
	d := NewWebhookDispatcher()
	payload := WebhookPayload{
		Event:     config.WebhookEventProviderDown,
		Timestamp: time.Now().UTC(),
		Data: &ProviderEventData{
			Provider: "anthropic",
			Status:   "unhealthy",
			Error:    "connection timeout",
		},
	}

	body := d.formatDiscord(payload)
	if len(body) == 0 {
		t.Fatal("formatDiscord returned empty body")
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(body, &msg); err != nil {
		t.Fatalf("failed to unmarshal Discord message: %v", err)
	}

	if _, ok := msg["content"]; !ok {
		t.Error("Discord message missing 'content' field")
	}
	if _, ok := msg["embeds"]; !ok {
		t.Error("Discord message missing 'embeds' field")
	}
}

func TestFormatGeneric(t *testing.T) {
	d := NewWebhookDispatcher()
	payload := WebhookPayload{
		Event:     config.WebhookEventFailover,
		Timestamp: time.Now().UTC(),
		Data: &FailoverEventData{
			FromProvider: "anthropic",
			ToProvider:   "openai",
			Reason:       "rate limit",
		},
	}

	body := d.formatGeneric(payload)
	if len(body) == 0 {
		t.Fatal("formatGeneric returned empty body")
	}

	var result WebhookPayload
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal generic payload: %v", err)
	}

	if result.Event != payload.Event {
		t.Errorf("expected event %s, got %s", payload.Event, result.Event)
	}
}

func TestFormatMessage(t *testing.T) {
	d := NewWebhookDispatcher()

	tests := []struct {
		name     string
		payload  WebhookPayload
		contains string
	}{
		{
			name: "budget warning",
			payload: WebhookPayload{
				Event: config.WebhookEventBudgetWarning,
				Data: &BudgetEventData{
					Period:  "daily",
					Spent:   8.5,
					Limit:   10.0,
					Percent: 85.0,
				},
			},
			contains: "Budget Warning",
		},
		{
			name: "budget exceeded",
			payload: WebhookPayload{
				Event: config.WebhookEventBudgetExceeded,
				Data: &BudgetEventData{
					Period: "monthly",
					Spent:  105.0,
					Limit:  100.0,
					Action: "block",
				},
			},
			contains: "Budget Exceeded",
		},
		{
			name: "provider down",
			payload: WebhookPayload{
				Event: config.WebhookEventProviderDown,
				Data: &ProviderEventData{
					Provider: "anthropic",
					Error:    "timeout",
				},
			},
			contains: "Provider Down",
		},
		{
			name: "provider up",
			payload: WebhookPayload{
				Event: config.WebhookEventProviderUp,
				Data: &ProviderEventData{
					Provider:  "anthropic",
					LatencyMs: 150,
				},
			},
			contains: "Provider Up",
		},
		{
			name: "failover",
			payload: WebhookPayload{
				Event: config.WebhookEventFailover,
				Data: &FailoverEventData{
					FromProvider: "anthropic",
					ToProvider:   "openai",
					Reason:       "rate limit",
				},
			},
			contains: "Failover",
		},
		{
			name: "daily summary",
			payload: WebhookPayload{
				Event: config.WebhookEventDailySummary,
				Data: &DailySummaryData{
					Date:          "2026-03-05",
					TotalCost:     25.50,
					TotalRequests: 100,
					TotalInput:    50000,
					TotalOutput:   10000,
				},
			},
			contains: "Daily Summary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := d.formatMessage(tt.payload)
			if msg == "" {
				t.Error("formatMessage returned empty string")
			}
			if len(tt.contains) > 0 && !contains(msg, tt.contains) {
				t.Errorf("expected message to contain %q, got %q", tt.contains, msg)
			}
		})
	}
}

func TestGetColorForEvent(t *testing.T) {
	d := NewWebhookDispatcher()

	tests := []struct {
		event config.WebhookEvent
		color int
	}{
		{config.WebhookEventBudgetWarning, 0xFBBF24},
		{config.WebhookEventBudgetExceeded, 0xFB7185},
		{config.WebhookEventProviderDown, 0xFB7185},
		{config.WebhookEventProviderUp, 0x86EFAC},
		{config.WebhookEventFailover, 0xC4B5FD},
		{config.WebhookEventDailySummary, 0x5EEAD4},
		{"unknown", 0x93C5FD},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			color := d.getColorForEvent(tt.event)
			if color != tt.color {
				t.Errorf("expected color 0x%X, got 0x%X", tt.color, color)
			}
		})
	}
}

func TestDispatch(t *testing.T) {
	// Create a test HTTP server
	var receivedCount int
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create dispatcher with test webhooks
	d := &WebhookDispatcher{
		webhooks: []*config.WebhookConfig{
			{
				Enabled: true,
				URL:     server.URL,
				Events:  []config.WebhookEvent{config.WebhookEventBudgetWarning},
			},
			{
				Enabled: false, // disabled, should not receive
				URL:     server.URL,
				Events:  []config.WebhookEvent{config.WebhookEventBudgetWarning},
			},
			{
				Enabled: true,
				URL:     server.URL,
				Events:  []config.WebhookEvent{config.WebhookEventProviderDown}, // different event
			},
		},
		client: &http.Client{Timeout: 5 * time.Second},
	}

	// Dispatch event
	d.Dispatch(config.WebhookEventBudgetWarning, &BudgetEventData{
		Period:  "daily",
		Spent:   8.5,
		Limit:   10.0,
		Percent: 85.0,
	})

	// Wait for async dispatch
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := receivedCount
	mu.Unlock()

	// Should only receive 1 request (first webhook matches, second is disabled, third doesn't match event)
	if count != 1 {
		t.Errorf("expected 1 webhook call, got %d", count)
	}
}

func TestTestWebhook(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "GoZen-Webhook/1.0" {
			t.Error("missing or incorrect User-Agent header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	d := NewWebhookDispatcher()

	tests := []struct {
		name      string
		wh        *config.WebhookConfig
		expectErr bool
	}{
		{
			name: "successful test",
			wh: &config.WebhookConfig{
				URL: server.URL,
			},
			expectErr: false,
		},
		{
			name: "invalid URL",
			wh: &config.WebhookConfig{
				URL: "http://invalid-url-that-does-not-exist-12345.com",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.TestWebhook(tt.wh)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestReloadConfig(t *testing.T) {
	d := NewWebhookDispatcher()
	initialCount := len(d.webhooks)

	// ReloadConfig should not panic
	d.ReloadConfig()

	// Should still have webhooks (or empty list)
	if d.webhooks == nil {
		t.Error("webhooks is nil after reload")
	}

	// Count may change depending on config, just verify it doesn't panic
	_ = initialCount
}

func TestGetGlobalDispatcher(t *testing.T) {
	d1 := GetGlobalDispatcher()
	d2 := GetGlobalDispatcher()

	if d1 == nil {
		t.Fatal("GetGlobalDispatcher returned nil")
	}
	if d1 != d2 {
		t.Error("GetGlobalDispatcher should return the same instance")
	}
}

func TestHelperFunctions(t *testing.T) {
	// These functions should not panic
	t.Run("NotifyBudgetWarning", func(t *testing.T) {
		NotifyBudgetWarning("daily", 8.5, 10.0, 85.0, "test-project")
	})

	t.Run("NotifyBudgetExceeded", func(t *testing.T) {
		NotifyBudgetExceeded("monthly", 105.0, 100.0, "block", "test-project")
	})

	t.Run("NotifyProviderDown", func(t *testing.T) {
		NotifyProviderDown("anthropic", "connection timeout")
	})

	t.Run("NotifyProviderUp", func(t *testing.T) {
		NotifyProviderUp("anthropic", 150)
	})

	t.Run("NotifyFailover", func(t *testing.T) {
		NotifyFailover("anthropic", "openai", "rate limit", "session-123")
	})

	t.Run("NotifyDailySummary", func(t *testing.T) {
		NotifyDailySummary("2026-03-05", 25.50, 100, 50000, 10000, map[string]float64{
			"anthropic": 15.0,
			"openai":    10.5,
		})
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

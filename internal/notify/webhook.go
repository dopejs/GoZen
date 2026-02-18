package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// WebhookPayload is the generic payload sent to webhooks.
type WebhookPayload struct {
	Event     config.WebhookEvent `json:"event"`
	Timestamp time.Time           `json:"timestamp"`
	Data      interface{}         `json:"data"`
}

// BudgetEventData contains data for budget-related events.
type BudgetEventData struct {
	Period     string  `json:"period"`
	Spent      float64 `json:"spent"`
	Limit      float64 `json:"limit"`
	Percent    float64 `json:"percent"`
	Action     string  `json:"action,omitempty"`
	Project    string  `json:"project,omitempty"`
}

// ProviderEventData contains data for provider-related events.
type ProviderEventData struct {
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	LatencyMs int    `json:"latency_ms,omitempty"`
}

// FailoverEventData contains data for failover events.
type FailoverEventData struct {
	FromProvider string `json:"from_provider"`
	ToProvider   string `json:"to_provider"`
	Reason       string `json:"reason"`
	SessionID    string `json:"session_id,omitempty"`
}

// DailySummaryData contains data for daily summary events.
type DailySummaryData struct {
	Date          string             `json:"date"`
	TotalCost     float64            `json:"total_cost"`
	TotalRequests int                `json:"total_requests"`
	TotalInput    int                `json:"total_input_tokens"`
	TotalOutput   int                `json:"total_output_tokens"`
	ByProvider    map[string]float64 `json:"by_provider,omitempty"`
}

// WebhookDispatcher sends notifications to configured webhooks.
type WebhookDispatcher struct {
	mu       sync.RWMutex
	webhooks []*config.WebhookConfig
	client   *http.Client
}

// NewWebhookDispatcher creates a new webhook dispatcher.
func NewWebhookDispatcher() *WebhookDispatcher {
	return &WebhookDispatcher{
		webhooks: config.GetWebhooks(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ReloadConfig refreshes the webhook configuration.
func (d *WebhookDispatcher) ReloadConfig() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.webhooks = config.GetWebhooks()
}

// Dispatch sends an event to all matching webhooks.
func (d *WebhookDispatcher) Dispatch(event config.WebhookEvent, data interface{}) {
	d.mu.RLock()
	webhooks := d.webhooks
	d.mu.RUnlock()

	payload := WebhookPayload{
		Event:     event,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}

	for _, wh := range webhooks {
		if !wh.Enabled {
			continue
		}
		if !d.matchesEvent(wh, event) {
			continue
		}

		go d.send(wh, payload)
	}
}

func (d *WebhookDispatcher) matchesEvent(wh *config.WebhookConfig, event config.WebhookEvent) bool {
	for _, e := range wh.Events {
		if e == event {
			return true
		}
	}
	return false
}

func (d *WebhookDispatcher) send(wh *config.WebhookConfig, payload WebhookPayload) {
	var body []byte
	var contentType string

	// Detect webhook type from URL and format accordingly
	if strings.Contains(wh.URL, "slack.com") {
		body = d.formatSlack(payload)
		contentType = "application/json"
	} else if strings.Contains(wh.URL, "discord.com") {
		body = d.formatDiscord(payload)
		contentType = "application/json"
	} else {
		body = d.formatGeneric(payload)
		contentType = "application/json"
	}

	req, err := http.NewRequest(http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "GoZen-Webhook/1.0")

	// Add custom headers
	for k, v := range wh.Headers {
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// formatSlack formats the payload for Slack webhooks.
func (d *WebhookDispatcher) formatSlack(payload WebhookPayload) []byte {
	text := d.formatMessage(payload)

	// Slack message format
	msg := map[string]interface{}{
		"text": text,
		"blocks": []map[string]interface{}{
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": text,
				},
			},
		},
	}

	data, _ := json.Marshal(msg)
	return data
}

// formatDiscord formats the payload for Discord webhooks.
func (d *WebhookDispatcher) formatDiscord(payload WebhookPayload) []byte {
	text := d.formatMessage(payload)

	// Discord message format
	msg := map[string]interface{}{
		"content": text,
		"embeds": []map[string]interface{}{
			{
				"title":       string(payload.Event),
				"description": text,
				"timestamp":   payload.Timestamp.Format(time.RFC3339),
				"color":       d.getColorForEvent(payload.Event),
			},
		},
	}

	data, _ := json.Marshal(msg)
	return data
}

// formatGeneric formats the payload as generic JSON.
func (d *WebhookDispatcher) formatGeneric(payload WebhookPayload) []byte {
	data, _ := json.Marshal(payload)
	return data
}

// formatMessage creates a human-readable message for the event.
func (d *WebhookDispatcher) formatMessage(payload WebhookPayload) string {
	switch payload.Event {
	case config.WebhookEventBudgetWarning:
		if data, ok := payload.Data.(*BudgetEventData); ok {
			return fmt.Sprintf("⚠️ Budget Warning: %s budget at %.1f%% ($%.2f / $%.2f)",
				data.Period, data.Percent, data.Spent, data.Limit)
		}

	case config.WebhookEventBudgetExceeded:
		if data, ok := payload.Data.(*BudgetEventData); ok {
			return fmt.Sprintf("🚫 Budget Exceeded: %s limit of $%.2f reached (spent: $%.2f). Action: %s",
				data.Period, data.Limit, data.Spent, data.Action)
		}

	case config.WebhookEventProviderDown:
		if data, ok := payload.Data.(*ProviderEventData); ok {
			return fmt.Sprintf("🔴 Provider Down: %s is unhealthy. Error: %s",
				data.Provider, data.Error)
		}

	case config.WebhookEventProviderUp:
		if data, ok := payload.Data.(*ProviderEventData); ok {
			return fmt.Sprintf("🟢 Provider Up: %s is healthy again (latency: %dms)",
				data.Provider, data.LatencyMs)
		}

	case config.WebhookEventFailover:
		if data, ok := payload.Data.(*FailoverEventData); ok {
			return fmt.Sprintf("🔄 Failover: Switched from %s to %s. Reason: %s",
				data.FromProvider, data.ToProvider, data.Reason)
		}

	case config.WebhookEventDailySummary:
		if data, ok := payload.Data.(*DailySummaryData); ok {
			return fmt.Sprintf("📊 Daily Summary (%s): %d requests, $%.2f total cost, %d input / %d output tokens",
				data.Date, data.TotalRequests, data.TotalCost, data.TotalInput, data.TotalOutput)
		}
	}

	// Fallback
	return fmt.Sprintf("GoZen Event: %s", payload.Event)
}

// getColorForEvent returns a Discord embed color for the event type.
func (d *WebhookDispatcher) getColorForEvent(event config.WebhookEvent) int {
	switch event {
	case config.WebhookEventBudgetWarning:
		return 0xFBBF24 // Amber
	case config.WebhookEventBudgetExceeded:
		return 0xFB7185 // Red
	case config.WebhookEventProviderDown:
		return 0xFB7185 // Red
	case config.WebhookEventProviderUp:
		return 0x86EFAC // Sage/Green
	case config.WebhookEventFailover:
		return 0xC4B5FD // Lavender
	case config.WebhookEventDailySummary:
		return 0x5EEAD4 // Teal
	default:
		return 0x93C5FD // Blue
	}
}

// TestWebhook sends a test message to a webhook.
func (d *WebhookDispatcher) TestWebhook(wh *config.WebhookConfig) error {
	payload := WebhookPayload{
		Event:     "test",
		Timestamp: time.Now().UTC(),
		Data: map[string]string{
			"message": "This is a test notification from GoZen",
		},
	}

	var body []byte
	if strings.Contains(wh.URL, "slack.com") {
		body = d.formatSlack(payload)
	} else if strings.Contains(wh.URL, "discord.com") {
		body = d.formatDiscord(payload)
	} else {
		body = d.formatGeneric(payload)
	}

	req, err := http.NewRequest(http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GoZen-Webhook/1.0")

	for k, v := range wh.Headers {
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// --- Global dispatcher ---

var globalDispatcher *WebhookDispatcher
var dispatcherOnce sync.Once

// GetGlobalDispatcher returns the global webhook dispatcher.
func GetGlobalDispatcher() *WebhookDispatcher {
	dispatcherOnce.Do(func() {
		globalDispatcher = NewWebhookDispatcher()
	})
	return globalDispatcher
}

// DispatchEvent is a convenience function to dispatch an event globally.
func DispatchEvent(event config.WebhookEvent, data interface{}) {
	GetGlobalDispatcher().Dispatch(event, data)
}

// --- Event helper functions ---

// NotifyBudgetWarning sends a budget warning notification.
func NotifyBudgetWarning(period string, spent, limit, percent float64, project string) {
	DispatchEvent(config.WebhookEventBudgetWarning, &BudgetEventData{
		Period:  period,
		Spent:   spent,
		Limit:   limit,
		Percent: percent,
		Project: project,
	})
}

// NotifyBudgetExceeded sends a budget exceeded notification.
func NotifyBudgetExceeded(period string, spent, limit float64, action, project string) {
	DispatchEvent(config.WebhookEventBudgetExceeded, &BudgetEventData{
		Period:  period,
		Spent:   spent,
		Limit:   limit,
		Percent: 100,
		Action:  action,
		Project: project,
	})
}

// NotifyProviderDown sends a provider down notification.
func NotifyProviderDown(provider, errorMsg string) {
	DispatchEvent(config.WebhookEventProviderDown, &ProviderEventData{
		Provider: provider,
		Status:   "unhealthy",
		Error:    errorMsg,
	})
}

// NotifyProviderUp sends a provider up notification.
func NotifyProviderUp(provider string, latencyMs int) {
	DispatchEvent(config.WebhookEventProviderUp, &ProviderEventData{
		Provider:  provider,
		Status:    "healthy",
		LatencyMs: latencyMs,
	})
}

// NotifyFailover sends a failover notification.
func NotifyFailover(from, to, reason, sessionID string) {
	DispatchEvent(config.WebhookEventFailover, &FailoverEventData{
		FromProvider: from,
		ToProvider:   to,
		Reason:       reason,
		SessionID:    sessionID,
	})
}

// NotifyDailySummary sends a daily summary notification.
func NotifyDailySummary(date string, cost float64, requests, input, output int, byProvider map[string]float64) {
	DispatchEvent(config.WebhookEventDailySummary, &DailySummaryData{
		Date:          date,
		TotalCost:     cost,
		TotalRequests: requests,
		TotalInput:    input,
		TotalOutput:   output,
		ByProvider:    byProvider,
	})
}

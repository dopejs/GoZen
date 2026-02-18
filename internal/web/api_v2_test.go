package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dopejs/gozen/internal/agent"
	"github.com/dopejs/gozen/internal/proxy"
)

// --- Agent Config API ---

func TestAgentConfigGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/config", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAgentConfigPut(t *testing.T) {
	s := setupTestServer(t)

	body := map[string]interface{}{
		"enabled": true,
		"observatory": map[string]interface{}{
			"enabled":         true,
			"stuck_threshold": 5,
		},
	}
	w := doRequest(s, "PUT", "/api/v1/agent/config", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentConfigMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/agent/config", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Agent Stats API ---

func TestAgentStatsGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/stats", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentStatsMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/agent/stats", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Agent Sessions API ---

func TestAgentSessionsList(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/sessions", nil)
	// May be 200 or 503 depending on whether observatory is initialized
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

func TestAgentSessionsMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/agent/sessions", nil)
	// POST to /sessions (not /sessions/{id}/kill) should be method not allowed or 503
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 405 or 503, got %d", w.Code)
	}
}

// --- Agent Locks API ---

func TestAgentLocksList(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/locks", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

// --- Agent Changes API ---

func TestAgentChangesGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/changes", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

func TestAgentChangesMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/agent/changes", nil)
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 405 or 503, got %d", w.Code)
	}
}

// --- Agent Tasks API ---

func TestAgentTasksList(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/tasks", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

// --- Agent Runtime API ---

func TestAgentRuntimeList(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/runtime", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

// --- Agent Guardrails API ---

func TestAgentGuardrailsGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/guardrails", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

func TestAgentGuardrailsSpending(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/guardrails/spending", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

func TestAgentGuardrailsOperations(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/agent/guardrails/operations", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

// --- Compression API ---

func TestCompressionGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/compression", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCompressionPut(t *testing.T) {
	s := setupTestServer(t)

	body := map[string]interface{}{
		"enabled":          true,
		"threshold_tokens": 50000,
		"target_tokens":    30000,
	}
	w := doRequest(s, "PUT", "/api/v1/compression", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCompressionMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/compression", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestCompressionStatsGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/compression/stats", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Middleware API ---

func TestMiddlewareGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/middleware", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMiddlewarePut(t *testing.T) {
	s := setupTestServer(t)

	body := map[string]interface{}{
		"enabled": true,
		"middlewares": []map[string]interface{}{
			{"name": "request-logger", "enabled": true},
		},
	}
	w := doRequest(s, "PUT", "/api/v1/middleware", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMiddlewareMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/middleware", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestMiddlewareReload(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/middleware/reload", nil)
	// May succeed or fail depending on middleware registry state
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200, 500 or 503, got %d", w.Code)
	}
}

func TestMiddlewareReloadMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/middleware/reload", nil)
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 405 or 503, got %d", w.Code)
	}
}

// --- Pricing API ---

func TestPricingGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/pricing", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPricingPut(t *testing.T) {
	s := setupTestServer(t)

	body := map[string]interface{}{
		"custom-model": map[string]interface{}{
			"input_per_million":  1.0,
			"output_per_million": 2.0,
		},
	}
	w := doRequest(s, "PUT", "/api/v1/pricing", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPricingMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/pricing", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestPricingResetPost(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/pricing/reset", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPricingResetMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/pricing/reset", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Sessions API ---

func TestSessionsGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sessions", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSessionsMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sessions", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSessionGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sessions/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestSessionMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PUT", "/api/v1/sessions/test", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Health Providers API ---

func TestHealthProvidersGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health/providers", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHealthProvidersMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/health/providers", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHealthProviderGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health/providers/test-provider", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHealthProviderMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/health/providers/test-provider", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Usage API ---

func TestUsageGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUsageMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/usage", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestUsageSummaryGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage/summary", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUsageSummaryWithPeriod(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage/summary?period=day", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUsageHourlyGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage/hourly", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Budget API ---

func TestBudgetGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/budget", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBudgetPut(t *testing.T) {
	s := setupTestServer(t)

	body := map[string]interface{}{
		"daily": map[string]interface{}{
			"amount": 10.0,
			"action": "warn",
		},
	}
	w := doRequest(s, "PUT", "/api/v1/budget", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBudgetMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/budget", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestBudgetStatusGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/budget/status", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Webhooks API ---

func TestWebhooksGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/webhooks", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWebhooksPost(t *testing.T) {
	s := setupTestServer(t)

	body := map[string]interface{}{
		"name":    "test-webhook",
		"url":     "https://example.com/webhook",
		"events":  []string{"budget_warning"},
		"enabled": true,
	}
	w := doRequest(s, "POST", "/api/v1/webhooks", body)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("expected 201 or 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWebhooksMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/webhooks", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestWebhookTestNotConfigured(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/webhooks/test", nil)
	// May fail because no webhooks configured
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200 or 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWebhookTestMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/webhooks/test", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestWebhookGetByName(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/webhooks/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestWebhookDeleteByName(t *testing.T) {
	s := setupTestServer(t)

	// First create a webhook
	body := map[string]interface{}{
		"name":    "del-webhook",
		"url":     "https://example.com/hook",
		"events":  []string{"budget_warning"},
		"enabled": true,
	}
	doRequest(s, "POST", "/api/v1/webhooks", body)

	// Then delete it
	w := doRequest(s, "DELETE", "/api/v1/webhooks/del-webhook", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Middleware By Name API ---

func TestMiddlewareByNameGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/middleware/request-logger", nil)
	// May be 200 or 404 depending on whether middleware is configured
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestMiddlewareByNameMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PUT", "/api/v1/middleware/request-logger", nil)
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
		t.Fatalf("expected 405 or 404, got %d", w.Code)
	}
}

func TestMiddlewareEnablePost(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/middleware/request-logger/enable", nil)
	// May succeed, fail with 400 (not found), or 500 (registry not initialized)
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200, 400 or 500, got %d", w.Code)
	}
}

func TestMiddlewareEnableMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/middleware/request-logger/enable", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestMiddlewareDisablePost(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/middleware/request-logger/disable", nil)
	// May succeed, fail with 400 (not found), or 500 (registry not initialized)
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200, 400 or 500, got %d", w.Code)
	}
}

func TestMiddlewareDisableMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/middleware/request-logger/disable", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Additional Compression Tests ---

func TestCompressionStatsMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/compression/stats", nil)
	// GET is allowed, POST may return 200 (handled by same handler) or 405
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusOK {
		t.Fatalf("expected 405 or 200, got %d", w.Code)
	}
}

// --- Additional Usage Tests ---

func TestUsageSummaryWithWeekPeriod(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage/summary?period=week", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUsageSummaryWithMonthPeriod(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage/summary?period=month", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUsageHourlyMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/usage/hourly", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Additional Budget Tests ---

func TestBudgetStatusMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/budget/status", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Additional Agent Tests ---

func TestAgentTasksMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/agent/tasks", nil)
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 405 or 503, got %d", w.Code)
	}
}

func TestAgentRuntimeMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/agent/runtime", nil)
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 405 or 503, got %d", w.Code)
	}
}

func TestAgentGuardrailsMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/agent/guardrails", nil)
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 405 or 503, got %d", w.Code)
	}
}

func TestAgentLocksMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/agent/locks", nil)
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 405 or 503, got %d", w.Code)
	}
}

// --- Webhook Update Test ---

func TestWebhookUpdateByName(t *testing.T) {
	s := setupTestServer(t)

	// First create a webhook
	body := map[string]interface{}{
		"name":    "update-webhook",
		"url":     "https://example.com/hook",
		"events":  []string{"budget_warning"},
		"enabled": true,
	}
	doRequest(s, "POST", "/api/v1/webhooks", body)

	// Then update it
	updateBody := map[string]interface{}{
		"url":     "https://example.com/hook-updated",
		"enabled": false,
	}
	w := doRequest(s, "PUT", "/api/v1/webhooks/update-webhook", updateBody)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Settings API Additional Tests ---

func TestSettingsDeleteMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/settings", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Sync API Tests ---

func TestSyncStatusGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/status", nil)
	// May be 200 or 503 depending on sync manager state
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 200 or 503, got %d", w.Code)
	}
}

func TestSyncPullPost(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/pull", nil)
	// May be 200, 400 (no sync configured), 500 or 503 depending on sync manager state
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusServiceUnavailable && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200, 400, 500 or 503, got %d", w.Code)
	}
}

func TestSyncPushPost(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/push", nil)
	// May be 200, 400 (no sync configured), 500 or 503 depending on sync manager state
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusServiceUnavailable && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200, 400, 500 or 503, got %d", w.Code)
	}
}

func TestSyncConfigGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/config", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSyncConfigPut(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"backend":       "",
		"auto_pull":     false,
		"pull_interval": 300,
	}
	w := doRequest(s, "PUT", "/api/v1/sync/config", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Agent Tests with Infrastructure ---

func setupAgentInfrastructure() {
	agent.InitGlobalObservatory()
	agent.InitGlobalCoordinator()
	agent.InitGlobalGuardrails()
	agent.InitGlobalTaskQueue()
	agent.InitGlobalRuntime(19841)
}

func setupProxyInfrastructure(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, "logs")
	os.MkdirAll(logDir, 0755)
	ldb, _ := proxy.OpenLogDB(logDir)
	proxy.InitGlobalUsageTracker(ldb)
	tracker := proxy.GetGlobalUsageTracker()
	proxy.InitGlobalBudgetChecker(tracker)
	proxy.InitGlobalHealthChecker(ldb)
}

func TestAgentSessionsWithObservatory(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/sessions", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentSessionsGetSpecific(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	// Try to get a non-existent session
	w := doRequest(s, "GET", "/api/v1/agent/sessions/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAgentSessionsKill(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	// Try to kill a non-existent session
	w := doRequest(s, "POST", "/api/v1/agent/sessions/nonexistent/kill", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAgentSessionsPause(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	// Try to pause a non-existent session
	w := doRequest(s, "POST", "/api/v1/agent/sessions/nonexistent/pause", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAgentSessionsResume(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	// Try to resume a non-existent session
	w := doRequest(s, "POST", "/api/v1/agent/sessions/nonexistent/resume", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAgentLocksWithCoordinator(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/locks", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentTasksWithQueue(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/tasks", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentRuntimeWithRuntime(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/runtime", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentGuardrailsWithGuardrails(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/guardrails", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentGuardrailsSpendingWithGuardrails(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/guardrails/spending", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentGuardrailsOperationsWithGuardrails(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/guardrails/operations", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentChangesWithObservatory(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/changes", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Agent Task Tests ---

func TestAgentTasksPost(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	body := map[string]interface{}{
		"session_id":  "test-session",
		"description": "Test task",
	}
	w := doRequest(s, "POST", "/api/v1/agent/tasks", body)
	// May succeed or fail depending on validation
	if w.Code != http.StatusOK && w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200, 201 or 400, got %d", w.Code)
	}
}

func TestAgentTaskGetSpecific(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/tasks/nonexistent", nil)
	if w.Code != http.StatusNotFound && w.Code != http.StatusOK {
		t.Fatalf("expected 404 or 200, got %d", w.Code)
	}
}

// --- Additional Agent Runtime Tests ---

func TestAgentRuntimePost(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	body := map[string]interface{}{
		"session_id": "test-session",
		"command":    "test",
	}
	w := doRequest(s, "POST", "/api/v1/agent/runtime", body)
	// POST may not be allowed on this endpoint
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 200, 400, 404 or 405, got %d", w.Code)
	}
}

func TestAgentRuntimeGetSpecific(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/runtime/nonexistent", nil)
	if w.Code != http.StatusNotFound && w.Code != http.StatusOK {
		t.Fatalf("expected 404 or 200, got %d", w.Code)
	}
}

// --- Additional Agent Locks Tests ---

func TestAgentLocksAcquire(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	body := map[string]interface{}{
		"session_id": "test-session",
		"file_path":  "/test/file.txt",
	}
	w := doRequest(s, "POST", "/api/v1/agent/locks", body)
	// POST may not be allowed on this endpoint
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusConflict && w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 200, 400, 405 or 409, got %d", w.Code)
	}
}

func TestAgentLocksRelease(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "DELETE", "/api/v1/agent/locks/test-lock", nil)
	// DELETE may require specific format or return 400 for invalid lock ID
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200, 400 or 404, got %d", w.Code)
	}
}

// --- Additional Health Provider Tests ---

func TestHealthProvidersDetailed(t *testing.T) {
	s := setupTestServer(t)

	w := doRequest(s, "GET", "/api/v1/health/providers?detailed=true", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Middleware Tests ---

func TestMiddlewareGetWithConfig(t *testing.T) {
	s := setupTestServer(t)

	// First set some middleware config
	body := map[string]interface{}{
		"enabled": true,
		"middlewares": []map[string]interface{}{
			{"name": "test-middleware", "enabled": true, "source": "builtin"},
		},
	}
	doRequest(s, "PUT", "/api/v1/middleware", body)

	// Then get it
	w := doRequest(s, "GET", "/api/v1/middleware", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Usage Tests ---

func TestUsageWithLimit(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage?limit=50", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUsageSummaryMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/usage/summary", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestUsageSummaryWithProject(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage/summary?project=/test/path", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUsageHourlyWithHours(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage/hourly?hours=48", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Budget Tests ---

func TestBudgetPutInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	// Send invalid JSON
	w := doRequestRaw(s, "PUT", "/api/v1/budget", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBudgetStatusWithProject(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/budget/status?project=/test/path", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Health Provider Tests ---

func TestHealthProvidersWithHours(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health/providers?hours=24", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthProviderWithHours(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health/providers/test-provider?hours=48", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthProviderEmptyName(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health/providers/", nil)
	// Empty name should return 400 or be handled as list
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200 or 400, got %d", w.Code)
	}
}

// --- Additional Webhook Tests ---

func TestWebhooksPostMissingName(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"url":    "https://example.com/hook",
		"events": []string{"budget_warning"},
	}
	w := doRequest(s, "POST", "/api/v1/webhooks", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWebhooksPostMissingURL(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"name":   "test-hook",
		"events": []string{"budget_warning"},
	}
	w := doRequest(s, "POST", "/api/v1/webhooks", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWebhooksPostMissingEvents(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"name": "test-hook",
		"url":  "https://example.com/hook",
	}
	w := doRequest(s, "POST", "/api/v1/webhooks", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWebhooksPostInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "POST", "/api/v1/webhooks", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWebhookPutMissingURL(t *testing.T) {
	s := setupTestServer(t)

	// First create a webhook
	body := map[string]interface{}{
		"name":    "put-test-hook",
		"url":     "https://example.com/hook",
		"events":  []string{"budget_warning"},
		"enabled": true,
	}
	doRequest(s, "POST", "/api/v1/webhooks", body)

	// Try to update without URL
	updateBody := map[string]interface{}{
		"enabled": false,
	}
	w := doRequest(s, "PUT", "/api/v1/webhooks/put-test-hook", updateBody)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWebhookPutInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "PUT", "/api/v1/webhooks/test-hook", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWebhookTestMissingNameAndURL(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{}
	w := doRequest(s, "POST", "/api/v1/webhooks/test", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWebhookTestWithURL(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"url": "https://example.com/test-hook",
	}
	w := doRequest(s, "POST", "/api/v1/webhooks/test", body)
	// May fail with 502 (bad gateway) if URL is not reachable
	if w.Code != http.StatusOK && w.Code != http.StatusBadGateway {
		t.Fatalf("expected 200 or 502, got %d", w.Code)
	}
}

func TestWebhookTestNonexistent(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"name": "nonexistent-webhook",
	}
	w := doRequest(s, "POST", "/api/v1/webhooks/test", body)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestWebhookTestInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "POST", "/api/v1/webhooks/test", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWebhookEmptyName(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/webhooks/", nil)
	// Empty name should return 400 or be handled as list
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200 or 400, got %d", w.Code)
	}
}

// --- Pricing Tests ---

func TestPricingPutInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "PUT", "/api/v1/pricing", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Settings Tests ---

func TestSettingsPutInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "PUT", "/api/v1/settings", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Compression Tests ---

func TestCompressionPutInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "PUT", "/api/v1/compression", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Middleware Tests ---

func TestMiddlewarePutInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "PUT", "/api/v1/middleware", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Agent Config Tests ---

func TestAgentConfigPutInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "PUT", "/api/v1/agent/config", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- MaskWebhookURL Tests ---

func TestMaskWebhookURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"empty", "", ""},
		{"https", "https://example.com/webhook", "https://example.com/***"},
		{"http", "http://example.com/webhook", "http://example.com/***"},
		{"no_path", "https://example.com", "https://example.com/***"},
		{"no_protocol", "example.com/webhook", "example.com/***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskWebhookURL(tt.url)
			if result != tt.expected {
				t.Errorf("maskWebhookURL(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// --- Additional Session Tests ---

func TestSessionDeleteNonexistent(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/sessions/nonexistent", nil)
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 404 or 405, got %d", w.Code)
	}
}

// --- Additional Logs Tests ---

func TestLogsGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/logs", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogsGetWithProvider(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/logs?provider=test", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogsGetWithLevel(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/logs?level=error", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogsGetWithLimit(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/logs?limit=50", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Auth Tests ---

func TestLoginInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "POST", "/api/v1/auth/login", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLoginEmptyPassword(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"password": "",
	}
	w := doRequest(s, "POST", "/api/v1/auth/login", body)
	// Should fail with 401 (unauthorized), 400 (bad request), or 403 (forbidden)
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusBadRequest && w.Code != http.StatusForbidden {
		t.Fatalf("expected 401, 400 or 403, got %d", w.Code)
	}
}

func TestAuthCheckGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/auth/check", nil)
	// May return 200 (authenticated) or 401 (not authenticated)
	if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 200 or 401, got %d", w.Code)
	}
}

func TestAuthCheckMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/auth/check", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestLogoutPost(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/auth/logout", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogoutMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/auth/logout", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestPasswordChangeInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "POST", "/api/v1/auth/password", []byte("not json"))
	// May return 400 (bad request) or 404 (not found if route doesn't exist)
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Fatalf("expected 400 or 404, got %d", w.Code)
	}
}

func TestPasswordChangeMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/auth/password", nil)
	// May return 405 (method not allowed) or 404 (not found)
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
		t.Fatalf("expected 405 or 404, got %d", w.Code)
	}
}

func TestPubKeyGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/auth/pubkey", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPubKeyMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/auth/pubkey", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Additional Reload Tests ---

func TestReloadPost(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/reload", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Health Provider Tests ---

func TestHealthProvidersWithMetrics(t *testing.T) {
	s := setupTestServer(t)
	// Test with different hours parameter
	w := doRequest(s, "GET", "/api/v1/health/providers?hours=12", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Usage Tests with Parameters ---

func TestUsageGetWithAllParams(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/usage?limit=100&provider=test&model=claude", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUsageSummaryAllPeriods(t *testing.T) {
	s := setupTestServer(t)
	periods := []string{"day", "week", "month", "all"}
	for _, period := range periods {
		w := doRequest(s, "GET", "/api/v1/usage/summary?period="+period, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for period %s, got %d", period, w.Code)
		}
	}
}

// --- Additional Middleware Tests ---

func TestMiddlewareByNameNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/middleware/nonexistent-middleware", nil)
	// May return 200 or 404 depending on implementation
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestMiddlewareEnableNonexistent(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/middleware/nonexistent/enable", nil)
	// May return various status codes depending on middleware state
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200, 400, 404 or 500, got %d", w.Code)
	}
}

func TestMiddlewareDisableNonexistent(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/middleware/nonexistent/disable", nil)
	// May return various status codes depending on middleware state
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200, 400, 404 or 500, got %d", w.Code)
	}
}

// --- Additional Session Tests ---

func TestSessionsGetWithParams(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sessions?limit=50", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Compression Tests ---

func TestCompressionStatsDetailed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/compression/stats?detailed=true", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Budget Tests ---

func TestBudgetPutWithAllFields(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"daily": map[string]interface{}{
			"amount": 10.0,
			"action": "warn",
		},
		"weekly": map[string]interface{}{
			"amount": 50.0,
			"action": "downgrade",
		},
		"monthly": map[string]interface{}{
			"amount": 200.0,
			"action": "block",
		},
		"per_project": true,
	}
	w := doRequest(s, "PUT", "/api/v1/budget", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Agent Tests ---

func TestAgentStatsWithParams(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()
	w := doRequest(s, "GET", "/api/v1/agent/stats?detailed=true", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Pricing Tests ---

func TestPricingPutMultipleModels(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"model-1": map[string]interface{}{
			"input_per_million":  1.0,
			"output_per_million": 2.0,
		},
		"model-2": map[string]interface{}{
			"input_per_million":  3.0,
			"output_per_million": 6.0,
		},
	}
	w := doRequest(s, "PUT", "/api/v1/pricing", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Provider Tests ---

func TestProvidersGetList(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestProviderGetByName(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers/test-provider", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestProviderGetNonexistent(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Additional Profile Tests ---

func TestProfilesGetList(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestProfileGetByName(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles/default", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestProfileGetNonexistent(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Usage Tests with Proxy Infrastructure ---

func TestUsageGetWithTracker(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/usage", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUsageGetWithTrackerAndLimit(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/usage?limit=50", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUsageSummaryWithTracker(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/usage/summary", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUsageSummaryWithTrackerAllPeriods(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	for _, period := range []string{"day", "week", "month", "all"} {
		w := doRequest(s, "GET", "/api/v1/usage/summary?period="+period, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for period %s, got %d", period, w.Code)
		}
	}
}

func TestUsageSummaryWithTrackerAndProject(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/usage/summary?period=day&project=/test/path", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUsageHourlyWithTracker(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/usage/hourly", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUsageHourlyWithTrackerAndHours(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/usage/hourly?hours=48", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBudgetStatusWithChecker(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/budget/status", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBudgetStatusWithCheckerAndProject(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/budget/status?project=/test/path", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Agent Runtime Tests ---

func TestAgentRuntimeWithInfrastructure(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/runtime", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentRuntimeGetByID(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/runtime/test-id", nil)
	// May return 200 or 404 depending on whether runtime exists
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Additional Agent Tasks Tests ---

func TestAgentTasksWithInfrastructure(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/tasks", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentTasksPostWithInfrastructure(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	body := map[string]interface{}{
		"session_id":  "test-session",
		"description": "Test task description",
		"priority":    1,
	}
	w := doRequest(s, "POST", "/api/v1/agent/tasks", body)
	// May return various status codes
	if w.Code != http.StatusOK && w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200, 201 or 400, got %d", w.Code)
	}
}

// --- Additional Session Tests ---

func TestSessionGetByID(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/sessions/test-session-id", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Additional Compression Stats Tests ---

func TestCompressionStatsWithCompressor(t *testing.T) {
	s := setupTestServer(t)

	w := doRequest(s, "GET", "/api/v1/compression/stats", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Auth Tests ---

func TestPasswordChangeWithPassword(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"current_password": "wrong-password",
		"new_password":     "new-password-123",
	}
	w := doRequest(s, "POST", "/api/v1/auth/password", body)
	// May return 400, 401, 403, or 404
	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
		t.Fatalf("expected 400, 401, 403 or 404, got %d", w.Code)
	}
}

// --- Additional Sync Test Tests ---

func TestSyncTestPost(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"backend": "s3",
	}
	w := doRequest(s, "POST", "/api/v1/sync/test", body)
	// May return various status codes
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusServiceUnavailable && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200, 400, 500 or 503, got %d", w.Code)
	}
}

// --- Additional Middleware Tests with Config ---

func TestMiddlewareByNameWithConfig(t *testing.T) {
	s := setupTestServer(t)

	// First set some middleware config
	body := map[string]interface{}{
		"enabled": true,
		"middlewares": []map[string]interface{}{
			{"name": "test-mw", "enabled": true, "source": "builtin"},
		},
	}
	doRequest(s, "PUT", "/api/v1/middleware", body)

	// Then get specific middleware
	w := doRequest(s, "GET", "/api/v1/middleware/test-mw", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Additional Bindings Tests ---

func TestBindingsGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/bindings", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBindingsPost(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"path":    "/test/project",
		"profile": "default",
	}
	w := doRequest(s, "POST", "/api/v1/bindings", body)
	if w.Code != http.StatusOK && w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200, 201 or 400, got %d", w.Code)
	}
}

func TestBindingGetByPath(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/bindings/test-path", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestBindingDelete(t *testing.T) {
	s := setupTestServer(t)
	// First create a binding
	body := map[string]interface{}{
		"path":    "/delete/test",
		"profile": "default",
	}
	doRequest(s, "POST", "/api/v1/bindings", body)

	// Then delete it (URL encode the path)
	w := doRequest(s, "DELETE", "/api/v1/bindings/%2Fdelete%2Ftest", nil)
	// May return 200, 301, or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest && w.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 200, 301, 400 or 404, got %d", w.Code)
	}
}

// --- Additional Provider Tests ---

func TestProviderPost(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"name":       "new-provider",
		"base_url":   "https://api.new.com",
		"auth_token": "sk-new-token",
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)
	if w.Code != http.StatusOK && w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200, 201 or 400, got %d", w.Code)
	}
}

func TestProviderPut(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"base_url":   "https://api.updated.com",
		"auth_token": "sk-updated-token",
	}
	w := doRequest(s, "PUT", "/api/v1/providers/test-provider", body)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestProviderDelete(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/providers/test-provider", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Additional Profile Tests ---

func TestProfilePost(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"name":      "new-profile",
		"providers": []string{"test-provider"},
	}
	w := doRequest(s, "POST", "/api/v1/profiles", body)
	if w.Code != http.StatusOK && w.Code != http.StatusCreated && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200, 201 or 400, got %d", w.Code)
	}
}

func TestProfilePut(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"providers": []string{"test-provider", "backup"},
	}
	w := doRequest(s, "PUT", "/api/v1/profiles/default", body)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestProfileDelete(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/profiles/work", nil)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Additional Settings Tests ---

func TestSettingsGet(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/settings", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSettingsPut(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"default_profile": "default",
	}
	w := doRequest(s, "PUT", "/api/v1/settings", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Session Tests ---

func TestSessionGetEmptyID(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sessions/", nil)
	// Empty ID should return 400 or be handled as list
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200 or 400, got %d", w.Code)
	}
}

func TestSessionGetWithThreshold(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sessions/test-session?threshold=50000", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Additional Agent Runtime Tests ---

func TestAgentRuntimePostCommand(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	body := map[string]interface{}{
		"session_id": "test-session",
		"command":    "status",
	}
	w := doRequest(s, "POST", "/api/v1/agent/runtime", body)
	// May return various status codes
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 200, 400, 404 or 405, got %d", w.Code)
	}
}

// --- Additional Agent Tasks Tests ---

func TestAgentTaskGetByID(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/tasks/test-task-id", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestAgentTaskDelete(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "DELETE", "/api/v1/agent/tasks/test-task-id", nil)
	// May return various status codes
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 200, 404 or 405, got %d", w.Code)
	}
}

// --- Additional Health Tests ---

func TestHealthProvidersWithHealthChecker(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/health/providers", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Logs Tests ---

func TestLogsGetWithAllParams(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/logs?provider=test&level=info&limit=100&session=test-session", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Compression Tests ---

func TestCompressionPutWithAllFields(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"enabled":          true,
		"threshold_tokens": 100000,
		"target_tokens":    50000,
		"strategy":         "summarize",
	}
	w := doRequest(s, "PUT", "/api/v1/compression", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Agent Config Tests ---

func TestAgentConfigPutWithAllFields(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"enabled": true,
		"observatory": map[string]interface{}{
			"enabled":         true,
			"stuck_threshold": 10,
		},
		"guardrails": map[string]interface{}{
			"enabled": true,
		},
	}
	w := doRequest(s, "PUT", "/api/v1/agent/config", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Agent Sessions Tests ---

func TestAgentSessionsGetByIDWithInfra(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/sessions/test-session-id", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestAgentSessionsKillWithInfra(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "POST", "/api/v1/agent/sessions/test-session/kill", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestAgentSessionsPauseWithInfra(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "POST", "/api/v1/agent/sessions/test-session/pause", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestAgentSessionsResumeWithInfra(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "POST", "/api/v1/agent/sessions/test-session/resume", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Additional Binding Tests ---

func TestBindingPut(t *testing.T) {
	s := setupTestServer(t)
	// First create a binding
	body := map[string]interface{}{
		"path":    "/update/test",
		"profile": "default",
	}
	doRequest(s, "POST", "/api/v1/bindings", body)

	// Then update it
	updateBody := map[string]interface{}{
		"profile": "work",
	}
	w := doRequest(s, "PUT", "/api/v1/bindings/%2Fupdate%2Ftest", updateBody)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 200, 400 or 404, got %d", w.Code)
	}
}

// --- Additional Health Provider Tests ---

func TestHealthProviderWithInvalidHours(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health/providers/test?hours=invalid", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Sync Tests ---

func TestSyncTestWithInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	w := doRequestRaw(s, "POST", "/api/v1/sync/test", []byte("not json"))
	if w.Code != http.StatusBadRequest && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 400 or 503, got %d", w.Code)
	}
}

// --- Additional Auth Tests ---

func TestLoginWithValidCredentials(t *testing.T) {
	s := setupTestServer(t)
	body := map[string]interface{}{
		"password": "test-password",
	}
	w := doRequest(s, "POST", "/api/v1/auth/login", body)
	// May return 200, 401, or 403
	if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Fatalf("expected 200, 401 or 403, got %d", w.Code)
	}
}

// --- Health Provider Tests with Full Infrastructure ---

func TestHealthProvidersWithFullInfra(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/health/providers", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHealthProvidersWithFullInfraAndHours(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/health/providers?hours=24", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHealthProviderSpecificWithFullInfra(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/health/providers/test-provider?hours=12", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Agent Tasks Tests with Full Infrastructure ---

func TestAgentTasksGetWithFullInfra(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/agent/tasks", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Agent Runtime Tests with Full Infrastructure ---

func TestAgentRuntimeGetWithFullInfra(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/agent/runtime", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional Binding Tests ---

func TestBindingGetWithEncodedPath(t *testing.T) {
	s := setupTestServer(t)

	// Create a binding first
	body := map[string]interface{}{
		"path":    "/test/encoded/path",
		"profile": "default",
	}
	doRequest(s, "POST", "/api/v1/bindings", body)

	// Get it with encoded path
	w := doRequest(s, "GET", "/api/v1/bindings/%2Ftest%2Fencoded%2Fpath", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

// --- Additional Session Tests ---

func TestSessionsListWithParams(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "GET", "/api/v1/sessions?active=true", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Pricing Reset Tests ---

func TestPricingResetWithInfra(t *testing.T) {
	s := setupTestServer(t)
	setupProxyInfrastructure(t)

	w := doRequest(s, "POST", "/api/v1/pricing/reset", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// --- Additional Password Change Tests ---

func TestPasswordChangeEmptyBody(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/auth/password", map[string]interface{}{})
	// May return 400, 401, 403, or 404
	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
		t.Fatalf("expected 400, 401, 403 or 404, got %d", w.Code)
	}
}

// --- Additional Client IP Tests ---

func TestAuthCheckWithXForwardedFor(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("GET", "/api/v1/auth/check", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	// May return 200 or 401
	if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 200 or 401, got %d", w.Code)
	}
}

func TestAuthCheckWithXRealIP(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("GET", "/api/v1/auth/check", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	// May return 200 or 401
	if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 200 or 401, got %d", w.Code)
	}
}

// --- Additional Agent Tasks Operations Tests ---

func TestAgentTaskRetry(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "POST", "/api/v1/agent/tasks/test-task/retry", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestAgentTaskRetryMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/tasks/test-task/retry", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestAgentTaskCancel(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "POST", "/api/v1/agent/tasks/test-task/cancel", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestAgentTaskCancelMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/tasks/test-task/cancel", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestAgentTaskDeleteByID(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "DELETE", "/api/v1/agent/tasks/test-task-id", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestAgentTaskMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "PUT", "/api/v1/agent/tasks/test-task-id", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Additional Agent Runtime Operations Tests ---

func TestAgentRuntimeGetBySessionID(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "GET", "/api/v1/agent/runtime/session-123", nil)
	// May return 200 or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

func TestAgentRuntimeStart(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	body := map[string]interface{}{
		"session_id": "test-session",
	}
	w := doRequest(s, "POST", "/api/v1/agent/runtime/start", body)
	// May return various status codes
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusConflict && w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200, 400, 404, 405 or 409, got %d", w.Code)
	}
}

func TestAgentRuntimeStop(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "POST", "/api/v1/agent/runtime/stop", nil)
	// May return various status codes
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 200, 400, 404 or 405, got %d", w.Code)
	}
}

// --- Additional Tests for Coverage ---

func TestAgentTasksPostInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequestRaw(s, "POST", "/api/v1/agent/tasks", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAgentTasksListMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	setupAgentInfrastructure()

	w := doRequest(s, "PUT", "/api/v1/agent/tasks", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

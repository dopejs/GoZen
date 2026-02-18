package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompatWebPassword(t *testing.T) {
	// Setup temp config
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create config dir
	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)

	// Reset store
	ResetDefaultStore()

	// Test GetWebPasswordHash (empty initially)
	hash := GetWebPasswordHash()
	if hash != "" {
		t.Errorf("Expected empty hash, got %s", hash)
	}

	// Test SetWebPasswordHash
	if err := SetWebPasswordHash("$2a$10$test"); err != nil {
		t.Errorf("SetWebPasswordHash failed: %v", err)
	}

	hash = GetWebPasswordHash()
	if hash != "$2a$10$test" {
		t.Errorf("Expected $2a$10$test, got %s", hash)
	}
}

func TestCompatSyncConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	ResetDefaultStore()

	// Test GetSyncConfig (nil initially)
	cfg := GetSyncConfig()
	if cfg != nil {
		t.Errorf("Expected nil sync config, got %+v", cfg)
	}

	// Test SetSyncConfig
	syncCfg := &SyncConfig{
		Backend:  "gist",
		GistID:   "abc123",
		AutoPull: true,
	}
	if err := SetSyncConfig(syncCfg); err != nil {
		t.Errorf("SetSyncConfig failed: %v", err)
	}

	cfg = GetSyncConfig()
	if cfg == nil || cfg.Backend != "gist" {
		t.Errorf("Expected gist backend, got %+v", cfg)
	}
}

func TestCompatPricing(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	ResetDefaultStore()

	// Test GetPricing (returns defaults)
	pricing := GetPricing()
	if pricing == nil {
		t.Error("Expected default pricing, got nil")
	}

	// Test SetPricing
	customPricing := map[string]*ModelPricing{
		"custom-model": {InputPerMillion: 1.0, OutputPerMillion: 2.0},
	}
	if err := SetPricing(customPricing); err != nil {
		t.Errorf("SetPricing failed: %v", err)
	}

	pricing = GetPricing()
	if pricing["custom-model"] == nil {
		t.Error("Expected custom-model in pricing")
	}
}

func TestCompatBudgets(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	ResetDefaultStore()

	// Test GetBudgets (nil initially)
	budgets := GetBudgets()
	if budgets != nil {
		t.Errorf("Expected nil budgets, got %+v", budgets)
	}

	// Test SetBudgets
	budgetCfg := &BudgetConfig{
		Daily: &BudgetLimit{Amount: 10.0, Action: "warn"},
	}
	if err := SetBudgets(budgetCfg); err != nil {
		t.Errorf("SetBudgets failed: %v", err)
	}

	budgets = GetBudgets()
	if budgets == nil || budgets.Daily == nil {
		t.Error("Expected daily budget to be set")
	}
}

func TestCompatWebhooks(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	ResetDefaultStore()

	// Test GetWebhooks (empty initially)
	webhooks := GetWebhooks()
	if len(webhooks) != 0 {
		t.Errorf("Expected empty webhooks, got %d", len(webhooks))
	}

	// Test AddWebhook
	webhook := &WebhookConfig{
		Name:    "test-webhook",
		URL:     "https://example.com/webhook",
		Events:  []WebhookEvent{"budget_warning"},
		Enabled: true,
	}
	if err := AddWebhook(webhook); err != nil {
		t.Errorf("AddWebhook failed: %v", err)
	}

	// Test GetWebhook
	retrieved := GetWebhook("test-webhook")
	if retrieved == nil || retrieved.URL != "https://example.com/webhook" {
		t.Error("Expected to retrieve webhook")
	}

	// Test GetWebhooks
	webhooks = GetWebhooks()
	if len(webhooks) != 1 {
		t.Errorf("Expected 1 webhook, got %d", len(webhooks))
	}

	// Test SetWebhooks
	newWebhooks := []*WebhookConfig{
		{Name: "webhook1", URL: "https://example.com/1", Enabled: true},
		{Name: "webhook2", URL: "https://example.com/2", Enabled: false},
	}
	if err := SetWebhooks(newWebhooks); err != nil {
		t.Errorf("SetWebhooks failed: %v", err)
	}

	webhooks = GetWebhooks()
	if len(webhooks) != 2 {
		t.Errorf("Expected 2 webhooks, got %d", len(webhooks))
	}

	// Test DeleteWebhook
	if err := DeleteWebhook("webhook1"); err != nil {
		t.Errorf("DeleteWebhook failed: %v", err)
	}

	webhooks = GetWebhooks()
	if len(webhooks) != 1 {
		t.Errorf("Expected 1 webhook after delete, got %d", len(webhooks))
	}
}

func TestCompatHealthCheck(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	ResetDefaultStore()

	// Test GetHealthCheck (nil initially)
	hc := GetHealthCheck()
	if hc != nil {
		t.Errorf("Expected nil health check, got %+v", hc)
	}

	// Test SetHealthCheck
	healthCfg := &HealthCheckConfig{
		Enabled:      true,
		IntervalSecs: 60,
		TimeoutSecs:  10,
	}
	if err := SetHealthCheck(healthCfg); err != nil {
		t.Errorf("SetHealthCheck failed: %v", err)
	}

	hc = GetHealthCheck()
	if hc == nil || !hc.Enabled {
		t.Error("Expected health check to be enabled")
	}
}

func TestCompatCompression(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	ResetDefaultStore()

	// Test GetCompression (nil initially)
	cc := GetCompression()
	if cc != nil {
		t.Errorf("Expected nil compression config, got %+v", cc)
	}

	// Test SetCompression
	compressionCfg := &CompressionConfig{
		Enabled:         true,
		ThresholdTokens: 50000,
		TargetTokens:    30000,
	}
	if err := SetCompression(compressionCfg); err != nil {
		t.Errorf("SetCompression failed: %v", err)
	}

	cc = GetCompression()
	if cc == nil || !cc.Enabled {
		t.Error("Expected compression to be enabled")
	}
}

func TestCompatMiddleware(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	ResetDefaultStore()

	// Test GetMiddleware (nil initially)
	mc := GetMiddleware()
	if mc != nil {
		t.Errorf("Expected nil middleware config, got %+v", mc)
	}

	// Test SetMiddleware
	middlewareCfg := &MiddlewareConfig{
		Enabled: true,
		Middlewares: []*MiddlewareEntry{
			{Name: "request-logger", Enabled: true},
		},
	}
	if err := SetMiddleware(middlewareCfg); err != nil {
		t.Errorf("SetMiddleware failed: %v", err)
	}

	mc = GetMiddleware()
	if mc == nil || !mc.Enabled {
		t.Error("Expected middleware to be enabled")
	}
}

func TestCompatAgent(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	ResetDefaultStore()

	// Test GetAgent (nil initially)
	ac := GetAgent()
	if ac != nil {
		t.Errorf("Expected nil agent config, got %+v", ac)
	}

	// Test SetAgent
	agentCfg := &AgentConfig{
		Enabled: true,
		Observatory: &ObservatoryConfig{
			Enabled:        true,
			StuckThreshold: 5,
		},
	}
	if err := SetAgent(agentCfg); err != nil {
		t.Errorf("SetAgent failed: %v", err)
	}

	ac = GetAgent()
	if ac == nil || !ac.Enabled {
		t.Error("Expected agent to be enabled")
	}
}

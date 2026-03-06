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

// TestGetFeatureGates tests GetFeatureGates with various states (table-driven).
func TestGetFeatureGates(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Store)
		want      *FeatureGates
	}{
		{
			name:      "nil FeatureGates returns empty struct",
			setupFunc: func(s *Store) {
				// Don't set FeatureGates - should be nil
			},
			want: &FeatureGates{},
		},
		{
			name: "existing FeatureGates returned",
			setupFunc: func(s *Store) {
				s.SetFeatureGates(&FeatureGates{Bot: true, Agent: true})
			},
			want: &FeatureGates{Bot: true, Agent: true},
		},
		{
			name: "all features enabled",
			setupFunc: func(s *Store) {
				s.SetFeatureGates(&FeatureGates{Bot: true, Compression: true, Middleware: true, Agent: true})
			},
			want: &FeatureGates{Bot: true, Compression: true, Middleware: true, Agent: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", oldHome)

			configDir := filepath.Join(tmpDir, ".zen")
			os.MkdirAll(configDir, 0755)
			ResetDefaultStore()

			store := DefaultStore()
			tt.setupFunc(store)

			got := GetFeatureGates()
			if *got != *tt.want {
				t.Errorf("GetFeatureGates() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestSetFeatureGatesCreate tests SetFeatureGates creating new FeatureGates (table-driven).
func TestSetFeatureGatesCreate(t *testing.T) {
	tests := []struct {
		name  string
		gates *FeatureGates
	}{
		{
			name:  "create with bot enabled",
			gates: &FeatureGates{Bot: true},
		},
		{
			name:  "create with all enabled",
			gates: &FeatureGates{Bot: true, Compression: true, Middleware: true, Agent: true},
		},
		{
			name:  "create with mixed state",
			gates: &FeatureGates{Compression: true, Agent: true},
		},
		{
			name:  "create with all disabled",
			gates: &FeatureGates{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", oldHome)

			configDir := filepath.Join(tmpDir, ".zen")
			os.MkdirAll(configDir, 0755)
			ResetDefaultStore()

			// Initially nil
			got := GetFeatureGates()
			if *got != (FeatureGates{}) {
				t.Errorf("Initial GetFeatureGates() = %+v, want empty", got)
			}

			// Set feature gates
			if err := SetFeatureGates(tt.gates); err != nil {
				t.Fatalf("SetFeatureGates() error = %v", err)
			}

			// Verify it was set
			got = GetFeatureGates()
			if *got != *tt.gates {
				t.Errorf("After SetFeatureGates(), GetFeatureGates() = %+v, want %+v", got, tt.gates)
			}
		})
	}
}

// TestSetFeatureGatesModify tests SetFeatureGates modifying existing FeatureGates (table-driven).
func TestSetFeatureGatesModify(t *testing.T) {
	tests := []struct {
		name     string
		initial  *FeatureGates
		modified *FeatureGates
	}{
		{
			name:     "enable bot",
			initial:  &FeatureGates{},
			modified: &FeatureGates{Bot: true},
		},
		{
			name:     "disable bot",
			initial:  &FeatureGates{Bot: true},
			modified: &FeatureGates{},
		},
		{
			name:     "enable multiple features",
			initial:  &FeatureGates{Bot: true},
			modified: &FeatureGates{Bot: true, Compression: true, Agent: true},
		},
		{
			name:     "toggle features",
			initial:  &FeatureGates{Bot: true, Compression: true},
			modified: &FeatureGates{Middleware: true, Agent: true},
		},
		{
			name:     "enable all then disable all",
			initial:  &FeatureGates{Bot: true, Compression: true, Middleware: true, Agent: true},
			modified: &FeatureGates{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", oldHome)

			configDir := filepath.Join(tmpDir, ".zen")
			os.MkdirAll(configDir, 0755)
			ResetDefaultStore()

			// Set initial state
			if err := SetFeatureGates(tt.initial); err != nil {
				t.Fatalf("SetFeatureGates(initial) error = %v", err)
			}

			// Verify initial state
			got := GetFeatureGates()
			if *got != *tt.initial {
				t.Errorf("Initial GetFeatureGates() = %+v, want %+v", got, tt.initial)
			}

			// Modify feature gates
			if err := SetFeatureGates(tt.modified); err != nil {
				t.Fatalf("SetFeatureGates(modified) error = %v", err)
			}

			// Verify modified state
			got = GetFeatureGates()
			if *got != *tt.modified {
				t.Errorf("After modification, GetFeatureGates() = %+v, want %+v", got, tt.modified)
			}
		})
	}
}

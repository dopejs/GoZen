package proxy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

func TestNewHealthChecker(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	hc := NewHealthChecker(nil)
	if hc == nil {
		t.Fatal("Expected non-nil health checker")
	}
	if hc.client == nil {
		t.Error("Expected client to be set")
	}
	if hc.statuses == nil {
		t.Error("Expected statuses map to be initialized")
	}
}

func TestHealthChecker_ReloadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	hc := NewHealthChecker(nil)

	// Set health check config
	config.SetHealthCheck(&config.HealthCheckConfig{
		Enabled:      true,
		IntervalSecs: 30,
		TimeoutSecs:  5,
	})

	hc.ReloadConfig()

	if hc.config == nil {
		t.Error("Expected config to be reloaded")
	}
	if hc.client.Timeout != 5*time.Second {
		t.Error("Expected timeout to be updated")
	}
}

func TestHealthChecker_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	hc := NewHealthChecker(nil)

	// Start
	hc.Start()
	if !hc.IsRunning() {
		t.Error("Expected health checker to be running")
	}

	// Start again (should be no-op)
	hc.Start()
	if !hc.IsRunning() {
		t.Error("Expected health checker to still be running")
	}

	// Stop
	hc.Stop()
	if hc.IsRunning() {
		t.Error("Expected health checker to be stopped")
	}

	// Stop again (should be no-op)
	hc.Stop()
}

func TestHealthChecker_CheckProvider(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hc := &HealthChecker{
		client:   &http.Client{Timeout: 5 * time.Second},
		statuses: make(map[string]*ProviderHealthStatus),
	}

	result := hc.CheckProvider("test", server.URL)
	if !result.Healthy {
		t.Errorf("Expected healthy result, got error: %s", result.Error)
	}
	// Latency can be 0 for very fast local requests
	if result.LatencyMs < 0 {
		t.Error("Expected non-negative latency")
	}
}

func TestHealthChecker_CheckProvider_NoURL(t *testing.T) {
	hc := &HealthChecker{
		client:   &http.Client{Timeout: 5 * time.Second},
		statuses: make(map[string]*ProviderHealthStatus),
	}

	result := hc.CheckProvider("test", "")
	if result.Healthy {
		t.Error("Expected unhealthy result for empty URL")
	}
	if result.Error != "no base URL configured" {
		t.Errorf("Expected 'no base URL configured' error, got: %s", result.Error)
	}
}

func TestHealthChecker_CheckProvider_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	hc := &HealthChecker{
		client:   &http.Client{Timeout: 5 * time.Second},
		statuses: make(map[string]*ProviderHealthStatus),
	}

	result := hc.CheckProvider("test", server.URL)
	if result.Healthy {
		t.Error("Expected unhealthy result for server error")
	}
}

func TestHealthChecker_GetStatus(t *testing.T) {
	now := time.Now()
	hc := &HealthChecker{
		statuses: map[string]*ProviderHealthStatus{
			"test": {
				Provider:    "test",
				Status:      HealthStatusHealthy,
				SuccessRate: 100,
				LastSuccess: &now,
			},
		},
	}

	status := hc.GetStatus("test")
	if status.Provider != "test" {
		t.Error("Expected test provider")
	}
	if status.Status != HealthStatusHealthy {
		t.Error("Expected healthy status")
	}

	// Unknown provider
	status = hc.GetStatus("unknown")
	if status.Status != HealthStatusUnknown {
		t.Error("Expected unknown status for unknown provider")
	}
}

func TestHealthChecker_GetAllStatus(t *testing.T) {
	hc := &HealthChecker{
		statuses: map[string]*ProviderHealthStatus{
			"p1": {Provider: "p1", Status: HealthStatusHealthy},
			"p2": {Provider: "p2", Status: HealthStatusDegraded},
		},
	}

	statuses := hc.GetAllStatus()
	if len(statuses) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statuses))
	}
}

func TestHealthChecker_GetStatusFromMetrics_NilDB(t *testing.T) {
	hc := &HealthChecker{
		db: nil,
		statuses: map[string]*ProviderHealthStatus{
			"test": {Provider: "test", Status: HealthStatusHealthy},
		},
	}

	status, err := hc.GetStatusFromMetrics("test", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status.Provider != "test" {
		t.Error("Expected test provider")
	}
}

func TestHealthChecker_DetermineStatus(t *testing.T) {
	hc := &HealthChecker{}
	now := time.Now()
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name     string
		status   *ProviderHealthStatus
		expected HealthStatus
	}{
		{
			name:     "no checks",
			status:   &ProviderHealthStatus{CheckCount: 0},
			expected: HealthStatusUnknown,
		},
		{
			name: "high success rate",
			status: &ProviderHealthStatus{
				CheckCount:  100,
				SuccessRate: 99,
				LastSuccess: &now,
			},
			expected: HealthStatusHealthy,
		},
		{
			name: "medium success rate",
			status: &ProviderHealthStatus{
				CheckCount:  100,
				SuccessRate: 80,
				LastSuccess: &now,
			},
			expected: HealthStatusDegraded,
		},
		{
			name: "low success rate",
			status: &ProviderHealthStatus{
				CheckCount:  100,
				SuccessRate: 50,
				LastSuccess: &now,
			},
			expected: HealthStatusUnhealthy,
		},
		{
			name: "recent failure with low success",
			status: &ProviderHealthStatus{
				CheckCount:  100,
				SuccessRate: 40,
				LastSuccess: &past,
				LastError:   &now,
			},
			expected: HealthStatusUnhealthy,
		},
		{
			name: "only failures",
			status: &ProviderHealthStatus{
				CheckCount: 10,
				LastError:  &now,
			},
			expected: HealthStatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hc.determineStatus(tt.status)
			if result != tt.expected {
				t.Errorf("determineStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGlobalHealthChecker(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Save old global
	oldGlobal := globalHealthChecker
	defer func() { globalHealthChecker = oldGlobal }()

	InitGlobalHealthChecker(nil)

	hc := GetGlobalHealthChecker()
	if hc == nil {
		t.Error("Expected non-nil global health checker")
	}

	// Test StartGlobalHealthChecker without enabled config
	StartGlobalHealthChecker()

	// Test StopGlobalHealthChecker
	StopGlobalHealthChecker()
}

func TestStartGlobalHealthChecker_Enabled(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Save old global
	oldGlobal := globalHealthChecker
	defer func() { globalHealthChecker = oldGlobal }()

	// Enable health check
	config.SetHealthCheck(&config.HealthCheckConfig{
		Enabled:      true,
		IntervalSecs: 60,
	})

	InitGlobalHealthChecker(nil)
	StartGlobalHealthChecker()

	hc := GetGlobalHealthChecker()
	if !hc.IsRunning() {
		t.Error("Expected health checker to be running when enabled")
	}

	StopGlobalHealthChecker()
}

func TestHealthChecker_UpdateStatus(t *testing.T) {
	hc := &HealthChecker{
		statuses: make(map[string]*ProviderHealthStatus),
	}

	// First check - creates new status
	result := &HealthResult{
		Provider:  "test",
		Healthy:   true,
		LatencyMs: 100,
		Timestamp: time.Now(),
	}
	hc.updateStatus(result)

	status := hc.statuses["test"]
	if status == nil {
		t.Fatal("Expected status to be created")
	}
	if status.CheckCount != 1 {
		t.Errorf("Expected check count 1, got %d", status.CheckCount)
	}
	if status.FailCount != 0 {
		t.Errorf("Expected fail count 0, got %d", status.FailCount)
	}

	// Second check - failure
	result = &HealthResult{
		Provider:  "test",
		Healthy:   false,
		Error:     "connection refused",
		Timestamp: time.Now(),
	}
	hc.updateStatus(result)

	status = hc.statuses["test"]
	if status.CheckCount != 2 {
		t.Errorf("Expected check count 2, got %d", status.CheckCount)
	}
	if status.FailCount != 1 {
		t.Errorf("Expected fail count 1, got %d", status.FailCount)
	}
	if status.LastErrorMsg != "connection refused" {
		t.Errorf("Expected error message, got %s", status.LastErrorMsg)
	}
}

func TestHealthChecker_CheckAllProviders(t *testing.T) {
	// Create test servers
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthyServer.Close()

	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Add providers to config using SetProvider
	config.SetProvider("healthy", &config.ProviderConfig{
		BaseURL: healthyServer.URL,
	})

	hc := NewHealthChecker(nil)
	hc.checkAllProviders()

	// Check that statuses were updated
	if len(hc.statuses) < 1 {
		t.Errorf("Expected at least 1 status, got %d", len(hc.statuses))
	}
}

func TestHealthChecker_GetStatusFromMetrics_WithStatus(t *testing.T) {
	now := time.Now()
	hc := &HealthChecker{
		db: nil,
		statuses: map[string]*ProviderHealthStatus{
			"test": {
				Provider:    "test",
				Status:      HealthStatusHealthy,
				SuccessRate: 99.5,
				LatencyMs:   150,
				LastSuccess: &now,
			},
		},
	}

	status, err := hc.GetStatusFromMetrics("test", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status.SuccessRate != 99.5 {
		t.Errorf("Expected success rate 99.5, got %f", status.SuccessRate)
	}
	if status.LatencyMs != 150 {
		t.Errorf("Expected latency 150, got %d", status.LatencyMs)
	}
}

func TestHealthChecker_CheckProvider_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	hc := &HealthChecker{
		client:   &http.Client{Timeout: 5 * time.Second},
		statuses: make(map[string]*ProviderHealthStatus),
	}

	result := hc.CheckProvider("test", server.URL)
	// 429 is < 500, so it's considered "reachable" (healthy)
	if !result.Healthy {
		t.Error("Expected healthy result for rate limit (429 < 500)")
	}
}

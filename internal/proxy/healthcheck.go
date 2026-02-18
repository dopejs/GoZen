package proxy

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// HealthStatus represents the health status of a provider.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ProviderHealthStatus holds the current health status of a provider.
type ProviderHealthStatus struct {
	Provider      string        `json:"provider"`
	Status        HealthStatus  `json:"status"`
	LastCheck     *time.Time    `json:"last_check,omitempty"`
	LastSuccess   *time.Time    `json:"last_success,omitempty"`
	LastError     *time.Time    `json:"last_error,omitempty"`
	LastErrorMsg  string        `json:"last_error_msg,omitempty"`
	LatencyMs     int           `json:"latency_ms,omitempty"`
	SuccessRate   float64       `json:"success_rate"`
	CheckCount    int           `json:"check_count"`
	FailCount     int           `json:"fail_count"`
}

// HealthResult represents the result of a single health check.
type HealthResult struct {
	Provider  string
	Healthy   bool
	LatencyMs int
	Error     string
	Timestamp time.Time
}

// HealthChecker performs periodic health checks on providers.
type HealthChecker struct {
	db       *LogDB
	config   *config.HealthCheckConfig
	client   *http.Client
	stopCh   chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	statuses map[string]*ProviderHealthStatus
	running  bool
	stopped  bool // tracks if stopCh has been closed
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(db *LogDB) *HealthChecker {
	cfg := config.GetHealthCheck()

	timeout := 10 * time.Second
	if cfg != nil && cfg.TimeoutSecs > 0 {
		timeout = time.Duration(cfg.TimeoutSecs) * time.Second
	}

	return &HealthChecker{
		db:       db,
		config:   cfg,
		client:   &http.Client{Timeout: timeout},
		stopCh:   make(chan struct{}),
		statuses: make(map[string]*ProviderHealthStatus),
	}
}

// ReloadConfig refreshes the health check configuration.
func (h *HealthChecker) ReloadConfig() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config.GetHealthCheck()

	if h.config != nil && h.config.TimeoutSecs > 0 {
		h.client.Timeout = time.Duration(h.config.TimeoutSecs) * time.Second
	}
}

// Start begins periodic health checking.
func (h *HealthChecker) Start() {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return
	}
	h.running = true
	h.mu.Unlock()

	h.wg.Add(1)
	go h.checkLoop()
}

// Stop stops the health checker.
func (h *HealthChecker) Stop() {
	h.mu.Lock()
	if !h.running || h.stopped {
		h.mu.Unlock()
		return
	}
	h.running = false
	h.stopped = true
	h.mu.Unlock()

	close(h.stopCh)
	h.wg.Wait()
}

// IsRunning returns whether the health checker is running.
func (h *HealthChecker) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

func (h *HealthChecker) checkLoop() {
	defer h.wg.Done()

	// Initial check
	h.checkAllProviders()

	interval := 60 * time.Second
	h.mu.RLock()
	if h.config != nil && h.config.IntervalSecs > 0 {
		interval = time.Duration(h.config.IntervalSecs) * time.Second
	}
	h.mu.RUnlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.mu.RLock()
			cfg := h.config
			h.mu.RUnlock()

			if cfg == nil || !cfg.Enabled {
				continue
			}

			h.checkAllProviders()
		}
	}
}

func (h *HealthChecker) checkAllProviders() {
	providers := config.ProviderNames()

	for _, name := range providers {
		p := config.GetProvider(name)
		if p == nil {
			continue
		}

		result := h.CheckProvider(name, p.BaseURL)
		h.updateStatus(result)
	}
}

// CheckProvider performs a health check on a single provider.
func (h *HealthChecker) CheckProvider(name string, baseURL string) *HealthResult {
	result := &HealthResult{
		Provider:  name,
		Timestamp: time.Now(),
	}

	if baseURL == "" {
		result.Error = "no base URL configured"
		return result
	}

	// Simple connectivity check - HEAD request to base URL
	ctx, cancel := context.WithTimeout(context.Background(), h.client.Timeout)
	defer cancel()

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, baseURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	resp, err := h.client.Do(req)
	result.LatencyMs = int(time.Since(start).Milliseconds())

	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	// Consider 2xx, 3xx, and 401/403 (auth required) as "reachable"
	if resp.StatusCode < 500 {
		result.Healthy = true
	} else {
		result.Error = "server error: " + resp.Status
	}

	return result
}

func (h *HealthChecker) updateStatus(result *HealthResult) {
	h.mu.Lock()
	defer h.mu.Unlock()

	status, ok := h.statuses[result.Provider]
	if !ok {
		status = &ProviderHealthStatus{
			Provider: result.Provider,
			Status:   HealthStatusUnknown,
		}
		h.statuses[result.Provider] = status
	}

	now := result.Timestamp
	status.LastCheck = &now
	status.CheckCount++
	status.LatencyMs = result.LatencyMs

	if result.Healthy {
		status.LastSuccess = &now
		status.LastErrorMsg = ""
	} else {
		status.FailCount++
		status.LastError = &now
		status.LastErrorMsg = result.Error
	}

	// Calculate success rate
	if status.CheckCount > 0 {
		status.SuccessRate = float64(status.CheckCount-status.FailCount) / float64(status.CheckCount) * 100
	}

	// Determine overall status
	status.Status = h.determineStatus(status)

	// Record metric in database
	if h.db != nil {
		h.db.RecordMetric(result.Provider, result.LatencyMs, 0, !result.Healthy, false)
	}
}

func (h *HealthChecker) determineStatus(status *ProviderHealthStatus) HealthStatus {
	if status.CheckCount == 0 {
		return HealthStatusUnknown
	}

	// Recent failure
	if status.LastError != nil && status.LastSuccess != nil {
		if status.LastError.After(*status.LastSuccess) {
			// Last check was a failure
			if status.SuccessRate < 50 {
				return HealthStatusUnhealthy
			}
			return HealthStatusDegraded
		}
	}

	// Only failures
	if status.LastSuccess == nil && status.LastError != nil {
		return HealthStatusUnhealthy
	}

	// Check success rate
	if status.SuccessRate >= 95 {
		return HealthStatusHealthy
	} else if status.SuccessRate >= 70 {
		return HealthStatusDegraded
	}

	return HealthStatusUnhealthy
}

// GetStatus returns the health status for a provider.
func (h *HealthChecker) GetStatus(provider string) *ProviderHealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if status, ok := h.statuses[provider]; ok {
		// Return a copy
		copy := *status
		return &copy
	}

	return &ProviderHealthStatus{
		Provider: provider,
		Status:   HealthStatusUnknown,
	}
}

// GetAllStatus returns health status for all known providers.
func (h *HealthChecker) GetAllStatus() []*ProviderHealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*ProviderHealthStatus, 0, len(h.statuses))
	for _, status := range h.statuses {
		copy := *status
		result = append(result, &copy)
	}

	return result
}

// GetStatusFromMetrics returns health status based on historical metrics.
func (h *HealthChecker) GetStatusFromMetrics(provider string, since time.Time) (*ProviderHealthStatus, error) {
	if h.db == nil {
		return h.GetStatus(provider), nil
	}

	metrics, err := h.db.GetProviderMetrics(provider, since)
	if err != nil {
		return nil, err
	}

	status := &ProviderHealthStatus{
		Provider:    provider,
		SuccessRate: metrics.SuccessRate,
		CheckCount:  metrics.TotalRequests,
		FailCount:   metrics.ErrorCount,
		LatencyMs:   int(metrics.AvgLatencyMs),
	}

	if metrics.LastSuccess != nil {
		status.LastSuccess = metrics.LastSuccess
	}
	if metrics.LastError != nil {
		status.LastError = metrics.LastError
	}

	// Determine status from metrics
	if metrics.TotalRequests == 0 {
		status.Status = HealthStatusUnknown
	} else if metrics.SuccessRate >= 95 {
		status.Status = HealthStatusHealthy
	} else if metrics.SuccessRate >= 70 {
		status.Status = HealthStatusDegraded
	} else {
		status.Status = HealthStatusUnhealthy
	}

	return status, nil
}

// --- Global health checker ---

var globalHealthChecker *HealthChecker

// InitGlobalHealthChecker initializes the global health checker.
func InitGlobalHealthChecker(db *LogDB) {
	globalHealthChecker = NewHealthChecker(db)
}

// GetGlobalHealthChecker returns the global health checker.
func GetGlobalHealthChecker() *HealthChecker {
	return globalHealthChecker
}

// StartGlobalHealthChecker starts the global health checker if enabled.
func StartGlobalHealthChecker() {
	if globalHealthChecker == nil {
		return
	}

	cfg := config.GetHealthCheck()
	if cfg != nil && cfg.Enabled {
		globalHealthChecker.Start()
	}
}

// StopGlobalHealthChecker stops the global health checker.
func StopGlobalHealthChecker() {
	if globalHealthChecker != nil {
		globalHealthChecker.Stop()
	}
}

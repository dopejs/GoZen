package daemon

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Metrics tracks request statistics for the daemon
type Metrics struct {
	mu sync.RWMutex

	// Counters
	totalRequests int64
	successCount  int64
	errorCount    int64

	// Latency ring buffer (for percentile calculation)
	latencies     []float64
	latencyIndex  int
	latencyFilled bool

	// Error tracking
	errorsByProvider map[string]int64
	errorsByType     map[string]int64

	// Resource peaks
	peakGoroutines int
	peakMemoryMB   int64

	startTime time.Time
}

// MetricsStats is the response schema for GET /api/v1/daemon/metrics
type MetricsStats struct {
	TotalRequests    int64            `json:"total_requests"`
	SuccessCount     int64            `json:"success_count"`
	ErrorCount       int64            `json:"error_count"`
	LatencyP50Ms     float64          `json:"latency_p50_ms"`
	LatencyP95Ms     float64          `json:"latency_p95_ms"`
	LatencyP99Ms     float64          `json:"latency_p99_ms"`
	ErrorsByProvider map[string]int64 `json:"errors_by_provider"`
	ErrorsByType     map[string]int64 `json:"errors_by_type"`
	PeakGoroutines   int              `json:"peak_goroutines"`
	PeakMemoryMB     int64            `json:"peak_memory_mb"`
	UptimeSeconds    int64            `json:"uptime_seconds"`
}

// RequestError represents an error from a request
type RequestError struct {
	Provider string
	Type     string
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Provider, e.Type)
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		latencies:        make([]float64, 1000), // Ring buffer size
		errorsByProvider: make(map[string]int64),
		errorsByType:     make(map[string]int64),
		startTime:        time.Now(),
	}
}

// RecordRequest records a request with latency and optional error
func (m *Metrics) RecordRequest(provider string, latency time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++

	// Record latency in ring buffer
	latencyMs := float64(latency.Milliseconds())
	m.latencies[m.latencyIndex] = latencyMs
	m.latencyIndex++
	if m.latencyIndex >= len(m.latencies) {
		m.latencyIndex = 0
		m.latencyFilled = true
	}

	if err != nil {
		m.errorCount++

		// Only record to errors_by_provider if provider is specified
		// Empty provider means system-level error (e.g., concurrency limit)
		if provider != "" {
			m.errorsByProvider[provider]++
		}

		// Always record error type for classification
		errType := classifyError(err)
		m.errorsByType[errType]++
	} else {
		m.successCount++
	}
}

// classifyError attempts to classify an error by type
func classifyError(err error) string {
	if err == nil {
		return "unknown"
	}

	// Check if it's a RequestError (for backward compatibility)
	if reqErr, ok := err.(*RequestError); ok {
		return reqErr.Type
	}

	// Check if error has a Type field (duck typing for ProxyError)
	type typedError interface {
		Error() string
		Type() string
	}
	if te, ok := err.(interface{ Type() string }); ok {
		return te.Type()
	}

	// Fallback: classify by error message content
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "auth"):
		return "auth"
	case strings.Contains(errMsg, "rate limit"):
		return "rate_limit"
	case strings.Contains(errMsg, "request error"):
		return "request"
	case strings.Contains(errMsg, "server error"):
		return "server"
	case strings.Contains(errMsg, "timeout"):
		return "timeout"
	case strings.Contains(errMsg, "concurrency limit"):
		return "concurrency"
	case strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection"):
		return "network"
	default:
		return "unknown"
	}
}

// UpdateResourcePeaks updates peak resource usage
func (m *Metrics) UpdateResourcePeaks(goroutines int, memoryMB int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if goroutines > m.peakGoroutines {
		m.peakGoroutines = goroutines
	}
	if memoryMB > m.peakMemoryMB {
		m.peakMemoryMB = memoryMB
	}
}

// GetStats returns aggregated metrics statistics
func (m *Metrics) GetStats() MetricsStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy error maps
	errorsByProvider := make(map[string]int64)
	for k, v := range m.errorsByProvider {
		errorsByProvider[k] = v
	}
	errorsByType := make(map[string]int64)
	for k, v := range m.errorsByType {
		errorsByType[k] = v
	}

	return MetricsStats{
		TotalRequests:    m.totalRequests,
		SuccessCount:     m.successCount,
		ErrorCount:       m.errorCount,
		LatencyP50Ms:     m.getPercentile(0.50),
		LatencyP95Ms:     m.getPercentile(0.95),
		LatencyP99Ms:     m.getPercentile(0.99),
		ErrorsByProvider: errorsByProvider,
		ErrorsByType:     errorsByType,
		PeakGoroutines:   m.peakGoroutines,
		PeakMemoryMB:     m.peakMemoryMB,
		UptimeSeconds:    int64(time.Since(m.startTime).Seconds()),
	}
}

// getPercentile calculates percentile from ring buffer (caller must hold lock)
func (m *Metrics) getPercentile(p float64) float64 {
	// Determine how many samples we have
	sampleCount := m.latencyIndex
	if m.latencyFilled {
		sampleCount = len(m.latencies)
	}

	if sampleCount == 0 {
		return 0
	}

	// Copy and sort samples
	samples := make([]float64, sampleCount)
	copy(samples, m.latencies[:sampleCount])
	sort.Float64s(samples)

	// Calculate percentile index
	index := int(float64(sampleCount-1) * p)
	if index < 0 {
		index = 0
	}
	if index >= sampleCount {
		index = sampleCount - 1
	}

	return samples[index]
}

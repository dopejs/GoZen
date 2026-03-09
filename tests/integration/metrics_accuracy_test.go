package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/daemon"
	"github.com/dopejs/gozen/internal/proxy"
)

// TestMetricsAccuracyUnderLoad verifies metrics accuracy with 100 concurrent requests
func TestMetricsAccuracyUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	// Create mock provider
	var requestCount atomic.Int64
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		// Simulate realistic latency
		time.Sleep(time.Duration(50+time.Now().UnixNano()%50) * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "response"}},
			"usage":   map[string]int{"input_tokens": 10, "output_tokens": 20},
		})
	}))
	defer mockProvider.Close()

	// Create metrics directly
	metrics := daemon.NewMetrics()

	// Create proxy with metrics recording
	provider := createTestProvider(mockProvider.URL)
	srv := proxy.NewProxyServer([]*proxy.Provider{provider}, testLogger())
	srv.Limiter = proxy.NewLimiter(100)

	const concurrency = 100
	const requestsPerWorker = 5
	const totalRequests = concurrency * requestsPerWorker

	var wg sync.WaitGroup
	var successCount atomic.Int64
	var errorCount atomic.Int64

	startTime := time.Now()

	// Send 100 concurrent workers, each sending 5 requests
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`)
				req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				reqStart := time.Now()
				srv.ServeHTTP(w, req)
				latency := time.Since(reqStart)

				// Record in metrics
				var err error
				if w.Code >= 400 {
					err = &metricsTestError{statusCode: w.Code}
					errorCount.Add(1)
				} else {
					successCount.Add(1)
				}

				metrics.RecordRequest(provider.Name, latency, err)
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Verify all requests completed
	actualSuccess := successCount.Load()
	actualErrors := errorCount.Load()
	actualTotal := actualSuccess + actualErrors

	t.Logf("Load test completed in %v", totalDuration)
	t.Logf("Requests: total=%d, success=%d, errors=%d", actualTotal, actualSuccess, actualErrors)
	t.Logf("Provider received: %d requests", requestCount.Load())

	if actualTotal != totalRequests {
		t.Errorf("total requests = %d, want %d", actualTotal, totalRequests)
	}

	// Verify metrics accuracy
	stats := metrics.GetStats()

	t.Logf("Metrics: total=%d, success=%d, errors=%d", stats.TotalRequests, stats.SuccessCount, stats.ErrorCount)
	t.Logf("Latency: P50=%.0fms, P95=%.0fms, P99=%.0fms", stats.LatencyP50Ms, stats.LatencyP95Ms, stats.LatencyP99Ms)

	// Verify request counts match
	if stats.TotalRequests != actualTotal {
		t.Errorf("metrics total_requests = %d, want %d", stats.TotalRequests, actualTotal)
	}

	if stats.SuccessCount != actualSuccess {
		t.Errorf("metrics success_count = %d, want %d", stats.SuccessCount, actualSuccess)
	}

	if stats.ErrorCount != actualErrors {
		t.Errorf("metrics error_count = %d, want %d", stats.ErrorCount, actualErrors)
	}

	// Verify latency percentiles are reasonable (50-150ms range given our mock)
	if stats.LatencyP50Ms < 40 || stats.LatencyP50Ms > 200 {
		t.Errorf("P50 latency = %.0fms, expected in range 40-200ms", stats.LatencyP50Ms)
	}

	if stats.LatencyP95Ms < 50 || stats.LatencyP95Ms > 300 {
		t.Errorf("P95 latency = %.0fms, expected in range 50-300ms", stats.LatencyP95Ms)
	}

	if stats.LatencyP99Ms < 60 || stats.LatencyP99Ms > 400 {
		t.Errorf("P99 latency = %.0fms, expected in range 60-400ms", stats.LatencyP99Ms)
	}

	// Note: Resource peaks (goroutines, memory) are tracked by the daemon's
	// goroutineLeakMonitor in production. In this unit test, we only verify
	// the metrics collection accuracy for request counts and latencies.

	// Verify throughput
	rps := float64(actualTotal) / totalDuration.Seconds()
	t.Logf("Throughput: %.1f requests/second", rps)

	if rps < 50 {
		t.Errorf("throughput = %.1f req/s, want >= 50 req/s", rps)
	}
}

// metricsTestError implements error interface for metrics testing
type metricsTestError struct {
	statusCode int
}

func (e *metricsTestError) Error() string {
	return "test error"
}

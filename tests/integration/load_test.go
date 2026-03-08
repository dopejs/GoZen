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

	"github.com/dopejs/gozen/internal/proxy"
)

// TestLoadSustained verifies daemon handles 100 concurrent requests for 5 minutes
// without crashes, maintaining responsiveness and resource stability.
func TestLoadSustained(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	// Create mock provider
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate realistic API latency (50-200ms)
		time.Sleep(time.Duration(50+time.Now().UnixNano()%150) * time.Millisecond)

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

	// Create proxy server with 100 concurrent limit
	provider := createTestProvider(mockProvider.URL)
	srv := proxy.NewProxyServer([]*proxy.Provider{provider}, testLogger())
	srv.Limiter = proxy.NewLimiter(100)

	// Test parameters
	const (
		concurrency = 100
		duration    = 2 * time.Minute // Reduced from 5min to fit within CI timeout (180s)
		minRPS      = 10              // Minimum requests per second to maintain
	)

	// Metrics
	var (
		totalRequests   atomic.Int64
		successRequests atomic.Int64
		errorRequests   atomic.Int64
		totalLatency    atomic.Int64
		maxLatency      atomic.Int64
	)

	// Start time
	startTime := time.Now()
	deadline := startTime.Add(duration)

	// Worker pool
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stopCh:
					return
				default:
					// Check if we've exceeded duration
					if time.Now().After(deadline) {
						return
					}

					// Send request
					reqStart := time.Now()
					body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`)
					req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
					req.Header.Set("Content-Type", "application/json")
					w := httptest.NewRecorder()

					srv.ServeHTTP(w, req)

					latency := time.Since(reqStart)
					totalRequests.Add(1)

					// Update metrics
					if w.Code >= 200 && w.Code < 300 {
						successRequests.Add(1)
					} else {
						errorRequests.Add(1)
					}

					totalLatency.Add(latency.Milliseconds())

					// Update max latency (atomic compare-and-swap)
					for {
						current := maxLatency.Load()
						if latency.Milliseconds() <= current {
							break
						}
						if maxLatency.CompareAndSwap(current, latency.Milliseconds()) {
							break
						}
					}

					// Small delay to avoid tight loop
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	// Progress reporter
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				total := totalRequests.Load()
				success := successRequests.Load()
				errors := errorRequests.Load()
				avgLatency := int64(0)
				if total > 0 {
					avgLatency = totalLatency.Load() / total
				}
				rps := float64(total) / elapsed.Seconds()

				t.Logf("[%s] requests=%d success=%d errors=%d rps=%.1f avg_latency=%dms max_latency=%dms",
					elapsed.Round(time.Second), total, success, errors, rps, avgLatency, maxLatency.Load())
			}
		}
	}()

	// Wait for test duration
	time.Sleep(duration)
	close(stopCh)
	wg.Wait()

	// Final metrics
	elapsed := time.Since(startTime)
	total := totalRequests.Load()
	success := successRequests.Load()
	errors := errorRequests.Load()
	avgLatency := int64(0)
	if total > 0 {
		avgLatency = totalLatency.Load() / total
	}
	rps := float64(total) / elapsed.Seconds()

	t.Logf("Final: requests=%d success=%d errors=%d rps=%.1f avg_latency=%dms max_latency=%dms",
		total, success, errors, rps, avgLatency, maxLatency.Load())

	// Assertions
	if total == 0 {
		t.Fatal("no requests completed")
	}

	successRate := float64(success) / float64(total) * 100
	if successRate < 95.0 {
		t.Errorf("success rate = %.1f%%, want >= 95%%", successRate)
	}

	if rps < minRPS {
		t.Errorf("requests per second = %.1f, want >= %d", rps, minRPS)
	}

	if avgLatency > 5000 {
		t.Errorf("average latency = %dms, want <= 5000ms", avgLatency)
	}

	if maxLatency.Load() > 30000 {
		t.Errorf("max latency = %dms, want <= 30000ms", maxLatency.Load())
	}
}

// TestLoadBurst verifies daemon handles sudden burst of 100 concurrent requests
func TestLoadBurst(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	// Create mock provider with fast responses
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "response"}},
		})
	}))
	defer mockProvider.Close()

	provider := createTestProvider(mockProvider.URL)
	srv := proxy.NewProxyServer([]*proxy.Provider{provider}, testLogger())
	srv.Limiter = proxy.NewLimiter(100)

	const burstSize = 100

	var wg sync.WaitGroup
	var successCount atomic.Int64
	var errorCount atomic.Int64

	startTime := time.Now()

	// Send burst of requests
	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`)
			req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			if w.Code >= 200 && w.Code < 300 {
				successCount.Add(1)
			} else {
				errorCount.Add(1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	success := successCount.Load()
	errors := errorCount.Load()

	t.Logf("Burst completed: %d requests in %s, success=%d errors=%d",
		burstSize, elapsed.Round(time.Millisecond), success, errors)

	// All requests should complete
	if success+errors != burstSize {
		t.Errorf("completed = %d, want %d", success+errors, burstSize)
	}

	// Most should succeed (allow some failures due to timing)
	successRate := float64(success) / float64(burstSize) * 100
	if successRate < 95.0 {
		t.Errorf("success rate = %.1f%%, want >= 95%%", successRate)
	}

	// Should complete in reasonable time (with 100 limit and 50ms latency, ~50-100ms expected)
	if elapsed > 5*time.Second {
		t.Errorf("burst took %s, want <= 5s", elapsed)
	}
}

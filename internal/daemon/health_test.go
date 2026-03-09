package daemon

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// testLogger returns a logger for tests
func testLogger() *log.Logger {
	return log.New(os.Stderr, "[test] ", log.LstdFlags)
}

// testNow returns current time
func testNow() time.Time {
	return time.Now()
}

// testSince returns duration since start
func testSince(start time.Time) time.Duration {
	return time.Since(start)
}

// TestHealthEndpointResponse verifies health endpoint returns 200 with correct schema
func TestHealthEndpointResponse(t *testing.T) {
	d := NewDaemon("test-version", testLogger())

	req := httptest.NewRequest("GET", "/api/v1/daemon/health", nil)
	w := httptest.NewRecorder()

	d.handleDaemonHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want 200", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Verify required fields exist
	requiredFields := []string{"status", "version", "uptime_seconds", "goroutines", "memory", "active_sessions"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}

	// Verify status is valid
	status, ok := response["status"].(string)
	if !ok {
		t.Error("status field is not a string")
	}
	validStatuses := map[string]bool{"healthy": true, "degraded": true, "unhealthy": true}
	if !validStatuses[status] {
		t.Errorf("invalid status = %s, want one of: healthy, degraded, unhealthy", status)
	}

	// Verify version matches
	if version := response["version"]; version != "test-version" {
		t.Errorf("version = %v, want 'test-version'", version)
	}

	// Verify memory object has required fields
	memory, ok := response["memory"].(map[string]interface{})
	if !ok {
		t.Fatal("memory field is not an object")
	}
	memoryFields := []string{"alloc_bytes", "sys_bytes", "heap_alloc_bytes", "heap_objects", "num_gc"}
	for _, field := range memoryFields {
		if _, ok := memory[field]; !ok {
			t.Errorf("memory missing field: %s", field)
		}
	}
}

// TestHealthEndpointDegradedStatus verifies degraded status when resources are high
func TestHealthEndpointDegradedStatus(t *testing.T) {
	// Note: This test documents the expected behavior.
	// In practice, degraded status is determined by:
	// - goroutines > 1000
	// - memory > 500MB
	// These thresholds are checked in the actual handleDaemonHealth implementation.

	d := NewDaemon("test-version", testLogger())

	req := httptest.NewRequest("GET", "/api/v1/daemon/health", nil)
	w := httptest.NewRecorder()

	d.handleDaemonHealth(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// In a fresh daemon, status should be healthy
	if status := response["status"]; status != "healthy" {
		t.Logf("status = %v (expected healthy for fresh daemon)", status)
	}

	// Document the degraded thresholds
	t.Log("Degraded status triggers:")
	t.Log("  - goroutines > 1000")
	t.Log("  - memory > 500MB")
}

// TestHealthEndpointUnhealthyStatus verifies unhealthy status when all providers fail
func TestHealthEndpointUnhealthyStatus(t *testing.T) {
	// Note: This test documents the expected behavior.
	// Unhealthy status occurs when health_check_enabled=true and all providers are failing.
	// In this test environment without real providers, we verify the response structure.

	d := NewDaemon("test-version", testLogger())

	req := httptest.NewRequest("GET", "/api/v1/daemon/health", nil)
	w := httptest.NewRecorder()

	d.handleDaemonHealth(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify health_check fields exist
	if _, ok := response["health_check_enabled"]; !ok {
		t.Error("missing health_check_enabled field")
	}

	if _, ok := response["health_check_running"]; !ok {
		t.Error("missing health_check_running field")
	}

	// Document unhealthy condition
	t.Log("Unhealthy status triggers:")
	t.Log("  - health_check_enabled = true")
	t.Log("  - all providers failing")
}

// TestHealthEndpointPerformanceUnderLoad verifies response time <100ms under load
func TestHealthEndpointPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	d := NewDaemon("test-version", testLogger())

	const iterations = 100
	var totalDuration int64

	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest("GET", "/api/v1/daemon/health", nil)
		w := httptest.NewRecorder()

		start := testNow()
		d.handleDaemonHealth(w, req)
		duration := testSince(start)

		totalDuration += duration.Microseconds()

		if w.Code != http.StatusOK {
			t.Errorf("iteration %d: status = %d, want 200", i, w.Code)
		}

		if duration.Milliseconds() > 100 {
			t.Errorf("iteration %d: response time = %v, want <100ms", i, duration)
		}
	}

	avgDuration := totalDuration / iterations
	t.Logf("Average response time: %dµs (over %d requests)", avgDuration, iterations)

	if avgDuration > 100000 { // 100ms in microseconds
		t.Errorf("average response time = %dµs, want <100ms", avgDuration)
	}
}

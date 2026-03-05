//go:build integration

// Package tests contains the stress test for proxy stability.
package tests

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestE2E_StressTest sends 500+ requests through the proxy and monitors
// memory growth to verify the daemon handles sustained load without leaks.
func TestE2E_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	// Create mock providers with mixed responses
	mockOK := newMockProvider(t)
	mockFail := newMockProviderWithStatus(t, http.StatusServiceUnavailable)

	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"primary": map[string]interface{}{
				"auth_token": "key-primary",
				"base_url":   mockOK.URL,
			},
			"backup": map[string]interface{}{
				"auth_token": "key-backup",
				"base_url":   mockFail.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"primary", "backup"},
			},
		},
	)

	env.startDaemon(t)

	// Read daemon PID for memory monitoring
	pidData, err := os.ReadFile(filepath.Join(env.configDir, "zend.pid"))
	if err != nil {
		t.Fatalf("read PID: %v", err)
	}
	pid := strings.TrimSpace(string(pidData))

	// Measure memory before
	memBefore := getProcessRSS(t, pid)
	t.Logf("Memory before: %d KB (PID %s)", memBefore, pid)

	// Send 500+ requests
	const totalRequests = 500
	successCount := 0
	failCount := 0

	for i := 0; i < totalRequests; i++ {
		resp, err := env.sendProxyRequest(t, "default", fmt.Sprintf("stress-%d", i))
		if err != nil {
			failCount++
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			successCount++
		} else {
			failCount++
		}

		// Every 100 requests, switch the mock behavior to create mixed patterns
		if i%100 == 99 {
			if mockOK.DefaultResponse.StatusCode == http.StatusOK {
				mockOK.DefaultResponse = MockResponse{
					StatusCode: http.StatusServiceUnavailable,
					Body:       mockErrorBody(http.StatusServiceUnavailable),
				}
				mockFail.DefaultResponse = MockResponse{
					StatusCode: http.StatusOK,
					Body:       defaultAnthropicResponse,
				}
			} else {
				mockOK.DefaultResponse = MockResponse{
					StatusCode: http.StatusOK,
					Body:       defaultAnthropicResponse,
				}
				mockFail.DefaultResponse = MockResponse{
					StatusCode: http.StatusServiceUnavailable,
					Body:       mockErrorBody(http.StatusServiceUnavailable),
				}
			}
		}
	}

	// Wait for any pending cleanup
	time.Sleep(1 * time.Second)

	// Measure memory after
	memAfter := getProcessRSS(t, pid)
	memGrowthKB := memAfter - memBefore
	memGrowthMB := memGrowthKB / 1024

	t.Logf("Memory after: %d KB (PID %s)", memAfter, pid)
	t.Logf("Memory growth: %d KB (%d MB)", memGrowthKB, memGrowthMB)
	t.Logf("Requests: %d total, %d success, %d fail", totalRequests, successCount, failCount)

	// Verify daemon is still running
	if !env.isDaemonUp() {
		t.Fatal("daemon crashed during stress test")
	}

	// Memory growth should not exceed 50MB
	if memGrowthMB > 50 {
		t.Errorf("memory growth %d MB exceeds 50MB threshold", memGrowthMB)
	}

	// Most requests should succeed (allowing for some failures during mock switch)
	if successCount < totalRequests/2 {
		t.Errorf("too many failures: %d/%d", failCount, totalRequests)
	}
}

// getProcessRSS returns the RSS (Resident Set Size) in KB for the given PID.
// Uses `ps -o rss= -p <PID>` which works on both macOS and Linux.
func getProcessRSS(t *testing.T, pid string) int64 {
	t.Helper()

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ps", "-o", "rss=", "-p", pid)
	} else {
		cmd = exec.Command("ps", "-o", "rss=", "-p", pid)
	}

	out, err := cmd.Output()
	if err != nil {
		t.Logf("warning: could not read RSS for PID %s: %v", pid, err)
		return 0
	}

	rss, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		t.Logf("warning: could not parse RSS output %q: %v", strings.TrimSpace(string(out)), err)
		return 0
	}

	return rss
}

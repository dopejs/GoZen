package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dopejs/gozen/internal/proxy"
)

// handleHealthProviders handles GET /api/v1/health/providers - returns health status for all providers.
func (s *Server) handleHealthProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	checker := proxy.GetGlobalHealthChecker()
	db := proxy.GetGlobalLogDB()

	// Get time range for metrics
	hours := 1
	if h := r.URL.Query().Get("hours"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 {
			hours = n
		}
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var statuses []*proxy.ProviderHealthStatus

	if checker != nil {
		// Get live status from health checker
		statuses = checker.GetAllStatus()
	}

	// Enrich with metrics from database
	if db != nil {
		metrics, err := db.GetAllProviderMetrics(since)
		if err == nil {
			// Create a map for quick lookup
			statusMap := make(map[string]*proxy.ProviderHealthStatus)
			for _, s := range statuses {
				statusMap[s.Provider] = s
			}

			// Add/update with metrics data
			for provider, m := range metrics {
				if existing, ok := statusMap[provider]; ok {
					// Update with metrics
					existing.SuccessRate = m.SuccessRate
					existing.CheckCount = m.TotalRequests
					existing.FailCount = m.ErrorCount
					if existing.LatencyMs == 0 {
						existing.LatencyMs = int(m.AvgLatencyMs)
					}
				} else {
					// Add new from metrics
					status := &proxy.ProviderHealthStatus{
						Provider:    provider,
						SuccessRate: m.SuccessRate,
						CheckCount:  m.TotalRequests,
						FailCount:   m.ErrorCount,
						LatencyMs:   int(m.AvgLatencyMs),
						LastSuccess: m.LastSuccess,
						LastError:   m.LastError,
					}
					// Determine status
					if m.TotalRequests == 0 {
						status.Status = proxy.HealthStatusUnknown
					} else if m.SuccessRate >= 95 {
						status.Status = proxy.HealthStatusHealthy
					} else if m.SuccessRate >= 70 {
						status.Status = proxy.HealthStatusDegraded
					} else {
						status.Status = proxy.HealthStatusUnhealthy
					}
					statuses = append(statuses, status)
				}
			}
		}
	}

	if statuses == nil {
		statuses = []*proxy.ProviderHealthStatus{}
	}

	writeJSON(w, http.StatusOK, statuses)
}

// handleHealthProvider handles GET /api/v1/health/providers/{name} - returns health status for a specific provider.
func (s *Server) handleHealthProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract provider name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/health/providers/")
	providerName := strings.TrimSuffix(path, "/")

	if providerName == "" {
		writeError(w, http.StatusBadRequest, "provider name required")
		return
	}

	checker := proxy.GetGlobalHealthChecker()
	db := proxy.GetGlobalLogDB()

	// Get time range for metrics
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 {
			hours = n
		}
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	response := struct {
		Status  *proxy.ProviderHealthStatus `json:"status"`
		Metrics *proxy.ProviderMetrics      `json:"metrics,omitempty"`
		Latency []proxy.LatencyPoint        `json:"latency,omitempty"`
	}{}

	// Get live status
	if checker != nil {
		response.Status = checker.GetStatus(providerName)
	} else {
		response.Status = &proxy.ProviderHealthStatus{
			Provider: providerName,
			Status:   proxy.HealthStatusUnknown,
		}
	}

	// Get detailed metrics
	if db != nil {
		if metrics, err := db.GetProviderMetrics(providerName, since); err == nil {
			response.Metrics = metrics
		}

		// Get latency history for charts
		bucketMinutes := 5
		if hours > 24 {
			bucketMinutes = 30
		} else if hours > 6 {
			bucketMinutes = 15
		}

		if latency, err := db.GetLatencyHistory(providerName, since, bucketMinutes); err == nil {
			response.Latency = latency
		}
	}

	writeJSON(w, http.StatusOK, response)
}

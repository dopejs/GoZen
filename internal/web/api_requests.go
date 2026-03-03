package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dopejs/gozen/internal/proxy"
)

// requestsResponse is the JSON shape for the monitoring requests API.
type requestsResponse struct {
	Requests []proxy.RequestRecord `json:"requests"`
	Total    int                   `json:"total"`
	Limit    int                   `json:"limit"`
}

// handleRequests handles GET /api/v1/monitoring/requests
func (s *Server) handleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	limit := 50 // default
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Build filter
	filter := proxy.RequestFilter{
		Provider:  query.Get("provider"),
		SessionID: query.Get("session"),
		Model:     query.Get("model"),
	}

	if minStatus := query.Get("status_min"); minStatus != "" {
		if s, err := strconv.Atoi(minStatus); err == nil {
			filter.MinStatus = s
		}
	}

	if maxStatus := query.Get("status_max"); maxStatus != "" {
		if s, err := strconv.Atoi(maxStatus); err == nil {
			filter.MaxStatus = s
		}
	}

	if startTime := query.Get("start_time"); startTime != "" {
		if ts, err := strconv.ParseInt(startTime, 10, 64); err == nil {
			filter.StartTime = time.Unix(ts, 0)
		}
	}

	if endTime := query.Get("end_time"); endTime != "" {
		if ts, err := strconv.ParseInt(endTime, 10, 64); err == nil {
			filter.EndTime = time.Unix(ts, 0)
		}
	}

	// Get records from monitor
	monitor := proxy.GetGlobalRequestMonitor()
	records := monitor.GetRecent(limit, filter)

	resp := requestsResponse{
		Requests: records,
		Total:    len(records),
		Limit:    limit,
	}

	writeJSON(w, http.StatusOK, resp)
}

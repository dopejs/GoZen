package web

import (
	"net/http"
	"strconv"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// handleUsage handles GET /api/v1/usage - returns recent usage entries.
func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tracker := proxy.GetGlobalUsageTracker()
	if tracker == nil {
		writeJSON(w, http.StatusOK, []proxy.UsageEntry{})
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	entries, err := tracker.GetRecentUsage(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, entries)
}

// handleUsageSummary handles GET /api/v1/usage/summary - returns usage summary.
func (s *Server) handleUsageSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tracker := proxy.GetGlobalUsageTracker()
	if tracker == nil {
		writeJSON(w, http.StatusOK, &proxy.UsageSummary{
			ByProvider: make(map[string]*proxy.UsageStats),
			ByModel:    make(map[string]*proxy.UsageStats),
			ByProject:  make(map[string]*proxy.UsageStats),
		})
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	projectPath := r.URL.Query().Get("project")

	summary, err := tracker.GetSummary(period, projectPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// handleUsageHourly handles GET /api/v1/usage/hourly - returns hourly usage for charts.
func (s *Server) handleUsageHourly(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tracker := proxy.GetGlobalUsageTracker()
	if tracker == nil {
		writeJSON(w, http.StatusOK, []proxy.HourlyUsage{})
		return
	}

	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 {
			hours = n
		}
	}

	data, err := tracker.GetHourlySummary(hours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// handleBudget handles GET/PUT /api/v1/budget - get or set budget config.
func (s *Server) handleBudget(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		budgets := config.GetBudgets()
		if budgets == nil {
			budgets = &config.BudgetConfig{}
		}
		writeJSON(w, http.StatusOK, budgets)

	case http.MethodPut:
		var budgets config.BudgetConfig
		if err := readJSON(r, &budgets); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if err := config.SetBudgets(&budgets); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Reload budget checker
		if checker := proxy.GetGlobalBudgetChecker(); checker != nil {
			checker.ReloadConfig()
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleBudgetStatus handles GET /api/v1/budget/status - returns current budget status.
func (s *Server) handleBudgetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	checker := proxy.GetGlobalBudgetChecker()
	if checker == nil {
		writeJSON(w, http.StatusOK, &proxy.BudgetStatus{})
		return
	}

	projectPath := r.URL.Query().Get("project")

	status, err := checker.Check(projectPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, status)
}

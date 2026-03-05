package web

import (
	"net/http"
	"strings"

	"github.com/dopejs/gozen/internal/agent"
	"github.com/dopejs/gozen/internal/config"
)

// handleAgentConfig handles GET/PUT for agent configuration.
func (s *Server) handleAgentConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := config.GetAgent()
		if cfg == nil {
			cfg = &config.AgentConfig{Enabled: false}
		}
		writeJSON(w, http.StatusOK, cfg)

	case http.MethodPut:
		var cfg config.AgentConfig
		if err := readJSON(r, &cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := config.SetAgent(&cfg); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAgentSessions handles agent session operations.
func (s *Server) handleAgentSessions(w http.ResponseWriter, r *http.Request) {
	obs := agent.GetGlobalObservatory()
	if obs == nil {
		writeError(w, http.StatusServiceUnavailable, "observatory not initialized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agent/sessions")
	if path == "" || path == "/" {
		// List all sessions
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		sessions := obs.GetAllSessions()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"sessions": sessions,
			"stats":    obs.GetStats(),
		})
		return
	}

	// Handle specific session
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	sessionID := parts[0]

	if len(parts) > 1 && parts[1] == "kill" {
		// Kill session
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if obs.KillSession(sessionID) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "killed"})
		} else {
			writeError(w, http.StatusNotFound, "session not found")
		}
		return
	}

	if len(parts) > 1 && parts[1] == "pause" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if obs.PauseSession(sessionID) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
		} else {
			writeError(w, http.StatusNotFound, "session not found")
		}
		return
	}

	if len(parts) > 1 && parts[1] == "resume" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if obs.ResumeSession(sessionID) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
		} else {
			writeError(w, http.StatusNotFound, "session not found or not paused")
		}
		return
	}

	// Get session details
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	session := obs.GetSession(sessionID)
	if session == nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// handleAgentLocks handles file lock operations.
func (s *Server) handleAgentLocks(w http.ResponseWriter, r *http.Request) {
	coord := agent.GetGlobalCoordinator()
	if coord == nil {
		writeError(w, http.StatusServiceUnavailable, "coordinator not initialized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agent/locks")
	if path == "" || path == "/" {
		// List all locks
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		locks := coord.GetAllLocks()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"locks": locks,
		})
		return
	}

	// Release specific lock
	lockPath := strings.TrimPrefix(path, "/")
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get session ID from query param
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session_id required")
		return
	}

	if coord.ReleaseLock(lockPath, sessionID) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "released"})
	} else {
		writeError(w, http.StatusNotFound, "lock not found or not owned by session")
	}
}

// handleAgentChanges handles file change history.
func (s *Server) handleAgentChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	coord := agent.GetGlobalCoordinator()
	if coord == nil {
		writeError(w, http.StatusServiceUnavailable, "coordinator not initialized")
		return
	}

	limit := 50
	changes := coord.GetRecentChanges(limit)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"changes": changes,
	})
}

// handleAgentTasks handles task queue operations.
func (s *Server) handleAgentTasks(w http.ResponseWriter, r *http.Request) {
	tq := agent.GetGlobalTaskQueue()
	if tq == nil {
		writeError(w, http.StatusServiceUnavailable, "task queue not initialized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agent/tasks")
	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			// List all tasks
			tasks := tq.GetAllTasks()
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"tasks": tasks,
				"stats": tq.GetStats(),
			})

		case http.MethodPost:
			// Add new task
			var req struct {
				Description string `json:"description"`
				Priority    int    `json:"priority"`
			}
			if err := readJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}
			task := tq.AddTask(req.Description, req.Priority)
			writeJSON(w, http.StatusCreated, task)

		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// Handle specific task
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	taskID := parts[0]

	if len(parts) > 1 && parts[1] == "retry" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if tq.RetryTask(taskID) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "retrying"})
		} else {
			writeError(w, http.StatusNotFound, "task not found or not failed")
		}
		return
	}

	if len(parts) > 1 && parts[1] == "cancel" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if tq.CancelTask(taskID) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
		} else {
			writeError(w, http.StatusNotFound, "task not found or already completed")
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		task := tq.GetTask(taskID)
		if task == nil {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeJSON(w, http.StatusOK, task)

	case http.MethodDelete:
		if tq.DeleteTask(taskID) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		} else {
			writeError(w, http.StatusNotFound, "task not found")
		}

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAgentRuntime handles autonomous runtime operations.
func (s *Server) handleAgentRuntime(w http.ResponseWriter, r *http.Request) {
	rt := agent.GetGlobalRuntime()
	if rt == nil {
		writeError(w, http.StatusServiceUnavailable, "runtime not initialized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agent/runtime")

	// POST /api/v1/agent/runtime/run - Start new task
	if path == "/run" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req struct {
			Description string `json:"description"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		task, err := rt.StartTask(req.Description)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, task)
		return
	}

	// GET /api/v1/agent/runtime - List all tasks
	if path == "" || path == "/" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		tasks := rt.GetAllTasks()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"tasks": tasks,
		})
		return
	}

	// Handle specific task
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	taskID := parts[0]

	if len(parts) > 1 && parts[1] == "cancel" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if rt.CancelTask(taskID) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
		} else {
			writeError(w, http.StatusNotFound, "task not found or already completed")
		}
		return
	}

	// GET task details
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	task := rt.GetTask(taskID)
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// handleAgentGuardrails handles guardrails operations.
func (s *Server) handleAgentGuardrails(w http.ResponseWriter, r *http.Request) {
	gr := agent.GetGlobalGuardrails()
	if gr == nil {
		writeError(w, http.StatusServiceUnavailable, "guardrails not initialized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agent/guardrails")

	// GET /api/v1/agent/guardrails/spending - Get all spending
	if path == "/spending" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		spending := gr.GetAllSpending()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"spending": spending,
		})
		return
	}

	// GET /api/v1/agent/guardrails/operations - Get sensitive operations
	if path == "/operations" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		ops := gr.GetRecentOperations(50)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"operations": ops,
		})
		return
	}

	// GET /api/v1/agent/guardrails - Get config
	if path == "" || path == "/" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		cfg := gr.GetConfig()
		writeJSON(w, http.StatusOK, cfg)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

// AgentStatsResponse holds aggregate agent statistics.
type AgentStatsResponse struct {
	Observatory map[string]interface{} `json:"observatory"`
	TaskQueue   map[string]int         `json:"task_queue"`
	Guardrails  map[string]interface{} `json:"guardrails"`
}

// handleAgentStats returns aggregate agent statistics.
func (s *Server) handleAgentStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	resp := AgentStatsResponse{
		Observatory: make(map[string]interface{}),
		TaskQueue:   make(map[string]int),
		Guardrails:  make(map[string]interface{}),
	}

	if obs := agent.GetGlobalObservatory(); obs != nil {
		resp.Observatory = obs.GetStats()
	}

	if tq := agent.GetGlobalTaskQueue(); tq != nil {
		resp.TaskQueue = tq.GetStats()
	}

	if gr := agent.GetGlobalGuardrails(); gr != nil {
		resp.Guardrails = map[string]interface{}{
			"enabled":  gr.IsEnabled(),
			"spending": gr.GetAllSpending(),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

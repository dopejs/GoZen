package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dopejs/gozen/internal/bot"
	"github.com/dopejs/gozen/internal/config"
)

// handleBotSkills handles all skill-related API endpoints.
func (s *Server) handleBotSkills(w http.ResponseWriter, r *http.Request) {
	// Trim prefix to get subpath
	prefix := "/api/v1/bot/skills"
	path := strings.TrimPrefix(r.URL.Path, prefix)
	if path == "" {
		// Handle collection operations
		switch r.Method {
		case http.MethodGet:
			s.listSkills(w, r)
		case http.MethodPost:
			s.createSkill(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// Handle item operations: /api/v1/bot/skills/{name}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 1 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	name := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getSkill(w, r, name)
	case http.MethodPut:
		s.updateSkill(w, r, name)
	case http.MethodDelete:
		s.deleteSkill(w, r, name)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listSkills returns all skills (builtin + custom) with their current state.
func (s *Server) listSkills(w http.ResponseWriter, r *http.Request) {
	gw := s.getBotGateway()
	if gw == nil {
		// Fall back to config-only view
		sc := config.GetSkillsConfig()
		if sc == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, sc.Custom)
		return
	}
	skills := gw.Skills().List()
	writeJSON(w, http.StatusOK, skills)
}

// createSkill creates a new custom skill and reloads.
func (s *Server) createSkill(w http.ResponseWriter, r *http.Request) {
	var def config.SkillDefinition
	if err := readJSON(r, &def); err != nil {
		writeError(w, http.StatusBadRequest, "invalid skill JSON")
		return
	}
	if def.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if def.Intent == "" {
		writeError(w, http.StatusBadRequest, "intent is required")
		return
	}
	if len(def.Keywords) == 0 {
		writeError(w, http.StatusBadRequest, "at least one keyword group is required")
		return
	}

	// Get current config
	sc := config.GetSkillsConfig()
	if sc == nil {
		sc = config.DefaultSkillsConfig()
	}

	// Check for duplicate name
	for _, existing := range sc.Custom {
		if existing.Name == def.Name {
			writeError(w, http.StatusConflict, "skill with this name already exists")
			return
		}
	}

	// Add skill
	sc.Custom = append(sc.Custom, def)
	if err := config.SetSkillsConfig(sc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Reload skills in gateway
	if gw := s.getBotGateway(); gw != nil {
		if err := gw.ReloadSkills(); err != nil {
			s.logger.Printf("Warning: failed to reload skills after create: %v", err)
		}
	}

	writeJSON(w, http.StatusCreated, def)
}

// getSkill returns a specific skill by name.
func (s *Server) getSkill(w http.ResponseWriter, r *http.Request, name string) {
	gw := s.getBotGateway()
	if gw == nil {
		writeError(w, http.StatusServiceUnavailable, "bot gateway not available")
		return
	}
	skill := gw.Skills().Get(name)
	if skill == nil {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	writeJSON(w, http.StatusOK, skill)
}

// updateSkill updates an existing custom skill.
func (s *Server) updateSkill(w http.ResponseWriter, r *http.Request, name string) {
	var def config.SkillDefinition
	if err := readJSON(r, &def); err != nil {
		writeError(w, http.StatusBadRequest, "invalid skill JSON")
		return
	}

	// Get current config
	sc := config.GetSkillsConfig()
	if sc == nil {
		writeError(w, http.StatusNotFound, "no skills configured")
		return
	}

	// Find and update skill
	found := false
	for i, existing := range sc.Custom {
		if existing.Name == name {
			def.Name = name // Ensure name consistency
			sc.Custom[i] = def
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "custom skill not found")
		return
	}

	if err := config.SetSkillsConfig(sc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Reload skills in gateway
	if gw := s.getBotGateway(); gw != nil {
		if err := gw.ReloadSkills(); err != nil {
			s.logger.Printf("Warning: failed to reload skills after update: %v", err)
		}
	}

	writeJSON(w, http.StatusOK, def)
}

// deleteSkill deletes a custom skill.
func (s *Server) deleteSkill(w http.ResponseWriter, r *http.Request, name string) {
	// Get current config
	sc := config.GetSkillsConfig()
	if sc == nil {
		writeError(w, http.StatusNotFound, "no skills configured")
		return
	}

	// Find and remove skill
	found := false
	for i, existing := range sc.Custom {
		if existing.Name == name {
			sc.Custom = append(sc.Custom[:i], sc.Custom[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "custom skill not found")
		return
	}

	if err := config.SetSkillsConfig(sc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Reload skills in gateway
	if gw := s.getBotGateway(); gw != nil {
		if err := gw.ReloadSkills(); err != nil {
			s.logger.Printf("Warning: failed to reload skills after delete: %v", err)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "name": name})
}

// handleBotSkillsConfig handles GET/PUT /api/v1/bot/skills/config
func (s *Server) handleBotSkillsConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sc := config.GetSkillsConfig()
		if sc == nil {
			sc = config.DefaultSkillsConfig()
		}
		// Return config without custom skills list
		resp := struct {
			Enabled             bool    `json:"enabled"`
			ConfidenceThreshold float64 `json:"confidence_threshold"`
			LLMFallback         bool    `json:"llm_fallback"`
			LogBufferSize       int     `json:"log_buffer_size"`
		}{
			Enabled:             sc.Enabled,
			ConfidenceThreshold: sc.ConfidenceThreshold,
			LLMFallback:         sc.LLMFallback,
			LogBufferSize:       sc.LogBufferSize,
		}
		writeJSON(w, http.StatusOK, resp)

	case http.MethodPut:
		var update struct {
			Enabled             *bool    `json:"enabled"`
			ConfidenceThreshold *float64 `json:"confidence_threshold"`
			LLMFallback         *bool    `json:"llm_fallback"`
			LogBufferSize       *int     `json:"log_buffer_size"`
		}
		if err := readJSON(r, &update); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		sc := config.GetSkillsConfig()
		if sc == nil {
			sc = config.DefaultSkillsConfig()
		}

		// Apply updates (only non-nil fields)
		if update.Enabled != nil {
			sc.Enabled = *update.Enabled
		}
		if update.ConfidenceThreshold != nil {
			sc.ConfidenceThreshold = *update.ConfidenceThreshold
		}
		if update.LLMFallback != nil {
			sc.LLMFallback = *update.LLMFallback
		}
		if update.LogBufferSize != nil {
			sc.LogBufferSize = *update.LogBufferSize
		}

		if err := config.SetSkillsConfig(sc); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Reload skills in gateway
		if gw := s.getBotGateway(); gw != nil {
			if err := gw.ReloadSkills(); err != nil {
				s.logger.Printf("Warning: failed to reload skills after config update: %v", err)
			}
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleBotSkillsTest handles POST /api/v1/bot/skills/test
func (s *Server) handleBotSkillsTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	gw := s.getBotGateway()
	if gw == nil {
		writeError(w, http.StatusServiceUnavailable, "bot gateway not available")
		return
	}

	result := gw.TestMatch(r.Context(), req.Message)
	if result == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"matched": false,
			"message": req.Message,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"matched":    true,
		"message":    req.Message,
		"skill":      result.Skill,
		"intent":     result.Intent,
		"confidence": result.Confidence,
		"method":     result.Method,
	})
}

// handleBotSkillsLogs handles GET /api/v1/bot/skills/logs
func (s *Server) handleBotSkillsLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	gw := s.getBotGateway()
	if gw == nil {
		writeError(w, http.StatusServiceUnavailable, "bot gateway not available")
		return
	}

	// Parse limit parameter
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	logs := gw.GetMatchLogs(limit)
	if logs == nil {
		logs = []*bot.MatchLog{}
	}
	writeJSON(w, http.StatusOK, logs)
}

package web

import (
	"net/http"
	"strings"

	"github.com/dopejs/gozen/internal/bot"
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
		writeError(w, http.StatusServiceUnavailable, "bot gateway not available")
		return
	}
	skills := gw.Skills().List()
	writeJSON(w, http.StatusOK, skills)
}

// createSkill creates a new custom skill.
func (s *Server) createSkill(w http.ResponseWriter, r *http.Request) {
	var skill bot.Skill
	if err := readJSON(r, &skill); err != nil {
		writeError(w, http.StatusBadRequest, "invalid skill JSON")
		return
	}
	// Validate skill
	if err := skill.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// TODO: Save to config and reload
	writeError(w, http.StatusNotImplemented, "not implemented")
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

// updateSkill updates an existing skill.
func (s *Server) updateSkill(w http.ResponseWriter, r *http.Request, name string) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// deleteSkill deletes a custom skill.
func (s *Server) deleteSkill(w http.ResponseWriter, r *http.Request, name string) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// handleBotSkillsConfig handles PUT /api/v1/bot/skills/config
func (s *Server) handleBotSkillsConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// handleBotSkillsTest handles POST /api/v1/bot/skills/test
func (s *Server) handleBotSkillsTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// handleBotSkillsLogs handles GET /api/v1/bot/skills/logs
func (s *Server) handleBotSkillsLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeError(w, http.StatusNotImplemented, "not implemented")
}
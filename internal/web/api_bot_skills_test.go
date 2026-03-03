package web

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dopejs/gozen/internal/bot"
	"github.com/dopejs/gozen/internal/config"
)

func setupTestServerWithSkills(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	configDir := filepath.Join(dir, config.ConfigDir)
	os.MkdirAll(configDir, 0755)
	cfg := &config.OpenCCConfig{
		Version: config.CurrentConfigVersion,
		Bot: &config.BotConfig{
			Enabled:    true,
			Profile:    "default",
			SocketPath: filepath.Join(dir, "test.sock"),
			Skills: &config.SkillsConfig{
				Enabled:             true,
				ConfidenceThreshold: 0.7,
				LLMFallback:         false,
				LogBufferSize:       100,
				Custom: []config.SkillDefinition{
					{
						Name:        "test-skill",
						Description: "A test skill",
						Intent:      "chat",
						Priority:    100,
						Keywords:    map[string][]string{"en": {"testword"}},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, config.ConfigFile), data, 0600)

	// Force reload
	config.DefaultStore()

	logger := log.New(io.Discard, "", 0)
	s := NewServer("1.0.0-test", logger, 0)

	// Create a minimal gateway with skill registry
	gwConfig := &bot.GatewayConfig{
		SocketPath: filepath.Join(dir, "test-gw.sock"),
	}
	gw := bot.NewGateway(gwConfig, logger)
	s.SetBotGateway(gw)

	return s
}

// --- T025: Skill Web API endpoint tests ---

func TestSkillAPIListSkills(t *testing.T) {
	s := setupTestServerWithSkills(t)
	w := doRequest(s, "GET", "/api/v1/bot/skills", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var skills []*bot.Skill
	decodeJSON(t, w, &skills)

	if len(skills) == 0 {
		t.Fatal("expected at least builtin skills")
	}

	// Verify builtin skills are present
	found := false
	for _, s := range skills {
		if s.Name == "process-control" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected process-control builtin skill in list")
	}
}

func TestSkillAPIGetSkill(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// Get existing builtin skill
	w := doRequest(s, "GET", "/api/v1/bot/skills/process-control", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var skill bot.Skill
	decodeJSON(t, w, &skill)
	if skill.Name != "process-control" {
		t.Errorf("skill.Name = %q, want %q", skill.Name, "process-control")
	}

	// Get non-existent skill
	w = doRequest(s, "GET", "/api/v1/bot/skills/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent skill, got %d", w.Code)
	}
}

func TestSkillAPICreateSkill(t *testing.T) {
	s := setupTestServerWithSkills(t)

	newSkill := config.SkillDefinition{
		Name:        "my-custom-skill",
		Description: "A custom skill for testing",
		Intent:      "chat",
		Priority:    100,
		Keywords:    map[string][]string{"en": {"custom", "test"}},
	}

	w := doRequest(s, "POST", "/api/v1/bot/skills", newSkill)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify skill was saved to config
	sc := config.GetSkillsConfig()
	if sc == nil {
		t.Fatal("skills config should not be nil")
	}
	found := false
	for _, s := range sc.Custom {
		if s.Name == "my-custom-skill" {
			found = true
			break
		}
	}
	if !found {
		t.Error("custom skill should be saved in config")
	}
}

func TestSkillAPICreateSkillValidation(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// Invalid JSON
	w := doRequest(s, "POST", "/api/v1/bot/skills", "not-json")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}

	// Missing name
	w = doRequest(s, "POST", "/api/v1/bot/skills", map[string]interface{}{
		"intent":   "chat",
		"keywords": map[string][]string{"en": {"test"}},
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name, got %d", w.Code)
	}

	// Missing intent
	w = doRequest(s, "POST", "/api/v1/bot/skills", map[string]interface{}{
		"name":     "test",
		"keywords": map[string][]string{"en": {"test"}},
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing intent, got %d", w.Code)
	}

	// Missing keywords
	w = doRequest(s, "POST", "/api/v1/bot/skills", map[string]interface{}{
		"name":   "test",
		"intent": "chat",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing keywords, got %d", w.Code)
	}
}

func TestSkillAPICreateSkillDuplicate(t *testing.T) {
	s := setupTestServerWithSkills(t)

	skill := config.SkillDefinition{
		Name:     "test-skill",
		Intent:   "chat",
		Keywords: map[string][]string{"en": {"dup"}},
	}

	w := doRequest(s, "POST", "/api/v1/bot/skills", skill)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate skill, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSkillAPIDeleteSkill(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// Delete existing custom skill
	w := doRequest(s, "DELETE", "/api/v1/bot/skills/test-skill", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify skill was removed from config
	sc := config.GetSkillsConfig()
	for _, s := range sc.Custom {
		if s.Name == "test-skill" {
			t.Error("skill should have been deleted from config")
		}
	}

	// Delete non-existent skill
	w = doRequest(s, "DELETE", "/api/v1/bot/skills/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for deleting nonexistent skill, got %d", w.Code)
	}
}

func TestSkillAPIUpdateSkill(t *testing.T) {
	s := setupTestServerWithSkills(t)

	update := config.SkillDefinition{
		Name:        "test-skill",
		Description: "Updated description",
		Intent:      "chat",
		Priority:    200,
		Keywords:    map[string][]string{"en": {"updated"}},
	}

	w := doRequest(s, "PUT", "/api/v1/bot/skills/test-skill", update)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify skill was updated in config
	sc := config.GetSkillsConfig()
	found := false
	for _, s := range sc.Custom {
		if s.Name == "test-skill" {
			found = true
			if s.Description != "Updated description" {
				t.Errorf("description = %q, want %q", s.Description, "Updated description")
			}
			if s.Priority != 200 {
				t.Errorf("priority = %d, want %d", s.Priority, 200)
			}
		}
	}
	if !found {
		t.Error("updated skill should still be in config")
	}

	// Update non-existent skill
	w = doRequest(s, "PUT", "/api/v1/bot/skills/nonexistent", update)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for updating nonexistent skill, got %d", w.Code)
	}

	// Update with invalid JSON
	w = doRequest(s, "PUT", "/api/v1/bot/skills/test-skill", "not-json")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestSkillAPIConfig(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// GET config
	w := doRequest(s, "GET", "/api/v1/bot/skills/config", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var cfg struct {
		Enabled             bool    `json:"enabled"`
		ConfidenceThreshold float64 `json:"confidence_threshold"`
		LLMFallback         bool    `json:"llm_fallback"`
		LogBufferSize       int     `json:"log_buffer_size"`
	}
	decodeJSON(t, w, &cfg)
	if !cfg.Enabled {
		t.Error("expected enabled=true")
	}
	if cfg.ConfidenceThreshold != 0.7 {
		t.Errorf("confidence_threshold = %f, want 0.7", cfg.ConfidenceThreshold)
	}

	// PUT config – update all fields
	newEnabled := false
	newThreshold := 0.8
	newLLM := true
	newBuf := 200
	w = doRequest(s, "PUT", "/api/v1/bot/skills/config", map[string]interface{}{
		"enabled":              newEnabled,
		"confidence_threshold": newThreshold,
		"llm_fallback":         newLLM,
		"log_buffer_size":      newBuf,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("PUT config expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify config was updated
	sc := config.GetSkillsConfig()
	if sc.ConfidenceThreshold != 0.8 {
		t.Errorf("after update, confidence_threshold = %f, want 0.8", sc.ConfidenceThreshold)
	}
	if sc.Enabled != false {
		t.Error("after update, expected enabled=false")
	}
	if sc.LLMFallback != true {
		t.Error("after update, expected llm_fallback=true")
	}
	if sc.LogBufferSize != 200 {
		t.Errorf("after update, log_buffer_size = %d, want 200", sc.LogBufferSize)
	}

	// PUT with invalid JSON
	badReq := doRequest(s, "PUT", "/api/v1/bot/skills/config", "not-json")
	if badReq.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", badReq.Code)
	}
}

func TestSkillAPITest(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// Test matching a message
	w := doRequest(s, "POST", "/api/v1/bot/skills/test", map[string]string{
		"message": "pause",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	decodeJSON(t, w, &result)
	if result["matched"] != true {
		t.Error("expected matched=true for 'pause'")
	}
	if result["skill"] != "process-control" {
		t.Errorf("skill = %v, want process-control", result["skill"])
	}

	// Test non-matching message
	w = doRequest(s, "POST", "/api/v1/bot/skills/test", map[string]string{
		"message": "random gibberish xyzzy",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	decodeJSON(t, w, &result)
	if result["matched"] != false {
		t.Error("expected matched=false for random gibberish")
	}

	// Test empty message
	w = doRequest(s, "POST", "/api/v1/bot/skills/test", map[string]string{
		"message": "",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty message, got %d", w.Code)
	}
}

func TestSkillAPILogs(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// First, generate some match logs via test endpoint
	doRequest(s, "POST", "/api/v1/bot/skills/test", map[string]string{
		"message": "pause",
	})
	doRequest(s, "POST", "/api/v1/bot/skills/test", map[string]string{
		"message": "hello world",
	})

	// Get logs
	w := doRequest(s, "GET", "/api/v1/bot/skills/logs", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var logs []map[string]interface{}
	decodeJSON(t, w, &logs)
	// Should have at least some logs from the test matches
	if len(logs) == 0 {
		t.Error("expected at least some match logs")
	}

	// Get logs with limit parameter
	w = doRequest(s, "GET", "/api/v1/bot/skills/logs?limit=1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for logs with limit, got %d", w.Code)
	}
	var limitedLogs []map[string]interface{}
	decodeJSON(t, w, &limitedLogs)
	if len(limitedLogs) > 1 {
		t.Errorf("expected at most 1 log with limit=1, got %d", len(limitedLogs))
	}
}

func TestSkillAPIMethodNotAllowed(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// DELETE on collection endpoint
	w := doRequest(s, "DELETE", "/api/v1/bot/skills", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}

	// GET on test endpoint
	w = doRequest(s, "GET", "/api/v1/bot/skills/test", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET on test endpoint, got %d", w.Code)
	}

	// POST on logs endpoint
	w = doRequest(s, "POST", "/api/v1/bot/skills/logs", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST on logs endpoint, got %d", w.Code)
	}

	// DELETE on config endpoint
	w = doRequest(s, "DELETE", "/api/v1/bot/skills/config", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for DELETE on config endpoint, got %d", w.Code)
	}
}

// --- T045: Frontend smoke tests ---

func TestSkillUIStaticFiles(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// Verify index.html is served (built by Vite from web/)
	w := doRequest(s, "GET", "/", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "GoZen") {
		t.Error("index.html should contain GoZen")
	}
}

func TestSkillUIAppJS(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// Vite build produces hashed JS in assets/, verify index.html references a JS bundle
	w := doRequest(s, "GET", "/", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<script") {
		t.Error("index.html should include a script tag")
	}
}

func TestSkillUIStyleCSS(t *testing.T) {
	s := setupTestServerWithSkills(t)

	// Vite build produces hashed CSS in assets/, verify index.html references a stylesheet
	w := doRequest(s, "GET", "/", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "stylesheet") {
		t.Error("index.html should include a stylesheet link")
	}
}

package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func TestAutoPermissionAPI_GetAll(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auto-permission", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/auto-permission status = %d, want 200", w.Code)
	}

	var resp autoPermissionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// Initially all should be nil
	if resp.Claude != nil {
		t.Errorf("claude should be nil initially, got %+v", resp.Claude)
	}
}

func TestAutoPermissionAPI_PutAndGet(t *testing.T) {
	s := setupTestServer(t)

	// PUT claude auto-permission
	body, _ := json.Marshal(autoPermissionRequest{
		Enabled: true,
		Mode:    "bypassPermissions",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/auto-permission/claude", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	// GET it back
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auto-permission/claude", nil)
	w = httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d", w.Code)
	}

	var ap config.AutoPermissionConfig
	if err := json.Unmarshal(w.Body.Bytes(), &ap); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !ap.Enabled {
		t.Error("enabled should be true")
	}
	if ap.Mode != "bypassPermissions" {
		t.Errorf("mode = %q, want %q", ap.Mode, "bypassPermissions")
	}
}

func TestAutoPermissionAPI_InvalidClient(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auto-permission/invalid-client", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("GET invalid client status = %d, want 400", w.Code)
	}
}

func TestAutoPermissionAPI_SettingsIncludesAutoPermission(t *testing.T) {
	s := setupTestServer(t)

	// Set auto-permission
	config.SetAutoPermission("claude", &config.AutoPermissionConfig{
		Enabled: true,
		Mode:    "acceptEdits",
	})

	// GET settings should include auto-permission
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/settings status = %d", w.Code)
	}

	var resp settingsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.ClaudeAutoPermission == nil {
		t.Fatal("claude_auto_permission should not be nil")
	}
	if resp.ClaudeAutoPermission.Mode != "acceptEdits" {
		t.Errorf("mode = %q, want %q", resp.ClaudeAutoPermission.Mode, "acceptEdits")
	}
}

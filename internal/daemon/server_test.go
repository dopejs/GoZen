package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDaemonStatusAPI(t *testing.T) {
	d := newTestDaemon()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/daemon/status", nil)
	d.handleDaemonStatus(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp daemonStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Status != "running" {
		t.Errorf("status = %q, want running", resp.Status)
	}
	if resp.Version != "test" {
		t.Errorf("version = %q, want test", resp.Version)
	}
}

func TestDaemonSessionsAPI(t *testing.T) {
	d := newTestDaemon()

	// Register a session
	d.RegisterSession("abc123", "default", "claude")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/daemon/sessions", nil)
	d.handleDaemonSessions(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Sessions []*SessionInfo `json:"sessions"`
		Count    int            `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Count != 1 {
		t.Errorf("count = %d, want 1", resp.Count)
	}
	if resp.Sessions[0].ID != "abc123" {
		t.Errorf("session ID = %q, want abc123", resp.Sessions[0].ID)
	}
}

func TestTempProfileAPI(t *testing.T) {
	d := newTestDaemon()

	// Create temp profile
	body := `{"providers":["p1","p2"]}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/profiles/temp", strings.NewReader(body))
	d.handleTempProfiles(w, r)

	// Should fail because providers don't exist in config
	// (no test config store set up)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for nonexistent provider, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSessionManagement(t *testing.T) {
	d := newTestDaemon()

	if d.ActiveSessionCount() != 0 {
		t.Fatalf("expected 0 sessions, got %d", d.ActiveSessionCount())
	}

	d.RegisterSession("s1", "default", "claude")
	d.RegisterSession("s2", "work", "codex")

	if d.ActiveSessionCount() != 2 {
		t.Fatalf("expected 2 sessions, got %d", d.ActiveSessionCount())
	}

	d.TouchSession("s1")
	d.RemoveSession("s2")

	if d.ActiveSessionCount() != 1 {
		t.Fatalf("expected 1 session, got %d", d.ActiveSessionCount())
	}
}

func TestTempProfileManagement(t *testing.T) {
	d := newTestDaemon()

	id := d.RegisterTempProfile([]string{"p1", "p2"})
	if !strings.HasPrefix(id, "_tmp_") {
		t.Errorf("temp profile ID should start with _tmp_, got %q", id)
	}

	tp := d.GetTempProfile(id)
	if tp == nil {
		t.Fatal("expected temp profile, got nil")
	}
	if len(tp.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(tp.Providers))
	}

	d.RemoveTempProfile(id)
	if d.GetTempProfile(id) != nil {
		t.Error("expected nil after removal")
	}
}

func newTestDaemon() *Daemon {
	d := NewDaemon("test", nil)
	d.startTime = d.startTime // already set by zero value, but Start() sets it
	d.proxyPort = 19841
	d.webPort = 19840
	return d
}

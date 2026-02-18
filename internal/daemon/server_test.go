package daemon

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
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

func TestDaemonStatusAPIMethodNotAllowed(t *testing.T) {
	d := newTestDaemon()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/daemon/status", nil)
	d.handleDaemonStatus(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestDaemonStatusWithSessions(t *testing.T) {
	d := newTestDaemon()
	d.RegisterSession("s1", "default", "claude")
	d.RegisterSession("s2", "work", "codex")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/daemon/status", nil)
	d.handleDaemonStatus(w, r)

	var resp daemonStatusResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ActiveSessions != 2 {
		t.Errorf("active_sessions = %d, want 2", resp.ActiveSessions)
	}
	if resp.ProxyPort != 19841 {
		t.Errorf("proxy_port = %d, want 19841", resp.ProxyPort)
	}
	if resp.WebPort != 19840 {
		t.Errorf("web_port = %d, want 19840", resp.WebPort)
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

func TestDaemonSessionsAPIMethodNotAllowed(t *testing.T) {
	d := newTestDaemon()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/daemon/sessions", nil)
	d.handleDaemonSessions(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestDaemonReloadAPI(t *testing.T) {
	d := newTestDaemon()

	// POST should succeed
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/daemon/reload", nil)
	d.handleDaemonReload(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// GET should fail
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/daemon/reload", nil)
	d.handleDaemonReload(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestTempProfileAPIValidation(t *testing.T) {
	d := newTestDaemon()

	// Method not allowed
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/profiles/temp", nil)
	d.handleTempProfiles(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}

	// Invalid JSON
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/v1/profiles/temp", strings.NewReader("not json"))
	d.handleTempProfiles(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}

	// Empty providers
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/api/v1/profiles/temp", strings.NewReader(`{"providers":[]}`))
	d.handleTempProfiles(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty providers, got %d", w.Code)
	}
}

func TestTempProfileDeleteAPI(t *testing.T) {
	d := newTestDaemon()
	id := d.RegisterTempProfile([]string{"p1"})

	// DELETE
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/profiles/temp/"+id, nil)
	d.handleTempProfile(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if d.GetTempProfile(id) != nil {
		t.Error("temp profile should be deleted")
	}

	// Empty ID
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/profiles/temp/", nil)
	d.handleTempProfile(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty ID, got %d", w.Code)
	}

	// Method not allowed
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/profiles/temp/someid", nil)
	d.handleTempProfile(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
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

func TestGetTempProfileProviders(t *testing.T) {
	d := newTestDaemon()

	// Non-existent profile
	if providers := d.GetTempProfileProviders("_tmp_nonexistent"); providers != nil {
		t.Errorf("expected nil for nonexistent, got %v", providers)
	}

	// Register and retrieve
	id := d.RegisterTempProfile([]string{"a", "b", "c"})
	providers := d.GetTempProfileProviders(id)
	if len(providers) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(providers))
	}
	if providers[0] != "a" || providers[1] != "b" || providers[2] != "c" {
		t.Errorf("providers = %v, want [a b c]", providers)
	}
}

func TestDaemonPidManagement(t *testing.T) {
	// Use a temp dir to avoid touching real config
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Write PID
	if err := WriteDaemonPid(12345); err != nil {
		t.Fatalf("WriteDaemonPid: %v", err)
	}

	// Read PID
	pid, err := ReadDaemonPid()
	if err != nil {
		t.Fatalf("ReadDaemonPid: %v", err)
	}
	if pid != 12345 {
		t.Errorf("pid = %d, want 12345", pid)
	}

	// Remove PID
	RemoveDaemonPid()
	_, err = ReadDaemonPid()
	if err == nil {
		t.Error("expected error after RemoveDaemonPid")
	}
}

func TestReadDaemonPidNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, err := ReadDaemonPid()
	if err == nil {
		t.Error("expected error for missing PID file")
	}
}

func TestCleanupLegacyPidFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	zenDir := filepath.Join(dir, ".zen")
	os.MkdirAll(zenDir, 0755)

	// Create legacy PID files
	os.WriteFile(filepath.Join(zenDir, "web.pid"), []byte("123\n"), 0600)
	os.WriteFile(filepath.Join(zenDir, "web-abcd1234.pid"), []byte("456\n"), 0600)
	os.WriteFile(filepath.Join(zenDir, "web-ef567890.pid"), []byte("789\n"), 0600)
	// This should NOT be removed
	os.WriteFile(filepath.Join(zenDir, "zend.pid"), []byte("999\n"), 0600)

	CleanupLegacyPidFiles()

	// Legacy files should be gone
	for _, name := range []string{"web.pid", "web-abcd1234.pid", "web-ef567890.pid"} {
		if _, err := os.Stat(filepath.Join(zenDir, name)); err == nil {
			t.Errorf("%s should have been removed", name)
		}
	}

	// zend.pid should still exist
	if _, err := os.Stat(filepath.Join(zenDir, "zend.pid")); err != nil {
		t.Error("zend.pid should NOT have been removed")
	}
}

func TestDaemonPidPath(t *testing.T) {
	path := DaemonPidPath()
	if !strings.HasSuffix(path, "zend.pid") {
		t.Errorf("DaemonPidPath = %q, want suffix zend.pid", path)
	}
}

func TestDaemonLogPath(t *testing.T) {
	path := DaemonLogPath()
	if !strings.HasSuffix(path, "zend.log") {
		t.Errorf("DaemonLogPath = %q, want suffix zend.log", path)
	}
}

func TestIsDaemonPortListening(t *testing.T) {
	// Port that's not listening
	if IsDaemonPortListening(59999) {
		t.Error("expected port 59999 to not be listening")
	}
}

func TestConfigWatcher(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	zenDir := filepath.Join(dir, ".zen")
	os.MkdirAll(zenDir, 0755)
	configPath := filepath.Join(zenDir, "zen.json")
	os.WriteFile(configPath, []byte(`{}`), 0600)

	reloaded := make(chan struct{}, 1)
	logger := log.New(os.Stderr, "[test] ", 0)
	w := NewConfigWatcher(logger, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})

	// Start watcher in background
	go w.Start()
	defer w.Stop()

	// Modify the config file
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(configPath, []byte(`{"version":7}`), 0600)

	// Manually trigger check (don't wait for ticker)
	w.check()

	select {
	case <-reloaded:
		// OK
	case <-time.After(time.Second):
		t.Error("expected reload callback after config change")
	}
}

func TestConfigWatcherNoFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	logger := log.New(os.Stderr, "[test] ", 0)
	w := NewConfigWatcher(logger, func() {
		t.Error("should not reload when file doesn't exist")
	})

	// check should not panic when file doesn't exist
	w.check()
}

func TestConfigWatcherStopIdempotent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	logger := log.New(os.Stderr, "[test] ", 0)
	w := NewConfigWatcher(logger, func() {})

	// Stop multiple times should not panic
	w.Stop()
	w.Stop()
}

func TestRandomID(t *testing.T) {
	id1 := randomID()
	id2 := randomID()

	if len(id1) == 0 {
		t.Error("randomID returned empty string")
	}
	// IDs should be different (extremely unlikely to collide)
	if id1 == id2 {
		t.Errorf("two randomID calls returned same value: %q", id1)
	}
}

func TestTouchSessionNonExistent(t *testing.T) {
	d := newTestDaemon()
	// Should not panic
	d.TouchSession("nonexistent")
}

func TestNewDaemon(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	d := NewDaemon("1.0.0", logger)
	if d == nil {
		t.Fatal("Expected non-nil daemon")
	}
	if d.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", d.version)
	}
	if d.sessions == nil {
		t.Error("Expected sessions map to be initialized")
	}
	if d.tmpProfiles == nil {
		t.Error("Expected tmpProfiles map to be initialized")
	}
}

func TestDaemonOnConfigReload(t *testing.T) {
	d := newTestDaemon()
	// Should not panic even without profileProxy
	d.onConfigReload()
}

func TestWriteDaemonPidError(t *testing.T) {
	// Test with invalid path
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/nonexistent/path/that/does/not/exist")
	defer os.Setenv("HOME", oldHome)

	err := WriteDaemonPid(12345)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestIsDaemonPortListeningWithServer(t *testing.T) {
	// Start a simple server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("Cannot create listener")
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	if !IsDaemonPortListening(port) {
		t.Error("Expected port to be listening")
	}
}

func newTestDaemon() *Daemon {
	d := NewDaemon("test", log.New(os.Stderr, "[test] ", 0))
	d.proxyPort = 19841
	d.webPort = 19840
	return d
}

func TestDaemonShutdown(t *testing.T) {
	d := newTestDaemon()

	// Set up some resources that Shutdown should clean up
	ctx, cancel := context.WithCancel(context.Background())
	d.syncCancel = cancel

	timer := time.NewTimer(time.Hour)
	d.pushTimer = timer

	// Create a mock watcher
	d.watcher = NewConfigWatcher(d.logger, func() {})

	// Shutdown should not panic and should clean up resources
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	err := d.Shutdown(shutdownCtx)
	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Good - context was cancelled
	default:
		t.Error("syncCancel should have been called")
	}
}

func TestDaemonShutdownNilResources(t *testing.T) {
	d := newTestDaemon()

	// Shutdown with nil resources should not panic
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := d.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown with nil resources returned error: %v", err)
	}
}

func TestSessionCleanupRemovesStale(t *testing.T) {
	d := newTestDaemon()

	// Add a fresh session
	d.RegisterSession("fresh", "default", "claude")

	// Add a stale session (manually set LastSeen to past)
	d.RegisterSession("stale", "default", "claude")
	d.mu.Lock()
	d.sessions["stale"].LastSeen = time.Now().Add(-3 * time.Hour)
	d.mu.Unlock()

	// Run cleanup manually (simulating what sessionCleanupLoop does)
	d.mu.Lock()
	now := time.Now()
	for id, s := range d.sessions {
		if now.Sub(s.LastSeen) > 2*time.Hour {
			delete(d.sessions, id)
		}
	}
	d.mu.Unlock()

	// Fresh session should remain
	if d.ActiveSessionCount() != 1 {
		t.Errorf("expected 1 session after cleanup, got %d", d.ActiveSessionCount())
	}

	// Verify it's the fresh one
	d.mu.RLock()
	_, hasFresh := d.sessions["fresh"]
	_, hasStale := d.sessions["stale"]
	d.mu.RUnlock()

	if !hasFresh {
		t.Error("fresh session should remain")
	}
	if hasStale {
		t.Error("stale session should be removed")
	}
}

func TestInitSyncCancelsExisting(t *testing.T) {
	d := newTestDaemon()

	// Set up an existing sync cancel function
	cancelled := false
	d.syncCancel = func() {
		cancelled = true
	}

	// Call initSync - it should cancel the existing one
	d.initSync()

	if !cancelled {
		t.Error("initSync should cancel existing syncCancel")
	}

	// syncCancel should be nil after initSync (no sync config)
	if d.syncCancel != nil {
		t.Error("syncCancel should be nil when no sync config")
	}
}

func TestDaemonStartProxy(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Create minimal config
	zenDir := filepath.Join(dir, ".zen")
	os.MkdirAll(zenDir, 0755)
	configPath := filepath.Join(zenDir, "zen.json")
	os.WriteFile(configPath, []byte(`{"providers":{},"profiles":{"default":{"providers":[]}}}`), 0600)

	d := newTestDaemon()

	// Find an available port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("Cannot find available port")
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	d.proxyPort = port

	// startProxy should work
	err = d.startProxy()
	if err != nil {
		t.Fatalf("startProxy failed: %v", err)
	}

	// Clean up
	if d.proxyServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		d.proxyServer.Shutdown(ctx)
		cancel()
	}
}

func TestDaemonResourceCleanupOnShutdown(t *testing.T) {
	d := newTestDaemon()

	// Track if resources are properly cleaned up
	var cleanupOrder []string
	var mu sync.Mutex

	// Set up sync cancel
	ctx, cancel := context.WithCancel(context.Background())
	d.syncCancel = func() {
		mu.Lock()
		cleanupOrder = append(cleanupOrder, "syncCancel")
		mu.Unlock()
		cancel()
	}

	// Set up push timer
	d.pushTimer = time.AfterFunc(time.Hour, func() {})

	// Set up watcher
	d.watcher = NewConfigWatcher(d.logger, func() {})

	// Shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	d.Shutdown(shutdownCtx)

	// Verify sync was cancelled
	select {
	case <-ctx.Done():
		// Good
	default:
		t.Error("sync context should be cancelled")
	}

	mu.Lock()
	if len(cleanupOrder) == 0 || cleanupOrder[0] != "syncCancel" {
		t.Error("syncCancel should be called during shutdown")
	}
	mu.Unlock()
}

func TestDaemonSysProcAttr(t *testing.T) {
	attr := DaemonSysProcAttr()
	if attr == nil {
		t.Fatal("DaemonSysProcAttr returned nil")
	}
	if !attr.Setsid {
		t.Error("Setsid should be true")
	}
}

func TestIsDaemonRunningNoPidFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	pid, running := IsDaemonRunning()
	if running {
		t.Error("should not be running without PID file")
	}
	if pid != 0 {
		t.Errorf("pid should be 0, got %d", pid)
	}
}

func TestIsDaemonRunningDeadProcess(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Write a PID that doesn't exist (very high PID)
	WriteDaemonPid(999999999)

	pid, running := IsDaemonRunning()
	if running {
		t.Error("should not be running with dead process PID")
	}
	if pid != 0 {
		t.Errorf("pid should be 0, got %d", pid)
	}

	// PID file should be cleaned up
	_, err := ReadDaemonPid()
	if err == nil {
		t.Error("PID file should be removed for dead process")
	}
}

func TestStopDaemonProcessNotRunning(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	err := StopDaemonProcess(time.Second)
	if err == nil {
		t.Error("should return error when daemon is not running")
	}
}

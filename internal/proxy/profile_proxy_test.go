package proxy

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type mockTempProfileProvider struct {
	profiles map[string][]string
}

func (m *mockTempProfileProvider) GetTempProfileProviders(id string) []string {
	return m.profiles[id]
}

func TestNewProfileProxy(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)
	if pp == nil {
		t.Fatal("NewProfileProxy returned nil")
	}
	if pp.cache == nil {
		t.Error("cache should be initialized")
	}
}

func TestProfileProxyInvalidPath(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid path, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "invalid_path" {
		t.Errorf("error type = %q, want invalid_path", errObj["type"])
	}
}

func TestProfileProxySingleSegmentPath(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/onlyone", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProfileProxyTempProfileNotSupported(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)
	// TempProfiles is nil

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/_tmp_abc/sess1/v1/messages", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProfileProxyTempProfileNotFound(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)
	pp.TempProfiles = &mockTempProfileProvider{profiles: map[string][]string{}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/_tmp_abc/sess1/v1/messages", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProfileProxyProfileNotFound(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	// Use a temp HOME so config store is empty
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/nonexistent/sess1/v1/messages", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProfileProxyInvalidateCache(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	// Manually populate cache
	pp.cache["test"] = &ProxyServer{}
	if len(pp.cache) != 1 {
		t.Fatal("cache should have 1 entry")
	}

	pp.InvalidateCache()
	if len(pp.cache) != 0 {
		t.Error("cache should be empty after InvalidateCache")
	}
}

func TestProfileProxyWriteError(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	w := httptest.NewRecorder()
	pp.writeError(w, http.StatusTeapot, "test_error", "test message")

	if w.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["type"] != "test_error" {
		t.Errorf("error type = %q, want test_error", errObj["type"])
	}
	if errObj["message"] != "test message" {
		t.Errorf("error message = %q, want 'test message'", errObj["message"])
	}
}

func TestProfileProxyGetOrCreateProxy(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	providers := []*Provider{
		{Name: "test", Healthy: true},
	}

	// First call creates
	srv1 := pp.getOrCreateProxy("prof1", providers, "anthropic")
	if srv1 == nil {
		t.Fatal("expected non-nil proxy server")
	}

	// Second call returns cached
	srv2 := pp.getOrCreateProxy("prof1", providers, "anthropic")
	if srv1 != srv2 {
		t.Error("expected same cached proxy server")
	}

	// Different profile creates new
	srv3 := pp.getOrCreateProxy("prof2", providers, "openai")
	if srv3 == srv1 {
		t.Error("expected different proxy server for different profile")
	}
}

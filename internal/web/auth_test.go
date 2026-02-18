package web

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func setupTestAuth(t *testing.T) (*Server, func()) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	configDir := filepath.Join(dir, config.ConfigDir)
	os.MkdirAll(configDir, 0755)
	cfg := &config.OpenCCConfig{
		Providers: map[string]*config.ProviderConfig{},
		Profiles:  map[string]*config.ProfileConfig{"default": {Providers: []string{}}},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, config.ConfigFile), data, 0600)

	config.DefaultStore()

	logger := log.New(io.Discard, "", 0)
	s := NewServer("test", logger, 0)
	cleanup := func() {}
	return s, cleanup
}

func TestAuthManagerCreateAndValidateSession(t *testing.T) {
	am := NewAuthManager()

	token := am.createSession()
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if !am.validateSession(token) {
		t.Error("session should be valid")
	}
	if am.validateSession("invalid-token") {
		t.Error("invalid token should not be valid")
	}
	if am.validateSession("") {
		t.Error("empty token should not be valid")
	}
}

func TestAuthManagerDeleteSession(t *testing.T) {
	am := NewAuthManager()
	token := am.createSession()
	am.deleteSession(token)
	if am.validateSession(token) {
		t.Error("deleted session should not be valid")
	}
}

func TestAuthManagerInvalidateAllSessions(t *testing.T) {
	am := NewAuthManager()
	t1 := am.createSession()
	t2 := am.createSession()
	am.invalidateAllSessions()
	if am.validateSession(t1) || am.validateSession(t2) {
		t.Error("all sessions should be invalidated")
	}
}

func TestAuthManagerSessionExpiry(t *testing.T) {
	am := NewAuthManager()
	token := am.createSession()

	// Manually set last access to past expiry
	am.mu.Lock()
	am.sessions[token] = time.Now().Add(-25 * time.Hour)
	am.mu.Unlock()

	if am.validateSession(token) {
		t.Error("expired session should not be valid")
	}
}

func TestAuthManagerCleanExpired(t *testing.T) {
	am := NewAuthManager()
	t1 := am.createSession()
	t2 := am.createSession()

	// Expire t1
	am.mu.Lock()
	am.sessions[t1] = time.Now().Add(-25 * time.Hour)
	am.mu.Unlock()

	am.CleanExpired()

	if am.validateSession(t1) {
		t.Error("expired session should be cleaned")
	}
	if !am.validateSession(t2) {
		t.Error("valid session should not be cleaned")
	}
}

func TestBruteForceProtection(t *testing.T) {
	am := NewAuthManager()
	ip := "1.2.3.4"

	// No delay initially
	if d := am.checkBruteForce(ip); d > 0 {
		t.Errorf("expected no delay, got %v", d)
	}

	// After 1 failure, 2s delay
	am.recordFailure(ip)
	if d := am.checkBruteForce(ip); d == 0 {
		t.Error("expected delay after failure")
	}

	// Reset clears
	am.resetFailures(ip)
	if d := am.checkBruteForce(ip); d > 0 {
		t.Errorf("expected no delay after reset, got %v", d)
	}
}

func TestIsLocalRequest(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       bool
	}{
		{"loopback ipv4", "127.0.0.1:12345", nil, true},
		{"loopback ipv6", "[::1]:12345", nil, true},
		{"remote", "192.168.1.1:12345", nil, false},
		{"forwarded local", "127.0.0.1:12345", map[string]string{"X-Forwarded-For": "1.2.3.4"}, false},
		{"real-ip local", "127.0.0.1:12345", map[string]string{"X-Real-IP": "1.2.3.4"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}
			if got := isLocalRequest(r); got != tt.want {
				t.Errorf("isLocalRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginEndpoint(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	// Set a password
	hash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)
	config.SetWebPasswordHash(string(hash))

	// Test successful login
	body := `{"password":"testpass"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "10.0.0.1:12345" // non-local
	w := httptest.NewRecorder()
	s.handleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("login got %d, want 200", w.Code)
	}
	// Should set cookie
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			found = true
			if !c.HttpOnly {
				t.Error("cookie should be HttpOnly")
			}
		}
	}
	if !found {
		t.Error("expected session cookie to be set")
	}

	// Test wrong password
	body = `{"password":"wrong"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "10.0.0.1:12345"
	w = httptest.NewRecorder()
	s.handleLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("wrong password got %d, want 401", w.Code)
	}

	// Test method not allowed
	req = httptest.NewRequest("GET", "/api/v1/auth/login", nil)
	w = httptest.NewRecorder()
	s.handleLogin(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET login got %d, want 405", w.Code)
	}
}

func TestLogoutEndpoint(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	token := s.auth.createSession()

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	w := httptest.NewRecorder()
	s.handleLogout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("logout got %d, want 200", w.Code)
	}
	if s.auth.validateSession(token) {
		t.Error("session should be invalidated after logout")
	}
}

func TestAuthMiddlewareLocalBypass(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)
	config.SetWebPasswordHash(string(hash))

	called := false
	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// Local request should bypass auth
	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called for local request")
	}
	if w.Code != http.StatusOK {
		t.Errorf("local request got %d, want 200", w.Code)
	}
}

func TestAuthMiddlewareRemoteBlocked(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)
	config.SetWebPasswordHash(string(hash))

	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Remote API request without session should get 401
	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("remote request without auth got %d, want 401", w.Code)
	}
}

func TestAuthMiddlewareRemoteWithSession(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)
	config.SetWebPasswordHash(string(hash))

	called := false
	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	token := s.auth.createSession()

	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called for authenticated remote request")
	}
}

func TestAuthMiddlewareNoPassword(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	// No password set — should allow all requests
	called := false
	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called when no password is configured")
	}
}

func TestPasswordChange(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	// Set initial password
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.MinCost)
	config.SetWebPasswordHash(string(hash))

	token := s.auth.createSession()

	// Change password
	body := `{"old_password":"oldpass","new_password":"newpass123"}`
	req := httptest.NewRequest("PUT", "/api/v1/settings/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	w := httptest.NewRecorder()
	s.handlePasswordChange(w, req)

	if w.Code != http.StatusOK {
		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		t.Errorf("password change got %d: %s", w.Code, resp["error"])
	}

	// Old session should be invalidated
	if s.auth.validateSession(token) {
		t.Error("session should be invalidated after password change")
	}

	// Verify new password works
	newHash := config.GetWebPasswordHash()
	if err := bcrypt.CompareHashAndPassword([]byte(newHash), []byte("newpass123")); err != nil {
		t.Error("new password should validate")
	}
}

func TestPasswordChangeWrongOld(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.MinCost)
	config.SetWebPasswordHash(string(hash))

	body := `{"old_password":"wrong","new_password":"newpass123"}`
	req := httptest.NewRequest("PUT", "/api/v1/settings/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePasswordChange(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("wrong old password got %d, want 401", w.Code)
	}
}

func TestPasswordChangeTooShort(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	body := `{"old_password":"","new_password":"abc"}`
	req := httptest.NewRequest("PUT", "/api/v1/settings/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handlePasswordChange(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("short password got %d, want 400", w.Code)
	}
}

func TestAuthCheckEndpoint(t *testing.T) {
	s, cleanup := setupTestAuth(t)
	defer cleanup()

	// No password set — should report authenticated
	req := httptest.NewRequest("GET", "/api/v1/auth/check", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	s.handleAuthCheck(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["authenticated"] != true {
		t.Error("should be authenticated when no password set")
	}
	if resp["password_required"] != false {
		t.Error("password should not be required when not set")
	}
}

func TestGeneratePassword(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	configDir := filepath.Join(dir, config.ConfigDir)
	os.MkdirAll(configDir, 0755)
	cfg := &config.OpenCCConfig{
		Providers: map[string]*config.ProviderConfig{},
		Profiles:  map[string]*config.ProfileConfig{"default": {Providers: []string{}}},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, config.ConfigFile), data, 0600)
	config.DefaultStore()

	password, err := GeneratePassword()
	if err != nil {
		t.Fatalf("GeneratePassword: %v", err)
	}
	if len(password) != 16 {
		t.Errorf("password length = %d, want 16", len(password))
	}

	// Verify hash was stored
	hash := config.GetWebPasswordHash()
	if hash == "" {
		t.Fatal("password hash should be stored")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Error("stored hash should match generated password")
	}
}

func TestSessionCleanupLoopAndStop(t *testing.T) {
	am := NewAuthManager()

	// Create a session
	token := am.createSession()
	if !am.validateSession(token) {
		t.Fatal("session should be valid")
	}

	// Start cleanup loop in background (it won't actually clean anything in this short test)
	go am.sessionCleanupLoop()

	// Stop the cleanup loop
	am.StopCleanup()

	// Session should still be valid (we didn't wait for cleanup)
	if !am.validateSession(token) {
		t.Error("session should still be valid after stopping cleanup")
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		realIP     string
		want       string
	}{
		{"remote addr only", "192.168.1.1:12345", "", "", "192.168.1.1"},
		{"x-forwarded-for single", "127.0.0.1:12345", "10.0.0.1", "", "10.0.0.1"},
		{"x-forwarded-for multiple", "127.0.0.1:12345", "10.0.0.1, 10.0.0.2", "", "10.0.0.1"},
		{"x-real-ip", "127.0.0.1:12345", "", "10.0.0.5", "10.0.0.5"},
		{"xff takes precedence", "127.0.0.1:12345", "10.0.0.1", "10.0.0.5", "10.0.0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.realIP != "" {
				r.Header.Set("X-Real-IP", tt.realIP)
			}
			if got := clientIP(r); got != tt.want {
				t.Errorf("clientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

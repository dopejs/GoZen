package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "zen_session"
	sessionMaxAge     = 24 * time.Hour
)

// AuthManager handles session-based authentication for the Web UI.
type AuthManager struct {
	mu       sync.RWMutex
	sessions map[string]time.Time // token -> last accessed

	failMu   sync.Mutex
	failures map[string]*loginFailure // IP -> failure info
}

type loginFailure struct {
	count    int
	lastFail time.Time
}

// NewAuthManager creates a new auth manager.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		sessions: make(map[string]time.Time),
		failures: make(map[string]*loginFailure),
	}
}

// GeneratePassword creates a random 16-character password and stores its bcrypt hash.
// Returns the plaintext password (for one-time display to the user).
func GeneratePassword() (string, error) {
	b := make([]byte, 12) // 12 bytes = 16 base64-ish chars
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Use hex for simplicity (24 chars, all printable)
	password := hex.EncodeToString(b)[:16]

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if err := config.SetWebPasswordHash(string(hash)); err != nil {
		return "", err
	}
	return password, nil
}

// HasPassword returns true if a password is configured.
func HasPassword() bool {
	return config.GetWebPasswordHash() != ""
}

// createSession generates a new session token and stores it.
func (am *AuthManager) createSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)

	am.mu.Lock()
	am.sessions[token] = time.Now()
	am.mu.Unlock()

	return token
}

// validateSession checks if a session token is valid and not expired.
func (am *AuthManager) validateSession(token string) bool {
	if token == "" {
		return false
	}
	am.mu.RLock()
	lastAccess, ok := am.sessions[token]
	am.mu.RUnlock()

	if !ok {
		return false
	}

	if time.Since(lastAccess) > sessionMaxAge {
		am.mu.Lock()
		delete(am.sessions, token)
		am.mu.Unlock()
		return false
	}

	// Refresh last access time
	am.mu.Lock()
	am.sessions[token] = time.Now()
	am.mu.Unlock()

	return true
}

// deleteSession removes a session token.
func (am *AuthManager) deleteSession(token string) {
	am.mu.Lock()
	delete(am.sessions, token)
	am.mu.Unlock()
}

// invalidateAllSessions removes all sessions (used after password change).
func (am *AuthManager) invalidateAllSessions() {
	am.mu.Lock()
	am.sessions = make(map[string]time.Time)
	am.mu.Unlock()
}

// CleanExpired removes expired sessions. Called periodically.
func (am *AuthManager) CleanExpired() {
	am.mu.Lock()
	defer am.mu.Unlock()
	now := time.Now()
	for token, lastAccess := range am.sessions {
		if now.Sub(lastAccess) > sessionMaxAge {
			delete(am.sessions, token)
		}
	}
}

// sessionCleanupLoop runs CleanExpired periodically.
func (am *AuthManager) sessionCleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		am.CleanExpired()
	}
}

// checkBruteForce returns the delay (in seconds) the client should wait before
// attempting login, based on past failures from this IP.
func (am *AuthManager) checkBruteForce(ip string) time.Duration {
	am.failMu.Lock()
	defer am.failMu.Unlock()

	f, ok := am.failures[ip]
	if !ok || f.count == 0 {
		return 0
	}

	// Reset after 10 minutes of no failures
	if time.Since(f.lastFail) > 10*time.Minute {
		delete(am.failures, ip)
		return 0
	}

	// Exponential backoff: 2^count seconds, max 30s
	delay := time.Duration(math.Min(math.Pow(2, float64(f.count)), 30)) * time.Second
	elapsed := time.Since(f.lastFail)
	if elapsed >= delay {
		return 0
	}
	return delay - elapsed
}

// recordFailure increments the failure counter for an IP.
func (am *AuthManager) recordFailure(ip string) {
	am.failMu.Lock()
	defer am.failMu.Unlock()

	f, ok := am.failures[ip]
	if !ok {
		f = &loginFailure{}
		am.failures[ip] = f
	}
	f.count++
	f.lastFail = time.Now()
}

// resetFailures clears the failure counter for an IP.
func (am *AuthManager) resetFailures(ip string) {
	am.failMu.Lock()
	delete(am.failures, ip)
	am.failMu.Unlock()
}

// isLocalRequest checks whether the request originates from localhost.
// Returns false if X-Forwarded-For or X-Real-IP headers are present (reverse proxy).
func isLocalRequest(r *http.Request) bool {
	if r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Real-IP") != "" {
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// clientIP extracts the client IP from the request.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return rip
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// authMiddleware returns an HTTP middleware that enforces authentication.
// Local requests are allowed through without authentication.
// The login and pubkey endpoints are always accessible.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always allow auth endpoints
		if r.URL.Path == "/api/v1/auth/login" ||
			r.URL.Path == "/api/v1/auth/pubkey" {
			next.ServeHTTP(w, r)
			return
		}

		// Local requests bypass auth
		if isLocalRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		// No password configured = no auth required
		if !HasPassword() {
			next.ServeHTTP(w, r)
			return
		}

		// Check session cookie
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil && s.auth.validateSession(cookie.Value) {
			next.ServeHTTP(w, r)
			return
		}

		// Not authenticated
		if isAPIRequest(r) {
			writeError(w, http.StatusUnauthorized, "authentication required")
		} else {
			// Serve login page for browser requests
			http.Redirect(w, r, "/login.html", http.StatusFound)
		}
	})
}

func isAPIRequest(r *http.Request) bool {
	return len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/"
}

// handleLogin handles POST /api/v1/auth/login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ip := clientIP(r)

	// Check brute force delay
	if delay := s.auth.checkBruteForce(ip); delay > 0 {
		w.Header().Set("Retry-After", time.Now().Add(delay).Format(time.RFC1123))
		writeError(w, http.StatusTooManyRequests, "too many login attempts, try again later")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	hash := config.GetWebPasswordHash()
	if hash == "" {
		writeError(w, http.StatusForbidden, "no password configured")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		s.auth.recordFailure(ip)
		writeError(w, http.StatusUnauthorized, "invalid password")
		return
	}

	// Success
	s.auth.resetFailures(ip)
	token := s.auth.createSession()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.Header.Get("X-Forwarded-Proto") == "https",
		MaxAge:   int(sessionMaxAge.Seconds()),
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleLogout handles POST /api/v1/auth/logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		s.auth.deleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handlePasswordChange handles PUT /api/v1/settings/password
func (s *Server) handlePasswordChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "new password is required")
		return
	}
	if len(req.NewPassword) < 6 {
		writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	// Verify old password (if one exists)
	hash := config.GetWebPasswordHash()
	if hash != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.OldPassword)); err != nil {
			writeError(w, http.StatusUnauthorized, "current password is incorrect")
			return
		}
	}

	// Set new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := config.SetWebPasswordHash(string(newHash)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save password")
		return
	}

	// Invalidate all sessions
	s.auth.invalidateAllSessions()

	writeJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
}

// handleAuthCheck handles GET /api/v1/auth/check — returns auth status.
func (s *Server) handleAuthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	authenticated := isLocalRequest(r)
	if !authenticated && HasPassword() {
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil {
			authenticated = s.auth.validateSession(cookie.Value)
		}
	} else if !HasPassword() {
		authenticated = true
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated":     authenticated,
		"password_required": HasPassword() && !isLocalRequest(r),
		"is_local":          isLocalRequest(r),
	})
}

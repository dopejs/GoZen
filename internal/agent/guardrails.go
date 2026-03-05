package agent

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// Guardrails provides safety controls for agent sessions.
type Guardrails struct {
	config     *config.GuardrailsConfig
	spending   map[string]float64          // session ID -> total spent
	requests   map[string][]time.Time      // session ID -> request timestamps
	operations []*SensitiveOperation       // recent sensitive operations
	mu         sync.RWMutex
}

// Global guardrails instance
var (
	globalGuardrails     *Guardrails
	globalGuardrailsOnce sync.Once
	globalGuardrailsMu   sync.RWMutex
)

// InitGlobalGuardrails initializes the global guardrails.
func InitGlobalGuardrails() {
	globalGuardrailsOnce.Do(func() {
		cfg := config.GetAgent()
		var grCfg *config.GuardrailsConfig
		if cfg != nil {
			grCfg = cfg.Guardrails
		}
		globalGuardrailsMu.Lock()
		globalGuardrails = NewGuardrails(grCfg)
		globalGuardrailsMu.Unlock()
	})
}

// GetGlobalGuardrails returns the global guardrails.
func GetGlobalGuardrails() *Guardrails {
	globalGuardrailsMu.RLock()
	defer globalGuardrailsMu.RUnlock()
	return globalGuardrails
}

// NewGuardrails creates a new guardrails instance.
func NewGuardrails(cfg *config.GuardrailsConfig) *Guardrails {
	if cfg == nil {
		cfg = &config.GuardrailsConfig{
			Enabled:            false,
			SessionSpendingCap: 10.0,  // $10 default
			RequestRateLimit:   60,    // 60 requests per minute
			SensitiveOpsDetect: true,
			AutoPauseOnCap:     true,
		}
	}
	return &Guardrails{
		config:     cfg,
		spending:   make(map[string]float64),
		requests:   make(map[string][]time.Time),
		operations: make([]*SensitiveOperation, 0),
	}
}

// IsEnabled returns whether guardrails are enabled.
func (g *Guardrails) IsEnabled() bool {
	return g.config != nil && g.config.Enabled
}

// UpdateConfig updates the guardrails configuration.
func (g *Guardrails) UpdateConfig(cfg *config.GuardrailsConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.config = cfg
}

// CheckRequest checks if a request should be allowed.
// Returns (allowed, reason).
func (g *Guardrails) CheckRequest(sessionID string) (bool, string) {
	if !g.IsEnabled() {
		return true, ""
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Check spending cap
	if g.config.SessionSpendingCap > 0 {
		spent := g.spending[sessionID]
		if spent >= g.config.SessionSpendingCap {
			if g.config.AutoPauseOnCap {
				return false, "session spending cap reached"
			}
		}
	}

	// Check rate limit
	if g.config.RequestRateLimit > 0 {
		now := time.Now()
		windowStart := now.Add(-time.Minute)

		// Clean old requests
		var recent []time.Time
		for _, t := range g.requests[sessionID] {
			if t.After(windowStart) {
				recent = append(recent, t)
			}
		}
		g.requests[sessionID] = recent

		if len(recent) >= g.config.RequestRateLimit {
			return false, "rate limit exceeded"
		}

		// Record this request
		g.requests[sessionID] = append(g.requests[sessionID], now)
	}

	return true, ""
}

// RecordSpending records spending for a session.
func (g *Guardrails) RecordSpending(sessionID string, cost float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.spending[sessionID] += cost
}

// GetSpending returns the total spending for a session.
func (g *Guardrails) GetSpending(sessionID string) float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.spending[sessionID]
}

// ResetSpending resets spending for a session.
func (g *Guardrails) ResetSpending(sessionID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.spending, sessionID)
}

// CheckSensitiveOperation checks if a request contains sensitive operations.
// Returns detected operations.
func (g *Guardrails) CheckSensitiveOperation(sessionID string, requestBody []byte) []*SensitiveOperation {
	if !g.IsEnabled() || !g.config.SensitiveOpsDetect {
		return nil
	}

	body := string(requestBody)
	var ops []*SensitiveOperation

	// Check for file deletion patterns
	deletePatterns := []string{
		`rm\s+-rf?\s+`,
		`unlink\s*\(`,
		`os\.Remove`,
		`fs\.unlink`,
		`shutil\.rmtree`,
		`delete\s+file`,
		`remove\s+directory`,
	}
	for _, pattern := range deletePatterns {
		if matched, _ := regexp.MatchString(pattern, body); matched {
			ops = append(ops, &SensitiveOperation{
				SessionID:   sessionID,
				Type:        SensitiveOpFileDelete,
				Description: "File deletion operation detected",
				Timestamp:   time.Now(),
			})
			break
		}
	}

	// Check for config file modifications
	configPatterns := []string{
		`\.env`,
		`config\.json`,
		`settings\.yaml`,
		`credentials`,
		`\.ssh/`,
		`\.aws/`,
		`\.kube/`,
	}
	for _, pattern := range configPatterns {
		if strings.Contains(strings.ToLower(body), pattern) {
			ops = append(ops, &SensitiveOperation{
				SessionID:   sessionID,
				Type:        SensitiveOpConfigModify,
				Description: "Configuration file access detected: " + pattern,
				Timestamp:   time.Now(),
			})
			break
		}
	}

	// Check for database operations
	dbPatterns := []string{
		`DROP\s+TABLE`,
		`DROP\s+DATABASE`,
		`TRUNCATE\s+TABLE`,
		`DELETE\s+FROM`,
		`db\.drop`,
		`migrate.*down`,
	}
	for _, pattern := range dbPatterns {
		if matched, _ := regexp.MatchString("(?i)"+pattern, body); matched {
			ops = append(ops, &SensitiveOperation{
				SessionID:   sessionID,
				Type:        SensitiveOpDBOperation,
				Description: "Database modification operation detected",
				Timestamp:   time.Now(),
			})
			break
		}
	}

	// Record operations
	if len(ops) > 0 {
		g.mu.Lock()
		g.operations = append(g.operations, ops...)
		// Keep only last 100 operations
		if len(g.operations) > 100 {
			g.operations = g.operations[len(g.operations)-100:]
		}
		g.mu.Unlock()
	}

	return ops
}

// GetRecentOperations returns recent sensitive operations.
func (g *Guardrails) GetRecentOperations(limit int) []*SensitiveOperation {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if limit <= 0 || limit > len(g.operations) {
		limit = len(g.operations)
	}

	result := make([]*SensitiveOperation, limit)
	copy(result, g.operations[len(g.operations)-limit:])
	return result
}

// GetSessionOperations returns sensitive operations for a specific session.
func (g *Guardrails) GetSessionOperations(sessionID string) []*SensitiveOperation {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []*SensitiveOperation
	for _, op := range g.operations {
		if op.SessionID == sessionID {
			result = append(result, op)
		}
	}
	return result
}

// GetAllSpending returns spending for all sessions.
func (g *Guardrails) GetAllSpending() map[string]float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make(map[string]float64, len(g.spending))
	for k, v := range g.spending {
		result[k] = v
	}
	return result
}

// GetConfig returns the current guardrails configuration.
func (g *Guardrails) GetConfig() *config.GuardrailsConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.config
}

package proxy

import (
	"testing"
	"time"
)

func TestSessionUsage(t *testing.T) {
	// Clear any existing data
	globalSessionCache = &SessionCache{
		maxSize: defaultMaxCacheSize,
	}

	// Test GetSessionUsage with empty session
	if usage := GetSessionUsage(""); usage != nil {
		t.Error("Expected nil for empty session ID")
	}

	// Test GetSessionUsage with non-existent session
	if usage := GetSessionUsage("nonexistent"); usage != nil {
		t.Error("Expected nil for non-existent session")
	}

	// Test UpdateSessionUsage
	usage := &SessionUsage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalCost:    0.05,
		TurnCount:    1,
	}
	UpdateSessionUsage("test-session", usage)

	// Test GetSessionUsage
	retrieved := GetSessionUsage("test-session")
	if retrieved == nil {
		t.Fatal("Expected to retrieve session usage")
	}
	if retrieved.InputTokens != 1000 {
		t.Errorf("Expected 1000 input tokens, got %d", retrieved.InputTokens)
	}

	// Test UpdateSessionUsage with nil
	UpdateSessionUsage("", nil)
	UpdateSessionUsage("test", nil)

	// Test ClearSessionUsage
	ClearSessionUsage("test-session")
	if GetSessionUsage("test-session") != nil {
		t.Error("Expected nil after clearing session")
	}

	// Test ClearSessionUsage with empty ID
	ClearSessionUsage("")
}

func TestSessionCacheEvictionLRU(t *testing.T) {
	// Create a small cache
	globalSessionCache = &SessionCache{
		maxSize: 3,
	}

	// Add sessions
	for i := 0; i < 5; i++ {
		usage := &SessionUsage{InputTokens: i * 100}
		UpdateSessionUsage("session-"+string(rune('a'+i)), usage)
	}

	// Check cache size
	size, maxSize := GetCacheStats()
	if size > maxSize {
		t.Errorf("Cache size %d exceeds max %d", size, maxSize)
	}
}

func TestCleanupOldSessionsCache(t *testing.T) {
	globalSessionCache = &SessionCache{
		maxSize: defaultMaxCacheSize,
	}

	// Add a session
	usage := &SessionUsage{
		InputTokens: 100,
		Timestamp:   time.Now().Add(-3 * time.Hour),
	}
	globalSessionCache.data.Store("old-session", usage)
	globalSessionCache.keyOrder = append(globalSessionCache.keyOrder, "old-session")

	// Add a recent session
	recentUsage := &SessionUsage{
		InputTokens: 200,
		Timestamp:   time.Now(),
	}
	globalSessionCache.data.Store("recent-session", recentUsage)
	globalSessionCache.keyOrder = append(globalSessionCache.keyOrder, "recent-session")

	// Cleanup sessions older than 2 hours
	deleted := CleanupOldSessions(2 * time.Hour)
	if deleted != 1 {
		t.Errorf("Expected 1 deleted session, got %d", deleted)
	}

	if GetSessionUsage("old-session") != nil {
		t.Error("Old session should have been cleaned up")
	}
}

func TestAddTurnToSession(t *testing.T) {
	globalSessionCache = &SessionCache{
		maxSize: defaultMaxCacheSize,
	}

	// Test with empty session ID
	AddTurnToSession("", TurnUsage{})

	// Add turn to new session
	turn := TurnUsage{
		InputTokens:  500,
		OutputTokens: 200,
		Cost:         0.02,
		Timestamp:    time.Now(),
	}
	AddTurnToSession("new-session", turn)

	usage := GetSessionUsage("new-session")
	if usage == nil {
		t.Fatal("Expected session to be created")
	}
	if usage.TurnCount != 1 {
		t.Errorf("Expected 1 turn, got %d", usage.TurnCount)
	}

	// Add more turns
	for i := 0; i < 25; i++ {
		AddTurnToSession("new-session", TurnUsage{
			InputTokens:  100 * (i + 1),
			OutputTokens: 50,
			Cost:         0.01,
			Timestamp:    time.Now(),
		})
	}

	usage = GetSessionUsage("new-session")
	if len(usage.Turns) > 20 {
		t.Errorf("Expected max 20 turns, got %d", len(usage.Turns))
	}
}

func TestGetSessionInsight(t *testing.T) {
	globalSessionCache = &SessionCache{
		maxSize: defaultMaxCacheSize,
	}

	// Test with empty session ID
	if insight := GetSessionInsight(""); insight != nil {
		t.Error("Expected nil for empty session ID")
	}

	// Test with non-existent session
	if insight := GetSessionInsight("nonexistent"); insight != nil {
		t.Error("Expected nil for non-existent session")
	}

	// Add session with turns
	now := time.Now()
	usage := &SessionUsage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalCost:    0.05,
		TurnCount:    2,
		Turns: []TurnUsage{
			{InputTokens: 400, OutputTokens: 200, Cost: 0.02, Timestamp: now.Add(-10 * time.Minute)},
			{InputTokens: 600, OutputTokens: 300, Cost: 0.03, Timestamp: now},
		},
	}
	globalSessionCache.data.Store("insight-session", usage)
	globalSessionCache.keyOrder = append(globalSessionCache.keyOrder, "insight-session")

	insight := GetSessionInsight("insight-session")
	if insight == nil {
		t.Fatal("Expected insight")
	}
	if insight.TotalInput != 1000 {
		t.Errorf("Expected 1000 total input, got %d", insight.TotalInput)
	}
	if insight.TurnCount != 2 {
		t.Errorf("Expected 2 turns, got %d", insight.TurnCount)
	}
	if insight.AvgOutputPerTurn != 250 {
		t.Errorf("Expected 250 avg output per turn, got %f", insight.AvgOutputPerTurn)
	}
}

func TestGetContextWarning(t *testing.T) {
	globalSessionCache = &SessionCache{
		maxSize: defaultMaxCacheSize,
	}

	// Test with empty session ID
	if warning := GetContextWarning("", 100000); warning != nil {
		t.Error("Expected nil for empty session ID")
	}

	// Test with non-existent session
	if warning := GetContextWarning("nonexistent", 100000); warning != nil {
		t.Error("Expected nil for non-existent session")
	}

	// Add session with low token count (no warning)
	lowUsage := &SessionUsage{InputTokens: 10000}
	globalSessionCache.data.Store("low-session", lowUsage)

	if warning := GetContextWarning("low-session", 100000); warning != nil {
		t.Error("Expected no warning for low token count")
	}

	// Add session with high token count (warning)
	highUsage := &SessionUsage{InputTokens: 85000}
	globalSessionCache.data.Store("high-session", highUsage)

	warning := GetContextWarning("high-session", 100000)
	if warning == nil {
		t.Fatal("Expected warning for high token count")
	}
	if warning.PercentUsed < 80 {
		t.Errorf("Expected percent used >= 80, got %f", warning.PercentUsed)
	}

	// Test with very high token count
	veryHighUsage := &SessionUsage{InputTokens: 95000}
	globalSessionCache.data.Store("very-high-session", veryHighUsage)

	warning = GetContextWarning("very-high-session", 100000)
	if warning == nil {
		t.Fatal("Expected warning for very high token count")
	}
	if warning.Warning != "Context is nearly full" {
		t.Errorf("Expected 'Context is nearly full', got %q", warning.Warning)
	}

	// Test with default threshold
	warning = GetContextWarning("high-session", 0)
	if warning == nil {
		t.Fatal("Expected warning with default threshold")
	}
}

func TestGetAllSessionInsights(t *testing.T) {
	globalSessionCache = &SessionCache{
		maxSize: defaultMaxCacheSize,
	}

	// Add multiple sessions
	for i := 0; i < 3; i++ {
		usage := &SessionUsage{
			InputTokens:  100 * (i + 1),
			OutputTokens: 50 * (i + 1),
			TurnCount:    i + 1,
		}
		sessionID := "session-" + string(rune('a'+i))
		globalSessionCache.data.Store(sessionID, usage)
		globalSessionCache.keyOrder = append(globalSessionCache.keyOrder, sessionID)
	}

	insights := GetAllSessionInsights()
	if len(insights) != 3 {
		t.Errorf("Expected 3 insights, got %d", len(insights))
	}
}

func TestExtractSessionIDFromBody(t *testing.T) {
	tests := []struct {
		name     string
		body     map[string]interface{}
		expected string
	}{
		{
			name:     "no metadata",
			body:     map[string]interface{}{},
			expected: "",
		},
		{
			name: "no user_id",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{},
			},
			expected: "",
		},
		{
			name: "invalid user_id format",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"user_id": "invalid",
				},
			},
			expected: "",
		},
		{
			name: "valid user_session format",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"user_id": "user_session_abc123",
				},
			},
			expected: "abc123",
		},
		{
			name: "user_id not string",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"user_id": 12345,
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSessionID(tt.body)
			if result != tt.expected {
				t.Errorf("extractSessionID() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "< 1 minute"},
		{1 * time.Minute, "1 minute"},
		{5 * time.Minute, "05 minutes"},
		{15 * time.Minute, "15 minutes"},
		{1 * time.Hour, "1 hour 00 minutes"},
		{2*time.Hour + 30*time.Minute, "02 hours"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
		}
	}
}

func TestGetCacheStats(t *testing.T) {
	globalSessionCache = &SessionCache{
		maxSize:  100,
		keyOrder: []string{"a", "b", "c"},
	}

	size, maxSize := GetCacheStats()
	if size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}
	if maxSize != 100 {
		t.Errorf("Expected maxSize 100, got %d", maxSize)
	}
}

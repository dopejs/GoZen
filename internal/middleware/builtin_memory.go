package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SessionMemoryConfig holds configuration for the session memory middleware.
type SessionMemoryConfig struct {
	Enabled       bool   `json:"enabled"`
	StoragePath   string `json:"storage_path,omitempty"`   // path to store memory data (default: ~/.zen/memory)
	MaxMemories   int    `json:"max_memories,omitempty"`   // max memories per project (default: 100)
	AutoExtract   bool   `json:"auto_extract,omitempty"`   // auto-extract insights from responses
	InjectContext bool   `json:"inject_context,omitempty"` // inject relevant memories into requests
	MaxInjectSize int    `json:"max_inject_size,omitempty"` // max characters to inject (default: 2000)
}

// MemoryEntry represents a stored memory/insight.
type MemoryEntry struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`      // "decision", "preference", "pattern", "todo", "context"
	Content   string    `json:"content"`
	Project   string    `json:"project"`
	SessionID string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
	Score     float64   `json:"score"` // relevance score
}

// SessionMemoryMiddleware provides cross-session intelligence by storing and retrieving
// insights from past conversations.
// [BETA] This feature is experimental.
type SessionMemoryMiddleware struct {
	config    SessionMemoryConfig
	memories  map[string][]*MemoryEntry // project -> memories
	mu        sync.RWMutex
	storePath string
}

// NewSessionMemory creates a new session memory middleware.
func NewSessionMemory() Middleware {
	return &SessionMemoryMiddleware{
		config: SessionMemoryConfig{
			MaxMemories:   100,
			AutoExtract:   true,
			InjectContext: true,
			MaxInjectSize: 2000,
		},
		memories: make(map[string][]*MemoryEntry),
	}
}

func (m *SessionMemoryMiddleware) Name() string {
	return "session-memory"
}

func (m *SessionMemoryMiddleware) Version() string {
	return "1.0.0"
}

func (m *SessionMemoryMiddleware) Description() string {
	return "Cross-session intelligence: stores and retrieves insights from past conversations"
}

func (m *SessionMemoryMiddleware) Priority() int {
	return 15 // After context-injection, before request-logger
}

func (m *SessionMemoryMiddleware) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := json.Unmarshal(config, &m.config); err != nil {
			return err
		}
	}

	// Set defaults
	if m.config.MaxMemories == 0 {
		m.config.MaxMemories = 100
	}
	if m.config.MaxInjectSize == 0 {
		m.config.MaxInjectSize = 2000
	}

	// Set storage path
	if m.config.StoragePath != "" {
		m.storePath = m.config.StoragePath
	} else {
		home, _ := os.UserHomeDir()
		m.storePath = filepath.Join(home, ".zen", "memory")
	}
	os.MkdirAll(m.storePath, 0755)

	// Load existing memories
	m.loadMemories()

	return nil
}

func (m *SessionMemoryMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
	if !m.config.InjectContext || ctx.ProjectPath == "" {
		return ctx, nil
	}

	// Get relevant memories for this project
	memories := m.getRelevantMemories(ctx.ProjectPath, ctx.Messages)
	if len(memories) == 0 {
		return ctx, nil
	}

	// Build memory context
	var memoryContext strings.Builder
	memoryContext.WriteString("\n[Session Memory - Relevant context from past conversations]\n")

	totalSize := 0
	for _, mem := range memories {
		entry := fmt.Sprintf("- [%s] %s\n", mem.Type, mem.Content)
		if totalSize+len(entry) > m.config.MaxInjectSize {
			break
		}
		memoryContext.WriteString(entry)
		totalSize += len(entry)
	}
	memoryContext.WriteString("[End Session Memory]\n")

	// Inject into first user message
	newCtx := ctx.Clone()
	for i, msg := range newCtx.Messages {
		if msg.Role == "user" {
			if content, ok := msg.Content.(string); ok {
				newCtx.Messages[i].Content = memoryContext.String() + content
			}
			break
		}
	}

	// Re-marshal body
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(newCtx.Body, &bodyMap); err == nil {
		bodyMap["messages"] = newCtx.Messages
		if newBody, err := json.Marshal(bodyMap); err == nil {
			newCtx.Body = newBody
		}
	}

	return newCtx, nil
}

func (m *SessionMemoryMiddleware) ProcessResponse(ctx *ResponseContext) (*ResponseContext, error) {
	if !m.config.AutoExtract || ctx.Request.ProjectPath == "" {
		return ctx, nil
	}

	// Extract insights from response
	go m.extractInsights(ctx)

	return ctx, nil
}

func (m *SessionMemoryMiddleware) Close() error {
	return m.saveMemories()
}

// getRelevantMemories returns memories relevant to the current context.
func (m *SessionMemoryMiddleware) getRelevantMemories(project string, messages []Message) []*MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	memories, ok := m.memories[project]
	if !ok || len(memories) == 0 {
		return nil
	}

	// Simple relevance: return most recent memories
	// TODO: Implement semantic similarity matching
	limit := 5
	if len(memories) < limit {
		limit = len(memories)
	}

	result := make([]*MemoryEntry, limit)
	copy(result, memories[:limit])
	return result
}

// extractInsights extracts insights from the response and stores them.
func (m *SessionMemoryMiddleware) extractInsights(ctx *ResponseContext) {
	// Parse response body
	var respData struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(ctx.Body, &respData); err != nil {
		return
	}

	if len(respData.Content) == 0 {
		return
	}

	text := respData.Content[0].Text

	// Simple pattern matching for insights
	// TODO: Use AI to extract insights
	insights := m.findInsights(text)

	for _, insight := range insights {
		m.addMemory(&MemoryEntry{
			ID:        m.generateID(insight.Content),
			Type:      insight.Type,
			Content:   insight.Content,
			Project:   ctx.Request.ProjectPath,
			SessionID: ctx.Request.SessionID,
			CreatedAt: time.Now(),
		})
	}
}

// findInsights finds insights in text using pattern matching.
func (m *SessionMemoryMiddleware) findInsights(text string) []struct{ Type, Content string } {
	var insights []struct{ Type, Content string }

	// Look for decision patterns
	decisionPatterns := []string{
		"I'll use", "I've decided", "Let's go with", "The approach will be",
		"I recommend", "We should", "The best option is",
	}
	for _, pattern := range decisionPatterns {
		if idx := strings.Index(strings.ToLower(text), strings.ToLower(pattern)); idx != -1 {
			// Extract sentence containing the pattern
			start := idx
			end := idx + 200
			if end > len(text) {
				end = len(text)
			}
			// Find sentence end
			for i := idx; i < end; i++ {
				if text[i] == '.' || text[i] == '\n' {
					end = i + 1
					break
				}
			}
			insights = append(insights, struct{ Type, Content string }{
				Type:    "decision",
				Content: strings.TrimSpace(text[start:end]),
			})
			break
		}
	}

	// Look for TODO patterns
	todoPatterns := []string{"TODO:", "FIXME:", "later we", "we'll need to", "remember to"}
	for _, pattern := range todoPatterns {
		if idx := strings.Index(strings.ToLower(text), strings.ToLower(pattern)); idx != -1 {
			start := idx
			end := idx + 150
			if end > len(text) {
				end = len(text)
			}
			for i := idx; i < end; i++ {
				if text[i] == '\n' {
					end = i
					break
				}
			}
			insights = append(insights, struct{ Type, Content string }{
				Type:    "todo",
				Content: strings.TrimSpace(text[start:end]),
			})
			break
		}
	}

	return insights
}

// addMemory adds a memory entry.
func (m *SessionMemoryMiddleware) addMemory(entry *MemoryEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	memories := m.memories[entry.Project]

	// Check for duplicates
	for _, existing := range memories {
		if existing.ID == entry.ID {
			return
		}
	}

	// Add to front (most recent first)
	memories = append([]*MemoryEntry{entry}, memories...)

	// Trim to max size
	if len(memories) > m.config.MaxMemories {
		memories = memories[:m.config.MaxMemories]
	}

	m.memories[entry.Project] = memories

	// Save asynchronously
	go m.saveMemories()
}

// generateID generates a unique ID for a memory entry.
func (m *SessionMemoryMiddleware) generateID(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:8])
}

// loadMemories loads memories from disk.
func (m *SessionMemoryMiddleware) loadMemories() {
	files, err := os.ReadDir(m.storePath)
	if err != nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(m.storePath, file.Name()))
		if err != nil {
			continue
		}

		var memories []*MemoryEntry
		if err := json.Unmarshal(data, &memories); err != nil {
			continue
		}

		project := strings.TrimSuffix(file.Name(), ".json")
		m.memories[project] = memories
	}
}

// saveMemories saves memories to disk.
func (m *SessionMemoryMiddleware) saveMemories() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for project, memories := range m.memories {
		if len(memories) == 0 {
			continue
		}

		data, err := json.MarshalIndent(memories, "", "  ")
		if err != nil {
			continue
		}

		// Use hash of project path as filename
		hash := sha256.Sum256([]byte(project))
		filename := hex.EncodeToString(hash[:8]) + ".json"

		if err := os.WriteFile(filepath.Join(m.storePath, filename), data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// GetMemories returns all memories for a project (for API access).
func (m *SessionMemoryMiddleware) GetMemories(project string) []*MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memories[project]
}

// ClearMemories clears all memories for a project.
func (m *SessionMemoryMiddleware) ClearMemories(project string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.memories, project)
	go m.saveMemories()
}

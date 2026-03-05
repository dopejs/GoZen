package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/bot"
	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// ChatSession stores conversation history for a chat session.
type ChatSession struct {
	ID        string
	Messages  []bot.ChatMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// chatSessionStore manages in-memory chat sessions.
type chatSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*ChatSession
}

var globalChatSessions = &chatSessionStore{
	sessions: make(map[string]*ChatSession),
}

func (s *chatSessionStore) Get(id string) *ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

func (s *chatSessionStore) Create() *ChatSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("chat-%d", time.Now().UnixNano())
	session := &ChatSession{
		ID:        id,
		Messages:  []bot.ChatMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.sessions[id] = session
	return session
}

func (s *chatSessionStore) AddMessage(id string, msg bot.ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[id]; ok {
		session.Messages = append(session.Messages, msg)
		session.UpdatedAt = time.Now()
	}
}

func (s *chatSessionStore) Clear(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[id]; ok {
		session.Messages = []bot.ChatMessage{}
		session.UpdatedAt = time.Now()
	}
}

func (s *chatSessionStore) Cleanup(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, session := range s.sessions {
		if session.UpdatedAt.Before(cutoff) {
			delete(s.sessions, id)
		}
	}
}

// StartChatSessionCleanup starts a goroutine that cleans up old sessions.
func StartChatSessionCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				globalChatSessions.Cleanup(1 * time.Hour)
			}
		}
	}()
}

type chatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	Clear     bool   `json:"clear,omitempty"`
}

func (s *Server) handleBotChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req chatRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get or create session
	var session *ChatSession
	if req.SessionID != "" {
		session = globalChatSessions.Get(req.SessionID)
	}
	if session == nil {
		session = globalChatSessions.Create()
	}

	// Handle clear request
	if req.Clear {
		globalChatSessions.Clear(session.ID)
		writeJSON(w, http.StatusOK, map[string]string{
			"session_id": session.ID,
			"status":     "cleared",
		})
		return
	}

	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	// Get bot config
	store := config.DefaultStore()
	botCfg := store.GetBot()
	if botCfg == nil || botCfg.Profile == "" {
		writeError(w, http.StatusBadRequest, "bot not configured - set a profile first")
		return
	}

	proxyPort := store.GetProxyPort()
	if proxyPort == 0 {
		proxyPort = 19841
	}

	// Add user message to session
	userMsg := bot.ChatMessage{Role: "user", Content: req.Message}
	globalChatSessions.AddMessage(session.ID, userMsg)

	// Build messages for LLM (get fresh copy after adding)
	currentSession := globalChatSessions.Get(session.ID)
	messages := make([]bot.ChatMessage, len(currentSession.Messages))
	copy(messages, currentSession.Messages)

	// Build system prompt with process info from bot bridge
	var processes []*bot.ProcessInfo
	if bridge := proxy.GetBotBridge(); bridge != nil {
		processes = bridge.GetProcessInfo()
	}
	memory, _ := bot.LoadMemory(bot.MemoryDir())
	systemPrompt := bot.BuildSystemPrompt(processes, botCfg.Profile, memory)

	// Set up SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	// Send session ID
	sendSSE(w, flusher, "session", map[string]string{"session_id": session.ID})

	// Create LLM client and stream response
	llm := bot.NewLLMClient(proxyPort, botCfg.Profile, botCfg.Model)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	var fullResponse string
	err := llm.ChatStream(ctx, systemPrompt, messages, func(delta string) {
		fullResponse += delta
		sendSSE(w, flusher, "delta", map[string]string{"content": delta})
	})

	if err != nil {
		sendSSE(w, flusher, "error", map[string]string{"error": err.Error()})
		return
	}

	// Add assistant response to session
	assistantMsg := bot.ChatMessage{Role: "assistant", Content: fullResponse}
	globalChatSessions.AddMessage(session.ID, assistantMsg)

	sendSSE(w, flusher, "done", map[string]string{"content": fullResponse})
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
	flusher.Flush()
}

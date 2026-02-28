package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- LLM Chat/ChatStream with mock HTTP server ---

func TestLLMClient_Chat_Success(t *testing.T) {
	// Mock Anthropic API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": "Hello from LLM!"},
			},
		})
	}))
	defer server.Close()

	// Parse port from test server URL
	port := serverPort(t, server)

	client := NewLLMClient(port, "default", "test-model")
	client.client = server.Client()

	// Override the URL by using the test server's port
	history := []ChatMessage{{Role: "user", Content: "hello"}}
	resp, err := client.Chat(context.Background(), "system prompt", history)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	if resp != "Hello from LLM!" {
		t.Errorf("expected 'Hello from LLM!', got %q", resp)
	}
}

func TestLLMClient_Chat_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	port := serverPort(t, server)
	client := NewLLMClient(port, "default", "test-model")

	history := []ChatMessage{{Role: "user", Content: "hello"}}
	_, err := client.Chat(context.Background(), "system", history)
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code, got: %v", err)
	}
}

func TestLLMClient_Chat_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]string{},
		})
	}))
	defer server.Close()

	port := serverPort(t, server)
	client := NewLLMClient(port, "default", "test-model")

	history := []ChatMessage{{Role: "user", Content: "hello"}}
	_, err := client.Chat(context.Background(), "system", history)
	if err == nil {
		t.Fatal("expected error for empty response")
	}
	if !strings.Contains(err.Error(), "empty response") {
		t.Errorf("expected 'empty response' error, got: %v", err)
	}
}

func TestLLMClient_ChatStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		events := []string{
			`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}`,
			`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":" world"}}`,
			`data: [DONE]`,
		}
		for _, e := range events {
			fmt.Fprintln(w, e)
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	defer server.Close()

	port := serverPort(t, server)
	client := NewLLMClient(port, "default", "test-model")

	var chunks []string
	history := []ChatMessage{{Role: "user", Content: "hello"}}
	err := client.ChatStream(context.Background(), "system", history, func(delta string) {
		chunks = append(chunks, delta)
	})
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0] != "Hello" || chunks[1] != " world" {
		t.Errorf("unexpected chunks: %v", chunks)
	}
}

func TestLLMClient_ChatStream_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	port := serverPort(t, server)
	client := NewLLMClient(port, "default", "test-model")

	history := []ChatMessage{{Role: "user", Content: "hello"}}
	err := client.ChatStream(context.Background(), "system", history, func(delta string) {})
	if err == nil {
		t.Fatal("expected error for 400 status")
	}
}

// --- handleChat with LLM ---

func TestGateway_handleChat_WithLLM(t *testing.T) {
	// Start mock LLM server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": "I'm Zen, your assistant!"},
			},
		})
	}))
	defer server.Close()

	port := serverPort(t, server)

	g := newTestGateway()
	g.config.Profile = "default"
	g.config.ProxyPort = port
	g.config.MemoryDir = t.TempDir()
	g.llm = NewLLMClient(port, "default", "test-model")

	adapter := newMockAdapter(PlatformTelegram)
	g.adapters = append(g.adapters, adapter)

	session := g.sessions.GetOrCreate(PlatformTelegram, "user-1", "chat-1")
	intent := &ParsedIntent{Intent: IntentChat, Raw: "hello"}
	replyTo := ReplyContext{Platform: PlatformTelegram, ChatID: "chat-1"}

	g.handleChat(intent, session, replyTo)

	if len(adapter.sentMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(adapter.sentMessages))
	}
	if !strings.Contains(adapter.sentMessages[0].Text, "Zen") {
		t.Errorf("expected response containing 'Zen', got: %s", adapter.sentMessages[0].Text)
	}
}

func TestGateway_handleChat_LLMError(t *testing.T) {
	// Start mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	port := serverPort(t, server)

	g := newTestGateway()
	g.config.Profile = "default"
	g.config.ProxyPort = port
	g.config.MemoryDir = t.TempDir()
	g.llm = NewLLMClient(port, "default", "test-model")

	adapter := newMockAdapter(PlatformTelegram)
	g.adapters = append(g.adapters, adapter)

	session := g.sessions.GetOrCreate(PlatformTelegram, "user-1", "chat-1")
	intent := &ParsedIntent{Intent: IntentChat, Raw: "hello"}
	replyTo := ReplyContext{Platform: PlatformTelegram, ChatID: "chat-1"}

	g.handleChat(intent, session, replyTo)

	// Should get fallback response on error
	if len(adapter.sentMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(adapter.sentMessages))
	}
}

func TestGateway_handleChat_WithTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": "Done!"},
			},
		})
	}))
	defer server.Close()

	port := serverPort(t, server)
	g := newTestGateway()
	g.config.Profile = "default"
	g.config.ProxyPort = port
	g.config.MemoryDir = t.TempDir()
	g.llm = NewLLMClient(port, "default", "test-model")

	adapter := newMockAdapter(PlatformTelegram)
	g.adapters = append(g.adapters, adapter)

	session := g.sessions.GetOrCreate(PlatformTelegram, "user-1", "chat-1")
	// Intent with Task field set (takes priority over Raw)
	intent := &ParsedIntent{Intent: IntentChat, Raw: "original", Task: "specific task"}
	replyTo := ReplyContext{Platform: PlatformTelegram, ChatID: "chat-1"}

	g.handleChat(intent, session, replyTo)

	if len(adapter.sentMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(adapter.sentMessages))
	}
}

// --- UpdateStatusExtended ---

func TestClient_UpdateStatusExtended_NotConnected(t *testing.T) {
	client := NewClient("/path/to/project", "")
	err := client.UpdateStatusExtended(StatusUpdate{
		Status:      "busy",
		CurrentTask: "running tests",
	})
	if err == nil {
		t.Error("UpdateStatusExtended should fail when not connected")
	}
}

func TestClient_UpdateStatusExtended_TruncatesLongMessage(t *testing.T) {
	client := NewClient("/path/to/project", "")

	longMsg := strings.Repeat("x", 300)
	update := StatusUpdate{
		Status:      "busy",
		LastMessage: longMsg,
	}

	// Call will fail (not connected), but we can verify truncation
	// by checking the cached status after the call
	_ = client.UpdateStatusExtended(update)

	client.mu.Lock()
	cached := client.currentStatus
	client.mu.Unlock()

	if len(cached.LastMessage) > 204 { // 200 + "..."
		t.Errorf("expected truncated message, got length %d", len(cached.LastMessage))
	}
	if !strings.HasSuffix(cached.LastMessage, "...") {
		t.Error("expected truncated message to end with '...'")
	}
}

// --- formatProcessList edge cases ---

func TestFormatProcessList_WithAllFields(t *testing.T) {
	processes := []*ProcessInfo{
		{
			Name:          "api",
			Path:          "/path/to/api",
			Status:        "busy",
			CurrentTask:   "deploying",
			WaitingFor:    "approval",
			PendingAction: "confirm deploy",
			LastMessage:   "Waiting for user confirmation",
			MessageRole:   "assistant",
			TurnCount:     5,
			StartTime:     time.Now().Add(-1 * time.Hour),
		},
	}
	result := formatProcessList(processes)
	if !strings.Contains(result, "waiting for: approval") {
		t.Error("should contain waiting for")
	}
	if !strings.Contains(result, "pending action: confirm deploy") {
		t.Error("should contain pending action")
	}
	if !strings.Contains(result, "last assistant message") {
		t.Error("should contain last message with role")
	}
	if !strings.Contains(result, "turns: 5") {
		t.Error("should contain turn count")
	}
}

func TestFormatProcessList_LongLastMessage(t *testing.T) {
	longMsg := strings.Repeat("a", 150)
	processes := []*ProcessInfo{
		{
			Name:        "api",
			Path:        "/p",
			Status:      "idle",
			LastMessage: longMsg,
			StartTime:   time.Now(),
		},
	}
	result := formatProcessList(processes)
	if !strings.Contains(result, "...") {
		t.Error("long message should be truncated with ...")
	}
}

func TestFormatProcessList_EmptyMessageRole(t *testing.T) {
	processes := []*ProcessInfo{
		{
			Name:        "api",
			Path:        "/p",
			Status:      "idle",
			LastMessage: "hello",
			MessageRole: "",
			StartTime:   time.Now(),
		},
	}
	result := formatProcessList(processes)
	if !strings.Contains(result, "last unknown message") {
		t.Error("empty role should default to 'unknown'")
	}
}

// --- helper ---

func serverPort(t *testing.T, server *httptest.Server) int {
	t.Helper()
	addr := server.Listener.Addr().String()
	var port int
	fmt.Sscanf(addr[strings.LastIndex(addr, ":")+1:], "%d", &port)
	return port
}

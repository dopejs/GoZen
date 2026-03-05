package bot

import (
	"testing"
	"time"
)

func TestNewLLMClient(t *testing.T) {
	client := NewLLMClient(19841, "default", "")
	if client == nil {
		t.Fatal("NewLLMClient returned nil")
	}
	if client.proxyPort != 19841 {
		t.Errorf("expected proxyPort 19841, got %d", client.proxyPort)
	}
	if client.profile != "default" {
		t.Errorf("expected profile 'default', got %s", client.profile)
	}
	if client.client == nil {
		t.Error("HTTP client should not be nil")
	}
}

func TestBuildSystemPrompt_NoProcesses(t *testing.T) {
	prompt := BuildSystemPrompt(nil, "default", "")
	if prompt == "" {
		t.Error("BuildSystemPrompt should not return empty string")
	}
	if !contains(prompt, "No connected sessions") {
		t.Error("prompt should contain 'No connected sessions' when no processes")
	}
	if !contains(prompt, "default") {
		t.Error("prompt should contain profile name")
	}
}

func TestBuildSystemPrompt_WithProcesses(t *testing.T) {
	processes := []*ProcessInfo{
		{Name: "api", Path: "/path/to/api", Status: "idle", StartTime: time.Now()},
		{Name: "backend", Path: "/path/to/backend", Status: "busy", CurrentTask: "running tests", StartTime: time.Now()},
		{Name: "frontend", Alias: "web", Path: "/path/to/frontend", Status: "idle", StartTime: time.Now()},
	}
	prompt := BuildSystemPrompt(processes, "myprofile", "")
	if prompt == "" {
		t.Error("BuildSystemPrompt should not return empty string")
	}
	if !contains(prompt, "api") {
		t.Error("prompt should contain process name 'api'")
	}
	if !contains(prompt, "backend") {
		t.Error("prompt should contain process name 'backend'")
	}
	// frontend has alias "web", so it should show "web" not "frontend"
	if !contains(prompt, "web") {
		t.Error("prompt should contain alias 'web' for frontend")
	}
	if !contains(prompt, "myprofile") {
		t.Error("prompt should contain profile name")
	}
	if !contains(prompt, "running tests") {
		t.Error("prompt should contain current task")
	}
}

func TestBuildSystemPrompt_SingleProcess(t *testing.T) {
	processes := []*ProcessInfo{
		{Name: "myproject", Path: "/path/to/myproject", Status: "idle", StartTime: time.Now()},
	}
	prompt := BuildSystemPrompt(processes, "test", "")
	if !contains(prompt, "myproject") {
		t.Error("prompt should contain single process name")
	}
}

func TestBuildSystemPrompt_WithMemory(t *testing.T) {
	prompt := BuildSystemPrompt(nil, "default", "你是一个猫娘助手，说话要加喵~")
	if !contains(prompt, "猫娘") {
		t.Error("prompt should contain memory persona")
	}
	// Memory is appended as persona instructions; base identity is always present
	if !contains(prompt, "You are Zen") {
		t.Error("prompt should always contain base identity")
	}
	if !contains(prompt, "Persona Instructions") {
		t.Error("prompt should contain persona instructions section when memory is set")
	}
}

func TestBuildSystemPrompt_WithoutMemory(t *testing.T) {
	prompt := BuildSystemPrompt(nil, "default", "")
	if !contains(prompt, "You are Zen") {
		t.Error("prompt should contain default base when no memory")
	}
}

func TestBuildSystemPrompt_ProcessDetails(t *testing.T) {
	processes := []*ProcessInfo{
		{Name: "api", Path: "/path/to/api", Status: "busy", CurrentTask: "deploying", StartTime: time.Now().Add(-2 * time.Hour)},
	}
	prompt := BuildSystemPrompt(processes, "default", "")
	if !contains(prompt, "busy") {
		t.Error("prompt should contain process status")
	}
	if !contains(prompt, "/path/to/api") {
		t.Error("prompt should contain process path")
	}
	if !contains(prompt, "deploying") {
		t.Error("prompt should contain current task")
	}
}

func TestGateway_handleChat_NoLLM(t *testing.T) {
	g := newTestGateway()
	adapter := newMockAdapter(PlatformTelegram)
	g.adapters = append(g.adapters, adapter)

	// Gateway without LLM client
	g.llm = nil

	replyTo := ReplyContext{
		Platform: PlatformTelegram,
		ChatID:   "chat-1",
	}

	session := g.sessions.GetOrCreate(PlatformTelegram, "user-1", "chat-1")
	intent := &ParsedIntent{Intent: IntentChat, Raw: "hello"}

	g.handleChat(intent, session, replyTo)

	if len(adapter.sentMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(adapter.sentMessages))
	}

	// Should get fallback help message
	msg := adapter.sentMessages[0]
	if !contains(msg.Text, "Zen") {
		t.Errorf("expected message to mention Zen, got: %s", msg.Text)
	}
	if msg.Format != "markdown" {
		t.Errorf("expected markdown format, got %s", msg.Format)
	}
}

func TestGateway_processIntent_Chat(t *testing.T) {
	g := newTestGateway()
	adapter := newMockAdapter(PlatformTelegram)
	g.adapters = append(g.adapters, adapter)

	session := g.sessions.GetOrCreate(PlatformTelegram, "user-1", "chat-1")

	replyTo := ReplyContext{
		Platform: PlatformTelegram,
		ChatID:   "chat-1",
	}

	intent := &ParsedIntent{Intent: IntentChat, Raw: "what can you do?"}
	msg := &Message{
		Platform: PlatformTelegram,
		ChatID:   "chat-1",
		UserID:   "user-1",
	}

	g.processIntent(intent, session, replyTo, msg)

	if len(adapter.sentMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(adapter.sentMessages))
	}

	// Should get a response (fallback since no LLM configured)
	if adapter.sentMessages[0].Text == "" {
		t.Error("expected non-empty response")
	}
}

func TestNLUParser_Parse_ChatFallback(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// Random text that doesn't match any command should go to chat
	tests := []struct {
		content string
	}{
		{"hello there"},
		{"what's the weather like"},
		{"tell me a joke"},
		{"how are you"},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != IntentChat {
			t.Errorf("Parse(%q) = %v, want IntentChat", tt.content, result.Intent)
		}
	}
}

func TestNLUParser_Parse_ExplicitSendTask(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// Test explicit "send <target> <task>" syntax
	msg := &Message{Content: "send api run all tests", IsMention: true}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Fatal("Parse returned nil")
	}
	if result.Intent != IntentSendTask {
		t.Errorf("expected IntentSendTask, got %v", result.Intent)
	}
	if result.Target != "api" {
		t.Errorf("expected target 'api', got %s", result.Target)
	}
	if result.Task != "run all tests" {
		t.Errorf("expected task 'run all tests', got %s", result.Task)
	}
}

func TestNLUParser_Parse_ColonSyntax(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// Test "<target>: <task>" syntax
	msg := &Message{Content: "backend: deploy to staging", IsMention: true}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Fatal("Parse returned nil")
	}
	if result.Intent != IntentSendTask {
		t.Errorf("expected IntentSendTask, got %v", result.Intent)
	}
	if result.Target != "backend" {
		t.Errorf("expected target 'backend', got %s", result.Target)
	}
	if result.Task != "deploy to staging" {
		t.Errorf("expected task 'deploy to staging', got %s", result.Task)
	}
}

func TestGateway_NewGateway_WithLLM(t *testing.T) {
	config := &GatewayConfig{
		Enabled:   true,
		Profile:   "default",
		ProxyPort: 19841,
	}

	g := NewGateway(config, nil)
	if g.llm == nil {
		t.Error("LLM client should be initialized when profile and proxy port are set")
	}
}

func TestGateway_NewGateway_WithoutLLM(t *testing.T) {
	config := &GatewayConfig{
		Enabled: true,
		// No profile or proxy port
	}

	g := NewGateway(config, nil)
	if g.llm != nil {
		t.Error("LLM client should be nil when profile or proxy port not set")
	}
}

func TestGateway_NewGateway_PartialLLMConfig(t *testing.T) {
	// Only profile, no proxy port
	config1 := &GatewayConfig{
		Enabled: true,
		Profile: "default",
	}
	g1 := NewGateway(config1, nil)
	if g1.llm != nil {
		t.Error("LLM client should be nil when proxy port not set")
	}

	// Only proxy port, no profile
	config2 := &GatewayConfig{
		Enabled:   true,
		ProxyPort: 19841,
	}
	g2 := NewGateway(config2, nil)
	if g2.llm != nil {
		t.Error("LLM client should be nil when profile not set")
	}
}

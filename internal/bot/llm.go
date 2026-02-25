package bot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultBotModel = "claude-3-haiku-20240307"

// LLMClient handles LLM calls through the proxy.
type LLMClient struct {
	proxyPort int
	profile   string
	model     string
	client    *http.Client
}

// NewLLMClient creates a new LLM client.
func NewLLMClient(proxyPort int, profile string, model string) *LLMClient {
	if model == "" {
		model = DefaultBotModel
	}
	return &LLMClient{
		proxyPort: proxyPort,
		profile:   profile,
		model:     model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat sends a multi-turn conversation to the LLM and returns the response.
func (c *LLMClient) Chat(ctx context.Context, systemPrompt string, history []ChatMessage) (string, error) {
	messages := make([]map[string]string, len(history))
	for i, msg := range history {
		messages[i] = map[string]string{"role": msg.Role, "content": msg.Content}
	}

	reqBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 1024,
		"system":     systemPrompt,
		"messages":   messages,
	}

	return c.sendRequest(ctx, reqBody)
}

// sendRequest sends a request to the LLM via the proxy.
func (c *LLMClient) sendRequest(ctx context.Context, reqBody map[string]interface{}) (string, error) {
	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	sessionID := fmt.Sprintf("bot-%d", time.Now().UnixNano())
	url := fmt.Sprintf("http://127.0.0.1:%d/%s/%s/v1/messages", c.proxyPort, c.profile, sessionID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var respData struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(respData.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}

	// Extract text from all content blocks
	var result string
	for _, block := range respData.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	return result, nil
}

// StreamCallback is called for each chunk of streamed response.
type StreamCallback func(delta string)

// ChatStream sends a streaming chat request and calls the callback for each chunk.
func (c *LLMClient) ChatStream(ctx context.Context, systemPrompt string, history []ChatMessage, callback StreamCallback) error {
	messages := make([]map[string]string, len(history))
	for i, msg := range history {
		messages[i] = map[string]string{"role": msg.Role, "content": msg.Content}
	}

	reqBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 1024,
		"system":     systemPrompt,
		"messages":   messages,
		"stream":     true,
	}

	return c.sendStreamRequest(ctx, reqBody, callback)
}

// sendStreamRequest sends a streaming request to the LLM via the proxy.
func (c *LLMClient) sendStreamRequest(ctx context.Context, reqBody map[string]interface{}, callback StreamCallback) error {
	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	sessionID := fmt.Sprintf("bot-%d", time.Now().UnixNano())
	url := fmt.Sprintf("http://127.0.0.1:%d/%s/%s/v1/messages", c.proxyPort, c.profile, sessionID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	// Use a client without timeout for streaming
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			callback(event.Delta.Text)
		}
	}

	return scanner.Err()
}

// BuildSystemPrompt builds the system prompt for chat mode.
// If memory is non-empty, it is appended as persona/behavioral instructions.
func BuildSystemPrompt(processes []*ProcessInfo, profile string, memory string) string {
	// Build process list
	processSection := formatProcessList(processes)

	// Build persona section
	personaSection := ""
	if memory != "" {
		personaSection = fmt.Sprintf("\n## Persona Instructions\n%s\n", memory)
	}

	return fmt.Sprintf(`You are Zen, the GoZen assistant bot. Your name is Zen. Never identify yourself by any other name.

GoZen is a tool that manages AI API providers and routes requests to coding AI sessions (like Claude Code). You are the control interface for GoZen.

## Your Capabilities
- Report on connected coding sessions managed by GoZen
- Relay the status of these sessions (idle, busy, waiting)
- Help control sessions (pause/resume/cancel)
- Answer questions about GoZen configuration

## IMPORTANT: What "tasks", "processes", "sessions" mean
When users ask about "tasks", "processes", "sessions", "进程", "任务", "状态", "タスク", "작업", or any similar term in ANY language, they are asking about **GoZen-managed coding sessions** — NOT operating system processes. You must ALWAYS respond with information from the "Connected Sessions" section below. If there are no connected sessions, say so clearly.

## Connected Sessions
%s
Active profile: %s
%s
## Response Guidelines
- Keep responses concise and friendly
- Use markdown formatting
- Respond in the same language the user writes in
- When listing sessions, format them clearly with status indicators
- If asked about something outside your capabilities, briefly explain what you can help with`, processSection, profile, personaSection)
}

// formatProcessList formats the process list for the system prompt.
func formatProcessList(processes []*ProcessInfo) string {
	if len(processes) == 0 {
		return "No connected sessions."
	}

	result := fmt.Sprintf("%d connected session(s):\n", len(processes))
	for _, p := range processes {
		name := p.Name
		if p.Alias != "" {
			name = p.Alias
		}
		status := p.Status
		if status == "" {
			status = "idle"
		}
		uptime := time.Since(p.StartTime).Round(time.Second)

		result += fmt.Sprintf("- %s | status: %s | path: %s | uptime: %s", name, status, p.Path, uptime)

		if p.CurrentTask != "" {
			result += fmt.Sprintf(" | task: %s", p.CurrentTask)
		}
		if p.WaitingFor != "" {
			result += fmt.Sprintf(" | waiting for: %s", p.WaitingFor)
		}
		if p.PendingAction != "" {
			result += fmt.Sprintf(" | pending action: %s", p.PendingAction)
		}
		if p.LastMessage != "" {
			msg := p.LastMessage
			if len(msg) > 100 {
				msg = msg[:100] + "..."
			}
			role := p.MessageRole
			if role == "" {
				role = "unknown"
			}
			result += fmt.Sprintf(" | last %s message: \"%s\"", role, msg)
		}
		if p.TurnCount > 0 {
			result += fmt.Sprintf(" | turns: %d", p.TurnCount)
		}
		result += "\n"
	}
	return result
}

package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMClient handles LLM calls through the proxy.
type LLMClient struct {
	proxyPort int
	profile   string
	client    *http.Client
}

// NewLLMClient creates a new LLM client.
func NewLLMClient(proxyPort int, profile string) *LLMClient {
	return &LLMClient{
		proxyPort: proxyPort,
		profile:   profile,
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
		"model":      "claude-3-haiku-20240307",
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

// BuildSystemPrompt builds the system prompt for chat mode.
// If memory is non-empty, it is used as persona/behavioral instructions.
func BuildSystemPrompt(processes []*ProcessInfo, profile string, memory string) string {
	processesStr := "No connected processes."
	if len(processes) > 0 {
		processesStr = "Connected processes:\n"
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
			processesStr += fmt.Sprintf("- %s (%s) — path: %s, uptime: %s", name, status, p.Path, uptime)
			if p.CurrentTask != "" {
				processesStr += fmt.Sprintf(", task: %s", p.CurrentTask)
			}
			processesStr += "\n"
		}
	}

	var base string
	if memory != "" {
		base = memory
	} else {
		base = "You are Zen, the GoZen assistant bot. GoZen manages AI API providers and routes requests."
	}

	return fmt.Sprintf(`%s

You can help with:
- Checking status of connected coding processes
- Listing connected processes
- Controlling tasks (pause/resume/cancel)
- Answering questions about GoZen configuration

Current state:
%s
Profile: %s

Keep responses concise and friendly. Use the chat platform's formatting (markdown).
If the user asks something you can't help with, suggest what you can do.`, base, processesStr, profile)
}

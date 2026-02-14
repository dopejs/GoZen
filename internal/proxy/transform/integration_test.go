package transform

import (
	"encoding/json"
	"testing"
)

// TestIntegration_ClaudeCodeWithOpenAIProvider tests the scenario where
// Claude Code (Anthropic format) uses an OpenAI-format provider.
func TestIntegration_ClaudeCodeWithOpenAIProvider(t *testing.T) {
	clientFormat := "anthropic"  // Claude Code uses Anthropic format
	providerFormat := "openai"   // Provider uses OpenAI format

	// Verify transform is needed
	if !NeedsTransform(clientFormat, providerFormat) {
		t.Fatal("NeedsTransform should return true for anthropic -> openai")
	}

	// Get the transformer for the provider format
	transformer := GetTransformer(providerFormat)
	if transformer.Name() != "openai" {
		t.Fatalf("GetTransformer(%q) should return OpenAI transformer", providerFormat)
	}

	// === Request Transform: Anthropic → OpenAI ===
	// Claude Code sends Anthropic format request
	anthropicRequest := map[string]interface{}{
		"model":      "claude-sonnet-4-5",
		"max_tokens": float64(1024),
		"system":     "You are a helpful assistant.",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "Hello"},
		},
		"stop_sequences": []interface{}{"END"},
	}
	requestBytes, _ := json.Marshal(anthropicRequest)

	// Transform to OpenAI format for the provider
	transformedRequest, err := transformer.TransformRequest(requestBytes, clientFormat)
	if err != nil {
		t.Fatalf("TransformRequest error: %v", err)
	}

	var openAIRequest map[string]interface{}
	if err := json.Unmarshal(transformedRequest, &openAIRequest); err != nil {
		t.Fatalf("Failed to parse transformed request: %v", err)
	}

	// Verify OpenAI format
	if openAIRequest["max_completion_tokens"] != float64(1024) {
		t.Errorf("max_completion_tokens = %v, want 1024", openAIRequest["max_completion_tokens"])
	}
	if _, exists := openAIRequest["max_tokens"]; exists {
		t.Error("max_tokens should be removed in OpenAI format")
	}
	if _, exists := openAIRequest["system"]; exists {
		t.Error("system field should be converted to message")
	}
	if openAIRequest["stop"] == nil {
		t.Error("stop should be set (converted from stop_sequences)")
	}

	// System message should be prepended to messages
	messages := openAIRequest["messages"].([]interface{})
	if len(messages) != 2 {
		t.Fatalf("messages should have 2 items (system + user), got %d", len(messages))
	}
	systemMsg := messages[0].(map[string]interface{})
	if systemMsg["role"] != "system" {
		t.Errorf("first message role = %v, want system", systemMsg["role"])
	}

	// === Response Transform: OpenAI → Anthropic ===
	// Provider returns OpenAI format response
	openAIResponse := map[string]interface{}{
		"id":      "chatcmpl-123",
		"object":  "chat.completion",
		"created": float64(1234567890),
		"model":   "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"index": float64(0),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello! How can I help you?",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     float64(20),
			"completion_tokens": float64(10),
			"total_tokens":      float64(30),
		},
	}
	responseBytes, _ := json.Marshal(openAIResponse)

	// Transform back to Anthropic format for Claude Code
	transformedResponse, err := transformer.TransformResponse(responseBytes, clientFormat)
	if err != nil {
		t.Fatalf("TransformResponse error: %v", err)
	}

	var anthropicResponse map[string]interface{}
	if err := json.Unmarshal(transformedResponse, &anthropicResponse); err != nil {
		t.Fatalf("Failed to parse transformed response: %v", err)
	}

	// Verify Anthropic format
	if anthropicResponse["type"] != "message" {
		t.Errorf("type = %v, want message", anthropicResponse["type"])
	}
	if anthropicResponse["role"] != "assistant" {
		t.Errorf("role = %v, want assistant", anthropicResponse["role"])
	}
	if anthropicResponse["stop_reason"] != "end_turn" {
		t.Errorf("stop_reason = %v, want end_turn", anthropicResponse["stop_reason"])
	}

	// Content should be in Anthropic format
	content := anthropicResponse["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content should have 1 item, got %d", len(content))
	}
	textBlock := content[0].(map[string]interface{})
	if textBlock["type"] != "text" {
		t.Errorf("content type = %v, want text", textBlock["type"])
	}
	if textBlock["text"] != "Hello! How can I help you?" {
		t.Errorf("content text = %v, want 'Hello! How can I help you?'", textBlock["text"])
	}

	// Usage should be in Anthropic format
	usage := anthropicResponse["usage"].(map[string]interface{})
	if usage["input_tokens"] != float64(20) {
		t.Errorf("input_tokens = %v, want 20", usage["input_tokens"])
	}
	if usage["output_tokens"] != float64(10) {
		t.Errorf("output_tokens = %v, want 10", usage["output_tokens"])
	}
}

// TestIntegration_CodexWithAnthropicProvider tests the scenario where
// Codex (OpenAI format) uses an Anthropic-format provider.
func TestIntegration_CodexWithAnthropicProvider(t *testing.T) {
	clientFormat := "openai"      // Codex uses OpenAI format
	providerFormat := "anthropic" // Provider uses Anthropic format

	// Verify transform is needed
	if !NeedsTransform(clientFormat, providerFormat) {
		t.Fatal("NeedsTransform should return true for openai -> anthropic")
	}

	// Get the transformer for the provider format
	transformer := GetTransformer(providerFormat)
	if transformer.Name() != "anthropic" {
		t.Fatalf("GetTransformer(%q) should return Anthropic transformer", providerFormat)
	}

	// === Request Transform: OpenAI → Anthropic ===
	// Codex sends OpenAI format request
	openAIRequest := map[string]interface{}{
		"model":                 "gpt-4",
		"max_completion_tokens": float64(2048),
		"messages": []interface{}{
			map[string]interface{}{"role": "system", "content": "You are a coding assistant."},
			map[string]interface{}{"role": "user", "content": "Write a hello world in Go"},
		},
		"stop":              []interface{}{"```"},
		"temperature":       float64(0.7),
		"presence_penalty":  float64(0.5),
		"frequency_penalty": float64(0.5),
	}
	requestBytes, _ := json.Marshal(openAIRequest)

	// Transform to Anthropic format for the provider
	transformedRequest, err := transformer.TransformRequest(requestBytes, clientFormat)
	if err != nil {
		t.Fatalf("TransformRequest error: %v", err)
	}

	var anthropicRequest map[string]interface{}
	if err := json.Unmarshal(transformedRequest, &anthropicRequest); err != nil {
		t.Fatalf("Failed to parse transformed request: %v", err)
	}

	// Verify Anthropic format
	if anthropicRequest["max_tokens"] != float64(2048) {
		t.Errorf("max_tokens = %v, want 2048", anthropicRequest["max_tokens"])
	}
	if _, exists := anthropicRequest["max_completion_tokens"]; exists {
		t.Error("max_completion_tokens should be removed in Anthropic format")
	}
	if anthropicRequest["stop_sequences"] == nil {
		t.Error("stop_sequences should be set (converted from stop)")
	}
	// OpenAI-specific fields should be removed
	if _, exists := anthropicRequest["presence_penalty"]; exists {
		t.Error("presence_penalty should be removed")
	}
	if _, exists := anthropicRequest["frequency_penalty"]; exists {
		t.Error("frequency_penalty should be removed")
	}

	// === Response Transform: Anthropic → OpenAI ===
	// Provider returns Anthropic format response
	anthropicResponse := map[string]interface{}{
		"id":    "msg_456",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-sonnet-4-5",
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": "```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```",
			},
		},
		"stop_reason": "end_turn",
		"usage": map[string]interface{}{
			"input_tokens":  float64(50),
			"output_tokens": float64(30),
		},
	}
	responseBytes, _ := json.Marshal(anthropicResponse)

	// Transform back to OpenAI format for Codex
	transformedResponse, err := transformer.TransformResponse(responseBytes, clientFormat)
	if err != nil {
		t.Fatalf("TransformResponse error: %v", err)
	}

	var openAIResponse map[string]interface{}
	if err := json.Unmarshal(transformedResponse, &openAIResponse); err != nil {
		t.Fatalf("Failed to parse transformed response: %v", err)
	}

	// Verify OpenAI format
	if openAIResponse["object"] != "chat.completion" {
		t.Errorf("object = %v, want chat.completion", openAIResponse["object"])
	}
	if openAIResponse["id"] != "msg_456" {
		t.Errorf("id = %v, want msg_456", openAIResponse["id"])
	}

	// Choices should be in OpenAI format
	choices := openAIResponse["choices"].([]interface{})
	if len(choices) != 1 {
		t.Fatalf("choices should have 1 item, got %d", len(choices))
	}
	choice := choices[0].(map[string]interface{})
	if choice["finish_reason"] != "stop" {
		t.Errorf("finish_reason = %v, want stop", choice["finish_reason"])
	}

	message := choice["message"].(map[string]interface{})
	if message["role"] != "assistant" {
		t.Errorf("message role = %v, want assistant", message["role"])
	}

	// Usage should be in OpenAI format
	usage := openAIResponse["usage"].(map[string]interface{})
	if usage["prompt_tokens"] != float64(50) {
		t.Errorf("prompt_tokens = %v, want 50", usage["prompt_tokens"])
	}
	if usage["completion_tokens"] != float64(30) {
		t.Errorf("completion_tokens = %v, want 30", usage["completion_tokens"])
	}
	if usage["total_tokens"] != float64(80) {
		t.Errorf("total_tokens = %v, want 80", usage["total_tokens"])
	}
}

// TestIntegration_NoTransformNeeded tests scenarios where no transform is needed.
func TestIntegration_NoTransformNeeded(t *testing.T) {
	tests := []struct {
		name           string
		clientFormat   string
		providerFormat string
	}{
		{"Claude Code with Anthropic provider", "anthropic", "anthropic"},
		{"Codex with OpenAI provider", "openai", "openai"},
		{"Default client with default provider", "", ""},
		{"Default client with Anthropic provider", "", "anthropic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if NeedsTransform(tt.clientFormat, tt.providerFormat) {
				t.Errorf("NeedsTransform(%q, %q) should return false", tt.clientFormat, tt.providerFormat)
			}
		})
	}
}

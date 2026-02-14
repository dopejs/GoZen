package transform

import (
	"encoding/json"
	"testing"
)

func TestOpenAITransformer_Name(t *testing.T) {
	tr := &OpenAITransformer{}
	if tr.Name() != "openai" {
		t.Errorf("Name() = %q, want %q", tr.Name(), "openai")
	}
}

func TestOpenAITransformer_TransformRequest_NoTransform(t *testing.T) {
	tr := &OpenAITransformer{}

	// OpenAI client → OpenAI provider: no transform needed
	input := `{"model": "gpt-4", "max_completion_tokens": 1024, "messages": [{"role": "user", "content": "Hello"}]}`

	result, err := tr.TransformRequest([]byte(input), "openai")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	if string(result) != input {
		t.Errorf("TransformRequest() should not modify when client is openai")
	}
}

func TestOpenAITransformer_TransformRequest_AnthropicToOpenAI(t *testing.T) {
	tr := &OpenAITransformer{}

	// Anthropic format request
	input := map[string]interface{}{
		"model":          "claude-sonnet-4-5",
		"max_tokens":     float64(1024),
		"messages":       []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
		"stop_sequences": []interface{}{"END"},
		"metadata":       map[string]interface{}{"user_id": "123"},
		"thinking":       map[string]interface{}{"type": "enabled"},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformRequest(inputBytes, "anthropic")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// max_tokens → max_completion_tokens
	if output["max_completion_tokens"] != float64(1024) {
		t.Errorf("max_completion_tokens = %v, want %v", output["max_completion_tokens"], float64(1024))
	}
	if _, exists := output["max_tokens"]; exists {
		t.Error("max_tokens should be removed")
	}

	// stop_sequences → stop
	if output["stop"] == nil {
		t.Error("stop should be set")
	}
	if _, exists := output["stop_sequences"]; exists {
		t.Error("stop_sequences should be removed")
	}

	// Anthropic-specific fields should be removed
	if _, exists := output["metadata"]; exists {
		t.Error("metadata should be removed")
	}
	if _, exists := output["thinking"]; exists {
		t.Error("thinking should be removed")
	}
}

func TestOpenAITransformer_TransformRequest_SystemMessage(t *testing.T) {
	tr := &OpenAITransformer{}

	// Anthropic format with system field
	input := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"system":   "You are a helpful assistant.",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformRequest(inputBytes, "anthropic")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// system field should be removed
	if _, exists := output["system"]; exists {
		t.Error("system field should be removed")
	}

	// messages should have system message prepended
	messages, ok := output["messages"].([]interface{})
	if !ok || len(messages) != 2 {
		t.Fatalf("messages should have 2 items, got %d", len(messages))
	}

	systemMsg := messages[0].(map[string]interface{})
	if systemMsg["role"] != "system" {
		t.Errorf("first message role = %v, want %v", systemMsg["role"], "system")
	}
	if systemMsg["content"] != "You are a helpful assistant." {
		t.Errorf("first message content = %v, want %v", systemMsg["content"], "You are a helpful assistant.")
	}

	userMsg := messages[1].(map[string]interface{})
	if userMsg["role"] != "user" {
		t.Errorf("second message role = %v, want %v", userMsg["role"], "user")
	}
}

func TestOpenAITransformer_TransformRequest_ToolsConversion(t *testing.T) {
	tr := &OpenAITransformer{}

	// Anthropic tools format
	input := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"messages": []interface{}{},
		"tools": []interface{}{
			map[string]interface{}{
				"name":        "get_weather",
				"description": "Get weather for a location",
				"input_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformRequest(inputBytes, "anthropic")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	tools, ok := output["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		t.Fatal("tools should be converted")
	}

	tool := tools[0].(map[string]interface{})
	if tool["type"] != "function" {
		t.Errorf("tool type = %v, want %v", tool["type"], "function")
	}

	fn := tool["function"].(map[string]interface{})
	if fn["name"] != "get_weather" {
		t.Errorf("function name = %v, want %v", fn["name"], "get_weather")
	}
	if fn["description"] != "Get weather for a location" {
		t.Errorf("function description = %v, want %v", fn["description"], "Get weather for a location")
	}
	if fn["parameters"] == nil {
		t.Error("function parameters should be set")
	}
}

func TestOpenAITransformer_TransformResponse_NoTransform(t *testing.T) {
	tr := &OpenAITransformer{}

	input := `{"id": "chatcmpl-123", "object": "chat.completion", "choices": [{"message": {"content": "Hello"}}]}`

	result, err := tr.TransformResponse([]byte(input), "openai")
	if err != nil {
		t.Fatalf("TransformResponse() error = %v", err)
	}

	if string(result) != input {
		t.Errorf("TransformResponse() should not modify when client is openai")
	}
}

func TestOpenAITransformer_TransformResponse_OpenAIToAnthropic(t *testing.T) {
	tr := &OpenAITransformer{}

	// OpenAI response format
	input := map[string]interface{}{
		"id":      "chatcmpl-123",
		"object":  "chat.completion",
		"created": float64(1234567890),
		"model":   "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"index": float64(0),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello, world!",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     float64(10),
			"completion_tokens": float64(5),
			"total_tokens":      float64(15),
		},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformResponse(inputBytes, "anthropic")
	if err != nil {
		t.Fatalf("TransformResponse() error = %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// Check Anthropic format
	if output["type"] != "message" {
		t.Errorf("type = %v, want %v", output["type"], "message")
	}
	if output["role"] != "assistant" {
		t.Errorf("role = %v, want %v", output["role"], "assistant")
	}
	if output["id"] != "chatcmpl-123" {
		t.Errorf("id = %v, want %v", output["id"], "chatcmpl-123")
	}
	if output["model"] != "gpt-4" {
		t.Errorf("model = %v, want %v", output["model"], "gpt-4")
	}

	// Check content
	content, ok := output["content"].([]interface{})
	if !ok || len(content) == 0 {
		t.Fatal("content should be set")
	}
	contentBlock := content[0].(map[string]interface{})
	if contentBlock["type"] != "text" {
		t.Errorf("content type = %v, want %v", contentBlock["type"], "text")
	}
	if contentBlock["text"] != "Hello, world!" {
		t.Errorf("content text = %v, want %v", contentBlock["text"], "Hello, world!")
	}

	// Check stop_reason
	if output["stop_reason"] != "end_turn" {
		t.Errorf("stop_reason = %v, want %v", output["stop_reason"], "end_turn")
	}

	// Check usage
	usage := output["usage"].(map[string]interface{})
	if usage["input_tokens"] != float64(10) {
		t.Errorf("input_tokens = %v, want %v", usage["input_tokens"], float64(10))
	}
	if usage["output_tokens"] != float64(5) {
		t.Errorf("output_tokens = %v, want %v", usage["output_tokens"], float64(5))
	}
}

func TestOpenAITransformer_FinishReasonMapping(t *testing.T) {
	tr := &OpenAITransformer{}

	tests := []struct {
		finishReason string
		stopReason   string
	}{
		{"stop", "end_turn"},
		{"length", "max_tokens"},
		{"tool_calls", "tool_use"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.finishReason, func(t *testing.T) {
			input := map[string]interface{}{
				"id":     "chatcmpl-123",
				"object": "chat.completion",
				"model":  "gpt-4",
				"choices": []interface{}{
					map[string]interface{}{
						"index":         float64(0),
						"message":       map[string]interface{}{"role": "assistant", "content": "Hi"},
						"finish_reason": tt.finishReason,
					},
				},
			}
			inputBytes, _ := json.Marshal(input)

			result, err := tr.TransformResponse(inputBytes, "anthropic")
			if err != nil {
				t.Fatalf("TransformResponse() error = %v", err)
			}

			var output map[string]interface{}
			json.Unmarshal(result, &output)

			if output["stop_reason"] != tt.stopReason {
				t.Errorf("stop_reason = %v, want %v", output["stop_reason"], tt.stopReason)
			}
		})
	}
}

func TestOpenAITransformer_InvalidJSON(t *testing.T) {
	tr := &OpenAITransformer{}

	// Invalid JSON should return original
	invalid := []byte(`{invalid json}`)

	result, err := tr.TransformRequest(invalid, "anthropic")
	if err != nil {
		t.Fatalf("TransformRequest() should not error on invalid JSON")
	}
	if string(result) != string(invalid) {
		t.Error("TransformRequest() should return original on invalid JSON")
	}

	result, err = tr.TransformResponse(invalid, "anthropic")
	if err != nil {
		t.Fatalf("TransformResponse() should not error on invalid JSON")
	}
	if string(result) != string(invalid) {
		t.Error("TransformResponse() should return original on invalid JSON")
	}
}

func TestOpenAITransformer_EmptyMessages(t *testing.T) {
	tr := &OpenAITransformer{}

	// Request with no messages
	input := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"system":   "You are helpful.",
		"messages": []interface{}{},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformRequest(inputBytes, "anthropic")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// Should have system message prepended
	messages := output["messages"].([]interface{})
	if len(messages) != 1 {
		t.Errorf("messages length = %d, want 1", len(messages))
	}
}

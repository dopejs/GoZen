package transform

import (
	"encoding/json"
	"testing"
)

func TestAnthropicTransformer_Name(t *testing.T) {
	tr := &AnthropicTransformer{}
	if tr.Name() != "anthropic" {
		t.Errorf("Name() = %q, want %q", tr.Name(), "anthropic")
	}
}

func TestAnthropicTransformer_TransformRequest_NoTransform(t *testing.T) {
	tr := &AnthropicTransformer{}

	// Anthropic client → Anthropic provider: no transform needed
	input := `{"model": "claude-sonnet-4-5", "max_tokens": 1024, "messages": [{"role": "user", "content": "Hello"}]}`

	result, err := tr.TransformRequest([]byte(input), "anthropic")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	// Should return unchanged
	if string(result) != input {
		t.Errorf("TransformRequest() should not modify when client is anthropic")
	}

	// Empty client format should also not transform
	result, err = tr.TransformRequest([]byte(input), "")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}
	if string(result) != input {
		t.Errorf("TransformRequest() should not modify when client is empty")
	}
}

func TestAnthropicTransformer_TransformRequest_OpenAIToAnthropic(t *testing.T) {
	tr := &AnthropicTransformer{}

	// OpenAI format request
	input := map[string]interface{}{
		"model":                 "claude-sonnet-4-5",
		"max_completion_tokens": float64(1024),
		"messages":              []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
		"stop":                  []interface{}{"END"},
		"n":                     float64(1),
		"presence_penalty":      float64(0.5),
		"frequency_penalty":     float64(0.5),
		"seed":                  float64(42),
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformRequest(inputBytes, "openai")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// max_completion_tokens → max_tokens
	if output["max_tokens"] != float64(1024) {
		t.Errorf("max_tokens = %v, want %v", output["max_tokens"], float64(1024))
	}
	if _, exists := output["max_completion_tokens"]; exists {
		t.Error("max_completion_tokens should be removed")
	}

	// stop → stop_sequences
	if output["stop_sequences"] == nil {
		t.Error("stop_sequences should be set")
	}
	if _, exists := output["stop"]; exists {
		t.Error("stop should be removed")
	}

	// OpenAI-specific fields should be removed
	if _, exists := output["n"]; exists {
		t.Error("n should be removed")
	}
	if _, exists := output["presence_penalty"]; exists {
		t.Error("presence_penalty should be removed")
	}
	if _, exists := output["frequency_penalty"]; exists {
		t.Error("frequency_penalty should be removed")
	}
	if _, exists := output["seed"]; exists {
		t.Error("seed should be removed")
	}
}

func TestAnthropicTransformer_TransformRequest_ToolsConversion(t *testing.T) {
	tr := &AnthropicTransformer{}

	// OpenAI tools format
	input := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"messages": []interface{}{},
		"tools": []interface{}{
			map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        "get_weather",
					"description": "Get weather for a location",
					"parameters": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformRequest(inputBytes, "openai")
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
	if tool["name"] != "get_weather" {
		t.Errorf("tool name = %v, want %v", tool["name"], "get_weather")
	}
	if tool["description"] != "Get weather for a location" {
		t.Errorf("tool description = %v, want %v", tool["description"], "Get weather for a location")
	}
	if tool["input_schema"] == nil {
		t.Error("tool input_schema should be set")
	}
}

func TestAnthropicTransformer_TransformResponse_NoTransform(t *testing.T) {
	tr := &AnthropicTransformer{}

	input := `{"id": "msg_123", "type": "message", "content": [{"type": "text", "text": "Hello"}]}`

	result, err := tr.TransformResponse([]byte(input), "anthropic")
	if err != nil {
		t.Fatalf("TransformResponse() error = %v", err)
	}

	if string(result) != input {
		t.Errorf("TransformResponse() should not modify when client is anthropic")
	}
}

func TestAnthropicTransformer_TransformResponse_AnthropicToOpenAI(t *testing.T) {
	tr := &AnthropicTransformer{}

	// Anthropic response format
	input := map[string]interface{}{
		"id":    "msg_123",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-sonnet-4-5",
		"content": []interface{}{
			map[string]interface{}{"type": "text", "text": "Hello, world!"},
		},
		"stop_reason": "end_turn",
		"usage": map[string]interface{}{
			"input_tokens":  float64(10),
			"output_tokens": float64(5),
		},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformResponse(inputBytes, "openai")
	if err != nil {
		t.Fatalf("TransformResponse() error = %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// Check OpenAI format
	if output["object"] != "chat.completion" {
		t.Errorf("object = %v, want %v", output["object"], "chat.completion")
	}
	if output["id"] != "msg_123" {
		t.Errorf("id = %v, want %v", output["id"], "msg_123")
	}
	if output["model"] != "claude-sonnet-4-5" {
		t.Errorf("model = %v, want %v", output["model"], "claude-sonnet-4-5")
	}

	// Check choices
	choices, ok := output["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Fatal("choices should be set")
	}
	choice := choices[0].(map[string]interface{})
	if choice["index"] != float64(0) {
		t.Errorf("choice index = %v, want %v", choice["index"], float64(0))
	}
	if choice["finish_reason"] != "stop" {
		t.Errorf("finish_reason = %v, want %v", choice["finish_reason"], "stop")
	}

	message := choice["message"].(map[string]interface{})
	if message["role"] != "assistant" {
		t.Errorf("message role = %v, want %v", message["role"], "assistant")
	}
	if message["content"] != "Hello, world!" {
		t.Errorf("message content = %v, want %v", message["content"], "Hello, world!")
	}

	// Check usage
	usage := output["usage"].(map[string]interface{})
	if usage["prompt_tokens"] != float64(10) {
		t.Errorf("prompt_tokens = %v, want %v", usage["prompt_tokens"], float64(10))
	}
	if usage["completion_tokens"] != float64(5) {
		t.Errorf("completion_tokens = %v, want %v", usage["completion_tokens"], float64(5))
	}
	if usage["total_tokens"] != float64(15) {
		t.Errorf("total_tokens = %v, want %v", usage["total_tokens"], float64(15))
	}
}

func TestAnthropicTransformer_StopReasonMapping(t *testing.T) {
	tr := &AnthropicTransformer{}

	tests := []struct {
		stopReason   string
		finishReason string
	}{
		{"end_turn", "stop"},
		{"max_tokens", "length"},
		{"tool_use", "tool_calls"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.stopReason, func(t *testing.T) {
			input := map[string]interface{}{
				"id":          "msg_123",
				"type":        "message",
				"model":       "claude-sonnet-4-5",
				"content":     []interface{}{map[string]interface{}{"type": "text", "text": "Hi"}},
				"stop_reason": tt.stopReason,
			}
			inputBytes, _ := json.Marshal(input)

			result, err := tr.TransformResponse(inputBytes, "openai")
			if err != nil {
				t.Fatalf("TransformResponse() error = %v", err)
			}

			var output map[string]interface{}
			json.Unmarshal(result, &output)

			choices := output["choices"].([]interface{})
			choice := choices[0].(map[string]interface{})
			if choice["finish_reason"] != tt.finishReason {
				t.Errorf("finish_reason = %v, want %v", choice["finish_reason"], tt.finishReason)
			}
		})
	}
}

func TestAnthropicTransformer_InvalidJSON(t *testing.T) {
	tr := &AnthropicTransformer{}

	// Invalid JSON should return original
	invalid := []byte(`{invalid json}`)

	result, err := tr.TransformRequest(invalid, "openai")
	if err != nil {
		t.Fatalf("TransformRequest() should not error on invalid JSON")
	}
	if string(result) != string(invalid) {
		t.Error("TransformRequest() should return original on invalid JSON")
	}

	result, err = tr.TransformResponse(invalid, "openai")
	if err != nil {
		t.Fatalf("TransformResponse() should not error on invalid JSON")
	}
	if string(result) != string(invalid) {
		t.Error("TransformResponse() should return original on invalid JSON")
	}
}

func TestConvertInputToMessages_StringInput(t *testing.T) {
	messages := convertInputToMessages("Hello, world!")
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	msg := messages[0].(map[string]interface{})
	if msg["role"] != "user" {
		t.Errorf("role = %v, want user", msg["role"])
	}
	if msg["content"] != "Hello, world!" {
		t.Errorf("content = %v, want Hello, world!", msg["content"])
	}
}

func TestConvertInputToMessages_ArrayOfStrings(t *testing.T) {
	input := []interface{}{"Hello", "World"}
	messages := convertInputToMessages(input)
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	msg0 := messages[0].(map[string]interface{})
	if msg0["content"] != "Hello" {
		t.Errorf("first content = %v, want Hello", msg0["content"])
	}
}

func TestConvertInputToMessages_TypedMessages(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"type":    "message",
			"role":    "user",
			"content": "What is 2+2?",
		},
		map[string]interface{}{
			"type":    "message",
			"role":    "assistant",
			"content": "4",
		},
	}
	messages := convertInputToMessages(input)
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	msg1 := messages[1].(map[string]interface{})
	if msg1["role"] != "assistant" {
		t.Errorf("role = %v, want assistant", msg1["role"])
	}
}

func TestConvertInputToMessages_DeveloperRole(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"type":    "message",
			"role":    "developer",
			"content": "You are a helpful assistant.",
		},
		map[string]interface{}{
			"type":    "message",
			"role":    "user",
			"content": "Hello",
		},
	}
	messages := convertInputToMessages(input)
	// Developer message should be extracted as _system marker, not included as regular message
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages (system marker + user), got %d", len(messages))
	}
	// First message should be the system marker
	first := messages[0].(map[string]interface{})
	if _, ok := first["_system"]; !ok {
		t.Error("expected _system marker in first message")
	}
}

func TestConvertInputToMessages_SystemRole(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"role":    "system",
			"content": "System instructions",
		},
		map[string]interface{}{
			"role":    "user",
			"content": "Hello",
		},
	}
	messages := convertInputToMessages(input)
	// System message should be extracted as _system marker
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	first := messages[0].(map[string]interface{})
	if sysContent, ok := first["_system"].(string); !ok || sysContent != "System instructions" {
		t.Errorf("expected _system='System instructions', got %v", first["_system"])
	}
}

func TestConvertInputToMessages_ContentItems(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"type": "input_text",
			"text": "Hello from input_text",
		},
		map[string]interface{}{
			"type": "text",
			"text": "Hello from text",
		},
	}
	messages := convertInputToMessages(input)
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
}

func TestConvertContent_String(t *testing.T) {
	result := convertContent("hello")
	if result != "hello" {
		t.Errorf("expected hello, got %v", result)
	}
}

func TestConvertContent_Array(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "input_text", "text": "hello"},
		map[string]interface{}{"type": "text", "text": "world"},
		map[string]interface{}{"type": "image", "url": "http://example.com"},
	}
	result := convertContent(input)
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 items, got %d", len(arr))
	}
	// input_text should be converted to text type
	first := arr[0].(map[string]interface{})
	if first["type"] != "text" {
		t.Errorf("first type = %v, want text", first["type"])
	}
}

func TestConvertContent_EmptyArray(t *testing.T) {
	input := []interface{}{}
	result := convertContent(input)
	// Empty array returns original input
	if result != nil {
		arr, ok := result.([]interface{})
		if ok && len(arr) != 0 {
			t.Errorf("expected empty or nil, got %v", result)
		}
	}
}

func TestConvertContent_NonStringNonArray(t *testing.T) {
	result := convertContent(42)
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestExtractTextFromContent_String(t *testing.T) {
	result := extractTextFromContent("hello")
	if result != "hello" {
		t.Errorf("expected hello, got %v", result)
	}
}

func TestExtractTextFromContent_Array(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "text", "text": "hello"},
		map[string]interface{}{"type": "text", "text": "world"},
	}
	result := extractTextFromContent(input)
	if result != "hello\nworld" {
		t.Errorf("expected 'hello\\nworld', got %q", result)
	}
}

func TestExtractTextFromContent_EmptyArray(t *testing.T) {
	result := extractTextFromContent([]interface{}{})
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestExtractTextFromContent_OtherType(t *testing.T) {
	result := extractTextFromContent(42)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestMapRole(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"developer", "user"},
		{"system", "user"},
		{"assistant", "assistant"},
		{"user", "user"},
		{"unknown", "user"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapRole(tt.input)
			if got != tt.want {
				t.Errorf("mapRole(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAnthropicTransformer_TransformRequest_ResponsesAPI(t *testing.T) {
	tr := &AnthropicTransformer{}

	// OpenAI Responses API format with "input" and "instructions"
	input := map[string]interface{}{
		"model":             "gpt-4",
		"input":             "What is the capital of France?",
		"instructions":      "You are a geography expert.",
		"max_output_tokens": float64(500),
		"stream_options":    map[string]interface{}{"include_usage": true},
		"logprobs":          true,
		"top_logprobs":      float64(5),
		"response_format":   map[string]interface{}{"type": "json"},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformRequest(inputBytes, "openai")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// "input" should be converted to "messages"
	if _, exists := output["input"]; exists {
		t.Error("input should be removed")
	}
	if output["messages"] == nil {
		t.Error("messages should be set")
	}

	// "instructions" should be converted to "system"
	if _, exists := output["instructions"]; exists {
		t.Error("instructions should be removed")
	}
	if output["system"] != "You are a geography expert." {
		t.Errorf("system = %v, want 'You are a geography expert.'", output["system"])
	}

	// max_output_tokens → max_tokens
	if output["max_tokens"] != float64(500) {
		t.Errorf("max_tokens = %v, want 500", output["max_tokens"])
	}
	if _, exists := output["max_output_tokens"]; exists {
		t.Error("max_output_tokens should be removed")
	}

	// OpenAI-specific fields should be removed
	for _, field := range []string{"stream_options", "logprobs", "top_logprobs", "response_format"} {
		if _, exists := output[field]; exists {
			t.Errorf("%s should be removed", field)
		}
	}
}

func TestAnthropicTransformer_TransformRequest_InputWithSystemMessages(t *testing.T) {
	tr := &AnthropicTransformer{}

	// Input with developer role messages
	input := map[string]interface{}{
		"model": "gpt-4",
		"input": []interface{}{
			map[string]interface{}{
				"type":    "message",
				"role":    "developer",
				"content": "Be concise.",
			},
			map[string]interface{}{
				"type":    "message",
				"role":    "user",
				"content": "Hello",
			},
		},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := tr.TransformRequest(inputBytes, "openai")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	var output map[string]interface{}
	json.Unmarshal(result, &output)

	// System content should be extracted
	if output["system"] != "Be concise." {
		t.Errorf("system = %v, want 'Be concise.'", output["system"])
	}
}

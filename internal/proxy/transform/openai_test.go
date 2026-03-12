package transform

import (
	"encoding/json"
	"strings"
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

func TestOpenAITransformer_ToolsTransformation(t *testing.T) {
	tr := &AnthropicTransformer{}

	// OpenAI Chat tools format
	input := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "What's the weather?",
			},
		},
		"tools": []interface{}{
			map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        "get_weather",
					"description": "Get weather info",
					"parameters": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "City name",
							},
						},
						"required": []interface{}{"location"},
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

	// Verify tools transformed to Anthropic format
	tools, ok := output["tools"].([]interface{})
	if !ok || len(tools) != 1 {
		t.Fatalf("tools not transformed correctly")
	}

	tool := tools[0].(map[string]interface{})
	if tool["name"] != "get_weather" {
		t.Errorf("tool name = %v, want get_weather", tool["name"])
	}
	if tool["description"] != "Get weather info" {
		t.Errorf("tool description = %v, want Get weather info", tool["description"])
	}

	// Verify input_schema exists (Anthropic format)
	inputSchema, ok := tool["input_schema"].(map[string]interface{})
	if !ok {
		t.Fatal("input_schema not found in Anthropic tool format")
	}
	if inputSchema["type"] != "object" {
		t.Errorf("input_schema type = %v, want object", inputSchema["type"])
	}
}

// Test OpenAI tool_calls transformation to Anthropic tool_use
func TestOpenAITransformer_TransformResponse_ToolCalls(t *testing.T) {
	transformer := &OpenAITransformer{}

	// OpenAI response with tool_calls
	openaiResp := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": null,
				"tool_calls": [{
					"id": "call_abc123",
					"type": "function",
					"function": {
						"name": "get_weather",
						"arguments": "{\"location\":\"San Francisco\",\"unit\":\"celsius\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 20,
			"total_tokens": 30
		}
	}`

	result, err := transformer.TransformResponse([]byte(openaiResp), "anthropic-messages")
	if err != nil {
		t.Fatalf("TransformResponse failed: %v", err)
	}

	var anthropicResp map[string]interface{}
	if err := json.Unmarshal(result, &anthropicResp); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// Verify stop_reason
	if anthropicResp["stop_reason"] != "tool_use" {
		t.Errorf("expected stop_reason=tool_use, got %v", anthropicResp["stop_reason"])
	}

	// Verify content blocks
	content, ok := anthropicResp["content"].([]interface{})
	if !ok || len(content) == 0 {
		t.Fatalf("expected content blocks, got %v", anthropicResp["content"])
	}

	// Verify tool_use block
	toolUse, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tool_use block, got %v", content[0])
	}

	if toolUse["type"] != "tool_use" {
		t.Errorf("expected type=tool_use, got %v", toolUse["type"])
	}

	if toolUse["id"] != "call_abc123" {
		t.Errorf("expected id=call_abc123, got %v", toolUse["id"])
	}

	if toolUse["name"] != "get_weather" {
		t.Errorf("expected name=get_weather, got %v", toolUse["name"])
	}

	// Verify input arguments
	input, ok := toolUse["input"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected input object, got %v", toolUse["input"])
	}

	if input["location"] != "San Francisco" {
		t.Errorf("expected location=San Francisco, got %v", input["location"])
	}
}

// Test OpenAI response with both content and tool_calls
func TestOpenAITransformer_TransformResponse_ContentAndToolCalls(t *testing.T) {
	transformer := &OpenAITransformer{}

	openaiResp := `{
		"id": "chatcmpl-123",
		"choices": [{
			"message": {
				"role": "assistant",
				"content": "Let me check the weather for you.",
				"tool_calls": [{
					"id": "call_123",
					"type": "function",
					"function": {
						"name": "get_weather",
						"arguments": "{\"location\":\"NYC\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`

	result, err := transformer.TransformResponse([]byte(openaiResp), "anthropic")
	if err != nil {
		t.Fatalf("TransformResponse failed: %v", err)
	}

	var anthropicResp map[string]interface{}
	json.Unmarshal(result, &anthropicResp)

	content := anthropicResp["content"].([]interface{})
	if len(content) != 2 {
		t.Errorf("expected 2 content blocks (text + tool_use), got %d", len(content))
	}

	// First block should be text
	textBlock := content[0].(map[string]interface{})
	if textBlock["type"] != "text" {
		t.Errorf("expected first block type=text, got %v", textBlock["type"])
	}

	// Second block should be tool_use
	toolBlock := content[1].(map[string]interface{})
	if toolBlock["type"] != "tool_use" {
		t.Errorf("expected second block type=tool_use, got %v", toolBlock["type"])
	}
}

// Test Anthropic -> OpenAI request transformation with tool_use
func TestOpenAITransformer_TransformRequest_ToolUse(t *testing.T) {
	transformer := &OpenAITransformer{}

	// Anthropic request with assistant message containing tool_use
	anthropicReq := `{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": [
			{
				"role": "user",
				"content": "What's the weather in SF?"
			},
			{
				"role": "assistant",
				"content": [
					{
						"type": "text",
						"text": "Let me check the weather for you."
					},
					{
						"type": "tool_use",
						"id": "toolu_123",
						"name": "get_weather",
						"input": {"location": "San Francisco", "unit": "celsius"}
					}
				]
			}
		]
	}`

	result, err := transformer.TransformRequest([]byte(anthropicReq), "anthropic-messages")
	if err != nil {
		t.Fatalf("TransformRequest failed: %v", err)
	}

	var openaiReq map[string]interface{}
	if err := json.Unmarshal(result, &openaiReq); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// Verify max_tokens -> max_completion_tokens
	if openaiReq["max_completion_tokens"] != float64(1024) {
		t.Errorf("expected max_completion_tokens=1024, got %v", openaiReq["max_completion_tokens"])
	}

	// Verify messages
	messages := openaiReq["messages"].([]interface{})
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	// Check assistant message with tool_calls
	assistantMsg := messages[1].(map[string]interface{})
	if assistantMsg["role"] != "assistant" {
		t.Errorf("expected role=assistant, got %v", assistantMsg["role"])
	}

	if assistantMsg["content"] != "Let me check the weather for you." {
		t.Errorf("expected text content, got %v", assistantMsg["content"])
	}

	toolCalls := assistantMsg["tool_calls"].([]interface{})
	if len(toolCalls) != 1 {
		t.Fatalf("expected 1 tool_call, got %d", len(toolCalls))
	}

	toolCall := toolCalls[0].(map[string]interface{})
	if toolCall["id"] != "toolu_123" {
		t.Errorf("expected id=toolu_123, got %v", toolCall["id"])
	}

	function := toolCall["function"].(map[string]interface{})
	if function["name"] != "get_weather" {
		t.Errorf("expected name=get_weather, got %v", function["name"])
	}

	// Verify arguments is JSON string
	argsStr := function["arguments"].(string)
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
		t.Errorf("arguments should be valid JSON: %v", err)
	}
	if args["location"] != "San Francisco" {
		t.Errorf("expected location=San Francisco, got %v", args["location"])
	}
}

// Test Anthropic -> OpenAI request transformation with tool_result
func TestOpenAITransformer_TransformRequest_ToolResult(t *testing.T) {
	transformer := &OpenAITransformer{}

	anthropicReq := `{
		"model": "claude-3-5-sonnet-20241022",
		"messages": [
			{
				"role": "user",
				"content": "What's the weather?"
			},
			{
				"role": "assistant",
				"content": [
					{
						"type": "tool_use",
						"id": "toolu_123",
						"name": "get_weather",
						"input": {"location": "SF"}
					}
				]
			},
			{
				"role": "user",
				"content": [
					{
						"type": "tool_result",
						"tool_use_id": "toolu_123",
						"content": "72°F, sunny"
					}
				]
			}
		]
	}`

	result, err := transformer.TransformRequest([]byte(anthropicReq), "anthropic")
	if err != nil {
		t.Fatalf("TransformRequest failed: %v", err)
	}

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	messages := openaiReq["messages"].([]interface{})
	
	// Should have: user, assistant, tool
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages (user, assistant, tool), got %d", len(messages))
	}

	// Check tool message
	toolMsg := messages[2].(map[string]interface{})
	if toolMsg["role"] != "tool" {
		t.Errorf("expected role=tool, got %v", toolMsg["role"])
	}

	if toolMsg["tool_call_id"] != "toolu_123" {
		t.Errorf("expected tool_call_id=toolu_123, got %v", toolMsg["tool_call_id"])
	}

	if toolMsg["content"] != "72°F, sunny" {
		t.Errorf("expected content='72°F, sunny', got %v", toolMsg["content"])
	}
}

// Test Anthropic -> OpenAI with multiple text blocks (should concatenate)
func TestOpenAITransformer_TransformRequest_MultipleTextBlocks(t *testing.T) {
	transformer := &OpenAITransformer{}

	anthropicReq := `{
		"model": "claude-3-5-sonnet-20241022",
		"messages": [
			{
				"role": "assistant",
				"content": [
					{"type": "text", "text": "First part."},
					{"type": "text", "text": "Second part."},
					{"type": "text", "text": "Third part."}
				]
			}
		]
	}`

	result, err := transformer.TransformRequest([]byte(anthropicReq), "anthropic-messages")
	if err != nil {
		t.Fatalf("TransformRequest failed: %v", err)
	}

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	messages := openaiReq["messages"].([]interface{})
	assistantMsg := messages[0].(map[string]interface{})
	content := assistantMsg["content"].(string)

	// Should concatenate all text blocks
	if !strings.Contains(content, "First part.") || !strings.Contains(content, "Second part.") || !strings.Contains(content, "Third part.") {
		t.Errorf("expected all text blocks concatenated, got: %s", content)
	}
}

// Test Anthropic -> OpenAI with mixed text and tool_result (should preserve both)
func TestOpenAITransformer_TransformRequest_MixedTextAndToolResult(t *testing.T) {
	transformer := &OpenAITransformer{}

	anthropicReq := `{
		"model": "claude-3-5-sonnet-20241022",
		"messages": [
			{
				"role": "user",
				"content": "Initial question"
			},
			{
				"role": "assistant",
				"content": [
					{
						"type": "tool_use",
						"id": "toolu_123",
						"name": "get_weather",
						"input": {"location": "SF"}
					}
				]
			},
			{
				"role": "user",
				"content": [
					{"type": "text", "text": "Here's some context:"},
					{
						"type": "tool_result",
						"tool_use_id": "toolu_123",
						"content": "72°F, sunny"
					},
					{"type": "text", "text": "What do you think?"}
				]
			}
		]
	}`

	result, err := transformer.TransformRequest([]byte(anthropicReq), "anthropic")
	if err != nil {
		t.Fatalf("TransformRequest failed: %v", err)
	}

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	messages := openaiReq["messages"].([]interface{})

	// Should have: user, assistant, user (text1), tool, user (text2)
	// Total: 5 messages preserving original ordering
	if len(messages) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(messages))
	}

	// Verify ordering is preserved: text -> tool_result -> text
	// Message 0: user("Initial question")
	// Message 1: assistant(tool_use)
	// Message 2: user("Here's some context:")
	// Message 3: tool(result)
	// Message 4: user("What do you think?")

	// Check message 2: first text block
	msg2 := messages[2].(map[string]interface{})
	if msg2["role"] != "user" {
		t.Errorf("message 2 should be user, got %s", msg2["role"])
	}
	if content, ok := msg2["content"].(string); !ok || content != "Here's some context:" {
		t.Errorf("message 2 should have first text content, got: %v", msg2["content"])
	}

	// Check message 3: tool result
	msg3 := messages[3].(map[string]interface{})
	if msg3["role"] != "tool" {
		t.Errorf("message 3 should be tool, got %s", msg3["role"])
	}
	if msg3["tool_call_id"] != "toolu_123" {
		t.Errorf("message 3 should have tool_call_id toolu_123, got: %v", msg3["tool_call_id"])
	}

	// Check message 4: second text block
	msg4 := messages[4].(map[string]interface{})
	if msg4["role"] != "user" {
		t.Errorf("message 4 should be user, got %s", msg4["role"])
	}
	if content, ok := msg4["content"].(string); !ok || content != "What do you think?" {
		t.Errorf("message 4 should have second text content, got: %v", msg4["content"])
	}
}

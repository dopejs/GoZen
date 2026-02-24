package transform

import (
	"io"
	"strings"
	"testing"
)

func TestStreamTransformer_NoTransformNeeded(t *testing.T) {
	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "anthropic",
	}

	input := "data: test\n\n"
	reader := st.TransformSSEStream(strings.NewReader(input))

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(output) != input {
		t.Errorf("expected passthrough, got %q", string(output))
	}
}

func TestStreamTransformer_OpenAIToAnthropic(t *testing.T) {
	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai",
	}

	// Simulate OpenAI Chat Completions SSE stream
	input := `data: {"id":"chatcmpl-123","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":2}}

data: [DONE]

`

	reader := st.TransformSSEStream(strings.NewReader(input))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := string(output)

	// Verify Anthropic SSE events are present
	if !strings.Contains(outputStr, "event: message_start") {
		t.Error("missing message_start event")
	}
	if !strings.Contains(outputStr, "event: content_block_start") {
		t.Error("missing content_block_start event")
	}
	if !strings.Contains(outputStr, "event: content_block_delta") {
		t.Error("missing content_block_delta event")
	}
	if !strings.Contains(outputStr, `"text":"Hello"`) {
		t.Error("missing Hello text delta")
	}
	if !strings.Contains(outputStr, `"text":" world"`) {
		t.Error("missing world text delta")
	}
	if !strings.Contains(outputStr, "event: content_block_stop") {
		t.Error("missing content_block_stop event")
	}
	if !strings.Contains(outputStr, "event: message_delta") {
		t.Error("missing message_delta event")
	}
	if !strings.Contains(outputStr, `"stop_reason":"end_turn"`) {
		t.Error("missing stop_reason in message_delta")
	}
	if !strings.Contains(outputStr, "event: message_stop") {
		t.Error("missing message_stop event")
	}
}

func TestStreamTransformer_OpenAIToAnthropic_FinishReasonMapping(t *testing.T) {
	tests := []struct {
		name         string
		finishReason string
		wantStop     string
	}{
		{"stop", "stop", "end_turn"},
		{"length", "length", "max_tokens"},
		{"tool_calls", "tool_calls", "tool_use"},
		{"content_filter", "content_filter", "end_turn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &StreamTransformer{
				ClientFormat:   "anthropic",
				ProviderFormat: "openai",
			}

			input := `data: {"id":"chatcmpl-123","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant"}}]}

data: {"id":"chatcmpl-123","model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hi"}}]}

data: {"id":"chatcmpl-123","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"` + tt.finishReason + `"}]}

`

			reader := st.TransformSSEStream(strings.NewReader(input))
			output, _ := io.ReadAll(reader)

			if !strings.Contains(string(output), `"stop_reason":"`+tt.wantStop+`"`) {
				t.Errorf("expected stop_reason %q, got output: %s", tt.wantStop, string(output))
			}
		})
	}
}

func TestStreamTransformer_AnthropicToOpenAI(t *testing.T) {
	st := &StreamTransformer{
		ClientFormat:   "openai",
		ProviderFormat: "anthropic",
	}

	// Simulate Anthropic Messages SSE stream
	input := `event: message_start
data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-sonnet","content":[],"usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":2}}

event: message_stop
data: {"type":"message_stop"}

`

	reader := st.TransformSSEStream(strings.NewReader(input))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := string(output)

	// Verify OpenAI Responses API events are present
	if !strings.Contains(outputStr, "event: response.created") {
		t.Error("missing response.created event")
	}
	if !strings.Contains(outputStr, "event: response.output_text.delta") {
		t.Error("missing response.output_text.delta event")
	}
	if !strings.Contains(outputStr, `"delta":"Hello"`) {
		t.Error("missing Hello delta")
	}
	if !strings.Contains(outputStr, `"delta":" world"`) {
		t.Error("missing world delta")
	}
	if !strings.Contains(outputStr, "event: response.completed") {
		t.Error("missing response.completed event")
	}
}

func TestStreamTransformer_EmptyStream(t *testing.T) {
	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai",
	}

	reader := st.TransformSSEStream(strings.NewReader(""))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output) != 0 {
		t.Errorf("expected empty output for empty input, got %q", string(output))
	}
}

func TestStreamTransformer_InvalidJSON(t *testing.T) {
	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai",
	}

	input := `data: not valid json

data: {"id":"chatcmpl-123","model":"gpt-4","choices":[{"delta":{"content":"Hi"}}]}

data: [DONE]

`

	reader := st.TransformSSEStream(strings.NewReader(input))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still process valid chunks
	if !strings.Contains(string(output), `"text":"Hi"`) {
		t.Error("should process valid JSON chunks even with invalid ones")
	}
}

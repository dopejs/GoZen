package transform

import (
	"encoding/json"
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

func TestTransformResponsesAPIToAnthropic_Text(t *testing.T) {
	// Simulate Responses API SSE stream with text output
	input := strings.Join([]string{
		`event: response.created`,
		`data: {"type":"response.created","response":{"id":"resp_1","object":"response","status":"in_progress","model":"gpt-5","output":[]}}`,
		``,
		`event: response.output_item.added`,
		`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg_1","type":"message","role":"assistant","content":[]}}`,
		``,
		`event: response.content_part.added`,
		`data: {"type":"response.content_part.added","item_id":"msg_1","output_index":0,"content_index":0,"part":{"type":"output_text","text":""}}`,
		``,
		`event: response.output_text.delta`,
		`data: {"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":"Hello"}`,
		``,
		`event: response.output_text.delta`,
		`data: {"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":" world"}`,
		``,
		`event: response.output_text.delta`,
		`data: {"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":"!"}`,
		``,
		`event: response.output_item.done`,
		`data: {"type":"response.output_item.done","output_index":0,"item":{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello world!"}]}}`,
		``,
		`event: response.completed`,
		`data: {"type":"response.completed","response":{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[],"usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}}`,
		``,
	}, "\n")

	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai-responses",
	}
	reader := st.TransformSSEStream(strings.NewReader(input))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := string(output)

	// Verify Anthropic SSE events
	if !strings.Contains(result, "event: message_start") {
		t.Error("should emit message_start event")
	}
	if !strings.Contains(result, "event: content_block_start") {
		t.Error("should emit content_block_start event")
	}
	if !strings.Contains(result, `"text_delta"`) {
		t.Error("should emit content_block_delta with text_delta")
	}
	if !strings.Contains(result, `"Hello"`) {
		t.Error("should include first delta text")
	}
	if !strings.Contains(result, `" world"`) {
		t.Error("should include second delta text")
	}
	if !strings.Contains(result, "event: content_block_stop") {
		t.Error("should emit content_block_stop event")
	}
	if !strings.Contains(result, "event: message_delta") {
		t.Error("should emit message_delta event")
	}
	if !strings.Contains(result, `"end_turn"`) {
		t.Error("should include stop_reason end_turn")
	}
	if !strings.Contains(result, "event: message_stop") {
		t.Error("should emit message_stop event")
	}
}

func TestTransformResponsesAPIToAnthropic_ToolCall(t *testing.T) {
	// Simulate Responses API SSE stream with function_call output
	input := strings.Join([]string{
		`event: response.created`,
		`data: {"type":"response.created","response":{"id":"resp_2","object":"response","status":"in_progress","model":"gpt-5","output":[]}}`,
		``,
		`event: response.output_item.added`,
		`data: {"type":"response.output_item.added","output_index":0,"item":{"id":"fc_1","type":"function_call","call_id":"call_1","name":"get_weather","arguments":""}}`,
		``,
		`event: response.function_call_arguments.delta`,
		`data: {"type":"response.function_call_arguments.delta","item_id":"fc_1","output_index":0,"delta":"{\"loc"}`,
		``,
		`event: response.function_call_arguments.delta`,
		`data: {"type":"response.function_call_arguments.delta","item_id":"fc_1","output_index":0,"delta":"ation\":\"Paris\"}"}`,
		``,
		`event: response.output_item.done`,
		`data: {"type":"response.output_item.done","output_index":0,"item":{"id":"fc_1","type":"function_call","call_id":"call_1","name":"get_weather","arguments":"{\"location\":\"Paris\"}","status":"completed"}}`,
		``,
		`event: response.completed`,
		`data: {"type":"response.completed","response":{"id":"resp_2","object":"response","status":"completed","model":"gpt-5","output":[],"usage":{"input_tokens":20,"output_tokens":10,"total_tokens":30}}}`,
		``,
	}, "\n")

	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai-responses",
	}
	reader := st.TransformSSEStream(strings.NewReader(input))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := string(output)

	// Verify tool call events
	if !strings.Contains(result, "event: message_start") {
		t.Error("should emit message_start event")
	}
	if !strings.Contains(result, "event: content_block_start") {
		t.Error("should emit content_block_start event")
	}
	if !strings.Contains(result, `"tool_use"`) {
		t.Error("should emit content_block_start with type tool_use")
	}
	if !strings.Contains(result, `"get_weather"`) {
		t.Error("should include tool name")
	}
	if !strings.Contains(result, `"input_json_delta"`) {
		t.Error("should emit content_block_delta with input_json_delta")
	}
	if !strings.Contains(result, "event: content_block_stop") {
		t.Error("should emit content_block_stop event")
	}
	if !strings.Contains(result, `"tool_use"`) {
		t.Error("should include stop_reason tool_use in message_delta")
	}
}

func TestStreamTransformerRouting(t *testing.T) {
	tests := []struct {
		name           string
		clientFormat   string
		providerFormat string
		wantPassthrough bool
	}{
		{
			name:           "openai-chat to anthropic",
			clientFormat:   FormatOpenAIChat,
			providerFormat: "anthropic",
			wantPassthrough: false,
		},
		{
			name:           "openai-responses to anthropic",
			clientFormat:   FormatOpenAIResponses,
			providerFormat: "anthropic",
			wantPassthrough: false,
		},
		{
			name:           "anthropic-messages to openai",
			clientFormat:   FormatAnthropicMessages,
			providerFormat: "openai",
			wantPassthrough: false,
		},
		{
			name:           "same format passthrough",
			clientFormat:   FormatAnthropicMessages,
			providerFormat: "anthropic",
			wantPassthrough: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &StreamTransformer{
				ClientFormat:   tt.clientFormat,
				ProviderFormat: tt.providerFormat,
				MessageID:      "test-id",
				Model:          "test-model",
			}

			input := "event: message_start\ndata: {\"type\":\"message_start\"}\n\n"
			reader := st.TransformSSEStream(strings.NewReader(input))
			output, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result := string(output)
			if tt.wantPassthrough {
				if result != input {
					t.Errorf("expected passthrough, got transformation")
				}
			} else {
				// Verify transformation occurred (output differs from input)
				if result == input {
					t.Errorf("expected transformation, got passthrough")
				}
			}
		})
	}
}

func TestStreamTransformer_AnthropicToolUseToOpenAIChat(t *testing.T) {
	// Anthropic streaming tool_use events
	input := `event: message_start
data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-sonnet-4-5"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_abc","name":"get_weather"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"location\":"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"SF\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}

`

	st := &StreamTransformer{
		ClientFormat:   FormatOpenAIChat,
		ProviderFormat: "anthropic",
		MessageID:      "chatcmpl_123",
		Model:          "claude-sonnet-4-5",
	}

	reader := st.TransformSSEStream(strings.NewReader(input))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := string(output)

	// Verify tool_calls delta events
	if !strings.Contains(result, `"tool_calls"`) {
		t.Error("should emit tool_calls in delta")
	}
	if !strings.Contains(result, `"get_weather"`) {
		t.Error("should include tool name in delta")
	}
	if !strings.Contains(result, `"arguments"`) {
		t.Error("should emit function arguments delta")
	}
	if !strings.Contains(result, `"finish_reason":"tool_calls"`) {
		t.Error("should set finish_reason to tool_calls")
	}
}

func TestStreamTransformer_AnthropicToolUseToOpenAIResponses(t *testing.T) {
	// Anthropic streaming tool_use events
	input := `event: message_start
data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-sonnet-4-5"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_abc","name":"get_weather"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"location\":\"SF\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_stop
data: {"type":"message_stop"}

`

	st := &StreamTransformer{
		ClientFormat:   FormatOpenAIResponses,
		ProviderFormat: "anthropic",
		MessageID:      "resp_123",
		Model:          "claude-sonnet-4-5",
	}

	reader := st.TransformSSEStream(strings.NewReader(input))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := string(output)

	// Verify Responses API function_call_arguments.delta events
	if !strings.Contains(result, `function_call_arguments`) {
		t.Error("should emit function_call_arguments in Responses API format")
	}
	if !strings.Contains(result, `delta`) {
		t.Error("should emit delta field")
	}
}

// Phase 5: SSE Error Handling Tests

// truncatedReader simulates a stream that ends abruptly with an error
type truncatedReader struct {
	data []byte
	pos  int
}

func (r *truncatedReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// T024: Anthropic to OpenAI - truncated stream should emit error event
func TestStreamTransformer_ErrorHandling_AnthropicToOpenAI(t *testing.T) {
	// Truncated Anthropic stream (missing message_stop)
	input := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-sonnet","content":[],"usage":{"input_tokens":10,"output_tokens":0}}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
		``,
		// Stream ends abruptly here
	}, "\n")

	tr := &truncatedReader{data: []byte(input)}
	st := &StreamTransformer{
		ClientFormat:   "openai",
		ProviderFormat: "anthropic",
	}
	reader := st.TransformSSEStream(tr)
	output, _ := io.ReadAll(reader)
	result := string(output)

	// Should emit error event instead of completion
	if strings.Contains(result, "response.completed") {
		t.Error("should emit error event instead of completion")
	}
}

// T025: OpenAI to Anthropic - truncated stream should emit error event
func TestStreamTransformer_ErrorHandling_OpenAIToAnthropic(t *testing.T) {
	input := strings.Join([]string{
		`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
		``,
		// Stream ends abruptly
	}, "\n")

	tr := &truncatedReader{data: []byte(input)}
	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai",
	}
	reader := st.TransformSSEStream(tr)
	output, _ := io.ReadAll(reader)
	result := string(output)

	// Should emit error event instead of completion
	if strings.Contains(result, "message_stop") {
		t.Error("should emit error event instead of completion")
	}
}

// T026: Responses API to Anthropic - truncated stream should emit error event
func TestStreamTransformer_ErrorHandling_ResponsesAPIToAnthropic(t *testing.T) {
	input := strings.Join([]string{
		`event: response.created`,
		`data: {"type":"response.created","response":{"id":"resp_1","object":"response","status":"in_progress","model":"gpt-5","output":[]}}`,
		``,
		// Stream ends abruptly
	}, "\n")

	tr := &truncatedReader{data: []byte(input)}
	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai-responses",
	}
	reader := st.TransformSSEStream(tr)
	output, _ := io.ReadAll(reader)
	result := string(output)

	// Should emit error event instead of completion
	if strings.Contains(result, "message_stop") {
		t.Error("should emit error event instead of completion")
	}
}

// T027: Clean EOF should emit correct completion event
func TestStreamTransformer_ErrorHandling_CleanEOF(t *testing.T) {
	input := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-sonnet","content":[],"usage":{"input_tokens":10,"output_tokens":0}}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")

	st := &StreamTransformer{
		ClientFormat:   "openai",
		ProviderFormat: "anthropic",
	}
	reader := st.TransformSSEStream(strings.NewReader(input))
	output, _ := io.ReadAll(reader)
	result := string(output)

	// Should emit completion event (not error)
	if !strings.Contains(result, "response.completed") {
		t.Error("should emit completion event for clean EOF")
	}
}

// Test OpenAI Chat SSE -> Anthropic SSE with streaming tool_calls
func TestStreamTransformer_OpenAIChatToAnthropic_StreamingToolCalls(t *testing.T) {
	// Simulate OpenAI streaming response with tool_calls
	openaiStream := `data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_abc123","type":"function","function":{"name":"get_weather","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"location\""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":":\"SF\"}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`

	st := &StreamTransformer{
		ClientFormat:   "anthropic-messages",
		ProviderFormat: "openai-chat",
	}

	reader := st.TransformSSEStream(strings.NewReader(openaiStream))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	outputStr := string(output)

	// Verify message_start
	if !strings.Contains(outputStr, "event: message_start") {
		t.Error("expected message_start event")
	}

	// Verify content_block_start with tool_use
	if !strings.Contains(outputStr, "\"type\":\"tool_use\"") {
		t.Error("expected tool_use content block")
	}

	if !strings.Contains(outputStr, "\"name\":\"get_weather\"") {
		t.Error("expected tool name get_weather")
	}

	if !strings.Contains(outputStr, "\"id\":\"call_abc123\"") {
		t.Error("expected tool call id")
	}

	// Verify input_json_delta events
	if !strings.Contains(outputStr, "\"type\":\"input_json_delta\"") {
		t.Error("expected input_json_delta events")
	}

	if !strings.Contains(outputStr, "\"partial_json\"") {
		t.Error("expected partial_json in delta")
	}

	// Verify content_block_stop
	if !strings.Contains(outputStr, "event: content_block_stop") {
		t.Error("expected content_block_stop event")
	}

	// Verify stop_reason is tool_use
	if !strings.Contains(outputStr, "\"stop_reason\":\"tool_use\"") {
		t.Error("expected stop_reason=tool_use")
	}

	// Verify message_delta with usage
	if !strings.Contains(outputStr, "event: message_delta") {
		t.Error("expected message_delta event")
	}
}

// Test OpenAI Chat SSE -> Anthropic SSE with text then tool_calls
func TestStreamTransformer_OpenAIChatToAnthropic_TextThenToolCalls(t *testing.T) {
	openaiStream := `data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"role":"assistant","content":"Let me check"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"content":" that"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_123","type":"function","function":{"name":"get_weather","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"loc\":\"NYC\"}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`

	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai-chat",
	}

	reader := st.TransformSSEStream(strings.NewReader(openaiStream))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	outputStr := string(output)

	// Should have text content block first
	if !strings.Contains(outputStr, "\"type\":\"text\"") {
		t.Error("expected text content block")
	}

	if !strings.Contains(outputStr, "\"text\":\"Let me check\"") {
		t.Error("expected text content")
	}

	// Then tool_use block
	if !strings.Contains(outputStr, "\"type\":\"tool_use\"") {
		t.Error("expected tool_use content block")
	}

	// Should have two content_block_stop events (one for text, one for tool)
	stopCount := strings.Count(outputStr, "event: content_block_stop")
	if stopCount < 2 {
		t.Errorf("expected at least 2 content_block_stop events, got %d", stopCount)
	}
}

// Test that content block indices are correctly assigned (text=0, tool=1)
// and that blocks are closed before opening new ones
func TestStreamTransformer_OpenAIChatToAnthropic_ContentBlockIndices(t *testing.T) {
	openaiStream := `data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"role":"assistant","content":"Text first"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_123","type":"function","function":{"name":"get_weather","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"loc\":\"NYC\"}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`

	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai-chat",
	}

	reader := st.TransformSSEStream(strings.NewReader(openaiStream))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	outputStr := string(output)

	// Parse events to verify indices and lifecycle
	events := parseSSEEvents(outputStr)

	// Find content_block_start events
	var textBlockIndex, toolBlockIndex int = -1, -1
	for _, event := range events {
		if event.Type == "content_block_start" {
			if blockType, ok := event.Data["content_block"].(map[string]interface{})["type"].(string); ok {
				index := int(event.Data["index"].(float64))
				if blockType == "text" {
					textBlockIndex = index
				} else if blockType == "tool_use" {
					toolBlockIndex = index
				}
			}
		}
	}

	// Verify text block is index 0
	if textBlockIndex != 0 {
		t.Errorf("expected text block at index 0, got %d", textBlockIndex)
	}

	// Verify tool block is index 1 (after text)
	if toolBlockIndex != 1 {
		t.Errorf("expected tool block at index 1, got %d", toolBlockIndex)
	}

	// Verify strict lifecycle: text block must be stopped before tool block starts
	var textStartPos, textStopPos, toolStartPos int = -1, -1, -1
	for i, event := range events {
		if event.Type == "content_block_start" {
			if blockType, ok := event.Data["content_block"].(map[string]interface{})["type"].(string); ok {
				index := int(event.Data["index"].(float64))
				if blockType == "text" && index == 0 {
					textStartPos = i
				} else if blockType == "tool_use" && index == 1 {
					toolStartPos = i
				}
			}
		} else if event.Type == "content_block_stop" {
			index := int(event.Data["index"].(float64))
			if index == 0 && textStopPos == -1 {
				textStopPos = i
			}
		}
	}

	// Verify lifecycle order: start(text) < stop(text) < start(tool)
	if textStartPos == -1 {
		t.Error("text block start not found")
	}
	if textStopPos == -1 {
		t.Error("text block stop not found")
	}
	if toolStartPos == -1 {
		t.Error("tool block start not found")
	}

	if textStopPos <= textStartPos {
		t.Errorf("text block stop (%d) should come after start (%d)", textStopPos, textStartPos)
	}
	if toolStartPos <= textStopPos {
		t.Errorf("tool block start (%d) should come after text block stop (%d)", toolStartPos, textStopPos)
	}
}

// Test parallel tool calls lifecycle: each tool block must be closed before the next starts
func TestStreamTransformer_OpenAIChatToAnthropic_ParallelToolCallsLifecycle(t *testing.T) {
	openaiStream := `data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_abc","type":"function","function":{"name":"tool_a","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"x\":1}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"id":"call_def","type":"function","function":{"name":"tool_b","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"function":{"arguments":"{\"y\":2}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`

	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai-chat",
	}

	reader := st.TransformSSEStream(strings.NewReader(openaiStream))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	outputStr := string(output)

	// Parse events to verify strict lifecycle
	events := parseSSEEvents(outputStr)

	// Find tool block start/stop positions
	var tool0StartPos, tool0StopPos, tool1StartPos, tool1StopPos int = -1, -1, -1, -1
	var tool0Index, tool1Index int = -1, -1

	for i, event := range events {
		if event.Type == "content_block_start" {
			if blockType, ok := event.Data["content_block"].(map[string]interface{})["type"].(string); ok {
				if blockType == "tool_use" {
					index := int(event.Data["index"].(float64))
					if tool0Index == -1 {
						tool0Index = index
						tool0StartPos = i
					} else if tool1Index == -1 {
						tool1Index = index
						tool1StartPos = i
					}
				}
			}
		} else if event.Type == "content_block_stop" {
			index := int(event.Data["index"].(float64))
			if tool0Index != -1 && index == tool0Index && tool0StopPos == -1 {
				tool0StopPos = i
			} else if tool1Index != -1 && index == tool1Index && tool1StopPos == -1 {
				tool1StopPos = i
			}
		}
	}

	// Verify both tools were found
	if tool0StartPos == -1 {
		t.Error("tool 0 start not found")
	}
	if tool0StopPos == -1 {
		t.Error("tool 0 stop not found")
	}
	if tool1StartPos == -1 {
		t.Error("tool 1 start not found")
	}
	if tool1StopPos == -1 {
		t.Error("tool 1 stop not found")
	}

	// Verify strict lifecycle: start(tool0) < stop(tool0) < start(tool1) < stop(tool1)
	if tool0StopPos <= tool0StartPos {
		t.Errorf("tool 0 stop (%d) should come after start (%d)", tool0StopPos, tool0StartPos)
	}
	if tool1StartPos <= tool0StopPos {
		t.Errorf("tool 1 start (%d) should come after tool 0 stop (%d)", tool1StartPos, tool0StopPos)
	}
	if tool1StopPos <= tool1StartPos {
		t.Errorf("tool 1 stop (%d) should come after start (%d)", tool1StopPos, tool1StartPos)
	}

	// Verify indices are sequential
	if tool1Index != tool0Index+1 {
		t.Errorf("expected tool 1 index (%d) to be tool 0 index (%d) + 1", tool1Index, tool0Index)
	}
}

// Test interleaved tool call deltas: deltas for closed blocks should be ignored
func TestStreamTransformer_OpenAIChatToAnthropic_InterleavedToolCallDeltas(t *testing.T) {
	// Simulate real parallel tool call scenario with interleaved deltas
	openaiStream := `data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_abc","type":"function","function":{"name":"tool_a","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"x\":"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"id":"call_def","type":"function","function":{"name":"tool_b","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"function":{"arguments":"{\"y\":"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"1}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"function":{"arguments":"2}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`

	st := &StreamTransformer{
		ClientFormat:   "anthropic",
		ProviderFormat: "openai-chat",
	}

	reader := st.TransformSSEStream(strings.NewReader(openaiStream))
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	outputStr := string(output)

	// Parse events
	events := parseSSEEvents(outputStr)

	// Track which Anthropic indices received deltas
	deltasByIndex := make(map[int][]string)
	for _, event := range events {
		if event.Type == "content_block_delta" {
			index := int(event.Data["index"].(float64))
			if delta, ok := event.Data["delta"].(map[string]interface{}); ok {
				if deltaType, ok := delta["type"].(string); ok {
					if deltaType == "input_json_delta" {
						if partialJSON, ok := delta["partial_json"].(string); ok {
							deltasByIndex[index] = append(deltasByIndex[index], partialJSON)
						}
					}
				}
			}
		}
	}

	// Find tool block indices
	var tool0Index, tool1Index int = -1, -1
	for _, event := range events {
		if event.Type == "content_block_start" {
			if blockType, ok := event.Data["content_block"].(map[string]interface{})["type"].(string); ok {
				if blockType == "tool_use" {
					index := int(event.Data["index"].(float64))
					if tool0Index == -1 {
						tool0Index = index
					} else if tool1Index == -1 {
						tool1Index = index
					}
				}
			}
		}
	}

	// Verify tool 0 only received deltas before it was closed
	// After tool 1 starts, tool 0 is closed, so no more deltas should go to tool 0
	tool0Deltas := deltasByIndex[tool0Index]
	if len(tool0Deltas) != 1 {
		t.Errorf("expected tool 0 to receive 1 delta (before being closed), got %d: %v", len(tool0Deltas), tool0Deltas)
	}
	if len(tool0Deltas) > 0 && tool0Deltas[0] != "{\"x\":" {
		t.Errorf("expected tool 0 first delta to be '{\"x\":', got %s", tool0Deltas[0])
	}

	// Verify tool 1 received all its deltas
	tool1Deltas := deltasByIndex[tool1Index]
	if len(tool1Deltas) != 2 {
		t.Errorf("expected tool 1 to receive 2 deltas, got %d: %v", len(tool1Deltas), tool1Deltas)
	}
	if len(tool1Deltas) >= 2 {
		if tool1Deltas[0] != "{\"y\":" {
			t.Errorf("expected tool 1 first delta to be '{\"y\":', got %s", tool1Deltas[0])
		}
		if tool1Deltas[1] != "2}" {
			t.Errorf("expected tool 1 second delta to be '2}', got %s", tool1Deltas[1])
		}
	}

	// Verify no deltas were sent to non-existent indices
	for idx := range deltasByIndex {
		if idx != tool0Index && idx != tool1Index {
			t.Errorf("unexpected deltas sent to index %d", idx)
		}
	}
}

// Helper to parse SSE events for testing
type sseEvent struct {
	Type string
	Data map[string]interface{}
}

func parseSSEEvents(output string) []sseEvent {
	var events []sseEvent
	lines := strings.Split(output, "\n")
	var currentEvent string
	var dataBuffer string

	for _, line := range lines {
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataBuffer = strings.TrimPrefix(line, "data: ")
		} else if line == "" && dataBuffer != "" {
			var data map[string]interface{}
			json.Unmarshal([]byte(dataBuffer), &data)
			events = append(events, sseEvent{
				Type: currentEvent,
				Data: data,
			})
			currentEvent = ""
			dataBuffer = ""
		}
	}

	return events
}

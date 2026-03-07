package transform

import (
	"encoding/json"
	"testing"
)

func TestChatCompletionsToResponsesAPI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkFn  func(t *testing.T, result map[string]interface{})
	}{
		{
			name:  "messages_renamed_to_input",
			input: `{"model":"gpt-5","messages":[{"role":"user","content":"hi"}]}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				if _, ok := result["messages"]; ok {
					t.Error("messages field should be removed")
				}
				input, ok := result["input"]
				if !ok {
					t.Fatal("input field should be present")
				}
				arr, ok := input.([]interface{})
				if !ok || len(arr) == 0 {
					t.Fatal("input should be a non-empty array")
				}
			},
		},
		{
			name:  "max_completion_tokens_renamed",
			input: `{"model":"gpt-5","messages":[],"max_completion_tokens":4096}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				if _, ok := result["max_completion_tokens"]; ok {
					t.Error("max_completion_tokens should be removed")
				}
				v, ok := result["max_output_tokens"]
				if !ok {
					t.Fatal("max_output_tokens should be present")
				}
				if int(v.(float64)) != 4096 {
					t.Errorf("max_output_tokens = %v, want 4096", v)
				}
			},
		},
		{
			name: "tool_flattening",
			input: `{"model":"gpt-5","messages":[],"tools":[{"type":"function","function":{"name":"get_weather","description":"Get weather","parameters":{"type":"object","properties":{}}}}]}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				tools, ok := result["tools"].([]interface{})
				if !ok || len(tools) == 0 {
					t.Fatal("tools should be a non-empty array")
				}
				tool := tools[0].(map[string]interface{})
				if tool["type"] != "function" {
					t.Errorf("tool type = %v, want function", tool["type"])
				}
				// Should be flattened: name at top level, no "function" wrapper
				if tool["name"] != "get_weather" {
					t.Errorf("tool name = %v, want get_weather", tool["name"])
				}
				if tool["description"] != "Get weather" {
					t.Errorf("tool description = %v, want Get weather", tool["description"])
				}
				if _, ok := tool["parameters"]; !ok {
					t.Error("tool should have parameters at top level")
				}
				if _, ok := tool["function"]; ok {
					t.Error("tool should NOT have function wrapper")
				}
			},
		},
		{
			name:  "unsupported_fields_removed",
			input: `{"model":"gpt-5","messages":[],"n":2,"logprobs":true,"stream_options":{"include_usage":true},"presence_penalty":0.5,"frequency_penalty":0.5,"seed":42,"response_format":{"type":"json_object"}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				removed := []string{"n", "logprobs", "stream_options", "presence_penalty", "frequency_penalty", "seed", "response_format"}
				for _, field := range removed {
					if _, ok := result[field]; ok {
						t.Errorf("field %q should be removed", field)
					}
				}
			},
		},
		{
			name:  "passthrough_fields",
			input: `{"model":"gpt-5","messages":[],"stream":true,"temperature":0.7,"top_p":0.9,"tool_choice":"auto","stop":["END"]}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				if result["model"] != "gpt-5" {
					t.Errorf("model = %v, want gpt-5", result["model"])
				}
				if result["stream"] != true {
					t.Errorf("stream = %v, want true", result["stream"])
				}
				if result["temperature"].(float64) != 0.7 {
					t.Errorf("temperature = %v, want 0.7", result["temperature"])
				}
				if result["top_p"].(float64) != 0.9 {
					t.Errorf("top_p = %v, want 0.9", result["top_p"])
				}
				if result["tool_choice"] != "auto" {
					t.Errorf("tool_choice = %v, want auto", result["tool_choice"])
				}
				stops := result["stop"].([]interface{})
				if len(stops) != 1 || stops[0] != "END" {
					t.Errorf("stop = %v, want [END]", result["stop"])
				}
			},
		},
		{
			name:  "store_set_to_false",
			input: `{"model":"gpt-5","messages":[]}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				v, ok := result["store"]
				if !ok {
					t.Fatal("store field should be present")
				}
				if v != false {
					t.Errorf("store = %v, want false", v)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ChatCompletionsToResponsesAPI([]byte(tt.input))
			if err != nil {
				t.Fatalf("ChatCompletionsToResponsesAPI() error: %v", err)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("failed to parse output: %v", err)
			}
			tt.checkFn(t, result)
		})
	}
}

func TestResponsesAPIToAnthropic(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "text_message_output",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello!"}]}],"usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				content := result["content"].([]interface{})
				if len(content) != 1 {
					t.Fatalf("content length = %d, want 1", len(content))
				}
				block := content[0].(map[string]interface{})
				if block["type"] != "text" {
					t.Errorf("content type = %v, want text", block["type"])
				}
				if block["text"] != "Hello!" {
					t.Errorf("text = %v, want Hello!", block["text"])
				}
			},
		},
		{
			name: "status_completed_to_end_turn",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[{"type":"message","content":[{"type":"output_text","text":"hi"}]}],"usage":{"input_tokens":1,"output_tokens":1}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				if result["stop_reason"] != "end_turn" {
					t.Errorf("stop_reason = %v, want end_turn", result["stop_reason"])
				}
			},
		},
		{
			name: "status_incomplete_to_max_tokens",
			input: `{"id":"resp_1","object":"response","status":"incomplete","model":"gpt-5","output":[{"type":"message","content":[{"type":"output_text","text":"partial"}]}],"usage":{"input_tokens":1,"output_tokens":100}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				if result["stop_reason"] != "max_tokens" {
					t.Errorf("stop_reason = %v, want max_tokens", result["stop_reason"])
				}
			},
		},
		{
			name: "usage_field_mapping",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[],"usage":{"input_tokens":100,"output_tokens":50,"total_tokens":150}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				usage := result["usage"].(map[string]interface{})
				if int(usage["input_tokens"].(float64)) != 100 {
					t.Errorf("input_tokens = %v, want 100", usage["input_tokens"])
				}
				if int(usage["output_tokens"].(float64)) != 50 {
					t.Errorf("output_tokens = %v, want 50", usage["output_tokens"])
				}
			},
		},
		{
			name: "missing_usage_zero_defaults",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[]}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				usage := result["usage"].(map[string]interface{})
				if int(usage["input_tokens"].(float64)) != 0 {
					t.Errorf("input_tokens = %v, want 0", usage["input_tokens"])
				}
				if int(usage["output_tokens"].(float64)) != 0 {
					t.Errorf("output_tokens = %v, want 0", usage["output_tokens"])
				}
			},
		},
		{
			name: "empty_output_array",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[],"usage":{"input_tokens":1,"output_tokens":0}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				content := result["content"].([]interface{})
				if len(content) != 0 {
					t.Errorf("content length = %d, want 0", len(content))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ResponsesAPIToAnthropic([]byte(tt.input))
			if err != nil {
				t.Fatalf("ResponsesAPIToAnthropic() error: %v", err)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("failed to parse output: %v", err)
			}
			tt.checkFn(t, result)
		})
	}
}

func TestResponsesAPIToAnthropic_ToolCall(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		checkFn func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "single_function_call",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[{"id":"fc_1","type":"function_call","call_id":"call_1","name":"get_weather","arguments":"{\"location\":\"Paris\"}","status":"completed"}],"usage":{"input_tokens":10,"output_tokens":5}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				content := result["content"].([]interface{})
				if len(content) != 1 {
					t.Fatalf("content length = %d, want 1", len(content))
				}
				block := content[0].(map[string]interface{})
				if block["type"] != "tool_use" {
					t.Errorf("content type = %v, want tool_use", block["type"])
				}
				if block["id"] != "call_1" {
					t.Errorf("id = %v, want call_1", block["id"])
				}
				if block["name"] != "get_weather" {
					t.Errorf("name = %v, want get_weather", block["name"])
				}
				input := block["input"].(map[string]interface{})
				if input["location"] != "Paris" {
					t.Errorf("input.location = %v, want Paris", input["location"])
				}
			},
		},
		{
			name: "mixed_message_and_function_call",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"Let me check."}]},{"id":"fc_1","type":"function_call","call_id":"call_2","name":"search","arguments":"{\"q\":\"test\"}","status":"completed"}],"usage":{"input_tokens":10,"output_tokens":5}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				content := result["content"].([]interface{})
				if len(content) != 2 {
					t.Fatalf("content length = %d, want 2", len(content))
				}
				text := content[0].(map[string]interface{})
				if text["type"] != "text" {
					t.Errorf("first content type = %v, want text", text["type"])
				}
				tool := content[1].(map[string]interface{})
				if tool["type"] != "tool_use" {
					t.Errorf("second content type = %v, want tool_use", tool["type"])
				}
			},
		},
		{
			name: "function_call_sets_stop_reason_tool_use",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[{"id":"fc_1","type":"function_call","call_id":"call_1","name":"test","arguments":"{}","status":"completed"}],"usage":{"input_tokens":1,"output_tokens":1}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				if result["stop_reason"] != "tool_use" {
					t.Errorf("stop_reason = %v, want tool_use", result["stop_reason"])
				}
			},
		},
		{
			name: "malformed_arguments_json",
			input: `{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[{"id":"fc_1","type":"function_call","call_id":"call_1","name":"test","arguments":"not valid json","status":"completed"}],"usage":{"input_tokens":1,"output_tokens":1}}`,
			checkFn: func(t *testing.T, result map[string]interface{}) {
				content := result["content"].([]interface{})
				block := content[0].(map[string]interface{})
				input := block["input"].(map[string]interface{})
				if len(input) != 0 {
					t.Errorf("input should be empty object for malformed JSON, got %v", input)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ResponsesAPIToAnthropic([]byte(tt.input))
			if err != nil {
				t.Fatalf("ResponsesAPIToAnthropic() error: %v", err)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("failed to parse output: %v", err)
			}
			tt.checkFn(t, result)
		})
	}
}

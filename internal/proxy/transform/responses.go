package transform

import "encoding/json"

// ChatCompletionsToResponsesAPI transforms an OpenAI Chat Completions request body
// to OpenAI Responses API format. Key changes:
// - messages → input
// - max_completion_tokens → max_output_tokens
// - tools: flatten (remove "function" wrapper)
// - remove unsupported fields (n, logprobs, stream_options, etc.)
// - set store: false to prevent server-side storage
func ChatCompletionsToResponsesAPI(body []byte) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body, err
	}

	// messages → input
	if messages, ok := data["messages"]; ok {
		data["input"] = messages
		delete(data, "messages")
	}

	// max_completion_tokens → max_output_tokens
	if v, ok := data["max_completion_tokens"]; ok {
		data["max_output_tokens"] = v
		delete(data, "max_completion_tokens")
	}

	// Flatten tools: remove "function" wrapper
	if tools, ok := data["tools"].([]interface{}); ok {
		flattened := make([]interface{}, 0, len(tools))
		for _, tool := range tools {
			toolMap, ok := tool.(map[string]interface{})
			if !ok {
				flattened = append(flattened, tool)
				continue
			}
			fn, ok := toolMap["function"].(map[string]interface{})
			if !ok {
				flattened = append(flattened, tool)
				continue
			}
			flat := map[string]interface{}{
				"type": toolMap["type"],
			}
			for k, v := range fn {
				flat[k] = v
			}
			flattened = append(flattened, flat)
		}
		data["tools"] = flattened
	}

	// Remove Chat Completions-only fields
	unsupported := []string{"n", "logprobs", "stream_options", "presence_penalty", "frequency_penalty", "seed", "response_format"}
	for _, field := range unsupported {
		delete(data, field)
	}

	// Prevent server-side storage on provider
	data["store"] = false

	return json.Marshal(data)
}

// ResponsesAPIToAnthropic transforms an OpenAI Responses API response body
// to Anthropic Messages API format.
func ResponsesAPIToAnthropic(body []byte) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body, err
	}

	anthropic := map[string]interface{}{
		"id":    data["id"],
		"type":  "message",
		"role":  "assistant",
		"model": data["model"],
	}

	// Map status → stop_reason
	switch data["status"] {
	case "completed":
		anthropic["stop_reason"] = "end_turn"
	case "incomplete":
		anthropic["stop_reason"] = "max_tokens"
	default:
		anthropic["stop_reason"] = "end_turn"
	}

	// Extract content from output array
	var content []interface{}
	hasToolUse := false

	if output, ok := data["output"].([]interface{}); ok {
		for _, item := range output {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			itemType, _ := itemMap["type"].(string)

			switch itemType {
			case "message":
				// Extract text content from message output
				if parts, ok := itemMap["content"].([]interface{}); ok {
					for _, part := range parts {
						partMap, ok := part.(map[string]interface{})
						if !ok {
							continue
						}
						if partMap["type"] == "output_text" {
							content = append(content, map[string]interface{}{
								"type": "text",
								"text": partMap["text"],
							})
						}
					}
				}

			case "function_call":
				hasToolUse = true
				// Parse arguments JSON string to object
				var inputObj interface{} = map[string]interface{}{}
				if argsStr, ok := itemMap["arguments"].(string); ok {
					if err := json.Unmarshal([]byte(argsStr), &inputObj); err != nil {
						inputObj = map[string]interface{}{}
					}
				}
				content = append(content, map[string]interface{}{
					"type":  "tool_use",
					"id":    itemMap["call_id"],
					"name":  itemMap["name"],
					"input": inputObj,
				})
			}
		}
	}

	if content == nil {
		content = []interface{}{}
	}
	anthropic["content"] = content

	// Override stop_reason if tool calls present
	if hasToolUse {
		anthropic["stop_reason"] = "tool_use"
	}

	// Map usage
	inputTokens := 0.0
	outputTokens := 0.0
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		if v, ok := usage["input_tokens"].(float64); ok {
			inputTokens = v
		}
		if v, ok := usage["output_tokens"].(float64); ok {
			outputTokens = v
		}
	}
	anthropic["usage"] = map[string]interface{}{
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
	}

	return json.Marshal(anthropic)
}

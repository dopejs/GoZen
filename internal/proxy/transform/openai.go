package transform

import "encoding/json"

// OpenAITransformer handles OpenAI Chat Completions API format.
// This is used by Codex and other OpenAI-compatible clients.
type OpenAITransformer struct{}

func (t *OpenAITransformer) Name() string {
	return "openai"
}

// TransformRequest transforms a request to OpenAI format.
// If the client is already using OpenAI format, no transformation is needed.
// If the client is using Anthropic format, convert to OpenAI format.
func (t *OpenAITransformer) TransformRequest(body []byte, clientFormat string) ([]byte, error) {
	// Normalize format
	normalized := NormalizeFormat(clientFormat)
	if normalized == "openai" || normalized == "" {
		// No transformation needed (empty defaults to openai for this transformer)
		return body, nil
	}

	// Anthropic → OpenAI transformation
	data, err := parseJSON(body)
	if err != nil {
		return body, nil
	}

	// Transform max_tokens → max_completion_tokens
	if maxTokens, ok := data["max_tokens"]; ok {
		data["max_completion_tokens"] = maxTokens
		delete(data, "max_tokens")
	}

	// Transform stop_sequences → stop
	if stopSeq, ok := data["stop_sequences"]; ok {
		data["stop"] = stopSeq
		delete(data, "stop_sequences")
	}

	// Remove Anthropic-specific fields
	delete(data, "metadata")
	delete(data, "thinking")

	// Transform tools format if present
	// Anthropic tools format is different from OpenAI
	if tools, ok := data["tools"].([]interface{}); ok {
		openAITools := make([]interface{}, 0, len(tools))
		for _, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				openAITool := map[string]interface{}{
					"type": "function",
					"function": map[string]interface{}{
						"name":        toolMap["name"],
						"description": toolMap["description"],
					},
				}
				if inputSchema, ok := toolMap["input_schema"]; ok {
					openAITool["function"].(map[string]interface{})["parameters"] = inputSchema
				}
				openAITools = append(openAITools, openAITool)
			}
		}
		if len(openAITools) > 0 {
			data["tools"] = openAITools
		}
	}

	// Transform system message format
	// Anthropic uses "system" field, OpenAI uses system role in messages
	if system, ok := data["system"].(string); ok && system != "" {
		messages, _ := data["messages"].([]interface{})
		// Prepend system message
		systemMsg := map[string]interface{}{
			"role":    "system",
			"content": system,
		}
		data["messages"] = append([]interface{}{systemMsg}, messages...)
		delete(data, "system")
	}

	// Transform Anthropic message content blocks to OpenAI format
	if messages, ok := data["messages"].([]interface{}); ok {
		transformedMessages := make([]interface{}, 0, len(messages))
		for _, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				transformed := t.transformAnthropicMessageToOpenAI(msgMap)

				// Check if this is a tool results marker that needs expansion
				if toolResults, ok := transformed["_anthropic_tool_results"].([]interface{}); ok {
					// Expand each tool_result into a separate OpenAI "tool" message
					for _, tr := range toolResults {
						if trMap, ok := tr.(map[string]interface{}); ok {
							toolMsg := map[string]interface{}{
								"role":         "tool",
								"tool_call_id": trMap["tool_use_id"],
							}

							// Extract content from tool_result
							if content, ok := trMap["content"].([]interface{}); ok {
								// Concatenate text blocks
								var textParts []string
								for _, c := range content {
									if cMap, ok := c.(map[string]interface{}); ok {
										if cMap["type"] == "text" {
											if text, ok := cMap["text"].(string); ok {
												textParts = append(textParts, text)
											}
										}
									}
								}
								if len(textParts) > 0 {
									toolMsg["content"] = textParts[0]
								}
							} else if contentStr, ok := trMap["content"].(string); ok {
								toolMsg["content"] = contentStr
							}

							transformedMessages = append(transformedMessages, toolMsg)
						}
					}
				} else {
					transformedMessages = append(transformedMessages, transformed)
				}
			}
		}
		data["messages"] = transformedMessages
	}

	return toJSON(data)
}

// transformAnthropicMessageToOpenAI converts Anthropic message format to OpenAI format.
// Handles tool_use and tool_result content blocks.
func (t *OpenAITransformer) transformAnthropicMessageToOpenAI(msg map[string]interface{}) map[string]interface{} {
	role, _ := msg["role"].(string)
	content := msg["content"]

	// If content is a string, return as-is
	if contentStr, ok := content.(string); ok {
		return map[string]interface{}{
			"role":    role,
			"content": contentStr,
		}
	}

	// If content is an array of blocks, transform based on role
	if contentBlocks, ok := content.([]interface{}); ok {
		// Assistant message with tool_use blocks -> OpenAI assistant with tool_calls
		if role == "assistant" {
			return t.transformAssistantMessage(contentBlocks)
		}

		// User message with tool_result blocks -> OpenAI tool messages
		if role == "user" {
			return t.transformUserMessageWithToolResults(contentBlocks)
		}

		// Regular user message with text blocks
		var textParts []string
		for _, block := range contentBlocks {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockMap["type"] == "text" {
					if text, ok := blockMap["text"].(string); ok {
						textParts = append(textParts, text)
					}
				}
			}
		}
		if len(textParts) > 0 {
			return map[string]interface{}{
				"role":    role,
				"content": textParts[0], // OpenAI expects single string
			}
		}
	}

	// Fallback
	return msg
}

// transformAssistantMessage converts Anthropic assistant message with tool_use to OpenAI format.
func (t *OpenAITransformer) transformAssistantMessage(contentBlocks []interface{}) map[string]interface{} {
	var textContent string
	var toolCalls []interface{}

	for _, block := range contentBlocks {
		if blockMap, ok := block.(map[string]interface{}); ok {
			blockType, _ := blockMap["type"].(string)

			switch blockType {
			case "text":
				if text, ok := blockMap["text"].(string); ok {
					textContent = text
				}

			case "tool_use":
				// Convert Anthropic tool_use to OpenAI tool_call
				toolCall := map[string]interface{}{
					"id":   blockMap["id"],
					"type": "function",
					"function": map[string]interface{}{
						"name": blockMap["name"],
					},
				}

				// Serialize input as JSON string
				if input := blockMap["input"]; input != nil {
					if inputBytes, err := json.Marshal(input); err == nil {
						toolCall["function"].(map[string]interface{})["arguments"] = string(inputBytes)
					}
				}

				toolCalls = append(toolCalls, toolCall)
			}
		}
	}

	result := map[string]interface{}{
		"role": "assistant",
	}

	// OpenAI requires content to be present (can be null if tool_calls exist)
	if textContent != "" {
		result["content"] = textContent
	} else if len(toolCalls) > 0 {
		result["content"] = nil
	}

	if len(toolCalls) > 0 {
		result["tool_calls"] = toolCalls
	}

	return result
}

// transformUserMessageWithToolResults converts Anthropic user message with tool_result blocks.
// In OpenAI format, tool results are separate "tool" role messages.
func (t *OpenAITransformer) transformUserMessageWithToolResults(contentBlocks []interface{}) map[string]interface{} {
	// Check if this message contains tool_result blocks
	hasToolResults := false
	for _, block := range contentBlocks {
		if blockMap, ok := block.(map[string]interface{}); ok {
			if blockMap["type"] == "tool_result" {
				hasToolResults = true
				break
			}
		}
	}

	// If no tool results, treat as regular user message
	if !hasToolResults {
		var textParts []string
		for _, block := range contentBlocks {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockMap["type"] == "text" {
					if text, ok := blockMap["text"].(string); ok {
						textParts = append(textParts, text)
					}
				}
			}
		}
		if len(textParts) > 0 {
			return map[string]interface{}{
				"role":    "user",
				"content": textParts[0],
			}
		}
	}

	// For tool results, we need to return a special marker that will be expanded later
	// OpenAI expects separate messages for each tool result
	// We'll use a special structure that the caller can detect and expand
	toolResults := make([]interface{}, 0)
	for _, block := range contentBlocks {
		if blockMap, ok := block.(map[string]interface{}); ok {
			if blockMap["type"] == "tool_result" {
				toolResults = append(toolResults, blockMap)
			}
		}
	}

	// Return a marker structure
	return map[string]interface{}{
		"_anthropic_tool_results": toolResults,
	}
}

// TransformResponse transforms a response from OpenAI format.
// If the client expects OpenAI format, no transformation is needed.
// If the client expects Anthropic format, convert from OpenAI format.
func (t *OpenAITransformer) TransformResponse(body []byte, clientFormat string) ([]byte, error) {
	// Normalize format
	normalized := NormalizeFormat(clientFormat)
	if normalized == "openai" || normalized == "" {
		// No transformation needed (empty defaults to openai for this transformer)
		return body, nil
	}

	// OpenAI → Anthropic response transformation
	data, err := parseJSON(body)
	if err != nil {
		return body, nil
	}

	// Transform OpenAI response to Anthropic format
	// OpenAI: { id, object, created, model, choices: [{index, message, finish_reason}], usage }
	// Anthropic: { id, type, role, content: [{type, text}], model, stop_reason, usage }

	anthropicResponse := map[string]interface{}{
		"id":    data["id"],
		"type":  "message",
		"role":  "assistant",
		"model": data["model"],
	}

	// Transform choices to content
	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				var contentBlocks []interface{}

				// Add text content if present
				if content, ok := message["content"].(string); ok && content != "" {
					contentBlocks = append(contentBlocks, map[string]interface{}{
						"type": "text",
						"text": content,
					})
				}

				// Transform tool_calls to tool_use blocks
				if toolCalls, ok := message["tool_calls"].([]interface{}); ok {
					for _, tc := range toolCalls {
						if toolCall, ok := tc.(map[string]interface{}); ok {
							if function, ok := toolCall["function"].(map[string]interface{}); ok {
								// Parse arguments JSON string
								var args interface{}
								if argsStr, ok := function["arguments"].(string); ok {
									json.Unmarshal([]byte(argsStr), &args)
								}

								contentBlocks = append(contentBlocks, map[string]interface{}{
									"type":  "tool_use",
									"id":    toolCall["id"],
									"name":  function["name"],
									"input": args,
								})
							}
						}
					}
				}

				if len(contentBlocks) > 0 {
					anthropicResponse["content"] = contentBlocks
				}
			}

			// Map finish_reason to stop_reason
			if finishReason, ok := choice["finish_reason"].(string); ok {
				switch finishReason {
				case "stop":
					anthropicResponse["stop_reason"] = "end_turn"
				case "length":
					anthropicResponse["stop_reason"] = "max_tokens"
				case "tool_calls":
					anthropicResponse["stop_reason"] = "tool_use"
				default:
					anthropicResponse["stop_reason"] = finishReason
				}
			}
		}
	}

	// Transform usage
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		anthropicResponse["usage"] = map[string]interface{}{
			"input_tokens":  usage["prompt_tokens"],
			"output_tokens": usage["completion_tokens"],
		}
	}

	return toJSON(anthropicResponse)
}

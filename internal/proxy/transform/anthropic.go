package transform

import (
	"strings"
)

// AnthropicTransformer handles Anthropic Messages API format.
// This is the default format used by Claude Code.
type AnthropicTransformer struct{}

func (t *AnthropicTransformer) Name() string {
	return "anthropic"
}

// TransformRequest transforms a request to Anthropic format.
// If the client is already using Anthropic format, no transformation is needed.
// If the client is using OpenAI format, convert to Anthropic format.
func (t *AnthropicTransformer) TransformRequest(body []byte, clientFormat string) ([]byte, error) {
	// Normalize format
	normalized := normalizeFormat(clientFormat)
	if normalized == "anthropic" {
		// No transformation needed
		return body, nil
	}

	// OpenAI → Anthropic transformation
	data, err := parseJSON(body)
	if err != nil {
		return body, nil // Return original on parse error
	}

	// Handle OpenAI Responses API format (uses "input" instead of "messages")
	if clientFormat == FormatOpenAIResponses || (clientFormat == "openai" && data["input"] != nil) {
		if input, ok := data["input"]; ok {
			// Convert input to messages format
			messages := convertInputToMessages(input)
			if len(messages) > 0 {
				// Check for _system marker in first message
				if first, ok := messages[0].(map[string]interface{}); ok {
					if sysContent, ok := first["_system"].(string); ok {
						// Extract system content and remove marker
						if existing, ok := data["system"].(string); ok && existing != "" {
							data["system"] = existing + "\n\n" + sysContent
						} else {
							data["system"] = sysContent
						}
						messages = messages[1:] // Remove the marker
					}
				}
				data["messages"] = messages
				delete(data, "input")
			}
		}
	}

	// Handle "instructions" field (system prompt in Responses API)
	if instructions, ok := data["instructions"].(string); ok && instructions != "" {
		data["system"] = instructions
		delete(data, "instructions")
	}

	// Transform max_completion_tokens → max_tokens (also used by Responses API as max_output_tokens)
	if maxCompletionTokens, ok := data["max_completion_tokens"]; ok {
		data["max_tokens"] = maxCompletionTokens
		delete(data, "max_completion_tokens")
	}
	if maxOutputTokens, ok := data["max_output_tokens"]; ok {
		data["max_tokens"] = maxOutputTokens
		delete(data, "max_output_tokens")
	}

	// Transform n parameter (OpenAI uses n for number of completions)
	delete(data, "n")

	// Transform temperature (both use same format, no change needed)

	// Transform stop sequences (OpenAI: stop, Anthropic: stop_sequences)
	if stop, ok := data["stop"]; ok {
		data["stop_sequences"] = stop
		delete(data, "stop")
	}

	// Transform stream_options (OpenAI specific)
	delete(data, "stream_options")

	// Transform logprobs (OpenAI specific)
	delete(data, "logprobs")
	delete(data, "top_logprobs")

	// Transform presence_penalty/frequency_penalty (OpenAI specific, not supported in Anthropic)
	delete(data, "presence_penalty")
	delete(data, "frequency_penalty")

	// Transform seed (OpenAI specific)
	delete(data, "seed")

	// Transform response_format (OpenAI specific)
	delete(data, "response_format")

	// Transform tools format if present
	// OpenAI tools format is different from Anthropic
	// This is a simplified transformation - full implementation would need more work
	if tools, ok := data["tools"].([]interface{}); ok {
		anthropicTools := make([]interface{}, 0, len(tools))
		for _, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				if toolMap["type"] == "function" {
					if fn, ok := toolMap["function"].(map[string]interface{}); ok {
						anthropicTool := map[string]interface{}{
							"name":        fn["name"],
							"description": fn["description"],
						}
						if params, ok := fn["parameters"]; ok {
							anthropicTool["input_schema"] = params
						}
						anthropicTools = append(anthropicTools, anthropicTool)
					}
				}
			}
		}
		if len(anthropicTools) > 0 {
			data["tools"] = anthropicTools
		}
	}

	result, err := toJSON(data)
	if err != nil {
		return body, err
	}

	return result, nil
}

// TransformResponse transforms a response from Anthropic format.
// If the client expects Anthropic format, no transformation is needed.
// If the client expects OpenAI format, convert from Anthropic format.
func (t *AnthropicTransformer) TransformResponse(body []byte, clientFormat string) ([]byte, error) {
	// Normalize format
	normalized := normalizeFormat(clientFormat)
	if normalized == "anthropic" {
		// No transformation needed
		return body, nil
	}

	// Anthropic → OpenAI response transformation
	data, err := parseJSON(body)
	if err != nil {
		return body, nil
	}

	// Determine target format: Chat Completions or Responses API
	if clientFormat == FormatOpenAIResponses {
		return t.transformToResponsesAPI(data)
	}

	// Default: Transform to Chat Completions format
	return t.transformToChatCompletions(data)
}

// transformToChatCompletions transforms Anthropic response to OpenAI Chat Completions format.
func (t *AnthropicTransformer) transformToChatCompletions(data map[string]interface{}) ([]byte, error) {
	// Transform Anthropic response to OpenAI format
	// Anthropic: { id, type, role, content: [{type, text}], model, stop_reason, usage }
	// OpenAI: { id, object, created, model, choices: [{index, message, finish_reason}], usage }

	openAIResponse := map[string]interface{}{
		"id":      data["id"],
		"object":  "chat.completion",
		"created": 0, // Anthropic doesn't provide this
		"model":   data["model"],
	}

	// Transform content to choices
	var messageContent string
	if content, ok := data["content"].([]interface{}); ok {
		for _, c := range content {
			if cMap, ok := c.(map[string]interface{}); ok {
				if cMap["type"] == "text" {
					if text, ok := cMap["text"].(string); ok {
						messageContent = text
						break
					}
				}
			}
		}
	}

	// Map stop_reason to finish_reason
	finishReason := "stop"
	if stopReason, ok := data["stop_reason"].(string); ok {
		switch stopReason {
		case "end_turn":
			finishReason = "stop"
		case "max_tokens":
			finishReason = "length"
		case "tool_use":
			finishReason = "tool_calls"
		default:
			finishReason = stopReason
		}
	}

	openAIResponse["choices"] = []interface{}{
		map[string]interface{}{
			"index": 0,
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": messageContent,
			},
			"finish_reason": finishReason,
		},
	}

	// Transform usage
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		openAIResponse["usage"] = map[string]interface{}{
			"prompt_tokens":     usage["input_tokens"],
			"completion_tokens": usage["output_tokens"],
			"total_tokens": func() interface{} {
				input, _ := usage["input_tokens"].(float64)
				output, _ := usage["output_tokens"].(float64)
				return input + output
			}(),
		}
	}

	return toJSON(openAIResponse)
}

// transformToResponsesAPI transforms Anthropic response to OpenAI Responses API format.
func (t *AnthropicTransformer) transformToResponsesAPI(data map[string]interface{}) ([]byte, error) {
	// Anthropic: { id, type, role, content: [{type, text}], model, stop_reason, usage }
	// Responses API: { id, object, status, output: [{type, content}], usage }

	responsesAPIResponse := map[string]interface{}{
		"id":     data["id"],
		"object": "response",
		"status": "completed",
		"model":  data["model"],
	}

	// Transform content to output
	var output []interface{}
	if content, ok := data["content"].([]interface{}); ok {
		for _, c := range content {
			if cMap, ok := c.(map[string]interface{}); ok {
				if cMap["type"] == "text" {
					output = append(output, map[string]interface{}{
						"type":    "message",
						"role":    "assistant",
						"content": []interface{}{cMap},
					})
				}
			}
		}
	}
	responsesAPIResponse["output"] = output

	// Transform usage
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		responsesAPIResponse["usage"] = map[string]interface{}{
			"input_tokens":  usage["input_tokens"],
			"output_tokens": usage["output_tokens"],
		}
	}

	return toJSON(responsesAPIResponse)
}

// convertInputToMessages converts OpenAI Responses API "input" field to Anthropic messages format.
// The input can be a string or an array of message objects.
func convertInputToMessages(input interface{}) []interface{} {
	var messages []interface{}
	var systemParts []string // Collect developer/system messages

	switch v := input.(type) {
	case string:
		// Simple string input → single user message
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": v,
		})
	case []interface{}:
		// Array of messages or content items
		for _, item := range v {
			switch msg := item.(type) {
			case string:
				// String item → user message
				messages = append(messages, map[string]interface{}{
					"role":    "user",
					"content": msg,
				})
			case map[string]interface{}:
				// Check if it's a message with type: "message"
				if msgType, ok := msg["type"].(string); ok && msgType == "message" {
					role, _ := msg["role"].(string)
					content := convertContent(msg["content"])

					// Handle developer role → collect as system prompt
					if role == "developer" || role == "system" {
						if text := extractTextFromContent(content); text != "" {
							systemParts = append(systemParts, text)
						}
						continue
					}

					// Map roles
					anthropicRole := mapRole(role)
					messages = append(messages, map[string]interface{}{
						"role":    anthropicRole,
						"content": content,
					})
				} else if role, ok := msg["role"].(string); ok {
					// Direct message with role (not wrapped in type: "message")
					content := convertContent(msg["content"])

					if role == "developer" || role == "system" {
						if text := extractTextFromContent(content); text != "" {
							systemParts = append(systemParts, text)
						}
						continue
					}

					anthropicRole := mapRole(role)
					messages = append(messages, map[string]interface{}{
						"role":    anthropicRole,
						"content": content,
					})
				} else if msgType, ok := msg["type"].(string); ok {
					// Content item (e.g., {type: "input_text", text: "..."})
					if msgType == "input_text" || msgType == "text" {
						if text, ok := msg["text"].(string); ok {
							messages = append(messages, map[string]interface{}{
								"role":    "user",
								"content": text,
							})
						}
					}
				}
			}
		}
	}

	// If we collected system parts, prepend as first message or handle separately
	// For now, we'll return them in a special way that TransformRequest can handle
	if len(systemParts) > 0 && len(messages) > 0 {
		// Insert system content marker that TransformRequest will extract
		messages = append([]interface{}{
			map[string]interface{}{
				"_system": strings.Join(systemParts, "\n\n"),
			},
		}, messages...)
	}

	return messages
}

// convertContent transforms OpenAI content format to Anthropic format.
// OpenAI uses "input_text" type, Anthropic uses "text" type.
func convertContent(content interface{}) interface{} {
	switch c := content.(type) {
	case string:
		return c
	case []interface{}:
		var result []interface{}
		for _, item := range c {
			if itemMap, ok := item.(map[string]interface{}); ok {
				itemType, _ := itemMap["type"].(string)
				// Convert input_text → text
				if itemType == "input_text" {
					result = append(result, map[string]interface{}{
						"type": "text",
						"text": itemMap["text"],
					})
				} else if itemType == "text" {
					result = append(result, itemMap)
				} else {
					// Pass through other types (images, etc.)
					result = append(result, itemMap)
				}
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return content
}

// extractTextFromContent extracts plain text from content (string or array).
func extractTextFromContent(content interface{}) string {
	switch c := content.(type) {
	case string:
		return c
	case []interface{}:
		var parts []string
		for _, item := range c {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, ok := itemMap["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// mapRole converts OpenAI roles to Anthropic roles.
func mapRole(role string) string {
	switch role {
	case "developer", "system":
		return "user" // Will be handled separately as system prompt
	case "assistant":
		return "assistant"
	default:
		return "user"
	}
}

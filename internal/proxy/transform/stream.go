package transform

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// StreamTransformer transforms SSE streams between API formats.
type StreamTransformer struct {
	ClientFormat   string
	ProviderFormat string
	MessageID      string
	Model          string
}

// writeStreamError emits a protocol-native error event based on client format
func (st *StreamTransformer) writeStreamError(w io.Writer, err error) {
	normalizedClient := NormalizeFormat(st.ClientFormat)

	if normalizedClient == "openai" {
		// OpenAI formats use error event
		if st.ClientFormat == FormatOpenAIChat {
			// Chat Completions format
			fmt.Fprintf(w, "data: {\"error\":{\"message\":\"%s\",\"type\":\"stream_error\"}}\n\n", err.Error())
		} else {
			// Responses API format
			fmt.Fprintf(w, "event: error\ndata: {\"type\":\"error\",\"error\":{\"message\":\"%s\",\"type\":\"stream_error\"}}\n\n", err.Error())
		}
	} else {
		// Anthropic format uses error event
		fmt.Fprintf(w, "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"stream_error\",\"message\":\"%s\"}}\n\n", err.Error())
	}
}

// TransformSSEStream transforms SSE streams between API formats.
// Returns a reader that produces the appropriate SSE events.
func (st *StreamTransformer) TransformSSEStream(r io.Reader) io.Reader {
	// Normalize formats for comparison
	normalizedClient := NormalizeFormat(st.ClientFormat)
	normalizedProvider := NormalizeFormat(st.ProviderFormat)

	if normalizedClient == normalizedProvider {
		return r
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		// Check specific format first before normalized comparison
		if st.ProviderFormat == FormatOpenAIResponses && normalizedClient == "anthropic" {
			st.transformResponsesAPIToAnthropic(r, pw)
		} else if normalizedProvider == "anthropic" && normalizedClient == "openai" {
			// Provider is Anthropic, client expects OpenAI
			// Distinguish between openai-chat and openai-responses
			// Default to Responses API for backward compatibility with legacy "openai"
			if st.ClientFormat == FormatOpenAIChat {
				st.transformAnthropicToOpenAIChat(r, pw)
			} else {
				// FormatOpenAIResponses or legacy "openai" → Responses API
				st.transformAnthropicToOpenAIResponses(r, pw)
			}
		} else if normalizedProvider == "openai" && normalizedClient == "anthropic" {
			st.transformOpenAIToAnthropic(r, pw)
		}
	}()

	return pr
}

// transformAnthropicToOpenAIResponses converts Anthropic SSE events to OpenAI Responses API format.
func (st *StreamTransformer) transformAnthropicToOpenAIResponses(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for large events
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentEvent string
	var dataBuffer bytes.Buffer
	created := time.Now().Unix()

	// Track state for building response
	var outputIndex int = 0
	var contentIndex int = 0
	var itemID string = "item_0"
	var fullText strings.Builder
	var inputTokens, outputTokens int

	// Send response.created first
	responseCreated := false

	for scanner.Scan() {
		line := scanner.Text()

		// Parse SSE format
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			dataBuffer.WriteString(strings.TrimPrefix(line, "data: "))
			continue
		}

		// Empty line = end of event
		if line == "" && dataBuffer.Len() > 0 {
			data := dataBuffer.String()
			dataBuffer.Reset()

			// Transform based on event type
			events := st.transformEventToResponses(currentEvent, data, created, &responseCreated,
				&outputIndex, &contentIndex, itemID, &fullText, &inputTokens, &outputTokens)
			for _, event := range events {
				fmt.Fprint(w, event)
			}
			currentEvent = ""
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		st.writeStreamError(w, err)
		return
	}

	// Send response.completed
	st.writeResponseCompleted(w, created, fullText.String(), inputTokens, outputTokens)
}

// transformAnthropicToOpenAIChat converts Anthropic SSE events to OpenAI Chat Completions format.
func (st *StreamTransformer) transformAnthropicToOpenAIChat(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentEvent string
	var dataBuffer bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			dataBuffer.WriteString(strings.TrimPrefix(line, "data: "))
			continue
		}

		// Empty line = end of event
		if line == "" && dataBuffer.Len() > 0 {
			data := dataBuffer.String()
			dataBuffer.Reset()

			// Transform Anthropic event to OpenAI Chat Completions chunk
			chunk := st.transformAnthropicEventToChat(currentEvent, data)
			if chunk != "" {
				fmt.Fprintf(w, "data: %s\n\n", chunk)
			}
			currentEvent = ""
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		st.writeStreamError(w, err)
		return
	}

	// Send [DONE]
	fmt.Fprintf(w, "data: [DONE]\n\n")
}

// transformAnthropicEventToChat transforms a single Anthropic event to OpenAI Chat Completions format.
func (st *StreamTransformer) transformAnthropicEventToChat(eventType, data string) string {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &eventData); err != nil {
		return ""
	}

	chunk := map[string]interface{}{
		"id":      st.MessageID,
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   st.Model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"delta": map[string]interface{}{},
			},
		},
	}

	delta := chunk["choices"].([]map[string]interface{})[0]["delta"].(map[string]interface{})

	switch eventType {
	case "message_start":
		delta["role"] = "assistant"
		delta["content"] = ""

	case "content_block_start":
		// Handle tool_use block start
		if contentBlock, ok := eventData["content_block"].(map[string]interface{}); ok {
			if contentBlock["type"] == "tool_use" {
				toolCall := map[string]interface{}{
					"index": eventData["index"],
					"id":    contentBlock["id"],
					"type":  "function",
					"function": map[string]interface{}{
						"name":      contentBlock["name"],
						"arguments": "",
					},
				}
				delta["tool_calls"] = []interface{}{toolCall}
			}
		}

	case "content_block_delta":
		if deltaData, ok := eventData["delta"].(map[string]interface{}); ok {
			if text, ok := deltaData["text"].(string); ok {
				delta["content"] = text
			} else if deltaData["type"] == "input_json_delta" {
				// Tool use arguments delta
				if partialJSON, ok := deltaData["partial_json"].(string); ok {
					index := 0
					if idx, ok := eventData["index"].(float64); ok {
						index = int(idx)
					}
					toolCall := map[string]interface{}{
						"index": index,
						"function": map[string]interface{}{
							"arguments": partialJSON,
						},
					}
					delta["tool_calls"] = []interface{}{toolCall}
				}
			}
		}

	case "message_delta":
		if stopReason, ok := eventData["delta"].(map[string]interface{})["stop_reason"].(string); ok {
			finishReason := "stop"
			if stopReason == "max_tokens" {
				finishReason = "length"
			} else if stopReason == "tool_use" {
				finishReason = "tool_calls"
			}
			chunk["choices"].([]map[string]interface{})[0]["finish_reason"] = finishReason
		}

	case "message_stop":
		return "" // Skip, handled by message_delta

	default:
		return ""
	}

	result, _ := json.Marshal(chunk)
	return string(result)
}

// transformEventToResponses transforms a single Anthropic event to OpenAI Responses API format.
func (st *StreamTransformer) transformEventToResponses(eventType, data string, created int64,
	responseCreated *bool, outputIndex, contentIndex *int, itemID string,
	fullText *strings.Builder, inputTokens, outputTokens *int) []string {

	var events []string
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &eventData); err != nil {
		return events
	}

	switch eventType {
	case "message_start":
		// Extract message info
		if msg, ok := eventData["message"].(map[string]interface{}); ok {
			if id, ok := msg["id"].(string); ok {
				st.MessageID = id
			}
			if model, ok := msg["model"].(string); ok {
				st.Model = model
			}
			// Extract usage if present
			if usage, ok := msg["usage"].(map[string]interface{}); ok {
				if it, ok := usage["input_tokens"].(float64); ok {
					*inputTokens = int(it)
				}
			}
		}

		// Send response.created on message_start
		if !*responseCreated {
			events = append(events, st.createResponseCreated(created))
			// Send response.in_progress
			events = append(events, st.createResponseInProgress(created))
			// Send response.output_item.added
			events = append(events, st.createOutputItemAdded(created, *outputIndex, itemID))
			// Send response.content_part.added
			events = append(events, st.createContentPartAdded(created, *outputIndex, itemID, *contentIndex))
			*responseCreated = true
		}

	case "content_block_start":
		// Handle content block start (text or tool_use)
		if contentBlock, ok := eventData["content_block"].(map[string]interface{}); ok {
			if contentBlock["type"] == "tool_use" {
				// Tool use block - emit function_call_output.added
				toolID := contentBlock["id"].(string)
				toolName := contentBlock["name"].(string)
				event := map[string]interface{}{
					"type":         "response.function_call_arguments.added",
					"item_id":      toolID,
					"output_index": *outputIndex,
					"call_id":      toolID,
					"name":         toolName,
					"arguments":    "",
				}
				events = append(events, formatSSEEvent("response.function_call_arguments.added", event))
			}
		}

	case "content_block_delta":
		// Extract text delta or tool input delta
		if delta, ok := eventData["delta"].(map[string]interface{}); ok {
			if deltaType, ok := delta["type"].(string); ok {
				if deltaType == "text_delta" {
					if text, ok := delta["text"].(string); ok {
						fullText.WriteString(text)
						events = append(events, st.createOutputTextDelta(created, *outputIndex, itemID, *contentIndex, text))
					}
				} else if deltaType == "input_json_delta" {
					// Tool arguments delta
					if partialJSON, ok := delta["partial_json"].(string); ok {
						event := map[string]interface{}{
							"type":       "response.function_call_arguments.delta",
							"item_id":    itemID,
							"output_index": *outputIndex,
							"delta":      partialJSON,
						}
						events = append(events, formatSSEEvent("response.function_call_arguments.delta", event))
					}
				}
			}
		}

	case "message_delta":
		// Extract usage and stop reason
		if usage, ok := eventData["usage"].(map[string]interface{}); ok {
			if ot, ok := usage["output_tokens"].(float64); ok {
				*outputTokens = int(ot)
			}
		}
		// Send content_part.done and output_item.done
		events = append(events, st.createContentPartDone(created, *outputIndex, itemID, *contentIndex, fullText.String()))
		events = append(events, st.createOutputItemDone(created, *outputIndex, itemID, fullText.String()))

	case "message_stop":
		// Final message - response.completed will be sent after loop
	}

	return events
}

// Helper functions to create OpenAI Responses API SSE events

func (st *StreamTransformer) createResponseCreated(created int64) string {
	event := map[string]interface{}{
		"type":       "response.created",
		"response": map[string]interface{}{
			"id":         st.MessageID,
			"object":     "response",
			"created_at": created,
			"status":     "in_progress",
			"model":      st.Model,
			"output":     []interface{}{},
		},
	}
	return formatSSEEvent("response.created", event)
}

func (st *StreamTransformer) createResponseInProgress(created int64) string {
	event := map[string]interface{}{
		"type":       "response.in_progress",
		"response": map[string]interface{}{
			"id":         st.MessageID,
			"object":     "response",
			"created_at": created,
			"status":     "in_progress",
			"model":      st.Model,
			"output":     []interface{}{},
		},
	}
	return formatSSEEvent("response.in_progress", event)
}

func (st *StreamTransformer) createOutputItemAdded(created int64, outputIndex int, itemID string) string {
	event := map[string]interface{}{
		"type":         "response.output_item.added",
		"output_index": outputIndex,
		"item": map[string]interface{}{
			"id":      itemID,
			"type":    "message",
			"role":    "assistant",
			"content": []interface{}{},
		},
	}
	return formatSSEEvent("response.output_item.added", event)
}

func (st *StreamTransformer) createContentPartAdded(created int64, outputIndex int, itemID string, contentIndex int) string {
	event := map[string]interface{}{
		"type":          "response.content_part.added",
		"item_id":       itemID,
		"output_index":  outputIndex,
		"content_index": contentIndex,
		"part": map[string]interface{}{
			"type": "output_text",
			"text": "",
		},
	}
	return formatSSEEvent("response.content_part.added", event)
}

func (st *StreamTransformer) createOutputTextDelta(created int64, outputIndex int, itemID string, contentIndex int, text string) string {
	event := map[string]interface{}{
		"type":          "response.output_text.delta",
		"item_id":       itemID,
		"output_index":  outputIndex,
		"content_index": contentIndex,
		"delta":         text,
	}
	return formatSSEEvent("response.output_text.delta", event)
}

func (st *StreamTransformer) createContentPartDone(created int64, outputIndex int, itemID string, contentIndex int, text string) string {
	event := map[string]interface{}{
		"type":          "response.content_part.done",
		"item_id":       itemID,
		"output_index":  outputIndex,
		"content_index": contentIndex,
		"part": map[string]interface{}{
			"type": "output_text",
			"text": text,
		},
	}
	return formatSSEEvent("response.content_part.done", event)
}

func (st *StreamTransformer) createOutputItemDone(created int64, outputIndex int, itemID string, text string) string {
	event := map[string]interface{}{
		"type":         "response.output_item.done",
		"output_index": outputIndex,
		"item": map[string]interface{}{
			"id":   itemID,
			"type": "message",
			"role": "assistant",
			"content": []interface{}{
				map[string]interface{}{
					"type": "output_text",
					"text": text,
				},
			},
		},
	}
	return formatSSEEvent("response.output_item.done", event)
}

func (st *StreamTransformer) writeResponseCompleted(w io.Writer, created int64, text string, inputTokens, outputTokens int) {
	event := map[string]interface{}{
		"type": "response.completed",
		"response": map[string]interface{}{
			"id":         st.MessageID,
			"object":     "response",
			"created_at": created,
			"status":     "completed",
			"model":      st.Model,
			"output": []interface{}{
				map[string]interface{}{
					"id":   "item_0",
					"type": "message",
					"role": "assistant",
					"content": []interface{}{
						map[string]interface{}{
							"type": "output_text",
							"text": text,
						},
					},
				},
			},
			"usage": map[string]interface{}{
				"input_tokens":  inputTokens,
				"output_tokens": outputTokens,
				"total_tokens":  inputTokens + outputTokens,
			},
		},
	}
	fmt.Fprint(w, formatSSEEvent("response.completed", event))
}

// formatSSEEvent formats a map as an SSE event with event type and JSON data.
func formatSSEEvent(eventType string, data map[string]interface{}) string {
	jsonData, _ := json.Marshal(data)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))
}

// transformOpenAIToAnthropic converts OpenAI Chat Completions SSE events to Anthropic Messages API format.
func (st *StreamTransformer) transformOpenAIToAnthropic(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var messageStarted bool
	var inputTokens, outputTokens int
	var finalStopReason string
	var messageStopped bool

	// Track content blocks: map OpenAI tool_call index to Anthropic content block index
	type blockState struct {
		started          bool
		anthropicIndex   int    // Anthropic content array index
		typ              string // "text" or "tool_use"
	}
	// Map OpenAI tool_call index to block state
	toolBlocksByOpenAIIndex := make(map[int]*blockState)
	var textBlock *blockState
	nextAnthropicIndex := 0 // Global counter for Anthropic content block indices

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse SSE data line
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Handle [DONE] signal
		if data == "[DONE]" {
			// Only send termination if we haven't already sent it via finish_reason
			if !messageStopped {
				// Send content_block_stop for all open blocks
				if textBlock != nil && textBlock.started {
					fmt.Fprint(w, formatSSEEvent("content_block_stop", map[string]interface{}{
						"type":  "content_block_stop",
						"index": textBlock.anthropicIndex,
					}))
				}
				for _, block := range toolBlocksByOpenAIIndex {
					if block.started {
						fmt.Fprint(w, formatSSEEvent("content_block_stop", map[string]interface{}{
							"type":  "content_block_stop",
							"index": block.anthropicIndex,
						}))
					}
				}

				// Use finalStopReason if set, otherwise default to end_turn
				stopReason := finalStopReason
				if stopReason == "" {
					stopReason = "end_turn"
				}

				// Send message_delta with stop_reason
				fmt.Fprint(w, formatSSEEvent("message_delta", map[string]interface{}{
					"type": "message_delta",
					"delta": map[string]interface{}{
						"stop_reason":   stopReason,
						"stop_sequence": nil,
					},
					"usage": map[string]interface{}{
						"output_tokens": outputTokens,
					},
				}))

				// Send message_stop
				fmt.Fprint(w, formatSSEEvent("message_stop", map[string]interface{}{
					"type": "message_stop",
				}))
				messageStopped = true
			}
			continue
		}

		// Parse JSON data
		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// Extract message ID and model on first chunk
		if !messageStarted {
			if id, ok := chunk["id"].(string); ok {
				st.MessageID = id
			}
			if model, ok := chunk["model"].(string); ok {
				st.Model = model
			}

			// Send message_start
			fmt.Fprint(w, formatSSEEvent("message_start", map[string]interface{}{
				"type": "message_start",
				"message": map[string]interface{}{
					"id":            st.MessageID,
					"type":          "message",
					"role":          "assistant",
					"content":       []interface{}{},
					"model":         st.Model,
					"stop_reason":   nil,
					"stop_sequence": nil,
					"usage": map[string]interface{}{
						"input_tokens":  inputTokens,
						"output_tokens": 0,
					},
				},
			}))
			messageStarted = true
		}

		// Extract usage if present
		if usage, ok := chunk["usage"].(map[string]interface{}); ok {
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				inputTokens = int(pt)
			}
			if ct, ok := usage["completion_tokens"].(float64); ok {
				outputTokens = int(ct)
			}
		}

		// Process choices
		choices, ok := chunk["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}

		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			continue
		}

		delta, ok := choice["delta"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check for tool_calls delta
		if toolCalls, ok := delta["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
			for _, tc := range toolCalls {
				toolCall, ok := tc.(map[string]interface{})
				if !ok {
					continue
				}

				// OpenAI tool_call index (for parallel tool calls)
				openaiToolIndex := 0
				if idx, ok := toolCall["index"].(float64); ok {
					openaiToolIndex = int(idx)
				}

				// Check if this is a new tool call (has id)
				if id, ok := toolCall["id"].(string); ok && id != "" {
					// Close previous block at this OpenAI tool index if open
					if existingBlock, exists := toolBlocksByOpenAIIndex[openaiToolIndex]; exists && existingBlock.started {
						fmt.Fprint(w, formatSSEEvent("content_block_stop", map[string]interface{}{
							"type":  "content_block_stop",
							"index": existingBlock.anthropicIndex,
						}))
					}

					// Get function name
					var functionName string
					if function, ok := toolCall["function"].(map[string]interface{}); ok {
						if name, ok := function["name"].(string); ok {
							functionName = name
						}
					}

					// Allocate new Anthropic content block index
					anthropicIndex := nextAnthropicIndex
					nextAnthropicIndex++

					// Send content_block_start for tool_use
					fmt.Fprint(w, formatSSEEvent("content_block_start", map[string]interface{}{
						"type":  "content_block_start",
						"index": anthropicIndex,
						"content_block": map[string]interface{}{
							"type":  "tool_use",
							"id":    id,
							"name":  functionName,
							"input": map[string]interface{}{},
						},
					}))
					toolBlocksByOpenAIIndex[openaiToolIndex] = &blockState{
						started:        true,
						anthropicIndex: anthropicIndex,
						typ:            "tool_use",
					}
				}

				// Check for function arguments delta
				if function, ok := toolCall["function"].(map[string]interface{}); ok {
					if args, ok := function["arguments"].(string); ok && args != "" {
						// Get the block for this OpenAI tool index
						if block, exists := toolBlocksByOpenAIIndex[openaiToolIndex]; exists {
							// Send input_json_delta
							fmt.Fprint(w, formatSSEEvent("content_block_delta", map[string]interface{}{
								"type":  "content_block_delta",
								"index": block.anthropicIndex,
								"delta": map[string]interface{}{
									"type":         "input_json_delta",
									"partial_json": args,
								},
							}))
						}
					}
				}
			}
			continue
		}

		// Check for content delta
		if content, ok := delta["content"].(string); ok && content != "" {
			// Start text block if not started
			if textBlock == nil || !textBlock.started {
				// Close previous text block if it exists
				if textBlock != nil && textBlock.started {
					fmt.Fprint(w, formatSSEEvent("content_block_stop", map[string]interface{}{
						"type":  "content_block_stop",
						"index": textBlock.anthropicIndex,
					}))
				}

				// Allocate new Anthropic content block index for text
				anthropicIndex := nextAnthropicIndex
				nextAnthropicIndex++

				// Start new text block
				fmt.Fprint(w, formatSSEEvent("content_block_start", map[string]interface{}{
					"type":  "content_block_start",
					"index": anthropicIndex,
					"content_block": map[string]interface{}{
						"type": "text",
						"text": "",
					},
				}))
				textBlock = &blockState{
					started:        true,
					anthropicIndex: anthropicIndex,
					typ:            "text",
				}
			}

			// Send content_block_delta
			fmt.Fprint(w, formatSSEEvent("content_block_delta", map[string]interface{}{
				"type":  "content_block_delta",
				"index": textBlock.anthropicIndex,
				"delta": map[string]interface{}{
					"type": "text_delta",
					"text": content,
				},
			}))
		}

		// Check for finish_reason
		if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
			// Close all open content blocks
			if textBlock != nil && textBlock.started {
				fmt.Fprint(w, formatSSEEvent("content_block_stop", map[string]interface{}{
					"type":  "content_block_stop",
					"index": textBlock.anthropicIndex,
				}))
				textBlock.started = false
			}
			for _, block := range toolBlocksByOpenAIIndex {
				if block.started {
					fmt.Fprint(w, formatSSEEvent("content_block_stop", map[string]interface{}{
						"type":  "content_block_stop",
						"index": block.anthropicIndex,
					}))
					block.started = false
				}
			}

			// Map finish_reason to stop_reason
			stopReason := "end_turn"
			switch finishReason {
			case "length":
				stopReason = "max_tokens"
			case "tool_calls":
				stopReason = "tool_use"
			case "content_filter":
				stopReason = "end_turn"
			}

			// Store the stop reason for potential [DONE] handling
			finalStopReason = stopReason

			// Send message_delta with stop_reason
			fmt.Fprint(w, formatSSEEvent("message_delta", map[string]interface{}{
				"type": "message_delta",
				"delta": map[string]interface{}{
					"stop_reason":   stopReason,
					"stop_sequence": nil,
				},
				"usage": map[string]interface{}{
					"output_tokens": outputTokens,
				},
			}))

			// Send message_stop
			fmt.Fprint(w, formatSSEEvent("message_stop", map[string]interface{}{
				"type": "message_stop",
			}))
			messageStopped = true
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		st.writeStreamError(w, err)
		return
	}
}

// transformResponsesAPIToAnthropic converts OpenAI Responses API SSE events
// to Anthropic Messages API SSE format.
// Responses API uses event: + data: lines with typed event names.
// Anthropic uses event: + data: lines with different event names.
func (st *StreamTransformer) transformResponsesAPIToAnthropic(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentEvent string
	var dataBuffer bytes.Buffer
	var messageStarted bool
	var contentBlockIndex int
	var hasToolUse bool
	var inputTokens, outputTokens int

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			dataBuffer.WriteString(strings.TrimPrefix(line, "data: "))
			continue
		}

		// Empty line = end of event
		if line == "" && dataBuffer.Len() > 0 {
			data := dataBuffer.String()
			dataBuffer.Reset()

			var eventData map[string]interface{}
			if err := json.Unmarshal([]byte(data), &eventData); err != nil {
				currentEvent = ""
				continue
			}

			switch currentEvent {
			case "response.created":
				// Extract response metadata
				if resp, ok := eventData["response"].(map[string]interface{}); ok {
					if id, ok := resp["id"].(string); ok {
						st.MessageID = id
					}
					if model, ok := resp["model"].(string); ok {
						st.Model = model
					}
				}

				if !messageStarted {
					fmt.Fprint(w, formatSSEEvent("message_start", map[string]interface{}{
						"type": "message_start",
						"message": map[string]interface{}{
							"id":            st.MessageID,
							"type":          "message",
							"role":          "assistant",
							"content":       []interface{}{},
							"model":         st.Model,
							"stop_reason":   nil,
							"stop_sequence": nil,
							"usage": map[string]interface{}{
								"input_tokens":  0,
								"output_tokens": 0,
							},
						},
					}))
					messageStarted = true
				}

			case "response.output_item.added":
				item, ok := eventData["item"].(map[string]interface{})
				if !ok {
					break
				}
				itemType, _ := item["type"].(string)

				if itemType == "message" {
					// Text message — content_block_start emitted on content_part.added
				} else if itemType == "function_call" {
					hasToolUse = true
					callID, _ := item["call_id"].(string)
					name, _ := item["name"].(string)
					fmt.Fprint(w, formatSSEEvent("content_block_start", map[string]interface{}{
						"type":  "content_block_start",
						"index": contentBlockIndex,
						"content_block": map[string]interface{}{
							"type":  "tool_use",
							"id":    callID,
							"name":  name,
							"input": map[string]interface{}{},
						},
					}))
				}

			case "response.content_part.added":
				// Text content block start
				fmt.Fprint(w, formatSSEEvent("content_block_start", map[string]interface{}{
					"type":  "content_block_start",
					"index": contentBlockIndex,
					"content_block": map[string]interface{}{
						"type": "text",
						"text": "",
					},
				}))

			case "response.output_text.delta":
				delta, _ := eventData["delta"].(string)
				fmt.Fprint(w, formatSSEEvent("content_block_delta", map[string]interface{}{
					"type":  "content_block_delta",
					"index": contentBlockIndex,
					"delta": map[string]interface{}{
						"type": "text_delta",
						"text": delta,
					},
				}))

			case "response.function_call_arguments.delta":
				delta, _ := eventData["delta"].(string)
				fmt.Fprint(w, formatSSEEvent("content_block_delta", map[string]interface{}{
					"type":  "content_block_delta",
					"index": contentBlockIndex,
					"delta": map[string]interface{}{
						"type":         "input_json_delta",
						"partial_json": delta,
					},
				}))

			case "response.output_item.done":
				fmt.Fprint(w, formatSSEEvent("content_block_stop", map[string]interface{}{
					"type":  "content_block_stop",
					"index": contentBlockIndex,
				}))
				contentBlockIndex++

			case "response.completed":
				// Extract usage from completed response
				if resp, ok := eventData["response"].(map[string]interface{}); ok {
					if usage, ok := resp["usage"].(map[string]interface{}); ok {
						if v, ok := usage["input_tokens"].(float64); ok {
							inputTokens = int(v)
						}
						if v, ok := usage["output_tokens"].(float64); ok {
							outputTokens = int(v)
						}
					}
				}

				stopReason := "end_turn"
				if hasToolUse {
					stopReason = "tool_use"
				}

				fmt.Fprint(w, formatSSEEvent("message_delta", map[string]interface{}{
					"type": "message_delta",
					"delta": map[string]interface{}{
						"stop_reason":   stopReason,
						"stop_sequence": nil,
					},
					"usage": map[string]interface{}{
						"output_tokens": outputTokens,
					},
				}))

				fmt.Fprint(w, formatSSEEvent("message_stop", map[string]interface{}{
					"type": "message_stop",
				}))
			}

			currentEvent = ""
		}
	}

	// Process remaining buffered data (stream may end without trailing blank line)
	if dataBuffer.Len() > 0 && currentEvent != "" {
		data := dataBuffer.String()
		var eventData map[string]interface{}
		if err := json.Unmarshal([]byte(data), &eventData); err == nil {
			if currentEvent == "response.completed" {
				if resp, ok := eventData["response"].(map[string]interface{}); ok {
					if usage, ok := resp["usage"].(map[string]interface{}); ok {
						if v, ok := usage["input_tokens"].(float64); ok {
							inputTokens = int(v)
						}
						if v, ok := usage["output_tokens"].(float64); ok {
							outputTokens = int(v)
						}
					}
				}

				stopReason := "end_turn"
				if hasToolUse {
					stopReason = "tool_use"
				}

				fmt.Fprint(w, formatSSEEvent("message_delta", map[string]interface{}{
					"type": "message_delta",
					"delta": map[string]interface{}{
						"stop_reason":   stopReason,
						"stop_sequence": nil,
					},
					"usage": map[string]interface{}{
						"output_tokens": outputTokens,
					},
				}))

				fmt.Fprint(w, formatSSEEvent("message_stop", map[string]interface{}{
					"type": "message_stop",
				}))
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		st.writeStreamError(w, err)
		return
	}

	_ = inputTokens
}

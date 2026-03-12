package proxy

import (
	"fmt"
	"net/http"
	"strings"
)

// NormalizedRequest represents a protocol-agnostic request structure.
// All API protocols (Anthropic Messages, OpenAI Chat, OpenAI Responses) are normalized to this format.
type NormalizedRequest struct {
	// Model is the requested model identifier
	Model string

	// SystemPrompt is the system message (if any)
	SystemPrompt string

	// Messages contains the conversation messages
	Messages []NormalizedMessage

	// HasTools indicates if the request includes tool/function definitions
	HasTools bool

	// HasWebSearch indicates if the request includes web_search tool
	HasWebSearch bool

	// HasThinking indicates if thinking/reasoning mode is enabled
	HasThinking bool

	// MaxTokens is the requested maximum output tokens (if specified)
	MaxTokens int

	// Temperature is the sampling temperature (if specified)
	Temperature float64

	// OriginalProtocol identifies the source API format
	OriginalProtocol string
}

// NormalizedMessage represents a single message in protocol-agnostic format.
type NormalizedMessage struct {
	// Role is the message role (user, assistant, system)
	Role string

	// Content is the text content of the message
	Content string

	// HasImage indicates if this message contains image content
	HasImage bool

	// TokenCount is the estimated token count for this message
	TokenCount int
}

// RequestFeatures contains extracted features used for routing classification.
type RequestFeatures struct {
	// HasImage indicates if any message contains image content
	HasImage bool

	// HasTools indicates if the request includes tool definitions
	HasTools bool

	// HasWebSearch indicates if the request includes web_search tool
	HasWebSearch bool

	// HasThinking indicates if thinking/reasoning mode is enabled
	HasThinking bool

	// IsLongContext indicates if the total token count exceeds the threshold
	IsLongContext bool

	// MessageCount is the number of messages in the conversation
	MessageCount int

	// TotalTokens is the estimated total token count
	TotalTokens int

	// Model is the requested model
	Model string
}

// NormalizeAnthropicMessages normalizes an Anthropic Messages API request.
func NormalizeAnthropicMessages(body map[string]interface{}) (*NormalizedRequest, error) {
	if body == nil {
		return nil, fmt.Errorf("request body is nil")
	}

	// Extract model (required)
	model, ok := body["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("missing or invalid 'model' field")
	}

	// Extract messages (required)
	messagesRaw, ok := body["messages"]
	if !ok {
		return nil, fmt.Errorf("missing 'messages' field")
	}

	messages, ok := messagesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("'messages' field is not an array")
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("'messages' array is empty")
	}

	// Extract system prompt (optional)
	systemPrompt, _ := body["system"].(string)

	// Normalize messages
	normalized := &NormalizedRequest{
		Model:            model,
		SystemPrompt:     systemPrompt,
		OriginalProtocol: "anthropic",
		Messages:         make([]NormalizedMessage, 0, len(messages)),
	}

	for _, msgRaw := range messages {
		msg, ok := msgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msg["role"].(string)
		if role == "" {
			continue
		}

		// Handle both string and array content formats
		var content string
		var hasImage bool

		switch c := msg["content"].(type) {
		case string:
			content = c
		case []interface{}:
			// Multi-part content (text + images)
			for _, part := range c {
				partMap, ok := part.(map[string]interface{})
				if !ok {
					continue
				}
				partType, _ := partMap["type"].(string)
				if partType == "text" {
					if text, ok := partMap["text"].(string); ok {
						content += text
					}
				} else if partType == "image" {
					hasImage = true
				}
			}
		}

		normalized.Messages = append(normalized.Messages, NormalizedMessage{
			Role:       role,
			Content:    content,
			HasImage:   hasImage,
			TokenCount: estimateTokens(content),
		})
	}

	// Extract optional fields
	if maxTokens, ok := body["max_tokens"].(float64); ok {
		normalized.MaxTokens = int(maxTokens)
	}
	if temp, ok := body["temperature"].(float64); ok {
		normalized.Temperature = temp
	}
	if tools, ok := body["tools"].([]interface{}); ok && len(tools) > 0 {
		normalized.HasTools = true
		// Check for web_search tool
		for _, tool := range tools {
			t, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}
			if toolType, ok := t["type"].(string); ok && strings.HasPrefix(toolType, "web_search") {
				normalized.HasWebSearch = true
				break
			}
		}
	}

	// Check for thinking mode
	if thinking, ok := body["thinking"]; ok {
		// Check if thinking is a boolean true
		if b, ok := thinking.(bool); ok {
			normalized.HasThinking = b
		} else if m, ok := thinking.(map[string]interface{}); ok {
			// Check if thinking is a map with type="enabled"
			if t, ok := m["type"].(string); ok {
				normalized.HasThinking = (t == "enabled")
			}
		}
	}

	return normalized, nil
}

// NormalizeOpenAIChat normalizes an OpenAI Chat Completions API request.
func NormalizeOpenAIChat(body map[string]interface{}) (*NormalizedRequest, error) {
	if body == nil {
		return nil, fmt.Errorf("request body is nil")
	}

	// Extract model (required)
	model, ok := body["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("missing or invalid 'model' field")
	}

	// Extract messages (required)
	messagesRaw, ok := body["messages"]
	if !ok {
		return nil, fmt.Errorf("missing 'messages' field")
	}

	messages, ok := messagesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("'messages' field is not an array")
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("'messages' array is empty")
	}

	normalized := &NormalizedRequest{
		Model:            model,
		OriginalProtocol: "openai_chat",
		Messages:         make([]NormalizedMessage, 0, len(messages)),
	}

	// Process messages, extracting system prompt if present
	for _, msgRaw := range messages {
		msg, ok := msgRaw.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msg["role"].(string)
		if role == "" {
			continue
		}

		// Handle system message separately
		if role == "system" {
			if content, ok := msg["content"].(string); ok {
				normalized.SystemPrompt = content
			}
			continue
		}

		// Handle both string and array content formats
		var content string
		var hasImage bool

		switch c := msg["content"].(type) {
		case string:
			content = c
		case []interface{}:
			// Multi-part content (text + images)
			for _, part := range c {
				partMap, ok := part.(map[string]interface{})
				if !ok {
					continue
				}
				partType, _ := partMap["type"].(string)
				if partType == "text" {
					if text, ok := partMap["text"].(string); ok {
						content += text
					}
				} else if partType == "image_url" {
					hasImage = true
				}
			}
		}

		normalized.Messages = append(normalized.Messages, NormalizedMessage{
			Role:       role,
			Content:    content,
			HasImage:   hasImage,
			TokenCount: estimateTokens(content),
		})
	}

	// Extract optional fields
	if maxTokens, ok := body["max_tokens"].(float64); ok {
		normalized.MaxTokens = int(maxTokens)
	}
	if temp, ok := body["temperature"].(float64); ok {
		normalized.Temperature = temp
	}
	if tools, ok := body["tools"].([]interface{}); ok && len(tools) > 0 {
		normalized.HasTools = true
		// Check for web_search tool
		for _, tool := range tools {
			t, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}
			if toolType, ok := t["type"].(string); ok && strings.HasPrefix(toolType, "web_search") {
				normalized.HasWebSearch = true
				break
			}
		}
	}
	if functions, ok := body["functions"].([]interface{}); ok && len(functions) > 0 {
		normalized.HasTools = true
	}

	// Check for thinking mode (OpenAI reasoning models or explicit thinking parameter)
	if thinking, ok := body["thinking"]; ok {
		if b, ok := thinking.(bool); ok {
			normalized.HasThinking = b
		} else if m, ok := thinking.(map[string]interface{}); ok {
			if t, ok := m["type"].(string); ok {
				normalized.HasThinking = (t == "enabled")
			}
		}
	}

	return normalized, nil
}

// NormalizeOpenAIResponses normalizes an OpenAI Responses API request.
func NormalizeOpenAIResponses(body map[string]interface{}) (*NormalizedRequest, error) {
	if body == nil {
		return nil, fmt.Errorf("request body is nil")
	}

	// Extract model (required)
	model, ok := body["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("missing or invalid 'model' field")
	}

	// Extract input (required)
	inputRaw, ok := body["input"]
	if !ok {
		return nil, fmt.Errorf("missing 'input' field")
	}

	normalized := &NormalizedRequest{
		Model:            model,
		OriginalProtocol: "openai_responses",
		Messages:         make([]NormalizedMessage, 0),
	}

	// Handle both string and array input formats
	switch input := inputRaw.(type) {
	case string:
		normalized.Messages = append(normalized.Messages, NormalizedMessage{
			Role:       "user",
			Content:    input,
			TokenCount: estimateTokens(input),
		})
	case []interface{}:
		// Handle structured input items (text, image, input_text, output_text, etc.)
		for _, item := range input {
			// Handle string items (legacy format)
			if str, ok := item.(string); ok {
				normalized.Messages = append(normalized.Messages, NormalizedMessage{
					Role:       "user",
					Content:    str,
					TokenCount: estimateTokens(str),
				})
				continue
			}

			// Handle structured items (new format)
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			itemType, _ := itemMap["type"].(string)
			switch itemType {
			case "text", "input_text":
				// Both "text" and "input_text" are text content
				if text, ok := itemMap["text"].(string); ok {
					normalized.Messages = append(normalized.Messages, NormalizedMessage{
						Role:       "user",
						Content:    text,
						TokenCount: estimateTokens(text),
					})
				}
			case "output_text":
				// Assistant output text
				if text, ok := itemMap["text"].(string); ok {
					normalized.Messages = append(normalized.Messages, NormalizedMessage{
						Role:       "assistant",
						Content:    text,
						TokenCount: estimateTokens(text),
					})
				}
			case "image":
				// Image item detected
				normalized.Messages = append(normalized.Messages, NormalizedMessage{
					Role:     "user",
					HasImage: true,
				})
			}
		}
	default:
		return nil, fmt.Errorf("'input' field must be string or array")
	}

	if len(normalized.Messages) == 0 {
		return nil, fmt.Errorf("no valid input messages found")
	}

	// Extract optional fields
	if tools, ok := body["tools"].([]interface{}); ok && len(tools) > 0 {
		normalized.HasTools = true
		// Check for web_search tool
		for _, tool := range tools {
			t, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}
			if toolType, ok := t["type"].(string); ok && strings.HasPrefix(toolType, "web_search") {
				normalized.HasWebSearch = true
				break
			}
		}
	}

	// Check for thinking mode
	if thinking, ok := body["thinking"]; ok {
		if b, ok := thinking.(bool); ok {
			normalized.HasThinking = b
		} else if m, ok := thinking.(map[string]interface{}); ok {
			if t, ok := m["type"].(string); ok {
				normalized.HasThinking = (t == "enabled")
			}
		}
	}

	return normalized, nil
}

// ExtractFeatures extracts routing-relevant features from a normalized request.
func ExtractFeatures(normalized *NormalizedRequest) *RequestFeatures {
	if normalized == nil {
		return &RequestFeatures{}
	}

	features := &RequestFeatures{
		Model:        normalized.Model,
		HasTools:     normalized.HasTools,
		HasWebSearch: normalized.HasWebSearch,
		HasThinking:  normalized.HasThinking,
		MessageCount: len(normalized.Messages),
	}

	// Check for images and calculate total tokens
	for _, msg := range normalized.Messages {
		if msg.HasImage {
			features.HasImage = true
		}
		features.TotalTokens += msg.TokenCount
	}

	// Determine if this is a long context request (threshold: 32000 tokens)
	// This is a default threshold; actual threshold comes from profile config
	features.IsLongContext = features.TotalTokens > 32000

	return features
}

// estimateTokens estimates token count for a text string.
// Uses tiktoken if available, falls back to character-based estimation.
func estimateTokens(text string) int {
	enc, err := getTokenEncoder()
	if err != nil {
		// Fallback: ~4 characters per token
		return len(text) / 4
	}
	return len(enc.Encode(text, nil, nil))
}

// DetectProtocol detects the API protocol from request context.
// Priority: URL path → X-Zen-Client header → body structure → default openai_chat
func DetectProtocol(path string, headers http.Header, body map[string]interface{}) string {
	// Priority 1: URL path detection
	if strings.Contains(path, "/v1/messages") || strings.Contains(path, "/messages") {
		return "anthropic"
	}
	if strings.Contains(path, "/v1/chat/completions") || strings.Contains(path, "/chat/completions") {
		return "openai_chat"
	}
	if strings.Contains(path, "/v1/completions") || strings.Contains(path, "/completions") {
		// Check if it's the Responses API (has "input" field) or legacy Completions API
		if body != nil {
			if _, hasInput := body["input"]; hasInput {
				return "openai_responses"
			}
		}
		return "openai_chat" // Default to chat for ambiguous /completions
	}

	// Priority 2: X-Zen-Client header
	if clientHeader := headers.Get("X-Zen-Client"); clientHeader != "" {
		switch strings.ToLower(clientHeader) {
		case "anthropic", "claude":
			return "anthropic"
		case "openai", "openai_chat":
			return "openai_chat"
		case "openai_responses":
			return "openai_responses"
		}
	}

	// Priority 3: Body structure detection
	if body != nil {
		// Anthropic Messages API has "messages" array and typically "model" starting with "claude"
		if _, hasMessages := body["messages"]; hasMessages {
			if model, hasModel := body["model"].(string); hasModel {
				if strings.HasPrefix(model, "claude") {
					return "anthropic"
				}
			}
			// Has messages but not Claude model - likely OpenAI Chat
			return "openai_chat"
		}

		// OpenAI Responses API has "input" field
		if _, hasInput := body["input"]; hasInput {
			return "openai_responses"
		}
	}

	// Priority 4: Default to openai_chat (most common)
	return "openai_chat"
}

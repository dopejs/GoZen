// Package transform provides request/response transformation between different API formats.
package transform

import (
	"encoding/json"
	"strings"
)

// Protocol format identifiers for client-side format detection.
const (
	FormatAnthropicMessages = "anthropic-messages"
	FormatOpenAIChat        = "openai-chat"
	FormatOpenAIResponses   = "openai-responses"
)

// Transformer defines the interface for API format transformation.
type Transformer interface {
	// Name returns the transformer name (e.g., "anthropic", "openai")
	Name() string

	// TransformRequest transforms an incoming request body to the target format.
	// clientFormat is the format the client is using (e.g., "anthropic" for Claude Code)
	// Returns the transformed body and any error.
	TransformRequest(body []byte, clientFormat string) ([]byte, error)

	// TransformResponse transforms an outgoing response body from the target format.
	// clientFormat is the format the client expects.
	// Returns the transformed body and any error.
	TransformResponse(body []byte, clientFormat string) ([]byte, error)
}

// GetTransformer returns the appropriate transformer for the given provider type.
func GetTransformer(providerType string) Transformer {
	switch providerType {
	case "openai":
		return &OpenAITransformer{}
	default:
		return &AnthropicTransformer{}
	}
}

// NeedsTransform returns true if transformation is needed between client and provider formats.
func NeedsTransform(clientFormat, providerFormat string) bool {
	// Normalize empty to anthropic (default)
	if clientFormat == "" {
		clientFormat = "anthropic"
	}
	if providerFormat == "" {
		providerFormat = "anthropic"
	}

	// Normalize new format constants to legacy provider types for comparison
	normalizedClient := NormalizeFormat(clientFormat)
	normalizedProvider := NormalizeFormat(providerFormat)

	return normalizedClient != normalizedProvider
}

// NormalizeFormat converts fine-grained format identifiers to provider types.
// anthropic-messages → anthropic
// openai-chat → openai
// openai-responses → openai
// empty string → anthropic (default)
func NormalizeFormat(format string) string {
	// Empty defaults to anthropic
	if format == "" {
		return "anthropic"
	}

	switch format {
	case FormatAnthropicMessages:
		return "anthropic"
	case FormatOpenAIChat, FormatOpenAIResponses:
		return "openai"
	default:
		return format // legacy "openai" or "anthropic"
	}
}

// TransformPath converts API endpoint paths between OpenAI and Anthropic formats.
// clientFormat: the format the client is using ("openai-chat", "openai-responses", "anthropic-messages", or legacy "openai"/"anthropic")
// providerFormat: the format the provider expects ("openai" or "anthropic")
// path: the original request path
// Returns the transformed path.
func TransformPath(clientFormat, providerFormat, path string) string {
	// Normalize formats to base types for path transformation
	normalizedClient := NormalizeFormat(clientFormat)
	normalizedProvider := NormalizeFormat(providerFormat)

	// No transformation needed if formats match
	if normalizedClient == normalizedProvider {
		return path
	}

	// OpenAI client → Anthropic provider
	if normalizedClient == "openai" && normalizedProvider == "anthropic" {
		// OpenAI Responses API: /v1/responses or /responses
		if strings.HasSuffix(path, "/responses") || strings.Contains(path, "/responses/") {
			return "/v1/messages"
		}
		// OpenAI Chat Completions API: /v1/chat/completions
		if strings.HasSuffix(path, "/chat/completions") {
			return "/v1/messages"
		}
	}

	// Anthropic client → OpenAI provider
	if normalizedClient == "anthropic" && normalizedProvider == "openai" {
		// Anthropic Messages API: /v1/messages
		if strings.HasSuffix(path, "/messages") {
			return "/v1/chat/completions"
		}
	}

	return path
}

// parseJSON is a helper to parse JSON body into a map.
func parseJSON(body []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// toJSON is a helper to convert a map back to JSON.
func toJSON(data map[string]interface{}) ([]byte, error) {
	return json.Marshal(data)
}

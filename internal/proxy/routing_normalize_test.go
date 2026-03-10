package proxy

import (
	"encoding/json"
	"testing"
)

// TestNormalizeAnthropicMessages tests normalization of Anthropic Messages API requests
func TestNormalizeAnthropicMessages(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantModel   string
		wantSystem  string
		wantMsgLen  int
		wantErr     bool
	}{
		{
			name: "basic anthropic request",
			requestBody: `{
				"model": "claude-3-opus-20240229",
				"messages": [
					{"role": "user", "content": "Hello"}
				],
				"max_tokens": 1024
			}`,
			wantModel:  "claude-3-opus-20240229",
			wantMsgLen: 1,
			wantErr:    false,
		},
		{
			name: "anthropic with system message",
			requestBody: `{
				"model": "claude-3-sonnet-20240229",
				"system": "You are a helpful assistant",
				"messages": [
					{"role": "user", "content": "Hello"}
				],
				"max_tokens": 1024
			}`,
			wantModel:  "claude-3-sonnet-20240229",
			wantSystem: "You are a helpful assistant",
			wantMsgLen: 1,
			wantErr:    false,
		},
		{
			name: "anthropic with multiple messages",
			requestBody: `{
				"model": "claude-3-haiku-20240307",
				"messages": [
					{"role": "user", "content": "Hello"},
					{"role": "assistant", "content": "Hi there!"},
					{"role": "user", "content": "How are you?"}
				],
				"max_tokens": 1024
			}`,
			wantModel:  "claude-3-haiku-20240307",
			wantMsgLen: 3,
			wantErr:    false,
		},
		{
			name: "anthropic with image content",
			requestBody: `{
				"model": "claude-3-opus-20240229",
				"messages": [
					{
						"role": "user",
						"content": [
							{"type": "text", "text": "What's in this image?"},
							{"type": "image", "source": {"type": "base64", "media_type": "image/jpeg", "data": "..."}}
						]
					}
				],
				"max_tokens": 1024
			}`,
			wantModel:  "claude-3-opus-20240229",
			wantMsgLen: 1,
			wantErr:    false,
		},
		{
			name:        "malformed json",
			requestBody: `{invalid json`,
			wantErr:     true,
		},
		{
			name: "missing model field",
			requestBody: `{
				"messages": [
					{"role": "user", "content": "Hello"}
				]
			}`,
			wantErr: true,
		},
		{
			name: "missing messages field",
			requestBody: `{
				"model": "claude-3-opus-20240229"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body map[string]interface{}
			if err := json.Unmarshal([]byte(tt.requestBody), &body); err != nil && !tt.wantErr {
				t.Fatalf("failed to parse test request body: %v", err)
			}

			normalized, err := NormalizeAnthropicMessages(body)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeAnthropicMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if normalized.Model != tt.wantModel {
				t.Errorf("Model = %v, want %v", normalized.Model, tt.wantModel)
			}

			if tt.wantSystem != "" && normalized.SystemPrompt != tt.wantSystem {
				t.Errorf("SystemPrompt = %v, want %v", normalized.SystemPrompt, tt.wantSystem)
			}

			if len(normalized.Messages) != tt.wantMsgLen {
				t.Errorf("Messages length = %v, want %v", len(normalized.Messages), tt.wantMsgLen)
			}
		})
	}
}

// TestNormalizeOpenAIChat tests normalization of OpenAI Chat Completions API requests
func TestNormalizeOpenAIChat(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantModel   string
		wantSystem  string
		wantMsgLen  int
		wantErr     bool
	}{
		{
			name: "basic openai chat request",
			requestBody: `{
				"model": "gpt-4",
				"messages": [
					{"role": "user", "content": "Hello"}
				]
			}`,
			wantModel:  "gpt-4",
			wantMsgLen: 1,
			wantErr:    false,
		},
		{
			name: "openai with system message",
			requestBody: `{
				"model": "gpt-4-turbo",
				"messages": [
					{"role": "system", "content": "You are a helpful assistant"},
					{"role": "user", "content": "Hello"}
				]
			}`,
			wantModel:  "gpt-4-turbo",
			wantSystem: "You are a helpful assistant",
			wantMsgLen: 1,
			wantErr:    false,
		},
		{
			name: "openai with multiple messages",
			requestBody: `{
				"model": "gpt-3.5-turbo",
				"messages": [
					{"role": "user", "content": "Hello"},
					{"role": "assistant", "content": "Hi there!"},
					{"role": "user", "content": "How are you?"}
				]
			}`,
			wantModel:  "gpt-3.5-turbo",
			wantMsgLen: 3,
			wantErr:    false,
		},
		{
			name: "openai with vision content",
			requestBody: `{
				"model": "gpt-4-vision-preview",
				"messages": [
					{
						"role": "user",
						"content": [
							{"type": "text", "text": "What's in this image?"},
							{"type": "image_url", "image_url": {"url": "https://example.com/image.jpg"}}
						]
					}
				]
			}`,
			wantModel:  "gpt-4-vision-preview",
			wantMsgLen: 1,
			wantErr:    false,
		},
		{
			name:        "malformed json",
			requestBody: `{invalid json`,
			wantErr:     true,
		},
		{
			name: "missing model field",
			requestBody: `{
				"messages": [
					{"role": "user", "content": "Hello"}
				]
			}`,
			wantErr: true,
		},
		{
			name: "missing messages field",
			requestBody: `{
				"model": "gpt-4"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body map[string]interface{}
			if err := json.Unmarshal([]byte(tt.requestBody), &body); err != nil && !tt.wantErr {
				t.Fatalf("failed to parse test request body: %v", err)
			}

			normalized, err := NormalizeOpenAIChat(body)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeOpenAIChat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if normalized.Model != tt.wantModel {
				t.Errorf("Model = %v, want %v", normalized.Model, tt.wantModel)
			}

			if tt.wantSystem != "" && normalized.SystemPrompt != tt.wantSystem {
				t.Errorf("SystemPrompt = %v, want %v", normalized.SystemPrompt, tt.wantSystem)
			}

			if len(normalized.Messages) != tt.wantMsgLen {
				t.Errorf("Messages length = %v, want %v", len(normalized.Messages), tt.wantMsgLen)
			}
		})
	}
}

// TestNormalizeOpenAIResponses tests normalization of OpenAI Responses API requests
func TestNormalizeOpenAIResponses(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantModel   string
		wantMsgLen  int
		wantErr     bool
	}{
		{
			name: "basic openai responses request",
			requestBody: `{
				"model": "gpt-4",
				"input": "Hello, how are you?"
			}`,
			wantModel:  "gpt-4",
			wantMsgLen: 1,
			wantErr:    false,
		},
		{
			name: "openai responses with array input",
			requestBody: `{
				"model": "gpt-3.5-turbo",
				"input": ["Hello", "How are you?", "What's the weather?"]
			}`,
			wantModel:  "gpt-3.5-turbo",
			wantMsgLen: 3,
			wantErr:    false,
		},
		{
			name:        "malformed json",
			requestBody: `{invalid json`,
			wantErr:     true,
		},
		{
			name: "missing model field",
			requestBody: `{
				"input": "Hello"
			}`,
			wantErr: true,
		},
		{
			name: "missing input field",
			requestBody: `{
				"model": "gpt-4"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body map[string]interface{}
			if err := json.Unmarshal([]byte(tt.requestBody), &body); err != nil && !tt.wantErr {
				t.Fatalf("failed to parse test request body: %v", err)
			}

			normalized, err := NormalizeOpenAIResponses(body)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeOpenAIResponses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if normalized.Model != tt.wantModel {
				t.Errorf("Model = %v, want %v", normalized.Model, tt.wantModel)
			}

			if len(normalized.Messages) != tt.wantMsgLen {
				t.Errorf("Messages length = %v, want %v", len(normalized.Messages), tt.wantMsgLen)
			}
		})
	}
}

// TestMalformedRequestHandling tests error handling for malformed requests
func TestMalformedRequestHandling(t *testing.T) {
	tests := []struct {
		name        string
		requestBody map[string]interface{}
		protocol    string
		wantErr     bool
	}{
		{
			name:        "nil body",
			requestBody: nil,
			protocol:    "anthropic",
			wantErr:     true,
		},
		{
			name:        "empty body",
			requestBody: map[string]interface{}{},
			protocol:    "anthropic",
			wantErr:     true,
		},
		{
			name: "anthropic missing model",
			requestBody: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{"role": "user", "content": "Hello"},
				},
			},
			protocol: "anthropic",
			wantErr:  true,
		},
		{
			name: "openai_chat missing messages",
			requestBody: map[string]interface{}{
				"model": "gpt-4",
			},
			protocol: "openai_chat",
			wantErr:  true,
		},
		{
			name: "openai_responses missing input",
			requestBody: map[string]interface{}{
				"model": "gpt-4",
			},
			protocol: "openai_responses",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			switch tt.protocol {
			case "anthropic":
				_, err = NormalizeAnthropicMessages(tt.requestBody)
			case "openai_chat":
				_, err = NormalizeOpenAIChat(tt.requestBody)
			case "openai_responses":
				_, err = NormalizeOpenAIResponses(tt.requestBody)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestExtractFeatures tests feature extraction from normalized requests
func TestExtractFeatures(t *testing.T) {
	tests := []struct {
		name             string
		normalized       *NormalizedRequest
		wantHasImage     bool
		wantHasTools     bool
		wantIsLongCtx    bool
		wantMessageCount int
	}{
		{
			name: "simple text request",
			normalized: &NormalizedRequest{
				Model: "claude-3-opus-20240229",
				Messages: []NormalizedMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantHasImage:     false,
			wantHasTools:     false,
			wantIsLongCtx:    false,
			wantMessageCount: 1,
		},
		{
			name: "request with image",
			normalized: &NormalizedRequest{
				Model: "claude-3-opus-20240229",
				Messages: []NormalizedMessage{
					{Role: "user", Content: "What's in this image?", HasImage: true},
				},
			},
			wantHasImage:     true,
			wantHasTools:     false,
			wantIsLongCtx:    false,
			wantMessageCount: 1,
		},
		{
			name: "request with tools",
			normalized: &NormalizedRequest{
				Model: "claude-3-opus-20240229",
				Messages: []NormalizedMessage{
					{Role: "user", Content: "Call a function"},
				},
				HasTools: true,
			},
			wantHasImage:     false,
			wantHasTools:     true,
			wantIsLongCtx:    false,
			wantMessageCount: 1,
		},
		{
			name: "long context request",
			normalized: &NormalizedRequest{
				Model: "claude-3-opus-20240229",
				Messages: []NormalizedMessage{
					{Role: "user", Content: "Short message", TokenCount: 50000},
				},
			},
			wantHasImage:     false,
			wantHasTools:     false,
			wantIsLongCtx:    true,
			wantMessageCount: 1,
		},
		{
			name: "multi-turn conversation",
			normalized: &NormalizedRequest{
				Model: "claude-3-opus-20240229",
				Messages: []NormalizedMessage{
					{Role: "user", Content: "Hello"},
					{Role: "assistant", Content: "Hi there!"},
					{Role: "user", Content: "How are you?"},
				},
			},
			wantHasImage:     false,
			wantHasTools:     false,
			wantIsLongCtx:    false,
			wantMessageCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := ExtractFeatures(tt.normalized)

			if features.HasImage != tt.wantHasImage {
				t.Errorf("HasImage = %v, want %v", features.HasImage, tt.wantHasImage)
			}

			if features.HasTools != tt.wantHasTools {
				t.Errorf("HasTools = %v, want %v", features.HasTools, tt.wantHasTools)
			}

			if features.IsLongContext != tt.wantIsLongCtx {
				t.Errorf("IsLongContext = %v, want %v", features.IsLongContext, tt.wantIsLongCtx)
			}

			if features.MessageCount != tt.wantMessageCount {
				t.Errorf("MessageCount = %v, want %v", features.MessageCount, tt.wantMessageCount)
			}
		})
	}
}

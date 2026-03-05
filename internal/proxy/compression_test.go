package proxy

import (
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func TestContextCompressor_EstimateTokens(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{
		Enabled:         true,
		ThresholdTokens: 1000,
		TargetTokens:    500,
	}, nil)

	tests := []struct {
		name     string
		messages []Message
		wantMin  int
		wantMax  int
	}{
		{
			name:     "empty messages",
			messages: []Message{},
			wantMin:  0,
			wantMax:  0,
		},
		{
			name: "single short message",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			wantMin: 1,
			wantMax: 10,
		},
		{
			name: "multiple messages",
			messages: []Message{
				{Role: "user", Content: "Hello, how are you?"},
				{Role: "assistant", Content: "I'm doing well, thank you for asking!"},
			},
			wantMin: 10,
			wantMax: 30,
		},
		{
			name: "long message",
			messages: []Message{
				{Role: "user", Content: string(make([]byte, 1000))},
			},
			wantMin: 200,
			wantMax: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compressor.EstimateTokens(tt.messages)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("EstimateTokens() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestContextCompressor_ShouldCompress(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.CompressionConfig
		messages  []Message
		want      bool
	}{
		{
			name: "disabled compression",
			config: &config.CompressionConfig{
				Enabled:         false,
				ThresholdTokens: 100,
			},
			messages: []Message{
				{Role: "user", Content: string(make([]byte, 1000))},
			},
			want: false,
		},
		{
			name: "below threshold",
			config: &config.CompressionConfig{
				Enabled:         true,
				ThresholdTokens: 1000,
			},
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			want: false,
		},
		{
			name: "above threshold",
			config: &config.CompressionConfig{
				Enabled:         true,
				ThresholdTokens: 100,
			},
			messages: []Message{
				{Role: "user", Content: string(make([]byte, 1000))},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewContextCompressor(tt.config, nil)
			got := compressor.ShouldCompress(tt.messages)
			if got != tt.want {
				t.Errorf("ShouldCompress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextCompressor_IsEnabled(t *testing.T) {
	tests := []struct {
		name   string
		config *config.CompressionConfig
		want   bool
	}{
		{
			name:   "nil config",
			config: nil,
			want:   false,
		},
		{
			name: "disabled",
			config: &config.CompressionConfig{
				Enabled: false,
			},
			want: false,
		},
		{
			name: "enabled",
			config: &config.CompressionConfig{
				Enabled: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewContextCompressor(tt.config, nil)
			got := compressor.IsEnabled()
			if got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstimateContentLength(t *testing.T) {
	tests := []struct {
		name    string
		content interface{}
		want    int
	}{
		{
			name:    "string content",
			content: "Hello, world!",
			want:    13,
		},
		{
			name:    "empty string",
			content: "",
			want:    0,
		},
		{
			name: "array content with text",
			content: []interface{}{
				map[string]interface{}{"type": "text", "text": "Hello"},
				map[string]interface{}{"type": "text", "text": "World"},
			},
			want: 10,
		},
		{
			name:    "nil content",
			content: nil,
			want:    4, // json.Marshal(nil) returns "null"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateContentLength(tt.content)
			if got != tt.want {
				t.Errorf("estimateContentLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextCompressor_GetStats(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{
		Enabled: true,
	}, nil)

	stats := compressor.GetStats()
	if stats.RequestsCompressed != 0 {
		t.Errorf("Initial RequestsCompressed = %v, want 0", stats.RequestsCompressed)
	}
	if stats.TokensSaved != 0 {
		t.Errorf("Initial TokensSaved = %v, want 0", stats.TokensSaved)
	}
}

func TestContextCompressor_UpdateConfig(t *testing.T) {
	compressor := NewContextCompressor(nil, nil)

	newCfg := &config.CompressionConfig{
		Enabled:         true,
		ThresholdTokens: 30000,
		TargetTokens:    15000,
	}
	compressor.UpdateConfig(newCfg)

	if !compressor.IsEnabled() {
		t.Error("Expected compressor to be enabled after update")
	}
	if compressor.config.ThresholdTokens != 30000 {
		t.Errorf("Expected threshold 30000, got %d", compressor.config.ThresholdTokens)
	}
}

func TestContextCompressor_SetProviders(t *testing.T) {
	compressor := NewContextCompressor(nil, nil)

	providers := []*Provider{
		{Name: "p1", Healthy: true},
		{Name: "p2", Healthy: true},
	}
	compressor.SetProviders(providers)

	compressor.mu.RLock()
	if len(compressor.providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(compressor.providers))
	}
	compressor.mu.RUnlock()
}

func TestContextCompressor_Compress_Disabled(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{Enabled: false}, nil)

	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	result, err := compressor.Compress(messages)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != len(messages) {
		t.Error("Expected messages unchanged when disabled")
	}
}

func TestContextCompressor_Compress_EmptyMessages(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{Enabled: true}, nil)

	result, err := compressor.Compress([]Message{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty input")
	}
}

func TestContextCompressor_Compress_FewMessages(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{
		Enabled:        true,
		PreserveRecent: 4,
	}, nil)

	// Only 3 messages, less than PreserveRecent
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
		{Role: "user", Content: "How are you?"},
	}

	result, err := compressor.Compress(messages)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Error("Expected messages unchanged when fewer than PreserveRecent")
	}
}

func TestContextCompressor_Summarize_NoProvider(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{Enabled: true}, nil)

	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := compressor.Summarize(messages)
	if err == nil {
		t.Error("Expected error when no provider available")
	}
}

func TestContextCompressor_CompressRequestBody_Disabled(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{Enabled: false}, nil)

	body := []byte(`{"messages":[{"role":"user","content":"Hello"}]}`)
	result, compressed, err := compressor.CompressRequestBody(body)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if compressed {
		t.Error("Should not compress when disabled")
	}
	if string(result) != string(body) {
		t.Error("Body should be unchanged")
	}
}

func TestContextCompressor_CompressRequestBody_InvalidJSON(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{Enabled: true}, nil)

	body := []byte(`not json`)
	result, compressed, err := compressor.CompressRequestBody(body)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if compressed {
		t.Error("Should not compress invalid JSON")
	}
	if string(result) != string(body) {
		t.Error("Body should be unchanged")
	}
}

func TestContextCompressor_CompressRequestBody_NoMessages(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{Enabled: true}, nil)

	body := []byte(`{"model":"test"}`)
	result, compressed, err := compressor.CompressRequestBody(body)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if compressed {
		t.Error("Should not compress when no messages")
	}
	if string(result) != string(body) {
		t.Error("Body should be unchanged")
	}
}

func TestContextCompressor_CompressRequestBody_BelowThreshold(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{
		Enabled:         true,
		ThresholdTokens: 100000,
	}, nil)

	body := []byte(`{"messages":[{"role":"user","content":"Hello"}]}`)
	result, compressed, err := compressor.CompressRequestBody(body)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if compressed {
		t.Error("Should not compress below threshold")
	}
	if string(result) != string(body) {
		t.Error("Body should be unchanged")
	}
}

func TestGlobalCompressorFunctions(t *testing.T) {
	// Test InitGlobalCompressor
	InitGlobalCompressor(nil)

	c := GetGlobalCompressor()
	if c == nil {
		t.Fatal("Expected non-nil global compressor")
	}

	// Test UpdateGlobalCompressorConfig
	UpdateGlobalCompressorConfig(&config.CompressionConfig{
		Enabled:         true,
		ThresholdTokens: 25000,
	})

	if !c.IsEnabled() {
		t.Error("Expected compressor to be enabled after update")
	}

	// Test UpdateGlobalCompressorProviders
	providers := []*Provider{{Name: "test", Healthy: true}}
	UpdateGlobalCompressorProviders(providers)

	c.mu.RLock()
	if len(c.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(c.providers))
	}
	c.mu.RUnlock()
}

func TestContextCompressor_ShouldCompress_DefaultThreshold(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{
		Enabled:         true,
		ThresholdTokens: 0, // Should use default
	}, nil)

	// Short message - should not compress
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}
	if compressor.ShouldCompress(messages) {
		t.Error("Should not compress short messages")
	}
}

func TestContextCompressor_Compress_DefaultPreserveRecent(t *testing.T) {
	compressor := NewContextCompressor(&config.CompressionConfig{
		Enabled:        true,
		PreserveRecent: 0, // Should use default (4)
	}, nil)

	// 3 messages - less than default PreserveRecent
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
		{Role: "user", Content: "Bye"},
	}

	result, err := compressor.Compress(messages)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Error("Expected messages unchanged")
	}
}

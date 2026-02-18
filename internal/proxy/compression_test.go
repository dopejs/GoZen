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

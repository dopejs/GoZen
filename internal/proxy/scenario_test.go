package proxy

import (
	"strings"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

// generateLongText creates varied text to get realistic token counts.
// Approximately 4 characters per token for English text.
func generateLongText(chars int) string {
	var sb strings.Builder
	words := []string{"hello", "world", "this", "is", "a", "test", "message", "with", "varied", "content"}
	wordIndex := 0
	for sb.Len() < chars {
		if wordIndex > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(words[wordIndex%len(words)])
		wordIndex++
	}
	return sb.String()
}


func TestDetectScenarioThink(t *testing.T) {
	body := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"thinking": map[string]interface{}{"type": "enabled"},
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hi"},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioThink {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioThink)
	}
}

func TestDetectScenarioThinkDisabled(t *testing.T) {
	body := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"thinking": map[string]interface{}{"type": "disabled"},
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hi"},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioDefault {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioDefault)
	}
}

func TestDetectScenarioImage(t *testing.T) {
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "What is this?"},
					map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "image/png",
							"data":       "iVBOR...",
						},
					},
				},
			},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioImage {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioImage)
	}
}

func TestDetectScenarioLongContext(t *testing.T) {
	// Generate text that will exceed token threshold
	// Using varied text to get realistic token count (~5.5 chars per token)
	longText := generateLongText(defaultLongContextThreshold * 6)
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": longText},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioLongContext)
	}
}

func TestDetectScenarioLongContextFromBlocks(t *testing.T) {
	longText := generateLongText(defaultLongContextThreshold * 6)
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": longText},
				},
			},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioLongContext)
	}
}

func TestDetectScenarioLongContextFromSystem(t *testing.T) {
	system := generateLongText(defaultLongContextThreshold * 6)
	body := map[string]interface{}{
		"model":  "claude-sonnet-4-5",
		"system": system,
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hi"},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioLongContext)
	}
}

func TestDetectScenarioDefault(t *testing.T) {
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hello"},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioDefault {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioDefault)
	}
}

func TestDetectScenarioPriority_ThinkOverImage(t *testing.T) {
	body := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"thinking": map[string]interface{}{"type": "enabled"},
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "image", "source": map[string]interface{}{}},
				},
			},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioThink {
		t.Errorf("DetectScenario() = %q, want %q (think takes priority over image)", got, config.ScenarioThink)
	}
}

func TestDetectScenarioPriority_ImageOverLongContext(t *testing.T) {
	longText := generateLongText(defaultLongContextThreshold * 6)
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": longText},
					map[string]interface{}{"type": "image", "source": map[string]interface{}{}},
				},
			},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioImage {
		t.Errorf("DetectScenario() = %q, want %q (image takes priority over longContext)", got, config.ScenarioImage)
	}
}

func TestDetectScenarioFromJSON(t *testing.T) {
	data := []byte(`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"hi"}]}`)
	scenario, body := DetectScenarioFromJSON(data, 0, "")
	if scenario != config.ScenarioThink {
		t.Errorf("scenario = %q, want %q", scenario, config.ScenarioThink)
	}
	if body == nil {
		t.Error("body should not be nil")
	}
}

func TestDetectScenarioFromJSONInvalid(t *testing.T) {
	scenario, body := DetectScenarioFromJSON([]byte("not json"), 0, "")
	if scenario != config.ScenarioDefault {
		t.Errorf("scenario = %q, want %q for invalid JSON", scenario, config.ScenarioDefault)
	}
	if body != nil {
		t.Error("body should be nil for invalid JSON")
	}
}

func TestHasImageContentNoMessages(t *testing.T) {
	body := map[string]interface{}{}
	if hasImageContent(body) {
		t.Error("expected false for empty body")
	}
}

func TestIsLongContextShort(t *testing.T) {
	body := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "short"},
		},
	}
	if isLongContext(body, 0, "") {
		t.Error("expected false for short content")
	}
}

func TestIsLongContextMultipleMessages(t *testing.T) {
	halfText := generateLongText(defaultLongContextThreshold * 3)
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": halfText},
			map[string]interface{}{"role": "assistant", "content": halfText},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() = %q, want %q for multiple messages totaling > threshold", got, config.ScenarioLongContext)
	}
}

func TestDetectScenarioWebSearch(t *testing.T) {
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"tools": []interface{}{
			map[string]interface{}{
				"type": "web_search_20241111",
				"name": "web_search",
			},
		},
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "search for something"},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioWebSearch {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioWebSearch)
	}
}

func TestDetectScenarioBackground(t *testing.T) {
	body := map[string]interface{}{
		"model": "claude-3-5-haiku-20241022",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "quick task"},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioBackground {
		t.Errorf("DetectScenario() = %q, want %q", got, config.ScenarioBackground)
	}
}

func TestDetectScenarioPriority_WebSearchOverThink(t *testing.T) {
	body := map[string]interface{}{
		"model":    "claude-sonnet-4-5",
		"thinking": map[string]interface{}{"type": "enabled"},
		"tools": []interface{}{
			map[string]interface{}{"type": "web_search_20241111"},
		},
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "search and think"},
		},
	}
	got := DetectScenario(body, 0, "")
	if got != config.ScenarioWebSearch {
		t.Errorf("DetectScenario() = %q, want %q (webSearch takes priority over think)", got, config.ScenarioWebSearch)
	}
}

func TestDetectScenarioCustomThreshold(t *testing.T) {
	text := generateLongText(40000) // ~10000 tokens
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": text},
		},
	}
	// With custom threshold of 5000, should be longContext
	got := DetectScenario(body, 5000, "")
	if got != config.ScenarioLongContext {
		t.Errorf("DetectScenario() with threshold 5000 = %q, want %q", got, config.ScenarioLongContext)
	}
	// With custom threshold of 20000, should be default
	got = DetectScenario(body, 20000, "")
	if got != config.ScenarioDefault {
		t.Errorf("DetectScenario() with threshold 20000 = %q, want %q", got, config.ScenarioDefault)
	}
}

func TestSessionCacheIntegration(t *testing.T) {
	sessionID := "test-session-123"

	// Create a request that's just below the threshold (~25000 tokens)
	text := generateLongText(140000) // ~25000 tokens
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": text},
		},
	}

	// First request: should be default (below threshold of 30000)
	got := DetectScenario(body, 30000, sessionID)
	if got != config.ScenarioDefault {
		t.Errorf("first request: got %q, want %q", got, config.ScenarioDefault)
	}

	// Simulate a large previous request
	UpdateSessionUsage(sessionID, &SessionUsage{
		InputTokens:  50000, // Above threshold
		OutputTokens: 5000,
	})

	// Second request: should be longContext due to session history
	// (current request > 20000 tokens and last request > threshold)
	got = DetectScenario(body, 30000, sessionID)
	if got != config.ScenarioLongContext {
		t.Errorf("second request with session history: got %q, want %q", got, config.ScenarioLongContext)
	}

	// Third request with small content: should be default
	// (current request < 20000 tokens)
	smallBody := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hi"},
		},
	}
	got = DetectScenario(smallBody, 30000, sessionID)
	if got != config.ScenarioDefault {
		t.Errorf("small request with session history: got %q, want %q", got, config.ScenarioDefault)
	}
}

func TestExtractSessionID(t *testing.T) {
	tests := []struct {
		name string
		body map[string]interface{}
		want string
	}{
		{
			name: "valid session ID",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"user_id": "user_session_abc123",
				},
			},
			want: "abc123",
		},
		{
			name: "no metadata",
			body: map[string]interface{}{
				"model": "claude-sonnet-4-5",
			},
			want: "",
		},
		{
			name: "no user_id",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"other": "value",
				},
			},
			want: "",
		},
		{
			name: "invalid format",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"user_id": "invalid_format",
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSessionID(tt.body)
			if got != tt.want {
				t.Errorf("extractSessionID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTokenCalculation(t *testing.T) {
	// Test basic token calculation
	body := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hello, how are you?",
			},
		},
	}

	tokens, err := calculateTokenCount(body)
	if err != nil {
		t.Fatalf("calculateTokenCount() error: %v", err)
	}

	// "Hello, how are you?" should be around 5-6 tokens
	if tokens < 3 || tokens > 10 {
		t.Errorf("calculateTokenCount() = %d, expected 3-10 tokens", tokens)
	}
}

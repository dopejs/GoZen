package transform

import (
	"encoding/json"
	"testing"
)

func TestGetTransformer(t *testing.T) {
	tests := []struct {
		providerType string
		wantName     string
	}{
		{"anthropic", "anthropic"},
		{"openai", "openai"},
		{"", "anthropic"}, // default
		{"unknown", "anthropic"}, // fallback to default
	}

	for _, tt := range tests {
		t.Run(tt.providerType, func(t *testing.T) {
			transformer := GetTransformer(tt.providerType)
			if transformer.Name() != tt.wantName {
				t.Errorf("GetTransformer(%q).Name() = %q, want %q", tt.providerType, transformer.Name(), tt.wantName)
			}
		})
	}
}

func TestNeedsTransform(t *testing.T) {
	tests := []struct {
		clientFormat   string
		providerFormat string
		want           bool
	}{
		{"anthropic", "anthropic", false},
		{"openai", "openai", false},
		{"anthropic", "openai", true},
		{"openai", "anthropic", true},
		{"", "anthropic", false},  // empty defaults to anthropic
		{"anthropic", "", false},  // empty defaults to anthropic
		{"", "", false},           // both default to anthropic
		{"", "openai", true},      // empty vs openai
		{"openai", "", true},      // openai vs empty (anthropic)
	}

	for _, tt := range tests {
		t.Run(tt.clientFormat+"_"+tt.providerFormat, func(t *testing.T) {
			got := NeedsTransform(tt.clientFormat, tt.providerFormat)
			if got != tt.want {
				t.Errorf("NeedsTransform(%q, %q) = %v, want %v", tt.clientFormat, tt.providerFormat, got, tt.want)
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", `{"key": "value"}`, false},
		{"empty object", `{}`, false},
		{"invalid", `{invalid}`, true},
		{"empty", ``, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	input := map[string]interface{}{
		"key": "value",
		"num": float64(42),
	}

	result, err := toJSON(input)
	if err != nil {
		t.Fatalf("toJSON() error = %v", err)
	}

	// Parse back to verify
	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed["key"] != "value" {
		t.Errorf("key = %v, want %v", parsed["key"], "value")
	}
	if parsed["num"] != float64(42) {
		t.Errorf("num = %v, want %v", parsed["num"], float64(42))
	}
}

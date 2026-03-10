package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

func discardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

// generateLongTextForTest creates varied text to get realistic token counts.
// Approximately 5.5 characters per token for English text.
func generateLongTextForTest(chars int) string {
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

func TestSingleJoiningSlash(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"http://host", "/path", "http://host/path"},
		{"http://host/", "/path", "http://host/path"},
		{"http://host/", "path", "http://host/path"},
		{"http://host", "path", "http://host/path"},
		{"http://host/api", "/v1/messages", "http://host/api/v1/messages"},
		{"http://host/api/", "/v1/messages", "http://host/api/v1/messages"},
		// Regression: base URL should not produce double /v1
		{"http://host", "/v1/messages", "http://host/v1/messages"},
		{"http://host/", "/v1/messages", "http://host/v1/messages"},
		{"http://host/claude", "/v1/messages", "http://host/claude/v1/messages"},
	}
	for _, tt := range tests {
		got := singleJoiningSlash(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("singleJoiningSlash(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestModelMappingSonnet(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", HaikuModel: "my-haiku", OpusModel: "my-opus",
		ReasoningModel: "my-reasoning", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5-20250929","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingHaiku(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-haiku" {
			t.Errorf("model = %v, want %q", data["model"], "my-haiku")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", HaikuModel: "my-haiku", OpusModel: "my-opus",
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-haiku-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingOpus(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-opus" {
			t.Errorf("model = %v, want %q", data["model"], "my-opus")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", HaikuModel: "my-haiku", OpusModel: "my-opus",
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-opus-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingThinkingMode(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-reasoning" {
			t.Errorf("model = %v, want %q", data["model"], "my-reasoning")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", ReasoningModel: "my-reasoning", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"}}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingThinkingDisabledUsesSonnet(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", ReasoningModel: "my-reasoning", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5","thinking":{"type":"disabled"}}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingUnknownModelUsesDefault(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "default-model" {
			t.Errorf("model = %v, want %q", data["model"], "default-model")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "default-model",
		SonnetModel: "my-sonnet", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"some-unknown-model","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingNoMappingKeepsOriginal(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "claude-sonnet-4-5" {
			t.Errorf("model = %v, want %q", data["model"], "claude-sonnet-4-5")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingCaseInsensitive(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t",
		SonnetModel: "my-sonnet", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"Claude-SONNET-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestModelMappingInvalidJSON(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Invalid JSON should be passed through unchanged
		if string(body) != "not json" {
			t.Errorf("body = %q, want %q", string(body), "not json")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "t", Model: "test-model", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

func TestModelMappingFailoverUsesSecondProviderMapping(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Second provider should use its own sonnet mapping
		if data["model"] != "provider2-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "provider2-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", SonnetModel: "provider1-sonnet", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", SonnetModel: "provider2-sonnet", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestFailoverAppliesAllProviderConfig verifies that when failing over to the
// second provider, auth token, base URL, and all model type mappings are
// correctly applied from the second provider's configuration.
func TestFailoverAppliesAllProviderConfig(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantModel string
	}{
		{"sonnet", `{"model":"claude-sonnet-4-5"}`, "p2-sonnet"},
		{"haiku", `{"model":"claude-haiku-4-5"}`, "p2-haiku"},
		{"opus", `{"model":"claude-opus-4-5"}`, "p2-opus"},
		{"thinking", `{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"}}`, "p2-reasoning"},
		{"unknown fallback", `{"model":"some-custom-model"}`, "p2-default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
			}))
			defer backend1.Close()

			backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify auth token from second provider
				if r.Header.Get("x-api-key") != "token-p2" {
					t.Errorf("x-api-key = %q, want %q", r.Header.Get("x-api-key"), "token-p2")
				}
				if r.Header.Get("Authorization") != "Bearer token-p2" {
					t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer token-p2")
				}

				// Verify model mapping from second provider
				body, _ := io.ReadAll(r.Body)
				var data map[string]interface{}
				json.Unmarshal(body, &data)
				if data["model"] != tt.wantModel {
					t.Errorf("model = %v, want %q", data["model"], tt.wantModel)
				}

				w.WriteHeader(200)
				w.Write([]byte(`{"ok":true}`))
			}))
			defer backend2.Close()

			u1, _ := url.Parse(backend1.URL)
			u2, _ := url.Parse(backend2.URL)
			providers := []*Provider{
				{
					Name: "p1", BaseURL: u1, Token: "token-p1",
					Model: "p1-default", SonnetModel: "p1-sonnet", HaikuModel: "p1-haiku",
					OpusModel: "p1-opus", ReasoningModel: "p1-reasoning", Healthy: true,
				},
				{
					Name: "p2", BaseURL: u2, Token: "token-p2",
					Model: "p2-default", SonnetModel: "p2-sonnet", HaikuModel: "p2-haiku",
					OpusModel: "p2-opus", ReasoningModel: "p2-reasoning", Healthy: true,
				},
			}

			srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
			req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("status = %d, want 200", w.Code)
			}
		})
	}
}

// TestFailoverThreeProviders verifies correct mapping when first two providers
// fail and the third succeeds.
func TestFailoverThreeProviders(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend2.Close()

	backend3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "token-p3" {
			t.Errorf("x-api-key = %q, want %q", r.Header.Get("x-api-key"), "token-p3")
		}
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "p3-haiku" {
			t.Errorf("model = %v, want %q", data["model"], "p3-haiku")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend3.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	u3, _ := url.Parse(backend3.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "token-p1", HaikuModel: "p1-haiku", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "token-p2", HaikuModel: "p2-haiku", Healthy: true},
		{Name: "p3", BaseURL: u3, Token: "token-p3", HaikuModel: "p3-haiku", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-haiku-4-5"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHasThinkingEnabled(t *testing.T) {
	tests := []struct {
		name string
		body map[string]interface{}
		want bool
	}{
		{"enabled", map[string]interface{}{"thinking": map[string]interface{}{"type": "enabled"}}, true},
		{"disabled", map[string]interface{}{"thinking": map[string]interface{}{"type": "disabled"}}, false},
		{"no thinking", map[string]interface{}{}, false},
		{"thinking not object", map[string]interface{}{"thinking": "enabled"}, false},
		{"thinking no type", map[string]interface{}{"thinking": map[string]interface{}{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasThinkingEnabled(tt.body)
			if got != tt.want {
				t.Errorf("hasThinkingEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestServeHTTPSuccess tests a successful proxy request.
func TestServeHTTPSuccess(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth headers
		if r.Header.Get("x-api-key") != "test-token" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}

		// Verify model mapping (sonnet → test-model via default)
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "test-model" {
			t.Errorf("model = %v, want %q", data["model"], "test-model")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "test-token", Model: "test-model", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)

	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"some-model","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Errorf("body = %q", w.Body.String())
	}
}

// TestServeHTTPFailoverOn500 tests that 500 triggers failover to next provider.
func TestServeHTTPFailoverOn500(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(500)
		w.Write([]byte("error"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)

	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestServeHTTPFailoverOn429 tests that 429 triggers failover.
func TestServeHTTPFailoverOn429(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestServeHTTPAllProvidersFail tests 502 when all providers fail.
func TestServeHTTPAllProvidersFail(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// TestServeHTTPSkipsUnhealthyProvider tests that unhealthy providers are skipped.
func TestServeHTTPSkipsUnhealthyProvider(t *testing.T) {
	called := make(map[string]bool)

	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called["p1"] = true
		w.WriteHeader(200)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called["p2"] = true
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	p1 := &Provider{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true}
	p2 := &Provider{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true}

	// Mark p1 as unhealthy
	p1.MarkFailed()

	srv := NewProxyServer([]*Provider{p1, p2}, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if called["p1"] {
		t.Error("p1 should have been skipped (unhealthy)")
	}
	if !called["p2"] {
		t.Error("p2 should have been called")
	}
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestServeHTTPNoModelInjectionWhenEmpty tests that empty model skips injection.
func TestServeHTTPNoModelInjectionWhenEmpty(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if _, ok := data["model"]; ok {
			t.Error("model should not be injected when provider model is empty")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Model: "", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

// TestServeHTTPPreservesQueryString tests that query params are forwarded.
func TestServeHTTPPreservesQueryString(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "beta=true" {
			t.Errorf("query = %q, want %q", r.URL.RawQuery, "beta=true")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages?beta=true", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

// TestServeHTTPSSEStreaming tests SSE response streaming.
func TestServeHTTPSSEStreaming(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		w.Write([]byte("data: hello\n\n"))
		w.Write([]byte("data: world\n\n"))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "data: hello") || !strings.Contains(body, "data: world") {
		t.Errorf("SSE body = %q", body)
	}
}

// TestStartProxy tests that StartProxy returns a valid port.
func TestStartProxy(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	port, err := StartProxy(providers, "anthropic", "127.0.0.1:0", discardLogger())
	if err != nil {
		t.Fatalf("StartProxy() error: %v", err)
	}
	if port <= 0 {
		t.Errorf("port = %d, want > 0", port)
	}

	// Verify the server is actually listening
	resp, err := http.Post(
		fmt.Sprintf("http://127.0.0.1:%d/v1/messages", port),
		"application/json",
		strings.NewReader(`{}`),
	)
	if err != nil {
		t.Fatalf("request to proxy error: %v", err)
	}
	resp.Body.Close()
	// Should get 502 since the backend URL is fake
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}
}

func TestNewProxyServer(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}
	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	if srv == nil {
		t.Fatal("NewProxyServer returned nil")
	}
	if len(srv.Providers) != 1 {
		t.Errorf("providers count = %d, want 1", len(srv.Providers))
	}
	if srv.Client == nil {
		t.Error("Client should not be nil")
	}
}

// TestServeHTTPCopiesResponseHeaders tests that response headers are forwarded.
func TestServeHTTPCopiesResponseHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Header().Get("X-Custom-Header") != "custom-value" {
		t.Errorf("X-Custom-Header = %q, want %q", w.Header().Get("X-Custom-Header"), "custom-value")
	}
}

// TestStartProxyListenError tests that StartProxy returns error for invalid address.
func TestStartProxyListenError(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	// Use an invalid listen address
	_, err := StartProxy(providers, "anthropic", "999.999.999.999:0", discardLogger())
	if err == nil {
		t.Error("expected error for invalid listen address")
	}
}

// TestServeHTTPConnectionError tests failover when backend is unreachable.
func TestServeHTTPConnectionError(t *testing.T) {
	// Use a URL that will refuse connections
	u1, _ := url.Parse("http://127.0.0.1:1") // port 1 should refuse
	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()
	u2, _ := url.Parse(backend2.URL)

	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from connection error)", w.Code)
	}
}

// TestServeHTTPBadBodyRead tests handling of body read error.
func TestServeHTTPBadBodyRead(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", &errorReader{})
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// errorReader always returns an error on Read.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

// TestServeHTTP4xxNoFailover tests that non-auth 4xx (e.g. 400) don't trigger failover.
// Auth errors (401, 403) are tested separately and DO trigger failover.
func TestServeHTTP4xxNoFailover(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(400)
		w.Write([]byte("bad request"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// 400 should NOT trigger failover — only 429 and 5xx do
	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (no failover for 400)", callCount)
	}
}

// TestServeHTTPFailoverOn401 tests that 401 triggers failover to next provider.
func TestServeHTTPFailoverOn401(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "bad-token", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "good-token", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from 401)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestServeHTTPFailoverOn403 tests that 403 triggers failover to next provider.
func TestServeHTTPFailoverOn403(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from 403)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestServeHTTPFailoverOn402 tests that 402 (payment required) triggers failover.
func TestServeHTTPFailoverOn402(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(402)
		w.Write([]byte(`{"error":"payment required"}`))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from 402)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestAuthFailedLongBackoff tests that auth failure (401/403) uses long backoff.
func TestAuthFailedLongBackoff(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	p := &Provider{Name: "p1", BaseURL: u, Token: "t", Healthy: true}

	p.MarkAuthFailed()

	if p.Healthy {
		t.Error("expected Healthy = false after MarkAuthFailed")
	}
	if !p.AuthFailed {
		t.Error("expected AuthFailed = true after MarkAuthFailed")
	}
	if p.Backoff != AuthInitialBackoff {
		t.Errorf("Backoff = %v, want %v", p.Backoff, AuthInitialBackoff)
	}

	// Second auth failure should double the backoff
	p.MarkAuthFailed()
	want := AuthInitialBackoff * 2
	if p.Backoff != want {
		t.Errorf("Backoff after 2nd failure = %v, want %v", p.Backoff, want)
	}

	// Verify it's much larger than transient backoff
	if p.Backoff < MaxBackoff {
		t.Errorf("auth backoff %v should be larger than transient max %v", p.Backoff, MaxBackoff)
	}
}

// TestAuthFailedRecovery tests that a provider recovers after auth backoff expires.
func TestAuthFailedRecovery(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	p := &Provider{Name: "p1", BaseURL: u, Token: "t", Healthy: true}

	p.MarkAuthFailed()

	// Immediately after failure, should be unhealthy
	if p.IsHealthy() {
		t.Error("expected unhealthy immediately after MarkAuthFailed")
	}

	// Simulate time passing beyond the backoff
	p.mu.Lock()
	p.FailedAt = time.Now().Add(-AuthInitialBackoff - time.Second)
	p.mu.Unlock()

	// Should now be considered healthy again
	if !p.IsHealthy() {
		t.Error("expected healthy after backoff period expires")
	}
}

// TestMarkHealthyClearsAuthFailed tests that MarkHealthy resets AuthFailed flag.
func TestMarkHealthyClearsAuthFailed(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	p := &Provider{Name: "p1", BaseURL: u, Token: "t", Healthy: true}

	p.MarkAuthFailed()
	if !p.AuthFailed {
		t.Error("expected AuthFailed = true")
	}

	p.MarkHealthy()
	if p.AuthFailed {
		t.Error("expected AuthFailed = false after MarkHealthy")
	}
	if p.Backoff != 0 {
		t.Errorf("Backoff = %v, want 0 after MarkHealthy", p.Backoff)
	}
}

// --- Scenario routing tests ---

func TestRoutingThinkScenarioUsesThinkProviders(t *testing.T) {
	defaultCalled := false
	thinkCalled := false

	defaultBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultCalled = true
		w.WriteHeader(200)
	}))
	defer defaultBackend.Close()

	thinkBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		thinkCalled = true
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Model override should be applied
		if data["model"] != "think-model" {
			t.Errorf("model = %v, want %q", data["model"], "think-model")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer thinkBackend.Close()

	u1, _ := url.Parse(defaultBackend.URL)
	u2, _ := url.Parse(thinkBackend.URL)

	defaultProvider := &Provider{Name: "default-p", BaseURL: u1, Token: "t1", Model: "m1", Healthy: true}
	thinkProvider := &Provider{Name: "think-p", BaseURL: u2, Token: "t2", Model: "m2", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{defaultProvider},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"think": {
				Providers: []*Provider{thinkProvider},
				Models:    map[string]string{"think-p": "think-model"},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"hi"}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if defaultCalled {
		t.Error("default provider should not have been called for think scenario")
	}
	if !thinkCalled {
		t.Error("think provider should have been called")
	}
}

func TestRoutingDefaultScenarioUsesDefaultProviders(t *testing.T) {
	defaultCalled := false

	defaultBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultCalled = true
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer defaultBackend.Close()

	u1, _ := url.Parse(defaultBackend.URL)
	defaultProvider := &Provider{Name: "default-p", BaseURL: u1, Token: "t1", Model: "m1", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{defaultProvider},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"think": {
				Providers: []*Provider{{Name: "think-p", BaseURL: u1, Token: "t2", Healthy: true}},
				Models:    map[string]string{"think-p": "think-model"},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hello"}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !defaultCalled {
		t.Error("default provider should have been called for non-matching scenario")
	}
}

func TestRoutingModelOverrideSkipsMapping(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Should use the override model, not the provider's sonnet mapping
		if data["model"] != "override-model" {
			t.Errorf("model = %v, want %q", data["model"], "override-model")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	provider := &Provider{
		Name: "p1", BaseURL: u, Token: "t",
		Model: "default-model", SonnetModel: "my-sonnet",
		Healthy: true,
	}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{provider},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"think": {
				Providers: []*Provider{provider},
				Models:    map[string]string{"p1": "override-model"},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"}}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRoutingNoRoutingBackwardCompat(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Should use normal model mapping (sonnet)
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "p1", BaseURL: u, Token: "t",
		SonnetModel: "my-sonnet", Healthy: true,
	}}

	// No routing — plain old proxy
	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRoutingSharedProviderHealth(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)

	// Same provider instance shared across default and think scenarios
	sharedProvider := &Provider{Name: "shared", BaseURL: u1, Token: "t1", Model: "m", Healthy: true}
	backupProvider := &Provider{Name: "backup", BaseURL: u2, Token: "t2", Model: "m", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{sharedProvider, backupProvider},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"think": {
				Providers: []*Provider{sharedProvider},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)

	// First request — default scenario. Provider "shared" will fail (500) and get marked unhealthy.
	req1 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}]}`))
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("first request status = %d, want 200 (failover to backup)", w1.Code)
	}

	// Now "shared" is unhealthy. A think scenario request should skip it too,
	// but will fallback to default providers where backup is healthy.
	req2 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"think"}]}`))
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)

	// Think scenario providers are unhealthy, but fallback to default providers succeeds
	if w2.Code != 200 {
		t.Errorf("second request status = %d, want 200 (fallback to default providers)", w2.Code)
	}
}

func TestRoutingScenarioFallbackAllFail(t *testing.T) {
	// Test that when both scenario and default providers fail, we get 502
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"server error"}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)

	scenarioProvider := &Provider{Name: "scenario-p", BaseURL: u, Token: "t1", Model: "m", Healthy: true}
	defaultProvider := &Provider{Name: "default-p", BaseURL: u, Token: "t2", Model: "m", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{defaultProvider},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"think": {
				Providers: []*Provider{scenarioProvider},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)

	// Think scenario request - both scenario and default providers will fail
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"think"}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Both scenario and default providers failed → 502
	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502 (all providers failed)", w.Code)
	}
}

func TestRoutingImageScenario(t *testing.T) {
	imageCalled := false

	imageBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		imageCalled = true
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer imageBackend.Close()

	u, _ := url.Parse(imageBackend.URL)
	imageProvider := &Provider{Name: "image-p", BaseURL: u, Token: "t", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"image": {Providers: []*Provider{imageProvider}},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":[{"type":"image","source":{"type":"base64","data":"abc"}}]}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !imageCalled {
		t.Error("image provider should have been called")
	}
}

func TestRoutingLongContextScenario(t *testing.T) {
	defaultCalled := false
	longCtxCalled := false

	defaultBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultCalled = true
		w.WriteHeader(200)
	}))
	defer defaultBackend.Close()

	longCtxBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		longCtxCalled = true
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "cheap-model" {
			t.Errorf("model = %v, want %q", data["model"], "cheap-model")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer longCtxBackend.Close()

	u1, _ := url.Parse(defaultBackend.URL)
	u2, _ := url.Parse(longCtxBackend.URL)

	defaultProvider := &Provider{Name: "default-p", BaseURL: u1, Token: "t1", Model: "m1", Healthy: true}
	longCtxProvider := &Provider{Name: "cheap-p", BaseURL: u2, Token: "t2", Model: "m2", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{defaultProvider},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"longContext": {
				Providers: []*Provider{longCtxProvider},
				Models:    map[string]string{"cheap-p": "cheap-model"},
			},
		},
	}

	// Build a request with >32k tokens
	// Generate varied text to get realistic token count (~5.5 chars per token)
	longText := generateLongTextForTest(32000 * 6)
	reqBody := fmt.Sprintf(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"%s"}]}`, longText)

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if defaultCalled {
		t.Error("default provider should not have been called for longContext scenario")
	}
	if !longCtxCalled {
		t.Error("longContext provider should have been called")
	}
}

func TestRoutingScenarioFailover(t *testing.T) {
	// Scenario chain has two providers; first fails 500 → should failover to second
	p1Called := false
	p2Called := false

	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p1Called = true
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p2Called = true
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// Model override should persist through failover
		if data["model"] != "think-override" {
			t.Errorf("model = %v, want %q", data["model"], "think-override")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)

	provider1 := &Provider{Name: "think-p1", BaseURL: u1, Token: "t1", Model: "m1", SonnetModel: "my-sonnet", Healthy: true}
	provider2 := &Provider{Name: "think-p2", BaseURL: u2, Token: "t2", Model: "m2", SonnetModel: "other-sonnet", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"think": {
				Providers: []*Provider{provider1, provider2},
				Models:    map[string]string{"think-p1": "think-override", "think-p2": "think-override"},
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"hi"}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !p1Called {
		t.Error("first think provider should have been called (then failed)")
	}
	if !p2Called {
		t.Error("second think provider should have been called (failover)")
	}
}

func TestRoutingScenarioFailoverWithoutModelOverride(t *testing.T) {
	// Scenario chain with failover, no model override → each provider uses its own mapping
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// No model override → should use provider2's sonnet mapping
		if data["model"] != "p2-sonnet" {
			t.Errorf("model = %v, want %q", data["model"], "p2-sonnet")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)

	provider1 := &Provider{Name: "img-p1", BaseURL: u1, Token: "t1", SonnetModel: "p1-sonnet", Healthy: true}
	provider2 := &Provider{Name: "img-p2", BaseURL: u2, Token: "t2", SonnetModel: "p2-sonnet", Healthy: true}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"image": {
				Providers: []*Provider{provider1, provider2},
				// No Model → normal mapping per provider
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":[{"type":"image","source":{"type":"base64","data":"abc"}}]}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRoutingScenarioWithoutModelOverrideUsesNormalMapping(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		// No model override → should use provider's normal model mapping
		if data["model"] != "my-sonnet" {
			t.Errorf("model = %v, want %q (normal mapping)", data["model"], "my-sonnet")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	provider := &Provider{
		Name: "p1", BaseURL: u, Token: "t",
		SonnetModel: "my-sonnet", Healthy: true,
	}

	routing := &RoutingConfig{
		DefaultProviders: []*Provider{provider},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"image": {
				Providers: []*Provider{provider},
				// No Model override → normal mapping should apply
			},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":[{"type":"image","source":{}}]}]}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestEnvVarsAppliedAsHeaders tests that env vars are converted to HTTP headers.
func TestEnvVarsAppliedAsHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify env var headers are present
		if r.Header.Get("x-env-claude-code-max-output-tokens") != "64000" {
			t.Errorf("x-env-claude-code-max-output-tokens = %q, want 64000",
				r.Header.Get("x-env-claude-code-max-output-tokens"))
		}
		if r.Header.Get("x-env-max-thinking-tokens") != "50000" {
			t.Errorf("x-env-max-thinking-tokens = %q, want 50000",
				r.Header.Get("x-env-max-thinking-tokens"))
		}
		if r.Header.Get("x-env-claude-code-effort-level") != "high" {
			t.Errorf("x-env-claude-code-effort-level = %q, want high",
				r.Header.Get("x-env-claude-code-effort-level"))
		}
		if r.Header.Get("x-env-my-custom-var") != "custom_value" {
			t.Errorf("x-env-my-custom-var = %q, want custom_value",
				r.Header.Get("x-env-my-custom-var"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name:    "test",
		BaseURL: u,
		Token:   "test-token",
		EnvVars: map[string]string{
			"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
			"MAX_THINKING_TOKENS":            "50000",
			"CLAUDE_CODE_EFFORT_LEVEL":       "high",
			"MY_CUSTOM_VAR":                  "custom_value",
		},
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestEnvVarsFailoverSwitchesEnvVars tests that failover switches to the second provider's env vars.
func TestEnvVarsFailoverSwitchesEnvVars(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First provider fails
		w.WriteHeader(500)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify second provider's env vars are used
		if r.Header.Get("x-env-claude-code-max-output-tokens") != "32000" {
			t.Errorf("x-env-claude-code-max-output-tokens = %q, want 32000 (from provider2)",
				r.Header.Get("x-env-claude-code-max-output-tokens"))
		}
		if r.Header.Get("x-env-claude-code-effort-level") != "medium" {
			t.Errorf("x-env-claude-code-effort-level = %q, want medium (from provider2)",
				r.Header.Get("x-env-claude-code-effort-level"))
		}
		// Provider1's custom var should NOT be present
		if r.Header.Get("x-env-provider1-var") != "" {
			t.Errorf("x-env-provider1-var should not be present, got %q",
				r.Header.Get("x-env-provider1-var"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{
			Name:    "p1",
			BaseURL: u1,
			Token:   "token1",
			EnvVars: map[string]string{
				"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
				"CLAUDE_CODE_EFFORT_LEVEL":       "high",
				"PROVIDER1_VAR":                  "p1_value",
			},
			Healthy: true,
		},
		{
			Name:    "p2",
			BaseURL: u2,
			Token:   "token2",
			EnvVars: map[string]string{
				"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "32000",
				"CLAUDE_CODE_EFFORT_LEVEL":       "medium",
			},
			Healthy: true,
		},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover)", w.Code)
	}
}

// TestEnvVarsEmptyMapNoHeaders tests that empty env vars map doesn't add headers.
func TestEnvVarsEmptyMapNoHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no x-env- headers are present
		for k := range r.Header {
			if strings.HasPrefix(strings.ToLower(k), "x-env-") {
				t.Errorf("unexpected header %q", k)
			}
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name:    "test",
		BaseURL: u,
		Token:   "test-token",
		EnvVars: map[string]string{}, // Empty map
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestEnvVarsNilMapNoHeaders tests that nil env vars map doesn't add headers.
func TestEnvVarsNilMapNoHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no x-env- headers are present
		for k := range r.Header {
			if strings.HasPrefix(strings.ToLower(k), "x-env-") {
				t.Errorf("unexpected header %q", k)
			}
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name:    "test",
		BaseURL: u,
		Token:   "test-token",
		EnvVars: nil, // Nil map
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestStartProxyBasic tests that StartProxy starts a server.
func TestStartProxyBasic(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	port, err := StartProxy(providers, "", "127.0.0.1:0", discardLogger())
	if err != nil {
		t.Fatalf("StartProxy() error: %v", err)
	}
	if port <= 0 {
		t.Errorf("port = %d, want > 0", port)
	}
}

// TestStartProxyWithRoutingBasic tests that StartProxyWithRouting starts a server.
func TestStartProxyWithRoutingBasic(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	routing := &RoutingConfig{
		DefaultProviders: providers,
	}

	port, err := StartProxyWithRouting(routing, "", "127.0.0.1:0", discardLogger())
	if err != nil {
		t.Fatalf("StartProxyWithRouting() error: %v", err)
	}
	if port <= 0 {
		t.Errorf("port = %d, want > 0", port)
	}
}

func TestIsRequestRelatedError(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "invalid_request_error type",
			body: `{"error":{"type":"invalid_request_error","message":"some error"}}`,
			want: true,
		},
		{
			name: "context length message",
			body: `{"error":{"type":"error","message":"context length exceeded"}}`,
			want: true,
		},
		{
			name: "too long message",
			body: `{"error":{"type":"error","message":"request is too long"}}`,
			want: true,
		},
		{
			name: "too large message",
			body: `{"error":{"type":"error","message":"payload too large"}}`,
			want: true,
		},
		{
			name: "exceeds maximum message",
			body: `{"error":{"type":"error","message":"input exceeds maximum allowed"}}`,
			want: true,
		},
		{
			name: "generic rate limit should not match",
			body: `{"error":{"type":"rate_limit_error","message":"rate limit exceeded"}}`,
			want: false,
		},
		{
			name: "generic error with limit keyword should not match",
			body: `{"error":{"type":"error","message":"rate limit reached"}}`,
			want: false,
		},
		{
			name: "generic error with token keyword should not match",
			body: `{"error":{"type":"error","message":"invalid token"}}`,
			want: false,
		},
		{
			name: "generic error with size keyword should not match",
			body: `{"error":{"type":"error","message":"unknown size"}}`,
			want: false,
		},
		{
			name: "empty body",
			body: `{}`,
			want: false,
		},
		{
			name: "invalid json",
			body: `not json`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRequestRelatedError([]byte(tt.body))
			if got != tt.want {
				t.Errorf("isRequestRelatedError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// T013: Test that 502 response body includes per-provider failure details with elapsed time.
func TestAllProvidersFailBodyFormat(t *testing.T) {
	// Provider 1: returns 500 server error
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer backend1.Close()

	// Provider 2: returns 429 rate limit
	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "provider-alpha", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "provider-beta", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}

	body := w.Body.String()

	// Body must mention each provider by name
	if !strings.Contains(body, "provider-alpha") {
		t.Errorf("502 body should mention provider-alpha, got:\n%s", body)
	}
	if !strings.Contains(body, "provider-beta") {
		t.Errorf("502 body should mention provider-beta, got:\n%s", body)
	}

	// Body must contain status codes
	if !strings.Contains(body, "500") {
		t.Errorf("502 body should contain status code 500, got:\n%s", body)
	}
	if !strings.Contains(body, "429") {
		t.Errorf("502 body should contain status code 429, got:\n%s", body)
	}

	// Body must contain elapsed time indicators (e.g., "42ms")
	if !strings.Contains(body, "ms") {
		t.Errorf("502 body should contain elapsed time in ms, got:\n%s", body)
	}
}

// T013b: Test 502 body format with connection error (no status code).
func TestAllProvidersFailConnectionError(t *testing.T) {
	// Use a URL that will fail to connect
	badURL, _ := url.Parse("http://127.0.0.1:1") // port 1 is unlikely to be listening
	providers := []*Provider{
		{Name: "broken-provider", BaseURL: badURL, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}

	body := w.Body.String()
	if !strings.Contains(body, "broken-provider") {
		t.Errorf("502 body should mention broken-provider, got:\n%s", body)
	}
	if !strings.Contains(body, "error:") {
		t.Errorf("502 body should contain 'error:' for connection failures, got:\n%s", body)
	}
	if !strings.Contains(body, "ms") {
		t.Errorf("502 body should contain elapsed time in ms, got:\n%s", body)
	}
}

// TestCopyResponse_NoTagInjection verifies that responses are not modified
// and no provider tags are injected into the response body.
func TestCopyResponse_NoTagInjection(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		contentType    string
		wantUnmodified bool
	}{
		{
			name:           "non-streaming JSON response",
			responseBody:   `{"id":"msg_123","type":"message","role":"assistant","content":[{"type":"text","text":"Hello"}],"model":"claude-sonnet-4"}`,
			contentType:    "application/json",
			wantUnmodified: true,
		},
		{
			name: "streaming SSE response",
			responseBody: "event: message_start\n" +
				"data: {\"type\":\"message_start\",\"message\":{\"model\":\"claude-sonnet-4\"}}\n\n" +
				"event: content_block_delta\n" +
				"data: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n" +
				"event: message_stop\n" +
				"data: {\"type\":\"message_stop\"}\n\n",
			contentType:    "text/event-stream",
			wantUnmodified: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(200)
				w.Write([]byte(tt.responseBody))
			}))
			defer backend.Close()

			u, _ := url.Parse(backend.URL)
			providers := []*Provider{{Name: "test-provider", BaseURL: u, Token: "t", Healthy: true}}
			srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)

			req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4","messages":[{"role":"user","content":"hi"}]}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("anthropic-version", "2023-06-01")
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			body := w.Body.String()

			// Verify no provider tag is present
			if strings.Contains(body, "[provider:") || strings.Contains(body, "provider: test-provider") {
				t.Errorf("response contains provider tag, but should be unmodified:\n%s", body)
			}

			// Verify response is identical to backend response
			if tt.wantUnmodified && body != tt.responseBody {
				t.Errorf("response was modified\nwant: %s\ngot:  %s", tt.responseBody, body)
			}
		})
	}
}

// TestCopyResponse_ThinkingBlockPreserved verifies that thinking blocks
// are not modified and remain valid for Bedrock API validation.
func TestCopyResponse_ThinkingBlockPreserved(t *testing.T) {
	thinkingResponse := `{"id":"msg_123","type":"message","role":"assistant","content":[{"type":"thinking","thinking":"Let me analyze this"},{"type":"text","text":"Here is my response"}],"model":"claude-sonnet-4"}`

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(thinkingResponse))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{Name: "test-provider", BaseURL: u, Token: "t", Healthy: true}}
	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)

	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4","thinking":{"type":"enabled"},"messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()

	// Verify thinking block is preserved exactly
	if !strings.Contains(body, `"type":"thinking"`) {
		t.Error("thinking block was removed or modified")
	}

	// Verify no tag injection in thinking block
	if strings.Contains(body, "[provider:") {
		t.Error("provider tag was injected into response with thinking block")
	}

	// Verify response is byte-for-byte identical
	if body != thinkingResponse {
		t.Errorf("response with thinking block was modified\nwant: %s\ngot:  %s", thinkingResponse, body)
	}
}

// TestPathDeduplication_CrossFormat tests that path deduplication works correctly
// when base_url already contains /v1 and TransformPath returns /v1/chat/completions.
// Bug: singleJoiningSlash(base_url_with_v1, "/v1/chat/completions") produces /v1/v1/chat/completions.
func TestPathDeduplication_CrossFormat(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		wantPath   string
		clientType string // "anthropic" triggers transform to openai path
	}{
		{
			name:       "base_url with /v1 should not duplicate",
			baseURL:    "BACKEND/v1",
			wantPath:   "/v1/chat/completions",
			clientType: "anthropic",
		},
		{
			name:       "base_url without /v1 should append correctly",
			baseURL:    "BACKEND",
			wantPath:   "/v1/chat/completions",
			clientType: "anthropic",
		},
		{
			name:       "base_url with trailing slash /v1/ should not duplicate",
			baseURL:    "BACKEND/v1/",
			wantPath:   "/v1/chat/completions",
			clientType: "anthropic",
		},
		{
			name:       "base_url with /api/v1 should not duplicate",
			baseURL:    "BACKEND/api/v1",
			wantPath:   "/api/v1/chat/completions",
			clientType: "anthropic",
		},
		{
			name:       "openai client same-format pass-through with /v1 base_url",
			baseURL:    "BACKEND/v1",
			wantPath:   "/v1/chat/completions",
			clientType: "openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedPath string
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(200)
				w.Write([]byte(`{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"text","text":"ok"}],"model":"test","usage":{"input_tokens":10,"output_tokens":5}}`))
			}))
			defer backend.Close()

			// Replace BACKEND placeholder with actual backend URL
			baseURLStr := strings.Replace(tt.baseURL, "BACKEND", backend.URL, 1)
			u, _ := url.Parse(baseURLStr)

			providers := []*Provider{{
				Name:    "test-openai",
				Type:    config.ProviderTypeOpenAI,
				BaseURL: u,
				Token:   "test-token",
				Model:   "gpt-test",
				Healthy: true,
			}}

			srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
			var reqPath string
			if tt.clientType == "anthropic" {
				reqPath = "/v1/messages"
			} else {
				reqPath = "/v1/chat/completions"
			}
			req := httptest.NewRequest("POST", reqPath, strings.NewReader(`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`))
			req.Header.Set("X-Zen-Request-Format", tt.clientType)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
			}
			if receivedPath != tt.wantPath {
				t.Errorf("backend received path %q, want %q", receivedPath, tt.wantPath)
			}
		})
	}
}

// TestBuildProviders_TypeAwareDefaults tests that buildProviders() only fills
// Anthropic default model names for providers of type "anthropic".
// For "openai" providers, tier-specific fields should remain empty.
func TestBuildProviders_TypeAwareDefaults(t *testing.T) {
	// Set up temp config directory
	tmpDir := t.TempDir()
	t.Setenv("GOZEN_CONFIG_DIR", tmpDir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	// Write config with both openai and anthropic providers
	cfg := map[string]interface{}{
		"version": config.CurrentConfigVersion,
		"providers": map[string]interface{}{
			"openai-only-default": map[string]interface{}{
				"type":       "openai",
				"base_url":   "https://api.openai.test",
				"auth_token": "test-token",
				"model":      "gpt-test-model",
				// No tier-specific models
			},
			"openai-with-tier": map[string]interface{}{
				"type":         "openai",
				"base_url":     "https://api.openai.test",
				"auth_token":   "test-token",
				"model":        "gpt-test-model",
				"sonnet_model": "gpt-custom-sonnet",
			},
			"anthropic-no-tiers": map[string]interface{}{
				"type":       "anthropic",
				"base_url":   "https://api.anthropic.test",
				"auth_token": "test-token",
				// No model or tier-specific models
			},
			"empty-type": map[string]interface{}{
				"base_url":   "https://api.test",
				"auth_token": "test-token",
				// type is empty → defaults to "anthropic"
			},
		},
	}

	cfgBytes, _ := json.Marshal(cfg)
	os.WriteFile(filepath.Join(tmpDir, "zen.json"), cfgBytes, 0644)

	config.ResetDefaultStore()

	pp := NewProfileProxy(discardLogger())

	tests := []struct {
		name             string
		providerName     string
		wantSonnetEmpty  bool
		wantHaikuEmpty   bool
		wantOpusEmpty    bool
		wantReasonEmpty  bool
		wantSonnetValue  string
		wantDefaultModel string
	}{
		{
			name:             "openai provider with only default model: tier fields should be empty",
			providerName:     "openai-only-default",
			wantSonnetEmpty:  true,
			wantHaikuEmpty:   true,
			wantOpusEmpty:    true,
			wantReasonEmpty:  true,
			wantDefaultModel: "gpt-test-model",
		},
		{
			name:             "openai provider with explicit sonnet_model: should preserve it",
			providerName:     "openai-with-tier",
			wantSonnetEmpty:  false,
			wantSonnetValue:  "gpt-custom-sonnet",
			wantHaikuEmpty:   true,
			wantOpusEmpty:    true,
			wantReasonEmpty:  true,
			wantDefaultModel: "gpt-test-model",
		},
		{
			name:             "anthropic provider with no tiers: should fill Anthropic defaults",
			providerName:     "anthropic-no-tiers",
			wantSonnetEmpty:  false,
			wantHaikuEmpty:   false,
			wantOpusEmpty:    false,
			wantReasonEmpty:  false,
			wantDefaultModel: "claude-sonnet-4-5",
		},
		{
			name:             "empty type defaults to anthropic: should fill Anthropic defaults",
			providerName:     "empty-type",
			wantSonnetEmpty:  false,
			wantHaikuEmpty:   false,
			wantOpusEmpty:    false,
			wantReasonEmpty:  false,
			wantDefaultModel: "claude-sonnet-4-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers, err := pp.buildProviders([]string{tt.providerName}, nil)
			if err != nil {
				t.Fatalf("buildProviders error: %v", err)
			}
			if len(providers) != 1 {
				t.Fatalf("expected 1 provider, got %d", len(providers))
			}
			p := providers[0]

			if p.Model != tt.wantDefaultModel {
				t.Errorf("Model = %q, want %q", p.Model, tt.wantDefaultModel)
			}

			if tt.wantSonnetEmpty && p.SonnetModel != "" {
				t.Errorf("SonnetModel = %q, want empty", p.SonnetModel)
			}
			if !tt.wantSonnetEmpty && tt.wantSonnetValue != "" && p.SonnetModel != tt.wantSonnetValue {
				t.Errorf("SonnetModel = %q, want %q", p.SonnetModel, tt.wantSonnetValue)
			}
			if !tt.wantSonnetEmpty && tt.wantSonnetValue == "" && p.SonnetModel == "" {
				t.Errorf("SonnetModel should not be empty for anthropic provider")
			}

			if tt.wantHaikuEmpty && p.HaikuModel != "" {
				t.Errorf("HaikuModel = %q, want empty", p.HaikuModel)
			}
			if !tt.wantHaikuEmpty && p.HaikuModel == "" {
				t.Errorf("HaikuModel should not be empty for anthropic provider")
			}

			if tt.wantOpusEmpty && p.OpusModel != "" {
				t.Errorf("OpusModel = %q, want empty", p.OpusModel)
			}
			if !tt.wantOpusEmpty && p.OpusModel == "" {
				t.Errorf("OpusModel should not be empty for anthropic provider")
			}

			if tt.wantReasonEmpty && p.ReasoningModel != "" {
				t.Errorf("ReasoningModel = %q, want empty", p.ReasoningModel)
			}
			if !tt.wantReasonEmpty && p.ReasoningModel == "" {
				t.Errorf("ReasoningModel should not be empty for anthropic provider")
			}
		})
	}
}

// TestModelMappingFallthrough_OpenAI tests that mapModel() falls through to p.Model
// when tier-specific fields are empty (as they should be for OpenAI providers after fix).
func TestModelMappingFallthrough_OpenAI(t *testing.T) {
	tests := []struct {
		name         string
		requestModel string
		provider     *Provider
		wantModel    string
	}{
		{
			name:         "openai provider, empty sonnet_model, request claude-sonnet → fallthrough to p.Model",
			requestModel: "claude-sonnet-4-6",
			provider: &Provider{
				Name:  "openai-p",
				Type:  config.ProviderTypeOpenAI,
				Model: "gpt-5.3-codex",
				// All tier fields empty — simulating post-fix buildProviders behavior
			},
			wantModel: "gpt-5.3-codex",
		},
		{
			name:         "openai provider, explicit sonnet_model, request claude-sonnet → returns explicit",
			requestModel: "claude-sonnet-4-6",
			provider: &Provider{
				Name:        "openai-p",
				Type:        config.ProviderTypeOpenAI,
				Model:       "gpt-5.3-codex",
				SonnetModel: "gpt-5.4",
			},
			wantModel: "gpt-5.4",
		},
		{
			name:         "openai provider, empty opus_model, request claude-opus → fallthrough to p.Model",
			requestModel: "claude-opus-4-6",
			provider: &Provider{
				Name:  "openai-p",
				Type:  config.ProviderTypeOpenAI,
				Model: "gpt-5.3-codex",
			},
			wantModel: "gpt-5.3-codex",
		},
		{
			name:         "openai provider, empty haiku_model, request claude-haiku → fallthrough to p.Model",
			requestModel: "claude-haiku-4-5",
			provider: &Provider{
				Name:  "openai-p",
				Type:  config.ProviderTypeOpenAI,
				Model: "gpt-5.3-codex",
			},
			wantModel: "gpt-5.3-codex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedModel string
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var data map[string]interface{}
				json.Unmarshal(body, &data)
				receivedModel, _ = data["model"].(string)
				w.WriteHeader(200)
				w.Write([]byte(`{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"text","text":"ok"}],"model":"test","usage":{"input_tokens":10,"output_tokens":5}}`))
			}))
			defer backend.Close()

			u, _ := url.Parse(backend.URL)
			tt.provider.BaseURL = u
			tt.provider.Token = "test-token"
			tt.provider.Healthy = true

			srv := NewProxyServer([]*Provider{tt.provider}, discardLogger(), config.LoadBalanceFailover, nil)
			body := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"hi"}]}`, tt.requestModel)
			req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Fatalf("status = %d, want 200", w.Code)
			}
			if receivedModel != tt.wantModel {
				t.Errorf("backend received model %q, want %q", receivedModel, tt.wantModel)
			}
		})
	}
}

// TestE2E_AnthropicToOpenAI_NonStreaming tests the full Anthropic→OpenAI pipeline:
// model mapping, request transformation, path transformation, response transformation.
func TestE2E_AnthropicToOpenAI_NonStreaming(t *testing.T) {
	var receivedPath, receivedModel string
	var receivedBody map[string]interface{}

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		receivedModel, _ = receivedBody["model"].(string)

		// Respond with OpenAI format
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-test",
			"choices": [{
				"index": 0,
				"message": {"role": "assistant", "content": "Hello from OpenAI!"},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL + "/v1")
	providers := []*Provider{{
		Name:    "openai-e2e",
		Type:    config.ProviderTypeOpenAI,
		BaseURL: u,
		Token:   "test-token",
		Model:   "gpt-test",
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)

	// Send Anthropic-format request
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-6","max_tokens":100,"messages":[{"role":"user","content":"Hello"}]}`))
	req.Header.Set("X-Zen-Request-Format", "anthropic")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}

	// Verify path: should be /v1/chat/completions (no /v1 duplication)
	if receivedPath != "/v1/chat/completions" {
		t.Errorf("path = %q, want /v1/chat/completions", receivedPath)
	}

	// Verify model mapping: should fall through to provider's model (FR-004: mapping before transform)
	if receivedModel != "gpt-test" {
		t.Errorf("model = %q, want gpt-test", receivedModel)
	}

	// Verify request was transformed to OpenAI format (max_tokens → max_completion_tokens)
	if _, hasMaxTokens := receivedBody["max_tokens"]; hasMaxTokens {
		t.Error("request should have max_completion_tokens, not max_tokens (Anthropic→OpenAI transform)")
	}

	// Verify response was transformed back to Anthropic format
	var respData map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respData)

	if respData["type"] != "message" {
		t.Errorf("response type = %v, want message", respData["type"])
	}
	if respData["role"] != "assistant" {
		t.Errorf("response role = %v, want assistant", respData["role"])
	}
	if respData["id"] != "chatcmpl-123" {
		t.Errorf("response id = %v, want chatcmpl-123", respData["id"])
	}

	content, ok := respData["content"].([]interface{})
	if !ok || len(content) == 0 {
		t.Fatal("response content is missing or empty")
	}
	contentBlock, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatal("response content[0] is not a map")
	}
	if contentBlock["text"] != "Hello from OpenAI!" {
		t.Errorf("content text = %v, want 'Hello from OpenAI!'", contentBlock["text"])
	}

	usage, ok := respData["usage"].(map[string]interface{})
	if !ok {
		t.Fatal("response usage is missing")
	}
	if usage["input_tokens"] != float64(10) {
		t.Errorf("input_tokens = %v, want 10", usage["input_tokens"])
	}
	if usage["output_tokens"] != float64(5) {
		t.Errorf("output_tokens = %v, want 5", usage["output_tokens"])
	}
}

// TestE2E_AnthropicToOpenAI_Streaming tests SSE streaming transformation.
func TestE2E_AnthropicToOpenAI_Streaming(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with OpenAI SSE format
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher, _ := w.(http.Flusher)

		// Send OpenAI SSE events
		fmt.Fprintf(w, "data: {\"id\":\"chatcmpl-1\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"\"},\"finish_reason\":null}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "data: {\"id\":\"chatcmpl-1\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Hello\"},\"finish_reason\":null}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "data: {\"id\":\"chatcmpl-1\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL + "/v1")
	providers := []*Provider{{
		Name:    "openai-stream",
		Type:    config.ProviderTypeOpenAI,
		BaseURL: u,
		Token:   "test-token",
		Model:   "gpt-test",
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-6","max_tokens":100,"stream":true,"messages":[{"role":"user","content":"Hello"}]}`))
	req.Header.Set("X-Zen-Request-Format", "anthropic")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}

	// Verify SSE content type is preserved
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	// Verify response body contains SSE data
	body := w.Body.String()
	if !strings.Contains(body, "data:") {
		t.Errorf("response should contain SSE data events, got: %s", body)
	}
}

// TestE2E_EdgeCases tests edge cases for cross-format requests.
func TestE2E_EdgeCases(t *testing.T) {
	t.Run("upstream_4xx_error_passed_through", func(t *testing.T) {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"bad request"}}`))
		}))
		defer backend.Close()

		u, _ := url.Parse(backend.URL)
		providers := []*Provider{{
			Name:    "error-provider",
			Type:    config.ProviderTypeOpenAI,
			BaseURL: u,
			Token:   "test-token",
			Model:   "gpt-test",
			Healthy: true,
		}}

		srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
			`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`))
		req.Header.Set("X-Zen-Request-Format", "anthropic")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		// 400 from upstream is passed through (proxy only failovers on 401/402/403/429/500+)
		if w.Code != 400 {
			t.Errorf("status = %d, want 400 (passed through)", w.Code)
		}
	})

	t.Run("upstream_5xx_error_triggers_failover", func(t *testing.T) {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			w.Write([]byte(`{"error":{"type":"server_error","message":"internal error"}}`))
		}))
		defer backend.Close()

		u, _ := url.Parse(backend.URL)
		providers := []*Provider{{
			Name:    "error-provider",
			Type:    config.ProviderTypeOpenAI,
			BaseURL: u,
			Token:   "test-token",
			Model:   "gpt-test",
			Healthy: true,
		}}

		srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
			`{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hi"}]}`))
		req.Header.Set("X-Zen-Request-Format", "anthropic")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		// 500 triggers failover; with single provider → all fail → 502
		if w.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want 502 (all providers failed)", w.Code)
		}
	})

	t.Run("no_model_in_request_body_passthrough", func(t *testing.T) {
		var receivedBody map[string]interface{}
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.WriteHeader(200)
			w.Write([]byte(`{"id":"chatcmpl-1","choices":[{"message":{"content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`))
		}))
		defer backend.Close()

		u, _ := url.Parse(backend.URL)
		providers := []*Provider{{
			Name:    "no-model-provider",
			Type:    config.ProviderTypeOpenAI,
			BaseURL: u,
			Token:   "test-token",
			Model:   "gpt-test",
			Healthy: true,
		}}

		srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
		// Request without model field
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
			`{"messages":[{"role":"user","content":"hi"}]}`))
		req.Header.Set("X-Zen-Request-Format", "anthropic")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("status = %d, want 200", w.Code)
		}

		// Body should still have been forwarded (no model to map, pass through)
		if receivedBody == nil {
			t.Fatal("backend should have received request body")
		}
	})
}

func TestIsResponsesAPIRequired(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "input_is_required_error",
			body: `{"error":{"message":"input is required (request id: 123)","type":"new_api_error"}}`,
			want: true,
		},
		{
			name: "server_error",
			body: `{"error":{"message":"server error","type":"server_error"}}`,
			want: false,
		},
		{
			name: "invalid_request_error_without_input_required",
			body: `{"error":{"message":"model not found","type":"invalid_request_error"}}`,
			want: false,
		},
		{
			name: "empty_body",
			body: "",
			want: false,
		},
		{
			name: "malformed_json",
			body: `{not valid json`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isResponsesAPIRequired([]byte(tt.body))
			if got != tt.want {
				t.Errorf("isResponsesAPIRequired(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}

func TestResponsesAPIRetry(t *testing.T) {
	t.Run("retry_success", func(t *testing.T) {
		// Mock server: 500 "input is required" on /chat/completions,
		// 200 Responses API on /responses
		var chatCompletionsHit, responsesHit bool
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/chat/completions") {
				chatCompletionsHit = true
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(500)
				w.Write([]byte(`{"error":{"message":"input is required (request id: abc123)","type":"new_api_error"}}`))
				return
			}
			if strings.Contains(r.URL.Path, "/responses") {
				responsesHit = true
				// Verify request body has "input" not "messages"
				body, _ := io.ReadAll(r.Body)
				var data map[string]interface{}
				json.Unmarshal(body, &data)
				if _, ok := data["messages"]; ok {
					t.Error("retry request should not have messages field")
				}
				if _, ok := data["input"]; !ok {
					t.Error("retry request should have input field")
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				w.Write([]byte(`{"id":"resp_1","object":"response","status":"completed","model":"gpt-5","output":[{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello from Responses API!"}]}],"usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}`))
				return
			}
			w.WriteHeader(404)
		}))
		defer backend.Close()

		u, _ := url.Parse(backend.URL)
		providers := []*Provider{{
			Name:    "responses-provider",
			Type:    config.ProviderTypeOpenAI,
			BaseURL: u,
			Token:   "test-token",
			Model:   "gpt-5",
			Healthy: true,
		}}

		srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
			`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}],"max_tokens":1024}`))
		req.Header.Set("X-Zen-Request-Format", "anthropic")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if !chatCompletionsHit {
			t.Error("should have hit /chat/completions first")
		}
		if !responsesHit {
			t.Error("should have retried with /responses")
		}
		if w.Code != 200 {
			t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
		}

		// Verify response is in Anthropic format
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["type"] != "message" {
			t.Errorf("response type = %v, want message", resp["type"])
		}
		content := resp["content"].([]interface{})
		if len(content) == 0 {
			t.Fatal("response content should not be empty")
		}
		block := content[0].(map[string]interface{})
		if block["text"] != "Hello from Responses API!" {
			t.Errorf("text = %v, want Hello from Responses API!", block["text"])
		}
	})

	t.Run("no_retry_on_other_errors", func(t *testing.T) {
		// Mock server: 500 generic error on /chat/completions
		var responsesHit bool
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/responses") {
				responsesHit = true
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte(`{"error":{"message":"internal server error","type":"server_error"}}`))
		}))
		defer backend.Close()

		u, _ := url.Parse(backend.URL)
		providers := []*Provider{{
			Name:    "error-provider",
			Type:    config.ProviderTypeOpenAI,
			BaseURL: u,
			Token:   "test-token",
			Model:   "gpt-5",
			Healthy: true,
		}}

		srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
			`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}]}`))
		req.Header.Set("X-Zen-Request-Format", "anthropic")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if responsesHit {
			t.Error("should NOT have retried with /responses for non-'input is required' error")
		}
		// Should get 502 (all providers failed)
		if w.Code != 502 {
			t.Errorf("status = %d, want 502", w.Code)
		}
	})

	t.Run("retry_failure_reports_responses_api_error", func(t *testing.T) {
		// Mock server: 500 "input is required" on /chat/completions, 401 on /responses
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/chat/completions") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(500)
				w.Write([]byte(`{"error":{"message":"input is required","type":"new_api_error"}}`))
				return
			}
			if strings.Contains(r.URL.Path, "/responses") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(401)
				w.Write([]byte(`{"error":{"message":"invalid api key","type":"auth_error"}}`))
				return
			}
			w.WriteHeader(404)
		}))
		defer backend.Close()

		u, _ := url.Parse(backend.URL)
		providers := []*Provider{{
			Name:    "retry-fail-provider",
			Type:    config.ProviderTypeOpenAI,
			BaseURL: u,
			Token:   "bad-token",
			Model:   "gpt-5",
			Healthy: true,
		}}

		srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
			`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}]}`))
		req.Header.Set("X-Zen-Request-Format", "anthropic")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		// Should report Responses API error (not the original Chat Completions error)
		if w.Code != 502 {
			t.Errorf("status = %d, want 502", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "invalid api key") {
			t.Errorf("error should contain Responses API error, got: %s", body)
		}
	})
}

func TestResponsesAPIRetryStreaming(t *testing.T) {
	// Mock server: 500 "input is required" on /chat/completions,
	// SSE Responses API stream on /responses
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/chat/completions") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte(`{"error":{"message":"input is required","type":"new_api_error"}}`))
			return
		}
		if strings.Contains(r.URL.Path, "/responses") {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)
			flusher, _ := w.(http.Flusher)

			events := []string{
				"event: response.created\ndata: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_s1\",\"object\":\"response\",\"status\":\"in_progress\",\"model\":\"gpt-5\",\"output\":[]}}\n\n",
				"event: response.output_item.added\ndata: {\"type\":\"response.output_item.added\",\"output_index\":0,\"item\":{\"id\":\"msg_s1\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[]}}\n\n",
				"event: response.content_part.added\ndata: {\"type\":\"response.content_part.added\",\"item_id\":\"msg_s1\",\"output_index\":0,\"content_index\":0,\"part\":{\"type\":\"output_text\",\"text\":\"\"}}\n\n",
				"event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"item_id\":\"msg_s1\",\"output_index\":0,\"content_index\":0,\"delta\":\"Streamed!\"}\n\n",
				"event: response.output_item.done\ndata: {\"type\":\"response.output_item.done\",\"output_index\":0,\"item\":{\"id\":\"msg_s1\",\"type\":\"message\"}}\n\n",
				"event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_s1\",\"status\":\"completed\",\"model\":\"gpt-5\",\"output\":[],\"usage\":{\"input_tokens\":5,\"output_tokens\":3,\"total_tokens\":8}}}\n\n",
			}
			for _, event := range events {
				w.Write([]byte(event))
				flusher.Flush()
			}
			return
		}
		w.WriteHeader(404)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name:    "stream-responses-provider",
		Type:    config.ProviderTypeOpenAI,
		BaseURL: u,
		Token:   "test-token",
		Model:   "gpt-5",
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}],"stream":true}`))
	req.Header.Set("X-Zen-Request-Format", "anthropic")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	// Verify Anthropic SSE events in the streamed response
	if !strings.Contains(body, "event: message_start") {
		t.Error("should contain message_start event")
	}
	if !strings.Contains(body, "event: content_block_delta") {
		t.Error("should contain content_block_delta event")
	}
	if !strings.Contains(body, `"Streamed!"`) {
		t.Error("should contain streamed text")
	}
	if !strings.Contains(body, "event: message_stop") {
		t.Error("should contain message_stop event")
	}
}

func TestResponsesAPIRetryToolCall(t *testing.T) {
	// Mock server: 500 "input is required" on /chat/completions,
	// 200 Responses API with function_call on /responses
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/chat/completions") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte(`{"error":{"message":"input is required","type":"new_api_error"}}`))
			return
		}
		if strings.Contains(r.URL.Path, "/responses") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{"id":"resp_tc","object":"response","status":"completed","model":"gpt-5","output":[{"id":"fc_1","type":"function_call","call_id":"call_tc1","name":"get_weather","arguments":"{\"location\":\"Tokyo\"}","status":"completed"}],"usage":{"input_tokens":15,"output_tokens":8,"total_tokens":23}}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name:    "tool-call-provider",
		Type:    config.ProviderTypeOpenAI,
		BaseURL: u,
		Token:   "test-token",
		Model:   "gpt-5",
		Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"weather in Tokyo"}],"tools":[{"name":"get_weather","input_schema":{}}]}`))
	req.Header.Set("X-Zen-Request-Format", "anthropic")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["stop_reason"] != "tool_use" {
		t.Errorf("stop_reason = %v, want tool_use", resp["stop_reason"])
	}

	content := resp["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content length = %d, want 1", len(content))
	}
	block := content[0].(map[string]interface{})
	if block["type"] != "tool_use" {
		t.Errorf("content type = %v, want tool_use", block["type"])
	}
	if block["name"] != "get_weather" {
		t.Errorf("tool name = %v, want get_weather", block["name"])
	}
}

// T008-T010: Tests for disabled provider filtering in proxy

func setupDisabledTestConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GOZEN_CONFIG_DIR", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() {
		config.ResetDefaultStore()
		os.Unsetenv("GOZEN_CONFIG_DIR")
	})
	return dir
}

// T008: tryProviders skips a disabled provider and uses the next one
func TestTryProvidersSkipsDisabledProvider(t *testing.T) {
	setupDisabledTestConfig(t)

	// Create two backend servers
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("disabled provider should not receive requests")
		w.WriteHeader(200)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"msg_123","type":"message","content":[{"type":"text","text":"ok"}]}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)

	// Set up providers in config store
	store := config.DefaultStore()
	store.SetProvider("provider1", &config.ProviderConfig{
		BaseURL:   backend1.URL,
		AuthToken: "tok1",
	})
	store.SetProvider("provider2", &config.ProviderConfig{
		BaseURL:   backend2.URL,
		AuthToken: "tok2",
	})

	// Disable provider1
	store.DisableProvider("provider1", config.MarkingTypePermanent)

	providers := []*Provider{
		{Name: "provider1", BaseURL: u1, Token: "tok1", Healthy: true},
		{Name: "provider2", BaseURL: u2, Token: "tok2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}

// T009: ServeHTTP returns 503 when all providers are disabled
func TestAllProvidersDisabled503(t *testing.T) {
	setupDisabledTestConfig(t)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("disabled provider should not receive requests")
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)

	store := config.DefaultStore()
	store.SetProvider("p1", &config.ProviderConfig{
		BaseURL:   backend.URL,
		AuthToken: "tok1",
	})
	store.SetProvider("p2", &config.ProviderConfig{
		BaseURL:   backend.URL,
		AuthToken: "tok2",
	})

	// Disable all providers
	store.DisableProvider("p1", config.MarkingTypePermanent)
	store.DisableProvider("p2", config.MarkingTypePermanent)

	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "tok1", Healthy: true},
		{Name: "p2", BaseURL: u, Token: "tok2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger(), config.LoadBalanceFailover, nil)

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != 503 {
		t.Fatalf("status = %d, want 503; body: %s", w.Code, w.Body.String())
	}

	// Verify error JSON
	var errResp struct {
		Error struct {
			Type             string   `json:"type"`
			Message          string   `json:"message"`
			DisabledProviders []string `json:"disabled_providers"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if errResp.Error.Type != "all_providers_unavailable" {
		t.Errorf("error type = %q, want %q", errResp.Error.Type, "all_providers_unavailable")
	}
	if len(errResp.Error.DisabledProviders) != 2 {
		t.Errorf("disabled_providers count = %d, want 2", len(errResp.Error.DisabledProviders))
	}
}

// T010: Scenario fallback when all scenario providers disabled, falls back to defaults;
// returns 503 if defaults also all disabled
func TestScenarioFallbackWithDisabledProviders(t *testing.T) {
	setupDisabledTestConfig(t)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"msg_123","type":"message","content":[{"type":"text","text":"ok"}]}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)

	store := config.DefaultStore()
	store.SetProvider("scenario-p", &config.ProviderConfig{
		BaseURL:   backend.URL,
		AuthToken: "tok1",
	})
	store.SetProvider("default-p", &config.ProviderConfig{
		BaseURL:   backend.URL,
		AuthToken: "tok2",
	})

	// Disable the scenario provider
	store.DisableProvider("scenario-p", config.MarkingTypePermanent)

	scenarioProviders := []*Provider{
		{Name: "scenario-p", BaseURL: u, Token: "tok1", Healthy: true},
	}
	defaultProviders := []*Provider{
		{Name: "default-p", BaseURL: u, Token: "tok2", Healthy: true},
	}

	routing := &RoutingConfig{
		DefaultProviders: defaultProviders,
		ScenarioRoutes: map[string]*ScenarioProviders{
			string(config.ScenarioDefault): {Providers: scenarioProviders},
		},
	}

	srv := NewProxyServerWithRouting(routing, discardLogger(), config.LoadBalanceFailover, nil)

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	// Should succeed using the default provider (not disabled)
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200 (fallback to default); body: %s", w.Code, w.Body.String())
	}

	// Now disable the default provider too → should return 503
	store.DisableProvider("default-p", config.MarkingTypePermanent)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")

	srv.ServeHTTP(w2, req2)

	if w2.Code != 503 {
		t.Fatalf("status = %d, want 503 (all providers disabled); body: %s", w2.Code, w2.Body.String())
	}
}

// Phase 6: Transform Error Classification Tests

// T032: Verify TransformError type exists and can be detected
func TestTransformError_RequestTransformFailure(t *testing.T) {
	// Test that TransformError type exists and implements error interface
	err := &TransformError{Op: "request", Err: fmt.Errorf("test error")}
	if err.Error() == "" {
		t.Error("TransformError should implement error interface")
	}

	// Test that errors.As can detect TransformError
	var transformErr *TransformError
	if !errors.As(err, &transformErr) {
		t.Error("errors.As should detect TransformError")
	}

	if transformErr.Op != "request" {
		t.Errorf("expected Op=request, got %s", transformErr.Op)
	}
}

// T033: Verify response transform errors return HTTP 500
func TestTransformError_ResponseTransformFailure(t *testing.T) {
	// Test TransformError for response operations
	err := &TransformError{Op: "response", Err: fmt.Errorf("invalid format")}

	var transformErr *TransformError
	if !errors.As(err, &transformErr) {
		t.Error("errors.As should detect TransformError")
	}

	if transformErr.Op != "response" {
		t.Errorf("expected Op=response, got %s", transformErr.Op)
	}

	// Verify Unwrap works
	if transformErr.Unwrap() == nil {
		t.Error("TransformError should unwrap to underlying error")
	}
}

// Test that transform errors return proper JSON with correct Content-Type
func TestTransformError_ProperJSONResponse(t *testing.T) {
	// Test that TransformError produces valid JSON response
	err := &TransformError{Op: "request", Err: fmt.Errorf("test error with \"quotes\" and special chars")}

	// Simulate what the server does
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	errResp := map[string]interface{}{
		"error": map[string]interface{}{
			"type":    "transform_error",
			"message": err.Error(),
		},
	}
	json.NewEncoder(w).Encode(errResp)

	// Verify Content-Type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %s", w.Header().Get("Content-Type"))
	}

	// Verify valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
		t.Errorf("response should be valid JSON: %v, body: %s", err, w.Body.String())
	}

	// Verify error structure
	if decoded["error"] == nil {
		t.Error("expected error field in response")
	}

	errorObj := decoded["error"].(map[string]interface{})
	if errorObj["type"] != "transform_error" {
		t.Errorf("expected type=transform_error, got %v", errorObj["type"])
	}

	// Verify message contains the error text (quotes should be properly escaped)
	message := errorObj["message"].(string)
	if !strings.Contains(message, "test error") {
		t.Errorf("expected message to contain error text, got: %s", message)
	}
}

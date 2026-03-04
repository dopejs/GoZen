//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// MockResponse configures a single response from a MockProvider.
type MockResponse struct {
	StatusCode int
	Body       string
	Delay      time.Duration
	Headers    map[string]string
}

// MockProvider wraps httptest.Server with configurable response behavior for e2e tests.
type MockProvider struct {
	Server          *httptest.Server
	URL             string
	RequestCount    atomic.Int64
	Responses       []MockResponse
	DefaultResponse MockResponse
	mu              sync.Mutex
}

// newMockProvider creates a mock provider with a default 200 OK Anthropic response.
func newMockProvider(t *testing.T) *MockProvider {
	t.Helper()
	mp := &MockProvider{
		DefaultResponse: MockResponse{
			StatusCode: http.StatusOK,
			Body:       defaultAnthropicResponse,
		},
	}

	mp.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mp.RequestCount.Add(1)
		resp := mp.nextResponse()

		if resp.Delay > 0 {
			time.Sleep(resp.Delay)
		}

		for k, v := range resp.Headers {
			w.Header().Set(k, v)
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		w.WriteHeader(resp.StatusCode)
		w.Write([]byte(resp.Body))
	}))
	mp.URL = mp.Server.URL

	t.Cleanup(func() {
		mp.Server.Close()
	})

	return mp
}

// newMockProviderWithStatus creates a mock provider that always returns the given status code.
func newMockProviderWithStatus(t *testing.T, statusCode int) *MockProvider {
	t.Helper()
	mp := newMockProvider(t)
	mp.DefaultResponse = MockResponse{
		StatusCode: statusCode,
		Body:       mockErrorBody(statusCode),
	}
	return mp
}

func (mp *MockProvider) enqueueResponse(resp MockResponse) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.Responses = append(mp.Responses, resp)
}

func (mp *MockProvider) nextResponse() MockResponse {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if len(mp.Responses) > 0 {
		resp := mp.Responses[0]
		mp.Responses = mp.Responses[1:]
		return resp
	}
	return mp.DefaultResponse
}

// writeConfigWithProviders writes a config that includes mock provider URLs.
func (e *testEnv) writeConfigWithProviders(t *testing.T, providers map[string]interface{}, profiles map[string]interface{}) {
	t.Helper()
	cfg := map[string]interface{}{
		"version":    6,
		"web_port":   e.webPort,
		"proxy_port": e.proxyPort,
		"providers":  providers,
		"profiles":   profiles,
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(e.configDir, "zen.json"), data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

// sendProxyRequest sends a request through the proxy and returns the response.
func (e *testEnv) sendProxyRequest(t *testing.T, profile, session string) (*http.Response, error) {
	t.Helper()
	reqBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("http://127.0.0.1:%d/%s/%s/v1/messages", e.proxyPort, profile, session)
	return http.Post(url, "application/json", bytes.NewReader(body))
}

// sendProxyRequestWithBody sends a custom body through the proxy.
func (e *testEnv) sendProxyRequestWithBody(t *testing.T, profile, session string, reqBody map[string]interface{}) (*http.Response, error) {
	t.Helper()
	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("http://127.0.0.1:%d/%s/%s/v1/messages", e.proxyPort, profile, session)
	return http.Post(url, "application/json", bytes.NewReader(body))
}

// postJSON sends a POST request with JSON body to the web API.
func (e *testEnv) postJSON(t *testing.T, path string, body interface{}, result interface{}) int {
	t.Helper()
	data, _ := json.Marshal(body)
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d%s", e.webPort, path)
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	if result != nil {
		respBody, _ := io.ReadAll(resp.Body)
		json.Unmarshal(respBody, result)
	}
	return resp.StatusCode
}

// putJSON sends a PUT request with JSON body to the web API.
func (e *testEnv) putJSON(t *testing.T, path string, body interface{}, result interface{}) int {
	t.Helper()
	data, _ := json.Marshal(body)
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d%s", e.webPort, path)
	req, _ := http.NewRequest("PUT", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("PUT %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	if result != nil {
		respBody, _ := io.ReadAll(resp.Body)
		json.Unmarshal(respBody, result)
	}
	return resp.StatusCode
}

// deleteJSON sends a DELETE request to the web API.
func (e *testEnv) deleteJSON(t *testing.T, path string) int {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d%s", e.webPort, path)
	req, _ := http.NewRequest("DELETE", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

const defaultAnthropicResponse = `{
	"id": "msg_mock_001",
	"type": "message",
	"role": "assistant",
	"model": "claude-sonnet-4-20250514",
	"content": [{"type": "text", "text": "Hello from mock provider!"}],
	"stop_reason": "end_turn",
	"usage": {"input_tokens": 10, "output_tokens": 5}
}`

func mockErrorBody(statusCode int) string {
	errType := "api_error"
	msg := "Internal server error"
	switch statusCode {
	case http.StatusServiceUnavailable:
		errType = "overloaded_error"
		msg = "Server overloaded"
	case http.StatusBadGateway:
		errType = "api_error"
		msg = "Bad gateway"
	case http.StatusTooManyRequests:
		errType = "rate_limit_error"
		msg = "Rate limit exceeded"
	case http.StatusInternalServerError:
		errType = "api_error"
		msg = "Internal server error"
	}
	body, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errType,
			"message": msg,
		},
	})
	return string(body)
}

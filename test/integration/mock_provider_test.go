// Package integration contains a configurable mock provider server for integration tests.
//
//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// MockProvider wraps httptest.Server with configurable response behavior.
type MockProvider struct {
	Server          *httptest.Server
	URL             string
	RequestCount    atomic.Int64
	Responses       []MockResponse
	DefaultResponse MockResponse
	mu              sync.Mutex
}

// NewMockProvider creates a mock provider with a default 200 OK Anthropic response.
func NewMockProvider(t *testing.T) *MockProvider {
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

// NewMockProviderWithStatus creates a mock provider that always returns the given status code.
func NewMockProviderWithStatus(t *testing.T, statusCode int) *MockProvider {
	t.Helper()
	mp := NewMockProvider(t)
	mp.DefaultResponse = MockResponse{
		StatusCode: statusCode,
		Body:       errorResponseBody(statusCode),
	}
	return mp
}

// EnqueueResponse adds a response to the FIFO queue.
func (mp *MockProvider) EnqueueResponse(resp MockResponse) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.Responses = append(mp.Responses, resp)
}

// EnqueueResponses adds multiple responses to the FIFO queue.
func (mp *MockProvider) EnqueueResponses(resps ...MockResponse) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.Responses = append(mp.Responses, resps...)
}

// nextResponse dequeues the first response or returns the default.
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

// GetRequestCount returns the number of requests received.
func (mp *MockProvider) GetRequestCount() int64 {
	return mp.RequestCount.Load()
}

// ResetRequestCount resets the request counter to zero.
func (mp *MockProvider) ResetRequestCount() {
	mp.RequestCount.Store(0)
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

func errorResponseBody(statusCode int) string {
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

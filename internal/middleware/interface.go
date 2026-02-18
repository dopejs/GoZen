// Package middleware provides a pluggable middleware pipeline for GoZen.
// [BETA] This feature is experimental and disabled by default.
package middleware

import (
	"encoding/json"
	"net/http"
)

// Middleware is the interface that all middleware must implement.
// Third-party developers implement this interface to create custom middleware.
type Middleware interface {
	// Name returns the unique identifier for this middleware
	Name() string

	// Version returns the middleware version (semver)
	Version() string

	// Description returns a human-readable description
	Description() string

	// Init is called once when the middleware is loaded
	// config contains middleware-specific configuration from zen.json
	Init(config json.RawMessage) error

	// ProcessRequest is called before the request is sent to the provider
	// Return modified context, or error to abort
	ProcessRequest(ctx *RequestContext) (*RequestContext, error)

	// ProcessResponse is called after receiving the response
	// Return modified context, or error to abort
	ProcessResponse(ctx *ResponseContext) (*ResponseContext, error)

	// Priority returns the execution order (lower = earlier)
	// Built-in middleware: 0-99, User middleware: 100+
	Priority() int

	// Close is called when the middleware is unloaded
	Close() error
}

// RequestContext contains all request information passed through the middleware pipeline.
type RequestContext struct {
	// Request metadata
	SessionID   string `json:"session_id"`
	Profile     string `json:"profile"`
	Provider    string `json:"provider"`
	ClientType  string `json:"client_type"`
	ProjectPath string `json:"project_path"`

	// Request data
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`

	// Parsed body (for convenience)
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`

	// Middleware can store data here for use in ProcessResponse
	Metadata map[string]interface{} `json:"metadata"`
}

// ResponseContext contains all response information passed through the middleware pipeline.
type ResponseContext struct {
	// Original request context
	Request *RequestContext `json:"request"`

	// Response data
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"body"`

	// Parsed usage (if available)
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Message represents a conversation message.
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// NewRequestContext creates a new RequestContext with initialized maps.
func NewRequestContext() *RequestContext {
	return &RequestContext{
		Headers:  make(http.Header),
		Metadata: make(map[string]interface{}),
	}
}

// NewResponseContext creates a new ResponseContext with initialized maps.
func NewResponseContext(req *RequestContext) *ResponseContext {
	return &ResponseContext{
		Request: req,
		Headers: make(http.Header),
	}
}

// Clone creates a deep copy of the RequestContext.
func (ctx *RequestContext) Clone() *RequestContext {
	clone := &RequestContext{
		SessionID:   ctx.SessionID,
		Profile:     ctx.Profile,
		Provider:    ctx.Provider,
		ClientType:  ctx.ClientType,
		ProjectPath: ctx.ProjectPath,
		Method:      ctx.Method,
		Path:        ctx.Path,
		Model:       ctx.Model,
		Headers:     make(http.Header),
		Metadata:    make(map[string]interface{}),
	}

	// Copy body
	if ctx.Body != nil {
		clone.Body = make([]byte, len(ctx.Body))
		copy(clone.Body, ctx.Body)
	}

	// Copy headers
	for k, v := range ctx.Headers {
		clone.Headers[k] = append([]string{}, v...)
	}

	// Copy messages
	if ctx.Messages != nil {
		clone.Messages = make([]Message, len(ctx.Messages))
		copy(clone.Messages, ctx.Messages)
	}

	// Copy metadata
	for k, v := range ctx.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

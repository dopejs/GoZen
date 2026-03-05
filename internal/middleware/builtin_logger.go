package middleware

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// RequestLoggerConfig holds configuration for the request logger middleware.
type RequestLoggerConfig struct {
	LogFile     string `json:"log_file,omitempty"`     // path to log file (default: stdout)
	LogBody     bool   `json:"log_body,omitempty"`     // whether to log request/response bodies
	MaxBodySize int    `json:"max_body_size,omitempty"` // max body size to log (default: 1000)
}

// RequestLoggerMiddleware logs requests and responses.
type RequestLoggerMiddleware struct {
	config RequestLoggerConfig
	logger *log.Logger
	file   *os.File
}

// NewRequestLogger creates a new request logger middleware.
func NewRequestLogger() Middleware {
	return &RequestLoggerMiddleware{
		config: RequestLoggerConfig{
			MaxBodySize: 1000,
		},
	}
}

func (m *RequestLoggerMiddleware) Name() string {
	return "request-logger"
}

func (m *RequestLoggerMiddleware) Version() string {
	return "1.0.0"
}

func (m *RequestLoggerMiddleware) Description() string {
	return "Logs all requests and responses passing through the proxy"
}

func (m *RequestLoggerMiddleware) Priority() int {
	return 20 // Early in the chain to capture all requests
}

func (m *RequestLoggerMiddleware) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := json.Unmarshal(config, &m.config); err != nil {
			return err
		}
	}

	// Set up logger
	if m.config.LogFile != "" {
		f, err := os.OpenFile(m.config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		m.file = f
		m.logger = log.New(f, "[request-logger] ", log.LstdFlags)
	} else {
		m.logger = log.New(os.Stdout, "[request-logger] ", log.LstdFlags)
	}

	if m.config.MaxBodySize == 0 {
		m.config.MaxBodySize = 1000
	}

	return nil
}

func (m *RequestLoggerMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
	// Store start time in metadata
	ctx.Metadata["request_start"] = time.Now()

	m.logger.Printf("REQUEST session=%s profile=%s provider=%s method=%s path=%s model=%s",
		ctx.SessionID, ctx.Profile, ctx.Provider, ctx.Method, ctx.Path, ctx.Model)

	if m.config.LogBody && len(ctx.Body) > 0 {
		body := string(ctx.Body)
		if len(body) > m.config.MaxBodySize {
			body = body[:m.config.MaxBodySize] + "..."
		}
		m.logger.Printf("REQUEST BODY: %s", body)
	}

	return ctx, nil
}

func (m *RequestLoggerMiddleware) ProcessResponse(ctx *ResponseContext) (*ResponseContext, error) {
	// Calculate duration
	var duration time.Duration
	if start, ok := ctx.Request.Metadata["request_start"].(time.Time); ok {
		duration = time.Since(start)
	}

	m.logger.Printf("RESPONSE session=%s status=%d duration=%v input_tokens=%d output_tokens=%d",
		ctx.Request.SessionID, ctx.StatusCode, duration, ctx.InputTokens, ctx.OutputTokens)

	if m.config.LogBody && len(ctx.Body) > 0 {
		body := string(ctx.Body)
		if len(body) > m.config.MaxBodySize {
			body = body[:m.config.MaxBodySize] + "..."
		}
		m.logger.Printf("RESPONSE BODY: %s", body)
	}

	return ctx, nil
}

func (m *RequestLoggerMiddleware) Close() error {
	if m.file != nil {
		return m.file.Close()
	}
	return nil
}

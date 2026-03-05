package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ContextInjectionConfig holds configuration for the context injection middleware.
type ContextInjectionConfig struct {
	Files       []string `json:"files,omitempty"`        // files to inject (default: [".cursorrules", "CLAUDE.md", ".claude/CLAUDE.md"])
	MaxFileSize int      `json:"max_file_size,omitempty"` // max file size to inject (default: 10000)
	Prefix      string   `json:"prefix,omitempty"`       // prefix for injected content
	Suffix      string   `json:"suffix,omitempty"`       // suffix for injected content
}

// ContextInjectionMiddleware auto-injects project context files into requests.
type ContextInjectionMiddleware struct {
	config ContextInjectionConfig
}

// DefaultContextFiles are the default files to look for.
var DefaultContextFiles = []string{
	".cursorrules",
	"CLAUDE.md",
	".claude/CLAUDE.md",
	".github/copilot-instructions.md",
}

// NewContextInjection creates a new context injection middleware.
func NewContextInjection() Middleware {
	return &ContextInjectionMiddleware{
		config: ContextInjectionConfig{
			Files:       DefaultContextFiles,
			MaxFileSize: 10000,
			Prefix:      "\n\n[Project Context]\n",
			Suffix:      "\n[End Project Context]\n\n",
		},
	}
}

func (m *ContextInjectionMiddleware) Name() string {
	return "context-injection"
}

func (m *ContextInjectionMiddleware) Version() string {
	return "1.0.0"
}

func (m *ContextInjectionMiddleware) Description() string {
	return "Auto-injects project context files (.cursorrules, CLAUDE.md) into requests"
}

func (m *ContextInjectionMiddleware) Priority() int {
	return 10 // Very early to inject context before other processing
}

func (m *ContextInjectionMiddleware) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := json.Unmarshal(config, &m.config); err != nil {
			return err
		}
	}

	// Apply defaults
	if len(m.config.Files) == 0 {
		m.config.Files = DefaultContextFiles
	}
	if m.config.MaxFileSize == 0 {
		m.config.MaxFileSize = 10000
	}
	if m.config.Prefix == "" {
		m.config.Prefix = "\n\n[Project Context]\n"
	}
	if m.config.Suffix == "" {
		m.config.Suffix = "\n[End Project Context]\n\n"
	}

	return nil
}

func (m *ContextInjectionMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
	// Skip if no project path
	if ctx.ProjectPath == "" {
		return ctx, nil
	}

	// Find and read context files
	var contextParts []string
	for _, filename := range m.config.Files {
		fullPath := filepath.Join(ctx.ProjectPath, filename)
		content, err := m.readContextFile(fullPath)
		if err != nil {
			continue // File doesn't exist or can't be read
		}
		if content != "" {
			contextParts = append(contextParts, fmt.Sprintf("--- %s ---\n%s", filename, content))
		}
	}

	// If no context files found, return unchanged
	if len(contextParts) == 0 {
		return ctx, nil
	}

	// Build context string
	contextStr := m.config.Prefix + strings.Join(contextParts, "\n\n") + m.config.Suffix

	// Inject into the first user message or system message
	ctx = m.injectContext(ctx, contextStr)

	return ctx, nil
}

func (m *ContextInjectionMiddleware) ProcessResponse(ctx *ResponseContext) (*ResponseContext, error) {
	// No processing needed for responses
	return ctx, nil
}

func (m *ContextInjectionMiddleware) Close() error {
	return nil
}

// readContextFile reads a context file if it exists and is within size limits.
func (m *ContextInjectionMiddleware) readContextFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.Size() > int64(m.config.MaxFileSize) {
		return "", fmt.Errorf("file too large: %d > %d", info.Size(), m.config.MaxFileSize)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// injectContext injects the context string into the request messages.
func (m *ContextInjectionMiddleware) injectContext(ctx *RequestContext, contextStr string) *RequestContext {
	if len(ctx.Messages) == 0 {
		return ctx
	}

	// Clone the context to avoid modifying the original
	newCtx := ctx.Clone()

	// Find the first user message and prepend context
	for i, msg := range newCtx.Messages {
		if msg.Role == "user" {
			switch content := msg.Content.(type) {
			case string:
				newCtx.Messages[i].Content = contextStr + content
			case []interface{}:
				// For array content (multimodal), prepend as text block
				textBlock := map[string]interface{}{
					"type": "text",
					"text": contextStr,
				}
				newCtx.Messages[i].Content = append([]interface{}{textBlock}, content...)
			}
			break
		}
	}

	// Re-marshal the body with updated messages
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(newCtx.Body, &bodyMap); err == nil {
		bodyMap["messages"] = newCtx.Messages
		if newBody, err := json.Marshal(bodyMap); err == nil {
			newCtx.Body = newBody
		}
	}

	return newCtx
}

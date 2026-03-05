package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ChatMessage represents a single message in conversation history.
// Named ChatMessage to avoid collision with existing Message in protocol.go.
type ChatMessage struct {
	Role      string    `json:"role"`    // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ConversationBuffer is a bounded ring buffer for conversation history.
type ConversationBuffer struct {
	mu      sync.RWMutex
	msgs    []ChatMessage
	maxSize int
}

// NewConversationBuffer creates a new conversation buffer with the given max size.
func NewConversationBuffer(maxSize int) *ConversationBuffer {
	if maxSize <= 0 {
		maxSize = 20
	}
	return &ConversationBuffer{
		msgs:    make([]ChatMessage, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a message to the buffer, evicting the oldest if full.
func (b *ConversationBuffer) Add(msg ChatMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.msgs) >= b.maxSize {
		b.msgs = b.msgs[1:]
	}
	b.msgs = append(b.msgs, msg)
}

// Messages returns all messages in chronological order.
func (b *ConversationBuffer) Messages() []ChatMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]ChatMessage, len(b.msgs))
	copy(out, b.msgs)
	return out
}

// Clear removes all messages from the buffer.
func (b *ConversationBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.msgs = b.msgs[:0]
}

// Len returns the number of messages in the buffer.
func (b *ConversationBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.msgs)
}

// MemoryDir returns the default bots memory directory path.
// Respects GOZEN_CONFIG_DIR for dev environments.
func MemoryDir() string {
	base := os.Getenv("GOZEN_CONFIG_DIR")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".zen")
	}
	return filepath.Join(base, "bots")
}

// MemoryFilePath returns the path to memory.md in the given directory.
func MemoryFilePath(dir string) string {
	return filepath.Join(dir, "memory.md")
}

// LoadMemory reads the memory.md file from the given directory.
// Returns empty string if the file doesn't exist.
func LoadMemory(dir string) (string, error) {
	path := MemoryFilePath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read memory: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveMemory writes content to memory.md in the given directory.
// Creates the directory if it doesn't exist.
func SaveMemory(dir, content string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}
	path := MemoryFilePath(dir)
	return os.WriteFile(path, []byte(content+"\n"), 0644)
}

// ClearMemory removes the memory.md file from the given directory.
func ClearMemory(dir string) error {
	path := MemoryFilePath(dir)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear memory: %w", err)
	}
	return nil
}

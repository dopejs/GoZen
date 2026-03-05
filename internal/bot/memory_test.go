package bot

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConversationBuffer_Basic(t *testing.T) {
	buf := NewConversationBuffer(3)

	if buf.Len() != 0 {
		t.Errorf("expected len 0, got %d", buf.Len())
	}

	buf.Add(ChatMessage{Role: "user", Content: "hello", Timestamp: time.Now()})
	buf.Add(ChatMessage{Role: "assistant", Content: "hi", Timestamp: time.Now()})

	if buf.Len() != 2 {
		t.Errorf("expected len 2, got %d", buf.Len())
	}

	msgs := buf.Messages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "hello" {
		t.Errorf("unexpected first message: %+v", msgs[0])
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "hi" {
		t.Errorf("unexpected second message: %+v", msgs[1])
	}
}

func TestConversationBuffer_Eviction(t *testing.T) {
	buf := NewConversationBuffer(2)

	buf.Add(ChatMessage{Role: "user", Content: "msg1"})
	buf.Add(ChatMessage{Role: "assistant", Content: "msg2"})
	buf.Add(ChatMessage{Role: "user", Content: "msg3"})

	if buf.Len() != 2 {
		t.Errorf("expected len 2 after eviction, got %d", buf.Len())
	}

	msgs := buf.Messages()
	if msgs[0].Content != "msg2" {
		t.Errorf("expected oldest to be msg2, got %s", msgs[0].Content)
	}
	if msgs[1].Content != "msg3" {
		t.Errorf("expected newest to be msg3, got %s", msgs[1].Content)
	}
}

func TestConversationBuffer_Clear(t *testing.T) {
	buf := NewConversationBuffer(10)
	buf.Add(ChatMessage{Role: "user", Content: "hello"})
	buf.Add(ChatMessage{Role: "assistant", Content: "hi"})

	buf.Clear()

	if buf.Len() != 0 {
		t.Errorf("expected len 0 after clear, got %d", buf.Len())
	}
	if len(buf.Messages()) != 0 {
		t.Error("expected empty messages after clear")
	}
}

func TestConversationBuffer_DefaultSize(t *testing.T) {
	buf := NewConversationBuffer(0)
	// Should default to 20
	for i := 0; i < 25; i++ {
		buf.Add(ChatMessage{Role: "user", Content: "msg"})
	}
	if buf.Len() != 20 {
		t.Errorf("expected max 20 with default size, got %d", buf.Len())
	}
}

func TestConversationBuffer_MessagesIsCopy(t *testing.T) {
	buf := NewConversationBuffer(10)
	buf.Add(ChatMessage{Role: "user", Content: "hello"})

	msgs := buf.Messages()
	msgs[0].Content = "modified"

	// Original should be unchanged
	original := buf.Messages()
	if original[0].Content != "hello" {
		t.Error("Messages() should return a copy")
	}
}

func TestLoadSaveClearMemory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bots")

	// Load from non-existent dir returns empty
	content, err := LoadMemory(dir)
	if err != nil {
		t.Fatalf("LoadMemory: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty, got %q", content)
	}

	// Save creates dir and file
	err = SaveMemory(dir, "你是一个猫娘助手")
	if err != nil {
		t.Fatalf("SaveMemory: %v", err)
	}

	content, err = LoadMemory(dir)
	if err != nil {
		t.Fatalf("LoadMemory: %v", err)
	}
	if content != "你是一个猫娘助手" {
		t.Errorf("expected persona content, got %q", content)
	}

	// Clear removes file
	err = ClearMemory(dir)
	if err != nil {
		t.Fatalf("ClearMemory: %v", err)
	}

	content, err = LoadMemory(dir)
	if err != nil {
		t.Fatalf("LoadMemory after clear: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty after clear, got %q", content)
	}
}

func TestClearMemory_NonExistent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent")
	// Should not error when file doesn't exist
	err := ClearMemory(dir)
	if err != nil {
		t.Errorf("ClearMemory on non-existent should not error: %v", err)
	}
}

func TestMemoryFilePath(t *testing.T) {
	path := MemoryFilePath("/home/user/.zen/bots")
	expected := filepath.Join("/home/user/.zen/bots", "memory.md")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestSaveMemory_Overwrite(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bots")

	SaveMemory(dir, "first persona")
	SaveMemory(dir, "second persona")

	content, _ := LoadMemory(dir)
	if content != "second persona" {
		t.Errorf("expected overwritten content, got %q", content)
	}
}

func TestMemoryDir(t *testing.T) {
	dir := MemoryDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".zen", "bots")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

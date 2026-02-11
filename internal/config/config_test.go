package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	os.MkdirAll(filepath.Join(dir, ".cc_envs"), 0755)
	return dir
}

func TestReadWriteFallbackOrder(t *testing.T) {
	setTestHome(t)

	names := []string{"yunyi", "cctq", "minimax"}
	if err := WriteFallbackOrder(names); err != nil {
		t.Fatalf("WriteFallbackOrder() error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}

	if len(got) != len(names) {
		t.Fatalf("got %d names, want %d", len(got), len(names))
	}
	for i, n := range names {
		if got[i] != n {
			t.Errorf("got[%d] = %q, want %q", i, got[i], n)
		}
	}
}

func TestReadFallbackOrderSkipsCommentsAndBlanks(t *testing.T) {
	home := setTestHome(t)
	confPath := filepath.Join(home, ".cc_envs", "fallback.conf")

	content := "# comment\nyunyi\n\ncctq\n# another\nminimax\n"
	os.WriteFile(confPath, []byte(content), 0644)

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 names, got %d: %v", len(got), got)
	}
	if got[0] != "yunyi" || got[1] != "cctq" || got[2] != "minimax" {
		t.Errorf("got %v", got)
	}
}

func TestReadFallbackOrderMissing(t *testing.T) {
	setTestHome(t)
	// Don't create fallback.conf

	_, err := ReadFallbackOrder()
	if err == nil {
		t.Error("expected error for missing fallback.conf")
	}
}

func TestWriteFallbackOrderEmpty(t *testing.T) {
	setTestHome(t)

	if err := WriteFallbackOrder(nil); err != nil {
		t.Fatalf("WriteFallbackOrder(nil) error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 names, got %d", len(got))
	}
}

func TestWriteFallbackOrderCreatesDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Don't pre-create .cc_envs

	if err := WriteFallbackOrder([]string{"a"}); err != nil {
		t.Fatalf("WriteFallbackOrder() error: %v", err)
	}

	got, err := ReadFallbackOrder()
	if err != nil {
		t.Fatalf("ReadFallbackOrder() error: %v", err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("got %v, want [a]", got)
	}
}

func TestFallbackConfPath(t *testing.T) {
	home := setTestHome(t)
	got := FallbackConfPath()
	want := filepath.Join(home, ".cc_envs", "fallback.conf")
	if got != want {
		t.Errorf("FallbackConfPath() = %q, want %q", got, want)
	}
}

func TestWriteFallbackOrderErrorBadDir(t *testing.T) {
	t.Setenv("HOME", "/dev/null/impossible")

	err := WriteFallbackOrder([]string{"a"})
	if err == nil {
		t.Error("expected error when dir can't be created")
	}
}

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

func TestRemoveFromFallbackOrder(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("b"); err != nil {
		t.Fatalf("RemoveFromFallbackOrder() error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("got %v, want [a c]", got)
	}
}

func TestRemoveFromFallbackOrderMissingFile(t *testing.T) {
	setTestHome(t)
	// No fallback.conf — should be a no-op
	if err := RemoveFromFallbackOrder("x"); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRemoveFromFallbackOrderNotPresent(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b"})

	if err := RemoveFromFallbackOrder("z"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 {
		t.Errorf("expected 2 names unchanged, got %v", got)
	}
}

func TestRemoveFromFallbackOrderFirst(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("a"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Errorf("got %v, want [b c]", got)
	}
}

func TestRemoveFromFallbackOrderLast(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "c"})

	if err := RemoveFromFallbackOrder("c"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

func TestRemoveFromFallbackOrderOnlyItem(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"solo"})

	if err := RemoveFromFallbackOrder("solo"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestRemoveFromFallbackOrderDuplicates(t *testing.T) {
	setTestHome(t)
	WriteFallbackOrder([]string{"a", "b", "a", "c"})

	if err := RemoveFromFallbackOrder("a"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadFallbackOrder()
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Errorf("got %v, want [b c]", got)
	}
}

// --- Profile tests ---

func TestProfileConfPathDefault(t *testing.T) {
	home := setTestHome(t)
	got := ProfileConfPath("default")
	want := filepath.Join(home, ".cc_envs", "fallback.conf")
	if got != want {
		t.Errorf("ProfileConfPath(\"default\") = %q, want %q", got, want)
	}
}

func TestProfileConfPathEmpty(t *testing.T) {
	home := setTestHome(t)
	got := ProfileConfPath("")
	want := filepath.Join(home, ".cc_envs", "fallback.conf")
	if got != want {
		t.Errorf("ProfileConfPath(\"\") = %q, want %q", got, want)
	}
}

func TestProfileConfPathNamed(t *testing.T) {
	home := setTestHome(t)
	got := ProfileConfPath("work")
	want := filepath.Join(home, ".cc_envs", "fallback.work.conf")
	if got != want {
		t.Errorf("ProfileConfPath(\"work\") = %q, want %q", got, want)
	}
}

func TestReadWriteProfileOrder(t *testing.T) {
	setTestHome(t)

	names := []string{"p1", "p2"}
	if err := WriteProfileOrder("work", names); err != nil {
		t.Fatalf("WriteProfileOrder() error: %v", err)
	}

	got, err := ReadProfileOrder("work")
	if err != nil {
		t.Fatalf("ReadProfileOrder() error: %v", err)
	}
	if len(got) != 2 || got[0] != "p1" || got[1] != "p2" {
		t.Errorf("got %v, want [p1 p2]", got)
	}

	// Default profile should be unaffected
	_, err = ReadProfileOrder("default")
	if err == nil {
		t.Error("expected error for missing default profile")
	}
}

func TestListProfiles(t *testing.T) {
	setTestHome(t)

	WriteProfileOrder("default", []string{"a"})
	WriteProfileOrder("work", []string{"b"})
	WriteProfileOrder("staging", []string{"c"})

	profiles := ListProfiles()
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d: %v", len(profiles), profiles)
	}
	// Should be sorted
	if profiles[0] != "default" || profiles[1] != "staging" || profiles[2] != "work" {
		t.Errorf("got %v, want [default staging work]", profiles)
	}
}

func TestListProfilesEmpty(t *testing.T) {
	setTestHome(t)
	profiles := ListProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %v", profiles)
	}
}

func TestDeleteProfile(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("work", []string{"a"})

	if err := DeleteProfile("work"); err != nil {
		t.Fatalf("DeleteProfile() error: %v", err)
	}

	_, err := ReadProfileOrder("work")
	if err == nil {
		t.Error("expected error after deleting profile")
	}
}

func TestDeleteProfileDefault(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("default", []string{"a"})

	err := DeleteProfile("default")
	if err == nil {
		t.Error("expected error when deleting default profile")
	}

	// Default should still exist
	got, err := ReadProfileOrder("default")
	if err != nil {
		t.Fatalf("default profile should still exist: %v", err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Errorf("got %v, want [a]", got)
	}
}

func TestDeleteProfileEmpty(t *testing.T) {
	setTestHome(t)
	err := DeleteProfile("")
	if err == nil {
		t.Error("expected error when deleting empty profile name")
	}
}

func TestRemoveFromProfileOrder(t *testing.T) {
	setTestHome(t)
	WriteProfileOrder("work", []string{"a", "b", "c"})

	if err := RemoveFromProfileOrder("work", "b"); err != nil {
		t.Fatalf("error: %v", err)
	}

	got, _ := ReadProfileOrder("work")
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("got %v, want [a c]", got)
	}
}

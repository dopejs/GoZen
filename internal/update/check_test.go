package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"2.0.0", "2.0.0", 0},
		{"2.0.0", "2.0.1", -1},
		{"2.1.0", "2.0.1", 1},
		{"2.0.0", "3.0.0", -1},
		{"10.0.0", "9.0.0", 1},
	}

	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFormatNotificationUpToDate(t *testing.T) {
	c := NewChecker("2.0.0")
	msg := c.formatNotification("2.0.0")
	if msg != "" {
		t.Errorf("expected empty notification when up-to-date, got %q", msg)
	}
}

func TestFormatNotificationNewerAvailable(t *testing.T) {
	c := NewChecker("2.0.0")
	msg := c.formatNotification("2.1.0")
	if msg == "" {
		t.Error("expected non-empty notification when update available")
	}
	if !contains(msg, "2.0.0") || !contains(msg, "2.1.0") {
		t.Errorf("notification should contain both versions: %q", msg)
	}
	if !contains(msg, "zen upgrade") {
		t.Errorf("notification should mention zen upgrade: %q", msg)
	}
}

func TestFormatNotificationEmpty(t *testing.T) {
	c := NewChecker("2.0.0")
	msg := c.formatNotification("")
	if msg != "" {
		t.Errorf("expected empty notification for empty version, got %q", msg)
	}
}

func TestCacheReadWrite(t *testing.T) {
	home := setTestHome(t)
	cachePath := filepath.Join(home, ".zen", cacheFile)

	c := &cache{
		LastCheck:     time.Now(),
		LatestVersion: "2.1.0",
	}
	writeCache(cachePath, c)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("cache file should exist: %v", err)
	}

	var loaded cache
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse cache: %v", err)
	}

	if loaded.LatestVersion != "2.1.0" {
		t.Errorf("cached version = %q, want %q", loaded.LatestVersion, "2.1.0")
	}
}

func TestFreshCacheSkipsHTTP(t *testing.T) {
	home := setTestHome(t)
	cachePath := filepath.Join(home, ".zen", cacheFile)

	// Write a fresh cache entry
	c := &cache{
		LastCheck:     time.Now(),
		LatestVersion: "2.1.0",
	}
	writeCache(cachePath, c)

	// Checker should use cache and not hit network
	checker := NewChecker("2.0.0")
	msg := checker.check()
	if !contains(msg, "2.1.0") {
		t.Errorf("expected notification from cache, got %q", msg)
	}
}

func TestNoNotificationWhenUpToDate(t *testing.T) {
	home := setTestHome(t)
	cachePath := filepath.Join(home, ".zen", cacheFile)

	c := &cache{
		LastCheck:     time.Now(),
		LatestVersion: "2.0.0",
	}
	writeCache(cachePath, c)

	checker := NewChecker("2.0.0")
	msg := checker.check()
	if msg != "" {
		t.Errorf("expected no notification when up-to-date, got %q", msg)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

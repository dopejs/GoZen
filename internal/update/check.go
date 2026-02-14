package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

const (
	checkInterval = 24 * time.Hour
	httpTimeout   = 3 * time.Second
	cacheFile     = "update-check.json"
	releaseAPI    = "https://api.github.com/repos/dopejs/gozen/releases/latest"
)

// cache stores the last check result on disk.
type cache struct {
	LastCheck     time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
}

// Checker performs a non-blocking version update check in the background.
type Checker struct {
	currentVersion string
	mu             sync.Mutex
	result         string // notification message, empty if up-to-date
	done           chan struct{}
}

// NewChecker creates a new update checker for the given current version.
func NewChecker(currentVersion string) *Checker {
	return &Checker{
		currentVersion: currentVersion,
		done:           make(chan struct{}),
	}
}

// Start launches the background check goroutine.
func (c *Checker) Start() {
	go func() {
		defer close(c.done)
		msg := c.check()
		c.mu.Lock()
		c.result = msg
		c.mu.Unlock()
	}()
}

// Notification blocks until the check is done, then returns the notification
// message. Returns empty string if up-to-date or if the check failed silently.
func (c *Checker) Notification() string {
	<-c.done
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.result
}

func (c *Checker) check() string {
	cachePath := filepath.Join(config.ConfigDirPath(), cacheFile)

	// Try to read cached result
	var cached cache
	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &cached)
	}

	latest := cached.LatestVersion

	// If cache is fresh, use it
	if !cached.LastCheck.IsZero() && time.Since(cached.LastCheck) < checkInterval && latest != "" {
		return c.formatNotification(latest)
	}

	// Fetch latest version from GitHub
	if fetched, err := fetchLatestVersion(); err == nil {
		latest = fetched
		// Update cache
		cached.LastCheck = time.Now()
		cached.LatestVersion = latest
		writeCache(cachePath, &cached)
	} else if latest == "" {
		// No cached version and fetch failed — nothing to show
		return ""
	}
	// If fetch failed but we have a stale cache, still show the notification

	return c.formatNotification(latest)
}

func (c *Checker) formatNotification(latest string) string {
	if latest == "" || compareVersions(c.currentVersion, latest) >= 0 {
		return ""
	}

	inner := fmt.Sprintf("  Update available: %s → %s", c.currentVersion, latest)
	hint := "  Run `zen upgrade` to update"

	// Pad to same width
	width := len(inner)
	if len(hint) > width {
		width = len(hint)
	}
	inner = padRight(inner, width)
	hint = padRight(hint, width)

	top := "╭" + strings.Repeat("─", width+2) + "╮"
	mid1 := "│ " + inner + " │"
	mid2 := "│ " + hint + " │"
	bot := "╰" + strings.Repeat("─", width+2) + "╯"

	return fmt.Sprintf("\n%s\n%s\n%s\n%s\n", top, mid1, mid2, bot)
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(releaseAPI)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

func writeCache(path string, c *cache) {
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, data, 0600)
}

// compareVersions returns -1, 0, or 1.
func compareVersions(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")

	maxLen := len(ap)
	if len(bp) > maxLen {
		maxLen = len(bp)
	}

	for i := 0; i < maxLen; i++ {
		var ai, bi int
		if i < len(ap) {
			ai, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bi, _ = strconv.Atoi(bp[i])
		}
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
	}
	return 0
}

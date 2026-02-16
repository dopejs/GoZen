package daemon

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// ConfigWatcher watches the config file for changes and triggers a reload callback.
// Uses polling instead of fsnotify to avoid the external dependency.
type ConfigWatcher struct {
	logger   *log.Logger
	onReload func()
	stop     chan struct{}
	path     string
	modTime  time.Time
}

// NewConfigWatcher creates a new config file watcher.
func NewConfigWatcher(logger *log.Logger, onReload func()) *ConfigWatcher {
	path := filepath.Join(config.ConfigDirPath(), config.ConfigFile)
	var modTime time.Time
	if info, err := os.Stat(path); err == nil {
		modTime = info.ModTime()
	}
	return &ConfigWatcher{
		logger:   logger,
		onReload: onReload,
		stop:     make(chan struct{}),
		path:     path,
		modTime:  modTime,
	}
}

// Start begins watching the config file. Blocks until Stop is called.
func (w *ConfigWatcher) Start() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			w.check()
		}
	}
}

// Stop stops the config watcher.
func (w *ConfigWatcher) Stop() {
	select {
	case <-w.stop:
		// Already stopped
	default:
		close(w.stop)
	}
}

// check polls the config file for modifications.
func (w *ConfigWatcher) check() {
	info, err := os.Stat(w.path)
	if err != nil {
		return
	}

	if info.ModTime().After(w.modTime) {
		w.modTime = info.ModTime()
		w.logger.Printf("config file modified, triggering reload")
		w.onReload()
	}
}

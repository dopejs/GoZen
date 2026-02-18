package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"runtime"
)

// PluginLoader handles loading middleware plugins from various sources.
type PluginLoader struct {
	pluginDir string // directory for downloaded/cached plugins
}

// NewPluginLoader creates a new plugin loader.
func NewPluginLoader(pluginDir string) *PluginLoader {
	if pluginDir == "" {
		home, _ := os.UserHomeDir()
		pluginDir = filepath.Join(home, ".zen", "plugins")
	}
	os.MkdirAll(pluginDir, 0755)
	return &PluginLoader{pluginDir: pluginDir}
}

// LoadLocal loads a middleware from a local .so plugin file.
// The plugin must export a variable named "Middleware" that implements the Middleware interface.
func (l *PluginLoader) LoadLocal(path string) (Middleware, error) {
	// Check if plugin loading is supported on this platform
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("plugin loading is not supported on Windows")
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("plugin file not found: %w", err)
	}

	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look up the Middleware symbol
	sym, err := p.Lookup("Middleware")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export 'Middleware' symbol: %w", err)
	}

	// Type assert to Middleware interface
	m, ok := sym.(Middleware)
	if !ok {
		// Try pointer to Middleware
		mp, ok := sym.(*Middleware)
		if !ok {
			return nil, fmt.Errorf("exported 'Middleware' does not implement Middleware interface")
		}
		m = *mp
	}

	return m, nil
}

// RemotePluginManifest describes a remote plugin.
type RemotePluginManifest struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Downloads   map[string]string `json:"downloads"` // platform -> URL (e.g., "linux-amd64" -> "https://...")
	Checksums   map[string]string `json:"checksums"` // platform -> SHA256
}

// LoadRemote downloads and loads a middleware from a remote URL.
// The URL should point to a JSON manifest file describing the plugin.
func (l *PluginLoader) LoadRemote(manifestURL string) (Middleware, error) {
	// Check if plugin loading is supported on this platform
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("plugin loading is not supported on Windows")
	}

	// Fetch manifest
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest fetch failed with status %d", resp.StatusCode)
	}

	var manifest RemotePluginManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Determine platform
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	downloadURL, ok := manifest.Downloads[platform]
	if !ok {
		return nil, fmt.Errorf("no download available for platform %s", platform)
	}

	expectedChecksum := manifest.Checksums[platform]

	// Check if already cached
	pluginPath := filepath.Join(l.pluginDir, fmt.Sprintf("%s-%s.so", manifest.Name, manifest.Version))
	if _, err := os.Stat(pluginPath); err == nil {
		// Verify checksum of cached file
		if l.verifyChecksum(pluginPath, expectedChecksum) {
			return l.LoadLocal(pluginPath)
		}
		// Checksum mismatch, re-download - ignore remove error
		_ = os.Remove(pluginPath)
	}

	// Download plugin
	if err := l.downloadPlugin(downloadURL, pluginPath, expectedChecksum); err != nil {
		return nil, fmt.Errorf("failed to download plugin: %w", err)
	}

	return l.LoadLocal(pluginPath)
}

// downloadPlugin downloads a plugin file and verifies its checksum.
func (l *PluginLoader) downloadPlugin(url, destPath, expectedChecksum string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp(l.pluginDir, "plugin-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Download and compute checksum
	hasher := sha256.New()
	writer := io.MultiWriter(tmpFile, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Verify checksum
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if expectedChecksum != "" && actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	// Move to final location
	if err := os.Rename(tmpPath, destPath); err != nil {
		return err
	}

	// Make executable
	return os.Chmod(destPath, 0755)
}

// verifyChecksum verifies the SHA256 checksum of a file.
func (l *PluginLoader) verifyChecksum(path, expected string) bool {
	if expected == "" {
		return true
	}

	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return false
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	return actual == expected
}

// ListCachedPlugins returns a list of cached plugin files.
func (l *PluginLoader) ListCachedPlugins() ([]string, error) {
	entries, err := os.ReadDir(l.pluginDir)
	if err != nil {
		return nil, err
	}

	var plugins []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".so" {
			plugins = append(plugins, filepath.Join(l.pluginDir, entry.Name()))
		}
	}
	return plugins, nil
}

// ClearCache removes all cached plugins.
func (l *PluginLoader) ClearCache() error {
	entries, err := os.ReadDir(l.pluginDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".so" {
			// Best-effort removal - ignore errors
			_ = os.Remove(filepath.Join(l.pluginDir, entry.Name()))
		}
	}
	return nil
}

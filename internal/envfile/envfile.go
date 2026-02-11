package envfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const EnvsDir = ".cc_envs"

// EnvsPath returns the full path to ~/.cc_envs
func EnvsPath() string {
	return filepath.Join(os.Getenv("HOME"), EnvsDir)
}

// Entry represents a key=value pair in an .env file.
type Entry struct {
	Key   string
	Value string
}

// Config represents a parsed .env file.
type Config struct {
	Name    string
	Path    string
	Entries []Entry
}

// Get returns the value for a key, or empty string.
func (c *Config) Get(key string) string {
	for _, e := range c.Entries {
		if e.Key == key {
			return e.Value
		}
	}
	return ""
}

// Set sets a key to a value. If the key exists, it updates; otherwise appends.
func (c *Config) Set(key, value string) {
	for i, e := range c.Entries {
		if e.Key == key {
			c.Entries[i].Value = value
			return
		}
	}
	c.Entries = append(c.Entries, Entry{Key: key, Value: value})
}

// Load parses an .env file. Lines starting with # or empty lines are skipped.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	name := strings.TrimSuffix(filepath.Base(path), ".env")
	cfg := &Config{Name: name, Path: path}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip inline comments
		if idx := strings.Index(line, " #"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		cfg.Entries = append(cfg.Entries, Entry{Key: k, Value: v})
	}
	return cfg, scanner.Err()
}

// Save writes the config back to its .env file.
func (c *Config) Save() error {
	var b strings.Builder
	for _, e := range c.Entries {
		fmt.Fprintf(&b, "%s=%s\n", e.Key, e.Value)
	}
	return os.WriteFile(c.Path, []byte(b.String()), 0644)
}

// Delete removes the .env file.
func (c *Config) Delete() error {
	return os.Remove(c.Path)
}

// ListConfigs returns all .env configs in ~/.cc_envs/.
func ListConfigs() ([]*Config, error) {
	dir := EnvsPath()
	matches, err := filepath.Glob(filepath.Join(dir, "*.env"))
	if err != nil {
		return nil, err
	}
	var configs []*Config
	for _, m := range matches {
		cfg, err := Load(m)
		if err != nil {
			continue
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

// ConfigNames returns just the names of available configs.
func ConfigNames() []string {
	dir := EnvsPath()
	matches, _ := filepath.Glob(filepath.Join(dir, "*.env"))
	var names []string
	for _, m := range matches {
		names = append(names, strings.TrimSuffix(filepath.Base(m), ".env"))
	}
	return names
}

// LoadByName loads a config by name from ~/.cc_envs/<name>.env.
func LoadByName(name string) (*Config, error) {
	path := filepath.Join(EnvsPath(), name+".env")
	return Load(path)
}

// Create creates a new config file.
func Create(name string, entries []Entry) (*Config, error) {
	dir := EnvsPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, name+".env")
	cfg := &Config{Name: name, Path: path, Entries: entries}
	if err := cfg.Save(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ExportToEnv sets all entries as environment variables.
func (c *Config) ExportToEnv() {
	for _, e := range c.Entries {
		os.Setenv(e.Key, e.Value)
	}
}

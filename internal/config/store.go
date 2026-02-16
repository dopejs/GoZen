package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// --- Path helpers ---

// ConfigDirPath returns ~/.zen
func ConfigDirPath() string {
	return filepath.Join(os.Getenv("HOME"), ConfigDir)
}

// ConfigFilePath returns ~/.zen/zen.json
func ConfigFilePath() string {
	return filepath.Join(ConfigDirPath(), ConfigFile)
}

// LogPath returns ~/.zen/proxy.log
func LogPath() string {
	return filepath.Join(ConfigDirPath(), "proxy.log")
}

// legacyDirPath returns ~/.cc_envs
func legacyDirPath() string {
	return filepath.Join(os.Getenv("HOME"), LegacyDir)
}

// legacyOpenCCDirPath returns ~/.opencc
func legacyOpenCCDirPath() string {
	return filepath.Join(os.Getenv("HOME"), LegacyOpenCCDir)
}

// legacyOpenCCFilePath returns ~/.opencc/opencc.json
func legacyOpenCCFilePath() string {
	return filepath.Join(legacyOpenCCDirPath(), LegacyOpenCCFile)
}

// --- Store ---

// Store manages reading and writing the unified JSON config.
type Store struct {
	mu       sync.Mutex
	path     string
	config   *OpenCCConfig
	modTime  time.Time // last known modification time of config file
}

var (
	defaultStore *Store
	defaultOnce  sync.Once
	defaultMu    sync.Mutex
)

// DefaultStore returns the global Store singleton.
// On first call it loads from disk (with legacy migration if needed).
// On subsequent calls, it checks if the config file has been modified
// and reloads if necessary.
func DefaultStore() *Store {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultStore == nil {
		defaultStore = &Store{path: ConfigFilePath()}
		if err := defaultStore.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		}
	} else {
		// Check if config file has been modified since last load
		if info, err := os.Stat(defaultStore.path); err == nil {
			if info.ModTime().After(defaultStore.modTime) {
				// File has been modified, reload
				defaultStore.Load()
			}
		}
	}
	return defaultStore
}

// ResetDefaultStore clears the singleton so the next DefaultStore() call
// re-initializes. Intended for tests.
func ResetDefaultStore() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultStore = nil
}

// --- Provider operations ---

// GetProvider returns the config for a named provider, or nil.
func (s *Store) GetProvider(name string) *ProviderConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil {
		return nil
	}
	return s.config.Providers[name]
}

// SetProvider creates or updates a provider and saves.
func (s *Store) SetProvider(name string, p *ProviderConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()
	s.config.Providers[name] = p
	return s.saveLocked()
}

// DeleteProvider removes a provider and removes it from all profiles
// (including routing scenarios), then saves.
func (s *Store) DeleteProvider(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()
	delete(s.config.Providers, name)
	for _, pc := range s.config.Profiles {
		pc.Providers = removeString(pc.Providers, name)
		for scenario, route := range pc.Routing {
			route.Providers = filterProviderRoutes(route.Providers, name)
			if len(route.Providers) == 0 {
				delete(pc.Routing, scenario)
			}
		}
	}
	return s.saveLocked()
}

// ProviderNames returns sorted provider names.
func (s *Store) ProviderNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil {
		return nil
	}
	names := make([]string, 0, len(s.config.Providers))
	for n := range s.config.Providers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ProviderMap returns a copy of all providers.
func (s *Store) ProviderMap() map[string]*ProviderConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil {
		return nil
	}
	// Return a copy to avoid concurrent modification
	copy := make(map[string]*ProviderConfig, len(s.config.Providers))
	for k, v := range s.config.Providers {
		copy[k] = v
	}
	return copy
}

// ExportProviderToEnv sets ANTHROPIC_* env vars for the named provider.
func (s *Store) ExportProviderToEnv(name string) error {
	p := s.GetProvider(name)
	if p == nil {
		return fmt.Errorf("provider %q not found", name)
	}
	p.ExportToEnv()
	return nil
}

// --- Profile operations ---

// GetProfileOrder returns a copy of the provider list for a profile.
func (s *Store) GetProfileOrder(profile string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil {
		return nil
	}
	pc := s.config.Profiles[profile]
	if pc == nil {
		return nil
	}
	// Return a copy to avoid callers mutating internal state
	out := make([]string, len(pc.Providers))
	copy(out, pc.Providers)
	return out
}

// SetProfileOrder sets the provider list for a profile and saves.
// Preserves existing routing configuration if the profile already exists.
func (s *Store) SetProfileOrder(profile string, names []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()
	if names == nil {
		names = []string{}
	}
	pc := s.config.Profiles[profile]
	if pc == nil {
		pc = &ProfileConfig{}
		s.config.Profiles[profile] = pc
	}
	pc.Providers = names
	return s.saveLocked()
}

// GetProfileConfig returns the full profile configuration.
func (s *Store) GetProfileConfig(profile string) *ProfileConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil {
		return nil
	}
	return s.config.Profiles[profile]
}

// SetProfileConfig sets the full profile configuration and saves.
func (s *Store) SetProfileConfig(profile string, pc *ProfileConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()
	if pc == nil {
		pc = &ProfileConfig{Providers: []string{}}
	}
	s.config.Profiles[profile] = pc
	return s.saveLocked()
}

// RemoveFromProfile removes a provider name from a specific profile
// (both from the main providers list and from all routing scenarios) and saves.
func (s *Store) RemoveFromProfile(profile, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()
	pc := s.config.Profiles[profile]
	if pc == nil {
		return nil
	}
	pc.Providers = removeString(pc.Providers, name)
	for scenario, route := range pc.Routing {
		route.Providers = filterProviderRoutes(route.Providers, name)
		if len(route.Providers) == 0 {
			delete(pc.Routing, scenario)
		}
	}
	return s.saveLocked()
}

// DeleteProfile deletes a profile. Cannot delete the default profile.
func (s *Store) DeleteProfile(profile string) error {
	if profile == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()

	// Check if this is the default profile
	defaultProfile := s.config.DefaultProfile
	if defaultProfile == "" {
		defaultProfile = DefaultProfileName
	}
	if profile == defaultProfile {
		return fmt.Errorf("cannot delete the default profile '%s'", profile)
	}

	delete(s.config.Profiles, profile)
	return s.saveLocked()
}

// ListProfiles returns sorted profile names.
func (s *Store) ListProfiles() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil {
		return nil
	}
	names := make([]string, 0, len(s.config.Profiles))
	for n := range s.config.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// --- Global Settings ---

// GetDefaultProfile returns the configured default profile name.
// Returns "default" if not set.
func (s *Store) GetDefaultProfile() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil || s.config.DefaultProfile == "" {
		return DefaultProfileName
	}
	return s.config.DefaultProfile
}

// SetDefaultProfile sets the default profile name.
func (s *Store) SetDefaultProfile(profile string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()
	s.config.DefaultProfile = profile
	return s.saveLocked()
}

// GetDefaultClient returns the configured default client.
// Returns "claude" if not set.
func (s *Store) GetDefaultClient() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil || s.config.DefaultClient == "" {
		return DefaultClientName
	}
	return s.config.DefaultClient
}

// SetDefaultClient sets the default client.
func (s *Store) SetDefaultClient(client string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()
	s.config.DefaultClient = client
	return s.saveLocked()
}

// GetWebPort returns the configured web UI port.
// Returns DefaultWebPort if not set.
func (s *Store) GetWebPort() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil || s.config.WebPort == 0 {
		return DefaultWebPort
	}
	return s.config.WebPort
}

// SetWebPort sets the web UI port.
func (s *Store) SetWebPort(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()
	s.config.WebPort = port
	return s.saveLocked()
}

// --- I/O ---

// reloadIfModified checks if the config file has been modified since last load
// and reloads if necessary. Must be called with s.mu held.
func (s *Store) reloadIfModified() {
	if info, err := os.Stat(s.path); err == nil {
		if info.ModTime().After(s.modTime) {
			// File has been modified, reload (ignore errors to avoid breaking operations)
			s.loadLocked()
		}
	}
}

// loadLocked is the internal load implementation. Must be called with s.mu held.
func (s *Store) loadLocked() error {
	data, err := os.ReadFile(s.path)
	if err == nil {
		var cfg OpenCCConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse %s: %w", s.path, err)
		}

		// Check config version
		if cfg.Version > CurrentConfigVersion {
			// Config is newer than this version of zen can handle
			return fmt.Errorf("config version %d is newer than supported version %d, please upgrade zen to the latest version",
				cfg.Version, CurrentConfigVersion)
		}
		if cfg.Version < CurrentConfigVersion {
			// Older config (including version 0 = no version field), upgrade to current
			cfg.Version = CurrentConfigVersion
		}

		if cfg.Providers == nil {
			cfg.Providers = make(map[string]*ProviderConfig)
		}
		if cfg.Profiles == nil {
			cfg.Profiles = make(map[string]*ProfileConfig)
		}
		s.config = &cfg
		// Update modification time
		if info, statErr := os.Stat(s.path); statErr == nil {
			s.modTime = info.ModTime()
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", s.path, err)
	}

	// Try migrating from legacy ~/.opencc/ directory
	legacyOpenCCPath := legacyOpenCCFilePath()
	if _, statErr := os.Stat(legacyOpenCCPath); statErr == nil {
		cfg, migrateErr := MigrateFromOpenCC()
		if migrateErr != nil {
			return fmt.Errorf("migration from ~/.opencc failed: %w", migrateErr)
		}
		if cfg != nil {
			s.config = cfg
			return s.saveLocked()
		}
	}

	// JSON doesn't exist — try legacy migration from ~/.cc_envs
	legacyDir := legacyDirPath()
	if info, statErr := os.Stat(legacyDir); statErr == nil && info.IsDir() {
		cfg, migrateErr := MigrateFromLegacy()
		if migrateErr != nil {
			return fmt.Errorf("migration failed: %w", migrateErr)
		}
		if cfg != nil {
			s.config = cfg
			return s.saveLocked()
		}
	}

	// Nothing exists — create empty config
	s.config = &OpenCCConfig{
		Version:   CurrentConfigVersion,
		Providers: make(map[string]*ProviderConfig),
		Profiles:  make(map[string]*ProfileConfig),
	}
	s.modTime = time.Time{} // zero time for non-existent file
	return nil
}

// Load reads the JSON config from disk. If the file doesn't exist, it tries
// to migrate from the legacy .cc_envs format. If neither exists, it creates
// an empty config.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

// Save writes the config to disk atomically (temp + rename), with 0600 permissions.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	s.ensureConfig()
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, "zen-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename config file: %w", err)
	}
	// Update modification time after successful save
	if info, statErr := os.Stat(s.path); statErr == nil {
		s.modTime = info.ModTime()
	}
	return nil
}

// Reload re-reads the config from disk.
func (s *Store) Reload() error {
	return s.Load()
}

// ensureConfig makes sure s.config is non-nil with initialized maps.
func (s *Store) ensureConfig() {
	if s.config == nil {
		s.config = &OpenCCConfig{
			Version:   CurrentConfigVersion,
			Providers: make(map[string]*ProviderConfig),
			Profiles:  make(map[string]*ProfileConfig),
		}
	}
	if s.config.Providers == nil {
		s.config.Providers = make(map[string]*ProviderConfig)
	}
	if s.config.Profiles == nil {
		s.config.Profiles = make(map[string]*ProfileConfig)
	}
	if s.config.ProjectBindings == nil {
		s.config.ProjectBindings = make(map[string]*ProjectBinding)
	}
	// Ensure version is set
	if s.config.Version == 0 {
		s.config.Version = CurrentConfigVersion
	}
}

// --- helpers ---

func removeString(ss []string, s string) []string {
	var out []string
	for _, v := range ss {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

func filterProviderRoutes(routes []*ProviderRoute, name string) []*ProviderRoute {
	var out []*ProviderRoute
	for _, pr := range routes {
		if pr.Name != name {
			out = append(out, pr)
		}
	}
	return out
}

// --- Project Bindings ---

// resolveProjectPath resolves symlinks and cleans the path to ensure
// consistent binding keys regardless of how the path was reached.
func resolveProjectPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

// BindProject binds a directory path to a profile and/or CLI.
// Either profile or cli can be empty to use the default.
func (s *Store) BindProject(path string, profile string, cli string) error {
	path = resolveProjectPath(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()

	// Verify profile exists if specified
	if profile != "" {
		if _, ok := s.config.Profiles[profile]; !ok {
			return fmt.Errorf("profile '%s' does not exist", profile)
		}
	}

	// Verify client is valid if specified
	if cli != "" && !IsValidClient(cli) {
		return fmt.Errorf("invalid client '%s' (must be %v)", cli, AvailableClients)
	}

	s.config.ProjectBindings[path] = &ProjectBinding{
		Profile: profile,
		Client:  cli,
	}
	return s.saveLocked()
}

// UnbindProject removes the binding for a directory path.
func (s *Store) UnbindProject(path string) error {
	path = resolveProjectPath(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	s.ensureConfig()

	delete(s.config.ProjectBindings, path)
	return s.saveLocked()
}

// GetProjectBinding returns the binding for a directory path.
// Returns nil if no binding exists.
func (s *Store) GetProjectBinding(path string) *ProjectBinding {
	path = resolveProjectPath(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil || s.config.ProjectBindings == nil {
		return nil
	}
	return s.config.ProjectBindings[path]
}

// GetAllProjectBindings returns all project bindings.
func (s *Store) GetAllProjectBindings() map[string]*ProjectBinding {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadIfModified()
	if s.config == nil || s.config.ProjectBindings == nil {
		return make(map[string]*ProjectBinding)
	}
	// Return a copy to avoid concurrent modification
	bindings := make(map[string]*ProjectBinding, len(s.config.ProjectBindings))
	for k, v := range s.config.ProjectBindings {
		bindings[k] = v
	}
	return bindings
}

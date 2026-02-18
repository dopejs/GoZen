package config

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MigrateFromLegacy reads the legacy ~/.cc_envs directory and converts
// *.env files and fallback*.conf files into an OpenCCConfig.
// Returns nil if the legacy directory has no .env files.
func MigrateFromLegacy() (*OpenCCConfig, error) {
	dir := legacyDirPath()

	// 1. Read all *.env files → providers
	envMatches, _ := filepath.Glob(filepath.Join(dir, "*.env"))
	if len(envMatches) == 0 {
		return nil, nil
	}

	providers := make(map[string]*ProviderConfig)
	for _, path := range envMatches {
		name := strings.TrimSuffix(filepath.Base(path), ".env")
		p, err := parseLegacyEnvFile(path)
		if err != nil {
			continue
		}
		providers[name] = p
	}

	if len(providers) == 0 {
		return nil, nil
	}

	// 2. Read fallback*.conf files → profiles
	profiles := make(map[string]*ProfileConfig)
	confMatches, _ := filepath.Glob(filepath.Join(dir, "fallback*.conf"))
	for _, path := range confMatches {
		base := filepath.Base(path)
		var profileName string
		switch {
		case base == "fallback.conf":
			profileName = "default"
		case strings.HasPrefix(base, "fallback.") && strings.HasSuffix(base, ".conf"):
			profileName = strings.TrimPrefix(base, "fallback.")
			profileName = strings.TrimSuffix(profileName, ".conf")
			if profileName == "" {
				continue
			}
		default:
			continue
		}

		names, err := parseLegacyConfFile(path)
		if err != nil {
			continue
		}
		if len(names) > 0 {
			profiles[profileName] = &ProfileConfig{Providers: names}
		}
	}

	return &OpenCCConfig{
		Providers: providers,
		Profiles:  profiles,
	}, nil
}

// parseLegacyEnvFile parses a key=value .env file into a ProviderConfig.
func parseLegacyEnvFile(path string) (*ProviderConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	kv := make(map[string]string)
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
		kv[k] = v
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &ProviderConfig{
		BaseURL:        kv["ANTHROPIC_BASE_URL"],
		AuthToken:      kv["ANTHROPIC_AUTH_TOKEN"],
		Model:          kv["ANTHROPIC_MODEL"],
		ReasoningModel: kv["ANTHROPIC_REASONING_MODEL"],
		HaikuModel:     kv["ANTHROPIC_DEFAULT_HAIKU_MODEL"],
		OpusModel:      kv["ANTHROPIC_DEFAULT_OPUS_MODEL"],
		SonnetModel:    kv["ANTHROPIC_DEFAULT_SONNET_MODEL"],
	}, nil
}

// parseLegacyConfFile reads a fallback*.conf file (one provider name per line).
func parseLegacyConfFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		names = append(names, line)
	}
	return names, scanner.Err()
}

// MigrateFromOpenCC reads the legacy ~/.opencc/opencc.json and converts it
// to the new config format. Also copies auxiliary files (logs, PID files, DB)
// from ~/.opencc/ to ~/.zen/.
func MigrateFromOpenCC() (*OpenCCConfig, error) {
	oldPath := legacyOpenCCFilePath()
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return nil, err
	}

	var cfg OpenCCConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cfg.Version = CurrentConfigVersion

	if cfg.Providers == nil {
		cfg.Providers = make(map[string]*ProviderConfig)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*ProfileConfig)
	}

	// Copy auxiliary files from ~/.opencc/ to ~/.zen/
	newDir := ConfigDirPath()
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return nil, err
	}

	oldDir := legacyOpenCCDirPath()
	auxiliaryFiles := []string{"proxy.log", "web.log", "*.pid", "*.db"}
	for _, pattern := range auxiliaryFiles {
		matches, _ := filepath.Glob(filepath.Join(oldDir, pattern))
		for _, src := range matches {
			dst := filepath.Join(newDir, filepath.Base(src))
			copyFile(src, dst)
		}
	}

	return &cfg, nil
}

// copyFile copies a single file from src to dst, ignoring errors.
// If copy fails, the incomplete destination file is removed.
func copyFile(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst) // Clean up incomplete file
		return
	}
	out.Close()
}

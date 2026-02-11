package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/anthropics/opencc/internal/envfile"
)

// ProfileConfPath returns the conf file path for a profile.
// "default" or "" → fallback.conf, others → fallback.<name>.conf
func ProfileConfPath(profile string) string {
	if profile == "" || profile == "default" {
		return filepath.Join(envfile.EnvsPath(), "fallback.conf")
	}
	return filepath.Join(envfile.EnvsPath(), "fallback."+profile+".conf")
}

// FallbackConfPath returns the path to fallback.conf (default profile).
func FallbackConfPath() string {
	return ProfileConfPath("default")
}

// ReadProfileOrder reads the provider names from a profile's conf file.
func ReadProfileOrder(profile string) ([]string, error) {
	path := ProfileConfPath(profile)
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

// WriteProfileOrder writes the provider names to a profile's conf file.
func WriteProfileOrder(profile string, names []string) error {
	dir := envfile.EnvsPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	var b strings.Builder
	for _, n := range names {
		fmt.Fprintln(&b, n)
	}
	return os.WriteFile(ProfileConfPath(profile), []byte(b.String()), 0644)
}

// RemoveFromProfileOrder removes a name from a profile's conf file.
func RemoveFromProfileOrder(profile, name string) error {
	names, err := ReadProfileOrder(profile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var filtered []string
	for _, n := range names {
		if n != name {
			filtered = append(filtered, n)
		}
	}
	return WriteProfileOrder(profile, filtered)
}

// DeleteProfile deletes a profile's conf file. Cannot delete "default".
func DeleteProfile(profile string) error {
	if profile == "" || profile == "default" {
		return fmt.Errorf("cannot delete the default profile")
	}
	return os.Remove(ProfileConfPath(profile))
}

// ListProfiles returns all profile names by globbing fallback*.conf files.
func ListProfiles() []string {
	dir := envfile.EnvsPath()
	matches, _ := filepath.Glob(filepath.Join(dir, "fallback*.conf"))
	var profiles []string
	for _, m := range matches {
		base := filepath.Base(m)
		switch {
		case base == "fallback.conf":
			profiles = append(profiles, "default")
		case strings.HasPrefix(base, "fallback.") && strings.HasSuffix(base, ".conf"):
			name := strings.TrimPrefix(base, "fallback.")
			name = strings.TrimSuffix(name, ".conf")
			if name != "" {
				profiles = append(profiles, name)
			}
		}
	}
	sort.Strings(profiles)
	return profiles
}

// Delegate functions for backward compatibility.

func ReadFallbackOrder() ([]string, error)      { return ReadProfileOrder("default") }
func WriteFallbackOrder(names []string) error    { return WriteProfileOrder("default", names) }
func RemoveFromFallbackOrder(name string) error  { return RemoveFromProfileOrder("default", name) }

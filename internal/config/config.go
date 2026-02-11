package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/opencc/internal/envfile"
)

// FallbackConfPath returns the path to fallback.conf.
func FallbackConfPath() string {
	return filepath.Join(envfile.EnvsPath(), "fallback.conf")
}

// ReadFallbackOrder reads the provider names from fallback.conf.
func ReadFallbackOrder() ([]string, error) {
	path := FallbackConfPath()
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

// WriteFallbackOrder writes the provider names to fallback.conf.
func WriteFallbackOrder(names []string) error {
	dir := envfile.EnvsPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	var b strings.Builder
	for _, n := range names {
		fmt.Fprintln(&b, n)
	}
	return os.WriteFile(FallbackConfPath(), []byte(b.String()), 0644)
}

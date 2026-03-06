package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GetGitModTime gets the last modification time of a file from git log
func GetGitModTime(filePath string) (time.Time, error) {
	// Run git log to get the last commit timestamp for this file
	cmd := exec.Command("git", "log", "-1", "--format=%ct", filePath)
	output, err := cmd.Output()
	if err != nil {
		// If git command fails, return zero time
		return time.Time{}, fmt.Errorf("git log failed: %w", err)
	}

	// Parse Unix timestamp
	timestampStr := strings.TrimSpace(string(output))
	if timestampStr == "" {
		return time.Time{}, fmt.Errorf("no git history for file")
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return time.Unix(timestamp, 0), nil
}

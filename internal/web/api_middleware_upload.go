package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dopejs/gozen/internal/config"
)

const maxUploadSize = 50 * 1024 * 1024 // 50MB

// handleMiddlewareUpload handles POST /api/v1/middleware/upload
func (s *Server) handleMiddlewareUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "file too large or invalid form data")
		return
	}

	// Get the file from form
	file, header, err := r.FormFile("plugin")
	if err != nil {
		writeError(w, http.StatusBadRequest, "plugin file is required")
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(header.Filename, ".so") {
		writeError(w, http.StatusBadRequest, "only .so files are allowed")
		return
	}

	// Get plugin name from form or use filename
	pluginName := r.FormValue("name")
	if pluginName == "" {
		pluginName = strings.TrimSuffix(header.Filename, ".so")
	}

	// Sanitize plugin name
	pluginName = sanitizePluginName(pluginName)
	if pluginName == "" {
		writeError(w, http.StatusBadRequest, "invalid plugin name")
		return
	}

	// Create plugins directory if not exists
	pluginsDir := filepath.Join(config.ConfigDirPath(), "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create plugins directory")
		return
	}

	// Create temp file first
	tempFile, err := os.CreateTemp(pluginsDir, "upload-*.tmp")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp file")
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up temp file on error

	// Copy uploaded file to temp file and calculate checksum
	hash := sha256.New()
	writer := io.MultiWriter(tempFile, hash)

	written, err := io.Copy(writer, file)
	if err != nil {
		tempFile.Close()
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	tempFile.Close()

	// Generate final filename with checksum prefix for cache busting
	checksum := hex.EncodeToString(hash.Sum(nil))[:8]
	finalName := fmt.Sprintf("%s-%s.so", pluginName, checksum)
	finalPath := filepath.Join(pluginsDir, finalName)

	// Move temp file to final location
	if err := os.Rename(tempPath, finalPath); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save plugin file")
		return
	}

	s.logger.Printf("Plugin uploaded: %s (%d bytes)", finalPath, written)

	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"name":     pluginName,
		"path":     finalPath,
		"checksum": hex.EncodeToString(hash.Sum(nil)),
		"size":     fmt.Sprintf("%d", written),
	})
}

// sanitizePluginName removes unsafe characters from plugin name
func sanitizePluginName(name string) string {
	// Only allow alphanumeric, dash, and underscore
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

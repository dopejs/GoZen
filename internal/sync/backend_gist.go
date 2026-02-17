package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const gistFileName = "zen-sync.json"

// GistBackend implements Backend using GitHub Gist REST API.
type GistBackend struct {
	GistID string
	Token  string // GitHub PAT with gist scope
}

func (b *GistBackend) Name() string { return "gist" }

// gistResponse is the subset of the Gist API response we need.
type gistResponse struct {
	ID    string                    `json:"id"`
	Files map[string]*gistFileInfo `json:"files"`
}

type gistFileInfo struct {
	Content string `json:"content"`
}

func (b *GistBackend) Download(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/gists/%s", b.GistID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("gist download: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+b.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gist download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gist download: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var gist gistResponse
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, fmt.Errorf("gist download: %w", err)
	}

	file, ok := gist.Files[gistFileName]
	if !ok || file == nil {
		return nil, nil
	}
	return []byte(file.Content), nil
}

func (b *GistBackend) Upload(ctx context.Context, data []byte) error {
	url := fmt.Sprintf("https://api.github.com/gists/%s", b.GistID)
	payload := map[string]interface{}{
		"files": map[string]interface{}{
			gistFileName: map[string]string{
				"content": string(data),
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gist upload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gist upload: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+b.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("gist upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gist upload: HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// CreateGist creates a new private gist and returns its ID.
func CreateGist(ctx context.Context, token string) (string, error) {
	payload := map[string]interface{}{
		"description": "GoZen Config Sync",
		"public":      false,
		"files": map[string]interface{}{
			gistFileName: map[string]string{
				"content": "{}",
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("create gist: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/gists", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create gist: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create gist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create gist: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var gist gistResponse
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return "", fmt.Errorf("create gist: %w", err)
	}
	return gist.ID, nil
}

package sync

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RepoBackend implements Backend using GitHub Repository Contents API.
type RepoBackend struct {
	Owner  string
	Repo   string
	Path   string // default: "zen-sync.json"
	Branch string // default: "main"
	Token  string // GitHub PAT with repo scope
}

func (b *RepoBackend) Name() string { return "repo" }

func (b *RepoBackend) filePath() string {
	if b.Path == "" {
		return "zen-sync.json"
	}
	return b.Path
}

func (b *RepoBackend) branch() string {
	if b.Branch == "" {
		return "main"
	}
	return b.Branch
}

// contentsResponse is the subset of the Contents API response we need.
type contentsResponse struct {
	SHA     string `json:"sha"`
	Content string `json:"content"`
}

func (b *RepoBackend) Download(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		b.Owner, b.Repo, b.filePath(), b.branch())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("repo download: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+b.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("repo download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("repo download: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var cr contentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("repo download: %w", err)
	}

	data, err := base64.StdEncoding.DecodeString(cr.Content)
	if err != nil {
		return nil, fmt.Errorf("repo download: decode base64: %w", err)
	}
	return data, nil
}

func (b *RepoBackend) Upload(ctx context.Context, data []byte) error {
	// First GET to obtain current SHA (needed for update)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		b.Owner, b.Repo, b.filePath(), b.branch())

	var currentSHA string
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("repo upload: %w", err)
	}
	getReq.Header.Set("Authorization", "Bearer "+b.Token)
	getReq.Header.Set("Accept", "application/vnd.github+json")

	getResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		return fmt.Errorf("repo upload: %w", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode == http.StatusOK {
		var cr contentsResponse
		if err := json.NewDecoder(getResp.Body).Decode(&cr); err == nil {
			currentSHA = cr.SHA
		}
	}
	// If 404, file doesn't exist yet — create without SHA

	// PUT to create/update
	putURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		b.Owner, b.Repo, b.filePath())

	payload := map[string]interface{}{
		"message": "Update GoZen sync config",
		"content": base64.StdEncoding.EncodeToString(data),
		"branch":  b.branch(),
	}
	if currentSHA != "" {
		payload["sha"] = currentSHA
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("repo upload: %w", err)
	}

	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("repo upload: %w", err)
	}
	putReq.Header.Set("Authorization", "Bearer "+b.Token)
	putReq.Header.Set("Accept", "application/vnd.github+json")
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		return fmt.Errorf("repo upload: %w", err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(putResp.Body)
		return fmt.Errorf("repo upload: HTTP %d: %s", putResp.StatusCode, string(respBody))
	}
	return nil
}

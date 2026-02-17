package sync

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// WebDAVBackend implements Backend using HTTP GET/PUT with optional Basic Auth.
type WebDAVBackend struct {
	Endpoint string // Full URL to the remote file, e.g. https://dav.example.com/zen-sync.json
	Username string
	Password string
}

func (b *WebDAVBackend) Name() string { return "webdav" }

func (b *WebDAVBackend) Download(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("webdav download: %w", err)
	}
	if b.Username != "" || b.Password != "" {
		req.SetBasicAuth(b.Username, b.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("webdav download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("webdav download: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("webdav download: %w", err)
	}
	return data, nil
}

func (b *WebDAVBackend) Upload(ctx context.Context, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, b.Endpoint, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("webdav upload: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if b.Username != "" || b.Password != "" {
		req.SetBasicAuth(b.Username, b.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("webdav upload: %w", err)
	}
	defer resp.Body.Close()

	// WebDAV PUT typically returns 200, 201, or 204
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("webdav upload: HTTP %d: %s", resp.StatusCode, string(body))
}

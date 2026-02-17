package sync

import (
	"context"
	"fmt"

	"github.com/dopejs/gozen/internal/config"
)

// Backend is the interface for remote sync storage.
type Backend interface {
	// Download fetches the remote payload. Returns nil,nil if not found.
	Download(ctx context.Context) ([]byte, error)
	// Upload writes data to the remote storage.
	Upload(ctx context.Context, data []byte) error
	// Name returns the backend type name.
	Name() string
}

// NewBackend creates a Backend from the given SyncConfig.
func NewBackend(cfg *config.SyncConfig) (Backend, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sync config is nil")
	}
	switch cfg.Backend {
	case "webdav":
		return &WebDAVBackend{
			Endpoint: cfg.Endpoint,
			Username: cfg.Username,
			Password: cfg.Token,
		}, nil
	case "s3":
		return NewS3Backend(cfg)
	case "gist":
		return &GistBackend{
			GistID: cfg.GistID,
			Token:  cfg.Token,
		}, nil
	case "repo":
		return &RepoBackend{
			Owner:  cfg.RepoOwner,
			Repo:   cfg.RepoName,
			Path:   cfg.RepoPath,
			Branch: cfg.RepoBranch,
			Token:  cfg.Token,
		}, nil
	default:
		return nil, fmt.Errorf("unknown sync backend: %q", cfg.Backend)
	}
}

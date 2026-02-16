package proxy

import (
	"fmt"
	"strings"
)

// RouteInfo holds the parsed routing information from a request URL path.
type RouteInfo struct {
	Profile   string // profile name (e.g., "default", "work", "_tmp_8f3a2b")
	SessionID string // session UUID (e.g., "f47ac10b")
	Remainder string // remaining path after profile/session (e.g., "/v1/messages")
}

// ParseRoutePath extracts profile and session from a URL path.
// Expected format: /<profile>/<session>/v1/...
// Returns an error if the path doesn't match the expected format.
func ParseRoutePath(path string) (*RouteInfo, error) {
	// Trim leading slash
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	// Split into segments
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("path must contain at least /<profile>/<session>/...")
	}

	profile := parts[0]
	session := parts[1]

	if profile == "" {
		return nil, fmt.Errorf("profile segment is empty")
	}
	if session == "" {
		return nil, fmt.Errorf("session segment is empty")
	}

	remainder := ""
	if len(parts) == 3 {
		remainder = "/" + parts[2]
	}

	return &RouteInfo{
		Profile:   profile,
		SessionID: session,
		Remainder: remainder,
	}, nil
}

// CacheKey returns the session cache key in the format "<profile>:<session_id>".
func (ri *RouteInfo) CacheKey() string {
	return ri.Profile + ":" + ri.SessionID
}

// IsTempProfile returns true if this is a temporary profile (from zen pick).
func (ri *RouteInfo) IsTempProfile() bool {
	return strings.HasPrefix(ri.Profile, "_tmp_")
}

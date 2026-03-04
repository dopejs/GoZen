package web

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetSettings_IncludesProxyPort(t *testing.T) {
	s := setupTestServer(t)

	w := doRequest(s, http.MethodGet, "/api/v1/settings", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var resp settingsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.ProxyPort == 0 {
		t.Error("proxy_port should not be zero in settings response")
	}
}

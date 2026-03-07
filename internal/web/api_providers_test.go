package web

import (
	"encoding/json"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

// T016: Test disable/enable endpoints
func TestProviderDisableEnable(t *testing.T) {
	s := setupTestServer(t)

	// Test disable with valid type
	w := doRequest(s, "POST", "/api/v1/providers/test-provider/disable", map[string]string{"type": "today"})
	if w.Code != 200 {
		t.Fatalf("disable status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var disableResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &disableResp)
	if disableResp["provider"] != "test-provider" {
		t.Errorf("provider = %v, want test-provider", disableResp["provider"])
	}
	if disableResp["disabled"] != true {
		t.Errorf("disabled = %v, want true", disableResp["disabled"])
	}
	if disableResp["type"] != "today" {
		t.Errorf("type = %v, want today", disableResp["type"])
	}

	// Verify provider is actually disabled
	if !config.IsProviderDisabled("test-provider") {
		t.Error("provider should be disabled after disable endpoint")
	}

	// Test disable with invalid type
	w = doRequest(s, "POST", "/api/v1/providers/test-provider/disable", map[string]string{"type": "invalid"})
	if w.Code != 400 {
		t.Errorf("disable invalid type status = %d, want 400", w.Code)
	}

	// Test disable nonexistent provider
	w = doRequest(s, "POST", "/api/v1/providers/nonexistent/disable", map[string]string{"type": "today"})
	if w.Code != 404 {
		t.Errorf("disable nonexistent status = %d, want 404", w.Code)
	}

	// Test enable
	w = doRequest(s, "POST", "/api/v1/providers/test-provider/enable", nil)
	if w.Code != 200 {
		t.Fatalf("enable status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var enableResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &enableResp)
	if enableResp["disabled"] != false {
		t.Errorf("disabled = %v, want false", enableResp["disabled"])
	}

	// Verify provider is actually enabled
	if config.IsProviderDisabled("test-provider") {
		t.Error("provider should be enabled after enable endpoint")
	}

	// Test enable nonexistent provider
	w = doRequest(s, "POST", "/api/v1/providers/nonexistent/enable", nil)
	if w.Code != 404 {
		t.Errorf("enable nonexistent status = %d, want 404", w.Code)
	}

	// Test disable with month type
	w = doRequest(s, "POST", "/api/v1/providers/backup/disable", map[string]string{"type": "month"})
	if w.Code != 200 {
		t.Fatalf("disable month status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	// Test disable with permanent type
	w = doRequest(s, "POST", "/api/v1/providers/test-provider/disable", map[string]string{"type": "permanent"})
	if w.Code != 200 {
		t.Fatalf("disable permanent status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}

// T017: Test disabled list endpoint and provider list includes disabled field
func TestDisabledProvidersList(t *testing.T) {
	s := setupTestServer(t)

	// Initially no disabled providers
	w := doRequest(s, "GET", "/api/v1/providers/disabled", nil)
	if w.Code != 200 {
		t.Fatalf("disabled list status = %d, want 200", w.Code)
	}
	var listResp struct {
		DisabledProviders []struct {
			Provider string `json:"provider"`
			Type     string `json:"type"`
		} `json:"disabled_providers"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if len(listResp.DisabledProviders) != 0 {
		t.Errorf("initially disabled = %d, want 0", len(listResp.DisabledProviders))
	}

	// Disable a provider
	doRequest(s, "POST", "/api/v1/providers/test-provider/disable", map[string]string{"type": "permanent"})

	// Now list should show it
	w = doRequest(s, "GET", "/api/v1/providers/disabled", nil)
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if len(listResp.DisabledProviders) != 1 {
		t.Fatalf("disabled list = %d, want 1", len(listResp.DisabledProviders))
	}
	if listResp.DisabledProviders[0].Provider != "test-provider" {
		t.Errorf("disabled provider = %q, want %q", listResp.DisabledProviders[0].Provider, "test-provider")
	}
	if listResp.DisabledProviders[0].Type != "permanent" {
		t.Errorf("disabled type = %q, want %q", listResp.DisabledProviders[0].Type, "permanent")
	}

	// Test that GET /api/v1/providers includes disabled field
	w = doRequest(s, "GET", "/api/v1/providers", nil)
	if w.Code != 200 {
		t.Fatalf("providers list status = %d, want 200", w.Code)
	}
	var providers []struct {
		Name     string                     `json:"name"`
		Disabled *config.UnavailableMarking `json:"disabled,omitempty"`
	}
	json.Unmarshal(w.Body.Bytes(), &providers)

	// Find the disabled provider
	var found bool
	for _, p := range providers {
		if p.Name == "test-provider" {
			found = true
			if p.Disabled == nil {
				t.Error("test-provider should have disabled field in provider list")
			} else if p.Disabled.Type != "permanent" {
				t.Errorf("disabled type = %q, want %q", p.Disabled.Type, "permanent")
			}
		}
		if p.Name == "backup" {
			if p.Disabled != nil {
				t.Error("backup should not have disabled field")
			}
		}
	}
	if !found {
		t.Error("test-provider not found in provider list")
	}
}

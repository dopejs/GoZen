package web

import (
	"encoding/json"
	"net/http"
	"testing"
)

// T005: Test settings API GET/PUT with show_provider_tag field

func TestSettingsGetIncludesShowProviderTag(t *testing.T) {
	s := setupTestServer(t)

	w := doRequest(s, http.MethodGet, "/api/v1/settings", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /settings status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// show_provider_tag should be present and false by default
	v, ok := resp["show_provider_tag"]
	if !ok {
		t.Fatal("response missing show_provider_tag field")
	}
	if v != false {
		t.Errorf("show_provider_tag = %v, want false", v)
	}
}

func TestSettingsPutShowProviderTagTrue(t *testing.T) {
	s := setupTestServer(t)

	// Enable show_provider_tag
	body := map[string]interface{}{"show_provider_tag": true}
	w := doRequest(s, http.MethodPut, "/api/v1/settings", body)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT /settings status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	// Verify GET reflects the change
	w2 := doRequest(s, http.MethodGet, "/api/v1/settings", nil)
	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)

	if resp["show_provider_tag"] != true {
		t.Errorf("after PUT true, show_provider_tag = %v, want true", resp["show_provider_tag"])
	}
}

func TestSettingsPutShowProviderTagFalse(t *testing.T) {
	s := setupTestServer(t)

	// First enable
	doRequest(s, http.MethodPut, "/api/v1/settings", map[string]interface{}{"show_provider_tag": true})

	// Then disable
	body := map[string]interface{}{"show_provider_tag": false}
	w := doRequest(s, http.MethodPut, "/api/v1/settings", body)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT /settings (false) status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	// Verify GET reflects the change
	w2 := doRequest(s, http.MethodGet, "/api/v1/settings", nil)
	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)

	if resp["show_provider_tag"] != false {
		t.Errorf("after PUT false, show_provider_tag = %v, want false", resp["show_provider_tag"])
	}
}

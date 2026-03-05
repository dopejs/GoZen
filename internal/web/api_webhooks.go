package web

import (
	"net/http"
	"strings"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/notify"
)

// handleWebhooks handles GET/POST /api/v1/webhooks - list or create webhooks.
func (s *Server) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		webhooks := config.GetWebhooks()
		if webhooks == nil {
			webhooks = []*config.WebhookConfig{}
		}

		// Mask URLs for security (show only domain)
		masked := make([]map[string]interface{}, len(webhooks))
		for i, wh := range webhooks {
			masked[i] = map[string]interface{}{
				"name":    wh.Name,
				"url":     maskWebhookURL(wh.URL),
				"events":  wh.Events,
				"enabled": wh.Enabled,
			}
		}

		writeJSON(w, http.StatusOK, masked)

	case http.MethodPost:
		var webhook config.WebhookConfig
		if err := readJSON(r, &webhook); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if webhook.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		if webhook.URL == "" {
			writeError(w, http.StatusBadRequest, "url is required")
			return
		}
		if len(webhook.Events) == 0 {
			writeError(w, http.StatusBadRequest, "at least one event is required")
			return
		}

		if err := config.AddWebhook(&webhook); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Reload dispatcher
		notify.GetGlobalDispatcher().ReloadConfig()

		writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "name": webhook.Name})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleWebhook handles GET/PUT/DELETE /api/v1/webhooks/{name} - manage a specific webhook.
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Extract webhook name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/webhooks/")
	webhookName := strings.TrimSuffix(path, "/")

	if webhookName == "" {
		writeError(w, http.StatusBadRequest, "webhook name required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		webhook := config.GetWebhook(webhookName)
		if webhook == nil {
			writeError(w, http.StatusNotFound, "webhook not found")
			return
		}

		// Mask URL for security
		masked := map[string]interface{}{
			"name":    webhook.Name,
			"url":     maskWebhookURL(webhook.URL),
			"events":  webhook.Events,
			"enabled": webhook.Enabled,
			"headers": len(webhook.Headers) > 0,
		}

		writeJSON(w, http.StatusOK, masked)

	case http.MethodPut:
		var webhook config.WebhookConfig
		if err := readJSON(r, &webhook); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		// Preserve name from URL
		webhook.Name = webhookName

		if webhook.URL == "" {
			writeError(w, http.StatusBadRequest, "url is required")
			return
		}

		if err := config.AddWebhook(&webhook); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Reload dispatcher
		notify.GetGlobalDispatcher().ReloadConfig()

		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	case http.MethodDelete:
		if err := config.DeleteWebhook(webhookName); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Reload dispatcher
		notify.GetGlobalDispatcher().ReloadConfig()

		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleWebhookTest handles POST /api/v1/webhooks/test - test a webhook.
func (s *Server) handleWebhookTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Name string `json:"name"`
		URL  string `json:"url,omitempty"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	var webhook *config.WebhookConfig

	if req.Name != "" {
		// Test existing webhook
		webhook = config.GetWebhook(req.Name)
		if webhook == nil {
			writeError(w, http.StatusNotFound, "webhook not found")
			return
		}
	} else if req.URL != "" {
		// Test with provided URL
		webhook = &config.WebhookConfig{
			Name:    "test",
			URL:     req.URL,
			Enabled: true,
		}
	} else {
		writeError(w, http.StatusBadRequest, "name or url required")
		return
	}

	dispatcher := notify.GetGlobalDispatcher()
	if err := dispatcher.TestWebhook(webhook); err != nil {
		writeError(w, http.StatusBadGateway, "webhook test failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "test notification sent"})
}

// maskWebhookURL masks a webhook URL for display, showing only the domain.
func maskWebhookURL(url string) string {
	if url == "" {
		return ""
	}

	// Find the domain part
	start := 0
	if strings.HasPrefix(url, "https://") {
		start = 8
	} else if strings.HasPrefix(url, "http://") {
		start = 7
	}

	end := strings.Index(url[start:], "/")
	if end == -1 {
		end = len(url) - start
	}

	domain := url[start : start+end]
	return url[:start] + domain + "/***"
}

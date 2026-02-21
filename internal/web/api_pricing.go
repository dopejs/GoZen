package web

import (
	"net/http"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// handlePricing handles GET/PUT /api/v1/pricing - get or set model pricing.
func (s *Server) handlePricing(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return merged pricing (defaults + custom overrides)
		pricing := config.GetPricing()

		// Also include info about which are defaults vs custom
		store := config.DefaultStore()
		customPricing := store.GetPricing()

		response := struct {
			Pricing  map[string]*config.ModelPricing `json:"pricing"`
			Defaults map[string]*config.ModelPricing `json:"defaults"`
			Custom   map[string]*config.ModelPricing `json:"custom,omitempty"`
		}{
			Pricing:  pricing,
			Defaults: config.DefaultModelPricing,
			Custom:   customPricing,
		}

		writeJSON(w, http.StatusOK, response)

	case http.MethodPut:
		var pricing map[string]*config.ModelPricing
		if err := readJSON(r, &pricing); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		// Validate pricing values
		for model, p := range pricing {
			if p.InputPerMillion < 0 || p.OutputPerMillion < 0 {
				writeError(w, http.StatusBadRequest, "pricing values must be non-negative for model: "+model)
				return
			}
		}

		if err := config.SetPricing(pricing); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Reload usage tracker pricing
		if tracker := proxy.GetGlobalUsageTracker(); tracker != nil {
			tracker.ReloadPricing()
		}

		// Reload load balancer pricing
		if lb := proxy.GetGlobalLoadBalancer(); lb != nil {
			lb.ReloadPricing()
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handlePricingReset handles POST /api/v1/pricing/reset - reset to default pricing.
func (s *Server) handlePricingReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Clear custom pricing (nil means use defaults only)
	if err := config.SetPricing(nil); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Reload usage tracker pricing
	if tracker := proxy.GetGlobalUsageTracker(); tracker != nil {
		tracker.ReloadPricing()
	}

	// Reload load balancer pricing
	if lb := proxy.GetGlobalLoadBalancer(); lb != nil {
		lb.ReloadPricing()
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "reset to defaults"})
}

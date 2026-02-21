package web

import (
	"encoding/json"
	"net/http"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// CompressionConfigResponse is the API response for compression config.
type CompressionConfigResponse struct {
	Enabled         bool   `json:"enabled"`
	ThresholdTokens int    `json:"threshold_tokens"`
	TargetTokens    int    `json:"target_tokens"`
	SummaryModel    string `json:"summary_model"`
	PreserveRecent  int    `json:"preserve_recent"`
	SummaryProvider string `json:"summary_provider"`
}

// CompressionStatsResponse is the API response for compression stats.
type CompressionStatsResponse struct {
	RequestsCompressed int64 `json:"requests_compressed"`
	TokensSaved        int64 `json:"tokens_saved"`
}

// handleCompression routes GET and PUT requests for compression config.
// GET/PUT /api/v1/compression
func (s *Server) handleCompression(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetCompression(w, r)
	case http.MethodPut:
		s.handleSetCompression(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetCompression returns the compression configuration.
// GET /api/v1/compression
func (s *Server) handleGetCompression(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetCompression()
	if cfg == nil {
		cfg = &config.CompressionConfig{
			Enabled:         false,
			ThresholdTokens: proxy.DefaultThresholdTokens,
			TargetTokens:    proxy.DefaultTargetTokens,
			SummaryModel:    proxy.DefaultSummaryModel,
			PreserveRecent:  proxy.DefaultPreserveRecent,
		}
	}

	resp := CompressionConfigResponse{
		Enabled:         cfg.Enabled,
		ThresholdTokens: cfg.ThresholdTokens,
		TargetTokens:    cfg.TargetTokens,
		SummaryModel:    cfg.SummaryModel,
		PreserveRecent:  cfg.PreserveRecent,
		SummaryProvider: cfg.SummaryProvider,
	}

	// Apply defaults for display
	if resp.ThresholdTokens == 0 {
		resp.ThresholdTokens = proxy.DefaultThresholdTokens
	}
	if resp.TargetTokens == 0 {
		resp.TargetTokens = proxy.DefaultTargetTokens
	}
	if resp.SummaryModel == "" {
		resp.SummaryModel = proxy.DefaultSummaryModel
	}
	if resp.PreserveRecent == 0 {
		resp.PreserveRecent = proxy.DefaultPreserveRecent
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSetCompression updates the compression configuration.
// PUT /api/v1/compression
func (s *Server) handleSetCompression(w http.ResponseWriter, r *http.Request) {
	var req CompressionConfigResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cfg := &config.CompressionConfig{
		Enabled:         req.Enabled,
		ThresholdTokens: req.ThresholdTokens,
		TargetTokens:    req.TargetTokens,
		SummaryModel:    req.SummaryModel,
		PreserveRecent:  req.PreserveRecent,
		SummaryProvider: req.SummaryProvider,
	}

	if err := config.SetCompression(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update global compressor config
	proxy.UpdateGlobalCompressorConfig(cfg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleGetCompressionStats returns compression statistics.
// GET /api/v1/compression/stats
func (s *Server) handleGetCompressionStats(w http.ResponseWriter, r *http.Request) {
	compressor := proxy.GetGlobalCompressor()
	if compressor == nil {
		resp := CompressionStatsResponse{
			RequestsCompressed: 0,
			TokensSaved:        0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	stats := compressor.GetStats()
	resp := CompressionStatsResponse{
		RequestsCompressed: stats.RequestsCompressed,
		TokensSaved:        stats.TokensSaved,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

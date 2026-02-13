package handlers

import (
	"net/http"

	"github.com/colton/futurebuild/internal/service"
)

// AssetHandler handles asset/vision status operations
type AssetHandler struct {
	assetService *service.AssetService
}

// NewAssetHandler creates a new AssetHandler
func NewAssetHandler(assetService *service.AssetService) *AssetHandler {
	return &AssetHandler{assetService: assetService}
}

// ListProjectAssets returns assets for a project
func (h *AssetHandler) ListProjectAssets(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// GetVisionStatus returns vision processing status for an asset
func (h *AssetHandler) GetVisionStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

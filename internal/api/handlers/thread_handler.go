package handlers

import (
	"net/http"

	"github.com/colton/futurebuild/internal/service"
)

// ThreadHandler handles conversation thread operations
type ThreadHandler struct {
	threadService *service.ThreadService
}

// NewThreadHandler creates a new ThreadHandler
func NewThreadHandler(threadService *service.ThreadService) *ThreadHandler {
	return &ThreadHandler{threadService: threadService}
}

// ListThreads returns threads for a project
func (h *ThreadHandler) ListThreads(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// CreateThread creates a new thread
func (h *ThreadHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// GetThread returns a specific thread
func (h *ThreadHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// ArchiveThread archives a thread
func (h *ThreadHandler) ArchiveThread(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// UnarchiveThread unarchives a thread
func (h *ThreadHandler) UnarchiveThread(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

// GetThreadMessages returns messages for a thread
func (h *ThreadHandler) GetThreadMessages(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	w.WriteHeader(http.StatusNotImplemented)
}

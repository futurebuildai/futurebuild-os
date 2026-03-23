package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/colton/futurebuild/internal/chat"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// StreamChatProcessor defines the interface for streaming chat requests.
// Satisfied by *chat.ClaudeOrchestrator.
type StreamChatProcessor interface {
	ProcessRequestStreaming(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, req chat.ChatRequest, out chan<- ai.StreamChunk) error
}

// StreamChatHandler handles streaming chat SSE endpoints.
type StreamChatHandler struct {
	orchestrator StreamChatProcessor
}

// NewStreamChatHandler creates a streaming chat handler.
func NewStreamChatHandler(orchestrator StreamChatProcessor) *StreamChatHandler {
	return &StreamChatHandler{orchestrator: orchestrator}
}

// HandleStreamChat handles POST /api/v1/projects/{projectId}/chat/stream.
// Streams Claude's response via SSE, including text deltas and tool events.
func (h *StreamChatHandler) HandleStreamChat(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusInternalServerError)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project_id", http.StatusBadRequest)
		return
	}

	// Parse request body (limit to 64KB to prevent abuse)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body struct {
		ThreadID string `json:"thread_id"`
		Message  string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	threadID, err := uuid.Parse(body.ThreadID)
	if err != nil {
		threadID = uuid.New()
	}

	// Verify SSE support
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send initial connection event
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Create stream channel
	out := make(chan ai.StreamChunk, 64)

	req := chat.ChatRequest{
		ProjectID: projectID,
		ThreadID:  threadID,
		Message:   body.Message,
	}

	// Run streaming in goroutine
	errCh := make(chan error, 1)
	go func() {
		defer close(out)
		errCh <- h.orchestrator.ProcessRequestStreaming(r.Context(), userID, orgID, req, out)
	}()

	// Keepalive ticker
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()

		case chunk, ok := <-out:
			if !ok {
				// Stream ended — check for errors
				if err := <-errCh; err != nil {
					slog.Error("stream chat: error", "error", err)
					errData, _ := json.Marshal(map[string]string{"error": "An internal error occurred"})
					_, _ = fmt.Fprintf(w, "event: error\ndata: %s\n\n", errData)
					flusher.Flush()
				}
				// Send final done event
				doneData, _ := json.Marshal(map[string]interface{}{
					"done":      true,
					"thread_id": threadID.String(),
				})
				_, _ = fmt.Fprintf(w, "event: done\ndata: %s\n\n", doneData)
				flusher.Flush()
				return
			}

			// Marshal and send the chunk
			data, err := json.Marshal(chunk)
			if err != nil {
				slog.Error("stream chat: marshal chunk failed", "error", err)
				continue
			}

			eventType := "delta"
			if chunk.ToolUse != nil {
				eventType = "tool_use"
			} else if chunk.ToolResult != nil {
				eventType = "tool_result"
			} else if chunk.Done {
				eventType = "done"
			}

			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
			flusher.Flush()
		}
	}
}

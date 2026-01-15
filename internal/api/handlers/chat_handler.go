package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/chat"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/google/uuid"
)

// ChatHandler handles chat API requests.
// See PRODUCTION_PLAN.md Step 43.4
type ChatHandler struct {
	orchestrator *chat.Orchestrator
}

// NewChatHandler creates a new ChatHandler with the given orchestrator.
func NewChatHandler(orchestrator *chat.Orchestrator) *ChatHandler {
	return &ChatHandler{orchestrator: orchestrator}
}

// HandleChat processes incoming chat messages.
// See BACKEND_SCOPE.md Section 5.2 (Chat Endpoint)
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 1. Extract identity from Context (RBAC Middleware populates this)
	// See PRODUCTION_PLAN.md Step 43.4: Context only - DO NOT trust request body.
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("chat: missing claims in context", "error", err)
		response.JSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		slog.Error("chat: invalid UserID in claims", "raw_user_id", claims.UserID, "error", err)
		response.JSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("chat: invalid OrgID in claims", "raw_org_id", claims.OrgID, "error", err)
		response.JSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// 2. Parse Request Body
	var req chat.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("chat: invalid request body", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// 3. Validate Request (Non-empty message)
	if req.Message == "" {
		slog.Warn("chat: empty message received", "user_id", userID, "project_id", req.ProjectID)
		response.JSONError(w, http.StatusBadRequest, "Message cannot be empty")
		return
	}

	// Log: Request received (sanitized message preview)
	messagePreview := req.Message
	if len(messagePreview) > 50 {
		messagePreview = messagePreview[:50] + "..."
	}
	slog.Info("chat: request received",
		"user_id", userID,
		"org_id", orgID,
		"project_id", req.ProjectID,
		"message_preview", messagePreview,
	)

	// 4. Process Request via Orchestrator
	resp, err := h.orchestrator.ProcessRequest(r.Context(), userID, orgID, req)
	if err != nil {
		slog.Error("chat: orchestrator error",
			"user_id", userID,
			"project_id", req.ProjectID,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		response.JSONError(w, http.StatusInternalServerError, "Failed to process chat request")
		return
	}

	// Log: Intent classified + duration
	slog.Info("chat: request completed",
		"user_id", userID,
		"project_id", req.ProjectID,
		"intent", resp.Intent,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	// 5. Return Response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("chat: failed to encode response", "error", err)
	}
}

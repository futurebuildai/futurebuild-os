package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/worker"
	"github.com/hibiken/asynq"
)

const (
	maxWebhookBodySize = 1024 * 1024 // 1MB limit
)

// GitHubWebhookHandler handles GitHub webhook events for Automated PR Review.
// See docs/AUTOMATED_PR_REVIEW_PRD.md
type GitHubWebhookHandler struct {
	webhookSecret string
	asynqClient   *asynq.Client
}

// NewGitHubWebhookHandler creates a new handler with HMAC verification.
func NewGitHubWebhookHandler(webhookSecret, redisAddr string) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{
		webhookSecret: webhookSecret,
		asynqClient:   asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

// HandleGitHubWebhook processes incoming GitHub webhook events.
// POST /api/v1/webhooks/github
//
// Security: HMAC-SHA256 verification with fail-closed behavior.
// See docs/AUTOMATED_PR_REVIEW_PRD.md Security Considerations
func (h *GitHubWebhookHandler) HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// Step 1: Fail-Closed - Reject if secret not configured
	if h.webhookSecret == "" {
		slog.Error("webhook/github: webhook secret not configured (fail-closed)")
		response.JSONError(w, http.StatusForbidden, "Webhook not configured")
		return
	}

	// Step 2: Read body with size limit (DoS prevention)
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodySize)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("webhook/github: failed to read body", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Step 3: Verify HMAC signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !h.verifySignature(bodyBytes, signature) {
		slog.Warn("webhook/github: invalid signature",
			"remote_addr", r.RemoteAddr,
		)
		response.JSONError(w, http.StatusForbidden, "Invalid signature")
		return
	}

	// Step 4: Check event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "pull_request" {
		// Silently ignore non-PR events (e.g., push, issue_comment)
		slog.Debug("webhook/github: ignoring non-PR event", "event", eventType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		return
	}

	// Step 5: Parse webhook payload
	var payload githubPRPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		slog.Warn("webhook/github: invalid JSON payload", "error", err)
		response.JSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Step 6: Filter actions - only process opened, synchronize, reopened
	validActions := map[string]bool{
		"opened":      true,
		"synchronize": true,
		"reopened":    true,
	}
	if !validActions[payload.Action] {
		slog.Debug("webhook/github: ignoring action", "action", payload.Action)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		return
	}

	// Step 7: Build ReviewPRPayload
	// Case ID format: GH_{owner}/{repo}#{number}_{sha}
	caseID := fmt.Sprintf("GH_%s/%s#%d_%s",
		payload.Repository.Owner.Login,
		payload.Repository.Name,
		payload.PullRequest.Number,
		payload.PullRequest.Head.SHA,
	)

	reviewPayload := worker.ReviewPRPayload{
		CaseID:   caseID,
		Owner:    payload.Repository.Owner.Login,
		Repo:     payload.Repository.Name,
		PRNumber: payload.PullRequest.Number,
		HeadSHA:  payload.PullRequest.Head.SHA,
		PRTitle:  payload.PullRequest.Title,
	}

	slog.Info("webhook/github: processing PR event",
		"action", payload.Action,
		"owner", reviewPayload.Owner,
		"repo", reviewPayload.Repo,
		"pr_number", reviewPayload.PRNumber,
		"case_id", caseID,
	)

	// Step 8: Enqueue task
	task, err := worker.NewReviewPRTask(reviewPayload)
	if err != nil {
		slog.Error("webhook/github: failed to create task", "error", err)
		response.JSONError(w, http.StatusInternalServerError, "Failed to create review task")
		return
	}

	if _, err := h.asynqClient.Enqueue(task); err != nil {
		slog.Error("webhook/github: failed to enqueue task", "error", err)
		response.JSONError(w, http.StatusInternalServerError, "Failed to enqueue review task")
		return
	}

	slog.Info("webhook/github: review task enqueued", "case_id", caseID)

	// Return 202 Accepted (async processing)
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"accepted","case_id":"` + caseID + `"}`))
}

// verifySignature validates the HMAC-SHA256 signature.
// Uses constant-time comparison to prevent timing attacks.
func (h *GitHubWebhookHandler) verifySignature(payload []byte, signature string) bool {
	if h.webhookSecret == "" {
		return false // Fail closed
	}

	// Strip "sha256=" prefix
	sig := strings.TrimPrefix(signature, "sha256=")
	if sig == signature {
		// No prefix found - invalid format
		return false
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	return hmac.Equal([]byte(sig), []byte(expected))
}

// Close releases resources held by the handler.
func (h *GitHubWebhookHandler) Close() error {
	return h.asynqClient.Close()
}

// --- GitHub Webhook Payload Types ---

type githubPRPayload struct {
	Action      string       `json:"action"`
	PullRequest githubPR     `json:"pull_request"`
	Repository  githubRepo   `json:"repository"`
	Sender      githubSender `json:"sender"`
}

type githubPR struct {
	Number int        `json:"number"`
	Title  string     `json:"title"`
	Head   githubHead `json:"head"`
	Base   githubBase `json:"base"`
}

type githubHead struct {
	SHA string `json:"sha"`
	Ref string `json:"ref"`
}

type githubBase struct {
	Ref string `json:"ref"`
}

type githubRepo struct {
	Name  string      `json:"name"`
	Owner githubOwner `json:"owner"`
}

type githubOwner struct {
	Login string `json:"login"`
}

type githubSender struct {
	Login string `json:"login"`
}

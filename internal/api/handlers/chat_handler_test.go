package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/chat"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Orchestrator Dependencies ---

type mockMessagePersister struct {
	fail bool
}

func (m *mockMessagePersister) SaveMessage(_ context.Context, _ models.ChatMessage) error {
	if m.fail {
		return assert.AnError
	}
	return nil
}

// Pool returns nil for unit tests that don't test transactional behavior.
func (m *mockMessagePersister) Pool() chat.Transactor {
	return nil
}

type mockTaskService struct{}

func (m *mockTaskService) UpdateTaskStatus(_ context.Context, _, _, _ uuid.UUID, _ types.TaskStatus) error {
	return nil
}

type mockScheduleService struct{}

func (m *mockScheduleService) GetTask(_ context.Context, _, _, _ uuid.UUID) (*models.ProjectTask, error) {
	return nil, nil
}

func (m *mockScheduleService) GetProjectSchedule(_ context.Context, _, _ uuid.UUID) (*service.ProjectScheduleSummary, error) {
	return &service.ProjectScheduleSummary{
		ProjectEnd:        time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		CriticalPathCount: 5,
		TotalTasks:        20,
		CompletedTasks:    8,
	}, nil
}

type mockInvoiceService struct{}

func (m *mockInvoiceService) AnalyzeInvoice(_ context.Context, _ uuid.UUID, _ uuid.UUID) (uuid.UUID, *types.InvoiceExtraction, error) {
	return uuid.Nil, nil, nil
}

func (m *mockInvoiceService) SaveExtraction(_ context.Context, _ uuid.UUID, _ *types.InvoiceExtraction, _ *uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

// mockDLQPersister is a no-op DLQPersister for testing.
type mockDLQPersister struct{}

func (m *mockDLQPersister) EnqueueRetry(_ context.Context, _ models.ChatMessage) error {
	return nil
}

// --- Helpers ---

func newTestOrchestrator() *chat.Orchestrator {
	orch, err := chat.NewOrchestratorWithPersister(
		&mockMessagePersister{},
		chat.NewDefaultRegexClassifier(),
		&mockTaskService{},
		&mockScheduleService{},
		&mockInvoiceService{},
		&mockDLQPersister{},
		nil, // AuditWAL
		nil, // AuditCircuitBreaker
	)
	if err != nil {
		panic(err)
	}
	return orch
}

func injectClaims(r *http.Request, userID, orgID string) *http.Request {
	claims := &types.Claims{
		UserID: userID,
		OrgID:  orgID,
		Role:   types.UserRoleAdmin,
	}
	ctx := middleware.WithClaims(r.Context(), claims)
	return r.WithContext(ctx)
}

// --- Tests ---

// TestHandleChat_HappyPath tests the successful flow with valid context and request.
func TestHandleChat_HappyPath(t *testing.T) {
	// Arrange
	handler := NewChatHandler(newTestOrchestrator())
	userID := uuid.New().String()
	orgID := uuid.New().String()

	reqBody := chat.ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Show me the schedule",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	req = injectClaims(req, userID, orgID)
	rr := httptest.NewRecorder()

	// Act
	handler.HandleChat(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp chat.ChatResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Reply)
	assert.Equal(t, types.IntentGetSchedule, resp.Intent)
}

// TestHandleChat_MissingContext tests security: handler should fail if claims are missing.
func TestHandleChat_MissingContext(t *testing.T) {
	// Arrange
	handler := NewChatHandler(newTestOrchestrator())

	reqBody := chat.ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Test",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// No context injection - simulates missing auth middleware
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	// Act
	handler.HandleChat(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Internal Server Error")
}

// TestHandleChat_EmptyMessage tests validation: empty messages should be rejected.
func TestHandleChat_EmptyMessage(t *testing.T) {
	// Arrange
	handler := NewChatHandler(newTestOrchestrator())
	userID := uuid.New().String()
	orgID := uuid.New().String()

	reqBody := chat.ChatRequest{
		ProjectID: uuid.New(),
		Message:   "", // Empty message
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	req = injectClaims(req, userID, orgID)
	rr := httptest.NewRecorder()

	// Act
	handler.HandleChat(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Message cannot be empty")
}

// TestHandleChat_InvalidJSON tests that malformed JSON returns 400.
func TestHandleChat_InvalidJSON(t *testing.T) {
	// Arrange
	handler := NewChatHandler(newTestOrchestrator())
	userID := uuid.New().String()
	orgID := uuid.New().String()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader([]byte("{invalid json")))
	req = injectClaims(req, userID, orgID)
	rr := httptest.NewRecorder()

	// Act
	handler.HandleChat(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request payload")
}

// TestHandleChat_InvalidClaims tests security: handler should fail if claims have invalid UUIDs.
func TestHandleChat_InvalidClaims(t *testing.T) {
	// Arrange
	handler := NewChatHandler(newTestOrchestrator())

	// Case 1: Invalid UserID
	reqBody := chat.ChatRequest{ProjectID: uuid.New(), Message: "Test"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	req = injectClaims(req, "invalid-uuid", uuid.New().String())
	rr := httptest.NewRecorder()

	handler.HandleChat(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Case 2: Invalid OrgID
	req = httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	req = injectClaims(req, uuid.New().String(), "invalid-uuid")
	rr = httptest.NewRecorder()

	handler.HandleChat(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// Mock Orchestrator for Error Testing
type errorOrchestrator struct {
	*chat.Orchestrator
}

// TestHandleChat_OrchestratorError tests internal error handling when orchestrator fails.
func TestHandleChat_OrchestratorError(t *testing.T) {
	// Arrange
	// We need a way to make the Orchestrator fail.
	// Since we can't easily mock the method on the struct without an interface,
	// we'll rely on the fact that Orchestrator.ProcessRequest calls db.SaveMessage.
	// We can inject a failing persister.

	failingPersister := &mockMessagePersister{fail: true}
	// We need to define a new constructor or just build it manually since we are in the same package (test)
	// Actually, we are in 'handlers' package, so we use the public constructor we added.

	// We need to extend the mock helper to support failure
	failingOrch, err := chat.NewOrchestratorWithPersister(
		failingPersister,
		chat.NewDefaultRegexClassifier(),
		&mockTaskService{},
		&mockScheduleService{},
		&mockInvoiceService{},
		&mockDLQPersister{},
		nil, // AuditWAL
		nil, // AuditCircuitBreaker
	)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	handler := NewChatHandler(failingOrch)

	userID := uuid.New().String()
	orgID := uuid.New().String()

	reqBody := chat.ChatRequest{ProjectID: uuid.New(), Message: "Test"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	req = injectClaims(req, userID, orgID)
	rr := httptest.NewRecorder()

	// Act
	handler.HandleChat(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to process chat request")
}

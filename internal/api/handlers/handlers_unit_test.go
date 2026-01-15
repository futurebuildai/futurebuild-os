package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock ProjectService ---

type mockProjectService struct {
	createErr     error
	getErr        error
	getProject    *models.Project
	createCalled  bool
	getCalled     bool
	lastCreatedID uuid.UUID
}

func (m *mockProjectService) CreateProject(ctx context.Context, p *models.Project) error {
	m.createCalled = true
	m.lastCreatedID = p.ID
	return m.createErr
}

func (m *mockProjectService) GetProject(ctx context.Context, projectID, orgID uuid.UUID) (*models.Project, error) {
	m.getCalled = true
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getProject, nil
}

// --- ProjectHandler Tests ---

func TestProjectHandler_CreateProject_Success(t *testing.T) {
	mock := &mockProjectService{}
	handler := NewProjectHandler(mock)

	orgID := uuid.New()
	body := models.Project{Name: "Test Project", OrgID: orgID}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rr := httptest.NewRecorder()
	handler.CreateProject(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.True(t, mock.createCalled)
}

func TestProjectHandler_CreateProject_MissingOrgHeader(t *testing.T) {
	mock := &mockProjectService{}
	handler := NewProjectHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/projects", nil)
	rr := httptest.NewRecorder()
	handler.CreateProject(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "X-Org-ID header is required")
}

func TestProjectHandler_CreateProject_InvalidOrgID(t *testing.T) {
	mock := &mockProjectService{}
	handler := NewProjectHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/projects", nil)
	req.Header.Set("X-Org-ID", "not-a-uuid")
	rr := httptest.NewRecorder()
	handler.CreateProject(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid X-Org-ID")
}

func TestProjectHandler_CreateProject_OrgMismatch(t *testing.T) {
	mock := &mockProjectService{}
	handler := NewProjectHandler(mock)

	headerOrgID := uuid.New()
	bodyOrgID := uuid.New()
	body := models.Project{Name: "Test Project", OrgID: bodyOrgID}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", headerOrgID.String())
	rr := httptest.NewRecorder()
	handler.CreateProject(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "does not match")
}

func TestProjectHandler_CreateProject_MissingName(t *testing.T) {
	mock := &mockProjectService{}
	handler := NewProjectHandler(mock)

	orgID := uuid.New()
	body := models.Project{OrgID: orgID} // No name
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())
	rr := httptest.NewRecorder()
	handler.CreateProject(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "name is required")
}

func TestProjectHandler_CreateProject_ServiceError(t *testing.T) {
	mock := &mockProjectService{createErr: errors.New("database error")}
	handler := NewProjectHandler(mock)

	orgID := uuid.New()
	body := models.Project{Name: "Test Project", OrgID: orgID}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())
	rr := httptest.NewRecorder()
	handler.CreateProject(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestProjectHandler_CreateProject_Conflict(t *testing.T) {
	mock := &mockProjectService{createErr: errors.New("project already exists")}
	handler := NewProjectHandler(mock)

	orgID := uuid.New()
	body := models.Project{Name: "Test Project", OrgID: orgID}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())
	rr := httptest.NewRecorder()
	handler.CreateProject(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestProjectHandler_GetProject_Success(t *testing.T) {
	projectID := uuid.New()
	orgID := uuid.New()
	mock := &mockProjectService{
		getProject: &models.Project{ID: projectID, OrgID: orgID, Name: "Test"},
	}
	handler := NewProjectHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
	req.Header.Set("X-Org-ID", orgID.String())

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetProject(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, mock.getCalled)
}

func TestProjectHandler_GetProject_NotFound(t *testing.T) {
	mock := &mockProjectService{getErr: errors.New("not found")}
	handler := NewProjectHandler(mock)

	projectID := uuid.New()
	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetProject(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestProjectHandler_GetProject_InvalidID(t *testing.T) {
	mock := &mockProjectService{}
	handler := NewProjectHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/projects/invalid", nil)
	req.Header.Set("X-Org-ID", uuid.New().String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetProject(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProjectHandler_GetProject_MissingOrgHeader(t *testing.T) {
	mock := &mockProjectService{}
	handler := NewProjectHandler(mock)

	projectID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetProject(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- taskHandlerMockScheduleService (for TaskHandler tests) ---
// This is named differently from chat_handler_test.go to avoid redeclaration.

type taskHandlerMockScheduleService struct {
	getTaskResult       *models.ProjectTask
	getTaskErr          error
	updateDurationErr   error
	updateStatusErr     error
	createProgressErr   error
	createInspectionErr error
	recalcResult        *physics.CPMResult
	recalcErr           error

	getTaskCalled          bool
	updateDurationCalled   bool
	updateStatusCalled     bool
	createProgressCalled   bool
	createInspectionCalled bool
	recalcCalled           bool
}

func (m *taskHandlerMockScheduleService) GetTask(ctx context.Context, taskID, projectID, orgID uuid.UUID) (*models.ProjectTask, error) {
	m.getTaskCalled = true
	return m.getTaskResult, m.getTaskErr
}

func (m *taskHandlerMockScheduleService) UpdateTaskDuration(ctx context.Context, taskID, projectID, orgID uuid.UUID, days float64, reason string) error {
	m.updateDurationCalled = true
	return m.updateDurationErr
}

func (m *taskHandlerMockScheduleService) UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status types.TaskStatus) error {
	m.updateStatusCalled = true
	return m.updateStatusErr
}

func (m *taskHandlerMockScheduleService) CreateTaskProgress(ctx context.Context, projectID, taskID, userID uuid.UUID, percent int, notes string) error {
	m.createProgressCalled = true
	return m.createProgressErr
}

func (m *taskHandlerMockScheduleService) CreateInspectionRecord(ctx context.Context, projectID, taskID uuid.UUID, inspector, result, notes string, date time.Time) error {
	m.createInspectionCalled = true
	return m.createInspectionErr
}

func (m *taskHandlerMockScheduleService) RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*physics.CPMResult, error) {
	m.recalcCalled = true
	return m.recalcResult, m.recalcErr
}

// --- TaskHandler Tests ---

func TestTaskHandler_UpdateTask_Success(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID, Name: "Test Task"},
		recalcResult:  &physics.CPMResult{},
	}
	handler := NewTaskHandler(mock)

	overrideDays := 5.0
	body := UpdateTaskRequest{ManualOverrideDays: &overrideDays, OverrideReason: "Weather delay"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID.String(), bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, mock.updateDurationCalled)
	assert.True(t, mock.recalcCalled)
}

func TestTaskHandler_UpdateTask_RecalcFail(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID},
		recalcErr:     errors.New("schedule calculation failed"),
	}
	handler := NewTaskHandler(mock)

	overrideDays := 5.0
	body := UpdateTaskRequest{ManualOverrideDays: &overrideDays}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID.String(), bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "recalculation failed")
}

func TestTaskHandler_RecordProgress_Complete(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID, Status: types.TaskStatusInProgress},
		recalcResult:  &physics.CPMResult{},
	}
	handler := NewTaskHandler(mock)

	body := ProgressRequest{PercentComplete: 100, Notes: "Done"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/progress", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordProgress(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, mock.createProgressCalled)
	assert.True(t, mock.updateStatusCalled)
	assert.True(t, mock.recalcCalled)
}

func TestTaskHandler_RecordProgress_Partial(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID, Status: types.TaskStatusInProgress},
	}
	handler := NewTaskHandler(mock)

	body := ProgressRequest{PercentComplete: 50, Notes: "Halfway"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/progress", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordProgress(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, mock.createProgressCalled)
	assert.False(t, mock.updateStatusCalled) // Not 100%, no status change
	assert.False(t, mock.recalcCalled)
}

func TestTaskHandler_RecordInspection_Pass(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID, IsInspection: true, Status: types.TaskStatusPending},
		recalcResult:  &physics.CPMResult{},
	}
	handler := NewTaskHandler(mock)

	body := InspectionRequest{
		Result:         types.InspectionResultPassed,
		InspectorName:  "John Doe",
		InspectionDate: "2026-01-14",
		Notes:          "All good",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/inspection", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordInspection(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, mock.createInspectionCalled)
	assert.True(t, mock.updateStatusCalled) // Passed triggers completion
	assert.True(t, mock.recalcCalled)
}

func TestTaskHandler_RecordInspection_Fail(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID, IsInspection: true, Status: types.TaskStatusPending},
	}
	handler := NewTaskHandler(mock)

	body := InspectionRequest{
		Result:         types.InspectionResultFailed,
		InspectorName:  "John Doe",
		InspectionDate: "2026-01-14",
		Notes:          "Issues found",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/inspection", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordInspection(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, mock.createInspectionCalled)
	assert.False(t, mock.updateStatusCalled) // Failed does NOT complete
	assert.False(t, mock.recalcCalled)
}

func TestTaskHandler_RecordInspection_InvalidEnum(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID, IsInspection: true},
	}
	handler := NewTaskHandler(mock)

	body := map[string]interface{}{
		"result":          "Maybe",
		"inspector_name":  "John Doe",
		"inspection_date": "2026-01-14",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/inspection", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordInspection(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid inspection result")
}

func TestTaskHandler_RecordInspection_NotInspectionTask(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID, IsInspection: false}, // Not an inspection
	}
	handler := NewTaskHandler(mock)

	body := InspectionRequest{
		Result:         types.InspectionResultPassed,
		InspectorName:  "John Doe",
		InspectionDate: "2026-01-14",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/inspection", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordInspection(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "not an inspection task")
}

func TestTaskHandler_UpdateTask_TaskNotFound(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskErr: errors.New("not found"),
	}
	handler := NewTaskHandler(mock)

	body := UpdateTaskRequest{}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID.String(), bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestTaskHandler_UpdateTask_InvalidBody(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID},
	}
	handler := NewTaskHandler(mock)

	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID.String(), bytes.NewReader([]byte("invalid")))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestTaskHandler_UpdateTask_InvalidTaskID(t *testing.T) {
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{}
	handler := NewTaskHandler(mock)

	req := httptest.NewRequest(http.MethodPut, "/tasks/invalid", bytes.NewReader([]byte("{}")))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestTaskHandler_UpdateTask_NoOverride(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID},
	}
	handler := NewTaskHandler(mock)

	body := UpdateTaskRequest{} // No ManualOverrideDays
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID.String(), bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, mock.updateDurationCalled) // No override, no update
	assert.False(t, mock.recalcCalled)
}

func TestTaskHandler_UpdateTask_UpdateDurationFails(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult:     &models.ProjectTask{ID: taskID},
		updateDurationErr: errors.New("db error"),
	}
	handler := NewTaskHandler(mock)

	overrideDays := 5.0
	body := UpdateTaskRequest{ManualOverrideDays: &overrideDays}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID.String(), bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestTaskHandler_RecordProgress_InvalidBody(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{}
	handler := NewTaskHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/progress", bytes.NewReader([]byte("invalid")))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordProgress(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestTaskHandler_RecordProgress_TaskNotFound(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskErr: errors.New("not found"),
	}
	handler := NewTaskHandler(mock)

	body := ProgressRequest{PercentComplete: 50}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/progress", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordProgress(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestTaskHandler_RecordProgress_CreateProgressFails(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult:     &models.ProjectTask{ID: taskID, Status: types.TaskStatusInProgress},
		createProgressErr: errors.New("db error"),
	}
	handler := NewTaskHandler(mock)

	body := ProgressRequest{PercentComplete: 50}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/progress", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordProgress(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestTaskHandler_RecordInspection_InvalidDate(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult: &models.ProjectTask{ID: taskID, IsInspection: true},
	}
	handler := NewTaskHandler(mock)

	body := map[string]interface{}{
		"result":          "Passed",
		"inspector_name":  "John Doe",
		"inspection_date": "not-a-date", // Invalid
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/inspection", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordInspection(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "YYYY-MM-DD")
}

func TestTaskHandler_RecordInspection_CreateRecordFails(t *testing.T) {
	taskID := uuid.New()
	projectID := uuid.New()
	orgID := uuid.New()

	mock := &taskHandlerMockScheduleService{
		getTaskResult:       &models.ProjectTask{ID: taskID, IsInspection: true},
		createInspectionErr: errors.New("db error"),
	}
	handler := NewTaskHandler(mock)

	body := InspectionRequest{
		Result:         types.InspectionResultPassed,
		InspectorName:  "John Doe",
		InspectionDate: "2026-01-14",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID.String()+"/inspection", bytes.NewReader(jsonBody))
	req.Header.Set("X-Org-ID", orgID.String())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	rctx.URLParams.Add("task_id", taskID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RecordInspection(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

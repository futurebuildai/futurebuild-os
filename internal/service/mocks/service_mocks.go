package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// MockProjectService is a spy for ProjectServicer.
type MockProjectService struct {
	mu sync.Mutex

	CreateProjectErr error
	GetProjectResp   *models.Project
	GetProjectErr    error

	CreateProjectCalls []*models.Project
	GetProjectCalls    []struct{ OrgID, ProjectID uuid.UUID }
}

func (m *MockProjectService) CreateProject(ctx context.Context, p *models.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateProjectCalls = append(m.CreateProjectCalls, p)
	return m.CreateProjectErr
}

func (m *MockProjectService) GetProject(ctx context.Context, id, orgID uuid.UUID) (*models.Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetProjectCalls = append(m.GetProjectCalls, struct{ OrgID, ProjectID uuid.UUID }{orgID, id})
	return m.GetProjectResp, m.GetProjectErr
}

func (m *MockProjectService) StreamActiveProjects(ctx context.Context, process service.ProjectProcessor) error {
	// Simple mock: assumes we don't need to stream in tests most of the time
	// or we can iterate over a predefined list if needed.
	return nil
}

// MockScheduleService is a spy for ScheduleServicer.
type MockScheduleService struct {
	mu sync.Mutex

	GetScheduleResp *service.ProjectScheduleSummary
	GetScheduleErr  error

	GetAgentFocusTasksResp []models.ProjectTask
	GetAgentFocusTasksErr  error

	RecalculateResp *physics.CPMResult
	RecalculateErr  error

	GetTaskResp *models.ProjectTask
	GetTaskErr  error

	UpdateDurationErr   error
	UpdateStatusErr     error
	CreateProgressErr   error
	CreateInspectionErr error

	// Call recording for assertions
	UpdateDurationCalls []struct {
		TaskID, ProjectID, OrgID uuid.UUID
		Days                     float64
		Reason                   string
	}
	UpdateStatusCalls []struct {
		TaskID, ProjectID, OrgID uuid.UUID
		Status                   types.TaskStatus
	}
	CreateProgressCalls []struct {
		ProjectID, TaskID, UserID uuid.UUID
		Percent                   int
		Notes                     string
	}
	CreateInspectionCalls []struct {
		ProjectID, TaskID        uuid.UUID
		Inspector, Result, Notes string
		Date                     time.Time
	}
}

func (m *MockScheduleService) GetProjectSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*service.ProjectScheduleSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.GetScheduleResp, m.GetScheduleErr
}

func (m *MockScheduleService) GetAgentFocusTasks(ctx context.Context, projectID uuid.UUID) ([]models.ProjectTask, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.GetAgentFocusTasksResp, m.GetAgentFocusTasksErr
}

func (m *MockScheduleService) RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*physics.CPMResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.RecalculateResp, m.RecalculateErr
}

func (m *MockScheduleService) GetTask(ctx context.Context, taskID, projectID, orgID uuid.UUID) (*models.ProjectTask, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.GetTaskResp, m.GetTaskErr
}

func (m *MockScheduleService) UpdateTaskDuration(ctx context.Context, taskID, projectID, orgID uuid.UUID, overrideDays float64, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateDurationCalls = append(m.UpdateDurationCalls, struct {
		TaskID, ProjectID, OrgID uuid.UUID
		Days                     float64
		Reason                   string
	}{taskID, projectID, orgID, overrideDays, reason})
	return m.UpdateDurationErr
}

func (m *MockScheduleService) UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status types.TaskStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateStatusCalls = append(m.UpdateStatusCalls, struct {
		TaskID, ProjectID, OrgID uuid.UUID
		Status                   types.TaskStatus
	}{taskID, projectID, orgID, status})
	return m.UpdateStatusErr
}

func (m *MockScheduleService) CreateTaskProgress(ctx context.Context, projectID, taskID, userID uuid.UUID, percentComplete int, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateProgressCalls = append(m.CreateProgressCalls, struct {
		ProjectID, TaskID, UserID uuid.UUID
		Percent                   int
		Notes                     string
	}{projectID, taskID, userID, percentComplete, notes})
	return m.CreateProgressErr
}

func (m *MockScheduleService) CreateInspectionRecord(ctx context.Context, projectID, taskID uuid.UUID, inspectorName, result, notes string, inspectionDate time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateInspectionCalls = append(m.CreateInspectionCalls, struct {
		ProjectID, TaskID        uuid.UUID
		Inspector, Result, Notes string
		Date                     time.Time
	}{projectID, taskID, inspectorName, result, notes, inspectionDate})
	return m.CreateInspectionErr
}

// MockInvoiceService is a spy for InvoiceServicer.
type MockInvoiceService struct {
	mu              sync.Mutex
	AnalyzeRespUUID uuid.UUID
	AnalyzeResp     *types.InvoiceExtraction
	AnalyzeErr      error
	SaveRespUUID    uuid.UUID
	SaveErr         error

	AnalyzeCalls []struct{ OrgID, DocID uuid.UUID }
	SaveCalls    []struct {
		ProjectID   uuid.UUID
		SourceDocID *uuid.UUID
	}
}

func (m *MockInvoiceService) AnalyzeInvoice(ctx context.Context, orgID, docID uuid.UUID) (uuid.UUID, *types.InvoiceExtraction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AnalyzeCalls = append(m.AnalyzeCalls, struct{ OrgID, DocID uuid.UUID }{orgID, docID})
	return m.AnalyzeRespUUID, m.AnalyzeResp, m.AnalyzeErr
}

func (m *MockInvoiceService) SaveExtraction(ctx context.Context, projectID uuid.UUID, extraction *types.InvoiceExtraction, sourceDocID *uuid.UUID) (uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SaveCalls = append(m.SaveCalls, struct {
		ProjectID   uuid.UUID
		SourceDocID *uuid.UUID
	}{projectID, sourceDocID})
	return m.SaveRespUUID, m.SaveErr
}

// MockWeatherService is a spy for WeatherServicer.
type MockWeatherService struct {
	mu           sync.Mutex
	ForecastResp types.Forecast
	ForecastErr  error
	Calls        []struct{ Lat, Long float64 }
}

func (m *MockWeatherService) GetForecast(lat, long float64) (types.Forecast, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, struct{ Lat, Long float64 }{lat, long})
	return m.ForecastResp, m.ForecastErr
}

// MockDirectoryService is a spy for DirectoryServicer.
type MockDirectoryService struct {
	mu          sync.Mutex
	ContactResp *types.Contact
	ContactErr  error

	GetContactCalls []struct {
		ProjectID, OrgID uuid.UUID
		PhaseCode        string
	}
	GetPMCalls []struct{ ProjectID, OrgID uuid.UUID }
}

func (m *MockDirectoryService) GetContactForPhase(ctx context.Context, projectID, orgID uuid.UUID, phaseCode string) (*types.Contact, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetContactCalls = append(m.GetContactCalls, struct {
		ProjectID, OrgID uuid.UUID
		PhaseCode        string
	}{projectID, orgID, phaseCode})
	return m.ContactResp, m.ContactErr
}

func (m *MockDirectoryService) GetProjectManager(ctx context.Context, projectID, orgID uuid.UUID) (*types.Contact, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetPMCalls = append(m.GetPMCalls, struct{ ProjectID, OrgID uuid.UUID }{projectID, orgID})
	return m.ContactResp, m.ContactErr
}

// MockNotificationService is a spy.
type MockNotificationService struct {
	mu         sync.Mutex
	SentEmails []struct{ To, Subject, Body string }
	SentSMS    []struct{ ContactID, Message string }
	SendErr    error
}

func (m *MockNotificationService) SendSMS(contactID string, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentSMS = append(m.SentSMS, struct{ ContactID, Message string }{contactID, message})
	return m.SendErr
}

func (m *MockNotificationService) SendEmail(to string, subject string, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentEmails = append(m.SentEmails, struct{ To, Subject, Body string }{to, subject, body})
	return m.SendErr
}

// MockVisionService is a spy.
type MockVisionService struct {
	mu          sync.Mutex
	VerifyResp  bool
	VerifyScore float64
	VerifyErr   error
	Calls       []struct{ ImageURL, TaskDesc string }
}

func (m *MockVisionService) VerifyTask(ctx context.Context, imageURL string, taskDescription string) (bool, float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, struct{ ImageURL, TaskDesc string }{imageURL, taskDescription})
	return m.VerifyResp, m.VerifyScore, m.VerifyErr
}

// MockDocumentService is a spy.
type MockDocumentService struct {
	mu           sync.Mutex
	IngestErr    error
	ReprocessErr error
	StatusResp   string
	StatusCount  int
	StatusErr    error

	IngestCalls    []uuid.UUID
	ReprocessCalls []struct{ OrgID, DocID uuid.UUID }
	StatusCalls    []uuid.UUID
}

func (m *MockDocumentService) GetDocumentStatus(ctx context.Context, docID uuid.UUID) (string, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StatusCalls = append(m.StatusCalls, docID)
	return m.StatusResp, m.StatusCount, m.StatusErr
}

func (m *MockDocumentService) IngestDocument(ctx context.Context, docID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IngestCalls = append(m.IngestCalls, docID)
	return m.IngestErr
}

func (m *MockDocumentService) ReprocessDocument(ctx context.Context, orgID, docID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ReprocessCalls = append(m.ReprocessCalls, struct{ OrgID, DocID uuid.UUID }{orgID, docID})
	return m.ReprocessErr
}

// --- Compile-time interface assertions ---
// These ensure mocks satisfy their interfaces at compile time, not runtime.
var (
	_ service.ProjectServicer      = (*MockProjectService)(nil)
	_ service.ScheduleServicer     = (*MockScheduleService)(nil)
	_ service.InvoiceServicer      = (*MockInvoiceService)(nil)
	_ service.WeatherServicer      = (*MockWeatherService)(nil)
	_ service.DirectoryServicer    = (*MockDirectoryService)(nil)
	_ service.NotificationServicer = (*MockNotificationService)(nil)
	_ service.VisionServicer       = (*MockVisionService)(nil)
	_ service.DocumentServicer     = (*MockDocumentService)(nil)
)

// Helper for formatting call records in failure messages
func formatCalls(calls interface{}) string {
	return fmt.Sprintf("%+v", calls)
}

package models

import (
	"time"

	"github.com/google/uuid"
)

// AnalysisStatus represents the vision analysis status of a project asset.
// See STEP_84_FIELD_FEEDBACK.md Section 2.2
type AnalysisStatus string

const (
	AnalysisStatusProcessing AnalysisStatus = "processing"
	AnalysisStatusCompleted  AnalysisStatus = "completed"
	AnalysisStatusFailed     AnalysisStatus = "failed"
)

// ProjectAsset represents an uploaded file (photo) linked to a project.
// See STEP_84_FIELD_FEEDBACK.md Section 2
type ProjectAsset struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	ProjectID      uuid.UUID       `json:"project_id" db:"project_id"`
	TaskID         *uuid.UUID      `json:"task_id,omitempty" db:"task_id"`
	UploadedBy     string          `json:"uploaded_by" db:"uploaded_by"`
	FileName       string          `json:"file_name" db:"file_name"`
	FileURL        string          `json:"file_url" db:"file_url"`
	MimeType       string          `json:"mime_type" db:"mime_type"`
	FileSizeBytes  int64           `json:"file_size_bytes" db:"file_size_bytes"`
	AnalysisStatus AnalysisStatus  `json:"analysis_status" db:"analysis_status"`
	AnalysisResult *map[string]any `json:"analysis_result,omitempty" db:"analysis_result"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

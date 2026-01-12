package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// DocumentType defines the type of uploaded document
type DocumentType string

const (
	DocumentTypeBlueprint       DocumentType = "blueprint"
	DocumentTypePermit          DocumentType = "permit"
	DocumentTypeContract        DocumentType = "contract"
	DocumentTypeWorkOrder       DocumentType = "work_order"
	DocumentTypeDesignSelection DocumentType = "design_selection"
	DocumentTypeInvoice         DocumentType = "invoice"
	DocumentTypeSitePhoto       DocumentType = "site_photo"
	DocumentTypeSiteVideo       DocumentType = "site_video"
	DocumentTypeOther           DocumentType = "other"
)

// ProcessingStatus tracks the RAG/Vision processing state
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

// Document represents a file stored in Object Storage
type Document struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	ProjectID        uuid.UUID        `json:"project_id" db:"project_id"`
	Type             DocumentType     `json:"type" db:"type"`
	Filename         string           `json:"filename" db:"filename"`
	StoragePath      string           `json:"storage_path" db:"storage_path"`
	MimeType         string           `json:"mime_type" db:"mime_type"`
	FileSizeBytes    int64            `json:"file_size_bytes" db:"file_size_bytes"`
	ProcessingStatus ProcessingStatus `json:"processing_status" db:"processing_status"`
	ExtractedText    string           `json:"extracted_text" db:"extracted_text"`
	Metadata         map[string]any   `json:"metadata" db:"metadata"`
	UploadedBy       uuid.UUID        `json:"uploaded_by" db:"uploaded_by"`
	UploadedAt       time.Time        `json:"uploaded_at" db:"uploaded_at"`
}

// DocumentChunk represents a semantic chunk of a document for RAG
type DocumentChunk struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	DocumentID   uuid.UUID       `json:"document_id" db:"document_id"`
	ChunkIndex   int             `json:"chunk_index" db:"chunk_index"`
	ChunkContent string          `json:"chunk_content" db:"chunk_content"`
	Embedding    pgvector.Vector `json:"embedding" db:"embedding"` // Vector(768)
	Metadata     map[string]any  `json:"metadata" db:"metadata"`
}

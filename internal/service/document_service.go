package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/colton/futurebuild/internal/rag"
	"github.com/colton/futurebuild/pkg/ai"
)

// DocumentService handles document operations including RAG ingestion
type DocumentService struct {
	db       *pgxpool.Pool
	embedder *rag.Embedder
	chunker  *rag.Chunker
}

// NewDocumentService creates a new DocumentService
func NewDocumentService(db *pgxpool.Pool, client ai.Client) *DocumentService {
	return &DocumentService{
		db:       db,
		embedder: rag.NewEmbedder(client),
		chunker:  rag.NewChunker(),
	}
}

// IngestDocument processes a document for RAG: Read -> Chunk -> Embed -> Save
func (s *DocumentService) IngestDocument(ctx context.Context, docID uuid.UUID) error {
	// 1. Fetch Document Content
	var extractedText string
	var processingStatus string

	query := `SELECT extracted_text, processing_status FROM documents WHERE id = $1`
	err := s.db.QueryRow(ctx, query, docID).Scan(&extractedText, &processingStatus)
	if err != nil {
		return fmt.Errorf("failed to fetch document: %w", err)
	}

	if extractedText == "" {
		return fmt.Errorf("document has no extracted text to process")
	}

	// 2. Chunk Text
	chunks := s.chunker.ChunkDocument(extractedText)
	if len(chunks) == 0 {
		return fmt.Errorf("no chunks generated from document")
	}

	// 3. Generate Embeddings
	embeddings, err := s.embedder.GenerateEmbeddings(ctx, chunks)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if len(chunks) != len(embeddings) {
		return fmt.Errorf("mismatch between chunks count (%d) and embeddings count (%d)", len(chunks), len(embeddings))
	}

	// 4. Save Chunks to DB
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// First, delete any existing chunks for this doc to allow re-ingestion
	_, err = tx.Exec(ctx, "DELETE FROM document_chunks WHERE document_id = $1", docID)
	if err != nil {
		return fmt.Errorf("failed to cleanup old chunks: %w", err)
	}

	insertQuery := `
		INSERT INTO document_chunks (document_id, chunk_index, chunk_content, embedding)
		VALUES ($1, $2, $3, $4)
	`

	for i, chunk := range chunks {
		embedding := embeddings[i]

		// Use pgvector.NewVector to correctly format the embedding for pgx
		_, err := tx.Exec(ctx, insertQuery, docID, i, chunk, pgvector.NewVector(embedding))
		if err != nil {
			return fmt.Errorf("failed to insert chunk %d: %w", i, err)
		}
	}

	// 5. Update Status
	_, err = tx.Exec(ctx, "UPDATE documents SET processing_status = 'completed' WHERE id = $1", docID)
	if err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	return tx.Commit(ctx)
}

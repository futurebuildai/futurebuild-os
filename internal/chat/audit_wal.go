// Package chat provides Write-Ahead Log (WAL) for audit trail durability.
// See PRODUCTION_PLAN.md Audit Trail Durability Remediation.
package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/colton/futurebuild/internal/models"
)

// AuditWAL defines the interface for local audit record fallback.
// Used when both primary DB and DLQ are unavailable.
// See Audit Trail Durability remediation.
type AuditWAL interface {
	// AppendRecord writes a chat message to the local WAL.
	// Returns nil on success, error on failure.
	AppendRecord(ctx context.Context, msg models.ChatMessage) error

	// Flush forces any buffered data to disk.
	Flush() error

	// Close releases resources and flushes pending writes.
	Close() error
}

// FileAuditWAL implements AuditWAL using append-only local file writes.
// Uses newline-delimited JSON for easy parsing by sidecar scraper.
// Thread-safe for concurrent writes.
type FileAuditWAL struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

// NewFileAuditWAL creates a new file-based WAL at the specified path.
// Creates the parent directory and file if they don't exist, opens in append mode if file exists.
// Returns error if file cannot be created or opened.
func NewFileAuditWAL(path string) (*FileAuditWAL, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory %s: %w", dir, err)
	}

	// Open file in append-only mode with create flag
	// O_SYNC ensures immediate write to disk (durability)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file: %w", err)
	}

	return &FileAuditWAL{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// AppendRecord writes a chat message to the WAL in newline-delimited JSON.
// Thread-safe and immediately flushes to disk for durability.
func (w *FileAuditWAL) AppendRecord(ctx context.Context, msg models.ChatMessage) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message for WAL: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Write JSON followed by newline
	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write to WAL: %w", err)
	}
	if err := w.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to write newline to WAL: %w", err)
	}

	// Flush to ensure durability
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush WAL: %w", err)
	}

	slog.Debug("audit record written to WAL",
		"message_id", msg.ID,
		"project_id", msg.ProjectID,
	)

	return nil
}

// Flush forces any buffered data to disk.
func (w *FileAuditWAL) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Flush()
}

// Close flushes pending writes and closes the file.
func (w *FileAuditWAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush on close: %w", err)
	}
	return w.file.Close()
}

// NoOpAuditWAL is a no-op implementation for testing or when WAL is disabled.
type NoOpAuditWAL struct{}

// AppendRecord does nothing and returns nil.
func (w *NoOpAuditWAL) AppendRecord(_ context.Context, _ models.ChatMessage) error {
	return nil
}

// Flush does nothing and returns nil.
func (w *NoOpAuditWAL) Flush() error {
	return nil
}

// Close does nothing and returns nil.
func (w *NoOpAuditWAL) Close() error {
	return nil
}

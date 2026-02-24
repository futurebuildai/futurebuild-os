package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentDecisionEntry represents a single AI agent decision for auditing and tracing.
type AgentDecisionEntry struct {
	ID           string         `json:"id"`
	Timestamp    time.Time      `json:"timestamp"`
	Agent        string         `json:"agent"`
	Action       string         `json:"action"`
	InputSummary string         `json:"input_summary"`
	Decision     string         `json:"decision"`
	Confidence   float64        `json:"confidence"`
	Model        string         `json:"model"`
	LatencyMS    int64          `json:"latency_ms"`
	ProjectID    string         `json:"project_id"`
	UserID       string         `json:"user_id"`
	TraceID      string         `json:"trace_id"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// AgentLogger describes the interface for logging AI agent decisions.
type AgentLogger interface {
	LogDecision(ctx context.Context, entry AgentDecisionEntry) error
}

// PgxAgentLogger implements AgentLogger using a PostgreSQL database pool.
type PgxAgentLogger struct {
	db *pgxpool.Pool
}

// NewPgxAgentLogger creates a new PgxAgentLogger.
func NewPgxAgentLogger(db *pgxpool.Pool) *PgxAgentLogger {
	return &PgxAgentLogger{db: db}
}

// LogDecision persists the agent decision asynchronously or synchronously depending on context.
// In this implementation, it performs a synchronous insert but can easily be sent to a channel.
func (l *PgxAgentLogger) LogDecision(ctx context.Context, entry AgentDecisionEntry) error {
	var metadataBytes []byte
	if entry.Metadata != nil {
		b, err := json.Marshal(entry.Metadata)
		if err != nil {
			slog.Warn("failed to marshal agent decision metadata", "error", err)
			b = []byte("{}")
		}
		metadataBytes = b
	}

	query := `
		INSERT INTO audit_decisions (
			timestamp, agent, action, input_summary, decision,
			confidence, model, latency_ms, project_id, user_id,
			trace_id, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	_, err := l.db.Exec(ctx, query,
		entry.Timestamp,
		entry.Agent,
		entry.Action,
		entry.InputSummary,
		entry.Decision,
		entry.Confidence,
		entry.Model,
		entry.LatencyMS,
		entry.ProjectID,
		entry.UserID,
		entry.TraceID,
		metadataBytes,
	)

	if err != nil {
		slog.Error("failed to record agent decision to audit_decisions", "error", err, "trace_id", entry.TraceID)
		return fmt.Errorf("audit insertion failed: %w", err)
	}

	return nil
}

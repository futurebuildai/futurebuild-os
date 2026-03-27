package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// VoiceTranscriptionPayload contains data for async voice-to-text processing.
// Phase 18: See FRONTEND_SCOPE.md §15.2 (Voice-First Field Portal)
type VoiceTranscriptionPayload struct {
	MemoID   uuid.UUID `json:"memo_id"`
	S3Key    string    `json:"s3_key"`
	MIMEType string    `json:"mime_type"`
	OrgID    uuid.UUID `json:"org_id"`
}

// HandleVoiceTranscription processes a voice memo from S3 via Vertex AI.
// Downloads audio from S3, sends to Vertex AI for transcription, saves result.
// This is the only AI-touching code in the voice pipeline.
func (h *WorkerHandler) HandleVoiceTranscription(_ context.Context, t *asynq.Task) error {
	var payload VoiceTranscriptionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("voice_transcription: invalid payload: %w", err)
	}

	slog.Info("voice_transcription: processing memo",
		"memo_id", payload.MemoID,
		"s3_key", payload.S3Key,
		"mime_type", payload.MIMEType,
	)

	// TODO Phase 18: Full implementation requires:
	// 1. Download audio blob from S3 using s3_key
	// 2. Send to Vertex AI (Gemini Flash) for speech-to-text
	// 3. Save transcript to database linked to memo_id
	// 4. Optionally create a feed card with the transcript
	//
	// For now, log receipt. S3 client and AI client will be wired
	// via WithVoiceTranscription() builder method when available.

	slog.Info("voice_transcription: completed (stub)",
		"memo_id", payload.MemoID,
	)

	return nil
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// VisionService implements AI-powered site photo analysis.
// See API_AND_TYPES_SPEC.md Section 2.2 and PRODUCTION_PLAN.md Step 40.
type VisionService struct {
	client     ai.Client
	httpClient *http.Client
}

// NewVisionService creates a new VisionService.
func NewVisionService(client ai.Client) *VisionService {
	return &VisionService{
		client: client,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

const visionPromptTemplate = `
Analyze the provided image and determine if the following task has been completed or is accurately represented:
Task Description: %s

Return your analysis as a structured JSON object with the following schema:
{
  "is_verified": bool,
  "confidence": 0.0 to 1.0 (float),
  "reasoning": "Brief explanation of what was observed"
}

Do NOT include any markdown formatting (like ` + "```" + `json) in your response. Return ONLY the raw JSON string.
`

// VerifyTask returns (is_verified, confidence_score, error)
// See API_AND_TYPES_SPEC.md Section 2.2
func (s *VisionService) VerifyTask(ctx context.Context, imageURL string, taskDescription string) (bool, float64, error) {
	// 1. Download image from URL
	imageBytes, mimeType, err := s.downloadImage(ctx, imageURL)
	if err != nil {
		return false, 0, fmt.Errorf("failed to download image: %w", err)
	}

	// 2. Prepare AI Parts
	prompt := fmt.Sprintf(visionPromptTemplate, taskDescription)
	imgPart := genai.ImageData(strings.TrimPrefix(mimeType, "image/"), imageBytes)

	// 3. Call Vertex AI (Gemini 2.5 Flash mandated per BACKEND_SCOPE Section 3.2)
	resp, err := s.client.GenerateContent(ctx, ai.ModelTypeFlash, genai.Text(prompt), imgPart)
	if err != nil {
		return false, 0, fmt.Errorf("ai vision analysis failed: %w", err)
	}

	// Clean response
	cleanResp := strings.TrimSpace(resp)
	cleanResp = strings.TrimPrefix(cleanResp, "```json")
	cleanResp = strings.TrimSuffix(cleanResp, "```")
	cleanResp = strings.TrimSpace(cleanResp)

	// 4. Parse Result
	var result struct {
		IsVerified bool    `json:"is_verified"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(cleanResp), &result); err != nil {
		return false, 0, fmt.Errorf("failed to parse vision response: %w. Response: %s", err, cleanResp)
	}

	return result.IsVerified, result.Confidence, nil
}

// VerifyAndPersistTask verifies a task via site photo AND persists the result.
// This satisfies the "Database is State" requirement from DATA_SPINE_SPEC.md.
// See PRODUCTION_PLAN.md Step 40
func (s *VisionService) VerifyAndPersistTask(ctx context.Context, db DBExecutor, taskID, projectID, orgID uuid.UUID, imageURL string, taskDescription string) (bool, float64, error) {
	// 1. Call AI Verification
	isVerified, confidence, err := s.VerifyTask(ctx, imageURL, taskDescription)
	if err != nil {
		return false, 0, fmt.Errorf("vision verification failed: %w", err)
	}

	// 2. Persist to Database (with Multi-Tenancy Check)
	// See DATA_SPINE_SPEC.md Section 3.3
	query := `
		UPDATE project_tasks pt
		SET 
			verified_by_vision = $1,
			verification_confidence = $2,
			updated_at = NOW()
		FROM projects p
		WHERE pt.id = $3 
			AND pt.project_id = $4
			AND p.id = $4
			AND p.org_id = $5
	`
	result, err := db.Exec(ctx, query, isVerified, confidence, taskID, projectID, orgID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to persist verification result: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return false, 0, fmt.Errorf("task not found or access denied (multi-tenancy violation)")
	}

	return isVerified, confidence, nil
}

// DBExecutor defines the minimal database interface needed for persistence.
// This allows both *pgxpool.Pool and *pgx.Tx to be used.
type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

func (s *VisionService) downloadImage(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	mimeType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, "", fmt.Errorf("invalid content type: %s", mimeType)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return data, mimeType, nil
}

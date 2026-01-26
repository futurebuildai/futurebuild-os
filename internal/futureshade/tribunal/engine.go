package tribunal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// ConsensusEngine orchestrates the multi-model decision process.
type ConsensusEngine struct {
	jury Jury
	repo *Repository
}

// NewConsensusEngine creates a new engine with the given jury and storage.
func NewConsensusEngine(jury Jury, repo *Repository) *ConsensusEngine {
	return &ConsensusEngine{
		jury: jury,
		repo: repo,
	}
}

// Review processes a request through the Tribunal.
func (e *ConsensusEngine) Review(ctx context.Context, req TribunalRequest) (*TribunalResponse, error) {
	// Create Decision Record
	decisionID := uuid.New()

	// parallelize expert opinions
	type opinionResult struct {
		role string
		vote ModelVote
		err  error
	}

	resultChan := make(chan opinionResult, 2)
	var wg sync.WaitGroup

	// The Architect (Claude)
	wg.Add(1)
	go func() {
		defer wg.Done()
		vote, err := e.consultExpert(ctx, e.jury.Architect, ai.ModelTypeOpus, "The Architect", ArchitectSystemPrompt, req)
		resultChan <- opinionResult{role: "architect", vote: vote, err: err}
	}()

	// The Historian (Gemini)
	wg.Add(1)
	go func() {
		defer wg.Done()
		vote, err := e.consultExpert(ctx, e.jury.Historian, ai.ModelTypeCodeAssist, "The Historian", HistorianSystemPrompt, req)
		resultChan <- opinionResult{role: "historian", vote: vote, err: err}
	}()

	wg.Wait()
	close(resultChan)

	var votes []ModelVote
	var expertContext strings.Builder

	expertContext.WriteString(fmt.Sprintf("Intent: %s\nContext: %s\n\n", req.Intent, req.Context))

	for res := range resultChan {
		if res.err != nil {
			// Log error but continue (fail partial?)
			// For L7 strictness, we might want to fail hard, but for resilience we log.
			// Let's create an ABSTAIN vote for the error case.
			votes = append(votes, ModelVote{
				DecisionID: decisionID,
				ModelName:  res.role,
				Vote:       VoteAbstain,
				Reasoning:  fmt.Sprintf("Error consulting expert: %v", res.err),
			})
			expertContext.WriteString(fmt.Sprintf("Expert %s failed: %v\n\n", res.role, res.err))
		} else {
			res.vote.DecisionID = decisionID
			votes = append(votes, res.vote)
			expertContext.WriteString(fmt.Sprintf("Expert %s voted %s:\n%s\n\n", res.role, res.vote.Vote, res.vote.Reasoning))
		}
	}

	// Coordinator Synthesis (Gemini Flash)
	coordinatorReq := ai.NewTextRequest(ai.ModelTypeFlashPreview,
		fmt.Sprintf("%s\n\n---\n\n%s", CoordinatorSystemPrompt, expertContext.String()))

	coordResp, err := e.jury.Coordinator.GenerateContent(ctx, coordinatorReq)
	if err != nil {
		return nil, fmt.Errorf("coordinator failed: %w", err)
	}

	// Parse Coordinator JSON Output
	// Note: In production, use structured output mode if available, or robust JSON parsing.
	// Here we assume standard JSON response.
	var finalVerdict struct {
		Status         DecisionStatus `json:"status"`
		ConsensusScore float64        `json:"consensus_score"`
		Summary        string         `json:"summary"`
		Plan           string         `json:"plan"`
	}

	// Strip markdown code blocks if present (basic sanitization)
	cleanText := strings.TrimPrefix(coordResp.Text, "```json")
	cleanText = strings.TrimSuffix(cleanText, "```")
	cleanText = strings.TrimSpace(cleanText)

	// P1 Fix: Default to REJECTED (Fail-Closed) if JSON parsing fails
	finalVerdict.Status = DecisionRejected
	finalVerdict.Summary = "Coordinator produced invalid JSON or failed to reason. Defaulting to REJECTED."
	finalVerdict.ConsensusScore = 0.0

	if err := json.Unmarshal([]byte(cleanText), &finalVerdict); err != nil {
		// Log error (implicitly done by defaulting above)
	}

	// Persist Everything
	// 1. Save Decision
	if err := e.saveDecision(ctx, decisionID, req, finalVerdict.Status, finalVerdict.ConsensusScore, finalVerdict.Summary); err != nil {
		return nil, err
	}

	// 2. Save Votes
	for _, v := range votes {
		if err := e.saveVote(ctx, v); err != nil {
			// Log error?
		}
	}

	return &TribunalResponse{
		DecisionID:     decisionID,
		Status:         finalVerdict.Status,
		ConsensusScore: finalVerdict.ConsensusScore,
		Summary:        finalVerdict.Summary,
		Plan:           finalVerdict.Plan,
	}, nil
}

func (e *ConsensusEngine) consultExpert(ctx context.Context, client ai.Client, model ai.ModelType, name, systemPrompt string, req TribunalRequest) (ModelVote, error) {
	// P1 Fix: Sanitize Input (Prompt Injection Vector)
	// Simple sanitization: remove potential delimiter abuse
	cleanIntent := strings.ReplaceAll(req.Intent, "---", " ")
	cleanContext := strings.ReplaceAll(req.Context, "---", " ")

	prompt := fmt.Sprintf("%s\n\nInput Intent: %s\nContext: %s", systemPrompt, cleanIntent, cleanContext)
	aiReq := ai.NewTextRequest(model, prompt)

	// P2 Fix: Enforce Timeout for specific expert calls
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := client.GenerateContent(ctxWithTimeout, aiReq)
	duration := time.Since(start)

	if err != nil {
		return ModelVote{}, err
	}

	// Parse Vote
	// P2 Fix: More robust parsing (case insensitive, loose matching)
	vote := VoteAbstain
	normalizedText := strings.ToUpper(resp.Text)
	if strings.Contains(normalizedText, "VOTE: YEA") || strings.Contains(normalizedText, "[VOTE]: YEA") || strings.Contains(normalizedText, "VOTE:YEA") {
		vote = VoteYea
	} else if strings.Contains(normalizedText, "VOTE: NAY") || strings.Contains(normalizedText, "[VOTE]: NAY") || strings.Contains(normalizedText, "VOTE:NAY") {
		vote = VoteNay
	}

	return ModelVote{
		ModelName: name,
		Vote:      vote,
		Reasoning: resp.Text,
		LatencyMs: int(duration.Milliseconds()),
		CostUSD:   0.0, // TODO: Implement cost calculation based on tokens
	}, nil
}

func (e *ConsensusEngine) saveDecision(ctx context.Context, id uuid.UUID, req TribunalRequest, status DecisionStatus, score float64, summary string) error {
	// P0 Fix: Wire up repository call
	return e.repo.CreateDecision(ctx, id, req, status, score, summary)
}

func (e *ConsensusEngine) saveVote(ctx context.Context, v ModelVote) error {
	// P0 Fix: Wire up repository call
	return e.repo.CreateVote(ctx, v)
}

// Diagnose performs self-healing analysis on a runtime error.
// Uses Gemini Flash with Temperature=0 for deterministic output.
// See Tree Planting integration test for usage.
func (e *ConsensusEngine) Diagnose(ctx context.Context, req DiagnosisRequest) (*DiagnosisResponse, error) {
	sessionID := uuid.New()
	start := time.Now()

	// Build the diagnostic prompt
	var stateStr strings.Builder
	for k, v := range req.SystemState {
		stateStr.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
	}

	prompt := fmt.Sprintf(`%s

---
Error Trace:
%s

Method Context: %s

System State:
%s
---

Output the JSON diagnosis:`, DiagnosticianSystemPrompt, req.ErrorTrace, req.MethodContext, stateStr.String())

	// Use Gemini Flash with Temperature=0 for determinism
	aiReq := ai.GenerateRequest{
		Model:       ai.ModelTypeFlashPreview,
		Parts:       []ai.ContentPart{{Text: prompt}},
		Temperature: 0.0, // Deterministic output
	}

	// Enforce timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := e.jury.Coordinator.GenerateContent(ctxWithTimeout, aiReq)
	if err != nil {
		return nil, fmt.Errorf("diagnosis failed: %w", err)
	}

	duration := time.Since(start)

	// Parse the JSON response
	cleanText := strings.TrimPrefix(resp.Text, "```json")
	cleanText = strings.TrimSuffix(cleanText, "```")
	cleanText = strings.TrimSpace(cleanText)

	var decision types.TribunalDecision
	if err := json.Unmarshal([]byte(cleanText), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse diagnosis JSON: %w (raw: %s)", err, cleanText)
	}

	// Validate the proposed action type
	if !types.IsValidActionType(string(decision.ProposedAction.Type)) {
		return nil, fmt.Errorf("invalid action type proposed: %s", decision.ProposedAction.Type)
	}

	return &DiagnosisResponse{
		Decision:  decision,
		SessionID: sessionID,
		LatencyMs: int(duration.Milliseconds()),
		ModelUsed: "gemini-flash",
	}, nil
}

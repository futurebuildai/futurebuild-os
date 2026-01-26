package tribunal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides data access for Tribunal decisions.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new Tribunal repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ListDecisions returns paginated tribunal decisions with optional filtering.
// See SHADOW_VIEWER_specs.md Section 3.1 GET /api/v1/tribunal/decisions
func (r *Repository) ListDecisions(ctx context.Context, filter ListDecisionsFilter) (*ListDecisionsResponse, error) {
	// Set defaults
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Build query with conditions
	baseQuery := `
		SELECT
			d.id,
			d.case_id,
			d.status,
			d.context_summary,
			d.created_at,
			ARRAY_AGG(DISTINCT v.model_name) FILTER (WHERE v.model_name IS NOT NULL) as models_consulted
		FROM tribunal_decisions d
		LEFT JOIN tribunal_votes v ON d.id = v.decision_id
	`

	countQuery := `SELECT COUNT(*) FROM tribunal_decisions d`

	var conditions []string
	var args []interface{}
	argNum := 1

	// Add filters
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("d.status = $%d", argNum))
		args = append(args, filter.Status)
		argNum++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("d.created_at >= $%d", argNum))
		args = append(args, filter.StartDate)
		argNum++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("d.created_at <= $%d", argNum))
		args = append(args, filter.EndDate)
		argNum++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("d.context_summary ILIKE $%d", argNum))
		args = append(args, "%"+filter.Search+"%")
		argNum++
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	var total int
	countQueryFull := countQuery + whereClause
	if err := r.db.QueryRow(ctx, countQueryFull, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count decisions: %w", err)
	}

	// Add GROUP BY, ORDER BY, and pagination
	fullQuery := baseQuery + whereClause + fmt.Sprintf(`
		GROUP BY d.id
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d
	`, argNum, argNum+1)
	args = append(args, filter.Limit, filter.Offset)

	// Execute query
	rows, err := r.db.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query decisions: %w", err)
	}
	defer rows.Close()

	decisions := make([]DecisionSummary, 0)
	for rows.Next() {
		var d DecisionSummary
		var models []string

		if err := rows.Scan(&d.ID, &d.CaseID, &d.Status, &d.Context, &d.Timestamp, &models); err != nil {
			return nil, fmt.Errorf("scan decision: %w", err)
		}

		if models == nil {
			d.ModelsConsulted = []string{}
		} else {
			d.ModelsConsulted = models
		}

		decisions = append(decisions, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate decisions: %w", err)
	}

	// Filter by model if specified (post-filter since we aggregated)
	if filter.Model != "" {
		filtered := make([]DecisionSummary, 0)
		for _, d := range decisions {
			for _, m := range d.ModelsConsulted {
				if strings.Contains(strings.ToLower(m), strings.ToLower(filter.Model)) {
					filtered = append(filtered, d)
					break
				}
			}
		}
		decisions = filtered
	}

	return &ListDecisionsResponse{
		Decisions: decisions,
		Total:     total,
		HasMore:   filter.Offset+len(decisions) < total,
	}, nil
}

// GetDecision returns a single decision with all its votes.
// See SHADOW_VIEWER_specs.md Section 3.1 GET /api/v1/tribunal/decisions/{id}
func (r *Repository) GetDecision(ctx context.Context, id uuid.UUID) (*DecisionDetail, error) {
	// Get decision
	decisionQuery := `
		SELECT id, case_id, status, context_summary, consensus_score, created_at
		FROM tribunal_decisions
		WHERE id = $1
	`

	var d DecisionDetail
	err := r.db.QueryRow(ctx, decisionQuery, id).Scan(
		&d.ID, &d.CaseID, &d.Status, &d.Context, &d.ConsensusScore, &d.Timestamp,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query decision: %w", err)
	}

	// Get votes
	votesQuery := `
		SELECT id, decision_id, model_name, vote, reasoning, latency_ms, token_count, cost_usd
		FROM tribunal_votes
		WHERE decision_id = $1
		ORDER BY model_name
	`

	rows, err := r.db.Query(ctx, votesQuery, id)
	if err != nil {
		return nil, fmt.Errorf("query votes: %w", err)
	}
	defer rows.Close()

	votes := make([]ModelVote, 0)
	for rows.Next() {
		var v ModelVote
		if err := rows.Scan(&v.ID, &v.DecisionID, &v.ModelName, &v.Vote, &v.Reasoning, &v.LatencyMs, &v.TokenCount, &v.CostUSD); err != nil {
			return nil, fmt.Errorf("scan vote: %w", err)
		}
		votes = append(votes, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate votes: %w", err)
	}

	d.Votes = votes

	// Extract policy links from vote reasoning (markdown links to docs/specs)
	d.PolicyLinks = extractPolicyLinks(votes)

	return &d, nil
}

// extractPolicyLinks finds markdown links in vote reasoning that point to docs or specs.
// Pattern: [text](path) where path starts with docs/ or specs/
func extractPolicyLinks(votes []ModelVote) []string {
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(((?:docs|specs)/[^)]+)\)`)
	linkSet := make(map[string]bool)

	for _, v := range votes {
		matches := linkRegex.FindAllStringSubmatch(v.Reasoning, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				linkSet[match[2]] = true
			}
		}
	}

	links := make([]string, 0, len(linkSet))
	for link := range linkSet {
		links = append(links, link)
	}
	return links
}

// CreateDecision stores a new Tribunal decision.
func (r *Repository) CreateDecision(ctx context.Context, id uuid.UUID, req TribunalRequest, status DecisionStatus, score float64, summary string) error {
	query := `
		INSERT INTO tribunal_decisions (id, case_id, status, context_summary, consensus_score, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	// Note: We are storing the "Summary" in context_summary for now as per schema
	// Ideally we should have separate columns, but schema reuse is efficient.
	_, err := r.db.Exec(ctx, query, id, req.CaseID, status, summary, score)
	if err != nil {
		return fmt.Errorf("create decision: %w", err)
	}
	return nil
}

// CreateVote stores a model's vote.
func (r *Repository) CreateVote(ctx context.Context, v ModelVote) error {
	query := `
		INSERT INTO tribunal_votes (id, decision_id, model_name, vote, reasoning, latency_ms, token_count, cost_usd)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	id := uuid.New()
	if v.ID != uuid.Nil {
		id = v.ID
	}

	_, err := r.db.Exec(ctx, query, id, v.DecisionID, v.ModelName, v.Vote, v.Reasoning, v.LatencyMs, v.TokenCount, v.CostUSD)
	if err != nil {
		return fmt.Errorf("create vote: %w", err)
	}
	return nil
}

// DecisionExistsByCaseID checks if a decision with the given case_id already exists.
// Used for idempotency in Automated PR Review. See docs/AUTOMATED_PR_REVIEW_PRD.md
func (r *Repository) DecisionExistsByCaseID(ctx context.Context, caseID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tribunal_decisions WHERE case_id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, caseID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check decision exists: %w", err)
	}
	return exists, nil
}

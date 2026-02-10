package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FeedServicer defines the contract for portfolio feed operations.
// See FRONTEND_V2_SPEC.md §5.1, §5.5
type FeedServicer interface {
	GetFeed(ctx context.Context, orgID uuid.UUID, projectFilter *uuid.UUID) (*FeedResponse, error)
	DismissCard(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID) error
	SnoozeCard(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID, until time.Time) error
	WriteCard(ctx context.Context, card *models.FeedCard) error
	GetCardByID(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID) (*models.FeedCard, error)
	MarkProcurementOrdered(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID) error
}

// FeedResponse is the response from GetFeed.
type FeedResponse struct {
	Greeting string                  `json:"greeting"`
	Summary  PortfolioSummary        `json:"summary"`
	Cards    []models.FeedCard       `json:"cards"`
}

// PortfolioSummary provides a high-level overview of the user's projects.
type PortfolioSummary struct {
	ActiveProjectCount   int                        `json:"active_project_count"`
	TotalTasks           int                        `json:"total_tasks"`
	CriticalAlerts       int                        `json:"critical_alerts"`
	ProjectedCompletions []ProjectCompletionSummary `json:"projected_completions"`
}

// ProjectCompletionSummary shows completion status for a single project.
type ProjectCompletionSummary struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	EndDate     string    `json:"end_date"`
	OnTrack     bool      `json:"on_track"`
	SlipDays    int       `json:"slip_days"`
}

// FeedService implements FeedServicer.
type FeedService struct {
	db *pgxpool.Pool
}

// NewFeedService creates a new FeedService.
func NewFeedService(db *pgxpool.Pool) *FeedService {
	return &FeedService{db: db}
}

// GetFeed retrieves the portfolio feed for an organization.
// Cards are sorted by horizon (today first), then priority (critical first), then deadline.
func (s *FeedService) GetFeed(ctx context.Context, orgID uuid.UUID, projectFilter *uuid.UUID) (*FeedResponse, error) {
	// 1. Get portfolio summary
	summary, err := s.getPortfolioSummary(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("feed: failed to get summary: %w", err)
	}

	// 2. Get active feed cards
	cards, err := s.getActiveCards(ctx, orgID, projectFilter)
	if err != nil {
		return nil, fmt.Errorf("feed: failed to get cards: %w", err)
	}

	// 3. Build greeting
	greeting := buildGreeting(time.Now())

	return &FeedResponse{
		Greeting: greeting,
		Summary:  *summary,
		Cards:    cards,
	}, nil
}

// DismissCard marks a feed card as dismissed.
func (s *FeedService) DismissCard(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID) error {
	result, err := s.db.Exec(ctx,
		`UPDATE feed_cards SET dismissed_at = NOW() WHERE id = $1 AND org_id = $2 AND dismissed_at IS NULL`,
		cardID, orgID,
	)
	if err != nil {
		return fmt.Errorf("feed: dismiss failed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("feed: card not found or already dismissed")
	}
	return nil
}

// SnoozeCard snoozes a feed card until a specified time.
func (s *FeedService) SnoozeCard(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID, until time.Time) error {
	result, err := s.db.Exec(ctx,
		`UPDATE feed_cards SET snoozed_until = $3 WHERE id = $1 AND org_id = $2 AND dismissed_at IS NULL`,
		cardID, orgID, until,
	)
	if err != nil {
		return fmt.Errorf("feed: snooze failed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("feed: card not found or already dismissed")
	}
	return nil
}

// WriteCard inserts or upserts a feed card.
// If a card with the same (project_id, card_type, task_id) exists, it updates in place.
func (s *FeedService) WriteCard(ctx context.Context, card *models.FeedCard) error {
	actionsJSON, err := json.Marshal(card.Actions)
	if err != nil {
		return fmt.Errorf("feed: marshal actions: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO feed_cards (
			org_id, project_id, card_type, priority, headline, body, consequence,
			horizon, deadline, actions, engine_data, agent_source, task_id, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (project_id, card_type, task_id)
			WHERE task_id IS NOT NULL AND dismissed_at IS NULL
		DO UPDATE SET
			priority = EXCLUDED.priority,
			headline = EXCLUDED.headline,
			body = EXCLUDED.body,
			consequence = EXCLUDED.consequence,
			horizon = EXCLUDED.horizon,
			deadline = EXCLUDED.deadline,
			actions = EXCLUDED.actions,
			engine_data = EXCLUDED.engine_data,
			expires_at = EXCLUDED.expires_at
	`,
		card.OrgID, card.ProjectID, card.CardType, card.Priority,
		card.Headline, card.Body, card.Consequence,
		card.Horizon, card.Deadline, actionsJSON, card.EngineData,
		card.AgentSource, card.TaskID, card.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("feed: write card: %w", err)
	}
	return nil
}

// GetCardByID retrieves a single feed card by ID, scoped to the org.
func (s *FeedService) GetCardByID(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID) (*models.FeedCard, error) {
	var card models.FeedCard
	err := s.db.QueryRow(ctx, `
		SELECT id, org_id, project_id, card_type, priority, headline, body,
			consequence, horizon, deadline, actions, engine_data, agent_source, task_id,
			created_at, expires_at
		FROM feed_cards
		WHERE id = $1 AND org_id = $2
	`, cardID, orgID).Scan(
		&card.ID, &card.OrgID, &card.ProjectID, &card.CardType, &card.Priority,
		&card.Headline, &card.Body, &card.Consequence, &card.Horizon, &card.Deadline,
		&card.ActionsJSON, &card.EngineData, &card.AgentSource, &card.TaskID,
		&card.CreatedAt, &card.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("feed: get card: %w", err)
	}
	if card.ActionsJSON != nil {
		if err := json.Unmarshal(card.ActionsJSON, &card.Actions); err != nil {
			card.Actions = []models.FeedCardAction{}
		}
	}
	return &card, nil
}

// MarkProcurementOrdered marks the procurement item linked to a card as 'Ordered'
// and dismisses the card. Returns error if card has no task_id or item not found.
func (s *FeedService) MarkProcurementOrdered(ctx context.Context, orgID uuid.UUID, cardID uuid.UUID) error {
	// Look up the card to get its task_id
	card, err := s.GetCardByID(ctx, orgID, cardID)
	if err != nil {
		return fmt.Errorf("mark ordered: card lookup: %w", err)
	}
	if card.TaskID == nil {
		return fmt.Errorf("mark ordered: card has no associated task")
	}

	// Update procurement_items linked to this project_task_id
	result, err := s.db.Exec(ctx, `
		UPDATE procurement_items
		SET status = 'Ordered', last_checked_at = NOW()
		WHERE project_task_id = $1
	`, *card.TaskID)
	if err != nil {
		return fmt.Errorf("mark ordered: update procurement: %w", err)
	}
	if result.RowsAffected() == 0 {
		slog.Warn("mark ordered: no procurement item found for task", "task_id", *card.TaskID)
	}

	// Dismiss the card
	return s.DismissCard(ctx, orgID, cardID)
}

func (s *FeedService) getActiveCards(ctx context.Context, orgID uuid.UUID, projectFilter *uuid.UUID) ([]models.FeedCard, error) {
	query := `
		SELECT fc.id, fc.org_id, fc.project_id, fc.card_type, fc.priority,
			fc.headline, fc.body, fc.consequence, fc.horizon, fc.deadline,
			fc.actions, fc.engine_data, fc.agent_source, fc.task_id,
			fc.created_at, fc.expires_at,
			p.name AS project_name
		FROM feed_cards fc
		JOIN projects p ON p.id = fc.project_id
		WHERE fc.org_id = $1
			AND fc.dismissed_at IS NULL
			AND (fc.snoozed_until IS NULL OR fc.snoozed_until < NOW())
			AND (fc.expires_at IS NULL OR fc.expires_at > NOW())
	`

	args := []interface{}{orgID}
	if projectFilter != nil {
		query += ` AND fc.project_id = $2`
		args = append(args, *projectFilter)
	}

	query += `
		ORDER BY
			CASE fc.horizon
				WHEN 'today' THEN 0
				WHEN 'this_week' THEN 1
				WHEN 'horizon' THEN 2
			END,
			fc.priority ASC,
			fc.deadline ASC NULLS LAST,
			fc.created_at DESC
		LIMIT 200
	`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query cards: %w", err)
	}
	defer rows.Close()

	var cards []models.FeedCard
	for rows.Next() {
		var card models.FeedCard
		err := rows.Scan(
			&card.ID, &card.OrgID, &card.ProjectID, &card.CardType, &card.Priority,
			&card.Headline, &card.Body, &card.Consequence, &card.Horizon, &card.Deadline,
			&card.ActionsJSON, &card.EngineData, &card.AgentSource, &card.TaskID,
			&card.CreatedAt, &card.ExpiresAt,
			&card.ProjectName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}

		// Unmarshal actions JSONB
		if card.ActionsJSON != nil {
			if err := json.Unmarshal(card.ActionsJSON, &card.Actions); err != nil {
				slog.Warn("feed: failed to unmarshal card actions", "card_id", card.ID, "error", err)
				card.Actions = []models.FeedCardAction{}
			}
		}

		cards = append(cards, card)
	}

	if cards == nil {
		cards = []models.FeedCard{}
	}

	return cards, rows.Err()
}

func (s *FeedService) getPortfolioSummary(ctx context.Context, orgID uuid.UUID) (*PortfolioSummary, error) {
	summary := &PortfolioSummary{
		ProjectedCompletions: []ProjectCompletionSummary{},
	}

	// Active project count and total tasks
	err := s.db.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT p.id),
			COALESCE(SUM(task_counts.cnt), 0)
		FROM projects p
		LEFT JOIN (
			SELECT project_id, COUNT(*) AS cnt FROM project_tasks GROUP BY project_id
		) task_counts ON task_counts.project_id = p.id
		WHERE p.org_id = $1 AND p.status IN ('Active', 'Preconstruction')
	`, orgID).Scan(&summary.ActiveProjectCount, &summary.TotalTasks)
	if err != nil {
		return nil, fmt.Errorf("summary counts: %w", err)
	}

	// Critical alerts count
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM feed_cards
		WHERE org_id = $1 AND priority <= 1 AND dismissed_at IS NULL
			AND (snoozed_until IS NULL OR snoozed_until < NOW())
	`, orgID).Scan(&summary.CriticalAlerts)
	if err != nil {
		return nil, fmt.Errorf("summary alerts: %w", err)
	}

	// Project completions — get the latest task end date per project as projected end
	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.name, p.target_end_date,
			MAX(pt.planned_end) AS projected_end
		FROM projects p
		LEFT JOIN project_tasks pt ON pt.project_id = p.id
		WHERE p.org_id = $1 AND p.status IN ('Active', 'Preconstruction')
		GROUP BY p.id, p.name, p.target_end_date
		ORDER BY p.name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("summary completions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			pcs           ProjectCompletionSummary
			targetEnd     *time.Time
			projectedEnd  *time.Time
		)
		if err := rows.Scan(&pcs.ProjectID, &pcs.ProjectName, &targetEnd, &projectedEnd); err != nil {
			return nil, fmt.Errorf("scan completion: %w", err)
		}
		if projectedEnd != nil {
			pcs.EndDate = projectedEnd.Format("2006-01-02")
			if targetEnd != nil {
				slip := projectedEnd.Sub(*targetEnd).Hours() / 24
				if slip > 0 {
					pcs.SlipDays = int(slip)
				}
				pcs.OnTrack = slip <= 0
			} else {
				pcs.OnTrack = true
			}
		} else if targetEnd != nil {
			pcs.EndDate = targetEnd.Format("2006-01-02")
			pcs.OnTrack = true
		}
		summary.ProjectedCompletions = append(summary.ProjectedCompletions, pcs)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summary, nil
}

func buildGreeting(now time.Time) string {
	hour := now.Hour()
	switch {
	case hour < 12:
		return "Good morning"
	case hour < 17:
		return "Good afternoon"
	default:
		return "Good evening"
	}
}

// Ensure FeedService satisfies FeedServicer at compile time.
var _ FeedServicer = (*FeedService)(nil)

// ListActiveProjectsForOrg returns active/preconstruction projects for an org.
// Used by the feed handler to populate project pills.
func (s *FeedService) ListActiveProjectsForOrg(ctx context.Context, orgID uuid.UUID) ([]models.Project, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, org_id, name, address, status, permit_issued_date, target_end_date
		FROM projects
		WHERE org_id = $1 AND status IN ('Active', 'Preconstruction')
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("feed: list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Address, &p.Status,
			&p.PermitIssuedDate, &p.TargetEndDate); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []models.Project{}
	}
	return projects, rows.Err()
}

// Suppress unused import warnings for pgx (used by interface alignment)
var _ = pgx.ErrNoRows

package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ResourceConflict represents a single contact booked on overlapping dates across projects.
type ResourceConflict struct {
	ContactID   uuid.UUID        `json:"contact_id"`
	ContactName string           `json:"contact_name"`
	Trade       string           `json:"trade"`
	Conflicts   []ProjectOverlap `json:"conflicts"`
}

// ProjectOverlap represents overlapping schedules for the same contact across two projects.
type ProjectOverlap struct {
	ProjectAID   uuid.UUID `json:"project_a_id"`
	ProjectAName string    `json:"project_a_name"`
	ProjectBID   uuid.UUID `json:"project_b_id"`
	ProjectBName string    `json:"project_b_name"`
	OverlapStart time.Time `json:"overlap_start"`
	OverlapEnd   time.Time `json:"overlap_end"`
	OverlapDays  int       `json:"overlap_days"`
}

// ResourceConflictService detects when the same subcontractor is booked on
// overlapping dates across multiple projects in the same org.
type ResourceConflictService struct {
	db         *pgxpool.Pool
	feedWriter FeedWriter
}

// NewResourceConflictService creates a new resource conflict detection service.
func NewResourceConflictService(db *pgxpool.Pool, feedWriter FeedWriter) *ResourceConflictService {
	return &ResourceConflictService{db: db, feedWriter: feedWriter}
}

// contactBooking represents a contact's scheduled time on a specific project.
type contactBooking struct {
	ContactID   uuid.UUID
	ContactName string
	Trade       string
	ProjectID   uuid.UUID
	ProjectName string
	Start       time.Time
	End         time.Time
}

// DetectConflicts finds all contacts booked on overlapping dates across projects in the org.
func (s *ResourceConflictService) DetectConflicts(ctx context.Context, orgID uuid.UUID) ([]ResourceConflict, error) {
	// Query all active project task assignments with scheduled dates
	query := `
		SELECT
			ca.contact_id,
			c.name AS contact_name,
			c.trade,
			pt.project_id,
			p.name AS project_name,
			pt.early_start,
			pt.early_finish
		FROM contact_assignments ca
		JOIN contacts c ON ca.contact_id = c.id
		JOIN project_tasks pt ON ca.task_id = pt.id
		JOIN projects p ON pt.project_id = p.id
		WHERE p.org_id = $1
		  AND p.status = 'Active'
		  AND pt.early_start IS NOT NULL
		  AND pt.early_finish IS NOT NULL
		  AND pt.status != 'Completed'
		ORDER BY ca.contact_id, pt.early_start
	`
	rows, err := s.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query contact bookings: %w", err)
	}
	defer rows.Close()

	// Group bookings by contact
	bookingsByContact := make(map[uuid.UUID][]contactBooking)
	for rows.Next() {
		var b contactBooking
		if err := rows.Scan(&b.ContactID, &b.ContactName, &b.Trade,
			&b.ProjectID, &b.ProjectName, &b.Start, &b.End); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookingsByContact[b.ContactID] = append(bookingsByContact[b.ContactID], b)
	}

	// Detect overlaps across different projects for each contact
	var conflicts []ResourceConflict
	for _, bookings := range bookingsByContact {
		if len(bookings) < 2 {
			continue
		}

		var overlaps []ProjectOverlap
		for i := 0; i < len(bookings); i++ {
			for j := i + 1; j < len(bookings); j++ {
				a, b := bookings[i], bookings[j]
				if a.ProjectID == b.ProjectID {
					continue // Same project, not a conflict
				}
				// Check overlap: a.Start < b.End && b.Start < a.End
				if a.Start.Before(b.End) && b.Start.Before(a.End) {
					overlapStart := a.Start
					if b.Start.After(overlapStart) {
						overlapStart = b.Start
					}
					overlapEnd := a.End
					if b.End.Before(overlapEnd) {
						overlapEnd = b.End
					}
					days := int(overlapEnd.Sub(overlapStart).Hours() / 24)
					if days > 0 {
						overlaps = append(overlaps, ProjectOverlap{
							ProjectAID:   a.ProjectID,
							ProjectAName: a.ProjectName,
							ProjectBID:   b.ProjectID,
							ProjectBName: b.ProjectName,
							OverlapStart: overlapStart,
							OverlapEnd:   overlapEnd,
							OverlapDays:  days,
						})
					}
				}
			}
		}

		if len(overlaps) > 0 {
			conflicts = append(conflicts, ResourceConflict{
				ContactID:   bookings[0].ContactID,
				ContactName: bookings[0].ContactName,
				Trade:       bookings[0].Trade,
				Conflicts:   overlaps,
			})
		}
	}

	return conflicts, nil
}

// DetectAndNotify runs conflict detection and writes feed cards for each conflict.
func (s *ResourceConflictService) DetectAndNotify(ctx context.Context, orgID uuid.UUID) error {
	conflicts, err := s.DetectConflicts(ctx, orgID)
	if err != nil {
		return err
	}

	for _, c := range conflicts {
		s.writeConflictCard(ctx, orgID, c)
	}

	slog.Info("resource conflict scan complete", "org_id", orgID, "conflicts_found", len(conflicts))
	return nil
}

// writeConflictCard creates a feed card for a resource conflict.
func (s *ResourceConflictService) writeConflictCard(ctx context.Context, orgID uuid.UUID, conflict ResourceConflict) {
	if s.feedWriter == nil || len(conflict.Conflicts) == 0 {
		return
	}

	headline := fmt.Sprintf("%s (%s) double-booked across %d projects",
		conflict.ContactName, conflict.Trade, len(conflict.Conflicts)+1)

	body := fmt.Sprintf("%s is scheduled on overlapping dates across multiple projects:", conflict.ContactName)
	limit := 3
	if len(conflict.Conflicts) < limit {
		limit = len(conflict.Conflicts)
	}
	for i := 0; i < limit; i++ {
		o := conflict.Conflicts[i]
		body += fmt.Sprintf("\n- %s vs %s: %d days overlap (%s – %s)",
			o.ProjectAName, o.ProjectBName, o.OverlapDays,
			o.OverlapStart.Format("Jan 02"), o.OverlapEnd.Format("Jan 02"))
	}

	agentSource := "ResourceConflictService"
	consequence := fmt.Sprintf("One or more projects may experience delays if %s cannot cover all commitments", conflict.ContactName)

	// Use the first conflict's project for the feed card
	projectID := conflict.Conflicts[0].ProjectAID

	card := &models.FeedCard{
		OrgID:       orgID,
		ProjectID:   projectID,
		CardType:    models.FeedCardScheduleRecalc,
		Priority:    models.FeedCardPriorityNormal,
		Headline:    headline,
		Body:        body,
		Consequence: &consequence,
		Horizon:     models.FeedCardHorizonThisWeek,
		AgentSource: &agentSource,
		Actions: []models.FeedCardAction{
			{ID: "view_contacts", Label: "View Contacts", Style: "primary"},
			{ID: "dismiss", Label: "Dismiss", Style: "secondary"},
		},
	}

	if err := s.feedWriter.WriteCard(ctx, card); err != nil {
		slog.Error("failed to write resource conflict feed card", "contact", conflict.ContactName, "error", err)
	}
}

// DetectAndNotifyAll scans all active orgs for resource conflicts.
// Used by the weekly cron job handler.
func (s *ResourceConflictService) DetectAndNotifyAll(ctx context.Context) error {
	rows, err := s.db.Query(ctx, `SELECT DISTINCT org_id FROM projects WHERE status = 'Active'`)
	if err != nil {
		return fmt.Errorf("list active orgs: %w", err)
	}
	defer rows.Close()

	var orgIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
		orgIDs = append(orgIDs, id)
	}

	for _, orgID := range orgIDs {
		if err := s.DetectAndNotify(ctx, orgID); err != nil {
			slog.Warn("resource conflict scan failed for org", "org_id", orgID, "error", err)
		}
	}
	return nil
}

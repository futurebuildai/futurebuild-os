package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// See FRONTEND_V2_SPEC.md §11.2 — Passive Drift Detection
//
// The DriftDetectionAgent silently tracks the ratio of actual vs. predicted
// task durations for each project. It only emits a calibration_drift feed
// card when ALL of these conditions are met:
//   1. >= 8 tasks have been completed (statistical minimum)
//   2. Rolling average ratio deviates >25% from 1.0 consistently
//   3. Deviation has persisted across the last 5+ completed tasks (sustained)
//
// This is a background agent — most users will never see a drift card.

// MinCompletedTasks is the statistical minimum before drift analysis is valid.
const MinCompletedTasks = 8

// SustainedWindow is the number of recent tasks that must consistently show drift.
const SustainedWindow = 5

// DriftThreshold is the minimum ratio deviation from 1.0 to be considered drift.
// 0.25 = 25% faster or slower than predicted.
const DriftThreshold = 0.25

// CompletedTaskRow is a projection of a completed task with actual and predicted durations.
type CompletedTaskRow struct {
	TaskID              uuid.UUID
	ProjectID           uuid.UUID
	OrgID               uuid.UUID
	PredictedDuration   float64 // DHSM-adjusted days (calculated_duration)
	ActualDurationDays  float64 // actual_end - actual_start in days
}

// DriftRepository abstracts the database queries needed by the drift agent.
type DriftRepository interface {
	// StreamCompletedTasksByProject calls fn for each project's batch of completed tasks.
	// Tasks must have both actual_start, actual_end, and calculated_duration > 0.
	StreamCompletedTasksByProject(ctx context.Context, fn func(projectID, orgID uuid.UUID, tasks []CompletedTaskRow) error) error
}

// DriftDetectionAgent analyzes actual vs. predicted task durations and emits
// calibration_drift feed cards when sustained deviation is detected.
type DriftDetectionAgent struct {
	repo        DriftRepository
	clock       clock.Clock
	feedWriter  FeedWriter
	asynqClient *asynq.Client
}

// NewDriftDetectionAgent creates a new drift detection agent.
func NewDriftDetectionAgent(repo DriftRepository, clk clock.Clock) *DriftDetectionAgent {
	return &DriftDetectionAgent{repo: repo, clock: clk}
}

// WithFeedWriter injects the optional feed card writer.
func (a *DriftDetectionAgent) WithFeedWriter(fw FeedWriter) *DriftDetectionAgent {
	a.feedWriter = fw
	return a
}

// WithAsynqClient injects the asynq client for enqueuing delay cascade tasks.
func (a *DriftDetectionAgent) WithAsynqClient(c *asynq.Client) *DriftDetectionAgent {
	a.asynqClient = c
	return a
}

// Execute runs the drift analysis across all projects.
func (a *DriftDetectionAgent) Execute(ctx context.Context) error {
	if a.feedWriter == nil {
		return fmt.Errorf("drift detection: feedWriter not configured")
	}

	var cardsWritten int
	err := a.repo.StreamCompletedTasksByProject(ctx, func(projectID, orgID uuid.UUID, tasks []CompletedTaskRow) error {
		if len(tasks) < MinCompletedTasks {
			return nil // Not enough data for statistical significance
		}

		// Calculate rolling ratio for each task: actual / predicted
		ratios := make([]float64, len(tasks))
		for i, t := range tasks {
			if t.PredictedDuration <= 0 {
				ratios[i] = 1.0 // Avoid division by zero
			} else {
				ratios[i] = t.ActualDurationDays / t.PredictedDuration
			}
		}

		// Check the last SustainedWindow tasks for consistent drift
		windowStart := len(ratios) - SustainedWindow
		if windowStart < 0 {
			windowStart = 0
		}
		recentRatios := ratios[windowStart:]

		// All recent tasks must deviate in the same direction
		var allFaster, allSlower bool
		allFaster = true
		allSlower = true
		var sumRatio float64
		for _, r := range recentRatios {
			sumRatio += r
			if r >= (1.0 - DriftThreshold) {
				allFaster = false
			}
			if r <= (1.0 + DriftThreshold) {
				allSlower = false
			}
		}

		if !allFaster && !allSlower {
			return nil // No sustained drift
		}

		avgRatio := sumRatio / float64(len(recentRatios))
		deviationPct := int(math.Abs(avgRatio-1.0) * 100)

		// Build the card
		var direction, consequence string
		if allFaster {
			direction = "faster"
			consequence = fmt.Sprintf("Your crew is completing tasks %d%% faster than predicted. The schedule may be more conservative than needed.", deviationPct)
		} else {
			direction = "slower"
			consequence = fmt.Sprintf("Tasks are taking %d%% longer than predicted. The schedule may need adjustment to stay on track.", deviationPct)
		}

		headline := fmt.Sprintf("Crew trending %d%% %s than predicted", deviationPct, direction)
		agentSource := "drift_detection"

		card := &models.FeedCard{
			ID:          uuid.New(),
			OrgID:       orgID,
			ProjectID:   projectID,
			CardType:    models.FeedCardCalibrationDrift,
			Priority:    models.FeedCardPriorityLow, // P2 blue dot — low urgency
			Headline:    headline,
			Body:        fmt.Sprintf("Based on %d completed tasks, your crew is consistently %s than the engine predicted.", len(tasks), direction),
			Consequence: &consequence,
			Horizon:     models.FeedCardHorizonThisWeek,
			AgentSource: &agentSource,
			Actions: []models.FeedCardAction{
				{ID: "recalibrate", Label: "Adjust predictions", Style: "primary"},
				{ID: "dismiss", Label: "Keep current", Style: "secondary"},
			},
		}

		if err := a.feedWriter.WriteCard(ctx, card); err != nil {
			slog.Warn("drift detection: failed to write card", "project_id", projectID, "error", err)
			return nil // Don't stop processing other projects
		}
		cardsWritten++

		// When crew is slower than predicted, enqueue delay cascade analysis
		// to show cascading impact on downstream tasks and project end date.
		if allSlower && a.asynqClient != nil {
			lastTask := tasks[len(tasks)-1]
			avgSlipDays := int(math.Round((avgRatio - 1.0) * lastTask.PredictedDuration))
			if avgSlipDays > 0 {
				cascadeTask, cErr := newDelayCascadeTask(projectID, orgID, lastTask.TaskID, avgSlipDays)
				if cErr == nil {
					if _, enqErr := a.asynqClient.EnqueueContext(ctx, cascadeTask); enqErr != nil {
						slog.Warn("drift detection: failed to enqueue delay cascade", "project_id", projectID, "error", enqErr)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("drift detection: %w", err)
	}

	slog.Info("drift detection complete", "cards_written", cardsWritten)
	return nil
}

// newDelayCascadeTask creates a delay cascade asynq task without importing worker package.
func newDelayCascadeTask(projectID, orgID, taskID uuid.UUID, slipDays int) (*asynq.Task, error) {
	payload, err := json.Marshal(struct {
		ProjectID uuid.UUID `json:"project_id"`
		OrgID     uuid.UUID `json:"org_id"`
		TaskID    uuid.UUID `json:"task_id"`
		SlipDays  int       `json:"slip_days"`
	}{
		ProjectID: projectID,
		OrgID:     orgID,
		TaskID:    taskID,
		SlipDays:  slipDays,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("task:delay_cascade", payload, asynq.Queue("default")), nil
}


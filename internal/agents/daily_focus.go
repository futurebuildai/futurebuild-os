package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// MaxConcurrentProjects limits concurrent AI/DB calls to prevent thundering herds.
// P1 Scalability Fix: Worker pool pattern.
const MaxConcurrentProjects = 10

// DailyFocusAgent orchestrates the morning briefing generation.
// See PRODUCTION_PLAN.md Step 49 (Service Layer Pattern)
// Refactored for deterministic simulation: PRODUCTION_PLAN.md Step 49
// Critical Blocker A Remediation: Added geocoder and directory dependencies
// P1 Scalability Fix: Uses streaming + worker pool for O(1) memory at scale
type DailyFocusAgent struct {
	projects    ProjectRepository // Abstraction for streaming (replaces *service.ProjectService)
	schedule    *service.ScheduleService
	weather     types.WeatherService
	notifier    types.NotificationService
	aiClient    ai.Client
	clock       clock.Clock
	geocoder    types.GeocodingService // Blocker A: Address → lat/long
	directory   types.DirectoryService // Blocker A: PM email lookup
	feedWriter  FeedWriter             // V2: writes daily_briefing cards to portfolio feed
	claudeRunner *AgentRunner          // Phase 6: Claude reasoning for actionable recommendations
	asynqClient *asynq.Client          // Feature 4: Enqueue briefing notifications
}

// NewDailyFocusAgent creates a new agent instance.
// Clock is required for deterministic time simulation (Step 49).
// Critical Blocker A Remediation: geocoder and directory are required
// P1 Scalability Fix: uses ProjectRepository interface for streaming
func NewDailyFocusAgent(
	projects ProjectRepository, // Abstraction for streaming
	schedule *service.ScheduleService,
	weather types.WeatherService,
	notifier types.NotificationService,
	aiClient ai.Client,
	clk clock.Clock,
	geocoder types.GeocodingService,
	directory types.DirectoryService,
) *DailyFocusAgent {
	return &DailyFocusAgent{
		projects:  projects,
		schedule:  schedule,
		weather:   weather,
		notifier:  notifier,
		aiClient:  aiClient,
		clock:     clk,
		geocoder:  geocoder,
		directory: directory,
	}
}

// NewDailyFocusAgentWithService is a convenience constructor using ProjectService directly.
// Maintains backward compatibility with existing callers.
func NewDailyFocusAgentWithService(
	projectSvc *service.ProjectService,
	schedule *service.ScheduleService,
	weather types.WeatherService,
	notifier types.NotificationService,
	aiClient ai.Client,
	clk clock.Clock,
	geocoder types.GeocodingService,
	directory types.DirectoryService,
) *DailyFocusAgent {
	return NewDailyFocusAgent(
		NewPgProjectRepository(projectSvc),
		schedule, weather, notifier, aiClient, clk, geocoder, directory,
	)
}

// WithFeedWriter sets the feed writer for V2 portfolio feed card generation.
func (a *DailyFocusAgent) WithFeedWriter(fw FeedWriter) *DailyFocusAgent {
	a.feedWriter = fw
	return a
}

// WithAsynqClient sets the Asynq client for enqueuing follow-up tasks (briefing notifications).
func (a *DailyFocusAgent) WithAsynqClient(client *asynq.Client) *DailyFocusAgent {
	a.asynqClient = client
	return a
}

// Execute runs the daily briefing logic for all active projects.
// P1 Scalability Fix: Uses streaming + worker pool for O(1) memory at scale.
// - Streams projects one-by-one (no unbounded slice allocation)
// - Worker pool limits concurrent AI/DB calls (prevents thundering herd)
// - Observability: logs projects_processed_count and batch_execution_duration
func (a *DailyFocusAgent) Execute(ctx context.Context) error {
	slog.Info("Starting Daily Focus Agent...")
	start := a.clock.Now()

	// Worker pool for concurrency control (semaphore pattern)
	sem := make(chan struct{}, MaxConcurrentProjects)
	var wg sync.WaitGroup
	var processed int64
	var errors int64

	// Stream projects one-by-one to avoid unbounded memory allocation
	err := a.projects.StreamActiveProjects(ctx, func(p models.Project) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Acquire semaphore slot
		sem <- struct{}{}
		wg.Add(1)

		// Process in goroutine with worker pool limiting
		go func(project models.Project) {
			defer func() {
				<-sem // Release semaphore slot
				wg.Done()
			}()

			if err := a.processProject(ctx, project); err != nil {
				slog.Error("ERROR processing project", "project_id", project.ID, "error", err)
				atomic.AddInt64(&errors, 1)
				// Continue with other projects even if one fails
			}
			atomic.AddInt64(&processed, 1)
		}(p)

		return nil
	})

	// Wait for all goroutines to complete
	wg.Wait()

	// Observability: Log metrics
	duration := a.clock.Now().Sub(start)
	slog.Info("Daily Focus Agent completed",
		"projects_processed_count", processed,
		"projects_error_count", errors,
		"batch_execution_duration", duration,
	)

	if err != nil {
		return fmt.Errorf("streaming projects failed: %w", err)
	}

	return nil
}

func (a *DailyFocusAgent) processProject(ctx context.Context, p models.Project) error {
	// 2. Fetch Context Data
	// A. Weather - Critical Blocker A Remediation: Use geocoded project address
	// See BACKEND_SCOPE.md Section 2.4 (Weather-Sensitive Phases)
	lat, lng := 30.2672, -97.7431 // Default fallback (Austin, TX)
	if a.geocoder != nil {
		if geoLat, geoLng, err := a.geocoder.Geocode(p.Address); err == nil {
			lat, lng = geoLat, geoLng
		} else {
			slog.Warn("geocoding failed for project, using default coords", "project_id", p.ID, "error", err)
		}
	}
	forecast, err := a.weather.GetForecast(lat, lng)
	if err != nil {
		slog.Warn("failed to get weather for project", "project_id", p.ID, "error", err)
		// Graceful degradation: proceed without weather data
	}

	// B. Schedule Status (Critical Path & Today's Tasks)
	// We'll query tasks relevant to "Today"
	// For "Today", we check tasks where today falls between PlannedStart and PlannedEnd
	// OR tasks that are starting soon.
	// Simplify: Fetch schedule summary and critical path.

	// Re-using ScheduleService.GetProjectSchedule gives high level stats,
	// but we need specific task details for the prompt.

	// Clean Service Call - enables mocking for Time-Travel simulation
	tasks, err := a.schedule.GetAgentFocusTasks(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch relevant tasks: %w", err)
	}

	// 3. Generate briefing — try Claude first, fall back to Gemini text generation
	var briefing string
	if a.claudeRunner != nil && a.feedWriter != nil {
		// Phase 6: Claude-powered daily focus with actionable approval cards.
		// Claude calls tools to create approval cards + generates briefing text.
		claudeBriefing, claudeErr := a.processProjectWithClaude(ctx, p, tasks)
		if claudeErr != nil {
			slog.Warn("claude daily focus failed, falling back to gemini",
				"project_id", p.ID, "error", claudeErr)
			// Fall through to legacy path below
		} else {
			briefing = claudeBriefing
		}
	}

	// Legacy path: Gemini text generation (also used as fallback)
	if briefing == "" {
		prompt := a.buildPrompt(p, forecast, tasks)
		if a.aiClient != nil {
			req := ai.NewTextRequest(ai.ModelTypeFlash, prompt)
			resp, err := a.aiClient.GenerateContent(ctx, req)
			if err != nil {
				return fmt.Errorf("AI generation failed: %w", err)
			}
			briefing = resp.Text
		} else {
			slog.Warn("AI client not available for project, skipping briefing generation", "project_id", p.ID)
			briefing = "[AI Unavailable] Manual briefing required. Please review project schedule and weather conditions."
		}
	}

	// 4. Deliver - Critical Blocker A Remediation: Dynamic PM email lookup
	slog.Info("DAILY BRIEFING generated", "project_name", p.Name)

	// Look up Project Manager contact via DirectoryService (fallback to generic)
	recipientEmail := "superintendent@futurebuild.sh"
	if a.directory != nil {
		if contact, err := a.directory.GetProjectManager(ctx, p.ID, p.OrgID); err == nil && contact.Email != "" {
			recipientEmail = contact.Email
		} else {
			slog.Warn("PM lookup failed for project, using fallback email", "project_id", p.ID, "error", err)
		}
	}
	if err := a.notifier.SendEmail(recipientEmail, fmt.Sprintf("Daily Briefing: %s", p.Name), briefing); err != nil {
		slog.Warn("failed to send notification", "error", err)
	}

	// V2 Feed: Write daily_briefing card to portfolio feed
	if a.feedWriter != nil {
		if a.claudeRunner != nil {
			a.writeDailyBriefingCardWithClaude(ctx, p, briefing, tasks)
		} else {
			a.writeDailyBriefingCard(ctx, p, briefing, tasks)
		}
	}

	// Feature 12: Detect upcoming inspections and write feed cards
	if a.feedWriter != nil {
		a.detectUpcomingInspections(ctx, p, tasks)
	}

	// Feature 4: Enqueue briefing push notification
	if a.asynqClient != nil {
		summary := extractBriefingSummary(briefing)
		task, err := NewDailyBriefingNotificationTask(p.ID, p.OrgID, summary)
		if err == nil {
			if _, err := a.asynqClient.Enqueue(task); err != nil {
				slog.Warn("failed to enqueue briefing notification", "project_id", p.ID, "error", err)
			}
		}
	}

	return nil
}

// writeDailyBriefingCard creates a daily_briefing feed card from the AI-generated briefing.
func (a *DailyFocusAgent) writeDailyBriefingCard(ctx context.Context, p models.Project, briefing string, tasks []models.ProjectTask) {
	headline := "Daily Briefing: " + p.Name
	for _, line := range strings.Split(briefing, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			headline = trimmed
			if len(headline) > 120 {
				headline = headline[:117] + "..."
			}
			break
		}
	}

	var critCount int
	for _, t := range tasks {
		if t.IsOnCriticalPath {
			critCount++
		}
	}
	var consequence *string
	if critCount > 0 {
		c := fmt.Sprintf("%d critical-path task(s) active today", critCount)
		consequence = &c
	}

	agentSource := "DailyFocusAgent"
	endOfDay := time.Date(
		a.clock.Now().Year(), a.clock.Now().Month(), a.clock.Now().Day(),
		23, 59, 59, 0, a.clock.Now().Location(),
	)

	card := &models.FeedCard{
		OrgID:       p.OrgID,
		ProjectID:   p.ID,
		CardType:    models.FeedCardDailyBriefing,
		Priority:    models.FeedCardPriorityNormal,
		Headline:    headline,
		Body:        briefing,
		Consequence: consequence,
		Horizon:     models.FeedCardHorizonToday,
		AgentSource: &agentSource,
		ExpiresAt:   &endOfDay,
		Actions: []models.FeedCardAction{
			{ID: "view_briefing", Label: "View Full Briefing", Style: "primary"},
			{ID: "dismiss", Label: "Dismiss", Style: "secondary"},
		},
	}

	if err := a.feedWriter.WriteCard(ctx, card); err != nil {
		slog.Error("failed to write daily briefing feed card", "project_id", p.ID, "error", err)
	}
}

func (a *DailyFocusAgent) buildPrompt(p models.Project, w types.Forecast, tasks []models.ProjectTask) string {
	taskContext := ""
	for _, t := range tasks {
		criticalMarker := ""
		if t.IsOnCriticalPath {
			criticalMarker = "[CRITICAL PATH]"
		}
		start := "N/A"
		if t.PlannedStart != nil {
			start = t.PlannedStart.Format("2006-01-02")
		}
		taskContext += fmt.Sprintf("- %s %s (Status: %s, Start: %s)\n", criticalMarker, t.Name, t.Status, start)
	}

	return fmt.Sprintf(`
You are the Superintendent for the project "%s".
Today is %s.
Location: %s.

Weather Forecast:
- High: %.1fC, Low: %.1fC
- Rain: %.1fmm (%.0f%% chance)
- Condition: %s

Current Task Focus (Top priorities):
%s

INSTRUCTIONS:
Generate a concise "Morning Briefing" for the site team.
1. Highlight the weather constraints (e.g. "Rain expected, cover the lumber").
2. List the Top 3 Critical actions for today.
3. Identify any blocked tasks or risks based on the task list.
4. Keep it professional, direct, and motivating.
5. Use Markdown formatting.
`, p.Name, a.clock.Now().Format("Monday, Jan 02"), p.Address,
		w.HighTempC, w.LowTempC, w.PrecipitationMM, w.PrecipitationProbability*100, w.Conditions,
		taskContext)
}

// NewDailyBriefingNotificationTask creates a task payload for briefing push notification.
// Local helper to avoid circular import with worker package.
func NewDailyBriefingNotificationTask(projectID, orgID uuid.UUID, summary string) (*asynq.Task, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"project_id": projectID,
		"org_id":     orgID,
		"card_id":    uuid.New(), // placeholder
		"summary":    summary,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("task:daily_briefing_notification", payload, asynq.Queue("default")), nil
}

// extractBriefingSummary extracts the first 2-3 sentences from a briefing for notification body.
func extractBriefingSummary(briefing string) string {
	lines := strings.Split(briefing, "\n")
	var sentences []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") {
			continue
		}
		sentences = append(sentences, trimmed)
		if len(sentences) >= 2 {
			break
		}
	}
	if len(sentences) == 0 {
		return "Your daily briefing is ready. Check the feed for details."
	}
	return strings.Join(sentences, " ")
}

// detectUpcomingInspections identifies inspection tasks starting within 10 business days
// and writes feed cards alerting the PM. See Feature 12.
func (a *DailyFocusAgent) detectUpcomingInspections(ctx context.Context, p models.Project, tasks []models.ProjectTask) {
	today := a.clock.Now().Truncate(24 * time.Hour)

	for _, t := range tasks {
		if !t.IsInspection {
			continue
		}
		if t.PlannedStart == nil {
			continue
		}
		// Skip completed inspections
		if string(t.Status) == "Completed" || string(t.Status) == "complete" {
			continue
		}

		daysUntil := int(t.PlannedStart.Sub(today).Hours() / 24)

		if daysUntil < 0 || daysUntil > 10 {
			continue
		}

		var priority int
		var headline, body string
		agentSource := "DailyFocusAgent"

		if daysUntil <= 5 {
			// Urgent: schedule inspection now
			priority = models.FeedCardPriorityUrgent
			headline = fmt.Sprintf("Schedule inspection: %s in %d days", t.Name, daysUntil)
			body = fmt.Sprintf("Inspection \"%s\" (WBS %s) is scheduled to start on %s — %d business days from now. "+
				"Contact your inspector to confirm the appointment.",
				t.Name, t.WBSCode, t.PlannedStart.Format("Jan 02"), daysUntil)

			// Check if prerequisite tasks are complete (simplified check)
			body += a.checkInspectionPrereqs(tasks, t)
		} else {
			// Advisory: prepare for inspection
			priority = models.FeedCardPriorityNormal
			headline = fmt.Sprintf("Prepare for inspection: %s in %d days", t.Name, daysUntil)
			body = fmt.Sprintf("Inspection \"%s\" (WBS %s) is coming up on %s. "+
				"Ensure all prerequisite work is completed and the site is ready.",
				t.Name, t.WBSCode, t.PlannedStart.Format("Jan 02"))
		}

		card := &models.FeedCard{
			OrgID:       p.OrgID,
			ProjectID:   p.ID,
			CardType:    models.FeedCardInspectionUpcoming,
			Priority:    priority,
			Headline:    headline,
			Body:        body,
			Horizon:     models.FeedCardHorizonThisWeek,
			AgentSource: &agentSource,
			TaskID:      &t.ID,
			Actions: []models.FeedCardAction{
				{ID: "view_schedule", Label: "View Schedule", Style: "primary"},
				{ID: "dismiss", Label: "Dismiss", Style: "secondary"},
			},
		}

		if err := a.feedWriter.WriteCard(ctx, card); err != nil {
			slog.Error("failed to write inspection feed card", "project_id", p.ID, "task_id", t.ID, "error", err)
		}
	}
}

// checkInspectionPrereqs checks if predecessor tasks are complete and adds warnings.
func (a *DailyFocusAgent) checkInspectionPrereqs(allTasks []models.ProjectTask, inspection models.ProjectTask) string {
	// Build a map of task names to check if immediate predecessor tasks are complete
	// This is a simplified check — full dependency graph would require task deps query
	inspWBS := inspection.WBSCode
	if inspWBS == "" {
		return ""
	}

	var blockers []string
	for _, t := range allTasks {
		if t.ID == inspection.ID {
			continue
		}
		// Check tasks in the same phase that aren't inspections and aren't complete
		if !t.IsInspection && t.WBSCode != "" && string(t.Status) != "Completed" && string(t.Status) != "complete" {
			// Simple heuristic: non-inspection tasks with lower WBS code in same phase
			if len(t.WBSCode) > 0 && len(inspWBS) > 0 && t.WBSCode[0] == inspWBS[0] && t.WBSCode < inspWBS {
				blockers = append(blockers, t.Name)
			}
		}
	}

	if len(blockers) == 0 {
		return ""
	}
	if len(blockers) > 3 {
		blockers = blockers[:3]
	}
	return fmt.Sprintf("\n\n**Prerequisite tasks still in progress:** %s", strings.Join(blockers, ", "))
}



package agents

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
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
	projects  ProjectRepository // Abstraction for streaming (replaces *service.ProjectService)
	schedule  *service.ScheduleService
	weather   types.WeatherService
	notifier  types.NotificationService
	aiClient  ai.Client
	clock     clock.Clock
	geocoder  types.GeocodingService // Blocker A: Address → lat/long
	directory types.DirectoryService // Blocker A: PM email lookup
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

	// 3. Synthesize AI Prompt
	prompt := a.buildPrompt(p, forecast, tasks)

	// 4. Generate Content
	// Graceful degradation: if AI client is unavailable, log warning and skip generation
	var briefing string
	if a.aiClient != nil {
		// L7 Vendor Abstraction: Use ai.NewTextRequest instead of genai.Part
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

	// 5. Deliver - Critical Blocker A Remediation: Dynamic PM email lookup
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

	return nil
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

package agents

import (
	"context"
	"fmt"
	"log"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"google.golang.org/genai"
)

// DailyFocusAgent orchestrates the morning briefing generation.
// See PRODUCTION_PLAN.md Step 49 (Service Layer Pattern)
// Refactored for deterministic simulation: PRODUCTION_PLAN.md Step 49
// Critical Blocker A Remediation: Added geocoder and directory dependencies
type DailyFocusAgent struct {
	projects  *service.ProjectService // Replaces *pgxpool.Pool
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
func NewDailyFocusAgent(
	projects *service.ProjectService, // Replaces db
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

// Execute runs the daily briefing logic for all active projects.
func (a *DailyFocusAgent) Execute(ctx context.Context) error {
	log.Println("Starting Daily Focus Agent...")

	// 1. Fetch Active Projects via Service Layer
	// Clean Service Call - enables mocking for Time-Travel simulation
	projects, err := a.projects.ListActiveProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active projects: %w", err)
	}

	for _, p := range projects {
		if err := a.processProject(ctx, p); err != nil {
			log.Printf("ERROR processing project %s: %v", p.ID, err)
			// Continue with other projects even if one fails
			continue
		}
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
			log.Printf("WARN: geocoding failed for project %s, using default coords: %v", p.ID, err)
		}
	}
	forecast, err := a.weather.GetForecast(lat, lng)
	if err != nil {
		log.Printf("WARN: failed to get weather for project %s: %v", p.ID, err)
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
		part := &genai.Part{Text: prompt}
		briefing, err = a.aiClient.GenerateContent(ctx, ai.ModelTypeFlash, part)
		if err != nil {
			return fmt.Errorf("AI generation failed: %w", err)
		}
	} else {
		log.Printf("WARN: AI client not available for project %s, skipping briefing generation", p.ID)
		briefing = "[AI Unavailable] Manual briefing required. Please review project schedule and weather conditions."
	}

	// 5. Deliver - Critical Blocker A Remediation: Dynamic PM email lookup
	log.Printf("--- DAILY BRIEFING FOR %s ---\n%s\n-----------------------------", p.Name, briefing)

	// Look up Project Manager contact via DirectoryService (fallback to generic)
	recipientEmail := "superintendent@futurebuild.sh"
	if a.directory != nil {
		if contact, err := a.directory.GetProjectManager(ctx, p.ID, p.OrgID); err == nil && contact.Email != "" {
			recipientEmail = contact.Email
		} else {
			log.Printf("WARN: PM lookup failed for project %s, using fallback email: %v", p.ID, err)
		}
	}
	if err := a.notifier.SendEmail(recipientEmail, fmt.Sprintf("Daily Briefing: %s", p.Name), briefing); err != nil {
		log.Printf("WARN: failed to send notification: %v", err)
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

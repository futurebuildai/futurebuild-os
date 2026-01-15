package agents

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/genai"
)

// DailyFocusAgent orchestrates the morning briefing generation.
type DailyFocusAgent struct {
	db       *pgxpool.Pool
	schedule *service.ScheduleService
	weather  types.WeatherService
	notifier types.NotificationService
	aiClient ai.Client
}

// NewDailyFocusAgent creates a new agent instance.
func NewDailyFocusAgent(
	db *pgxpool.Pool,
	schedule *service.ScheduleService,
	weather types.WeatherService,
	notifier types.NotificationService,
	aiClient ai.Client,
) *DailyFocusAgent {
	return &DailyFocusAgent{
		db:       db,
		schedule: schedule,
		weather:  weather,
		notifier: notifier,
		aiClient: aiClient,
	}
}

// Execute runs the daily briefing logic for all active projects.
func (a *DailyFocusAgent) Execute(ctx context.Context) error {
	log.Println("Starting Daily Focus Agent...")

	// 1. Fetch Active Projects
	// We query the DB directly here for batch processing, skipping service layer for efficiency if needed,
	// but strictly we should use service. However, ProjectService doesn't have ListProjects yet.
	// For now, simple query.
	projects, err := a.fetchActiveProjects(ctx)
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

func (a *DailyFocusAgent) fetchActiveProjects(ctx context.Context) ([]models.Project, error) {
	// Helper query
	// Ideally this belongs in a Repository
	query := `
		SELECT id, org_id, name, address, permit_issued_date, target_end_date, status
		FROM projects
		WHERE status IN ('Active', 'Preconstruction') -- Include Precon for planning
	`
	rows, err := a.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Address, &p.PermitIssuedDate, &p.TargetEndDate, &p.Status)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (a *DailyFocusAgent) processProject(ctx context.Context, p models.Project) error {
	// 2. Fetch Context Data
	// A. Weather
	// TODO: Geocode p.Address to get lat/long. For MVP, using Austin, TX coordinates.
	// See BACKEND_SCOPE.md Section 2.4 (Weather-Sensitive Phases)
	forecast, err := a.weather.GetForecast(30.2672, -97.7431)
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

	// Let's implement a specific query here for the agent's needs.
	tasks, err := a.fetchRelevantTasks(ctx, p.ID)
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

	// 5. Deliver
	// Send to "Project Owner" (stubbed as generic email for now)
	log.Printf("--- DAILY BRIEFING FOR %s ---\n%s\n-----------------------------", p.Name, briefing)

	// In production, look up Project Manager contact via DirectoryService
	if err := a.notifier.SendEmail("superintendent@futurebuild.sh", fmt.Sprintf("Daily Briefing: %s", p.Name), briefing); err != nil {
		log.Printf("WARN: failed to send notification: %v", err)
	}

	return nil
}

func (a *DailyFocusAgent) fetchRelevantTasks(ctx context.Context, projectID uuid.UUID) ([]models.ProjectTask, error) {
	// Select tasks that are:
	// 1. In Progress
	// 2. Scheduled to start/end this week
	// 3. On Critical Path and not complete
	// Limit to 20 to fit context window comfortably.
	// See BACKEND_SCOPE.md Section 3.4 (Live Project Graph)
	//
	// NOTE: Status literals must match DB enum values. See pkg/types/enums.go.
	// DB uses lowercase, types use Title_Case. Queries use lowercase per schema.
	query := `
        SELECT id, name, status, planned_start, planned_end, is_on_critical_path
        FROM project_tasks
        WHERE project_id = $1
          AND (
              status = 'in_progress'
              OR (status = 'pending' AND planned_start <= CURRENT_DATE + INTERVAL '7 days')
              OR (is_on_critical_path = true AND status != 'completed')
          )
        ORDER BY planned_start ASC
        LIMIT 20
    `
	// Note: using minimal struct scan or map
	rows, err := a.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.ProjectTask
	for rows.Next() {
		var t models.ProjectTask
		// We only fetch a subset of fields, so we need to be careful with the full struct scan
		// OR just scan what we asked for. The struct has more fields.
		// Let's scan into variables and populate the struct.
		err := rows.Scan(&t.ID, &t.Name, &t.Status, &t.PlannedStart, &t.PlannedEnd, &t.IsOnCriticalPath)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
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
`, p.Name, time.Now().Format("Monday, Jan 02"), p.Address,
		w.HighTempC, w.LowTempC, w.PrecipitationMM, w.PrecipitationProbability*100, w.Conditions,
		taskContext)
}

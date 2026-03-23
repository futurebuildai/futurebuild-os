package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/data"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// SchedulePreviewService generates instant schedule previews from onboarding data.
// It runs the full physics pipeline (WBS scoping → DHSM → CPM) without a database.
type SchedulePreviewService struct {
	weatherSvc     WeatherServicer
	physicsCfg     config.PhysicsConfig
	calibrationSvc CalibrationServicer
}

// NewSchedulePreviewService creates a new preview service.
// weatherSvc may be nil; weather overlay will be skipped if unavailable.
func NewSchedulePreviewService(
	weatherSvc WeatherServicer,
	physicsCfg config.PhysicsConfig,
) *SchedulePreviewService {
	return &SchedulePreviewService{
		weatherSvc: weatherSvc,
		physicsCfg: physicsCfg.WithDefaults(),
	}
}

// WithCalibration injects the CalibrationService for org-learned duration multipliers.
func (s *SchedulePreviewService) WithCalibration(svc CalibrationServicer) *SchedulePreviewService {
	s.calibrationSvc = svc
	return s
}

// SchedulePreviewRequest is the input for generating a schedule preview.
type SchedulePreviewRequest struct {
	SquareFootage  float64              `json:"square_footage"`
	FoundationType string               `json:"foundation_type"`
	StartDate      string               `json:"start_date"` // YYYY-MM-DD
	Stories        int                  `json:"stories"`
	Address        string               `json:"address,omitempty"`
	Latitude       float64              `json:"latitude,omitempty"`
	Longitude      float64              `json:"longitude,omitempty"`
	Topography     string               `json:"topography,omitempty"`
	SoilConditions string               `json:"soil_conditions,omitempty"`
	Bedrooms       int                  `json:"bedrooms,omitempty"`
	Bathrooms      int                  `json:"bathrooms,omitempty"`
	LongLeadItems  []models.LongLeadItem `json:"long_lead_items,omitempty"`

	// Org-level calibration support (optional)
	OrgID *uuid.UUID `json:"org_id,omitempty"`

	// In-progress project support
	IsInProgress    bool                  `json:"is_in_progress"`
	CompletedPhases []CompletedPhaseInput `json:"completed_phases,omitempty"`
	CurrentDate     string                `json:"current_date,omitempty"` // defaults to today
}

// CompletedPhaseInput captures actual completion data for in-progress projects.
type CompletedPhaseInput struct {
	WBSCode   string `json:"wbs_code"`   // e.g. "8.0" or "8.x" for whole phase
	ActualEnd string `json:"actual_end"` // YYYY-MM-DD
	Status    string `json:"status"`     // "completed" or "in_progress"
}

// SchedulePreviewResponse is the output of the preview pipeline.
type SchedulePreviewResponse struct {
	ProjectedEnd     string            `json:"projected_end"`
	TotalWorkingDays int               `json:"total_working_days"`
	RemainingDays    int               `json:"remaining_days"`
	CriticalPath     []string          `json:"critical_path"`
	PhaseTimeline    []PhasePreview    `json:"phase_timeline"`
	ProcurementDates []ProcurementDate `json:"procurement_dates,omitempty"`
	WeatherImpact    *WeatherImpact    `json:"weather_impact,omitempty"`
	ScopeChanges     []physics.ScopeChange `json:"scope_changes"`
	TradeGaps        []TradeGap        `json:"trade_gaps,omitempty"`
	GanttPreview     *types.GanttData  `json:"gantt_preview"`
	CompletionPercent float64          `json:"completion_percent"`
}

// PhasePreview summarizes a single phase in the schedule preview.
type PhasePreview struct {
	PhaseName    string `json:"phase_name"`
	WBSCode      string `json:"wbs_code"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	DurationDays int    `json:"duration_days"`
	IsCritical   bool   `json:"is_critical"`
	Status       string `json:"status"` // pending, in_progress, completed
}

// ProcurementDate represents order-by timing for a long-lead material.
type ProcurementDate struct {
	ItemName    string `json:"item_name"`
	Brand       string `json:"brand,omitempty"`
	WBSCode     string `json:"wbs_code"`
	LeadWeeks   int    `json:"lead_weeks"`
	OrderByDate string `json:"order_by_date"`
	InstallDate string `json:"install_date"`
	Status      string `json:"status"` // overdue, urgent, upcoming, ok
}

// WeatherImpact captures how weather affects the schedule.
type WeatherImpact struct {
	AffectedPhases []WeatherPhaseImpact `json:"affected_phases"`
	TotalExtraDays int                  `json:"total_extra_days"`
	RiskMonths     []string             `json:"risk_months"`
	Summary        string               `json:"summary"`
}

// WeatherPhaseImpact is a single phase affected by weather.
type WeatherPhaseImpact struct {
	PhaseName string  `json:"phase_name"`
	WBSCode   string  `json:"wbs_code"`
	ExtraDays float64 `json:"extra_days"`
	Reason    string  `json:"reason"`
}

// TradeGap identifies a missing subcontractor for a required trade.
type TradeGap struct {
	PhaseName     string `json:"phase_name"`
	WBSCode       string `json:"wbs_code"`
	RequiredTrade string `json:"required_trade"`
	StartDate     string `json:"start_date"`
	HasContact    bool   `json:"has_contact"`
	ContactName   string `json:"contact_name,omitempty"`
}

// ScenarioComparisonRequest runs multiple previews for what-if analysis.
type ScenarioComparisonRequest struct {
	Base         SchedulePreviewRequest   `json:"base"`
	Alternatives []SchedulePreviewRequest `json:"alternatives"` // max 3
}

// ScenarioComparisonResponse contains all scenario results.
type ScenarioComparisonResponse struct {
	Scenarios []ScenarioResult `json:"scenarios"`
}

// ScenarioResult is a single scenario with delta calculations.
type ScenarioResult struct {
	Label            string                  `json:"label"`
	Preview          SchedulePreviewResponse `json:"preview"`
	DeltaDays        int                     `json:"delta_days"`
	DeltaCostCents   int64                   `json:"delta_cost_cents"`
	CriticalPathDiff []string                `json:"critical_path_diff"`
}

// GeneratePreview runs the full physics pipeline and returns a schedule preview.
func (s *SchedulePreviewService) GeneratePreview(req SchedulePreviewRequest) (*SchedulePreviewResponse, error) {
	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	// Parse current date for in-progress (defaults to today)
	currentDate := time.Now().Truncate(24 * time.Hour)
	if req.CurrentDate != "" {
		cd, err := time.Parse("2006-01-02", req.CurrentDate)
		if err != nil {
			return nil, fmt.Errorf("invalid current_date: %w", err)
		}
		currentDate = cd
	}

	// Step 1: Load WBS master template
	masterTasks, masterDeps, err := data.LoadMasterWBS()
	if err != nil {
		return nil, fmt.Errorf("failed to load WBS: %w", err)
	}

	// Step 2: Build scope context from request
	var completedCodes []string
	if req.IsInProgress {
		for _, cp := range req.CompletedPhases {
			completedCodes = append(completedCodes, cp.WBSCode)
		}
	}

	scopeCtx := physics.ProjectScopeContext{
		FoundationType:    req.FoundationType,
		Stories:           req.Stories,
		GSF:               req.SquareFootage,
		Bedrooms:          req.Bedrooms,
		Bathrooms:         req.Bathrooms,
		Topography:        req.Topography,
		SoilConditions:    req.SoilConditions,
		CompletedWBSCodes: completedCodes,
	}

	// Step 3: Apply scoping rules
	scopedTasks, scopedDeps, scopeChanges := physics.ApplyScope(masterTasks, masterDeps, scopeCtx)

	// Step 4: Calculate DHSM durations
	wbsTasks := toModelWBSTasks(scopedTasks)
	forecast := s.getWeatherForecast(req)

	// Load org multipliers if OrgID is provided and CalibrationService is available
	var orgMultipliers []models.DurationMultiplier
	if req.OrgID != nil && s.calibrationSvc != nil {
		multiplierMap, mErr := s.calibrationSvc.GetOrgMultiplierMap(context.Background(), *req.OrgID)
		if mErr == nil && len(multiplierMap) > 0 {
			for wbs, weight := range multiplierMap {
				orgMultipliers = append(orgMultipliers, models.DurationMultiplier{
					WBSTaskCode: wbs,
					VariableKey: "org_calibration",
					Weight:      weight,
				})
			}
		}
	}

	durations := physics.CalculateBatchDurationsV2(
		wbsTasks,
		req.SquareFootage,
		models.ProjectContext{}, // Default context for preview
		orgMultipliers,          // Org-learned multipliers (nil if no org)
		forecast,
		s.physicsCfg,
	)

	// Step 5: Build CPM graph
	projectTasks, taskDeps, codeToID := buildPreviewGraph(scopedTasks, scopedDeps, durations)

	// Step 6: Set up material constraints (procurement + completed task dates)
	materialConstraints := make(map[uuid.UUID]time.Time)

	// Procurement constraints: long-lead items
	procDates := s.calculateProcurementDates(req.LongLeadItems, scopedTasks, codeToID, startDate, currentDate, req.IsInProgress, completedCodes)
	for _, pd := range procDates {
		installDate, err := time.Parse("2006-01-02", pd.InstallDate)
		if err != nil {
			continue
		}
		if taskID, ok := codeToID[pd.WBSCode]; ok {
			if existing, exists := materialConstraints[taskID]; !exists || installDate.After(existing) {
				materialConstraints[taskID] = installDate
			}
		}
	}

	// In-progress constraints: completed tasks have fixed dates
	completedDates := make(map[string]time.Time)
	if req.IsInProgress {
		expandedCompleted := physics.CompletedTaskCodes(completedCodes, scopedTasks)
		for _, cp := range req.CompletedPhases {
			if cp.ActualEnd == "" {
				continue
			}
			actualEnd, err := time.Parse("2006-01-02", cp.ActualEnd)
			if err != nil {
				continue
			}

			// Expand phase wildcards to individual codes
			if strings.HasSuffix(cp.WBSCode, ".x") {
				prefix := strings.TrimSuffix(cp.WBSCode, "x")
				for _, code := range expandedCompleted {
					if strings.HasPrefix(code, prefix) {
						completedDates[code] = actualEnd
					}
				}
			} else {
				completedDates[cp.WBSCode] = actualEnd
			}
		}

		// For completed tasks, constrain successors to start no earlier than completion date
		for code, finishDate := range completedDates {
			if taskID, ok := codeToID[code]; ok {
				// Set the completed task's actual end as a constraint for scheduling
				materialConstraints[taskID] = finishDate
			}
		}

		// Also constrain remaining tasks to start no earlier than today
		expandedSet := make(map[string]bool, len(expandedCompleted))
		for _, c := range expandedCompleted {
			expandedSet[c] = true
		}
		for _, task := range projectTasks {
			if !expandedSet[task.WBSCode] {
				// Pending task: ensure it doesn't start before today
				if existing, exists := materialConstraints[task.ID]; !exists || currentDate.After(existing) {
					materialConstraints[task.ID] = currentDate
				}
			}
		}
	}

	// Step 7: Run CPM Forward Pass
	cal := &physics.StandardCalendar{}
	graph := physics.BuildDependencyGraph(projectTasks, taskDeps)
	schedule, err := physics.ForwardPass(graph, startDate, cal, materialConstraints)
	if err != nil {
		return nil, fmt.Errorf("CPM forward pass failed: %w", err)
	}

	// Step 8: Run CPM Backward Pass
	criticalPath, err := physics.BackwardPass(graph, schedule, cal, nil)
	if err != nil {
		return nil, fmt.Errorf("CPM backward pass failed: %w", err)
	}

	// Step 9: Calculate weather impact (comparison without weather)
	weatherImpact := s.calculateWeatherImpact(scopedTasks, durations, forecast, req.IsInProgress, completedCodes)

	// Step 10: Build response
	return s.buildResponse(
		schedule, criticalPath, scopedTasks, scopedDeps,
		procDates, weatherImpact, scopeChanges,
		startDate, currentDate, req.IsInProgress, completedCodes,
		codeToID,
	), nil
}

// CompareScenarios runs multiple previews and calculates deltas.
func (s *SchedulePreviewService) CompareScenarios(req ScenarioComparisonRequest) (*ScenarioComparisonResponse, error) {
	if len(req.Alternatives) > 3 {
		return nil, fmt.Errorf("maximum 3 alternative scenarios allowed")
	}

	basePreview, err := s.GeneratePreview(req.Base)
	if err != nil {
		return nil, fmt.Errorf("base scenario failed: %w", err)
	}

	baseEnd, _ := time.Parse("2006-01-02", basePreview.ProjectedEnd)
	results := []ScenarioResult{{
		Label:   "Base",
		Preview: *basePreview,
	}}

	for i, alt := range req.Alternatives {
		altPreview, err := s.GeneratePreview(alt)
		if err != nil {
			return nil, fmt.Errorf("alternative %d failed: %w", i+1, err)
		}

		altEnd, _ := time.Parse("2006-01-02", altPreview.ProjectedEnd)
		deltaDays := int(altEnd.Sub(baseEnd).Hours() / 24)

		// Compute critical path diff
		baseCPSet := make(map[string]bool, len(basePreview.CriticalPath))
		for _, cp := range basePreview.CriticalPath {
			baseCPSet[cp] = true
		}
		var diff []string
		for _, cp := range altPreview.CriticalPath {
			if !baseCPSet[cp] {
				diff = append(diff, "+"+cp)
			}
		}
		altCPSet := make(map[string]bool, len(altPreview.CriticalPath))
		for _, a := range altPreview.CriticalPath {
			altCPSet[a] = true
		}
		for _, cp := range basePreview.CriticalPath {
			if !altCPSet[cp] {
				diff = append(diff, "-"+cp)
			}
		}

		// Calculate cost delta between base and alternative scenarios
		var deltaCost int64
		if alt.SquareFootage != req.Base.SquareFootage {
			baseCost := data.CalculateProjectCost(req.Base.SquareFootage, req.Base.FoundationType, req.Base.Stories, "")
			altCost := data.CalculateProjectCost(alt.SquareFootage, alt.FoundationType, alt.Stories, "")
			deltaCost = altCost - baseCost
		}

		results = append(results, ScenarioResult{
			Label:            fmt.Sprintf("Alternative %d", i+1),
			Preview:          *altPreview,
			DeltaDays:        deltaDays,
			DeltaCostCents:   deltaCost,
			CriticalPathDiff: diff,
		})
	}

	return &ScenarioComparisonResponse{Scenarios: results}, nil
}

// getWeatherForecast retrieves forecast data if coordinates are available.
func (s *SchedulePreviewService) getWeatherForecast(req SchedulePreviewRequest) types.Forecast {
	if s.weatherSvc == nil || (req.Latitude == 0 && req.Longitude == 0) {
		return types.Forecast{} // No weather data — will not affect durations
	}

	forecast, err := s.weatherSvc.GetForecast(req.Latitude, req.Longitude)
	if err != nil {
		return types.Forecast{} // Graceful degradation
	}
	return forecast
}

// toModelWBSTasks converts data.WBSTask → models.WBSTask for DHSM compatibility.
func toModelWBSTasks(tasks []data.WBSTask) []models.WBSTask {
	result := make([]models.WBSTask, len(tasks))
	for i, t := range tasks {
		result[i] = models.WBSTask{
			Code:             t.Code,
			Name:             t.Name,
			BaseDurationDays: t.BaseDurationDays,
			ResponsibleParty: t.ResponsibleParty,
			IsInspection:     t.IsInspection,
			IsLongLead:       t.IsLongLead,
			LeadTimeWeeksMin: t.LeadTimeWeeksMin,
			LeadTimeWeeksMax: t.LeadTimeWeeksMax,
		}
	}
	return result
}

// buildPreviewGraph converts scoped tasks into CPM-compatible structures.
func buildPreviewGraph(
	tasks []data.WBSTask,
	deps []data.WBSDependency,
	durations []physics.DHSMResultV2,
) ([]models.ProjectTask, []models.TaskDependency, map[string]uuid.UUID) {
	codeToID := make(map[string]uuid.UUID, len(tasks))
	projectID := uuid.New() // Ephemeral project ID for preview

	// Build duration map
	durationMap := make(map[string]float64)
	for _, d := range durations {
		durationMap[d.WBSCode] = physics.DurationToDays(d.CalculatedDuration)
	}

	var projectTasks []models.ProjectTask
	for _, t := range tasks {
		id := uuid.New()
		codeToID[t.Code] = id

		dur := t.BaseDurationDays
		if d, ok := durationMap[t.Code]; ok && d > 0 {
			dur = d
		}
		// CPM requires CalculatedDuration > 0 (fail-loudly design).
		// Milestones and zero-duration tasks get a 0.5-day minimum quantum.
		if dur <= 0 {
			dur = 0.5
		}

		projectTasks = append(projectTasks, models.ProjectTask{
			ID:                 id,
			ProjectID:          projectID,
			WBSCode:            t.Code,
			Name:               t.Name,
			IsInspection:       t.IsInspection,
			CalculatedDuration: dur,
			Status:             types.TaskStatusPending,
		})
	}

	var taskDeps []models.TaskDependency
	for _, d := range deps {
		predID, predOK := codeToID[d.PredecessorCode]
		succID, succOK := codeToID[d.SuccessorCode]
		if !predOK || !succOK {
			continue
		}
		taskDeps = append(taskDeps, models.TaskDependency{
			ID:             uuid.New(),
			ProjectID:      projectID,
			PredecessorID:  predID,
			SuccessorID:    succID,
			DependencyType: d.Type,
			LagDays:        d.LagDays,
		})
	}

	return projectTasks, taskDeps, codeToID
}

// calculateProcurementDates computes order-by dates for long-lead materials.
func (s *SchedulePreviewService) calculateProcurementDates(
	longLeadItems []models.LongLeadItem,
	tasks []data.WBSTask,
	codeToID map[string]uuid.UUID,
	startDate, currentDate time.Time,
	isInProgress bool,
	completedCodes []string,
) []ProcurementDate {
	var results []ProcurementDate

	for _, item := range longLeadItems {
		// Skip procurement for completed installation tasks
		if isInProgress && physics.IsTaskCompleted(item.WBSCode, completedCodes) {
			continue
		}

		// Find installation task to determine install date
		installWBS := item.WBSCode
		if installWBS == "" {
			installWBS = mapCategoryToWBS(item.Category)
		}

		// Estimate install date based on task position in sequence
		// For preview, use startDate + rough offset based on WBS position
		installDate := estimateInstallDate(installWBS, startDate, tasks)

		leadWeeks := item.EstimatedLeadWeeks
		if leadWeeks <= 0 {
			leadWeeks = lookupBrandLeadTime(item.Brand, item.Category)
		}

		orderByDate := installDate.AddDate(0, 0, -leadWeeks*7)

		status := "ok"
		if orderByDate.Before(currentDate) {
			status = "overdue"
		} else if orderByDate.Before(currentDate.AddDate(0, 0, 14)) {
			status = "urgent"
		} else if orderByDate.Before(currentDate.AddDate(0, 0, 30)) {
			status = "upcoming"
		}

		results = append(results, ProcurementDate{
			ItemName:    item.Name,
			Brand:       item.Brand,
			WBSCode:     installWBS,
			LeadWeeks:   leadWeeks,
			OrderByDate: orderByDate.Format("2006-01-02"),
			InstallDate: installDate.Format("2006-01-02"),
			Status:      status,
		})
	}

	return results
}

// calculateWeatherImpact estimates extra days due to weather.
func (s *SchedulePreviewService) calculateWeatherImpact(
	tasks []data.WBSTask,
	durations []physics.DHSMResultV2,
	forecast types.Forecast,
	isInProgress bool,
	completedCodes []string,
) *WeatherImpact {
	// No weather data → no impact
	if forecast.PrecipitationMM == 0 && forecast.LowTempC == 0 && forecast.HighTempC == 0 {
		return nil
	}

	var affected []WeatherPhaseImpact
	totalExtra := 0.0

	durationMap := make(map[string]float64)
	for _, d := range durations {
		durationMap[d.WBSCode] = physics.DurationToDays(d.BaseDuration)
	}

	for _, task := range tasks {
		// Skip completed tasks for in-progress
		if isInProgress && physics.IsTaskCompleted(task.Code, completedCodes) {
			continue
		}

		baseDur := durationMap[task.Code]
		if baseDur <= 0 {
			continue
		}

		tempTask := models.ProjectTask{
			WBSCode:            task.Code,
			CalculatedDuration: baseDur,
		}
		adjusted := physics.ApplyWeatherAdjustment(tempTask, forecast)
		extra := adjusted - baseDur

		if extra > 0 {
			reason := buildWeatherReason(forecast)
			affected = append(affected, WeatherPhaseImpact{
				PhaseName: task.Name,
				WBSCode:   task.Code,
				ExtraDays: extra,
				Reason:    reason,
			})
			totalExtra += extra
		}
	}

	if len(affected) == 0 {
		return nil
	}

	// Identify risk months based on when weather-sensitive tasks fall
	riskMonths := identifyRiskMonths(forecast)

	return &WeatherImpact{
		AffectedPhases: affected,
		TotalExtraDays: int(totalExtra + 0.5), // Round
		RiskMonths:     riskMonths,
		Summary:        fmt.Sprintf("Weather conditions may add ~%d working days to the schedule", int(totalExtra+0.5)),
	}
}

// buildResponse assembles the final preview response from CPM results.
func (s *SchedulePreviewService) buildResponse(
	schedule map[uuid.UUID]physics.TaskSchedule,
	criticalPath []string,
	tasks []data.WBSTask,
	deps []data.WBSDependency,
	procDates []ProcurementDate,
	weatherImpact *WeatherImpact,
	scopeChanges []physics.ScopeChange,
	startDate, currentDate time.Time,
	isInProgress bool,
	completedCodes []string,
	codeToID map[string]uuid.UUID,
) *SchedulePreviewResponse {
	// Find project end date
	var projectEnd time.Time
	for _, sched := range schedule {
		if sched.EarlyFinish.After(projectEnd) {
			projectEnd = sched.EarlyFinish
		}
	}

	totalDays := int(projectEnd.Sub(startDate).Hours()/24) + 1
	remainingDays := totalDays
	if isInProgress {
		remainingDays = int(projectEnd.Sub(currentDate).Hours()/24) + 1
		if remainingDays < 0 {
			remainingDays = 0
		}
	}

	// Build phase timeline
	phases, _ := data.LoadMasterWBSPhases()
	phaseTimeline := buildPhaseTimeline(phases, schedule, codeToID, criticalPath, isInProgress, completedCodes)

	// Compute completion percent
	completionPercent := 0.0
	if isInProgress && len(tasks) > 0 {
		expandedCompleted := physics.CompletedTaskCodes(completedCodes, tasks)
		completionPercent = float64(len(expandedCompleted)) / float64(len(tasks)) * 100
	}

	// Build Gantt preview
	gantt := buildGanttPreview(tasks, deps, schedule, criticalPath, codeToID, isInProgress, completedCodes, projectEnd)

	critPathCopy := make([]string, len(criticalPath))
	copy(critPathCopy, criticalPath)

	// Detect trade gaps (required trades without known contacts)
	tradeGaps := detectTradeGaps(tasks, schedule, codeToID, isInProgress, completedCodes)

	return &SchedulePreviewResponse{
		ProjectedEnd:      projectEnd.Format("2006-01-02"),
		TotalWorkingDays:  totalDays,
		RemainingDays:     remainingDays,
		CriticalPath:      critPathCopy,
		PhaseTimeline:     phaseTimeline,
		ProcurementDates:  procDates,
		WeatherImpact:     weatherImpact,
		ScopeChanges:      scopeChanges,
		TradeGaps:         tradeGaps,
		GanttPreview:      gantt,
		CompletionPercent: completionPercent,
	}
}

// buildPhaseTimeline groups tasks into phase-level summaries.
func buildPhaseTimeline(
	phases []data.WBSPhase,
	schedule map[uuid.UUID]physics.TaskSchedule,
	codeToID map[string]uuid.UUID,
	criticalPath []string,
	isInProgress bool,
	completedCodes []string,
) []PhasePreview {
	critPathSet := make(map[string]bool, len(criticalPath))
	for _, cp := range criticalPath {
		critPathSet[cp] = true
	}

	var timeline []PhasePreview
	for _, phase := range phases {
		var phaseStart, phaseEnd time.Time
		hasCritical := false
		first := true

		for _, task := range phase.Tasks {
			taskID, ok := codeToID[task.Code]
			if !ok {
				continue
			}
			sched, ok := schedule[taskID]
			if !ok {
				continue
			}

			if first || sched.EarlyStart.Before(phaseStart) {
				phaseStart = sched.EarlyStart
			}
			if first || sched.EarlyFinish.After(phaseEnd) {
				phaseEnd = sched.EarlyFinish
			}
			first = false

			if critPathSet[task.Code] {
				hasCritical = true
			}
		}

		if first {
			continue // No scheduled tasks in this phase
		}

		status := "pending"
		if isInProgress {
			allCompleted := true
			anyCompleted := false
			for _, task := range phase.Tasks {
				if physics.IsTaskCompleted(task.Code, completedCodes) {
					anyCompleted = true
				} else {
					allCompleted = false
				}
			}
			if allCompleted {
				status = "completed"
			} else if anyCompleted {
				status = "in_progress"
			}
		}

		duration := int(phaseEnd.Sub(phaseStart).Hours()/24) + 1
		timeline = append(timeline, PhasePreview{
			PhaseName:    phase.Name,
			WBSCode:      phase.Code,
			StartDate:    phaseStart.Format("2006-01-02"),
			EndDate:      phaseEnd.Format("2006-01-02"),
			DurationDays: duration,
			IsCritical:   hasCritical,
			Status:       status,
		})
	}

	return timeline
}

// buildGanttPreview converts CPM results into a GanttData structure for the frontend.
func buildGanttPreview(
	tasks []data.WBSTask,
	deps []data.WBSDependency,
	schedule map[uuid.UUID]physics.TaskSchedule,
	criticalPath []string,
	codeToID map[string]uuid.UUID,
	isInProgress bool,
	completedCodes []string,
	projectEnd time.Time,
) *types.GanttData {
	critPathSet := make(map[string]bool, len(criticalPath))
	for _, cp := range criticalPath {
		critPathSet[cp] = true
	}

	var ganttTasks []types.GanttTask
	for _, task := range tasks {
		taskID, ok := codeToID[task.Code]
		if !ok {
			continue
		}
		sched, ok := schedule[taskID]
		if !ok {
			continue
		}

		status := types.TaskStatusPending
		if isInProgress && physics.IsTaskCompleted(task.Code, completedCodes) {
			status = types.TaskStatusCompleted
		}

		dur := sched.EarlyFinish.Sub(sched.EarlyStart).Hours() / 24
		ganttTasks = append(ganttTasks, types.GanttTask{
			WBSCode:      task.Code,
			Name:         task.Name,
			Status:       status,
			EarlyStart:   sched.EarlyStart.Format("2006-01-02"),
			EarlyFinish:  sched.EarlyFinish.Format("2006-01-02"),
			DurationDays: dur,
			IsCritical:   critPathSet[task.Code],
		})
	}

	var ganttDeps []types.GanttDependency
	for _, d := range deps {
		ganttDeps = append(ganttDeps, types.GanttDependency{
			From: d.PredecessorCode,
			To:   d.SuccessorCode,
		})
	}

	return &types.GanttData{
		CalculatedAt:     time.Now().Format(time.RFC3339),
		ProjectedEndDate: projectEnd.Format("2006-01-02"),
		CriticalPath:     criticalPath,
		Tasks:            ganttTasks,
		Dependencies:     ganttDeps,
	}
}

// mapCategoryToWBS maps a long-lead item category to the typical installation WBS code.
func mapCategoryToWBS(category string) string {
	switch strings.ToLower(category) {
	case "windows", "doors":
		return "9.5" // Window & Door Install
	case "hvac":
		return "10.2" // HVAC Rough-In
	case "appliances":
		return "12.6" // Appliance Install
	case "millwork", "cabinetry":
		return "12.0" // Cabinet Install
	case "roofing":
		return "9.7" // Roofing
	case "plumbing":
		return "10.0" // Plumbing Rough-In
	case "electrical":
		return "10.1" // Electrical Rough-In
	default:
		return "12.0" // Default to interior finishes
	}
}

// lookupBrandLeadTime returns estimated lead weeks for known brands.
func lookupBrandLeadTime(brand, category string) int {
	knownLeadTimes := models.KnownBrandLeadTimes()
	if weeks, ok := knownLeadTimes[strings.ToLower(brand)]; ok {
		return weeks
	}
	// Default lead times by category
	switch strings.ToLower(category) {
	case "windows":
		return 8
	case "doors":
		return 6
	case "hvac":
		return 4
	case "appliances":
		return 6
	case "cabinetry", "millwork":
		return 10
	default:
		return 6
	}
}

// estimateInstallDate estimates when an installation task will occur based on WBS position.
func estimateInstallDate(wbsCode string, projectStart time.Time, tasks []data.WBSTask) time.Time {
	// Find position of install task in sequence
	sortedTasks := make([]data.WBSTask, len(tasks))
	copy(sortedTasks, tasks)
	sort.Slice(sortedTasks, func(i, j int) bool {
		return compareWBSCodes(sortedTasks[i].Code, sortedTasks[j].Code) < 0
	})

	totalDuration := 0.0
	for _, t := range sortedTasks {
		totalDuration += t.BaseDurationDays
		if t.Code == wbsCode {
			break
		}
	}

	cal := &physics.StandardCalendar{}
	return cal.AddWorkDuration(projectStart, time.Duration(totalDuration*float64(24*time.Hour)))
}

// buildWeatherReason constructs a human-readable weather impact explanation.
func buildWeatherReason(forecast types.Forecast) string {
	var reasons []string
	if forecast.PrecipitationMM > 10.0 {
		reasons = append(reasons, "precipitation")
	}
	if forecast.LowTempC < 0.0 {
		reasons = append(reasons, "freezing temperatures")
	}
	if forecast.HighTempC > 35.0 {
		reasons = append(reasons, "extreme heat")
	}
	if len(reasons) == 0 {
		return "weather conditions"
	}
	return strings.Join(reasons, ", ")
}

// detectTradeGaps identifies required trades for remaining phases.
// During preview (pre-project creation), we flag all unique trades needed
// so the builder knows which subcontractors to line up.
func detectTradeGaps(
	tasks []data.WBSTask,
	schedule map[uuid.UUID]physics.TaskSchedule,
	codeToID map[string]uuid.UUID,
	isInProgress bool,
	completedCodes []string,
) []TradeGap {
	// Build set of completed tasks
	completedSet := make(map[string]bool)
	if isInProgress {
		expanded := physics.CompletedTaskCodes(completedCodes, tasks)
		for _, c := range expanded {
			completedSet[c] = true
		}
	}

	// Deduplicate by trade name — one gap per unique trade
	seenTrades := make(map[string]bool)
	var gaps []TradeGap

	for _, task := range tasks {
		// Skip completed tasks
		if completedSet[task.Code] {
			continue
		}

		trade := task.ResponsibleParty
		if trade == "" || trade == "GC" || trade == "gc" || trade == "General Contractor" {
			continue // GC-managed tasks don't need a sub
		}

		if seenTrades[trade] {
			continue
		}
		seenTrades[trade] = true

		// Find the earliest start date for this trade
		startDateStr := ""
		if taskID, ok := codeToID[task.Code]; ok {
			if sched, exists := schedule[taskID]; exists {
				startDateStr = sched.EarlyStart.Format("2006-01-02")
			}
		}

		gaps = append(gaps, TradeGap{
			PhaseName:     task.Name,
			WBSCode:       task.Code,
			RequiredTrade: trade,
			StartDate:     startDateStr,
			HasContact:    false, // Pre-creation: no directory to check
		})
	}

	// Sort by start date
	sort.Slice(gaps, func(i, j int) bool {
		return gaps[i].StartDate < gaps[j].StartDate
	})

	return gaps
}

// compareWBSCodes compares two WBS codes numerically (e.g., "9.5" < "10.0").
func compareWBSCodes(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aNum, aErr := strconv.Atoi(aParts[i])
		bNum, bErr := strconv.Atoi(bParts[i])
		if aErr != nil || bErr != nil {
			// Fallback to lexicographic for non-numeric segments
			if aParts[i] < bParts[i] {
				return -1
			}
			if aParts[i] > bParts[i] {
				return 1
			}
			continue
		}
		if aNum != bNum {
			return aNum - bNum
		}
	}
	return len(aParts) - len(bParts)
}

// identifyRiskMonths returns months that have adverse weather conditions.
func identifyRiskMonths(forecast types.Forecast) []string {
	var months []string
	if forecast.LowTempC < 0.0 {
		months = append(months, "December", "January", "February")
	}
	if forecast.PrecipitationMM > 10.0 {
		months = append(months, "March", "April", "November")
	}
	if forecast.HighTempC > 35.0 {
		months = append(months, "July", "August")
	}
	return months
}

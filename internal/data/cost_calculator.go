package data

// CalculateProjectCost estimates total project cost in cents from core attributes.
// Aggregates NationalCostIndices × RegionalMultiplier × GSF, with foundation and
// stories adjustments applied. Shared by budget tools and market-aware cost features.
func CalculateProjectCost(gsf float64, foundation string, stories int, region string) int64 {
	if gsf <= 0 {
		return 0
	}
	baseCostPerSqFt := TotalNationalCostPerSqFtCents()

	// Apply regional multiplier
	regionalMult := 1.0
	if m, ok := RegionalMultipliers()[region]; ok {
		regionalMult = m
	}

	// Apply foundation and stories adjustments
	foundationMult := FoundationCostAdjustment(foundation)
	storiesMult := StoriesCostAdjustment(stories)

	totalCents := int64(float64(baseCostPerSqFt) * gsf * regionalMult * foundationMult * storiesMult)
	return totalCents
}

// CalculateProjectCostByPhase returns per-phase cost breakdown in cents.
// Each entry maps WBS phase code to estimated cost for that phase.
func CalculateProjectCostByPhase(gsf float64, foundation string, stories int, region string) map[string]int64 {
	regionalMult := 1.0
	if m, ok := RegionalMultipliers()[region]; ok {
		regionalMult = m
	}
	foundationMult := FoundationCostAdjustment(foundation)
	storiesMult := StoriesCostAdjustment(stories)

	result := make(map[string]int64)
	for _, idx := range NationalCostIndices() {
		phaseCost := int64(float64(idx.CostPerSqFtCents) * gsf * regionalMult * foundationMult * storiesMult)
		result[idx.WBSPhaseCode] = phaseCost
	}
	return result
}

// CalculateProjectCostWithSeason estimates project cost with seasonal material indices.
// Uses the start month to determine which seasonal factors apply to each phase.
func CalculateProjectCostWithSeason(gsf float64, foundation string, stories int, region string, startMonth int) int64 {
	if gsf <= 0 {
		return 0
	}
	if startMonth < 1 || startMonth > 12 {
		startMonth = 1
	}
	regionalMult := 1.0
	if m, ok := RegionalMultipliers()[region]; ok {
		regionalMult = m
	}
	foundationMult := FoundationCostAdjustment(foundation)
	storiesMult := StoriesCostAdjustment(stories)

	// Map WBS phases to approximate month offsets from project start
	phaseMonthOffset := map[string]int{
		"7.x":  0, // Site Prep starts immediately
		"8.x":  1, // Foundation ~1 month in
		"9.x":  2, // Framing ~2 months in
		"10.x": 3, // Rough-ins ~3 months in
		"11.x": 4, // Insulation/Drywall ~4 months in
		"12.x": 5, // Interior Finishes ~5 months in
		"13.x": 5, // Exterior ~5 months in (parallel)
		"14.x": 7, // Commissioning ~7 months in
		"15.x": 8, // Warranty ~8 months in
	}

	var totalCents int64
	for _, idx := range NationalCostIndices() {
		offset := phaseMonthOffset[idx.WBSPhaseCode]
		phaseMonth := ((startMonth - 1 + offset) % 12) + 1
		seasonalFactor := MonthlySeasonalCostFactor(phaseMonth)
		phaseCost := int64(float64(idx.CostPerSqFtCents) * gsf * regionalMult * foundationMult * storiesMult * seasonalFactor)
		totalCents += phaseCost
	}
	return totalCents
}

// EstimateCostDelta estimates the cost impact of a change in square footage
// on specific WBS categories. Returns the delta in cents.
func EstimateCostDelta(sqftDelta float64, wbsCategories []string, region string) int64 {
	regionalMult := 1.0
	if m, ok := RegionalMultipliers()[region]; ok {
		regionalMult = m
	}

	indices := NationalCostIndices()
	categorySet := make(map[string]bool, len(wbsCategories))
	for _, c := range wbsCategories {
		categorySet[c] = true
	}

	var deltaCents int64
	for _, idx := range indices {
		if len(wbsCategories) > 0 && !categorySet[idx.WBSPhaseCode] {
			continue
		}
		deltaCents += int64(float64(idx.CostPerSqFtCents) * sqftDelta * regionalMult)
	}
	return deltaCents
}

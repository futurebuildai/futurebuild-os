package data

// ResidentialCostIndex holds national-average base costs per gross square foot
// for a specific WBS construction phase. Derived from publicly available
// residential construction cost data (national averages, 2024-2025).
//
// MONETARY PRECISION: All costs in int64 cents per square foot.
type ResidentialCostIndex struct {
	WBSPhaseCode     string  // e.g., "7.x"
	PhaseName        string  // e.g., "Site Prep"
	CostPerSqFtCents int64   // National average base cost in cents/sqft
	LaborSharePct    float64 // 0.0-1.0 share of cost that is labor
}

// NationalCostIndices returns per-phase cost indices for residential construction.
// These represent national averages for a standard 2,250 sqft single-story home.
// Costs scale with GSF via the SAF formula (see physics/dhsm.go).
func NationalCostIndices() []ResidentialCostIndex {
	return []ResidentialCostIndex{
		{WBSPhaseCode: "7.x", PhaseName: "Site Prep", CostPerSqFtCents: 650, LaborSharePct: 0.70},
		{WBSPhaseCode: "8.x", PhaseName: "Foundation", CostPerSqFtCents: 2000, LaborSharePct: 0.55},
		{WBSPhaseCode: "9.x", PhaseName: "Framing", CostPerSqFtCents: 2500, LaborSharePct: 0.50},
		{WBSPhaseCode: "10.x", PhaseName: "Rough-Ins", CostPerSqFtCents: 2200, LaborSharePct: 0.65},
		{WBSPhaseCode: "11.x", PhaseName: "Insulation/Drywall", CostPerSqFtCents: 1200, LaborSharePct: 0.60},
		{WBSPhaseCode: "12.x", PhaseName: "Interior Finishes", CostPerSqFtCents: 3800, LaborSharePct: 0.55},
		{WBSPhaseCode: "13.x", PhaseName: "Exterior", CostPerSqFtCents: 1500, LaborSharePct: 0.50},
		{WBSPhaseCode: "14.x", PhaseName: "Commissioning & Closeout", CostPerSqFtCents: 400, LaborSharePct: 0.80},
		{WBSPhaseCode: "15.x", PhaseName: "Warranty", CostPerSqFtCents: 150, LaborSharePct: 0.90},
	}
}

// TotalNationalCostPerSqFtCents returns the sum of all phase costs.
// Useful for quick project-level estimation.
func TotalNationalCostPerSqFtCents() int64 {
	var total int64
	for _, idx := range NationalCostIndices() {
		total += idx.CostPerSqFtCents
	}
	return total
}

// RegionalMultipliers maps geographic regions to cost adjustment factors.
// 1.0 = national average. Values derived from publicly available BLS
// and construction cost survey data.
func RegionalMultipliers() map[string]float64 {
	return map[string]float64{
		// West Coast
		"CA-Bay Area":     1.45,
		"CA-Los Angeles":  1.30,
		"CA-San Diego":    1.25,
		"OR-Portland":     1.15,
		"WA-Seattle":      1.25,

		// Mountain West
		"CO-Denver":       1.05,
		"CO-Mountain":     1.20,
		"UT-Salt Lake":    0.95,
		"AZ-Phoenix":      0.90,
		"NV-Las Vegas":    1.00,
		"MT-Bozeman":      1.10,
		"ID-Boise":        0.95,

		// Southwest
		"TX-Austin":       0.92,
		"TX-Dallas":       0.90,
		"TX-Houston":      0.88,
		"NM-Albuquerque":  0.85,

		// Midwest
		"IL-Chicago":      1.10,
		"MN-Minneapolis":  1.05,
		"OH-Columbus":     0.88,
		"MI-Detroit":      0.92,
		"MO-Kansas City":  0.85,
		"WI-Milwaukee":    0.95,
		"IN-Indianapolis": 0.85,

		// Southeast
		"FL-Miami":        1.05,
		"FL-Tampa":        0.95,
		"GA-Atlanta":      0.92,
		"NC-Charlotte":    0.90,
		"NC-Raleigh":      0.88,
		"TN-Nashville":    0.90,
		"SC-Charleston":   0.92,
		"VA-Richmond":     0.95,
		"AL-Birmingham":   0.82,

		// Northeast
		"NY-New York City": 1.55,
		"NY-Upstate":       1.05,
		"MA-Boston":        1.35,
		"CT-Hartford":      1.20,
		"NJ-Northern":      1.30,
		"PA-Philadelphia":  1.15,
		"PA-Pittsburgh":    1.00,
		"MD-Baltimore":     1.05,
		"DC-Washington":    1.15,
		"ME-Portland":      1.05,
		"NH-Manchester":    1.05,
		"VT-Burlington":    1.10,

		// Pacific Northwest / Other
		"HI-Honolulu":     1.50,
		"AK-Anchorage":    1.40,
	}
}

// FoundationCostAdjustment returns a multiplier for foundation cost based on type.
// Basement construction is significantly more expensive than slab-on-grade.
func FoundationCostAdjustment(foundationType string) float64 {
	switch foundationType {
	case "basement":
		return 1.40
	case "crawlspace":
		return 1.15
	case "slab":
		return 1.00
	default:
		return 1.00
	}
}

// StoriesCostAdjustment returns a multiplier for multi-story construction.
// Two-story homes have slightly higher per-sqft costs due to structural
// requirements and vertical logistics, but lower foundation/roof cost per sqft.
func StoriesCostAdjustment(stories int) float64 {
	switch {
	case stories >= 3:
		return 1.15
	case stories == 2:
		return 1.05
	default:
		return 1.00
	}
}

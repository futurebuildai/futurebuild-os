package data

// SeasonalMaterialIndex holds the seasonal cost multiplier for a material category.
// Month 1-12. Index 1.0 = baseline. Based on PPI (Producer Price Index) patterns.
type SeasonalMaterialIndex struct {
	Material string
	Month    int
	Index    float64
}

// GetSeasonalIndices returns monthly cost indices for key construction materials.
// Based on publicly available BLS PPI seasonal patterns.
func GetSeasonalIndices() []SeasonalMaterialIndex {
	return []SeasonalMaterialIndex{
		// Lumber — peaks May-August (construction season demand)
		{"lumber", 1, 0.92}, {"lumber", 2, 0.94}, {"lumber", 3, 0.98},
		{"lumber", 4, 1.05}, {"lumber", 5, 1.12}, {"lumber", 6, 1.15},
		{"lumber", 7, 1.18}, {"lumber", 8, 1.14}, {"lumber", 9, 1.08},
		{"lumber", 10, 1.02}, {"lumber", 11, 0.95}, {"lumber", 12, 0.90},

		// Concrete — peaks summer, drops winter
		{"concrete", 1, 0.94}, {"concrete", 2, 0.96}, {"concrete", 3, 1.00},
		{"concrete", 4, 1.04}, {"concrete", 5, 1.08}, {"concrete", 6, 1.10},
		{"concrete", 7, 1.10}, {"concrete", 8, 1.08}, {"concrete", 9, 1.04},
		{"concrete", 10, 1.00}, {"concrete", 11, 0.96}, {"concrete", 12, 0.94},

		// Steel — moderate seasonal variation
		{"steel", 1, 0.97}, {"steel", 2, 0.98}, {"steel", 3, 1.00},
		{"steel", 4, 1.02}, {"steel", 5, 1.04}, {"steel", 6, 1.05},
		{"steel", 7, 1.04}, {"steel", 8, 1.03}, {"steel", 9, 1.01},
		{"steel", 10, 1.00}, {"steel", 11, 0.98}, {"steel", 12, 0.97},

		// HVAC equipment — fairly steady, slight winter increase
		{"hvac", 1, 1.03}, {"hvac", 2, 1.02}, {"hvac", 3, 1.00},
		{"hvac", 4, 0.99}, {"hvac", 5, 0.98}, {"hvac", 6, 0.98},
		{"hvac", 7, 0.99}, {"hvac", 8, 1.00}, {"hvac", 9, 1.01},
		{"hvac", 10, 1.02}, {"hvac", 11, 1.03}, {"hvac", 12, 1.04},

		// Drywall/Gypsum — slight summer peak
		{"drywall", 1, 0.97}, {"drywall", 2, 0.98}, {"drywall", 3, 1.00},
		{"drywall", 4, 1.02}, {"drywall", 5, 1.03}, {"drywall", 6, 1.04},
		{"drywall", 7, 1.04}, {"drywall", 8, 1.03}, {"drywall", 9, 1.01},
		{"drywall", 10, 1.00}, {"drywall", 11, 0.98}, {"drywall", 12, 0.97},
	}
}

// LaborAvailabilityIndex represents trade-specific labor market tightness by month.
// Index 1.0 = normal availability. Lower = tighter (delays more likely).
type LaborAvailabilityIndex struct {
	Trade string
	Month int
	Index float64
}

// GetLaborAvailability returns monthly labor availability indices by trade.
// Based on construction industry employment seasonal patterns.
func GetLaborAvailability() []LaborAvailabilityIndex {
	return []LaborAvailabilityIndex{
		// General labor — tightest in summer peak season
		{"general", 1, 1.10}, {"general", 2, 1.08}, {"general", 3, 1.00},
		{"general", 4, 0.95}, {"general", 5, 0.88}, {"general", 6, 0.85},
		{"general", 7, 0.85}, {"general", 8, 0.87}, {"general", 9, 0.92},
		{"general", 10, 0.98}, {"general", 11, 1.05}, {"general", 12, 1.10},

		// Electricians — demand spikes in summer
		{"electrician", 1, 1.05}, {"electrician", 2, 1.03}, {"electrician", 3, 0.98},
		{"electrician", 4, 0.94}, {"electrician", 5, 0.90}, {"electrician", 6, 0.87},
		{"electrician", 7, 0.85}, {"electrician", 8, 0.88}, {"electrician", 9, 0.93},
		{"electrician", 10, 0.98}, {"electrician", 11, 1.03}, {"electrician", 12, 1.06},

		// Plumbers — steadier demand, slight summer dip
		{"plumber", 1, 1.02}, {"plumber", 2, 1.01}, {"plumber", 3, 0.99},
		{"plumber", 4, 0.97}, {"plumber", 5, 0.94}, {"plumber", 6, 0.92},
		{"plumber", 7, 0.92}, {"plumber", 8, 0.94}, {"plumber", 9, 0.97},
		{"plumber", 10, 1.00}, {"plumber", 11, 1.02}, {"plumber", 12, 1.03},

		// Framers — very seasonal, scarce in peak summer
		{"framer", 1, 1.12}, {"framer", 2, 1.08}, {"framer", 3, 0.98},
		{"framer", 4, 0.92}, {"framer", 5, 0.85}, {"framer", 6, 0.82},
		{"framer", 7, 0.82}, {"framer", 8, 0.85}, {"framer", 9, 0.92},
		{"framer", 10, 1.00}, {"framer", 11, 1.08}, {"framer", 12, 1.12},

		// HVAC — peaks in Q3/Q4 (winter prep)
		{"hvac", 1, 1.00}, {"hvac", 2, 1.02}, {"hvac", 3, 1.03},
		{"hvac", 4, 0.98}, {"hvac", 5, 0.95}, {"hvac", 6, 0.92},
		{"hvac", 7, 0.90}, {"hvac", 8, 0.88}, {"hvac", 9, 0.90},
		{"hvac", 10, 0.93}, {"hvac", 11, 0.97}, {"hvac", 12, 1.00},
	}
}

// GetSeasonalMaterialIndex returns the cost index for a material in a given month.
// Returns 1.0 if material/month combination is not found.
func GetSeasonalMaterialIndex(material string, month int) float64 {
	for _, idx := range GetSeasonalIndices() {
		if idx.Material == material && idx.Month == month {
			return idx.Index
		}
	}
	return 1.0
}

// GetLaborAvailabilityIndex returns the availability index for a trade in a given month.
// Returns 1.0 if trade/month combination is not found.
func GetLaborAvailabilityIndex(trade string, month int) float64 {
	for _, idx := range GetLaborAvailability() {
		if idx.Trade == trade && idx.Month == month {
			return idx.Index
		}
	}
	return 1.0
}

// MonthlySeasonalCostFactor returns the blended seasonal cost multiplier for a given month.
// Averages across all material categories.
func MonthlySeasonalCostFactor(month int) float64 {
	materials := []string{"lumber", "concrete", "steel", "hvac", "drywall"}
	var sum float64
	for _, m := range materials {
		sum += GetSeasonalMaterialIndex(m, month)
	}
	return sum / float64(len(materials))
}

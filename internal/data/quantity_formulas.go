package data

import "math"

// QuantityEstimate represents a heuristic-derived material quantity for a project.
// Used when blueprint-level detail is unavailable — estimates from project attributes.
type QuantityEstimate struct {
	MaterialName  string  // Human-readable name
	Category      string  // structural, framing, roofing, etc.
	WBSPhaseCode  string  // Maps to WBS phase for budget allocation
	Quantity      float64 // Estimated count/measure
	Unit          string  // sqft, lf, ea, cy, bf, etc.
	UnitCostCents int64   // Cost per unit in cents
	Confidence    float64 // 0.0-1.0 confidence in this estimate
	Formula       string  // Human-readable formula description
}

// EstimateQuantities produces material quantity estimates using project attributes.
// These are heuristic-based approximations using industry rules of thumb.
// Costs are national averages — apply RegionalMultipliers() for local pricing.
func EstimateQuantities(gsf float64, stories int, foundationType string, bedrooms int, bathrooms int) []QuantityEstimate {
	if gsf <= 0 {
		return nil
	}
	if stories <= 0 {
		stories = 1
	}

	footprint := gsf / float64(stories)

	estimates := []QuantityEstimate{
		// === Phase 8.x: Foundation ===
		concreteEstimate(footprint, foundationType),
		rebarEstimate(footprint, foundationType),

		// === Phase 9.x: Framing ===
		lumberFramingEstimate(gsf, stories),
		roofTrussEstimate(footprint),
		sheathingEstimate(gsf, stories),

		// === Phase 9.x: Roofing & Exterior ===
		roofingEstimate(footprint),

		// === Phase 10.x: Rough-Ins ===
		electricalWiringEstimate(gsf),
		electricalOutletEstimate(gsf),
		plumbingPipeEstimate(gsf, bathrooms),
		plumbingFixtureEstimate(bedrooms, bathrooms),
		hvacDuctworkEstimate(gsf),

		// === Phase 11.x: Insulation/Drywall ===
		insulationEstimate(gsf, stories),
		drywallEstimate(gsf, stories),

		// === Phase 12.x: Interior Finishes ===
		interiorPaintEstimate(gsf, stories),
		flooringEstimate(gsf),
		cabinetryEstimate(gsf),
		windowEstimate(gsf, stories),
		interiorDoorEstimate(bedrooms, bathrooms),

		// === Phase 13.x: Exterior ===
		sidingEstimate(gsf, stories),
	}

	return estimates
}

func concreteEstimate(footprint float64, foundationType string) QuantityEstimate {
	// Slab: footprint * 4" thick = footprint * 0.333 / 27 cy
	// Basement: footprint * 0.5 for walls + footprint * 0.333 / 27 for floor
	var qty float64
	switch foundationType {
	case "basement":
		qty = math.Ceil((footprint*0.5 + footprint*0.333) / 27.0)
	case "crawlspace":
		qty = math.Ceil(footprint * 0.012) // Footings only
	default: // slab
		qty = math.Ceil(footprint * 0.333 / 27.0)
	}
	return QuantityEstimate{
		MaterialName:  "Ready-Mix Concrete",
		Category:      "structural",
		WBSPhaseCode:  "8.x",
		Quantity:      qty,
		Unit:          "cy",
		UnitCostCents: 15000, // $150/cy
		Confidence:    0.65,
		Formula:       "Footprint-based volume estimate",
	}
}

func rebarEstimate(footprint float64, foundationType string) QuantityEstimate {
	// ~0.5 lb rebar per sqft of slab, more for basement
	multiplier := 0.5
	if foundationType == "basement" {
		multiplier = 1.2
	}
	qty := math.Ceil(footprint * multiplier / 100.0) // Convert to hundredweight
	return QuantityEstimate{
		MaterialName:  "Rebar (#4 & #5)",
		Category:      "structural",
		WBSPhaseCode:  "8.x",
		Quantity:      qty * 100, // lbs
		Unit:          "lb",
		UnitCostCents: 75, // $0.75/lb
		Confidence:    0.55,
		Formula:       "0.5-1.2 lb/sqft of footprint",
	}
}

func lumberFramingEstimate(gsf float64, stories int) QuantityEstimate {
	// ~1.1 board feet per sqft per story for wall/floor framing
	qty := math.Ceil(gsf * 1.1 * float64(stories))
	return QuantityEstimate{
		MaterialName:  "Dimensional Lumber (2x4, 2x6, 2x10)",
		Category:      "framing",
		WBSPhaseCode:  "9.x",
		Quantity:      qty,
		Unit:          "bf",
		UnitCostCents: 550, // $5.50/bf
		Confidence:    0.60,
		Formula:       "GSF × 1.1 × stories",
	}
}

func roofTrussEstimate(footprint float64) QuantityEstimate {
	// Trusses at 24" OC across the shorter dimension (approximated)
	qty := math.Ceil(footprint / 48.0) // Rough estimate: 1 truss per 48 sqft
	return QuantityEstimate{
		MaterialName:  "Engineered Roof Trusses",
		Category:      "framing",
		WBSPhaseCode:  "9.x",
		Quantity:      qty,
		Unit:          "ea",
		UnitCostCents: 25000, // $250/truss average
		Confidence:    0.50,
		Formula:       "Footprint / 48 (24\" OC approximation)",
	}
}

func sheathingEstimate(gsf float64, stories int) QuantityEstimate {
	// OSB/plywood for walls + roof. ~1.5x GSF per story for walls, + roof area
	qty := math.Ceil(gsf*0.4*float64(stories) + gsf/float64(stories)*1.15) // wall + roof
	// Convert to 4x8 sheets (32 sqft)
	sheets := math.Ceil(qty / 32.0)
	return QuantityEstimate{
		MaterialName:  "OSB/Plywood Sheathing (4×8 sheets)",
		Category:      "framing",
		WBSPhaseCode:  "9.x",
		Quantity:      sheets,
		Unit:          "ea",
		UnitCostCents: 3500, // $35/sheet
		Confidence:    0.50,
		Formula:       "Wall area + roof area / 32 sqft per sheet",
	}
}

func roofingEstimate(footprint float64) QuantityEstimate {
	// Roof area ≈ footprint × 1.15 (accounting for pitch and overhang)
	// 3 bundles per 100 sqft (1 square)
	roofSqft := footprint * 1.15
	squares := math.Ceil(roofSqft / 100.0)
	return QuantityEstimate{
		MaterialName:  "Architectural Shingles",
		Category:      "roofing",
		WBSPhaseCode:  "13.x",
		Quantity:      squares * 3, // bundles
		Unit:          "bundle",
		UnitCostCents: 4500, // $45/bundle
		Confidence:    0.55,
		Formula:       "Footprint × 1.15 / 100 sqft per square × 3 bundles",
	}
}

func electricalWiringEstimate(gsf float64) QuantityEstimate {
	// ~2 linear feet of Romex per sqft of conditioned space
	qty := math.Ceil(gsf * 2.0)
	return QuantityEstimate{
		MaterialName:  "Electrical Wire (Romex 14/2 & 12/2)",
		Category:      "electrical",
		WBSPhaseCode:  "10.x",
		Quantity:      qty,
		Unit:          "lf",
		UnitCostCents: 45, // $0.45/lf average
		Confidence:    0.55,
		Formula:       "GSF × 2.0 linear feet",
	}
}

func electricalOutletEstimate(gsf float64) QuantityEstimate {
	// NEC requires outlet every 12 feet along wall, roughly 1 per 80 sqft
	qty := math.Ceil(gsf / 80.0)
	return QuantityEstimate{
		MaterialName:  "Electrical Outlets & Switches",
		Category:      "electrical",
		WBSPhaseCode:  "10.x",
		Quantity:      qty,
		Unit:          "ea",
		UnitCostCents: 4500, // $45/ea installed
		Confidence:    0.60,
		Formula:       "GSF / 80 (NEC spacing rule)",
	}
}

func plumbingPipeEstimate(gsf float64, bathrooms int) QuantityEstimate {
	// ~0.15 lf per sqft base + 25 lf per bathroom
	qty := math.Ceil(gsf*0.15 + float64(bathrooms)*25.0)
	return QuantityEstimate{
		MaterialName:  "PEX/Copper Supply & DWV Pipe",
		Category:      "plumbing",
		WBSPhaseCode:  "10.x",
		Quantity:      qty,
		Unit:          "lf",
		UnitCostCents: 350, // $3.50/lf average (PEX + DWV mix)
		Confidence:    0.50,
		Formula:       "GSF × 0.15 + bathrooms × 25",
	}
}

func plumbingFixtureEstimate(bedrooms int, bathrooms int) QuantityEstimate {
	// Per bathroom: toilet, sink, tub/shower. Kitchen: sink, dishwasher hookup.
	// Laundry: washer hookup
	qty := float64(bathrooms)*3 + 3 // bath fixtures + kitchen + laundry
	return QuantityEstimate{
		MaterialName:  "Plumbing Fixtures (toilets, sinks, tubs/showers)",
		Category:      "plumbing",
		WBSPhaseCode:  "10.x",
		Quantity:      qty,
		Unit:          "ea",
		UnitCostCents: 35000, // $350/fixture average
		Confidence:    0.65,
		Formula:       "Bathrooms × 3 + 3 (kitchen/laundry)",
	}
}

func hvacDuctworkEstimate(gsf float64) QuantityEstimate {
	// ~1 ton per 500 sqft. Ductwork ~30 lf per ton.
	tons := math.Ceil(gsf / 500.0)
	qty := tons * 30
	return QuantityEstimate{
		MaterialName:  "HVAC Ductwork (flex & rigid)",
		Category:      "hvac",
		WBSPhaseCode:  "10.x",
		Quantity:      qty,
		Unit:          "lf",
		UnitCostCents: 1200, // $12/lf installed
		Confidence:    0.50,
		Formula:       "GSF / 500 tons × 30 lf/ton",
	}
}

func insulationEstimate(gsf float64, stories int) QuantityEstimate {
	// Wall + ceiling insulation. ~1.2x GSF for walls per story + ceiling = footprint
	wallArea := gsf * 0.4 * float64(stories) // Approximate wall area
	ceilingArea := gsf / float64(stories)     // Top floor ceiling
	qty := math.Ceil(wallArea + ceilingArea)
	return QuantityEstimate{
		MaterialName:  "Batt & Blown Insulation",
		Category:      "insulation",
		WBSPhaseCode:  "11.x",
		Quantity:      qty,
		Unit:          "sqft",
		UnitCostCents: 120, // $1.20/sqft average
		Confidence:    0.55,
		Formula:       "Wall area + ceiling area",
	}
}

func drywallEstimate(gsf float64, stories int) QuantityEstimate {
	// ~3.5 sqft of drywall per sqft of conditioned space (walls + ceilings)
	sqft := gsf * 3.5
	sheets := math.Ceil(sqft / 32.0) // 4x8 sheets
	return QuantityEstimate{
		MaterialName:  "Drywall (4×8 sheets, 1/2\" & 5/8\")",
		Category:      "drywall",
		WBSPhaseCode:  "11.x",
		Quantity:      sheets,
		Unit:          "ea",
		UnitCostCents: 1400, // $14/sheet
		Confidence:    0.65,
		Formula:       "GSF × 3.5 / 32 sqft per sheet",
	}
}

func interiorPaintEstimate(gsf float64, stories int) QuantityEstimate {
	// ~3.5 sqft of paintable surface per sqft, 350 sqft coverage per gallon
	paintableSqft := gsf * 3.5
	gallons := math.Ceil(paintableSqft / 350.0 * 2.0) // 2 coats
	return QuantityEstimate{
		MaterialName:  "Interior Paint (primer + 2 coats)",
		Category:      "finishes",
		WBSPhaseCode:  "12.x",
		Quantity:      gallons,
		Unit:          "gal",
		UnitCostCents: 4500, // $45/gallon
		Confidence:    0.60,
		Formula:       "GSF × 3.5 / 350 sqft/gal × 2 coats",
	}
}

func flooringEstimate(gsf float64) QuantityEstimate {
	// ~90% of GSF gets flooring (minus closets, utility rooms with bare concrete)
	qty := math.Ceil(gsf * 0.90)
	return QuantityEstimate{
		MaterialName:  "Flooring (hardwood/tile/carpet mix)",
		Category:      "flooring",
		WBSPhaseCode:  "12.x",
		Quantity:      qty,
		Unit:          "sqft",
		UnitCostCents: 800, // $8/sqft blended average
		Confidence:    0.50,
		Formula:       "GSF × 0.90",
	}
}

func cabinetryEstimate(gsf float64) QuantityEstimate {
	// Kitchen + bath cabinetry. Roughly 30-40 lf for average home.
	lf := math.Max(25, gsf/80.0)
	return QuantityEstimate{
		MaterialName:  "Cabinetry (kitchen & bath)",
		Category:      "millwork",
		WBSPhaseCode:  "12.x",
		Quantity:      math.Ceil(lf),
		Unit:          "lf",
		UnitCostCents: 25000, // $250/lf mid-grade
		Confidence:    0.45,
		Formula:       "max(25, GSF / 80) linear feet",
	}
}

func windowEstimate(gsf float64, stories int) QuantityEstimate {
	// ~1 window per 200 sqft per story
	qty := math.Ceil(gsf / 200.0)
	return QuantityEstimate{
		MaterialName:  "Windows (double-hung/casement, vinyl or wood)",
		Category:      "fixtures",
		WBSPhaseCode:  "9.x",
		Quantity:      qty,
		Unit:          "ea",
		UnitCostCents: 50000, // $500/window average
		Confidence:    0.50,
		Formula:       "GSF / 200",
	}
}

func interiorDoorEstimate(bedrooms int, bathrooms int) QuantityEstimate {
	// Bedroom doors + bathroom doors + closet doors + utility
	qty := float64(bedrooms) + float64(bathrooms) + float64(bedrooms)*1.5 + 3 // closets + utility
	return QuantityEstimate{
		MaterialName:  "Interior Doors (prehung)",
		Category:      "millwork",
		WBSPhaseCode:  "12.x",
		Quantity:      math.Ceil(qty),
		Unit:          "ea",
		UnitCostCents: 25000, // $250/door prehung
		Confidence:    0.55,
		Formula:       "Bedrooms + bathrooms + closets + utility",
	}
}

func sidingEstimate(gsf float64, stories int) QuantityEstimate {
	// Exterior wall area ≈ perimeter × wall height
	// Perimeter ≈ 4 × sqrt(footprint)
	footprint := gsf / float64(stories)
	perimeter := 4.0 * math.Sqrt(footprint)
	wallHeight := 9.0 * float64(stories) // 9ft per story
	exteriorSqft := perimeter * wallHeight * 0.85 // 15% window/door openings
	return QuantityEstimate{
		MaterialName:  "Exterior Siding/Cladding",
		Category:      "siding",
		WBSPhaseCode:  "13.x",
		Quantity:      math.Ceil(exteriorSqft),
		Unit:          "sqft",
		UnitCostCents: 800, // $8/sqft mid-grade
		Confidence:    0.50,
		Formula:       "Perimeter × height × 0.85",
	}
}

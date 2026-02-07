package prompts

import (
	"encoding/json"
	"fmt"
)

// BlueprintExtractionPrompt returns the system prompt for document analysis.
// Extracts building specs and long-lead procurement items for schedule generation.
func BlueprintExtractionPrompt() string {
	return `You are a construction project analyst. Extract the following information from this architectural blueprint or floor plan:

REQUIRED FIELDS:
- name: Project name (from title block)
- address: Full street address (including zip code if visible)
- square_footage: Gross Square Footage (total conditioned space)
- foundation_type: "slab", "crawlspace", or "basement"
- stories: Number of stories/levels
- bedrooms: Bedroom count
- bathrooms: Bathroom count

PROCUREMENT INDICATORS (long-lead items that affect schedule):
Look for specific brands, models, and specifications for items that typically have long lead times:
- Windows: Brand (Marvin, Andersen, Pella, Milgard), model, sizes
- Doors: Entry doors, garage doors, specialty doors
- HVAC equipment: Brand, model, tonnage
- Appliances: Built-in, commercial-grade, Sub-Zero, Wolf, Viking, etc.
- Custom millwork: Cabinetry notes, specialty woodwork
- Special finishes: Stone, tile with specific sourcing

RESPOND IN JSON FORMAT ONLY:
{
  "name": "string or null",
  "address": "string or null",
  "square_footage": number or null,
  "foundation_type": "slab" | "crawlspace" | "basement" | null,
  "stories": number or null,
  "bedrooms": number or null,
  "bathrooms": number or null,
  "long_lead_items": [
    {
      "name": "item description",
      "brand": "brand name or null",
      "model": "model number or null",
      "category": "windows" | "doors" | "hvac" | "appliances" | "millwork" | "finishes",
      "notes": "any relevant specs"
    }
  ],
  "confidence": {
    "name": 0.0-1.0,
    "address": 0.0-1.0,
    "square_footage": 0.0-1.0,
    "foundation_type": 0.0-1.0,
    "stories": 0.0-1.0,
    "bedrooms": 0.0-1.0,
    "bathrooms": 0.0-1.0
  }
}

For fields you cannot determine, use null and set confidence to 0.
For fields you're uncertain about, set confidence between 0.5-0.8.
Return an empty array for long_lead_items if none are found.`
}

// MessageParsingPrompt returns the prompt for natural language parsing.
func MessageParsingPrompt(message string, currentState map[string]interface{}) string {
	// Format current state as JSON for clarity
	stateJSON, _ := json.Marshal(currentState)
	if stateJSON == nil {
		stateJSON = []byte("{}")
	}

	return fmt.Sprintf(`You are extracting construction project details from a user's message.

CURRENT PROJECT STATE (already collected):
%s

USER MESSAGE:
"%s"

Extract any NEW information. Look for:
- name: Project name (e.g., "Project: Oak Ridge" → "Oak Ridge", "called Sunset Heights" → "Sunset Heights", "building Mountain View Estate" → "Mountain View Estate")
- address: Full address with city/state/zip
- square_footage: Number only (e.g., "3,200 sq ft" → 3200, "2800 square feet" → 2800)
- foundation_type: Must be exactly "slab", "crawlspace", or "basement"
- stories: Number (e.g., "two-story" → 2, "single story" → 1)
- bedrooms: Number
- bathrooms: Number (e.g., "2.5 bath" → 2.5)
- start_date: ISO date format YYYY-MM-DD (e.g., "January 15, 2026" → "2026-01-15")
- topography: Must be exactly "flat", "sloped", or "hillside"
- soil_conditions: Must be exactly "normal", "rocky", "clay", or "sandy"

IMPORTANT: Return ONLY valid JSON, no other text. Include only fields with new values.

{
  "values": {
    "name": "extracted project name",
    "address": "full address"
  },
  "confidence": {
    "name": 0.95,
    "address": 0.9
  }
}`, string(stateJSON), message)
}

// ClarifyingQuestionPrompt generates context-aware follow-up questions.
func ClarifyingQuestionPrompt(field string, extractedValue any, confidence float64) string {
	templates := map[string]string{
		"square_footage":   "The plans show approximately %.0f square feet. Is that correct?",
		"foundation_type": "I see what looks like a %s foundation. Can you confirm?",
		"bedrooms":        "I found %d potential bedrooms. One looks like it might be an office—should I count it?",
		"bathrooms":       "I see %d bathrooms. Is one of them a master en-suite?",
	}

	if template, ok := templates[field]; ok {
		return fmt.Sprintf(template, extractedValue)
	}
	return ""
}

// LongLeadItemsPrompt returns the prompt for extracting procurement items with lead times.
func LongLeadItemsPrompt() string {
	return `Identify any materials or equipment in these plans that typically have
long lead times (4+ weeks). Look for:
- Specific window/door brands and models
- HVAC equipment specifications
- Custom cabinetry or millwork
- Specialty fixtures (commercial-grade appliances, etc.)
- Stone/tile with specific sourcing requirements

Known lead time estimates by brand:
WINDOWS:
- Marvin Ultimate/Signature: 12-16 weeks
- Andersen E-Series: 8-12 weeks
- Pella Reserve: 10-14 weeks
- Milgard: 4-6 weeks

APPLIANCES:
- Sub-Zero: 8-12 weeks
- Wolf: 8-12 weeks
- Viking: 6-10 weeks
- La Cornue: 16-24 weeks

HVAC:
- Standard equipment: 2-4 weeks
- High-efficiency/Geothermal: 6-10 weeks

CUSTOM MILLWORK:
- Standard cabinets: 4-6 weeks
- Custom cabinetry: 8-12 weeks

Return JSON with items and estimated lead time weeks:
{
  "long_lead_items": [
    {
      "name": "item description",
      "brand": "brand name",
      "model": "model if known",
      "category": "windows" | "doors" | "hvac" | "appliances" | "millwork" | "finishes",
      "estimated_lead_weeks": number,
      "wbs_code": "affected WBS code if known (e.g., 8.1 for windows)"
    }
  ]
}`
}

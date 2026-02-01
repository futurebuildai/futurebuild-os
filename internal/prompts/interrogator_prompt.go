package prompts

import "fmt"

// BlueprintExtractionPrompt returns the system prompt for document analysis.
func BlueprintExtractionPrompt() string {
	return `You are a construction project analyst. Extract the following information from this architectural blueprint or floor plan:

REQUIRED FIELDS:
- name: Project name (from title block)
- address: Full street address
- square_footage: Gross Square Footage (total conditioned space)
- foundation_type: "slab", "crawlspace", or "basement"
- stories: Number of stories/levels
- bedrooms: Bedroom count
- bathrooms: Bathroom count

RESPOND IN JSON FORMAT ONLY:
{
  "name": "string or null",
  "address": "string or null",
  "square_footage": number or null,
  "foundation_type": "slab" | "crawlspace" | "basement" | null,
  "stories": number or null,
  "bedrooms": number or null,
  "bathrooms": number or null,
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
For fields you're uncertain about, set confidence between 0.5-0.8.`
}

// MessageParsingPrompt returns the prompt for natural language parsing.
func MessageParsingPrompt(message string, currentState map[string]interface{}) string {
	return fmt.Sprintf(`You are extracting project details from a user's message.

CURRENT PROJECT STATE:
%v

USER MESSAGE:
"%s"

Extract any NEW information about:
- name (project name)
- address (location)
- square_footage (square footage, look for "sq ft", "square feet", numbers like "3200")
- foundation_type ("slab", "crawlspace", "basement")
- stories (1, 2, etc or "single story", "two story")
- bedrooms (number)
- bathrooms (number)
- topography ("flat", "sloped", "hillside")
- soil_conditions ("normal", "rocky", "clay", "sandy")

RESPOND IN JSON FORMAT ONLY:
{
  "values": {
    "field_name": "extracted_value"
  },
  "confidence": {
    "field_name": 0.0-1.0
  }
}

Only include fields that have new values from this message. Do not repeat existing values.`, currentState, message)
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

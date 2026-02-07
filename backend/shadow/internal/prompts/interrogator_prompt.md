# interrogator_prompt

## Intent
*   **High Level:** AI prompts for project onboarding extraction
*   **Business Value:** Enables accurate extraction of construction project data from user input

## Responsibility
*   Define prompts for blueprint/document analysis
*   Define prompts for natural language message parsing
*   Define prompts for long-lead item detection

## Key Prompts

### BlueprintExtractionPrompt
Extracts from architectural blueprints:
- Building specs: name, address, sqft, foundation, stories, bed/bath
- Long-lead items: windows, doors, HVAC, appliances with brand detection
- Returns JSON with values and confidence scores

### MessageParsingPrompt
Parses natural language user messages:
- Current state passed as JSON (not Go map syntax)
- Explicit examples for name extraction ("Project: X" → "X", "called Y" → "Y")
- Instructs AI to return ONLY valid JSON, no markdown wrappers
- Extracts: name, address, square_footage, foundation_type, stories, bedrooms, bathrooms, start_date, topography, soil_conditions

### LongLeadItemsPrompt
Identifies materials with long lead times:
- Known brand lead times (Marvin: 12-16 weeks, Sub-Zero: 8-12 weeks, etc.)
- Categories: windows, doors, hvac, appliances, millwork, finishes
- Returns estimated lead weeks for schedule constraint planning

## Dependencies
*   **Upstream:** InterrogatorService
*   **Downstream:** Gemini AI models (Flash for text, Pro for vision)

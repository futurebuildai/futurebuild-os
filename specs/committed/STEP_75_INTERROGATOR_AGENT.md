# Step 75: The Interrogator Agent (Backend)

**Phase:** 11 (The Conversational Hook)  
**Status:** READY FOR IMPLEMENTATION  
**Est. Duration:** 2 Days  
**Owner:** Backend Developer

---

## 🎯 Objective

Implement the backend logic for "The Interrogator" — Agent 2's onboarding persona that asks clarifying questions to extract physics engine inputs from user conversation and uploaded blueprints.

---

## 📐 Architectural Guardrails

> [!IMPORTANT]
> These constraints are non-negotiable and align with the system architecture.

### Layer Separation (CRITICAL)
Reference: [PRODUCT_VISION.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/PRODUCT_VISION.md) Section 2.2

| Layer | Responsibility | This Step |
|-------|----------------|-----------|
| **Layer 1 (Context Engine)** | Probabilistic extraction | ✅ Gemini extracts data from documents |
| **Layer 3 (Physics Engine)** | Deterministic calculation | ❌ DO NOT call physics here |
| **Layer 4 (Action Engine)** | Agent orchestration | ✅ Interrogator generates questions |

```
⚠️ THE INTERROGATOR DOES NOT CALCULATE SCHEDULES.
It ONLY extracts data and populates the project_context table.
Schedule generation happens AFTER project creation via a separate CPM call.
```

### Physics Engine Data Targets
Reference: [PHASE_11_PRD.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/planning/PHASE_11_PRD.md) Section 5.5

The Interrogator MUST prioritize extraction of:
1. **P0 (Critical):** `gsf`, `foundation_type`, `address/zip_code`
2. **P1 (Important):** `topography`, `stories`, `soil_conditions`
3. **P2 (Helpful):** `supply_chain_volatility`, `inspection_latency`

### Go Backend Patterns
Reference: [BACKEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/BACKEND_SCOPE.md)

- Use Chi Router for endpoints
- Implement as a service in `internal/services/`
- Use repository pattern for data access
- All Vertex AI calls go through `internal/vertex/client.go`

---

## 📋 Implementation Checklist

### Files to Create/Modify

```
velocity-backend/internal/
├── handlers/
│   └── onboarding_handler.go       # HTTP handler for /api/v1/agent/onboard
├── services/
│   └── interrogator_service.go     # Core agent logic
├── models/
│   └── onboarding.go               # Request/response types
└── prompts/
    └── interrogator_prompt.go      # System prompts for Gemini
```

---

### Task 1: Define Data Models (30 min)

```go
// internal/models/onboarding.go

package models

import (
    "time"
)

// OnboardRequest is the incoming payload from the frontend wizard.
type OnboardRequest struct {
    SessionID    string                 `json:"session_id"`
    Message      string                 `json:"message,omitempty"`
    DocumentURL  string                 `json:"document_url,omitempty"`
    CurrentState map[string]interface{} `json:"current_state"`
}

// OnboardResponse is returned to the frontend with extracted values.
type OnboardResponse struct {
    SessionID          string             `json:"session_id"`
    Reply              string             `json:"reply"`
    ExtractedValues    map[string]any     `json:"extracted_values"`
    ConfidenceScores   map[string]float64 `json:"confidence_scores"`
    ClarifyingQuestion string             `json:"clarifying_question,omitempty"`
    ReadyToCreate      bool               `json:"ready_to_create"`
    NextPriorityField  string             `json:"next_priority_field,omitempty"`
}

// OnboardingSession persists wizard state (optional for MVP).
type OnboardingSession struct {
    ID               string                 `json:"id"`
    TenantID         string                 `json:"tenant_id"`
    UserID           string                 `json:"user_id"`
    FormState        map[string]interface{} `json:"form_state"`
    ExtractionHistory []ExtractionResult    `json:"extraction_history"`
    Status           string                 `json:"status"` // in_progress, completed, abandoned
    CreatedAt        time.Time              `json:"created_at"`
}

// ExtractionResult logs what was extracted from a document.
type ExtractionResult struct {
    DocumentURL  string             `json:"document_url"`
    ExtractedAt  time.Time          `json:"extracted_at"`
    Values       map[string]any     `json:"values"`
    Confidence   map[string]float64 `json:"confidence"`
}

// PhysicsFieldPriority defines extraction priority.
// P0 = critical, P1 = important, P2 = helpful
type PhysicsFieldPriority struct {
    Field    string
    Priority int // 0, 1, or 2
    Question string
}

// GetPriorityFields returns the ordered list of physics fields.
func GetPriorityFields() []PhysicsFieldPriority {
    return []PhysicsFieldPriority{
        {"name", 0, "What would you like to call this project?"},
        {"address", 0, "Where is the project located?"},
        {"gsf", 0, "What's the approximate square footage?"},
        {"foundation_type", 0, "What type of foundation? Slab, crawlspace, or basement?"},
        {"stories", 1, "Is this a single-story or multi-story home?"},
        {"topography", 1, "Is the lot flat, moderately sloped, or steeply sloped?"},
        {"soil_conditions", 1, "Any special soil conditions like rock or clay?"},
        {"bedrooms", 2, "How many bedrooms?"},
        {"bathrooms", 2, "How many bathrooms?"},
        {"supply_chain_volatility", 2, "Any supply chain concerns for this project?"},
    }
}
```

---

### Task 2: Implement Handler (30 min)

```go
// internal/handlers/onboarding_handler.go

package handlers

import (
    "encoding/json"
    "net/http"
    
    "velocity-backend/internal/models"
    "velocity-backend/internal/services"
)

type OnboardingHandler struct {
    interrogator *services.InterrogatorService
}

func NewOnboardingHandler(svc *services.InterrogatorService) *OnboardingHandler {
    return &OnboardingHandler{interrogator: svc}
}

// HandleOnboard processes the /api/v1/agent/onboard endpoint.
// @route POST /api/v1/agent/onboard
func (h *OnboardingHandler) HandleOnboard(w http.ResponseWriter, r *http.Request) {
    var req models.OnboardRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Get user context from JWT (middleware should have set this)
    userID := r.Context().Value("user_id").(string)
    tenantID := r.Context().Value("tenant_id").(string)

    resp, err := h.interrogator.ProcessMessage(r.Context(), userID, tenantID, &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

---

### Task 3: Implement Interrogator Service (3-4 hours)

```go
// internal/services/interrogator_service.go

package services

import (
    "context"
    "encoding/json"
    "fmt"
    
    "velocity-backend/internal/models"
    "velocity-backend/internal/prompts"
    "velocity-backend/internal/vertex"
)

type InterrogatorService struct {
    vertexClient vertex.Client
    docService   *DocumentService // For document analysis
}

func NewInterrogatorService(vc vertex.Client, ds *DocumentService) *InterrogatorService {
    return &InterrogatorService{
        vertexClient: vc,
        docService:   ds,
    }
}

// ProcessMessage handles a single turn of the onboarding conversation.
func (s *InterrogatorService) ProcessMessage(
    ctx context.Context,
    userID, tenantID string,
    req *models.OnboardRequest,
) (*models.OnboardResponse, error) {
    
    resp := &models.OnboardResponse{
        SessionID:        req.SessionID,
        ExtractedValues:  make(map[string]any),
        ConfidenceScores: make(map[string]float64),
        ReadyToCreate:    false,
    }

    // BRANCH 1: Document uploaded → Extract via Vision API
    if req.DocumentURL != "" {
        extraction, err := s.extractFromDocument(ctx, req.DocumentURL)
        if err != nil {
            resp.Reply = "I couldn't read that file. Could you try a clearer scan or describe the project?"
            return resp, nil
        }
        
        // Merge extracted values into response
        for k, v := range extraction.Values {
            resp.ExtractedValues[k] = v
            resp.ConfidenceScores[k] = extraction.Confidence[k]
        }
        
        // Generate a summary message
        resp.Reply = s.generateExtractionSummary(extraction)
    }

    // BRANCH 2: User message → Parse natural language
    if req.Message != "" {
        extraction, err := s.parseUserMessage(ctx, req.Message, req.CurrentState)
        if err == nil {
            for k, v := range extraction.Values {
                resp.ExtractedValues[k] = v
                resp.ConfidenceScores[k] = extraction.Confidence[k]
            }
        }
    }

    // Merge with existing state
    mergedState := s.mergeStates(req.CurrentState, resp.ExtractedValues)

    // Check if ready to create (name + address are required)
    resp.ReadyToCreate = s.checkReadyToCreate(mergedState)

    // Generate next question if not ready
    if !resp.ReadyToCreate {
        nextField, question := s.getNextQuestion(mergedState)
        resp.NextPriorityField = nextField
        resp.ClarifyingQuestion = question
        
        // If no explicit reply yet, use the clarifying question
        if resp.Reply == "" {
            resp.Reply = question
        }
    } else if resp.Reply == "" {
        resp.Reply = "Great! Your project is ready to create. Review the details and click 'Create Project' when ready."
    }

    return resp, nil
}

// extractFromDocument uses Vision API to extract structured data from blueprints.
func (s *InterrogatorService) extractFromDocument(
    ctx context.Context, 
    documentURL string,
) (*models.ExtractionResult, error) {
    prompt := prompts.BlueprintExtractionPrompt()
    
    result, err := s.vertexClient.GenerateContent(ctx, vertex.ModelFlash, 
        prompt,
        vertex.ImagePart(documentURL),
    )
    if err != nil {
        return nil, err
    }

    // Parse JSON response from Gemini
    var extraction struct {
        Name           string  `json:"name"`
        Address        string  `json:"address"`
        GSF            float64 `json:"gsf"`
        FoundationType string  `json:"foundation_type"`
        Stories        int     `json:"stories"`
        Bedrooms       int     `json:"bedrooms"`
        Bathrooms      int     `json:"bathrooms"`
        Confidence     map[string]float64 `json:"confidence"`
    }
    
    if err := json.Unmarshal([]byte(result), &extraction); err != nil {
        return nil, err
    }

    values := make(map[string]any)
    if extraction.Name != "" { values["name"] = extraction.Name }
    if extraction.Address != "" { values["address"] = extraction.Address }
    if extraction.GSF > 0 { values["gsf"] = extraction.GSF }
    if extraction.FoundationType != "" { values["foundation_type"] = extraction.FoundationType }
    if extraction.Stories > 0 { values["stories"] = extraction.Stories }
    if extraction.Bedrooms > 0 { values["bedrooms"] = extraction.Bedrooms }
    if extraction.Bathrooms > 0 { values["bathrooms"] = extraction.Bathrooms }

    return &models.ExtractionResult{
        DocumentURL: documentURL,
        Values:      values,
        Confidence:  extraction.Confidence,
    }, nil
}

// parseUserMessage extracts structured data from natural language.
func (s *InterrogatorService) parseUserMessage(
    ctx context.Context,
    message string,
    currentState map[string]interface{},
) (*models.ExtractionResult, error) {
    prompt := prompts.MessageParsingPrompt(message, currentState)
    
    result, err := s.vertexClient.GenerateContent(ctx, vertex.ModelFlash, prompt)
    if err != nil {
        return nil, err
    }

    var extraction struct {
        Values     map[string]any     `json:"values"`
        Confidence map[string]float64 `json:"confidence"`
    }
    
    if err := json.Unmarshal([]byte(result), &extraction); err != nil {
        return nil, err
    }

    return &models.ExtractionResult{
        Values:     extraction.Values,
        Confidence: extraction.Confidence,
    }, nil
}

// getNextQuestion determines what to ask based on missing fields.
func (s *InterrogatorService) getNextQuestion(state map[string]interface{}) (string, string) {
    for _, field := range models.GetPriorityFields() {
        if _, exists := state[field.Field]; !exists {
            return field.Field, field.Question
        }
    }
    return "", ""
}

// checkReadyToCreate verifies minimum required fields.
func (s *InterrogatorService) checkReadyToCreate(state map[string]interface{}) bool {
    _, hasName := state["name"]
    _, hasAddress := state["address"]
    return hasName && hasAddress
}

// generateExtractionSummary creates a natural language summary of what was extracted.
func (s *InterrogatorService) generateExtractionSummary(extraction *models.ExtractionResult) string {
    // TODO: Generate more natural summaries
    count := len(extraction.Values)
    return fmt.Sprintf("I found %d details from your blueprint. Review them in the form and let me know if anything needs to be corrected.", count)
}

// mergeStates combines current state with new extractions (new values win).
func (s *InterrogatorService) mergeStates(
    current map[string]interface{},
    extracted map[string]any,
) map[string]interface{} {
    merged := make(map[string]interface{})
    for k, v := range current {
        merged[k] = v
    }
    for k, v := range extracted {
        merged[k] = v
    }
    return merged
}
```

---

### Task 4: Create Prompts (1 hour)

```go
// internal/prompts/interrogator_prompt.go

package prompts

import "fmt"

// BlueprintExtractionPrompt returns the system prompt for document analysis.
func BlueprintExtractionPrompt() string {
    return `You are a construction project analyst. Extract the following information from this architectural blueprint or floor plan:

REQUIRED FIELDS:
- name: Project name (from title block)
- address: Full street address
- gsf: Gross Square Footage (total conditioned space)
- foundation_type: "slab", "crawlspace", or "basement"
- stories: Number of stories/levels
- bedrooms: Bedroom count
- bathrooms: Bathroom count

RESPOND IN JSON FORMAT ONLY:
{
  "name": "string or null",
  "address": "string or null", 
  "gsf": number or null,
  "foundation_type": "slab" | "crawlspace" | "basement" | null,
  "stories": number or null,
  "bedrooms": number or null,
  "bathrooms": number or null,
  "confidence": {
    "name": 0.0-1.0,
    "address": 0.0-1.0,
    "gsf": 0.0-1.0,
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
- gsf (square footage, look for "sq ft", "square feet", numbers like "3200")
- foundation_type ("slab", "crawlspace", "basement")
- stories (1, 2, etc or "single story", "two story")
- bedrooms (number)
- bathrooms (number)
- topography ("flat", "moderate", "steep")
- soil_conditions ("standard", "rock", "clay")

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
        "gsf": "The plans show approximately %.0f square feet. Is that correct?",
        "foundation_type": "I see what looks like a %s foundation. Can you confirm?",
        "bedrooms": "I found %d potential bedrooms. One looks like it might be an office—should I count it?",
        "bathrooms": "I see %d bathrooms. Is one of them a master en-suite?",
    }
    
    if template, ok := templates[field]; ok {
        return fmt.Sprintf(template, extractedValue)
    }
    return ""
}
```

---

### Task 5: Register Route (15 min)

```go
// In cmd/api/main.go or router setup

import "velocity-backend/internal/handlers"

// Add to router
r.Route("/api/v1/agent", func(r chi.Router) {
    r.Use(middleware.RequireAuth)
    r.Post("/onboard", onboardingHandler.HandleOnboard)
})
```

---

## 🔌 API Contract

### Endpoint: `POST /api/v1/agent/onboard`

**Request:**
```json
{
  "session_id": "temp_abc123",
  "message": "It's a 3,200 square foot custom home in Austin",
  "document_url": null,
  "current_state": {
    "name": "Smith Residence"
  }
}
```

**Response:**
```json
{
  "session_id": "temp_abc123",
  "reply": "Got it! A 3,200 sqft home in Austin. What type of foundation will this project have—slab, crawlspace, or basement?",
  "extracted_values": {
    "gsf": 3200,
    "address": "Austin, TX"
  },
  "confidence_scores": {
    "gsf": 0.95,
    "address": 0.60
  },
  "clarifying_question": null,
  "ready_to_create": false,
  "next_priority_field": "foundation_type"
}
```

---

## ✅ Acceptance Criteria

- [ ] `POST /api/v1/agent/onboard` returns valid JSON response
- [ ] Document upload triggers Vision API extraction
- [ ] User messages are parsed for structured data
- [ ] Extracted values include confidence scores
- [ ] `ready_to_create` is true only when name + address are present
- [ ] `next_priority_field` follows the priority matrix (P0 → P1 → P2)
- [ ] Low-confidence extractions (< 0.8) trigger clarifying questions
- [ ] All extracted fields map to valid physics engine inputs
- [ ] No schedule calculations occur in this service (Layer 4 only)
- [ ] Error handling returns user-friendly messages

---

## 🧪 Verification Plan

### Unit Tests
```go
// internal/services/interrogator_service_test.go
func TestGetNextQuestion_ReturnsNameFirst(t *testing.T) { ... }
func TestCheckReadyToCreate_RequiresNameAndAddress(t *testing.T) { ... }
func TestMergeStates_NewValuesWin(t *testing.T) { ... }
```

### Integration Tests
```bash
# Test with curl
curl -X POST http://localhost:8080/api/v1/agent/onboard \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"test","message":"3 bedroom home in Austin","current_state":{}}'
```

### Manual Testing
1. Send message with square footage → Verify GSF extracted
2. Upload a test blueprint PDF → Verify extraction response
3. Test conversation flow until `ready_to_create` is true

---

## 📚 Reference Documents

- [PHASE_11_PRD.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/planning/PHASE_11_PRD.md) Section 5.2, 5.5
- [AGENT_BEHAVIOR_SPEC.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/AGENT_BEHAVIOR_SPEC.md) - Agent patterns
- [BACKEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/BACKEND_SCOPE.md) - Go patterns
- [CPM_RES_MODEL_SPEC.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/CPM_RES_MODEL_SPEC.md) - Physics engine inputs
- [API_AND_TYPES_SPEC.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/API_AND_TYPES_SPEC.md) - API patterns

# Onboarding API Documentation

## POST /api/v1/agent/onboard

Smart onboarding wizard - extracts project details from conversation or blueprints.

**Authentication:** Required (JWT Bearer token)

**Rate Limiting:** Future implementation planned (currently protected by auth only)

**Timeout:** 60 seconds (AI operations can be slow for large blueprints)

### Request

**Content-Type:** `application/json`

**Request Body:**

| Field | Type | Required | Description | Constraints |
|-------|------|----------|-------------|-------------|
| `session_id` | string | Yes | Wizard session identifier | Max 100 characters |
| `message` | string | No | User's natural language message | Max 10,000 characters |
| `document_url` | string | No | Blueprint image URL | Max 2,000 characters, http/https only |
| `current_state` | object | Yes | Current form values (map of field name to value) | Max 50 fields |

**Example Request:**

```json
{
  "session_id": "onboard_abc123",
  "message": "3200 sqft home in Austin with slab foundation",
  "document_url": "",
  "current_state": {}
}
```

**Example Request with Document:**

```json
{
  "session_id": "onboard_abc123",
  "message": "",
  "document_url": "https://example.com/blueprints/plan.pdf",
  "current_state": {
    "name": "Smith Residence"
  }
}
```

### Response

**Success Status Code:** `200 OK`

**Content-Type:** `application/json`

**Response Body:**

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | Echo of request session_id |
| `reply` | string | Agent's conversational response to display to user |
| `extracted_values` | object | New field values extracted from message/document |
| `confidence_scores` | object | Confidence per field (0.0-1.0) |
| `ready_to_create` | boolean | True if name + address are present (minimum required fields) |
| `next_priority_field` | string | Next missing P0/P1 field to ask about |
| `clarifying_question` | string | Question to ask user for next_priority_field |

**Example Response (Partial Data):**

```json
{
  "session_id": "onboard_abc123",
  "reply": "I see you're planning a 3200 sqft home in Austin with a slab foundation. What would you like to name this project?",
  "extracted_values": {
    "address": "Austin",
    "gsf": 3200,
    "foundation_type": "slab"
  },
  "confidence_scores": {
    "address": 0.90,
    "gsf": 0.95,
    "foundation_type": 0.88
  },
  "ready_to_create": false,
  "next_priority_field": "name",
  "clarifying_question": "What would you like to name this project?"
}
```

**Example Response (Ready to Create):**

```json
{
  "session_id": "onboard_abc123",
  "reply": "Great! Your project is ready to create. Review the details and click 'Create Project' when ready.",
  "extracted_values": {
    "name": "Smith Residence"
  },
  "confidence_scores": {
    "name": 0.98
  },
  "ready_to_create": true,
  "next_priority_field": "",
  "clarifying_question": ""
}
```

### Extractable Fields

The following fields can be extracted from user messages or blueprint documents:

| Field | Type | Priority | Description |
|-------|------|----------|-------------|
| `name` | string | P0 | Project name (required to create) |
| `address` | string | P0 | Project address (required to create) |
| `gsf` | number | P1 | Gross square footage |
| `foundation_type` | string | P1 | Foundation type (slab, crawl_space, basement, pier) |
| `stories` | integer | P1 | Number of stories |
| `bedrooms` | integer | P2 | Number of bedrooms |
| `bathrooms` | integer | P2 | Number of bathrooms |

**Priority Matrix:**
- **P0:** Required fields for project creation (name, address)
- **P1:** Critical for accurate scheduling (gsf, foundation_type, stories)
- **P2:** Nice to have for better estimates (bedrooms, bathrooms)

### Error Responses

#### 400 Bad Request

Returned when request validation fails.

**Possible Reasons:**
- `session_id` is empty or too long (>100 chars)
- `message` is too long (>10,000 chars)
- `document_url` is invalid format or too long (>2,000 chars)
- `current_state` is null or has too many fields (>50)

**Example:**

```json
HTTP/1.1 400 Bad Request
Content-Type: text/plain

session_id is required
```

#### 401 Unauthorized

Returned when authentication is missing or invalid.

**Possible Reasons:**
- No `Authorization` header
- Invalid or expired JWT token
- User context missing from token

**Example:**

```json
HTTP/1.1 401 Unauthorized
Content-Type: text/plain

Missing user_id in context
```

#### 413 Request Entity Too Large

Returned when request body exceeds 1MB limit.

**Example:**

```json
HTTP/1.1 413 Request Entity Too Large
Content-Type: text/plain

http: request body too large
```

#### 500 Internal Server Error

Returned when AI processing fails or internal errors occur.

**Note:** Error details are logged server-side but not exposed to clients for security.

**Example:**

```json
HTTP/1.1 500 Internal Server Error
Content-Type: text/plain

Failed to process request. Please try again.
```

### Security

#### SSRF Protection

Document URLs are validated to prevent Server-Side Request Forgery attacks:

- **Allowed schemes:** `http://`, `https://` only (blocks `file://`, `ftp://`, etc.)
- **Blocked IPs:** Private IP ranges (10.0.0.0/8, 192.168.0.0/16, 172.16.0.0/12, 127.0.0.0/8)
- **AWS metadata blocked:** 169.254.169.254 (common SSRF target)
- **Redirects disabled:** Prevents redirect-based SSRF attacks
- **Timeout:** 30 seconds for document download

#### File Upload Restrictions

When uploading blueprints via `document_url`:

- **Max file size:** 50MB
- **Allowed MIME types:**
  - `image/jpeg`
  - `image/png`
  - `image/webp`
  - `application/pdf`
- **Disallowed types:** Executables, scripts, HTML, JSON, etc.

#### Rate Limiting

**Current:** None (protected by auth middleware only)

**Future:** Rate limiting will be added if abuse is detected:
- Limit: 10 requests per minute per user
- Burst: 20 requests
- Scope: Per-user quota based on JWT `user_id`

### Usage Example

#### Multi-Turn Conversation Flow

**Turn 1: Initial Message**

```bash
curl -X POST https://api.futurebuild.com/api/v1/agent/onboard \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "onboard_session_1",
    "message": "3200 sqft home in Austin",
    "current_state": {}
  }'
```

**Response:**

```json
{
  "session_id": "onboard_session_1",
  "reply": "I see you're planning a 3200 sqft home in Austin. What would you like to name this project?",
  "extracted_values": {
    "address": "Austin",
    "gsf": 3200
  },
  "confidence_scores": {
    "address": 0.90,
    "gsf": 0.95
  },
  "ready_to_create": false,
  "next_priority_field": "name"
}
```

**Turn 2: Provide Name**

```bash
curl -X POST https://api.futurebuild.com/api/v1/agent/onboard \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "onboard_session_1",
    "message": "Smith Residence",
    "current_state": {
      "address": "Austin",
      "gsf": 3200
    }
  }'
```

**Response:**

```json
{
  "session_id": "onboard_session_1",
  "reply": "Great! Your project is ready to create.",
  "extracted_values": {
    "name": "Smith Residence"
  },
  "confidence_scores": {
    "name": 0.98
  },
  "ready_to_create": true,
  "next_priority_field": ""
}
```

#### Blueprint Upload Flow

```bash
curl -X POST https://api.futurebuild.com/api/v1/agent/onboard \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "onboard_session_2",
    "document_url": "https://storage.example.com/blueprints/plan.pdf",
    "current_state": {}
  }'
```

**Response:**

```json
{
  "session_id": "onboard_session_2",
  "reply": "I found a 3200 sqft, 2-story home with slab foundation and 4 bedrooms. What would you like to name this project?",
  "extracted_values": {
    "address": "123 Main St, Austin, TX",
    "gsf": 3200,
    "foundation_type": "slab",
    "stories": 2,
    "bedrooms": 4,
    "bathrooms": 3
  },
  "confidence_scores": {
    "address": 0.85,
    "gsf": 0.92,
    "foundation_type": 0.88,
    "stories": 0.95,
    "bedrooms": 0.90,
    "bathrooms": 0.87
  },
  "ready_to_create": false,
  "next_priority_field": "name"
}
```

### Implementation Notes

**Architecture Layer:** Layer 4 (Action Engine)

**No Physics Calculations:** This endpoint only extracts and validates data. Schedule calculations happen in the Physics Engine (Layer 3) after project creation.

**Stateless Service:** Session state is maintained client-side in `current_state`. The service is stateless and does not persist sessions to database.

**AI Model:** Google Vertex AI (Gemini 2.5 Flash)
- Text extraction: Gemini Flash (fast, cost-effective)
- Vision extraction: Gemini Flash with multimodal input

**Database:** None required (stateless)

### Monitoring

**Logs:**
- `onboarding_message_received` - Request received
- `onboarding_message_completed` - Request completed with metrics
- `blueprint_download_failed` - Document download error
- `ai_extraction_failed` - AI processing error
- `onboarding_request_failed` - Handler-level error

**Metrics (Future):**
- `interrogator_requests_total` - Total requests counter
- `interrogator_extraction_duration_seconds` - Latency histogram
- `interrogator_confidence_avg` - Average confidence by field
- `interrogator_errors_total` - Error counter by type

### Related Documentation

- **Spec:** `/specs/committed/STEP_75_INTERROGATOR_AGENT.md`
- **PRD:** `/planning/PHASE_11_PRD.md` (Step 75)
- **Frontend:** `/frontend/src/components/views/fb-view-onboarding.ts`
- **State Management:** `/frontend/src/store/onboarding-store.ts`

### Changelog

**Version 1.0 (Step 75):**
- Initial release with conversational extraction
- Blueprint vision analysis
- SSRF protection and security hardening
- P0/P1/P2 priority matrix for questioning

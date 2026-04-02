# API Contract Specification

**System:** FutureBuild Brain (System of Connection)
**Pipeline Stage:** 07 - Architecture Spec
**Date:** 2026-04-02
**Status:** COMPLETE

---

## 1. OIDC Provider Endpoints (Public)

FB-Brain IS the OIDC identity provider for the entire FutureBuild ecosystem. All endpoints follow the OpenID Connect Core 1.0 specification.

### 1.1 Discovery

### GET /.well-known/openid-configuration
- **Auth:** None
- **Response:**
```json
{
  "issuer": "https://brain.futurebuild.io",
  "authorization_endpoint": "https://brain.futurebuild.io/authorize",
  "token_endpoint": "https://brain.futurebuild.io/token",
  "userinfo_endpoint": "https://brain.futurebuild.io/userinfo",
  "jwks_uri": "https://brain.futurebuild.io/jwks",
  "revocation_endpoint": "https://brain.futurebuild.io/revoke",
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code", "refresh_token"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "profile", "email", "org"],
  "token_endpoint_auth_methods_supported": ["none", "client_secret_post"],
  "code_challenge_methods_supported": ["S256"]
}
```

### 1.2 Authorization

### GET /authorize
- **Auth:** None (redirects to login)
- **Query Parameters:**
  - `response_type=code` (required)
  - `client_id` (required) — registered OIDC client ID
  - `redirect_uri` (required) — must match registered redirect URIs
  - `scope=openid+profile+email+org` (required)
  - `state` (required) — CSRF protection
  - `nonce` (optional) — replay protection
  - `code_challenge` (required for public clients) — PKCE S256
  - `code_challenge_method=S256` (required with code_challenge)
- **Flow:**
  1. Validates client_id and redirect_uri
  2. Redirects to login page (magic link email)
  3. After authentication, shows consent screen (first login only)
  4. Redirects to redirect_uri with `?code={auth_code}&state={state}`
- **Errors:** Redirects with `?error=invalid_request&error_description=...`

### 1.3 Token Exchange

### POST /token
- **Auth:** None (public clients with PKCE) or client_secret_post
- **Content-Type:** application/x-www-form-urlencoded
- **Body (Authorization Code):**
  - `grant_type=authorization_code`
  - `code` — authorization code from /authorize
  - `redirect_uri` — must match the original request
  - `client_id` — OIDC client ID
  - `code_verifier` — PKCE verifier (required for public clients)
- **Body (Refresh Token):**
  - `grant_type=refresh_token`
  - `refresh_token` — previously issued refresh token
  - `client_id` — OIDC client ID
- **Response (200):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
  "id_token": "eyJhbGciOiJSUzI1NiIs..."
}
```
- **Errors:** `400 { error: "invalid_grant" }`, `401 { error: "invalid_client" }`

### 1.4 UserInfo

### GET /userinfo
- **Auth:** Bearer access_token
- **Response (200):**
```json
{
  "sub": "uuid",
  "email": "user@example.com",
  "name": "Tom Builder",
  "org_id": "uuid",
  "role": "owner"
}
```

### 1.5 JWKS (JSON Web Key Set)

### GET /jwks
- **Auth:** None
- **Response (200):**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "brain-key-2026-04",
      "use": "sig",
      "alg": "RS256",
      "n": "base64url-encoded-modulus",
      "e": "AQAB"
    }
  ]
}
```
- **Caching:** Clients should cache with 1-hour TTL. Refresh on JWT validation failure (key rotation).

### 1.6 Token Revocation

### POST /revoke
- **Auth:** client_id in body (or client_secret_post)
- **Content-Type:** application/x-www-form-urlencoded
- **Body:** `token={refresh_token}&client_id={client_id}`
- **Response:** `200 OK` (always, even if token was already revoked — per RFC 7009)

### 1.7 Consent

### GET /consent
- **Auth:** Session cookie (authenticated user)
- **Query:** `?auth_request_id={uuid}`
- **Response:** HTML consent page showing requested scopes and client name

### POST /consent
- **Auth:** Session cookie
- **Body:** `auth_request_id={uuid}&action=accept|deny`
- **Response:** Redirect to client's redirect_uri with auth code (accept) or error (deny)

---

## 2. Token Claims

### 2.1 Access Token Claims (JWT, RS256)

| Claim | Type | Description |
|-------|------|-------------|
| `iss` | string | `https://brain.futurebuild.io` |
| `sub` | string | User ID (UUID) |
| `aud` | string | Client ID (e.g., `fb-os`) |
| `exp` | int | Expiry (1 hour from issuance) |
| `iat` | int | Issued-at timestamp |
| `jti` | string | Unique token ID |
| `org_id` | string | Organization ID (UUID) |
| `role` | string | `owner`, `admin`, `member` |
| `plan_tier` | string | `free`, `pro`, `enterprise` |

### 2.2 ID Token Claims (JWT, RS256)

All access token claims plus:
| Claim | Type | Description |
|-------|------|-------------|
| `email` | string | User's email address |
| `name` | string | Display name |
| `auth_time` | int | Time of authentication |
| `nonce` | string | Nonce from authorization request (if provided) |

---

## 3. Hub Admin API

All Hub endpoints require session authentication (magic link login). These power the Lit Admin UI.

### 3.1 Dashboard

### GET /api/hub/dashboard
- **Auth:** Session
- **Response:**
```json
{
  "data": {
    "connected_integrations": 4,
    "active_rfqs": 7,
    "pending_approvals": 2,
    "recent_events": [{ "event_type": "quote_created", "source": "gable", "created_at": "timestamp" }],
    "maestro_session_count": 12
  }
}
```

### 3.2 MCP Server Registry

### GET /api/hub/registry/servers
- **Auth:** Session
- **Response:** `200 { data: { servers: []MCPServer } }`

### GET /api/hub/registry/servers/{slug}
- **Auth:** Session
- **Response:** `200 { data: { server: MCPServer, tools: []MCPTool, triggers: []MCPTrigger } }`

### POST /api/hub/registry/servers
- **Auth:** Session (admin only)
- **Body:** `{ slug, name, description?, base_url, auth_type, auth_config? }`
- **Response:** `201 { data: { server: MCPServer } }`

### PUT /api/hub/registry/servers/{slug}
- **Auth:** Session (admin only)
- **Body:** `{ name?, description?, base_url?, auth_type?, auth_config?, is_active? }`
- **Response:** `200 { data: { server: MCPServer } }`

### DELETE /api/hub/registry/servers/{slug}
- **Auth:** Session (admin only)
- **Response:** `204 No Content`

### MCPServer Object
```json
{
  "id": "uuid",
  "slug": "gable",
  "name": "GableERP",
  "description": "Material supplier integration",
  "base_url": "https://api.gable.com/v1",
  "auth_type": "api_key",
  "health_status": "healthy|degraded|down|unknown",
  "health_checked_at": "timestamp",
  "is_active": true
}
```

### MCPTool Object
```json
{
  "id": "uuid",
  "server_id": "uuid",
  "slug": "get_products_by_category",
  "name": "Get Products by Category",
  "description": "Retrieve products filtered by category",
  "http_method": "POST",
  "path_template": "/products/search",
  "input_schema": { "type": "object", "properties": { "category": { "type": "string" } }, "required": ["category"] }
}
```

### 3.3 OIDC Client Management

### GET /api/hub/oidc/clients
- **Auth:** Session (admin only)
- **Response:** `200 { data: { clients: []OIDCClient } }`

### POST /api/hub/oidc/clients
- **Auth:** Session (admin only)
- **Body:** `{ client_id, client_name, redirect_uris: [], grant_types?, response_types?, scopes? }`
- **Response:** `201 { data: { client: OIDCClient } }`

### PUT /api/hub/oidc/clients/{id}
- **Auth:** Session (admin only)
- **Body:** `{ client_name?, redirect_uris?, is_active? }`
- **Response:** `200 { data: { client: OIDCClient } }`

### DELETE /api/hub/oidc/clients/{id}
- **Auth:** Session (admin only)
- **Response:** `204 No Content`

### 3.4 Integration Connections

### GET /api/hub/connections
- **Auth:** Session
- **Response:** `200 { data: { connections: []Connection } }`

### POST /api/hub/connections/{serverSlug}/connect
- **Auth:** Session
- **Body:** `{ api_key?: "string" }` (for API key auth) or empty (for OAuth2 — redirects)
- **Response:** `200 { data: { connection: Connection } }` | `302 Redirect` (OAuth2 flow)

### DELETE /api/hub/connections/{serverSlug}/disconnect
- **Auth:** Session
- **Response:** `204 No Content`

### 3.5 OAuth2 External Flows

### GET /api/hub/oauth/{provider}/authorize
- **Auth:** Session
- **Description:** Initiates OAuth2 authorization code flow with external provider (QuickBooks, Gmail, Outlook)
- **Response:** `302 Redirect` to provider's authorization endpoint

### GET /api/hub/oauth/{provider}/callback
- **Auth:** Session (via cookie)
- **Query:** `?code={auth_code}&state={state}`
- **Description:** Handles OAuth2 callback, exchanges code for tokens, stores encrypted credentials
- **Response:** `302 Redirect` to Hub connections page

### 3.6 Maestro Chat

### POST /api/hub/chat/message
- **Auth:** Session
- **Body:** `{ message: "string", session_id?: "uuid" }`
- **Response:**
```json
{
  "data": {
    "session_id": "uuid",
    "response": "I found 3 matching products from GableERP...",
    "intent": "product_search",
    "confidence": 0.92,
    "tier_used": "claude",
    "tools_called": [
      { "server": "gable", "tool": "get_products_by_category", "latency_ms": 450 }
    ]
  }
}
```

### GET /api/hub/chat/sessions
- **Auth:** Session
- **Response:** `200 { data: { sessions: []MaestroSession } }`

### GET /api/hub/chat/sessions/{sessionID}/messages
- **Auth:** Session
- **Response:** `200 { data: { messages: []MaestroMessage } }`

### 3.7 Real-time Events (SSE)

### GET /api/hub/events/stream
- **Auth:** Session
- **Content-Type:** text/event-stream
- **Events:**
```
event: integration_event
data: { "event_type": "quote_created", "source": "gable", "rfq_id": "uuid", "timestamp": "..." }

event: webhook_delivered
data: { "event_type": "review_material_quote", "target": "fb-os", "status": "delivered" }

event: health_check
data: { "server_slug": "gable", "status": "healthy" }

event: maestro_tool_call
data: { "server": "gable", "tool": "get_products_by_category", "latency_ms": 450 }
```

---

## 4. MCP Protocol Endpoints

### POST /mcp/tools/list
- **Auth:** Session
- **Response:**
```json
{
  "tools": [
    {
      "server_slug": "gable",
      "tool_slug": "get_products_by_category",
      "name": "Get Products by Category",
      "description": "...",
      "input_schema": { ... }
    }
  ]
}
```
- **Note:** Only returns tools from servers the user's org has connected

### POST /mcp/tools/call
- **Auth:** Session
- **Body:**
```json
{
  "server_slug": "gable",
  "tool_slug": "get_products_by_category",
  "input": { "category": "lumber" }
}
```
- **Response:**
```json
{
  "data": {
    "result": { ... },
    "server_slug": "gable",
    "tool_slug": "get_products_by_category",
    "latency_ms": 450,
    "trace_id": "otel-trace-id"
  }
}
```
- **Flow:**
  1. Validates input against tool's JSON Schema (input_schema)
  2. Retrieves encrypted credentials for user's org
  3. Constructs and executes HTTP request to external server
  4. Logs IntegrationEvent
  5. Returns result

### POST /mcp/triggers/subscribe
- **Auth:** Session
- **Body:** `{ server_slug: "gable", trigger_slug: "price_updated", callback_url?: "string" }`
- **Response:** `200 { data: { subscription_id: "uuid" } }`

---

## 5. A2A Webhook Architecture (Outbound)

FB-Brain emits JWS-signed webhooks to FB-OS for cross-system coordination.

### 5.1 Webhook Emission

- **Target:** FB-OS at `POST /api/v1/a2a/webhook`
- **Signing:** RS256 JWS detached compact serialization (go-jose/v4)
- **Headers:**
  - `X-JWS-Signature`: Detached JWS signature
  - `X-Idempotency-Key`: UUID v7 for deduplication
  - `Content-Type`: application/json

### 5.2 Webhook Event Envelope
```json
{
  "event_type": "string",
  "payload": { ... },
  "trace_id": "otel-trace-id",
  "idempotency_key": "uuid",
  "timestamp": "2026-04-02T14:30:00Z",
  "iss": "fb-brain"
}
```

### 5.3 Event Types and Payloads (Composite Currency Pattern)

#### review_material_quote
Emitted when a materials quote is ready for review on FB-OS.
```json
{
  "rfq_id": "uuid",
  "line_items": [
    { "name": "2x4 Lumber", "quantity": 500, "unit_price_cents": 450, "currency_code": "USD" }
  ],
  "total_cents": 225000,
  "currency_code": "USD",
  "vendor": "GableERP"
}
```

#### review_labor_bid
Emitted when a labor bid is received and analyzed.
```json
{
  "rfq_id": "uuid",
  "bidder": "Apex Roofing",
  "amount_cents": 1200000,
  "currency_code": "USD",
  "timeline": "14 days",
  "ai_analysis": "Competitive bid, strong safety record, 4.2/5 past rating"
}
```

#### update_schedule
Emitted when material delivery dates affect CPM schedule.
```json
{
  "event_type": "material_delivery",
  "delivery_date": "2026-05-15",
  "constraints": { "wbs_codes": ["9.1", "9.2"] }
}
```

#### delivery_confirmation
Emitted when materials + labor procurement convergence check completes.
```json
{
  "materials_ordered": true,
  "labor_approved": false,
  "convergence_status": "partial"
}
```

#### create_feed_card
Emitted to create a notification card in FB-OS feed.
```json
{
  "card_type": "procurement",
  "title": "Quote Ready for Review",
  "body": "GableERP quote #Q-2026-0042 is ready",
  "actions": [{ "label": "Review", "action_type": "open_quote", "payload": { "rfq_id": "uuid" } }],
  "priority": "urgent"
}
```

### 5.4 Retry Strategy (Exponential Backoff)

| Attempt | Delay |
|---------|-------|
| 1 | Immediate |
| 2 | 30 seconds |
| 3 | 1 minute |
| 4 | 2 minutes |
| 5 | 5 minutes |
| 6 | 15 minutes |
| 7 | 1 hour |
| Dead letter | After 7 failures → alert + manual intervention |

### 5.5 Webhook Delivery Log

All webhook emissions are logged in `a2a_webhook_log`:
- `status`: pending → delivered | failed → dead_letter
- `retry_count`: 0-7
- `idempotency_key`: prevents duplicate processing on FB-OS side
- `delivery_ms`: round-trip latency

---

## 6. MCP Server Inventory

### 6.1 Registered Servers (MVP)

| Server | Slug | Auth | Tools | Triggers |
|--------|------|------|-------|----------|
| GableERP | `gable` | API Key | get_products_by_category, bulk_calculate_price, create_quote, accept_and_convert_quote | price_updated, order_status_changed |
| LocalBlue | `localblue` | Host Header | push_rfq, update_bid_status | bid_submitted |
| XUI Projects | `xui` | API Key | create_feed_card, assign_contact_to_phase | card_action_approved, card_action_rejected, card_action_custom |
| QuickBooks | `quickbooks` | OAuth2 | create_purchase_order, create_invoice, create_bill, get_company_info | payment_received, invoice_updated |
| 1Build | `onebuild` | Auth Header | get_estimate, search_cost_data, get_line_items | — |
| Gmail | `gmail` | OAuth2 | send_email, list_emails, get_email | email_received |
| Outlook | `outlook` | OAuth2 | send_email, list_emails, get_email | email_received |

### 6.2 Tool Call Latency Budgets

| Server | Max Latency | Notes |
|--------|------------|-------|
| GableERP | <3s | Product search may be slower for large catalogs |
| LocalBlue | <3s | Host header auth, minimal overhead |
| QuickBooks | <3s | OAuth2 token refresh adds ~200ms when expired |
| 1Build | <3s | Cost data search can be heavy |
| Gmail/Outlook | <3s | OAuth2 token refresh if expired |
| Anthropic Claude | <2s | Intent classification only |

---

## 7. Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `VALIDATION_ERROR` | Invalid request body or query parameters |
| 400 | `INVALID_GRANT` | Invalid authorization code or refresh token |
| 401 | `UNAUTHORIZED` | Missing or invalid session / access token |
| 401 | `INVALID_CLIENT` | Unknown client_id |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 403 | `CONSENT_DENIED` | User denied consent |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Duplicate resource |
| 422 | `SCHEMA_VALIDATION_ERROR` | Tool input doesn't match JSON Schema |
| 429 | `RATE_LIMITED` | Brute force protection on auth endpoints |
| 500 | `INTERNAL_ERROR` | Server error |
| 502 | `UPSTREAM_ERROR` | External MCP server returned an error |
| 504 | `UPSTREAM_TIMEOUT` | External MCP server call timed out (>3s) |

---

## 8. Rate Limits

### 8.1 Auth Endpoints (Brute Force Protection)

| Endpoint | Limit | Window |
|----------|-------|--------|
| POST /token | 20 requests | per minute per client_id |
| Magic link request | 5 requests | per hour per email |

### 8.2 API Endpoints

| Tier | Requests/min | Burst |
|------|-------------|-------|
| free | 60 | 10 |
| pro | 300 | 50 |
| enterprise | 1000 | 200 |

### 8.3 MCP Tool Calls

| Limit | Value | Reason |
|-------|-------|--------|
| Per-org concurrent tool calls | 5 | Prevent external API abuse |
| Per-session tool calls/min | 30 | Fair usage |

---

## 9. Cross-System Contract Summary

### 9.1 Brain Provides to OS

| Interface | Protocol | Endpoint | SLA |
|-----------|----------|----------|-----|
| Identity | OIDC | `/.well-known/openid-configuration` | 99.99% uptime |
| JWT Signing | JWKS | `/jwks` | <10ms validation (cached) |
| Token Issuance | OIDC | `/token` | <500ms p99 |
| Webhooks | A2A (JWS) | Brain → OS `/api/v1/a2a/webhook` | <500ms delivery, >99.9% rate |

### 9.2 Brain Depends on OS

| Interface | Protocol | Endpoint | Purpose |
|-----------|----------|----------|---------|
| Webhook Receiver | A2A | OS `/api/v1/a2a/webhook` | Deliver integration events |
| Approval Callbacks | REST | OS → Brain callback | Quote/bid approval events |

### 9.3 Brain Depends on External Systems

| System | Protocol | Auth | Latency Budget |
|--------|----------|------|----------------|
| GableERP | REST | API Key | <3s per tool call |
| LocalBlue | REST | Host Header | <3s per tool call |
| QuickBooks | REST + CloudEvents | OAuth2 | <3s per tool call |
| 1Build | REST | Auth Header | <3s per tool call |
| Gmail | REST | OAuth2 | <3s per tool call |
| Outlook | REST | OAuth2 | <3s per tool call |
| Anthropic Claude | REST | API Key | <2s classification |

---

## 10. Versioning

- OIDC endpoints are unversioned (standard-compliant, stable)
- Hub API: `/api/hub/...` (unversioned, admin-only, internal)
- MCP Protocol: `/mcp/...` (unversioned, follows MCP spec)
- Breaking changes communicated via `Sunset` header (RFC 8594) with 6-month lead time

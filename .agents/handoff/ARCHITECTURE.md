# Architecture Specification

**System:** FutureBuild Brain (System of Connection)
**Pipeline Stage:** 07 - Architecture Spec
**Date:** 2026-04-02
**Status:** COMPLETE

---

## 1. System Overview

FutureBuild Brain is the identity, integration, and orchestration layer for the FutureBuild ecosystem. It operates as an OIDC identity provider, hosts the MCP Server Registry, provides the Maestro AI co-pilot for natural-language workflow execution, and emits JWS-signed A2A webhooks to coordinate with FutureBuild OS.

```
┌─────────────────────────────────────────────────────────────────┐
│                       FutureBuild Brain                         │
│                                                                 │
│  ┌──────────┐  ┌──────────────┐  ┌───────────┐  ┌───────────┐ │
│  │ Lit Admin│  │ OIDC Provider│  │ MCP       │  │ A2A       │ │
│  │ UI       │  │ Endpoints   │  │ Registry  │  │ Webhook   │ │
│  └────┬─────┘  └──────┬───────┘  └─────┬─────┘  └─────┬─────┘ │
│       │               │                │               │       │
│       └───────────────┴────────────────┴───────────────┘       │
│                            │                                    │
│  ┌─────────────────────────┴──────────────────────────────┐    │
│  │                   Maestro Orchestrator                   │    │
│  │  Intent Classification (regex → Claude → human)         │    │
│  │  Flow Coordination (Materials, Labor, Post-Approval)    │    │
│  └─────────────────────────┬──────────────────────────────┘    │
│                            │                                    │
│  ┌────────────┐  ┌─────────┴─────────┐  ┌──────────────────┐  │
│  │ OIDC       │  │   Repository      │  │ MCP Server       │  │
│  │ Token      │  │   (pgx/v5)        │  │ Clients          │  │
│  │ Service    │  │                    │  │ GableERP, QB,    │  │
│  │            │  │                    │  │ LocalBlue, XUI   │  │
│  └────────────┘  └─────────┬─────────┘  └──────────────────┘  │
│                            │                                    │
│                   ┌────────┴────────┐                          │
│                   │  PostgreSQL 16  │                          │
│                   └─────────────────┘                          │
└─────────────────────────────────────────────────────────────────┘
          │                                    │
          │ OIDC (JWT issuance)               │ A2A Webhooks (JWS-signed)
          ▼                                    ▼
    ┌─────────────┐                     ┌──────────────┐
    │  FB-OS      │                     │  FB-OS       │
    │ JWT Validate│                     │ Webhook Recv │
    └─────────────┘                     └──────────────┘

          │                                    │
          │ MCP Tool Calls                    │ MCP Triggers
          ▼                                    ▲
    ┌──────────────────────────────────────────────────┐
    │          External Integration Systems             │
    │  GableERP · LocalBlue · QuickBooks · 1Build      │
    │  Gmail · Outlook · XUI Projects                  │
    └──────────────────────────────────────────────────┘
```

---

## 2. Go Package Structure

```
futurebuild-brain/
├── cmd/
│   └── server/
│       └── main.go          # Wires all components, starts HTTP server
├── internal/
│   ├── api/                 # HTTP layer
│   │   ├── router.go        # Chi router with middleware stack
│   │   └── middleware/
│   │       ├── cors.go      # CORS for OIDC endpoints
│   │       ├── session.go   # Session middleware (legacy auth during migration)
│   │       ├── telemetry.go # OpenTelemetry + Prometheus
│   │       └── ratelimit.go # Brute force protection for auth endpoints
│   │
│   ├── oidc/                # OIDC Provider (zitadel/oidc v3)
│   │   ├── provider.go      # OIDC server configuration and initialization
│   │   ├── storage.go       # OIDCStorage interface implementation (PostgreSQL)
│   │   ├── keys.go          # RSA key pair management for JWT signing
│   │   ├── claims.go        # Custom claims: org_id, role, plan_tier
│   │   ├── consent.go       # Consent screen handler
│   │   └── clients.go       # OIDC client CRUD (admin API)
│   │
│   ├── mcp/                 # MCP Server Registry
│   │   ├── registry.go      # MCP registry service (PostgreSQL-backed)
│   │   ├── handler.go       # tools/list, tools/call HTTP handlers
│   │   ├── schema.go        # JSON Schema 2020-12 validation for tool inputs
│   │   ├── health.go        # Per-server health check runner
│   │   └── types.go         # MCPServer, MCPTool, MCPTrigger types
│   │
│   ├── maestro/             # AI Orchestrator
│   │   ├── orchestrator.go  # Main entry point: Classify → Route → Execute
│   │   ├── classifier.go    # Three-tier intent classification
│   │   │                    # Tier 1: regex fast-path (deterministic)
│   │   │                    # Tier 2: Claude tool-use (probabilistic)
│   │   │                    # Tier 3: human confirmation (low confidence)
│   │   ├── intents.go       # Intent definitions (top-20 construction intents)
│   │   ├── flows/
│   │   │   ├── materials.go # MaterialsFlow (evolves legacy 9-step pipeline)
│   │   │   ├── labor.go     # LaborFlow (evolves legacy 6-step pipeline)
│   │   │   └── post_approval.go  # Convergence check (materials + labor)
│   │   └── chat.go          # Maestro chat session management
│   │
│   ├── a2a/                 # Agent-to-Agent Webhooks
│   │   ├── client.go        # HTTP client with JWS signing (go-jose/v4)
│   │   ├── signer.go        # RS256 JWS detached signature generation
│   │   ├── types.go         # WebhookEvent types and payload schemas
│   │   └── retry.go         # Retry logic with exponential backoff
│   │
│   ├── hub/                 # Admin UI Backend (evolves legacy Hub)
│   │   ├── hub.go           # Hub coordinator — route registration
│   │   ├── registry_handlers.go   # MCP registry admin endpoints
│   │   ├── oidc_handlers.go       # OIDC client management endpoints
│   │   ├── integration_handlers.go # Integration connection management
│   │   ├── oauth_handlers.go      # OAuth2 flows for external systems (QuickBooks, Gmail)
│   │   └── event_bus.go           # SSE-based real-time event streaming
│   │
│   ├── clients/             # External system MCP server implementations
│   │   ├── gable.go         # GableERP MCP server (4 tools, 2 triggers)
│   │   ├── localblue.go     # LocalBlue MCP server (2 tools, 1 trigger)
│   │   ├── xui.go           # XUI Projects MCP server (2 tools, 3 triggers)
│   │   ├── quickbooks.go    # QuickBooks MCP server (4 tools, 2 triggers + CloudEvents)
│   │   ├── onebuild.go      # 1Build MCP server (3 tools)
│   │   ├── gmail.go         # Gmail MCP server (3 tools, 1 trigger)
│   │   └── outlook.go       # Outlook MCP server (3 tools, 1 trigger)
│   │
│   ├── models/              # Domain models
│   │   ├── user.go          # HubUser, OIDCSession
│   │   ├── org.go           # Organization, OrgMembership
│   │   ├── account_link.go  # AccountLink (cross-system identity mapping)
│   │   ├── rfq.go           # RFQ state machine (quote lifecycle)
│   │   ├── integration_event.go  # IntegrationEvent (audit trail)
│   │   ├── mcp_server.go    # MCPServer, MCPTool, MCPTrigger
│   │   └── oidc_client.go   # OIDCClient registration model
│   │
│   ├── store/               # Data access layer (raw SQL via pgx)
│   │   ├── pool.go          # pgxpool.Pool initialization
│   │   ├── user.go          # User/session queries
│   │   ├── org.go           # Organization queries
│   │   ├── account_link.go  # AccountLink CRUD
│   │   ├── rfq.go           # RFQ state transitions
│   │   ├── integration_event.go  # IntegrationEvent logging
│   │   ├── mcp_server.go    # MCP server registry queries
│   │   ├── oidc_client.go   # OIDC client queries
│   │   └── oidc_storage.go  # OIDC storage adapter (auth requests, tokens, keys)
│   │
│   └── config/              # Configuration
│       └── config.go        # Environment variable parsing
│
├── migrations/              # PostgreSQL migrations (goose)
│   ├── 001_initial_schema.sql
│   ├── 002_oidc_tables.sql
│   ├── 003_mcp_registry.sql
│   └── ...
├── frontend/                # Lit Web Components (Admin UI)
│   ├── src/
│   │   ├── components/
│   │   │   ├── base/
│   │   │   │   └── fb-element.ts        # FBBaseElement (shared with FB-OS)
│   │   │   ├── atoms/                   # Shared atom components
│   │   │   ├── molecules/
│   │   │   │   ├── fb-integration-card.ts
│   │   │   │   ├── fb-mcp-tool-card.ts
│   │   │   │   ├── fb-chat-bubble.ts
│   │   │   │   ├── fb-execution-row.ts
│   │   │   │   ├── fb-oidc-client-row.ts
│   │   │   │   └── fb-webhook-event.ts
│   │   │   ├── organisms/
│   │   │   │   ├── fb-mcp-registry.ts
│   │   │   │   ├── fb-oidc-manager.ts
│   │   │   │   ├── fb-ecosystem-canvas.ts
│   │   │   │   ├── fb-maestro-chat.ts
│   │   │   │   ├── fb-maestro-drawer.ts
│   │   │   │   └── fb-activity-ticker.ts
│   │   │   └── pages/
│   │   │       ├── fb-home-page.ts
│   │   │       ├── fb-ecosystem-page.ts
│   │   │       ├── fb-marketplace-page.ts
│   │   │       ├── fb-settings-page.ts
│   │   │       ├── fb-admin-registry.ts
│   │   │       └── fb-login-page.ts
│   │   ├── styles/
│   │   │   └── variables.css            # GableLBM tokens + integration colors
│   │   └── router.ts
│   ├── tailwind.config.ts
│   └── vite.config.ts
├── Makefile
├── Dockerfile
└── docker-compose.yml
```

---

## 3. PostgreSQL Schema

### 3.1 OIDC Provider Tables

```sql
-- OIDC Clients (registered relying parties)
CREATE TABLE oidc_clients (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id           TEXT NOT NULL UNIQUE,            -- Public identifier
    client_name         TEXT NOT NULL,
    redirect_uris       TEXT[] NOT NULL,
    grant_types         TEXT[] NOT NULL DEFAULT '{authorization_code}',
    response_types      TEXT[] NOT NULL DEFAULT '{code}',
    token_endpoint_auth TEXT NOT NULL DEFAULT 'none',    -- none (public+PKCE), client_secret_post
    scopes              TEXT[] NOT NULL DEFAULT '{openid,profile,email,org}',
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- OIDC Authorization Requests (ephemeral)
CREATE TABLE oidc_auth_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       TEXT NOT NULL REFERENCES oidc_clients(client_id),
    redirect_uri    TEXT NOT NULL,
    scopes          TEXT[] NOT NULL,
    state           TEXT NOT NULL,
    nonce           TEXT,
    code_challenge  TEXT,                    -- PKCE S256
    code_challenge_method TEXT DEFAULT 'S256',
    response_type   TEXT NOT NULL DEFAULT 'code',
    user_id         UUID,                    -- Set after authentication
    auth_code       TEXT UNIQUE,             -- Set after consent
    authenticated   BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL     -- Short-lived (10 min)
);
CREATE INDEX idx_auth_requests_code ON oidc_auth_requests(auth_code);

-- OIDC Refresh Tokens
CREATE TABLE oidc_refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash      TEXT NOT NULL UNIQUE,     -- SHA-256 of refresh token
    client_id       TEXT NOT NULL REFERENCES oidc_clients(client_id),
    user_id         UUID NOT NULL,
    scopes          TEXT[] NOT NULL,
    auth_time       TIMESTAMPTZ NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,             -- Null = active
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user ON oidc_refresh_tokens(user_id);

-- OIDC Signing Keys (RSA key pairs)
CREATE TABLE oidc_signing_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kid             TEXT NOT NULL UNIQUE,     -- Key ID for JWKS
    algorithm       TEXT NOT NULL DEFAULT 'RS256',
    private_key_pem TEXT NOT NULL,            -- Encrypted at rest
    public_key_pem  TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    rotated_at      TIMESTAMPTZ              -- Set when key is rotated
);
```

### 3.2 Identity & Organization Tables

```sql
-- Users (Brain-managed identity)
CREATE TABLE hub_users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    display_name    TEXT NOT NULL,
    magic_link_hash TEXT,                    -- SHA-256 of magic link token
    magic_link_exp  TIMESTAMPTZ,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Organizations
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    plan_tier   TEXT NOT NULL DEFAULT 'free',   -- free, pro, enterprise
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Org Memberships
CREATE TABLE org_memberships (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES hub_users(id),
    org_id      UUID NOT NULL REFERENCES organizations(id),
    role        TEXT NOT NULL DEFAULT 'member',  -- owner, admin, member
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, org_id)
);

-- Account Links (cross-system identity mapping)
CREATE TABLE account_links (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    user_id             UUID REFERENCES hub_users(id),
    gable_customer_id   TEXT,
    localblue_site_id   TEXT,
    localblue_user_id   TEXT,
    quickbooks_realm_id TEXT,
    link_type           TEXT NOT NULL DEFAULT 'full',  -- supplier, labor, full
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_account_links_org ON account_links(org_id);
```

### 3.3 MCP Server Registry Tables

```sql
-- MCP Servers (registered integration systems)
CREATE TABLE mcp_servers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    description     TEXT,
    icon            TEXT,
    base_url        TEXT NOT NULL,
    auth_type       TEXT NOT NULL,            -- api_key, oauth2, host_header
    auth_config     JSONB,                    -- Encrypted credential reference
    health_status   TEXT NOT NULL DEFAULT 'unknown',  -- healthy, degraded, down, unknown
    health_checked_at TIMESTAMPTZ,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- MCP Tools (operations on an MCP server)
CREATE TABLE mcp_tools (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    slug            TEXT NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    http_method     TEXT NOT NULL DEFAULT 'POST',
    path_template   TEXT NOT NULL,
    input_schema    JSONB NOT NULL,           -- JSON Schema 2020-12
    output_schema   JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(server_id, slug)
);
CREATE INDEX idx_tools_server ON mcp_tools(server_id);

-- MCP Triggers (events from an MCP server)
CREATE TABLE mcp_triggers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    slug            TEXT NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    event_type      TEXT NOT NULL DEFAULT 'webhook',  -- webhook, polling
    payload_schema  JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(server_id, slug)
);

-- Integration Credentials (encrypted per-org per-server)
CREATE TABLE integration_credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    server_id       UUID NOT NULL REFERENCES mcp_servers(id),
    credential_type TEXT NOT NULL,            -- api_key, oauth2_token
    encrypted_data  BYTEA NOT NULL,           -- AES-256-GCM encrypted
    expires_at      TIMESTAMPTZ,             -- For OAuth2 tokens
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, server_id)
);
```

### 3.4 Workflow & Audit Tables

```sql
-- RFQs (Request for Quote state machine) — Composite Currency Pattern
CREATE TABLE rfqs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    flow_type       TEXT NOT NULL,            -- materials, labor
    project_id      TEXT NOT NULL,            -- XUI project ID (external)
    card_id         TEXT,                     -- XUI card ID (external)
    status          TEXT NOT NULL DEFAULT 'pending',
    -- materials: pending -> quote_received -> approved -> ordered -> delivered
    -- labor: pending -> rfq_sent -> bid_received -> approved -> assigned
    scope_data      JSONB,                    -- Flexible scope items
    quote_data      JSONB,                    -- Quote/bid details
    total_cents     BIGINT,                   -- Total quote/bid amount
    currency_code   VARCHAR(3) NOT NULL DEFAULT 'USD',
    gable_quote_id  TEXT,
    gable_order_id  TEXT,
    localblue_rfq_id INTEGER,
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rfqs_org ON rfqs(org_id);
CREATE INDEX idx_rfqs_status ON rfqs(status);

-- Integration Events (immutable audit trail)
CREATE TABLE integration_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    rfq_id          UUID REFERENCES rfqs(id),
    event_type      TEXT NOT NULL,
    -- Event types: materials_flow_started, quote_created, quote_approved,
    --   order_placed, labor_flow_started, rfq_sent, bid_received,
    --   bid_approved, po_created, webhook_emitted, webhook_received
    source_system   TEXT NOT NULL,            -- fb-brain, gable, localblue, quickbooks, xui
    target_system   TEXT NOT NULL,
    payload         JSONB NOT NULL,
    trace_id        TEXT,                     -- OpenTelemetry trace ID
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_events_org ON integration_events(org_id);
CREATE INDEX idx_events_type ON integration_events(event_type);
CREATE INDEX idx_events_created ON integration_events(created_at);

-- A2A Webhook Delivery Log
CREATE TABLE a2a_webhook_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type      TEXT NOT NULL,
    target_url      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    jws_signature   TEXT NOT NULL,
    http_status     INTEGER,
    delivery_ms     INTEGER,
    retry_count     INTEGER NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending, delivered, failed, dead_letter
    idempotency_key UUID NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at    TIMESTAMPTZ
);
```

### 3.5 Maestro AI Tables

```sql
-- Maestro Chat Sessions
CREATE TABLE maestro_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES hub_users(id),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Maestro Messages (conversation history)
CREATE TABLE maestro_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES maestro_sessions(id),
    role            TEXT NOT NULL,             -- user, assistant, system
    content         TEXT NOT NULL,
    intent          TEXT,                      -- Classified intent (if applicable)
    confidence      DOUBLE PRECISION,          -- Classification confidence
    tier_used       TEXT,                      -- regex, claude, human
    tools_called    JSONB,                     -- MCP tools invoked [{server, tool, latency_ms}]
    token_count     INTEGER,
    cost_cents      BIGINT,                    -- Claude API cost in cents
    cost_currency_code VARCHAR(3) DEFAULT 'USD', -- Always USD for API billing
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_messages_session ON maestro_messages(session_id);
```

---

## 4. OIDC Provider Architecture

### 4.1 Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/.well-known/openid-configuration` | GET | OpenID Configuration document |
| `/authorize` | GET | Authorization endpoint (redirects to login) |
| `/token` | POST | Token endpoint (code exchange, refresh) |
| `/userinfo` | GET | UserInfo endpoint (protected, returns profile) |
| `/jwks` | GET | JSON Web Key Set (public keys for verification) |
| `/revoke` | POST | Token revocation |
| `/consent` | GET/POST | Consent screen (first-party only) |

### 4.2 Token Claims

```go
// Access Token Claims (JWT)
type AccessTokenClaims struct {
    jwt.RegisteredClaims            // iss, sub, aud, exp, iat, jti
    OrgID    string `json:"org_id"`
    Role     string `json:"role"`      // owner, admin, member
    PlanTier string `json:"plan_tier"` // free, pro, enterprise
}

// ID Token Claims (JWT)
type IDTokenClaims struct {
    jwt.RegisteredClaims
    Email       string `json:"email"`
    Name        string `json:"name"`
    OrgID       string `json:"org_id"`
    Role        string `json:"role"`
    AuthTime    int64  `json:"auth_time"`
    Nonce       string `json:"nonce,omitempty"`
}
```

### 4.3 Authorization Code Flow with PKCE

```
1. FB-OS redirects to Brain:
   GET /authorize?response_type=code&client_id=fb-os&redirect_uri=...
       &scope=openid+profile+email+org&state=...
       &code_challenge=...&code_challenge_method=S256

2. Brain authenticates user (magic link email)
3. Brain shows consent screen (first login only)
4. Brain redirects back with authorization code:
   302 -> redirect_uri?code=...&state=...

5. FB-OS exchanges code for tokens:
   POST /token
   Body: grant_type=authorization_code&code=...&redirect_uri=...
         &client_id=fb-os&code_verifier=...

6. Brain returns:
   { access_token: "eyJ...", refresh_token: "...",
     id_token: "eyJ...", token_type: "Bearer", expires_in: 3600 }
```

### 4.4 JWKS Endpoint

```json
// GET /jwks
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "brain-key-2026-04",
      "use": "sig",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

FB-OS caches this response with a 1-hour TTL and refreshes on JWT validation failure (key rotation detection).

---

## 5. MCP Server Registry Architecture

### 5.1 Registry Service

```go
// MCPRegistry manages the lifecycle of MCP servers and their tools
type MCPRegistry struct {
    store   *store.MCPServerStore
    clients map[string]MCPServerClient  // slug -> live client
    mu      sync.RWMutex
}

// Core Operations
func (r *MCPRegistry) RegisterServer(ctx context.Context, server MCPServer) error
func (r *MCPRegistry) ListServers(ctx context.Context) ([]MCPServer, error)
func (r *MCPRegistry) GetServer(ctx context.Context, slug string) (*MCPServer, error)
func (r *MCPRegistry) ListTools(ctx context.Context, serverSlug string) ([]MCPTool, error)
func (r *MCPRegistry) CallTool(ctx context.Context, serverSlug, toolSlug string, input json.RawMessage, creds Credentials) (*ToolResult, error)
func (r *MCPRegistry) HealthCheck(ctx context.Context, serverSlug string) HealthStatus
```

### 5.2 MCP Server Inventory (MVP)

| Server | Slug | Auth | Tools | Triggers | Legacy File |
|--------|------|------|-------|----------|-------------|
| GableERP | `gable` | API Key | get_products_by_category, bulk_calculate_price, create_quote, accept_and_convert_quote | price_updated, order_status_changed | `registry/gable.go` |
| LocalBlue | `localblue` | Host Header | push_rfq, update_bid_status | bid_submitted | `registry/localblue.go` |
| XUI Projects | `xui` | API Key | create_feed_card, assign_contact_to_phase | card_action_approved, card_action_rejected, card_action_custom | `registry/xui.go` |
| QuickBooks | `quickbooks` | OAuth2 | create_purchase_order, create_invoice, create_bill, get_company_info | payment_received, invoice_updated | `registry/quickbooks.go` |
| 1Build | `onebuild` | Auth Header | get_estimate, search_cost_data, get_line_items | — | `registry/onebuild.go` |
| Gmail | `gmail` | OAuth2 | send_email, list_emails, get_email | email_received | `registry/gmail.go` |
| Outlook | `outlook` | OAuth2 | send_email, list_emails, get_email | email_received | `registry/outlook.go` |

### 5.3 Tool Call Flow

```
Maestro -> MCPRegistry.CallTool("gable", "get_products_by_category", input)
  1. Registry looks up server by slug
  2. Validates input against tool's JSON Schema (input_schema)
  3. Retrieves encrypted credentials for user's org from integration_credentials
  4. Constructs HTTP request using tool's path_template and http_method
  5. Executes request against server's base_url with auth headers
  6. Logs IntegrationEvent (source=fb-brain, target=gable, action=get_products)
  7. Returns ToolResult to Maestro
```

---

## 6. Maestro Orchestrator Architecture

### 6.1 Three-Tier Intent Classification

```go
// Orchestrator processes natural language and executes cross-system workflows
type Orchestrator struct {
    classifier  *Classifier
    registry    *MCPRegistry
    a2a         *A2AClient
    store       *store.Repository
}

// Classify determines intent from user message
func (o *Orchestrator) Process(ctx context.Context, session *Session, message string) (*Response, error) {
    // Tier 1: Regex fast-path (deterministic, <1ms)
    intent, confidence := o.classifier.RegexMatch(message)
    if confidence > 0.95 {
        return o.executeFlow(ctx, session, intent)
    }

    // Tier 2: Claude tool-use (probabilistic, <2s)
    intent, confidence = o.classifier.ClaudeClassify(ctx, message, session.History)
    if confidence > 0.70 {
        return o.executeFlow(ctx, session, intent)
    }

    // Tier 3: Human confirmation (low confidence)
    return o.requestConfirmation(ctx, session, intent, confidence)
}
```

### 6.2 Intent Definitions (Top-10 MVP)

| Intent | Regex Pattern | MCP Servers | Flow |
|--------|--------------|-------------|------|
| `materials_procurement` | `(quote\|order\|price).*(material\|lumber\|roofing)` | GableERP, XUI | MaterialsFlow |
| `labor_bidding` | `(find\|hire\|bid).*(roofer\|plumber\|electrician\|sub)` | LocalBlue, XUI | LaborFlow |
| `product_search` | `(list\|show\|search).*(product\|inventory)` | GableERP | Direct tool call |
| `bid_comparison` | `(compare\|analyze).*(bid\|quote)` | LocalBlue, 1Build | BidComparison |
| `cost_estimation` | `(estimate\|cost\|price).*(project\|phase\|task)` | 1Build | Direct tool call |
| `invoice_check` | `(invoice\|bill\|payment).*(status\|check)` | QuickBooks | Direct tool call |
| `send_email` | `(send\|email\|message).*(vendor\|sub\|client)` | Gmail, Outlook | Direct tool call |
| `po_creation` | `(create\|generate).*(po\|purchase order)` | QuickBooks | Direct tool call |
| `project_status` | `(status\|update\|progress).*(project\|site)` | XUI | Direct tool call |
| `permit_check` | `(permit\|inspection).*(status\|schedule)` | XUI | Direct tool call |

### 6.3 MaterialsFlow (Evolved Architecture)

```go
// MaterialsFlow coordinates multi-system materials procurement
type MaterialsFlow struct {
    registry *MCPRegistry
    a2a      *A2AClient
    store    *store.Repository
}

func (f *MaterialsFlow) Start(ctx context.Context, orgID, projectID string, scope MaterialScope) (*FlowResult, error) {
    // 1. Resolve cross-system identity
    link, _ := f.store.GetAccountLink(ctx, orgID)

    // 2. Get products via MCP
    products, _ := f.registry.CallTool(ctx, "gable", "get_products_by_category",
        json.Marshal(map[string]string{"category": scope.Category}), link.GableCredentials())

    // 3. Calculate pricing via MCP
    pricing, _ := f.registry.CallTool(ctx, "gable", "bulk_calculate_price",
        json.Marshal(BulkPriceRequest{CustomerID: link.GableCustomerID, Items: products}), link.GableCredentials())

    // 4. Create quote via MCP
    quote, _ := f.registry.CallTool(ctx, "gable", "create_quote",
        json.Marshal(QuoteRequest{CustomerID: link.GableCustomerID, Lines: pricing}), link.GableCredentials())

    // 5. Persist RFQ state
    rfq, _ := f.store.CreateRFQ(ctx, RFQ{OrgID: orgID, FlowType: "materials", QuoteData: quote})

    // 6. Emit A2A webhook to FB-OS (currency_code explicit per Composite Currency Pattern)
    f.a2a.Emit(ctx, WebhookEvent{
        Type:    "review_material_quote",
        Payload: ReviewPayload{RFQID: rfq.ID, QuoteData: quote, CurrencyCode: rfq.CurrencyCode},
    })

    // 7. Log integration event
    f.store.LogEvent(ctx, IntegrationEvent{EventType: "materials_flow_started", Source: "fb-brain", Target: "gable"})

    return &FlowResult{RFQID: rfq.ID, Status: "quote_created"}, nil
}
```

---

## 7. A2A Webhook Architecture

### 7.1 Client

```go
// A2AClient emits JWS-signed webhooks to FB-OS
type A2AClient struct {
    targetURL  string            // FB-OS webhook receiver URL
    privateKey *rsa.PrivateKey   // Brain's signing key
    httpClient *http.Client
}

func (c *A2AClient) Emit(ctx context.Context, event WebhookEvent) error {
    payload, _ := json.Marshal(event)

    // Sign with JWS RS256 (detached signature)
    signer, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: c.privateKey},
        &jose.SignerOptions{ExtraHeaders: map[jose.HeaderKey]any{
            "kid": "brain-key-2026-04",
        }})
    jws, _ := signer.Sign(payload)
    signature := jws.DetachedCompactSerialize()

    req, _ := http.NewRequestWithContext(ctx, "POST", c.targetURL, bytes.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-JWS-Signature", signature)
    req.Header.Set("X-Idempotency-Key", event.IdempotencyKey.String())

    resp, err := c.httpClient.Do(req)
    // Log delivery to a2a_webhook_log
    return err
}
```

### 7.2 Webhook Event Schema

```go
type WebhookEvent struct {
    Type           string          `json:"event_type"`
    Payload        json.RawMessage `json:"payload"`
    TraceID        string          `json:"trace_id"`
    IdempotencyKey uuid.UUID       `json:"idempotency_key"`
    Timestamp      time.Time       `json:"timestamp"`
    Issuer         string          `json:"iss"`  // "fb-brain"
}

// Event Types and Payloads (Composite Currency Pattern — currency_code explicit):
// review_material_quote -> { rfq_id, line_items, total_cents, currency_code, vendor }
// review_labor_bid      -> { rfq_id, bidder, amount_cents, currency_code, timeline, ai_analysis }
// update_schedule       -> { event_type, delivery_date, constraints }
// delivery_confirmation -> { materials_ordered, labor_approved, convergence_status }
// create_feed_card      -> { card_type, title, body, actions, priority }
```

### 7.3 Retry Strategy

```
Attempt 1: Immediate
Attempt 2: 30 seconds
Attempt 3: 1 minute
Attempt 4: 2 minutes
Attempt 5: 5 minutes
Attempt 6: 15 minutes
Attempt 7: 1 hour
After 7 failures: Move to a2a_webhook_log with status=dead_letter + alert
```

---

## 8. API Route Contracts

### 8.1 OIDC Endpoints (Public)

```
GET  /.well-known/openid-configuration
GET  /authorize
POST /token
GET  /userinfo  (Bearer token required)
GET  /jwks
POST /revoke
```

### 8.2 Hub Admin API (Session auth)

```
# Registry
GET  /api/hub/registry/servers                    # List all MCP servers
GET  /api/hub/registry/servers/{slug}             # Get server details + tools
POST /api/hub/registry/servers                    # Register new MCP server
PUT  /api/hub/registry/servers/{slug}             # Update server config
DELETE /api/hub/registry/servers/{slug}           # Remove server

# OIDC Client Management
GET  /api/hub/oidc/clients                        # List OIDC clients
POST /api/hub/oidc/clients                        # Register new OIDC client
PUT  /api/hub/oidc/clients/{id}                   # Update client
DELETE /api/hub/oidc/clients/{id}                 # Remove client

# Integration Connections
GET  /api/hub/connections                          # List org's connected integrations
POST /api/hub/connections/{serverSlug}/connect      # Connect integration (API key or OAuth2)
DELETE /api/hub/connections/{serverSlug}/disconnect  # Disconnect integration

# OAuth2 Flows (external systems)
GET  /api/hub/oauth/{provider}/authorize           # Start OAuth2 flow (QuickBooks, Gmail, etc.)
GET  /api/hub/oauth/{provider}/callback            # OAuth2 callback handler

# Maestro Chat
POST /api/hub/chat/message                         # Send message to Maestro
GET  /api/hub/chat/sessions                        # List chat sessions
GET  /api/hub/chat/sessions/{id}/messages           # Get session messages

# Real-time Events
GET  /api/hub/events/stream                        # SSE event stream

# Dashboard
GET  /api/hub/dashboard                            # Home dashboard data
```

### 8.3 MCP Protocol Endpoints

```
POST /mcp/tools/list                               # List available tools (per user's connected servers)
POST /mcp/tools/call                               # Execute a tool call
     Body: { server_slug, tool_slug, input: {} }
POST /mcp/triggers/subscribe                       # Subscribe to trigger events
```

---

## 9. Lit Admin UI Architecture

```
fb-home-page (Split Layout)
├── Left Panel: fb-activity-ticker (recent integration events)
└── Right Panel: fb-maestro-chat (AI co-pilot)

fb-ecosystem-page
└── fb-ecosystem-canvas (D3 force-directed graph)
    ├── fb-platform-node (Brain, FB-OS, GableERP, etc.)
    └── fb-connection-edge (data flow lines)

fb-admin-registry
└── fb-mcp-registry (table of servers)
    ├── fb-mcp-tool-card (tool details + test panel)
    └── fb-oidc-client-row (OIDC client management)

fb-marketplace-page
└── Available MCP servers for connection

fb-settings-page
├── fb-oidc-manager (OIDC client CRUD)
├── Integration credentials management
└── Organization settings

fb-login-page
├── Magic link email input
└── SSO redirect handling
```

### Maestro Drawer (Global FAB)

```
Floating Action Button (bottom-right, all pages)
└── fb-maestro-drawer (slide-out panel)
    └── fb-maestro-chat
        ├── fb-chat-bubble (user/assistant messages)
        └── fb-execution-row (MCP tool call results)
```

---

## 10. CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  backend:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: brain_test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }

      # SQL Migration Linter (HARD FAIL — no exemptions)
      # Checks: (1) No DECIMAL/NUMERIC/FLOAT/MONEY on monetary columns
      #          (2) No orphan amount_cents without currency_code
      - name: Composite Currency Pattern Enforcement
        run: ./scripts/lint-migrations.sh

      # OIDC conformance tests
      - name: OIDC Tests
        run: go test ./internal/oidc/...

      # MCP registry tests
      - name: MCP Tests
        run: go test ./internal/mcp/...

      # Full test suite
      - name: Test
        run: go test ./...

      # Lint
      - name: Lint
        run: golangci-lint run ./...

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: cd frontend && npm ci && npm run lint && npm run build
      - name: Lighthouse CI
        run: npx lhci autorun  # Performance >90, Accessibility >95, Bundle <50KB
```

---

## 11. Cross-System Contract Summary

### Brain Provides to OS

| Interface | Protocol | Endpoint | SLA |
|-----------|----------|----------|-----|
| Identity | OIDC | `/.well-known/openid-configuration` | 99.99% uptime |
| JWT Signing | JWKS | `/jwks` | <10ms validation (cached) |
| Token Issuance | OIDC | `/token` | <500ms p99 |
| Webhooks | A2A | Brain -> OS `/api/v1/a2a/webhook` | <500ms delivery, >99.9% rate |

### Brain Depends on OS

| Interface | Protocol | Endpoint | Purpose |
|-----------|----------|----------|---------|
| Webhook Receiver | A2A | OS `/api/v1/a2a/webhook` | Deliver integration events |
| Approval Callbacks | REST | OS -> Brain callback | Quote/bid approval events |

### Brain Depends on External Systems

| System | Protocol | Auth | Latency Budget |
|--------|----------|------|---------------|
| GableERP | REST | API Key | <3s per tool call |
| LocalBlue | REST | Host Header | <3s per tool call |
| QuickBooks | REST + CloudEvents | OAuth2 | <3s per tool call |
| 1Build | REST | Auth Header | <3s per tool call |
| Gmail | REST | OAuth2 | <3s per tool call |
| Outlook | REST | OAuth2 | <3s per tool call |
| Anthropic Claude | REST | API Key | <2s for classification |

---

## 12. Observability Stack

| Layer | Tool | Data |
|-------|------|------|
| API Metrics | Prometheus (promhttp on Chi) | Request latency, error rate, per-pillar histograms |
| Distributed Tracing | OpenTelemetry | Brain -> external API calls, cross-system spans |
| Structured Logging | slog (JSON) | Correlation IDs, event types, tool call details |
| Dashboards | Grafana | Per-pillar panels: OIDC, MCP, Maestro, A2A, Admin UI |
| Error Tracking | Sentry | Go panic handler, frontend error boundary |
| AI Cost Tracking | Custom Prometheus counter | Cost per intent, token usage, tier distribution |

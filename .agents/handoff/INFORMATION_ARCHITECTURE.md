# Information Architecture

**Document ID:** AG-05-IA
**System:** FutureBuild Brain (System of Connection)
**Created:** 2026-04-02
**Pipeline Stage:** 05 - Design System
**Status:** COMPLETE

---

## 1. Workspace Model

FutureBuild Brain organizes its user experience into **three surfaces**, each serving a distinct interaction paradigm within the Connection Plane.

```
┌────────────────────────────────────────────────────────────────────────────┐
│                         FutureBuild Brain                                  │
├──────────────────┬─────────────────────────┬───────────────────────────────┤
│  HUB ADMIN       │  MAESTRO AI CO-PILOT    │  INTEGRATION REGISTRY        │
│  Dashboard       │  (Chat Interface)       │  (MCP + OIDC Management)     │
│                  │                          │                              │
│  Integration     │  General Contractor      │  Integration Admin           │
│  Admin           │  Bookkeeper              │  DevOps                      │
│                  │  All Authenticated       │                              │
│                  │  Users                    │                              │
│                  │                          │                              │
│  Web Desktop     │  Web Desktop + Drawer    │  Web Desktop                 │
├──────────────────┼─────────────────────────┼───────────────────────────────┤
│  Ecosystem Map   │  Natural Language       │  MCP Server Browser          │
│  Platform Health │  Materials Flow         │  Tool Schema Viewer          │
│  Connection Mgmt │  Labor Flow             │  OIDC Client Manager         │
│  Activity Feed   │  QuickBooks Automation  │  Webhook Delivery Logs       │
│  Setup Wizard    │  Suggested Prompts      │  A2A Agent Cards             │
└──────────────────┴─────────────────────────┴───────────────────────────────┘
```

### Surface Definitions

| Surface | Primary Users | Access | Purpose |
|---------|--------------|--------|---------|
| **Hub Admin Dashboard** | Integration Admin, Owner | Web (desktop ≥1200px) | Ecosystem visualization, platform health monitoring, connection management, onboarding |
| **Maestro AI Co-pilot** | GC, Bookkeeper, all authenticated users | Web (desktop full-panel + FAB drawer) | Natural language orchestration of materials procurement, labor bidding, QuickBooks automation |
| **Integration Registry** | Integration Admin, DevOps | Web (desktop, admin-only) | MCP server management, OIDC client CRUD, webhook logs, A2A agent card configuration |

---

## 2. Navigation Model

### 2.1 Primary Navigation (Top Bar)

Brain uses a **top navigation bar** (not sidebar) because its surfaces are fewer and flatter than FB-OS.

```
┌───────────────────────────────────────────────────────────────────────────┐
│ ⬡ FutureBuild Brain  │  Dashboard  │  Ecosystem  │  Admin  │  🔔  👤    │
└───────────────────────────────────────────────────────────────────────────┘
                          │              │             │
                          ▼              ▼             ▼
                     Home Page      Ecosystem     Registry +
                     (Dashboard     Canvas        OIDC Manager
                      + Maestro)
```

### 2.2 Maestro Access Points

Maestro is available everywhere via two mechanisms:

| Access Point | Trigger | Layout |
|-------------|---------|--------|
| **Home Page (inline)** | Default on `/` | Right 45% of split layout |
| **FAB Drawer** | Floating action button (bottom-right) | 420×600px slide-out panel |
| **Mobile** | Full-screen overlay | Replaces current view |

### 2.3 Admin Sub-Navigation

The Admin surface uses tabs within the page:

```
┌─────────────────────────────────────────────────┐
│  Admin                                           │
│  [MCP Registry] [OIDC Clients] [Webhooks]       │
│  ─────────────────────────────────────────       │
│  ... content based on selected tab ...           │
└─────────────────────────────────────────────────┘
```

### 2.4 Contextual Navigation

- **Ecosystem Canvas:** Click platform node → slide-over detail panel
- **MCP Registry:** Click server row → expand to show tools
- **Setup Wizard:** Linear stepper (1→2→3→4→5)
- **Maestro Chat:** Conversation history in drawer header

---

## 3. Screen Inventory

### 3.1 Hub Admin Dashboard

| Screen | Route | Components | Data Source |
|--------|-------|------------|------------|
| **Home (Split View)** | `/` | `<fb-home-page>`: Dashboard grid (left 55%) + Maestro chat (right 45%) | `GET /api/hub/dashboard`, SSE `/api/hub/events/stream` |
| **Ecosystem Canvas** | `/ecosystem` | `<fb-ecosystem-canvas>`: XY Flow graph with platform nodes, connection edges, minimap | `GET /api/hub/platforms`, `GET /api/hub/connections` |
| **Platform Detail** | `/ecosystem` (slide-over) | `<fb-platform-detail>`: Health, connectivity test, API URL, credentials | `GET /api/hub/platforms/:id` |
| **Marketplace** | `/marketplace` | `<fb-marketplace-page>`: Template catalog grid | `GET /api/hub/templates` |
| **Setup Wizard** | `/setup` | `<fb-setup-wizard>`: 5-step onboarding flow | Multi-step API calls |
| **Org Settings** | `/settings/org` | `<fb-settings-page>`: Member management, billing | `GET /api/hub/orgs/current` |

### 3.2 Maestro AI Co-pilot

| Screen | Route | Components | Data Source |
|--------|-------|------------|------------|
| **Chat (Home)** | `/` (right panel) | `<fb-maestro-chat>`: Message list + input + suggested prompts | `POST /api/hub/chat/sessions/:id/messages` (SSE) |
| **Chat (Drawer)** | Any page (FAB) | `<fb-maestro-drawer>`: Compact chat in 420×600px overlay | Same SSE endpoint |
| **Chat History** | Drawer header dropdown | Session list with timestamps | `GET /api/hub/chat/sessions` |

### 3.3 Integration Registry

| Screen | Route | Components | Data Source |
|--------|-------|------------|------------|
| **MCP Server Browser** | `/admin/registry` | `<fb-mcp-registry>`: Server list with tool count, health, actions | `GET /api/hub/registry/systems` |
| **Tool Detail** | `/admin/registry` (expanded row) | `<fb-mcp-tool-card>`: JSON schema, trigger/action badge, test runner | `GET /api/hub/registry/systems/:id/tools` |
| **OIDC Client Manager** | `/admin/oidc` | `<fb-oidc-manager>`: Client CRUD, redirect URIs, grant types | `GET /api/hub/oidc/clients` |
| **Webhook Logs** | `/admin/webhooks` | `<fb-webhook-log>`: Delivery history, retry status, payload preview | `GET /api/hub/a2a/webhooks` |
| **A2A Agent Cards** | `/admin/agents` | `<fb-agent-card-list>`: Published agent cards, skill inventory | `GET /api/hub/a2a/agent-cards` |

### 3.4 Auth Screens

| Screen | Route | Components |
|--------|-------|------------|
| **Login** | `/login` | `<fb-login-page>`: Standard login + magic link + SSO |
| **Magic Link Verify** | `/auth/verify` | `<fb-verify-page>`: Token verification |
| **OAuth Callback** | `/auth/callback` | `<fb-oauth-callback>`: Gmail/Outlook OAuth handler |
| **OIDC Authorize** | `/authorize` | `<fb-oidc-authorize>`: Consent screen for downstream clients |

---

## 4. Content Inventory

### 4.1 Data Objects

| Object | Source | Display Context | Typography |
|--------|--------|----------------|-----------|
| Platform | PostgreSQL | Ecosystem nodes, detail panels | Name: Outfit; URL: JetBrains Mono |
| Connection | PostgreSQL | Ecosystem edges, connection list | ID: JetBrains Mono; Label: Outfit |
| Integration | PostgreSQL | Integration cards, monitor panels | Name: Outfit; Timestamps/counts: JetBrains Mono |
| MCP Server | PostgreSQL | Registry browser rows | Name: Outfit; Tool schemas: JetBrains Mono |
| MCP Tool | PostgreSQL | Tool cards, schema viewer | Name + schema: JetBrains Mono; Description: Outfit |
| OIDC Client | PostgreSQL | Client manager rows | Client ID + URIs: JetBrains Mono; Name: Outfit |
| Chat Session | PostgreSQL | Maestro message list | Messages: Outfit; Code/data: JetBrains Mono |
| Webhook Event | PostgreSQL | Webhook log table | URL + status + payload: JetBrains Mono; Action: Outfit |
| Agent Card | PostgreSQL | A2A agent card list | Skills: JetBrains Mono; Name: Outfit |
| Integration Event | PostgreSQL | Activity feed | Timestamp + IDs: JetBrains Mono; Description: Outfit |

### 4.2 Actions

| Action | Trigger | Surface | Permission |
|--------|---------|---------|-----------|
| Test platform connectivity | Button on platform detail | Hub Admin | Admin |
| Create connection | Form in ecosystem canvas | Hub Admin | Admin |
| Send chat message | Input bar enter/send | Maestro | Authenticated |
| Approve materials quote | Chat inline action card | Maestro | Owner, Admin |
| Register MCP server | Form in registry | Integration Registry | Admin |
| Create OIDC client | Form in client manager | Integration Registry | Admin |
| Activate integration template | Setup wizard step 5 | Hub Admin | Admin |
| Run integration test | Button on integration detail | Integration Registry | Admin |

### 4.3 System Messages

| Message Type | Display | Context |
|-------------|---------|---------|
| Integration health change | Activity ticker + toast | Ecosystem canvas |
| Webhook delivery failure | Toast (error) + log entry | Admin webhook log |
| Chat response streaming | Inline SSE rendering | Maestro |
| OIDC token issued | Activity log entry | Silent (admin-visible) |
| MCP tool execution | Activity ticker + execution log | Ecosystem + Admin |

---

## 5. URL Structure

### 5.1 Application Routes

```
/                                    → Home: Dashboard (left) + Maestro (right)
/login                               → Login page (standard + magic link + SSO)
/auth/verify                         → Magic link verification
/auth/callback                       → OAuth callback (Gmail/Outlook)

# Hub Admin
/ecosystem                           → Ecosystem canvas (XY Flow graph)
/marketplace                         → Integration template catalog
/settings/org                        → Organization settings

# Integration Registry (Admin)
/admin                               → Redirect to /admin/registry
/admin/registry                      → MCP server browser
/admin/oidc                          → OIDC client manager
/admin/webhooks                      → Webhook delivery logs
/admin/agents                        → A2A agent cards

# Setup
/setup                               → Onboarding wizard
/setup/configure/:sessionId          → AI-driven configuration chat
```

### 5.2 OIDC Provider Endpoints (Served by Brain)

```
/.well-known/openid-configuration    → OIDC discovery document
/authorize                           → Authorization endpoint (consent screen)
/token                               → Token endpoint (code → tokens)
/userinfo                            → UserInfo endpoint
/jwks                                → JSON Web Key Set
/revoke                              → Token revocation
```

### 5.3 API Routes

```
# Auth
POST   /api/hub/auth/login
POST   /api/hub/auth/logout
POST   /api/hub/auth/magic-link/request
POST   /api/hub/auth/magic-link/verify
GET    /api/hub/auth/me

# Platforms & Connections
GET    /api/hub/platforms
POST   /api/hub/platforms
PUT    /api/hub/platforms/:id
POST   /api/hub/platforms/:id/test
GET    /api/hub/connections
POST   /api/hub/connections/direct

# Integrations
GET    /api/hub/integrations
POST   /api/hub/integrations
PUT    /api/hub/integrations/:id

# Chat (Maestro)
POST   /api/hub/chat/sessions
POST   /api/hub/chat/sessions/:id/messages   (SSE stream)
GET    /api/hub/chat/sessions/:id

# Registry
GET    /api/hub/registry/systems
POST   /api/hub/registry/systems
GET    /api/hub/registry/systems/:id/tools
POST   /api/hub/registry/systems/:id/tools/:toolId/test

# OIDC Admin
GET    /api/hub/oidc/clients
POST   /api/hub/oidc/clients
PUT    /api/hub/oidc/clients/:id
DELETE /api/hub/oidc/clients/:id

# A2A
POST   /api/hub/a2a/webhooks/dispatch
GET    /api/hub/a2a/webhooks
GET    /api/hub/a2a/agent-cards

# Dashboard & Activity
GET    /api/hub/dashboard
GET    /api/hub/activity
GET    /api/hub/events/stream             (SSE)
GET    /api/hub/notifications
```

---

## 6. Taxonomy & Naming Conventions

### 6.1 User-Facing Vocabulary

| Internal Term | User-Facing Term | Context |
|--------------|-----------------|---------|
| MCP Server | Integration Server | Registry browser |
| MCP Tool | Integration Tool | Tool cards, schema viewer |
| OIDC Client | Connected Application | Client manager |
| A2A Webhook | Event Notification | Webhook logs |
| Agent Card | Agent Profile | A2A agent list |
| IntegrationEvent | Activity | Activity feed |
| ChatSession | Conversation | Maestro chat |
| Platform | Connected Platform | Ecosystem canvas |
| Connection | Data Link | Ecosystem canvas |

### 6.2 Component Naming Convention

- **Tag prefix:** `fb-` (shared with FB-OS for ecosystem consistency)
- **File naming:** PascalCase matching class (e.g., `FBMaestroChat.ts`)
- **Event naming:** `fb-{noun}-{verb}` (e.g., `fb-platform-tested`, `fb-tool-executed`)
- **React component naming:** PascalCase (e.g., `PlatformNode.tsx`) — during migration period

### 6.3 Status Vocabulary

| Status | Badge Color | Usage |
|--------|------------|-------|
| Healthy | Gable Green | Platform connected, integration active |
| Degraded | Amber Warning | High error rate, slow responses |
| Down | Safety Red | Platform unreachable, integration failed |
| Pending | Blueprint Blue | Connection being established, webhook in retry |
| Inactive | Gray (#5A5B66) | Paused integration, disabled OIDC client |

---

## 7. Information Hierarchy

### 7.1 Progressive Disclosure Pattern

```
Level 0: Surface Selection (Top Nav)
  └─ Dashboard | Ecosystem | Admin

Level 1: Overview (aggregated status)
  └─ Platform health grid, integration counts, activity stream

Level 2: Entity List (all items of a type)
  └─ MCP servers, OIDC clients, webhook events

Level 3: Entity Detail (single item, slide-over or expanded)
  └─ Platform health, tool schemas, execution history

Level 4: Action (modify state)
  └─ Test connectivity, register server, create client
```

### 7.2 Data Density by Surface

| Surface | Density | Rationale |
|---------|---------|-----------|
| Hub Admin | **Medium** | Ecosystem canvas is visual (low density), but activity feed and dashboard grid are data-rich. |
| Maestro | **Low** | Chat is conversational. Rich artifacts appear inline but the primary interaction is natural language. |
| Integration Registry | **High** | Admin needs JSON schemas, URIs, execution logs, webhook payloads. JetBrains Mono-heavy. |

---

## 8. Maestro Interaction Model

### 8.1 Chat Flow

```
┌─────────────────────────────────────────────┐
│  Maestro AI Co-pilot                         │
├─────────────────────────────────────────────┤
│                                              │
│  [Suggested Prompt Cards]                    │
│   "List products in GableERP"                │
│   "Order roofing materials for Riverside"    │
│   "Check QuickBooks PO status"               │
│                                              │
│  ─────────────────────────────────          │
│                                              │
│  [User]: Order roofing materials for the     │
│          Riverside project                    │
│                                              │
│  [Maestro]: I'll help you order roofing      │
│  materials. Let me check GableERP...         │
│                                              │
│  ┌─────────────────────────────────────┐    │
│  │ 🛒 Materials Quote                   │    │
│  │ Vendor: ABC Roofing Supply           │    │
│  │ Items: 45 squares GAF Timberline     │    │
│  │ Total: $12,450.00          (mono)    │    │
│  │ Delivery: Apr 15, 2026     (mono)    │    │
│  │                                       │    │
│  │ [Approve & Create PO]  [Edit]  [✕]   │    │
│  └─────────────────────────────────────┘    │
│                                              │
│  ┌───────────────────────────────┐          │
│  │  Type a message...        [⏎] │          │
│  └───────────────────────────────┘          │
└─────────────────────────────────────────────┘
```

### 8.2 Intent → MCP Tool Mapping

| User Intent | Classified As | MCP Server | MCP Tool |
|------------|--------------|------------|----------|
| "List products" | `product_query` | GableERP | `get_products_by_category` |
| "Order materials" | `materials_procurement` | GableERP + XUI | `create_quote` → `create_feed_card` |
| "Send RFQ to subs" | `labor_bidding` | LocalBlue | `push_rfq` |
| "Create PO in QuickBooks" | `accounting_entry` | QuickBooks | `create_purchase_order` |
| "Check bid status" | `bid_status` | LocalBlue | `get_bid_status` |

---

## 9. Cross-System Navigation

### 9.1 Brain → OS Handoffs

| Trigger | Brain Action | OS Action |
|---------|-------------|-----------|
| Materials quote approved | A2A webhook with order details | Creates feed card in Command Center |
| PO created in QuickBooks | A2A webhook with PO data | Updates project budget |
| Schedule-affecting event | A2A webhook with event type | Triggers CPM recalculation |

### 9.2 OS → Brain Handoffs

| Trigger | OS Action | Brain Response |
|---------|-----------|---------------|
| User opens Brain link | JWT validated by Brain OIDC | Brain serves Hub Admin |
| Integration health query | Settings page calls Brain API | Returns platform status |

Brain and OS are separate web applications. Users switch between them via direct URL navigation. The shared JWT (issued by Brain, validated by OS) provides seamless authentication.

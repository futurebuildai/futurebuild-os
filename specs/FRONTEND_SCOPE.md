# FutureBuild Frontend Scope
## Chat + Dashboard UX Architecture

**Version:** 1.0  
**Date:** December 30, 2026  
**Status:** Planning

---

## 1. Executive Summary

FutureBuild's frontend embodies the principle: **"Simple Interface. Powerful Engine."**

The UI is built around a **Chat + Dashboard** paradigm where users interact primarily through conversation, with visual artifacts (schedules, charts, reports) generated on-demand and displayed inline. The same core experience serves all user types (builders, clients, subcontractors) with role-based permissions.

### Design Philosophy

| Principle | Implementation |
|-----------|----------------|
| **Chat-First** | Primary interaction is conversational; no complex nav menus. |
| **Agent-as-Sieve** | Agent proactively surfaces priorities; User doesn't "hunt" for data. |
| **Unified** | Same interface for all user roles (Builder, Sub, Client). |
| **Responsive** | Mobile-first Web App design; touch-optimized. |
| **Fast** | Lit + Vite, optimized reactivity. |
| Type-Safe | Strict TypeScript 5.0+, shared backend types |

### Technology Stack

| Layer | Technology |
|-------|------------|
| Language | TypeScript 5.0+ (Strict Mode) |
| Components | Lit 3.0 (Shadow DOM, Reactive Properties) |
| Styling | CSS Custom Properties + Scoped Component Styles |
| State | Signals-based Store (@lit-labs/preact-signals) |
| Routing | Custom SPA router (History API) |
| Charts | D3.js for complex visualizations |
| Gantt | Custom Canvas/SVG timeline component |
| Build | Vite (configured for ES Modules output) |
| Icons | Custom SVG icon system |
| Backend API | Go REST API (Chi Router) |

---

## 2. User Types & Access

### 2.1 Access Matrix

| User Type | Access Method | Primary View | Key Actions |
|-----------|---------------|--------------|-------------|
| Builder/Project Lead | Email magic link | Multi-project dashboard | Full project control, chat, overrides |
| Admin | Email magic link | Org settings + all projects | User management, settings |
| Client (Homeowner) | Email magic link | Single project portal | View progress, ask questions, approve |
| Subcontractor | Email magic link | Task-focused portal | View tasks, report progress, upload photos |
| Vendor | Email magic link | Procurement portal | View orders, confirm deliveries |

### 2.2 Unified Chat Experience

All users interact through the same chat interface. The system adapts responses based on role:

**Builder asks:** "What's the critical path?"
→ Full technical response with Gantt artifact

**Client asks:** "When will my house be done?"
→ Friendly summary with milestone timeline

**Sub asks:** "What do I need to do this week?"
→ Task list with completion buttons

---

## 3. Application Architecture

### 3.1 Lit Integration

FutureBuild uses **Lit 3.0** to provide a structured, high-performance component architecture. All UI elements extend `LitElement`, leveraging:

- **Shadow DOM:** Scoped CSS and encapsulated DOM structures.
- **Reactive Properties:** Automatic re-rendering when element properties or state (Signals) change.
- **Declarative Templates:** Clean, readable HTML templates using the `html` tag literal.
- **Lifecycle Management:** Standardized hooks for component mounting, updates, and cleanup.

### 3.2 Type Sharing & Data Integrity

To ensure end-to-end data integrity, the frontend MUST maintain a `types/` directory. All TypeScript interfaces and enums in this directory MUST be generated or manually aligned to match the definitions in [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md).

- **Automated Alignment:** Shared types for `Task`, `Phase`, `Invoice`, `Budget`, and `Contact`.
- **Enum Synchronization:** `TaskStatus`, `UserRole`, and `ArtifactType` must match the Go backend exactly.
- **Strict Typing:** All components and services use these shared interfaces to eliminate "any" types and reduce integration bugs.

### 3.3 High-Level Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  <fb-header>  Logo | Project Selector | Notifications | User    │
├─────────────┬───────────────────────────────────────────────────┤
│             │                                                   │
│  <fb-nav>   │  <fb-main>                                        │
│             │  ┌─────────────────────────────────────────────┐  │
│  Dashboard  │  │  <fb-dashboard>                             │  │
│  Schedule   │  │  Status cards, metrics, quick actions       │  │
│  Tasks      │  └─────────────────────────────────────────────┘  │
│  Documents  │  ┌─────────────────────────────────────────────┐  │
│  Budget     │  │  <fb-chat>                                  │  │
│  Team       │  │  Conversational interface with artifacts    │  │
│  Settings   │  │                                             │  │
│             │  └─────────────────────────────────────────────┘  │
│             │                                                   │
└─────────────┴───────────────────────────────────────────────────┘
```

### 3.2 Responsive Behavior

| Breakpoint | Layout |
|------------|--------|
| Desktop (≥1200px) | Sidebar + Dashboard + Chat side-by-side |
| Tablet (768-1199px) | Collapsible sidebar, stacked dashboard/chat |
| Mobile (<768px) | Bottom nav, full-screen views, swipe between dashboard/chat |

### 3.3 Component Tree

```
<fb-app>
├── <fb-header>
│   ├── <fb-logo>
│   ├── <fb-project-selector>
│   ├── <fb-notification-bell>
│   └── <fb-user-menu>
│
├── <fb-sidebar>
│   ├── <fb-nav-item icon="dashboard">
│   ├── <fb-nav-item icon="schedule">
│   ├── <fb-nav-item icon="tasks">
│   ├── <fb-nav-item icon="documents">
│   ├── <fb-nav-item icon="budget">
│   ├── <fb-nav-item icon="team">
│   └── <fb-nav-item icon="settings">
│
├── <fb-main>
│   ├── <fb-dashboard-view>
│   │   ├── <fb-status-card type="progress">
│   │   ├── <fb-status-card type="budget">
│   │   ├── <fb-status-card type="weather">
│   │   ├── <fb-status-card type="inspections">
│   │   ├── <fb-upcoming-tasks>
│   │   └── <fb-recent-activity>
│   │
│   ├── <fb-schedule-view>
│   │   └── <fb-gantt-chart>
│   │
│   ├── <fb-tasks-view>
│   │   ├── <fb-task-filters>
│   │   └── <fb-task-list>
│   │       └── <fb-task-card>
│   │
│   ├── <fb-documents-view>
│   │   ├── <fb-document-upload>
│   │   └── <fb-document-list>
│   │
│   ├── <fb-budget-view>
│   │   ├── <fb-budget-chart>
│   │   └── <fb-cost-table>
│   │
│   └── <fb-team-view>
│       └── <fb-team-member-card>
│
├── <fb-chat-panel>
│   ├── <fb-chat-header>
│   ├── <fb-message-list>
│   │   ├── <fb-message type="user">
│   │   ├── <fb-message type="assistant">
│   │   └── <fb-artifact>
│   │       ├── <fb-artifact-gantt>
│   │       ├── <fb-artifact-table>
│   │       ├── <fb-artifact-chart>
│   │       └── <fb-artifact-card>
│   ├── <fb-chat-input>
│   └── <fb-voice-button>
│
├── <fb-notification-center>
│   └── <fb-notification-item>
│
├── <fb-modal>
│   └── <fb-artifact-fullscreen>
│
└── <fb-toast>
```

---

## 4. Web Component Library

### 4.1 Base Component Class (Lit)

```typescript
import { LitElement, html, css } from 'lit';
import { property, state } from 'lit/decorators.js';

export class FBElement extends LitElement {
  // Common styles shared across all components
  static styles = css`
    :host {
      box-sizing: border-box;
    }
  `;

  // Helper for emitting custom events
  emit(eventName: string, detail: any = {}) {
    this.dispatchEvent(new CustomEvent(eventName, {
      bubbles: true,
      composed: true,
      detail
    }));
  }
}
```

### 4.2 Component Registration

```typescript
import { FBApp } from './app/FBApp';
import { FBHeader } from './layout/FBHeader';
import { FBSidebar } from './layout/FBSidebar';
import { FBChat } from './chat/FBChat';
import { FBArtifactInvoice } from './artifacts/FBArtifactInvoice';
import { FBArtifactBudget } from './artifacts/FBArtifactBudget';
import { FBContactCard } from './contacts/FBContactCard';

const components = {
  'fb-app': FBApp,
  'fb-header': FBHeader,
  'fb-sidebar': FBSidebar,
  'fb-chat': FBChat,
  'fb-artifact-invoice': FBArtifactInvoice,
  'fb-artifact-budget': FBArtifactBudget,
  'fb-contact-card': FBContactCard,
};

export function registerComponents() {
  Object.entries(components).forEach(([name, component]) => {
    if (!customElements.get(name)) {
      customElements.define(name, component);
    }
  });
}
```

### 4.3 New & Refactored Components

#### <fb-artifact-invoice> (Invoice Processor)
Split-screen interface for validating AI-extracted invoice data.
- **Left Panel:** Document Viewer (PDF or Image) with zoom/rotate.
- **Right Panel:** Editable form mirroring the `Invoice` type.
- **Fields:** Vendor (Autocomplete from Rolodex), Date, Amount, WBS Code, Line Items.
- **Action:** "Approve" button to commit data to the Data Spine.

#### <fb-artifact-budget> (Budget Overview)
A detailed 3-column financial table grouped by Project Phase.
- **Columns:** Estimated (from specs), Committed (from contracts/POs), Actual (from approved invoices).
- **Interactions:** Expandable rows to see individual cost actuals.
- **Visuals:** Delta indicators (green/red) for budget variance.

#### <fb-contact-card> (Project Rolodex)
UI for managing project-specific team assignments.
- **Attributes:** `contact-id`, `role`, `assigned-phases[]`.
- **Functionality:** Link a contact to specific WBS phases for automated outreach.
- **Social:** One-tap SMS/Email triggers for the builder.

#### <fb-chat-panel>
The primary conversational container, now utilizing Lit's reactive repeaters for the message list.
- **Reactivity:** Uses `SignalWatch` to update when the global message store changes.
- **Artifact Rendering:** Dynamically maps agent tool outputs to components.

---

## 5. State Management

### 5.1 Event-Driven Store

```javascript
class Store {
  constructor() {
    this.state = {
      user: null,
      currentProject: null,
      projects: [],
      messages: [],
      notifications: [],
      tasks: [],
      isLoading: false
    };
    this.listeners = new Map();
  }

  getState() {
    return { ...this.state };
  }

  setState(updates) {
    const prevState = { ...this.state };
    this.state = { ...this.state, ...updates };
    
    Object.keys(updates).forEach(key => {
      if (this.listeners.has(key)) {
        this.listeners.get(key).forEach(callback => {
          callback(this.state[key], prevState[key]);
        });
      }
    });
  }

  subscribe(key, callback) {
    if (!this.listeners.has(key)) {
      this.listeners.set(key, new Set());
    }
    this.listeners.get(key).add(callback);
    
    return () => {
      this.listeners.get(key).delete(callback);
    };
  }
}

export const store = new Store();
```

### 5.2 API Service

```javascript
const API_BASE = '/api/v1';

class ApiService {
  constructor() {
    this.token = localStorage.getItem('fb_token');
  }

  setToken(token) {
    this.token = token;
    localStorage.setItem('fb_token', token);
  }

  async request(endpoint, options = {}) {
    const headers = {
      'Content-Type': 'application/json',
      ...(this.token && { 'Authorization': `Bearer ${this.token}` }),
      ...options.headers
    };

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }

    return response.json();
  }

  async requestMagicLink(email) {
    return this.request('/auth/magic-link', {
      method: 'POST',
      body: JSON.stringify({ email })
    });
  }

  async getProjects() {
    return this.request('/projects');
  }

  async getProject(id) {
    return this.request(`/projects/${id}`);
  }

  async getSchedule(projectId) {
    return this.request(`/projects/${projectId}/schedule`);
  }

  async sendMessage(projectId, message) {
    return this.request('/chat', {
      method: 'POST',
      body: JSON.stringify({ project_id: projectId, message })
    });
  }

  async *streamChat(projectId, message) {
    const response = await fetch(`${API_BASE}/chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`
      },
      body: JSON.stringify({ project_id: projectId, message })
    });

    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      yield decoder.decode(value);
    }
  }

  async getTasks(projectId) {
    return this.request(`/projects/${projectId}/tasks`);
  }

  async updateTaskProgress(projectId, taskId, progress) {
    return this.request(`/projects/${projectId}/tasks/${taskId}/progress`, {
      method: 'POST',
      body: JSON.stringify(progress)
    });
  }

  async uploadDocument(projectId, file) {
    const formData = new FormData();
    formData.append('file', file);

    return fetch(`${API_BASE}/projects/${projectId}/documents`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`
      },
      body: formData
    }).then(r => r.json());
  }
}

export const api = new ApiService();
```

---

## 6. Routing

### 6.1 SPA Router

```javascript
class Router {
  constructor() {
    this.routes = new Map();
    this.currentRoute = null;
    
    window.addEventListener('popstate', () => this.handleRoute());
  }

  register(path, handler) {
    this.routes.set(path, handler);
  }

  navigate(path) {
    history.pushState(null, '', path);
    this.handleRoute();
  }

  handleRoute() {
    const path = window.location.pathname;
    
    for (const [pattern, handler] of this.routes) {
      const match = this.matchRoute(pattern, path);
      if (match) {
        this.currentRoute = { pattern, params: match.params };
        handler(match.params);
        return;
      }
    }
    
    this.navigate('/dashboard');
  }

  matchRoute(pattern, path) {
    const patternParts = pattern.split('/');
    const pathParts = path.split('/');
    
    if (patternParts.length !== pathParts.length) return null;
    
    const params = {};
    for (let i = 0; i < patternParts.length; i++) {
      if (patternParts[i].startsWith(':')) {
        params[patternParts[i].slice(1)] = pathParts[i];
      } else if (patternParts[i] !== pathParts[i]) {
        return null;
      }
    }
    
    return { params };
  }
}

export const router = new Router();

router.register('/dashboard', () => showView('dashboard'));
router.register('/projects/:id', (params) => showView('project', params.id));
router.register('/projects/:id/schedule', (params) => showView('schedule', params.id));
router.register('/projects/:id/tasks', (params) => showView('tasks', params.id));
router.register('/projects/:id/documents', (params) => showView('documents', params.id));
router.register('/projects/:id/budget', (params) => showView('budget', params.id));
router.register('/projects/:id/team', (params) => showView('team', params.id));
router.register('/settings', () => showView('settings'));
router.register('/portal', () => showView('portal'));
```

---

## 7. Design System

### 7.1 CSS Custom Properties

```css
:root {
  --fb-bg-primary: #000000;
  --fb-bg-secondary: #0a0a0a;
  --fb-bg-tertiary: #1a1a1a;
  --fb-bg-card: #111111;
  
  --fb-text-primary: #ffffff;
  --fb-text-secondary: #aaaaaa;
  --fb-text-muted: #666666;
  
  --fb-primary: #667eea;
  --fb-primary-hover: #5a6fd6;
  --fb-secondary: #764ba2;
  
  --fb-success: #2e7d32;
  --fb-warning: #e65100;
  --fb-error: #c62828;
  --fb-info: #1565c0;
  
  --fb-border: #333333;
  --fb-border-light: #222222;
  
  --fb-font-family: 'Poppins', sans-serif;
  --fb-text-xs: 0.75rem;
  --fb-text-sm: 0.875rem;
  --fb-text-base: 1rem;
  --fb-text-lg: 1.125rem;
  --fb-text-xl: 1.25rem;
  --fb-text-2xl: 1.5rem;
  --fb-text-3xl: 2rem;
  
  --fb-spacing-xs: 0.25rem;
  --fb-spacing-sm: 0.5rem;
  --fb-spacing-md: 1rem;
  --fb-spacing-lg: 1.5rem;
  --fb-spacing-xl: 2rem;
  --fb-spacing-2xl: 3rem;
  
  --fb-radius-sm: 4px;
  --fb-radius-md: 8px;
  --fb-radius-lg: 12px;
  --fb-radius-xl: 16px;
  --fb-radius-full: 9999px;
  
  --fb-shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.5);
  --fb-shadow-md: 0 4px 6px rgba(0, 0, 0, 0.5);
  --fb-shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.5);
}

@media (prefers-color-scheme: light) {
  :root {
    --fb-bg-primary: #ffffff;
    --fb-bg-secondary: #f5f5f5;
    --fb-bg-tertiary: #eeeeee;
    --fb-bg-card: #ffffff;
    
    --fb-text-primary: #111111;
    --fb-text-secondary: #666666;
    --fb-text-muted: #999999;
    
    --fb-border: #e0e0e0;
    --fb-border-light: #f0f0f0;
    
    --fb-shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.1);
    --fb-shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
    --fb-shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1);
  }
}
```

### 7.2 Typography Scale

```css
.text-xs { font-size: var(--fb-text-xs); }
.text-sm { font-size: var(--fb-text-sm); }
.text-base { font-size: var(--fb-text-base); }
.text-lg { font-size: var(--fb-text-lg); }
.text-xl { font-size: var(--fb-text-xl); }
.text-2xl { font-size: var(--fb-text-2xl); }
.text-3xl { font-size: var(--fb-text-3xl); }

.font-normal { font-weight: 400; }
.font-medium { font-weight: 500; }
.font-semibold { font-weight: 600; }
.font-bold { font-weight: 700; }
```

### 7.3 Component Styles Pattern

Each Web Component uses scoped styles via Shadow DOM:

```javascript
export class FBStatusCard extends FBElement {
  styles() {
    return `
      :host {
        display: block;
        background: var(--fb-bg-card);
        border: 1px solid var(--fb-border);
        border-radius: var(--fb-radius-lg);
        padding: var(--fb-spacing-lg);
      }
      
      .title {
        font-size: var(--fb-text-sm);
        color: var(--fb-text-secondary);
        margin-bottom: var(--fb-spacing-sm);
      }
      
      .value {
        font-size: var(--fb-text-2xl);
        font-weight: 600;
        color: var(--fb-text-primary);
      }
      
      .trend {
        font-size: var(--fb-text-sm);
        margin-top: var(--fb-spacing-xs);
      }
      
      .trend.positive { color: var(--fb-success); }
      .trend.negative { color: var(--fb-error); }
    `;
  }
}
```

---

## 8. Chat Interface

### 8.1 Message Types

| Type | Appearance | Content |
|------|------------|---------|
| user | Right-aligned, primary color bg | User's text input |
| assistant | Left-aligned, secondary bg | AI response text |
| artifact | Full-width, bordered | Interactive chart/table/gantt |
| system | Center-aligned, muted | System notifications |

### 8.2 Invoice Ingestion Workflow

FutureBuild supports a streamlined "Drag and Drop" workflow for processing site documents:

1. **User Action:** User drags an invoice (PDF/Image) into the chat area.
2. **UI State:** The chat input switches to a "Processing..." skeleton state.
3. **Backend Trigger:** The file is uploaded, and the `PROCESS_INVOICE` agent intent is triggered.
4. **Agent Response:** The agent returns a JSON payload containing the extracted data.
5. **UI Update:** The chat renders the `<fb-artifact-invoice>` component, allowing the user to review and approve the extraction.

### 8.3 Agent Intent & Artifact Mapping

The Chat Orchestrator maps specific Layer 4 Agent intents to specialized UI Artifacts as defined in the `ArtifactType` enum in [API_AND_TYPES_SPEC.md](file:///home/colton/Replit%20Specs/API_AND_TYPES_SPEC.md):

| Agent Intent | UI Artifact | ArtifactType Enum | Description |
|--------------|-------------|-------------------|-------------|
| `PROCESS_INVOICE` | `<fb-artifact-invoice>` | `Invoice` | OCR validation and WBS coding |
| `QUERY_BUDGET` | `<fb-artifact-budget>` | `Budget_View` | 3-column financial overview |
| `GET_SCHEDULE` | `<fb-artifact-gantt>` | `Gantt_View` | Interactive project timeline |
| `MANAGE_CONTACTS` | `<fb-contact-card>` | `N/A` | Project Rolodex / Team management |
| `EXPLAIN_DELAY` | `<fb-artifact-card>` | `N/A` | Critical path bottleneck analysis |

### 8.4 Streaming Response Handler

```javascript
async function handleStreamingResponse(projectId, message) {
  const assistantMessage = createPendingMessage();
  
  for await (const chunk of api.streamChat(projectId, message)) {
    const data = JSON.parse(chunk);
    
    if (data.type === 'text') {
      assistantMessage.appendText(data.content);
    } else if (data.type === 'artifact') {
      assistantMessage.addArtifact(data.artifact);
    } else if (data.type === 'done') {
      assistantMessage.finalize();
    }
  }
}
```

---

## 9. Gantt Chart Component

### 9.1 Architecture

```javascript
export class FBArtifactGantt extends FBElement {
  constructor() {
    super();
    this.tasks = [];
    this.viewMode = 'week';
    this.scrollPosition = 0;
  }

  set data(value) {
    this.tasks = value.tasks || [];
    this.criticalPath = value.critical_path || [];
    this.renderGantt();
  }

  styles() {
    return `
      :host {
        display: block;
        overflow: hidden;
      }
      
      .gantt-container {
        display: flex;
        height: 400px;
      }
      
      .task-list {
        width: 250px;
        border-right: 1px solid var(--fb-border);
        overflow-y: auto;
      }
      
      .timeline {
        flex: 1;
        overflow-x: auto;
        overflow-y: auto;
      }
      
      .task-bar {
        height: 24px;
        border-radius: 4px;
        position: absolute;
      }
      
      .task-bar.critical {
        background: var(--fb-error);
      }
      
      .task-bar.normal {
        background: var(--fb-primary);
      }
      
      .task-bar.completed {
        background: var(--fb-success);
      }
    `;
  }

  renderGantt() {
    const timeline = this.$('.timeline');
    if (!timeline) return;
    
    const startDate = this.getProjectStart();
    const dayWidth = this.getDayWidth();
    
    this.tasks.forEach((task, index) => {
      const bar = document.createElement('div');
      bar.className = `task-bar ${this.getTaskClass(task)}`;
      bar.style.left = `${this.getTaskLeft(task, startDate, dayWidth)}px`;
      bar.style.width = `${task.duration * dayWidth}px`;
      bar.style.top = `${index * 30 + 5}px`;
      timeline.appendChild(bar);
    });
  }
}
```

### 9.2 Interactions

| Action | Behavior |
|--------|----------|
| Hover task | Show tooltip with details |
| Click task | Open task detail modal |
| Scroll horizontal | Pan timeline |
| Zoom buttons | Change day width |
| Critical path toggle | Highlight/dim non-critical |

---

## 10. Mobile Experience

### 10.1 Bottom Navigation

```javascript
export class FBMobileNav extends FBElement {
  styles() {
    return `
      :host {
        display: none;
        position: fixed;
        bottom: 0;
        left: 0;
        right: 0;
        background: var(--fb-bg-secondary);
        border-top: 1px solid var(--fb-border);
        padding: var(--fb-spacing-sm);
        z-index: 100;
      }
      
      @media (max-width: 767px) {
        :host { display: flex; }
      }
      
      .nav-items {
        display: flex;
        justify-content: space-around;
        width: 100%;
      }
      
      .nav-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: var(--fb-spacing-xs);
        color: var(--fb-text-secondary);
        text-decoration: none;
      }
      
      .nav-item.active {
        color: var(--fb-primary);
      }
    `;
  }

  template() {
    return `
      <nav class="nav-items">
        <a class="nav-item" href="/dashboard">
          <fb-icon name="dashboard"></fb-icon>
          <span>Dashboard</span>
        </a>
        <a class="nav-item" href="/chat">
          <fb-icon name="chat"></fb-icon>
          <span>Chat</span>
        </a>
        <a class="nav-item" href="/tasks">
          <fb-icon name="tasks"></fb-icon>
          <span>Tasks</span>
        </a>
        <a class="nav-item" href="/more">
          <fb-icon name="menu"></fb-icon>
          <span>More</span>
        </a>
      </nav>
    `;
  }
}
```

### 10.2 Touch Gestures

| Gesture | Action |
|---------|--------|
| Swipe left | Navigate to next view |
| Swipe right | Navigate to previous view |
| Pull down | Refresh current data |
| Long press task | Open context menu |

---

## 11. Notifications

### 11.1 Notification Types

| Type | Icon | Priority | Auto-dismiss |
|------|------|----------|--------------|
| inspection | calendar | High | No |
| task_due | clock | Medium | 10s |
| message | chat | Medium | 10s |
| weather | cloud | Low | 15s |
| system | info | Low | 5s |

### 11.2 Notification Center

```javascript
export class FBNotificationCenter extends FBElement {
  constructor() {
    super();
    this.notifications = [];
    this.unreadCount = 0;
  }

  connectedCallback() {
    super.connectedCallback();
    this.setupWebSocket();
  }

  setupWebSocket() {
    const ws = new WebSocket(`wss://${location.host}/ws/notifications`);
    
    ws.onmessage = (event) => {
      const notification = JSON.parse(event.data);
      this.addNotification(notification);
    };
  }

  addNotification(notification) {
    this.notifications.unshift(notification);
    this.unreadCount++;
    this.render();
    this.emit('notification-received', notification);
  }
}
```

---

## 12. Build & Development

### 12.1 Vite Configuration (TypeScript)

```typescript
import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  root: 'src',
  base: '/',
  build: {
    outDir: '../dist',
    target: 'esnext',
    modulePreload: { polyfill: false },
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'src/index.html')
      }
    }
  },
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src')
    }
  }
});
```

### 12.2 TSConfig (Strict Mode)

```json
{
  "compilerOptions": {
    "target": "ESNext",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ESNext", "DOM", "DOM.Iterable"],
    "moduleResolution": "Node",
    "strict": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "esModuleInterop": true,
    "noEmit": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "skipLibCheck": true,
    "experimentalDecorators": true
  },
  "include": ["src/**/*.ts"]
}
```

### 12.3 File Structure

```
frontend/
├── src/
│   ├── index.html
│   ├── main.ts
│   ├── vite-env.d.ts
│   ├── components/
│   │   ├── base/
│   │   │   └── FBElement.ts
│   │   ├── chat/
│   │   │   ├── FBChat.ts
│   │   │   └── FBMessage.ts
│   │   ├── artifacts/
│   │   │   ├── FBArtifactInvoice.ts
│   │   │   ├── FBArtifactBudget.ts
│   │   │   └── FBArtifactGantt.ts
│   │   └── shared/
│   │       └── FBIcon.ts
│   ├── store/
│   │   └── signals.ts
│   ├── types/           <-- Mirrored from Go
│   │   ├── task.ts
│   │   ├── invoice.ts
│   │   └── contact.ts
│   └── styles/
│       └── app.css
├── package.json
├── tsconfig.json
└── vite.config.ts
```

---

## 13. Standards & Linting

To enforce strict typing and consistent quality, FutureBuild requires:

- **ESLint:** Configured with `@typescript-eslint/recommended` and `lit/recommended`.
- **Prettier:** Integrated into ESLint to enforce formatting on save.
- **Commit Hooks:** Husky + lint-staged to prevent non-compliant code from entering the repo.
- **Documentation:** All components must use JSDoc for property/method descriptions.

---

## 14. Performance Considerations

### 13.1 Code Splitting

```javascript
const routes = {
  '/dashboard': () => import('./views/Dashboard'),
  '/schedule': () => import('./views/Schedule'),
  '/tasks': () => import('./views/Tasks'),
  '/documents': () => import('./views/Documents'),
};

async function loadView(path) {
  const loader = routes[path];
  if (loader) {
    const module = await loader();
    return new module.default();
  }
}
```

### 13.2 Virtual Scrolling

For long task lists, implement virtual scrolling:

```javascript
export class FBVirtualList extends FBElement {
  constructor() {
    super();
    this.items = [];
    this.itemHeight = 50;
    this.visibleCount = 20;
    this.scrollTop = 0;
  }

  get visibleItems() {
    const startIndex = Math.floor(this.scrollTop / this.itemHeight);
    const endIndex = startIndex + this.visibleCount;
    return this.items.slice(startIndex, endIndex);
  }

  handleScroll(e) {
    this.scrollTop = e.target.scrollTop;
    this.render();
  }
}
```

---

## 15. Accessibility

### 14.1 ARIA Labels

```javascript
template() {
  return `
    <nav aria-label="Main navigation">
      <button 
        aria-expanded="${this.isOpen}"
        aria-controls="menu"
        aria-label="Toggle navigation"
      >
        <fb-icon name="menu"></fb-icon>
      </button>
      <ul id="menu" role="menu" ${!this.isOpen ? 'hidden' : ''}>
        ...
      </ul>
    </nav>
  `;
}
```

### 14.2 Keyboard Navigation

| Key | Action |
|-----|--------|
| Tab | Move focus to next element |
| Shift+Tab | Move focus to previous element |
| Enter/Space | Activate focused element |
| Escape | Close modal/dropdown |
| Arrow keys | Navigate within menus/lists |

---

*Document Version: 1.1.0*

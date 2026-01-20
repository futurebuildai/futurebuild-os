# FutureBuild Command Center - Demo Script

> **Purpose**: Step-by-step guide for demonstrating the AI Agent Command Center to prospects.
> **Duration**: ~10 minutes

---

## Pre-Demo Setup

```bash
# Ensure dev server is running
cd frontend && npm run dev -- --port 5174
```

Open: http://localhost:5174

---

## Demo Flow

### 1. Login Experience (1 min)
1. Show the **clean login page** with "FutureBuild" branding
2. Enter any email (e.g., `demo@futurebuild.com`)
3. Click **"Request Magic Link"**
4. Explain: *"In production, this sends a secure magic link. For demo, we simulate login."*

### 2. Command Center Overview (2 min)
After login, highlight the **3-panel layout**:

| Panel | Purpose |
|-------|---------|
| **Left** | Daily Focus tasks, Projects list, Agent Activity |
| **Center** | Conversation with the AI agent |
| **Right** | Artifacts (invoices, budgets, schedules) |

**Key talking points:**
- *"Daily Focus surfaces what matters today"*
- *"AI Agent lives in the center — always available"*
- *"Artifacts appear automatically as the agent works"*

### 3. Invoice Processing Demo (3 min)

Open DevTools Console (`F12` → Console) and run:

```javascript
window.fb.triggerScenario('invoice_success')
```

**What happens:**
1. Message appears: *"I've analyzed your invoice..."*
2. **Invoice artifact** renders in right panel
3. Shows: **ABC Lumber Co.** - **$5,600.00**

**Script:**
> *"When a subcontractor emails an invoice, our AI extracts it, validates against your budget, and presents it for approval. One click to approve, deny, or edit."*

### 4. Budget Overview Demo (2 min)

```javascript
window.fb.triggerScenario('budget_overview')
```

**What happens:**
1. Message: *"Here's the current budget overview..."*
2. **Budget artifact** shows:
   - Total: **$450,000**
   - Spent: **$125,000**
   - Materials & Labor breakdown

### 5. Schedule Update Demo (2 min)

```javascript
window.fb.triggerScenario('schedule_change')
```

**What happens:**
1. **Gantt artifact** appears showing updated timeline
2. Foundation and Framing phases with durations

---

## Mobile Demo (Optional)

Resize browser to **375px width** (or use Chrome DevTools device mode):
- Panels collapse to single-view mode
- Hamburger icon opens navigation overlay
- Touch-friendly mobile experience

---

## Available DevTools Commands

| Command | Description |
|---------|-------------|
| `window.fb.getScenarios()` | List all available scenarios |
| `window.fb.triggerScenario('name')` | Trigger a specific scenario |
| `window.fb.triggerMessage('text')` | Inject a custom agent message |
| `window.fb.setTyping(true/false)` | Toggle typing indicator |

### All Scenarios
- `text_reply` - Simple text response
- `invoice_success` - Invoice extraction
- `budget_overview` - Project budget
- `schedule_change` - Gantt schedule update
- `typing_long` - Extended typing indicator
- `error_network` - Network error simulation

---

## Closing Points

1. **"AI handles the grunt work"** — invoice processing, schedule updates
2. **"Human stays in control"** — approve/deny workflow
3. **"Everything in one place"** — no app switching
4. **"Built for construction"** — understands trades, schedules, budgets

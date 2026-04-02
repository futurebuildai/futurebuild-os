# Design System — GableLBM Industrial Dark

**Document ID:** AG-05-DS
**System:** FutureBuild Brain (System of Connection)
**Created:** 2026-04-02
**Pipeline Stage:** 05 - Design System
**Status:** COMPLETE
**Source of Truth:** `reference-vault/futurebuild-os/specs/GABLE_LBM_DESIGN_SYSTEM.md`

---

## 1. Design Principles

Grounded in persona research (General Contractor, Bookkeeper, Subcontractor Admin, Integration Admin) and the Brain vision: "The Connection Plane."

| # | Principle | Rationale | Persona Link |
|---|-----------|-----------|--------------|
| 1 | **Chat-First, Dashboard-Second** | Brain is an orchestration layer. Users interact primarily through Maestro (natural language), with admin dashboards as secondary management surfaces. | GC uses Maestro to "order roofing materials"; Admin uses dashboard to manage OIDC clients |
| 2 | **Ecosystem Visibility** | Brain connects 7+ external systems. Users must see the health and status of every connection at a glance — like a network operations center. | Integration Admin monitoring GableERP/QuickBooks/LocalBlue connectivity |
| 3 | **Trust Through Transparency** | Every AI decision (intent classification, tool selection, A2A dispatch) must be auditable. Show the reasoning chain, not just the result. | Bookkeeper verifying QuickBooks PO was auto-created correctly |
| 4 | **Consistent Industrial Dark** | Brain and OS share the same GableLBM visual language. Users moving between systems experience zero visual context-switching. | Same Gable Green, Deep Space, Slate Steel across both products |
| 5 | **Progressive Migration** | Brain's legacy UI is React 19. New features use Lit 3.0, existing pages migrate incrementally. Both frameworks render with the same token system. | React→Lit migration strategy; Tailwind CSS 4 bridges both |

---

## 2. Color System

Brain shares the **identical color palette** with FB-OS. The GableLBM Industrial Dark palette is ecosystem-wide.

### 2.1 Core Palette

| Token | Hex | Role | WCAG on Deep Space |
|-------|-----|------|-------------------|
| **Gable Green** | `#00FFA3` | Primary accent, CTAs, success states | AAA (12.8:1) |
| **Deep Space** | `#0A0B10` | Base background | N/A (base) |
| **Slate Steel** | `#161821` | Surface containers, cards, panels | N/A (surface) |
| **Blueprint Blue** | `#38BDF8` | Secondary accent, info, links | AAA (8.6:1) |
| **Safety Red** | `#F43F5E` | Error, destructive, critical alerts | AA (5.1:1) |
| **Amber Warning** | `#F59E0B` | Warning, approaching deadlines | AA (6.3:1) |

### 2.2 Brain-Specific Semantic Tokens

In addition to the shared palette, Brain adds integration-specific color coding:

| Token | Hex | Usage |
|-------|-----|-------|
| `--fb-color-gable-erp` | `#22C55E` | GableERP integration badge |
| `--fb-color-xui` | `#3B82F6` | XUI integration badge |
| `--fb-color-localblue` | `#F43F5E` | LocalBlue integration badge |
| `--fb-color-quickbooks` | `#2CA01C` | QuickBooks integration badge |
| `--fb-color-onebuild` | `#F59E0B` | 1Build integration badge |
| `--fb-color-gmail` | `#EA4335` | Gmail integration badge |
| `--fb-color-outlook` | `#0078D4` | Outlook integration badge |

### 2.3 Surface Elevation Scale

Identical to FB-OS:

| Layer | Token | Hex | Usage |
|-------|-------|-----|-------|
| 0 | `--fb-surface-0` | `#0A0B10` | Page background |
| 1 | `--fb-surface-1` | `#161821` | Cards, panels |
| 2 | `--fb-surface-2` | `#1E2029` | Nested containers, hover |
| 3 | `--fb-surface-3` | `#252836` | Modals, dropdowns |

### 2.4 Tailwind CSS 4 Extension

Brain uses Tailwind CSS 4 as its primary styling utility (React 19 components + Lit 3.0 components during migration).

```css
@theme {
  /* Core Gable Palette (shared with FB-OS) */
  --color-gable-green: #00FFA3;
  --color-gable-green-dim: #005234;
  --color-gable-green-bright: #66FFC8;
  --color-deep-space: #0A0B10;
  --color-deep-space-soft: #0E0F15;
  --color-slate-steel: #161821;
  --color-slate-steel-raised: #1E2029;
  --color-slate-steel-elevated: #252836;

  /* Semantic Colors */
  --color-blueprint-blue: #38BDF8;
  --color-safety-red: #F43F5E;
  --color-amber-warning: #F59E0B;

  /* Text Colors */
  --color-text-primary: #F0F0F5;
  --color-text-secondary: #8B8D98;
  --color-text-muted: #5A5B66;

  /* Integration Colors */
  --color-gable-erp: #22C55E;
  --color-xui: #3B82F6;
  --color-localblue: #F43F5E;
  --color-quickbooks: #2CA01C;
  --color-onebuild: #F59E0B;
  --color-gmail: #EA4335;
  --color-outlook: #0078D4;

  /* Glass Effects */
  --color-glass-bg: rgba(22, 24, 33, 0.6);
  --color-glass-border: rgba(255, 255, 255, 0.05);
  --color-glass-panel: rgba(10, 11, 16, 0.8);

  /* Typography */
  --font-sans: 'Outfit', system-ui, -apple-system, sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;

  /* Spacing Scale (4px base) */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-2xl: 48px;

  /* Border Radius */
  --radius-xs: 6px;
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 16px;
  --radius-xl: 24px;
  --radius-full: 9999px;

  /* Elevation Shadows */
  --shadow-elevation-1: 0px 1px 3px 1px rgba(0,0,0,0.25), 0px 1px 2px 0px rgba(0,0,0,0.4);
  --shadow-elevation-2: 0px 2px 6px 2px rgba(0,0,0,0.25), 0px 1px 2px 0px rgba(0,0,0,0.4);
  --shadow-elevation-3: 0px 4px 8px 3px rgba(0,0,0,0.25), 0px 1px 3px 0px rgba(0,0,0,0.4);
  --shadow-glow: 0 0 20px rgba(0,255,163,0.15);
  --shadow-glow-strong: 0 0 30px rgba(0,255,163,0.3), 0 0 60px rgba(0,255,163,0.1);

  /* Animation */
  --animate-pulse-glow: pulse-glow 2s ease-in-out infinite;
}
```

---

## 3. Typography System

### 3.1 Font Families

| Font | Role | Weight Range |
|------|------|-------------|
| **Outfit** | All UI labels, headings, body copy, navigation, chat messages | 400, 500, 700 |
| **JetBrains Mono** | Integration IDs, webhook payloads, JSON previews, execution timestamps, tool schemas, monetary values | 400, 500, 700 |

### 3.2 JetBrains Mono Mandate (Brain-Specific)

JetBrains Mono MUST be used for:
- MCP tool names and JSON schemas
- Webhook payload previews
- Integration execution timestamps and durations
- OIDC client IDs, redirect URIs
- A2A task IDs, correlation IDs
- All monetary values in QuickBooks data
- API endpoint paths in admin UI
- Execution log entries

Outfit MUST be used for:
- Navigation labels, menu items
- Maestro chat messages (user and assistant)
- Button text, form labels
- Card titles, section headers
- Integration names (human-readable)
- Status text, descriptions

### 3.3 Type Scale

Identical to FB-OS — see shared GableLBM type scale (Display through Label, all Outfit; Data variants use JetBrains Mono).

---

## 4. Spacing, Shape, Elevation

All shared with FB-OS. See the GableLBM specification:
- **Spacing:** 4px base (xs=4, sm=8, md=16, lg=24, xl=32, 2xl=48)
- **Shape:** 6px through 24px + 9999px (full)
- **Elevation:** 3-tier shadow system + glow accents

---

## 5. Glassmorphism Specification

Identical to FB-OS:
- `.glass-card`: `rgba(22, 24, 33, 0.6)`, `blur(24px)`, `1px solid rgba(255,255,255,0.05)`
- `.glass-panel`: `rgba(10, 11, 16, 0.8)`, `blur(48px)`
- Rules: 24px minimum blur, 80% maximum opacity, always 5% white border

---

## 6. FBBaseElement Class (Lit 3.0)

Brain uses the **same FBBaseElement** as FB-OS for all new Lit components. During the React→Lit migration, React components use Tailwind CSS 4 utilities that map to the same tokens.

```typescript
import { LitElement, css, CSSResultGroup } from 'lit';

export abstract class FBElement extends LitElement {
  static override styles: CSSResultGroup = css`
    :host { box-sizing: border-box; }
    :host *, :host *::before, :host *::after { box-sizing: inherit; }

    .glass-card {
      background: rgba(22, 24, 33, 0.6);
      backdrop-filter: blur(24px);
      -webkit-backdrop-filter: blur(24px);
      border: 1px solid rgba(255, 255, 255, 0.05);
      border-radius: 16px;
      box-shadow: var(--md-sys-elevation-1);
    }
    .glass-panel {
      background: rgba(10, 11, 16, 0.8);
      backdrop-filter: blur(48px);
      -webkit-backdrop-filter: blur(48px);
      border: 1px solid rgba(255, 255, 255, 0.05);
    }
    .hover-lift { transition: transform 0.3s ease-out, box-shadow 0.3s ease-out; }
    .hover-lift:hover { transform: translateY(-4px); box-shadow: var(--md-sys-elevation-2); }
    .glow-accent { box-shadow: 0 0 20px rgba(0, 255, 163, 0.15); }
    .glow-accent-strong {
      box-shadow: 0 0 30px rgba(0, 255, 163, 0.3), 0 0 60px rgba(0, 255, 163, 0.1);
    }
    .active-indicator { position: relative; }
    .active-indicator::before {
      content: ''; position: absolute; left: 0; top: 50%;
      transform: translateY(-50%); width: 3px; height: 60%;
      background: #00FFA3; border-radius: 0 3px 3px 0;
    }
    .btn-primary { transition: transform 0.15s ease-out, box-shadow 0.15s ease-out; }
    .btn-primary:hover { box-shadow: 0 0 20px rgba(0, 255, 163, 0.15); transform: translateY(-1px); }
    .btn-primary:active { transform: scale(0.95); box-shadow: none; }
    .skeleton {
      background: linear-gradient(90deg,
        var(--md-sys-color-surface-container-high) 25%,
        var(--md-sys-color-surface-container) 50%,
        var(--md-sys-color-surface-container-high) 75%
      );
      background-size: 200% 100%; animation: shimmer 2s linear infinite; border-radius: 4px;
    }
    @keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }

    .data-mono { font-family: 'JetBrains Mono', monospace; font-variant-numeric: tabular-nums; }
  `;

  protected emit<T = unknown>(name: string, detail?: T): CustomEvent<T> {
    const event = new CustomEvent<T>(name, { bubbles: true, composed: true, detail: detail as T });
    this.dispatchEvent(event);
    return event;
  }
}
```

---

## 7. Component Catalog

### 7.1 Shared Atoms

Brain uses the same atoms as FB-OS: `<fb-button>`, `<fb-icon>`, `<fb-badge>`, `<fb-text>`, `<fb-input>`, `<fb-select>`, `<fb-chip>`, `<fb-spinner>`, `<fb-avatar>`.

### 7.2 Brain-Specific Molecules

| Component | Tag | Description |
|-----------|-----|-------------|
| Integration Card | `<fb-integration-card>` | Glass-card showing integration name, color-coded icon, health status, last sync time, execution count |
| MCP Tool Card | `<fb-mcp-tool-card>` | Tool name, JSON schema preview (JetBrains Mono), trigger/action badge, test button |
| Chat Bubble | `<fb-chat-bubble>` | User/assistant message in Maestro. Outfit for text; inline code in JetBrains Mono |
| Execution Log Row | `<fb-execution-row>` | Timestamp (mono) + action + source→target + status badge + duration (mono) |
| Platform Node | `<fb-platform-node>` | XY Flow canvas node for ecosystem visualization: integration icon + health dot |
| Connection Edge | `<fb-connection-edge>` | XY Flow edge with animated pulse on active data flow |
| OIDC Client Row | `<fb-oidc-client-row>` | Client ID (mono), redirect URIs, grant types, status toggle |
| Webhook Event | `<fb-webhook-event>` | A2A webhook delivery: timestamp, target URL, HTTP status, retry count |

### 7.3 Brain-Specific Organisms

| Component | Tag | Description |
|-----------|-----|-------------|
| MCP Registry Browser | `<fb-mcp-registry>` | Searchable/filterable list of registered MCP servers and their tools |
| OIDC Client Manager | `<fb-oidc-manager>` | CRUD interface for OIDC clients with test token generation |
| Ecosystem Canvas | `<fb-ecosystem-canvas>` | XY Flow graph of platforms, connections, integrations with real-time health |
| Maestro Chat | `<fb-maestro-chat>` | Full chat interface: message list + input bar + suggested prompts |
| Maestro Drawer | `<fb-maestro-drawer>` | Slide-out 420x600px chat panel triggered by FAB button |
| Integration Monitor | `<fb-integration-monitor>` | Detail panel: execution history, error rates, latency chart, test runner |
| Setup Wizard | `<fb-setup-wizard>` | 5-step onboarding: system selection → templates → AI config → review → activate |
| Activity Ticker | `<fb-activity-ticker>` | Real-time event feed showing webhook dispatches, tool executions, auth events |

### 7.4 Page Components

| Component | Tag | Surface | Description |
|-----------|-----|---------|-------------|
| Home Page | `<fb-home-page>` | Hub Admin | Split: Dashboard grid (left) + Maestro chat (right) |
| Ecosystem Page | `<fb-ecosystem-page>` | Hub Admin | XY Flow canvas + detail panels |
| Marketplace Page | `<fb-marketplace-page>` | Hub Admin | Template catalog grid |
| Settings Page | `<fb-settings-page>` | Hub Admin | Org settings, member management |
| Admin Registry | `<fb-admin-registry>` | Hub Admin | MCP registry browser + OIDC client manager |
| Login Page | `<fb-login-page>` | Auth | Login form + magic link + SSO options |

---

## 8. Interaction Patterns

### 8.1 States

Identical to FB-OS: Default, Hover (lift), Active (scale), Focus (green outline), Disabled (40% opacity), Loading (skeleton), Empty (centered message), Error (red border-left), Success (glow flash), Offline (amber badge).

### 8.2 Motion & Animation

Identical to FB-OS timing scale. Brain adds:

| Motion | Duration | Usage |
|--------|----------|-------|
| Pulse Glow | 2000ms ease-in-out infinite | Active webhook/event indicator |
| Slide-in Right | 300ms emphasized | Maestro drawer open |
| Fade-in | 200ms ease-out | Chat message appearance |
| Canvas Zoom | 300ms emphasized | Ecosystem graph zoom/pan |

### 8.3 Responsive Breakpoints

| Breakpoint | Layout |
|------------|--------|
| ≥1440px | Dashboard grid (left 55%) + Maestro chat (right 45%) |
| ≥1200px | Dashboard grid full-width; Maestro as FAB drawer |
| ≥768px | Single column; Maestro as FAB drawer |
| <768px | Full-width; Maestro as full-screen overlay |

---

## 9. React ↔ Lit Migration Strategy

### 9.1 Coexistence Model

During migration, React and Lit components coexist:

| Pattern | Implementation |
|---------|---------------|
| Lit in React | React wrapper component uses `ref` to mount Lit element in `useEffect` |
| Shared tokens | Both frameworks consume the same CSS custom properties from `variables.css` |
| Shared Tailwind | Tailwind CSS 4 `@theme` block is framework-agnostic |
| Event bridge | Lit `emit()` → React `addEventListener` in wrapper `useEffect` |

### 9.2 Migration Phases

| Phase | Pages to Migrate | Framework |
|-------|-----------------|-----------|
| MVP | Admin Registry, OIDC Manager | Built new in Lit 3.0 |
| Post-MVP 1 | Login, Setup Wizard | Rewrite React → Lit |
| Post-MVP 2 | Ecosystem Canvas, Marketplace | Rewrite React → Lit |
| Post-MVP 3 | Home Page, Settings | Rewrite React → Lit |

---

## 10. Accessibility Requirements

Identical to FB-OS:
- WCAG AA contrast (4.5:1 text, 3:1 large text)
- Full keyboard navigation with `:focus-visible` green outline
- ARIA landmarks on all surfaces
- `prefers-reduced-motion` support
- 48px minimum touch targets
- Error identification by color AND text AND icon

---

## 11. Dark Mode Policy

Dark-only, identical to FB-OS. No light mode. The Industrial Dark aesthetic is ecosystem-wide brand identity.

---

## 12. Icon System

- **Provider:** Material Symbols (Outlined, weight 400)
- **Brain-specific icons:** SVG sprite for integration logos (GableERP, QuickBooks, LocalBlue, XUI, 1Build)
- **Sizing:** 20px (sm), 24px (md), 32px (lg)
- **Color:** Inherits `currentColor`; integration icons use brand colors

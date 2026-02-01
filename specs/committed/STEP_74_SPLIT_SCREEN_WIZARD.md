# Step 74: Split-Screen Wizard (`fb-view-onboarding`)

**Phase:** 11 (The Conversational Hook)  
**Status:** READY FOR IMPLEMENTATION  
**Est. Duration:** 1 Day  
**Owner:** Frontend Developer

---

## 🎯 Objective

Create a new view component `<fb-view-onboarding>` that implements a split-screen wizard layout with Chat (left) and Live Form (right) for AI-driven project creation.

---

## 📐 Architectural Guardrails

> [!IMPORTANT]
> These constraints are non-negotiable and align with the system architecture.

### Layer Compliance
- **Layer 4 (Action Engine)**: This component RECEIVES data from agents; it does NOT perform physics calculations
- **Layer 2 (Data Spine)**: Form submissions create records via REST API, NOT direct database access
- The component is **stateless** regarding schedule math — all calculations happen server-side

### Design System Alignment
Reference: [FRONTEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_SCOPE.md) Section 7

```css
/* REQUIRED: Use existing CSS custom properties */
--fb-bg-primary, --fb-bg-card, --fb-text-primary
--fb-spacing-md, --fb-spacing-lg, --fb-spacing-xl
--fb-radius-lg, --fb-border
```

### Component Patterns
Reference: [FRONTEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_SCOPE.md) Section 4

- Extend `FBViewElement` (not raw `LitElement`)
- Use Shadow DOM for style encapsulation
- Emit custom events via `this.emit()` helper
- Register in component index

### Type Safety
Reference: [FRONTEND_TYPES_SPEC.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_TYPES_SPEC.md)

- All props MUST be typed with TypeScript interfaces
- NO `any` types allowed
- Form state MUST align with `CreateProjectRequest` interface

---

## 📋 Implementation Checklist

### Pre-Implementation
- [ ] Read existing view components for patterns: `fb-view-projects.ts`, `fb-view-chat.ts`
- [ ] Review `FBViewElement` base class
- [ ] Confirm router registration pattern in `fb-app-shell.ts`

### Component Files to Create

```
frontend/src/components/
├── views/
│   └── fb-view-onboarding.ts          # Main wizard view
├── features/onboarding/
│   ├── fb-onboarding-chat.ts          # Left panel (chat)
│   ├── fb-onboarding-form.ts          # Right panel (form)
│   └── fb-onboarding-dropzone.ts      # File drop area
```

### Step-by-Step Tasks

#### Task 1: Create View Skeleton (30 min)
```typescript
// frontend/src/components/views/fb-view-onboarding.ts

import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/fb-view-element';

// Import sub-components
import '../features/onboarding/fb-onboarding-chat';
import '../features/onboarding/fb-onboarding-form';

/**
 * Split-screen wizard for AI-driven project creation.
 * @see planning/PHASE_11_PRD.md Section 5.1
 */
@customElement('fb-view-onboarding')
export class FBViewOnboarding extends FBViewElement {
  static override styles = [
    FBViewElement.styles,
    css`
      :host {
        display: flex;
        flex-direction: column;
        height: 100%;
        background: var(--fb-bg-primary);
      }

      .wizard-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--fb-spacing-md) var(--fb-spacing-lg);
        border-bottom: 1px solid var(--fb-border);
      }

      .wizard-title {
        font-size: var(--fb-text-xl);
        font-weight: 600;
        color: var(--fb-text-primary);
      }

      .wizard-body {
        display: flex;
        flex: 1;
        overflow: hidden;
      }

      .panel-chat {
        flex: 1;
        display: flex;
        flex-direction: column;
        border-right: 1px solid var(--fb-border);
      }

      .panel-form {
        flex: 1;
        overflow-y: auto;
        padding: var(--fb-spacing-xl);
      }

      /* Responsive: Stack on tablet/mobile */
      @media (max-width: 1023px) {
        .wizard-body {
          flex-direction: column;
        }
        .panel-chat {
          border-right: none;
          border-bottom: 1px solid var(--fb-border);
          max-height: 50%;
        }
      }

      @media (max-width: 767px) {
        /* Tab toggle mode - implement later */
      }
    `
  ];

  render(): TemplateResult {
    return html`
      <div class="wizard-header">
        <span class="wizard-title">New Project</span>
        <fb-icon-button
          icon="close"
          @click=${this._handleClose}
        ></fb-icon-button>
      </div>
      <div class="wizard-body">
        <div class="panel-chat">
          <fb-onboarding-chat></fb-onboarding-chat>
        </div>
        <div class="panel-form">
          <fb-onboarding-form></fb-onboarding-form>
        </div>
      </div>
    `;
  }

  private _handleClose(): void {
    // Navigate back to projects view
    this.emit('navigate', { path: '/projects' });
  }
}
```

#### Task 2: Create Chat Panel Component (45 min)
```typescript
// frontend/src/components/features/onboarding/fb-onboarding-chat.ts

/**
 * Left panel: Conversational interface for project onboarding.
 * Reuses fb-message-list and fb-input-bar from Phase 10.
 */
@customElement('fb-onboarding-chat')
export class FBOnboardingChat extends LitElement {
  @state() private _messages: OnboardingMessage[] = [];
  @state() private _isProcessing = false;

  // Initial greeting from "The Interrogator"
  connectedCallback() {
    super.connectedCallback();
    this._addSystemMessage(
      "Hi! Let's set up your new project. You can describe it, or drag a blueprint here to get started."
    );
  }

  // MUST reuse existing chat components
  render() {
    return html`
      <fb-message-list .messages=${this._messages}></fb-message-list>
      <fb-onboarding-dropzone
        @file-dropped=${this._handleFileDrop}
      ></fb-onboarding-dropzone>
      <fb-input-bar
        placeholder="Describe your project..."
        .disabled=${this._isProcessing}
        @send=${this._handleSend}
      ></fb-input-bar>
    `;
  }
}
```

#### Task 3: Create Form Panel Component (60 min)
```typescript
// frontend/src/components/features/onboarding/fb-onboarding-form.ts

/**
 * Right panel: Live form that updates as AI extracts data.
 * Implements visual indicators for AI vs user-populated fields.
 * @see planning/PHASE_11_PRD.md Section 5.3
 */
@customElement('fb-onboarding-form')
export class FBOnboardingForm extends LitElement {
  @property({ type: Object }) values: Partial<CreateProjectRequest> = {};
  @property({ type: Object }) sources: Record<string, 'user' | 'ai' | 'default'> = {};
  @property({ type: Object }) confidence: Record<string, number> = {};

  // Field rendering with visual state indicators
  private _renderField(
    name: keyof CreateProjectRequest,
    label: string,
    type: 'text' | 'number' | 'date' | 'select'
  ) {
    const source = this.sources[name];
    const conf = this.confidence[name];
    
    return html`
      <div class="field ${source === 'ai' ? 'ai-populated' : ''}">
        <label>${label}</label>
        ${source === 'ai' ? html`<span class="ai-badge">✨ AI</span>` : ''}
        ${conf && conf < 0.8 ? html`<span class="verify-badge">Verify</span>` : ''}
        <input
          type=${type}
          .value=${this.values[name] ?? ''}
          @input=${(e: Event) => this._handleInput(name, e)}
        />
      </div>
    `;
  }

  render() {
    return html`
      <form @submit=${this._handleSubmit}>
        <section class="required-fields">
          <h3>Project Details</h3>
          ${this._renderField('name', 'Project Name', 'text')}
          ${this._renderField('address', 'Address', 'text')}
        </section>

        <section class="physics-fields">
          <h3>Building Specifications</h3>
          <p class="hint">These help calculate your schedule accurately.</p>
          ${this._renderField('square_footage', 'Square Feet', 'number')}
          ${this._renderField('bedrooms', 'Bedrooms', 'number')}
          ${this._renderField('bathrooms', 'Bathrooms', 'number')}
          ${this._renderField('stories', 'Stories', 'number')}
          <!-- Foundation type dropdown -->
          <!-- Topography dropdown -->
        </section>

        <button
          type="submit"
          class="btn-primary"
          ?disabled=${!this._canCreate}
        >
          Create Project
        </button>
      </form>
    `;
  }

  private get _canCreate(): boolean {
    return !!(this.values.name && this.values.address);
  }
}
```

#### Task 4: Register Route (15 min)
```typescript
// In fb-app-shell.ts or router configuration
router.register('/projects/new', () => showView('onboarding'));
```

#### Task 5: Apply Responsive Breakpoints (30 min)
- Desktop (≥1024px): Side-by-side 50/50
- Tablet (768-1023px): Stacked, chat on top
- Mobile (<768px): Tab toggle (implement toggle buttons)

---

## 🔌 Interface Contracts

### Events Emitted

| Event | Payload | Description |
|-------|---------|-------------|
| `navigate` | `{ path: string }` | Request navigation (close wizard) |
| `form-updated` | `Partial<CreateProjectRequest>` | Form state changed (for sync) |
| `project-created` | `{ projectId: string }` | Project successfully created |

### Props Received (from parent)

| Prop | Type | Description |
|------|------|-------------|
| `initialValues` | `Partial<CreateProjectRequest>` | Optional pre-filled values |

---

## ✅ Acceptance Criteria

- [ ] `fb-view-onboarding` renders split-screen layout at desktop widths
- [ ] `fb-view-onboarding` stacks panels at tablet widths
- [ ] Close button navigates back to `/projects`
- [ ] Form displays all physics-critical fields from PRD Section 5.5
- [ ] AI-populated fields show distinct visual treatment (blue border + ✨ badge)
- [ ] Low-confidence fields show "Verify" badge
- [ ] "Create Project" button is disabled until name + address are filled
- [ ] Component extends `FBViewElement` correctly
- [ ] All TypeScript types are properly defined (no `any`)

---

## 🧪 Verification Plan

### Manual Testing
1. Navigate to `/projects/new` → Verify wizard loads
2. Resize browser to tablet width → Verify stacked layout
3. Test keyboard navigation through form fields
4. Click close button → Verify navigation to `/projects`

### Lint/Type Check
```bash
cd frontend && npm run lint && npm run typecheck
```

### Unit Tests (if time permits)
Create test file: `frontend/src/components/views/fb-view-onboarding.test.ts`
- Test that component renders without errors
- Test form validation logic

---

## 📚 Reference Documents

- [PHASE_11_PRD.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/planning/PHASE_11_PRD.md) - Full PRD
- [FRONTEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_SCOPE.md) - Component patterns
- [MASTER_PRD.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/MASTER_PRD.md) Section 7.3 - Quick-Add Wizard reference
- [PRODUCT_VISION.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/PRODUCT_VISION.md) - System architecture
- [fb-view-chat.ts](file:///home/colton/Desktop/FutureBuild_HQ/dev/frontend/src/components/views/fb-view-chat.ts) - Chat view pattern
- [fb-view-projects.ts](file:///home/colton/Desktop/FutureBuild_HQ/dev/frontend/src/components/views/fb-view-projects.ts) - Projects view pattern

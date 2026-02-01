# Step 76: Real-Time Form Filling (State Sync)

**Phase:** 11 (The Conversational Hook)  
**Status:** READY FOR IMPLEMENTATION  
**Est. Duration:** 1 Day  
**Owner:** Frontend Developer

---

## 🎯 Objective

Implement bidirectional state synchronization between the chat panel and form panel in the onboarding wizard. Form fields must update in real-time as the AI extracts data, with visual indicators distinguishing AI-populated vs user-edited values.

---

## 📐 Architectural Guardrails

> [!IMPORTANT]
> These constraints are non-negotiable and align with the system architecture.

### State Management Pattern
Reference: [FRONTEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_SCOPE.md) Section 5

- Use **Signals** (@lit-labs/preact-signals) for reactive state
- State lives in a shared store, NOT duplicated across components
- Components subscribe to state slices, not the entire store
- Avoid prop-drilling; use event bubbling + store reads

### Data Flow Rules
```
┌─────────────────────────────────────────────────────────────────┐
│                       OnboardingStore                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ values$      │  │ sources$     │  │ confidence$  │          │
│  │ (form data)  │  │ (ai/user)    │  │ (0.0-1.0)    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
         ▲                    ▲                    ▲
         │                    │                    │
    ┌────┴────┐          ┌────┴────┐          ┌────┴────┐
    │ Chat    │          │ Form    │          │ API     │
    │ Panel   │          │ Panel   │          │ Response│
    └─────────┘          └─────────┘          └─────────┘
```

### Visual States
Reference: [PHASE_11_PRD.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/planning/PHASE_11_PRD.md) Section 5.3

| Source | Visual Treatment |
|--------|------------------|
| `ai` | Blue left border + sparkle icon ✨ + "(AI)" badge |
| `user` | Normal styling (clears AI indicator on edit) |
| `default` | Muted text + "(Default)" label |
| Low confidence (<0.8) | Yellow left border + "Verify" badge |

---

## 📋 Implementation Checklist

### Files to Create/Modify

```
frontend/src/
├── store/
│   └── onboarding-store.ts          # Dedicated store for wizard state
├── components/features/onboarding/
│   ├── fb-onboarding-form.ts        # Modify to use store
│   └── fb-onboarding-chat.ts        # Modify to dispatch updates
└── styles/
    └── onboarding.css               # Field state styles
```

---

### Task 1: Create Onboarding Store (45 min)

```typescript
// frontend/src/store/onboarding-store.ts

import { signal, computed } from '@preact/signals-core';
import type { CreateProjectRequest } from '../types/models';

// Core state signals
export const onboardingValues = signal<Partial<CreateProjectRequest>>({});
export const onboardingSources = signal<Record<string, 'user' | 'ai' | 'default'>>({});
export const onboardingConfidence = signal<Record<string, number>>({});
export const onboardingMessages = signal<OnboardingMessage[]>([]);
export const isProcessing = signal<boolean>(false);

// Computed: Check if ready to create
export const isReadyToCreate = computed(() => {
  const v = onboardingValues.value;
  return !!(v.name && v.address);
});

// Computed: Get fields that need verification
export const fieldsNeedingVerification = computed(() => {
  const conf = onboardingConfidence.value;
  return Object.entries(conf)
    .filter(([_, score]) => score < 0.8)
    .map(([field, _]) => field);
});

// Action: Update a single field (from user input)
export function setFieldValue(field: keyof CreateProjectRequest, value: any): void {
  onboardingValues.value = {
    ...onboardingValues.value,
    [field]: value
  };
  onboardingSources.value = {
    ...onboardingSources.value,
    [field]: 'user'
  };
  // Clear confidence when user manually edits
  const newConf = { ...onboardingConfidence.value };
  delete newConf[field];
  onboardingConfidence.value = newConf;
}

// Action: Apply AI extraction results (from API response)
export function applyAIExtraction(
  extractedValues: Record<string, any>,
  confidenceScores: Record<string, number>
): void {
  const currentValues = { ...onboardingValues.value };
  const currentSources = { ...onboardingSources.value };
  const currentConf = { ...onboardingConfidence.value };

  for (const [field, value] of Object.entries(extractedValues)) {
    // Only apply if field is empty OR was previously AI-populated
    const existingSource = currentSources[field];
    if (!currentValues[field] || existingSource === 'ai' || existingSource === 'default') {
      currentValues[field as keyof CreateProjectRequest] = value;
      currentSources[field] = 'ai';
      currentConf[field] = confidenceScores[field] ?? 0.5;
    }
  }

  onboardingValues.value = currentValues;
  onboardingSources.value = currentSources;
  onboardingConfidence.value = currentConf;
}

// Action: Add message to conversation
export function addMessage(message: OnboardingMessage): void {
  onboardingMessages.value = [...onboardingMessages.value, message];
}

// Action: Reset wizard state
export function resetOnboarding(): void {
  onboardingValues.value = {};
  onboardingSources.value = {};
  onboardingConfidence.value = {};
  onboardingMessages.value = [];
  isProcessing.value = false;
}

// Types
export interface OnboardingMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
}
```

---

### Task 2: Connect Chat Panel to Store (30 min)

```typescript
// frontend/src/components/features/onboarding/fb-onboarding-chat.ts

import { SignalWatcher } from '@lit-labs/preact-signals';
import { 
  onboardingMessages, 
  isProcessing,
  addMessage,
  applyAIExtraction
} from '../../../store/onboarding-store';
import { api } from '../../../services/api';

@customElement('fb-onboarding-chat')
export class FBOnboardingChat extends SignalWatcher(LitElement) {
  
  async _handleSend(e: CustomEvent<{ content: string }>) {
    const content = e.detail.content;
    
    // Add user message to store
    addMessage({
      id: crypto.randomUUID(),
      role: 'user',
      content,
      timestamp: new Date()
    });
    
    // Call API
    isProcessing.value = true;
    try {
      const response = await api.onboard.process({
        session_id: this._sessionId,
        message: content,
        current_state: onboardingValues.value
      });
      
      // Apply AI extractions to store (triggers form update)
      if (response.extracted_values) {
        applyAIExtraction(
          response.extracted_values,
          response.confidence_scores || {}
        );
      }
      
      // Add assistant message
      addMessage({
        id: crypto.randomUUID(),
        role: 'assistant',
        content: response.reply,
        timestamp: new Date()
      });
      
    } catch (error) {
      addMessage({
        id: crypto.randomUUID(),
        role: 'system',
        content: 'Something went wrong. Please try again.',
        timestamp: new Date()
      });
    } finally {
      isProcessing.value = false;
    }
  }

  render() {
    return html`
      <fb-message-list .messages=${onboardingMessages.value}></fb-message-list>
      <fb-onboarding-dropzone @file-dropped=${this._handleFileDrop}></fb-onboarding-dropzone>
      <fb-input-bar
        ?disabled=${isProcessing.value}
        @send=${this._handleSend}
      ></fb-input-bar>
    `;
  }
}
```

---

### Task 3: Connect Form Panel to Store (45 min)

```typescript
// frontend/src/components/features/onboarding/fb-onboarding-form.ts

import { SignalWatcher } from '@lit-labs/preact-signals';
import {
  onboardingValues,
  onboardingSources,
  onboardingConfidence,
  isReadyToCreate,
  fieldsNeedingVerification,
  setFieldValue
} from '../../../store/onboarding-store';

@customElement('fb-onboarding-form')
export class FBOnboardingForm extends SignalWatcher(LitElement) {
  
  static override styles = css`
    .field {
      margin-bottom: var(--fb-spacing-md);
      position: relative;
    }

    .field.ai-populated {
      border-left: 3px solid var(--fb-primary);
      padding-left: var(--fb-spacing-sm);
    }

    .field.needs-verification {
      border-left: 3px solid var(--fb-warning);
      padding-left: var(--fb-spacing-sm);
    }

    .ai-badge {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      font-size: var(--fb-text-xs);
      color: var(--fb-primary);
      margin-left: var(--fb-spacing-xs);
    }

    .verify-badge {
      display: inline-flex;
      align-items: center;
      font-size: var(--fb-text-xs);
      color: var(--fb-warning);
      background: rgba(230, 81, 0, 0.1);
      padding: 2px 6px;
      border-radius: var(--fb-radius-sm);
      margin-left: var(--fb-spacing-xs);
    }

    /* Animation for AI-populated fields */
    @keyframes ai-glow {
      0% { box-shadow: 0 0 0 0 rgba(102, 126, 234, 0.4); }
      70% { box-shadow: 0 0 0 6px rgba(102, 126, 234, 0); }
      100% { box-shadow: 0 0 0 0 rgba(102, 126, 234, 0); }
    }

    .field.just-populated input {
      animation: ai-glow 0.6s ease-out;
    }
  `;

  private _renderField(
    name: keyof CreateProjectRequest,
    label: string,
    type: 'text' | 'number' | 'date' | 'select',
    options?: string[]
  ) {
    const value = onboardingValues.value[name];
    const source = onboardingSources.value[name];
    const confidence = onboardingConfidence.value[name];
    const needsVerify = confidence !== undefined && confidence < 0.8;

    const classes = {
      field: true,
      'ai-populated': source === 'ai',
      'needs-verification': needsVerify
    };

    return html`
      <div class=${classMap(classes)}>
        <label for=${name}>
          ${label}
          ${source === 'ai' ? html`<span class="ai-badge">✨ AI</span>` : ''}
          ${needsVerify ? html`<span class="verify-badge">Verify</span>` : ''}
        </label>
        
        ${type === 'select' && options 
          ? html`
              <select
                id=${name}
                .value=${value ?? ''}
                @change=${(e: Event) => this._handleChange(name, (e.target as HTMLSelectElement).value)}
              >
                <option value="">Select...</option>
                ${options.map(opt => html`<option value=${opt}>${opt}</option>`)}
              </select>
            `
          : html`
              <input
                id=${name}
                type=${type}
                .value=${value ?? ''}
                @input=${(e: Event) => this._handleChange(name, (e.target as HTMLInputElement).value)}
              />
            `
        }
      </div>
    `;
  }

  private _handleChange(field: keyof CreateProjectRequest, value: any) {
    // Convert to appropriate type
    const typedValue = this._parseValue(field, value);
    setFieldValue(field, typedValue);
  }

  private _parseValue(field: string, value: string): any {
    const numericFields = ['square_footage', 'bedrooms', 'bathrooms', 'stories'];
    if (numericFields.includes(field)) {
      const num = parseFloat(value);
      return isNaN(num) ? undefined : num;
    }
    return value || undefined;
  }

  private async _handleSubmit(e: Event) {
    e.preventDefault();
    if (!isReadyToCreate.value) return;

    // Emit create event with form values
    this.dispatchEvent(new CustomEvent('create-project', {
      bubbles: true,
      composed: true,
      detail: { values: onboardingValues.value }
    }));
  }

  render() {
    return html`
      <form @submit=${this._handleSubmit}>
        <section>
          <h3>Project Details</h3>
          ${this._renderField('name', 'Project Name', 'text')}
          ${this._renderField('address', 'Address', 'text')}
        </section>

        <section>
          <h3>Building Specifications</h3>
          <p class="hint">These help calculate your schedule accurately.</p>
          ${this._renderField('square_footage', 'Square Feet', 'number')}
          ${this._renderField('foundation_type', 'Foundation Type', 'select', 
            ['slab', 'crawlspace', 'basement'])}
          ${this._renderField('stories', 'Stories', 'number')}
          ${this._renderField('bedrooms', 'Bedrooms', 'number')}
          ${this._renderField('bathrooms', 'Bathrooms', 'number')}
          ${this._renderField('topography', 'Topography', 'select',
            ['flat', 'moderate', 'steep'])}
        </section>

        ${fieldsNeedingVerification.value.length > 0 ? html`
          <div class="verification-notice">
            ⚠️ Please verify the highlighted fields before creating.
          </div>
        ` : ''}

        <button
          type="submit"
          class="btn-primary"
          ?disabled=${!isReadyToCreate.value}
        >
          Create Project
        </button>
      </form>
    `;
  }
}
```

---

### Task 4: Implement Glow Animation (15 min)

Track "just populated" state to trigger animation:

```typescript
// In store, track recently updated fields
export const recentlyUpdatedFields = signal<Set<string>>(new Set());

export function applyAIExtraction(
  extractedValues: Record<string, any>,
  confidenceScores: Record<string, number>
): void {
  // ... existing logic ...
  
  // Mark fields as recently updated for animation
  recentlyUpdatedFields.value = new Set(Object.keys(extractedValues));
  
  // Clear after animation duration
  setTimeout(() => {
    recentlyUpdatedFields.value = new Set();
  }, 600);
}
```

---

## 🔌 Interface Contracts

### Store Actions

| Action | Params | Effect |
|--------|--------|--------|
| `setFieldValue` | `(field, value)` | Updates value, sets source to 'user', clears confidence |
| `applyAIExtraction` | `(values, confidence)` | Bulk update with 'ai' source, triggers animation |
| `addMessage` | `(message)` | Appends to message history |
| `resetOnboarding` | `()` | Clears all wizard state |

### Events Emitted by Form

| Event | Payload | When |
|-------|---------|------|
| `create-project` | `{ values: CreateProjectRequest }` | Form submitted |
| `field-verified` | `{ field: string }` | User edits AI-populated field |

---

## ✅ Acceptance Criteria

- [ ] Form fields update immediately when AI extracts data (no page refresh)
- [ ] AI-populated fields show blue left border + ✨ AI badge
- [ ] Low-confidence fields (<0.8) show yellow border + "Verify" badge  
- [ ] Editing an AI-populated field clears AI indicator (becomes user source)
- [ ] Editing an AI-populated field does NOT re-apply on next AI response
- [ ] New AI extractions animate with glow effect (0.6s)
- [ ] Chat panel shows processing state during API call
- [ ] Form validation prevents submit until name + address filled
- [ ] All state changes are reactive (no manual `requestUpdate()` calls)

---

## 🧪 Verification Plan

### Manual Testing
1. Type message "3 bedroom home" → Verify bedrooms field updates
2. Upload blueprint → Verify multiple fields populate with animation
3. Edit AI-populated GSF → Verify AI badge disappears
4. Clear name field → Verify Create button disables

### Console Verification
```javascript
// In browser console, test store reactivity
import { onboardingValues, applyAIExtraction } from '/src/store/onboarding-store.ts';
applyAIExtraction({ gsf: 2500, bedrooms: 4 }, { gsf: 0.95, bedrooms: 0.72 });
console.log(onboardingValues.value); // Should show updated values
```

---

## 📚 Reference Documents

- [PHASE_11_PRD.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/planning/PHASE_11_PRD.md) Section 5.3
- [FRONTEND_SCOPE.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_SCOPE.md) Section 5 (State Management)
- [FRONTEND_TYPES_SPEC.md](file:///home/colton/Desktop/FutureBuild_HQ/dev/specs/FRONTEND_TYPES_SPEC.md) - Type definitions

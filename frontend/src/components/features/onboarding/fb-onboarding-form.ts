/**
 * FBOnboardingForm - Right Panel Live Form for Project Onboarding
 * See STEP_74_SPLIT_SCREEN_WIZARD.md Task 3
 *
 * Live form that updates as AI extracts data from conversations and documents.
 * Implements visual indicators for AI vs user-populated fields:
 * - Blue left border + ✨ badge for AI-populated fields
 * - Yellow "Verify" badge for low-confidence fields (< 0.8)
 * - User edits override AI values and remove visual indicators
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';
import { api } from '../../../services/api';
import type { CreateProjectRequest } from '../../../services/api';

@customElement('fb-onboarding-form')
export class FBOnboardingForm extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
            }

            form {
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-xl, 24px);
            }

            section {
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-md, 16px);
            }

            h3 {
                font-size: var(--fb-text-lg, 18px);
                font-weight: 600;
                color: var(--fb-text-primary, #1a1a1a);
                margin: 0;
                padding-bottom: var(--fb-spacing-sm, 8px);
                border-bottom: 1px solid var(--fb-border, #e5e5e5);
            }

            .hint {
                font-size: var(--fb-text-sm, 14px);
                color: var(--fb-text-muted, #666);
                margin: 0;
                font-style: italic;
            }

            .field {
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-xs, 4px);
                position: relative;
            }

            .field-header {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-xs, 4px);
            }

            label {
                font-size: var(--fb-text-sm, 14px);
                font-weight: 500;
                color: var(--fb-text-primary, #1a1a1a);
            }

            .ai-badge {
                display: inline-flex;
                align-items: center;
                gap: 2px;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                font-size: 11px;
                font-weight: 600;
                padding: 2px 6px;
                border-radius: 4px;
                text-transform: uppercase;
                letter-spacing: 0.5px;
            }

            .verify-badge {
                display: inline-flex;
                align-items: center;
                background: #fbbf24;
                color: #78350f;
                font-size: 11px;
                font-weight: 600;
                padding: 2px 6px;
                border-radius: 4px;
                text-transform: uppercase;
                letter-spacing: 0.5px;
            }

            input,
            select {
                font-size: var(--fb-text-sm, 14px);
                padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 16px);
                border: 1px solid var(--fb-border, #e5e5e5);
                border-radius: var(--fb-radius-sm, 4px);
                background: var(--fb-bg-card, white);
                color: var(--fb-text-primary, #1a1a1a);
                transition: all 0.2s;
            }

            input:focus,
            select:focus {
                outline: none;
                border-color: #667eea;
                box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
            }

            /* AI-populated field styling */
            .field.ai-populated input,
            .field.ai-populated select {
                border-left: 3px solid #667eea;
                padding-left: calc(var(--fb-spacing-md, 16px) - 2px);
                background: linear-gradient(90deg, rgba(102, 126, 234, 0.05) 0%, var(--fb-bg-card, white) 100%);
            }

            /* Low-confidence field styling */
            .field.low-confidence input,
            .field.low-confidence select {
                border-left: 3px solid #fbbf24;
                padding-left: calc(var(--fb-spacing-md, 16px) - 2px);
                background: linear-gradient(90deg, rgba(251, 191, 36, 0.05) 0%, var(--fb-bg-card, white) 100%);
            }

            .btn-primary {
                display: inline-flex;
                align-items: center;
                justify-content: center;
                padding: var(--fb-spacing-md, 16px) var(--fb-spacing-xl, 24px);
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                font-size: var(--fb-text-md, 16px);
                font-weight: 600;
                border: none;
                border-radius: var(--fb-radius-md, 8px);
                cursor: pointer;
                transition: all 0.2s;
                margin-top: var(--fb-spacing-md, 16px);
            }

            .btn-primary:hover:not(:disabled) {
                transform: translateY(-1px);
                box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
            }

            .btn-primary:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .error-message {
                color: #dc2626;
                font-size: var(--fb-text-sm, 14px);
                margin-top: var(--fb-spacing-xs, 4px);
            }

            .success-message {
                color: #16a34a;
                font-size: var(--fb-text-sm, 14px);
                margin-top: var(--fb-spacing-xs, 4px);
            }

            @media (max-width: 767px) {
                form {
                    gap: var(--fb-spacing-md, 16px);
                }
            }
        `
    ];

    @property({ type: Object }) values: Partial<CreateProjectRequest> = {};
    @property({ type: Object }) sources: Record<string, 'user' | 'ai' | 'default'> = {};
    @property({ type: Object }) confidence: Record<string, number> = {};

    @state() private _errorMessage = '';
    @state() private _isSubmitting = false;

    private _handleInput(field: keyof CreateProjectRequest, e: Event): void {
        const input = e.target as HTMLInputElement | HTMLSelectElement;
        const value = input.type === 'number' ? Number(input.value) : input.value;

        // Update values
        const updatedValues = { ...this.values, [field]: value };

        // Mark as user-edited (removes AI styling)
        const updatedSources = { ...this.sources, [field]: 'user' as const };

        // Remove confidence for user-edited fields
        const updatedConfidence = { ...this.confidence };
        delete updatedConfidence[field];

        // Emit update event
        this.emit('form-updated', {
            values: updatedValues,
            sources: updatedSources,
            confidence: updatedConfidence
        });
    }

    private async _handleSubmit(e: Event): Promise<void> {
        e.preventDefault();
        this._errorMessage = '';

        if (!this._canCreate) {
            this._errorMessage = 'Please fill in all required fields (Name and Address)';
            return;
        }

        this._isSubmitting = true;

        try {
            // Create the project via API
            const projectData: CreateProjectRequest = {
                name: this.values.name!,
                address: this.values.address!,
                square_footage: this.values.square_footage ?? 0,
                bedrooms: this.values.bedrooms ?? 0,
                bathrooms: this.values.bathrooms ?? 0,
                lot_size: this.values.lot_size,
                foundation_type: this.values.foundation_type,
                start_date: this.values.start_date ?? new Date().toISOString().split('T')[0]
            };

            const response = await api.projects.create(projectData);

            // Emit success event
            this.emit('project-created', { projectId: response.id });
        } catch (err) {
            console.error('[FBOnboardingForm] Project creation failed:', err);
            this._errorMessage = 'Failed to create project. Please try again.';
        } finally {
            this._isSubmitting = false;
        }
    }

    private get _canCreate(): boolean {
        return !!(this.values.name && this.values.address);
    }

    private _renderField(
        name: keyof CreateProjectRequest,
        label: string,
        type: 'text' | 'number' | 'date' | 'select',
        options?: { value: string; label: string }[],
        required = false
    ): TemplateResult {
        const source = this.sources[name];
        const conf = this.confidence[name];
        const isAiPopulated = source === 'ai';
        const isLowConfidence = isAiPopulated && conf !== undefined && conf < 0.8;

        const fieldClass = isLowConfidence ? 'low-confidence' : (isAiPopulated ? 'ai-populated' : '');

        return html`
            <div class="field ${fieldClass}">
                <div class="field-header">
                    <label>${label}${required ? ' *' : ''}</label>
                    ${isAiPopulated ? html`<span class="ai-badge">✨ AI</span>` : ''}
                    ${isLowConfidence ? html`<span class="verify-badge">⚠️ Verify</span>` : ''}
                </div>
                ${type === 'select' && options ? html`
                    <select
                        .value=${String(this.values[name] ?? '')}
                        @input=${(e: Event) => this._handleInput(name, e)}
                    >
                        <option value="">-- Select --</option>
                        ${options.map(opt => html`
                            <option value=${opt.value}>${opt.label}</option>
                        `)}
                    </select>
                ` : html`
                    <input
                        type=${type}
                        .value=${String(this.values[name] ?? '')}
                        @input=${(e: Event) => this._handleInput(name, e)}
                        ?required=${required}
                    />
                `}
            </div>
        `;
    }

    override render(): TemplateResult {
        return html`
            <form @submit=${this._handleSubmit}>
                <section class="required-fields">
                    <h3>Project Details</h3>
                    ${this._renderField('name', 'Project Name', 'text', undefined, true)}
                    ${this._renderField('address', 'Address', 'text', undefined, true)}
                    ${this._renderField('start_date', 'Start Date', 'date')}
                </section>

                <section class="physics-fields">
                    <h3>Building Specifications</h3>
                    <p class="hint">These help calculate your schedule accurately.</p>
                    ${this._renderField('square_footage', 'Square Feet', 'number')}
                    ${this._renderField('bedrooms', 'Bedrooms', 'number')}
                    ${this._renderField('bathrooms', 'Bathrooms', 'number')}
                    ${this._renderField('lot_size', 'Lot Size (sq ft)', 'number')}
                    ${this._renderField('foundation_type', 'Foundation Type', 'select', [
                        { value: 'slab', label: 'Slab' },
                        { value: 'crawlspace', label: 'Crawlspace' },
                        { value: 'basement', label: 'Basement' }
                    ])}
                </section>

                ${this._errorMessage ? html`
                    <div class="error-message">${this._errorMessage}</div>
                ` : ''}

                <button
                    type="submit"
                    class="btn-primary"
                    ?disabled=${!this._canCreate || this._isSubmitting}
                >
                    ${this._isSubmitting ? 'Creating...' : 'Create Project'}
                </button>
            </form>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboarding-form': FBOnboardingForm;
    }
}

/**
 * FBOnboardingForm - Right Panel Live Form for Project Onboarding
 * See STEP_74_SPLIT_SCREEN_WIZARD.md Task 3, STEP_76_REALTIME_FORM_FILLING.md
 *
 * Live form that updates as AI extracts data from conversations and documents.
 * Reads directly from onboarding-store signals (no prop-drilling).
 * Implements visual indicators for AI vs user-populated fields:
 * - Blue left border + AI badge for AI-populated fields
 * - Yellow "Verify" badge for low-confidence fields (< 0.8)
 * - Glow animation on newly AI-populated fields
 * - User edits override AI values and remove visual indicators
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { FBElement } from '../../base/FBElement';
import { api } from '../../../services/api';
import type { CreateProjectRequest } from '../../../services/api';
import {
    onboardingValues,
    onboardingSources,
    onboardingConfidence,
    recentlyUpdatedFields,
    isReadyToCreate,
    isProcessing,
    fieldsNeedingVerification,
    setFieldValue,
} from '../../../store/onboarding-store';

@customElement('fb-onboarding-form')
export class FBOnboardingForm extends SignalWatcher(FBElement) {
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

            /* Glow animation for newly AI-populated fields */
            .field.just-populated input,
            .field.just-populated select {
                animation: ai-glow 0.6s ease-out;
            }

            @keyframes ai-glow {
                0% {
                    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0.5);
                }
                50% {
                    box-shadow: 0 0 8px 4px rgba(102, 126, 234, 0.3);
                }
                100% {
                    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0);
                }
            }

            .verification-notice {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm, 8px);
                padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 16px);
                background: rgba(251, 191, 36, 0.1);
                border: 1px solid #fbbf24;
                border-radius: var(--fb-radius-sm, 4px);
                font-size: var(--fb-text-sm, 14px);
                color: #78350f;
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

            @media (max-width: 767px) {
                form {
                    gap: var(--fb-spacing-md, 16px);
                }
            }
        `
    ];

    @state() private _errorMessage = '';
    @state() private _isSubmitting = false;

    private _handleInput(field: keyof CreateProjectRequest, e: Event): void {
        const input = e.target as HTMLInputElement | HTMLSelectElement;
        const value = input.type === 'number' ? Number(input.value) : input.value;
        setFieldValue(field, value);
    }

    private async _handleSubmit(e: Event): Promise<void> {
        e.preventDefault();
        this._errorMessage = '';

        // Fix 8: Guard against submit while AI is processing or already submitting
        if (isProcessing.value || this._isSubmitting) return;

        if (!isReadyToCreate.value) {
            this._errorMessage = 'Please fill in all required fields (Name and Address)';
            return;
        }

        this._isSubmitting = true;
        const values = onboardingValues.value;
        const name = values.name ?? '';
        const address = values.address ?? '';

        try {
            const startDate = values.start_date ?? new Date().toISOString().split('T')[0] ?? '';
            const projectData: CreateProjectRequest = {
                name,
                address,
                square_footage: values.square_footage ?? 0,
                bedrooms: values.bedrooms ?? 0,
                bathrooms: values.bathrooms ?? 0,
                start_date: startDate,
            };
            if (values.lot_size !== undefined) projectData.lot_size = values.lot_size;
            if (values.foundation_type !== undefined) projectData.foundation_type = values.foundation_type;
            if (values.stories !== undefined) projectData.stories = values.stories;
            if (values.topography !== undefined) projectData.topography = values.topography;
            if (values.soil_conditions !== undefined) projectData.soil_conditions = values.soil_conditions;

            const response = await api.projects.create(projectData);
            this.emit('project-created', { projectId: response.id });
        } catch (err) {
            console.error('[FBOnboardingForm] Project creation failed:', err);
            this._errorMessage = 'Failed to create project. Please try again.';
        } finally {
            this._isSubmitting = false;
        }
    }

    private _renderField(
        name: keyof CreateProjectRequest,
        label: string,
        type: 'text' | 'number' | 'date' | 'select',
        options?: { value: string; label: string }[],
        required = false
    ): TemplateResult {
        const source = onboardingSources.value[name];
        const conf = onboardingConfidence.value[name];
        const isAiPopulated = source === 'ai';
        const isLowConfidence = isAiPopulated && conf !== undefined && conf < 0.8;
        const isJustPopulated = recentlyUpdatedFields.value.has(name);

        const classes = [
            'field',
            isLowConfidence ? 'low-confidence' : (isAiPopulated ? 'ai-populated' : ''),
            isJustPopulated ? 'just-populated' : '',
        ].filter(Boolean).join(' ');

        return html`
            <div class=${classes}>
                <div class="field-header">
                    <label>${label}${required ? ' *' : ''}</label>
                    ${isAiPopulated ? html`<span class="ai-badge">AI</span>` : ''}
                    ${isLowConfidence ? html`<span class="verify-badge">Verify</span>` : ''}
                </div>
                ${type === 'select' && options ? html`
                    <select
                        .value=${String(onboardingValues.value[name] ?? '')}
                        @input=${(e: Event): void => { this._handleInput(name, e); }}
                    >
                        <option value="">-- Select --</option>
                        ${options.map(opt => html`
                            <option value=${opt.value}>${opt.label}</option>
                        `)}
                    </select>
                ` : html`
                    <input
                        type=${type}
                        .value=${String(onboardingValues.value[name] ?? '')}
                        @input=${(e: Event): void => { this._handleInput(name, e); }}
                        ?required=${required}
                    />
                `}
            </div>
        `;
    }

    override render(): TemplateResult {
        const needsVerification = fieldsNeedingVerification.value;

        return html`
            ${needsVerification.length > 0 ? html`
                <div class="verification-notice">
                    Some fields have low AI confidence and may need your review.
                </div>
            ` : ''}

            <form @submit=${(e: Event): void => { void this._handleSubmit(e); }}>
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
                    ${this._renderField('stories', 'Stories', 'number')}
                    ${this._renderField('lot_size', 'Lot Size (sq ft)', 'number')}
                    ${this._renderField('foundation_type', 'Foundation Type', 'select', [
                        { value: 'slab', label: 'Slab' },
                        { value: 'crawlspace', label: 'Crawlspace' },
                        { value: 'basement', label: 'Basement' }
                    ])}
                    ${this._renderField('topography', 'Topography', 'select', [
                        { value: 'flat', label: 'Flat' },
                        { value: 'sloped', label: 'Sloped' },
                        { value: 'hillside', label: 'Hillside' }
                    ])}
                    ${this._renderField('soil_conditions', 'Soil Conditions', 'select', [
                        { value: 'normal', label: 'Normal' },
                        { value: 'rocky', label: 'Rocky' },
                        { value: 'clay', label: 'Clay' },
                        { value: 'sandy', label: 'Sandy' }
                    ])}
                </section>

                ${this._errorMessage ? html`
                    <div class="error-message">${this._errorMessage}</div>
                ` : ''}

                <button
                    type="submit"
                    class="btn-primary"
                    ?disabled=${!isReadyToCreate.value || this._isSubmitting || isProcessing.value}
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

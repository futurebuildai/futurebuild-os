/**
 * FBOnboardingSteps - Horizontal Progress Bar for Project Onboarding
 *
 * Displays onboarding progress as a horizontal step indicator:
 * Upload → Extract → Details → Review
 *
 * - Active step highlighted with accent color and pulse animation
 * - Completed steps show checkmarks
 * - Compact height (~48px) to maximize chat space
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { FBElement } from '../../base/FBElement';
import { currentStage, type OnboardingStage } from '../../../store/onboarding-store';

interface StepConfig {
    id: OnboardingStage;
    label: string;
    icon: string;
}

const STEPS: StepConfig[] = [
    { id: 'upload', label: 'Upload', icon: 'upload' },
    { id: 'extract', label: 'Extract', icon: 'sparkles' },
    { id: 'details', label: 'Details', icon: 'chat' },
    { id: 'review', label: 'Review', icon: 'check-circle' },
];

@customElement('fb-onboarding-steps')
export class FBOnboardingSteps extends SignalWatcher(FBElement) {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                align-items: center;
                justify-content: center;
                padding: var(--fb-spacing-sm) var(--fb-spacing-xl);
                background: var(--md-sys-color-surface-container-low);
                border-bottom: 1px solid var(--md-sys-color-outline-variant);
                min-height: 56px;
                box-shadow: var(--md-sys-elevation-1);
                position: relative;
                z-index: 5;
            }

            .steps {
                display: flex;
                align-items: center;
                gap: 0;
            }

            .step {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-xs);
                position: relative;
            }

            .step-content {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                padding: 6px 12px;
                border-radius: var(--md-sys-shape-corner-full);
                transition: all var(--fb-transition-base);
            }

            .step-indicator {
                width: 28px;
                height: 28px;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
                font: var(--md-sys-typescale-label-small);
                transition: all var(--fb-transition-base);
                flex-shrink: 0;
            }

            .step-indicator svg {
                width: 16px;
                height: 16px;
                fill: currentColor;
            }

            .step-label {
                font: var(--md-sys-typescale-label-medium);
                white-space: nowrap;
                transition: color var(--fb-transition-base);
            }

            /* Completed state */
            .step.completed .step-indicator {
                background-color: var(--md-sys-color-primary);
                color: var(--md-sys-color-on-primary);
            }

            .step.completed .step-label {
                color: var(--md-sys-color-on-surface);
            }

            /* Active state */
            .step.active .step-content {
                background-color: var(--md-sys-color-primary-container);
            }

            .step.active .step-indicator {
                background-color: var(--md-sys-color-primary);
                color: var(--md-sys-color-on-primary);
                box-shadow: 0 0 0 2px var(--md-sys-color-background), 0 0 0 4px var(--md-sys-color-primary-container);
            }

            .step.active .step-label {
                color: var(--md-sys-color-on-primary-container);
                font-weight: 500;
            }

            /* Pending state */
            .step.pending .step-indicator {
                background-color: var(--md-sys-color-surface-variant);
                color: var(--md-sys-color-on-surface-variant);
            }

            .step.pending .step-label {
                color: var(--md-sys-color-on-surface-variant);
            }

            /* Connector line between steps */
            .connector {
                width: 40px;
                height: 2px;
                background-color: var(--md-sys-color-surface-variant);
                transition: background-color var(--fb-transition-base);
                margin: 0 var(--fb-spacing-xs);
                border-radius: 1px;
            }

            .connector.completed {
                background-color: var(--md-sys-color-primary);
            }

            /* Responsive: Hide labels on small screens */
            @media (max-width: 600px) {
                .step-label {
                    display: none;
                }

                .connector {
                    width: 24px;
                }
            }
        `
    ];

    private _getStepState(stepId: OnboardingStage): 'completed' | 'active' | 'pending' {
        const current = currentStage.value;
        const currentIndex = STEPS.findIndex(s => s.id === current);
        const stepIndex = STEPS.findIndex(s => s.id === stepId);

        if (stepIndex < currentIndex) return 'completed';
        if (stepIndex === currentIndex) return 'active';
        return 'pending';
    }

    private _renderIcon(step: StepConfig, state: 'completed' | 'active' | 'pending'): TemplateResult {
        // Show checkmark for completed steps
        if (state === 'completed') {
            return html`
                <svg viewBox="0 0 24 24" fill="currentColor">
                    <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                </svg>
            `;
        }

        // Show step-specific icon
        switch (step.icon) {
            case 'upload':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M9 16h6v-6h4l-7-7-7 7h4zm-4 2h14v2H5z"/>
                    </svg>
                `;
            case 'sparkles':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2L9.5 8.5 2 12l7.5 3.5L12 22l2.5-6.5L22 12l-7.5-3.5z"/>
                    </svg>
                `;
            case 'chat':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z"/>
                    </svg>
                `;
            case 'check-circle':
                return html`
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                    </svg>
                `;
            default:
                return html``;
        }
    }

    private _renderStep(step: StepConfig, index: number): TemplateResult {
        const state = this._getStepState(step.id);
        const prevStep = index > 0 ? STEPS[index - 1] : undefined;
        const prevState = prevStep ? this._getStepState(prevStep.id) : 'pending';

        return html`
            ${index > 0 ? html`
                <div class="connector ${prevState === 'completed' ? 'completed' : ''}"></div>
            ` : nothing}
            <div class="step ${state}">
                <div class="step-content">
                    <div class="step-indicator">
                        ${this._renderIcon(step, state)}
                    </div>
                    <span class="step-label">${step.label}</span>
                </div>
            </div>
        `;
    }

    override render(): TemplateResult {
        return html`
            <div class="steps">
                ${STEPS.map((step, i) => this._renderStep(step, i))}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboarding-steps': FBOnboardingSteps;
    }
}

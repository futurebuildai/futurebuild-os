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
                background: var(--fb-bg-secondary);
                border-bottom: 1px solid var(--fb-border);
                min-height: 48px;
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
            }

            .step-content {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-xs);
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                border-radius: var(--fb-radius-full);
                transition: all 0.3s ease;
            }

            .step-indicator {
                width: 24px;
                height: 24px;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 12px;
                font-weight: 600;
                transition: all 0.3s ease;
                flex-shrink: 0;
            }

            .step-indicator svg {
                width: 14px;
                height: 14px;
            }

            .step-label {
                font-size: var(--fb-text-sm);
                font-weight: 500;
                white-space: nowrap;
                transition: color 0.3s ease;
            }

            /* Completed state */
            .step.completed .step-indicator {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
            }

            .step.completed .step-label {
                color: var(--fb-text-primary);
            }

            /* Active state */
            .step.active .step-content {
                background: rgba(102, 126, 234, 0.1);
            }

            .step.active .step-indicator {
                border: 2px solid #667eea;
                background: white;
                color: #667eea;
                animation: pulse 2s ease-in-out infinite;
            }

            .step.active .step-label {
                color: #667eea;
                font-weight: 600;
            }

            /* Pending state */
            .step.pending .step-indicator {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-muted);
            }

            .step.pending .step-label {
                color: var(--fb-text-muted);
            }

            /* Connector line between steps */
            .connector {
                width: 32px;
                height: 2px;
                background: var(--fb-border);
                transition: background 0.3s ease;
                margin: 0 var(--fb-spacing-xs);
            }

            .connector.completed {
                background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
            }

            @keyframes pulse {
                0%, 100% {
                    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0.4);
                }
                50% {
                    box-shadow: 0 0 0 6px rgba(102, 126, 234, 0);
                }
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

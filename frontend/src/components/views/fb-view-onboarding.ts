/**
 * FBViewOnboarding - Split-Screen Wizard for AI-Driven Project Creation
 * See STEP_74_SPLIT_SCREEN_WIZARD.md, STEP_76_REALTIME_FORM_FILLING.md
 *
 * Pure layout component. State lives in onboarding-store; children read from it directly.
 * - Left Panel: Chat interface with "The Interrogator" agent
 * - Right Panel: Live form that auto-populates as AI extracts data
 * - Responsive: Side-by-side on desktop, stacked on tablet, tabs on mobile
 */
import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { resetOnboarding } from '../../store/onboarding-store';

import '../features/onboarding/fb-onboarding-chat';
import '../features/onboarding/fb-onboarding-form';

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
                overflow: hidden;
            }

            .wizard-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: var(--fb-spacing-md) var(--fb-spacing-lg);
                border-bottom: 1px solid var(--fb-border);
                flex-shrink: 0;
            }

            .wizard-title {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .close-btn {
                background: none;
                border: none;
                color: var(--fb-text-muted);
                cursor: pointer;
                padding: var(--fb-spacing-xs);
                border-radius: var(--fb-radius-sm);
                transition: all 0.2s;
                display: flex;
                align-items: center;
                justify-content: center;
                width: 32px;
                height: 32px;
            }

            .close-btn:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .close-btn svg {
                width: 20px;
                height: 20px;
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
                overflow: hidden;
            }

            .panel-form {
                flex: 1;
                overflow-y: auto;
                padding: var(--fb-spacing-xl);
            }

            /* Responsive: Stack on tablet (768px - 1023px) */
            @media (max-width: 1023px) {
                .wizard-body {
                    flex-direction: column;
                }
                .panel-chat {
                    border-right: none;
                    border-bottom: 1px solid var(--fb-border);
                    max-height: 50%;
                }
                .panel-form {
                    max-height: 50%;
                }
            }

            /* Mobile: Tab toggle mode (< 768px) - Phase 2 */
            @media (max-width: 767px) {
                .wizard-body {
                    position: relative;
                }
                /* Future: Implement tab toggle UI */
            }
        `
    ];

    override disconnectedCallback(): void {
        super.disconnectedCallback();
        resetOnboarding();
    }

    private _handleClose(): void {
        this.emit('navigate', { path: '/projects' });
    }

    private _handleProjectCreated(e: CustomEvent<{ projectId: string }>): void {
        this.emit('project-created', e.detail);
        this.emit('navigate', { path: `/projects/${e.detail.projectId}` });
    }

    override render(): TemplateResult {
        return html`
            <div class="wizard-header">
                <span class="wizard-title">New Project</span>
                <button
                    class="close-btn"
                    @click=${(): void => { this._handleClose(); }}
                    aria-label="Close wizard"
                    title="Close"
                >
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                    </svg>
                </button>
            </div>
            <div class="wizard-body">
                <div class="panel-chat">
                    <fb-onboarding-chat></fb-onboarding-chat>
                </div>
                <div class="panel-form">
                    <fb-onboarding-form
                        @project-created=${(e: CustomEvent<{ projectId: string }>): void => { this._handleProjectCreated(e); }}
                    ></fb-onboarding-form>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-onboarding': FBViewOnboarding;
    }
}

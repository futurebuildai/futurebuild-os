/**
 * FBViewOnboarding - Document-First Chat-Only Project Creation
 *
 * Full-width chat experience with horizontal progress bar:
 * - Top: Progress steps (Upload → Extract → Details → Review)
 * - Center: Full-width chat with document extraction and conversation
 * - Bottom: Create button when ready
 *
 * Collects all inputs needed for initial deterministic schedule generation
 * through AI document extraction and natural conversation.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { FBViewElement } from '../base/FBViewElement';
import { api } from '../../services/api';
import type { CreateProjectRequest } from '../../services/api';
import { store } from '../../store/store';
import type { ProjectSummary } from '../../store/types';
import {
    onboardingValues,
    resetOnboarding,
    uploadedPdfUrl,
    hasDocumentUploaded,
    schedulePreview,
} from '../../store/onboarding-store';

import '../features/onboarding/fb-onboarding-chat';
import '../features/onboarding/fb-onboarding-steps';
import '../features/onboarding/fb-engine-calibration';
import '../features/onboarding/fb-interrogator-wizard';
import '../features/onboarding/fb-onboard-schedule-preview';
import '../features/onboarding/fb-onboard-progress-selector';

@customElement('fb-view-onboarding')
export class FBViewOnboarding extends SignalWatcher(FBViewElement) {
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
                flex-direction: column;
                flex: 1;
                overflow: hidden;
            }

            .panel-chat {
                flex: 1;
                display: flex;
                flex-direction: column;
                overflow: hidden;
            }

            fb-interrogator-wizard {
                flex: 1;
                overflow: hidden;
            }
        `
    ];

    /**
     * Optional pre-filled values for the form (e.g., from "duplicate project").
     * Applied to the store on first connect.
     */
    @property({ type: Object, attribute: 'initial-values' }) initialValues: Partial<CreateProjectRequest> = {};

    /** When true, show calibration step before navigating home */
    @state() private _showCalibration = false;

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback(): void {
        super.connectedCallback();

        // Seed store with initial values if provided
        if (Object.keys(this.initialValues).length > 0) {
            onboardingValues.value = { ...this.initialValues };
        }

        // Emit form-updated whenever store values change (for parent sync)
        this._disposeEffect = effect(() => {
            const values = onboardingValues.value;
            this.emit('form-updated', values);
        });
    }

    override disconnectedCallback(): void {
        if (this._disposeEffect) {
            this._disposeEffect();
            this._disposeEffect = null;
        }
        resetOnboarding();
        super.disconnectedCallback();
    }

    private _navigateTo(path: string): void {
        window.history.pushState({}, '', path);
        window.dispatchEvent(new PopStateEvent('popstate'));
    }

    private _handleClose(): void {
        this._navigateTo('/projects');
    }

    private async _handleProjectCreated(e: CustomEvent<{ projectId: string }>): Promise<void> {
        const { projectId } = e.detail;
        this.emit('project-created', e.detail);

        try {
            // Fetch the newly created project and add it to the store's project list
            const detail = await api.projects.get(projectId);
            const summary: ProjectSummary = {
                id: detail.id,
                name: detail.name,
                address: detail.address,
                status: detail.status,
                completionPercentage: detail.completion_percentage,
                createdAt: detail.created_at,
                updatedAt: detail.updated_at,
            };
            const current = store.projects$.value;
            store.actions.setProjects([...current, summary]);

            // Select the new project (triggers thread loading + General thread auto-select)
            store.actions.setActiveProject(projectId);

            // First project → show calibration step before navigating home
            if (current.length === 0) {
                this._showCalibration = true;
                return;
            }

            // Navigate to chat view
            this._navigateTo('/');
        } catch {
            // Fallback: navigate to project detail page if store update fails
            this._navigateTo(`/projects/${projectId}`);
        }
    }

    private _handleCalibrationDone(): void {
        this._showCalibration = false;
        this._navigateTo('/');
    }

    override render(): TemplateResult {
        if (this._showCalibration) {
            return html`
                <div class="wizard-header">
                    <span class="wizard-title">Calibrate Your Engine</span>
                    <button
                        class="close-btn"
                        @click=${(): void => { this._handleCalibrationDone(); }}
                        aria-label="Skip calibration"
                        title="Skip"
                    >
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                        </svg>
                    </button>
                </div>
                <div class="wizard-body">
                    <fb-engine-calibration
                        @fb-calibration-applied=${(): void => { this._handleCalibrationDone(); }}
                        @fb-calibration-skipped=${(): void => { this._handleCalibrationDone(); }}
                    ></fb-engine-calibration>
                </div>
            `;
        }

        const showWizard = hasDocumentUploaded.value && uploadedPdfUrl.value;

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
            <fb-onboarding-steps></fb-onboarding-steps>
            <div class="wizard-body">
                ${showWizard ? html`
                    <fb-interrogator-wizard
                        @project-created=${(e: CustomEvent<{ projectId: string }>): void => { void this._handleProjectCreated(e); }}
                    ></fb-interrogator-wizard>
                ` : html`
                    <div class="panel-chat">
                        <fb-onboarding-chat
                            @project-created=${(e: CustomEvent<{ projectId: string }>): void => { void this._handleProjectCreated(e); }}
                        ></fb-onboarding-chat>
                        <fb-onboard-progress-selector></fb-onboard-progress-selector>
                        ${schedulePreview.value ? html`
                            <fb-onboard-schedule-preview></fb-onboard-schedule-preview>
                        ` : nothing}
                    </div>
                `}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-onboarding': FBViewOnboarding;
    }
}

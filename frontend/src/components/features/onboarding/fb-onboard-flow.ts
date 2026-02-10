/**
 * fb-onboard-flow — Full-screen onboarding orchestration.
 * See FRONTEND_V2_SPEC.md §2.3.B
 *
 * Combines: Drop zone → Extraction stream → Correction loop → Calibration → Activate
 *
 * Flow:
 * 1. Upload: Drop zone for contract/scope documents
 * 2. Extract: SSE-driven extraction narration
 * 3. Review: Chat-based correction loop for extracted values
 * 4. Calibrate: Work days + inspection latency (first project only)
 * 5. Activate: Create project and redirect to feed
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';
import './fb-onboarding-dropzone';
import './fb-extraction-stream';
import './fb-onboarding-chat';
import './fb-engine-calibration';
import type { ExtractionEvent } from './fb-extraction-stream';

type OnboardingStep = 'upload' | 'extract' | 'review' | 'calibrate' | 'complete';

interface ExtractedValues {
    address?: string;
    scope?: string;
    contractValue?: number;
    startDate?: string;
    completionDate?: string;
    contractor?: string;
    client?: string;
}

@customElement('fb-onboard-flow')
export class FBOnboardFlow extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                min-height: 100vh;
                background: var(--fb-bg-primary, #0d0d1a);
            }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 16px 24px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .logo {
                font-size: 18px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .logo span {
                color: var(--fb-accent, #6366f1);
            }

            .close-btn {
                display: flex;
                align-items: center;
                gap: 6px;
                padding: 8px 16px;
                border-radius: 8px;
                background: transparent;
                border: 1px solid var(--fb-border, #2a2a3e);
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 14px;
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .close-btn:hover {
                background: var(--fb-surface-1, #1a1a2e);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .progress-bar {
                display: flex;
                align-items: center;
                gap: 4px;
                padding: 12px 24px;
                background: var(--fb-surface-1, #1a1a2e);
            }

            .progress-step {
                flex: 1;
                height: 4px;
                background: var(--fb-border, #2a2a3e);
                border-radius: 2px;
                transition: background 0.3s ease;
            }

            .progress-step.active {
                background: var(--fb-accent, #6366f1);
            }

            .progress-step.complete {
                background: #22c55e;
            }

            .content {
                flex: 1;
                display: flex;
                align-items: center;
                justify-content: center;
                padding: 40px 24px;
            }

            .step-container {
                width: 100%;
                max-width: 600px;
            }

            .step-title {
                font-size: 28px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
                text-align: center;
                margin-bottom: 8px;
            }

            .step-subtitle {
                font-size: 16px;
                color: var(--fb-text-secondary, #a0a0b0);
                text-align: center;
                margin-bottom: 32px;
            }

            .two-column {
                display: grid;
                grid-template-columns: 1fr 1fr;
                gap: 24px;
                max-width: 900px;
            }

            .actions {
                display: flex;
                justify-content: center;
                gap: 12px;
                margin-top: 32px;
            }

            .btn {
                padding: 12px 28px;
                border-radius: 10px;
                font-size: 15px;
                font-weight: 600;
                cursor: pointer;
                border: none;
                transition: all 0.2s ease;
            }

            .btn-primary {
                background: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .btn-primary:hover:not(:disabled) {
                transform: translateY(-1px);
                box-shadow: 0 4px 16px rgba(99, 102, 241, 0.3);
            }

            .btn-primary:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .btn-secondary {
                background: transparent;
                border: 1px solid var(--fb-border, #2a2a3e);
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .btn-secondary:hover {
                background: var(--fb-surface-1, #1a1a2e);
                color: var(--fb-text-primary, #e0e0e0);
            }

            @media (max-width: 768px) {
                .two-column {
                    grid-template-columns: 1fr;
                }

                .step-title {
                    font-size: 22px;
                }
            }
        `,
    ];

    @state() private _step: OnboardingStep = 'upload';
    @state() private _extractedValues: ExtractedValues = {};
    // Reserved for session tracking when SSE is implemented
    // private _sessionId: string | null = null;
    @state() private _isFirstProject = true; // TODO: Check from API
    @state() private _extractionComplete = false;

    private _getStepIndex(): number {
        const steps: OnboardingStep[] = ['upload', 'extract', 'review', 'calibrate', 'complete'];
        return steps.indexOf(this._step);
    }

    private _handleClose() {
        this.emit('fb-navigate', { view: 'home' });
    }

    private _handleFilesSelected(e: CustomEvent<{ files: FileList }>) {
        // Start extraction process
        this._step = 'extract';
        this._startExtraction(e.detail.files);
    }

    private async _startExtraction(_files: FileList) {
        const stream = this.shadowRoot?.querySelector('fb-extraction-stream');
        if (!stream) return;

        // TODO: Send files to backend and receive SSE stream
        // Simulate extraction events for now (replace with real SSE when available)
        const events: ExtractionEvent[] = [
            { type: 'extraction', step: 'address', value: '123 Main St, Austin TX 78701', confidence: 0.95 },
            { type: 'extraction', step: 'scope', value: 'Single-family residential renovation', confidence: 0.88 },
            { type: 'extraction', step: 'contract_value', value: '$485,000', confidence: 0.92 },
            { type: 'scheduling', step: 'tasks_generated', count: 47 },
            { type: 'scheduling', step: 'dependencies_mapped', count: 62 },
            { type: 'weather', step: 'weather_check', value: 'Checking forecast for outdoor work' },
            { type: 'procurement', step: 'long_lead_detected', items: [
                { name: 'Roof Trusses (Engineered)', lead_weeks: 6, order_by: '2026-02-14' },
                { name: 'Custom Windows', lead_weeks: 8, order_by: '2026-02-01' },
            ]},
            { type: 'complete', step: 'done', ready_to_create: true },
        ];

        for (const event of events) {
            await new Promise(resolve => setTimeout(resolve, 800 + Math.random() * 400));
            (stream as any).addEvent(event);

            // Update extracted values
            if (event.type === 'extraction') {
                this._extractedValues = {
                    ...this._extractedValues,
                    [event.step]: event.value,
                };
            }
        }

        this._extractionComplete = true;
    }

    private _handleContinueToReview() {
        this._step = 'review';
    }

    private _handleReviewComplete() {
        if (this._isFirstProject) {
            this._step = 'calibrate';
        } else {
            this._createProject();
        }
    }

    private _handleCalibrationComplete() {
        this._createProject();
    }

    private _handleSkipCalibration() {
        this._createProject();
    }

    private async _createProject() {
        this._step = 'complete';
        // TODO: Call API to create project
        await new Promise(resolve => setTimeout(resolve, 1500));
        this.emit('fb-navigate', { view: 'home' });
    }

    private _renderProgressBar() {
        const currentIndex = this._getStepIndex();
        const steps = ['upload', 'extract', 'review', 'calibrate'];

        return html`
            <div class="progress-bar">
                ${steps.map((_, i) => html`
                    <div class="progress-step ${i < currentIndex ? 'complete' : i === currentIndex ? 'active' : ''}"></div>
                `)}
            </div>
        `;
    }

    private _renderUploadStep() {
        return html`
            <div class="step-container">
                <h1 class="step-title">Start Your Project</h1>
                <p class="step-subtitle">Drop your contract, scope document, or bid sheet and we'll extract everything</p>
                <fb-onboarding-dropzone
                    @files-selected=${this._handleFilesSelected}
                ></fb-onboarding-dropzone>
            </div>
        `;
    }

    private _renderExtractStep() {
        return html`
            <div class="step-container">
                <h1 class="step-title">Analyzing Your Document</h1>
                <p class="step-subtitle">Our AI is reading and understanding your project</p>
                <fb-extraction-stream></fb-extraction-stream>
                ${this._extractionComplete ? html`
                    <div class="actions">
                        <button class="btn btn-primary" @click=${this._handleContinueToReview}>
                            Continue to Review
                        </button>
                    </div>
                ` : nothing}
            </div>
        `;
    }

    private _renderReviewStep() {
        return html`
            <div class="step-container">
                <h1 class="step-title">Review & Refine</h1>
                <p class="step-subtitle">Chat with our AI to correct any details or add more context</p>
                <fb-onboarding-chat
                    .extractedValues=${this._extractedValues}
                    @review-complete=${this._handleReviewComplete}
                ></fb-onboarding-chat>
            </div>
        `;
    }

    private _renderCalibrateStep() {
        return html`
            <div class="step-container">
                <h1 class="step-title">Calibrate Your Engine</h1>
                <p class="step-subtitle">Tell us about your crew's schedule so we can build accurate timelines</p>
                <fb-engine-calibration
                    @calibration-complete=${this._handleCalibrationComplete}
                    @skip=${this._handleSkipCalibration}
                ></fb-engine-calibration>
            </div>
        `;
    }

    private _renderCompleteStep() {
        return html`
            <div class="step-container" style="text-align: center;">
                <div style="width: 64px; height: 64px; margin: 0 auto 24px; border-radius: 50%; background: #22c55e; display: flex; align-items: center; justify-content: center;">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"/>
                    </svg>
                </div>
                <h1 class="step-title">Project Created!</h1>
                <p class="step-subtitle">Redirecting to your feed...</p>
            </div>
        `;
    }

    override render() {
        return html`
            <header class="header">
                <div class="logo">Future<span>Build</span></div>
                <button class="close-btn" @click=${this._handleClose}>
                    Cancel
                </button>
            </header>

            ${this._renderProgressBar()}

            <div class="content">
                ${this._step === 'upload' ? this._renderUploadStep() : nothing}
                ${this._step === 'extract' ? this._renderExtractStep() : nothing}
                ${this._step === 'review' ? this._renderReviewStep() : nothing}
                ${this._step === 'calibrate' ? this._renderCalibrateStep() : nothing}
                ${this._step === 'complete' ? this._renderCompleteStep() : nothing}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboard-flow': FBOnboardFlow;
    }
}

/**
 * FBPhotoUpload - Mobile Camera/Gallery Upload with Vision Feedback Loop
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 * See STEP_84_FIELD_FEEDBACK.md Section 1: State Machine
 *
 * State Machine:
 *   IDLE → UPLOADING → ANALYZING → COMPLETE | ERROR
 *
 * Features:
 * - Camera capture button (uses capture="environment" for back camera)
 * - Gallery picker fallback
 * - Preview before upload
 * - Progress indicator during upload
 * - Exponential backoff polling for vision analysis status
 * - "Verifying..." spinner during ANALYZING state
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import type { VisionStatusResponse } from '../../services/api';

export interface PhotoUploadResult {
    id: string;
    url: string;
    fileName: string;
}

/**
 * Upload state machine states.
 * See STEP_84_FIELD_FEEDBACK.md Section 1.1
 */
type UploadState = 'IDLE' | 'UPLOADING' | 'ANALYZING' | 'COMPLETE' | 'ERROR';

/** Max polling duration before timeout (30 seconds) */
const MAX_POLL_DURATION_MS = 30_000;

/** Initial polling interval (1 second) */
const INITIAL_POLL_INTERVAL_MS = 1_000;

/** Maximum polling interval (5 seconds) */
const MAX_POLL_INTERVAL_MS = 5_000;

/** Maximum consecutive poll failures before giving up */
const MAX_POLL_FAILURES = 5;

/**
 * Mobile-friendly photo upload component with vision analysis feedback loop.
 * @element fb-photo-upload
 *
 * @fires fb-photo-selected - Fired when a photo is selected (before upload)
 * @fires fb-photo-uploaded - Fired after successful upload (before analysis)
 * @fires fb-analysis-complete - Fired when vision analysis completes
 * @fires fb-photo-error - Fired on upload or analysis error
 */
@customElement('fb-photo-upload')
export class FBPhotoUpload extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .upload-container {
                display: flex;
                flex-direction: column;
                gap: 16px;
            }

            .upload-area {
                position: relative;
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                gap: 12px;
                padding: 32px;
                background: var(--fb-bg-card, #161821);
                border: 2px dashed var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 12px;
                min-height: 200px;
                cursor: pointer;
                transition: border-color 0.2s ease;
            }

            .upload-area:hover {
                border-color: var(--fb-primary, #00FFA3);
            }

            .upload-area--has-preview {
                padding: 0;
                border-style: solid;
            }

            input[type="file"] {
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                opacity: 0;
                cursor: pointer;
            }

            .upload-icon {
                width: 48px;
                height: 48px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .upload-text {
                color: var(--fb-text-primary, #fff);
                font-size: 16px;
                font-weight: 500;
                margin: 0;
            }

            .upload-hint {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
                margin: 0;
            }

            .preview {
                width: 100%;
                max-height: 300px;
                object-fit: contain;
                border-radius: 10px;
            }

            .actions {
                display: flex;
                gap: 12px;
            }

            .btn {
                flex: 1;
                padding: 14px 20px;
                font-size: 16px;
                font-weight: 500;
                border: none;
                border-radius: 8px;
                cursor: pointer;
                transition: all 0.2s ease;
            }

            .btn--primary {
                background: var(--fb-primary, #00FFA3);
                color: white;
            }

            .btn--primary:hover:not([disabled]) {
                background: var(--fb-primary-hover, #5a6fd6);
            }

            .btn--secondary {
                background: var(--fb-bg-tertiary, #1a1a1a);
                color: var(--fb-text-primary, #fff);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .btn--secondary:hover:not([disabled]) {
                background: var(--fb-bg-card, #161821);
            }

            .btn[disabled] {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .progress-bar {
                height: 4px;
                background: var(--fb-bg-tertiary, #1a1a1a);
                border-radius: 2px;
                overflow: hidden;
            }

            .progress-fill {
                height: 100%;
                background: var(--fb-primary, #00FFA3);
                border-radius: 2px;
                transition: width 0.3s ease;
            }

            .error {
                display: flex;
                align-items: center;
                gap: 8px;
                padding: 12px;
                background: var(--fb-error-alpha, rgba(244, 63, 94, 0.1));
                border: 1px solid var(--fb-error, #F43F5E);
                border-radius: 8px;
                color: var(--fb-error, #F43F5E);
                font-size: 14px;
            }

            /* Analyzing state */
            .analyzing {
                display: flex;
                flex-direction: column;
                align-items: center;
                gap: 12px;
                padding: 24px;
            }

            .analyzing-spinner {
                width: 32px;
                height: 32px;
                border: 3px solid var(--fb-border, rgba(255,255,255,0.05));
                border-top-color: var(--fb-primary, #00FFA3);
                border-radius: 50%;
                animation: spin 0.8s linear infinite;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            .analyzing-text {
                color: var(--fb-text-primary, #fff);
                font-size: 14px;
                font-weight: 500;
            }

            .analyzing-hint {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 12px;
            }

            /* Complete state */
            .complete-badge {
                display: flex;
                align-items: center;
                gap: 8px;
                padding: 12px;
                background: rgba(5, 150, 105, 0.1);
                border: 1px solid #059669;
                border-radius: 8px;
                color: #059669;
                font-size: 14px;
                font-weight: 500;
            }

            .complete-icon {
                width: 20px;
                height: 20px;
                flex-shrink: 0;
            }
        `,
    ];

    @property({ type: String, attribute: 'upload-url' }) uploadUrl = '/api/v1/upload/photo';
    @property({ type: Boolean }) disabled = false;

    @state() private _selectedFile: File | null = null;
    @state() private _previewUrl: string | null = null;
    @state() private _uploadState: UploadState = 'IDLE';
    @state() private _progress = 0;
    @state() private _error: string | null = null;
    @state() private _analysisResult: VisionStatusResponse | null = null;

    /** Asset ID returned from upload, used for polling */
    private _assetId: string | null = null;

    /** Polling timer handle */
    private _pollTimer: ReturnType<typeof setTimeout> | null = null;

    /** Timestamp when polling started, for timeout detection */
    private _pollStartTime = 0;

    /** Consecutive polling failure count */
    private _pollFailures = 0;

    private _handleFileSelect(e: Event): void {
        const input = e.target as HTMLInputElement;
        const file = input.files?.[0];
        if (!file) return;

        // Validate file type
        if (!file.type.startsWith('image/')) {
            this._error = 'Please select an image file';
            return;
        }

        // Validate file size (max 10MB)
        if (file.size > 10 * 1024 * 1024) {
            this._error = 'Image must be less than 10MB';
            return;
        }

        this._selectedFile = file;
        this._error = null;
        this._uploadState = 'IDLE';
        this._analysisResult = null;

        // Create preview URL
        if (this._previewUrl) {
            URL.revokeObjectURL(this._previewUrl);
        }
        this._previewUrl = URL.createObjectURL(file);

        this.emit('fb-photo-selected', { file });
    }

    private _handleClear(): void {
        this._stopPolling();
        this._selectedFile = null;
        if (this._previewUrl) {
            URL.revokeObjectURL(this._previewUrl);
            this._previewUrl = null;
        }
        this._error = null;
        this._progress = 0;
        this._uploadState = 'IDLE';
        this._assetId = null;
        this._analysisResult = null;
    }

    /**
     * Upload the selected file and transition through the state machine.
     * IDLE → UPLOADING → ANALYZING → COMPLETE | ERROR
     */
    async upload(): Promise<PhotoUploadResult | null> {
        if (!this._selectedFile || this.disabled) return null;
        // H1 Fix: State guard prevents concurrent uploads from rapid taps
        if (this._uploadState === 'UPLOADING' || this._uploadState === 'ANALYZING') return null;

        this._uploadState = 'UPLOADING';
        this._error = null;
        this._progress = 0;

        try {
            const formData = new FormData();
            formData.append('file', this._selectedFile);

            // Simulate upload progress (real implementation would use XMLHttpRequest)
            const progressInterval = setInterval(() => {
                if (this._progress < 90) {
                    this._progress += 10;
                }
            }, 100);

            // M2: Portal uses magic-link auth (not Clerk JWT).
            // Upload handler must accept the portal session token or be public.
            // When the upload handler is built, integrate auth here.
            const response = await fetch(this.uploadUrl, {
                method: 'POST',
                body: formData,
            });

            clearInterval(progressInterval);
            this._progress = 100;

            if (!response.ok) {
                throw new Error('Upload failed');
            }

            const result = await response.json() as PhotoUploadResult;
            this._assetId = result.id;
            this.emit('fb-photo-uploaded', result);

            // Transition to ANALYZING and start polling
            this._uploadState = 'ANALYZING';
            this._startPolling();

            return result;
        } catch {
            this._uploadState = 'ERROR';
            this._error = 'Upload failed. Please try again.';
            this.emit('fb-photo-error', { error: this._error });
            return null;
        }
    }

    /**
     * Start exponential backoff polling for vision analysis status.
     * See STEP_84_FIELD_FEEDBACK.md Section 1.2
     */
    private _startPolling(): void {
        this._pollStartTime = Date.now();
        this._pollFailures = 0;
        this._pollWithBackoff(INITIAL_POLL_INTERVAL_MS);
    }

    /**
     * Poll with exponential backoff: 1s, 2s, 4s, 5s (capped).
     * Timeout after MAX_POLL_DURATION_MS.
     */
    private _pollWithBackoff(intervalMs: number): void {
        // Check for timeout
        if (Date.now() - this._pollStartTime > MAX_POLL_DURATION_MS) {
            this._uploadState = 'ERROR';
            this._error = 'Analysis timed out. Please try again.';
            this.emit('fb-photo-error', { error: this._error });
            return;
        }

        this._pollTimer = setTimeout(async () => {
            if (!this._assetId || this._uploadState !== 'ANALYZING') return;

            try {
                const status = await api.vision.status(this._assetId);
                this._pollFailures = 0; // Reset on successful response

                if (status.status === 'completed') {
                    this._uploadState = 'COMPLETE';
                    this._analysisResult = status;
                    this.emit('fb-analysis-complete', {
                        assetId: this._assetId,
                        analysis: status.analysis,
                    });
                    return;
                }

                if (status.status === 'failed') {
                    this._uploadState = 'ERROR';
                    this._error = 'Vision analysis failed. Please try again.';
                    this.emit('fb-photo-error', { error: this._error });
                    return;
                }

                // Still processing — poll again with increased interval
                const nextInterval = Math.min(intervalMs * 2, MAX_POLL_INTERVAL_MS);
                this._pollWithBackoff(nextInterval);
            } catch {
                // H2 Fix: Track consecutive failures with a hard limit
                this._pollFailures++;
                if (this._pollFailures >= MAX_POLL_FAILURES) {
                    this._uploadState = 'ERROR';
                    this._error = 'Unable to check analysis status. Please try again later.';
                    this.emit('fb-photo-error', { error: this._error });
                    return;
                }
                // Network error during polling — retry with increased backoff
                const nextInterval = Math.min(intervalMs * 2, MAX_POLL_INTERVAL_MS);
                this._pollWithBackoff(nextInterval);
            }
        }, intervalMs);
    }

    /** Stop any active polling timer. */
    private _stopPolling(): void {
        if (this._pollTimer !== null) {
            clearTimeout(this._pollTimer);
            this._pollTimer = null;
        }
    }

    override disconnectedCallback(): void {
        this._stopPolling();
        if (this._previewUrl) {
            URL.revokeObjectURL(this._previewUrl);
        }
        super.disconnectedCallback();
    }

    private _renderAnalyzing(): TemplateResult {
        return html`
            <div class="analyzing">
                <div class="analyzing-spinner"></div>
                <span class="analyzing-text">Verifying...</span>
                <span class="analyzing-hint">AI is analyzing your photo</span>
            </div>
        `;
    }

    private _renderComplete(): TemplateResult {
        const hasAnalysis = this._analysisResult?.analysis != null;
        return html`
            <div class="complete-badge">
                <svg class="complete-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                </svg>
                ${hasAnalysis ? 'Verified' : 'Analysis Complete'}
            </div>
        `;
    }

    override render(): TemplateResult {
        const isProcessing = this._uploadState === 'UPLOADING' || this._uploadState === 'ANALYZING';

        return html`
            <div class="upload-container">
                ${this._error
                    ? html`
                          <div class="error">
                              <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                                  <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                              </svg>
                              ${this._error}
                          </div>
                      `
                    : nothing}

                <div class="upload-area ${this._previewUrl ? 'upload-area--has-preview' : ''}">
                    ${this._previewUrl
                        ? html`<img class="preview" src="${this._previewUrl}" alt="Preview" />`
                        : html`
                              <svg class="upload-icon" viewBox="0 0 24 24" fill="currentColor">
                                  <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-4h2v-4h3l-4-4-4 4h3z"/>
                              </svg>
                              <p class="upload-text">Tap to take a photo</p>
                              <p class="upload-hint">or choose from gallery</p>
                          `}
                    <input
                        type="file"
                        accept="image/*"
                        capture="environment"
                        ?disabled=${this.disabled || isProcessing}
                        @change=${this._handleFileSelect.bind(this)}
                    />
                </div>

                ${this._uploadState === 'UPLOADING'
                    ? html`
                          <div class="progress-bar">
                              <div class="progress-fill" style="width: ${this._progress}%"></div>
                          </div>
                      `
                    : nothing}

                ${this._uploadState === 'ANALYZING' ? this._renderAnalyzing() : nothing}
                ${this._uploadState === 'COMPLETE' ? this._renderComplete() : nothing}

                ${this._selectedFile && !isProcessing && this._uploadState !== 'COMPLETE'
                    ? html`
                          <div class="actions">
                              <button
                                  class="btn btn--secondary"
                                  ?disabled=${isProcessing}
                                  @click=${this._handleClear.bind(this)}
                              >
                                  Clear
                              </button>
                              <button
                                  class="btn btn--primary"
                                  ?disabled=${isProcessing}
                                  @click=${() => this.upload()}
                              >
                                  Upload Photo
                              </button>
                          </div>
                      `
                    : nothing}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-photo-upload': FBPhotoUpload;
    }
}

/**
 * FBOnboardingDropzone - Drag-and-Drop Zone for Blueprint/Document Upload
 * See STEP_77_MAGIC_UPLOAD_TRIGGER.md
 *
 * Compact drag-and-drop zone for the onboarding chat panel.
 * - Accepts PDF, PNG, JPG files (50MB max)
 * - Visual feedback on drag over
 * - Progress bar with uploading/analyzing states
 * - Keyboard accessible (Enter/Space to browse)
 * - Emits file-dropped event for parent to handle
 *
 * @fires file-dropped - When a valid file is selected with { files: [file] }
 * @fires upload-error - When validation fails with { error: string }
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';

const MAX_FILE_SIZE = 50 * 1024 * 1024; // 50MB
const ACCEPTED_TYPES = ['application/pdf', 'image/png', 'image/jpeg'];

@customElement('fb-onboarding-dropzone')
export class FBOnboardingDropzone extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
            }

            .dropzone {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-lg);
                border: 2px dashed var(--md-sys-color-outline-variant);
                border-radius: var(--md-sys-shape-corner-medium);
                background: var(--md-sys-color-surface-container);
                cursor: pointer;
                transition: all var(--fb-transition-base);
                min-height: var(--dropzone-min-height, 120px);
                outline: none;
                position: relative;
                overflow: hidden;
            }

            .dropzone:focus-visible {
                outline: 2px solid var(--md-sys-color-primary);
                outline-offset: 2px;
            }

            .dropzone:hover:not(.disabled) {
                border-color: var(--md-sys-color-primary);
                background: var(--md-sys-color-surface-container-high);
            }

            .dropzone.drag-over {
                border-color: var(--md-sys-color-primary);
                background: var(--md-sys-color-primary-container);
                border-style: solid;
            }

            .dropzone.drag-over .dropzone-text,
            .dropzone.drag-over .dropzone-icon {
                color: var(--md-sys-color-on-primary-container);
            }

            .dropzone.uploading {
                border-color: var(--md-sys-color-primary);
                border-style: solid;
                cursor: default;
                background: var(--md-sys-color-surface-container-high);
            }

            .dropzone.disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .dropzone-icon {
                font-size: 32px;
                color: var(--md-sys-color-primary);
                transition: color var(--fb-transition-base);
            }

            .dropzone-text {
                font: var(--md-sys-typescale-body-medium);
                color: var(--md-sys-color-on-surface);
                text-align: center;
            }

            .dropzone-hint {
                font: var(--md-sys-typescale-body-small);
                color: var(--md-sys-color-on-surface-variant);
                text-align: center;
            }

            .error-message {
                color: var(--md-sys-color-error);
                font: var(--md-sys-typescale-body-small);
                text-align: center;
                margin-top: var(--fb-spacing-xs);
            }

            .progress-bg {
                width: 100%;
                max-width: 240px;
                height: 4px;
                background: var(--md-sys-color-surface-variant);
                border-radius: var(--md-sys-shape-corner-full);
                margin-top: var(--fb-spacing-md);
                overflow: hidden;
            }

            .progress-fill {
                height: 100%;
                background-color: var(--md-sys-color-primary);
                transition: width 0.3s ease;
                width: 0%;
            }

            input[type="file"] {
                display: none;
            }
        `
    ];

    @property({ type: Boolean }) disabled = false;
    @state() private _isDragOver = false;
    @state() private _errorMessage = '';
    @state() private _isUploading = false;
    @state() private _uploadProgress = 0;
    @state() private _uploadFileName = '';

    /** Called by parent to indicate upload has started. */
    public setUploading(fileName: string): void {
        this._isUploading = true;
        this._uploadProgress = 0;
        this._uploadFileName = fileName;
        this._errorMessage = '';
    }

    /** Called by parent to update upload progress (0-100). */
    public setProgress(pct: number): void {
        this._uploadProgress = Math.min(100, Math.max(0, pct));
    }

    /** Called by parent when upload/analysis is done. */
    public setComplete(): void {
        this._uploadProgress = 100;
        setTimeout(() => {
            this._isUploading = false;
            this._uploadProgress = 0;
            this._uploadFileName = '';
        }, 500);
    }

    /** Called by parent to reset on error. */
    public setError(message: string): void {
        this._isUploading = false;
        this._uploadProgress = 0;
        this._uploadFileName = '';
        this._errorMessage = message;
    }

    private get _isDisabled(): boolean {
        return this.disabled || this._isUploading;
    }

    private _handleDragEnter(e: DragEvent): void {
        if (this._isDisabled) return;
        e.preventDefault();
        e.stopPropagation();
        this._isDragOver = true;
    }

    private _handleDragLeave(e: DragEvent): void {
        if (this._isDisabled) return;
        e.preventDefault();
        e.stopPropagation();
        this._isDragOver = false;
    }

    private _handleDragOver(e: DragEvent): void {
        if (this._isDisabled) return;
        e.preventDefault();
        e.stopPropagation();
    }

    private _handleDrop(e: DragEvent): void {
        if (this._isDisabled) return;
        e.preventDefault();
        e.stopPropagation();
        this._isDragOver = false;
        this._errorMessage = '';

        const files = Array.from(e.dataTransfer?.files ?? []);
        this._processFiles(files);
    }

    private _handleClick(): void {
        if (this._isDisabled) return;
        const input = this.shadowRoot?.querySelector<HTMLInputElement>('input[type="file"]');
        if (input) input.click();
    }

    private _handleKeyDown(e: KeyboardEvent): void {
        if (this._isDisabled) return;
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            this._handleClick();
        }
    }

    private _handleFileSelect(e: Event): void {
        if (this._isDisabled) return;
        const input = e.target as HTMLInputElement;
        const files = Array.from(input.files ?? []);
        this._processFiles(files);
        // Reset input so the same file can be selected again
        input.value = '';
    }

    private _processFiles(files: File[]): void {
        if (files.length === 0) {
            return;
        }

        // Fix 9: Warn user about multiple files instead of silently discarding
        if (files.length > 1) {
            this._errorMessage = 'Please upload one file at a time';
            return;
        }

        const file = files[0];
        if (!file) return;

        // Validate file type
        if (!ACCEPTED_TYPES.includes(file.type)) {
            this._errorMessage = 'Please upload a PDF, PNG, or JPG file';
            return;
        }

        // Validate file size
        if (file.size > MAX_FILE_SIZE) {
            this._errorMessage = 'File too large. Maximum size is 50MB.';
            return;
        }

        this._errorMessage = '';

        // Emit file-dropped event
        this.emit('file-dropped', { files: [file] });
    }

    private _renderIdle(): TemplateResult {
        return html`
            <div class="dropzone-icon">📄</div>
            <div class="dropzone-text">
                ${this._isDragOver ? 'Drop file here' : 'Drag blueprint or document here'}
            </div>
            <div class="dropzone-hint">
                or click to browse (PDF, PNG, JPG • Max 50MB)
            </div>
        `;
    }

    private _renderUploading(): TemplateResult {
        return html`
            <div class="dropzone-icon">⏳</div>
            <div class="dropzone-text">
                ${this._uploadProgress < 100
                ? `Uploading ${this._uploadFileName}...`
                : 'Analyzing blueprint...'}
            </div>
            <div class="progress-bg">
                <div
                    class="progress-fill"
                    style="width: ${String(this._uploadProgress)}%"
                ></div>
            </div>
        `;
    }

    override render(): TemplateResult {
        const dzClass = [
            'dropzone',
            this._isDragOver ? 'drag-over' : '',
            this._isDisabled ? 'disabled' : '',
            this._isUploading ? 'uploading' : '',
        ].filter(Boolean).join(' ');

        return html`
            <div
                class=${dzClass}
                tabindex="0"
                role="button"
                aria-label="Upload blueprint file. Drag and drop or press Enter to browse."
                @dragenter=${this._handleDragEnter}
                @dragleave=${this._handleDragLeave}
                @dragover=${this._handleDragOver}
                @drop=${this._handleDrop}
                @click=${this._handleClick}
                @keydown=${this._handleKeyDown}
            >
                ${this._isUploading ? this._renderUploading() : this._renderIdle()}
            </div>

            ${this._errorMessage ? html`
                <div class="error-message">${this._errorMessage}</div>
            ` : nothing}

            <input
                type="file"
                accept=".pdf,.png,.jpg,.jpeg"
                @change=${this._handleFileSelect}
            />
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboarding-dropzone': FBOnboardingDropzone;
    }
}

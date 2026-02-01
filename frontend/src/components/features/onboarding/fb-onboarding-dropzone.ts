/**
 * FBOnboardingDropzone - Drag-and-Drop Zone for Blueprint/Document Upload
 * See STEP_74_SPLIT_SCREEN_WIZARD.md Task 2
 *
 * Compact drag-and-drop zone for the onboarding chat panel.
 * - Accepts PDF, PNG, JPG files
 * - Visual feedback on drag over
 * - Emits file-dropped event for parent to handle
 * - 50MB max file size (Step 77 requirement)
 */
import { html, css, TemplateResult } from 'lit';
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
                gap: var(--fb-spacing-xs, 4px);
                padding: var(--fb-spacing-md, 16px);
                border: 2px dashed var(--fb-border, #e5e5e5);
                border-radius: var(--fb-radius-md, 8px);
                background: var(--fb-bg-card, white);
                cursor: pointer;
                transition: all 0.2s;
                min-height: 80px;
            }

            .dropzone:hover:not(.disabled) {
                border-color: #667eea;
                background: rgba(102, 126, 234, 0.02);
            }

            .dropzone.drag-over {
                border-color: #667eea;
                background: rgba(102, 126, 234, 0.08);
                border-style: solid;
            }

            .dropzone.disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .dropzone-icon {
                font-size: 32px;
                opacity: 0.6;
            }

            .dropzone-text {
                font-size: var(--fb-text-sm, 14px);
                color: var(--fb-text-muted, #666);
                text-align: center;
            }

            .dropzone-hint {
                font-size: var(--fb-text-xs, 12px);
                color: var(--fb-text-muted, #666);
                text-align: center;
            }

            .error-message {
                color: #dc2626;
                font-size: var(--fb-text-sm, 14px);
                text-align: center;
                margin-top: var(--fb-spacing-xs, 4px);
            }

            input[type="file"] {
                display: none;
            }
        `
    ];

    @property({ type: Boolean }) disabled = false;
    @state() private _isDragOver = false;
    @state() private _errorMessage = '';

    private _handleDragEnter(e: DragEvent): void {
        if (this.disabled) return;
        e.preventDefault();
        e.stopPropagation();
        this._isDragOver = true;
    }

    private _handleDragLeave(e: DragEvent): void {
        if (this.disabled) return;
        e.preventDefault();
        e.stopPropagation();
        this._isDragOver = false;
    }

    private _handleDragOver(e: DragEvent): void {
        if (this.disabled) return;
        e.preventDefault();
        e.stopPropagation();
    }

    private _handleDrop(e: DragEvent): void {
        if (this.disabled) return;
        e.preventDefault();
        e.stopPropagation();
        this._isDragOver = false;
        this._errorMessage = '';

        const files = Array.from(e.dataTransfer?.files ?? []);
        this._processFiles(files);
    }

    private _handleClick(): void {
        if (this.disabled) return;
        const input = this.shadowRoot?.querySelector('input[type="file"]') as HTMLInputElement;
        input?.click();
    }

    private _handleFileSelect(e: Event): void {
        if (this.disabled) return;
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

        const file = files[0]; // Only process first file for now

        // Validate file type
        if (!ACCEPTED_TYPES.includes(file.type)) {
            this._errorMessage = 'Please upload a PDF, PNG, or JPG file';
            return;
        }

        // Validate file size
        if (file.size > MAX_FILE_SIZE) {
            this._errorMessage = 'File size must be less than 50MB';
            return;
        }

        // Emit file-dropped event
        this.emit('file-dropped', { files: [file] });
    }

    override render(): TemplateResult {
        return html`
            <div
                class="dropzone ${this._isDragOver ? 'drag-over' : ''} ${this.disabled ? 'disabled' : ''}"
                @dragenter=${this._handleDragEnter}
                @dragleave=${this._handleDragLeave}
                @dragover=${this._handleDragOver}
                @drop=${this._handleDrop}
                @click=${this._handleClick}
            >
                <div class="dropzone-icon">📄</div>
                <div class="dropzone-text">
                    ${this._isDragOver ? 'Drop file here' : 'Drag blueprint or document here'}
                </div>
                <div class="dropzone-hint">
                    or click to browse (PDF, PNG, JPG • Max 50MB)
                </div>
            </div>

            ${this._errorMessage ? html`
                <div class="error-message">${this._errorMessage}</div>
            ` : ''}

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

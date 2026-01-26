/**
 * FBPhotoUpload - Mobile Camera/Gallery Upload Component
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Features:
 * - Camera capture button (uses capture="environment" for back camera)
 * - Gallery picker fallback
 * - Preview before upload
 * - Progress indicator during upload
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export interface PhotoUploadResult {
    id: string;
    url: string;
    fileName: string;
}

/**
 * Mobile-friendly photo upload component.
 * @element fb-photo-upload
 *
 * @fires fb-photo-selected - Fired when a photo is selected (before upload)
 * @fires fb-photo-uploaded - Fired after successful upload
 * @fires fb-photo-error - Fired on upload error
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
                background: var(--fb-bg-card, #111);
                border: 2px dashed var(--fb-border, #333);
                border-radius: 12px;
                min-height: 200px;
                cursor: pointer;
                transition: border-color 0.2s ease;
            }

            .upload-area:hover {
                border-color: var(--fb-primary, #667eea);
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
                color: var(--fb-text-secondary, #aaa);
            }

            .upload-text {
                color: var(--fb-text-primary, #fff);
                font-size: 16px;
                font-weight: 500;
                margin: 0;
            }

            .upload-hint {
                color: var(--fb-text-secondary, #aaa);
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
                background: var(--fb-primary, #667eea);
                color: white;
            }

            .btn--primary:hover:not([disabled]) {
                background: var(--fb-primary-hover, #5a6fd6);
            }

            .btn--secondary {
                background: var(--fb-bg-tertiary, #1a1a1a);
                color: var(--fb-text-primary, #fff);
                border: 1px solid var(--fb-border, #333);
            }

            .btn--secondary:hover:not([disabled]) {
                background: var(--fb-bg-card, #111);
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
                background: var(--fb-primary, #667eea);
                border-radius: 2px;
                transition: width 0.3s ease;
            }

            .error {
                display: flex;
                align-items: center;
                gap: 8px;
                padding: 12px;
                background: var(--fb-error-alpha, rgba(198, 40, 40, 0.1));
                border: 1px solid var(--fb-error, #c62828);
                border-radius: 8px;
                color: var(--fb-error, #c62828);
                font-size: 14px;
            }
        `,
    ];

    @property({ type: String, attribute: 'upload-url' }) uploadUrl = '/api/v1/upload/photo';
    @property({ type: Boolean }) disabled = false;

    @state() private _selectedFile: File | null = null;
    @state() private _previewUrl: string | null = null;
    @state() private _uploading = false;
    @state() private _progress = 0;
    @state() private _error: string | null = null;

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

        // Create preview URL
        if (this._previewUrl) {
            URL.revokeObjectURL(this._previewUrl);
        }
        this._previewUrl = URL.createObjectURL(file);

        this.emit('fb-photo-selected', { file });
    }

    private _handleClear(): void {
        this._selectedFile = null;
        if (this._previewUrl) {
            URL.revokeObjectURL(this._previewUrl);
            this._previewUrl = null;
        }
        this._error = null;
        this._progress = 0;
    }

    async upload(): Promise<PhotoUploadResult | null> {
        if (!this._selectedFile || this.disabled) return null;

        this._uploading = true;
        this._error = null;
        this._progress = 0;

        try {
            const formData = new FormData();
            formData.append('file', this._selectedFile);

            // Note: In a real implementation, this would use XMLHttpRequest for progress tracking
            // or the Fetch API. For now, we simulate progress.
            const progressInterval = setInterval(() => {
                if (this._progress < 90) {
                    this._progress += 10;
                }
            }, 100);

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
            this.emit('fb-photo-uploaded', result);
            return result;
        } catch {
            this._error = 'Upload failed. Please try again.';
            this.emit('fb-photo-error', { error: this._error });
            return null;
        } finally {
            this._uploading = false;
        }
    }

    override disconnectedCallback(): void {
        if (this._previewUrl) {
            URL.revokeObjectURL(this._previewUrl);
        }
        super.disconnectedCallback();
    }

    override render(): TemplateResult {
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
                        ?disabled=${this.disabled || this._uploading}
                        @change=${this._handleFileSelect.bind(this)}
                    />
                </div>

                ${this._uploading
                    ? html`
                          <div class="progress-bar">
                              <div class="progress-fill" style="width: ${this._progress}%"></div>
                          </div>
                      `
                    : nothing}

                ${this._selectedFile
                    ? html`
                          <div class="actions">
                              <button
                                  class="btn btn--secondary"
                                  ?disabled=${this._uploading}
                                  @click=${this._handleClear.bind(this)}
                              >
                                  Clear
                              </button>
                              <button
                                  class="btn btn--primary"
                                  ?disabled=${this._uploading}
                                  @click=${() => this.upload()}
                              >
                                  ${this._uploading ? 'Uploading...' : 'Upload Photo'}
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

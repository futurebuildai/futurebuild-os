import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import { notify } from '../../store/notifications';

@customElement('fb-view-create-project')
export class FBViewCreateProject extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                padding: 24px;
                max-width: 600px;
                margin: 0 auto;
            }

            h1 {
                font-size: 24px;
                margin-bottom: 8px;
                color: var(--fb-text-primary);
            }

            .subtitle {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-bottom: 24px;
            }

            .drop-zone {
                border: 2px dashed var(--fb-border, #2a2a3e);
                border-radius: 12px;
                padding: 40px 24px;
                text-align: center;
                cursor: pointer;
                transition: all 0.2s ease;
                margin-bottom: 24px;
                background: var(--fb-surface-1, #1a1a2e);
            }

            .drop-zone:hover,
            .drop-zone.dragover {
                border-color: var(--fb-accent, #6366f1);
                background: rgba(99, 102, 241, 0.05);
            }

            .drop-icon {
                font-size: 36px;
                margin-bottom: 12px;
            }

            .drop-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 4px;
            }

            .drop-hint {
                font-size: 13px;
                color: var(--fb-text-tertiary, #707080);
            }

            .file-list {
                margin-bottom: 24px;
            }

            .file-item {
                display: flex;
                align-items: center;
                gap: 10px;
                padding: 10px 14px;
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 8px;
                margin-bottom: 8px;
                font-size: 13px;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .file-item .name { flex: 1; }
            .file-item .size {
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
            }

            .file-remove {
                background: none;
                border: none;
                color: var(--fb-text-tertiary, #707080);
                cursor: pointer;
                padding: 4px;
                font-size: 16px;
            }
            .file-remove:hover { color: #ef4444; }

            .divider {
                display: flex;
                align-items: center;
                gap: 12px;
                margin: 24px 0;
                color: var(--fb-text-tertiary, #707080);
                font-size: 12px;
                text-transform: uppercase;
            }
            .divider::before, .divider::after {
                content: '';
                flex: 1;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .form-group {
                margin-bottom: 16px;
            }

            label {
                display: block;
                margin-bottom: 8px;
                color: var(--fb-text-secondary);
                font-size: 14px;
            }

            input, select, textarea {
                width: 100%;
                padding: 10px;
                border-radius: 8px;
                border: 1px solid var(--fb-border);
                background: var(--fb-surface-2);
                color: var(--fb-text-primary);
                font-family: inherit;
                box-sizing: border-box;
            }

            textarea {
                min-height: 100px;
                resize: vertical;
            }

            .actions {
                display: flex;
                justify-content: flex-end;
                gap: 12px;
                margin-top: 32px;
            }

            button {
                padding: 10px 20px;
                border-radius: 8px;
                font-weight: 500;
                cursor: pointer;
                border: none;
            }

            .btn-cancel {
                background: transparent;
                color: var(--fb-text-secondary);
            }

            .btn-primary {
                background: var(--fb-accent);
                color: white;
            }

            .btn-primary:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }
        `,
    ];

    @state() private _name = '';
    @state() private _client = '';
    @state() private _description = '';
    @state() private _files: File[] = [];
    @state() private _dragover = false;
    @state() private _isExtracting = false;
    @state() private _extractionError = '';

    private async _processFiles(newFiles: File[]) {
        if (newFiles.length === 0) return;

        // Add all files visually
        this._files = [...this._files, ...newFiles];

        // Take the first valid file to extract (usually a plan set)
        const fileToExtract = newFiles[0];
        if (!fileToExtract) return;

        this._isExtracting = true;
        this._extractionError = '';

        try {
            const resp = await api.vision.extract(fileToExtract);

            // Map extracted values to our form state
            if (resp && resp.extracted_values) {
                const values = resp.extracted_values as Record<string, string>;

                // Map common fields that might be returned
                if (values.name || values.project_name) {
                    this._name = values.name || values.project_name || this._name;
                }

                if (values.client || values.client_name || values.owner) {
                    this._client = values.client || values.client_name || values.owner || this._client;
                }

                // Build a description from other useful extracted values
                const descLines = [];
                if (values.address) descLines.push(`Address: ${values.address}`);
                if (values.square_footage) descLines.push(`Square Footage: ${values.square_footage}`);
                if (values.bedrooms) descLines.push(`Bedrooms: ${values.bedrooms}`);
                if (values.bathrooms) descLines.push(`Bathrooms: ${values.bathrooms}`);

                if (descLines.length > 0) {
                    this._description = this._description
                        ? `${this._description}\n\nExtracted Details:\n${descLines.join('\n')}`
                        : `Extracted Details:\n${descLines.join('\n')}`;
                }

                // Add success toast
                notify.success('Project details extracted from plan set.');
            }
        } catch (err: any) {
            console.error('Extraction failed:', err);
            this._extractionError = err.message || 'Failed to extract data from document.';
            notify.error(this._extractionError);
        } finally {
            this._isExtracting = false;
        }
    }

    private _handleSubmit(e: Event) {
        e.preventDefault();
        console.log('Creates project:', {
            name: this._name,
            client: this._client,
            description: this._description,
            files: this._files.map(f => f.name),
        });
        this.emit('fb-navigate', { view: 'home' });
    }

    private _handleCancel() {
        this.emit('fb-navigate', { view: 'home' });
    }

    private _handleDragOver(e: DragEvent) {
        e.preventDefault();
        this._dragover = true;
    }

    private _handleDragLeave() {
        this._dragover = false;
    }

    private _handleDrop(e: DragEvent) {
        e.preventDefault();
        this._dragover = false;
        if (e.dataTransfer?.files) {
            void this._processFiles(Array.from(e.dataTransfer.files));
        }
    }

    private _handleFileSelect() {
        const input = document.createElement('input');
        input.type = 'file';
        input.multiple = true;
        input.accept = '.pdf,.doc,.docx,.xls,.xlsx,.csv,.txt,.jpg,.jpeg,.png,.webp';
        input.addEventListener('change', () => {
            if (input.files) {
                void this._processFiles(Array.from(input.files));
            }
        });
        input.click();
    }

    private _removeFile(index: number) {
        this._files = this._files.filter((_, i) => i !== index);
    }

    private _formatFileSize(bytes: number): string {
        if (bytes < 1024) return `${bytes} B`;
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
        return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    }

    override render() {
        return html`
            <h1>Create New Project</h1>
            <div class="subtitle">Upload project documents or enter details manually</div>

            <div
                class="drop-zone ${this._dragover ? 'dragover' : ''}"
                @dragover=${this._handleDragOver}
                @dragleave=${this._handleDragLeave}
                @drop=${this._handleDrop}
                @click=${this._handleFileSelect}
            >
                <div class="drop-icon">📁</div>
                <div class="drop-title">Drop project files here</div>
                <div class="drop-hint">
                    PDF, Word, Excel, or CSV — plans, specs, budgets, schedules
                </div>
                ${this._isExtracting ? html`
                    <div style="margin-top: 16px; color: var(--fb-accent); font-weight: 500; font-size: 14px;">
                        ✨ Extracting project data...
                    </div>
                ` : nothing}
            </div>

            ${this._files.length > 0 ? html`
                <div class="file-list">
                    ${this._files.map((file, i) => html`
                        <div class="file-item">
                            <span>📄</span>
                            <span class="name">${file.name}</span>
                            <span class="size">${this._formatFileSize(file.size)}</span>
                            <button class="file-remove" @click=${() => this._removeFile(i)} aria-label="Remove file">✕</button>
                        </div>
                    `)}
                </div>
            ` : nothing}

            <div class="divider">or enter details manually</div>

            <form @submit=${this._handleSubmit}>
                <div class="form-group">
                    <label for="name">Project Name</label>
                    <input
                        type="text"
                        id="name"
                        .value=${this._name}
                        @input=${(e: any) => this._name = e.target.value}
                        required
                        placeholder="e.g. Skyline Tower"
                    >
                </div>

                <div class="form-group">
                    <label for="client">Client</label>
                    <input
                        type="text"
                        id="client"
                        .value=${this._client}
                        @input=${(e: any) => this._client = e.target.value}
                        placeholder="Client Name"
                    >
                </div>

                <div class="form-group">
                    <label for="description">Project Description</label>
                    <textarea
                        id="description"
                        .value=${this._description}
                        @input=${(e: any) => this._description = e.target.value}
                        placeholder="Paste project details, scope, or notes here..."
                    ></textarea>
                </div>

                <div class="actions">
                    <button type="button" class="btn-cancel" @click=${this._handleCancel}>Cancel</button>
                    <button type="submit" class="btn-primary">Create Project</button>
                </div>
            </form>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-create-project': FBViewCreateProject;
    }
}

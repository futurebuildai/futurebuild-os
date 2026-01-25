/**
 * FBProjectForm - Smart Form Component for Creating Projects
 * See PROJECT_ONBOARDING_SPEC.md Step 62.5
 *
 * Handles form validation and API submission for creating new projects.
 * Emits 'project-created' event with ProjectDetail on success.
 *
 * @element fb-project-form
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';
import { api } from '../../../services/api';
import type { ProjectDetail } from '../../../store/types';

@customElement('fb-project-form')
export class FBProjectForm extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            form {
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-lg);
            }

            .field {
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-xs);
            }

            label {
                font-size: var(--fb-text-sm);
                font-weight: 500;
                color: var(--fb-text-secondary);
            }

            .required::after {
                content: ' *';
                color: #e53e3e;
            }

            input {
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                background: var(--fb-bg-primary);
                color: var(--fb-text-primary);
                font-size: var(--fb-text-base);
                font-family: inherit;
                transition: border-color 0.15s ease, box-shadow 0.15s ease;
            }

            input:focus {
                outline: none;
                border-color: var(--fb-primary);
                box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.2);
            }

            input:invalid:not(:placeholder-shown) {
                border-color: #e53e3e;
            }

            input[type="number"] {
                -moz-appearance: textfield;
            }

            input[type="number"]::-webkit-outer-spin-button,
            input[type="number"]::-webkit-inner-spin-button {
                -webkit-appearance: none;
                margin: 0;
            }

            .error-message {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-xs);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: rgba(229, 62, 62, 0.1);
                border: 1px solid rgba(229, 62, 62, 0.3);
                border-radius: var(--fb-radius-md);
                color: #e53e3e;
                font-size: var(--fb-text-sm);
            }

            .actions {
                display: flex;
                justify-content: flex-end;
                gap: var(--fb-spacing-md);
                margin-top: var(--fb-spacing-md);
            }

            button {
                padding: var(--fb-spacing-sm) var(--fb-spacing-lg);
                border-radius: var(--fb-radius-md);
                font-size: var(--fb-text-sm);
                font-weight: 500;
                font-family: inherit;
                cursor: pointer;
                transition: background 0.15s ease, opacity 0.15s ease;
            }

            button:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .btn-cancel {
                background: transparent;
                border: 1px solid var(--fb-border);
                color: var(--fb-text-secondary);
            }

            .btn-cancel:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .btn-submit {
                background: var(--fb-primary, #667eea);
                border: none;
                color: white;
            }

            .btn-submit:hover:not(:disabled) {
                background: var(--fb-primary-hover, #5a67d8);
            }

            .btn-submit:disabled {
                opacity: 0.6;
                cursor: not-allowed;
            }

            .spinner {
                display: inline-block;
                width: 14px;
                height: 14px;
                border: 2px solid transparent;
                border-top-color: currentColor;
                border-radius: 50%;
                animation: spin 0.8s linear infinite;
                margin-right: var(--fb-spacing-xs);
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }
        `,
    ];

    @state() private _name = '';
    @state() private _address = '';
    @state() private _squareFootage = 2500;
    @state() private _loading = false;
    @state() private _error = '';

    private _handleNameInput(e: Event): void {
        const target = e.target as HTMLInputElement;
        this._name = target.value;
    }

    private _handleAddressInput(e: Event): void {
        const target = e.target as HTMLInputElement;
        this._address = target.value;
    }

    private _handleSquareFootageInput(e: Event): void {
        const target = e.target as HTMLInputElement;
        this._squareFootage = parseInt(target.value, 10) || 0;
    }

    private _handleCancel(): void {
        this.emit('cancel');
    }

    private async _handleSubmit(e: Event): Promise<void> {
        e.preventDefault();

        // Validation
        if (!this._name.trim()) {
            this._error = 'Project name is required';
            return;
        }
        if (!this._address.trim()) {
            this._error = 'Address is required';
            return;
        }

        this._loading = true;
        this._error = '';

        try {
            const projectDetail = await api.projects.create({
                name: this._name.trim(),
                address: this._address.trim(),
                square_footage: this._squareFootage,
                bedrooms: 0,
                bathrooms: 0,
                start_date: new Date().toISOString(),
            });

            // Convert API response to store format
            const project: ProjectDetail = {
                id: projectDetail.id,
                name: projectDetail.name,
                address: projectDetail.address,
                status: projectDetail.status,
                completionPercentage: projectDetail.completion_percentage,
                createdAt: projectDetail.created_at,
                updatedAt: projectDetail.updated_at,
                orgId: projectDetail.org_id,
                squareFootage: projectDetail.square_footage,
                bedrooms: projectDetail.bedrooms,
                bathrooms: projectDetail.bathrooms,
                lotSize: projectDetail.lot_size,
                foundationType: projectDetail.foundation_type,
                startDate: projectDetail.start_date,
                projectedEndDate: projectDetail.projected_end_date,
            };

            this.emit('project-created', project);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to create project';
        } finally {
            this._loading = false;
        }
    }

    override render(): TemplateResult {
        return html`
            <form @submit=${this._handleSubmit.bind(this)}>
                ${this._error ? html`
                    <div class="error-message" role="alert">
                        <span aria-hidden="true">!</span>
                        ${this._error}
                    </div>
                ` : nothing}

                <div class="field">
                    <label for="name" class="required">Project Name</label>
                    <input
                        id="name"
                        type="text"
                        .value=${this._name}
                        @input=${this._handleNameInput.bind(this)}
                        placeholder="e.g., Smith Residence"
                        required
                        ?disabled=${this._loading}
                        autocomplete="off"
                    />
                </div>

                <div class="field">
                    <label for="address" class="required">Address</label>
                    <input
                        id="address"
                        type="text"
                        .value=${this._address}
                        @input=${this._handleAddressInput.bind(this)}
                        placeholder="e.g., 123 Main St, Austin, TX"
                        required
                        ?disabled=${this._loading}
                        autocomplete="street-address"
                    />
                </div>

                <div class="field">
                    <label for="sqft">Square Footage</label>
                    <input
                        id="sqft"
                        type="number"
                        .value=${String(this._squareFootage)}
                        @input=${this._handleSquareFootageInput.bind(this)}
                        min="0"
                        max="100000"
                        ?disabled=${this._loading}
                    />
                </div>

                <div class="actions">
                    <button
                        type="button"
                        class="btn-cancel"
                        @click=${this._handleCancel.bind(this)}
                        ?disabled=${this._loading}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        class="btn-submit"
                        ?disabled=${this._loading}
                    >
                        ${this._loading ? html`<span class="spinner"></span>Creating...` : 'Create Project'}
                    </button>
                </div>
            </form>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-project-form': FBProjectForm;
    }
}

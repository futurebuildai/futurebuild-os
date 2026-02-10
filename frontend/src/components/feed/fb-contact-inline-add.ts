/**
 * fb-contact-inline-add — Minimal name+phone inline form for feed cards and phase grid.
 * See FRONTEND_V2_SPEC.md §10.3.E
 *
 * Used inside `setup_contacts` feed cards and `fb-contact-phase-grid`.
 * Two fields, one tap — designed for maximum speed.
 */
import { html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import type { Contact } from '../../types/models';

@customElement('fb-contact-inline-add')
export class FBContactInlineAdd extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .form {
                display: flex;
                flex-direction: column;
                gap: 8px;
            }

            .row {
                display: flex;
                gap: 8px;
            }

            input {
                flex: 1;
                padding: 8px 12px;
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 6px;
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 14px;
                outline: none;
            }

            input:focus {
                border-color: var(--fb-accent, #6366f1);
            }

            input::placeholder {
                color: var(--fb-text-tertiary, #707080);
            }

            .preference-row {
                display: flex;
                align-items: center;
                gap: 12px;
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .preference-row label {
                display: flex;
                align-items: center;
                gap: 4px;
                cursor: pointer;
            }

            .btn-row {
                display: flex;
                gap: 8px;
                margin-top: 4px;
            }

            button {
                padding: 8px 16px;
                border-radius: 6px;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
                border: none;
                transition: opacity 0.15s ease;
            }

            button:hover {
                opacity: 0.9;
            }

            .btn-save {
                background: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .btn-save:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .btn-cancel {
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                border: 1px solid var(--fb-border, #2a2a3e);
            }

            .suggestions {
                display: flex;
                flex-wrap: wrap;
                gap: 6px;
                margin-bottom: 4px;
            }

            .suggestion {
                display: inline-flex;
                align-items: center;
                padding: 4px 10px;
                border-radius: 16px;
                font-size: 12px;
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-secondary, #a0a0b0);
                border: 1px solid var(--fb-border, #2a2a3e);
                cursor: pointer;
            }

            .suggestion:hover {
                border-color: var(--fb-accent, #6366f1);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .error {
                font-size: 12px;
                color: #ef4444;
            }
        `,
    ];

    /** Project ID for the assignment (if assigning to a phase) */
    @property({ type: String }) projectId = '';
    /** WBS phase ID for auto-assignment after save */
    @property({ type: String }) phaseId = '';
    /** Label shown above the form */
    @property({ type: String }) label = '';

    @state() private _name = '';
    @state() private _phone = '';
    @state() private _email = '';
    @state() private _preference: 'SMS' | 'Email' | 'Both' = 'SMS';
    @state() private _saving = false;
    @state() private _error = '';
    @state() private _suggestions: Contact[] = [];

    override connectedCallback() {
        super.connectedCallback();
        this._loadSuggestions();
    }

    private async _loadSuggestions() {
        try {
            const contacts = await api.contacts.list();
            this._suggestions = contacts.slice(0, 5);
        } catch {
            // Suggestions are optional
        }
    }

    private _handleSuggestionClick(contact: Contact) {
        this.emit('fb-contact-saved', {
            contact,
            projectId: this.projectId,
            phaseId: this.phaseId,
            isExisting: true,
        });
    }

    private async _handleSave() {
        if (!this._name.trim()) return;
        if (!this._phone.trim() && !this._email.trim()) {
            this._error = 'Phone or email is required';
            return;
        }

        this._saving = true;
        this._error = '';

        try {
            const req: Parameters<typeof api.contacts.create>[0] = {
                name: this._name.trim(),
                contact_preference: this._preference,
            };
            if (this._phone.trim()) req.phone = this._phone.trim();
            if (this._email.trim()) req.email = this._email.trim();
            const resp = await api.contacts.create(req);

            // Auto-assign to phase if projectId and phaseId are set
            if (this.projectId && this.phaseId) {
                await api.contacts.createAssignment(this.projectId, resp.contact.id, this.phaseId);
            }

            this.emit('fb-contact-saved', {
                contact: resp.contact,
                matched: resp.matched,
                projectId: this.projectId,
                phaseId: this.phaseId,
                isExisting: false,
            });

            // Reset form
            this._name = '';
            this._phone = '';
            this._email = '';
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to save contact';
        } finally {
            this._saving = false;
        }
    }

    private _handleCancel() {
        this.emit('fb-contact-cancelled', {});
    }

    private get _isValid(): boolean {
        return this._name.trim().length > 0 && (this._phone.trim().length > 0 || this._email.trim().length > 0);
    }

    override render() {
        return html`
            ${this._suggestions.length > 0
                ? html`
                      <div class="suggestions">
                          ${this._suggestions.map(
                              (c) => html`
                                  <button class="suggestion" @click=${() => this._handleSuggestionClick(c)}>
                                      ${c.name}${c.phone ? ` \u2014 ${c.phone}` : ''}
                                  </button>
                              `
                          )}
                      </div>
                  `
                : nothing}

            <div class="form">
                <div class="row">
                    <input
                        type="text"
                        placeholder="Name"
                        .value=${this._name}
                        @input=${(e: InputEvent) => { this._name = (e.target as HTMLInputElement).value; }}
                    />
                    <input
                        type="tel"
                        placeholder="Phone"
                        .value=${this._phone}
                        @input=${(e: InputEvent) => { this._phone = (e.target as HTMLInputElement).value; }}
                    />
                </div>

                <div class="preference-row">
                    Contact via:
                    <label>
                        <input type="radio" name="pref" value="SMS"
                            .checked=${this._preference === 'SMS'}
                            @change=${() => { this._preference = 'SMS'; }} />
                        SMS
                    </label>
                    <label>
                        <input type="radio" name="pref" value="Email"
                            .checked=${this._preference === 'Email'}
                            @change=${() => { this._preference = 'Email'; }} />
                        Email
                    </label>
                    <label>
                        <input type="radio" name="pref" value="Both"
                            .checked=${this._preference === 'Both'}
                            @change=${() => { this._preference = 'Both'; }} />
                        Both
                    </label>
                </div>

                ${this._error ? html`<div class="error">${this._error}</div>` : nothing}

                <div class="btn-row">
                    <button class="btn-save" ?disabled=${!this._isValid || this._saving} @click=${this._handleSave}>
                        ${this._saving ? 'Saving...' : 'Save & Assign'}
                    </button>
                    <button class="btn-cancel" @click=${this._handleCancel}>Cancel</button>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-contact-inline-add': FBContactInlineAdd;
    }
}

/**
 * fb-contact-phase-grid — Phase-by-phase contact assignment grid.
 * See FRONTEND_V2_SPEC.md §10.3.A
 *
 * Shows all WBS phases for the project with contact assignment slots.
 * Features:
 * - Inline add form with name + phone + contact preference
 * - Autocomplete from org directory
 * - Portal access toggle when email provided
 * - Trade auto-set from phase name
 */
import { html, css, nothing, type TemplateResult } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
// import { api } from '../../services/api'; // TODO: Wire up when backend is ready

export interface PhaseAssignment {
    phase_code: string;
    phase_name: string;
    contact: Contact | null;
}

export interface Contact {
    id: string;
    name: string;
    phone?: string | undefined;
    email?: string | undefined;
    company?: string | undefined;
    role: 'Subcontractor' | 'Client';
    contact_preference?: 'SMS' | 'Email' | 'Both' | undefined;
    portal_enabled?: boolean | undefined;
    trades?: string[] | undefined;
}

interface ContactFormData {
    name: string;
    phone: string;
    email: string;
    contactPreference: 'SMS' | 'Email' | 'Both';
    portalEnabled: boolean;
}

@customElement('fb-contact-phase-grid')
export class FBContactPhaseGrid extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .grid-container {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                overflow: hidden;
            }

            .grid-header {
                padding: 20px 24px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .grid-title {
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 6px;
            }

            .grid-subtitle {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .phase-list {
                list-style: none;
                margin: 0;
                padding: 0;
            }

            .phase-row {
                display: flex;
                align-items: center;
                padding: 16px 24px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
                min-height: 56px;
            }

            .phase-row:last-child {
                border-bottom: none;
            }

            .phase-row.expanded {
                flex-direction: column;
                align-items: stretch;
            }

            .phase-row-main {
                display: flex;
                align-items: center;
                width: 100%;
            }

            .phase-name {
                width: 140px;
                flex-shrink: 0;
                font-size: 14px;
                font-weight: 500;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .phase-content {
                flex: 1;
                display: flex;
                align-items: center;
                gap: 12px;
            }

            .add-btn {
                display: flex;
                align-items: center;
                gap: 6px;
                padding: 8px 16px;
                border-radius: 6px;
                border: 1px dashed var(--fb-border, #2a2a3e);
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 14px;
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .add-btn:hover {
                border-color: var(--fb-accent, #6366f1);
                color: var(--fb-accent, #6366f1);
            }

            .contact-info {
                display: flex;
                align-items: center;
                gap: 12px;
                flex: 1;
            }

            .contact-name {
                font-size: 14px;
                font-weight: 500;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .contact-phone {
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
                font-family: monospace;
            }

            .contact-badges {
                display: flex;
                gap: 6px;
            }

            .badge {
                padding: 2px 8px;
                border-radius: 4px;
                font-size: 10px;
                font-weight: 600;
                text-transform: uppercase;
            }

            .badge.portal {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .badge.sms {
                background: rgba(156, 163, 175, 0.15);
                color: #9ca3af;
            }

            .edit-btn, .remove-btn {
                padding: 6px 12px;
                border-radius: 4px;
                border: none;
                font-size: 12px;
                font-weight: 500;
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .edit-btn {
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .edit-btn:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .remove-btn {
                background: transparent;
                color: #ef4444;
            }

            .remove-btn:hover {
                background: rgba(239, 68, 68, 0.1);
            }

            /* Inline form */
            .inline-form {
                width: 100%;
                padding: 16px 0 0 0;
                margin-left: 140px;
                border-top: 1px solid var(--fb-border, #2a2a3e);
            }

            .form-grid {
                display: grid;
                grid-template-columns: 1fr 1fr;
                gap: 12px;
            }

            .form-group {
                display: flex;
                flex-direction: column;
                gap: 4px;
            }

            .form-group.full-width {
                grid-column: 1 / -1;
            }

            label {
                font-size: 11px;
                font-weight: 500;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            input {
                padding: 10px 12px;
                background: var(--fb-surface-2, #252540);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 6px;
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 14px;
            }

            input:focus {
                outline: none;
                border-color: var(--fb-accent, #6366f1);
            }

            input::placeholder {
                color: var(--fb-text-tertiary, #707080);
            }

            .radio-group {
                display: flex;
                gap: 16px;
                padding: 8px 0;
            }

            .radio-option {
                display: flex;
                align-items: center;
                gap: 6px;
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
                cursor: pointer;
            }

            .radio-option input[type="radio"] {
                width: auto;
                padding: 0;
            }

            .checkbox-group {
                display: flex;
                align-items: center;
                gap: 8px;
                padding: 8px 0;
            }

            .checkbox-group input[type="checkbox"] {
                width: auto;
                padding: 0;
            }

            .checkbox-label {
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .checkbox-help {
                font-size: 11px;
                color: var(--fb-text-tertiary, #707080);
                margin-top: 4px;
            }

            .form-actions {
                display: flex;
                gap: 12px;
                margin-top: 16px;
            }

            .btn {
                padding: 8px 20px;
                border-radius: 6px;
                font-size: 14px;
                font-weight: 600;
                cursor: pointer;
                border: none;
                transition: all 0.15s ease;
            }

            .btn-primary {
                background: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .btn-primary:hover:not(:disabled) {
                opacity: 0.9;
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
                background: var(--fb-surface-2, #252540);
            }

            /* Suggestions */
            .suggestions {
                margin-bottom: 12px;
            }

            .suggestions-label {
                font-size: 11px;
                color: var(--fb-text-tertiary, #707080);
                margin-bottom: 8px;
            }

            .suggestion-chips {
                display: flex;
                flex-wrap: wrap;
                gap: 8px;
            }

            .suggestion-chip {
                display: flex;
                align-items: center;
                gap: 6px;
                padding: 6px 12px;
                background: var(--fb-surface-2, #252540);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 6px;
                font-size: 13px;
                color: var(--fb-text-primary, #e0e0e0);
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .suggestion-chip:hover {
                border-color: var(--fb-accent, #6366f1);
            }

            .suggestion-phone {
                font-size: 11px;
                color: var(--fb-text-tertiary, #707080);
                font-family: monospace;
            }

            /* Footer */
            .grid-footer {
                display: flex;
                justify-content: flex-end;
                padding: 16px 24px;
                border-top: 1px solid var(--fb-border, #2a2a3e);
            }

            .loading {
                padding: 40px;
                text-align: center;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .error {
                padding: 16px 24px;
                background: rgba(239, 68, 68, 0.1);
                color: #ef4444;
                font-size: 14px;
            }

            @media (max-width: 600px) {
                .phase-name {
                    width: 100px;
                }

                .inline-form {
                    margin-left: 0;
                }

                .form-grid {
                    grid-template-columns: 1fr;
                }
            }
        `,
    ];

    /** Project ID for assignment */
    @property({ type: String }) projectId = '';

    /** Project address for header */
    @property({ type: String }) projectAddress = '';

    @state() private _loading = true;
    @state() private _saving = false;
    @state() private _error = '';
    @state() private _phases: PhaseAssignment[] = [];
    @state() private _expandedPhase: string | null = null;
    @state() private _suggestions: Contact[] = [];
    @state() private _formData: ContactFormData = this._getEmptyForm();

    private _getEmptyForm(): ContactFormData {
        return {
            name: '',
            phone: '',
            email: '',
            contactPreference: 'SMS',
            portalEnabled: false,
        };
    }

    override async connectedCallback() {
        super.connectedCallback();
        if (this.projectId) {
            await this._loadAssignments();
            await this._loadSuggestions();
        }
    }

    private async _loadAssignments() {
        this._loading = true;
        this._error = '';
        try {
            // TODO: Replace with actual API call
            // this._phases = await api.contacts.getAssignments(this.projectId);

            // Mock data for now
            this._phases = [
                { phase_code: '3.0', phase_name: 'Foundation', contact: null },
                { phase_code: '7.0', phase_name: 'Framing', contact: { id: '1', name: 'Rodriguez Framing', phone: '555-0101', role: 'Subcontractor', contact_preference: 'SMS' } },
                { phase_code: '8.0', phase_name: 'Roofing', contact: null },
                { phase_code: '9.0', phase_name: 'Electrical', contact: null },
                { phase_code: '9.1', phase_name: 'Plumbing', contact: null },
                { phase_code: '9.2', phase_name: 'HVAC', contact: null },
                { phase_code: '10.0', phase_name: 'Insulation', contact: null },
                { phase_code: '11.0', phase_name: 'Drywall', contact: null },
                { phase_code: '12.0', phase_name: 'Finishes', contact: null },
            ];
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to load assignments';
        } finally {
            this._loading = false;
        }
    }

    private async _loadSuggestions() {
        try {
            // TODO: Replace with actual API call
            // this._suggestions = await api.contacts.search('');

            this._suggestions = [
                { id: '1', name: 'Rodriguez Framing', phone: '555-0101', role: 'Subcontractor' },
                { id: '2', name: "Jake's Electric", phone: '555-0199', role: 'Subcontractor' },
                { id: '3', name: "Mike's Plumbing", phone: '555-0201', role: 'Subcontractor' },
            ];
        } catch {
            // Suggestions are optional, don't show error
        }
    }

    private _handleAddClick(phaseCode: string) {
        this._expandedPhase = phaseCode;
        this._formData = this._getEmptyForm();
    }

    private _handleCancelForm() {
        this._expandedPhase = null;
        this._formData = this._getEmptyForm();
    }

    private _handleInputChange(field: keyof ContactFormData, value: string | boolean) {
        this._formData = { ...this._formData, [field]: value };

        // Auto-set contact preference based on phone/email
        if (field === 'phone' && value && !this._formData.email) {
            this._formData.contactPreference = 'SMS';
        } else if (field === 'email' && value && !this._formData.phone) {
            this._formData.contactPreference = 'Email';
        }
    }

    private _handleSuggestionClick(contact: Contact) {
        // Assign existing contact to phase
        this._assignContact(contact);
    }

    private async _handleSaveContact() {
        if (!this._formData.name.trim() || (!this._formData.phone.trim() && !this._formData.email.trim())) {
            this._error = 'Name and phone or email are required';
            return;
        }

        this._saving = true;
        this._error = '';

        try {
            // TODO: Replace with actual API call
            // const contact = await api.contacts.create({ ... });
            // await api.contacts.assign(this.projectId, this._expandedPhase, contact.id);

            // Mock: Update local state
            const phase = this._phases.find((p) => p.phase_code === this._expandedPhase);
            if (phase) {
                phase.contact = {
                    id: Date.now().toString(),
                    name: this._formData.name,
                    phone: this._formData.phone || undefined,
                    email: this._formData.email || undefined,
                    role: 'Subcontractor',
                    contact_preference: this._formData.contactPreference,
                    portal_enabled: this._formData.portalEnabled,
                    trades: [phase.phase_name],
                };
            }

            this._phases = [...this._phases];
            this._expandedPhase = null;
            this._formData = this._getEmptyForm();

            this.emit('fb-contact-saved', { phaseCode: phase?.phase_code, contact: phase?.contact });
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to save contact';
        } finally {
            this._saving = false;
        }
    }

    private async _assignContact(contact: Contact) {
        if (!this._expandedPhase) return;

        this._saving = true;
        try {
            // TODO: Replace with actual API call
            // await api.contacts.assign(this.projectId, this._expandedPhase, contact.id);

            const phase = this._phases.find((p) => p.phase_code === this._expandedPhase);
            if (phase) {
                phase.contact = contact;
            }
            this._phases = [...this._phases];
            this._expandedPhase = null;

            this.emit('fb-contact-assigned', { phaseCode: phase?.phase_code, contact });
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to assign contact';
        } finally {
            this._saving = false;
        }
    }

    private async _handleRemoveContact(phaseCode: string) {
        try {
            // TODO: Replace with actual API call
            // await api.contacts.unassign(this.projectId, phaseCode);

            const phase = this._phases.find((p) => p.phase_code === phaseCode);
            if (phase) {
                phase.contact = null;
            }
            this._phases = [...this._phases];

            this.emit('fb-contact-removed', { phaseCode });
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to remove contact';
        }
    }

    private _handleDone() {
        this.emit('fb-phase-grid-done', { projectId: this.projectId });
    }

    private _renderPhaseRow(phase: PhaseAssignment): TemplateResult {
        const isExpanded = this._expandedPhase === phase.phase_code;
        const availableSuggestions = this._suggestions.filter(
            (s) => !this._phases.some((p) => p.contact?.id === s.id)
        );

        return html`
            <li class="phase-row ${isExpanded ? 'expanded' : ''}">
                <div class="phase-row-main">
                    <span class="phase-name">${phase.phase_name}</span>
                    <div class="phase-content">
                        ${phase.contact
                            ? this._renderContact(phase)
                            : html`
                                  <button class="add-btn" @click=${() => this._handleAddClick(phase.phase_code)}>
                                      + Add contact
                                  </button>
                              `}
                    </div>
                </div>
                ${isExpanded
                    ? html`
                          <div class="inline-form">
                              ${availableSuggestions.length > 0
                                  ? html`
                                        <div class="suggestions">
                                            <div class="suggestions-label">From your directory:</div>
                                            <div class="suggestion-chips">
                                                ${availableSuggestions.map(
                                                    (s) => html`
                                                        <button class="suggestion-chip" @click=${() => this._handleSuggestionClick(s)}>
                                                            ${s.name}
                                                            ${s.phone ? html`<span class="suggestion-phone">${s.phone}</span>` : nothing}
                                                        </button>
                                                    `
                                                )}
                                            </div>
                                        </div>
                                    `
                                  : nothing}

                              <div class="form-grid">
                                  <div class="form-group">
                                      <label>Name</label>
                                      <input
                                          type="text"
                                          placeholder="Contact name"
                                          .value=${this._formData.name}
                                          @input=${(e: Event) => this._handleInputChange('name', (e.target as HTMLInputElement).value)}
                                          ?disabled=${this._saving}
                                      />
                                  </div>
                                  <div class="form-group">
                                      <label>Phone</label>
                                      <input
                                          type="tel"
                                          placeholder="555-0123"
                                          .value=${this._formData.phone}
                                          @input=${(e: Event) => this._handleInputChange('phone', (e.target as HTMLInputElement).value)}
                                          ?disabled=${this._saving}
                                      />
                                  </div>
                                  <div class="form-group full-width">
                                      <label>Email (optional)</label>
                                      <input
                                          type="email"
                                          placeholder="contact@example.com"
                                          .value=${this._formData.email}
                                          @input=${(e: Event) => this._handleInputChange('email', (e.target as HTMLInputElement).value)}
                                          ?disabled=${this._saving}
                                      />
                                  </div>
                                  <div class="form-group full-width">
                                      <label>Contact via</label>
                                      <div class="radio-group">
                                          <label class="radio-option">
                                              <input
                                                  type="radio"
                                                  name="pref-${phase.phase_code}"
                                                  .checked=${this._formData.contactPreference === 'SMS'}
                                                  @change=${() => this._handleInputChange('contactPreference', 'SMS')}
                                                  ?disabled=${this._saving}
                                              />
                                              SMS
                                          </label>
                                          <label class="radio-option">
                                              <input
                                                  type="radio"
                                                  name="pref-${phase.phase_code}"
                                                  .checked=${this._formData.contactPreference === 'Email'}
                                                  @change=${() => this._handleInputChange('contactPreference', 'Email')}
                                                  ?disabled=${this._saving}
                                              />
                                              Email
                                          </label>
                                          <label class="radio-option">
                                              <input
                                                  type="radio"
                                                  name="pref-${phase.phase_code}"
                                                  .checked=${this._formData.contactPreference === 'Both'}
                                                  @change=${() => this._handleInputChange('contactPreference', 'Both')}
                                                  ?disabled=${this._saving}
                                              />
                                              Both
                                          </label>
                                      </div>
                                  </div>
                                  ${this._formData.email
                                      ? html`
                                            <div class="form-group full-width">
                                                <div class="checkbox-group">
                                                    <input
                                                        type="checkbox"
                                                        id="portal-${phase.phase_code}"
                                                        .checked=${this._formData.portalEnabled}
                                                        @change=${(e: Event) => this._handleInputChange('portalEnabled', (e.target as HTMLInputElement).checked)}
                                                        ?disabled=${this._saving}
                                                    />
                                                    <label for="portal-${phase.phase_code}" class="checkbox-label">
                                                        Grant portal access
                                                    </label>
                                                </div>
                                                <div class="checkbox-help">
                                                    Portal contacts can view schedules, upload documents, and message you through FutureBuild.
                                                </div>
                                            </div>
                                        `
                                      : nothing}
                              </div>
                              <div class="form-actions">
                                  <button class="btn btn-secondary" @click=${this._handleCancelForm} ?disabled=${this._saving}>
                                      Cancel
                                  </button>
                                  <button
                                      class="btn btn-primary"
                                      @click=${this._handleSaveContact}
                                      ?disabled=${this._saving || !this._formData.name.trim()}
                                  >
                                      ${this._saving ? 'Saving...' : 'Save'}
                                  </button>
                              </div>
                          </div>
                      `
                    : nothing}
            </li>
        `;
    }

    private _renderContact(phase: PhaseAssignment): TemplateResult {
        const contact = phase.contact!;
        return html`
            <div class="contact-info">
                <span class="contact-name">${contact.name}</span>
                ${contact.phone ? html`<span class="contact-phone">${contact.phone}</span>` : nothing}
                <div class="contact-badges">
                    ${contact.portal_enabled
                        ? html`<span class="badge portal">Portal</span>`
                        : html`<span class="badge sms">${contact.contact_preference ?? 'SMS'}</span>`}
                </div>
            </div>
            <button class="edit-btn" @click=${() => this._handleAddClick(phase.phase_code)}>
                Edit
            </button>
            <button class="remove-btn" @click=${() => this._handleRemoveContact(phase.phase_code)}>
                Remove
            </button>
        `;
    }

    override render(): TemplateResult {
        if (this._loading) {
            return html`
                <div class="grid-container">
                    <div class="loading">Loading contacts...</div>
                </div>
            `;
        }

        return html`
            <div class="grid-container">
                <div class="grid-header">
                    <div class="grid-title">Contacts — ${this.projectAddress || 'Project'}</div>
                    <div class="grid-subtitle">
                        Assign your subs to each trade. I'll handle confirmations, reminders, and status checks.
                    </div>
                </div>

                ${this._error ? html`<div class="error">${this._error}</div>` : nothing}

                <ul class="phase-list">
                    ${this._phases.map((phase) => this._renderPhaseRow(phase))}
                </ul>

                <div class="grid-footer">
                    <button class="btn btn-primary" @click=${this._handleDone}>
                        Done
                    </button>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-contact-phase-grid': FBContactPhaseGrid;
    }
}

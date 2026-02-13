/**
 * fb-contact-bulk-input — Bulk paste contact input.
 * See FRONTEND_V2_SPEC.md §10.3.B
 *
 * Allows pasting a list of contacts in various formats:
 * - Comma-separated, tab-separated, or natural text
 * - Parses phone numbers and trade hints
 * - Review step before saving
 */
import { html, css, nothing, type TemplateResult } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

interface ParsedContact {
    id: string;
    name: string;
    phone: string;
    trade: string;
    tradeMatch: string | undefined;  // Matched phase name
    valid: boolean;
    error: string | undefined;
}

const TRADE_MAPPINGS: Record<string, string> = {
    'electric': 'Electrical',
    'electrical': 'Electrical',
    'plumb': 'Plumbing',
    'plumbing': 'Plumbing',
    'hvac': 'HVAC',
    'heating': 'HVAC',
    'cooling': 'HVAC',
    'fram': 'Framing',
    'framing': 'Framing',
    'roof': 'Roofing',
    'roofing': 'Roofing',
    'foundation': 'Foundation',
    'concrete': 'Foundation',
    'insulation': 'Insulation',
    'drywall': 'Drywall',
    'finish': 'Finishes',
    'finishes': 'Finishes',
    'paint': 'Finishes',
    'flooring': 'Finishes',
    'tile': 'Finishes',
};

@customElement('fb-contact-bulk-input')
export class FBContactBulkInput extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .container {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                overflow: hidden;
            }

            .header {
                padding: 20px 24px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .title {
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 6px;
            }

            .subtitle {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .content {
                padding: 24px;
            }

            /* Input step */
            .textarea-container {
                margin-bottom: 16px;
            }

            textarea {
                width: 100%;
                height: 200px;
                padding: 16px;
                background: var(--fb-surface-2, #252540);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 8px;
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 14px;
                font-family: monospace;
                line-height: 1.6;
                resize: vertical;
                box-sizing: border-box;
            }

            textarea:focus {
                outline: none;
                border-color: var(--fb-accent, #6366f1);
            }

            textarea::placeholder {
                color: var(--fb-text-tertiary, #707080);
            }

            .format-help {
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
                margin-top: 8px;
            }

            /* Review step */
            .review-table {
                width: 100%;
                border-collapse: collapse;
                margin-bottom: 16px;
            }

            .review-table th {
                text-align: left;
                padding: 12px 16px;
                background: var(--fb-surface-2, #252540);
                font-size: 11px;
                font-weight: 600;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
                letter-spacing: 0.05em;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .review-table td {
                padding: 12px 16px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
                font-size: 14px;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .review-table tr:last-child td {
                border-bottom: none;
            }

            .review-table tr.invalid {
                background: rgba(239, 68, 68, 0.1);
            }

            .review-table tr.invalid td {
                color: #ef4444;
            }

            .trade-select {
                padding: 6px 10px;
                background: var(--fb-surface-2, #252540);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 4px;
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 13px;
                min-width: 120px;
            }

            .trade-select:focus {
                outline: none;
                border-color: var(--fb-accent, #6366f1);
            }

            .trade-badge {
                display: inline-block;
                padding: 4px 10px;
                background: rgba(99, 102, 241, 0.15);
                color: var(--fb-accent, #6366f1);
                border-radius: 4px;
                font-size: 12px;
                font-weight: 500;
            }

            .trade-badge.no-match {
                background: rgba(245, 158, 11, 0.15);
                color: #f59e0b;
            }

            .remove-row-btn {
                padding: 4px 8px;
                background: transparent;
                border: none;
                color: var(--fb-text-tertiary, #707080);
                cursor: pointer;
                font-size: 16px;
                transition: color 0.15s;
            }

            .remove-row-btn:hover {
                color: #ef4444;
            }

            .summary {
                padding: 16px;
                background: var(--fb-surface-2, #252540);
                border-radius: 8px;
                margin-bottom: 16px;
            }

            .summary-row {
                display: flex;
                justify-content: space-between;
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .summary-row + .summary-row {
                margin-top: 8px;
            }

            .summary-value {
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .summary-value.warning {
                color: #f59e0b;
            }

            .summary-value.error {
                color: #ef4444;
            }

            /* Actions */
            .actions {
                display: flex;
                justify-content: flex-end;
                gap: 12px;
                padding: 16px 24px;
                border-top: 1px solid var(--fb-border, #2a2a3e);
            }

            .btn {
                padding: 10px 24px;
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

            .error-banner {
                padding: 12px 16px;
                background: rgba(239, 68, 68, 0.1);
                color: #ef4444;
                font-size: 14px;
                margin-bottom: 16px;
                border-radius: 8px;
            }

            .success-banner {
                padding: 12px 16px;
                background: rgba(34, 197, 94, 0.1);
                color: #22c55e;
                font-size: 14px;
                margin-bottom: 16px;
                border-radius: 8px;
            }
        `,
    ];

    /** Available phases for assignment */
    @property({ attribute: false })
    phases: Array<{ code: string; name: string }> = [];

    @state() private _step: 'input' | 'review' | 'success' = 'input';
    @state() private _rawInput = '';
    @state() private _parsedContacts: ParsedContact[] = [];
    @state() private _saving = false;
    @state() private _error = '';

    private _parseInput() {
        const lines = this._rawInput.trim().split('\n').filter((l) => l.trim());
        const parsed: ParsedContact[] = [];

        for (const line of lines) {
            const contact = this._parseLine(line);
            parsed.push(contact);
        }

        this._parsedContacts = parsed;
        this._step = 'review';
    }

    private _parseLine(line: string): ParsedContact {
        const id = Math.random().toString(36).slice(2);

        // Look for phone number pattern
        const phoneMatch = line.match(/(?:\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})/);
        let phone = '';
        let remaining = line;

        if (phoneMatch) {
            phone = `${phoneMatch[1]}-${phoneMatch[2]}-${phoneMatch[3]}`;
            remaining = line.replace(phoneMatch[0], '').trim();
        }

        // Split remaining by common delimiters
        const parts = remaining.split(/[,\t|]/).map((p) => p.trim()).filter(Boolean);
        const name = parts[0] ?? '';
        const tradeHint = parts[1] ?? '';

        // Try to match trade
        let tradeMatch: string | undefined;
        const searchTerms = [name.toLowerCase(), tradeHint.toLowerCase()].join(' ');

        for (const [keyword, tradeName] of Object.entries(TRADE_MAPPINGS)) {
            if (searchTerms.includes(keyword)) {
                tradeMatch = tradeName;
                break;
            }
        }

        // Validation
        const valid = name.length > 0 && (phone.length > 0 || tradeHint.includes('@'));
        const error = !valid ? 'Name and phone required' : undefined;

        return {
            id,
            name,
            phone,
            trade: tradeHint,
            tradeMatch,
            valid,
            error,
        };
    }

    private _handleRemoveRow(id: string) {
        this._parsedContacts = this._parsedContacts.filter((c) => c.id !== id);
    }

    private _handleTradeChange(id: string, tradeName: string) {
        this._parsedContacts = this._parsedContacts.map((c) =>
            c.id === id ? { ...c, tradeMatch: tradeName } : c
        );
    }

    private async _handleSaveAll() {
        const validContacts = this._parsedContacts.filter((c) => c.valid);
        if (validContacts.length === 0) {
            this._error = 'No valid contacts to save';
            return;
        }

        this._saving = true;
        this._error = '';

        try {
            // TODO: Replace with actual API call
            // await api.contacts.bulkCreate(validContacts);

            await new Promise((resolve) => setTimeout(resolve, 1000)); // Simulate API call

            this._step = 'success';
            this.emit('fb-bulk-contacts-saved', { count: validContacts.length, contacts: validContacts });
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to save contacts';
        } finally {
            this._saving = false;
        }
    }

    private _handleBack() {
        this._step = 'input';
    }

    private _handleClose() {
        this.emit('fb-bulk-input-close');
    }

    private _handleAddMore() {
        this._step = 'input';
        this._rawInput = '';
        this._parsedContacts = [];
    }

    private _renderInputStep(): TemplateResult {
        return html`
            <div class="content">
                <div class="textarea-container">
                    <textarea
                        placeholder="Paste your contact list here. One per line.

Example:
Jake Williams, 555-0199, Electrical
Rodriguez Framing, 555-0101
Mike's Plumbing, (555) 020-1234"
                        .value=${this._rawInput}
                        @input=${(e: Event) => { this._rawInput = (e.target as HTMLTextAreaElement).value; }}
                    ></textarea>
                    <div class="format-help">
                        Format: Name, Phone, Trade (optional). Accepts comma or tab separated.
                    </div>
                </div>
            </div>
            <div class="actions">
                <button class="btn btn-secondary" @click=${this._handleClose}>Cancel</button>
                <button
                    class="btn btn-primary"
                    @click=${this._parseInput}
                    ?disabled=${!this._rawInput.trim()}
                >
                    Parse & Review
                </button>
            </div>
        `;
    }

    private _renderReviewStep(): TemplateResult {
        const validCount = this._parsedContacts.filter((c) => c.valid).length;
        const invalidCount = this._parsedContacts.length - validCount;
        const unmatchedCount = this._parsedContacts.filter((c) => c.valid && !c.tradeMatch).length;

        return html`
            <div class="content">
                ${this._error ? html`<div class="error-banner">${this._error}</div>` : nothing}

                <div class="summary">
                    <div class="summary-row">
                        <span>Total parsed</span>
                        <span class="summary-value">${this._parsedContacts.length}</span>
                    </div>
                    <div class="summary-row">
                        <span>Ready to save</span>
                        <span class="summary-value">${validCount}</span>
                    </div>
                    ${invalidCount > 0
                        ? html`
                              <div class="summary-row">
                                  <span>Invalid (missing data)</span>
                                  <span class="summary-value error">${invalidCount}</span>
                              </div>
                          `
                        : nothing}
                    ${unmatchedCount > 0
                        ? html`
                              <div class="summary-row">
                                  <span>No trade match (assign manually)</span>
                                  <span class="summary-value warning">${unmatchedCount}</span>
                              </div>
                          `
                        : nothing}
                </div>

                <table class="review-table">
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Phone</th>
                            <th>Trade</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        ${this._parsedContacts.map(
                            (contact) => html`
                                <tr class="${contact.valid ? '' : 'invalid'}">
                                    <td>${contact.name || '—'}</td>
                                    <td>${contact.phone || '—'}</td>
                                    <td>
                                        ${contact.valid
                                            ? contact.tradeMatch
                                                ? html`<span class="trade-badge">${contact.tradeMatch}</span>`
                                                : html`
                                                      <select
                                                          class="trade-select"
                                                          @change=${(e: Event) =>
                                                              this._handleTradeChange(
                                                                  contact.id,
                                                                  (e.target as HTMLSelectElement).value
                                                              )}
                                                      >
                                                          <option value="">Assign trade...</option>
                                                          ${this.phases.map(
                                                              (p) => html`<option value=${p.name}>${p.name}</option>`
                                                          )}
                                                      </select>
                                                  `
                                            : html`<span style="color: #ef4444">${contact.error}</span>`}
                                    </td>
                                    <td>
                                        <button
                                            class="remove-row-btn"
                                            @click=${() => this._handleRemoveRow(contact.id)}
                                            title="Remove"
                                        >
                                            &times;
                                        </button>
                                    </td>
                                </tr>
                            `
                        )}
                    </tbody>
                </table>
            </div>
            <div class="actions">
                <button class="btn btn-secondary" @click=${this._handleBack} ?disabled=${this._saving}>
                    Back
                </button>
                <button
                    class="btn btn-primary"
                    @click=${this._handleSaveAll}
                    ?disabled=${this._saving || validCount === 0}
                >
                    ${this._saving ? 'Saving...' : `Save ${validCount} Contacts`}
                </button>
            </div>
        `;
    }

    private _renderSuccessStep(): TemplateResult {
        const savedCount = this._parsedContacts.filter((c) => c.valid).length;

        return html`
            <div class="content">
                <div class="success-banner">
                    Successfully saved ${savedCount} contact${savedCount !== 1 ? 's' : ''} to your directory.
                </div>
            </div>
            <div class="actions">
                <button class="btn btn-secondary" @click=${this._handleAddMore}>Add More</button>
                <button class="btn btn-primary" @click=${this._handleClose}>Done</button>
            </div>
        `;
    }

    override render(): TemplateResult {
        return html`
            <div class="container">
                <div class="header">
                    <div class="title">Bulk Add Contacts</div>
                    <div class="subtitle">
                        ${this._step === 'input'
                            ? 'Paste your contact list. One per line.'
                            : this._step === 'review'
                            ? 'Review and assign trades before saving.'
                            : 'Contacts saved successfully.'}
                    </div>
                </div>

                ${this._step === 'input' ? this._renderInputStep() : nothing}
                ${this._step === 'review' ? this._renderReviewStep() : nothing}
                ${this._step === 'success' ? this._renderSuccessStep() : nothing}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-contact-bulk-input': FBContactBulkInput;
    }
}

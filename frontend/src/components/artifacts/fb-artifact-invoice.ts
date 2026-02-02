import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { InvoiceArtifactData } from '../../types/artifacts';
import { InvoiceStatus } from '../../types/enums';
import { api } from '../../services/api';
import './fb-artifact-actions'; // Register <fb-artifact-actions>

/**
 * Draft line item for local editing state.
 * Mirrors InvoiceExtractionItem but uses number types for input binding.
 */
interface DraftLineItem {
    description: string;
    quantity: number;
    unitPriceCents: number;
    totalCents: number;
}

@customElement('fb-artifact-invoice')
export class FBArtifactInvoice extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
                background: white;
                color: black;
                border-radius: var(--fb-radius-md);
                border: 1px solid var(--fb-border);
                overflow: hidden;
                position: relative;
            }

            .invoice-container {
                padding: var(--fb-spacing-lg);
                font-family: 'Courier New', Courier, monospace;
            }

            .header {
                display: flex;
                justify-content: space-between;
                margin-bottom: var(--fb-spacing-xl);
                border-bottom: 2px solid #000;
                padding-bottom: var(--fb-spacing-md);
            }

            .vendor-info h2 {
                margin: 0 0 8px 0;
                font-size: 18px;
                text-transform: uppercase;
            }

            .invoice-meta {
                text-align: right;
            }

            .meta-row {
                margin-bottom: 4px;
                font-size: 12px;
            }

            .label { font-weight: bold; margin-right: 8px; }

            .line-items {
                width: 100%;
                border-collapse: collapse;
                margin-bottom: var(--fb-spacing-xl);
                font-size: 13px;
            }

            .line-items th {
                text-align: left;
                border-bottom: 1px solid #000;
                padding: 8px 4px;
                text-transform: uppercase;
            }

            .line-items td {
                padding: 8px 4px;
                border-bottom: 1px dashed #ccc;
            }

            .col-desc { width: 45%; }
            .col-qty { width: 10%; text-align: center; }
            .col-price { width: 15%; text-align: right; }
            .col-total { width: 15%; text-align: right; }
            .col-actions { width: 15%; text-align: center; }

            /* Edit Mode Inputs */
            .edit-input {
                width: 100%;
                padding: 4px 6px;
                border: 1px solid #ccc;
                border-radius: 3px;
                font-family: inherit;
                font-size: 13px;
                box-sizing: border-box;
                background: #fefefe;
            }

            .edit-input:focus {
                outline: none;
                border-color: var(--fb-accent, #2563eb);
                box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.15);
            }

            .edit-input[type="number"] {
                text-align: right;
                -moz-appearance: textfield;
            }

            .edit-input[type="number"]::-webkit-inner-spin-button,
            .edit-input[type="number"]::-webkit-outer-spin-button {
                -webkit-appearance: none;
                margin: 0;
            }

            .total-section {
                display: flex;
                justify-content: flex-end;
                margin-top: var(--fb-spacing-lg);
            }

            .total-box {
                width: 200px;
                border-top: 2px solid #000;
                padding-top: var(--fb-spacing-sm);
            }

            .total-row {
                display: flex;
                justify-content: space-between;
                margin-bottom: 4px;
                font-size: 14px;
            }

            .grand-total {
                font-weight: bold;
                font-size: 16px;
                margin-top: 8px;
            }

            /* Status Stamp */
            .stamp {
                position: absolute;
                top: 40px;
                right: 40px;
                border: 3px solid;
                padding: 4px 12px;
                font-size: 20px;
                font-weight: bold;
                text-transform: uppercase;
                transform: rotate(-15deg);
                opacity: 0.2;
                pointer-events: none;
            }

            .stamp--draft { color: #6b7280; border-color: #6b7280; }
            .stamp--approved { color: #059669; border-color: #059669; }
            .stamp--rejected { color: #dc2626; border-color: #dc2626; }
            .stamp--exported { color: #2563eb; border-color: #2563eb; }
            .stamp--pending { color: #d97706; border-color: #d97706; }

            /* Action Bar */
            .action-bar {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: var(--fb-spacing-sm) 0;
                margin-bottom: var(--fb-spacing-md);
                border-bottom: 1px solid #eee;
            }

            .status-badge {
                display: inline-block;
                padding: 2px 8px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: bold;
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            .status-badge--draft { background: #f3f4f6; color: #374151; }
            .status-badge--approved { background: #d1fae5; color: #065f46; }
            .status-badge--rejected { background: #fee2e2; color: #991b1b; }
            .status-badge--exported { background: #dbeafe; color: #1e40af; }
            .status-badge--pending { background: #fef3c7; color: #92400e; }

            .btn {
                padding: 6px 14px;
                border: none;
                border-radius: 4px;
                font-size: 12px;
                font-weight: 600;
                cursor: pointer;
                font-family: inherit;
                transition: background 0.15s, opacity 0.15s;
            }

            .btn:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .btn--edit {
                background: #2563eb;
                color: white;
            }

            .btn--edit:hover:not(:disabled) { background: #1d4ed8; }

            .btn--save {
                background: #059669;
                color: white;
            }

            .btn--save:hover:not(:disabled) { background: #047857; }

            .btn--cancel {
                background: #f3f4f6;
                color: #374151;
                margin-right: 8px;
            }

            .btn--cancel:hover:not(:disabled) { background: #e5e7eb; }

            .btn--add {
                background: none;
                color: #2563eb;
                border: 1px dashed #2563eb;
                padding: 4px 10px;
                font-size: 11px;
            }

            .btn--add:hover:not(:disabled) { background: #eff6ff; }

            .btn--remove {
                background: none;
                color: #dc2626;
                border: none;
                padding: 2px 6px;
                font-size: 16px;
                cursor: pointer;
                opacity: 0.4;
                transition: opacity 0.15s;
            }

            .btn--remove:hover { opacity: 1; }

            .edit-actions {
                display: flex;
                align-items: center;
            }

            .save-error {
                color: #dc2626;
                font-size: 11px;
                margin-top: 4px;
            }
        `
    ];

    @property({ attribute: false })
    data: InvoiceArtifactData | null = null;

    /** Whether the user has the finance:edit permission (set by parent) */
    @property({ type: Boolean, attribute: 'can-edit' })
    canEdit = false;

    /** Whether the user has the budget:approve permission (set by parent) */
    @property({ type: Boolean, attribute: 'can-approve' })
    canApprove = false;

    @state()
    private _isEditing = false;

    @state()
    private _draftItems: DraftLineItem[] = [];

    @state()
    private _isSaving = false;

    @state()
    private _saveError = '';

    /**
     * Formats cents (string or number) to USD currency.
     * Handles P1 Fix: Backend now sends int64 as string to preserve precision.
     */
    private _formatCurrency(cents: string | number): string {
        const numCents = typeof cents === 'string' ? parseInt(cents, 10) : cents;
        if (Number.isNaN(numCents)) return '$0.00';
        const dollars = numCents / 100;
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(dollars);
    }

    /** Returns the current invoice status. Defaults to Pending (non-editable) when unknown. */
    private get _status(): InvoiceStatus {
        return this.data?.status ?? InvoiceStatus.Pending;
    }

    /** Whether the invoice is in an editable state — requires both Draft status and edit permission */
    private get _isEditable(): boolean {
        return this._status === InvoiceStatus.Draft && this.canEdit;
    }

    /** Enter edit mode — clone items for draft mutation */
    private _enterEditMode(): void {
        if (!this.data || !this._isEditable) return;

        this._draftItems = this.data.line_items.map(item => ({
            description: item.description,
            quantity: item.quantity,
            unitPriceCents: typeof item.unit_price_cents === 'string'
                ? parseInt(item.unit_price_cents, 10)
                : item.unit_price_cents,
            totalCents: typeof item.total_cents === 'string'
                ? parseInt(item.total_cents, 10)
                : item.total_cents,
        }));
        this._isEditing = true;
        this._saveError = '';
    }

    /** Cancel edit — discard draft changes */
    private _cancelEdit(): void {
        this._isEditing = false;
        this._draftItems = [];
        this._saveError = '';
    }

    /** Update a field on a draft line item and recalculate */
    private _updateItem(index: number, field: keyof DraftLineItem, value: string): void {
        const items = [...this._draftItems];
        const existing = items[index];
        if (!existing) return;
        const item: DraftLineItem = {
            description: existing.description,
            quantity: existing.quantity,
            unitPriceCents: existing.unitPriceCents,
            totalCents: existing.totalCents,
        };

        if (field === 'description') {
            item.description = value;
        } else if (field === 'quantity') {
            item.quantity = parseFloat(value) || 0;
        } else if (field === 'unitPriceCents') {
            // Input is in dollars, convert to cents
            const dollars = parseFloat(value) || 0;
            item.unitPriceCents = Math.round(dollars * 100);
        }

        // Recalculate line total
        item.totalCents = Math.round(item.quantity * item.unitPriceCents);

        items[index] = item;
        this._draftItems = items;
    }

    /** Add a new empty line item */
    private _addItem(): void {
        this._draftItems = [
            ...this._draftItems,
            { description: '', quantity: 1, unitPriceCents: 0, totalCents: 0 },
        ];
    }

    /** Remove a line item by index */
    private _removeItem(index: number): void {
        if (this._draftItems.length <= 1) return; // Keep at least one item
        this._draftItems = this._draftItems.filter((_, i) => i !== index);
    }

    /** Calculate the draft grand total */
    private get _draftTotal(): number {
        return this._draftItems.reduce((sum, item) => sum + item.totalCents, 0);
    }

    /** Validate draft items before save */
    private _validateDraft(): string | null {
        for (const [i, item] of this._draftItems.entries()) {
            if (!item.description.trim()) {
                return `Line item ${String(i + 1)}: description is required`;
            }
            if (item.quantity <= 0) {
                return `Line item ${String(i + 1)}: quantity must be greater than 0`;
            }
            if (item.unitPriceCents < 0) {
                return `Line item ${String(i + 1)}: unit price cannot be negative`;
            }
        }
        return null;
    }

    /** Handle artifact approved event — update local status */
    private _onApproved(): void {
        if (this.data) {
            this.data = { ...this.data, status: InvoiceStatus.Approved };
            this._isEditing = false;
        }
    }

    /** Handle artifact rejected event — update local status */
    private _onRejected(): void {
        if (this.data) {
            this.data = { ...this.data, status: InvoiceStatus.Rejected };
            this._isEditing = false;
        }
    }

    /** Save draft changes to the backend */
    private async _save(): Promise<void> {
        if (!this.data?.id) {
            this._saveError = 'Invoice ID not available';
            return;
        }

        const validationError = this._validateDraft();
        if (validationError) {
            this._saveError = validationError;
            return;
        }

        this._isSaving = true;
        this._saveError = '';

        try {
            const response = await api.invoices.update(this.data.id, {
                items: this._draftItems.map(item => ({
                    description: item.description,
                    quantity: item.quantity,
                    unit_price_cents: item.unitPriceCents,
                })),
            });

            // Update parent data with response
            this.data = {
                ...this.data,
                line_items: response.line_items.map(li => ({
                    description: li.description,
                    quantity: li.quantity,
                    unit_price_cents: String(li.unit_price_cents),
                    total_cents: String(li.total_cents),
                })),
                total_amount_cents: String(response.amount_cents),
            };

            this._isEditing = false;
            this._draftItems = [];

            // Notify parent of successful save
            this.emit('invoice-updated', { invoiceId: this.data.id });
        } catch (err: unknown) {
            if (err instanceof Error) {
                this._saveError = err.message;
            } else {
                this._saveError = 'Failed to save changes';
            }
        } finally {
            this._isSaving = false;
        }
    }

    private _renderSkeleton(): TemplateResult {
        return html`
            <div class="invoice-container">
                <div class="header">
                    <div class="vendor-info" style="width: 200px;">
                        <div class="skeleton skeleton-text" style="width: 150px; height: 24px;"></div>
                        <div class="skeleton skeleton-text" style="width: 100px;"></div>
                    </div>
                </div>
                <div class="skeleton skeleton-box"></div>
                <div class="total-section">
                    <div class="skeleton skeleton-box" style="width: 200px; height: 80px;"></div>
                </div>
            </div>
        `;
    }

    private _renderStatusBadge(): TemplateResult {
        const status = this._status;
        const cssClass = `status-badge status-badge--${status.toLowerCase()}`;
        return html`<span class="${cssClass}">${status}</span>`;
    }

    private _renderStamp(): TemplateResult | typeof nothing {
        const status = this._status;
        if (status === InvoiceStatus.Draft) return nothing;
        const cssClass = `stamp stamp--${status.toLowerCase()}`;
        return html`<div class="${cssClass}">${status}</div>`;
    }

    private _renderActionBar(): TemplateResult {
        return html`
            <div class="action-bar">
                <div>${this._renderStatusBadge()}</div>
                <div>
                    ${this._isEditing
                        ? html`
                            <div class="edit-actions">
                                <button
                                    class="btn btn--cancel"
                                    @click=${this._cancelEdit}
                                    ?disabled=${this._isSaving}
                                >Cancel</button>
                                <button
                                    class="btn btn--save"
                                    @click=${this._save}
                                    ?disabled=${this._isSaving}
                                >${this._isSaving ? 'Saving...' : 'Save'}</button>
                            </div>
                        `
                        : this._isEditable
                            ? html`
                                <button
                                    class="btn btn--edit"
                                    @click=${this._enterEditMode}
                                >Edit Invoice</button>
                            `
                            : nothing
                    }
                </div>
            </div>
            ${this._saveError
                ? html`<div class="save-error">${this._saveError}</div>`
                : nothing
            }
        `;
    }

    private _renderViewRow(item: { description: string; quantity: number; unit_price_cents: string | number; total_cents: string | number }): TemplateResult {
        return html`
            <tr>
                <td class="col-desc">${item.description}</td>
                <td class="col-qty">${String(item.quantity)}</td>
                <td class="col-price">${this._formatCurrency(item.unit_price_cents)}</td>
                <td class="col-total">${this._formatCurrency(item.total_cents)}</td>
                ${this._isEditing ? html`<td class="col-actions"></td>` : nothing}
            </tr>
        `;
    }

    private _renderEditRow(item: DraftLineItem, index: number): TemplateResult {
        return html`
            <tr>
                <td class="col-desc">
                    <input
                        class="edit-input"
                        type="text"
                        .value=${item.description}
                        @input=${(e: Event) => this._updateItem(index, 'description', (e.target as HTMLInputElement).value)}
                        placeholder="Description"
                        maxlength="500"
                    />
                </td>
                <td class="col-qty">
                    <input
                        class="edit-input"
                        type="number"
                        .value=${String(item.quantity)}
                        @input=${(e: Event) => this._updateItem(index, 'quantity', (e.target as HTMLInputElement).value)}
                        min="0.01"
                        step="0.01"
                    />
                </td>
                <td class="col-price">
                    <input
                        class="edit-input"
                        type="number"
                        .value=${String(item.unitPriceCents / 100)}
                        @input=${(e: Event) => this._updateItem(index, 'unitPriceCents', (e.target as HTMLInputElement).value)}
                        min="0"
                        step="0.01"
                    />
                </td>
                <td class="col-total">${this._formatCurrency(item.totalCents)}</td>
                <td class="col-actions">
                    <button
                        class="btn--remove"
                        @click=${() => this._removeItem(index)}
                        title="Remove line item"
                        ?disabled=${this._draftItems.length <= 1}
                        aria-label="Remove line item"
                    >&times;</button>
                </td>
            </tr>
        `;
    }

    override render(): TemplateResult {
        if (!this.data) return this._renderSkeleton();

        const totalCents = this._isEditing ? this._draftTotal : this.data.total_amount_cents;

        return html`
            <div class="invoice-container">
                ${this._renderStamp()}
                ${this._renderActionBar()}

                <div class="header">
                    <div class="vendor-info">
                        <h2>${this.data.vendor}</h2>
                        <div>${this.data.address || ''}</div>
                    </div>
                    <div class="invoice-meta">
                        <div class="meta-row"><span class="label">Invoice #:</span> ${this.data.invoice_number}</div>
                        <div class="meta-row"><span class="label">Date:</span> ${new Date(this.data.date).toLocaleDateString()}</div>
                    </div>
                </div>

                <table class="line-items">
                    <thead>
                        <tr>
                            <th class="col-desc">Description</th>
                            <th class="col-qty">Qty</th>
                            <th class="col-price">Price</th>
                            <th class="col-total">Total</th>
                            ${this._isEditing ? html`<th class="col-actions"></th>` : nothing}
                        </tr>
                    </thead>
                    <tbody>
                        ${this._isEditing
                            ? this._draftItems.map((item, i) => this._renderEditRow(item, i))
                            : this.data.line_items.map(item => this._renderViewRow(item))
                        }
                    </tbody>
                </table>

                ${this._isEditing
                    ? html`
                        <button class="btn btn--add" @click=${this._addItem}>+ Add Item</button>
                    `
                    : nothing
                }

                <div class="total-section">
                    <div class="total-box">
                        <div class="total-row grand-total">
                            <span>TOTAL:</span>
                            <span>${this._formatCurrency(totalCents)}</span>
                        </div>
                    </div>
                </div>

                ${this.data.id ? html`
                    <fb-artifact-actions
                        .artifactId=${this.data.id}
                        .status=${this._status}
                        ?can-approve=${this.canApprove}
                        .approvedBy=${this.data.approved_by_id ?? ''}
                        .rejectedBy=${this.data.rejected_by_id ?? ''}
                        .rejectionReason=${this.data.rejection_reason ?? ''}
                        @artifact-approved=${this._onApproved}
                        @artifact-rejected=${this._onRejected}
                    ></fb-artifact-actions>
                ` : nothing}
            </div>
        `;
    }
}

import { html, css, TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { InvoiceArtifactData } from '../../types/artifacts';

@customElement('fb-artifact-invoice')
export class FBArtifactInvoice extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
                background: white; /* Paper look */
                color: black;
                border-radius: var(--fb-radius-md);
                border: 1px solid var(--fb-border);
                overflow: hidden;
            }

            .invoice-container {
                padding: var(--fb-spacing-lg);
                font-family: 'Courier New', Courier, monospace; /* Invoice font feel */
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
                padding: 8px 0;
                text-transform: uppercase;
            }

            .line-items td {
                padding: 8px 0;
                border-bottom: 1px dashed #ccc;
            }

            .col-desc { width: 60%; }
            .col-qty { width: 10%; text-align: center; }
            .col-price { width: 15%; text-align: right; }
            .col-total { width: 15%; text-align: right; }

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
            
            /* Status Stamp - TODO: Add @property status when approval workflow is implemented */
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
            /* Skeleton styles inherited from FBElement */
        `
    ];

    @property({ attribute: false })
    data: InvoiceArtifactData | null = null;


    private _formatCurrency(amount: number): string {
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
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

    override render(): TemplateResult {
        if (!this.data) return this._renderSkeleton();

        const total = this.data.total_amount;

        return html`
            <div class="invoice-container">
                <div class="stamp">PAID</div>
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
                        </tr>
                    </thead>
                    <tbody>
                        ${this.data.line_items.map(item => html`
                            <tr>
                                <td class="col-desc">${item.description}</td>
                                <td class="col-qty">${item.quantity}</td>
                                <td class="col-price">${this._formatCurrency(item.unit_price)}</td>
                                <td class="col-total">${this._formatCurrency(item.total)}</td>
                            </tr>
                        `)}
                    </tbody>
                </table>

                <div class="total-section">
                    <div class="total-box">
                        <div class="total-row grand-total">
                            <span>TOTAL:</span>
                            <span>${this._formatCurrency(total)}</span>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
}

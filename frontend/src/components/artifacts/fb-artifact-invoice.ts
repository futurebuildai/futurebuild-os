import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

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
            :host([status="approved"]) .stamp { color: var(--fb-success); border-color: var(--fb-success); }
            :host([status="denied"]) .stamp { color: var(--fb-error); border-color: var(--fb-error); }
            :host([status="pending"]) .stamp { opacity: 0; }
        `
    ];

    private _data = {
        number: 'INV-2024-001',
        date: '2024-01-15',
        vendor: 'ABC Lumber Co.',
        address: '123 Tree St, Woodville, OR',
        items: [
            { desc: '2x4x8 Doug Fir Studs', qty: 500, price: 4.50 },
            { desc: '3/4" Plywood Sheets (4x8)', qty: 100, price: 32.00 },
            { desc: 'Delivery Fee', qty: 1, price: 150.00 },
        ]
    };

    private _formatCurrency(amount: number): string {
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
    }

    private _getTotal(): number {
        return this._data.items.reduce((acc, item) => acc + (item.qty * item.price), 0);
    }

    override render(): TemplateResult {
        const subtotal = this._getTotal();
        const tax = subtotal * 0.08;
        const total = subtotal + tax;

        return html`
            <div class="invoice-container">
                <div class="stamp">PAID</div>
                <div class="header">
                    <div class="vendor-info">
                        <h2>${this._data.vendor}</h2>
                        <div>${this._data.address}</div>
                    </div>
                    <div class="invoice-meta">
                        <div class="meta-row"><span class="label">Invoice #:</span> ${this._data.number}</div>
                        <div class="meta-row"><span class="label">Date:</span> ${this._data.date}</div>
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
                        ${this._data.items.map(item => html`
                            <tr>
                                <td class="col-desc">${item.desc}</td>
                                <td class="col-qty">${item.qty}</td>
                                <td class="col-price">${this._formatCurrency(item.price)}</td>
                                <td class="col-total">${this._formatCurrency(item.qty * item.price)}</td>
                            </tr>
                        `)}
                    </tbody>
                </table>

                <div class="total-section">
                    <div class="total-box">
                        <div class="total-row">
                            <span>Subtotal:</span>
                            <span>${this._formatCurrency(subtotal)}</span>
                        </div>
                        <div class="total-row">
                            <span>Tax (8%):</span>
                            <span>${this._formatCurrency(tax)}</span>
                        </div>
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

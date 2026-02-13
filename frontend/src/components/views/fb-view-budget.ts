
import { html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { mockFinancialService, FinancialSummary } from '../../services/mock-financial-service';

@customElement('fb-view-budget')
export class FBViewBudget extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: block;
                max-width: 900px;
                margin: 0 auto;
                padding: 24px 16px 80px;
            }

            .header {
                margin-bottom: 24px;
            }

            h1 {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
                margin: 0 0 8px 0;
            }

            .subtitle {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .summary-cards {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                gap: 16px;
                margin-bottom: 32px;
            }

            .card {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                padding: 20px;
            }

            .card-label {
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-bottom: 8px;
            }

            .card-value {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .card-value.positive { color: #10b981; }
            .card-value.negative { color: #ef4444; }

            .breakdown-table {
                width: 100%;
                border-collapse: collapse;
                background: var(--fb-surface-1, #1a1a2e);
                border-radius: 12px;
                overflow: hidden;
                border: 1px solid var(--fb-border, #2a2a3e);
            }

            th, td {
                padding: 16px;
                text-align: left;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            th {
                background: var(--fb-surface-2, #252540);
                font-size: 12px;
                font-weight: 600;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
            }

            td {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .col-name { color: var(--fb-text-primary, #e0e0e0); font-weight: 500; }
            .col-money { font-family: monospace; }
            
            .status-badge {
                display: inline-block;
                padding: 4px 8px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 600;
                text-transform: uppercase;
            }

            .status-on_track { background: rgba(16, 185, 129, 0.1); color: #10b981; }
            .status-at_risk { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
            .status-over_budget { background: rgba(239, 68, 68, 0.1); color: #ef4444; }
        `
    ];

    @state() private _data: FinancialSummary | null = null;
    @state() private _loading = true;

    override connectedCallback() {
        super.connectedCallback();
        this._loadData();
    }

    override onViewActive(): void {
        this._loadData();
    }

    private async _loadData() {
        this._loading = true;
        try {
            this._data = await mockFinancialService.getSummary('p1');
        } catch (err) {
            console.error('[FBViewBudget] Failed to load data:', err);
            this._data = null;
        } finally {
            this._loading = false;
        }
    }

    private _formatCurrency(amount: number): string {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            maximumFractionDigits: 0
        }).format(amount);
    }

    override render() {
        if (this._loading) return html`<div>Loading budget data...</div>`;
        if (!this._data) return html`<div>Error loading data</div>`;

        return html`
            <div class="header">
                <h1>Budget Overview</h1>
                <div class="subtitle">Last updated: ${new Date(this._data.last_updated).toLocaleDateString()}</div>
            </div>

            <div class="summary-cards">
                <div class="card">
                    <div class="card-label">Total Budget</div>
                    <div class="card-value">${this._formatCurrency(this._data.budget_total)}</div>
                </div>
                <div class="card">
                    <div class="card-label">Total Spend</div>
                    <div class="card-value">${this._formatCurrency(this._data.spend_total)}</div>
                </div>
                <div class="card">
                    <div class="card-label">Variance</div>
                    <div class="card-value ${this._data.variance >= 0 ? 'positive' : 'negative'}">
                        ${this._data.variance >= 0 ? '+' : ''}${this._formatCurrency(this._data.variance)}
                    </div>
                </div>
            </div>

            <table class="breakdown-table">
                <thead>
                    <tr>
                        <th>Category</th>
                        <th>Budget</th>
                        <th>Spend</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    ${this._data.categories.map(cat => html`
                        <tr>
                            <td class="col-name">${cat.name}</td>
                            <td class="col-money">${this._formatCurrency(cat.budget)}</td>
                            <td class="col-money">${this._formatCurrency(cat.spend)}</td>
                            <td>
                                <span class="status-badge status-${cat.status}">
                                    ${cat.status.replace('_', ' ')}
                                </span>
                            </td>
                        </tr>
                    `)}
                </tbody>
            </table>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-budget': FBViewBudget;
    }
}

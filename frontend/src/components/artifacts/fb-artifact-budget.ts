import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

@customElement('fb-artifact-budget')
export class FBArtifactBudget extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
                background: var(--fb-bg-card);
                border-radius: var(--fb-radius-md);
                border: 1px solid var(--fb-border);
                overflow: hidden;
            }

            .container {
                padding: var(--fb-spacing-md);
            }

            .summary {
                display: grid;
                grid-template-columns: 1fr 1fr;
                gap: var(--fb-spacing-md);
                margin-bottom: var(--fb-spacing-lg);
                padding-bottom: var(--fb-spacing-md);
                border-bottom: 1px solid var(--fb-border-light);
            }

            .metric {
                display: flex;
                flex-direction: column;
            }

            .label {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                text-transform: uppercase;
                letter-spacing: 0.05em;
                margin-bottom: 4px;
            }

            .value {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .value.total { color: var(--fb-text-primary); }
            .value.spent { color: var(--fb-primary); }

            .category-list {
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-md);
            }

            .category-row {
                display: flex;
                flex-direction: column;
                gap: 4px;
            }

            .cat-header {
                display: flex;
                justify-content: space-between;
                font-size: var(--fb-text-sm);
                font-weight: 500;
            }

            .cat-name { color: var(--fb-text-primary); }
            .cat-val { color: var(--fb-text-secondary); }

            .progress-bg {
                height: 8px;
                background: var(--fb-bg-tertiary);
                border-radius: 4px;
                overflow: hidden;
            }

            .progress-fill {
                height: 100%;
                background: var(--fb-primary);
                border-radius: 4px;
            }

            .progress-fill.warning { background: var(--fb-warning); }
            .progress-fill.danger { background: var(--fb-error); }
        `
    ];

    private _data = {
        totalBudget: 450000,
        totalSpent: 125000,
        categories: [
            { name: 'Materials', budget: 200000, spent: 65000 },
            { name: 'Labor', budget: 150000, spent: 45000 },
            { name: 'Permits & Fees', budget: 25000, spent: 12000 },
            { name: 'Contingency', budget: 75000, spent: 3000 }
        ]
    };

    private _formatCurrency(amount: number): string {
        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(amount);
    }

    private _getPercent(spent: number, total: number): number {
        return Math.min(100, Math.round((spent / total) * 100));
    }

    private _getStatusClass(spent: number, total: number): string {
        const pct = spent / total;
        if (pct > 0.9) return 'danger';
        if (pct > 0.75) return 'warning';
        return '';
    }

    override render(): TemplateResult {
        return html`
            <div class="container">
                <div class="summary">
                    <div class="metric">
                        <span class="label">Total Budget</span>
                        <span class="value total">${this._formatCurrency(this._data.totalBudget)}</span>
                    </div>
                    <div class="metric">
                        <span class="label">Spent to Date</span>
                        <span class="value spent">${this._formatCurrency(this._data.totalSpent)}</span>
                    </div>
                </div>

                <div class="category-list">
                    ${this._data.categories.map(cat => html`
                        <div class="category-row">
                            <div class="cat-header">
                                <span class="cat-name">${cat.name}</span>
                                <span class="cat-val">${this._formatCurrency(cat.spent)} / ${this._formatCurrency(cat.budget)}</span>
                            </div>
                            <div class="progress-bg">
                                <div 
                                    class="progress-fill ${this._getStatusClass(cat.spent, cat.budget)}" 
                                    style="width: ${this._getPercent(cat.spent, cat.budget)}%"
                                ></div>
                            </div>
                        </div>
                    `)}
                </div>
            </div>
        `;
    }
}

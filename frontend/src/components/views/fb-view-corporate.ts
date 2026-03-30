/**
 * FBViewCorporate - Corporate Financials Dashboard
 * Phase 18 ERP: Budget rollups, AR aging, GL sync history.
 * See BACKEND_SCOPE.md Section 20.1
 */

import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api } from '../../services/api';
import type { CorporateBudget, ARAgingSnapshot, GLSyncLog } from '../../types/corporate';

@customElement('fb-view-corporate')
export class FBViewCorporate extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: block;
                max-width: 960px;
                margin: 0 auto;
                padding: 24px 16px 80px;
                background: var(--fb-bg-primary);
            }

            h1 {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #F0F0F5);
                margin: 0 0 4px 0;
            }

            .subtitle {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
                margin-bottom: 24px;
            }

            .section {
                background: rgba(22, 24, 33, 0.6);
                backdrop-filter: blur(24px);
                -webkit-backdrop-filter: blur(24px);
                border: 1px solid rgba(255, 255, 255, 0.05);
                border-radius: 12px;
                padding: 24px;
                margin-bottom: 20px;
            }

            .section-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
                margin: 0 0 16px 0;
            }

            /* Quarter selector */
            .quarter-nav {
                display: flex;
                align-items: center;
                gap: 12px;
                margin-bottom: 20px;
            }

            .quarter-nav button {
                background: var(--fb-surface-1, #161821);
                border: 1px solid rgba(255, 255, 255, 0.05);
                border-radius: 8px;
                color: var(--fb-text-primary, #F0F0F5);
                padding: 6px 14px;
                cursor: pointer;
                font-size: 13px;
                transition: background 0.15s ease;
            }

            .quarter-nav button:hover {
                background: var(--fb-surface-2, #1E2029);
            }

            .quarter-label {
                font-size: 14px;
                font-weight: 500;
                color: var(--fb-text-secondary, #8B8D98);
                min-width: 80px;
                text-align: center;
            }

            /* Budget big numbers */
            .budget-numbers {
                display: grid;
                grid-template-columns: repeat(3, 1fr);
                gap: 16px;
                margin-bottom: 16px;
            }

            .big-number-card {
                text-align: center;
                padding: 16px 8px;
                background: var(--fb-surface-1, #161821);
                border-radius: 8px;
                border: 1px solid rgba(255, 255, 255, 0.04);
            }

            .big-number-label {
                font-size: 12px;
                color: var(--fb-text-secondary, #8B8D98);
                text-transform: uppercase;
                margin-bottom: 6px;
            }

            .big-number-value {
                font-family: var(--fb-font-mono, monospace);
                font-size: 22px;
                font-weight: 700;
                color: var(--fb-text-primary, #F0F0F5);
            }

            .project-count {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
                margin-bottom: 16px;
            }

            .btn-rollup {
                background: var(--fb-primary, #00FFA3);
                color: #0A0B10;
                border: none;
                border-radius: 8px;
                padding: 10px 20px;
                font-size: 14px;
                font-weight: 600;
                cursor: pointer;
                transition: opacity 0.15s ease;
            }

            .btn-rollup:hover { opacity: 0.9; }
            .btn-rollup:disabled { opacity: 0.5; cursor: not-allowed; }

            /* AR Aging bar */
            .aging-bar {
                display: flex;
                height: 32px;
                border-radius: 6px;
                overflow: hidden;
                margin-bottom: 12px;
            }

            .aging-segment { transition: width 0.3s ease; }
            .aging-current { background: #00FFA3; }
            .aging-30 { background: #f59e0b; }
            .aging-60 { background: #f97316; }
            .aging-90 { background: #F43F5E; }

            .aging-legend {
                display: flex;
                flex-wrap: wrap;
                gap: 16px;
                align-items: center;
            }

            .legend-item {
                display: flex;
                align-items: center;
                gap: 6px;
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .legend-dot {
                width: 10px;
                height: 10px;
                border-radius: 3px;
                flex-shrink: 0;
            }

            .legend-amount {
                font-family: var(--fb-font-mono, monospace);
                color: var(--fb-text-primary, #F0F0F5);
                font-weight: 500;
            }

            .aging-total {
                margin-left: auto;
                font-size: 14px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
                font-family: var(--fb-font-mono, monospace);
            }

            /* GL Sync table */
            .sync-table {
                width: 100%;
                border-collapse: collapse;
            }

            .sync-table th,
            .sync-table td {
                padding: 12px 14px;
                text-align: left;
                border-bottom: 1px solid rgba(255, 255, 255, 0.04);
            }

            .sync-table th {
                font-size: 11px;
                font-weight: 600;
                color: var(--fb-text-tertiary, #5A5B66);
                text-transform: uppercase;
            }

            .sync-table td {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .status-badge {
                display: inline-block;
                padding: 3px 8px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 600;
                text-transform: uppercase;
            }

            .status-completed, .status-success { background: rgba(16, 185, 129, 0.1); color: #10b981; }
            .status-failed { background: rgba(244, 63, 94, 0.1); color: #F43F5E; }
            .status-pending, .status-in_progress { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }

            .empty-state {
                text-align: center;
                padding: 32px 16px;
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
            }

            .error-state {
                text-align: center;
                padding: 48px 16px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .error-state button {
                margin-top: 16px;
                padding: 8px 20px;
                border-radius: 8px;
                border: 1px solid rgba(255, 255, 255, 0.05);
                background: var(--fb-surface-1, #161821);
                color: var(--fb-text-primary, #F0F0F5);
                cursor: pointer;
                font-size: 14px;
            }

            .error-state button:hover { background: var(--fb-surface-2, #1E2029); }

            .td-error { color: var(--fb-text-tertiary, #5A5B66); font-size: 12px; }

            .skeleton {
                background: rgba(255, 255, 255, 0.05);
                border-radius: 4px;
                position: relative;
                overflow: hidden;
            }
            .skeleton::after {
                content: '';
                position: absolute;
                top: 0; left: -100%; width: 100%; height: 100%;
                background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.08), transparent);
                animation: fb-skeleton-shimmer 1.5s infinite;
            }
            @keyframes fb-skeleton-shimmer {
                100% { left: 100%; }
            }
        `,
    ];

    @state() private _budget: CorporateBudget | null = null;
    @state() private _aging: ARAgingSnapshot | null = null;
    @state() private _syncLogs: GLSyncLog[] = [];
    @state() private _viewState: 'loading' | 'ready' | 'error' = 'loading';
    @state() private _selectedYear: number = new Date().getFullYear();
    @state() private _selectedQuarter: number = Math.ceil((new Date().getMonth() + 1) / 3);
    @state() private _isRollingUp = false;

    override async onViewActive(): Promise<void> {
        await this._loadAll();
    }

    private async _loadAll(): Promise<void> {
        this._viewState = 'loading';
        try {
            const [budget, aging, logs] = await Promise.all([
                api.corporate.getBudget(this._selectedYear, this._selectedQuarter),
                api.corporate.getARAging(),
                api.corporate.getGLSyncLogs(),
            ]);
            this._budget = budget;
            this._aging = aging;
            this._syncLogs = logs;
            this._viewState = 'ready';
        } catch {
            this._viewState = 'error';
        }
    }

    private async _triggerRollup(): Promise<void> {
        this._isRollingUp = true;
        try {
            await api.corporate.rollupBudget(this._selectedYear, this._selectedQuarter);
            await this._loadAll();
        } catch {
            // Rollup failed; data reload in _loadAll handles state
        } finally {
            this._isRollingUp = false;
        }
    }

    private _prevQuarter(): void {
        if (this._selectedQuarter === 1) {
            this._selectedQuarter = 4;
            this._selectedYear -= 1;
        } else {
            this._selectedQuarter -= 1;
        }
        void this._loadAll();
    }

    private _nextQuarter(): void {
        if (this._selectedQuarter === 4) {
            this._selectedQuarter = 1;
            this._selectedYear += 1;
        } else {
            this._selectedQuarter += 1;
        }
        void this._loadAll();
    }

    private _formatCents(cents: number): string {
        return (cents / 100).toLocaleString('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 0,
        });
    }

    private _formatDate(iso: string): string {
        return new Date(iso).toLocaleString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    }

    override render(): TemplateResult {
        if (this._viewState === 'loading') {
            return html`
                <h1>Corporate Financials</h1>
                <div class="subtitle">Organization-wide budget rollups, receivables, and sync status</div>
                <div class="section">
                    <h2 class="section-title skeleton" style="width: 250px; height: 24px; margin-bottom: 20px; border-radius: 4px;"></h2>
                    <div class="budget-numbers">
                        <div class="big-number-card skeleton" style="height: 80px;"></div>
                        <div class="big-number-card skeleton" style="height: 80px;"></div>
                        <div class="big-number-card skeleton" style="height: 80px;"></div>
                    </div>
                </div>
                <div class="section">
                    <h2 class="section-title skeleton" style="width: 200px; height: 24px; margin-bottom: 20px;"></h2>
                    <div class="skeleton" style="width: 100%; height: 32px; border-radius: 6px; margin-bottom: 12px;"></div>
                </div>
            `;
        }
        if (this._viewState === 'error') {
            return html`
                <div class="error-state">
                    <div>Failed to load corporate data.</div>
                    <button @click=${() => this._loadAll()}>Retry</button>
                </div>
            `;
        }

        return html`
            <h1>Corporate Financials</h1>
            <div class="subtitle">Organization-wide budget rollups, receivables, and sync status</div>
            ${this._renderBudget()}
            ${this._renderAging()}
            ${this._renderSyncLogs()}
        `;
    }

    private _renderBudget(): TemplateResult {
        return html`
            <div class="section">
                <h2 class="section-title">Q${this._selectedQuarter} ${this._selectedYear} Budget Summary</h2>
                <div class="quarter-nav">
                    <button @click=${this._prevQuarter}>&larr; Prev</button>
                    <span class="quarter-label">Q${this._selectedQuarter} ${this._selectedYear}</span>
                    <button @click=${this._nextQuarter}>Next &rarr;</button>
                </div>
                ${this._budget
                    ? html`
                          <div class="budget-numbers">
                              <div class="big-number-card">
                                  <div class="big-number-label">Estimated</div>
                                  <div class="big-number-value">${this._formatCents(this._budget.total_estimated_cents)}</div>
                              </div>
                              <div class="big-number-card">
                                  <div class="big-number-label">Committed</div>
                                  <div class="big-number-value">${this._formatCents(this._budget.total_committed_cents)}</div>
                              </div>
                              <div class="big-number-card">
                                  <div class="big-number-label">Actual</div>
                                  <div class="big-number-value">${this._formatCents(this._budget.total_actual_cents)}</div>
                              </div>
                          </div>
                          <div class="project-count">${this._budget.project_count} project${this._budget.project_count !== 1 ? 's' : ''} in rollup</div>
                          <button
                              class="btn-rollup"
                              ?disabled=${this._isRollingUp}
                              @click=${this._triggerRollup}
                          >
                              ${this._isRollingUp ? 'Rolling up...' : 'Trigger Rollup'}
                          </button>
                      `
                    : html`<div class="empty-state">No budget data for this period</div>`}
            </div>
        `;
    }

    private _renderAging(): TemplateResult {
        return html`
            <div class="section">
                <h2 class="section-title">Accounts Receivable Aging</h2>
                ${this._aging
                    ? this._renderAgingContent(this._aging)
                    : html`<div class="empty-state">No aging data</div>`}
            </div>
        `;
    }

    private _renderAgingContent(aging: ARAgingSnapshot): TemplateResult {
        const total = aging.total_receivable_cents;
        const pct = (cents: number): string => (total > 0 ? `${(cents / total) * 100}%` : '0%');

        return html`
            <div class="aging-bar">
                ${aging.current_cents > 0 ? html`<div class="aging-segment aging-current" style="width:${pct(aging.current_cents)}"></div>` : nothing}
                ${aging.days_30_cents > 0 ? html`<div class="aging-segment aging-30" style="width:${pct(aging.days_30_cents)}"></div>` : nothing}
                ${aging.days_60_cents > 0 ? html`<div class="aging-segment aging-60" style="width:${pct(aging.days_60_cents)}"></div>` : nothing}
                ${aging.days_90_plus_cents > 0 ? html`<div class="aging-segment aging-90" style="width:${pct(aging.days_90_plus_cents)}"></div>` : nothing}
            </div>
            <div class="aging-legend">
                <div class="legend-item">
                    <span class="legend-dot" style="background:#00FFA3"></span>
                    Current
                    <span class="legend-amount">${this._formatCents(aging.current_cents)}</span>
                </div>
                <div class="legend-item">
                    <span class="legend-dot" style="background:#f59e0b"></span>
                    30 Day
                    <span class="legend-amount">${this._formatCents(aging.days_30_cents)}</span>
                </div>
                <div class="legend-item">
                    <span class="legend-dot" style="background:#f97316"></span>
                    60 Day
                    <span class="legend-amount">${this._formatCents(aging.days_60_cents)}</span>
                </div>
                <div class="legend-item">
                    <span class="legend-dot" style="background:#F43F5E"></span>
                    90+ Day
                    <span class="legend-amount">${this._formatCents(aging.days_90_plus_cents)}</span>
                </div>
                <span class="aging-total">Total: ${this._formatCents(total)}</span>
            </div>
        `;
    }

    private _renderSyncLogs(): TemplateResult {
        return html`
            <div class="section">
                <h2 class="section-title">GL Sync History</h2>
                ${this._syncLogs.length > 0
                    ? html`
                          <table class="sync-table">
                              <thead>
                                  <tr>
                                      <th>Time</th>
                                      <th>Type</th>
                                      <th>Status</th>
                                      <th>Records</th>
                                      <th>Error</th>
                                  </tr>
                              </thead>
                              <tbody>
                                  ${this._syncLogs.map(
                                      (log) => html`
                                          <tr>
                                              <td>${log.synced_at ? this._formatDate(log.synced_at) : '\u2014'}</td>
                                              <td>${log.sync_type}</td>
                                              <td>
                                                  <span class="status-badge status-${log.status}">
                                                      ${log.status.replace('_', ' ')}
                                                  </span>
                                              </td>
                                              <td>${log.records_synced ?? '\u2014'}</td>
                                              <td class="td-error">${log.error_message ?? '\u2014'}</td>
                                          </tr>
                                      `,
                                  )}
                              </tbody>
                          </table>
                      `
                    : html`<div class="empty-state">No sync logs yet</div>`}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-corporate': FBViewCorporate;
    }
}

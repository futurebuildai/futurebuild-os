/**
 * FBViewCompletionReport - Project Completion Report View
 *
 * Displays the completion report for a project that has been marked complete.
 * Shows schedule summary, budget summary, weather impact, and procurement data.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import type { CompletionReport } from '../../types/models';

/**
 * Formats cents to a dollar string (e.g., 150000 -> "$1,500.00").
 */
function formatCents(cents: number): string {
    const dollars = cents / 100;
    return dollars.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
}

/**
 * Completion report view component.
 * @element fb-view-completion-report
 */
@customElement('fb-view-completion-report')
export class FBViewCompletionReport extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                padding: 24px;
                max-width: 800px;
            }

            .report-header {
                margin-bottom: 24px;
            }

            .report-title {
                color: var(--fb-text-primary, #fff);
                font-size: 20px;
                font-weight: 600;
                margin: 0 0 4px 0;
            }

            .report-meta {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 13px;
                margin: 0;
            }

            .cards {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
                gap: 16px;
            }

            .card {
                background: var(--fb-bg-secondary, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: var(--fb-radius-md, 8px);
                padding: 20px;
            }

            .card-title {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 12px;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.5px;
                margin: 0 0 16px 0;
            }

            .stat-row {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 6px 0;
            }

            .stat-row + .stat-row {
                border-top: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .stat-label {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 13px;
            }

            .stat-value {
                color: var(--fb-text-primary, #fff);
                font-size: 14px;
                font-weight: 500;
            }

            .stat-value--positive {
                color: var(--fb-success, #00FFA3);
            }

            .stat-value--negative {
                color: var(--fb-danger, #F43F5E);
            }

            .notes-section {
                margin-top: 16px;
            }

            .notes-label {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 12px;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.5px;
                margin: 0 0 8px 0;
            }

            .notes-text {
                color: var(--fb-text-primary, #fff);
                font-size: 14px;
                line-height: 1.5;
                white-space: pre-wrap;
            }

            .loading {
                display: flex;
                align-items: center;
                justify-content: center;
                min-height: 200px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .error {
                color: var(--fb-danger, #F43F5E);
                padding: 16px;
                text-align: center;
            }
        `,
    ];

    @property({ type: String }) projectId = '';
    @state() private _report: CompletionReport | null = null;
    @state() private _loading = false;
    @state() private _error: string | null = null;
    private _loadedProjectId: string | null = null;

    override connectedCallback(): void {
        super.connectedCallback();
        if (this.projectId) {
            void this._loadReport();
        }
    }

    override updated(changedProperties: Map<string, unknown>): void {
        if (changedProperties.has('projectId') && this.projectId && this.projectId !== this._loadedProjectId) {
            void this._loadReport();
        }
    }

    private async _loadReport(): Promise<void> {
        this._loadedProjectId = this.projectId;
        this._loading = true;
        this._error = null;
        try {
            this._report = await api.completion.getReport(this.projectId);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to load report';
        } finally {
            this._loading = false;
        }
    }

    private _renderScheduleCard(): TemplateResult {
        const s = this._report!.schedule_summary;
        return html`
            <div class="card">
                <h3 class="card-title">Schedule</h3>
                <div class="stat-row">
                    <span class="stat-label">Total Tasks</span>
                    <span class="stat-value">${s.total_tasks}</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Completed Tasks</span>
                    <span class="stat-value">${s.completed_tasks}</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">On-Time %</span>
                    <span class="stat-value">${s.on_time_percent.toFixed(1)}%</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Planned Duration</span>
                    <span class="stat-value">${s.total_duration_days} days</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Actual Duration</span>
                    <span class="stat-value">${s.actual_duration_days} days</span>
                </div>
            </div>
        `;
    }

    private _renderBudgetCard(): TemplateResult {
        const b = this._report!.budget_summary;
        const varianceClass = b.variance_cents > 0 ? 'stat-value--negative' : 'stat-value--positive';
        return html`
            <div class="card">
                <h3 class="card-title">Budget</h3>
                <div class="stat-row">
                    <span class="stat-label">Estimated</span>
                    <span class="stat-value">${formatCents(b.estimated_cents)}</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Committed</span>
                    <span class="stat-value">${formatCents(b.committed_cents)}</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Actual</span>
                    <span class="stat-value">${formatCents(b.actual_cents)}</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Variance</span>
                    <span class="stat-value ${varianceClass}">${formatCents(b.variance_cents)}</span>
                </div>
            </div>
        `;
    }

    private _renderWeatherCard(): TemplateResult | typeof nothing {
        const w = this._report!.weather_impact_summary;
        if (!w) return nothing;
        return html`
            <div class="card">
                <h3 class="card-title">Weather Impact</h3>
                <div class="stat-row">
                    <span class="stat-label">Total Delay</span>
                    <span class="stat-value">${w.total_delay_days} days</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Phases Affected</span>
                    <span class="stat-value">${w.phases_affected}</span>
                </div>
            </div>
        `;
    }

    private _renderProcurementCard(): TemplateResult | typeof nothing {
        const p = this._report!.procurement_summary;
        if (!p) return nothing;
        return html`
            <div class="card">
                <h3 class="card-title">Procurement</h3>
                <div class="stat-row">
                    <span class="stat-label">Total Items</span>
                    <span class="stat-value">${p.total_items}</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Total Spend</span>
                    <span class="stat-value">${formatCents(p.total_spend_cents)}</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Vendors</span>
                    <span class="stat-value">${p.vendor_count}</span>
                </div>
            </div>
        `;
    }

    override render(): TemplateResult {
        if (this._loading) {
            return html`<div class="loading">Loading completion report...</div>`;
        }
        if (this._error) {
            return html`<div class="error">${this._error}</div>`;
        }
        if (!this._report) {
            return html`<div class="loading">No completion report available.</div>`;
        }

        const createdAt = new Date(this._report.created_at).toLocaleDateString();

        return html`
            <div class="report-header">
                <h2 class="report-title">Completion Report</h2>
                <p class="report-meta">Generated ${createdAt}</p>
            </div>

            <div class="cards">
                ${this._renderScheduleCard()}
                ${this._renderBudgetCard()}
                ${this._renderWeatherCard()}
                ${this._renderProcurementCard()}
            </div>

            ${this._report.notes ? html`
                <div class="notes-section">
                    <h3 class="notes-label">Notes</h3>
                    <p class="notes-text">${this._report.notes}</p>
                </div>
            ` : nothing}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-completion-report': FBViewCompletionReport;
    }
}

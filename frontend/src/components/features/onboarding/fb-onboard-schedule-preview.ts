/**
 * FBOnboardSchedulePreview - Instant Schedule Preview for Onboarding
 *
 * Renders a compact Gantt-style preview from the schedule preview signal.
 * Shows:
 * - Phase timeline bars (colored by status: completed/in-progress/pending)
 * - Critical path highlighting
 * - Procurement deadline warnings
 * - Projected end date and trade gap summary
 *
 * Appears after P0 fields (sqft, foundation, start_date) are captured.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../../base/FBElement';
import {
    schedulePreview,
    isInProgressProject,
} from '../../../store/onboarding-store';
import type { PhasePreview, ProcurementDate, TradeGap } from '../../../types/schedule';

@customElement('fb-onboard-schedule-preview')
export class FBOnboardSchedulePreview extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                padding: var(--fb-spacing-md);
            }

            .preview-container {
                background: var(--fb-surface-1);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-md);
            }

            .preview-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: var(--fb-spacing-md);
            }

            .preview-title {
                font-size: var(--fb-text-sm);
                font-weight: 600;
                color: var(--fb-text-primary);
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            .projected-end {
                font-size: var(--fb-text-sm);
                color: var(--fb-accent);
                font-weight: 600;
            }

            .phase-list {
                display: flex;
                flex-direction: column;
                gap: 4px;
            }

            .phase-row {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                height: 28px;
            }

            .phase-label {
                width: 140px;
                font-size: 11px;
                color: var(--fb-text-secondary);
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
                flex-shrink: 0;
            }

            .phase-bar-container {
                flex: 1;
                height: 16px;
                position: relative;
                background: var(--fb-bg-tertiary);
                border-radius: 3px;
                overflow: hidden;
            }

            .phase-bar {
                position: absolute;
                top: 0;
                height: 100%;
                border-radius: 3px;
                transition: width 0.3s ease;
                min-width: 2px;
            }

            .phase-bar.pending {
                background: var(--fb-accent);
                opacity: 0.6;
            }

            .phase-bar.critical {
                background: #F43F5E;
                opacity: 0.9;
            }

            .phase-bar.completed {
                background: var(--fb-accent);
                opacity: 1.0;
            }

            .phase-bar.in_progress {
                background: #FBBF24;
                opacity: 0.9;
            }

            .phase-duration {
                width: 50px;
                font-size: 10px;
                color: var(--fb-text-muted);
                text-align: right;
                flex-shrink: 0;
            }

            .summary-section {
                margin-top: var(--fb-spacing-md);
                padding-top: var(--fb-spacing-md);
                border-top: 1px solid var(--fb-border);
                display: flex;
                gap: var(--fb-spacing-lg);
                flex-wrap: wrap;
            }

            .summary-stat {
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .stat-label {
                font-size: 10px;
                color: var(--fb-text-muted);
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            .stat-value {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-primary);
                font-weight: 600;
            }

            .warnings-section {
                margin-top: var(--fb-spacing-sm);
                display: flex;
                flex-direction: column;
                gap: 4px;
            }

            .warning-item {
                font-size: 11px;
                color: #FBBF24;
                display: flex;
                align-items: center;
                gap: 4px;
            }

            .warning-icon {
                flex-shrink: 0;
            }

            .skeleton {
                background: var(--fb-surface-1);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-lg);
                text-align: center;
                color: var(--fb-text-muted);
                font-size: var(--fb-text-sm);
            }

            .skeleton-bar {
                height: 16px;
                background: var(--fb-bg-tertiary);
                border-radius: 3px;
                margin: 4px 0;
                animation: pulse 1.5s ease-in-out infinite;
            }

            @keyframes pulse {
                0%, 100% { opacity: 0.4; }
                50% { opacity: 0.8; }
            }
        `
    ];

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();
        // Subscribe to signal changes to trigger re-render
        this._disposeEffects.push(
            effect(() => {
                void schedulePreview.value;
                this.requestUpdate();
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach(d => d());
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    override render(): TemplateResult {
        const preview = schedulePreview.value;

        if (!preview) {
            return html`
                <div class="skeleton">
                    <div class="skeleton-bar" style="width: 80%"></div>
                    <div class="skeleton-bar" style="width: 60%"></div>
                    <div class="skeleton-bar" style="width: 90%"></div>
                    <div class="skeleton-bar" style="width: 40%"></div>
                    <p>Schedule preview will appear after project details are captured</p>
                </div>
            `;
        }

        const phases = preview.phase_timeline ?? [];
        const procDates = preview.procurement_dates ?? [];
        const tradeGaps = preview.trade_gaps ?? [];
        const inProgress = isInProgressProject.value;

        return html`
            <div class="preview-container">
                <div class="preview-header">
                    <span class="preview-title">Schedule Preview</span>
                    <span class="projected-end">
                        ${inProgress ? 'Projected completion' : 'Est. completion'}: ${preview.projected_end}
                    </span>
                </div>

                <div class="phase-list">
                    ${phases.map(phase => this._renderPhaseBar(phase, phases))}
                </div>

                <div class="summary-section">
                    <div class="summary-stat">
                        <span class="stat-label">Total Duration</span>
                        <span class="stat-value">${preview.total_working_days} days</span>
                    </div>
                    ${inProgress ? html`
                        <div class="summary-stat">
                            <span class="stat-label">Remaining</span>
                            <span class="stat-value">${preview.remaining_days} days</span>
                        </div>
                        <div class="summary-stat">
                            <span class="stat-label">Complete</span>
                            <span class="stat-value">${(preview.completion_percent ?? 0).toFixed(0)}%</span>
                        </div>
                    ` : nothing}
                    <div class="summary-stat">
                        <span class="stat-label">Critical Tasks</span>
                        <span class="stat-value">${(preview.critical_path ?? []).length}</span>
                    </div>
                </div>

                ${procDates.length > 0 || tradeGaps.length > 0 ? html`
                    <div class="warnings-section">
                        ${this._renderProcurementWarnings(procDates)}
                        ${this._renderTradeGapWarnings(tradeGaps)}
                    </div>
                ` : nothing}
            </div>
        `;
    }

    private _renderPhaseBar(phase: PhasePreview, allPhases: PhasePreview[]): TemplateResult {
        // Calculate relative position/width within the timeline
        const totalDuration = allPhases.reduce((sum, p) => sum + p.duration_days, 0);
        const widthPercent = totalDuration > 0 ? (phase.duration_days / totalDuration) * 100 : 10;

        let barClass = 'pending';
        if (phase.status === 'completed') barClass = 'completed';
        else if (phase.status === 'in_progress') barClass = 'in_progress';
        else if (phase.is_critical) barClass = 'critical';

        return html`
            <div class="phase-row">
                <span class="phase-label" title="${phase.phase_name}">${phase.phase_name}</span>
                <div class="phase-bar-container">
                    <div
                        class="phase-bar ${barClass}"
                        style="width: ${Math.max(widthPercent, 3)}%; left: 0"
                    ></div>
                </div>
                <span class="phase-duration">${phase.duration_days}d</span>
            </div>
        `;
    }

    private _renderProcurementWarnings(dates: ProcurementDate[]): TemplateResult | typeof nothing {
        const urgent = dates.filter(d => d.status === 'overdue' || d.status === 'urgent');
        if (urgent.length === 0) return nothing;

        return html`
            ${urgent.map(d => html`
                <div class="warning-item">
                    <span class="warning-icon">&#9888;</span>
                    <span>${d.item_name}: order by ${d.order_by_date} (${d.status})</span>
                </div>
            `)}
        `;
    }

    private _renderTradeGapWarnings(gaps: TradeGap[]): TemplateResult | typeof nothing {
        if (gaps.length === 0) return nothing;

        const firstFew = gaps.slice(0, 3);
        return html`
            ${firstFew.map(g => html`
                <div class="warning-item">
                    <span class="warning-icon">&#128736;</span>
                    <span>${g.required_trade} needed by ${g.start_date}</span>
                </div>
            `)}
            ${gaps.length > 3 ? html`
                <div class="warning-item">
                    <span>+${gaps.length - 3} more trades needed</span>
                </div>
            ` : nothing}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboard-schedule-preview': FBOnboardSchedulePreview;
    }
}

/**
 * fb-schedule-diff — Before/after schedule comparison overlay.
 * See FRONTEND_V2_SPEC.md §6.3, Phase 6 Step 39
 *
 * Shows the delta between old and new schedule after CPM recalculation:
 * - End date change (+/- days)
 * - Tasks that moved
 * - Critical path changes
 * - New bottlenecks
 */
import { html, css, nothing, type TemplateResult } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export interface ScheduleDiff {
    old_end_date: string;
    new_end_date: string;
    end_date_delta_days: number;
    tasks_moved: TaskMovement[];
    critical_path_changed: boolean;
    old_critical_path?: string[];
    new_critical_path?: string[];
    new_bottlenecks?: string[];
}

export interface TaskMovement {
    wbs_code: string;
    name: string;
    old_start: string;
    new_start: string;
    old_finish: string;
    new_finish: string;
    delta_days: number;
    became_critical?: boolean;
    no_longer_critical?: boolean;
}

@customElement('fb-schedule-diff')
export class FBScheduleDiff extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .diff-overlay {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                overflow: hidden;
            }

            .diff-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 16px 20px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .diff-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .close-btn {
                width: 28px;
                height: 28px;
                border-radius: 6px;
                border: none;
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                cursor: pointer;
                display: flex;
                align-items: center;
                justify-content: center;
                transition: all 0.15s ease;
            }

            .close-btn:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .diff-summary {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
                gap: 16px;
                padding: 20px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .summary-card {
                padding: 16px;
                background: var(--fb-surface-2, #252540);
                border-radius: 8px;
                text-align: center;
            }

            .summary-label {
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
                letter-spacing: 0.05em;
                margin-bottom: 8px;
            }

            .summary-value {
                font-size: 24px;
                font-weight: 700;
            }

            .summary-value.positive {
                color: #22c55e;
            }

            .summary-value.negative {
                color: #ef4444;
            }

            .summary-value.neutral {
                color: var(--fb-text-primary, #e0e0e0);
            }

            .summary-detail {
                font-size: 12px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-top: 4px;
            }

            .end-date-comparison {
                display: flex;
                align-items: center;
                justify-content: center;
                gap: 12px;
                padding: 20px;
                background: var(--fb-bg-primary, #0f0f1a);
            }

            .date-box {
                padding: 12px 20px;
                background: var(--fb-surface-2, #252540);
                border-radius: 8px;
                text-align: center;
            }

            .date-box.old {
                opacity: 0.6;
            }

            .date-box-label {
                font-size: 11px;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
                margin-bottom: 4px;
            }

            .date-box-value {
                font-size: 15px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .date-box.old .date-box-value {
                text-decoration: line-through;
            }

            .arrow {
                font-size: 18px;
                color: var(--fb-text-tertiary, #707080);
            }

            .delta-badge {
                padding: 6px 12px;
                border-radius: 6px;
                font-size: 13px;
                font-weight: 600;
            }

            .delta-badge.positive {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .delta-badge.negative {
                background: rgba(239, 68, 68, 0.15);
                color: #ef4444;
            }

            .task-changes {
                padding: 20px;
            }

            .section-title {
                font-size: 13px;
                font-weight: 600;
                color: var(--fb-text-secondary, #a0a0b0);
                text-transform: uppercase;
                letter-spacing: 0.05em;
                margin-bottom: 12px;
            }

            .task-list {
                display: flex;
                flex-direction: column;
                gap: 8px;
            }

            .task-item {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 12px 16px;
                background: var(--fb-surface-2, #252540);
                border-radius: 8px;
            }

            .task-info {
                display: flex;
                flex-direction: column;
                gap: 4px;
            }

            .task-name {
                font-size: 14px;
                font-weight: 500;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .task-wbs {
                font-size: 11px;
                font-family: monospace;
                color: var(--fb-text-tertiary, #707080);
            }

            .task-delta {
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .task-delta-value {
                font-size: 14px;
                font-weight: 600;
            }

            .task-delta-value.slip {
                color: #ef4444;
            }

            .task-delta-value.advance {
                color: #22c55e;
            }

            .critical-badge {
                padding: 2px 8px;
                border-radius: 4px;
                font-size: 10px;
                font-weight: 600;
                text-transform: uppercase;
            }

            .critical-badge.became {
                background: rgba(239, 68, 68, 0.15);
                color: #ef4444;
            }

            .critical-badge.removed {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .empty-changes {
                padding: 40px 20px;
                text-align: center;
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 14px;
            }

            .diff-actions {
                display: flex;
                justify-content: flex-end;
                gap: 12px;
                padding: 16px 20px;
                border-top: 1px solid var(--fb-border, #2a2a3e);
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

            .btn-secondary {
                background: transparent;
                border: 1px solid var(--fb-border, #2a2a3e);
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .btn-secondary:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .btn-primary {
                background: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .btn-primary:hover {
                opacity: 0.9;
            }

            @media (max-width: 600px) {
                .diff-summary {
                    grid-template-columns: 1fr;
                }

                .end-date-comparison {
                    flex-direction: column;
                }

                .arrow {
                    transform: rotate(90deg);
                }
            }
        `,
    ];

    /** The schedule diff data to display */
    @property({ attribute: false })
    diff: ScheduleDiff | null = null;

    /** Whether to show the expanded task list */
    @state() private _showAllTasks = false;

    private _formatDate(iso: string): string {
        const d = new Date(iso + 'T00:00:00');
        return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    }

    private _handleClose() {
        this.emit('fb-schedule-diff-close');
    }

    private _handleViewSchedule() {
        this.emit('fb-navigate', { view: 'schedule' });
        this._handleClose();
    }

    private _renderSummary(): TemplateResult {
        if (!this.diff) return html``;

        const delta = this.diff.end_date_delta_days;
        const deltaClass = delta > 0 ? 'negative' : delta < 0 ? 'positive' : 'neutral';
        const deltaPrefix = delta > 0 ? '+' : '';
        const tasksMoved = this.diff.tasks_moved.length;
        const criticalChanged = this.diff.critical_path_changed;

        return html`
            <div class="diff-summary">
                <div class="summary-card">
                    <div class="summary-label">End Date Change</div>
                    <div class="summary-value ${deltaClass}">
                        ${delta === 0 ? '—' : `${deltaPrefix}${delta}d`}
                    </div>
                    <div class="summary-detail">
                        ${delta === 0
                            ? 'No change'
                            : delta > 0
                            ? 'Completion delayed'
                            : 'Completion advanced'}
                    </div>
                </div>

                <div class="summary-card">
                    <div class="summary-label">Tasks Moved</div>
                    <div class="summary-value neutral">${tasksMoved}</div>
                    <div class="summary-detail">
                        ${tasksMoved === 0 ? 'No changes' : tasksMoved === 1 ? '1 task affected' : `${tasksMoved} tasks affected`}
                    </div>
                </div>

                <div class="summary-card">
                    <div class="summary-label">Critical Path</div>
                    <div class="summary-value ${criticalChanged ? 'negative' : 'positive'}">
                        ${criticalChanged ? 'Changed' : 'Unchanged'}
                    </div>
                    <div class="summary-detail">
                        ${criticalChanged ? 'New bottleneck detected' : 'Same sequence'}
                    </div>
                </div>
            </div>
        `;
    }

    private _renderEndDateComparison(): TemplateResult {
        if (!this.diff) return html``;

        const delta = this.diff.end_date_delta_days;
        const deltaClass = delta > 0 ? 'negative' : 'positive';

        return html`
            <div class="end-date-comparison">
                <div class="date-box old">
                    <div class="date-box-label">Previous</div>
                    <div class="date-box-value">${this._formatDate(this.diff.old_end_date)}</div>
                </div>
                <span class="arrow">→</span>
                <div class="date-box">
                    <div class="date-box-label">New</div>
                    <div class="date-box-value">${this._formatDate(this.diff.new_end_date)}</div>
                </div>
                ${delta !== 0
                    ? html`<span class="delta-badge ${deltaClass}">${delta > 0 ? '+' : ''}${delta} days</span>`
                    : nothing}
            </div>
        `;
    }

    private _renderTaskChanges(): TemplateResult {
        if (!this.diff) return html``;

        const tasks = this._showAllTasks
            ? this.diff.tasks_moved
            : this.diff.tasks_moved.slice(0, 5);

        if (this.diff.tasks_moved.length === 0) {
            return html`
                <div class="empty-changes">
                    No individual task changes to display
                </div>
            `;
        }

        return html`
            <div class="task-changes">
                <div class="section-title">Affected Tasks</div>
                <div class="task-list">
                    ${tasks.map((task) => this._renderTaskItem(task))}
                </div>
                ${this.diff.tasks_moved.length > 5 && !this._showAllTasks
                    ? html`
                          <button
                              class="btn btn-secondary"
                              style="width: 100%; margin-top: 12px;"
                              @click=${() => { this._showAllTasks = true; }}
                          >
                              Show all ${this.diff.tasks_moved.length} tasks
                          </button>
                      `
                    : nothing}
            </div>
        `;
    }

    private _renderTaskItem(task: TaskMovement): TemplateResult {
        const deltaClass = task.delta_days > 0 ? 'slip' : 'advance';
        const deltaPrefix = task.delta_days > 0 ? '+' : '';

        return html`
            <div class="task-item">
                <div class="task-info">
                    <span class="task-name">${task.name}</span>
                    <span class="task-wbs">${task.wbs_code}</span>
                </div>
                <div class="task-delta">
                    ${task.became_critical
                        ? html`<span class="critical-badge became">Now Critical</span>`
                        : nothing}
                    ${task.no_longer_critical
                        ? html`<span class="critical-badge removed">No Longer Critical</span>`
                        : nothing}
                    <span class="task-delta-value ${deltaClass}">
                        ${deltaPrefix}${task.delta_days}d
                    </span>
                </div>
            </div>
        `;
    }

    override render(): TemplateResult {
        if (!this.diff) {
            return html`
                <div class="diff-overlay">
                    <div class="empty-changes">No schedule changes to display</div>
                </div>
            `;
        }

        return html`
            <div class="diff-overlay" role="dialog" aria-label="Schedule Changes">
                <div class="diff-header">
                    <span class="diff-title">Schedule Recalculated</span>
                    <button class="close-btn" @click=${this._handleClose} aria-label="Close">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <line x1="18" y1="6" x2="6" y2="18"/>
                            <line x1="6" y1="6" x2="18" y2="18"/>
                        </svg>
                    </button>
                </div>

                ${this._renderSummary()}
                ${this._renderEndDateComparison()}
                ${this._renderTaskChanges()}

                <div class="diff-actions">
                    <button class="btn btn-secondary" @click=${this._handleClose}>
                        Dismiss
                    </button>
                    <button class="btn btn-primary" @click=${this._handleViewSchedule}>
                        View Schedule
                    </button>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-schedule-diff': FBScheduleDiff;
    }
}

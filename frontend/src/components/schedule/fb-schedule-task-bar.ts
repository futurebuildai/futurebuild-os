/**
 * fb-schedule-task-bar — Individual task bar on Gantt timeline.
 * See FRONTEND_V2_SPEC.md §2.3.E, Phase 6 Step 36
 *
 * Positioned by date, styled by status. Supports:
 * - Normal, critical, completed, delayed states
 * - Progress fill overlay
 * - Float visualization (dashed ghost bar)
 * - Hover tooltip with task details
 * - Click to emit selection event
 */
import { html, css, nothing, type TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { TaskStatus } from '../../types/enums';

export interface TaskBarData {
    wbs_code: string;
    name: string;
    early_start: string;
    early_finish: string;
    late_start?: string;
    late_finish?: string;
    duration_days: number;
    total_float?: number;
    status: TaskStatus;
    is_critical?: boolean;
    progress?: number; // 0-100
}

@customElement('fb-schedule-task-bar')
export class FBScheduleTaskBar extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                position: absolute;
                height: 24px;
            }

            .task-bar-wrapper {
                position: relative;
                height: 100%;
                cursor: pointer;
            }

            /* Main bar */
            .task-bar {
                position: absolute;
                height: 100%;
                border-radius: 4px;
                min-width: 4px;
                transition: transform 0.1s ease, box-shadow 0.1s ease;
            }

            .task-bar-wrapper:hover .task-bar {
                transform: scaleY(1.1);
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
            }

            /* Status colors */
            .task-bar.normal {
                background: var(--fb-accent, #00FFA3);
            }

            .task-bar.critical {
                background: #F43F5E;
            }

            .task-bar.completed {
                background: #00FFA3;
            }

            .task-bar.delayed {
                background: #f97316;
            }

            .task-bar.pending {
                background: var(--fb-surface-2, #1E2029);
                border: 1px dashed var(--fb-border, rgba(255,255,255,0.05));
            }

            /* Progress overlay */
            .progress-fill {
                position: absolute;
                top: 0;
                left: 0;
                height: 100%;
                background: rgba(255, 255, 255, 0.2);
                border-radius: 4px 0 0 4px;
                pointer-events: none;
            }

            /* Float bar (late dates) */
            .float-bar {
                position: absolute;
                height: 100%;
                background: transparent;
                border: 1px dashed var(--fb-text-tertiary, #5A5B66);
                border-radius: 4px;
                opacity: 0.5;
            }

            /* Critical path indicator */
            .critical-indicator {
                position: absolute;
                top: -2px;
                right: -2px;
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background: #F43F5E;
                border: 1px solid var(--fb-bg-primary, #0f0f1a);
            }

            /* Tooltip */
            .tooltip {
                position: absolute;
                bottom: calc(100% + 8px);
                left: 50%;
                transform: translateX(-50%);
                background: var(--fb-surface-1, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 6px;
                padding: 8px 12px;
                font-size: 12px;
                white-space: nowrap;
                opacity: 0;
                visibility: hidden;
                transition: opacity 0.15s ease, visibility 0.15s ease;
                z-index: 100;
                pointer-events: none;
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
            }

            .task-bar-wrapper:hover .tooltip {
                opacity: 1;
                visibility: visible;
            }

            .tooltip::after {
                content: '';
                position: absolute;
                top: 100%;
                left: 50%;
                transform: translateX(-50%);
                border: 6px solid transparent;
                border-top-color: var(--fb-border, rgba(255,255,255,0.05));
            }

            .tooltip-title {
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
                margin-bottom: 4px;
            }

            .tooltip-wbs {
                font-family: var(--fb-font-mono, monospace);
                font-size: 10px;
                color: var(--fb-text-tertiary, #5A5B66);
                margin-bottom: 6px;
            }

            .tooltip-dates {
                font-family: var(--fb-font-mono, monospace);
                color: var(--fb-text-secondary, #8B8D98);
            }

            .tooltip-status {
                display: flex;
                align-items: center;
                gap: 6px;
                margin-top: 6px;
                font-size: 11px;
            }

            .status-dot {
                width: 6px;
                height: 6px;
                border-radius: 50%;
            }

            .status-dot.normal { background: var(--fb-accent, #00FFA3); }
            .status-dot.critical { background: #F43F5E; }
            .status-dot.completed { background: #00FFA3; }
            .status-dot.delayed { background: #f97316; }
            .status-dot.pending { background: var(--fb-text-tertiary, #5A5B66); }

            .tooltip-float {
                font-family: var(--fb-font-mono, monospace);
                color: var(--fb-text-tertiary, #5A5B66);
                font-size: 11px;
                margin-top: 4px;
            }
        `,
    ];

    /** Task data to render */
    @property({ attribute: false })
    task: TaskBarData | null = null;

    /** Bar width in pixels */
    @property({ type: Number }) width = 0;

    /** Float bar width (late_finish - early_finish) */
    @property({ type: Number }) floatWidth = 0;

    /** Whether to show float bar */
    @property({ type: Boolean, attribute: 'show-float' }) showFloat = false;

    private _getStatusClass(): string {
        if (!this.task) return 'pending';
        switch (this.task.status) {
            case TaskStatus.Completed:
                return 'completed';
            case TaskStatus.Delayed:
                return 'delayed';
            case TaskStatus.InProgress:
                return this.task.is_critical ? 'critical' : 'normal';
            default:
                return this.task.is_critical ? 'critical' : 'normal';
        }
    }

    private _getStatusLabel(): string {
        if (!this.task) return 'Pending';
        switch (this.task.status) {
            case TaskStatus.Completed:
                return 'Completed';
            case TaskStatus.Delayed:
                return 'Delayed';
            case TaskStatus.InProgress:
                return this.task.is_critical ? 'Critical Path' : 'In Progress';
            default:
                return this.task.is_critical ? 'Critical Path' : 'Pending';
        }
    }

    private _formatDate(iso: string): string {
        const d = new Date(iso + 'T00:00:00');
        return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    }

    private _handleClick() {
        if (!this.task) return;
        this.emit('fb-task-bar-click', {
            wbsCode: this.task.wbs_code,
            name: this.task.name,
            task: this.task,
        });
    }

    private _renderTooltip(): TemplateResult {
        if (!this.task) return html``;
        const statusClass = this._getStatusClass();

        return html`
            <div class="tooltip">
                <div class="tooltip-title">${this.task.name}</div>
                <div class="tooltip-wbs">${this.task.wbs_code}</div>
                <div class="tooltip-dates">
                    ${this._formatDate(this.task.early_start)} → ${this._formatDate(this.task.early_finish)}
                    <span style="color: var(--fb-text-tertiary)">(${this.task.duration_days}d)</span>
                </div>
                <div class="tooltip-status">
                    <span class="status-dot ${statusClass}"></span>
                    ${this._getStatusLabel()}
                    ${this.task.progress !== undefined && this.task.progress > 0
                        ? html`<span style="margin-left: auto;">${this.task.progress}%</span>`
                        : nothing}
                </div>
                ${this.task.total_float !== undefined && this.task.total_float > 0
                    ? html`<div class="tooltip-float">${this.task.total_float}d float</div>`
                    : nothing}
            </div>
        `;
    }

    override render(): TemplateResult {
        if (!this.task) return html``;

        const statusClass = this._getStatusClass();
        const progress = this.task.progress ?? 0;
        const showProgress = progress > 0 && progress < 100;

        return html`
            <div
                class="task-bar-wrapper"
                @click=${this._handleClick}
                role="button"
                tabindex="0"
                aria-label="${this.task.name}: ${this.task.early_start} to ${this.task.early_finish}"
            >
                <!-- Main task bar -->
                <div
                    class="task-bar ${statusClass}"
                    style="width: ${this.width}px"
                >
                    ${showProgress
                        ? html`<div class="progress-fill" style="width: ${progress}%"></div>`
                        : nothing}
                </div>

                <!-- Float bar (ghost bar showing slack) -->
                ${this.showFloat && this.floatWidth > 0
                    ? html`
                          <div
                              class="float-bar"
                              style="left: ${this.width}px; width: ${this.floatWidth}px"
                          ></div>
                      `
                    : nothing}

                <!-- Critical path indicator -->
                ${this.task.is_critical
                    ? html`<div class="critical-indicator"></div>`
                    : nothing}

                ${this._renderTooltip()}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-schedule-task-bar': FBScheduleTaskBar;
    }
}

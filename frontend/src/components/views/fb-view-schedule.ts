/**
 * fb-view-schedule — Timeline-based Gantt chart.
 * See FRONTEND_V2_SPEC.md §2.3.E, Phase 6 Steps 35-38
 *
 * Full-width horizontal timeline with:
 * - Date axis (month + week markers)
 * - Task bars positioned by early_start → early_finish
 * - Critical path highlighting (red bars)
 * - SVG dependency arrows (bezier curves)
 * - Today marker line
 * - Click task → emit detail event
 */
import { html, css, svg, nothing, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBViewElement } from '../base/FBViewElement';
import { store } from '../../store/store';
import { api } from '../../services/api';
import type { GanttData, GanttTask, GanttDependency } from '../../types/models';
import { TaskStatus } from '../../types/enums';

const ROW_HEIGHT = 40;
const LABEL_WIDTH = 220;
const DAY_WIDTH = 18;

function parseDate(iso: string): Date {
    return new Date(iso + 'T00:00:00');
}

function daysBetween(a: Date, b: Date): number {
    return Math.round((b.getTime() - a.getTime()) / (1000 * 60 * 60 * 24));
}

function formatMonth(d: Date): string {
    return d.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
}

@customElement('fb-view-schedule')
export class FBViewSchedule extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                overflow: hidden;
            }

            .toolbar {
                display: flex;
                align-items: center;
                gap: 12px;
                padding: 12px 20px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
                flex-shrink: 0;
            }

            .toolbar-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .toolbar-spacer { flex: 1; }

            .toolbar-info {
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .recalc-btn {
                padding: 6px 14px;
                border-radius: 6px;
                font-size: 13px;
                font-weight: 500;
                cursor: pointer;
                border: 1px solid var(--fb-border, #2a2a3e);
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                transition: all 0.15s;
            }

            .recalc-btn:hover {
                border-color: var(--fb-accent, #6366f1);
                color: var(--fb-accent, #6366f1);
            }

            .recalc-btn:disabled { opacity: 0.5; cursor: not-allowed; }

            .scroll-container {
                flex: 1;
                overflow: auto;
                position: relative;
            }

            .gantt-canvas {
                position: relative;
                min-width: 100%;
            }

            .date-axis {
                position: sticky;
                top: 0;
                z-index: 2;
                display: flex;
                background: var(--fb-bg-primary, #0f0f1a);
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .date-axis-label-col {
                width: 220px;
                min-width: 220px;
                flex-shrink: 0;
                padding: 8px 12px;
                font-size: 11px;
                font-weight: 600;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
                letter-spacing: 0.05em;
                display: flex;
                align-items: flex-end;
            }

            .date-axis-timeline {
                flex: 1;
                position: relative;
                height: 48px;
            }

            .month-label {
                position: absolute;
                top: 4px;
                font-size: 11px;
                font-weight: 600;
                color: var(--fb-text-secondary, #a0a0b0);
                white-space: nowrap;
            }

            .week-tick {
                position: absolute;
                bottom: 0;
                width: 1px;
                height: 12px;
                background: var(--fb-border, #2a2a3e);
            }

            .week-label {
                position: absolute;
                bottom: 14px;
                font-size: 10px;
                color: var(--fb-text-tertiary, #707080);
                transform: translateX(-50%);
                white-space: nowrap;
            }

            .task-rows { position: relative; }

            .task-row {
                display: flex;
                height: 40px;
                border-bottom: 1px solid var(--fb-border-light, #1e1e32);
                cursor: pointer;
                transition: background 0.1s;
            }

            .task-row:hover { background: var(--fb-surface-1, #1a1a2e); }

            .task-label {
                width: 220px;
                min-width: 220px;
                flex-shrink: 0;
                display: flex;
                align-items: center;
                gap: 8px;
                padding: 0 12px;
                font-size: 13px;
                color: var(--fb-text-primary, #e0e0e0);
                overflow: hidden;
            }

            .task-wbs {
                font-size: 11px;
                color: var(--fb-text-tertiary, #707080);
                font-family: monospace;
                flex-shrink: 0;
            }

            .task-name {
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .task-timeline {
                flex: 1;
                position: relative;
            }

            .task-bar {
                position: absolute;
                height: 24px;
                top: 8px;
                border-radius: 4px;
                min-width: 4px;
            }

            .task-bar.normal { background: var(--fb-accent, #6366f1); }
            .task-bar.critical { background: #ef4444; }
            .task-bar.completed { background: #22c55e; }
            .task-bar.delayed { background: #f97316; }

            .dep-overlay {
                position: absolute;
                top: 0;
                left: 220px;
                pointer-events: none;
                z-index: 1;
            }

            .dep-arrow {
                fill: none;
                stroke: var(--fb-text-tertiary, #707080);
                stroke-width: 1.5;
                opacity: 0.5;
            }

            .dep-arrowhead {
                fill: var(--fb-text-tertiary, #707080);
                opacity: 0.5;
            }

            .today-line {
                position: absolute;
                top: 0;
                bottom: 0;
                width: 2px;
                background: var(--fb-accent, #6366f1);
                opacity: 0.4;
                z-index: 1;
            }

            .loading, .empty, .error-state {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                height: 200px;
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 14px;
                gap: 12px;
            }

            .error-state {
                color: #ef4444;
            }

            .retry-btn {
                padding: 8px 20px;
                border-radius: 6px;
                border: 1px solid #ef4444;
                background: transparent;
                color: #ef4444;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .retry-btn:hover {
                background: #ef4444;
                color: #fff;
            }

            .legend {
                display: flex;
                gap: 16px;
                padding: 8px 20px;
                border-top: 1px solid var(--fb-border, #2a2a3e);
                flex-shrink: 0;
            }

            .legend-item {
                display: flex;
                align-items: center;
                gap: 6px;
                font-size: 12px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .legend-dot {
                width: 10px;
                height: 10px;
                border-radius: 2px;
            }

            @media (max-width: 768px) {
                .toolbar {
                    flex-direction: column;
                    gap: 8px;
                    padding: 12px;
                }

                .task-label {
                    width: 120px;
                    min-width: 120px;
                    font-size: 11px;
                }

                .date-axis-label-col {
                    width: 120px;
                    min-width: 120px;
                }

                .legend {
                    flex-wrap: wrap;
                    padding: 8px 12px;
                }
            }
        `,
    ];

    @state() private _data: GanttData | null = null;
    @state() private _loading = false;
    @state() private _error: string | null = null;
    @state() private _projectId: string | null = null;
    @state() private _recalculating = false;

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffects.push(
            effect(() => {
                const pid = store.contextProjectId$.value;
                if (pid && pid !== this._projectId) {
                    this._projectId = pid;
                    void this._loadSchedule(pid);
                }
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private async _loadSchedule(projectId: string): Promise<void> {
        this._loading = true;
        this._error = null;
        try {
            this._data = await api.schedule.get(projectId);
        } catch (err) {
            console.error('[FBViewSchedule] Failed to load schedule:', err);
            this._data = null;
            this._error = err instanceof Error ? err.message : 'Failed to load schedule';
        } finally {
            this._loading = false;
        }
    }

    private async _handleRecalculate(): Promise<void> {
        if (!this._projectId || this._recalculating) return;
        this._recalculating = true;
        try {
            this._data = await api.schedule.recalculate(this._projectId);
        } catch {
            // Keep existing data on failure
        } finally {
            this._recalculating = false;
        }
    }

    private _getTimelineBounds(): { startDate: Date; totalDays: number } | null {
        if (!this._data?.tasks.length) return null;
        let min = Infinity;
        let max = -Infinity;
        for (const t of this._data.tasks) {
            const s = parseDate(t.early_start).getTime();
            const e = parseDate(t.early_finish).getTime();
            if (s < min) min = s;
            if (e > max) max = e;
        }
        const startDate = new Date(min - 7 * 86400000);
        const endDate = new Date(max + 7 * 86400000);
        return { startDate, totalDays: daysBetween(startDate, endDate) };
    }

    private _dayToX(date: Date, startDate: Date): number {
        return daysBetween(startDate, date) * DAY_WIDTH;
    }

    private _renderDateAxis(startDate: Date, totalDays: number): TemplateResult {
        const months: TemplateResult[] = [];
        const weeks: TemplateResult[] = [];
        let lastMonth = -1;

        for (let d = 0; d <= totalDays; d++) {
            const date = new Date(startDate.getTime() + d * 86400000);
            const m = date.getMonth();

            if (m !== lastMonth) {
                months.push(html`<span class="month-label" style="left: ${d * DAY_WIDTH}px">${formatMonth(date)}</span>`);
                lastMonth = m;
            }

            if (date.getDay() === 1) {
                const x = d * DAY_WIDTH;
                weeks.push(html`
                    <span class="week-tick" style="left: ${x}px"></span>
                    <span class="week-label" style="left: ${x}px">${date.getDate()}</span>
                `);
            }
        }

        return html`
            <div class="date-axis">
                <div class="date-axis-label-col">Tasks</div>
                <div class="date-axis-timeline" style="width: ${totalDays * DAY_WIDTH}px">
                    ${months} ${weeks}
                </div>
            </div>
        `;
    }

    private _getBarClass(task: GanttTask): string {
        if (task.status === TaskStatus.Completed) return 'completed';
        if (task.status === TaskStatus.Delayed) return 'delayed';
        if (task.is_critical) return 'critical';
        return 'normal';
    }

    private _renderTaskRow(task: GanttTask, startDate: Date): TemplateResult {
        const barStart = parseDate(task.early_start);
        const barEnd = parseDate(task.early_finish);
        const x = this._dayToX(barStart, startDate);
        const width = Math.max(daysBetween(barStart, barEnd) * DAY_WIDTH, 4);

        return html`
            <div class="task-row" role="listitem" @click=${() => this.emit('fb-task-selected', { wbsCode: task.wbs_code, name: task.name })}
                aria-label="${task.wbs_code} ${task.name}: ${task.early_start} to ${task.early_finish}">
                <div class="task-label">
                    <span class="task-wbs">${task.wbs_code}</span>
                    <span class="task-name">${task.name}</span>
                </div>
                <div class="task-timeline">
                    <div
                        class="task-bar ${this._getBarClass(task)}"
                        style="left: ${x}px; width: ${width}px"
                        title="${task.name}: ${task.early_start} \u2192 ${task.early_finish} (${task.duration_days}d)"
                    ></div>
                </div>
            </div>
        `;
    }

    private _renderDependencyArrows(tasks: GanttTask[], deps: GanttDependency[], startDate: Date): TemplateResult {
        const taskIndex = new Map<string, number>();
        tasks.forEach((t, i) => { taskIndex.set(t.wbs_code, i); });

        const taskMap = new Map<string, GanttTask>();
        tasks.forEach((t) => { taskMap.set(t.wbs_code, t); });

        const canvasWidth = this._getTimelineBounds()!.totalDays * DAY_WIDTH;
        const canvasHeight = tasks.length * ROW_HEIGHT;

        const arrows = deps.map((dep) => {
            const fromTask = taskMap.get(dep.from);
            const toTask = taskMap.get(dep.to);
            const fromIdx = taskIndex.get(dep.from);
            const toIdx = taskIndex.get(dep.to);
            if (!fromTask || !toTask || fromIdx === undefined || toIdx === undefined) return nothing;

            const x1 = this._dayToX(parseDate(fromTask.early_finish), startDate);
            const y1 = fromIdx * ROW_HEIGHT + ROW_HEIGHT / 2;
            const x2 = this._dayToX(parseDate(toTask.early_start), startDate);
            const y2 = toIdx * ROW_HEIGHT + ROW_HEIGHT / 2;
            const midX = (x1 + x2) / 2;

            const path = `M ${x1} ${y1} C ${midX} ${y1}, ${midX} ${y2}, ${x2} ${y2}`;
            const aSize = 5;
            const arrowhead = `M ${x2} ${y2} L ${x2 - aSize} ${y2 - aSize} L ${x2 - aSize} ${y2 + aSize} Z`;

            return svg`
                <path class="dep-arrow" d=${path}/>
                <path class="dep-arrowhead" d=${arrowhead}/>
            `;
        });

        return html`
            <svg class="dep-overlay" width=${canvasWidth} height=${canvasHeight}>
                ${arrows}
            </svg>
        `;
    }

    private _renderTodayLine(startDate: Date, totalDays: number): TemplateResult | typeof nothing {
        const now = new Date();
        now.setHours(0, 0, 0, 0);
        const daysFromStart = daysBetween(startDate, now);
        if (daysFromStart < 0 || daysFromStart > totalDays) return nothing;
        const x = LABEL_WIDTH + daysFromStart * DAY_WIDTH;
        return html`<div class="today-line" style="left: ${x}px"></div>`;
    }

    override render(): TemplateResult {
        if (this._loading) {
            return html`
                <div class="toolbar"><span class="toolbar-title">Schedule</span></div>
                <div class="loading" aria-busy="true">Loading schedule...</div>
            `;
        }

        if (this._error) {
            return html`
                <div class="toolbar"><span class="toolbar-title">Schedule</span></div>
                <div class="error-state" role="alert">
                    <span>${this._error}</span>
                    <button class="retry-btn" @click=${() => { if (this._projectId) void this._loadSchedule(this._projectId); }}>
                        Retry
                    </button>
                </div>
            `;
        }

        if (!this._data || !this._data.tasks.length) {
            return html`
                <div class="toolbar"><span class="toolbar-title">Schedule</span></div>
                <div class="empty">No schedule data available. Create tasks to generate a schedule.</div>
            `;
        }

        const bounds = this._getTimelineBounds()!;
        const { startDate, totalDays } = bounds;
        const canvasWidth = LABEL_WIDTH + totalDays * DAY_WIDTH;
        const projEnd = this._data.projected_end_date;

        return html`
            <div class="toolbar" role="toolbar" aria-label="Schedule controls">
                <span class="toolbar-title">Schedule</span>
                <span class="toolbar-spacer"></span>
                ${projEnd ? html`
                    <span class="toolbar-info">Projected end: ${new Date(projEnd + 'T00:00:00').toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</span>
                ` : nothing}
                <button class="recalc-btn" ?disabled=${this._recalculating} @click=${this._handleRecalculate.bind(this)}
                    aria-label="${this._recalculating ? 'Recalculating schedule' : 'Recalculate schedule'}">
                    ${this._recalculating ? 'Recalculating...' : 'Recalculate'}
                </button>
            </div>

            <div class="scroll-container" role="region" aria-label="Gantt timeline" tabindex="0">
                <div class="gantt-canvas" style="width: ${canvasWidth}px">
                    ${this._renderDateAxis(startDate, totalDays)}
                    <div class="task-rows" role="list" aria-label="Project tasks">
                        ${this._data.tasks.map((t) => this._renderTaskRow(t, startDate))}
                        ${(this._data.dependencies?.length ?? 0) > 0
                ? this._renderDependencyArrows(this._data.tasks, this._data.dependencies!, startDate)
                : nothing}
                    </div>
                    ${this._renderTodayLine(startDate, totalDays)}
                </div>
            </div>

            <div class="legend" role="complementary" aria-label="Chart legend">
                <div class="legend-item"><span class="legend-dot" style="background: var(--fb-accent, #6366f1)"></span> Normal</div>
                <div class="legend-item"><span class="legend-dot" style="background: #ef4444"></span> Critical Path</div>
                <div class="legend-item"><span class="legend-dot" style="background: #22c55e"></span> Completed</div>
                <div class="legend-item"><span class="legend-dot" style="background: #f97316"></span> Delayed</div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-schedule': FBViewSchedule;
    }
}

import { html, css, svg, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { GanttArtifactData } from '../../types/artifacts';
import { GanttTask, GanttDependency } from '../../types/models';
import { TaskStatus } from '../../types/enums';

/** Height of each task row group (row + bar + padding) in pixels. */
const ROW_HEIGHT = 42;
/** Vertical offset for the first row from the top of the task list. */
const TOP_OFFSET = 0;
/** Horizontal padding for arrow start/end. */
const ARROW_PAD_X = 12;

@customElement('fb-artifact-gantt')
export class FBArtifactGantt extends FBElement {
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

            .gantt-wrapper {
                position: relative;
            }

            .gantt-container {
                padding: var(--fb-spacing-md);
            }

            .timeline-header {
                display: flex;
                margin-bottom: var(--fb-spacing-sm);
                border-bottom: 1px solid var(--fb-border-light);
                padding-bottom: var(--fb-spacing-xs);
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            .col-task { flex: 1; }
            .col-date { width: 80px; text-align: center; }

            .task-list {
                position: relative;
            }

            .task-row {
                display: flex;
                align-items: center;
                padding: var(--fb-spacing-sm) 0;
                font-size: var(--fb-text-sm);
                border-radius: var(--fb-radius-sm);
                transition: background 0.15s ease;
            }

            /* Step 88: Critical path row highlight */
            .task-row.critical {
                background: rgba(244, 63, 94, 0.05);
            }

            .task-name {
                flex: 1;
                font-weight: 500;
                color: var(--fb-text-primary);
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .status-dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background: var(--fb-text-muted);
                flex-shrink: 0;
            }

            .status-dot.completed { background: var(--fb-success); }
            .status-dot.in-progress { background: var(--fb-primary); }
            .status-dot.delayed { background: var(--fb-error); }

            /* Step 88: Critical path flame icon */
            .critical-icon {
                flex-shrink: 0;
                font-size: 12px;
                line-height: 1;
            }

            .task-bar-container {
                margin-top: 4px;
                height: 6px;
                background: var(--fb-bg-tertiary);
                border-radius: 3px;
                width: 100%;
                position: relative;
            }

            .task-bar {
                height: 100%;
                border-radius: 3px;
                background: var(--fb-primary);
                position: absolute;
            }

            /* Step 88: Critical path bar color */
            .task-bar.critical {
                background: #F43F5E;
            }

            .task-meta {
                width: 80px;
                text-align: right;
                font-family: var(--fb-font-mono, monospace);
                font-size: var(--fb-text-xs);
                color: var(--fb-text-secondary);
            }

            /* Step 88: Tooltip for critical path */
            .task-group {
                position: relative;
            }

            .critical-tooltip {
                display: none;
                position: absolute;
                right: 0;
                top: 0;
                background: rgba(244, 63, 94, 0.9);
                color: white;
                font-size: var(--fb-text-xs);
                padding: 2px 8px;
                border-radius: var(--fb-radius-sm);
                white-space: nowrap;
                z-index: 5;
                pointer-events: none;
            }

            .task-group:hover .critical-tooltip {
                display: block;
            }

            /* Step 89: SVG dependency layer */
            .dependency-layer {
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                pointer-events: none;
                z-index: 2;
            }

            /* Skeleton styles inherited from FBElement */
        `
    ];

    @property({ attribute: false })
    data: GanttArtifactData | null = null;

    @state() private _hoveredTask: string | null = null;

    private _renderSkeleton(): TemplateResult {
        return html`
            <div class="gantt-container">
                <div class="timeline-header">
                    <div class="col-task">Task Phase</div>
                    <div class="col-date">Duration</div>
                </div>
                ${[1, 2, 3, 4].map(() => html`
                    <div class="task-group">
                         <div class="task-row">
                             <div class="skeleton skeleton-box" style="width: 60%"></div>
                         </div>
                    </div>
                `)}
            </div>
        `;
    }

    private _getBarWidth(task: GanttTask): number {
        if (task.status === TaskStatus.Completed) return 100;
        if (task.status === TaskStatus.InProgress) return 50;
        return 0;
    }

    // ========================================================================
    // Step 89: Dependency resolution
    // ========================================================================

    /**
     * Resolve dependencies from the data. Uses explicit dependencies if available,
     * otherwise infers from the critical_path array (consecutive pairs).
     */
    private _resolveDependencies(): GanttDependency[] {
        if (!this.data) return [];

        // Prefer explicit dependencies
        if (this.data.dependencies && this.data.dependencies.length > 0) {
            return this.data.dependencies;
        }

        // Fallback: infer from critical_path (consecutive pairs form FS dependencies)
        const cp = this.data.critical_path;
        if (!cp || cp.length < 2) return [];

        const deps: GanttDependency[] = [];
        for (let i = 0; i < cp.length - 1; i++) {
            const from = cp[i];
            const to = cp[i + 1];
            if (from && to) deps.push({ from, to });
        }
        return deps;
    }

    /**
     * Build a WBS code to row index map for coordinate calculation.
     */
    private _buildTaskIndexMap(): Map<string, number> {
        const map = new Map<string, number>();
        if (this.data) {
            this.data.tasks.forEach((task, i) => {
                map.set(task.wbs_code, i);
            });
        }
        return map;
    }

    /**
     * Render SVG dependency arrows as bezier curves.
     */
    private _renderDependencyLines(): TemplateResult {
        const deps = this._resolveDependencies();
        if (deps.length === 0) return html``;

        const indexMap = this._buildTaskIndexMap();
        const totalHeight = (this.data?.tasks.length ?? 0) * ROW_HEIGHT;

        return html`
            <svg
                class="dependency-layer"
                viewBox="0 0 24 ${totalHeight}"
                preserveAspectRatio="none"
                style="height: ${totalHeight}px"
                aria-hidden="true"
            >
                <defs>
                    <marker
                        id="arrowhead"
                        viewBox="0 0 10 7"
                        refX="10"
                        refY="3.5"
                        markerWidth="8"
                        markerHeight="6"
                        orient="auto-start-reverse"
                    >
                        ${svg`<polygon points="0 0, 10 3.5, 0 7" fill="var(--fb-border, #555)" />`}
                    </marker>
                    <marker
                        id="arrowhead-highlight"
                        viewBox="0 0 10 7"
                        refX="10"
                        refY="3.5"
                        markerWidth="8"
                        markerHeight="6"
                        orient="auto-start-reverse"
                    >
                        ${svg`<polygon points="0 0, 10 3.5, 0 7" fill="var(--fb-primary, #00FFA3)" />`}
                    </marker>
                </defs>
                ${deps.map(dep => {
                    if (dep.from === dep.to) return nothing; // Skip self-referencing dependencies
                    const fromIdx = indexMap.get(dep.from);
                    const toIdx = indexMap.get(dep.to);
                    if (fromIdx === undefined || toIdx === undefined) return nothing;

                    const isHighlighted = this._hoveredTask === dep.from || this._hoveredTask === dep.to;

                    // Calculate Y positions (center of each row)
                    const y1 = TOP_OFFSET + fromIdx * ROW_HEIGHT + ROW_HEIGHT / 2;
                    const y2 = TOP_OFFSET + toIdx * ROW_HEIGHT + ROW_HEIGHT / 2;

                    // X positions: right edge (from) -> left edge (to)
                    const x1 = 24 - ARROW_PAD_X;
                    const x2 = ARROW_PAD_X;

                    // Bezier control points for a smooth curve
                    const cpx1 = x1 + 4;
                    const cpx2 = x2 - 4;

                    const stroke = isHighlighted ? 'var(--fb-primary, #00FFA3)' : 'var(--fb-border, #555)';
                    const marker = isHighlighted ? 'url(#arrowhead-highlight)' : 'url(#arrowhead)';
                    const strokeWidth = isHighlighted ? '0.4' : '0.25';

                    return svg`
                        <path
                            d="M ${x1} ${y1} C ${cpx1} ${y1}, ${cpx2} ${y2}, ${x2} ${y2}"
                            fill="none"
                            stroke="${stroke}"
                            stroke-width="${strokeWidth}"
                            marker-end="${marker}"
                            opacity="${isHighlighted ? 1 : 0.5}"
                        />
                    `;
                })}
            </svg>
        `;
    }

    // ========================================================================
    // Task rendering
    // ========================================================================

    private _renderTaskRow(task: GanttTask): TemplateResult {
        const isCritical = task.is_critical;
        const statusClass = task.status.toLowerCase().replace('_', '-');

        return html`
            <div
                class="task-group"
                @mouseenter=${() => { this._hoveredTask = task.wbs_code; }}
                @mouseleave=${() => { this._hoveredTask = null; }}
            >
                <div class="task-row ${isCritical ? 'critical' : ''}">
                    <div class="task-name">
                        <span class="status-dot ${statusClass}"></span>
                        ${isCritical ? html`<span class="critical-icon" title="Critical Path">&#128293;</span>` : nothing}
                        ${task.name}
                    </div>
                    <div class="task-meta">${task.duration_days}d</div>
                </div>
                <div class="task-bar-container">
                    <div
                        class="task-bar ${isCritical ? 'critical' : ''}"
                        style="width: ${this._getBarWidth(task)}%"
                    ></div>
                </div>
                ${isCritical ? html`<span class="critical-tooltip">Critical Path</span>` : nothing}
            </div>
        `;
    }

    override render(): TemplateResult {
        if (!this.data) return this._renderSkeleton();

        return html`
            <div class="gantt-wrapper">
                <div class="gantt-container">
                    <div class="timeline-header">
                        <div class="col-task">Task Phase</div>
                        <div class="col-date">Duration</div>
                    </div>
                    <div class="task-list">
                        ${this._renderDependencyLines()}
                        ${this.data.tasks.map(task => this._renderTaskRow(task))}
                    </div>
                </div>
            </div>
        `;
    }
}

import { html, css, TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { GanttArtifactData } from '../../types/artifacts';
import { TaskStatus } from '../../types/enums';

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

            .task-row {
                display: flex;
                align-items: center;
                padding: var(--fb-spacing-sm) 0;
                font-size: var(--fb-text-sm);
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
            }

            .status-dot.completed { background: var(--fb-success); }
            .status-dot.in-progress { background: var(--fb-primary); }
            .status-dot.delayed { background: var(--fb-error); }

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

            .task-meta {
                width: 80px;
                text-align: right;
                font-size: var(--fb-text-xs);
                color: var(--fb-text-secondary);
            }
            /* Skeleton styles inherited from FBElement */
        `
    ];

    @property({ attribute: false })
    data: GanttArtifactData | null = null;


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

    override render(): TemplateResult {
        if (!this.data) return this._renderSkeleton();

        return html`
            <div class="gantt-container">
                <div class="timeline-header">
                    <div class="col-task">Task Phase</div>
                    <div class="col-date">Duration</div>
                </div>
                ${this.data.tasks.map(task => html`
                    <div class="task-group">
                        <div class="task-row">
                            <div class="task-name">
                                <span class="status-dot ${task.status.toLowerCase().replace('_', '-')}"></span>
                                ${task.name}
                            </div>
                            <div class="task-meta">${task.duration_days}d</div>
                        </div>
                        <div class="task-bar-container">
                            <div class="task-bar" style="width: ${task.status === TaskStatus.Completed ? 100 :
                task.status === TaskStatus.InProgress ? 50 : 0
            }%"></div>
                        </div>
                    </div>
                `)}
            </div>
        `;
    }
}

import { html, css, TemplateResult } from 'lit';
import { customElement } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

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
        `
    ];

    // Placeholder mock data
    private _tasks = [
        { id: 1, name: 'Foundation', status: 'completed', progress: 100, start: 'Jan 1', end: 'Jan 10' },
        { id: 2, name: 'Framing', status: 'in-progress', progress: 60, start: 'Jan 12', end: 'Feb 15' },
        { id: 3, name: 'Rough Plumbing', status: 'pending', progress: 0, start: 'Feb 16', end: 'Feb 28' },
        { id: 4, name: 'Electrical', status: 'pending', progress: 0, start: 'Mar 1', end: 'Mar 14' },
    ];

    override render(): TemplateResult {
        return html`
            <div class="gantt-container">
                <div class="timeline-header">
                    <div class="col-task">Task Phase</div>
                    <div class="col-date">Duration</div>
                </div>
                ${this._tasks.map(task => html`
                    <div class="task-group">
                        <div class="task-row">
                            <div class="task-name">
                                <span class="status-dot ${task.status}"></span>
                                ${task.name}
                            </div>
                            <div class="task-meta">${task.start} - ${task.end}</div>
                        </div>
                        <div class="task-bar-container">
                            <div class="task-bar" style="width: ${task.progress}%"></div>
                        </div>
                    </div>
                `)}
            </div>
        `;
    }
}

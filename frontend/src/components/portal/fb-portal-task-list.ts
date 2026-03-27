/**
 * fb-portal-task-list — Mobile-optimized task list for field workers.
 *
 * Features: 48px+ touch targets, neon status badges, glassmorphism cards.
 * Phase 18: See FRONTEND_SCOPE.md §15.2 (Voice-First Field Portal)
 */
import { html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export interface PortalTask {
    id: string;
    wbs_code: string;
    name: string;
    status: string;
    percent_complete: number;
    assignee?: string;
    due_date?: string;
}

@customElement('fb-portal-task-list')
export class FBPortalTaskList extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                font-size: 16px;
            }

            .task-card {
                background: rgba(22, 24, 33, 0.6);
                backdrop-filter: blur(24px);
                -webkit-backdrop-filter: blur(24px);
                border: 1px solid rgba(255, 255, 255, 0.05);
                border-radius: 12px;
                padding: 16px;
                margin-bottom: 12px;
                min-height: 64px;
            }

            .task-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: 8px;
            }

            .task-name {
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
                line-height: 1.3;
            }

            .task-wbs {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
                font-family: monospace;
            }

            .task-meta {
                font-size: 14px;
                color: var(--fb-text-secondary, #8B8D98);
                margin-bottom: 12px;
            }

            .status-badge {
                display: inline-block;
                padding: 4px 10px;
                border-radius: 6px;
                font-size: 13px;
                font-weight: 700;
                text-transform: uppercase;
                letter-spacing: 0.5px;
            }

            .status-badge.passed,
            .status-badge.completed {
                background: rgba(0, 255, 163, 0.12);
                color: #00FFA3;
            }

            .status-badge.failed {
                background: rgba(244, 63, 94, 0.12);
                color: #F43F5E;
            }

            .status-badge.pending,
            .status-badge.in_progress {
                background: rgba(245, 158, 11, 0.12);
                color: #f59e0b;
            }

            .status-badge.not_started {
                background: rgba(139, 141, 152, 0.12);
                color: #8B8D98;
            }

            .progress-bar {
                height: 6px;
                background: rgba(255, 255, 255, 0.05);
                border-radius: 3px;
                overflow: hidden;
                margin: 8px 0 12px;
            }

            .progress-fill {
                height: 100%;
                background: var(--fb-accent, #00FFA3);
                border-radius: 3px;
                transition: width 0.3s ease;
            }

            .mark-complete-btn {
                width: 100%;
                height: 48px;
                border: none;
                border-radius: 8px;
                background: var(--fb-accent, #00FFA3);
                color: #0A0B10;
                font-size: 16px;
                font-weight: 700;
                cursor: pointer;
                transition: opacity 0.15s;
            }

            .mark-complete-btn:hover { opacity: 0.9; }
            .mark-complete-btn:active { opacity: 0.8; }
            .mark-complete-btn:disabled { opacity: 0.4; cursor: not-allowed; }

            .empty-state {
                text-align: center;
                padding: 40px 20px;
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 16px;
            }
        `,
    ];

    @property({ type: Array }) tasks: PortalTask[] = [];

    private _handleMarkComplete(taskId: string) {
        this.dispatchEvent(new CustomEvent('task-complete', {
            detail: { taskId },
            bubbles: true,
            composed: true,
        }));
    }

    override render() {
        if (this.tasks.length === 0) {
            return html`<div class="empty-state">No tasks assigned to you yet.</div>`;
        }

        return html`
            ${this.tasks.map(t => html`
                <div class="task-card">
                    <div class="task-header">
                        <span class="task-name">${t.name}</span>
                        <span class="task-wbs">${t.wbs_code}</span>
                    </div>

                    <div class="task-meta">
                        <span class="status-badge ${t.status.toLowerCase().replace(' ', '_')}">${t.status}</span>
                        ${t.due_date ? html` &middot; Due ${t.due_date}` : nothing}
                    </div>

                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${t.percent_complete}%"></div>
                    </div>

                    ${t.status !== 'Completed' && t.status !== 'completed' ? html`
                        <button class="mark-complete-btn" @click=${() => this._handleMarkComplete(t.id)}>
                            Mark Complete
                        </button>
                    ` : nothing}
                </div>
            `)}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-portal-task-list': FBPortalTaskList;
    }
}

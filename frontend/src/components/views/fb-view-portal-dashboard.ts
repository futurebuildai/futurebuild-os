/**
 * FBViewPortalDashboard - Portal Dashboard View
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Dashboard for contacts with permanent portal accounts.
 * Shows list of assigned tasks grouped by project.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';

import '../portal/fb-portal-shell';
import '../portal/fb-task-card';
import type { TaskCardData } from '../portal/fb-task-card';

interface TaskGroup {
    projectId: string;
    projectName: string;
    tasks: TaskCardData[];
}

/**
 * Portal dashboard view component.
 * @element fb-view-portal-dashboard
 */
@customElement('fb-view-portal-dashboard')
export class FBViewPortalDashboard extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: block;
            }

            .header {
                margin-bottom: 24px;
            }

            .greeting {
                color: var(--fb-text-primary, #fff);
                font-size: 24px;
                font-weight: 600;
                margin: 0 0 4px 0;
            }

            .subtitle {
                color: var(--fb-text-secondary, #aaa);
                font-size: 14px;
                margin: 0;
            }

            .loading {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                min-height: 200px;
                gap: 16px;
            }

            .spinner {
                width: 32px;
                height: 32px;
                border: 3px solid var(--fb-border, #333);
                border-top-color: var(--fb-primary, #667eea);
                border-radius: 50%;
                animation: spin 1s linear infinite;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            .project-group {
                margin-bottom: 24px;
            }

            .project-title {
                color: var(--fb-text-secondary, #aaa);
                font-size: 12px;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.5px;
                margin: 0 0 12px 0;
            }

            .task-list {
                display: flex;
                flex-direction: column;
                gap: 12px;
            }

            .empty-state {
                text-align: center;
                padding: 48px 24px;
            }

            .empty-icon {
                width: 64px;
                height: 64px;
                color: var(--fb-text-muted, #666);
                margin-bottom: 16px;
            }

            .empty-title {
                color: var(--fb-text-primary, #fff);
                font-size: 18px;
                font-weight: 600;
                margin: 0 0 8px 0;
            }

            .empty-text {
                color: var(--fb-text-secondary, #aaa);
                font-size: 14px;
                margin: 0;
            }

            .nav-bar {
                position: fixed;
                bottom: 0;
                left: 0;
                right: 0;
                display: flex;
                justify-content: space-around;
                padding: 12px 0;
                background: var(--fb-bg-secondary, #0a0a0a);
                border-top: 1px solid var(--fb-border, #333);
            }

            .nav-item {
                display: flex;
                flex-direction: column;
                align-items: center;
                gap: 4px;
                padding: 8px 16px;
                background: none;
                border: none;
                cursor: pointer;
                color: var(--fb-text-muted, #666);
                transition: color 0.2s ease;
            }

            .nav-item:hover,
            .nav-item--active {
                color: var(--fb-primary, #667eea);
            }

            .nav-item svg {
                width: 24px;
                height: 24px;
            }

            .nav-label {
                font-size: 11px;
                font-weight: 500;
            }

            .content {
                padding-bottom: 80px;
            }
        `,
    ];

    @state() private _loading = true;
    @state() private _contactName = 'there';
    @state() private _taskGroups: TaskGroup[] = [];

    override connectedCallback(): void {
        super.connectedCallback();
        void this._loadTasks();
    }

    private async _loadTasks(): Promise<void> {
        // TODO: Replace with actual API call: GET /api/v1/portal/tasks
        const isDev = (import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV === true;

        if (isDev) {
            // Mock data for development only
            await new Promise((resolve) => setTimeout(resolve, 1000));
            this._contactName = 'John';
            this._taskGroups = [
                {
                    projectId: '1',
                    projectName: '123 Main Street',
                    tasks: [
                        { id: 't1', wbsCode: '7.1.1', name: 'Rough-In Plumbing', status: 'in_progress', projectName: '123 Main Street', dueDate: 'Jan 28' },
                        { id: 't2', wbsCode: '7.2.1', name: 'Rough-In Electrical', status: 'pending', projectName: '123 Main Street', dueDate: 'Feb 1' },
                    ],
                },
                {
                    projectId: '2',
                    projectName: '456 Oak Avenue',
                    tasks: [
                        { id: 't3', wbsCode: '6.1.2', name: 'Frame Second Floor', status: 'completed', projectName: '456 Oak Avenue', hasPhotos: true },
                    ],
                },
            ];
        } else {
            // Production: empty until API is wired
            this._contactName = 'there';
            this._taskGroups = [];
        }
        this._loading = false;
    }

    private _handleTaskSelect(_e: CustomEvent<{ taskId: string }>): void {
        // TODO: Navigate to task detail view
    }

    private _renderLoading(): TemplateResult {
        return html`
            <div class="loading">
                <div class="spinner"></div>
                <span>Loading your tasks...</span>
            </div>
        `;
    }

    private _renderEmpty(): TemplateResult {
        return html`
            <div class="empty-state">
                <svg class="empty-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
                </svg>
                <h2 class="empty-title">All caught up!</h2>
                <p class="empty-text">You have no assigned tasks at the moment.</p>
            </div>
        `;
    }

    private _renderTasks(): TemplateResult {
        if (this._taskGroups.length === 0) {
            return this._renderEmpty();
        }

        return html`
            ${this._taskGroups.map(
                (group) => html`
                    <div class="project-group">
                        <h2 class="project-title">${group.projectName}</h2>
                        <div class="task-list">
                            ${group.tasks.map(
                                (task) => html`
                                    <fb-task-card
                                        .task=${task}
                                        @fb-task-select=${this._handleTaskSelect.bind(this)}
                                    ></fb-task-card>
                                `
                            )}
                        </div>
                    </div>
                `
            )}
        `;
    }

    override render(): TemplateResult {
        return html`
            <fb-portal-shell>
                <div class="content">
                    <div class="header">
                        <h1 class="greeting">Hi, ${this._contactName}!</h1>
                        <p class="subtitle">Here are your assigned tasks</p>
                    </div>

                    ${this._loading ? this._renderLoading() : this._renderTasks()}
                </div>

                <nav class="nav-bar">
                    <button class="nav-item nav-item--active">
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-7 14l-5-5 1.41-1.41L12 14.17l4.59-4.58L18 11l-6 6z"/>
                        </svg>
                        <span class="nav-label">Tasks</span>
                    </button>
                    <button class="nav-item">
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                        </svg>
                        <span class="nav-label">Profile</span>
                    </button>
                    <button class="nav-item">
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 17h-2v-2h2v2zm2.07-7.75l-.9.92C13.45 12.9 13 13.5 13 15h-2v-.5c0-1.1.45-2.1 1.17-2.83l1.24-1.26c.37-.36.59-.86.59-1.41 0-1.1-.9-2-2-2s-2 .9-2 2H8c0-2.21 1.79-4 4-4s4 1.79 4 4c0 .88-.36 1.68-.93 2.25z"/>
                        </svg>
                        <span class="nav-label">Help</span>
                    </button>
                </nav>
            </fb-portal-shell>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-portal-dashboard': FBViewPortalDashboard;
    }
}

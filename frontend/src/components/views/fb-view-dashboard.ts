/**
 * FBViewDashboard - Default Landing View
 * See PRODUCTION_PLAN.md Step 51.4, STEP_70_DASHBOARD_WIRING.md
 *
 * The home base for authenticated users.
 * Displays project overview, metrics derived from store signals,
 * and daily focus tasks driven by the agentic backend.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBViewElement } from '../base/FBViewElement';
import { store } from '../../store/store';
import { api, type Project as APIProject } from '../../services/api';
import type { FocusTask, ProjectSummary } from '../../store/types';
import type { StatusCardType } from '../widgets/fb-status-card';
import '../widgets/fb-status-card';

/**
 * Convert API Project to store ProjectSummary format.
 * Duplicated from fb-view-projects.ts to avoid cross-view imports.
 */
function mapAPIProjectToSummary(project: APIProject): ProjectSummary {
    return {
        id: project.id,
        name: project.name,
        address: project.address,
        status: project.status,
        completionPercentage: project.completion_percentage,
        createdAt: project.created_at,
        updatedAt: project.updated_at,
    };
}

@customElement('fb-view-dashboard')
export class FBViewDashboard extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                padding: var(--fb-spacing-xl);
            }

            .header {
                margin-bottom: var(--fb-spacing-xl);
            }

            h1 {
                font-size: var(--fb-text-2xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0 0 var(--fb-spacing-sm) 0;
            }

            .subtitle {
                color: var(--fb-text-secondary);
                font-size: var(--fb-text-base);
            }

            .metrics-grid {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                gap: var(--fb-spacing-lg);
                margin-bottom: var(--fb-spacing-xl);
            }

            .metric-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-lg);
            }

            .metric-label {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-xs);
            }

            .metric-value {
                font-size: var(--fb-text-2xl);
                font-weight: 700;
                color: var(--fb-text-primary);
            }

            .metric-trend {
                font-size: var(--fb-text-sm);
                margin-top: var(--fb-spacing-xs);
            }

            .trend-up {
                color: var(--fb-success);
            }

            .trend-down {
                color: var(--fb-error);
            }

            .section-title {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-md);
            }

            .placeholder-content {
                background: var(--fb-bg-card);
                border: 1px dashed var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-2xl);
                text-align: center;
                color: var(--fb-text-muted);
            }

            .task-list {
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-sm);
            }

        `,
    ];

    @state() private _userName = '';
    @state() private _focusTasks: FocusTask[] = [];
    @state() private _projects: ProjectSummary[] = [];

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffect = effect(() => {
            this._userName = store.user$.value?.name ?? 'Builder';
            this._focusTasks = store.focusTasks$.value;
            this._projects = store.projects$.value;
        });
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        this._disposeEffect = null;
        super.disconnectedCallback();
    }

    override onViewActive(): void {
        // Fetch projects to populate metrics if store is empty
        if (store.projects$.value.length === 0) {
            void this._fetchProjects();
        }
    }

    private async _fetchProjects(): Promise<void> {
        try {
            const apiProjects = await api.projects.list();
            store.actions.setProjects(apiProjects.map(mapAPIProjectToSummary));
        } catch (err) {
            console.error('[FBViewDashboard] Failed to fetch projects:', err);
        }
    }

    private get _activeProjectCount(): number {
        return this._projects.length;
    }

    private get _criticalTaskCount(): number {
        return this._focusTasks.filter(
            (t) => t.priority === 'high' || t.actionType === 'urgent'
        ).length;
    }

    private get _avgCompletion(): string {
        if (this._projects.length === 0) return '\u2014';
        const avg =
            this._projects.reduce((sum, p) => sum + p.completionPercentage, 0) /
            this._projects.length;
        return `${Math.round(avg)}%`;
    }

    private _mapPriorityToType(task: FocusTask): StatusCardType {
        if (task.actionType === 'urgent' || task.priority === 'high')
            return 'critical';
        if (task.priority === 'medium') return 'warning';
        return 'info';
    }

    private _mapPriorityToIcon(task: FocusTask): string {
        if (task.actionType === 'urgent') return 'priority_high';
        if (task.actionType === 'approval') return 'approval';
        if (task.actionType === 'review') return 'rate_review';
        return 'task_alt';
    }

    private _renderFocusTasks(): TemplateResult {
        if (this._focusTasks.length === 0) {
            return html`
                <div class="placeholder-content">
                    No focus tasks right now. The AI agent will surface priorities here.
                </div>
            `;
        }

        return html`
            <div class="task-list">
                ${this._focusTasks.map(
                    (task) => html`
                        <fb-status-card
                            type=${this._mapPriorityToType(task)}
                            title=${task.title}
                            subtitle="${task.description} — ${task.projectName}"
                            icon=${this._mapPriorityToIcon(task)}
                        ></fb-status-card>
                    `
                )}
            </div>
        `;
    }

    override render(): TemplateResult {
        return html`
            <div class="header">
                <h1>Welcome back, ${this._userName}</h1>
                <p class="subtitle">Here's what's happening with your projects today.</p>
            </div>

            <div class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-label">Active Projects</div>
                    <div class="metric-value">${this._activeProjectCount}</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">Critical Tasks</div>
                    <div class="metric-value">${this._criticalTaskCount}</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">Focus Items</div>
                    <div class="metric-value">${this._focusTasks.length}</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">Avg. Completion</div>
                    <div class="metric-value">${this._avgCompletion}</div>
                </div>
            </div>

            <h2 class="section-title">Daily Focus</h2>
            ${this._renderFocusTasks()}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-dashboard': FBViewDashboard;
    }
}

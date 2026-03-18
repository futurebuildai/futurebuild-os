/**
 * FBViewProjects - Projects Gallery View (V2 Simplified)
 * See FRONTEND_V2_SPEC.md §8 - V1 project components removed
 *
 * Displays projects in a grid. Clicking navigates to /project/:id.
 * "New Project" navigates to /onboard.
 *
 * Note: In V2, the primary project navigation is via top-bar pills.
 * This view exists for direct /projects route access.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBViewElement } from '../base/FBViewElement';
import { store } from '../../store/store';
import { api, Project as APIProject } from '../../services/api';
import type { ProjectSummary } from '../../store/types';

/**
 * Convert API Project to store ProjectSummary format.
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

@customElement('fb-view-projects')
export class FBViewProjects extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                padding: var(--fb-spacing-xl);
            }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: var(--fb-spacing-xl);
            }

            h1 {
                font-size: var(--fb-text-2xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0;
            }

            .btn-new {
                display: inline-flex;
                align-items: center;
                gap: var(--fb-spacing-xs);
                padding: var(--fb-spacing-sm) var(--fb-spacing-lg);
                background: var(--fb-primary, #00FFA3);
                border: none;
                border-radius: var(--fb-radius-md);
                color: white;
                font-size: var(--fb-text-sm);
                font-weight: 500;
                font-family: inherit;
                cursor: pointer;
                transition: background 0.15s ease;
            }

            .btn-new:hover {
                background: var(--fb-primary-hover, #5a67d8);
            }

            .btn-new:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .btn-new svg {
                width: 16px;
                height: 16px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .grid {
                display: grid;
                grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
                gap: var(--fb-spacing-lg);
            }

            /* V2 Simple Project Card */
            .project-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-lg);
                cursor: pointer;
                transition: border-color 0.15s ease, transform 0.15s ease;
            }

            .project-card:hover {
                border-color: var(--fb-primary);
                transform: translateY(-2px);
            }

            .project-card:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .project-name {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0 0 var(--fb-spacing-xs) 0;
            }

            .project-address {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin: 0 0 var(--fb-spacing-md) 0;
            }

            .project-meta {
                display: flex;
                justify-content: space-between;
                align-items: center;
            }

            .project-status {
                display: inline-flex;
                align-items: center;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                background: var(--fb-bg-tertiary);
                border-radius: var(--fb-radius-full, 9999px);
                font-size: var(--fb-text-xs);
                font-weight: 500;
                color: var(--fb-text-secondary);
            }

            .project-status[data-status="Active"] {
                background: rgba(34, 197, 94, 0.1);
                color: #00FFA3;
            }

            .project-status[data-status="Preconstruction"] {
                background: rgba(59, 130, 246, 0.1);
                color: #3b82f6;
            }

            .project-status[data-status="Completed"] {
                background: rgba(107, 114, 128, 0.1);
                color: #6b7280;
            }

            .project-completion {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-muted);
            }

            /* Empty State */
            .empty-state {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                padding: var(--fb-spacing-3xl, 64px) var(--fb-spacing-xl);
                text-align: center;
            }

            .empty-icon {
                width: 80px;
                height: 80px;
                margin-bottom: var(--fb-spacing-lg);
                color: var(--fb-text-muted);
            }

            .empty-icon svg {
                width: 100%;
                height: 100%;
                stroke: currentColor;
                fill: none;
                stroke-width: 1.5;
            }

            .empty-title {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0 0 var(--fb-spacing-sm) 0;
            }

            .empty-description {
                font-size: var(--fb-text-base);
                color: var(--fb-text-secondary);
                margin: 0 0 var(--fb-spacing-lg) 0;
                max-width: 400px;
            }

            /* Skeleton Cards */
            .skeleton-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-lg);
            }

            .skeleton-title {
                height: 24px;
                width: 70%;
                margin-bottom: var(--fb-spacing-sm);
            }

            .skeleton-subtitle {
                height: 16px;
                width: 90%;
                margin-bottom: var(--fb-spacing-md);
            }

            .skeleton-footer {
                display: flex;
                justify-content: space-between;
            }

            .skeleton-badge {
                height: 24px;
                width: 80px;
                border-radius: var(--fb-radius-full, 9999px);
            }

            .skeleton-progress {
                height: 16px;
                width: 60px;
            }
        `,
    ];

    @state() private _projects: ProjectSummary[] = [];
    @state() private _loading = true;

    private _disposeEffect?: () => void;

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffect = effect(() => {
            this._projects = store.projects$.value;
        });
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        super.disconnectedCallback();
    }

    override onViewActive(): void {
        void this._fetchProjects();
    }

    private async _fetchProjects(): Promise<void> {
        this._loading = true;
        try {
            const apiProjects = await api.projects.list();
            const projects = apiProjects.map(mapAPIProjectToSummary);
            store.actions.setProjects(projects);
        } catch (err) {
            console.error('[FBViewProjects] Failed to fetch projects:', err);
            store.actions.setProjectError(
                err instanceof Error ? err.message : 'Failed to load projects'
            );
        } finally {
            this._loading = false;
        }
    }

    private _handleNewProject(): void {
        // V2: Navigate to onboarding flow instead of opening dialog
        window.history.pushState({}, '', '/onboard');
        window.dispatchEvent(new PopStateEvent('popstate'));
    }

    private _handleProjectClick(projectId: string): void {
        // Navigate to project detail (filtered feed)
        window.history.pushState({}, '', `/project/${projectId}`);
        window.dispatchEvent(new PopStateEvent('popstate'));
    }

    private _renderSkeleton(): TemplateResult {
        return html`
            <div class="grid">
                ${[1, 2, 3].map(() => html`
                    <div class="skeleton-card">
                        <div class="skeleton skeleton-title"></div>
                        <div class="skeleton skeleton-subtitle"></div>
                        <div class="skeleton-footer">
                            <div class="skeleton skeleton-badge"></div>
                            <div class="skeleton skeleton-progress"></div>
                        </div>
                    </div>
                `)}
            </div>
        `;
    }

    private _renderEmptyState(): TemplateResult {
        return html`
            <div class="empty-state">
                <div class="empty-icon" aria-hidden="true">
                    <svg viewBox="0 0 24 24">
                        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
                        <polyline points="9 22 9 12 15 12 15 22" />
                    </svg>
                </div>
                <h2 class="empty-title">No projects yet</h2>
                <p class="empty-description">
                    Create your first project to get started. FutureBuild will help you manage schedules, budgets, and communication.
                </p>
                <button
                    class="btn-new"
                    @click=${this._handleNewProject.bind(this)}
                >
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <path d="M12 5v14M5 12h14" />
                    </svg>
                    Create First Project
                </button>
            </div>
        `;
    }

    private _renderGrid(): TemplateResult {
        return html`
            <div class="grid">
                ${this._projects.map(
                    (project) => html`
                        <div
                            class="project-card"
                            tabindex="0"
                            role="button"
                            aria-label="View ${project.name}"
                            @click=${(): void => { this._handleProjectClick(project.id); }}
                            @keydown=${(e: KeyboardEvent): void => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                    e.preventDefault();
                                    this._handleProjectClick(project.id);
                                }
                            }}
                        >
                            <h3 class="project-name">${project.name}</h3>
                            <p class="project-address">${project.address}</p>
                            <div class="project-meta">
                                <span class="project-status" data-status="${project.status}">
                                    ${project.status}
                                </span>
                                <span class="project-completion">
                                    ${project.completionPercentage}% complete
                                </span>
                            </div>
                        </div>
                    `
                )}
            </div>
        `;
    }

    override render(): TemplateResult {
        return html`
            <header class="header">
                <h1>Projects</h1>
                ${this._projects.length > 0 || this._loading ? html`
                    <button
                        class="btn-new"
                        @click=${this._handleNewProject.bind(this)}
                        ?disabled=${this._loading}
                    >
                        <svg viewBox="0 0 24 24" aria-hidden="true">
                            <path d="M12 5v14M5 12h14" />
                        </svg>
                        New Project
                    </button>
                ` : nothing}
            </header>

            ${this._loading
                ? this._renderSkeleton()
                : this._projects.length === 0
                    ? this._renderEmptyState()
                    : this._renderGrid()
            }
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-projects': FBViewProjects;
    }
}

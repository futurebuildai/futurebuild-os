/**
 * FBViewProjects - Projects Gallery View
 * See PROJECT_ONBOARDING_SPEC.md Step 62.5
 *
 * Displays all projects in a responsive grid layout with:
 * - Project cards showing name, address, status, and completion
 * - Empty state when no projects exist
 * - Loading skeleton during data fetch
 * - "New Project" button that opens creation dialog
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBViewElement } from '../base/FBViewElement';
import { store } from '../../store/store';
import { api, Project as APIProject } from '../../services/api';
import type { ProjectSummary, ProjectDetail } from '../../store/types';

// Import project components
import '../features/project/fb-project-card';
import '../features/project/fb-project-dialog';

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
                background: var(--fb-primary, #667eea);
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
    @state() private _dialogOpen = false;

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
        this._dialogOpen = true;
    }

    private _handleDialogClose(): void {
        this._dialogOpen = false;
    }

    private _handleProjectCreated(e: CustomEvent<ProjectDetail>): void {
        const project = e.detail;

        // Add to store as summary
        const summary: ProjectSummary = {
            id: project.id,
            name: project.name,
            address: project.address,
            status: project.status,
            completionPercentage: project.completionPercentage,
            createdAt: project.createdAt,
            updatedAt: project.updatedAt,
        };

        // Update store with new project
        store.actions.setProjects([...this._projects, summary]);

        // Close dialog
        this._dialogOpen = false;

        // Select the new project
        store.actions.selectProject(project.id);
    }

    private _handleProjectSelected(e: CustomEvent<{ id: string }>): void {
        const { id } = e.detail;
        store.actions.selectProject(id);
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
                        <fb-project-card
                            .project=${project}
                            @project-selected=${this._handleProjectSelected.bind(this)}
                        ></fb-project-card>
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

            <fb-project-dialog
                ?open=${this._dialogOpen}
                @close=${this._handleDialogClose.bind(this)}
                @project-created=${this._handleProjectCreated.bind(this)}
            ></fb-project-dialog>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-projects': FBViewProjects;
    }
}

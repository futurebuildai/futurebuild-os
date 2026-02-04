/**
 * FBPanelLeft - Left Panel Component (Projects, Threads, Daily Focus, Agent Activity)
 * See FRONTEND_SCOPE.md Section 3.3
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state, query } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import { api } from '../../services/api';
import type { ProjectSummary, Thread, FocusTask } from '../../store/types';
import type { Thread as ApiThread } from '../../types/models';
import { UserRole } from '../../types/enums';
import { clerkService } from '../../services/clerk';
import { isPlatformAdmin } from '../../services/platform-admin';
import '../agent/fb-agent-activity';

@customElement('fb-panel-left')
export class FBPanelLeft extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                background: var(--fb-bg-panel);
                border-right: 1px solid var(--fb-border);
                overflow: hidden;
            }

            .header {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-md);
                border-bottom: 1px solid var(--fb-border-light);
                cursor: pointer;
            }

            .header:hover {
                background: var(--fb-bg-tertiary);
            }

            .logo {
                width: 32px;
                height: 32px;
            }

            .brand {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .brand span {
                font-weight: 300;
            }

            .section {
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
            }

            .section-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: var(--fb-spacing-xs) 0;
                color: var(--fb-text-muted);
                font-size: var(--fb-text-xs);
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            .section-content {
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .item {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                border-radius: var(--fb-radius-md);
                cursor: pointer;
                color: var(--fb-text-secondary);
                font-size: var(--fb-text-sm);
                transition: background 0.15s ease, color 0.15s ease;
                border: none;
                background: transparent;
                width: 100%;
                text-align: left;
            }

            .item:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .item:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: -2px;
            }

            .item.active {
                background: var(--fb-bg-tertiary);
                color: var(--fb-primary);
            }

            .item-icon {
                width: 16px;
                height: 16px;
                flex-shrink: 0;
            }

            .item-icon svg {
                width: 100%;
                height: 100%;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .focus-badge {
                display: flex;
                align-items: center;
                justify-content: center;
                min-width: 18px;
                height: 18px;
                border-radius: 9px;
                background: var(--fb-primary);
                color: white;
                font-size: 10px;
                font-weight: 600;
            }

            .threads {
                margin-left: var(--fb-spacing-lg);
            }

            .thread-item {
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                font-size: var(--fb-text-xs);
            }

            .thread-item.unread {
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .projects-list {
                flex: 1;
                overflow-y: auto;
            }

            .user-section {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-md);
                border-top: 1px solid var(--fb-border-light);
            }

            .avatar {
                width: 32px;
                height: 32px;
                border-radius: 50%;
                background: var(--fb-dawn-gradient, linear-gradient(135deg, #667eea 0%, #764ba2 100%));
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: var(--fb-text-sm);
                font-weight: 600;
                color: white;
            }

            .user-info {
                flex: 1;
                min-width: 0;
            }

            .user-name {
                font-size: var(--fb-text-sm);
                font-weight: 500;
                color: var(--fb-text-primary);
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .user-role {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
            }

            .theme-toggle {
                padding: var(--fb-spacing-xs);
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-sm);
            }

            .theme-toggle:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .theme-toggle:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .theme-toggle svg {
                width: 18px;
                height: 18px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .admin-section {
                border-top: 1px solid var(--fb-border-light);
                margin-top: auto;
            }

            .admin-section .section-header {
                color: var(--fb-warning, #f59e0b);
            }

            .admin-item {
                color: var(--fb-text-secondary);
            }

            .admin-item:hover {
                color: var(--fb-warning, #f59e0b);
            }

            .add-btn {
                padding: 2px;
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-sm);
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .add-btn:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-primary);
            }

            .add-btn svg {
                width: 14px;
                height: 14px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .org-switcher {
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                border-bottom: 1px solid var(--fb-border-light);
            }

            /* Clerk org switcher renders inside shadow DOM but its CSS-in-JS
               stylesheets are injected into document.head and can't penetrate
               the shadow boundary. These rules restyle the unstyled Clerk elements. */
            .org-switcher .cl-organizationSwitcherTrigger {
                display: flex;
                align-items: center;
                gap: 8px;
                width: 100%;
                padding: 6px 8px;
                background: transparent;
                border: 1px solid var(--fb-border, #333);
                border-radius: var(--fb-radius-md, 8px);
                color: var(--fb-text-primary, #fff);
                cursor: pointer;
                font-family: inherit;
                font-size: var(--fb-text-sm, 13px);
            }

            .org-switcher .cl-organizationSwitcherTrigger:hover {
                background: var(--fb-bg-tertiary, #1a1a1a);
            }

            .org-switcher .cl-organizationPreview {
                display: flex;
                align-items: center;
                gap: 8px;
                min-width: 0;
                flex: 1;
            }

            .org-switcher .cl-organizationPreviewAvatarContainer,
            .org-switcher .cl-avatarBox,
            .org-switcher .cl-organizationPreviewAvatarBox {
                width: 24px !important;
                height: 24px !important;
                min-width: 24px;
                border-radius: 4px;
                overflow: hidden;
                flex-shrink: 0;
            }

            .org-switcher .cl-avatarImage {
                width: 100%;
                height: 100%;
                object-fit: cover;
            }

            .org-switcher .cl-organizationPreviewTextContainer {
                display: flex;
                flex-direction: column;
                min-width: 0;
                color: var(--fb-text-primary, #fff);
            }

            .org-switcher .cl-organizationPreviewTextContainer > * {
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .org-switcher .cl-organizationSwitcherTriggerIcon {
                width: 14px !important;
                height: 14px !important;
                min-width: 14px;
                color: var(--fb-text-muted, #666);
                flex-shrink: 0;
                fill: currentColor;
            }
        `,
    ];

    @query('#org-switcher-container') private _orgSwitcherContainer: HTMLDivElement | undefined;

    @state() private _projects: ProjectSummary[] = [];
    @state() private _threads: Thread[] = [];
    @state() private _focusTasks: FocusTask[] = [];
    @state() private _activeProjectId: string | null = null;
    @state() private _activeThreadId: string | null = null;
    @state() private _userName = '';
    @state() private _userInitials = '';
    @state() private _userRole: UserRole | null = null;
    @state() private _theme: 'light' | 'dark' | 'system' = 'system';
    @state() private _isCreatingThread = false;
    @state() private _newThreadTitle = '';
    @state() private _showCompleteConfirm = false;
    @state() private _isCompleting = false;
    @state() private _completeError: string | null = null;
    @state() private _isPlatformAdmin = false;

    private _disposeEffects: (() => void)[] = [];

    /** Maps an API Thread to a store Thread. */
    private _mapApiThread(t: ApiThread): Thread {
        const thread: Thread = {
            id: t.id,
            projectId: t.project_id,
            title: t.title,
            isGeneral: t.is_general,
            createdAt: t.created_at,
            updatedAt: t.updated_at,
            messages: [],
            hasUnread: false,
        };
        if (t.archived_at) {
            thread.archivedAt = t.archived_at;
        }
        return thread;
    }

    override connectedCallback(): void {
        super.connectedCallback();

        this._disposeEffects.push(
            effect(() => {
                this._projects = store.projects$.value;
            }),
            effect(() => {
                this._threads = store.projectThreads$.value;
            }),
            effect(() => {
                this._focusTasks = store.focusTasks$.value;
            }),
            effect(() => {
                const projectId = store.activeProjectId$.value;
                if (projectId && projectId !== this._activeProjectId) {
                    void this._loadThreads(projectId);
                }
                this._activeProjectId = projectId;
            }),
            effect(() => {
                this._activeThreadId = store.activeThreadId$.value;
            }),
            effect(() => {
                const user = store.user$.value;
                this._userName = user?.name ?? '';
                this._userInitials = this._computeInitials(user?.name);
                this._userRole = user?.role ?? null;
                this._isPlatformAdmin = isPlatformAdmin(user?.email ?? '');
            }),
            effect(() => {
                this._theme = store.theme$.value;
            })
        );
    }

    protected override firstUpdated(): void {
        if (this._orgSwitcherContainer) {
            clerkService.mountOrganizationSwitcher(this._orgSwitcherContainer);
        }
    }

    override disconnectedCallback(): void {
        if (this._orgSwitcherContainer) {
            clerkService.unmountOrganizationSwitcher(this._orgSwitcherContainer);
        }
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _computeInitials(name?: string): string {
        if (!name) return '?';
        const parts = name.split(' ');
        if (parts.length >= 2) {
            const first = parts[0]?.[0] ?? '';
            const second = parts[1]?.[0] ?? '';
            return `${first}${second}`.toUpperCase();
        }
        return name.substring(0, 2).toUpperCase();
    }

    private _formatRole(role: UserRole | null): string {
        if (!role) return 'User';
        // Capitalize first letter of role
        return role.charAt(0).toUpperCase() + role.slice(1).toLowerCase();
    }

    private _handleProjectClick(projectId: string): void {
        store.actions.setActiveProject(projectId);
    }

    private _handleThreadClick(threadId: string): void {
        store.actions.setActiveThread(threadId);
        store.actions.markThreadRead(threadId);
    }

    private _handleFocusClick(task: FocusTask): void {
        store.actions.setActiveProject(task.projectId);
        if (task.threadId) {
            store.actions.setActiveThread(task.threadId);
        } else {
            // Agent-driven: create a thread with the task title
            void this._createThreadFromFocus(task);
        }
    }

    private async _createThreadFromFocus(task: FocusTask): Promise<void> {
        try {
            const apiThread = await api.threads.create(task.projectId, task.title);
            const thread = this._mapApiThread(apiThread);
            store.actions.addThread(thread);
            store.actions.setActiveThread(thread.id);
        } catch (err) {
            console.error('[FBPanelLeft] Failed to create thread from focus task:', err);
        }
    }

    private async _loadThreads(projectId: string): Promise<void> {
        try {
            const apiThreads = await api.threads.list(projectId);
            const threads = apiThreads.map((t) => this._mapApiThread(t));
            store.actions.setThreads(threads);
            // Auto-select General thread
            const general = threads.find((t) => t.isGeneral);
            if (general) {
                store.actions.setActiveThread(general.id);
            }
        } catch (err) {
            console.error('[FBPanelLeft] Failed to load threads:', err);
        }
    }

    private _showCreateThread(): void {
        this._isCreatingThread = true;
        this._newThreadTitle = '';
    }

    private _cancelCreateThread(): void {
        this._isCreatingThread = false;
        this._newThreadTitle = '';
    }

    private async _submitCreateThread(): Promise<void> {
        const title = this._newThreadTitle.trim();
        if (!title || !this._activeProjectId) return;

        try {
            const apiThread = await api.threads.create(this._activeProjectId, title);
            const thread = this._mapApiThread(apiThread);
            store.actions.addThread(thread);
            store.actions.setActiveThread(thread.id);
            this._isCreatingThread = false;
            this._newThreadTitle = '';
        } catch (err) {
            console.error('[FBPanelLeft] Failed to create thread:', err);
        }
    }

    private _handleCreateThreadKeydown(e: KeyboardEvent): void {
        if (e.key === 'Enter') {
            void this._submitCreateThread();
        } else if (e.key === 'Escape') {
            this._cancelCreateThread();
        }
    }

    private _handleThemeToggle(): void {
        const next = this._theme === 'dark' ? 'light' : 'dark';
        store.actions.setTheme(next);
    }

    private _canCompleteProject(project: ProjectSummary): boolean {
        const role = this._userRole;
        const isBuilderOrAdmin = role === UserRole.Admin || role === UserRole.Builder;
        return isBuilderOrAdmin && project.status === 'Active' && this._activeProjectId === project.id;
    }

    private async _handleCompleteProject(): Promise<void> {
        if (!this._activeProjectId || this._isCompleting) return;
        this._isCompleting = true;
        this._completeError = null;
        try {
            const report = await api.completion.complete(this._activeProjectId);
            store.actions.setCompletionReport(report);
            // Update the project status in the local list
            const updated = this._projects.map((p) =>
                p.id === this._activeProjectId ? { ...p, status: 'Completed' } : p
            );
            store.actions.setProjects(updated);
            this._showCompleteConfirm = false;
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to complete project';
            this._completeError = message;
            console.error('[FBPanelLeft] Failed to complete project:', err);
        } finally {
            this._isCompleting = false;
        }
    }

    private _handleAdminNav(path: string): void {
        window.history.pushState({}, '', path);
        window.dispatchEvent(new PopStateEvent('popstate'));
    }

    override render(): TemplateResult {
        return html`
            <header class="header" @click=${(): void => { this._handleAdminNav('/'); }}>
                <svg class="logo" viewBox="0 0 100 100" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M10 45 L50 15 L90 45"/>
                    <path d="M20 50 L50 25 L80 50"/>
                    <path d="M25 48 L25 85 L75 85 L75 48"/>
                    <path d="M50 50 L40 56 L40 68 L50 74 L60 68 L60 56 Z"/>
                </svg>
                <div class="brand">FUTURE<span>BUILD</span></div>
            </header>

            <!-- Organization Switcher (Step 80) -->
            <div class="org-switcher">
                <div id="org-switcher-container"></div>
            </div>

            <!-- Daily Focus -->
            ${this._focusTasks.length > 0 ? html`
                <section class="section" aria-label="Daily Focus Tasks">
                    <div class="section-header">
                        <span>🔔 Daily Focus</span>
                        <span class="focus-badge" aria-label="${this._focusTasks.length} tasks">${this._focusTasks.length}</span>
                    </div>
                    <div class="section-content" role="list">
                        ${this._focusTasks.map((task) => html`
                            <button 
                                class="item" 
                                role="listitem"
                                @click=${(): void => { this._handleFocusClick(task); }}
                                aria-label="Focus task: ${task.title}"
                            >
                                <span class="item-icon" aria-hidden="true">
                                    <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>
                                </span>
                                ${task.title}
                            </button>
                        `)}
                    </div>
                </section>
            ` : null}

            <!-- Projects -->
            <nav class="section projects-list" aria-label="Projects">
                <div class="section-header">
                    <span>📁 Projects</span>
                    <button
                        class="add-btn"
                        @click=${(): void => { this._handleAdminNav('/projects'); }}
                        aria-label="View all projects"
                        title="All Projects"
                    >
                        <svg viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
                    </button>
                </div>
                <div class="section-content" role="list">
                    ${this._projects.map((project) => html`
                        <div role="listitem">
                            <button 
                                class="item ${this._activeProjectId === project.id ? 'active' : ''}"
                                @click=${(): void => { this._handleProjectClick(project.id); }}
                                aria-expanded="${this._activeProjectId === project.id}"
                                aria-label="Project: ${project.name}"
                            >
                                <span class="item-icon" aria-hidden="true">
                                    <svg viewBox="0 0 24 24"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>
                                </span>
                                ${project.name}
                            </button>
                            ${this._activeProjectId === project.id ? html`
                                <div class="threads" role="list" aria-label="Threads in ${project.name}">
                                    ${this._threads.map((thread) => html`
                                        <button
                                            class="item thread-item ${thread.hasUnread ? 'unread' : ''} ${this._activeThreadId === thread.id ? 'active' : ''}"
                                            role="listitem"
                                            @click=${(): void => { this._handleThreadClick(thread.id); }}
                                            aria-label="Thread: ${thread.title}${thread.hasUnread ? ' (unread)' : ''}"
                                        >
                                            💬 ${thread.title}
                                        </button>
                                    `)}
                                    ${this._isCreatingThread ? html`
                                        <input
                                            class="item thread-item"
                                            type="text"
                                            placeholder="Thread title..."
                                            .value=${this._newThreadTitle}
                                            @input=${(e: InputEvent): void => { this._newThreadTitle = (e.target as HTMLInputElement).value; }}
                                            @keydown=${this._handleCreateThreadKeydown.bind(this)}
                                            @blur=${(): void => { this._cancelCreateThread(); }}
                                            style="border: 1px solid var(--fb-border); border-radius: var(--fb-radius-sm); outline: none;"
                                        />
                                    ` : html`
                                        <button
                                            class="item thread-item"
                                            @click=${(): void => { this._showCreateThread(); }}
                                            aria-label="Create new thread"
                                        >
                                            + New Thread
                                        </button>
                                    `}
                                </div>
                                ${this._canCompleteProject(project) ? html`
                                    ${this._completeError ? html`
                                        <div style="padding: 4px 12px; font-size: 11px; color: var(--fb-danger, #ef4444);">
                                            ${this._completeError}
                                        </div>
                                    ` : nothing}
                                    ${this._showCompleteConfirm ? html`
                                        <div class="complete-confirm" style="padding: 8px 12px; display: flex; gap: 8px; align-items: center;">
                                            <span style="font-size: 12px; color: var(--fb-text-secondary);">Complete?</span>
                                            <button
                                                class="item thread-item"
                                                style="color: var(--fb-success, #22c55e); font-size: 12px;"
                                                @click=${(): void => { void this._handleCompleteProject(); }}
                                                ?disabled=${this._isCompleting}
                                            >
                                                ${this._isCompleting ? 'Completing...' : 'Confirm'}
                                            </button>
                                            <button
                                                class="item thread-item"
                                                style="font-size: 12px;"
                                                @click=${(): void => { this._showCompleteConfirm = false; this._completeError = null; }}
                                            >
                                                Cancel
                                            </button>
                                        </div>
                                    ` : html`
                                        <button
                                            class="item thread-item"
                                            style="color: var(--fb-success, #22c55e);"
                                            @click=${(): void => { this._showCompleteConfirm = true; this._completeError = null; }}
                                            aria-label="Mark project as completed"
                                        >
                                            Complete Project
                                        </button>
                                    `}
                                ` : nothing}
                            ` : null}
                        </div>
                    `)}
                </div>
            </nav>

            <!-- Agent Activity -->
            <fb-agent-activity></fb-agent-activity>

            <!-- Organization (Admin only) -->
            ${this._userRole === UserRole.Admin ? html`
                <section class="section admin-section" aria-label="Organization">
                    <div class="section-header">Organization</div>
                    <div class="section-content" role="list">
                        <button
                            class="item admin-item"
                            role="listitem"
                            @click=${(): void => { this._handleAdminNav('/settings/team'); }}
                            aria-label="Manage team members"
                        >
                            <span class="item-icon" aria-hidden="true">
                                <svg viewBox="0 0 24 24"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                            </span>
                            Team
                        </button>
                    </div>
                </section>
            ` : nothing}

            <!-- Platform Admin Link (allowlisted emails only) -->
            ${this._isPlatformAdmin ? html`
                <div class="section" style="padding-bottom: 0;">
                    <button
                        class="item admin-item"
                        @click=${(): void => { this._handleAdminNav('/admin'); }}
                        aria-label="Platform Administration"
                    >
                        <span class="item-icon" aria-hidden="true">
                            <svg viewBox="0 0 24 24"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
                        </span>
                        Platform Admin
                    </button>
                </div>
            ` : nothing}

            <!-- User -->
            <footer class="user-section">
                <div class="avatar" aria-hidden="true">${this._userInitials}</div>
                <div class="user-info">
                    <div class="user-name">${this._userName || 'Guest'}</div>
                    <div class="user-role">${this._formatRole(this._userRole)} · v2.1.0-beta</div>
                </div>
                <button
                    class="theme-toggle"
                    @click=${(): void => { this._handleAdminNav('/settings'); }}
                    aria-label="Open settings"
                    title="Settings"
                >
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <circle cx="12" cy="12" r="3"/>
                        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                    </svg>
                </button>
                <button
                    class="theme-toggle"
                    @click=${this._handleThemeToggle.bind(this)}
                    aria-label="Toggle theme, current: ${this._theme}"
                >
                    ${this._theme === 'dark' ? html`
                        <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
                    ` : html`
                        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
                    `}
                </button>
            </footer>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-panel-left': FBPanelLeft;
    }
}

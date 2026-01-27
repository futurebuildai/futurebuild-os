/**
 * FBPanelLeft - Left Panel Component (Projects, Threads, Daily Focus, Agent Activity)
 * See FRONTEND_SCOPE.md Section 3.3
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { ProjectSummary, Thread, FocusTask } from '../../store/types';
import { UserRole } from '../../types/enums';
import '../agent/fb-agent-activity';
import '../shadow/shadow-toggle';

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
        `,
    ];

    @state() private _projects: ProjectSummary[] = [];
    @state() private _threads: Thread[] = [];
    @state() private _focusTasks: FocusTask[] = [];
    @state() private _activeProjectId: string | null = null;
    @state() private _activeThreadId: string | null = null;
    @state() private _userName = '';
    @state() private _userInitials = '';
    @state() private _userRole: UserRole | null = null;
    @state() private _theme: 'light' | 'dark' | 'system' = 'system';

    private _disposeEffects: (() => void)[] = [];

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
                this._activeProjectId = store.activeProjectId$.value;
            }),
            effect(() => {
                this._activeThreadId = store.activeThreadId$.value;
            }),
            effect(() => {
                const user = store.user$.value;
                this._userName = user?.name ?? '';
                this._userInitials = this._computeInitials(user?.name);
                this._userRole = user?.role ?? null;
            }),
            effect(() => {
                this._theme = store.theme$.value;
            })
        );
    }

    override disconnectedCallback(): void {
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
        }
    }

    private _handleThemeToggle(): void {
        const next = this._theme === 'dark' ? 'light' : 'dark';
        store.actions.setTheme(next);
    }

    private _handleAdminNav(path: string): void {
        window.history.pushState({}, '', path);
        window.dispatchEvent(new PopStateEvent('popstate'));
    }

    override render(): TemplateResult {
        return html`
            <header class="header">
                <svg class="logo" viewBox="0 0 100 100" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M10 45 L50 15 L90 45"/>
                    <path d="M20 50 L50 25 L80 50"/>
                    <path d="M25 48 L25 85 L75 85 L75 48"/>
                    <path d="M50 50 L40 56 L40 68 L50 74 L60 68 L60 56 Z"/>
                </svg>
                <div class="brand">FUTURE<span>BUILD</span></div>
            </header>

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
                <div class="section-header">📁 Projects</div>
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
                                </div>
                            ` : null}
                        </div>
                    `)}
                </div>
            </nav>

            <!-- Agent Activity -->
            <fb-agent-activity></fb-agent-activity>

            <!-- Admin Navigation (Admin only) -->
            ${this._userRole === UserRole.Admin ? html`
                <section class="section admin-section" aria-label="Admin">
                    <div class="section-header">Admin</div>
                    <div class="section-content" role="list">
                        <button
                            class="item admin-item"
                            role="listitem"
                            @click=${(): void => { this._handleAdminNav('/admin/invites'); }}
                            aria-label="Manage user invitations"
                        >
                            <span class="item-icon" aria-hidden="true">
                                <svg viewBox="0 0 24 24"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><path d="M20 8v6M23 11h-6"/></svg>
                            </span>
                            Invitations
                        </button>
                    </div>
                </section>
            ` : nothing}

            <!-- Shadow Mode Toggle (Admin only) -->
            ${this._userRole === UserRole.Admin ? html`
                <div class="section">
                    <shadow-toggle></shadow-toggle>
                </div>
            ` : nothing}

            <!-- User -->
            <footer class="user-section">
                <div class="avatar" aria-hidden="true">${this._userInitials}</div>
                <div class="user-info">
                    <div class="user-name">${this._userName || 'Guest'}</div>
                    <div class="user-role">${this._formatRole(this._userRole)}</div>
                </div>
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

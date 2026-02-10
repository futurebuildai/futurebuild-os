/**
 * FBAppShell - V2 Root Application Container
 * See FRONTEND_V2_SPEC.md §4.2
 *
 * V2 Layout: Top bar + content area + optional right panel.
 * No left sidebar. Project navigation via top-bar pills.
 *
 * Routing:
 *   /              → Home Feed (fb-home-feed)
 *   /onboard       → Onboarding (fb-onboard-flow) [future]
 *   /project/:id   → Project detail (filtered feed)
 *   /project/:id/chat → Project chat
 *   /project/:id/schedule → Schedule view
 *   /settings/*    → Settings pages
 *   /contacts      → Contact directory
 *   /admin         → Admin shell
 *   /portal/*      → Portal routes
 */
import { html, css, type TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store, initializeStore } from '../../store/store';
import { clerkService } from '../../services/clerk';
import { isPlatformAdmin } from '../../services/platform-admin';
import type { ProjectPill } from '../../types/feed';
import { api } from '../../services/api';

// V2 Components
import './fb-top-bar';
import '../feed/fb-home-feed';
import '../feed/fb-feed-card';

// Kept from V1
import './fb-panel-right';
import '../features/fb-file-drop';
import '../artifacts/fb-artifact-modal';
import '../feedback/fb-toast-container';
import './fb-mobile-nav';
import '../shadow/shadow-layout';
import '../admin/fb-admin-shell';

// V1 components still used for chat/schedule views until replaced
import '../chat/fb-message-list';
import '../chat/fb-input-bar';

/**
 * Route type for V2 URL-based routing.
 * See FRONTEND_V2_SPEC.md §4.3
 */
type Route =
    | { view: 'home' }
    | { view: 'onboard' }
    | { view: 'project'; projectId: string }
    | { view: 'project-chat'; projectId: string; threadId?: string }
    | { view: 'project-schedule'; projectId: string }
    | { view: 'settings-profile' }
    | { view: 'settings-org' }
    | { view: 'settings-team' }
    | { view: 'contacts' }
    | { view: 'admin' }
    | { view: 'login' }
    | { view: 'portal'; subpath: string };

function matchRoute(path: string): Route {
    if (path === '/' || path === '') return { view: 'home' };
    if (path === '/onboard' || path === '/projects/new') return { view: 'onboard' };
    if (path === '/settings/profile') return { view: 'settings-profile' };
    if (path === '/settings/org') return { view: 'settings-org' };
    if (path === '/settings/team') return { view: 'settings-team' };
    if (path === '/contacts') return { view: 'contacts' };
    if (path.startsWith('/admin')) return { view: 'admin' };
    if (path.startsWith('/portal')) return { view: 'portal', subpath: path };
    if (path === '/login') return { view: 'login' };

    // /project/:id/chat/:threadId?
    const chatMatch = path.match(/^\/project\/([^/]+)\/chat(?:\/([^/]+))?$/);
    if (chatMatch) {
        const projectId = chatMatch[1] ?? '';
        const threadId = chatMatch[2];
        return threadId
            ? { view: 'project-chat', projectId, threadId }
            : { view: 'project-chat', projectId };
    }

    // /project/:id/schedule
    const scheduleMatch = path.match(/^\/project\/([^/]+)\/schedule$/);
    if (scheduleMatch) {
        return { view: 'project-schedule', projectId: scheduleMatch[1] ?? '' };
    }

    // /project/:id
    const projectMatch = path.match(/^\/project\/([^/]+)$/);
    if (projectMatch) {
        return { view: 'project', projectId: projectMatch[1] ?? '' };
    }

    // Fallback to home
    return { view: 'home' };
}

@customElement('fb-app-shell')
export class FBAppShell extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: grid;
                grid-template-columns: 1fr;
                grid-template-rows: 56px 1fr;
                height: 100vh;
                width: 100vw;
                overflow: hidden;
                background: var(--fb-bg-primary);
                color: var(--fb-text-primary);
                font-family: var(--fb-font-family);
            }

            /* With right artifact panel */
            :host([panel-open]) {
                grid-template-columns: 1fr var(--fb-right-panel-width, 380px);
            }

            .top-bar-area {
                grid-column: 1 / -1;
            }

            .content-area {
                overflow-y: auto;
                overflow-x: hidden;
            }

            fb-panel-right {
                overflow-y: auto;
            }

            /* Login mode: no top bar, full screen */
            :host([login-mode]) {
                grid-template-rows: 1fr;
            }

            :host([login-mode]) .top-bar-area {
                display: none;
            }

            /* Mobile */
            @media (max-width: 768px) {
                :host {
                    grid-template-columns: 1fr;
                    grid-template-rows: 56px 1fr;
                    padding-bottom: calc(64px + env(safe-area-inset-bottom, 0px));
                }

                :host([panel-open]) {
                    grid-template-columns: 1fr;
                }

                fb-panel-right {
                    position: fixed;
                    top: 56px;
                    right: 0;
                    bottom: 64px;
                    width: min(380px, 100vw);
                    z-index: var(--fb-z-panel, 100);
                    background: var(--fb-bg-primary);
                    border-left: 1px solid var(--fb-border, #2a2a3e);
                    transform: translateX(100%);
                    transition: transform 0.3s ease;
                }

                :host([panel-open]) fb-panel-right {
                    transform: translateX(0);
                }
            }

            .backdrop {
                position: fixed;
                top: 0;
                left: 0;
                width: 100vw;
                height: 100vh;
                background: rgba(0, 0, 0, 0.5);
                backdrop-filter: blur(2px);
                z-index: calc(var(--fb-z-panel, 100) - 1);
                animation: fadeIn 0.2s ease;
            }

            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }

            .loading-screen {
                display: flex;
                align-items: center;
                justify-content: center;
                height: 100vh;
                color: var(--fb-text-muted, #666);
            }
        `,
    ];

    @state() private _resolvedTheme: 'light' | 'dark' = 'dark';
    @state() private _isAuthenticated = false;
    @state() private _clerkLoaded = false;
    @state() private _route: Route = { view: 'home' };
    @state() private _rightPanelOpen = false;
    @state() private _isMobile = false;
    @state() private _hasPopoutArtifact = false;
    @state() private _shadowModeEnabled = false;
    @state() private _isPlatformAdmin = false;
    @state() private _userName = '';
    @state() private _userRole = '';
    @state() private _projects: ProjectPill[] = [];

    private _disposeEffects: (() => void)[] = [];
    private _dragCounter = 0;

    // ---- Drag-and-Drop Handlers (kept from V1) ----

    private _handleDragEnter = (e: DragEvent): void => {
        e.preventDefault();
        e.stopPropagation();
        this._dragCounter++;
        if (this._dragCounter === 1) {
            store.actions.setDragging(true);
        }
    };

    private _handleDragOver = (e: DragEvent): void => {
        e.preventDefault();
        e.stopPropagation();
    };

    private _handleDragLeave = (e: DragEvent): void => {
        e.preventDefault();
        e.stopPropagation();
        this._dragCounter--;
        if (this._dragCounter === 0) {
            store.actions.setDragging(false);
        }
    };

    private _handleDrop = (e: DragEvent): void => {
        e.preventDefault();
        e.stopPropagation();
        this._dragCounter = 0;

        const files = e.dataTransfer?.files;
        if (files && files.length > 0) {
            store.actions.handleFileDrop(files);
        } else {
            store.actions.setDragging(false);
        }
    };

    override connectedCallback(): void {
        super.connectedCallback();

        void this._initClerkAndStore();

        // Drag-and-drop
        this.addEventListener('dragenter', this._handleDragEnter);
        this.addEventListener('dragover', this._handleDragOver);
        this.addEventListener('dragleave', this._handleDragLeave);
        this.addEventListener('drop', this._handleDrop);

        // Navigation events from child components
        this.addEventListener('fb-navigate', this._handleNavigate as EventListener);
        this.addEventListener('fb-filter-change', this._handleFilterChange as EventListener);

        // Theme
        this._disposeEffects.push(
            effect(() => {
                const theme = store.theme$.value;
                this._resolvedTheme = this._resolveTheme(theme);
                document.documentElement.setAttribute('data-theme', this._resolvedTheme);
            })
        );

        // Auth state
        this._disposeEffects.push(
            effect(() => {
                this._isAuthenticated = store.isAuthenticated$.value;
                if (this._isAuthenticated) {
                    this.removeAttribute('login-mode');
                    // Load projects for pills
                    void this._loadProjects();
                } else {
                    this.setAttribute('login-mode', '');
                }
            })
        );

        // User name + role for avatar and menu
        this._disposeEffects.push(
            effect(() => {
                this._userName = store.user$.value?.name ?? '';
                this._userRole = store.user$.value?.role ?? '';
            })
        );

        // Right panel
        this._disposeEffects.push(
            effect(() => {
                this._rightPanelOpen = store.rightPanelOpen$.value;
                if (this._rightPanelOpen) {
                    this.setAttribute('panel-open', '');
                } else {
                    this.removeAttribute('panel-open');
                }
            })
        );

        // Mobile
        this._disposeEffects.push(
            effect(() => {
                this._isMobile = store.isMobile$.value;
            })
        );

        // Popout artifact
        this._disposeEffects.push(
            effect(() => {
                this._hasPopoutArtifact = store.popoutArtifact$.value !== null;
            })
        );

        // Shadow mode
        this._disposeEffects.push(
            effect(() => {
                this._shadowModeEnabled = store.shadowModeEnabled$.value;
            })
        );

        // Platform admin
        this._disposeEffects.push(
            effect(() => {
                const email = store.user$.value?.email ?? '';
                this._isPlatformAdmin = isPlatformAdmin(email);
            })
        );

        // URL routing
        this._syncRoute();
        window.addEventListener('popstate', this._handlePopState);
        this._patchHistory();
    }

    private async _initClerkAndStore(): Promise<void> {
        const publishableKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY as string | undefined;
        if (!publishableKey) {
            console.error('[FBAppShell] VITE_CLERK_PUBLISHABLE_KEY not set. Auth will not work.');
            this._clerkLoaded = true;
            initializeStore();
            return;
        }

        try {
            await clerkService.init(publishableKey);
        } catch (err) {
            console.error('[FBAppShell] Clerk initialization failed:', err);
        }

        this._clerkLoaded = true;
        initializeStore();
        await clerkService.syncAuthState();
    }

    private async _loadProjects(): Promise<void> {
        try {
            const resp = await api.portfolio.getFeed();
            this._projects = resp.projects;
        } catch {
            // Non-critical; pills just won't show
        }
    }

    // ---- Routing ----

    private _handlePopState = (): void => {
        this._syncRoute();
    };

    private _syncRoute(): void {
        this._route = matchRoute(window.location.pathname);
    }

    private _patchHistory(): void {
        const originalPushState = history.pushState;
        const originalReplaceState = history.replaceState;
        const self = this;

        history.pushState = function (...args) {
            originalPushState.apply(this, args);
            self._syncRoute();
        };

        history.replaceState = function (...args) {
            originalReplaceState.apply(this, args);
            self._syncRoute();
        };
    }

    private _navigate(path: string): void {
        if (window.location.pathname !== path) {
            history.pushState({}, '', path);
        }
    }

    private _handleNavigate = (e: CustomEvent<{ view: string; projectId?: string | undefined }>): void => {
        e.stopPropagation();
        const { view, projectId } = e.detail;
        switch (view) {
            case 'home':
                this._navigate('/');
                break;
            case 'onboard':
                this._navigate('/onboard');
                break;
            case 'project':
                if (projectId) this._navigate(`/project/${projectId}`);
                break;
            case 'project-chat':
                if (projectId) this._navigate(`/project/${projectId}/chat`);
                break;
            case 'project-schedule':
                if (projectId) this._navigate(`/project/${projectId}/schedule`);
                break;
            case 'settings-profile':
                this._navigate('/settings/profile');
                break;
            case 'settings-org':
                this._navigate('/settings/org');
                break;
            case 'settings-team':
                this._navigate('/settings/team');
                break;
            case 'contacts':
                this._navigate('/contacts');
                break;
        }
    };

    private _handleFilterChange = (e: CustomEvent<{ projectId: string | null }>): void => {
        e.stopPropagation();
        const { projectId } = e.detail;

        // If filtering to a specific project, navigate to project view
        if (projectId) {
            this._navigate(`/project/${projectId}`);
        } else {
            this._navigate('/');
        }
    };

    override disconnectedCallback(): void {
        this.removeEventListener('dragenter', this._handleDragEnter);
        this.removeEventListener('dragover', this._handleDragOver);
        this.removeEventListener('dragleave', this._handleDragLeave);
        this.removeEventListener('drop', this._handleDrop);
        this.removeEventListener('fb-navigate', this._handleNavigate as EventListener);
        this.removeEventListener('fb-filter-change', this._handleFilterChange as EventListener);

        window.removeEventListener('popstate', this._handlePopState);

        this._disposeEffects.forEach((dispose) => { dispose(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _resolveTheme(theme: string): 'light' | 'dark' {
        if (theme === 'system') {
            return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
        }
        return theme === 'light' ? 'light' : 'dark';
    }

    private _closeRightPanel(): void {
        if (this._rightPanelOpen) store.actions.setRightPanelOpen(false);
    }

    // ---- Render ----

    private _renderContent(): TemplateResult | typeof nothing {
        switch (this._route.view) {
            case 'home':
                return html`<fb-home-feed></fb-home-feed>`;

            case 'project': {
                const route = this._route as { view: 'project'; projectId: string };
                // Project view is a filtered feed
                return html`<fb-home-feed .setFilter=${route.projectId}></fb-home-feed>`;
            }

            case 'onboard':
                // Phase 4: Full onboarding flow component
                return html`
                    <div style="display: flex; align-items: center; justify-content: center; height: 100%; color: var(--fb-text-secondary);">
                        Onboarding flow — coming in Phase 4
                    </div>
                `;

            case 'project-chat': {
                // V1 chat components until replaced in Phase 5
                return html`
                    <div style="display: flex; flex-direction: column; height: 100%;">
                        <fb-message-list></fb-message-list>
                        <fb-input-bar></fb-input-bar>
                    </div>
                `;
            }

            case 'project-schedule': {
                // Phase 6: Full schedule view
                return html`
                    <div style="display: flex; align-items: center; justify-content: center; height: 100%; color: var(--fb-text-secondary);">
                        Schedule view — coming in Phase 6
                    </div>
                `;
            }

            case 'settings-profile':
            case 'settings-org':
            case 'settings-team':
                return html`<fb-view-settings></fb-view-settings>`;

            case 'contacts':
                return html`<fb-view-directory></fb-view-directory>`;

            case 'login':
                return nothing;

            default:
                return html`<fb-home-feed></fb-home-feed>`;
        }
    }

    override render(): TemplateResult {
        if (!this._clerkLoaded) {
            return html`<div class="loading-screen">Loading...</div>`;
        }

        // Admin route
        if (this._route.view === 'admin' && this._isAuthenticated) {
            if (!this._isPlatformAdmin) {
                window.history.replaceState({}, '', '/');
                return html`<div class="loading-screen"></div>`;
            }
            return html`
                <fb-admin-shell></fb-admin-shell>
                <fb-toast-container></fb-toast-container>
            `;
        }

        // Shadow mode
        if (this._isAuthenticated && this._shadowModeEnabled) {
            return html`
                <shadow-layout></shadow-layout>
                <fb-toast-container></fb-toast-container>
            `;
        }

        // Portal routes pass through to V1 panel-center for now
        if (this._route.view === 'portal') {
            return html`
                <fb-panel-center .isAuthenticated=${false}></fb-panel-center>
                <fb-toast-container></fb-toast-container>
            `;
        }

        // Determine active filter for top bar pills
        const activeFilter =
            this._route.view === 'project' || this._route.view === 'project-chat' || this._route.view === 'project-schedule'
                ? (this._route as { projectId: string }).projectId
                : null;

        return html`
            <fb-file-drop></fb-file-drop>

            <div class="top-bar-area">
                ${this._isAuthenticated
                    ? html`
                          <fb-top-bar
                              .projects=${this._projects}
                              .activeFilter=${activeFilter}
                              .userName=${this._userName}
                              .userRole=${this._userRole}
                          ></fb-top-bar>
                      `
                    : nothing}
            </div>

            <div class="content-area">
                ${this._isAuthenticated ? this._renderContent() : html`
                    <fb-panel-center .isAuthenticated=${false}></fb-panel-center>
                `}
            </div>

            ${this._rightPanelOpen ? html`<fb-panel-right></fb-panel-right>` : nothing}

            ${this._isMobile && this._rightPanelOpen ? html`
                <div class="backdrop" @click=${this._closeRightPanel.bind(this)}></div>
            ` : nothing}

            ${this._hasPopoutArtifact ? html`<fb-artifact-modal></fb-artifact-modal>` : nothing}

            ${this._isAuthenticated && this._isMobile ? html`<fb-mobile-nav></fb-mobile-nav>` : nothing}

            <fb-toast-container></fb-toast-container>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-app-shell': FBAppShell;
    }
}

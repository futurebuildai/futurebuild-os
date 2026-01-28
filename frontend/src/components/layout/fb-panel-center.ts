/**
 * FBPanelCenter - Center Panel (Conversation / Login)
 * See FRONTEND_SCOPE.md Section 3.3
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { ChatMessage, ProjectSummary, Thread } from '../../store/types';

// Import view components
import '../views/fb-view-login';
import '../views/fb-view-verify';
import '../views/fb-view-invite-accept';
import '../views/fb-view-admin-invites';
import '../views/fb-view-settings';
import '../views/fb-view-projects';

// Import portal view components (LAUNCH_PLAN.md P2)
import '../views/fb-view-portal-action';
import '../views/fb-view-portal-login';
import '../views/fb-view-portal-signup';
import '../views/fb-view-portal-verify';
import '../views/fb-view-portal-dashboard';

// Import chat components (Step 52 Integration)
import '../chat/fb-message-list';
import '../chat/fb-input-bar';

@customElement('fb-panel-center')
export class FBPanelCenter extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                background: var(--fb-bg-primary);
                overflow: hidden;
            }

            .breadcrumb {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-md) var(--fb-spacing-lg);
                border-bottom: 1px solid var(--fb-border-light);
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                flex-shrink: 0;
            }

            .breadcrumb-item {
                cursor: pointer;
            }

            .breadcrumb-item:hover {
                color: var(--fb-text-primary);
            }

            .breadcrumb-separator {
                color: var(--fb-text-muted);
            }

            .breadcrumb-current {
                color: var(--fb-text-primary);
                font-weight: 500;
            }

            .panel-toggle {
                margin-left: auto;
                padding: var(--fb-spacing-xs);
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-sm);
            }

            .panel-toggle:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .panel-toggle:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .panel-toggle svg {
                width: 18px;
                height: 18px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .empty-state {
                flex: 1;
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                color: var(--fb-text-muted);
                text-align: center;
                padding: var(--fb-spacing-2xl);
            }

            .empty-icon {
                font-size: 48px;
                margin-bottom: var(--fb-spacing-md);
            }

            .empty-title {
                font-size: var(--fb-text-lg);
                font-weight: 500;
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-sm);
            }

            fb-view-login,
            fb-view-verify,
            fb-view-invite-accept {
                flex: 1;
            }
        `,
    ];

    @property({ type: Boolean, attribute: 'is-authenticated' }) isAuthenticated = false;

    @state() private _activeProject: ProjectSummary | null = null;
    @state() private _activeThread: Thread | null = null;
    @state() private _isMobile = false;
    @state() private _hasMessages = false;
    @state() private _isVerifyRoute = false;
    @state() private _isInviteAcceptRoute = false;
    @state() private _isAdminInvitesRoute = false;
    @state() private _isSettingsRoute = false;
    @state() private _isProjectsRoute = false;

    // Portal routes (LAUNCH_PLAN.md P2)
    @state() private _isPortalActionRoute = false;
    @state() private _portalActionToken: string | null = null;
    @state() private _isPortalLoginRoute = false;
    @state() private _isPortalSignupRoute = false;
    @state() private _isPortalVerifyRoute = false;
    @state() private _isPortalDashboardRoute = false;

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();

        // Check for verify route
        this._checkRoute();
        window.addEventListener('popstate', this._handlePopState);

        this._disposeEffects.push(
            effect(() => {
                this._activeProject = store.currentProject$.value;
            }),
            effect(() => {
                this._activeThread = store.activeThread$.value;
            }),
            effect(() => {
                this._isMobile = store.isMobile$.value;
            }),
            // Step 56: Track messages for file drop display
            effect(() => {
                this._hasMessages = store.messages$.value.length > 0;
            })
        );
    }

    private _handlePopState = (): void => {
        this._checkRoute();
    };

    private _checkRoute(): void {
        const path = window.location.pathname;
        this._isVerifyRoute = path === '/auth/verify';
        this._isInviteAcceptRoute = path === '/invite/accept';
        this._isAdminInvitesRoute = path === '/admin/invites';
        this._isSettingsRoute = path === '/settings';
        this._isProjectsRoute = path === '/projects';

        // Portal routes (LAUNCH_PLAN.md P2)
        this._isPortalLoginRoute = path === '/portal/login';
        this._isPortalSignupRoute = path === '/portal/signup';
        this._isPortalVerifyRoute = path === '/portal/verify';
        this._isPortalDashboardRoute = path === '/portal/dashboard';

        // Portal action route: /portal/action/:token
        const actionMatch = path.match(/^\/portal\/action\/([a-zA-Z0-9_-]+)$/);
        if (actionMatch && actionMatch[1]) {
            this._isPortalActionRoute = true;
            this._portalActionToken = actionMatch[1];
        } else {
            this._isPortalActionRoute = false;
            this._portalActionToken = null;
        }
    }

    override disconnectedCallback(): void {
        window.removeEventListener('popstate', this._handlePopState);
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _handleSend(e: CustomEvent<{ content: string }>): void {
        const content = e.detail.content;

        const message: ChatMessage = {
            id: `msg-${String(Date.now())}`,
            role: 'user',
            content: content,
            createdAt: new Date().toISOString(),
        };

        store.actions.addMessage(message);

        // Step 57: Send chat message via RealtimeService
        // Response is handled by store's realtime event listener
        // See FRONTEND_SCOPE.md Section 8.4
        void import('../../services/realtime').then(({ realtimeService }) => {
            realtimeService.send({
                type: 'chat',
                payload: { content },
            });
        });
    }

    private _toggleLeftPanel(): void {
        store.actions.toggleLeftPanel();
    }

    private _toggleRightPanel(): void {
        store.actions.toggleRightPanel();
    }

    override render(): TemplateResult {
        // Portal routes (LAUNCH_PLAN.md P2) - These use portal-specific auth, not main app auth
        if (this._isPortalActionRoute && this._portalActionToken) {
            return html`<fb-view-portal-action .token=${this._portalActionToken}></fb-view-portal-action>`;
        }
        if (this._isPortalLoginRoute) {
            return html`<fb-view-portal-login></fb-view-portal-login>`;
        }
        if (this._isPortalSignupRoute) {
            return html`<fb-view-portal-signup></fb-view-portal-signup>`;
        }
        if (this._isPortalVerifyRoute) {
            return html`<fb-view-portal-verify></fb-view-portal-verify>`;
        }
        if (this._isPortalDashboardRoute) {
            return html`<fb-view-portal-dashboard></fb-view-portal-dashboard>`;
        }

        // Show verify view for magic link verification
        if (this._isVerifyRoute) {
            return html`<fb-view-verify></fb-view-verify>`;
        }

        // Show invite acceptance view
        if (this._isInviteAcceptRoute) {
            return html`<fb-view-invite-accept></fb-view-invite-accept>`;
        }

        // Show login if not authenticated
        if (!this.isAuthenticated) {
            return html`<fb-view-login></fb-view-login>`;
        }

        // Show admin invites view for authenticated admin users
        if (this._isAdminInvitesRoute) {
            return html`<fb-view-admin-invites></fb-view-admin-invites>`;
        }

        // Show settings view for authenticated users
        if (this._isSettingsRoute) {
            return html`<fb-view-settings></fb-view-settings>`;
        }

        // Show projects view for authenticated users
        if (this._isProjectsRoute) {
            return html`<fb-view-projects></fb-view-projects>`;
        }

        return html`
            <!-- Breadcrumb -->
            <nav class="breadcrumb" aria-label="Current context">
                ${this._isMobile ? html`
                    <button 
                        class="panel-toggle" 
                        @click=${this._toggleLeftPanel.bind(this)} 
                        aria-label="Open navigation panel"
                    >
                        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 12h18M3 6h18M3 18h18"/></svg>
                    </button>
                ` : nothing}
                
                ${this._activeProject ? html`
                    <span class="breadcrumb-item">${this._activeProject.name}</span>
                    ${this._activeThread ? html`
                        <span class="breadcrumb-separator" aria-hidden="true">›</span>
                        <span class="breadcrumb-current">${this._activeThread.title}</span>
                    ` : nothing}
                ` : html`
                    <span class="breadcrumb-current">Select a project</span>
                `}
                
                <button 
                    class="panel-toggle" 
                    @click=${this._toggleRightPanel.bind(this)} 
                    aria-label="Toggle artifacts panel"
                >
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <rect x="3" y="3" width="18" height="18" rx="2"/>
                        <path d="M9 3v18"/>
                    </svg>
                </button>
            </nav>

            <!-- Conversation / Empty State -->
            ${this._activeThread || this._hasMessages ? html`
                <fb-message-list></fb-message-list>
                <fb-input-bar @send=${this._handleSend.bind(this)}></fb-input-bar>
            ` : html`
                <div class="empty-state" role="status">
                    <div class="empty-icon" aria-hidden="true">💬</div>
                    <div class="empty-title">No conversation selected</div>
                    <div>Select a project and thread from the left panel to start chatting.</div>
                    <div style="margin-top: var(--fb-spacing-md); font-size: var(--fb-text-sm);">
                        Or drag and drop a file anywhere to upload.
                    </div>
                </div>
            `}
        `;
    }
}

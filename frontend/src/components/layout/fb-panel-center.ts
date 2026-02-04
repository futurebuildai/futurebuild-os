/**
 * FBPanelCenter - Center Panel (Conversation / Login)
 * See FRONTEND_SCOPE.md Section 3.3
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

// Import view components
import '../views/fb-view-login';
// fb-view-verify removed — Clerk handles email verification internally (Phase 12)
import '../views/fb-view-invite-accept';
import '../views/fb-view-settings';
import '../views/fb-view-projects';
import '../views/fb-view-team';

// Import portal view components (LAUNCH_PLAN.md P2)
import '../views/fb-view-portal-action';
import '../views/fb-view-portal-login';
import '../views/fb-view-portal-signup';
import '../views/fb-view-portal-verify';
import '../views/fb-view-portal-dashboard';

// Import chat view (Step 72 Integration)
import '../views/fb-view-chat';

// Import onboarding view (Step 74 Split-Screen Wizard)
import '../views/fb-view-onboarding';


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
            fb-view-invite-accept,
            fb-view-chat,
            fb-view-onboarding,
            fb-view-team {
                flex: 1;
            }

        `,
    ];

    @property({ type: Boolean, attribute: 'is-authenticated' }) isAuthenticated = false;

    @state() private _isInviteAcceptRoute = false;
    @state() private _isSettingsRoute = false;
    @state() private _isProjectsRoute = false;
    @state() private _isOnboardingRoute = false;
    @state() private _isTeamRoute = false;

    // Portal routes (LAUNCH_PLAN.md P2)
    @state() private _isPortalActionRoute = false;
    @state() private _portalActionToken: string | null = null;
    @state() private _isPortalLoginRoute = false;
    @state() private _isPortalSignupRoute = false;
    @state() private _isPortalVerifyRoute = false;
    @state() private _isPortalDashboardRoute = false;

    override connectedCallback(): void {
        super.connectedCallback();
        this._checkRoute();
        window.addEventListener('popstate', this._handlePopState);
    }

    private _handlePopState = (): void => {
        this._checkRoute();
    };

    private _checkRoute(): void {
        const path = window.location.pathname;
        this._isInviteAcceptRoute = path === '/invite/accept';
        this._isSettingsRoute = path === '/settings';
        this._isProjectsRoute = path === '/projects';
        this._isOnboardingRoute = path === '/projects/new';
        this._isTeamRoute = path === '/settings/team';

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
        super.disconnectedCallback();
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

        // Show invite acceptance view
        if (this._isInviteAcceptRoute) {
            return html`<fb-view-invite-accept></fb-view-invite-accept>`;
        }

        // Show login if not authenticated
        if (!this.isAuthenticated) {
            return html`<fb-view-login></fb-view-login>`;
        }

        // Show settings view for authenticated users
        if (this._isSettingsRoute) {
            return html`<fb-view-settings></fb-view-settings>`;
        }

        // Show team management view for authenticated users (Step 80)
        if (this._isTeamRoute) {
            return html`<fb-view-team></fb-view-team>`;
        }

        // Show onboarding wizard (Step 74)
        if (this._isOnboardingRoute) {
            return html`<fb-view-onboarding></fb-view-onboarding>`;
        }

        // Show projects view for authenticated users
        if (this._isProjectsRoute) {
            return html`<fb-view-projects></fb-view-projects>`;
        }

        return html`
            <fb-view-chat></fb-view-chat>
        `;
    }
}

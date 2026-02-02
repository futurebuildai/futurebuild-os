/**
 * FBMobileNav - Bottom Tab Bar for Mobile Navigation
 * See STEP_90_MOBILE_NAV.md
 *
 * Fixed bottom navigation bar visible only on screens < 768px.
 * Provides thumb-friendly access to: Dashboard, Projects, Chat, Settings.
 * Active state derived from current URL path.
 *
 * @element fb-mobile-nav
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';

interface NavTab {
    id: string;
    label: string;
    path: string;
    /** Match function: returns true if this tab should be active for the given path */
    match: (path: string) => boolean;
}

const TABS: NavTab[] = [
    {
        id: 'dashboard',
        label: 'Home',
        path: '/',
        match: (p) => p === '/' || p === '',
    },
    {
        id: 'projects',
        label: 'Projects',
        path: '/projects',
        match: (p) => p === '/projects' || p.startsWith('/projects/'),
    },
    {
        id: 'chat',
        label: 'Chat',
        path: '/chat',
        match: (p) => p === '/chat',
    },
    {
        id: 'settings',
        label: 'Settings',
        path: '/settings',
        match: (p) => p === '/settings' || p.startsWith('/settings/'),
    },
];

@customElement('fb-mobile-nav')
export class FBMobileNav extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: none;
                position: fixed;
                bottom: 0;
                left: 0;
                right: 0;
                z-index: 1000;
                background: var(--fb-bg-secondary, #0a0a0a);
                border-top: 1px solid var(--fb-border, #333);
                /* Safe area for iOS notch devices */
                padding-bottom: env(safe-area-inset-bottom, 0px);
            }

            @media (max-width: 767px) {
                :host {
                    display: flex;
                }
            }

            nav {
                display: flex;
                align-items: stretch;
                justify-content: space-around;
                width: 100%;
                height: 64px;
            }

            .tab {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                gap: 4px;
                flex: 1;
                min-width: 0;
                padding: var(--fb-spacing-xs, 4px) 0;
                border: none;
                background: transparent;
                color: var(--fb-text-muted, #666);
                cursor: pointer;
                /* Minimum 44x44px touch target (WCAG 2.5.8) */
                min-height: 44px;
                min-width: 44px;
                -webkit-tap-highlight-color: transparent;
                transition: color var(--fb-transition-fast, 150ms ease);
                font-family: var(--fb-font-family, 'Poppins', system-ui, sans-serif);
                text-decoration: none;
            }

            .tab:focus-visible {
                outline: 2px solid var(--fb-primary, #667eea);
                outline-offset: -2px;
                border-radius: var(--fb-radius-sm, 4px);
            }

            .tab.active {
                color: var(--fb-primary, #667eea);
            }

            .tab:not(.active):hover {
                color: var(--fb-text-secondary, #aaa);
            }

            .tab-icon {
                width: 22px;
                height: 22px;
                flex-shrink: 0;
            }

            .tab-icon svg {
                width: 100%;
                height: 100%;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
                stroke-linecap: round;
                stroke-linejoin: round;
            }

            .tab-label {
                font-size: 10px;
                font-weight: 500;
                line-height: 1;
                white-space: nowrap;
            }

            .tab.active .tab-label {
                font-weight: 600;
            }
        `,
    ];

    @state() private _activePath = '/';

    private _handlePopState = (): void => {
        this._activePath = window.location.pathname;
    };

    override connectedCallback(): void {
        super.connectedCallback();
        this._activePath = window.location.pathname;
        window.addEventListener('popstate', this._handlePopState);
    }

    override disconnectedCallback(): void {
        window.removeEventListener('popstate', this._handlePopState);
        super.disconnectedCallback();
    }

    private _navigate(path: string): void {
        // Close any open mobile panels first
        if (store.leftPanelOpen$.value) {
            store.actions.setLeftPanelOpen(false);
        }
        if (store.rightPanelOpen$.value) {
            store.actions.setRightPanelOpen(false);
        }

        window.history.pushState({}, '', path);
        window.dispatchEvent(new PopStateEvent('popstate'));
        this._activePath = path;
    }

    private _renderIcon(id: string): TemplateResult | typeof nothing {
        switch (id) {
            case 'dashboard':
                return html`<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>`;
            case 'projects':
                return html`<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>`;
            case 'chat':
                return html`<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`;
            case 'settings':
                return html`<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>`;
            default:
                return nothing;
        }
    }

    override render(): TemplateResult {
        return html`
            <nav aria-label="Mobile navigation" role="navigation">
                ${TABS.map((tab) => {
                    const isActive = tab.match(this._activePath);
                    return html`
                        <button
                            class="tab ${isActive ? 'active' : ''}"
                            @click=${(): void => { this._navigate(tab.path); }}
                            aria-label="${tab.label}"
                            aria-current=${isActive ? 'page' : 'false'}
                            role="tab"
                            aria-selected=${isActive ? 'true' : 'false'}
                        >
                            <span class="tab-icon">
                                ${this._renderIcon(tab.id)}
                            </span>
                            <span class="tab-label">${tab.label}</span>
                        </button>
                    `;
                })}
            </nav>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-mobile-nav': FBMobileNav;
    }
}
